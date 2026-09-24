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

        <!-- 通用 -->
        <div class="menu-group">
          <div class="menu-group-title">通用</div>
          <router-link
            v-for="item in generalMenuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isMenuActive(item.path) }"
          >
            <span class="nav-icon">
              <component :is="item.icon" />
            </span>
            <span class="nav-text">{{ item.label }}</span>
          </router-link>
        </div>

        <!-- 音乐 -->
        <div v-if="musicMenuItems.length" class="menu-group">
          <div class="menu-group-title">音乐</div>
          <router-link
            v-for="item in musicMenuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isMenuActive(item.path) }"
          >
            <span class="nav-icon">
              <component :is="item.icon" />
            </span>
            <span class="nav-text">{{ item.label }}</span>
          </router-link>
        </div>

        <!-- 视频分享 -->
        <div v-if="videoMenuItems.length" class="menu-group">
          <div class="menu-group-title">视频分享</div>
          <router-link
            v-for="item in videoMenuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isMenuActive(item.path) }"
          >
            <span class="nav-icon">
              <component :is="item.icon" />
            </span>
            <span class="nav-text">{{ item.label }}</span>
          </router-link>
        </div>

        <!-- 旅行分享 -->
        <div v-if="travelMenuItems.length" class="menu-group">
          <div class="menu-group-title">旅行分享</div>
          <router-link
            v-for="item in travelMenuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isMenuActive(item.path) }"
          >
            <span class="nav-icon">
              <component :is="item.icon" />
            </span>
            <span class="nav-text">{{ item.label }}</span>
          </router-link>
        </div>

        <!-- 摄影 -->
        <div v-if="photoMenuItems.length" class="menu-group">
          <div class="menu-group-title">摄影</div>
          <router-link
            v-for="item in photoMenuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isMenuActive(item.path) }"
          >
            <span class="nav-icon">
              <component :is="item.icon" />
            </span>
            <span class="nav-text">{{ item.label }}</span>
          </router-link>
        </div>

        <!-- 技术 -->
        <div v-if="techMenuItems.length" class="menu-group">
          <div class="menu-group-title">技术</div>
          <router-link
            v-for="item in techMenuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isMenuActive(item.path) }"
          >
            <span class="nav-icon">
              <component :is="item.icon" />
            </span>
            <span class="nav-text">{{ item.label }}</span>
          </router-link>
        </div>

        <!-- 安全 -->
        <div class="menu-group">
          <div class="menu-group-title">安全</div>
          <router-link
            v-for="item in securityMenuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isMenuActive(item.path) }"
          >
            <span class="nav-icon">
              <component :is="item.icon" />
            </span>
            <span class="nav-text">{{ item.label }}</span>
          </router-link>
        </div>

        <!-- 媒体库 -->
        <div v-if="moduleStore.isEnabled('media_enabled')" class="menu-group">
          <router-link
            to="/media"
            class="nav-item"
            :class="{ active: isActive('/media') }"
          >
            <span class="nav-icon"><PictureOutlined /></span>
            <span class="nav-text">媒体库</span>
          </router-link>
        </div>

        <!-- 系统 -->
        <div class="menu-group">
          <div class="menu-group-title">系统</div>
          <router-link
            v-for="item in systemMenuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isMenuActive(item.path) }"
          >
            <span class="nav-icon">
              <component :is="item.icon" />
            </span>
            <span class="nav-text">{{ item.label }}</span>
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
import { computed, ref, onMounted, onUnmounted } from 'vue'
import type { Component } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useModuleStore, type ModuleKey } from '@/stores/module'
import {
  DashboardOutlined,
  FileTextOutlined,
  PictureOutlined,
  CameraOutlined,
  LaptopOutlined,
  ProjectOutlined,
  PlaySquareOutlined,
  CompassOutlined,
  CustomerServiceOutlined,
  MessageOutlined,
  SkinOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  QuestionCircleOutlined,
  CloseOutlined,
  BookOutlined,
  SafetyCertificateOutlined,
  GithubOutlined,
	  EyeOutlined,
	  UserOutlined,
	  CloudServerOutlined,
	  SettingOutlined,
	} from '@ant-design/icons-vue'
import AppLogo from '@/components/common/AppLogo.vue'

const route = useRoute()
const appStore = useAppStore()
const moduleStore = useModuleStore()
const isMobile = ref(false)

interface SidebarMenuItem {
  path: string
  label: string
  icon: Component
  moduleKey?: ModuleKey
}

const allGeneralMenuItems: SidebarMenuItem[] = [
  { path: '/articles', label: '文章管理', icon: FileTextOutlined, moduleKey: 'article_enabled' },
  { path: '/comments', label: '评论管理', icon: MessageOutlined },
  { path: '/profile/info', label: '个人资料', icon: UserOutlined },
  { path: '/equipments', label: '个人设备', icon: LaptopOutlined, moduleKey: 'equipment_enabled' },
  { path: '/projects', label: '项目经历', icon: ProjectOutlined, moduleKey: 'project_enabled' },
]

const allMusicMenuItems: SidebarMenuItem[] = [
  { path: '/playlists', label: '音乐播放列表', icon: CustomerServiceOutlined, moduleKey: 'music_enabled' },
]

const allVideoMenuItems: SidebarMenuItem[] = [
  { path: '/videos', label: '视频作品', icon: PlaySquareOutlined, moduleKey: 'video_enabled' },
]

const allTravelMenuItems: SidebarMenuItem[] = [
  { path: '/travels', label: '旅行攻略', icon: CompassOutlined, moduleKey: 'travel_enabled' },
]

const allPhotoMenuItems: SidebarMenuItem[] = [
  { path: '/portfolios', label: '摄影作品集', icon: CameraOutlined, moduleKey: 'portfolio_enabled' },
]

const allTechMenuItems: SidebarMenuItem[] = [
  { path: '/open-sources', label: '开源作品', icon: GithubOutlined, moduleKey: 'open_source_enabled' },
]

// 按模块开关过滤：未配置开关的菜单项（评论、资料、安全、系统等）始终显示
function visibleItems(items: SidebarMenuItem[]): SidebarMenuItem[] {
  return items.filter((item) => !item.moduleKey || moduleStore.isEnabled(item.moduleKey))
}

const generalMenuItems = computed(() => visibleItems(allGeneralMenuItems))
const musicMenuItems = computed(() => visibleItems(allMusicMenuItems))
const videoMenuItems = computed(() => visibleItems(allVideoMenuItems))
const travelMenuItems = computed(() => visibleItems(allTravelMenuItems))
const photoMenuItems = computed(() => visibleItems(allPhotoMenuItems))
const techMenuItems = computed(() => visibleItems(allTechMenuItems))

const securityMenuItems = [
  { path: '/api-doc', label: 'API 文档', icon: BookOutlined },
  { path: '/security/config', label: '黑名单管理', icon: SafetyCertificateOutlined },
  { path: '/security/monitor', label: '访问统计', icon: EyeOutlined },
]

		const systemMenuItems = [
		  { path: '/themes', label: '主题市场', icon: SkinOutlined },
		  { path: '/module-config', label: '模块管理', icon: SettingOutlined },
		  { path: '/profile/storage', label: '对象存储', icon: CloudServerOutlined },
		  { path: '/profile/cors', label: '跨域配置', icon: SettingOutlined },
		]

function isActive(path: string): boolean {
  const currentRoot = '/' + route.path.split('/')[1]
  if (path === '/dashboard') return currentRoot === '/dashboard'
  return currentRoot === path
}

function isMenuActive(path: string): boolean {
  if (path.startsWith('/security/') || path.startsWith('/profile/')) {
    return route.path === path
  }
  return isActive(path)
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
  moduleStore.fetchConfig()
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

.collapsed .sidebar-logo-wrapper {
  justify-content: center;
  padding: 0 8px;
}
</style>
