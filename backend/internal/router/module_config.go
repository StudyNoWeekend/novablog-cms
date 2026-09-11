package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterModuleConfigRoutes 注册模块开关配置管理路由。
func RegisterModuleConfigRoutes(r *gin.RouterGroup, moduleConfigController *controller.ModuleConfigController, authMiddleware gin.HandlerFunc) {
	module := r.Group("/module-config")
	module.Use(authMiddleware)
	{
		module.GET("", moduleConfigController.GetConfig)
		module.PUT("", moduleConfigController.UpdateConfig)
	}
}
