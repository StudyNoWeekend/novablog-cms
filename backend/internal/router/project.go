package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterProjectRoutes 注册项目经历管理路由。
func RegisterProjectRoutes(r *gin.RouterGroup, projectController *controller.ProjectController, authMiddleware gin.HandlerFunc) {
	projects := r.Group("/projects")
	projects.Use(authMiddleware)
	{
		projects.POST("", projectController.Create)
		projects.GET("", projectController.GetList)
		projects.GET("/:id", projectController.GetByID)
		projects.PUT("/:id", projectController.Update)
		projects.DELETE("/:id", projectController.Delete)
	}
}
