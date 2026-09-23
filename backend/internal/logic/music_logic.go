package logic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/internal/storage"
	"novablog/pkg/bilibili"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MusicLogic 歌曲业务逻辑结构体。
//
// 依赖：
//   - songModel：歌曲数据访问
//   - manager：存储管理器，用于自动保存封面到对象存储
//   - logger：业务日志
type MusicLogic struct {
	songModel *model.SongModel
	manager   *storage.Manager
	logger    *zap.Logger
}

// NewMusicLogic 创建 MusicLogic 实例。
func NewMusicLogic(manager *storage.Manager) *MusicLogic {
	return &MusicLogic{
		songModel: model.NewSong(),
		manager:   manager,
		logger:    MusicLogger,
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
		CoverURL:   l.saveCoverToStorage(ctx, r.CoverURL),
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
func (l *MusicLogic) UpdateSong(ctx context.Context, id string, r *req.UpdateSongReq) (*res.SongRes, error) {
	// 使用 GetByIDRaw 跳过 AfterFind 钩子，读取存储中的原始 cover_url，
	// 避免将钩子解析出的完整 URL 原样写回，覆盖相对路径存储值。
	song, err := l.songModel.GetByIDRaw(ctx, id)
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
		song.CoverURL = l.saveCoverToStorage(ctx, *r.CoverURL)
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

// DeleteSong 删除歌曲。
func (l *MusicLogic) DeleteSong(ctx context.Context, id string) error {
	_, err := l.songModel.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("歌曲不存在")
	}
	if err := l.songModel.Delete(ctx, id); err != nil {
		return err
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
			CoverURL:   l.saveCoverToStorage(ctx, songReq.CoverURL),
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

// GetAudioURL 获取歌曲的播放地址（管理端接口）。
//
// 返回 B 站官方外链播放器地址，前端用 iframe 内嵌播放，
// 不解析 CDN 直链（官方 CDN 存在 Referer 白名单防盗链，浏览器直连必 403）。
func (l *MusicLogic) GetAudioURL(ctx context.Context, songID string) (string, error) {
	song, err := l.songModel.GetByID(ctx, songID)
	if err != nil {
		return "", fmt.Errorf("歌曲不存在")
	}
	return bilibili.BuildEmbedURL(song.BVID), nil
}

// GetPublicSongList 获取公开歌曲列表（无需状态过滤）。
func (l *MusicLogic) GetPublicSongList(ctx context.Context, r *req.SongListReq) (*res.PageRes[res.SongRes], error) {
	return l.GetSongList(ctx, r)
}

// GetPublicSongByID 根据 ID 获取公开歌曲。
func (l *MusicLogic) GetPublicSongByID(ctx context.Context, id string) (*res.SongRes, error) {
	return l.GetSongByID(ctx, id)
}

// GetPublicAudioURL 获取歌曲的公开播放地址（B 站外链播放器地址）。
func (l *MusicLogic) GetPublicAudioURL(ctx context.Context, songID string) (string, error) {
	return l.GetAudioURL(ctx, songID)
}

// saveCoverToStorage 将外部封面 URL 下载并保存到已配置的对象存储中。
// 路径格式：music/cover/{uuid}.{ext}（前置 PathPrefix）。
// 若存储未配置或保存失败，降级使用原始 URL，仅记录警告。
// 对于已托管在当前存储中的 URL（本地 BaseURL 或云存储自定义域名，如媒体库选择），直接复用，不重复下载。
func (l *MusicLogic) saveCoverToStorage(ctx context.Context, coverURL string) string {
	if coverURL == "" {
		return coverURL
	}
	// 仅处理外部 HTTP(S) URL（Bilibili CDN 等），已保存的本地路径跳过
	if !strings.HasPrefix(coverURL, "http://") && !strings.HasPrefix(coverURL, "https://") {
		return coverURL
	}
	// 已属于本站/当前存储托管的地址（如媒体库选择），直接复用
	if isSelfHostedURL(coverURL, l.manager.GetActiveConfig()) {
		return coverURL
	}

	provider := l.manager.GetProviderOrReload(ctx)
	if provider == nil {
		l.logger.Warn("存储未配置，跳过封面保存", zap.String("cover_url", coverURL))
		return coverURL
	}

	// HTTP GET 下载封面（超时 30s）
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(coverURL)
	if err != nil {
		l.logger.Warn("下载封面失败", zap.String("cover_url", coverURL), zap.Error(err))
		return coverURL
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.logger.Warn("下载封面返回非 200", zap.String("cover_url", coverURL), zap.Int("status", resp.StatusCode))
		return coverURL
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		l.logger.Warn("读取封面数据失败", zap.Error(err))
		return coverURL
	}
	if len(data) == 0 {
		return coverURL
	}

	// 根据 Content-Type 确定扩展名
	contentType := resp.Header.Get("Content-Type")
	ext := extFromContentType(contentType)
	if ext == "" {
		ext = ".jpg"
	}

	// 生成对象 key：music/cover/{uuid}.{ext}
	storeFilename := uuid.New().String() + ext
	key := fmt.Sprintf("music/cover/%s", storeFilename)
	if activeCfg := l.manager.GetActiveConfig(); activeCfg != nil && activeCfg.PathPrefix != "" {
		key = fmt.Sprintf("%s/%s", strings.Trim(activeCfg.PathPrefix, "/"), key)
	}

	newURL, err := provider.Upload(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
	if err != nil {
		l.logger.Warn("上传封面到存储失败", zap.Error(err))
		return coverURL
	}

	l.logger.Info("封面已自动保存到存储",
		zap.String("source", coverURL),
		zap.String("saved", newURL),
	)
	return newURL
}

// extFromContentType 根据 Content-Type 返回文件扩展名。
func extFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		return ".jpg"
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	default:
		return ".jpg"
	}
}

// isSelfHostedURL 判断 URL 是否已托管在本地站点或当前存储上，避免重复下载落地。
// 命中以下任一情况即视为自托管：
//   - 以当前存储配置的 CustomDomain 为前缀；
//   - 以本地 BaseURL（/files 静态资源前缀）为前缀。
//
// 云存储默认桶域名（COS/OSS/MinIO）生成的 URL 无法在此精确匹配，但这类地址通常
// 是媒体库或历史封面数据，即便被再次下载也只是多存一份，不影响正确性。
func isSelfHostedURL(coverURL string, activeCfg *model.StorageConfig) bool {
	if activeCfg != nil && activeCfg.CustomDomain != "" &&
		strings.HasPrefix(coverURL, strings.TrimRight(activeCfg.CustomDomain, "/")+"/") {
		return true
	}
	if model.BaseURL != "" {
		base := strings.TrimRight(model.BaseURL, "/")
		if strings.HasPrefix(coverURL, base+"/files/") {
			return true
		}
	}
	return false
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
