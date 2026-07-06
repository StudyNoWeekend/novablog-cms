<template>
  <div class="article-form">
    <!-- 摘要 -->
    <div class="form-row">
      <a-textarea
        v-model:value="form.summary"
        :rows="2"
        placeholder="文章摘要，将显示在列表和SEO中..."
        class="summary-input"
      />
    </div>

    <!-- 工具栏 -->
    <div class="form-toolbar">
      <!-- 封面图 -->
      <div class="toolbar-item cover-item">
        <div class="cover-upload-compact" @click="showMediaPicker = true">
          <template v-if="form.cover_image">
            <img :src="form.cover_image" class="cover-preview-compact" />
            <div class="cover-overlay-compact">
              <PictureOutlined />
            </div>
          </template>
          <div v-else class="cover-placeholder-compact">
            <PictureOutlined />
            <span>封面</span>
          </div>
        </div>
      </div>

      <!-- 分类 -->
      <div class="toolbar-item">
        <a-select
          v-model:value="form.category_id"
          placeholder="分类"
          style="width: 140px"
          :loading="categoriesLoading"
        >
          <a-select-option v-for="cat in categories" :key="cat.id" :value="cat.id">
            {{ cat.name }}
          </a-select-option>
        </a-select>
      </div>

      <!-- 标签 -->
      <div class="toolbar-item">
        <a-select
          v-model:value="form.tag_ids"
          mode="multiple"
          placeholder="标签"
          style="min-width: 160px; max-width: 240px"
          :loading="tagsLoading"
          :options="tagOptions"
          :max-tag-count="1"
        />
      </div>

      <!-- 置顶 -->
      <div class="toolbar-item switch-item">
        <a-switch v-model:checked="form.is_top" size="small" />
        <span class="switch-label">置顶</span>
      </div>

      <!-- 评论 -->
      <div class="toolbar-item switch-item">
        <a-switch v-model:checked="form.is_comment" size="small" />
        <span class="switch-label">评论</span>
      </div>
    </div>

    <!-- 媒体选择器弹窗 -->
    <MediaPicker v-model:visible="showMediaPicker" @selected="handleMediaSelected" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, nextTick } from 'vue'
import { PictureOutlined } from '@ant-design/icons-vue'
import { categoryApi } from '@/api/category'
import { tagApi } from '@/api/tag'
import type { Category } from '@/types/category'
import type { Tag } from '@/types/tag'
import MediaPicker from '@/components/media/MediaPicker.vue'

export interface ArticleFormData {
  title: string
  content: string
  summary: string
  cover_image: string
  category_id: string | undefined
  tag_ids: string[]
  type: number           // 1=Markdown 2=富文本
  is_top: boolean
  is_comment: boolean
}

const props = defineProps<{
  modelValue: ArticleFormData
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ArticleFormData): void
}>()

const form = reactive<ArticleFormData>({ ...props.modelValue })
const categories = ref<Category[]>([])
const tags = ref<Tag[]>([])
const categoriesLoading = ref(false)
const tagsLoading = ref(false)
const showMediaPicker = ref(false)

// 防止循环 watch：从父组件同步过来的变更不再 emit 回去
let isSyncingFromParent = false

const tagOptions = computed(() =>
  tags.value.map(t => ({ label: t.name, value: t.id }))
)

// 当 form 变化时同步到父组件
watch(() => ({ ...form }), (val) => {
  if (!isSyncingFromParent) {
    emit('update:modelValue', val as ArticleFormData)
  }
}, { deep: true })

// 当 modelValue 外部变化时同步到 form
watch(() => props.modelValue, (val) => {
  isSyncingFromParent = true
  Object.assign(form, val)
  nextTick(() => { isSyncingFromParent = false })
}, { deep: true })

onMounted(async () => {
  loadCategories()
  loadTags()
})

async function loadCategories() {
  categoriesLoading.value = true
  try { categories.value = await categoryApi.getList() as unknown as Category[] } catch {}
  finally { categoriesLoading.value = false }
}

async function loadTags() {
  tagsLoading.value = true
  try { tags.value = await tagApi.getList() as unknown as Tag[] } catch {}
  finally { tagsLoading.value = false }
}

function handleMediaSelected(media: { url: string }) {
  form.cover_image = media.url
}
</script>

<style scoped>
.article-form {
  width: 100%;
}
.form-row {
  margin-bottom: 8px;
}
.summary-input {
  resize: none;
}
.form-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.toolbar-item {
  display: flex;
  align-items: center;
}
.cover-item {
  flex-shrink: 0;
}
.cover-upload-compact {
  position: relative;
  cursor: pointer;
  width: 80px;
  height: 60px;
  border-radius: 6px;
  overflow: hidden;
}
.cover-upload-compact:hover .cover-overlay-compact {
  opacity: 1;
}
.cover-preview-compact {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.cover-placeholder-compact {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #f8fafc;
  border: 1px dashed var(--border-color, #cbd5e1);
  border-radius: 6px;
  font-size: 11px;
  color: var(--text-tertiary, #94a3b8);
  gap: 2px;
  transition: all 0.2s;
}
.cover-placeholder-compact:hover {
  border-color: var(--primary, #3b82f6);
  color: var(--primary, #3b82f6);
  background: #f0f7ff;
}
.cover-overlay-compact {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 12px;
  opacity: 0;
  transition: opacity 0.2s;
  border-radius: 6px;
}
.switch-item {
  display: flex;
  align-items: center;
  gap: 6px;
}
.switch-label {
  font-size: 13px;
  color: var(--text-secondary, #64748b);
  white-space: nowrap;
}
</style>
