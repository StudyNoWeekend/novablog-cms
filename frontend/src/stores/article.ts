import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { articleApi } from '@/api/article'
import type { Article, ArticleFilters, ArticleCreateReq, ArticleUpdateReq } from '@/types/article'
import { message } from 'ant-design-vue'

export const useArticleStore = defineStore('article', () => {
  const list = ref<Article[]>([])
  const total = ref(0)
  const loading = ref(false)
  const currentArticle = ref<Article | null>(null)
  const filters = reactive<ArticleFilters>({
    status: undefined,
    category_id: undefined,
    keyword: undefined,
  })
  const pagination = reactive({
    page: 1,
    pageSize: 20,
  })

  async function fetchList() {
    loading.value = true
    try {
      const res = await articleApi.getList({
        page: pagination.page,
        page_size: pagination.pageSize,
        ...filters,
      })
      list.value = res?.list || []
      total.value = res?.total || 0
    } catch {
      // 错误由 request.ts 拦截器统一处理
      list.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  async function fetchDetail(id: string) {
    loading.value = true
    try {
      currentArticle.value = await articleApi.getDetail(id)
    } finally {
      loading.value = false
    }
  }

  async function create(data: ArticleCreateReq) {
    const res = await articleApi.create(data)
    return res
  }

  async function update(id: string, data: ArticleUpdateReq) {
    const res = await articleApi.update(id, data)
    return res
  }

  async function remove(id: string) {
    await articleApi.remove(id)
    message.success('文章已删除')
    await fetchList()
  }

  async function updateStatus(id: string, status: number) {
    await articleApi.updateStatus(id, status)
    const statusText = { 1: '转为草稿', 2: '已发布', 3: '已下架' }
    message.success(statusText[status as keyof typeof statusText] || '状态已更新')
    await fetchList()
  }

  function setFilters(newFilters: Partial<ArticleFilters>) {
    Object.assign(filters, newFilters)
    pagination.page = 1
  }

  function setPage(page: number) {
    pagination.page = page
  }

  return {
    list, total, loading, currentArticle, filters, pagination,
    fetchList, fetchDetail, create, update, remove, updateStatus,
    setFilters, setPage,
  }
})