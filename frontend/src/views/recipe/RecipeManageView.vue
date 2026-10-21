<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ t('menu.recipe.list') }}</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> {{ t('recipe.actions.create') }}
      </a-button>
    </div>

    <!-- 分类标签栏 -->
    <CategoryBar type="recipe" @select="handleCategorySelect" />

    <a-card>
      <div class="toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="t('recipe.list.searchPlaceholder')"
          style="width: 240px"
          allow-clear
        />
      </div>

      <a-table
        :columns="columns"
        :data-source="list"
        :loading="loading"
        :pagination="paginationConfig"
        row-key="id"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'title'">
            <div class="title-cell">
              <img v-if="record.cover" :src="record.cover" class="cover-thumb" referrerpolicy="no-referrer" />
              <div v-else class="cover-thumb placeholder"><CoffeeOutlined /></div>
              <div>
                <div class="title-text">{{ record.title }}</div>
                <div v-if="record.category_name" class="sub-text">{{ record.category_name }}</div>
              </div>
            </div>
          </template>
          <template v-else-if="column.key === 'difficulty'">
            {{ difficultyLabel(record.difficulty) }}
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 1 ? 'success' : 'default'">
              {{ record.status === 1 ? t('recipe.status.published') : t('recipe.status.draft') }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-button type="link" size="small" @click="toggleStatus(record)">
              {{ record.status === 1 ? t('common.actions.offline') : t('common.actions.publish') }}
            </a-button>
            <a-button type="link" size="small" @click="handleEdit(record)">{{ t('common.actions.edit') }}</a-button>
            <PushSyncButton
              target-type="recipe"
              :target-id="record.id"
              :published="record.status === 1"
              @updated="fetchList"
            />
            <a-button type="link" size="small" danger @click="handleDelete(record)">{{ t('common.actions.delete') }}</a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="modalVisible"
      :title="modalMode === 'create' ? t('recipe.actions.create') : t('recipe.actions.edit')"
      :confirm-loading="modalLoading"
      width="720px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item :label="t('recipe.form.title')" required>
          <a-input v-model:value="form.title" :placeholder="t('recipe.form.titlePlaceholder')" :maxlength="255" />
        </a-form-item>
        <a-form-item :label="t('recipe.form.cover')">
          <MediaPicker v-model:visible="mediaPickerVisible" module="recipe" @selected="handleMediaSelected" />
          <ImageField v-model="form.cover" type="recipe" with-preset :search-keyword="form.title" :placeholder="t('recipe.form.coverUrlPlaceholder')">
            <template #extra>
              <a-button size="small" @click="mediaPickerVisible = true">
                <PictureOutlined /> {{ t('recipe.form.mediaLibrary') }}
              </a-button>
            </template>
          </ImageField>
        </a-form-item>
        <a-form-item :label="t('recipe.form.summary')">
          <a-textarea v-model:value="form.summary" :placeholder="t('recipe.form.summaryPlaceholder')" :rows="2" :maxlength="500" />
        </a-form-item>
        <div class="form-row">
          <a-form-item :label="t('recipe.form.difficulty')">
            <a-select v-model:value="form.difficulty">
              <a-select-option :value="1">{{ t('recipe.difficulty.easy') }}</a-select-option>
              <a-select-option :value="2">{{ t('recipe.difficulty.medium') }}</a-select-option>
              <a-select-option :value="3">{{ t('recipe.difficulty.hard') }}</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item :label="t('recipe.form.minutes')">
            <a-input-number v-model:value="form.minutes" :min="0" style="width: 100%" />
          </a-form-item>
          <a-form-item :label="t('recipe.form.servings')">
            <a-input-number v-model:value="form.servings" :min="1" style="width: 100%" />
          </a-form-item>
        </div>
        <div class="form-row">
          <a-form-item :label="t('common.field.category')">
            <CategorySelect v-model="form.category_id" type="recipe" :placeholder="t('recipe.form.categoryPlaceholder')" />
          </a-form-item>
          <a-form-item :label="t('recipe.form.tags')" class="span-2">
            <a-input v-model:value="form.tags" :placeholder="t('recipe.form.tagsPlaceholder')" :maxlength="500" />
          </a-form-item>
        </div>

        <a-form-item :label="t('recipe.form.ingredients')">
          <div class="dynamic-list">
            <div v-for="(item, index) in form.ingredients" :key="index" class="dynamic-row">
              <a-input v-model:value="item.name" :placeholder="t('recipe.form.ingredientNamePlaceholder')" />
              <a-input v-model:value="item.amount" :placeholder="t('recipe.form.ingredientAmountPlaceholder')" />
              <a-button type="text" danger @click="form.ingredients.splice(index, 1)">
                <DeleteOutlined />
              </a-button>
            </div>
            <a-button type="dashed" block @click="form.ingredients.push({ name: '', amount: '' })">
              <PlusOutlined /> {{ t('recipe.form.addIngredient') }}
            </a-button>
          </div>
        </a-form-item>

        <a-form-item :label="t('recipe.form.steps')">
          <div class="dynamic-list">
            <div v-for="(item, index) in form.steps" :key="index" class="dynamic-row">
              <span class="step-index">{{ index + 1 }}</span>
              <a-textarea v-model:value="item.text" :placeholder="t('recipe.form.stepPlaceholder')" :rows="1" :auto-size="true" />
              <a-button type="text" danger @click="form.steps.splice(index, 1)">
                <DeleteOutlined />
              </a-button>
            </div>
            <a-button type="dashed" block @click="form.steps.push({ text: '' })">
              <PlusOutlined /> {{ t('recipe.form.addStep') }}
            </a-button>
          </div>
        </a-form-item>

        <a-form-item :label="t('common.field.status')">
          <a-radio-group v-model:value="form.status">
            <a-radio :value="0">{{ t('recipe.status.draft') }}</a-radio>
            <a-radio :value="1">{{ t('common.actions.publish') }}</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import PushSyncButton from '@/components/community/PushSyncButton.vue'
import { message, Modal } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import {
  PlusOutlined,
  DeleteOutlined,
  PictureOutlined,
  CoffeeOutlined,
} from '@ant-design/icons-vue'
import { recipeApi } from '@/api/recipe'
import type { Recipe, RecipeIngredient, RecipeStep } from '@/types/recipe'
import MediaPicker from '@/components/media/MediaPicker.vue'
import ImageField from '@/components/media/ImageField.vue'
import CategoryBar from '@/components/common/CategoryBar.vue'
import CategorySelect from '@/components/article/CategorySelect.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const { t } = useI18n()

const keyword = ref('')
const categoryFilter = ref<string | undefined>(undefined)
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const list = ref<Recipe[]>([])
const loading = ref(false)
const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(10)

const modalVisible = ref(false)
const modalLoading = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref('')
const mediaPickerVisible = ref(false)

const form = reactive({
  title: '',
  cover: '',
  summary: '',
  difficulty: 1,
  minutes: 0,
  servings: 1,
  tags: '',
  category_id: undefined as string | undefined,
  ingredients: [] as RecipeIngredient[],
  steps: [] as RecipeStep[],
  status: 0,
})

const columns = computed(() => [
  { key: 'title', title: t('recipe.columns.recipe'), ellipsis: true },
  { key: 'difficulty', title: t('recipe.columns.difficulty'), width: 90 },
  { key: 'minutes', title: t('recipe.columns.minutes'), width: 100 },
  { key: 'status', title: t('common.field.status'), width: 90 },
  { key: 'updated_at', title: t('common.field.updateTime'), width: 170 },
  { key: 'action', title: t('common.table.actions'), width: 200 },
])

const paginationConfig = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (count: number) => t('recipe.list.total', { total: count }),
}))

watch(keyword, setDebounce)
watch(debouncedKeyword, () => {
  reset()
  fetchList()
})

onMounted(() => {
  fetchList()
})

async function fetchList() {
  loading.value = true
  try {
    const res = await recipeApi.getList({
      page: page.value,
      page_size: pageSize.value,
      keyword: debouncedKeyword.value || undefined,
      category_id: categoryFilter.value,
    })
    list.value = res.list || []
    total.value = res.total || 0
  } catch {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handleCategorySelect(id: string | undefined) {
  categoryFilter.value = id
  reset()
  fetchList()
}

function handleTableChange(p: { current?: number; pageSize?: number }) {
  changePagination(p.current || 1, p.pageSize)
  fetchList()
}

const DIFFICULTY_LABEL_KEYS = ['recipe.difficulty.easy', 'recipe.difficulty.medium', 'recipe.difficulty.hard']

function difficultyLabel(d: number): string {
  return t(DIFFICULTY_LABEL_KEYS[d - 1] || DIFFICULTY_LABEL_KEYS[0])
}

function resetForm() {
  form.title = ''
  form.cover = ''
  form.summary = ''
  form.difficulty = 1
  form.minutes = 0
  form.servings = 1
  form.tags = ''
  form.category_id = undefined
  form.ingredients = []
  form.steps = []
  form.status = 0
}

function openCreateModal() {
  modalMode.value = 'create'
  editingId.value = ''
  resetForm()
  modalVisible.value = true
}

function handleEdit(record: Recipe) {
  modalMode.value = 'edit'
  editingId.value = record.id
  resetForm()
  form.title = record.title
  form.cover = record.cover || ''
  form.summary = record.summary || ''
  form.difficulty = record.difficulty || 1
  form.minutes = record.minutes || 0
  form.servings = record.servings || 1
  form.tags = record.tags || ''
  form.category_id = record.category_id || undefined
  form.ingredients = (record.ingredients || []).map((i) => ({ ...i }))
  form.steps = (record.steps || []).map((s) => ({ ...s }))
  form.status = record.status
  modalVisible.value = true
}

function handleMediaSelected(media: { url: string }) {
  form.cover = media.url
}

function buildPayload() {
  return {
    title: form.title.trim(),
    cover: form.cover.trim() || undefined,
    summary: form.summary.trim() || undefined,
    difficulty: form.difficulty,
    minutes: form.minutes,
    servings: form.servings,
    tags: form.tags.trim() || undefined,
    category_id: form.category_id,
    ingredients: form.ingredients.filter((i) => i.name.trim()),
    steps: form.steps.filter((s) => s.text.trim()),
    status: form.status,
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning(t('recipe.message.titleRequired'))
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await recipeApi.create(buildPayload())
      message.success(t('common.message.createSuccess'))
    } else {
      await recipeApi.update(editingId.value, buildPayload())
      message.success(t('common.message.updateSuccess'))
    }
    modalVisible.value = false
    fetchList()
  } catch {
    // 错误由拦截器处理
  } finally {
    modalLoading.value = false
  }
}

async function toggleStatus(record: Recipe) {
  try {
    await recipeApi.update(record.id, { status: record.status === 1 ? 0 : 1 })
    message.success(record.status === 1 ? t('recipe.status.unpublished') : t('recipe.status.published'))
    fetchList()
  } catch {
    // 错误由拦截器处理
  }
}

function handleDelete(record: Recipe) {
  Modal.confirm({
    title: t('recipe.message.deleteTitle'),
    content: t('recipe.message.deleteConfirm', { name: record.title }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await recipeApi.remove(record.id)
        message.success(t('common.message.deleteSuccess'))
        if (list.value.length === 1 && page.value > 1) {
          page.value -= 1
        }
        fetchList()
      } catch {
        // 错误由拦截器处理
      }
    },
  })
}
</script>

<style scoped>
.toolbar {
  margin-bottom: 20px;
}

.title-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cover-thumb {
  width: 48px;
  height: 48px;
  object-fit: cover;
  border-radius: 6px;
  flex-shrink: 0;
}

.cover-thumb.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #F5F7FC;
  color: #8A93A8;
  font-size: 20px;
}

.title-text {
  font-weight: 500;
}

.sub-text {
  font-size: 12px;
  color: #8A93A8;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.span-2 {
  grid-column: span 2;
}

.cover-picker {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cover-preview {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid #E5E9F2;
}

.dynamic-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dynamic-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.step-index {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #F5F7FC;
  color: #667085;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
</style>
