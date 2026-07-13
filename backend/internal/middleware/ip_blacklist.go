package middleware

import (
	"novablog/enum"
	"novablog/internal/cache"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// IPBlacklistMiddleware IP 黑名单中间件：检查客户端 IP 是否已被加入黑名单，
// 若在黑名单中则返回 403 错误，否则放行请求。
func IPBlacklistMiddleware(securityCache *cache.SecurityCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		blocked, err := securityCache.IsBlacklisted(c.Request.Context(), ip)
		if err != nil {
			// 检查异常时放行，避免影响正常请求
			c.Next()
			return
		}
		if blocked {
			response.Fail(c, enum.ErrIPBlocked.Code, enum.ErrIPBlocked.Msg, enum.ErrIPBlocked.HttpCode)
			c.Abort()
			return
		}
		c.Next()
	}
}
