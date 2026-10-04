package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// BookController 读书书架控制器结构体。
type BookController struct {
	logic *logic.BookLogic
}

// NewBookController 创建 BookController 实例。
func NewBookController() *BookController {
	return &BookController{logic: logic.NewBookLogic()}
}

// Create 创建书籍 POST /api/v1/books
func (c *BookController) Create(ctx *gin.Context) {
	var r req.CreateBookReq
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

// GetList 获取书籍列表 GET /api/v1/books
func (c *BookController) GetList(ctx *gin.Context) {
	var r req.BookListReq
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

// GetByID 获取书籍详情 GET /api/v1/books/:id
func (c *BookController) GetByID(ctx *gin.Context) {
	result, err := c.logic.GetByID(ctx, ctx.Param("id"))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新书籍 PUT /api/v1/books/:id
func (c *BookController) Update(ctx *gin.Context) {
	var r req.UpdateBookReq
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

// Delete 删除书籍 DELETE /api/v1/books/:id
func (c *BookController) Delete(ctx *gin.Context) {
	if err := c.logic.Delete(ctx, ctx.Param("id")); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
