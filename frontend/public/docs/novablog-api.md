---
name: novablog-api
description: novablog 全部公开接口完整参考文档，涵盖公开接口（无需认证），包括接口路径、请求方法、输入输出参数、错误码、请求/响应示例等。当用户问及 novablog API、接口文档、公开接口、接口使用方式、接口调用等话题时触发该 skill。
---

# novablog API 接口参考文档

## 目录

- [一、概述](#一概述)
- [二、统一响应格式](#二统一响应格式)
- [三、错误码说明](#三错误码说明)
- [四、安全与访问控制](#四安全与访问控制)
- [五、健康检查接口](#五健康检查接口)
- [六、公开接口（无需认证）](#六公开接口无需认证)
  - [6.1 公共配置下发](#61-公共配置下发)
  - [6.2 博主信息](#62-博主信息)
  - [6.3 文章模块（公开）](#63-文章模块公开)
  - [6.4 分类与标签（公开）](#64-分类与标签公开)
  - [6.5 评论模块（公开）](#65-评论模块公开)
  - [6.6 旅行攻略模块（公开）](#66-旅行攻略模块公开)
  - [6.7 摄影作品集模块（公开）](#67-摄影作品集模块公开)
  - [6.8 项目经历模块（公开）](#68-项目经历模块公开)
  - [6.9 视频作品模块（公开）](#69-视频作品模块公开)
  - [6.10 音乐播放器模块（公开）](#610-音乐播放器模块公开)
  - [6.11 摄影器材模块（公开）](#611-摄影器材模块公开)
  - [6.12 模块开关配置（公开）](#612-模块开关配置公开)
  - [6.13 第三方歌单（公开）](#613-第三方歌单公开)
- [七、数据脱敏规则](#七数据脱敏规则)
- [八、使用限制](#八使用限制)

---

## 一、概述

### 技术栈

- **API 前缀**: `/api/v1`
- **公开路由组**: `/api/v1/public/*`（无需认证）
- **全局中间件**: TraceMiddleware、LoggerMiddleware、RecoveryMiddleware、AccessLogMiddleware、CORSMiddleware
- **IP 黑名单中间件**: 通过 `api.Use(ipBlacklistMiddleware)` 统一应用于全部 `/api/v1/*` 路由（含公开与管理接口），非仅公开路由组

### 接口分类

| 分类 | 路径前缀 | 认证 | 中间件 |
|------|----------|------|--------|
| 健康检查 | `/health`、`/ready` | 无 | 全局中间件 |
| 公开接口 | `/api/v1/public/*` | 无 | IP 黑名单 |

### 分页请求参数（通用）

所有列表接口均支持以下分页参数：

| 参数 | 类型 | 位置 | 必填 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| page | int | query | 否 | 1 | 页码，min=1 |
| page_size | int | query | 否 | 20 | 每页条数，min=1, max=100 |

### 分页响应格式（通用）

```json
{
  "list": [],
  "total": 0,
  "page": 1,
  "page_size": 20,
  "total_pages": 0
}
```

---

## 二、统一响应格式

所有接口均返回统一的 JSON 响应结构：

```json
{
  "data": {},
  "code": 0,
  "msg": "success",
  "trace_id": "xxx"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| data | any | 响应数据，成功时返回；失败时省略（omitempty） |
| code | int | 业务状态码，0 表示成功 |
| msg | string | 响应信息，成功为 "success" |
| trace_id | string | 链路追踪 ID（omitempty） |

### 成功响应

```json
{
  "code": 0,
  "msg": "success",
  "data": { ... },
  "trace_id": "a1b2c3d4"
}
```

### 错误响应

```json
{
  "code": 400001,
  "msg": "请求参数错误",
  "trace_id": "a1b2c3d4"
}
```

---

## 三、错误码说明

| 错误码 | HTTP 状态码 | 错误信息 | 说明 |
|--------|------------|----------|------|
| 0 | 200 | success | 成功 |
| 400 | 400 | (自定义消息) | 通用参数错误（response.Error） |
| 400001 | 400 | 请求参数错误 | 请求参数校验失败（ShouldBindJSON/ShouldBindQuery） |
| 400103 | 400 | 官方地址不合法，请检查输入 | 官方主题市场地址不合法 |
| 403000 | 403 | 无权限访问该资源 | IP 黑名单命中时实际返回的错误码 |
| 403001 | 403 | 系统已初始化，无法重复创建 | 重复初始化博主账号 |
| 403002 | 403 | 您已被暂时限制访问，请稍后再试 | 已定义但当前未使用（预留错误码） |
| 404001 | 404 | 资源不存在 | 资源未找到 |
| 500001 | 500 | 系统内部错误 | 服务器内部异常 |

---

## 四、安全与访问控制

### IP 黑名单中间件

通过 `api.Use(ipBlacklistMiddleware)` 统一应用于全部 `/api/v1/*` 路由（含公开接口与管理接口）：
- 检查客户端 IP 是否在黑名单中
- 若在黑名单中返回 `403000`（HTTP 403，错误信息"无权限访问该资源"）
- 检查异常时放行（fail open）

### IP 访问日志

系统通过全局 AccessLogMiddleware 中间件记录每个请求的：客户端 IP、请求时间、HTTP 方法、请求路径、响应状态码、User-Agent 摘要。日志写入为异步/非阻塞方式，并按配置的保留周期自动清理过期数据。

### 管理员黑名单管理

管理员可通过 `/security/blacklists` 等接口手动添加、查询、更新备注与删除黑名单条目。黑名单仅通过管理接口维护，系统不会自动限流或自动封禁 IP。

### 限流说明

当前后端**未实现限流中间件**（middleware 目录仅含 IP 黑名单）。业务上的限流策略需由客户端或网关层自行实现。

---

## 五、健康检查接口

### GET /health

健康检查，返回 200 表示服务正常运行。

**响应示例:**

```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| status | string | 固定为 "ok" |
| version | string | 应用版本号（构建期 -ldflags 注入，未注入时为 "dev"） |

### GET /ready

就绪检查，返回 200 表示服务已就绪。

**响应示例:**

```json
{
  "status": "ready"
}
```

---

## 六、公开接口（无需认证）

所有公开接口前缀为 `/api/v1/public`，应用 IP 黑名单中间件。

### 6.1 公共配置下发

#### GET /api/v1/public/config

获取公共配置（免鉴权下发，官方市场地址默认值由后端控制）。

**请求参数:** 无

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| market_base_url | string | 官方主题市场地址（DB 自定义值优先，回退出厂默认值） |

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "market_base_url": "http://novablogapi.ditancafebar.cn"
  }
}
```

---

### 6.2 博主信息

#### GET /api/v1/public/blogger

获取博主公开信息（脱敏）。

**请求参数:** 无

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| nickname | string | 昵称 |
| avatar | string | 头像 URL |
| bio | string | 个人简介 |
| email | string | 邮箱 |
| city | string | 所在城市 |
| blog_title | string | 博客标题 |
| blog_description | string | 博客描述 |
| page_background | string | 页面背景图 URL |
| blog_icon | string | 博客 icon 图 URL |
| social_links | array | 社交平台链接数组 [{platform, name, icon, color, url, sort_order}] |
| tags | array | 标签字符串数组 |

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "nickname": "Cus",
    "avatar": "https://example.com/avatar.jpg",
    "bio": "全栈开发者",
    "email": "cus@example.com",
    "city": "上海",
    "blog_title": "Cus Blog",
    "blog_description": "记录技术与生活",
    "page_background": "https://example.com/bg.jpg",
    "blog_icon": "https://example.com/icon.png",
    "social_links": [
      {"platform": "github", "name": "GitHub", "icon": "M12 .297c-6.63...", "color": "#181717", "url": "https://github.com/xxx", "sort_order": 0}
    ],
    "tags": ["Go", "Vue", "摄影"]
  }
}
```

---

### 6.3 文章模块（公开）

#### GET /api/v1/public/articles

获取已发布文章列表（仅返回 status=2 的文章）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数，max=100 |
| category_id | string | 否 | - | 按分类筛选 |
| keyword | string | 否 | - | 模糊搜索标题 |

**响应 data 字段（分页）:**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | array | 文章列表 |
| total | int64 | 总数 |
| page | int | 当前页 |
| page_size | int | 每页条数 |
| total_pages | int | 总页数 |

**list 中每项字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 文章 ID |
| title | string | 标题 |
| slug | string | URL 标识 |
| summary | string | 摘要 |
| cover_image | string | 封面图 URL（相对路径会被 `AfterFind` 钩子自动拼接为绝对 URL） |
| category_id | string | 分类 ID，无分类时为空串 `""` |
| category_name | string | 分类名称，无分类时为空串 `""` |
| tag_ids | []string | 标签 ID 列表，无标签时为 `[]` |
| tag_names | []string | 标签名称列表，无标签时为 `[]` |
| status | int16 | 状态（公开接口固定 2=已发布；1=草稿、3=已下架不会返回） |
| type | int16 | 编辑器类型 |
| view_count | int | 浏览量 |
| comment_count | int | 评论数 |
| is_top | bool | 是否置顶 |
| is_comment | bool | 是否允许评论 |
| published_at | *time | 发布时间，未发布时为 `null` |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则（默认）**: `is_top DESC, published_at DESC, created_at DESC`（置顶优先；同置顶状态时按发布时间倒序）。
> **status 强制**: controller 强制覆盖入参为 2，前端传入 `status` 无效。

---

#### GET /api/v1/public/articles/hot

获取热门文章列表（按 view_count 降序）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| count | int | 否 | 5 | 返回数量 |

**响应:** 文章列表数组（字段同上）

---

#### GET /api/v1/public/articles/random

获取随机推荐文章。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| count | int | 否 | 5 | 返回数量 |

**响应:** 文章列表数组（字段同上）

---

#### GET /api/v1/public/articles/:slug

根据 slug 获取文章详情（仅返回已发布文章）。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| slug | string | 是 | 文章 URL 标识 |

**响应 data 字段:**

包含文章列表项所有字段，另加：

| 字段 | 类型 | 说明 |
|------|------|------|
| content | string | 文章正文 |
| extra | map | 扩展字段（JSONB） |

---

#### GET /api/v1/public/articles/:slug/view

递增文章浏览量（仅对已发布文章生效）。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| slug | string | 是 | 文章 URL 标识 |

**响应 data:** `null`

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

---

### 6.4 分类与标签（公开）

#### GET /api/v1/public/categories

获取分类列表（默认返回 type=article 的分类）。

**请求参数:** 无

**响应 data 字段（数组）:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 分类 ID |
| name | string | 分类名称 |
| slug | string | URL 标识 |
| description | string | 分类描述 |
| type | string | 分类类型 |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |

> **注意:** 公开接口固定返回 type=article 的分类。

---

#### GET /api/v1/public/tags

获取全部标签列表。

**请求参数:** 无

**响应 data 字段（数组）:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 标签 ID |
| name | string | 标签名称 |
| created_at | time | 创建时间 |

---

### 6.5 评论模块（公开）

#### GET /api/v1/public/comments

获取已通过审核的评论列表（仅返回 status=2）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| target_type | string | 否 | - | 目标类型: article / travel_guide |
| target_id | string | 否 | - | 目标 ID |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 评论 ID |
| target_type | string | 目标类型 |
| target_id | string | 目标 ID |
| parent_id | *string | 父评论 ID |
| nickname | string | 评论者昵称 |
| website | string | 评论者网站 |
| content | string | 评论内容 |
| is_blogger | bool | 是否为博主回复 |
| created_at | time | 创建时间 |

> **脱敏:** 公开评论接口不返回 `ip_address`、`blogger_id` 等敏感字段。

---

#### POST /api/v1/public/comments

发表评论（访客）。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| target_type | string | 是 | 目标类型: article / travel_guide |
| target_id | string | 是 | 目标 ID |
| parent_id | string | 否 | 父评论 ID（回复评论时传） |
| nickname | string | 是 | 评论者昵称，1-50 字符 |
| website | string | 否 | 评论者博客地址 |
| content | string | 是 | 评论内容，1-2000 字符 |

**请求示例:**

```json
{
  "target_type": "article",
  "target_id": "550e8400-e29b-41d4-a716-446655440000",
  "nickname": "访客小明",
  "content": "写得很好！"
}
```

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 评论 ID |
| target_type | string | 目标类型 |
| target_id | string | 目标 ID |
| parent_id | *string | 父评论 ID |
| nickname | string | 昵称 |
| website | string | 网站 |
| content | string | 内容 |
| is_blogger | bool | 是否博主 |
| created_at | time | 创建时间 |

---

### 6.6 旅行攻略模块（公开）

#### GET /api/v1/public/travels

获取已发布旅行攻略列表（仅返回 status=2）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| keyword | string | 否 | - | 模糊搜索标题/目的地/摘要 |
| region | string | 否 | - | 按地区筛选 |
| category_id | string | 否 | - | 按分类筛选 |
| days_range | string | 否 | - | 天数范围: 1-3 / 4-7 / 8-14 / 15+ / all |
| sort | string | 否 | - | 排序: views / rating / likes |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 攻略 ID |
| title | string | 标题 |
| summary | string | 摘要 |
| cover_image | string | 封面图 URL（`AfterFind` 自动拼接为绝对 URL） |
| status | int16 | 状态（公开接口固定 2=已发布） |
| destination | string | 目的地 |
| region | string | 地区 |
| category_id | string | 分类 ID，无分类时为空串 `""` |
| category_name | string | 分类名称，无分类时为空串 `""` |
| days | int | 天数 |
| best_month | string | 最佳月份 |
| view_count | int | 浏览量 |
| like_count | int | 点赞数 |
| rating | float64 | 评分 |
| review_count | int | 评价数 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort` 参数可选 `views`/`rating`/`likes`/默认（`created_at DESC`）。
> **status 强制**: controller 强制覆盖为 2，仅返回已发布攻略。

---

#### GET /api/v1/public/travels/hot

获取热门旅行攻略（按 view_count 降序）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| count | int | 否 | 5 | 返回数量 |

**响应:** 旅行攻略列表数组（字段同上）

---

#### GET /api/v1/public/travels/:id

获取已发布旅行攻略详情。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 攻略 ID |

**响应 data 字段:**

包含列表项所有字段，另加：

| 字段 | 类型 | 说明 |
|------|------|------|
| attractions | []map | 景点列表 |
| itinerary | []map | 行程安排 |
| reviews | []map | 评价列表 |

---

#### GET /api/v1/public/travels/:id/view

递增旅行攻略浏览量。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 攻略 ID |

**响应 data:** `null`

---

#### POST /api/v1/public/travels/:id/like

点赞旅行攻略（like_count + 1）。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 攻略 ID |

**响应 data:** `null`

---

### 6.7 摄影作品集模块（公开）

#### GET /api/v1/public/portfolios

获取已发布作品集列表（仅返回 status=1）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| keyword | string | 否 | - | 模糊搜索名称 |
| category_id | string | 否 | - | 按分类筛选 |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| cover_mode | int | 封面模式: 0=首图, 1=独立封面 |
| cover_preset_id | string | 封面预设 ID（独立封面模式下使用；无关联时为空串 `""`） |
| cover_url | string | 封面图 URL（已自动解析为绝对 URL） |
| status | int | 状态（公开接口固定 1=已发布） |
| sort_order | int | 排序 |
| category_id | string | 分类 ID，无分类时为空串 `""` |
| category_name | string | 分类名称，无分类时为空串 `""` |
| item_count | int64 | 作品项数量 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, created_at DESC`。
> **status 强制**: controller 强制覆盖为 1，仅返回已发布作品集。

---

#### GET /api/v1/public/portfolios/:id

获取已发布作品集详情（含作品项列表）。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 作品集 ID |

**响应 data 字段:**

包含列表项所有字段，另加：

| 字段 | 类型 | 说明 |
|------|------|------|
| items | array | 作品项列表 |

**items 中每项字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 作品项 ID |
| portfolio_id | string | 所属作品集 ID |
| preset_id | string | 媒体预设 ID |
| title | string | 标题 |
| description | string | 描述 |
| sort_order | int | 排序 |
| output_url | string | 输出图 URL |
| mime_type | string | MIME 类型 |
| output_size | int64 | 文件大小 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

---

### 6.8 项目经历模块（公开）

#### GET /api/v1/public/projects

获取已发布项目经历列表（仅返回 status=1）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| keyword | string | 否 | - | 模糊搜索标题/简介 |
| category | string | 否 | - | 按领域分类筛选（摄影/视频剪辑/技术开发等） |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 项目 ID |
| title | string | 项目名称 |
| category | string | 领域分类 |
| role | string | 担任角色 |
| client | string | 客户/所属组织 |
| cover_url | string | 封面 URL（`AfterFind` 自动拼接为绝对 URL） |
| summary | string | 一句话简介 |
| description | string | 详细描述 |
| tech_stack | string | 技能/工具标签，逗号分隔 |
| start_date | string | 开始时间（YYYY-MM） |
| end_date | string | 结束时间（YYYY-MM，空表示至今） |
| project_url | string | 项目/作品链接 |
| repo_url | string | 代码仓库链接 |
| status | int | 状态（公开接口固定 1=已发布；0=草稿不返回） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, start_date DESC, created_at DESC`。
> **status 强制**: controller 强制覆盖为 1，仅返回已发布项目经历。

---

#### GET /api/v1/public/projects/:id

获取已发布项目经历详情。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 项目经历 ID |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 项目 ID |
| title | string | 项目名称 |
| category | string | 领域分类 |
| role | string | 担任角色 |
| client | string | 客户/所属组织 |
| cover_url | string | 封面 URL |
| summary | string | 一句话简介 |
| description | string | 详细描述 |
| tech_stack | string | 技能/工具标签，逗号分隔 |
| start_date | string | 开始时间（YYYY-MM） |
| end_date | string | 结束时间（YYYY-MM，空表示至今） |
| project_url | string | 项目/作品链接 |
| repo_url | string | 代码仓库链接 |
| status | int | 状态（1=已发布） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

---

### 6.9 视频作品模块（公开）

#### GET /api/v1/public/videos

获取已发布视频列表（仅返回 status=1）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| keyword | string | 否 | - | 模糊搜索标题 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 视频 ID |
| title | string | 标题 |
| cover_url | string | 封面 URL（`AfterFind` 自动拼接为绝对 URL） |
| description | string | 描述 |
| status | int | 状态（公开接口固定 1=已发布） |
| sort_order | int | 排序 |
| platforms | array | 平台链接列表 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, created_at DESC`。
> **status 强制**: controller 强制覆盖为 1，仅返回已发布视频。

**platforms 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 链接 ID |
| video_id | string | 视频 ID |
| platform | string | 平台名称 |
| url | string | 链接 URL |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

---

#### GET /api/v1/public/videos/:id

获取已发布视频详情（含平台链接）。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 视频 ID |

**响应 data:** 同列表项字段结构

---

### 6.10 音乐播放器模块（公开）

#### GET /api/v1/public/music/songs

获取歌曲列表。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| category_id | string | 否 | - | 按分类筛选 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 歌曲 ID |
| title | string | 标题 |
| artist | string | 艺术家 |
| cover_url | string | 封面 URL |
| bvid | string | B站视频 ID |
| cid | int64 | B站内容 ID |
| source_url | string | 来源 URL |
| source_type | string | 来源类型 |
| category_id | *string | 分类 ID |
| duration | int | 时长（秒） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

---

#### GET /api/v1/public/music/songs/:id

获取歌曲详情。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 歌曲 ID |

**响应 data:** 同列表项字段结构

---

#### GET /api/v1/public/music/audio-url/:song_id

获取歌曲播放地址（返回 B 站官方外链播放器地址，前端用 iframe 内嵌播放）。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| song_id | string | 是 | 歌曲 ID |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | B 站官方外链播放器地址，供 iframe.src 使用 |

**说明：**

- 播放器地址格式：`https://player.bilibili.com/player.html?bvid=<BVID>&autoplay=1&danmaku=0&high_quality=1`。
- 该地址为 B 站官方外链播放器页，需用 `<iframe>` 内嵌播放，而非 `<audio>`/`<video>` 直接引用。
- 不解析 CDN 直链：B 站 CDN（bilivideo/upos 等）对请求来源做 Referer 白名单校验，浏览器直连第三方站点会返回 403；外链播放器不受防盗链与链接时效限制。
- 若前端传入 `autoplay=1`（接口默认带），需在 iframe 上配置 `allow="autoplay; fullscreen; encrypted-media"` 以允许自动播放与全屏。

**iframe 用法示例:**

```html
<iframe
  :src="url"
  allow="autoplay; fullscreen; encrypted-media"
  allowfullscreen
  scrolling="no"
  frameborder="0"
></iframe>
```

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "url": "https://player.bilibili.com/player.html?bvid=BV1GJ411x7h7&autoplay=1&danmaku=0&high_quality=1"
  }
}
```

---

#### GET /api/v1/public/music/playlists

获取歌曲列表（第三方歌单）。与 `/api/v1/public/playlists` 为同一处理函数。

**请求参数:** 无

**响应 data 字段（数组）:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 歌单 ID |
| title | string | 歌单标题 |
| cover_url | string | 歌单封面 URL |
| platform | string | 平台名称 |
| platform_url | string | 平台链接 URL |
| description | string | 描述 |
| sort_order | int | 排序 |
| enabled | bool | 是否启用（公开接口仅返回 enabled=true） |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, created_at DESC`。

---

### 6.11 摄影器材模块（公开）

#### GET /api/v1/public/equipments

获取摄影器材列表。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数，max=100 |
| keyword | string | 否 | - | 模糊搜索器材名称 |
| brand | string | 否 | - | 按品牌筛选 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 器材 ID |
| name | string | 器材名称 |
| image_url | string | 器材图片 URL（`AfterFind` 自动拼接为绝对 URL） |
| brand | string | 器材品牌 |
| description | string | 器材介绍 |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, created_at DESC`。

---

#### GET /api/v1/public/equipments/:id

获取摄影器材详情。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 器材 ID |

**响应 data:** 同列表项字段结构

---

### 6.12 模块开关配置（公开）

#### GET /api/v1/public/module-config

获取各模块的开关状态（无需认证）。

**请求参数:** 无

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| article_enabled | bool | 文章管理模块是否开启 |
| media_enabled | bool | 媒体管理模块是否开启 |
| music_enabled | bool | 音乐管理模块是否开启 |
| video_enabled | bool | 视频管理模块是否开启 |
| travel_enabled | bool | 旅行管理模块是否开启 |
| portfolio_enabled | bool | 作品集管理模块是否开启 |
| equipment_enabled | bool | 设备管理模块是否开启 |
| updated_at | string | 配置最后更新时间 |

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "article_enabled": true,
    "media_enabled": true,
    "music_enabled": false,
    "video_enabled": true,
    "travel_enabled": true,
    "portfolio_enabled": true,
    "equipment_enabled": true,
    "updated_at": "2026-09-09T12:00:00Z"
  }
}
```

> **注意:** 模块开关仅作用于公开侧（前端展示），管理后台始终可访问模块管理入口，以方便管理员重新开启模块。

---

### 6.13 第三方歌单（公开）

#### GET /api/v1/public/playlists

获取前台展示的第三方歌单列表。与 `/api/v1/public/music/playlists` 为同一处理函数。

**请求参数:** 无

**响应 data 字段（数组）:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 歌单 ID |
| title | string | 歌单标题 |
| cover_url | string | 歌单封面 URL |
| platform | string | 平台名称 |
| platform_url | string | 平台链接 URL |
| description | string | 描述 |
| sort_order | int | 排序 |
| enabled | bool | 是否启用（公开接口仅返回 enabled=true） |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, created_at DESC`。

---

## 七、数据脱敏规则

公开接口返回数据时，必须过滤敏感字段：

| 模块 | 返回字段 | 排除字段 |
|------|----------|----------|
| 博主信息 | nickname, avatar, bio, email, city, blog_title, blog_description, page_background, blog_icon, social_links, tags | password_hash, username, last_login_at |
| 文章 | 全部字段（不含 deleted_at） | - |
| 评论（公开） | id, target_type, target_id, parent_id, nickname, website, content, is_blogger, created_at | ip_address, blogger_id, updated_at, target_title |
| 旅行攻略 | 全部字段（不含 deleted_at） | - |
| 作品集 | 全部字段（不含 deleted_at） | - |
| 视频 | 全部字段（不含 deleted_at） | - |
| 项目经历 | 全部字段（不含 deleted_at） | - |
| 音乐歌曲 | 全部字段（不含 deleted_at） | - |
| 摄影器材 | 全部字段（不含 deleted_at） | - |
| 模块开关配置 | article_enabled, media_enabled, music_enabled, video_enabled, travel_enabled, portfolio_enabled, equipment_enabled, updated_at | - |
| 第三方歌单（公开） | id, title, cover_url, platform, platform_url, description, sort_order, enabled, created_at, updated_at | - |

---

## 八、使用限制

### 文件上传限制

| 限制项 | 值 |
|--------|-----|
| 单文件最大大小 | 100MB |
| 上传字段名 | file |

### 内容长度限制

| 内容 | 最小长度 | 最大长度 |
|------|----------|----------|
| 昵称 | 1 | 50 |
| 文章标题 | 1 | 200 |
| 旅行攻略标题 | 1 | 200 |
| 旅行攻略目的地 | 1 | 200 |
| 作品集名称 | 1 | 255 |
| 作品项标题 | 1 | 255 |
| 视频标题 | 1 | 255 |
| 歌曲标题 | 1 | 500 |
| 歌曲艺术家 | 1 | 255 |
| 分类名称 | 1 | 50 |
| 标签名称 | 1 | 50 |
| 器材名称 | 1 | 255 |
| 器材品牌 | 0 | 255 |
| 评论内容 | 1 | 2000 |

### 状态值约定

| 模块 | 状态值 | 含义 |
|------|--------|------|
| 文章 | 1 | 草稿 |
| 文章 | 2 | 已发布 |
| 文章 | 3 | 已下架 |
| 旅行攻略 | 1 | 草稿 |
| 旅行攻略 | 2 | 已发布 |
| 旅行攻略 | 3 | 已下架 |
| 作品集 | 0 | 未发布 |
| 作品集 | 1 | 已发布 |
| 视频 | 0 | 未发布 |
| 视频 | 1 | 已发布 |
| 项目经历 | 0 | 草稿 |
| 项目经历 | 1 | 已发布 |
| 评论 | 2 | 已通过审核（默认） |

### 分类类型

| 类型值 | 说明 |
|--------|------|
| article | 文章分类 |
| travel | 旅行攻略分类 |
| portfolio | 作品集分类 |
| music | 音乐分类 |

### 安全配置项

以下配置项位于 `config.yaml` 的 `security` 与 `themes` 段：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| security.blacklist_ttl_minutes | 60 | 黑名单条目有效期（分钟） |
| security.log_retention_days | 7 | 访问日志保留天数 |
| themes.max_artifact_mb | 100 | 主题制品下载上限（MB） |
| themes.market_base_url | 见 config.yaml | 官方主题市场地址（出厂默认值） |
