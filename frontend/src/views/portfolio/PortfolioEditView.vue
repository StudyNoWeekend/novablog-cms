<template>
  <div class="page-container">
    <div v-if="loading" class="loading-wrapper">
      <a-skeleton active :paragraph="{ rows: 8 }" />
    </div>
    <template v-else>
      <!-- 顶部：返回 + 作品集信息编辑 -->
      <div class="edit-header">
        <div class="header-top">
          <a-button type="text" @click="router.push('/portfolios')">
            <ArrowLeftOutlined /> {{ t('portfolio.edit.backToList') }}
          </a-button>
          <div class="header-actions">
            <a-button :loading="saving" @click="handleSave">{{ t('portfolio.edit.saveInfo') }}</a-button>
            <PushSyncButton
              v-if="portfolioId"
              target-type="portfolio"
              :target-id="portfolioId"
            />
          </div>
        </div>
        <div class="header-form">
          <a-input
            v-model:value="form.name"
            :placeholder="t('portfolio.edit.namePlaceholder')"
            size="large"
            class="name-input"
          />
          <a-textarea
            v-model:value="form.description"
            :placeholder="t('portfolio.edit.descPlaceholder')"
            :rows="2"
            :maxlength="500"
            class="desc-input"
          />
          <!-- 分类设置 -->
          <div class="category-setting">
            <div class="category-label">{{ t('common.field.category') }}</div>
            <CategorySelect
              v-model="form.category_id"
              type="portfolio"
              :placeholder="t('portfolio.form.categoryPlaceholder')"
              class="category-select"
            />
          </div>
          <!-- 封面设置 -->
          <div class="cover-setting">
            <div class="cover-label">{{ t('portfolio.cover.setting') }}</div>
            <a-radio-group v-model:value="coverMode">
              <a-radio :value="0">{{ t('portfolio.cover.useFirst') }}</a-radio>
              <a-radio :value="1">{{ t('portfolio.cover.custom') }}</a-radio>
            </a-radio-group>
            <div v-if="coverMode === 1" class="cover-custom">
              <div v-if="coverPreviewUrl" class="cover-preview">
                <img :src="coverPreviewUrl" :alt="t('portfolio.cover.previewAlt')" />
                <div class="cover-preview-actions">
                  <a-button type="link" size="small" @click="openCoverPicker">{{ t('portfolio.cover.change') }}</a-button>
                  <a-button type="link" size="small" danger @click="clearCoverPreset">{{ t('portfolio.cover.clear') }}</a-button>
                </div>
              </div>
              <a-button v-else type="dashed" @click="openCoverPicker">
                <PictureOutlined /> {{ t('portfolio.cover.selectPreset') }}
              </a-button>
            </div>
            <div v-else class="cover-tip">
              {{ t('portfolio.cover.autoTip') }}
            </div>
          </div>
        </div>
      </div>

      <!-- 工具栏 -->
      <div class="items-toolbar">
        <span class="items-title">
          {{ t('portfolio.item.list') }} <span class="items-count">({{ items.length }})</span>
        </span>
        <a-button type="primary" @click="openAddPicker">
          <PlusOutlined /> {{ t('portfolio.item.add') }}
        </a-button>
      </div>

      <!-- 作品项列表 -->
      <a-spin :spinning="actionLoading">
        <a-empty
          v-if="items.length === 0"
          :description="t('portfolio.item.empty')"
          style="margin-top: 48px"
        />
        <div v-else class="items-list">
          <div
            v-for="(item, index) in items"
            :key="item.id"
            class="item-card"
            :class="{ dragging: dragIndex === index, 'drag-over': dragOverIndex === index }"
            :draggable="true"
            @dragstart="handleDragStart($event, index)"
            @dragover.prevent="handleDragOver(index)"
            @dragleave="handleDragLeave"
            @drop.prevent="handleDrop"
            @dragend="handleDragEnd"
          >
            <div class="drag-handle">
              <HolderOutlined />
            </div>
            <div class="item-index">{{ index + 1 }}</div>
            <div class="item-preview">
              <img
                v-if="item.output_url"
                :src="item.output_url"
                :alt="item.title"
              />
              <div v-else class="preset-deleted">
                <PictureOutlined />
                <span>{{ t('portfolio.item.presetDeleted') }}</span>
              </div>
            </div>
            <div class="item-content">
              <div class="item-title" :title="item.title">{{ item.title }}</div>
              <div class="item-desc">
                {{ item.description || t('portfolio.item.noDesc') }}
              </div>
            </div>
            <div class="item-actions">
              <a-button type="link" size="small" @click="openEditItem(item)">
                <EditOutlined /> {{ t('common.actions.edit') }}
              </a-button>
              <a-button type="link" size="small" danger @click="handleDeleteItem(item)">
                <DeleteOutlined /> {{ t('common.actions.delete') }}
              </a-button>
            </div>
          </div>
        </div>
      </a-spin>

      <!-- 添加作品选择器 -->
      <PortfolioItemPicker
        v-model:visible="pickerVisible"
        :title="t('portfolio.item.add')"
        @selected="handleAddItem"
      />

      <!-- 编辑作品项弹窗 -->
      <a-modal
        v-model:open="editItemVisible"
        :title="t('portfolio.item.editTitle')"
        :confirm-loading="savingItem"
        @ok="handleSaveItem"
      >
        <a-form layout="vertical">
          <a-form-item :label="t('portfolio.item.titleLabel')" required>
            <a-input
              v-model:value="editItemForm.title"
              :placeholder="t('portfolio.item.titlePlaceholder')"
              :maxlength="100"
              show-count
            />
          </a-form-item>
          <a-form-item :label="t('portfolio.item.introLabel')">
            <a-textarea
              v-model:value="editItemForm.description"
              :placeholder="t('portfolio.item.introPlaceholder')"
              :rows="4"
              :maxlength="500"
              show-count
            />
          </a-form-item>
          <a-form-item :label="t('portfolio.item.presetLabel')">
            <div v-if="presetPreview" class="preset-preview">
              <img
                v-if="presetPreview.output_url"
                :src="presetPreview.output_url"
                :alt="presetPreview.name"
              />
              <div v-else class="preset-preview-placeholder">
                <PictureOutlined />
              </div>
              <span>{{ presetPreview.name }}</span>
              <a-button type="link" size="small" @click="openReselectPreset">{{ t('portfolio.item.reselect') }}</a-button>
            </div>
            <a-button v-else type="dashed" block @click="openReselectPreset">
              <SwapOutlined /> {{ t('portfolio.item.selectPreset') }}
            </a-button>
          </a-form-item>
        </a-form>
      </a-modal>

      <!-- 重选预设选择器 -->
      <PortfolioItemPicker
        v-model:visible="reselectPickerVisible"
        :title="t('portfolio.item.reselectTitle')"
        @selected="handleReselectPreset"
      />

      <!-- 封面预设选择器 -->
      <PortfolioItemPicker
        v-model:visible="coverPickerVisible"
        :title="t('portfolio.cover.selectPreset')"
        preset-only
        @selected="handleSelectCover"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import PushSyncButton from '@/components/community/PushSyncButton.vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { message, Modal } from 'ant-design-vue'
import {
  ArrowLeftOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PictureOutlined,
  HolderOutlined,
  SwapOutlined,
} from '@ant-design/icons-vue'
import { portfolioApi } from '@/api/portfolio'
import type { PortfolioDetail, PortfolioItem } from '@/types/portfolio'
import PortfolioItemPicker from '@/components/portfolio/PortfolioItemPicker.vue'
import CategorySelect from '@/components/article/CategorySelect.vue'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const portfolioId = computed(() => String(route.params.id))

const loading = ref(true)
const actionLoading = ref(false)
const saving = ref(false)
const detail = ref<PortfolioDetail | null>(null)
const items = ref<PortfolioItem[]>([])

const form = reactive({
  name: '',
  description: '',
  category_id: undefined as string | undefined,
})

// 封面设置
const coverMode = ref(0)
const coverPresetId = ref<string | null>(null)
const coverPreviewUrl = ref('')
const coverPickerVisible = ref(false)

// 拖拽排序
const dragIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

// 添加作品
const pickerVisible = ref(false)

// 编辑作品项
const editItemVisible = ref(false)
const savingItem = ref(false)
const editingItem = ref<PortfolioItem | null>(null)
const editItemForm = reactive({
  title: '',
  description: '',
  preset_id: '',
})
const reselectPickerVisible = ref(false)
const reselectPresetCache = ref<{ id: string; name: string; output_url: string } | null>(null)

const presetPreview = computed(() => {
  if (reselectPresetCache.value) {
    return reselectPresetCache.value
  }
  const editing = editingItem.value
  if (editing && editing.output_url) {
    return { name: t('portfolio.item.currentPreset'), output_url: editing.output_url }
  }
  if (editing && editing.preset_id) {
    return { name: t('portfolio.item.presetDeleted'), output_url: '' }
  }
  return null
})

onMounted(() => {
  fetchDetail()
})

async function fetchDetail() {
  loading.value = true
  try {
    const res = await portfolioApi.getById(portfolioId.value)
    detail.value = res
    form.name = res.name
    form.description = res.description || ''
    form.category_id = res.category_id || undefined
    coverMode.value = res.cover_mode ?? 0
    coverPresetId.value = res.cover_preset_id || null
    coverPreviewUrl.value = res.cover_url || ''
    items.value = [...(res.items || [])].sort((a, b) => a.sort_order - b.sort_order)
  } catch {
    // 错误由拦截器处理
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  if (!form.name.trim()) {
    message.warning(t('portfolio.form.nameRequired'))
    return
  }
  if (coverMode.value === 1 && !coverPresetId.value) {
    message.warning(t('portfolio.cover.presetRequired'))
    return
  }
  saving.value = true
  try {
    await portfolioApi.update(portfolioId.value, {
      name: form.name.trim(),
      description: form.description.trim(),
      category_id: form.category_id,
      cover_mode: coverMode.value,
      cover_preset_id:
        coverMode.value === 1 ? coverPresetId.value || undefined : undefined,
    })
    message.success(t('common.message.saveSuccess'))
    await fetchDetail()
  } catch {
    // 错误由拦截器处理
  } finally {
    saving.value = false
  }
}

function openCoverPicker() {
  coverPickerVisible.value = true
}

function handleSelectCover(payload: {
  preset_id: string
  output_url?: string
  preset_name?: string
}) {
  coverPresetId.value = payload.preset_id
  coverPreviewUrl.value = payload.output_url || ''
  coverPickerVisible.value = false
  message.success(t('portfolio.edit.coverSelected'))
}

function clearCoverPreset() {
  coverPresetId.value = null
  coverPreviewUrl.value = ''
}

function openAddPicker() {
  pickerVisible.value = true
}

async function handleAddItem(payload: { preset_id: string; title: string; description: string }) {
  actionLoading.value = true
  try {
    await portfolioApi.addItem(portfolioId.value, {
      preset_id: payload.preset_id,
      title: payload.title,
      description: payload.description || undefined,
    })
    message.success(t('portfolio.item.added'))
    await fetchDetail()
  } catch {
    // 错误由拦截器处理
  } finally {
    actionLoading.value = false
  }
}

function openEditItem(item: PortfolioItem) {
  editingItem.value = item
  editItemForm.title = item.title
  editItemForm.description = item.description || ''
  editItemForm.preset_id = item.preset_id
  reselectPresetCache.value = null
  editItemVisible.value = true
}

function openReselectPreset() {
  reselectPickerVisible.value = true
}

function handleReselectPreset(payload: { preset_id: string; title: string; description: string }) {
  // picker 仅返回 preset_id，预设名称与缩略图在保存后由详情接口刷新
  reselectPresetCache.value = {
    id: payload.preset_id,
    name: t('portfolio.item.newPresetPending'),
    output_url: '',
  }
  editItemForm.preset_id = payload.preset_id
  message.success(t('portfolio.edit.presetSelected'))
}

async function handleSaveItem() {
  if (!editingItem.value) return
  if (!editItemForm.title.trim()) {
    message.warning(t('portfolio.item.titleRequired'))
    return
  }
  savingItem.value = true
  try {
    await portfolioApi.updateItem(portfolioId.value, editingItem.value.id, {
      title: editItemForm.title.trim(),
      description: editItemForm.description.trim() || undefined,
      preset_id: editItemForm.preset_id,
    })
    message.success(t('common.message.saveSuccess'))
    editItemVisible.value = false
    await fetchDetail()
  } catch {
    // 错误由拦截器处理
  } finally {
    savingItem.value = false
  }
}

function handleDeleteItem(item: PortfolioItem) {
  Modal.confirm({
    title: t('portfolio.item.deleteTitle'),
    content: t('portfolio.item.deleteConfirm', { name: item.title }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await portfolioApi.removeItem(portfolioId.value, item.id)
        message.success(t('common.message.deleteSuccess'))
        await fetchDetail()
      } catch {
        // 错误由拦截器处理
      }
    },
  })
}

// 拖拽排序
function handleDragStart(_e: DragEvent, index: number) {
  dragIndex.value = index
}

function handleDragOver(index: number) {
  if (dragIndex.value === null || dragIndex.value === index) return
  dragOverIndex.value = index
}

function handleDragLeave() {
  dragOverIndex.value = null
}

function handleDrop() {
  if (dragIndex.value === null || dragOverIndex.value === null) return
  const from = dragIndex.value
  const to = dragOverIndex.value
  if (from === to) return
  const moved = items.value.splice(from, 1)[0]
  items.value.splice(to, 0, moved)
  persistSort()
  dragIndex.value = null
  dragOverIndex.value = null
}

function handleDragEnd() {
  dragIndex.value = null
  dragOverIndex.value = null
}

async function persistSort() {
  const payload = {
    items: items.value.map((it, idx) => ({ id: it.id, sort_order: idx + 1 })),
  }
  try {
    await portfolioApi.sortItems(portfolioId.value, payload)
    message.success(t('portfolio.edit.sortUpdated'))
  } catch {
    // 排序失败时回滚到服务端数据
    await fetchDetail()
  }
}
</script>

<style scoped>
.loading-wrapper {
  padding: 24px;
}

.edit-header {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: 8px;
  padding: 16px 20px;
  margin-bottom: 20px;
}

.header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.header-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.name-input {
  font-size: 18px;
  font-weight: 600;
}

.category-setting {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.category-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #29365C);
}

.category-select {
  max-width: 320px;
}

.cover-setting {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cover-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #29365C);
}

.cover-custom {
  display: flex;
  align-items: center;
}

.cover-preview {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cover-preview img {
  width: 96px;
  height: 64px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid var(--border-color, #E5E9F2);
}

.cover-preview-actions {
  display: flex;
  flex-direction: column;
}

.cover-tip {
  font-size: 13px;
  color: var(--text-secondary, #667085);
}

.items-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.items-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #29365C);
}

.items-count {
  font-size: 13px;
  font-weight: normal;
  color: var(--text-secondary, #667085);
}

.items-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.item-card {
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: 8px;
  padding: 12px 16px;
  transition: box-shadow 0.2s, border-color 0.2s, opacity 0.2s;
}

.item-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.item-card.dragging {
  opacity: 0.4;
}

.item-card.drag-over {
  border-color: var(--primary, #526FE8);
  border-style: dashed;
}

.drag-handle {
  cursor: grab;
  color: var(--text-secondary, #8A93A8);
  font-size: 18px;
  display: flex;
  align-items: center;
}

.drag-handle:active {
  cursor: grabbing;
}

.item-index {
  width: 24px;
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-secondary, #667085);
}

.item-preview {
  width: 96px;
  height: 96px;
  flex-shrink: 0;
  border-radius: 6px;
  overflow: hidden;
  background: var(--bg-secondary, #F5F7FC);
}

.item-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preset-deleted {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 11px;
  color: var(--text-secondary, #8A93A8);
}

.preset-deleted .anticon {
  font-size: 28px;
}

.item-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #29365C);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-desc {
  font-size: 13px;
  color: var(--text-secondary, #667085);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.item-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;
}

.preset-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px;
  background: var(--bg-secondary, #F5F7FC);
  border-radius: 6px;
}

.preset-preview img {
  width: 56px;
  height: 56px;
  object-fit: cover;
  border-radius: 4px;
}

.preset-preview span {
  flex: 1;
  font-size: 13px;
  color: var(--text-secondary, #667085);
}

.preset-preview-placeholder {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-card, #fff);
  border-radius: 4px;
  color: var(--text-secondary, #8A93A8);
  font-size: 24px;
}
</style>
