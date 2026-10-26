package req

// UpdateSystemConfigReq 更新通用系统配置请求（键值部分更新；仅接受白名单键）。
type UpdateSystemConfigReq struct {
	Updates map[string]*string `json:"updates" binding:"required"`
}

// ImportSystemConfigReq 导入系统配置请求（结构同导出 JSON）。
type ImportSystemConfigReq struct {
	Version int                          `json:"version"`
	Groups  map[string]map[string]string `json:"groups"`
}
