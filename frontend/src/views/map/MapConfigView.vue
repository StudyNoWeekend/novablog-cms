<template>
  <div class="map-config-page">
    <div class="page-header">
      <h1>{{ $t('map.title') }}</h1>
      <p class="page-desc">
        {{ $t('map.desc') }}
      </p>
    </div>

    <a-card class="config-card" :bordered="false">
      <a-skeleton :loading="loading" active>
        <a-form layout="vertical">
          <a-form-item required :label="$t('map.form.amapKey')">
            <a-input-password
              v-model:value="form.amap_key"
              :placeholder="$t('map.form.amapKeyPlaceholder')"
              autocomplete="new-password"
            />
            <div class="field-tip">
              {{ $t('map.form.amapKeyTipBefore') }}<a href="https://console.amap.com" target="_blank" rel="noopener noreferrer">{{ $t('map.form.amapConsoleLink') }}</a>{{ $t('map.form.amapKeyTipAfter') }}
            </div>
          </a-form-item>

          <a-form-item :label="$t('map.form.amapSecurityCode')">
            <a-input-password
              v-model:value="form.amap_security_code"
              :placeholder="$t('map.form.amapSecurityCodePlaceholder')"
              autocomplete="new-password"
            />
            <div class="field-tip">{{ $t('map.form.amapSecurityCodeTip') }}</div>
          </a-form-item>

          <a-form-item required :label="$t('map.form.googleKey')">
            <a-input-password
              v-model:value="form.google_key"
              :placeholder="$t('map.form.googleKeyPlaceholder')"
              autocomplete="new-password"
            />
            <div class="field-tip">{{ $t('map.form.googleKeyTip') }}</div>
          </a-form-item>

          <div class="form-actions">
            <a-button type="primary" :loading="saving" @click="handleSave">
              {{ $t('map.actions.save') }}
            </a-button>
          </div>
        </a-form>
      </a-skeleton>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import { mapApi } from '@/api/map'

const { t } = useI18n()
const loading = ref(true)
const saving = ref(false)

const form = reactive({
  amap_key: '',
  amap_security_code: '',
  google_key: '',
})

async function fetchConfig() {
  loading.value = true
  try {
    const config = await mapApi.getConfig()
    if (config) {
      form.amap_key = config.amap_key || ''
      form.amap_security_code = config.amap_security_code || ''
      form.google_key = config.google_key || ''
    }
  } catch {
    // Error handled by interceptor
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    const config = await mapApi.updateConfig({ ...form })
    form.amap_key = config.amap_key || ''
    form.amap_security_code = config.amap_security_code || ''
    form.google_key = config.google_key || ''
    message.success(t('map.message.saveSuccess'))
  } catch {
    // Error handled by interceptor
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.map-config-page {
  max-width: 720px;
  margin: 0 auto;
}

.page-header {
  /* 显式块级布局，避免全局 .page-header flex 规则把标题与描述排成一行 */
  display: block;
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #29365C;
  margin: 0 0 8px 0;
}

.page-desc {
  color: #667085;
  font-size: 14px;
  margin: 0;
  line-height: 1.6;
}

.config-card {
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.field-tip {
  font-size: 12px;
  color: #8A93A8;
  margin-top: 4px;
  line-height: 1.5;
}

.field-tip a {
  color: #526FE8;
}

.form-actions {
  margin-top: 8px;
  padding-top: 20px;
  border-top: 1px solid #F5F7FC;
  display: flex;
  justify-content: flex-end;
}
</style>
