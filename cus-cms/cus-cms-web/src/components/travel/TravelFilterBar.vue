<template>
  <div class="travel-filter-bar">
    <a-cascader
      :options="regionTree"
      :value="findRegionPath(localFilters.region)"
      allow-clear
      placeholder="选择地区"
      change-on-select
      class="region-cascader"
      @change="onRegionChange"
    />

    <div class="filter-controls">
      <a-select
        v-model:value="localFilters.days"
        :options="daysOptions"
        aria-label="天数筛选"
        class="filter-select"
        @change="onDaysChange"
      />
      <a-select
        v-model:value="localFilters.sort"
        :options="sortOptions"
        aria-label="排序方式"
        class="filter-select"
        @change="onSortChange"
      />
      <a-input-search
        v-model:value="keyword"
        placeholder="搜索攻略标题、目的地..."
        allow-clear
        aria-label="搜索攻略"
        class="filter-search"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useDebounce } from '@/composables/useDebounce'
import { findRegionPath, regionTree } from '@/types/travel'
import type { TravelFilters, TravelRegionFilter } from '@/types/travel'

const props = defineProps<{
  modelValue: TravelFilters
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: TravelFilters): void
}>()

const daysOptions = [
  { value: 'all', label: '全部' },
  { value: '1-3', label: '1-3 天' },
  { value: '4-7', label: '4-7 天' },
  { value: '8-14', label: '8-14 天' },
  { value: '15+', label: '15 天以上' },
]

const sortOptions = [
  { value: 'latest', label: '最新发布' },
  { value: 'views', label: '最多浏览' },
  { value: 'rating', label: '最高评分' },
  { value: 'likes', label: '最多点赞' },
]

const keyword = ref(props.modelValue.keyword ?? '')
const { debouncedValue, setDebounce } = useDebounce(keyword, 300)

const localFilters = ref<Required<TravelFilters>>({
  keyword: props.modelValue.keyword ?? '',
  region: props.modelValue.region ?? 'all',
  days: props.modelValue.days ?? 'all',
  sort: props.modelValue.sort ?? 'latest',
  categoryId: props.modelValue.categoryId ?? '',
})

watch(
  () => props.modelValue,
  (val) => {
    localFilters.value = {
      keyword: val.keyword ?? '',
      region: val.region ?? 'all',
      days: val.days ?? 'all',
      sort: val.sort ?? 'latest',
      categoryId: val.categoryId ?? '',
    }
    keyword.value = val.keyword ?? ''
  },
  { deep: true },
)

watch(keyword, (val) => {
  setDebounce(val)
})

watch(debouncedValue, (val) => {
  localFilters.value.keyword = val
  emitUpdate()
})

function onRegionChange(value: string[] | undefined) {
  const selected = value?.length ? value[value.length - 1] : 'all'
  localFilters.value.region = selected as TravelRegionFilter
  emitUpdate()
}

function onDaysChange() {
  emitUpdate()
}

function onSortChange() {
  emitUpdate()
}

function emitUpdate() {
  emit('update:modelValue', { ...localFilters.value })
}
</script>

<style scoped>
.travel-filter-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
}

.region-cascader {
  width: 180px;
}

.filter-controls {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-left: auto;
}

.filter-controls :deep(.ant-select),
.filter-controls :deep(.ant-input-affix-wrapper) {
  min-height: 44px;
}

.filter-select {
  width: 140px;
}

.filter-search {
  width: 240px;
}

@media (max-width: 768px) {
  .travel-filter-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .region-cascader {
    width: 100%;
  }

  .filter-controls {
    margin-left: 0;
    width: 100%;
  }

  .filter-select {
    flex: 1;
    min-width: 120px;
  }

  .filter-search {
    width: 100%;
  }
}
</style>
