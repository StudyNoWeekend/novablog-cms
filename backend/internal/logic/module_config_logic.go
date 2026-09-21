package logic

import (
	"context"
	"fmt"
	"time"

	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
)

// moduleConfigCacheTTL 模块开关配置缓存过期时间。
const moduleConfigCacheTTL = 10 * time.Minute

// ModuleConfigLogic 模块开关配置业务逻辑结构体。
type ModuleConfigLogic struct {
	configModel *model.ModuleConfigModel
	configCache *cache.ModuleConfigCache
}

// NewModuleConfigLogic 创建 ModuleConfigLogic 实例。
func NewModuleConfigLogic() *ModuleConfigLogic {
	return &ModuleConfigLogic{
		configModel: model.NewModuleConfig(),
		configCache: cache.NewModuleConfigCache(),
	}
}

// modelToCache 将数据库模型转换为缓存数据结构。
func moduleConfigToCache(m *model.ModuleConfig) *cache.ModuleConfigCacheData {
	return &cache.ModuleConfigCacheData{
		ArticleEnabled:   m.ArticleEnabled,
		MediaEnabled:     m.MediaEnabled,
		MusicEnabled:     m.MusicEnabled,
		VideoEnabled:     m.VideoEnabled,
		TravelEnabled:    m.TravelEnabled,
		PortfolioEnabled: m.PortfolioEnabled,
		EquipmentEnabled: m.EquipmentEnabled,
		ProjectEnabled:   m.ProjectEnabled,
	}
}

// cacheToRes 将缓存数据转换为响应结构体。
func moduleConfigCacheToRes(c *cache.ModuleConfigCacheData) *res.ModuleConfigRes {
	return &res.ModuleConfigRes{
		ArticleEnabled:   c.ArticleEnabled,
		MediaEnabled:     c.MediaEnabled,
		MusicEnabled:     c.MusicEnabled,
		VideoEnabled:     c.VideoEnabled,
		TravelEnabled:    c.TravelEnabled,
		PortfolioEnabled: c.PortfolioEnabled,
		EquipmentEnabled: c.EquipmentEnabled,
		ProjectEnabled:   c.ProjectEnabled,
	}
}

// modelToRes 将数据库模型直接转换为响应结构体。
func moduleConfigToRes(m *model.ModuleConfig) *res.ModuleConfigRes {
	return &res.ModuleConfigRes{
		ArticleEnabled:   m.ArticleEnabled,
		MediaEnabled:     m.MediaEnabled,
		MusicEnabled:     m.MusicEnabled,
		VideoEnabled:     m.VideoEnabled,
		TravelEnabled:    m.TravelEnabled,
		PortfolioEnabled: m.PortfolioEnabled,
		EquipmentEnabled: m.EquipmentEnabled,
		ProjectEnabled:   m.ProjectEnabled,
		UpdatedAt:        m.UpdatedAt,
	}
}

// GetConfig 获取模块开关配置（优先读缓存，缓存未命中则查数据库并写入缓存）。
func (l *ModuleConfigLogic) GetConfig(ctx context.Context) (*res.ModuleConfigRes, error) {
	// 优先读取缓存
	cached, err := l.configCache.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取模块开关配置缓存失败: %w", err)
	}
	if cached != nil {
		return moduleConfigCacheToRes(cached), nil
	}

	// 缓存未命中，查询数据库
	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询模块开关配置失败: %w", err)
	}

	// 写入缓存（失败不影响主流程）
	configCache := moduleConfigToCache(config)
	_ = l.configCache.SetConfig(ctx, configCache, moduleConfigCacheTTL)

	return moduleConfigToRes(config), nil
}

// UpdateConfig 更新模块开关配置，更新数据库后刷新缓存以确保立即生效。
func (l *ModuleConfigLogic) UpdateConfig(ctx context.Context, r *req.UpdateModuleConfigReq) error {
	// 先获取当前配置
	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("查询模块开关配置失败: %w", err)
	}

	// 仅更新提供的字段
	if r.ArticleEnabled != nil {
		config.ArticleEnabled = *r.ArticleEnabled
	}
	if r.MediaEnabled != nil {
		config.MediaEnabled = *r.MediaEnabled
	}
	if r.MusicEnabled != nil {
		config.MusicEnabled = *r.MusicEnabled
	}
	if r.VideoEnabled != nil {
		config.VideoEnabled = *r.VideoEnabled
	}
	if r.TravelEnabled != nil {
		config.TravelEnabled = *r.TravelEnabled
	}
	if r.PortfolioEnabled != nil {
		config.PortfolioEnabled = *r.PortfolioEnabled
	}
	if r.EquipmentEnabled != nil {
		config.EquipmentEnabled = *r.EquipmentEnabled
	}
	if r.ProjectEnabled != nil {
		config.ProjectEnabled = *r.ProjectEnabled
	}

	// 更新数据库
	if err := l.configModel.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("更新模块开关配置失败: %w", err)
	}

	// 直接写入新缓存，确保立即生效（热更新）
	configCache := moduleConfigToCache(config)
	_ = l.configCache.SetConfig(ctx, configCache, moduleConfigCacheTTL)

	return nil
}
