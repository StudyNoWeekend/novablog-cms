<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ t('menu.book.list') }}</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> {{ t('book.actions.create') }}
      </a-button>
    </div>

    <!-- 分类标签栏 -->
    <CategoryBar type="book" @select="handleCategorySelect" />

    <a-card>
      <div class="toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="t('book.list.searchPlaceholder')"
          style="width: 240px"
          allow-clear
        />
        <a-select v-model:value="readingStatusFilter" style="width: 140px" allow-clear>
          <a-select-option value="want">{{ t('book.readingStatus.want') }}</a-select-option>
          <a-select-option value="reading">{{ t('book.readingStatus.reading') }}</a-select-option>
          <a-select-option value="done">{{ t('book.readingStatus.done') }}</a-select-option>
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
                <div class="sub-text">
                  {{ [record.author, record.category_name].filter(Boolean).join(' · ') || t('book.list.unknownAuthor') }}
                </div>
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
              {{ record.status === 1 ? t('book.status.published') : t('book.status.draft') }}
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
      :title="modalMode === 'create' ? t('book.actions.create') : t('book.actions.edit')"
      :confirm-loading="modalLoading"
      width="640px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item :label="t('book.form.title')" required>
          <a-input v-model:value="form.title" :placeholder="t('book.form.titlePlaceholder')" :maxlength="255" />
        </a-form-item>
        <div class="form-row">
          <a-form-item :label="t('book.form.author')">
            <a-input v-model:value="form.author" :placeholder="t('book.form.authorPlaceholder')" :maxlength="255" />
          </a-form-item>
          <a-form-item :label="t('book.form.readingStatus')">
            <a-select v-model:value="form.reading_status">
            <a-select-option value="want">{{ t('book.readingStatus.want') }}</a-select-option>
            <a-select-option value="reading">{{ t('book.readingStatus.reading') }}</a-select-option>
            <a-select-option value="done">{{ t('book.readingStatus.done') }}</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item :label="t('book.form.rating')">
            <a-rate v-model:value="form.rating" :count="5" />
          </a-form-item>
        </div>
        <div class="form-row">
          <a-form-item :label="t('common.field.category')">
            <CategorySelect v-model="form.category_id" type="book" :placeholder="t('book.form.categoryPlaceholder')" />
          </a-form-item>
        </div>
        <a-form-item :label="t('book.form.cover')">
          <MediaPicker v-model:visible="mediaPickerVisible" module="book" @selected="handleMediaSelected" />
          <ImageField v-model="form.cover" type="book" with-preset :search-keyword="form.title" :placeholder="t('book.form.coverUrlPlaceholder')">
            <template #extra>
              <a-button size="small" @click="mediaPickerVisible = true">
                <PictureOutlined /> {{ t('book.form.mediaLibrary') }}
              </a-button>
            </template>
          </ImageField>
        </a-form-item>
        <div class="form-row">
          <a-form-item :label="t('book.form.startedAt')">
            <a-date-picker v-model:value="startedAt" style="width: 100%" />
          </a-form-item>
          <a-form-item :label="t('book.form.finishedAt')">
            <a-date-picker v-model:value="finishedAt" style="width: 100%" />
          </a-form-item>
        </div>
        <a-form-item :label="t('book.form.review')">
          <a-textarea v-model:value="form.review" :placeholder="t('book.form.reviewPlaceholder')" :rows="5" />
        </a-form-item>
        <a-form-item :label="t('common.field.status')">
          <a-radio-group v-model:value="form.status">
            <a-radio :value="0">{{ t('book.status.draft') }}</a-radio>
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
import { PlusOutlined, PictureOutlined, ReadOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { bookApi } from '@/api/book'
import type { Book } from '@/types/book'
import MediaPicker from '@/components/media/MediaPicker.vue'
import ImageField from '@/components/media/ImageField.vue'
import CategoryBar from '@/components/common/CategoryBar.vue'
import CategorySelect from '@/components/article/CategorySelect.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const { t } = useI18n()

const keyword = ref('')
const readingStatusFilter = ref<string | undefined>(undefined)
const categoryFilter = ref<string | undefined>(undefined)
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
  category_id: undefined as string | undefined,
  status: 0,
})

const columns = computed(() => [
  { key: 'title', title: t('book.columns.book'), ellipsis: true },
  { key: 'rating', title: t('book.columns.rating'), width: 160 },
  { key: 'reading_status', title: t('common.field.status'), width: 90 },
  { key: 'status', title: t('book.columns.publish'), width: 90 },
  { key: 'updated_at', title: t('common.field.updateTime'), width: 170 },
  { key: 'action', title: t('common.table.actions'), width: 200 },
])

const paginationConfig = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (count: number) => t('book.list.total', { total: count }),
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

const READING_STATUS_LABEL_KEYS: Record<string, string> = {
  want: 'book.readingStatus.want',
  reading: 'book.readingStatus.reading',
  done: 'book.readingStatus.done',
}

function statusLabel(s: string): string {
  const key = READING_STATUS_LABEL_KEYS[s]
  return key ? t(key) : s
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
  form.category_id = undefined
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
  form.category_id = record.category_id || undefined
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
    category_id: form.category_id,
    status: form.status,
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning(t('book.message.titleRequired'))
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await bookApi.create(buildPayload())
      message.success(t('common.message.createSuccess'))
    } else {
      await bookApi.update(editingId.value, buildPayload())
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

async function toggleStatus(record: Book) {
  try {
    await bookApi.update(record.id, { status: record.status === 1 ? 0 : 1 })
    message.success(record.status === 1 ? t('book.status.unpublished') : t('book.status.published'))
    fetchList()
  } catch {
    // 错误由拦截器处理
  }
}

function handleDelete(record: Book) {
  Modal.confirm({
    title: t('book.message.deleteTitle'),
    content: t('book.message.deleteConfirm', { name: record.title }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await bookApi.remove(record.id)
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
