package res

import "time"

// MapConfigRes 地图服务配置响应结构体。
// 安全密钥返回解密后的真实值：管理端选点器需要注入 _AMapSecurityConfig，暴露面与原编译期 env 一致。
type MapConfigRes struct {
	AmapKey          string    `json:"amap_key"`
	AmapSecurityCode string    `json:"amap_security_code"`
	GoogleKey        string    `json:"google_key"`
	UpdatedAt        time.Time `json:"updated_at"`
}
