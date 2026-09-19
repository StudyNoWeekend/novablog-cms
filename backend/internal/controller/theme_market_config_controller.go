package controller

import (
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// ThemeMarketConfigController 官方主题市场配置控制器结构体。
type ThemeMarketConfigController struct {
	configLogic *logic.ThemeMarketConfigLogic
}

// NewThemeMarketConfigController 创建 ThemeMarketConfigController 实例。
func NewThemeMarketConfigController() *ThemeMarketConfigController {
	return &ThemeMarketConfigController{
		configLogic: logic.NewThemeMarketConfigLogic(),
	}
}

// GetConfig 获取官方市场配置 GET /theme-market-config
func (ctrl *ThemeMarketConfigController) GetConfig(c *gin.Context) {
	config, err := ctrl.configLogic.GetConfig(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, config)
}

// UpdateConfig 更新官方市场地址 PUT /theme-market-config
func (ctrl *ThemeMarketConfigController) UpdateConfig(c *gin.Context) {
	var updateReq req.UpdateThemeMarketConfigReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.HandleError(c, err)
		return
	}
	if err := ctrl.configLogic.UpdateConfig(c.Request.Context(), &updateReq); err != nil {
		response.HandleError(c, err)
		return
	}
	// 返回更新后的配置，便于前端直接刷新
	config, err := ctrl.configLogic.GetConfig(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, config)
}
