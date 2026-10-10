package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"novablog/internal/cache"

	"github.com/gin-gonic/gin"
)

// MCP 每密钥限流参数：每分钟最多 120 次请求。
const (
	mcpRateLimitWindow  = time.Minute
	mcpRateLimitMaxReqs = 120
)

// MCPRateLimitMiddleware MCP 端点限流中间件：按密钥 ID 做固定窗口限流。
// Redis 不可用时放行（鉴权已由 MCPAuthMiddleware 保证）。
func MCPRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := cache.RedisClient
		if client == nil {
			c.Next()
			return
		}
		keyID, _ := c.Get("mcp_key_id")
		if keyID == nil {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		key := fmt.Sprintf("ratelimit:mcp:%s:%d", keyID, time.Now().Unix()/int64(mcpRateLimitWindow.Seconds()))
		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			// 限流器故障不应阻断正常调用
			c.Next()
			return
		}
		if count == 1 {
			client.Expire(ctx, key, mcpRateLimitWindow)
		}
		if count > mcpRateLimitMaxReqs {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": 429001,
				"msg":  "MCP 请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
