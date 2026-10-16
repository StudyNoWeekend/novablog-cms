<template>
  <img
    :src="currentSrc"
    :alt="alt"
    referrerpolicy="no-referrer"
    @error="handleError"
  />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { buildSongCoverCandidates } from '@/utils/image'
import fallbackCover from '@/assets/brand/novablog-logo-mark.png'

const props = defineProps<{
  url?: string
  alt?: string
  /** 缩略图目标宽度（像素），仅对 COS 地址生效 */
  width?: number
}>()

// 降级阶段：0=缩略处理 URL → 1=原始 URL → 2=图片代理 → 兜底占位图
const stage = ref(0)
const exhausted = ref(false)

const candidates = computed(() => buildSongCoverCandidates(props.url || '', props.width ?? 160))

const currentSrc = computed(() => {
  if (candidates.value.length === 0 || exhausted.value) {
    return fallbackCover
  }
  return candidates.value[Math.min(stage.value, candidates.value.length - 1)]
})

watch(candidates, () => {
  stage.value = 0
  exhausted.value = false
})

function handleError() {
  if (stage.value < candidates.value.length - 1) {
    stage.value++
  } else {
    exhausted.value = true
  }
}
</script>
