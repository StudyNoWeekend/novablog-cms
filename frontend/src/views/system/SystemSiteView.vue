<template>
  <div class="system-site-view">
    <div class="site-header">
      <h2 class="site-title">{{ $t('system.site.title') }}</h2>
      <p class="site-desc">{{ $t('system.site.desc') }}</p>
    </div>

    <!-- 卡 1：站点与集成 -->
    <a-card :title="$t('system.site.integrationTitle')" class="config-card" :bordered="false">
      <template #extra>
        <a-tag color="blue">{{ $t('system.site.effectiveImmediate') }}</a-tag>
      </template>

      <a-skeleton :loading="loading" active>
        <a-form layout="vertical">
          <a-form-item :label="$t('system.site.siteUrl')">
            <a-input
              v-model:value="siteForm.site_url"
              :placeholder="$t('system.site.siteUrlPlaceholder')"
              allow-clear
            />
            <div class="field-tip">{{ $t('system.site.siteUrlTip') }}</div>
          </a-form-item>

          <a-form-item :label="$t('system.site.proxyUrl')">
            <a-input
              v-model:value="siteForm.proxy_url"
              :placeholder="$t('system.site.proxyUrlPlaceholder')"
              allow-clear
            />
            <div class="field-tip">{{ $t('system.site.proxyUrlTip') }}</div>
          </a-form-item>

          <a-form-item :label="$t('system.site.githubToken')">
            <a-input-password
              v-model:value="siteForm.github_token"
              :placeholder="$t('system.site.githubTokenPlaceholder')"
              autocomplete="new-password"
            />
            <div class="field-tip">{{ $t('system.site.githubTokenTip') }}</div>
          </a-form-item>

          <a-form-item :label="$t('system.site.maxArtifactMb')">
            <a-input-number v-model:value="siteForm.max_artifact_mb" :min="1" style="width: 100%" />
            <div class="field-tip">{{ $t('system.site.maxArtifactMbTip') }}</div>
          </a-form-item>

          <a-form-item :label="$t('system.site.cloudMaxFileSizeMb')">
            <a-input-number v-model:value="siteForm.cloud_max_file_size_mb" :min="1" style="width: 100%" />
            <div class="field-tip">{{ $t('system.site.cloudMaxFileSizeMbTip') }}</div>
          </a-form-item>

          <a-form-item :label="$t('system.site.cloudImportBatchLimit')">
            <a-input-number v-model:value="siteForm.cloud_import_batch_limit" :min="1" style="width: 100%" />
            <div class="field-tip">{{ $t('system.site.cloudImportBatchLimitTip') }}</div>
          </a-form-item>

          <a-form-item :label="$t('system.site.geoipDownloadUrl')">
            <a-input
              v-model:value="siteForm.geoip_download_url"
              :placeholder="$t('system.site.geoipDownloadUrlPlaceholder')"
              allow-clear
            />
            <div class="field-tip">{{ $t('system.site.geoipDownloadUrlTip') }}</div>
          </a-form-item>

          <div class="form-actions">
            <a-button type="primary" :loading="saving" @click="handleSaveSite">
              {{ $t('system.site.saveSite') }}
            </a-button>
          </div>
        </a-form>
      </a-skeleton>
    </a-card>

    <!-- 卡 2：安全与日志 -->
    <a-card :title="$t('system.site.securityTitle')" class="config-card" :bordered="false">
      <template #extra>
        <a-tag color="blue">{{ $t('system.site.effectiveImmediate') }}</a-tag>
      </template>

      <a-skeleton :loading="loading" active>
        <a-form layout="vertical">
          <a-form-item :label="$t('system.site.securityEnabled')">
            <a-switch v-model:checked="secForm.security_enabled" />
            <div class="field-tip">{{ $t('system.site.securityEnabledTip') }}</div>
          </a-form-item>

          <a-form-item :label="$t('system.site.blacklistTtl')">
            <a-input-number
              v-model:value="secForm.blacklist_ttl_minutes"
              :min="1"
              :max="1440"
              style="width: 100%"
            />
            <div class="field-tip">{{ $t('system.site.blacklistTtlTip') }}</div>
          </a-form-item>

          <a-form-item :label="$t('system.site.logRetention')">
            <a-input-number
              v-model:value="secForm.log_retention_days"
              :min="1"
              :max="365"
              style="width: 100%"
            />
            <div class="field-tip">{{ $t('system.site.logRetentionTip') }}</div>
          </a-form-item>

          <div class="form-actions">
            <a-button type="primary" :loading="savingSecurity" @click="handleSaveSecurity">
              {{ $t('system.site.saveSecurity') }}
            </a-button>
          </div>
        </a-form>
      </a-skeleton>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import { getSystemConfigAPI, updateSystemConfigAPI } from '@/api/system'
import { getSecurityConfigAPI, updateSecurityConfigAPI } from '@/api/security'

const { t } = useI18n()

const loading = ref(true)
const saving = ref(false)
const savingSecurity = ref(false)

// 键固定为这七项：文本键直接比对字符串，数字键留空提交空串恢复内置默认
const TEXT_KEYS = ['site_url', 'proxy_url', 'github_token', 'geoip_download_url'] as const
const NUMBER_KEYS = ['max_artifact_mb', 'cloud_max_file_size_mb', 'cloud_import_batch_limit'] as const

const siteForm = reactive({
  site_url: '',
  proxy_url: '',
  github_token: '',
  geoip_download_url: '',
  max_artifact_mb: null as number | null,
  cloud_max_file_size_mb: null as number | null,
  cloud_import_batch_limit: null as number | null,
})

// original 保存加载时的原始值（字符串形态），保存时只提交变化过的键
const original = reactive<Record<string, string>>({})

const secForm = reactive({
  security_enabled: false,
  blacklist_ttl_minutes: 1 as number | null,
  log_retention_days: 1 as number | null,
})

// toNumberOrNull 数字配置解析：空串/非法值视为未设置（null），保存时提交空串清空
function toNumberOrNull(raw: string | undefined): number | null {
  if (!raw) return null
  const n = Number(raw)
  return Number.isFinite(n) ? n : null
}

async function fetchSiteConfig() {
  loading.value = true
  try {
    const res = await getSystemConfigAPI()
    const map: Record<string, string> = {}
    for (const item of res.items) {
      map[item.key] = item.value ?? ''
    }
    for (const key of [...TEXT_KEYS, ...NUMBER_KEYS]) {
      original[key] = map[key] ?? ''
    }
    siteForm.site_url = original.site_url
    siteForm.proxy_url = original.proxy_url
    siteForm.github_token = original.github_token
    siteForm.geoip_download_url = original.geoip_download_url
    siteForm.max_artifact_mb = toNumberOrNull(original.max_artifact_mb)
    siteForm.cloud_max_file_size_mb = toNumberOrNull(original.cloud_max_file_size_mb)
    siteForm.cloud_import_batch_limit = toNumberOrNull(original.cloud_import_batch_limit)
  } catch {
    // 错误由拦截器统一提示
  } finally {
    loading.value = false
  }
}

async function handleSaveSite() {
  const updates: Record<string, string | null> = {}
  for (const key of TEXT_KEYS) {
    const next = siteForm[key].trim()
    if (next !== original[key]) updates[key] = next
  }
  for (const key of NUMBER_KEYS) {
    const next = siteForm[key] == null ? '' : String(siteForm[key])
    if (next !== original[key]) updates[key] = next
  }
  if (Object.keys(updates).length === 0) {
    message.info(t('system.site.noChanges'))
    return
  }
  saving.value = true
  try {
    const res = await updateSystemConfigAPI(updates)
    message.success(t('system.site.saveSuccess'))
    if (res.site_url_changed) {
      message.warning(t('system.site.siteUrlChangedWarning'), 5)
    }
    await fetchSiteConfig()
  } catch {
    // 错误由拦截器统一提示
  } finally {
    saving.value = false
  }
}

async function fetchSecurityConfig() {
  try {
    const config = await getSecurityConfigAPI()
    if (config) {
      secForm.security_enabled = config.security_enabled
      secForm.blacklist_ttl_minutes = config.blacklist_ttl_minutes
      secForm.log_retention_days = config.log_retention_days
    }
  } catch {
    // 错误由拦截器统一提示
  }
}

async function handleSaveSecurity() {
  savingSecurity.value = true
  try {
    await updateSecurityConfigAPI({
      security_enabled: secForm.security_enabled,
      // 清空表示保持后端现值（该接口不支持恢复默认，故不传空）
      blacklist_ttl_minutes: secForm.blacklist_ttl_minutes ?? undefined,
      log_retention_days: secForm.log_retention_days ?? undefined,
    })
    message.success(t('system.site.securitySaveSuccess'))
    await fetchSecurityConfig()
  } catch {
    // 错误由拦截器统一提示
  } finally {
    savingSecurity.value = false
  }
}

onMounted(() => {
  fetchSiteConfig()
  fetchSecurityConfig()
})
</script>

<style scoped>
.system-site-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 760px;
}

.site-header {
  margin-bottom: 4px;
}

.site-title {
  font-size: 18px;
  font-weight: 600;
  color: #29365C;
  margin: 0 0 4px 0;
}

.site-desc {
  color: #667085;
  font-size: 13px;
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

.form-actions {
  margin-top: 8px;
  padding-top: 20px;
  border-top: 1px solid #F5F7FC;
  display: flex;
  justify-content: flex-end;
}
</style>
