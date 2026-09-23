package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterMediaRoutes 注册媒体管理路由。
func RegisterMediaRoutes(r *gin.RouterGroup, mediaController *controller.MediaController, authMiddleware gin.HandlerFunc) {
	media := r.Group("/media")
	media.Use(authMiddleware)
	{
		media.POST("/upload", mediaController.Upload)
		media.POST("/upload-with-preset", mediaController.UploadWithPreset)
		media.GET("", mediaController.GetList)
		media.DELETE("/:id", mediaController.Delete)

		media.GET("/:id/usages", mediaController.GetUsages)
		media.GET("/:id/presets", mediaController.GetPresets)
		media.POST("/preset", mediaController.CreatePreset)
		media.DELETE("/preset/:id", mediaController.DeletePreset)
	}
}
