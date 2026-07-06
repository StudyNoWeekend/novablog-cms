package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware 请求日志中间件：记录请求方法、路径、状态码、耗时和客户端 IP。
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// 构建日志字段
		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("method", method),
			zap.String("path", path),
		}

		if query != "" {
			fields = append(fields, zap.String("query", query))
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			logger.Error("请求处理异常", fields...)
		} else if statusCode >= 400 {
			logger.Warn("客户端请求错误", fields...)
		} else {
			logger.Info("请求处理完成", fields...)
		}
	}
}
