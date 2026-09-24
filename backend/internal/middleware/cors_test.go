package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// newCORSRouter 构建仅挂载 CORS 中间件的测试路由。
func newCORSRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	return r
}

// doPreflight 发送带 Origin 的 OPTIONS 预检请求。
func doPreflight(r *gin.Engine, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestCORSMiddlewareAllowedOrigin 白名单内的来源：预检返回 204 并回显 ACAO。
func TestCORSMiddlewareAllowedOrigin(t *testing.T) {
	SetAllowedOrigins("https://db.example.com,https://blog.example.com")
	r := newCORSRouter()

	w := doPreflight(r, "https://blog.example.com")
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://blog.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

// TestCORSMiddlewareDisallowedOrigin 白名单外的来源：预检返回 403 且不带 ACAO。
func TestCORSMiddlewareDisallowedOrigin(t *testing.T) {
	SetAllowedOrigins("https://db.example.com")
	r := newCORSRouter()

	w := doPreflight(r, "https://evil.example.com")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORSMiddlewareEnvVarNoLongerOverrides 环境变量 CORS_ALLOWED_ORIGINS 不再参与：
// 即使设置了环境变量，也只以 SetAllowedOrigins 写入的白名单（DB 权威）为准。
func TestCORSMiddlewareEnvVarNoLongerOverrides(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://env-override.example.com")
	SetAllowedOrigins("https://db.example.com")
	r := newCORSRouter()

	// 环境变量中的来源不在 DB 白名单内 → 拒绝
	w := doPreflight(r, "https://env-override.example.com")
	assert.Equal(t, http.StatusForbidden, w.Code, "环境变量不应再覆盖 DB 配置")
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))

	// DB 白名单内的来源照常放行
	w = doPreflight(r, "https://db.example.com")
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://db.example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORSMiddlewareDefaultFallback 包级变量为空时回退到硬编码默认值。
func TestCORSMiddlewareDefaultFallback(t *testing.T) {
	SetAllowedOrigins("")
	r := newCORSRouter()

	w := doPreflight(r, "http://localhost:5173")
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORSMiddlewareActualRequest 实际请求（非预检）同样携带 ACAO 头，无 Origin 不带该头。
func TestCORSMiddlewareActualRequest(t *testing.T) {
	SetAllowedOrigins("https://db.example.com")
	r := newCORSRouter()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://db.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://db.example.com", w.Header().Get("Access-Control-Allow-Origin"))

	// 无 Origin 头（同源请求）不放行 ACAO，但请求正常处理
	req = httptest.NewRequest(http.MethodGet, "/ping", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORSMiddlewareOriginTrimming 白名单条目两侧空白应被忽略（前端以逗号拼接保存）。
func TestCORSMiddlewareOriginTrimming(t *testing.T) {
	SetAllowedOrigins(" https://db.example.com , https://blog.example.com ")
	r := newCORSRouter()

	w := doPreflight(r, "https://blog.example.com")
	assert.Equal(t, http.StatusNoContent, w.Code)
}
