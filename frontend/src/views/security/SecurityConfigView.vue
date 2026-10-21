<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('security.blacklist.title') }}</h1>
      <a-button type="primary" @click="openAddModal">
        <template #icon><PlusOutlined /></template>
        {{ $t('security.blacklist.add') }}
      </a-button>
    </div>

    <a-card class="table-card">
      <div class="toolbar">
        <a-input-search
          v-model:value="searchKeyword"
          :placeholder="$t('security.blacklist.searchPlaceholder')"
          allow-clear
          enter-button
          style="max-width: 320px"
          @search="handleSearch"
        />
      </div>

      <a-table
        :columns="columns"
        :data-source="blacklist"
        :loading="loading"
        :pagination="pagination"
        row-key="id"
        :scroll="{ x: 700 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'created_at'">
            {{ formatDateTime(record.created_at) }}
          </template>
          <template v-if="column.key === 'action'">
            <a-button type="link" size="small" @click="openEditModal(record)">
              {{ $t('security.blacklist.editRemark') }}
            </a-button>
            <a-popconfirm
              :title="$t('security.blacklist.deleteConfirm')"
              :ok-text="$t('common.actions.confirm')"
              :cancel-text="$t('common.actions.cancel')"
              @confirm="handleDelete(record.id)"
            >
              <a-button type="link" size="small" danger>{{ $t('common.actions.delete') }}</a-button>
            </a-popconfirm>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="modalOpen"
      :title="isEdit ? $t('security.blacklist.editTitle') : $t('security.blacklist.add')"
      :confirm-loading="submitting"
      :ok-text="$t('common.actions.confirm')"
      :cancel-text="$t('common.actions.cancel')"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item :label="$t('security.blacklist.form.ip')">
          <a-input
            v-model:value="form.ip"
            :placeholder="$t('security.blacklist.form.ipPlaceholder')"
            :disabled="isEdit"
          />
        </a-form-item>
        <a-form-item :label="$t('security.blacklist.form.reason')">
          <a-textarea
            v-model:value="form.reason"
            :placeholder="$t('security.blacklist.form.reasonPlaceholder')"
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
import { PlusOutlined } from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import {
  getBlacklistsAPI,
  createBlacklistAPI,
  updateBlacklistAPI,
  deleteBlacklistAPI,
} from '@/api/security'
import type { BlacklistItem } from '@/types/security'
import { formatDateTime } from '@/utils/datetime'

const { t } = useI18n()

const loading = ref(false)
const blacklist = ref<BlacklistItem[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchKeyword = ref('')

const pagination = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showTotal: (count: number) => t('common.table.total', { total: count }),
  showSizeChanger: true,
}))

// 表格列定义（computed 包裹，切换语言时刷新）
const columns = computed<TableColumnsType<BlacklistItem>>(() => [
  { title: t('security.blacklist.columns.ip'), dataIndex: 'ip', key: 'ip', width: 180 },
  { title: t('security.blacklist.columns.reason'), dataIndex: 'reason', key: 'reason', ellipsis: true },
  { title: t('common.field.createTime'), dataIndex: 'created_at', key: 'created_at', width: 200 },
  { title: t('common.table.actions'), key: 'action', width: 160, fixed: 'right' },
])

async function fetchBlacklist() {
  loading.value = true
  try {
    const data = await getBlacklistsAPI({
      page: page.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value,
    })
    blacklist.value = data.list
    total.value = data.total
  } catch {
    blacklist.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  fetchBlacklist()
}

function handleTableChange(paginationState: { current?: number; pageSize?: number }) {
  page.value = paginationState.current ?? 1
  pageSize.value = paginationState.pageSize ?? 10
  fetchBlacklist()
}

const modalOpen = ref(false)
const submitting = ref(false)
const isEdit = ref(false)
const editId = ref<string>('')
const form = ref({ ip: '', reason: '' })

function openAddModal() {
  isEdit.value = false
  editId.value = ''
  form.value = { ip: '', reason: '' }
  modalOpen.value = true
}

function openEditModal(record: BlacklistItem) {
  isEdit.value = true
  editId.value = record.id
  form.value = { ip: record.ip, reason: record.reason }
  modalOpen.value = true
}

async function handleSubmit() {
  const ip = form.value.ip.trim()
  const reason = form.value.reason.trim()

  if (!ip) {
    message.warning(t('security.blacklist.message.ipRequired'))
    return
  }

  submitting.value = true
  try {
    if (isEdit.value) {
      await updateBlacklistAPI(editId.value, { reason })
      message.success(t('security.blacklist.message.remarkUpdated'))
    } else {
      await createBlacklistAPI({ ip, reason })
      message.success(t('security.blacklist.message.addSuccess'))
    }
    modalOpen.value = false
    await fetchBlacklist()
  } catch {
    // 错误由请求拦截器处理
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: string) {
  try {
    await deleteBlacklistAPI(id)
    message.success(t('common.message.deleteSuccess'))
    await fetchBlacklist()
  } catch {
    // 错误由请求拦截器处理
  }
}

onMounted(() => {
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
  color: var(--text-primary, #29365C);
  margin: 0;
}

.table-card {
  border-radius: var(--border-radius-lg, 12px);
}

.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
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
