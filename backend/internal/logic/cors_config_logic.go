package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/middleware"
	"novablog/internal/model"

	"gorm.io/gorm"
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

// GetConfig 获取跨域配置（优先读缓存，缓存未命中则查数据库并写入缓存；DB 无记录时返回空配置）。
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// DB 无记录：返回空配置且不写缓存（等待启动播种或首次保存）
			return &res.CorsConfigRes{}, nil
		}
		return nil, fmt.Errorf("查询跨域配置失败: %w", err)
	}

	// 写入缓存（失败不影响主流程）
	configCache := corsConfigToCache(config)
	_ = l.configCache.SetConfig(ctx, configCache, corsConfigCacheTTL)

	return corsConfigToRes(config), nil
}

// UpdateConfig 更新跨域配置（DB 无记录时创建），更新数据库后刷新缓存并同步到中间件内存变量。
func (l *CorsConfigLogic) UpdateConfig(ctx context.Context, r *req.UpdateCorsConfigReq) error {
	// 先获取当前配置
	config, err := l.configModel.GetConfig(ctx)
	switch {
	case err == nil:
		// 仅更新提供的字段
		if r.AllowedOrigins != nil {
			config.AllowedOrigins = *r.AllowedOrigins
		}
		// 更新数据库
		if err := l.configModel.UpdateConfig(ctx, config); err != nil {
			return fmt.Errorf("更新跨域配置失败: %w", err)
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		// DB 无记录：直接创建
		origins := ""
		if r.AllowedOrigins != nil {
			origins = *r.AllowedOrigins
		}
		if config, err = l.configModel.CreateConfig(ctx, origins); err != nil {
			return fmt.Errorf("创建跨域配置失败: %w", err)
		}
	default:
		return fmt.Errorf("查询跨域配置失败: %w", err)
	}

	// 直接写入新缓存，确保立即生效（热更新）
	configCache := corsConfigToCache(config)
	_ = l.configCache.SetConfig(ctx, configCache, corsConfigCacheTTL)

	// 同步到中间件内存变量，使 CORS 策略即时生效
	middleware.SetAllowedOrigins(config.AllowedOrigins)

	return nil
}

// InitFromFallback 应用启动引导：DB 是唯一权威数据源，任何其他来源都不能覆盖它。
//   - DB 存在用户自定义配置（记录更新时间晚于创建时间）→ 无条件以 DB 为准；
//   - DB 仅有从未修改过的初始记录（如迁移播种行）→ 视为无数据，用 fallback（config.yaml）回填并生效；
//   - DB 无记录 → 用 fallback 播种并生效；fallback 为空则不落库，中间件走硬编码默认值。
//
// 一旦 DB 落库有效数据，后续启动一律以 DB 为准，config.yaml 不再覆盖。
func (l *CorsConfigLogic) InitFromFallback(ctx context.Context, fallback string) error {
	config, err := l.configModel.GetConfig(ctx)
	switch {
	case err == nil:
		// 更新时间等于创建时间 → 记录自创建后从未被修改（迁移播种的初始行）
		if config.UpdatedAt.Equal(config.CreatedAt) && fallback != "" {
			config.AllowedOrigins = fallback
			if err := l.configModel.UpdateConfig(ctx, config); err != nil {
				return fmt.Errorf("回填跨域配置失败: %w", err)
			}
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		// DB 无数据：用 config.yaml 播种（为空则不落库）
		if fallback == "" {
			return nil
		}
		if config, err = l.configModel.CreateConfig(ctx, fallback); err != nil {
			return fmt.Errorf("播种跨域配置失败: %w", err)
		}
	default:
		return fmt.Errorf("查询跨域配置失败: %w", err)
	}

	// 刷新缓存并同步中间件
	_ = l.configCache.SetConfig(ctx, corsConfigToCache(config), corsConfigCacheTTL)
	middleware.SetAllowedOrigins(config.AllowedOrigins)

	return nil
}
