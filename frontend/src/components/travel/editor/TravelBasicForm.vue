<template>
  <div class="travel-basic-form">
    <div class="section-title">
      <span class="title-bar" aria-hidden="true" />
      <h3 class="title-text">{{ $t('travel.form.basicInfo') }}</h3>
    </div>

    <a-form layout="vertical" class="basic-form">
      <!-- 攻略标题 -->
      <a-form-item
        :label="$t('travel.form.titleLabel')"
        :help="errors.title"
        :validate-status="errors.title ? 'error' : ''"
        class="form-item-full"
      >
        <a-input
          v-model:value="form.title"
          maxlength="80"
          show-count
          :placeholder="$t('travel.form.titlePlaceholder')"
          aria-required="true"
          :aria-invalid="!!errors.title"
          :aria-describedby="errors.title ? 'basic-form-title-error' : undefined"
          @blur="touched.title = true"
        />
      </a-form-item>

      <div class="form-row">
        <!-- 目的地 -->
        <a-form-item
          :label="$t('travel.form.destinationLabel')"
          :help="errors.destination"
          :validate-status="errors.destination ? 'error' : ''"
          class="form-item-half"
        >
          <a-input
            v-model:value="form.destination"
            :placeholder="$t('travel.form.destinationPlaceholder')"
            aria-required="true"
            :aria-invalid="!!errors.destination"
            :aria-describedby="errors.destination ? 'basic-form-destination-error' : undefined"
            @blur="touched.destination = true"
          />
        </a-form-item>

        <!-- 地区分类 -->
        <a-form-item
          :label="$t('travel.form.regionLabel')"
          :help="errors.region"
          :validate-status="errors.region ? 'error' : ''"
          class="form-item-half"
        >
          <a-cascader
            :options="regionOptions"
            :value="findRegionPath(form.region)"
            :placeholder="$t('travel.form.regionPlaceholder')"
            allow-clear
            class="region-cascader"
            @change="onRegionChange"
          />
          <div v-if="form.region" class="region-path">{{ regionPath(form.region) }}</div>
        </a-form-item>
      </div>

      <div class="form-row">
        <!-- 行程天数 -->
        <a-form-item
          :label="$t('travel.form.daysLabel')"
          :help="errors.days"
          :validate-status="errors.days ? 'error' : ''"
          class="form-item-half"
        >
          <a-input-number
            v-model:value="form.days"
            :min="1"
            :max="30"
            :placeholder="$t('travel.form.daysPlaceholder')"
            class="days-input"
            aria-required="true"
            :aria-invalid="!!errors.days"
            :aria-describedby="errors.days ? 'basic-form-days-error' : undefined"
            @blur="touched.days = true"
          />
        </a-form-item>

        <!-- 最佳出行月份 -->
        <a-form-item :label="$t('travel.form.bestMonthLabel')" class="form-item-half">
          <a-input v-model:value="form.bestMonth" :placeholder="$t('travel.form.bestMonthPlaceholder')" />
        </a-form-item>
      </div>

      <div class="form-row">
        <!-- 分类 -->
        <a-form-item :label="$t('travel.form.categoryLabel')" class="form-item-half">
          <a-select
            v-model:value="form.categoryId"
            :placeholder="$t('travel.form.categoryPlaceholder')"
            allow-clear
            :loading="categoriesLoading"
          >
            <a-select-option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item class="form-item-half" />
      </div>

      <!-- 封面图 -->
      <a-form-item :label="$t('common.field.cover')" class="form-item-full">
        <div class="cover-field">
          <div
            class="cover-card"
            :class="{ 'has-image': !!form.coverImage }"
            role="button"
            tabindex="0"
            :aria-label="$t('travel.form.selectCoverAria')"
            @click="pickerVisible = true"
            @keydown.enter.space.prevent="pickerVisible = true"
          >
            <template v-if="form.coverImage">
              <img :src="form.coverImage" :alt="$t('common.field.cover')" />
              <div class="cover-overlay">
                <span>{{ $t('travel.form.changeCover') }}</span>
              </div>
              <a-button
                type="text"
                danger
                size="small"
                class="cover-remove"
                :aria-label="$t('travel.form.removeCoverAria')"
                @click.stop="form.coverImage = null"
              >
                <template #icon>
                  <CloseOutlined />
                </template>
              </a-button>
            </template>
            <template v-else>
              <div class="cover-placeholder">
                <PictureOutlined class="cover-placeholder-icon" />
                <span class="cover-placeholder-text">{{ $t('travel.form.selectCoverAria') }}</span>
              </div>
            </template>
          </div>
          <div v-if="form.coverImage" class="cover-url-readonly">
            {{ form.coverImage }}
          </div>
        </div>
        <TravelCoverPicker v-model:visible="pickerVisible" @selected="onCoverSelected" />
      </a-form-item>

      <!-- 攻略摘要 -->
      <a-form-item :label="$t('travel.form.summaryLabel')" class="form-item-full">
        <a-textarea
          v-model:value="form.summary"
          :rows="3"
          maxlength="200"
          show-count
          :placeholder="$t('travel.form.summaryPlaceholder')"
        />
      </a-form-item>
    </a-form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, watch, nextTick, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { PictureOutlined, CloseOutlined } from '@ant-design/icons-vue'
import TravelCoverPicker from '@/components/travel/editor/TravelCoverPicker.vue'
import { categoryApi } from '@/api/category'
import { findRegionPath, getRegionLabelKeys, regionTree } from '@/types/travel'
import type { TravelGuideFormData, TravelRegion, TravelRegionFilter } from '@/types/travel'
import type { Category } from '@/types/category'

const { t } = useI18n()

const props = defineProps<{
  modelValue: TravelGuideFormData
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: TravelGuideFormData): void
  (e: 'daysChange', days: number): void
}>()

const form = reactive<TravelGuideFormData>({ ...props.modelValue })

const touched = reactive({
  title: false,
  destination: false,
  region: false,
  days: false,
})

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

// 地区路径（大洲 / 地区，跟随当前语言）
function regionPath(region: TravelRegionFilter): string {
  return getRegionLabelKeys(region).map((key) => t(key)).join(' / ')
}

// 校验文案（跟随当前语言）
const errors = computed(() => ({
  title: touched.title && !form.title.trim() ? t('travel.form.titleRequired') : '',
  destination: touched.destination && !form.destination.trim() ? t('travel.form.destinationRequired') : '',
  region: touched.region && !form.region ? t('travel.form.regionRequired') : '',
  days: touched.days && (!form.days || form.days < 1) ? t('travel.form.daysRequired') : '',
}))

const pickerVisible = ref(false)

const categories = ref<Category[]>([])
const categoriesLoading = ref(false)

onMounted(async () => {
  categoriesLoading.value = true
  try {
    categories.value = await categoryApi.getList('travel')
  } catch {}
  finally { categoriesLoading.value = false }
})

function onCoverSelected(payload: { url: string }) {
  form.coverImage = payload.url
  pickerVisible.value = false
}

let isSyncingFromParent = false

watch(
  () => ({ ...form }),
  (val) => {
    if (!isSyncingFromParent) {
      emit('update:modelValue', val as TravelGuideFormData)
    }
  },
  { deep: true },
)

watch(
  () => props.modelValue,
  (val) => {
    isSyncingFromParent = true
    Object.assign(form, val)
    nextTick(() => {
      isSyncingFromParent = false
    })
  },
  { deep: true },
)

watch(
  () => form.days,
  (newDays, oldDays) => {
    if (newDays === null || newDays === undefined || Number.isNaN(newDays)) {
      const fallback = typeof oldDays === 'number' && !Number.isNaN(oldDays) ? oldDays : 1
      form.days = fallback
      return
    }
    const clamped = Math.max(1, Math.min(30, newDays))
    if (clamped !== newDays) {
      form.days = clamped
      return
    }
    if (newDays !== oldDays) {
      emit('daysChange', newDays)
    }
  },
)

function onRegionChange(value: string[] | undefined) {
  const selected = value?.length ? value[value.length - 1] : ''
  form.region = selected as TravelRegion
}

function isValid(): boolean {
  touched.title = true
  touched.destination = true
  touched.region = true
  touched.days = true
  return (
    !!form.title.trim() &&
    !!form.destination.trim() &&
    !!form.region &&
    form.days >= 1
  )
}

defineExpose({ isValid })
</script>

<style scoped>
.travel-basic-form {
  width: 100%;
  padding: 24px;
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
}

.title-bar {
  width: 4px;
  height: 20px;
  background: var(--color-primary, #526FE8);
  border-radius: 2px;
}

.title-text {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary, #29365C);
}

.basic-form {
  width: 100%;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.form-item-full {
  width: 100%;
}

.form-item-half {
  width: 100%;
  min-width: 0;
}

.region-cascader {
  width: 100%;
}

.region-path {
  margin-top: 8px;
  font-size: 13px;
  color: var(--text-secondary, #667085);
}

.days-input {
  width: 100%;
}

.cover-field {
  width: 100%;
}

.cover-card {
  position: relative;
  width: 100%;
  max-width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: var(--border-radius-lg, 12px);
  overflow: hidden;
  cursor: pointer;
  background: var(--bg-secondary, #F5F7FC);
  border: 1px solid var(--border-color, #E5E9F2);
  transition: border-color 0.2s ease;
}

.cover-card.has-image {
  border: none;
}

.cover-card:not(.has-image):hover,
.cover-card:not(.has-image):focus-visible {
  border-color: var(--color-primary, #526FE8);
}

.cover-card img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.cover-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  border: 2px dashed var(--border-color, #E5E9F2);
  border-radius: inherit;
  color: var(--text-secondary, #667085);
  transition: border-color 0.2s ease, color 0.2s ease;
}

.cover-card:hover .cover-placeholder,
.cover-card:focus-visible .cover-placeholder {
  border-color: var(--color-primary, #526FE8);
  color: var(--color-primary, #526FE8);
}

.cover-placeholder-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.cover-placeholder-text {
  font-size: 14px;
}

.cover-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  color: #ffffff;
  font-size: 14px;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s ease;
}

.cover-card.has-image:hover .cover-overlay,
.cover-card.has-image:focus-visible .cover-overlay {
  opacity: 1;
}

.cover-remove {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 2;
  background: rgba(255, 255, 255, 0.85) !important;
  border-radius: 50% !important;
}

.cover-url-readonly {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-secondary, #667085);
  word-break: break-all;
}

@media (max-width: 768px) {
  .travel-basic-form {
    padding: 16px;
  }

  .form-row {
    grid-template-columns: 1fr;
    gap: 0;
  }
}
</style>
