<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ t('project.list.title') }}</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> {{ t('project.list.create') }}
      </a-button>
    </div>

    <!-- 分类标签栏 -->
    <CategoryBar type="project" @select="handleCategorySelect" />

    <a-card>
      <div class="project-toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="t('project.list.searchPlaceholder')"
          style="width: 220px"
          allow-clear
        />
        <a-select
          v-model:value="filterStatus"
          :placeholder="t('project.list.allStatus')"
          style="width: 120px"
          allow-clear
          :options="statusOptions"
        />
      </div>

      <a-spin :spinning="loading">
        <a-empty
          v-if="!loading && list.length === 0"
          :description="t('project.list.empty')"
          style="margin-top: 48px"
        />
        <div v-else class="project-grid">
          <div v-for="item in list" :key="item.id" class="project-card">
            <div class="project-cover">
              <img
                v-if="item.cover_url"
                :src="getThumbUrl(item.cover_url, 640)"
                :alt="item.title"
                referrerpolicy="no-referrer"
              />
              <div v-else class="cover-placeholder">
                <ProjectOutlined />
              </div>
              <a-tag v-if="item.category_name || item.category" class="category-tag" color="blue">
                {{ item.category_name || item.category }}
              </a-tag>
              <a-tag v-else class="category-tag">{{ t('project.list.uncategorized') }}</a-tag>
            </div>
            <div class="project-info">
              <div class="project-title-row">
                <div class="project-title" :title="item.title">{{ item.title }}</div>
                <a-tag :color="item.status === 1 ? 'green' : 'default'" class="status-tag">
                  {{ item.status === 1 ? t('project.status.published') : t('project.status.draft') }}
                </a-tag>
              </div>
              <div class="project-sub">
                <template v-if="item.role || item.client">
                  <span v-if="item.role">{{ item.role }}</span>
                  <span v-if="item.role && item.client">·</span>
                  <span v-if="item.client">{{ item.client }}</span>
                </template>
              </div>
              <div class="project-summary" :title="item.summary">
                {{ item.summary || t('project.list.noSummary') }}
              </div>
              <div v-if="splitTechStack(item.tech_stack).length" class="project-tech">
                <a-tag
                  v-for="tech in splitTechStack(item.tech_stack).slice(0, 4)"
                  :key="tech"
                  class="tech-tag"
                >
                  {{ tech }}
                </a-tag>
                <a-tag v-if="splitTechStack(item.tech_stack).length > 4" class="tech-tag">
                  +{{ splitTechStack(item.tech_stack).length - 4 }}
                </a-tag>
              </div>
              <div class="project-meta">
                <span class="meta-item">
                  <CalendarOutlined />
                  {{ formatRange(item.start_date, item.end_date) }}
                </span>
                <span class="meta-links">
                  <a
                    v-if="item.project_url"
                    :href="item.project_url"
                    target="_blank"
                    rel="noopener noreferrer"
                    :title="t('project.list.projectLinkTitle')"
                  >
                    <LinkOutlined />
                  </a>
                  <a
                    v-if="item.repo_url"
                    :href="item.repo_url"
                    target="_blank"
                    rel="noopener noreferrer"
                    :title="t('project.list.repoTitle')"
                  >
                    <GithubOutlined />
                  </a>
                </span>
              </div>
              <div class="project-actions">
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
          :show-total="(count: number) => t('project.list.total', { total: count })"
          @change="handlePageChange"
          @show-size-change="handlePageChange"
        />
      </div>
    </a-card>

    <!-- 新建/编辑项目弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalMode === 'create' ? t('project.list.create') : t('project.list.edit')"
      :confirm-loading="modalLoading"
      width="720px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item :label="t('project.form.titleLabel')" required>
          <a-input
            v-model:value="form.title"
            :placeholder="t('project.form.titlePlaceholder')"
            :maxlength="255"
            show-count
          />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item :label="t('common.field.category')">
              <CategorySelect v-model="form.category_id" type="project" :placeholder="t('project.form.categoryPlaceholder')" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item :label="t('project.form.roleLabel')">
              <a-input
                v-model:value="form.role"
                :placeholder="t('project.form.rolePlaceholder')"
                :maxlength="100"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item :label="t('project.form.clientLabel')">
          <a-input
            v-model:value="form.client"
            :placeholder="t('project.form.clientPlaceholder')"
            :maxlength="255"
          />
        </a-form-item>
        <a-form-item :label="t('project.form.coverLabel')">
          <div class="image-picker">
            <div v-if="form.cover_url" class="image-preview">
              <img
                :src="getThumbUrl(form.cover_url, 400)"
                :alt="t('project.form.coverAlt')"
                referrerpolicy="no-referrer"
              />
              <a-button size="small" @click="mediaPickerVisible = true">
                <PictureOutlined /> {{ t('project.form.changeCover') }}
              </a-button>
            </div>
            <a-button v-else @click="mediaPickerVisible = true">
              <PictureOutlined /> {{ t('project.form.selectCover') }}
            </a-button>
          </div>
        </a-form-item>
        <a-form-item :label="t('project.form.summaryLabel')">
          <a-textarea
            v-model:value="form.summary"
            :placeholder="t('project.form.summaryPlaceholder')"
            :rows="2"
            :maxlength="500"
            show-count
          />
        </a-form-item>
        <a-form-item :label="t('project.form.descLabel')">
          <a-textarea
            v-model:value="form.description"
            :placeholder="t('project.form.descPlaceholder')"
            :rows="5"
          />
        </a-form-item>
        <a-form-item :label="t('project.form.techLabel')">
          <a-select
            v-model:value="form.techStack"
            mode="tags"
            :placeholder="t('project.form.techPlaceholder')"
            :max-tag-count="8"
          />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item :label="t('project.form.startDateLabel')">
              <a-date-picker
                v-model:value="form.start_date"
                picker="month"
                value-format="YYYY-MM"
                :placeholder="t('project.form.monthPlaceholder')"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item :label="t('project.form.endDateLabel')">
              <a-date-picker
                v-model:value="form.end_date"
                picker="month"
                value-format="YYYY-MM"
                :placeholder="t('project.form.endDatePlaceholder')"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item :label="t('project.form.projectUrlLabel')">
          <a-input
            v-model:value="form.project_url"
            :placeholder="t('project.form.urlPlaceholder')"
            :maxlength="1024"
          />
        </a-form-item>
        <a-form-item :label="t('project.form.repoUrlLabel')">
          <a-input
            v-model:value="form.repo_url"
            :placeholder="t('project.form.repoUrlPlaceholder')"
            :maxlength="1024"
          />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item :label="t('common.field.status')">
              <a-radio-group v-model:value="form.status">
                <a-radio :value="0">{{ t('project.status.draft') }}</a-radio>
                <a-radio :value="1">{{ t('project.status.published') }}</a-radio>
              </a-radio-group>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item :label="t('project.form.sortWeight')">
              <a-input-number
                v-model:value="form.sort_order"
                :min="0"
                :precision="0"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-modal>

    <MediaPicker v-model:visible="mediaPickerVisible" module="project" @selected="handleMediaSelected" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ProjectOutlined,
  PictureOutlined,
  CalendarOutlined,
  LinkOutlined,
  GithubOutlined,
} from '@ant-design/icons-vue'
import { projectApi } from '@/api/project'
import type { Project } from '@/types/project'
import { getThumbUrl } from '@/utils/image'
import MediaPicker from '@/components/media/MediaPicker.vue'
import CategoryBar from '@/components/common/CategoryBar.vue'
import CategorySelect from '@/components/article/CategorySelect.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const { t } = useI18n()

const statusOptions = computed(() => [
  { value: 0, label: t('project.status.draft') },
  { value: 1, label: t('project.status.published') },
])

const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const filterStatus = ref<number | undefined>(undefined)
const categoryFilter = ref<string | undefined>(undefined)

const list = ref<Project[]>([])
const loading = ref(false)

const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(12)

const modalVisible = ref(false)
const modalLoading = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref('')

const mediaPickerVisible = ref(false)

const form = reactive({
  title: '',
  category_id: undefined as string | undefined,
  role: '',
  client: '',
  cover_url: '',
  coverMediaId: '',
  coverPresetId: '',
  summary: '',
  description: '',
  techStack: [] as string[],
  start_date: undefined as string | undefined,
  end_date: undefined as string | undefined,
  project_url: '',
  repo_url: '',
  status: 0,
  sort_order: 0,
})

watch(keyword, setDebounce)

watch([debouncedKeyword, filterStatus], () => {
  reset()
  fetchList()
})

onMounted(() => {
  fetchList()
})

async function fetchList() {
  loading.value = true
  try {
    const res = await projectApi.getList({
      page: page.value,
      page_size: pageSize.value,
      keyword: debouncedKeyword.value || undefined,
      category_id: categoryFilter.value,
      status: filterStatus.value,
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

function handleCategorySelect(id: string | undefined) {
  categoryFilter.value = id
  reset()
  fetchList()
}

function handlePageChange(newPage: number, newPageSize?: number) {
  changePagination(newPage, newPageSize)
  fetchList()
}

function resetForm() {
  form.title = ''
  form.category_id = undefined
  form.role = ''
  form.client = ''
  form.cover_url = ''
  form.coverMediaId = ''
  form.coverPresetId = ''
  form.summary = ''
  form.description = ''
  form.techStack = []
  form.start_date = undefined
  form.end_date = undefined
  form.project_url = ''
  form.repo_url = ''
  form.status = 0
  form.sort_order = 0
}

function openCreateModal() {
  modalMode.value = 'create'
  editingId.value = ''
  resetForm()
  modalVisible.value = true
}

function handleEdit(item: Project) {
  modalMode.value = 'edit'
  editingId.value = item.id
  resetForm()
  form.title = item.title
  form.category_id = item.category_id || undefined
  form.role = item.role || ''
  form.client = item.client || ''
  form.cover_url = item.cover_url || ''
  form.coverMediaId = item.cover_media_id || ''
  form.coverPresetId = item.cover_preset_id || ''
  form.summary = item.summary || ''
  form.description = item.description || ''
  form.techStack = splitTechStack(item.tech_stack)
  form.start_date = item.start_date || undefined
  form.end_date = item.end_date || undefined
  form.project_url = item.project_url || ''
  form.repo_url = item.repo_url || ''
  form.status = item.status
  form.sort_order = item.sort_order
  modalVisible.value = true
}

function handleMediaSelected(media: { id: string; url: string; presetId?: string }) {
  form.coverMediaId = media.id
  form.coverPresetId = media.presetId || ''
  form.cover_url = media.url // 仅展示；后端按 ID 优先解析
}

function splitTechStack(techStack?: string): string[] {
  if (!techStack) return []
  return techStack.split(',').map((s) => s.trim()).filter(Boolean)
}

function formatRange(start?: string, end?: string): string {
  if (!start && !end) return t('project.meta.timeTbd')
  const endText = end || t('project.meta.present')
  return `${start || '?'} ~ ${endText}`
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning(t('project.form.titleRequired'))
    return
  }
  if (form.start_date && form.end_date && form.start_date > form.end_date) {
    message.warning(t('project.message.dateRangeInvalid'))
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await projectApi.create({
        title: form.title.trim(),
        category_id: form.category_id,
        role: form.role.trim() || undefined,
        client: form.client.trim() || undefined,
        cover_url: form.cover_url.trim() || undefined,
        cover_media_id: form.coverMediaId || undefined,
        cover_preset_id: form.coverPresetId || undefined,
        summary: form.summary.trim() || undefined,
        description: form.description.trim() || undefined,
        tech_stack: form.techStack.join(',') || undefined,
        start_date: form.start_date || undefined,
        end_date: form.end_date || undefined,
        project_url: form.project_url.trim() || undefined,
        repo_url: form.repo_url.trim() || undefined,
        status: form.status,
        sort_order: form.sort_order,
      })
      message.success(t('common.message.createSuccess'))
    } else {
      await projectApi.update(editingId.value, {
        title: form.title.trim(),
        category_id: form.category_id,
        role: form.role.trim(),
        client: form.client.trim(),
        cover_url: form.cover_url.trim(),
        cover_media_id: form.coverMediaId || undefined,
        cover_preset_id: form.coverPresetId || undefined,
        summary: form.summary.trim(),
        description: form.description.trim(),
        tech_stack: form.techStack.join(','),
        // 更新时空字符串表示清空该字段
        start_date: form.start_date || '',
        end_date: form.end_date || '',
        project_url: form.project_url.trim(),
        repo_url: form.repo_url.trim(),
        status: form.status,
        sort_order: form.sort_order,
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

function handleDelete(item: Project) {
  Modal.confirm({
    title: t('project.message.deleteTitle'),
    content: t('project.message.deleteConfirm', { name: item.title }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await projectApi.remove(item.id)
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
.project-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.project-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.project-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: box-shadow 0.2s;
}

.project-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.project-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  background: var(--bg-secondary, #F5F7FC);
}

.project-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}

.project-card:hover .project-cover img {
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

.category-tag {
  position: absolute;
  top: 8px;
  left: 8px;
  margin: 0;
}

.project-info {
  padding: 12px 16px 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.project-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.project-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #29365C);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.status-tag {
  flex-shrink: 0;
  margin: 0;
}

.project-sub {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-secondary, #667085);
}

.project-summary {
  font-size: 13px;
  color: var(--text-secondary, #667085);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 39px;
}

.project-tech {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tech-tag {
  margin: 0;
  font-size: 12px;
}

.project-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: var(--text-secondary, #8A93A8);
  margin-top: 4px;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.meta-links {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
}

.meta-links a {
  color: var(--text-secondary, #667085);
  transition: color 0.2s;
}

.meta-links a:hover {
  color: #526FE8;
}

.project-actions {
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
  width: 120px;
  height: 68px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid var(--border-color, #E5E9F2);
}
</style>
