package controller

import (
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// ArticleController 文章控制器结构体。
type ArticleController struct {
	logic *logic.ArticleLogic
}

// NewArticleController 创建 ArticleController 实例。
func NewArticleController() *ArticleController {
	return &ArticleController{logic: logic.NewArticleLogic()}
}

// Create 创建文章 POST /api/v1/articles
func (c *ArticleController) Create(ctx *gin.Context) {
	var r req.CreateArticleReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	if r.Status == 0 {
		r.Status = 1 // 默认草稿
	}
	result, err := c.logic.Create(ctx, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetList 获取文章列表 GET /api/v1/articles
func (c *ArticleController) GetList(ctx *gin.Context) {
	var r req.ArticleListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误")
		return
	}
	result, err := c.logic.GetList(ctx, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetDetail 获取文章详情 GET /api/v1/articles/:id
func (c *ArticleController) GetDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetDetail(ctx, id)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// Update 更新文章 PUT /api/v1/articles/:id
func (c *ArticleController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateArticleReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	result, err := c.logic.Update(ctx, id, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// UpdateStatus 更新文章状态 PUT /api/v1/articles/:id/status
func (c *ArticleController) UpdateStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateStatusReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: status 必须为 1/2/3")
		return
	}
	if err := c.logic.UpdateStatus(ctx, id, &r); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// Delete 删除文章 DELETE /api/v1/articles/:id
func (c *ArticleController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}
