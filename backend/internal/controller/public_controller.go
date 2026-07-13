package controller

import (
	"errors"
	"strconv"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PublicController 公开接口控制器结构体。
type PublicController struct {
	setupLogic     *logic.SetupLogic
	articleLogic   *logic.ArticleLogic
	categoryLogic  *logic.CategoryLogic
	tagLogic       *logic.TagLogic
	bloggerLogic   *logic.BloggerLogic
	commentLogic   *logic.CommentLogic
	travelLogic    *logic.TravelGuideLogic
	portfolioLogic *logic.PortfolioLogic
	videoLogic     *logic.VideoLogic
	musicLogic     *logic.MusicLogic
}

// NewPublicController 创建 PublicController 实例。
func NewPublicController() *PublicController {
	return &PublicController{
		setupLogic:     logic.NewSetupLogic(),
		articleLogic:   logic.NewArticleLogic(),
		categoryLogic:  logic.NewCategoryLogic(),
		tagLogic:       logic.NewTagLogic(),
		bloggerLogic:   logic.NewBloggerLogic(),
		commentLogic:   logic.NewCommentLogic(),
		travelLogic:    logic.NewTravelGuideLogic(),
		portfolioLogic: logic.NewPortfolioLogic(),
		videoLogic:     logic.NewVideoLogic(),
		musicLogic:     logic.NewMusicLogic(),
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

// GetArticles 公开文章列表 GET /api/v1/public/articles
func (ctrl *PublicController) GetArticles(ctx *gin.Context) {
	var r req.ArticleListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误")
		return
	}
	// 公开接口只返回已发布文章
	published := int16(2)
	r.Status = &published
	result, err := ctrl.articleLogic.GetList(ctx, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetArticleBySlug 根据 slug 获取文章详情 GET /api/v1/public/articles/:slug
func (ctrl *PublicController) GetArticleBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")
	result, err := ctrl.articleLogic.GetBySlug(ctx, slug)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	// 公开接口只返回已发布文章
	if result.Status != 2 {
		response.Error(ctx, "文章不存在")
		return
	}
	response.Success(ctx, result)
}

// GetCategories 公开分类列表 GET /api/v1/public/categories
func (ctrl *PublicController) GetCategories(ctx *gin.Context) {
	result, err := ctrl.categoryLogic.GetAll(ctx, "article")
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetTags 公开标签列表 GET /api/v1/public/tags
func (ctrl *PublicController) GetTags(ctx *gin.Context) {
	result, err := ctrl.tagLogic.GetAll(ctx)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetBlogger 获取博主公开信息 GET /api/v1/public/blogger
func (ctrl *PublicController) GetBlogger(ctx *gin.Context) {
	result, err := ctrl.bloggerLogic.GetPublicInfo(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetHotArticles 获取热门文章列表 GET /api/v1/public/articles/hot
func (ctrl *PublicController) GetHotArticles(ctx *gin.Context) {
	var r req.HotArticleReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误")
		return
	}
	count := r.Count
	if count <= 0 {
		count = 5
	}
	result, err := ctrl.articleLogic.GetHotList(ctx.Request.Context(), count)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetRandomArticles 获取随机文章列表 GET /api/v1/public/articles/random
func (ctrl *PublicController) GetRandomArticles(ctx *gin.Context) {
	var r req.RandomArticleReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误")
		return
	}
	count := r.Count
	if count <= 0 {
		count = 5
	}
	result, err := ctrl.articleLogic.GetRandomList(ctx.Request.Context(), count)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// IncrementArticleView 增加文章浏览量 GET /api/v1/public/articles/:slug/view
func (ctrl *PublicController) IncrementArticleView(ctx *gin.Context) {
	slug := ctx.Param("slug")
	if err := ctrl.articleLogic.IncrementView(ctx.Request.Context(), slug); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// GetComments 获取公开评论列表 GET /api/v1/public/comments
func (ctrl *PublicController) GetComments(ctx *gin.Context) {
	var r req.CommentListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误")
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
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// CreateComment 创建公开评论 POST /api/v1/public/comments
func (ctrl *PublicController) CreateComment(ctx *gin.Context) {
	var r req.CreatePublicCommentReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误")
		return
	}
	ip := ctx.ClientIP()
	result, err := ctrl.commentLogic.CreatePublic(ctx.Request.Context(), &r, ip)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetTravels 获取已发布旅行攻略列表 GET /api/v1/public/travels
func (ctrl *PublicController) GetTravels(ctx *gin.Context) {
	var r req.TravelGuideListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误")
		return
	}
	result, err := ctrl.travelLogic.GetPublicList(ctx.Request.Context(), &r)
	if err != nil {
		response.Error(ctx, err.Error())
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
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetTravelDetail 获取已发布旅行攻略详情 GET /api/v1/public/travels/:id
func (ctrl *PublicController) GetTravelDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.travelLogic.GetPublicDetail(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// IncrementTravelView 增加旅行攻略浏览量 GET /api/v1/public/travels/:id/view
func (ctrl *PublicController) IncrementTravelView(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := ctrl.travelLogic.IncrementView(ctx.Request.Context(), id); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// LikeTravel 增加旅行攻略点赞数 POST /api/v1/public/travels/:id/like
func (ctrl *PublicController) LikeTravel(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := ctrl.travelLogic.IncrementLike(ctx.Request.Context(), id); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// GetPortfolios 获取已发布作品集列表 GET /api/v1/public/portfolios
func (ctrl *PublicController) GetPortfolios(ctx *gin.Context) {
	var r req.PortfolioListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误")
		return
	}
	result, err := ctrl.portfolioLogic.GetPublicList(ctx.Request.Context(), &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetPortfolioDetail 获取已发布作品集详情 GET /api/v1/public/portfolios/:id
func (ctrl *PublicController) GetPortfolioDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.portfolioLogic.GetPublicDetail(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetVideos 获取已发布视频作品列表 GET /api/v1/public/videos
func (ctrl *PublicController) GetVideos(ctx *gin.Context) {
	var r req.VideoListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误")
		return
	}
	result, err := ctrl.videoLogic.GetPublicList(ctx.Request.Context(), &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetVideoDetail 获取已发布视频作品详情 GET /api/v1/public/videos/:id
func (ctrl *PublicController) GetVideoDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.videoLogic.GetPublicDetail(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetSongs 获取歌曲列表 GET /api/v1/public/music/songs
func (ctrl *PublicController) GetSongs(ctx *gin.Context) {
	var r req.SongListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误")
		return
	}
	result, err := ctrl.musicLogic.GetPublicSongList(ctx.Request.Context(), &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetSongDetail 获取歌曲详情 GET /api/v1/public/music/songs/:id
func (ctrl *PublicController) GetSongDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := ctrl.musicLogic.GetPublicSongByID(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetAudioURL 获取歌曲音频播放地址 GET /api/v1/public/music/audio-url/:song_id
func (ctrl *PublicController) GetAudioURL(ctx *gin.Context) {
	songID := ctx.Param("song_id")
	url, err := ctrl.musicLogic.GetPublicAudioURL(ctx.Request.Context(), songID)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, gin.H{"url": url})
}
