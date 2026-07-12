---
name: api-reference
description: Cus-CMS 全部开放接口完整参考文档，涵盖公开接口（无需认证）与管理接口（JWT 认证），包括接口路径、请求方法、输入输出参数、错误码、请求/响应示例等
---

# Cus-CMS API 接口参考文档

## 目录

- [一、概述](#一概述)
- [二、统一响应格式](#二统一响应格式)
- [三、错误码说明](#三错误码说明)
- [四、认证机制](#四认证机制)
- [五、安全与限流](#五安全与限流)
- [六、健康检查接口](#六健康检查接口)
- [七、公开接口（无需认证）](#七公开接口无需认证)
  - [7.1 系统初始化](#71-系统初始化)
  - [7.2 博主信息](#72-博主信息)
  - [7.3 文章模块（公开）](#73-文章模块公开)
  - [7.4 分类与标签（公开）](#74-分类与标签公开)
  - [7.5 评论模块（公开）](#75-评论模块公开)
  - [7.6 旅行攻略模块（公开）](#76-旅行攻略模块公开)
  - [7.7 摄影作品集模块（公开）](#77-摄影作品集模块公开)
  - [7.8 视频作品模块（公开）](#78-视频作品模块公开)
  - [7.9 音乐播放器模块（公开）](#79-音乐播放器模块公开)
- [八、管理接口（需 JWT 认证）](#八管理接口需-jwt-认证)
  - [8.1 认证模块](#81-认证模块)
  - [8.2 文章管理](#82-文章管理)
  - [8.3 分类管理](#83-分类管理)
  - [8.4 标签管理](#84-标签管理)
  - [8.5 媒体管理](#85-媒体管理)
  - [8.6 摄影作品集管理](#86-摄影作品集管理)
  - [8.7 视频作品管理](#87-视频作品管理)
  - [8.8 旅行攻略管理](#88-旅行攻略管理)
  - [8.9 音乐播放器管理](#89-音乐播放器管理)
  - [8.10 评论管理](#810-评论管理)
  - [8.11 工作台统计](#811-工作台统计)
  - [8.12 安全管理](#812-安全管理)
  - [8.13 存储配置管理](#813-存储配置管理)
  - [8.14 素材迁移管理](#814-素材迁移管理)
  - [8.15 API 文档](#815-api-文档)
  - [8.16 个人资料管理](#816-个人资料管理)
- [九、数据脱敏规则](#九数据脱敏规则)
- [十、使用限制](#十使用限制)

---

## 一、概述

### 技术栈

- **API 前缀**: `/api/v1`
- **公开路由组**: `/api/v1/public/*`（无需认证，应用 IP 黑名单与限流中间件）
- **管理路由组**: `/api/v1/articles`、`/api/v1/travels` 等（需 JWT 认证）
- **全局中间件**: TraceMiddleware、LoggerMiddleware、RecoveryMiddleware、CORSMiddleware

### 接口分类

| 分类 | 路径前缀 | 认证 | 中间件 |
|------|----------|------|--------|
| 健康检查 | `/health`、`/ready` | 无 | 全局中间件 |
| 公开接口 | `/api/v1/public/*` | 无 | IP 黑名单 + 限流 |
| 管理接口 | `/api/v1/*`（非 public） | JWT Bearer Token | AuthMiddleware |

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
| 400001 | 400 | 请求参数错误 | 请求参数校验失败 |
| 401000 | 401 | 未认证，请先登录 | 缺少 Authorization 头 |
| 401001 | 401 | 用户名或密码错误 | 登录失败 |
| 401002 | 401 | Token 已过期 | JWT 过期 |
| 401003 | 401 | Token 无效 | JWT 无效或在黑名单中 |
| 403000 | 403 | 无权限访问该资源 | 权限不足 |
| 403001 | 403 | 系统已初始化，无法重复创建 | 重复初始化 |
| 403002 | 403 | 您已被暂时限制访问，请稍后再试 | IP 被加入黑名单 |
| 404001 | 404 | 资源不存在 | 资源未找到 |
| 500001 | 500 | 系统内部错误 | 服务器内部异常 |
| 429001 | 429 | 操作过于频繁，请稍后再试 | 触发限流 |

---

## 四、认证机制

### JWT Bearer Token 认证

管理接口（除 `/auth/login`、`/auth/refresh`、`/auth/logout` 外）均需在请求头中携带 JWT Token：

```
Authorization: Bearer <access_token>
```

### 认证流程

1. **登录**: `POST /api/v1/auth/login` 获取 `access_token` 和 `refresh_token`
2. **请求**: 在请求头中携带 `Authorization: Bearer <access_token>`
3. **刷新**: `access_token` 过期后，使用 `refresh_token` 调用 `POST /api/v1/auth/refresh` 获取新令牌
4. **登出**: `POST /api/v1/auth/logout` 将当前 Token 加入黑名单

### AuthMiddleware 执行流程

1. 提取 `Authorization` 头，若为空返回 `401000`
2. 解析 Bearer 格式，若格式错误返回 `401003`
3. 解析 JWT Token，过期返回 `401002`，其他错误返回 `401003`
4. 检查 Token JTI 是否在黑名单中，若在黑名单返回 `401003`
5. 将 `user_id`、`username` 注入 Gin Context，放行请求

---

## 五、安全与限流

### IP 黑名单中间件

应用于公开路由组 `/api/v1/public/*`：
- 检查客户端 IP 是否在黑名单中
- 若在黑名单中返回 `403002`（HTTP 403）
- 检查异常时放行（fail open）

### 限流中间件

应用于公开路由组 `/api/v1/public/*`，基于令牌桶算法：

| 请求类型 | 匹配规则 | 限流参数（可配置） |
|----------|----------|-------------------|
| POST 类 | method=POST 或路径含 "comment" | post_max_tokens / post_window_seconds |
| 浏览量递增 | 路径含 "view" | view_max_tokens / view_window_seconds |
| 点赞 | 路径含 "like" | like_max_tokens / like_window_seconds |
| GET 查询 | 其他 GET 请求 | get_max_tokens / get_window_seconds |

- 超出限流返回 `429001`（HTTP 429）
- 连续触发限流达到阈值（blacklist_threshold）后自动将 IP 加入黑名单
- 黑名单有效期：blacklist_ttl_minutes 分钟

---

## 六、健康检查接口

### GET /health

健康检查，返回 200 表示服务正常运行。

**响应示例:**

```json
{
  "status": "ok"
}
```

### GET /ready

就绪检查，返回 200 表示服务已就绪。

**响应示例:**

```json
{
  "status": "ready"
}
```

---

## 七、公开接口（无需认证）

所有公开接口前缀为 `/api/v1/public`，应用 IP 黑名单与限流中间件。

### 7.1 系统初始化

#### GET /api/v1/public/setup/status

获取系统初始化状态。

**请求参数:** 无

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| initialized | bool | 是否已初始化 |

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "initialized": false
  }
}
```

---

#### POST /api/v1/public/setup/init

初始化博主账号（仅系统未初始化时可调用）。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，3-50 字符 |
| password | string | 是 | 密码，最少 6 字符 |
| nickname | string | 否 | 昵称，最多 50 字符 |

**请求示例:**

```json
{
  "username": "admin",
  "password": "123456",
  "nickname": "Cus"
}
```

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| success | bool | 是否成功 |
| message | string | 响应消息 |

**错误码:**

| 错误码 | 说明 |
|--------|------|
| 400001 | 请求参数错误 |
| 403001 | 系统已初始化，无法重复创建 |

---

### 7.2 博主信息

#### GET /api/v1/public/blogger

获取博主公开信息（脱敏）。

**请求参数:** 无

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| nickname | string | 昵称 |
| avatar | string | 头像 URL |
| bio | string | 个人简介 |
| blog_title | string | 博客标题 |
| blog_description | string | 博客描述 |
| page_background | string | 页面背景图 URL |
| blog_icon | string | 博客 icon 图 URL |

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "nickname": "Cus",
    "avatar": "https://example.com/avatar.jpg",
    "bio": "全栈开发者",
    "blog_title": "Cus Blog",
    "blog_description": "记录技术与生活",
    "page_background": "https://example.com/bg.jpg",
    "blog_icon": "https://example.com/icon.png"
  }
}
```

---

### 7.3 文章模块（公开）

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
| cover_image | string | 封面图 URL |
| category_id | string | 分类 ID |
| category_name | string | 分类名称 |
| tag_ids | []string | 标签 ID 列表 |
| tag_names | []string | 标签名称列表 |
| status | int16 | 状态（2=已发布） |
| type | int16 | 类型 |
| view_count | int | 浏览量 |
| comment_count | int | 评论数 |
| is_top | bool | 是否置顶 |
| is_comment | bool | 是否允许评论 |
| published_at | *time | 发布时间 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

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

### 7.4 分类与标签（公开）

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

> **注意:** 公开接口固定返回 type=article 的分类。管理接口 `/api/v1/categories` 支持通过 `type` 查询参数返回不同模块分类。

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

### 7.5 评论模块（公开）

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

### 7.6 旅行攻略模块（公开）

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
| cover_image | string | 封面图 URL |
| status | int16 | 状态 |
| destination | string | 目的地 |
| region | string | 地区 |
| category_id | string | 分类 ID |
| category_name | string | 分类名称 |
| days | int | 天数 |
| best_month | string | 最佳月份 |
| view_count | int | 浏览量 |
| like_count | int | 点赞数 |
| rating | float64 | 评分 |
| review_count | int | 评价数 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

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

### 7.7 摄影作品集模块（公开）

#### GET /api/v1/public/portfolios

获取已发布作品集列表（仅返回 status=1）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| keyword | string | 否 | - | 模糊搜索名称 |
| category_id | string | 否 | - | 按分类筛选 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 作品集 ID |
| name | string | 名称 |
| description | string | 描述 |
| cover_mode | int | 封面模式: 0=首图, 1=独立封面 |
| cover_preset_id | string | 封面预设 ID |
| cover_url | string | 封面图 URL |
| status | int | 状态 |
| sort_order | int | 排序 |
| category_id | string | 分类 ID |
| category_name | string | 分类名称 |
| item_count | int64 | 作品项数量 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

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

### 7.8 视频作品模块（公开）

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
| cover_url | string | 封面 URL |
| description | string | 描述 |
| status | int | 状态 |
| sort_order | int | 排序 |
| platforms | array | 平台链接列表 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

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

### 7.9 音乐播放器模块（公开）

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

获取歌曲音频播放地址（返回 B站 CDN 直链）。

**路径参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| song_id | string | 是 | 歌曲 ID |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 音频播放地址 |

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "url": "https://example.com/audio.mp3"
  }
}
```

---

## 八、管理接口（需 JWT 认证）

所有管理接口均需在请求头中携带 `Authorization: Bearer <access_token>`。

### 8.1 认证模块

#### POST /api/v1/auth/login

用户登录，获取访问令牌。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**请求示例:**

```json
{
  "username": "admin",
  "password": "123456"
}
```

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| access_token | string | 访问令牌 |
| refresh_token | string | 刷新令牌 |
| expires_in | int64 | 过期时间（秒） |

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "expires_in": 3600
  }
}
```

**错误码:**

| 错误码 | 说明 |
|--------|------|
| 400001 | 请求参数错误 |
| 401001 | 用户名或密码错误 |

---

#### POST /api/v1/auth/refresh

刷新访问令牌。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| refresh_token | string | 是 | 刷新令牌 |

**响应 data:** 同登录响应

**错误码:**

| 错误码 | 说明 |
|--------|------|
| 400001 | 请求参数错误 |
| 401003 | Token 无效 |

---

#### POST /api/v1/auth/logout

登出，将当前 Token 加入黑名单。

**请求头:**

| Header | 说明 |
|--------|------|
| Authorization | Bearer \<access_token\> |

**请求体:** 无

**响应 data:** `null`

---

#### PUT /api/v1/auth/password

修改密码（需认证）。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| old_password | string | 是 | 旧密码 |
| new_password | string | 是 | 新密码 |

**响应 data:** `null`

**错误码:**

| 错误码 | 说明 |
|--------|------|
| 400001 | 请求参数错误 |
| 401000 | 未认证 |
| 401001 | 旧密码错误 |

---

### 8.2 文章管理

#### POST /api/v1/articles

创建文章。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 标题，1-200 字符 |
| content | string | 是 | 正文 |
| summary | string | 否 | 摘要 |
| cover_image | string | 否 | 封面图 URL |
| category_id | string | 否 | 分类 ID |
| tag_ids | []string | 否 | 标签 ID 列表 |
| status | int16 | 否 | 状态: 0/1=草稿(默认), 2=已发布, 3=已下架 |
| type | int16 | 否 | 类型 |
| is_top | *bool | 否 | 是否置顶 |
| is_comment | *bool | 否 | 是否允许评论 |
| extra | map | 否 | 扩展字段 |

**响应 data:** 文章详情对象（含 content、extra）

---

#### GET /api/v1/articles

获取文章列表（管理端，可查看所有状态）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| status | int16 | 否 | - | 按状态筛选 |
| category_id | string | 否 | - | 按分类筛选 |
| keyword | string | 否 | - | 模糊搜索标题 |

**响应 data:** 分页文章列表

---

#### GET /api/v1/articles/:id

获取文章详情。

**路径参数:** `id` (string) - 文章 ID

**响应 data:** 文章详情对象（含 content、extra）

---

#### PUT /api/v1/articles/:id

更新文章（支持部分更新，所有字段为指针类型）。

**路径参数:** `id` (string)

**请求体:** 所有字段均为可选（指针类型）：

| 参数 | 类型 | 说明 |
|------|------|------|
| title | *string | 标题 |
| content | *string | 正文 |
| summary | *string | 摘要 |
| cover_image | *string | 封面图 |
| category_id | *string | 分类 ID |
| tag_ids | []string | 标签 ID 列表 |
| status | *int16 | 状态 |
| type | *int16 | 类型 |
| is_top | *bool | 是否置顶 |
| is_comment | *bool | 是否允许评论 |
| extra | map | 扩展字段 |

**响应 data:** 更新后的文章对象

---

#### DELETE /api/v1/articles/:id

删除文章（软删除）。

**路径参数:** `id` (string)

**响应 data:** `null`

---

#### PUT /api/v1/articles/:id/status

更新文章状态。

**路径参数:** `id` (string)

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | int16 | 是 | 状态: 1=草稿, 2=已发布, 3=已下架 |

> 当 status=2（发布）时，系统自动设置 `published_at` 为当前时间。

**响应 data:** `null`

---

### 8.3 分类管理

#### POST /api/v1/categories

创建分类。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 分类名称，1-50 字符 |
| slug | string | 否 | URL 标识 |
| description | string | 否 | 描述 |
| type | string | 否 | 类型: article / travel / portfolio / music |
| sort_order | int | 否 | 排序 |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 分类 ID |
| name | string | 名称 |
| slug | string | URL 标识 |
| description | string | 描述 |
| type | string | 类型 |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |

---

#### GET /api/v1/categories

获取所有分类列表。

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 否 | 分类类型筛选 |

**响应 data:** 分类数组（无分页）

---

#### GET /api/v1/categories/:id

根据 ID 获取分类。

**路径参数:** `id` (string)

**响应 data:** 分类对象

---

#### PUT /api/v1/categories/:id

更新分类（支持部分更新）。

**路径参数:** `id` (string)

**请求体:** 所有字段可选（指针类型）

| 参数 | 类型 | 说明 |
|------|------|------|
| name | *string | 名称 |
| slug | *string | URL 标识 |
| description | *string | 描述 |
| type | *string | 类型 |
| sort_order | *int | 排序 |

**响应 data:** 更新后的分类对象

---

#### DELETE /api/v1/categories/:id

删除分类。

**路径参数:** `id` (string)

**响应 data:** `null`

---

### 8.4 标签管理

#### POST /api/v1/tags

创建标签。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 标签名称，1-50 字符 |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 标签 ID |
| name | string | 名称 |
| created_at | time | 创建时间 |

---

#### GET /api/v1/tags

获取所有标签列表（无分页）。

**响应 data:** 标签数组

---

#### GET /api/v1/tags/:id

根据 ID 获取标签。

**路径参数:** `id` (string)

**响应 data:** 标签对象

---

#### PUT /api/v1/tags/:id

更新标签。

**路径参数:** `id` (string)

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 标签名称，1-50 字符 |

**响应 data:** 更新后的标签对象

---

#### DELETE /api/v1/tags/:id

删除标签。

**路径参数:** `id` (string)

**响应 data:** `null`

---

### 8.5 媒体管理

#### POST /api/v1/media/upload

上传文件。

**请求体:** `multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 上传的文件，最大 100MB |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 媒体 ID |
| filename | string | 文件名 |
| file_type | int16 | 文件类型 |
| mime_type | string | MIME 类型 |
| size | int64 | 文件大小 |
| url | string | 访问 URL |
| thumb_url | string | 缩略图 URL |
| width | *int | 宽度 |
| height | *int | 高度 |
| created_at | time | 创建时间 |

---

#### POST /api/v1/media/upload-with-preset

上传原图并自动生成预设。

**请求体:** `multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 原图文件，最大 100MB |
| name | string | 否 | 预设名称 |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| media | object | 原图媒体对象（同上传响应） |
| preset | object | 预设对象（见下方） |

**preset 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 预设 ID |
| media_id | string | 原图媒体 ID |
| name | string | 预设名称 |
| frame_config | string | 帧配置 |
| display_params | string | 显示参数 |
| output_url | string | 输出图 URL |
| output_storage_path | string | 存储路径 |
| output_size | int64 | 输出文件大小 |
| mime_type | string | MIME 类型 |
| created_at | time | 创建时间 |

---

#### GET /api/v1/media

获取媒体列表。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| file_type | int16 | 否 | - | 按文件类型筛选 |
| keyword | string | 否 | - | 关键词搜索 |

**响应 data:** 分页媒体列表

---

#### DELETE /api/v1/media/:id

删除媒体。

**路径参数:** `id` (string)

**响应 data:** `null`

---

#### GET /api/v1/media/:id/presets

获取原图下的所有预设。

**路径参数:** `id` (string) - 原图媒体 ID

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | array | 预设列表 |

---

#### POST /api/v1/media/preset

创建媒体预设（为已有原图创建预设）。

**请求体:** `multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| media_id | string | 是 | 原图媒体 ID |
| name | string | 是 | 预设名称 |
| frame_config | string | 是 | 帧配置 |
| display_params | string | 是 | 显示参数 |
| file | file | 是 | 预设输出文件，最大 100MB |

**响应 data:** 预设对象（同上）

---

#### DELETE /api/v1/media/preset/:id

删除预设。

**路径参数:** `id` (string) - 预设 ID

**响应 data:** `null`

---

### 8.6 摄影作品集管理

#### POST /api/v1/portfolios

创建作品集。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 名称，1-255 字符 |
| description | string | 否 | 描述 |
| cover_mode | *int | 否 | 封面模式: 0=首图, 1=独立封面 |
| cover_preset_id | *string | 否 | 封面预设 ID（cover_mode=1 时必填） |
| status | *int | 否 | 状态: 0=未发布, 1=已发布 |
| sort_order | *int | 否 | 排序 |
| category_id | *string | 否 | 分类 ID |

**响应 data:** 作品集对象

---

#### GET /api/v1/portfolios

获取作品集列表。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| keyword | string | 否 | - | 模糊搜索名称 |
| status | int | 否 | - | 按状态筛选 |
| category_id | string | 否 | - | 按分类筛选 |

**响应 data:** 分页作品集列表

---

#### GET /api/v1/portfolios/:id

获取作品集详情（含作品项列表）。

**路径参数:** `id` (string)

**响应 data:** 作品集详情对象（含 items 数组）

---

#### PUT /api/v1/portfolios/:id

更新作品集（支持部分更新）。

**路径参数:** `id` (string)

**请求体:** 所有字段可选（指针类型）

| 参数 | 类型 | 说明 |
|------|------|------|
| name | *string | 名称 |
| description | *string | 描述 |
| cover_mode | *int | 封面模式 |
| cover_preset_id | *string | 封面预设 ID |
| status | *int | 状态 |
| sort_order | *int | 排序 |
| category_id | *string | 分类 ID |

**响应 data:** 更新后的作品集对象

---

#### DELETE /api/v1/portfolios/:id

删除作品集。

**路径参数:** `id` (string)

**响应 data:** `null`

---

#### POST /api/v1/portfolios/:id/items

添加作品项到作品集。

**路径参数:** `id` (string) - 作品集 ID

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| preset_id | string | 是 | 媒体预设 ID |
| title | string | 是 | 标题，1-255 字符 |
| description | string | 否 | 描述 |

**响应 data:** 作品项对象

---

#### PUT /api/v1/portfolios/:id/items/sort

批量排序作品项。

**路径参数:** `id` (string) - 作品集 ID

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| items | array | 是 | 排序项列表，至少 1 项 |

**items 中每项:**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 作品项 ID |
| sort_order | int | 否 | 排序值 |

**响应 data:** `null`

---

#### PUT /api/v1/portfolios/:id/items/:itemId

更新作品项。

**路径参数:** `id` (作品集 ID), `itemId` (作品项 ID)

**请求体:** 所有字段可选

| 参数 | 类型 | 说明 |
|------|------|------|
| preset_id | *string | 预设 ID |
| title | *string | 标题 |
| description | *string | 描述 |

**响应 data:** 更新后的作品项对象

---

#### DELETE /api/v1/portfolios/:id/items/:itemId

删除作品项。

**路径参数:** `id` (作品集 ID), `itemId` (作品项 ID)

**响应 data:** `null`

---

### 8.7 视频作品管理

#### POST /api/v1/videos

创建视频作品。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 标题，1-255 字符 |
| cover_url | string | 否 | 封面 URL |
| description | string | 否 | 描述 |
| status | *int | 否 | 状态: 0=未发布, 1=已发布 |
| platforms | array | 是 | 平台链接列表，至少 1 项 |

**platforms 中每项:**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台名称 |
| url | string | 是 | 链接 URL |

**响应 data:** 视频作品对象（含 platforms）

---

#### GET /api/v1/videos

获取视频作品列表。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| keyword | string | 否 | - | 模糊搜索标题 |
| status | int | 否 | - | 按状态筛选 |

**响应 data:** 分页视频列表

---

#### POST /api/v1/videos/parse

解析视频元信息（从平台链接获取标题、封面、描述）。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台名称 |
| url | string | 是 | 视频 URL |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| title | string | 视频标题 |
| cover_url | string | 封面 URL |
| description | string | 视频描述 |

---

#### GET /api/v1/videos/:id

获取视频作品详情。

**路径参数:** `id` (string)

**响应 data:** 视频作品对象（含 platforms）

---

#### PUT /api/v1/videos/:id

更新视频作品（支持部分更新）。

**路径参数:** `id` (string)

**请求体:** 所有标量字段可选（指针类型），platforms 提供时替换全部链接

| 参数 | 类型 | 说明 |
|------|------|------|
| title | *string | 标题 |
| cover_url | *string | 封面 URL |
| description | *string | 描述 |
| status | *int | 状态 |
| platforms | array | 平台链接列表（提供时替换全部） |

**响应 data:** 更新后的视频对象

---

#### DELETE /api/v1/videos/:id

删除视频作品。

**路径参数:** `id` (string)

**响应 data:** `null`

---

### 8.8 旅行攻略管理

#### POST /api/v1/travels

创建旅行攻略。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 标题，1-200 字符 |
| summary | string | 否 | 摘要 |
| cover_image | string | 否 | 封面图 URL |
| status | int16 | 否 | 状态: 0/1=草稿(默认), 2=已发布, 3=已下架 |
| destination | string | 是 | 目的地，1-200 字符 |
| region | string | 是 | 地区 |
| category_id | *string | 否 | 分类 ID |
| days | int | 否 | 天数 |
| best_month | string | 否 | 最佳月份 |
| attractions | []map | 否 | 景点列表 |
| itinerary | []map | 否 | 行程安排 |
| reviews | []map | 否 | 评价列表 |

**响应 data:** 旅行攻略对象

---

#### GET /api/v1/travels

获取旅行攻略列表（管理端）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| keyword | string | 否 | - | 关键词搜索 |
| region | string | 否 | - | 地区筛选 |
| category_id | string | 否 | - | 分类筛选 |
| status | int16 | 否 | - | 状态筛选 |
| days_range | string | 否 | - | 天数范围: 1-3/4-7/8-14/15+/all |
| sort | string | 否 | - | 排序: views/rating/likes |

**响应 data:** 分页旅行攻略列表

---

#### GET /api/v1/travels/:id

获取旅行攻略详情。

**路径参数:** `id` (string)

**响应 data:** 旅行攻略详情对象（含 attractions、itinerary、reviews）

---

#### PUT /api/v1/travels/:id

更新旅行攻略（支持部分更新）。

**路径参数:** `id` (string)

**请求体:** 所有字段可选（指针类型），同创建参数

**响应 data:** 更新后的旅行攻略对象

---

#### DELETE /api/v1/travels/:id

删除旅行攻略（软删除）。

**路径参数:** `id` (string)

**响应 data:** `null`

---

#### PUT /api/v1/travels/:id/status

更新旅行攻略状态。

**路径参数:** `id` (string)

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | int16 | 是 | 状态: 1=草稿, 2=已发布, 3=已下架 |

**响应 data:** `null`

---

### 8.9 音乐播放器管理

#### POST /api/v1/music/songs

创建歌曲。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 标题，1-500 字符 |
| artist | string | 是 | 艺术家，1-255 字符 |
| cover_url | string | 否 | 封面 URL |
| bvid | string | 是 | B站视频 ID |
| cid | int64 | 是 | B站内容 ID |
| source_url | string | 是 | 来源 URL |
| source_type | string | 否 | 来源类型 |
| category_id | *string | 否 | 分类 ID |
| duration | int | 否 | 时长（秒） |
| sort_order | int | 否 | 排序 |

**响应 data:** 歌曲对象

---

#### POST /api/v1/music/songs/batch

批量创建歌曲。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| songs | array | 是 | 歌曲列表，至少 1 项（每项结构同创建歌曲） |

**响应 data:** 批量创建结果

---

#### GET /api/v1/music/songs

获取歌曲列表。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| category_id | string | 否 | - | 按分类筛选 |

**响应 data:** 分页歌曲列表

---

#### GET /api/v1/music/songs/:id

获取歌曲详情。

**路径参数:** `id` (string)

**响应 data:** 歌曲对象

---

#### PUT /api/v1/music/songs/:id

更新歌曲（支持部分更新）。

**路径参数:** `id` (string)

**请求体:** 所有字段可选（指针类型）

| 参数 | 类型 | 说明 |
|------|------|------|
| title | *string | 标题 |
| artist | *string | 艺术家 |
| cover_url | *string | 封面 URL |
| category_id | *string | 分类 ID |
| duration | *int | 时长 |
| sort_order | *int | 排序 |

**响应 data:** 更新后的歌曲对象

---

#### DELETE /api/v1/music/songs/:id

删除歌曲。

**路径参数:** `id` (string)

**响应 data:** `null`

---

#### POST /api/v1/music/parse

解析 B站音乐链接（异步任务）。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| url | string | 是 | B站链接 |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| task_id | string | 任务 ID |

---

#### GET /api/v1/music/parse/:task_id

获取解析任务状态。

**路径参数:** `task_id` (string)

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| task_id | string | 任务 ID |
| status | string | 状态: pending / processing / success / failed |
| results | array | 解析结果列表（成功时返回） |
| error | string | 错误信息（失败时返回） |

**results 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| title | string | 歌曲标题 |
| artist | string | 艺术家 |
| cover_url | string | 封面 URL |
| duration | int | 时长 |
| bvid | string | B站视频 ID |
| cid | int64 | B站内容 ID |

---

#### GET /api/v1/music/audio-url/:song_id

获取音频播放地址。

**路径参数:** `song_id` (string)

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 音频播放地址 |

---

### 8.10 评论管理

#### GET /api/v1/comments

获取评论列表（管理端，可查看所有状态）。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |
| target_type | string | 否 | - | 目标类型: article / travel_guide |
| target_id | string | 否 | - | 目标 ID |
| keyword | string | 否 | - | 模糊搜索评论内容 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 评论 ID |
| target_type | string | 目标类型 |
| target_id | string | 目标 ID |
| target_title | string | 目标标题 |
| parent_id | *string | 父评论 ID |
| blogger_id | *string | 博主 ID |
| nickname | string | 评论者昵称 |
| website | string | 评论者网站 |
| content | string | 评论内容 |
| is_blogger | bool | 是否博主回复 |
| ip_address | string | IP 地址 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

---

#### POST /api/v1/comments/:id/reply

博主回复评论。

**路径参数:** `id` (string) - 被回复的评论 ID

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | 是 | 回复内容，1-2000 字符 |

**响应 data:** 评论对象（回复记录）

---

#### DELETE /api/v1/comments/:id

删除评论（软删除，级联删除子回复）。

**路径参数:** `id` (string)

**响应 data:** `null`

---

### 8.11 工作台统计

#### GET /api/v1/analytics/overview

获取工作台概览统计数据。

**请求参数:** 无

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| article_total | int64 | 文章总数 |
| article_published | int64 | 已发布文章数 |
| article_draft | int64 | 草稿文章数 |
| portfolio_total | int64 | 作品集总数 |
| portfolio_published | int64 | 已发布作品集数 |
| video_total | int64 | 视频总数 |
| video_published | int64 | 已发布视频数 |
| travel_total | int64 | 旅行攻略总数 |
| travel_published | int64 | 已发布旅行攻略数 |
| song_total | int64 | 歌曲总数 |
| comment_total | int64 | 评论总数 |
| travel_views | int64 | 旅行攻略总浏览量 |
| last_publish_at | *time | 最后发布时间 |
| days_since_last_publish | int | 距上次发布天数 |

---

#### GET /api/v1/analytics/content-trend

获取内容产出趋势。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| range | string | 否 | 30d | 时间范围: 7d / 30d / 90d |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| range | string | 时间范围 |
| items | array | 趋势数据列表 |

**items 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| date | string | 日期 |
| article_count | int64 | 文章数 |
| travel_count | int64 | 旅行攻略数 |

---

#### GET /api/v1/analytics/top-content

获取热门内容排行。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| type | string | 否 | article | 内容类型: article / travel |
| sort | string | 否 | views | 排序方式: views / comments |
| limit | int | 否 | 5 | 返回数量，1-20 |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 内容类型 |
| sort | string | 排序方式 |
| items | array | 排行列表 |

**items 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 内容 ID |
| title | string | 标题 |
| view_count | int64 | 浏览量 |
| comment_count | int64 | 评论数 |
| slug | string | URL 标识（文章） |
| published_at | *time | 发布时间 |

---

#### GET /api/v1/analytics/distribution

获取内容类型分布。

**请求参数:** 无

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| items | array | 分布数据列表 |
| total | int64 | 总数 |

**items 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 内容类型 |
| name | string | 类型名称 |
| count | int64 | 数量 |
| percentage | float64 | 百分比 |

---

#### GET /api/v1/analytics/recent-comments

获取最近评论。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| limit | int | 否 | 5 | 返回数量，1-20 |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| items | array | 最近评论列表 |

**items 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 评论 ID |
| nickname | string | 评论者昵称 |
| content | string | 评论内容 |
| target_type | string | 目标类型 |
| target_id | string | 目标 ID |
| target_title | string | 目标标题 |
| is_blogger | bool | 是否博主 |
| created_at | time | 创建时间 |

---

### 8.12 安全管理

#### GET /api/v1/security/config

获取安全配置。

**请求参数:** 无

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| get_max_tokens | int | GET 请求最大令牌数 |
| get_window_seconds | int | GET 限流窗口（秒） |
| post_max_tokens | int | POST 请求最大令牌数 |
| post_window_seconds | int | POST 限流窗口（秒） |
| view_max_tokens | int | 浏览量递增最大令牌数 |
| view_window_seconds | int | 浏览量限流窗口（秒） |
| like_max_tokens | int | 点赞最大令牌数 |
| like_window_seconds | int | 点赞限流窗口（秒） |
| blacklist_threshold | int | 加入黑名单的违规阈值 |
| blacklist_ttl_minutes | int | 黑名单有效期（分钟） |

---

#### PUT /api/v1/security/config

更新安全配置（支持部分更新）。

**请求体:** 所有字段可选（指针类型），同上方响应字段，每项 min=1。

**响应 data:** `null`

---

#### GET /api/v1/security/blacklist

获取 IP 黑名单列表。

**查询参数:**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页条数 |

**响应 data 字段（分页），list 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 记录 ID |
| ip_address | string | IP 地址 |
| reason | string | 封禁原因 |
| banned_at | time | 封禁时间 |
| expires_at | time | 过期时间 |
| is_active | bool | 是否生效中 |

---

#### DELETE /api/v1/security/blacklist/:ip

解封 IP。

**路径参数:** `ip` (string) - IP 地址

**响应 data:** `null`

---

#### GET /api/v1/security/stats

获取安全统计数据。

**请求参数:** 无

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| blocked_ip_count | int64 | 封禁 IP 数 |
| today_rate_limit_count | int64 | 今日限流次数 |
| daily_trend | array | 每日趋势 |
| top_violations | array | 违规 IP 排行 |

**daily_trend 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| date | string | 日期 |
| count | int64 | 次数 |

**top_violations 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| ip_address | string | IP 地址 |
| count | int64 | 违规次数 |

---

### 8.13 存储配置管理

#### GET /api/v1/storage/config

获取所有存储配置列表。

**请求参数:** 无

**响应 data 字段（数组）:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 配置 ID |
| provider | string | 存储商: aliyun / tencent / minio |
| endpoint | string | 端点地址 |
| region | string | 区域 |
| bucket | string | 存储桶 |
| access_key | string | 访问密钥 |
| access_secret | string | 访问密钥（脱敏返回 "******"） |
| path_prefix | string | 路径前缀 |
| custom_domain | string | 自定义域名 |
| extra | string | 额外配置（JSON 字符串） |
| is_active | bool | 是否激活 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

---

#### POST /api/v1/storage/config

创建或更新存储配置（upsert）。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| provider | string | 是 | 存储商: aliyun / tencent / minio |
| endpoint | string | 是 | 端点地址 |
| region | string | 否 | 区域 |
| bucket | string | 否 | 存储桶 |
| access_key | string | 是 | 访问密钥 |
| access_secret | string | 是 | 访问密钥 |
| path_prefix | string | 否 | 路径前缀 |
| custom_domain | string | 否 | 自定义域名 |
| extra | string | 否 | 额外配置（JSON 字符串，如 {"use_ssl":true}） |

**响应 data:** `null`

---

#### DELETE /api/v1/storage/config/:provider

删除存储配置。

**路径参数:** `provider` (string) - 存储商名称

**响应 data:** `null`

---

#### PUT /api/v1/storage/config/:provider/activate

激活存储配置（设为当前使用的存储商）。

**路径参数:** `provider` (string)

**响应 data:** `null`

---

#### POST /api/v1/storage/config/test

测试存储连通性（不落库）。

**请求体:** 同创建存储配置请求体

**响应 data:** `null`（测试失败返回错误信息）

---

### 8.14 素材迁移管理

#### POST /api/v1/storage/migration/analyze

发起迁移分析任务（异步），分析目标存储中缺失的素材。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| target_provider | string | 是 | 目标存储商: aliyun / tencent / minio |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| task_id | string | 任务 ID |

---

#### GET /api/v1/storage/migration/analyze/:taskId

获取分析结果。

**路径参数:** `taskId` (string)

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| task | object | 任务信息 |
| missing | array | 缺失素材列表 |
| existing | array | 已存在素材列表 |

**task 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 任务 ID |
| task_type | string | 任务类型: analyze |
| target_provider | string | 目标存储商 |
| status | string | 状态: pending / running / completed / failed / canceled |
| total | int | 总数 |
| succeeded | int | 成功数 |
| failed | int | 失败数 |
| started_at | *time | 开始时间 |
| finished_at | *time | 完成时间 |
| error | string | 错误信息 |
| created_at | time | 创建时间 |

**missing/existing 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 记录 ID |
| media_id | string | 媒体 ID |
| status | string | 状态: exist / missing |
| error | string | 错误信息 |

---

#### POST /api/v1/storage/migration/start

发起迁移执行任务（异步）。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| target_provider | string | 是 | 目标存储商 |
| media_ids | []string | 否 | 指定迁移的媒体 ID 列表（与 all 互斥） |
| all | bool | 否 | 一键迁移所有缺失素材 |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| task_id | string | 任务 ID |

---

#### GET /api/v1/storage/migration/status/:taskId

获取迁移进度。

**路径参数:** `taskId` (string)

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| task | object | 任务信息（同分析结果中的 task，task_type=migrate） |
| failed | array | 失败明细列表 |

**failed 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 记录 ID |
| media_id | string | 媒体 ID |
| status | string | 状态: pending / success / failed |
| error | string | 错误信息 |

---

### 8.15 API 文档

#### GET /api/v1/api-docs

获取所有公开 API 的定义文档。

**请求参数:** 无

**响应 data 字段（数组）:**

每项结构：

| 字段 | 类型 | 说明 |
|------|------|------|
| module | string | 模块名称 |
| method | string | HTTP 方法 |
| path | string | 路径（相对于 /api/v1/public） |
| description | string | 接口描述 |
| params | array | 参数列表 |
| response | array | 响应字段列表 |

**params 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 参数名 |
| type | string | 参数类型 |
| required | bool | 是否必填 |
| desc | string | 参数说明 |

**response 中每项:**

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 字段名 |
| type | string | 字段类型 |
| desc | string | 字段说明 |

---

### 8.16 个人资料管理

#### GET /api/v1/profile

获取当前登录博主的完整个人资料。

**请求参数:** 无

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| nickname | string | 昵称 |
| avatar | string | 头像 URL |
| bio | string | 个人简介 |
| page_background | string | 页面背景图 URL |
| blog_icon | string | 博客 icon 图 URL |
| blog_title | string | 博客标题 |
| blog_description | string | 博客描述 |
| email | string | 邮箱 |

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "nickname": "Cus",
    "avatar": "https://example.com/avatar.jpg",
    "bio": "全栈开发者",
    "page_background": "https://example.com/bg.jpg",
    "blog_icon": "https://example.com/icon.png",
    "blog_title": "Cus Blog",
    "blog_description": "记录技术与生活",
    "email": "cus@example.com"
  }
}
```

---

#### PUT /api/v1/profile

更新个人资料（支持部分更新，所有字段均为可选）。

**请求体:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nickname | *string | 否 | 昵称，最多 50 字符 |
| avatar | *string | 否 | 头像 URL，最多 500 字符 |
| bio | *string | 否 | 个人简介 |
| page_background | *string | 否 | 页面背景图 URL，最多 500 字符 |
| blog_icon | *string | 否 | 博客 icon 图 URL，最多 500 字符 |
| blog_title | *string | 否 | 博客标题，最多 100 字符 |
| blog_description | *string | 否 | 博客描述 |

**请求示例:**

```json
{
  "nickname": "Cus",
  "avatar": "https://example.com/new-avatar.jpg",
  "page_background": "https://example.com/new-bg.jpg"
}
```

**响应 data 字段:** 同 GET /api/v1/profile 响应（返回更新后的完整资料）

**错误码:**

| 错误码 | 说明 |
|--------|------|
| 400001 | 请求参数错误 |
| 401000 | 未认证 |

---

#### POST /api/v1/profile/upload-icon

上传博客 Icon 图（需 JWT 认证，上传后返回图片 URL，需通过 PUT /api/v1/profile 保存）。

**请求体:** `multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | Icon 图片文件，最大 10MB，支持 .jpg/.jpeg/.png/.gif/.webp/.svg/.ico |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 上传后的图片 URL |

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "url": "https://example.com/profile/icon/2026/07/uuid.png"
  }
}
```

---

#### POST /api/v1/profile/upload-background

上传页面背景图（需 JWT 认证，上传后返回图片 URL，需通过 PUT /api/v1/profile 保存）。

**请求体:** `multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 背景图片文件，最大 10MB，支持 .jpg/.jpeg/.png/.gif/.webp/.svg/.ico |

**响应 data 字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 上传后的图片 URL |

**响应示例:**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "url": "https://example.com/profile/background/2026/07/uuid.jpg"
  }
}
```

**错误码:**

| 错误码 | 说明 |
|--------|------|
| 400001 | 请求参数错误 |
| 401000 | 未认证 |

---

## 九、数据脱敏规则

公开接口返回数据时，必须过滤敏感字段：

| 模块 | 返回字段 | 排除字段 |
|------|----------|----------|
| 博主信息 | nickname, avatar, bio, blog_title, blog_description, page_background, blog_icon | password_hash, email, username, last_login_at |
| 文章 | 全部字段（不含 deleted_at） | - |
| 评论（公开） | id, target_type, target_id, parent_id, nickname, website, content, is_blogger, created_at | ip_address, blogger_id, updated_at |
| 评论（管理） | 全部字段（含 ip_address, blogger_id, target_title） | - |
| 旅行攻略 | 全部字段（不含 deleted_at） | - |
| 存储配置 | 全部字段 | access_secret 返回 "******" |

---

## 十、使用限制

### 文件上传限制

| 限制项 | 值 |
|--------|-----|
| 单文件最大大小 | 100MB |
| 上传字段名 | file |

### 内容长度限制

| 内容 | 最小长度 | 最大长度 |
|------|----------|----------|
| 用户名 | 3 | 50 |
| 密码 | 6 | - |
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
| 评论内容 | 1 | 2000 |
| 博主回复内容 | 1 | 2000 |

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
| 评论 | 2 | 已通过审核（默认） |

### 分类类型

| 类型值 | 说明 |
|--------|------|
| article | 文章分类 |
| travel | 旅行攻略分类 |
| portfolio | 作品集分类 |
| music | 音乐分类 |

### 限流建议规则

| 接口类型 | 建议限流规则 |
|----------|-------------|
| GET 查询类 | 可配置（默认由安全配置决定） |
| POST 评论 | 可配置（默认由安全配置决定） |
| 浏览量递增 | 可配置（默认由安全配置决定） |
| 点赞 | 可配置（默认由安全配置决定） |
| 违规阈值 | 达到 blacklist_threshold 次后加入黑名单 |
| 黑名单有效期 | blacklist_ttl_minutes 分钟 |
