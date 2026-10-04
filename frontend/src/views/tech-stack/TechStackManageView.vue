<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">技术栈</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> 添加技术
      </a-button>
    </div>

    <a-card>
      <div class="toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索技术..."
          style="width: 240px"
          allow-clear
        />
        <a-select v-model:value="categoryFilter" style="width: 160px" allow-clear>
          <a-select-option value="language">语言</a-select-option>
          <a-select-option value="framework">框架</a-select-option>
          <a-select-option value="tool">工具</a-select-option>
          <a-select-option value="database">数据库</a-select-option>
          <a-select-option value="other">其他</a-select-option>
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
          <template v-if="column.key === 'name'">
            <div class="title-cell">
              <span class="name-text">{{ record.name }}</span>
              <a-tag v-if="record.category">{{ categoryLabel(record.category) }}</a-tag>
            </div>
          </template>
          <template v-else-if="column.key === 'level'">
            <a-tag :color="levelColor(record.level)">{{ levelLabel(record.level) }}</a-tag>
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
      :title="modalMode === 'create' ? '添加技术' : '编辑技术'"
      :confirm-loading="modalLoading"
      width="560px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item label="名称" required>
          <a-input v-model:value="form.name" placeholder="如：Go / Vue 3 / PostgreSQL" :maxlength="100" />
        </a-form-item>
        <div class="form-row">
          <a-form-item label="分类">
            <a-select v-model:value="form.category" allow-clear>
              <a-select-option value="language">语言</a-select-option>
              <a-select-option value="framework">框架</a-select-option>
              <a-select-option value="tool">工具</a-select-option>
              <a-select-option value="database">数据库</a-select-option>
              <a-select-option value="other">其他</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="熟练度">
            <a-select v-model:value="form.level">
              <a-select-option :value="1">了解</a-select-option>
              <a-select-option :value="2">熟悉</a-select-option>
              <a-select-option :value="3">熟练</a-select-option>
              <a-select-option :value="4">精通</a-select-option>
            </a-select>
          </a-form-item>
        </div>
        <a-form-item label="图标">
          <ImageField
            v-model="form.icon"
            type="tech"
            :search-keyword="form.name"
            placeholder="或直接粘贴图标地址"
          />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea v-model:value="form.description" placeholder="使用经验、代表作品等" :rows="3" :maxlength="500" />
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
import { PlusOutlined } from '@ant-design/icons-vue'
import { techStackApi } from '@/api/tech-stack'
import type { TechStackItem } from '@/types/tech-stack'
import ImageField from '@/components/media/ImageField.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const keyword = ref('')
const categoryFilter = ref<string | undefined>(undefined)
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const list = ref<TechStackItem[]>([])
const loading = ref(false)
const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(10)

const modalVisible = ref(false)
const modalLoading = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref('')

const form = reactive({
  name: '',
  category: undefined as string | undefined,
  icon: '',
  level: 2,
  description: '',
  status: 0,
})

const columns = [
  { key: 'name', title: '技术', ellipsis: true },
  { key: 'level', title: '熟练度', width: 100 },
  { key: 'description', title: '描述', ellipsis: true },
  { key: 'status', title: '发布', width: 90 },
  { key: 'action', title: '操作', width: 200 },
]

const paginationConfig = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (t: number) => `共 ${t} 项`,
}))

watch(keyword, setDebounce)
watch([debouncedKeyword, categoryFilter], () => {
  reset()
  fetchList()
})

onMounted(() => {
  fetchList()
})

async function fetchList() {
  loading.value = true
  try {
    const res = await techStackApi.getList({
      page: page.value,
      page_size: pageSize.value,
      keyword: debouncedKeyword.value || undefined,
      category: categoryFilter.value || undefined,
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

function categoryLabel(c: string): string {
  return { language: '语言', framework: '框架', tool: '工具', database: '数据库', other: '其他' }[c] || c
}

function levelLabel(l: number): string {
  return { 1: '了解', 2: '熟悉', 3: '熟练', 4: '精通' }[l] || '了解'
}

function levelColor(l: number): string {
  return { 1: 'default', 2: 'blue', 3: 'gold', 4: 'volcano' }[l] || 'default'
}

function resetForm() {
  form.name = ''
  form.category = undefined
  form.icon = ''
  form.level = 2
  form.description = ''
  form.status = 0
}

function openCreateModal() {
  modalMode.value = 'create'
  editingId.value = ''
  resetForm()
  modalVisible.value = true
}

function handleEdit(record: TechStackItem) {
  modalMode.value = 'edit'
  editingId.value = record.id
  resetForm()
  form.name = record.name
  form.category = record.category || undefined
  form.icon = record.icon || ''
  form.level = record.level || 1
  form.description = record.description || ''
  form.status = record.status
  modalVisible.value = true
}

function buildPayload() {
  return {
    name: form.name.trim(),
    category: form.category || undefined,
    icon: form.icon.trim() || undefined,
    level: form.level,
    description: form.description.trim() || undefined,
    status: form.status,
  }
}

async function handleSubmit() {
  if (!form.name.trim()) {
    message.warning('请输入技术名称')
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await techStackApi.create(buildPayload())
      message.success('创建成功')
    } else {
      await techStackApi.update(editingId.value, buildPayload())
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

async function toggleStatus(record: TechStackItem) {
  try {
    await techStackApi.update(record.id, { status: record.status === 1 ? 0 : 1 })
    message.success(record.status === 1 ? '已下架' : '已发布')
    fetchList()
  } catch {
    // 错误由拦截器处理
  }
}

function handleDelete(record: TechStackItem) {
  Modal.confirm({
    title: '确认删除',
    content: `删除「${record.name}」后无法恢复，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await techStackApi.remove(record.id)
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

.title-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.name-text {
  font-weight: 500;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
</style>
