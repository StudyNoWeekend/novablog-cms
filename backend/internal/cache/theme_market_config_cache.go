package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// CacheKeyThemeMarketConfig 官方主题市场配置缓存键
	CacheKeyThemeMarketConfig = "theme:market:config"
)

// ThemeMarketConfigCacheData 主题模块运行配置缓存数据结构体，用于 Redis 缓存序列化。
type ThemeMarketConfigCacheData struct {
	MarketBaseURL string    `json:"market_base_url"`
	PublicAPIBase string    `json:"public_api_base"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ThemeMarketConfigCache 官方主题市场配置缓存操作结构体。
type ThemeMarketConfigCache struct {
	client *redis.Client
}

// NewThemeMarketConfigCache 创建 ThemeMarketConfigCache 实例。
func NewThemeMarketConfigCache() *ThemeMarketConfigCache {
	return &ThemeMarketConfigCache{client: RedisClient}
}

// GetConfig 从缓存获取官方市场配置。
func (c *ThemeMarketConfigCache) GetConfig(ctx context.Context) (*ThemeMarketConfigCacheData, error) {
	data, err := c.client.Get(ctx, CacheKeyThemeMarketConfig).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var config ThemeMarketConfigCacheData
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetConfig 将官方市场配置写入缓存。
func (c *ThemeMarketConfigCache) SetConfig(ctx context.Context, config *ThemeMarketConfigCacheData, ttl time.Duration) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, CacheKeyThemeMarketConfig, data, ttl).Err()
}
