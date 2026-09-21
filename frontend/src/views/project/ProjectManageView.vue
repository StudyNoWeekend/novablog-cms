<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">项目经历</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> 新建项目
      </a-button>
    </div>

    <a-card>
      <div class="project-toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索项目名称/简介..."
          style="width: 220px"
          allow-clear
        />
        <a-select
          v-model:value="filterCategory"
          placeholder="全部领域"
          style="width: 150px"
          allow-clear
          :options="filterCategoryOptions"
        />
        <a-select
          v-model:value="filterStatus"
          placeholder="全部状态"
          style="width: 120px"
          allow-clear
          :options="statusOptions"
        />
      </div>

      <a-spin :spinning="loading">
        <a-empty
          v-if="!loading && list.length === 0"
          description="还没有项目经历，开始添加吧"
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
              <a-tag v-if="item.category" class="category-tag" color="blue">
                {{ item.category }}
              </a-tag>
              <a-tag v-else class="category-tag">未分类</a-tag>
            </div>
            <div class="project-info">
              <div class="project-title-row">
                <div class="project-title" :title="item.title">{{ item.title }}</div>
                <a-tag :color="item.status === 1 ? 'green' : 'default'" class="status-tag">
                  {{ item.status === 1 ? '已发布' : '草稿' }}
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
                {{ item.summary || '暂无简介' }}
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
                    title="项目/作品链接"
                  >
                    <LinkOutlined />
                  </a>
                  <a
                    v-if="item.repo_url"
                    :href="item.repo_url"
                    target="_blank"
                    rel="noopener noreferrer"
                    title="代码仓库"
                  >
                    <GithubOutlined />
                  </a>
                </span>
              </div>
              <div class="project-actions">
                <a-button type="link" size="small" @click="handleEdit(item)">
                  <EditOutlined /> 编辑
                </a-button>
                <a-button type="link" size="small" danger @click="handleDelete(item)">
                  <DeleteOutlined /> 删除
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
          :show-total="(t: number) => `共 ${t} 个项目`"
          @change="handlePageChange"
          @show-size-change="handlePageChange"
        />
      </div>
    </a-card>

    <!-- 新建/编辑项目弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalMode === 'create' ? '新建项目' : '编辑项目'"
      :confirm-loading="modalLoading"
      width="720px"
      @ok="handleSubmit"
    >
      <a-form layout="vertical">
        <a-form-item label="项目名称" required>
          <a-input
            v-model:value="form.title"
            placeholder="请输入项目名称"
            :maxlength="255"
            show-count
          />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="领域分类">
              <a-auto-complete
                v-model:value="form.category"
                :options="presetCategoryOptions"
                placeholder="摄影 / 视频剪辑 / 技术开发，可自定义"
                allow-clear
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="担任角色">
              <a-input
                v-model:value="form.role"
                placeholder="如：摄影师 / 剪辑师 / 全栈开发"
                :maxlength="100"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="客户 / 所属组织">
          <a-input
            v-model:value="form.client"
            placeholder="请输入客户或所属组织（可选）"
            :maxlength="255"
          />
        </a-form-item>
        <a-form-item label="封面图">
          <div class="image-picker">
            <div v-if="form.cover_url" class="image-preview">
              <img
                :src="getThumbUrl(form.cover_url, 400)"
                alt="项目封面"
                referrerpolicy="no-referrer"
              />
              <a-button size="small" @click="mediaPickerVisible = true">
                <PictureOutlined /> 更换封面
              </a-button>
            </div>
            <a-button v-else @click="mediaPickerVisible = true">
              <PictureOutlined /> 选择封面
            </a-button>
          </div>
        </a-form-item>
        <a-form-item label="一句话简介">
          <a-textarea
            v-model:value="form.summary"
            placeholder="用一句话介绍这个项目（列表卡片展示用）"
            :rows="2"
            :maxlength="500"
            show-count
          />
        </a-form-item>
        <a-form-item label="详细描述">
          <a-textarea
            v-model:value="form.description"
            placeholder="项目背景、职责、成果亮点等（可选）"
            :rows="5"
          />
        </a-form-item>
        <a-form-item label="技能 / 工具标签">
          <a-select
            v-model:value="form.techStack"
            mode="tags"
            placeholder="输入后回车添加，如 Go、Vue、Lightroom、Premiere"
            :max-tag-count="8"
          />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="开始时间">
              <a-date-picker
                v-model:value="form.start_date"
                picker="month"
                value-format="YYYY-MM"
                placeholder="选择月份"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="结束时间">
              <a-date-picker
                v-model:value="form.end_date"
                picker="month"
                value-format="YYYY-MM"
                placeholder="留空表示至今"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="项目 / 作品链接">
          <a-input
            v-model:value="form.project_url"
            placeholder="https://（可选）"
            :maxlength="1024"
          />
        </a-form-item>
        <a-form-item label="代码仓库链接">
          <a-input
            v-model:value="form.repo_url"
            placeholder="https://github.com/...（可选）"
            :maxlength="1024"
          />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="状态">
              <a-radio-group v-model:value="form.status">
                <a-radio :value="0">草稿</a-radio>
                <a-radio :value="1">已发布</a-radio>
              </a-radio-group>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="排序权重">
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

    <MediaPicker v-model:visible="mediaPickerVisible" @selected="handleMediaSelected" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
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
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

/** 领域分类预设项，支持自由输入其他行业 */
const PRESET_CATEGORIES = ['摄影', '视频剪辑', '技术开发', '设计', '其他']

const presetCategoryOptions = PRESET_CATEGORIES.map((c) => ({ value: c }))

const statusOptions = [
  { value: 0, label: '草稿' },
  { value: 1, label: '已发布' },
]

const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const filterCategory = ref<string | undefined>(undefined)
const filterStatus = ref<number | undefined>(undefined)

const list = ref<Project[]>([])
const loading = ref(false)

const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(12)

/** 筛选下拉选项 = 预设分类 + 当前列表中实际出现的分类 */
const filterCategoryOptions = computed(() => {
  const existed = list.value
    .map((item) => item.category)
    .filter((c): c is string => !!c && !PRESET_CATEGORIES.includes(c))
  return Array.from(new Set([...PRESET_CATEGORIES, ...existed])).map((c) => ({ value: c, label: c }))
})

const modalVisible = ref(false)
const modalLoading = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref('')

const mediaPickerVisible = ref(false)

const form = reactive({
  title: '',
  category: '',
  role: '',
  client: '',
  cover_url: '',
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

watch([debouncedKeyword, filterCategory, filterStatus], () => {
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
      category: filterCategory.value || undefined,
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

function handlePageChange(newPage: number, newPageSize?: number) {
  changePagination(newPage, newPageSize)
  fetchList()
}

function resetForm() {
  form.title = ''
  form.category = ''
  form.role = ''
  form.client = ''
  form.cover_url = ''
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
  form.category = item.category || ''
  form.role = item.role || ''
  form.client = item.client || ''
  form.cover_url = item.cover_url || ''
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

function handleMediaSelected(media: { url: string }) {
  form.cover_url = media.url
}

function splitTechStack(techStack?: string): string[] {
  if (!techStack) return []
  return techStack.split(',').map((s) => s.trim()).filter(Boolean)
}

function formatRange(start?: string, end?: string): string {
  if (!start && !end) return '时间待定'
  const endText = end || '至今'
  return `${start || '?'} ~ ${endText}`
}

async function handleSubmit() {
  if (!form.title.trim()) {
    message.warning('请输入项目名称')
    return
  }
  if (form.start_date && form.end_date && form.start_date > form.end_date) {
    message.warning('开始时间不能晚于结束时间')
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await projectApi.create({
        title: form.title.trim(),
        category: form.category.trim() || undefined,
        role: form.role.trim() || undefined,
        client: form.client.trim() || undefined,
        cover_url: form.cover_url.trim() || undefined,
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
      message.success('创建成功')
    } else {
      await projectApi.update(editingId.value, {
        title: form.title.trim(),
        category: form.category.trim(),
        role: form.role.trim(),
        client: form.client.trim(),
        cover_url: form.cover_url.trim(),
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
      message.success('更新成功')
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
    title: '确认删除',
    content: `删除「${item.title}」后无法恢复，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await projectApi.remove(item.id)
        message.success('删除成功')
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
  border: 1px solid var(--border-color, #f0f0f0);
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
  background: var(--bg-secondary, #f5f5f5);
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
  color: var(--text-secondary, #bfbfbf);
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
  color: var(--text-primary, #262626);
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
  color: var(--text-secondary, #8c8c8c);
}

.project-summary {
  font-size: 13px;
  color: var(--text-secondary, #8c8c8c);
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
  color: var(--text-secondary, #bfbfbf);
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
  color: var(--text-secondary, #8c8c8c);
  transition: color 0.2s;
}

.meta-links a:hover {
  color: #1890ff;
}

.project-actions {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  border-top: 1px solid var(--border-color, #f0f0f0);
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
  border: 1px solid var(--border-color, #f0f0f0);
}
</style>
