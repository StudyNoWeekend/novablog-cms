// Package res 定义响应 DTO（数据传输对象）。
package res

import "time"

// ThemeItemRes 已安装主题条目。
type ThemeItemRes struct {
	ID           string            `json:"id"`            // 安装实例 UUID
	ThemeID      string            `json:"theme_id"`      // 主题标识（theme.json id）
	Name         string            `json:"name"`          // 展示名
	Version      string            `json:"version"`       // 版本
	Engine       string            `json:"engine"`        // 渲染引擎
	APICompat    string            `json:"api_compat"`    // 兼容 API 版本
	Author       string            `json:"author"`        // 作者
	Description  string            `json:"description"`   // 简介
	Source       string            `json:"source"`        // 来源：official | builtin
	Screenshots  []string          `json:"screenshots"`   // 截图相对路径
	Fallbacks    map[string]string `json:"fallbacks"`     // 壳页面映射
	MarketID     int64             `json:"market_id"`     // 官方市场主题 ID
	MarketSlug   string            `json:"market_slug"`   // 官方市场主题 slug
	ArtifactPath string            `json:"artifact_path"` // 制品目录名（相对 data_dir）
	Checksum     string            `json:"checksum"`      // 制品 sha256
	Active       bool              `json:"active"`        // 是否使用中
	CreatedAt    time.Time         `json:"created_at"`    // 安装时间
}
