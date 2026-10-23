<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('cloudDrive.title') }}</h1>
    </div>

    <a-card>
      <a-tabs v-model:activeKey="activeTab">
        <!-- WebDAV -->
        <a-tab-pane key="webdav">
          <template #tab>
            <span><CloudServerOutlined /> {{ $t('cloudDrive.tabs.webdav') }}</span>
          </template>

          <a-alert
            class="provider-hint"
            type="info"
            show-icon
            :message="$t('cloudDrive.webdav.appPasswordHint')"
          />

          <div class="section-card">
            <h3 class="section-title">{{ $t('cloudDrive.webdav.formTitle') }}</h3>
            <a-form layout="vertical" class="bind-form">
              <a-form-item :label="$t('cloudDrive.webdav.baseUrl')" required>
                <a-input v-model:value="webdavForm.base_url" :placeholder="$t('cloudDrive.webdav.baseUrlPlaceholder')" />
              </a-form-item>
              <a-form-item :label="$t('cloudDrive.webdav.username')" required>
                <a-input v-model:value="webdavForm.username" autocomplete="off" />
              </a-form-item>
              <a-form-item :label="$t('cloudDrive.webdav.password')" required>
                <a-input-password v-model:value="webdavForm.password" autocomplete="new-password" />
              </a-form-item>
              <a-space>
                <a-button type="primary" :loading="webdavSaving" @click="handleAddWebDAV">
                  <LinkOutlined /> {{ $t('cloudDrive.webdav.add') }}
                </a-button>
                <a-button :loading="webdavTesting" @click="handleTestWebDAV">
                  {{ $t('cloudDrive.webdav.test') }}
                </a-button>
              </a-space>
            </a-form>
          </div>
        </a-tab-pane>

        <!-- 百度网盘 / 阿里云盘：应用凭据配置 -->
        <a-tab-pane v-for="p in oauthProviders" :key="p">
          <template #tab>
            <span><CloudOutlined /> {{ $t(`cloudDrive.tabs.${p}`) }}</span>
          </template>

          <div class="section-card">
            <div class="section-head">
              <h3 class="section-title">{{ $t('cloudDrive.config.title') }}</h3>
              <a class="apply-link" :href="applyPortals[p]" target="_blank" rel="noopener">
                <ExportOutlined /> {{ $t('cloudDrive.config.applyPortal') }}
              </a>
            </div>
            <a-form layout="vertical" class="bind-form">
              <a-form-item :label="$t('cloudDrive.config.clientId')" required>
                <a-input v-model:value="oauthForms[p].client_id" autocomplete="off" />
              </a-form-item>
              <a-form-item :label="$t('cloudDrive.config.clientSecret')" :extra="$t('cloudDrive.config.secretKeep')">
                <a-input-password
                  v-model:value="oauthForms[p].client_secret"
                  :placeholder="oauthForms[p].masked ? '******' : ''"
                  autocomplete="new-password"
                />
              </a-form-item>
              <a-form-item :label="$t('cloudDrive.config.callbackUrl')" :extra="$t('cloudDrive.config.callbackBaseHint')">
                <a-input :value="callbackURI(p)" readonly class="callback-input">
                  <template #suffix>
                    <a class="copy-btn" @click.prevent="copyCallback(callbackURI(p))">
                      <CopyOutlined />
                    </a>
                  </template>
                </a-input>
              </a-form-item>
              <a-button type="primary" :loading="oauthSaving === p" @click="handleSaveConfig(p)">
                <SaveOutlined /> {{ $t('cloudDrive.config.save') }}
              </a-button>
            </a-form>
          </div>
        </a-tab-pane>
      </a-tabs>

      <!-- 已绑定账号（按当前 Tab 的 provider 过滤，三 Tab 共用） -->
      <div class="section-card">
        <div class="accounts-head">
          <h3 class="section-title">{{ $t('cloudDrive.accounts.title') }}</h3>
          <a-button
            v-if="activeTab !== 'webdav'"
            type="primary"
            size="small"
            :disabled="oauthPending"
            @click="handleAuthorize(activeTab)"
          >
            <CloudDownloadOutlined /> {{ $t('cloudDrive.accounts.add') }}
          </a-button>
        </div>

        <div v-if="oauthPending" class="oauth-pending">
          <LoadingOutlined /> <span>{{ $t('cloudDrive.accounts.oauthPending') }}</span>
        </div>

        <a-spin :spinning="accountsLoading">
          <a-empty v-if="filteredAccounts.length === 0" :description="$t('cloudDrive.accounts.empty')" />
          <div v-else class="account-grid">
            <div v-for="account in filteredAccounts" :key="account.id" class="account-card">
              <div class="account-head">
                <UserOutlined class="account-icon" />
                <span class="account-name" :title="account.display_name || account.username">
                  {{ account.display_name || account.username }}
                </span>
                <span v-if="account.status === 'expired'" class="account-status expired">
                  {{ $t('cloudDrive.accounts.expired') }}
                </span>
              </div>
              <div class="account-meta" :title="account.base_url">
                <template v-if="account.provider === 'webdav'">{{ account.base_url }}</template>
                <template v-else>{{ $t('cloudDrive.accounts.oauthAccount') }}</template>
              </div>
              <div v-if="account.quota_total > 0" class="account-quota">
                <span class="quota-label">{{ $t('cloudDrive.accounts.quota') }}</span>
                <div class="quota-bar">
                  <div
                    class="quota-bar-inner"
                    :style="{ width: quotaPercent(account) + '%' }"
                  />
                </div>
                <span class="quota-text">{{ formatQuota(account.quota_used, account.quota_total) }}</span>
              </div>
              <div class="account-actions">
                <a-button size="small" danger @click="handleUnbind(account)">
                  <DeleteOutlined /> {{ $t('cloudDrive.accounts.unbind') }}
                </a-button>
              </div>
            </div>
          </div>
        </a-spin>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { message, Modal } from 'ant-design-vue'
import {
  CloudOutlined,
  CloudServerOutlined,
  CloudDownloadOutlined,
  CopyOutlined,
  ExportOutlined,
  LinkOutlined,
  SaveOutlined,
  UserOutlined,
  DeleteOutlined,
  LoadingOutlined,
} from '@ant-design/icons-vue'
import { cloudDriveApi } from '@/api/cloud-drive'
import type { CloudDriveAccount, CloudDriveProvider } from '@/types/cloud-drive'

const { t } = useI18n()

const oauthProviders = ['baidu', 'aliyun'] as const
type OAuthProvider = (typeof oauthProviders)[number]

/** 各开放平台申请入口（注册应用获取 AppKey/ClientID） */
const applyPortals: Record<OAuthProvider, string> = {
  baidu: 'https://pan.baidu.com/union',
  aliyun: 'https://www.alipan.com/developer',
}

const activeTab = ref<CloudDriveProvider>('webdav')
const accounts = ref<CloudDriveAccount[]>([])
const accountsLoading = ref(false)

interface OAuthFormState {
  client_id: string
  client_secret: string
  /** 后端返回脱敏 "******" 时，secret 输入框留空表示不修改 */
  masked: boolean
}
const oauthForms = reactive<Record<OAuthProvider, OAuthFormState>>({
  baidu: { client_id: '', client_secret: '', masked: false },
  aliyun: { client_id: '', client_secret: '', masked: false },
})
const oauthSaving = ref<OAuthProvider | ''>('')

const webdavForm = reactive({ base_url: '', username: '', password: '' })
const webdavSaving = ref(false)
const webdavTesting = ref(false)

/** 授权进行中（弹窗打开 + 轮询账号） */
const oauthPending = ref(false)
let oauthPollTimer: ReturnType<typeof setInterval> | null = null
let oauthPollDeadline = 0

const filteredAccounts = computed(() => accounts.value.filter((a) => a.provider === activeTab.value))

function quotaPercent(account: CloudDriveAccount): number {
  if (account.quota_total <= 0) return 0
  return Math.min(100, Math.round((account.quota_used / account.quota_total) * 100))
}

function formatQuota(used: number, total: number): string {
  const gb = 1024 ** 3
  const fmt = (v: number) => (v >= gb ? `${(v / gb).toFixed(1)}G` : `${Math.round(v / 1024 ** 2)}M`)
  return `${fmt(used)} / ${fmt(total)}`
}

async function fetchAll() {
  accountsLoading.value = true
  try {
    const [accountList, configs] = await Promise.all([cloudDriveApi.getAccounts(), cloudDriveApi.getConfigs()])
    accounts.value = accountList || []
    for (const cfg of configs || []) {
      if (cfg.provider === 'baidu' || cfg.provider === 'aliyun') {
        const form = oauthForms[cfg.provider]
        form.client_id = cfg.client_id || ''
        form.masked = cfg.client_secret === '******'
      }
    }
  } catch (e: any) {
    message.error(e?.message || t('cloudDrive.errors.loadFailed'))
  } finally {
    accountsLoading.value = false
  }
}

/**
 * 回调地址由当前访问地址自动生成（自部署单实例，管理后台所在域名即本实例外部地址），
 * 保存时上报 origin，后端用它拼接 OAuth redirect_uri；展示给用户登记到开放平台。
 */
function callbackURI(provider: OAuthProvider): string {
  return `${window.location.origin}/api/v1/clouddrive/callback/${provider}`
}

async function copyCallback(url: string) {
  try {
    await navigator.clipboard.writeText(url)
    message.success(t('cloudDrive.config.copied'))
  } catch {
    message.error(t('cloudDrive.errors.saveFailed'))
  }
}

async function handleSaveConfig(provider: OAuthProvider) {
  const form = oauthForms[provider]
  if (!form.client_id || (!form.client_secret && !form.masked)) {
    message.warning(t('cloudDrive.config.required'))
    return
  }
  oauthSaving.value = provider
  try {
    await cloudDriveApi.upsertConfig({
      provider,
      client_id: form.client_id,
      client_secret: form.client_secret, // 留空=不修改
      callback_base_url: window.location.origin,
    })
    message.success(t('cloudDrive.config.saved'))
    await fetchAll()
  } catch (e: any) {
    message.error(e?.message || t('cloudDrive.errors.saveFailed'))
  } finally {
    oauthSaving.value = ''
  }
}

function validateWebDAVForm(): boolean {
  if (!webdavForm.base_url || !webdavForm.username || !webdavForm.password) {
    message.warning(t('cloudDrive.config.required'))
    return false
  }
  return true
}

async function handleTestWebDAV() {
  if (!validateWebDAVForm()) return
  webdavTesting.value = true
  try {
    await cloudDriveApi.testWebDAV({ ...webdavForm })
    message.success(t('cloudDrive.webdav.testOk'))
  } catch (e: any) {
    message.error(e?.message || t('cloudDrive.errors.webdavTestFailed'))
  } finally {
    webdavTesting.value = false
  }
}

async function handleAddWebDAV() {
  if (!validateWebDAVForm()) return
  webdavSaving.value = true
  try {
    await cloudDriveApi.addWebDAVAccount({ ...webdavForm })
    message.success(t('cloudDrive.webdav.added'))
    webdavForm.password = ''
    await fetchAll()
  } catch (e: any) {
    message.error(e?.message || t('cloudDrive.errors.webdavAddFailed'))
  } finally {
    webdavSaving.value = false
  }
}

async function handleAuthorize(provider: OAuthProvider) {
  try {
    const res = await cloudDriveApi.getAuthorizeURL(provider)
    window.open(res.auth_url, '_blank', 'noopener,width=520,height=680')
    startOAuthPolling()
  } catch (e: any) {
    message.error(e?.message || t('cloudDrive.errors.authorizeFailed'))
  }
}

/** 授权弹窗打开期间轮询账号列表（后端回调直接落库），超时 90s 自动停止 */
function startOAuthPolling() {
  stopOAuthPolling()
  oauthPending.value = true
  oauthPollDeadline = Date.now() + 90_000
  oauthPollTimer = setInterval(async () => {
    if (Date.now() > oauthPollDeadline) {
      stopOAuthPolling()
      return
    }
    try {
      const list = await cloudDriveApi.getAccounts()
      const before = accounts.value.length
      accounts.value = list || []
      if (accounts.value.length > before) {
        stopOAuthPolling()
        message.success(t('cloudDrive.accounts.added'))
      }
    } catch {
      // 轮询失败继续
    }
  }, 2000)
}

function stopOAuthPolling() {
  if (oauthPollTimer) {
    clearInterval(oauthPollTimer)
    oauthPollTimer = null
  }
  oauthPending.value = false
}

/** 监听授权回调页 postMessage，立即刷新 */
function onOAuthMessage(event: MessageEvent) {
  const data = event.data as { source?: string; status?: string } | null
  if (data && data.source === 'novablog-clouddrive' && data.status === 'ok') {
    stopOAuthPolling()
    fetchAll()
    message.success(t('cloudDrive.accounts.added'))
  }
}

async function handleUnbind(account: CloudDriveAccount) {
  Modal.confirm({
    title: t('cloudDrive.accounts.unbindConfirm'),
    okText: t('cloudDrive.accounts.unbind'),
    okType: 'danger',
    onOk: async () => {
      try {
        await cloudDriveApi.removeAccount(account.id)
        message.success(t('cloudDrive.accounts.unbound'))
        await fetchAll()
      } catch (e: any) {
        message.error(e?.message || t('cloudDrive.errors.unbindFailed'))
      }
    },
  })
}

onMounted(fetchAll)
onUnmounted(() => {
  stopOAuthPolling()
  window.removeEventListener('message', onOAuthMessage)
})
</script>

<style scoped>
.provider-hint {
  margin-bottom: 16px;
}

.section-card {
  border: 1px solid #e5e9f2;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #29365c;
  margin: 0;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.apply-link {
  font-size: 13px;
  color: #526fe8;
  text-decoration: none;
}

.apply-link:hover {
  text-decoration: underline;
}

.accounts-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.bind-form {
  max-width: 560px;
}

.callback-input {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  color: #526fe8;
}

.callback-input :deep(input) {
  color: #526fe8;
  cursor: default;
}

.copy-btn {
  color: #526fe8;
}

.account-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 12px;
}

.account-card {
  border: 1px solid #e5e9f2;
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.account-head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.account-icon {
  color: #526fe8;
}

.account-name {
  font-weight: 500;
  color: #29365c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-status.expired {
  font-size: 12px;
  color: #d4380d;
  border: 1px solid #ffa39e;
  border-radius: 4px;
  padding: 0 4px;
}

.account-meta {
  font-size: 12px;
  color: #8a93a8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-quota {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #8a93a8;
}

.quota-label {
  flex-shrink: 0;
}

.quota-bar {
  flex: 1;
  height: 6px;
  background: #eef1f6;
  border-radius: 3px;
  overflow: hidden;
}

.quota-bar-inner {
  height: 100%;
  background: #526fe8;
  border-radius: 3px;
}

.quota-text {
  flex-shrink: 0;
}

.account-actions {
  display: flex;
  justify-content: flex-end;
}

.oauth-pending {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #b7791f;
  margin-bottom: 12px;
}
</style>
