<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">美食菜谱</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> 新建菜谱
      </a-button>
    </div>

    <a-card>
      <div class="toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索菜谱..."
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
              <span class="title-text">{{ record.title }}</span>
            </div>
          </template>
          <template v-else-if="column.key === 'difficulty'">
            {{ difficultyLabel(record.difficulty) }}
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 1 ? 'success' : 'default'">
              {{ record.status === 1 ? '已发布' : '草稿' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-button type="link" size="small" @click="toggleStatus(record)">
              {{ record.status === 1 ? '下架' : '发布' }}
            </a-button>
            <a-button type="link" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="link" size="small" danger @click="handleDelete(record)">删除</a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="modalVisible"
      :title="modalMode === 'create' ? '新建菜谱' : '编辑菜谱'"
      :confirm-loading="modalLoading"
      width="720px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item label="菜谱名称" required>
          <a-input v-model:value="form.title" placeholder="请输入菜谱名称" :maxlength="255" />
        </a-form-item>
        <a-form-item label="封面图">
          <MediaPicker v-model:visible="mediaPickerVisible" module="recipe" @selected="handleMediaSelected" />
          <ImageField v-model="form.cover" type="recipe" :search-keyword="form.title" placeholder="或直接粘贴封面地址">
            <template #extra>
              <a-button size="small" @click="mediaPickerVisible = true">
                <PictureOutlined /> 媒体库
              </a-button>
            </template>
          </ImageField>
        </a-form-item>
        <a-form-item label="简介">
          <a-textarea v-model:value="form.summary" placeholder="一句话介绍这道菜" :rows="2" :maxlength="500" />
        </a-form-item>
        <div class="form-row">
          <a-form-item label="难度">
            <a-select v-model:value="form.difficulty">
              <a-select-option :value="1">简单</a-select-option>
              <a-select-option :value="2">中等</a-select-option>
              <a-select-option :value="3">困难</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="耗时（分钟）">
            <a-input-number v-model:value="form.minutes" :min="0" style="width: 100%" />
          </a-form-item>
          <a-form-item label="份量（人份）">
            <a-input-number v-model:value="form.servings" :min="1" style="width: 100%" />
          </a-form-item>
        </div>
        <a-form-item label="标签（逗号分隔）">
          <a-input v-model:value="form.tags" placeholder="如：家常菜,下饭菜" :maxlength="500" />
        </a-form-item>

        <a-form-item label="食材清单">
          <div class="dynamic-list">
            <div v-for="(item, index) in form.ingredients" :key="index" class="dynamic-row">
              <a-input v-model:value="item.name" placeholder="食材名，如：番茄" />
              <a-input v-model:value="item.amount" placeholder="用量，如：2个" />
              <a-button type="text" danger @click="form.ingredients.splice(index, 1)">
                <DeleteOutlined />
              </a-button>
            </div>
            <a-button type="dashed" block @click="form.ingredients.push({ name: '', amount: '' })">
              <PlusOutlined /> 添加食材
            </a-button>
          </div>
        </a-form-item>

        <a-form-item label="烹饪步骤">
          <div class="dynamic-list">
            <div v-for="(item, index) in form.steps" :key="index" class="dynamic-row">
              <span class="step-index">{{ index + 1 }}</span>
              <a-textarea v-model:value="item.text" placeholder="这一步做什么" :rows="1" :auto-size="true" />
              <a-button type="text" danger @click="form.steps.splice(index, 1)">
                <DeleteOutlined />
              </a-button>
            </div>
            <a-button type="dashed" block @click="form.steps.push({ text: '' })">
              <PlusOutlined /> 添加步骤
            </a-button>
          </div>
        </a-form-item>

        <a-form-item label="状态">
          <a-radio-group v-model:value="form.status">
            <a-radio :value="0">草稿</a-radio>
            <a-radio :value="1">发布</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
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
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const keyword = ref('')
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
  ingredients: [] as RecipeIngredient[],
  steps: [] as RecipeStep[],
  status: 0,
})

const columns = [
  { key: 'title', title: '菜谱', ellipsis: true },
  { key: 'difficulty', title: '难度', width: 90 },
  { key: 'minutes', title: '耗时(分)', width: 100 },
  { key: 'status', title: '状态', width: 90 },
  { key: 'updated_at', title: '更新时间', width: 170 },
  { key: 'action', title: '操作', width: 200 },
]

const paginationConfig = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (t: number) => `共 ${t} 篇`,
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

function handleTableChange(p: { current?: number; pageSize?: number }) {
  changePagination(p.current || 1, p.pageSize)
  fetchList()
}

function difficultyLabel(d: number): string {
  return ['简单', '中等', '困难'][d - 1] || '简单'
}

function resetForm() {
  form.title = ''
  form.cover = ''
  form.summary = ''
  form.difficulty = 1
  form.minutes = 0
  form.servings = 1
  form.tags = ''
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
    ingredients: form.ingredients.filter((i) => i.name.trim()),
    steps: form.steps.filter((s) => s.text.trim()),
    status: form.status,
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning('请输入菜谱名称')
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await recipeApi.create(buildPayload())
      message.success('创建成功')
    } else {
      await recipeApi.update(editingId.value, buildPayload())
      message.success('更新成功')
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
    message.success(record.status === 1 ? '已下架' : '已发布')
    fetchList()
  } catch {
    // 错误由拦截器处理
  }
}

function handleDelete(record: Recipe) {
  Modal.confirm({
    title: '确认删除',
    content: `删除「${record.title}」后无法恢复，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await recipeApi.remove(record.id)
        message.success('删除成功')
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
  background: #f5f5f5;
  color: #bfbfbf;
  font-size: 20px;
}

.title-text {
  font-weight: 500;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
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
  border: 1px solid #f0f0f0;
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
  background: #f0f2f5;
  color: #64748b;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
</style>
