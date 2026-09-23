# NovaBlog 容器部署指南（单镜像：nginx 入口 + Go 后端 + 管理后台前端）

## 部署拓扑

```
                    访客浏览器                    管理员浏览器
                        │ :80                          │ :8080
        ┌───────────────┴──────────────────────────────┴───────────────┐
        │  novablog 容器（单镜像，supervisord 管理两个进程）              │
        │  ┌──────────────────────────────────────────────────────────┐ │
        │  │ nginx   :80 博客入口   / :8080 后台入口                    │ │
        │  │   /admin/ → 管理后台静态资源（镜像内置 dist）              │ │
        │  │   其余     → 反代 127.0.0.1:8111                          │ │
        │  └───────────────────────────┬──────────────────────────────┘ │
        │  ┌───────────────────────────┴──────────────────────────────┐ │
        │  │ Go 后端 :8111                                             │ │
        │  │   业务 API + 博客前端托管（挂载的自备前端 或 激活主题）    │ │
        │  └───────────────────────────┬──────────────────────────────┘ │
        └──────────────────────────────┼────────────────────────────────┘
                        ┌──────────────┴──────────────┐
                   PostgreSQL（外部实例）           Redis（外部实例）
```

| 入口 | 端口（默认） | 内容 |
|---|---|---|
| 博客入口 | `NOVABLOG_HTTP_PORT`（80） | 自备前端挂载目录 或 激活主题页面 · `/preview` 预览 · `/api/v1` `/files` 公开数据 |
| 后台入口 | `NOVABLOG_ADMIN_HTTP_PORT`（8080） | `/admin/` 管理后台 · `/api/v1` `/files` 管理接口 · `/` 跳转后台 |

- **博客页面**由 Go 后端托管：配置了 `themes.frontend_dir` 时托管挂载目录，否则按激活指针（`bloggers.active_theme_id`）托管已安装主题，切换秒级生效；
- **首装**：打开后台入口 → 跳转 `/admin/setup` → 创建博主账号 → 配置存储 → 第 3 步输入官方主题市场地址获取默认主题。

## 快速开始

```bash
cd deploy
./deploy.sh                          # 交互管理菜单：全新安装 / 容器升级 / 退出
./deploy.sh --name myblog            # 按参数安装：问答补全（端口、外部 PG/Redis 等，回车沿用已有配置）
./deploy.sh --name myblog --yes      # 全默认安装：须显式 --db-host / --redis-host（内置 PG/Redis 已移除）
```

部署完成后：

- 博客入口：`http://<博客入口域名>/`（未配域名时 `http://<主机>:80/`）
- 后台入口：`http://<后台入口域名>/admin/`
- 首装向导：`http://<后台入口域名>/admin/setup`

## 部署脚本

脚本按**博客名称**隔离部署：生成 `<名称>/.env` 与 `<名称>/config/config.yaml`，全部挂载数据默认收拢在脚本同级 `<名称>/` 目录下，再按选择拼接 compose 文件启动容器（容器前缀 `<名称>-`）。可重复执行（幂等更新/升级）。

**交互管理菜单（零参数运行）**：进入时自动获取最新发布版本号，并扫描本地部署记录与实际容器状态对齐展示（记录版本 vs 运行版本、端口、DB/Redis 模式；不对齐情况显式标警），然后可选：

1. **全新安装**——重名时自动识别：可转为升级（沿用现有配置）、或重新安装（保留数据 / 删除数据重建，后者需输入名称二次确认）；
2. **容器升级**——单个（可自定目标版本）或批量（统一升到最新版）；升级前在线对比目标版本的 `config.example.yaml`，新增配置逐条让你确认填写（默认取 example 值）、废弃配置提示保留，补写前自动备份 `config.yaml.old`；
3. 退出。

按参数运行（`./deploy.sh --name ...`）保持原有问答部署流程，适合脚本化与改参数重部署。

```bash
./deploy.sh --name llcms --help      # 查看全部参数
./deploy.sh --name llcms --status    # 容器状态（仅一个博客时可省略 --name）
./deploy.sh --name llcms --logs      # 跟踪应用日志
./deploy.sh --name llcms --down      # 停止并移除容器（挂载目录数据保留）
```

常用参数：

| 参数 | 默认 | 说明 |
|---|---|---|
| `--name <名称>` | **必填** | 博客名称 = 隔离键：数据目录 `<名称>/`、容器前缀 `<名称>-`、网络 `<名称>_novablog`；小写字母/数字/`-`/`_` |
| `--version <tag>` | `latest` | 镜像版本号；不指定即 latest |
| `--image <repo>` | `ghcr.io/studynoweekend/novablog-cms` | 镜像仓库 |
| `--build` | 关 | 本地构建镜像而非拉取 |
| `--blog-port <端口>` | `80` | 博客前端端口（宿主）；多博客同机时各用不同端口 |
| `--admin-port <端口>` | `8080` | CMS 后台端口（宿主） |
| `--public-url <地址>` | `http://<博客入口域名>`（未配域名时 `http://localhost:<博客端口>`） | 博客对外访问地址，写入 `upload.base_url`（媒体文件 URL 前缀） |
| `--domain` / `--admin-domain` | `_` | 两个入口的 server_name |
| `--db-host/--db-port/--db-user/--db-password/--db-name/--db-sslmode` | — | 外部 PostgreSQL 地址等信息，**必填**（容器网络内勿用 localhost，宿主机实例可填 `host.docker.internal` 或宿主机 IP）；用户/库名/sslmode 默认 `postgres`/`<博客名>`/`disable` |
| `--redis-host/--redis-port/--redis-password/--redis-db` | — | 外部 Redis 地址等信息，**必填**（同上，容器网络内勿用 localhost） |
| `--config-dir <目录>` | `./<名称>/config` | config.yaml 存放目录 |
| `--uploads-dir <目录>` | `./<名称>/data/uploads` | 媒体上传目录 |
| `--themes-dir <目录>` | `./<名称>/data/themes` | 主题制品目录 |
| `--logs-dir <目录>` | `./<名称>/data/logs` | 日志目录 |
| `--frontend-dir <目录>` | 无 | 自备博客前端目录（含 `theme.json` 与 `dist/`），传 `-` 清除 |
| `--pgdata-dir` / `--redisdata-dir` | `./<名称>/data/pg`、`./<名称>/data/redis` | 内置数据库数据目录（仅旧版内置模式存量部署使用，新安装不再创建） |
| `--market-url <地址>` | 无 | 官方主题市场地址（首装拉取默认主题用） |
| `-y, --yes` | 关 | 全部使用默认值/已有配置，不进入问答 |

示例：

```bash
# 指定版本 + 外部 PostgreSQL/Redis（库名默认与博客名称相同）
./deploy.sh --name llcms --version v1.0.1 \
  --db-host 10.0.0.5 --db-password '***' \
  --redis-host 10.0.0.6 --redis-password '***' --redis-db 4 \
  --blog-port 9001 --admin-port 9002 \
  --domain blog.example.com --admin-domain cms.example.com

# 使用挂载的自备博客前端（目录内需有 theme.json 与 dist/）
./deploy.sh --name llcms --frontend-dir /srv/novablog/blog-frontend

# 本地构建镜像（不打 tag 发布）
./deploy.sh --name llcms --build
```

> 外部数据库在容器内访问宿主机时，主机名可用 `host.docker.internal`（Linux 下需 `--add-host=host.docker.internal:host-gateway`，compose 已默认处理该场景的解析）。

## 单独脚本部署（免克隆仓库）

`deploy.sh` 支持脱离仓库单独运行：只把脚本拷到服务器任意目录（或用一行命令下载），执行时会**自动检测并在线拉取缺失的 compose 文件**（`docker-compose.yml` 及旧版内置 PG/Redis 的存量 overlay 等）。已存在的文件不会被覆盖。

```bash
mkdir -p /srv/novablog && cd /srv/novablog
curl -fsSL https://raw.githubusercontent.com/StudyNoWeekend/novablog-cms/main/deploy/deploy.sh -o deploy.sh
chmod +x deploy.sh

./deploy.sh --name llcms --version v1.0.1   # 首次运行自动补齐 compose 文件并部署
./deploy.sh --name llcms --status | --logs | --down
```

说明：

- **拉取源**：默认依次尝试 `raw.githubusercontent.com` 与 `cdn.jsdelivr.net`（国内网络友好）；指定 `--version vX.Y.Z` 时按同名 tag 拉取，保证 compose 与镜像版本一致。
- **镜像源覆盖**：内网/自建源可导出环境变量 `NOVABLOG_RAW_BASE`（URL 前缀，不含文件名，如 `https://mirror.example.com/novablog/main/deploy`），设置后仅使用该源。
- **`--build` 例外**：本地构建镜像需要完整仓库（`frontend/`、`backend/` 等），脚本会检测并在缺失时明确提示，请 `git clone` 后在仓库 `deploy/` 目录内运行。

## 镜像版本号与升级

- 镜像：`ghcr.io/studynoweekend/novablog-cms:<tag>`，同时提供 `linux/amd64` 与 `linux/arm64`；
- 版本标签：`v1.0.0` 等语义化版本，稳定版同时更新 `latest`；
- 不指定 `--version` 时使用 `latest`；
- 升级（推荐）：`./deploy.sh` → 菜单选「容器升级」——自动获取最新版本号，扫描本地部署并展示记录版本与实际运行版本的对齐情况，支持单个（可自定目标版本）与批量升级；升级前在线对比目标版本 `config.example.yaml`，新增配置逐条确认、废弃配置提示，补写前备份 `config.yaml.old`；容器自动拉取新镜像重建，挂载数据保留；
- 升级（按参数）：`./deploy.sh --name llcms --version v1.0.1`（问答沿用现有配置，自动拉取新镜像并重建容器）；
- 回退：`./deploy.sh --name llcms --version v1.0.0`；
- 确认运行版本：`curl http://<主机>:<博客端口>/health` 返回 `{"status":"ok","version":"v1.0.1"}`。

## 外层 nginx 反代部署（免端口访问）

镜像**自带容器内 nginx**（统一入口 + 静态托管），宿主机无需为 NovaBlog 额外安装 nginx。若宿主机已有 nginx（如承载主域名），可将二级域名反代到映射端口，实现 80/443 免端口访问。

**① deploy.sh 填法（反代场景）**

```
博客前端端口: 9001
CMS 后台端口: 9002
博客对外访问地址: http://blog.example.com   ← 访客实际访问地址，走反代时不带端口！
博客入口域名: blog.example.com              ← 裸域名，不带 http://
后台入口域名: cms.example.com               ← 裸域名，不带 http://
```

> 「博客对外访问地址」写入 `upload.base_url`，决定媒体文件 URL 形态：直接带端口访问时须填 `http://域名:端口`；走反代时填反代后的对外地址。域名值渲染进 nginx `server_name`，必须为裸域名，带 `http://` 会使域名匹配失效。

**② 宿主机 nginx 反代配置**

完整示例见 [`nginx/host-proxy.example.conf`](nginx/host-proxy.example.conf)，核心两块：

```nginx
# WebSocket 升级映射（conf.d 在 http 上下文加载，顶层 map 即可）
map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

server {
    listen 80;
    server_name blog.example.com;
    client_max_body_size 200m;        # 必设！默认 1m 会导致媒体上传/主题安装 413
    location / {
        proxy_pass http://127.0.0.1:9001;
        proxy_http_version 1.1;       # 默认 1.0 回源，SSE/流式与 keepalive 会异常
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;            # WebSocket/SSE 必留
        proxy_set_header Connection $connection_upgrade;
        proxy_buffering off;
    }
}

server {
    listen 80;
    server_name cms.example.com;
    client_max_body_size 200m;
    location / {
        proxy_pass http://127.0.0.1:9002;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
    }
}
```

修改后 `nginx -t && nginx -s reload` 生效。

> **排错：反代返回 502？** 502 表示容器内 nginx 收到了请求、但 Go 后端（127.0.0.1:8111）没有响应，与外层反代配置无关——先执行 `./deploy.sh --logs` 查看后端崩溃原因（典型如外部数据库/Redis 地址填错导致启动失败）。后端恢复后，未安装博客时博客入口会自动显示「博客即将上线」引导页。

**③ DNS 与安全组**

- 两个子域名 A 记录解析到本机 IP；
- 映射端口（9001/9002）无需对公网开放——反代仅走本机回环，安全组只保留 80/443。

**④ 后续启用 HTTPS**

证书配在宿主机 nginx 的 443 server 块（反代写法同上），生效后重跑 `./deploy.sh --version <tag>`，将「博客对外访问地址」改为 `https://blog.example.com`（媒体 URL 随协议切换，已有数据不受影响）。

## 独立 API 域名部署（可选）

默认拓扑下主题页面与 public 接口同域（博客入口域名即 API 域名，零配置）。如需把公开接口拆到**独立域名**（如 `api.example.com`，只暴露公开接口与媒体文件、不暴露后台），按以下步骤：

1. **反代**：在宿主机 nginx 增加 API 域名 server 块——只放行 `/api/v1/public/`、`/files/`（可选 `/health`、`/ready`），其余 404。完整示例见 [`nginx/host-proxy.example.conf`](nginx/host-proxy.example.conf) 第三个 server 块。
   > 切勿对该域名 `location /` 全量反代：Go 端会把未匹配路径兜底为主题静态页，等于把整站（含后台页面跳转）暴露在 API 域名下。
2. **设置 API 地址**：管理后台「主题 → 服务地址设置」填入 `http://api.example.com`（可带或不带 `/api/v1` 后缀，主题侧自动归一化）并保存。
   - 保存后后端立即对**全部已安装主题**重写 `dist/theme-config.js`（注入 `window.__NOVA_CONFIG__.apiBase`），主题前台随即走新域名取数，**无需重装主题**；
   - 配置持久化到数据库，重启不丢失；填空则清除注入，主题回退同域相对路径取数；
   - 主题安装/更新时也会按当前配置自动注入。
3. **验证**：浏览器访问 `http://api.example.com/api/v1/public/config` 应返回配置 JSON；查看主题目录 `data/themes/<主题>/dist/theme-config.js` 内容为注入后的 apiBase。
4. **CORS**：博客页面跨域调用 API 域名，需博客域名在后端白名单内——deploy.sh 生成的 `cors.allowed_origins` 已包含「博客对外访问地址」，通常无需额外操作；特殊拓扑可在后台「跨域配置」页自行调整。

## 多博客同机部署（按名称隔离）

同一台机器部署多个互相独立的博客：`--name` 是隔离键，每个博客一个专属目录、一组专属容器，互不影响。

```bash
# 第一个博客
./deploy.sh --name llcms --blog-port 9001 --admin-port 9002 \
  --domain ll.example.com --admin-domain llcms.example.com

# 第二个博客（换名称、换端口即可；PG/Redis 可共用，库名默认=博客名称、Redis 库编号错开）
./deploy.sh --name zhang3 --blog-port 9003 --admin-port 9004 \
  --domain z3.example.com --admin-domain z3cms.example.com
```

目录布局（全部默认收拢在 `<名称>/` 下，无需逐项指定）：

```
deploy.sh 所在目录/
├── deploy.sh                  # 共享脚本
├── docker-compose*.yml        # 共享模板（运行时 -p 隔离，升级只改一份）
├── llcms/                     # 博客 llcms 的专属目录
│   ├── .env
│   ├── config/config.yaml
│   └── data/{uploads,themes,logs,blog-frontend,pg,redis}
└── zhang3/                    # 博客 zhang3 的专属目录（结构同上）
```

隔离机制一览：

| 维度 | 隔离方式 |
|---|---|
| 数据 | 各博客独立目录 `<名称>/`（bind mount 互不可见） |
| 容器/网络 | compose 项目名 `-p <名称>`：容器 `<名称>-novablog-1`、网络 `<名称>_novablog` |
| 端口 | 部署前预检宿主端口占用，冲突时点名占用容器并拒绝部署（保护在运行的博客） |
| PostgreSQL | 共用同一实例时默认按博客名称分库（`--db-name` 可覆盖） |
| Redis | 共用同一实例时请为每个博客分配不同库编号（`--redis-db`） |
| 域名/对外地址 | 各博客独立填写，反代场景各配一个 server 块（见「外层 nginx 反代部署」） |

日常运维均带 `--name` 定位博客：`--status` / `--logs` / `--down` / `--version <tag>`；仅存在一个博客部署时可省略 `--name`，存在多个时脚本会列出候选。

## 手动 Compose（不使用脚本）

```bash
cp .env.example .env    # 编辑端口、挂载目录、数据库密码

# 外部 PG/Redis（推荐，deploy.sh 也仅支持外部实例）
docker compose up -d

# 内置 PostgreSQL / Redis（仅旧版存量部署升级时由脚本按 .env 记录继续带起；手动使用需自行维护）
docker compose -f docker-compose.yml -f docker-compose.local-pg.yml up -d
docker compose -f docker-compose.yml -f docker-compose.local-redis.yml up -d

# 本地构建镜像
docker compose -f docker-compose.yml -f docker-compose.build.yml up -d --build
```

`config.yaml` 需自行准备（可先跑一次 `./deploy.sh` 生成，再按其结构修改）：容器内路径固定为 `/app/uploads`、`/app/data/themes`、`/app/logs/app.log`，`http.port` 需保持 `8111`（nginx 反代目标）。

## 挂载目录

默认全部收拢在博客目录 `<名称>/` 下（可用 `--config-dir` 等 flag 覆盖）：

| 宿主目录（默认） | 容器路径 | 用途 |
|---|---|---|
| `./<名称>/config/config.yaml` | `/app/config/config.yaml` | 后端配置（只读挂载） |
| `./<名称>/data/uploads` | `/app/uploads` | 媒体文件（`upload.dir`） |
| `./<名称>/data/themes` | `/app/data/themes` | 主题制品（`themes.data_dir`） |
| `./<名称>/data/logs` | `/app/logs` | 应用日志（`log.file_path`） |
| `./<名称>/data/blog-frontend` | `/app/blog-frontend` | 自备博客前端（只读，`themes.frontend_dir`） |
| `./<名称>/data/pg`、`./<名称>/data/redis` | — | 内置数据库数据目录（需本地文件系统，勿用 NFS） |

## 挂载自备博客前端

目录结构与主题制品包一致（与 [novablog-web](https://github.com/StudyNoWeekend/novablog-web) 的 Release 制品相同）：

```
blog-frontend/
├── theme.json        # 主题清单（必需；缺失时按无壳页面模式托管）
└── dist/             # 构建产物（必需，含 index.html）
    ├── index.html
    ├── article.html  # 动态路由壳页面（可选，由 theme.json 的 routes.fallback 指定）
    └── _next/…
```

```bash
./deploy.sh --frontend-dir /srv/novablog/blog-frontend
```

- 该目录优先于后台安装的主题；目录为空或缺少 `dist/index.html` 时自动回退到已安装主题（日志中会给出提示）；
- 取消挂载：`./deploy.sh --frontend-dir -`（或在交互问答中对该项输入 `-`），即改回使用后台安装的主题；
- 托管语义与已安装主题完全一致：精确文件 → `{path}.html` → `{path}/index.html` → fallback 壳页面 → `404.html`；`theme.json` 与 dotfile 不对外；`_next/`、`static/` 资源下发一年强缓存；
- 由于与 API 同源（同一入口），无需注入 `theme-config.js` 即可相对路径取数。

## 日常运维

```bash
./deploy.sh --name llcms --status         # 状态（仅一个博客时可省略 --name）
./deploy.sh --name llcms --logs           # 日志（Ctrl+C 退出）
docker stats llcms-novablog-1             # 资源占用（容器前缀 = 博客名称）
./deploy.sh --name llcms --down           # 停止（数据保留）
```

## 从旧版（backend + nginx 双镜像）迁移

旧部署数据在命名卷中，按以下步骤迁移：

```bash
# 1) 导出（旧容器运行时执行）
docker exec novablog-postgres-1 pg_dump -U postgres novablog > novablog.sql
docker run --rm -v novablog_uploads_data:/from -v /srv/novablog/uploads:/to alpine sh -c 'cp -a /from/. /to/'
docker run --rm -v novablog_themes_data:/from  -v /srv/novablog/themes:/to  alpine sh -c 'cp -a /from/. /to/'

# 2) 停掉旧容器（数据卷先保留）
docker stop novablog-backend-1 novablog-nginx-1 novablog-postgres-1 novablog-redis-1
docker rm   novablog-backend-1 novablog-nginx-1 novablog-postgres-1 novablog-redis-1

# 3) 新方式部署，指向备份目录
./deploy.sh --uploads-dir /srv/novablog/uploads --themes-dir /srv/novablog/themes

# 4) 导入旧数据
docker exec -i novablog-postgres-1 psql -U postgres -d novablog < novablog.sql
```

## 裸机部署（不使用容器）

1. 后端：`cd backend && cp config/config.example.yaml config/config.yaml`，按需修改后 `make run`（默认 `:8111`，自动建库建表）；
2. 管理后台：`cd frontend && pnpm install && pnpm build`，将 `dist/` 交给任意静态服务器，base 路径需与访问路径一致（默认 `/admin/`）；
3. nginx：参考 `deploy/nginx/novablog.conf.template`，替换占位符、upstream 指向 `127.0.0.1:8111`、`/admin` 指向后台 `dist/`，然后 `nginx -t && systemctl reload nginx`。

## 首装流程

1. 访问后台入口 → 自动跳转 `/admin/setup`；
2. 创建博主账号（系统无默认账号，首装向导创建第一个账号）；
3. 配置存储（本地或对象存储）；
4. 初始化博客外观：输入官方主题市场地址 → 后端拉取默认主题制品（sha256 校验 + 安全解压 + 清单校验）→ 自动启用；也可跳过，之后在「主题」页安装。
