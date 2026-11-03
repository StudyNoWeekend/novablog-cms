<template>
  <a-modal
    v-model:open="open"
    :title="$t('travel.coverPicker.title')"
    width="920px"
    :footer="null"
    @cancel="handleClose"
  >
    <div class="cover-picker">
      <!-- 上传进度 -->
      <div v-if="uploading" class="upload-progress-card">
        <a-progress :percent="uploadPercent" size="small" status="active" />
      </div>

      <!-- 两栏媒体浏览器：文件夹树 + 文件区（含文件夹/文件名搜索） -->
      <MediaBrowser
        v-if="open"
        ref="browserRef"
        module="travel"
        :file-type="1"
        :selected-id="selectedId"
        @select="handleSelect"
      />

      <!-- 底部操作栏 -->
      <div class="picker-footer">
        <div class="selected-info">
          <template v-if="selectedId && selectedUrl">
            <span class="selected-label">{{ $t('travel.coverPicker.selected') }}</span>
            <img
              class="selected-thumb"
              :src="selectedUrl"
              :alt="$t('travel.coverPicker.selectedAlt')"
            />
          </template>
          <span v-else class="selected-label">{{ $t('travel.coverPicker.notSelected') }}</span>
        </div>

        <div class="footer-actions">
          <input
            ref="fileInputRef"
            type="file"
            accept="image/*"
            style="display: none"
            aria-hidden="true"
            @change="handleFileChange"
          />
          <a-button
            :loading="uploading"
            :disabled="uploading"
            :aria-label="$t('travel.coverPicker.uploadLocal')"
            @click="handleUploadClick"
          >
            <UploadOutlined /> {{ $t('travel.coverPicker.uploadLocal') }}
          </a-button>
          <a-button :aria-label="$t('common.actions.cancel')" @click="handleClose">
            {{ $t('common.actions.cancel') }}
          </a-button>
          <a-button
            type="primary"
            :disabled="!selectedId"
            :aria-label="$t('travel.coverPicker.confirm')"
            @click="handleConfirm"
          >
            {{ $t('travel.coverPicker.confirm') }}
          </a-button>
        </div>
      </div>

      <MediaPresetChooser
        v-model:visible="chooserVisible"
        :media-id="selectedId"
        :original-url="selectedUrl"
        :original-file-url="selectedFileUrl"
        @confirm="handleChooserConfirm"
      />
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import { UploadOutlined } from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaItem } from '@/api/media'
import MediaBrowser from '@/components/media/MediaBrowser.vue'
import MediaPresetChooser from '@/components/media/MediaPresetChooser.vue'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  // fileUrl 为按媒体 ID 索引的稳定链接（jsonb 内嵌等场景用；选中预设时带 preset 参数）
  (e: 'selected', payload: { mediaId: string; presetId?: string; url: string; fileUrl?: string }): void
}>()

const open = ref(false)
const selectedId = ref<string | null>(null)
const selectedItem = ref<MediaItem | null>(null)
const selectedUrl = ref('')
const selectedPresetId = ref<string | undefined>(undefined)
const selectedFileUrl = ref<string | undefined>(undefined)
const uploading = ref(false)
const uploadPercent = ref(0)
const fileInputRef = ref<HTMLInputElement | null>(null)
const chooserVisible = ref(false)
const browserRef = ref<InstanceType<typeof MediaBrowser> | null>(null)

watch(
  () => props.visible,
  (v) => {
    open.value = v
    if (v) {
      selectedId.value = null
      selectedItem.value = null
      selectedUrl.value = ''
      selectedPresetId.value = undefined
      selectedFileUrl.value = undefined
    }
  }
)

watch(open, (v) => {
  if (!v) {
    emit('update:visible', false)
  }
})

function handleSelect(item: MediaItem) {
  selectedId.value = item.id
  selectedItem.value = item
  selectedUrl.value = item.url
  selectedPresetId.value = undefined
  selectedFileUrl.value = undefined
}

function handleUploadClick() {
  if (uploading.value) return
  fileInputRef.value?.click()
}

async function handleFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  uploading.value = true
  uploadPercent.value = 0

  try {
    // 「默认无 EXIF」为后端预设约定的固定名称，参与后端匹配，不做翻译
    const res = await mediaApi.uploadWithPreset(
      file,
      '默认无 EXIF',
      {
        module: 'travel',
        onProgress: (percent) => {
          uploadPercent.value = percent
        },
      }
    )
    if (res.media && res.preset) {
      selectedId.value = res.media.id
      selectedItem.value = res.media
      selectedUrl.value = res.preset.output_url || res.media.url
      selectedPresetId.value = res.preset.id
      selectedFileUrl.value = res.preset.file_url || res.media.file_url
      browserRef.value?.refresh()
    }
  } catch {
    message.error(t('travel.coverPicker.uploadFailed'))
  } finally {
    uploading.value = false
    uploadPercent.value = 0
    if (fileInputRef.value) {
      fileInputRef.value.value = ''
    }
  }
}

function handleConfirm() {
  if (!selectedId.value || !selectedUrl.value) return
  // 上传流程已确定预设，直接回传；从列表选的图先经预设选择层（无预设时直接回传原图）
  if (selectedPresetId.value) {
    emitSelected()
    return
  }
  chooserVisible.value = true
}

function handleChooserConfirm(payload: { url: string; fileUrl?: string; presetId?: string; presetName?: string }) {
  chooserVisible.value = false
  selectedUrl.value = payload.url
  selectedPresetId.value = payload.presetId
  selectedFileUrl.value = payload.fileUrl
  emitSelected()
}

function emitSelected() {
  if (!selectedId.value || !selectedUrl.value) return
  emit('selected', {
    mediaId: selectedId.value,
    presetId: selectedPresetId.value,
    url: selectedUrl.value,
    fileUrl: selectedFileUrl.value,
  })
  open.value = false
}

function handleClose() {
  open.value = false
}
</script>

<style scoped>
.cover-picker {
  display: flex;
  flex-direction: column;
}

.upload-progress-card {
  margin-bottom: 12px;
  padding: 8px 12px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius);
  box-shadow: var(--shadow-sm);
}

.picker-footer {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.selected-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.selected-label {
  color: var(--text-secondary);
  font-size: 14px;
}

.selected-thumb {
  width: 40px;
  height: 40px;
  object-fit: cover;
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.footer-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
