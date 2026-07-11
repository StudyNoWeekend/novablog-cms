package router

import (
	"cus-cms/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterTravelRoutes 注册旅行攻略管理路由。
func RegisterTravelRoutes(r *gin.RouterGroup, travelController *controller.TravelGuideController, authMiddleware gin.HandlerFunc) {
	travels := r.Group("/travels")
	travels.Use(authMiddleware)
	{
		travels.POST("", travelController.Create)
		travels.GET("", travelController.GetList)
		travels.GET("/:id", travelController.GetDetail)
		travels.PUT("/:id", travelController.Update)
		travels.DELETE("/:id", travelController.Delete)
		travels.PUT("/:id/status", travelController.UpdateStatus)
	}
}
