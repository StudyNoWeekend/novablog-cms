package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterThemeMarketRoutes 注册官方主题市场代理路由（需后台登录）。
func RegisterThemeMarketRoutes(r *gin.RouterGroup, themeMarketController *controller.ThemeMarketController, authMiddleware gin.HandlerFunc) {
	market := r.Group("/themes/market")
	market.Use(authMiddleware)
	{
		// 官方账号登录态
		market.POST("/auth/login", themeMarketController.MarketLogin)
		market.POST("/auth/refresh", themeMarketController.MarketRefresh)
		market.POST("/auth/logout", themeMarketController.MarketLogout)

		// 市场浏览（公开数据）
		market.GET("", themeMarketController.GetList)
		market.GET("/stats", themeMarketController.GetStats)
		market.GET("/hot-tags", themeMarketController.GetHotTags)
		market.GET("/detail/:id", themeMarketController.GetDetail)

		// 我的收藏
		market.GET("/favorites", themeMarketController.GetFavorites)
		market.GET("/favorites/ids", themeMarketController.GetFavoriteIds)

		// 互动（需官方登录）
		market.POST("/:id/like", themeMarketController.Like)
		market.POST("/:id/favorite", themeMarketController.Favorite)
		market.POST("/:id/rating", themeMarketController.Rating)
		market.POST("/:id/download", themeMarketController.Download)
	}
}
