package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// RecipeController 美食菜谱控制器结构体。
type RecipeController struct {
	logic *logic.RecipeLogic
}

// NewRecipeController 创建 RecipeController 实例。
func NewRecipeController() *RecipeController {
	return &RecipeController{logic: logic.NewRecipeLogic()}
}

// Create 创建菜谱 POST /api/v1/recipes
func (c *RecipeController) Create(ctx *gin.Context) {
	var r req.CreateRecipeReq
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

// GetList 获取菜谱列表 GET /api/v1/recipes
func (c *RecipeController) GetList(ctx *gin.Context) {
	var r req.RecipeListReq
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

// GetByID 获取菜谱详情 GET /api/v1/recipes/:id
func (c *RecipeController) GetByID(ctx *gin.Context) {
	result, err := c.logic.GetByID(ctx, ctx.Param("id"))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新菜谱 PUT /api/v1/recipes/:id
func (c *RecipeController) Update(ctx *gin.Context) {
	var r req.UpdateRecipeReq
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

// Delete 删除菜谱 DELETE /api/v1/recipes/:id
func (c *RecipeController) Delete(ctx *gin.Context) {
	if err := c.logic.Delete(ctx, ctx.Param("id")); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
