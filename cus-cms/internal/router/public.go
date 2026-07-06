package router

import (
	"cus-cms/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes 注册公开路由（无需认证）。
func RegisterPublicRoutes(r *gin.RouterGroup, pc *controller.PublicController) {
	setup := r.Group("/setup")
	{
		setup.GET("/status", pc.GetStatus)
		setup.POST("/init", pc.Init)
	}

	public := r.Group("")
	{
		public.GET("/articles", pc.GetArticles)
		public.GET("/articles/:slug", pc.GetArticleBySlug)
		public.GET("/categories", pc.GetCategories)
		public.GET("/tags", pc.GetTags)
	}
}
