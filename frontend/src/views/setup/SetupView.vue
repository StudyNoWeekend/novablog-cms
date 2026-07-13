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
      <p class="setup-desc">创建您的博主账号以开始使用系统</p>

      <!-- 表单 -->
      <a-form
        ref="formRef"
        :model="form"
        layout="vertical"
        class="setup-form"
        autocomplete="off"
        @finish="handleInit"
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
            完成初始化
          </a-button>
        </a-form-item>
      </a-form>

      <!-- 分隔线 -->
      <a-divider class="setup-divider" />

      <!-- 底部标语 -->
      <p class="setup-slogan">用专业的方式，展示你的创作</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { UserOutlined, LockOutlined, SmileOutlined } from '@ant-design/icons-vue'
import { setupApi } from '@/api/setup'
import { storage } from '@/utils/storage'
import type { FormInstance } from 'ant-design-vue'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
  nickname: '',
})

const usernameRules = [
  { required: true, message: '请输入用户名', trigger: 'blur' },
  { min: 3, max: 50, message: '用户名长度为 3-50 个字符', trigger: 'blur' },
]

const passwordRules = [
  { required: true, message: '请输入密码', trigger: 'blur' },
  { min: 6, message: '密码长度至少为 6 个字符', trigger: 'blur' },
]

async function handleInit() {
  loading.value = true
  try {
    const res = await setupApi.init({
      username: form.username,
      password: form.password,
      nickname: form.nickname || undefined,
    })
    if (res.success) {
      storage.setInitialized(true)
      message.success('初始化成功！请使用新账号登录')
      router.push('/auth/login')
    } else {
      message.error(res.message || '初始化失败')
    }
  } catch (error: any) {
    message.error(error?.message || '初始化失败，请重试')
  } finally {
    loading.value = false
  }
}
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

/* 底部标语 */
.setup-slogan {
  text-align: center;
  font-size: 14px;
  color: var(--text-tertiary, #94a3b8);
  margin: 0;
  letter-spacing: 1px;
  user-select: none;
}
</style>