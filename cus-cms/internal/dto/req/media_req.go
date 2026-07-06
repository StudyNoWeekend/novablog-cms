package req

// MediaListReq 媒体列表查询请求参数。
type MediaListReq struct {
	PageReq
	FileType *int16  `form:"file_type" json:"file_type"`
	Keyword  *string `form:"keyword" json:"keyword"`
}

// CreatePresetReq 创建媒体预设请求参数（multipart 表单）。
type CreatePresetReq struct {
	MediaID       string `form:"media_id" json:"media_id" binding:"required"`
	Name          string `form:"name" json:"name" binding:"required"`
	FrameConfig   string `form:"frame_config" json:"frame_config" binding:"required"`
	DisplayParams string `form:"display_params" json:"display_params" binding:"required"`
}
