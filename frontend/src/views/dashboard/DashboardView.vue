<template>
  <div class="dashboard">
    <!-- Loading skeleton -->
    <div v-if="loading" class="dashboard-skeleton">
      <div class="skeleton-grid">
        <div class="skeleton-card" v-for="i in 4" :key="i"></div>
      </div>
      <div class="skeleton-row">
        <div class="skeleton-block skeleton-left"></div>
        <div class="skeleton-block skeleton-right"></div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else-if="!overview" class="empty-state">
      <div class="empty-icon">
        <FileTextOutlined />
      </div>
      <p class="empty-text">暂无工作台数据</p>
    </div>

    <!-- Main content -->
    <template v-else>
      <!-- 顶部统计卡片 -->
      <div class="stat-grid">
        <div v-for="(s, i) in statsCards" :key="i" class="stat-card">
          <div class="stat-card-label">
            <span class="stat-icon" :class="s.iconBg">
              <component :is="s.icon" />
            </span>
            <span class="stat-name">{{ s.label }}</span>
          </div>
          <div class="stat-card-value">{{ s.value.toLocaleString() }}</div>
          <div class="stat-card-sub">{{ s.sub }}</div>
        </div>
      </div>

      <!-- 待办提醒区 -->
      <div v-if="todoReminder.show" class="todo-reminder">
        <div v-if="todoReminder.drafts > 0 && moduleStore.isEnabled('article_enabled')" class="todo-item" @click="goToArticles">
          <span class="todo-count">{{ todoReminder.drafts }}</span>
          <span class="todo-text">篇草稿待发布</span>
          <span class="todo-arrow">&rarr;</span>
        </div>
      </div>

      <!-- 中间两列布局 -->
      <div class="dashboard-row">
        <!-- 左列：内容产出趋势 -->
        <div class="dashboard-col-left">
          <div class="card">
            <div class="chart-header">
              <h3 class="card-title">内容产出趋势</h3>
              <div class="time-tabs">
                <button
                  v-for="t in timeTabs"
                  :key="t.value"
                  :class="['tab-btn', { active: activeRange === t.value }]"
                  @click="switchRange(t.value)"
                >
                  {{ t.label }}
                </button>
              </div>
            </div>
            <div class="chart-wrap">
              <v-chart
                v-if="trendData && trendData.items.length"
                class="trend-chart"
                :option="trendOption"
                autoresize
              />
              <div v-else class="chart-empty">暂无趋势数据</div>
            </div>
            <div class="summary-bar">
              <div v-for="(item, idx) in summaryStats" :key="idx" class="summary-pill">
                <div class="summary-pill-label">{{ item.label }}</div>
                <div class="summary-pill-value">{{ item.value }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 右列：内容分布 -->
        <div class="dashboard-col-right">
          <div class="card">
            <h3 class="card-title">内容分布</h3>
            <div class="pie-wrap">
              <v-chart
                v-if="distribution && distribution.items.length"
                class="pie-chart"
                :option="pieOption"
                autoresize
              />
              <div v-else class="chart-empty">暂无分布数据</div>
            </div>
            <div class="pie-legend">
              <div v-for="(item, idx) in pieLegend" :key="idx" class="legend-row">
                <span class="legend-dot" :style="{ background: item.color }" />
                <span class="legend-name">{{ item.name }}</span>
                <span class="legend-percent">{{ formatPercentage(item.percentage) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 底部：热门内容排行 + 最近评论 -->
      <div class="dashboard-row bottom-row">
        <div class="dashboard-col-left">
          <div class="card top-content-card">
            <h3 class="card-title">热门内容排行</h3>
            <div class="top-content-list">
              <div
                v-for="(item, idx) in topContentList"
                :key="item.id"
                class="top-content-item"
              >
                <span :class="['rank-badge', `rank-${idx + 1}`]">{{ idx + 1 }}</span>
                <span class="top-content-title">{{ item.title }}</span>
                <span class="top-content-views">{{ item.view_count.toLocaleString() }} 浏览</span>
                <span class="top-content-comments">{{ item.comment_count }} 评论</span>
              </div>
              <div v-if="topContentList.length === 0" class="list-empty">暂无热门内容</div>
            </div>
          </div>
        </div>
        <div class="dashboard-col-right">
          <div class="card">
            <h3 class="card-title">最近评论</h3>
            <div class="comment-list">
              <div v-for="item in recentCommentsList" :key="item.id" class="comment-item">
                <div class="comment-avatar" :class="{ blogger: item.is_blogger }">
                  {{ item.nickname.charAt(0) }}
                </div>
                <div class="comment-body">
                  <div class="comment-top">
                    <span class="comment-name">{{ item.nickname }}</span>
                    <span v-if="item.is_blogger" class="blogger-tag">博主</span>
                    <span class="comment-time">{{ formatCommentTime(item.created_at) }}</span>
                  </div>
                  <div class="comment-text">{{ item.content }}</div>
                  <div class="comment-target">{{ item.target_title }}</div>
                </div>
              </div>
              <div v-if="recentCommentsList.length === 0" class="list-empty">暂无评论</div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { Component } from 'vue'
import { useRouter } from 'vue-router'
import {
  FileTextOutlined,
  EyeOutlined,
  MessageOutlined,
  PictureOutlined,
} from '@ant-design/icons-vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import { analyticsApi } from '@/api/analytics'
import { useModuleStore } from '@/stores/module'
import type {
  OverviewData,
  ContentTrendData,
  TopContentData,
  DistributionData,
  RecentCommentsData,
} from '@/types/analytics'

use([
  CanvasRenderer,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
])

const distributionColors: Record<string, string> = {
  article: '#4a6cf7',
  portfolio: '#10b981',
  video: '#f59e0b',
  travel: '#8b5cf6',
  song: '#ec4899',
}

const timeTabs = [
  { label: '7天', value: '7d' },
  { label: '30天', value: '30d' },
  { label: '90天', value: '90d' },
]

const loading = ref(true)
const overview = ref<OverviewData | null>(null)
const trendData = ref<ContentTrendData | null>(null)
const topContent = ref<TopContentData | null>(null)
const distribution = ref<DistributionData | null>(null)
const recentComments = ref<RecentCommentsData | null>(null)
const activeRange = ref('30d')
const router = useRouter()
const moduleStore = useModuleStore()

async function loadOverview() {
  try {
    overview.value = await analyticsApi.getOverview()
  } catch (e) {
    console.error('Failed to load overview:', e)
  }
}

async function loadTrend(range: string) {
  try {
    trendData.value = await analyticsApi.getContentTrend({
      range: range as '7d' | '30d' | '90d',
    })
  } catch (e) {
    console.error('Failed to load trend:', e)
  }
}

async function loadTopContent() {
  try {
    topContent.value = await analyticsApi.getTopContent({ limit: 5 })
  } catch (e) {
    console.error('Failed to load top content:', e)
  }
}

async function loadDistribution() {
  try {
    distribution.value = await analyticsApi.getDistribution()
  } catch (e) {
    console.error('Failed to load distribution:', e)
  }
}

async function loadRecentComments() {
  try {
    recentComments.value = await analyticsApi.getRecentComments({ limit: 5 })
  } catch (e) {
    console.error('Failed to load recent comments:', e)
  }
}

onMounted(async () => {
  await Promise.all([
    loadOverview(),
    loadTrend('30d'),
    loadTopContent(),
    loadDistribution(),
    loadRecentComments(),
  ])
  loading.value = false
})

async function switchRange(range: string) {
  activeRange.value = range
  await loadTrend(range)
}

function goToArticles() {
  router.push('/articles')
}

interface StatsCardItem {
  label: string
  value: number
  sub: string
  icon: Component
  iconBg: string
}

const statsCards = computed((): StatsCardItem[] => {
  if (!overview.value) return []
  const o = overview.value
  return [
    {
      label: '文章总数',
      value: o.article_total,
      sub: `已发布 ${o.article_published} · 草稿 ${o.article_draft}`,
      icon: FileTextOutlined,
      iconBg: 'bg-blue',
    },
    {
      label: '攻略浏览',
      value: o.travel_views,
      sub: '旅行攻略总浏览量',
      icon: EyeOutlined,
      iconBg: 'bg-green',
    },
    {
      label: '评论总数',
      value: o.comment_total,
      sub: '全部评论',
      icon: MessageOutlined,
      iconBg: 'bg-purple',
    },
    {
      label: '作品总数',
      value: o.portfolio_total + o.video_total + o.song_total,
      sub: `摄影 ${o.portfolio_total} · 视频 ${o.video_total} · 音乐 ${o.song_total}`,
      icon: PictureOutlined,
      iconBg: 'bg-orange',
    },
  ]
})

const todoReminder = computed(() => {
  if (!overview.value) return { show: false, drafts: 0 }
  return {
    show: overview.value.article_draft > 0,
    drafts: overview.value.article_draft,
  }
})

const trendOption = computed(() => {
  if (!trendData.value) return {}
  const items = trendData.value.items
  return {
    grid: { left: 0, right: 0, top: 42, bottom: 8, containLabel: true },
    tooltip: { trigger: 'axis' },
    legend: { data: ['文章', '旅行攻略'], top: 0, textStyle: { color: '#64748b' } },
    xAxis: {
      type: 'category',
      data: items.map((i) => i.date.slice(5)),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: '#94a3b8', fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { type: 'dashed', color: '#f1f5f9' } },
      axisLabel: { color: '#94a3b8' },
    },
    series: [
      {
        name: '文章',
        type: 'bar',
        data: items.map((i) => i.article_count),
        barWidth: '35%',
        itemStyle: { borderRadius: [4, 4, 0, 0], color: '#4a6cf7' },
      },
      {
        name: '旅行攻略',
        type: 'bar',
        data: items.map((i) => i.travel_count),
        barWidth: '35%',
        itemStyle: { borderRadius: [4, 4, 0, 0], color: '#10b981' },
      },
    ],
  }
})

const pieOption = computed(() => {
  if (!distribution.value) return {}
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    series: [
      {
        type: 'pie',
        radius: ['55%', '75%'],
        center: ['50%', '50%'],
        label: { show: false },
        labelLine: { show: false },
        data: distribution.value.items.map((i) => ({
          value: i.count,
          name: i.name,
          itemStyle: { color: distributionColors[i.type] || '#94a3b8' },
        })),
      },
    ],
  }
})

const pieLegend = computed(() => {
  if (!distribution.value) return []
  return distribution.value.items.map((i) => ({
    name: i.name,
    percentage: i.percentage,
    color: distributionColors[i.type] || '#94a3b8',
  }))
})

const summaryStats = computed(() => {
  if (!trendData.value) return []
  const items = trendData.value.items
  const totalPublish = items.reduce((s, i) => s + i.article_count + i.travel_count, 0)
  const totalArticles = items.reduce((s, i) => s + i.article_count, 0)
  const totalTravel = items.reduce((s, i) => s + i.travel_count, 0)
  const avgPerDay = items.length > 0 ? (totalPublish / items.length).toFixed(1) : '0'
  return [
    { label: '总发布', value: totalPublish },
    { label: '文章', value: totalArticles },
    { label: '旅行攻略', value: totalTravel },
    { label: '日均发布', value: avgPerDay },
  ]
})

const recentCommentsList = computed(() => {
  if (!recentComments.value) return []
  return recentComments.value.items
})

const topContentList = computed(() => {
  if (!topContent.value) return []
  return topContent.value.items
})

function formatCommentTime(dateStr: string): string {
  const d = new Date(dateStr)
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const min = String(d.getMinutes()).padStart(2, '0')
  return `${mm}-${dd} ${hh}:${min}`
}

function formatPercentage(value: number): string {
  return `${value.toFixed(1)}%`
}
</script>

<style scoped>
.dashboard {
  background: #f5f7fa;
  padding: 24px;
  min-height: 100%;
}

/* Stat Grid */
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  display: flex;
  flex-direction: column;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.stat-card-label {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.stat-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  font-size: 20px;
}

.bg-blue {
  background: rgba(74, 108, 247, 0.1);
  color: #4a6cf7;
}

.bg-green {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.bg-orange {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.bg-purple {
  background: rgba(139, 92, 246, 0.1);
  color: #8b5cf6;
}

.stat-name {
  font-size: 14px;
  color: #64748b;
  font-weight: 500;
}

.stat-card-value {
  font-size: 28px;
  font-weight: 700;
  color: #1e293b;
  line-height: 1.2;
  margin-bottom: 8px;
}

.stat-card-sub {
  font-size: 13px;
  color: #64748b;
}

/* Todo Reminder */
.todo-reminder {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  padding: 16px 20px;
  background: #fff7ed;
  border-left: 4px solid #f59e0b;
  border-radius: 8px;
}

.todo-item {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 6px;
  transition: background 0.15s ease;
}

.todo-item:hover {
  background: rgba(245, 158, 11, 0.1);
}

.todo-count {
  font-size: 18px;
  font-weight: 700;
  color: #f59e0b;
}

.todo-text {
  font-size: 14px;
  color: #92400e;
}

.todo-arrow {
  font-size: 14px;
  color: #f59e0b;
}

/* Dashboard Row */
.dashboard-row {
  display: flex;
  gap: 20px;
}

.dashboard-col-left {
  flex: 0 0 65%;
  max-width: 65%;
}

.dashboard-col-right {
  flex: 0 0 35%;
  max-width: 35%;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card {
  background: #ffffff;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
}

/* Chart Header */
.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 16px 0;
}

.chart-header .card-title {
  margin: 0;
}

.time-tabs {
  display: flex;
  gap: 4px;
  background: #f1f5f9;
  border-radius: 8px;
  padding: 4px;
}

.tab-btn {
  border: none;
  background: transparent;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  color: #64748b;
  cursor: pointer;
  transition: all 0.15s ease;
  font-weight: 500;
}

.tab-btn.active {
  background: #4a6cf7;
  color: #ffffff;
}

.chart-wrap {
  height: 260px;
  margin-bottom: 20px;
}

.trend-chart {
  width: 100%;
  height: 100%;
}

.chart-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  font-size: 14px;
  color: #94a3b8;
}

/* Summary Bar */
.summary-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.summary-pill {
  background: #f8fafc;
  border-radius: 10px;
  padding: 14px;
  text-align: center;
}

.summary-pill-label {
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 6px;
}

.summary-pill-value {
  font-size: 16px;
  font-weight: 700;
  color: #1e293b;
}

/* Pie Chart */
.pie-wrap {
  height: 200px;
  margin-bottom: 16px;
}

.pie-chart {
  width: 100%;
  height: 100%;
}

.pie-legend {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.legend-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.legend-name {
  font-size: 13px;
  color: #475569;
  flex: 1;
}

.legend-percent {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
}

/* Comment List */
.comment-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.comment-item {
  display: flex;
  gap: 12px;
}

.comment-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  color: #ffffff;
  background: #4a6cf7;
  flex-shrink: 0;
}

.comment-avatar.blogger {
  background: #8b5cf6;
}

.comment-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.comment-top {
  display: flex;
  align-items: center;
  gap: 6px;
}

.comment-name {
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
}

.blogger-tag {
  font-size: 11px;
  font-weight: 600;
  color: #8b5cf6;
  background: rgba(139, 92, 246, 0.1);
  padding: 1px 6px;
  border-radius: 4px;
}

.comment-time {
  font-size: 12px;
  color: #94a3b8;
  margin-left: auto;
}

.comment-text {
  font-size: 13px;
  color: #475569;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
}

.comment-target {
  font-size: 12px;
  color: #94a3b8;
}

/* Bottom Row */
.bottom-row {
  margin-top: 24px;
}

.bottom-row .dashboard-col-left,
.bottom-row .dashboard-col-right {
  display: flex;
  flex-direction: column;
}

.bottom-row .card {
  flex: 1;
}

.top-content-list {
  display: flex;
  flex-direction: column;
}

.top-content-item {
  display: grid;
  grid-template-columns: 40px 1fr auto auto;
  align-items: center;
  gap: 16px;
  padding: 14px 0;
  border-bottom: 1px solid #f8fafc;
}

.top-content-item:last-child {
  border-bottom: none;
}

.rank-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 700;
  background: #f1f5f9;
  color: #94a3b8;
}

.rank-badge.rank-1 {
  background: rgba(245, 158, 11, 0.12);
  color: #f59e0b;
}

.rank-badge.rank-2 {
  background: rgba(100, 116, 139, 0.12);
  color: #64748b;
}

.rank-badge.rank-3 {
  background: rgba(180, 83, 9, 0.1);
  color: #b45309;
}

.top-content-title {
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.top-content-views {
  font-size: 13px;
  color: #64748b;
  white-space: nowrap;
}

.top-content-comments {
  font-size: 13px;
  color: #64748b;
  white-space: nowrap;
}

.list-empty {
  text-align: center;
  padding: 24px 0;
  font-size: 14px;
  color: #94a3b8;
}

/* Skeleton */
.dashboard-skeleton {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.skeleton-card {
  height: 120px;
  background: #f1f5f9;
  border-radius: 12px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-row {
  display: flex;
  gap: 20px;
}

.skeleton-block {
  background: #f1f5f9;
  border-radius: 12px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-left {
  flex: 0 0 65%;
  height: 400px;
}

.skeleton-right {
  flex: 0 0 35%;
  height: 400px;
}

@keyframes skeleton-pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  gap: 16px;
}

.empty-icon {
  font-size: 48px;
  color: #cbd5e1;
}

.empty-text {
  font-size: 15px;
  color: #94a3b8;
  margin: 0;
}

/* Responsive */
@media (max-width: 1199px) {
  .dashboard {
    padding: 16px;
  }

  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .dashboard-row {
    flex-direction: column;
  }

  .dashboard-col-left,
  .dashboard-col-right {
    flex: 1 1 auto;
    max-width: 100%;
  }

  .summary-bar {
    grid-template-columns: repeat(2, 1fr);
  }

  .skeleton-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .skeleton-row {
    flex-direction: column;
  }

  .skeleton-left,
  .skeleton-right {
    flex: 1 1 auto;
    max-width: 100%;
  }
}

@media (max-width: 767px) {
  .dashboard {
    padding: 12px;
  }

  .stat-grid {
    grid-template-columns: 1fr;
  }

  .stat-card {
    padding: 12px;
  }

  .card {
    padding: 16px;
  }

  .chart-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .chart-wrap {
    height: 200px;
  }

  .pie-wrap {
    height: 160px;
  }

  .summary-pill {
    padding: 10px;
  }

  .summary-bar {
    grid-template-columns: repeat(2, 1fr);
  }

  .time-tabs {
    flex-wrap: wrap;
  }

  .tab-btn {
    padding: 6px 10px;
    font-size: 12px;
  }

  .top-content-item {
    grid-template-columns: 28px 1fr auto;
    gap: 8px;
  }

  .top-content-comments {
    display: none;
  }

  .todo-reminder {
    flex-direction: column;
    gap: 8px;
  }

  .skeleton-grid {
    grid-template-columns: 1fr;
  }
}
</style>
