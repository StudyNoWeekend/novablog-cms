package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheKeyCloudDriveState 云盘 OAuth 一次性 state 的 key 前缀。
const CacheKeyCloudDriveState = "clouddrive:state"

// ErrRedisUnavailable Redis 客户端未初始化（bootstrap 未完成或被测试置空）。
var ErrRedisUnavailable = errors.New("Redis 客户端未初始化")

// CloudDriveCache 云盘 OAuth 授权 state 的 Redis 操作。
// state 为防 CSRF 的一次性凭据，语义 fail-closed：Redis 不可用时返回错误拒绝授权流程。
type CloudDriveCache struct {
	client *redis.Client
}

// NewCloudDriveCache 创建 CloudDriveCache 实例。
func NewCloudDriveCache() *CloudDriveCache {
	return &CloudDriveCache{client: RedisClient}
}

// SetState 存储 state → provider 映射，有效期 ttl。
func (c *CloudDriveCache) SetState(ctx context.Context, state, provider string, ttl time.Duration) error {
	if c.client == nil {
		return ErrRedisUnavailable
	}
	return c.client.Set(ctx, CacheKeyCloudDriveState+state, provider, ttl).Err()
}

// GetState 读取 state 对应的 provider；未命中返回空串与 redis.Nil 错误。
func (c *CloudDriveCache) GetState(ctx context.Context, state string) (string, error) {
	if c.client == nil {
		return "", ErrRedisUnavailable
	}
	return c.client.Get(ctx, CacheKeyCloudDriveState+state).Result()
}

// DelState 消费（删除）state。
func (c *CloudDriveCache) DelState(ctx context.Context, state string) {
	if c.client == nil {
		return
	}
	c.client.Del(ctx, CacheKeyCloudDriveState+state)
}
