package res

import "time"

// CorsConfigRes 跨域配置响应。
type CorsConfigRes struct {
	AllowedOrigins string    `json:"allowed_origins"`
	UpdatedAt      time.Time `json:"updated_at"`
}
