package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"cus-cms/enum"
	"cus-cms/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware 统一异常恢复中间件：捕获 panic，记录堆栈，返回 500 错误。
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				stack := string(debug.Stack())
				logger.Error("捕获到 panic",
					zap.Any("panic", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("stack", stack),
				)

				// 返回 500 错误
				response.Fail(c, enum.ErrInternalServer.Code,
					fmt.Sprintf("%s", err),
					http.StatusInternalServerError)

				c.Abort()
			}
		}()
		c.Next()
	}
}
