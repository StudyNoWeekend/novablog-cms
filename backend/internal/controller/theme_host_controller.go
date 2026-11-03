package controller

import (
	"net/http"
	"strings"

	"novablog/enum"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// ThemeHostController 博客主题静态托管控制器（NoRoute 兜底与主题预览）。
type ThemeHostController struct{}

// NewThemeHostController 创建 ThemeHostController 实例。
func NewThemeHostController() *ThemeHostController {
	return &ThemeHostController{}
}

// Handler 博客托管兜底路由（NoRoute）：非保留路径按激活主题静态解析。
func (ctl *ThemeHostController) Handler(themeLogic *logic.ThemeLogic) gin.HandlerFunc {
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

		resp, err := logic.ThemeHost.ResolveHosted(c.Request.Context(), themeLogic, reqPath)
		if err != nil {
			response.HandleError(c, err)
			return
		}
		renderThemeHost(c, resp)
	}
}

// Preview 主题预览路由 GET /preview/:theme_id 与 /preview/:theme_id/*path。
func (ctl *ThemeHostController) Preview(c *gin.Context) {
	resp, err := logic.ThemeHost.ResolvePreview(c.Request.Context(), c.Param("theme_id"), c.Param("path"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	renderThemeHost(c, resp)
}

// renderThemeHost 执行托管响应决策（文件/HTML/文本）。
func renderThemeHost(c *gin.Context, resp logic.ThemeHostResponse) {
	switch resp.Kind {
	case logic.ThemeHostKindFile:
		if resp.Immutable {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			c.Header("Cache-Control", "no-cache")
		}
		http.ServeFile(c.Writer, c.Request, resp.FilePath)
	case logic.ThemeHostKindHTML:
		c.Data(resp.Status, "text/html; charset=utf-8", resp.HTML)
	default:
		if resp.Text == "" {
			c.Status(resp.Status)
			return
		}
		c.String(resp.Status, "%s", resp.Text)
	}
}
