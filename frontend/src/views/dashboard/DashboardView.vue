<template>
  <div class="dashboard">
    <!-- Loading skeleton -->
    <div v-if="loading" class="dashboard-skeleton">
      <div class="skeleton-strip"></div>
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
      <!-- 核心指标条 -->
      <div class="metric-strip">
        <div class="metric-item">
          <div class="metric-value">{{ trafficSummary?.today_pv.toLocaleString() ?? '0' }}</div>
          <div class="metric-label">今日访问</div>
        </div>
        <div class="metric-item">
          <div class="metric-value">{{ trafficSummary?.today_uv.toLocaleString() ?? '0' }}</div>
          <div class="metric-label">今日访客</div>
        </div>
        <div class="metric-item">
          <div class="metric-value">{{ trafficSummary?.total_pv.toLocaleString() ?? '0' }}</div>
          <div class="metric-label">总访问量</div>
        </div>
        <div class="metric-item">
          <div class="metric-value">{{ totalContentCount.toLocaleString() }}</div>
          <div class="metric-label">内容总数</div>
        </div>
        <div class="metric-item">
          <div class="metric-value">{{ overview.comment_total.toLocaleString() }}</div>
          <div class="metric-label">总评论</div>
        </div>
      </div>

      <!-- 待办提醒区 -->
      <div v-if="todoReminder.show" class="todo-reminder">
        <div class="todo-item" @click="goToArticles">
          <span class="todo-count">{{ todoReminder.drafts }}</span>
          <span class="todo-text">篇草稿待发布</span>
          <span class="todo-arrow">&rarr;</span>
        </div>
      </div>

      <!-- 趋势卡：访问量 / 内容产出 -->
      <div class="card trend-card">
        <div class="chart-header">
          <div class="header-tabs">
            <div class="time-tabs">
              <button
                :class="['tab-btn', { active: metricTab === 'traffic' }]"
                @click="metricTab = 'traffic'"
              >
                访问量
              </button>
              <button
                :class="['tab-btn', { active: metricTab === 'production' }]"
                @click="metricTab = 'production'"
              >
                内容产出
              </button>
            </div>
          </div>
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
            v-if="trendChartDataReady"
            class="trend-chart"
            :option="trendOption"
            autoresize
          />
          <div v-else class="chart-empty">暂无趋势数据</div>
        </div>
        <div v-if="metricTab === 'production'" class="summary-bar">
          <div v-for="(item, idx) in productionSummary" :key="idx" class="summary-pill">
            <div class="summary-pill-label">{{ item.label }}</div>
            <div class="summary-pill-value">{{ item.value }}</div>
          </div>
        </div>
      </div>

      <!-- 中间两列布局：模块数据表 + 访问分布 -->
      <div class="dashboard-row">
        <div class="dashboard-col-left">
          <div class="card">
            <h3 class="card-title">模块数据</h3>
            <table v-if="moduleRows.length" class="module-table">
              <thead>
                <tr>
                  <th>模块</th>
                  <th>发布</th>
                  <th>累计访问</th>
                  <th>今日</th>
                  <th class="col-week">7日</th>
                  <th class="col-share">访问占比</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in moduleRows" :key="row.type">
                  <td>
                    <span class="legend-dot" :style="{ background: colorOf(row.type) }" />
                    <span class="module-name">{{ labelOf(row.type) }}</span>
                  </td>
                  <td>
                    {{ row.total.toLocaleString() }}
                    <span class="cell-sub">已发布 {{ row.published.toLocaleString() }}</span>
                  </td>
                  <td>{{ row.total_views.toLocaleString() }}</td>
                  <td>{{ row.today_views.toLocaleString() }}</td>
                  <td class="col-week">{{ row.week_views.toLocaleString() }}</td>
                  <td class="col-share">
                    <div class="share-cell">
                      <div class="share-bar">
                        <div
                          class="share-fill"
                          :style="{ width: shareOf(row) + '%', background: colorOf(row.type) }"
                        />
                      </div>
                      <span class="share-text">{{ shareOf(row) }}%</span>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
            <div v-else class="list-empty">暂无启用的内容模块</div>
          </div>
        </div>
        <div class="dashboard-col-right">
          <div class="card">
            <h3 class="card-title">访问分布</h3>
            <div class="pie-wrap">
              <v-chart
                v-if="viewDistribution.length"
                class="pie-chart"
                :option="pieOption"
                autoresize
              />
              <div v-else class="chart-empty">暂无访问数据</div>
            </div>
            <div class="pie-legend">
              <div v-for="(item, idx) in viewLegend" :key="idx" class="legend-row">
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
                <span class="top-content-type" :style="{ color: colorOf(item.type) }">
                  {{ labelOf(item.type) }}
                </span>
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
import { useRouter } from 'vue-router'
import { FileTextOutlined } from '@ant-design/icons-vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import { analyticsApi } from '@/api/analytics'
import { useModuleStore } from '@/stores/module'
import { CONTENT_TYPE_DEFS, CONTENT_TYPE_MAP } from '@/constants/dashboardCards'
import type {
  OverviewData,
  ContentTrendData,
  TopContentData,
  RecentCommentsData,
  TrafficSummaryData,
  TrafficTrendData,
  ModuleStatsData,
} from '@/types/analytics'

use([
  CanvasRenderer,
  BarChart,
  LineChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
])

const timeTabs = [
  { label: '7天', value: '7d' },
  { label: '30天', value: '30d' },
  { label: '90天', value: '90d' },
]

const loading = ref(true)
const overview = ref<OverviewData | null>(null)
const trafficSummary = ref<TrafficSummaryData | null>(null)
const trafficTrend = ref<TrafficTrendData | null>(null)
const productionTrend = ref<ContentTrendData | null>(null)
const moduleStats = ref<ModuleStatsData | null>(null)
const topContent = ref<TopContentData | null>(null)
const recentComments = ref<RecentCommentsData | null>(null)
const activeRange = ref('30d')
const metricTab = ref<'traffic' | 'production'>('traffic')
const router = useRouter()
const moduleStore = useModuleStore()

async function loadOverview() {
  try {
    overview.value = await analyticsApi.getOverview()
  } catch (e) {
    console.error('Failed to load overview:', e)
  }
}

async function loadTrafficSummary() {
  try {
    trafficSummary.value = await analyticsApi.getTrafficSummary()
  } catch (e) {
    console.error('Failed to load traffic summary:', e)
  }
}

async function loadTrafficTrend(range: string) {
  try {
    trafficTrend.value = await analyticsApi.getTrafficTrend({
      range: range as '7d' | '30d' | '90d',
    })
  } catch (e) {
    console.error('Failed to load traffic trend:', e)
  }
}

async function loadProductionTrend(range: string) {
  try {
    productionTrend.value = await analyticsApi.getContentTrend({
      range: range as '7d' | '30d' | '90d',
    })
  } catch (e) {
    console.error('Failed to load production trend:', e)
  }
}

async function loadModuleStats() {
  try {
    moduleStats.value = await analyticsApi.getModuleStats()
  } catch (e) {
    console.error('Failed to load module stats:', e)
  }
}

async function loadTopContent() {
  try {
    topContent.value = await analyticsApi.getTopContent({ limit: 5 })
  } catch (e) {
    console.error('Failed to load top content:', e)
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
    // 模块开关就绪后再渲染，避免首屏闪烁；拉取失败按全部开启处理
    moduleStore.fetchConfig().catch(() => {}),
    loadOverview(),
    loadTrafficSummary(),
    loadTrafficTrend('30d'),
    loadProductionTrend('30d'),
    loadModuleStats(),
    loadTopContent(),
    loadRecentComments(),
  ])
  loading.value = false
})

async function switchRange(range: string) {
  activeRange.value = range
  await Promise.all([loadTrafficTrend(range), loadProductionTrend(range)])
}

function goToArticles() {
  router.push('/articles')
}

// 当前趋势卡的数据是否可渲染
const trendChartDataReady = computed(() => {
  if (metricTab.value === 'traffic') {
    return !!trafficTrend.value?.items.length
  }
  return !!productionTrend.value?.items.length && enabledDefs.value.length > 0
})

// 随模块开关启停的内容类型系列
const enabledDefs = computed(() =>
  CONTENT_TYPE_DEFS.filter((d) => moduleStore.isEnabled(d.moduleKey))
)

// 模块数据表行（仅启用模块）
const moduleRows = computed(() => {
  if (!moduleStats.value) return []
  return moduleStats.value.items.filter((row) => {
    const meta = CONTENT_TYPE_MAP[row.type]
    return !meta || moduleStore.isEnabled(meta.moduleKey)
  })
})

// 启用模块内容总数（指标条）
const totalContentCount = computed(() =>
  moduleRows.value.reduce((s, row) => s + row.total, 0)
)

function colorOf(type: string): string {
  return CONTENT_TYPE_MAP[type]?.color || '#8A93A8'
}

function labelOf(type: string): string {
  return CONTENT_TYPE_MAP[type]?.label || type
}

// 访问占比（基于启用模块累计访问总量，保留一位小数）
function shareOf(row: { total_views: number }): number {
  const total = moduleRows.value.reduce((s, r) => s + r.total_views, 0)
  if (total <= 0) return 0
  const pct = (row.total_views / total) * 100
  return Math.round(pct * 10) / 10
}

const todoReminder = computed(() => {
  if (!overview.value || !moduleStore.isEnabled('article_enabled')) {
    return { show: false, drafts: 0 }
  }
  return {
    show: overview.value.article_draft > 0,
    drafts: overview.value.article_draft,
  }
})

// 访问趋势：各启用模块堆叠柱 + 全站 PV 折线
const trendOption = computed(() => {
  if (metricTab.value === 'production') {
    return productionOption.value
  }
  if (!trafficTrend.value) return {}
  const items = trafficTrend.value.items
  const series: unknown[] = enabledDefs.value.map((def) => ({
    name: def.label,
    type: 'bar',
    stack: 'total',
    data: items.map((i) => i.counts?.[def.type] ?? 0),
    itemStyle: { color: def.color },
  }))
  series.push({
    name: '全站',
    type: 'line',
    data: items.map((i) => i.total_pv),
    smooth: true,
    symbol: 'none',
    lineStyle: { color: '#B7791F', width: 2 },
    itemStyle: { color: '#B7791F' },
  })
  return {
    grid: { left: 0, right: 0, top: 42, bottom: 8, containLabel: true },
    tooltip: { trigger: 'axis' },
    legend: {
      data: [...enabledDefs.value.map((d) => d.label), '全站'],
      type: 'scroll',
      top: 0,
      textStyle: { color: '#667085' },
    },
    xAxis: {
      type: 'category',
      data: items.map((i) => i.date.slice(5)),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: '#8A93A8', fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { type: 'dashed', color: '#F5F7FC' } },
      axisLabel: { color: '#8A93A8' },
    },
    series,
  }
})

// 内容产出趋势：各启用模块堆叠柱
const productionOption = computed(() => {
  if (!productionTrend.value) return {}
  const items = productionTrend.value.items
  const series = enabledDefs.value.map((def) => ({
    name: def.label,
    type: 'bar',
    stack: 'total',
    data: items.map((i) => i.counts?.[def.type] ?? 0),
    itemStyle: { color: def.color },
  }))
  return {
    grid: { left: 0, right: 0, top: 42, bottom: 8, containLabel: true },
    tooltip: { trigger: 'axis' },
    legend: {
      data: enabledDefs.value.map((d) => d.label),
      type: 'scroll',
      top: 0,
      textStyle: { color: '#667085' },
    },
    xAxis: {
      type: 'category',
      data: items.map((i) => i.date.slice(5)),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: '#8A93A8', fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { type: 'dashed', color: '#F5F7FC' } },
      axisLabel: { color: '#8A93A8' },
    },
    series,
  }
})

// 内容产出汇总条（仅产出视图显示）
const productionSummary = computed(() => {
  if (!productionTrend.value) return []
  const items = productionTrend.value.items
  const sumOf = (type: string) => items.reduce((s, i) => s + (i.counts?.[type] ?? 0), 0)
  const total = enabledDefs.value.reduce((s, def) => s + sumOf(def.type), 0)
  const avgPerDay = items.length > 0 ? (total / items.length).toFixed(1) : '0'
  const stats: { label: string; value: number | string }[] = [
    { label: '总产出', value: total },
  ]
  for (const def of enabledDefs.value) {
    stats.push({ label: def.label, value: sumOf(def.type) })
  }
  stats.push({ label: '日均产出', value: avgPerDay })
  return stats
})

// 访问分布：按模块累计访问量占比（剔除无访问的模块）
const viewDistribution = computed(() => {
  return moduleRows.value
    .filter((row) => row.total_views > 0)
    .map((row) => ({
      type: row.type,
      name: labelOf(row.type),
      count: row.total_views,
      color: colorOf(row.type),
    }))
})

const viewLegend = computed(() => {
  const total = viewDistribution.value.reduce((s, i) => s + i.count, 0)
  return viewDistribution.value.map((i) => ({
    name: i.name,
    color: i.color,
    percentage: total > 0 ? (i.count / total) * 100 : 0,
  }))
})

const pieOption = computed(() => {
  if (!viewDistribution.value.length) return {}
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    series: [
      {
        type: 'pie',
        radius: ['55%', '75%'],
        center: ['50%', '50%'],
        label: { show: false },
        labelLine: { show: false },
        data: viewDistribution.value.map((i) => ({
          value: i.count,
          name: i.name,
          itemStyle: { color: i.color },
        })),
      },
    ],
  }
})

const topContentList = computed(() => {
  if (!topContent.value) return []
  // 过滤掉所属模块已关闭的内容条目
  return topContent.value.items.filter((i) => {
    const meta = CONTENT_TYPE_MAP[i.type]
    return !meta || moduleStore.isEnabled(meta.moduleKey)
  })
})

const recentCommentsList = computed(() => {
  if (!recentComments.value) return []
  return recentComments.value.items
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
  background: #F5F7FC;
  padding: 24px;
  min-height: 100%;
}

/* Metric Strip */
.metric-strip {
  display: flex;
  background: #ffffff;
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: 12px;
  padding: 22px 12px;
  margin-bottom: 24px;
  box-shadow: 0 1px 3px rgba(41, 54, 92, 0.05);
}

.metric-item {
  flex: 1;
  text-align: center;
  border-right: 1px solid #F0F2F7;
  padding: 0 12px;
}

.metric-item:last-child {
  border-right: none;
}

.metric-value {
  font-size: 26px;
  font-weight: 700;
  color: #29365C;
  line-height: 1.2;
  margin-bottom: 6px;
}

.metric-label {
  font-size: 13px;
  color: #667085;
}

/* Dashboard Row */
.dashboard-row {
  display: flex;
  gap: 20px;
  margin-bottom: 24px;
}

.dashboard-row.bottom-row {
  margin-bottom: 0;
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
  gap: 12px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #29365C;
  margin: 0 0 16px 0;
}

.time-tabs {
  display: flex;
  gap: 4px;
  background: #F5F7FC;
  border-radius: 8px;
  padding: 4px;
}

.tab-btn {
  border: none;
  background: transparent;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  color: #667085;
  cursor: pointer;
  transition: all 0.15s ease;
  font-weight: 500;
  white-space: nowrap;
}

.tab-btn.active {
  background: #526FE8;
  color: #ffffff;
}

.chart-wrap {
  height: 280px;
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
  color: #8A93A8;
}

/* Summary Bar（内容产出视图） */
.summary-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.summary-pill {
  flex: 1 1 120px;
  background: #F5F7FC;
  border-radius: 10px;
  padding: 14px;
  text-align: center;
}

.summary-pill-label {
  font-size: 12px;
  color: #8A93A8;
  margin-bottom: 6px;
}

.summary-pill-value {
  font-size: 16px;
  font-weight: 700;
  color: #29365C;
}

/* Module Table */
.module-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.module-table th {
  text-align: left;
  font-size: 12px;
  font-weight: 500;
  color: #8A93A8;
  padding: 10px 8px;
  border-bottom: 1px solid #F0F2F7;
  white-space: nowrap;
}

.module-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #F5F7FC;
  color: #29365C;
  white-space: nowrap;
}

.module-table tr:last-child td {
  border-bottom: none;
}

.module-name {
  font-weight: 500;
}

.cell-sub {
  font-size: 12px;
  color: #8A93A8;
  margin-left: 4px;
}

.share-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.share-bar {
  width: 72px;
  height: 6px;
  background: #F5F7FC;
  border-radius: 3px;
  overflow: hidden;
  flex-shrink: 0;
}

.share-fill {
  height: 100%;
  border-radius: 3px;
  min-width: 0;
}

.share-text {
  font-size: 12px;
  color: #667085;
  min-width: 40px;
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
  color: #667085;
  flex: 1;
}

.legend-percent {
  font-size: 13px;
  font-weight: 600;
  color: #29365C;
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
  background: #526FE8;
  flex-shrink: 0;
}

.comment-avatar.blogger {
  background: rgba(255, 188, 104, 0.25);
  color: #8F6414;
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
  color: #29365C;
}

.blogger-tag {
  font-size: 11px;
  font-weight: 600;
  color: #8EA5FF;
  background: rgba(142, 165, 255, 0.1);
  padding: 1px 6px;
  border-radius: 4px;
}

.comment-time {
  font-size: 12px;
  color: #8A93A8;
  margin-left: auto;
}

.comment-text {
  font-size: 13px;
  color: #667085;
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
  color: #8A93A8;
}

/* Bottom Row */
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
  grid-template-columns: 40px 1fr auto auto auto;
  align-items: center;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid #F5F7FC;
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
  background: #F5F7FC;
  color: #8A93A8;
}

.rank-badge.rank-1 {
  background: rgba(183, 121, 31, 0.12);
  color: #B7791F;
}

.rank-badge.rank-2 {
  background: rgba(102, 112, 133, 0.12);
  color: #667085;
}

.rank-badge.rank-3 {
  background: rgba(183, 121, 31, 0.1);
  color: #B7791F;
}

.top-content-title {
  font-size: 14px;
  font-weight: 500;
  color: #29365C;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.top-content-type {
  font-size: 12px;
  white-space: nowrap;
}

.top-content-views {
  font-size: 13px;
  color: #667085;
  white-space: nowrap;
}

.top-content-comments {
  font-size: 13px;
  color: #667085;
  white-space: nowrap;
}

.list-empty {
  text-align: center;
  padding: 24px 0;
  font-size: 14px;
  color: #8A93A8;
}

/* Todo Reminder */
.todo-reminder {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  padding: 16px 20px;
  background: #F9F1E4;
  border-left: 4px solid #B7791F;
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
  background: rgba(183, 121, 31, 0.1);
}

.todo-count {
  font-size: 18px;
  font-weight: 700;
  color: #B7791F;
}

.todo-text {
  font-size: 14px;
  color: #8F6414;
}

.todo-arrow {
  font-size: 14px;
  color: #B7791F;
}

/* Skeleton */
.dashboard-skeleton {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.skeleton-strip {
  height: 96px;
  background: #F5F7FC;
  border-radius: 12px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-row {
  display: flex;
  gap: 20px;
}

.skeleton-block {
  background: #F5F7FC;
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
  color: #8A93A8;
}

.empty-text {
  font-size: 15px;
  color: #8A93A8;
  margin: 0;
}

/* Responsive */
@media (max-width: 1199px) {
  .dashboard {
    padding: 16px;
  }

  .dashboard-row {
    flex-direction: column;
  }

  .dashboard-col-left,
  .dashboard-col-right {
    flex: 1 1 auto;
    max-width: 100%;
  }

  .skeleton-row {
    flex-direction: column;
  }

  .skeleton-left,
  .skeleton-right {
    flex: 1 1 auto;
    max-width: 100%;
  }

  .metric-strip {
    flex-wrap: wrap;
    gap: 16px;
  }

  .metric-item {
    flex: 1 1 30%;
    border-right: none;
    text-align: left;
  }
}

@media (max-width: 767px) {
  .dashboard {
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

  .time-tabs {
    flex-wrap: wrap;
  }

  .chart-wrap {
    height: 220px;
  }

  .metric-strip {
    padding: 16px 12px;
  }

  .metric-value {
    font-size: 20px;
  }

  .col-week,
  .col-share {
    display: none;
  }

  .top-content-item {
    grid-template-columns: 28px 1fr auto;
    gap: 8px;
  }

  .top-content-type,
  .top-content-comments {
    display: none;
  }

  .todo-reminder {
    flex-direction: column;
    gap: 8px;
  }
}
</style>
