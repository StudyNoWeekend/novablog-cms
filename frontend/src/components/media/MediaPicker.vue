<template>
  <a-modal
    v-model:open="open"
    :title="$t('media.picker.title')"
    width="920px"
    :footer="null"
    @cancel="handleCancel"
  >
    <a-input-search
      v-model:value="keyword"
      :placeholder="$t('media.picker.searchPlaceholder')"
      allow-clear
      class="picker-search"
    />

    <div class="picker-body">
      <!-- 左侧：文件夹树 -->
      <div class="picker-tree-panel">
        <a-spin :spinning="folderLoading">
          <a-tree
            v-if="treeData.length > 0"
            :tree-data="treeData"
            :selected-keys="treeSelectedKeys"
            :expanded-keys="expandedKeys"
            block-node
            @select="handleTreeSelect"
            @expand="handleTreeExpand"
          >
            <template #title="{ title, count }">
              <span class="tree-folder">
                <FolderOutlined />
                <span class="tree-title" :title="title">{{ title }}</span>
                <span v-if="count > 0" class="tree-count">{{ $t('media.itemCount', { count }) }}</span>
              </span>
            </template>
          </a-tree>
          <a-empty v-else-if="!folderLoading" :description="$t('media.folder.empty')" />
        </a-spin>
      </div>

      <!-- 右侧：面包屑/搜索提示 + 文件区 -->
      <div class="picker-content">
        <div v-if="searching" class="picker-search-hint">
          {{ $t('media.picker.searchResultFor', { keyword: debouncedKeyword.trim() }) }}
        </div>
        <div v-else class="picker-breadcrumb">
          <a-breadcrumb>
            <a-breadcrumb-item>
              <a @click.prevent="enterFolder(null)">{{ $t('media.folder.rootName') }}</a>
            </a-breadcrumb-item>
            <a-breadcrumb-item v-for="folder in currentPath" :key="folder.id">
              <a @click.prevent="enterFolder(folder.id)">{{ folderDisplayName(folder.name, folder.module_key) }}</a>
            </a-breadcrumb-item>
          </a-breadcrumb>
        </div>

        <!-- 搜索态：匹配的文件夹（客户端过滤目录树） -->
        <div v-if="searching && matchedFolders.length > 0" class="picker-folder-results">
          <div class="picker-section-label">{{ $t('media.picker.folderResults') }}</div>
          <div
            v-for="match in matchedFolders"
            :key="match.folder.id"
            class="picker-folder-result"
            @click="enterFolder(match.folder.id)"
          >
            <FolderFilled class="picker-folder-icon" />
            <span
              class="picker-folder-result-name"
              :title="folderDisplayName(match.folder.name, match.folder.module_key)"
            >
              {{ folderDisplayName(match.folder.name, match.folder.module_key) }}
            </span>
            <span class="picker-folder-result-path" :title="match.pathText">{{ match.pathText }}</span>
          </div>
        </div>

        <a-spin :spinning="loading">
          <div v-if="mediaList.length > 0" class="media-picker-grid">
            <div
              v-for="item in mediaList"
              :key="item.id"
              class="media-picker-item"
              :class="{ selected: selectedId === item.id }"
              @click="selectedId = item.id"
            >
              <div class="media-picker-thumb">
                <img
                  v-if="item.file_type === 1"
                  :src="item.thumb_url || getThumbUrl(item.url, 300)"
                  :alt="item.filename"
                />
                <div v-else class="video-thumb"><PlayCircleOutlined /></div>
                <div v-if="selectedId === item.id" class="selected-overlay">
                  <CheckCircleFilled />
                </div>
              </div>
              <div class="media-picker-name" :title="item.filename">
                {{ item.filename }}
                <span v-if="searching && item.folder_name" class="media-folder-tag">
                  {{ folderDisplayName(item.folder_name, item.folder_module_key) }}
                </span>
              </div>
            </div>
          </div>
          <a-empty
            v-else-if="!loading && (searching ? matchedFolders.length === 0 : true)"
            :description="emptyText"
            style="margin: 48px 0"
          />
        </a-spin>

        <a-pagination
          v-if="total > PAGE_SIZE"
          :current="page"
          :page-size="PAGE_SIZE"
          :total="total"
          @change="handlePageChange"
          style="margin-top: 16px; text-align: center;"
        />
      </div>
    </div>

    <div v-if="selectedId" class="media-picker-footer">
      <a-button type="primary" @click="handleConfirm">{{ $t('media.picker.confirm') }}</a-button>
    </div>

    <MediaPresetChooser
      v-model:visible="chooserVisible"
      :media-id="pendingItem?.id || null"
      :original-url="pendingItem?.url || ''"
      :original-file-url="pendingItem?.file_url || ''"
      @confirm="handleChooserConfirm"
    />
  </a-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { TreeProps } from 'ant-design-vue'
import { CheckCircleFilled, FolderFilled, FolderOutlined, PlayCircleOutlined } from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaFolderNode, MediaItem } from '@/api/media'
import { getThumbUrl } from '@/utils/image'
import { useMediaFolders, folderDisplayName } from '@/composables/useMediaFolders'
import { useDebounce } from '@/composables/useDebounce'
import MediaPresetChooser from './MediaPresetChooser.vue'

const PAGE_SIZE = 20
/** 文件夹树根节点哨兵 key */
const ROOT_KEY = '__root__'

/** 选图结果：选中预设时 url 为成品图 URL，presetId/presetName 标记来源预设；
 * fileUrl 为按媒体 ID 索引的稳定链接（jsonb 内嵌等场景用，选中预设时带 preset 参数） */
export interface MediaPickerSelected {
  id: string
  url: string
  fileUrl?: string
  presetId?: string
  presetName?: string
}

const props = withDefaults(
  defineProps<{
    visible: boolean
    /** 业务模块 key：打开时自动定位到对应模块文件夹（如 recipe/game/book） */
    module?: string
    /** 是否启用图片预设工作台/版本选择：确认时弹出（默认开启） */
    preset?: boolean
  }>(),
  { preset: true }
)

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'selected', media: MediaPickerSelected): void
}>()

const { t } = useI18n()
const { folderTree, folderLoading, fetchFolderTree, findFolder, findFolderPath, findModuleFolder } = useMediaFolders()

const open = ref(props.visible)
const mediaList = ref<MediaItem[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)
const selectedId = ref<string | null>(null)
const currentFolderId = ref<string | null>(null)
const chooserVisible = ref(false)
const pendingItem = ref<MediaItem | null>(null)

// 搜索：输入防抖后进入搜索态——文件走全局 keyword 检索，文件夹由客户端过滤目录树
const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

watch(keyword, (val) => setDebounce(val))
watch(debouncedKeyword, () => {
  page.value = 1
  selectedId.value = null
  fetchMedia()
})

const searching = computed(() => debouncedKeyword.value.trim().length > 0)

// 文件夹树（受控选中/展开），treeData 与树面板写法对齐 MoveMediaModal
const treeSelectedKeys = ref<string[]>([])
const expandedKeys = ref<string[]>([])

interface TreeItem {
  title: string
  key: string
  count: number
  children?: TreeItem[]
}

const treeData = computed<TreeItem[]>(() => {
  const build = (nodes: MediaFolderNode[]): TreeItem[] =>
    nodes.map((n) => ({
      title: folderDisplayName(n.name, n.module_key),
      key: n.id,
      count: n.media_count,
      children: build(n.children || []),
    }))
  return [{ title: t('media.folder.rootName'), key: ROOT_KEY, count: 0, children: build(folderTree.value) }]
})

const currentFolder = computed(() =>
  currentFolderId.value ? findFolder(folderTree.value, currentFolderId.value) : null
)

const currentPath = computed(() =>
  currentFolderId.value ? findFolderPath(folderTree.value, currentFolderId.value) : []
)

interface MatchedFolder {
  folder: MediaFolderNode
  /** 从根目录到该文件夹的路径文本（不含自身） */
  pathText: string
}

const matchedFolders = computed<MatchedFolder[]>(() => {
  const kw = debouncedKeyword.value.trim().toLowerCase()
  if (!kw) return []
  const results: MatchedFolder[] = []
  const walk = (nodes: MediaFolderNode[], ancestors: MediaFolderNode[]) => {
    for (const node of nodes) {
      if (folderDisplayName(node.name, node.module_key).toLowerCase().includes(kw)) {
        const pathText = [
          t('media.folder.rootName'),
          ...ancestors.map((f) => folderDisplayName(f.name, f.module_key)),
        ].join(' / ')
        results.push({ folder: node, pathText })
      }
      walk(node.children || [], [...ancestors, node])
    }
  }
  walk(folderTree.value, [])
  return results
})

const emptyText = computed(() => {
  if (searching.value) return t('media.picker.noResults')
  return currentFolderId.value ? t('media.empty.folder') : t('media.empty.root')
})

/** 同步树的选中项与展开路径（进入某文件夹时调用） */
function syncTreeSelection() {
  treeSelectedKeys.value = [currentFolderId.value || ROOT_KEY]
  const path = currentFolderId.value ? findFolderPath(folderTree.value, currentFolderId.value) : []
  const ids = new Set(expandedKeys.value)
  ids.add(ROOT_KEY)
  for (const folder of path) ids.add(folder.id)
  expandedKeys.value = [...ids]
}

watch(() => props.visible, async (v) => {
  open.value = v
  if (v) {
    page.value = 1
    selectedId.value = null
    keyword.value = ''
    debouncedKeyword.value = ''
    expandedKeys.value = [ROOT_KEY]
    await fetchFolderTree()
    // 优先定位到业务模块对应的文件夹
    const moduleFolder = props.module ? findModuleFolder(folderTree.value, props.module) : null
    currentFolderId.value = moduleFolder?.id || null
    syncTreeSelection()
    fetchMedia()
  }
})

watch(open, (v) => {
  if (!v) {
    emit('update:visible', false)
  }
})

function clearSearch() {
  keyword.value = ''
  debouncedKeyword.value = ''
}

function enterFolder(folderId: string | null) {
  if (currentFolderId.value === folderId && !searching.value) return
  clearSearch()
  currentFolderId.value = folderId
  page.value = 1
  selectedId.value = null
  syncTreeSelection()
  fetchMedia()
}

function handleTreeSelect(keys: (string | number)[]) {
  if (keys.length === 0) return // 点击已选中节点会触发反选，忽略
  const key = String(keys[0])
  enterFolder(key === ROOT_KEY ? null : key)
}

const handleTreeExpand: TreeProps['onExpand'] = (keys) => {
  expandedKeys.value = keys.map(String)
}

/** seq 自增守卫：仅应用最后一次请求的响应，避免快速切换目录/搜索时旧响应覆盖新结果 */
let fetchSeq = 0

async function fetchMedia() {
  const seq = ++fetchSeq
  loading.value = true
  try {
    const res = await mediaApi.getList({
      page: page.value,
      page_size: PAGE_SIZE,
      // 搜索时不传 folder_id = 全局检索；浏览时限定当前文件夹
      ...(searching.value
        ? { keyword: debouncedKeyword.value.trim() }
        : { folder_id: currentFolderId.value || 'root' }),
    })
    if (seq !== fetchSeq) return
    mediaList.value = res.list || []
    total.value = res.total || 0
  } catch {}
  finally {
    if (seq === fetchSeq) loading.value = false
  }
}

function handlePageChange(p: number) {
  page.value = p
  fetchMedia()
}

function handleConfirm() {
  const item = mediaList.value.find(m => m.id === selectedId.value)
  if (!item) return
  // 非图片（视频等）或关闭预设能力时：保持原行为直接回传
  if (!props.preset || item.file_type !== 1) {
    emit('selected', { id: item.id, url: item.url, fileUrl: item.file_url })
    open.value = false
    return
  }
  // 图片：弹出版本选择层（层内可点图片直接打开预设工作台制作/编辑）
  pendingItem.value = item
  chooserVisible.value = true
}

function handleChooserConfirm(payload: { url: string; fileUrl?: string; presetId?: string; presetName?: string }) {
  chooserVisible.value = false
  const item = pendingItem.value
  if (item) {
    emit('selected', {
      id: item.id,
      url: payload.url,
      fileUrl: payload.fileUrl || item.file_url,
      presetId: payload.presetId,
      presetName: payload.presetName,
    })
  }
  pendingItem.value = null
  open.value = false
}

function handleCancel() {
  open.value = false
}
</script>

<style scoped>
.picker-search {
  margin-bottom: 12px;
}

.picker-body {
  display: flex;
  gap: 16px;
}

.picker-tree-panel {
  width: 200px;
  flex-shrink: 0;
  max-height: 480px;
  overflow-y: auto;
  padding-right: 4px;
  border-right: 1px solid var(--border-color, #E5E9F2);
}

.picker-tree-panel :deep(.ant-tree) {
  background: transparent;
}

.picker-tree-panel :deep(.ant-tree-node-content-wrapper) {
  min-width: 0;
}

.tree-folder {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  min-width: 0;
}

.tree-folder :deep(.anticon) {
  color: #B7791F;
}

.tree-title {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-count {
  flex-shrink: 0;
  color: var(--text-tertiary, #8A93A8);
  font-size: 12px;
}

.picker-content {
  flex: 1;
  min-width: 0;
}

.picker-breadcrumb {
  margin-bottom: 12px;
}

.picker-search-hint {
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--text-secondary, #667085);
}

.picker-folder-results {
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px dashed var(--border-color, #E5E9F2);
}

.picker-section-label {
  font-size: 12px;
  color: var(--text-tertiary, #8A93A8);
  margin-bottom: 6px;
}

.picker-folder-result {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.picker-folder-result:hover {
  background: var(--bg-secondary, #F5F7FC);
}

.picker-folder-icon {
  color: #B7791F;
  font-size: 16px;
  flex-shrink: 0;
}

.picker-folder-result-name {
  max-width: 45%;
  min-width: 0;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.picker-folder-result-path {
  flex: 1;
  min-width: 0;
  margin-left: auto;
  text-align: right;
  font-size: 12px;
  color: var(--text-tertiary, #8A93A8);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.media-picker-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}

.media-picker-item {
  padding: 4px;
  border: 2px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.2s;
}

.media-picker-item:hover { border-color: var(--primary, #526FE8); }
.media-picker-item.selected { border-color: var(--primary, #526FE8); }

.media-picker-thumb {
  position: relative;
  aspect-ratio: 1;
  border-radius: 6px;
  overflow: hidden;
}

.media-picker-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-thumb {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary, #F5F7FC);
  color: var(--text-secondary, #666);
  font-size: 28px;
}

.selected-overlay {
  position: absolute;
  inset: 0;
  background: rgba(82, 111, 232, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  color: var(--primary, #526FE8);
}

.media-picker-name {
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.4;
  text-align: center;
  color: var(--text-secondary, #667085);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.media-folder-tag {
  color: var(--primary, #526FE8);
  margin-left: 4px;
}

.media-picker-footer {
  margin-top: 16px;
  text-align: right;
}
</style>
