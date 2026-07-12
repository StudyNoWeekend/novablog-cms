package cache

import (
	"context"
	"encoding/json"
	"time"

	"cus-cms/internal/dto/res"

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
	analyticsOverviewKey     = "analytics:overview"
	analyticsDistributionKey = "analytics:distribution"
	analyticsCacheTTL        = 5 * time.Minute
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
