package logic

import (
	"context"
	"fmt"
	"time"

	"cus-cms/internal/cache"
	"cus-cms/internal/dto/req"
	"cus-cms/internal/dto/res"
	"cus-cms/internal/model"
)

// securityConfigCacheTTL 安全配置缓存过期时间。
const securityConfigCacheTTL = 10 * time.Minute

// SecurityLogic 安全管理业务逻辑结构体。
type SecurityLogic struct {
	configModel    *model.SecurityConfigModel
	blacklistModel *model.IPBlacklistRecordModel
	securityCache  *cache.SecurityCache
}

// NewSecurityLogic 创建 SecurityLogic 实例。
func NewSecurityLogic() *SecurityLogic {
	return &SecurityLogic{
		configModel:    model.NewSecurityConfig(),
		blacklistModel: model.NewIPBlacklistRecord(),
		securityCache:  cache.NewSecurityCache(),
	}
}

// InitCache 从数据库加载安全配置到 Redis 缓存，供应用启动时调用。
// 如果缓存已存在则跳过，避免覆盖刚更新的配置。
func (l *SecurityLogic) InitCache(ctx context.Context) error {
	// 检查缓存是否已存在
	cached, err := l.securityCache.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("检查安全配置缓存失败: %w", err)
	}
	if cached != nil {
		return nil // 缓存已存在，跳过
	}

	// 从数据库加载配置
	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("加载安全配置到缓存失败: %w", err)
	}

	// 写入缓存
	configCache := securityConfigToCache(config)
	if err := l.securityCache.SetConfig(ctx, configCache, securityConfigCacheTTL); err != nil {
		return fmt.Errorf("写入安全配置缓存失败: %w", err)
	}

	return nil
}

// GetConfig 获取安全配置（优先读缓存，缓存未命中则查数据库并写入缓存）。
func (l *SecurityLogic) GetConfig(ctx context.Context) (*res.SecurityConfigRes, error) {
	// 优先读取缓存
	cached, err := l.securityCache.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取安全配置缓存失败: %w", err)
	}
	if cached != nil {
		return securityConfigCacheToRes(cached), nil
	}

	// 缓存未命中，查询数据库
	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询安全配置失败: %w", err)
	}

	// 写入缓存（失败不影响主流程）
	configCache := securityConfigToCache(config)
	_ = l.securityCache.SetConfig(ctx, configCache, securityConfigCacheTTL)

	return securityConfigToRes(config), nil
}

// UpdateConfig 更新安全配置，更新数据库后删除缓存以便下次读取刷新。
func (l *SecurityLogic) UpdateConfig(ctx context.Context, r *req.UpdateSecurityConfigReq) error {
	// 先获取当前配置
	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("查询安全配置失败: %w", err)
	}

	// 仅更新提供的字段
	if r.GetMaxTokens != nil {
		config.GetMaxTokens = *r.GetMaxTokens
	}
	if r.GetWindowSeconds != nil {
		config.GetWindowSeconds = *r.GetWindowSeconds
	}
	if r.PostMaxTokens != nil {
		config.PostMaxTokens = *r.PostMaxTokens
	}
	if r.PostWindowSeconds != nil {
		config.PostWindowSeconds = *r.PostWindowSeconds
	}
	if r.ViewMaxTokens != nil {
		config.ViewMaxTokens = *r.ViewMaxTokens
	}
	if r.ViewWindowSeconds != nil {
		config.ViewWindowSeconds = *r.ViewWindowSeconds
	}
	if r.LikeMaxTokens != nil {
		config.LikeMaxTokens = *r.LikeMaxTokens
	}
	if r.LikeWindowSeconds != nil {
		config.LikeWindowSeconds = *r.LikeWindowSeconds
	}
	if r.BlacklistThreshold != nil {
		config.BlacklistThreshold = *r.BlacklistThreshold
	}
	if r.BlacklistTTLMinutes != nil {
		config.BlacklistTTLMinutes = *r.BlacklistTTLMinutes
	}

	// 更新数据库
	if err := l.configModel.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("更新安全配置失败: %w", err)
	}

	// 直接写入新缓存，确保立即生效（热更新）
	configCache := securityConfigToCache(config)
	_ = l.securityCache.SetConfig(ctx, configCache, securityConfigCacheTTL)

	return nil
}

// GetBlacklist 获取 IP 黑名单列表（分页）。
func (l *SecurityLogic) GetBlacklist(ctx context.Context, page, pageSize int) (*res.PageRes[res.BlacklistItemRes], error) {
	records, total, err := l.blacklistModel.GetList(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询黑名单列表失败: %w", err)
	}

	items := make([]res.BlacklistItemRes, 0, len(records))
	for _, record := range records {
		items = append(items, res.BlacklistItemRes{
			ID:        record.ID,
			IPAddress: record.IPAddress,
			Reason:    record.Reason,
			BannedAt:  record.BannedAt,
			ExpiresAt: record.ExpiresAt,
			IsActive:  record.IsActive,
		})
	}

	return res.NewPageRes(items, total, page, pageSize), nil
}

// UnbanIP 解封 IP，同时从数据库和缓存中移除。
func (l *SecurityLogic) UnbanIP(ctx context.Context, ip string) error {
	// 从数据库中删除（停用）
	if err := l.blacklistModel.Delete(ctx, ip); err != nil {
		return fmt.Errorf("解封 IP 失败: %w", err)
	}

	// 从缓存中移除
	_ = l.securityCache.RemoveFromBlacklist(ctx, ip)

	return nil
}

// GetStats 获取安全统计数据。
func (l *SecurityLogic) GetStats(ctx context.Context) (*res.SecurityStatsRes, error) {
	// 获取当前封禁 IP 数量
	blockedCount, err := l.securityCache.GetBlacklistCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取封禁 IP 数量失败: %w", err)
	}

	// 获取今日限流次数
	today := time.Now().Format("2006-01-02")
	todayCount, err := l.securityCache.GetDailyRateLimitCount(ctx, today)
	if err != nil {
		return nil, fmt.Errorf("获取今日限流次数失败: %w", err)
	}

	// 获取最近 7 天的限流趋势
	dailyTrend := make([]res.DailyCountRes, 0, 7)
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		count, err := l.securityCache.GetDailyRateLimitCount(ctx, date)
		if err != nil {
			return nil, fmt.Errorf("获取每日限流次数失败: %w", err)
		}
		dailyTrend = append(dailyTrend, res.DailyCountRes{
			Date:  date,
			Count: count,
		})
	}

	// TopViolations 暂时返回空数组
	topViolations := []res.TopIPRes{}

	return &res.SecurityStatsRes{
		BlockedIPCount:      blockedCount,
		TodayRateLimitCount: todayCount,
		DailyTrend:          dailyTrend,
		TopViolations:       topViolations,
	}, nil
}

// securityConfigToRes 将数据库模型转换为响应结构体。
func securityConfigToRes(config *model.SecurityConfig) *res.SecurityConfigRes {
	return &res.SecurityConfigRes{
		GetMaxTokens:        config.GetMaxTokens,
		GetWindowSeconds:    config.GetWindowSeconds,
		PostMaxTokens:       config.PostMaxTokens,
		PostWindowSeconds:   config.PostWindowSeconds,
		ViewMaxTokens:       config.ViewMaxTokens,
		ViewWindowSeconds:   config.ViewWindowSeconds,
		LikeMaxTokens:       config.LikeMaxTokens,
		LikeWindowSeconds:   config.LikeWindowSeconds,
		BlacklistThreshold:  config.BlacklistThreshold,
		BlacklistTTLMinutes: config.BlacklistTTLMinutes,
	}
}

// securityConfigCacheToRes 将缓存结构体转换为响应结构体。
func securityConfigCacheToRes(cached *cache.SecurityConfigCache) *res.SecurityConfigRes {
	return &res.SecurityConfigRes{
		GetMaxTokens:        cached.GetMaxTokens,
		GetWindowSeconds:    cached.GetWindowSeconds,
		PostMaxTokens:       cached.PostMaxTokens,
		PostWindowSeconds:   cached.PostWindowSeconds,
		ViewMaxTokens:       cached.ViewMaxTokens,
		ViewWindowSeconds:   cached.ViewWindowSeconds,
		LikeMaxTokens:       cached.LikeMaxTokens,
		LikeWindowSeconds:   cached.LikeWindowSeconds,
		BlacklistThreshold:  cached.BlacklistThreshold,
		BlacklistTTLMinutes: cached.BlacklistTTLMinutes,
	}
}

// securityConfigToCache 将数据库模型转换为缓存结构体。
func securityConfigToCache(config *model.SecurityConfig) *cache.SecurityConfigCache {
	return &cache.SecurityConfigCache{
		GetMaxTokens:        config.GetMaxTokens,
		GetWindowSeconds:    config.GetWindowSeconds,
		PostMaxTokens:       config.PostMaxTokens,
		PostWindowSeconds:   config.PostWindowSeconds,
		ViewMaxTokens:       config.ViewMaxTokens,
		ViewWindowSeconds:   config.ViewWindowSeconds,
		LikeMaxTokens:       config.LikeMaxTokens,
		LikeWindowSeconds:   config.LikeWindowSeconds,
		BlacklistThreshold:  config.BlacklistThreshold,
		BlacklistTTLMinutes: config.BlacklistTTLMinutes,
	}
}
