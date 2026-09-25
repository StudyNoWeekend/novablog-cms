# NovaBlog CMS

> 开源博客系统：一个管内容的核心 CMS + 随时可换的博客前端（主题）

<p>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Vue-3-4FC08D?style=flat-square&logo=vue.js&logoColor=white" alt="Vue 3" />
  <img src="https://img.shields.io/badge/TypeScript-5-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Redis-7+-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
</p>

---

## 它如何工作：CMS × 主题

NovaBlog 把博客拆成两层：

- **NovaBlog CMS（本项目）** —— 内容与托管的唯一核心。Vue 3 管理后台供博主创作与管理，Go 后端负责数据、接口与安全，同时**直接托管博客页面**；
- **博客主题** —— 面向访客的博客前端，是可整体替换的「主题资产」，全部来自官方主题仓库 [novablog-web](https://github.com/StudyNoWeekend/novablog-web)：极客风、镜头生活、漫游世界、旋律笔记等 11+ 风格各异的主题持续上架。

```
┌──────────────────────────────────────────────┐
│            NovaBlog CMS（本项目）              │
│                                              │
│   Vue 3 管理后台  —— 内容创作 / 系统配置         │
│   Go 后端        —— 数据中心 / 公开 API / 安全  │
│   博客页面托管    —— 直接对外服务访客            │
└──────────────┬────────────────▲──────────────┘
               │ 拉取安装        │ 调用免认证公开 API 读取内容
               ▼                │
┌──────────────────────────────┴──────────────┐
│         novablog-web（官方主题仓库）           │
│    极客风 / 镜头生活 / 漫游世界 / 旋律笔记 …     │
│    Release CI 自动打包制品，后台一键安装         │
└─────────────────────────────────────────────┘
```

一句话：**主题负责外观与体验，CMS 负责内容与托管**。主题随时可换、可回滚，内容纹丝不动；主题通过 CMS 的免认证公开 API 取数，页面也由 CMS 直接托管——部署时只需要这一个服务，主题切换秒级生效，无需重新构建或部署任何前端项目。

---

## ✨ 核心亮点

- **🎨 主题热插拔**：博客前端是可整体替换的资产——市场一键安装、激活前 `/preview` 预览真实效果、版本管理与一键回滚；切换只更新一个指针，秒级生效、不重启服务。
- **🔌 两条路拿到主题**：官方市场在线安装（自动完成下载 → SHA256 校验 → 安全解压 → 清单校验），或本地挂载自备前端目录，自定义主题同样零改造接入。
- **🤖 AI 友好的开放 API**：内容全量读取走免认证公开接口，并配套 **Skill 格式接口文档**——投喂给 AI 编码助手，它就能掌握全部接口帮你写主题、做集成。
- **🧩 模块化博客**：文章、作品集、视频、旅行、音乐、设备、项目经历、开源作品等 9 大模块在后台随时开关，主题按配置自动显隐对应入口。
- **💾 存储可插拔**：本地 / MinIO / 阿里云 OSS / 腾讯云 COS 后台可视化配置、热切换，并支持存量文件跨存储迁移。
- **🚀 三步首装向导**：创建账号 → 初始化存储 → 自动拉取默认主题，完成后即得一个能访问的完整博客，零配置门槛。
- **🔒 安全开箱即用**：JWT 双令牌无感刷新、令牌/IP 黑名单、后台动态管理 CORS、访问日志（离线 IP 归属地解析，不依赖外部 API）。

---

## 🎨 主题：获取与开发

### 方式一：官方市场在线安装（推荐）

后台「主题 → 官方市场」浏览、搜索、评分、收藏社区主题，选中即可一键安装激活：

1. CMS 从官方市场查询主题与版本，自动从 [novablog-web](https://github.com/StudyNoWeekend/novablog-web) 的 GitHub Release 拉取预构建 tar.gz 制品；
2. 安装链路内置安全闭环：SHA256 完整性校验 → 路径穿越防护解压 → `theme.json` 清单严格校验 → 落盘；
3. 激活即对外生效，支持版本管理与一键回滚；激活前可通过 `/preview/:theme_id` 预览真实渲染效果；
4. 主题通过 `theme.json` 声明配置项，后台自动生成可视化设置面板，改主题设置无需碰代码。

首次安装向导最后一步会自动拉取并激活默认主题——装完就是一个能访问的完整博客。

### 方式二：本地挂载自备前端

想用自己的前端？把一个与主题制品包同构的目录（含 `theme.json` 与 `dist/`）通过配置挂载给 CMS 即可：

- **裸机部署**：`backend/config/config.yaml` 中设置 `themes.frontend_dir` 指向目录；
- **Docker 部署**：`./deploy.sh --frontend-dir <目录>`；
- CMS 将优先托管该目录，无需走市场安装流程，适合深度自定义与本地开发调试。

### 想写一个自己的主题？

- 主题包规范：[novablog-web/docs/theme-spec.md](https://github.com/StudyNoWeekend/novablog-web/blob/main/docs/theme-spec.md)
- 打包与发布指南：[novablog-主题包规范与发布指南.md](novablog-主题包规范与发布指南.md)

### 主题相关关键配置

| 配置项 | 说明 |
|--------|------|
| `themes.market_base_url` | 官方主题市场地址（空 = 禁用市场拉取） |
| `themes.frontend_dir` | 本地挂载的自备前端目录（优先于已安装主题） |
| `themes.public_api_base` | 跨域部署时注入主题的 API 地址（空 = 同域相对路径） |
| `themes.max_artifact_mb` | 主题制品下载大小上限 |

完整配置模板见 [`backend/config/config.example.yaml`](backend/config/config.example.yaml)。

---

## 📡 开放 API 与 Skill

博客主题与第三方集成通过一套**全部免认证的公开 API**（`/api/v1/public/*`）读取内容：文章与分类标签、评论（含提交）、旅行攻略、摄影作品集、视频、个人设备、项目经历、开源作品、音乐播放、博主信息、模块开关配置等。

我们还提供一份 **Skill 格式的接口参考文档**：把 [frontend/public/docs/novablog-api.md](https://github.com/StudyNoWeekend/novablog-cms/blob/main/frontend/public/docs/novablog-api.md) 放进 Claude Code / Cursor / ZCode 等 AI 编码助手的工作区，助手即可掌握全部公开接口的路径、参数、错误码与响应示例，帮你零门槛开发或改造主题前端。

后台另内置动态 OpenAPI 文档页面，可在线浏览与调试全部接口（含管理接口）。

---

## 📦 功能特性

### 📝 文章管理

- **双编辑器**：Markdown（ByteMD，支持 GFM 与语法高亮）与富文本（wangEditor 5）按需切换；
- **分类与标签**：多分类 + 自由标签体系；
- **状态管理**：草稿 / 已发布 / 已下架三级状态一键切换；
- **SEO**：自定义 slug、摘要、封面图、关键词；
- **阅读统计**：独立阅读计数，支持热门与随机推荐接口。

### 🖼️ 多媒体库

- 图片、视频、音频上传，大文件分片上传；
- 可配置的尺寸裁剪 / 缩放预设工作台；
- EXIF 自动提取（相机、镜头、光圈、快门、ISO、焦距）；
- 存储间文件迁移，带进度追踪与状态记录。

### 📸 摄影作品集

- 按主题 / 项目组织照片集，支持排序分类；
- 每张照片可携带完整 EXIF 拍摄参数；
- 多种封面展示效果，可选图片水印保护。

### 🎥 视频作品

- 视频管理与多对多专辑系统；
- 内置 B 站 URL 解析器，支持嵌入第三方视频。

### 🗺️ 旅行攻略

- 图文游记与结构化多日行程（itinerary）；
- 高德地图 / Google Maps 地点标记；
- 点赞与阅读统计，含热门攻略推荐。

### 🎵 音乐播放

- 曲库与自定义歌单管理；
- 粘贴 B 站链接自动解析标题、封面、时长，B 站官方外链播放器内嵌播放（不受防盗链与链接时效限制）；
- 导入网易云 / QQ 音乐 / Spotify 等平台公开歌单。

### 💼 项目经历与开源作品

- **项目经历**：通用项目履历展示——领域、担任角色、客户/组织、技术栈、起止时间，覆盖摄影、剪辑、开发等任意行业；
- **开源作品**：开源项目展示与链接管理。

### 🎒 个人设备库

- 设备信息管理（图片、品牌、描述）；
- 上传照片自动识别拍摄参数填充。

### 💬 评论系统

- 统一支持文章、作品集、视频等多内容类型评论；
- 待审核 / 已通过 / 已拒绝三级审核流程，楼中楼嵌套回复；
- IP 黑名单 + 人工审核双重反垃圾。

### 📊 数据统计

- 工作台核心指标卡片、ECharts 内容趋势图；
- PV / UV 与近期访问记录，Redis 缓存聚合结果。

### 🔒 安全与监控

- JWT Access + Refresh 双令牌，无感刷新，登出即时失效（Redis 黑名单）；
- IP 黑名单支持自动过期；
- 访问日志全量落库，基于 ip2region 离线解析 IP 归属地，按保留天数自动清理；
- CORS 白名单后台动态管理，即时生效。

### ⚙️ 系统配置

- 模块开关：动态启停各内容模块，主题据此显隐入口；
- 存储配置：后台可视化配置对象存储，支持连接测试与热切换；
- 博主资料：个人信息与 20+ 社交平台链接（内置 SVG 图标与品牌色）；
- 内置动态 OpenAPI 文档页面。

---

## 🚀 快速开始

### 前置要求

Go 1.25+ · Node.js 20+（pnpm 9+）· PostgreSQL 15+ · Redis 7+

### 本地开发

```bash
# 1. 启动后端
cd backend
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml：postgres / redis 连接信息、jwt 密钥、crypto.secret_key
go mod download
make run                     # 默认监听 :8111，首次启动自动执行数据库迁移

# 2. 启动管理后台前端
cd ../frontend
pnpm install
pnpm dev                     # 默认 :5173，/api 已代理到 :8111
```

浏览器访问 `http://localhost:5173`，自动进入 `/setup` 三步安装向导，完成后即可进入后台。

### Docker 部署

需要自备 PostgreSQL 15+ 与 Redis 7+ 实例（脚本会做连通性预检；宿主机实例填 `host.docker.internal` 或宿主机 IP）：

```bash
cd deploy
./deploy.sh --name myblog                        # 交互问答补全剩余配置
./deploy.sh --name myblog --yes \
  --db-host 10.0.0.5 --db-password '***' \
  --redis-host 10.0.0.5 --redis-password '***'   # 非交互全默认部署
```

启动后：博客入口 `http://<域名>:80`，后台入口 `http://<域名>:8080/admin/`。

---

## 🐳 部署

单镜像内置 nginx 双入口、Go 后端与管理后台前端：`ghcr.io/studynoweekend/novablog-cms`（`linux/amd64` / `linux/arm64`）。

- **`--name` 必填**：作为同机多博客的隔离键——数据目录收拢在脚本同级 `<名称>/` 下，容器与网络按名称隔离，不同博客用不同名称与端口即可同机共存；
- **外部 PostgreSQL / Redis**：内置数据库容器自动部署已移除，通过 `--db-host` / `--redis-host` 指向你的实例；
- **本地挂载自备前端**：`--frontend-dir <目录>`（传 `-` 清除，回退到后台安装的主题）；
- **宿主机 nginx 反代**：`--host-nginx` 自动生成并 reload `conf.d/<名称>.conf`，适合已有 nginx 的环境；
- **升级**：重跑 `./deploy.sh --name myblog --version v1.0.6`，挂载数据保留。

```bash
./deploy.sh --name myblog --status    # 查看容器状态
./deploy.sh --name myblog --logs      # 跟踪应用日志
./deploy.sh --name myblog --down      # 停止并移除容器（挂载数据保留）
```

| 入口 | 默认端口 | 内容 |
|------|----------|------|
| 博客入口 | `80` | 激活主题页面（或本地挂载前端）· `/preview` · `/api/v1` · `/files` |
| 后台入口 | `8080` | `/admin/` 管理后台 · `/api/v1` · `/files` |

博客入口与后台入口相互独立，可分别叠加 IP 白名单、Basic Auth 与 TLS 证书；后端本身也是完整的 HTTP 服务，不依赖 nginx 可独立运行。

数据通过挂载目录持久化（`config` / `uploads` / `themes` / `logs` 等），升级重跑 `./deploy.sh` 即可，数据保留。完整的部署参数、外部数据库、旧版迁移与裸机 nginx 复用说明见 [deploy/README.md](deploy/README.md)。

---

## 📚 设计文档

| 文档 | 内容概要 |
|------|----------|
| [novablog-主题模板皮肤系统设计文档.md](novablog-主题模板皮肤系统设计文档.md) | 主题系统三层分离架构、theme.json 清单规范、预览机制、安全设计 |
| [novablog-主题包规范与发布指南.md](novablog-主题包规范与发布指南.md) | 面向主题作者：打包规范、清单校验、Next.js 静态导出、GitHub Release 分发 |
| [novablog-内容管理系统设计文档.md](novablog-内容管理系统设计文档.md) | CMS 模块拆解、数据库 Schema、API 清单、认证与缓存策略 |
| [novablog-前端设计文档.md](novablog-前端设计文档.md) | 管理后台设计系统：配色、排版、布局、组件与页面细节 |

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
