// Package logic 安装码校验单元测试。
// 纯函数用例不依赖 DB/Redis；防爆破用例连接本地 Redis（127.0.0.1，DB 14），不可用时自动跳过。
package logic

import (
	"context"
	"testing"
	"time"

	"novablog/enum"
	"novablog/internal/cache"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

// withInstallCode 临时设置安装码，测试结束后还原，避免污染其他测试。
func withInstallCode(t *testing.T, code string) {
	t.Helper()
	old := InstallInitCode
	InstallInitCode = code
	t.Cleanup(func() { InstallInitCode = old })
}

// TestVerifyInstallCodeDisabled 自部署（未配置安装码）时任何输入都放行。
func TestVerifyInstallCodeDisabled(t *testing.T) {
	withInstallCode(t, "")

	assert.NoError(t, verifyInstallCode(""))
	assert.NoError(t, verifyInstallCode("anything"))
}

// TestVerifyInstallCodeEnabled 已配置安装码：正确放行；空/错误/大小写不同均拒绝。
func TestVerifyInstallCodeEnabled(t *testing.T) {
	withInstallCode(t, "nbi_0f3a9c8b7d6e5f4a3b2c1d0e9f8a7b6c")

	assert.NoError(t, verifyInstallCode("nbi_0f3a9c8b7d6e5f4a3b2c1d0e9f8a7b6c"))
	assert.ErrorIs(t, verifyInstallCode(""), enum.ErrInstallCodeInvalid)
	assert.ErrorIs(t, verifyInstallCode("wrong-code"), enum.ErrInstallCodeInvalid)
	// 大小写敏感，防止彩虹表式变体碰撞
	assert.ErrorIs(t, verifyInstallCode("NBI_0F3A9C8B7D6E5F4A3B2C1D0E9F8A7B6C"), enum.ErrInstallCodeInvalid)
}

// TestEnforceInstallCodeFailOpen 安装码校验不依赖 Redis：Redis 不可用时跳过防爆破计数，
// 但密钥本身仍正常比对（fail-open 只作用于限流，不作用于校验）。
func TestEnforceInstallCodeFailOpen(t *testing.T) {
	withInstallCode(t, "nbi_0f3a9c8b7d6e5f4a3b2c1d0e9f8a7b6c")

	old := cache.RedisClient
	cache.RedisClient = nil
	t.Cleanup(func() { cache.RedisClient = old })

	ctx := context.Background()
	assert.NoError(t, EnforceInstallCode(ctx, "nbi_0f3a9c8b7d6e5f4a3b2c1d0e9f8a7b6c", "203.0.113.10"))
	assert.ErrorIs(t, EnforceInstallCode(ctx, "wrong-code", "203.0.113.10"), enum.ErrInstallCodeInvalid)
}

// TestCheckInstallAttemptsEmptyIP 空 IP 直接放行（无法定位来源时宁可不拦）。
func TestCheckInstallAttemptsEmptyIP(t *testing.T) {
	assert.NoError(t, checkInstallAttempts(context.Background(), ""))
}

// setupInstallGuardRedis 连接本地 Redis（DB 14，独立于其他测试），不可用则跳过。
func setupInstallGuardRedis(t *testing.T) context.Context {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 14})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("本地 Redis 不可用，跳过防爆破集成测试 (%v)", err)
	}
	old := cache.RedisClient
	cache.RedisClient = rdb
	t.Cleanup(func() {
		cache.RedisClient = old
		_ = rdb.Close()
	})
	return context.Background()
}

// TestEnforceInstallCodeRateLimit 连续失败达到上限后，即使输入正确密钥也被限流；
// 计数清除后恢复正常。
func TestEnforceInstallCodeRateLimit(t *testing.T) {
	withInstallCode(t, "nbi_0f3a9c8b7d6e5f4a3b2c1d0e9f8a7b6c")
	ctx := setupInstallGuardRedis(t)

	const ip = "203.0.113.77"
	clearInstallFails(ctx, ip)
	t.Cleanup(func() { clearInstallFails(ctx, ip) })

	// 上限内的失败只返回密钥错误，不触发限流
	for i := 0; i < installFailMax; i++ {
		err := EnforceInstallCode(ctx, "wrong", ip)
		assert.ErrorIs(t, err, enum.ErrInstallCodeInvalid, "第 %d 次失败", i+1)
	}

	// 达到上限：正确密钥也被拒绝为限流
	assert.ErrorIs(t, EnforceInstallCode(ctx, "nbi_0f3a9c8b7d6e5f4a3b2c1d0e9f8a7b6c", ip), enum.ErrInstallCodeRateLimited)

	// 计数过期/清除后恢复
	clearInstallFails(ctx, ip)
	assert.NoError(t, EnforceInstallCode(ctx, "nbi_0f3a9c8b7d6e5f4a3b2c1d0e9f8a7b6c", ip))
}

// TestInstallFailKeyTTL 失败计数必须带过期时间，避免半开状态下永久封禁。
func TestInstallFailKeyTTL(t *testing.T) {
	ctx := setupInstallGuardRedis(t)

	const ip = "203.0.113.88"
	clearInstallFails(ctx, ip)
	t.Cleanup(func() { clearInstallFails(ctx, ip) })

	recordInstallFail(ctx, ip)
	ttl, err := cache.RedisClient.TTL(ctx, installFailKey(ip)).Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, installFailWindow)
}
