<div align="center">

<img src="docs/assets/novablog-logo.png" alt="NovaBlog CMS" width="280" />

# NovaBlog CMS

**按你的爱好与职业预置好一切的开源博客系统 —— CMS 管内容，主题管颜值，AI 替你写**

<p>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Vue-3-4FC08D?style=flat-square&logo=vue.js&logoColor=white" alt="Vue 3" />
  <img src="https://img.shields.io/badge/TypeScript-5-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Redis-7+-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
</p>

[官方站点](http://www.novablog.cn) · [免费注册](http://www.novablog.cn/register) · [作者博客](http://czf.ditancafebar.cn) · [主题仓库](https://github.com/StudyNoWeekend/novablog-web) · [Docker 镜像](https://github.com/StudyNoWeekend/novablog-cms/pkgs/container/novablog-cms) · [Skill 文档](https://github.com/StudyNoWeekend/novablog-cms/blob/main/frontend/public/docs/novablog-api.md)

</div>

<div align="center">

**语言**：简体中文 · [繁體中文](README.zh-TW.md) · [English](README.en.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

</div>

---

## 🎯 三步拥有你的博客

不需要懂代码，也不需要懂服务器：

1. **开通博客** —— 到 [官方托管申请](http://www.novablog.cn/hosting#apply)即用官方托管，零部署开箱即用；想完全自己掌控，就按下方 [部署](#-部署两种方式) 章节用 Docker 自建；
2. **首装向导** —— 创建账号 → **选择创作方向** → 配置存储 → 自动装好默认主题，全程点几下鼠标；
3. **开始创作** —— 写文章、传照片、发菜谱、记旅行……你选的方向决定了后台预置哪些模块，之后随时可调。

首装完成后你就拥有一个能访问的完整博客：CMS 直接托管博客页面，访客访问的是你选的主题，内容数据纹丝不动。

### 按爱好 / 职业定制：创作方向

首装向导第 2 步会展示 **17 张「创作方向」卡片**（可多选），选中所爱，系统自动开启对应的模块；一个都不选则默认开启全部。登录后也能在后台顶部随时修改创作方向。

| 分组 | 创作方向 |
|------|----------|
| 技术 · 数码 | 技术开发者 · 数码测评 |
| 影像 · 声音 | 摄影爱好者 · 视频创作者 · 音乐人 · 旅行 Vlogger |
| 旅行 · 生活 | 旅行博主 · 美食爱好者 · 健身运动 · 时尚穿搭 · 宠物日常 |
| 兴趣 · 收藏 | 读书 · 游戏玩家 · 设计师/插画师 · 手工 DIY |
| 特殊 | 生活方式 · 全能创作者（一键开启全部模块） |

---

## 🧩 模块一览

后台模块随时可在「系统配置 → 模块管理」开关，关闭后前台与后台菜单同步隐藏（文章模块恒开，媒体库为基础设施恒可用）。

### ✍️ 内容创作（按你的爱好选装）

| 模块 | 给你什么 |
|------|----------|
| 📝 文章管理 | Markdown / 富文本双编辑器，分类标签、草稿/发布/下架、SEO（自定义 slug、摘要、封面、关键词）、阅读统计，一键推送官方社区 |
| 📸 摄影作品集 | 按主题组织照片集，自动提取 EXIF 拍摄参数，多种封面样式与水印保护 |
| 🗺️ 旅行攻略 | 结构化多日行程编辑器，高德 / Google 地图选点与详情页地图展示，点赞与热门推荐 |
| 🍳 美食菜谱 | 食材与烹饪步骤管理，一键推送官方社区 |
| 📖 读书书架 | 藏书、评分与书评，支持关联豆瓣图书信息 |
| 🎮 游戏库 | 游戏收藏、游玩时长与短评，支持关联 Bangumi 条目 |
| 💪 健身训练 | 训练记录管理与前台展示 |
| 🎵 音乐播放列表 | 上传歌曲建歌单，粘贴 B 站链接自动解析播放，导入网易云 / QQ 音乐 / Spotify 公开歌单 |
| 🎥 视频作品 | 视频与多对多专辑管理，B 站链接解析嵌入 |

### 👤 关于我（打造个人名片）

| 模块 | 给你什么 |
|------|----------|
| 个人资料 | 昵称头像简介、博客标题/描述/背景、20+ 社交平台链接（内置品牌色图标） |
| 个人设备 | 数码/摄影装备清单，上传照片自动识别拍摄参数 |
| 项目经历 | 在线简历式项目履历：领域、角色、技术栈、起止时间 |
| 技术栈 | 技能标签管理，前台用官方品牌图标展示 |
| 开源作品 | GitHub 开源项目展示卡，一键推送官方社区 |

### 🧰 基础设施（开箱自带）

| 模块 | 给你什么 |
|------|----------|
| 工作台 | 访问量/内容量指标、各模块内容卡片、ECharts 趋势图 |
| 评论管理 | 汇总文章、旅行攻略等来源的访客评论，三级审核、楼中楼回复、IP 黑名单 |
| 媒体库 | 图片/视频/音频上传（大文件分片）、文件夹管理、尺寸预设工作台、外部图片导入与搜索 |
| 主题市场 | 官方主题浏览、预览、一键安装激活（见下方主题章节） |
| 系统配置 | 模块开关、对象存储（本地/MinIO/阿里云 OSS/腾讯云 COS 热切换与迁移）、跨域、地图 Key、IP 黑名单、社区推送 |
| 访问统计 | 前台访客 IP 统计，ip2region 离线解析归属地，不依赖外部服务 |
| API 文档 / MCP 服务 | 免认证公开 API 在线浏览、Skill 文档下载、MCP 密钥管理（见下方 AI 章节） |

---

## 🤖 让 AI 替你写：MCP + Skill

NovaBlog CMS 内置 **MCP Server** 与 **Skill 格式接口文档**，Codex、ZCode、Claude Code、Cursor 等 AI Agent 都能直接在你的博客上创作内容——你动嘴，它动手。

### 方式一：MCP 接入（推荐，可写可发可管理）

后台「AI → MCP 服务」三步接入：

1. **新建密钥**：自助开通 MCP API Key（`nbt_mcp_` 前缀，可设有效期，内置限流保护）；
2. **粘贴配置**：复制页面自动生成的 JSON，粘进 AI 客户端的 MCP 配置：

```json
{
  "mcpServers": {
    "novablog": {
      "type": "http",
      "url": "https://你的博客域名/api/v1/mcp",
      "headers": {
        "Authorization": "Bearer nbt_mcp_你的密钥"
      }
    }
  }
}
```

3. **开口创作**：对 AI 说「帮我写一篇 XX 主题的文章并发布」即可。

MCP 端点为 Streamable HTTP（`/api/v1/mcp`），当前提供 9 个工具：

| 工具 | 作用 |
|------|------|
| `create_article` | 写文章：标题、Markdown 正文、摘要、标签/分类（自动创建）、封面、状态 |
| `update_article` | 修改文章 |
| `set_article_status` | 发布 / 转草稿 / 下架 |
| `delete_article` | 删除文章（需显式 confirm，防误删） |
| `list_articles` / `get_article` | 查询文章列表与详情（关键词、分页） |
| `upload_image` | 上传图片：http 链接自动转存或 base64，返回可引用 URL |
| `list_categories` / `list_tags` | 查询分类与标签 |

### 方式二：Skill 文档（让 AI 掌握全部公开 API）

后台「AI → API 文档」页可下载 **SKILL.md 格式的接口参考**（[novablog-api.md](frontend/public/docs/novablog-api.md)），放进 Claude Code / Cursor / ZCode 等 AI 助手的工作区，它就能掌握全部免认证公开 API 的路径、参数与响应示例，帮你写主题、做集成、批量取数。

---

## 🏠 官方生态：NovaBlog 官方站与 CMS 的关系

一句话：**CMS 是你自己的博客核心，官方 NovaBlog 站是你的内容社区。**

- **NovaBlog CMS（本项目）**：自托管的开源博客系统，管内容、管外观、直接托管博客页面；
- **NovaBlog 官方站**：社区门户——官方账号体系、主题市场、内容社区；CMS 的主题市场与社区推送都连向官方站；
- **官方地址统一托管**：出厂内置、后台「系统配置 → 社区推送」可查改，主题市场与社区推送共用同一入口，服务端直连。

| 入口 | 地址 | 用途 |
|------|------|------|
| 官方站点 | <http://www.novablog.cn> | 社区门户 · 主题市场 |
| 免费注册 | <http://www.novablog.cn/register> | 开通官方账号：市场登录、托管部署、推送密钥 |
| 官方 API | <http://api.novablog.cn> | 社区推送与主题市场的统一服务入口 |

> 后台「系统配置 → 社区推送」中的官方地址以实际配置为准。

**社区推送**：文章、旅行攻略、美食菜谱、摄影作品集、开源作品 5 类内容可一键同步到官方社区获得曝光——内容创建/更新后自动异步推送、删除自动下架、无变化自动跳过、失败自动重试，编辑页内的推送按钮也支持手动重推与查看同步状态。

---

## 🎨 主题：换外观像换衣服

博客前端是可整体替换的「主题」：主题负责外观与体验，CMS 负责内容与托管。切换主题只更新一个指针，秒级生效、不重启服务、内容纹丝不动。

### 方式一：官方市场一键安装（推荐）

后台「主题市场」浏览、搜索、评分、收藏 11+ 风格各异的官方主题（极客风、镜头生活、漫游世界、旋律笔记……），选中即可：

- **激活前预览**真实渲染效果（`/preview`）；
- **一键安装**：自动完成下载 → SHA256 校验 → 安全校验解压 → `theme.json` 清单校验；
- **版本管理与一键回滚**，主题通过 `theme.json` 声明配置项，后台自动生成可视化设置面板，改设置不碰代码。

### 方式二：本地挂载自备前端

想用自己的前端？把含 `theme.json` 与 `dist/` 的目录挂载给 CMS 即可，零改造接入：裸机部署设 `themes.frontend_dir`，Docker 部署用 `./deploy.sh --frontend-dir <目录>`。

### 🌍 主题开源开放，和我们一起共建

全部官方主题开源在主题仓库 [StudyNoWeekend/novablog-web](https://github.com/StudyNoWeekend/novablog-web)，采用**官方 + 社区共同维护**的开放模式：官方持续产出与维护官方主题，社区作者按规范开发自己的主题提交 Issue / PR，经官方市场分发给所有博主。

- 主题包规范：[novablog-web/docs/theme-spec.md](https://github.com/StudyNoWeekend/novablog-web/blob/main/docs/theme-spec.md)
- 打包与发布指南：[novablog-主题包规范与发布指南.md](novablog-主题包规范与发布指南.md)

---

## 🚀 部署：两种方式

### 方式一：官方托管（零部署，小白推荐）

到 [官方站点免费注册](http://www.novablog.cn/register)即可开通托管博客，无需服务器、无需 Docker，注册后进入与自部署完全一致的首装向导。

### 方式二：Docker 自部署

单镜像内置 nginx 双入口、Go 后端与管理后台前端：`ghcr.io/studynoweekend/novablog-cms`（`linux/amd64` / `linux/arm64`）。需要自备 PostgreSQL 15+ 与 Redis 7+ 实例（脚本会做连通性预检）：

```bash
cd deploy
./deploy.sh --name myblog                        # 交互问答补全剩余配置
./deploy.sh --name myblog --yes \
  --db-host 10.0.0.5 --db-password '***' \
  --redis-host 10.0.0.5 --redis-password '***'   # 非交互全默认部署
```

启动后：博客入口 `http://<域名>:80`，后台入口 `http://<域名>:8080/admin/`。

| 入口 | 默认端口 | 内容 |
|------|----------|------|
| 博客入口 | `80` | 激活主题页面（或本地挂载前端）· `/preview` · `/api/v1` · `/files` |
| 后台入口 | `8080` | `/admin/` 管理后台 · `/api/v1` · `/files` |

常用运维：

```bash
./deploy.sh --name myblog --status    # 查看容器状态
./deploy.sh --name myblog --logs      # 跟踪应用日志
./deploy.sh --name myblog --down      # 停止并移除容器（挂载数据保留）
```

- `--name` 必填，作为同机多博客的隔离键，不同名称与端口即可同机共存；
- 升级重跑 `./deploy.sh --name myblog --version vX.Y.Z`，挂载数据保留；
- 完整参数、外部数据库、宿主机 nginx 反代说明见 [deploy/README.md](deploy/README.md)。

### 本地开发

```bash
# 1. 启动后端
cd backend
cp config/config.example.yaml config/config.yaml   # 编辑 postgres/redis 连接与密钥
make run                                            # :8111，首次启动自动迁移

# 2. 启动管理后台前端
cd ../frontend
pnpm install && pnpm dev                            # :5173，/api 代理到 :8111
```

浏览器访问 `http://localhost:5173`，自动进入首装向导。

---

## 🤝 贡献

欢迎提交 Issue 与 Pull Request：

1. 后端提交前执行 `make fmt && make vet && make lint`；
2. 前端提交前执行 `pnpm lint`；
3. 新增 API 时同步更新 DTO 与 API 文档。

---

## 📜 开源协议

本项目基于 [MIT](LICENSE) 协议开源。

---

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) · [GORM](https://gorm.io) · [Vue.js](https://vuejs.org) · [Ant Design Vue](https://antdv.com) · [PostgreSQL](https://www.postgresql.org) · [Redis](https://redis.io)
- 以及所有为本项目提供灵感与支持的社区贡献者

---

<p align="center">
  <sub>Made with ❤️ for independent bloggers</sub>
</p>
