package logic

import (
	"context"
	"testing"

	"novablog/internal/dto/req"

	"github.com/stretchr/testify/assert"
)

func TestModuleConfigGetDefault(t *testing.T) {
	// 测试：获取模块配置应返回默认值（全开启）
	// 注意：此测试需要连接数据库，若 DB 不可用则跳过
	logic := NewModuleConfigLogic()
	config, err := logic.GetConfig(context.Background())
	if err != nil {
		t.Skipf("数据库不可用，跳过集成测试: %v", err)
	}

	assert.True(t, config.ArticleEnabled, "文章模块应默认开启")
	assert.True(t, config.MediaEnabled, "媒体模块应默认开启")
	assert.True(t, config.MusicEnabled, "音乐模块应默认开启")
	assert.True(t, config.VideoEnabled, "视频模块应默认开启")
	assert.True(t, config.TravelEnabled, "旅行模块应默认开启")
	assert.True(t, config.PortfolioEnabled, "作品集模块应默认开启")
	assert.True(t, config.EquipmentEnabled, "设备模块应默认开启")
	assert.True(t, config.OpenSourceEnabled, "开源作品模块应默认开启")
}

func TestModuleConfigUpdatePartial(t *testing.T) {
	// 测试：局部更新（只关闭音乐和视频）
	logic := NewModuleConfigLogic()

	// 先获取当前配置（若 DB 不可用则跳过）
	_, err := logic.GetConfig(context.Background())
	if err != nil {
		t.Skipf("数据库不可用，跳过集成测试: %v", err)
	}

	// 只关闭音乐和视频
	musicEnabled := false
	videoEnabled := false
	updateReq := &req.UpdateModuleConfigReq{
		MusicEnabled: &musicEnabled,
		VideoEnabled: &videoEnabled,
	}

	err = logic.UpdateConfig(context.Background(), updateReq)
	assert.NoError(t, err, "局部更新应成功")

	// 验证更新结果
	config, err := logic.GetConfig(context.Background())
	assert.NoError(t, err)
	assert.False(t, config.MusicEnabled, "音乐模块应已关闭")
	assert.False(t, config.VideoEnabled, "视频模块应已关闭")
	assert.True(t, config.ArticleEnabled, "文章模块应保持开启")
	assert.True(t, config.EquipmentEnabled, "设备模块应保持开启")

	// 恢复（全开）
	trueVal := true
	resetReq := &req.UpdateModuleConfigReq{
		MusicEnabled: &trueVal,
		VideoEnabled: &trueVal,
	}
	err = logic.UpdateConfig(context.Background(), resetReq)
	assert.NoError(t, err, "重置应成功")
}
