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
            <span class="nav-text">{{ $t('menu.dashboard') }}</span>
          </router-link>
          <router-link
            to="/comments"
            class="nav-item"
            :class="{ active: isActive('/comments') }"
          >
            <span class="nav-icon"><MessageOutlined /></span>
            <span class="nav-text">{{ $t('menu.comment.list') }}</span>
          </router-link>
          <router-link
            to="/media"
            class="nav-item"
            :class="{ active: isActive('/media') }"
          >
            <span class="nav-icon"><PictureOutlined /></span>
            <span class="nav-text">{{ $t('menu.media.list') }}</span>
          </router-link>
        </div>

        <!-- 分组菜单（点击标题折叠/展开） -->
        <div v-for="group in menuGroups" :key="group.titleKey" class="menu-group">
          <button
            type="button"
            class="menu-group-title"
            :aria-expanded="isGroupExpanded(group.titleKey)"
            @click="toggleGroup(group.titleKey)"
          >
            <span class="menu-group-title__text">{{ $t(group.titleKey) }}</span>
            <CaretRightOutlined
              class="menu-group-title__arrow"
              :class="{ expanded: isGroupExpanded(group.titleKey) }"
            />
          </button>
          <template v-if="isGroupExpanded(group.titleKey)">
          <!-- 普通菜单项 -->
          <router-link
            v-for="item in group.items"
            :key="item.labelKey"
            :to="item.path"
            class="nav-item"
            :class="{ active: isMenuActive(item.path) }"
          >
            <span class="nav-icon">
              <component :is="item.icon" />
            </span>
            <span class="nav-text">{{ $t(item.labelKey) }}</span>
          </router-link>
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
  ApiOutlined,
  GithubOutlined,
  CoffeeOutlined,
  ReadOutlined,
  RocketOutlined,
  HeartOutlined,
  CodeOutlined,
  CloudOutlined,
  EyeOutlined,
  UserOutlined,
  SettingOutlined,
} from '@ant-design/icons-vue'
import AppLogo from '@/components/common/AppLogo.vue'
import { sidebarStorage } from '@/utils/storage'

const route = useRoute()
const appStore = useAppStore()
const moduleStore = useModuleStore()
const isMobile = ref(false)

interface SidebarMenuBase {
  labelKey: string
  icon: Component
  moduleKey?: ModuleKey
}

interface SidebarMenuLink extends SidebarMenuBase {
  path: string
}

type SidebarMenuItem = SidebarMenuLink

// labelKey/titleKey 为 menu.* 的 i18n key；折叠状态持久化以 titleKey 为身份（跨语言稳定）
const menuGroupDefs: { titleKey: string; items: SidebarMenuItem[] }[] = [
  {
    titleKey: 'menu.group.about',
    items: [
      { path: '/profile/info', labelKey: 'menu.profile.info', icon: UserOutlined },
      { path: '/equipments', labelKey: 'menu.equipment.list', icon: LaptopOutlined, moduleKey: 'equipment_enabled' },
      { path: '/projects', labelKey: 'menu.project.list', icon: ProjectOutlined, moduleKey: 'project_enabled' },
      { path: '/tech-stacks', labelKey: 'menu.techStack.list', icon: CodeOutlined, moduleKey: 'tech_stack_enabled' },
      { path: '/open-sources', labelKey: 'menu.openSource.list', icon: GithubOutlined, moduleKey: 'open_source_enabled' },
    ],
  },
  {
    titleKey: 'menu.group.content',
    items: [
      { path: '/articles', labelKey: 'menu.article.list', icon: FileTextOutlined, moduleKey: 'article_enabled' },
      { path: '/playlists', labelKey: 'menu.playlist.list', icon: CustomerServiceOutlined, moduleKey: 'music_enabled' },
      { path: '/videos', labelKey: 'menu.video.list', icon: PlaySquareOutlined, moduleKey: 'video_enabled' },
      { path: '/travels', labelKey: 'menu.travel.list', icon: CompassOutlined, moduleKey: 'travel_enabled' },
      { path: '/portfolios', labelKey: 'menu.portfolio.list', icon: CameraOutlined, moduleKey: 'portfolio_enabled' },
      { path: '/recipes', labelKey: 'menu.recipe.list', icon: CoffeeOutlined, moduleKey: 'recipe_enabled' },
      { path: '/books', labelKey: 'menu.book.list', icon: ReadOutlined, moduleKey: 'book_enabled' },
      { path: '/games', labelKey: 'menu.game.list', icon: RocketOutlined, moduleKey: 'game_enabled' },
      { path: '/fitness', labelKey: 'menu.fitness.list', icon: HeartOutlined, moduleKey: 'fitness_enabled' },
    ],
  },
  {
    titleKey: 'menu.group.system',
    items: [
      { path: '/themes', labelKey: 'menu.theme.market', icon: SkinOutlined },
      { path: '/cloud-drive', labelKey: 'menu.cloudDrive.accounts', icon: CloudOutlined },
      { path: '/system-config', labelKey: 'menu.system.config', icon: SettingOutlined },
      { path: '/security/monitor', labelKey: 'menu.security.monitor', icon: EyeOutlined },
    ],
  },
  {
    titleKey: 'menu.group.ai',
    items: [
      { path: '/api-doc', labelKey: 'menu.apiDoc.list', icon: BookOutlined },
      { path: '/mcp', labelKey: 'menu.mcp.list', icon: ApiOutlined },
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

function isGroupExpanded(titleKey: string): boolean {
  if (!isMobile.value && appStore.sidebarCollapsed) return true
  return !collapsedGroups.value[titleKey]
}

function toggleGroup(titleKey: string) {
  collapsedGroups.value = { ...collapsedGroups.value, [titleKey]: !collapsedGroups.value[titleKey] }
  sidebarStorage.setCollapsedGroups(collapsedGroups.value)
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
  width: var(--sidebar-width);
  background: var(--bg-sidebar);
  display: flex;
  flex-direction: column;
  z-index: 100;
  overflow: hidden;
  transition: width var(--transition-slow);
  border-right: 1px solid #E5E9F2;
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
  border-bottom: 1px solid #E5E9F2;
  flex-shrink: 0;
}

.collapse-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #667085;
  cursor: pointer;
  transition: color var(--transition-fast);
  flex-shrink: 0;
}

.collapse-btn:hover {
  color: #526FE8;
}

.close-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #667085;
  cursor: pointer;
  transition: color var(--transition-fast);
  flex-shrink: 0;
}

.close-btn:hover {
  color: #526FE8;
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
  color: #8A93A8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  transition: color var(--transition-fast);
}

.menu-group-title:hover {
  color: #526FE8;
}

.menu-group-title:focus-visible {
  outline: 2px solid #526FE8;
  outline-offset: -2px;
  border-radius: 6px;
}

.menu-group-title__arrow {
  font-size: 10px;
  color: #8A93A8;
  transition: transform var(--transition-fast);
}

.menu-group-title__arrow:hover {
  color: inherit;
}

.menu-group-title__arrow.expanded {
  transform: rotate(90deg);
}

.nav-item {
  display: flex;
  align-items: center;
  height: 42px;
  margin: 2px 12px;
  padding: 0 16px;
  border-radius: 6px;
  color: #29365C;
  text-decoration: none;
  transition: all var(--transition-fast);
  white-space: nowrap;
  overflow: hidden;
  cursor: pointer;
}

.nav-item:hover {
  background: var(--bg-sidebar-hover);
}

.nav-item.active {
  color: #526FE8;
  background: var(--color-primary-soft);
  box-shadow: inset 3px 0 0 #526FE8;
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
