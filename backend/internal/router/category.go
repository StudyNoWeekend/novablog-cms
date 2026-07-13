package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterCategoryRoutes 注册分类与标签管理路由。
func RegisterCategoryRoutes(r *gin.RouterGroup, categoryController *controller.CategoryController, tagController *controller.TagController, authMiddleware gin.HandlerFunc) {
	cat := r.Group("/categories")
	cat.Use(authMiddleware)
	{
		cat.POST("", categoryController.Create)
		cat.GET("", categoryController.GetAll)
		cat.GET("/:id", categoryController.GetByID)
		cat.PUT("/:id", categoryController.Update)
		cat.DELETE("/:id", categoryController.Delete)
	}

	// 标签路由
	tag := r.Group("/tags")
	tag.Use(authMiddleware)
	{
		tag.POST("", tagController.Create)
		tag.GET("", tagController.GetAll)
		tag.GET("/:id", tagController.GetByID)
		tag.PUT("/:id", tagController.Update)
		tag.DELETE("/:id", tagController.Delete)
	}
}
