<template>
  <div class="media-browser">
    <a-input-search
      v-model:value="keyword"
      :placeholder="$t('media.picker.searchPlaceholder')"
      allow-clear
      class="browser-search"
    />

    <div class="browser-body">
      <!-- 左侧：文件夹树 -->
      <div class="browser-tree-panel">
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
      <div class="browser-content">
        <div v-if="searching" class="browser-search-hint">
          {{ $t('media.picker.searchResultFor', { keyword: debouncedKeyword.trim() }) }}
        </div>
        <div v-else class="browser-breadcrumb">
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
        <div v-if="searching && matchedFolders.length > 0" class="browser-folder-results">
          <div class="browser-section-label">{{ $t('media.picker.folderResults') }}</div>
          <div
            v-for="match in matchedFolders"
            :key="match.folder.id"
            class="browser-folder-result"
            @click="enterFolder(match.folder.id)"
          >
            <FolderFilled class="browser-folder-icon" />
            <span
              class="browser-folder-result-name"
              :title="folderDisplayName(match.folder.name, match.folder.module_key)"
            >
              {{ folderDisplayName(match.folder.name, match.folder.module_key) }}
            </span>
            <span class="browser-folder-result-path" :title="match.pathText">{{ match.pathText }}</span>
          </div>
        </div>

        <a-spin :spinning="loading">
          <div v-if="mediaList.length > 0" class="media-grid">
            <div
              v-for="item in mediaList"
              :key="item.id"
              class="media-item"
              :class="{ selected: selectedId === item.id }"
              role="button"
              tabindex="0"
              :aria-label="$t('media.picker.selectAria', { name: item.filename })"
              :aria-pressed="selectedId === item.id"
              @click="handleSelect(item)"
              @keydown.enter="handleSelect(item)"
            >
              <div class="media-thumb">
                <img
                  v-if="item.file_type === 1"
                  :src="item.thumb_url || getThumbUrl(item.url, 300)"
                  :alt="item.filename"
                  loading="lazy"
                />
                <div v-else class="video-thumb"><PlayCircleOutlined /></div>
                <div v-if="selectedId === item.id" class="selected-overlay">
                  <CheckCircleFilled />
                </div>
              </div>
              <div class="media-name" :title="item.filename">
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { TreeProps } from 'ant-design-vue'
import { CheckCircleFilled, FolderFilled, FolderOutlined, PlayCircleOutlined } from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaFolderNode, MediaItem } from '@/api/media'
import { getThumbUrl } from '@/utils/image'
import { useMediaFolders, folderDisplayName } from '@/composables/useMediaFolders'
import { useDebounce } from '@/composables/useDebounce'

const PAGE_SIZE = 20
/** 文件夹树根节点哨兵 key */
const ROOT_KEY = '__root__'

const props = defineProps<{
  /** 业务模块 key：挂载时自动定位到对应模块文件夹（如 recipe/game/book） */
  module?: string
  /** 按文件类型过滤（1=图片 2=视频），不传=全部 */
  fileType?: number
  /** 当前选中媒体 ID（受控），高亮由父组件驱动 */
  selectedId?: string | null
}>()

const emit = defineEmits<{
  (e: 'select', item: MediaItem): void
}>()

const { t } = useI18n()
const { folderTree, folderLoading, fetchFolderTree, findFolder, findFolderPath, findModuleFolder } = useMediaFolders()

const mediaList = ref<MediaItem[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)
const currentFolderId = ref<string | null>(null)

// 搜索：输入防抖后进入搜索态——文件走全局 keyword 检索，文件夹由客户端过滤目录树
const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

watch(keyword, (val) => setDebounce(val))
watch(debouncedKeyword, () => {
  page.value = 1
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

function handleSelect(item: MediaItem) {
  emit('select', item)
}

/** 同步树的选中项与展开路径（进入某文件夹时调用） */
function syncTreeSelection() {
  treeSelectedKeys.value = [currentFolderId.value || ROOT_KEY]
  const path = currentFolderId.value ? findFolderPath(folderTree.value, currentFolderId.value) : []
  const ids = new Set(expandedKeys.value)
  ids.add(ROOT_KEY)
  for (const folder of path) ids.add(folder.id)
  expandedKeys.value = [...ids]
}

function enterFolder(folderId: string | null) {
  if (currentFolderId.value === folderId && !searching.value) return
  clearSearch()
  currentFolderId.value = folderId
  page.value = 1
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

function clearSearch() {
  keyword.value = ''
  debouncedKeyword.value = ''
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
      file_type: props.fileType,
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

/** 挂载：加载目录树并按业务模块定位，随后拉取当前文件夹媒体 */
onMounted(async () => {
  expandedKeys.value = [ROOT_KEY]
  await fetchFolderTree()
  const moduleFolder = props.module ? findModuleFolder(folderTree.value, props.module) : null
  currentFolderId.value = moduleFolder?.id || null
  syncTreeSelection()
  fetchMedia()
})

/** 重新拉取当前视图（上传/删除后由父组件调用） */
function refresh() {
  fetchMedia()
}

defineExpose({ refresh })
</script>

<style scoped>
.browser-search {
  margin-bottom: 12px;
}

.browser-body {
  display: flex;
  gap: 16px;
}

.browser-tree-panel {
  width: 200px;
  flex-shrink: 0;
  max-height: 480px;
  overflow-y: auto;
  padding-right: 4px;
  border-right: 1px solid var(--border-color, #E5E9F2);
}

.browser-tree-panel :deep(.ant-tree) {
  background: transparent;
}

.browser-tree-panel :deep(.ant-tree-node-content-wrapper) {
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

.browser-content {
  flex: 1;
  min-width: 0;
}

.browser-breadcrumb {
  margin-bottom: 12px;
}

.browser-search-hint {
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--text-secondary, #667085);
}

.browser-folder-results {
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px dashed var(--border-color, #E5E9F2);
}

.browser-section-label {
  font-size: 12px;
  color: var(--text-tertiary, #8A93A8);
  margin-bottom: 6px;
}

.browser-folder-result {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.browser-folder-result:hover {
  background: var(--bg-secondary, #F5F7FC);
}

.browser-folder-icon {
  color: #B7791F;
  font-size: 16px;
  flex-shrink: 0;
}

.browser-folder-result-name {
  max-width: 45%;
  min-width: 0;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.browser-folder-result-path {
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

.media-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}

.media-item {
  padding: 4px;
  border: 2px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.2s;
}

.media-item:hover { border-color: var(--primary, #526FE8); }
.media-item.selected { border-color: var(--primary, #526FE8); }

.media-item:focus-visible {
  outline: 2px solid var(--primary, #526FE8);
  outline-offset: 2px;
}

.media-thumb {
  position: relative;
  aspect-ratio: 1;
  border-radius: 6px;
  overflow: hidden;
}

.media-thumb img {
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

.media-name {
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
</style>
