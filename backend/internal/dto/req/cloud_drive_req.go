package req

// CloudDriveConfigReq 云盘应用凭据保存请求（百度/阿里各一行）。
type CloudDriveConfigReq struct {
	Provider        string `json:"provider" binding:"required,oneof=baidu aliyun"`
	ClientID        string `json:"client_id"`
	ClientSecret    string `json:"client_secret"` // 编辑时留空表示不修改
	CallbackBaseURL string `json:"callback_base_url"`
}

// WebDAVAccountReq WebDAV 账号创建请求（账密即用，无需 OAuth）。
type WebDAVAccountReq struct {
	BaseURL  string `json:"base_url" binding:"required,url"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// WebDAVTestReq WebDAV 连通性测试请求（不落库）。
type WebDAVTestReq struct {
	BaseURL  string `json:"base_url" binding:"required,url"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CloudFile 云盘文件（导入请求载荷）。
type CloudFile struct {
	FileID string `json:"file_id" binding:"required"`
	Path   string `json:"path"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
}

// CloudImportStartReq 发起云盘导入请求。
type CloudImportStartReq struct {
	AccountID string      `json:"account_id" binding:"required"`
	Files     []CloudFile `json:"files" binding:"required,min=1"`
	FolderID  *string     `json:"folder_id"` // NULL=媒体库根目录
}
