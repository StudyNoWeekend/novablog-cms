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
  - [6.12 开源作品模块（公开）](#612-开源作品模块公开)
  - [6.13 美食菜谱模块（公开）](#613-美食菜谱模块公开)
  - [6.14 读书书架模块（公开）](#614-读书书架模块公开)
  - [6.15 游戏库模块（公开）](#615-游戏库模块公开)
  - [6.16 健身训练模块（公开）](#616-健身训练模块公开)
  - [6.17 技术栈模块（公开）](#617-技术栈模块公开)
  - [6.18 模块开关配置（公开）](#618-模块开关配置公开)
  - [6.19 第三方歌单（公开）](#619-第三方歌单公开)
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

所有分页列表接口均支持以下分页参数（例外：第三方歌单 `/api/v1/public/playlists` 与 `/api/v1/public/music/playlists` 为非分页接口，分页参数不生效，见 6.19）：

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

### 字段类型约定

- **time**：RFC3339 格式的 JSON 字符串（例：`2026-09-09T12:00:00Z`，由后端 `time.Time` 序列化而来）。文中个别以 `string` 标注的时间字段（如 6.18 的 `updated_at`）格式相同。
- **URL 类字段**（各模块 `cover`/`cover_url`/`icon`/`image_url` 等，模型带 `AfterFind` 钩子）：后端返回前将相对路径自动拼接为 `{upload.base_url}/files/{相对路径}`（`upload.base_url` 为后端配置的访问基础 URL）；以 `http://`/`https://` 开头的值原样返回；未配置 `upload.base_url` 时相对路径原样返回。未带钩子的字段（如菜谱详情的 `steps[].image`）不做拼接、原样返回相对路径，接入者可按同一规则自行拼接。

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

**请求参数:** 无（非分页接口；传入 `page`/`page_size` 会被忽略，响应为裸数组而非分页结构）

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

### 6.12 开源作品模块（公开）

#### GET /api/v1/public/open-sources

获取已发布开源作品列表（仅返回 status=1）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数（最大 100） |
| keyword | string | 否 | - | 模糊搜索仓库名称/一句话介绍 |
| status | int | 否 | - | 公开接口忽略该参数，强制覆盖为 1 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 作品 ID |
| name | string | 仓库名称 |
| repo_url | string | 仓库链接（GitHub 仓库地址） |
| summary | string | 一句话介绍（列表卡片展示） |
| language | string | 主语言 |
| topics | string | 主题标签，逗号分隔 |
| stars | int | Star 数（管理端刷新远端仓库时快照） |
| homepage | string | 主页/演示地址 |
| status | int | 状态（公开接口固定 1=已发布；0=草稿不返回） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, stars DESC, created_at DESC`。
> **status 强制**: logic 层强制覆盖为 1，仅返回已发布开源作品。
> **列表不含 README**：列表查询剔除 `readme` 大字段，需要 README 时调用详情接口。

---

#### GET /api/v1/public/open-sources/:id

获取已发布开源作品详情（含 README 原文）。未发布（status≠1）或已删除的作品视为不存在，返回「开源作品不存在」。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 开源作品 ID |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 作品 ID |
| name | string | 仓库名称 |
| repo_url | string | 仓库链接（GitHub 仓库地址） |
| summary | string | 一句话介绍 |
| readme | string | README 原文（Markdown，管理端创建/刷新时自动从 GitHub 拉取入库） |
| language | string | 主语言 |
| topics | string | 主题标签，逗号分隔 |
| stars | int | Star 数（管理端刷新远端仓库时快照） |
| homepage | string | 主页/演示地址 |
| status | int | 状态（公开接口固定 1=已发布；0=草稿不返回） |
| sort_order | int | 排序 |
| readme_updated_at | time | README 最近拉取时间（null 表示尚未拉取） |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

---

### 6.13 美食菜谱模块（公开）

#### GET /api/v1/public/recipes

获取已发布菜谱列表（仅返回 status=1）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数（最大 100） |
| keyword | string | 否 | - | 模糊搜索标题/摘要/标签 |
| difficulty | int | 否 | - | 按难度筛选（1=简单, 2=中等, 3=困难） |
| status | int | 否 | - | 公开接口忽略该参数，强制覆盖为 1 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 菜谱 ID |
| title | string | 菜谱名称 |
| cover | string | 封面 URL（`AfterFind` 自动拼接为绝对 URL） |
| summary | string | 一句话简介 |
| difficulty | int | 难度（1=简单, 2=中等, 3=困难） |
| minutes | int | 总耗时（分钟） |
| servings | int | 份量 |
| tags | string | 标签，逗号分隔 |
| status | int | 状态（公开接口固定 1=已发布；0=草稿不返回） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, created_at DESC`。
> **status 强制**: logic 层强制覆盖为 1，传入的 `status` 参数不生效，仅返回已发布菜谱。
> **列表裁剪**: 列表项不含 `ingredients`/`steps` 大字段，仅详情接口返回。

---

#### GET /api/v1/public/recipes/:id

获取已发布菜谱详情（未发布或不存在均视为不存在）。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 菜谱 ID |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 菜谱 ID |
| title | string | 菜谱名称 |
| cover | string | 封面 URL（`AfterFind` 自动拼接为绝对 URL） |
| summary | string | 一句话简介 |
| ingredients | array | 食材清单，`[{name, amount}]` |
| steps | array | 步骤列表，`[{text, image}]`（`image` 为相对路径，不自动拼接绝对 URL） |
| difficulty | int | 难度（1=简单, 2=中等, 3=困难） |
| minutes | int | 总耗时（分钟） |
| servings | int | 份量 |
| tags | string | 标签，逗号分隔 |
| status | int | 状态（1=已发布） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **相对路径说明**：`steps[].image` 原样返回相对路径，不做自动拼接。本系统相对路径资源的统一拼接规则为 `{upload.base_url}/files/{相对路径}`（`cover` 等字段由后端 `AfterFind` 钩子自动完成，见「字段类型约定」）；渲染 `steps[].image` 时可按同一规则自行拼接，或直接存储绝对 URL。

---

### 6.14 读书书架模块（公开）

#### GET /api/v1/public/books

获取已发布书籍列表（仅返回 status=1）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码（min=1） |
| page_size | int | 否 | 20 | 每页条数（min=1，max=100） |
| keyword | string | 否 | - | 模糊搜索书名/作者（title ILIKE / author ILIKE） |
| reading_status | string | 否 | - | 按阅读状态筛选（want=想读 / reading=在读 / done=读完），留空或传 `all` 表示不过滤 |
| status | int | 否 | - | 公开接口忽略该参数，强制覆盖为 1 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 书籍 ID |
| title | string | 书名 |
| author | string | 作者 |
| cover | string | 封面 URL（`AfterFind` 自动拼接为绝对 URL） |
| rating | int | 评分（0-5 星，0=未评分） |
| reading_status | string | 阅读状态（want=想读，reading=在读，done=读完） |
| status | int | 状态（公开接口固定 1=已发布；0=草稿不返回） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, updated_at DESC`。
> **status 强制**：logic 层强制覆盖为 1，仅返回已发布书籍。
> 列表为卡片响应，不含书评大字段 `review`（完整字段见下方详情接口）。

---

#### GET /api/v1/public/books/:id

获取已发布书籍详情（含书评）。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 书籍 ID |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 书籍 ID |
| title | string | 书名 |
| author | string | 作者 |
| cover | string | 封面 URL（`AfterFind` 自动拼接为绝对 URL） |
| rating | int | 评分（0-5 星，0=未评分） |
| reading_status | string | 阅读状态（want=想读，reading=在读，done=读完） |
| review | string | 书评（Markdown） |
| started_at | time | 开始阅读时间（可为 null） |
| finished_at | time | 读完时间（可为 null） |
| status | int | 状态（1=已发布） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **status 强制**：详情不区分状态查询，但 logic 层校验 `status != 1` 时按「书籍不存在」返回，草稿对外不可见。

---

### 6.15 游戏库模块（公开）

#### GET /api/v1/public/games

获取已发布游戏列表（仅返回 status=1）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数（最大 100） |
| keyword | string | 否 | - | 模糊搜索标题 |
| platform | string | 否 | - | 按平台模糊筛选（PC/PS5/Switch/Xbox/移动端等） |
| genre | string | 否 | - | 按游戏类型模糊筛选（RPG/FPS/独立游戏等） |
| play_status | string | 否 | - | 按游玩状态精确筛选：want=想玩 / playing=在玩 / played=玩过（传 all 时不过滤） |
| status | int | 否 | - | 公开接口忽略该参数，强制覆盖为 1 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 游戏 ID |
| title | string | 游戏名称 |
| cover | string | 封面 URL（`AfterFind` 自动拼接为绝对 URL） |
| platform | string | 平台（PC/PS5/Switch/Xbox/移动端等） |
| genre | string | 游戏类型（RPG/FPS/独立游戏等） |
| play_status | string | 游玩状态（want=想玩，playing=在玩，played=玩过） |
| play_hours | int | 累计游玩时长（小时） |
| rating | int | 评分（0-10 分，0=未评分） |
| status | int | 状态（公开接口固定 1=已发布；0=草稿不返回） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, updated_at DESC`。
> **status 强制**: logic 层强制覆盖为 1，仅返回已发布游戏。
> 列表为卡片视图，不含 short_review 短评字段（仅详情返回）。

---

#### GET /api/v1/public/games/:id

获取已发布游戏详情（未发布的游戏视为不存在）。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 游戏 ID |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 游戏 ID |
| title | string | 游戏名称 |
| cover | string | 封面 URL（`AfterFind` 自动拼接为绝对 URL） |
| platform | string | 平台（PC/PS5/Switch/Xbox/移动端等） |
| genre | string | 游戏类型（RPG/FPS/独立游戏等） |
| play_status | string | 游玩状态（want=想玩，playing=在玩，played=玩过） |
| play_hours | int | 累计游玩时长（小时） |
| rating | int | 评分（0-10 分，0=未评分） |
| short_review | string | 短评 |
| status | int | 状态（1=已发布） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

---

### 6.16 健身训练模块（公开）

#### GET /api/v1/public/fitness

获取已发布训练记录列表（仅返回 status=1）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数（最大 100） |
| keyword | string | 否 | - | 模糊搜索标题 |
| type | string | 否 | - | 训练类型筛选：strength=力量 / cardio=有氧 / stretch=拉伸；传 `all` 或留空查询全部 |
| status | int | 否 | - | 状态筛选；公开接口强制覆盖为 1，传入其他值不生效 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 训练记录 ID |
| date | time | 训练日期 |
| title | string | 训练标题 |
| type | string | 训练类型（strength=力量 / cardio=有氧 / stretch=拉伸） |
| duration_min | int | 训练时长（分钟） |
| calories | int | 消耗热量（千卡） |
| status | int | 状态（公开接口固定 1=已发布；0=草稿不返回） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`date DESC, created_at DESC`（按训练日期倒序，同日再按创建时间倒序）。
> **status 强制**: logic 层强制覆盖为 1，仅返回已发布训练记录。
> **字段裁剪**: 列表项为卡片结构，不含 `content`（动作清单 `[{name,sets,reps,note}]`）与 `notes`（备注）大字段；本模块未提供公开详情路由。模型无 `AfterFind` 钩子，所有字段原样返回，不涉及 URL 自动拼接。

---

### 6.17 技术栈模块（公开）

#### GET /api/v1/public/tech-stacks

获取已发布技术栈条目列表（仅返回 status=1）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数，max=100 |
| keyword | string | 否 | - | 模糊搜索名称 |
| category | string | 否 | - | 按分类精确筛选（全等匹配，非模糊），如 language/framework/tool/database；分类为管理端维护的自由文本，无固定枚举 |
| level | int | 否 | - | 按熟练度筛选（1=了解, 2=熟悉, 3=熟练, 4=精通） |
| status | int | 否 | - | 公开接口忽略该参数，强制覆盖为 1 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 技术栈条目 ID |
| name | string | 技术名称 |
| category | string | 分类（language/framework/tool/database 等） |
| icon | string | 图标 URL（`AfterFind` 自动拼接为绝对 URL，相对路径追加 `/files/` 前缀） |
| level | int | 熟练度（1=了解, 2=熟悉, 3=熟练, 4=精通） |
| description | string | 描述 |
| status | int | 状态（公开接口固定 1=已发布；0=草稿不返回） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

> **排序规则**：`sort_order ASC, created_at DESC`。
> **status 强制**: logic 层强制覆盖为 1（查询参数中的 status 不生效），仅返回已发布技术栈条目。

---

### 6.18 模块开关配置（公开）

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
| project_enabled | bool | 项目经历管理模块是否开启 |
| open_source_enabled | bool | 开源作品模块是否开启 |
| recipe_enabled | bool | 美食菜谱模块是否开启 |
| book_enabled | bool | 读书书架模块是否开启 |
| game_enabled | bool | 游戏库模块是否开启 |
| fitness_enabled | bool | 健身训练模块是否开启 |
| tech_stack_enabled | bool | 技术栈模块是否开启 |
| updated_at | string | 配置最后更新时间（RFC3339 格式字符串，与全文 `time` 类型字段格式一致） |

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
    "project_enabled": true,
    "open_source_enabled": true,
    "recipe_enabled": true,
    "book_enabled": true,
    "game_enabled": true,
    "fitness_enabled": true,
    "tech_stack_enabled": true,
    "updated_at": "2026-09-09T12:00:00Z"
  }
}
```

> **开关的作用方式**：模块开关**不拦截任何公开接口**。本组 `/api/v1/public/*` 接口始终可用并正常返回数据（即使对应开关为 `false`，也不会返回错误或空列表）；开关仅供前端做展示决策——前端先调用本接口读取各开关，再自行决定是否展示对应模块入口、是否调用对应接口。管理后台始终可访问模块管理入口，以方便管理员重新开启模块。
> **默认值**：配置尚未初始化时，后端自动创建默认配置，全部开关均为 `true`。
> **第三方歌单**（6.19）无独立开关，公开接口层也不受 `music_enabled` 等任何开关拦截；`music_enabled` 仅影响前端相关入口的展示。

---

### 6.19 第三方歌单（公开）

#### GET /api/v1/public/playlists

获取前台展示的第三方歌单列表。与 `/api/v1/public/music/playlists` 为同一处理函数。

**请求参数:** 无（非分页接口；传入 `page`/`page_size` 会被忽略，响应为裸数组而非分页结构）

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
| 开源作品 | 列表接口返回分页结构（list/total/page/page_size/total_pages），list 项含 id、name、repo_url、summary、language、topics、stars、homepage、status、sort_order、created_at、updated_at；详情接口在列表字段基础上增加 readme、readme_updated_at；均不含 deleted_at | 列表接口不返回 readme 与 readme_updated_at（仅详情接口返回） |
| 美食菜谱 | 列表项含 id、title、cover、summary、difficulty、minutes、servings、tags、status、sort_order、created_at、updated_at；详情接口在列表字段基础上增加 ingredients、steps；均不含 deleted_at | 列表接口不返回 ingredients/steps（仅详情接口返回） |
| 读书书架 | 列表项含 id、title、author、cover、rating、reading_status、status、sort_order、created_at、updated_at；详情接口在列表字段基础上增加 review、started_at、finished_at；均不含 deleted_at | 列表接口不返回 review/started_at/finished_at（仅详情接口返回） |
| 游戏库 | 列表项含 id、title、cover、platform、genre、play_status、play_hours、rating、status、sort_order、created_at、updated_at；详情接口在列表字段基础上增加 short_review；均不含 deleted_at | 列表接口不返回 short_review（仅详情接口返回） |
| 健身训练 | 列表接口返回分页结构（list/total/page/page_size/total_pages），list 项含 id、date、title、type、duration_min、calories、status、sort_order、created_at、updated_at；本模块无公开详情接口；不含 deleted_at | deleted_at；content（动作清单）与 notes（备注）不在公开接口返回范围内（无公开详情接口） |
| 技术栈 | 列表接口返回分页结构（list/total/page/page_size/total_pages），list 项含 id、name、category、icon、level、description、status、sort_order、created_at、updated_at；不含 deleted_at | - |
| 模块开关配置 | article_enabled, media_enabled, music_enabled, video_enabled, travel_enabled, portfolio_enabled, equipment_enabled, project_enabled, open_source_enabled, recipe_enabled, book_enabled, game_enabled, fitness_enabled, tech_stack_enabled, updated_at | - |
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
| 评论 | 2 | 已通过审核（评论创建时即写为该值，默认免审核；当前未定义其他状态值，公开列表仅返回该状态） |
| 开源作品 | 0 | 草稿 |
| 开源作品 | 1 | 已发布 |
| 美食菜谱 | 0 | 草稿 |
| 美食菜谱 | 1 | 已发布 |
| 读书书架 | 0 | 草稿 |
| 读书书架 | 1 | 已发布 |
| 游戏库 | 0 | 草稿 |
| 游戏库 | 1 | 已发布 |
| 健身训练 | 0 | 草稿 |
| 健身训练 | 1 | 已发布 |
| 技术栈 | 0 | 草稿 |
| 技术栈 | 1 | 已发布 |

> 上列采用「0=草稿 / 1=已发布」两态的模块（作品集、视频、项目经历、开源作品、美食菜谱、读书书架、游戏库、健身训练、技术栈），其公开列表/详情接口均强制只返回 status=1 的记录，草稿不对外可见。

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
