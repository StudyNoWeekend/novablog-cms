<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ t('menu.video.list') }}</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> {{ t('video.actions.create') }}
      </a-button>
    </div>

    <!-- 分类标签栏 -->
    <CategoryBar type="video" @select="handleCategorySelect" />

    <a-card>
      <div class="video-toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="t('video.list.searchPlaceholder')"
          style="width: 240px"
          allow-clear
        />
        <a-select
          v-model:value="statusFilter"
          style="width: 120px"
          :placeholder="t('common.field.status')"
          allow-clear
          @change="reset(); fetchList()"
        >
          <a-select-option value="">{{ t('common.status.all') }}</a-select-option>
          <a-select-option :value="1">{{ t('video.status.published') }}</a-select-option>
          <a-select-option :value="0">{{ t('video.status.draft') }}</a-select-option>
        </a-select>
      </div>

      <a-spin :spinning="loading">
        <a-empty
          v-if="!loading && list.length === 0"
          :description="t('video.list.empty')"
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
              <a-tag
                v-if="item.category_name"
                color="blue"
                class="status-tag category-cover-tag"
              >{{ item.category_name }}</a-tag>
              <a-tag
                v-if="item.status === 1"
                color="green"
                class="status-tag"
              >{{ t('video.status.published') }}</a-tag>
              <a-tag
                v-else
                color="orange"
                class="status-tag"
              >{{ t('video.status.draft') }}</a-tag>
            </div>
            <div class="video-info">
              <div class="video-title" :title="item.title">{{ item.title }}</div>
              <div class="video-desc" :title="item.description">
                {{ item.description || t('video.list.noDescription') }}
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
                <span v-else class="no-platform">{{ t('video.list.noPlatforms') }}</span>
              </div>
              <div class="video-meta">
                <span class="meta-item">{{ formatDateTime(item.created_at) }}</span>
              </div>
              <div class="video-actions">
                <a-button type="link" size="small" @click="handleEdit(item)">
                  <EditOutlined /> {{ t('common.actions.edit') }}
                </a-button>
                <a-button type="link" size="small" danger @click="handleDelete(item)">
                  <DeleteOutlined /> {{ t('common.actions.delete') }}
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
          :show-total="(count: number) => t('video.list.total', { total: count })"
          @change="handlePageChange"
          @show-size-change="handlePageChange"
        />
      </div>
    </a-card>

    <!-- 新建/编辑视频弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalMode === 'create' ? t('video.actions.create') : t('video.actions.edit')"
      :confirm-loading="modalLoading"
      width="640px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item :label="t('common.field.title')" required>
          <a-input
            v-model:value="form.title"
            :placeholder="t('video.form.titlePlaceholder')"
            :maxlength="100"
            show-count
          />
        </a-form-item>
        <a-form-item :label="t('video.form.coverUrl')">
          <a-input
            v-model:value="form.cover_url"
            :placeholder="t('video.form.coverUrlPlaceholder')"
            @change="handleCoverUrlInput"
          >
            <template #addonAfter>
              <a :aria-label="t('video.form.selectCoverFromMedia')" @click.prevent="mediaPickerVisible = true">
                <PictureOutlined /> {{ t('video.form.mediaLibrary') }}
              </a>
            </template>
          </a-input>
        </a-form-item>
        <a-form-item :label="t('video.form.intro')">
          <a-textarea
            v-model:value="form.description"
            :placeholder="t('video.form.introPlaceholder')"
            :rows="3"
            :maxlength="500"
            show-count
          />
        </a-form-item>
        <a-form-item :label="t('common.field.category')">
          <CategorySelect v-model="form.category_id" type="video" :placeholder="t('video.form.categoryPlaceholder')" />
        </a-form-item>
        <a-form-item :label="t('video.form.platforms')" required>
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
              <span class="platform-toggle-name">{{ t(p.labelKey) }}</span>
            </div>
          </div>
          <div v-if="Object.keys(selectedPlatforms).length > 0" class="platform-urls">
            <div v-for="p in selectedPlatformList" :key="p.key" class="platform-url-item">
              <div class="platform-url-label">
                <PlatformIcon :platform="p.key" :size="16" />
                <span>{{ t(p.labelKey) }}</span>
              </div>
              <a-input-group compact class="platform-url-input-group">
                <a-input
                  v-model:value="selectedPlatforms[p.key]"
                  :placeholder="t('video.form.platformLinkPlaceholder', { name: t(p.labelKey) })"
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
                  {{ t('video.actions.parse') }}
                </a-button>
              </a-input-group>
            </div>
          </div>
        </a-form-item>
        <a-form-item :label="t('common.field.status')">
          <a-switch
            v-model:checked="form.published"
            :checked-children="t('common.actions.publish')"
            :un-checked-children="t('video.status.draft')"
          />
          <span style="margin-left: 8px; color: var(--text-secondary, #667085); font-size: 13px;">
            {{ form.published ? t('video.form.publishedHint') : t('video.form.draftHint') }}
          </span>
        </a-form-item>
      </a-form>
    </a-modal>

    <MediaPicker
      v-model:visible="mediaPickerVisible"
      module="video"
      @selected="handleMediaSelected"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PlaySquareOutlined,
  LinkOutlined,
  PictureOutlined,
} from '@ant-design/icons-vue'
import { videoApi } from '@/api/video'
import type { VideoWork, ParseVideoReq } from '@/types/video'
import MediaPicker from '@/components/media/MediaPicker.vue'
import CategoryBar from '@/components/common/CategoryBar.vue'
import CategorySelect from '@/components/article/CategorySelect.vue'
import { getThumbUrl } from '@/utils/image'
import PlatformIcon from '@/components/video/PlatformIcon.vue'
import { PLATFORMS, PLATFORM_MAP } from '@/components/video/platforms'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'
import { formatDateTime } from '@/utils/datetime'

const { t } = useI18n()

const keyword = ref('')
const statusFilter = ref<number | undefined>(undefined)
const categoryFilter = ref<string | undefined>(undefined)
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const list = ref<VideoWork[]>([])
const loading = ref(false)

const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(12)

const modalVisible = ref(false)
const modalLoading = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref('')
const mediaPickerVisible = ref(false)

const form = reactive({
  title: '',
  cover_url: '',
  coverMediaId: '',
  coverPresetId: '',
  description: '',
  category_id: undefined as string | undefined,
  published: true,
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
      status: statusFilter.value,
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

function handlePageChange(newPage: number, newPageSize?: number) {
  changePagination(newPage, newPageSize)
  fetchList()
}

function handleCategorySelect(id: string | undefined) {
  categoryFilter.value = id
  reset()
  fetchList()
}

function resetForm() {
  form.title = ''
  form.cover_url = ''
  form.coverMediaId = ''
  form.coverPresetId = ''
  form.description = ''
  form.category_id = undefined
  form.published = true
  Object.keys(selectedPlatforms).forEach((key) => delete selectedPlatforms[key])
}

function openCreateModal() {
  modalMode.value = 'create'
  editingId.value = ''
  resetForm()
  modalVisible.value = true
}

// 媒体库选图：记录媒体引用（结构化封面字段走 ID 优先），同时回填 URL 用于展示
function handleMediaSelected(media: { id: string; url: string; presetId?: string }) {
  form.cover_url = media.url
  form.coverMediaId = media.id
  form.coverPresetId = media.presetId || ''
}

// 手输 URL = 外链模式，清空媒体引用
function handleCoverUrlInput() {
  form.coverMediaId = ''
  form.coverPresetId = ''
}

function handleEdit(item: VideoWork) {
  modalMode.value = 'edit'
  editingId.value = item.id
  resetForm()
  form.title = item.title
  form.cover_url = item.cover_url || ''
  form.coverMediaId = item.cover_media_id || ''
  form.coverPresetId = item.cover_preset_id || ''
  form.description = item.description || ''
  form.category_id = item.category_id || undefined
  form.published = item.status === 1
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
    message.warning(t('video.message.linkRequired'))
    return
  }
  parsingPlatform.value = platformKey
  try {
    const res = await videoApi.parse({ platform: platformKey, url: url.trim() })
    const hasExisting = form.title || form.cover_url || form.description
    const doFill = () => {
      if (res.title) form.title = res.title
      if (res.cover_url) {
        // 解析到的封面为平台外链，切换为外链模式并清空媒体引用
        form.cover_url = res.cover_url
        form.coverMediaId = ''
        form.coverPresetId = ''
      }
      if (res.description) form.description = res.description
      message.success(t('video.message.parseSuccess'))
    }
    if (hasExisting) {
      Modal.confirm({
        title: t('video.message.overwriteTitle'),
        content: t('video.message.overwriteContent'),
        okText: t('video.actions.overwrite'),
        cancelText: t('common.actions.cancel'),
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
    message.warning(t('video.message.titleRequired'))
    return
  }
  const platforms = buildPlatforms()
  if (platforms.length === 0) {
    message.warning(t('video.message.platformRequired'))
    return
  }
  modalLoading.value = true
  try {
    const payload = {
      title: form.title.trim(),
      cover_url: form.cover_url.trim() || undefined,
      cover_media_id: form.coverMediaId || undefined,
      cover_preset_id: form.coverPresetId || undefined,
      description: form.description.trim() || undefined,
      category_id: form.category_id,
      status: form.published ? 1 : 0,
      platforms,
    }
    if (modalMode.value === 'create') {
      await videoApi.create(payload)
      message.success(t('common.message.createSuccess'))
    } else {
      await videoApi.update(editingId.value, payload)
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

function handleDelete(item: VideoWork) {
  Modal.confirm({
    title: t('video.message.deleteTitle'),
    content: t('video.message.deleteConfirm', { name: item.title }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await videoApi.remove(item.id)
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
  border: 1px solid var(--border-color, #E5E9F2);
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
  background: var(--bg-secondary, #F5F7FC);
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
  color: var(--text-secondary, #8A93A8);
}

.status-tag {
  position: absolute;
  top: 8px;
  right: 8px;
  pointer-events: none;
}

.category-cover-tag {
  right: auto;
  left: 8px;
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
  color: var(--text-primary, #29365C);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.video-desc {
  font-size: 13px;
  color: var(--text-secondary, #667085);
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
  color: var(--text-secondary, #8A93A8);
}

.video-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-secondary, #8A93A8);
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
  border-top: 1px solid var(--border-color, #E5E9F2);
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
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary, #667085);
  transition: all 0.2s;
}

.platform-toggle:hover {
  border-color: var(--primary-color, #526FE8);
}

.platform-toggle.active {
  background: rgba(82, 111, 232, 0.06);
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
  color: var(--text-primary, #29365C);
}

.platform-url-input-group {
  display: flex;
}
</style>
