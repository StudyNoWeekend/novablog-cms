<template>
  <div class="cors-config-page">
    <a-spin :spinning="loading">
      <div class="section">
        <h3 class="section-title">允许的来源（Origin）白名单</h3>
        <p class="section-desc">
          设置允许访问 API 的跨域来源地址。每行一个，支持端口号。修改后点击保存即时生效。
        </p>

        <div class="origins-input-wrapper">
          <div class="tags-container">
            <a-tag
              v-for="(origin, index) in origins"
              :key="index"
              closable
              :color="origin.includes('localhost') ? 'blue' : 'green'"
              @close="removeOrigin(index)"
            >
              {{ origin }}
            </a-tag>
            <a-input
              ref="inputRef"
              v-model:value="newOrigin"
              placeholder="输入地址后按回车添加，例如 http://localhost:3001"
              class="origin-input"
              @press-enter="addOrigin"
            />
          </div>
        </div>
      </div>

      <div class="actions">
        <a-button type="primary" :loading="saving" @click="handleSave">
          保存配置
        </a-button>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { corsApi } from '@/api/cors'

const loading = ref(false)
const saving = ref(false)
const origins = ref<string[]>([])
const newOrigin = ref('')
const inputRef = ref()

function addOrigin() {
  const val = newOrigin.value.trim()
  if (!val) return

  // 简单校验：必须以 http:// 或 https:// 开头
  if (!/^https?:\/\/.+/.test(val)) {
    message.warning('请输入有效的 URL（以 http:// 或 https:// 开头）')
    return
  }

  // 去重
  if (origins.value.includes(val)) {
    message.info('该地址已在列表中')
    newOrigin.value = ''
    return
  }

  origins.value.push(val)
  newOrigin.value = ''
}

function removeOrigin(index: number) {
  origins.value.splice(index, 1)
}

async function loadConfig() {
  loading.value = true
  try {
    const config = await corsApi.getConfig()
    origins.value = config.allowed_origins
      .split(',')
      .map((o: string) => o.trim())
      .filter(Boolean)
  } catch (err) {
    // 静默处理
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await corsApi.updateConfig({
      allowed_origins: origins.value.join(','),
    })
    message.success('跨域配置已更新，立即生效')
  } catch {
    message.error('保存失败，请重试')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.cors-config-page {
  max-width: 720px;
}

.section {
  margin-bottom: 32px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 8px;
}

.section-desc {
  font-size: 14px;
  color: #64748b;
  margin: 0 0 20px;
  line-height: 1.6;
}

.origins-input-wrapper {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 12px;
  min-height: 100px;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: flex-start;
}

.origin-input {
  width: 280px;
  border: none;
  background: transparent;
  box-shadow: none;
  padding: 4px 8px;
  font-size: 13px;
  outline: none;
}

.origin-input:focus {
  border: none;
  box-shadow: none;
}

.actions {
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}
</style>
