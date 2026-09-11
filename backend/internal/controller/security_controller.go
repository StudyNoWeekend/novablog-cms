package controller

import (
	"strconv"

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
		response.HandleError(c, err)
		return
	}

	response.Success(c, config)
}

// UpdateConfig 更新安全配置 PUT /security/config
func (ctrl *SecurityController) UpdateConfig(c *gin.Context) {
	var updateReq req.UpdateSecurityConfigReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.securityLogic.UpdateConfig(c.Request.Context(), &updateReq); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// CreateBlacklist 创建黑名单 POST /security/blacklist
func (ctrl *SecurityController) CreateBlacklist(c *gin.Context) {
	var r req.BlacklistReq
	if err := c.ShouldBindJSON(&r); err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.securityLogic.CreateBlacklist(c.Request.Context(), &r); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteBlacklist 删除黑名单 DELETE /security/blacklist/:id
func (ctrl *SecurityController) DeleteBlacklist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.securityLogic.DeleteBlacklist(c.Request.Context(), uint(id)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateBlacklist 更新黑名单 PUT /security/blacklist/:id
func (ctrl *SecurityController) UpdateBlacklist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var r req.BlacklistReq
	if err := c.ShouldBindJSON(&r); err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.securityLogic.UpdateBlacklist(c.Request.Context(), uint(id), &r); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetBlacklist 获取黑名单详情 GET /security/blacklist/:id
func (ctrl *SecurityController) GetBlacklist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := ctrl.securityLogic.GetBlacklist(c.Request.Context(), uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// ListBlacklists 获取黑名单列表 GET /security/blacklists
func (ctrl *SecurityController) ListBlacklists(c *gin.Context) {
	var r req.ListBlacklistReq
	if err := c.ShouldBindQuery(&r); err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := ctrl.securityLogic.ListBlacklists(c.Request.Context(), &r)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// GetAccessStatistics 获取 IP 访问统计 GET /security/access-stats
func (ctrl *SecurityController) GetAccessStatistics(c *gin.Context) {
	var r req.ListAccessLogReq
	if err := c.ShouldBindQuery(&r); err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := ctrl.securityLogic.GetAccessStatistics(c.Request.Context(), &r)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}
