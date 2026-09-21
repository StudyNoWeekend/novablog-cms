package router

import (
	"novablog/internal/controller"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes 注册公开路由（无需认证）。
func RegisterPublicRoutes(r *gin.RouterGroup, pc *controller.PublicController) {
	// 首次安装引导接口（公开，避免与"管理后台设置"语义混淆）
	install := r.Group("/install")
	{
		install.GET("/status", pc.GetStatus)
		install.POST("/init", pc.Init)
		// 首装初始化对象存储配置（可选，本地存储可跳过）
		install.POST("/storage", pc.SetupStorage)
		// 首装初始化博客外观：拉取官方默认主题（失败回退内置兜底）并激活
		install.POST("/theme", pc.InitTheme)
		install.GET("/theme/status", pc.GetThemeInstallStatus)
	}

	public := r.Group("")
	{
		// 应用版本号下发（免鉴权）：构建期 -ldflags 注入，供管理端侧边栏展示
		public.GET("/version", func(c *gin.Context) {
			response.Success(c, gin.H{"version": Version})
		})

		// 公共配置下发（免鉴权）：官方市场地址默认值等由后端控制
		public.GET("/config", pc.GetPublicConfig)

		// 博主信息
		public.GET("/blogger", pc.GetBlogger)

		// 文章（静态路由必须在参数路由之前注册，避免 Gin 路由冲突）
		public.GET("/articles", pc.GetArticles)
		public.GET("/articles/hot", pc.GetHotArticles)
		public.GET("/articles/random", pc.GetRandomArticles)
		public.GET("/articles/:slug", pc.GetArticleBySlug)
		public.GET("/articles/:slug/view", pc.IncrementArticleView)

		// 分类和标签
		public.GET("/categories", pc.GetCategories)
		public.GET("/tags", pc.GetTags)

		// 评论
		public.GET("/comments", pc.GetComments)
		public.POST("/comments", pc.CreateComment)

		// 旅行攻略（静态路由必须在参数路由之前注册）
		public.GET("/travels", pc.GetTravels)
		public.GET("/travels/hot", pc.GetHotTravels)
		public.GET("/travels/:id", pc.GetTravelDetail)
		public.GET("/travels/:id/view", pc.IncrementTravelView)
		public.POST("/travels/:id/like", pc.LikeTravel)

		// 作品集
		public.GET("/portfolios", pc.GetPortfolios)
		public.GET("/portfolios/:id", pc.GetPortfolioDetail)

		// 视频作品
		public.GET("/videos", pc.GetVideos)
		public.GET("/videos/:id", pc.GetVideoDetail)

		// 个人设备
		public.GET("/equipments", pc.GetEquipments)
		public.GET("/equipments/:id", pc.GetEquipmentDetail)

		// 项目经历
		public.GET("/projects", pc.GetProjects)
		public.GET("/projects/:id", pc.GetProjectDetail)

		// 开源作品
		public.GET("/open-sources", pc.GetOpenSources)
		public.GET("/open-sources/:id", pc.GetOpenSourceDetail)

		// 音乐
		public.GET("/music/songs", pc.GetSongs)
		public.GET("/music/songs/:id", pc.GetSongDetail)
		public.GET("/music/audio-url/:song_id", pc.GetAudioURL)
		public.GET("/music/playlists", pc.GetPlaylists)

		// 模块开关配置
		public.GET("/module-config", pc.GetModuleConfig)

		// 第三方歌单
		public.GET("/playlists", pc.GetPlaylists)
	}
}
