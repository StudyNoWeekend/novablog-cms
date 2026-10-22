package mcp

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"novablog/internal/dto/res"
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

// TestValidateMonth 验证 YYYY-MM 月份格式校验。
func TestValidateMonth(t *testing.T) {
	valid := []string{"", "2024-01", "2024-12", "1999-06"}
	for _, v := range valid {
		if err := validateMonth(v, "start_date"); err != nil {
			t.Errorf("validateMonth(%q) 不应报错，实际 %v", v, err)
		}
	}
	invalid := []string{"2024-1", "2024/01", "24-01", "2024-01-01", "abcd-01", "2024-ab", "2024-13x"}
	for _, v := range invalid {
		if err := validateMonth(v, "start_date"); err == nil {
			t.Errorf("validateMonth(%q) 应返回错误", v)
		}
	}
}

// TestStringPtr 验证空串转 nil 的可选参数转换。
func TestStringPtr(t *testing.T) {
	if stringPtr("") != nil {
		t.Error("空串应返回 nil")
	}
	if s := stringPtr("abc"); s == nil || *s != "abc" {
		t.Error("非空串应返回指向原值的指针")
	}
	if trimSpacePtr("  x ") == nil || *trimSpacePtr("  x ") != "x" {
		t.Error("trimSpacePtr 应去除首尾空白")
	}
	if trimSpacePtr("   ") != nil {
		t.Error("全空白串应返回 nil")
	}
}

// TestFlattenFolderTree 验证文件夹树拍平的深度、路径拼接。
func TestFlattenFolderTree(t *testing.T) {
	parent := "root"
	tree := []res.MediaFolderRes{
		{ID: "a", Name: "文章", ParentID: &parent, MediaCount: 3, Children: []res.MediaFolderRes{
			{ID: "b", Name: "封面", MediaCount: 2},
		}},
		{ID: "c", Name: "菜谱", MediaCount: 1},
	}
	items := flattenFolderTree(tree)
	if len(items) != 3 {
		t.Fatalf("应拍平为 3 个节点，实际 %d", len(items))
	}
	if items[0].Path != "文章" || items[0].Depth != 0 {
		t.Errorf("根节点 path/depth 错误: %+v", items[0])
	}
	if items[1].Path != "文章 / 封面" || items[1].Depth != 1 {
		t.Errorf("子节点 path/depth 错误: %+v", items[1])
	}
}

// TestToolsList 通过真实 Streamable HTTP 链路拉取 tools/list，
// 兜底捕获注册期才会暴露的问题（jsonschema 循环类型、非法 tag、重名工具等）。
func TestToolsList(t *testing.T) {
	svc := NewService("test", nil, nil)
	srv := httptest.NewServer(svc.Handler())
	defer srv.Close()

	ctx := context.Background()
	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "0"}, nil)
	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{
		Endpoint:             srv.URL,
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("连接 MCP 服务失败: %v", err)
	}
	defer session.Close()

	res, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("tools/list 失败: %v", err)
	}
	if len(res.Tools) < 90 {
		t.Errorf("工具数量应不少于 90，实际 %d", len(res.Tools))
	}
	seen := make(map[string]bool, len(res.Tools))
	for _, tool := range res.Tools {
		if tool.Name == "" {
			t.Fatal("存在未命名工具")
		}
		if tool.InputSchema == nil {
			t.Errorf("工具 %s 缺少 InputSchema", tool.Name)
		}
		if tool.Description == "" {
			t.Errorf("工具 %s 缺少 Description", tool.Name)
		}
		if seen[tool.Name] {
			t.Errorf("工具 %s 重复注册", tool.Name)
		}
		seen[tool.Name] = true
	}
	if !seen["create_article"] || !seen["upload_image"] || !seen["list_media_folders"] {
		t.Error("核心工具缺失")
	}
}
