package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// newMapEmbedEngine 构造仅注册地图嵌入路由的测试引擎。
func newMapEmbedEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	c := NewMapEmbedController()
	r.GET("/map-embed/travel/:id", c.Travel)
	r.GET("/map-embed/assets/:file", c.Asset)
	return r
}

func TestMapEmbedTravelPage(t *testing.T) {
	r := newMapEmbedEngine()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/map-embed/travel/550e8400-e29b-41d4-a716-446655440000", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "novablog-map-embed") {
		t.Error("页面缺少 postMessage 高度协议标记 novablog-map-embed")
	}
	if !strings.Contains(body, "/api/v1/public/travels/") {
		t.Error("页面缺少公开 API 取数逻辑")
	}
	if !strings.Contains(body, "wgs84FromGcj02") {
		t.Error("页面缺少 GCJ-02 纠偏逻辑")
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, 期望 text/html", ct)
	}
}

func TestMapEmbedAssets(t *testing.T) {
	r := newMapEmbedEngine()

	cases := []struct {
		file        string
		wantCode    int
		wantContain string
	}{
		{"leaflet.js", http.StatusOK, "Leaflet"},
		{"leaflet.css", http.StatusOK, ".leaflet-pane"},
		{"../secret", http.StatusNotFound, ""},
		{"unknown.txt", http.StatusNotFound, ""},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/map-embed/assets/"+tc.file, nil)
		r.ServeHTTP(w, req)
		if w.Code != tc.wantCode {
			t.Errorf("GET %s 状态码 = %d, 期望 %d", tc.file, w.Code, tc.wantCode)
			continue
		}
		if tc.wantContain != "" && !strings.Contains(w.Body.String(), tc.wantContain) {
			t.Errorf("GET %s 响应未包含 %q", tc.file, tc.wantContain)
		}
	}
}
