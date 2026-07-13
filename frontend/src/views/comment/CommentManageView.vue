<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">评论管理</h1>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <a-select
        v-model:value="filterTargetType"
        placeholder="全部来源"
        allow-clear
        style="width: 150px"
        @change="handleFilterChange"
      >
        <a-select-option :value="undefined">全部来源</a-select-option>
        <a-select-option value="article">文章</a-select-option>
        <a-select-option value="travel">旅行攻略</a-select-option>
      </a-select>

      <a-input-search
        v-model:value="filterKeyword"
        placeholder="搜索评论内容..."
        style="width: 240px"
        @search="handleFilterChange"
        :loading="loading"
      />
    </div>

    <!-- 错误状态 -->
    <a-result
      v-if="error"
      status="error"
      title="加载失败"
      sub-title="获取评论列表时出错，请重试"
    >
      <template #extra>
        <a-button type="primary" @click="retry">重试</a-button>
      </template>
    </a-result>

    <!-- 评论表格 -->
    <div v-else class="table-wrapper">
      <a-table
        :columns="columns"
        :data-source="list"
        :loading="loading"
        :pagination="false"
        row-key="id"
        :scroll="{ x: 900 }"
      >
        <template #emptyText>
          <a-empty description="暂无评论" />
        </template>

        <template #bodyCell="{ column, record }">
          <!-- 评论者 -->
          <template v-if="column.key === 'commenter'">
            <div class="commenter-cell">
              <a-avatar :size="32">
                {{ record.nickname?.charAt(0) || '?' }}
              </a-avatar>
              <div class="commenter-info">
                <div class="commenter-name-row">
                  <span class="commenter-name">{{ record.nickname }}</span>
                  <a-tag v-if="record.is_blogger" color="blue" class="blogger-tag">博主</a-tag>
                </div>
                <a
                  v-if="record.website"
                  :href="record.website"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="commenter-website"
                >
                  {{ record.website }}
                </a>
              </div>
            </div>
          </template>

          <!-- 内容 -->
          <template v-else-if="column.key === 'content'">
            <div class="comment-content" :title="record.content">{{ record.content }}</div>
          </template>

          <!-- 来源目标 -->
          <template v-else-if="column.key === 'target'">
            <div class="target-cell">
              <FileTextOutlined v-if="record.target_type === 'article'" class="target-icon" />
              <CompassOutlined v-else class="target-icon" />
              <span class="target-title">{{ record.target_title }}</span>
            </div>
          </template>

          <!-- 时间 -->
          <template v-else-if="column.key === 'created_at'">
            <span class="time-text">{{ formatDate(record.created_at) }}</span>
          </template>

          <!-- 操作 -->
          <template v-else-if="column.key === 'action'">
            <div class="action-cell">
              <a-button type="link" size="small" @click="openReply(record)">
                <MessageOutlined /> 回复
              </a-button>
              <a-button type="link" size="small" danger @click="handleDelete(record.id)">
                <DeleteOutlined /> 删除
              </a-button>
            </div>
          </template>
        </template>
      </a-table>
    </div>

    <!-- 分页 -->
    <div v-if="total > pageSize" class="pagination-wrapper">
      <a-pagination
        :current="page"
        :page-size="pageSize"
        :total="total"
        :show-size-changer="true"
        :show-total="(total: number) => `共 ${total} 条`"
        @change="handlePageChange"
        @show-size-change="handlePageSizeChange"
      />
    </div>

    <!-- 回复 Modal -->
    <a-modal
      v-model:open="replyModalVisible"
      title="回复评论"
      :confirm-loading="replySubmitting"
      ok-text="发送回复"
      cancel-text="取消"
      @ok="submitReply"
    >
      <div class="reply-context" v-if="replyTarget">
        <span class="reply-context-label">回复 {{ replyTarget.nickname }}：</span>
        <span class="reply-context-content">{{ replyTarget.content }}</span>
      </div>
      <a-textarea
        v-model:value="replyContent"
        :rows="4"
        placeholder="请输入回复内容..."
        :maxlength="500"
        show-count
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Modal, message, type TableColumnsType } from 'ant-design-vue'
import {
  MessageOutlined,
  DeleteOutlined,
  FileTextOutlined,
  CompassOutlined,
} from '@ant-design/icons-vue'
import { commentApi } from '@/api/comment'
import type { Comment, CommentListReq } from '@/types/comment'

// 列表状态
const loading = ref(false)
const list = ref<Comment[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filterTargetType = ref<string | undefined>(undefined)
const filterKeyword = ref('')
const error = ref(false)

// 回复 Modal
const replyModalVisible = ref(false)
const replyContent = ref('')
const replyTarget = ref<Comment | null>(null)
const replySubmitting = ref(false)

// 表格列定义
const columns: TableColumnsType<Comment> = [
  { title: '评论者', key: 'commenter', width: 220 },
  { title: '内容', key: 'content', ellipsis: true },
  { title: '来源', key: 'target', width: 200 },
  { title: '时间', key: 'created_at', width: 170 },
  { title: '操作', key: 'action', width: 200 },
]

onMounted(() => {
  fetchList()
})

async function fetchList() {
  loading.value = true
  error.value = false
  try {
    const params: CommentListReq = {
      page: page.value,
      page_size: pageSize.value,
      target_type: filterTargetType.value,
      keyword: filterKeyword.value || undefined,
    }
    const data = await commentApi.getList(params)
    list.value = data.list
    total.value = data.total
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

function retry() {
  fetchList()
}

function handleFilterChange() {
  page.value = 1
  fetchList()
}

function handlePageChange(p: number, ps: number) {
  page.value = p
  pageSize.value = ps
  fetchList()
}

function handlePageSizeChange(_current: number, size: number) {
  page.value = 1
  pageSize.value = size
  fetchList()
}

function openReply(comment: Comment) {
  replyTarget.value = comment
  replyContent.value = ''
  replyModalVisible.value = true
}

async function submitReply() {
  if (!replyTarget.value) return
  const content = replyContent.value.trim()
  if (!content) {
    message.warning('请输入回复内容')
    return
  }
  replySubmitting.value = true
  try {
    await commentApi.reply(replyTarget.value.id, { content })
    message.success('回复成功')
    replyModalVisible.value = false
    fetchList()
  } catch {
    // 错误由请求拦截器处理
  } finally {
    replySubmitting.value = false
  }
}

function handleDelete(id: string) {
  Modal.confirm({
    title: '确认删除',
    content: '删除后评论将无法恢复，确定要删除吗？',
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await commentApi.remove(id)
        message.success('删除成功')
        fetchList()
      } catch {
        // 错误由请求拦截器处理
      }
    },
  })
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
</script>

<style scoped>
.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.table-wrapper {
  background: var(--bg-card);
  border-radius: var(--border-radius-lg);
  padding: 16px;
  box-shadow: var(--shadow-card);
}

.commenter-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.commenter-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.commenter-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.commenter-name {
  font-weight: 500;
  color: var(--text-primary);
}

.commenter-website {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 140px;
}

.commenter-website:hover {
  color: var(--color-primary);
}

.blogger-tag {
  margin-left: 0;
}

.comment-content {
  color: var(--text-primary);
  word-break: break-word;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.target-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-secondary);
}

.target-icon {
  color: var(--color-primary);
  flex-shrink: 0;
}

.target-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time-text {
  color: var(--text-secondary);
  font-size: 13px;
}

.action-cell {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  align-items: center;
}

.action-cell :deep(.ant-btn-link) {
  padding: 0 4px;
  height: auto;
  cursor: pointer;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 32px;
}

.reply-context {
  background: var(--color-primary-light);
  border-radius: var(--border-radius);
  padding: 10px 12px;
  margin-bottom: 16px;
  font-size: 13px;
  line-height: 1.6;
}

.reply-context-label {
  color: var(--color-primary);
  font-weight: 500;
}

.reply-context-content {
  color: var(--text-secondary);
  display: block;
  margin-top: 4px;
}

:deep(.ant-table) {
  border-radius: var(--border-radius-lg);
}

:deep(.ant-btn) {
  cursor: pointer;
}

:deep(.ant-select),
:deep(.ant-input) {
  cursor: pointer;
}

@media (max-width: 768px) {
  .filter-bar {
    flex-direction: column;
  }

  .filter-bar .ant-select,
  .filter-bar .ant-input-search {
    width: 100% !important;
  }
}
</style>
