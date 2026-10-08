package controller

import (
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// MapConfigController 地图服务配置控制器结构体。
type MapConfigController struct {
	configLogic *logic.MapConfigLogic
}

// NewMapConfigController 创建 MapConfigController 实例。
func NewMapConfigController(cryptoKey string) *MapConfigController {
	return &MapConfigController{
		configLogic: logic.NewMapConfigLogic(cryptoKey),
	}
}

// GetConfig 获取地图服务配置 GET /map-config
func (ctrl *MapConfigController) GetConfig(c *gin.Context) {
	config, err := ctrl.configLogic.GetConfig(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, config)
}

// UpdateConfig 更新地图服务配置 PUT /map-config
func (ctrl *MapConfigController) UpdateConfig(c *gin.Context) {
	var updateReq req.UpdateMapConfigReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.HandleError(c, err)
		return
	}
	config, err := ctrl.configLogic.UpdateConfig(c.Request.Context(), &updateReq)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	// 返回更新后的配置，便于前端直接刷新
	response.Success(c, config)
}
