package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterThemeMarketConfigRoutes 注册官方主题市场配置管理路由。
func RegisterThemeMarketConfigRoutes(r *gin.RouterGroup, configController *controller.ThemeMarketConfigController, authMiddleware gin.HandlerFunc) {
	config := r.Group("/theme-market-config")
	config.Use(authMiddleware)
	{
		config.GET("", configController.GetConfig)
		config.PUT("", configController.UpdateConfig)
	}
}
