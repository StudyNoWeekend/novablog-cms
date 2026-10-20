<template>
  <a-tooltip :title="platformLabel">
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
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n()

const config = computed(() => PLATFORM_MAP[props.platform])

// 平台显示名：跟随当前语言；未知平台时回退显示原始 key
const platformLabel = computed(() => (config.value ? t(config.value.labelKey) : props.platform))
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
