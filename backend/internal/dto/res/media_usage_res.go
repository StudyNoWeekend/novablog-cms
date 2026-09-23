package res

// MediaUsageItem 媒体引用条目响应。
type MediaUsageItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Field string `json:"field"`
}

// MediaUsageGroup 按内容模块聚合的媒体引用响应。
type MediaUsageGroup struct {
	Module string           `json:"module"`
	Items  []MediaUsageItem `json:"items"`
}

// MediaUsageRes 媒体引用扫描响应。
type MediaUsageRes struct {
	MediaID string            `json:"media_id"`
	Used    bool              `json:"used"`
	Total   int               `json:"total"`
	Groups  []MediaUsageGroup `json:"groups"`
}
