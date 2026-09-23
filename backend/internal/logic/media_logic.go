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
	"sort"
	"strings"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/internal/storage"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/image/webp"
)

// MediaLogger 媒体逻辑日志器，由 bootstrap 注入。
var MediaLogger *zap.Logger

// mediaLog 返回媒体逻辑日志器，未注入时返回 Nop。
func mediaLog() *zap.Logger {
	if MediaLogger != nil {
		return MediaLogger
	}
	return zap.NewNop()
}

// MediaLogic 媒体业务逻辑结构体。
type MediaLogic struct {
	model       *model.MediaModel
	presetModel *model.MediaPresetModel
	usageModel  *model.MediaUsageModel
	manager     *storage.Manager
}

// NewMediaLogic 创建 MediaLogic 实例。
func NewMediaLogic(manager *storage.Manager) *MediaLogic {
	return &MediaLogic{
		model:       model.NewMedia(),
		presetModel: model.NewMediaPreset(),
		usageModel:  model.NewMediaUsage(),
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

	provider := l.manager.GetProviderOrReload(ctx)
	if provider == nil {
		return nil, fmt.Errorf("对象存储未配置，请在存储配置页面创建并激活存储配置")
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

	provider := l.manager.GetProvider()

	var items []res.MediaRes
	for _, m := range list {
		thumbURL := ptrToString(m.ThumbURL)
		if thumbURL == "" && provider != nil {
			thumbURL = provider.GetThumbURL(m.URL, 300)
		}
		items = append(items, res.MediaRes{
			ID:        m.ID,
			Filename:  m.Filename,
			FileType:  m.FileType,
			MimeType:  m.MimeType,
			Size:      m.Size,
			URL:       m.URL,
			ThumbURL:  thumbURL,
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

// GetUsages 扫描媒体被哪些内容模块引用（文章、旅行攻略、摄影作品集等）。
func (l *MediaLogic) GetUsages(ctx context.Context, id string) (*res.MediaUsageRes, error) {
	media, presets, err := l.getMediaWithPresets(ctx, id)
	if err != nil {
		return nil, err
	}

	hits, err := l.usageModel.FindUsages(ctx, media.URL, media.StoragePath, presetIDs(presets), presetOutputKeys(presets))
	if err != nil {
		return nil, err
	}

	return &res.MediaUsageRes{
		MediaID: id,
		Used:    len(hits) > 0,
		Total:   len(hits),
		Groups:  groupUsageHits(hits),
	}, nil
}

// Delete 删除媒体：先扫描引用，被引用时需 force=true 强制删除；确认后
// 硬删数据库记录（含预设）并尽力删除存储中的原文件与预设成品图。
func (l *MediaLogic) Delete(ctx context.Context, id string, force bool) error {
	media, presets, err := l.getMediaWithPresets(ctx, id)
	if err != nil {
		return err
	}

	hits, err := l.usageModel.FindUsages(ctx, media.URL, media.StoragePath, presetIDs(presets), presetOutputKeys(presets))
	if err != nil {
		return err
	}
	if len(hits) > 0 && !force {
		return fmt.Errorf("该文件被 %d 处内容引用，无法直接删除", len(hits))
	}

	if err := l.model.HardDeleteWithPresets(ctx, id); err != nil {
		return fmt.Errorf("删除媒体记录失败: %w", err)
	}

	l.deleteStorageFiles(ctx, media, presets)
	return nil
}

// getMediaWithPresets 获取媒体原始记录（未经 URL 钩子解析）及其全部预设（含软删）。
func (l *MediaLogic) getMediaWithPresets(ctx context.Context, id string) (*model.Media, []model.MediaPreset, error) {
	media, err := l.model.GetByIDRaw(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("媒体不存在: %w", err)
	}
	presets, err := l.presetModel.GetAllByMediaIDUnscoped(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("查询媒体预设失败: %w", err)
	}
	return media, presets, nil
}

// deleteStorageFiles 尽力删除存储中的原文件与预设成品图，失败仅记录日志不回滚。
func (l *MediaLogic) deleteStorageFiles(ctx context.Context, media *model.Media, presets []model.MediaPreset) {
	provider := l.manager.GetProvider()
	if provider == nil {
		mediaLog().Warn("删除媒体文件跳过：无可用存储 Provider", zap.String("media_id", media.ID))
		return
	}

	keys := make([]string, 0, 1+len(presets))
	if media.StoragePath != "" {
		keys = append(keys, media.StoragePath)
	}
	for _, p := range presets {
		if p.OutputStoragePath != "" {
			keys = append(keys, p.OutputStoragePath)
		}
	}

	for _, key := range keys {
		if err := provider.Delete(ctx, key); err != nil {
			mediaLog().Warn("删除存储文件失败",
				zap.String("media_id", media.ID),
				zap.String("key", key),
				zap.Error(err),
			)
		}
	}
}

// presetIDs 提取预设 ID 列表。
func presetIDs(presets []model.MediaPreset) []string {
	ids := make([]string, 0, len(presets))
	for _, p := range presets {
		ids = append(ids, p.ID)
	}
	return ids
}

// presetOutputKeys 提取预设成品图的存储路径列表。
func presetOutputKeys(presets []model.MediaPreset) []string {
	keys := make([]string, 0, len(presets))
	for _, p := range presets {
		keys = append(keys, p.OutputStoragePath)
	}
	return keys
}

// usageModuleOrder 引用分组在前端展示时的固定模块顺序。
var usageModuleOrder = []string{
	model.UsageModuleArticle,
	model.UsageModuleTravel,
	model.UsageModulePortfolio,
	model.UsageModuleProject,
	model.UsageModuleEquipment,
	model.UsageModuleVideo,
	model.UsageModuleSong,
	model.UsageModuleBlogger,
}

// groupUsageHits 将引用命中按模块聚合，按固定模块顺序输出，未知模块排最后。
func groupUsageHits(hits []model.UsageHit) []res.MediaUsageGroup {
	order := make(map[string]int, len(usageModuleOrder))
	for i, m := range usageModuleOrder {
		order[m] = i
	}

	groups := make([]res.MediaUsageGroup, 0)
	index := make(map[string]int)
	for _, h := range hits {
		i, ok := index[h.Module]
		if !ok {
			groups = append(groups, res.MediaUsageGroup{Module: h.Module, Items: []res.MediaUsageItem{}})
			i = len(groups) - 1
			index[h.Module] = i
		}
		groups[i].Items = append(groups[i].Items, res.MediaUsageItem{ID: h.ID, Title: h.Title, Field: h.Field})
	}

	sort.SliceStable(groups, func(a, b int) bool {
		return moduleOrder(order, groups[a].Module) < moduleOrder(order, groups[b].Module)
	})
	return groups
}

// moduleOrder 返回模块的展示顺序，未知模块排最后。
func moduleOrder(order map[string]int, module string) int {
	if i, ok := order[module]; ok {
		return i
	}
	return len(usageModuleOrder)
}

// CreatePreset 创建媒体预设：将前端合成的成品图转存对象存储并记录。
func (l *MediaLogic) CreatePreset(ctx context.Context, req *req.CreatePresetReq, fileHeader *multipart.FileHeader) (*res.MediaPresetRes, error) {
	// 校验原图是否存在
	media, err := l.model.GetByID(ctx, req.MediaID)
	if err != nil {
		return nil, fmt.Errorf("原图不存在: %w", err)
	}

	provider := l.manager.GetProviderOrReload(ctx)
	if provider == nil {
		return nil, fmt.Errorf("对象存储未配置，请在存储配置页面创建并激活存储配置")
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

	provider := l.manager.GetProviderOrReload(ctx)
	if provider == nil {
		return nil, fmt.Errorf("对象存储未配置，请在存储配置页面创建并激活存储配置")
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
