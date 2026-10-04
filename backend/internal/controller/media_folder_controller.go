package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// MediaFolderController 媒体文件夹控制器结构体。
type MediaFolderController struct {
	logic *logic.MediaFolderLogic
}

// NewMediaFolderController 创建 MediaFolderController 实例。
func NewMediaFolderController(folderLogic *logic.MediaFolderLogic) *MediaFolderController {
	return &MediaFolderController{logic: folderLogic}
}

// Tree 获取文件夹树 GET /api/v1/media/folders/tree
func (c *MediaFolderController) Tree(ctx *gin.Context) {
	result, err := c.logic.Tree(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Create 创建文件夹 POST /api/v1/media/folders
func (c *MediaFolderController) Create(ctx *gin.Context) {
	var r req.MediaFolderCreateReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.Create(ctx.Request.Context(), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新文件夹（重命名/移动） PUT /api/v1/media/folders/:id
func (c *MediaFolderController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.MediaFolderUpdateReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.Update(ctx.Request.Context(), id, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Delete 删除空文件夹 DELETE /api/v1/media/folders/:id
func (c *MediaFolderController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx.Request.Context(), id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
