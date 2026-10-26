package model

import (
	"context"
)

// ResolveMediaURLMap 批量解析媒体文件与预设的当前完整 URL，返回 map[ID]URL。
// 查询走 AfterFind 钩子：本地相对路径拼 BaseURL/files，对象存储按当前存储配置域名动态拼接，
// 历史完整 URL 原样返回。业务表图片字段引用 media_id/preset_id 后由此解析展示链接。
func ResolveMediaURLMap(ctx context.Context, mediaIDs, presetIDs []string) (mediaURLs map[string]string, presetURLs map[string]string, err error) {
	mediaURLs = make(map[string]string, len(mediaIDs))
	presetURLs = make(map[string]string, len(presetIDs))

	if len(mediaIDs) > 0 {
		mediaList, err := NewMedia().GetByIDs(ctx, mediaIDs)
		if err != nil {
			return nil, nil, err
		}
		for _, m := range mediaList {
			mediaURLs[m.ID] = m.URL
		}
	}

	if len(presetIDs) > 0 {
		presets, err := NewMediaPreset().GetByIDs(ctx, presetIDs)
		if err != nil {
			return nil, nil, err
		}
		for _, p := range presets {
			presetURLs[p.ID] = p.OutputURL
		}
	}

	return mediaURLs, presetURLs, nil
}

// ResolveMediaRefURL 按引用优先级解析单条业务记录的图片 URL：
// preset_id → media_id → 历史落库 URL（外链）。任一引用在媒体库中不存在时自动降级。
func ResolveMediaRefURL(ctx context.Context, mediaID, presetID, legacyURL string) string {
	if presetID != "" {
		presets, err := NewMediaPreset().GetByIDs(ctx, []string{presetID})
		if err == nil && len(presets) == 1 {
			return presets[0].OutputURL
		}
	}
	if mediaID != "" {
		mediaList, err := NewMedia().GetByIDs(ctx, []string{mediaID})
		if err == nil && len(mediaList) == 1 {
			return mediaList[0].URL
		}
	}
	return legacyURL
}

// DerefStr 空指针转空串。
func DerefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ResolveCoverURLs 批量解析业务模型的媒体/预设引用并回填 URL 列。
// ref 返回该行的 (mediaID, presetID)（空串表示未引用），setURL 回填解析出的完整 URL；
// 优先级：preset_id → media_id；引用在媒体库中不存在时保持原值（URL 列兜底）。
// 解析失败静默跳过，不影响业务读路径。
func ResolveCoverURLs[T any](ctx context.Context, items []*T, ref func(*T) (mediaID, presetID string), setURL func(*T, string)) {
	if len(items) == 0 {
		return
	}
	mediaIDs := make([]string, 0, len(items))
	presetIDs := make([]string, 0, len(items))
	for _, it := range items {
		mid, pid := ref(it)
		if mid != "" {
			mediaIDs = append(mediaIDs, mid)
		}
		if pid != "" {
			presetIDs = append(presetIDs, pid)
		}
	}
	if len(mediaIDs) == 0 && len(presetIDs) == 0 {
		return
	}

	mediaURLs, presetURLs, err := ResolveMediaURLMap(ctx, mediaIDs, presetIDs)
	if err != nil {
		return
	}

	for _, it := range items {
		mid, pid := ref(it)
		if pid != "" {
			if u, ok := presetURLs[pid]; ok && u != "" {
				setURL(it, u)
				continue
			}
		}
		if mid != "" {
			if u, ok := mediaURLs[mid]; ok && u != "" {
				setURL(it, u)
			}
		}
	}
}
