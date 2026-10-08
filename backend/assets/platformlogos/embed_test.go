package platformlogos

import (
	"strings"
	"testing"
)

// TestGetKnownPlatforms 内置平台应能取到非空 SVG 内容。
func TestGetKnownPlatforms(t *testing.T) {
	for _, platform := range []string{"qq_music", "netease", "bilibili", "spotify", "apple_music", "other"} {
		data, ok := Get(platform)
		if !ok {
			t.Errorf("平台 %s 应有内置 Logo", platform)
			continue
		}
		svg := string(data)
		if !strings.HasPrefix(svg, "<svg") || !strings.Contains(svg, "fill=\"#fff\"") {
			t.Errorf("平台 %s 的 Logo 应为白色填充 SVG，实际：%s", platform, svg[:min(len(svg), 80)])
		}
	}
}

// TestGetUnknownPlatform 未知平台应返回 false。
func TestGetUnknownPlatform(t *testing.T) {
	for _, platform := range []string{"", "unknown", "NETEASE", "netease.svg"} {
		if _, ok := Get(platform); ok {
			t.Errorf("平台 %q 不应有内置 Logo", platform)
		}
	}
}

// TestURLPath URLPath 应返回与文件名一致的公开路径，未知平台返回空串。
func TestURLPath(t *testing.T) {
	if got := URLPath("netease"); got != "/api/v1/public/platform-logos/netease.svg" {
		t.Errorf("netease Logo 路径不符，实际：%s", got)
	}
	if got := URLPath("nope"); got != "" {
		t.Errorf("未知平台 Logo 路径应为空串，实际：%s", got)
	}
}
