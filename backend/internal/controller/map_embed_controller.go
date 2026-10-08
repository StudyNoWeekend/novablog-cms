package controller

import (
	"net/http"

	"novablog/assets/mapembed"

	"github.com/gin-gonic/gin"
)

// MapEmbedController 地图嵌入页控制器。
// 为博客主题提供免 key 的景点地图展示页（Leaflet + OSM），主题通过 iframe 嵌入；
// 访客点击景点后按坐标系自动跳转对应地图平台。
type MapEmbedController struct{}

// NewMapEmbedController 创建 MapEmbedController 实例。
func NewMapEmbedController() *MapEmbedController {
	return &MapEmbedController{}
}

// Travel 渲染攻略景点地图嵌入页 GET /map-embed/travel/:id
// 页面本身为静态模板，数据由页面脚本从同源公开接口 /api/v1/public/travels/:id 拉取。
func (c *MapEmbedController) Travel(ctx *gin.Context) {
	html, err := mapembed.Files.ReadFile("travel.html")
	if err != nil {
		ctx.String(http.StatusInternalServerError, "embed page unavailable")
		return
	}
	ctx.Header("Cache-Control", "no-cache")
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", html)
}

// Asset 提供嵌入页静态资产 GET /map-embed/assets/:file
func (c *MapEmbedController) Asset(ctx *gin.Context) {
	var contentType string
	switch ctx.Param("file") {
	case "leaflet.js":
		contentType = "application/javascript; charset=utf-8"
	case "leaflet.css":
		contentType = "text/css; charset=utf-8"
	default:
		ctx.String(http.StatusNotFound, "not found")
		return
	}
	data, err := mapembed.Files.ReadFile(ctx.Param("file"))
	if err != nil {
		ctx.String(http.StatusNotFound, "not found")
		return
	}
	ctx.Header("Cache-Control", "public, max-age=86400")
	ctx.Data(http.StatusOK, contentType, data)
}
