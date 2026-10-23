package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterCloudDriveRoutes 注册云盘导入路由。
// callback 为 OAuth 回调（公开，state 防 CSRF）；其余为管理能力（需认证）。
func RegisterCloudDriveRoutes(r *gin.RouterGroup, ctrl *controller.CloudDriveController, authMiddleware gin.HandlerFunc) {
	// OAuth 授权回调（公开路由：浏览器跳转携带 code，无法附加 Authorization 头）
	r.GET("/clouddrive/callback/:provider", ctrl.OAuthCallback)

	cd := r.Group("/clouddrive")
	cd.Use(authMiddleware)
	{
		// 应用凭据（百度/阿里开放平台 AppKey/Secret）
		cd.GET("/config", ctrl.GetConfigs)
		cd.POST("/config", ctrl.UpsertConfig)
		cd.DELETE("/config/:provider", ctrl.DeleteConfig)

		// 授权
		cd.POST("/authorize/:provider", ctrl.GetAuthorizeURL)

		// 账号
		cd.GET("/accounts", ctrl.ListAccounts)
		cd.POST("/accounts/webdav", ctrl.AddWebDAVAccount)
		cd.POST("/accounts/webdav/test", ctrl.TestWebDAV)
		cd.DELETE("/accounts/:id", ctrl.DeleteAccount)
		cd.GET("/accounts/:id/files", ctrl.ListFiles)

		// 导入
		cd.POST("/import/start", ctrl.StartImport)
		cd.GET("/import/status/:taskId", ctrl.GetImportStatus)
		cd.GET("/import/tasks", ctrl.GetImportTasks)
	}
}
