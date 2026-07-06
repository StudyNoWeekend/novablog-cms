# 文章管理功能任务列表

## 后端任务

- [x] Task 1: 数据库迁移 — 创建 articles/categories/tags/article_tags/media 五张表
  - [x] 创建 002_article_migration.up.sql：articles 表（字段：id, title, slug, summary, content, cover_image, category_id, status, type, article_type, extra, view_count, comment_count, is_top, is_comment, published_at, created_at, updated_at, deleted_at）
  - [x] 创建 categories 表
  - [x] 创建 tags 表
  - [x] 创建 article_tags 关联表
  - [x] 创建 media 表（字段：id, filename, file_type, mime_type, size, url, thumb_url, width, height, storage_path, created_at, deleted_at）
  - [x] 创建 002_article_migration.down.sql 回滚脚本
  - [x] 创建必要索引（articles_status, articles_category, articles_published, articles_slug, media_type, comments_target）

- [x] Task 2: 后端 Model 层 — GORM 模型定义
  - [x] model/article.go：Article 结构体 + ArticleModel（CRUD + 列表查询 + 状态更新 + 软删除 + slug 查重）
  - [x] model/category.go：Category 结构体 + CategoryModel（CRUD + 名称查重 + 按 sort_order 排序列表）
  - [x] model/tag.go：Tag 结构体 + TagModel（CRUD + 名称查重）+ ArticleTag 结构体 + ArticleTagModel（批量关联/替换关联）
  - [x] model/media.go：Media 结构体 + MediaModel（创建 + 列表查询 + 删除）

- [x] Task 3: 后端 DTO 层 — 请求/响应结构体
  - [x] dto/req/page_req.go：通用分页请求（Page, PageSize）
  - [x] dto/req/article_req.go：CreateArticleReq, UpdateArticleReq, ArticleListReq, UpdateStatusReq
  - [x] dto/req/category_req.go：CreateCategoryReq, UpdateCategoryReq
  - [x] dto/req/tag_req.go：CreateTagReq, UpdateTagReq
  - [x] dto/res/page_res.go：通用分页响应（List, Total, Page, PageSize, TotalPages）
  - [x] dto/res/article_res.go：ArticleRes（列表项）, ArticleDetailRes（详情）
  - [x] dto/res/category_res.go：CategoryRes
  - [x] dto/res/tag_res.go：TagRes
  - [x] dto/res/media_res.go：MediaRes

- [x] Task 4: 后端 Logic 层 — 媒体上传业务逻辑
  - [x] logic/media_logic.go：UploadFile（接收 multipart file → 存本地 → 写 DB → 返回 URL）, GetList, GetByID, Delete

- [x] Task 5: 后端 Logic 层 — 分类标签业务逻辑
  - [x] logic/category_logic.go：CRUD + 删除前关联检查
  - [x] logic/tag_logic.go：CRUD + 批量查询

- [x] Task 6: 后端 Logic 层 — 文章业务逻辑
  - [x] logic/article_logic.go：Create（自动生成 slug + 事务处理分类/标签关联）, Update（含标签关联更新）, GetList（条件筛选+分页+批量加载分类/标签名）, GetDetail（含分类/标签信息）, UpdateStatus, Delete

- [x] Task 7: 后端 Controller + Router 层
  - [x] controller/media_controller.go：POST /media/upload, GET /media, DELETE /media/:id
  - [x] controller/category_controller.go：CRUD 4 个接口
  - [x] controller/tag_controller.go：CRUD 4 个接口
  - [x] controller/article_controller.go：CRUD + 状态变更 6 个接口
  - [x] controller/public_controller.go：GET /public/articles, GET /public/articles/:slug, GET /public/categories, GET /public/tags
  - [x] router/article.go：注册文章路由组
  - [x] router/category.go：注册分类路由组
  - [x] router/media.go：注册媒体路由组
  - [x] 更新 router/router.go：注入新 controller 和路由

- [x] Task 8: 后端缓存层
  - [x] cache/article_cache.go：文章列表/详情缓存，更新时清除

## 前端任务

- [x] Task 9: 前端类型定义 + API 层
  - [x] types/article.ts：Article, ArticleFilters, ArticleCreateReq, PaginatedData 类型
  - [x] types/category.ts：Category 类型
  - [x] types/tag.ts：Tag 类型
  - [x] api/article.ts：文章 API（getList, getDetail, create, update, remove, updateStatus）
  - [x] api/category.ts：分类 API（getList, create, update, remove）
  - [x] api/tag.ts：标签 API（getList, create, update, remove）
  - [x] api/media.ts：媒体 API（upload, getList, remove）
  - [x] api/setup.ts：修改 API 基础路径常量（如有需要）

- [x] Task 10: 前端 Store 层 + Composables
  - [x] stores/article.ts：articleStore（list, total, current, filters, pagination + fetchList, fetchDetail, create, update, remove, updateStatus actions）
  - [x] composables/usePagination.ts：分页逻辑（page, pageSize, total, handlePageChange）
  - [x] composables/useDebounce.ts：防抖函数

- [x] Task 11: 前端 Markdown 编辑器组件
  - [x] 安装 bytemd 依赖（@bytemd/vue-next, @bytemd/plugin-gfm, @bytemd/plugin-highlight, bytemd）
  - [x] components/editor/MarkdownEditor.vue：ByteMD 编辑器封装，支持 v-model 双向绑定，粘贴/拖拽图片自动上传，GMF + 代码高亮插件，中文语言包

- [x] Task 12: 前端文章列表页
  - [x] components/article/ArticleCard.vue：文章卡片组件（封面/标题/摘要/状态标签/分类/时间/操作栏）
  - [x] 重写 views/article/ArticleListView.vue：筛选栏（状态/分类/搜索） + 卡片网格 + 分页 + 写文章按钮 + 空状态 + 加载骨架屏

- [x] Task 13: 前端文章创建/编辑页
  - [x] components/article/ArticleForm.vue：文章编辑表单组件（标题/摘要/封面/分类/标签/状态/置顶/评论开关）
  - [x] views/article/ArticleCreateView.vue：新建文章页（编辑器 + ArticleForm + 保存草稿/发布按钮）
  - [x] views/article/ArticleEditView.vue：编辑文章页（复用 ArticleCreateView 逻辑）
  - [x] 更新 router/routes.ts：添加 articles/create 和 articles/:id/edit 路由

- [x] Task 14: 前端分类标签管理页
  - [x] components/article/CategorySelect.vue：分类选择器（下拉搜索）
  - [x] 重写 views/category/CategoryManageView.vue：左右分栏（分类列表+CRUD | 标签列表+CRUD）

# 任务依赖

- Task 2 依赖 Task 1（先建表再写 Model）
- Task 3 可与 Task 2 并行
- Task 4 依赖 Task 2（Media model）
- Task 5 依赖 Task 2（Category/Tag model）
- Task 6 依赖 Task 2, 3, 5
- Task 7 依赖 Task 4, 5, 6
- Task 8 可与 Task 7 并行
- Task 9 可与 Task 2~8 并行（纯前端，无需后端完成）
- Task 10 依赖 Task 9
- Task 11 可与 Task 10 并行
- Task 12 依赖 Task 9, 10
- Task 13 依赖 Task 9, 10, 11
- Task 14 依赖 Task 9, 10