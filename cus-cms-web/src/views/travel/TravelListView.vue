<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">旅行攻略</h1>
      <a-button
        type="primary"
        aria-label="新建攻略"
        @click="handleCreate"
      >
        <PlusOutlined />
        新建攻略
      </a-button>
    </div>

    <div class="travel-stats" role="region" aria-label="攻略统计">
      <div
        v-for="item in statItems"
        :key="item.label"
        class="stat-card"
      >
        <div class="stat-icon" :style="{ background: item.iconBg, color: item.iconColor }">
          <component :is="item.icon" />
        </div>
        <div class="stat-body">
          <div class="stat-value">{{ item.value }}</div>
          <div class="stat-label">{{ item.label }}</div>
        </div>
      </div>
    </div>

    <!-- 分类标签栏 -->
    <CategoryBar type="travel" @select="handleCategorySelect" />

    <div class="travel-filter-bar">
      <TravelFilterBar v-model="filters" />
    </div>

    <div class="travel-list-section">
      <div
        v-if="loading"
        class="travel-grid travel-grid-skeleton"
        role="status"
        aria-label="加载中"
      >
        <div
          v-for="n in 6"
          :key="`skeleton-${n}`"
          class="skeleton-card"
        >
          <a-skeleton active :paragraph="{ rows: 4 }" />
        </div>
      </div>

      <div
        v-else-if="filteredGuides.length > 0"
        :key="listKey"
        class="travel-grid"
      >
        <TravelCard
          v-for="(guide, index) in filteredGuides"
          :key="guide.id"
          :guide="guide"
          :index="index"
          @view="handleView"
          @edit="handleEdit"
          @delete="handleDelete"
          @statusChange="handleStatusChange"
        />
      </div>

      <a-empty
        v-else
        description="没有找到匹配的攻略"
      />
    </div>

    <!-- 分页 -->
    <div v-if="total > pageSize" class="pagination-wrapper">
      <a-pagination
        :current="page"
        :page-size="pageSize"
        :total="total"
        :show-size-changer="true"
        :show-total="(total: number) => `共 ${total} 篇`"
        @change="onPageChange"
        @show-size-change="handlePageSizeChange"
      />
    </div>

    <TravelDetailDrawer
      v-model:open="drawerOpen"
      :guide="selectedGuide"
      @close="handleDrawerClose"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  FileTextOutlined,
  CalendarOutlined,
  EyeOutlined,
  StarOutlined,
} from '@ant-design/icons-vue'
import TravelCard from '@/components/travel/TravelCard.vue'
import TravelFilterBar from '@/components/travel/TravelFilterBar.vue'
import TravelDetailDrawer from '@/components/travel/TravelDetailDrawer.vue'
import CategoryBar from '@/components/common/CategoryBar.vue'
import { travelApi, type TravelGuideListParams } from '@/api/travel'
import { usePagination } from '@/composables/usePagination'
import { TravelStatus, type TravelGuide, type TravelFilters } from '@/types/travel'

const guides = ref<TravelGuide[]>([])

const { page, pageSize, total, handlePageChange } = usePagination(20)

function onPageChange(newPage: number, newPageSize?: number) {
  handlePageChange(newPage, newPageSize)
  fetchGuides()
}

function handlePageSizeChange(_p: number, size: number) {
  pageSize.value = size
  page.value = 1
  fetchGuides()
}

const filters = reactive<Required<TravelFilters>>({
  keyword: '',
  region: 'all',
  days: 'all',
  sort: 'latest',
  categoryId: '',
})

const filterChangeCounter = ref(0)
const drawerOpen = ref(false)
const selectedGuide = ref<TravelGuide | null>(null)
const loading = ref(false)

const router = useRouter()

const filteredGuides = computed(() => guides.value)

const listKey = computed(() => `travel-list-${filterChangeCounter.value}`)

const stats = computed(() => {
  const totalGuides = total.value
  const now = new Date()
  const currentMonthPrefix = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  const thisMonth = guides.value.filter((g) => g.createdAt.startsWith(currentMonthPrefix)).length
  const totalViews = guides.value.reduce((sum, g) => sum + g.viewCount, 0)
  const avgRating = guides.value.length > 0
    ? (guides.value.reduce((sum, g) => sum + g.rating, 0) / guides.value.length).toFixed(1)
    : '0.0'

  return { total: totalGuides, thisMonth, totalViews, avgRating }
})

const statItems = computed(() => [
  {
    label: '攻略总数',
    value: stats.value.total,
    icon: FileTextOutlined,
    iconBg: 'rgba(74, 108, 247, 0.1)',
    iconColor: '#4a6cf7',
  },
  {
    label: '本月新增',
    value: stats.value.thisMonth,
    icon: CalendarOutlined,
    iconBg: 'rgba(16, 185, 129, 0.1)',
    iconColor: '#10b981',
  },
  {
    label: '总浏览量',
    value: stats.value.totalViews.toLocaleString('zh-CN'),
    icon: EyeOutlined,
    iconBg: 'rgba(245, 158, 11, 0.1)',
    iconColor: '#f59e0b',
  },
  {
    label: '平均评分',
    value: stats.value.avgRating,
    icon: StarOutlined,
    iconBg: 'rgba(239, 68, 68, 0.1)',
    iconColor: '#ef4444',
  },
])

let debounceTimer: ReturnType<typeof setTimeout> | null = null
watch(
  () => ({ ...filters }),
  () => {
    filterChangeCounter.value++
    page.value = 1
    if (debounceTimer) clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
      fetchGuides()
    }, 300)
  },
  { deep: true },
)

async function fetchGuides() {
  loading.value = true
  try {
    const params: TravelGuideListParams = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (filters.keyword) params.keyword = filters.keyword
    if (filters.region && filters.region !== 'all') params.region = filters.region
    if (filters.days && filters.days !== 'all') params.days_range = filters.days
    if (filters.sort) params.sort = filters.sort
    if (filters.categoryId) params.category_id = filters.categoryId

    const res = await travelApi.getList(params)
    guides.value = res.list
    total.value = res.total || 0
  } catch {
    // 错误已由拦截器处理
  } finally {
    loading.value = false
  }
}

function handleCategorySelect(id: string | undefined) {
  filters.categoryId = id ?? ''
}

onMounted(() => {
  fetchGuides()
})

function handleCreate() {
  router.push('/travels/create')
}

async function handleView(id: string) {
  try {
    const detail = await travelApi.getDetail(id)
    selectedGuide.value = detail
    drawerOpen.value = true
  } catch {
    // 错误已由拦截器处理
  }
}

function handleEdit(id: string) {
  router.push(`/travels/${id}/edit`)
}

function handleDelete(id: string) {
  const guide = guides.value.find((g) => g.id === id)
  if (!guide) return

  Modal.confirm({
    title: '确认删除',
    content: `确定要删除「${guide.title}」吗？删除后无法恢复。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await travelApi.remove(id)
        message.success('删除成功')
        fetchGuides()
      } catch {
        // 错误已由拦截器处理
      }
    },
  })
}

async function handleStatusChange(id: string, status: TravelStatus) {
  try {
    await travelApi.updateStatus(id, status)
    message.success('状态更新成功')
    fetchGuides()
  } catch {
    // 错误已由拦截器处理
  }
}

function handleDrawerClose() {
  selectedGuide.value = null
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0;
}

.travel-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
  animation: fade-in-up 400ms cubic-bezier(0.4, 0, 0.2, 1) forwards;
  opacity: 0;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
  transition:
    transform var(--transition-base, 250ms cubic-bezier(0.4, 0, 0.2, 1)),
    box-shadow var(--transition-base, 250ms cubic-bezier(0.4, 0, 0.2, 1));
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-dropdown, 0 4px 16px rgba(0, 0, 0, 0.08));
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 12px;
  font-size: 24px;
  flex-shrink: 0;
}

.stat-body {
  min-width: 0;
}

.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary, #1e293b);
  line-height: 1.3;
}

.stat-label {
  font-size: 13px;
  color: var(--text-secondary, #64748b);
  margin-top: 2px;
}

.travel-filter-bar {
  margin-bottom: 24px;
  animation: fade-in-up 400ms cubic-bezier(0.4, 0, 0.2, 1) 80ms forwards;
  opacity: 0;
}

.travel-list-section {
  min-height: 200px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 32px;
}

.travel-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.travel-grid-skeleton {
  opacity: 1;
}

.skeleton-card {
  padding: 16px;
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
}

@keyframes fade-in-up {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 1024px) {
  .travel-stats {
    grid-template-columns: repeat(2, 1fr);
  }

  .travel-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .travel-stats {
    grid-template-columns: 1fr;
  }

  .travel-grid {
    grid-template-columns: 1fr;
  }

  .stat-card {
    padding: 16px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .travel-stats,
  .travel-filter-bar {
    animation: none;
    opacity: 1;
  }

  .stat-card {
    transform: none !important;
    transition: none;
  }
}
</style>
