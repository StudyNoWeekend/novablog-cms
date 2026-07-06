// Package cache 提供 Redis 缓存操作。
package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient 全局 Redis 客户端实例。
var RedisClient *redis.Client

// TokenCache Token 缓存操作结构体。
type TokenCache struct {
	client *redis.Client
}

// NewTokenCache 创建 TokenCache 实例。
func NewTokenCache() *TokenCache {
	return &TokenCache{client: RedisClient}
}

// AddToBlacklist 将 Token 的 JTI 加入黑名单，设置过期时间。
func (c *TokenCache) AddToBlacklist(ctx context.Context, tokenJTI string, expiration time.Duration) error {
	return c.client.Set(ctx, "blacklist:"+tokenJTI, "1", expiration).Err()
}

// IsBlacklisted 检查 Token 的 JTI 是否在黑名单中。
func (c *TokenCache) IsBlacklisted(ctx context.Context, tokenJTI string) (bool, error) {
	exists, err := c.client.Exists(ctx, "blacklist:"+tokenJTI).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
