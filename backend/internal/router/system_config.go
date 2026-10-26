package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterSystemConfigRoutes 注册通用系统配置路由（站点信息分组 + 配置导出/导入）。
func RegisterSystemConfigRoutes(r *gin.RouterGroup, systemConfigController *controller.SystemConfigController, authMiddleware gin.HandlerFunc) {
	group := r.Group("/system-config")
	group.Use(authMiddleware)
	{
		group.GET("", systemConfigController.GetSystemConfig)
		group.PUT("", systemConfigController.UpdateSystemConfig)
		group.GET("/export", systemConfigController.ExportConfig)
		group.POST("/import", systemConfigController.ImportConfig)
	}
}
