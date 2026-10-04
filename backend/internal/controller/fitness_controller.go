package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// FitnessController 健身训练控制器结构体。
type FitnessController struct {
	logic *logic.FitnessLogic
}

// NewFitnessController 创建 FitnessController 实例。
func NewFitnessController() *FitnessController {
	return &FitnessController{logic: logic.NewFitnessLogic()}
}

// Create 创建训练记录 POST /api/v1/fitness
func (c *FitnessController) Create(ctx *gin.Context) {
	var r req.CreateFitnessReq
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

// GetList 获取训练记录列表 GET /api/v1/fitness
func (c *FitnessController) GetList(ctx *gin.Context) {
	var r req.FitnessListReq
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

// GetByID 获取训练记录详情 GET /api/v1/fitness/:id
func (c *FitnessController) GetByID(ctx *gin.Context) {
	result, err := c.logic.GetByID(ctx, ctx.Param("id"))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新训练记录 PUT /api/v1/fitness/:id
func (c *FitnessController) Update(ctx *gin.Context) {
	var r req.UpdateFitnessReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.Update(ctx, ctx.Param("id"), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Delete 删除训练记录 DELETE /api/v1/fitness/:id
func (c *FitnessController) Delete(ctx *gin.Context) {
	if err := c.logic.Delete(ctx, ctx.Param("id")); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
