// Package enum 定义业务错误码和错误结构体。
package enum

import "fmt"

// BizError 业务错误结构体，包含错误码、错误信息和对应的 HTTP 状态码。
type BizError struct {
	Code     int    `json:"code"`     // 业务错误码
	Msg      string `json:"msg"`      // 错误信息
	HttpCode int    `json:"httpCode"` // HTTP 状态码
}

// Error 实现 error 接口，返回错误信息。
func (e *BizError) Error() string {
	return fmt.Sprintf("BizError[%d]: %s", e.Code, e.Msg)
}

// NewBizError 创建新的业务错误。
func NewBizError(code int, msg string, httpCode int) *BizError {
	return &BizError{
		Code:     code,
		Msg:      msg,
		HttpCode: httpCode,
	}
}

// 预定义业务错误码
var (
	// ErrInvalidParam 请求参数错误
	ErrInvalidParam = NewBizError(400001, "请求参数错误", 400)
	// ErrUnauthorized 未认证
	ErrUnauthorized = NewBizError(401000, "未认证，请先登录", 401)
	// ErrLoginFailed 用户名或密码错误
	ErrLoginFailed = NewBizError(401001, "用户名或密码错误", 401)
	// ErrTokenExpired Token 已过期
	ErrTokenExpired = NewBizError(401002, "Token 已过期", 401)
	// ErrTokenInvalid Token 无效
	ErrTokenInvalid = NewBizError(401003, "Token 无效", 401)
	// ErrForbidden 无权限访问该资源
	ErrForbidden = NewBizError(403000, "无权限访问该资源", 403)
	// ErrAlreadyInitialized 系统已初始化，无法重复创建
	ErrAlreadyInitialized = NewBizError(403001, "系统已初始化，无法重复创建", 403)
	// ErrIPBlocked IP 已被暂时限制访问
	ErrIPBlocked = NewBizError(403002, "您已被暂时限制访问，请稍后再试", 403)
	// ErrNotFound 资源不存在
	ErrNotFound = NewBizError(404001, "资源不存在", 404)
	// ErrInternalServer 系统内部错误
	ErrInternalServer = NewBizError(500001, "系统内部错误", 500)
	// ErrTooManyRequests 操作过于频繁
	ErrTooManyRequests = NewBizError(429001, "操作过于频繁，请稍后再试", 429)
)
