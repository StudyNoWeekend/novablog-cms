# NovaBlog 主题模板皮肤系统设计文档

> 关联文档：《novablog-内容管理系统设计文档》（本目录）
>
> 本文档定义「主题模板皮肤系统」的完整设计：主题包规范、分发渠道契约、CMS 主题模块、后台管理 UI、安全设计与分期路线图。评审通过后按第九章路线图排期实现。

---

## 一、背景与目标

### 1.1 现状

- 博客前台（访客侧）由多个**独立的 Next.js 静态导出主题**组成（`novablog-web/` 下的 tech-geek、photo-creator、media-creator、lens-life-template），均为 npm 构建、通过 `/api/v1/public/*` 拉取数据，**切换主题目前靠手工部署**。
- CMS 管理后台已有「模板风格」菜单入口（`frontend/src/views/template/TemplateView.vue`），但为空壳页。
- 后端数据库中没有任何主题相关表；官网（`novablog/frontend`）ThemeMarket 板块为硬编码展示。

### 1.2 目标

1. 博主在 CMS 后台**浏览模板市场 → 一键安装 → 一键切换**，访客侧博客页面即时换肤。
2. 建立**开源可扩展的主题包规范**，社区可按规范制作并发布主题。
3. **双分发渠道**：官方 Registry 服务（另起项目承载）为主渠道，GitHub 仓库/Release 直接导入为通用渠道。
4. 为后期多租户（CMS 托管多个用户的博客页面）预留演进路径。

### 1.3 已确认的关键决策

| 决策点 | 结论 |
|--------|------|
| 主题形态 | npm 构建的**静态 dist 制品**（无 Node 常驻进程，切换零重建） |
| 分发渠道 | 官方 Registry API + GitHub 导入，双渠道并存 |
| v1 范围 | 只做模板的浏览/安装/切换；不做主题内部可视化配置（模板自身对接后端公开 API） |
| 在线构建 | v1 不做服务端 npm 构建，制品必须预构建 |

---

## 二、总体架构

### 2.1 三层分离

```
┌─────────────────────────────────────────────────────────────────────┐
│                      主题包规范层（Theme Spec）                       │
│   theme.json 清单 + 静态 dist 制品（tar.gz + sha256）                │
│   主题只依赖「公开数据契约 /api/v1/public/*」，与后端版本解耦          │
└───────────────────────────────┬─────────────────────────────────────┘
                                │ 按规范发布
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          分发层（Distribution）                      │
│  ┌───────────────────────────┐    ┌───────────────────────────────┐ │
│  │ 官方 Registry 服务（主渠道）│    │ GitHub 导入（通用渠道）         │ │
│  │  另起独立项目承载           │    │  Release 制品 / 仓库 tarball   │ │
│  │  上架/版本/统计/审核        │    │  社区主题、官方未上架主题       │ │
│  └───────────────────────────┘    └───────────────────────────────┘ │
└───────────────────────────────┬─────────────────────────────────────┘
                                │ 制品下载（tar.gz + sha256）
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       消费层（CMS 主题模块）                          │
│  • 安装：下载 → 校验 → 解压 → 落库                                   │
│  • 激活：DB 激活指针切换，静态托管即时生效                             │
│  • 托管：Go 内置静态托管器直接对外提供博客页面                         │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 部署拓扑

**推荐拓扑 A：CMS 一体托管（默认）**

```
                        blog.example.com（访客）
                              │
                              ▼
              ┌───────────────────────────────┐
              │        CMS 后端 (Go)          │
              │  /            → 主题静态托管器 │ ← 读取激活主题目录
              │  /preview/:id → 预览托管       │
              │  /api/v1      → 业务 API       │
              │  /files       → 媒体文件       │
              └───────────────────────────────┘
```

访客域名直接解析到 CMS 后端。主题静态页面与 `/api/v1` **天然同域**，无 CORS 问题，主题默认使用相对路径 `/api/v1` 取数。

**备选拓扑 B：Nginx 托管（无 Go 参与访客流量时）**

```
              nginx root → /data/themes/active (软链)
                             active -> tech-geek-1.0.0/
```

激活 = 原子切换软链（`ln -sfn` + `mv -T`）。此拓扑下 API 跨域，需 CMS 在安装时向制品注入 `theme-config.js`（见 3.3）。v1 以拓扑 A 为主，拓扑 B 作为运维备选。

### 2.3 核心切换机制

激活状态只存一个 DB 指针（`bloggers.active_theme_id`）。Go 静态托管器按请求解析激活主题目录（内存缓存，激活时主动失效），因此：

- **切换是秒级的**：无构建、无进程重启、无 Nginx reload；
- **回滚是天然的**：同主题多个版本目录共存，激活旧版本记录即可；
- **预览零成本**：`/preview/{theme_id}/` 路由以指定主题响应同一套托管逻辑。

---

## 三、主题包规范（Theme Spec）

### 3.1 主题的定义

一个主题 = 一份 `theme.json` 清单 + 一个可静态托管的 `dist/` 目录，打包为 `tar.gz` 制品。

```
tech-geek/
├── theme.json              # 主题清单（必需）
├── dist/                   # 构建产物（必需，静态文件）
│   ├── index.html
│   ├── articles/
│   ├── article.html        # 动态内容壳页面（见 3.4）
│   ├── _next/              # Next.js 静态资源
│   └── ...
├── screenshots/            # 截图（必需，市场展示用）
│   ├── home.webp
│   └── article.webp
└── package.json / src/     # 源码（不进入制品）
```

### 3.2 theme.json 清单定义

```json
{
  "id": "tech-geek",
  "name": "极客风",
  "version": "1.0.0",
  "engine": "next-static",
  "api_compat": "v1",
  "author": "novablog",
  "description": "深色极客风格，适合技术类博客",
  "homepage": "https://github.com/novablog-themes/tech-geek",
  "screenshots": ["screenshots/home.webp", "screenshots/article.webp"],
  "preview_url": "https://demo.novablog.example/tech-geek/",
  "routes": {
    "fallback": {
      "/articles/*": "/article.html",
      "/travels/*": "/travel.html",
      "/portfolios/*": "/portfolio.html"
    }
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `id` | string | ✅ | 主题唯一标识，`^[a-z0-9-]{2,50}$`，全局唯一、不可变 |
| `name` | string | ✅ | 展示名 |
| `version` | string | ✅ | 语义化版本 `x.y.z` |
| `engine` | string | ✅ | 渲染引擎，v1 仅 `next-static`；预留 `next-standalone`（SSR）等 |
| `api_compat` | string | ✅ | 兼容的公开 API 版本，当前 `v1`；CMS 版本不兼容时拒绝安装 |
| `author` | string | ✅ | 作者/组织 |
| `description` | string | ✅ | 一句话简介 |
| `homepage` | string | ❌ | 仓库/主页链接 |
| `screenshots` | string[] | ✅ | 制品内相对路径，至少 1 张，市场卡片展示 |
| `preview_url` | string | ❌ | 官方在线 Demo |
| `routes.fallback` | map | ❌ | 动态路由回退壳页面声明（见 3.4），托管器据此实现回退 |

**校验规则**：CMS 安装时按上表校验必填字段与格式，`engine` 不在支持列表、`api_compat` 与当前后端不兼容时直接拒绝。

### 3.3 运行时配置注入机制（关键设计）

静态制品是**中心化构建**的，而每个部署者的 API 地址不同（跨域部署时）。为做到「一份制品到处部署」，引入部署时注入：

CMS 在安装完成前，向制品根目录写入 `theme-config.js`：

```js
// 同域部署（拓扑 A）时 apiBase 为空，主题走相对路径；跨域时为绝对地址
window.__NOVA_CONFIG__ = { apiBase: "https://api.example.com/api/v1" };
```

主题数据层按以下优先级取 API 地址：

```ts
// lib/api/client.ts 约定
const API_BASE =
  (typeof window !== "undefined" && window.__NOVA_CONFIG__?.apiBase) ||
  "/api/v1"; // 同域默认值，且 dist/index.html 引入 <script src="/theme-config.js">
```

- 同域拓扑下 CMS 不写此文件（或写空值），主题回退相对路径 `/api/v1`；
- `theme-config.js` 由部署方管理，**主题更新/重装时自动重写**，用户不可手改。

### 3.4 动态路由与壳页面（fallback）约定

Next.js 静态导出**不支持动态路由 fallback**（`generateStaticParams` 未覆盖的路径构建后即 404）。博客文章持续增长，不可能靠重新构建覆盖，因此规范强制约定：

1. 动态内容路由（文章/攻略/作品详情等）必须提供**壳页面**：一个预渲染的静态 HTML，页面 JS 从 URL 解析 slug 后客户端取数渲染；
2. 主题在 `theme.json.routes.fallback` 中声明「路径模式 → 壳页面」映射；
3. CMS 托管器按「精确文件 → `${path}.html` → 目录 `index.html` → fallback 壳页面 → 404」的顺序解析请求。

> ⚠️ 现有主题 tech-geek 预生成前 100 篇文章详情（`generateStaticParams`），第 101 篇起将 404。Phase 1 示例改造时必须为其补壳页面并声明 fallback。

### 3.5 主题数据契约（必须对接的公开 API）

主题只允许消费 CMS 公开只读接口（免鉴权，前缀 `/api/v1/public`），当前契约：

| 接口 | 用途 |
|------|------|
| `GET /public/blogger` | 站点信息：标题/描述/头像/社交链接/标签 |
| `GET /public/articles`、`/public/articles/:slug`、`/hot`、`/random`、`/view` | 文章列表/详情/热点/随机/浏览计数 |
| `GET /public/categories`、`/public/tags` | 分类、标签 |
| `GET /public/comments`（及发表评论接口） | 评论 |
| `GET /public/travels`、`/portfolios`、`/videos`、`/equipments`、`/music` | 垂直内容 |

- 媒体资源使用返回的完整 URL（`/files/...` 或 OSS 直链），主题不感知存储细节；
- 响应统一 wrapper：`{code, data, msg, trace_id}`，`code=0` 为成功；
- 新增接口只做加法；`api_compat` 版本号在发生**破坏性变更**时递增，CMS 以版本比对决定安装准入。

### 3.6 工程与打包约定

主题源码仓库（如 `novablog-themes` 仓库内的每个子目录）：

```
themes/tech-geek/
├── theme.json          # 源码仓库同样维护清单（build 段声明产物目录）
├── package.json
├── next.config.ts      # output: "export", distDir: "dist"
└── src|app/ ...
```

打包命令约定（CI 或本地执行）：

```bash
npm ci && npm run build
# 产出制品：tech-geek-1.0.0.tar.gz = theme.json + dist/ + screenshots/
tar -czf tech-geek-1.0.0.tar.gz theme.json dist screenshots
sha256sum tech-geek-1.0.0.tar.gz
```

制品内**必须包含** theme.json、dist、screenshots；**不得包含**源码、node_modules、.env。

---

## 四、分发渠道

### 4.1 双渠道总览

| | 官方 Registry（主渠道） | GitHub 导入（通用渠道） |
|---|---|---|
| 承载 | 另起的官网后端项目（Phase 2） | GitHub Releases / 仓库 tarball |
| 后台体验 | 市场卡片流：浏览/搜索/筛选，一键安装 | 输入仓库地址导入 |
| 校验强度 | 强制 sha256 比对 | Release 制品校验；tarball 记录 checksum |
| 适用 | 官方与审核过的社区主题 | 官方未上架、自研私有主题 |
| CMS 配置 | `themes.market_base_url`（现行实现对接 novablog 官方市场代理，Registry 契约保留为 Phase 2 设计） | GitHub 导入当前常开，无独立开关 |

### 4.2 官方 Registry API 契约

契约先行、双方并行开发。CMS 端只依赖以下三个接口，`registry_base_url` 可配置：

```
# 1. 主题列表（市场页数据源）
GET {registry_base}/api/registry/v1/themes
    ?page=1&page_size=20&keyword=&engine=next-static&sort=downloads
→ 200 { "code": 0, "data": { "list": [ ThemeSummary ], "total": 42 } }

ThemeSummary = {
  "id": "tech-geek", "name": "极客风", "latest_version": "1.2.0",
  "author": "novablog", "description": "...",
  "screenshots": ["https://cdn.../home.webp"],   // 绝对 URL
  "downloads": 1024, "created_at": "..."
}

# 2. 主题详情（含全部版本）
GET {registry_base}/api/registry/v1/themes/{theme_id}
→ 200 { "code": 0, "data": { ...ThemeSummary,
        "versions": [ { "version": "1.2.0", "changelog": "...",
                        "checksum": "sha256hex", "download_url": "https://cdn.../tech-geek-1.2.0.tar.gz",
                        "created_at": "..." } ] } }

# 3. 制品下载（可 302 到 CDN）
GET {registry_base}/api/registry/v1/themes/{theme_id}/versions/{version}/download
→ 200 (tar.gz 流) 或 302 Location: CDN
```

**官方 Registry 项目（Phase 2，另起独立项目）** 职责：主题上架/版本管理、制品存储（复用 MinIO/OSS/COS 方案）、下载统计、审核流程、向官网 ThemeMarket 页面供数。v1 上架采用管理员后台上传，作者自助投稿后续开放。

### 4.3 GitHub 导入约定

后台输入 `https://github.com/{owner}/{repo}`（可选 tag/branch），CMS 按以下顺序解析：

1. **Release 制品优先**：最新 Release（或指定 tag）附件中匹配 `{主题目录名}-{语义化版本}.tar.gz`，且解压后根目录含 `theme.json` → 直接导入；
2. **仓库 tarball 兜底**：`codeload.github.com/{owner}/{repo}/tar.gz/{ref}` 下载源码归档，要求根目录（或一级子目录）同时存在 `theme.json` 与 `dist/` → 导入；
3. 两者都不满足 → 报错提示：「未找到预构建制品，请使用官方 Registry，或让主题作者在 Release 附带制品 / 提交 dist 目录」。

> v1 **不做服务端 npm 构建**（避免 CMS 依赖 Node 工具链与构建资源风险）。源码托管构建交给主题作者的 GitHub Actions，规范仓库提供标准 workflow 模板。

### 4.4 过渡期策略

官方 Registry 上线前，`market_base_url` 可临时指向任何兼容该契约的来源（例如 GitHub 仓库内的静态 `registry.json` 索引服务）。后台市场页在 Registry 不可用/未配置时置灰并提示走 GitHub 导入。

---

## 五、CMS 主题模块设计（novablog-admin/backend）

### 5.1 目录结构（遵循现有分层 router → controller → logic → model）

```
backend/internal/
├── router/theme.go              # 路由注册
├── controller/theme_controller.go
├── logic/theme_logic.go         # 安装/激活/卸载/列表编排
├── logic/theme_host.go          # 静态托管器（含 fallback 解析）
├── logic/registry_client.go     # 官方 Registry HTTP 客户端
├── dto/req/theme_req.go
├── dto/res/theme_res.go
├── model/theme.go
└── storage/                     # 复用现有（无需改动）
```

### 5.2 数据模型

新表 `themes`（已安装主题实例，同主题多版本共存）：

```sql
CREATE TABLE themes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    theme_id      VARCHAR(100) NOT NULL,              -- theme.json 的 id
    name          VARCHAR(100) NOT NULL,
    version       VARCHAR(50)  NOT NULL,
    engine        VARCHAR(50)  NOT NULL,              -- next-static
    api_compat    VARCHAR(20)  NOT NULL,              -- v1
    author        VARCHAR(100),
    description   TEXT,
    screenshots   JSONB,                              -- 制品内相对路径数组
    source        VARCHAR(20)  NOT NULL,              -- official | github
    source_ref    VARCHAR(500),                       -- registry url / github owner/repo@ref
    artifact_path VARCHAR(500) NOT NULL,              -- 解压目录（theme_root）
    checksum      VARCHAR(64),                        -- sha256 hex
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT uk_themes_id_version UNIQUE (theme_id, version)
);
```

`bloggers` 表新增字段（激活指针；站点配置本就存放于 bloggers 单行表，多租户演进时 blogger 即站点）：

```sql
ALTER TABLE bloggers ADD COLUMN active_theme_id UUID REFERENCES themes(id);
```

- 激活引用 `themes.id`（具体安装实例），回滚 = 激活另一条记录；
- 卸载被激活的主题时后端拒绝（需先切换到其他主题）；
- 遵循项目现状：模型加入 `bootstrap/db.go` 的 AutoMigrate 列表，并同步补写 `migrations/` SQL 文件。

### 5.3 管理 API 契约

全部挂载在 `/api/v1` 下、走现有 JWT 鉴权中间件：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/themes/registry` | 代理官方 Registry 列表（透传分页/关键词），市场页数据源 |
| GET | `/api/v1/themes/registry/:theme_id` | 代理官方主题详情（含版本列表、changelog） |
| GET | `/api/v1/themes` | 已安装主题列表（含各主题最新 registry 版本号，供「检查更新」） |
| POST | `/api/v1/themes/install` | 安装：`{ "source": "official"\|"github", "theme_id"?, "version"?, "repo"?, "ref"? }` |
| POST | `/api/v1/themes/:id/activate` | 激活指定安装实例（UUID） |
| DELETE | `/api/v1/themes/:id` | 卸载（激活中的主题拒绝卸载） |

响应遵循现有 wrapper `{code, data, msg, trace_id}`；错误复用 `enum.BizError` 语义（如 4001 清单校验失败、4002 引擎不兼容、4003 校验和不匹配、4004 未找到预构建制品）。

### 5.4 安装流程

```
后台点击安装
   │
   ▼
① 解析来源 ── official: registry 拉取制品 URL ── github: Release/tarball 解析（4.3）
   │
   ▼
② 流式下载到临时文件，边下边算 sha256
   │      official 渠道：与 Registry 声明 checksum 比对，不一致即失败
   │      github 渠道：无法比对上游宣称值，checksum 仅记录入库用于完整性跟踪
   │
   ▼
③ 安全解压到临时目录（tar-slip 防护 + 配额限制）
   │
   ▼
④ 校验 theme.json：必填字段 / engine 支持 / api_compat 兼容 / dist 与截图存在
   │
   ▼
⑤ 注入 theme-config.js（跨域场景，见 3.3）
   │
   ▼
⑥ 原子移动到 {themes.data_dir}/{theme_id}-{version}/
   │      目标已存在（同 id 同版本）→ 返回冲突，由用户选择覆盖更新
   ▼
⑦ 落库 themes 表，清理临时文件
```

解压安全限制：拒绝 `..`/绝对路径/符号链接条目；总解压大小 ≤ 200MB、文件数 ≤ 5000、强制 UTF-8 文件名。

### 5.5 激活与静态托管

**激活（秒级）**：校验实例存在且兼容 → 事务更新 `bloggers.active_theme_id` → 失效托管器内存缓存 → 立即生效。

**静态托管器（拓扑 A，Go 内置）**：注册一条兜底路由（优先级低于 `/api`、`/files`、`/preview`），解析顺序：

```
① 精确文件存在？        →  直接返回（设置长缓存于 _next/ 静态资源）
② {path}.html 存在？    →  返回
③ {path}/index.html？  →  返回
④ 命中激活主题 fallback 壳页面（3.4）？ → 返回壳页面
⑤ / 404.html 或内置 404
```

防护：基于 `http.Dir` 的目录穿越防护；拒绝 `.env`、`theme.json` 等 dotfile/清单文件直出；`Content-Type` 按扩展名正确设置。

**预览路由**：`GET /preview/{theme_id}/...` 与主托管逻辑相同，仅把「激活主题」替换为 URL 指定主题（要求该主题已安装）。后台「预览」按钮新窗口打开此地址，激活前即可看真实效果。

---

## 六、后台管理 UI 设计（frontend）

填充现有空壳页 `frontend/src/views/template/TemplateView.vue`（路由 `templates` 已存在，无需改路由），新增 `frontend/api/theme.ts` 对接后端。

### 6.1 页面结构

```
模板风格（/templates）
├── Tabs「模板市场」
│    ├── 顶部：搜索框 + 引擎筛选 + 「从 GitHub 导入」按钮
│    ├── 卡片流：截图 / 名称 / 版本 / 作者 / 简介标签 / 下载量
│    │    └── 卡片操作：安装（进度反馈）· 详情抽屉（多版本 + changelog + 截图大图）
│    └── 空态：Registry 未配置/不可达 → 引导走 GitHub 导入
└── Tabs「已安装模板」
     ├── 卡片：截图 / 名称 / 版本 / 来源 / 安装时间
     │    ├── 「使用中」徽标（当前激活）
     │    └── 操作：启用 · 预览（新窗口 /preview/{id}/）· 检查更新 · 卸载
     └── 更新提示：已安装版本 < registry 最新版本时卡片角标提醒
```

### 6.2 交互要点

- **安装**为长任务：按钮进入 loading 并分阶段提示（下载中/校验中/安装完成）；失败展示具体错误码文案（见 5.3）；
- **激活**二次确认（「切换后访客侧立即生效」）；成功后全局 message 提示 + 「使用中」徽标迁移；
- **卸载**二次确认；激活中主题的卸载按钮禁用并 tooltip 说明；
- 遵循现有 Ant Design Vue 4 组件与 UnoCSS 原子类风格，图标用 SkinOutlined 体系。

---

## 七、安全设计

| 风险 | 措施 |
|------|------|
| 制品被篡改 | 官方渠道强制 sha256 比对 Registry 声明值；GitHub Release 制品校验、tarball 记录 checksum；全程 HTTPS，可选下载域白名单配置 |
| 路径穿越（tar-slip） | 解压期拒绝 `..`、绝对路径、符号链接；大小/文件数配额 |
| 恶意清单 | theme.json 按 schema 强校验；engine 白名单；api_compat 准入 |
| 静态托管逃逸 | 制品为纯静态文件，无代码执行面；托管器禁 dotfile/清单直出；`http.Dir` 防穿越 |
| 越权操作 | 主题管理 API 全部在 JWT 鉴权之后；Registry 代理仅服务后台 |
| SSRF | GitHub 导入仅接受 `github.com` 域；Registry 仅请求配置的 `market_base_url` |

---

## 八、配置项（backend/config/config.yaml 新增）

```yaml
themes:
  data_dir: ./data/themes          # 主题制品解压根目录
  market_base_url: ""              # 官方主题市场地址（服务端直连，用于首装拉默认主题；空 = 首装向导必须手填）
  github_token: ""                 # 可选 GitHub PAT（提升 Release 查询限流额度）
  public_api_base: ""              # 跨域部署时注入 theme-config.js 的 apiBase；空 = 同域相对路径
  max_artifact_mb: 100             # 制品下载上限
```

---

## 九、分期路线图

### Phase 1 —— CMS 全链路 MVP（本设计的主实现期）

1. 后端：themes 表 + `active_theme_id`、theme_logic / registry_client / theme_host、管理 API；
2. 前端：模板市场 / 已安装模板两个 Tab 填充 TemplateView.vue；
3. 主题示例：将 `novablog-web/tech-geek` 改造为首个标准主题——补 `theme.json`、动态路由壳页面 + fallback 声明、`theme-config.js` 读取逻辑、打包脚本与发布 workflow 模板；
4. 验收：后台从 Registry（或 GitHub）安装 tech-geek 制品 → 激活 → 博客域名即时呈现新皮肤 → 预览另一主题并切回。

### Phase 2 —— 官方 Registry 项目（另起独立项目）

1. Registry 后端：上架/版本管理、制品对象存储、下载统计、管理员审核；
2. 官网 ThemeMarket 页面接 Registry 真实数据；
3. CMS 后台配置 `registry_base_url` 指向正式服务。

### Phase 3 —— 多租户演进（远期）

1. blogger = 站点：主题激活指针天然随 blogger 走，themes 表升级为「主题实例 (blogger, theme_id, version)」；
2. 托管器按 Host 头路由到对应站点的激活主题；
3. 模板作者生态：自助上架、分成（依赖 Phase 2 的 Registry 能力）。

---

## 十、开放问题（不阻塞 Phase 1）

1. **静态导出 SEO**：动态内容走壳页面客户端渲染，SEO 弱于 SSR。后续可为规范增加 `engine: next-standalone`（预构建 standalone 产物 + Node 运行时托管），托管器按 engine 分派；
2. **主题更新策略**：v1 由用户手动「检查更新 → 安装新版本 → 激活」；自动更新（安全窗口/灰度）待 Registry 统计数据成熟后评估；
3. **评论等交互组件**：主题内评论 UI 直接调用现有公开接口，若后续开放第三方评论系统（giscus 等）作为主题可选项，再扩展规范；
4. **官方 Registry 域名与 CDN 选型**：Phase 2 立项时确定。
