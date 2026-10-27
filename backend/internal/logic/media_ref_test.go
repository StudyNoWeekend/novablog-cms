package logic

import "testing"

// TestCategoryRefPtr 验证分类引用归一化：未选择（nil/空串/纯空白）置为 NULL，
// 非空原样透传——uuid 列不接受空串（22P02）。
func TestCategoryRefPtr(t *testing.T) {
	empty := ""
	blank := "   "
	valid := "56f731f0-6897-48a9-b1bb-cf7de31cb793"

	cases := []struct {
		name string
		in   *string
		want *string
	}{
		{"nil 保持 nil", nil, nil},
		{"空串置为 NULL", &empty, nil},
		{"纯空白置为 NULL", &blank, nil},
		{"合法 UUID 原样透传", &valid, &valid},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := categoryRefPtr(c.in)
			if (got == nil) != (c.want == nil) {
				t.Fatalf("categoryRefPtr(%v) = %v, want %v", c.in, got, c.want)
			}
			if got != nil && *got != *c.want {
				t.Fatalf("categoryRefPtr(%v) = %q, want %q", c.in, *got, *c.want)
			}
		})
	}
}
