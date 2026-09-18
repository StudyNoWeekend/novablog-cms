package logic

import (
	"context"
	"fmt"
	"time"

	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/middleware"
	"novablog/internal/model"
)

// corsConfigCacheTTL 跨域配置缓存过期时间。
const corsConfigCacheTTL = 10 * time.Minute

// CorsConfigLogic 跨域配置业务逻辑结构体。
type CorsConfigLogic struct {
	configModel *model.CorsConfigModel
	configCache *cache.CorsConfigCache
}

// NewCorsConfigLogic 创建 CorsConfigLogic 实例。
func NewCorsConfigLogic() *CorsConfigLogic {
	return &CorsConfigLogic{
		configModel: model.NewCorsConfig(),
		configCache: cache.NewCorsConfigCache(),
	}
}

// modelToCache 将数据库模型转换为缓存数据结构。
func corsConfigToCache(m *model.CorsConfig) *cache.CorsConfigCacheData {
	return &cache.CorsConfigCacheData{
		AllowedOrigins: m.AllowedOrigins,
	}
}

// cacheToRes 将缓存数据转换为响应结构体。
func corsConfigCacheToRes(c *cache.CorsConfigCacheData) *res.CorsConfigRes {
	return &res.CorsConfigRes{
		AllowedOrigins: c.AllowedOrigins,
	}
}

// modelToRes 将数据库模型直接转换为响应结构体。
func corsConfigToRes(m *model.CorsConfig) *res.CorsConfigRes {
	return &res.CorsConfigRes{
		AllowedOrigins: m.AllowedOrigins,
		UpdatedAt:      m.UpdatedAt,
	}
}

// GetConfig 获取跨域配置（优先读缓存，缓存未命中则查数据库并写入缓存）。
func (l *CorsConfigLogic) GetConfig(ctx context.Context) (*res.CorsConfigRes, error) {
	// 优先读取缓存
	cached, err := l.configCache.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取跨域配置缓存失败: %w", err)
	}
	if cached != nil {
		return corsConfigCacheToRes(cached), nil
	}

	// 缓存未命中，查询数据库
	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询跨域配置失败: %w", err)
	}

	// 写入缓存（失败不影响主流程）
	configCache := corsConfigToCache(config)
	_ = l.configCache.SetConfig(ctx, configCache, corsConfigCacheTTL)

	return corsConfigToRes(config), nil
}

// UpdateConfig 更新跨域配置，更新数据库后刷新缓存并同步到中间件内存变量。
func (l *CorsConfigLogic) UpdateConfig(ctx context.Context, r *req.UpdateCorsConfigReq) error {
	// 先获取当前配置
	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("查询跨域配置失败: %w", err)
	}

	// 仅更新提供的字段
	if r.AllowedOrigins != nil {
		config.AllowedOrigins = *r.AllowedOrigins
	}

	// 更新数据库
	if err := l.configModel.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("更新跨域配置失败: %w", err)
	}

	// 直接写入新缓存，确保立即生效（热更新）
	configCache := corsConfigToCache(config)
	_ = l.configCache.SetConfig(ctx, configCache, corsConfigCacheTTL)

	// 同步到中间件内存变量，使 CORS 策略即时生效
	middleware.SetAllowedOrigins(config.AllowedOrigins)

	return nil
}
