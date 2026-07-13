package middleware

import (
	"strings"
	"time"

	"novablog/enum"
	"novablog/internal/cache"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware 限流中间件：基于令牌桶算法对请求进行频率控制，
// 超出限制时累计违规次数，达到阈值后自动将 IP 加入黑名单。
func RateLimitMiddleware(securityCache *cache.SecurityCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// 从缓存获取安全配置，缓存未命中时跳过限流（fail open）
		config, err := securityCache.GetConfig(c.Request.Context())
		if err != nil || config == nil {
			c.Next()
			return
		}

		// 根据请求方法和路径确定限流类别
		path := c.Request.URL.Path
		var maxTokens, windowSeconds int
		var route string
		if c.Request.Method == "POST" || strings.Contains(path, "comment") {
			maxTokens = config.PostMaxTokens
			windowSeconds = config.PostWindowSeconds
			route = "post"
		} else if strings.Contains(path, "view") {
			maxTokens = config.ViewMaxTokens
			windowSeconds = config.ViewWindowSeconds
			route = "view"
		} else if strings.Contains(path, "like") {
			maxTokens = config.LikeMaxTokens
			windowSeconds = config.LikeWindowSeconds
			route = "like"
		} else {
			maxTokens = config.GetMaxTokens
			windowSeconds = config.GetWindowSeconds
			route = "get"
		}

		// 检查是否允许请求
		allowed, err := securityCache.AllowRequest(c.Request.Context(), ip, route, maxTokens, windowSeconds)
		if err != nil {
			// 限流检查异常时放行，避免影响正常请求
			c.Next()
			return
		}
		if allowed {
			c.Next()
			return
		}

		// 请求被限流：递增违规计数
		violations, err := securityCache.IncrementViolation(c.Request.Context(), ip)
		if err == nil && violations >= int64(config.BlacklistThreshold) {
			// 违规次数达到阈值，将 IP 加入黑名单
			ttl := time.Duration(config.BlacklistTTLMinutes) * time.Minute
			_ = securityCache.AddToBlacklist(c.Request.Context(), ip, ttl)
			_ = securityCache.ResetViolations(c.Request.Context(), ip)
		}

		// 记录每日限流触发次数（用于统计）
		_ = securityCache.IncrementDailyRateLimitLog(c.Request.Context(), time.Now().Format("20060102"))

		response.Fail(c, enum.ErrTooManyRequests.Code, enum.ErrTooManyRequests.Msg, enum.ErrTooManyRequests.HttpCode)
		c.Abort()
	}
}
