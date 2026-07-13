<template>
  <div class="markdown-editor">
    <Editor
      :value="modelValue"
      :plugins="plugins"
      :locale="zhHans"
      :upload-images="handleUploadImage"
      @change="handleChange"
    />
  </div>
</template>

<script setup lang="ts">
import { Editor } from '@bytemd/vue-next'
import gfm from '@bytemd/plugin-gfm'
import highlight from '@bytemd/plugin-highlight'
import zhHans from 'bytemd/locales/zh_Hans.json'
import 'bytemd/dist/index.css'
import { mediaApi } from '@/api/media'
import type { MediaItem } from '@/api/media'
import { message } from 'ant-design-vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const plugins = [
  gfm(),
  highlight(),
]

function handleChange(v: string) {
  emit('update:modelValue', v)
}

async function handleUploadImage(files: File[]): Promise<{ title: string; url: string }[]> {
  const results: { title: string; url: string }[] = []
  for (const file of files) {
    try {
      const res: MediaItem = await mediaApi.upload(file)
      results.push({
        title: file.name,
        url: res.url,
      })
    } catch {
      message.error(`上传 ${file.name} 失败`)
    }
  }
  return results
}
</script>

<style scoped>
.markdown-editor {
  height: 100%;
}

:deep(.bytemd) {
  height: 100%;
  min-height: 400px;
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
}
</style>