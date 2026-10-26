package model

import (
	"fmt"

	"gorm.io/gorm"
)

// mediaReference 业务表媒体引用列配置：引用扫描（ID 精确匹配）与媒体强制删除时清空引用共用。
type mediaReference struct {
	table     string // 业务表名
	mediaCol  string // 媒体文件引用列（关联 media.id），空=无
	presetCol string // 预设引用列（关联 media_presets.id），空=无
}

// mediaReferences 各业务表图片字段的媒体引用列清单（内容创作 + 关于我）。
var mediaReferences = []mediaReference{
	{"articles", "cover_media_id", "cover_preset_id"},
	{"songs", "cover_media_id", "cover_preset_id"},
	{"third_party_playlists", "cover_media_id", "cover_preset_id"},
	{"video_works", "cover_media_id", "cover_preset_id"},
	{"travel_guides", "cover_media_id", "cover_preset_id"},
	{"recipes", "cover_media_id", "cover_preset_id"},
	{"books", "cover_media_id", "cover_preset_id"},
	{"games", "cover_media_id", "cover_preset_id"},
	{"photo_equipment", "image_media_id", "image_preset_id"},
	{"projects", "cover_media_id", "cover_preset_id"},
	{"tech_stack_items", "icon_media_id", ""},
	{"bloggers", "avatar_media_id", ""},
	{"bloggers", "blog_icon_media_id", ""},
	{"bloggers", "page_background_media_id", ""},
}

// clearMediaReferences 在事务内清空各业务表对指定媒体及其预设的引用列，避免遗留死 ID。
// 仅清引用列，历史 URL 列保持原值（引用行展示随之降级为空/外链）。
func clearMediaReferences(tx *gorm.DB, mediaID string, presetIDs []string) error {
	for _, ref := range mediaReferences {
		conds := make([]string, 0, 2)
		args := make([]any, 0, 2)
		updates := map[string]any{}
		if ref.mediaCol != "" {
			conds = append(conds, ref.mediaCol+" = ?")
			args = append(args, mediaID)
			updates[ref.mediaCol] = nil
		}
		if ref.presetCol != "" && len(presetIDs) > 0 {
			conds = append(conds, ref.presetCol+" IN ?")
			args = append(args, presetIDs)
			updates[ref.presetCol] = nil
		}
		if len(conds) == 0 {
			continue
		}
		// conds 依次展开为条件与参数： Where(mediaCol = ? OR presetCol IN ?, mediaID, presetIDs)
		cond := "(" + conds[0]
		bindArgs := []any{args[0]}
		for i := 1; i < len(conds); i++ {
			cond += " OR " + conds[i]
			bindArgs = append(bindArgs, args[i])
		}
		cond += ")"
		if err := tx.Table(ref.table).Where(cond, bindArgs...).Updates(updates).Error; err != nil {
			return fmt.Errorf("清空 %s 媒体引用失败: %w", ref.table, err)
		}
	}
	return nil
}
