<template>
  <a-modal
    v-model:open="open"
    title="选择图片"
    width="680px"
    :footer="null"
    @cancel="handleCancel"
  >
    <div class="media-picker-grid">
      <div
        v-for="item in mediaList"
        :key="item.id"
        class="media-picker-item"
        :class="{ selected: selectedId === item.id }"
        @click="selectedId = item.id"
      >
        <img :src="item.thumb_url || getThumbUrl(item.url, 300)" :alt="item.filename" />
        <div v-if="selectedId === item.id" class="selected-overlay">
          <CheckCircleFilled />
        </div>
      </div>
    </div>
    <a-empty v-if="!loading && mediaList.length === 0" description="暂无图片" />
    <a-pagination
      v-if="total > 20"
      :current="page"
      :page-size="20"
      :total="total"
      @change="handlePageChange"
      style="margin-top: 16px; text-align: center;"
    />

    <div v-if="selectedId" class="media-picker-footer">
      <a-button type="primary" @click="handleConfirm">确认选择</a-button>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { CheckCircleFilled } from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaItem } from '@/api/media'
import { getThumbUrl } from '@/utils/image'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'selected', media: { id: string; url: string }): void
}>()

const open = ref(props.visible)
const mediaList = ref<MediaItem[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)
const selectedId = ref<string | null>(null)

watch(() => props.visible, (v) => {
  open.value = v
  if (v) {
    page.value = 1
    selectedId.value = null
    fetchMedia()
  }
})

watch(open, (v) => {
  if (!v) {
    emit('update:visible', false)
  }
})

async function fetchMedia() {
  loading.value = true
  try {
    const res = await mediaApi.getList({ page: page.value, page_size: 20 })
    mediaList.value = res.list || []
    total.value = res.total || 0
  } catch {}
  finally { loading.value = false }
}

function handlePageChange(p: number) {
  page.value = p
  fetchMedia()
}

function handleConfirm() {
  const item = mediaList.value.find(m => m.id === selectedId.value)
  if (item) {
    emit('selected', { id: item.id, url: item.url })
  }
  open.value = false
}

function handleCancel() {
  open.value = false
}
</script>

<style scoped>
.media-picker-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}
.media-picker-item {
  aspect-ratio: 1;
  position: relative;
  cursor: pointer;
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid transparent;
  transition: border-color 0.2s;
}
.media-picker-item:hover { border-color: var(--primary, #4a6cf7); }
.media-picker-item.selected { border-color: var(--primary, #4a6cf7); }
.media-picker-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.selected-overlay {
  position: absolute;
  inset: 0;
  background: rgba(74, 108, 247, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  color: var(--primary, #4a6cf7);
}
.media-picker-footer {
  margin-top: 16px;
  text-align: right;
}
</style>