package model

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// 引用扫描的模块名称（与媒体库模块文件夹名保持一致，供自动归类按名匹配）。
const (
	UsageModuleArticle   = "文章"
	UsageModuleRecipe    = "美食菜谱"
	UsageModuleBook      = "读书书架"
	UsageModuleGame      = "游戏库"
	UsageModuleTechStack = "技术栈"
	UsageModuleTravel    = "旅行攻略"
	UsageModulePortfolio = "摄影作品集"
	UsageModuleProject   = "项目经历"
	UsageModuleEquipment = "个人设备"
	UsageModuleVideo     = "视频作品"
	UsageModuleSong      = "音乐"
	UsageModuleBlogger   = "博主资料"
)

// UsageHit 媒体引用命中：某内容模块中的某条内容在某个位置引用了该媒体。
type UsageHit struct {
	Module string // 模块名，如 文章/旅行攻略
	ID     string // 内容条目 ID
	Title  string // 内容条目标题
	Field  string // 引用位置，如 封面/正文
}

// usageRow 引用扫描的行结果。
type usageRow struct {
	ID    string `gorm:"column:id"`
	Title string `gorm:"column:title"`
}

// MediaUsageModel 媒体引用扫描模型，删除前实时查询各内容表中的引用。
type MediaUsageModel struct {
	db *gorm.DB
}

// NewMediaUsage 创建 MediaUsageModel 实例。
func NewMediaUsage() *MediaUsageModel {
	return &MediaUsageModel{db: DB}
}

// FindUsages 扫描各内容表中对指定媒体的引用。
// mediaID 为媒体 ID：结构化字段的 *_media_id/*_preset_id 引用列按 ID 精确匹配；
// mediaURL 为 media.url 的原始存储值（相对路径或历史完整 URL），mediaKey 为
// media.storage_path；presetIDs 为该媒体的预设 ID，presetKeys 为预设成品图
// 的 storage_path。引用处无论存相对路径还是完整 URL 都包含 key 子串，故对
// 文本/jsonb 以 key 做 LIKE 匹配（含稳定链接 /media/{id}/file 片段），URL 列以
// 全等/LIKE 兜底。命中按 (模块,内容ID,位置) 去重。
// songs/third_party_playlists 表无软删除列，扫描时不做 deleted_at 过滤。
func (m *MediaUsageModel) FindUsages(ctx context.Context, mediaID, mediaURL, mediaKey string, presetIDs, presetKeys []string) ([]UsageHit, error) {
	keys := dedupNonEmpty(append([]string{mediaKey}, presetKeys...))
	if mediaID != "" {
		// 稳定链接按媒体 ID 索引，正文/jsonb 中的稳定链接引用同样可扫描到
		keys = append(keys, "api/v1/public/media/"+mediaID+"/file")
	}
	if mediaURL == "" && len(keys) == 0 && len(presetIDs) == 0 && mediaID == "" {
		return nil, nil
	}

	hits := make([]UsageHit, 0)

	// 文章：封面 / 正文
	if err := m.scanURLField(ctx, "articles", UsageModuleArticle, "封面", "title", "cover_image", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}
	if err := m.scanTextField(ctx, "articles", UsageModuleArticle, "正文", "title", "content", keys, true, &hits); err != nil {
		return nil, err
	}

	// 美食菜谱：封面 / 步骤图（steps 为 jsonb）
	if err := m.scanURLField(ctx, "recipes", UsageModuleRecipe, "封面", "title", "cover", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}
	if err := m.scanTextField(ctx, "recipes", UsageModuleRecipe, "步骤图", "title", "steps::text", keys, true, &hits); err != nil {
		return nil, err
	}

	// 读书书架：封面
	if err := m.scanURLField(ctx, "books", UsageModuleBook, "封面", "title", "cover", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}

	// 游戏库：封面
	if err := m.scanURLField(ctx, "games", UsageModuleGame, "封面", "title", "cover", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}

	// 技术栈：图标
	if err := m.scanURLField(ctx, "tech_stack_items", UsageModuleTechStack, "图标", "name", "icon", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}

	// 旅行攻略：封面 / 景点图 / 行程图（attractions、itinerary 为 jsonb）
	if err := m.scanURLField(ctx, "travel_guides", UsageModuleTravel, "封面", "title", "cover_image", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}
	if err := m.scanTextField(ctx, "travel_guides", UsageModuleTravel, "景点图", "title", "attractions::text", keys, true, &hits); err != nil {
		return nil, err
	}
	if err := m.scanTextField(ctx, "travel_guides", UsageModuleTravel, "行程图", "title", "itinerary::text", keys, true, &hits); err != nil {
		return nil, err
	}

	// 摄影作品集：通过预设 ID 间接引用
	if err := m.scanPortfolios(ctx, presetIDs, &hits); err != nil {
		return nil, err
	}

	// 个人设备：图片
	if err := m.scanURLField(ctx, "photo_equipment", UsageModuleEquipment, "图片", "name", "image_url", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}

	// 项目经历：封面
	if err := m.scanURLField(ctx, "projects", UsageModuleProject, "封面", "title", "cover_url", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}

	// 视频作品：封面
	if err := m.scanURLField(ctx, "video_works", UsageModuleVideo, "封面", "title", "cover_url", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}

	// 音乐：封面 / 音源
	if err := m.scanURLField(ctx, "songs", UsageModuleSong, "封面", "title", "cover_url", mediaURL, keys, false, &hits); err != nil {
		return nil, err
	}
	if err := m.scanURLField(ctx, "songs", UsageModuleSong, "音源", "title", "source_url", mediaURL, keys, false, &hits); err != nil {
		return nil, err
	}

	// 博主信息：头像 / 博客图标 / 页面背景
	bloggerTitle := "COALESCE(NULLIF(nickname, ''), username)"
	if err := m.scanURLField(ctx, "bloggers", UsageModuleBlogger, "头像", bloggerTitle, "avatar", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}
	if err := m.scanURLField(ctx, "bloggers", UsageModuleBlogger, "博客图标", bloggerTitle, "blog_icon", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}
	if err := m.scanURLField(ctx, "bloggers", UsageModuleBlogger, "页面背景", bloggerTitle, "page_background", mediaURL, keys, true, &hits); err != nil {
		return nil, err
	}

	// 结构化字段的媒体引用列（*_media_id / *_preset_id）ID 精确匹配
	if err := m.scanMediaRefs(ctx, mediaID, presetIDs, &hits); err != nil {
		return nil, err
	}

	return dedupeUsageHits(hits), nil
}

// mediaRefScan 结构化字段媒体引用列的扫描配置。
type mediaRefScan struct {
	table         string
	module        string
	field         string
	titleExpr     string
	mediaCol      string
	presetCol     string
	filterDeleted bool
}

// mediaRefScans 各业务表 *_media_id / *_preset_id 引用列的扫描配置（内容创作 + 关于我）。
var mediaRefScans = []mediaRefScan{
	{table: "articles", module: UsageModuleArticle, field: "封面", titleExpr: "title", mediaCol: "cover_media_id", presetCol: "cover_preset_id", filterDeleted: true},
	{table: "songs", module: UsageModuleSong, field: "封面", titleExpr: "title", mediaCol: "cover_media_id", presetCol: "cover_preset_id"},
	{table: "third_party_playlists", module: UsageModuleSong, field: "歌单封面", titleExpr: "title", mediaCol: "cover_media_id", presetCol: "cover_preset_id"},
	{table: "video_works", module: UsageModuleVideo, field: "封面", titleExpr: "title", mediaCol: "cover_media_id", presetCol: "cover_preset_id", filterDeleted: true},
	{table: "travel_guides", module: UsageModuleTravel, field: "封面", titleExpr: "title", mediaCol: "cover_media_id", presetCol: "cover_preset_id", filterDeleted: true},
	{table: "recipes", module: UsageModuleRecipe, field: "封面", titleExpr: "title", mediaCol: "cover_media_id", presetCol: "cover_preset_id", filterDeleted: true},
	{table: "books", module: UsageModuleBook, field: "封面", titleExpr: "title", mediaCol: "cover_media_id", presetCol: "cover_preset_id", filterDeleted: true},
	{table: "games", module: UsageModuleGame, field: "封面", titleExpr: "title", mediaCol: "cover_media_id", presetCol: "cover_preset_id", filterDeleted: true},
	{table: "photo_equipment", module: UsageModuleEquipment, field: "图片", titleExpr: "name", mediaCol: "image_media_id", presetCol: "image_preset_id", filterDeleted: true},
	{table: "projects", module: UsageModuleProject, field: "封面", titleExpr: "title", mediaCol: "cover_media_id", presetCol: "cover_preset_id", filterDeleted: true},
	{table: "tech_stack_items", module: UsageModuleTechStack, field: "图标", titleExpr: "name", mediaCol: "icon_media_id", filterDeleted: true},
	{table: "bloggers", module: UsageModuleBlogger, field: "头像", titleExpr: "COALESCE(NULLIF(nickname, ''), username)", mediaCol: "avatar_media_id", filterDeleted: true},
	{table: "bloggers", module: UsageModuleBlogger, field: "博客图标", titleExpr: "COALESCE(NULLIF(nickname, ''), username)", mediaCol: "blog_icon_media_id", filterDeleted: true},
	{table: "bloggers", module: UsageModuleBlogger, field: "页面背景", titleExpr: "COALESCE(NULLIF(nickname, ''), username)", mediaCol: "page_background_media_id", filterDeleted: true},
}

// scanMediaRefs 按 ID 精确扫描各业务表的媒体/预设引用列。
func (m *MediaUsageModel) scanMediaRefs(ctx context.Context, mediaID string, presetIDs []string, hits *[]UsageHit) error {
	if mediaID == "" && len(presetIDs) == 0 {
		return nil
	}
	for _, ref := range mediaRefScans {
		conds := make([]string, 0, 2)
		args := make([]any, 0, 2)
		if mediaID != "" && ref.mediaCol != "" {
			conds = append(conds, ref.mediaCol+" = ?")
			args = append(args, mediaID)
		}
		if ref.presetCol != "" && len(presetIDs) > 0 {
			conds = append(conds, ref.presetCol+" IN ?")
			args = append(args, presetIDs)
		}
		if len(conds) == 0 {
			continue
		}
		cond := "(" + strings.Join(conds, " OR ") + ")"
		if ref.filterDeleted {
			cond = "deleted_at IS NULL AND " + cond
		}
		var rows []usageRow
		err := m.db.WithContext(ctx).
			Table(ref.table).
			Select("id, "+ref.titleExpr+" AS title").
			Where(cond, args...).
			Scan(&rows).Error
		if err != nil {
			return fmt.Errorf("扫描%s媒体引用失败: %w", ref.module, err)
		}
		for _, r := range rows {
			*hits = append(*hits, UsageHit{Module: ref.module, ID: r.ID, Title: r.Title, Field: ref.field})
		}
	}
	return nil
}

// dedupeUsageHits 按 (模块,内容ID,位置) 去重：同一行可能同时被 URL 兜底与 ID 精确扫描命中。
func dedupeUsageHits(hits []UsageHit) []UsageHit {
	seen := make(map[string]struct{}, len(hits))
	result := make([]UsageHit, 0, len(hits))
	for _, h := range hits {
		key := h.Module + "|" + h.ID + "|" + h.Field
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, h)
	}
	return result
}

// scanPortfolios 扫描摄影作品集对媒体预设的引用：作品项的作品图、作品集独立封面。
func (m *MediaUsageModel) scanPortfolios(ctx context.Context, presetIDs []string, hits *[]UsageHit) error {
	if len(presetIDs) == 0 {
		return nil
	}

	var items []usageRow
	err := m.db.WithContext(ctx).
		Table("portfolio_items pi").
		Select("pi.id, p.name || ' / ' || pi.title AS title").
		Joins("JOIN portfolios p ON p.id = pi.portfolio_id AND p.deleted_at IS NULL").
		Where("pi.preset_id IN ? AND pi.deleted_at IS NULL", presetIDs).
		Scan(&items).Error
	if err != nil {
		return fmt.Errorf("扫描摄影作品集引用失败: %w", err)
	}
	for _, r := range items {
		*hits = append(*hits, UsageHit{Module: UsageModulePortfolio, ID: r.ID, Title: r.Title, Field: "作品图"})
	}

	var covers []usageRow
	err = m.db.WithContext(ctx).
		Table("portfolios").
		Select("id, name AS title").
		Where("cover_preset_id IN ? AND deleted_at IS NULL", presetIDs).
		Scan(&covers).Error
	if err != nil {
		return fmt.Errorf("扫描摄影作品集封面引用失败: %w", err)
	}
	for _, r := range covers {
		*hits = append(*hits, UsageHit{Module: UsageModulePortfolio, ID: r.ID, Title: r.Title, Field: "封面"})
	}

	return nil
}

// scanURLField 扫描单个 URL 字段（封面类）：精确匹配 URL 或 LIKE 任意 key。
func (m *MediaUsageModel) scanURLField(ctx context.Context, table, module, field, titleExpr, col, mediaURL string, keys []string, filterDeleted bool, hits *[]UsageHit) error {
	cond, args := urlCond(col, mediaURL, keys)
	return m.scan(ctx, table, module, field, titleExpr, cond, args, filterDeleted, hits)
}

// scanTextField 扫描单个文本/jsonb 字段（正文类）：LIKE 任意 key，keys 为空时跳过。
func (m *MediaUsageModel) scanTextField(ctx context.Context, table, module, field, titleExpr, col string, keys []string, filterDeleted bool, hits *[]UsageHit) error {
	cond, args := likeCond(col, keys)
	return m.scan(ctx, table, module, field, titleExpr, cond, args, filterDeleted, hits)
}

// scan 执行单字段引用查询并把命中追加到 hits。
func (m *MediaUsageModel) scan(ctx context.Context, table, module, field, titleExpr, cond string, args []any, filterDeleted bool, hits *[]UsageHit) error {
	if cond == "" {
		return nil
	}
	if filterDeleted {
		cond = "deleted_at IS NULL AND " + cond
	}
	var rows []usageRow
	err := m.db.WithContext(ctx).
		Table(table).
		Select("id, "+titleExpr+" AS title").
		Where(cond, args...).
		Scan(&rows).Error
	if err != nil {
		return fmt.Errorf("扫描%s引用失败: %w", module, err)
	}
	for _, r := range rows {
		*hits = append(*hits, UsageHit{Module: module, ID: r.ID, Title: r.Title, Field: field})
	}
	return nil
}

// urlCond 生成「精确等于 URL 或 LIKE 任意 key」的条件片段及参数。
func urlCond(col, mediaURL string, keys []string) (string, []any) {
	parts := make([]string, 0, 1+len(keys))
	args := make([]any, 0, 1+len(keys))
	if mediaURL != "" {
		parts = append(parts, col+" = ?")
		args = append(args, mediaURL)
	}
	for _, k := range keys {
		parts = append(parts, col+" LIKE ?")
		args = append(args, "%"+k+"%")
	}
	if len(parts) == 0 {
		return "", nil
	}
	return "(" + strings.Join(parts, " OR ") + ")", args
}

// likeCond 生成「LIKE 任意 key」的条件片段及参数，keys 为空时返回空串。
func likeCond(col string, keys []string) (string, []any) {
	if len(keys) == 0 {
		return "", nil
	}
	parts := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, col+" LIKE ?")
		args = append(args, "%"+k+"%")
	}
	return "(" + strings.Join(parts, " OR ") + ")", args
}

// dedupNonEmpty 去除空字符串并去重，保持原顺序。
func dedupNonEmpty(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
