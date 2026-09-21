// Package proxyhttp 提供"代理优先、探测缓存、失败自动直连兜底"的 HTTP 客户端。
//
// 典型场景：服务器访问 GitHub 不稳定，可为下载配置代理；代理挂掉时自动降级直连，
// 不让代理成为强依赖。三种状态：
//  1. 未配置代理 → 纯直连（与原生 http.Client 行为一致）
//  2. 配置代理且可达 → 请求经代理转发
//  3. 配置代理但不可用（探测失败或请求传输层报错）→ 自动直连兜底
package proxyhttp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// 默认探测地址与超时：GitHub 轻量端点，代理通不通以"能否经代理访问 GitHub"为准。
const (
	DefaultProbeURL     = "https://api.github.com/zen"
	DefaultProbeTimeout = 5 * time.Second
)

// 探测结果缓存时长：成功长缓存（代理一般稳定），失败短缓存（尽快重试恢复）。
const (
	probeSuccessTTL = 5 * time.Minute
	probeFailureTTL = 60 * time.Second
)

// Config 客户端配置。
type Config struct {
	// ProxyURL 代理地址（形如 http://host:port）；空 = 纯直连。
	ProxyURL string
	// Timeout 整体请求超时（同 http.Client.Timeout）。
	Timeout time.Duration
	// ProbeURL 代理可用性探测地址；空则用 DefaultProbeURL。
	ProbeURL string
	// ProbeTimeout 单次探测超时；0 则用 DefaultProbeTimeout。
	ProbeTimeout time.Duration
	// OnFallback 降级直连时的回调（记日志用）；可为 nil。
	OnFallback func(err error)
}

// probeResult 一次探测的结论。
type probeResult struct {
	ok bool
	at time.Time
}

var (
	probeMu     sync.Mutex
	probeCache  = map[string]probeResult{}
	probeClient *http.Client // 探测自身用的客户端（走代理，超时独立）
)

// NewResilient 按配置构建代理感知客户端。
func NewResilient(cfg Config) *http.Client {
	direct := &http.Client{Timeout: cfg.Timeout}
	if cfg.ProxyURL == "" {
		return direct
	}

	proxyURL, err := url.Parse(cfg.ProxyURL)
	if err != nil || proxyURL.Host == "" {
		// 代理地址非法：按未配置处理（兜底直连），不让错误配置阻断下载
		if cfg.OnFallback != nil {
			cfg.OnFallback(fmt.Errorf("代理地址非法 %q，已按直连处理: %v", cfg.ProxyURL, err))
		}
		return direct
	}

	proxyTransport := http.DefaultTransport.(*http.Transport).Clone()
	proxyTransport.Proxy = http.ProxyURL(proxyURL)
	directTransport := http.DefaultTransport.(*http.Transport).Clone()

	return &http.Client{
		Timeout: cfg.Timeout,
		Transport: &resilientTransport{
			proxyURL:       cfg.ProxyURL,
			proxyTransport: proxyTransport,
			direct:         directTransport,
			cfg:            cfg,
		},
	}
}

// resilientTransport 代理优先的 RoundTripper：探测可用走代理，失败直连兜底。
type resilientTransport struct {
	proxyURL       string
	proxyTransport http.RoundTripper
	direct         http.RoundTripper
	cfg            Config
}

// RoundTrip 执行一次请求：代理可用 → 经代理；经代理出现传输层错误 → 直连重试一次。
func (t *resilientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !proxyAvailable(t.cfg) {
		return t.direct.RoundTrip(req)
	}

	resp, err := t.proxyTransport.RoundTrip(req)
	if err == nil {
		return resp, nil
	}

	// 传输层错误（连接拒绝/超时/中断）才降级；HTTP 状态码错误由调用方按响应处理
	// Body 已消费的请求无法重放（如流式上传），放弃重试
	if !canRetry(req) {
		return nil, err
	}

	if t.cfg.OnFallback != nil {
		t.cfg.OnFallback(err)
	}
	retry := req.Clone(req.Context())
	if req.Body != nil && req.GetBody != nil {
		body, bodyErr := req.GetBody()
		if bodyErr != nil {
			return nil, err
		}
		retry.Body = body
	}
	return t.direct.RoundTrip(retry)
}

// canRetry 判断请求是否可安全重放：无 Body，或 Body 支持 GetBody 重放。
func canRetry(req *http.Request) bool {
	return req.Body == nil || req.GetBody != nil
}

// proxyAvailable 判断代理当前是否可用（探测结果带 TTL 缓存）。
func proxyAvailable(cfg Config) bool {
	probeURL := cfg.ProbeURL
	if probeURL == "" {
		probeURL = DefaultProbeURL
	}

	probeMu.Lock()
	defer probeMu.Unlock()

	if cached, ok := probeCache[cfg.ProxyURL]; ok {
		ttl := probeFailureTTL
		if cached.ok {
			ttl = probeSuccessTTL
		}
		if time.Since(cached.at) < ttl {
			return cached.ok
		}
	}

	ok := probeOnce(cfg, probeURL)
	probeCache[cfg.ProxyURL] = probeResult{ok: ok, at: time.Now()}
	return ok
}

// probeOnce 立即执行一次探测：经代理 GET probeURL，2xx 视为可用。
func probeOnce(cfg Config, probeURL string) bool {
	timeout := cfg.ProbeTimeout
	if timeout <= 0 {
		timeout = DefaultProbeTimeout
	}
	if probeClient == nil || probeClient.Timeout != timeout {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err != nil || proxyURL.Host == "" {
			return false
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.Proxy = http.ProxyURL(proxyURL)
		probeClient = &http.Client{Timeout: timeout, Transport: transport}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	if err != nil {
		return false
	}
	resp, err := probeClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// ResetProbeCache 清空探测缓存（测试用）。
func ResetProbeCache() {
	probeMu.Lock()
	defer probeMu.Unlock()
	probeCache = map[string]probeResult{}
	probeClient = nil
}
