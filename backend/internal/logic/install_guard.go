// Package logic 首装安装码校验与防爆破（官方部署专用）。
// 官方部署在 config.yaml 配置 install.init_code（deploy.sh --install-code 写入），
// 首装初始化必须携带该密钥，防止知道实例地址的恶意者抢先初始化；自部署不配置即不启用。
package logic

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"novablog/enum"
	"novablog/internal/cache"
)

// InstallInitCode 官方部署安装码（bootstrap 从 config.yaml install.init_code 注入）；
// 空字符串表示自部署，不启用校验。
var InstallInitCode string

// 防爆破参数：每 IP 在窗口内最多允许 installFailMax 次密钥校验失败。
// 安装码本身为 128 位随机熵，爆破在数学上不可行，限流仅作纵深防御。
const (
	installFailWindow = time.Hour
	installFailMax    = 10
)

// installFailKey 安装码失败计数的 Redis key（按 IP 隔离）。
func installFailKey(ip string) string {
	return fmt.Sprintf("install:fail:%s", ip)
}

// verifyInstallCode 常量时间比对安装码；未配置（自部署）直接放行。
func verifyInstallCode(input string) error {
	if InstallInitCode == "" {
		return nil
	}
	if subtle.ConstantTimeCompare([]byte(input), []byte(InstallInitCode)) != 1 {
		return enum.ErrInstallCodeInvalid
	}
	return nil
}

// checkInstallAttempts 校验前检查该 IP 失败次数是否已达上限。
// Redis 不可用或 IP 为空时放行（与 mcp_rate_limit 的 fail-open 语义一致）。
func checkInstallAttempts(ctx context.Context, ip string) error {
	client := cache.RedisClient
	if client == nil || ip == "" {
		return nil
	}
	count, err := client.Get(ctx, installFailKey(ip)).Int64()
	if err != nil {
		// redis.Nil（无记录）与其他错误均放行
		return nil
	}
	if count >= installFailMax {
		return enum.ErrInstallCodeRateLimited
	}
	return nil
}

// recordInstallFail 记录一次校验失败，窗口内首次失败时设置过期时间。
func recordInstallFail(ctx context.Context, ip string) {
	client := cache.RedisClient
	if client == nil || ip == "" {
		return
	}
	key := installFailKey(ip)
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		return
	}
	if count == 1 {
		client.Expire(ctx, key, installFailWindow)
	}
}

// clearInstallFails 校验成功后清除失败计数。
func clearInstallFails(ctx context.Context, ip string) {
	if client := cache.RedisClient; client != nil && ip != "" {
		client.Del(ctx, installFailKey(ip))
	}
}

// EnforceInstallCode 首装初始化的安装码校验入口：未配置直接放行；
// 配置了则先查防爆破计数，再做常量时间比对（失败记账，成功清账）。
func EnforceInstallCode(ctx context.Context, input, ip string) error {
	if InstallInitCode == "" {
		return nil
	}
	if err := checkInstallAttempts(ctx, ip); err != nil {
		return err
	}
	if err := verifyInstallCode(input); err != nil {
		recordInstallFail(ctx, ip)
		return err
	}
	clearInstallFails(ctx, ip)
	return nil
}
