package res

import "time"

// CloudDriveConfigRes 云盘应用凭据响应（client_secret 脱敏）。
type CloudDriveConfigRes struct {
	ID              string    `json:"id"`
	Provider        string    `json:"provider"`
	ClientID        string    `json:"client_id"`
	ClientSecret    string    `json:"client_secret"` // 返回 "******"
	CallbackBaseURL string    `json:"callback_base_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CloudAuthorizeRes 发起授权响应。
type CloudAuthorizeRes struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// CloudDriveAccountRes 云盘账号响应（token/密码脱敏）。
type CloudDriveAccountRes struct {
	ID          string     `json:"id"`
	Provider    string     `json:"provider"` // webdav/baidu/aliyun
	OpenUserID  string     `json:"open_user_id"`
	DisplayName string     `json:"display_name"`
	BaseURL     string     `json:"base_url"`  // WebDAV
	Username    string     `json:"username"`  // WebDAV
	Status      string     `json:"status"`    // active/expired/revoked
	ExpiresAt   *time.Time `json:"expires_at"`
	QuotaTotal  int64      `json:"quota_total"`
	QuotaUsed   int64      `json:"quota_used"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CloudFileRes 云盘文件浏览响应条目。
type CloudFileRes struct {
	FileID   string `json:"file_id"`
	ParentID string `json:"parent_id"`
	Path     string `json:"path"`
	Name     string `json:"name"`
	IsDir    bool   `json:"is_dir"`
	Size     int64  `json:"size"`
	Ext      string `json:"ext"`
}

// CloudFileListRes 云盘文件浏览响应。
type CloudFileListRes struct {
	List       []CloudFileRes `json:"list"`
	NextCursor string         `json:"next_cursor"`
	Parent     string         `json:"parent"`
}

// CloudImportItemRes 导入条目响应。
type CloudImportItemRes struct {
	ID               string `json:"id"`
	SourceFileID     string `json:"source_file_id"`
	SourcePath       string `json:"source_path"`
	Filename         string `json:"filename"`
	FileSize         int64  `json:"file_size"`
	TransferredBytes int64  `json:"transferred_bytes"` // 已下载字节（下载中节流回写，前端逐文件进度/速度）
	Status           string `json:"status"`
	MediaID          string `json:"media_id"`
	Error            string `json:"error"`
}

// CloudImportTaskRes 导入任务响应。
type CloudImportTaskRes struct {
	ID             string     `json:"id"`
	AccountID      string     `json:"account_id"`
	Provider       string     `json:"provider"`
	TargetFolderID *string    `json:"target_folder_id"`
	Status         string     `json:"status"` // pending/running/completed/failed
	Total          int        `json:"total"`
	Succeeded      int        `json:"succeeded"`
	Failed         int        `json:"failed"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	Error          string     `json:"error"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CloudImportStatusRes 导入进度响应（任务 + 全部条目逐文件进度 + 失败明细）。
type CloudImportStatusRes struct {
	Task   CloudImportTaskRes   `json:"task"`
	Items  []CloudImportItemRes `json:"items"`  // 全部条目（含下载中条目的 transferred_bytes，供前端逐文件进度条）
	Failed []CloudImportItemRes `json:"failed"` // 失败明细（结果页折叠面板）
}
