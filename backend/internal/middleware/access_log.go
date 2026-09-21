package middleware

import (
	"context"
	"time"

	"novablog/pkg/geoip"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"novablog/internal/model"
)

// Logger 中间件包级日志实例，由 bootstrap 注入。
var Logger *zap.Logger

// GeoIP IP 归属地解析器，由 bootstrap 注入；未注入时跳过地区解析。
var GeoIP *geoip.Searcher

// AccessLogMiddleware 记录每个 HTTP 请求的 IP 访问日志。
// 日志写入采用 fire-and-forget 模式，带 2 秒超时，避免阻塞响应。
// 地区归属在写入时解析一次并随日志持久化，聚合统计直接读库。
func AccessLogMiddleware() gin.HandlerFunc {
	accessLogModel := model.NewAccessLog()
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		ip := c.ClientIP()
		if ip == "" {
			ip = c.RemoteIP()
		}

		ua := c.Request.UserAgent()
		if len(ua) > 255 {
			ua = ua[:255]
		}

		log := &model.AccessLog{
			IP:         ip,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			UserAgent:  ua,
			CreatedAt:  start,
		}

		// 解析 IP 归属地区（数据未就绪时返回空，不影响日志写入）
		if GeoIP != nil {
			log.Region = GeoIP.Lookup(ip)
		}

		// 异步写入，带超时控制，不影响响应
		// 使用 context.Background() 而非请求 context，避免请求返回后 context 被取消导致写入失败
		go func(l *model.AccessLog) {
			writeCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := accessLogModel.Create(writeCtx, l); err != nil {
				// 使用 zap logger 如果可用，否则忽略错误避免阻塞
				if Logger != nil {
					Logger.Error("写入访问日志失败", zap.Error(err))
				}
			}
		}(log)
	}
}
