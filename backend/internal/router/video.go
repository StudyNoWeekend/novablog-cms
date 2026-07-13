package router

import (
	"cus-cms/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterVideoRoutes 注册视频作品管理路由。
func RegisterVideoRoutes(r *gin.RouterGroup, videoController *controller.VideoController, authMiddleware gin.HandlerFunc) {
	videos := r.Group("/videos")
	videos.Use(authMiddleware)
	{
		videos.POST("", videoController.Create)
		videos.GET("", videoController.GetList)
		videos.POST("/parse", videoController.Parse)
		videos.GET("/:id", videoController.GetByID)
		videos.PUT("/:id", videoController.Update)
		videos.DELETE("/:id", videoController.Delete)
	}
}
