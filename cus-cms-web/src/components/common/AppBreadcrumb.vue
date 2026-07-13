<template>
  <a-breadcrumb class="app-breadcrumb">
    <a-breadcrumb-item v-for="item in breadcrumbs" :key="item.path || item.title">
      <router-link v-if="item.path" :to="item.path" class="breadcrumb-link">
        {{ item.title }}
      </router-link>
      <span v-else class="breadcrumb-current">{{ item.title }}</span>
    </a-breadcrumb-item>
  </a-breadcrumb>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const breadcrumbs = computed(() => {
  const items: Array<{ title: string; path?: string }> = []

  const matched = route.matched.filter((r) => r.meta.title)

  for (const record of matched) {
    const title = record.meta.title as string
    const path = record.path

    if (record.name === 'Dashboard' && matched.length > 1) {
      continue
    }

    if (path === route.path) {
      items.push({ title })
    } else {
      items.push({ title, path: record.redirect as string | undefined || path })
    }
  }

  if (items.length === 0) {
    items.push({ title: '首页', path: '/dashboard' })
  }

  return items
})
</script>

<style scoped>
.app-breadcrumb {
  font-size: 13px;
}

.app-breadcrumb :deep(.ant-breadcrumb-separator) {
  color: var(--text-tertiary);
}

.breadcrumb-link {
  color: var(--text-tertiary);
  transition: color var(--transition-fast);
  text-decoration: none;
}

.breadcrumb-link:hover {
  color: var(--color-primary);
}

.breadcrumb-current {
  color: var(--text-primary);
  font-weight: 500;
}
</style>