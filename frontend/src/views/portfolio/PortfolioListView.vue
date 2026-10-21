<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ t('portfolio.list.title') }}</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> {{ t('portfolio.list.create') }}
      </a-button>
    </div>

    <!-- 分类标签栏 -->
    <CategoryBar type="portfolio" @select="handleCategorySelect" />

    <a-card>
      <div class="portfolio-toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="t('portfolio.list.searchPlaceholder')"
          style="width: 240px"
          allow-clear
        />
      </div>

      <a-spin :spinning="loading">
        <a-empty
          v-if="!loading && list.length === 0"
          :description="t('portfolio.list.empty')"
          style="margin-top: 48px"
        />
        <div v-else class="portfolio-grid">
          <div v-for="item in list" :key="item.id" class="portfolio-card">
            <div class="portfolio-cover" @click="handleEdit(item.id)">
              <img
                v-if="item.cover_url"
                :src="getThumbUrl(item.cover_url, 400)"
                :alt="item.name"
              />
              <div v-else class="cover-placeholder">
                <PictureOutlined />
              </div>
              <div class="cover-overlay">
                <span>{{ t('common.actions.edit') }}</span>
              </div>
            </div>
            <div class="portfolio-info">
              <div class="portfolio-name" :title="item.name">{{ item.name }}</div>
              <div class="portfolio-desc" :title="item.description">
                {{ item.description || t('portfolio.item.noDesc') }}
              </div>
              <div v-if="item.category_name" class="portfolio-category">
                <a-tag color="blue" :bordered="false">{{ item.category_name }}</a-tag>
              </div>
              <div class="portfolio-meta">
                <span class="meta-item">
                  <PictureOutlined /> {{ t('portfolio.list.itemCount', { count: item.item_count }) }}
                </span>
                <span class="meta-item">{{ formatDateTime(item.created_at) }}</span>
              </div>
              <div class="portfolio-actions">
                <a-button type="link" size="small" @click="handleEdit(item.id)">
                  <EditOutlined /> {{ t('common.actions.edit') }}
                </a-button>
                <PushSyncButton
                  target-type="portfolio"
                  :target-id="item.id"
                  :published="item.status === 1"
                  @updated="fetchList"
                />
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
          :show-total="(count: number) => t('portfolio.list.total', { total: count })"
          @change="handlePageChange"
          @show-size-change="handlePageChange"
        />
      </div>
    </a-card>

    <!-- 新建作品集弹窗 -->
    <a-modal
      v-model:open="createVisible"
      :title="t('portfolio.list.create')"
      :confirm-loading="creating"
      @ok="handleCreate"
    >
      <a-form layout="vertical">
        <a-form-item :label="t('common.field.name')" required>
          <a-input
            v-model:value="createForm.name"
            :placeholder="t('portfolio.form.namePlaceholder')"
            :maxlength="100"
            show-count
          />
        </a-form-item>
        <a-form-item :label="t('portfolio.form.introLabel')">
          <a-textarea
            v-model:value="createForm.description"
            :placeholder="t('portfolio.form.introPlaceholder')"
            :rows="4"
            :maxlength="500"
            show-count
          />
        </a-form-item>
        <a-form-item :label="t('common.field.category')">
          <CategorySelect v-model="createForm.category_id" type="portfolio" :placeholder="t('portfolio.form.categoryPlaceholder')" />
        </a-form-item>
        <a-form-item :label="t('portfolio.cover.setting')">
          <a-radio-group v-model:value="createForm.cover_mode">
            <a-radio :value="0">{{ t('portfolio.cover.useFirst') }}</a-radio>
            <a-radio :value="1">{{ t('portfolio.cover.custom') }}</a-radio>
          </a-radio-group>
          <div v-if="createForm.cover_mode === 1" class="create-cover-custom">
            <div v-if="createCoverPreviewUrl" class="create-cover-preview">
              <img :src="createCoverPreviewUrl" :alt="t('portfolio.cover.previewAlt')" />
              <div class="create-cover-actions">
                <a-button type="link" size="small" @click="openCreateCoverPicker">{{ t('portfolio.cover.change') }}</a-button>
                <a-button type="link" size="small" danger @click="clearCreateCoverPreset">{{ t('portfolio.cover.clear') }}</a-button>
              </div>
            </div>
            <a-button v-else type="dashed" @click="openCreateCoverPicker">
              <PictureOutlined /> {{ t('portfolio.cover.selectPreset') }}
            </a-button>
          </div>
          <div v-else class="create-cover-tip">
            {{ t('portfolio.cover.autoTip') }}
          </div>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 封面预设选择器 -->
    <PortfolioItemPicker
      v-model:visible="createCoverPickerVisible"
      :title="t('portfolio.cover.selectPreset')"
      preset-only
      @selected="handleSelectCreateCover"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import PushSyncButton from '@/components/community/PushSyncButton.vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PictureOutlined,
} from '@ant-design/icons-vue'
import { portfolioApi } from '@/api/portfolio'
import type { Portfolio, CreatePortfolioReq } from '@/types/portfolio'
import { getThumbUrl } from '@/utils/image'
import { formatDateTime } from '@/utils/datetime'
import PortfolioItemPicker from '@/components/portfolio/PortfolioItemPicker.vue'
import CategoryBar from '@/components/common/CategoryBar.vue'
import CategorySelect from '@/components/article/CategorySelect.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const router = useRouter()
const { t } = useI18n()

const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const list = ref<Portfolio[]>([])
const loading = ref(false)
const selectedCategory = ref<string | undefined>(undefined)

const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(12)

const createVisible = ref(false)
const creating = ref(false)
const createForm = ref<CreatePortfolioReq>({ name: '', description: '', cover_mode: 0, category_id: undefined })
const createCoverPresetId = ref<string | null>(null)
const createCoverPreviewUrl = ref('')
const createCoverPickerVisible = ref(false)

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
    const res = await portfolioApi.getList({
      page: page.value,
      page_size: pageSize.value,
      keyword: debouncedKeyword.value || undefined,
      category_id: selectedCategory.value,
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

function openCreateModal() {
  createForm.value = { name: '', description: '', cover_mode: 0, category_id: undefined }
  createCoverPresetId.value = null
  createCoverPreviewUrl.value = ''
  createVisible.value = true
}

async function handleCreate() {
  if (!createForm.value.name.trim()) {
    message.warning(t('portfolio.form.nameRequired'))
    return
  }
  if (createForm.value.cover_mode === 1 && !createCoverPresetId.value) {
    message.warning(t('portfolio.cover.presetRequired'))
    return
  }
  creating.value = true
  try {
    const res = await portfolioApi.create({
      name: createForm.value.name.trim(),
      description: createForm.value.description?.trim() || undefined,
      cover_mode: createForm.value.cover_mode,
      cover_preset_id:
        createForm.value.cover_mode === 1
          ? createCoverPresetId.value || undefined
          : undefined,
      category_id: createForm.value.category_id,
    })
    message.success(t('common.message.createSuccess'))
    createVisible.value = false
    router.push(`/portfolios/${res.id}/edit`)
  } catch {
    // 错误由拦截器处理
  } finally {
    creating.value = false
  }
}

function openCreateCoverPicker() {
  createCoverPickerVisible.value = true
}

function handleSelectCreateCover(payload: {
  preset_id: string
  output_url?: string
  preset_name?: string
}) {
  createCoverPresetId.value = payload.preset_id
  createCoverPreviewUrl.value = payload.output_url || ''
  createCoverPickerVisible.value = false
}

function clearCreateCoverPreset() {
  createCoverPresetId.value = null
  createCoverPreviewUrl.value = ''
}

function handleCategorySelect(id: string | undefined) {
  selectedCategory.value = id
  reset()
  fetchList()
}

function handleEdit(id: string) {
  router.push(`/portfolios/${id}/edit`)
}

function handleDelete(item: Portfolio) {
  Modal.confirm({
    title: t('portfolio.list.deleteTitle'),
    content: t('portfolio.list.deleteConfirm', { name: item.name }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await portfolioApi.remove(item.id)
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
.portfolio-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.portfolio-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.portfolio-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: box-shadow 0.2s;
}

.portfolio-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.portfolio-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 10;
  overflow: hidden;
  cursor: pointer;
  background: var(--bg-secondary, #F5F7FC);
}

.portfolio-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}

.portfolio-card:hover .portfolio-cover img {
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

.cover-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  color: #fff;
  font-size: 14px;
  opacity: 0;
  transition: opacity 0.2s;
}

.portfolio-cover:hover .cover-overlay {
  opacity: 1;
}

.portfolio-info {
  padding: 12px 16px 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.portfolio-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #29365C);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.portfolio-desc {
  font-size: 13px;
  color: var(--text-secondary, #667085);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 39px;
}

.portfolio-category {
  margin-top: 2px;
}

.portfolio-meta {
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

.portfolio-actions {
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

.create-cover-custom {
  display: flex;
  align-items: center;
  margin-top: 8px;
}

.create-cover-preview {
  display: flex;
  align-items: center;
  gap: 12px;
}

.create-cover-preview img {
  width: 96px;
  height: 64px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid var(--border-color, #E5E9F2);
}

.create-cover-actions {
  display: flex;
  flex-direction: column;
}

.create-cover-tip {
  margin-top: 8px;
  font-size: 13px;
  color: var(--text-secondary, #667085);
}
</style>
