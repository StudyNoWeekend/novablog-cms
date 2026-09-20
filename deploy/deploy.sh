#!/usr/bin/env bash
#
# NovaBlog 部署脚本
#   - 生成 deploy/.env 与 config.yaml，并按选择拼接 compose 文件后启动容器
#   - 可选内置 PostgreSQL / Redis，或使用你自己的外部实例
#   - 可指定博客前端端口、CMS 后台端口，以及 uploads / themes / config.yaml 等挂载目录
#   - 可指定镜像版本号（不指定则使用 latest）
#
# 示例：
#   ./deploy.sh                                  # 交互式问答
#   ./deploy.sh --yes                            # 全部默认（内置 PG/Redis，端口 80/8080，latest）
#   ./deploy.sh --version v1.0.0 --db external --db-host 10.0.0.5 --db-password '***'
#   ./deploy.sh --blog-port 80 --admin-port 8080 --frontend-dir /srv/blog-frontend
#   ./deploy.sh --build                          # 本地构建镜像（不打 tag 发布）
#   ./deploy.sh --status | --logs | --down
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
ENV_FILE="$SCRIPT_DIR/.env"

# ============================== 输出辅助 ==============================
c_info() { printf '\033[32m[信息]\033[0m %s\n' "$*"; }
c_warn() { printf '\033[33m[提示]\033[0m %s\n' "$*"; }
c_err()  { printf '\033[31m[错误]\033[0m %s\n' "$*" >&2; }
die()    { c_err "$*"; exit 1; }
# TCP 连通性探测：错误地址（如 IP 手误）可能长时间挂起，优先用 timeout 限时
tcp_reachable() {
  local host="$1" port="$2"
  if command -v timeout >/dev/null 2>&1; then
    timeout 3 bash -c "exec 3<>/dev/tcp/$host/$port" 2>/dev/null
  else
    bash -c "exec 3<>/dev/tcp/$host/$port" 2>/dev/null
  fi
}

# ============================== 默认值 ==============================
IMAGE_REPO="ghcr.io/studynoweekend/novablog-cms"
VERSION=""
BUILD=0
ASSUME_YES=0
SKIP_IMAGE_CHECK=0
ACTION="up"

BLOG_PORT=""
ADMIN_PORT=""
PUBLIC_URL=""
DOMAIN=""
ADMIN_DOMAIN=""

DB_MODE=""
DB_HOST="" DB_PORT="" DB_USER="" DB_PASSWORD="" DB_NAME="" DB_SSLMODE=""

REDIS_MODE=""
REDIS_HOST="" REDIS_PORT="" REDIS_PASSWORD="" REDIS_DB=""

CONFIG_DIR="" UPLOADS_DIR="" THEMES_DIR="" LOGS_DIR="" FRONTEND_DIR=""
PGDATA_DIR="" REDISDATA_DIR=""
MARKET_URL=""

VERSION_SET=0

usage() {
  cat <<'EOF'
NovaBlog 部署脚本

用法：./deploy.sh [选项]

镜像与版本
  --version <tag>        镜像版本号，不指定则使用 latest（例：--version v1.0.0）
  --image <repo>         镜像仓库地址（默认 ghcr.io/studynoweekend/novablog-cms）
  --build                本地构建镜像，而非拉取 registry 镜像
  --skip-image-check     跳过镜像 tag 存在性校验

入口端口
  --blog-port <端口>     博客前端端口（宿主，默认 80）
  --admin-port <端口>    CMS 后台端口（宿主，默认 8080）
  --public-url <地址>    博客对外访问地址（默认 http://localhost:<博客端口>，用于媒体文件 URL）
  --domain <域名>        博客入口 server_name（默认 _，即任意域名/ IP）
  --admin-domain <域名>  后台入口 server_name（默认 _）

数据库（PostgreSQL）
  --db-host <主机>     PostgreSQL 地址；留空或 localhost/127.0.0.1 时自动部署内置容器
  --db-port <端口>       默认 5432
  --db-user <用户>       默认 postgres
  --db-password <密码>   必填（内置模式下自动生成）
  --db-name <库名>       默认 novablog
  --db-sslmode <模式>    默认 disable

缓存（Redis）
  --redis-host <主机>    Redis 地址；留空或 localhost/127.0.0.1 时自动部署内置容器
  --redis-port <端口>    默认 6379
  --redis-password <密码> 无密码可留空（内置模式下自动生成）
  --redis-db <编号>      默认 3

挂载目录（相对路径以 deploy/ 为基准，建议用绝对路径）
  --config-dir <目录>    config.yaml 存放目录（默认 ./config）
  --uploads-dir <目录>   媒体上传目录（默认 ./data/uploads）
  --themes-dir <目录>    主题制品目录（默认 ./data/themes）
  --logs-dir <目录>      日志目录（默认 ./data/logs）
  --frontend-dir <目录>  自备博客前端目录（含 theme.json 与 dist/），不指定则使用后台安装的主题；
                         传 - 可清除已配置的挂载前端
  --pgdata-dir <目录>    内置 PostgreSQL 数据目录（默认 ./data/pg）
  --redisdata-dir <目录> 内置 Redis 数据目录（默认 ./data/redis）

其他
  --market-url <地址>    官方主题市场地址（首装拉取默认主题用，可留空）
  -y, --yes              全部使用默认值/已有配置，不进入交互问答
  --status               查看容器状态
  --logs                 跟踪应用日志
  --down                 停止并移除容器（挂载目录中的数据保留）
  -h, --help             显示本帮助
EOF
}

# ============================== 参数解析 ==============================
while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="${2:?--version 需要一个值}"; VERSION_SET=1; shift 2 ;;
    --image) IMAGE_REPO="${2:?--image 需要一个值}"; shift 2 ;;
    --build) BUILD=1; shift ;;
    --skip-image-check) SKIP_IMAGE_CHECK=1; shift ;;
    --blog-port) BLOG_PORT="${2:?--blog-port 需要一个值}"; shift 2 ;;
    --admin-port) ADMIN_PORT="${2:?--admin-port 需要一个值}"; shift 2 ;;
    --public-url) PUBLIC_URL="${2:?--public-url 需要一个值}"; shift 2 ;;
    --domain) DOMAIN="${2:?--domain 需要一个值}"; shift 2 ;;
    --admin-domain) ADMIN_DOMAIN="${2:?--admin-domain 需要一个值}"; shift 2 ;;
    --db) DB_MODE="${2:?--db 需要一个值}"; shift 2 ;;
    --db-host) DB_HOST="${2:?--db-host 需要一个值}"; shift 2 ;;
    --db-port) DB_PORT="${2:?--db-port 需要一个值}"; shift 2 ;;
    --db-user) DB_USER="${2:?--db-user 需要一个值}"; shift 2 ;;
    --db-password) DB_PASSWORD="${2:?--db-password 需要一个值}"; shift 2 ;;
    --db-name) DB_NAME="${2:?--db-name 需要一个值}"; shift 2 ;;
    --db-sslmode) DB_SSLMODE="${2:?--db-sslmode 需要一个值}"; shift 2 ;;
    --redis) REDIS_MODE="${2:?--redis 需要一个值}"; shift 2 ;;
    --redis-host) REDIS_HOST="${2:?--redis-host 需要一个值}"; shift 2 ;;
    --redis-port) REDIS_PORT="${2:?--redis-port 需要一个值}"; shift 2 ;;
    --redis-password) REDIS_PASSWORD="${2:?--redis-password 需要一个值}"; shift 2 ;;
    --redis-db) REDIS_DB="${2:?--redis-db 需要一个值}"; shift 2 ;;
    --config-dir) CONFIG_DIR="${2:?--config-dir 需要一个值}"; shift 2 ;;
    --uploads-dir) UPLOADS_DIR="${2:?--uploads-dir 需要一个值}"; shift 2 ;;
    --themes-dir) THEMES_DIR="${2:?--themes-dir 需要一个值}"; shift 2 ;;
    --logs-dir) LOGS_DIR="${2:?--logs-dir 需要一个值}"; shift 2 ;;
    --frontend-dir) FRONTEND_DIR="${2:?--frontend-dir 需要一个值}"; shift 2 ;;
    --pgdata-dir) PGDATA_DIR="${2:?--pgdata-dir 需要一个值}"; shift 2 ;;
    --redisdata-dir) REDISDATA_DIR="${2:?--redisdata-dir 需要一个值}"; shift 2 ;;
    --market-url) MARKET_URL="${2:?--market-url 需要一个值}"; shift 2 ;;
    -y|--yes) ASSUME_YES=1; shift ;;
    --status) ACTION="status"; shift ;;
    --logs) ACTION="logs"; shift ;;
    --down) ACTION="down"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "未知参数：$1（使用 --help 查看用法）" ;;
  esac
done

# ============================== 交互辅助 ==============================
INTERACTIVE=1
[ -t 0 ] || INTERACTIVE=0
[ "$ASSUME_YES" = 1 ] && INTERACTIVE=0

# env_get 从已有 .env 读取变量（不执行文件内容）
env_get() {
  [ -f "$ENV_FILE" ] || return 0
  sed -n "s/^$1=//p" "$ENV_FILE" | tail -1
}

# ask "提示" "默认值" → 结果输出到 stdout
ask() {
  local prompt="$1" def="${2:-}" ans=""
  if [ "$INTERACTIVE" = 1 ]; then
    if [ -n "$def" ]; then
      read -r -p "$prompt [$def]: " ans || true
    else
      read -r -p "$prompt: " ans || true
    fi
  fi
  printf '%s' "${ans:-$def}"
}

# ask_choice "提示" "候选1|候选2" "默认"
ask_choice() {
  local prompt="$1" opts="$2" def="$3" ans=""
  if [ "$INTERACTIVE" = 1 ]; then
    read -r -p "$prompt ($opts) [$def]: " ans || true
  fi
  ans="${ans:-$def}"
  case "|$opts|" in
    *"|$ans|"*) printf '%s' "$ans" ;;
    *) die "取值必须是 $opts 之一，收到：$ans" ;;
  esac
}

# ask_secret "提示" "默认值" → 输入不回显
ask_secret() {
  local prompt="$1" def="${2:-}" ans=""
  if [ "$INTERACTIVE" = 1 ]; then
    read -r -s -p "$prompt: " ans || true
    printf '\n' >&2
  fi
  printf '%s' "${ans:-$def}"
}

gen_secret() {
  local bytes="${1:-32}"
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex "$bytes"
  else
    head -c "$bytes" /dev/urandom | od -An -tx1 | tr -d ' \n'
  fi
}

# yaml_quote 转义双引号字符串
yaml_quote() { printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g'; }

# abs_dir 相对 deploy/ 解析并创建目录，输出绝对路径
abs_dir() {
  local p="$1"
  case "$p" in
    /*) ;;
    *) p="$SCRIPT_DIR/$p" ;;
  esac
  mkdir -p "$p" || die "无法创建目录：$p"
  (cd "$p" && pwd)
}

# compose_up_failed 启动失败时的排查提示
compose_up_failed() {
  c_err "容器启动失败。"
  cat <<EOF

常见原因：
  1) 端口被占用 —— 检查博客端口 ${BLOG_PORT} 与后台端口 ${ADMIN_PORT} 是否已被占用：
       lsof -i :${BLOG_PORT} ; lsof -i :${ADMIN_PORT}
     被占用时可改端口重试：./deploy.sh --blog-port 8080 --admin-port 8081
  2) 挂载目录不可读写 —— 确认 ${CONFIG_DIR} / ${UPLOADS_DIR} / ${THEMES_DIR} 权限正常
  3) 配置有误 —— 检查 ${CONFIG_FILE}

查看详细日志：./deploy.sh --logs
EOF
  exit 1
}

# ============================== 前置检查 ==============================
command -v docker >/dev/null 2>&1 || die "未找到 docker，请先安装 Docker"
docker info >/dev/null 2>&1 || die "无法连接 Docker 守护进程，请确认 Docker 已启动"
docker compose version >/dev/null 2>&1 || die "未找到 docker compose（需 Compose V2）"

# ============================== 依赖文件在线自举 ==============================
# 脚本支持脱离仓库单独运行：compose 依赖文件缺失时自动从 GitHub 在线拉取。
# 已存在的文件绝不覆盖（仓库克隆场景行为不变）。
# 国内网络可导出 NOVABLOG_RAW_BASE 指向自建/镜像源（URL 前缀，不含文件名）。
RAW_REF="main"
case "$VERSION" in v[0-9]*) RAW_REF="$VERSION" ;; esac

fetch_url() { # $1=url $2=目标文件
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$1" -o "$2"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$2" "$1"
  else
    return 127
  fi
}

fetch_missing_file() { # $1=文件名（相对 deploy/ 目录）
  local f="$1" dest="$SCRIPT_DIR/$1" url src
  [ -f "$dest" ] && return 0
  local sources=()
  if [ -n "${NOVABLOG_RAW_BASE:-}" ]; then
    sources=("$NOVABLOG_RAW_BASE")
  else
    sources=(
      "https://raw.githubusercontent.com/StudyNoWeekend/novablog-cms/${RAW_REF}/deploy"
      "https://cdn.jsdelivr.net/gh/StudyNoWeekend/novablog-cms@${RAW_REF}/deploy"
    )
  fi
  for src in "${sources[@]}"; do
    url="${src%/}/$f"
    c_info "缺失 ${f}，尝试在线拉取：${url}"
    if fetch_url "$url" "$dest" && [ -s "$dest" ] && grep -q '^services:' "$dest" 2>/dev/null; then
      c_info "已下载 $f"
      return 0
    fi
    rm -f "$dest"
  done
  c_err "在线拉取 $f 失败。请检查网络后重试，或手动下载放到 $dest"
  c_err "  curl -fsSL ${sources[0]%/}/$f -o $dest"
  return 1
}

# ensure_compose_files 预检 docker compose 参数中的 -f 文件，缺失则在线拉取
ensure_compose_files() {
  local arg
  for arg in "$@"; do
    [ "$arg" = "-f" ] && continue
    case "$arg" in -*) continue ;; esac
    fetch_missing_file "$arg" || exit 1
  done
}

compose_files() {
  local files=(-f docker-compose.yml)
  [ "$DB_MODE" = "local" ] && files+=(-f docker-compose.local-pg.yml)
  [ "$REDIS_MODE" = "local" ] && files+=(-f docker-compose.local-redis.yml)
  [ "$BUILD" = 1 ] && files+=(-f docker-compose.build.yml)
  printf '%s\n' "${files[@]}"
}

# 从 .env 推断文件组合（用于 --status / --logs / --down）
compose_files_from_env() {
  local db_mode redis_mode files=(-f docker-compose.yml)
  db_mode="$(env_get NOVABLOG_DB_MODE)"
  redis_mode="$(env_get NOVABLOG_REDIS_MODE)"
  [ "$db_mode" = "local" ] && files+=(-f docker-compose.local-pg.yml)
  [ "$redis_mode" = "local" ] && files+=(-f docker-compose.local-redis.yml)
  printf '%s\n' "${files[@]}"
}

run_compose() {
  local f
  local args=()
  while IFS= read -r f; do [ -n "$f" ] && args+=("$f"); done < <(compose_files)
  ensure_compose_files "${args[@]}"
  docker compose "${args[@]}" "$@"
}

run_compose_from_env() {
  local f
  local args=()
  while IFS= read -r f; do [ -n "$f" ] && args+=("$f"); done < <(compose_files_from_env)
  ensure_compose_files "${args[@]}"
  docker compose "${args[@]}" "$@"
}

case "$ACTION" in
  status)
    [ -f "$ENV_FILE" ] || die "未找到 .env，请先运行 ./deploy.sh 完成部署"
    run_compose_from_env ps
    exit 0
    ;;
  logs)
    [ -f "$ENV_FILE" ] || die "未找到 .env，请先运行 ./deploy.sh 完成部署"
    run_compose_from_env logs -f --tail=200 novablog
    exit 0
    ;;
  down)
    [ -f "$ENV_FILE" ] || die "未找到 .env，请先运行 ./deploy.sh 完成部署"
    c_info "停止并移除容器（挂载目录中的数据保留）..."
    run_compose_from_env down
    c_info "已停止。数据仍在 .env 中配置的挂载目录里，重新部署执行 ./deploy.sh 即可。"
    exit 0
    ;;
esac

# --build 需要完整仓库作为构建上下文（frontend/ backend/ 等），缺失时早失败（仅部署路径）
if [ "$BUILD" = 1 ] && { [ ! -f "$SCRIPT_DIR/Dockerfile" ] || [ ! -d "$SCRIPT_DIR/../backend" ]; }; then
  die "--build 需要完整仓库（构建上下文为仓库根，需 frontend/ 与 backend/）。请 git clone 仓库后在 deploy/ 目录内运行，或去掉 --build 直接使用线上镜像。"
fi

# ============================== 收集配置 ==============================
if [ -f "$ENV_FILE" ]; then
  c_info "检测到已有 .env，将复用其中的配置作为默认值（可直接回车沿用）。"
fi

[ -z "$IMAGE_REPO" ] && IMAGE_REPO="$(env_get NOVABLOG_IMAGE_REPO)"
[ "$VERSION_SET" = 0 ] && VERSION="$(env_get NOVABLOG_VERSION)"
[ -z "$BLOG_PORT" ] && BLOG_PORT="$(env_get NOVABLOG_HTTP_PORT)"
[ -z "$ADMIN_PORT" ] && ADMIN_PORT="$(env_get NOVABLOG_ADMIN_HTTP_PORT)"
[ -z "$DOMAIN" ] && DOMAIN="$(env_get NOVABLOG_DOMAIN)"
[ -z "$ADMIN_DOMAIN" ] && ADMIN_DOMAIN="$(env_get NOVABLOG_ADMIN_DOMAIN)"
[ -z "$CONFIG_DIR" ] && CONFIG_DIR="$(env_get NOVABLOG_CONFIG_DIR)"
[ -z "$UPLOADS_DIR" ] && UPLOADS_DIR="$(env_get NOVABLOG_UPLOADS_DIR)"
[ -z "$THEMES_DIR" ] && THEMES_DIR="$(env_get NOVABLOG_THEMES_DIR)"
[ -z "$LOGS_DIR" ] && LOGS_DIR="$(env_get NOVABLOG_LOGS_DIR)"
[ -z "$FRONTEND_DIR" ] && FRONTEND_DIR="$(env_get NOVABLOG_BLOG_FRONTEND_DIR)"
[ -z "$PGDATA_DIR" ] && PGDATA_DIR="$(env_get NOVABLOG_PGDATA_DIR)"
[ -z "$REDISDATA_DIR" ] && REDISDATA_DIR="$(env_get NOVABLOG_REDISDATA_DIR)"
[ -z "$DB_MODE" ] && DB_MODE="$(env_get NOVABLOG_DB_MODE)"
[ -z "$REDIS_MODE" ] && REDIS_MODE="$(env_get NOVABLOG_REDIS_MODE)"
[ -z "$DB_PASSWORD" ] && DB_PASSWORD="$(env_get POSTGRES_PASSWORD)"
[ -z "$REDIS_PASSWORD" ] && REDIS_PASSWORD="$(env_get REDIS_PASSWORD)"

echo
c_info "===== 入口端口 ====="
BLOG_PORT="$(ask '博客前端端口' "${BLOG_PORT:-80}")"
ADMIN_PORT="$(ask 'CMS 后台端口' "${ADMIN_PORT:-8080}")"
PUBLIC_URL="$(ask '博客对外访问地址（用于媒体文件 URL，填访客实际访问地址，带反代时不带端口）' "${PUBLIC_URL:-http://localhost:$BLOG_PORT}")"
# 未带协议时自动补 http://，否则媒体 URL 与 CORS 白名单会生成非法值
case "$PUBLIC_URL" in
  http://*|https://*) ;;
  *) PUBLIC_URL="http://$PUBLIC_URL"; c_info "博客对外访问地址已自动补全为：$PUBLIC_URL" ;;
esac
DOMAIN="$(ask '博客入口域名（裸域名如 blog.example.com，不带 http://；任意域名/IP 填 _）' "${DOMAIN:-_}")"
ADMIN_DOMAIN="$(ask '后台入口域名（裸域名，不带 http://；任意域名/IP 填 _）' "${ADMIN_DOMAIN:-_}")"
# 误带协议或结尾斜杠时自动清洗，否则会渲染进 nginx server_name 导致域名匹配失效
strip_url_prefix() { local v="$1"; v="${v#http://}"; v="${v#https://}"; v="${v%/}"; printf '%s' "$v"; }
DOMAIN="$(strip_url_prefix "$DOMAIN")"
ADMIN_DOMAIN="$(strip_url_prefix "$ADMIN_DOMAIN")"

echo
c_info "===== PostgreSQL ====="
c_info "（直接填写外部地址；留空或填 localhost / 127.0.0.1 则自动部署内置 PostgreSQL）"
DB_HOST="$(ask 'PostgreSQL 地址（主机名或 IP）' "${DB_HOST:-}")"
DB_PORT="$(ask 'PostgreSQL 端口' "${DB_PORT:-5432}")"
DB_USER="$(ask 'PostgreSQL 用户' "${DB_USER:-postgres}")"
DB_PASSWORD="$(ask_secret 'PostgreSQL 密码' "${DB_PASSWORD:-}")"
[ -n "$DB_PASSWORD" ] || DB_PASSWORD="$(gen_secret 16)"
DB_NAME="$(ask 'PostgreSQL 库名' "${DB_NAME:-novablog}")"
DB_SSLMODE="$(ask 'PostgreSQL sslmode' "${DB_SSLMODE:-disable}")"
# 判断是否为内置模式
if [ -z "$DB_HOST" ] || [ "$DB_HOST" = "localhost" ] || [ "$DB_HOST" = "127.0.0.1" ]; then
  DB_MODE="local"
  DB_HOST="postgres"
else
  DB_MODE="external"
fi

echo
c_info "===== Redis ====="
c_info "（直接填写外部地址；留空或填 localhost / 127.0.0.1 则自动部署内置 Redis）"
REDIS_HOST="$(ask 'Redis 地址（主机名或 IP）' "${REDIS_HOST:-}")"
REDIS_PORT="$(ask 'Redis 端口' "${REDIS_PORT:-6379}")"
REDIS_PASSWORD="$(ask_secret 'Redis 密码（无密码直接回车）' "${REDIS_PASSWORD:-}")"
REDIS_DB="$(ask 'Redis 库编号' "${REDIS_DB:-3}")"
# 判断是否为内置模式
if [ -z "$REDIS_HOST" ] || [ "$REDIS_HOST" = "localhost" ] || [ "$REDIS_HOST" = "127.0.0.1" ]; then
  REDIS_MODE="local"
  REDIS_HOST="redis"
else
  REDIS_MODE="external"
fi

# 外部服务连通性预检：地址填错若等到容器启动后才暴露，后端会反复崩溃重启且现象只有 502
echo
c_info "===== 外部服务连通性预检 ====="
if [ "$DB_MODE" = "external" ]; then
  if tcp_reachable "$DB_HOST" "$DB_PORT"; then
    c_info "PostgreSQL $DB_HOST:$DB_PORT 可达"
  else
    die "无法连接 PostgreSQL $DB_HOST:$DB_PORT —— 请核对地址/端口是否填对、实例是否放行本机 IP"
  fi
fi
if [ "$REDIS_MODE" = "external" ]; then
  if tcp_reachable "$REDIS_HOST" "$REDIS_PORT"; then
    c_info "Redis $REDIS_HOST:$REDIS_PORT 可达"
  else
    die "无法连接 Redis $REDIS_HOST:$REDIS_PORT —— 请核对地址/端口是否填对、实例是否放行本机 IP"
  fi
fi

echo
c_info "===== 挂载目录 ====="
c_info "（相对路径以 deploy/ 为基准，建议填绝对路径）"
CONFIG_DIR="$(abs_dir "$(ask 'config.yaml 存放目录' "${CONFIG_DIR:-./config}")")"
UPLOADS_DIR="$(abs_dir "$(ask '媒体上传目录 uploads_data' "${UPLOADS_DIR:-./data/uploads}")")"
THEMES_DIR="$(abs_dir "$(ask '主题制品目录 themes_data' "${THEMES_DIR:-./data/themes}")")"
LOGS_DIR="$(abs_dir "$(ask '日志目录' "${LOGS_DIR:-./data/logs}")")"
FRONTEND_DIR_INPUT="$(ask '自备博客前端目录（含 theme.json 与 dist/；回车沿用当前值，输入 - 清除）' "${FRONTEND_DIR:-}")"
# 输入 - / none / off 表示清除已配置的挂载前端，回退到后台安装的主题
case "$FRONTEND_DIR_INPUT" in
  -|none|off|no|NONE) FRONTEND_DIR_INPUT="" ;;
esac
if [ -n "$FRONTEND_DIR_INPUT" ]; then
  FRONTEND_DIR="$(abs_dir "$FRONTEND_DIR_INPUT")"
else
  FRONTEND_DIR="$(abs_dir './data/blog-frontend')"
fi
if [ "$DB_MODE" = "local" ]; then
  PGDATA_DIR="$(abs_dir "$(ask '内置 PostgreSQL 数据目录' "${PGDATA_DIR:-./data/pg}")")"
fi
if [ "$REDIS_MODE" = "local" ]; then
  REDISDATA_DIR="$(abs_dir "$(ask '内置 Redis 数据目录' "${REDISDATA_DIR:-./data/redis}")")"
fi
[ -n "$MARKET_URL" ] || MARKET_URL="$(env_get NOVABLOG_MARKET_URL)"

echo
c_info "===== 镜像 ====="
if [ "$BUILD" = 1 ]; then
  VERSION="$(ask '本地构建的镜像标签' "${VERSION:-dev}")"
else
  VERSION="$(ask '镜像版本号' "${VERSION:-latest}")"
fi
IMAGE_REPO="$(ask '镜像仓库' "$IMAGE_REPO")"

[ -n "$BLOG_PORT" ] || die "博客前端端口不能为空"
[ -n "$ADMIN_PORT" ] || die "CMS 后台端口不能为空"
[ "$BLOG_PORT" != "$ADMIN_PORT" ] || die "博客前端端口与 CMS 后台端口不能相同"

# ============================== 生成 config.yaml ==============================
CONFIG_FILE="$CONFIG_DIR/config.yaml"
WRITE_CONFIG=1
if [ -f "$CONFIG_FILE" ]; then
  if [ "$ASSUME_YES" = 1 ]; then
    c_info "已存在 ${CONFIG_FILE}，--yes 模式沿用现有配置（不覆盖）。"
    WRITE_CONFIG=0
  else
    keep="$(ask_choice "已存在 ${CONFIG_FILE}，是否覆盖？覆盖会重写数据库/端口等配置（JWT 密钥会保留）" 'keep|overwrite' 'keep')"
    [ "$keep" = "keep" ] && WRITE_CONFIG=0
  fi
fi

if [ "$WRITE_CONFIG" = 1 ]; then
  # 复用已有密钥，避免覆盖后登录态失效
  ACCESS_SECRET=""; REFRESH_SECRET=""; CRYPTO_SECRET=""
  if [ -f "$CONFIG_FILE" ]; then
    ACCESS_SECRET="$(sed -n 's/^[[:space:]]*access_secret:[[:space:]]*"\{0,1\}\([^"]*\)"\{0,1\}[[:space:]]*$/\1/p' "$CONFIG_FILE" | head -1)"
    REFRESH_SECRET="$(sed -n 's/^[[:space:]]*refresh_secret:[[:space:]]*"\{0,1\}\([^"]*\)"\{0,1\}[[:space:]]*$/\1/p' "$CONFIG_FILE" | head -1)"
    CRYPTO_SECRET="$(sed -n 's/^[[:space:]]*secret_key:[[:space:]]*"\{0,1\}\([^"]*\)"\{0,1\}[[:space:]]*$/\1/p' "$CONFIG_FILE" | head -1)"
  fi
  [ -n "$ACCESS_SECRET" ] || ACCESS_SECRET="$(gen_secret 32)"
  [ -n "$REFRESH_SECRET" ] || REFRESH_SECRET="$(gen_secret 32)"
  [ -n "$CRYPTO_SECRET" ] || CRYPTO_SECRET="$(gen_secret 16)"

  if [ -n "$FRONTEND_DIR_INPUT" ]; then
    FRONTEND_DIR_YAML="/app/blog-frontend"
  else
    FRONTEND_DIR_YAML=""
  fi

  c_info "写入配置：$CONFIG_FILE"
  cat > "$CONFIG_FILE" <<YAML
# 由 deploy.sh 生成于 $(date '+%Y-%m-%d %H:%M:%S')；字段说明见 backend/config/config.example.yaml
app:
  name: backend
  env: production
  debug: false

http:
  port: 8111
  read_timeout: 30s
  write_timeout: 30s

log:
  level: info
  file_path: /app/logs/app.log
  max_size: 100
  max_backups: 10
  max_age: 30

postgres:
  host: "$(yaml_quote "$DB_HOST")"
  port: $DB_PORT
  user: "$(yaml_quote "$DB_USER")"
  password: "$(yaml_quote "$DB_PASSWORD")"
  dbname: "$(yaml_quote "$DB_NAME")"
  sslmode: "$(yaml_quote "$DB_SSLMODE")"
  max_open_conns: 100
  max_idle_conns: 10

redis:
  host: "$(yaml_quote "$REDIS_HOST")"
  port: $REDIS_PORT
  password: "$(yaml_quote "$REDIS_PASSWORD")"
  db: $REDIS_DB

upload:
  dir: /app/uploads
  base_url: "$(yaml_quote "$PUBLIC_URL")"

jwt:
  access_secret: "$(yaml_quote "$ACCESS_SECRET")"
  refresh_secret: "$(yaml_quote "$REFRESH_SECRET")"
  access_expire: 2h
  refresh_expire: 168h

crypto:
  secret_key: "$(yaml_quote "$CRYPTO_SECRET")"

security:
  enabled: true
  blacklist_ttl_minutes: 60
  log_retention_days: 7

themes:
  data_dir: /app/data/themes
  frontend_dir: "$(yaml_quote "$FRONTEND_DIR_YAML")"
  market_base_url: "$(yaml_quote "$MARKET_URL")"
  github_token: ""
  public_api_base: ""
  max_artifact_mb: 100

cors:
  allowed_origins: "$(yaml_quote "$PUBLIC_URL,http://localhost:5173,http://localhost:5174")"
YAML
  chmod 600 "$CONFIG_FILE"
else
  c_info "沿用现有配置：$CONFIG_FILE"
fi

# ============================== 生成 .env ==============================
c_info "写入环境变量：$ENV_FILE"
cat > "$ENV_FILE" <<ENV
# 由 deploy.sh 生成于 $(date '+%Y-%m-%d %H:%M:%S')

# ---- 镜像与版本号 ----
NOVABLOG_IMAGE_REPO=$IMAGE_REPO
NOVABLOG_VERSION=$VERSION
NOVABLOG_BUILD_VERSION=$( [ "$VERSION" = "latest" ] && printf 'dev' || printf '%s' "$VERSION" )

# ---- 入口端口与域名 ----
NOVABLOG_DOMAIN=$DOMAIN
NOVABLOG_HTTP_PORT=$BLOG_PORT
NOVABLOG_ADMIN_DOMAIN=$ADMIN_DOMAIN
NOVABLOG_ADMIN_HTTP_PORT=$ADMIN_PORT
NOVABLOG_ADMIN_CONTAINER_PORT=8080
NOVABLOG_ADMIN_BASE=/admin/

# ---- 挂载目录 ----
NOVABLOG_CONFIG_DIR=$CONFIG_DIR
NOVABLOG_UPLOADS_DIR=$UPLOADS_DIR
NOVABLOG_THEMES_DIR=$THEMES_DIR
NOVABLOG_LOGS_DIR=$LOGS_DIR
NOVABLOG_BLOG_FRONTEND_DIR=$FRONTEND_DIR

# ---- 部署方式（供 --status/--logs/--down 还原 compose 文件组合） ----
NOVABLOG_DB_MODE=$DB_MODE
NOVABLOG_REDIS_MODE=$REDIS_MODE

# ---- 内置数据库 ----
NOVABLOG_PGDATA_DIR=$PGDATA_DIR
NOVABLOG_REDISDATA_DIR=$REDISDATA_DIR
POSTGRES_USER=$DB_USER
POSTGRES_PASSWORD=$DB_PASSWORD
POSTGRES_DB=$DB_NAME
REDIS_PASSWORD=$REDIS_PASSWORD

# ---- 其他 ----
NOVABLOG_PUBLIC_URL=$PUBLIC_URL
NOVABLOG_MARKET_URL=$MARKET_URL
ENV
chmod 600 "$ENV_FILE"

# ============================== 镜像校验 ==============================
if [ "$BUILD" = 0 ] && [ "$SKIP_IMAGE_CHECK" = 0 ]; then
  c_info "校验镜像 $IMAGE_REPO:$VERSION ..."
  if ! docker buildx imagetools inspect "$IMAGE_REPO:$VERSION" >/dev/null 2>&1; then
    c_err "镜像 $IMAGE_REPO:$VERSION 不可用。"
    cat <<EOF

可能原因：
  1) 版本号不存在 —— 可用版本见 https://github.com/StudyNoWeekend/novablog-admin/pkgs/container/novablog-admin
  2) 私有镜像未登录 —— 先执行：echo \$GH_TOKEN | docker login ghcr.io -u <用户名> --password-stdin
  3) 需要本地构建 —— 改用：./deploy.sh --build

确认版本号后可重试，或加 --skip-image-check 跳过校验。
EOF
    exit 1
  fi
fi

# ============================== 启动 ==============================
c_info "启动容器（镜像 ${IMAGE_REPO}:${VERSION}）..."
if [ "$BUILD" = 1 ]; then
  run_compose up -d --build || compose_up_failed
else
  run_compose pull --quiet || c_warn "镜像拉取失败（将使用本地已有镜像继续）"
  run_compose up -d || compose_up_failed
fi

# 等待健康检查
c_info "等待服务就绪 ..."
HEALTHY=0
for _ in $(seq 1 60); do
  cid="$(run_compose ps -q novablog 2>/dev/null || true)"
  if [ -n "$cid" ]; then
    status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$cid" 2>/dev/null || true)"
    if [ "$status" = "healthy" ]; then HEALTHY=1; break; fi
    if [ "$status" = "unhealthy" ] || [ "$status" = "exited" ]; then break; fi
  fi
  sleep 2
done

if [ "$HEALTHY" != 1 ]; then
  c_err "服务未能就绪（容器未进入 healthy 状态），最近日志如下："
  cid="$(run_compose ps -q novablog 2>/dev/null || true)"
  if [ -n "$cid" ]; then
    docker logs --tail 30 "$cid" 2>&1 | tail -30
  fi
  c_err "请根据上方日志排查（典型原因：外部数据库/Redis 地址或密码填错），完整日志：./deploy.sh --logs"
  exit 1
fi

HOST_FOR_URL="localhost"
[ "$DOMAIN" != "_" ] && HOST_FOR_URL="$DOMAIN"

cat <<EOF

$(c_info '部署完成')

  博客入口     http://$HOST_FOR_URL:$BLOG_PORT/
  后台入口     http://$HOST_FOR_URL:$ADMIN_PORT/admin/
  首装向导     http://$HOST_FOR_URL:$ADMIN_PORT/admin/setup
  镜像版本     $IMAGE_REPO:$VERSION

  配置目录     $CONFIG_FILE
  上传目录     $UPLOADS_DIR
  主题目录     $THEMES_DIR
  日志目录     $LOGS_DIR
  前端挂载     $FRONTEND_DIR$([ -n "$FRONTEND_DIR_INPUT" ] && printf ' （已启用 themes.frontend_dir）' || printf ' （空目录，使用后台安装的主题）')

  常用命令：
    ./deploy.sh --status     查看状态
    ./deploy.sh --logs       查看日志
    ./deploy.sh --down       停止容器
    ./deploy.sh --version v1.0.0   切换到指定版本（数据保留）

  首次使用：打开后台入口完成安装向导（创建博主账号 → 存储 → 主题）。
  若使用外部 PostgreSQL，请确认库已存在或该账号有建库权限（服务启动会自动建库与建表）。
EOF
