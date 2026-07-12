import request from './request'
import type {
  OverviewData,
  ContentTrendData,
  ContentTrendParams,
  TopContentData,
  TopContentParams,
  DistributionData,
  RecentCommentsData,
  RecentCommentsParams,
} from '@/types/analytics'

export const analyticsApi = {
  getOverview() {
    return request.get<OverviewData>('/analytics/overview')
  },
  getContentTrend(params: ContentTrendParams) {
    return request.get<ContentTrendData>('/analytics/content-trend', { params })
  },
  getTopContent(params: TopContentParams) {
    return request.get<TopContentData>('/analytics/top-content', { params })
  },
  getDistribution() {
    return request.get<DistributionData>('/analytics/distribution')
  },
  getRecentComments(params: RecentCommentsParams) {
    return request.get<RecentCommentsData>('/analytics/recent-comments', { params })
  },
}
