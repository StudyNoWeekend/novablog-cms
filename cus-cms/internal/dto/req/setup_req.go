// Package req 定义请求 DTO（数据传输对象）。
package req

// InitReq 初始化博主账号请求。
type InitReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"` // 用户名
	Password string `json:"password" binding:"required,min=6"`        // 密码
	Nickname string `json:"nickname" binding:"omitempty,max=50"`     // 昵称
}