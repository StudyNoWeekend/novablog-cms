package logic

import (
	"context"
	"fmt"

	"novablog/internal/model"
)

// normalizeMediaRef 保存时的媒体引用/外链互斥处理（字段级）：
//   - mediaID 非空：校验媒体存在（附 presetID 时校验预设归属该媒体），返回引用 ID，
//     URL 返回空串（落库清空，读取时按 ID 动态解析最新链接）；
//   - mediaID 为空：URL 原样返回（第三方外链/历史模式），引用返回空串（落库清空）。
func normalizeMediaRef(ctx context.Context, mediaID, presetID, url string) (string, string, string, error) {
	if mediaID == "" {
		return "", "", url, nil
	}
	media, err := model.NewMedia().GetByID(ctx, mediaID)
	if err != nil {
		return "", "", "", fmt.Errorf("媒体文件不存在")
	}
	if presetID != "" {
		preset, err := model.NewMediaPreset().GetByID(ctx, presetID)
		if err != nil || preset.MediaID != media.ID {
			return "", "", "", fmt.Errorf("预设不存在或不属于该媒体文件")
		}
		return mediaID, presetID, "", nil
	}
	return mediaID, "", "", nil
}

// mediaRefPtr 空串转 nil（引用列落库置空），非空转指针。
func mediaRefPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

// refStr 空指针转空串（logic 包别名，与 model.DerefStr 等价）。
func refStr(p *string) string {
	return model.DerefStr(p)
}
