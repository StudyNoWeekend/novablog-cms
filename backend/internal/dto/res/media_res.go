package res

import "time"

// MediaRes 媒体响应结构体。
type MediaRes struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	FileType  int16     `json:"file_type"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	URL       string    `json:"url"`
	ThumbURL  string    `json:"thumb_url"`
	Width     *int      `json:"width"`
	Height    *int      `json:"height"`
	CreatedAt time.Time `json:"created_at"`
}

// MediaListRes 媒体列表响应结构体。
type MediaListRes struct {
	List       []MediaRes `json:"list"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// MediaPresetRes 媒体预设响应结构体。
type MediaPresetRes struct {
	ID                string    `json:"id"`
	MediaID           string    `json:"media_id"`
	Name              string    `json:"name"`
	FrameConfig       string    `json:"frame_config"`
	DisplayParams     string    `json:"display_params"`
	OutputURL         string    `json:"output_url"`
	OutputStoragePath string    `json:"output_storage_path"`
	OutputSize        int64     `json:"output_size"`
	MimeType          string    `json:"mime_type"`
	CreatedAt         time.Time `json:"created_at"`
}

// MediaPresetListRes 媒体预设列表响应结构体。
type MediaPresetListRes struct {
	List []MediaPresetRes `json:"list"`
}

// UploadWithPresetRes 上传原图并自动生成预设响应结构体。
type UploadWithPresetRes struct {
	Media  MediaRes       `json:"media"`
	Preset MediaPresetRes `json:"preset"`
}
