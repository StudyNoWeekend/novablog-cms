<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">游戏库</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> 添加游戏
      </a-button>
    </div>

    <a-card>
      <div class="toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索游戏..."
          style="width: 240px"
          allow-clear
        />
        <a-select v-model:value="playStatusFilter" style="width: 140px" allow-clear>
          <a-select-option value="want">想玩</a-select-option>
          <a-select-option value="playing">在玩</a-select-option>
          <a-select-option value="played">玩过</a-select-option>
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
                <div class="sub-text">{{ [record.platform, record.genre].filter(Boolean).join(' · ') || '—' }}</div>
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
      :title="modalMode === 'create' ? '添加游戏' : '编辑游戏'"
      :confirm-loading="modalLoading"
      width="640px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item label="游戏名称" required>
          <a-input v-model:value="form.title" placeholder="请输入游戏名称" :maxlength="255" />
        </a-form-item>
        <div class="form-row">
          <a-form-item label="平台">
            <a-input v-model:value="form.platform" placeholder="如：PC / PS5 / Switch" :maxlength="100" />
          </a-form-item>
          <a-form-item label="类型">
            <a-input v-model:value="form.genre" placeholder="如：RPG / FPS" :maxlength="100" />
          </a-form-item>
          <a-form-item label="游玩状态">
            <a-select v-model:value="form.play_status">
              <a-select-option value="want">想玩</a-select-option>
              <a-select-option value="playing">在玩</a-select-option>
              <a-select-option value="played">玩过</a-select-option>
            </a-select>
          </a-form-item>
        </div>
        <div class="form-row">
          <a-form-item label="时长（小时）">
            <a-input-number v-model:value="form.play_hours" :min="0" style="width: 100%" />
          </a-form-item>
          <a-form-item label="评分（0-10）">
            <a-slider v-model:value="form.rating" :min="0" :max="10" style="margin-top: 8px" />
          </a-form-item>
        </div>
        <a-form-item label="封面图">
          <MediaPicker v-model:visible="mediaPickerVisible" module="game" @selected="handleMediaSelected" />
          <ImageField v-model="form.cover" type="game" :search-keyword="form.title" placeholder="或直接粘贴封面地址">
            <template #extra>
              <a-button size="small" @click="mediaPickerVisible = true">
                <PictureOutlined /> 媒体库
              </a-button>
            </template>
          </ImageField>
        </a-form-item>
        <a-form-item label="短评">
          <a-textarea v-model:value="form.short_review" placeholder="一句话短评" :rows="3" :maxlength="1000" />
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
import { PlusOutlined, PictureOutlined, RocketOutlined } from '@ant-design/icons-vue'
import { gameApi } from '@/api/game'
import type { Game } from '@/types/game'
import MediaPicker from '@/components/media/MediaPicker.vue'
import ImageField from '@/components/media/ImageField.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const keyword = ref('')
const playStatusFilter = ref<string | undefined>(undefined)
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
  status: 0,
})

const columns = [
  { key: 'title', title: '游戏', ellipsis: true },
  { key: 'play_status', title: '状态', width: 90 },
  { key: 'play_hours', title: '时长', width: 80 },
  { key: 'rating', title: '评分', width: 90 },
  { key: 'status', title: '发布', width: 90 },
  { key: 'action', title: '操作', width: 200 },
]

const paginationConfig = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (t: number) => `共 ${t} 款`,
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
  return { want: '想玩', playing: '在玩', played: '玩过' }[s] || s
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
    status: form.status,
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning('请输入游戏名称')
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await gameApi.create(buildPayload())
      message.success('创建成功')
    } else {
      await gameApi.update(editingId.value, buildPayload())
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

async function toggleStatus(record: Game) {
  try {
    await gameApi.update(record.id, { status: record.status === 1 ? 0 : 1 })
    message.success(record.status === 1 ? '已下架' : '已发布')
    fetchList()
  } catch {
    // 错误由拦截器处理
  }
}

function handleDelete(record: Game) {
  Modal.confirm({
    title: '确认删除',
    content: `删除「${record.title}」后无法恢复，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await gameApi.remove(record.id)
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

.sub-text {
  font-size: 12px;
  color: #94a3b8;
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
</style>
