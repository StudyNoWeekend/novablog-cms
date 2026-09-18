<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ routeTitle }}</h1>
    </div>
    <div class="settings-content">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const routeTitle = computed(() => {
  return (route.meta?.title as string) || '个人资料'
})
</script>

<style scoped>
.page-container {
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.settings-content {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  min-height: 400px;
}
</style>
