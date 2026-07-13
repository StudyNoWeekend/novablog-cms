package controller

import (
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// APIDocParam API 参数定义。
type APIDocParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Desc     string `json:"desc"`
}

// APIDocField API 响应字段定义。
type APIDocField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Desc string `json:"desc"`
}

// APIDocItem 单个 API 定义。
type APIDocItem struct {
	Module      string        `json:"module"`
	Method      string        `json:"method"`
	Path        string        `json:"path"`
	Description string        `json:"description"`
	Params      []APIDocParam `json:"params"`
	Response    []APIDocField `json:"response"`
}

// apiDocs 所有公开 API 定义（路径相对于 /api/v1/public）。
var apiDocs = []APIDocItem{
	// 博主信息模块
	{
		Module:      "博主信息",
		Method:      "GET",
		Path:        "/blogger",
		Description: "获取博主公开信息",
		Params:      []APIDocParam{},
		Response: []APIDocField{
			{Name: "nickname", Type: "string", Desc: "昵称"},
			{Name: "avatar", Type: "string", Desc: "头像URL"},
			{Name: "bio", Type: "string", Desc: "个人简介"},
			{Name: "blog_title", Type: "string", Desc: "博客标题"},
			{Name: "blog_description", Type: "string", Desc: "博客描述"},
			{Name: "page_background", Type: "string", Desc: "页面背景图URL"},
			{Name: "blog_icon", Type: "string", Desc: "博客icon图URL"},
			{Name: "social_links", Type: "array", Desc: "社交平台链接数组 [{platform, url, sort_order}]"},
		},
	},

	// 文章模块
	{
		Module:      "文章",
		Method:      "GET",
		Path:        "/articles",
		Description: "获取已发布文章列表",
		Params: []APIDocParam{
			{Name: "page", Type: "int", Required: false, Desc: "页码默认1"},
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数默认20"},
			{Name: "category_id", Type: "string", Required: false, Desc: "分类ID筛选"},
			{Name: "keyword", Type: "string", Required: false, Desc: "标题关键词搜索"},
		},
		Response: []APIDocField{
			{Name: "list", Type: "array", Desc: "文章列表"},
			{Name: "total", Type: "int", Desc: "总数"},
			{Name: "page", Type: "int", Desc: "当前页"},
			{Name: "page_size", Type: "int", Desc: "每页条数"},
			{Name: "total_pages", Type: "int", Desc: "总页数"},
		},
	},
	{
		Module:      "文章",
		Method:      "GET",
		Path:        "/articles/:slug",
		Description: "获取文章详情",
		Params: []APIDocParam{
			{Name: "slug", Type: "string", Required: true, Desc: "文章URL标识"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "文章ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "slug", Type: "string", Desc: "URL标识"},
			{Name: "summary", Type: "string", Desc: "摘要"},
			{Name: "content", Type: "string", Desc: "正文"},
			{Name: "cover_image", Type: "string", Desc: "封面图"},
			{Name: "category_name", Type: "string", Desc: "分类名"},
			{Name: "tag_names", Type: "array", Desc: "标签名列表"},
			{Name: "view_count", Type: "int", Desc: "浏览量"},
			{Name: "comment_count", Type: "int", Desc: "评论数"},
			{Name: "published_at", Type: "string", Desc: "发布时间"},
		},
	},
	{
		Module:      "文章",
		Method:      "GET",
		Path:        "/articles/hot",
		Description: "获取热门文章",
		Params: []APIDocParam{
			{Name: "count", Type: "int", Required: false, Desc: "返回数量默认5"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "文章ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "slug", Type: "string", Desc: "URL标识"},
			{Name: "summary", Type: "string", Desc: "摘要"},
			{Name: "cover_image", Type: "string", Desc: "封面图"},
			{Name: "view_count", Type: "int", Desc: "浏览量"},
		},
	},
	{
		Module:      "文章",
		Method:      "GET",
		Path:        "/articles/random",
		Description: "获取随机文章推荐",
		Params: []APIDocParam{
			{Name: "count", Type: "int", Required: false, Desc: "返回数量默认5"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "文章ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "slug", Type: "string", Desc: "URL标识"},
			{Name: "summary", Type: "string", Desc: "摘要"},
			{Name: "cover_image", Type: "string", Desc: "封面图"},
		},
	},
	{
		Module:      "文章",
		Method:      "GET",
		Path:        "/articles/:slug/view",
		Description: "递增文章浏览量",
		Params: []APIDocParam{
			{Name: "slug", Type: "string", Required: true, Desc: "文章URL标识"},
		},
		Response: []APIDocField{},
	},
	{
		Module:      "文章",
		Method:      "GET",
		Path:        "/categories",
		Description: "获取分类列表",
		Params: []APIDocParam{
			{Name: "type", Type: "string", Required: false, Desc: "分类类型:article/travel/portfolio/music"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "分类ID"},
			{Name: "name", Type: "string", Desc: "分类名称"},
			{Name: "slug", Type: "string", Desc: "URL标识"},
			{Name: "description", Type: "string", Desc: "分类描述"},
			{Name: "type", Type: "string", Desc: "分类类型"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
		},
	},
	{
		Module:      "文章",
		Method:      "GET",
		Path:        "/tags",
		Description: "获取全部标签",
		Params:      []APIDocParam{},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "标签ID"},
			{Name: "name", Type: "string", Desc: "标签名称"},
		},
	},

	// 评论模块
	{
		Module:      "评论",
		Method:      "GET",
		Path:        "/comments",
		Description: "获取评论列表",
		Params: []APIDocParam{
			{Name: "page", Type: "int", Required: false, Desc: "页码"},
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数"},
			{Name: "target_type", Type: "string", Required: false, Desc: "目标类型:article/travel_guide"},
			{Name: "target_id", Type: "string", Required: false, Desc: "目标ID"},
		},
		Response: []APIDocField{
			{Name: "list", Type: "array", Desc: "评论列表"},
			{Name: "total", Type: "int", Desc: "总数"},
			{Name: "page", Type: "int", Desc: "当前页"},
			{Name: "page_size", Type: "int", Desc: "每页条数"},
			{Name: "total_pages", Type: "int", Desc: "总页数"},
		},
	},
	{
		Module:      "评论",
		Method:      "POST",
		Path:        "/comments",
		Description: "发表评论",
		Params: []APIDocParam{
			{Name: "target_type", Type: "string", Required: true, Desc: "目标类型"},
			{Name: "target_id", Type: "string", Required: true, Desc: "目标ID"},
			{Name: "parent_id", Type: "string", Required: false, Desc: "父评论ID"},
			{Name: "nickname", Type: "string", Required: true, Desc: "昵称"},
			{Name: "website", Type: "string", Required: false, Desc: "网站"},
			{Name: "content", Type: "string", Required: true, Desc: "评论内容"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "评论ID"},
			{Name: "nickname", Type: "string", Desc: "昵称"},
			{Name: "content", Type: "string", Desc: "内容"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
		},
	},

	// 旅行攻略模块
	{
		Module:      "旅行攻略",
		Method:      "GET",
		Path:        "/travels",
		Description: "获取已发布旅行攻略列表",
		Params: []APIDocParam{
			{Name: "page", Type: "int", Required: false, Desc: "页码"},
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数"},
			{Name: "keyword", Type: "string", Required: false, Desc: "关键词"},
			{Name: "region", Type: "string", Required: false, Desc: "地区"},
			{Name: "category_id", Type: "string", Required: false, Desc: "分类"},
			{Name: "days_range", Type: "string", Required: false, Desc: "天数范围:1-3/4-7/8-14/15+/all"},
			{Name: "sort", Type: "string", Required: false, Desc: "排序:views/rating/likes"},
		},
		Response: []APIDocField{
			{Name: "list", Type: "array", Desc: "攻略列表"},
			{Name: "total", Type: "int", Desc: "总数"},
			{Name: "page", Type: "int", Desc: "当前页"},
			{Name: "page_size", Type: "int", Desc: "每页条数"},
			{Name: "total_pages", Type: "int", Desc: "总页数"},
		},
	},
	{
		Module:      "旅行攻略",
		Method:      "GET",
		Path:        "/travels/:id",
		Description: "获取旅行攻略详情",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "攻略ID"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "攻略ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "summary", Type: "string", Desc: "摘要"},
			{Name: "cover_image", Type: "string", Desc: "封面图"},
			{Name: "destination", Type: "string", Desc: "目的地"},
			{Name: "region", Type: "string", Desc: "地区"},
			{Name: "days", Type: "int", Desc: "天数"},
			{Name: "best_month", Type: "string", Desc: "最佳月份"},
			{Name: "view_count", Type: "int", Desc: "浏览量"},
			{Name: "like_count", Type: "int", Desc: "点赞数"},
			{Name: "rating", Type: "number", Desc: "评分"},
			{Name: "attractions", Type: "array", Desc: "景点列表"},
			{Name: "itinerary", Type: "array", Desc: "行程安排"},
			{Name: "reviews", Type: "array", Desc: "评价列表"},
		},
	},
	{
		Module:      "旅行攻略",
		Method:      "GET",
		Path:        "/travels/hot",
		Description: "获取热门旅行攻略",
		Params: []APIDocParam{
			{Name: "count", Type: "int", Required: false, Desc: "返回数量默认5"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "攻略ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "summary", Type: "string", Desc: "摘要"},
			{Name: "cover_image", Type: "string", Desc: "封面图"},
			{Name: "view_count", Type: "int", Desc: "浏览量"},
		},
	},
	{
		Module:      "旅行攻略",
		Method:      "GET",
		Path:        "/travels/:id/view",
		Description: "递增攻略浏览量",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "攻略ID"},
		},
		Response: []APIDocField{},
	},
	{
		Module:      "旅行攻略",
		Method:      "POST",
		Path:        "/travels/:id/like",
		Description: "点赞攻略",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "攻略ID"},
		},
		Response: []APIDocField{},
	},

	// 摄影作品集模块
	{
		Module:      "摄影作品集",
		Method:      "GET",
		Path:        "/portfolios",
		Description: "获取已发布作品集列表",
		Params: []APIDocParam{
			{Name: "page", Type: "int", Required: false, Desc: "页码"},
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数"},
			{Name: "keyword", Type: "string", Required: false, Desc: "名称关键词"},
			{Name: "category_id", Type: "string", Required: false, Desc: "分类"},
		},
		Response: []APIDocField{
			{Name: "list", Type: "array", Desc: "作品集列表"},
			{Name: "total", Type: "int", Desc: "总数"},
			{Name: "page", Type: "int", Desc: "当前页"},
			{Name: "page_size", Type: "int", Desc: "每页条数"},
			{Name: "total_pages", Type: "int", Desc: "总页数"},
		},
	},
	{
		Module:      "摄影作品集",
		Method:      "GET",
		Path:        "/portfolios/:id",
		Description: "获取作品集详情",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "作品集ID"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "作品集ID"},
			{Name: "name", Type: "string", Desc: "名称"},
			{Name: "description", Type: "string", Desc: "描述"},
			{Name: "cover_url", Type: "string", Desc: "封面URL"},
			{Name: "items", Type: "array", Desc: "作品列表"},
		},
	},

	// 视频作品模块
	{
		Module:      "视频作品",
		Method:      "GET",
		Path:        "/videos",
		Description: "获取已发布视频列表",
		Params: []APIDocParam{
			{Name: "page", Type: "int", Required: false, Desc: "页码"},
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数"},
			{Name: "keyword", Type: "string", Required: false, Desc: "标题关键词"},
		},
		Response: []APIDocField{
			{Name: "list", Type: "array", Desc: "视频列表"},
			{Name: "total", Type: "int", Desc: "总数"},
			{Name: "page", Type: "int", Desc: "当前页"},
			{Name: "page_size", Type: "int", Desc: "每页条数"},
			{Name: "total_pages", Type: "int", Desc: "总页数"},
		},
	},
	{
		Module:      "视频作品",
		Method:      "GET",
		Path:        "/videos/:id",
		Description: "获取视频详情",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "视频ID"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "视频ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "cover_url", Type: "string", Desc: "封面URL"},
			{Name: "description", Type: "string", Desc: "描述"},
			{Name: "platforms", Type: "array", Desc: "平台列表({platform,url})"},
		},
	},

	// 音乐模块
	{
		Module:      "音乐",
		Method:      "GET",
		Path:        "/music/songs",
		Description: "获取歌曲列表",
		Params: []APIDocParam{
			{Name: "page", Type: "int", Required: false, Desc: "页码"},
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数"},
			{Name: "category_id", Type: "string", Required: false, Desc: "分类"},
		},
		Response: []APIDocField{
			{Name: "list", Type: "array", Desc: "歌曲列表"},
			{Name: "total", Type: "int", Desc: "总数"},
			{Name: "page", Type: "int", Desc: "当前页"},
			{Name: "page_size", Type: "int", Desc: "每页条数"},
			{Name: "total_pages", Type: "int", Desc: "总页数"},
		},
	},
	{
		Module:      "音乐",
		Method:      "GET",
		Path:        "/music/songs/:id",
		Description: "获取歌曲详情",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "歌曲ID"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "歌曲ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "artist", Type: "string", Desc: "艺术家"},
			{Name: "cover_url", Type: "string", Desc: "封面URL"},
			{Name: "duration", Type: "int", Desc: "时长"},
		},
	},
	{
		Module:      "音乐",
		Method:      "GET",
		Path:        "/music/audio-url/:song_id",
		Description: "获取音频播放地址",
		Params: []APIDocParam{
			{Name: "song_id", Type: "string", Required: true, Desc: "歌曲ID"},
		},
		Response: []APIDocField{
			{Name: "url", Type: "string", Desc: "音频播放地址"},
		},
	},
}

// APIDocController API 文档控制器结构体。
type APIDocController struct{}

// NewAPIDocController 创建 APIDocController 实例。
func NewAPIDocController() *APIDocController {
	return &APIDocController{}
}

// GetOpenAPIDocs 返回所有公开 API 的定义文档 GET /api/v1/api-docs
func (ctrl *APIDocController) GetOpenAPIDocs(c *gin.Context) {
	response.Success(c, apiDocs)
}
