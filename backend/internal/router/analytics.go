package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterAnalyticsRoutes 注册工作台统计路由。
func RegisterAnalyticsRoutes(r *gin.RouterGroup, analyticsController *controller.AnalyticsController, authMiddleware gin.HandlerFunc) {
	analytics := r.Group("/analytics")
	analytics.Use(authMiddleware)
	{
		analytics.GET("/overview", analyticsController.GetOverview)
		analytics.GET("/content-trend", analyticsController.GetContentTrend)
		analytics.GET("/top-content", analyticsController.GetTopContent)
		analytics.GET("/distribution", analyticsController.GetDistribution)
		analytics.GET("/recent-comments", analyticsController.GetRecentComments)
		analytics.GET("/traffic/summary", analyticsController.GetTrafficSummary)
		analytics.GET("/traffic/trend", analyticsController.GetTrafficTrend)
		analytics.GET("/module-stats", analyticsController.GetModuleStats)
	}
}
