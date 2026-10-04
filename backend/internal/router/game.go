package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterGameRoutes 注册游戏库管理路由。
func RegisterGameRoutes(r *gin.RouterGroup, gameController *controller.GameController, authMiddleware gin.HandlerFunc) {
	games := r.Group("/games")
	games.Use(authMiddleware)
	{
		games.POST("", gameController.Create)
		games.GET("", gameController.GetList)
		games.GET("/:id", gameController.GetByID)
		games.PUT("/:id", gameController.Update)
		games.DELETE("/:id", gameController.Delete)
	}
}
