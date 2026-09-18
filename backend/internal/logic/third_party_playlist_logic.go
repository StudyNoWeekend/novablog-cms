package logic

import (
	"context"
	"errors"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ThirdPartyPlaylistLogic 第三方歌单业务逻辑结构体。
type ThirdPartyPlaylistLogic struct {
	playlistModel *model.ThirdPartyPlaylistModel
	logger        *zap.Logger
}

// NewThirdPartyPlaylistLogic 创建 ThirdPartyPlaylistLogic 实例。
func NewThirdPartyPlaylistLogic() *ThirdPartyPlaylistLogic {
	return &ThirdPartyPlaylistLogic{
		playlistModel: model.NewThirdPartyPlaylist(),
		logger:        MusicLogger,
	}
}

// Create 创建第三方歌单。
func (l *ThirdPartyPlaylistLogic) Create(ctx context.Context, r *req.CreatePlaylistReq) (*res.PlaylistRes, error) {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	playlist := &model.ThirdPartyPlaylist{
		ID:          uuid.New().String(),
		Title:       r.Title,
		CoverURL:    r.CoverURL,
		Platform:    r.Platform,
		PlatformURL: r.PlatformURL,
		Description: r.Description,
		SortOrder:   r.SortOrder,
		Enabled:     enabled,
	}

	if err := l.playlistModel.Create(ctx, playlist); err != nil {
		l.logger.Error("创建第三方歌单失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	return l.toRes(playlist), nil
}

// GetList 分页获取第三方歌单列表。
func (l *ThirdPartyPlaylistLogic) GetList(ctx context.Context, r *req.PlaylistListReq) (*res.PageRes[res.PlaylistRes], error) {
	playlists, total, err := l.playlistModel.GetList(ctx, r.GetPage(), r.GetPageSize())
	if err != nil {
		l.logger.Error("获取第三方歌单列表失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	items := make([]res.PlaylistRes, len(playlists))
	for i, p := range playlists {
		items[i] = *l.toRes(&p)
	}

	return res.NewPageRes(items, total, r.GetPage(), r.GetPageSize()), nil
}

// GetByID 根据 ID 获取第三方歌单。
func (l *ThirdPartyPlaylistLogic) GetByID(ctx context.Context, id string) (*res.PlaylistRes, error) {
	playlist, err := l.playlistModel.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, enum.ErrNotFound
		}
		l.logger.Error("获取第三方歌单详情失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}
	return l.toRes(playlist), nil
}

// Update 更新第三方歌单。
func (l *ThirdPartyPlaylistLogic) Update(ctx context.Context, id string, r *req.UpdatePlaylistReq) (*res.PlaylistRes, error) {
	playlist, err := l.playlistModel.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, enum.ErrNotFound
		}
		l.logger.Error("获取第三方歌单详情失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	if r.Title != nil {
		playlist.Title = *r.Title
	}
	if r.CoverURL != nil {
		playlist.CoverURL = *r.CoverURL
	}
	if r.Platform != nil {
		playlist.Platform = *r.Platform
	}
	if r.PlatformURL != nil {
		playlist.PlatformURL = *r.PlatformURL
	}
	if r.Description != nil {
		playlist.Description = *r.Description
	}
	if r.SortOrder != nil {
		playlist.SortOrder = *r.SortOrder
	}
	if r.Enabled != nil {
		playlist.Enabled = *r.Enabled
	}

	if err := l.playlistModel.Update(ctx, playlist); err != nil {
		l.logger.Error("更新第三方歌单失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	return l.toRes(playlist), nil
}

// Delete 删除第三方歌单。
func (l *ThirdPartyPlaylistLogic) Delete(ctx context.Context, id string) error {
	if err := l.playlistModel.Delete(ctx, id); err != nil {
		l.logger.Error("删除第三方歌单失败", zap.Error(err))
		return enum.ErrInternalServer
	}
	return nil
}

// GetPublicList 获取前台展示的第三方歌单列表。
func (l *ThirdPartyPlaylistLogic) GetPublicList(ctx context.Context) ([]res.PlaylistRes, error) {
	playlists, err := l.playlistModel.GetPublicList(ctx)
	if err != nil {
		l.logger.Error("获取前台第三方歌单列表失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	items := make([]res.PlaylistRes, len(playlists))
	for i, p := range playlists {
		items[i] = *l.toRes(&p)
	}
	return items, nil
}

// toRes 将模型转换为响应结构体。
func (l *ThirdPartyPlaylistLogic) toRes(p *model.ThirdPartyPlaylist) *res.PlaylistRes {
	return &res.PlaylistRes{
		ID:          p.ID,
		Title:       p.Title,
		CoverURL:    p.CoverURL,
		Platform:    p.Platform,
		PlatformURL: p.PlatformURL,
		Description: p.Description,
		SortOrder:   p.SortOrder,
		Enabled:     p.Enabled,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
