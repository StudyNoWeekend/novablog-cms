// Package req 定义请求 DTO（数据传输对象）。
package req

// InitReq 初始化博主账号请求。
type InitReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"` // 用户名
	Password string `json:"password" binding:"required,min=6"`        // 密码
	Nickname string `json:"nickname" binding:"omitempty,max=50"`      // 昵称
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
