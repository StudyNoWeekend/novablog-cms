package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterMapConfigRoutes 注册地图服务配置管理路由。
func RegisterMapConfigRoutes(r *gin.RouterGroup, mapConfigController *controller.MapConfigController, authMiddleware gin.HandlerFunc) {
	mapConfig := r.Group("/map-config")
	mapConfig.Use(authMiddleware)
	{
		mapConfig.GET("", mapConfigController.GetConfig)
		mapConfig.PUT("", mapConfigController.UpdateConfig)
	}
}
