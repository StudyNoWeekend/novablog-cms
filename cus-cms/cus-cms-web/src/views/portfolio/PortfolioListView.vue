<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">摄影作品集</h1>
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined /> 新建作品集
      </a-button>
    </div>

    <a-card>
      <div class="portfolio-toolbar">
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索作品集名称..."
          style="width: 240px"
          allow-clear
        />
      </div>

      <a-spin :spinning="loading">
        <a-empty
          v-if="!loading && list.length === 0"
          description="还没有作品集，开始创建吧"
          style="margin-top: 48px"
        />
        <div v-else class="portfolio-grid">
          <div v-for="item in list" :key="item.id" class="portfolio-card">
            <div class="portfolio-cover" @click="handleEdit(item.id)">
              <img
                v-if="item.cover_url"
                :src="item.cover_url"
                :alt="item.name"
              />
              <div v-else class="cover-placeholder">
                <PictureOutlined />
              </div>
              <div class="cover-overlay">
                <span>编辑</span>
              </div>
            </div>
            <div class="portfolio-info">
              <div class="portfolio-name" :title="item.name">{{ item.name }}</div>
              <div class="portfolio-desc" :title="item.description">
                {{ item.description || '暂无介绍' }}
              </div>
              <div class="portfolio-meta">
                <span class="meta-item">
                  <PictureOutlined /> {{ item.item_count }} 件作品
                </span>
                <span class="meta-item">{{ formatDateTime(item.created_at) }}</span>
              </div>
              <div class="portfolio-actions">
                <a-button type="link" size="small" @click="handleEdit(item.id)">
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
          :show-total="(t: number) => `共 ${t} 个作品集`"
          @change="handlePageChange"
          @show-size-change="handlePageChange"
        />
      </div>
    </a-card>

    <!-- 新建作品集弹窗 -->
    <a-modal
      v-model:open="createVisible"
      title="新建作品集"
      :confirm-loading="creating"
      @ok="handleCreate"
    >
      <a-form layout="vertical">
        <a-form-item label="名称" required>
          <a-input
            v-model:value="createForm.name"
            placeholder="请输入作品集名称"
            :maxlength="100"
            show-count
          />
        </a-form-item>
        <a-form-item label="介绍">
          <a-textarea
            v-model:value="createForm.description"
            placeholder="请输入作品集介绍（可选）"
            :rows="4"
            :maxlength="500"
            show-count
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PictureOutlined,
} from '@ant-design/icons-vue'
import { portfolioApi } from '@/api/portfolio'
import type { Portfolio, CreatePortfolioReq } from '@/types/portfolio'
import { usePagination } from '@/composables/usePagination'
import { useDebounce } from '@/composables/useDebounce'

const router = useRouter()

const keyword = ref('')
const { debouncedValue: debouncedKeyword, setDebounce } = useDebounce(keyword)

const list = ref<Portfolio[]>([])
const loading = ref(false)

const { page, pageSize, total, reset, handlePageChange: changePagination } = usePagination(12)

const createVisible = ref(false)
const creating = ref(false)
const createForm = ref<CreatePortfolioReq>({ name: '', description: '' })

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
  createForm.value = { name: '', description: '' }
  createVisible.value = true
}

async function handleCreate() {
  if (!createForm.value.name.trim()) {
    message.warning('请输入作品集名称')
    return
  }
  creating.value = true
  try {
    const res = await portfolioApi.create({
      name: createForm.value.name.trim(),
      description: createForm.value.description?.trim() || undefined,
    })
    message.success('创建成功')
    createVisible.value = false
    router.push(`/portfolios/${res.id}/edit`)
  } catch {
    // 错误由拦截器处理
  } finally {
    creating.value = false
  }
}

function handleEdit(id: string) {
  router.push(`/portfolios/${id}/edit`)
}

function handleDelete(item: Portfolio) {
  Modal.confirm({
    title: '确认删除',
    content: `删除「${item.name}」后无法恢复，作品项也将一并删除，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await portfolioApi.remove(item.id)
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

function formatDateTime(time?: string): string {
  if (!time) return '-'
  const d = new Date(time)
  if (Number.isNaN(d.getTime())) return time
  return d.toLocaleString('zh-CN')
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
  border: 1px solid var(--border-color, #f0f0f0);
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
  background: var(--bg-secondary, #f5f5f5);
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
  color: var(--text-secondary, #bfbfbf);
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
  color: var(--text-primary, #262626);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.portfolio-desc {
  font-size: 13px;
  color: var(--text-secondary, #8c8c8c);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 39px;
}

.portfolio-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-secondary, #bfbfbf);
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
  border-top: 1px solid var(--border-color, #f0f0f0);
  margin-top: 4px;
  padding-top: 4px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}
</style>
