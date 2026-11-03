// Package cron 承载独立于 HTTP 服务的后台定时任务（由 cmd/cron 入口运行）。
package cron

import (
	"context"
	"time"

	"novablog/internal/logic"
	"novablog/internal/model"

	"go.uber.org/zap"
)

// RunAccessLogCleanup 访问日志清理：每小时删除超过保留天数的访问日志。
// 保留天数以 security_configs 表为权威（后台可改），读取失败时回退内置默认 7 天。
// 启动时立即执行一次；ctx 取消后不再进入下一周期（在跑的一次执行完后退出）。
func RunAccessLogCleanup(ctx context.Context, logger *zap.Logger) {
	securityModel := model.NewSecurityConfig()
	resolveDays := func() int {
		qCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if sec, err := securityModel.GetConfig(qCtx); err == nil && sec.LogRetentionDays > 0 {
			return sec.LogRetentionDays
		}
		return 7
	}

	accessLogModel := model.NewAccessLog()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	// 启动时立即执行一次清理
	cleanupAccessLogs(accessLogModel, resolveDays(), logger)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleanupAccessLogs(accessLogModel, resolveDays(), logger)
		}
	}
}

// RunCommunityPushRetry 社区推送失败重试：每 5 分钟扫描同步失败的记录重推最新内容
// （单条失败不影响后续，重试上限见 logic 层）。启动时不立即执行（避免启动期网络抖动误判失败）。
func RunCommunityPushRetry(ctx context.Context, logger *zap.Logger) {
	pushLogic := logic.NewCommunityPushLogic()
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	run := func() {
		runCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		retried, err := pushLogic.RetryFailed(runCtx)
		if err != nil {
			logger.Error("社区推送重试任务执行失败", zap.Error(err))
		} else if retried > 0 {
			logger.Info("社区推送重试完成", zap.Int("retried", retried))
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

// RunViewStatsMaintenance 访问统计维护：每小时将最近 48 小时明细聚合进按日统计表
// （幂等覆盖），并清理超过 180 天的明细。启动时立即执行一次。
func RunViewStatsMaintenance(ctx context.Context, logger *zap.Logger) {
	viewStats := logic.NewViewStatsLogic()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	run := func() {
		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := viewStats.RunMaintenance(runCtx); err != nil {
			logger.Error("访问统计维护失败", zap.Error(err))
		} else {
			logger.Info("访问统计维护完成")
		}
	}

	run()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
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
