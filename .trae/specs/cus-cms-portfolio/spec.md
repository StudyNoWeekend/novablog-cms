# 摄影作品集模块规格说明

## Why
当前 CMS 已具备媒体库与图片预设能力（`media` 原图表 + `media_presets` 成品图表），用户可以为一张原图生成多张不同相框/EXIF 参数的成品图。但目前缺少将这些成品图组织成「作品集」对外展示的能力。本规格新增摄影作品集模块，允许用户创建带名称和介绍的作品集，并从媒体库中图片类型的预设成品图中挑选作品加入集合，每张作品可单独设置作品名称与作品介绍。

## What Changes
- 新增数据表 `portfolios`（作品集）与 `portfolio_items`（作品集中的作品项）。
- 后端新增 portfolio 模块的完整 CRUD（model / logic / controller / router / dto）。
- 后端 `portfolio_items` 通过 `preset_id` 关联 `media_presets` 表，引用已有预设成品图，不重复存储图片数据。
- 前端实现 `PortfolioListView.vue` 列表页与 `PortfolioEditView.vue` 编辑页，支持作品集创建、编辑、删除，以及作品项的添加、排序、删除。
- 前端新增作品选择器组件，从媒体库图片类型中选择某张原图的某个预设加入作品集。
- 前端新增 `api/portfolio.ts` 与类型定义。
- 前端路由新增作品集编辑页路由（隐藏菜单）。
- 数据库迁移文件 `005_portfolio.up.sql` / `005_portfolio.down.sql`。

## Impact
- 新增后端文件：
  - `internal/model/portfolio.go`
  - `internal/dto/req/portfolio_req.go`
  - `internal/dto/res/portfolio_res.go`
  - `internal/logic/portfolio_logic.go`
  - `internal/controller/portfolio_controller.go`
  - `internal/router/portfolio.go`
  - `migrations/005_portfolio.up.sql`
  - `migrations/005_portfolio.down.sql`
- 受影响后端文件：
  - `internal/router/router.go`（注册 portfolio 路由）
- 新增前端文件：
  - `cus-cms-web/src/api/portfolio.ts`
  - `cus-cms-web/src/views/portfolio/PortfolioEditView.vue`
  - `cus-cms-web/src/components/portfolio/PortfolioItemPicker.vue`（从媒体图片预设中挑选作品）
  - `cus-cms-web/src/types/portfolio.ts`
- 受影响前端文件：
  - `cus-cms-web/src/views/portfolio/PortfolioListView.vue`（重建为完整页面）
  - `cus-cms-web/src/router/routes.ts`（新增编辑路由）
  - `cus-cms-web/src/types/api.ts`（可选，若复用通用类型）

## ADDED Requirements

### Requirement: 作品集数据模型
系统 SHALL 新增 `portfolios` 表，存储作品集名称、介绍、排序、发布状态等元数据；并新增 `portfolio_items` 表，存储作品集下的每一项作品，通过 `preset_id` 关联 `media_presets` 表的成品图，并为每项提供单独的作品名称与作品介绍。

#### Scenario: 作品集与作品项的关系
- **GIVEN** 一个已创建的作品集（portfolio 记录）
- **WHEN** 用户向该作品集添加一个作品项
- **THEN** 系统在 `portfolio_items` 表插入一条记录，`portfolio_id` 指向所属作品集
- **AND** `preset_id` 指向 `media_presets` 表中的一条成品图记录
- **AND** 该作品项可单独设置 `title`（作品名称）与 `description`（作品介绍）

#### Scenario: 字段约束
- `portfolios.id` / `portfolio_items.id` 使用 UUID 主键
- `portfolio_items.preset_id` 必须引用未软删除的 `media_presets` 记录
- `portfolio_items.sort_order` 用于前端展示排序，默认按添加顺序递增
- `portfolios.deleted_at` / `portfolio_items.deleted_at` 软删除

### Requirement: 创建作品集接口
后端 SHALL 提供 `POST /api/v1/portfolios` 接口，创建一个新作品集，接收 `name`（必填）、`description`（可选）、`cover_preset_id`（可选封面预设 ID）。

#### Scenario: 创建成功
- **WHEN** 前端提交合法的名称
- **THEN** 后端在 `portfolios` 表插入一条记录
- **AND** 返回新建作品集信息（含 ID、名称、介绍、创建时间）

#### Scenario: 名称缺失
- **WHEN** 前端未提交 `name` 或 `name` 为空
- **THEN** 后端返回参数错误

### Requirement: 作品集列表接口
后端 SHALL 提供 `GET /api/v1/portfolios` 接口，分页返回作品集列表，支持按 `keyword` 模糊搜索名称，按 `status` 过滤状态。

#### Scenario: 分页查询
- **WHEN** 前端请求 `GET /api/v1/portfolios?page=1&page_size=20`
- **THEN** 后端返回 `list`、`total`、`page`、`page_size`、`total_pages`
- **AND** 列表项包含每个作品集的作品数量 `item_count`

### Requirement: 作品集详情接口
后端 SHALL 提供 `GET /api/v1/portfolios/:id` 接口，返回作品集元数据及其下所有作品项列表（按 `sort_order` 升序）。

#### Scenario: 查询详情
- **WHEN** 前端请求某作品集详情
- **THEN** 后端返回作品集信息
- **AND** 返回 `items` 数组，每项包含 `id`、`preset_id`、`title`、`description`、`sort_order`、`output_url`（来自关联预设的成品图 URL）、`mime_type`、`output_size`

#### Scenario: 作品集不存在
- **WHEN** 前端请求不存在的作品集 ID
- **THEN** 后端返回 404 资源不存在

### Requirement: 更新作品集接口
后端 SHALL 提供 `PUT /api/v1/portfolios/:id` 接口，更新作品集的名称、介绍、封面、状态。

#### Scenario: 更新成功
- **WHEN** 前端提交合法参数
- **THEN** 后端更新对应记录
- **AND** 返回更新后的作品集信息

### Requirement: 删除作品集接口
后端 SHALL 提供 `DELETE /api/v1/portfolios/:id` 接口，软删除作品集，并级联软删除其下所有作品项。

#### Scenario: 删除成功
- **WHEN** 前端请求删除某作品集
- **THEN** 后端软删除该作品集
- **AND** 软删除其下所有 `portfolio_items` 记录

### Requirement: 添加作品项接口
后端 SHALL 提供 `POST /api/v1/portfolios/:id/items` 接口，向指定作品集添加一个作品项，接收 `preset_id`（必填，引用 `media_presets.id`）、`title`（作品名称，必填）、`description`（作品介绍，可选）。

#### Scenario: 添加成功
- **WHEN** 前端提交合法的 `preset_id` 与 `title`
- **THEN** 后端校验 `preset_id` 对应的预设存在
- **AND** 在 `portfolio_items` 表插入记录，`sort_order` 取当前作品集下最大值 +1
- **AND** 返回新建作品项信息（含关联预设的成品图 URL）

#### Scenario: 预设不存在
- **WHEN** 前端提交的 `preset_id` 在 `media_presets` 中不存在（或已软删除）
- **THEN** 后端返回 404 资源不存在

#### Scenario: 作品集不存在
- **WHEN** 前端向不存在的作品集添加作品项
- **THEN** 后端返回 404 资源不存在

### Requirement: 更新作品项接口
后端 SHALL 提供 `PUT /api/v1/portfolios/:id/items/:itemId` 接口，更新作品项的名称、介绍、关联预设。

#### Scenario: 更新成功
- **WHEN** 前端提交合法参数
- **THEN** 后端更新对应 `portfolio_items` 记录
- **AND** 返回更新后的作品项信息

### Requirement: 删除作品项接口
后端 SHALL 提供 `DELETE /api/v1/portfolios/:id/items/:itemId` 接口，软删除指定作品项。

#### Scenario: 删除成功
- **WHEN** 前端请求删除某作品项
- **THEN** 后端软删除该 `portfolio_items` 记录

### Requirement: 作品项排序接口
后端 SHALL 提供 `PUT /api/v1/portfolios/:id/items/sort` 接口，批量更新作品项的 `sort_order`。

#### Scenario: 排序成功
- **WHEN** 前端提交 `{ "items": [{ "id": "...", "sort_order": 1 }, ...] }`
- **THEN** 后端在事务内更新所有作品项的 `sort_order`
- **AND** 返回成功

### Requirement: 前端作品集列表页
前端 SHALL 实现 `PortfolioListView.vue`：
- 顶部页头含「摄影作品集」标题与「新建作品集」按钮。
- 工具栏含搜索框。
- 内容区以网格卡片形式展示作品集，每张卡片显示封面图（取 `cover_preset_id` 关联预设的成品图 URL，若无则取第一个作品项的成品图，再无则占位）、名称、介绍摘要、作品数量。
- 卡片操作：编辑、删除。
- 底部分页组件。

#### Scenario: 创建作品集
- **WHEN** 用户点击「新建作品集」
- **THEN** 弹出表单（名称、介绍），提交后调用 `POST /api/v1/portfolios`
- **AND** 成功后刷新列表并跳转到编辑页

#### Scenario: 删除作品集
- **WHEN** 用户点击卡片「删除」并确认
- **THEN** 调用 `DELETE /api/v1/portfolios/:id`
- **AND** 成功后从列表移除

### Requirement: 前端作品集编辑页
前端 SHALL 实现 `PortfolioEditView.vue`：
- 顶部展示作品集名称与介绍（可编辑）。
- 内容区展示当前作品集下的所有作品项，按 `sort_order` 排序，每项显示成品图预览、作品名称、作品介绍。
- 作品项支持编辑（名称、介绍）、删除、拖拽排序。
- 提供「添加作品」按钮，打开作品选择器。

#### Scenario: 添加作品
- **WHEN** 用户点击「添加作品」
- **THEN** 打开 `PortfolioItemPicker.vue` 选择器
- **AND** 用户从媒体库图片中选择一张原图，再选择该原图下的一个预设
- **AND** 填写作品名称与作品介绍
- **AND** 提交后调用 `POST /api/v1/portfolios/:id/items`

#### Scenario: 编辑作品
- **WHEN** 用户点击某作品项的「编辑」
- **THEN** 弹出表单（作品名称、作品介绍、可重新选择预设）
- **AND** 提交后调用 `PUT /api/v1/portfolios/:id/items/:itemId`

#### Scenario: 拖拽排序
- **WHEN** 用户拖拽作品项调整顺序
- **THEN** 前端即时更新视图
- **AND** 拖拽结束后调用 `PUT /api/v1/portfolios/:id/items/sort` 持久化

### Requirement: 作品选择器组件
前端 SHALL 实现 `PortfolioItemPicker.vue`：
- 步骤一：从媒体库的图片类型（`file_type=1`）中选择一张原图，支持分页与搜索。
- 步骤二：选择该原图下的某个预设（调用 `GET /api/v1/media/:id/presets`），展示预设缩略图。
- 步骤三：填写作品名称与作品介绍，确认后通过 `emit('selected', { preset_id, title, description })` 返回。

#### Scenario: 无预设的原图
- **WHEN** 用户选择一张原图但该原图没有任何预设
- **THEN** 提示「该图片暂无预设，请先在媒体库创建预设」
- **AND** 禁用确认按钮

## MODIFIED Requirements
无。本规格全部为新增能力。

## REMOVED Requirements
无。

## 约束与注意事项
- `portfolio_items.preset_id` 仅引用 `media_presets` 表记录，不复制图片数据，删除作品项不影响预设本身。
- 删除作品集时级联软删除其下所有作品项（在 logic 层事务内完成）。
- 删除预设（`media_presets`）时不级联删除引用它的 `portfolio_items`，由后续清理任务处理；前端在作品集详情中遇到预设已删除的项，显示占位图与提示。
- 作品项排序接口必须在事务内完成，避免并发导致 `sort_order` 冲突。
- 所有接口必须经过认证中间件保护。
- 前端作品集编辑页路由不在侧边栏菜单显示（`hidden: true`），仅通过列表页跳转进入。
- 主键使用 UUID，禁止对外暴露自增 ID。
