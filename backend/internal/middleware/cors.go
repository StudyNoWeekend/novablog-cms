package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// CORS 默认兜底值（仅当 DB 与 config.yaml 均未提供有效白名单时生效）
const defaultAllowedOrigins = "http://localhost:5173,http://localhost:5174"

// 包级变量：由 SetAllowedOrigins 写入，中间件从该变量读取（RWMutex 保护）。
// DB 是唯一权威数据源：启动时由 bootstrap 从 DB 读取写入（DB 无数据时以
// config.yaml 播种），运行期由后台保存配置（CorsConfigLogic.UpdateConfig）热更新。
var (
	corsAllowedOrigins   = defaultAllowedOrigins
	corsAllowedOriginsMu sync.RWMutex
)

// SetAllowedOrigins 设置允许的来源白名单（启动引导与后台保存配置时调用）。
func SetAllowedOrigins(origins string) {
	corsAllowedOriginsMu.Lock()
	corsAllowedOrigins = origins
	corsAllowedOriginsMu.Unlock()
}

// getEffectiveOrigins 获取实际生效的来源白名单，空值时回退硬编码默认值。
func getEffectiveOrigins() string {
	corsAllowedOriginsMu.RLock()
	val := corsAllowedOrigins
	corsAllowedOriginsMu.RUnlock()
	if val != "" {
		return val
	}
	return defaultAllowedOrigins
}

// CORSMiddleware 跨域资源共享中间件（基于白名单校验来源）。
// 白名单唯一来源为 SetAllowedOrigins 写入的包级变量（DB 权威，config.yaml 播种兜底）。
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
