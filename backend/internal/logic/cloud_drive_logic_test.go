package logic

import (
	"testing"
	"time"

	"novablog/internal/model"

	"github.com/stretchr/testify/assert"
)

// TestNeedRefresh 验证 token 刷新判定：无过期时间/剩余不足 10 分钟 → 需要刷新。
func TestNeedRefresh(t *testing.T) {
	l := &CloudDriveLogic{}

	// 无过期时间 → 刷新
	assert.True(t, l.needRefresh(&model.CloudDriveAccount{}))

	// 剩余 1 小时（百度 30 天有效期场景）→ 不刷新
	future := time.Now().Add(time.Hour)
	assert.False(t, l.needRefresh(&model.CloudDriveAccount{ExpiresAt: &future}))

	// 剩余 5 分钟 → 刷新
	soon := time.Now().Add(5 * time.Minute)
	assert.True(t, l.needRefresh(&model.CloudDriveAccount{ExpiresAt: &soon}))

	// 已过期 → 刷新
	past := time.Now().Add(-time.Minute)
	assert.True(t, l.needRefresh(&model.CloudDriveAccount{ExpiresAt: &past}))
}

// TestWebDAVOpenUserID 验证 WebDAV 唯一标识拼接。
func TestWebDAVOpenUserID(t *testing.T) {
	assert.Equal(t,
		"https://dav.jianguoyun.com/dav|alice",
		webDAVOpenUserID("https://dav.jianguoyun.com/dav", "alice"),
	)
}

// TestCallbackURI 验证回调地址拼接（去除尾斜杠）。
func TestCallbackURI(t *testing.T) {
	assert.Equal(t,
		"https://cms.example.com/api/v1/clouddrive/callback/baidu",
		callbackURI("https://cms.example.com", "baidu"),
	)
	assert.Equal(t,
		"https://cms.example.com/api/v1/clouddrive/callback/aliyun",
		callbackURI("https://cms.example.com/", "aliyun"),
	)
}

// TestMaskSecret 验证凭据脱敏。
func TestMaskSecret(t *testing.T) {
	assert.Empty(t, maskSecret(""))
	assert.Equal(t, "******", maskSecret("encrypted-payload"))
}
