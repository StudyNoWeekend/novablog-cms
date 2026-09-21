<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">访问统计</h1>
      <a-button type="primary" :loading="loading" @click="fetchAccessStats">
        <template #icon><ReloadOutlined /></template>
        刷新
      </a-button>
    </div>

    <a-card class="table-card">
      <div class="toolbar">
        <a-input-search
          v-model:value="searchIP"
          placeholder="搜索 IP 地址"
          allow-clear
          enter-button
          style="max-width: 280px"
          @search="handleSearch"
        />
        <a-input-search
          v-model:value="searchRegion"
          placeholder="搜索地区"
          allow-clear
          enter-button
          style="max-width: 280px"
          @search="handleSearch"
        />
      </div>

      <a-table
        :columns="columns"
        :data-source="accessStats"
        :loading="loading"
        :pagination="pagination"
        row-key="ip"
        :scroll="{ x: 700 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'region'">
            <span v-if="record.region">{{ record.region }}</span>
            <span v-else class="text-muted">未知</span>
          </template>
          <template v-if="column.key === 'total_count'">
            <a-tag color="blue">{{ record.total_count }}</a-tag>
          </template>
          <template v-if="column.key === 'error_count'">
            <a-tag :color="record.error_count > 0 ? 'error' : 'success'">
              {{ record.error_count }}
            </a-tag>
          </template>
          <template v-if="column.key === 'last_access_at'">
            {{ formatDateTime(record.last_access_at) }}
          </template>
          <template v-if="column.key === 'action'">
            <a-button type="link" danger size="small" @click="openBlockModal(record.ip)">
              封禁
            </a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="blockModalOpen"
      title="封禁 IP"
      :confirm-loading="blockSubmitting"
      ok-text="确认封禁"
      cancel-text="取消"
      @ok="handleBlockSubmit"
    >
      <a-form layout="vertical">
        <a-form-item label="IP 地址">
          <a-input v-model:value="blockForm.ip" disabled />
        </a-form-item>
        <a-form-item label="封禁原因">
          <a-textarea
            v-model:value="blockForm.reason"
            placeholder="请输入封禁原因"
            :rows="3"
            :maxlength="255"
            show-count
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { message, type TableColumnsType } from 'ant-design-vue'
import { ReloadOutlined } from '@ant-design/icons-vue'
import { getAccessStatsAPI, createBlacklistAPI } from '@/api/security'
import type { IPAccessStats } from '@/types/security'

const loading = ref(false)
const accessStats = ref<IPAccessStats[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchIP = ref('')
const searchRegion = ref('')

const pagination = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showTotal: (t: number) => `共 ${t} 条`,
  showSizeChanger: true,
}))

const columns: TableColumnsType<IPAccessStats> = [
  { title: 'IP 地址', dataIndex: 'ip', key: 'ip', width: 160 },
  { title: '地区', dataIndex: 'region', key: 'region', width: 220 },
  { title: '总请求数', dataIndex: 'total_count', key: 'total_count', width: 120, align: 'center' },
  { title: '错误次数', dataIndex: 'error_count', key: 'error_count', width: 120, align: 'center' },
  { title: '最近访问时间', dataIndex: 'last_access_at', key: 'last_access_at', width: 190 },
  { title: '操作', key: 'action', width: 100, fixed: 'right' },
]

async function fetchAccessStats() {
  loading.value = true
  try {
    const data = await getAccessStatsAPI({
      page: page.value,
      page_size: pageSize.value,
      ip: searchIP.value,
      region: searchRegion.value,
    })
    accessStats.value = data.list
    total.value = data.total
  } catch {
    accessStats.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  fetchAccessStats()
}

function handleTableChange(paginationState: { current?: number; pageSize?: number }) {
  page.value = paginationState.current ?? 1
  pageSize.value = paginationState.pageSize ?? 10
  fetchAccessStats()
}

const blockModalOpen = ref(false)
const blockSubmitting = ref(false)
const blockForm = ref({ ip: '', reason: '' })

function openBlockModal(ip: string) {
  blockForm.value = { ip, reason: '' }
  blockModalOpen.value = true
}

async function handleBlockSubmit() {
  if (!blockForm.value.reason.trim()) {
    message.warning('请输入封禁原因')
    return
  }
  blockSubmitting.value = true
  try {
    await createBlacklistAPI({
      ip: blockForm.value.ip,
      reason: blockForm.value.reason.trim(),
    })
    message.success('IP 封禁成功')
    blockModalOpen.value = false
  } catch {
    // 错误由请求拦截器处理
  } finally {
    blockSubmitting.value = false
  }
}

function formatDateTime(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

onMounted(() => {
  fetchAccessStats()
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

.table-card {
  border-radius: var(--border-radius-lg, 12px);
}

.toolbar {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.text-muted {
  color: var(--text-muted, #94a3b8);
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

  .toolbar {
    justify-content: flex-start;
  }
}
</style>
