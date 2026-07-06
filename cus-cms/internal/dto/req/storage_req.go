package req

// StorageConfigReq 存储配置创建/更新请求
type StorageConfigReq struct {
	Provider     string `json:"provider" binding:"required,oneof=aliyun tencent minio"`
	Endpoint     string `json:"endpoint" binding:"required"`
	Region       string `json:"region"`
	Bucket       string `json:"bucket"`
	AccessKey    string `json:"access_key" binding:"required"`
	AccessSecret string `json:"access_secret" binding:"required"`
	PathPrefix   string `json:"path_prefix"`
	CustomDomain string `json:"custom_domain"`
	Extra        string `json:"extra"` // JSON字符串，如 {"use_ssl":true}
}

// StorageTestReq 存储连通性测试请求（不落库）
type StorageTestReq struct {
	Provider     string `json:"provider" binding:"required,oneof=aliyun tencent minio"`
	Endpoint     string `json:"endpoint" binding:"required"`
	Region       string `json:"region"`
	Bucket       string `json:"bucket"`
	AccessKey    string `json:"access_key" binding:"required"`
	AccessSecret string `json:"access_secret" binding:"required"`
	PathPrefix   string `json:"path_prefix"`
	CustomDomain string `json:"custom_domain"`
	Extra        string `json:"extra"`
}

// MigrationAnalyzeReq 迁移分析请求
type MigrationAnalyzeReq struct {
	TargetProvider string `json:"target_provider" binding:"required"` // aliyun/tencent/minio
}

// MigrationStartReq 迁移执行请求
type MigrationStartReq struct {
	TargetProvider string   `json:"target_provider" binding:"required"`
	MediaIDs       []string `json:"media_ids"` // 指定 ID 迁移（与 All 互斥）
	All            bool     `json:"all"`      // 一键迁移所有缺失素材
}
