package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheKeyViewDedup 内容浏览量 IP 去重 key 前缀。
const CacheKeyViewDedup = "view"

// ViewStatsCache 内容浏览去重的 Redis 操作（SetNX 短窗口去重）。
type ViewStatsCache struct {
	client *redis.Client
}

// NewViewStatsCache 创建 ViewStatsCache 实例。
func NewViewStatsCache() *ViewStatsCache {
	return &ViewStatsCache{client: RedisClient}
}

// MarkViewed 窗口内首次出现的 (contentType, contentID, ip) 返回 true（应计数）。
// Redis 未初始化或操作失败时返回 true（fail-open：降级为直接计数）。
func (c *ViewStatsCache) MarkViewed(ctx context.Context, contentType, contentID, ip string, ttl time.Duration) bool {
	if c.client == nil || ip == "" {
		return true
	}
	ok, err := c.client.SetNX(ctx,
		CacheKeyViewDedup+":"+contentType+":"+contentID+":"+ip, "1", ttl).Result()
	if err != nil {
		return true
	}
	return ok
}
