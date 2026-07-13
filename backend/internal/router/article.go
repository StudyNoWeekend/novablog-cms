package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterArticleRoutes 注册文章管理路由。
func RegisterArticleRoutes(r *gin.RouterGroup, articleController *controller.ArticleController, authMiddleware gin.HandlerFunc) {
	articles := r.Group("/articles")
	articles.Use(authMiddleware)
	{
		articles.POST("", articleController.Create)
		articles.GET("", articleController.GetList)
		articles.GET("/:id", articleController.GetDetail)
		articles.PUT("/:id", articleController.Update)
		articles.DELETE("/:id", articleController.Delete)
		articles.PUT("/:id/status", articleController.UpdateStatus)
	}
}
