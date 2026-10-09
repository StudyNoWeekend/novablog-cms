<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">读书书架</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> 添加书籍
      </a-button>
    </div>

    <a-card>
      <div class="toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索书名/作者..."
          style="width: 240px"
          allow-clear
        />
        <a-select v-model:value="readingStatusFilter" style="width: 140px" allow-clear>
          <a-select-option value="want">想读</a-select-option>
          <a-select-option value="reading">在读</a-select-option>
          <a-select-option value="done">读完</a-select-option>
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
            <div class="title-cell">
              <img v-if="record.cover" :src="record.cover" class="cover-thumb" referrerpolicy="no-referrer" />
              <div v-else class="cover-thumb placeholder"><ReadOutlined /></div>
              <div>
                <div class="title-text">{{ record.title }}</div>
                <div class="sub-text">{{ record.author || '未知作者' }}</div>
              </div>
            </div>
          </template>
          <template v-else-if="column.key === 'rating'">
            <a-rate :value="record.rating" disabled :count="5" style="font-size: 14px" />
          </template>
          <template v-else-if="column.key === 'reading_status'">
            <a-tag :color="statusColor(record.reading_status)">{{ statusLabel(record.reading_status) }}</a-tag>
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
      :title="modalMode === 'create' ? '添加书籍' : '编辑书籍'"
      :confirm-loading="modalLoading"
      width="640px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item label="书名" required>
          <a-input v-model:value="form.title" placeholder="请输入书名" :maxlength="255" />
        </a-form-item>
        <div class="form-row">
          <a-form-item label="作者">
            <a-input v-model:value="form.author" placeholder="请输入作者" :maxlength="255" />
          </a-form-item>
          <a-form-item label="阅读状态">
            <a-select v-model:value="form.reading_status">
              <a-select-option value="want">想读</a-select-option>
              <a-select-option value="reading">在读</a-select-option>
              <a-select-option value="done">读完</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="评分">
            <a-rate v-model:value="form.rating" :count="5" />
          </a-form-item>
        </div>
        <a-form-item label="封面图">
          <MediaPicker v-model:visible="mediaPickerVisible" module="book" @selected="handleMediaSelected" />
          <ImageField v-model="form.cover" type="book" :search-keyword="form.title" placeholder="或直接粘贴封面地址">
            <template #extra>
              <a-button size="small" @click="mediaPickerVisible = true">
                <PictureOutlined /> 媒体库
              </a-button>
            </template>
          </ImageField>
        </a-form-item>
        <div class="form-row">
          <a-form-item label="开始阅读">
            <a-date-picker v-model:value="startedAt" style="width: 100%" />
          </a-form-item>
          <a-form-item label="读完">
            <a-date-picker v-model:value="finishedAt" style="width: 100%" />
          </a-form-item>
        </div>
        <a-form-item label="书评">
          <a-textarea v-model:value="form.review" placeholder="写点什么（Markdown）" :rows="5" />
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
import { PlusOutlined, PictureOutlined, ReadOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { bookApi } from '@/api/book'
import type { Book } from '@/types/book'
import MediaPicker from '@/components/media/MediaPicker.vue'
import ImageField from '@/components/media/ImageField.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const keyword = ref('')
const readingStatusFilter = ref<string | undefined>(undefined)
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const list = ref<Book[]>([])
const loading = ref(false)
const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(10)

const modalVisible = ref(false)
const modalLoading = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref('')
const mediaPickerVisible = ref(false)
const startedAt = ref<Dayjs | null>(null)
const finishedAt = ref<Dayjs | null>(null)

const form = reactive({
  title: '',
  author: '',
  cover: '',
  rating: 0,
  reading_status: 'want' as Book['reading_status'],
  review: '',
  status: 0,
})

const columns = [
  { key: 'title', title: '书籍', ellipsis: true },
  { key: 'rating', title: '评分', width: 160 },
  { key: 'reading_status', title: '状态', width: 90 },
  { key: 'status', title: '发布', width: 90 },
  { key: 'updated_at', title: '更新时间', width: 170 },
  { key: 'action', title: '操作', width: 200 },
]

const paginationConfig = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (t: number) => `共 ${t} 本`,
}))

watch(keyword, setDebounce)
watch([debouncedKeyword, readingStatusFilter], () => {
  reset()
  fetchList()
})

onMounted(() => {
  fetchList()
})

async function fetchList() {
  loading.value = true
  try {
    const res = await bookApi.getList({
      page: page.value,
      page_size: pageSize.value,
      keyword: debouncedKeyword.value || undefined,
      reading_status: readingStatusFilter.value || undefined,
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

function statusLabel(s: string): string {
  return { want: '想读', reading: '在读', done: '读完' }[s] || s
}

function statusColor(s: string): string {
  return { want: 'default', reading: 'processing', done: 'success' }[s] || 'default'
}

function resetForm() {
  form.title = ''
  form.author = ''
  form.cover = ''
  form.rating = 0
  form.reading_status = 'want'
  form.review = ''
  form.status = 0
  startedAt.value = null
  finishedAt.value = null
}

function openCreateModal() {
  modalMode.value = 'create'
  editingId.value = ''
  resetForm()
  modalVisible.value = true
}

function handleEdit(record: Book) {
  modalMode.value = 'edit'
  editingId.value = record.id
  resetForm()
  form.title = record.title
  form.author = record.author || ''
  form.cover = record.cover || ''
  form.rating = record.rating || 0
  form.reading_status = record.reading_status
  form.review = record.review || ''
  form.status = record.status
  startedAt.value = record.started_at ? dayjs(record.started_at) : null
  finishedAt.value = record.finished_at ? dayjs(record.finished_at) : null
  modalVisible.value = true
}

function handleMediaSelected(media: { url: string }) {
  form.cover = media.url
}

function buildPayload() {
  return {
    title: form.title.trim(),
    author: form.author.trim() || undefined,
    cover: form.cover.trim() || undefined,
    rating: form.rating,
    reading_status: form.reading_status,
    review: form.review.trim() || undefined,
    started_at: startedAt.value ? startedAt.value.toISOString() : undefined,
    finished_at: finishedAt.value ? finishedAt.value.toISOString() : undefined,
    status: form.status,
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning('请输入书名')
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await bookApi.create(buildPayload())
      message.success('创建成功')
    } else {
      await bookApi.update(editingId.value, buildPayload())
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

async function toggleStatus(record: Book) {
  try {
    await bookApi.update(record.id, { status: record.status === 1 ? 0 : 1 })
    message.success(record.status === 1 ? '已下架' : '已发布')
    fetchList()
  } catch {
    // 错误由拦截器处理
  }
}

function handleDelete(record: Book) {
  Modal.confirm({
    title: '确认删除',
    content: `删除「${record.title}」后无法恢复，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await bookApi.remove(record.id)
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
  gap: 12px;
}

.cover-thumb {
  width: 48px;
  height: 64px;
  object-fit: cover;
  border-radius: 4px;
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
</style>
