package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterAPIDocRoutes 注册 API 文档路由（需认证）。
func RegisterAPIDocRoutes(r *gin.RouterGroup, apiDocController *controller.APIDocController, authMiddleware gin.HandlerFunc) {
	docs := r.Group("/api-docs")
	docs.Use(authMiddleware)
	{
		docs.GET("", apiDocController.GetOpenAPIDocs)
	}
}
