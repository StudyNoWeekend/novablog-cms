<template>
  <div class="setup-page">
    <!-- 几何网格背景 -->
    <div class="bg-pattern"></div>

    <div class="setup-card">
      <!-- Logo -->
      <div class="setup-logo">Novablog</div>

      <!-- 标题 -->
      <h1 class="setup-title">Novablog</h1>

      <!-- 欢迎文字 -->
      <p class="setup-welcome">欢迎使用 Novablog</p>
      <p class="setup-desc">{{ phaseDescriptions[phase] }}</p>

      <!-- 步骤指示条 -->
      <a-steps :current="stepCurrent" size="small" class="setup-steps" :items="stepItems" />

      <!-- 步骤一：创建博主账号 -->
      <a-form
        v-if="phase === 'form'"
        ref="formRef"
        :model="form"
        layout="vertical"
        class="setup-form"
        autocomplete="off"
        @finish="goRoleStep"
      >
        <!-- 用户名 -->
        <a-form-item name="username" :rules="usernameRules">
          <a-input
            v-model:value="form.username"
            placeholder="设置登录用户名"
            size="large"
            class="setup-input"
          >
            <template #prefix>
              <user-outlined />
            </template>
          </a-input>
        </a-form-item>

        <!-- 密码 -->
        <a-form-item name="password" :rules="passwordRules">
          <a-input-password
            v-model:value="form.password"
            placeholder="设置登录密码"
            size="large"
            class="setup-input"
          >
            <template #prefix>
              <lock-outlined />
            </template>
          </a-input-password>
        </a-form-item>

        <!-- 确认密码 -->
        <a-form-item name="confirmPassword" :rules="confirmPasswordRules">
          <a-input-password
            v-model:value="form.confirmPassword"
            placeholder="再次输入密码以确认"
            size="large"
            class="setup-input"
          >
            <template #prefix>
              <lock-outlined />
            </template>
          </a-input-password>
        </a-form-item>

        <!-- 昵称 -->
        <a-form-item name="nickname">
          <a-input
            v-model:value="form.nickname"
            placeholder="您的昵称"
            size="large"
            class="setup-input"
          >
            <template #prefix>
              <smile-outlined />
            </template>
          </a-input>
        </a-form-item>

        <!-- 提交按钮 -->
        <a-form-item class="submit-item">
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            block
            :loading="loading"
            :class="{ 'btn-loading': loading }"
          >
            下一步：选择创作方向
          </a-button>
        </a-form-item>
      </a-form>

      <!-- 分隔线 -->
      <a-divider v-if="phase === 'form'" class="setup-divider" />

      <!-- 步骤二：选择创作方向（可多选，随创建账号一起提交） -->
      <div v-if="phase === 'role'" class="role-step">
        <h2 class="role-step__title">你想打造怎样的博客？</h2>
        <p class="role-step__desc">选择你的创作方向（可多选），系统将按方向预置后台功能模块，之后可在「模块管理」中随时调整。</p>

        <RolePickerCards v-model:selected="selectedRoles" class="role-step__picker" />

        <div class="role-preset">
          <div class="role-preset__header">
            <span class="role-preset__title">将为你开启的模块</span>
            <a-button type="link" size="small" @click="showCustomize = !showCustomize">
              {{ showCustomize ? '收起微调' : '自定义微调' }}
            </a-button>
          </div>
          <div class="role-preset__chips">
            <span
              v-for="item in moduleChips"
              :key="item.key"
              class="module-chip"
              :class="{ 'module-chip--off': !item.enabled }"
            >
              <CheckOutlined v-if="item.enabled" class="module-chip__check" />
              {{ item.label }}
              <em v-if="item.common" class="module-chip__badge">通用</em>
            </span>
          </div>
          <div v-if="showCustomize" class="role-preset__customize">
            <div v-for="item in customizableModules" :key="item.key" class="customize-row">
              <span class="customize-row__label">{{ item.label }}</span>
              <a-switch
                v-model:checked="moduleOverrides[item.key]"
                size="small"
                :aria-label="`开关 ${item.label} 模块`"
              />
            </div>
          </div>
        </div>

        <div class="role-step__actions">
          <a-space direction="vertical" style="width: 100%">
            <a-button
              type="primary"
              size="large"
              block
              :loading="loading"
              :class="{ 'btn-loading': loading }"
              @click="handleInit"
            >
              创建账号
            </a-button>
            <a-button size="large" block :disabled="loading" @click="skipRole">
              跳过，稍后再选
            </a-button>
          </a-space>
        </div>
      </div>

      <a-divider v-if="phase === 'role'" class="setup-divider" />

      <!-- 步骤二：配置对象存储 -->
      <div v-if="phase === 'storage'" class="storage-step">
        <div class="storage-step__icon"><CloudServerOutlined /></div>
        <h2 class="storage-step__title">配置对象存储</h2>
        <p class="storage-step__desc">选择文件存储方式，用于保存上传的图片、视频等素材。可跳过，后续在管理后台中随时配置。</p>

        <a-form ref="storageFormRef" layout="vertical" class="setup-form" autocomplete="off">
          <!-- 存储提供商 -->
          <a-form-item label="存储类型" name="provider">
            <a-select
              v-model:value="storageForm.provider"
              placeholder="选择存储类型"
              size="large"
              class="setup-input"
              @change="onProviderChange"
            >
              <a-select-option value="local">本地存储</a-select-option>
              <a-select-option value="aliyun">阿里云 OSS</a-select-option>
              <a-select-option value="tencent">腾讯云 COS</a-select-option>
              <a-select-option value="minio">MinIO</a-select-option>
            </a-select>
          </a-form-item>

          <!-- 云存储参数（选择云存储后展开） -->
          <template v-if="storageForm.provider && storageForm.provider !== 'local'">
            <a-form-item label="Endpoint" name="endpoint" :rules="cloudRequiredRules">
              <a-input v-model:value="storageForm.endpoint" placeholder="https://oss-cn-hangzhou.aliyuncs.com" size="large" class="setup-input" />
            </a-form-item>

            <a-form-item label="Region" name="region">
              <a-input v-model:value="storageForm.region" placeholder="oss-cn-hangzhou" size="large" class="setup-input" />
            </a-form-item>

            <a-form-item label="Bucket" name="bucket" :rules="cloudRequiredRules">
              <a-input v-model:value="storageForm.bucket" placeholder="my-blog" size="large" class="setup-input" />
            </a-form-item>

            <a-form-item label="AccessKey" name="access_key" :rules="cloudRequiredRules">
              <a-input v-model:value="storageForm.access_key" placeholder="LTAI******" size="large" class="setup-input" />
            </a-form-item>

            <a-form-item label="AccessSecret" name="access_secret" :rules="cloudRequiredRules">
              <a-input-password v-model:value="storageForm.access_secret" placeholder="********" size="large" class="setup-input" />
            </a-form-item>

            <a-form-item v-if="storageForm.provider === 'minio'" label="启用 SSL" name="use_ssl">
              <a-switch v-model:checked="storageForm.use_ssl" />
            </a-form-item>

            <a-form-item label="PathPrefix（可选）" name="path_prefix">
              <a-input v-model:value="storageForm.path_prefix" placeholder="blog/images" size="large" class="setup-input" />
            </a-form-item>

            <a-form-item label="CustomDomain（可选）" name="custom_domain">
              <a-input v-model:value="storageForm.custom_domain" placeholder="https://cdn.yourdomain.com" size="large" class="setup-input" />
            </a-form-item>
          </template>
        </a-form>

        <div class="storage-step__actions">
          <a-space direction="vertical" style="width: 100%">
            <a-button
              type="primary"
              size="large"
              block
              :loading="storageSaving"
              @click="handleSaveStorage"
            >
              <template #icon><SaveOutlined /></template>
              保存并继续
            </a-button>
            <a-button
              v-if="storageForm.provider && storageForm.provider !== 'local'"
              size="large"
              block
              :loading="storageTesting"
              @click="handleTestConnection"
            >
              <template #icon><ApiOutlined /></template>
              测试连接
            </a-button>
            <a-button size="large" block @click="skipStorage">
              跳过，使用本地存储
            </a-button>
          </a-space>
        </div>
      </div>

      <!-- 分隔线（storage → theme） -->
      <a-divider v-if="phase === 'storage'" class="setup-divider" />

      <!-- 步骤三：初始化博客外观（登录官方账号并拉取官方默认主题） -->
      <div v-if="phase === 'theme'" class="theme-step">
        <div class="theme-step__icon"><SkinOutlined /></div>
        <h2 class="theme-step__title">初始化博客外观</h2>
        <p class="theme-step__desc">输入 NovaBlog 官方主题市场地址并登录官方账号，将自动获取并启用默认主题</p>

        <a-form ref="marketFormRef" layout="vertical" class="setup-form" autocomplete="off">
          <a-form-item label="官方市场地址" name="marketBaseURL" :rules="marketURLRules">
            <a-input v-model:value="marketBaseURL" :placeholder="defaultMarketURL || '请输入官方市场地址'" allow-clear size="large">
              <template #prefix><LinkOutlined /></template>
            </a-input>
          </a-form-item>
          <a-form-item label="官方账号" name="marketEmail" :rules="marketAccountRules">
            <a-input v-model:value="marketAccount.email" placeholder="官方市场登录邮箱" allow-clear size="large">
              <template #prefix><UserOutlined /></template>
            </a-input>
          </a-form-item>
          <a-form-item label="官方密码" name="marketPassword" :rules="marketAccountRules">
            <a-input-password v-model:value="marketAccount.password" placeholder="官方市场登录密码" size="large">
              <template #prefix><LockOutlined /></template>
            </a-input-password>
          </a-form-item>
        </a-form>

        <div class="theme-step__progress" v-if="themeStatus">
          <div class="theme-step__stage" :class="{ 'theme-step__stage--active': themeStatus.status === 'running' }">
            <template v-if="themeStatus.status === 'running'">
              <LoadingOutlined /> {{ stageText }}
            </template>
            <template v-else-if="themeStatus.status === 'success'">
              <CheckCircleOutlined class="theme-step__ok" />
              已启用 {{ themeStatus.theme_name }} v{{ themeStatus.theme_version }}
            </template>
            <template v-else-if="themeStatus.status === 'failed'">
              <CloseCircleOutlined class="theme-step__err" /> {{ themeStatus.message }}
            </template>
          </div>
        </div>

        <div class="theme-step__actions">
          <a-button
            v-if="!themeStatus || themeStatus.status !== 'success'"
            type="primary"
            size="large"
            block
            :loading="themeStarting"
            @click="startThemeInit"
          >
            {{ themeStatus?.status === 'failed' ? '重试' : '获取默认主题' }}
          </a-button>
          <a-button
            v-if="themeStatus?.status === 'success'"
            type="primary"
            size="large"
            block
            @click="goLogin"
          >
            完成，进入登录
          </a-button>
          <a-button size="large" block :disabled="themeStatus?.status === 'running'" @click="goLogin">
            跳过，稍后在主题页安装
          </a-button>
        </div>
      </div>

      <!-- 底部标语 -->
      <p class="setup-slogan">用专业的方式，展示你的创作</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  ApiOutlined,
  CheckCircleOutlined,
  CheckOutlined,
  CloseCircleOutlined,
  CloudServerOutlined,
  LinkOutlined,
  LoadingOutlined,
  LockOutlined,
  SaveOutlined,
  SkinOutlined,
  UserOutlined,
} from '@ant-design/icons-vue'
import { setupApi } from '@/api/setup'
import { configApi } from '@/api/config'
import { storageApi } from '@/api/storage'
import { marketStorage } from '@/utils/storage'
import { storage } from '@/utils/storage'
import type { ThemeInstallStatus } from '@/types/theme'
import type { FormInstance } from 'ant-design-vue'
import { MODULE_DEFS, COMMON_MODULE_KEYS, type ModuleKey } from '@/constants/modules'
import { computeModulePreset, type RoleKey } from '@/constants/setupRoles'
import RolePickerCards from '@/components/common/RolePickerCards.vue'

const router = useRouter()
const formRef = ref<FormInstance>()
const marketFormRef = ref<FormInstance>()
const storageFormRef = ref<FormInstance>()
const loading = ref(false)

// 首装阶段：form 创建账号 → role 选择创作方向 → storage 对象存储 → theme 初始化博客外观
const phase = ref<'form' | 'role' | 'storage' | 'theme'>('form')
const themeStarting = ref(false)
const themeStatus = ref<ThemeInstallStatus | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null

const stepItems = [
  { title: '账号' },
  { title: '方向' },
  { title: '存储' },
  { title: '外观' },
]
const stepCurrent = computed(() => ({ form: 0, role: 1, storage: 2, theme: 3 })[phase.value])
const phaseDescriptions: Record<string, string> = {
  form: '创建您的博主账号以开始使用系统',
  role: '选择创作方向，系统将预置对应的功能模块',
  storage: '选择文件存储方式，可跳过稍后配置',
  theme: '获取官方默认主题，一键完成博客外观初始化',
}

// ---- 创作方向选择 ----
const selectedRoles = ref<RoleKey[]>([])
const showCustomize = ref(false)
const moduleOverrides = reactive<Record<string, boolean>>({})

// 角色变化时清空手动微调，重新应用预设
watch(selectedRoles, () => {
  Object.keys(moduleOverrides).forEach((key) => delete moduleOverrides[key])
})

/** 最终模块开关 = 角色并集预设 + 手动微调 */
const finalModules = computed<Record<ModuleKey, boolean>>(() => {
  const preset = computeModulePreset(selectedRoles.value)
  for (const [key, value] of Object.entries(moduleOverrides)) {
    if (key in preset) {
      preset[key as ModuleKey] = value
    }
  }
  return preset
})

const moduleChips = computed(() =>
  MODULE_DEFS.map((def) => ({
    key: def.key,
    label: def.label,
    common: COMMON_MODULE_KEYS.includes(def.key),
    enabled: finalModules.value[def.key],
  })),
)

const customizableModules = computed(() =>
  MODULE_DEFS.filter((def) => !COMMON_MODULE_KEYS.includes(def.key)),
)

// 存储配置表单
const storageForm = reactive({
  provider: 'local',
  endpoint: '',
  region: '',
  bucket: '',
  access_key: '',
  access_secret: '',
  path_prefix: '',
  custom_domain: '',
  extra: '',
  use_ssl: false,
})
const storageSaving = ref(false)
const storageTesting = ref(false)

// 云存储必填校验规则
const cloudRequiredRules = [{ required: true, message: '此项为必填', trigger: 'blur' }]

// 官方市场地址（向导输入）；默认值由后端下发，本地缓存（上次向导输入）优先
const marketBaseURL = ref(marketStorage.getBaseURL() || '')
const defaultMarketURL = ref('')

// 官方市场账号（必填）：官方代理下载要求登录态，由后端临时换取 Token，不落库
const marketAccount = reactive({ email: '', password: '' })
const marketAccountRules = [{ required: true, message: '官方市场下载需要先登录官方账号', trigger: 'blur' }]

onMounted(async () => {
  try {
    const config = await configApi.getPublicConfig()
    defaultMarketURL.value = config.market_base_url || ''
    if (!marketBaseURL.value && defaultMarketURL.value) {
      marketBaseURL.value = defaultMarketURL.value
    }
  } catch {
    // 下发失败保持为空，允许手动输入
  }
})

const marketURLRules = [
  { pattern: /^https?:\/\//, message: '官方地址需以 http:// 或 https:// 开头', trigger: 'blur' },
]

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  nickname: '',
})

const stageText = computed(() => {
  switch (themeStatus.value?.stage) {
    case 'fetching':
      return '正在从官方市场获取默认主题…'
    case 'installing':
      return '正在下载并安装主题制品…'
    case 'activating':
      return '正在启用主题…'
    default:
      return '处理中…'
  }
})

const usernameRules = [
  { required: true, message: '请输入用户名', trigger: 'blur' },
  { min: 3, max: 50, message: '用户名长度为 3-50 个字符', trigger: 'blur' },
]

const passwordRules = [
  { required: true, message: '请输入密码', trigger: 'blur' },
  { min: 6, message: '密码长度至少为 6 个字符', trigger: 'blur' },
]

const confirmPasswordRules = [
  { required: true, message: '请再次输入密码', trigger: 'blur' },
  {
    validator: (_rule: any, value: string) => {
      if (value && value !== form.password) {
        return Promise.reject(new Error('两次输入的密码不一致'))
      }
      return Promise.resolve()
    },
    trigger: 'blur',
  },
]

/** 账号步本地校验通过后进入创作方向选择步 */
async function goRoleStep() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  phase.value = 'role'
}

/** 创建账号：携带所选创作方向与最终模块开关一起提交 */
async function handleInit() {
  loading.value = true
  try {
    const payload: Parameters<typeof setupApi.init>[0] = {
      username: form.username,
      password: form.password,
      nickname: form.nickname || undefined,
    }
    if (selectedRoles.value.length > 0) {
      payload.role = selectedRoles.value.join(',')
      payload.modules = { ...finalModules.value }
    }
    const res = await setupApi.init(payload)
    if (res.success) {
      storage.setInitialized(true)
      message.success('博主账号创建成功')
      phase.value = 'storage'
    } else {
      message.error(res.message || '初始化失败')
      phase.value = 'form'
    }
  } catch (error: any) {
    message.error(error?.message || '初始化失败，请重试')
    phase.value = 'form'
  } finally {
    loading.value = false
  }
}

/** 跳过创作方向选择：不带角色直接初始化（模块保持默认全开） */
async function skipRole() {
  selectedRoles.value = []
  await handleInit()
}

/** 存储提供商切换时重置云存储参数 */
function onProviderChange() {
  if (storageForm.provider === 'local') {
    storageForm.endpoint = ''
    storageForm.region = ''
    storageForm.bucket = ''
    storageForm.access_key = ''
    storageForm.access_secret = ''
    storageForm.path_prefix = ''
    storageForm.custom_domain = ''
    storageForm.use_ssl = false
  }
}

/** 保存存储配置并进入主题步骤 */
async function handleSaveStorage() {
  // 如果是云存储，先校验必填项
  if (storageForm.provider && storageForm.provider !== 'local') {
    const missing: string[] = []
    if (!storageForm.endpoint) missing.push('Endpoint')
    if (!storageForm.bucket) missing.push('Bucket')
    if (!storageForm.access_key) missing.push('AccessKey')
    if (!storageForm.access_secret) missing.push('AccessSecret')
    if (missing.length > 0) {
      message.warning(`请填写以下必填项：${missing.join('、')}`)
      return
    }
  }

  storageSaving.value = true
  try {
    // 构建 extra JSON（minio use_ssl）
    let extra = storageForm.extra
    if (storageForm.provider === 'minio' && storageForm.use_ssl) {
      extra = JSON.stringify({ use_ssl: true })
    }

    const res = await setupApi.setupStorage({
      provider: storageForm.provider,
      endpoint: storageForm.endpoint || undefined,
      region: storageForm.region || undefined,
      bucket: storageForm.bucket || undefined,
      access_key: storageForm.access_key || undefined,
      access_secret: storageForm.access_secret || undefined,
      path_prefix: storageForm.path_prefix || undefined,
      custom_domain: storageForm.custom_domain || undefined,
      extra: extra || undefined,
    })
    if (res.success) {
      message.success(res.message || '存储配置已保存')
      phase.value = 'theme'
    } else {
      message.error(res.message || '保存存储配置失败')
    }
  } catch (error: any) {
    message.error(error?.message || '保存存储配置失败，请重试')
  } finally {
    storageSaving.value = false
  }
}

/** 测试云存储连接 */
async function handleTestConnection() {
  if (!storageForm.provider || storageForm.provider === 'local') {
    message.info('本地存储无需测试连接')
    return
  }

  const missing: string[] = []
  if (!storageForm.endpoint) missing.push('Endpoint')
  if (!storageForm.bucket) missing.push('Bucket')
  if (!storageForm.access_key) missing.push('AccessKey')
  if (!storageForm.access_secret) missing.push('AccessSecret')
  if (missing.length > 0) {
    message.warning(`请先填写以下必填项：${missing.join('、')}`)
    return
  }

  storageTesting.value = true
  try {
    let extra = storageForm.extra
    if (storageForm.provider === 'minio' && storageForm.use_ssl) {
      extra = JSON.stringify({ use_ssl: true })
    }
    await storageApi.testConfig({
      provider: storageForm.provider as 'aliyun' | 'tencent' | 'minio' | 'local',
      endpoint: storageForm.endpoint,
      region: storageForm.region,
      bucket: storageForm.bucket,
      access_key: storageForm.access_key,
      access_secret: storageForm.access_secret,
      path_prefix: storageForm.path_prefix,
      custom_domain: storageForm.custom_domain,
      extra: extra,
    })
    message.success('连接测试通过')
  } catch (error: any) {
    message.error(error?.message || '连接测试失败，请检查配置')
  } finally {
    storageTesting.value = false
  }
}

/** 跳过存储配置，使用本地存储 */
async function skipStorage() {
  // 跳过后自动调用 API 保存本地存储配置（确保流程完整）
  storageSaving.value = true
  try {
    await setupApi.setupStorage({ provider: 'local' })
  } catch {
    // 静默处理，不阻断跳转
  } finally {
    storageSaving.value = false
  }
  phase.value = 'theme'
}

/** 触发官方默认主题拉取并轮询任务状态（先校验官方地址与账号填写完整） */
async function startThemeInit() {
  try {
    await marketFormRef.value?.validate()
  } catch {
    return
  }
  themeStarting.value = true
  try {
    themeStatus.value = await setupApi.initTheme(marketBaseURL.value || undefined, {
      email: marketAccount.email.trim(),
      password: marketAccount.password,
    })
    if (themeStatus.value.status === 'running' && !pollTimer) {
      pollTimer = setInterval(pollThemeStatus, 1500)
    }
  } catch (error: any) {
    message.error(error?.message || '启动主题安装失败')
  } finally {
    themeStarting.value = false
  }
}

async function pollThemeStatus() {
  try {
    const status = await setupApi.getThemeStatus()
    themeStatus.value = status
    if (status.status !== 'running') {
      stopPolling()
      if (status.status === 'success') {
        // 持久化官方市场地址，主题页登录弹窗自动填入
        if (marketBaseURL.value) {
          marketStorage.setBaseURL(marketBaseURL.value)
        }
        message.success(`主题「${status.theme_name}」已启用，博客即将上线`)
      }
    }
  } catch {
    // 轮询失败静默重试
  }
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function goLogin() {
  stopPolling()
  router.push('/auth/login')
}

onUnmounted(stopPolling)
</script>

<style scoped>
.setup-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  overflow: hidden;
}

/* 几何网格覆盖层 */
.bg-pattern {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
  background-size: 48px 48px;
  pointer-events: none;
}

/* 角落装饰光晕 */
.bg-pattern::before {
  content: '';
  position: absolute;
  top: -120px;
  right: -120px;
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(74, 108, 247, 0.12) 0%, transparent 70%);
  border-radius: 50%;
}

.bg-pattern::after {
  content: '';
  position: absolute;
  bottom: -100px;
  left: -100px;
  width: 350px;
  height: 350px;
  background: radial-gradient(circle, rgba(124, 58, 237, 0.1) 0%, transparent 70%);
  border-radius: 50%;
}

/* 设置卡片 */
.setup-card {
  position: relative;
  z-index: 1;
  max-width: 420px;
  width: 90%;
  padding: 48px 40px 40px;
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  box-shadow:
    0 4px 24px rgba(0, 0, 0, 0.25),
    0 0 0 1px rgba(255, 255, 255, 0.05);
}

/* Logo */
.setup-logo {
  text-align: center;
  font-size: 36px;
  font-weight: 800;
  letter-spacing: 2px;
  color: var(--color-primary, #4a6cf7);
  margin-bottom: 4px;
  user-select: none;
}

/* 标题 */
.setup-title {
  text-align: center;
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0 0 8px;
  letter-spacing: 1px;
}

/* 欢迎文字 */
.setup-welcome {
  text-align: center;
  font-size: 16px;
  color: var(--text-secondary, #64748b);
  margin: 0 0 4px;
}

.setup-desc {
  text-align: center;
  font-size: 14px;
  color: var(--text-tertiary, #94a3b8);
  margin: 0 0 28px;
}

/* 移动端适配 */
@media (max-width: 480px) {
  .setup-card {
    padding: 32px 24px 24px;
  }

  .setup-logo {
    font-size: 28px;
  }

  .setup-title {
    font-size: 18px;
  }

  .setup-welcome {
    font-size: 14px;
  }

  .setup-desc {
    font-size: 13px;
    margin-bottom: 20px;
  }
}

/* 表单 */
.setup-form {
  width: 100%;
}

/* 输入框增强 */
.setup-input :deep(.ant-input),
.setup-input :deep(.ant-input-affix-wrapper) {
  border-color: var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
  transition: border-color var(--transition-fast, 150ms), box-shadow var(--transition-fast, 150ms);
}

.setup-input :deep(.ant-input-affix-wrapper):hover,
.setup-input :deep(.ant-input):hover {
  border-color: var(--color-primary, #4a6cf7);
}

.setup-input :deep(.ant-input-affix-wrapper):focus,
.setup-input :deep(.ant-input-affix-wrapper)-focused,
.setup-input :deep(.ant-input):focus {
  border-color: var(--color-primary, #4a6cf7);
  box-shadow: 0 0 0 2px rgba(74, 108, 247, 0.15);
}

.setup-input :deep(.ant-input-prefix) {
  margin-right: 10px;
  color: var(--text-tertiary, #94a3b8);
}

/* 提交按钮 */
.submit-item {
  margin-bottom: 0;
  margin-top: 8px;
}

.submit-item :deep(.ant-btn) {
  height: 44px;
  font-size: 16px;
  font-weight: 500;
  letter-spacing: 2px;
  border-radius: var(--border-radius, 8px);
  background: var(--color-primary, #4a6cf7);
  border-color: var(--color-primary, #4a6cf7);
  transition: all var(--transition-fast, 150ms);
}

.submit-item :deep(.ant-btn):hover {
  background: var(--color-primary-hover, #3b5de7);
  border-color: var(--color-primary-hover, #3b5de7);
}

/* 加载时的脉冲动画 */
.btn-loading {
  animation: btn-pulse 1.8s ease-in-out infinite;
}

@keyframes btn-pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(74, 108, 247, 0.4);
  }
  50% {
    box-shadow: 0 0 0 12px rgba(74, 108, 247, 0);
  }
}

/* 分隔线 */
.setup-divider {
  margin: 32px 0 24px;
}

.setup-divider :deep(.ant-divider-inner-text) {
  color: var(--text-tertiary, #94a3b8);
}

/* ===== 第 2 步：配置对象存储 ===== */
.storage-step {
  text-align: left;
}

.storage-step__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  margin-bottom: 12px;
  border-radius: 16px;
  background: linear-gradient(135deg, #f59e0b 0%, #ef4444 100%);
  color: #fff;
  font-size: 26px;
}

.storage-step__title {
  text-align: center;
  margin: 0 0 6px;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
}

.storage-step__desc {
  text-align: center;
  margin: 0 0 20px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary, #64748b);
}

.storage-step__actions {
  margin-top: 16px;
}

/* 存储步骤表单标签 */
.storage-step :deep(.ant-form-item-label > label) {
  font-size: 13px;
  color: var(--text-secondary, #64748b);
}

/* ===== 第 3 步：初始化博客外观 ===== */
.theme-step {
  text-align: center;
}

.theme-step__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  margin-bottom: 12px;
  border-radius: 16px;
  background: linear-gradient(135deg, #4a6cf7 0%, #7c3aed 100%);
  color: #fff;
  font-size: 26px;
}

.theme-step__title {
  margin: 0 0 6px;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
}

.theme-step__desc {
  margin: 0 0 20px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary, #64748b);
}

.theme-step__progress {
  padding: 16px;
  margin-bottom: 20px;
  border: 1px dashed var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
  background: var(--bg-layout, rgba(74, 108, 247, 0.04));
}

.theme-step__stage {
  font-size: 14px;
  color: var(--text-secondary, #64748b);
  word-break: break-all;
}

.theme-step__stage--active {
  color: var(--color-primary, #4a6cf7);
}

.theme-step__ok {
  color: #52c41a;
  margin-right: 6px;
}

.theme-step__err {
  color: #ff4d4f;
  margin-right: 6px;
}

.theme-step__actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/* 底部标语 */
.setup-slogan {
  text-align: center;
  font-size: 14px;
  color: var(--text-tertiary, #94a3b8);
  margin: 0;
  letter-spacing: 1px;
  user-select: none;
}

/* 步骤指示条 */
.setup-steps {
  margin-bottom: 20px;
}

.setup-steps :deep(.ant-steps-item-title) {
  font-size: 13px !important;
}

/* 创作方向选择步 */
.role-step__title {
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
  text-align: center;
  margin: 0 0 6px;
}

.role-step__desc {
  font-size: 13px;
  color: #64748b;
  text-align: center;
  margin: 0 0 16px;
  line-height: 1.6;
}

/* 角色卡片滚动容器（卡片网格与卡片样式在 RolePickerCards 组件内） */
.role-step__picker {
  max-height: 320px;
  overflow-y: auto;
  padding-right: 4px;
}

.role-preset {
  margin-top: 14px;
  padding: 12px 14px;
  background: #f8f9fb;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.role-preset__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.role-preset__title {
  font-size: 13px;
  font-weight: 500;
  color: #1e293b;
}

.role-preset__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.module-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  font-size: 12px;
  border-radius: 999px;
  background: rgba(74, 108, 247, 0.08);
  color: #4a6cf7;
  transition: opacity 0.2s;
}

.module-chip--off {
  background: #eef1f5;
  color: #94a3b8;
}

.module-chip__check {
  font-size: 10px;
}

.module-chip__badge {
  font-style: normal;
  font-size: 10px;
  color: #94a3b8;
}

.role-preset__customize {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px dashed #e2e8f0;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 6px 16px;
}

.customize-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 28px;
}

.customize-row__label {
  font-size: 13px;
  color: #475569;
}

.role-step__actions {
  margin-top: 16px;
}
</style>