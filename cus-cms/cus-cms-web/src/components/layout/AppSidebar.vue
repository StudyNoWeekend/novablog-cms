<template>
  <aside
    class="app-sidebar"
    :class="{
      collapsed: !isMobile && appStore.sidebarCollapsed,
      'mobile-open': isMobile && !appStore.sidebarCollapsed,
      'mobile-closed': isMobile && appStore.sidebarCollapsed,
    }"
  >
    <div class="sidebar-top">
      <div class="sidebar-logo-wrapper">
        <span v-if="isMobile" class="close-btn" @click="appStore.sidebarCollapsed = true">
          <CloseOutlined />
        </span>
        <span v-else class="collapse-btn" @click="appStore.toggleSidebar">
          <MenuFoldOutlined v-if="!appStore.sidebarCollapsed" />
          <MenuUnfoldOutlined v-else />
        </span>
        <AppLogo :collapsed="!isMobile && appStore.sidebarCollapsed" />
      </div>

      <nav class="sidebar-nav">
        <!-- 工作台 -->
        <div class="menu-group">
          <router-link
            to="/dashboard"
            class="nav-item"
            :class="{ active: isActive('/dashboard') }"
          >
            <span class="nav-icon"><DashboardOutlined /></span>
            <span class="nav-text">工作台</span>
          </router-link>
        </div>

        <!-- 内容管理（可展开） -->
        <div class="menu-group">
          <div
            class="nav-item submenu-title"
            :class="{ active: isContentActive }"
            @click="contentExpanded = !contentExpanded"
          >
            <span class="nav-icon"><FileTextOutlined /></span>
            <span class="nav-text">内容管理</span>
            <span class="submenu-arrow" :class="{ expanded: contentExpanded }">
              <RightOutlined />
            </span>
          </div>
          <div class="submenu" :class="{ expanded: contentExpanded }">
            <router-link
              v-for="item in contentMenuItems"
              :key="item.path"
              :to="item.path"
              class="nav-item sub-item"
              :class="{ active: isActive(item.path) }"
            >
              <span class="nav-icon">
                <component :is="item.icon" />
              </span>
              <span class="nav-text">{{ item.label }}</span>
            </router-link>
          </div>
        </div>

        <!-- 媒体库 -->
        <div class="menu-group">
          <router-link
            to="/media"
            class="nav-item"
            :class="{ active: isActive('/media') }"
          >
            <span class="nav-icon"><PictureOutlined /></span>
            <span class="nav-text">媒体库</span>
          </router-link>
        </div>

        <!-- 评论管理 -->
        <div class="menu-group">
          <router-link
            to="/comments"
            class="nav-item"
            :class="{ active: isActive('/comments') }"
          >
            <span class="nav-icon"><MessageOutlined /></span>
            <span class="nav-text">评论管理</span>
          </router-link>
        </div>

        <!-- API 文档 -->
        <div class="menu-group">
          <router-link
            to="/api-doc"
            class="nav-item"
            :class="{ active: isActive('/api-doc') }"
          >
            <span class="nav-icon"><BookOutlined /></span>
            <span class="nav-text">API 文档</span>
          </router-link>
        </div>

        <!-- 安全中心（可展开） -->
        <div class="menu-group">
          <div
            class="nav-item submenu-title"
            :class="{ active: isSecurityActive }"
            @click="securityExpanded = !securityExpanded"
          >
            <span class="nav-icon"><SafetyOutlined /></span>
            <span class="nav-text">安全中心</span>
            <span class="submenu-arrow" :class="{ expanded: securityExpanded }">
              <RightOutlined />
            </span>
          </div>
          <div class="submenu" :class="{ expanded: securityExpanded }">
            <router-link
              v-for="item in securityMenuItems"
              :key="item.path"
              :to="item.path"
              class="nav-item sub-item"
              :class="{ active: isActiveExact(item.path) }"
            >
              <span class="nav-icon">
                <component :is="item.icon" />
              </span>
              <span class="nav-text">{{ item.label }}</span>
            </router-link>
          </div>
        </div>

        <!-- 模板风格 -->
        <div class="menu-group">
          <router-link
            to="/templates"
            class="nav-item"
            :class="{ active: isActive('/templates') }"
          >
            <span class="nav-icon"><SkinOutlined /></span>
            <span class="nav-text">模板风格</span>
          </router-link>
        </div>

        <!-- 设置 -->
        <div class="menu-group">
          <div class="menu-group-title">系统</div>
          <router-link
            to="/profile/info"
            class="nav-item"
            :class="{ active: isActive('/profile') }"
          >
            <span class="nav-icon"><SettingOutlined /></span>
            <span class="nav-text">设置</span>
          </router-link>
          <a class="nav-item" @click.prevent="">
            <span class="nav-icon"><QuestionCircleOutlined /></span>
            <span class="nav-text">帮助</span>
          </a>
        </div>
      </nav>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import {
  DashboardOutlined,
  FileTextOutlined,
  PictureOutlined,
  CameraOutlined,
  PlaySquareOutlined,
  CompassOutlined,
  CustomerServiceOutlined,
  MessageOutlined,
  SkinOutlined,
  SettingOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  QuestionCircleOutlined,
  CloseOutlined,
  RightOutlined,
  BookOutlined,
  SafetyOutlined,
  SafetyCertificateOutlined,
  EyeOutlined,
} from '@ant-design/icons-vue'
import AppLogo from '@/components/common/AppLogo.vue'

const route = useRoute()
const appStore = useAppStore()
const isMobile = ref(false)
const contentExpanded = ref(true)
const securityExpanded = ref(false)

const contentMenuItems = [
  { path: '/articles', label: '文章管理', icon: FileTextOutlined },
  { path: '/portfolios', label: '摄影作品集', icon: CameraOutlined },
  { path: '/videos', label: '视频作品', icon: PlaySquareOutlined },
  { path: '/travels', label: '旅行攻略', icon: CompassOutlined },
  { path: '/playlists', label: '音乐播放列表', icon: CustomerServiceOutlined },
]

const securityMenuItems = [
  { path: '/security/config', label: '安全配置', icon: SafetyCertificateOutlined },
  { path: '/security/monitor', label: '安全监控', icon: EyeOutlined },
]

const securityPaths = securityMenuItems.map((item) => item.path)

const contentPaths = contentMenuItems.map((item) => item.path)

const isContentActive = computed(() => {
  const currentRoot = '/' + route.path.split('/')[1]
  return contentPaths.includes(currentRoot)
})

const isSecurityActive = computed(() => {
  return securityPaths.some((p) => route.path.startsWith(p))
})

function isActive(path: string): boolean {
  const currentRoot = '/' + route.path.split('/')[1]
  if (path === '/dashboard') return currentRoot === '/dashboard'
  return currentRoot === path
}

function isActiveExact(path: string): boolean {
  return route.path === path
}

function handleResize() {
  const mobile = window.innerWidth <= 768
  if (mobile && !isMobile.value) {
    appStore.sidebarCollapsed = true
  } else if (!mobile && isMobile.value) {
    appStore.sidebarCollapsed = false
  }
  isMobile.value = mobile
}

onMounted(() => {
  handleResize()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.app-sidebar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 220px;
  background: #f0f2f5;
  display: flex;
  flex-direction: column;
  z-index: 100;
  overflow: hidden;
  transition: width var(--transition-slow);
  border-right: 1px solid #e8e8e8;
}

.app-sidebar.collapsed {
  width: 72px;
}

@media (max-width: 768px) {
  .app-sidebar {
    width: 260px;
    transition: transform var(--transition-slow);
    z-index: 200;
  }

  .app-sidebar.mobile-closed {
    transform: translateX(-100%);
  }

  .app-sidebar.mobile-open {
    transform: translateX(0);
  }
}

.sidebar-top {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sidebar-logo-wrapper {
  height: 72px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 12px;
  border-bottom: 1px solid #e8e8e8;
  flex-shrink: 0;
}

.collapse-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #64748b;
  cursor: pointer;
  transition: color var(--transition-fast);
  flex-shrink: 0;
}

.collapse-btn:hover {
  color: #4a6cf7;
}

.close-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #64748b;
  cursor: pointer;
  transition: color var(--transition-fast);
  flex-shrink: 0;
}

.close-btn:hover {
  color: #4a6cf7;
}

.sidebar-nav {
  flex: 1;
  padding: 12px 0;
  overflow-y: auto;
  overflow-x: hidden;
}

.menu-group {
  margin-bottom: 4px;
}

.menu-group-title {
  padding: 8px 20px;
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.nav-item {
  display: flex;
  align-items: center;
  height: 42px;
  margin: 2px 12px;
  padding: 0 16px;
  border-radius: 6px;
  color: #1e293b;
  text-decoration: none;
  transition: all var(--transition-fast);
  white-space: nowrap;
  overflow: hidden;
  cursor: pointer;
}

.nav-item:hover {
  background: #e8ecf1;
}

.nav-item.active {
  color: #4a6cf7;
  background: rgba(74, 108, 247, 0.08);
  border-left: 3px solid #4a6cf7;
  padding-left: 13px;
  border-radius: 0 6px 6px 0;
  margin-left: 9px;
}

/* 子菜单标题 */
.submenu-title {
  font-weight: 500;
}

.submenu-arrow {
  margin-left: auto;
  font-size: 10px;
  color: #94a3b8;
  transition: transform var(--transition-fast);
  display: inline-flex;
  align-items: center;
}

.submenu-arrow.expanded {
  transform: rotate(90deg);
}

/* 子菜单容器 */
.submenu {
  max-height: 0;
  overflow: hidden;
  transition: max-height var(--transition-slow);
}

.submenu.expanded {
  max-height: 300px;
}

/* 子菜单项 */
.sub-item {
  height: 38px;
  padding-left: 24px;
  margin: 1px 12px;
  font-size: 13px;
}

.sub-item .nav-icon {
  font-size: 14px;
  opacity: 0.7;
}

.nav-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  min-width: 18px;
  flex-shrink: 0;
}

.nav-text {
  margin-left: 12px;
  font-size: 14px;
  opacity: 1;
  transition: opacity var(--transition-slow);
}

.collapsed .nav-text,
.collapsed .menu-group-title {
  opacity: 0;
  width: 0;
  margin-left: 0;
  display: none;
}

.collapsed .nav-item {
  justify-content: center;
  padding: 0;
  margin: 2px 8px;
}

.collapsed .nav-item.active {
  padding-left: 0;
  margin-left: 8px;
  border-left: none;
  border-radius: 6px;
}

.collapsed .submenu {
  max-height: 0 !important;
}

.collapsed .submenu-arrow {
  display: none;
}

.collapsed .sidebar-logo-wrapper {
  justify-content: center;
  padding: 0 8px;
}
</style>
