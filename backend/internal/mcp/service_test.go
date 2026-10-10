package mcp

import (
	"strings"
	"testing"
)

// TestStatusMapping 验证状态字符串与内部状态码的双向映射。
func TestStatusMapping(t *testing.T) {
	cases := []struct {
		in   string
		want int16
	}{
		{"", 0},
		{"draft", 1},
		{"published", 2},
		{"offline", 3},
		{" PUBLISHED ", 2},
	}
	for _, c := range cases {
		got, err := statusFromString(c.in)
		if err != nil {
			t.Fatalf("statusFromString(%q) error = %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("statusFromString(%q) = %d, want %d", c.in, got, c.want)
		}
	}
	if _, err := statusFromString("bogus"); err == nil {
		t.Error("非法状态值应返回错误")
	}
	if statusToString(1) != statusDraft || statusToString(2) != statusPublished || statusToString(3) != statusOffline {
		t.Error("statusToString 映射错误")
	}
}

// TestTruncateSummary 验证默认摘要提取。
func TestTruncateSummary(t *testing.T) {
	long := strings.Repeat("# 标题\n\n正文内容。", 50)
	got := truncateSummary(long)
	if len([]rune(got)) > 201 {
		t.Errorf("摘要应不超过 201 字符（含省略号），实际 %d", len([]rune(got)))
	}
	if !strings.Contains(got, "…") {
		t.Error("超长摘要应以省略号结尾")
	}
	if got := truncateSummary("短文"); strings.Contains(got, "…") {
		t.Error("短文摘要不应有省略号")
	}
}

// TestSanitizeFilename 验证文件名清理。
func TestSanitizeFilename(t *testing.T) {
	cases := []struct{ in, want string }{
		{"cover.png", "cover.png"},
		{"a/b/c.jpg", "c.jpg"},
		{"..\\evil.png", "evil.png"},
		{"", ""},
	}
	for _, c := range cases {
		if got := sanitizeFilename(c.in); got != c.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestExtFromMIME 验证 MIME 到扩展名的映射。
func TestExtFromMIME(t *testing.T) {
	if extFromMIME("image/png") != ".png" || extFromMIME("image/jpeg") != ".jpg" {
		t.Error("常见图片类型映射错误")
	}
	if extFromMIME("application/json") != "" {
		t.Error("非图片类型应返回空扩展名")
	}
}
