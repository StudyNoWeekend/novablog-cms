# NovaBlog 主题包规范与发布指南

> 关联文档：《novablog-主题模板皮肤系统设计文档》（总体设计，本仓库）、《deploy/README.md》（容器部署，本仓库）。
>
> **权威版本随主题仓库（novablog-web）维护：`docs/theme-spec.md`**。本文为 CMS 仓库侧的同步副本，面向**主题作者**（如何制作可安装的主题包并发布）与**部署者/运维**（CMS 如何消费、配置与排障）；两处不一致时以主题仓库为准。所有规则与 CMS 后端实现一一对应，评审通过后作为主题上架的唯一准入依据。
>
> 当前已上架主题与状态见主题仓库根 `README.md`（novablog-web）。

**目录**：[一、主题包定义](#一主题包定义) · [二、theme.json 清单规范](#二themejson-清单规范) · [三、主题工程要求](#三主题工程要求nextjs-静态导出) · [四、打包规范](#四打包规范) · [五、发布渠道](#五发布渠道github-release) · [六、CMS 安装链路](#六cms-安装链路部署者参考) · [七、CMS 管理接口速查](#七cms-管理接口速查) · [八、错误码与排障对照](#八错误码与排障对照) · [九、新主题改造清单](#九新主题改造清单checklist) · [十、分期与边界](#十分期与边界)

---

## 一、主题包定义

一个**可安装的主题包** = 一份 `theme.json` 清单 + 一个可静态托管的 `dist/` 目录（可选 `screenshots/`），打包为 `tar.gz` 制品：

```
tech-geek/
├── theme.json              # 主题清单（必需）
├── dist/                   # 构建产物（必需，纯静态文件）
│   ├── index.html          # 入口页（必需）
│   ├── article.html        # 动态内容壳页面（按需）
│   ├── _next/              # Next.js 静态资源
│   └── ...
├── screenshots/            # 截图（可选，市场卡片展示用）
│   └── home.webp
└── package.json / src/     # 源码（不进入制品）
```

**硬性红线**：

- 制品内**必须包含** `theme.json` 与 `dist/index.html`；
- 制品内**禁止出现**：源码目录、`node_modules`、`.env`、符号链接（解压时直接拒绝）；
- 制品总解压大小 ≤ 200MB、文件数 ≤ 5000；
- 制品文件名必须为 `{主题目录名}-{语义化版本}.tar.gz`（见第五章）。

---

## 二、theme.json 清单规范

```json
{
  "id": "tech-geek",
  "name": "极客风",
  "version": "0.1.0",
  "engine": "next-static",
  "api_compat": "v1",
  "author": "novablog",
  "description": "面向开发者与硬核技术读者的深色终端风格博客主题",
  "homepage": "https://github.com/StudyNoWeekend/novablog-web/tree/HEAD/tech-geek",
  "screenshots": [],
  "routes": {
    "fallback": {
      "/articles/*": "/article.html"
    }
  }
}
```

| 字段 | 类型 | 必填 | 校验规则（CMS 安装时强制） |
|------|------|:----:|------|
| `id` | string | ✅ | 匹配 `^[a-z0-9-]{2,50}$`，全局唯一、不可变；**必须等于制品所在的仓库目录名**（安装时比对，不一致拒绝） |
| `name` | string | ✅ | 非空展示名 |
| `version` | string | ✅ | 语义化版本 `x.y.z`（可带 `-pre` 后缀） |
| `engine` | string | ✅ | 白名单校验，**当前仅 `next-static`**（SSR 主题会被拒绝安装） |
| `api_compat` | string | ✅ | 必须为 `v1`（与 CMS 公开 API 版本一致，破坏性变更时递增） |
| `author` | string | ❌ | 作者/组织 |
| `description` | string | ❌ | 一句话简介 |
| `homepage` | string | ❌ | 仓库/主页链接 |
| `screenshots` | string[] | ❌ | 制品内相对路径；为空时后台卡片用渐变占位 |
| `routes.fallback` | map | ❌ | 动态路由壳页面映射，键为路径前缀模式（`*` 结尾），值为 dist 内壳页面路径；托管时**最长前缀匹配** |

> ⚠️ **版本三处一致**：`tag 版本 == theme.json version == package.json version`。官方市场通过 GitHub 同步读取 package.json 版本，CMS 按市场版本匹配 Release 资产名，任一处不一致会导致「制品版本与请求版本不一致」或 CI 直接构建失败（Release CI 强制校验，见 5.3）。

---

## 三、主题工程要求（Next.js 静态导出）

### 3.1 构建配置

`next.config.ts` 必须为纯静态导出：

```ts
const nextConfig: NextConfig = {
  output: "export",   // 必需：静态导出，产物无 Node 常驻进程
  distDir: "dist",    // 约定：产物目录为 dist（打包脚本依赖此约定）
  images: { unoptimized: true },
};
```

### 3.2 API 地址解析（运行时注入优先）

主题数据层**禁止硬编码 API 地址**，按以下优先级取值（参考实现 `tech-geek/lib/api/client.ts`）：

```ts
declare global {
  interface Window { __NOVA_CONFIG__?: { apiBase?: string } }
}

function getBaseUrl(): string {
  // ① 部署端注入的 theme-config.js（允许携带 /api/v1 前缀，需归一化为站点根）
  const injected = typeof window !== "undefined" && window.__NOVA_CONFIG__?.apiBase;
  if (injected) return injected.replace(/\/api\/v1\/?$/, "").replace(/\/$/, "");
  // ② 编译期环境变量（同样归一化 /api/v1 前缀）
  const base = process.env.NEXT_PUBLIC_API_BASE_URL;
  if (base) return base.replace(/\/api\/v1\/?$/, "").replace(/\/$/, "");
  // ③ 同域相对路径（CMS 统一托管时的默认值）
  return "";
}
```

- 同域部署（CMS/nginx 统一入口）下主题走相对路径取数，**一份制品到处部署**；
- ③ 的缺省值必须是空字符串（相对路径），**禁止回退到 `localhost:端口`**——否则 CMS 托管后所有访客浏览器都会去请求自己的 localhost；
- 跨域部署时 CMS 安装阶段自动向 `dist/theme-config.js` 写入 `window.__NOVA_CONFIG__`，主题更新/重装时自动重写，用户不可手改；
- **所有请求必须带超时**（参考实现默认 15s，`AbortSignal.timeout`），杜绝构建期/运行期取数无限挂起。

### 3.3 theme-config.js 引入

根布局必须引入（文件可能不存在，静默容错）：

```tsx
// app/layout.tsx
import Script from "next/script";
// <body> 内：
<Script src="/theme-config.js" strategy="beforeInteractive" />
```

### 3.4 动态内容路由：壳页面与混合渲染（关键）

Next.js 静态导出**不支持动态路由 fallback**——`generateStaticParams` 未覆盖的详情页（如第 101 篇起）构建后即 404。规范强制约定「**预渲染 + 壳页面兜底**」的混合渲染模型：

1. 每类动态内容路由提供**壳页面**：一个自包含的静态 HTML（置于 `public/`，构建后落在 dist 根），页面 JS 从 `location.pathname` 解析 slug 后客户端取数渲染；
2. 在 `theme.json.routes.fallback` 声明「路径模式 → 壳页面」映射（如 `/articles/* → /article.html`、`/travels/* → /travel.html`）；
3. CMS 托管器按「精确文件 → `{path}.html` → `{path}/index.html` → fallback 壳页面（最长前缀）→ `404.html` → 404」顺序解析——未预渲染的详情 URL 由壳页面接住，实时取数渲染。

**壳页面要求**：自包含（内联 CSS/JS，仅允许引用 `/theme-config.js`）；先整体 HTML 转义再渲染（防 XSS）；链接/图片 URL 仅放行 `http(s)://`、`/`、`#`；slug 解析对非法百分号编码容错（try/catch 后显示错误态，禁止卡在加载中）；取数失败展示错误态与返回链接。参考实现：`tech-geek/public/article.html`（深色）、`media-creator/public/article.html`（亮色）、`lens-life-template/public/travel.html`（结构化 JSONB 数据渲染）。

**详情页预渲染规则**：

```ts
export const dynamicParams = false;

// 后端不可达时的构建占位 slug（output: export 不允许 generateStaticParams 返回空数组）；
// 该占位页渲染时不发任何请求，直接 404，保证构建不依赖后端可达
const FALLBACK_SLUG = "__fallback__";

export async function generateStaticParams() {
  try {
    const res = await articles.list({ page: 1, page_size: 100 });
    if (res.list.length === 0) return [{ slug: FALLBACK_SLUG }];
    return res.list.map((a) => ({ slug: a.slug }));
  } catch {
    return [{ slug: FALLBACK_SLUG }];
  }
}
```

- 占位页的页面组件与 `generateMetadata` 遇到占位 slug 必须**短路**：不发任何网络请求，直接 `notFound()` / 返回默认元数据。若占位页仍发起取数，CI 离线构建会挂起至 60 秒超时×3 次后失败；
- 构建时后端可达则预渲染真实详情页（SEO 友好），不可达则全部由壳页面兜底——两种制品均合规。

**列表页（searchParams 筛选/分页）模式**：静态导出禁止在服务端组件读取 `searchParams`。将列表页拆为「静态页壳 + 客户端列表组件」：

- 页壳（服务端组件）：保留页头、纯 HTML GET 搜索表单（`<form action="/articles" method="GET">` 静态可用）、模块开关判断；用 `<Suspense>` 包裹客户端组件并提供骨架屏 fallback；
- 客户端列表组件（`"use client"`）：`useSearchParams` 读筛选/分页参数，`useEffect` 取数渲染分类筛选、卡片网格、分页与侧栏。参考实现：`lens-life-template/components/ArticlesListContent.tsx`。

**服务端副作用迁移**：静态导出后服务端组件只在构建期执行一次，原先的 fire-and-forget 行为全部失效：

- 浏览计数（`POST /articles/{slug}/view`、`/travels/{id}/view`）迁移为客户端上报组件（挂载后 `useEffect` 上报一次，参考 `lens-life-template/components/ViewTracker.tsx`）；
- 点赞、评论发表等本就是客户端组件的行为无需改动；评论拉取改为客户端取数即可。

**静态导出禁用清单**（主题内不得出现）：

| 禁用项 | 原因 | 替代方案 |
|---|---|---|
| `export const dynamic = "force-dynamic"` | 与静态导出互斥 | 删除，数据改为构建期取数 + 客户端取数 |
| 服务端组件读取 `searchParams` | 动态 API，构建期无请求上下文 | §3.4 列表页模式 |
| route handlers（`app/api/**/route.ts`） | 产物无 Node 进程 | 直接调用 CMS 公开 API |
| Server Actions（`"use server"`） | 需服务端运行时 | 客户端 fetch 公开 API |
| `middleware.ts` / `getServerSideProps` / `cookies()` / `headers()` | 同为服务端运行时能力 | — |
| `generateStaticParams` 返回 `[]` | `output: export` 直接报错 | 返回占位 slug 并短路 |

### 3.5 数据契约

主题只允许消费 CMS 公开只读接口（免鉴权，前缀 `/api/v1/public`）：`blogger`、`articles`（列表/详情/hot/random/view）、`categories`、`tags`、`comments`、`travels`、`portfolios`、`videos`、`equipments`、`music`、`module-config`。响应统一 wrapper `{code, data, msg, trace_id}`，`code=0` 为成功。媒体资源使用返回的完整 URL，主题不感知存储细节。

**API 层必须容错**：每个取数函数 catch 异常后回退空数据/`null`（参考 `lens-life-template/lib/api/*`），保证任何一处后端不可达都不会让 `pnpm build` 失败——构建期失败只允许降级（少预渲染几个页面），不允许中断。

### 3.6 工程一致性要求

- **版本三处一致**：tag = theme.json = package.json（Release CI 强制校验）；
- **lockfile 必须与 package.json 同步**：CI 以 `pnpm install --frozen-lockfile` 安装，改依赖后必须重新 `pnpm install` 并提交 `pnpm-lock.yaml`，否则 CI 直接失败；
- **不要声明 `packageManager` 字段**（或必须与 CI 的 pnpm 9 兼容）：CI 固定使用 pnpm 9 + Node 20，字段值不一致会导致安装失败；lockfile 版本须为 `9.0`；
- `devDependencies` 中的 lint 工具版本仅影响本地 lint，不进入制品。

---

## 四、打包规范

日常发布直接走第五章的 Release CI，无需手动打包。手动打包时：

```bash
cd tech-geek
pnpm install && pnpm build
# 产物：tech-geek-{version}.tar.gz = theme.json + dist [+ screenshots]
tar -czf tech-geek-0.1.0.tar.gz theme.json dist
# 有截图时：
tar -czf tech-geek-0.1.0.tar.gz theme.json dist screenshots
sha256sum tech-geek-0.1.0.tar.gz > tech-geek-0.1.0.tar.gz.sha256
```

- tar 根目录必须直接是 `theme.json`、`dist/` 条目（不要多包一层目录）；
- 附加 `.sha256` 校验文件为推荐做法（CMS 官方渠道强制比对 checksum 时使用）；**校验和文件内使用相对文件名**（`shasum -a 256 tech-geek-0.1.0.tar.gz > …` 而非绝对路径），下游可直接 `shasum -c`。

---

## 五、发布渠道：GitHub Release

CMS 安装链路从主题的 `download_url`（形如 `https://github.com/{owner}/{repo}/tree/HEAD/{目录}`）解析出 `owner/repo/目录名`，然后调用 GitHub API 遍历该仓库**最近 30 个 Release**，查找资产名为 `{目录名}-{版本}.tar.gz` 的制品。因此：

### 5.1 资产命名（强制）

```
{主题目录名}-{theme.json版本}.tar.gz
例：tech-geek-0.1.0.tar.gz
```

版本匹配规则：CMS 请求安装时携带市场主题版本（来自 package.json 同步），精确匹配资产名；未指定版本时取**最近的、能按模式匹配上的** Release。

### 5.2 发布 tag 规范（强制）

一次主题发布 = 一个主题级 tag。**tag 只用于触发 CI 发布，CMS 匹配只认资产名（5.1），与 tag 命名无关**：

```
{主题目录名}-v{语义化版本}
例：tech-geek-v0.1.0、tech-geek-v0.2.0-beta.1
```

- 一个 tag 只构建、只发布一个主题，避免多主题共用全局版本号带来的语义混乱与「全家桶」式重复构建；
- 新增主题**无需修改任何 CI 文件**：建目录 + `theme.json` + 打 tag 即可。

### 5.3 方式一：Release CI（已内置，推荐）

主题仓库根已内置 `.github/workflows/release.yml`，无需新建，行为如下：

- **触发**：推送 `*-v*` 主题 tag（正式发布）；或手动 `workflow_dispatch`（输入主题目录名，产物以 Actions Artifact 供测试下载，不建 Release）；
- **流程**：解析 tag（格式不符直接失败）→ 校验主题目录存在且含 `theme.json` → 校验三重版本一致（tag 版本 == `theme.json` version == `package.json` version）→ `pnpm install --frozen-lockfile` → `pnpm build`（未设置 API 地址时走相对路径回退，构建不依赖后端可达）→ 按第四章打包 `{目录名}-{版本}.tar.gz` + `.sha256`（相对文件名）→ 挂载到该 tag 的 Release。

发布操作：

```bash
git add . && git commit -m "主题规范改造"
git tag tech-geek-v0.1.0 && git push origin tech-geek-v0.1.0   # 2~3 分钟后 Release 产出制品
```

### 5.4 方式二：手动发 Release

```bash
gh release create tech-geek-v0.1.0 tech-geek-0.1.0.tar.gz tech-geek-0.1.0.tar.gz.sha256 \
  --title "tech-geek v0.1.0" --notes "tech-geek 主题制品"
```

或 GitHub 网页 → Releases → Draft a new release（tag 命名仍须遵循 5.2）→ 上传资产。

### 5.5 发布前置自检（避免烧 CI 轮次）

打 tag 前在本地做一次**与 CI 完全一致**的构建：

```bash
NEXT_PUBLIC_API_BASE_URL= pnpm build   # 空值覆盖本地 .env.local，保证与 CI（无该变量）环境一致
```

构建必须成功且不挂起（取数失败自动回退占位 slug，与 CI 同路径）；产物 `dist/` 应包含 `index.html`、`404.html` 与全部壳页面，且**客户端 chunk 中不得出现任何具体 API 地址**（可用 `grep -r "localhost\|127.0.0.1" dist/` 复查）。

> ⚠️ **禁止**通过设置 `NEXT_PUBLIC_API_BASE_URL` 为死地址（如 `http://127.0.0.1:59999`）来模拟后端不可达后打包发布——该变量为编译期内联，死地址会被烤进客户端 chunk，访客浏览器将向死地址取数（违反 §3.2）。本地 `.env.local` 存有调试地址时同理：不覆盖直接构建的产物**不得用于发布**（这正是 §8.2「本地构建成功、CI 失败」的常见根因）。

---

## 六、CMS 安装链路（部署者参考）

```
后台点击安装 / 首装向导
   │
   ▼
① 解析来源 ── 官方市场 GET /themes/default 或 /themes/:id（download_url）
   │
   ▼
② githubartifact.ParseRepoDir：从 download_url 提取 owner/repo/目录
   │    （仅接受 github.com 域，防 SSRF）
   ▼
③ GitHub API 遍历最近 30 个 Release，匹配资产 {目录}-{版本}.tar.gz
   │    （未配置 github_token 时受匿名 60 次/小时限流约束）
   ▼
④ 流式下载制品，边下边算 sha256（大小上限 themes.max_artifact_mb，默认 100MB）
   │
   ▼
⑤ 安全解压到 {data_dir}/.tmp/：拒绝 ..、绝对路径、符号链接；≤200MB / ≤5000 文件
   │
   ▼
⑥ 校验 theme.json：必填字段 / engine 白名单 / api_compat / dist/index.html 存在 / id==目录名
   │
   ▼
⑦ 跨域部署时向 dist/ 注入 theme-config.js（apiBase 取 themes.public_api_base）
   │
   ▼
⑧ 原子 mv 到 {data_dir}/{id}-{version}/，落库 themes 表（UNIQUE(theme_id, version)）
   │
   ▼
⑨ 激活：UPDATE bloggers.active_theme_id → 失效托管缓存 → 博客页面秒级生效
```

**静态托管解析顺序**（Go 兜底路由，位于 `/api`、`/files`、`/preview` 之后）：

```
精确文件 → {path}.html → {path}/index.html → fallback 壳页面（最长前缀）→ 404.html → 404
```

防护：拒绝 dotfile 与 `theme.json` 直出；`_next/`、`static/` 下发一年强缓存；`http.Dir` 防目录穿越。

---

## 七、CMS 管理接口速查

| 方法 | 路径 | 认证 | 说明 |
|---|---|---|---|
| GET | `/api/v1/public/install/status` | ❌ | 是否已初始化（bloggers 表为空即未初始化） |
| POST | `/api/v1/public/install/init` | ❌ | 创建博主账号（首装第一步） |
| POST | `/api/v1/public/install/theme` | ❌ | 首装拉默认主题并激活；body 可带 `{"market_base_url":"http://…"}` 覆盖配置 |
| GET | `/api/v1/public/install/theme/status` | ❌ | 轮询安装任务状态（not_started/running/success/failed） |
| GET | `/api/v1/themes` | ✅ | 已安装主题列表（含 active 标记） |
| POST | `/api/v1/themes/install` | ✅ | 安装：`{"theme_id":1,"version":"…","force":false}` |
| POST | `/api/v1/themes/:id/activate` | ✅ | 激活（秒级切换） |
| DELETE | `/api/v1/themes/:id` | ✅ | 卸载（激活中的实例拒绝） |
| GET | `/api/v1/themes/market/default` | ✅ | 代理官方默认主题 |
| GET | `/preview/{theme_id}/` | ❌ | 主题预览（激活前即可看真实效果） |

**关键配置**（config.yaml `themes` 段）：

```yaml
themes:
  data_dir: ./data/themes            # 制品解压根目录
  market_base_url: ""                # 官方市场地址（服务端直连；空=首装向导必须手填）
  github_token: ""                   # 可选 PAT，提升 Release 查询限流额度
  public_api_base: ""                # 跨域注入 apiBase；空=同域相对路径
  max_artifact_mb: 100               # 制品下载上限
```

---

## 八、错误码与排障对照

### 8.1 CMS 错误码

| 报错 | 业务码 | 根因 | 处理 |
|---|---|---|---|
| 未配置官方市场地址（themes.market_base_url） | 400201 | 首装未填地址且配置为空 | 向导输入地址后重试，或写入 config.yaml |
| 主题清单（theme.json）校验失败 | 400202 | 清单缺字段 / id 与目录不一致 / 缺 dist/index.html | 对照第二章修正后重新打包 |
| 主题引擎不受支持 | 400203 | SSR 主题或 engine 拼写错误 | 改造为 `output: "export"` 静态导出 |
| 主题与当前系统 API 版本不兼容 | 400204 | api_compat ≠ v1 | 等待 CMS 升级或主题适配 |
| 制品校验和不匹配 | 400205 | 下载被篡改/截断 | 重新上传制品并重试 |
| **未找到预构建制品** | 400206 | Release 无 `{目录}-{版本}.tar.gz` 资产 | 按第五章发布制品；若市场已显示新版本号但未打 tag，补打 `{目录名}-v{版本}` tag |
| 该主题版本已安装 | 400207 | 同版本重复安装 | 幂等返回已有实例；需覆盖时 `force: true` |
| 该主题正在使用中，请先切换 | 400208 | 卸载/覆盖激活中实例 | 先启用其他主题 |
| GitHub API 访问受限 | — | 匿名限流（60 次/小时） | 配置 `themes.github_token` |
| 访问 `/setup` 返回 404/502 | — | 少输 `/admin/` 前缀 | 已由 nginx 兜底跳转；确认访问 `:8188/admin/setup/` |

### 8.2 发布 CI 常见失败

| 现象 | 根因 | 处理 |
|---|---|---|
| `tag「x」不符合 {主题目录名}-v{语义化版本} 规范` | tag 命名不合法（如 `v0.1.0`、`foo-0.1.0`） | 按 5.2 重打 tag |
| `tag 版本(x) 与 theme.json 版本(y) 不一致` | 三重版本校验失败 | 对齐后删旧 tag 重打 |
| `theme.json 版本与 package.json 版本不一致` | 同上 | 对齐后重打 |
| `目录 x/ 不存在或缺少 theme.json` | tag 前缀与实际目录不符 / 清单缺失 | 核对目录名与清单 |
| `specifiers in the lockfile … don't match specs in package.json` | 改了依赖未重新生成 lockfile | `pnpm install` 更新 lockfile 并提交 |
| `Export encountered an error … took more than 60 seconds` | 构建期取数挂起（占位页未短路 / 请求无超时） | 按 §3.4 占位短路 + §3.2 超时改造 |
| 本地构建成功、CI 失败 | 本地 `.env.local` 有可达后端而 CI 没有 | 按 5.5 自检（`NEXT_PUBLIC_API_BASE_URL=` 空值覆盖本地 `.env.local`） |

重打 tag 的标准操作：

```bash
git tag -d tech-geek-v0.1.0 && git push origin :refs/tags/tech-geek-v0.1.0
# 修复后：
git tag tech-geek-v0.1.0 && git push origin tech-geek-v0.1.0
```

---

## 九、新主题改造清单（checklist）

以 tech-geek / lens-life-template 为参考实现：

- [ ] 仓库根新建主题目录（**目录名 = 主题 id**，匹配 `^[a-z0-9-]{2,50}$`），根目录新增 `theme.json`（id=目录名，version 与 package.json 一致，engine=next-static，api_compat=v1，routes.fallback 声明全部壳页面映射）
- [ ] `next.config.ts`：`output: "export"` + `distDir: "dist"` + `images.unoptimized`
- [ ] `lib/api/client.ts`：三级回退取 API 地址（`__NOVA_CONFIG__` → 环境变量 → 相对路径），缺省**不抛错、不回退 localhost**，请求带 15s 超时；全部取数函数 catch 后回退空数据
- [ ] `app/layout.tsx`：注入 `<Script src="/theme-config.js" strategy="beforeInteractive" />`
- [ ] 清除全部静态导出禁用项（§3.4 禁用清单）：`force-dynamic`、服务端 `searchParams`、route handlers、Server Actions 等
- [ ] 动态详情路由：`generateStaticParams` + `dynamicParams = false` + 占位 slug 短路（不发请求直接 404）
- [ ] 动态内容路由补壳页面（`public/article.html` 等），自包含、转义渲染、安全 URL、slug 解析容错、错误态
- [ ] 列表页 searchParams 改「静态页壳 + Suspense + 客户端列表组件」
- [ ] 服务端 fire-and-forget 副作用（浏览计数）迁移为客户端上报
- [ ] 工程一致性：lockfile 与 package.json 同步；不声明冲突的 `packageManager`
- [ ] （可选）`screenshots/` 提供市场展示图
- [ ] **离线构建自检**（5.5）：`NEXT_PUBLIC_API_BASE_URL=http://127.0.0.1:59999 pnpm build` 成功不挂起
- [ ] PR 合入后打 `{目录名}-v{版本}` tag 发布（如 `tech-geek-v0.1.0`），Release CI 自动产出制品，**无需改 CI**
- [ ] 验收：后台向导/模板页安装成功 → 激活 → 访客域即时呈现 → `/preview/{id}/` 可预览

---

## 十、分期与边界

- **v1 范围**：浏览 / 安装 / 切换 / 预览 / 卸载；不做主题内可视化配置、不做服务端 npm 构建（制品必须预构建）、不做自动更新（手动「检查更新→安装新版本→激活」）；
- **静态主题的固有限制**：`module-config` 等构建期取数的配置在制品内固化，CMS 侧变更后需重装/更新主题生效；未预渲染的详情页由壳页面客户端渲染，内容实时；
- **SSR 主题**（engine ≠ next-static）安装时被拒绝，待后续版本引入 `next-standalone` 引擎与 Node 运行时托管；
- **多租户**：blogger 即站点，激活指针天然随 blogger 走，远期按 Host 头路由到对应站点主题。
