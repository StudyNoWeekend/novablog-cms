// Package req 定义请求 DTO（数据传输对象）。
package req

// LoginReq 登录请求参数。
type LoginReq struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// RefreshReq 刷新 Token 请求参数。
type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // 刷新令牌
}

// ChangePasswordReq 修改密码请求参数。
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"` // 旧密码
	NewPassword string `json:"new_password" binding:"required"` // 新密码
}
