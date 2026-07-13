package controller

import (
	"errors"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// SecurityController 安全管理控制器结构体。
type SecurityController struct {
	securityLogic *logic.SecurityLogic
}

// NewSecurityController 创建 SecurityController 实例。
func NewSecurityController() *SecurityController {
	return &SecurityController{
		securityLogic: logic.NewSecurityLogic(),
	}
}

// GetConfig 获取安全配置 GET /security/config
func (ctrl *SecurityController) GetConfig(c *gin.Context) {
	config, err := ctrl.securityLogic.GetConfig(c.Request.Context())
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, config)
}

// UpdateConfig 更新安全配置 PUT /security/config
func (ctrl *SecurityController) UpdateConfig(c *gin.Context) {
	var updateReq req.UpdateSecurityConfigReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	if err := ctrl.securityLogic.UpdateConfig(c.Request.Context(), &updateReq); err != nil {
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

// GetBlacklist 获取 IP 黑名单列表 GET /security/blacklist
func (ctrl *SecurityController) GetBlacklist(c *gin.Context) {
	var queryReq req.BlacklistQueryReq
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	result, err := ctrl.securityLogic.GetBlacklist(c.Request.Context(), queryReq.GetPage(), queryReq.GetPageSize())
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, result)
}

// UnbanIP 解封 IP DELETE /security/blacklist/:ip
func (ctrl *SecurityController) UnbanIP(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	if err := ctrl.securityLogic.UnbanIP(c.Request.Context(), ip); err != nil {
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

// GetStats 获取安全统计数据 GET /security/stats
func (ctrl *SecurityController) GetStats(c *gin.Context) {
	stats, err := ctrl.securityLogic.GetStats(c.Request.Context())
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, stats)
}
