import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { message } from 'ant-design-vue'

import { themeMarketApi, resolveMarketRetries, abortMarketRetries } from '@/api/theme'
import { configApi } from '@/api/config'
import { marketStorage } from '@/utils/storage'
import type {
  HotTag,
  MarketUser,
  ThemeDetail,
  ThemeItem,
  ThemeMarketFilters,
  ThemeReleaseItem,
  ThemeStats,
} from '@/types/theme'

// 官方服务地址默认值由后端下发（GET /public/config），前端不硬编码；
// defaultBaseURL 保存后端当前生效值，marketBaseURL 为本浏览器实际使用的地址（登录弹窗可改）
export const defaultBaseURL = ref('')

export type MarketTabKey = 'market' | 'favorites'

export const useThemeMarketStore = defineStore('themeMarket', () => {
  // ===== 官方账号登录态 =====
  const loggedIn = ref(!!(marketStorage.getBaseURL() && marketStorage.getToken()))
  const marketUser = ref<MarketUser | null>(marketStorage.getUser())
  const marketBaseURL = ref(marketStorage.getBaseURL() || '')
  const activeTab = ref<MarketTabKey>('market')

  // ===== 市场数据 =====
  const list = ref<ThemeItem[]>([])
  const total = ref(0)
  const loading = ref(false)
  const loadError = ref(false)
  const stats = ref<ThemeStats | null>(null)
  const hotTags = ref<HotTag[]>([])
  const favoriteIds = ref<Set<number>>(new Set())

  const filters = reactive<ThemeMarketFilters>({
    type: undefined,
    styles: [],
    price: undefined,
    search: '',
    sort: 'latest',
  })
  const pagination = reactive({ page: 1, pageSize: 9 })

  // ===== 我的收藏 =====
  const favList = ref<ThemeItem[]>([])
  const favTotal = ref(0)
  const favLoading = ref(false)
  const favLoadError = ref(false)
  const favPagination = reactive({ page: 1, pageSize: 9 })

  // ===== 主题详情页 =====
  const detailLoading = ref(false)
  const detailError = ref(false)
  const currentDetail = ref<ThemeDetail | null>(null)
  const releases = ref<ThemeReleaseItem[]>([])
  const releasesLoading = ref(false)
  const releasesError = ref(false)

  function buildListParams() {
    return {
      page: pagination.page,
      page_size: pagination.pageSize,
      type: filters.type,
      styles: filters.styles?.length ? filters.styles : undefined,
      price: filters.price,
      search: filters.search || undefined,
      sort: filters.sort,
    }
  }

  async function fetchList() {
    loading.value = true
    loadError.value = false
    try {
      const res = await themeMarketApi.getList(buildListParams())
      list.value = res?.list || []
      total.value = res?.total || 0
    } catch {
      loadError.value = true
      list.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  /** 筛选变更后回到第一页并重新拉取 */
  async function fetchListWithReset() {
    pagination.page = 1
    await fetchList()
  }

  async function fetchStats() {
    try {
      stats.value = await themeMarketApi.getStats()
    } catch {
      // 错误由 request.ts 拦截器统一提示
    }
  }

  async function fetchHotTags() {
    try {
      hotTags.value = await themeMarketApi.getHotTags()
    } catch {
      // 错误由 request.ts 拦截器统一提示
    }
  }

  async function fetchFavoriteIds() {
    try {
      favoriteIds.value = new Set(await themeMarketApi.getFavoriteIds())
    } catch {
      // 错误由 request.ts 拦截器统一提示
    }
  }

  async function fetchFavorites() {
    favLoading.value = true
    favLoadError.value = false
    try {
      const res = await themeMarketApi.getFavorites({
        page: favPagination.page,
        page_size: favPagination.pageSize,
      })
      favList.value = res?.list || []
      favTotal.value = res?.total || 0
    } catch {
      favLoadError.value = true
      favList.value = []
      favTotal.value = 0
    } finally {
      favLoading.value = false
    }
  }

  /** 拉取后端下发的官方地址默认值；本地无使用值时以之为准 */
  async function fetchDefaultBaseURL() {
    try {
      const config = await configApi.getPublicConfig()
      defaultBaseURL.value = config.market_base_url || ''
      if (!marketBaseURL.value && defaultBaseURL.value) {
        marketBaseURL.value = defaultBaseURL.value
      }
    } catch {
      // 下发失败时保持为空，登录弹窗允许手动输入
    }
  }

  /** 登录成功后拉取市场全量数据 */
  async function initMarketData() {
    pagination.page = 1
    favPagination.page = 1
    await Promise.all([fetchList(), fetchStats(), fetchHotTags(), fetchFavoriteIds()])
  }

  // ===== 官方账号登录态操作 =====

  async function login(baseURL: string, email: string, password: string) {
    const res = await themeMarketApi.login(baseURL, { email, password })
    marketStorage.setBaseURL(baseURL)
    marketStorage.setToken(res.access_token)
    marketStorage.setUser(res.user)
    marketBaseURL.value = baseURL
    // 后端已在登录成功时持久化该地址为全局默认
    defaultBaseURL.value = baseURL
    marketUser.value = res.user
    loggedIn.value = true
    // 用新 Token 重放登录失效期间挂起的请求（安装/更新/下载等自动继续）
    resolveMarketRetries()
    await initMarketData()
  }

  /** 官方登录失效（api 层已清空凭据并广播事件） */
  function handleAuthExpired() {
    marketUser.value = null
    loggedIn.value = false
  }

  async function logout() {
    // 挂起中的请求不再等待重放，直接拒绝
    abortMarketRetries()
    try {
      await themeMarketApi.logout()
    } catch {
      // 官方侧已失效也继续本地退出
    }
    marketStorage.clear()
    marketUser.value = null
    loggedIn.value = false
    activeTab.value = 'market'
    // 退出后回到后端下发的默认地址
    marketBaseURL.value = defaultBaseURL.value
  }

  // ===== 浏览与互动 =====

  async function toggleFavorite(id: number) {
    const res = await themeMarketApi.toggleFavorite(id)
    const next = new Set(favoriteIds.value)
    if (res.favorited) {
      next.add(id)
    } else {
      next.delete(id)
      // 收藏页处于当前取消的主题页时同步移除
      if (activeTab.value === 'favorites') {
        favList.value = favList.value.filter((t) => t.id !== id)
        favTotal.value = Math.max(0, favTotal.value - 1)
      }
    }
    favoriteIds.value = next
    message.success(res.favorited ? '已收藏' : '已取消收藏')
    return res.favorited
  }

  async function toggleLike(id: number) {
    const res = await themeMarketApi.toggleLike(id)
    const delta = res.liked ? 1 : -1
    if (currentDetail.value?.id === id) {
      currentDetail.value.liked = res.liked
      currentDetail.value.likes += delta
    }
    const item = list.value.find((t) => t.id === id)
    if (item) item.likes += delta
    return res.liked
  }

  async function rate(id: number, score: number) {
    const res = await themeMarketApi.rate(id, score)
    if (currentDetail.value?.id === id) {
      currentDetail.value.user_rating = score
      currentDetail.value.rating = res.rating
    }
    const item = list.value.find((t) => t.id === id)
    if (item) item.rating = res.rating
    const fav = favList.value.find((t) => t.id === id)
    if (fav) fav.rating = res.rating
    message.success('评分成功')
    return res.rating
  }

  async function download(id: number) {
    const res = await themeMarketApi.download(id)
    if (res.download_url) {
      // 异步回调中 window.open 可能被拦截，被拦时降级为提示
      const win = window.open(res.download_url, '_blank')
      if (!win) message.info(`下载地址：${res.download_url}`)
    } else {
      message.warning('该主题未配置下载地址')
    }
    if (currentDetail.value?.id === id) currentDetail.value.downloads += 1
    const item = list.value.find((t) => t.id === id)
    if (item) item.downloads += 1
    const fav = favList.value.find((t) => t.id === id)
    if (fav) fav.downloads += 1
  }

  async function fetchDetail(id: number) {
    detailLoading.value = true
    detailError.value = false
    currentDetail.value = null
    try {
      currentDetail.value = await themeMarketApi.getDetail(id)
    } catch {
      detailError.value = true
    } finally {
      detailLoading.value = false
    }
  }

  /** 拉取主题版本历史与更新日志（按发布时间倒序） */
  async function fetchReleases(id: number) {
    releasesLoading.value = true
    releasesError.value = false
    releases.value = []
    try {
      releases.value = await themeMarketApi.getReleases(id)
    } catch {
      releasesError.value = true
    } finally {
      releasesLoading.value = false
    }
  }

  return {
    loggedIn, marketUser, marketBaseURL, defaultBaseURL, activeTab,
    list, total, loading, loadError, stats, hotTags, favoriteIds, filters, pagination,
    favList, favTotal, favLoading, favLoadError, favPagination,
    detailLoading, detailError, currentDetail, releases, releasesLoading, releasesError,
    fetchList, fetchListWithReset, fetchStats, fetchHotTags, fetchFavoriteIds, fetchFavorites,
    initMarketData, login, handleAuthExpired, logout, fetchDefaultBaseURL,
    toggleFavorite, toggleLike, rate, download,
    fetchDetail, fetchReleases,
  }
})
