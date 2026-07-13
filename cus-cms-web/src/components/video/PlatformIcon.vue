<template>
  <a-tooltip :title="config?.name || platform">
    <span class="platform-icon" :style="{ color: config?.color || '#999' }">
      <svg
        :width="size"
        :height="size"
        viewBox="0 0 24 24"
        fill="currentColor"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path :d="config?.icon || ''" />
      </svg>
    </span>
  </a-tooltip>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { PLATFORM_MAP } from './platforms'

const props = withDefaults(
  defineProps<{
    platform: string
    size?: number
  }>(),
  {
    size: 24,
  },
)

const config = computed(() => PLATFORM_MAP[props.platform])
</script>

<style scoped>
.platform-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: opacity 0.2s;
}
.platform-icon:hover {
  opacity: 0.7;
}
</style>
