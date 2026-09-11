package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// TagController 标签控制器结构体。
type TagController struct {
	logic *logic.TagLogic
}

// NewTagController 创建 TagController 实例。
func NewTagController() *TagController {
	return &TagController{logic: logic.NewTagLogic()}
}

// Create 创建标签 POST /api/v1/tags
func (c *TagController) Create(ctx *gin.Context) {
	var r req.CreateTagReq
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

// GetAll 获取所有标签 GET /api/v1/tags
func (c *TagController) GetAll(ctx *gin.Context) {
	result, err := c.logic.GetAll(ctx)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetByID 根据ID获取标签 GET /api/v1/tags/:id
func (c *TagController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetByID(ctx, id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新标签 PUT /api/v1/tags/:id
func (c *TagController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateTagReq
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

// Delete 删除标签 DELETE /api/v1/tags/:id
func (c *TagController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
