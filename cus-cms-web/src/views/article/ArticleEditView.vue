<template>
  <div v-if="loading" class="p-8">
    <a-skeleton active :paragraph="{ rows: 10 }" />
  </div>
  <div v-else class="editor-page">
    <!-- 顶部工具栏 -->
    <div class="editor-toolbar">
      <a-button type="text" @click="router.push('/articles')">
        <ArrowLeftOutlined /> 返回列表
      </a-button>
      <a-input
        v-model:value="form.title"
        placeholder="输入文章标题..."
        class="title-input"
        size="large"
      />

      <!-- 编辑器切换 -->
      <a-radio-group v-model:value="form.type" size="small" class="editor-switch">
        <a-radio-button :value="1">
          <FileMarkdownOutlined /> Markdown
        </a-radio-button>
        <a-radio-button :value="2">
          <FileTextOutlined /> 富文本
        </a-radio-button>
      </a-radio-group>

      <div class="toolbar-actions">
        <a-button @click="handleSaveDraft" :loading="saving">保存草稿</a-button>
        <a-button type="primary" @click="handlePublish" :loading="saving">
          <SendOutlined /> 发布
        </a-button>
      </div>
    </div>

    <!-- 编辑器 -->
    <div class="editor-body">
      <!-- 元信息栏 -->
      <div class="meta-bar-wrapper">
        <ArticleForm :modelValue="form" @update:modelValue="Object.assign(form, $event)" />
      </div>

      <MarkdownEditor v-if="form.type === 1" v-model="form.content" />
      <RichTextEditor v-else v-model="form.content" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowLeftOutlined, SendOutlined, FileMarkdownOutlined, FileTextOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { useArticleStore } from '@/stores/article'
import MarkdownEditor from '@/components/editor/MarkdownEditor.vue'
import RichTextEditor from '@/components/editor/RichTextEditor.vue'
import ArticleForm from '@/components/article/ArticleForm.vue'

const router = useRouter()
const route = useRoute()
const store = useArticleStore()
const saving = ref(false)
const loading = ref(true)

const form = reactive({
  title: '',
  content: '',
  summary: '',
  cover_image: '',
  category_id: undefined as string | undefined,
  tag_ids: [] as string[],
  type: 1,
  is_top: false,
  is_comment: true,
})

onMounted(async () => {
  const id = route.params.id as string
  try {
    await store.fetchDetail(id)
    const article = store.currentArticle
    if (article) {
      form.title = article.title
      form.content = article.content
      form.summary = article.summary || ''
      form.cover_image = article.cover_image || ''
      form.category_id = article.category_id || undefined
      form.tag_ids = article.tag_ids || []
      form.type = article.type || 1
      form.is_top = article.is_top
      form.is_comment = article.is_comment
    }
  } finally {
    loading.value = false
  }
})

async function handleSaveDraft() {
  if (!form.title.trim()) { message.warning('请输入文章标题'); return }
  saving.value = true
  try {
    await store.update(route.params.id as string, { ...form, status: 1 })
    message.success('草稿已保存')
  } finally { saving.value = false }
}

async function handlePublish() {
  if (!form.title.trim()) { message.warning('请输入文章标题'); return }
  saving.value = true
  try {
    await store.update(route.params.id as string, { ...form, status: 2 })
    message.success('文章已发布')
    router.push('/articles')
  } finally { saving.value = false }
}
</script>

<style scoped>
.editor-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 56px - 48px);
  max-width: 1600px;
  margin: 0 auto;
  width: 100%;
}
.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  flex-shrink: 0;
  flex-wrap: wrap;
}
.title-input {
  flex: 1;
  min-width: 200px;
  max-width: 800px;
}
.editor-switch {
  flex-shrink: 0;
}
.toolbar-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}
.meta-bar-wrapper {
  width: 100%;
  padding: 0 0 12px 0;
  flex-shrink: 0;
}
.editor-body {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

@media (max-width: 768px) {
  .editor-page {
    height: auto;
    min-height: calc(100vh - 56px - 48px);
  }

  .toolbar-actions {
    margin-left: 0;
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
