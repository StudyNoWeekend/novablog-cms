package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterOpenSourceRoutes 注册开源作品管理路由。
func RegisterOpenSourceRoutes(r *gin.RouterGroup, openSourceController *controller.OpenSourceController, authMiddleware gin.HandlerFunc) {
	openSources := r.Group("/open-sources")
	openSources.Use(authMiddleware)
	{
		openSources.POST("", openSourceController.Create)
		openSources.GET("", openSourceController.GetList)
		// 静态路由必须在参数路由之前注册，避免 Gin 路由冲突
		openSources.POST("/:id/refresh", openSourceController.Refresh)
		openSources.GET("/:id", openSourceController.GetByID)
		openSources.PUT("/:id", openSourceController.Update)
		openSources.DELETE("/:id", openSourceController.Delete)
	}
}
