// 媒体引用改造启动回填：把业务表写死的图片 URL 迁移为媒体库文件 ID 引用，
// 并把 media.url / media_presets.output_url 归一为相对对象键（完整 URL 读取时动态拼接）。
// 项目没有自动执行 migrations/*.sql 的机制（AutoMigrate 只同步表结构），数据级迁移在此幂等执行，
// 与 migrations/043_media_reference 的手工执行互不冲突。
// 整个流程在独立 goroutine 中异步运行，不阻塞服务启动；读取解析走「ID 优先、URL 兜底、
// 历史完整 URL 直通」规则，回填进行中的短暂窗口内页面无感知。
package bootstrap

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"novablog/enum"
	"novablog/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// mediaRefField 业务表图片字段的媒体引用回填配置。
// table/urlCol/mediaCol/presetCol/filter/moduleKey 均为编译期常量，仅用于拼接 SQL，不涉及用户输入。
type mediaRefField struct {
	table     string
	urlCol    string
	mediaCol  string
	presetCol string // 空表示该字段不使用预设
	filter    string // 附加过滤条件（已限定别名 t.，如软删除），空表示无
	moduleKey string // 遗留文件收编时归类的媒体库模块文件夹 key
}

// mediaRefFields 全部需要回填的业务表图片字段（内容创作 + 关于我）。
var mediaRefFields = []mediaRefField{
	// 内容创作
	{"articles", "cover_image", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL", enum.MediaModuleArticle},
	{"songs", "cover_url", "cover_media_id", "cover_preset_id", "", enum.MediaModuleMusic},
	{"third_party_playlists", "cover_url", "cover_media_id", "cover_preset_id", "", enum.MediaModuleMusic},
	{"video_works", "cover_url", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL", enum.MediaModuleVideo},
	{"travel_guides", "cover_image", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL", enum.MediaModuleTravel},
	{"recipes", "cover", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL", enum.MediaModuleRecipe},
	{"books", "cover", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL", enum.MediaModuleBook},
	{"games", "cover", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL", enum.MediaModuleGame},
	// 关于我
	{"photo_equipment", "image_url", "image_media_id", "image_preset_id", "t.deleted_at IS NULL", enum.MediaModuleEquipment},
	{"projects", "cover_url", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL", enum.MediaModuleProject},
	{"tech_stack_items", "icon", "icon_media_id", "", "t.deleted_at IS NULL", enum.MediaModuleTechStack},
	{"bloggers", "avatar", "avatar_media_id", "", "t.deleted_at IS NULL", enum.MediaModuleProfile},
	{"bloggers", "blog_icon", "blog_icon_media_id", "", "t.deleted_at IS NULL", enum.MediaModuleProfile},
	{"bloggers", "page_background", "page_background_media_id", "", "t.deleted_at IS NULL", enum.MediaModuleProfile},
}

// backfillMediaReferencesAsync 异步执行媒体引用回填（goroutine，不阻塞启动）。
// uploadDir 为本地存储根目录（遗留文件收编时做磁盘存在性判定，空=跳过磁盘检查）。
// 各步骤独立执行、失败仅记日志，下次启动自动重试；仅处理引用列为空的行，天然幂等，
// 与后台并发编辑互不覆盖。
func backfillMediaReferencesAsync(db *gorm.DB, logger *zap.Logger, uploadDir string) {
	go func() {
		if logger == nil {
			logger = zap.NewNop()
		}
		defer func() {
			if r := recover(); r != nil {
				logger.Error("媒体引用回填异常退出", zap.Any("panic", r))
			}
		}()

		if !mediaBackfillNeeded(db) {
			logger.Info("媒体引用回填：无待处理数据，跳过")
			return
		}

		normalizeMediaURLs(db, logger)
		backfillMediaRefFields(db, logger)
		adoptLegacyLocalFiles(db, logger, uploadDir)
		logger.Info("媒体引用回填完成")
	}()
}

// mediaBackfillNeeded 判断是否需要执行媒体引用回填：
// 任一相关表缺失（AutoMigrate 异常等极端情况）→ 不执行，避免无谓报错；
// 归一与各业务字段均无待处理行 → 无需执行。
func mediaBackfillNeeded(db *gorm.DB) bool {
	tables := []string{"media", "media_presets", "articles", "songs", "third_party_playlists",
		"video_works", "travel_guides", "recipes", "books", "games",
		"photo_equipment", "projects", "tech_stack_items", "bloggers"}
	for _, t := range tables {
		if !db.Migrator().HasTable(t) {
			return false
		}
	}

	// 归一待处理：对象存储记录的 url 仍是完整 URL，或预设缺 storage_type / output_url 未归一
	var need bool
	err := db.Raw(`SELECT EXISTS (
		SELECT 1 FROM media
		WHERE storage_type IS NOT NULL AND storage_type <> '' AND storage_type <> 'local'
		  AND storage_path IS NOT NULL AND storage_path <> ''
		  AND url <> storage_path
	) OR EXISTS (
		SELECT 1 FROM media_presets
		WHERE output_storage_path IS NOT NULL AND output_storage_path <> ''
		  AND output_url <> output_storage_path
	)`).Scan(&need).Error
	if err != nil {
		// 探测失败按无需回填处理，不影响启动
		return false
	}
	if need {
		return true
	}

	// 业务字段待处理：存在引用列为空且 URL 非空的行（纯外链行也会命中，
	// 但回填 UPDATE 在无匹配时是零成本扫描，可接受）。
	// filter 为 UPDATE...FROM 拼接用的 "t.deleted_at IS NULL"，裸 SELECT 需去掉别名前缀
	// （否则 SQL 报错被当成"无需回填"，整个回填被跳过——v1.2.12 首发即栽在这里）。
	for _, f := range mediaRefFields {
		query := fmt.Sprintf(
			"SELECT EXISTS (SELECT 1 FROM %s WHERE %s IS NULL AND COALESCE(%s, '') <> ''",
			f.table, f.mediaCol, f.urlCol)
		if f.filter != "" {
			query += " AND " + strings.Replace(f.filter, "t.", "", 1)
		}
		query += ")"
		var pending bool
		if err := db.Raw(query).Scan(&pending).Error; err != nil {
			return false
		}
		if pending {
			return true
		}
	}
	return false
}

// normalizeMediaURLs 把 media.url / media_presets.output_url 归一为相对对象键，
// 并补齐预设的 storage_type（继承所属媒体）。
func normalizeMediaURLs(db *gorm.DB, logger *zap.Logger) {
	if err := db.Exec(`UPDATE media SET url = storage_path
		WHERE storage_type IS NOT NULL AND storage_type <> '' AND storage_type <> 'local'
		  AND storage_path IS NOT NULL AND storage_path <> ''
		  AND url <> storage_path`).Error; err != nil {
		logger.Warn("media.url 归一失败（下次启动重试）", zap.Error(err))
	}

	if err := db.Exec(`UPDATE media_presets p SET storage_type = m.storage_type
		FROM media m
		WHERE p.media_id = m.id AND (p.storage_type IS NULL OR p.storage_type = '')`).Error; err != nil {
		logger.Warn("media_presets.storage_type 回填失败（下次启动重试）", zap.Error(err))
	}

	if err := db.Exec(`UPDATE media_presets SET output_url = output_storage_path
		WHERE output_storage_path IS NOT NULL AND output_storage_path <> ''
		  AND output_url <> output_storage_path`).Error; err != nil {
		logger.Warn("media_presets.output_url 归一失败（下次启动重试）", zap.Error(err))
	}
}

// backfillMediaRefFields 按业务表逐字段回填媒体引用：
// 先按 URL 与 media.url 全等 / storage_path 包含匹配填 *_media_id，
// 再按预设成品图 output_storage_path 包含匹配填 *_preset_id（指向预设成品图的 URL
// 不含原图 key，故预设独立匹配并同填 media_id；两组 key 均含唯一 uuid，互不冲突）。
func backfillMediaRefFields(db *gorm.DB, logger *zap.Logger) {
	for _, f := range mediaRefFields {
		query := fmt.Sprintf(`
UPDATE %s t SET %s = m.id
FROM media m
WHERE t.%s IS NULL
  AND COALESCE(t.%s, '') <> ''
  AND (t.%s = m.url OR (m.storage_path <> '' AND t.%s LIKE '%%' || m.storage_path || '%%'))`,
			f.table, f.mediaCol, f.mediaCol, f.urlCol, f.urlCol, f.urlCol)
		if f.filter != "" {
			query += " AND " + f.filter
		}
		if err := db.Exec(query).Error; err != nil {
			logger.Warn("业务表媒体引用回填失败（下次启动重试）",
				zap.String("table", f.table), zap.String("column", f.mediaCol), zap.Error(err))
		}

		if f.presetCol == "" {
			continue
		}
		// 预设成品图 URL 不包含原图 storage_path，需独立匹配；同填 media_id（预设必然归属某媒体）
		presetQuery := fmt.Sprintf(`
UPDATE %s t SET %s = p.id, %s = p.media_id
FROM media_presets p
WHERE t.%s IS NULL AND t.%s IS NULL
  AND p.output_storage_path <> ''
  AND COALESCE(t.%s, '') LIKE '%%' || p.output_storage_path || '%%'`,
			f.table, f.presetCol, f.mediaCol, f.mediaCol, f.presetCol, f.urlCol)
		if f.filter != "" {
			presetQuery += " AND " + f.filter
		}
		if err := db.Exec(presetQuery).Error; err != nil {
			logger.Warn("业务表预设引用回填失败（下次启动重试）",
				zap.String("table", f.table), zap.String("column", f.presetCol), zap.Error(err))
		}
	}
}

// adoptLegacyLocalFiles 遗留文件收编（引用回填之后的第三步）：
// 早期直写磁盘流程不登记媒体库，业务表里烙的是完整 URL（含当时的域名）、文件在
// media 表中无记录，ID 引用统一机制覆盖不到。本步骤扫描引用列为空且 URL 含
// /files/ 的行，提取对象 key（解码中文路径，不依赖域名——旧域名失效也能处理）：
//   - media 已有该 storage_path → 仅回填引用列；
//   - media 无记录但磁盘文件存在（uploadDir 非空且文件存在）→ 以 local 存储补建
//     media 记录并归类到所属模块文件夹，再回填引用列；
//   - 磁盘也没有 → 保持 URL 模式原样。
//
// 幂等：仅处理引用列为空的行；media 建行按 storage_path 防重；单行失败仅记日志。
func adoptLegacyLocalFiles(db *gorm.DB, logger *zap.Logger, uploadDir string) {
	for _, f := range mediaRefFields {
		query := fmt.Sprintf(
			"SELECT id, %s AS url FROM %s WHERE %s IS NULL AND %s LIKE '%%/files/%%'",
			f.urlCol, f.table, f.mediaCol, f.urlCol)
		if f.filter != "" {
			// filter 形如 "t.deleted_at IS NULL"(为 UPDATE...FROM 的歧义消解限定别名),
			// 裸 SELECT 无别名,去掉 "t." 前缀
			query += " AND " + strings.Replace(f.filter, "t.", "", 1)
		}
		var rows []struct {
			ID  string
			URL string
		}
		if err := db.Raw(query).Scan(&rows).Error; err != nil {
			logger.Warn("遗留文件收编扫描失败", zap.String("table", f.table), zap.Error(err))
			continue
		}
		for _, row := range rows {
			key := extractFilesKey(row.URL)
			if key == "" {
				continue
			}
			mediaID, ok, err := ensureMediaForLocalFile(db, uploadDir, key, f.moduleKey)
			if err != nil {
				logger.Warn("遗留文件收编失败", zap.String("key", key), zap.String("table", f.table), zap.Error(err))
				continue
			}
			if !ok {
				continue // 磁盘无文件，保持 URL 模式
			}
			update := fmt.Sprintf("UPDATE %s SET %s = ? WHERE id = ?", f.table, f.mediaCol)
			if err := db.Exec(update, mediaID, row.ID).Error; err != nil {
				logger.Warn("遗留文件引用回填失败", zap.String("table", f.table), zap.Error(err))
				continue
			}
			logger.Info("遗留文件已收编进媒体库", zap.String("key", key), zap.String("table", f.table))
		}
	}
}

// extractFilesKey 从本站 /files/ 完整 URL 中提取对象 key（解码中文路径），
// 非 /files/ 形态或含异常片段返回空串。
func extractFilesKey(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if i := strings.IndexAny(raw, "?#"); i >= 0 {
		raw = raw[:i]
	}
	parts := strings.SplitN(raw, "/files/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return ""
	}
	key, err := url.PathUnescape(parts[1])
	if err != nil {
		key = parts[1]
	}
	key = strings.Trim(key, "/")
	if key == "" || strings.Contains(key, "..") || strings.Contains(key, "://") {
		return ""
	}
	return key
}

// ensureMediaForLocalFile 保证 key 对应的媒体记录存在：
// 已有 → 返回其 ID；缺失且磁盘文件存在 → 以 local 存储补建并归类模块文件夹；
// 磁盘也缺失（或 uploadDir 为空无法判定）→ ok=false（调用方保持 URL 模式）。
func ensureMediaForLocalFile(db *gorm.DB, uploadDir, key, moduleKey string) (mediaID string, ok bool, err error) {
	var existing model.Media
	switch findErr := db.Where("storage_path = ?", key).First(&existing).Error; {
	case findErr == nil:
		return existing.ID, true, nil
	case !errors.Is(findErr, gorm.ErrRecordNotFound):
		return "", false, findErr
	}

	if uploadDir == "" {
		return "", false, nil
	}
	uploadRoot := filepath.Clean(uploadDir)
	full := filepath.Join(uploadRoot, filepath.FromSlash(key))
	if !strings.HasPrefix(full, uploadRoot+string(os.PathSeparator)) {
		return "", false, nil
	}
	info, statErr := os.Stat(full)
	if statErr != nil || info.IsDir() {
		return "", false, nil
	}

	ext := strings.ToLower(filepath.Ext(key))
	media := model.Media{
		ID:          uuid.New().String(),
		Filename:    filepath.Base(key),
		FileType:    legacyFileType(ext),
		MimeType:    legacyMimeType(ext),
		URL:         key, // 只存相对 key，完整 URL 读取时按当前存储动态拼接
		StoragePath: key,
		StorageType: "local",
	}
	if moduleKey != "" {
		var folder model.MediaFolder
		if folderErr := db.Where("module_key = ? AND parent_id IS NULL", moduleKey).First(&folder).Error; folderErr == nil {
			media.FolderID = &folder.ID
		}
	}
	if createErr := db.Create(&media).Error; createErr != nil {
		return "", false, createErr
	}
	return media.ID, true, nil
}

// legacyFileType 按扩展名推断媒体类型（与 logic.getFileType 口径一致）。
func legacyFileType(ext string) int16 {
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

// legacyMimeType 按扩展名推断 MIME（与 logic.getMimeType 口径一致）。
func legacyMimeType(ext string) string {
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
