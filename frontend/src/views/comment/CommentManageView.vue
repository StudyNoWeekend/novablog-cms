<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('comment.title') }}</h1>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <a-select
        v-model:value="filterTargetType"
        :placeholder="$t('comment.filter.allSources')"
        allow-clear
        style="width: 150px"
        @change="handleFilterChange"
      >
        <a-select-option :value="undefined">{{ $t('comment.filter.allSources') }}</a-select-option>
        <a-select-option value="article">{{ $t('comment.source.article') }}</a-select-option>
        <a-select-option value="travel">{{ $t('comment.source.travel') }}</a-select-option>
      </a-select>

      <a-input-search
        v-model:value="filterKeyword"
        :placeholder="$t('comment.filter.searchPlaceholder')"
        style="width: 240px"
        @search="handleFilterChange"
        :loading="loading"
      />
    </div>

    <!-- 错误状态 -->
    <a-result
      v-if="error"
      status="error"
      :title="$t('common.state.error')"
      :sub-title="$t('comment.error.getListFailed')"
    >
      <template #extra>
        <a-button type="primary" @click="retry">{{ $t('common.actions.retry') }}</a-button>
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
          <a-empty :description="$t('comment.empty')" />
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
                  <a-tag v-if="record.is_blogger" color="blue" class="blogger-tag">{{ $t('comment.blogger') }}</a-tag>
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
            <span class="time-text">{{ formatDateTime(record.created_at) }}</span>
          </template>

          <!-- 操作 -->
          <template v-else-if="column.key === 'action'">
            <div class="action-cell">
              <a-button type="link" size="small" @click="openReply(record)">
                <MessageOutlined /> {{ $t('comment.action.reply') }}
              </a-button>
              <a-button type="link" size="small" danger @click="handleDelete(record.id)">
                <DeleteOutlined /> {{ $t('common.actions.delete') }}
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
        :show-total="(total: number) => t('common.table.total', { total })"
        @change="handlePageChange"
        @show-size-change="handlePageSizeChange"
      />
    </div>

    <!-- 回复 Modal -->
    <a-modal
      v-model:open="replyModalVisible"
      :title="$t('comment.reply.title')"
      :confirm-loading="replySubmitting"
      :ok-text="$t('comment.reply.okText')"
      :cancel-text="$t('common.actions.cancel')"
      @ok="submitReply"
    >
      <div class="reply-context" v-if="replyTarget">
        <span class="reply-context-label">{{ $t('comment.reply.replyTo', { name: replyTarget.nickname }) }}</span>
        <span class="reply-context-content">{{ replyTarget.content }}</span>
      </div>
      <a-textarea
        v-model:value="replyContent"
        :rows="4"
        :placeholder="$t('comment.reply.placeholder')"
        :maxlength="500"
        show-count
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Modal, message, type TableColumnsType } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import {
  MessageOutlined,
  DeleteOutlined,
  FileTextOutlined,
  CompassOutlined,
} from '@ant-design/icons-vue'
import { commentApi } from '@/api/comment'
import type { Comment, CommentListReq } from '@/types/comment'
import { formatDateTime } from '@/utils/datetime'

const { t } = useI18n()

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

// 表格列定义（包裹 computed，切换语言时刷新）
const columns = computed<TableColumnsType<Comment>>(() => [
  { title: t('comment.columns.commenter'), key: 'commenter', width: 220 },
  { title: t('common.field.content'), key: 'content', ellipsis: true },
  { title: t('comment.columns.source'), key: 'target', width: 200 },
  { title: t('comment.columns.time'), key: 'created_at', width: 170 },
  { title: t('common.table.actions'), key: 'action', width: 200 },
])

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
    message.warning(t('comment.message.replyEmpty'))
    return
  }
  replySubmitting.value = true
  try {
    await commentApi.reply(replyTarget.value.id, { content })
    message.success(t('comment.message.replySuccess'))
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
    title: t('comment.message.deleteTitle'),
    content: t('comment.message.deleteContent'),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await commentApi.remove(id)
        message.success(t('common.message.deleteSuccess'))
        fetchList()
      } catch {
        // 错误由请求拦截器处理
      }
    },
  })
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
