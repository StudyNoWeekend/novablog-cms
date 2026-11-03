package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheKeyInstallFail 首装安装码失败计数 key 前缀（按 IP 隔离）。
const CacheKeyInstallFail = "install:fail"

// installFailKey 生成失败计数 key。
func installFailKey(ip string) string {
	return fmt.Sprintf("%s:%s", CacheKeyInstallFail, ip)
}

// InstallGuardCache 首装安装码防爆破的 Redis 计数操作。
// 全部方法在 Redis 未初始化或 IP 为空时静默跳过（fail-open 由调用方语义保证）。
type InstallGuardCache struct {
	client *redis.Client
}

// NewInstallGuardCache 创建 InstallGuardCache 实例。
func NewInstallGuardCache() *InstallGuardCache {
	return &InstallGuardCache{client: RedisClient}
}

// GetFailCount 返回 IP 的失败计数；无记录或不可用时返回 0。
func (c *InstallGuardCache) GetFailCount(ctx context.Context, ip string) int64 {
	if c.client == nil || ip == "" {
		return 0
	}
	count, err := c.client.Get(ctx, installFailKey(ip)).Int64()
	if err != nil {
		return 0
	}
	return count
}

// RecordFail 记录一次校验失败，窗口内首次失败时设置过期时间。
func (c *InstallGuardCache) RecordFail(ctx context.Context, ip string, window time.Duration) {
	if c.client == nil || ip == "" {
		return
	}
	key := installFailKey(ip)
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return
	}
	if count == 1 {
		c.client.Expire(ctx, key, window)
	}
}

// ClearFails 校验成功后清除失败计数。
func (c *InstallGuardCache) ClearFails(ctx context.Context, ip string) {
	if c.client == nil || ip == "" {
		return
	}
	c.client.Del(ctx, installFailKey(ip))
}
