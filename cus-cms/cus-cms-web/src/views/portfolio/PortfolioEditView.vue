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
            <ArrowLeftOutlined /> 返回列表
          </a-button>
          <div class="header-actions">
            <a-button :loading="saving" @click="handleSave">保存信息</a-button>
          </div>
        </div>
        <div class="header-form">
          <a-input
            v-model:value="form.name"
            placeholder="作品集名称"
            size="large"
            class="name-input"
          />
          <a-textarea
            v-model:value="form.description"
            placeholder="作品集介绍（可选）"
            :rows="2"
            :maxlength="500"
            class="desc-input"
          />
        </div>
      </div>

      <!-- 工具栏 -->
      <div class="items-toolbar">
        <span class="items-title">
          作品列表 <span class="items-count">({{ items.length }})</span>
        </span>
        <a-button type="primary" @click="openAddPicker">
          <PlusOutlined /> 添加作品
        </a-button>
      </div>

      <!-- 作品项列表 -->
      <a-spin :spinning="actionLoading">
        <a-empty
          v-if="items.length === 0"
          description="还没有作品，点击「添加作品」开始"
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
                <span>预设已删除</span>
              </div>
            </div>
            <div class="item-content">
              <div class="item-title" :title="item.title">{{ item.title }}</div>
              <div class="item-desc">
                {{ item.description || '暂无介绍' }}
              </div>
            </div>
            <div class="item-actions">
              <a-button type="link" size="small" @click="openEditItem(item)">
                <EditOutlined /> 编辑
              </a-button>
              <a-button type="link" size="small" danger @click="handleDeleteItem(item)">
                <DeleteOutlined /> 删除
              </a-button>
            </div>
          </div>
        </div>
      </a-spin>

      <!-- 添加作品选择器 -->
      <PortfolioItemPicker
        v-model:visible="pickerVisible"
        title="添加作品"
        @selected="handleAddItem"
      />

      <!-- 编辑作品项弹窗 -->
      <a-modal
        v-model:open="editItemVisible"
        title="编辑作品"
        :confirm-loading="savingItem"
        @ok="handleSaveItem"
      >
        <a-form layout="vertical">
          <a-form-item label="作品名称" required>
            <a-input
              v-model:value="editItemForm.title"
              placeholder="请输入作品名称"
              :maxlength="100"
              show-count
            />
          </a-form-item>
          <a-form-item label="作品介绍">
            <a-textarea
              v-model:value="editItemForm.description"
              placeholder="请输入作品介绍（可选）"
              :rows="4"
              :maxlength="500"
              show-count
            />
          </a-form-item>
          <a-form-item label="关联预设">
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
              <a-button type="link" size="small" @click="openReselectPreset">重新选择</a-button>
            </div>
            <a-button v-else type="dashed" block @click="openReselectPreset">
              <SwapOutlined /> 选择预设
            </a-button>
          </a-form-item>
        </a-form>
      </a-modal>

      <!-- 重选预设选择器 -->
      <PortfolioItemPicker
        v-model:visible="reselectPickerVisible"
        title="重新选择预设"
        @selected="handleReselectPreset"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
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

const router = useRouter()
const route = useRoute()

const portfolioId = computed(() => String(route.params.id))

const loading = ref(true)
const actionLoading = ref(false)
const saving = ref(false)
const detail = ref<PortfolioDetail | null>(null)
const items = ref<PortfolioItem[]>([])

const form = reactive({
  name: '',
  description: '',
})

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
    return { name: '当前预设', output_url: editing.output_url }
  }
  if (editing && editing.preset_id) {
    return { name: '预设已删除', output_url: '' }
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
    items.value = [...(res.items || [])].sort((a, b) => a.sort_order - b.sort_order)
  } catch {
    // 错误由拦截器处理
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  if (!form.name.trim()) {
    message.warning('请输入作品集名称')
    return
  }
  saving.value = true
  try {
    await portfolioApi.update(portfolioId.value, {
      name: form.name.trim(),
      description: form.description.trim(),
    })
    message.success('保存成功')
  } catch {
    // 错误由拦截器处理
  } finally {
    saving.value = false
  }
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
    message.success('添加成功')
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
    name: '新预设（保存后显示预览）',
    output_url: '',
  }
  editItemForm.preset_id = payload.preset_id
  message.success('已选择新预设，请点击确定保存')
}

async function handleSaveItem() {
  if (!editingItem.value) return
  if (!editItemForm.title.trim()) {
    message.warning('请输入作品名称')
    return
  }
  savingItem.value = true
  try {
    await portfolioApi.updateItem(portfolioId.value, editingItem.value.id, {
      title: editItemForm.title.trim(),
      description: editItemForm.description.trim() || undefined,
      preset_id: editItemForm.preset_id,
    })
    message.success('保存成功')
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
    title: '确认删除',
    content: `删除作品「${item.title}」后无法恢复，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await portfolioApi.removeItem(portfolioId.value, item.id)
        message.success('删除成功')
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
    message.success('排序已更新')
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
  border: 1px solid var(--border-color, #f0f0f0);
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

.items-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.items-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #262626);
}

.items-count {
  font-size: 13px;
  font-weight: normal;
  color: var(--text-secondary, #8c8c8c);
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
  border: 1px solid var(--border-color, #f0f0f0);
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
  border-color: var(--primary, #4a6cf7);
  border-style: dashed;
}

.drag-handle {
  cursor: grab;
  color: var(--text-secondary, #bfbfbf);
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
  color: var(--text-secondary, #8c8c8c);
}

.item-preview {
  width: 96px;
  height: 96px;
  flex-shrink: 0;
  border-radius: 6px;
  overflow: hidden;
  background: var(--bg-secondary, #f5f5f5);
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
  color: var(--text-secondary, #bfbfbf);
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
  color: var(--text-primary, #262626);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-desc {
  font-size: 13px;
  color: var(--text-secondary, #8c8c8c);
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
  background: var(--bg-secondary, #f5f5f5);
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
  color: var(--text-secondary, #595959);
}

.preset-preview-placeholder {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-card, #fff);
  border-radius: 4px;
  color: var(--text-secondary, #bfbfbf);
  font-size: 24px;
}
</style>
