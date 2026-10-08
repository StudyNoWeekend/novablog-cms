package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterMapEmbedRoutes 注册地图嵌入页公开路由（根级，供博客主题 iframe 嵌入）。
// 显式注册于 NoRoute 主题托管兜底之前，不会被主题路由吞掉。
func RegisterMapEmbedRoutes(r *gin.Engine) {
	mapEmbedController := controller.NewMapEmbedController()
	r.GET("/map-embed/travel/:id", mapEmbedController.Travel)
	r.GET("/map-embed/assets/:file", mapEmbedController.Asset)
}
