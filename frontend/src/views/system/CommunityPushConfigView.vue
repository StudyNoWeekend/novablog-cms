<template>
  <div class="community-push-view">
    <!-- 连接配置 -->
    <a-card title="社区推送配置" class="config-card">
      <template #extra>
        <a-tag v-if="config?.has_push_key" color="success">已配置密钥</a-tag>
        <a-tag v-else color="warning">未配置密钥</a-tag>
      </template>

      <a-form layout="vertical" class="config-form">
        <a-form-item label="NovaBlog 社区地址" required extra="社区推送与主题市场的统一官方入口，主题模块同步使用此地址">
          <a-input
            v-model:value="form.base_url"
            placeholder="https://community.example.com（自动补 /api/v1）"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="博客对外地址" extra="用于拼接推送内容的「阅读原文」链接">
          <a-input v-model:value="form.site_url" placeholder="https://blog.example.com" allow-clear />
        </a-form-item>
        <a-form-item :extra="pushKeyExtra">
          <template #label>
            推送密钥
            <a-tag v-if="config?.has_push_key" class="ml-2" color="default">{{ config.push_key_prefix }}…</a-tag>
          </template>
          <a-input-password
            v-model:value="form.push_key"
            :placeholder="config?.has_push_key ? '留空表示不修改已有密钥' : '在 NovaBlog 个人中心生成，nbk_ 开头'"
            autocomplete="new-password"
          />
        </a-form-item>
        <div class="config-actions">
          <a-button type="primary" :loading="saving" @click="handleSave">
            <SaveOutlined /> 保存配置
          </a-button>
          <a-button :loading="testing" @click="handleTest">
            <ApiOutlined /> 测试连接
          </a-button>
        </div>
        <a-alert
          v-if="testResult"
          type="success"
          show-icon
          class="test-result"
          :message="`连接成功：${testResult.credential_name}（博主 ${testResult.username}）`"
          :description="`社区账号博客地址：${testResult.blog_url || '未绑定'}`"
        />
      </a-form>
    </a-card>

    <!-- 同步记录 -->
    <a-card title="同步记录" class="records-card">
      <template #extra>
        <a-select
          v-model:value="statusFilter"
          style="width: 140px"
          :options="statusOptions"
          @change="loadRecords(1)"
        />
      </template>

      <a-table
        :columns="recordColumns"
        :data-source="records"
        :loading="recordsLoading"
        :pagination="recordsPagination"
        row-key="id"
        size="small"
        @change="handleRecordsChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'content_title'">
            <span class="record-title">{{ record.content_title || record.target_id }}</span>
          </template>
          <template v-else-if="column.key === 'target_type'">
            {{ targetTypeLabel(record.target_type) }}
          </template>
          <template v-else-if="column.key === 'sync_status'">
            <a-tag :color="statusColor(record.sync_status)">{{ statusLabel(record.sync_status) }}</a-tag>
          </template>
          <template v-else-if="column.key === 'last_error'">
            <a-tooltip :title="record.last_error">
              <span class="record-error">{{ record.last_error || '—' }}</span>
            </a-tooltip>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-button
              type="link"
              size="small"
              :loading="retryingId === record.target_id"
              :disabled="record.sync_status === 'deleted'"
              @click="handleRetry(record)"
            >
              重新推送
            </a-button>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import { ApiOutlined, SaveOutlined } from '@ant-design/icons-vue'

import {
  communityPushApi,
  type CommunityPushConfig,
  type CommunitySyncRecord,
} from '@/api/communityPush'

const config = ref<CommunityPushConfig | null>(null)
const form = reactive({ base_url: '', site_url: '', push_key: '' })
const saving = ref(false)
const testing = ref(false)
const testResult = ref<{ username: string; blog_url: string; credential_name: string } | null>(null)

const records = ref<CommunitySyncRecord[]>([])
const recordsTotal = ref(0)
const recordsPage = ref(1)
const recordsLoading = ref(false)
const statusFilter = ref('')
const retryingId = ref('')

const pushKeyExtra = '在 NovaBlog 个人中心「推送密钥」生成；保存后仅显示前缀，不会回显完整密钥'

const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'synced', label: '已同步' },
  { value: 'hidden', label: '已下架' },
  { value: 'failed', label: '失败' },
  { value: 'deleted', label: '已删除' },
]

const recordColumns = [
  { title: '内容', key: 'content_title', ellipsis: true },
  { title: '类型', key: 'target_type', width: 90 },
  { title: '社区状态', key: 'sync_status', width: 96 },
  { title: '社区文章 ID', key: 'community_article_id', width: 110 },
  { title: '最后推送', key: 'pushed_at', width: 140 },
  { title: '重试', key: 'retry_count', width: 64 },
  { title: '错误信息', key: 'last_error', ellipsis: true },
  { title: '操作', key: 'action', width: 100 },
]

const recordsPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: false,
})

const targetTypeLabels: Record<string, string> = {
  article: '文章',
  travel: '旅行攻略',
  recipe: '美食食谱',
  portfolio: '作品集',
  open_source: '开源作品',
}

const statusLabels: Record<string, string> = {
  synced: '已同步',
  hidden: '已下架',
  failed: '失败',
  deleted: '已删除',
  none: '未推送',
}

const statusColors: Record<string, string> = {
  synced: 'success',
  hidden: 'warning',
  failed: 'error',
  deleted: 'default',
  none: 'default',
}

function targetTypeLabel(type: string) {
  return targetTypeLabels[type] ?? type
}

function statusLabel(status: string) {
  return statusLabels[status] ?? status
}

function statusColor(status: string) {
  return statusColors[status] ?? 'default'
}

async function loadConfig() {
  try {
    config.value = await communityPushApi.getConfig()
    form.base_url = config.value.base_url
    form.site_url = config.value.site_url
    form.push_key = ''
  } catch {
    // 错误由拦截器统一提示
  }
}

async function handleSave() {
  if (!form.base_url.trim()) {
    message.warning('请填写社区地址')
    return
  }
  saving.value = true
  try {
    config.value = await communityPushApi.updateConfig({
      base_url: form.base_url.trim(),
      site_url: form.site_url.trim(),
      push_key: form.push_key.trim() || undefined,
    })
    form.push_key = ''
    message.success('推送配置已保存')
  } finally {
    saving.value = false
  }
}

async function handleTest() {
  testing.value = true
  testResult.value = null
  try {
    testResult.value = await communityPushApi.testConnection()
  } finally {
    testing.value = false
  }
}

async function loadRecords(page = 1) {
  recordsLoading.value = true
  try {
    const res = await communityPushApi.getRecords({
      page,
      size: recordsPagination.pageSize,
      sync_status: statusFilter.value || undefined,
    })
    records.value = res.items
    recordsTotal.value = res.total
    recordsPage.value = page
    recordsPagination.current = page
    recordsPagination.total = res.total
  } finally {
    recordsLoading.value = false
  }
}

function handleRecordsChange(pagination: { current?: number }) {
  loadRecords(pagination.current ?? 1)
}

async function handleRetry(record: CommunitySyncRecord) {
  retryingId.value = record.target_id
  try {
    await communityPushApi.sync(record.target_type, record.target_id, 'push')
    message.success('已重新推送')
    await loadRecords(recordsPage.value)
  } finally {
    retryingId.value = ''
  }
}

onMounted(() => {
  loadConfig()
  loadRecords(1)
})
</script>

<style scoped>
.community-push-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.config-actions {
  display: flex;
  gap: 12px;
}

.test-result {
  margin-top: 12px;
}

.record-title {
  font-weight: 500;
}

.record-error {
  color: var(--color-text-tertiary, #8a93a8);
}
</style>
