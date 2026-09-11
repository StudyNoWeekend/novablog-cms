package logic

import (
	"context"
	"fmt"
	"time"

	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/pkg/bilibili"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MusicLogic 歌曲业务逻辑结构体。
//
// 依赖：
//   - songModel：歌曲数据访问
//   - musicCache：音频 URL 缓存（B 站 CDN 链接 120 分钟过期，缓存 TTL 与 B 站 URL 中 expire 字段对齐）
//   - logger：业务日志
type MusicLogic struct {
	songModel  *model.SongModel
	musicCache *cache.MusicCache
	logger     *zap.Logger
}

// NewMusicLogic 创建 MusicLogic 实例。
func NewMusicLogic() *MusicLogic {
	return &MusicLogic{
		songModel:  model.NewSong(),
		musicCache: cache.NewMusicCache(),
		logger:     MusicLogger,
	}
}

// CreateSong 创建歌曲。
func (l *MusicLogic) CreateSong(ctx context.Context, r *req.CreateSongReq) (*res.SongRes, error) {
	sourceType := r.SourceType
	if sourceType == "" {
		sourceType = "bilibili"
	}

	song := &model.Song{
		ID:         uuid.New().String(),
		Title:      r.Title,
		Artist:     r.Artist,
		CoverURL:   r.CoverURL,
		BVID:       r.BVID,
		CID:        r.CID,
		SourceURL:  r.SourceURL,
		SourceType: sourceType,
		CategoryID: r.CategoryID,
		Duration:   r.Duration,
		SortOrder:  r.SortOrder,
	}

	if err := l.songModel.Create(ctx, song); err != nil {
		return nil, fmt.Errorf("创建歌曲失败: %w", err)
	}

	return l.toSongRes(song), nil
}

// GetSongList 分页获取歌曲列表（管理端）。
func (l *MusicLogic) GetSongList(ctx context.Context, r *req.SongListReq) (*res.PageRes[res.SongRes], error) {
	page := r.GetPage()
	pageSize := r.GetPageSize()

	var categoryID *string
	if r.CategoryID != "" {
		categoryID = &r.CategoryID
	}

	songs, total, err := l.songModel.GetList(ctx, categoryID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询歌曲列表失败: %w", err)
	}

	items := make([]res.SongRes, 0, len(songs))
	for i := range songs {
		items = append(items, *l.toSongRes(&songs[i]))
	}

	return res.NewPageRes(items, total, page, pageSize), nil
}

// GetSongByID 根据 ID 获取歌曲（管理端）。
func (l *MusicLogic) GetSongByID(ctx context.Context, id string) (*res.SongRes, error) {
	song, err := l.songModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("歌曲不存在")
	}
	return l.toSongRes(song), nil
}

// UpdateSong 更新歌曲。
//
// 副作用：更新成功后使对应的音频 URL 缓存失效，避免 B 站新链接与旧缓存不一致。
func (l *MusicLogic) UpdateSong(ctx context.Context, id string, r *req.UpdateSongReq) (*res.SongRes, error) {
	song, err := l.songModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("歌曲不存在")
	}

	if r.Title != nil {
		song.Title = *r.Title
	}
	if r.Artist != nil {
		song.Artist = *r.Artist
	}
	if r.CoverURL != nil {
		song.CoverURL = *r.CoverURL
	}
	if r.CategoryID != nil {
		song.CategoryID = r.CategoryID
	}
	if r.Duration != nil {
		song.Duration = *r.Duration
	}
	if r.SortOrder != nil {
		song.SortOrder = *r.SortOrder
	}

	if err := l.songModel.Update(ctx, song); err != nil {
		return nil, fmt.Errorf("更新歌曲失败: %w", err)
	}

	// 失效音频 URL 缓存（BVID/CID 可能已变更）。
	if err := l.musicCache.InvalidateAudioURL(ctx, id); err != nil {
		l.logger.Warn("失效音频 URL 缓存失败", zap.String("song_id", id), zap.Error(err))
	}

	return l.toSongRes(song), nil
}

// DeleteSong 删除歌曲。
//
// 副作用：删除成功后使对应的音频 URL 缓存失效。
func (l *MusicLogic) DeleteSong(ctx context.Context, id string) error {
	_, err := l.songModel.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("歌曲不存在")
	}
	if err := l.songModel.Delete(ctx, id); err != nil {
		return err
	}
	if err := l.musicCache.InvalidateAudioURL(ctx, id); err != nil {
		l.logger.Warn("失效音频 URL 缓存失败", zap.String("song_id", id), zap.Error(err))
	}
	return nil
}

// StartParse 启动 B 站链接解析任务。
func (l *MusicLogic) StartParse(ctx context.Context, r *req.ParseMusicReq) (string, error) {
	return StartParseTask(r.URL)
}

// GetParseTask 获取解析任务状态。
func (l *MusicLogic) GetParseTask(ctx context.Context, taskID string) (*res.ParseTaskRes, error) {
	task, ok := GetParseTask(taskID)
	if !ok {
		return nil, fmt.Errorf("解析任务不存在或已过期")
	}

	taskRes := &res.ParseTaskRes{
		TaskID: task.ID,
		Status: task.Status,
		Error:  task.Error,
	}

	if len(task.Results) > 0 {
		taskRes.Results = make([]res.ParseResultRes, len(task.Results))
		for i, r := range task.Results {
			taskRes.Results[i] = res.ParseResultRes{
				Title:    r.Title,
				Artist:   r.Artist,
				CoverURL: r.CoverURL,
				Duration: r.Duration,
				BVID:     r.BVID,
				CID:      r.CID,
			}
		}
	}

	return taskRes, nil
}

// BatchCreateSongs 批量创建歌曲。
func (l *MusicLogic) BatchCreateSongs(ctx context.Context, r *req.BatchCreateSongReq) ([]res.SongRes, error) {
	results := make([]res.SongRes, 0, len(r.Songs))
	for _, songReq := range r.Songs {
		sourceType := songReq.SourceType
		if sourceType == "" {
			sourceType = "bilibili"
		}
		song := &model.Song{
			ID:         uuid.New().String(),
			Title:      songReq.Title,
			Artist:     songReq.Artist,
			CoverURL:   songReq.CoverURL,
			BVID:       songReq.BVID,
			CID:        songReq.CID,
			SourceURL:  songReq.SourceURL,
			SourceType: sourceType,
			CategoryID: songReq.CategoryID,
			Duration:   songReq.Duration,
			SortOrder:  songReq.SortOrder,
		}
		if err := l.songModel.Create(ctx, song); err != nil {
			return nil, fmt.Errorf("批量创建歌曲失败: %w", err)
		}
		results = append(results, *l.toSongRes(song))
	}
	return results, nil
}

// GetAudioURL 获取歌曲的音频播放地址（管理端接口）。
//
// 缓存策略：
//  1. 先查 Redis（key=music:audio:<song_id>）；
//  2. 命中直接返回；未命中调用 B 站 /x/player/playurl 拉取，并按 URL 中 expire 字段设 TTL。
func (l *MusicLogic) GetAudioURL(ctx context.Context, songID string) (string, error) {
	return l.fetchAudioURL(ctx, songID)
}

// GetPublicSongList 获取公开歌曲列表（无需状态过滤）。
func (l *MusicLogic) GetPublicSongList(ctx context.Context, r *req.SongListReq) (*res.PageRes[res.SongRes], error) {
	return l.GetSongList(ctx, r)
}

// GetPublicSongByID 根据 ID 获取公开歌曲。
func (l *MusicLogic) GetPublicSongByID(ctx context.Context, id string) (*res.SongRes, error) {
	return l.GetSongByID(ctx, id)
}

// GetPublicAudioURL 获取歌曲的公开音频播放地址。
func (l *MusicLogic) GetPublicAudioURL(ctx context.Context, songID string) (string, error) {
	return l.fetchAudioURL(ctx, songID)
}

// fetchAudioURL 统一从缓存 / B 站获取音频 URL。
func (l *MusicLogic) fetchAudioURL(ctx context.Context, songID string) (string, error) {
	// 1) 缓存命中。
	if cached, err := l.musicCache.GetAudioURL(ctx, songID); err != nil {
		l.logger.Warn("读取音频 URL 缓存失败，将直连 B 站", zap.String("song_id", songID), zap.Error(err))
	} else if cached != nil {
		return cached.URL, nil
	}

	// 2) 查 song。
	song, err := l.songModel.GetByID(ctx, songID)
	if err != nil {
		return "", fmt.Errorf("歌曲不存在")
	}

	// 3) 调 B 站。
	start := time.Now()
	entries, err := bilibili.FetchAudioEntries(ctx, song.BVID, song.CID)
	if err != nil {
		return "", fmt.Errorf("获取音频地址失败: %w", err)
	}
	audioURL := pickBestAudioURL(entries)
	if audioURL == "" {
		return "", fmt.Errorf("未找到可用的音频流")
	}
	l.logger.Info("获取音频地址成功",
		zap.String("song_id", songID),
		zap.String("bvid", song.BVID),
		zap.Int64("cid", song.CID),
		zap.Duration("cost_ms", time.Since(start)),
	)

	// 4) 写缓存（TTL 与 B 站 URL 中 expire 对齐）。
	if err := l.musicCache.SetAudioURL(ctx, songID, audioURL); err != nil {
		l.logger.Warn("写入音频 URL 缓存失败", zap.String("song_id", songID), zap.Error(err))
	}
	return audioURL, nil
}

// pickBestAudioURL 选取 bandwidth 最大的音频流。
func pickBestAudioURL(entries []bilibili.AudioEntry) string {
	var bestURL string
	var bestBandwidth int
	for _, a := range entries {
		if a.Bandwidth > bestBandwidth {
			bestBandwidth = a.Bandwidth
			bestURL = a.BaseURL
		}
	}
	return bestURL
}

// toSongRes 转换为歌曲响应。
func (l *MusicLogic) toSongRes(s *model.Song) *res.SongRes {
	return &res.SongRes{
		ID:         s.ID,
		Title:      s.Title,
		Artist:     s.Artist,
		CoverURL:   s.CoverURL,
		BVID:       s.BVID,
		CID:        s.CID,
		SourceURL:  s.SourceURL,
		SourceType: s.SourceType,
		CategoryID: s.CategoryID,
		Duration:   s.Duration,
		SortOrder:  s.SortOrder,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}
