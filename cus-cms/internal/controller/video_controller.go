package controller

import (
	"cus-cms/internal/dto/req"
	"cus-cms/internal/logic"
	"cus-cms/utils/response"

	"github.com/gin-gonic/gin"
)

// VideoController 视频作品控制器结构体。
type VideoController struct {
	logic *logic.VideoLogic
}

// NewVideoController 创建 VideoController 实例。
func NewVideoController() *VideoController {
	return &VideoController{logic: logic.NewVideoLogic()}
}

// Create 创建视频作品 POST /api/v1/videos
func (c *VideoController) Create(ctx *gin.Context) {
	var r req.CreateVideoReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	result, err := c.logic.Create(ctx, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetList 获取视频作品列表 GET /api/v1/videos
func (c *VideoController) GetList(ctx *gin.Context) {
	var r req.VideoListReq
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

// GetByID 获取视频作品详情 GET /api/v1/videos/:id
func (c *VideoController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetByID(ctx, id)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// Update 更新视频作品 PUT /api/v1/videos/:id
func (c *VideoController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateVideoReq
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

// Delete 删除视频作品 DELETE /api/v1/videos/:id
func (c *VideoController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// Parse 解析视频元信息 POST /api/v1/videos/parse
func (c *VideoController) Parse(ctx *gin.Context) {
	var r req.ParseVideoReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	result, err := c.logic.ParseVideo(ctx, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}
