package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheKeyMCPRateLimit MCP 按密钥限流 key 前缀（固定窗口：ratelimit:mcp:<keyID>:<窗口序号>）。
const CacheKeyMCPRateLimit = "ratelimit:mcp"

// MCPRateLimitCache MCP 按密钥固定窗口限流的 Redis 计数操作。
type MCPRateLimitCache struct {
	client *redis.Client
}

// NewMCPRateLimitCache 创建 MCPRateLimitCache 实例。
func NewMCPRateLimitCache() *MCPRateLimitCache {
	return &MCPRateLimitCache{client: RedisClient}
}

// Hit 记录一次请求并返回窗口内累计计数；Redis 未初始化、keyID 为空或操作失败时
// 返回 available=false（fail-open：限流器故障不阻断调用）。
func (c *MCPRateLimitCache) Hit(ctx context.Context, keyID string, window time.Duration) (count int64, available bool) {
	if c.client == nil || keyID == "" {
		return 0, false
	}
	key := fmt.Sprintf("%s:%s:%d", CacheKeyMCPRateLimit, keyID, time.Now().Unix()/int64(window.Seconds()))
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, false
	}
	if count == 1 {
		c.client.Expire(ctx, key, window)
	}
	return count, true
}
