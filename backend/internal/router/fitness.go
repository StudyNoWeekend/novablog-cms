package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterFitnessRoutes 注册健身训练管理路由。
func RegisterFitnessRoutes(r *gin.RouterGroup, fitnessController *controller.FitnessController, authMiddleware gin.HandlerFunc) {
	fitness := r.Group("/fitness")
	fitness.Use(authMiddleware)
	{
		fitness.POST("", fitnessController.Create)
		fitness.GET("", fitnessController.GetList)
		fitness.GET("/:id", fitnessController.GetByID)
		fitness.PUT("/:id", fitnessController.Update)
		fitness.DELETE("/:id", fitnessController.Delete)
	}
}
