package logic

import (
	"context"
	"errors"
	"strings"

	"novablog/assets/platformlogos"
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

	// 封面：媒体库引用（ID 索引）与外链互斥写入
	coverMediaID, coverPresetID, coverURL, err := normalizeMediaRef(ctx, r.CoverMediaID, r.CoverPresetID, r.CoverURL)
	if err != nil {
		return nil, err
	}

	playlist := &model.ThirdPartyPlaylist{
		ID:            uuid.New().String(),
		Title:         r.Title,
		CoverURL:      coverURL,
		CoverMediaID:  mediaRefPtr(coverMediaID),
		CoverPresetID: mediaRefPtr(coverPresetID),
		Platform:      r.Platform,
		PlatformURL:   r.PlatformURL,
		Description:   r.Description,
		SortOrder:     r.SortOrder,
		Enabled:       enabled,
	}

	if err := l.playlistModel.Create(ctx, playlist); err != nil {
		l.logger.Error("创建第三方歌单失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	// 解析封面媒体引用为当前完整 URL（外链行保持原值）
	l.resolveCoverURL(ctx, playlist)
	return l.toRes(playlist), nil
}

// GetList 分页获取第三方歌单列表。
func (l *ThirdPartyPlaylistLogic) GetList(ctx context.Context, r *req.PlaylistListReq) (*res.PageRes[res.PlaylistRes], error) {
	playlists, total, err := l.playlistModel.GetList(ctx, r.GetPage(), r.GetPageSize())
	if err != nil {
		l.logger.Error("获取第三方歌单列表失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	// 批量解析封面媒体引用为当前完整 URL（外链行保持原值）
	l.resolveCoverURLs(ctx, playlists)

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
	// 解析封面媒体引用为当前完整 URL（外链行保持原值）
	l.resolveCoverURL(ctx, playlist)
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
	if r.CoverURL != nil || r.CoverMediaID != nil {
		mediaID, presetID := "", ""
		if r.CoverMediaID != nil {
			mediaID = *r.CoverMediaID
		}
		if r.CoverPresetID != nil {
			presetID = *r.CoverPresetID
		}
		coverURL := ""
		if r.CoverURL != nil {
			coverURL = *r.CoverURL
		}
		// 封面：媒体库引用（ID 索引）与外链互斥写入
		coverMediaID, coverPresetID, coverImage, err := normalizeMediaRef(ctx, mediaID, presetID, coverURL)
		if err != nil {
			return nil, err
		}
		playlist.CoverURL = coverImage
		playlist.CoverMediaID = mediaRefPtr(coverMediaID)
		playlist.CoverPresetID = mediaRefPtr(coverPresetID)
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

	// 解析封面媒体引用为当前完整 URL（外链行保持原值）
	l.resolveCoverURL(ctx, playlist)
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

	// 批量解析封面媒体引用为当前完整 URL（外链行保持原值）
	l.resolveCoverURLs(ctx, playlists)

	items := make([]res.PlaylistRes, len(playlists))
	for i, p := range playlists {
		items[i] = *l.toRes(&p)
	}
	return items, nil
}

// toRes 将模型转换为响应结构体。
func (l *ThirdPartyPlaylistLogic) toRes(p *model.ThirdPartyPlaylist) *res.PlaylistRes {
	return &res.PlaylistRes{
		ID:            p.ID,
		Title:         p.Title,
		CoverURL:      p.CoverURL,
		CoverMediaID:  refStr(p.CoverMediaID),
		CoverPresetID: refStr(p.CoverPresetID),
		Platform:      p.Platform,
		PlatformURL:   p.PlatformURL,
		PlatformLogo:  platformLogoURL(p.Platform),
		Description:   p.Description,
		SortOrder:     p.SortOrder,
		Enabled:       p.Enabled,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// resolveCoverURLs 批量解析第三方歌单封面媒体引用为当前完整 URL（外链行保持原值）。
func (l *ThirdPartyPlaylistLogic) resolveCoverURLs(ctx context.Context, playlists []model.ThirdPartyPlaylist) {
	ptrs := make([]*model.ThirdPartyPlaylist, len(playlists))
	for i := range playlists {
		ptrs[i] = &playlists[i]
	}
	model.ResolveCoverURLs(ctx, ptrs,
		func(p *model.ThirdPartyPlaylist) (string, string) { return refStr(p.CoverMediaID), refStr(p.CoverPresetID) },
		func(p *model.ThirdPartyPlaylist, u string) { p.CoverURL = u },
	)
}

// resolveCoverURL 解析单个第三方歌单封面媒体引用。
func (l *ThirdPartyPlaylistLogic) resolveCoverURL(ctx context.Context, p *model.ThirdPartyPlaylist) {
	model.ResolveCoverURLs(ctx, []*model.ThirdPartyPlaylist{p},
		func(x *model.ThirdPartyPlaylist) (string, string) { return refStr(x.CoverMediaID), refStr(x.CoverPresetID) },
		func(x *model.ThirdPartyPlaylist, u string) { x.CoverURL = u },
	)
}

// platformLogoURL 拼接平台 Logo 完整 URL；与歌曲封面 resolveURL 一致，BaseURL 未配置时返回相对路径。
func platformLogoURL(platform string) string {
	path := platformlogos.URLPath(platform)
	if path == "" {
		return ""
	}
	if model.BaseURL == "" {
		return path
	}
	return strings.TrimRight(model.BaseURL, "/") + path
}
