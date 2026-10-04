package res

// ImageSearchResult 图片搜索统一结果。
type ImageSearchResult struct {
	Name string `json:"name"` // 名称（图标为图标名、条目为条目名；无名称元数据的图源为空串）
	URL  string `json:"url"`  // 图片直链
}
