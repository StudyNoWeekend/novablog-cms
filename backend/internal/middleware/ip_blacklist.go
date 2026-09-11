package middleware

import (
	"novablog/enum"
	"novablog/internal/cache"
	"novablog/internal/model"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IPBlacklistMiddleware IP 黑名单中间件：检查客户端 IP 是否已被加入黑名单，
// 若在黑名单中则返回 403 错误，否则放行请求。
// 当 Redis 或任何其他检查出现异常时，采用 fail-open 策略放行请求。
func IPBlacklistMiddleware(securityCache *cache.SecurityCache, securityModel *model.SecurityModel) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = c.RemoteIP()
		}

		blacklisted, err := securityCache.IsBlacklisted(c.Request.Context(), ip)
		if err != nil {
			// fail-open：记录错误并放行，避免影响正常请求
			if Logger != nil {
				Logger.Error("检查 IP 黑名单失败", zap.Error(err))
			}
			c.Next()
			return
		}

		if blacklisted {
			response.Fail(c, enum.ErrForbidden.Code, enum.ErrForbidden.Msg, enum.ErrForbidden.HttpCode)
			c.Abort()
			return
		}

		c.Next()
	}
}
