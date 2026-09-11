package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// AnalyticsController 工作台统计控制器结构体。
type AnalyticsController struct {
	logic *logic.AnalyticsLogic
}

// NewAnalyticsController 创建 AnalyticsController 实例。
func NewAnalyticsController() *AnalyticsController {
	return &AnalyticsController{logic: logic.NewAnalyticsLogic()}
}

// GetOverview 获取工作台概览统计 GET /api/v1/analytics/overview
func (ctl *AnalyticsController) GetOverview(ctx *gin.Context) {
	result, err := ctl.logic.GetOverview(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetContentTrend 获取内容产出趋势 GET /api/v1/analytics/content-trend
func (ctl *AnalyticsController) GetContentTrend(ctx *gin.Context) {
	var r req.ContentTrendReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := ctl.logic.GetContentTrend(ctx.Request.Context(), r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetTopContent 获取热门内容排行 GET /api/v1/analytics/top-content
func (ctl *AnalyticsController) GetTopContent(ctx *gin.Context) {
	var r req.TopContentReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := ctl.logic.GetTopContent(ctx.Request.Context(), r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetDistribution 获取内容类型分布 GET /api/v1/analytics/distribution
func (ctl *AnalyticsController) GetDistribution(ctx *gin.Context) {
	result, err := ctl.logic.GetDistribution(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetRecentComments 获取最近评论 GET /api/v1/analytics/recent-comments
func (ctl *AnalyticsController) GetRecentComments(ctx *gin.Context) {
	var r req.RecentCommentsReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := ctl.logic.GetRecentComments(ctx.Request.Context(), r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}
