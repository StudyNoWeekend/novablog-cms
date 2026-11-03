// Package main 为后台定时任务入口（cmd/cron）：
// 访问日志清理、访问统计维护、社区推送失败重试。
// 与 cmd/api 同镜像部署（容器内由 supervisord 托管独立进程），
// 收到 SIGINT/SIGTERM 后停止进入新周期、等待在跑任务完成（30s 上限）再关闭连接。
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"novablog/bootstrap"
	"novablog/internal/cron"
	"novablog/internal/logic"
	"novablog/internal/model"
	"novablog/internal/storage"

	"go.uber.org/zap"
)

func main() {
	// 确定配置文件路径
	cfgPath := "config/config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	// 初始化应用（与 cmd/api 共用 bootstrap：配置/日志/DB/Redis/种子逻辑）
	app, err := bootstrap.NewApp(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "应用初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer app.Logger.Sync()

	// 媒体 URL 解析依赖：站点对外地址（DB 权威值优先）与存储管理器（对象存储 URL 拼接），
	// 与 cmd/api 保持一致，确保推送 payload 等产出的链接与线上形态一致
	model.BaseURL = app.Config.GetString("upload.base_url")
	if siteURL := logic.NewSystemConfig().GetSiteURL(context.Background()); siteURL != "" {
		model.BaseURL = siteURL
	}
	storageMgr := storage.NewManager(
		model.NewStorageConfig(),
		app.Config.GetString("crypto.secret_key"),
		app.Config.GetString("upload.dir"),
		app.Logger,
	)
	if err := storageMgr.Reload(context.Background()); err != nil {
		app.Logger.Warn("存储管理器初始化失败", zap.Error(err))
	}
	model.ResolveStorageURL = storageMgr.ResolvePublicURL

	app.Logger.Info("NovaBlog 定时任务启动", zap.String("config", cfgPath))

	// SIGINT/SIGTERM 触发 ctx 取消，各任务退出循环
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); cron.RunAccessLogCleanup(ctx, app.Logger) }()
	go func() { defer wg.Done(); cron.RunViewStatsMaintenance(ctx, app.Logger) }()
	go func() { defer wg.Done(); cron.RunCommunityPushRetry(ctx, app.Logger) }()

	<-ctx.Done()
	app.Logger.Info("正在停止定时任务...")

	// 等待在跑任务收尾，30s 上限
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		app.Logger.Warn("等待定时任务收尾超时，强制退出")
	}

	// 关闭数据库与 Redis 连接
	if sqlDB, err := app.DB.DB(); err == nil {
		sqlDB.Close()
	}
	app.Redis.Close()

	app.Logger.Info("定时任务已安全退出")
}
