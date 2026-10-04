package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// TechStackController 技术栈控制器结构体。
type TechStackController struct {
	logic *logic.TechStackLogic
}

// NewTechStackController 创建 TechStackController 实例。
func NewTechStackController() *TechStackController {
	return &TechStackController{logic: logic.NewTechStackLogic()}
}

// Create 创建技术栈条目 POST /api/v1/tech-stacks
func (c *TechStackController) Create(ctx *gin.Context) {
	var r req.CreateTechStackReq
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

// GetList 获取技术栈条目列表 GET /api/v1/tech-stacks
func (c *TechStackController) GetList(ctx *gin.Context) {
	var r req.TechStackListReq
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

// GetByID 获取技术栈条目详情 GET /api/v1/tech-stacks/:id
func (c *TechStackController) GetByID(ctx *gin.Context) {
	result, err := c.logic.GetByID(ctx, ctx.Param("id"))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新技术栈条目 PUT /api/v1/tech-stacks/:id
func (c *TechStackController) Update(ctx *gin.Context) {
	var r req.UpdateTechStackReq
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

// Delete 删除技术栈条目 DELETE /api/v1/tech-stacks/:id
func (c *TechStackController) Delete(ctx *gin.Context) {
	if err := c.logic.Delete(ctx, ctx.Param("id")); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
