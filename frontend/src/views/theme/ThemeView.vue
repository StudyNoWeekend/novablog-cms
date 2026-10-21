<template>
  <div class="page-container">
    <!-- 官方账号登录弹窗：未登录时弹出，可主动关闭（关闭后页面展示登录引导占位） -->
    <a-modal
      :open="!store.loggedIn && !loginModalDismissed"
      :mask-closable="false"
      :keyboard="false"
      :footer="null"
      centered
      width="400px"
      @cancel="handleLoginModalClose"
    >
      <div class="market-login">
        <div class="market-login__icon"><SkinOutlined /></div>
        <h2 class="market-login__title">{{ $t('theme.login.title') }}</h2>
        <p class="market-login__desc">{{ $t('theme.login.desc') }}</p>
        <div class="market-login__official">
          <GlobalOutlined />
          <span class="market-login__official-url" :title="store.defaultBaseURL || ''">
            {{ store.defaultBaseURL === null ? $t('theme.login.officialLoading') : store.defaultBaseURL || $t('theme.login.officialMissing') }}
          </span>
          <a @click="goOfficialConfig">{{ $t('theme.login.goModify') }}</a>
        </div>
        <p class="market-login__official-hint">{{ $t('theme.login.officialHint') }}</p>
        <a-form :model="loginForm" layout="vertical" autocomplete="off" @finish="handleLogin">
          <a-form-item :label="$t('theme.login.email')" name="email" :rules="emailRules">
            <a-input v-model:value="loginForm.email" :placeholder="$t('theme.login.emailPlaceholder')" allow-clear>
              <template #prefix><UserOutlined /></template>
            </a-input>
          </a-form-item>
          <a-form-item :label="$t('theme.login.password')" name="password" :rules="passwordRules">
            <a-input-password v-model:value="loginForm.password" :placeholder="$t('theme.login.passwordPlaceholder')">
              <template #prefix><LockOutlined /></template>
            </a-input-password>
          </a-form-item>
          <a-button type="primary" html-type="submit" block :loading="loginLoading">{{ $t('theme.login.submit') }}</a-button>
        </a-form>
        <div class="market-login__actions">
          <a-button block @click="handleGoRegister">{{ $t('theme.login.goRegister') }}</a-button>
          <a-button block @click="handleLoginModalClose">{{ $t('common.actions.close') }}</a-button>
        </div>
      </div>
    </a-modal>

    <!-- 博客 API 地址设置弹窗（官方站点地址统一在「系统设置 → 社区推送配置」维护） -->
    <a-modal
      v-model:open="marketConfigOpen"
      :title="$t('theme.apiModal.title')"
      :confirm-loading="marketConfigSaving"
      :ok-text="$t('common.actions.save')"
      :cancel-text="$t('common.actions.cancel')"
      @ok="handleSaveMarketConfig"
    >
      <a-form layout="vertical" autocomplete="off">
        <a-form-item :label="$t('theme.apiModal.label')" :validate-status="apiURLError ? 'error' : ''" :help="apiURLError">
          <a-input v-model:value="marketAPIURL" :placeholder="$t('theme.apiModal.placeholder')" allow-clear>
            <template #prefix><GlobalOutlined /></template>
          </a-input>
        </a-form-item>
        <p class="market-config-tip">{{ $t('theme.apiModal.tip') }}<a @click="goOfficialConfig">{{ $t('theme.apiModal.officialLink') }}</a></p>
      </a-form>
    </a-modal>

    <template v-if="store.loggedIn">
      <div class="page-header">
        <h1 class="page-title">{{ $t('theme.page.title') }}</h1>
        <div class="market-header-actions">
          <span v-if="store.marketUser" class="market-account">
            <UserOutlined /> {{ store.marketUser.username }}
          </span>
          <a-button @click="openMarketConfig">
            <template #icon><SettingOutlined /></template>
            {{ $t('theme.header.officialUrl') }}
          </a-button>
          <a-button @click="handleLogout">
            <template #icon><LogoutOutlined /></template>
            {{ $t('theme.header.logout') }}
          </a-button>
        </div>
      </div>

      <!-- 市场统计 -->
      <div class="market-stats">
        <div class="market-stat">
          <span class="market-stat__value">{{ store.stats?.total ?? '—' }}</span>
          <span class="market-stat__label">{{ $t('theme.stats.total') }}</span>
        </div>
        <div class="market-stat">
          <span class="market-stat__value">{{ store.stats?.authors ?? '—' }}</span>
          <span class="market-stat__label">{{ $t('theme.stats.authors') }}</span>
        </div>
        <div class="market-stat">
          <span class="market-stat__value">{{ store.stats?.downloads ?? '—' }}</span>
          <span class="market-stat__label">{{ $t('theme.stats.downloads') }}</span>
        </div>
      </div>

      <a-tabs v-model:active-key="store.activeTab" @change="handleTabChange">
        <!-- 主题市场 -->
        <a-tab-pane key="market" :tab="$t('menu.theme.market')">
          <div class="filter-bar">
            <a-select
              v-model:value="store.filters.type"
              :placeholder="$t('theme.filter.allTypes')"
              allow-clear
              style="width: 130px"
              :options="typeOptions"
              @change="store.fetchListWithReset()"
            />
            <a-select
              v-model:value="store.filters.styles"
              mode="multiple"
              :placeholder="$t('theme.filter.styles')"
              allow-clear
              :max-tag-count="2"
              style="min-width: 170px"
              :options="styleOptions"
              @change="store.fetchListWithReset()"
            />
            <a-select
              v-model:value="store.filters.price"
              :placeholder="$t('theme.filter.price')"
              allow-clear
              style="width: 120px"
              :options="priceOptions"
              @change="store.fetchListWithReset()"
            />
            <a-select
              v-model:value="store.filters.sort"
              style="width: 130px"
              :options="sortOptions"
              @change="store.fetchListWithReset()"
            />
            <a-input-search
              v-model:value="store.filters.search"
              :placeholder="$t('theme.filter.searchPlaceholder')"
              style="width: 240px"
              allow-clear
              :loading="store.loading"
              @search="store.fetchListWithReset()"
            />
          </div>

          <!-- 加载骨架屏 -->
          <div v-if="store.loading && store.list.length === 0" class="market-grid">
            <div v-for="i in 6" :key="i" class="skeleton-card">
              <a-skeleton active :paragraph="{ rows: 3 }" />
            </div>
          </div>

          <!-- 错误状态 -->
          <a-result
            v-else-if="store.loadError"
            status="error"
            :title="$t('common.state.error')"
            :sub-title="$t('theme.error.marketLoad')"
          >
            <template #extra>
              <a-button type="primary" @click="store.fetchList()">{{ $t('common.actions.retry') }}</a-button>
            </template>
          </a-result>

          <!-- 空状态 -->
          <a-empty v-else-if="store.list.length === 0" :description="$t('theme.empty.noMatch')" />

          <!-- 主题卡片网格 -->
          <div v-else class="market-grid">
            <ThemeCard
              v-for="theme in store.list"
              :key="theme.id"
              :theme="theme"
              :favorited="store.favoriteIds.has(theme.id)"
              :installed="installedMarketIds.has(theme.id)"
              :installing="installingId === theme.id"
              @detail="goDetail(theme)"
              @favorite="store.toggleFavorite(theme.id)"
              @install="handleInstall(theme)"
            />
          </div>

          <!-- 分页 -->
          <div v-if="store.total > store.pagination.pageSize" class="pagination-wrapper">
            <a-pagination
              :current="store.pagination.page"
              :page-size="store.pagination.pageSize"
              :total="store.total"
              :page-size-options="['9', '18', '36']"
              show-size-changer
              :show-total="(total: number) => $t('theme.pagination.totalThemes', { total })"
              @change="handlePageChange"
              @show-size-change="handlePageSizeChange"
            />
          </div>
        </a-tab-pane>

        <!-- 我的收藏 -->
        <a-tab-pane key="favorites" :tab="$t('theme.tabs.favorites')">
          <div v-if="store.favLoading && store.favList.length === 0" class="market-grid">
            <div v-for="i in 6" :key="i" class="skeleton-card">
              <a-skeleton active :paragraph="{ rows: 3 }" />
            </div>
          </div>

          <a-result
            v-else-if="store.favLoadError"
            status="error"
            :title="$t('common.state.error')"
            :sub-title="$t('theme.error.favoritesLoad')"
          >
            <template #extra>
              <a-button type="primary" @click="store.fetchFavorites()">{{ $t('common.actions.retry') }}</a-button>
            </template>
          </a-result>

          <a-empty v-else-if="store.favList.length === 0" :description="$t('theme.empty.noFavorites')">
            <a-button type="primary" @click="store.activeTab = 'market'">{{ $t('theme.empty.goMarket') }}</a-button>
          </a-empty>

          <div v-else class="market-grid">
            <ThemeCard
              v-for="theme in store.favList"
              :key="theme.id"
              :theme="theme"
              favorited
              :installed="installedMarketIds.has(theme.id)"
              :installing="installingId === theme.id"
              @detail="goDetail(theme)"
              @favorite="store.toggleFavorite(theme.id)"
              @install="handleInstall(theme)"
            />
          </div>

          <div v-if="store.favTotal > store.favPagination.pageSize" class="pagination-wrapper">
            <a-pagination
              :current="store.favPagination.page"
              :page-size="store.favPagination.pageSize"
              :total="store.favTotal"
              :page-size-options="['9', '18', '36']"
              show-size-changer
              :show-total="(total: number) => $t('theme.pagination.totalFavorites', { total })"
              @change="handleFavPageChange"
              @show-size-change="handleFavPageSizeChange"
            />
          </div>
        </a-tab-pane>

        <!-- 已安装主题 -->
        <a-tab-pane key="installed" :tab="$t('theme.tabs.installed')">
          <div v-if="installedLoading && installedList.length === 0" class="market-grid">
            <div v-for="i in 6" :key="i" class="skeleton-card">
              <a-skeleton active :paragraph="{ rows: 3 }" />
            </div>
          </div>

          <a-result
            v-else-if="installedError"
            status="error"
            :title="$t('common.state.error')"
            :sub-title="$t('theme.error.installedLoad')"
          >
            <template #extra>
              <a-button type="primary" @click="fetchInstalled()">{{ $t('common.actions.retry') }}</a-button>
            </template>
          </a-result>

          <a-empty v-else-if="installedList.length === 0" :description="$t('theme.empty.noInstalled')">
            <a-button type="primary" @click="store.activeTab = 'market'">{{ $t('theme.empty.goInstall') }}</a-button>
          </a-empty>

          <div v-else class="market-grid">
            <InstalledThemeCard
              v-for="theme in installedList"
              :key="theme.id"
              :theme="theme"
              :market="theme.market_id > 0 ? installedMarketMap.get(theme.market_id) : undefined"
              :activating="activatingId === theme.id"
              :updating="updatingId === theme.id"
              @activate="handleActivate(theme)"
              @preview="handlePreview(theme)"
              @update="handleUpdate(theme)"
              @uninstall="handleUninstall(theme)"
            />
          </div>

          <p class="installed-tip">
            <GlobalOutlined />
            {{ $t('theme.installedTip') }}
          </p>
        </a-tab-pane>
      </a-tabs>
    </template>

    <!-- 未登录且已关闭登录弹窗时的占位提示 -->
    <div v-else class="market-login-required">
      <p class="market-login-required__text">{{ $t('theme.login.required') }}</p>
      <a-button type="link" @click="loginModalDismissed = false">{{ $t('theme.login.action') }}</a-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Modal, message } from 'ant-design-vue'
import {
  GlobalOutlined,
  LockOutlined,
  LogoutOutlined,
  SettingOutlined,
  SkinOutlined,
  UserOutlined,
} from '@ant-design/icons-vue'
import { MARKET_AUTH_EXPIRED_EVENT, abortMarketRetries, themeApi, themeMarketApi } from '@/api/theme'
import { configApi } from '@/api/config'
import { themeDisplayName } from '@/utils/themeDisplay'
import type { InstalledTheme, ThemeItem } from '@/types/theme'
import { useThemeMarketStore } from '@/stores/themeMarket'
import ThemeCard from '@/components/theme/ThemeCard.vue'
import InstalledThemeCard from '@/components/theme/InstalledThemeCard.vue'

const { t } = useI18n()
const router = useRouter()
const store = useThemeMarketStore()

/** 跳转主题详情页（含版本历史） */
function goDetail(theme: ThemeItem) {
  router.push(`/themes/market/${theme.id}`)
}

// ===== 已安装主题 =====
const installedList = ref<InstalledTheme[]>([])
const installedLoading = ref(false)
const installedError = ref(false)
const installingId = ref<number | null>(null)
const activatingId = ref<string | null>(null)
const updatingId = ref<string | null>(null)
// 博客前台独立域名时经 VITE_BLOG_BASE_URL 指定（同域部署留空）
const blogBase = import.meta.env.VITE_BLOG_BASE_URL || ''

// installedMarketIds 已安装主题对应的市场主题 ID 集合（builtin/手动导入主题无 market_id，不计入），
// 供市场与收藏卡片标记"已安装"状态
const installedMarketIds = computed(() => {
  const ids = new Set<number>()
  for (const t of installedList.value) {
    if (t.market_id > 0) ids.add(t.market_id)
  }
  return ids
})

// installedMarketMap market_id → 市场主题元数据，已安装卡片据此对齐市场展示（封面/标题/描述）
const installedMarketMap = ref<Map<number, ThemeItem>>(new Map())

/** 补拉已安装主题的市场元数据（市场列表分页内未必包含），
 * 拉取失败（市场不可达/登录失效）时静默，卡片回退展示制品内信息 */
async function hydrateInstalledMarketMeta() {
  const missing = installedList.value.filter(
    (t) => t.market_id > 0 && !installedMarketMap.value.has(t.market_id),
  )
  await Promise.all(
    missing.map(async (t) => {
      try {
        const detail = await themeMarketApi.getDetail(t.market_id)
        const next = new Map(installedMarketMap.value)
        next.set(t.market_id, detail)
        installedMarketMap.value = next
      } catch {
        // 静默：保持渐变占位展示
      }
    }),
  )
}

async function fetchInstalled() {
  installedLoading.value = true
  installedError.value = false
  try {
    installedList.value = await themeApi.getList()
    // 异步补齐市场元数据，不阻塞列表渲染
    hydrateInstalledMarketMeta()
  } catch {
    installedError.value = true
  } finally {
    installedLoading.value = false
  }
}

/** 从官方市场安装主题（长任务：下载/校验/解压/落库） */
function handleInstall(theme: ThemeItem) {
  const name = themeDisplayName(theme.title, theme.slug)
  Modal.confirm({
    title: t('theme.modal.installTitle', { name }),
    content: t('theme.modal.installContent'),
    okText: t('theme.action.install'),
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      installingId.value = theme.id
      try {
        await themeApi.install({ theme_id: theme.id })
        message.success(t('theme.modal.installSuccess', { name }))
        fetchInstalled()
      } catch {
        // 失败原因（制品缺失/引擎不支持等）由拦截器统一提示
      } finally {
        installingId.value = null
      }
    },
  })
}

/** 启用主题（切换后访客侧立即生效） */
function handleActivate(theme: InstalledTheme) {
  Modal.confirm({
    title: t('theme.modal.activateTitle', { name: theme.name, version: theme.version }),
    content: t('theme.modal.activateContent'),
    okText: t('common.actions.enable'),
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      activatingId.value = theme.id
      try {
        await themeApi.activate(theme.id)
        message.success(t('theme.modal.activateSuccess', { name: theme.name }))
        fetchInstalled()
      } catch {
        // 错误由拦截器统一提示
      } finally {
        activatingId.value = null
      }
    },
  })
}

/** 预览主题（新窗口打开预览路由） */
function handlePreview(theme: InstalledTheme) {
  window.open(`${blogBase}/preview/${theme.theme_id}/`, '_blank', 'noopener')
}

/** 卸载主题（使用中的主题由后端拒绝并在卡片上禁用） */
function handleUninstall(theme: InstalledTheme) {
  installingId.value = null
  themeApi
    .uninstall(theme.id)
    .then(() => {
      message.success(t('theme.modal.uninstallSuccess', { name: theme.name, version: theme.version }))
      fetchInstalled()
    })
    .catch(() => {
      // 错误由拦截器统一提示
    })
}

/** 从官方市场更新主题到最新版本 */
async function handleUpdate(theme: InstalledTheme) {
  Modal.confirm({
    title: t('theme.modal.updateTitle', { name: theme.name }),
    content: t('theme.modal.updateContent'),
    okText: t('common.actions.update'),
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      updatingId.value = theme.id
      try {
        await themeApi.updateTheme(theme.id)
        message.success(t('theme.modal.updateSuccess', { name: theme.name }))
        fetchInstalled()
      } catch (e: any) {
        // "已是最新版本"等错误由拦截器统一提示
      } finally {
        updatingId.value = null
      }
    },
  })
}

// ===== 登录弹窗 =====
/** 用户主动关闭登录弹窗后的占位态（内存态，刷新页面后未登录仍会重新弹出） */
const loginModalDismissed = ref(false)
/** 官方站点注册页（去注册按钮跳转地址） */
const OFFICIAL_REGISTER_URL = 'http://www.novablog.cn/register'
const loginLoading = ref(false)
const loginForm = reactive({
  email: '',
  password: '',
})

const emailRules = computed(() => [
  { required: true, message: t('theme.validation.emailRequired'), trigger: 'blur' },
  { pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: t('theme.validation.emailInvalid'), trigger: 'blur' },
])
const passwordRules = computed(() => [{ required: true, message: t('theme.validation.passwordRequired'), trigger: 'blur' }])

async function handleLogin() {
  loginLoading.value = true
  try {
    await store.login(loginForm.email.trim(), loginForm.password)
    loginForm.password = ''
    message.success(t('theme.message.loginSuccess'))
    // 登录后补拉已安装列表，市场卡片的"已安装"状态立即可用
    fetchInstalled()
  } catch {
    // 登录错误（如邮箱或密码错误）由响应拦截器统一提示
  } finally {
    loginLoading.value = false
  }
}

/** 去注册：新标签页打开官方站点注册页 */
function handleGoRegister() {
  window.open(OFFICIAL_REGISTER_URL, '_blank', 'noopener')
}

/** 前往系统配置的社区推送页修改官方地址（社区推送与主题市场共用，深链直达对应标签页） */
function goOfficialConfig() {
  router.push({ path: '/system-config', query: { tab: 'community' } })
}

/** 关闭官方登录弹窗：因登录失效挂起的请求直接拒绝，避免安装/更新按钮无限 loading */
function handleLoginModalClose() {
  loginModalDismissed.value = true
  abortMarketRetries()
}

function handleLogout() {
  Modal.confirm({
    title: t('theme.modal.logoutTitle'),
    content: t('theme.modal.logoutContent'),
    okText: t('theme.modal.logoutOk'),
    cancelText: t('common.actions.cancel'),
    onOk: () => store.logout(),
  })
}

// ===== 博客 API 地址设置（官方站点地址统一在社区推送配置页维护）=====
const marketConfigOpen = ref(false)
const marketAPIURL = ref('')
const marketConfigSaving = ref(false)
const apiURLError = ref('')

async function openMarketConfig() {
  marketAPIURL.value = ''
  apiURLError.value = ''
  marketConfigOpen.value = true
  // 回显后端持久化的博客 API 地址
  try {
    const cfg = await configApi.getThemeMarketConfig()
    marketAPIURL.value = cfg.public_api_base || ''
  } catch {
    // 回显失败时保留空值
  }
}

async function handleSaveMarketConfig() {
  const apiURL = marketAPIURL.value.trim()
  if (apiURL && !/^https?:\/\//.test(apiURL)) {
    apiURLError.value = t('theme.validation.apiURLFormat')
    return
  }
  marketConfigSaving.value = true
  try {
    await configApi.updateThemeMarketConfig({ public_api_base: apiURL })
    marketConfigOpen.value = false
    message.success(t('theme.apiModal.saved'))
  } catch {
    // 错误由拦截器统一提示
  } finally {
    marketConfigSaving.value = false
  }
}

// 官方登录失效（Token 刷新失败）时重新弹出登录弹窗
function handleAuthExpiredEvent() {
  store.handleAuthExpired()
  loginModalDismissed.value = false
}

// ===== 市场浏览 =====
// typeLabels 主题类型展示文案（i18n），未知类型回退原始值
const typeLabels = computed<Record<string, string>>(() => ({
  personal: t('theme.type.personal'),
  tech: t('theme.type.tech'),
  photo: t('theme.type.photo'),
  travel: t('theme.type.travel'),
  music: t('theme.type.music'),
  team: t('theme.type.team'),
}))
const typeOptions = computed(() =>
  Object.entries(typeLabels.value).map(([value, label]) => ({ value, label })),
)
const styleOptions = computed(() =>
  store.hotTags.map((t) => ({ value: t.name, label: `${t.name} (${t.count})` })),
)
const priceOptions = computed(() => [
  { value: 'free', label: t('theme.price.free') },
  { value: 'paid', label: t('theme.price.paid') },
])
const sortOptions = computed(() => [
  { value: 'latest', label: t('theme.sort.latest') },
  { value: 'downloads', label: t('theme.sort.downloads') },
  { value: 'rating', label: t('theme.sort.rating') },
])

function handleTabChange(key: string | number) {
  if (key === 'favorites') store.fetchFavorites()
  if (key === 'installed') fetchInstalled()
}

function handlePageChange(page: number) {
  store.pagination.page = page
  store.fetchList()
}

function handlePageSizeChange(_page: number, size: number) {
  store.pagination.pageSize = size
  store.pagination.page = 1
  store.fetchList()
}

function handleFavPageChange(page: number) {
  store.favPagination.page = page
  store.fetchFavorites()
}

function handleFavPageSizeChange(_page: number, size: number) {
  store.favPagination.pageSize = size
  store.favPagination.page = 1
  store.fetchFavorites()
}

onMounted(async () => {
  window.addEventListener(MARKET_AUTH_EXPIRED_EVENT, handleAuthExpiredEvent)
  // 官方站点地址由后端统一配置，仅拉取用于展示
  await store.fetchDefaultBaseURL()
  if (store.loggedIn) {
    // 页面刷新后恢复数据（登录态持久于 localStorage）
    if (!store.list.length) store.fetchList()
    if (!store.stats) store.fetchStats()
    if (!store.hotTags.length) store.fetchHotTags()
    if (store.favoriteIds.size === 0) store.fetchFavoriteIds()
    // 预拉已安装列表：市场/收藏卡片的"已安装"状态依赖它
    fetchInstalled()
  }
})

onUnmounted(() => {
  window.removeEventListener(MARKET_AUTH_EXPIRED_EVENT, handleAuthExpiredEvent)
})
</script>

<style scoped>
/* ===== 登录遮罩 ===== */
.market-login {
  padding: 8px 4px 4px;
  text-align: center;
}

.market-login__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  margin-bottom: 12px;
  border-radius: 16px;
  background: linear-gradient(135deg, #526FE8 0%, #8EA5FF 100%);
  color: #fff;
  font-size: 26px;
}

.market-login__title {
  margin: 0 0 6px;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
}

.market-login__desc {
  margin: 0 0 12px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
}

.market-login__official {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0 0 6px;
  padding: 7px 10px;
  background: #F5F7FB;
  border: 1px solid #E5E9F2;
  border-radius: 6px;
  font-size: 12px;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
}

.market-login__official-url {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.market-login__official-hint {
  margin: 0 0 16px;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
}

.market-login :deep(.ant-form) {
  text-align: left;
}

.market-login__actions {
  display: flex;
  gap: 12px;
  margin-top: 12px;
}

/* ===== 未登录占位提示 ===== */
.market-login-required {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 140px 16px;
}

.market-login-required__text {
  margin: 0;
  font-size: 15px;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
}

/* ===== 官方地址设置 ===== */
.market-config-tip {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.45));
}

/* ===== 页头 ===== */
.market-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.market-account {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 999px;
  background: var(--bg-layout, rgba(0, 0, 0, 0.04));
  color: var(--text-color, rgba(0, 0, 0, 0.88));
  font-size: 13px;
}

/* ===== 统计条 ===== */
.market-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}

.market-stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 14px 18px;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: var(--border-radius-lg, 12px);
}

.market-stat__value {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
}

.market-stat__label {
  font-size: 12px;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
}

/* ===== 筛选栏与卡片网格 ===== */
.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.market-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

@media (max-width: 1400px) {
  .market-grid { grid-template-columns: repeat(3, 1fr); }
}

@media (max-width: 1100px) {
  .market-grid { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 768px) {
  .market-grid { grid-template-columns: 1fr; }
  .market-stats { flex-direction: column; }
}

.skeleton-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: var(--border-radius-lg, 12px);
  padding: 16px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 32px;
}

.installed-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 24px;
  font-size: 12px;
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
}
</style>
