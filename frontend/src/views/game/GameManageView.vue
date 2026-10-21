<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ t('menu.game.list') }}</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> {{ t('game.actions.create') }}
      </a-button>
    </div>

    <!-- 分类标签栏 -->
    <CategoryBar type="game" @select="handleCategorySelect" />

    <a-card>
      <div class="toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="t('game.list.searchPlaceholder')"
          style="width: 240px"
          allow-clear
        />
        <a-select v-model:value="playStatusFilter" style="width: 140px" allow-clear>
          <a-select-option value="want">{{ t('game.playStatus.want') }}</a-select-option>
          <a-select-option value="playing">{{ t('game.playStatus.playing') }}</a-select-option>
          <a-select-option value="played">{{ t('game.playStatus.played') }}</a-select-option>
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
              <div v-else class="cover-thumb placeholder"><RocketOutlined /></div>
              <div>
                <div class="title-text">{{ record.title }}</div>
                <div class="sub-text">
                  {{ [record.category_name, record.platform, record.genre].filter(Boolean).join(' · ') || '—' }}
                </div>
              </div>
            </div>
          </template>
          <template v-else-if="column.key === 'play_status'">
            <a-tag :color="statusColor(record.play_status)">{{ statusLabel(record.play_status) }}</a-tag>
          </template>
          <template v-else-if="column.key === 'play_hours'">
            {{ record.play_hours }}h
          </template>
          <template v-else-if="column.key === 'rating'">
            {{ record.rating > 0 ? record.rating + ' / 10' : '—' }}
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 1 ? 'success' : 'default'">
              {{ record.status === 1 ? t('game.status.published') : t('game.status.draft') }}
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
      :title="modalMode === 'create' ? t('game.actions.create') : t('game.actions.edit')"
      :confirm-loading="modalLoading"
      width="640px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item :label="t('game.form.title')" required>
          <a-input v-model:value="form.title" :placeholder="t('game.form.titlePlaceholder')" :maxlength="255" />
        </a-form-item>
        <div class="form-row">
          <a-form-item :label="t('game.form.platform')">
            <a-input v-model:value="form.platform" :placeholder="t('game.form.platformPlaceholder')" :maxlength="100" />
          </a-form-item>
          <a-form-item :label="t('common.field.type')">
            <a-input v-model:value="form.genre" :placeholder="t('game.form.genrePlaceholder')" :maxlength="100" />
          </a-form-item>
          <a-form-item :label="t('game.form.playStatus')">
            <a-select v-model:value="form.play_status">
            <a-select-option value="want">{{ t('game.playStatus.want') }}</a-select-option>
            <a-select-option value="playing">{{ t('game.playStatus.playing') }}</a-select-option>
            <a-select-option value="played">{{ t('game.playStatus.played') }}</a-select-option>
            </a-select>
          </a-form-item>
        </div>
        <div class="form-row">
          <a-form-item :label="t('common.field.category')">
            <CategorySelect v-model="form.category_id" type="game" :placeholder="t('game.form.categoryPlaceholder')" />
          </a-form-item>
          <a-form-item :label="t('game.form.playHours')">
            <a-input-number v-model:value="form.play_hours" :min="0" style="width: 100%" />
          </a-form-item>
          <a-form-item :label="t('game.form.rating')">
            <a-slider v-model:value="form.rating" :min="0" :max="10" style="margin-top: 8px" />
          </a-form-item>
        </div>
        <a-form-item :label="t('game.form.cover')">
          <MediaPicker v-model:visible="mediaPickerVisible" module="game" @selected="handleMediaSelected" />
          <ImageField v-model="form.cover" type="game" with-preset :search-keyword="form.title" :placeholder="t('game.form.coverUrlPlaceholder')">
            <template #extra>
              <a-button size="small" @click="mediaPickerVisible = true">
                <PictureOutlined /> {{ t('game.form.mediaLibrary') }}
              </a-button>
            </template>
          </ImageField>
        </a-form-item>
        <a-form-item :label="t('game.form.shortReview')">
          <a-textarea v-model:value="form.short_review" :placeholder="t('game.form.shortReviewPlaceholder')" :rows="3" :maxlength="1000" />
        </a-form-item>
        <a-form-item :label="t('common.field.status')">
          <a-radio-group v-model:value="form.status">
            <a-radio :value="0">{{ t('game.status.draft') }}</a-radio>
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
import { PlusOutlined, PictureOutlined, RocketOutlined } from '@ant-design/icons-vue'
import { gameApi } from '@/api/game'
import type { Game } from '@/types/game'
import MediaPicker from '@/components/media/MediaPicker.vue'
import ImageField from '@/components/media/ImageField.vue'
import CategoryBar from '@/components/common/CategoryBar.vue'
import CategorySelect from '@/components/article/CategorySelect.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const { t } = useI18n()

const keyword = ref('')
const playStatusFilter = ref<string | undefined>(undefined)
const categoryFilter = ref<string | undefined>(undefined)
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const list = ref<Game[]>([])
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
  platform: '',
  genre: '',
  play_status: 'want' as Game['play_status'],
  play_hours: 0,
  rating: 0,
  short_review: '',
  category_id: undefined as string | undefined,
  status: 0,
})

const columns = computed(() => [
  { key: 'title', title: t('game.columns.game'), ellipsis: true },
  { key: 'play_status', title: t('common.field.status'), width: 90 },
  { key: 'play_hours', title: t('game.columns.playHours'), width: 80 },
  { key: 'rating', title: t('game.columns.rating'), width: 90 },
  { key: 'status', title: t('game.columns.publish'), width: 90 },
  { key: 'action', title: t('common.table.actions'), width: 200 },
])

const paginationConfig = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (count: number) => t('game.list.total', { total: count }),
}))

watch(keyword, setDebounce)
watch([debouncedKeyword, playStatusFilter], () => {
  reset()
  fetchList()
})

onMounted(() => {
  fetchList()
})

async function fetchList() {
  loading.value = true
  try {
    const res = await gameApi.getList({
      page: page.value,
      page_size: pageSize.value,
      keyword: debouncedKeyword.value || undefined,
      play_status: playStatusFilter.value || undefined,
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

const PLAY_STATUS_LABEL_KEYS: Record<string, string> = {
  want: 'game.playStatus.want',
  playing: 'game.playStatus.playing',
  played: 'game.playStatus.played',
}

function statusLabel(s: string): string {
  const key = PLAY_STATUS_LABEL_KEYS[s]
  return key ? t(key) : s
}

function statusColor(s: string): string {
  return { want: 'default', playing: 'processing', played: 'success' }[s] || 'default'
}

function resetForm() {
  form.title = ''
  form.cover = ''
  form.platform = ''
  form.genre = ''
  form.play_status = 'want'
  form.play_hours = 0
  form.rating = 0
  form.short_review = ''
  form.category_id = undefined
  form.status = 0
}

function openCreateModal() {
  modalMode.value = 'create'
  editingId.value = ''
  resetForm()
  modalVisible.value = true
}

function handleEdit(record: Game) {
  modalMode.value = 'edit'
  editingId.value = record.id
  resetForm()
  form.title = record.title
  form.cover = record.cover || ''
  form.platform = record.platform || ''
  form.genre = record.genre || ''
  form.play_status = record.play_status
  form.play_hours = record.play_hours || 0
  form.rating = record.rating || 0
  form.short_review = record.short_review || ''
  form.category_id = record.category_id || undefined
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
    platform: form.platform.trim() || undefined,
    genre: form.genre.trim() || undefined,
    play_status: form.play_status,
    play_hours: form.play_hours,
    rating: form.rating,
    short_review: form.short_review.trim() || undefined,
    category_id: form.category_id,
    status: form.status,
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning(t('game.message.titleRequired'))
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await gameApi.create(buildPayload())
      message.success(t('common.message.createSuccess'))
    } else {
      await gameApi.update(editingId.value, buildPayload())
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

async function toggleStatus(record: Game) {
  try {
    await gameApi.update(record.id, { status: record.status === 1 ? 0 : 1 })
    message.success(record.status === 1 ? t('game.status.unpublished') : t('game.status.published'))
    fetchList()
  } catch {
    // 错误由拦截器处理
  }
}

function handleDelete(record: Game) {
  Modal.confirm({
    title: t('game.message.deleteTitle'),
    content: t('game.message.deleteConfirm', { name: record.title }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await gameApi.remove(record.id)
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
