# NovaBlog CMS

> 面向独立博主的全功能内容管理系统（CMS）— Go + Gin + PostgreSQL + Redis + Vue 3

<p>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Vue-3-4FC08D?style=flat-square&logo=vue.js&logoColor=white" alt="Vue 3" />
  <img src="https://img.shields.io/badge/TypeScript-5-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Redis-7+-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
</p>

---

## 📖 项目简介

**NovaBlog** 是一套面向**独立博主**的全功能内容管理系统。它由两部分组成：

- **Go 后端服务**：提供 RESTful API、博客前端的静态资源托管、主题安装与切换、对象存储抽象、安全防护等能力；
- **Vue 3 管理后台**：博主日常使用的工作台，涵盖内容创作、多媒体管理、作品集维护、评论审核、系统配置等全部管理操作。

NovaBlog 不仅是一套内容管理后台，同时也是**博客前端的运行宿主**——已安装的主题（前端页面）由后端直接托管，主题切换秒级生效，无需重新构建或部署前端项目。

### 系统定位

整个博客体系拆分为三个独立项目，职责边界清晰：

```
┌─────────────────────────────────────────────────────────────────┐
│                   cus-official（官方网站）                        │
│  • 系统介绍 / 部署文档 / GitHub 链接 / 博主案例展示                 │
│  • 纯展示型，无业务数据交互                                        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                ★ NovaBlog（内容管理系统 / 本项目）★                │
│  • 博主登录后台 → 管理文章 / 作品 / 媒体 / 评论 / 主题              │
│  • 提供 RESTful API 供博客前端调用                                 │
│  • 核心数据中心，负责数据持久化与业务规则                            │
│  • 托管已激活的主题页面（博客前端运行宿主）                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   cus-blog-web（博客展示页 / 主题）                │
│  • 面向访客的博客前端，调用 CMS API 获取数据                        │
│  • 文章阅读 / 作品集浏览 / 视频播放 / 评论互动                      │
│  • 模板渲染 / 响应式设计 / SEO 优化                                │
└─────────────────────────────────────────────────────────────────┘
```

> **设计原则**：NovaBlog 作为唯一的**数据与业务核心**，博客前端与官方网站均为**消费端**，不直接操作数据库，所有数据通过 CMS API 获取。

---

## 🏗️ 架构概览

### 部署拓扑

```
              访客浏览器                       管理员浏览器
                  │ :80                            │ :8080
              ┌───┴────────────────────────────────┴───┐
              │  nginx 容器（双入口 server 块）          │
              │  :80   博客入口                          │
              │  :8080 后台入口（可独立加白名单/证书）     │
              └───┬────────────────────────────────┬───┘
   / /preview /api /files（博客侧）│                   │ /admin 静态 + /api /files（后台侧）
                ┌──────────────┴──────────────────┴──┐
                │  backend 容器（单端口 :8111）        │
                │  Go：业务 API + 激活主题静态托管 + 首装 │
                └───┬───────────────────────────┬────┘
              ┌─────┴───┐                 ┌─────┴────┐
              │postgres │                 │  redis   │
              └─────────┘                 └──────────┘
```

- **博客页面**由 Go 后端按激活指针（`bloggers.active_theme_id`）直接托管，主题切换秒级生效；
- **双入口分离**：博客入口与后台入口使用独立端口，可分别配置访问控制（IP 白名单 / Basic Auth）与 TLS 证书；
- **单端口裸机部署**：后端本身是一个完整的 HTTP 服务，不依赖 nginx 也能独立运行。

### 技术栈

| 层级 | 技术选型 | 说明 |
|------|----------|------|
| **后端语言** | Go 1.25+ | 高性能、并发友好、单二进制部署 |
| **Web 框架** | Gin v1.10 | 轻量级 HTTP 框架，中间件生态完善 |
| **ORM** | GORM v2 | 全功能 ORM，支持钩子与自动迁移 |
| **主数据库** | PostgreSQL 15+ | 支持 JSONB、全文检索 |
| **缓存** | Redis 7+ | 热点数据缓存、令牌黑名单、限流 |
| **对象存储** | 本地文件系统 / MinIO / 阿里云 OSS / 腾讯云 COS | 可插拔存储抽象层，支持热切换与文件迁移 |
| **配置管理** | Viper + YAML | 支持环境隔离 |
| **日志** | zap + lumberjack | 结构化日志，自动轮转 |
| **认证** | JWT（golang-jwt/jwt v5） | Access + Refresh 双令牌机制 |
| **加密** | AES-256-GCM / bcrypt | 敏感配置加密存储、密码哈希 |
| **前端框架** | Vue 3 + TypeScript + Vite 8 | 组合式 API + `<script setup>` |
| **UI 组件库** | Ant Design Vue 4 | 企业级 UI 组件 |
| **状态管理** | Pinia 3 | Vue 3 官方状态管理方案 |
| **样式方案** | UnoCSS | 原子化 CSS 引擎 |
| **富文本编辑** | wangEditor 5 + ByteMD | 富文本 / Markdown 双编辑器 |
| **数据可视化** | ECharts 6 | 工作台统计图表 |
| **地图集成** | 高德地图 / Google Maps | 旅行攻略地点标记 |
| **图片处理** | exifr / piexifjs | EXIF 元数据解析 |

---

## ✨ 功能特性

### 📝 文章管理

- **双编辑器支持**：Markdown（ByteMD，支持 GFM 与语法高亮）与富文本（wangEditor 5），按需切换；
- **分类与标签**：多分类 + 自由标签体系，灵活组织内容；
- **状态管理**：草稿 / 已发布 / 已下架三级状态，支持一键切换；
- **SEO 优化**：自定义 slug、摘要、封面图、关键词；
- **阅读统计**：独立阅读计数接口，支持热门文章、随机文章推荐；
- **公开 API**：为博客前端提供文章列表、详情、热门、随机等查询接口。

### 🖼️ 多媒体库

- **文件上传**：支持图片、视频、音频，含分片上传机制，适配大文件场景；
- **资源浏览**：网格预览、类型筛选、关键词搜索；
- **媒体预设**：可配置的尺寸裁剪 / 缩放预设工作台；
- **EXIF 提取**：自动读取照片拍摄参数（相机、镜头、光圈、快门、ISO、焦距）；
- **存储迁移**：支持在不同存储提供商之间迁移已有文件，带进度追踪与状态记录。

### 📸 摄影作品集

- **作品集管理**：按主题 / 项目组织照片集，支持排序与分类；
- **EXIF 元数据展示**：每张照片可携带完整拍摄参数；
- **封面模式**：多种封面展示效果可选；
- **水印保护**：可选的图片水印处理。

### 🎥 视频作品

- **视频管理**：上传与管理视频作品，支持封面、简介、标签；
- **视频专辑**：多对多关联的专辑系统，一个视频可属于多个专辑；
- **Bilibili 集成**：内置 B 站视频 URL 解析器，支持嵌入第三方视频内容。

### 🎒 个人设备库

- **设备管理**：记录个人设备信息，含图片、品牌与描述；
- **EXIF 自动填充**：上传照片后自动识别拍摄参数并填充，减少手工录入。

### 🗺️ 旅行攻略

- **攻略创作**：撰写图文并茂的旅行游记与攻略；
- **地图集成**：支持高德地图与 Google Maps，可添加地点标记；
- **行程规划**：结构化的多日行程（itinerary）数据模型；
- **景点管理**：关联景点、坐标、图文描述；
- **访客互动**：点赞计数与阅读统计，含热门攻略推荐接口。

### 🎵 音乐播放列表

- **曲库管理**：歌曲信息维护，支持音频文件上传与元数据（封面、歌手、专辑）；
- **第三方歌单**：导入网易云音乐、QQ 音乐、Spotify 等平台的公开歌单；
- **受控音频流**：音频 URL 带权限校验，防止原创资源被直接盗链；
- **播放列表**：自定义歌单与排序。

### 💬 评论系统

- **多目标评论**：统一支持文章、作品集、视频等不同内容类型的评论；
- **审核流程**：待审核 / 已通过 / 已拒绝三级状态，前台仅展示已通过评论；
- **嵌套回复**：支持楼中楼式回复互动；
- **反垃圾**：IP 黑名单 + 人工审核双重保障。

### 🎨 主题市场系统

这是 NovaBlog 最具特色的模块，让博客前端成为「可插拔」的资产：

- **主题市场**：集成官方主题市场，浏览、搜索、筛选、评分、收藏社区主题；
- **一键安装**：从官方市场或 GitHub Release 拉取预构建 tar.gz 制品 → SHA256 完整性校验 → 安全解压 → 清单校验 → 落盘；
- **秒级切换**：激活主题仅更新一个指针字段，无需重建镜像或重启服务；
- **预览模式**：通过 `/preview/:theme_id` 在激活前预览任意已安装主题的真实渲染效果；
- **静态托管**：Gin 的兜底路由按激活主题提供 HTML/CSS/JS/静态资源服务；
- **主题设置**：主题可通过 `theme.json` 声明配置项，后台自动生成设置面板；
- **安全机制**：解压路径穿越防护、制品大小上限、清单 schema 严格校验；
- **版本管理**：保留历史版本制品，支持回滚。

### 🔒 安全与监控

- **JWT 双令牌**：Access Token（短期）+ Refresh Token（长期），支持无感刷新；
- **令牌黑名单**：登出后令牌立即失效（Redis 存储，带 TTL）；
- **IP 黑名单**：可配置的 IP 封禁，支持自动过期；
- **访问日志**：所有请求落库（含 IP 归属地区，基于 ip2region 离线库解析，不依赖外部 API），支持按保留天数自动清理（默认 7 天，每小时检查一次）；
- **CORS 白名单**：三层优先级控制（环境变量 > 数据库配置 > 配置文件默认值），后台界面可动态修改；
- **请求追踪**：全链路 `X-Trace-Id`，便于问题排查；
- **优雅关闭**：接收 SIGINT/SIGTERM 后安全关闭服务器、数据库与 Redis 连接。

### 📊 数据统计

- **工作台概览**：内容总数、评论数、访问量等核心指标卡片；
- **内容趋势图**：文章发布趋势、评论增长趋势（ECharts 可视化）；
- **访问分析**：PV / UV 统计与近期访问记录；
- **Redis 缓存**：统计结果缓存，避免高频聚合查询压垮数据库。

### ⚙️ 系统配置

- **模块开关**：动态启用 / 禁用各功能模块，前端读取 `/public/module-config` 决定展示哪些入口；
- **存储配置**：后台可视化配置对象存储（本地 / MinIO / 阿里云 OSS / 腾讯云 COS），支持连接测试与热切换；
- **跨域配置**：后台界面动态管理 CORS 白名单，即时生效；
- **博主资料**：个人信息、头像、背景图、城市、标签，以及 20+ 社交平台链接（内置 SVG 图标与品牌色映射）；
- **API 文档**：内置动态 OpenAPI 文档页面，可在线浏览与调试接口。

### 🚀 首次安装向导

三步引导式初始化，零配置门槛：

1. **创建博主账号** —— 设置用户名与密码；
2. **初始化对象存储** —— 选择本地存储或配置云存储；
3. **拉取默认主题** —— 输入官方市场地址，自动下载并激活默认主题（失败可重试或跳过）。

系统自动检测初始化状态，未初始化时访问后台会重定向到 `/setup`。

---

## 🚀 快速开始

### 前置要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | 1.25+ | 后端编译运行 |
| Node.js | 20+ | 前端构建 |
| pnpm | 9+ | 前端包管理（推荐，也可用 npm/yarn） |
| PostgreSQL | 15+ | 主数据库 |
| Redis | 7+ | 缓存与令牌黑名单 |

### 方式一：本地开发

#### 1. 准备数据库

```bash
# 创建数据库
createdb novablog

# 确保 Redis 已启动
redis-server
```

#### 2. 启动后端

```bash
cd backend

# 复制配置模板
cp config/config.example.yaml config/config.yaml

# 编辑 config.yaml，填入 PostgreSQL / Redis 连接信息
# 重点修改：postgres.password、redis.password、jwt 密钥、crypto.secret_key

# 下载依赖
go mod download

# 启动服务（默认监听 :8111，首次启动自动执行数据库迁移）
make run
# 或：go run ./cmd/api
```

#### 3. 启动前端

```bash
cd frontend

# 安装依赖
pnpm install

# 启动开发服务器（默认 :5173，已配置 /api 代理到 localhost:8111）
pnpm dev
```

#### 4. 完成初始化

浏览器访问 `http://localhost:5173`，系统会自动跳转到 `/setup` 安装向导，按三步完成初始化即可进入后台。

### 方式二：Docker 一键部署（单镜像）

```bash
cd deploy

# 交互式问答：端口、PG/Redis、挂载目录
./deploy.sh

# 或全部使用默认值（内置 PG/Redis，端口 80/8080，镜像版本 latest）
./deploy.sh --yes

# 指定镜像版本与外部数据库
./deploy.sh --version v1.0.1 --db external --db-host 10.0.0.5 --db-password '***'
```

启动后：

- 博客入口：`http://<域名>:80`
- 后台入口：`http://<域名>:8080/admin/`

详细部署说明（镜像版本、挂载目录、外部数据库、入口分离、访问控制、旧版迁移）请参阅 [deploy/README.md](deploy/README.md)。

---

## 📂 项目结构

```
novablog-cms/
├── backend/                          # Go 后端服务
│   ├── bootstrap/                    # 应用引导
│   │   ├── app.go                    # App 初始化（配置加载、依赖装配）
│   │   ├── db.go                     # PostgreSQL (GORM) 初始化
│   │   ├── redis.go                  # Redis 客户端初始化
│   │   └── logger.go                 # zap 日志初始化（含轮转）
│   ├── cmd/
│   │   ├── api/main.go               # HTTP 服务主入口（含优雅关闭）
│   │   └── seed/main.go              # 数据库种子工具
│   ├── config/                       # YAML 配置
│   │   └── config.example.yaml       # 配置模板（含 themes.frontend_dir 等全部字段）
│   ├── enum/                         # 业务错误码定义
│   ├── internal/
│   │   ├── cache/                    # Redis 缓存层
│   │   │   ├── article_cache.go      # 文章缓存
│   │   │   ├── analytics_cache.go    # 统计缓存
│   │   │   ├── music_cache.go        # 音乐缓存
│   │   │   ├── security_cache.go     # 安全配置缓存
│   │   │   ├── module_config_cache.go# 模块开关缓存
│   │   │   ├── cors_config_cache.go  # 跨域配置缓存
│   │   │   └── token_cache.go        # JWT 令牌黑名单
│   │   ├── controller/               # HTTP 处理器（24 个控制器）
│   │   ├── dto/
│   │   │   ├── req/                  # 请求 DTO
│   │   │   └── res/                  # 响应 DTO
│   │   ├── logic/                    # 业务逻辑层（33+ 文件）
│   │   │   ├── auth_logic.go         # 认证逻辑
│   │   │   ├── article_logic.go      # 文章逻辑
│   │   │   ├── media_logic.go        # 媒体逻辑
│   │   │   ├── theme_logic.go        # 主题管理
│   │   │   ├── theme_artifact_logic.go # 主题包解析与校验
│   │   │   ├── theme_host_logic.go   # 主题静态托管
│   │   │   ├── theme_market_logic.go # 主题市场代理
│   │   │   ├── migration_logic.go    # 存储迁移
│   │   │   ├── security_logic.go     # 安全与 IP 黑名单
│   │   │   ├── setup_logic.go        # 首次安装流程
│   │   │   └── social_platforms.go   # 社交平台图标映射
│   │   ├── middleware/               # Gin 中间件
│   │   │   ├── auth.go               # JWT 验证 + 黑名单检查
│   │   │   ├── cors.go               # 三层优先级 CORS
│   │   │   ├── ip_blacklist.go       # IP 黑名单拦截
│   │   │   ├── access_log.go         # 访问日志落库
│   │   │   ├── logger.go             # 请求日志
│   │   │   ├── recovery.go           # panic 恢复
│   │   │   └── trace.go              # 请求追踪 ID
│   │   ├── model/                    # GORM 数据模型（24+ 模型）
│   │   ├── router/                   # 路由注册（按模块拆分，24 个文件）
│   │   └── storage/                  # 对象存储抽象层
│   │       ├── provider.go           # 存储接口定义
│   │       ├── factory.go            # 提供商工厂
│   │       ├── manager.go            # 热重载管理器
│   │       ├── local.go              # 本地文件系统
│   │       ├── aliyun.go             # 阿里云 OSS
│   │       ├── tencent.go            # 腾讯云 COS
│   │       └── minio.go              # MinIO (S3 兼容)
│   ├── migrations/                   # SQL 迁移脚本（40+ up/down 对）
│   ├── pkg/                          # 外部集成
│   │   ├── bilibili/client.go        # Bilibili 视频解析
│   │   ├── githubartifact/client.go  # GitHub Release 制品发现与下载
│   │   └── novablogapi/client.go     # 官方主题市场 HTTP 客户端
│   ├── utils/
│   │   ├── crypto/aes.go             # AES-256-GCM 加解密
│   │   ├── hash/bcrypt.go            # bcrypt 密码哈希
│   │   ├── jwt/jwt.go                # JWT 创建与解析
│   │   └── response/response.go      # 统一 JSON 响应
│   ├── scripts/                      # SQL 种子数据
│   ├── Makefile                      # 构建 / 运行 / 测试 / Docker 命令
│   └── go.mod
│
├── frontend/                         # Vue 3 管理后台
│   ├── src/
│   │   ├── views/                    # 页面组件
│   │   │   ├── setup/                # 首次安装向导
│   │   │   ├── auth/                 # 登录页
│   │   │   ├── dashboard/            # 工作台（统计 + 趋势图）
│   │   │   ├── article/              # 文章管理（Markdown / 富文本）
│   │   │   ├── media/                # 媒体库
│   │   │   ├── portfolio/            # 摄影作品集
│   │   │   ├── equipment/            # 个人设备
│   │   │   ├── video/                # 视频作品
│   │   │   ├── travel/               # 旅行攻略（含地图）
│   │   │   ├── playlist/             # 音乐播放列表
│   │   │   ├── comment/              # 评论审核
│   │   │   ├── template/             # 主题市场与已装主题管理
│   │   │   ├── security/             # 安全配置与访问监控
│   │   │   ├── profile/              # 个人资料 / 跨域 / 存储配置
│   │   │   ├── api-doc/              # API 文档
│   │   │   └── module-config/        # 模块开关
│   │   ├── router/                   # 路由定义与导航守卫
│   │   ├── stores/                   # Pinia 状态管理
│   │   ├── api/                      # Axios 封装与接口定义
│   │   ├── components/               # 通用组件
│   │   └── composables/              # 组合式函数（地图等）
│   ├── vite.config.ts                # Vite 配置（含 /api 代理）
│   ├── uno.config.ts                 # UnoCSS 配置
│   └── package.json
│
├── deploy/                           # Docker 部署配置
│   ├── Dockerfile                    # 单镜像多阶段构建（管理后台 + 后端 + nginx）
│   ├── docker-compose.yml            # 应用服务编排
│   ├── docker-compose.local-pg.yml   # 可选：内置 PostgreSQL
│   ├── docker-compose.local-redis.yml# 可选：内置 Redis
│   ├── docker-compose.build.yml      # 可选：本地构建镜像
│   ├── deploy.sh                     # 部署脚本（参数 + 交互问答）
│   ├── entrypoint.sh                 # 镜像入口（渲染 nginx 模板）
│   ├── supervisord.conf              # 容器内进程管理
│   ├── nginx/novablog.conf.template  # Nginx 双入口配置模板
│   ├── .env.example                  # 环境变量模板
│   └── README.md                     # 部署指南
│
├── novablog-前端设计文档.md             # 前端设计系统（配色 / 排版 / 组件 / 页面）
├── novablog-内容管理系统设计文档.md      # CMS 模块拆解 / 数据库 / API / 缓存策略
├── novablog-主题模板皮肤系统设计文档.md   # 主题系统三层架构 / 清单规范 / 预览机制
├── novablog-主题包规范与发布指南.md      # 面向主题作者的打包与发布规范
└── README.md                          # 本文件
```

---

## 🔧 配置说明

### 核心配置文件

后端配置位于 `backend/config/config.yaml`，完整模板见 `config.example.yaml`：

```yaml
app:
  name: backend            # 应用名称
  env: development         # 运行环境：development / production
  debug: true              # Gin 调试模式（生产环境请设为 false）

http:
  port: 8111               # HTTP 服务端口
  read_timeout: 30s        # 请求读超时
  write_timeout: 600s      # 响应写超时（放宽到 10 分钟：主题安装/更新需从 GitHub 下载制品）

log:
  level: debug             # 日志级别：debug / info / warn / error
  file_path: logs/app.log  # 日志文件路径
  max_size: 100            # 单个日志文件上限（MB）
  max_backups: 10          # 保留的历史日志文件数
  max_age: 30              # 日志保留天数

postgres:
  host: 127.0.0.1
  port: 5432
  user: postgres
  password: "<your-postgres-password>"
  dbname: novablog
  sslmode: disable
  max_open_conns: 100      # 最大连接数
  max_idle_conns: 10       # 最大空闲连接数

redis:
  host: 127.0.0.1
  port: 6379
  password: "<your-redis-password>"
  db: 3                    # Redis 库编号

upload:
  dir: ./uploads           # 本地存储目录
  base_url: http://localhost:8111  # 文件访问基础 URL（需与 http.port 一致）

jwt:
  access_secret: "<your-access-secret-key>"    # Access Token 密钥
  refresh_secret: "<your-refresh-secret-key>"  # Refresh Token 密钥
  access_expire: 2h        # Access Token 有效期
  refresh_expire: 168h     # Refresh Token 有效期（7 天）

crypto:
  secret_key: "<a-random-32-char-secret-key>"  # AES 加密密钥（用于加密存储配置中的敏感字段）

security:
  enabled: true            # 是否启用安全模块
  blacklist_ttl_minutes: 60    # IP 黑名单默认封禁时长（分钟）
  log_retention_days: 7        # 访问日志保留天数

geoip:
  data_dir: ./data/geoip   # IP 地区库目录（缺失 ip2region_v4.xdb 时自动下载）
  download_url: "https://github.com/lionsoul2014/ip2region/raw/v3.18.0/data/ip2region_v4.xdb"
  proxy_url: ""            # 地区库下载代理（空 = 直连，代理不可用自动降级直连）

themes:
  data_dir: ./data/themes  # 主题制品解压根目录
  market_base_url: ""      # 官方主题市场地址（空 = 禁用首装拉取）
  github_token: ""         # 可选 GitHub PAT（提升 Release 查询限流额度）
  public_api_base: ""      # 跨域部署时注入 theme-config.js 的 apiBase（空 = 同域相对路径）
  max_artifact_mb: 100     # 主题制品下载上限（MB）

cors:
  allowed_origins: "http://localhost:5173,http://localhost:5174"  # 首次启动的默认白名单
```

### 配置优先级说明

**CORS 白名单**按三层优先级解析（高 → 低）：

1. 环境变量 `CORS_ALLOWED_ORIGINS`
2. 数据库中持久化的配置（后台「个人资料 → 跨域配置」页面可编辑）
3. `config.yaml` 中的 `cors.allowed_origins`（仅作为首次启动的默认值）

### Docker 部署环境变量

`deploy/.env` 由 `./deploy.sh` 生成（也可从 `.env.example` 复制手改）：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `NOVABLOG_IMAGE_REPO` | `ghcr.io/studynoweekend/novablog-cms` | 镜像仓库 |
| `NOVABLOG_VERSION` | `latest` | 镜像版本号（不指定即 latest） |
| `NOVABLOG_DOMAIN` | `_` | 博客入口 server_name（任意域名/IP 用 `_`） |
| `NOVABLOG_HTTP_PORT` | `80` | 博客入口端口 |
| `NOVABLOG_ADMIN_DOMAIN` | `_` | 后台入口 server_name |
| `NOVABLOG_ADMIN_HTTP_PORT` | `8080` | 后台入口端口 |
| `NOVABLOG_CONFIG_DIR` | `./config` | config.yaml 所在目录 |
| `NOVABLOG_UPLOADS_DIR` | `./data/uploads` | 媒体文件目录 |
| `NOVABLOG_THEMES_DIR` | `./data/themes` | 主题制品目录 |
| `NOVABLOG_LOGS_DIR` | `./data/logs` | 日志目录 |
| `NOVABLOG_BLOG_FRONTEND_DIR` | `./data/blog-frontend` | 自备博客前端目录（空目录=使用后台安装的主题） |
| `POSTGRES_PASSWORD` | — | 内置数据库密码（使用外部数据库时不生效） |
| `REDIS_PASSWORD` | — | 内置 Redis 密码（使用外部实例时不生效） |

---

## 📡 API 概览

所有接口统一前缀 `/api/v1`，采用标准 JSON 响应格式。

### 公开接口（无需认证，供博客前端调用）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/public/install/status` | 系统安装状态检测 |
| POST | `/public/install/init` | 初始化博主账号 |
| POST | `/public/install/storage` | 初始化对象存储配置 |
| POST | `/public/install/theme` | 拉取并激活默认主题 |
| GET | `/public/blogger` | 博主信息（含社交链接） |
| GET | `/public/articles` | 文章列表（分页 / 分类 / 标签筛选） |
| GET | `/public/articles/hot` | 热门文章 |
| GET | `/public/articles/random` | 随机文章 |
| GET | `/public/articles/:slug` | 文章详情（按 slug） |
| POST | `/public/articles/:slug/view` | 阅读计数 +1 |
| GET | `/public/categories` | 分类列表 |
| GET | `/public/tags` | 标签列表 |
| GET | `/public/comments` | 评论列表（仅已通过） |
| POST | `/public/comments` | 提交评论 |
| GET | `/public/travels` | 旅行攻略列表 |
| GET | `/public/travels/hot` | 热门攻略 |
| GET | `/public/travels/:id` | 攻略详情 |
| POST | `/public/travels/:id/like` | 点赞攻略 |
| GET | `/public/portfolios` | 摄影作品集列表 |
| GET | `/public/portfolios/:id` | 作品集详情 |
| GET | `/public/videos` | 视频作品列表 |
| GET | `/public/videos/:id` | 视频详情 |
| GET | `/public/equipments` | 个人设备列表 |
| GET | `/public/music/songs` | 音乐曲目列表 |
| GET | `/public/music/audio-url/:song_id` | 获取受控音频流 URL |
| GET | `/public/music/playlists` | 音乐播放列表 |
| GET | `/public/playlists` | 第三方歌单 |
| GET | `/public/module-config` | 模块开关配置（前端据此隐藏入口） |

### 管理接口（需 `Authorization: Bearer <access_token>`）

| 模块 | 路径前缀 | 能力 |
|------|----------|------|
| 认证 | `/auth/*` | 登录、刷新令牌、登出、修改密码 |
| 媒体 | `/media/*` | 上传、列表、更新、删除 |
| 存储 | `/storage/*` | 存储配置 CRUD、连接测试、状态查询、文件迁移 |
| 分类 / 标签 | `/categories/*` `/tags/*` | 增删改查 |
| 文章 | `/articles/*` | 增删改查、状态切换 |
| 作品集 | `/portfolios/*` | 增删改查 |
| 视频 | `/videos/*` | 增删改查、专辑管理 |
| 个人设备 | `/equipments/*` | 增删改查 |
| 攻略 | `/travels/*` | 增删改查 |
| 音乐 | `/music/*` | 歌曲 CRUD、音频上传、歌单管理 |
| 第三方歌单 | `/playlists/*` | 歌单导入与管理 |
| 评论 | `/comments/*` | 审核、回复、删除 |
| 统计 | `/analytics/*` | 工作台概览、数据趋势 |
| 安全 | `/security/*` | 安全配置、IP 黑名单、访问日志 |
| 主题市场 | `/themes/market/*` | 浏览、搜索、详情、评分、收藏 |
| 已装主题 | `/themes/*` | 安装、激活、卸载、版本管理 |
| 个人资料 | `/profile` | 博主信息、社交链接、CORS 配置 |
| 模块开关 | `/module-config` | 模块启用 / 禁用 |
| API 文档 | `/api-doc` | 动态 OpenAPI 文档 |

### 其他端点

| 路径 | 说明 |
|------|------|
| `GET /health` | 健康检查（返回 200） |
| `GET /ready` | 就绪检查 |
| `GET /files/*` | 本地存储的媒体文件访问 |
| `GET /preview/:theme_id` | 主题预览（激活前查看真实效果） |
| `GET /*` | 博客前端页面（由激活主题托管） |

> 完整的接口参数与响应结构可在后台「API 文档」页面在线浏览，或访问 `/api/v1/api-doc` 获取 OpenAPI 规范。

---

## 🧩 主题系统详解

主题系统是 NovaBlog 的核心特色，让博客前端成为可热插拔的资产。

### 主题包结构

主题以 `tar.gz` 制品形式分发，解压后目录结构：

```
<theme-id>/
├── theme.json          # 主题清单（必需）
│   ├── id / name / version / author
│   ├── description / screenshot / preview
│   ├── settings schema（声明可配置项）
│   └── api 兼容性声明
├── index.html          # 入口页面
├── assets/             # 静态资源（CSS / JS / 图片）
└── theme-config.js     # 运行时注入的 API 地址配置（可选）
```

### 安装与激活流程

```
官方主题市场
    │
    │ 1. 查询主题详情，获取 Release 制品地址
    ▼
GitHub Release (tar.gz)
    │
    │ 2. 流式下载（受 max_artifact_mb 限制）
    ▼
SHA256 校验
    │
    │ 3. 完整性验证，拒绝被篡改的制品
    ▼
安全解压
    │
    │ 4. 路径穿越防护，拒绝 ../ 越界写入
    ▼
清单校验
    │
    │ 5. theme.json schema 验证
    ▼
落盘至 themes/data_dir/<theme-id>
    │
    │ 6. 写入主题记录（含版本、安装时间）
    ▼
激活（更新 bloggers.active_theme_id 指针）
    │
    ▼
秒级生效 —— 无需重启、无需重建镜像
```

### 安全机制

| 机制 | 说明 |
|------|------|
| **制品大小限制** | `themes.max_artifact_mb` 控制单次下载上限，防止磁盘打满 |
| **SHA256 校验** | 下载完成后校验制品哈希，防止传输篡改 |
| **路径穿越防护** | 解压时逐项校验文件路径，拒绝写入主题目录之外的路径 |
| **清单 Schema 校验** | 严格验证 `theme.json` 必需字段与类型 |
| **版本隔离** | 不同版本解压至独立目录，支持回滚 |
| **预览沙箱** | `/preview/:theme_id` 仅提供静态资源，不执行服务端代码 |

### 主题开发

面向主题作者的完整打包与发布规范，请参阅 [novablog-主题包规范与发布指南.md](novablog-主题包规范与发布指南.md)。

---

## 🐳 部署

### 单镜像 + 部署脚本（推荐）

nginx 入口、Go 后端与管理后台前端已合并在**一个镜像**内，镜像地址 `ghcr.io/studynoweekend/novablog-cms`（支持 `linux/amd64` / `linux/arm64`）：

```bash
cd deploy
./deploy.sh                 # 交互式问答：端口、PG/Redis、挂载目录
./deploy.sh --yes           # 全部默认：内置 PG/Redis，端口 80 / 8080，版本 latest
./deploy.sh --version v1.0.1 --db external --db-host 10.0.0.5 --db-password '***'
```

脚本会生成 `deploy/.env` 与 `config.yaml`，按选择拼接 compose 文件并启动容器。

| 入口 | 默认端口 | 内容 |
|------|----------|------|
| 博客入口 | `NOVABLOG_HTTP_PORT`（80） | 自备前端挂载目录 或 激活主题页面 · `/preview` · `/api/v1` · `/files` |
| 后台入口 | `NOVABLOG_ADMIN_HTTP_PORT`（8080） | `/admin/` 管理后台 · `/api/v1` · `/files` 管理接口 |

- **版本号**：`--version v1.0.0` 指定镜像版本，不指定则使用 `latest`；升级重跑 `./deploy.sh --version v1.0.1` 即可，挂载目录数据保留；
- **数据库**：`--db local|external` 与 `--redis local|external` 可分别选择内置容器或你自己的实例；
- **挂载目录**：`--config-dir`（config.yaml）、`--uploads-dir`、`--themes-dir`、`--logs-dir`、`--pgdata-dir`、`--redisdata-dir`；
- **自备博客前端**：`--frontend-dir` 指定含 `theme.json` 与 `dist/` 的目录（与主题制品包同构），未指定则使用后台安装的主题。

```bash
./deploy.sh --status   # 查看状态
./deploy.sh --logs     # 查看日志
./deploy.sh --down     # 停止容器（挂载目录数据保留）
./deploy.sh --help     # 查看全部参数
```

### 手动 Compose

```bash
cd deploy
cp .env.example .env    # 编辑端口、挂载目录、数据库密码
docker compose up -d                                                              # 使用外部 PG/Redis
docker compose -f docker-compose.yml -f docker-compose.local-pg.yml up -d          # 附带内置 PostgreSQL
docker compose -f docker-compose.yml -f docker-compose.local-redis.yml up -d       # 附带内置 Redis
```

| 服务 | 镜像 | 职责 |
|------|------|------|
| `novablog` | `ghcr.io/studynoweekend/novablog-cms:<版本>` | nginx 双端口入口 + Go API + 主题托管 + 管理后台静态资源 |
| `postgres` | `postgres:15-alpine` | 可选：内置主数据库 |
| `redis` | `redis:7-alpine` | 可选：内置缓存与令牌黑名单 |

数据默认落在 `deploy/data/` 下（`uploads` / `themes` / `logs` / `pg` / `redis`），可在 `.env` 中改为任意目录。

### 从旧版（backend + nginx 双镜像）迁移

旧部署的数据在命名卷 `novablog_pg_data` / `novablog_uploads_data` / `novablog_themes_data` 中，按以下步骤迁移到新的挂载目录：

```bash
# 1) 导出数据（旧容器仍在运行时）
docker exec novablog-postgres-1 pg_dump -U postgres novablog > novablog.sql
docker run --rm -v novablog_uploads_data:/from -v /srv/novablog/uploads:/to alpine sh -c 'cp -a /from/. /to/'
docker run --rm -v novablog_themes_data:/from  -v /srv/novablog/themes:/to  alpine sh -c 'cp -a /from/. /to/'

# 2) 停掉并移除旧容器（数据卷保留，确认无误后再删）
docker stop novablog-backend-1 novablog-nginx-1 novablog-postgres-1 novablog-redis-1
docker rm   novablog-backend-1 novablog-nginx-1 novablog-postgres-1 novablog-redis-1

# 3) 用新方式部署，指向刚备份的目录
./deploy.sh --uploads-dir /srv/novablog/uploads --themes-dir /srv/novablog/themes

# 4) 导入旧数据
docker exec -i novablog-postgres-1 psql -U postgres -d novablog < novablog.sql
```

### 入口分离与访问控制

博客入口（:80）与后台入口（:8080）默认独立：

- **访问控制**：在 `deploy/nginx/novablog.conf.template` 的后台 server 块中叠加 `allow/deny`（IP 白名单）或 `auth_basic`（模板顶部有示例）；
- **独立证书**：为博客域名与后台域名分别配置 443 server 块；
- **内网后台**：后台端口可只对内网映射，或改用 SSH 隧道访问。

### 裸机复用 nginx

`deploy/nginx/novablog.conf.template` 是标准 nginx 配置模板，可直接用于已有 nginx 环境：

1. 替换 `${NOVABLOG_DOMAIN}` / `${NOVABLOG_ADMIN_PORT}` 等占位符后拷贝到 `/etc/nginx/conf.d/novablog.conf`；
2. upstream 使用 `127.0.0.1:8111`（模板默认值，即后端 `http.port`）；
3. `/admin` 静态目录指向本地构建的 `frontend/dist`；
4. `nginx -t && systemctl reload nginx`。

> 提示：镜像内 nginx 固定反代 `127.0.0.1:8111`，因此 `config.yaml` 的 `http.port` 需保持 `8111`；如需改动，请同时设置环境变量 `NOVABLOG_BACKEND_ADDR`。

更多细节请参阅 [deploy/README.md](deploy/README.md)。

---

## 🧪 开发

### 后端命令

```bash
cd backend

make build          # 编译二进制到 bin/novablog
make run            # 直接运行
make test           # 运行测试（含 race 检测与覆盖率）
make fmt            # 格式化代码
make vet            # 静态检查
make lint           # golangci-lint 代码规范检查
make clean          # 清理编译产物
make docker-build   # 构建单镜像（nginx + 后端 + 管理后台前端）
make docker-up      # 本地构建并启动容器编排
make docker-down    # 停止容器编排
```

### 前端命令

```bash
cd frontend

pnpm dev            # 开发服务器（:5173，/api 代理到 :8111）
pnpm build          # 生产构建（输出到 dist/）
pnpm preview        # 预览生产构建
pnpm lint           # 代码检查
```

### 构建期变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `VITE_BASE_PATH` | `/admin/` | 管理后台前端 base 路径（须与 nginx 的 `NOVABLOG_ADMIN_BASE` 一致） |
| `VERSION` | `dev` | 注入二进制的版本号，启动日志与 `GET /health` 可见 |

### 数据库迁移

迁移脚本位于 `backend/migrations/`，采用 `{序号}_{描述}.up.sql` / `.down.sql` 成对命名。服务启动时自动执行未应用的迁移；也可通过 `cmd/seed` 工具手动执行种子数据。

---

## 📄 设计文档

项目根目录包含四份详细的中文设计文档，是理解系统设计意图的最佳入口：

| 文档 | 内容概要 |
|------|----------|
| [novablog-前端设计文档.md](novablog-前端设计文档.md) | 设计系统全貌：配色方案（"Creator's Darkroom" 暗房主题）、字体排版、布局系统、组件规范、逐页面设计细节 |
| [novablog-内容管理系统设计文档.md](novablog-内容管理系统设计文档.md) | CMS 模块拆解、完整数据库 Schema、50+ API 接口清单、认证流程、缓存策略、开发阶段规划 |
| [novablog-主题模板皮肤系统设计文档.md](novablog-主题模板皮肤系统设计文档.md) | 主题系统三层分离架构（规范层 / 分发层 / 消费层）、theme.json 清单规范、部署拓扑、预览机制、安全设计 |
| [novablog-主题包规范与发布指南.md](novablog-主题包规范与发布指南.md) | 面向主题作者的打包规范：tar.gz 格式要求、清单校验规则、Next.js 静态导出要求、API 地址解析策略、GitHub Release 分发流程 |

---

## 🔍 核心设计要点

理解以下几点，能快速把握 NovaBlog 的设计思路：

1. **单博主模型**：系统面向单个博主设计，不存在多用户 / 多角色体系。`bloggers` 表中只有一条记录，`active_theme_id` 字段即是整个博客的「外观开关」。

2. **后端即博客宿主**：Gin 的 `NoRoute` 兜底路由接管所有未匹配请求，按激活主题提供静态文件。这意味着部署时只需要一个后端服务 + 数据库，博客前端不需要独立部署。

3. **存储抽象层**：`internal/storage` 定义了统一的 `Provider` 接口，本地 / MinIO / 阿里云 / 腾讯云四种实现可热切换。存储管理器从数据库读取当前激活配置，配置变更后调用 `Reload()` 即可切换，无需重启。

4. **缓存分层**：Redis 承担四类职责——热点数据缓存（文章、统计、音乐）、配置缓存（安全、CORS、模块开关）、令牌黑名单、以及缓存的主动失效（写操作后清理对应 key）。

5. **配置三层优先级**：环境变量 > 数据库持久化配置 > YAML 默认值。这让「后台界面可改」和「运维强制覆盖」两种需求可以共存。

6. **主题安装的安全闭环**：下载 → 哈希校验 → 路径穿越防护解压 → 清单 schema 验证，每一步都拒绝不可信输入，因为主题制品来自第三方社区。

---

## 🤝 贡献

欢迎提交 Issue 与 Pull Request。参与开发前建议：

1. 阅读对应模块的设计文档，理解设计意图；
2. 后端代码遵循 Go 官方风格，提交前执行 `make fmt && make vet && make lint`；
3. 前端代码遵循项目 ESLint / Prettier 配置，提交前执行 `pnpm lint`；
4. 新增 API 时同步更新 DTO 与 API 文档。

---

## 📜 开源协议

本项目基于 MIT 协议开源。

---

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) — 优秀的 Go Web 框架
- [GORM](https://gorm.io) — 功能完备的 Go ORM
- [Vue.js](https://vuejs.org) — 渐进式 JavaScript 框架
- [Ant Design Vue](https://antdv.com) — 企业级 UI 组件库
- [PostgreSQL](https://www.postgresql.org) — 最先进的开源关系型数据库
- [Redis](https://redis.io) — 高性能内存数据结构存储
- 以及所有为本项目提供灵感与支持的社区贡献者

---

<p align="center">
  <sub>Made with ❤️ for independent bloggers</sub>
</p>
