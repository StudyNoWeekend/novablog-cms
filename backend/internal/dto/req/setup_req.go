// Package req 定义请求 DTO（数据传输对象）。
package req

// InitReq 初始化博主账号请求。
type InitReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"` // 用户名
	Password string `json:"password" binding:"required,min=6"`        // 密码
	Nickname string `json:"nickname" binding:"omitempty,max=50"`      // 昵称
	Role     string `json:"role" binding:"omitempty,max=50"`          // 创作方向角色 key（如 tech/travel），在 logic 层按白名单校验
	// InitCode 官方部署安装码：config.yaml 配置了 install.init_code 时必填，
	// 未配置（自部署）时忽略
	InitCode string `json:"init_code" binding:"omitempty,max=128"`
	// Modules 模块开关覆盖（key 为 module_configs 开关名，如 "open_source_enabled"）。
	// 非 nil 时以此为准应用最终开关；为 nil 且 Role 非空时按角色预设应用。
	Modules map[string]bool `json:"modules"`
}

// SetupStorageReq 首装向导存储配置请求。
// provider 为空或 "local" 表示跳过，保持 config.yaml 的本地存储配置。
type SetupStorageReq struct {
	Provider     string `json:"provider" binding:"omitempty,oneof=aliyun tencent minio local"`
	Endpoint     string `json:"endpoint"`
	Region       string `json:"region"`
	Bucket       string `json:"bucket"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
	PathPrefix   string `json:"path_prefix"`
	CustomDomain string `json:"custom_domain"`
	Extra        string `json:"extra"` // JSON 字符串，如 {"use_ssl":true}
}
