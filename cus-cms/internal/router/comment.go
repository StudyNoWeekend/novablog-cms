package router

import (
	"cus-cms/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterCommentRoutes 注册评论管理路由。
func RegisterCommentRoutes(r *gin.RouterGroup, commentController *controller.CommentController, authMiddleware gin.HandlerFunc) {
	comments := r.Group("/comments")
	comments.Use(authMiddleware)
	{
		comments.GET("", commentController.GetList)
		comments.POST("/:id/reply", commentController.Reply)
		comments.DELETE("/:id", commentController.Delete)
	}
}
