package bilibili

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	// mockNavBody 匿名 nav 响应（code=-101 但仍含 wbi_img），密钥为官方测试向量。
	mockNavBody = `{"code":-101,"message":"账号未登录","data":{"wbi_img":{"img_url":"https://i0.hdslb.com/bfs/wbi/7cd084941338484aae1ad9425b242ee3.png","sub_url":"https://i0.hdslb.com/bfs/wbi/4932caff0ff746eab6f01bf08b70ac45.jpg"}}}`
	// mockSPIBody spi 访客身份响应。
	mockSPIBody = `{"code":0,"data":{"b_3":"mock-buvid3","b_4":"mock-buvid4"}}`
)

// overrideAPIBase 将接口基地址指向 mock 服务并在测试结束后还原。
func overrideAPIBase(t *testing.T, serverURL string) {
	t.Helper()
	old := apiBaseURL
	apiBaseURL = serverURL
	t.Cleanup(func() { apiBaseURL = old })
}

// resetBiliCache 清空 wbi 密钥与 buvid 进程内缓存，避免测试间串扰。
func resetBiliCache(t *testing.T) {
	t.Helper()
	wbiKeyMu.Lock()
	wbiKeyCache, wbiKeyExpireAt = "", time.Time{}
	wbiKeyMu.Unlock()
	buvidMu.Lock()
	buvidCookieCache, buvidExpireAt = "", time.Time{}
	buvidMu.Unlock()
}

// TestFetchVideoInfoWBIHappyPath 验证 wbi/view 主路径的完整映射与签名参数。
func TestFetchVideoInfoWBIHappyPath(t *testing.T) {
	resetBiliCache(t)
	var viewQuery url.Values
	var viewCookie string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/x/web-interface/nav":
			_, _ = w.Write([]byte(mockNavBody))
		case "/x/frontend/finger/spi":
			_, _ = w.Write([]byte(mockSPIBody))
		case "/x/web-interface/wbi/view":
			viewQuery = r.URL.Query()
			viewCookie = r.Header.Get("Cookie")
			_, _ = w.Write([]byte(`{"code":0,"data":{"title":"不该","pic":"http://i0.hdslb.com/bfs/archive/cover.jpg","desc":"d","cid":111,"duration":283,"owner":{"name":"UP"},"pages":[{"cid":111,"part":"P1","duration":283}]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	overrideAPIBase(t, ts.URL)

	info, err := FetchVideoInfo(context.Background(), "BV1DTJR6REcK")
	if err != nil {
		t.Fatalf("FetchVideoInfo() error = %v", err)
	}
	if info.Title != "不该" || info.OwnerName != "UP" || info.CID != 111 || info.Duration != 283 || len(info.Pages) != 1 {
		t.Fatalf("映射结果不符合预期: %+v", info)
	}
	if info.Pic != "https://i0.hdslb.com/bfs/archive/cover.jpg" {
		t.Fatalf("封面 http 未升级为 https: %q", info.Pic)
	}
	if viewQuery.Get("bvid") != "BV1DTJR6REcK" || viewQuery.Get("wts") == "" || viewQuery.Get("w_rid") == "" {
		t.Fatalf("wbi 签名参数缺失: %v", viewQuery)
	}
	if !strings.Contains(viewCookie, "buvid3=mock-buvid3") {
		t.Fatalf("wbi/view 请求未携带 buvid Cookie: %q", viewCookie)
	}
}

// TestFetchVideoInfoFallsBackToPagelist 验证 412 风控时降级 pagelist。
func TestFetchVideoInfoFallsBackToPagelist(t *testing.T) {
	resetBiliCache(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/x/web-interface/nav":
			_, _ = w.Write([]byte(mockNavBody))
		case "/x/frontend/finger/spi":
			_, _ = w.Write([]byte(mockSPIBody))
		case "/x/web-interface/wbi/view":
			w.WriteHeader(http.StatusPreconditionFailed)
			_, _ = w.Write([]byte("<html>出错啦!</html>"))
		case "/x/player/pagelist":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"cid":42877715092,"page":1,"part":"周杰伦&aMEI 《不该》","duration":283}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	overrideAPIBase(t, ts.URL)

	info, err := FetchVideoInfo(context.Background(), "BV1DTJR6REcK")
	if err != nil {
		t.Fatalf("降级解析失败: %v", err)
	}
	if info.Title != "周杰伦&aMEI 《不该》" || info.CID != 42877715092 || info.Duration != 283 || len(info.Pages) != 1 {
		t.Fatalf("降级结果不符合预期: %+v", info)
	}
	if info.Pic != "" || info.Desc != "" || info.OwnerName != "" {
		t.Fatalf("降级结果不应包含封面/简介/UP主: %+v", info)
	}
}

// TestFetchVideoInfoBusinessErrorNoFallback 验证业务错误（如视频不存在）不触发降级。
func TestFetchVideoInfoBusinessErrorNoFallback(t *testing.T) {
	resetBiliCache(t)
	pagelistCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/x/web-interface/nav":
			_, _ = w.Write([]byte(mockNavBody))
		case "/x/frontend/finger/spi":
			_, _ = w.Write([]byte(mockSPIBody))
		case "/x/web-interface/wbi/view":
			_, _ = w.Write([]byte(`{"code":-404,"message":"啥都木有"}`))
		case "/x/player/pagelist":
			pagelistCalled = true
			_, _ = w.Write([]byte(`{"code":0,"data":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	overrideAPIBase(t, ts.URL)

	_, err := FetchVideoInfo(context.Background(), "BV1DTJR6REcK")
	if err == nil || !strings.Contains(err.Error(), "-404") {
		t.Fatalf("应返回 code -404 错误, got: %v", err)
	}
	if pagelistCalled {
		t.Fatalf("业务错误不应触发 pagelist 降级")
	}
}

// TestFetchVideoInfoBothFailRiskMessage 验证两级都被风控时返回明确错误。
func TestFetchVideoInfoBothFailRiskMessage(t *testing.T) {
	resetBiliCache(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/x/web-interface/nav":
			_, _ = w.Write([]byte(mockNavBody))
		case "/x/frontend/finger/spi":
			_, _ = w.Write([]byte(mockSPIBody))
		case "/x/web-interface/wbi/view", "/x/player/pagelist":
			w.WriteHeader(http.StatusPreconditionFailed)
			_, _ = w.Write([]byte("<html>出错啦!</html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	overrideAPIBase(t, ts.URL)

	_, err := FetchVideoInfo(context.Background(), "BV1DTJR6REcK")
	if err == nil {
		t.Fatalf("两级都失败时应返回错误")
	}
	if !strings.Contains(err.Error(), "风控拦截") || !strings.Contains(err.Error(), "pagelist") {
		t.Fatalf("错误信息应包含风控原因与降级原因, got: %v", err)
	}
}

// TestFetchVideoInfoWBIRefreshKeyOnInvalid 验证 -403 时刷新密钥并重试。
func TestFetchVideoInfoWBIRefreshKeyOnInvalid(t *testing.T) {
	resetBiliCache(t)
	navCalls, viewCalls := 0, 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/x/web-interface/nav":
			navCalls++
			_, _ = w.Write([]byte(mockNavBody))
		case "/x/frontend/finger/spi":
			_, _ = w.Write([]byte(mockSPIBody))
		case "/x/web-interface/wbi/view":
			viewCalls++
			if viewCalls == 1 {
				_, _ = w.Write([]byte(`{"code":-403,"message":"访问权限不足"}`))
				return
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"title":"不该","cid":111,"duration":283,"owner":{"name":"UP"},"pages":[{"cid":111,"part":"P1","duration":283}]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	overrideAPIBase(t, ts.URL)

	info, err := FetchVideoInfo(context.Background(), "BV1DTJR6REcK")
	if err != nil {
		t.Fatalf("刷新密钥重试后应成功: %v", err)
	}
	if navCalls != 2 {
		t.Fatalf("应刷新一次签名密钥, nav 调用次数 = %d", navCalls)
	}
	if viewCalls != 2 || info.Title != "不该" {
		t.Fatalf("重试结果不符合预期: calls=%d info=%+v", viewCalls, info)
	}
}
