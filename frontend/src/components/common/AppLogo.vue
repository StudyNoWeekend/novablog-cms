<template>
  <router-link to="/dashboard" class="app-logo">
    <span class="logo-icon">
      <img :src="logoMark" alt="Novablog" class="logo-img" />
    </span>
    <transition name="logo-text-fade">
      <div v-if="!collapsed" class="logo-brand">
        <span class="logo-text">Novablog</span>
        <span class="logo-version">{{ appStore.appVersion }}</span>
      </div>
    </transition>
  </router-link>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import logoMark from '@/assets/brand/novablog-logo-mark.png'

defineProps<{
  collapsed?: boolean
}>()

const appStore = useAppStore()

onMounted(() => {
  appStore.fetchVersion()
})
</script>

<style scoped>
.app-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
}

.logo-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-img {
  display: block;
  width: 28px;
  height: 28px;
}

.logo-brand {
  display: flex;
  flex-direction: column;
  justify-content: center;
  line-height: 1;
  gap: 3px;
  white-space: nowrap;
  overflow: hidden;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: #29365C;
  letter-spacing: 0.5px;
  white-space: nowrap;
  overflow: hidden;
}

.logo-version {
  font-size: 11px;
  color: #8A93A8;
  white-space: nowrap;
  overflow: hidden;
}

.logo-text-fade-enter-active,
.logo-text-fade-leave-active {
  transition: opacity 0.25s ease, max-width 0.25s ease;
  max-width: 120px;
}

.logo-text-fade-enter-from,
.logo-text-fade-leave-to {
  opacity: 0;
  max-width: 0;
}
</style>
