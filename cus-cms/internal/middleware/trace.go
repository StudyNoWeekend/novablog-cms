package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceMiddleware 链路追踪中间件：为每个请求生成唯一 TraceID 并存入 Context 和响应头。
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先使用请求头中的 TraceID
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 存入 Context
		c.Set("trace_id", traceID)

		// 设置响应头
		c.Header("X-Trace-Id", traceID)

		c.Next()
	}
}
