<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">媒体库</h1>
      <!-- 图片/视频分开上传：通配 accept 保证 iOS/安卓点击后直接进系统图库 -->
      <a-space wrap>
        <a-upload
          :show-upload-list="false"
          :before-upload="handleBeforeUpload"
          :custom-request="handleCustomUpload"
          :accept="IMAGE_ACCEPT"
        >
          <a-button type="primary" :loading="uploading">
            <UploadOutlined /> 上传图片
          </a-button>
        </a-upload>
        <a-upload
          :show-upload-list="false"
          :before-upload="handleBeforeUpload"
          :custom-request="handleCustomUpload"
          :accept="VIDEO_ACCEPT"
        >
          <a-button type="primary" :loading="uploading">
            <VideoCameraOutlined /> 上传视频
          </a-button>
        </a-upload>
      </a-space>
    </div>

    <a-card>
      <div class="folder-bar">
        <MediaFolderBreadcrumbs
          :path="folderPath"
          :current-id="currentFolderId"
          @navigate="enterFolder"
        />
        <a-button size="small" @click="openCreateFolder">
          <FolderAddOutlined /> 新建文件夹
        </a-button>
      </div>

      <a-tabs v-model:activeKey="activeTab" @change="handleTabChange">
        <a-tab-pane key="1">
          <template #tab>
            <span><PictureOutlined /> 图片</span>
          </template>
        </a-tab-pane>
        <a-tab-pane key="2">
          <template #tab>
            <span><VideoCameraOutlined /> 视频</span>
          </template>
        </a-tab-pane>
      </a-tabs>

      <div class="media-toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索全部文件..."
          style="width: 240px"
          allow-clear
        />
        <a-radio-group v-model:value="viewType">
          <a-radio-button value="grid">
            <AppstoreOutlined /> 图标视图
          </a-radio-button>
          <a-radio-button value="list">
            <UnorderedListOutlined /> 文件视图
          </a-radio-button>
        </a-radio-group>
      </div>

      <a-spin :spinning="loading || folderLoading">
        <a-empty
          v-if="!loading && !searching && visibleFolders.length === 0 && list.length === 0"
          :description="currentFolderId === 'root' ? '暂无媒体文件' : '此文件夹为空'"
          style="margin-top: 48px"
        />
        <template v-else>
          <!-- 图标视图 -->
          <div v-if="viewType === 'grid'">
            <!-- 文件夹区（搜索时不展示） -->
            <div v-if="!searching && visibleFolders.length > 0" class="folder-grid">
              <div
                v-for="folder in visibleFolders"
                :key="folder.id"
                class="folder-card"
                @click="enterFolder(folder.id)"
              >
                <div class="folder-body">
                  <FolderFilled class="folder-icon" />
                  <span class="folder-count">{{ folder.media_count }} 项</span>
                </div>
                <div class="folder-name">
                  <span class="folder-name-text" :title="folder.name">{{ folder.name }}</span>
                  <a-tag v-if="isSystemFolder(folder)" class="folder-system-tag" color="gold">系统</a-tag>
                </div>
                <a-dropdown :trigger="['click']">
                  <a-button
                    class="folder-more"
                    type="text"
                    size="small"
                    @click.stop
                  >
                    <MoreOutlined />
                  </a-button>
                  <template #overlay>
                    <a-menu @click="({ key }: any) => handleFolderMenu(key as string, folder)">
                      <a-menu-item key="rename" :disabled="isSystemFolder(folder)"><EditOutlined /> 重命名</a-menu-item>
                      <a-menu-item key="move" :disabled="isSystemFolder(folder)"><FolderOpenOutlined /> 移动到</a-menu-item>
                      <a-menu-item key="delete" :disabled="isSystemFolder(folder)" danger><DeleteOutlined /> 删除</a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
              </div>
            </div>

            <!-- 媒体区 -->
            <div v-if="list.length > 0" class="media-grid" :class="{ 'no-folder': searching || visibleFolders.length === 0 }">
              <div
                v-for="item in list"
                :key="item.id"
                class="media-card"
              >
                <div class="media-content">
                  <img
                    v-if="item.file_type === 1"
                    :src="item.thumb_url || getThumbUrl(item.url, 300)"
                    :alt="item.filename"
                  />
                  <div v-else class="video-card">
                    <PlayCircleOutlined class="video-icon" />
                  </div>
                  <div class="media-overlay">
                    <a-button
                      v-if="item.file_type === 1"
                      type="primary"
                      size="small"
                      ghost
                      @click.stop="openWorkbench(item)"
                    >
                      <SettingOutlined /> 预设
                    </a-button>
                    <a-button size="small" ghost @click.stop="openMoveForMedia(item)">
                      <FolderOpenOutlined /> 移动
                    </a-button>
                    <a-button size="small" danger ghost @click.stop="openDelete(item)">
                      <DeleteOutlined /> 删除
                    </a-button>
                  </div>
                </div>
                <div class="media-name" :title="item.filename">
                  {{ item.filename }}
                  <span v-if="searching && item.folder_name" class="media-folder-tag">
                    {{ item.folder_name }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- 文件视图：文件夹行 + 媒体行合并展示 -->
          <a-table
            v-else
            :columns="columns"
            :data-source="tableRows"
            :pagination="false"
            row-key="key"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <a
                  v-if="record.rowType === 'folder'"
                  class="folder-cell"
                  @click="enterFolder(record.folder.id)"
                >
                  <FolderFilled class="folder-cell-icon" />
                  <span>{{ record.folder.name }}</span>
                </a>
                <span v-else class="media-cell">{{ record.media.filename }}</span>
              </template>
              <template v-else-if="column.key === 'type'">
                <a-tag v-if="record.rowType === 'folder'" color="gold">文件夹</a-tag>
                <span v-else>{{ record.media.file_type === 1 ? '图片' : '视频' }}</span>
              </template>
              <template v-else-if="column.key === 'dimensions'">
                <span v-if="record.rowType === 'media' && record.media.width && record.media.height">
                  {{ record.media.width }}x{{ record.media.height }}
                </span>
                <span v-else>-</span>
              </template>
              <template v-else-if="column.key === 'size'">
                <span v-if="record.rowType === 'folder'">{{ record.folder.media_count }} 项</span>
                <span v-else>{{ formatBytes(record.media.size) }}</span>
              </template>
              <template v-else-if="column.key === 'created_at'">
                {{ formatDateTime(record.rowType === 'folder' ? record.folder.created_at : record.media.created_at) }}
              </template>
              <template v-else-if="column.key === 'actions'">
                <template v-if="record.rowType === 'folder'">
                  <a-tooltip :title="isSystemFolder(record.folder) ? '系统文件夹不支持重命名' : undefined">
                    <a-button type="link" size="small" :disabled="isSystemFolder(record.folder)" @click="handleFolderMenu('rename', record.folder)">重命名</a-button>
                  </a-tooltip>
                  <a-tooltip :title="isSystemFolder(record.folder) ? '系统文件夹不支持移动' : undefined">
                    <a-button type="link" size="small" :disabled="isSystemFolder(record.folder)" @click="handleFolderMenu('move', record.folder)">移动</a-button>
                  </a-tooltip>
                  <a-tooltip :title="isSystemFolder(record.folder) ? '系统文件夹不支持删除' : undefined">
                    <a-button type="link" danger size="small" :disabled="isSystemFolder(record.folder)" @click="handleFolderMenu('delete', record.folder)">删除</a-button>
                  </a-tooltip>
                </template>
                <template v-else>
                  <a-button type="link" size="small" @click="openMoveForMedia(record.media)">移动</a-button>
                  <a-button type="link" danger size="small" @click="openDelete(record.media)">删除</a-button>
                </template>
              </template>
            </template>
          </a-table>
        </template>
      </a-spin>

      <div v-if="total > 0" class="pagination-wrapper">
        <a-pagination
          :current="page"
          :page-size="pageSize"
          :total="total"
          :page-size-options="[20, 40, 60]"
          show-size-changer
          :show-total="(t: number) => `共 ${t} 条`"
          @change="handlePageChange"
          @show-size-change="handlePageChange"
        />
      </div>
    </a-card>

    <MediaPresetWorkbench
      v-if="selectedMedia"
      v-model:visible="workbenchVisible"
      :media-id="selectedMedia.id"
      :original-url="selectedMedia.url"
      :filename="selectedMedia.filename"
      :size="selectedMedia.size"
      @saved="handleWorkbenchSaved"
    />

    <MediaDeleteModal
      v-model:visible="deleteVisible"
      :media="deleteTarget"
      @success="handleDeleteSuccess"
    />

    <MoveMediaModal
      v-model:visible="moveVisible"
      :media-ids="moveMediaIds"
      :folder="moveFolderTarget"
      @success="handleMoveSuccess"
    />

    <a-modal
      v-model:visible="folderModalVisible"
      :title="folderModalMode === 'create' ? '新建文件夹' : '重命名文件夹'"
      :confirm-loading="folderSubmitting"
      @ok="handleFolderModalOk"
    >
      <a-form layout="vertical">
        <a-form-item label="文件夹名称" required>
          <a-input
            v-model:value="folderModalName"
            placeholder="请输入文件夹名称"
            :maxlength="50"
            @press-enter="handleFolderModalOk"
          />
        </a-form-item>
        <p class="folder-modal-hint">创建于「{{ folderModalParentName }}」内；名称不能包含 / \ : * ? " &lt; &gt; |</p>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { message, Modal, type TableColumnsType } from 'ant-design-vue'
import {
  PictureOutlined,
  VideoCameraOutlined,
  UploadOutlined,
  AppstoreOutlined,
  UnorderedListOutlined,
  PlayCircleOutlined,
  SettingOutlined,
  DeleteOutlined,
  FolderFilled,
  FolderAddOutlined,
  FolderOpenOutlined,
  EditOutlined,
  MoreOutlined,
} from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaItem, MediaFolderNode } from '@/api/media'
import { getThumbUrl } from '@/utils/image'
import { IMAGE_ACCEPT, VIDEO_ACCEPT, MAX_UPLOAD_SIZE } from '@/utils/upload'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'
import { useMediaFolders } from '@/composables/useMediaFolders'
import MediaFolderBreadcrumbs from '@/components/media/MediaFolderBreadcrumbs.vue'
import MediaPresetWorkbench from '@/components/media/MediaPresetWorkbench.vue'
import MediaDeleteModal from '@/components/media/MediaDeleteModal.vue'
import MoveMediaModal from '@/components/media/MoveMediaModal.vue'

const activeTab = ref<'1' | '2'>('1')
const fileType = computed(() => Number(activeTab.value))

const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const viewType = ref<'grid' | 'list'>('grid')
const list = ref<MediaItem[]>([])
const loading = ref(false)
const uploading = ref(false)

const currentFolderId = ref('root')
const {
  folderTree,
  folderLoading,
  fetchFolderTree,
  findFolderPath,
  findFolder,
  collectSubtreeIds,
} = useMediaFolders()

const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(20)

const workbenchVisible = ref(false)
const selectedMedia = ref<MediaItem | null>(null)

const deleteVisible = ref(false)
const deleteTarget = ref<MediaItem | null>(null)

const moveVisible = ref(false)
const moveMediaIds = ref<string[]>([])
const moveFolderTarget = ref<MediaFolderNode | null>(null)

const folderModalVisible = ref(false)
const folderModalMode = ref<'create' | 'rename'>('create')
const folderModalName = ref('')
const folderModalTarget = ref<MediaFolderNode | null>(null)
const folderSubmitting = ref(false)

const searching = computed(() => !!debouncedKeyword.value)

/** 系统文件夹（模块播种，module_key 非空）：名称参与存储路径与文件 URL 链路，禁止重命名/移动/删除 */
const isSystemFolder = (folder: MediaFolderNode) => !!folder.module_key

/** 当前路径（面包屑），根目录为空数组 */
const folderPath = computed(() =>
  currentFolderId.value === 'root' ? [] : findFolderPath(folderTree.value, currentFolderId.value)
)

/** 当前文件夹的直接子文件夹 */
const visibleFolders = computed<MediaFolderNode[]>(() => {
  if (searching.value) return []
  if (currentFolderId.value === 'root') return folderTree.value
  const folder = findFolder(folderTree.value, currentFolderId.value)
  return folder?.children || []
})

/** 当前文件夹的父文件夹名（新建/重命名弹窗提示用） */
const folderModalParentName = computed(() => {
  if (folderModalMode.value === 'rename' && folderModalTarget.value) {
    const parent = folderModalTarget.value.parent_id
      ? findFolder(folderTree.value, folderModalTarget.value.parent_id!)?.name
      : null
    return parent || '全部文件'
  }
  return currentFolderId.value === 'root'
    ? '全部文件'
    : findFolder(folderTree.value, currentFolderId.value)?.name || '全部文件'
})

type FolderRow = { rowType: 'folder'; folder: MediaFolderNode; key: string }
type MediaRow = { rowType: 'media'; media: MediaItem; key: string }
type TableRow = FolderRow | MediaRow

/** 文件视图数据源：文件夹行在前、媒体行在后 */
const tableRows = computed<TableRow[]>(() => {
  const rows: TableRow[] = visibleFolders.value.map((f) => ({
    rowType: 'folder',
    folder: f,
    key: `folder-${f.id}`,
  }))
  for (const m of list.value) {
    rows.push({ rowType: 'media', media: m, key: m.id })
  }
  return rows
})

const columns: TableColumnsType = [
  { title: '名称', key: 'name', ellipsis: true },
  { title: '类型', key: 'type', width: 100 },
  { title: '尺寸', key: 'dimensions', width: 120 },
  { title: '大小', key: 'size', width: 120 },
  { title: '上传时间', key: 'created_at', width: 180 },
  { title: '操作', key: 'actions', width: 180 },
]

watch(keyword, setDebounce)

watch(debouncedKeyword, () => {
  reset()
  fetchList()
})

onMounted(() => {
  fetchFolderTree()
  fetchList()
})

async function fetchList() {
  loading.value = true
  try {
    // 搜索时全局检索（不传 folder_id）；浏览时限定当前文件夹
    const res = await mediaApi.getList({
      file_type: fileType.value,
      keyword: debouncedKeyword.value || undefined,
      folder_id: debouncedKeyword.value
        ? undefined
        : currentFolderId.value === 'root'
          ? 'root'
          : currentFolderId.value,
      page: page.value,
      page_size: pageSize.value,
    })
    list.value = res.list || []
    total.value = res.total || 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

function enterFolder(folderId: string | null) {
  currentFolderId.value = folderId || 'root'
  reset()
  fetchList()
}

function handleTabChange(key: string | number) {
  activeTab.value = String(key) as '1' | '2'
  reset()
  fetchList()
}

function handlePageChange(newPage: number, newPageSize?: number) {
  changePagination(newPage, newPageSize)
  fetchList()
}

function refreshAll() {
  fetchFolderTree()
  fetchList()
}

function handleBeforeUpload(file: File) {
  if (file.size > MAX_UPLOAD_SIZE) {
    message.error('文件大小不能超过 100MB')
    return false
  }
  return true
}

function handleCustomUpload({ file, onSuccess, onError }: any) {
  uploading.value = true
  mediaApi
    .upload(file as File, {
      folder_id: currentFolderId.value === 'root' ? undefined : currentFolderId.value,
      onProgress: (percent) => {
        message.loading({ content: `上传中 ${percent}%`, key: 'upload', duration: 0 })
      },
    })
    .then(() => {
      message.success({ content: '上传成功', key: 'upload' })
      onSuccess?.()
      refreshAll()
    })
    .catch((err) => {
      message.error({ content: err?.message || '上传失败', key: 'upload' })
      onError?.(err)
    })
    .finally(() => {
      uploading.value = false
    })
}

function handleFolderMenu(key: string, folder: MediaFolderNode) {
  // 兜底拦截（防陈旧树数据/绕过禁用态）：系统文件夹三种操作一律拒绝
  if (isSystemFolder(folder)) {
    message.warning('系统文件夹由模块固定使用，不支持重命名、移动或删除')
    return
  }
  if (key === 'rename') {
    openRenameFolder(folder)
  } else if (key === 'move') {
    moveFolderTarget.value = folder
    moveMediaIds.value = []
    moveVisible.value = true
  } else if (key === 'delete') {
    confirmDeleteFolder(folder)
  }
}

function openCreateFolder() {
  folderModalMode.value = 'create'
  folderModalName.value = ''
  folderModalTarget.value = null
  folderModalVisible.value = true
}

function openRenameFolder(folder: MediaFolderNode) {
  folderModalMode.value = 'rename'
  folderModalName.value = folder.name
  folderModalTarget.value = folder
  folderModalVisible.value = true
}

async function handleFolderModalOk() {
  const name = folderModalName.value.trim()
  if (!name) {
    message.warning('请输入文件夹名称')
    return
  }
  folderSubmitting.value = true
  try {
    if (folderModalMode.value === 'create') {
      const parentId = currentFolderId.value === 'root' ? null : currentFolderId.value
      await mediaApi.createFolder({ name, parent_id: parentId })
      message.success('文件夹已创建')
    } else if (folderModalTarget.value) {
      await mediaApi.updateFolder(folderModalTarget.value.id, { name })
      message.success('重命名成功')
    }
    folderModalVisible.value = false
    fetchFolderTree()
  } catch (err: any) {
    message.error(err?.message || '操作失败')
  } finally {
    folderSubmitting.value = false
  }
}

function confirmDeleteFolder(folder: MediaFolderNode) {
  const subtree = collectSubtreeIds(folder).length
  Modal.confirm({
    title: '删除文件夹',
    content:
      subtree > 1
        ? `「${folder.name}」包含子文件夹，仅允许删除空文件夹，确定继续？`
        : `确定删除文件夹「${folder.name}」？仅允许删除空文件夹。`,
    okText: '删除',
    okType: 'danger',
    onOk: async () => {
      try {
        await mediaApi.deleteFolder(folder.id)
        message.success('文件夹已删除')
        // 若删除的是当前所在文件夹（理论上不可发生，兜底回到根目录）
        if (currentFolderId.value === folder.id) enterFolder(null)
        else fetchFolderTree()
      } catch (err: any) {
        message.error(err?.message || '删除失败')
      }
    },
  })
}

function openMoveForMedia(item: MediaItem) {
  moveMediaIds.value = [item.id]
  moveFolderTarget.value = null
  moveVisible.value = true
}

function handleMoveSuccess() {
  refreshAll()
}

function openWorkbench(item: MediaItem) {
  selectedMedia.value = item
  workbenchVisible.value = true
}

function handleWorkbenchSaved() {
  fetchList()
}

function openDelete(item: MediaItem) {
  deleteTarget.value = item
  deleteVisible.value = true
}

function handleDeleteSuccess() {
  refreshAll()
}

function formatBytes(bytes?: number): string {
  if (bytes == null || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

function formatDateTime(time?: string): string {
  if (!time) return '-'
  const d = new Date(time)
  if (Number.isNaN(d.getTime())) return time
  return d.toLocaleString('zh-CN')
}
</script>

<style scoped>
.folder-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 4px;
}

.media-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin: 16px 0 20px;
  flex-wrap: wrap;
}

/* 文件夹网格（图标视图） */
.folder-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.folder-card {
  position: relative;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #f0f0f0);
  border-radius: 8px;
  padding: 16px 12px 10px;
  text-align: center;
  cursor: pointer;
  transition: box-shadow 0.2s, border-color 0.2s;
}

.folder-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-color: var(--primary, #4a6cf7);
}

.folder-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.folder-icon {
  font-size: 52px;
  color: #faad14;
  filter: drop-shadow(0 2px 3px rgba(0, 0, 0, 0.15));
}

.folder-count {
  font-size: 12px;
  color: var(--text-tertiary, #bfbfbf);
}

.folder-name {
  margin-top: 6px;
  font-size: 13px;
  color: var(--text-primary, #262626);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-width: 0;
}

.folder-name-text {
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.folder-system-tag {
  flex-shrink: 0;
  margin-right: 0;
  margin-inline-end: 0;
  font-size: 10px;
  line-height: 16px;
  padding: 0 4px;
}

.folder-more {
  position: absolute;
  top: 4px;
  right: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.folder-card:hover .folder-more {
  opacity: 1;
}

.folder-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--text-primary, #262626);
  cursor: pointer;
}

.folder-cell:hover {
  color: var(--primary, #4a6cf7);
}

.folder-cell-icon {
  font-size: 18px;
  color: #faad14;
}

.media-cell {
  color: var(--text-primary, #262626);
}

/* 媒体网格（图标视图） */
.media-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
}

.media-grid.no-folder {
  margin-top: 0;
}

.media-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #f0f0f0);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  transition: box-shadow 0.2s;
  display: flex;
  flex-direction: column;
}

.media-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.media-card:hover .media-overlay {
  opacity: 1;
}

.media-content {
  position: relative;
  width: 100%;
  aspect-ratio: 1;
  overflow: hidden;
}

.media-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 8px;
  background: rgba(0, 0, 0, 0.4);
  opacity: 0;
  transition: opacity 0.2s;
}

.media-content img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-card {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary, #f5f5f5);
  color: var(--text-secondary, #666);
  padding: 16px;
}

.video-icon {
  font-size: 48px;
}

.media-name {
  font-size: 12px;
  line-height: 1.4;
  padding: 6px 8px;
  text-align: center;
  color: var(--text-secondary, #595959);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.media-folder-tag {
  color: var(--primary, #4a6cf7);
  margin-left: 4px;
}

.folder-modal-hint {
  font-size: 12px;
  color: var(--text-tertiary, #bfbfbf);
  margin: 0;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}
</style>
