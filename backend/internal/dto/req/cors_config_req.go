package req

// UpdateCorsConfigReq 更新跨域配置请求。
type UpdateCorsConfigReq struct {
	AllowedOrigins *string `json:"allowed_origins"`
}
