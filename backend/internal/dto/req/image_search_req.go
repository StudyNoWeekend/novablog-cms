package req

// ImageSearchReq 图片搜索请求参数。
type ImageSearchReq struct {
	Type  string `form:"type" binding:"required,oneof=icons games books food"` // 图源类型
	Q     string `form:"q" binding:"required,min=1,max=100"`                   // 搜索关键词
	Limit int    `form:"limit"`                                                // 返回条数，缺省按图源默认值
}

// SaveImageReq 外部图片转存请求参数。
type SaveImageReq struct {
	URL    string `json:"url" binding:"required,max=1024"`   // 外部图片直链
	Module string `json:"module" binding:"omitempty,max=50"` // 归属业务模块 key（转存后归入对应模块文件夹）
}

// ImageThumbReq 图片缩略图代理请求参数。
type ImageThumbReq struct {
	URL string `form:"url" binding:"required,max=1024"` // 外部图片直链
}
