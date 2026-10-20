<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('article.list.title') }}</h1>
      <a-button type="primary" @click="router.push('/articles/create')">
        <PlusOutlined /> {{ $t('article.list.writeArticle') }}
      </a-button>
    </div>

    <!-- 分类标签栏 -->
    <CategoryBar type="article" @select="handleCategorySelect" />

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <a-select
        v-model:value="filterStatus"
        :placeholder="$t('article.filter.allStatus')"
        allow-clear
        style="width: 140px"
        @change="handleFilterChange"
      >
        <a-select-option :value="undefined">{{ $t('article.filter.allStatus') }}</a-select-option>
        <a-select-option :value="1">{{ $t('article.status.draft') }}</a-select-option>
        <a-select-option :value="2">{{ $t('article.status.published') }}</a-select-option>
        <a-select-option :value="3">{{ $t('article.status.unpublished') }}</a-select-option>
      </a-select>

      <a-input-search
        v-model:value="filterKeyword"
        :placeholder="$t('article.list.searchPlaceholder')"
        style="width: 240px"
        @search="handleFilterChange"
        :loading="store.loading"
      />
    </div>

    <!-- 加载骨架屏 -->
    <div v-if="store.loading && (!store.list || store.list.length === 0)" class="article-grid">
      <div v-for="i in 8" :key="i" class="skeleton-card">
        <a-skeleton active :paragraph="{ rows: 3 }" />
      </div>
    </div>

    <!-- 错误状态 -->
    <a-result
      v-else-if="error"
      status="error"
      :title="$t('common.state.error')"
      :sub-title="$t('article.list.loadError')"
    >
      <template #extra>
        <a-button type="primary" @click="retry">{{ $t('common.actions.retry') }}</a-button>
      </template>
    </a-result>

    <!-- 空状态 -->
    <a-empty
      v-else-if="!store.loading && (!store.list || store.list.length === 0)"
      :description="$t('article.list.empty')"
    >
      <a-button type="primary" @click="router.push('/articles/create')">
        {{ $t('article.list.writeFirst') }}
      </a-button>
    </a-empty>

    <!-- 文章卡片网格 -->
    <div v-else-if="store.list && store.list.length > 0" class="article-grid">
      <ArticleCard
        v-for="article in store.list"
        :key="article.id"
        :article="article"
        @edit="handleEdit"
        @delete="handleDelete"
        @pushUpdated="fetchList"
        @status-change="handleStatusChange"
      />
    </div>

    <!-- 分页 -->
    <div v-if="store.total > store.pagination.pageSize" class="pagination-wrapper">
      <a-pagination
        :current="store.pagination.page"
        :page-size="store.pagination.pageSize"
        :total="store.total"
        :show-size-changer="true"
        :show-total="(total: number) => $t('article.list.totalArticles', { total })"
        @change="handlePageChange"
        @show-size-change="handlePageSizeChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { PlusOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { useArticleStore } from '@/stores/article'
import ArticleCard from '@/components/article/ArticleCard.vue'
import CategoryBar from '@/components/common/CategoryBar.vue'

const { t } = useI18n()
const router = useRouter()
const store = useArticleStore()

const filterStatus = ref<number | undefined>(undefined)
const filterKeyword = ref('')
const filterCategory = ref<string | undefined>(undefined)
const error = ref(false)

onMounted(() => {
  fetchList()
})

async function fetchList() {
  error.value = false
  try {
    await store.fetchList()
  } catch {
    error.value = true
  }
}

async function retry() {
  error.value = false
  await fetchList()
}

function handleFilterChange() {
  store.setFilters({
    status: filterStatus.value,
    category_id: filterCategory.value,
    keyword: filterKeyword.value || undefined,
  })
  fetchList()
}

function handleCategorySelect(id: string | undefined) {
  filterCategory.value = id
  handleFilterChange()
}

function handlePageChange(page: number) {
  store.setPage(page)
  fetchList()
}

function handlePageSizeChange(_page: number, size: number) {
  store.pagination.pageSize = size
  store.pagination.page = 1
  fetchList()
}

function handleEdit(id: string) {
  router.push(`/articles/${id}/edit`)
}

function handleDelete(id: string) {
  Modal.confirm({
    title: t('article.list.deleteTitle'),
    content: t('article.list.deleteConfirm'),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: () => store.remove(id),
  })
}

function handleStatusChange(id: string, status: number) {
  store.updateStatus(id, status).then(() => fetchList())
}
</script>

<style scoped>
.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.article-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

@media (max-width: 1400px) {
  .article-grid { grid-template-columns: repeat(3, 1fr); }
}

@media (max-width: 1100px) {
  .article-grid { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 768px) {
  .article-grid { grid-template-columns: 1fr; }
}

.skeleton-card {
  background: var(--bg-card, #fff);
  border-radius: var(--border-radius-lg, 12px);
  padding: 16px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 32px;
}
</style>