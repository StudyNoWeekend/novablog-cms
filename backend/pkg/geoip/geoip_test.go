package geoip

import (
	"testing"
)

func TestFormatRegion(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"中国省份城市运营商", "中国|0|广东省|深圳市|电信", "中国 广东省 深圳市"},
		{"无运营商", "中国|0|浙江省|杭州市|0", "中国 浙江省 杭州市"},
		{"区域有值", "中国|华南|广东省|广州市|移动", "中国 广东省 广州市"},
		{"只到国家", "美国|0|0|0|0", "美国"},
		{"市级重复省", "中国|0|北京市|北京市|联通", "中国 北京市"},
		{"直辖市无市级", "中国|0|北京市|0|0", "中国 北京市"},
		{"内网IP", "0|0|0|内网IP|内网IP", "内网IP"},
		{"Reserved英文保留", "0|0|0|Reserved|APNIC", "Reserved"},
		{"国外国家加组织", "美国|0|0|0|Google LLC", "美国 Google LLC"},
		{"国外仅国家", "美国|0|0|0|0", "美国"},
		{"IPv4保留", "0|0|0|0|0", ""},
		{"纯0带运营商", "0|0|0|0|电信", "电信"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := formatRegion(c.raw); got != c.want {
				t.Fatalf("formatRegion(%q) = %q, want %q", c.raw, got, c.want)
			}
		})
	}
}
