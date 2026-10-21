<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ t('menu.fitness.list') }}</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> {{ t('fitness.actions.create') }}
      </a-button>
    </div>

    <!-- 分类标签栏 -->
    <CategoryBar type="fitness" @select="handleCategorySelect" />

    <a-card>
      <div class="toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="t('fitness.list.searchPlaceholder')"
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
            <div>
              <div class="title-text">{{ record.title }}</div>
              <div class="sub-text">{{ formatDate(record.date) }}</div>
            </div>
          </template>
          <template v-else-if="column.key === 'type'">
            <a-tag v-if="record.category_name" color="blue">{{ record.category_name }}</a-tag>
            <a-tag v-else-if="record.type" :color="typeColor(record.type)">{{ typeLabel(record.type) }}</a-tag>
            <span v-else>—</span>
          </template>
          <template v-else-if="column.key === 'duration_min'">
            {{ record.duration_min }} {{ t('fitness.unit.minutes') }}
          </template>
          <template v-else-if="column.key === 'calories'">
            {{ record.calories }} {{ t('fitness.unit.kcal') }}
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 1 ? 'success' : 'default'">
              {{ record.status === 1 ? t('fitness.status.published') : t('fitness.status.draft') }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-button type="link" size="small" @click="toggleStatus(record)">
              {{ record.status === 1 ? t('common.actions.offline') : t('common.actions.publish') }}
            </a-button>
            <a-button type="link" size="small" @click="handleEdit(record)">{{ t('common.actions.edit') }}</a-button>
            <a-button type="link" size="small" danger @click="handleDelete(record)">{{ t('common.actions.delete') }}</a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="modalVisible"
      :title="modalMode === 'create' ? t('fitness.actions.create') : t('fitness.actions.edit')"
      :confirm-loading="modalLoading"
      width="680px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <div class="form-row">
          <a-form-item :label="t('fitness.form.date')" required>
            <a-date-picker v-model:value="trainDate" style="width: 100%" />
          </a-form-item>
          <a-form-item :label="t('common.field.category')">
            <CategorySelect v-model="form.category_id" type="fitness" :placeholder="t('fitness.form.categoryPlaceholder')" />
          </a-form-item>
        </div>
        <a-form-item :label="t('common.field.title')" required>
          <a-input v-model:value="form.title" :placeholder="t('fitness.form.titlePlaceholder')" :maxlength="255" />
        </a-form-item>
        <div class="form-row">
          <a-form-item :label="t('fitness.form.duration')">
            <a-input-number v-model:value="form.duration_min" :min="0" style="width: 100%" />
          </a-form-item>
          <a-form-item :label="t('fitness.form.calories')">
            <a-input-number v-model:value="form.calories" :min="0" style="width: 100%" />
          </a-form-item>
        </div>
        <a-form-item :label="t('fitness.form.exerciseList')">
          <div class="dynamic-list">
            <div v-for="(item, index) in form.content" :key="index" class="dynamic-row">
              <a-input v-model:value="item.name" :placeholder="t('fitness.form.exerciseNamePlaceholder')" />
              <a-input v-model:value="item.sets" :placeholder="t('fitness.form.setsPlaceholder')" style="width: 90px" />
              <a-input v-model:value="item.reps" :placeholder="t('fitness.form.repsPlaceholder')" style="width: 110px" />
              <a-button type="text" danger @click="form.content.splice(index, 1)">
                <DeleteOutlined />
              </a-button>
            </div>
            <a-button type="dashed" block @click="form.content.push({ name: '', sets: '', reps: '' })">
              <PlusOutlined /> {{ t('fitness.form.addExercise') }}
            </a-button>
          </div>
        </a-form-item>
        <a-form-item :label="t('common.field.remark')">
          <a-textarea v-model:value="form.notes" :placeholder="t('fitness.form.notesPlaceholder')" :rows="3" />
        </a-form-item>
        <a-form-item :label="t('common.field.status')">
          <a-radio-group v-model:value="form.status">
            <a-radio :value="0">{{ t('fitness.status.draft') }}</a-radio>
            <a-radio :value="1">{{ t('common.actions.publish') }}</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { fitnessApi } from '@/api/fitness'
import type { FitnessRecord, FitnessExercise } from '@/types/fitness'
import CategoryBar from '@/components/common/CategoryBar.vue'
import CategorySelect from '@/components/article/CategorySelect.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'
import { formatDate } from '@/utils/datetime'

const { t } = useI18n()

const keyword = ref('')
const categoryFilter = ref<string | undefined>(undefined)
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const list = ref<FitnessRecord[]>([])
const loading = ref(false)
const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(10)

const modalVisible = ref(false)
const modalLoading = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref('')
const trainDate = ref<Dayjs>(dayjs())

const form = reactive({
  title: '',
  category_id: undefined as string | undefined,
  duration_min: 0,
  calories: 0,
  content: [] as FitnessExercise[],
  notes: '',
  status: 0,
})

const columns = computed(() => [
  { key: 'title', title: t('fitness.columns.title'), ellipsis: true },
  { key: 'type', title: t('common.field.type'), width: 90 },
  { key: 'duration_min', title: t('fitness.columns.duration'), width: 100 },
  { key: 'calories', title: t('fitness.columns.calories'), width: 100 },
  { key: 'status', title: t('fitness.columns.status'), width: 90 },
  { key: 'action', title: t('common.table.actions'), width: 200 },
])

const paginationConfig = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (count: number) => t('common.table.total', { total: count }),
}))

watch(keyword, setDebounce)
watch([debouncedKeyword], () => {
  reset()
  fetchList()
})

onMounted(() => {
  fetchList()
})

async function fetchList() {
  loading.value = true
  try {
    const res = await fitnessApi.getList({
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

function typeLabel(type: string): string {
  return { strength: t('fitness.type.strength'), cardio: t('fitness.type.cardio'), stretch: t('fitness.type.stretch') }[type] || type
}

function typeColor(t: string): string {
  return { strength: 'volcano', cardio: 'geekblue', stretch: 'green' }[t] || 'default'
}

function resetForm() {
  trainDate.value = dayjs()
  form.title = ''
  form.category_id = undefined
  form.duration_min = 0
  form.calories = 0
  form.content = []
  form.notes = ''
  form.status = 0
}

function openCreateModal() {
  modalMode.value = 'create'
  editingId.value = ''
  resetForm()
  modalVisible.value = true
}

function handleEdit(record: FitnessRecord) {
  modalMode.value = 'edit'
  editingId.value = record.id
  resetForm()
  trainDate.value = record.date ? dayjs(record.date) : dayjs()
  form.title = record.title
  form.category_id = record.category_id || undefined
  form.duration_min = record.duration_min || 0
  form.calories = record.calories || 0
  form.content = (record.content || []).map((c) => ({ ...c }))
  form.notes = record.notes || ''
  form.status = record.status
  modalVisible.value = true
}

function buildPayload() {
  return {
    date: trainDate.value ? trainDate.value.toISOString() : undefined,
    title: form.title.trim(),
    category_id: form.category_id,
    duration_min: form.duration_min,
    calories: form.calories,
    content: form.content.filter((c) => c.name.trim()),
    notes: form.notes.trim() || undefined,
    status: form.status,
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning(t('fitness.message.titleRequired'))
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await fitnessApi.create(buildPayload())
      message.success(t('common.message.createSuccess'))
    } else {
      await fitnessApi.update(editingId.value, buildPayload())
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

async function toggleStatus(record: FitnessRecord) {
  try {
    await fitnessApi.update(record.id, { status: record.status === 1 ? 0 : 1 })
    message.success(record.status === 1 ? t('fitness.status.offline') : t('fitness.status.published'))
    fetchList()
  } catch {
    // 错误由拦截器处理
  }
}

function handleDelete(record: FitnessRecord) {
  Modal.confirm({
    title: t('fitness.message.deleteTitle'),
    content: t('fitness.message.deleteConfirm', { name: record.title }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await fitnessApi.remove(record.id)
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
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
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
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
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
</style>
