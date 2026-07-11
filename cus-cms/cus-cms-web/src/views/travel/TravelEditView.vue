<template>
  <div v-if="notFound" class="not-found-wrapper">
    <a-result status="404" title="攻略不存在" sub-title="您访问的攻略可能已被删除或从未创建">
      <template #extra>
        <a-button type="primary" aria-label="返回攻略列表" @click="router.push('/travels')">
          <ArrowLeftOutlined />
          返回攻略列表
        </a-button>
      </template>
    </a-result>
  </div>

  <div v-else class="travel-edit-view">
    <header class="action-bar">
      <div class="action-bar-left">
        <a-button
          type="text"
          aria-label="返回"
          @click="handleBack"
        >
          <ArrowLeftOutlined />
        </a-button>
        <h1 ref="pageTitleRef" tabindex="-1" class="page-title">
          {{ pageTitle }}
        </h1>
      </div>

      <div class="action-bar-right">
        <a-button
          aria-label="保存草稿"
          :loading="saving"
          @click="handleSaveDraft"
        >
          <SaveOutlined />
          保存草稿
        </a-button>

        <a-button
          v-show="!isDesktop"
          aria-label="预览"
          @click="previewVisible = true"
        >
          <EyeOutlined />
          预览
        </a-button>

        <a-button
          type="primary"
          aria-label="发布攻略"
          :loading="saving"
          @click="handlePublish"
        >
          <SendOutlined />
          发布
        </a-button>
      </div>
    </header>

    <div class="edit-layout">
      <main class="edit-main">
        <TravelBasicForm
          ref="basicFormRef"
          :modelValue="guide"
          @update:modelValue="handleUpdate"
          @daysChange="handleDaysChange"
        />

        <TravelAttractionPool
          :modelValue="guide"
          @update:modelValue="handleUpdate"
          @dragStart="handleDragStart"
        />

        <TravelItineraryBuilder
          :modelValue="guide"
          @update:modelValue="handleUpdate"
        />
      </main>

      <div class="edit-preview-spacer" aria-hidden="true" />
    </div>

    <TravelEditorPreview
      v-model:visible="previewVisible"
      :guide="guide"
      class="travel-edit-preview"
    />

    <a-button
      v-show="!isDesktop"
      type="primary"
      shape="circle"
      size="large"
      class="mobile-preview-float"
      aria-label="打开预览"
      @click="previewVisible = true"
    >
      <EyeOutlined />
    </a-button>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  nextTick,
  onMounted,
  onUnmounted,
  reactive,
  ref,
} from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  ArrowLeftOutlined,
  EyeOutlined,
  PlusOutlined,
  SaveOutlined,
  SendOutlined,
} from '@ant-design/icons-vue'
import TravelBasicForm from '@/components/travel/editor/TravelBasicForm.vue'
import TravelAttractionPool from '@/components/travel/editor/TravelAttractionPool.vue'
import TravelItineraryBuilder from '@/components/travel/editor/TravelItineraryBuilder.vue'
import TravelEditorPreview from '@/components/travel/editor/TravelEditorPreview.vue'
import { createEmptyGuide } from '@/mocks/travel'
import { travelApi } from '@/api/travel'
import { TravelStatus, toFormData } from '@/types/travel'
import type { TravelAttraction, TravelGuideFormData } from '@/types/travel'

const route = useRoute()
const router = useRouter()

const basicFormRef = ref<InstanceType<typeof TravelBasicForm> | null>(null)
const pageTitleRef = ref<HTMLElement | null>(null)

const isCreateMode = route.path === '/travels/create'
const pageTitle = computed(() =>
  isCreateMode ? '新建攻略' : '编辑攻略',
)

const guide = reactive<TravelGuideFormData>(createEmptyGuide())
const hasUnsavedChanges = ref(false)
const previewVisible = ref(false)
const draggingAttraction = ref<TravelAttraction | null>(null)
const notFound = ref(false)
const saving = ref(false)
const loadingDetail = ref(false)
const windowWidth = ref(window.innerWidth)

const isDesktop = computed(() => windowWidth.value >= 1024)

function onResize() {
  windowWidth.value = window.innerWidth
}

onMounted(async () => {
  window.addEventListener('resize', onResize)

  if (!isCreateMode) {
    const id = route.params.id as string
    loadingDetail.value = true
    try {
      const detail = await travelApi.getDetail(id)
      Object.assign(guide, toFormData(detail))
    } catch {
      notFound.value = true
    } finally {
      loadingDetail.value = false
    }
  }

  nextTick(() => {
    pageTitleRef.value?.focus()
  })
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
})

function handleUpdate(value: TravelGuideFormData) {
  Object.assign(guide, value)
  hasUnsavedChanges.value = true
}

function handleDaysChange(days: number) {
  hasUnsavedChanges.value = true

  const current = guide.itinerary
  if (days > current.length) {
    const added = []
    for (let d = current.length + 1; d <= days; d++) {
      added.push({
        day: d,
        title: '',
        description: '',
        attractions: [],
        attractionIds: [],
      })
    }
    guide.itinerary = [...current, ...added]
  } else if (days < current.length) {
    guide.itinerary = current.slice(0, days)
  }
}

function handleDragStart(attraction: TravelAttraction) {
  draggingAttraction.value = attraction
}

async function persistGuide(status: TravelStatus) {
  guide.status = status
  if (guide.id) {
    const saved = await travelApi.update(guide.id, guide)
    guide.id = saved.id
  } else {
    const saved = await travelApi.create(guide)
    guide.id = saved.id
  }
  hasUnsavedChanges.value = false
}

async function handleSaveDraft() {
  saving.value = true
  try {
    await persistGuide(TravelStatus.Draft)
    message.success('保存草稿成功')
    router.push('/travels')
  } catch {
    // 错误已由拦截器处理
  } finally {
    saving.value = false
  }
}

async function handlePublish() {
  const valid = basicFormRef.value?.isValid()
  if (!valid) {
    message.warning('请完善必填信息')
    return
  }

  saving.value = true
  try {
    await persistGuide(TravelStatus.Published)
    message.success('发布成功')
    router.push('/travels')
  } catch {
    // 错误已由拦截器处理
  } finally {
    saving.value = false
  }
}

function handleBack() {
  if (hasUnsavedChanges.value) {
    Modal.confirm({
      title: '确认离开',
      content: '有未保存的修改，是否确认离开？',
      okText: '离开',
      cancelText: '取消',
      onOk: () => {
        router.back()
      },
    })
  } else {
    router.back()
  }
}
</script>

<style scoped>
.travel-edit-view {
  --action-bar-height: 64px;

  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  flex-direction: column;
  background: var(--bg-page, #f8f9fb);
  overflow-x: hidden;
}

.not-found-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - var(--header-height, 56px) - 48px);
}

.action-bar {
  flex-shrink: 0;
  height: var(--action-bar-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 24px;
  background: var(--bg-card, #ffffff);
  border-bottom: 1px solid var(--border-color, #e2e8f0);
  box-shadow: var(--shadow-sm, 0 1px 2px rgba(0, 0, 0, 0.04));
}

.action-bar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.page-title:focus {
  outline: none;
}

.action-bar-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.edit-layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.edit-main {
  flex: 0 0 55%;
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 24px;
  overflow-y: auto;
  min-width: 0;
}

.edit-preview-spacer {
  flex: 0 0 45%;
}

.travel-edit-preview :deep(.travel-editor-preview--desktop) {
  position: absolute;
  top: var(--action-bar-height);
  right: 0;
  bottom: 0;
  width: 45%;
  max-width: none;
  height: auto;
  border-radius: 0;
}

.mobile-preview-float {
  position: fixed;
  right: 20px;
  bottom: 24px;
  z-index: 110;
  box-shadow: var(--shadow-dropdown, 0 4px 16px rgba(0, 0, 0, 0.08));
}

@media (max-width: 1023px) {
  .travel-edit-view {
    --action-bar-height: auto;
  }

  .action-bar {
    height: auto;
    min-height: 56px;
    flex-wrap: wrap;
    padding: 12px 16px;
  }

  .action-bar-right {
    width: 100%;
    justify-content: flex-end;
  }

  .edit-layout {
    flex-direction: column;
  }

  .edit-main {
    flex: 1 1 auto;
    width: 100%;
    padding: 16px;
  }

  .edit-preview-spacer {
    display: none;
  }

  .travel-edit-preview :deep(.travel-editor-preview--desktop) {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .mobile-preview-float {
    transition: none;
  }
}
</style>
