<template>
  <div class="module-config-page">
    <div class="page-header">
      <h1>模块管理</h1>
      <p class="page-desc">控制各功能模块的开启状态。关闭后，博客前台和后台菜单将隐藏对应模块，后台也无法通过输入网址访问；重新开启请从本页操作。</p>
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
                  :loading="pendingKey === item.key"
                  @change="handleToggle(item)"
                />
              </div>
            </div>
          </div>
        </a-form>
      </a-skeleton>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { moduleApi } from '@/api/module'
import { useModuleStore } from '@/stores/module'
import { MODULE_DEFS } from '@/constants/modules'
import type { UpdateModuleConfigReq } from '@/types/module'

const modules = MODULE_DEFS

const loading = ref(true)
const pendingKey = ref<string | null>(null)
const moduleStore = useModuleStore()

const formState = reactive<Record<string, boolean>>(
  Object.fromEntries(modules.map((item) => [item.key, true])),
)

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

async function handleToggle(item: (typeof modules)[number]) {
  const enabled = formState[item.key]
  pendingKey.value = item.key
  try {
    // 只提交这一个模块的字段，后端支持部分更新，其余开关不受影响
    await moduleApi.updateConfig({ [item.key]: enabled } as UpdateModuleConfigReq)
    // 强制刷新全局模块配置，让侧边栏菜单与路由拦截立即生效
    await moduleStore.fetchConfig(true)
  } catch {
    // 保存失败时回滚开关状态，错误提示由拦截器统一弹出
    formState[item.key] = !enabled
  } finally {
    pendingKey.value = null
  }
}

onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.module-config-page {
  max-width: 960px;
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
}

.config-card {
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.module-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
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
  background: #F5F7FC;
}

@media (max-width: 768px) {
  .module-list {
    grid-template-columns: 1fr;
  }
}

.module-info {
  flex: 1;
  min-width: 0;
}

.module-name {
  font-size: 15px;
  font-weight: 500;
  color: #29365C;
}

.module-desc {
  font-size: 13px;
  color: #8A93A8;
  margin-top: 2px;
}

.module-switch {
  flex-shrink: 0;
  margin-left: 24px;
}
</style>
