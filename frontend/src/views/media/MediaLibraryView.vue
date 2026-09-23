<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">媒体库</h1>
      <a-upload
        :show-upload-list="false"
        :before-upload="handleBeforeUpload"
        :custom-request="handleCustomUpload"
        accept="image/*,video/*"
      >
        <a-button type="primary" :loading="uploading">
          <UploadOutlined /> 上传
        </a-button>
      </a-upload>
    </div>

    <a-card>
      <a-tabs v-model:activeKey="activeTab" @change="handleTabChange">
        <a-tab-pane key="1">
          <template #tab>
            <span><PictureOutlined /> 图片</span>
          </template>
        </a-tab-pane>
        <a-tab-pane key="2">
          <template #tab>
            <span><VideoCameraOutlined /> 视频</span>
          </template>
        </a-tab-pane>
      </a-tabs>

      <div class="media-toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索文件名..."
          style="width: 240px"
          allow-clear
        />
        <a-radio-group v-model:value="viewType">
          <a-radio-button value="grid">
            <AppstoreOutlined /> 图标视图
          </a-radio-button>
          <a-radio-button value="list">
            <UnorderedListOutlined /> 文件视图
          </a-radio-button>
        </a-radio-group>
      </div>

      <a-spin :spinning="loading">
        <a-empty
          v-if="!loading && list.length === 0"
          description="暂无媒体文件"
          style="margin-top: 48px"
        />
        <template v-else>
          <div v-if="viewType === 'grid'" class="media-grid">
            <div
              v-for="item in list"
              :key="item.id"
              class="media-card"
            >
              <div class="media-content">
                <img
                  v-if="item.file_type === 1"
                  :src="item.thumb_url || getThumbUrl(item.url, 300)"
                  :alt="item.filename"
                />
                <div v-else class="video-card">
                  <PlayCircleOutlined class="video-icon" />
                </div>
                <div class="media-overlay">
                  <a-button
                    v-if="item.file_type === 1"
                    type="primary"
                    size="small"
                    ghost
                    @click.stop="openWorkbench(item)"
                  >
                    <SettingOutlined /> 预设
                  </a-button>
                  <a-button size="small" danger ghost @click.stop="openDelete(item)">
                    <DeleteOutlined /> 删除
                  </a-button>
                </div>
              </div>
              <div class="media-name" :title="item.filename">{{ item.filename }}</div>
            </div>
          </div>
          <a-table
            v-else
            :columns="columns"
            :data-source="list"
            :pagination="false"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'actions'">
                <a-button type="link" danger size="small" @click="openDelete(record as MediaItem)">
                  删除
                </a-button>
              </template>
            </template>
          </a-table>
        </template>
      </a-spin>

      <div v-if="total > 0" class="pagination-wrapper">
        <a-pagination
          :current="page"
          :page-size="pageSize"
          :total="total"
          :page-size-options="[20, 40, 60]"
          show-size-changer
          :show-total="(t: number) => `共 ${t} 条`"
          @change="handlePageChange"
          @show-size-change="handlePageChange"
        />
      </div>
    </a-card>

    <MediaPresetWorkbench
      v-if="selectedMedia"
      v-model:visible="workbenchVisible"
      :media-id="selectedMedia.id"
      :original-url="selectedMedia.url"
      :filename="selectedMedia.filename"
      :size="selectedMedia.size"
      @saved="handleWorkbenchSaved"
    />

    <MediaDeleteModal
      v-model:visible="deleteVisible"
      :media="deleteTarget"
      @success="handleDeleteSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { message, type TableColumnsType } from 'ant-design-vue'
import {
  PictureOutlined,
  VideoCameraOutlined,
  UploadOutlined,
  AppstoreOutlined,
  UnorderedListOutlined,
  PlayCircleOutlined,
  SettingOutlined,
  DeleteOutlined,
} from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaItem } from '@/api/media'
import { getThumbUrl } from '@/utils/image'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'
import MediaPresetWorkbench from '@/components/media/MediaPresetWorkbench.vue'
import MediaDeleteModal from '@/components/media/MediaDeleteModal.vue'

const activeTab = ref<'1' | '2'>('1')
const fileType = computed(() => Number(activeTab.value))

const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const viewType = ref<'grid' | 'list'>('grid')
const list = ref<MediaItem[]>([])
const loading = ref(false)
const uploading = ref(false)

const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(20)

const workbenchVisible = ref(false)
const selectedMedia = ref<MediaItem | null>(null)

const deleteVisible = ref(false)
const deleteTarget = ref<MediaItem | null>(null)

const columns: TableColumnsType<MediaItem> = [
  {
    title: '文件名',
    dataIndex: 'filename',
    key: 'filename',
    ellipsis: true,
  },
  {
    title: '类型',
    key: 'type',
    width: 100,
    customRender: ({ record }) => (record.file_type === 1 ? '图片' : '视频'),
  },
  {
    title: '尺寸',
    key: 'dimensions',
    width: 120,
    customRender: ({ record }) =>
      record.width && record.height ? `${record.width}x${record.height}` : '-',
  },
  {
    title: '大小',
    key: 'size',
    width: 120,
    customRender: ({ record }) => formatBytes(record.size),
  },
  {
    title: '上传时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 180,
    customRender: ({ text }) => formatDateTime(text),
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
  },
]

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
    const res = await mediaApi.getList({
      file_type: fileType.value,
      keyword: debouncedKeyword.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    })
    list.value = res.list || []
    total.value = res.total || 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

function handleTabChange(key: string | number) {
  activeTab.value = String(key) as '1' | '2'
  reset()
  fetchList()
}

function handlePageChange(newPage: number, newPageSize?: number) {
  changePagination(newPage, newPageSize)
  fetchList()
}

function handleBeforeUpload(file: File) {
  const maxSize = 100 * 1024 * 1024
  if (file.size > maxSize) {
    message.error('文件大小不能超过 100MB')
    return false
  }
  return true
}

function handleCustomUpload({ file, onSuccess, onError }: any) {
  uploading.value = true
  mediaApi
    .upload(file as File, (percent) => {
      message.loading({ content: `上传中 ${percent}%`, key: 'upload', duration: 0 })
    })
    .then(() => {
      message.success({ content: '上传成功', key: 'upload' })
      onSuccess?.()
      fetchList()
    })
    .catch((err) => {
      message.error({ content: err?.message || '上传失败', key: 'upload' })
      onError?.(err)
    })
    .finally(() => {
      uploading.value = false
    })
}

function openWorkbench(item: MediaItem) {
  selectedMedia.value = item
  workbenchVisible.value = true
}

function handleWorkbenchSaved() {
  fetchList()
}

function openDelete(item: MediaItem) {
  deleteTarget.value = item
  deleteVisible.value = true
}

function handleDeleteSuccess() {
  fetchList()
}

function formatBytes(bytes?: number): string {
  if (bytes == null || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

function formatDateTime(time?: string): string {
  if (!time) return '-'
  const d = new Date(time)
  if (Number.isNaN(d.getTime())) return time
  return d.toLocaleString('zh-CN')
}
</script>

<style scoped>
.media-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin: 16px 0 20px;
  flex-wrap: wrap;
}

.media-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
}

.media-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #f0f0f0);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  transition: box-shadow 0.2s;
  display: flex;
  flex-direction: column;
}

.media-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.media-card:hover .media-overlay {
  opacity: 1;
}

.media-content {
  position: relative;
  width: 100%;
  aspect-ratio: 1;
  overflow: hidden;
}

.media-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: rgba(0, 0, 0, 0.4);
  opacity: 0;
  transition: opacity 0.2s;
}

.media-content img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-card {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary, #f5f5f5);
  color: var(--text-secondary, #666);
  padding: 16px;
}

.video-icon {
  font-size: 48px;
}

.media-name {
  font-size: 12px;
  line-height: 1.4;
  padding: 6px 8px;
  text-align: center;
  color: var(--text-secondary, #595959);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}
</style>
