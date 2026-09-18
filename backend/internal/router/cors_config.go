package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterCorsConfigRoutes 注册跨域配置管理路由。
func RegisterCorsConfigRoutes(r *gin.RouterGroup, corsConfigController *controller.CorsConfigController, authMiddleware gin.HandlerFunc) {
	cors := r.Group("/cors-config")
	cors.Use(authMiddleware)
	{
		cors.GET("", corsConfigController.GetConfig)
		cors.PUT("", corsConfigController.UpdateConfig)
	}
}
