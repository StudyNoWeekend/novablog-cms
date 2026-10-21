<template>
  <div class="travel-filter-bar">
    <a-cascader
      :options="regionOptions"
      :value="findRegionPath(localFilters.region)"
      allow-clear
      :placeholder="$t('travel.filter.regionPlaceholder')"
      change-on-select
      class="region-cascader"
      @change="onRegionChange"
    />

    <div class="filter-controls">
      <a-select
        v-model:value="localFilters.days"
        :options="daysOptions"
        :aria-label="$t('travel.filter.daysAria')"
        class="filter-select"
        @change="onDaysChange"
      />
      <a-select
        v-model:value="localFilters.sort"
        :options="sortOptions"
        :aria-label="$t('travel.filter.sortAria')"
        class="filter-select"
        @change="onSortChange"
      />
      <a-input-search
        v-model:value="keyword"
        :placeholder="$t('travel.filter.searchPlaceholder')"
        allow-clear
        :aria-label="$t('travel.filter.searchAria')"
        class="filter-search"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDebounce } from '@/composables/useDebounce'
import { findRegionPath, regionTree } from '@/types/travel'
import type { TravelFilters, TravelRegionFilter } from '@/types/travel'

const { t } = useI18n()

const props = defineProps<{
  modelValue: TravelFilters
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: TravelFilters): void
}>()

// 地区级联选项（labelKey → 当前语言文案）
const regionOptions = computed(() =>
  regionTree.map((continent) => ({
    value: continent.value,
    label: t(continent.labelKey),
    children: continent.children?.map((child) => ({
      value: child.value,
      label: t(child.labelKey),
    })),
  })),
)

// 天数筛选选项（跟随当前语言）
const daysOptions = computed(() => [
  { value: 'all', label: t('common.status.all') },
  { value: '1-3', label: t('travel.filter.days1To3') },
  { value: '4-7', label: t('travel.filter.days4To7') },
  { value: '8-14', label: t('travel.filter.days8To14') },
  { value: '15+', label: t('travel.filter.days15Plus') },
])

// 排序选项（跟随当前语言）
const sortOptions = computed(() => [
  { value: 'latest', label: t('travel.sort.latest') },
  { value: 'views', label: t('travel.sort.mostViewed') },
  { value: 'rating', label: t('travel.sort.topRated') },
  { value: 'likes', label: t('travel.sort.mostLiked') },
])

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
