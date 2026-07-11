package controller

import (
	"errors"

	"cus-cms/enum"
	"cus-cms/internal/dto/req"
	"cus-cms/internal/logic"
	"cus-cms/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PublicController 公开接口控制器结构体。
type PublicController struct {
	setupLogic    *logic.SetupLogic
	articleLogic  *logic.ArticleLogic
	categoryLogic *logic.CategoryLogic
	tagLogic      *logic.TagLogic
}

// NewPublicController 创建 PublicController 实例。
func NewPublicController() *PublicController {
	return &PublicController{
		setupLogic:    logic.NewSetupLogic(),
		articleLogic:  logic.NewArticleLogic(),
		categoryLogic: logic.NewCategoryLogic(),
		tagLogic:      logic.NewTagLogic(),
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
