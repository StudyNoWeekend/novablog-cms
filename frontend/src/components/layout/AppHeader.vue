<template>
  <header class="app-header">
    <div class="header-left">
      <span class="hamburger-btn" @click="appStore.toggleSidebar">
        <MenuOutlined />
      </span>
      <div class="greeting-wrapper">
        <h1 class="greeting">{{ $t('app.header.greeting', { name: nickname }) }}</h1>
        <div class="subtitle-row">
          <p class="subtitle">{{ roleDisplay }}</p>
          <a-button
            v-if="hasRole"
            type="link"
            size="small"
            class="role-edit-btn"
            @click="appStore.roleModalVisible = true"
          >
            {{ $t('app.header.editRole') }}
          </a-button>
        </div>
      </div>
    </div>
    <div class="header-right">
      <a
        class="action-icon github-link"
        href="https://github.com/StudyNoWeekend/novablog-cms"
        target="_blank"
        rel="noopener noreferrer"
        :title="$t('app.header.githubRepo')"
        :aria-label="$t('app.header.githubRepoAria')"
      >
        <GithubOutlined />
      </a>
      <a-badge dot color="#C2414B" class="action-icon-badge">
        <span class="action-icon">
          <BellOutlined />
        </span>
      </a-badge>
      <LocaleSwitch />
      <AppAvatar />
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { BellOutlined, MenuOutlined, GithubOutlined } from '@ant-design/icons-vue'
import AppAvatar from '@/components/common/AppAvatar.vue'
import LocaleSwitch from '@/components/common/LocaleSwitch.vue'
import { ALL_ROLES } from '@/constants/setupRoles'

const authStore = useAuthStore()
const appStore = useAppStore()
const { t } = useI18n()

const nickname = computed(() => {
  return authStore.user?.nickname || authStore.username || t('app.header.defaultNickname')
})

const hasRole = computed(() => !!authStore.user?.role)

// 已选创作方向的展示文案：role 为空时回退到欢迎语（失效 key 自动过滤）
const roleDisplay = computed(() => {
  const role = authStore.user?.role
  if (!role) return t('app.header.welcome')
  const labels = role
    .split(',')
    .map((key) => key.trim())
    .map((key) => ALL_ROLES.find((r) => r.key === key)?.labelKey)
    .filter((labelKey): labelKey is string => !!labelKey)
    .map((labelKey) => t(labelKey))
  return labels.length ? labels.join(' · ') : t('app.header.welcome')
})
</script>

<style scoped>
.app-header {
  height: var(--header-height);
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  border-bottom: 1px solid #E5E9F2;
  flex-shrink: 0;
  z-index: 10;
  position: relative;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.greeting-wrapper {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 2px;
}

.subtitle-row {
  display: flex;
  align-items: center;
  gap: 4px;
}

.role-edit-btn {
  height: auto;
  padding: 0;
  font-size: 12px;
}

.hamburger-btn {
  display: none;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #667085;
  cursor: pointer;
  padding: 4px;
}

.greeting {
  font-size: 24px;
  font-weight: 700;
  color: #29365C;
  margin: 0;
  line-height: 1.3;
}

.subtitle {
  font-size: 14px;
  color: #8A93A8;
  margin: 0;
  line-height: 1.4;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.header-right .locale-switch {
  margin-right: -4px;
}

.action-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #667085;
  cursor: pointer;
  transition: color var(--transition-fast);
  padding: 4px;
}

.action-icon:hover {
  color: #526FE8;
}

.github-link {
  text-decoration: none;
}

.action-icon-badge :deep(.ant-badge-dot) {
  width: 8px;
  height: 8px;
  min-width: 8px;
}

@media (max-width: 768px) {
  .app-header {
    padding: 0 16px;
  }

  .hamburger-btn {
    display: inline-flex;
  }

  .greeting {
    font-size: 20px;
  }

  .header-right {
    gap: 12px;
  }
}
</style>
