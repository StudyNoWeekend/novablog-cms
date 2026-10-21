<template>
  <div class="community-push-view">
    <!-- 连接配置 -->
    <a-card :title="$t('system.push.configTitle')" class="config-card">
      <template #extra>
        <a-tag v-if="config?.has_push_key" color="success">{{ $t('system.push.keyConfigured') }}</a-tag>
        <a-tag v-else color="warning">{{ $t('system.push.keyMissing') }}</a-tag>
      </template>

      <a-form layout="vertical" class="config-form">
        <a-form-item :label="$t('system.push.baseUrl')" required :extra="$t('system.push.baseUrlExtra')">
          <a-input
            v-model:value="form.base_url"
            :placeholder="$t('system.push.baseUrlPlaceholder')"
            allow-clear
          />
        </a-form-item>
        <a-form-item :label="$t('system.push.siteUrl')" :extra="$t('system.push.siteUrlExtra')">
          <a-input v-model:value="form.site_url" placeholder="https://blog.example.com" allow-clear />
        </a-form-item>
        <a-form-item :extra="pushKeyExtra">
          <template #label>
            {{ $t('system.push.pushKey') }}
            <a-tag v-if="config?.has_push_key" class="ml-2" color="default">{{ config.push_key_prefix }}…</a-tag>
          </template>
          <a-input-password
            v-model:value="form.push_key"
            :placeholder="config?.has_push_key ? $t('system.push.pushKeyKeep') : $t('system.push.pushKeyPlaceholder')"
            autocomplete="new-password"
          />
        </a-form-item>
        <a-form-item
          :label="$t('system.push.verifyCode')"
          :extra="$t('system.push.verifyCodeExtra')"
        >
          <a-input v-model:value="form.verify_code" :placeholder="$t('system.push.verifyCodePlaceholder')" allow-clear />
        </a-form-item>
        <div class="config-actions">
          <a-button type="primary" :loading="saving" @click="handleSave">
            <SaveOutlined /> {{ $t('system.push.saveConfig') }}
          </a-button>
          <a-button :loading="testing" @click="handleTest">
            <ApiOutlined /> {{ $t('system.push.testConnection') }}
          </a-button>
        </div>
        <a-alert
          v-if="testResult"
          type="success"
          show-icon
          class="test-result"
          :message="$t('system.push.testSuccessTitle', { name: testResult.credential_name, username: testResult.username })"
          :description="$t('system.push.testSuccessDesc', { url: testResult.blog_url || $t('system.push.notBound') })"
        />
        <a-alert
          v-if="verifyResult"
          :type="verifyResult.status === 'verified' ? 'success' : 'error'"
          show-icon
          class="test-result"
          :message="verifyResult.status === 'verified' ? $t('system.push.verifyPassed') : $t('system.push.verifyFailed')"
          :description="verifyResult.message"
        />
      </a-form>
    </a-card>

    <!-- 同步记录 -->
    <a-card :title="$t('system.push.recordsTitle')" class="records-card">
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
              {{ $t('system.push.repush') }}
            </a-button>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import { ApiOutlined, SaveOutlined } from '@ant-design/icons-vue'

import {
  communityPushApi,
  type CommunityPushConfig,
  type CommunitySyncRecord,
} from '@/api/communityPush'

const { t } = useI18n()

const config = ref<CommunityPushConfig | null>(null)
const form = reactive({ base_url: '', site_url: '', push_key: '', verify_code: '' })
const saving = ref(false)
const testing = ref(false)
const testResult = ref<{ username: string; blog_url: string; credential_name: string } | null>(null)
const verifyResult = ref<{ status: 'verified' | 'failed'; message: string } | null>(null)

const records = ref<CommunitySyncRecord[]>([])
const recordsTotal = ref(0)
const recordsPage = ref(1)
const recordsLoading = ref(false)
const statusFilter = ref('')
const retryingId = ref('')

const pushKeyExtra = computed(() => t('system.push.pushKeyExtra'))

// statusOptions 同步状态下拉（含筛选全部项）
const statusOptions = computed(() => [
  { value: '', label: t('common.status.all') },
  { value: 'synced', label: t('system.sync.synced') },
  { value: 'hidden', label: t('system.sync.hidden') },
  { value: 'failed', label: t('system.sync.failed') },
  { value: 'deleted', label: t('system.sync.deleted') },
])

const recordColumns = computed(() => [
  { title: t('common.field.content'), key: 'content_title', ellipsis: true },
  { title: t('common.field.type'), key: 'target_type', width: 90 },
  { title: t('system.record.communityStatus'), key: 'sync_status', width: 96 },
  { title: t('system.record.communityArticleId'), key: 'community_article_id', width: 110 },
  { title: t('system.record.lastPush'), key: 'pushed_at', width: 140 },
  { title: t('system.record.retryCount'), key: 'retry_count', width: 64 },
  { title: t('system.record.lastError'), key: 'last_error', ellipsis: true },
  { title: t('common.table.actions'), key: 'action', width: 100 },
])

const recordsPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: false,
})

// targetTypeLabels 推送内容类型展示文案，未知类型回退原始值
const targetTypeLabels = computed<Record<string, string>>(() => ({
  article: t('system.target.article'),
  travel: t('system.target.travel'),
  recipe: t('system.target.recipe'),
  portfolio: t('system.target.portfolio'),
  open_source: t('system.target.openSource'),
}))

// statusLabels 同步状态展示文案，未知状态回退原始值
const statusLabels = computed<Record<string, string>>(() => ({
  synced: t('system.sync.synced'),
  hidden: t('system.sync.hidden'),
  failed: t('system.sync.failed'),
  deleted: t('system.sync.deleted'),
  none: t('system.sync.notPushed'),
}))

const statusColors: Record<string, string> = {
  synced: 'success',
  hidden: 'warning',
  failed: 'error',
  deleted: 'default',
  none: 'default',
}

function targetTypeLabel(type: string) {
  return targetTypeLabels.value[type] ?? type
}

function statusLabel(status: string) {
  return statusLabels.value[status] ?? status
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
    form.verify_code = config.value.verify_code || ''
  } catch {
    // 错误由拦截器统一提示
  }
}

async function handleSave() {
  if (!form.base_url.trim()) {
    message.warning(t('system.validation.baseUrlRequired'))
    return
  }
  saving.value = true
  try {
    const result = await communityPushApi.updateConfig({
      base_url: form.base_url.trim(),
      site_url: form.site_url.trim(),
      push_key: form.push_key.trim() || undefined,
      verify_code: form.verify_code.trim(),
    })
    config.value = result
    form.push_key = ''
    // 验证码随保存一并提交给社区，结果单独提示
    if (result.verify_status === 'verified') {
      verifyResult.value = { status: 'verified', message: result.verify_message || '' }
      message.success(t('system.message.saveVerified'))
    } else if (result.verify_status === 'failed') {
      verifyResult.value = { status: 'failed', message: result.verify_message || t('system.message.verifyRejected') }
      message.warning(t('system.message.saveVerifyFailed'))
    } else {
      verifyResult.value = null
      message.success(t('system.message.saveSuccess'))
    }
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
    message.success(t('system.message.repushSuccess'))
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
