// Package middleware 定义 HTTP 中间件。
package middleware

import (
	"net/http"
	"strings"

	"cus-cms/enum"
	"cus-cms/internal/cache"
	"cus-cms/utils/jwt"
	"cus-cms/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthLogger 认证中间件日志记录器（由启动时注入）。
var AuthLogger *zap.Logger

// AuthMiddleware 认证中间件：解析 Bearer Token，验证有效性，检查黑名单。
func AuthMiddleware(accessSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取 Bearer Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, enum.ErrUnauthorized.Code, enum.ErrUnauthorized.Msg, enum.ErrUnauthorized.HttpCode)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Fail(c, enum.ErrTokenInvalid.Code, enum.ErrTokenInvalid.Msg, enum.ErrTokenInvalid.HttpCode)
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 解析 Token
		claims, err := jwt.ParseToken(tokenString, accessSecret)
		if err != nil {
			AuthLogger.Warn("Token 解析失败", zap.Error(err))
			var httpCode int
			var msg string
			var code int
			if strings.Contains(err.Error(), "expired") {
				code = enum.ErrTokenExpired.Code
				msg = enum.ErrTokenExpired.Msg
				httpCode = enum.ErrTokenExpired.HttpCode
			} else {
				code = enum.ErrTokenInvalid.Code
				msg = enum.ErrTokenInvalid.Msg
				httpCode = enum.ErrTokenInvalid.HttpCode
			}
			response.Fail(c, code, msg, httpCode)
			c.Abort()
			return
		}

		// 检查 Token 是否在黑名单中
		tokenCache := cache.NewTokenCache()
		blacklisted, err := tokenCache.IsBlacklisted(c.Request.Context(), claims.ID)
		if err != nil {
			AuthLogger.Error("检查令牌黑名单失败", zap.Error(err))
		}
		if blacklisted {
			AuthLogger.Warn("Token 已在黑名单中", zap.String("jti", claims.ID))
			response.Fail(c, enum.ErrTokenInvalid.Code, enum.ErrTokenInvalid.Msg, enum.ErrTokenInvalid.HttpCode)
			c.Abort()
			return
		}

		// 将用户信息存入 Context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// 确保导入了 http 包
var _ = http.StatusOK
