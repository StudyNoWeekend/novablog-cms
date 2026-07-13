<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">安全配置</h1>
    </div>

    <a-spin :spinning="loading">
      <div class="config-content">
        <!-- GET 请求限流 -->
        <a-card title="GET 请求限流" class="config-card">
          <a-form layout="horizontal" :label-col="{ span: 6 }" :wrapper-col="{ span: 14 }">
            <a-form-item label="最大令牌数">
              <a-input-number
                v-model:value="form.get_max_tokens"
                :min="1"
                style="width: 200px"
              />
              <span class="unit-hint">次</span>
            </a-form-item>
            <a-form-item label="时间窗口">
              <a-input-number
                v-model:value="form.get_window_seconds"
                :min="1"
                style="width: 200px"
              />
              <span class="unit-hint">秒</span>
            </a-form-item>
          </a-form>
        </a-card>

        <!-- POST 请求限流 -->
        <a-card title="POST 请求限流" class="config-card">
          <a-form layout="horizontal" :label-col="{ span: 6 }" :wrapper-col="{ span: 14 }">
            <a-form-item label="最大令牌数">
              <a-input-number
                v-model:value="form.post_max_tokens"
                :min="1"
                style="width: 200px"
              />
              <span class="unit-hint">次</span>
            </a-form-item>
            <a-form-item label="时间窗口">
              <a-input-number
                v-model:value="form.post_window_seconds"
                :min="1"
                style="width: 200px"
              />
              <span class="unit-hint">秒</span>
            </a-form-item>
          </a-form>
        </a-card>

        <!-- 浏览量递增限流 -->
        <a-card title="浏览量递增限流" class="config-card">
          <a-form layout="horizontal" :label-col="{ span: 6 }" :wrapper-col="{ span: 14 }">
            <a-form-item label="最大令牌数">
              <a-input-number
                v-model:value="form.view_max_tokens"
                :min="1"
                style="width: 200px"
              />
              <span class="unit-hint">次</span>
            </a-form-item>
            <a-form-item label="时间窗口">
              <a-input-number
                v-model:value="form.view_window_seconds"
                :min="1"
                style="width: 200px"
              />
              <span class="unit-hint">秒</span>
            </a-form-item>
          </a-form>
        </a-card>

        <!-- 点赞限流 -->
        <a-card title="点赞限流" class="config-card">
          <a-form layout="horizontal" :label-col="{ span: 6 }" :wrapper-col="{ span: 14 }">
            <a-form-item label="最大令牌数">
              <a-input-number
                v-model:value="form.like_max_tokens"
                :min="1"
                style="width: 200px"
              />
              <span class="unit-hint">次</span>
            </a-form-item>
            <a-form-item label="时间窗口">
              <a-input-number
                v-model:value="form.like_window_seconds"
                :min="1"
                style="width: 200px"
              />
              <span class="unit-hint">秒</span>
            </a-form-item>
          </a-form>
        </a-card>

        <!-- IP 黑名单配置 -->
        <a-card title="IP 黑名单配置" class="config-card">
          <a-form layout="horizontal" :label-col="{ span: 6 }" :wrapper-col="{ span: 14 }">
            <a-form-item label="触发阈值">
              <a-input-number
                v-model:value="form.blacklist_threshold"
                :min="1"
                style="width: 200px"
              />
              <span class="unit-hint">次（触发限流次数达到该值后加入黑名单）</span>
            </a-form-item>
            <a-form-item label="封禁时长">
              <a-input-number
                v-model:value="form.blacklist_ttl_minutes"
                :min="1"
                style="width: 200px"
              />
              <span class="unit-hint">分钟</span>
            </a-form-item>
          </a-form>
        </a-card>

        <!-- 保存按钮 -->
        <div class="save-bar">
          <a-button
            type="primary"
            :loading="saving"
            @click="handleSave"
          >
            保存配置
          </a-button>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { getSecurityConfigAPI, updateSecurityConfigAPI } from '@/api/security'
import type { SecurityConfig, UpdateSecurityConfigReq } from '@/types/security'

const loading = ref(false)
const saving = ref(false)

const form = ref<SecurityConfig>({
  get_max_tokens: 0,
  get_window_seconds: 0,
  post_max_tokens: 0,
  post_window_seconds: 0,
  view_max_tokens: 0,
  view_window_seconds: 0,
  like_max_tokens: 0,
  like_window_seconds: 0,
  blacklist_threshold: 0,
  blacklist_ttl_minutes: 0,
})

async function fetchConfig() {
  loading.value = true
  try {
    const data = await getSecurityConfigAPI()
    form.value = { ...data }
  } catch {
    // 错误由请求拦截器处理
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    const payload: UpdateSecurityConfigReq = { ...form.value }
    await updateSecurityConfigAPI(payload)
    message.success('配置保存成功')
  } catch {
    // 错误由请求拦截器处理
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.page-container {
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0;
}

.config-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 720px;
}

.config-card {
  border-radius: var(--border-radius-lg, 12px);
}

.unit-hint {
  margin-left: 12px;
  font-size: 13px;
  color: var(--text-tertiary, #94a3b8);
}

.save-bar {
  padding-top: 8px;
}

:deep(.ant-btn-primary) {
  cursor: pointer;
}

:deep(.ant-input-number) {
  cursor: pointer;
}
</style>
