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
import { shallowRef, onBeforeUnmount } from 'vue'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import { type IDomEditor, type IEditorConfig } from '@wangeditor/editor'
import { mediaApi } from '@/api/media'
import { message } from 'ant-design-vue'
import '@wangeditor/editor/dist/css/style.css'

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

// 编辑器配置
const editorConfig: Partial<IEditorConfig> = {
  placeholder: '请输入文章内容...',
  scroll: false,
  MENU_CONF: {
    uploadImage: {
      customUpload: async (file: File, insertFn: (url: string, alt?: string, href?: string) => void) => {
        try {
          const res = await mediaApi.upload(file)
          insertFn(res.url, file.name, res.url)
        } catch {
          message.error('图片上传失败')
        }
      },
    },
    uploadVideo: {
      customUpload: async (file: File, insertFn: (url: string, poster?: string) => void) => {
        try {
          const res = await mediaApi.upload(file)
          insertFn(res.url, '')
        } catch {
          message.error('视频上传失败')
        }
      },
    },
  },
}

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
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
  overflow: hidden;
}

.toolbar {
  border-bottom: 1px solid var(--border-color, #e2e8f0);
}

.editor-content-area {
  flex: 1;
  overflow-y: auto;
}

:deep(.w-e-text-container) {
  background: #fff;
}
</style>
