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
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

const route = useRoute()
const { t } = useI18n()

// meta.title 为 menu.* 的 i18n key（见 README 规范），缺失时回退到个人资料
const routeTitle = computed(() => {
  const titleKey = route.meta?.title as string | undefined
  return titleKey ? t(titleKey) : t('menu.profile.info')
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
  color: #29365C;
  margin: 0;
}

.settings-content {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  min-height: 400px;
}
</style>
