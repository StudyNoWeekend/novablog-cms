package router

import (
	"cus-cms/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterPortfolioRoutes 注册摄影作品集管理路由。
func RegisterPortfolioRoutes(r *gin.RouterGroup, portfolioController *controller.PortfolioController, authMiddleware gin.HandlerFunc) {
	portfolios := r.Group("/portfolios")
	portfolios.Use(authMiddleware)
	{
		portfolios.POST("", portfolioController.Create)
		portfolios.GET("", portfolioController.GetList)
		portfolios.GET("/:id", portfolioController.GetByID)
		portfolios.PUT("/:id", portfolioController.Update)
		portfolios.DELETE("/:id", portfolioController.Delete)

		portfolios.POST("/:id/items", portfolioController.AddItem)
		// 注意：items/sort 必须在 items/:itemId 之前注册，否则会被 :itemId 匹配
		portfolios.PUT("/:id/items/sort", portfolioController.SortItems)
		portfolios.PUT("/:id/items/:itemId", portfolioController.UpdateItem)
		portfolios.DELETE("/:id/items/:itemId", portfolioController.DeleteItem)
	}
}
