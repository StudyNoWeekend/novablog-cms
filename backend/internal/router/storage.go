package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterStorageRoutes 注册存储配置管理路由。
func RegisterStorageRoutes(r *gin.RouterGroup, ctrl *controller.StorageController, migrationCtrl *controller.MigrationController, authMiddleware gin.HandlerFunc) {
	storage := r.Group("/storage")
	storage.Use(authMiddleware)
	{
		storage.GET("/config", ctrl.GetConfigs)
		storage.POST("/config", ctrl.UpsertConfig)
		storage.DELETE("/config/:provider", ctrl.DeleteConfig)
		storage.PUT("/config/:provider/activate", ctrl.ActivateConfig)
		storage.POST("/config/test", ctrl.TestConfig)

		// 素材迁移路由
		storage.POST("/migration/analyze", migrationCtrl.Analyze)
		storage.GET("/migration/analyze/:taskId", migrationCtrl.GetAnalyzeResult)
		storage.POST("/migration/start", migrationCtrl.StartMigration)
		storage.GET("/migration/status/:taskId", migrationCtrl.GetMigrationStatus)
	}
}
