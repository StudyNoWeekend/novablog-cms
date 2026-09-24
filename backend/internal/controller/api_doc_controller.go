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

// apiDocs 所有公开 API 定义（路径相对于 /api/v1/public，不含首次安装引导接口）。
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
			{Name: "email", Type: "string", Desc: "邮箱"},
			{Name: "city", Type: "string", Desc: "所在城市"},
			{Name: "blog_title", Type: "string", Desc: "博客标题"},
			{Name: "blog_description", Type: "string", Desc: "博客描述"},
			{Name: "page_background", Type: "string", Desc: "页面背景图URL"},
			{Name: "blog_icon", Type: "string", Desc: "博客icon图URL"},
			{Name: "social_links", Type: "array", Desc: "社交平台链接数组 [{platform,name,icon,color,url,sort_order}]"},
			{Name: "tags", Type: "array", Desc: "标签字符串数组"},
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
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数默认20，max=100"},
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
			{Name: "category_name", Type: "string", Desc: "分类名"},
			{Name: "tag_names", Type: "array", Desc: "标签名列表"},
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
			{Name: "category_name", Type: "string", Desc: "分类名"},
			{Name: "tag_names", Type: "array", Desc: "标签名列表"},
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
			{Name: "category_id", Type: "string", Desc: "分类ID，无分类时为空串"},
			{Name: "category_name", Type: "string", Desc: "分类名"},
			{Name: "tag_ids", Type: "array", Desc: "标签ID列表"},
			{Name: "tag_names", Type: "array", Desc: "标签名列表"},
			{Name: "is_top", Type: "bool", Desc: "是否置顶"},
			{Name: "is_comment", Type: "bool", Desc: "是否允许评论"},
			{Name: "view_count", Type: "int", Desc: "浏览量"},
			{Name: "comment_count", Type: "int", Desc: "评论数"},
			{Name: "published_at", Type: "string", Desc: "发布时间，未发布时为null"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
		},
	},
	{
		Module:      "文章",
		Method:      "GET",
		Path:        "/articles/:slug/view",
		Description: "递增文章浏览量（仅对已发布文章生效）",
		Params: []APIDocParam{
			{Name: "slug", Type: "string", Required: true, Desc: "文章URL标识"},
		},
		Response: []APIDocField{},
	},
	{
		Module:      "文章",
		Method:      "GET",
		Path:        "/categories",
		Description: "获取分类列表（固定返回 type=article 的分类）",
		Params:      []APIDocParam{},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "分类ID"},
			{Name: "name", Type: "string", Desc: "分类名称"},
			{Name: "slug", Type: "string", Desc: "URL标识"},
			{Name: "description", Type: "string", Desc: "分类描述"},
			{Name: "type", Type: "string", Desc: "分类类型（固定article）"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
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
			{Name: "created_at", Type: "string", Desc: "创建时间"},
		},
	},

	// 评论模块
	{
		Module:      "评论",
		Method:      "GET",
		Path:        "/comments",
		Description: "获取已通过审核的评论列表",
		Params: []APIDocParam{
			{Name: "page", Type: "int", Required: false, Desc: "页码"},
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数"},
			{Name: "target_type", Type: "string", Required: false, Desc: "目标类型:article/travel_guide"},
			{Name: "target_id", Type: "string", Required: false, Desc: "目标ID"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "评论ID"},
			{Name: "target_type", Type: "string", Desc: "目标类型"},
			{Name: "target_id", Type: "string", Desc: "目标ID"},
			{Name: "parent_id", Type: "string", Desc: "父评论ID，可为null"},
			{Name: "nickname", Type: "string", Desc: "评论者昵称"},
			{Name: "website", Type: "string", Desc: "评论者网站"},
			{Name: "content", Type: "string", Desc: "评论内容"},
			{Name: "is_blogger", Type: "bool", Desc: "是否为博主回复"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
		},
	},
	{
		Module:      "评论",
		Method:      "POST",
		Path:        "/comments",
		Description: "发表评论（访客）",
		Params: []APIDocParam{
			{Name: "target_type", Type: "string", Required: true, Desc: "目标类型:article/travel_guide"},
			{Name: "target_id", Type: "string", Required: true, Desc: "目标ID"},
			{Name: "parent_id", Type: "string", Required: false, Desc: "父评论ID（回复评论时传）"},
			{Name: "nickname", Type: "string", Required: true, Desc: "昵称，1-50字符"},
			{Name: "website", Type: "string", Required: false, Desc: "评论者博客地址"},
			{Name: "content", Type: "string", Required: true, Desc: "评论内容，1-2000字符"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "评论ID"},
			{Name: "target_type", Type: "string", Desc: "目标类型"},
			{Name: "target_id", Type: "string", Desc: "目标ID"},
			{Name: "parent_id", Type: "string", Desc: "父评论ID"},
			{Name: "nickname", Type: "string", Desc: "昵称"},
			{Name: "website", Type: "string", Desc: "网站"},
			{Name: "content", Type: "string", Desc: "内容"},
			{Name: "is_blogger", Type: "bool", Desc: "是否博主"},
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
			{Name: "keyword", Type: "string", Required: false, Desc: "模糊搜索标题/目的地/摘要"},
			{Name: "region", Type: "string", Required: false, Desc: "按地区筛选"},
			{Name: "category_id", Type: "string", Required: false, Desc: "按分类筛选"},
			{Name: "days_range", Type: "string", Required: false, Desc: "天数范围:1-3/4-7/8-14/15+/all"},
			{Name: "sort", Type: "string", Required: false, Desc: "排序:views/rating/likes"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "攻略ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "summary", Type: "string", Desc: "摘要"},
			{Name: "cover_image", Type: "string", Desc: "封面图"},
			{Name: "destination", Type: "string", Desc: "目的地"},
			{Name: "region", Type: "string", Desc: "地区"},
			{Name: "category_id", Type: "string", Desc: "分类ID，无分类时为空串"},
			{Name: "category_name", Type: "string", Desc: "分类名称"},
			{Name: "days", Type: "int", Desc: "天数"},
			{Name: "best_month", Type: "string", Desc: "最佳月份"},
			{Name: "view_count", Type: "int", Desc: "浏览量"},
			{Name: "like_count", Type: "int", Desc: "点赞数"},
			{Name: "rating", Type: "number", Desc: "评分"},
			{Name: "review_count", Type: "int", Desc: "评价数"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
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
			{Name: "destination", Type: "string", Desc: "目的地"},
			{Name: "region", Type: "string", Desc: "地区"},
			{Name: "view_count", Type: "int", Desc: "浏览量"},
		},
	},
	{
		Module:      "旅行攻略",
		Method:      "GET",
		Path:        "/travels/:id",
		Description: "获取已发布旅行攻略详情",
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
			{Name: "review_count", Type: "int", Desc: "评价数"},
			{Name: "attractions", Type: "array", Desc: "景点列表"},
			{Name: "itinerary", Type: "array", Desc: "行程安排"},
			{Name: "reviews", Type: "array", Desc: "评价列表"},
		},
	},
	{
		Module:      "旅行攻略",
		Method:      "GET",
		Path:        "/travels/:id/view",
		Description: "递增攻略浏览量（仅对已发布攻略生效）",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "攻略ID"},
		},
		Response: []APIDocField{},
	},
	{
		Module:      "旅行攻略",
		Method:      "POST",
		Path:        "/travels/:id/like",
		Description: "点赞旅行攻略（仅对已发布攻略生效）",
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
			{Name: "category_id", Type: "string", Required: false, Desc: "按分类筛选"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "作品集ID"},
			{Name: "name", Type: "string", Desc: "名称"},
			{Name: "description", Type: "string", Desc: "描述"},
			{Name: "cover_mode", Type: "int", Desc: "封面模式:0=首图,1=独立封面"},
			{Name: "cover_preset_id", Type: "string", Desc: "独立封面预设ID"},
			{Name: "cover_url", Type: "string", Desc: "封面图URL"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "category_id", Type: "string", Desc: "分类ID"},
			{Name: "category_name", Type: "string", Desc: "分类名称"},
			{Name: "item_count", Type: "int", Desc: "作品项数量"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
		},
	},
	{
		Module:      "摄影作品集",
		Method:      "GET",
		Path:        "/portfolios/:id",
		Description: "获取已发布作品集详情（含作品项列表）",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "作品集ID"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "作品集ID"},
			{Name: "name", Type: "string", Desc: "名称"},
			{Name: "description", Type: "string", Desc: "描述"},
			{Name: "cover_mode", Type: "int", Desc: "封面模式:0=首图,1=独立封面"},
			{Name: "cover_url", Type: "string", Desc: "封面URL"},
			{Name: "items", Type: "array", Desc: "作品项列表 [{id,preset_id,title,output_url,mime_type,output_size}]"},
		},
	},

	// 个人设备模块
	{
		Module:      "个人设备",
		Method:      "GET",
		Path:        "/equipments",
		Description: "获取摄影器材列表",
		Params: []APIDocParam{
			{Name: "page", Type: "int", Required: false, Desc: "页码"},
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数，max=100"},
			{Name: "keyword", Type: "string", Required: false, Desc: "模糊搜索器材名称"},
			{Name: "brand", Type: "string", Required: false, Desc: "按品牌筛选"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "器材ID"},
			{Name: "name", Type: "string", Desc: "器材名称"},
			{Name: "image_url", Type: "string", Desc: "器材图片URL"},
			{Name: "brand", Type: "string", Desc: "器材品牌"},
			{Name: "description", Type: "string", Desc: "器材介绍"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
		},
	},
	{
		Module:      "个人设备",
		Method:      "GET",
		Path:        "/equipments/:id",
		Description: "获取摄影器材详情",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "器材ID"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "器材ID"},
			{Name: "name", Type: "string", Desc: "器材名称"},
			{Name: "image_url", Type: "string", Desc: "器材图片URL"},
			{Name: "brand", Type: "string", Desc: "器材品牌"},
			{Name: "description", Type: "string", Desc: "器材介绍"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
		},
	},

	// 项目经历模块
	{
		Module:      "项目经历",
		Method:      "GET",
		Path:        "/projects",
		Description: "获取已发布项目经历列表",
		Params: []APIDocParam{
			{Name: "page", Type: "int", Required: false, Desc: "页码"},
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数"},
			{Name: "keyword", Type: "string", Required: false, Desc: "标题/简介关键词"},
			{Name: "category", Type: "string", Required: false, Desc: "领域分类：摄影/视频剪辑/技术开发等"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "项目ID"},
			{Name: "title", Type: "string", Desc: "项目名称"},
			{Name: "category", Type: "string", Desc: "领域分类"},
			{Name: "role", Type: "string", Desc: "担任角色"},
			{Name: "client", Type: "string", Desc: "客户/所属组织"},
			{Name: "cover_url", Type: "string", Desc: "封面URL"},
			{Name: "summary", Type: "string", Desc: "一句话简介"},
			{Name: "description", Type: "string", Desc: "详细描述"},
			{Name: "tech_stack", Type: "string", Desc: "技能/工具标签，逗号分隔"},
			{Name: "start_date", Type: "string", Desc: "开始时间（YYYY-MM）"},
			{Name: "end_date", Type: "string", Desc: "结束时间（YYYY-MM，空表示至今）"},
			{Name: "project_url", Type: "string", Desc: "项目/作品链接"},
			{Name: "repo_url", Type: "string", Desc: "代码仓库链接"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
		},
	},
	{
		Module:      "项目经历",
		Method:      "GET",
		Path:        "/projects/:id",
		Description: "获取已发布项目经历详情",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "项目经历ID"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "项目ID"},
			{Name: "title", Type: "string", Desc: "项目名称"},
			{Name: "category", Type: "string", Desc: "领域分类"},
			{Name: "role", Type: "string", Desc: "担任角色"},
			{Name: "client", Type: "string", Desc: "客户/所属组织"},
			{Name: "cover_url", Type: "string", Desc: "封面URL"},
			{Name: "summary", Type: "string", Desc: "一句话简介"},
			{Name: "description", Type: "string", Desc: "详细描述"},
			{Name: "tech_stack", Type: "string", Desc: "技能/工具标签，逗号分隔"},
			{Name: "start_date", Type: "string", Desc: "开始时间（YYYY-MM）"},
			{Name: "end_date", Type: "string", Desc: "结束时间（YYYY-MM，空表示至今）"},
			{Name: "project_url", Type: "string", Desc: "项目/作品链接"},
			{Name: "repo_url", Type: "string", Desc: "代码仓库链接"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
		},
	},

	// 开源作品模块
	{
		Module:      "开源作品",
		Method:      "GET",
		Path:        "/open-sources",
		Description: "获取已发布开源作品列表（按排序权重、Star 数排序，不含 README 正文）",
		Params: []APIDocParam{
			{Name: "page", Type: "int", Required: false, Desc: "页码"},
			{Name: "page_size", Type: "int", Required: false, Desc: "每页条数"},
			{Name: "keyword", Type: "string", Required: false, Desc: "仓库名称/简介关键词"},
		},
		Response: []APIDocField{
			{Name: "list", Type: "array", Desc: "开源作品列表（元素含 id/name/repo_url/summary/language/topics/stars/homepage）"},
			{Name: "total", Type: "int", Desc: "总数"},
			{Name: "page", Type: "int", Desc: "当前页"},
			{Name: "page_size", Type: "int", Desc: "每页条数"},
			{Name: "total_pages", Type: "int", Desc: "总页数"},
		},
	},
	{
		Module:      "开源作品",
		Method:      "GET",
		Path:        "/open-sources/:id",
		Description: "获取已发布开源作品详情（含自动拉取的 README 原文，Markdown 格式）",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "开源作品ID"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "开源作品ID"},
			{Name: "name", Type: "string", Desc: "仓库名称"},
			{Name: "repo_url", Type: "string", Desc: "仓库链接"},
			{Name: "summary", Type: "string", Desc: "一句话介绍（为空时自动回填仓库描述）"},
			{Name: "readme", Type: "string", Desc: "README 原文（Markdown，自动从仓库拉取）"},
			{Name: "language", Type: "string", Desc: "主语言"},
			{Name: "topics", Type: "string", Desc: "主题标签，逗号分隔"},
			{Name: "stars", Type: "int", Desc: "Star 数（同步时快照）"},
			{Name: "homepage", Type: "string", Desc: "主页/演示地址"},
			{Name: "readme_updated_at", Type: "string", Desc: "README 最近拉取时间"},
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
			{Name: "id", Type: "string", Desc: "视频ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "cover_url", Type: "string", Desc: "封面URL"},
			{Name: "description", Type: "string", Desc: "描述"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "platforms", Type: "array", Desc: "平台链接列表 [{id,platform,url}]"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
		},
	},
	{
		Module:      "视频作品",
		Method:      "GET",
		Path:        "/videos/:id",
		Description: "获取已发布视频详情（含平台链接）",
		Params: []APIDocParam{
			{Name: "id", Type: "string", Required: true, Desc: "视频ID"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "视频ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "cover_url", Type: "string", Desc: "封面URL"},
			{Name: "description", Type: "string", Desc: "描述"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "platforms", Type: "array", Desc: "平台链接列表 [{id,platform,url}]"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
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
			{Name: "category_id", Type: "string", Required: false, Desc: "按分类筛选"},
		},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "歌曲ID"},
			{Name: "title", Type: "string", Desc: "标题"},
			{Name: "artist", Type: "string", Desc: "艺术家"},
			{Name: "cover_url", Type: "string", Desc: "封面URL"},
			{Name: "bvid", Type: "string", Desc: "B站视频ID"},
			{Name: "cid", Type: "int", Desc: "B站内容ID"},
			{Name: "source_url", Type: "string", Desc: "来源URL"},
			{Name: "source_type", Type: "string", Desc: "来源类型"},
			{Name: "category_id", Type: "string", Desc: "分类ID"},
			{Name: "duration", Type: "int", Desc: "时长（秒）"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
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
			{Name: "source_url", Type: "string", Desc: "来源URL"},
			{Name: "source_type", Type: "string", Desc: "来源类型"},
			{Name: "duration", Type: "int", Desc: "时长（秒）"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
		},
	},
	{
		Module:      "音乐",
		Method:      "GET",
		Path:        "/music/audio-url/:song_id",
		Description: "获取歌曲播放地址（B站官方外链播放器，iframe 内嵌播放）",
		Params: []APIDocParam{
			{Name: "song_id", Type: "string", Required: true, Desc: "歌曲ID"},
		},
		Response: []APIDocField{
			{Name: "url", Type: "string", Desc: "B站外链播放器地址，供 iframe.src 使用"},
		},
	},
	{
		Module:      "音乐",
		Method:      "GET",
		Path:        "/music/playlists",
		Description: "获取前台展示的第三方歌单列表（与 /playlists 同处理函数）",
		Params:      []APIDocParam{},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "歌单ID"},
			{Name: "title", Type: "string", Desc: "歌单标题"},
			{Name: "cover_url", Type: "string", Desc: "歌单封面URL"},
			{Name: "platform", Type: "string", Desc: "平台名称"},
			{Name: "platform_url", Type: "string", Desc: "平台链接URL"},
			{Name: "description", Type: "string", Desc: "描述"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "enabled", Type: "bool", Desc: "是否启用（仅返回true）"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
		},
	},

	// 系统模块
	{
		Module:      "系统",
		Method:      "GET",
		Path:        "/playlists",
		Description: "获取前台展示的第三方歌单列表",
		Params:      []APIDocParam{},
		Response: []APIDocField{
			{Name: "id", Type: "string", Desc: "歌单ID"},
			{Name: "title", Type: "string", Desc: "歌单标题"},
			{Name: "cover_url", Type: "string", Desc: "歌单封面URL"},
			{Name: "platform", Type: "string", Desc: "平台名称"},
			{Name: "platform_url", Type: "string", Desc: "平台链接URL"},
			{Name: "description", Type: "string", Desc: "描述"},
			{Name: "sort_order", Type: "int", Desc: "排序"},
			{Name: "enabled", Type: "bool", Desc: "是否启用（仅返回true）"},
			{Name: "created_at", Type: "string", Desc: "创建时间"},
			{Name: "updated_at", Type: "string", Desc: "更新时间"},
		},
	},
	{
		Module:      "系统",
		Method:      "GET",
		Path:        "/config",
		Description: "获取公共配置（官方主题市场地址，免鉴权下发）",
		Params:      []APIDocParam{},
		Response: []APIDocField{
			{Name: "market_base_url", Type: "string", Desc: "官方主题市场地址"},
		},
	},
	{
		Module:      "系统",
		Method:      "GET",
		Path:        "/module-config",
		Description: "获取模块开关配置",
		Params:      []APIDocParam{},
		Response: []APIDocField{
			{Name: "article_enabled", Type: "bool", Desc: "文章管理模块是否开启"},
			{Name: "media_enabled", Type: "bool", Desc: "媒体管理模块是否开启"},
			{Name: "music_enabled", Type: "bool", Desc: "音乐管理模块是否开启"},
			{Name: "video_enabled", Type: "bool", Desc: "视频管理模块是否开启"},
			{Name: "travel_enabled", Type: "bool", Desc: "旅行管理模块是否开启"},
			{Name: "portfolio_enabled", Type: "bool", Desc: "作品集管理模块是否开启"},
			{Name: "equipment_enabled", Type: "bool", Desc: "设备管理模块是否开启"},
			{Name: "project_enabled", Type: "bool", Desc: "项目经历管理模块是否开启"},
			{Name: "open_source_enabled", Type: "bool", Desc: "开源作品模块是否开启"},
			{Name: "updated_at", Type: "string", Desc: "配置最后更新时间"},
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
