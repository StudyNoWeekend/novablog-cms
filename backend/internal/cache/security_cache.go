package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// CacheKeySecurityConfig 安全配置缓存键前缀
	CacheKeySecurityConfig = "security:config"
	// CacheKeyBlacklist IP 黑名单缓存键前缀
	CacheKeyBlacklist = "security:blacklist"
)

// SecurityConfigCache 安全配置缓存结构体，用于 Redis 缓存序列化。
type SecurityConfigCache struct {
	SecurityEnabled     bool `json:"security_enabled"`      // 是否开启安全防护
	BlacklistTTLMinutes int  `json:"blacklist_ttl_minutes"` // 黑名单封禁时长（分钟）
	LogRetentionDays    int  `json:"log_retention_days"`    // IP 访问日志保留天数
}

// SecurityCache 安全相关缓存操作结构体。
type SecurityCache struct {
	client *redis.Client
}

// NewSecurityCache 创建 SecurityCache 实例。
func NewSecurityCache() *SecurityCache {
	return &SecurityCache{client: RedisClient}
}

// GetConfig 从缓存获取安全配置。
func (c *SecurityCache) GetConfig(ctx context.Context) (*SecurityConfigCache, error) {
	data, err := c.client.Get(ctx, CacheKeySecurityConfig).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var config SecurityConfigCache
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetConfig 将安全配置写入缓存。
func (c *SecurityCache) SetConfig(ctx context.Context, config *SecurityConfigCache, ttl time.Duration) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, CacheKeySecurityConfig, data, ttl).Err()
}

// DeleteConfig 删除安全配置缓存（配置更新时调用）。
func (c *SecurityCache) DeleteConfig(ctx context.Context) error {
	return c.client.Del(ctx, CacheKeySecurityConfig).Err()
}

// IsBlacklisted 检查 IP 是否在黑名单中。
func (c *SecurityCache) IsBlacklisted(ctx context.Context, ip string) (bool, error) {
	key := fmt.Sprintf("%s:%s", CacheKeyBlacklist, ip)
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// AddToBlacklist 将 IP 加入黑名单，设置过期时间。
func (c *SecurityCache) AddToBlacklist(ctx context.Context, ip string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", CacheKeyBlacklist, ip)
	return c.client.Set(ctx, key, "1", ttl).Err()
}

// RemoveFromBlacklist 将 IP 从黑名单中移除。
func (c *SecurityCache) RemoveFromBlacklist(ctx context.Context, ip string) error {
	key := fmt.Sprintf("%s:%s", CacheKeyBlacklist, ip)
	return c.client.Del(ctx, key).Err()
}

// GetBlacklistCount 统计当前黑名单中的 IP 数量。
func (c *SecurityCache) GetBlacklistCount(ctx context.Context) (int64, error) {
	pattern := fmt.Sprintf("%s:*", CacheKeyBlacklist)
	var count int64
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		count++
	}
	if err := iter.Err(); err != nil {
		return 0, err
	}
	return count, nil
}
