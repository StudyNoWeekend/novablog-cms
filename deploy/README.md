# NovaBlog 容器部署指南（自带 nginx 双端口入口 + 博客主题托管）

## 部署拓扑

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

| 入口 | 端口（默认） | 内容 |
|---|---|---|
| 博客入口 | `NOVABLOG_HTTP_PORT`（80） | `/` 激活主题页面 · `/preview` 预览 · `/api/v1` `/files` 公开数据 |
| 后台入口 | `NOVABLOG_ADMIN_HTTP_PORT`（8080） | `/admin/` 管理后台 · `/api/v1` `/files` 管理接口 · `/` 跳转后台 |

- **博客页面**由 Go 后端按激活指针（`bloggers.active_theme_id`）托管，主题切换秒级生效；
- **主题安装**：从官方市场拉取预构建 tar.gz 制品（Release），安全解压到 `themes_data` 卷；
- **首装**：在 `/setup` 向导第三步输入官方市场地址后，自动获取默认主题并启用。

## 快速开始

```bash
# 1. 准备环境变量
cd deploy
cp .env.example .env
# 修改 POSTGRES_PASSWORD / REDIS_PASSWORD / NOVABLOG_DOMAIN

# 2. 准备后端配置（以 backend/config/config.yaml 为准）
#    容器内服务名互访：postgres.host=postgres  redis.host=redis
#    主题模块（可选）：
#    themes:
#      market_base_url: "http://<官方市场地址>:8081"   # 首装拉默认主题
#      public_api_base: ""                            # 同域部署留空

# 3. 一键构建并启动
docker compose up -d --build

# 4. 完成首装
#    打开 http://<域名>:8088/admin → 跳转 /setup → 创建账号 → 第 3 步输入官方地址获取默认主题
#    访问 http://<域名>:8088/ 即可看到博客前台
```

或在 backend 目录使用 Makefile：`make docker-up` / `make docker-build`。

## 构建期参数

| 参数 | 默认 | 说明 |
|---|---|---|
| `VITE_BASE_PATH` | `/admin/` | 管理后台前端 base 路径（须与 nginx 的 `NOVABLOG_ADMIN_BASE` 一致） |

## 分离入口 / 官方代托管

博客页面与管理后台默认已是**两个独立端口**（:80 / :8080），代托管时可按入口分别治理：

```yaml
# deploy/.env
NOVABLOG_HTTP_PORT=80          # 博客入口（对外）
NOVABLOG_ADMIN_HTTP_PORT=8080  # 后台入口（可只对内网/跳板映射，或不映射改为 SSH 隧道）
NOVABLOG_ADMIN_DOMAIN=admin.example.com
```

- **访问控制**：在 `deploy/nginx/novablog.conf.template` 的后台 server 块中叠加
  `allow/deny`（IP 白名单）或 `auth_basic`（模板顶部有示例），重建 nginx 容器即可生效；
- **独立证书**：为博客域名与后台域名分别配置 443 server 块（两入口互不影响）；
- **单域名合并模式**（旧行为）：把两个端口映射到同一宿主端口不可行（server 块按端口区分），
  如确需合并，删除模板中的后台 server 块、将 `/admin` location 移回博客 server 块即可——
  静态制品无需重建（base 仍为 `/admin/`）；
- 后台 base 路径默认 `/admin/`（与构建期 `VITE_BASE_PATH` 一致）。若希望后台在独立端口
  根路径直接呈现：`--build-arg VITE_BASE_PATH=/` 重建 nginx 镜像，`NOVABLOG_ADMIN_BASE=/`，
  并删除模板中 `location = /` 跳转行（模板内已注释说明）。

## 日常运维

```bash
docker compose ps                 # 状态
docker compose logs -f backend    # 后端日志
docker compose down               # 停止（保留数据卷）
docker compose down -v            # 停止并清空数据（谨慎）
```

数据卷：`pg_data`（数据库）、`redis_data`、`themes_data`（主题制品，切换/回滚状态）、`uploads_data`（媒体）。

## 裸机复用（可选）

`deploy/nginx/novablog.conf.template` 为标准 nginx 配置模板：

1. 手动替换 `${NOVABLOG_DOMAIN}` 为实际域名后拷贝到 `/etc/nginx/conf.d/novablog.conf`；
2. 将 `upstream` 中的 `backend:8111` 改为 `127.0.0.1:8111`；
3. `/admin` 静态目录指向本地构建的 `frontend/dist`；
4. `nginx -t && systemctl reload nginx`。

## 首装流程说明

1. `GET /api/v1/public/install/status` → 未初始化时访问 `/setup`；
2. 创建博主账号；
3. **初始化博客外观**：向导输入官方地址 → 后端调 `GET /api/v1/themes/default` → 解析 GitHub Release 制品（`{目录}-{版本}.tar.gz`）→ 流式下载 + sha256 → 安全解压 → 清单校验 → 激活。失败时前端展示明确错误提示（地址不可达/未找到制品等），允许重试或跳过。
