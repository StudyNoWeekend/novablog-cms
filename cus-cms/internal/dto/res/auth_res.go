// Package res 定义响应 DTO（数据传输对象）。
package res

// LoginRes 登录响应数据。
type LoginRes struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	ExpiresIn    int64  `json:"expires_in"`    // 过期时间（秒）
}
