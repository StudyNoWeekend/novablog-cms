package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterThemeRoutes 注册已安装主题管理路由（需后台登录）。
func RegisterThemeRoutes(r *gin.RouterGroup, themeController *controller.ThemeController, authMiddleware gin.HandlerFunc) {
	themes := r.Group("/themes")
	themes.Use(authMiddleware)
	{
		// 已安装主题列表（含激活标记）
		themes.GET("", themeController.List)
		// 从官方市场安装
		themes.POST("/install", themeController.Install)
		// 激活指定安装实例（秒级切换）
		themes.POST("/:id/activate", themeController.Activate)
		// 从官方市场更新到最新版本（自动切换激活）
		themes.POST("/:id/update", themeController.Update)
		// 卸载（激活中的实例拒绝）
		themes.DELETE("/:id", themeController.Uninstall)
	}
}
