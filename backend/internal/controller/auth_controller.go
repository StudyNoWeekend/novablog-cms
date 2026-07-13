// Package controller 定义 HTTP 控制器层。
package controller

import (
	"errors"
	"net/http"
	"strings"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthController 认证控制器结构体。
type AuthController struct {
	authLogic *logic.AuthLogic
}

// NewAuthController 创建 AuthController 实例。
func NewAuthController() *AuthController {
	return &AuthController{
		authLogic: logic.NewAuthLogic(),
	}
}

// Login 处理用户登录请求。
func (ctrl *AuthController) Login(c *gin.Context) {
	var loginReq req.LoginReq
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		logic.AuthLogger.Warn("登录请求参数错误", zap.Error(err))
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	loginRes, err := ctrl.authLogic.Login(c.Request.Context(), &loginReq)
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, loginRes)
}

// Refresh 处理刷新 Token 请求。
func (ctrl *AuthController) Refresh(c *gin.Context) {
	var refreshReq req.RefreshReq
	if err := c.ShouldBindJSON(&refreshReq); err != nil {
		logic.AuthLogger.Warn("刷新Token请求参数错误", zap.Error(err))
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	loginRes, err := ctrl.authLogic.Refresh(c.Request.Context(), refreshReq.RefreshToken)
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, loginRes)
}

// Logout 处理登出请求。
func (ctrl *AuthController) Logout(c *gin.Context) {
	// 从 Authorization 头中提取 Token
	token := extractBearerToken(c)
	if token == "" {
		response.Fail(c, enum.ErrUnauthorized.Code, enum.ErrUnauthorized.Msg, enum.ErrUnauthorized.HttpCode)
		return
	}

	if err := ctrl.authLogic.Logout(c.Request.Context(), token); err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, nil)
}

// ChangePassword 处理修改密码请求。
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	// 从上下文中获取用户 ID（由 AuthMiddleware 设置）
	userID, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, enum.ErrUnauthorized.Code, enum.ErrUnauthorized.Msg, enum.ErrUnauthorized.HttpCode)
		return
	}

	var changePwdReq req.ChangePasswordReq
	if err := c.ShouldBindJSON(&changePwdReq); err != nil {
		logic.AuthLogger.Warn("修改密码请求参数错误", zap.Error(err))
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	if err := ctrl.authLogic.ChangePassword(c.Request.Context(), userID.(string), changePwdReq.OldPassword, changePwdReq.NewPassword); err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, nil)
}

// extractBearerToken 从 Authorization 头中提取 Bearer Token。
func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// 确保导入了 http 包（用于保证编译通过）
var _ = http.StatusOK
