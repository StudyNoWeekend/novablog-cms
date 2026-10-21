<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('security.monitor.title') }}</h1>
      <a-button type="primary" :loading="loading" @click="fetchAccessStats">
        <template #icon><ReloadOutlined /></template>
        {{ $t('common.actions.refresh') }}
      </a-button>
    </div>
    <p class="page-desc">{{ $t('security.monitor.desc') }}</p>

    <a-card class="table-card">
      <div class="toolbar">
        <a-input-search
          v-model:value="searchIP"
          :placeholder="$t('security.monitor.searchIpPlaceholder')"
          allow-clear
          enter-button
          style="max-width: 280px"
          @search="handleSearch"
        />
        <a-input-search
          v-model:value="searchRegion"
          :placeholder="$t('security.monitor.searchRegionPlaceholder')"
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
            <span v-else class="text-muted">{{ $t('common.status.unknown') }}</span>
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
              {{ $t('security.monitor.block') }}
            </a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="blockModalOpen"
      :title="$t('security.monitor.blockTitle')"
      :confirm-loading="blockSubmitting"
      :ok-text="$t('security.monitor.confirmBlock')"
      :cancel-text="$t('common.actions.cancel')"
      @ok="handleBlockSubmit"
    >
      <a-form layout="vertical">
        <a-form-item :label="$t('security.blacklist.form.ip')">
          <a-input v-model:value="blockForm.ip" disabled />
        </a-form-item>
        <a-form-item :label="$t('security.monitor.form.reason')">
          <a-textarea
            v-model:value="blockForm.reason"
            :placeholder="$t('security.monitor.form.reasonPlaceholder')"
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
import { useI18n } from 'vue-i18n'
import { getAccessStatsAPI, createBlacklistAPI } from '@/api/security'
import type { IPAccessStats } from '@/types/security'
import { formatDateTime } from '@/utils/datetime'

const { t } = useI18n()

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
  showTotal: (count: number) => t('common.table.total', { total: count }),
  showSizeChanger: true,
}))

// 表格列定义（computed 包裹，切换语言时刷新）
const columns = computed<TableColumnsType<IPAccessStats>>(() => [
  { title: t('security.monitor.columns.ip'), dataIndex: 'ip', key: 'ip', width: 160 },
  { title: t('security.monitor.columns.region'), dataIndex: 'region', key: 'region', width: 220 },
  { title: t('security.monitor.columns.totalCount'), dataIndex: 'total_count', key: 'total_count', width: 120, align: 'center' },
  { title: t('security.monitor.columns.errorCount'), dataIndex: 'error_count', key: 'error_count', width: 120, align: 'center' },
  { title: t('security.monitor.columns.lastAccess'), dataIndex: 'last_access_at', key: 'last_access_at', width: 190 },
  { title: t('common.table.actions'), key: 'action', width: 100, fixed: 'right' },
])

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
    message.warning(t('security.monitor.message.reasonRequired'))
    return
  }
  blockSubmitting.value = true
  try {
    await createBlacklistAPI({
      ip: blockForm.value.ip,
      reason: blockForm.value.reason.trim(),
    })
    message.success(t('security.monitor.message.blockSuccess'))
    blockModalOpen.value = false
  } catch {
    // 错误由请求拦截器处理
  } finally {
    blockSubmitting.value = false
  }
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
  color: var(--text-primary, #29365C);
  margin: 0;
}

.page-desc {
  font-size: 13px;
  color: var(--text-secondary, #667085);
  margin: -16px 0 16px 0;
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
  color: var(--text-muted, #8A93A8);
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
