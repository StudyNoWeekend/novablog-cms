// Package mapembed 内嵌地图展示页所需静态资产（Leaflet 库与页面模板）。
// 主题通过 iframe 嵌入 /map-embed/travel/:id，本包资产随二进制分发，不依赖外部 CDN。
package mapembed

import "embed"

//go:embed leaflet.js leaflet.css travel.html
var Files embed.FS
