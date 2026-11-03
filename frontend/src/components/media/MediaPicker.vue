<template>
  <a-modal
    v-model:open="open"
    :title="$t('media.picker.title')"
    width="920px"
    :footer="null"
    @cancel="handleCancel"
  >
    <MediaBrowser
      v-if="open"
      :module="module"
      :selected-id="selectedId"
      @select="handleSelect"
    />

    <div v-if="selectedId" class="media-picker-footer">
      <a-button type="primary" @click="handleConfirm">{{ $t('media.picker.confirm') }}</a-button>
    </div>

    <MediaPresetChooser
      v-model:visible="chooserVisible"
      :media-id="selectedItem?.id || null"
      :original-url="selectedItem?.url || ''"
      :original-file-url="selectedItem?.file_url || ''"
      @confirm="handleChooserConfirm"
    />
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { MediaItem } from '@/api/media'
import MediaBrowser from './MediaBrowser.vue'
import MediaPresetChooser from './MediaPresetChooser.vue'

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

const open = ref(props.visible)
const selectedId = ref<string | null>(null)
const selectedItem = ref<MediaItem | null>(null)
const chooserVisible = ref(false)

watch(() => props.visible, (v) => {
  open.value = v
  if (v) {
    selectedId.value = null
    selectedItem.value = null
  }
})

watch(open, (v) => {
  if (!v) {
    emit('update:visible', false)
  }
})

function handleSelect(item: MediaItem | null) {
  selectedId.value = item?.id ?? null
  selectedItem.value = item
}

function handleConfirm() {
  const item = selectedItem.value
  if (!item) return
  // 非图片（视频等）或关闭预设能力时：保持原行为直接回传
  if (!props.preset || item.file_type !== 1) {
    emit('selected', { id: item.id, url: item.url, fileUrl: item.file_url })
    open.value = false
    return
  }
  // 图片：弹出版本选择层（层内可点图片直接打开预设工作台制作/编辑）
  chooserVisible.value = true
}

function handleChooserConfirm(payload: { url: string; fileUrl?: string; presetId?: string; presetName?: string }) {
  chooserVisible.value = false
  const item = selectedItem.value
  if (item) {
    emit('selected', {
      id: item.id,
      url: payload.url,
      fileUrl: payload.fileUrl || item.file_url,
      presetId: payload.presetId,
      presetName: payload.presetName,
    })
  }
  open.value = false
}

function handleCancel() {
  open.value = false
}
</script>

<style scoped>
.media-picker-footer {
  margin-top: 16px;
  text-align: right;
}
</style>
