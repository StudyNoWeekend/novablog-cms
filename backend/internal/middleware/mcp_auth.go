package middleware

import (
	"context"
	"strings"

	"novablog/enum"
	"novablog/internal/model"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MCPKeyVerifier MCP 密钥校验接口，由 logic 层实现，避免中间件反向依赖业务层。
type MCPKeyVerifier interface {
	VerifyKey(ctx context.Context, key string) (*model.MCPAPIKey, error)
	RecordUsage(ctx context.Context, id string)
}

// MCPAuthMiddleware MCP 端点鉴权中间件：校验 Authorization: Bearer <MCP Key>，
// 通过后将 mcp_key_id / mcp_key_name 写入 gin Context 并记录使用统计。
func MCPAuthMiddleware(verifier MCPKeyVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			failMCPAuth(c, enum.ErrMCPKeyInvalid)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			failMCPAuth(c, enum.ErrMCPKeyInvalid)
			return
		}

		record, err := verifier.VerifyKey(c.Request.Context(), strings.TrimSpace(parts[1]))
		if err != nil {
			if AuthLogger != nil {
				AuthLogger.Warn("MCP 密钥校验失败",
					zap.String("ip", c.ClientIP()),
					zap.String("error", err.Error()),
				)
			}
			var bizErr *enum.BizError
			if e, ok := err.(*enum.BizError); ok {
				bizErr = e
			} else {
				bizErr = enum.ErrMCPKeyInvalid
			}
			failMCPAuth(c, bizErr)
			return
		}

		c.Set("mcp_key_id", record.ID)
		c.Set("mcp_key_name", record.Name)

		// 记录使用统计（失败不影响请求）
		verifier.RecordUsage(c.Request.Context(), record.ID)
		c.Next()
	}
}

// failMCPAuth 返回 MCP 鉴权失败响应，并附带 WWW-Authenticate 头供客户端识别。
func failMCPAuth(c *gin.Context, bizErr *enum.BizError) {
	c.Header("WWW-Authenticate", `Bearer realm="novablog-mcp"`)
	response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
	c.Abort()
}
