<template>
  <a-select
    v-model:value="selectedValue"
    :placeholder="placeholder"
    style="width: 100%"
    :loading="loading"
    @change="handleChange"
  >
    <a-select-option v-for="cat in categories" :key="cat.id" :value="cat.id">
      {{ cat.name }}
    </a-select-option>
  </a-select>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { categoryApi } from '@/api/category'
import type { Category } from '@/types/category'

const props = withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
  type?: string
}>(), {
  placeholder: '选择分类',
  type: 'article',
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string | undefined): void
}>()

const categories = ref<Category[]>([])
const loading = ref(false)
const selectedValue = ref<string | undefined>(props.modelValue)

watch(() => props.modelValue, (val) => { selectedValue.value = val })

onMounted(async () => {
  loading.value = true
  try { categories.value = await categoryApi.getList(props.type) }
  catch {}
  finally { loading.value = false }
})

function handleChange(val: string) {
  emit('update:modelValue', val)
}
</script>