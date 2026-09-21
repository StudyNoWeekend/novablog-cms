// Package router 定义路由注册。
package router

import (
	"context"
	"net/http"

	"novablog/internal/cache"
	"novablog/internal/controller"
	"novablog/internal/logic"
	"novablog/internal/middleware"
	"novablog/internal/model"
	"novablog/internal/storage"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Version 应用版本号，由 cmd/api 在启动时注入（构建期 -ldflags 决定）。
var Version = "dev"

// RegisterRoutes 注册所有 API 路由。
func RegisterRoutes(r *gin.Engine, logger *zap.Logger, db *gorm.DB, accessSecret string, storageMgr *storage.Manager, cryptoKey, uploadDir string) {
	// 全局中间件
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.AccessLogMiddleware())
	r.Use(middleware.CORSMiddleware())

	// 健康检查路由
	RegisterHealthRouter(r)

	// 文件服务路由（公开访问，无需认证）
	r.Static("/files", uploadDir)

	// 初始化依赖
	authController := controller.NewAuthController()
	publicController := controller.NewPublicController(storageMgr, cryptoKey)
	mediaController := controller.NewMediaController(storageMgr)
	categoryController := controller.NewCategoryController()
	tagController := controller.NewTagController()
	articleController := controller.NewArticleController()
	portfolioController := controller.NewPortfolioController()
	videoController := controller.NewVideoController()
	equipmentController := controller.NewEquipmentController()
	projectController := controller.NewProjectController()
	travelController := controller.NewTravelGuideController()
	musicController := controller.NewMusicController(storageMgr)
	commentController := controller.NewCommentController()
	analyticsController := controller.NewAnalyticsController()
	securityController := controller.NewSecurityController()
	apiDocController := controller.NewAPIDocController()
	profileController := controller.NewProfileController(storageMgr)
	moduleConfigController := controller.NewModuleConfigController()
	authMiddleware := middleware.AuthMiddleware(accessSecret)

	// 安全缓存与 IP 黑名单中间件
	securityCache := cache.NewSecurityCache()
	securityModel := model.NewSecurity()
	ipBlacklistMiddleware := middleware.IPBlacklistMiddleware(securityCache, securityModel)

	// API 路由组（统一应用 IP 黑名单中间件）
	api := r.Group("/api/v1")
	api.Use(ipBlacklistMiddleware)

	// 启动时预加载安全配置到缓存
	securityLogic := logic.NewSecurityLogic()
	if err := securityLogic.InitCache(context.Background()); err != nil {
		logger.Warn("预加载安全配置缓存失败", zap.Error(err))
	}

	// 存储配置管理依赖
	storageLogic := logic.NewStorageLogic(storageMgr, cryptoKey, model.NewStorageMigration())
	storageController := controller.NewStorageController(storageLogic)

	// 素材迁移管理依赖
	migrationLogic := logic.NewMigrationLogic(storageMgr, cryptoKey, uploadDir, logger)
	migrationController := controller.NewMigrationController(migrationLogic)

	// 注册公开路由（无需认证）
	public := api.Group("/public")
	RegisterPublicRoutes(public, publicController)

	// 注册认证路由
	RegisterAuthRoutes(api, authController, authMiddleware)

	// 注册媒体管理路由
	RegisterMediaRoutes(api, mediaController, authMiddleware)

	// 注册存储配置管理路由
	RegisterStorageRoutes(api, storageController, migrationController, authMiddleware)

	// 注册分类与标签管理路由
	RegisterCategoryRoutes(api, categoryController, tagController, authMiddleware)

	// 注册文章管理路由
	RegisterArticleRoutes(api, articleController, authMiddleware)

	// 注册摄影作品集管理路由
	RegisterPortfolioRoutes(api, portfolioController, authMiddleware)

	// 注册视频作品管理路由
	RegisterVideoRoutes(api, videoController, authMiddleware)

	// 注册个人设备管理路由
	RegisterEquipmentRoutes(api, equipmentController, authMiddleware)

	// 注册项目经历管理路由
	RegisterProjectRoutes(api, projectController, authMiddleware)

	// 注册开源作品管理路由
	openSourceController := controller.NewOpenSourceController()
	RegisterOpenSourceRoutes(api, openSourceController, authMiddleware)

	// 注册旅行攻略管理路由
	RegisterTravelRoutes(api, travelController, authMiddleware)

	// 注册音乐播放器管理路由
	RegisterMusicRoutes(api, musicController, authMiddleware)

	// 注册评论管理路由
	RegisterCommentRoutes(api, commentController, authMiddleware)

	// 注册工作台统计路由
	RegisterAnalyticsRoutes(api, analyticsController, authMiddleware)

	// 注册安全管理路由
	RegisterSecurityRoutes(api, securityController, authMiddleware)

	// 注册 API 文档路由
	RegisterAPIDocRoutes(api, apiDocController, authMiddleware)

	// 注册官方主题市场代理路由
	themeMarketController := controller.NewThemeMarketController()
	RegisterThemeMarketRoutes(api, themeMarketController, authMiddleware)

	// 注册已安装主题管理路由
	themeController := controller.NewThemeController()
	RegisterThemeRoutes(api, themeController, authMiddleware)

	// 注册个人资料管理路由
	RegisterProfileRoutes(api, profileController, authMiddleware)

	// 注册模块开关配置管理路由
	RegisterModuleConfigRoutes(api, moduleConfigController, authMiddleware)

	// 注册跨域配置管理路由
	corsConfigController := controller.NewCorsConfigController()
	RegisterCorsConfigRoutes(api, corsConfigController, authMiddleware)

	// 注册官方主题市场配置管理路由
	themeMarketConfigController := controller.NewThemeMarketConfigController()
	RegisterThemeMarketConfigRoutes(api, themeMarketConfigController, authMiddleware)

	// 注册第三方歌单管理路由
	playlistController := controller.NewPlaylistController()
	RegisterPlaylistRoutes(api, playlistController, authMiddleware)

	// 博客主题预览路由（以指定主题响应托管逻辑，激活前即可预览真实效果）
	themeLogic := logic.NewThemeLogic()
	r.GET("/preview/:theme_id", logic.ThemeHost.PreviewHandler())
	r.GET("/preview/:theme_id/*path", logic.ThemeHost.PreviewHandler())

	// 博客主题静态托管兜底路由（激活主题；注册于所有业务路由之后）
	r.NoRoute(logic.ThemeHost.Handler(themeLogic))
}

// RegisterHealthRouter 注册健康检查路由。
func RegisterHealthRouter(r *gin.Engine) {
	// /health 返回 200
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": Version,
		})
	})

	// /ready 检查数据库连接
	r.GET("/ready", func(c *gin.Context) {
		// 如果需要在 ready 中使用 db，可以通过闭包注入
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})
}
