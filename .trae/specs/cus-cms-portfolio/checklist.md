# 验收清单

## 后端 - 数据库迁移
- [x] `migrations/005_portfolio.up.sql` 创建了 `portfolios` 表，包含 id(UUID PK)、name、description、cover_preset_id、status、sort_order、created_at、updated_at、deleted_at 字段
- [x] `migrations/005_portfolio.up.sql` 创建了 `portfolio_items` 表，包含 id(UUID PK)、portfolio_id、preset_id、title、description、sort_order、created_at、updated_at、deleted_at 字段
- [x] 两张表均建立了必要索引（portfolio_id、preset_id、deleted_at）
- [x] `portfolio_items.portfolio_id` 与 `portfolio_items.preset_id` 设置了外键约束
- [x] `migrations/005_portfolio.down.sql` 能正确回滚两张表

## 后端 - Model 层
- [x] `Portfolio` 与 `PortfolioItem` 结构体字段与数据库表一致，主键为 UUID
- [x] `PortfolioModel` 提供 `Create`、`GetByID`、`GetList`、`Update`、`SoftDelete`、`CountItems` 方法
- [x] `PortfolioItemModel` 提供 `Create`、`GetByPortfolioID`（JOIN media_presets 取 output_url）、`GetByID`、`Update`、`SoftDelete`、`SoftDeleteByPortfolioID`、`MaxSortOrder`、`BatchUpdateSort` 方法
- [x] 所有方法接收 `context.Context` 并使用 `WithContext(ctx)`

## 后端 - DTO 层
- [x] `CreatePortfolioReq` 含 name（必填）、description、cover_preset_id
- [x] `UpdatePortfolioReq` 含可更新字段
- [x] `PortfolioListReq` 继承 `PageReq`，含 keyword、status
- [x] `CreatePortfolioItemReq` 含 preset_id（必填）、title（必填）、description
- [x] `UpdatePortfolioItemReq` 含可更新字段
- [x] `SortPortfolioItemsReq` 含 items 数组（id + sort_order）
- [x] `PortfolioRes` 含 item_count 字段
- [x] `PortfolioDetailRes` 含 items 数组，每项含 output_url 等预设字段

## 后端 - Logic 层
- [x] `PortfolioLogic` 依赖 `PortfolioModel`、`PortfolioItemModel`、`MediaPresetModel`，通过构造函数注入
- [x] `AddItem` 校验 preset_id 对应的预设存在
- [x] `AddItem` 自动计算 sort_order（当前最大值 +1）
- [x] `Delete` 在事务内级联软删除作品集及其下所有作品项
- [x] `SortItems` 在事务内批量更新 sort_order
- [x] `GetList` 批量统计每个作品集的作品数量
- [x] 不依赖 Gin Context，所有方法接收 `context.Context`

## 后端 - Controller 层
- [x] 所有 handler 使用 `ShouldBindQuery`（GET）或 `ShouldBindJSON`（POST/PUT/DELETE）
- [x] 所有 handler 使用 `response.Success` / `response.Error` 统一输出
- [x] 不直接调用 Model，只调用 Logic

## 后端 - Router 层
- [x] `RegisterPortfolioRoutes` 注册了 9 个路由（作品集 CRUD 5 个 + 作品项 4 个）
- [x] 所有路由经过 authMiddleware
- [x] `router.go` 中初始化了 `PortfolioController` 并调用了 `RegisterPortfolioRoutes`
- [x] 路由路径符合 `/api/v1/portfolios` 前缀

## 后端 - 编译验证
- [x] `go build ./...` 编译通过
- [x] `go vet ./...` 无错误

## 前端 - 类型与 API
- [x] `types/portfolio.ts` 定义了完整的类型
- [x] `api/portfolio.ts` 封装了 9 个 API 方法
- [x] API 方法使用 `request` 实例，路径与后端一致

## 前端 - 作品选择器组件
- [x] `PortfolioItemPicker.vue` 实现三步选择流程
- [x] 步骤一从媒体库图片类型（file_type=1）中选择原图，支持分页与搜索
- [x] 步骤二调用 `GET /api/v1/media/:id/presets` 获取预设列表，单选
- [x] 步骤三填写作品名称与作品介绍
- [x] 原图无预设时给出提示并禁用确认
- [x] 通过 `emit('selected', { preset_id, title, description })` 返回结果

## 前端 - 作品集列表页
- [x] `PortfolioListView.vue` 显示页头标题与「新建作品集」按钮
- [x] 工具栏含搜索框
- [x] 网格卡片展示封面图、名称、介绍摘要、作品数量
- [x] 卡片支持编辑与删除操作
- [x] 分页组件正常工作
- [x] 新建作品集弹窗含名称与介绍字段
- [x] 删除操作有二次确认
- [x] 创建成功后跳转到编辑页

## 前端 - 作品集编辑页
- [x] `PortfolioEditView.vue` 顶部可编辑作品集名称与介绍
- [x] 内容区按 sort_order 展示作品项列表
- [x] 每个作品项显示成品图预览、作品名称、作品介绍
- [x] 作品项支持编辑与删除
- [x] 「添加作品」按钮打开 `PortfolioItemPicker`
- [x] 支持拖拽排序，拖拽结束后调用排序接口
- [x] 预设已删除的作品项显示占位图与提示

## 前端 - 路由
- [x] `routes.ts` 新增 `portfolios/:id/edit` 路由，`hidden: true`
- [x] 列表页通过 `router.push` 跳转编辑页
- [x] 侧边栏菜单仍只显示「摄影作品集」入口

## 前端 - 构建验证
- [x] `npm run build` 编译通过，无 TypeScript 错误

## 安全与规范
- [x] 所有接口经过认证中间件
- [x] 主键使用 UUID，不暴露自增 ID
- [x] 错误信息不暴露内部堆栈
- [x] 软删除使用 `gorm.DeletedAt`
- [x] 排序字段使用白名单或固定字段（sort_order）
- [x] 事务范围合理，无长事务
