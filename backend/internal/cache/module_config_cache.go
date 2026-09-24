package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// CacheKeyModuleConfig 模块开关配置缓存键（v2：新增 open_source_enabled 字段后换键，避免旧缓存缺字段被解析为 false）
	CacheKeyModuleConfig = "module:config:v2"
)

// ModuleConfigCacheData 模块开关配置缓存数据结构体，用于 Redis 缓存序列化。
type ModuleConfigCacheData struct {
	ArticleEnabled    bool `json:"article_enabled"`
	MediaEnabled      bool `json:"media_enabled"`
	MusicEnabled      bool `json:"music_enabled"`
	VideoEnabled      bool `json:"video_enabled"`
	TravelEnabled     bool `json:"travel_enabled"`
	PortfolioEnabled  bool `json:"portfolio_enabled"`
	EquipmentEnabled  bool `json:"equipment_enabled"`
	ProjectEnabled    bool `json:"project_enabled"`
	OpenSourceEnabled bool `json:"open_source_enabled"`
}

// ModuleConfigCache 模块开关配置缓存操作结构体。
type ModuleConfigCache struct {
	client *redis.Client
}

// NewModuleConfigCache 创建 ModuleConfigCache 实例。
func NewModuleConfigCache() *ModuleConfigCache {
	return &ModuleConfigCache{client: RedisClient}
}

// GetConfig 从缓存获取模块开关配置。
func (c *ModuleConfigCache) GetConfig(ctx context.Context) (*ModuleConfigCacheData, error) {
	data, err := c.client.Get(ctx, CacheKeyModuleConfig).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var config ModuleConfigCacheData
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetConfig 将模块开关配置写入缓存。
func (c *ModuleConfigCache) SetConfig(ctx context.Context, config *ModuleConfigCacheData, ttl time.Duration) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, CacheKeyModuleConfig, data, ttl).Err()
}

// DeleteConfig 删除模块开关配置缓存（配置更新时调用）。
func (c *ModuleConfigCache) DeleteConfig(ctx context.Context) error {
	return c.client.Del(ctx, CacheKeyModuleConfig).Err()
}
