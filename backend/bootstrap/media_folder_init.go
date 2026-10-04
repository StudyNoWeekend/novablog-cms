// 媒体文件夹启动初始化：首次启动（media_folders 表为空）时播种模块顶级文件夹，
// 并按引用扫描把存量媒体自动归入对应模块文件夹；此后启动直接跳过，保证幂等且
// 不覆盖用户后续手动整理的结果。
package bootstrap

import (
	"context"

	"novablog/enum"
	"novablog/internal/logic"
	"novablog/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// initMediaFolders 播种模块文件夹并归类存量媒体（仅首次启动执行一次）。
// 失败不阻塞启动：上传时仍会按 module 参数懒创建模块文件夹。
func initMediaFolders(db *gorm.DB) {
	logger := logic.MediaLogger
	if logger == nil {
		logger = zap.NewNop()
	}

	var count int64
	if err := db.Model(&model.MediaFolder{}).Count(&count).Error; err != nil {
		logger.Warn("统计媒体文件夹失败，跳过初始化", zap.Error(err))
		return
	}
	if count > 0 {
		return
	}

	ctx := context.Background()

	// 1. 播种模块顶级文件夹
	folders := make([]model.MediaFolder, 0, len(enum.MediaModuleOrder))
	folderIDByName := make(map[string]string, len(enum.MediaModuleOrder))
	for _, key := range enum.MediaModuleOrder {
		name := enum.MediaModuleFolders[key]
		moduleKey := key
		folders = append(folders, model.MediaFolder{
			ID:        uuid.New().String(),
			Name:      name,
			ModuleKey: &moduleKey,
		})
		folderIDByName[name] = folders[len(folders)-1].ID
	}
	if err := db.Create(&folders).Error; err != nil {
		logger.Warn("播种媒体模块文件夹失败", zap.Error(err))
		return
	}
	logger.Info("已播种媒体模块文件夹", zap.Int("count", len(folders)))

	// 2. 存量媒体自动归类：folder_id 为空且未被引用的保持根目录
	mediaModel := model.NewMedia()
	presetModel := model.NewMediaPreset()
	usageModel := model.NewMediaUsage()

	var mediaList []model.Media
	if err := db.WithContext(ctx).
		Session(&gorm.Session{SkipHooks: true}).
		Where("folder_id IS NULL AND deleted_at IS NULL").
		Find(&mediaList).Error; err != nil {
		logger.Warn("查询待归类媒体失败", zap.Error(err))
		return
	}

	classified := 0
	for i := range mediaList {
		m := &mediaList[i]
		presets, err := presetModel.GetAllByMediaIDUnscoped(ctx, m.ID)
		if err != nil {
			logger.Warn("查询媒体预设失败，跳过归类", zap.String("media_id", m.ID), zap.Error(err))
			continue
		}
		presetIDs := make([]string, 0, len(presets))
		presetKeys := make([]string, 0, len(presets))
		for _, p := range presets {
			presetIDs = append(presetIDs, p.ID)
			presetKeys = append(presetKeys, p.OutputStoragePath)
		}

		hits, err := usageModel.FindUsages(ctx, m.URL, m.StoragePath, presetIDs, presetKeys)
		if err != nil {
			logger.Warn("扫描媒体引用失败，跳过归类", zap.String("media_id", m.ID), zap.Error(err))
			continue
		}

		// 按模块固定顺序取第一个命中的模块文件夹
		for _, key := range enum.MediaModuleOrder {
			folderName := enum.MediaModuleFolders[key]
			matched := false
			for _, h := range hits {
				if h.Module == folderName {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
			folderID := folderIDByName[folderName]
			if err := mediaModel.UpdateFolderIDs(ctx, []string{m.ID}, &folderID); err != nil {
				logger.Warn("归类媒体失败", zap.String("media_id", m.ID), zap.String("folder", folderName), zap.Error(err))
			} else {
				classified++
			}
			break
		}
	}
	logger.Info("存量媒体自动归类完成",
		zap.Int("total", len(mediaList)),
		zap.Int("classified", classified),
	)
}
