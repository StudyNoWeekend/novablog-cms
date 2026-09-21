// Package geoip 提供 IP → 地区(国家/省/市) 的离线解析能力。
//
// 基于 ip2region 的 xdb 数据文件（v4），数据文件缺失时自动从
// DownloadURL 下载（支持通过 ProxyURL 走代理），下载/加载失败时
// 自动降级：Lookup 返回空字符串，不影响访问日志写入主流程。
package geoip

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"novablog/pkg/proxyhttp"

	"github.com/lionsoul2014/ip2region/binding/golang/service"
	"go.uber.org/zap"
)

// dataFileName xdb 数据文件名。
const dataFileName = "ip2region_v4.xdb"

// minDataFileSize 最小合法数据文件大小，用于校验下载结果（GitHub 的 404 页远小于此值）。
const minDataFileSize int64 = 1 << 20 // 1MB

// Config geoip 模块配置。
type Config struct {
	// DataDir xdb 数据文件所在目录，缺失时自动创建并下载。
	DataDir string
	// DownloadURL 数据文件下载地址；DataDir 下无文件且此值为空时仅降级并告警。
	DownloadURL string
	// ProxyURL 下载所用的 HTTP 代理（可空，复用主题模块代理语义）。
	ProxyURL string
}

// Searcher IP 地区解析器，并发安全。
// 底层 ip2region service 自带 searcher 池，可被并发调用。
type Searcher struct {
	mu     sync.RWMutex
	region *service.Ip2Region
	cfg    Config
	log    *zap.Logger

	// loaded 记录是否已完成一次加载尝试（成功或失败），避免每次 Lookup 重复触发下载。
	loaded bool
}

// New 创建解析器（不做加载，由 Load 触发）。
func New(cfg Config, log *zap.Logger) *Searcher {
	if log == nil {
		log = zap.NewNop()
	}
	return &Searcher{cfg: cfg, log: log}
}

// Load 加载 xdb 数据文件。数据文件不存在时自动下载；加载/下载失败
// 返回 error，但调用方不应因此中断启动——Lookup 会自动降级返回空地区。
func (s *Searcher) Load(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.loaded {
		return nil
	}

	full := filepath.Join(s.cfg.DataDir, dataFileName)

	// 已存在且大小合法 → 直接加载
	if ok, _ := validDataFile(full); ok {
		s.loaded = true
		return s.load(full)
	}

	// 缺失/文件损坏 → 尝试下载
	if s.cfg.DownloadURL == "" {
		s.loaded = true
		err := fmt.Errorf("geoip 数据文件缺失(%s)且未配置下载地址，地区解析已降级", full)
		s.log.Warn(err.Error())
		return err
	}

	if err := os.MkdirAll(s.cfg.DataDir, 0o755); err != nil {
		s.loaded = true
		return fmt.Errorf("创建 geoip 数据目录失败: %w", err)
	}

	tmp := full + ".tmp"
	if err := s.download(ctx, tmp); err != nil {
		s.loaded = true
		s.log.Warn("geoip 数据文件下载失败，地区解析已降级", zap.Error(err))
		return err
	}
	if err := os.Rename(tmp, full); err != nil {
		s.loaded = true
		_ = os.Remove(tmp)
		return fmt.Errorf("geoip 数据文件保存失败: %w", err)
	}

	return s.load(full)
}

// validDataFile 判断文件是否存在且大小超过最小阈值。
func validDataFile(path string) (bool, int64) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, 0
	}
	return fi.Size() >= minDataFileSize, fi.Size()
}

// load 加载已完成校验的 xdb 文件到内存 searcher。
func (s *Searcher) load(full string) error {
	// 使用 BufferCache 全量载入内存：解析并发安全、延迟低，
	// 11MB 文件载入内存对服务端可接受，换取每次查询 O(1) 且无文件句柄竞争。
	v4Config, err := service.NewV4Config(service.BufferCache, full, 1)
	if err != nil {
		return fmt.Errorf("geoip 数据文件加载失败(%s): %w", full, err)
	}

	r, err := service.NewIp2Region(v4Config, nil)
	if err != nil {
		return fmt.Errorf("geoip 解析器初始化失败: %w", err)
	}

	s.region = r
	s.loaded = true
	s.log.Info("geoip 地区解析器初始化完成", zap.String("xdb", full))
	return nil
}

// download 将数据文件下载到目标临时文件。
func (s *Searcher) download(ctx context.Context, dst string) error {
	client := proxyhttp.NewResilient(proxyhttp.Config{
		ProxyURL: s.cfg.ProxyURL,
		Timeout:  3 * time.Minute,
		OnFallback: func(err error) {
			if s.log != nil {
				s.log.Warn("geoip 数据下载经代理失败，已降级直连", zap.Error(err))
			}
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.DownloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载 geoip 数据文件状态异常: %s %s", resp.Status, s.cfg.DownloadURL)
	}
	if resp.ContentLength > 0 && resp.ContentLength < minDataFileSize {
		return fmt.Errorf("下载 geoip 数据文件内容异常(过小): %d bytes", resp.ContentLength)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if err != nil {
		_ = os.Remove(dst)
		return err
	}
	if closeErr != nil {
		_ = os.Remove(dst)
		return closeErr
	}
	if n < minDataFileSize {
		_ = os.Remove(dst)
		return fmt.Errorf("下载 geoip 数据文件内容异常(过小): %d bytes", n)
	}
	s.log.Info("geoip 数据文件下载完成", zap.Int64("bytes", n))
	return nil
}

// Lookup 解析 IP 所属地区，返回形如 "中国 广东省 深圳市" 的字符串；
// 解析失败或数据不可用时返回空字符串。
func (s *Searcher) Lookup(ip string) string {
	if ip == "" {
		return ""
	}

	s.mu.RLock()
	r := s.region
	s.mu.RUnlock()
	if r == nil {
		return ""
	}

	raw, err := r.Search(ip)
	if err != nil || raw == "" {
		return ""
	}
	return formatRegion(raw)
}

// formatRegion 将 ip2region 原始串（"国家|区域|省|市|运营商"）格式化为可读地区。
// 跳过 "0" 与空白段；省/市与上一段重复时去重。
// 保留地址（内网/本机/保留）与国外组织名（如 "United States Google LLC"）
// 按有含义段原样返回。
func formatRegion(raw string) string {
	parts := strings.Split(raw, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// 保留地址/内网标记：返回有含义的一段
	reservedHints := []string{"内网", "局域网", "本机", "reserved", "本地", "未分配"}
	for _, p := range parts {
		l := strings.ToLower(p)
		for _, hint := range reservedHints {
			if strings.Contains(l, hint) {
				return p
			}
		}
	}

	// ip2region v4 布局：国家|区域(常为0)|省|市|运营商
	country := pick(parts, 0)
	province := pick(parts, 2)
	city := pick(parts, 3)

	var segs []string
	for _, seg := range []string{country, province, city} {
		if seg != "" && (len(segs) == 0 || seg != segs[len(segs)-1]) {
			segs = append(segs, seg)
		}
	}

	// 国外地址常表现为 "国家|0|0|0|组织名"（如 "美国 Google LLC"）：取国家 + 组织
	if len(segs) == 1 {
		if org := pick(parts, 4); org != "" && !sameOrContained(segs[0], org) {
			segs = append(segs, org)
		}
	}

	// 地区全空时回退到运营商段，如纯 "0|0|0|0|电信"
	if len(segs) == 0 {
		if isp := pick(parts, 4); isp != "" {
			return isp
		}
	}
	return strings.Join(segs, " ")
}

// sameOrContained 判断两段是否重复或互相包含（如 "北京市" 与 "北京市"、"United States" 与 "United States Google"）。
func sameOrContained(a, b string) bool {
	a, b = strings.ToLower(a), strings.ToLower(b)
	return a == b || strings.Contains(a, b) || strings.Contains(b, a)
}

// pick 取第 idx 段，过滤空与 "0"。
func pick(parts []string, idx int) string {
	if idx < 0 || idx >= len(parts) {
		return ""
	}
	v := parts[idx]
	if v == "" || v == "0" {
		return ""
	}
	return v
}
