<template>
  <div class="login-page">
    <!-- 几何网格背景 -->
    <div class="bg-pattern"></div>

    <div class="login-card">
      <!-- Logo -->
      <div class="login-logo">Novablog</div>

      <!-- 标题 -->
      <h1 class="login-title">创作者后台</h1>

      <!-- 表单 -->
      <a-form
        ref="formRef"
        :model="form"
        layout="vertical"
        class="login-form"
        autocomplete="off"
        @finish="handleLogin"
      >
        <!-- 用户名 -->
        <a-form-item name="username" :rules="usernameRules">
          <a-input
            v-model:value="form.username"
            placeholder="用户名/邮箱"
            size="large"
            class="login-input"
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
            placeholder="密码"
            size="large"
            class="login-input"
          >
            <template #prefix>
              <lock-outlined />
            </template>
          </a-input-password>
        </a-form-item>

        <!-- 记住我 + 忘记密码 -->
        <div class="form-extra">
          <a-checkbox v-model:checked="rememberMe">记住我</a-checkbox>
          <a class="forgot-link" @click="handleForgotPassword">忘记密码</a>
        </div>

        <!-- 登录按钮 -->
        <a-form-item class="submit-item">
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            block
            :loading="loading"
            :class="{ 'btn-loading': loading }"
          >
            登 录
          </a-button>
        </a-form-item>
      </a-form>

      <!-- 分隔线 -->
      <a-divider class="login-divider" />

      <!-- 底部标语 -->
      <p class="login-slogan">用专业的方式，展示你的创作</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'
import type { FormInstance } from 'ant-design-vue'

const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const rememberMe = ref(true)

const form = reactive({
  username: '',
  password: '',
})

const usernameRules = [
  { required: true, message: '请输入用户名/邮箱', trigger: 'blur' },
]

const passwordRules = [
  { required: true, message: '请输入密码', trigger: 'blur' },
]

async function handleLogin() {
  loading.value = true
  try {
    await authStore.login(form, rememberMe.value)
  } catch {
    message.error('用户名或密码错误')
  } finally {
    loading.value = false
  }
}

function handleForgotPassword() {
  message.info('请联系管理员重置密码')
}
</script>

<style scoped>
.login-page {
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

/* 登录卡片 */
.login-card {
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
.login-logo {
  text-align: center;
  font-size: 36px;
  font-weight: 800;
  letter-spacing: 2px;
  color: var(--color-primary, #4a6cf7);
  margin-bottom: 4px;
  user-select: none;
}

/* 标题 */
.login-title {
  text-align: center;
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0 0 36px;
  letter-spacing: 1px;
}

/* 移动端适配 */
@media (max-width: 480px) {
  .login-card {
    padding: 32px 24px 24px;
  }

  .login-logo {
    font-size: 28px;
  }

  .login-title {
    font-size: 18px;
    margin-bottom: 28px;
  }
}

/* 表单 */
.login-form {
  width: 100%;
}

/* 输入框增强 */
.login-input :deep(.ant-input),
.login-input :deep(.ant-input-affix-wrapper) {
  border-color: var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
  transition: border-color var(--transition-fast, 150ms), box-shadow var(--transition-fast, 150ms);
}

.login-input :deep(.ant-input-affix-wrapper):hover,
.login-input :deep(.ant-input):hover {
  border-color: var(--color-primary, #4a6cf7);
}

.login-input :deep(.ant-input-affix-wrapper):focus,
.login-input :deep(.ant-input-affix-wrapper)-focused,
.login-input :deep(.ant-input):focus {
  border-color: var(--color-primary, #4a6cf7);
  box-shadow: 0 0 0 2px rgba(74, 108, 247, 0.15);
}

.login-input :deep(.ant-input-prefix) {
  margin-right: 10px;
  color: var(--text-tertiary, #94a3b8);
}

/* 记住我 + 忘记密码 */
.form-extra {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: -8px;
  margin-bottom: 16px;
}

.form-extra :deep(.ant-checkbox-wrapper) {
  color: var(--text-secondary, #64748b);
  font-size: 14px;
}

.forgot-link {
  font-size: 14px;
  color: var(--color-primary, #4a6cf7);
  cursor: pointer;
  transition: color var(--transition-fast, 150ms);
}

.forgot-link:hover {
  color: var(--color-primary-hover, #3b5de7);
}

/* 提交按钮 */
.submit-item {
  margin-bottom: 0;
}

.submit-item :deep(.ant-btn) {
  height: 44px;
  font-size: 16px;
  font-weight: 500;
  letter-spacing: 4px;
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
.login-divider {
  margin: 32px 0 24px;
}

.login-divider :deep(.ant-divider-inner-text) {
  color: var(--text-tertiary, #94a3b8);
}

/* 底部标语 */
.login-slogan {
  text-align: center;
  font-size: 14px;
  color: var(--text-tertiary, #94a3b8);
  margin: 0;
  letter-spacing: 1px;
  user-select: none;
}
</style>