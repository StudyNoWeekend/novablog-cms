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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import AppContent from './AppContent.vue'

const appStore = useAppStore()
const isMobile = ref(false)

function checkMobile() {
  isMobile.value = window.innerWidth <= 768
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
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
