<template>
  <div class="cors-config-page">
    <a-spin :spinning="loading">
      <div class="section">
        <h3 class="section-title">{{ $t('profile.cors.title') }}</h3>
        <p class="section-desc">
          {{ $t('profile.cors.desc') }}
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
              :placeholder="$t('profile.cors.inputPlaceholder')"
              class="origin-input"
              @press-enter="addOrigin"
            />
          </div>
        </div>
      </div>

      <div class="actions">
        <a-button type="primary" :loading="saving" @click="handleSave">
          {{ $t('profile.cors.saveConfig') }}
        </a-button>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import { corsApi } from '@/api/cors'

const { t } = useI18n()

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
    message.warning(t('profile.cors.invalidUrl'))
    return
  }

  // 去重
  if (origins.value.includes(val)) {
    message.info(t('profile.cors.duplicate'))
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
    message.success(t('profile.cors.updateSuccess'))
  } catch {
    message.error(t('profile.cors.saveFailed'))
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
  color: #29365C;
  margin: 0 0 8px;
}

.section-desc {
  font-size: 14px;
  color: #667085;
  margin: 0 0 20px;
  line-height: 1.6;
}

.origins-input-wrapper {
  background: #F5F7FC;
  border: 1px solid #E5E9F2;
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
  border-top: 1px solid #E5E9F2;
}
</style>
