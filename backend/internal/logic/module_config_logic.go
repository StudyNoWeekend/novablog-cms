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

// moduleKeySetters 模块开关 key 白名单到模型字段的映射。
// 首装角色预设与模块配置更新共用，新增模块开关时同步维护。
var moduleKeySetters = map[string]func(c *model.ModuleConfig, v bool){
	"article_enabled":     func(c *model.ModuleConfig, v bool) { c.ArticleEnabled = v },
	"media_enabled":       func(c *model.ModuleConfig, v bool) { c.MediaEnabled = v },
	"music_enabled":       func(c *model.ModuleConfig, v bool) { c.MusicEnabled = v },
	"video_enabled":       func(c *model.ModuleConfig, v bool) { c.VideoEnabled = v },
	"travel_enabled":      func(c *model.ModuleConfig, v bool) { c.TravelEnabled = v },
	"portfolio_enabled":   func(c *model.ModuleConfig, v bool) { c.PortfolioEnabled = v },
	"equipment_enabled":   func(c *model.ModuleConfig, v bool) { c.EquipmentEnabled = v },
	"project_enabled":     func(c *model.ModuleConfig, v bool) { c.ProjectEnabled = v },
	"open_source_enabled": func(c *model.ModuleConfig, v bool) { c.OpenSourceEnabled = v },
	"recipe_enabled":      func(c *model.ModuleConfig, v bool) { c.RecipeEnabled = v },
	"book_enabled":        func(c *model.ModuleConfig, v bool) { c.BookEnabled = v },
	"game_enabled":        func(c *model.ModuleConfig, v bool) { c.GameEnabled = v },
	"fitness_enabled":     func(c *model.ModuleConfig, v bool) { c.FitnessEnabled = v },
	"tech_stack_enabled":  func(c *model.ModuleConfig, v bool) { c.TechStackEnabled = v },
}

// IsValidModuleKey 判断模块开关 key 是否在白名单内。
func IsValidModuleKey(key string) bool {
	_, ok := moduleKeySetters[key]
	return ok
}

// modelToCache 将数据库模型转换为缓存数据结构。
func moduleConfigToCache(m *model.ModuleConfig) *cache.ModuleConfigCacheData {
	return &cache.ModuleConfigCacheData{
		ArticleEnabled:    m.ArticleEnabled,
		MediaEnabled:      m.MediaEnabled,
		MusicEnabled:      m.MusicEnabled,
		VideoEnabled:      m.VideoEnabled,
		TravelEnabled:     m.TravelEnabled,
		PortfolioEnabled:  m.PortfolioEnabled,
		EquipmentEnabled:  m.EquipmentEnabled,
		ProjectEnabled:    m.ProjectEnabled,
		OpenSourceEnabled: m.OpenSourceEnabled,
		RecipeEnabled:     m.RecipeEnabled,
		BookEnabled:       m.BookEnabled,
		GameEnabled:       m.GameEnabled,
		FitnessEnabled:    m.FitnessEnabled,
		TechStackEnabled:  m.TechStackEnabled,
	}
}

// cacheToRes 将缓存数据转换为响应结构体。
func moduleConfigCacheToRes(c *cache.ModuleConfigCacheData) *res.ModuleConfigRes {
	return &res.ModuleConfigRes{
		ArticleEnabled:    c.ArticleEnabled,
		MediaEnabled:      c.MediaEnabled,
		MusicEnabled:      c.MusicEnabled,
		VideoEnabled:      c.VideoEnabled,
		TravelEnabled:     c.TravelEnabled,
		PortfolioEnabled:  c.PortfolioEnabled,
		EquipmentEnabled:  c.EquipmentEnabled,
		ProjectEnabled:    c.ProjectEnabled,
		OpenSourceEnabled: c.OpenSourceEnabled,
		RecipeEnabled:     c.RecipeEnabled,
		BookEnabled:       c.BookEnabled,
		GameEnabled:       c.GameEnabled,
		FitnessEnabled:    c.FitnessEnabled,
		TechStackEnabled:  c.TechStackEnabled,
	}
}

// modelToRes 将数据库模型直接转换为响应结构体。
func moduleConfigToRes(m *model.ModuleConfig) *res.ModuleConfigRes {
	return &res.ModuleConfigRes{
		ArticleEnabled:    m.ArticleEnabled,
		MediaEnabled:      m.MediaEnabled,
		MusicEnabled:      m.MusicEnabled,
		VideoEnabled:      m.VideoEnabled,
		TravelEnabled:     m.TravelEnabled,
		PortfolioEnabled:  m.PortfolioEnabled,
		EquipmentEnabled:  m.EquipmentEnabled,
		ProjectEnabled:    m.ProjectEnabled,
		OpenSourceEnabled: m.OpenSourceEnabled,
		RecipeEnabled:     m.RecipeEnabled,
		BookEnabled:       m.BookEnabled,
		GameEnabled:       m.GameEnabled,
		FitnessEnabled:    m.FitnessEnabled,
		TechStackEnabled:  m.TechStackEnabled,
		UpdatedAt:         m.UpdatedAt,
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
	if r.OpenSourceEnabled != nil {
		config.OpenSourceEnabled = *r.OpenSourceEnabled
	}
	if r.RecipeEnabled != nil {
		config.RecipeEnabled = *r.RecipeEnabled
	}
	if r.BookEnabled != nil {
		config.BookEnabled = *r.BookEnabled
	}
	if r.GameEnabled != nil {
		config.GameEnabled = *r.GameEnabled
	}
	if r.FitnessEnabled != nil {
		config.FitnessEnabled = *r.FitnessEnabled
	}
	if r.TechStackEnabled != nil {
		config.TechStackEnabled = *r.TechStackEnabled
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

// ApplyModuleMap 按模块开关 key 集合覆盖式写模块配置（首装角色预设用）。
// map 中出现的 key 会被置为对应值，未出现的 key 保持当前配置不变；
// key 不在白名单时返回错误。写入后热更新缓存。
func (l *ModuleConfigLogic) ApplyModuleMap(ctx context.Context, modules map[string]bool) error {
	if len(modules) == 0 {
		return nil
	}
	for key := range modules {
		if !IsValidModuleKey(key) {
			return fmt.Errorf("未知的模块开关: %s", key)
		}
	}

	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("查询模块开关配置失败: %w", err)
	}

	for key, value := range modules {
		moduleKeySetters[key](config, value)
	}

	if err := l.configModel.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("更新模块开关配置失败: %w", err)
	}

	configCache := moduleConfigToCache(config)
	_ = l.configCache.SetConfig(ctx, configCache, moduleConfigCacheTTL)

	return nil
}
