<template>
  <a-popover
    v-model:open="open"
    trigger="click"
    placement="bottomLeft"
    :arrow="false"
  >
    <template #content>
      <div class="picker-panel" :style="{ width: panelWidth }">
        <a-input
          ref="searchInputRef"
          v-model:value="keyword"
          allow-clear
          :placeholder="searchPlaceholder"
        >
          <template #prefix>
            <SearchOutlined class="picker-search-icon" />
          </template>
        </a-input>
        <div class="picker-grid">
          <button
            v-for="opt in filteredOptions"
            :key="opt.key"
            type="button"
            class="picker-card"
            :class="{ selected: opt.key === modelValue }"
            :aria-pressed="opt.key === modelValue"
            @click="select(opt.key)"
          >
            <span class="picker-card__icon">
              <!-- eslint-disable-next-line vue/no-v-html -->
              <svg viewBox="0 0 24 24" aria-hidden="true" v-html="opt.image" />
            </span>
            <span class="picker-card__name">{{ opt.name }}</span>
            <span v-if="opt.description" class="picker-card__desc">{{ opt.description }}</span>
            <span v-else-if="opt.dateRange || opt.element" class="picker-card__meta">
              <span class="picker-card__date">{{ opt.dateRange }}</span>
              <span
                v-if="opt.element"
                class="picker-card__element"
                :style="{ color: elementColor(opt.element), borderColor: elementColor(opt.element) }"
              >
                {{ opt.element }}
              </span>
            </span>
            <span v-if="opt.key === modelValue" class="picker-card__check">
              <CheckOutlined />
            </span>
          </button>
        </div>
        <div v-if="filteredOptions.length === 0" class="picker-empty">未找到匹配项</div>
        <div v-if="modelValue" class="picker-footer">
          <a-button type="link" size="small" danger @click="select('')">清除选择</a-button>
        </div>
      </div>
    </template>

    <button
      type="button"
      class="picker-trigger"
      :class="{ 'is-placeholder': !selectedOption, 'is-open': open }"
      :aria-label="ariaLabel"
    >
      <template v-if="selectedOption">
        <svg class="picker-trigger__icon" viewBox="0 0 24 24" aria-hidden="true" v-html="selectedOption.image" />
        <span class="picker-trigger__text">{{ selectedOption.name }}</span>
      </template>
      <span v-else class="picker-trigger__text">{{ placeholder }}</span>
      <DownOutlined class="picker-trigger__arrow" />
    </button>
  </a-popover>
</template>

<script lang="ts">
/** 卡片选择器统一选项结构（星座/性格等元数据的 UI 视图模型） */
export interface PickerCardOption {
  key: string
  name: string
  /** 自包含 SVG 片段（24×24 viewBox 内） */
  image: string
  /** 性格：一句话介绍 */
  description?: string
  /** 星座：日期范围 */
  dateRange?: string
  /** 星座：类型（火象/土象/风象/水象） */
  element?: string
}
</script>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { CheckOutlined, DownOutlined, SearchOutlined } from '@ant-design/icons-vue'

const props = withDefaults(
  defineProps<{
    /** 已选 key（v-model:value，空串/undefined 表示未选择） */
    modelValue?: string
    options: PickerCardOption[]
    placeholder?: string
    searchPlaceholder?: string
    ariaLabel?: string
    panelWidth?: string
  }>(),
  {
    modelValue: '',
    placeholder: '请选择',
    searchPlaceholder: '搜索',
    ariaLabel: '选择',
    panelWidth: '520px',
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const open = ref(false)
const keyword = ref('')
const searchInputRef = ref()

const selectedOption = computed(() => props.options.find((opt) => opt.key === props.modelValue))

const filteredOptions = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return props.options
  return props.options.filter((opt) =>
    [opt.name, opt.key, opt.description, opt.dateRange, opt.element]
      .filter(Boolean)
      .some((text) => String(text).toLowerCase().includes(kw)),
  )
})

// 打开面板时清空搜索并聚焦输入框
watch(open, (visible) => {
  if (visible) {
    keyword.value = ''
    nextTick(() => searchInputRef.value?.focus())
  }
})

function select(key: string) {
  emit('update:modelValue', key)
  open.value = false
}

// 星座四象配色（与后端 profile_meta.go 保持一致）
const ELEMENT_COLORS: Record<string, string> = {
  火象: '#E2574C',
  土象: '#5B8C5A',
  风象: '#4A90D9',
  水象: '#7B68EE',
}

function elementColor(element: string): string {
  return ELEMENT_COLORS[element] || 'var(--text-secondary)'
}
</script>

<style scoped>
.picker-panel {
  padding: 4px 2px 2px;
}

.picker-search-icon {
  color: var(--text-tertiary);
}

.picker-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  max-height: 272px;
  margin-top: 8px;
  overflow-y: auto;
  padding-right: 2px;
}

.picker-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 3px;
  padding: 10px;
  background: #f8f9fb;
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius);
  cursor: pointer;
  text-align: left;
  transition: border-color var(--transition-fast), background var(--transition-fast), box-shadow var(--transition-fast);
}

.picker-card:hover {
  border-color: var(--color-primary);
}

.picker-card:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.picker-card.selected {
  background: var(--color-primary-light);
  border-color: var(--color-primary);
  box-shadow: 0 0 0 1px var(--color-primary) inset;
}

.picker-card__icon {
  display: inline-flex;
  width: 34px;
  height: 34px;
}

.picker-card__icon svg {
  width: 100%;
  height: 100%;
}

.picker-card__name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  line-height: 1.3;
}

.picker-card__desc {
  font-size: 11px;
  color: var(--text-tertiary);
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.picker-card__meta {
  display: flex;
  align-items: center;
  gap: 6px;
}

.picker-card__date {
  font-size: 11px;
  color: var(--text-tertiary);
}

.picker-card__element {
  font-size: 10px;
  line-height: 1;
  padding: 2px 6px;
  border: 1px solid;
  border-radius: 999px;
}

.picker-card__check {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.picker-empty {
  padding: 24px 0;
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
}

.picker-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 4px;
  border-top: 1px solid var(--border-color);
}

.picker-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-height: 32px;
  padding: 4px 11px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  cursor: pointer;
  text-align: left;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.picker-trigger:hover,
.picker-trigger.is-open {
  border-color: var(--color-primary);
}

.picker-trigger:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.picker-trigger__icon {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
}

.picker-trigger__text {
  flex: 1;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-size: 14px;
  color: var(--text-primary);
}

.picker-trigger.is-placeholder .picker-trigger__text {
  color: var(--text-tertiary);
}

.picker-trigger__arrow {
  flex-shrink: 0;
  font-size: 10px;
  color: var(--text-tertiary);
  transition: transform var(--transition-fast);
}

.picker-trigger.is-open .picker-trigger__arrow {
  transform: rotate(180deg);
}
</style>
