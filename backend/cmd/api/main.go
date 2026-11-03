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
	"novablog/internal/logic"
	"novablog/internal/middleware"
	"novablog/internal/model"
	"novablog/internal/router"
	"novablog/internal/storage"
	"novablog/pkg/geoip"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// version 应用版本号，构建期通过 -ldflags "-X main.version=..." 注入。
var version = "dev"

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
	// 数据文件缺失时转为后台下载，不阻塞启动；完成前 Lookup 返回空地区
	// 下载地址与代理为运行时配置（system_configs 表，后台「系统配置-站点信息」维护）
	{
		systemConfig := logic.NewSystemConfig()
		geoipSearcher := geoip.New(geoip.Config{
			DataDir:     app.Config.GetString("geoip.data_dir"),
			DownloadURL: systemConfig.GetGeoipDownloadURL(context.Background()),
			ProxyURL:    systemConfig.GetProxyURL(context.Background()),
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
	// 站点对外地址权威值在数据库（后台「系统配置-站点信息」维护），非空时覆盖 yaml 首装默认
	if siteURL := logic.NewSystemConfig().GetSiteURL(context.Background()); siteURL != "" {
		model.BaseURL = siteURL
	}

	// 注入媒体 URL 动态拼接器：model 层按媒体 storage_type 与相对 key
	// 经 storage.Manager 拼接当前配置下的完整 URL（换域名/换平台零改写）
	model.ResolveStorageURL = storageMgr.ResolvePublicURL

	// 注入云盘导入模块配置（WebDAV/百度网盘/阿里云盘；代理与限额为运行时配置走数据库）
	{
		systemConfig := logic.NewSystemConfig()
		logic.SetCloudDriveSettings(&logic.CloudDriveSettings{
			ProxyURL:         systemConfig.GetProxyURL(context.Background()),
			MaxFileSizeMB:    systemConfig.GetCloudMaxFileSizeMB(context.Background()),
			ImportBatchLimit: systemConfig.GetCloudImportBatchLimit(context.Background()),
		})
	}

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

	// 访问日志清理、访问统计维护、社区推送重试等后台定时任务
	// 已迁移至独立入口 cmd/cron（规范要求：禁止在 api 入口以 go 协程运行长驻任务）

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("正在关闭服务器...")

	// 优雅关闭（设置 10 秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		// 记录后继续关闭 DB/Redis，避免 Fatal 直接退出跳过资源释放
		app.Logger.Error("服务器强制关闭", zap.Error(err))
	}

	// 关闭数据库连接
	if sqlDB, err := app.DB.DB(); err == nil {
		sqlDB.Close()
	}

	// 关闭 Redis 连接
	app.Redis.Close()

	app.Logger.Info("服务器已安全退出")
}
