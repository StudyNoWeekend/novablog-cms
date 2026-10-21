<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('storage.title') }}</h1>
    </div>

    <a-card>
      <a-tabs v-model:activeKey="activeTab">
        <a-tab-pane key="config" :tab="$t('storage.tabs.config')" />
        <a-tab-pane key="migration" :tab="$t('storage.tabs.migration')" />
      </a-tabs>

      <!-- ============ Tab 1: 存储配置 ============ -->
      <template v-if="activeTab === 'config'">
        <div class="toolbar">
          <a-button v-if="!hasPendingConfig" type="primary" @click="handleAdd">{{ $t('storage.actions.switchPlatform') }}</a-button>
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
                <a-tag v-if="record.is_active" color="green">{{ $t('storage.status.active') }}</a-tag>
                <a-tag v-else color="orange">{{ $t('storage.status.pending') }}</a-tag>
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
                  {{ $t('storage.actions.activate') }}
                </a-button>
                <a-button
                  v-if="!record.is_active"
                  type="link"
                  @click="activeTab = 'migration'"
                >
                  {{ $t('storage.actions.goMigrate') }}
                </a-button>
                <a-button
                  v-if="record.is_active"
                  type="link"
                  @click="handleEdit(record)"
                >
                  {{ $t('common.actions.edit') }}
                </a-button>
                <a-button
                  v-if="!record.is_active"
                  type="link"
                  danger
                  @click="handleDelete(record)"
                >
                  {{ $t('common.actions.cancel') }}
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
              {{ $t('storage.migrate.fromTo', { sources: sourcePlatformLabels, target: providerLabel(targetConfig.provider) }) }}
            </template>
            <template v-else>
              {{ $t('storage.migrate.targetHint', { target: providerLabel(targetConfig.provider) }) }}
            </template>
          </template>
        </a-alert>
        <a-alert
          v-else
          type="warning"
          :message="$t('storage.migrate.needConfig')"
          show-icon
          style="margin-bottom: 16px"
        />

        <!-- 分析阶段 -->
        <template v-if="targetConfig && migrationState === 'idle'">
          <div v-if="analyzeState === 'idle'">
            <a-button type="primary" @click="handleAnalyze">
              {{ $t('storage.migrate.startAnalyze') }}
            </a-button>
          </div>

          <div v-if="analyzeState === 'analyzing'" style="text-align: center; padding: 48px 0">
            <a-spin :tip="$t('storage.migrate.analyzing')" />
          </div>

          <div v-if="analyzeState === 'completed' && analyzeResult">
            <a-statistic
              :title="$t('storage.migrate.analyzeResult')"
              :value="$t('storage.migrate.analyzeSummary', {
                total: (analyzeResult.missing || []).length + (analyzeResult.existing || []).length,
                existing: (analyzeResult.existing || []).length,
                missing: (analyzeResult.missing || []).length,
              })"
              style="margin-bottom: 16px"
            />
            <a-table
              :columns="existingColumns"
              :data-source="analyzeResult.existing"
              :pagination="false"
              row-key="id"
              style="margin-bottom: 16px; background: #F5F7FC"
              :row-class="() => 'existing-row'"
              bordered
            />
            <div class="migration-toolbar">
              <span style="font-weight: 500">{{ $t('storage.migrate.pendingLabel') }}</span>
              <a-button
                type="primary"
                :disabled="selectedRowKeys.length === 0"
                @click="handleBatchMigrate"
              >
                {{ $t('storage.migrate.batch', { count: selectedRowKeys.length }) }}
              </a-button>
              <a-button type="primary" @click="handleMigrateAll">
                {{ $t('storage.migrate.migrateAll') }}
              </a-button>
              <a-button @click="handleResetAnalyze">{{ $t('storage.migrate.reanalyze') }}</a-button>
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
            {{ $t('storage.migrate.progress', { done: migrateSucceeded, total: migrateTotal }) }}
            <template v-if="migrationState === 'completed'">
              {{ $t('storage.migrate.completedIn', { seconds: elapsedTime }) }}
            </template>
          </div>
          <a-button @click="handleResetAnalyze" style="margin-top: 8px">{{ $t('storage.migrate.reanalyze') }}</a-button>

          <!-- 失败项列表 -->
          <a-collapse v-if="migrateFailedItems.length > 0" style="margin-top: 16px">
            <a-collapse-panel key="failed" :header="$t('storage.migrate.failedItems', { count: migrateFailedItems.length })">
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
      :title="editingConfig ? $t('storage.modal.editTitle') : $t('storage.modal.createTitle')"
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
        <a-form-item :label="$t('storage.fields.provider')" name="provider">
          <a-select
            v-model:value="form.provider"
            :disabled="!!editingConfig"
            :placeholder="$t('storage.form.providerPlaceholder')"
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
            <span class="field-hint">{{ $t('storage.form.example', { example: endpointExample }) }}</span>
          </template>
        </a-form-item>

        <a-form-item
          v-if="form.provider === 'aliyun' || form.provider === 'tencent'"
          label="Region"
          name="region"
        >
          <a-input
            v-model:value="form.region"
            :placeholder="form.provider === 'tencent' ? $t('storage.form.regionExampleTencent') : $t('storage.form.regionExampleAliyun')"
          />
          <template #help>
            <span v-if="form.provider === 'aliyun'" class="field-hint">{{ $t('storage.form.regionRequiredAliyun') }}</span>
            <span v-else class="field-hint">{{ $t('storage.form.regionExampleTencent') }}</span>
          </template>
        </a-form-item>

        <a-form-item v-if="form.provider !== 'tencent' && form.provider !== 'local'" label="Bucket" name="bucket">
          <a-input
            v-model:value="form.bucket"
            :placeholder="bucketPlaceholder"
          />
          <template #help>
            <span class="field-hint">{{ $t('storage.form.example', { example: bucketExample }) }}</span>
          </template>
        </a-form-item>

        <a-form-item v-if="form.provider !== 'local'" :label="form.provider === 'tencent' ? 'Secret ID' : 'Access Key'" name="access_key">
          <a-input v-model:value="form.access_key" :placeholder="form.provider === 'tencent' ? $t('storage.form.secretIdPlaceholder') : $t('storage.form.accessKeyPlaceholder')" />
        </a-form-item>

        <a-form-item v-if="form.provider !== 'local'" :label="form.provider === 'tencent' ? 'Secret Key' : 'Access Secret'" name="access_secret">
          <a-input-password
            v-model:value="form.access_secret"
            :placeholder="editingConfig ? $t('storage.form.leaveEmptyToKeep') : (form.provider === 'tencent' ? $t('storage.form.secretKeyPlaceholder') : $t('storage.form.accessSecretPlaceholder'))"
          />
        </a-form-item>

        <a-form-item :label="$t('storage.form.pathPrefix')" name="path_prefix">
          <a-input v-model:value="form.path_prefix" :placeholder="$t('storage.form.pathPrefixPlaceholder')" />
        </a-form-item>

        <a-form-item :label="$t('storage.form.customDomain')" name="custom_domain">
          <a-input v-model:value="form.custom_domain" :placeholder="$t('storage.form.customDomainPlaceholder')" />
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
            <a-button :loading="testing" @click="handleTestConfig">{{ $t('storage.actions.testConnection') }}</a-button>
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
import { useI18n } from 'vue-i18n'
import { formatDateTime } from '@/utils/datetime'
import { storageApi } from '@/api/storage'
import type {
  StorageConfig,
  StorageConfigForm,
  StorageTestForm,
  MigrationTask,
  MigrationItem,
  AnalyzeResult,
} from '@/types/storage'

const { t } = useI18n()

// ========== 通用状态 ==========
const activeTab = ref<'config' | 'migration'>('config')

// ========== Tab 1: 存储配置 ==========
const configs = ref<StorageConfig[]>([])
const configLoading = ref(false)

const hasPendingConfig = computed(() => configs.value.some((c) => !c.is_active))

const availableProviders = computed(() => {
  const activeProvider = configs.value.find((c) => c.is_active)?.provider
  return [
    { value: 'local', label: t('storage.providers.local') },
    { value: 'aliyun', label: t('storage.providers.aliyun') },
    { value: 'tencent', label: t('storage.providers.tencent') },
    { value: 'minio', label: t('storage.providers.minio') },
  ].filter((p) => p.value !== activeProvider)
})

const providerLabels = computed<Record<string, string>>(() => ({
  local: t('storage.providers.local'),
  aliyun: t('storage.providers.aliyun'),
  tencent: t('storage.providers.tencent'),
  minio: t('storage.providers.minio'),
}))

const providerLabel = (p: string): string => {
  return providerLabels.value[p] || p
}

const configColumns = computed<TableColumnsType<StorageConfig>>(() => [
  { title: t('storage.fields.provider'), key: 'provider', width: 140 },
  { title: 'Endpoint', dataIndex: 'endpoint', key: 'endpoint', ellipsis: true },
  { title: 'Bucket', dataIndex: 'bucket', key: 'bucket', width: 180 },
  { title: t('common.field.status'), key: 'status', width: 120 },
  { title: t('common.field.updateTime'), key: 'updated_at', width: 180 },
  { title: t('common.table.actions'), key: 'action', width: 220 },
])

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
    title: t('storage.message.activateTitle'),
    content: t('storage.message.activateConfirm', { provider: providerLabel(record.provider) }),
    onOk: async () => {
      try {
        await storageApi.activateConfig(record.provider)
        message.success(t('storage.message.activated'))
        await fetchConfigs()
      } catch (e: any) {
        message.error(e?.message || t('storage.message.activateFailed'))
      }
    },
  })
}

async function handleDelete(record: StorageConfig) {
  Modal.confirm({
    title: t('storage.message.deleteTitle'),
    content: t('storage.message.deleteConfirm', { provider: providerLabel(record.provider) }),
    onOk: async () => {
      try {
        await storageApi.deleteConfig(record.provider)
        message.success(t('common.message.deleteSuccess'))
        await fetchConfigs()
      } catch (e: any) {
        message.error(e?.message || t('storage.message.deleteActiveDenied'))
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

// 表单校验规则（computed 包裹，切换语言时刷新）
const formRules = computed<Record<string, any>>(() => ({
  provider: [{ required: true, message: t('storage.validation.providerRequired'), trigger: 'change' }],
  endpoint: [
    {
      validator: (_: any, value: string) => {
        if (form.value.provider !== 'local' && !value) {
          return Promise.reject(new Error(t('storage.validation.endpointRequired')))
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
          return Promise.reject(new Error(t('storage.validation.bucketRequired')))
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
          return Promise.reject(new Error(t('storage.validation.accessKeyRequired')))
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
          return Promise.reject(new Error(t('storage.validation.accessSecretRequired')))
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
          return Promise.reject(new Error(t('storage.validation.regionRequired')))
        }
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
}))

// Provider 动态提示
const endpointPlaceholder = computed(() => {
  switch (form.value.provider) {
    case 'aliyun': return 'https://oss-cn-hangzhou.aliyuncs.com'
    case 'tencent': return 'https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com'
    case 'minio': return 'play.min.io:9000'
    default: return t('storage.form.endpointPlaceholder')
  }
})

const endpointExample = computed(() => {
  switch (form.value.provider) {
    case 'aliyun': return 'https://oss-cn-hangzhou.aliyuncs.com'
    case 'tencent': return 'https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com'
    case 'minio': return t('storage.form.endpointExampleMinio')
    default: return '—'
  }
})

const bucketPlaceholder = computed(() => {
  switch (form.value.provider) {
    case 'aliyun': return t('storage.form.bucketPlain')
    case 'tencent': return t('storage.form.bucketWithAppId')
    case 'minio': return t('storage.form.bucketPlain')
    default: return t('storage.form.bucketPlaceholder')
  }
})

const bucketExample = computed(() => {
  switch (form.value.provider) {
    case 'aliyun': return 'my-bucket'
    case 'tencent': return t('storage.form.bucketExampleWithAppId')
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
    testResult.value = { success: true, message: t('storage.message.testPassed') }
  } catch (e: any) {
    testResult.value = { success: false, message: e?.message || t('storage.message.testFailed') }
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
      title: t('storage.message.changeProviderDeniedTitle'),
      content: t('storage.message.changeProviderDenied'),
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
      message.success(t('storage.message.configUpdated'))
      closeModal()
      await fetchConfigs()
    } else {
      message.success(t('storage.message.configCreated'))
      closeModal()
      await fetchConfigs()
      activeTab.value = 'migration'
    }
  } catch (e: any) {
    message.error(e?.message || t('storage.message.saveFailed'))
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

// 分析结果表格列（computed 包裹，切换语言时刷新）
const existingColumns = computed<TableColumnsType<MigrationItem>>(() => [
  { title: t('common.field.type'), dataIndex: 'source_type', key: 'source_type', width: 80, customRender: ({ text }: { text: string }) => text === 'preset' ? t('storage.fields.typePreset') : t('storage.fields.typeMedia') },
  { title: t('storage.fields.filename'), dataIndex: 'filename', key: 'filename', ellipsis: true },
  { title: t('storage.fields.currentPlatform'), dataIndex: 'storage_type', key: 'storage_type', customRender: ({ text }: { text: string }) => providerLabel(text) },
  { title: 'URL', dataIndex: 'url', key: 'url', ellipsis: true },
  {
    title: t('common.field.status'),
    key: 'status',
    width: 100,
    customRender: () => h(Tag, { color: 'green' }, () => t('storage.fields.statusExisting')),
  },
])

const missingColumns = computed<TableColumnsType<MigrationItem>>(() => [
  { title: t('common.field.type'), dataIndex: 'source_type', key: 'source_type', width: 80, customRender: ({ text }: { text: string }) => text === 'preset' ? t('storage.fields.typePreset') : t('storage.fields.typeMedia') },
  { title: t('storage.fields.filename'), dataIndex: 'filename', key: 'filename', ellipsis: true },
  { title: t('storage.fields.currentPlatform'), dataIndex: 'storage_type', key: 'storage_type', customRender: ({ text }: { text: string }) => providerLabel(text) },
  { title: 'URL', dataIndex: 'url', key: 'url', ellipsis: true },
  {
    title: t('common.field.status'),
    key: 'status',
    width: 100,
    customRender: () => h(Tag, { color: 'orange' }, () => t('storage.fields.statusMissing')),
  },
])

const failedColumns = computed<TableColumnsType<MigrationItem>>(() => [
  { title: 'Media ID', dataIndex: 'media_id', key: 'media_id' },
  { title: t('storage.fields.filename'), dataIndex: 'filename', key: 'filename', ellipsis: true },
  { title: t('storage.fields.error'), dataIndex: 'error', key: 'error', ellipsis: true },
])

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
    message.warning(t('storage.migrate.addConfigFirst'))
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
          message.error(result.task.error || t('storage.message.analyzeFailed'))
          stopAnalyzePoll()
        }
      } catch {
        // 继续轮询
      }
    }, 2000)
  } catch (e: any) {
    analyzeState.value = 'failed'
    message.error(e?.message || t('storage.message.analyzeStartFailed'))
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
          message.success(t('storage.message.migrateCompleted'))
          // 提示用户激活新平台
          const pendingConfig = configs.value.find((c) => !c.is_active)
          if (pendingConfig) {
            Modal.confirm({
              title: t('storage.message.migrateCompleted'),
              content: t('storage.message.migrateActivateConfirm', { provider: providerLabel(pendingConfig.provider) }),
              onOk: async () => {
                try {
                  await storageApi.activateConfig(pendingConfig.provider)
                  message.success(t('storage.message.activated'))
                  await fetchConfigs()
                  activeTab.value = 'config'
                } catch (e: any) {
                  message.error(e?.message || t('storage.message.activateFailedSimple'))
                }
              },
            })
          }
        } else if (status.task.status === 'failed') {
          migrationState.value = 'failed'
          stopMigratePoll()
          message.error(status.task.error || t('storage.message.migrateFailed'))
        }
      } catch {
        // 继续轮询
      }
    }, 2000)
  } catch (e: any) {
    migrationState.value = 'failed'
    message.error(e?.message || t('storage.message.migrateStartFailed'))
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
  background-color: #F9FAFC;
}
</style>
