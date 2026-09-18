package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// CORS 默认兜底值（仅在 环境变量、DB配置、config.yaml 均未设置时生效）
const defaultAllowedOrigins = "http://localhost:5173,http://localhost:5174"

// 包级变量：由 SetAllowedOrigins 写入，中间件从该变量读取（RWMutex 保护）
var (
	corsAllowedOrigins   = defaultAllowedOrigins
	corsAllowedOriginsMu sync.RWMutex
)

// SetAllowedOrigins 设置允许的来源白名单（由 logic.CorsConfigLogic 在更新配置时调用）。
// 设置的优先级：环境变量 CORS_ALLOWED_ORIGINS > 此变量 > 硬编码默认值。
func SetAllowedOrigins(origins string) {
	corsAllowedOriginsMu.Lock()
	corsAllowedOrigins = origins
	corsAllowedOriginsMu.Unlock()
}

// getEffectiveOrigins 获取实际生效的来源白名单（按优先级）。
func getEffectiveOrigins() string {
	// 优先级 1：环境变量
	if env := os.Getenv("CORS_ALLOWED_ORIGINS"); env != "" {
		return env
	}
	// 优先级 2：包级变量（由后台页面配置写入）
	corsAllowedOriginsMu.RLock()
	val := corsAllowedOrigins
	corsAllowedOriginsMu.RUnlock()
	if val != "" {
		return val
	}
	// 优先级 3：硬编码默认值
	return defaultAllowedOrigins
}

// CORSMiddleware 跨域资源共享中间件（基于白名单校验来源）。
// 白名单来源按优先级读取：环境变量 → 后台页面配置 → 默认值。
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := getEffectiveOrigins()

		// 校验请求来源是否在白名单中
		allowed := false
		for _, o := range strings.Split(allowedOrigins, ",") {
			if strings.TrimSpace(o) == origin {
				allowed = true
				break
			}
		}

		// 设置 CORS 响应头
		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Market-Base-URL, X-Market-Token")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type, X-Trace-Id")
		c.Header("Access-Control-Max-Age", "86400")

		// 处理 OPTIONS 预检请求
		if c.Request.Method == http.MethodOptions {
			// 如果 origin 不在白名单中，返回 403 而不是 204
			if !allowed && origin != "" {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
