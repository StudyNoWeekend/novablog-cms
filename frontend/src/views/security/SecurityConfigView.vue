<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">黑名单管理</h1>
      <a-button type="primary" @click="openAddModal">
        <template #icon><PlusOutlined /></template>
        添加黑名单
      </a-button>
    </div>

    <a-card class="table-card">
      <div class="toolbar">
        <a-input-search
          v-model:value="searchKeyword"
          placeholder="搜索 IP 地址"
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
              编辑备注
            </a-button>
            <a-popconfirm
              title="确定要删除该黑名单吗？"
              ok-text="确定"
              cancel-text="取消"
              @confirm="handleDelete(record.id)"
            >
              <a-button type="link" size="small" danger>删除</a-button>
            </a-popconfirm>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="modalOpen"
      :title="isEdit ? '编辑黑名单备注' : '添加黑名单'"
      :confirm-loading="submitting"
      ok-text="确认"
      cancel-text="取消"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item label="IP 地址">
          <a-input
            v-model:value="form.ip"
            placeholder="请输入 IP 地址"
            :disabled="isEdit"
          />
        </a-form-item>
        <a-form-item label="备注 / 原因">
          <a-textarea
            v-model:value="form.reason"
            placeholder="请输入备注或封禁原因"
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
import {
  getBlacklistsAPI,
  createBlacklistAPI,
  updateBlacklistAPI,
  deleteBlacklistAPI,
} from '@/api/security'
import type { BlacklistItem } from '@/types/security'

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
  showTotal: (t: number) => `共 ${t} 条`,
  showSizeChanger: true,
}))

const columns: TableColumnsType<BlacklistItem> = [
  { title: 'IP 地址', dataIndex: 'ip', key: 'ip', width: 180 },
  { title: '备注 / 原因', dataIndex: 'reason', key: 'reason', ellipsis: true },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 200 },
  { title: '操作', key: 'action', width: 160, fixed: 'right' },
]

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
    message.warning('请输入 IP 地址')
    return
  }

  submitting.value = true
  try {
    if (isEdit.value) {
      await updateBlacklistAPI(editId.value, { reason })
      message.success('备注更新成功')
    } else {
      await createBlacklistAPI({ ip, reason })
      message.success('黑名单添加成功')
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
    message.success('删除成功')
    await fetchBlacklist()
  } catch {
    // 错误由请求拦截器处理
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
