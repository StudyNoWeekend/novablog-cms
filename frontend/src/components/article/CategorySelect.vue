<template>
  <a-select
    v-model:value="selectedValue"
    :placeholder="placeholderText"
    style="width: 100%"
    :loading="loading"
    allow-clear
    @change="handleChange"
  >
    <a-select-option v-for="cat in categories" :key="cat.id" :value="cat.id">
      {{ cat.name }}
    </a-select-option>
  </a-select>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { categoryApi } from '@/api/category'
import type { Category } from '@/types/category'

const props = withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
  type?: string
}>(), {
  type: 'article',
})

const { t } = useI18n()

// 未传入 placeholder 时使用 i18n 默认文案（computed 保证切换语言后跟随刷新）
const placeholderText = computed(() => props.placeholder ?? t('category.selectPlaceholder'))

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

function handleChange(val: string | undefined) {
  emit('update:modelValue', val || undefined)
}
</script>