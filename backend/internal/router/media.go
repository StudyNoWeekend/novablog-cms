package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterMediaRoutes 注册媒体管理路由。
func RegisterMediaRoutes(r *gin.RouterGroup, mediaController *controller.MediaController, folderController *controller.MediaFolderController, authMiddleware gin.HandlerFunc) {
	media := r.Group("/media")
	media.Use(authMiddleware)
	{
		media.POST("/upload", mediaController.Upload)
		media.POST("/upload-with-preset", mediaController.UploadWithPreset)
		media.GET("", mediaController.GetList)
		media.PUT("/move", mediaController.MoveMedia)
		media.DELETE("/:id", mediaController.Delete)

		// 文件夹管理
		media.GET("/folders/tree", folderController.Tree)
		media.POST("/folders", folderController.Create)
		media.PUT("/folders/:id", folderController.Update)
		media.DELETE("/folders/:id", folderController.Delete)

		media.GET("/:id/usages", mediaController.GetUsages)
		media.GET("/:id/presets", mediaController.GetPresets)
		media.POST("/preset", mediaController.CreatePreset)
		media.DELETE("/preset/:id", mediaController.DeletePreset)
	}
}
