<template>
  <div class="error-container">
    <h1 class="error-code">404</h1>
    <p class="error-text">页面不存在</p>
    <a-button type="primary" @click="$router.push('/dashboard')">返回首页</a-button>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.error-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
}
.error-code {
  font-size: 96px;
  font-weight: 700;
  color: var(--color-primary);
  line-height: 1;
}
.error-text {
  font-size: 18px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}
</style>