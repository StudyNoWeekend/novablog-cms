<template>
  <router-link to="/dashboard" class="app-logo">
    <span class="logo-icon">
      <svg width="28" height="28" viewBox="0 0 30 30" fill="none">
        <path
          d="M2 15L15 2L28 15L15 28L2 15Z"
          fill="#4a6cf7"
          stroke="#4a6cf7"
          stroke-width="2"
        />
        <path
          d="M10 15L13.5 18.5L20 12"
          stroke="white"
          stroke-width="2.2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
    </span>
    <transition name="logo-text-fade">
      <span v-if="!collapsed" class="logo-text">Cus CMS</span>
    </transition>
  </router-link>
</template>

<script setup lang="ts">
defineProps<{
  collapsed?: boolean
}>()
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

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: #1e293b;
  letter-spacing: 0.5px;
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
