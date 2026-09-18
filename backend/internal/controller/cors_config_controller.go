package controller

import (
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// CorsConfigController 跨域配置控制器结构体。
type CorsConfigController struct {
	corsConfigLogic *logic.CorsConfigLogic
}

// NewCorsConfigController 创建 CorsConfigController 实例。
func NewCorsConfigController() *CorsConfigController {
	return &CorsConfigController{
		corsConfigLogic: logic.NewCorsConfigLogic(),
	}
}

// GetConfig 获取跨域配置 GET /cors-config
func (ctrl *CorsConfigController) GetConfig(c *gin.Context) {
	config, err := ctrl.corsConfigLogic.GetConfig(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, config)
}

// UpdateConfig 更新跨域配置 PUT /cors-config
func (ctrl *CorsConfigController) UpdateConfig(c *gin.Context) {
	var updateReq req.UpdateCorsConfigReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.HandleError(c, err)
		return
	}
	if err := ctrl.corsConfigLogic.UpdateConfig(c.Request.Context(), &updateReq); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, nil)
}
