<template>
  <div class="map-config-page">
    <div class="page-header">
      <h1>地图配置</h1>
      <p class="page-desc">
        配置旅行攻略选点使用的地图服务 Key。未配置或加载失败时，选点弹窗自动降级为
        OSM 免 Key 模式（WGS-84 坐标）；博客端地图展示不受此处影响，始终使用免 Key 方案。
      </p>
    </div>

    <a-card class="config-card" :bordered="false">
      <a-skeleton :loading="loading" active>
        <a-form layout="vertical">
          <a-form-item required label="高德 Key">
            <a-input-password
              v-model:value="form.amap_key"
              placeholder="高德开放平台 Web端(JS API) Key"
              autocomplete="new-password"
            />
            <div class="field-tip">
              在 <a href="https://console.amap.com" target="_blank" rel="noopener noreferrer">高德开放平台控制台</a>
              创建「Web端(JS API)」类型 Key 后填入；配置后中国地区选点使用高德地图（GCJ-02 坐标）。
            </div>
          </a-form-item>

          <a-form-item label="高德安全密钥（securityJsCode）">
            <a-input-password
              v-model:value="form.amap_security_code"
              placeholder="与高德 Key 配套的安全密钥；留空表示保持后端已存值"
              autocomplete="new-password"
            />
            <div class="field-tip">与高德 2021-12 后申请的 Key 配套；保存后加密存储，读取仅限登录后台。</div>
          </a-form-item>

          <a-form-item required label="Google Maps Key">
            <a-input-password
              v-model:value="form.google_key"
              placeholder="Google Maps JavaScript API Key"
              autocomplete="new-password"
            />
            <div class="field-tip">配置后海外地区选点使用 Google Maps（WGS-84 坐标）。</div>
          </a-form-item>

          <div class="form-actions">
            <a-button type="primary" :loading="saving" @click="handleSave">
              保存配置
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
import { mapApi } from '@/api/map'

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
    message.success('地图配置已更新，选点弹窗下次打开时生效')
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
  color: #1e293b;
  margin: 0 0 8px 0;
}

.page-desc {
  color: #64748b;
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
  color: #94a3b8;
  margin-top: 4px;
  line-height: 1.5;
}

.field-tip a {
  color: #4a6cf7;
}

.form-actions {
  margin-top: 8px;
  padding-top: 20px;
  border-top: 1px solid #f1f5f9;
  display: flex;
  justify-content: flex-end;
}
</style>
