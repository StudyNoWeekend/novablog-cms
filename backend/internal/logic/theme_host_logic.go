package logic

import (
	"context"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"novablog/enum"
	"novablog/internal/model"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// hostedTheme 已解析的托管主题（制品目录 + 清单）。
type hostedTheme struct {
	root     string
	manifest *ThemeManifest
}

// ThemeHost 博客主题静态托管器。
// 全局单例：路由兜底处理器与激活/卸载时的缓存失效共享同一实例。
var ThemeHost = &ThemeHostLogic{cache: map[string]*hostedTheme{}}

// ThemeHostLogic 主题静态托管逻辑。
type ThemeHostLogic struct {
	mu    sync.RWMutex
	cache map[string]*hostedTheme
}

// Invalidate 清空托管缓存（激活/卸载/覆盖安装后调用，下一次请求重新解析）。
func (h *ThemeHostLogic) Invalidate() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cache = map[string]*hostedTheme{}
}

// Handler 博客托管兜底路由（NoRoute）：非保留路径按激活主题静态解析。
func (h *ThemeHostLogic) Handler(themeLogic *ThemeLogic) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqPath := c.Request.URL.Path

		// 仅 GET/HEAD 参与托管，其余方法按接口 404 处理
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			response.Fail(c, enum.ErrNotFound.Code, "接口不存在", http.StatusNotFound)
			return
		}
		// 保留路径（API/媒体文件/预览/健康检查）兜底 404
		if strings.HasPrefix(reqPath, "/api/") || strings.HasPrefix(reqPath, "/files/") ||
			strings.HasPrefix(reqPath, "/preview/") || reqPath == "/health" || reqPath == "/ready" {
			response.Fail(c, enum.ErrNotFound.Code, "接口不存在", http.StatusNotFound)
			return
		}

		theme, err := themeLogic.GetActiveTheme(c.Request.Context())
		if err != nil {
			response.HandleError(c, err)
			return
		}
		if theme == nil {
			serveNoTheme(c)
			return
		}

		ht, err := h.resolve(c.Request.Context(), theme.ID)
		if err != nil {
			response.HandleError(c, err)
			return
		}
		h.serveTheme(c, ht, reqPath)
	}
}

// PreviewHandler 主题预览路由：以指定主题（最新安装实例）响应同一套托管逻辑。
func (h *ThemeHostLogic) PreviewHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		themeID := c.Param("theme_id")
		subPath := c.Param("path")
		if subPath == "" {
			subPath = "/"
		}

		theme, err := model.NewTheme().GetLatestByThemeID(c.Request.Context(), themeID)
		if err != nil {
			response.Fail(c, enum.ErrThemeNotFound.Code, enum.ErrThemeNotFound.Msg, http.StatusNotFound)
			return
		}
		ht, err := h.resolve(c.Request.Context(), theme.ID)
		if err != nil {
			response.HandleError(c, err)
			return
		}
		h.serveTheme(c, ht, subPath)
	}
}

// resolve 解析指定安装实例的托管主题（制品目录 + 清单，带缓存）。
func (h *ThemeHostLogic) resolve(ctx context.Context, instanceID string) (*hostedTheme, error) {
	h.mu.RLock()
	ht, ok := h.cache[instanceID]
	h.mu.RUnlock()
	if ok {
		return ht, nil
	}

	theme, err := model.NewTheme().GetByID(ctx, instanceID)
	if err != nil {
		return nil, enum.ErrThemeNotFound
	}
	root := filepath.Join(getThemeSettings().DataDir, theme.ArtifactPath, "dist")
	manifest, err := NewThemeArtifactLogic().LoadManifest(filepath.Dir(root))
	if err != nil {
		return nil, err
	}

	ht = &hostedTheme{root: root, manifest: manifest}
	h.mu.Lock()
	h.cache[instanceID] = ht
	h.mu.Unlock()
	return ht, nil
}

// serveTheme 按托管解析顺序响应：
// 精确文件 → {path}.html → 目录 index.html → fallback 壳页面 → 404.html → 404。
func (h *ThemeHostLogic) serveTheme(c *gin.Context, ht *hostedTheme, reqPath string) {
	clean := path.Clean("/" + reqPath)

	// 拒绝 dotfile 与清单直出
	rel := strings.Trim(clean, "/")
	if rel == "theme.json" || strings.HasPrefix(path.Base(clean), ".") || containsDotSegment(rel) {
		c.Status(http.StatusNotFound)
		return
	}
	if rel == "" {
		rel = "index.html"
	}

	for _, cand := range []string{rel, rel + ".html", path.Join(rel, "index.html")} {
		if cand == "" {
			continue
		}
		full := filepath.Join(ht.root, filepath.FromSlash(cand))
		if isRegularFile(full) {
			serveThemeStatic(c, full, cand)
			return
		}
	}

	// fallback 壳页面（最长前缀匹配，见主题规范 3.4）
	if shell, ok := matchFallback(ht.manifest, clean); ok {
		shellRel := strings.TrimPrefix(shell, "/")
		if full := filepath.Join(ht.root, filepath.FromSlash(shellRel)); isRegularFile(full) {
			serveThemeStatic(c, full, shellRel)
			return
		}
	}

	// 404.html
	if full := filepath.Join(ht.root, "404.html"); isRegularFile(full) {
		data, err := os.ReadFile(full)
		if err == nil {
			c.Data(http.StatusNotFound, "text/html; charset=utf-8", data)
			return
		}
	}
	c.String(http.StatusNotFound, "404 page not found")
}

// serveThemeStatic 响应静态文件并设置缓存策略（Next.js 哈希资源长缓存）。
func serveThemeStatic(c *gin.Context, fullPath, rel string) {
	if strings.HasPrefix(rel, "_next/") || strings.HasPrefix(rel, "static/") {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Header("Cache-Control", "no-cache")
	}
	http.ServeFile(c.Writer, c.Request, fullPath)
}

// matchFallback 在清单 fallback 声明中做最长前缀匹配（pattern 形如 /articles/*）。
func matchFallback(m *ThemeManifest, cleanPath string) (string, bool) {
	if m == nil || m.Routes.Fallback == nil {
		return "", false
	}
	best, bestLen := "", 0
	for pattern, shell := range m.Routes.Fallback {
		prefix := strings.TrimSuffix(pattern, "*")
		if strings.HasPrefix(cleanPath, prefix) && len(prefix) > bestLen {
			best, bestLen = shell, len(prefix)
		}
	}
	return best, best != ""
}

// containsDotSegment 判断相对路径中是否包含 dotfile 片段。
func containsDotSegment(rel string) bool {
	if rel == "" {
		return false
	}
	for _, seg := range strings.Split(rel, "/") {
		if strings.HasPrefix(seg, ".") {
			return true
		}
	}
	return false
}

// isRegularFile 判断路径是否为常规文件。
func isRegularFile(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.Mode().IsRegular()
}

// serveNoTheme 未激活主题时的占位页（首装未完成或未切换主题）。
func serveNoTheme(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>NovaBlog</title>
<style>body{font-family:system-ui,-apple-system,"PingFang SC",sans-serif;background:#0a0a0a;color:#e5e5e5;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0}main{text-align:center;padding:24px}h1{font-size:22px;margin-bottom:12px}p{color:#8a8a8a;margin:6px 0}a{color:#4ade80}</style></head>
<body><main>
<h1>博客即将上线</h1>
<p>尚未启用博客主题，请前往管理后台完成初始化。</p>
</main></body></html>
`)
}
