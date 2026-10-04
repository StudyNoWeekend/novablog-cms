package logic

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/internal/storage"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/image/webp"
	"gorm.io/gorm"
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
	folderModel *model.MediaFolderModel
	manager     *storage.Manager
	uploadDir   string // 本地存储目录（未配置对象存储时降级使用）
}

// MediaUploadOptions 媒体上传归属选项：决定文件归属的文件夹与物理存储路径前缀。
type MediaUploadOptions struct {
	Module   string // 业务模块 key（enum.MediaModuleFolders 的键），归属到对应模块文件夹
	FolderID string // 显式指定目标文件夹 ID，优先于 Module
}

// NewMediaLogic 创建 MediaLogic 实例。
func NewMediaLogic(manager *storage.Manager, uploadDir string) *MediaLogic {
	return &MediaLogic{
		model:       model.NewMedia(),
		presetModel: model.NewMediaPreset(),
		usageModel:  model.NewMediaUsage(),
		folderModel: model.NewMediaFolder(),
		manager:     manager,
		uploadDir:   uploadDir,
	}
}

// activeProvider 获取当前对象存储 Provider；未配置时降级为本地存储，保证上传能力开箱可用。
func (l *MediaLogic) activeProvider(ctx context.Context) (storage.StorageProvider, error) {
	if p := l.manager.GetProviderOrReload(ctx); p != nil {
		return p, nil
	}
	fallback, err := storage.NewLocalProvider(l.uploadDir, "", "")
	if err != nil {
		return nil, fmt.Errorf("初始化本地存储失败: %w", err)
	}
	mediaLog().Warn("未配置对象存储，本次上传降级为本地存储", zap.String("upload_dir", l.uploadDir))
	return fallback, nil
}

// UploadFile 上传文件到对象存储并记录到数据库。归属优先级：opts.FolderID > opts.Module > 根目录。
func (l *MediaLogic) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, opts MediaUploadOptions) (*res.MediaRes, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	folder, err := l.resolveTargetFolder(ctx, opts)
	if err != nil {
		return nil, err
	}

	provider, perr := l.activeProvider(ctx)
	if perr != nil {
		return nil, perr
	}

	// 生成对象 key：{中文文件夹链}/2006/01/uuid.ext（根目录兜底 未分类/）
	now := time.Now()
	dateDir := now.Format("2006/01")
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		ext = ".bin"
	}
	storeFilename := uuid.New().String() + ext
	prefix, err := l.storagePrefix(ctx, folder)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s/%s/%s", prefix, dateDir, storeFilename)

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
		FolderID:    folderIDPtr(folder),
	}

	if err := l.model.Create(ctx, media); err != nil {
		return nil, fmt.Errorf("保存媒体记录失败: %w", err)
	}

	return l.toMediaRes(ctx, media)
}

// SaveRemoteBytes 将外部图片字节写入当前存储并登记媒体记录（供图片搜索转存等使用）。
func (l *MediaLogic) SaveRemoteBytes(ctx context.Context, filename, contentType string, data []byte, opts MediaUploadOptions) (*res.MediaRes, error) {
	folder, err := l.resolveTargetFolder(ctx, opts)
	if err != nil {
		return nil, err
	}

	provider, err := l.activeProvider(ctx)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".bin"
	}
	storeFilename := uuid.New().String() + ext
	prefix, err := l.storagePrefix(ctx, folder)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s/%s/%s", prefix, time.Now().Format("2006/01"), storeFilename)
	if activeCfg := l.manager.GetActiveConfig(); activeCfg != nil && activeCfg.PathPrefix != "" {
		key = fmt.Sprintf("%s/%s", strings.Trim(activeCfg.PathPrefix, "/"), key)
	}

	url, err := provider.Upload(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
	if err != nil {
		return nil, fmt.Errorf("转存图片失败: %w", err)
	}

	media := &model.Media{
		ID:          uuid.New().String(),
		Filename:    filename,
		FileType:    getFileType(ext),
		MimeType:    getMimeType(ext),
		Size:        int64(len(data)),
		URL:         url,
		StoragePath: key,
		StorageType: provider.Type(),
		FolderID:    folderIDPtr(folder),
	}
	if err := l.model.Create(ctx, media); err != nil {
		return nil, fmt.Errorf("保存媒体记录失败: %w", err)
	}

	return l.toMediaRes(ctx, media)
}

// resolveTargetFolder 解析上传归属文件夹：FolderID 优先，其次按模块 key 归入模块文件夹，否则根目录。
func (l *MediaLogic) resolveTargetFolder(ctx context.Context, opts MediaUploadOptions) (*model.MediaFolder, error) {
	if opts.FolderID != "" {
		folder, err := l.folderModel.GetByID(ctx, opts.FolderID)
		if err != nil {
			return nil, fmt.Errorf("目标文件夹不存在: %w", err)
		}
		return folder, nil
	}
	if opts.Module != "" {
		return l.ensureModuleFolder(ctx, opts.Module)
	}
	return nil, nil
}

// ensureModuleFolder 获取模块对应的顶级文件夹，不存在时自动创建；未知模块 key 返回根目录。
func (l *MediaLogic) ensureModuleFolder(ctx context.Context, moduleKey string) (*model.MediaFolder, error) {
	name, ok := enum.MediaModuleFolders[moduleKey]
	if !ok {
		mediaLog().Warn("上传携带未知模块 key，按根目录处理", zap.String("module", moduleKey))
		return nil, nil
	}
	folder, err := l.folderModel.GetTopByModuleKey(ctx, moduleKey)
	if err == nil {
		return folder, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询模块文件夹失败: %w", err)
	}
	// 兜底匹配同名顶级文件夹（兼容历史上仅有同名文件夹、无 module_key 的情况）
	folder, err = l.folderModel.GetTopByName(ctx, name)
	if err == nil {
		return folder, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询模块文件夹失败: %w", err)
	}
	folder = &model.MediaFolder{ID: uuid.New().String(), Name: name, ModuleKey: &moduleKey}
	if err := l.folderModel.Create(ctx, folder); err != nil {
		return nil, fmt.Errorf("创建模块文件夹失败: %w", err)
	}
	return folder, nil
}

// rootStorageDir 根目录上传时物理存储路径的一级目录名。
const rootStorageDir = "未分类"

// storagePrefix 计算物理存储路径前缀：文件夹祖先链名称；根目录返回 未分类。
func (l *MediaLogic) storagePrefix(ctx context.Context, folder *model.MediaFolder) (string, error) {
	if folder == nil {
		return rootStorageDir, nil
	}
	names, err := l.folderModel.AncestorNames(ctx, folder)
	if err != nil {
		return "", fmt.Errorf("计算存储路径失败: %w", err)
	}
	return strings.Join(names, "/"), nil
}

// folderIDPtr 提取文件夹 ID 指针（根目录为 nil）。
func folderIDPtr(folder *model.MediaFolder) *string {
	if folder == nil {
		return nil
	}
	return &folder.ID
}

// toMediaRes 构建媒体响应（含所属文件夹信息）。
func (l *MediaLogic) toMediaRes(ctx context.Context, media *model.Media) (*res.MediaRes, error) {
	folderName := ""
	if media.FolderID != nil {
		folder, err := l.folderModel.GetByID(ctx, *media.FolderID)
		if err != nil {
			return nil, fmt.Errorf("查询所属文件夹失败: %w", err)
		}
		folderName = folder.Name
	}
	return &res.MediaRes{
		ID:         media.ID,
		Filename:   media.Filename,
		FileType:   media.FileType,
		MimeType:   media.MimeType,
		Size:       media.Size,
		URL:        model.ResolveURL(media.URL),
		Width:      media.Width,
		Height:     media.Height,
		FolderID:   ptrToString(media.FolderID),
		FolderName: folderName,
		CreatedAt:  media.CreatedAt,
	}, nil
}

// GetList 获取媒体列表。FolderID 取值：nil=全部；"root"=根目录；其他=指定文件夹。
func (l *MediaLogic) GetList(ctx context.Context, req *req.MediaListReq) (*res.MediaListRes, error) {
	var folderID *string
	rootOnly := false
	if req.FolderID != nil && *req.FolderID != "" {
		if *req.FolderID == "root" {
			rootOnly = true
		} else {
			folderID = req.FolderID
		}
	}

	list, total, err := l.model.GetList(ctx, req.FileType, req.Keyword, folderID, rootOnly, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}

	provider := l.manager.GetProvider()

	// 批量查询所属文件夹名
	folderIDs := make([]string, 0, len(list))
	for _, m := range list {
		if m.FolderID != nil {
			folderIDs = append(folderIDs, *m.FolderID)
		}
	}
	folderNames := make(map[string]string)
	if len(folderIDs) > 0 {
		folders, err := l.folderModel.GetByIDs(ctx, folderIDs)
		if err != nil {
			return nil, err
		}
		for _, f := range folders {
			folderNames[f.ID] = f.Name
		}
	}

	items := make([]res.MediaRes, 0, len(list))
	for _, m := range list {
		thumbURL := ptrToString(m.ThumbURL)
		if thumbURL == "" && provider != nil {
			thumbURL = provider.GetThumbURL(m.URL, 300)
		}
		folderName := ""
		if m.FolderID != nil {
			folderName = folderNames[*m.FolderID]
		}
		items = append(items, res.MediaRes{
			ID:         m.ID,
			Filename:   m.Filename,
			FileType:   m.FileType,
			MimeType:   m.MimeType,
			Size:       m.Size,
			URL:        m.URL,
			ThumbURL:   thumbURL,
			Width:      m.Width,
			Height:     m.Height,
			FolderID:   ptrToString(m.FolderID),
			FolderName: folderName,
			CreatedAt:  m.CreatedAt,
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
	return l.toMediaRes(ctx, m)
}

// MoveMedia 批量移动媒体到指定文件夹（folderID 为空表示移回根目录）。
func (l *MediaLogic) MoveMedia(ctx context.Context, mediaIDs []string, folderID *string) error {
	if folderID != nil && *folderID != "" {
		if _, err := l.folderModel.GetByID(ctx, *folderID); err != nil {
			return fmt.Errorf("目标文件夹不存在: %w", err)
		}
	}
	if err := l.model.UpdateFolderIDs(ctx, mediaIDs, folderID); err != nil {
		return fmt.Errorf("移动媒体失败: %w", err)
	}
	return nil
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
	model.UsageModuleRecipe,
	model.UsageModuleBook,
	model.UsageModuleGame,
	model.UsageModuleTechStack,
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

	provider, perr := l.activeProvider(ctx)
	if perr != nil {
		return nil, perr
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

	mediaRes, err := l.UploadFile(ctx, fileHeader, MediaUploadOptions{Module: req.Module, FolderID: req.FolderID})
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

	provider, perr := l.activeProvider(ctx)
	if perr != nil {
		return nil, perr
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
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".ico":
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
	case ".ico":
		return "image/x-icon"
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
