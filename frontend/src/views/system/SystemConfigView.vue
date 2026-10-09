<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">系统配置</h1>
    </div>

    <div class="settings-layout">
      <!-- 左侧导航 -->
      <div class="settings-nav">
        <button
          v-for="item in navItems"
          :key="item.key"
          type="button"
          class="settings-nav__item"
          :class="{ active: activeKey === item.key }"
          @click="activeKey = item.key"
        >
          <span class="settings-nav__icon">
            <component :is="item.icon" />
          </span>
          <span class="settings-nav__label">{{ item.label }}</span>
        </button>
      </div>

      <!-- 右侧内容区 -->
      <div class="settings-content">
        <component :is="activeComponent" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Component } from 'vue'
import {
  AppstoreOutlined,
  CloudServerOutlined,
  ApiOutlined,
  KeyOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons-vue'
import ModuleConfigView from '@/views/module/ModuleConfigView.vue'
import StorageConfigView from '@/views/storage/StorageConfigView.vue'
import CorsConfigView from '@/views/user/CorsConfigView.vue'
import MapConfigView from '@/views/map/MapConfigView.vue'
import SecurityConfigView from '@/views/security/SecurityConfigView.vue'

const ACTIVE_KEY_STORAGE = 'novablog_system_config_active'

interface NavItem {
  key: string
  label: string
  icon: Component
  component: Component
}

const navItems: NavItem[] = [
  { key: 'module', label: '模块管理', icon: AppstoreOutlined, component: ModuleConfigView },
  { key: 'storage', label: '对象存储', icon: CloudServerOutlined, component: StorageConfigView },
  { key: 'cors', label: '跨域配置', icon: ApiOutlined, component: CorsConfigView },
  { key: 'map', label: '地图配置', icon: KeyOutlined, component: MapConfigView },
  { key: 'security', label: '黑名单管理', icon: SafetyCertificateOutlined, component: SecurityConfigView },
]

function readStoredKey(): string {
  try {
    const v = localStorage.getItem(ACTIVE_KEY_STORAGE)
    if (v && navItems.some((item) => item.key === v)) return v
  } catch {
    // 存储不可用时回退到默认项
  }
  return navItems[0].key
}

const activeKey = ref(readStoredKey())

watch(activeKey, (key) => {
  try {
    localStorage.setItem(ACTIVE_KEY_STORAGE, key)
  } catch {
    // 持久化失败不影响切换
  }
})

const activeComponent = computed(() => {
  return navItems.find((item) => item.key === activeKey.value)?.component ?? navItems[0].component
})
</script>

<style scoped>
.page-container {
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #29365C;
  margin: 0;
}

.settings-layout {
  display: flex;
  align-items: flex-start;
  gap: 24px;
}

.settings-nav {
  width: 200px;
  flex-shrink: 0;
  background: #fff;
  border: 1px solid #E5E9F2;
  border-radius: 8px;
  padding: 8px;
}

.settings-nav__item {
  display: flex;
  align-items: center;
  width: 100%;
  height: 40px;
  padding: 0 12px;
  margin-bottom: 4px;
  border: none;
  border-radius: 6px;
  background: none;
  color: #29365C;
  font-size: 14px;
  text-align: left;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
  overflow: hidden;
}

.settings-nav__item:last-child {
  margin-bottom: 0;
}

.settings-nav__item:hover {
  background: #F5F7FC;
}

.settings-nav__item.active {
  color: #526FE8;
  background: rgba(82, 111, 232, 0.08);
}

.settings-nav__icon {
  display: inline-flex;
  align-items: center;
  font-size: 16px;
  margin-right: 10px;
}

.settings-content {
  flex: 1;
  min-width: 0;
  background: #fff;
  border: 1px solid #E5E9F2;
  border-radius: 8px;
  padding: 24px;
}

@media (max-width: 768px) {
  .settings-layout {
    flex-direction: column;
  }

  .settings-nav {
    width: 100%;
    display: flex;
    overflow-x: auto;
  }

  .settings-nav__item {
    width: auto;
    flex-shrink: 0;
    margin-bottom: 0;
  }

  .settings-nav__item:last-child {
    margin-bottom: 0;
    margin-right: 0;
  }
}
</style>
