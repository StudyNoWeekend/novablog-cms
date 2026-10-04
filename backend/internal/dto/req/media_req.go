package req

// MediaListReq 媒体列表查询请求参数。
// FolderID 取值：不传=全部媒体（向后兼容）；"root"=根目录；其他=指定文件夹 ID。
type MediaListReq struct {
	PageReq
	FileType *int16  `form:"file_type" json:"file_type"`
	Keyword  *string `form:"keyword" json:"keyword"`
	FolderID *string `form:"folder_id" json:"folder_id"`
}

// CreatePresetReq 创建媒体预设请求参数（multipart 表单）。
type CreatePresetReq struct {
	MediaID       string `form:"media_id" json:"media_id" binding:"required"`
	Name          string `form:"name" json:"name" binding:"required"`
	FrameConfig   string `form:"frame_config" json:"frame_config" binding:"required"`
	DisplayParams string `form:"display_params" json:"display_params" binding:"required"`
}

// UploadWithPresetReq 上传原图并自动生成预设请求参数（multipart 表单）。
// Module/FolderID 决定原图归属的文件夹（FolderID 优先）。
type UploadWithPresetReq struct {
	Name     string `form:"name"`
	Module   string `form:"module"`
	FolderID string `form:"folder_id"`
}

// MediaFolderCreateReq 创建媒体文件夹请求参数。
type MediaFolderCreateReq struct {
	Name     string  `json:"name" binding:"required,max=50"`
	ParentID *string `json:"parent_id"` // 空=顶级文件夹
}

// MediaFolderUpdateReq 更新媒体文件夹请求参数（重命名和/或移动）。
// ParentID 传空字符串表示移动到根目录；不传该字段表示不改变父级。
type MediaFolderUpdateReq struct {
	Name     *string `json:"name" binding:"omitempty,max=50"`
	ParentID *string `json:"parent_id"`
}

// MediaMoveReq 批量移动媒体到指定文件夹请求参数。
type MediaMoveReq struct {
	MediaIDs []string `json:"media_ids" binding:"required,min=1,max=200"`
	FolderID *string  `json:"folder_id"` // 空=移到根目录
}
