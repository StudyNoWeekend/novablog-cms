<template>
  <div class="markdown-editor">
    <Editor
      :value="modelValue"
      :plugins="plugins"
      :locale="bytemdLocale"
      :upload-images="handleUploadImage"
      @change="handleChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Editor } from '@bytemd/vue-next'
import gfm from '@bytemd/plugin-gfm'
import highlight from '@bytemd/plugin-highlight'
// bytemd 内置语言包：按当前界面语言静态查表（ja 包缺 limited 键，缺失项回落 bytemd 默认文案）
import zhHans from 'bytemd/locales/zh_Hans.json'
import zhHant from 'bytemd/locales/zh_Hant.json'
import en from 'bytemd/locales/en.json'
import ja from 'bytemd/locales/ja.json'
import ko from 'bytemd/locales/ko.json'
import 'bytemd/dist/index.css'
import { mediaApi } from '@/api/media'
import type { MediaItem } from '@/api/media'
import { message } from 'ant-design-vue'

const { t, locale } = useI18n()

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

// 编辑器内置文案按当前语言映射：zh-CN→zh_Hans、zh-TW→zh_Hant、en→en、ja→ja、ko→ko
const BYTEMD_LOCALES: Record<string, object> = {
  'zh-CN': zhHans,
  'zh-TW': zhHant,
  en,
  ja,
  ko,
}

const bytemdLocale = computed(() => BYTEMD_LOCALES[locale.value] ?? zhHans)

function handleChange(v: string) {
  emit('update:modelValue', v)
}

async function handleUploadImage(files: File[]): Promise<{ title: string; url: string }[]> {
  const results: { title: string; url: string }[] = []
  for (const file of files) {
    try {
      const res: MediaItem = await mediaApi.upload(file, { module: 'article' })
      results.push({
        title: file.name,
        url: res.url,
      })
    } catch {
      message.error(t('profile.editor.uploadFileFailed', { name: file.name }))
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
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: var(--border-radius, 8px);
}
</style>
