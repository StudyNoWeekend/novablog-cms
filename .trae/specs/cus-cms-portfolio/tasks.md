# Tasks

## 后端

- [x] Task 1: 创建数据库迁移文件
  - [x] SubTask 1.1: 创建 `migrations/005_portfolio.up.sql`，定义 `portfolios` 表（id UUID PK, name, description, cover_preset_id, status, sort_order, created_at, updated_at, deleted_at）与 `portfolio_items` 表（id UUID PK, portfolio_id, preset_id, title, description, sort_order, created_at, updated_at, deleted_at），含外键、索引、注释
  - [x] SubTask 1.2: 创建 `migrations/005_portfolio.down.sql`，回滚上述两张表

- [x] Task 2: 实现 Model 层
  - [x] SubTask 2.1: 创建 `internal/model/portfolio.go`，定义 `Portfolio` 与 `PortfolioItem` 结构体、`PortfolioModel` 与 `PortfolioItemModel` 操作结构体及 `NewPortfolio()` / `NewPortfolioItem()` 构造函数
  - [x] SubTask 2.2: 实现 `PortfolioModel` 方法：`Create`、`GetByID`、`GetList`（分页+keyword+status）、`Update`、`SoftDelete`、`CountItems`（批量统计作品数）
  - [x] SubTask 2.3: 实现 `PortfolioItemModel` 方法：`Create`、`GetByPortfolioID`（含 JOIN media_presets 取 output_url）、`GetByID`、`Update`、`SoftDelete`、`SoftDeleteByPortfolioID`（级联软删除）、`MaxSortOrder`、`BatchUpdateSort`（事务内批量更新 sort_order）

- [x] Task 3: 实现 DTO 层
  - [x] SubTask 3.1: 创建 `internal/dto/req/portfolio_req.go`，定义 `CreatePortfolioReq`、`UpdatePortfolioReq`、`PortfolioListReq`、`CreatePortfolioItemReq`、`UpdatePortfolioItemReq`、`SortPortfolioItemsReq`
  - [x] SubTask 3.2: 创建 `internal/dto/res/portfolio_res.go`，定义 `PortfolioRes`（含 item_count）、`PortfolioListRes`、`PortfolioDetailRes`（含 items 数组）、`PortfolioItemRes`（含 output_url 等预设字段）

- [x] Task 4: 实现 Logic 层
  - [x] SubTask 4.1: 创建 `internal/logic/portfolio_logic.go`，定义 `PortfolioLogic` 结构体（依赖 `PortfolioModel`、`PortfolioItemModel`、`MediaPresetModel`）与 `NewPortfolioLogic()` 构造函数
  - [x] SubTask 4.2: 实现作品集方法：`Create`、`GetList`、`GetByID`（含 items）、`Update`、`Delete`（事务级联软删除 items）
  - [x] SubTask 4.3: 实现作品项方法：`AddItem`（校验预设存在、sort_order 自增）、`UpdateItem`、`DeleteItem`、`SortItems`（事务批量更新）
  - [x] SubTask 4.4: 在 `GetList` 中调用 `CountItems` 批量统计每个作品集的作品数量

- [x] Task 5: 实现 Controller 层
  - [x] SubTask 5.1: 创建 `internal/controller/portfolio_controller.go`，定义 `PortfolioController` 结构体与 `NewPortfolioController()` 构造函数
  - [x] SubTask 5.2: 实现 handler：`Create`、`GetList`、`GetByID`、`Update`、`Delete`、`AddItem`、`UpdateItem`、`DeleteItem`、`SortItems`，统一使用 `ShouldBindQuery` / `ShouldBindJSON` 与 `response.Success` / `response.Error`

- [x] Task 6: 实现 Router 层
  - [x] SubTask 6.1: 创建 `internal/router/portfolio.go`，定义 `RegisterPortfolioRoutes`，挂载以下路由（均经过 authMiddleware）：
    - `POST   /portfolios`
    - `GET    /portfolios`
    - `GET    /portfolios/:id`
    - `PUT    /portfolios/:id`
    - `DELETE /portfolios/:id`
    - `POST   /portfolios/:id/items`
    - `PUT    /portfolios/:id/items/sort`
    - `PUT    /portfolios/:id/items/:itemId`
    - `DELETE /portfolios/:id/items/:itemId`
  - [x] SubTask 6.2: 修改 `internal/router/router.go`，初始化 `PortfolioController` 并调用 `RegisterPortfolioRoutes`

- [x] Task 7: 后端编译验证
  - [x] SubTask 7.1: 在 `cus-cms/` 目录运行 `go build ./...` 确保编译通过
  - [x] SubTask 7.2: 运行 `go vet ./...` 确保无静态检查错误

## 前端

- [x] Task 8: 实现类型与 API 层
  - [x] SubTask 8.1: 创建 `cus-cms-web/src/types/portfolio.ts`，定义 `Portfolio`、`PortfolioItem`、`PortfolioDetail`、`PortfolioListRes`、`CreatePortfolioReq`、`UpdatePortfolioReq`、`CreatePortfolioItemReq`、`UpdatePortfolioItemReq`、`SortPortfolioItemsReq` 等类型
  - [x] SubTask 8.2: 创建 `cus-cms-web/src/api/portfolio.ts`，封装 `portfolioApi`：`create`、`getList`、`getById`、`update`、`remove`、`addItem`、`updateItem`、`removeItem`、`sortItems`

- [x] Task 9: 实现作品选择器组件
  - [x] SubTask 9.1: 创建 `cus-cms-web/src/components/portfolio/PortfolioItemPicker.vue`，实现三步选择流程：
    - 步骤一：媒体库图片列表（`file_type=1`，分页+搜索），单选
    - 步骤二：所选原图的预设列表（`GET /api/v1/media/:id/presets`），单选
    - 步骤三: 填写作品名称与作品介绍
  - [x] SubTask 9.2: 通过 `emit('selected', { preset_id, title, description })` 返回结果，校验必填项

- [x] Task 10: 实现作品集列表页
  - [x] SubTask 10.1: 重写 `cus-cms-web/src/views/portfolio/PortfolioListView.vue`：
    - 页头：标题 + 「新建作品集」按钮
    - 工具栏：搜索框
    - 网格卡片：封面图、名称、介绍摘要、作品数量、编辑/删除操作
    - 分页组件
  - [x] SubTask 10.2: 新建作品集弹窗（名称、介绍），提交后跳转编辑页
  - [x] SubTask 10.3: 删除作品集二次确认

- [x] Task 11: 实现作品集编辑页
  - [x] SubTask 11.1: 创建 `cus-cms-web/src/views/portfolio/PortfolioEditView.vue`：
    - 顶部：作品集名称与介绍（可编辑，保存调用 `PUT /portfolios/:id`）
    - 内容区：作品项列表（按 sort_order 排序），每项显示成品图预览、作品名称、作品介绍、编辑/删除按钮
    - 「添加作品」按钮：打开 `PortfolioItemPicker`
  - [x] SubTask 11.2: 实现作品项编辑弹窗（作品名称、作品介绍、可重选预设）
  - [x] SubTask 11.3: 实现拖拽排序，拖拽结束后调用 `PUT /portfolios/:id/items/sort`
  - [x] SubTask 11.4: 处理预设已删除的异常情况（占位图与提示）

- [x] Task 12: 实现路由与导航
  - [x] SubTask 12.1: 修改 `cus-cms-web/src/router/routes.ts`，新增 `portfolios/:id/edit` 路由（`hidden: true`），指向 `PortfolioEditView.vue`
  - [x] SubTask 12.2: 列表页跳转编辑页使用 `router.push`

- [x] Task 13: 前端构建验证
  - [x] SubTask 13.1: 在 `cus-cms/cus-cms-web/` 目录运行 `npm run build` 确保 TypeScript 编译通过

# Task Dependencies
- Task 2 依赖 Task 1（需要表结构定义参考）
- Task 3 可与 Task 2 并行（DTO 不依赖 Model 实现）
- Task 4 依赖 Task 2 与 Task 3
- Task 5 依赖 Task 4
- Task 6 依赖 Task 5
- Task 7 依赖 Task 6
- Task 9 依赖 Task 8
- Task 10 与 Task 11 可并行（不同页面）
- Task 11 依赖 Task 9（需要作品选择器）
- Task 12 依赖 Task 11
- Task 13 依赖 Task 8-12 全部完成
- 后端 Task 1-7 与前端 Task 8-13 可并行推进
