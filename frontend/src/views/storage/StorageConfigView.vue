<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">对象存储</h1>
    </div>

    <a-card>
      <a-tabs v-model:activeKey="activeTab">
        <a-tab-pane key="config" tab="存储配置" />
        <a-tab-pane key="migration" tab="素材迁移" />
      </a-tabs>

      <!-- ============ Tab 1: 存储配置 ============ -->
      <template v-if="activeTab === 'config'">
        <div class="toolbar">
          <a-button v-if="!hasPendingConfig" type="primary" @click="handleAdd">切换存储平台</a-button>
        </div>

        <a-spin :spinning="configLoading">
          <a-table
            :columns="configColumns"
            :data-source="configs"
            :pagination="false"
            row-key="provider"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'provider'">
                {{ providerLabel(record.provider) }}
              </template>
              <template v-if="column.key === 'status'">
                <a-tag v-if="record.is_active" color="green">使用中</a-tag>
                <a-tag v-else color="orange">待迁移</a-tag>
              </template>
              <template v-if="column.key === 'updated_at'">
                {{ formatDateTime(record.updated_at) }}
              </template>
              <template v-if="column.key === 'action'">
                <a-button
                  v-if="!record.is_active"
                  type="link"
                  @click="handleActivate(record)"
                >
                  激活
                </a-button>
                <a-button
                  v-if="!record.is_active"
                  type="link"
                  @click="activeTab = 'migration'"
                >
                  前往迁移
                </a-button>
                <a-button
                  v-if="record.is_active"
                  type="link"
                  @click="handleEdit(record)"
                >
                  编辑
                </a-button>
                <a-button
                  v-if="!record.is_active"
                  type="link"
                  danger
                  @click="handleDelete(record)"
                >
                  取消
                </a-button>
                <span v-if="record.is_active" style="color: #999; font-size: 13px">-</span>
              </template>
            </template>
          </a-table>
        </a-spin>
      </template>

      <!-- ============ Tab 2: 素材迁移 ============ -->
      <template v-if="activeTab === 'migration'">
        <!-- 迁移目标平台信息 -->
        <a-alert
          v-if="targetConfig"
          type="info"
          show-icon
          style="margin-bottom: 16px"
        >
          <template #message>
            <template v-if="sourcePlatforms.length > 0">
              将素材从 {{ sourcePlatformLabels }} 迁移至「{{ providerLabel(targetConfig.provider) }}」
            </template>
            <template v-else>
              迁移目标：{{ providerLabel(targetConfig.provider) }}，点击「开始分析」检查需迁移的素材
            </template>
          </template>
        </a-alert>
        <a-alert
          v-else
          type="warning"
          message="请先在「存储配置」中添加存储平台配置"
          show-icon
          style="margin-bottom: 16px"
        />

        <!-- 分析阶段 -->
        <template v-if="targetConfig && migrationState === 'idle'">
          <div v-if="analyzeState === 'idle'">
            <a-button type="primary" @click="handleAnalyze">
              开始分析
            </a-button>
          </div>

          <div v-if="analyzeState === 'analyzing'" style="text-align: center; padding: 48px 0">
            <a-spin tip="正在分析素材..." />
          </div>

          <div v-if="analyzeState === 'completed' && analyzeResult">
            <a-statistic
              title="分析结果"
              :value="`共 ${(analyzeResult.missing || []).length + (analyzeResult.existing || []).length} 条素材，${(analyzeResult.existing || []).length} 条已存在，${(analyzeResult.missing || []).length} 条需要迁移`"
              style="margin-bottom: 16px"
            />
            <a-table
              :columns="existingColumns"
              :data-source="analyzeResult.existing"
              :pagination="false"
              row-key="id"
              style="margin-bottom: 16px; background: #f5f5f5"
              :row-class="() => 'existing-row'"
              bordered
            />
            <div class="migration-toolbar">
              <span style="font-weight: 500">需迁移素材：</span>
              <a-button
                type="primary"
                :disabled="selectedRowKeys.length === 0"
                @click="handleBatchMigrate"
              >
                批量迁移（{{ selectedRowKeys.length }}）
              </a-button>
              <a-button type="primary" @click="handleMigrateAll">
                一键迁移所有
              </a-button>
              <a-button @click="handleResetAnalyze">重新分析</a-button>
            </div>
            <a-table
              :columns="missingColumns"
              :data-source="analyzeResult.missing"
              :pagination="false"
              row-key="id"
              :row-selection="{
                selectedRowKeys,
                onChange: (keys: string[]) => { selectedRowKeys = keys },
              }"
              bordered
            />
          </div>
        </template>

        <!-- 迁移阶段 -->
        <template v-if="migrationState === 'migrating' || migrationState === 'completed' || migrationState === 'failed'">
          <a-progress
            :percent="migratePercent"
            :status="migrateProgressStatus"
          />
          <div style="margin: 12px 0; color: #666">
            已迁移 {{ migrateSucceeded }} / {{ migrateTotal }} 条
            <template v-if="migrationState === 'completed'">
              ，迁移完成，耗时 {{ elapsedTime }} 秒
            </template>
          </div>
          <a-button @click="handleResetAnalyze" style="margin-top: 8px">重新分析</a-button>

          <!-- 失败项列表 -->
          <a-collapse v-if="migrateFailedItems.length > 0" style="margin-top: 16px">
            <a-collapse-panel key="failed" :header="`失败项（${migrateFailedItems.length}）`">
              <a-table
                :columns="failedColumns"
                :data-source="migrateFailedItems"
                :pagination="false"
                row-key="id"
                bordered
              />
            </a-collapse-panel>
          </a-collapse>
        </template>
      </template>
    </a-card>

    <!-- ============ 新增/编辑 Modal ============ -->
    <a-modal
      v-model:open="modalVisible"
      :title="editingConfig ? '编辑存储配置' : '新增存储配置'"
      :width="560"
      :confirm-loading="modalLoading"
      @ok="handleSave"
      @cancel="closeModal"
    >
      <a-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="存储平台" name="provider">
          <a-select
            v-model:value="form.provider"
            :disabled="!!editingConfig"
            placeholder="请选择存储平台"
            @change="handleProviderChange"
          >
            <a-select-option v-for="p in availableProviders" :key="p.value" :value="p.value">
              {{ p.label }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item v-if="form.provider !== 'local'" :label="form.provider === 'tencent' ? 'Bucket URL' : 'Endpoint'" name="endpoint">
          <a-input
            v-model:value="form.endpoint"
            :placeholder="endpointPlaceholder"
          />
          <template #help>
            <span class="field-hint">示例：{{ endpointExample }}</span>
          </template>
        </a-form-item>

        <a-form-item
          v-if="form.provider === 'aliyun' || form.provider === 'tencent'"
          label="Region"
          name="region"
        >
          <a-input
            v-model:value="form.region"
            :placeholder="form.provider === 'tencent' ? '如 ap-guangzhou' : '如 cn-hangzhou'"
          />
          <template #help>
            <span v-if="form.provider === 'aliyun'" class="field-hint">必填，如 cn-hangzhou</span>
            <span v-else class="field-hint">如 ap-guangzhou</span>
          </template>
        </a-form-item>

        <a-form-item v-if="form.provider !== 'tencent' && form.provider !== 'local'" label="Bucket" name="bucket">
          <a-input
            v-model:value="form.bucket"
            :placeholder="bucketPlaceholder"
          />
          <template #help>
            <span class="field-hint">示例：{{ bucketExample }}</span>
          </template>
        </a-form-item>

        <a-form-item v-if="form.provider !== 'local'" :label="form.provider === 'tencent' ? 'Secret ID' : 'Access Key'" name="access_key">
          <a-input v-model:value="form.access_key" :placeholder="form.provider === 'tencent' ? '请输入 SecretId' : '请输入 Access Key'" />
        </a-form-item>

        <a-form-item v-if="form.provider !== 'local'" :label="form.provider === 'tencent' ? 'Secret Key' : 'Access Secret'" name="access_secret">
          <a-input-password
            v-model:value="form.access_secret"
            :placeholder="editingConfig ? '留空不修改' : (form.provider === 'tencent' ? '请输入 SecretKey' : '请输入 Access Secret')"
          />
        </a-form-item>

        <a-form-item label="路径前缀" name="path_prefix">
          <a-input v-model:value="form.path_prefix" placeholder="可选，如 media/" />
        </a-form-item>

        <a-form-item label="自定义域名" name="custom_domain">
          <a-input v-model:value="form.custom_domain" placeholder="可选，如 https://cdn.example.com" />
        </a-form-item>

        <a-form-item v-if="form.provider === 'minio'" label="Use SSL" name="use_ssl">
          <a-switch v-model:checked="form.use_ssl" />
        </a-form-item>

        <!-- 测试连通性结果 -->
        <a-form-item v-if="testResult" :wrapper-col="{ offset: 6, span: 18 }">
          <a-alert
            :type="testResult.success ? 'success' : 'error'"
            :message="testResult.message"
            show-icon
          />
        </a-form-item>

        <a-form-item :wrapper-col="{ offset: 6, span: 18 }">
          <a-space>
            <a-button :loading="testing" @click="handleTestConfig">测试连通性</a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, h } from 'vue'
import { message, Modal, Tag, type TableColumnsType } from 'ant-design-vue'
import type { FormInstance } from 'ant-design-vue/es/form'
import { storageApi } from '@/api/storage'
import type {
  StorageConfig,
  StorageConfigForm,
  StorageTestForm,
  MigrationTask,
  MigrationItem,
  AnalyzeResult,
} from '@/types/storage'

// ========== 通用状态 ==========
const activeTab = ref<'config' | 'migration'>('config')

// ========== Tab 1: 存储配置 ==========
const configs = ref<StorageConfig[]>([])
const configLoading = ref(false)

const hasPendingConfig = computed(() => configs.value.some((c) => !c.is_active))

const availableProviders = computed(() => {
  const activeProvider = configs.value.find((c) => c.is_active)?.provider
  return [
    { value: 'local', label: '本地存储' },
    { value: 'aliyun', label: '阿里云 OSS' },
    { value: 'tencent', label: '腾讯云 COS' },
    { value: 'minio', label: 'MinIO' },
  ].filter((p) => p.value !== activeProvider)
})

const providerLabel = (p: string): string => {
  const map: Record<string, string> = {
    local: '本地存储',
    aliyun: '阿里云 OSS',
    tencent: '腾讯云 COS',
    minio: 'MinIO',
  }
  return map[p] || p
}

const configColumns: TableColumnsType<StorageConfig> = [
  { title: '存储平台', key: 'provider', width: 140 },
  { title: 'Endpoint', dataIndex: 'endpoint', key: 'endpoint', ellipsis: true },
  { title: 'Bucket', dataIndex: 'bucket', key: 'bucket', width: 180 },
  { title: '状态', key: 'status', width: 120 },
  { title: '更新时间', key: 'updated_at', width: 180 },
  { title: '操作', key: 'action', width: 220 },
]

async function fetchConfigs() {
  configLoading.value = true
  try {
    const res = await storageApi.getConfigs()
    configs.value = res || []
  } catch {
    configs.value = []
  } finally {
    configLoading.value = false
  }
}

async function handleActivate(record: StorageConfig) {
  Modal.confirm({
    title: '确认激活',
    content: `激活后存储平台将切换为「${providerLabel(record.provider)}」，旧配置将被删除。请确保已将所有素材迁移到目标平台。确定继续？`,
    onOk: async () => {
      try {
        await storageApi.activateConfig(record.provider)
        message.success('已激活，旧配置已清理')
        await fetchConfigs()
      } catch (e: any) {
        message.error(e?.message || '激活失败，请先完成素材迁移')
      }
    },
  })
}

async function handleDelete(record: StorageConfig) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除 ${providerLabel(record.provider)} 配置吗？`,
    onOk: async () => {
      try {
        await storageApi.deleteConfig(record.provider)
        message.success('已删除')
        await fetchConfigs()
      } catch (e: any) {
        message.error(e?.message || '已激活的配置不能直接删除')
      }
    },
  })
}

// ========== Modal 表单 ==========
const formRef = ref<FormInstance>()
const modalVisible = ref(false)
const modalLoading = ref(false)
const editingConfig = ref<StorageConfig | null>(null)
const testing = ref(false)
const testResult = ref<{ success: boolean; message: string } | null>(null)

const defaultForm = (): StorageConfigForm => ({
  provider: '',
  old_provider: '',
  endpoint: '',
  region: '',
  bucket: '',
  access_key: '',
  access_secret: '',
  path_prefix: '',
  custom_domain: '',
  extra: '',
  use_ssl: true,
})

const form = ref<StorageConfigForm>(defaultForm())

function handleAdd() {
  editingConfig.value = null
  testResult.value = null
  form.value = defaultForm()
  modalVisible.value = true
}

function handleEdit(record: StorageConfig) {
  editingConfig.value = record
  testResult.value = null
  form.value = {
    provider: record.provider,
    old_provider: record.provider,
    endpoint: record.endpoint,
    region: record.region,
    bucket: record.bucket,
    access_key: record.access_key,
    access_secret: '', // 编辑时留空表示不修改
    path_prefix: record.path_prefix,
    custom_domain: record.custom_domain,
    extra: record.extra,
    use_ssl: parseUseSsl(record.extra),
  }
  modalVisible.value = true
}

function parseUseSsl(extra: string): boolean {
  if (!extra) return true
  try {
    const obj = JSON.parse(extra)
    return obj.use_ssl !== false
  } catch {
    return true
  }
}

function closeModal() {
  modalVisible.value = false
  editingConfig.value = null
  testResult.value = null
}

// 表单校验规则
const formRules: Record<string, any> = {
  provider: [{ required: true, message: '请选择存储平台', trigger: 'change' }],
  endpoint: [
    {
      validator: (_: any, value: string) => {
        if (form.value.provider !== 'local' && !value) {
          return Promise.reject(new Error('请输入 Endpoint'))
        }
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
  bucket: [
    {
      validator: (_: any, value: string) => {
        if (form.value.provider !== 'tencent' && form.value.provider !== 'local' && !value) {
          return Promise.reject(new Error('请输入 Bucket'))
        }
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
  access_key: [
    {
      validator: (_: any, value: string) => {
        if (form.value.provider !== 'local' && !value) {
          return Promise.reject(new Error('请输入 Access Key'))
        }
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
  access_secret: [
    {
      validator: (_: any, value: string) => {
        if (form.value.provider !== 'local' && !editingConfig.value && !value) {
          return Promise.reject(new Error('请输入 Access Secret'))
        }
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
  region: [
    {
      validator: (_: any, value: string) => {
        if ((form.value.provider === 'aliyun' || form.value.provider === 'tencent') && !value) {
          return Promise.reject(new Error('Region 为必填'))
        }
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
}

// Provider 动态提示
const endpointPlaceholder = computed(() => {
  switch (form.value.provider) {
    case 'aliyun': return 'https://oss-cn-hangzhou.aliyuncs.com'
    case 'tencent': return 'https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com'
    case 'minio': return 'play.min.io:9000'
    default: return '请输入 Endpoint'
  }
})

const endpointExample = computed(() => {
  switch (form.value.provider) {
    case 'aliyun': return 'https://oss-cn-hangzhou.aliyuncs.com'
    case 'tencent': return 'https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com'
    case 'minio': return 'play.min.io:9000（不含 scheme）'
    default: return '—'
  }
})

const bucketPlaceholder = computed(() => {
  switch (form.value.provider) {
    case 'aliyun': return '纯桶名，如 my-bucket'
    case 'tencent': return '必须带 APPID，如 examplebucket-1250000000'
    case 'minio': return '纯桶名，如 my-bucket'
    default: return '请输入 Bucket'
  }
})

const bucketExample = computed(() => {
  switch (form.value.provider) {
    case 'aliyun': return 'my-bucket'
    case 'tencent': return 'examplebucket-1250000000（必须带 APPID）'
    case 'minio': return 'my-bucket'
    default: return '—'
  }
})

function handleProviderChange() {
  // 切换 provider 时重置 extra
  form.value.extra = ''
  testResult.value = null
}

// 测试连通性
async function handleTestConfig() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  testing.value = true
  testResult.value = null
  try {
    const data: StorageTestForm = {
      provider: form.value.provider as 'aliyun' | 'tencent' | 'minio' | 'local',
      endpoint: form.value.endpoint,
      region: form.value.region,
      bucket: form.value.bucket,
      access_key: form.value.access_key,
      access_secret: form.value.access_secret,
      path_prefix: form.value.path_prefix,
      custom_domain: form.value.custom_domain,
      extra: buildExtra(),
    }
    await storageApi.testConfig(data)
    testResult.value = { success: true, message: '连通性测试通过' }
  } catch (e: any) {
    testResult.value = { success: false, message: e?.message || '连通性测试失败' }
  } finally {
    testing.value = false
  }
}

function buildExtra(): string {
  if (form.value.provider === 'minio') {
    return JSON.stringify({ use_ssl: form.value.use_ssl ?? true })
  }
  return form.value.extra || ''
}

// 保存
async function handleSave() {
  // 检测平台类型是否被篡改
  if (editingConfig.value && form.value.provider !== editingConfig.value.provider) {
    Modal.warning({
      title: '禁止修改存储平台',
      content: '检测到存储平台类型发生变化！禁止直接修改存储平台类型。如需切换存储平台，请先在「素材迁移」中将已有素材迁移至目标平台，再激活目标平台配置。',
    })
    return
  }

  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  modalLoading.value = true
  try {
    const payload: StorageConfigForm = {
      ...form.value,
      extra: buildExtra(),
    }
    await storageApi.upsertConfig(payload)
    if (editingConfig.value) {
      message.success('配置已更新')
      closeModal()
      await fetchConfigs()
    } else {
      message.success('配置已创建，请前往「素材迁移」完成迁移后再激活')
      closeModal()
      await fetchConfigs()
      activeTab.value = 'migration'
    }
  } catch (e: any) {
    message.error(e?.message || '保存失败')
  } finally {
    modalLoading.value = false
  }
}

// ========== Tab 2: 素材迁移 ==========
const activeConfig = computed(() => configs.value.find((c) => c.is_active) || null)
// 迁移目标：优先使用待激活的配置，其次使用当前活跃配置
const targetConfig = computed(() => configs.value.find((c) => !c.is_active) || activeConfig.value || null)

const sourcePlatforms = computed(() => {
  if (!analyzeResult.value) return []
  const platforms = new Set<string>()
  for (const item of (analyzeResult.value.missing || [])) {
    if (item.storage_type && item.storage_type !== targetConfig.value?.provider) {
      platforms.add(item.storage_type)
    }
  }
  return Array.from(platforms)
})

const sourcePlatformLabels = computed(() => {
  return sourcePlatforms.value.map(p => providerLabel(p)).join('、')
})

// 分析状态
const analyzeState = ref<'idle' | 'analyzing' | 'completed' | 'failed'>('idle')
const analyzeResult = ref<AnalyzeResult | null>(null)
let analyzePollTimer: ReturnType<typeof setInterval> | null = null

// 迁移状态
const migrationState = ref<'idle' | 'migrating' | 'completed' | 'failed'>('idle')
const migrateTask = ref<MigrationTask | null>(null)
const migrateItems = ref<MigrationItem[]>([])
let migratePollTimer: ReturnType<typeof setInterval> | null = null

const selectedRowKeys = ref<string[]>([])

// 分析结果表格列
const existingColumns: TableColumnsType<MigrationItem> = [
  { title: '类型', dataIndex: 'source_type', key: 'source_type', width: 80, customRender: ({ text }: { text: string }) => text === 'preset' ? '预设' : '素材' },
  { title: '文件名', dataIndex: 'filename', key: 'filename', ellipsis: true },
  { title: '当前平台', dataIndex: 'storage_type', key: 'storage_type', customRender: ({ text }: { text: string }) => providerLabel(text) },
  { title: 'URL', dataIndex: 'url', key: 'url', ellipsis: true },
  {
    title: '状态',
    key: 'status',
    width: 100,
    customRender: () => h(Tag, { color: 'green' }, () => '存在 ✅'),
  },
]

const missingColumns: TableColumnsType<MigrationItem> = [
  { title: '类型', dataIndex: 'source_type', key: 'source_type', width: 80, customRender: ({ text }: { text: string }) => text === 'preset' ? '预设' : '素材' },
  { title: '文件名', dataIndex: 'filename', key: 'filename', ellipsis: true },
  { title: '当前平台', dataIndex: 'storage_type', key: 'storage_type', customRender: ({ text }: { text: string }) => providerLabel(text) },
  { title: 'URL', dataIndex: 'url', key: 'url', ellipsis: true },
  {
    title: '状态',
    key: 'status',
    width: 100,
    customRender: () => h(Tag, { color: 'orange' }, () => '缺失 ⚠️'),
  },
]

const failedColumns: TableColumnsType<MigrationItem> = [
  { title: 'Media ID', dataIndex: 'media_id', key: 'media_id' },
  { title: '文件名', dataIndex: 'filename', key: 'filename', ellipsis: true },
  { title: '错误信息', dataIndex: 'error', key: 'error', ellipsis: true },
]

// 迁移进度
const migratePercent = computed(() => {
  if (!migrateTask.value || migrateTask.value.total === 0) return 0
  return Math.round((migrateTask.value.succeeded / migrateTask.value.total) * 100)
})

const migrateProgressStatus = computed(() => {
  if (migrationState.value === 'failed') return 'exception'
  if (migrationState.value === 'completed') return 'success'
  return 'active'
})

const migrateSucceeded = computed(() => migrateTask.value?.succeeded ?? 0)
const migrateTotal = computed(() => migrateTask.value?.total ?? 0)

const migrateFailedItems = computed(() =>
  migrateItems.value.filter((item) => item.status === 'failed'),
)

const elapsedTime = computed(() => {
  if (!migrateTask.value?.started_at || !migrateTask.value?.finished_at) return 0
  const start = new Date(migrateTask.value.started_at).getTime()
  const end = new Date(migrateTask.value.finished_at).getTime()
  if (Number.isNaN(start) || Number.isNaN(end)) return 0
  return Math.round((end - start) / 1000)
})

// 开始分析
async function handleAnalyze() {
  if (!targetConfig.value) {
    message.warning('请先添加存储平台配置')
    return
  }

  analyzeState.value = 'analyzing'
  analyzeResult.value = null
  selectedRowKeys.value = []

  try {
    const task = await storageApi.analyze(targetConfig.value.provider)
    // 轮询分析结果
    analyzePollTimer = setInterval(async () => {
      try {
        const result = await storageApi.getAnalyzeResult(task.task_id)
        if (result.task.status === 'completed') {
          analyzeState.value = 'completed'
          analyzeResult.value = result
          stopAnalyzePoll()
        } else if (result.task.status === 'failed') {
          analyzeState.value = 'failed'
          message.error(result.task.error || '分析失败')
          stopAnalyzePoll()
        }
      } catch {
        // 继续轮询
      }
    }, 2000)
  } catch (e: any) {
    analyzeState.value = 'failed'
    message.error(e?.message || '启动分析失败')
  }
}

function stopAnalyzePoll() {
  if (analyzePollTimer) {
    clearInterval(analyzePollTimer)
    analyzePollTimer = null
  }
}

// 一键迁移所有
async function handleMigrateAll() {
  if (!targetConfig.value || !analyzeResult.value) return
  await startMigration({ target_provider: targetConfig.value.provider, all: true })
}

// 批量迁移
async function handleBatchMigrate() {
  if (!targetConfig.value || !analyzeResult.value || selectedRowKeys.value.length === 0) return
  await startMigration({
    target_provider: targetConfig.value.provider,
    media_ids: selectedRowKeys.value,
  })
}

async function startMigration(params: { target_provider: string; all?: boolean; media_ids?: string[] }) {
  migrationState.value = 'migrating'
  migrateTask.value = null
  migrateItems.value = []

  try {
    const task = await storageApi.startMigration(params)

    migratePollTimer = setInterval(async () => {
      try {
        const status = await storageApi.getMigrationStatus(task.task_id)
        migrateTask.value = status.task
        migrateItems.value = status.failed || []

        if (status.task.status === 'completed') {
          migrationState.value = 'completed'
          stopMigratePoll()
          message.success('迁移完成')
          // 提示用户激活新平台
          const pendingConfig = configs.value.find((c) => !c.is_active)
          if (pendingConfig) {
            Modal.confirm({
              title: '迁移完成',
              content: `所有素材已迁移完成，是否立即激活「${providerLabel(pendingConfig.provider)}」？激活后旧配置将被删除。`,
              onOk: async () => {
                try {
                  await storageApi.activateConfig(pendingConfig.provider)
                  message.success('已激活，旧配置已清理')
                  await fetchConfigs()
                  activeTab.value = 'config'
                } catch (e: any) {
                  message.error(e?.message || '激活失败')
                }
              },
            })
          }
        } else if (status.task.status === 'failed') {
          migrationState.value = 'failed'
          stopMigratePoll()
          message.error(status.task.error || '迁移失败')
        }
      } catch {
        // 继续轮询
      }
    }, 2000)
  } catch (e: any) {
    migrationState.value = 'failed'
    message.error(e?.message || '启动迁移失败')
  }
}

function stopMigratePoll() {
  if (migratePollTimer) {
    clearInterval(migratePollTimer)
    migratePollTimer = null
  }
}

// 重置分析/迁移状态
function handleResetAnalyze() {
  stopAnalyzePoll()
  stopMigratePoll()
  analyzeState.value = 'idle'
  analyzeResult.value = null
  migrationState.value = 'idle'
  migrateTask.value = null
  migrateItems.value = []
  selectedRowKeys.value = []
}

// ========== 生命周期 ==========
onMounted(() => {
  fetchConfigs()
})

onUnmounted(() => {
  stopAnalyzePoll()
  stopMigratePoll()
})

// 切换 Tab 时重置迁移状态
watch(activeTab, (val) => {
  if (val === 'config') {
    handleResetAnalyze()
  }
  if (val === 'migration') {
    fetchConfigs()
  }
})

function formatDateTime(time?: string): string {
  if (!time) return '-'
  const d = new Date(time)
  if (Number.isNaN(d.getTime())) return time
  return d.toLocaleString('zh-CN')
}
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: flex-start;
  gap: 12px;
  margin: 16px 0 20px;
}

.migration-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.field-hint {
  color: #999;
  font-size: 12px;
}

:deep(.existing-row) {
  background-color: #fafafa;
}
</style>
