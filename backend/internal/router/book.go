package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterBookRoutes 注册读书书架管理路由。
func RegisterBookRoutes(r *gin.RouterGroup, bookController *controller.BookController, authMiddleware gin.HandlerFunc) {
	books := r.Group("/books")
	books.Use(authMiddleware)
	{
		books.POST("", bookController.Create)
		books.GET("", bookController.GetList)
		books.GET("/:id", bookController.GetByID)
		books.PUT("/:id", bookController.Update)
		books.DELETE("/:id", bookController.Delete)
	}
}
