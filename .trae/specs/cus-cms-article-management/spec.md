# 文章管理功能 Spec

## Why
Cus CMS 需要文章管理核心功能，支持博主创建、编辑、发布文章，管理分类和标签。

## What Changes
- 后端新增 articles、categories、tags、article_tags、media 五张数据库表
- 后端新增文章 CRUD + 状态管理 API
- 后端新增分类标签 CRUD API
- 后端新增媒体上传 API（支持图片上传）
- 后端新增公共 API（文章列表/详情，供 blog-web 消费）
- 前端新增文章列表页（表格/卡片双视图 + 筛选 + 分页）
- 前端新增文章创建/编辑页（双编辑器：Markdown + 富文本）
- 前端新增分类标签管理页
- content 字段存储编辑器原始内容（Markdown 或 HTML），由 type 字段区分
- 渲染逻辑由前端负责，后端不做 Markdown→HTML 转换
- 媒体上传支持粘贴/拖拽图片自动上传

## Impact
- Affected specs: 无（全新功能）
- Affected code: cus-cms/internal/model, controller, logic, router, dto, cache; cus-cms-web/src/views, api, stores, types, components

## ADDED Requirements

### Requirement: 文章 CRUD
系统 SHALL 提供文章的创建、查询、更新、删除功能。

#### Scenario: 创建文章成功
- **WHEN** 博主通过 POST /api/v1/articles 提交标题和正文
- **THEN** 系统创建文章并返回文章 ID，状态默认为草稿

#### Scenario: 按条件查询文章列表
- **WHEN** 博主打后台 GET /api/v1/articles 并传入 status/category_id/keyword/page/page_size 参数
- **THEN** 系统返回分页文章列表，包含文章基本信息（标题/摘要/封面/状态/分类/标签/时间/浏览量）

#### Scenario: 更新文章
- **WHEN** 博主通过 PUT /api/v1/articles/:id 更新文章内容
- **THEN** 系统更新文章并返回最新信息

#### Scenario: 软删除文章
- **WHEN** 博主通过 DELETE /api/v1/articles/:id 删除文章
- **THEN** 系统软删除文章，不物理删除数据

### Requirement: 文章状态管理
系统 SHALL 支持文章状态流转：草稿 → 已发布 → 已下架。

#### Scenario: 发布草稿
- **WHEN** 博主调用 PUT /api/v1/articles/:id/status 将草稿设为已发布
- **THEN** published_at 记录当前时间，文章在公共 API 中可见

#### Scenario: 下架已发布文章
- **WHEN** 博主将已发布文章设为已下架
- **THEN** 文章不再通过公共 API 暴露

### Requirement: 分类管理
系统 SHALL 支持分类的 CRUD，分类为文章的单选归属。

#### Scenario: 创建分类
- **WHEN** 博主通过 POST /api/v1/categories 创建分类
- **THEN** 系统创建分类记录

#### Scenario: 删除分类前检查关联
- **WHEN** 博主尝试删除已被文章引用的分类
- **THEN** 系统拒绝删除并返回错误

### Requirement: 标签管理
系统 SHALL 支持标签的 CRUD，文章和标签为多对多关系，创建/更新文章时传入标签 ID 数组。

#### Scenario: 为文章关联标签
- **WHEN** 创建或更新文章时传入 tag_ids 数组
- **THEN** 系统自动维护 article_tags 关联表

### Requirement: 媒体上传
系统 SHALL 支持图片文件上传，返回可访问的 URL。

#### Scenario: 上传图片
- **WHEN** 用户通过 POST /api/v1/media/upload 上传图片文件
- **THEN** 系统保存文件到本地存储（开发阶段），返回 { id, url } 供编辑器插入使用

#### Scenario: 保存媒体记录
- **WHEN** 文件上传成功
- **THEN** 系统在 media 表中记录文件信息（文件名/类型/大小/URL）

### Requirement: 公共 API
系统 SHALL 通过 /api/v1/public 前缀提供无需认证的公共查询接口供 blog-web 调用。

#### Scenario: 公开文章列表
- **WHEN** 访客请求 GET /api/v1/public/articles
- **THEN** 返回已发布状态的文章列表（分页），不包含草稿和下架文章

#### Scenario: 公开文章详情
- **WHEN** 访客请求 GET /api/v1/public/articles/:slug
- **THEN** 返回文章完整内容

### Requirement: 双编辑器模式
前端 SHALL 支持 Markdown 和富文本两种编辑模式，通过 type 字段区分内容格式。

#### Scenario: Markdown 模式
- **WHEN** 用户选择 Markdown 模式编辑
- **THEN** 使用 ByteMD 编辑器，type=1，content 存储 Markdown 源码

#### Scenario: 富文本模式
- **WHEN** 用户选择富文本模式编辑
- **THEN** content 存储 HTML 字符串，type=2

### Requirement: 编辑器图片插入
前端编辑器 SHALL 支持粘贴和拖拽图片自动上传并插入。

#### Scenario: 粘贴图片自动上传
- **WHEN** 用户在编辑器中粘贴剪贴板中的图片
- **THEN** 系统自动将图片上传到媒体接口，将返回的 URL 插入编辑器

#### Scenario: 拖拽图片自动上传
- **WHEN** 用户拖拽图片文件到编辑器
- **THEN** 系统自动上传并插入图片 URL