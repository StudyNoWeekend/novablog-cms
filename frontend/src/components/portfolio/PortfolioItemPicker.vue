<template>
  <a-modal
    v-model:open="open"
    :title="title"
    width="780px"
    :footer="null"
    :destroy-on-close="true"
    @cancel="handleCancel"
  >
    <a-steps :current="currentStep" size="small" class="picker-steps">
      <a-step title="选择原图" />
      <a-step title="选择预设" />
      <a-step v-if="!presetOnly" title="填写信息" />
    </a-steps>

    <!-- 步骤一：媒体库图片列表 -->
    <div v-if="currentStep === 0">
      <div class="step1-toolbar">
        <a-input-search
          v-model:value="mediaKeyword"
          placeholder="搜索文件名..."
          style="width: 240px"
          allow-clear
        />
        <input
          ref="uploadInputRef"
          type="file"
          accept="image/*"
          style="display: none"
          aria-hidden="true"
          @change="handleUploadFile"
        />
        <a-button :loading="uploading" aria-label="本地上传" @click="uploadInputRef?.click()">
          <UploadOutlined /> 本地上传
        </a-button>
      </div>
      <a-spin :spinning="mediaLoading">
        <a-empty
          v-if="!mediaLoading && mediaList.length === 0"
          description="暂无图片"
          style="margin-top: 32px"
        />
        <div v-else class="picker-grid">
          <div
            v-for="item in mediaList"
            :key="item.id"
            class="picker-item"
            :class="{ selected: selectedMediaId === item.id }"
            @click="handleSelectMedia(item)"
          >
            <img :src="item.url" :alt="item.filename" />
            <div class="picker-item-name" :title="item.filename">{{ item.filename }}</div>
            <div v-if="selectedMediaId === item.id" class="selected-overlay">
              <CheckCircleFilled />
            </div>
          </div>
        </div>
      </a-spin>
      <div v-if="mediaTotal > mediaPageSize" class="picker-pagination">
        <a-pagination
          :current="mediaPage"
          :page-size="mediaPageSize"
          :total="mediaTotal"
          size="small"
          @change="handleMediaPageChange"
        />
      </div>
    </div>

    <!-- 步骤二：预设列表 -->
    <div v-else-if="currentStep === 1">
      <div class="step2-toolbar">
        <span class="step2-hint">选择要使用的预设版本，点击「制作预设」可现场制作新成品图</span>
        <a-button size="small" aria-label="制作预设" @click="workbenchVisible = true">
          <EditOutlined /> 制作预设
        </a-button>
      </div>
      <a-spin :spinning="presetLoading">
        <div v-if="!presetLoading && presetList.length === 0" class="preset-empty">
          <a-empty description="该图片暂无预设" />
          <a-button type="primary" aria-label="制作预设" @click="workbenchVisible = true">
            <EditOutlined /> 去制作预设
          </a-button>
        </div>
        <div v-else class="picker-grid">
          <div
            v-for="preset in presetList"
            :key="preset.id"
            class="picker-item"
            :class="{ selected: selectedPresetId === preset.id }"
            @click="selectedPresetId = preset.id"
          >
            <img :src="preset.output_url" :alt="preset.name" />
            <div class="picker-item-name" :title="preset.name">{{ preset.name }}</div>
            <div v-if="selectedPresetId === preset.id" class="selected-overlay">
              <CheckCircleFilled />
            </div>
          </div>
        </div>
      </a-spin>
    </div>

    <!-- 步骤三：填写信息 -->
    <div v-else-if="currentStep === 2">
      <a-form layout="vertical">
        <a-form-item label="作品名称" required>
          <a-input
            v-model:value="itemTitle"
            placeholder="请输入作品名称"
            :maxlength="100"
            show-count
          />
        </a-form-item>
        <a-form-item label="作品介绍">
          <a-textarea
            v-model:value="itemDescription"
            placeholder="请输入作品介绍（可选）"
            :rows="4"
            :maxlength="500"
            show-count
          />
        </a-form-item>
        <div v-if="selectedPreset" class="confirm-preview">
          <img :src="selectedPreset.output_url" alt="预览" />
          <div class="confirm-preview-name">{{ selectedPreset.name }}</div>
        </div>
      </a-form>
    </div>

    <!-- 底部操作 -->
    <div class="picker-footer">
      <a-button v-if="currentStep > 0" @click="handlePrev">上一步</a-button>
      <div class="footer-right">
        <a-button @click="handleCancel">取消</a-button>
        <a-button
          v-if="currentStep < (presetOnly ? 1 : 2)"
          type="primary"
          :disabled="!canNext"
          @click="handleNext"
        >
          下一步
        </a-button>
        <a-button
          v-else
          type="primary"
          :disabled="presetOnly ? !canNext : !itemTitle.trim()"
          @click="handleConfirm"
        >
          确认选择
        </a-button>
      </div>
    </div>
    <MediaPresetWorkbench
      v-if="selectedMediaId"
      v-model:visible="workbenchVisible"
      :media-id="selectedMediaId"
      :original-url="selectedMedia?.url || ''"
      @saved="fetchPresets"
    />
  </a-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import { CheckCircleFilled, EditOutlined, UploadOutlined } from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaItem } from '@/api/media'
import type { MediaPreset } from '@/types/api'
import { useDebounce } from '@/composables/useDebounce'
import { validateImageFile } from '@/utils/upload'
import MediaPresetWorkbench from '@/components/media/MediaPresetWorkbench.vue'

const props = defineProps<{
  visible: boolean
  title?: string
  presetOnly?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (
    e: 'selected',
    payload: {
      preset_id: string
      title: string
      description: string
      output_url?: string
      preset_name?: string
    },
  ): void
}>()

const open = ref(props.visible)
const currentStep = ref(0)

// 步骤一
const mediaKeyword = ref('')
const { debouncedValue: debouncedMediaKeyword, setDebounce: setMediaDebounce } = useDebounce(mediaKeyword)
const mediaList = ref<MediaItem[]>([])
const mediaLoading = ref(false)
const mediaPage = ref(1)
const mediaPageSize = 20
const mediaTotal = ref(0)
const selectedMediaId = ref<string | null>(null)
const selectedMedia = ref<MediaItem | null>(null)
const uploading = ref(false)
const uploadInputRef = ref<HTMLInputElement | null>(null)

// 步骤二
const presetList = ref<MediaPreset[]>([])
const presetLoading = ref(false)
const selectedPresetId = ref<string | null>(null)
const workbenchVisible = ref(false)

// 步骤三
const itemTitle = ref('')
const itemDescription = ref('')

const selectedPreset = computed(() =>
  presetList.value.find((p) => p.id === selectedPresetId.value) || null,
)

const canNext = computed(() => {
  if (currentStep.value === 0) return !!selectedMediaId.value
  if (currentStep.value === 1) return !!selectedPresetId.value
  return true
})

watch(
  () => props.visible,
  (v) => {
    open.value = v
    if (v) {
      resetState()
      fetchMediaList()
    }
  },
)

watch(open, (v) => {
  if (!v) emit('update:visible', false)
})

watch(mediaKeyword, setMediaDebounce)

watch(debouncedMediaKeyword, () => {
  mediaPage.value = 1
  fetchMediaList()
})

async function fetchMediaList() {
  mediaLoading.value = true
  try {
    const res = await mediaApi.getList({
      file_type: 1,
      keyword: debouncedMediaKeyword.value || undefined,
      page: mediaPage.value,
      page_size: mediaPageSize,
    })
    mediaList.value = res.list || []
    mediaTotal.value = res.total || 0
  } catch {
    mediaList.value = []
    mediaTotal.value = 0
  } finally {
    mediaLoading.value = false
  }
}

function handleMediaPageChange(p: number) {
  mediaPage.value = p
  fetchMediaList()
}

function handleSelectMedia(item: MediaItem) {
  selectedMediaId.value = item.id
  selectedMedia.value = item
}

/** 本地上传：同步生成"默认无 EXIF"预设，成功后自动选中并进入选预设步骤 */
async function handleUploadFile(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const error = validateImageFile(file, ['jpg', 'jpeg', 'png', 'webp'])
  if (error) {
    message.error(error)
    target.value = ''
    return
  }

  uploading.value = true
  try {
    const res = await mediaApi.uploadWithPreset(file, '默认无 EXIF', { module: 'portfolio' })
    if (res.media && res.preset) {
      mediaList.value = [res.media, ...mediaList.value]
      mediaTotal.value += 1
      selectedMediaId.value = res.media.id
      selectedMedia.value = res.media
      currentStep.value = 1
      selectedPresetId.value = null
      await fetchPresets()
      message.success('上传成功，请选择预设')
    }
  } catch {
    message.error('上传失败，请重试')
  } finally {
    uploading.value = false
    if (uploadInputRef.value) {
      uploadInputRef.value.value = ''
    }
  }
}

async function fetchPresets() {
  if (!selectedMediaId.value) return
  presetLoading.value = true
  try {
    const res = await mediaApi.getPresets(selectedMediaId.value)
    presetList.value = res.list || []
  } catch {
    presetList.value = []
  } finally {
    presetLoading.value = false
  }
}

function handleNext() {
  if (currentStep.value === 0 && selectedMediaId.value) {
    currentStep.value = 1
    selectedPresetId.value = null
    fetchPresets()
  } else if (currentStep.value === 1 && selectedPresetId.value && !props.presetOnly) {
    currentStep.value = 2
  }
}

function handlePrev() {
  if (currentStep.value > 0) currentStep.value -= 1
}

function handleConfirm() {
  if (!selectedPresetId.value) return
  const preset = selectedPreset.value
  if (props.presetOnly) {
    emit('selected', {
      preset_id: selectedPresetId.value,
      title: preset?.name || '',
      description: '',
      output_url: preset?.output_url || '',
      preset_name: preset?.name || '',
    })
    open.value = false
    return
  }
  if (!itemTitle.value.trim()) return
  emit('selected', {
    preset_id: selectedPresetId.value,
    title: itemTitle.value.trim(),
    description: itemDescription.value.trim(),
    output_url: preset?.output_url || '',
    preset_name: preset?.name || '',
  })
  open.value = false
}

function handleCancel() {
  open.value = false
}

function resetState() {
  currentStep.value = 0
  mediaKeyword.value = ''
  mediaPage.value = 1
  mediaList.value = []
  mediaTotal.value = 0
  selectedMediaId.value = null
  selectedMedia.value = null
  presetList.value = []
  selectedPresetId.value = null
  itemTitle.value = ''
  itemDescription.value = ''
  workbenchVisible.value = false
}
</script>

<style scoped>
.picker-steps {
  margin-bottom: 24px;
}

.step1-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.step2-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.step2-hint {
  font-size: 13px;
  color: var(--text-secondary, #667085);
}

.preset-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 24px 0;
}

.picker-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 420px;
  overflow-y: auto;
}

.picker-item {
  position: relative;
  cursor: pointer;
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid transparent;
  transition: border-color 0.2s;
  background: var(--bg-secondary, #F5F7FC);
}

.picker-item:hover {
  border-color: var(--primary, #526FE8);
}

.picker-item.selected {
  border-color: var(--primary, #526FE8);
}

.picker-item img {
  width: 100%;
  aspect-ratio: 1;
  object-fit: cover;
  display: block;
}

.picker-item-name {
  font-size: 12px;
  line-height: 1.4;
  padding: 4px 8px;
  color: var(--text-secondary, #667085);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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

.picker-pagination {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}

.picker-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 24px;
}

.footer-right {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.confirm-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: var(--bg-secondary, #F5F7FC);
  border-radius: 8px;
}

.confirm-preview img {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
}

.confirm-preview-name {
  font-size: 14px;
  color: var(--text-secondary, #667085);
}
</style>
