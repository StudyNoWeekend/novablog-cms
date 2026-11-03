package logic

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"novablog/enum"
	"novablog/internal/model"

	"go.uber.org/zap"
)

// mountedFrontendKey 自备博客前端在托管缓存中的保留键（与安装实例 ID 不冲突）。
const mountedFrontendKey = "@mounted-frontend"

// 响应决策类型（由 logic 计算由 controller 执行）。
const (
	// ThemeHostKindFile 静态文件，直接 http.ServeFile。
	ThemeHostKindFile = "file"
	// ThemeHostKindHTML HTML 字节（fallback/占位/404 页）。
	ThemeHostKindHTML = "html"
	// ThemeHostKindText 纯文本或裸状态码。
	ThemeHostKindText = "text"
)

// ThemeHostResponse 主题托管请求的响应决策。
// logic 层只计算「响应什么」，HTTP 执行（ServeFile/响应头/状态码）由 controller 完成。
type ThemeHostResponse struct {
	Kind      string // ThemeHostKind* 之一
	FilePath  string // Kind=file：直接响应的文件绝对路径
	Immutable bool   // Kind=file：是否长缓存（Next.js 哈希资源）
	Status    int    // Kind=html/text：HTTP 状态码
	HTML      []byte // Kind=html：页面字节
	Text      string // Kind=text：文本内容（空则仅返回状态码）
}

// hostedTheme 已解析的托管主题（制品目录 + 清单）。
type hostedTheme struct {
	root     string
	manifest *ThemeManifest
}

// ThemeHost 博客主题静态托管器。
// 全局单例：路由兜底处理器与激活/卸载时的缓存失效共享同一实例。
var ThemeHost = &ThemeHostLogic{cache: map[string]*hostedTheme{}}

// ThemeHostLogic 主题静态托管逻辑（纯解析与决策，不依赖 HTTP 框架）。
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

// ResolveHosted 博客托管兜底入口：解析自备前端或激活主题，返回响应决策。
func (h *ThemeHostLogic) ResolveHosted(ctx context.Context, themeLogic *ThemeLogic, reqPath string) (ThemeHostResponse, error) {
	// 自备博客前端优先（config: themes.frontend_dir，挂载目录与主题制品包同构）
	if dir := getThemeSettings().FrontendDir; dir != "" {
		if ht, err := h.resolveMounted(dir); err == nil && ht != nil {
			return h.ResolveThemePath(ht, reqPath), nil
		}
		// 不可用时静默回退到已安装主题流程（首次失败已记录日志）
	}

	theme, err := themeLogic.GetActiveTheme(ctx)
	if err != nil {
		return ThemeHostResponse{}, err
	}
	if theme == nil {
		return ThemeHostResponse{Kind: ThemeHostKindHTML, Status: http.StatusOK, HTML: []byte(noThemeHTML())}, nil
	}

	ht, err := h.resolve(ctx, theme.ID)
	if err != nil {
		return ThemeHostResponse{}, err
	}
	return h.ResolveThemePath(ht, reqPath), nil
}

// ResolvePreview 主题预览入口：以指定主题（最新安装实例）解析同一套托管逻辑。
func (h *ThemeHostLogic) ResolvePreview(ctx context.Context, themeID, subPath string) (ThemeHostResponse, error) {
	if subPath == "" {
		subPath = "/"
	}

	theme, err := model.NewTheme().GetLatestByThemeID(ctx, themeID)
	if err != nil {
		return ThemeHostResponse{}, enum.ErrThemeNotFound
	}
	ht, err := h.resolve(ctx, theme.ID)
	if err != nil {
		return ThemeHostResponse{}, err
	}
	return h.ResolveThemePath(ht, subPath), nil
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

// resolveMounted 解析自备博客前端（带缓存；失败结果同样缓存，避免每次请求重复解析与刷日志）。
func (h *ThemeHostLogic) resolveMounted(dir string) (*hostedTheme, error) {
	h.mu.RLock()
	ht, ok := h.cache[mountedFrontendKey]
	h.mu.RUnlock()
	if ok {
		return ht, nil
	}

	var err error
	ht, err = loadMountedFrontend(dir)
	if err != nil {
		themeLog().Warn("自备博客前端不可用，回退到已安装主题", zap.String("dir", dir), zap.Error(err))
		ht = nil
	} else {
		themeLog().Info("已启用自备博客前端",
			zap.String("dir", dir),
			zap.String("name", ht.manifest.Name),
			zap.String("version", ht.manifest.Version),
		)
	}

	h.mu.Lock()
	h.cache[mountedFrontendKey] = ht
	h.mu.Unlock()
	return ht, err
}

// loadMountedFrontend 解析自备博客前端目录：制品根含 theme.json，网页根为其 dist/。
// 清单缺失或非法时按空清单托管（无 fallback 壳页面），保证纯静态站点仍可访问。
func loadMountedFrontend(dir string) (*hostedTheme, error) {
	root := filepath.Join(dir, "dist")
	if !isRegularFile(filepath.Join(root, "index.html")) {
		return nil, fmt.Errorf("挂载目录缺少 dist/index.html")
	}

	manifest, err := NewThemeArtifactLogic().LoadManifest(dir)
	if err != nil {
		themeLog().Warn("自备博客前端清单不可用，按无壳页面模式托管", zap.String("dir", dir), zap.Error(err))
		manifest = &ThemeManifest{}
	}
	return &hostedTheme{root: root, manifest: manifest}, nil
}

// ResolveThemePath 按托管解析顺序把请求路径解析为响应决策：
// 精确文件 → {path}.html → 目录 index.html → fallback 壳页面 → 404.html → 404。
func (h *ThemeHostLogic) ResolveThemePath(ht *hostedTheme, reqPath string) ThemeHostResponse {
	clean := path.Clean("/" + reqPath)

	// 拒绝 dotfile 与清单直出
	rel := strings.Trim(clean, "/")
	if rel == "theme.json" || strings.HasPrefix(path.Base(clean), ".") || containsDotSegment(rel) {
		return ThemeHostResponse{Kind: ThemeHostKindText, Status: http.StatusNotFound}
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
			return themeHostFileDecision(full, cand)
		}
	}

	// fallback 壳页面（最长前缀匹配，见主题规范 3.4）
	if shell, ok := matchFallback(ht.manifest, clean); ok {
		shellRel := strings.TrimPrefix(shell, "/")
		if full := filepath.Join(ht.root, filepath.FromSlash(shellRel)); isRegularFile(full) {
			return themeHostFileDecision(full, shellRel)
		}
	}

	// 404.html
	if full := filepath.Join(ht.root, "404.html"); isRegularFile(full) {
		if data, err := os.ReadFile(full); err == nil {
			return ThemeHostResponse{Kind: ThemeHostKindHTML, Status: http.StatusNotFound, HTML: data}
		}
	}
	return ThemeHostResponse{Kind: ThemeHostKindText, Status: http.StatusNotFound, Text: "404 page not found"}
}

// themeHostFileDecision 构造静态文件响应决策（Next.js 哈希资源长缓存）。
func themeHostFileDecision(fullPath, rel string) ThemeHostResponse {
	immutable := strings.HasPrefix(rel, "_next/") || strings.HasPrefix(rel, "static/")
	return ThemeHostResponse{Kind: ThemeHostKindFile, FilePath: fullPath, Immutable: immutable}
}

// noThemeHTML 未激活主题时的占位页（首装未完成或未切换主题）。
func noThemeHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>NovaBlog</title>
<style>body{font-family:system-ui,-apple-system,"PingFang SC",sans-serif;background:#0a0a0a;color:#e5e5e5;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0}main{text-align:center;padding:24px}h1{font-size:22px;margin-bottom:12px}p{color:#8a8a8a;margin:6px 0}a{color:#4ade80}</style></head>
<body><main>
<h1>博客即将上线</h1>
<p>尚未启用博客主题，请前往管理后台完成初始化。</p>
</main></body></html>
`
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
