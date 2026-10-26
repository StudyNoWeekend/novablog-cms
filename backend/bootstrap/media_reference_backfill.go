// 媒体引用改造启动回填：把业务表写死的图片 URL 迁移为媒体库文件 ID 引用，
// 并把 media.url / media_presets.output_url 归一为相对对象键（完整 URL 读取时动态拼接）。
// 项目没有自动执行 migrations/*.sql 的机制（AutoMigrate 只同步表结构），数据级迁移在此幂等执行，
// 与 migrations/043_media_reference 的手工执行互不冲突。
// 整个流程在独立 goroutine 中异步运行，不阻塞服务启动；读取解析走「ID 优先、URL 兜底、
// 历史完整 URL 直通」规则，回填进行中的短暂窗口内页面无感知。
package bootstrap

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// mediaRefField 业务表图片字段的媒体引用回填配置。
// table/urlCol/mediaCol/presetCol/filter 均为编译期常量，仅用于拼接 SQL，不涉及用户输入。
type mediaRefField struct {
	table     string
	urlCol    string
	mediaCol  string
	presetCol string // 空表示该字段不使用预设
	filter    string // 附加过滤条件（已限定别名 t.，如软删除），空表示无
}

// mediaRefFields 全部需要回填的业务表图片字段（内容创作 + 关于我）。
var mediaRefFields = []mediaRefField{
	// 内容创作
	{"articles", "cover_image", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL"},
	{"songs", "cover_url", "cover_media_id", "cover_preset_id", ""},
	{"third_party_playlists", "cover_url", "cover_media_id", "cover_preset_id", ""},
	{"video_works", "cover_url", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL"},
	{"travel_guides", "cover_image", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL"},
	{"recipes", "cover", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL"},
	{"books", "cover", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL"},
	{"games", "cover", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL"},
	// 关于我
	{"photo_equipment", "image_url", "image_media_id", "image_preset_id", "t.deleted_at IS NULL"},
	{"projects", "cover_url", "cover_media_id", "cover_preset_id", "t.deleted_at IS NULL"},
	{"tech_stack_items", "icon", "icon_media_id", "", "t.deleted_at IS NULL"},
	{"bloggers", "avatar", "avatar_media_id", "", "t.deleted_at IS NULL"},
	{"bloggers", "blog_icon", "blog_icon_media_id", "", "t.deleted_at IS NULL"},
	{"bloggers", "page_background", "page_background_media_id", "", "t.deleted_at IS NULL"},
}

// backfillMediaReferencesAsync 异步执行媒体引用回填（goroutine，不阻塞启动）。
// 各步骤独立执行、失败仅记日志，下次启动自动重试；仅处理引用列为空的行，天然幂等，
// 与后台并发编辑互不覆盖。
func backfillMediaReferencesAsync(db *gorm.DB, logger *zap.Logger) {
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
	// 但回填 UPDATE 在无匹配时是零成本扫描，可接受）
	for _, f := range mediaRefFields {
		query := fmt.Sprintf(
			"SELECT EXISTS (SELECT 1 FROM %s WHERE %s IS NULL AND COALESCE(%s, '') <> ''",
			f.table, f.mediaCol, f.urlCol)
		if f.filter != "" {
			query += " AND " + f.filter
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
