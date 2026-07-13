<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">视频作品</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> 新建视频
      </a-button>
    </div>

    <a-card>
      <div class="video-toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索视频标题..."
          style="width: 240px"
          allow-clear
        />
      </div>

      <a-spin :spinning="loading">
        <a-empty
          v-if="!loading && list.length === 0"
          description="还没有视频作品，开始添加吧"
          style="margin-top: 48px"
        />
        <div v-else class="video-grid">
          <div v-for="item in list" :key="item.id" class="video-card">
            <div class="video-cover">
              <img
                v-if="item.cover_url"
                :src="getThumbUrl(item.cover_url, 400)"
                :alt="item.title"
                referrerpolicy="no-referrer"
              />
              <div v-else class="cover-placeholder">
                <PlaySquareOutlined />
              </div>
            </div>
            <div class="video-info">
              <div class="video-title" :title="item.title">{{ item.title }}</div>
              <div class="video-desc" :title="item.description">
                {{ item.description || '暂无介绍' }}
              </div>
              <div class="video-platforms">
                <template v-if="item.platforms && item.platforms.length">
                  <a
                    v-for="link in item.platforms"
                    :key="link.id"
                    :href="link.url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="platform-link"
                    @click.stop
                  >
                    <PlatformIcon :platform="link.platform" :size="22" />
                  </a>
                </template>
                <span v-else class="no-platform">暂无平台链接</span>
              </div>
              <div class="video-meta">
                <span class="meta-item">{{ formatDateTime(item.created_at) }}</span>
              </div>
              <div class="video-actions">
                <a-button type="link" size="small" @click="handleEdit(item)">
                  <EditOutlined /> 编辑
                </a-button>
                <a-button type="link" size="small" danger @click="handleDelete(item)">
                  <DeleteOutlined /> 删除
                </a-button>
              </div>
            </div>
          </div>
        </div>
      </a-spin>

      <div v-if="total > 0" class="pagination-wrapper">
        <a-pagination
          :current="page"
          :page-size="pageSize"
          :total="total"
          :page-size-options="[12, 24, 48]"
          show-size-changer
          :show-total="(t: number) => `共 ${t} 个视频`"
          @change="handlePageChange"
          @show-size-change="handlePageChange"
        />
      </div>
    </a-card>

    <!-- 新建/编辑视频弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalMode === 'create' ? '新建视频' : '编辑视频'"
      :confirm-loading="modalLoading"
      width="640px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item label="标题" required>
          <a-input
            v-model:value="form.title"
            placeholder="请输入视频标题"
            :maxlength="100"
            show-count
          />
        </a-form-item>
        <a-form-item label="封面 URL">
          <a-input
            v-model:value="form.cover_url"
            placeholder="请输入封面图片地址（可选）"
          />
        </a-form-item>
        <a-form-item label="介绍">
          <a-textarea
            v-model:value="form.description"
            placeholder="请输入视频介绍（可选）"
            :rows="3"
            :maxlength="500"
            show-count
          />
        </a-form-item>
        <a-form-item label="发布平台" required>
          <div class="platform-selector">
            <div
              v-for="p in PLATFORMS"
              :key="p.key"
              class="platform-toggle"
              :class="{ active: p.key in selectedPlatforms }"
              :style="p.key in selectedPlatforms ? { borderColor: p.color, color: p.color } : {}"
              @click="togglePlatform(p.key)"
            >
              <PlatformIcon :platform="p.key" :size="20" />
              <span class="platform-toggle-name">{{ p.name }}</span>
            </div>
          </div>
          <div v-if="Object.keys(selectedPlatforms).length > 0" class="platform-urls">
            <div v-for="p in selectedPlatformList" :key="p.key" class="platform-url-item">
              <div class="platform-url-label">
                <PlatformIcon :platform="p.key" :size="16" />
                <span>{{ p.name }}</span>
              </div>
              <a-input-group compact class="platform-url-input-group">
                <a-input
                  v-model:value="selectedPlatforms[p.key]"
                  :placeholder="`请输入${p.name}视频链接`"
                  style="width: calc(100% - 72px)"
                >
                  <template #prefix>
                    <LinkOutlined />
                  </template>
                </a-input>
                <a-button
                  type="primary"
                  ghost
                  :loading="parsingPlatform === p.key"
                  :disabled="parsingPlatform !== '' && parsingPlatform !== p.key"
                  style="width: 72px"
                  @click="handleParse(p.key)"
                >
                  <template #icon><SyncOutlined /></template>
                  解析
                </a-button>
              </a-input-group>
            </div>
          </div>
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
  EditOutlined,
  DeleteOutlined,
  PlaySquareOutlined,
  LinkOutlined,
} from '@ant-design/icons-vue'
import { videoApi } from '@/api/video'
import type { VideoWork, ParseVideoReq } from '@/types/video'
import { getThumbUrl } from '@/utils/image'
import PlatformIcon from '@/components/video/PlatformIcon.vue'
import { PLATFORMS, PLATFORM_MAP } from '@/components/video/platforms'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const list = ref<VideoWork[]>([])
const loading = ref(false)

const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(12)

const modalVisible = ref(false)
const modalLoading = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref('')

const form = reactive({
  title: '',
  cover_url: '',
  description: '',
})

const selectedPlatforms = reactive<Record<string, string>>({})

const parsingPlatform = ref('')

const selectedPlatformList = computed(() =>
  Object.keys(selectedPlatforms)
    .map((key) => PLATFORM_MAP[key])
    .filter(Boolean),
)

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
    const res = await videoApi.getList({
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

function handlePageChange(newPage: number, newPageSize?: number) {
  changePagination(newPage, newPageSize)
  fetchList()
}

function resetForm() {
  form.title = ''
  form.cover_url = ''
  form.description = ''
  Object.keys(selectedPlatforms).forEach((key) => delete selectedPlatforms[key])
}

function openCreateModal() {
  modalMode.value = 'create'
  editingId.value = ''
  resetForm()
  modalVisible.value = true
}

function handleEdit(item: VideoWork) {
  modalMode.value = 'edit'
  editingId.value = item.id
  resetForm()
  form.title = item.title
  form.cover_url = item.cover_url || ''
  form.description = item.description || ''
  if (item.platforms && item.platforms.length) {
    item.platforms.forEach((link) => {
      selectedPlatforms[link.platform] = link.url
    })
  }
  modalVisible.value = true
}

function togglePlatform(key: string) {
  if (key in selectedPlatforms) {
    delete selectedPlatforms[key]
  } else {
    selectedPlatforms[key] = ''
  }
}

function buildPlatforms() {
  return Object.entries(selectedPlatforms)
    .filter(([, url]) => url.trim())
    .map(([platform, url]) => ({ platform, url: url.trim() }))
}

async function handleParse(platformKey: string) {
  const url = selectedPlatforms[platformKey]
  if (!url || !url.trim()) {
    message.warning('请先输入视频链接')
    return
  }
  parsingPlatform.value = platformKey
  try {
    const res = await videoApi.parse({ platform: platformKey, url: url.trim() })
    const hasExisting = form.title || form.cover_url || form.description
    const doFill = () => {
      if (res.title) form.title = res.title
      if (res.cover_url) form.cover_url = res.cover_url
      if (res.description) form.description = res.description
      message.success('解析成功')
    }
    if (hasExisting) {
      Modal.confirm({
        title: '覆盖确认',
        content: '表单中已有内容，是否用解析结果覆盖？',
        okText: '覆盖',
        cancelText: '取消',
        onOk: doFill,
      })
    } else {
      doFill()
    }
  } catch {
    // 错误由拦截器处理
  } finally {
    parsingPlatform.value = ''
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning('请输入视频标题')
    return
  }
  const platforms = buildPlatforms()
  if (platforms.length === 0) {
    message.warning('请至少添加一个平台链接')
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await videoApi.create({
        title: form.title.trim(),
        cover_url: form.cover_url.trim() || undefined,
        description: form.description.trim() || undefined,
        platforms,
      })
      message.success('创建成功')
    } else {
      await videoApi.update(editingId.value, {
        title: form.title.trim(),
        cover_url: form.cover_url.trim() || undefined,
        description: form.description.trim() || undefined,
        platforms,
      })
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

function handleDelete(item: VideoWork) {
  Modal.confirm({
    title: '确认删除',
    content: `删除「${item.title}」后无法恢复，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await videoApi.remove(item.id)
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

function formatDateTime(time?: string): string {
  if (!time) return '-'
  const d = new Date(time)
  if (Number.isNaN(d.getTime())) return time
  return d.toLocaleString('zh-CN')
}
</script>

<style scoped>
.video-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.video-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.video-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #f0f0f0);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: box-shadow 0.2s;
}

.video-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.video-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  background: var(--bg-secondary, #f5f5f5);
}

.video-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}

.video-card:hover .video-cover img {
  transform: scale(1.05);
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  color: var(--text-secondary, #bfbfbf);
}

.video-info {
  padding: 12px 16px 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.video-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #262626);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.video-desc {
  font-size: 13px;
  color: var(--text-secondary, #8c8c8c);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 39px;
}

.video-platforms {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  min-height: 24px;
}

.platform-link {
  display: inline-flex;
  align-items: center;
}

.no-platform {
  font-size: 12px;
  color: var(--text-secondary, #bfbfbf);
}

.video-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-secondary, #bfbfbf);
  margin-top: 4px;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.video-actions {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  border-top: 1px solid var(--border-color, #f0f0f0);
  margin-top: 4px;
  padding-top: 4px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}

.platform-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.platform-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid var(--border-color, #d9d9d9);
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary, #595959);
  transition: all 0.2s;
}

.platform-toggle:hover {
  border-color: var(--primary-color, #40a9ff);
}

.platform-toggle.active {
  background: rgba(24, 144, 255, 0.06);
}

.platform-toggle-name {
  white-space: nowrap;
}

.platform-urls {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.platform-url-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.platform-url-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-primary, #262626);
}

.platform-url-input-group {
  display: flex;
}
</style>
