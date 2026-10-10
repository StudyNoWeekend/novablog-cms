<template>
  <div class="page-container">
    <div class="page-header">
      <div>
        <h1 class="page-title">MCP 服务</h1>
        <p class="page-desc">通过 MCP 协议让 AI 客户端直接发布和管理博客内容</p>
      </div>
      <a-button type="primary" @click="openCreateModal">
        <template #icon><PlusOutlined /></template>
        新建密钥
      </a-button>
    </div>

    <!-- 接入信息 -->
    <a-card class="section-card" title="接入信息">
      <template #extra>
        <a-tag color="green">Streamable HTTP</a-tag>
      </template>
      <div class="endpoint-row">
        <span class="endpoint-label">服务端点</span>
        <code class="endpoint-value">{{ endpoint }}</code>
        <a-button type="link" size="small" @click="handleCopyEndpoint">
          <template #icon><CopyOutlined /></template>
          复制
        </a-button>
      </div>
      <ul class="info-list">
        <li>鉴权方式：请求头携带 <code>Authorization: Bearer &lt;API Key&gt;</code>，密钥在下方创建。</li>
        <li>可用能力：创建 / 更新 / 发布 / 下架 / 删除文章（Markdown），上传图片到媒体库，查询分类与标签，共 9 个工具。</li>
        <li>安全限制：每个密钥限流 120 次/分钟；删除文章需要 AI 显式确认；密钥明文仅在创建时显示一次。</li>
      </ul>
    </a-card>

    <!-- API 密钥 -->
    <a-card class="section-card" title="API 密钥">
      <div class="toolbar">
        <a-input-search
          v-model:value="searchKeyword"
          placeholder="搜索密钥名称"
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
              checked-children="启用"
              un-checked-children="禁用"
              @change="(checked: boolean | string | number) => handleToggleStatus(record, checked)"
            />
          </template>
          <template v-else-if="column.key === 'call_count'">
            {{ record.call_count }} 次
          </template>
          <template v-else-if="column.key === 'last_used_at'">
            {{ record.last_used_at ? formatDateTime(record.last_used_at) : '从未使用' }}
          </template>
          <template v-else-if="column.key === 'expires_at'">
            {{ record.expires_at ? formatDateTime(record.expires_at) : '永久有效' }}
          </template>
          <template v-else-if="column.key === 'created_at'">
            {{ formatDateTime(record.created_at) }}
          </template>
          <template v-else-if="column.key === 'action'">
            <a-popconfirm
              title="删除后使用该密钥的 AI 客户端将立即失去访问权限，确定删除？"
              ok-text="删除"
              cancel-text="取消"
              @confirm="handleDelete(record.id)"
            >
              <a-button type="link" size="small" danger>删除</a-button>
            </a-popconfirm>
          </template>
        </template>

        <template #emptyText>
          <div class="empty-guide">
            <p class="empty-title">还没有 API 密钥，三步接入：</p>
            <div class="empty-steps">
              <div class="empty-step"><span class="step-no">1</span>点击「新建密钥」生成密钥</div>
              <div class="empty-step"><span class="step-no">2</span>复制 MCP 配置到 AI 客户端</div>
              <div class="empty-step"><span class="step-no">3</span>在 AI 客户端里让它发布文章</div>
            </div>
            <a-button type="primary" @click="openCreateModal">立即创建</a-button>
          </div>
        </template>
      </a-table>
    </a-card>

    <!-- 客户端接入配置 -->
    <a-card class="section-card" title="客户端接入配置">
      <template #extra>
        <span class="placeholder-tip">配置中的密钥为占位符，创建密钥后可获得预填完整配置</span>
      </template>
      <McpConfigJson :endpoint="endpoint" :key-value="''" />
    </a-card>

    <!-- 新建密钥弹窗 -->
    <a-modal
      v-model:open="createModalOpen"
      title="新建 MCP 密钥"
      :confirm-loading="creating"
      ok-text="创建"
      cancel-text="取消"
      @ok="handleCreate"
    >
      <a-form layout="vertical">
        <a-form-item label="密钥名称" required>
          <a-input
            v-model:value="createForm.name"
            placeholder="例如：我的工作电脑"
            :maxlength="100"
            show-count
          />
        </a-form-item>
        <a-form-item label="有效期">
          <a-radio-group v-model:value="createForm.expireDays">
            <a-radio-button :value="0">永久</a-radio-button>
            <a-radio-button :value="30">30 天</a-radio-button>
            <a-radio-button :value="90">90 天</a-radio-button>
            <a-radio-button :value="180">180 天</a-radio-button>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 创建成功弹窗（密钥仅显示一次） -->
    <a-modal
      v-model:open="createdModalOpen"
      title="密钥创建成功"
      :footer="null"
      :width="680"
      class="created-modal"
    >
      <a-alert
        type="warning"
        show-icon
        message="请立即保存密钥，关闭后将无法再次查看"
        description="密钥仅展示这一次。建议直接复制下方客户端配置使用。"
        class="key-alert"
      />
      <div class="created-key-row">
        <code class="created-key">{{ createdKeyInfo?.key }}</code>
        <a-button type="primary" size="small" @click="handleCopyKey">
          <template #icon><CopyOutlined /></template>
          复制密钥
        </a-button>
      </div>
      <McpConfigJson v-if="createdKeyInfo" :endpoint="endpoint" :key-value="createdKeyInfo.key" />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
import McpConfigJson from '@/components/mcp/McpConfigJson.vue'

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
  showTotal: (t: number) => `共 ${t} 条`,
  showSizeChanger: true,
}))

const columns: TableColumnsType<McpKeyItem> = [
  { title: '名称', dataIndex: 'name', key: 'name', width: 160, ellipsis: true },
  { title: '密钥', dataIndex: 'key_prefix', key: 'key_prefix', width: 190 },
  { title: '状态', key: 'status', width: 90 },
  { title: '调用次数', dataIndex: 'call_count', key: 'call_count', width: 90 },
  { title: '最近使用', key: 'last_used_at', width: 150 },
  { title: '有效期至', key: 'expires_at', width: 150 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 150 },
  { title: '操作', key: 'action', width: 80, fixed: 'right' },
]

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
    message.success(status === MCP_KEY_STATUS.ENABLED ? '已启用' : '已禁用，使用该密钥的客户端将无法调用')
  } catch {
    // 错误由请求拦截器处理
  }
}

async function handleDelete(id: string) {
  try {
    await deleteMcpKeyAPI(id)
    message.success('密钥已删除')
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
    message.warning('请输入密钥名称')
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
    message.success('端点地址已复制')
  } else {
    message.error('复制失败，请手动复制')
  }
}

async function handleCopyKey() {
  if (!createdKeyInfo.value) return
  const ok = await copyText(createdKeyInfo.value.key)
  if (ok) {
    message.success('密钥已复制')
  } else {
    message.error('复制失败，请手动选择复制')
  }
}

// ---------- 工具 ----------
function formatDateTime(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
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
