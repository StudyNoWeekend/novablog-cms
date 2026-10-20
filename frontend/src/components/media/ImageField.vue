<template>
  <div class="image-field">
    <div class="field-main">
      <div class="field-preview">
        <img
          v-if="modelValue"
          :src="modelValue"
          :alt="$t('media.field.previewAlt')"
          referrerpolicy="no-referrer"
        />
        <span v-else class="field-empty">{{ $t('media.field.noImage') }}</span>
      </div>
      <div class="field-actions">
        <a-space wrap>
          <a-button size="small" :aria-label="$t('media.field.searchOnlineAria')" @click="pickerVisible = true">
            <SearchOutlined /> {{ $t('media.field.searchOnline') }}
          </a-button>
          <a-button size="small" :loading="uploading" :aria-label="$t('media.field.uploadLocalAria')" @click="handleUploadClick">
            <UploadOutlined /> {{ $t('media.field.uploadLocal') }}
          </a-button>
          <a-button v-if="modelValue" size="small" danger :aria-label="$t('media.field.clearAria')" @click="handleClear">
            {{ $t('media.field.clear') }}
          </a-button>
          <slot name="extra" />
        </a-space>
        <a-input
          v-model:value="modelValueProxy"
          class="field-url"
          size="small"
          :placeholder="placeholder || $t('media.field.urlPlaceholder')"
          :maxlength="500"
          allow-clear
        />
      </div>
    </div>

    <input
      ref="fileInputRef"
      type="file"
      :accept="IMAGE_ACCEPT"
      style="display: none"
      aria-hidden="true"
      @change="handleFileChange"
    />
    <ExternalImagePickerModal
      v-model:visible="pickerVisible"
      :type="type"
      :module="moduleKey"
      :initial-keyword="searchKeyword"
      @selected="handlePicked"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import { SearchOutlined, UploadOutlined } from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import { IMAGE_ACCEPT, validateImageFile } from '@/utils/upload'
import ExternalImagePickerModal from './ExternalImagePickerModal.vue'

const props = defineProps<{
  /** 图片地址（v-model） */
  modelValue: string
  /** 图源类型：tech=品牌图标, game=游戏封面, book=书籍封面, recipe=美食图片 */
  type: 'tech' | 'game' | 'book' | 'recipe'
  /** 打开在线搜索时预填的关键词（通常传表单里的名称字段） */
  searchKeyword?: string
  /** URL 输入框占位文案 */
  placeholder?: string
  /** 本地上传时是否同时生成"默认无 EXIF"预设并回填成品图（默认关闭；注意成品为 JPEG，透明 PNG 不适用） */
  withPreset?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const modelValueProxy = computed({
  get: () => props.modelValue,
  set: (v: string) => emit('update:modelValue', v || ''),
})

const { t } = useI18n()

/** 上传/转存归属的媒体库模块文件夹 key */
const moduleKey = computed(() => {
  const map: Record<string, string> = {
    tech: 'tech_stack',
    game: 'game',
    book: 'book',
    recipe: 'recipe',
  }
  return map[props.type] || ''
})

const pickerVisible = ref(false)
const uploading = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

function handlePicked(url: string) {
  emit('update:modelValue', url)
}

function handleUploadClick() {
  if (uploading.value) return
  fileInputRef.value?.click()
}

async function handleFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // accept 用通配 image/* 保证移动端直接跳图库，格式/大小限制在选取后由前端拦截
  const error = validateImageFile(file, ['jpg', 'jpeg', 'png', 'webp'])
  if (error) {
    message.error(error)
    target.value = ''
    return
  }

  uploading.value = true
  try {
    if (props.withPreset) {
      // 「默认无 EXIF」为后端预设约定的固定名称，参与后端匹配，不做翻译
      const res = await mediaApi.uploadWithPreset(file, '默认无 EXIF', { module: moduleKey.value })
      const url = res?.preset?.output_url || res?.media?.url
      if (url) {
        emit('update:modelValue', url)
        message.success(t('common.message.uploadSuccess'))
      }
    } else {
      const res = await mediaApi.upload(file, { module: moduleKey.value })
      if (res?.url) {
        emit('update:modelValue', res.url)
        message.success(t('common.message.uploadSuccess'))
      }
    }
  } catch {
    // 错误提示已由 axios 拦截器统一弹出
  } finally {
    uploading.value = false
    if (fileInputRef.value) {
      fileInputRef.value.value = ''
    }
  }
}

function handleClear() {
  emit('update:modelValue', '')
}
</script>

<style scoped>
.image-field {
  width: 100%;
}

.field-main {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.field-preview {
  flex-shrink: 0;
  width: 72px;
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed var(--border-color);
  border-radius: var(--border-radius);
  background: var(--bg-card);
  overflow: hidden;
}

.field-preview img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.field-empty {
  color: var(--text-secondary);
  font-size: 12px;
}

.field-actions {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-url {
  max-width: 320px;
}
</style>
