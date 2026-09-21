package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterEquipmentRoutes 注册个人设备管理路由。
func RegisterEquipmentRoutes(r *gin.RouterGroup, equipmentController *controller.EquipmentController, authMiddleware gin.HandlerFunc) {
	equipments := r.Group("/equipments")
	equipments.Use(authMiddleware)
	{
		equipments.POST("", equipmentController.Create)
		equipments.GET("", equipmentController.GetList)
		equipments.GET("/:id", equipmentController.GetByID)
		equipments.PUT("/:id", equipmentController.Update)
		equipments.DELETE("/:id", equipmentController.Delete)
	}
}
