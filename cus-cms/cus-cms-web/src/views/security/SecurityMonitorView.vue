<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">安全监控</h1>
      <a-button type="primary" :loading="refreshing" @click="refreshAll">
        刷新
      </a-button>
    </div>

    <!-- 统计卡片 -->
    <a-row :gutter="16" class="stats-row">
      <a-col :xs="24" :sm="12">
        <a-card class="stat-card">
          <a-statistic
            title="当前被封 IP 数"
            :value="stats?.blocked_ip_count ?? 0"
            :value-style="{ color: '#ef4444' }"
          />
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="12">
        <a-card class="stat-card">
          <a-statistic
            title="今日限流触发次数"
            :value="stats?.today_rate_limit_count ?? 0"
            :value-style="{ color: '#f59e0b' }"
          />
        </a-card>
      </a-col>
    </a-row>

    <!-- 限流趋势图表 -->
    <a-card title="近 7 天限流趋势" class="chart-card">
      <a-spin :spinning="statsLoading">
        <div v-if="stats && stats.daily_trend.length > 0" class="chart-wrap">
          <v-chart class="trend-chart" :option="trendOption" autoresize />
        </div>
        <a-empty v-else-if="!statsLoading" description="暂无趋势数据" />
      </a-spin>
    </a-card>

    <!-- Top 违规 IP -->
    <a-card v-if="stats && stats.top_violations.length > 0" title="违规 IP 排行" class="top-card">
      <a-table
        :columns="topColumns"
        :data-source="stats.top_violations"
        :pagination="false"
        row-key="ip_address"
        size="small"
      />
    </a-card>

    <!-- 黑名单列表 -->
    <a-card title="IP 黑名单" class="blacklist-card">
      <a-table
        :columns="blacklistColumns"
        :data-source="blacklist"
        :loading="blacklistLoading"
        :pagination="blacklistPagination"
        row-key="id"
        :scroll="{ x: 800 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'banned_at'">
            {{ formatDate(record.banned_at) }}
          </template>
          <template v-if="column.key === 'expires_at'">
            {{ formatDate(record.expires_at) }}
          </template>
          <template v-if="column.key === 'is_active'">
            <a-tag :color="record.is_active ? 'red' : 'default'">
              {{ record.is_active ? '封禁中' : '已解封' }}
            </a-tag>
          </template>
          <template v-if="column.key === 'action'">
            <a-popconfirm
              v-if="record.is_active"
              title="确定要解封该 IP 吗？"
              ok-text="确定"
              cancel-text="取消"
              @confirm="handleUnban(record.ip_address)"
            >
              <a-button type="link" size="small" danger>解封</a-button>
            </a-popconfirm>
            <span v-else style="color: #cbd5e1">-</span>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { message, type TableColumnsType } from 'ant-design-vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import { getSecurityStatsAPI, getBlacklistAPI, unbanIPAPI } from '@/api/security'
import type { SecurityStats, BlacklistItem, TopIP } from '@/types/security'

use([
  CanvasRenderer,
  LineChart,
  GridComponent,
  TooltipComponent,
])

// 统计数据
const statsLoading = ref(false)
const refreshing = ref(false)
const stats = ref<SecurityStats | null>(null)

// 黑名单数据
const blacklistLoading = ref(false)
const blacklist = ref<BlacklistItem[]>([])
const blacklistPage = ref(1)
const blacklistPageSize = ref(10)
const blacklistTotal = ref(0)

const blacklistPagination = computed(() => ({
  current: blacklistPage.value,
  pageSize: blacklistPageSize.value,
  total: blacklistTotal.value,
  showTotal: (total: number) => `共 ${total} 条`,
}))

const topColumns: TableColumnsType<TopIP> = [
  { title: 'IP 地址', dataIndex: 'ip_address', key: 'ip_address' },
  { title: '违规次数', dataIndex: 'count', key: 'count', width: 150 },
]

const blacklistColumns: TableColumnsType<BlacklistItem> = [
  { title: 'IP 地址', dataIndex: 'ip_address', key: 'ip_address', width: 160 },
  { title: '封禁原因', dataIndex: 'reason', key: 'reason', ellipsis: true },
  { title: '封禁时间', key: 'banned_at', width: 180 },
  { title: '预计解封时间', key: 'expires_at', width: 180 },
  { title: '状态', key: 'is_active', width: 110 },
  { title: '操作', key: 'action', width: 100, fixed: 'right' },
]

const trendOption = computed(() => {
  if (!stats.value) return {}
  const items = stats.value.daily_trend
  return {
    grid: { left: 8, right: 8, top: 20, bottom: 8, containLabel: true },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: items.map((i) => i.date.slice(5)),
      axisLine: { lineStyle: { color: '#e2e8f0' } },
      axisTick: { show: false },
      axisLabel: { color: '#94a3b8', fontSize: 12 },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { type: 'dashed', color: '#f1f5f9' } },
      axisLabel: { color: '#94a3b8' },
    },
    series: [
      {
        name: '限流次数',
        type: 'line',
        data: items.map((i) => i.count),
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        itemStyle: { color: '#4a6cf7' },
        lineStyle: { width: 2, color: '#4a6cf7' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(74, 108, 247, 0.15)' },
              { offset: 1, color: 'rgba(74, 108, 247, 0)' },
            ],
          },
        },
      },
    ],
  }
})

async function fetchStats() {
  statsLoading.value = true
  try {
    stats.value = await getSecurityStatsAPI()
  } catch {
    // 错误由请求拦截器处理
  } finally {
    statsLoading.value = false
  }
}

async function fetchBlacklist() {
  blacklistLoading.value = true
  try {
    const data = await getBlacklistAPI({
      page: blacklistPage.value,
      page_size: blacklistPageSize.value,
    })
    blacklist.value = data.list
    blacklistTotal.value = data.total
  } catch {
    blacklist.value = []
    blacklistTotal.value = 0
  } finally {
    blacklistLoading.value = false
  }
}

function handleTableChange(pagination: { current?: number; pageSize?: number }) {
  blacklistPage.value = pagination.current ?? 1
  blacklistPageSize.value = pagination.pageSize ?? 10
  fetchBlacklist()
}

async function handleUnban(ip: string) {
  try {
    await unbanIPAPI(ip)
    message.success('解封成功')
    await Promise.all([fetchBlacklist(), fetchStats()])
  } catch {
    // 错误由请求拦截器处理
  }
}

async function refreshAll() {
  refreshing.value = true
  await Promise.all([fetchStats(), fetchBlacklist()])
  refreshing.value = false
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

onMounted(() => {
  fetchStats()
  fetchBlacklist()
})
</script>

<style scoped>
.page-container {
  padding: 0;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0;
}

.stats-row {
  margin-bottom: 16px;
}

.stat-card {
  border-radius: var(--border-radius-lg, 12px);
}

.chart-card {
  border-radius: var(--border-radius-lg, 12px);
  margin-bottom: 16px;
}

.chart-wrap {
  height: 300px;
}

.trend-chart {
  width: 100%;
  height: 100%;
}

.top-card {
  border-radius: var(--border-radius-lg, 12px);
  margin-bottom: 16px;
}

.blacklist-card {
  border-radius: var(--border-radius-lg, 12px);
}

:deep(.ant-btn-primary) {
  cursor: pointer;
}

:deep(.ant-btn-link) {
  cursor: pointer;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }
}
</style>
