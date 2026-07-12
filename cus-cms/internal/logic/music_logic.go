package logic

import (
	"context"
	"fmt"

	"cus-cms/internal/dto/req"
	"cus-cms/internal/dto/res"
	"cus-cms/internal/model"

	"github.com/google/uuid"
)

type MusicLogic struct {
	songModel *model.SongModel
}

func NewMusicLogic() *MusicLogic {
	return &MusicLogic{
		songModel: model.NewSong(),
	}
}

// CreateSong 创建歌曲
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

// GetSongList 分页获取歌曲列表
func (l *MusicLogic) GetSongList(ctx context.Context, r *req.SongListReq) (*res.PageRes[res.SongRes], error) {
	page := r.GetPage()
	pageSize := r.GetPageSize()

	var categoryID *string
	if r.CategoryID != "" {
		categoryID = &r.CategoryID
	}

	songs, total, err := l.songModel.GetList(ctx, categoryID, page, pageSize)
	if err != nil {
		return nil, err
	}

	var items []res.SongRes
	for i := range songs {
		items = append(items, *l.toSongRes(&songs[i]))
	}

	return res.NewPageRes(items, total, page, pageSize), nil
}

// GetSongByID 根据 ID 获取歌曲
func (l *MusicLogic) GetSongByID(ctx context.Context, id string) (*res.SongRes, error) {
	song, err := l.songModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("歌曲不存在")
	}
	return l.toSongRes(song), nil
}

// UpdateSong 更新歌曲
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

	return l.toSongRes(song), nil
}

// DeleteSong 删除歌曲
func (l *MusicLogic) DeleteSong(ctx context.Context, id string) error {
	_, err := l.songModel.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("歌曲不存在")
	}
	return l.songModel.Delete(ctx, id)
}

// StartParse 启动 B 站链接解析任务
func (l *MusicLogic) StartParse(ctx context.Context, r *req.ParseMusicReq) (string, error) {
	taskID, err := StartParseTask(r.URL)
	if err != nil {
		return "", err
	}
	return taskID, nil
}

// GetParseTask 获取解析任务状态
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

// BatchCreateSongs 批量创建歌曲
func (l *MusicLogic) BatchCreateSongs(ctx context.Context, r *req.BatchCreateSongReq) ([]res.SongRes, error) {
	var results []res.SongRes
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
	if results == nil {
		results = []res.SongRes{}
	}
	return results, nil
}

// GetAudioURL 获取歌曲的音频播放地址（B站CDN直链，不经过后端转发）
func (l *MusicLogic) GetAudioURL(ctx context.Context, songID string) (string, error) {
	song, err := l.songModel.GetByID(ctx, songID)
	if err != nil {
		return "", fmt.Errorf("歌曲不存在")
	}

	audioURL, err := FetchAudioURL(ctx, song.BVID, song.CID)
	if err != nil {
		return "", fmt.Errorf("获取音频地址失败: %w", err)
	}

	return audioURL, nil
}

// GetPublicSongList 获取公开歌曲列表（无需状态过滤）。
func (l *MusicLogic) GetPublicSongList(ctx context.Context, r *req.SongListReq) (*res.PageRes[res.SongRes], error) {
	return l.GetSongList(ctx, r)
}

// GetPublicSongByID 根据 ID 获取歌曲。
func (l *MusicLogic) GetPublicSongByID(ctx context.Context, id string) (*model.Song, error) {
	return l.songModel.GetByID(ctx, id)
}

// GetPublicAudioURL 获取歌曲的音频播放地址（复用已有 GetAudioURL 逻辑）。
func (l *MusicLogic) GetPublicAudioURL(ctx context.Context, songID string) (string, error) {
	return l.GetAudioURL(ctx, songID)
}

// toSongRes 转换为歌曲响应
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
