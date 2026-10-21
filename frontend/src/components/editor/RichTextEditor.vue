<template>
  <div class="rich-text-editor">
    <Toolbar
      :editor="editorRef"
      :defaultConfig="toolbarConfig"
      class="toolbar"
    />
    <Editor
      :modelValue="modelValue"
      :defaultConfig="editorConfig"
      @update:modelValue="handleChange"
      @onCreated="handleCreated"
      class="editor-content-area"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, shallowRef, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import { type IDomEditor, type IEditorConfig } from '@wangeditor/editor'
import { mediaApi } from '@/api/media'
import { message } from 'ant-design-vue'
import '@wangeditor/editor/dist/css/style.css'

const { t } = useI18n()

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const editorRef = shallowRef<IDomEditor>()

// 工具栏配置
const toolbarConfig = {
  excludeKeys: [],
}

// 编辑器配置（computed 包裹，placeholder 跟随当前语言）
const editorConfig = computed<Partial<IEditorConfig>>(() => ({
  placeholder: t('profile.editor.contentPlaceholder'),
  scroll: false,
  MENU_CONF: {
    uploadImage: {
      customUpload: async (file: File, insertFn: (url: string, alt?: string, href?: string) => void) => {
        try {
          const res = await mediaApi.upload(file, { module: 'article' })
          insertFn(res.url, file.name, res.url)
        } catch {
          message.error(t('profile.editor.imageUploadFailed'))
        }
      },
    },
    uploadVideo: {
      customUpload: async (file: File, insertFn: (url: string, poster?: string) => void) => {
        try {
          const res = await mediaApi.upload(file, { module: 'article' })
          insertFn(res.url, '')
        } catch {
          message.error(t('profile.editor.videoUploadFailed'))
        }
      },
    },
  },
}))

function handleCreated(editor: IDomEditor) {
  editorRef.value = editor
}

function handleChange(val: string) {
  emit('update:modelValue', val)
}

onBeforeUnmount(() => {
  const editor = editorRef.value
  if (editor) {
    editor.destroy()
  }
})
</script>

<style scoped>
.rich-text-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: var(--border-radius, 8px);
  overflow: hidden;
}

.toolbar {
  border-bottom: 1px solid var(--border-color, #E5E9F2);
}

.editor-content-area {
  flex: 1;
  overflow-y: auto;
}

:deep(.w-e-text-container) {
  background: #fff;
}
</style>
