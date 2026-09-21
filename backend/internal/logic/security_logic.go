package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"novablog/enum"
	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"gorm.io/gorm"
)

// securityConfigCacheTTL 安全配置缓存过期时间。
const securityConfigCacheTTL = 10 * time.Minute

// SecurityLogic 安全管理业务逻辑结构体。
type SecurityLogic struct {
	configModel    *model.SecurityConfigModel
	securityModel  *model.SecurityModel
	accessLogModel *model.AccessLogModel
	securityCache  *cache.SecurityCache
}

// NewSecurityLogic 创建 SecurityLogic 实例。
func NewSecurityLogic() *SecurityLogic {
	return &SecurityLogic{
		configModel:    model.NewSecurityConfig(),
		securityModel:  model.NewSecurity(),
		accessLogModel: model.NewAccessLog(),
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
	if r.SecurityEnabled != nil {
		config.SecurityEnabled = *r.SecurityEnabled
	}
	if r.BlacklistTTLMinutes != nil {
		config.BlacklistTTLMinutes = *r.BlacklistTTLMinutes
	}
	if r.LogRetentionDays != nil {
		config.LogRetentionDays = *r.LogRetentionDays
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

// CreateBlacklist 创建黑名单记录并同步到缓存（永久封禁，需手动解封）。
func (l *SecurityLogic) CreateBlacklist(ctx context.Context, r *req.BlacklistReq) error {
	exists, err := l.securityModel.GetBlacklistByIP(ctx, r.IP)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询黑名单失败: %w", err)
	}
	if exists != nil {
		return fmt.Errorf("IP %s 已存在黑名单中", r.IP)
	}

	record := &model.IPBlacklist{
		IP:     r.IP,
		Reason: r.Reason,
	}
	if err := l.securityModel.CreateBlacklist(ctx, record); err != nil {
		return fmt.Errorf("创建黑名单失败: %w", err)
	}

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_ = l.securityCache.AddToBlacklist(cacheCtx, record.IP, 0)

	return nil
}

// DeleteBlacklist 根据 ID 删除黑名单记录并从缓存移除。
func (l *SecurityLogic) DeleteBlacklist(ctx context.Context, id uint) error {
	record, err := l.securityModel.GetBlacklistByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return enum.ErrNotFound
		}
		return fmt.Errorf("查询黑名单失败: %w", err)
	}

	if err := l.securityModel.DeleteBlacklist(ctx, id); err != nil {
		return fmt.Errorf("删除黑名单失败: %w", err)
	}

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_ = l.securityCache.RemoveFromBlacklist(cacheCtx, record.IP)

	return nil
}

// UpdateBlacklist 更新黑名单记录并刷新缓存。
func (l *SecurityLogic) UpdateBlacklist(ctx context.Context, id uint, r *req.BlacklistReq) error {
	record, err := l.securityModel.GetBlacklistByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return enum.ErrNotFound
		}
		return fmt.Errorf("查询黑名单失败: %w", err)
	}

	oldIP := record.IP
	record.IP = r.IP
	record.Reason = r.Reason
	if err := l.securityModel.UpdateBlacklist(ctx, record); err != nil {
		return fmt.Errorf("更新黑名单失败: %w", err)
	}

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if oldIP != record.IP {
		_ = l.securityCache.RemoveFromBlacklist(cacheCtx, oldIP)
	}
	_ = l.securityCache.AddToBlacklist(cacheCtx, record.IP, 0)

	return nil
}

// GetBlacklist 根据 ID 获取黑名单记录。
func (l *SecurityLogic) GetBlacklist(ctx context.Context, id uint) (*res.BlacklistRes, error) {
	record, err := l.securityModel.GetBlacklistByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, enum.ErrNotFound
		}
		return nil, fmt.Errorf("查询黑名单失败: %w", err)
	}

	return &res.BlacklistRes{
		ID:        record.ID,
		IP:        record.IP,
		Reason:    record.Reason,
		CreatedAt: record.CreatedAt,
	}, nil
}

// ListBlacklists 分页获取黑名单列表。
func (l *SecurityLogic) ListBlacklists(ctx context.Context, r *req.ListBlacklistReq) (*res.ListBlacklistRes, error) {
	page, pageSize := r.GetPage(), r.GetPageSize()
	records, total, err := l.securityModel.ListBlacklists(ctx, r.Keyword, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询黑名单列表失败: %w", err)
	}

	list := make([]*res.BlacklistRes, 0, len(records))
	for _, record := range records {
		list = append(list, &res.BlacklistRes{
			ID:        record.ID,
			IP:        record.IP,
			Reason:    record.Reason,
			CreatedAt: record.CreatedAt,
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	if list == nil {
		list = []*res.BlacklistRes{}
	}

	return &res.ListBlacklistRes{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetAccessStatistics 获取按 IP 聚合的访问统计，支持 IP 前缀与地区模糊搜索。
func (l *SecurityLogic) GetAccessStatistics(ctx context.Context, r *req.ListAccessLogReq) (*res.ListIPAccessStatsRes, error) {
	page, pageSize := r.GetPage(), r.GetPageSize()
	records, total, err := l.accessLogModel.GetIPAccessStatistics(ctx, r.IP, r.Region, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询 IP 访问统计失败: %w", err)
	}

	list := make([]*res.IPAccessStatsRes, 0, len(records))
	for _, record := range records {
		list = append(list, &res.IPAccessStatsRes{
			IP:           record.IP,
			Region:       record.Region,
			TotalCount:   record.TotalCount,
			ErrorCount:   record.ErrorCount,
			LastAccessAt: record.LastAccessAt,
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	if list == nil {
		list = []*res.IPAccessStatsRes{}
	}

	return &res.ListIPAccessStatsRes{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// securityConfigToRes 将数据库模型转换为响应结构体。
func securityConfigToRes(config *model.SecurityConfig) *res.SecurityConfigRes {
	return &res.SecurityConfigRes{
		SecurityEnabled:     config.SecurityEnabled,
		BlacklistTTLMinutes: config.BlacklistTTLMinutes,
		LogRetentionDays:    config.LogRetentionDays,
	}
}

// securityConfigCacheToRes 将缓存结构体转换为响应结构体。
func securityConfigCacheToRes(cached *cache.SecurityConfigCache) *res.SecurityConfigRes {
	return &res.SecurityConfigRes{
		SecurityEnabled:     cached.SecurityEnabled,
		BlacklistTTLMinutes: cached.BlacklistTTLMinutes,
		LogRetentionDays:    cached.LogRetentionDays,
	}
}

// securityConfigToCache 将数据库模型转换为缓存结构体。
func securityConfigToCache(config *model.SecurityConfig) *cache.SecurityConfigCache {
	return &cache.SecurityConfigCache{
		SecurityEnabled:     config.SecurityEnabled,
		BlacklistTTLMinutes: config.BlacklistTTLMinutes,
		LogRetentionDays:    config.LogRetentionDays,
	}
}
