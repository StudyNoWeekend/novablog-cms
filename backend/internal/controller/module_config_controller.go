package controller

import (
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// ModuleConfigController 模块开关配置控制器结构体。
type ModuleConfigController struct {
	moduleConfigLogic *logic.ModuleConfigLogic
}

// NewModuleConfigController 创建 ModuleConfigController 实例。
func NewModuleConfigController() *ModuleConfigController {
	return &ModuleConfigController{
		moduleConfigLogic: logic.NewModuleConfigLogic(),
	}
}

// GetConfig 获取模块开关配置 GET /module-config
func (ctrl *ModuleConfigController) GetConfig(c *gin.Context) {
	config, err := ctrl.moduleConfigLogic.GetConfig(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, config)
}

// UpdateConfig 更新模块开关配置 PUT /module-config
func (ctrl *ModuleConfigController) UpdateConfig(c *gin.Context) {
	var updateReq req.UpdateModuleConfigReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.HandleError(c, err)
		return
	}
	if err := ctrl.moduleConfigLogic.UpdateConfig(c.Request.Context(), &updateReq); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, nil)
}
