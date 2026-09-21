package logic

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/internal/storage"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MigrationLogic 素材迁移业务逻辑。
type MigrationLogic struct {
	mediaModel     *model.MediaModel
	configModel    *model.StorageConfigModel
	migrationModel *model.StorageMigrationModel
	presetModel    *model.MediaPresetModel
	manager        *storage.Manager
	cryptoKey      string
	uploadDir      string // 本地迁移源目录
	logger         *zap.Logger
}

// NewMigrationLogic 创建 MigrationLogic 实例。
func NewMigrationLogic(manager *storage.Manager, cryptoKey, uploadDir string, logger *zap.Logger) *MigrationLogic {
	return &MigrationLogic{
		mediaModel:     model.NewMedia(),
		configModel:    model.NewStorageConfig(),
		migrationModel: model.NewStorageMigration(),
		presetModel:    model.NewMediaPreset(),
		manager:        manager,
		cryptoKey:      cryptoKey,
		uploadDir:      uploadDir,
		logger:         logger,
	}
}

// Analyze 异步分析目标平台中缺失的素材。
// 目标平台必须存在配置（不要求是当前活跃平台）。
func (l *MigrationLogic) Analyze(ctx context.Context, targetProvider string) (string, error) {
	// 校验目标平台配置存在（不要求是当前活跃平台）
	targetConfig, err := l.configModel.GetByProvider(ctx, targetProvider)
	if err != nil {
		return "", fmt.Errorf("目标平台配置不存在，请先添加配置")
	}
	// 验证配置能构建出可用的 Provider
	targetProviderInstance, err := storage.NewProvider(targetConfig, l.cryptoKey, l.uploadDir)
	if err != nil {
		return "", fmt.Errorf("目标平台配置无效: %w", err)
	}
	_ = targetProviderInstance // used in runAnalyze

	// 创建分析任务
	now := time.Now()
	task := &model.StorageMigrationTask{
		ID:             uuid.New().String(),
		TaskType:       "analyze",
		TargetProvider: targetProvider,
		Status:         "running",
		StartedAt:      &now,
	}
	if err := l.migrationModel.CreateTask(ctx, task); err != nil {
		return "", fmt.Errorf("创建分析任务失败: %w", err)
	}

	taskID := task.ID

	// 异步执行分析
	go l.runAnalyze(taskID, targetProvider)

	return taskID, nil
}

// runAnalyze 异步执行分析逻辑。
func (l *MigrationLogic) runAnalyze(taskID, targetProvider string) {
	// TODO: use cancellable context to support graceful shutdown
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			l.logger.Error("分析任务 panic", zap.String("task_id", taskID), zap.Any("panic", r))
			l.updateTaskFailed(ctx, taskID, fmt.Sprintf("分析任务 panic: %v", r))
		}
	}()

	// Build target provider from config (not necessarily active)
	targetConfig, err := l.configModel.GetByProvider(ctx, targetProvider)
	if err != nil {
		l.updateTaskFailed(ctx, taskID, "目标平台配置不存在")
		return
	}
	provider, err := storage.NewProvider(targetConfig, l.cryptoKey, l.uploadDir)
	if err != nil {
		l.updateTaskFailed(ctx, taskID, fmt.Sprintf("构建目标 Provider 失败: %v", err))
		return
	}

	var allItems []model.StorageMigrationItem
	var total int64
	page := 1
	pageSize := 100

	for {
		list, t, err := l.mediaModel.GetList(ctx, nil, nil, page, pageSize)
		if err != nil {
			l.updateTaskFailed(ctx, taskID, fmt.Sprintf("查询媒体列表失败: %v", err))
			return
		}
		total = t

		for _, media := range list {
			status := "exist"
			// 仅当素材不在目标平台时才检查存在性
			if media.StorageType != targetProvider {
				exists, err := provider.Exists(ctx, media.StoragePath)
				if err != nil {
					l.logger.Error("检查文件存在性失败",
						zap.String("media_id", media.ID),
						zap.String("storage_path", media.StoragePath),
						zap.Error(err))
					status = "missing"
				} else if !exists {
					status = "missing"
				}
			}

			allItems = append(allItems, model.StorageMigrationItem{
				ID:         uuid.New().String(),
				TaskID:     taskID,
				MediaID:    media.ID,
				SourceType: "media",
				Status:     status,
			})
		}

		if page*pageSize >= int(total) {
			break
		}
		page++
	}

	// 扫描 media_presets
	presetPage := 1
	for {
		presets, presetTotal, err := l.presetModel.GetAll(ctx, presetPage, pageSize)
		if err != nil {
			l.updateTaskFailed(ctx, taskID, fmt.Sprintf("查询预设列表失败: %v", err))
			return
		}
		for _, preset := range presets {
			status := "exist"
			// 检查预设的输出文件是否在目标平台存在
			exists, err := provider.Exists(ctx, preset.OutputStoragePath)
			if err != nil {
				l.logger.Error("检查预设文件存在性失败",
					zap.String("preset_id", preset.ID),
					zap.String("output_storage_path", preset.OutputStoragePath),
					zap.Error(err))
				status = "missing"
			} else if !exists {
				status = "missing"
			}
			allItems = append(allItems, model.StorageMigrationItem{
				ID:         uuid.New().String(),
				TaskID:     taskID,
				MediaID:    preset.ID, // 用 preset ID
				SourceType: "preset",
				Status:     status,
			})
		}
		if presetPage*pageSize >= int(presetTotal) {
			break
		}
		presetPage++
	}

	// 批量创建 items（每批 100 条）
	batchSize := 100
	for i := 0; i < len(allItems); i += batchSize {
		end := i + batchSize
		if end > len(allItems) {
			end = len(allItems)
		}
		if err := l.migrationModel.CreateItems(ctx, allItems[i:end]); err != nil {
			l.updateTaskFailed(ctx, taskID, fmt.Sprintf("创建分析条目失败: %v", err))
			return
		}
	}

	// 更新任务状态为完成
	task, err := l.migrationModel.GetTaskByID(ctx, taskID)
	if err != nil {
		l.logger.Error("获取分析任务失败", zap.Error(err))
		return
	}
	task.Total = len(allItems)
	task.Status = "completed"
	finishedAt := time.Now()
	task.FinishedAt = &finishedAt
	if err := l.migrationModel.UpdateTask(ctx, task); err != nil {
		l.logger.Error("更新分析任务状态失败", zap.Error(err))
	}
}

// GetAnalyzeResult 获取分析结果。
func (l *MigrationLogic) GetAnalyzeResult(ctx context.Context, taskID string) (*res.AnalyzeResultRes, error) {
	task, err := l.migrationModel.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询迁移任务失败: %w", err)
	}

	items, err := l.migrationModel.GetItemsByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询迁移条目失败: %w", err)
	}

	// 批量查询 media 和 preset 信息
	var mediaIDs, presetIDs []string
	for _, item := range items {
		if item.SourceType == "preset" {
			presetIDs = append(presetIDs, item.MediaID)
		} else {
			mediaIDs = append(mediaIDs, item.MediaID)
		}
	}

	// 查询 media
	mediaMap := make(map[string]model.Media)
	if len(mediaIDs) > 0 {
		mediaList, err := l.mediaModel.GetByIDs(ctx, mediaIDs)
		if err != nil {
			l.logger.Error("查询媒体信息失败", zap.Error(err))
		} else {
			for _, m := range mediaList {
				mediaMap[m.ID] = m
			}
		}
	}

	// 查询 presets - batch query by IDs
	presetMap := make(map[string]model.MediaPreset)
	if len(presetIDs) > 0 {
		presets, err := l.presetModel.GetByIDs(ctx, presetIDs)
		if err != nil {
			l.logger.Error("查询预设信息失败", zap.Error(err))
		} else {
			for _, p := range presets {
				presetMap[p.ID] = p
			}
		}
	}

	// Also query media for presets' media_ids (to get storage_type)
	var presetMediaIDs []string
	for _, p := range presetMap {
		presetMediaIDs = append(presetMediaIDs, p.MediaID)
	}
	if len(presetMediaIDs) > 0 {
		extraMedia, err := l.mediaModel.GetByIDs(ctx, presetMediaIDs)
		if err == nil {
			for _, m := range extraMedia {
				mediaMap[m.ID] = m
			}
		}
	}

	result := &res.AnalyzeResultRes{
		Task:     l.toTaskRes(task),
		Missing:  []res.MigrationItemRes{},
		Existing: []res.MigrationItemRes{},
	}
	// Build result
	for _, item := range items {
		itemRes := res.MigrationItemRes{
			ID:         item.ID,
			MediaID:    item.MediaID,
			Status:     item.Status,
			Error:      item.Error,
			SourceType: item.SourceType,
		}
		if item.SourceType == "preset" {
			if preset, ok := presetMap[item.MediaID]; ok {
				itemRes.Filename = preset.Name
				itemRes.URL = preset.OutputURL
				// Get storage type from associated media
				if media, ok := mediaMap[preset.MediaID]; ok {
					itemRes.StorageType = media.StorageType
				}
			}
		} else {
			if media, ok := mediaMap[item.MediaID]; ok {
				itemRes.Filename = media.Filename
				itemRes.StorageType = media.StorageType
				itemRes.URL = media.URL
			}
		}
		if item.Status == "missing" {
			result.Missing = append(result.Missing, itemRes)
		} else {
			result.Existing = append(result.Existing, itemRes)
		}
	}
	return result, nil
}

// StartMigration 异步执行素材迁移。
// 目标平台必须存在配置（不要求是当前活跃平台）。
func (l *MigrationLogic) StartMigration(ctx context.Context, r *req.MigrationStartReq) (string, error) {
	// 校验目标平台配置存在（不要求是当前活跃平台）
	targetConfig, err := l.configModel.GetByProvider(ctx, r.TargetProvider)
	if err != nil {
		return "", fmt.Errorf("目标平台配置不存在，请先添加配置")
	}
	targetProviderInstance, err := storage.NewProvider(targetConfig, l.cryptoKey, l.uploadDir)
	if err != nil {
		return "", fmt.Errorf("目标平台配置无效: %w", err)
	}
	_ = targetProviderInstance // used in runMigration

	// 校验参数：必须指定 MediaIDs 或 All
	if !r.All && len(r.MediaIDs) == 0 {
		return "", fmt.Errorf("请指定要迁移的媒体 ID 或选择全部迁移")
	}

	// 创建迁移任务
	now := time.Now()
	task := &model.StorageMigrationTask{
		ID:             uuid.New().String(),
		TaskType:       "migrate",
		TargetProvider: r.TargetProvider,
		Status:         "running",
		StartedAt:      &now,
	}
	if err := l.migrationModel.CreateTask(ctx, task); err != nil {
		return "", fmt.Errorf("创建迁移任务失败: %w", err)
	}

	// 确定要迁移的 media 列表
	var mediaList []model.Media
	if r.All {
		// 分页遍历所有 media
		page := 1
		pageSize := 100
		for {
			list, total, err := l.mediaModel.GetList(ctx, nil, nil, page, pageSize)
			if err != nil {
				l.failTask(ctx, task, fmt.Sprintf("查询媒体列表失败: %v", err))
				return "", fmt.Errorf("查询媒体列表失败: %w", err)
			}
			mediaList = append(mediaList, list...)
			if page*pageSize >= int(total) {
				break
			}
			page++
		}
	} else {
		var err error
		mediaList, err = l.mediaModel.GetByIDs(ctx, r.MediaIDs)
		if err != nil {
			l.failTask(ctx, task, fmt.Sprintf("查询媒体失败: %v", err))
			return "", fmt.Errorf("查询媒体失败: %w", err)
		}
	}

	// 过滤出 storage_type != targetProvider 的
	var toMigrate []model.Media
	for _, m := range mediaList {
		if m.StorageType != r.TargetProvider {
			toMigrate = append(toMigrate, m)
		}
	}

	// 创建 items
	items := make([]model.StorageMigrationItem, 0, len(toMigrate))
	for _, m := range toMigrate {
		items = append(items, model.StorageMigrationItem{
			ID:      uuid.New().String(),
			TaskID:  task.ID,
			MediaID: m.ID,
			Status:  "pending",
		})
	}

	// 批量创建 items
	batchSize := 100
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		if err := l.migrationModel.CreateItems(ctx, items[i:end]); err != nil {
			l.failTask(ctx, task, fmt.Sprintf("创建迁移条目失败: %v", err))
			return "", fmt.Errorf("创建迁移条目失败: %w", err)
		}
	}

	// 更新任务总数
	task.Total = len(items)
	if err := l.migrationModel.UpdateTask(ctx, task); err != nil {
		l.logger.Error("更新任务总数失败", zap.Error(err))
	}

	taskID := task.ID

	// 异步执行迁移
	go l.runMigration(taskID, r.TargetProvider)

	return taskID, nil
}

// runMigration 异步执行迁移逻辑。
func (l *MigrationLogic) runMigration(taskID, targetProvider string) {
	// TODO: use cancellable context to support graceful shutdown
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			l.logger.Error("迁移任务 panic", zap.String("task_id", taskID), zap.Any("panic", r))
			l.updateTaskFailed(ctx, taskID, fmt.Sprintf("迁移任务 panic: %v", r))
		}
	}()

	// 获取所有待迁移 items
	items, err := l.migrationModel.GetItemsByTaskID(ctx, taskID)
	if err != nil {
		l.updateTaskFailed(ctx, taskID, fmt.Sprintf("查询迁移条目失败: %v", err))
		return
	}

	// Build target provider from config (not necessarily active)
	targetConfig, err := l.configModel.GetByProvider(ctx, targetProvider)
	if err != nil {
		l.updateTaskFailed(ctx, taskID, "目标平台配置不存在")
		return
	}
	targetProvider_, err := storage.NewProvider(targetConfig, l.cryptoKey, l.uploadDir)
	if err != nil {
		l.updateTaskFailed(ctx, taskID, fmt.Sprintf("构建目标 Provider 失败: %v", err))
		return
	}

	var succeeded, failed int
	for _, item := range items {
		if err := l.migrateOne(ctx, item, targetProvider_, targetProvider); err != nil {
			failed++
			item.Status = "failed"
			item.Error = err.Error()
			if updateErr := l.migrationModel.UpdateItem(ctx, &item); updateErr != nil {
				l.logger.Error("更新迁移条目失败", zap.Error(updateErr))
			}
			l.logger.Error("迁移文件失败",
				zap.String("media_id", item.MediaID),
				zap.Error(err))
		} else {
			succeeded++
			item.Status = "success"
			if updateErr := l.migrationModel.UpdateItem(ctx, &item); updateErr != nil {
				l.logger.Error("更新迁移条目失败", zap.Error(updateErr))
			}
		}

		// 更新任务进度
		if task, err := l.migrationModel.GetTaskByID(ctx, taskID); err == nil {
			task.Succeeded = succeeded
			task.Failed = failed
			if updateErr := l.migrationModel.UpdateTask(ctx, task); updateErr != nil {
				l.logger.Error("更新任务进度失败", zap.Error(updateErr))
			}
		}
	}

	// 完成任务
	task, err := l.migrationModel.GetTaskByID(ctx, taskID)
	if err != nil {
		l.logger.Error("获取迁移任务失败", zap.Error(err))
		return
	}
	finishedAt := time.Now()
	task.FinishedAt = &finishedAt
	if failed > 0 {
		task.Status = "failed"
		task.Error = fmt.Sprintf("迁移完成，但有 %d 个文件失败", failed)
	} else {
		task.Status = "completed"
	}
	if err := l.migrationModel.UpdateTask(ctx, task); err != nil {
		l.logger.Error("更新迁移任务状态失败", zap.Error(err))
	}
}

// migrateOne 迁移单个文件：拉源 → 上传目标 → 验证 → 更新 media 记录。
func (l *MigrationLogic) migrateOne(ctx context.Context, item model.StorageMigrationItem, targetProvider storage.StorageProvider, targetProviderType string) error {
	if item.SourceType == "preset" {
		return l.migratePreset(ctx, item, targetProvider, targetProviderType)
	}
	media, err := l.mediaModel.GetByID(ctx, item.MediaID)
	if err != nil {
		return fmt.Errorf("查询媒体记录失败: %w", err)
	}

	// 拉源文件
	var reader io.ReadCloser
	if media.StorageType == "local" {
		file, err := os.Open(filepath.Join(l.uploadDir, media.StoragePath))
		if err != nil {
			return fmt.Errorf("打开本地文件失败: %w", err)
		}
		reader = file
	} else {
		// 云平台源：构建源 Provider 并下载
		sourceConfig, err := l.configModel.GetByProvider(ctx, media.StorageType)
		if err != nil {
			return fmt.Errorf("查询源存储配置失败: %w", err)
		}
		sourceProvider, err := storage.NewProvider(sourceConfig, l.cryptoKey, l.uploadDir)
		if err != nil {
			return fmt.Errorf("构建源存储 Provider 失败: %w", err)
		}
		rc, err := sourceProvider.Download(ctx, media.StoragePath)
		if err != nil {
			return fmt.Errorf("下载源文件失败: %w", err)
		}
		reader = rc
	}
	defer reader.Close()

	// 上传到目标存储
	newURL, err := targetProvider.Upload(ctx, media.StoragePath, reader, media.Size, media.MimeType)
	if err != nil {
		return fmt.Errorf("上传到目标存储失败: %w", err)
	}

	// 验证上传成功
	exists, err := targetProvider.Exists(ctx, media.StoragePath)
	if err != nil {
		return fmt.Errorf("验证上传结果失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("上传后文件不存在，验证失败")
	}

	// 更新 media 记录
	if err := l.mediaModel.UpdateStorageInfo(ctx, media.ID, newURL, targetProviderType); err != nil {
		return fmt.Errorf("更新媒体记录失败: %w", err)
	}

	// 更新内容管理中的 URL 引用
	oldURL := media.URL // This is the resolved URL (after AfterFind hook)
	resolvedNewURL := model.ResolveURL(newURL)
	if oldURL != "" && resolvedNewURL != "" && oldURL != resolvedNewURL {
		if err := l.updateContentURLs(ctx, oldURL, resolvedNewURL); err != nil {
			l.logger.Error("更新内容URL失败", zap.String("old_url", oldURL), zap.String("new_url", resolvedNewURL), zap.Error(err))
			// Don't fail the migration for this
		}
	}

	return nil
}

// migratePreset 迁移单个预设文件。
func (l *MigrationLogic) migratePreset(ctx context.Context, item model.StorageMigrationItem, targetProvider storage.StorageProvider, targetProviderType string) error {
	// 查询预设记录
	preset, err := l.presetModel.GetByID(ctx, item.MediaID)
	if err != nil {
		return fmt.Errorf("查询预设记录失败: %w", err)
	}

	// 查询关联的 media 以确定源存储平台
	media, err := l.mediaModel.GetByID(ctx, preset.MediaID)
	if err != nil {
		return fmt.Errorf("查询关联媒体记录失败: %w", err)
	}

	// 拉源文件
	var reader io.ReadCloser
	if media.StorageType == "local" {
		file, err := os.Open(filepath.Join(l.uploadDir, preset.OutputStoragePath))
		if err != nil {
			return fmt.Errorf("打开本地预设文件失败: %w", err)
		}
		reader = file
	} else {
		sourceConfig, err := l.configModel.GetByProvider(ctx, media.StorageType)
		if err != nil {
			return fmt.Errorf("查询源存储配置失败: %w", err)
		}
		sourceProvider, err := storage.NewProvider(sourceConfig, l.cryptoKey, l.uploadDir)
		if err != nil {
			return fmt.Errorf("构建源存储 Provider 失败: %w", err)
		}
		rc, err := sourceProvider.Download(ctx, preset.OutputStoragePath)
		if err != nil {
			return fmt.Errorf("下载源预设文件失败: %w", err)
		}
		reader = rc
	}
	defer reader.Close()

	// 上传到目标存储
	newURL, err := targetProvider.Upload(ctx, preset.OutputStoragePath, reader, preset.OutputSize, preset.MimeType)
	if err != nil {
		return fmt.Errorf("上传预设到目标存储失败: %w", err)
	}

	// 验证上传成功
	exists, err := targetProvider.Exists(ctx, preset.OutputStoragePath)
	if err != nil || !exists {
		return fmt.Errorf("预设上传后验证失败")
	}

	// 保存旧 URL 用于内容更新
	oldURL := preset.OutputURL // resolved URL after AfterFind

	// 更新预设记录
	if err := l.presetModel.UpdateStorageInfo(ctx, preset.ID, newURL); err != nil {
		return fmt.Errorf("更新预设记录失败: %w", err)
	}

	// 更新内容管理中的 URL 引用
	resolvedNewURL := model.ResolveURL(newURL)
	if oldURL != "" && resolvedNewURL != "" && oldURL != resolvedNewURL {
		if err := l.updateContentURLs(ctx, oldURL, resolvedNewURL); err != nil {
			l.logger.Error("更新内容URL失败", zap.Error(err))
		}
	}

	return nil
}

// updateContentURLs 更新内容管理表中的 URL 引用。
// TODO: 此处直接使用 model.DB 更新多张表，后续应考虑通过各 Model 方法封装。
func (l *MigrationLogic) updateContentURLs(ctx context.Context, oldURL, newURL string) error {
	db := model.DB.WithContext(ctx)

	// articles.cover_image (exact match)
	if err := db.Table("articles").Where("cover_image = ?", oldURL).Update("cover_image", newURL).Error; err != nil {
		return fmt.Errorf("更新文章封面失败: %w", err)
	}
	// articles.content (REPLACE in HTML)
	if err := db.Table("articles").Where("content LIKE ?", "%"+oldURL+"%").
		UpdateColumn("content", gorm.Expr("REPLACE(content, ?, ?)", oldURL, newURL)).Error; err != nil {
		return fmt.Errorf("更新文章内容失败: %w", err)
	}
	// travel_guides.cover_image
	if err := db.Table("travel_guides").Where("cover_image = ?", oldURL).Update("cover_image", newURL).Error; err != nil {
		return fmt.Errorf("更新攻略封面失败: %w", err)
	}
	// video_works.cover_url
	if err := db.Table("video_works").Where("cover_url = ?", oldURL).Update("cover_url", newURL).Error; err != nil {
		return fmt.Errorf("更新视频封面失败: %w", err)
	}
	// songs.cover_url
	if err := db.Table("songs").Where("cover_url = ?", oldURL).Update("cover_url", newURL).Error; err != nil {
		return fmt.Errorf("更新歌曲封面失败: %w", err)
	}
	// bloggers.avatar
	if err := db.Table("bloggers").Where("avatar = ?", oldURL).Update("avatar", newURL).Error; err != nil {
		return fmt.Errorf("更新博主头像失败: %w", err)
	}
	// bloggers.blog_icon
	if err := db.Table("bloggers").Where("blog_icon = ?", oldURL).Update("blog_icon", newURL).Error; err != nil {
		return fmt.Errorf("更新博客图标失败: %w", err)
	}
	// bloggers.page_background
	if err := db.Table("bloggers").Where("page_background = ?", oldURL).Update("page_background", newURL).Error; err != nil {
		return fmt.Errorf("更新页面背景失败: %w", err)
	}

	return nil
}

// GetMigrationStatus 获取迁移进度和失败明细。
func (l *MigrationLogic) GetMigrationStatus(ctx context.Context, taskID string) (*res.MigrationTaskRes, []res.MigrationItemRes, error) {
	task, err := l.migrationModel.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, nil, fmt.Errorf("查询迁移任务失败: %w", err)
	}

	items, err := l.migrationModel.GetItemsByTaskID(ctx, taskID)
	if err != nil {
		return nil, nil, fmt.Errorf("查询迁移条目失败: %w", err)
	}

	taskRes := l.toTaskRes(task)

	var failedItems []res.MigrationItemRes
	for _, item := range items {
		if item.Status == "failed" {
			failedItems = append(failedItems, res.MigrationItemRes{
				ID:      item.ID,
				MediaID: item.MediaID,
				Status:  item.Status,
				Error:   item.Error,
			})
		}
	}

	return &taskRes, failedItems, nil
}

// toTaskRes 将 StorageMigrationTask 模型转换为 DTO。
func (l *MigrationLogic) toTaskRes(task *model.StorageMigrationTask) res.MigrationTaskRes {
	return res.MigrationTaskRes{
		ID:             task.ID,
		TaskType:       task.TaskType,
		TargetProvider: task.TargetProvider,
		Status:         task.Status,
		Total:          task.Total,
		Succeeded:      task.Succeeded,
		Failed:         task.Failed,
		StartedAt:      task.StartedAt,
		FinishedAt:     task.FinishedAt,
		Error:          task.Error,
		CreatedAt:      task.CreatedAt,
	}
}

// updateTaskFailed 更新任务状态为失败。
func (l *MigrationLogic) updateTaskFailed(ctx context.Context, taskID, errMsg string) {
	task, err := l.migrationModel.GetTaskByID(ctx, taskID)
	if err != nil {
		l.logger.Error("获取任务失败", zap.String("task_id", taskID), zap.Error(err))
		return
	}
	task.Status = "failed"
	task.Error = errMsg
	finishedAt := time.Now()
	task.FinishedAt = &finishedAt
	if err := l.migrationModel.UpdateTask(ctx, task); err != nil {
		l.logger.Error("更新任务状态失败", zap.Error(err))
	}
}

// failTask 将任务标记为失败（同步调用，用于 StartMigration 中的前置错误）。
func (l *MigrationLogic) failTask(ctx context.Context, task *model.StorageMigrationTask, errMsg string) {
	task.Status = "failed"
	task.Error = errMsg
	finishedAt := time.Now()
	task.FinishedAt = &finishedAt
	if err := l.migrationModel.UpdateTask(ctx, task); err != nil {
		l.logger.Error("更新任务状态失败", zap.Error(err))
	}
}
