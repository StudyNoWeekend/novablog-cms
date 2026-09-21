package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// OpenSourceController 开源作品控制器结构体。
type OpenSourceController struct {
	logic *logic.OpenSourceLogic
}

// NewOpenSourceController 创建 OpenSourceController 实例。
func NewOpenSourceController() *OpenSourceController {
	return &OpenSourceController{logic: logic.NewOpenSourceLogic()}
}

// Create 创建开源作品（自动拉取 README） POST /api/v1/open-sources
func (c *OpenSourceController) Create(ctx *gin.Context) {
	var r req.CreateOpenSourceReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.Create(ctx, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetList 获取开源作品列表 GET /api/v1/open-sources
func (c *OpenSourceController) GetList(ctx *gin.Context) {
	var r req.OpenSourceListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.GetList(ctx, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetByID 获取开源作品详情（含 README） GET /api/v1/open-sources/:id
func (c *OpenSourceController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetByID(ctx, id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新开源作品（仓库链接变更时自动重新拉取 README） PUT /api/v1/open-sources/:id
func (c *OpenSourceController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateOpenSourceReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.Update(ctx, id, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Refresh 重新拉取远端仓库元数据与 README POST /api/v1/open-sources/:id/refresh
func (c *OpenSourceController) Refresh(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.Refresh(ctx, id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Delete 删除开源作品 DELETE /api/v1/open-sources/:id
func (c *OpenSourceController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
