package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/internal/storage"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// MusicController 音乐播放器控制器结构体。
type MusicController struct {
	logic *logic.MusicLogic
}

// NewMusicController 创建 MusicController 实例。
func NewMusicController(manager *storage.Manager) *MusicController {
	return &MusicController{logic: logic.NewMusicLogic(manager)}
}

// CreateSong 创建歌曲 POST /api/v1/music/songs
func (c *MusicController) CreateSong(ctx *gin.Context) {
	var r req.CreateSongReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.CreateSong(ctx, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// BatchCreateSongs 批量创建歌曲 POST /api/v1/music/songs/batch
func (c *MusicController) BatchCreateSongs(ctx *gin.Context) {
	var r req.BatchCreateSongReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.BatchCreateSongs(ctx, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetSongList 获取歌曲列表 GET /api/v1/music/songs
func (c *MusicController) GetSongList(ctx *gin.Context) {
	var r req.SongListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.GetSongList(ctx, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetSongByID 根据ID获取歌曲 GET /api/v1/music/songs/:id
func (c *MusicController) GetSongByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetSongByID(ctx, id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// UpdateSong 更新歌曲 PUT /api/v1/music/songs/:id
func (c *MusicController) UpdateSong(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateSongReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.UpdateSong(ctx, id, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// DeleteSong 删除歌曲 DELETE /api/v1/music/songs/:id
func (c *MusicController) DeleteSong(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.DeleteSong(ctx, id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// ParseMusic 解析B站音乐链接 POST /api/v1/music/parse
func (c *MusicController) ParseMusic(ctx *gin.Context) {
	var r req.ParseMusicReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	taskID, err := c.logic.StartParse(ctx, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, gin.H{"task_id": taskID})
}

// GetParseStatus 获取解析任务状态 GET /api/v1/music/parse/:task_id
func (c *MusicController) GetParseStatus(ctx *gin.Context) {
	taskID := ctx.Param("task_id")
	result, err := c.logic.GetParseTask(ctx, taskID)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetAudioURL 获取播放地址（B站外链播放器） GET /api/v1/music/audio-url/:song_id
func (c *MusicController) GetAudioURL(ctx *gin.Context) {
	songID := ctx.Param("song_id")

	url, err := c.logic.GetAudioURL(ctx, songID)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, gin.H{"url": url})
}
