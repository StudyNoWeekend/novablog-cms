package res

import "time"

// SystemConfigItemRes 单个系统配置项。
type SystemConfigItemRes struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// SystemConfigListRes 系统配置列表响应（站点信息分组）。
type SystemConfigListRes struct {
	Items []SystemConfigItemRes `json:"items"`
}

// SystemConfigUpdateRes 更新系统配置响应（site_url 变化时前端提示检查跨域白名单）。
type SystemConfigUpdateRes struct {
	SiteURLChanged bool `json:"site_url_changed"`
}

// SystemConfigExportRes 配置导出/导入载体（JSON 下载体与导入请求体同构）。
// Groups 按「分组 → 键 → 值」组织；敏感项在 include_secrets=false 时以占位符输出。
type SystemConfigExportRes struct {
	Version         int                          `json:"version"`
	ExportedAt      time.Time                    `json:"exported_at"`
	SecretsIncluded bool                         `json:"secrets_included"`
	Groups          map[string]map[string]string `json:"groups"`
}

// SystemConfigImportRes 配置导入响应（各分组实际写入的键数）。
type SystemConfigImportRes struct {
	Applied map[string]int `json:"applied"`
}
