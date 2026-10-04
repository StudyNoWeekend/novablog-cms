package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterImageSearchRoutes 注册图片搜索路由。
// thumbnail 代理输出外部图片（公开，供搜索弹窗 <img> 引用，无鉴权请求头）；
// 搜索与转存为管理能力（需认证）。
func RegisterImageSearchRoutes(r *gin.RouterGroup, imageSearchController *controller.ImageSearchController, authMiddleware gin.HandlerFunc) {
	search := r.Group("/image-search")
	{
		search.GET("/thumbnail", imageSearchController.Thumbnail)

		authed := search.Group("")
		authed.Use(authMiddleware)
		{
			authed.GET("", imageSearchController.Search)
			authed.POST("/save", imageSearchController.Save)
		}
	}
}
