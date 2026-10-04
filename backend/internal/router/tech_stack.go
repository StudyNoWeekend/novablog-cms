package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterTechStackRoutes 注册技术栈管理路由。
func RegisterTechStackRoutes(r *gin.RouterGroup, techStackController *controller.TechStackController, authMiddleware gin.HandlerFunc) {
	techStacks := r.Group("/tech-stacks")
	techStacks.Use(authMiddleware)
	{
		techStacks.POST("", techStackController.Create)
		techStacks.GET("", techStackController.GetList)
		techStacks.GET("/:id", techStackController.GetByID)
		techStacks.PUT("/:id", techStackController.Update)
		techStacks.DELETE("/:id", techStackController.Delete)
	}
}
