# NovaBlog 社区推送 · 内容类型与字段核对文档

> 目的：CMS（novablog-admin）推送到 NovaBlog 社区前，对「内容类型 + 内容字段 + 渲染形态」做两端一致性核对，避免同一篇文章在博客端与社区端展示出现误差。
>
> 依据：本文所有结论均来自 NovaBlog 官方仓库**源码实读**（`backend/` Go 后端 + `frontend/` Next.js 社区前端），非对接文档推断；关键位置标注了源码路径，可直接回查。
> 适用版本：novablog backend v1.0.6+（对接文档 v1.1 所对应版本）。

---

## 1. 内容类型接口

### 1.1 两个类型接口，同一数据源

| 接口 | 鉴权 | 用途 |
|---|---|---|
| `GET /api/v1/push/types` | 推送密钥（Bearer nbk_…） | CMS 推送前获取可用类型 |
| `GET /api/v1/articles/types` | 无需鉴权 | 社区前端筛选 chips 数据源 |

两者在官方后端调用同一个查询函数 `listEnabledArticleTypes`（`backend/internal/logic/article_logic.go:112`），返回完全一致。

**响应结构**（`backend/internal/dto/res/article_res.go:3`）：

```json
{
  "code": 200, "message": "success",
  "data": [
    { "id": 1, "slug": "article",    "name": "文章",     "sort_order": 1 },
    { "id": 2, "slug": "travel",     "name": "旅行攻略", "sort_order": 2 },
    { "id": 3, "slug": "recipe",     "name": "美食食谱", "sort_order": 3 },
    { "id": 4, "slug": "portfolio",  "name": "作品集",   "sort_order": 4 },
    { "id": 5, "slug": "opensource", "name": "开源作品", "sort_order": 5 }
  ]
}
```

### 1.2 类型字典机制（核对必读）

- 存储表 `article_types`（`backend/internal/model/article_type.go`）：`slug` 唯一索引、`status` 1 启用 / 0 停用；查询固定 `status = 1` 且按 `sort_order asc, id asc` 排序。
- **字典是动态的**：官方管理后台有 `/admin/article-types` 增删改（`router.go:267`），可用类型不限于初始 5 类。CMS 端**不可硬编码**这 5 个 slug，应启动/配置时调接口缓存。
- 推送校验（`backend/internal/logic/article_push_logic.go:63`）：`slug` 不存在**或已停用** → HTTP 400，message：`内容类型无效或已停用，可通过 GET /api/v1/push/types 获取可用类型`。
- ⚠️ **类型被官方停用后**，已推送文章在社区仍保留原 slug，但筛选 chips 中不再出现，新推送会被拒。CMS 端应对已推送文章记录所用 slug，类型字典变更时提示博主。

### 1.3 社区端类型展示形态（`frontend/components/article-card.tsx:9`）

- 列表卡片与详情页的类型徽章 = **色点 + 名称**，名称来自接口 `type_name`，**色值硬编码在前端**：

| slug | 徽章色点 |
|---|---|
| `article` | slate（灰蓝） |
| `travel` | emerald（绿） |
| `recipe` | orange（橙） |
| `portfolio` | violet（紫） |
| `opensource` | sky（天蓝） |
| 其他（官方新加的类型） | **中性灰**（回退色 `bg-muted-foreground`） |

> 核对点：如果官方后续新增类型，CMS 推送成功但社区徽章只有灰点，这是社区前端既定行为，不是推送故障。

---

## 2. 推送字段规格 × CMS 本地字段 × 社区展示 对照表

推送接口 `POST /api/v1/push/articles` 的请求 DTO（`backend/internal/dto/req/push_req.go:11`），与 CMS（novablog-admin `articles` 表）逐字段对照：

| # | 推送字段 | 官方约束（binding，超限→400） | CMS 本地字段 | 转换/存储规则（官方源码行为） | 社区展示用途 | 核对要点 |
|---|---|---|---|---|---|---|
| 1 | `external_id` | required，≤128 字符 | 文章 UUID 主键 | 与博主账号组成**账号内唯一**复合索引 `idx_user_external`（`model/article.go:22`）；重复推送=整体更新 | 不展示 | 同一文章**永不更换**；删除后新建的文章要用新 UUID |
| 2 | `type` | required，≤30 字符，须为启用中 slug | ❗CMS 无对应字段（CMS 的 `type` 是**编辑器类型** 1=MD/2=HTML，勿混淆） | 原样存 `articles.type` 列（存 slug 字符串，非外键） | 列表/详情类型徽章 + 类型筛选 | 需建立映射：推送弹窗选类型 / 配置默认类型；slug 失效即 400 |
| 3 | `title` | required，≤200 | `title` varchar(200) | 原样存储 | 列表卡片标题、详情页 H1 | 长度天然一致；注意首尾空白是否要 TrimSpace |
| 4 | `content_markdown` | required，≥1，**≤200KB（按 UTF-8 字节 `len()` 计）**（`article_push_logic.go:17,56`） | `content` text（仅 MD 文章） | 原样存 `content` 列，后端**不做任何处理** | 详情页 Markdown 渲染（见 §4） | ① CMS 端无大小限制，超 200KB 会被 400；② **正文内嵌 HTML 会被社区静默丢弃**（见 §4.2，最大误差源） |
| 5 | `summary` | omitempty，≤500 | `summary` varchar(500) | 原样存储；**为空时社区读取端截正文前 200 rune + `...`，截的是 Markdown 源码，不剥离语法**（`article_logic.go:134`、`community_logic.go:537`） | 列表卡片摘要，前端 CSS 固定 2 行截断（line-clamp-2） | 若 CMS 不传 summary，社区列表摘要会带 `#`、`![]()` 等 Markdown 符号 → **CMS 应保证已发布文章都有摘要** |
| 6 | `cover_image` | omitempty，≤500，**无 URL 格式校验** | `cover_image` | 原样存储 | 列表缩略图 176×128（md，object-cover 裁切；**移动端 <sm 不显示封面**）；详情头图最大高 420px | ⚠️ CMS 本地上传的封面若存**相对路径**，推送前必须拼成绝对 URL，否则社区破图 |
| 7 | `tags` | omitempty，≤10 个，单个 ≤30（`max=10,dive,max=30`） | 多对多 `tags`（TagNames） | **全删重建**（每次推送全量替换）；每篇每标签一行，无全局标签表、无去重约束；TrimSpace 后空白项跳过（`article_push_logic.go:119`） | 卡片/详情 `#标签` 纯展示，**不可点击、无标签聚合页** | CMS 标签 >10 个会被 400，需截断或提示；社区标签无筛选入口，仅装饰性展示 |
| 8 | `source_url` | omitempty，≤500，无 URL 格式校验 | ❗无直接对应（可由博客前台地址 + slug 拼接） | 原样存储 | 详情页「阅读原文」卡片（非空才显示，单行截断展示完整 URL） | 不传则详情页无「阅读原文」入口 |
| 9 | `published_at` | `*time.Time`，**必须 RFC3339**（如 `2026-10-10T09:00:00+08:00`），非法→400 | `published_at` timestamptz | 创建：缺省填 `time.Now()`；更新：**仅传入时才覆盖，不传保留旧值**（`publishedAtOrNow`，`article_push_logic.go:209`） | **列表排序依据**（`published_at DESC, id DESC`）+ 卡片/详情日期展示 `YYYY-MM-DD`（`slice(0,10)`） | CMS 端格式化为 RFC3339 时区串；编辑重推时**每次都要带上原 published_at**，否则若首次推送时没传，社区排序时间=首次推送时间而非真实发布时间 |
| 10 | `status` | omitempty，oneof=`published`/`hidden` | CMS 状态 1=草稿 / 2=已发布 / 3=已下架 | `published`→社区 1（发布）；`hidden`→社区 2（下架）；缺省 `published` | 社区列表/详情**仅返回 status=1**；hidden 即读者不可见 | 映射：CMS 2→`published`；CMS 3（或转草稿）→`hidden`；CMS 1（草稿）**不推送**。注意社区还有第 3 态 `deleted`（仅对账接口可见） |

**CMS 字段在社区无落点（提前知晓，不算误差但要确认预期）：**

| CMS 字段 | 社区对应 | 说明 |
|---|---|---|
| `category`（分类） | ❌ 无任何对应字段 | 社区只有「类型 + 标签」两个维度。分类语义只能：映射为 type（如"旅行"分类→travel），或并入 tags，或放弃同步。**这是两端信息架构最大的结构性差异，需要博主确认预期** |
| 浏览数 | 社区独立计数 | 社区 `view_count` 从 0 开始自己累计，详情页打开即 +1，与 CMS 阅读数无关 |
| 评论 | ❌ 不互通 | 推送文章在社区的评论是社区用户发的，CMS 端不可见 |

---

## 3. 读者侧公开接口（核对数据源）

推送完成后，用以下**无需鉴权**的接口拉取社区实际存储/展示的数据，与 CMS 本地比对：

### 3.1 列表 `GET /api/v1/articles?page=1&size=10&type=<slug>&tag=<标签名>&search=<关键词>`

- 响应 `data.items[]` 单元素结构（`backend/internal/dto/res/article_res.go:13`，字段名以此为准）：

```json
{
  "id": 128,
  "type": "travel",
  "type_name": "旅行攻略",
  "title": "大理三日游攻略",
  "summary": "洱海环线与古城漫步",
  "cover_image": "https://blog.example.com/images/dali.jpg",
  "tags": ["旅行", "大理"],
  "user_id": "…博主uuid…",
  "username": "alice",
  "user_avatar": "…",
  "blog_url": "https://blog.example.com",
  "view_count": 3,
  "status": 1,
  "published_at": "2026-10-10 09:00",
  "created_at": "3 天前",
  "updated_at": "2 小时前"
}
```

- **不含**正文、`source_url`、`external_id`。
- `published_at` 输出格式是 `2006-01-02 15:04`（`formatDateTime`，`push_credential_logic.go:132`）；`created_at/updated_at` 是相对时间文案（"N 天前"，超 30 天转日期）。
- 排序：`published_at DESC, id DESC`，仅返回 status=1。
- `type`/`tag` 均为**精确匹配**（tag 用子查询 `article_tags.name = ?`）；`search` 匹配 title/summary 的小写模糊。
- 分页：size 1–50，默认 10。
- ⚠️ 社区前端实际只用了 `page/size/type` 三个参数（`frontend/lib/api.ts:146`、`article-feed.tsx:37`），`tag`/`search` 后端支持但前端无入口。

### 3.2 详情 `GET /api/v1/articles/:id`

```json
{
  "…同列表字段…",
  "content": "# 大理\n\n环洱海骑行两天……（Markdown 原文，原样返回）",
  "source_url": "https://blog.example.com/posts/1024"
}
```

- `content` 是推送时 `content_markdown` 的**原文**，后端零处理；渲染与消毒完全由前端负责（`article_logic.go:15` 注释）。
- 仅 status=1 的文章可访问（否则 404）；访问会使社区 `view_count` +1（**核对时注意会污染阅读数**）。
- 详情响应含 `source_url`（列表不含），核对「阅读原文」链接以此为准。

### 3.3 对账 `GET /api/v1/push/articles?page=1&size=50`（需密钥）

返回本账号全部推送文章的 `article_id / external_id / title / status(1发布|2下架|3删除) / updated_at`，按 `updated_at` 倒序。用于批量核对两端 external_id 集合是否漂移。

---

## 4. 社区端渲染事实（误差风险清单）

### 4.1 Markdown 渲染能力边界（详情页正文）

社区前端用 `react-markdown v10 + remark-gfm`，**仅此一个插件**（`frontend/app/community/articles/[id]/page.tsx:121`）：

| 语法 | 社区端表现 |
|---|---|
| GFM 表格、任务列表、删除线、自动链接、脚注 | ✅ 支持（remark-gfm 内置） |
| 标题/列表/引用/代码块/行内代码/图片/链接 | ✅ 支持（标准 Markdown） |
| **代码语法高亮** | ❌ 无 highlight.js/prism，代码块只有底色边框纯文本样式（CMS 端 ``` 后带语言不报错，但不着色） |
| **数学公式**（`$...$`、KaTeX） | ❌ 不支持，按普通文本显示 |
| Markdown 扩展（如 `==高亮==`、上标下标） | ❌ 按原文显示符号 |
| **内嵌 HTML** | ❌ **静默丢弃**（见 4.2） |

### 4.2 ⚠️ 内嵌 HTML 会被静默丢弃（最高优先级误差源）

社区前端未启用 `rehype-raw`/`rehype-sanitize`，react-markdown 默认**直接跳过原始 HTML 节点**（源码注释自证：`page.tsx:18`）。即 CMS Markdown 正文里混写的：

```markdown
这段文字中的 <u>下划线内容</u> 和 <span style="color:red">红字</span> 会被社区丢弃
```

在社区端渲染为：`这段文字中的 和 会被社区丢弃` ——**既不显示效果、也不显示标签原文，内容直接消失，且推送校验不报错（推送成功 200）**。

> 核对动作：CMS 端推送前扫描正文中的 HTML 标签（`<[a-zA-Z][^>]*>`），命中即提示博主改写为纯 Markdown；或确认本文无 HTML 后再推送。这也决定了**只有编辑器类型=MD 的文章可推送**（HTML 文章正文几乎全是标签，推过去等于空白）。

### 4.3 图片

- 正文/封面图片一律**外链直出**（原生 `<img>`，不走 next/image 压缩）；`max-width:100%`、圆角，**无点击放大、无懒加载错误占位**。
- **防盗链风险**：图床若校验 Referer，社区端会破图（浏览器默认行为），CMS 端看图正常、社区端挂掉——图床需允许外站引用。
- CMS 本地上传的图片若为相对路径（`/uploads/...`），**必须拼绝对地址后再推送**（正文 `![]()` 内与 `cover_image` 字段同理）。

### 4.4 其他展示层既定行为

| 项 | 社区端行为 | 备注 |
|---|---|---|
| 摘要 | 显示接口 `summary` 原文，CSS 固定 2 行截断，前端无二次处理 | 摘要过长只影响截断观感 |
| 日期 | `published_at.slice(0,10)` → `YYYY-MM-DD`；**published_at 为空时回退 `created_at` 且未 slice，会显示完整 ISO 串**（前端小瑕疵，`article-card.tsx:38`） | CMS 保证 published_at 必传可规避 |
| 作者头像 | **首字母假头像**（按 username 首字符取色），接口 `user_avatar` 被忽略 | CMS/官方头像设置不影响文章卡片 |
| 作者博客链接 | 底部「作者博客：{blog_url}」，来自博主账号资料，非文章字段 | 在 NovaBlog 个人中心维护 |
| 类型徽章 | 色点+名称，色值硬编码 5 类，扩展类型灰点 | 见 §1.3 |
| 标签 | `#标签` 纯展示，不可点击、无聚合页 | 对接文档 8.1 所述「标签聚合浏览」目前前端未实现 |
| 阅读数 | 社区自计数，详情页打开 +1 | 与 CMS 阅读数无同步关系 |

---

## 5. 状态与生命周期映射

| CMS 端动作 | 推送调用 | 社区结果 | 核对点 |
|---|---|---|---|
| 发布（status 1→2） | POST，`status=published`（首次=创建，重复=更新） | status=1，列表/详情可见 | result 返回 `created`/`updated` |
| 编辑已发布文章 | POST 同 external_id（全量覆盖，标签全量替换） | 内容更新 | **每次都带 published_at 原值** |
| 本地下架/转草稿（status→3/1） | POST，`status=hidden` | status=2，读者不可见但数据保留 | 重新推 `published` 即恢复 |
| 本地删除 | DELETE `/push/articles/:external_id` | status=3（软删） | 幂等；`not_found` 也返回 200 |
| 社区被管理员下架 | 更新推送被拒 400：`文章已被管理员下架，无法推送更新…`（`admin_locked=true`，`admin_article_logic.go:73`） | 读者不可见 | CMS 标记锁定、停止重试；管理员恢复后自动解锁。**删除同步不受锁定拦截** |

**关键不对称**：CMS 三态（草稿/发布/下架）↔ 社区推送接口只有 `published/hidden` 两态 + 删除。草稿态文章不推送；「从社区隐藏」用 hidden 而非删除，才能保住社区端评论/阅读数。

---

## 6. 核对操作指引

### 6.1 推送前 CMS 本地自检清单

1. 文章编辑器类型 = MD（HTML 文章不推）。
2. 正文扫描无内嵌 HTML 标签（§4.2，否则社区静默丢内容）。
3. 正文字节数 ≤ 200KB；标题 ≤200 字；摘要 ≤500 字；标签 ≤10 个且单个 ≤30 字。
4. 封面与正文图片均为**绝对 URL** 且图床允许外站引用。
5. 摘要非空（避免社区截 Markdown 源码前 200 字符当摘要）。
6. `published_at` 准备为 RFC3339（CMS 真实发布时间，非推送时间）。
7. type 已从 `/push/types` 缓存中选定（勿硬编码）。
8. `source_url` 准备好博客永久链接。

### 6.2 推送后逐字段核对步骤

```bash
BASE="https://<官方社区域名>/api/v1"

# ① 类型字典核对（CMS 缓存 vs 官方当前字典）
curl -s "$BASE/articles/types"

# ② 列表核对：找到刚推的文章，比对 type/type_name/title/summary/cover_image/tags/published_at
curl -s "$BASE/articles?type=<slug>&page=1&size=10"

# ③ 详情核对：content 与 CMS Markdown 原文逐字节 diff；source_url 是否正确
curl -s "$BASE/articles/<article_id>"

# ④ 对账：external_id 集合与状态
curl -s "$BASE/push/articles?page=1&size=50" -H "Authorization: Bearer nbk_…"
```

浏览器侧最后一步：社区首页 → 文章 tab → 找到该文章，肉眼核对卡片（徽章颜色/摘要两行/封面/日期）与详情页（表格/代码块/图片/脚注渲染、「阅读原文」卡片）。

### 6.3 建议先发官方确认的问题清单

1. **HTML 内嵌静默丢弃**：是否考虑推送校验时返回警告，或前端至少显示占位提示？（当前 200 成功但内容缺失，博主无感知）
2. **summary 缺省截断**：截的是 Markdown 源码（含 `#`、`![]()` 符号），是否会改为剥离语法后的纯文本？
3. **代码高亮 / 数学公式**：社区是否计划支持（rehype-highlight / KaTeX）？CMS 端大量技术文依赖代码高亮。
4. **扩展类型徽章**：新类型在社区只有灰点，色值映射是否会做成接口下发？
5. **前端小瑕疵**：`published_at` 为空时卡片日期回退 `created_at` 未截断（显示完整 ISO 串）；`user_avatar` 接口有值但前端用首字母假头像；标签不可点击无聚合页——与对接文档 8.1「社区侧聚合浏览」描述不符，是否在 roadmap？
6. **published_at 更新语义**：更新不传则保留旧值——CMS 端按「每次全量传」处理是否为官方推荐姿势？
7. 分类维度：CMS 分类在社区无落点，官方建议映射到 type 还是并入 tags？

---

## 附：本文事实来源索引

| 主题 | 源码位置 |
|---|---|
| 类型 DTO / 字典表 / 种子 / 校验 | `novablog/backend/internal/dto/res/article_res.go:3`、`internal/model/article_type.go`、`internal/logic/article_logic.go:112`、`internal/logic/article_push_logic.go:63` |
| 推送请求 DTO / 200KB / 标签重建 / publishedAtOrNow / 锁定 | `novablog/backend/internal/dto/req/push_req.go:11`、`internal/logic/article_push_logic.go:17,56,99,119,209` |
| 读者列表/详情 DTO、截断、排序、格式化 | `novablog/backend/internal/dto/res/article_res.go:13,33`、`internal/logic/article_logic.go:33,88,134,152`、`internal/logic/community_logic.go:537`、`internal/logic/push_credential_logic.go:132` |
| 前端卡片 / 徽章色 / 日期 / 头像 | `novablog/frontend/components/article-card.tsx:9,38,51,82`、`frontend/components/avatar.tsx` |
| Markdown 渲染与 HTML 丢弃 | `novablog/frontend/app/community/articles/[id]/page.tsx:18,121`、`frontend/app/globals.css:158` |
| 管理员锁定 | `novablog/backend/internal/logic/admin_article_logic.go:73` |

---

## 7. 契约 v2：extra 结构化参数与 CMS 落地模块（2026-10 更新）

### 7.1 契约 v2：围绕 CMS 结构

社区推送接口 `POST /push/articles` 新增可选 `extra`（JSON 对象，≤64KB，更新时整体覆盖、缺省清空）。**键名与 CMS 五表字段同名直传**，社区存储为 jsonb 并按类型原生渲染——CMS 不再需要把结构化内容全部压平进 markdown。已知类型的已知键做类别宽松校验（类别不符返回 400），未知键透传。

| CMS 表 | 推送 type | content_markdown（合成保底） | extra（结构直传） |
|---|---|---|---|
| articles | `article` | `Content`（正文图片相对路径自动绝对化） | 无强约束 |
| travel_guides | `travel` | 信息行 + 景点 + 行程 + 游记 合成 | `{destination, region, days, best_month, attractions, itinerary, reviews}` |
| recipes | `recipe` | 食材表格 + 步骤 合成；`Cover`→cover_image | `{ingredients, steps, difficulty, minutes, servings}` |
| portfolios | `portfolio` | 图集合成；封面走 ResolveCoverURLs | `{items: [{title, description, image}]}` |
| open_source_works | `opensource`（CMS 内部 open_source，推送时映射） | Readme 直用 | `{repo_url, language, topics, stars, homepage}`；source_url 优先取 RepoURL |

其余契约不变：external_id=各表 UUID 主键、tags ≤10/≤30（超限截断并记入 warnings）、正文 ≤200KB、published_at=article.PublishedAt??CreatedAt（其余类型 CreatedAt，重推稳定）、status 映射（article/travel: 2→published、3/1→hidden；recipe/portfolio/open_source: 1→published、0→hidden）。

### 7.2 CMS 落地模块（backend）

| 模块 | 位置 | 说明 |
|---|---|---|
| 推送客户端 | `pkg/novapush/client.go` | Ping/Types/UpsertArticle/DeleteArticle/ListSync，仿 novablogapi 模式 |
| 配置 | `community:` yaml 段 + `community_push_configs` 单行表 | base_url/site_url/push_key，后台「系统设置→社区推送」可改，密钥读取脱敏 |
| 同步状态表 | `community_sync_records` | TargetType+TargetID 唯一；synced/hidden/failed/deleted + ContentHash（无变化跳过推送）+ RetryCount |
| 构建器/预检 | `internal/logic/push_payload.go` | 五表→payload；标签截断、摘要自动生成（剥离 markdown 语法）、图片相对路径绝对化 |
| 推送服务 | `internal/logic/community_push_logic.go` | PushToCommunity/HideFromCommunity/RemoveFromCommunity + SyncOnUpdatedAsync/SyncOnDeletedAsync 钩子（仅对已开启推送的内容生效） |
| 钩子接线 | article/travel/recipe/portfolio/open_source 五个 logic 的 Create/Update/UpdateStatus/Delete | 显式异步调用，失败落状态表不阻塞保存 |
| 重试 worker | `cmd/api/main.go` startCommunityPushRetryWorker | 每 5 分钟扫 failed 且 RetryCount<5 重推最新内容 |
| 管理 API | `/community-push/config|test|state|states|sync|records|types` | 后台登录态 |
| 管理界面 | 系统设置→社区推送 tab；各内容列表/编辑页 PushSyncButton | 按钮按推送状态渲染（推送/下架/失败重试），同步记录页支持单条重推 |

### 7.3 与 §2/§5 的关系

- §2 表格中「CMS 无对应字段」的 `type`（内容类型）在 v2 由推送侧按内容表决定并映射 slug；`source_url` 由 site_url + 类型路由模板拼接（开源作品优先 RepoURL）；`published_at` 各表策略见上表。
- §4.2「内嵌 HTML 静默丢弃」仍为社区既定行为；CMS 富文本编辑器计划移除，届时该风险自然消除（本期推送模块不做 HTML 预检）。
- §5 生命周期映射不变；新增语义：**博主开启推送的内容，后续更新/下架/删除自动同步**（未开启的内容不受影响）。
