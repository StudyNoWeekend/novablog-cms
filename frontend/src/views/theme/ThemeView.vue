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
        <h2 class="market-login__title">登录 NovaBlog 官方账号</h2>
        <p class="market-login__desc">连接官方主题市场后，可浏览主题并进行点赞、收藏、评分与下载</p>
        <a-form :model="loginForm" layout="vertical" autocomplete="off" @finish="handleLogin">
          <a-form-item label="官方地址" name="baseURL" :rules="baseURLRules">
            <a-input v-model:value="loginForm.baseURL" :placeholder="store.defaultBaseURL || '请输入官方市场地址'" allow-clear>
              <template #prefix><LinkOutlined /></template>
            </a-input>
          </a-form-item>
          <a-form-item label="邮箱" name="email" :rules="emailRules">
            <a-input v-model:value="loginForm.email" placeholder="请输入官方账号邮箱" allow-clear>
              <template #prefix><UserOutlined /></template>
            </a-input>
          </a-form-item>
          <a-form-item label="密码" name="password" :rules="passwordRules">
            <a-input-password v-model:value="loginForm.password" placeholder="请输入密码">
              <template #prefix><LockOutlined /></template>
            </a-input-password>
          </a-form-item>
          <a-button type="primary" html-type="submit" block :loading="loginLoading">登 录</a-button>
        </a-form>
        <div class="market-login__actions">
          <a-button block @click="handleGoRegister">去注册</a-button>
          <a-button block @click="handleLoginModalClose">关 闭</a-button>
        </div>
      </div>
    </a-modal>

    <!-- 服务地址设置弹窗 -->
    <a-modal
      v-model:open="marketConfigOpen"
      title="官方地址与博客 API"
      :confirm-loading="marketConfigSaving"
      ok-text="保存"
      cancel-text="取消"
      @ok="handleSaveMarketConfig"
    >
      <a-form layout="vertical" autocomplete="off">
        <a-form-item label="官方市场地址" :validate-status="marketConfigError ? 'error' : ''" :help="marketConfigError">
          <a-input v-model:value="marketConfigURL" :placeholder="store.defaultBaseURL || '请输入官方市场地址'" allow-clear>
            <template #prefix><LinkOutlined /></template>
          </a-input>
        </a-form-item>
        <a-form-item label="博客 API 地址（可选）" :validate-status="apiURLError ? 'error' : ''" :help="apiURLError">
          <a-input v-model:value="marketAPIURL" placeholder="https://api.example.com（留空 = 主题同域取数）" allow-clear>
            <template #prefix><GlobalOutlined /></template>
          </a-input>
        </a-form-item>
        <p class="market-config-tip">官方市场地址决定主题市场与主题安装的取数来源；博客 API 地址用于主题前台访问公开接口——填写独立 API 域名后保存，已安装主题会立即重新注入取数地址，无需重装主题；留空则同域相对路径取数</p>
      </a-form>
    </a-modal>

    <template v-if="store.loggedIn">
      <div class="page-header">
        <h1 class="page-title">主题</h1>
        <div class="market-header-actions">
          <span v-if="store.marketUser" class="market-account">
            <UserOutlined /> {{ store.marketUser.username }}
          </span>
          <a-button @click="openMarketConfig">
            <template #icon><SettingOutlined /></template>
            官方地址
          </a-button>
          <a-button @click="handleLogout">
            <template #icon><LogoutOutlined /></template>
            退出登录
          </a-button>
        </div>
      </div>

      <!-- 市场统计 -->
      <div class="market-stats">
        <div class="market-stat">
          <span class="market-stat__value">{{ store.stats?.total ?? '—' }}</span>
          <span class="market-stat__label">主题总数</span>
        </div>
        <div class="market-stat">
          <span class="market-stat__value">{{ store.stats?.authors ?? '—' }}</span>
          <span class="market-stat__label">作者数</span>
        </div>
        <div class="market-stat">
          <span class="market-stat__value">{{ store.stats?.downloads ?? '—' }}</span>
          <span class="market-stat__label">安装次数</span>
        </div>
      </div>

      <a-tabs v-model:active-key="store.activeTab" @change="handleTabChange">
        <!-- 主题市场 -->
        <a-tab-pane key="market" tab="主题市场">
          <div class="filter-bar">
            <a-select
              v-model:value="store.filters.type"
              placeholder="全部类型"
              allow-clear
              style="width: 130px"
              :options="typeOptions"
              @change="store.fetchListWithReset()"
            />
            <a-select
              v-model:value="store.filters.styles"
              mode="multiple"
              placeholder="风格筛选"
              allow-clear
              :max-tag-count="2"
              style="min-width: 170px"
              :options="styleOptions"
              @change="store.fetchListWithReset()"
            />
            <a-select
              v-model:value="store.filters.price"
              placeholder="免费/付费"
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
              placeholder="搜索主题名称或描述..."
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
            title="加载失败"
            sub-title="获取官方主题市场数据时出错，请重试"
          >
            <template #extra>
              <a-button type="primary" @click="store.fetchList()">重试</a-button>
            </template>
          </a-result>

          <!-- 空状态 -->
          <a-empty v-else-if="store.list.length === 0" description="没有符合条件的主题" />

          <!-- 主题卡片网格 -->
          <div v-else class="market-grid">
            <ThemeCard
              v-for="theme in store.list"
              :key="theme.id"
              :theme="theme"
              :favorited="store.favoriteIds.has(theme.id)"
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
              :show-total="(total: number) => `共 ${total} 个主题`"
              @change="handlePageChange"
              @show-size-change="handlePageSizeChange"
            />
          </div>
        </a-tab-pane>

        <!-- 我的收藏 -->
        <a-tab-pane key="favorites" tab="我的收藏">
          <div v-if="store.favLoading && store.favList.length === 0" class="market-grid">
            <div v-for="i in 6" :key="i" class="skeleton-card">
              <a-skeleton active :paragraph="{ rows: 3 }" />
            </div>
          </div>

          <a-result
            v-else-if="store.favLoadError"
            status="error"
            title="加载失败"
            sub-title="获取收藏列表时出错，请重试"
          >
            <template #extra>
              <a-button type="primary" @click="store.fetchFavorites()">重试</a-button>
            </template>
          </a-result>

          <a-empty v-else-if="store.favList.length === 0" description="暂无收藏">
            <a-button type="primary" @click="store.activeTab = 'market'">去主题市场逛逛</a-button>
          </a-empty>

          <div v-else class="market-grid">
            <ThemeCard
              v-for="theme in store.favList"
              :key="theme.id"
              :theme="theme"
              favorited
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
              :show-total="(total: number) => `共 ${total} 个收藏`"
              @change="handleFavPageChange"
              @show-size-change="handleFavPageSizeChange"
            />
          </div>
        </a-tab-pane>

        <!-- 已安装主题 -->
        <a-tab-pane key="installed" tab="已安装主题">
          <div v-if="installedLoading && installedList.length === 0" class="market-grid">
            <div v-for="i in 6" :key="i" class="skeleton-card">
              <a-skeleton active :paragraph="{ rows: 3 }" />
            </div>
          </div>

          <a-result
            v-else-if="installedError"
            status="error"
            title="加载失败"
            sub-title="获取已安装主题时出错，请重试"
          >
            <template #extra>
              <a-button type="primary" @click="fetchInstalled()">重试</a-button>
            </template>
          </a-result>

          <a-empty v-else-if="installedList.length === 0" description="尚未安装任何主题">
            <a-button type="primary" @click="store.activeTab = 'market'">去主题市场安装</a-button>
          </a-empty>

          <div v-else class="market-grid">
            <InstalledThemeCard
              v-for="theme in installedList"
              :key="theme.id"
              :theme="theme"
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
            启用后访客侧立即生效；「预览」在新窗口打开主题预览页，激活前即可对比真实效果
          </p>
        </a-tab-pane>
      </a-tabs>
    </template>

    <!-- 未登录且已关闭登录弹窗时的占位提示 -->
    <div v-else class="market-login-required">
      <p class="market-login-required__text">获取主题市场内容需要登录</p>
      <a-button type="link" @click="loginModalDismissed = false">登陆</a-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Modal, message } from 'ant-design-vue'
import {
  GlobalOutlined,
  LinkOutlined,
  LockOutlined,
  LogoutOutlined,
  SettingOutlined,
  SkinOutlined,
  UserOutlined,
} from '@ant-design/icons-vue'
import { MARKET_AUTH_EXPIRED_EVENT, abortMarketRetries } from '@/api/theme'
import { configApi } from '@/api/config'
import { themeApi } from '@/api/theme'
import { marketStorage } from '@/utils/storage'
import { THEME_TYPE_LABELS } from '@/utils/themeDisplay'
import type { InstalledTheme, ThemeItem } from '@/types/theme'
import { useThemeMarketStore } from '@/stores/themeMarket'
import ThemeCard from '@/components/theme/ThemeCard.vue'
import InstalledThemeCard from '@/components/theme/InstalledThemeCard.vue'

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

async function fetchInstalled() {
  installedLoading.value = true
  installedError.value = false
  try {
    installedList.value = await themeApi.getList()
  } catch {
    installedError.value = true
  } finally {
    installedLoading.value = false
  }
}

/** 从官方市场安装主题（长任务：下载/校验/解压/落库） */
function handleInstall(theme: ThemeItem) {
  Modal.confirm({
    title: `安装「${theme.title}」？`,
    content: '将从官方市场下载预构建制品并安装，完成后可在「已安装主题」中启用',
    okText: '安装',
    cancelText: '取消',
    onOk: async () => {
      installingId.value = theme.id
      try {
        await themeApi.install({ theme_id: theme.id })
        message.success(`「${theme.title}」安装成功，可在「已安装主题」中启用`)
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
    title: `启用「${theme.name} v${theme.version}」？`,
    content: '切换后博客前台将立即更新外观，确认切换？',
    okText: '启用',
    cancelText: '取消',
    onOk: async () => {
      activatingId.value = theme.id
      try {
        await themeApi.activate(theme.id)
        message.success(`已启用「${theme.name}」，访客侧已生效`)
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
      message.success(`已卸载「${theme.name} v${theme.version}」`)
      fetchInstalled()
    })
    .catch(() => {
      // 错误由拦截器统一提示
    })
}

/** 从官方市场更新主题到最新版本 */
async function handleUpdate(theme: InstalledTheme) {
  Modal.confirm({
    title: `更新「${theme.name}」？`,
    content: '将从官方市场拉取最新版本并重新部署，如果正在使用中将自动切换激活。',
    okText: '更新',
    cancelText: '取消',
    onOk: async () => {
      updatingId.value = theme.id
      try {
        await themeApi.updateTheme(theme.id)
        message.success(`「${theme.name}」已更新`)
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
const OFFICIAL_REGISTER_URL = 'http://novablog.ditancafebar.cn/register'
const loginLoading = ref(false)
const loginForm = reactive({
  baseURL: marketStorage.getBaseURL() || '',
  email: '',
  password: '',
})

const baseURLRules = [
  { required: true, message: '请输入官方地址', trigger: 'blur' },
  { pattern: /^https?:\/\//, message: '官方地址需以 http:// 或 https:// 开头', trigger: 'blur' },
]
const emailRules = [
  { required: true, message: '请输入邮箱', trigger: 'blur' },
  { pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: '邮箱格式不正确', trigger: 'blur' },
]
const passwordRules = [{ required: true, message: '请输入密码', trigger: 'blur' }]

async function handleLogin() {
  loginLoading.value = true
  try {
    await store.login(loginForm.baseURL.trim(), loginForm.email.trim(), loginForm.password)
    loginForm.password = ''
    message.success('已连接官方主题市场')
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

/** 关闭官方登录弹窗：因登录失效挂起的请求直接拒绝，避免安装/更新按钮无限 loading */
function handleLoginModalClose() {
  loginModalDismissed.value = true
  abortMarketRetries()
}

function handleLogout() {
  Modal.confirm({
    title: '退出官方账号？',
    content: '退出后需重新登录才能浏览主题市场',
    okText: '退出',
    cancelText: '取消',
    onOk: () => store.logout(),
  })
}

// ===== 服务地址设置（官方市场地址 + 博客公开 API 地址）=====
const marketConfigOpen = ref(false)
const marketConfigURL = ref('')
const marketAPIURL = ref('')
const marketConfigSaving = ref(false)
const marketConfigError = ref('')
const apiURLError = ref('')

async function openMarketConfig() {
  marketConfigURL.value = store.defaultBaseURL
  marketAPIURL.value = ''
  marketConfigError.value = ''
  apiURLError.value = ''
  marketConfigOpen.value = true
  // 回显后端持久化值（两个地址）
  try {
    const cfg = await configApi.getThemeMarketConfig()
    if (cfg.market_base_url) marketConfigURL.value = cfg.market_base_url
    marketAPIURL.value = cfg.public_api_base || ''
  } catch {
    // 回显失败时保留本地默认值
  }
}

async function handleSaveMarketConfig() {
  const url = marketConfigURL.value.trim()
  if (!/^https?:\/\//.test(url)) {
    marketConfigError.value = '官方地址需以 http:// 或 https:// 开头'
    return
  }
  const apiURL = marketAPIURL.value.trim()
  if (apiURL && !/^https?:\/\//.test(apiURL)) {
    apiURLError.value = 'API 地址需以 http:// 或 https:// 开头，留空表示同域取数'
    return
  }
  marketConfigSaving.value = true
  try {
    await configApi.updateThemeMarketConfig({ market_base_url: url, public_api_base: apiURL })
    store.defaultBaseURL = url
    store.marketBaseURL = url
    // 本浏览器后续安装/更新请求头同步使用新地址
    marketStorage.setBaseURL(url)
    marketConfigOpen.value = false
    message.success('服务地址已保存并全局生效')
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
const typeOptions = computed(() =>
  Object.entries(THEME_TYPE_LABELS).map(([value, label]) => ({ value, label })),
)
const styleOptions = computed(() =>
  store.hotTags.map((t) => ({ value: t.name, label: `${t.name} (${t.count})` })),
)
const priceOptions = [
  { value: 'free', label: '免费' },
  { value: 'paid', label: '付费' },
]
const sortOptions = [
  { value: 'latest', label: '最新上架' },
  { value: 'downloads', label: '下载最多' },
  { value: 'rating', label: '评分最高' },
]

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
  // 官方地址默认值由后端下发，登录弹窗预填
  await store.fetchDefaultBaseURL()
  if (!loginForm.baseURL && store.defaultBaseURL) loginForm.baseURL = store.defaultBaseURL
  if (store.loggedIn) {
    // 页面刷新后恢复数据（登录态持久于 localStorage）
    if (!store.list.length) store.fetchList()
    if (!store.stats) store.fetchStats()
    if (!store.hotTags.length) store.fetchHotTags()
    if (store.favoriteIds.size === 0) store.fetchFavoriteIds()
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
  background: linear-gradient(135deg, #6a11cb 0%, #2575fc 100%);
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
  margin: 0 0 20px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
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
  border: 1px solid var(--border-color, #f0f0f0);
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
  border: 1px solid var(--border-color, #f0f0f0);
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
