<template>
  <div class="markdown-viewer">
    <Viewer :value="value" :plugins="plugins" />
  </div>
</template>

<script setup lang="ts">
import { Viewer } from '@bytemd/vue-next'
import gfm from '@bytemd/plugin-gfm'
import highlight from '@bytemd/plugin-highlight'
import 'bytemd/dist/index.css'

defineProps<{
  value: string
}>()

const plugins = [
  gfm(),
  highlight(),
]
</script>

<style scoped>
.markdown-viewer :deep(.bytemd-viewer) {
  padding: 0;
  font-size: 14px;
  line-height: 1.7;
  background: transparent;
}

.markdown-viewer :deep(img) {
  max-width: 100%;
}
</style>
