// Package main 为 backend 服务入口。
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"novablog/bootstrap"
	"novablog/internal/middleware"
	"novablog/internal/model"
	"novablog/internal/router"
	"novablog/internal/storage"
	"novablog/pkg/geoip"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// version 应用版本号，构建期通过 -ldflags "-X main.version=..." 注入。
var version = "dev"

// startAccessLogCleanup 启动访问日志清理定时任务。
// 每隔一小时删除超过 log_retention_days 天（默认 7 天）的访问日志。
func startAccessLogCleanup(cfg *viper.Viper, logger *zap.Logger) {
	days := cfg.GetInt("security.log_retention_days")
	if days <= 0 {
		days = 7
	}

	accessLogModel := model.NewAccessLog()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	// 启动时立即执行一次清理
	cleanupAccessLogs(accessLogModel, days, logger)

	for range ticker.C {
		cleanupAccessLogs(accessLogModel, days, logger)
	}
}

// cleanupAccessLogs 删除指定保留天数之前的访问日志。
func cleanupAccessLogs(accessLogModel *model.AccessLogModel, days int, logger *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	before := time.Now().AddDate(0, 0, -days)
	if err := accessLogModel.CleanBefore(ctx, before); err != nil {
		logger.Error("清理访问日志失败", zap.Error(err), zap.Int("retention_days", days))
	} else {
		logger.Info("访问日志清理完成", zap.Time("before", before), zap.Int("retention_days", days))
	}
}

func main() {
	// 确定配置文件路径
	cfgPath := "config/config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	// 初始化应用
	app, err := bootstrap.NewApp(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "应用初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer app.Logger.Sync()

	// 设置 Gin 模式
	if app.Config.GetBool("app.debug") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化 IP 归属地解析器（访问日志写入时记录地区）
	{
		geoipSearcher := geoip.New(geoip.Config{
			DataDir:     app.Config.GetString("geoip.data_dir"),
			DownloadURL: app.Config.GetString("geoip.download_url"),
			ProxyURL:    app.Config.GetString("geoip.proxy_url"),
		}, app.Logger)
		if err := geoipSearcher.Load(context.Background()); err != nil {
			// 加载失败仅降级（地区列为空），不影响服务启动
			app.Logger.Warn("IP 归属地解析器初始化失败，访问日志将不记录地区", zap.Error(err))
		}
		middleware.GeoIP = geoipSearcher
	}

	router.Version = version
	app.Logger.Info("NovaBlog 后端启动",
		zap.String("version", version),
		zap.String("config", cfgPath),
	)

	// 创建 Gin 引擎
	r := gin.New()

	// 构建存储管理器
	configModel := model.NewStorageConfig()
	cryptoKey := app.Config.GetString("crypto.secret_key")
	storageMgr := storage.NewManager(configModel, cryptoKey, app.Config.GetString("upload.dir"), app.Logger)
	if err := storageMgr.Reload(context.Background()); err != nil {
		app.Logger.Warn("存储管理器初始化失败", zap.Error(err))
	}

	model.BaseURL = app.Config.GetString("upload.base_url")

	// 注册路由
	router.RegisterRoutes(
		r, app.Logger, app.DB,
		app.Config.GetString("jwt.access_secret"),
		storageMgr,
		cryptoKey,
		app.Config.GetString("upload.dir"),
	)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.GetInt("http.port")),
		Handler:      r,
		ReadTimeout:  app.Config.GetDuration("http.read_timeout"),
		WriteTimeout: app.Config.GetDuration("http.write_timeout"),
	}

	// 启动服务器（非阻塞）
	go func() {
		app.Logger.Info("服务器启动中...", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 启动访问日志清理定时任务
	go startAccessLogCleanup(app.Config, app.Logger)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("正在关闭服务器...")

	// 优雅关闭（设置 10 秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		app.Logger.Fatal("服务器强制关闭", zap.Error(err))
	}

	// 关闭数据库连接
	if sqlDB, err := app.DB.DB(); err == nil {
		sqlDB.Close()
	}

	// 关闭 Redis 连接
	app.Redis.Close()

	app.Logger.Info("服务器已安全退出")
}
