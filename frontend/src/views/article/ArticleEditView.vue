<template>
  <div v-if="loading" class="p-8">
    <a-skeleton active :paragraph="{ rows: 10 }" />
  </div>
  <div v-else class="editor-page">
    <!-- 顶部工具栏 -->
    <div class="editor-toolbar">
      <a-button type="text" @click="isEdit ? router.push('/articles') : router.back()">
        <ArrowLeftOutlined /> {{ isEdit ? $t('article.edit.backToList') : $t('common.actions.back') }}
      </a-button>
      <a-input
        v-model:value="form.title"
        :placeholder="$t('article.edit.titlePlaceholder')"
        class="title-input"
        size="large"
      />

      <!-- 编辑器切换 -->
      <a-radio-group v-model:value="form.type" size="small" class="editor-switch">
        <a-radio-button :value="1">
          <FileMarkdownOutlined /> Markdown
        </a-radio-button>
        <a-radio-button :value="2">
          <FileTextOutlined /> {{ $t('article.edit.richText') }}
        </a-radio-button>
      </a-radio-group>

      <div class="toolbar-actions">
        <a-button @click="handleSaveDraft" :loading="saving">{{ $t('article.edit.saveDraft') }}</a-button>
        <a-button type="primary" @click="handlePublish" :loading="saving">
          <SendOutlined /> {{ $t('common.actions.publish') }}
        </a-button>
        <PushSyncButton
          v-if="isEdit"
          target-type="article"
          :target-id="String(route.params.id)"
        />
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
import { reactive, ref, computed, onMounted, watchEffect } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import PushSyncButton from '@/components/community/PushSyncButton.vue'
import { ArrowLeftOutlined, SendOutlined, FileMarkdownOutlined, FileTextOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { useArticleStore } from '@/stores/article'
import MarkdownEditor from '@/components/editor/MarkdownEditor.vue'
import RichTextEditor from '@/components/editor/RichTextEditor.vue'
import ArticleForm from '@/components/article/ArticleForm.vue'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const store = useArticleStore()
const saving = ref(false)

const isEdit = computed(() => !!route.params.id)
// 页面标题复用 menu.article.*（与侧边栏/路由 meta 共用同一批 key）
const title = computed(() => (isEdit.value ? t('menu.article.edit') : t('menu.article.create')))

watchEffect(() => {
  document.title = title.value
})

const loading = ref(isEdit.value)

const form = reactive({
  title: '',
  content: '',
  summary: '',
  cover_image: '',
  coverMediaId: '',
  coverPresetId: '',
  category_id: undefined as string | undefined,
  tag_ids: [] as string[],
  type: 1,   // 默认 Markdown
  is_top: false,
  is_comment: true,
})

onMounted(async () => {
  if (!isEdit.value) return
  const id = route.params.id as string
  try {
    await store.fetchDetail(id)
    const article = store.currentArticle
    if (article) {
      form.title = article.title
      form.content = article.content
      form.summary = article.summary || ''
      form.cover_image = article.cover_image || ''
      form.coverMediaId = article.cover_media_id || ''
      form.coverPresetId = article.cover_preset_id || ''
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

// 组装提交参数：媒体库 ID 字段转 snake_case，ID 为空时走外链 URL 模式
function buildPayload(status: number) {
  const { coverMediaId, coverPresetId, ...rest } = form
  return {
    ...rest,
    cover_media_id: coverMediaId || undefined,
    cover_preset_id: coverPresetId || undefined,
    status,
  }
}

async function handleSaveDraft() {
  if (!form.title.trim()) {
    message.warning(t('article.edit.titleRequired'))
    return
  }
  saving.value = true
  try {
    if (isEdit.value) {
      await store.update(route.params.id as string, buildPayload(1))
      message.success(t('article.edit.draftSaved'))
    } else {
      const article = await store.create(buildPayload(1))
      message.success(t('article.edit.draftSaved'))
      router.replace(`/articles/${article.id}/edit`)
    }
  } finally {
    saving.value = false
  }
}

async function handlePublish() {
  if (!form.title.trim()) {
    message.warning(t('article.edit.titleRequired'))
    return
  }
  saving.value = true
  try {
    if (isEdit.value) {
      await store.update(route.params.id as string, buildPayload(2))
    } else {
      await store.create(buildPayload(2))
    }
    message.success(t('article.edit.published'))
    router.push('/articles')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.editor-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 56px - 48px); /* header height + padding */
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
