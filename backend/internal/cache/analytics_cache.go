package cache

import (
	"context"
	"encoding/json"
	"time"

	"novablog/internal/dto/res"

	"github.com/go-redis/redis/v8"
)

// AnalyticsCache 工作台统计缓存操作结构体。
type AnalyticsCache struct {
	client *redis.Client
}

// NewAnalyticsCache 创建 AnalyticsCache 实例。
func NewAnalyticsCache() *AnalyticsCache {
	return &AnalyticsCache{client: RedisClient}
}

const (
	// analyticsOverviewKey v2：OverviewRes 新增扩展模块统计字段后换键，避免旧缓存缺字段被解析为 0
	analyticsOverviewKey = "analytics:overview:v2"
	// analyticsDistributionKey v2：分布统计扩展到全部内容模块后换键，避免旧缓存只剩 5 类
	analyticsDistributionKey = "analytics:distribution:v2"
	// 访问概览/趋势缓存（实时性要求较高，60 秒）
	analyticsTrafficSummaryKey = "analytics:traffic:summary:v1"
	analyticsTrafficTrendKey   = "analytics:traffic:trend:v1:"
	analyticsCacheTTL          = 5 * time.Minute
	analyticsTrafficTTL        = 60 * time.Second
)

// GetOverview 从缓存获取概览数据。
func (c *AnalyticsCache) GetOverview(ctx context.Context) (*res.OverviewRes, error) {
	data, err := c.client.Get(ctx, analyticsOverviewKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var overview res.OverviewRes
	if err := json.Unmarshal([]byte(data), &overview); err != nil {
		return nil, err
	}
	return &overview, nil
}

// SetOverview 缓存概览数据。
func (c *AnalyticsCache) SetOverview(ctx context.Context, overview *res.OverviewRes) error {
	data, err := json.Marshal(overview)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, analyticsOverviewKey, data, analyticsCacheTTL).Err()
}

// GetDistribution 从缓存获取分布数据。
func (c *AnalyticsCache) GetDistribution(ctx context.Context) (*res.DistributionRes, error) {
	data, err := c.client.Get(ctx, analyticsDistributionKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var dist res.DistributionRes
	if err := json.Unmarshal([]byte(data), &dist); err != nil {
		return nil, err
	}
	return &dist, nil
}

// SetDistribution 缓存分布数据。
func (c *AnalyticsCache) SetDistribution(ctx context.Context, dist *res.DistributionRes) error {
	data, err := json.Marshal(dist)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, analyticsDistributionKey, data, analyticsCacheTTL).Err()
}

// GetTrafficSummary 从缓存获取访问概览。
func (c *AnalyticsCache) GetTrafficSummary(ctx context.Context) (*res.TrafficSummaryRes, error) {
	data, err := c.client.Get(ctx, analyticsTrafficSummaryKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var summary res.TrafficSummaryRes
	if err := json.Unmarshal([]byte(data), &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}

// SetTrafficSummary 缓存访问概览。
func (c *AnalyticsCache) SetTrafficSummary(ctx context.Context, summary *res.TrafficSummaryRes) error {
	data, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, analyticsTrafficSummaryKey, data, analyticsTrafficTTL).Err()
}

// trafficTrendRangeKey 访问趋势缓存键（按时间范围区分）。
func trafficTrendRangeKey(rangeStr string) string {
	return analyticsTrafficTrendKey + rangeStr
}

// GetTrafficTrend 从缓存获取访问趋势。
func (c *AnalyticsCache) GetTrafficTrend(ctx context.Context, rangeStr string) (*res.TrafficTrendRes, error) {
	data, err := c.client.Get(ctx, trafficTrendRangeKey(rangeStr)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var trend res.TrafficTrendRes
	if err := json.Unmarshal([]byte(data), &trend); err != nil {
		return nil, err
	}
	return &trend, nil
}

// SetTrafficTrend 缓存访问趋势。
func (c *AnalyticsCache) SetTrafficTrend(ctx context.Context, trend *res.TrafficTrendRes) error {
	data, err := json.Marshal(trend)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, trafficTrendRangeKey(trend.Range), data, analyticsTrafficTTL).Err()
}
