package logic

import (
	"context"
	"fmt"
	"time"

	"novablog/internal/cache"
	"novablog/internal/dto/res"
	"novablog/internal/model"
)

// viewDedupWindow 浏览量 IP 去重窗口（与文章/旅行详情计数保持一致）。
const viewDedupWindow = 5 * time.Minute

// ViewStatsLogic 访问统计业务逻辑结构体。
type ViewStatsLogic struct {
	model *model.ViewStatsModel
	cache *cache.AnalyticsCache
}

// NewViewStatsLogic 创建 ViewStatsLogic 实例。
func NewViewStatsLogic() *ViewStatsLogic {
	return &ViewStatsLogic{
		model: model.NewViewStats(),
		cache: cache.NewAnalyticsCache(),
	}
}

// IsValidViewType 判断内容类型是否支持浏览计数。
func (l *ViewStatsLogic) IsValidViewType(contentType string) bool {
	return model.IsValidViewTarget(contentType)
}

// ContentViewExists 校验目标内容存在且已发布。
func (l *ViewStatsLogic) ContentViewExists(ctx context.Context, contentType, contentID string) (bool, error) {
	return l.model.ContentViewExists(ctx, contentType, contentID)
}

// RecordView 记录一次内容访问：IP 去重后异步写入明细并自增内容表浏览量。
// 返回是否实际计数；Redis 异常时降级为直接计数（fail-open，与文章/旅行计数一致）。
func (l *ViewStatsLogic) RecordView(ctx context.Context, contentType, contentID, ip string) (bool, error) {
	if !model.IsValidViewTarget(contentType) {
		return false, fmt.Errorf("未知的内容类型: %s", contentType)
	}

	var ok bool
	if cache.RedisClient == nil {
		// 未配置 Redis：降级为直接计数
		ok = true
	} else {
		dedupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		var err error
		ok, err = cache.RedisClient.SetNX(dedupCtx, "view:"+contentType+":"+contentID+":"+ip, "1", viewDedupWindow).Result()
		cancel()
		if err != nil {
			ok = true
		}
	}
	if !ok {
		return false, nil
	}

	// 异步落库与自增，不阻塞公开接口响应；明细写入失败则跳过自增，保持明细与计数一致
	go func() {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer bgCancel()
		if err := l.model.InsertViewLog(bgCtx, contentType, contentID, ip); err != nil {
			return
		}
		_ = l.model.IncrementContentView(bgCtx, contentType, contentID)
	}()
	return true, nil
}

// RunMaintenance 聚合最近 48 小时明细进按日统计表，并清理超过 180 天的明细（幂等，每小时执行）。
func (l *ViewStatsLogic) RunMaintenance(ctx context.Context) error {
	if _, err := l.model.RollupDailyViews(ctx, time.Now().Add(-48*time.Hour)); err != nil {
		return err
	}
	_, err := l.model.CleanLogsBefore(ctx, time.Now().AddDate(0, 0, -180))
	return err
}

// GetTrafficSummary 获取访问概览（当日实时 + 昨日聚合 + 累计），带 60 秒缓存。
func (l *ViewStatsLogic) GetTrafficSummary(ctx context.Context) (*res.TrafficSummaryRes, error) {
	cached, err := l.cache.GetTrafficSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取访问概览缓存失败: %w", err)
	}
	if cached != nil {
		return cached, nil
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	todayPV, todayUV, err := l.model.LogsGlobalPvUv(ctx, todayStart)
	if err != nil {
		return nil, err
	}

	statRows, err := l.model.StatsDailySince(ctx, todayStart.AddDate(0, 0, -1))
	if err != nil {
		return nil, err
	}
	yesterday := todayStart.AddDate(0, 0, -1).Format("2006-01-02")
	var yesterdayPV int64
	for _, row := range statRows {
		if row.StatDate.Format("2006-01-02") == yesterday {
			yesterdayPV += row.PV
		}
	}

	totalPV, err := l.model.SumContentViews(ctx)
	if err != nil {
		return nil, err
	}

	summary := &res.TrafficSummaryRes{
		TodayPV:     todayPV,
		TodayUV:     todayUV,
		YesterdayPV: yesterdayPV,
		TotalPV:     totalPV,
	}
	_ = l.cache.SetTrafficSummary(ctx, summary)
	return summary, nil
}

// GetTrafficTrend 获取按天访问趋势（历史取聚合表 + 当日实时明细），带 60 秒缓存。
func (l *ViewStatsLogic) GetTrafficTrend(ctx context.Context, rangeStr string) (*res.TrafficTrendRes, error) {
	cached, err := l.cache.GetTrafficTrend(ctx, rangeStr)
	if err != nil {
		return nil, fmt.Errorf("读取访问趋势缓存失败: %w", err)
	}
	if cached != nil {
		return cached, nil
	}

	var days int
	switch rangeStr {
	case "7d":
		days = 7
	case "90d":
		days = 90
	default:
		days = 30
		rangeStr = "30d"
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startDate := todayStart.AddDate(0, 0, -(days - 1))
	todayKey := todayStart.Format("2006-01-02")

	// 历史日期来自按日聚合表
	statRows, err := l.model.StatsDailySince(ctx, startDate)
	if err != nil {
		return nil, err
	}
	history := make(map[string]map[string]int64, days)
	for _, row := range statRows {
		dateKey := row.StatDate.Format("2006-01-02")
		if dateKey == todayKey {
			continue // 当日以实时明细为准
		}
		if history[dateKey] == nil {
			history[dateKey] = make(map[string]int64)
		}
		history[dateKey][row.ContentType] += row.PV
	}

	// 当日来自明细表实时统计
	logRows, err := l.model.LogsPvByType(ctx, todayStart)
	if err != nil {
		return nil, err
	}
	todayCounts := make(map[string]int64, len(logRows))
	for _, row := range logRows {
		todayCounts[row.ContentType] = row.PV
	}

	items := make([]res.TrafficTrendItem, 0, days)
	for i := 0; i < days; i++ {
		dateKey := startDate.AddDate(0, 0, i).Format("2006-01-02")
		counts := history[dateKey]
		if counts == nil {
			counts = map[string]int64{} // 无数据日期补空 map，避免序列化为 null
		}
		if dateKey == todayKey {
			counts = todayCounts
		}
		total := int64(0)
		for _, pv := range counts {
			total += pv
		}
		items = append(items, res.TrafficTrendItem{Date: dateKey, Counts: counts, TotalPV: total})
	}

	trend := &res.TrafficTrendRes{Range: rangeStr, Items: items}
	_ = l.cache.SetTrafficTrend(ctx, trend)
	return trend, nil
}

// GetModuleStats 获取各模块发布数与访问量汇总（驱动工作台模块数据表）。
func (l *ViewStatsLogic) GetModuleStats(ctx context.Context) (*res.ModuleStatsRes, error) {
	counts, err := l.model.CountContentAll(ctx)
	if err != nil {
		return nil, err
	}

	totalViews, err := l.model.SumContentViewsByType(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -6)

	todayRows, err := l.model.LogsPvByType(ctx, todayStart)
	if err != nil {
		return nil, err
	}
	todayViews := make(map[string]int64, len(todayRows))
	for _, row := range todayRows {
		todayViews[row.ContentType] = row.PV
	}

	weekRows, err := l.model.LogsPvByType(ctx, weekStart)
	if err != nil {
		return nil, err
	}
	weekViews := make(map[string]int64, len(weekRows))
	for _, row := range weekRows {
		weekViews[row.ContentType] = row.PV
	}

	items := make([]res.ModuleStatRow, 0, len(model.ViewTypeOrder()))
	for _, contentType := range model.ViewTypeOrder() {
		count := counts[contentType]
		items = append(items, res.ModuleStatRow{
			Type:       contentType,
			Total:      count.Total,
			Published:  count.Published,
			TotalViews: totalViews[contentType],
			TodayViews: todayViews[contentType],
			WeekViews:  weekViews[contentType],
		})
	}
	return &res.ModuleStatsRes{Items: items}, nil
}
