<template>
  <main class="app-content">
    <router-view v-slot="{ Component, route }">
      <transition name="page-fade" mode="out-in">
        <component :is="Component" :key="route.path" />
      </transition>
    </router-view>
  </main>
</template>

<script setup lang="ts">
</script>

<style scoped>
.app-content {
  flex: 1;
  padding: 24px;
  width: 100%;
  background: var(--bg-page);
  min-height: calc(100vh - var(--header-height));
}

.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.2s ease;
}

.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
}
</style>