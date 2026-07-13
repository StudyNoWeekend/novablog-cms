package router

import (
	"cus-cms/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes 注册认证相关路由。
func RegisterAuthRoutes(r *gin.RouterGroup, authController *controller.AuthController, authMiddleware gin.HandlerFunc) {
	auth := r.Group("/auth")
	{
		// 无需认证的路由
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.Refresh)
		auth.POST("/logout", authController.Logout)

		// 需要认证的路由
		authPUT := auth.Group("")
		authPUT.Use(authMiddleware)
		{
			authPUT.PUT("/password", authController.ChangePassword)
		}
	}
}
