<template>
  <a-modal
    v-model:open="open"
    title="选择封面图"
    width="720px"
    :footer="null"
    @cancel="handleClose"
  >
    <div class="cover-picker">
      <!-- 顶部工具栏 -->
      <div class="picker-toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索图片文件名"
          allow-clear
          style="width: 260px"
        />
      </div>

      <!-- 上传进度 -->
      <div v-if="uploading" class="upload-progress-card">
        <a-progress :percent="uploadPercent" size="small" status="active" />
      </div>

      <!-- 图片网格 -->
      <a-spin :spinning="loading">
        <div v-if="mediaList.length > 0" class="media-grid">
          <div
            v-for="item in mediaList"
            :key="item.id"
            class="media-item"
            :class="{ selected: selectedId === item.id }"
            role="button"
            tabindex="0"
            :aria-label="`选择图片 ${item.filename}`"
            :aria-pressed="selectedId === item.id"
            @click="handleSelect(item)"
            @keydown.enter="handleSelect(item)"
          >
            <img
              :src="item.thumb_url || item.url"
              :alt="item.filename"
              loading="lazy"
            />
            <div v-if="selectedId === item.id" class="selected-overlay">
              <CheckCircleFilled />
            </div>
          </div>
        </div>
        <a-empty
          v-else-if="!loading"
          description="暂无图片"
          style="margin-top: 48px"
        />
      </a-spin>

      <!-- 分页 -->
      <a-pagination
        v-if="total > pageSize"
        :current="page"
        :page-size="pageSize"
        :total="total"
        class="pagination"
        @change="handlePageChange"
      />

      <!-- 底部操作栏 -->
      <div class="picker-footer">
        <div class="selected-info">
          <template v-if="selectedId && selectedUrl">
            <span class="selected-label">已选择</span>
            <img
              class="selected-thumb"
              :src="selectedUrl"
              alt="已选封面"
            />
          </template>
          <span v-else class="selected-label">未选择</span>
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
            aria-label="本地上传"
            @click="handleUploadClick"
          >
            <UploadOutlined /> 本地上传
          </a-button>
          <a-button aria-label="取消" @click="handleClose">
            取消
          </a-button>
          <a-button
            type="primary"
            :disabled="!selectedId"
            aria-label="确认选择"
            @click="handleConfirm"
          >
            确认选择
          </a-button>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import {
  PictureOutlined,
  CheckCircleFilled,
  UploadOutlined,
  CloseOutlined,
} from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaItem } from '@/api/media'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'selected', payload: { mediaId: string; presetId?: string; url: string }): void
}>()

const open = ref(false)
const mediaList = ref<MediaItem[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const keyword = ref('')
const selectedId = ref<string | null>(null)
const selectedUrl = ref('')
const selectedPresetId = ref<string | undefined>(undefined)
const uploading = ref(false)
const uploadPercent = ref(0)
const fileInputRef = ref<HTMLInputElement | null>(null)

let keywordDebounceTimer: ReturnType<typeof setTimeout> | null = null
let skipNextKeywordFetch = false

watch(
  () => props.visible,
  (v) => {
    open.value = v
    if (v) {
      page.value = 1
      skipNextKeywordFetch = true
      keyword.value = ''
      selectedId.value = null
      selectedUrl.value = ''
      selectedPresetId.value = undefined
      fetchMedia()
    }
  }
)

watch(open, (v) => {
  if (!v) {
    emit('update:visible', false)
  }
})

watch(keyword, () => {
  if (skipNextKeywordFetch) {
    skipNextKeywordFetch = false
    return
  }
  if (keywordDebounceTimer) {
    clearTimeout(keywordDebounceTimer)
  }
  keywordDebounceTimer = setTimeout(() => {
    page.value = 1
    fetchMedia()
  }, 300)
})

async function fetchMedia() {
  loading.value = true
  try {
    const res = await mediaApi.getList({
      file_type: 1,
      keyword: keyword.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    })
    mediaList.value = res.list || []
    total.value = res.total || 0
  } catch {
    mediaList.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handlePageChange(p: number) {
  page.value = p
  fetchMedia()
}

function handleSelect(item: MediaItem) {
  selectedId.value = item.id
  selectedUrl.value = item.url
  selectedPresetId.value = undefined
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
    const res = await mediaApi.uploadWithPreset(
      file,
      '默认无 EXIF',
      (percent) => {
        uploadPercent.value = percent
      }
    )
    if (res.media && res.preset) {
      selectedId.value = res.media.id
      selectedUrl.value = res.preset.output_url || res.media.url
      selectedPresetId.value = res.preset.id
      await fetchMedia()
    }
  } catch {
    message.error('上传失败，请重试')
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
  emit('selected', {
    mediaId: selectedId.value,
    presetId: selectedPresetId.value,
    url: selectedUrl.value,
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

.picker-toolbar {
  margin-bottom: 16px;
}

.upload-progress-card {
  margin-bottom: 12px;
  padding: 8px 12px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius);
  box-shadow: var(--shadow-sm);
}

.media-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 360px;
  overflow-y: auto;
  padding: 4px;
}

@media (max-width: 768px) {
  .media-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 480px) {
  .media-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.media-item {
  aspect-ratio: 1;
  position: relative;
  cursor: pointer;
  border-radius: var(--border-radius);
  overflow: hidden;
  border: 2px solid transparent;
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.media-item:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-card);
}

.media-item.selected {
  border-color: var(--color-primary);
}

.media-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.media-item:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.selected-overlay {
  position: absolute;
  inset: 0;
  background: rgba(74, 108, 247, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: var(--color-primary);
}

.pagination {
  margin-top: 16px;
  text-align: center;
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
