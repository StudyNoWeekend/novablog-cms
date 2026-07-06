# Tasks

- [x] Task 1: 创建 API 层与类型定义
  - [x] SubTask 1.1: 新建 `src/types/storage.ts`，定义 StorageConfig / StorageConfigForm / MigrationTask / MigrationItem / AnalyzeResult 等 TypeScript 类型
  - [x] SubTask 1.2: 新建 `src/api/storage.ts`，封装后端所有存储配置和迁移接口

- [x] Task 2: 设置页子导航布局（SettingsLayout）
  - [x] SubTask 2.1: 新建 `src/components/layout/SettingsLayout.vue` — 左侧垂直 Menu（个人资料 / 对象存储），右侧 `<router-view />` 渲染子路由内容

- [x] Task 3: 存储配置管理（Tab 1）
  - [x] SubTask 3.1: 新建 `src/views/storage/StorageConfigView.vue` — Tabs 布局（存储配置 + 素材迁移），默认展示存储配置 Tab
  - [x] SubTask 3.2: 实现配置列表表格 + 激活 / 删除 / 编辑 / 新增按钮
  - [x] SubTask 3.3: 实现新增/编辑 Modal 表单（平台参数差异化展示 + 测试连通性按钮）

- [x] Task 4: 素材迁移（Tab 2）
  - [x] SubTask 4.1: 在 StorageConfigView 中实现第二个 Tab「素材迁移」— 活跃平台信息 + 分析按钮
  - [x] SubTask 4.2: 实现 analyze 交互（轮询）+ 结果展示（已存在 / 需迁移表格）
  - [x] SubTask 4.3: 实现迁移交互（一键 / 批量）+ 进度轮播 + 完成展示

- [x] Task 5: 路由、菜单与个人资料页改造
  - [x] SubTask 5.1: `routes.ts` 将 `/profile` 改为子路由结构（redirect → /profile/info，children: info/storage）
  - [x] SubTask 5.2: `AppSidebar.vue` 中「设置」链接 `to="/profile"` 改为 `to="/profile/info"`
  - [x] SubTask 5.3: `ProfileView.vue` 精简为纯内容组件（移除 page-header，由 SettingsLayout 管理）

- [x] Task 6: 编译验证
  - [x] SubTask 6.1: `npm run build` 通过，无 TypeScript 编译错误

# Task Dependencies
- Task 1 是基础，Task 3/4 依赖 Task 1
- Task 2 可独立进行
- Task 5 依赖 Task 2
- Task 6 依赖全部完成
