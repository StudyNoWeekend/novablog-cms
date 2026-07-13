package controller

import (
	"cus-cms/internal/dto/req"
	"cus-cms/internal/logic"
	"cus-cms/utils/response"

	"github.com/gin-gonic/gin"
)

// StorageController 存储配置控制器结构体。
type StorageController struct {
	logic *logic.StorageLogic
}

// NewStorageController 创建 StorageController 实例。
func NewStorageController(logic *logic.StorageLogic) *StorageController {
	return &StorageController{logic: logic}
}

// GetConfigs 获取所有存储配置 GET /storage/config
func (c *StorageController) GetConfigs(ctx *gin.Context) {
	configs, err := c.logic.GetAllConfigs(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, configs)
}

// UpsertConfig 创建或更新存储配置 POST /storage/config
func (c *StorageController) UpsertConfig(ctx *gin.Context) {
	var r req.StorageConfigReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "请求参数错误")
		return
	}
	if err := c.logic.UpsertConfig(ctx.Request.Context(), &r); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// DeleteConfig 删除存储配置 DELETE /storage/config/:provider
func (c *StorageController) DeleteConfig(ctx *gin.Context) {
	provider := ctx.Param("provider")
	if err := c.logic.DeleteConfig(ctx.Request.Context(), provider); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// ActivateConfig 激活存储配置 PUT /storage/config/:provider/activate
func (c *StorageController) ActivateConfig(ctx *gin.Context) {
	provider := ctx.Param("provider")
	if err := c.logic.ActivateConfig(ctx.Request.Context(), provider); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// TestConfig 测试存储连通性 POST /storage/config/test
func (c *StorageController) TestConfig(ctx *gin.Context) {
	var r req.StorageTestReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "请求参数错误")
		return
	}
	if err := c.logic.TestConfig(ctx.Request.Context(), &r); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}
