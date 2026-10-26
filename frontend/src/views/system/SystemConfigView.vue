<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('menu.system.config') }}</h1>
      <div class="page-actions">
        <a-dropdown :trigger="['click']">
          <a-button :loading="exporting">
            <DownloadOutlined />
            {{ $t('system.export.button') }}
            <DownOutlined class="page-actions__caret" />
          </a-button>
          <template #overlay>
            <a-menu @click="handleExportMenu">
              <a-menu-item key="with-secrets">
                <ExportOutlined />
                {{ $t('system.export.withSecrets') }}
              </a-menu-item>
              <a-menu-item key="without-secrets">
                <ExportOutlined />
                {{ $t('system.export.withoutSecrets') }}
              </a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>

        <a-upload
          accept=".json,application/json"
          :show-upload-list="false"
          :before-upload="handleImportFile"
        >
          <a-button>
            <UploadOutlined />
            {{ $t('system.import.button') }}
          </a-button>
        </a-upload>
      </div>
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
        <component :is="activeComponent" :key="`${activeKey}:${refreshTick}`" />
      </div>
    </div>

    <!-- 导入确认 / 结果弹窗 -->
    <a-modal
      :open="importModalOpen"
      :title="importResult ? $t('system.import.resultTitle') : $t('system.import.confirmTitle')"
      :mask-closable="false"
      @cancel="closeImportModal"
    >
      <template #footer>
        <template v-if="importResult">
          <a-button type="primary" @click="closeImportModal">
            {{ $t('common.actions.confirm') }}
          </a-button>
        </template>
        <template v-else>
          <a-button :disabled="importing" @click="closeImportModal">
            {{ $t('common.actions.cancel') }}
          </a-button>
          <a-button type="primary" :loading="importing" @click="confirmImport">
            {{ $t('common.actions.confirm') }}
          </a-button>
        </template>
      </template>

      <template v-if="!importResult">
        <p class="import-groups-label">{{ $t('system.import.groupsLabel') }}</p>
        <div class="import-groups">
          <a-tag v-for="name in importGroupNames" :key="name" color="blue">{{ name }}</a-tag>
        </div>
        <a-alert type="warning" show-icon :message="$t('system.import.overwriteWarning')" />
      </template>
      <template v-else>
        <p class="import-groups-label">{{ $t('system.import.success') }}</p>
        <ul class="import-applied">
          <li v-for="(count, group) in importResult" :key="group">
            <span class="import-applied__group">{{ group }}</span>
            <span>{{ $t('system.import.appliedItem', { count }) }}</span>
          </li>
        </ul>
      </template>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Component } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { message, type UploadProps } from 'ant-design-vue'
import {
  AppstoreOutlined,
  CloudServerOutlined,
  ApiOutlined,
  KeyOutlined,
  SafetyCertificateOutlined,
  SendOutlined,
  GlobalOutlined,
  DownloadOutlined,
  UploadOutlined,
  ExportOutlined,
  DownOutlined,
} from '@ant-design/icons-vue'
import ModuleConfigView from '@/views/module/ModuleConfigView.vue'
import StorageConfigView from '@/views/storage/StorageConfigView.vue'
import CorsConfigView from '@/views/user/CorsConfigView.vue'
import MapConfigView from '@/views/map/MapConfigView.vue'
import SecurityConfigView from '@/views/security/SecurityConfigView.vue'
import CommunityPushConfigView from '@/views/system/CommunityPushConfigView.vue'
import SystemSiteView from '@/views/system/SystemSiteView.vue'
import { exportSystemConfigAPI, importSystemConfigAPI } from '@/api/system'
import type { SystemConfigExportRes } from '@/api/system'

const ACTIVE_KEY_STORAGE = 'novablog_system_config_active'

interface NavItem {
  key: string
  label: string
  icon: Component
  component: Component
}

const { t } = useI18n()

// navItems 左侧导航（标签随语言切换，包成 computed 保证响应式）
const navItems = computed<NavItem[]>(() => [
  { key: 'site', label: t('system.nav.site'), icon: GlobalOutlined, component: SystemSiteView },
  { key: 'module', label: t('menu.system.module'), icon: AppstoreOutlined, component: ModuleConfigView },
  { key: 'storage', label: t('menu.storage.config'), icon: CloudServerOutlined, component: StorageConfigView },
  { key: 'cors', label: t('menu.profile.cors'), icon: ApiOutlined, component: CorsConfigView },
  { key: 'map', label: t('menu.system.map'), icon: KeyOutlined, component: MapConfigView },
  { key: 'security', label: t('menu.security.blacklist'), icon: SafetyCertificateOutlined, component: SecurityConfigView },
  { key: 'community', label: t('system.nav.community'), icon: SendOutlined, component: CommunityPushConfigView },
])

function readStoredKey(): string {
  try {
    const v = localStorage.getItem(ACTIVE_KEY_STORAGE)
    if (v && navItems.value.some((item) => item.key === v)) return v
  } catch {
    // 存储不可用时回退到默认项
  }
  return navItems.value[0].key
}

const route = useRoute()

const activeKey = ref(readStoredKey())

// 支持 ?tab= 深链定位（如主题页登录弹窗的官方地址修改引导），合法时优先于本地记忆
const queryTab = typeof route.query.tab === 'string' ? route.query.tab : ''
if (navItems.value.some((item) => item.key === queryTab)) {
  activeKey.value = queryTab
}

watch(activeKey, (key) => {
  try {
    localStorage.setItem(ACTIVE_KEY_STORAGE, key)
  } catch {
    // 持久化失败不影响切换
  }
})

const activeComponent = computed(() => {
  return navItems.value.find((item) => item.key === activeKey.value)?.component ?? navItems.value[0].component
})

// ---------- 配置导出 / 导入 ----------

const exporting = ref(false)

async function handleExportMenu({ key }: { key: string | number }) {
  exporting.value = true
  try {
    await exportSystemConfigAPI(key === 'with-secrets')
    message.success(t('system.export.success'))
  } catch {
    // 错误在 exportSystemConfigAPI 内统一提示
  } finally {
    exporting.value = false
  }
}

const importModalOpen = ref(false)
const importing = ref(false)
const importPayload = ref<SystemConfigExportRes | null>(null)
const importResult = ref<Record<string, number> | null>(null)
// refreshTick 导入成功后递增，强制重渲染当前子视图以重新加载配置
const refreshTick = ref(0)

const importGroupNames = computed(() => Object.keys(importPayload.value?.groups ?? {}))

// handleImportFile 阻止 a-upload 自动上传，仅读取本地 JSON 文本并弹确认框
const handleImportFile: UploadProps['beforeUpload'] = (file) => {
  file
    .text()
    .then((text) => {
      const parsed = JSON.parse(text) as SystemConfigExportRes
      if (!parsed || typeof parsed !== 'object' || !parsed.groups || typeof parsed.groups !== 'object') {
        throw new Error('invalid export file')
      }
      importPayload.value = parsed
      importResult.value = null
      importModalOpen.value = true
    })
    .catch(() => {
      message.error(t('system.import.invalidFile'))
    })
  return false
}

async function confirmImport() {
  if (!importPayload.value) return
  importing.value = true
  try {
    const res = await importSystemConfigAPI(importPayload.value)
    importResult.value = res.applied
    refreshTick.value++
  } catch {
    // 错误由拦截器统一提示
  } finally {
    importing.value = false
  }
}

function closeImportModal() {
  importModalOpen.value = false
  importPayload.value = null
  importResult.value = null
  importing.value = false
}
</script>

<style scoped>
.page-container {
  padding: 24px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #29365C;
  margin: 0;
}

.page-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-actions__caret {
  font-size: 12px;
}

.import-groups-label {
  margin: 0 0 8px 0;
  color: #29365C;
  font-weight: 500;
}

.import-groups {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 0;
  margin-bottom: 12px;
}

.import-applied {
  margin: 0;
  padding-left: 4px;
  list-style: none;
}

.import-applied li {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 6px 8px;
  border-bottom: 1px solid #F5F7FC;
}

.import-applied li:last-child {
  border-bottom: none;
}

.import-applied__group {
  font-weight: 500;
  color: #29365C;
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
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

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
