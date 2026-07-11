package logic

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"cus-cms/internal/dto/req"
	"cus-cms/internal/dto/res"
	"cus-cms/internal/model"
	"cus-cms/internal/storage"

	"github.com/google/uuid"
	"golang.org/x/image/webp"
)

// MediaLogic 媒体业务逻辑结构体。
type MediaLogic struct {
	model       *model.MediaModel
	presetModel *model.MediaPresetModel
	manager     *storage.Manager
}

// NewMediaLogic 创建 MediaLogic 实例。
func NewMediaLogic(manager *storage.Manager) *MediaLogic {
	return &MediaLogic{
		model:       model.NewMedia(),
		presetModel: model.NewMediaPreset(),
		manager:     manager,
	}
}

// UploadFile 上传文件到对象存储并记录到数据库。
func (l *MediaLogic) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader) (*res.MediaRes, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	provider := l.manager.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("对象存储未配置")
	}

	// 生成对象 key：images/2006/01/uuid.ext
	now := time.Now()
	dateDir := now.Format("2006/01")
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		ext = ".bin"
	}
	storeFilename := uuid.New().String() + ext
	key := fmt.Sprintf("images/%s/%s", dateDir, storeFilename)

	// 若配置了 PathPrefix 则拼接前缀
	if activeCfg := l.manager.GetActiveConfig(); activeCfg != nil && activeCfg.PathPrefix != "" {
		key = fmt.Sprintf("%s/%s", strings.Trim(activeCfg.PathPrefix, "/"), key)
	}

	fileType := getFileType(ext)
	mimeType := getMimeType(ext)

	url, err := provider.Upload(ctx, key, src, fileHeader.Size, mimeType)
	if err != nil {
		return nil, fmt.Errorf("上传文件到对象存储失败: %w", err)
	}

	media := &model.Media{
		ID:          uuid.New().String(),
		Filename:    fileHeader.Filename,
		FileType:    fileType,
		MimeType:    mimeType,
		Size:        fileHeader.Size,
		URL:         url,
		StoragePath: key,
		StorageType: provider.Type(),
	}

	if err := l.model.Create(ctx, media); err != nil {
		return nil, fmt.Errorf("保存媒体记录失败: %w", err)
	}

	return &res.MediaRes{
		ID:        media.ID,
		Filename:  media.Filename,
		FileType:  media.FileType,
		MimeType:  media.MimeType,
		Size:      media.Size,
		URL:       media.URL,
		CreatedAt: media.CreatedAt,
	}, nil
}

// GetList 获取媒体列表。
func (l *MediaLogic) GetList(ctx context.Context, req *req.MediaListReq) (*res.MediaListRes, error) {
	list, total, err := l.model.GetList(ctx, req.FileType, req.Keyword, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}

	var items []res.MediaRes
	for _, m := range list {
		items = append(items, res.MediaRes{
			ID:        m.ID,
			Filename:  m.Filename,
			FileType:  m.FileType,
			MimeType:  m.MimeType,
			Size:      m.Size,
			URL:       m.URL,
			ThumbURL:  ptrToString(m.ThumbURL),
			Width:     m.Width,
			Height:    m.Height,
			CreatedAt: m.CreatedAt,
		})
	}

	return &res.MediaListRes{
		List:       items,
		Total:      total,
		Page:       req.GetPage(),
		PageSize:   req.GetPageSize(),
		TotalPages: calcTotalPages(total, req.GetPageSize()),
	}, nil
}

// GetByID 获取媒体详情。
func (l *MediaLogic) GetByID(ctx context.Context, id string) (*res.MediaRes, error) {
	m, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &res.MediaRes{
		ID:        m.ID,
		Filename:  m.Filename,
		FileType:  m.FileType,
		MimeType:  m.MimeType,
		Size:      m.Size,
		URL:       m.URL,
		ThumbURL:  ptrToString(m.ThumbURL),
		Width:     m.Width,
		Height:    m.Height,
		CreatedAt: m.CreatedAt,
	}, nil
}

// Delete 软删除媒体。
func (l *MediaLogic) Delete(ctx context.Context, id string) error {
	return l.model.SoftDelete(ctx, id)
}

// CreatePreset 创建媒体预设：将前端合成的成品图转存对象存储并记录。
func (l *MediaLogic) CreatePreset(ctx context.Context, req *req.CreatePresetReq, fileHeader *multipart.FileHeader) (*res.MediaPresetRes, error) {
	// 校验原图是否存在
	media, err := l.model.GetByID(ctx, req.MediaID)
	if err != nil {
		return nil, fmt.Errorf("原图不存在: %w", err)
	}

	provider := l.manager.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("对象存储未配置")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 生成对象 key：presets/{media_id}/{uuid}.jpg
	storeFilename := uuid.New().String() + ".jpg"
	key := fmt.Sprintf("presets/%s/%s", req.MediaID, storeFilename)
	if activeCfg := l.manager.GetActiveConfig(); activeCfg != nil && activeCfg.PathPrefix != "" {
		key = fmt.Sprintf("%s/%s", strings.Trim(activeCfg.PathPrefix, "/"), key)
	}

	mimeType := "image/jpeg"
	url, err := provider.Upload(ctx, key, src, fileHeader.Size, mimeType)
	if err != nil {
		return nil, fmt.Errorf("上传预设成品图失败: %w", err)
	}

	preset := &model.MediaPreset{
		ID:                uuid.New().String(),
		MediaID:           media.ID,
		Name:              req.Name,
		FrameConfig:       req.FrameConfig,
		DisplayParams:     req.DisplayParams,
		OutputURL:         url,
		OutputStoragePath: key,
		OutputSize:        fileHeader.Size,
		MimeType:          mimeType,
	}
	if err := l.presetModel.Create(ctx, preset); err != nil {
		return nil, fmt.Errorf("保存预设记录失败: %w", err)
	}

	return &res.MediaPresetRes{
		ID:                preset.ID,
		MediaID:           preset.MediaID,
		Name:              preset.Name,
		FrameConfig:       preset.FrameConfig,
		DisplayParams:     preset.DisplayParams,
		OutputURL:         preset.OutputURL,
		OutputStoragePath: preset.OutputStoragePath,
		OutputSize:        preset.OutputSize,
		MimeType:          preset.MimeType,
		CreatedAt:         preset.CreatedAt,
	}, nil
}

// UploadWithPreset 上传原图并自动生成一个默认无 EXIF 的预设成品图。
func (l *MediaLogic) UploadWithPreset(ctx context.Context, req *req.UploadWithPresetReq, fileHeader *multipart.FileHeader) (*res.UploadWithPresetRes, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return nil, fmt.Errorf("不支持的文件格式，仅支持 .jpg/.jpeg/.png/.webp")
	}

	mediaRes, err := l.UploadFile(ctx, fileHeader)
	if err != nil {
		return nil, err
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	var img image.Image
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(src)
	case ".png":
		img, err = png.Decode(src)
	case ".webp":
		img, err = webp.Decode(src)
	}
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95}); err != nil {
		return nil, fmt.Errorf("编码 JPEG 失败: %w", err)
	}

	provider := l.manager.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("对象存储未配置")
	}

	storeFilename := uuid.New().String() + ".jpg"
	key := fmt.Sprintf("presets/%s/%s", mediaRes.ID, storeFilename)
	if activeCfg := l.manager.GetActiveConfig(); activeCfg != nil && activeCfg.PathPrefix != "" {
		key = fmt.Sprintf("%s/%s", strings.Trim(activeCfg.PathPrefix, "/"), key)
	}

	mimeType := "image/jpeg"
	url, err := provider.Upload(ctx, key, bytes.NewReader(buf.Bytes()), int64(buf.Len()), mimeType)
	if err != nil {
		return nil, fmt.Errorf("上传预设成品图失败: %w", err)
	}

	name := req.Name
	if name == "" {
		name = "默认无 EXIF"
	}
	frameConfig := `{"template":"gallery","showExif":false,"fontScale":1,"borderScale":1,"borderColor":"#ffffff","textColor":"auto","fontFamily":"system","logoMode":"text"}`
	displayParams := "{}"

	preset := &model.MediaPreset{
		ID:                uuid.New().String(),
		MediaID:           mediaRes.ID,
		Name:              name,
		FrameConfig:       frameConfig,
		DisplayParams:     displayParams,
		OutputURL:         url,
		OutputStoragePath: key,
		OutputSize:        int64(buf.Len()),
		MimeType:          mimeType,
	}
	if err := l.presetModel.Create(ctx, preset); err != nil {
		return nil, fmt.Errorf("保存预设记录失败: %w", err)
	}

	return &res.UploadWithPresetRes{
		Media: *mediaRes,
		Preset: res.MediaPresetRes{
			ID:                preset.ID,
			MediaID:           preset.MediaID,
			Name:              preset.Name,
			FrameConfig:       preset.FrameConfig,
			DisplayParams:     preset.DisplayParams,
			OutputURL:         preset.OutputURL,
			OutputStoragePath: preset.OutputStoragePath,
			OutputSize:        preset.OutputSize,
			MimeType:          preset.MimeType,
			CreatedAt:         preset.CreatedAt,
		},
	}, nil
}

// GetPresetsByMediaID 获取指定原图下的所有预设。
func (l *MediaLogic) GetPresetsByMediaID(ctx context.Context, mediaID string) (*res.MediaPresetListRes, error) {
	presets, err := l.presetModel.GetByMediaID(ctx, mediaID)
	if err != nil {
		return nil, err
	}

	items := make([]res.MediaPresetRes, 0, len(presets))
	for _, p := range presets {
		items = append(items, res.MediaPresetRes{
			ID:                p.ID,
			MediaID:           p.MediaID,
			Name:              p.Name,
			FrameConfig:       p.FrameConfig,
			DisplayParams:     p.DisplayParams,
			OutputURL:         p.OutputURL,
			OutputStoragePath: p.OutputStoragePath,
			OutputSize:        p.OutputSize,
			MimeType:          p.MimeType,
			CreatedAt:         p.CreatedAt,
		})
	}

	return &res.MediaPresetListRes{List: items}, nil
}

// DeletePreset 软删除媒体预设。
func (l *MediaLogic) DeletePreset(ctx context.Context, id string) error {
	return l.presetModel.SoftDelete(ctx, id)
}

// getFileType 获取文件类型：1图片 2视频 3音频 0其他。
func getFileType(ext string) int16 {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp":
		return 1
	case ".mp4", ".avi", ".mov", ".mkv", ".webm":
		return 2
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return 3
	default:
		return 0
	}
}

// getMimeType 根据扩展名返回 MIME 类型。
func getMimeType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}

// ptrToString 将 *string 安全转换为 string。
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// calcTotalPages 计算总页数。
func calcTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 1
	}
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return pages
}
