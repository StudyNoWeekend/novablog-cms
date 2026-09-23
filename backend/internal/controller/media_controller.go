package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/internal/storage"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// MediaController 媒体控制器结构体。
type MediaController struct {
	logic *logic.MediaLogic
}

// NewMediaController 创建 MediaController 实例。
func NewMediaController(manager *storage.Manager) *MediaController {
	return &MediaController{logic: logic.NewMediaLogic(manager)}
}

// Upload 上传文件 POST /api/v1/media/upload
func (c *MediaController) Upload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	// 限制文件大小 100MB
	if file.Size > 100*1024*1024 {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	media, err := c.logic.UploadFile(ctx, file)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, media)
}

// GetList 获取媒体列表 GET /api/v1/media
func (c *MediaController) GetList(ctx *gin.Context) {
	var req req.MediaListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.GetList(ctx, &req)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Delete 删除媒体 DELETE /api/v1/media/:id?force=true
// 被内容引用时需携带 force=true 强制删除。
func (c *MediaController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	force := ctx.Query("force") == "true"
	if err := c.logic.Delete(ctx, id, force); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// GetUsages 获取媒体引用情况 GET /api/v1/media/:id/usages
func (c *MediaController) GetUsages(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetUsages(ctx, id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// CreatePreset 创建媒体预设 POST /api/v1/media/preset
func (c *MediaController) CreatePreset(ctx *gin.Context) {
	var req req.CreatePresetReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	// 限制文件大小 100MB
	if file.Size > 100*1024*1024 {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	preset, err := c.logic.CreatePreset(ctx, &req, file)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, preset)
}

// UploadWithPreset 上传原图并自动生成预设 POST /api/v1/media/upload-with-preset
func (c *MediaController) UploadWithPreset(ctx *gin.Context) {
	var req req.UploadWithPresetReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	// 限制文件大小 100MB
	if file.Size > 100*1024*1024 {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	result, err := c.logic.UploadWithPreset(ctx, &req, file)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetPresets 获取原图下的所有预设 GET /api/v1/media/:id/presets
func (c *MediaController) GetPresets(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetPresetsByMediaID(ctx, id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// DeletePreset 删除预设 DELETE /api/v1/media/preset/:id
func (c *MediaController) DeletePreset(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.DeletePreset(ctx, id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
