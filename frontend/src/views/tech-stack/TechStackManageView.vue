<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ t('techStack.list.title') }}</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> {{ t('techStack.list.addTech') }}
      </a-button>
    </div>

    <!-- 分类标签栏 -->
    <CategoryBar type="tech_stack" @select="handleCategorySelect" />

    <a-card>
      <div class="toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="t('techStack.list.searchPlaceholder')"
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
          <template v-if="column.key === 'name'">
            <div class="title-cell">
              <span class="name-text">{{ record.name }}</span>
              <a-tag v-if="record.category_name">{{ record.category_name }}</a-tag>
              <a-tag v-else-if="record.category">{{ categoryLabel(record.category) }}</a-tag>
            </div>
          </template>
          <template v-else-if="column.key === 'level'">
            <a-tag :color="levelColor(record.level)">{{ levelLabel(record.level) }}</a-tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 1 ? 'success' : 'default'">
              {{ record.status === 1 ? t('techStack.status.published') : t('techStack.status.draft') }}
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
      :title="modalMode === 'create' ? t('techStack.list.addTech') : t('techStack.list.editTech')"
      :confirm-loading="modalLoading"
      width="560px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item :label="t('common.field.name')" required>
          <a-input v-model:value="form.name" :placeholder="t('techStack.form.namePlaceholder')" :maxlength="100" />
        </a-form-item>
        <div class="form-row">
          <a-form-item :label="t('common.field.category')">
            <CategorySelect v-model="form.category_id" type="tech_stack" :placeholder="t('techStack.form.categoryPlaceholder')" />
          </a-form-item>
          <a-form-item :label="t('techStack.form.levelLabel')">
            <a-select v-model:value="form.level">
              <a-select-option :value="1">{{ t('techStack.level.aware') }}</a-select-option>
              <a-select-option :value="2">{{ t('techStack.level.familiar') }}</a-select-option>
              <a-select-option :value="3">{{ t('techStack.level.proficient') }}</a-select-option>
              <a-select-option :value="4">{{ t('techStack.level.expert') }}</a-select-option>
            </a-select>
          </a-form-item>
        </div>
        <a-form-item :label="t('common.field.icon')">
          <ImageField
            v-model="form.icon"
            v-model:media-id="form.iconMediaId"
            type="tech"
            :search-keyword="form.name"
            :placeholder="t('techStack.form.iconPlaceholder')"
          />
        </a-form-item>
        <a-form-item :label="t('common.field.description')">
          <a-textarea v-model:value="form.description" :placeholder="t('techStack.form.descPlaceholder')" :rows="3" :maxlength="500" />
        </a-form-item>
        <a-form-item :label="t('common.field.status')">
          <a-radio-group v-model:value="form.status">
            <a-radio :value="0">{{ t('techStack.status.draft') }}</a-radio>
            <a-radio :value="1">{{ t('common.actions.publish') }}</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { techStackApi } from '@/api/tech-stack'
import type { TechStackItem } from '@/types/tech-stack'
import ImageField from '@/components/media/ImageField.vue'
import CategoryBar from '@/components/common/CategoryBar.vue'
import CategorySelect from '@/components/article/CategorySelect.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const { t } = useI18n()

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
  category_id: undefined as string | undefined,
  icon: '',
  iconMediaId: '',
  level: 2,
  description: '',
  status: 0,
})

const columns = computed(() => [
  { key: 'name', title: t('techStack.columns.tech'), ellipsis: true },
  { key: 'level', title: t('techStack.columns.level'), width: 100 },
  { key: 'description', title: t('common.field.description'), ellipsis: true },
  { key: 'status', title: t('techStack.columns.publish'), width: 90 },
  { key: 'action', title: t('common.table.actions'), width: 200 },
])

const paginationConfig = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (count: number) => t('techStack.list.total', { total: count }),
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
    const res = await techStackApi.getList({
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

const categoryLabels = computed<Record<string, string>>(() => ({
  language: t('techStack.category.language'),
  framework: t('techStack.category.framework'),
  tool: t('techStack.category.tool'),
  database: t('techStack.category.database'),
  other: t('techStack.category.other'),
}))

function categoryLabel(c: string): string {
  return categoryLabels.value[c] || c
}

const levelLabels = computed<Record<number, string>>(() => ({
  1: t('techStack.level.aware'),
  2: t('techStack.level.familiar'),
  3: t('techStack.level.proficient'),
  4: t('techStack.level.expert'),
}))

function levelLabel(l: number): string {
  return levelLabels.value[l] || t('techStack.level.aware')
}

function levelColor(l: number): string {
  return { 1: 'default', 2: 'blue', 3: 'gold', 4: 'volcano' }[l] || 'default'
}

function resetForm() {
  form.name = ''
  form.category_id = undefined
  form.icon = ''
  form.iconMediaId = ''
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
  form.category_id = record.category_id || undefined
  form.icon = record.icon || ''
  form.iconMediaId = record.icon_media_id || ''
  form.level = record.level || 1
  form.description = record.description || ''
  form.status = record.status
  modalVisible.value = true
}

function buildPayload() {
  return {
    name: form.name.trim(),
    category_id: form.category_id,
    icon: form.icon.trim() || undefined,
    icon_media_id: form.iconMediaId || undefined,
    level: form.level,
    description: form.description.trim() || undefined,
    status: form.status,
  }
}

async function handleSubmit() {
  if (!form.name.trim()) {
    message.warning(t('techStack.form.nameRequired'))
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await techStackApi.create(buildPayload())
      message.success(t('common.message.createSuccess'))
    } else {
      await techStackApi.update(editingId.value, buildPayload())
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

async function toggleStatus(record: TechStackItem) {
  try {
    await techStackApi.update(record.id, { status: record.status === 1 ? 0 : 1 })
    message.success(record.status === 1 ? t('techStack.message.offlineSuccess') : t('techStack.message.publishSuccess'))
    fetchList()
  } catch {
    // 错误由拦截器处理
  }
}

function handleDelete(record: TechStackItem) {
  Modal.confirm({
    title: t('techStack.message.deleteTitle'),
    content: t('techStack.message.deleteConfirm', { name: record.name }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await techStackApi.remove(record.id)
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
