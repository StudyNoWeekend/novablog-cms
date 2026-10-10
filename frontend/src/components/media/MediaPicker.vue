<template>
  <a-modal
    v-model:open="open"
    title="选择图片"
    width="680px"
    :footer="null"
    @cancel="handleCancel"
  >
    <div class="picker-breadcrumb">
      <a-breadcrumb>
        <a-breadcrumb-item>
          <a @click.prevent="enterFolder(null)">全部文件</a>
        </a-breadcrumb-item>
        <a-breadcrumb-item v-for="folder in currentPath" :key="folder.id">
          <a @click.prevent="enterFolder(folder.id)">{{ folder.name }}</a>
        </a-breadcrumb-item>
      </a-breadcrumb>
      <span v-if="currentFolder" class="picker-folder-name">{{ currentFolder.name }}</span>
    </div>

    <!-- 当前文件夹的子文件夹 -->
    <div v-if="subFolders.length > 0" class="picker-folders">
      <div
        v-for="folder in subFolders"
        :key="folder.id"
        class="picker-folder-item"
        @click="enterFolder(folder.id)"
      >
        <FolderFilled class="picker-folder-icon" />
        <span class="picker-folder-title" :title="folder.name">{{ folder.name }}</span>
      </div>
    </div>

    <div class="media-picker-grid">
      <div
        v-for="item in mediaList"
        :key="item.id"
        class="media-picker-item"
        :class="{ selected: selectedId === item.id }"
        @click="selectedId = item.id"
      >
        <img :src="item.thumb_url || getThumbUrl(item.url, 300)" :alt="item.filename" />
        <div v-if="selectedId === item.id" class="selected-overlay">
          <CheckCircleFilled />
        </div>
      </div>
    </div>
    <a-empty v-if="!loading && subFolders.length === 0 && mediaList.length === 0" description="此文件夹为空" />
    <a-pagination
      v-if="total > 20"
      :current="page"
      :page-size="20"
      :total="total"
      @change="handlePageChange"
      style="margin-top: 16px; text-align: center;"
    />

    <div v-if="selectedId" class="media-picker-footer">
      <a-button type="primary" @click="handleConfirm">确认选择</a-button>
    </div>

    <MediaPresetChooser
      v-model:visible="chooserVisible"
      :media-id="pendingItem?.id || null"
      :original-url="pendingItem?.url || ''"
      @confirm="handleChooserConfirm"
    />
  </a-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { CheckCircleFilled, FolderFilled } from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaItem } from '@/api/media'
import { getThumbUrl } from '@/utils/image'
import { useMediaFolders } from '@/composables/useMediaFolders'
import MediaPresetChooser from './MediaPresetChooser.vue'

/** 选图结果：选中预设时 url 为成品图 URL，presetId/presetName 标记来源预设 */
export interface MediaPickerSelected {
  id: string
  url: string
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

const { folderTree, fetchFolderTree, findFolderPath, findFolder, findModuleFolder } = useMediaFolders()

const open = ref(props.visible)
const mediaList = ref<MediaItem[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)
const selectedId = ref<string | null>(null)
const currentFolderId = ref<string | null>(null)
const chooserVisible = ref(false)
const pendingItem = ref<MediaItem | null>(null)

const currentFolder = computed(() =>
  currentFolderId.value ? findFolder(folderTree.value, currentFolderId.value) : null
)

const currentPath = computed(() =>
  currentFolderId.value ? findFolderPath(folderTree.value, currentFolderId.value) : []
)

const subFolders = computed(() => {
  if (!currentFolderId.value) return folderTree.value
  return currentFolder.value?.children || []
})

watch(() => props.visible, async (v) => {
  open.value = v
  if (v) {
    page.value = 1
    selectedId.value = null
    await fetchFolderTree()
    // 优先定位到业务模块对应的文件夹
    const moduleFolder = props.module ? findModuleFolder(folderTree.value, props.module) : null
    currentFolderId.value = moduleFolder?.id || null
    fetchMedia()
  }
})

watch(open, (v) => {
  if (!v) {
    emit('update:visible', false)
  }
})

function enterFolder(folderId: string | null) {
  if (currentFolderId.value === folderId) return
  currentFolderId.value = folderId
  page.value = 1
  selectedId.value = null
  fetchMedia()
}

async function fetchMedia() {
  loading.value = true
  try {
    const res = await mediaApi.getList({
      page: page.value,
      page_size: 20,
      folder_id: currentFolderId.value || 'root',
    })
    mediaList.value = res.list || []
    total.value = res.total || 0
  } catch {}
  finally { loading.value = false }
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
    emit('selected', { id: item.id, url: item.url })
    open.value = false
    return
  }
  // 图片：弹出版本选择层（层内可点图片直接打开预设工作台制作/编辑）
  pendingItem.value = item
  chooserVisible.value = true
}

function handleChooserConfirm(payload: { url: string; presetId?: string; presetName?: string }) {
  chooserVisible.value = false
  const item = pendingItem.value
  if (item) {
    emit('selected', {
      id: item.id,
      url: payload.url,
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
.picker-breadcrumb {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.picker-folder-name {
  font-size: 13px;
  color: var(--text-secondary, #667085);
}

.picker-folders {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 12px;
}

.picker-folder-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.2s;
}

.picker-folder-item:hover {
  border-color: var(--primary, #526FE8);
}

.picker-folder-icon {
  color: #B7791F;
  font-size: 18px;
  flex-shrink: 0;
}

.picker-folder-title {
  font-size: 13px;
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
  aspect-ratio: 1;
  position: relative;
  cursor: pointer;
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid transparent;
  transition: border-color 0.2s;
}
.media-picker-item:hover { border-color: var(--primary, #526FE8); }
.media-picker-item.selected { border-color: var(--primary, #526FE8); }
.media-picker-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
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
.media-picker-footer {
  margin-top: 16px;
  text-align: right;
}
</style>
