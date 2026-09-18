package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// CacheKeyCorsConfig 跨域配置缓存键
	CacheKeyCorsConfig = "cors:config"
)

// CorsConfigCacheData 跨域配置缓存数据结构体，用于 Redis 缓存序列化。
type CorsConfigCacheData struct {
	AllowedOrigins string `json:"allowed_origins"`
}

// CorsConfigCache 跨域配置缓存操作结构体。
type CorsConfigCache struct {
	client *redis.Client
}

// NewCorsConfigCache 创建 CorsConfigCache 实例。
func NewCorsConfigCache() *CorsConfigCache {
	return &CorsConfigCache{client: RedisClient}
}

// GetConfig 从缓存获取跨域配置。
func (c *CorsConfigCache) GetConfig(ctx context.Context) (*CorsConfigCacheData, error) {
	data, err := c.client.Get(ctx, CacheKeyCorsConfig).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var config CorsConfigCacheData
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetConfig 将跨域配置写入缓存。
func (c *CorsConfigCache) SetConfig(ctx context.Context, config *CorsConfigCacheData, ttl time.Duration) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, CacheKeyCorsConfig, data, ttl).Err()
}

// DeleteConfig 删除跨域配置缓存（配置更新时调用）。
func (c *CorsConfigCache) DeleteConfig(ctx context.Context) error {
	return c.client.Del(ctx, CacheKeyCorsConfig).Err()
}
