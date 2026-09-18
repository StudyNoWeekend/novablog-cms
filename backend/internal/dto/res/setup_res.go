// Package res 定义响应 DTO（数据传输对象）。
package res

import "time"

// StatusRes 初始化状态响应。
type StatusRes struct {
	Initialized bool `json:"initialized"` // 是否已初始化
}

// InitRes 初始化结果响应。
type InitRes struct {
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"` // 响应消息
}

// ThemeInstallStatusRes 首装主题安装任务状态。
type ThemeInstallStatusRes struct {
	Status       string     `json:"status"`        // not_started | running | success | failed
	Stage        string     `json:"stage"`         // fetching | installing | activating | done
	Message      string     `json:"message"`       // 失败原因或补充说明
	ThemeID      string     `json:"theme_id"`      // 主题标识
	ThemeName    string     `json:"theme_name"`    // 主题名
	ThemeVersion string     `json:"theme_version"` // 主题版本
	StartedAt    *time.Time `json:"started_at"`    // 开始时间
	FinishedAt   *time.Time `json:"finished_at"`   // 结束时间
}
