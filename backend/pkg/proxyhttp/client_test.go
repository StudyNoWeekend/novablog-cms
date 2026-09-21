package proxyhttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// newForwardProxy 构造一个 HTTP 正向代理测试服务器：
// 收到绝对形式 URI 的请求（经代理）时向该 URI 转发并回传响应，模拟真实代理行为。
func newForwardProxy(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 经代理的请求 URI 为绝对地址；直连本代理的请求为路径形式，拒绝
		if r.URL.Host == "" {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		resp, err := http.Get(r.URL.String())
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}))
}

// TestEmptyProxyDirect 代理为空 → 纯直连可用。
func TestEmptyProxyDirect(t *testing.T) {
	ResetProbeCache()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "direct-ok")
	}))
	defer server.Close()

	client := NewResilient(Config{Timeout: 5 * time.Second})
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("直连请求应成功: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if string(data) != "direct-ok" {
		t.Fatalf("直连响应不正确: %q", string(data))
	}
}

// TestProxyAvailable 代理可达 → 请求经代理转发。
func TestProxyAvailable(t *testing.T) {
	ResetProbeCache()
	zen := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "zen-ok")
	}))
	defer zen.Close()

	proxy := newForwardProxy(t)
	defer proxy.Close()

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "via-proxy")
	}))
	defer target.Close()

	cfg := Config{
		Timeout:      5 * time.Second,
		ProxyURL:     proxy.URL,
		ProbeURL:     zen.URL,
		ProbeTimeout: 3 * time.Second,
		OnFallback: func(err error) {
			t.Errorf("代理可用时不应触发降级: %v", err)
		},
	}

	// 探测：经代理访问 zen 端点
	if !proxyAvailable(cfg) {
		t.Fatal("代理探测应通过")
	}

	client := NewResilient(cfg)
	resp, err := client.Get(target.URL)
	if err != nil {
		t.Fatalf("经代理请求应成功: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if string(data) != "via-proxy" {
		t.Fatalf("经代理响应不正确: %q", string(data))
	}
}

// TestProxyUnreachableFallback 代理不可达 → 自动降级直连兜底成功。
func TestProxyUnreachableFallback(t *testing.T) {
	ResetProbeCache()
	fallbacks := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "direct-fallback")
	}))
	defer target.Close()

	// 指向一个确定不可达的地址（httptest 关闭后的端口）
	dead := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	deadURL, _ := url.Parse(dead.URL)
	dead.Close()

	cfg := Config{
		Timeout:      5 * time.Second,
		ProxyURL:     deadURL.Host,
		ProbeURL:     dead.URL, // 探测同样不可达
		ProbeTimeout: 2 * time.Second,
		OnFallback: func(err error) {
			fallbacks++
		},
	}

	client := NewResilient(cfg)
	resp, err := client.Get(target.URL)
	if err != nil {
		t.Fatalf("代理不可达时应降级直连成功: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if string(data) != "direct-fallback" {
		t.Fatalf("降级直连响应不正确: %q", string(data))
	}
	if fallbacks == 0 {
		t.Fatal("降级时应触发 OnFallback 回调")
	}
}

// TestProbeCache 探测结果缓存：TTL 内不重复探测。
func TestProbeCache(t *testing.T) {
	ResetProbeCache()
	probeHits := 0
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probeHits++
		_, _ = io.WriteString(w, "probe-ok")
	}))
	defer proxy.Close()

	cfg := Config{
		Timeout:      5 * time.Second,
		ProxyURL:     proxy.URL,
		ProbeURL:     proxy.URL,
		ProbeTimeout: 3 * time.Second,
	}

	for i := 0; i < 3; i++ {
		if !proxyAvailable(cfg) {
			t.Fatalf("第 %d 次探测应通过", i+1)
		}
	}
	if probeHits != 1 {
		t.Fatalf("探测应命中缓存只发一次请求, 实际 %d 次", probeHits)
	}

	ResetProbeCache()
	if !proxyAvailable(cfg) {
		t.Fatal("清空缓存后应重新探测")
	}
	if probeHits != 2 {
		t.Fatalf("清空缓存后应重新探测一次, 实际 %d 次", probeHits)
	}
}

// TestProxyDiesMidway 探测通过后代理中途挂掉 → 请求级降级直连重试成功。
func TestProxyDiesMidway(t *testing.T) {
	ResetProbeCache()
	zen := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "zen-ok")
	}))
	defer zen.Close()

	proxy := newForwardProxy(t)
	proxyURL := proxy.URL

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "rescued")
	}))
	defer target.Close()

	fallbacks := 0
	cfg := Config{
		Timeout:      5 * time.Second,
		ProxyURL:     proxyURL,
		ProbeURL:     zen.URL,
		ProbeTimeout: 3 * time.Second,
		OnFallback: func(err error) {
			fallbacks++
		},
	}

	client := NewResilient(cfg)
	if !proxyAvailable(cfg) {
		t.Fatal("代理探测应通过")
	}

	// 探测通过后关闭代理，模拟"代理中途挂掉"
	proxy.Close()

	resp, err := client.Get(target.URL)
	if err != nil {
		t.Fatalf("代理中途挂掉应降级直连成功: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if string(data) != "rescued" {
		t.Fatalf("降级直连响应不正确: %q", string(data))
	}
	if fallbacks == 0 {
		t.Fatal("请求级降级应触发 OnFallback 回调")
	}
}

// TestInvalidProxyURL 代理地址非法 → 按直连处理。
func TestInvalidProxyURL(t *testing.T) {
	ResetProbeCache()
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "direct-ok")
	}))
	defer target.Close()

	client := NewResilient(Config{Timeout: 5 * time.Second, ProxyURL: "::::not-a-url"})
	resp, err := client.Get(target.URL)
	if err != nil {
		t.Fatalf("非法代理应兜底直连: %v", err)
	}
	resp.Body.Close()
}
