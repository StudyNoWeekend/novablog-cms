<template>
  <a-modal
    v-model:open="open"
    title="选择图片版本"
    width="680px"
    :footer="null"
    @cancel="handleCancel"
  >
    <a-spin :spinning="loading">
      <div class="chooser-tip">
        点击图片直接打开该图的预设工作台（制作/编辑相框 EXIF 成品图），点「使用」选定版本后确认。
        <a :href="mediaLibraryUrl" target="_blank" rel="noopener">在媒体库管理预设</a>
      </div>
      <div class="chooser-grid">
        <div
          v-for="card in cards"
          :key="card.key"
          class="chooser-item"
          :class="{ selected: choiceKind === card.key }"
        >
          <div
            class="chooser-item-media"
            role="button"
            tabindex="0"
            :aria-label="`打开「${card.name}」的预设工作台`"
            @click="openWorkbench"
            @keydown.enter="openWorkbench"
          >
            <img :src="card.url" :alt="card.name" referrerpolicy="no-referrer" loading="lazy" />
            <div class="media-hover-tip">
              <EditOutlined />
              <span>打开预设工作台</span>
            </div>
          </div>
          <div class="chooser-item-bar">
            <span class="chooser-item-name" :title="card.name">{{ card.name }}</span>
            <a-button
              size="small"
              :type="choiceKind === card.key ? 'primary' : 'default'"
              :aria-label="`使用${card.name}`"
              @click="choiceKind = card.key"
            >
              {{ choiceKind === card.key ? '使用中' : '使用' }}
            </a-button>
          </div>
        </div>

        <!-- 制作预设入口：始终展示，无预设时是唯一出路 -->
        <div
          class="chooser-item chooser-item-create"
          role="button"
          tabindex="0"
          aria-label="制作新预设"
          @click="openWorkbench"
          @keydown.enter="openWorkbench"
        >
          <PlusOutlined class="create-icon" />
          <span>制作预设</span>
        </div>
      </div>
      <div class="chooser-footer">
        <a-button aria-label="取消" @click="handleCancel">取消</a-button>
        <a-button type="primary" :disabled="!choiceKind" aria-label="确认" @click="handleConfirm">
          确认
        </a-button>
      </div>
    </a-spin>

    <MediaPresetWorkbench
      v-if="mediaId"
      v-model:visible="workbenchVisible"
      :media-id="mediaId"
      :original-url="originalUrl"
      @saved="load"
    />
  </a-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { EditOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaPreset } from '@/types/api'
import MediaPresetWorkbench from './MediaPresetWorkbench.vue'

const CHOICE_ORIGINAL = '__original__'

interface VersionCard {
  key: string
  name: string
  url: string
}

const props = defineProps<{
  /** 目标媒体 ID；为空时视为无预设场景，直接回传原图 */
  mediaId: string | null
  /** 原图 URL（"使用原图"选项） */
  originalUrl: string
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'confirm', payload: { url: string; presetId?: string; presetName?: string }): void
}>()

const router = useRouter()
const mediaLibraryUrl = router.resolve('/media').href

const open = ref(props.visible)
const presets = ref<MediaPreset[]>([])
const loading = ref(false)
const choiceKind = ref<string>(CHOICE_ORIGINAL)
const workbenchVisible = ref(false)

const cards = computed<VersionCard[]>(() => [
  { key: CHOICE_ORIGINAL, name: '使用原图', url: props.originalUrl },
  ...presets.value.map((p) => ({ key: p.id, name: p.name, url: p.output_url })),
])

watch(
  () => props.visible,
  (v) => {
    open.value = v
    if (v) load()
  }
)

watch(open, (v) => {
  if (!v) emit('update:visible', false)
})

async function load() {
  if (!props.mediaId) {
    emit('confirm', { url: props.originalUrl })
    return
  }
  loading.value = true
  try {
    const res = await mediaApi.getPresets(props.mediaId)
    presets.value = res.list || []
  } catch {
    presets.value = []
  } finally {
    loading.value = false
  }
  // 工作台里可能删掉了正选中的预设，失效时回退到原图
  if (
    choiceKind.value !== CHOICE_ORIGINAL &&
    !presets.value.some((p) => p.id === choiceKind.value)
  ) {
    choiceKind.value = CHOICE_ORIGINAL
  }
}

function openWorkbench() {
  workbenchVisible.value = true
}

function handleConfirm() {
  const card = cards.value.find((c) => c.key === choiceKind.value)
  if (!card) return
  if (card.key === CHOICE_ORIGINAL) {
    emit('confirm', { url: card.url })
  } else {
    emit('confirm', { url: card.url, presetId: card.key, presetName: card.name })
  }
  open.value = false
}

function handleCancel() {
  open.value = false
}
</script>

<style scoped>
.chooser-tip {
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--text-secondary, #667085);
}

.chooser-tip a {
  margin-left: 4px;
}

.chooser-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  max-height: 420px;
  overflow-y: auto;
  padding: 4px;
}

.chooser-item {
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid var(--border-color, #E5E9F2);
  background: var(--bg-secondary, #F5F7FC);
  transition: border-color 0.2s;
}

.chooser-item.selected {
  border-color: var(--primary, #526FE8);
}

.chooser-item-media {
  position: relative;
  cursor: pointer;
}

.chooser-item-media img {
  width: 100%;
  aspect-ratio: 1;
  object-fit: cover;
  display: block;
}

.media-hover-tip {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: rgba(15, 23, 42, 0.55);
  color: #fff;
  font-size: 13px;
  opacity: 0;
  transition: opacity 0.2s;
}

.media-hover-tip :deep(.anticon) {
  font-size: 20px;
}

.chooser-item-media:hover .media-hover-tip,
.chooser-item-media:focus-visible .media-hover-tip {
  opacity: 1;
}

.chooser-item-media:focus-visible {
  outline: 2px solid var(--primary, #526FE8);
  outline-offset: 2px;
}

.chooser-item-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
}

.chooser-item-name {
  flex: 1;
  min-width: 0;
  font-size: 12px;
  color: var(--text-secondary, #667085);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chooser-item-create {
  aspect-ratio: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border-style: dashed;
  border-color: var(--primary, #526FE8);
  color: var(--primary, #526FE8);
  font-size: 13px;
  cursor: pointer;
}

.chooser-item-create:hover {
  background: rgba(82, 111, 232, 0.06);
}

.create-icon {
  font-size: 24px;
}

.chooser-footer {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
