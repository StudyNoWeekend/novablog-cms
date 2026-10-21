<template>
  <div class="page-container">
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ $t('mcp.title') }}</h1>
        <p class="page-desc">{{ $t('mcp.desc') }}</p>
      </div>
      <a-button type="primary" @click="openCreateModal">
        <template #icon><PlusOutlined /></template>
        {{ $t('mcp.action.createKey') }}
      </a-button>
    </div>

    <!-- 接入信息 -->
    <a-card class="section-card" :title="$t('mcp.access.title')">
      <template #extra>
        <a-tag color="green">Streamable HTTP</a-tag>
      </template>
      <div class="endpoint-row">
        <span class="endpoint-label">{{ $t('mcp.access.endpoint') }}</span>
        <code class="endpoint-value">{{ endpoint }}</code>
        <a-button type="link" size="small" @click="handleCopyEndpoint">
          <template #icon><CopyOutlined /></template>
          {{ $t('common.actions.copy') }}
        </a-button>
      </div>
      <ul class="info-list">
        <li>{{ $t('mcp.access.authPrefix') }} <code>Authorization: Bearer &lt;API Key&gt;</code>{{ $t('mcp.access.authSuffix') }}</li>
        <li>{{ $t('mcp.access.capabilities') }}</li>
        <li>{{ $t('mcp.access.security') }}</li>
      </ul>
    </a-card>

    <!-- API 密钥 -->
    <a-card class="section-card" :title="$t('mcp.keys.title')">
      <div class="toolbar">
        <a-input-search
          v-model:value="searchKeyword"
          :placeholder="$t('mcp.keys.searchPlaceholder')"
          allow-clear
          enter-button
          style="max-width: 280px"
          @search="handleSearch"
        />
      </div>

      <a-table
        :columns="columns"
        :data-source="keys"
        :loading="loading"
        :pagination="pagination"
        row-key="id"
        :scroll="{ x: 860 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'key_prefix'">
            <code class="key-prefix">{{ record.key_prefix }}</code>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-switch
              :checked="record.status === MCP_KEY_STATUS.ENABLED"
              :checked-children="$t('common.actions.enable')"
              :un-checked-children="$t('common.actions.disable')"
              @change="(checked: boolean | string | number) => handleToggleStatus(record, checked)"
            />
          </template>
          <template v-else-if="column.key === 'call_count'">
            {{ $t('mcp.keys.callCount', { count: record.call_count }) }}
          </template>
          <template v-else-if="column.key === 'last_used_at'">
            {{ record.last_used_at ? formatDateTime(record.last_used_at) : $t('mcp.keys.neverUsed') }}
          </template>
          <template v-else-if="column.key === 'expires_at'">
            {{ record.expires_at ? formatDateTime(record.expires_at) : $t('mcp.keys.neverExpires') }}
          </template>
          <template v-else-if="column.key === 'created_at'">
            {{ formatDateTime(record.created_at) }}
          </template>
          <template v-else-if="column.key === 'action'">
            <a-popconfirm
              :title="$t('mcp.message.deleteConfirm')"
              :ok-text="$t('common.actions.delete')"
              :cancel-text="$t('common.actions.cancel')"
              @confirm="handleDelete(record.id)"
            >
              <a-button type="link" size="small" danger>{{ $t('common.actions.delete') }}</a-button>
            </a-popconfirm>
          </template>
        </template>

        <template #emptyText>
          <div class="empty-guide">
            <p class="empty-title">{{ $t('mcp.empty.title') }}</p>
            <div class="empty-steps">
              <div class="empty-step"><span class="step-no">1</span>{{ $t('mcp.empty.step1') }}</div>
              <div class="empty-step"><span class="step-no">2</span>{{ $t('mcp.empty.step2') }}</div>
              <div class="empty-step"><span class="step-no">3</span>{{ $t('mcp.empty.step3') }}</div>
            </div>
            <a-button type="primary" @click="openCreateModal">{{ $t('mcp.empty.create') }}</a-button>
          </div>
        </template>
      </a-table>
    </a-card>

    <!-- 客户端接入配置 -->
    <a-card class="section-card" :title="$t('mcp.clientConfig.title')">
      <template #extra>
        <span class="placeholder-tip">{{ $t('mcp.clientConfig.placeholderTip') }}</span>
      </template>
      <McpConfigJson :endpoint="endpoint" :key-value="''" />
    </a-card>

    <!-- 新建密钥弹窗 -->
    <a-modal
      v-model:open="createModalOpen"
      :title="$t('mcp.create.title')"
      :confirm-loading="creating"
      :ok-text="$t('mcp.create.okText')"
      :cancel-text="$t('common.actions.cancel')"
      @ok="handleCreate"
    >
      <a-form layout="vertical">
        <a-form-item :label="$t('mcp.create.nameLabel')" required>
          <a-input
            v-model:value="createForm.name"
            :placeholder="$t('mcp.create.namePlaceholder')"
            :maxlength="100"
            show-count
          />
        </a-form-item>
        <a-form-item :label="$t('mcp.create.expireLabel')">
          <a-radio-group v-model:value="createForm.expireDays">
            <a-radio-button :value="0">{{ $t('mcp.create.permanent') }}</a-radio-button>
            <a-radio-button :value="30">{{ $t('mcp.create.days', { days: 30 }) }}</a-radio-button>
            <a-radio-button :value="90">{{ $t('mcp.create.days', { days: 90 }) }}</a-radio-button>
            <a-radio-button :value="180">{{ $t('mcp.create.days', { days: 180 }) }}</a-radio-button>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 创建成功弹窗（密钥仅显示一次） -->
    <a-modal
      v-model:open="createdModalOpen"
      :title="$t('mcp.created.title')"
      :footer="null"
      :width="680"
      class="created-modal"
    >
      <a-alert
        type="warning"
        show-icon
        :message="$t('mcp.created.alertTitle')"
        :description="$t('mcp.created.alertDesc')"
        class="key-alert"
      />
      <div class="created-key-row">
        <code class="created-key">{{ createdKeyInfo?.key }}</code>
        <a-button type="primary" size="small" @click="handleCopyKey">
          <template #icon><CopyOutlined /></template>
          {{ $t('mcp.action.copyKey') }}
        </a-button>
      </div>
      <McpConfigJson v-if="createdKeyInfo" :endpoint="endpoint" :key-value="createdKeyInfo.key" />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { message, type TableColumnsType } from 'ant-design-vue'
import { PlusOutlined, CopyOutlined } from '@ant-design/icons-vue'
import {
  getMcpKeysAPI,
  createMcpKeyAPI,
  updateMcpKeyAPI,
  deleteMcpKeyAPI,
} from '@/api/mcp'
import { MCP_KEY_STATUS, type McpKeyItem } from '@/types/mcp'
import { getMcpEndpoint } from '@/constants/mcpSnippets'
import { copyText } from '@/utils/clipboard'
import { formatDateTime } from '@/utils/datetime'
import McpConfigJson from '@/components/mcp/McpConfigJson.vue'

const { t } = useI18n()

const endpoint = getMcpEndpoint()

// ---------- 列表 ----------
const loading = ref(false)
const keys = ref<McpKeyItem[]>([])
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

const columns = computed<TableColumnsType<McpKeyItem>>(() => [
  { title: t('common.field.name'), dataIndex: 'name', key: 'name', width: 160, ellipsis: true },
  { title: t('mcp.keys.columns.key'), dataIndex: 'key_prefix', key: 'key_prefix', width: 190 },
  { title: t('common.field.status'), key: 'status', width: 90 },
  { title: t('mcp.keys.columns.callCount'), dataIndex: 'call_count', key: 'call_count', width: 90 },
  { title: t('mcp.keys.columns.lastUsed'), key: 'last_used_at', width: 150 },
  { title: t('mcp.keys.columns.expiresAt'), key: 'expires_at', width: 150 },
  { title: t('common.field.createTime'), dataIndex: 'created_at', key: 'created_at', width: 150 },
  { title: t('common.table.actions'), key: 'action', width: 80, fixed: 'right' },
])

async function fetchKeys() {
  loading.value = true
  try {
    const data = await getMcpKeysAPI({
      page: page.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value || undefined,
    })
    keys.value = data.list
    total.value = data.total
  } catch {
    // 错误由请求拦截器处理
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  fetchKeys()
}

function handleTableChange(paginationState: { current?: number; pageSize?: number }) {
  page.value = paginationState.current ?? 1
  pageSize.value = paginationState.pageSize ?? 10
  fetchKeys()
}

async function handleToggleStatus(record: McpKeyItem, checked: boolean | string | number) {
  const status = checked ? MCP_KEY_STATUS.ENABLED : MCP_KEY_STATUS.DISABLED
  try {
    await updateMcpKeyAPI(record.id, { status })
    record.status = status
    message.success(status === MCP_KEY_STATUS.ENABLED ? t('common.status.enabled') : t('mcp.message.disabledNotice'))
  } catch {
    // 错误由请求拦截器处理
  }
}

async function handleDelete(id: string) {
  try {
    await deleteMcpKeyAPI(id)
    message.success(t('mcp.message.deleteSuccess'))
    await fetchKeys()
  } catch {
    // 错误由请求拦截器处理
  }
}

// ---------- 创建密钥 ----------
const createModalOpen = ref(false)
const creating = ref(false)
const createForm = ref({ name: '', expireDays: 0 })
const createdModalOpen = ref(false)
const createdKeyInfo = ref<{ key: string; name: string } | null>(null)

function openCreateModal() {
  createForm.value = { name: '', expireDays: 0 }
  createModalOpen.value = true
}

async function handleCreate() {
  const name = createForm.value.name.trim()
  if (!name) {
    message.warning(t('mcp.message.nameRequired'))
    return
  }
  creating.value = true
  try {
    const days = createForm.value.expireDays
    const expiresAt =
      days > 0 ? new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString() : null
    const res = await createMcpKeyAPI({ name, expires_at: expiresAt })
    createModalOpen.value = false
    createdKeyInfo.value = { key: res.key, name: res.name }
    createdModalOpen.value = true
    await fetchKeys()
  } catch {
    // 错误由请求拦截器处理
  } finally {
    creating.value = false
  }
}

// ---------- 复制 ----------
async function handleCopyEndpoint() {
  const ok = await copyText(endpoint)
  if (ok) {
    message.success(t('mcp.message.endpointCopied'))
  } else {
    message.error(t('mcp.message.copyFailedManual'))
  }
}

async function handleCopyKey() {
  if (!createdKeyInfo.value) return
  const ok = await copyText(createdKeyInfo.value.key)
  if (ok) {
    message.success(t('mcp.message.keyCopied'))
  } else {
    message.error(t('mcp.message.copyFailedSelect'))
  }
}

onMounted(() => {
  fetchKeys()
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
  color: var(--text-primary, #29365c);
  margin: 0;
}

.page-desc {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-secondary, #667085);
}

.section-card {
  border-radius: var(--border-radius-lg, 12px);
  margin-bottom: 20px;
}

.endpoint-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.endpoint-label {
  font-size: 13px;
  color: var(--text-secondary, #667085);
  flex-shrink: 0;
}

.endpoint-value {
  padding: 6px 12px;
  background: #f6f8fb;
  border: 1px solid #e8ecf3;
  border-radius: 6px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  color: var(--text-primary, #29365c);
  word-break: break-all;
}

.info-list {
  margin: 0;
  padding-left: 18px;
  font-size: 13px;
  line-height: 2;
  color: var(--text-secondary, #667085);
}

.info-list code {
  padding: 1px 6px;
  background: #f6f8fb;
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-primary, #29365c);
}

.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.key-prefix {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12.5px;
  color: var(--text-primary, #29365c);
}

.placeholder-tip {
  font-size: 12px;
  color: var(--text-secondary, #667085);
}

.key-alert {
  margin-bottom: 16px;
}

.created-key-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.created-key {
  flex: 1;
  padding: 10px 14px;
  background: #f6f8fb;
  border: 1px dashed #b9c4dd;
  border-radius: 8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  color: var(--text-primary, #29365c);
  word-break: break-all;
}

.empty-guide {
  padding: 24px 0 32px;
}

.empty-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #29365c);
  margin-bottom: 16px;
}

.empty-steps {
  display: flex;
  justify-content: center;
  gap: 24px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.empty-step {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary, #667085);
}

.step-no {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--color-primary, #526fe8);
  color: #fff;
  font-size: 12px;
  flex-shrink: 0;
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

  .created-key-row {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
