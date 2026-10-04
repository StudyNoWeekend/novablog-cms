<template>
  <div class="app-layout">
    <AppSidebar />
    <div
      class="app-main"
      :style="{
        marginLeft: isMobile ? '0' : (appStore.sidebarCollapsed ? '72px' : '220px')
      }"
    >
      <AppHeader />
      <AppContent />
    </div>
    <div
      v-if="isMobile && !appStore.sidebarCollapsed"
      class="mobile-overlay"
      @click="appStore.sidebarCollapsed = true"
    />
    <RoleSelectModal v-model:open="roleModalVisible" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import AppContent from './AppContent.vue'
import RoleSelectModal from './RoleSelectModal.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const isMobile = ref(false)
const roleModalVisible = ref(false)

function checkMobile() {
  isMobile.value = window.innerWidth <= 768
}

onMounted(async () => {
  checkMobile()
  window.addEventListener('resize', checkMobile)

  // 进入后台先刷新用户信息（旧版本用户 localStorage 缓存缺 role 字段），
  // role 为空（旧版本用户/首装跳过）则弹窗补选；关闭不持久化，刷新后仍未选会再次弹出
  await authStore.fetchUserInfo()
  if (authStore.user && !authStore.user.role) {
    roleModalVisible.value = true
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background: #f5f7fa;
}

.app-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.mobile-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 99;
}
</style>
