package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// CategoryController 分类控制器结构体。
type CategoryController struct {
	logic *logic.CategoryLogic
}

// NewCategoryController 创建 CategoryController 实例。
func NewCategoryController() *CategoryController {
	return &CategoryController{logic: logic.NewCategoryLogic()}
}

// Create 创建分类 POST /api/v1/categories
func (c *CategoryController) Create(ctx *gin.Context) {
	var r req.CreateCategoryReq
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

// GetAll 获取所有分类 GET /api/v1/categories
func (c *CategoryController) GetAll(ctx *gin.Context) {
	var r req.CategoryListReq
	_ = ctx.ShouldBindQuery(&r)
	result, err := c.logic.GetAll(ctx, r.Type)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetByID 根据ID获取分类 GET /api/v1/categories/:id
func (c *CategoryController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetByID(ctx, id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新分类 PUT /api/v1/categories/:id
func (c *CategoryController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateCategoryReq
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

// Delete 删除分类 DELETE /api/v1/categories/:id
func (c *CategoryController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
