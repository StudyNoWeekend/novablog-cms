package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterCommunityPushRoutes 注册 CMS 内容推送到 NovaBlog 社区的管理路由（需后台登录）。
func RegisterCommunityPushRoutes(r *gin.RouterGroup, communityPushController *controller.CommunityPushController, authMiddleware gin.HandlerFunc) {
	push := r.Group("/community-push")
	push.Use(authMiddleware)
	{
		push.GET("/config", communityPushController.GetConfig)
		push.PUT("/config", communityPushController.UpdateConfig)
		push.POST("/test", communityPushController.TestConnection)
		push.GET("/state", communityPushController.GetState)
		push.GET("/states", communityPushController.GetStates)
		push.POST("/sync", communityPushController.Sync)
		push.GET("/records", communityPushController.GetRecords)
		push.GET("/types", communityPushController.GetTypes)
	}
}
