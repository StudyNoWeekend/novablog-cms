package geoip

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
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

// TestLoadAsyncDownload 数据文件缺失时 Load 必须立即返回（转后台下载），
// 且后台 goroutine 最终完成下载落盘。测试服务端延迟 1.5s 再返回数据：
// 若实现退化为同步下载，Load 至少阻塞 1.5s，"1s 内返回"断言即失败。
func TestLoadAsyncDownload(t *testing.T) {
	payload := bytes.Repeat([]byte{0x00}, 2<<20) // 2MB ≥ minDataFileSize(1MB)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1500 * time.Millisecond)
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	dir := t.TempDir()
	full := filepath.Join(dir, dataFileName)
	s := New(Config{DataDir: dir, DownloadURL: srv.URL + "/ip2region_v4.xdb"}, nil)

	start := time.Now()
	if err := s.Load(context.Background()); err != nil {
		t.Fatalf("Load 缺失文件时应转后台并立即返回，却报错: %v", err)
	}
	if elapsed := time.Since(start); elapsed >= time.Second {
		t.Fatalf("Load 耗时 %v，下载阻塞了启动（未实现异步）", elapsed)
	}

	// 后台 goroutine 最终完成下载并落盘（tmp 改名为正式文件）
	deadline := time.Now().Add(5 * time.Second)
	for {
		if fi, err := os.Stat(full); err == nil && fi.Size() == int64(len(payload)) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("后台下载未在 5s 内完成落盘")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// TestLoadSyncWhenFileExists 文件已存在时走同步加载（不转后台下载）：
// 用足够大但非法的内容触发 load 解析错误，错误必须同步返回；
// 若实现错误地转了后台下载，Load 会返回 nil（下载 URL 不可达也只会留下告警日志）。
func TestLoadSyncWhenFileExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, dataFileName), bytes.Repeat([]byte{0x00}, 2<<20), 0o644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}
	s := New(Config{DataDir: dir, DownloadURL: "http://invalid.invalid/xdb"}, nil)
	if err := s.Load(context.Background()); err == nil {
		t.Fatal("文件已存在时应同步加载；损坏内容应同步返回错误而非静默转后台")
	}
}
