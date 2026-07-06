<template>
  <div class="dashboard">
    <!-- 顶部统计卡片 -->
    <div class="stat-grid">
      <div v-for="(s, i) in stats" :key="i" class="stat-card">
        <div class="stat-card-top">
          <div class="stat-card-label">
            <span class="stat-icon" :class="s.iconBg">
              <component :is="s.icon" />
            </span>
            <span class="stat-name">{{ s.label }}</span>
          </div>
          <div class="mini-chart">
            <div v-for="(h, idx) in s.miniBars" :key="idx" class="mini-bar" :style="{ height: h + '%' }" />
          </div>
        </div>
        <div class="stat-card-value">{{ s.value }}</div>
        <div class="stat-card-bottom">
          <span class="stat-compare">{{ s.compare }}</span>
          <span class="stat-badge">{{ s.trend }}</span>
        </div>
      </div>
    </div>

    <!-- 中间两列布局 -->
    <div class="dashboard-row">
      <!-- 左列：整体趋势 -->
      <div class="dashboard-col-left">
        <div class="card">
          <div class="chart-header">
            <div class="chart-header-left">
              <h3 class="card-title">整体趋势</h3>
              <div class="chart-metric">
                <span class="metric-value">¥23,8461</span>
                <span class="metric-trend">+218.23</span>
              </div>
            </div>
            <div class="time-tabs">
              <button
                v-for="t in timeTabs"
                :key="t"
                :class="['tab-btn', { active: activeTab === t }]"
                @click="activeTab = t"
              >
                {{ t }}
              </button>
            </div>
          </div>
          <div class="chart-wrap">
            <v-chart class="trend-chart" :option="trendOption" autoresize />
          </div>
          <div class="summary-bar">
            <div v-for="(item, idx) in summaryStats" :key="idx" class="summary-pill">
              <div class="summary-pill-label">{{ item.label }}</div>
              <div class="summary-pill-value">{{ item.value }}</div>
            </div>
          </div>
          <div class="data-table">
            <div class="table-row table-head">
              <div class="cell">月份</div>
              <div class="cell">7天阅读</div>
              <div class="cell">30天阅读</div>
              <div class="cell">本月阅读</div>
              <div class="cell">本年阅读</div>
              <div class="cell">增长率</div>
            </div>
            <div v-for="(row, idx) in tableData" :key="idx" class="table-row">
              <div class="cell">{{ row.month }}</div>
              <div class="cell">{{ row.d7 }}</div>
              <div class="cell">{{ row.d30 }}</div>
              <div class="cell">{{ row.monthly }}</div>
              <div class="cell">{{ row.yearly }}</div>
              <div class="cell">
                <span :class="['growth-badge', row.growth >= 0 ? 'up' : 'down']">
                  {{ row.growth >= 0 ? '+' : '' }}{{ row.growth }}%
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 右列：内容分布 + 互动历史 -->
      <div class="dashboard-col-right">
        <div class="card">
          <div class="card-header-between">
            <h3 class="card-title">内容分布</h3>
            <span class="dropdown-trigger">一月 ▼</span>
          </div>
          <div class="pie-wrap">
            <v-chart class="pie-chart" :option="pieOption" autoresize />
          </div>
          <div class="pie-legend">
            <div v-for="(item, idx) in pieLegend" :key="idx" class="legend-row">
              <span class="legend-dot" :style="{ background: item.color }" />
              <span class="legend-name">{{ item.name }}</span>
              <span class="legend-percent">{{ item.value }}%</span>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-header-between">
            <h3 class="card-title">互动历史</h3>
            <span class="menu-dots">⋯</span>
          </div>
          <div class="interaction-list">
            <div v-for="(item, idx) in interactions" :key="idx" class="interaction-item">
              <div class="interaction-avatar" :style="{ backgroundColor: item.avatarColor }">
                {{ item.initials }}
              </div>
              <div class="interaction-body">
                <div class="interaction-title">{{ item.name }}</div>
                <div class="interaction-desc">{{ item.date }} · {{ item.action }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  FileTextOutlined,
  EyeOutlined,
  ReadOutlined,
  MessageOutlined,
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

use([
  CanvasRenderer,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
])

const stats = [
  {
    label: '文章总数',
    value: '128',
    compare: '+12 篇本周新增',
    trend: '+14.67%',
    icon: FileTextOutlined,
    iconBg: 'bg-blue',
    miniBars: [40, 60, 35, 75, 50, 85, 65],
  },
  {
    label: '总访问量',
    value: '89,421',
    compare: '+2,560 本周新增',
    trend: '+0.67%',
    icon: EyeOutlined,
    iconBg: 'bg-green',
    miniBars: [55, 45, 70, 40, 60, 50, 80],
  },
  {
    label: '平均阅读',
    value: '892',
    compare: '+218 本周新增',
    trend: '+1.4%',
    icon: ReadOutlined,
    iconBg: 'bg-orange',
    miniBars: [30, 50, 40, 65, 45, 70, 55],
  },
  {
    label: '评论总数',
    value: '456',
    compare: '+560 本周新增',
    trend: '+0.52%',
    icon: MessageOutlined,
    iconBg: 'bg-purple',
    miniBars: [45, 35, 60, 50, 75, 55, 65],
  },
]

const timeTabs = ['12个月', '6个月', '30天', '7天']
const activeTab = ref('12个月')

const trendOption = computed(() => ({
  grid: { left: 0, right: 0, top: 10, bottom: 0, containLabel: true },
  tooltip: { trigger: 'axis' },
  xAxis: {
    type: 'category',
    data: ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月'],
    axisLine: { show: false },
    axisTick: { show: false },
    axisLabel: { color: '#94a3b8' },
  },
  yAxis: {
    type: 'value',
    splitLine: { lineStyle: { type: 'dashed', color: '#f1f5f9' } },
    axisLabel: { show: false },
  },
  series: [
    {
      type: 'bar',
      data: [4200, 5500, 4800, 6200, 5100, 7300, 6800, 5900, 7100, 6500, 7800, 8200],
      barWidth: '40%',
      itemStyle: {
        borderRadius: [4, 4, 0, 0],
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: '#4a6cf7' },
            { offset: 1, color: '#8faafa' },
          ],
        },
      },
    },
  ],
}))

const summaryStats = [
  { label: '7天阅读', value: '12,847' },
  { label: '30天阅读', value: '45,231' },
  { label: '本月阅读', value: '89,421' },
  { label: '本年阅读', value: '1,024,893' },
]

const tableData = [
  { month: '2024-06', d7: '12,847', d30: '45,231', monthly: '89,421', yearly: '1,024,893', growth: 14.67 },
  { month: '2024-05', d7: '11,230', d30: '42,100', monthly: '78,000', yearly: '935,472', growth: 8.32 },
  { month: '2024-04', d7: '10,500', d30: '38,900', monthly: '72,000', yearly: '857,472', growth: -2.14 },
  { month: '2024-03', d7: '13,200', d30: '41,500', monthly: '75,000', yearly: '785,472', growth: 5.67 },
  { month: '2024-02', d7: '9,800', d30: '35,200', monthly: '68,000', yearly: '710,472', growth: 12.3 },
  { month: '2024-01', d7: '11,000', d30: '39,000', monthly: '70,000', yearly: '642,472', growth: 3.45 },
]

const pieOption = computed(() => ({
  tooltip: { trigger: 'item' },
  series: [
    {
      type: 'pie',
      radius: ['55%', '75%'],
      center: ['50%', '50%'],
      avoidLabelOverlap: false,
      label: { show: false },
      emphasis: { label: { show: false } },
      labelLine: { show: false },
      data: [
        { value: 65, name: '文章', itemStyle: { color: '#4a6cf7' } },
        { value: 20.5, name: '摄影作品', itemStyle: { color: '#10b981' } },
        { value: 14.5, name: '视频', itemStyle: { color: '#f59e0b' } },
      ],
    },
  ],
}))

const pieLegend = [
  { name: '文章', value: 65, color: '#4a6cf7' },
  { name: '摄影作品', value: 20.5, color: '#10b981' },
  { name: '视频', value: 14.5, color: '#f59e0b' },
]

const interactions = [
  { name: '张三', date: '2024-06-12', action: '评论了文章', initials: '张', avatarColor: '#4a6cf7' },
  { name: '李四', date: '2024-06-11', action: '点赞了作品', initials: '李', avatarColor: '#10b981' },
  { name: '王五', date: '2024-06-11', action: '收藏了文章', initials: '王', avatarColor: '#f59e0b' },
  { name: '赵六', date: '2024-06-10', action: '评论了视频', initials: '赵', avatarColor: '#ef4444' },
  { name: '孙七', date: '2024-06-10', action: '点赞了文章', initials: '孙', avatarColor: '#8b5cf6' },
  { name: '周八', date: '2024-06-09', action: '评论了作品', initials: '周', avatarColor: '#06b6d4' },
  { name: '吴九', date: '2024-06-08', action: '收藏了视频', initials: '吴', avatarColor: '#ec4899' },
  { name: '郑十', date: '2024-06-08', action: '点赞了文章', initials: '郑', avatarColor: '#84cc16' },
]
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

.stat-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: nowrap;
  gap: 12px;
  margin-bottom: 16px;
}

.stat-card-label {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
  min-width: 0;
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

.mini-chart {
  display: flex;
  align-items: flex-end;
  gap: 3px;
  height: 28px;
  flex-shrink: 0;
}

.mini-bar {
  width: 4px;
  min-height: 2px;
  background: #dbeafe;
  border-radius: 2px;
  transition: height 0.3s ease;
}

.stat-card:hover .mini-bar {
  background: #93bbfc;
}

.stat-card-value {
  font-size: 28px;
  font-weight: 700;
  color: #1e293b;
  line-height: 1.2;
  margin-bottom: 12px;
}

.stat-card-bottom {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-compare {
  font-size: 13px;
  color: #64748b;
}

.stat-badge {
  font-size: 12px;
  font-weight: 600;
  color: #10b981;
  background: rgba(16, 185, 129, 0.08);
  padding: 2px 8px;
  border-radius: 6px;
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
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}

.chart-header-left {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.chart-metric {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.metric-value {
  font-size: 24px;
  font-weight: 700;
  color: #1e293b;
}

.metric-trend {
  font-size: 14px;
  font-weight: 600;
  color: #10b981;
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

/* Summary Bar */
.summary-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid #f1f5f9;
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

/* Data Table */
.data-table {
  display: flex;
  flex-direction: column;
}

.table-row {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1fr 1fr 1fr 0.8fr;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f8fafc;
}

.table-head {
  padding: 8px 0;
  border-bottom: 1px solid #f1f5f9;
}

.table-head .cell {
  font-size: 12px;
  color: #94a3b8;
  font-weight: 500;
}

.cell {
  font-size: 13px;
  color: #475569;
}

.growth-badge {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 6px;
}

.growth-badge.up {
  color: #10b981;
  background: rgba(16, 185, 129, 0.08);
}

.growth-badge.down {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.08);
}

/* Right Column */
.card-header-between {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.dropdown-trigger {
  font-size: 13px;
  color: #64748b;
  cursor: pointer;
}

.menu-dots {
  font-size: 18px;
  color: #94a3b8;
  cursor: pointer;
  letter-spacing: 1px;
}

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

/* Interaction List */
.interaction-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.interaction-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.interaction-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  color: #ffffff;
  flex-shrink: 0;
}

.interaction-body {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.interaction-title {
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
}

.interaction-desc {
  font-size: 12px;
  color: #94a3b8;
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

  .time-tabs {
    flex-wrap: wrap;
  }

  .tab-btn {
    padding: 6px 10px;
    font-size: 12px;
  }

  .data-table {
    overflow-x: auto;
  }

  .table-row {
    min-width: 600px;
  }
}
</style>
