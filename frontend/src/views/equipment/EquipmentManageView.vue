<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ t('menu.equipment.list') }}</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> {{ t('equipment.actions.create') }}
      </a-button>
    </div>

    <a-card>
      <div class="equipment-toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="t('equipment.list.searchPlaceholder')"
          style="width: 240px"
          allow-clear
        />
      </div>

      <a-spin :spinning="loading">
        <a-empty
          v-if="!loading && list.length === 0"
          :description="t('equipment.list.empty')"
          style="margin-top: 48px"
        />
        <div v-else class="equipment-grid">
          <div v-for="item in list" :key="item.id" class="equipment-card">
            <div class="equipment-cover">
              <img
                v-if="item.image_url"
                :src="getThumbUrl(item.image_url, 400)"
                :alt="item.name"
                referrerpolicy="no-referrer"
              />
              <div v-else class="cover-placeholder">
                <LaptopOutlined />
              </div>
            </div>
            <div class="equipment-info">
              <div class="equipment-title" :title="item.name">{{ item.name }}</div>
              <div class="equipment-brand">
                <span class="brand-badge" :class="{ 'no-brand': !item.brand }">
                  {{ item.brand || t('equipment.list.noBrand') }}
                </span>
              </div>
              <div class="equipment-desc" :title="item.description">
                {{ item.description || t('equipment.list.noDescription') }}
              </div>
              <div class="equipment-meta">
                <span class="meta-item">{{ formatDateTime(item.created_at) }}</span>
              </div>
              <div class="equipment-actions">
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
          :show-total="(count: number) => t('equipment.list.total', { total: count })"
          @change="handlePageChange"
          @show-size-change="handlePageChange"
        />
      </div>
    </a-card>

    <!-- 新建/编辑设备弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalMode === 'create' ? t('equipment.actions.create') : t('equipment.actions.edit')"
      :confirm-loading="modalLoading"
      width="640px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item :label="t('common.field.name')" required>
          <a-input
            v-model:value="form.name"
            :placeholder="t('equipment.form.namePlaceholder')"
            :maxlength="100"
            show-count
          />
        </a-form-item>
        <a-form-item :label="t('common.field.image')">
          <div class="image-picker">
            <div v-if="form.image_url" class="image-preview">
              <img
                :src="getThumbUrl(form.image_url, 400)"
                :alt="t('equipment.form.imageAlt')"
                referrerpolicy="no-referrer"
              />
              <a-button size="small" @click="mediaPickerVisible = true">
                <PictureOutlined /> {{ t('equipment.form.changeImage') }}
              </a-button>
            </div>
            <a-button v-else @click="mediaPickerVisible = true">
              <PictureOutlined /> {{ t('equipment.form.selectImage') }}
            </a-button>
          </div>
        </a-form-item>
        <a-form-item :label="t('equipment.form.brand')">
          <a-input
            v-model:value="form.brand"
            :placeholder="t('equipment.form.brandPlaceholder')"
            :maxlength="100"
          />
        </a-form-item>
        <a-form-item :label="t('common.field.description')">
          <a-textarea
            v-model:value="form.description"
            :placeholder="t('equipment.form.descriptionPlaceholder')"
            :rows="3"
            :maxlength="500"
            show-count
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <MediaPicker v-model:visible="mediaPickerVisible" module="equipment" @selected="handleMediaSelected" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  LaptopOutlined,
  PictureOutlined,
} from '@ant-design/icons-vue'
import { equipmentApi } from '@/api/equipment'
import type { Equipment } from '@/types/equipment'
import { getThumbUrl } from '@/utils/image'
import MediaPicker from '@/components/media/MediaPicker.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'
import { formatDateTime } from '@/utils/datetime'

const { t } = useI18n()

const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const list = ref<Equipment[]>([])
const loading = ref(false)

const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(12)

const modalVisible = ref(false)
const modalLoading = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref('')

const mediaPickerVisible = ref(false)

const form = reactive({
  name: '',
  image_url: '',
  brand: '',
  description: '',
})

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
    const res = await equipmentApi.getList({
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
  form.name = ''
  form.image_url = ''
  form.brand = ''
  form.description = ''
}

function openCreateModal() {
  modalMode.value = 'create'
  editingId.value = ''
  resetForm()
  modalVisible.value = true
}

function handleEdit(item: Equipment) {
  modalMode.value = 'edit'
  editingId.value = item.id
  resetForm()
  form.name = item.name
  form.image_url = item.image_url || ''
  form.brand = item.brand || ''
  form.description = item.description || ''
  modalVisible.value = true
}

function handleMediaSelected(media: { url: string }) {
  form.image_url = media.url
}

async function handleSubmit() {
  if (!form.name.trim()) {
    message.warning(t('equipment.message.nameRequired'))
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await equipmentApi.create({
        name: form.name.trim(),
        image_url: form.image_url.trim() || undefined,
        brand: form.brand.trim() || undefined,
        description: form.description.trim() || undefined,
      })
      message.success(t('common.message.createSuccess'))
    } else {
      await equipmentApi.update(editingId.value, {
        name: form.name.trim(),
        image_url: form.image_url.trim() || undefined,
        brand: form.brand.trim() || undefined,
        description: form.description.trim() || undefined,
      })
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

function handleDelete(item: Equipment) {
  Modal.confirm({
    title: t('equipment.message.deleteTitle'),
    content: t('equipment.message.deleteConfirm', { name: item.name }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await equipmentApi.remove(item.id)
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
.equipment-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.equipment-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 20px;
}

.equipment-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: box-shadow 0.2s;
}

.equipment-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.equipment-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  overflow: hidden;
  background: var(--bg-secondary, #F5F7FC);
}

.equipment-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}

.equipment-card:hover .equipment-cover img {
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

.equipment-info {
  padding: 12px 16px 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.equipment-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #29365C);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.equipment-brand {
  display: flex;
  align-items: center;
}

.brand-badge {
  display: inline-block;
  padding: 2px 8px;
  font-size: 12px;
  border-radius: 4px;
  background: rgba(82, 111, 232, 0.1);
  color: #526FE8;
}

.brand-badge.no-brand {
  background: var(--bg-secondary, #F5F7FC);
  color: var(--text-secondary, #8A93A8);
}

.equipment-desc {
  font-size: 13px;
  color: var(--text-secondary, #667085);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 39px;
}

.equipment-meta {
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

.equipment-actions {
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

.image-picker {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.image-preview {
  display: flex;
  align-items: center;
  gap: 12px;
}

.image-preview img {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid var(--border-color, #E5E9F2);
}
</style>
