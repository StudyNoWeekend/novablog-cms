package controller

import (
	"context"
	"errors"
	"strconv"
	"time"

	"novablog/enum"
	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/internal/storage"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PublicController 公开接口控制器结构体。
type PublicController struct {
	setupLogic             *logic.SetupLogic
	articleLogic           *logic.ArticleLogic
	categoryLogic          *logic.CategoryLogic
	tagLogic               *logic.TagLogic
	bloggerLogic           *logic.BloggerLogic
	commentLogic           *logic.CommentLogic
	travelLogic            *logic.TravelGuideLogic
	portfolioLogic         *logic.PortfolioLogic
	videoLogic             *logic.VideoLogic
	musicLogic             *logic.MusicLogic
	equipmentLogic         *logic.EquipmentLogic
	projectLogic           *logic.ProjectLogic
	openSourceLogic        *logic.OpenSourceLogic
	moduleConfigLogic      *logic.ModuleConfigLogic
	playlistLogic          *logic.ThirdPartyPlaylistLogic
	themeMarketConfigLogic *logic.ThemeMarketConfigLogic
}

// NewPublicController 创建 PublicController 实例。
func NewPublicController(manager *storage.Manager, cryptoKey string) *PublicController {
	return &PublicController{
		setupLogic:             logic.NewSetupLogic(manager, cryptoKey),
		articleLogic:           logic.NewArticleLogic(),
		categoryLogic:          logic.NewCategoryLogic(),
		tagLogic:               logic.NewTagLogic(),
		bloggerLogic:           logic.NewBloggerLogic(),
		commentLogic:           logic.NewCommentLogic(),
		travelLogic:            logic.NewTravelGuideLogic(),
		portfolioLogic:         logic.NewPortfolioLogic(),
		videoLogic:             logic.NewVideoLogic(),
		musicLogic:             logic.NewMusicLogic(manager),
		equipmentLogic:         logic.NewEquipmentLogic(),
		projectLogic:           logic.NewProjectLogic(),
		openSourceLogic:        logic.NewOpenSourceLogic(),
		moduleConfigLogic:      logic.NewModuleConfigLogic(),
		playlistLogic:          logic.NewThirdPartyPlaylistLogic(),
		themeMarketConfigLogic: logic.NewThemeMarketConfigLogic(),
	}
}

// GetStatus 获取系统初始化状态。
func (ctrl *PublicController) GetStatus(c *gin.Context) {
	statusRes, err := ctrl.setupLogic.CheckStatus(c.Request.Context())
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, statusRes)
}

// GetPublicConfig 公共配置下发（免鉴权）：官方市场地址默认值由后端控制。
func (ctrl *PublicController) GetPublicConfig(c *gin.Context) {
	config, err := ctrl.themeMarketConfigLogic.GetPublicConfig(c.Request.Context())
	if err != nil {
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}
	response.Success(c, config)
}

// Init 初始化博主账号。
func (ctrl *PublicController) Init(c *gin.Context) {
	var initReq req.InitReq
	if err := c.ShouldBindJSON(&initReq); err != nil {
		logic.SetupLogger.Warn("初始化请求参数错误", zap.Error(err))
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	initRes, err := ctrl.setupLogic.InitBlogger(c.Request.Context(), &initReq)
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, initRes)
}

// InitThemeReq 首装主题安装请求。
type InitThemeReq struct {
	MarketBaseURL  string `json:"market_base_url" binding:"omitempty"` // 官方市场地址（向导输入覆盖 config）
	MarketEmail    string `json:"market_email" binding:"omitempty"`    // 官方市场账号（官方代理下载要求登录态）
	MarketPassword string `json:"market_password" binding:"omitempty"` // 官方市场密码（仅本次服务端登录使用，不落库）
}

// InitTheme 首装初始化博客外观：拉取官方默认主题并激活 POST /api/v1/public/install/theme
func (ctrl *PublicController) InitTheme(c *gin.Context) {
	var r InitThemeReq
	_ = c.ShouldBindJSON(&r)
	result, err := ctrl.setupLogic.StartThemeInstall(c.Request.Context(), r.MarketBaseURL, r.MarketEmail, r.MarketPassword)
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}
	response.Success(c, result)
}

// GetThemeInstallStatus 查询主题安装任务状态 GET /api/v1/public/install/theme/status
func (ctrl *PublicController) GetThemeInstallStatus(c *gin.Context) {
	result, err := ctrl.setupLogic.GetThemeInstallStatus(c.Request.Context())
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}
	response.Success(c, result)
}

// SetupStorage 首装向导存储配置 POST /api/v1/public/install/storage
// provider 为空或 "local" 表示使用本地存储，非空时创建云存储配置并激活。
func (ctrl *PublicController) SetupStorage(c *gin.Context) {
	var r req.SetupStorageReq
	if err := c.ShouldBindJSON(&r); err != nil {
		logic.SetupLogger.Warn("存储配置请求参数错误", zap.Error(err))
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	result, err := ctrl.setupLogic.SetupStorage(c.Request.Context(), &r)
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, result)
}

// GetArticles 公开文章列表 GET /api/v1/public/articles
func (ctrl *PublicController) GetArticles(ctx *gin.Context) {
	var r req.ArticleListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	// 公开接口只返回已发布文章
	published := int16(2)
	r.Status = &published
	result, err := ctrl.articleLogic.GetList(ctx, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetArticleBySlug 根据 slug 获取文章详情 GET /api/v1/public/articles/:slug
func (ctrl *PublicController) GetArticleBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")
	result, err := ctrl.articleLogic.GetBySlug(ctx, slug)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	// 公开接口只返回已发布文章
	if result.Status != 2 {
		response.Fail(ctx, enum.ErrNotFound.Code, enum.ErrNotFound.Msg, enum.ErrNotFound.HttpCode)
		return
	}
	// 详情请求顺带累计浏览量（同 IP 在去重窗口内只计一次；部分旧版主题未调用独立上报接口，靠此兜底）
	ctrl.recordArticleView(ctx, result.ID, slug, ctx.ClientIP())
	response.Success(ctx, result)
}

// GetCategories 公开分类列表 GET /api/v1/public/categories
func (ctrl *PublicController) GetCategories(ctx *gin.Context) {
	result, err := ctrl.categoryLogic.GetAll(ctx, "article")
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetTags 公开标签列表 GET /api/v1/public/tags
func (ctrl *PublicController) GetTags(ctx *gin.Context) {
	result, err := ctrl.tagLogic.GetAll(ctx)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetBlogger 获取博主公开信息 GET /api/v1/public/blogger
func (ctrl *PublicController) GetBlogger(ctx *gin.Context) {
	result, err := ctrl.bloggerLogic.GetPublicInfo(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetHotArticles 获取热门文章列表 GET /api/v1/public/articles/hot
func (ctrl *PublicController) GetHotArticles(ctx *gin.Context) {
	var r req.HotArticleReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	count := r.Count
	if count <= 0 {
		count = 5
	}
	result, err := ctrl.articleLogic.GetHotList(ctx.Request.Context(), count)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetRandomArticles 获取随机文章列表 GET /api/v1/public/articles/random
func (ctrl *PublicController) GetRandomArticles(ctx *gin.Context) {
	var r req.RandomArticleReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	count := r.Count
	if count <= 0 {
		count = 5
	}
	result, err := ctrl.articleLogic.GetRandomList(ctx.Request.Context(), count)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// IncrementArticleView 增加文章浏览量 GET /api/v1/public/articles/:slug/view
// 与详情接口共用同一套 IP 去重字典，避免新版主题"详情+上报"双调用导致重复计数。
func (ctrl *PublicController) IncrementArticleView(ctx *gin.Context) {
	slug := ctx.Param("slug")
	// 先解析文章 ID，便于去重键与详情接口一致
	article, err := ctrl.articleLogic.GetBySlug(ctx.Request.Context(), slug)
	if err != nil || article == nil || article.Status != 2 {
		// 文章不存在或未发布：静默，保持与原行为一致（不计数）
		response.Success(ctx, nil)
		return
	}
	ctrl.recordArticleView(ctx, article.ID, slug, ctx.ClientIP())
	response.Success(ctx, nil)
}

// GetComments 获取公开评论列表 GET /api/v1/public/comments
func (ctrl *PublicController) GetComments(ctx *gin.Context) {
	var r req.CommentListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	targetType := ""
	if r.TargetType != nil {
		targetType = *r.TargetType
	}
	targetID := ""
	if r.TargetID != nil {
		targetID = *r.TargetID
	}
	result, err := ctrl.commentLogic.GetPublicList(ctx.Request.Context(), targetType, targetID, r.GetPage(), r.GetPageSize())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// CreateComment 创建公开评论 POST /api/v1/public/comments
func (ctrl *PublicController) CreateComment(ctx *gin.Context) {
	var r req.CreatePublicCommentReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	ip := ctx.ClientIP()
	result, err := ctrl.commentLogic.CreatePublic(ctx.Request.Context(), &r, ip)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetTravels 获取已发布旅行攻略列表 GET /api/v1/public/travels
func (ctrl *PublicController) GetTravels(ctx *gin.Context) {
	var r req.TravelGuideListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := ctrl.travelLogic.GetPublicList(ctx.Request.Context(), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetHotTravels 获取热门旅行攻略列表 GET /api/v1/public/travels/hot
func (ctrl *PublicController) GetHotTravels(ctx *gin.Context) {
	count := 5
	if countStr := ctx.Query("count"); countStr != "" {
		if n, err := strconv.Atoi(countStr); err == nil && n > 0 {
			count = n
		}
	}
	result, err := ctrl.travelLogic.GetHotList(ctx.Request.Context(), count)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetTravelDetail 获取已发布旅行攻略详情 GET /api/v1/public/travels/:id
func (ctrl *PublicController) GetTravelDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.travelLogic.GetPublicDetail(ctx.Request.Context(), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	// 详情请求顺带累计浏览量（同 IP 在去重窗口内只计一次；部分旧版主题未调用独立上报接口，靠此兜底）
	ctrl.recordTravelView(ctx, result.ID, ctx.ClientIP())
	response.Success(ctx, result)
}

// IncrementTravelView 增加旅行攻略浏览量 GET /api/v1/public/travels/:id/view
// 与详情接口共用同一套 IP 去重字典，避免新版主题"详情+上报"双调用导致重复计数。
func (ctrl *PublicController) IncrementTravelView(ctx *gin.Context) {
	id := ctx.Param("id")
	ctrl.recordTravelView(ctx, id, ctx.ClientIP())
	response.Success(ctx, nil)
}

// LikeTravel 增加旅行攻略点赞数 POST /api/v1/public/travels/:id/like
func (ctrl *PublicController) LikeTravel(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := ctrl.travelLogic.IncrementLike(ctx.Request.Context(), id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// GetPortfolios 获取已发布作品集列表 GET /api/v1/public/portfolios
func (ctrl *PublicController) GetPortfolios(ctx *gin.Context) {
	var r req.PortfolioListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := ctrl.portfolioLogic.GetPublicList(ctx.Request.Context(), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetPortfolioDetail 获取已发布作品集详情 GET /api/v1/public/portfolios/:id
func (ctrl *PublicController) GetPortfolioDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.portfolioLogic.GetPublicDetail(ctx.Request.Context(), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetVideos 获取已发布视频作品列表 GET /api/v1/public/videos
func (ctrl *PublicController) GetVideos(ctx *gin.Context) {
	var r req.VideoListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := ctrl.videoLogic.GetPublicList(ctx.Request.Context(), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetVideoDetail 获取已发布视频作品详情 GET /api/v1/public/videos/:id
func (ctrl *PublicController) GetVideoDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.videoLogic.GetPublicDetail(ctx.Request.Context(), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetSongs 获取歌曲列表 GET /api/v1/public/music/songs
func (ctrl *PublicController) GetSongs(ctx *gin.Context) {
	var r req.SongListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := ctrl.musicLogic.GetPublicSongList(ctx.Request.Context(), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetSongDetail 获取歌曲详情 GET /api/v1/public/music/songs/:id
func (ctrl *PublicController) GetSongDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.musicLogic.GetPublicSongByID(ctx.Request.Context(), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetAudioURL 获取播放地址（B站外链播放器） GET /api/v1/public/music/audio-url/:song_id
func (ctrl *PublicController) GetAudioURL(ctx *gin.Context) {
	songID := ctx.Param("song_id")
	url, err := ctrl.musicLogic.GetPublicAudioURL(ctx.Request.Context(), songID)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, gin.H{"url": url})
}

// GetEquipments 获取个人设备列表 GET /api/v1/public/equipments
func (ctrl *PublicController) GetEquipments(ctx *gin.Context) {
	var r req.EquipmentListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := ctrl.equipmentLogic.GetList(ctx.Request.Context(), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetEquipmentDetail 获取个人设备详情 GET /api/v1/public/equipments/:id
func (ctrl *PublicController) GetEquipmentDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.equipmentLogic.GetByID(ctx.Request.Context(), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetProjects 获取已发布项目经历列表 GET /api/v1/public/projects
func (ctrl *PublicController) GetProjects(ctx *gin.Context) {
	var r req.ProjectListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := ctrl.projectLogic.GetPublicList(ctx.Request.Context(), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetProjectDetail 获取已发布项目经历详情 GET /api/v1/public/projects/:id
func (ctrl *PublicController) GetProjectDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.projectLogic.GetPublicDetail(ctx.Request.Context(), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetOpenSources 获取已发布开源作品列表 GET /api/v1/public/open-sources
func (ctrl *PublicController) GetOpenSources(ctx *gin.Context) {
	var r req.OpenSourceListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := ctrl.openSourceLogic.GetPublicList(ctx.Request.Context(), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetOpenSourceDetail 获取已发布开源作品详情（含 README） GET /api/v1/public/open-sources/:id
func (ctrl *PublicController) GetOpenSourceDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.openSourceLogic.GetPublicDetail(ctx.Request.Context(), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetModuleConfig 获取模块开关配置 GET /api/v1/public/module-config
func (ctrl *PublicController) GetModuleConfig(ctx *gin.Context) {
	config, err := ctrl.moduleConfigLogic.GetConfig(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, config)
}

// GetPlaylists 获取前台展示的第三方歌单列表 GET /api/v1/public/playlists
func (ctrl *PublicController) GetPlaylists(ctx *gin.Context) {
	result, err := ctrl.playlistLogic.GetPublicList(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// viewDedupWindow 浏览去重窗口：同一 IP 对同一内容在窗口内只计一次浏览。
const viewDedupWindow = 5 * time.Minute

// recordArticleView 记录文章浏览：依赖 Redis SETNX 去重，Redis 不可用时降级为直接计数（不阻塞响应）。
func (ctrl *PublicController) recordArticleView(ctx *gin.Context, articleID, slug, ip string) {
	doIncrement := func() {
		if err := ctrl.articleLogic.IncrementViewByID(ctx.Request.Context(), articleID); err != nil {
			// 计数失败不阻塞响应，浏览统计可容忍少量丢失
			return
		}
	}

	if cache.RedisClient == nil {
		// 未配置 Redis：直接计数
		doIncrement()
		return
	}

	key := "view:article:" + articleID + ":" + ip
	dedupCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
	defer cancel()
	ok, err := cache.RedisClient.SetNX(dedupCtx, key, "1", viewDedupWindow).Result()
	if err != nil {
		// Redis 异常：降级为直接计数，保证浏览量仍能累计
		doIncrement()
		return
	}
	if !ok {
		return // 去重窗口内已计过，跳过
	}
	doIncrement()
}

// recordTravelView 记录旅行攻略浏览：依赖 Redis SETNX 去重，Redis 不可用时降级为直接计数（不阻塞响应）。
func (ctrl *PublicController) recordTravelView(ctx *gin.Context, travelID, ip string) {
	doIncrement := func() {
		if err := ctrl.travelLogic.IncrementView(ctx.Request.Context(), travelID); err != nil {
			// 计数失败不阻塞响应，浏览统计可容忍少量丢失
			return
		}
	}

	if cache.RedisClient == nil {
		doIncrement()
		return
	}

	key := "view:travel:" + travelID + ":" + ip
	dedupCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
	defer cancel()
	ok, err := cache.RedisClient.SetNX(dedupCtx, key, "1", viewDedupWindow).Result()
	if err != nil {
		doIncrement()
		return
	}
	if !ok {
		return
	}
	doIncrement()
}
