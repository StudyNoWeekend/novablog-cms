<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">设置</h1>
    </div>
    <div class="settings-body">
      <div class="settings-sidebar">
        <a-menu
          mode="vertical"
          :selectedKeys="[currentKey]"
          @click="handleMenuClick"
        >
          <a-menu-item key="info">个人资料</a-menu-item>
          <a-menu-item key="storage">对象存储</a-menu-item>
        </a-menu>
      </div>
      <div class="settings-content">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Menu } from 'ant-design-vue'

const router = useRouter()
const route = useRoute()

const currentKey = computed(() => {
  if (route.path.includes('/profile/storage')) return 'storage'
  return 'info'
})

function handleMenuClick({ key }: { key: string }) {
  if (key === 'info') router.push('/profile/info')
  else if (key === 'storage') router.push('/profile/storage')
}
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

.settings-body {
  display: flex;
  gap: 24px;
  align-items: flex-start;
}

.settings-sidebar {
  width: 200px;
  flex-shrink: 0;
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.settings-content {
  flex: 1;
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  min-height: 400px;
}
</style>
