package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// PlaylistController 第三方歌单控制器结构体。
type PlaylistController struct {
	logic *logic.ThirdPartyPlaylistLogic
}

// NewPlaylistController 创建 PlaylistController 实例。
func NewPlaylistController() *PlaylistController {
	return &PlaylistController{logic: logic.NewThirdPartyPlaylistLogic()}
}

// Create 创建第三方歌单 POST /api/v1/playlists
func (c *PlaylistController) Create(ctx *gin.Context) {
	var r req.CreatePlaylistReq
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

// GetList 获取第三方歌单列表 GET /api/v1/playlists
func (c *PlaylistController) GetList(ctx *gin.Context) {
	var r req.PlaylistListReq
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

// GetByID 根据ID获取第三方歌单 GET /api/v1/playlists/:id
func (c *PlaylistController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetByID(ctx, id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新第三方歌单 PUT /api/v1/playlists/:id
func (c *PlaylistController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdatePlaylistReq
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

// Delete 删除第三方歌单 DELETE /api/v1/playlists/:id
func (c *PlaylistController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
