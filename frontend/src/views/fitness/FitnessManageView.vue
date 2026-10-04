<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">健身训练</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> 记录训练
      </a-button>
    </div>

    <a-card>
      <div class="toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索训练..."
          style="width: 240px"
          allow-clear
        />
        <a-select v-model:value="typeFilter" style="width: 140px" allow-clear>
          <a-select-option value="strength">力量</a-select-option>
          <a-select-option value="cardio">有氧</a-select-option>
          <a-select-option value="stretch">拉伸</a-select-option>
        </a-select>
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
            <a-tag :color="typeColor(record.type)">{{ typeLabel(record.type) }}</a-tag>
          </template>
          <template v-else-if="column.key === 'duration_min'">
            {{ record.duration_min }} 分钟
          </template>
          <template v-else-if="column.key === 'calories'">
            {{ record.calories }} 千卡
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
      :title="modalMode === 'create' ? '记录训练' : '编辑训练'"
      :confirm-loading="modalLoading"
      width="680px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <div class="form-row">
          <a-form-item label="训练日期" required>
            <a-date-picker v-model:value="trainDate" style="width: 100%" />
          </a-form-item>
          <a-form-item label="训练类型">
            <a-select v-model:value="form.type">
              <a-select-option value="strength">力量</a-select-option>
              <a-select-option value="cardio">有氧</a-select-option>
              <a-select-option value="stretch">拉伸</a-select-option>
            </a-select>
          </a-form-item>
        </div>
        <a-form-item label="标题" required>
          <a-input v-model:value="form.title" placeholder="如：卧推 + 划船日" :maxlength="255" />
        </a-form-item>
        <div class="form-row">
          <a-form-item label="时长（分钟）">
            <a-input-number v-model:value="form.duration_min" :min="0" style="width: 100%" />
          </a-form-item>
          <a-form-item label="消耗（千卡）">
            <a-input-number v-model:value="form.calories" :min="0" style="width: 100%" />
          </a-form-item>
        </div>
        <a-form-item label="动作清单">
          <div class="dynamic-list">
            <div v-for="(item, index) in form.content" :key="index" class="dynamic-row">
              <a-input v-model:value="item.name" placeholder="动作，如：卧推" />
              <a-input v-model:value="item.sets" placeholder="组数，如：4" style="width: 90px" />
              <a-input v-model:value="item.reps" placeholder="次数，如：8-12" style="width: 110px" />
              <a-button type="text" danger @click="form.content.splice(index, 1)">
                <DeleteOutlined />
              </a-button>
            </div>
            <a-button type="dashed" block @click="form.content.push({ name: '', sets: '', reps: '' })">
              <PlusOutlined /> 添加动作
            </a-button>
          </div>
        </a-form-item>
        <a-form-item label="备注">
          <a-textarea v-model:value="form.notes" placeholder="训练感受、重量记录等" :rows="3" />
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
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { fitnessApi } from '@/api/fitness'
import type { FitnessRecord, FitnessExercise } from '@/types/fitness'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const keyword = ref('')
const typeFilter = ref<string | undefined>(undefined)
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
  type: 'strength' as FitnessRecord['type'],
  duration_min: 0,
  calories: 0,
  content: [] as FitnessExercise[],
  notes: '',
  status: 0,
})

const columns = [
  { key: 'title', title: '训练', ellipsis: true },
  { key: 'type', title: '类型', width: 90 },
  { key: 'duration_min', title: '时长', width: 100 },
  { key: 'calories', title: '消耗', width: 100 },
  { key: 'status', title: '发布', width: 90 },
  { key: 'action', title: '操作', width: 200 },
]

const paginationConfig = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (t: number) => `共 ${t} 条`,
}))

watch(keyword, setDebounce)
watch([debouncedKeyword, typeFilter], () => {
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
      type: typeFilter.value || undefined,
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

function formatDate(d: string): string {
  const date = new Date(d)
  return Number.isNaN(date.getTime()) ? d : date.toLocaleDateString('zh-CN')
}

function typeLabel(t: string): string {
  return { strength: '力量', cardio: '有氧', stretch: '拉伸' }[t] || t
}

function typeColor(t: string): string {
  return { strength: 'volcano', cardio: 'geekblue', stretch: 'green' }[t] || 'default'
}

function resetForm() {
  trainDate.value = dayjs()
  form.title = ''
  form.type = 'strength'
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
  form.type = record.type
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
    type: form.type,
    duration_min: form.duration_min,
    calories: form.calories,
    content: form.content.filter((c) => c.name.trim()),
    notes: form.notes.trim() || undefined,
    status: form.status,
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning('请输入训练标题')
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await fitnessApi.create(buildPayload())
      message.success('创建成功')
    } else {
      await fitnessApi.update(editingId.value, buildPayload())
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

async function toggleStatus(record: FitnessRecord) {
  try {
    await fitnessApi.update(record.id, { status: record.status === 1 ? 0 : 1 })
    message.success(record.status === 1 ? '已下架' : '已发布')
    fetchList()
  } catch {
    // 错误由拦截器处理
  }
}

function handleDelete(record: FitnessRecord) {
  Modal.confirm({
    title: '确认删除',
    content: `删除「${record.title}」后无法恢复，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await fitnessApi.remove(record.id)
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
  color: #94a3b8;
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
