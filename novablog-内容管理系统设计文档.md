# Cus 内容管理系统（CMS）设计文档

## 一、项目拆分总览

Cus博客系统整体拆分为 **3个独立项目**，各自职责清晰、前后端分离、通过RESTful API交互。

| 项目 | 名称 | 定位 | 目标用户 | 技术栈 |
|------|------|------|----------|--------|
| **Project 1** | `cus-blog-web` | 博客Web展示页面 | 访客/读者 | 前端框架（React/Vue） |
| **Project 2** | `novablog` | 内容管理系统（CMS） | 博主 | Go + Gin + GORM + PostgreSQL + Redis |
| **Project 3** | `cus-official` | 博客官方网站 | 潜在用户/开发者 | 静态站点/前端框架 |

### 各项目边界

```
┌─────────────────────────────────────────────────────────────────┐
│                     cus-official（官方网站）                      │
│  • 系统介绍 / 部署文档 / GitHub链接 / 博主案例展示                │
│  • 纯展示型，无业务数据交互                                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     novablog（内容管理系统）                       │
│  • 博主登录后台 → 管理文章/作品/媒体/评论/模板                    │
│  • 提供 RESTful API 供 blog-web 调用                              │
│  • 核心数据中心，负责数据持久化与业务规则                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    cus-blog-web（博客展示页）                     │
│  • 面向访客的博客前端，调用 CMS API 获取数据                       │
│  • 文章阅读 / 作品集浏览 / 视频播放 / 评论互动                     │
│  • 模板渲染 / 响应式设计 / SEO优化                               │
└─────────────────────────────────────────────────────────────────┘
```

> **设计原则**：CMS 作为唯一的**数据与业务核心**，blog-web 与 official 均为**消费端**，不直接操作数据库，所有数据通过 CMS API 获取。

---

## 二、CMS 系统概述

### 2.1 定位

novablog 是面向**单个博主**的后台内容管理系统，一套系统只有一个博主，无需多用户/多角色管理。负责：

- 博主身份认证（JWT Token 校验）
- 博客内容的创建、编辑、发布、归档
- 多媒体资源（图片/视频/音频）的上传与管理
- 职业作品集（摄影/视频剪辑）的管理
- 评论的审核与互动管理
- 博客模板风格的选择与切换
- 向 blog-web 提供数据查询 API

### 2.2 技术栈

| 层级 | 技术选型 | 说明 |
|------|----------|------|
| 语言/框架 | Go 1.22+ / Gin | HTTP Web 框架 |
| ORM | GORM v2 | 数据库操作 |
| 数据库 | PostgreSQL 15+ | 主数据库，支持JSONB、全文检索 |
| 缓存 | Redis 7+ | 会话、热点数据、限流、分布式锁 |
| 对象存储 | MinIO / 阿里云OSS | 多媒体文件存储 |
| 搜索 | PostgreSQL Full-Text Search | 初期内置，后续可扩展ES |
| 配置 | YAML + env | 环境隔离配置 |
| 日志 | zap | 结构化日志 |
| 文档 | swaggo/swag | Swagger API 文档 |

---

## 三、CMS 功能模块拆解

### 3.1 模块总览

```
CMS 系统
├── 认证模块 (auth)
├── 博主资料模块 (profile)
├── 内容管理
│   ├── 文章管理模块 (article)
│   ├── 分类标签模块 (category)
│   ├── 摄影作品集模块 (portfolio)
│   ├── 视频作品模块 (video)
│   ├── 旅行攻略模块 (travel)
│   └── 音乐播放列表模块 (music)
├── 媒体资源模块 (media)           ← 独立，作为素材仓库
├── 评论管理模块 (comment)
├── 模板管理模块 (template)
├── 系统配置模块 (system)
└── 数据统计模块 (analytics)
```

---

### 3.2 认证模块 (auth)

> **设计说明**：单博主系统，无需注册和 RBAC。博主账号通过初始化脚本或环境变量创建，登录后获取 JWT Token，所有后台接口通过 Token 校验即可。

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 博主登录 | 用户名+密码登录，返回 JWT Token | P0 |
| Token刷新 | Access Token + Refresh Token 双令牌机制 | P0 |
| 登出 | Redis 存储 Token 黑名单，支持登出 | P0 |
| 密码修改 | 登录后修改密码（需验证旧密码） | P0 |

**接口列表**：
- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/refresh` - 刷新Token
- `POST /api/v1/auth/logout` - 登出
- `PUT /api/v1/auth/password` - 修改密码

---

### 3.3 博主资料模块 (profile)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 资料查看/编辑 | 昵称、头像、简介、联系方式 | P0 |
| 头像上传 | 上传头像到对象存储 | P1 |
| 博客信息配置 | 博客标题、描述、SEO配置 | P1 |

**接口列表**：
- `GET /api/v1/profile` - 获取博主资料
- `PUT /api/v1/profile` - 更新博主资料
- `POST /api/v1/profile/avatar` - 上传头像
- `GET /api/v1/profile/blog-config` - 获取博客配置
- `PUT /api/v1/profile/blog-config` - 更新博客配置

---

### 3.4 文章管理模块 (article)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 文章创建 | 标题/正文/摘要/封面图，支持Markdown | P0 |
| 文章编辑 | 修改已发布或草稿文章 | P0 |
| 草稿保存 | 自动保存/手动保存草稿 | P0 |
| 文章发布 | 立即发布或定时发布 | P0 |
| 文章状态管理 | 草稿/已发布/已下架 | P0 |
| 文章删除 | 软删除文章 | P0 |
| 文章列表查询 | 按状态/分类/标签/时间筛选（分页） | P0 |
| 文章详情查询 | 获取单篇文章完整内容 | P0 |
| 文章置顶 | 设置文章置顶排序 | P1 |
| 文章SEO设置 | 自定义URL slug / meta描述 | P1 |

**文章状态流转**：

```
草稿 → 已发布 → 已下架
```

**接口列表**：
- `POST /api/v1/articles` - 创建文章
- `GET /api/v1/articles` - 文章列表
- `GET /api/v1/articles/:id` - 文章详情
- `PUT /api/v1/articles/:id` - 更新文章
- `DELETE /api/v1/articles/:id` - 删除文章
- `PUT /api/v1/articles/:id/status` - 修改文章状态（发布/下架/转为草稿）

---

### 3.5 分类标签模块 (category)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 分类管理 | 创建/编辑/删除文章分类 | P0 |
| 标签管理 | 创建/编辑/删除标签 | P0 |
| 文章关联 | 文章绑定分类和多个标签 | P0 |
| 分类树查询 | 支持多级分类（预留） | P2 |

**接口列表**：
- `GET /api/v1/categories` - 分类列表
- `POST /api/v1/categories` - 创建分类
- `PUT /api/v1/categories/:id` - 更新分类
- `DELETE /api/v1/categories/:id` - 删除分类
- `GET /api/v1/tags` - 标签列表
- `POST /api/v1/tags` - 创建标签
- `PUT /api/v1/tags/:id` - 更新标签
- `DELETE /api/v1/tags/:id` - 删除标签

---

### 3.6 媒体资源模块 (media)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 文件上传 | 图片/视频/音频上传，支持分片上传 | P0 |
| 上传进度 | 返回上传任务ID，支持进度查询 | P1 |
| 媒体库管理 | 查看/删除/搜索已上传文件 | P0 |
| 媒体分类 | 按类型/时间/用途分类管理 | P1 |
| 缩略图生成 | 图片自动生成多尺寸缩略图 | P1 |
| 视频转码（可选） | 上传视频后台异步转码为HLS | P2 |

**存储策略**：

| 文件类型 | 存储位置 | 命名规则 |
|----------|----------|----------|
| 图片 | MinIO / OSS | `images/{yyyy}/{mm}/{uuid}.{ext}` |
| 视频 | MinIO / OSS | `videos/{yyyy}/{mm}/{uuid}/{filename}` |
| 音频 | MinIO / OSS | `audios/{yyyy}/{mm}/{uuid}.{ext}` |
| 头像 | MinIO / OSS | `avatars/{uuid}.{ext}` |

**接口列表**：
- `POST /api/v1/media/upload` - 文件上传（预签名/直传）
- `POST /api/v1/media/upload/chunk` - 分片上传
- `GET /api/v1/media` - 媒体列表
- `DELETE /api/v1/media/:id` - 删除媒体
- `GET /api/v1/media/:id` - 媒体详情

---

### 3.7 摄影作品集模块 (portfolio)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 作品集创建 | 标题/描述/封面 | P1 |
| 作品集编辑/删除 | 修改信息或删除 | P1 |
| 照片上传 | 批量上传照片到作品集 | P1 |
| 照片管理 | 排序/删除/编辑照片信息 | P1 |
| 照片元数据 | 光圈/快门/ISO/拍摄设备等 | P2 |
| 作品集列表 | 查询所有作品集 | P1 |

**接口列表**：
- `POST /api/v1/portfolios` - 创建作品集
- `GET /api/v1/portfolios` - 作品集列表
- `GET /api/v1/portfolios/:id` - 作品集详情
- `PUT /api/v1/portfolios/:id` - 更新作品集
- `DELETE /api/v1/portfolios/:id` - 删除作品集
- `POST /api/v1/portfolios/:id/photos` - 上传照片
- `PUT /api/v1/portfolios/:id/photos/:photo_id` - 更新照片信息
- `DELETE /api/v1/portfolios/:id/photos/:photo_id` - 删除照片
- `PUT /api/v1/portfolios/:id/photos/order` - 照片排序

---

### 3.8 视频作品模块 (video)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 视频上传 | 上传视频文件，支持大文件 | P1 |
| 视频信息管理 | 标题/描述/封面/分类 | P1 |
| 视频专辑创建 | 创建专辑，一个视频可归属多个专辑 | P1 |
| 视频列表 | 查询视频/专辑列表 | P1 |
| 视频删除 | 软删除视频 | P1 |
| 转码状态查询 | 查询视频转码进度（如有） | P2 |

**接口列表**：
- `POST /api/v1/videos` - 创建视频记录
- `GET /api/v1/videos` - 视频列表
- `GET /api/v1/videos/:id` - 视频详情
- `PUT /api/v1/videos/:id` - 更新视频
- `DELETE /api/v1/videos/:id` - 删除视频
- `POST /api/v1/video-albums` - 创建专辑
- `GET /api/v1/video-albums` - 专辑列表
- `PUT /api/v1/video-albums/:id` - 更新专辑
- `DELETE /api/v1/video-albums/:id` - 删除专辑
- `PUT /api/v1/video-albums/:id/videos` - 专辑关联视频（批量设置视频列表）

---

### 3.9 旅行攻略模块 (travel)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 攻略创建/编辑 | 同文章管理，标记为攻略类型 | P1 |
| 地图标记 | 关联地图坐标点（JSON存储） | P2 |
| 行程规划 | 多日行程编辑（JSON存储） | P2 |
| 攻略分类 | 国内/国外/亲子游等 | P1 |

> **设计说明**：旅行攻略本质是文章的一种类型，复用文章表，通过 `type` 字段区分，额外信息存储在 `extra` JSONB 字段。

---

### 3.10 音乐播放列表模块 (music)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 播放列表创建 | 命名/描述 | P2 |
| 添加音乐 | 添加第三方音乐链接 | P2 |
| 播放列表管理 | 编辑/删除/排序 | P2 |
| 音乐元数据解析 | 尝试解析链接获取歌曲信息 | P2 |

**支持的第三方平台**：网易云音乐、QQ音乐、Spotify

**接口列表**：
- `POST /api/v1/playlists` - 创建播放列表
- `GET /api/v1/playlists` - 播放列表列表
- `GET /api/v1/playlists/:id` - 播放列表详情
- `PUT /api/v1/playlists/:id` - 更新播放列表
- `DELETE /api/v1/playlists/:id` - 删除播放列表
- `POST /api/v1/playlists/:id/items` - 添加音乐
- `PUT /api/v1/playlists/:id/items/:item_id` - 更新音乐信息
- `DELETE /api/v1/playlists/:id/items/:item_id` - 删除音乐
- `PUT /api/v1/playlists/:id/items/order` - 排序

---

### 3.11 评论管理模块 (comment)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 评论列表 | 查文章/作品集/视频的评论（分页，按 target_type + target_id 筛选） | P0 |
| 评论审核 | 通过/拒绝/标记垃圾评论 | P0 |
| 评论回复 | 博主回复评论 | P0 |
| 评论删除 | 删除不当评论 | P0 |
| 垃圾过滤 | 关键词过滤（可配置） | P1 |

**评论状态**：待审核 / 已通过 / 已拒绝 / 垃圾评论

**接口列表**：
- `GET /api/v1/comments?target_type=1&target_id=xxx` - 评论列表（按目标筛选）
- `PUT /api/v1/comments/:id/status` - 审核评论
- `DELETE /api/v1/comments/:id` - 删除评论
- `POST /api/v1/comments/:id/reply` - 回复评论

---

### 3.12 模板管理模块 (template)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 模板列表 | 查看系统预置模板 | P1 |
| 模板切换 | 选择并应用模板 | P1 |
| 模板预览 | 预览模板效果 | P1 |
| 自定义模板 | 上传/编辑模板代码（可选） | P2 |
| 模板配置 | 模板专属配置项（颜色/布局等） | P1 |

**接口列表**：
- `GET /api/v1/templates` - 模板列表
- `PUT /api/v1/template` - 切换当前模板
- `GET /api/v1/template` - 获取当前模板配置
- `PUT /api/v1/template/config` - 更新模板配置

---

### 3.13 系统配置模块 (system)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 站点基础配置 | 站点名称/Logo/ICP备案等 | P1 |
| SEO配置 | 全局SEO标题/关键词/描述 | P1 |
| 评论配置 | 是否开启评论/是否需要审核 | P1 |
| 第三方集成配置 | B站/音乐平台嵌入配置 | P2 |
| 存储配置 | 对象存储桶/域名配置 | P1 |

**接口列表**：
- `GET /api/v1/settings` - 获取所有配置项
- `GET /api/v1/settings/:key` - 获取单个配置项
- `PUT /api/v1/settings/:key` - 更新配置项

---

### 3.14 数据统计模块 (analytics)

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 文章访问量 | 各文章PV/UV统计 | P2 |
| 热门内容 | 按访问量排序 | P2 |
| 访客统计 | 日/周/月访客趋势 | P2 |
| 来源分析 | 访问来源分布 | P3 |

**接口列表**：
- `GET /api/v1/analytics/overview` - 概览数据（总文章数、总访问量等）
- `GET /api/v1/analytics/articles` - 文章访问量统计
- `GET /api/v1/analytics/visitors` - 访客趋势统计

---

## 四、数据库设计

### 4.1 ER 关系概览

```
blogger (博主，单记录)
├── articles (文章)  ←── categories, article_tags, tags
├── media (媒体资源)
├── portfolios (摄影作品集)  ←── portfolio_photos
├── videos (视频)  ←── video_album_items (多对多) → video_albums
├── playlists (音乐播放列表)  ←── playlist_items
├── comments (评论)  ←── 多态关联 (article/portfolio/video)
└── template_config (模板配置)  ←── templates
```

### 4.2 表结构定义

#### blogger（博主表）

> 单博主系统，此表只有一条记录。通过初始化脚本创建。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 博主唯一标识 |
| username | VARCHAR(50) | UNIQUE, NOT NULL | 登录用户名 |
| password_hash | VARCHAR(255) | NOT NULL | bcrypt哈希密码 |
| nickname | VARCHAR(50) | | 昵称 |
| avatar | VARCHAR(500) | | 头像URL |
| bio | TEXT | | 个人简介 |
| email | VARCHAR(100) | | 邮箱（用于通知） |
| blog_title | VARCHAR(100) | | 博客标题 |
| blog_description | TEXT | | 博客描述 |
| seo_keywords | VARCHAR(500) | | 全局SEO关键词 |
| seo_description | VARCHAR(500) | | 全局SEO描述 |
| last_login_at | TIMESTAMPTZ | | 最后登录时间 |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

#### articles（文章表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| title | VARCHAR(200) | NOT NULL | 标题 |
| slug | VARCHAR(200) | UNIQUE | URL别名 |
| summary | VARCHAR(500) | | 摘要 |
| content | TEXT | NOT NULL | Markdown正文 |
| html_content | TEXT | | 渲染后的HTML |
| cover_image | VARCHAR(500) | | 封面图URL |
| category_id | UUID | FK | 分类 |
| status | SMALLINT | DEFAULT 1 | 1草稿 2已发布 3已下架 |
| type | SMALLINT | DEFAULT 1 | 1普通 2攻略 |
| extra | JSONB | | 扩展字段（攻略信息） |
| view_count | INT | DEFAULT 0 | 浏览数 |
| comment_count | INT | DEFAULT 0 | 评论数 |
| is_top | BOOLEAN | DEFAULT false | 是否置顶 |
| published_at | TIMESTAMPTZ | | 发布时间 |
| created_at | TIMESTAMPTZ | | |
| updated_at | TIMESTAMPTZ | | |
| deleted_at | TIMESTAMPTZ | | 软删除 |

#### tags（标签表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| name | VARCHAR(50) | UNIQUE, NOT NULL | 标签名 |
| created_at | TIMESTAMPTZ | | |

#### article_tags（文章-标签关联表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| article_id | UUID | PK, FK | |
| tag_id | UUID | PK, FK | |

#### categories（分类表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| name | VARCHAR(50) | NOT NULL | 分类名 |
| slug | VARCHAR(50) | UNIQUE | URL别名 |
| description | VARCHAR(255) | | 描述 |
| sort_order | INT | DEFAULT 0 | 排序 |
| created_at | TIMESTAMPTZ | | |

#### media（媒体资源表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| filename | VARCHAR(255) | NOT NULL | 原始文件名 |
| file_type | SMALLINT | NOT NULL | 1图片 2视频 3音频 |
| mime_type | VARCHAR(100) | | MIME类型 |
| size | BIGINT | | 文件大小(字节) |
| url | VARCHAR(500) | NOT NULL | 访问URL |
| thumb_url | VARCHAR(500) | | 缩略图URL |
| width | INT | | 图片宽 |
| height | INT | | 图片高 |
| duration | INT | | 音视频时长(秒) |
| storage_path | VARCHAR(500) | | 存储路径 |
| created_at | TIMESTAMPTZ | | |
| deleted_at | TIMESTAMPTZ | | |

#### portfolios（摄影作品集表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| title | VARCHAR(100) | NOT NULL | 标题 |
| description | TEXT | | 描述 |
| cover_image | VARCHAR(500) | | 封面图 |
| sort_order | INT | DEFAULT 0 | 排序 |
| created_at | TIMESTAMPTZ | | |
| updated_at | TIMESTAMPTZ | | |

#### portfolio_photos（作品集照片表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| portfolio_id | UUID | FK | |
| media_id | UUID | FK | 关联media表 |
| title | VARCHAR(100) | | 照片标题 |
| description | TEXT | | 描述 |
| exif | JSONB | | EXIF元数据（光圈/快门/ISO等） |
| sort_order | INT | DEFAULT 0 | 排序 |
| created_at | TIMESTAMPTZ | | |

#### videos（视频表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| title | VARCHAR(200) | NOT NULL | 标题 |
| description | TEXT | | 描述 |
| cover_image | VARCHAR(500) | | 封面图 |
| media_id | UUID | FK | 关联media表 |
| duration | INT | | 时长 |
| status | SMALLINT | DEFAULT 1 | 1正常 2转码中 3失败 |
| created_at | TIMESTAMPTZ | | |

#### video_albums（视频专辑表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| title | VARCHAR(100) | NOT NULL | 标题 |
| description | TEXT | | 描述 |
| sort_order | INT | DEFAULT 0 | 排序 |
| created_at | TIMESTAMPTZ | | |

#### video_album_items（专辑-视频关联表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| album_id | UUID | PK, FK | |
| video_id | UUID | PK, FK | |
| sort_order | INT | DEFAULT 0 | 在专辑内的排序 |
| created_at | TIMESTAMPTZ | | |

#### playlists（音乐播放列表表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| title | VARCHAR(100) | NOT NULL | 播放列表名称 |
| description | TEXT | | 描述 |
| sort_order | INT | DEFAULT 0 | 排序 |
| created_at | TIMESTAMPTZ | | |
| updated_at | TIMESTAMPTZ | | |

#### playlist_items（播放列表项目表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| playlist_id | UUID | FK | |
| title | VARCHAR(200) | NOT NULL | 歌曲名称 |
| artist | VARCHAR(200) | | 艺术家 |
| platform | VARCHAR(20) | | 平台 netease/qqmusic/spotify |
| source_url | VARCHAR(500) | NOT NULL | 第三方音乐链接 |
| cover_url | VARCHAR(500) | | 封面图 |
| sort_order | INT | DEFAULT 0 | 排序 |
| created_at | TIMESTAMPTZ | | |

#### comments（评论表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| target_type | SMALLINT | NOT NULL, INDEX | 1文章 2作品集 3视频 |
| target_id | UUID | NOT NULL, INDEX | 关联的目标ID |
| parent_id | UUID | FK | 父评论ID（层级回复） |
| visitor_name | VARCHAR(50) | | 游客昵称 |
| visitor_email | VARCHAR(100) | | 游客邮箱 |
| content | TEXT | NOT NULL | 评论内容 |
| status | SMALLINT | DEFAULT 1 | 1待审核 2已通过 3已拒绝 4垃圾 |
| ip | INET | | 评论者IP |
| created_at | TIMESTAMPTZ | | |

#### templates（模板表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| name | VARCHAR(50) | NOT NULL | 模板名称 |
| code | VARCHAR(50) | UNIQUE | 模板编码 |
| description | VARCHAR(255) | | 描述 |
| preview_image | VARCHAR(500) | | 预览图 |
| config_schema | JSONB | | 配置项JSON Schema |
| is_system | BOOLEAN | DEFAULT false | 是否系统预置 |
| status | SMALLINT | DEFAULT 1 | 1启用 2禁用 |
| created_at | TIMESTAMPTZ | | |

#### template_config（模板配置表）

> 单博主系统，此表只有一条记录。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| template_id | UUID | FK | 当前使用模板 |
| config | JSONB | | 模板配置值 |
| updated_at | TIMESTAMPTZ | | |

#### settings（系统配置表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | |
| key | VARCHAR(100) | UNIQUE, NOT NULL | 配置键 |
| value | JSONB | | 配置值 |
| description | VARCHAR(255) | | 描述 |
| updated_at | TIMESTAMPTZ | | |

---

## 五、API 设计规范

### 5.1 接口前缀与认证

- CMS 后台管理 API: `/api/v1/` — **需 JWT Token 校验**
- 开放给 blog-web 的公共 API: `/api/v1/public/` — **无需认证**

> **认证模型**：单博主系统，登录后获取 JWT Token，后台接口仅需验证 Token 有效性即可，无需角色/权限判断。

### 5.2 统一响应格式

```json
{
  "code": 0,
  "msg": "success",
  "data": {},
  "trace_id": "abc123"
}
```

| code | 含义 |
|------|------|
| 0 | 成功 |
| 400xxx | 参数错误 |
| 401xxx | 未认证 |
| 404xxx | 资源不存在 |
| 500xxx | 系统内部错误 |

### 5.3 分页规范

请求：
```
GET /api/v1/articles?page=1&page_size=20
```

响应：
```json
{
  "code": 0,
  "data": {
    "list": [],
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  }
}
```

### 5.4 公共 API（供 blog-web 调用）

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/public/articles` | 公开文章列表（已发布） | 否 |
| GET | `/api/v1/public/articles/:slug` | 文章详情（按slug） | 否 |
| GET | `/api/v1/public/categories` | 分类列表 | 否 |
| GET | `/api/v1/public/tags` | 标签列表 | 否 |
| GET | `/api/v1/public/portfolios` | 作品集列表 | 否 |
| GET | `/api/v1/public/portfolios/:id` | 作品集详情 | 否 |
| GET | `/api/v1/public/videos` | 视频列表 | 否 |
| GET | `/api/v1/public/videos/:id` | 视频详情 | 否 |
| GET | `/api/v1/public/comments?target_type=1&target_id=xxx` | 评论列表（已通过） | 否 |
| POST | `/api/v1/public/comments` | 发表评论（需传 target_type + target_id） | 否 |
| GET | `/api/v1/public/playlists` | 播放列表 | 否 |
| GET | `/api/v1/public/profile` | 博主信息 | 否 |
| GET | `/api/v1/public/template` | 当前模板配置 | 否 |

> **设计说明**：公共 API 只返回已发布/已通过状态的数据，不暴露草稿和待审核内容。

---

## 六、项目目录结构

```
novablog/
├── cmd/
│   └── api/
│       └── main.go                 # HTTP服务入口
├── config/
│   ├── config.yaml                 # 默认配置
│   ├── config.dev.yaml             # 开发环境
│   └── config.prod.yaml            # 生产环境
├── bootstrap/
│   ├── app.go                      # 应用初始化编排
│   ├── db.go                       # 数据库初始化
│   ├── redis.go                    # Redis初始化
│   ├── logger.go                   # 日志初始化
│   └── storage.go                  # 对象存储初始化
├── internal/
│   ├── middleware/
│   │   ├── auth.go                 # JWT认证中间件（Token校验）
│   │   ├── cors.go                 # 跨域处理
│   │   ├── recovery.go             # Panic恢复
│   │   ├── logger.go               # 访问日志
│   │   └── trace.go                # TraceID注入
│   ├── router/
│   │   ├── router.go               # 路由总入口
│   │   ├── auth.go                 # 认证路由
│   │   ├── profile.go              # 博主资料路由
│   │   ├── article.go              # 文章路由
│   │   ├── category.go             # 分类标签路由
│   │   ├── media.go                # 媒体路由
│   │   ├── portfolio.go            # 作品集路由
│   │   ├── video.go                # 视频路由
│   │   ├── comment.go              # 评论路由
│   │   ├── template.go             # 模板路由
│   │   └── public.go               # 公共API路由
│   ├── controller/
│   │   ├── auth_controller.go
│   │   ├── profile_controller.go
│   │   ├── article_controller.go
│   │   ├── category_controller.go
│   │   ├── media_controller.go
│   │   ├── portfolio_controller.go
│   │   ├── video_controller.go
│   │   ├── comment_controller.go
│   │   ├── template_controller.go
│   │   └── public_controller.go
│   ├── dto/
│   │   ├── req/
│   │   │   ├── auth_req.go
│   │   │   ├── profile_req.go
│   │   │   ├── article_req.go
│   │   │   ├── page_req.go         # 分页请求
│   │   │   └── ...
│   │   └── res/
│   │       ├── auth_res.go
│   │       ├── profile_res.go
│   │       ├── article_res.go
│   │       ├── page_res.go         # 分页响应
│   │       └── ...
│   ├── logic/
│   │   ├── auth_logic.go
│   │   ├── profile_logic.go
│   │   ├── article_logic.go
│   │   ├── category_logic.go
│   │   ├── media_logic.go
│   │   ├── portfolio_logic.go
│   │   ├── video_logic.go
│   │   ├── comment_logic.go
│   │   ├── template_logic.go
│   │   └── public_logic.go
│   ├── model/
│   │   ├── blogger.go
│   │   ├── article.go
│   │   ├── tag.go
│   │   ├── category.go
│   │   ├── media.go
│   │   ├── portfolio.go
│   │   ├── portfolio_photo.go
│   │   ├── video.go
│   │   ├── video_album.go
│   │   ├── comment.go
│   │   ├── template.go
│   │   ├── template_config.go
│   │   └── setting.go
│   └── cache/
│       ├── profile_cache.go
│       ├── article_cache.go
│       └── token_cache.go
├── pkg/
│   ├── storage/
│   │   ├── minio.go
│   │   └── oss.go
│   └── email/
│       └── smtp.go
├── utils/
│   ├── response/
│   │   └── response.go             # 统一响应封装
│   ├── hash/
│   │   └── bcrypt.go               # 密码哈希
│   └── jwt/
│       └── jwt.go                  # Token生成与解析
├── enum/
│   └── error.go                    # 错误码定义
├── migrations/
│   ├── 001_init_schema.up.sql
│   └── 001_init_schema.down.sql
├── scripts/
│   ├── migrate.sh
│   └── seed_blogger.sh             # 初始化博主账号
├── logs/                           # 运行时日志（.gitignore）
├── Makefile
├── .golangci.yml
├── go.mod
└── go.sum
```

---

## 七、核心流程说明

### 7.1 认证流程

```
博主登录
    │
    ▼
POST /api/v1/auth/login
    │
    ├── 校验用户名+密码
    │
    ├── 生成 Access Token (2h) + Refresh Token (7d)
    │
    └── 返回 Token 对
```

### 7.2 接口鉴权流程

```
请求进入
    │
    ├── 公共路由 (/api/v1/public/*) → 直接放行
    │
    ├── 后台路由 (/api/v1/*) → Auth中间件校验
    │   ├── 解析 JWT Token
    │   ├── 检查 Token 是否在黑名单（登出场景）
    │   ├── Token 有效 → 放行
    │   └── Token 无效/过期 → 返回 401
    │
    └── 进入 Controller
```

### 7.3 文章发布流程

```
POST /api/v1/articles
    │
    ├── Controller: 参数校验（title, content）
    │
    ├── Logic: 业务处理
    │   ├── 生成slug（如未提供）
    │   └── 处理分类/标签关联
    │
    ├── Model: 数据库事务写入
    │   ├── INSERT articles
    │   ├── INSERT/UPDATE tags
    │   └── INSERT article_tags
    │
    └── 清除文章列表缓存
    │
    返回文章ID
```

### 7.4 文件上传流程

```
POST /api/v1/media/upload
    │
    ├── 生成预签名URL（直传对象存储）
    │   或：服务端接收文件 → 上传到对象存储
    │
    ├── 文件信息写入 media 表
    │
    └── 异步任务：图片生成缩略图（如有）
```

---

## 八、非功能设计要点

### 8.1 安全

- 密码使用 `bcrypt` 哈希存储
- JWT Token 有效期：Access Token 2小时，Refresh Token 7天
- 后台接口统一 JWT 校验，公共接口无需认证
- SQL注入防护：全量使用GORM参数绑定
- XSS防护：用户输入内容转义后存储/输出
- IP 黑名单由管理员手动维护，系统不再自动限流

### 8.2 缓存策略

| 缓存对象 | Key | TTL | 更新时机 |
|----------|-----|-----|----------|
| 博主资料 | `blogger:profile` | 1小时 | 资料更新时删除 |
| 文章列表 | `article:list:{条件}` | 10分钟 | 文章发布/更新时删除 |
| 文章详情 | `article:detail:{id}` | 30分钟 | 文章更新时删除 |
| Token黑名单 | `token:blacklist:{jti}` | Token剩余有效期 | 登出时写入 |

### 8.3 索引设计

```sql
-- articles 表
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_category ON articles(category_id);
CREATE INDEX idx_articles_published ON articles(published_at DESC);
CREATE INDEX idx_articles_slug ON articles(slug);

-- comments 表
CREATE INDEX idx_comments_target ON comments(target_type, target_id);
CREATE INDEX idx_comments_status ON comments(status);

-- media 表
CREATE INDEX idx_media_type ON media(file_type);
```

---

## 九、开发阶段规划

| 阶段 | 周期 | 目标 |
|------|------|------|
| **Phase 1** | 2周 | 项目脚手架 + 博主认证 + 文章管理 |
| **Phase 2** | 2周 | 分类标签 + 媒体上传 + 评论管理 + 公共API |
| **Phase 3** | 2周 | 摄影作品集 + 视频管理 + 模板管理 |
| **Phase 4** | 1周 | 音乐播放列表 + 旅行攻略 + 系统配置 |
| **Phase 5** | 1周 | 数据统计 + 性能优化 + 测试补全 |

---

## 十、接口汇总

### 认证模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 登录 |
| POST | `/api/v1/auth/refresh` | 刷新Token |
| POST | `/api/v1/auth/logout` | 登出 |
| PUT | `/api/v1/auth/password` | 修改密码 |

### 博主资料模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/profile` | 获取博主资料 |
| PUT | `/api/v1/profile` | 更新资料 |
| POST | `/api/v1/profile/avatar` | 上传头像 |
| GET | `/api/v1/profile/blog-config` | 博客配置 |
| PUT | `/api/v1/profile/blog-config` | 更新博客配置 |

### 文章模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/articles` | 创建文章 |
| GET | `/api/v1/articles` | 文章列表 |
| GET | `/api/v1/articles/:id` | 文章详情 |
| PUT | `/api/v1/articles/:id` | 更新文章 |
| DELETE | `/api/v1/articles/:id` | 删除文章 |
| PUT | `/api/v1/articles/:id/status` | 修改状态（发布/下架/草稿） |

### 分类/标签模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST/PUT/DELETE | `/api/v1/categories` | 分类CRUD |
| GET/POST/PUT/DELETE | `/api/v1/tags` | 标签CRUD |

### 媒体模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/media/upload` | 文件上传 |
| GET | `/api/v1/media` | 媒体列表 |
| DELETE | `/api/v1/media/:id` | 删除媒体 |

### 作品集模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/portfolios` | 创建作品集 |
| GET | `/api/v1/portfolios` | 列表 |
| GET | `/api/v1/portfolios/:id` | 详情 |
| PUT | `/api/v1/portfolios/:id` | 更新 |
| DELETE | `/api/v1/portfolios/:id` | 删除 |
| POST | `/api/v1/portfolios/:id/photos` | 上传照片 |

### 视频模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/videos` | 创建视频 |
| GET | `/api/v1/videos` | 列表 |
| PUT | `/api/v1/videos/:id` | 更新 |
| DELETE | `/api/v1/videos/:id` | 删除 |
| POST | `/api/v1/video-albums` | 创建专辑 |
| PUT | `/api/v1/video-albums/:id/videos` | 专辑关联视频 |

### 评论模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/comments` | 评论列表（按target_type+target_id筛选） |
| PUT | `/api/v1/comments/:id/status` | 审核 |
| DELETE | `/api/v1/comments/:id` | 删除 |

### 模板模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/templates` | 模板列表 |
| PUT | `/api/v1/template` | 切换模板 |
| GET | `/api/v1/template` | 当前模板 |
| PUT | `/api/v1/template/config` | 更新配置 |

### 音乐模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/playlists` | 创建播放列表 |
| GET | `/api/v1/playlists` | 播放列表列表 |
| PUT | `/api/v1/playlists/:id` | 更新播放列表 |
| DELETE | `/api/v1/playlists/:id` | 删除播放列表 |
| POST | `/api/v1/playlists/:id/items` | 添加音乐 |
| DELETE | `/api/v1/playlists/:id/items/:item_id` | 删除音乐 |

### 系统配置模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/settings` | 获取所有配置 |
| PUT | `/api/v1/settings/:key` | 更新配置项 |

### 数据统计模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/analytics/overview` | 概览数据 |
| GET | `/api/v1/analytics/articles` | 文章访问量 |
| GET | `/api/v1/analytics/visitors` | 访客趋势 |

### 公共API
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/public/articles` | 公开文章列表 |
| GET | `/api/v1/public/articles/:slug` | 文章详情 |
| GET | `/api/v1/public/portfolios` | 作品集 |
| GET | `/api/v1/public/videos` | 视频 |
| GET | `/api/v1/public/comments` | 评论 |
| POST | `/api/v1/public/comments` | 发表评论 |
| GET | `/api/v1/public/profile` | 博主信息 |
| GET | `/api/v1/public/template` | 模板配置 |