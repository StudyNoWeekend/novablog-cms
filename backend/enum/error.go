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
	// ErrMarketAuthFailed 官方主题市场登录状态已失效（注意：不复用 HTTP 401，避免与后台自身认证刷新流程冲突）
	ErrMarketAuthFailed = NewBizError(401101, "官方账号登录已失效，请重新登录", 400)
	// ErrMarketLoginFailed 官方主题市场登录凭据错误
	ErrMarketLoginFailed = NewBizError(401102, "官方邮箱或密码错误", 400)
	// ErrMarketBaseURLInvalid 官方主题市场地址不合法
	ErrMarketBaseURLInvalid = NewBizError(400103, "官方地址不合法，请检查输入", 400)
	// ErrMarketUpstream 官方主题市场服务不可用
	ErrMarketUpstream = NewBizError(400102, "官方主题市场服务不可用，请稍后再试", 400)
	// ErrInternalServer 系统内部错误
	ErrInternalServer = NewBizError(500001, "系统内部错误", 500)
)
