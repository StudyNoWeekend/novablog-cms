package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterMCPKeyRoutes 注册 MCP 密钥管理路由（管理后台，需登录）。
func RegisterMCPKeyRoutes(r *gin.RouterGroup, mcpKeyController *controller.MCPKeyController, authMiddleware gin.HandlerFunc) {
	keys := r.Group("/mcp/keys")
	keys.Use(authMiddleware)
	{
		keys.POST("", mcpKeyController.Create)
		keys.GET("", mcpKeyController.GetList)
		keys.PUT("/:id", mcpKeyController.Update)
		keys.DELETE("/:id", mcpKeyController.Delete)
	}
}
