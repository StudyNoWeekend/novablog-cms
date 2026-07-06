# Cus CMS 前端设计文档

## 一、项目概述

### 1.1 定位

`cus-cms-web` 是 Cus 博客系统的内容管理后台前端，面向博主和管理员，提供博客内容创作、媒体管理、作品展示、评论审核、模板配置等全部管理能力。

### 1.2 技术栈

| 层级 | 技术选型 | 版本 | 说明 |
|------|----------|------|------|
| 框架 | Vue 3 | 3.4+ | Composition API + `<script setup>` |
| 构建工具 | Vite | 5.x | 极速开发体验 |
| UI 组件库 | Ant Design Vue | 4.x | 企业级后台UI组件 |
| 状态管理 | Pinia | 2.x | Vue 3 官方推荐 |
| 路由 | Vue Router | 4.x | SPA 路由 |
| HTTP 客户端 | Axios | 1.x | 请求封装 + 拦截器 |
| 富文本编辑器 | Tiptap | 2.x | 基于 ProseMirror 的现代编辑器 |
| Markdown 编辑器 | ByteMD | 1.x | 字节跳动开源 Markdown 编辑器 |
| 图表 | ECharts | 5.x | 数据统计可视化 |
| 图标 | @ant-design/icons-vue | 7.x | Ant Design 图标库 |
| CSS 方案 | UnoCSS | 0.58+ | 原子化 CSS + 主题定制 |
| 文件上传 | 自封装 + 分片 | - | 大文件分片上传 |
| 国际化 | vue-i18n | 9.x | 预留 |

---

## 二、设计系统

### 2.1 设计理念

**"创作者的暗房"** —— 正如摄影师在暗房中精心打磨每一张作品，CMS 后台是创作者雕琢内容的工作空间。设计语言追求：

- **沉浸感**：深色侧边栏 + 浅色工作区，将注意力聚焦于内容创作
- **呼吸感**：充裕的留白让每个模块清晰可辨，避免后台拥挤感
- **精准反馈**：每一次操作都有细腻的动效反馈，让创作过程流畅愉悦
- **职业调性**：区别于通用后台的千篇一律，注入创作者工具的独特气质

### 2.2 色彩系统

```
┌─ 品牌色（墨蓝）
│  Primary:    #1a1a2e   →  深邃夜幕，用于侧边栏背景
│  Accent:     #4a6cf7   →  电光蓝，用于操作按钮、选中态
│  AccentAlt:  #7c3aed   →  紫罗兰，用于强调、高亮
│
├─ 功能色
│  Success:    #10b981   绿
│  Warning:    #f59e0b   琥珀
│  Error:      #ef4444   赤红
│  Info:       #3b82f6   蓝
│
├─ 中性色
│  BG-Primary:    #f8f9fb   页面底色
│  BG-Secondary:  #ffffff   卡片/容器白
│  BG-Sidebar:    #111827   侧边栏深色
│  BG-SidebarHov: #1e293b   侧边栏悬停
│  Text-Primary:  #1e293b   主文字
│  Text-Secondary:#64748b   次文字
│  Text-Disabled: #94a3b8   禁用文字
│  Border:        #e2e8f0   边框
│  Divider:       #f1f5f9   分割线
│
└─ 数据可视化色盘
   #4a6cf7 #7c3aed #10b981 #f59e0b #ef4444 #ec4899 #06b6d4
```

**CSS 变量定义**：

```css
:root {
  /* 品牌色 */
  --color-primary: #4a6cf7;
  --color-primary-hover: #3b5de7;
  --color-primary-light: rgba(74, 108, 247, 0.08);
  --color-accent: #7c3aed;

  /* 背景 */
  --bg-page: #f8f9fb;
  --bg-card: #ffffff;
  --bg-sidebar: #111827;
  --bg-sidebar-hover: #1e293b;
  --bg-sidebar-active: rgba(74, 108, 247, 0.15);

  /* 文字 */
  --text-primary: #1e293b;
  --text-secondary: #64748b;
  --text-tertiary: #94a3b8;
  --text-sidebar: #cbd5e1;
  --text-sidebar-active: #ffffff;

  /* 边框 */
  --border-color: #e2e8f0;
  --border-radius: 8px;
  --border-radius-lg: 12px;

  /* 阴影 */
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.04);
  --shadow-card: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
  --shadow-dropdown: 0 4px 16px rgba(0, 0, 0, 0.08);
  --shadow-modal: 0 20px 60px rgba(0, 0, 0, 0.12);

  /* 间距 */
  --sidebar-width: 240px;
  --sidebar-collapsed: 72px;
  --header-height: 56px;

  /* 动效 */
  --transition-fast: 150ms cubic-bezier(0.4, 0, 0.2, 1);
  --transition-base: 250ms cubic-bezier(0.4, 0, 0.2, 1);
  --transition-slow: 350ms cubic-bezier(0.4, 0, 0.2, 1);
}
```

### 2.3 字体系统

| 用途 | 字体 | 备选 |
|------|------|------|
| 中文正文 | `"PingFang SC"` | `"Microsoft YaHei"` |
| 英文/数字 | `"Inter"` | `"SF Pro Display"` |
| 代码/编辑器 | `"JetBrains Mono"` | `"Fira Code"`, `"Consolas"` |
| Logo/标题 | `"PingFang SC"` | weight 600-700 |

```css
body {
  font-family: "Inter", "PingFang SC", "Microsoft YaHei", -apple-system, sans-serif;
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* 代码块 */
code, pre, .markdown-editor {
  font-family: "JetBrains Mono", "Fira Code", "Consolas", monospace;
}
```

### 2.4 布局系统

```
┌──────────────────────────────────────────────────────────────┐
│  ┌──────────┐  ┌───────────────────────────────────────────┐ │
│  │          │  │  Header (56px)                             │ │
│  │          │  │  ┌───────────────────────────────────────┐│ │
│  │ Sidebar  │  │  │                                       ││ │
│  │ 240px    │  │  │  Content Area                          ││ │
│  │          │  │  │  (max-width: 1400px centered)          ││ │
│  │  Logo    │  │  │                                       ││ │
│  │  Nav     │  │  │                                       ││ │
│  │  Menu    │  │  │                                       ││ │
│  │  ...     │  │  │                                       ││ │
│  │          │  │  └───────────────────────────────────────┘│ │
│  │          │  └───────────────────────────────────────────┘ │
│  └──────────┘                                                │
└──────────────────────────────────────────────────────────────┘
```

### 2.5 动效定义

```css
/* 页面过渡 */
.page-enter-active { transition: opacity 0.3s, transform 0.3s; }
.page-enter-from { opacity: 0; transform: translateY(8px); }

/* 列表交错入场 */
.list-item { animation: fadeInUp 0.4s ease both; }
.list-item:nth-child(1) { animation-delay: 0.05s; }
.list-item:nth-child(2) { animation-delay: 0.1s; }
.list-item:nth-child(3) { animation-delay: 0.15s; }
/* ... 以此类推 */

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* 卡片悬浮 */
.card-hover {
  transition: transform var(--transition-base), box-shadow var(--transition-base);
}
.card-hover:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-dropdown);
}

/* 侧边栏折叠 */
.sidebar-collapse {
  transition: width var(--transition-slow);
}
```

---

## 三、项目目录结构

```
cus-cms-web/
├── public/
│   └── favicon.svg
├── src/
│   ├── assets/
│   │   ├── images/
│   │   │   ├── logo.svg
│   │   │   └── logo-symbol.svg        # 折叠后仅图标
│   │   └── styles/
│   │       ├── variables.css           # CSS 变量
│   │       ├── reset.css               # 基础重置
│   │       ├── global.css              # 全局样式
│   │       └── transition.css          # 动效定义
│   ├── components/
│   │   ├── common/                     # 通用组件
│   │   │   ├── AppLogo.vue
│   │   │   ├── AppBreadcrumb.vue
│   │   │   ├── AppAvatar.vue
│   │   │   ├── EmptyState.vue          # 空状态
│   │   │   ├── LoadingSkeleton.vue     # 骨架屏
│   │   │   ├── ConfirmModal.vue        # 确认弹窗
│   │   │   └── PageHeader.vue          # 页面标题栏
│   │   ├── layout/
│   │   │   ├── AppLayout.vue           # 主布局（侧边栏+顶栏+内容）
│   │   │   ├── AppSidebar.vue          # 侧边导航
│   │   │   ├── AppHeader.vue           # 顶部栏
│   │   │   └── AppContent.vue          # 内容区包装
│   │   ├── editor/
│   │   │   ├── MarkdownEditor.vue      # Markdown 编辑器
│   │   │   ├── RichTextEditor.vue      # 富文本编辑器
│   │   │   └── EditorToolbar.vue       # 编辑器工具栏
│   │   ├── media/
│   │   │   ├── MediaPicker.vue         # 媒体选择器（弹窗）
│   │   │   ├── MediaUploader.vue       # 文件上传组件
│   │   │   ├── ImagePreview.vue        # 图片预览
│   │   │   └── VideoPlayer.vue         # 视频播放器
│   │   ├── article/
│   │   │   ├── ArticleCard.vue         # 文章卡片
│   │   │   ├── ArticleForm.vue         # 文章编辑表单
│   │   │   ├── ArticleMeta.vue         # 文章元信息栏
│   │   │   └── CategorySelect.vue      # 分类选择器
│   │   └── portfolio/
│   │       ├── PortfolioCard.vue       # 作品集卡片
│   │       └── PhotoGrid.vue           # 照片网格
│   ├── views/
│   │   ├── auth/
│   │   │   ├── LoginView.vue
│   │   │   ├── RegisterView.vue
│   │   │   └── ForgotPasswordView.vue
│   │   ├── dashboard/
│   │   │   └── DashboardView.vue       # 仪表盘/数据概览
│   │   ├── article/
│   │   │   ├── ArticleListView.vue     # 文章列表
│   │   │   ├── ArticleCreateView.vue   # 新建文章
│   │   │   └── ArticleEditView.vue     # 编辑文章
│   │   ├── category/
│   │   │   └── CategoryManageView.vue  # 分类与标签管理
│   │   ├── media/
│   │   │   └── MediaLibraryView.vue    # 媒体库
│   │   ├── portfolio/
│   │   │   ├── PortfolioListView.vue   # 作品集列表
│   │   │   └── PortfolioDetailView.vue # 作品集详情/照片管理
│   │   ├── video/
│   │   │   └── VideoManageView.vue     # 视频管理
│   │   ├── comment/
│   │   │   └── CommentManageView.vue   # 评论管理
│   │   ├── template/
│   │   │   └── TemplateView.vue        # 模板选择与配置
│   │   ├── user/
│   │   │   ├── ProfileView.vue         # 个人资料
│   │   │   └── UserManageView.vue      # 用户管理（管理员）
│   │   ├── role/
│   │   │   └── RoleManageView.vue      # 角色权限管理（管理员）
│   │   └── settings/
│   │       └── SettingsView.vue        # 系统设置
│   ├── router/
│   │   ├── index.ts                    # 路由入口
│   │   ├── routes.ts                   # 路由配置
│   │   └── guards.ts                   # 路由守卫
│   ├── stores/
│   │   ├── index.ts                    # Pinia 入口
│   │   ├── auth.ts                     # 认证/用户状态
│   │   ├── app.ts                      # 应用全局状态（侧边栏/主题）
│   │   ├── article.ts                  # 文章状态
│   │   ├── media.ts                    # 媒体状态
│   │   └── notification.ts             # 通知状态
│   ├── api/
│   │   ├── request.ts                  # Axios 封装 + 拦截器
│   │   ├── auth.ts                     # 认证 API
│   │   ├── user.ts                     # 用户 API
│   │   ├── article.ts                  # 文章 API
│   │   ├── category.ts                 # 分类标签 API
│   │   ├── media.ts                    # 媒体 API
│   │   ├── portfolio.ts                # 作品集 API
│   │   ├── video.ts                    # 视频 API
│   │   ├── comment.ts                  # 评论 API
│   │   ├── template.ts                 # 模板 API
│   │   └── system.ts                   # 系统配置 API
│   ├── composables/
│   │   ├── usePagination.ts            # 分页逻辑
│   │   ├── useTable.ts                 # 表格通用逻辑
│   │   ├── useUpload.ts                # 上传逻辑
│   │   ├── usePermission.ts            # 权限判断
│   │   └── useDebounce.ts              # 防抖
│   ├── utils/
│   │   ├── format.ts                   # 格式化工具（日期/文件大小）
│   │   ├── storage.ts                  # localStorage 封装
│   │   └── validators.ts               # 表单校验规则
│   ├── types/
│   │   ├── api.ts                      # API 响应类型
│   │   ├── user.ts                     # 用户类型
│   │   ├── article.ts                  # 文章类型
│   │   └── ...
│   ├── App.vue
│   └── main.ts
├── uno.config.ts                        # UnoCSS 配置
├── vite.config.ts
├── tsconfig.json
├── package.json
└── .env.development                     # 开发环境变量
```

---

## 四、路由设计

### 4.1 路由结构

```typescript
const routes = [
  // 认证（无布局）
  {
    path: '/auth',
    component: () => import('@/views/auth/AuthLayout.vue'),
    children: [
      { path: 'login', name: 'Login', component: () => import('@/views/auth/LoginView.vue') },
      { path: 'register', name: 'Register', component: () => import('@/views/auth/RegisterView.vue') },
      { path: 'forgot-password', name: 'ForgotPassword', component: () => import('@/views/auth/ForgotPasswordView.vue') },
    ],
  },

  // 主应用（侧边栏布局）
  {
    path: '/',
    component: () => import('@/components/layout/AppLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        meta: { title: '工作台', icon: 'DashboardOutlined' },
        component: () => import('@/views/dashboard/DashboardView.vue'),
      },
      {
        path: 'articles',
        name: 'ArticleList',
        meta: { title: '文章管理', icon: 'FileTextOutlined' },
        component: () => import('@/views/article/ArticleListView.vue'),
      },
      {
        path: 'articles/create',
        name: 'ArticleCreate',
        meta: { title: '写文章', hidden: true },
        component: () => import('@/views/article/ArticleCreateView.vue'),
      },
      {
        path: 'articles/:id/edit',
        name: 'ArticleEdit',
        meta: { title: '编辑文章', hidden: true },
        component: () => import('@/views/article/ArticleEditView.vue'),
      },
      {
        path: 'categories',
        name: 'Categories',
        meta: { title: '分类标签', icon: 'TagsOutlined' },
        component: () => import('@/views/category/CategoryManageView.vue'),
      },
      {
        path: 'media',
        name: 'MediaLibrary',
        meta: { title: '媒体库', icon: 'PictureOutlined' },
        component: () => import('@/views/media/MediaLibraryView.vue'),
      },
      {
        path: 'portfolios',
        name: 'Portfolios',
        meta: { title: '摄影作品集', icon: 'CameraOutlined' },
        component: () => import('@/views/portfolio/PortfolioListView.vue'),
      },
      {
        path: 'portfolios/:id',
        name: 'PortfolioDetail',
        meta: { title: '作品集详情', hidden: true },
        component: () => import('@/views/portfolio/PortfolioDetailView.vue'),
      },
      {
        path: 'videos',
        name: 'Videos',
        meta: { title: '视频管理', icon: 'PlaySquareOutlined' },
        component: () => import('@/views/video/VideoManageView.vue'),
      },
      {
        path: 'comments',
        name: 'Comments',
        meta: { title: '评论管理', icon: 'MessageOutlined' },
        component: () => import('@/views/comment/CommentManageView.vue'),
      },
      {
        path: 'templates',
        name: 'Templates',
        meta: { title: '模板风格', icon: 'SkinOutlined' },
        component: () => import('@/views/template/TemplateView.vue'),
      },
      // 管理员专属
      {
        path: 'users',
        name: 'Users',
        meta: { title: '用户管理', icon: 'TeamOutlined', role: 'admin' },
        component: () => import('@/views/user/UserManageView.vue'),
      },
      {
        path: 'roles',
        name: 'Roles',
        meta: { title: '角色权限', icon: 'SafetyOutlined', role: 'admin' },
        component: () => import('@/views/role/RoleManageView.vue'),
      },
      {
        path: 'settings',
        name: 'Settings',
        meta: { title: '系统设置', icon: 'SettingOutlined', role: 'admin' },
        component: () => import('@/views/settings/SettingsView.vue'),
      },
      {
        path: 'profile',
        name: 'Profile',
        meta: { title: '个人资料', icon: 'UserOutlined', hidden: true },
        component: () => import('@/views/user/ProfileView.vue'),
      },
    ],
  },

  // 404
  { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/views/error/NotFoundView.vue') },
];
```

### 4.2 路由守卫

```typescript
// guards.ts
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore();

  // 白名单路由直接放行
  if (to.path.startsWith('/auth')) {
    if (authStore.isLoggedIn) return next('/dashboard');
    return next();
  }

  // 未登录跳转登录页
  if (!authStore.isLoggedIn) {
    return next(`/auth/login?redirect=${to.path}`);
  }

  // 角色权限校验
  if (to.meta.role && !authStore.hasRole(to.meta.role as string)) {
    return next('/dashboard');
  }

  next();
});
```

---

## 五、页面设计

### 5.1 登录页

**设计方向**：极简创作氛围，大面积留白 + 品牌色点缀。

```
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│                    ┌─────────────────────┐                   │
│                    │    [Logo] Cus        │                   │
│                    │    创作者后台         │                   │
│                    │                     │                   │
│                    │  ┌─────────────────┐│                   │
│                    │  │ 邮箱/用户名      ││                   │
│                    │  └─────────────────┘│                   │
│                    │  ┌─────────────────┐│                   │
│                    │  │ 密码       👁    ││                   │
│                    │  └─────────────────┘│                   │
│                    │  [✓] 记住我         │                   │
│                    │  ┌─────────────────┐│                   │
│                    │  │    登  录        ││                   │
│                    │  └─────────────────┘│                   │
│                    │                     │                   │
│                    │  忘记密码？  注册账号 │                   │
│                    └─────────────────────┘                   │
│                                                              │
│         ── 用专业的方式，展示你的创作 ──                        │
└──────────────────────────────────────────────────────────────┘
```

**关键交互**：
- 登录成功后过渡动画跳转 Dashboard
- 密码可见/不可见切换图标
- 表单校验实时提示，输入框 focus 时边框亮色
- 登录按钮 loading 状态 + 脉冲动画

---

### 5.2 工作台 Dashboard

**设计方向**：数据驱动 + 卡片式布局，关键指标一目了然。

```
┌──────────────────────────────────────────────────────────────┐
│  工作台                                        最近更新: 刚刚  │
│  ─────────────────────────────────────────────────────────── │
│                                                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │
│  │ 📄 文章   │ │ 👁 访问量 │ │ 💬 评论   │ │ 📸 作品   │        │
│  │   128     │ │  12.4k   │ │   356    │ │   15     │        │
│  │ ↑ 12%    │ │ ↑ 8.3%   │ │ ↑ 5%     │ │ 较上周   │        │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │
│                                                              │
│  ┌─────────────────────────┐ ┌─────────────────────────────┐ │
│  │  访问趋势 (近30天)       │ │  快速操作                    │ │
│  │  ┌───────────────────┐  │ │  [✏️ 写文章]  [📷 上传作品]   │ │
│  │  │   📈 折线图        │  │ │  [📁 媒体库]  [💬 审核评论]  │ │
│  │  │                   │  │ │                             │ │
│  │  └───────────────────┘  │ │  待处理                      │ │
│  └─────────────────────────┘ │  · 3 条评论待审核            │ │
│                              │  · 2 篇文章草稿              │ │
│  ┌─────────────────────────┐ │  · 视频转码完成 1 个         │ │
│  │  最近文章                │ └─────────────────────────────┘ │
│  │  · 西藏之旅 Day5  → 6h前 │                                │
│  │  · 人像调色指南  → 1d前  │                                │
│  │  · 器材评测      → 3d前  │                                │
│  └─────────────────────────┘                                │
└──────────────────────────────────────────────────────────────┘
```

**关键交互**：
- 统计卡片加载时展示骨架屏，数字滚动动画
- 折线图 SVG 路径动画
- 快速操作按钮 hover 时微放大 + 阴影
- 最近文章列表点击跳转编辑

---

### 5.3 文章列表页

**设计方向**：双视图切换（表格/卡片）+ 强大筛选 + 批量操作。

```
┌──────────────────────────────────────────────────────────────┐
│  文章管理                        [+ 写文章]  [表格 ▦ 卡片 ▥]  │
│  ─────────────────────────────────────────────────────────── │
│  ┌──────────┬──────────┬──────────┬──────────┬────────────┐ │
│  │ 状态 ▾   │ 分类 ▾   │ 标签 ▾   │ 🔍 搜索  │  筛选      │ │
│  │ 全部     │ 全部     │          │          │            │ │
│  └──────────┴──────────┴──────────┴──────────┴────────────┘ │
│                                                              │
│  ┌─ 卡片视图 ──────────────────────────────────────────────┐ │
│  │                                                        │ │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │ │
│  │  │▉▉▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉▉▉│ │          │ │▉▉▉▉▉▉▉▉▉▉│  │ │
│  │  │ 西藏之旅  │ │ 人像指南  │ │ 器材评测  │ │ 音乐推荐  │  │ │
│  │  │ 旅行·攻略 │ │ 摄影·教程 │ │ 数码·测评 │ │ 音乐·歌单 │  │ │
│  │  │ 6h前 已发 │ │ 1d前 已发 │ │ 3d前 草稿 │ │ 5d前 已发 │  │ │
│  │  │ ⋯  [✏️]  │ │ ⋯  [✏️]  │ │ ⋯  [✏️]  │ │ ⋯  [✏️]  │  │ │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘  │ │
│  │                                                        │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ◀ 1  2  3  ...  8  ▶         共 128 条                      │
└──────────────────────────────────────────────────────────────┘
```

**卡片组件状态**：

```
┌──────────────────────┐
│  ╔══════════════════╗│  ← 封面图（预加载 blurhash 占位）
│  ║    [封面图]      ║│
│  ╚══════════════════╝│
│                      │
│  ● 已发布             │  ← 状态标签（绿/黄/灰/蓝）
│  西藏之旅 Day5        │  ← 标题（最大2行截断）
│  旅行 · 攻略          │  ← 分类+标签
│  6 小时前 · 👁 1.2k · 💬 23  │  ← 时间+浏览量+评论数
│  ───                │
│  编辑  预览  更多 ▾   │  ← 操作栏（hover显示）
└──────────────────────┘
```

**表格视图列定义**：

| 列 | 宽度 | 说明 |
|----|------|------|
| 复选框 | 48px | 批量选择 |
| 封面 | 80px | 缩略图 60x40 |
| 标题 | auto | 可点击跳转编辑 |
| 分类 | 120px | 标签 |
| 状态 | 100px | 标签样式 |
| 浏览 | 80px | 数字 |
| 评论 | 80px | 数字 |
| 发布时间 | 150px | 相对时间 |
| 操作 | 120px | 编辑/预览/更多 |

**关键交互**：
- 卡片 hover 时封面图微放大 + 阴影加深
- 草稿状态卡片有虚线边框，已发布为实线
- 批量操作栏在选择后从顶部滑入
- 搜索输入 300ms 防抖

---

### 5.4 文章编辑器

**设计方向**：沉浸式写作体验，分屏预览 + 自适应布局。

```
┌──────────────────────────────────────────────────────────────┐
│  ← 返回      草稿自动保存 ✓    [保存草稿]  [预览]  [发布 ▾]  │
│  ─────────────────────────────────────────────────────────── │
│  ┌──────────────────────────────────────────────────────────┐│
│  │ 文章标题（输入你的标题...）                                ││
│  └──────────────────────────────────────────────────────────┘│
│  ┌──────────────────────────────────────────────────────────┐│
│  │  [Markdown] [富文本]   │  [B] [I] [H1] [H2] [🔗] [📷] ...││
│  ├────────────────────────┴─────────────────────────────────┤│
│  │                                                          ││
│  │  ┌──────────────────────┐  ┌──────────────────────────┐  ││
│  │  │                      │  │                          │  ││
│  │  │   Markdown 编辑区    │  │   实时预览（HTML 渲染）   │  ││
│  │  │                      │  │                          │  ││
│  │  │  # 西藏之旅 Day5     │  │  ⌈ 西藏之旅 Day5 ⌋       │  ││
│  │  │                      │  │                          │  ││
│  │  │  今天我们从拉萨出发.. │  │  今天我们从拉萨出发...    │  ││
│  │  │                      │  │                          │  ││
│  │  │  ![](image.jpg)      │  │  [图片渲染]              │  ││
│  │  │                      │  │                          │  ││
│  │  └──────────────────────┘  └──────────────────────────┘  ││
│  │                                                          ││
│  └──────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌── 发布设置 ──────────────────────────────────────────────┐│
│  │  摘要: 今天我们从拉萨出发，沿着318国道一路向西...         ││
│  │  封面: [上传封面图]  ┌────┐                              ││
│  │  分类: [旅行 ▾]      标签: [西藏] [自驾] [+ 添加]        ││
│  │  URL:  /tibet-day5                         [编辑]       ││
│  │  定时: □ 设定发布时间  [2024-01-15 08:00]                ││
│  └──────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
```

**关键交互**：
- 编辑器左侧行号 + 当前行高亮
- 拖拽/粘贴图片自动上传 + 进度条
- 大纲（TOC）浮动在右侧，点击跳转
- `Ctrl+S` 保存草稿，顶部提示"已保存"
- 发布弹窗确认（摘要、封面、分类、定时）
- 离开页面未保存时浏览器拦截提示

---

### 5.5 媒体库

**设计方向**：网格展示 + 快速筛选 + 拖拽上传。

```
┌──────────────────────────────────────────────────────────────┐
│  媒体库                        [上传文件]  [新建文件夹 ▾]     │
│  ─────────────────────────────────────────────────────────── │
│  ┌──────────┬──────────┬──────────┬────────────────────────┐ │
│  │ 类型 ▾   │ 时间 ▾   │ 🔍 搜索  │  ☐ 仅未使用           │ │
│  └──────────┴──────────┴──────────┴────────────────────────┘ │
│                                                              │
│  ┌──────────────────────────────────────────────────────────┐│
│  │                                                          ││
│  │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐          ││
│  │  │▉▉▉▉▉▉│ │▉▉▉▉▉▉│ │▶ 2:30│ │▉▉▉▉▉▉│ │▉▉▉▉▉▉│          ││
│  │  │      │ │      │ │      │ │      │ │      │          ││
│  │  │IMG001│ │IMG002│ │VID001│ │IMG003│ │IMG004│          ││
│  │  │2.1MB │ │3.4MB │ │45MB  │ │1.2MB │ │5.1MB │          ││
│  │  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘          ││
│  │                                                          ││
│  │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐          ││
│  │  │▉▉▉▉▉▉│ │▉▉▉▉▉▉│ │▉▉▉▉▉▉│ │▉▉▉▉▉▉│ │▉▉▉▉▉▉│          ││
│  │  │      │ │      │ │      │ │      │ │      │          ││
│  │  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘          ││
│  │                                                          ││
│  └──────────────────────────────────────────────────────────┘│
│                        ◀ 1  2  3  4  5  ▶                    │
│                                                              │
│  ┌── 拖拽文件到此处上传 ────────────────────────────────────┐ │
│  │        📁  支持 JPG/PNG/GIF/MP4/MP3，单文件最大 500MB     │ │
│  └──────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

**关键交互**：
- 选中文件后蓝色边框 + 勾选标记
- 右键菜单：复制链接 / 下载 / 删除 / 查看详情
- 点击图片打开大图预览（Lightbox）
- 侧边详情面板：文件信息、尺寸、引用文章列表
- 拖拽区域 hover 时高亮 + 虚线边框变实线

---

### 5.6 评论管理

**设计方向**：列表式审核流 + 快速操作。

```
┌──────────────────────────────────────────────────────────────┐
│  评论管理                            [全部] [待审核(3)] [垃圾]│
│  ─────────────────────────────────────────────────────────── │
│                                                              │
│  ┌── 评论项 ────────────────────────────────────────────────┐│
│  │  ┌────┐                                                   ││
│  │  │头像│  张三  ·  3分钟前  ·  在《西藏之旅》中             ││
│  │  └────┘                                                   ││
│  │  照片拍得太美了！请问是什么相机拍的？                       ││
│  │  ───────────────────────────────────────────────          ││
│  │  [通过]  [回复]  [标记垃圾]  [删除]                        ││
│  └──────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌── 评论项（待审核）───────────────────────────────────────┐│
│  │  ┌────┐  🟡 待审核                                        ││
│  │  │头像│  李四  ·  10分钟前  ·  在《人像调色指南》中         ││
│  │  └────┘                                                   ││
│  │  加微信 xxx123456 免费领取修图教程                          ││
│  │  ───────────────────────────────────────────────────────────│
│  │  [通过]  [标记垃圾]  [删除]                                ││
│  └──────────────────────────────────────────────────────────┘│
│                                                              │
│  ◀ 1  2  3  ▶                              共 56 条          │
└──────────────────────────────────────────────────────────────┘
```

---

### 5.7 模板选择页

**设计方向**：卡片预览 + 即时切换 + 实时预览链接。

```
┌──────────────────────────────────────────────────────────────┐
│  模板风格              当前: [极简白]  [预览我的博客 →]        │
│  ─────────────────────────────────────────────────────────── │
│                                                              │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐         │
│  │  ┌──────────┐│ │  ┌──────────┐│ │  ┌──────────┐│         │
│  │  │          ││ │  │  ████    ││ │  │░░░░░░░░░░││         │
│  │  │  极简白  ││ │  │  ████    ││ │  │ 暗夜黑    ││         │
│  │  │          ││ │  │  杂志风  ││ │  │░░░░░░░░░░││         │
│  │  └──────────┘│ │  └──────────┘│ │  └──────────┘│         │
│  │  ● 当前使用   │ │              │ │              │         │
│  │  简洁干净     │ │  图文并茂     │ │  护眼深色     │         │
│  │  [预览]      │ │  [预览] [应用]│ │  [预览] [应用]│         │
│  └──────────────┘ └──────────────┘ └──────────────┘         │
│                                                              │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐         │
│  │  ┌──────────┐│ │  ┌──────────┐│ │  ┌──────────┐│         │
│  │  │ ▓▓▓▓▓▓▓▓ ││ │  │░░░░░░░░░░││ │  │░░░░░░░░░░││         │
│  │  │ ▓ 摄影 ▓ ││ │  │░░ 视频 ░░││ │  │░░ 旅行 ░░││         │
│  │  │ ▓▓▓▓▓▓▓▓ ││ │  │░░░░░░░░░░││ │  │░░░░░░░░░░││         │
│  │  └──────────┘│ │  └──────────┘│ │  └──────────┘│         │
│  │  摄影专用     │ │  视频创作者   │ │  旅行博主     │         │
│  │  [预览] [应用]│ │  [预览] [应用]│ │  [预览] [应用]│         │
│  └──────────────┘ └──────────────┘ └──────────────┘         │
└──────────────────────────────────────────────────────────────┘
```

**关键交互**：
- 当前使用模板有选中边框 + 角标
- 切换模板弹窗确认："切换后博客将立即更新外观，确定切换？"
- "预览我的博客"按钮在新标签打开博客实际效果

---

### 5.8 作品集详情页

**设计方向**：照片墙 + 拖拽排序 + 批量编辑。

```
┌──────────────────────────────────────────────────────────────┐
│  ← 返回作品集列表    人像精选  [编辑信息]  [分享]  [···]      │
│  描述：2024年创作的人像作品合集  ·  12 张照片  ·  创建于 3月   │
│  ─────────────────────────────────────────────────────────── │
│  [上传照片]  [批量编辑]  [排序模式]              ☐ 全选       │
│                                                              │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐                │
│  │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│                │
│  │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│                │
│  │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│                │
│  │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│                │
│  │ 照片1  │ │ 照片2  │ │ 照片3  │ │ 照片4  │                │
│  │ ☐     │ │ ☐     │ │ ☐     │ │ ☐     │                │
│  └────────┘ └────────┘ └────────┘ └────────┘                │
│                                                              │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐                │
│  │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │        │ │        │                │
│  │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │  上传  │ │  上传  │                │
│  │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │  照片  │ │  照片  │                │
│  │▉▉▉▉▉▉▉▉│ │▉▉▉▉▉▉▉▉│ │   +    │ │   +    │                │
│  │ 照片5  │ │ 照片6  │ │        │ │        │                │
│  │ ☐     │ │ ☐     │ │        │ │        │                │
│  └────────┘ └────────┘ └────────┘ └────────┘                │
└──────────────────────────────────────────────────────────────┘
```

**关键交互**：
- 排序模式下卡片可拖拽重排
- 点击照片打开 Lightbox 大图浏览
- 侧边栏展示照片 EXIF 信息（光圈/快门/ISO）
- 批量编辑：一键修改标题模板、添加水印

---

### 5.9 系统设置页

**设计方向**：分组表单 + 实时保存。

```
┌──────────────────────────────────────────────────────────────┐
│  系统设置                                                     │
│  ─────────────────────────────────────────────────────────── │
│                                                              │
│  ┌── 站点信息 ──────────────────────────────────────────────┐│
│  │  站点名称:  [Cus Blog                        ]           ││
│  │  站点描述:  [一个创作者的内容平台              ]           ││
│  │  Logo:     [上传]  ┌────┐                                ││
│  │  Favicon:  [上传]  ┌────┐                                ││
│  └──────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌── SEO 配置 ──────────────────────────────────────────────┐│
│  │  首页标题:   [Cus Blog - 创作你的世界          ]           ││
│  │  关键词:     [博客, 摄影, 旅行, 创作            ]           ││
│  │  描述:       [Cus Blog 是一个面向创作者...      ]           ││
│  └──────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌── 评论设置 ──────────────────────────────────────────────┐│
│  │  开启评论:  [🟢 开启]                                      ││
│  │  评论审核:  [🟢 所有评论需审核]                             ││
│  │  垃圾过滤:  [🟢 开启关键词过滤]      [配置关键词]           ││
│  └──────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌── 第三方集成 ────────────────────────────────────────────┐│
│  │  B站嵌入:   [🟢 开启]                                      ││
│  │  音乐平台:  [☑ 网易云] [☑ QQ音乐] [☐ Spotify]              ││
│  └──────────────────────────────────────────────────────────┘│
│                                                              │
│                                           [恢复默认]  [保存]  │
└──────────────────────────────────────────────────────────────┘
```

---

## 六、状态管理（Pinia）

### 6.1 authStore

```typescript
// stores/auth.ts
interface AuthState {
  token: string | null;
  refreshToken: string | null;
  user: UserInfo | null;
  permissions: string[];
}

// 核心 actions
- login(credentials)       // 登录 → 存储 token → 获取用户信息
- logout()                 // 清除 token → 跳转登录页
- refreshToken()           // 无感刷新 access token
- fetchUserInfo()          // 获取当前用户信息 + 权限
- hasPermission(code)      // 判断是否有某权限
- hasRole(role)            // 判断角色
```

### 6.2 appStore

```typescript
// stores/app.ts
interface AppState {
  sidebarCollapsed: boolean;
  breadcrumbs: BreadcrumbItem[];
  globalLoading: boolean;
}

// 核心 actions
- toggleSidebar()          // 折叠/展开侧边栏
- setBreadcrumbs(items)    // 设置面包屑
```

### 6.3 articleStore

```typescript
// stores/article.ts
interface ArticleState {
  list: Article[];
  total: number;
  current: Article | null;
  filters: ArticleFilters;
  pagination: Pagination;
}

// 核心 actions
- fetchList(params)        // 获取文章列表
- fetchDetail(id)          // 获取文章详情
- create(data)             // 创建文章
- update(id, data)         // 更新文章
- remove(id)               // 删除文章
- updateStatus(id, status) // 修改状态
```

---

## 七、API 层设计

### 7.1 Axios 封装

```typescript
// api/request.ts
import axios from 'axios';
import { message } from 'ant-design-vue';
import { useAuthStore } from '@/stores/auth';
import router from '@/router';

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
});

// 请求拦截器
request.interceptors.request.use((config) => {
  const auth = useAuthStore();
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`;
  }
  return config;
});

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const { code, msg, data } = response.data;
    if (code !== 0) {
      message.error(msg || '请求失败');
      return Promise.reject(new Error(msg));
    }
    return data;
  },
  async (error) => {
    if (error.response?.status === 401) {
      // Token 过期，尝试刷新
      const auth = useAuthStore();
      try {
        await auth.refreshToken();
        return request(error.config); // 重试
      } catch {
        auth.logout();
        router.push('/auth/login');
      }
    }
    message.error(error.message || '网络错误');
    return Promise.reject(error);
  }
);

export default request;
```

### 7.2 API 模块示例

```typescript
// api/article.ts
import request from './request';

export const articleApi = {
  // 获取文章列表
  getList(params: ArticleListParams) {
    return request.get<PaginatedData<Article>>('/articles', { params });
  },
  // 获取文章详情
  getDetail(id: string) {
    return request.get<Article>(`/articles/${id}`);
  },
  // 创建文章
  create(data: ArticleCreateReq) {
    return request.post<Article>('/articles', data);
  },
  // 更新文章
  update(id: string, data: ArticleUpdateReq) {
    return request.put<Article>(`/articles/${id}`, data);
  },
  // 删除文章
  remove(id: string) {
    return request.delete(`/articles/${id}`);
  },
  // 发布文章
  publish(id: string) {
    return request.put(`/articles/${id}/publish`);
  },
  // 保存草稿
  saveDraft(id: string, data: ArticleCreateReq) {
    return request.post(`/articles/${id}/draft`, data);
  },
};
```

---

## 八、通用组件设计

### 8.1 侧边栏 AppSidebar

```
┌──────────────┐
│              │
│  [Logo] Cus  │  品牌区
│  ─────────── │
│              │
│  📊 工作台   │  导航菜单
│  📄 文章管理 │  · 图标 + 文字
│  🏷 分类标签 │  · 选中态：左侧蓝色竖条 + 背景高亮
│  📁 媒体库   │  · 折叠态：仅图标 + Tooltip
│  📸 摄影作品 │
│  ▶ 视频管理  │
│  💬 评论管理 │
│  🎨 模板风格 │
│  ─────────── │
│  👥 用户管理 │  管理员专区
│  🔐 角色权限 │  (v-if="isAdmin")
│  ⚙ 系统设置 │
│  ─────────── │
│              │
│  ┌──────────┐│  用户信息区
│  │ 头像 用户名││
│  │ 博主      ││
│  └──────────┘│
│  个人资料    │
│  退出登录    │
└──────────────┘
```

**折叠状态**：

```
┌─────┐
│     │
│ [◆] │  Logo 符号
│ ─── │
│ 📊  │
│ 📄  │
│ 🏷  │
│ 📁  │
│ 📸  │
│ ▶   │
│ 💬  │
│ 🎨  │
│ ─── │
│ 👥  │
│ 🔐  │
│ ⚙   │
│     │
│ ┌─┐ │
│ │头│ │
│ └─┘ │
└─────┘
```

**关键交互**：
- 折叠/展开按钮：侧边栏底部或顶部汉堡图标
- 折叠时鼠标悬停菜单项弹出 Tooltip 显示完整名称
- 当前路由自动高亮对应菜单项
- 菜单项过渡动画：文字淡入淡出

### 8.2 顶部栏 AppHeader

```
┌──────────────────────────────────────────────────────────────┐
│  [☰]  // 面包屑导航  工作台 / 文章管理 / 写文章               │
│                                                    [🔔] [👤]│
│                                  通知(3)    用户菜单 ▾        │
└──────────────────────────────────────────────────────────────┘
```

**通知下拉**：

```
┌──────────────────────────┐
│  通知 (3)         全部已读│
│  ─────────────────────── │
│  🟢 新评论：张三评论了...  │
│     3 分钟前              │
│  ─────────────────────── │
│  🟢 视频转码完成          │
│     1 小时前              │
│  ─────────────────────── │
│  🔵 系统更新通知          │
│     2 天前                │
│  ─────────────────────── │
│  查看全部通知 →           │
└──────────────────────────┘
```

### 8.3 空状态组件 EmptyState

```
┌──────────────────────────────────────┐
│                                      │
│            ┌──────────┐              │
│            │  📭      │              │
│            │  暂无数据 │              │
│            └──────────┘              │
│                                      │
│        还没有任何文章，开始创作吧     │
│                                      │
│           [写第一篇文章 →]            │
│                                      │
└──────────────────────────────────────┘
```

### 8.4 骨架屏 LoadingSkeleton

```
┌──────────────────────────────────────┐
│  ██████████████                      │  标题骨架
│  ─────────────────────────────────── │
│  ┌────────┐ ┌────────┐ ┌────────┐   │
│  │░░░░░░░░│ │░░░░░░░░│ │░░░░░░░░│   │  统计卡片骨架
│  │░░░░░░░░│ │░░░░░░░░│ │░░░░░░░░│   │
│  └────────┘ └────────┘ └────────┘   │
│  ┌──────────────────────────────────┐│
│  │░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░││  图表骨架
│  │░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░││
│  └──────────────────────────────────┘│
└──────────────────────────────────────┘
```

---

## 九、关键交互流程

### 9.1 文章发布流程

```
┌─────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ 点击     │───▶│ 打开发布  │───▶│ 表单校验  │───▶│ 调用 API │
│ [发布]   │    │ 设置弹窗  │    │通过后提交 │    │          │
└─────────┘    └──────────┘    └──────────┘    └────┬─────┘
                                                    │
                     ┌──────────────────────────────┘
                     ▼
              ┌──────────┐    ┌──────────┐
              │ 成功：    │    │ 失败：    │
              │ Toast提醒 │    │ 错误提示  │
              │ 跳转列表页│    │ 保留表单  │
              └──────────┘    └──────────┘
```

### 9.2 文件上传流程

```
┌─────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ 选择文件 │───▶│ 校验类型  │───▶│ 分片上传  │───▶│ 合并通知  │
│ 或拖拽   │    │ 和大小   │    │ 进度条   │    │ 服务端    │
└─────────┘    └──────────┘    └──────────┘    └────┬─────┘
                                                    │
              ┌─────────────────────────────────────┘
              ▼
       ┌──────────┐    ┌──────────┐
       │ 刷新列表  │    │ 自动插入  │
       │ 显示新文件│    │ 编辑器    │
       └──────────┘    └──────────┘
```

### 9.3 权限校验流程

```
┌─────────┐    ┌──────────┐    ┌──────────┐
│ 访问页面 │───▶│ 路由守卫  │───▶│ 检查     │
│          │    │ beforeEach│    │ token    │
└─────────┘    └──────────┘    └────┬─────┘
                                    │
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
              ┌──────────┐   ┌──────────┐   ┌──────────┐
              │ 无 token │   │ token过期 │   │ 有 token │
              │ 跳转登录 │   │ 尝试刷新  │   │ 检查角色 │
              └──────────┘   └──────────┘   └────┬─────┘
                                                │
                                    ┌───────────┼───────────┐
                                    ▼                       ▼
                              ┌──────────┐           ┌──────────┐
                              │ 有权限   │           │ 无权限   │
                              │ 正常渲染 │           │ 跳转403  │
                              └──────────┘           └──────────┘
```

---

## 十、UnoCSS 主题配置

```typescript
// uno.config.ts
import { defineConfig, presetUno, presetAttributify } from 'unocss';

export default defineConfig({
  presets: [presetUno(), presetAttributify()],
  theme: {
    colors: {
      primary: '#4a6cf7',
      accent: '#7c3aed',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
    },
  },
  shortcuts: {
    'page-container': 'max-w-1400px mx-auto p-6',
    'card': 'bg-white rounded-lg shadow-card p-6',
    'card-hover': 'card transition-transform duration-250 hover:-translate-y-2px hover:shadow-dropdown',
    'stat-card': 'card flex items-center gap-4',
    'stat-value': 'text-2xl font-bold text-primary',
    'stat-label': 'text-sm text-secondary',
    'stat-trend-up': 'text-xs text-success',
    'stat-trend-down': 'text-xs text-error',
    'btn-primary': 'bg-primary text-white px-4 py-2 rounded-md hover:bg-primary-hover transition-colors',
    'btn-ghost': 'text-secondary hover:text-primary hover:bg-primary-light px-3 py-1 rounded-md transition-all',
    'input-field': 'w-full px-3 py-2 border border-border rounded-md focus:border-primary focus:ring-2 focus:ring-primary-light outline-none transition-all',
    'tag': 'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
    'tag-default': 'tag bg-gray-100 text-secondary',
    'tag-primary': 'tag bg-primary-light text-primary',
    'tag-success': 'tag bg-green-50 text-success',
    'tag-warning': 'tag bg-amber-50 text-warning',
    'tag-error': 'tag bg-red-50 text-error',
    'sidebar-link': 'flex items-center gap-3 px-4 py-2.5 mx-2 rounded-lg text-sidebar hover:bg-sidebar-hover transition-all',
    'sidebar-link-active': 'sidebar-link bg-sidebar-active text-sidebar-active',
    'page-header': 'flex items-center justify-between mb-6',
    'page-title': 'text-xl font-semibold text-primary',
    'table-container': 'card overflow-hidden',
    'filter-bar': 'flex items-center gap-3 mb-4 flex-wrap',
  },
});
```

---

## 十一、环境变量

```bash
# .env.development
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_TITLE=Cus CMS
VITE_UPLOAD_MAX_SIZE=524288000  # 500MB
VITE_UPLOAD_CHUNK_SIZE=5242880  # 5MB

# .env.production
VITE_API_BASE_URL=https://api.cus-blog.com/api/v1
VITE_APP_TITLE=Cus CMS
VITE_UPLOAD_MAX_SIZE=524288000
VITE_UPLOAD_CHUNK_SIZE=5242880
```

---

## 十二、开发阶段规划

| 阶段 | 周期 | 目标 |
|------|------|------|
| **Phase 1** | 2周 | 项目脚手架 + 登录页 + 主布局 + Dashboard + 路由/权限 |
| **Phase 2** | 2周 | 文章管理全流程（列表/创建/编辑/发布）+ 分类标签 |
| **Phase 3** | 2周 | 媒体库 + 作品集管理 + 视频管理 |
| **Phase 4** | 1周 | 评论管理 + 模板选择 + 系统设置 |
| **Phase 5** | 1周 | 用户管理 + 角色权限 + 数据统计 + 联调优化 |

---

## 十三、技术实现要点

### 13.1 大文件分片上传

```typescript
// composables/useUpload.ts
const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB

async function uploadFile(file: File, onProgress?: (p: number) => void) {
  const chunks = Math.ceil(file.size / CHUNK_SIZE);
  const uploadId = uuid();

  for (let i = 0; i < chunks; i++) {
    const chunk = file.slice(i * CHUNK_SIZE, (i + 1) * CHUNK_SIZE);
    const formData = new FormData();
    formData.append('chunk', chunk);
    formData.append('upload_id', uploadId);
    formData.append('chunk_index', String(i));
    formData.append('total_chunks', String(chunks));
    formData.append('filename', file.name);

    await mediaApi.uploadChunk(formData);
    onProgress?.((i + 1) / chunks * 100);
  }

  return mediaApi.mergeChunks(uploadId, file.name);
}
```

### 13.2 Markdown 编辑器（ByteMD）

```vue
<!-- components/editor/MarkdownEditor.vue -->
<template>
  <div class="markdown-editor-wrapper">
    <Editor
      :value="modelValue"
      :plugins="plugins"
      @change="handleChange"
      :locale="zhHans"
    />
  </div>
</template>

<script setup lang="ts">
import { Editor, Viewer } from '@bytemd/vue-next';
import gfm from '@bytemd/plugin-gfm';
import highlight from '@bytemd/plugin-highlight';
import zhHans from 'bytemd/locales/zh_Hans.json';

const plugins = [gfm(), highlight()];
</script>
```

### 13.3 权限指令

```typescript
// directives/permission.ts
import { useAuthStore } from '@/stores/auth';

export const vPermission = {
  mounted(el: HTMLElement, binding: { value: string }) {
    const auth = useAuthStore();
    if (!auth.hasPermission(binding.value)) {
      el.parentNode?.removeChild(el);
    }
  },
};
```

```vue
<!-- 使用示例 -->
<a-button v-permission="'article:delete'">删除</a-button>
```

---

## 十四、页面状态覆盖

每个页面需覆盖以下状态：

| 状态 | 组件 | 说明 |
|------|------|------|
| **加载中** | `LoadingSkeleton` | 骨架屏，与真实布局一致 |
| **空数据** | `EmptyState` | 引导性空状态 + 操作入口 |
| **错误** | `a-result` | Ant Design 错误结果 + 重试按钮 |
| **正常** | 页面内容 | 数据渲染后的完整页面 |
| **403** | `a-result` | 无权限提示 |
| **404** | `a-result` | 页面不存在 |

---

## 十五、附录：Ant Design Vue 组件覆盖

| 场景 | 使用组件 |
|------|----------|
| 表格 | `a-table` + `a-table-column` |
| 表单 | `a-form` + `a-form-item` + `a-input` + `a-select` |
| 弹窗 | `a-modal` |
| 抽屉 | `a-drawer` |
| 消息提示 | `message` (全局) |
| 通知 | `notification` (全局) |
| 确认 | `a-popconfirm` / `Modal.confirm` |
| 下拉菜单 | `a-dropdown` + `a-menu` |
| 标签页 | `a-tabs` + `a-tab-pane` |
| 分页 | `a-pagination` |
| 上传 | `a-upload` + `a-upload-dragger` |
| 图片预览 | `a-image` + `a-image-preview-group` |
| 骨架屏 | `a-skeleton` |
| 空状态 | `a-empty` |
| 结果 | `a-result` |
| 面包屑 | `a-breadcrumb` + `a-breadcrumb-item` |
| 标签 | `a-tag` |
| 开关 | `a-switch` |
| 日期选择 | `a-date-picker` |
| 统计数值 | `a-statistic` |
| 卡片 | `a-card` + `a-card-meta` |
| 折叠面板 | `a-collapse` + `a-collapse-panel` |
| 步骤条 | `a-steps` + `a-step` |
| 进度条 | `a-progress` |
| 级联选择 | `a-cascader` |
| 树选择 | `a-tree-select` |
| 时间线 | `a-timeline` + `a-timeline-item` |