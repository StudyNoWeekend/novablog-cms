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

	"cus-cms/bootstrap"
	"cus-cms/internal/model"
	"cus-cms/internal/router"
	"cus-cms/internal/storage"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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

	// 创建 Gin 引擎
	r := gin.New()

	// 构建存储管理器
	configModel := model.NewStorageConfig()
	cryptoKey := app.Config.GetString("crypto.secret_key")
	storageMgr := storage.NewManager(configModel, cryptoKey, app.Logger)
	if err := storageMgr.Reload(context.Background()); err != nil {
		app.Logger.Warn("存储管理器初始化失败", zap.Error(err))
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
