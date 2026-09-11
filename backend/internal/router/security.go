package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterSecurityRoutes 注册安全管理相关路由。
func RegisterSecurityRoutes(r *gin.RouterGroup, securityController *controller.SecurityController, authMiddleware gin.HandlerFunc) {
	security := r.Group("/security")
	security.Use(authMiddleware)
	{
		security.GET("/config", securityController.GetConfig)
		security.PUT("/config", securityController.UpdateConfig)

		security.POST("/blacklist", securityController.CreateBlacklist)
		security.DELETE("/blacklist/:id", securityController.DeleteBlacklist)
		security.PUT("/blacklist/:id", securityController.UpdateBlacklist)
		security.GET("/blacklist/:id", securityController.GetBlacklist)
		security.GET("/blacklists", securityController.ListBlacklists)

		security.GET("/access-stats", securityController.GetAccessStatistics)
	}
}
