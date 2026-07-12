package router

import (
	"cus-cms/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterSecurityRoutes 注册安全管理相关路由。
func RegisterSecurityRoutes(r *gin.RouterGroup, securityController *controller.SecurityController, authMiddleware gin.HandlerFunc) {
	security := r.Group("/security")
	security.Use(authMiddleware)
	{
		security.GET("/config", securityController.GetConfig)
		security.PUT("/config", securityController.UpdateConfig)
		security.GET("/blacklist", securityController.GetBlacklist)
		security.DELETE("/blacklist/:ip", securityController.UnbanIP)
		security.GET("/stats", securityController.GetStats)
	}
}
