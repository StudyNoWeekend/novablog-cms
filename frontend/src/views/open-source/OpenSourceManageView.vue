<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">开源作品</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> 添加仓库
      </a-button>
    </div>

    <a-card>
      <div class="opensource-toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索仓库名称/简介..."
          style="width: 220px"
          allow-clear
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
          description="还没有开源作品，点击右上角添加 GitHub 仓库"
          style="margin-top: 48px"
        />
        <div v-else class="opensource-grid">
          <div v-for="item in list" :key="item.id" class="opensource-card">
            <div class="card-head">
              <GithubOutlined class="repo-icon" />
              <div class="repo-name" :title="item.name">{{ item.name }}</div>
              <a-tag :color="item.status === 1 ? 'green' : 'default'" class="status-tag">
                {{ item.status === 1 ? '已发布' : '草稿' }}
              </a-tag>
            </div>
            <div class="repo-summary" :title="item.summary">
              {{ item.summary || '暂无简介，保存后自动回填仓库描述' }}
            </div>
            <div class="repo-tags">
              <a-tag v-if="item.language" class="lang-tag" color="blue">{{ item.language }}</a-tag>
              <a-tag v-for="topic in splitTopics(item.topics).slice(0, 3)" :key="topic" class="topic-tag">
                {{ topic }}
              </a-tag>
            </div>
            <div class="repo-meta">
              <span class="meta-item" title="Star 数（同步时快照）">
                <StarOutlined /> {{ formatStars(item.stars) }}
              </span>
              <span class="meta-links">
                <a-button type="link" size="small" @click="openReadme(item)">
                  <FileTextOutlined /> README
                </a-button>
                <a
                  v-if="item.homepage"
                  :href="item.homepage"
                  target="_blank"
                  rel="noopener noreferrer"
                  title="主页/演示地址"
                >
                  <LinkOutlined />
                </a>
                <a
                  :href="item.repo_url"
                  target="_blank"
                  rel="noopener noreferrer"
                  title="仓库链接"
                >
                  <ExportOutlined />
                </a>
              </span>
            </div>
            <div class="repo-actions">
              <a-button type="link" size="small" :loading="refreshingId === item.id" @click="handleRefresh(item)">
                <SyncOutlined /> 同步
              </a-button>
              <a-button type="link" size="small" @click="handleEdit(item)">
                <EditOutlined /> 编辑
              </a-button>
              <a-button type="link" size="small" danger @click="handleDelete(item)">
                <DeleteOutlined /> 删除
              </a-button>
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
          :show-total="(t: number) => `共 ${t} 个仓库`"
          @change="handlePageChange"
          @show-size-change="handlePageChange"
        />
      </div>
    </a-card>

    <!-- 添加/编辑仓库弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalMode === 'create' ? '添加仓库' : '编辑仓库'"
      :confirm-loading="modalLoading"
      width="640px"
      @ok="handleSubmit"
    >
      <a-alert
        v-if="modalMode === 'create'"
        class="fetch-tip"
        type="info"
        show-icon
        message="保存时将自动拉取仓库的 README 与元数据（描述/语言/Star 等），请确保服务器可访问 GitHub"
      />
      <a-form layout="vertical">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="仓库名称" required>
              <a-input
                v-model:value="form.name"
                placeholder="如 novablog"
                :maxlength="255"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="主语言">
              <a-auto-complete
                v-model:value="form.language"
                :options="presetLanguageOptions"
                placeholder="如 Go / Vue，保存时自动回填"
                allow-clear
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="仓库链接" required>
          <a-input
            v-model:value="form.repo_url"
            placeholder="https://github.com/owner/repo"
            :maxlength="1024"
          />
        </a-form-item>
        <a-form-item label="一句话介绍">
          <a-textarea
            v-model:value="form.summary"
            placeholder="留空则保存时自动回填仓库描述（列表卡片展示用）"
            :rows="2"
            :maxlength="500"
            show-count
          />
        </a-form-item>
        <a-form-item label="主题标签">
          <a-select
            v-model:value="form.topics"
            mode="tags"
            placeholder="输入后回车添加，保存时自动回填仓库 topics"
            :max-tag-count="8"
          />
        </a-form-item>
        <a-form-item label="主页 / 演示地址">
          <a-input
            v-model:value="form.homepage"
            placeholder="https://（可选）"
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

    <!-- README 预览抽屉 -->
    <a-drawer
      v-model:open="readmeVisible"
      :title="readmeWork?.name || 'README'"
      width="720"
      placement="right"
    >
      <template #extra>
        <a-button
          size="small"
          :loading="readmeLoading"
          @click="readmeWork && openReadme(readmeWork, true)"
        >
          <SyncOutlined /> 重新加载
        </a-button>
      </template>
      <a-spin :spinning="readmeLoading">
        <div v-if="readmeWork?.readme_updated_at" class="readme-meta">
          最近拉取：{{ formatDateTime(readmeWork.readme_updated_at) }}
        </div>
        <a-empty
          v-if="!readmeLoading && !readmeWork?.readme"
          description="尚未拉取到 README，试试「同步」按钮"
        />
        <MarkdownViewer v-else-if="readmeWork?.readme" :value="readmeWork.readme" />
      </a-spin>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  GithubOutlined,
  StarOutlined,
  FileTextOutlined,
  LinkOutlined,
  ExportOutlined,
  SyncOutlined,
} from '@ant-design/icons-vue'
import { openSourceApi } from '@/api/openSource'
import type { OpenSourceWork } from '@/types/openSource'
import MarkdownViewer from '@/components/editor/MarkdownViewer.vue'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const statusOptions = [
  { value: 0, label: '草稿' },
  { value: 1, label: '已发布' },
]

/** 常见语言预设项，支持自由输入；实际值保存时以远端仓库为准回填 */
const PRESET_LANGUAGES = ['Go', 'Vue', 'TypeScript', 'JavaScript', 'Python', 'Rust', 'Java', 'Kotlin', 'Swift', 'Shell']

const presetLanguageOptions = PRESET_LANGUAGES.map((l) => ({ value: l }))

const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const filterStatus = ref<number | undefined>(undefined)

const list = ref<OpenSourceWork[]>([])
const loading = ref(false)

const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(12)

const modalVisible = ref(false)
const modalLoading = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref('')

const refreshingId = ref('')

const readmeVisible = ref(false)
const readmeLoading = ref(false)
const readmeWork = ref<OpenSourceWork | null>(null)

const form = reactive({
  name: '',
  repo_url: '',
  summary: '',
  language: '',
  topics: [] as string[],
  homepage: '',
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
    const res = await openSourceApi.getList({
      page: page.value,
      page_size: pageSize.value,
      keyword: debouncedKeyword.value || undefined,
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
  form.name = ''
  form.repo_url = ''
  form.summary = ''
  form.language = ''
  form.topics = []
  form.homepage = ''
  form.status = 0
  form.sort_order = 0
}

function openCreateModal() {
  modalMode.value = 'create'
  editingId.value = ''
  resetForm()
  modalVisible.value = true
}

function handleEdit(item: OpenSourceWork) {
  modalMode.value = 'edit'
  editingId.value = item.id
  resetForm()
  form.name = item.name
  form.repo_url = item.repo_url
  form.summary = item.summary || ''
  form.language = item.language || ''
  form.topics = splitTopics(item.topics)
  form.homepage = item.homepage || ''
  form.status = item.status
  form.sort_order = item.sort_order
  modalVisible.value = true
}

function splitTopics(topics?: string): string[] {
  if (!topics) return []
  return topics.split(',').map((s) => s.trim()).filter(Boolean)
}

function formatStars(stars: number): string {
  if (stars >= 1000) {
    return `${(stars / 1000).toFixed(1)}k`
  }
  return String(stars)
}

function formatDateTime(value?: string | null): string {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}

async function handleSubmit() {
  if (!form.name.trim()) {
    message.warning('请输入仓库名称')
    return
  }
  if (!form.repo_url.trim()) {
    message.warning('请输入仓库链接')
    return
  }
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await openSourceApi.create({
        name: form.name.trim(),
        repo_url: form.repo_url.trim(),
        summary: form.summary.trim() || undefined,
        language: form.language.trim() || undefined,
        topics: form.topics.join(',') || undefined,
        homepage: form.homepage.trim() || undefined,
        status: form.status,
        sort_order: form.sort_order,
      })
      message.success('添加成功，README 已自动拉取')
    } else {
      await openSourceApi.update(editingId.value, {
        name: form.name.trim(),
        repo_url: form.repo_url.trim(),
        summary: form.summary.trim(),
        language: form.language.trim(),
        topics: form.topics.join(','),
        homepage: form.homepage.trim(),
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

/** 同步：重新拉取远端元数据与 README */
async function handleRefresh(item: OpenSourceWork) {
  refreshingId.value = item.id
  try {
    await openSourceApi.refresh(item.id)
    message.success(`「${item.name}」已同步`)
    fetchList()
  } catch {
    // 错误由拦截器处理
  } finally {
    refreshingId.value = ''
  }
}

/** 查看 README：详情接口含 readme 字段 */
async function openReadme(item: OpenSourceWork, reload = false) {
  readmeVisible.value = true
  readmeLoading.value = true
  if (!reload) {
    readmeWork.value = { ...item, readme: '' }
  }
  try {
    const detail = await openSourceApi.getById(item.id)
    readmeWork.value = detail
  } catch {
    // 错误由拦截器处理
  } finally {
    readmeLoading.value = false
  }
}

function handleDelete(item: OpenSourceWork) {
  Modal.confirm({
    title: '确认删除',
    content: `删除「${item.name}」后无法恢复，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await openSourceApi.remove(item.id)
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
.opensource-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.opensource-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.opensource-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #f0f0f0);
  border-radius: 8px;
  transition: box-shadow 0.2s;
}

.opensource-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.card-head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.repo-icon {
  font-size: 18px;
  color: #4a6cf7;
  flex-shrink: 0;
}

.repo-name {
  flex: 1;
  min-width: 0;
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

.repo-summary {
  font-size: 13px;
  color: var(--text-secondary, #595959);
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 42px;
}

.repo-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.lang-tag,
.topic-tag {
  margin: 0;
  font-size: 12px;
}

.repo-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: var(--text-secondary, #8c8c8c);
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

.repo-actions {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  border-top: 1px solid var(--border-color, #f0f0f0);
  padding-top: 4px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}

.fetch-tip {
  margin-bottom: 16px;
}

.readme-meta {
  margin-bottom: 12px;
  font-size: 12px;
  color: var(--text-secondary, #8c8c8c);
}
</style>
