package controller

import (
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// CommentController 评论控制器结构体。
type CommentController struct {
	logic *logic.CommentLogic
}

// NewCommentController 创建 CommentController 实例。
func NewCommentController() *CommentController {
	return &CommentController{logic: logic.NewCommentLogic()}
}

// GetList 获取评论列表 GET /api/v1/comments
func (c *CommentController) GetList(ctx *gin.Context) {
	var r req.CommentListReq
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

// Reply 博主回复评论 POST /api/v1/comments/:id/reply
func (c *CommentController) Reply(ctx *gin.Context) {
	commentID := ctx.Param("id")
	bloggerID := ctx.GetString("user_id")
	var r req.ReplyCommentReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	result, err := c.logic.Reply(ctx, bloggerID, commentID, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// Delete 删除评论 DELETE /api/v1/comments/:id
func (c *CommentController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}
