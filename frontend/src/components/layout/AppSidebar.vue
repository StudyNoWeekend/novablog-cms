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
          <router-link
            to="/comments"
            class="nav-item"
            :class="{ active: isActive('/comments') }"
          >
            <span class="nav-icon"><MessageOutlined /></span>
            <span class="nav-text">评论管理</span>
          </router-link>
        </div>

        <!-- 分组菜单（点击标题折叠/展开） -->
        <div v-for="group in menuGroups" :key="group.title" class="menu-group">
          <button
            type="button"
            class="menu-group-title"
            :aria-expanded="isGroupExpanded(group.title)"
            @click="toggleGroup(group.title)"
          >
            <span class="menu-group-title__text">{{ group.title }}</span>
            <CaretRightOutlined
              class="menu-group-title__arrow"
              :class="{ expanded: isGroupExpanded(group.title) }"
            />
          </button>
          <template v-if="isGroupExpanded(group.title)">
            <template v-for="item in group.items" :key="item.label">
              <!-- 二级子菜单 -->
              <template v-if="item.children">
                <button
                  type="button"
                  class="submenu-title"
                  :aria-expanded="isSubmenuExpanded(item.label)"
                  @click="toggleSubmenu(item.label)"
                >
                  <span class="nav-icon">
                    <component :is="item.icon" />
                  </span>
                  <span class="nav-text">{{ item.label }}</span>
                  <CaretRightOutlined
                    class="submenu-arrow"
                    :class="{ expanded: isSubmenuExpanded(item.label) }"
                  />
                </button>
                <template v-if="isSubmenuExpanded(item.label)">
                  <router-link
                    v-for="child in item.children"
                    :key="child.path"
                    :to="child.path"
                    class="nav-item submenu-item"
                    :class="{ active: route.path === child.path }"
                  >
                    <span class="nav-icon">
                      <component :is="child.icon" />
                    </span>
                    <span class="nav-text">{{ child.label }}</span>
                  </router-link>
                </template>
              </template>
              <!-- 普通菜单项 -->
              <router-link
                v-else
                :to="item.path"
                class="nav-item"
                :class="{ active: isMenuActive(item.path) }"
              >
                <span class="nav-icon">
                  <component :is="item.icon" />
                </span>
                <span class="nav-text">{{ item.label }}</span>
              </router-link>
            </template>
          </template>
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
  CloseOutlined,
  CaretRightOutlined,
  BookOutlined,
  SafetyCertificateOutlined,
  GithubOutlined,
  CoffeeOutlined,
  ReadOutlined,
  RocketOutlined,
  HeartOutlined,
  CodeOutlined,
  EyeOutlined,
  UserOutlined,
  CloudServerOutlined,
  SettingOutlined,
  AppstoreOutlined,
  ApiOutlined,
  KeyOutlined,
} from '@ant-design/icons-vue'
import AppLogo from '@/components/common/AppLogo.vue'
import { sidebarStorage } from '@/utils/storage'

const route = useRoute()
const appStore = useAppStore()
const moduleStore = useModuleStore()
const isMobile = ref(false)

interface SidebarMenuBase {
  label: string
  icon: Component
  moduleKey?: ModuleKey
}

interface SidebarMenuLink extends SidebarMenuBase {
  path: string
  children?: undefined
}

interface SidebarMenuSub extends SidebarMenuBase {
  /** 有 children 时渲染为二级子菜单（子项不做模块开关过滤） */
  children: { path: string; label: string; icon: Component }[]
}

type SidebarMenuItem = SidebarMenuLink | SidebarMenuSub

const menuGroupDefs: { title: string; items: SidebarMenuItem[] }[] = [
  {
    title: '关于我',
    items: [
      { path: '/profile/info', label: '个人资料', icon: UserOutlined },
      { path: '/equipments', label: '个人设备', icon: LaptopOutlined, moduleKey: 'equipment_enabled' },
      { path: '/projects', label: '项目经历', icon: ProjectOutlined, moduleKey: 'project_enabled' },
      { path: '/tech-stacks', label: '技术栈', icon: CodeOutlined, moduleKey: 'tech_stack_enabled' },
      { path: '/open-sources', label: '开源作品', icon: GithubOutlined, moduleKey: 'open_source_enabled' },
    ],
  },
  {
    title: '内容创作',
    items: [
      { path: '/articles', label: '文章管理', icon: FileTextOutlined, moduleKey: 'article_enabled' },
      { path: '/media', label: '媒体库', icon: PictureOutlined, moduleKey: 'media_enabled' },
      { path: '/playlists', label: '音乐播放列表', icon: CustomerServiceOutlined, moduleKey: 'music_enabled' },
      { path: '/videos', label: '视频作品', icon: PlaySquareOutlined, moduleKey: 'video_enabled' },
      { path: '/travels', label: '旅行攻略', icon: CompassOutlined, moduleKey: 'travel_enabled' },
      { path: '/portfolios', label: '摄影作品集', icon: CameraOutlined, moduleKey: 'portfolio_enabled' },
      { path: '/recipes', label: '美食菜谱', icon: CoffeeOutlined, moduleKey: 'recipe_enabled' },
      { path: '/books', label: '读书书架', icon: ReadOutlined, moduleKey: 'book_enabled' },
      { path: '/games', label: '游戏库', icon: RocketOutlined, moduleKey: 'game_enabled' },
      { path: '/fitness', label: '健身训练', icon: HeartOutlined, moduleKey: 'fitness_enabled' },
    ],
  },
  {
    title: '安全',
    items: [
      { path: '/api-doc', label: 'API 文档', icon: BookOutlined },
      { path: '/security/config', label: '黑名单管理', icon: SafetyCertificateOutlined },
      { path: '/security/monitor', label: '访问统计', icon: EyeOutlined },
    ],
  },
  {
    title: '系统',
    items: [
      { path: '/themes', label: '主题市场', icon: SkinOutlined },
      {
        label: '系统配置',
        icon: SettingOutlined,
        children: [
          { path: '/module-config', label: '模块管理', icon: AppstoreOutlined },
          { path: '/profile/storage', label: '对象存储', icon: CloudServerOutlined },
          { path: '/profile/cors', label: '跨域配置', icon: ApiOutlined },
          { path: '/map-config', label: '地图配置', icon: KeyOutlined },
        ],
      },
    ],
  },
]

// 按模块开关过滤：未配置开关的菜单项（评论、资料、安全、系统等）始终显示，空组整体隐藏
function visibleItems(items: SidebarMenuItem[]): SidebarMenuItem[] {
  return items.filter((item) => !item.moduleKey || moduleStore.isEnabled(item.moduleKey))
}

const menuGroups = computed(() =>
  menuGroupDefs
    .map((group) => ({ ...group, items: visibleItems(group.items) }))
    .filter((group) => group.items.length > 0)
)

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

// ---- 分组折叠 ----
// 侧边栏收窄为图标栏时标题已隐藏，此时不折叠（保留全部图标导航）
const collapsedGroups = ref<Record<string, boolean>>(sidebarStorage.getCollapsedGroups())

function isGroupExpanded(title: string): boolean {
  if (!isMobile.value && appStore.sidebarCollapsed) return true
  return !collapsedGroups.value[title]
}

function toggleGroup(title: string) {
  collapsedGroups.value = { ...collapsedGroups.value, [title]: !collapsedGroups.value[title] }
  sidebarStorage.setCollapsedGroups(collapsedGroups.value)
}

// ---- 二级子菜单折叠 ----
const expandedSubmenus = ref<Record<string, boolean>>(sidebarStorage.getExpandedSubmenus())

function isSubmenuExpanded(label: string): boolean {
  // 侧边栏收窄为图标栏时隐藏子项，展开侧边栏后恢复折叠状态
  if (!isMobile.value && appStore.sidebarCollapsed) return false
  return !!expandedSubmenus.value[label]
}

function toggleSubmenu(label: string) {
  expandedSubmenus.value = { ...expandedSubmenus.value, [label]: !expandedSubmenus.value[label] }
  sidebarStorage.setExpandedSubmenus(expandedSubmenus.value)
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 8px 16px 8px 20px;
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  transition: color var(--transition-fast);
}

.menu-group-title:hover {
  color: #4a6cf7;
}

.menu-group-title:focus-visible {
  outline: 2px solid #4a6cf7;
  outline-offset: -2px;
  border-radius: 6px;
}

.menu-group-title__arrow {
  font-size: 10px;
  color: #cbd5e1;
  transition: transform var(--transition-fast);
}

.menu-group-title__arrow:hover {
  color: inherit;
}

.menu-group-title__arrow.expanded {
  transform: rotate(90deg);
}

/* ---- 二级子菜单 ---- */
.submenu-title {
  display: flex;
  align-items: center;
  width: 100%;
  height: 42px;
  margin: 2px 12px;
  padding: 0 16px;
  border: none;
  border-radius: 6px;
  background: none;
  color: #1e293b;
  text-align: left;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
  overflow: hidden;
}

.submenu-title:hover {
  background: #e8ecf1;
}

.submenu-title:focus-visible {
  outline: 2px solid #4a6cf7;
  outline-offset: -2px;
}

.submenu-arrow {
  margin-left: auto;
  font-size: 10px;
  color: #94a3b8;
  transition: transform var(--transition-fast);
}

.submenu-arrow.expanded {
  transform: rotate(90deg);
}

.submenu-item {
  margin: 2px 12px 2px 26px;
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
  box-shadow: inset 3px 0 0 #4a6cf7;
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

.collapsed .sidebar-logo-wrapper {
  justify-content: center;
  padding: 0 8px;
}
</style>
