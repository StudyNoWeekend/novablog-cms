package router

import (
	"novablog/internal/logic"
	"novablog/internal/mcp"
	"novablog/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterMCPRoutes 注册 MCP Streamable HTTP 端点（外部 AI 客户端调用，MCP Key 鉴权）。
// 端点路径 /api/v1/mcp，按协议接受 POST（JSON-RPC 请求）、GET（SSE 流）、DELETE（终止会话）。
func RegisterMCPRoutes(r *gin.RouterGroup, svc *mcp.Service) {
	g := r.Group("/mcp")
	g.Use(middleware.MCPAuthMiddleware(logic.NewMCPKeyLogic()))
	g.Use(middleware.MCPRateLimitMiddleware())

	// 将鉴权中间件写入 gin Context 的密钥 ID 注入请求上下文，供 MCP 工具 handler 读取
	g.Use(func(c *gin.Context) {
		if id, ok := c.Get("mcp_key_id"); ok {
			if keyID, ok := id.(string); ok {
				c.Request = c.Request.WithContext(mcp.WithKeyID(c.Request.Context(), keyID))
			}
		}
		c.Next()
	})

	handler := gin.WrapH(svc.Handler())
	g.POST("", handler)
	g.GET("", handler)
	g.DELETE("", handler)
}
