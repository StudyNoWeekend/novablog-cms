package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheKeyBiliParse B 站视频解析结果缓存 key 前缀。
const CacheKeyBiliParse = "bili:parse"

// BilibiliCache B 站解析结果的 Redis 缓存操作（存 JSON 字节，编解码由调用方负责）。
type BilibiliCache struct {
	client *redis.Client
}

// NewBilibiliCache 创建 BilibiliCache 实例。
func NewBilibiliCache() *BilibiliCache {
	return &BilibiliCache{client: RedisClient}
}

// Get 读取解析结果缓存；未命中或不可用时返回 nil（缓存缺失不视为错误）。
func (c *BilibiliCache) Get(ctx context.Context, bvid string) []byte {
	if c.client == nil || bvid == "" {
		return nil
	}
	data, err := c.client.Get(ctx, CacheKeyBiliParse+":"+bvid).Bytes()
	if err != nil || len(data) == 0 {
		return nil
	}
	return data
}

// Set 写入解析结果缓存，失败静默忽略（best-effort）。
func (c *BilibiliCache) Set(ctx context.Context, bvid string, data []byte, ttl time.Duration) {
	if c.client == nil || bvid == "" || len(data) == 0 {
		return
	}
	_ = c.client.Set(ctx, CacheKeyBiliParse+":"+bvid, data, ttl).Err()
}
