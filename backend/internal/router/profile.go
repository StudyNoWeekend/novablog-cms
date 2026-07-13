package router

import (
	"cus-cms/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterProfileRoutes 注册个人资料管理路由。
func RegisterProfileRoutes(r *gin.RouterGroup, profileController *controller.ProfileController, authMiddleware gin.HandlerFunc) {
	profile := r.Group("/profile")
	profile.Use(authMiddleware)
	{
		profile.GET("", profileController.GetProfile)
		profile.PUT("", profileController.UpdateProfile)
		profile.POST("/upload-icon", profileController.UploadIcon)
		profile.POST("/upload-background", profileController.UploadBackground)
	}
}
