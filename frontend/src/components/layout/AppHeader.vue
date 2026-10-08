<template>
  <header class="app-header">
    <div class="header-left">
      <span class="hamburger-btn" @click="appStore.toggleSidebar">
        <MenuOutlined />
      </span>
      <div class="greeting-wrapper">
        <h1 class="greeting">Hi, {{ nickname }}</h1>
        <div class="subtitle-row">
          <p class="subtitle">{{ roleDisplay }}</p>
          <a-button
            v-if="hasRole"
            type="link"
            size="small"
            class="role-edit-btn"
            @click="appStore.roleModalVisible = true"
          >
            修改
          </a-button>
        </div>
      </div>
    </div>
    <div class="header-right">
      <div class="search-box">
        <SearchOutlined class="search-icon" />
        <input type="text" placeholder="输入内容查询" class="search-input" />
      </div>
      <span class="action-icon">
        <MailOutlined />
      </span>
      <a-badge dot color="#ef4444" class="action-icon-badge">
        <span class="action-icon">
          <BellOutlined />
        </span>
      </a-badge>
      <AppAvatar />
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { SearchOutlined, MailOutlined, BellOutlined, MenuOutlined } from '@ant-design/icons-vue'
import AppAvatar from '@/components/common/AppAvatar.vue'
import { ALL_ROLES } from '@/constants/setupRoles'

const authStore = useAuthStore()
const appStore = useAppStore()

const nickname = computed(() => {
  return authStore.user?.nickname || authStore.username || '用户'
})

const hasRole = computed(() => !!authStore.user?.role)

// 已选创作方向的展示文案：role 为空时回退到欢迎语（失效 key 自动过滤）
const roleDisplay = computed(() => {
  const role = authStore.user?.role
  if (!role) return '欢迎来到你的创作空间'
  const labels = role
    .split(',')
    .map((key) => key.trim())
    .map((key) => ALL_ROLES.find((r) => r.key === key)?.label)
    .filter(Boolean)
  return labels.length ? labels.join(' · ') : '欢迎来到你的创作空间'
})
</script>

<style scoped>
.app-header {
  height: 72px;
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  border-bottom: 1px solid #f0f0f0;
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
  color: #64748b;
  cursor: pointer;
  padding: 4px;
}

.greeting {
  font-size: 24px;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
  line-height: 1.3;
}

.subtitle {
  font-size: 14px;
  color: #94a3b8;
  margin: 0;
  line-height: 1.4;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.search-box {
  display: flex;
  align-items: center;
  background: #f5f7fa;
  border-radius: 20px;
  padding: 0 14px;
  height: 38px;
  min-width: 200px;
  transition: background var(--transition-fast);
}

.search-box:hover {
  background: #eef0f4;
}

.search-icon {
  font-size: 16px;
  color: #94a3b8;
  margin-right: 8px;
  flex-shrink: 0;
}

.search-input {
  border: none;
  background: transparent;
  outline: none;
  font-size: 14px;
  color: #1e293b;
  width: 100%;
}

.search-input::placeholder {
  color: #94a3b8;
}

.action-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #64748b;
  cursor: pointer;
  transition: color var(--transition-fast);
  padding: 4px;
}

.action-icon:hover {
  color: #4a6cf7;
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

  .search-box {
    display: none;
  }

  .greeting {
    font-size: 20px;
  }

  .header-right {
    gap: 12px;
  }
}
</style>
