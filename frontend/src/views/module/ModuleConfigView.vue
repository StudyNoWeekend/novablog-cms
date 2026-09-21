<template>
  <div class="module-config-page">
    <div class="page-header">
      <h1>模块管理</h1>
      <p class="page-desc">控制各功能模块的开启状态。关闭后，博客前台将隐藏对应模块入口和内容，管理后台始终可访问。</p>
    </div>

    <a-card class="config-card" :bordered="false">
      <a-skeleton :loading="loading" active>
        <a-form layout="vertical">
          <div class="module-list">
            <div v-for="item in modules" :key="item.key" class="module-item">
              <div class="module-info">
                <div class="module-name">{{ item.label }}</div>
                <div class="module-desc">{{ item.description }}</div>
              </div>
              <div class="module-switch">
                <a-switch
                  v-model:checked="formState[item.key]"
                  :checked-children="'开启'"
                  :un-checked-children="'关闭'"
                />
              </div>
            </div>
          </div>

          <div class="form-actions">
            <a-button type="primary" :loading="saving" @click="handleSave">
              保存设置
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
import { moduleApi } from '@/api/module'
import type { UpdateModuleConfigReq } from '@/types/module'

interface ModuleItem {
  key: keyof UpdateModuleConfigReq
  label: string
  description: string
}

const modules: ModuleItem[] = [
  { key: 'article_enabled', label: '文章管理', description: '博客前台的文章列表、详情展示' },
  { key: 'media_enabled', label: '媒体管理', description: '博客前台的图片展示、媒体资源' },
  { key: 'music_enabled', label: '音乐管理', description: '博客前台的音乐播放器展示' },
  { key: 'video_enabled', label: '视频管理', description: '博客前台的视频作品展示' },
  { key: 'travel_enabled', label: '旅行管理', description: '博客前台的旅行攻略展示' },
  { key: 'portfolio_enabled', label: '作品集管理', description: '博客前台的摄影作品集展示' },
  { key: 'equipment_enabled', label: '个人设备', description: '博客前台的个人设备展示' },
  { key: 'project_enabled', label: '项目经历', description: '博客前台的项目经历展示' },
]

const loading = ref(true)
const saving = ref(false)

const formState = reactive<Record<string, boolean>>({
  article_enabled: true,
  media_enabled: true,
  music_enabled: true,
  video_enabled: true,
  travel_enabled: true,
  portfolio_enabled: true,
  equipment_enabled: true,
  project_enabled: true,
})

async function fetchConfig() {
  loading.value = true
  try {
    const config = await moduleApi.getConfig()
    if (config) {
      for (const key of Object.keys(formState)) {
        (formState as Record<string, boolean>)[key] = config[key as keyof typeof config] as boolean
      }
    }
  } catch {
    // Error handled by interceptor
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  // Only send changed fields (partial update pattern)
  const diff: UpdateModuleConfigReq = {}
  for (const item of modules) {
    diff[item.key] = formState[item.key] as boolean
  }
  saving.value = true
  try {
    await moduleApi.updateConfig(diff)
    message.success('模块配置已更新')
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
.module-config-page {
  max-width: 720px;
  margin: 0 auto;
}

.page-header {
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
}

.config-card {
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.module-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.module-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-radius: 8px;
  transition: background 0.2s;
}

.module-item:hover {
  background: #f8fafc;
}

.module-item:not(:last-child) {
  border-bottom: 1px solid #f1f5f9;
}

.module-info {
  flex: 1;
  min-width: 0;
}

.module-name {
  font-size: 15px;
  font-weight: 500;
  color: #1e293b;
}

.module-desc {
  font-size: 13px;
  color: #94a3b8;
  margin-top: 2px;
}

.module-switch {
  flex-shrink: 0;
  margin-left: 24px;
}

.form-actions {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #f1f5f9;
  display: flex;
  justify-content: flex-end;
}
</style>
