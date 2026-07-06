# 对象存储前端集成 Spec

## Why
后端已实现对象存储多平台配置（阿里云/腾讯云/MinIO）和素材迁移引擎（[后端 Spec](file:///Users/apple/GolandProjects/ai-cus/.trae/specs/add-object-storage-config/spec.md)），但缺少管理界面。管理员需要能在「设置」页面内通过子导航管理对象存储。

## What Changes
- 新增 API 层 `src/api/storage.ts`，封装存储配置和迁移所有接口
- 新增 `src/types/storage.ts`，定义前端存储配置和迁移的 TypeScript 类型
- 新增 `SettingsLayout.vue` — 设置页面的子导航布局，左侧垂直导航「个人资料」「对象存储」，右侧显示对应内容
- **MODIFIED** `ProfileView.vue` → 降级为纯个人资料内容组件（不再直接作为路由 page）
- 新增对象存储页面 `StorageConfigView.vue` — 包含「存储配置」和「素材迁移」两个 Tab
- **MODIFIED** 路由：`/profile` 改为 `/profile/info`（个人资料）和 `/profile/storage`（对象存储），`/profile` 重定向到 `/profile/info`
- **MODIFIED** `AppSidebar.vue`：「设置」菜单链接改为 `/profile/info`

## Impact
- Affected specs: cus-cms-admin-ui
- Affected code:
  - `src/api/storage.ts`（新建）
  - `src/types/storage.ts`（新建）
  - `src/components/layout/SettingsLayout.vue`（新建 — 设置页子导航布局）
  - `src/views/storage/StorageConfigView.vue`（新建 — 对象存储页，含配置+迁移双 Tab）
  - `src/views/user/ProfileView.vue`（修改 — 精简为纯个人资料内容）
  - `src/router/routes.ts`（修改 — /profile 改为子路由结构）
  - `src/components/layout/AppSidebar.vue`（修改 — 设置链接改为 /profile/info）

## 后端 API 对照（已实现，前端需对接）

### 存储配置 API
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/storage/config` | 列表（access_secret 脱敏） |
| POST | `/api/v1/storage/config` | 新建/更新（密钥加密后落库） |
| DELETE | `/api/v1/storage/config/:provider` | 删除（非 active 才能删） |
| PUT | `/api/v1/storage/config/:provider/activate` | 激活 → 热重载 |
| POST | `/api/v1/storage/config/test` | 连通性测试（不落库） |

### 迁移 API
| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/v1/storage/migration/analyze` | 启动分析任务 |
| GET | `/api/v1/storage/migration/analyze/:taskId` | 分析结果（缺失/已存在清单） |
| POST | `/api/v1/storage/migration/start` | 启动迁移（body: media_ids[] 或 all=true） |
| GET | `/api/v1/storage/migration/status/:taskId` | 轮询进度 |

## ADDED Requirements

### Requirement: 设置页面子导航布局
系统 SHALL 提供一个 SettingsLayout 布局，左侧为垂直导航菜单，右侧为对应的内容区域。

#### Scenario: 设置页面布局
- **WHEN** 管理员点击侧边栏「设置」
- **THEN** 页面整体布局为：
```
┌─────────────────────────────────────────────┐
│ 设置         (page header)                  │
├──────────┬──────────────────────────────────┤
│ 个人资料  │  (内容区域)                      │
│ 对象存储  │  根据左侧选中展示对应内容         │
│           │                                  │
└──────────┴──────────────────────────────────┘
```
- **AND** 左侧导航宽度约 200px，使用 ant-design-vue 的 Menu 组件（mode="vertical"、无图标纯文字）
- **AND** 右侧内容区域使用 `<router-view />` 嵌套渲染子路由内容
- **AND** 左侧菜单高亮当前选中的路由项

#### Scenario: 菜单项与路由对应
- **WHEN** 左侧点击「个人资料」
- **THEN** 路由跳转到 `/profile/info`，右侧显示个人资料内容
- **WHEN** 左侧点击「对象存储」
- **THEN** 路由跳转到 `/profile/storage`，右侧显示对象存储配置与迁移

### Requirement: 存储配置管理（Tab 1）
系统 SHALL 在对象存储页面第一个 Tab 提供存储配置管理，以表格展示所有已配置的平台，并提供新增/编辑/激活/测试/删除操作。

#### Scenario: 查看配置列表
- **WHEN** 管理员进入对象存储页面
- **THEN** 页面顶部为 Tabs：「存储配置」（默认）和「素材迁移」
- **AND** 存储配置 Tab 显示以表格展示的配置列表，每条展示 endpoint / bucket / provider 类型 / is_active 状态 / 最后编辑时间
- **AND** access_secret 列显示 `******`（后端已脱敏）
- **AND** 已激活的配置行有「当前使用中」Tag

#### Scenario: 新增/编辑配置（Modal 表单）
- **WHEN** 管理员点击「新增配置」按钮
- **THEN** 弹出 Modal，包含表单字段：provider 下拉（aliyun/tencent/minio）、endpoint、region、bucket、access_key、access_secret、path_prefix、custom_domain
- **AND** 选择 provider 时动态显示对应的**平台参数提示**：
  - 阿里云：endpoint 示例 `https://oss-cn-hangzhou.aliyuncs.com`、region 必填、bucket 纯桶名
  - 腾讯云：endpoint 示例 `https://cos.ap-guangzhou.myqcloud.com`、bucket 必须带 APPID（如 `examplebucket-1250000000`）
  - MinIO：endpoint 示例 `play.min.io:9000`（不含 scheme）、显示 use_ssl 开关（自动填充 extra 字段）
- **AND** 表单校验：provider/endpoint/bucket/access_key/access_secret 必填
- **WHEN** 编辑已有配置
- **THEN** access_secret 输入框为空表示不修改，placeholder 显示「留空不修改」

#### Scenario: 测试连通性
- **WHEN** 管理员在表单填写中点击「测试连通性」
- **THEN** 用当前表单数据（access_secret 明文）调 test 接口
- **AND** 结果展示：成功则绿色提示「连通成功」；失败则红色显示后端返回的错误详情

#### Scenario: 激活与删除配置
- **WHEN** 点击未激活行的「激活」按钮
- **THEN** 弹出确认框：「激活后将切换存储平台，确定继续？」
- **AND** 确认后调 activate 接口，成功后刷新列表
- **AND** 若后端返回「迁移进行中」错误，提示「当前有迁移任务运行中，请等待迁移完成后再切换」
- **WHEN** 点击非 active 行的「删除」按钮
- **THEN** 弹出确认框，确认后调 delete 接口
- **AND** 若后端拒绝（active 配置），提示「已激活的配置不能直接删除」

### Requirement: 素材迁移（Tab 2）
系统 SHALL 在对象存储页面第二个 Tab 提供素材迁移功能，支持分析当前素材 vs 目标平台的存在状态，以及执行迁移。

#### Scenario: 分析素材
- **WHEN** 管理员切换到「素材迁移」Tab
- **THEN** 页面显示当前活跃平台信息（「当前存储平台：阿里云 OSS / 腾讯云 COS / MinIO」）
- **AND** 显示「开始分析」按钮
- **WHEN** 点击「开始分析」
- **THEN** 调 analyze 接口，立即轮询分析结果（每 2 秒）
- **AND** 分析中显示 Spin 加载状态
- **WHEN** 分析完成
- **THEN** 展示结果统计：「共 N 条素材，M 条已存在，K 条需要迁移」
- **AND** 下方分两个表格展示：「已存在」（灰色，不可操作）和「需迁移」（可勾选）
- **AND** 每个表格展示列：文件名、当前存储平台(storage_type)、路径

#### Scenario: 执行迁移
- **WHEN** 分析结果展示后
- **THEN** 显示「一键迁移所有」和「批量迁移」（勾选后可用）两个操作按钮
- **WHEN** 点击任一迁移按钮
- **THEN** 调 start 接口，创建迁移任务
- **AND** 页面切换为进度展示模式：进度条（succeeded/total）、实时计数、失败清单
- **AND** 每 2 秒轮询 status 接口更新进度
- **WHEN** 迁移完成
- **THEN** 进度条 100%，显示总耗时，失败项可展开查看错误详情
- **AND** 提供「重新分析」按钮

#### Scenario: 无活跃配置时
- **WHEN** 当前无可使用的存储配置（is_active 不存在）
- **THEN** 素材迁移 Tab 显示提示「请先在「存储配置」中配置并激活一个存储平台」

## MODIFIED Requirements

### Requirement: 设置路由结构
原 `/profile` 单一路由（对应 `ProfileView.vue`）改为子路由结构：

```typescript
{
  path: 'profile',
  redirect: '/profile/info',
  component: () => import('@/components/layout/SettingsLayout.vue'),
  children: [
    {
      path: 'info',
      name: 'Profile',
      meta: { title: '个人资料' },
      component: () => import('@/views/user/ProfileView.vue'),
    },
    {
      path: 'storage',
      name: 'StorageConfig',
      meta: { title: '对象存储' },
      component: () => import('@/views/storage/StorageConfigView.vue'),
    },
  ],
},
```

注意：SettingsLayout 本身不含 `meta.hidden`，但每个子路由也不用「隐藏」，Sidebar 菜单直接链接到 `/profile/info`，设置页面内部的左侧导航通过路由激活状态自动高亮。

### Requirement: 侧边栏「设置」菜单链接
原 `AppSidebar.vue` 中「设置」链接 `to="/profile"` 改为 `to="/profile/info"`，匹配新的路由结构。

### Requirement: 个人资料页面
`ProfileView.vue` 精简为纯内容组件（移除 page-header，由 SettingsLayout 统一管理标题），只保留个人资料卡片内容。
