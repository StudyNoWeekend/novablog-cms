// Package req 定义请求 DTO（数据传输对象）。
package req

// ThemeInstallReq 从官方市场安装主题请求。
type ThemeInstallReq struct {
	ThemeID int64  `json:"theme_id" binding:"required,min=1"`  // 官方市场主题 ID
	Version string `json:"version" binding:"omitempty,max=50"` // 指定版本（空=市场最新）
	Force   bool   `json:"force"`                              // 同版本已安装时覆盖重装（激活中的实例拒绝覆盖）
}
