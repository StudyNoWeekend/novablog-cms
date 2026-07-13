package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// CacheKeySecurityConfig 安全配置缓存键前缀
	CacheKeySecurityConfig = "security:config"
	// CacheKeyRateLimit 限流缓存键前缀
	CacheKeyRateLimit = "security:ratelimit"
	// CacheKeyBlacklist IP 黑名单缓存键前缀
	CacheKeyBlacklist = "security:blacklist"
	// CacheKeyViolation 违规计数缓存键前缀
	CacheKeyViolation = "security:violation"
)

// SecurityConfigCache 安全配置缓存结构体，用于 Redis 缓存序列化。
type SecurityConfigCache struct {
	GetMaxTokens        int `json:"get_max_tokens"`        // GET 请求最大令牌数
	GetWindowSeconds    int `json:"get_window_seconds"`    // GET 请求时间窗口（秒）
	PostMaxTokens       int `json:"post_max_tokens"`       // POST 请求最大令牌数
	PostWindowSeconds   int `json:"post_window_seconds"`   // POST 请求时间窗口（秒）
	ViewMaxTokens       int `json:"view_max_tokens"`       // 浏览相关接口最大令牌数
	ViewWindowSeconds   int `json:"view_window_seconds"`   // 浏览相关接口时间窗口（秒）
	LikeMaxTokens       int `json:"like_max_tokens"`       // 点赞相关接口最大令牌数
	LikeWindowSeconds   int `json:"like_window_seconds"`   // 点赞相关接口时间窗口（秒）
	BlacklistThreshold  int `json:"blacklist_threshold"`   // 触发黑名单的违规次数阈值
	BlacklistTTLMinutes int `json:"blacklist_ttl_minutes"` // 黑名单封禁时长（分钟）
}

// SecurityCache 安全相关缓存操作结构体。
type SecurityCache struct {
	client *redis.Client
}

// NewSecurityCache 创建 SecurityCache 实例。
func NewSecurityCache() *SecurityCache {
	return &SecurityCache{client: RedisClient}
}

// GetConfig 从缓存获取安全配置。
func (c *SecurityCache) GetConfig(ctx context.Context) (*SecurityConfigCache, error) {
	data, err := c.client.Get(ctx, CacheKeySecurityConfig).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var config SecurityConfigCache
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetConfig 将安全配置写入缓存。
func (c *SecurityCache) SetConfig(ctx context.Context, config *SecurityConfigCache, ttl time.Duration) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, CacheKeySecurityConfig, data, ttl).Err()
}

// DeleteConfig 删除安全配置缓存（配置更新时调用）。
func (c *SecurityCache) DeleteConfig(ctx context.Context) error {
	return c.client.Del(ctx, CacheKeySecurityConfig).Err()
}

// AllowRequest 令牌桶限流：基于 Redis INCR 实现固定窗口计数，
// key 格式为 security:ratelimit:{route}:{ip}:{window}，
// 首次计数时设置过期时间为窗口时长，返回当前计数是否不超过最大令牌数。
func (c *SecurityCache) AllowRequest(ctx context.Context, ip, route string, maxTokens int, windowSeconds int) (bool, error) {
	windowBucket := time.Now().Unix() / int64(windowSeconds)
	key := fmt.Sprintf("%s:%s:%s:%d", CacheKeyRateLimit, route, ip, windowBucket)
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		if err := c.client.Expire(ctx, key, time.Duration(windowSeconds)*time.Second).Err(); err != nil {
			return false, err
		}
	}
	return count <= int64(maxTokens), nil
}

// IsBlacklisted 检查 IP 是否在黑名单中。
func (c *SecurityCache) IsBlacklisted(ctx context.Context, ip string) (bool, error) {
	key := fmt.Sprintf("%s:%s", CacheKeyBlacklist, ip)
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// AddToBlacklist 将 IP 加入黑名单，设置过期时间。
func (c *SecurityCache) AddToBlacklist(ctx context.Context, ip string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", CacheKeyBlacklist, ip)
	return c.client.Set(ctx, key, "1", ttl).Err()
}

// RemoveFromBlacklist 将 IP 从黑名单中移除。
func (c *SecurityCache) RemoveFromBlacklist(ctx context.Context, ip string) error {
	key := fmt.Sprintf("%s:%s", CacheKeyBlacklist, ip)
	return c.client.Del(ctx, key).Err()
}

// GetBlacklistCount 统计当前黑名单中的 IP 数量。
func (c *SecurityCache) GetBlacklistCount(ctx context.Context) (int64, error) {
	pattern := fmt.Sprintf("%s:*", CacheKeyBlacklist)
	var count int64
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		count++
	}
	if err := iter.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

// IncrementViolation 递增 IP 的违规计数，首次递增时设置 24 小时过期。
func (c *SecurityCache) IncrementViolation(ctx context.Context, ip string) (int64, error) {
	key := fmt.Sprintf("%s:%s", CacheKeyViolation, ip)
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := c.client.Expire(ctx, key, 24*time.Hour).Err(); err != nil {
			return 0, err
		}
	}
	return count, nil
}

// GetViolationCount 获取 IP 的违规计数。
func (c *SecurityCache) GetViolationCount(ctx context.Context, ip string) (int64, error) {
	key := fmt.Sprintf("%s:%s", CacheKeyViolation, ip)
	count, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ResetViolations 重置 IP 的违规计数。
func (c *SecurityCache) ResetViolations(ctx context.Context, ip string) error {
	key := fmt.Sprintf("%s:%s", CacheKeyViolation, ip)
	return c.client.Del(ctx, key).Err()
}

// IncrementDailyRateLimitLog 递增每日限流触发次数，用于统计。
func (c *SecurityCache) IncrementDailyRateLimitLog(ctx context.Context, date string) error {
	key := fmt.Sprintf("%s:daily:%s", CacheKeyRateLimit, date)
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		return c.client.Expire(ctx, key, 48*time.Hour).Err()
	}
	return nil
}

// GetDailyRateLimitCount 获取指定日期的限流触发次数。
func (c *SecurityCache) GetDailyRateLimitCount(ctx context.Context, date string) (int64, error) {
	key := fmt.Sprintf("%s:daily:%s", CacheKeyRateLimit, date)
	count, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}
