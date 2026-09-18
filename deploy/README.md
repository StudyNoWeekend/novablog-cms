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
                   PostgreSQL（内置容器或你的实例）  Redis（内置容器或你的实例）
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
./deploy.sh            # 交互式问答（端口、PG/Redis、挂载目录）
./deploy.sh --yes      # 全部默认：内置 PG/Redis，端口 80/8080，镜像版本 latest
```

部署完成后：

- 博客入口：`http://<域名>:80/`
- 后台入口：`http://<域名>:8080/admin/`
- 首装向导：`http://<域名>:8080/admin/setup`

## 部署脚本

脚本会生成 `deploy/.env` 与 `<config-dir>/config.yaml`，再按选择拼接 compose 文件启动容器；可重复执行（幂等更新/升级）。

```bash
./deploy.sh --help      # 查看全部参数
./deploy.sh --status    # 容器状态
./deploy.sh --logs      # 跟踪应用日志
./deploy.sh --down      # 停止并移除容器（挂载目录数据保留）
```

常用参数：

| 参数 | 默认 | 说明 |
|---|---|---|
| `--version <tag>` | `latest` | 镜像版本号；不指定即 latest |
| `--image <repo>` | `ghcr.io/studynoweekend/novablog-cms` | 镜像仓库 |
| `--build` | 关 | 本地构建镜像而非拉取 |
| `--blog-port <端口>` | `80` | 博客前端端口（宿主） |
| `--admin-port <端口>` | `8080` | CMS 后台端口（宿主） |
| `--public-url <地址>` | `http://localhost:<博客端口>` | 博客对外访问地址，写入 `upload.base_url`（媒体文件 URL 前缀） |
| `--domain` / `--admin-domain` | `_` | 两个入口的 server_name |
| `--db local\|external` | `local` | PostgreSQL 部署方式 |
| `--db-host/--db-port/--db-user/--db-password/--db-name/--db-sslmode` | — | 外部 PostgreSQL 连接信息 |
| `--redis local\|external` | `local` | Redis 部署方式 |
| `--redis-host/--redis-port/--redis-password/--redis-db` | — | 外部 Redis 连接信息 |
| `--config-dir <目录>` | `./config` | config.yaml 存放目录 |
| `--uploads-dir <目录>` | `./data/uploads` | 媒体上传目录 |
| `--themes-dir <目录>` | `./data/themes` | 主题制品目录 |
| `--logs-dir <目录>` | `./data/logs` | 日志目录 |
| `--frontend-dir <目录>` | 无 | 自备博客前端目录（含 `theme.json` 与 `dist/`） |
| `--pgdata-dir` / `--redisdata-dir` | `./data/pg`、`./data/redis` | 内置数据库数据目录 |
| `--market-url <地址>` | 无 | 官方主题市场地址（首装拉取默认主题用） |
| `-y, --yes` | 关 | 全部使用默认值/已有配置，不进入问答 |

示例：

```bash
# 指定版本 + 外部 PostgreSQL/Redis + 自定义挂载目录
./deploy.sh --version v1.0.0 \
  --db external --db-host 10.0.0.5 --db-user novablog --db-password '***' --db-name novablog \
  --redis external --redis-host 10.0.0.6 --redis-password '***' \
  --blog-port 80 --admin-port 8080 \
  --config-dir /srv/novablog/config --uploads-dir /srv/novablog/uploads --themes-dir /srv/novablog/themes

# 使用挂载的自备博客前端（目录内需有 theme.json 与 dist/）
./deploy.sh --frontend-dir /srv/novablog/blog-frontend

# 本地构建镜像（不打 tag 发布）
./deploy.sh --build
```

> 外部数据库在容器内访问宿主机时，主机名可用 `host.docker.internal`（Linux 下需 `--add-host=host.docker.internal:host-gateway`，compose 已默认处理该场景的解析）。

## 镜像版本号与升级

- 镜像：`ghcr.io/studynoweekend/novablog-cms:<tag>`，同时提供 `linux/amd64` 与 `linux/arm64`；
- 版本标签：`v1.0.0` 等语义化版本，稳定版同时更新 `latest`；
- 不指定 `--version` 时使用 `latest`；
- 升级：`./deploy.sh --version v1.0.1`（自动拉取新镜像并重建容器，挂载目录中的数据保留）；
- 回退：`./deploy.sh --version v1.0.0`；
- 确认运行版本：`curl http://<主机>:<博客端口>/health` 返回 `{"status":"ok","version":"v1.0.0"}`。

## 手动 Compose（不使用脚本）

```bash
cp .env.example .env    # 编辑端口、挂载目录、数据库密码

# 外部 PG/Redis
docker compose up -d

# 内置 PostgreSQL / Redis（可单独或同时使用）
docker compose -f docker-compose.yml -f docker-compose.local-pg.yml up -d
docker compose -f docker-compose.yml -f docker-compose.local-redis.yml up -d

# 本地构建镜像
docker compose -f docker-compose.yml -f docker-compose.build.yml up -d --build
```

`config.yaml` 需自行准备（可先跑一次 `./deploy.sh` 生成，再按其结构修改）：容器内路径固定为 `/app/uploads`、`/app/data/themes`、`/app/logs/app.log`，`http.port` 需保持 `8111`（nginx 反代目标）。

## 挂载目录

| 宿主目录（默认） | 容器路径 | 用途 |
|---|---|---|
| `./config/config.yaml` | `/app/config/config.yaml` | 后端配置（只读挂载） |
| `./data/uploads` | `/app/uploads` | 媒体文件（`upload.dir`） |
| `./data/themes` | `/app/data/themes` | 主题制品（`themes.data_dir`） |
| `./data/logs` | `/app/logs` | 应用日志（`log.file_path`） |
| `./data/blog-frontend` | `/app/blog-frontend` | 自备博客前端（只读，`themes.frontend_dir`） |
| `./data/pg`、`./data/redis` | — | 内置数据库数据目录（需本地文件系统，勿用 NFS） |

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
- 托管语义与已安装主题完全一致：精确文件 → `{path}.html` → `{path}/index.html` → fallback 壳页面 → `404.html`；`theme.json` 与 dotfile 不对外；`_next/`、`static/` 资源下发一年强缓存；
- 由于与 API 同源（同一入口），无需注入 `theme-config.js` 即可相对路径取数。

## 日常运维

```bash
./deploy.sh --status                      # 状态
./deploy.sh --logs                        # 日志（Ctrl+C 退出）
docker stats novablog-novablog-1          # 资源占用
./deploy.sh --down                        # 停止（数据保留）
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
4. 初始化博客外观：输入官方主题市场地址 → 后端拉取默认主题制品（sha256 校验 + 安全解压 + 清单校验）→ 自动启用；也可跳过，之后在「模板」页安装。
