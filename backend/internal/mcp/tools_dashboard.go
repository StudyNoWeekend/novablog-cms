package mcp

import (
	"context"
	"fmt"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// DashboardOverviewOut get_dashboard_overview 出参：工作台概览 + 流量摘要。
type DashboardOverviewOut struct {
	Overview *res.OverviewRes       `json:"overview" jsonschema:"各内容模块的内容/评论统计与发布节奏"`
	Traffic  *res.TrafficSummaryRes `json:"traffic" jsonschema:"流量摘要（今日/昨日 PV、UV 与总 PV）"`
}

// getDashboardOverview 工作台概览统计。
func (s *Service) getDashboardOverview(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, DashboardOverviewOut, error) {
	overview, err := s.analyticsLogic.GetOverview(ctx)
	if err != nil {
		return nil, DashboardOverviewOut{}, fmt.Errorf("获取工作台概览失败: %w", err)
	}
	traffic, err := s.viewStatsLogic.GetTrafficSummary(ctx)
	if err != nil {
		return nil, DashboardOverviewOut{}, fmt.Errorf("获取流量摘要失败: %w", err)
	}
	return nil, DashboardOverviewOut{Overview: overview, Traffic: traffic}, nil
}

// DashboardTrendIn get_dashboard_trend 入参。
type DashboardTrendIn struct {
	Range string `json:"range,omitempty" jsonschema:"时间范围: 7d / 30d / 90d，默认 30d"`
}

// DashboardTrendOut get_dashboard_trend 出参：内容产出趋势 + 流量趋势 + 内容分布。
type DashboardTrendOut struct {
	ContentTrend *res.ContentTrendRes `json:"content_trend" jsonschema:"内容产出趋势（按日统计各模块新增数）"`
	TrafficTrend *res.TrafficTrendRes `json:"traffic_trend" jsonschema:"访问量趋势"`
	Distribution *res.DistributionRes `json:"distribution" jsonschema:"内容分布统计"`
}

// getDashboardTrend 内容产出趋势与流量趋势。
func (s *Service) getDashboardTrend(ctx context.Context, _ *sdk.CallToolRequest, in DashboardTrendIn) (*sdk.CallToolResult, DashboardTrendOut, error) {
	rangeStr := in.Range
	switch rangeStr {
	case "", "7d", "30d", "90d":
	default:
		return nil, DashboardTrendOut{}, fmt.Errorf("非法的时间范围 %q，可选：7d / 30d / 90d", rangeStr)
	}
	contentTrend, err := s.analyticsLogic.GetContentTrend(ctx, req.ContentTrendReq{Range: rangeStr})
	if err != nil {
		return nil, DashboardTrendOut{}, fmt.Errorf("获取内容趋势失败: %w", err)
	}
	trafficTrend, err := s.viewStatsLogic.GetTrafficTrend(ctx, rangeStr)
	if err != nil {
		return nil, DashboardTrendOut{}, fmt.Errorf("获取流量趋势失败: %w", err)
	}
	distribution, err := s.analyticsLogic.GetDistribution(ctx)
	if err != nil {
		return nil, DashboardTrendOut{}, fmt.Errorf("获取内容分布失败: %w", err)
	}
	return nil, DashboardTrendOut{
		ContentTrend: contentTrend,
		TrafficTrend: trafficTrend,
		Distribution: distribution,
	}, nil
}

// TopContentIn get_top_content 入参。
type TopContentIn struct {
	Type  string `json:"type,omitempty" jsonschema:"内容类型: article(文章) / travel(旅行攻略) / all(全部，默认)"`
	Sort  string `json:"sort,omitempty" jsonschema:"排序方式: views(按浏览量，默认) / comments(按评论数)"`
	Limit int    `json:"limit,omitempty" jsonschema:"返回条数 1-20，默认 5"`
}

// getTopContent 热门内容排行。
func (s *Service) getTopContent(ctx context.Context, _ *sdk.CallToolRequest, in TopContentIn) (*sdk.CallToolResult, res.TopContentRes, error) {
	if in.Type != "" && in.Type != "article" && in.Type != "travel" && in.Type != "all" {
		return nil, res.TopContentRes{}, fmt.Errorf("非法的内容类型 %q，可选：article / travel / all", in.Type)
	}
	if in.Sort != "" && in.Sort != "views" && in.Sort != "comments" {
		return nil, res.TopContentRes{}, fmt.Errorf("非法的排序方式 %q，可选：views / comments", in.Sort)
	}
	top, err := s.analyticsLogic.GetTopContent(ctx, req.TopContentReq{Type: in.Type, Sort: in.Sort, Limit: in.Limit})
	if err != nil {
		return nil, res.TopContentRes{}, fmt.Errorf("获取热门内容失败: %w", err)
	}
	return nil, *top, nil
}
