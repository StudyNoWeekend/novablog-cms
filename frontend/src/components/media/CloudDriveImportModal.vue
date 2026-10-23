<template>
  <a-modal
    :open="visible"
    :title="$t('cloudDrive.import.title')"
    :width="760"
    :footer="null"
    :mask-closable="phase !== 'importing'"
    @cancel="handleClose"
  >
    <!-- 阶段一：选择与浏览 -->
    <template v-if="phase === 'browse'">
      <a-empty
        v-if="accounts.length === 0 && !accountsLoading"
        :description="$t('cloudDrive.import.noAccount')"
        style="margin: 24px 0"
      >
        <a-button type="primary" @click="goBindAccounts">
          <CloudOutlined /> {{ $t('cloudDrive.import.goBind') }}
        </a-button>
      </a-empty>

      <template v-else>
        <div class="account-row">
          <span class="field-label">{{ $t('cloudDrive.import.account') }}</span>
          <a-select
            v-model:value="selectedAccountId"
            class="account-select"
            :placeholder="$t('cloudDrive.import.selectAccount')"
            :loading="accountsLoading"
            @change="handleAccountChange"
          >
            <a-select-option v-for="account in accounts" :key="account.id" :value="account.id">
              {{ providerLabel(account.provider) }} · {{ account.display_name || account.username }}
            </a-select-option>
          </a-select>
        </div>

        <div class="browse-head">
          <a-breadcrumb>
            <a-breadcrumb-item>
              <a @click.prevent="enterBreadcrumb(-1)">{{ $t('cloudDrive.import.breadcrumbRoot') }}</a>
            </a-breadcrumb-item>
            <a-breadcrumb-item v-for="(crumb, idx) in breadcrumbs" :key="crumb.parent">
              <a @click.prevent="enterBreadcrumb(idx)">{{ crumb.name }}</a>
            </a-breadcrumb-item>
          </a-breadcrumb>
          <a-checkbox
            :checked="allPageSelected"
            :indeterminate="somePageSelected && !allPageSelected"
            :disabled="pageFiles.length === 0"
            @change="toggleSelectAll"
          >
            {{ $t('cloudDrive.import.selectAll') }}
          </a-checkbox>
        </div>

        <a-spin :spinning="loadingFiles">
          <div class="file-list">
            <a-empty
              v-if="!loadingFiles && files.length === 0"
              :description="$t('cloudDrive.import.empty')"
              style="margin: 32px 0"
            />
            <div v-for="file in files" :key="file.file_id" class="file-row" :class="{ dir: file.is_dir }">
              <template v-if="file.is_dir">
                <span class="file-icon"><FolderOutlined /></span>
                <a class="file-name" :title="file.name" @click.prevent="enterFolder(file)">{{ file.name }}</a>
              </template>
              <template v-else>
                <a-checkbox
                  class="file-check"
                  :checked="selectedIds.has(file.file_id)"
                  @change="() => toggleSelect(file)"
                />
                <span class="file-icon" :class="fileExtClass(file)"><FileOutlined /></span>
                <span class="file-name" :title="file.name">{{ file.name }}</span>
                <span class="file-size">{{ formatSize(file.size) }}</span>
              </template>
            </div>
          </div>
          <div v-if="nextCursor" class="load-more">
            <a-button size="small" :loading="loadingMore" @click="loadMore">
              {{ $t('cloudDrive.import.loadMore') }}
            </a-button>
          </div>
        </a-spin>

        <div class="target-row">
          <span class="field-label">{{ $t('cloudDrive.import.targetFolder') }}</span>
          <a-tree-select
            v-model:value="targetKey"
            class="target-select"
            :tree-data="folderTreeData"
            :loading="folderLoading"
            tree-default-expand-all
            :tree-line="true"
          >
            <template #title="{ title, key }">
              <span><FolderOutlined v-if="key !== ROOT_KEY" /> {{ title }}</span>
            </template>
          </a-tree-select>
        </div>

        <div class="selection-summary">
          {{ $t('cloudDrive.import.selectedCount', { count: selectedFiles.length }) }}
          <span v-if="overLimit" class="over-limit">{{ $t('cloudDrive.import.overLimit') }}</span>
        </div>
      </template>
    </template>

    <!-- 阶段二/三：导入进度与结果 -->
    <template v-else>
      <div class="progress-area">
        <a-progress
          :percent="progressPercent"
          :status="taskFailed ? 'exception' : progressPercent >= 100 ? 'success' : 'active'"
        />
        <div class="progress-stats">
          <span>{{ $t('cloudDrive.import.doneCount', { done: doneCount, total: task?.total ?? 0 }) }}</span>
          <span class="stat-ok">{{ $t('cloudDrive.import.successCount', { count: task?.succeeded ?? 0 }) }}</span>
          <span class="stat-fail" v-if="(task?.failed ?? 0) > 0">
            {{ $t('cloudDrive.import.failCount', { count: task?.failed ?? 0 }) }}
          </span>
        </div>

        <!-- 逐文件进度：网盘限速下载时每个文件独立进度条 + 实时速度 -->
        <div class="item-list">
          <div v-for="item in importItems" :key="item.id" class="item-row">
            <span class="item-icon">
              <CheckCircleFilled v-if="item.status === 'success'" class="icon-ok" />
              <CloseCircleFilled v-else-if="item.status === 'failed'" class="icon-fail" />
              <LoadingOutlined v-else-if="isDownloading(item)" spin />
              <ClockCircleOutlined v-else class="icon-queued" />
            </span>
            <span class="item-name" :title="item.error || item.filename">{{ item.filename }}</span>
            <span class="item-size">{{ formatSize(item.file_size) }}</span>
            <div class="item-progress">
              <a-progress
                :percent="itemPercent(item)"
                size="small"
                :show-info="item.file_size > 0"
                :status="item.status === 'failed' ? 'exception' : item.status === 'success' ? 'success' : 'active'"
              />
            </div>
            <span class="item-speed">{{ itemSpeedText(item) }}</span>
          </div>
        </div>

        <template v-if="phase === 'done'">
          <a-alert
            v-if="taskFailed && task?.error"
            class="task-error"
            type="warning"
            show-icon
            :message="task.error"
          />
          <a-collapse v-if="failedItems.length > 0" class="failed-collapse">
            <a-collapse-panel :header="$t('cloudDrive.import.failedDetail', { count: failedItems.length })">
              <div v-for="item in failedItems" :key="item.id" class="failed-item">
                <span class="failed-name">{{ item.filename }}</span>
                <span class="failed-reason">{{ item.error }}</span>
              </div>
            </a-collapse-panel>
          </a-collapse>
          <div class="done-actions">
            <a-button type="primary" @click="handleClose">{{ $t('cloudDrive.import.close') }}</a-button>
          </div>
        </template>
      </div>
    </template>

    <!-- 浏览阶段底部操作 -->
    <div v-if="phase === 'browse' && accounts.length > 0" class="modal-actions">
      <a-button @click="handleClose">{{ $t('cloudDrive.import.cancel') }}</a-button>
      <a-button
        type="primary"
        :disabled="selectedFiles.length === 0 || overLimit"
        :loading="starting"
        @click="handleStartImport"
      >
        <CloudDownloadOutlined /> {{ $t('cloudDrive.import.start', { count: selectedFiles.length }) }}
      </a-button>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  CloudOutlined,
  CloudDownloadOutlined,
  CheckCircleFilled,
  CloseCircleFilled,
  ClockCircleOutlined,
  FileOutlined,
  FolderOutlined,
  LoadingOutlined,
} from '@ant-design/icons-vue'
import { cloudDriveApi } from '@/api/cloud-drive'
import type {
  CloudDriveAccount,
  CloudDriveProvider,
  CloudFile,
  CloudImportTask,
  CloudImportItem,
} from '@/types/cloud-drive'
import { type MediaFolderNode } from '@/api/media'
import { useMediaFolders, folderDisplayName } from '@/composables/useMediaFolders'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'imported'): void
}>()

const { t } = useI18n()
const router = useRouter()
const { folderTree, folderLoading, fetchFolderTree } = useMediaFolders()

const ROOT_KEY = '__root__'

type Phase = 'browse' | 'importing' | 'done'

const phase = ref<Phase>('browse')
const accounts = ref<CloudDriveAccount[]>([])
const accountsLoading = ref(false)
const selectedAccountId = ref<string>('')

const files = ref<CloudFileItem[]>([])
const loadingFiles = ref(false)
const loadingMore = ref(false)
const nextCursor = ref('')
/** 当前目录标识：webdav/百度=路径（根 "/"），阿里云盘=file_id（根 "root"） */
const currentParent = ref('')
const breadcrumbs = ref<{ name: string; parent: string }[]>([])

interface CloudFileItem {
  file_id: string
  parent_id: string
  path: string
  name: string
  is_dir: boolean
  size: number
  ext: string
}

const selectedFiles = ref<CloudFileItem[]>([])
const selectedIds = computed(() => new Set(selectedFiles.value.map((f) => f.file_id)))

const targetKey = ref<string>(ROOT_KEY)
const starting = ref(false)

const task = ref<CloudImportTask | null>(null)
const failedItems = ref<CloudImportItem[]>([])
/** 全部条目（含下载中条目的 transferred_bytes，逐文件进度条数据源） */
const importItems = ref<CloudImportItem[]>([])
/** 相邻轮询快照：itemID → {bytes, ts}，用于计算实时下载速度 */
const speedSamples = new Map<string, { bytes: number; ts: number }>()
const lastSpeeds = reactive<Record<string, number>>({})
let pollTimer: ReturnType<typeof setInterval> | null = null

const selectedProvider = computed(
  () => accounts.value.find((a) => a.id === selectedAccountId.value)?.provider ?? '',
)

/** 当前页的文件（非文件夹） */
const pageFiles = computed(() => files.value.filter((f) => !f.is_dir))
const allPageSelected = computed(
  () => pageFiles.value.length > 0 && pageFiles.value.every((f) => selectedIds.value.has(f.file_id)),
)
const somePageSelected = computed(() => pageFiles.value.some((f) => selectedIds.value.has(f.file_id)))

/** 后端限制单次 200 个，前端提前拦截给出反馈 */
const IMPORT_LIMIT = 200
const overLimit = computed(() => selectedFiles.value.length > IMPORT_LIMIT)

const doneCount = computed(() => (task.value?.succeeded ?? 0) + (task.value?.failed ?? 0))
const progressPercent = computed(() => {
  const total = task.value?.total ?? 0
  if (total <= 0) return 0
  return Math.min(100, Math.round((doneCount.value / total) * 100))
})
const taskFailed = computed(() => (task.value?.failed ?? 0) > 0)

// ===== 逐文件进度与速度 =====

/** 串行导入：第一个 pending 条目即当前下载中的文件 */
const firstPendingId = computed(() => importItems.value.find((i) => i.status === 'pending')?.id ?? '')

function isDownloading(item: CloudImportItem): boolean {
  return item.status === 'pending' && (item.transferred_bytes > 0 || item.id === firstPendingId.value)
}

function itemPercent(item: CloudImportItem): number {
  if (item.status === 'success') return 100
  if (item.file_size <= 0) return 0
  return Math.min(100, Math.round((item.transferred_bytes / item.file_size) * 100))
}

/** 用相邻两次轮询的字节差计算实时速度（后端 ≥1s 节流回写，与 2s 轮询粒度匹配） */
function updateSpeeds(items: CloudImportItem[]) {
  const now = Date.now()
  for (const item of items) {
    if (item.status !== 'pending') {
      speedSamples.delete(item.id)
      continue
    }
    const prev = speedSamples.get(item.id)
    if (prev) {
      const dt = (now - prev.ts) / 1000
      const db = item.transferred_bytes - prev.bytes
      if (dt > 0 && db > 0) {
        lastSpeeds[item.id] = db / dt
      } else if (!(item.id in lastSpeeds)) {
        lastSpeeds[item.id] = 0
      }
    }
    speedSamples.set(item.id, { bytes: item.transferred_bytes, ts: now })
  }
}

function itemSpeedText(item: CloudImportItem): string {
  if (item.status !== 'pending') return ''
  const speed = lastSpeeds[item.id]
  if (!speed) return ''
  return `${formatSize(Math.round(speed))}/s`
}

function providerLabel(provider: string): string {
  return t(`cloudDrive.tabs.${provider}`)
}

function fileExtClass(file: CloudFileItem): string {
  if (['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg', '.bmp', '.ico'].includes(file.ext)) return 'is-image'
  if (['.mp4', '.avi', '.mov', '.mkv', '.webm'].includes(file.ext)) return 'is-video'
  if (['.mp3', '.wav', '.flac', '.aac', '.ogg'].includes(file.ext)) return 'is-audio'
  return ''
}

function formatSize(size: number): string {
  if (size <= 0) return '-'
  if (size >= 1024 ** 3) return `${(size / 1024 ** 3).toFixed(1)} GB`
  if (size >= 1024 ** 2) return `${(size / 1024 ** 2).toFixed(1)} MB`
  return `${Math.max(1, Math.round(size / 1024))} KB`
}

// ===== 数据加载 =====

async function fetchAccounts() {
  accountsLoading.value = true
  try {
    accounts.value = (await cloudDriveApi.getAccounts()) || []
  } catch (e: any) {
    message.error(e?.message || t('cloudDrive.errors.loadFailed'))
  } finally {
    accountsLoading.value = false
  }
}

function rootParentFor(provider: string): string {
  return provider === 'aliyun' ? 'root' : '/'
}

async function loadFiles(parent: string, cursor = '') {
  if (!selectedAccountId.value) return
  if (cursor === '') {
    loadingFiles.value = true
    files.value = []
  } else {
    loadingMore.value = true
  }
  try {
    const res = await cloudDriveApi.listFiles(selectedAccountId.value, {
      parent,
      cursor: cursor || undefined,
      limit: 100,
    })
    if (cursor === '') {
      files.value = res.list || []
    } else {
      files.value = [...files.value, ...(res.list || [])]
    }
    nextCursor.value = res.next_cursor || ''
    currentParent.value = parent
  } catch (e: any) {
    message.error(e?.message || t('cloudDrive.errors.listFailed'))
    if (cursor === '') {
      files.value = []
      nextCursor.value = ''
    }
  } finally {
    loadingFiles.value = false
    loadingMore.value = false
  }
}

async function loadMore() {
  if (!nextCursor.value) return
  await loadFiles(currentParent.value, nextCursor.value)
}

function handleAccountChange() {
  selectedFiles.value = []
  breadcrumbs.value = []
  const provider = selectedProvider.value
  const root = rootParentFor(provider)
  currentParent.value = root
  loadFiles(root)
}

function enterFolder(file: CloudFileItem) {
  // webdav/百度按路径列目录，阿里云盘按 file_id
  const parent = selectedProvider.value === 'aliyun' ? file.file_id : file.path
  breadcrumbs.value = [...breadcrumbs.value, { name: file.name, parent }]
  selectedFiles.value = selectedFiles.value.filter(
    (f) => !files.value.some((pf) => pf.file_id === f.file_id),
  )
  loadFiles(parent)
}

function enterBreadcrumb(idx: number) {
  // -1=根目录；否则截断面包屑到第 idx 层
  if (idx < 0) {
    breadcrumbs.value = []
    loadFiles(rootParentFor(selectedProvider.value))
    return
  }
  breadcrumbs.value = breadcrumbs.value.slice(0, idx + 1)
  loadFiles(breadcrumbs.value[idx].parent)
}

// ===== 选择 =====

function toggleSelect(file: CloudFileItem) {
  const exists = selectedFiles.value.some((f) => f.file_id === file.file_id)
  if (exists) {
    selectedFiles.value = selectedFiles.value.filter((f) => f.file_id !== file.file_id)
  } else {
    selectedFiles.value = [...selectedFiles.value, file]
  }
}

function toggleSelectAll() {
  if (allPageSelected.value) {
    const pageIds = new Set(pageFiles.value.map((f) => f.file_id))
    selectedFiles.value = selectedFiles.value.filter((f) => !pageIds.has(f.file_id))
  } else {
    const existing = new Set(selectedFiles.value.map((f) => f.file_id))
    const merged = [...selectedFiles.value]
    for (const f of pageFiles.value) {
      if (!existing.has(f.file_id)) merged.push(f)
    }
    selectedFiles.value = merged
  }
}

// ===== 目标文件夹 =====

interface TreeItem {
  title: string
  key: string
  children?: TreeItem[]
}
const folderTreeData = computed<TreeItem[]>(() => {
  const build = (nodes: MediaFolderNode[]): TreeItem[] =>
    nodes.map((n) => ({
      title: folderDisplayName(n.name, n.module_key),
      key: n.id,
      children: build(n.children || []),
    }))
  return [{ title: t('cloudDrive.import.rootFolder'), key: ROOT_KEY, children: build(folderTree.value) }]
})
// ===== 导入 =====

async function handleStartImport() {
  if (selectedFiles.value.length === 0) return
  starting.value = true
  try {
    const res = await cloudDriveApi.startImport({
      account_id: selectedAccountId.value,
      files: selectedFiles.value.map((f) => ({
        file_id: f.file_id,
        path: f.path,
        name: f.name,
        size: f.size,
      })),
      folder_id: targetKey.value === ROOT_KEY ? null : targetKey.value,
    })
    message.success(t('cloudDrive.import.started'))
    task.value = null
    failedItems.value = []
    phase.value = 'importing'
    startPolling(res.task_id)
  } catch (e: any) {
    message.error(e?.message || t('cloudDrive.errors.importStartFailed'))
  } finally {
    starting.value = false
  }
}

function startPolling(taskId: string) {
  stopPolling()
  pollTimer = setInterval(async () => {
    try {
      const status = await cloudDriveApi.getImportStatus(taskId)
      task.value = status.task
      failedItems.value = status.failed || []
      importItems.value = status.items || []
      updateSpeeds(status.items || [])
      if (status.task.status === 'completed' || status.task.status === 'failed') {
        phase.value = 'done'
        stopPolling()
        emit('imported')
        if (status.task.succeeded > 0) {
          message.success(
            t('cloudDrive.import.completed', {
              succeeded: status.task.succeeded,
              failed: status.task.failed,
            }),
          )
        }
      }
    } catch {
      // 轮询失败继续
    }
  }, 2000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function goBindAccounts() {
  emit('update:visible', false)
  router.push('/cloud-drive')
}

function handleClose() {
  if (phase.value === 'importing') return // 导入中不允许关闭，避免轮询中断
  emit('update:visible', false)
}

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      phase.value = 'browse'
      task.value = null
      failedItems.value = []
      importItems.value = []
      speedSamples.clear()
      for (const key of Object.keys(lastSpeeds)) {
        delete lastSpeeds[key]
      }
      selectedFiles.value = []
      breadcrumbs.value = []
      files.value = []
      nextCursor.value = ''
      targetKey.value = ROOT_KEY
      fetchAccounts()
      fetchFolderTree()
    } else {
      stopPolling()
    }
  },
)

onUnmounted(stopPolling)
</script>

<style scoped>
.account-row,
.target-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.field-label {
  flex-shrink: 0;
  color: #29365c;
  font-size: 14px;
}

.account-select,
.target-select {
  flex: 1;
}

.browse-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  gap: 12px;
}

.file-list {
  min-height: 200px;
  max-height: 320px;
  overflow-y: auto;
  border: 1px solid #e5e9f2;
  border-radius: 8px;
}

.file-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-bottom: 1px solid #f0f2f7;
  cursor: pointer;
  transition: background var(--transition-fast, 0.2s);
}

.file-row:last-child {
  border-bottom: none;
}

.file-row:hover {
  background: #f6f7fb;
}

.file-row.dir .file-icon {
  color: #b7791f;
}

.file-icon {
  display: inline-flex;
  font-size: 16px;
  color: #8a93a8;
  flex-shrink: 0;
}

.file-icon.is-image {
  color: #526fe8;
}

.file-icon.is-video {
  color: #d4380d;
}

.file-icon.is-audio {
  color: #1d39c4;
}

.file-check {
  flex-shrink: 0;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #29365c;
  font-size: 13px;
}

.file-size {
  flex-shrink: 0;
  font-size: 12px;
  color: #8a93a8;
}

.load-more {
  text-align: center;
  padding: 8px 0;
}

.selection-summary {
  margin-top: 8px;
  font-size: 13px;
  color: #8a93a8;
}

.over-limit {
  color: #d4380d;
  margin-left: 8px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.progress-area {
  padding: 8px 0;
}

.item-list {
  margin-top: 12px;
  max-height: 260px;
  overflow-y: auto;
  border: 1px solid #e5e9f2;
  border-radius: 8px;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 12px;
  border-bottom: 1px solid #f0f2f7;
  font-size: 13px;
}

.item-row:last-child {
  border-bottom: none;
}

.item-icon {
  display: inline-flex;
  flex-shrink: 0;
  color: #8a93a8;
}

.item-icon .icon-ok {
  color: #389e0d;
}

.item-icon .icon-fail {
  color: #d4380d;
}

.item-icon .icon-queued {
  color: #b7bfd3;
}

.item-name {
  flex: 0 1 auto;
  max-width: 34%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #29365c;
}

.item-size {
  flex-shrink: 0;
  color: #8a93a8;
  font-size: 12px;
}

.item-progress {
  flex: 1;
  min-width: 0;
}

.item-progress :deep(.ant-progress) {
  margin: 0;
  line-height: 1;
}

.item-progress :deep(.ant-progress-line) {
  padding-bottom: 0;
}

.item-speed {
  flex-shrink: 0;
  min-width: 76px;
  text-align: right;
  color: #526fe8;
  font-variant-numeric: tabular-nums;
}

.progress-stats {
  display: flex;
  gap: 16px;
  margin-top: 8px;
  font-size: 13px;
  color: #29365c;
}

.stat-ok {
  color: #389e0d;
}

.stat-fail {
  color: #d4380d;
}

.task-error {
  margin-top: 12px;
}

.failed-collapse {
  margin-top: 12px;
}

.failed-item {
  display: flex;
  gap: 12px;
  padding: 4px 0;
  font-size: 13px;
}

.failed-name {
  color: #29365c;
  max-width: 45%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.failed-reason {
  color: #d4380d;
}

.done-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
