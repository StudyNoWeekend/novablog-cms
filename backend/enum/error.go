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
	// ErrThemeMarketNotConfigured 未配置官方市场地址，服务端无法拉取主题
	ErrThemeMarketNotConfigured = NewBizError(400201, "未配置官方市场地址（themes.market_base_url），无法从官方获取主题", 400)
	// ErrThemeManifestInvalid 主题清单校验失败
	ErrThemeManifestInvalid = NewBizError(400202, "主题清单（theme.json）校验失败", 400)
	// ErrThemeEngineUnsupported 主题引擎不受支持
	ErrThemeEngineUnsupported = NewBizError(400203, "主题引擎不受支持，仅支持静态导出主题（next-static）", 400)
	// ErrThemeAPICompatIncompatible 主题与当前系统 API 版本不兼容
	ErrThemeAPICompatIncompatible = NewBizError(400204, "主题与当前系统 API 版本不兼容", 400)
	// ErrThemeChecksumMismatch 制品校验和不匹配
	ErrThemeChecksumMismatch = NewBizError(400205, "制品校验和不匹配，下载可能被篡改", 400)
	// ErrThemeArtifactNotFound 未找到预构建制品
	ErrThemeArtifactNotFound = NewBizError(400206, "未找到预构建制品，请让主题作者在 Release 附带 tar.gz 制品", 400)
	// ErrThemeVersionExists 该主题版本已安装
	ErrThemeVersionExists = NewBizError(400207, "该主题版本已安装", 400)
	// ErrThemeActiveCannotDelete 激活中的主题禁止卸载
	ErrThemeActiveCannotDelete = NewBizError(400208, "该主题正在使用中，请先切换到其他主题", 400)
	// ErrThemeNotFound 主题不存在
	ErrThemeNotFound = NewBizError(404002, "主题不存在", 404)
	// ErrThemeInstallRunning 已有主题安装任务进行中
	ErrThemeInstallRunning = NewBizError(400209, "已有主题安装任务在进行中，请稍候", 400)
	// ErrAlreadyLatestTheme 该主题已是最新版本
	ErrAlreadyLatestTheme = NewBizError(400210, "当前已是最新版本", 400)
	// ErrInternalServer 系统内部错误
	ErrInternalServer = NewBizError(500001, "系统内部错误", 500)
)
