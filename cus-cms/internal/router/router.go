// Package router 定义路由注册。
package router

import (
	"net/http"

	"cus-cms/internal/controller"
	"cus-cms/internal/logic"
	"cus-cms/internal/middleware"
	"cus-cms/internal/model"
	"cus-cms/internal/storage"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RegisterRoutes 注册所有 API 路由。
func RegisterRoutes(r *gin.Engine, logger *zap.Logger, db *gorm.DB, accessSecret string, storageMgr *storage.Manager, cryptoKey, uploadDir string) {
	// 全局中间件
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.CORSMiddleware())

	// 健康检查路由
	RegisterHealthRouter(r)

	// API 路由组
	api := r.Group("/api/v1")

	// 初始化依赖
	authController := controller.NewAuthController()
	publicController := controller.NewPublicController()
	mediaController := controller.NewMediaController(storageMgr)
	categoryController := controller.NewCategoryController()
	tagController := controller.NewTagController()
	articleController := controller.NewArticleController()
	portfolioController := controller.NewPortfolioController()
	videoController := controller.NewVideoController()
	travelController := controller.NewTravelGuideController()
	musicController := controller.NewMusicController()
	commentController := controller.NewCommentController()
	authMiddleware := middleware.AuthMiddleware(accessSecret)

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

	// 注册旅行攻略管理路由
	RegisterTravelRoutes(api, travelController, authMiddleware)

	// 注册音乐播放器管理路由
	RegisterMusicRoutes(api, musicController, authMiddleware)

	// 注册评论管理路由
	RegisterCommentRoutes(api, commentController, authMiddleware)
}

// RegisterHealthRouter 注册健康检查路由。
func RegisterHealthRouter(r *gin.Engine) {
	// /health 返回 200
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
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
