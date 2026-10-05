<template>
  <div class="image-field">
    <div class="field-main">
      <div class="field-preview">
        <img
          v-if="modelValue"
          :src="modelValue"
          alt="图片预览"
          referrerpolicy="no-referrer"
        />
        <span v-else class="field-empty">暂无图片</span>
      </div>
      <div class="field-actions">
        <a-space wrap>
          <a-button size="small" aria-label="在线搜索图片" @click="pickerVisible = true">
            <SearchOutlined /> 在线搜索
          </a-button>
          <a-button size="small" :loading="uploading" aria-label="本地上传图片" @click="handleUploadClick">
            <UploadOutlined /> 本地上传
          </a-button>
          <a-button v-if="modelValue" size="small" danger aria-label="清除图片" @click="handleClear">
            清除
          </a-button>
          <slot name="extra" />
        </a-space>
        <a-input
          v-model:value="modelValueProxy"
          class="field-url"
          size="small"
          :placeholder="placeholder || '或直接粘贴图片地址'"
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
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const modelValueProxy = computed({
  get: () => props.modelValue,
  set: (v: string) => emit('update:modelValue', v || ''),
})

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
    const res = await mediaApi.upload(file, { module: moduleKey.value })
    if (res?.url) {
      emit('update:modelValue', res.url)
      message.success('上传成功')
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
