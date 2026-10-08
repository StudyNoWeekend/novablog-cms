package logic

import (
	"testing"

	"novablog/enum"

	"github.com/stretchr/testify/assert"
)

func TestListZodiacConfigs(t *testing.T) {
	// 测试：星座选项应包含 12 项，按日期顺序排列，字段完整
	list := ListZodiacConfigs()
	assert.Len(t, list, 12, "星座选项应为 12 项")

	assert.Equal(t, "aries", list[0].Key, "第一项应为白羊座")
	assert.Equal(t, "capricorn", list[9].Key, "第十项应为摩羯座")
	assert.Equal(t, "pisces", list[11].Key, "最后一项应为双鱼座")

	for _, cfg := range list {
		assert.NotEmpty(t, cfg.Key, "星座 key 不应为空")
		assert.NotEmpty(t, cfg.Name, "星座名称不应为空")
		assert.NotEmpty(t, cfg.Image, "星座图片 SVG 片段不应为空")
		assert.NotEmpty(t, cfg.DateRange, "星座日期范围不应为空")
		assert.Contains(t, []string{"火象", "土象", "风象", "水象"}, cfg.Element, "星座类型应为四象之一")
	}

	// 四象各 3 个
	elementCount := map[string]int{}
	for _, cfg := range list {
		elementCount[cfg.Element]++
	}
	assert.Equal(t, map[string]int{"火象": 3, "土象": 3, "风象": 3, "水象": 3}, elementCount, "四象应各含 3 个星座")
}

func TestListPersonalityConfigs(t *testing.T) {
	// 测试：性格选项应包含 MBTI 16 型，字段完整
	list := ListPersonalityConfigs()
	assert.Len(t, list, 16, "性格选项应为 MBTI 16 型")

	seen := map[string]bool{}
	for _, cfg := range list {
		assert.NotEmpty(t, cfg.Key, "MBTI key 不应为空")
		assert.NotEmpty(t, cfg.Name, "性格名称不应为空")
		assert.NotEmpty(t, cfg.Image, "性格图片 SVG 片段不应为空")
		assert.NotEmpty(t, cfg.Description, "性格介绍不应为空")
		assert.Contains(t, cfg.Image, cfg.Key, "徽章 SVG 应包含 MBTI 四字母")
		assert.False(t, seen[cfg.Key], "MBTI key 不应重复")
		seen[cfg.Key] = true
	}
	assert.Len(t, seen, 16, "MBTI key 应覆盖 16 种类型")
}

func TestGetZodiacConfig(t *testing.T) {
	// 测试：按 key 获取星座配置，未知 key 返回 false
	cfg, ok := GetZodiacConfig("leo")
	assert.True(t, ok)
	assert.Equal(t, "狮子座", cfg.Name)
	assert.Equal(t, "7.23-8.22", cfg.DateRange)
	assert.Equal(t, "火象", cfg.Element)

	_, ok = GetZodiacConfig("not-exist")
	assert.False(t, ok, "未知星座 key 应返回 false")
}

func TestGetPersonalityConfig(t *testing.T) {
	// 测试：按 key 获取性格配置，未知 key 返回 false
	cfg, ok := GetPersonalityConfig("INTJ")
	assert.True(t, ok)
	assert.Equal(t, "建筑师", cfg.Name)
	assert.NotEmpty(t, cfg.Description)

	_, ok = GetPersonalityConfig("ABCD")
	assert.False(t, ok, "未知性格 key 应返回 false")
}

func TestValidateProfileMetaKey(t *testing.T) {
	// 测试：空串合法（表示清除）
	assert.NoError(t, validateProfileMetaKey(metaKindPersonality, ""))
	assert.NoError(t, validateProfileMetaKey(metaKindZodiac, ""))

	// 测试：合法 key
	assert.NoError(t, validateProfileMetaKey(metaKindPersonality, "INFJ"))
	assert.NoError(t, validateProfileMetaKey(metaKindZodiac, "taurus"))

	// 测试：非法 key 返回参数错误
	err := validateProfileMetaKey(metaKindPersonality, "ABCD")
	assert.Error(t, err)
	var bizErr *enum.BizError
	if assert.ErrorAs(t, err, &bizErr) {
		assert.Equal(t, enum.ErrInvalidParam.Code, bizErr.Code)
	}

	err = validateProfileMetaKey(metaKindZodiac, "not-exist")
	assert.Error(t, err)
	assert.ErrorAs(t, err, &bizErr)
	assert.Equal(t, enum.ErrInvalidParam.Code, bizErr.Code)
}

func TestBuildZodiacPublicMeta(t *testing.T) {
	// 测试：开关关闭 → nil（公开接口返回 null）
	assert.Nil(t, buildZodiacPublicMeta(false, "leo"))

	// 测试：未设置 → nil
	assert.Nil(t, buildZodiacPublicMeta(true, ""))

	// 测试：未知 key → nil（不暴露错误数据）
	assert.Nil(t, buildZodiacPublicMeta(true, "hacker"))

	// 测试：开启且已设置 → 返回完整 meta
	meta := buildZodiacPublicMeta(true, "leo")
	if assert.NotNil(t, meta) {
		assert.Equal(t, "leo", meta.Key)
		assert.Equal(t, "狮子座", meta.Name)
		assert.Equal(t, "7.23-8.22", meta.DateRange)
		assert.Equal(t, "火象", meta.Element)
		assert.NotEmpty(t, meta.Image)
	}
}

func TestBuildPersonalityPublicMeta(t *testing.T) {
	// 测试：开关关闭 → nil（公开接口返回 null）
	assert.Nil(t, buildPersonalityPublicMeta(false, "INTJ"))

	// 测试：未设置 → nil
	assert.Nil(t, buildPersonalityPublicMeta(true, ""))

	// 测试：未知 key → nil
	assert.Nil(t, buildPersonalityPublicMeta(true, "ABCD"))

	// 测试：开启且已设置 → 返回完整 meta
	meta := buildPersonalityPublicMeta(true, "INTJ")
	if assert.NotNil(t, meta) {
		assert.Equal(t, "INTJ", meta.Key)
		assert.Equal(t, "建筑师", meta.Name)
		assert.NotEmpty(t, meta.Image)
		assert.NotEmpty(t, meta.Description)
	}
}

func TestGetProfileMeta(t *testing.T) {
	// 测试：选项元数据接口数据源完整（星座 12 项 + 性格 16 项）
	meta := NewBloggerLogic().GetProfileMeta()
	assert.Len(t, meta.Zodiac, 12)
	assert.Len(t, meta.Personality, 16)
	for _, z := range meta.Zodiac {
		assert.NotEmpty(t, z.Image)
	}
	for _, p := range meta.Personality {
		assert.Contains(t, p.Image, p.Key)
	}
}
