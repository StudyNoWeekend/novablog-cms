<template>
  <a-dropdown :trigger="['click']">
    <span class="avatar-trigger">
      <a-avatar :size="32" :style="{ backgroundColor: '#4a6cf7' }">
        {{ avatarText }}
      </a-avatar>
      <span class="avatar-name">{{ authStore.username }}</span>
      <DownOutlined class="avatar-arrow" />
    </span>
    <template #overlay>
      <a-menu class="avatar-menu" @click="handleMenuClick">
        <a-menu-item key="profile">
          <UserOutlined class="menu-item-icon" />
          <span>个人资料</span>
        </a-menu-item>
        <a-menu-divider />
        <a-menu-item key="logout">
          <LogoutOutlined class="menu-item-icon" />
          <span>退出登录</span>
        </a-menu-item>
      </a-menu>
    </template>
  </a-dropdown>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { DownOutlined, UserOutlined, LogoutOutlined } from '@ant-design/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const avatarText = computed(() => {
  return authStore.username ? authStore.username.charAt(0).toUpperCase() : 'U'
})

function handleMenuClick({ key }: { key: string }) {
  if (key === 'profile') {
    router.push('/profile')
  } else if (key === 'logout') {
    authStore.logout()
  }
}
</script>

<style scoped>
.avatar-trigger {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 0;
}

.avatar-name {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
}

.avatar-arrow {
  font-size: 11px;
  color: var(--text-tertiary);
  transition: transform var(--transition-fast);
}

.avatar-trigger:hover .avatar-arrow {
  color: var(--text-secondary);
}

.avatar-menu {
  min-width: 140px;
}

.menu-item-icon {
  margin-right: 8px;
  font-size: 15px;
}
</style>