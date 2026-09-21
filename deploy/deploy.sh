#!/usr/bin/env bash
#
# NovaBlog 部署脚本
#   - 按博客名称隔离部署：每个博客一个 <名称>/ 目录（.env、config.yaml 与全部挂载数据）
#     和一组独立容器（前缀 <名称>-），同机部署多个博客互不影响
#   - 可选内置 PostgreSQL / Redis，或使用你自己的外部实例（多博客共用时默认按名称分库）
#   - 可指定博客前端端口、CMS 后台端口、入口域名，以及镜像版本号（不指定则使用 latest）
#
# 示例：
#   ./deploy.sh --name llcms                      # 交互式问答（博客名称必填）
#   ./deploy.sh --name llcms --yes                # 全部默认（内置 PG/Redis，端口 80/8080，latest）
#   ./deploy.sh --name llcms --blog-port 9001 --admin-port 9002 --domain blog.example.com
#   ./deploy.sh --name llcms --version v1.0.1     # 升级指定博客（数据保留）
#   ./deploy.sh --name llcms --status | --logs | --down
#   ./deploy.sh --name other --blog-port 9003 --admin-port 9004   # 再部署一个博客
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
# ENV_FILE / BLOG_DIR 在「博客名称解析」后确定：$SCRIPT_DIR/<名称>/.env
NAME=""
BLOG_DIR=""
ENV_FILE=""

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

CONFIG_DIR="" UPLOADS_DIR="" THEMES_DIR="" LOGS_DIR="" FRONTEND_DIR="" FRONTEND_DIR_INPUT=""
PGDATA_DIR="" REDISDATA_DIR=""
MARKET_URL=""

VERSION_SET=0

usage() {
  cat <<'EOF'
NovaBlog 部署脚本

用法：./deploy.sh --name <博客名称> [选项]

博客名称（必填，同机多博客的隔离键）
  --name <名称>          用作数据目录名（脚本同级 <名称>/，存放 .env / config.yaml / 全部挂载数据）
                         与容器项目名（容器前缀 <名称>-，网络 <名称>_novablog）。
                         小写字母/数字/中划线/下划线；不同博客用不同名称与端口即可完全隔离

镜像与版本
  --version <tag>        镜像版本号，不指定则使用 latest（例：--version v1.0.0）
  --image <repo>         镜像仓库地址（默认 ghcr.io/studynoweekend/novablog-cms）
  --build                本地构建镜像，而非拉取 registry 镜像
  --skip-image-check     跳过镜像 tag 存在性校验

入口端口与域名
  --blog-port <端口>     博客前端端口（宿主，默认 80；多博客同机时各用不同端口）
  --admin-port <端口>    CMS 后台端口（宿主，默认 8080）
  --public-url <地址>    博客对外访问地址（默认 http://<博客入口域名>，未配域名时 http://localhost:<博客端口>）
  --domain <域名>        博客入口 server_name（默认 _，即任意域名/ IP）
  --admin-domain <域名>  后台入口 server_name（默认 _）

数据库（PostgreSQL）
  --db-host <主机>     PostgreSQL 地址；留空或 localhost/127.0.0.1 时自动部署内置容器
  --db-port <端口>       默认 5432
  --db-user <用户>       默认 postgres
  --db-password <密码>   必填（内置模式下自动生成）
  --db-name <库名>       默认与博客名称相同（多博客共用同一 PG 时天然分库）
  --db-sslmode <模式>    默认 disable

缓存（Redis）
  --redis-host <主机>    Redis 地址；留空或 localhost/127.0.0.1 时自动部署内置容器
  --redis-port <端口>    默认 6379
  --redis-password <密码> 无密码可留空（内置模式下自动生成）
  --redis-db <编号>      默认 3（多博客共用同一 Redis 时请为每个博客分配不同编号）

挂载目录（默认全部收拢在脚本同级 <名称>/ 下；相对路径以脚本所在目录为基准）
  --config-dir <目录>    config.yaml 存放目录（默认 ./<名称>/config）
  --uploads-dir <目录>   媒体上传目录（默认 ./<名称>/data/uploads）
  --themes-dir <目录>    主题制品目录（默认 ./<名称>/data/themes）
  --logs-dir <目录>      日志目录（默认 ./<名称>/data/logs）
  --frontend-dir <目录>  自备博客前端目录（含 theme.json 与 dist/），传 - 清除；
                         不指定则使用后台安装的主题
  --pgdata-dir <目录>    内置 PostgreSQL 数据目录（默认 ./<名称>/data/pg）
  --redisdata-dir <目录> 内置 Redis 数据目录（默认 ./<名称>/data/redis）

其他
  --market-url <地址>    官方主题市场地址（首装拉取默认主题用，可留空）
  -y, --yes              全部使用默认值/已有配置，不进入交互问答
  --status               查看容器状态（可 --name 指定博客，仅一个博客时自动选中）
  --logs                 跟踪应用日志
  --down                 停止并移除容器（挂载目录中的数据保留）
  -h, --help             显示本帮助
EOF
}

# ============================== 参数解析 ==============================
while [ $# -gt 0 ]; do
  case "$1" in
    --name) NAME="${2:?--name 需要一个值}"; shift 2 ;;
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
    --frontend-dir) FRONTEND_DIR_INPUT="${2:?--frontend-dir 需要一个值}"; shift 2 ;;
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

查看详细日志：./deploy.sh --name ${NAME} --logs
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

# ============================== 博客名称解析 ==============================
# 博客名称是同机多博客的隔离键：数据目录 <名称>/、容器项目名 <名称>。
validate_name() {
  [ -n "$1" ] || die "博客名称不能为空（交互输入，或使用 --name <名称>）"
  printf '%s' "$1" | grep -qE '^[a-z0-9][a-z0-9_-]*$' \
    || die "博客名称不合法：$1（仅允许小写字母、数字、中划线、下划线，且以字母或数字开头）"
}

list_blogs() {
  local d
  for d in "$SCRIPT_DIR"/*/; do
    [ -f "${d}.env" ] && printf '%s\n' "$(basename "$d")"
  done
  return 0
}

if [ -z "$NAME" ] && [ "$ACTION" != "up" ]; then
  # status/logs/down：未指定名称时，仅一个博客自动选中，多个则交互选择
  blogs="$(list_blogs)"
  if [ -z "$blogs" ]; then
    if [ -f "$SCRIPT_DIR/.env" ]; then
      die "检测到旧版根目录 .env（单博客旧格式，不支持按名称管理）。请先在本目录执行 docker compose down 停止旧容器并删除该文件，再用 --name <名称> 重新部署"
    fi
    die "未找到任何博客部署记录（<名称>/.env）。先完成部署：./deploy.sh --name <名称>"
  fi
  if [ "$(printf '%s\n' "$blogs" | wc -l)" = 1 ]; then
    NAME="$(printf '%s' "$blogs" | head -1)"
  elif [ "$INTERACTIVE" = 1 ]; then
    c_info "检测到多个博客："
    printf '%s\n' "$blogs" | sed 's/^/  - /'
    NAME="$(ask_choice '选择要操作的博客' "$(printf '%s\n' "$blogs" | paste -sd'|' -)" "$(printf '%s' "$blogs" | head -1)")"
  else
    die "检测到多个博客，请用 --name 指定其一：$(printf '%s\n' "$blogs" | paste -sd' ' -)"
  fi
elif [ -z "$NAME" ] && [ "$INTERACTIVE" = 1 ]; then
  echo
  c_info "===== 博客名称 ====="
  c_info "同机多博客的隔离键：每个博客独立目录 <名称>/ 与独立容器（前缀 <名称>-），互不影响"
  NAME="$(ask '博客名称（如 llcms；字母/数字/-/_，自动转小写）' "")"
fi
# 统一转小写：名称同时用作目录名与容器项目名，保持全小写避免大小写歧义
NAME="$(printf '%s' "$NAME" | tr 'A-Z' 'a-z')"
validate_name "$NAME"
BLOG_DIR="$SCRIPT_DIR/$NAME"
ENV_FILE="$BLOG_DIR/.env"
if [ -e "$BLOG_DIR" ] && [ ! -d "$BLOG_DIR" ]; then
  die "$BLOG_DIR 已存在且不是目录，请换一个博客名称"
fi
if [ "$ACTION" != "up" ]; then
  [ -f "$ENV_FILE" ] || die "博客 $NAME 尚未部署（缺 ${ENV_FILE}），先运行 ./deploy.sh --name $NAME 完成部署"
fi
# 旧版根目录 .env（单博客格式）仅作迁移提示，不再参与部署
if [ "$ACTION" = "up" ] && [ -f "$SCRIPT_DIR/.env" ]; then
  c_warn "检测到旧版根目录 .env（旧版单博客格式）。若旧容器仍在运行请先停止（本目录执行 docker compose down），避免端口冲突；迁移完成后可删除该文件"
fi

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
  # --env-file 指向博客自己的 .env，-p 用博客名称作项目名：容器/网络按博客隔离
  docker compose --env-file "$ENV_FILE" -p "$NAME" "${args[@]}" "$@"
}

run_compose_from_env() {
  local f
  local args=()
  while IFS= read -r f; do [ -n "$f" ] && args+=("$f"); done < <(compose_files_from_env)
  ensure_compose_files "${args[@]}"
  docker compose --env-file "$ENV_FILE" -p "$NAME" "${args[@]}" "$@"
}

case "$ACTION" in
  status)
    run_compose_from_env ps
    exit 0
    ;;
  logs)
    run_compose_from_env logs -f --tail=200 novablog
    exit 0
    ;;
  down)
    c_info "停止并移除容器（挂载目录中的数据保留）..."
    run_compose_from_env down
    c_info "已停止。数据仍在 $ENV_FILE 配置的挂载目录里，重新部署执行 ./deploy.sh --name $NAME 即可。"
    exit 0
    ;;
esac

# --build 需要完整仓库作为构建上下文（frontend/ backend/ 等），缺失时早失败（仅部署路径）
if [ "$BUILD" = 1 ] && { [ ! -f "$SCRIPT_DIR/Dockerfile" ] || [ ! -d "$SCRIPT_DIR/../backend" ]; }; then
  die "--build 需要完整仓库（构建上下文为仓库根，需 frontend/ 与 backend/）。请 git clone 仓库后在 deploy/ 目录内运行，或去掉 --build 直接使用线上镜像。"
fi

# ============================== 收集配置 ==============================
if [ -f "$ENV_FILE" ]; then
  c_info "检测到博客 $NAME 已有部署记录（${ENV_FILE}），复用其中配置作为默认值（可直接回车沿用）。"
fi

[ -z "$IMAGE_REPO" ] && IMAGE_REPO="$(env_get NOVABLOG_IMAGE_REPO)"
[ "$VERSION_SET" = 0 ] && VERSION="$(env_get NOVABLOG_VERSION)"
[ -z "$BLOG_PORT" ] && BLOG_PORT="$(env_get NOVABLOG_HTTP_PORT)"
[ -z "$ADMIN_PORT" ] && ADMIN_PORT="$(env_get NOVABLOG_ADMIN_HTTP_PORT)"
[ -z "$DOMAIN" ] && DOMAIN="$(env_get NOVABLOG_DOMAIN)"
[ -z "$ADMIN_DOMAIN" ] && ADMIN_DOMAIN="$(env_get NOVABLOG_ADMIN_DOMAIN)"
[ -z "$PUBLIC_URL" ] && PUBLIC_URL="$(env_get NOVABLOG_PUBLIC_URL)"
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
[ -z "$DB_USER" ] && DB_USER="$(env_get POSTGRES_USER)"
[ -z "$DB_NAME" ] && DB_NAME="$(env_get POSTGRES_DB)"
[ -z "$REDIS_PASSWORD" ] && REDIS_PASSWORD="$(env_get REDIS_PASSWORD)"

echo
c_info "===== 入口端口与域名 ====="
BLOG_PORT="$(ask '博客前端端口' "${BLOG_PORT:-80}")"
ADMIN_PORT="$(ask 'CMS 后台端口' "${ADMIN_PORT:-8080}")"
DOMAIN="$(ask '博客入口域名（裸域名如 blog.example.com；任意域名/IP 填 _）' "${DOMAIN:-_}")"
ADMIN_DOMAIN="$(ask '后台入口域名（裸域名；任意域名/IP 填 _）' "${ADMIN_DOMAIN:-_}")"
# 误带协议或结尾斜杠时自动清洗，否则会渲染进 nginx server_name 导致域名匹配失效
strip_url_prefix() { local v="$1"; v="${v#http://}"; v="${v#https://}"; v="${v%/}"; printf '%s' "$v"; }
DOMAIN="$(strip_url_prefix "$DOMAIN")"
ADMIN_DOMAIN="$(strip_url_prefix "$ADMIN_DOMAIN")"
# 对外访问地址默认跟随博客域名（反代/标准端口场景）；未配域名时回退本机直连地址
public_url_def="http://localhost:$BLOG_PORT"
[ "$DOMAIN" != "_" ] && public_url_def="http://$DOMAIN"
PUBLIC_URL="$(ask '博客对外访问地址（媒体文件 URL 基址，访客实际访问地址）' "${PUBLIC_URL:-$public_url_def}")"
# 未带协议时自动补 http://，否则媒体 URL 与 CORS 白名单会生成非法值
case "$PUBLIC_URL" in
  http://*|https://*) ;;
  *) PUBLIC_URL="http://$PUBLIC_URL"; c_info "博客对外访问地址已自动补全为：$PUBLIC_URL" ;;
esac

echo
c_info "===== PostgreSQL ====="
c_info "（填外部地址；留空或填 localhost / 127.0.0.1 则自动部署内置 PostgreSQL）"
DB_HOST="$(ask 'PostgreSQL 地址（主机名或 IP）' "${DB_HOST:-}")"
DB_PORT="$(ask 'PostgreSQL 端口' "${DB_PORT:-5432}")"
DB_PASSWORD="$(ask_secret 'PostgreSQL 密码（内置模式自动生成，直接回车即可）' "${DB_PASSWORD:-}")"
[ -n "$DB_PASSWORD" ] || DB_PASSWORD="$(gen_secret 16)"
# 用户/库名/sslmode 属低频配置，不再逐项提问：默认 postgres / <博客名> / disable，可用 --db-user/--db-name/--db-sslmode 覆盖
[ -n "$DB_USER" ] || DB_USER="postgres"
[ -n "$DB_NAME" ] || DB_NAME="$NAME"
[ -n "$DB_SSLMODE" ] || DB_SSLMODE="disable"
# 判断是否为内置模式
if [ -z "$DB_HOST" ] || [ "$DB_HOST" = "localhost" ] || [ "$DB_HOST" = "127.0.0.1" ]; then
  DB_MODE="local"
  DB_HOST="postgres"
else
  DB_MODE="external"
fi

echo
c_info "===== Redis ====="
c_info "（填外部地址；留空或填 localhost / 127.0.0.1 则自动部署内置 Redis）"
REDIS_HOST="$(ask 'Redis 地址（主机名或 IP）' "${REDIS_HOST:-}")"
REDIS_PORT="$(ask 'Redis 端口' "${REDIS_PORT:-6379}")"
REDIS_PASSWORD="$(ask_secret 'Redis 密码（无密码直接回车）' "${REDIS_PASSWORD:-}")"
REDIS_DB="$(ask 'Redis 库编号（多博客共用同一 Redis 时，请为每个博客分配不同编号）' "${REDIS_DB:-3}")"
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

# ============================== 挂载目录 ==============================
# 目录不再逐项提问：默认全部收拢在博客目录 <名称>/ 下；高级场景用 --config-dir 等 flag 覆盖
# （--frontend-dir 传 - / none 表示清除自备前端，回退到后台安装的主题）
case "$FRONTEND_DIR_INPUT" in
  -|none|off|no|NONE) FRONTEND_DIR_INPUT="" ;;
esac
if [ -z "$FRONTEND_DIR_INPUT" ] && [ "$(env_get NOVABLOG_FRONTEND_ENABLED)" = "1" ] && [ -n "$FRONTEND_DIR" ]; then
  # 重跑沿用此前启用的自备前端
  FRONTEND_DIR_INPUT="$FRONTEND_DIR"
fi
if [ -n "$FRONTEND_DIR_INPUT" ]; then
  FRONTEND_DIR="$(abs_dir "$FRONTEND_DIR_INPUT")"
else
  FRONTEND_DIR="$(abs_dir "./$NAME/data/blog-frontend")"
fi
CONFIG_DIR="$(abs_dir "${CONFIG_DIR:-./$NAME/config}")"
UPLOADS_DIR="$(abs_dir "${UPLOADS_DIR:-./$NAME/data/uploads}")"
THEMES_DIR="$(abs_dir "${THEMES_DIR:-./$NAME/data/themes}")"
LOGS_DIR="$(abs_dir "${LOGS_DIR:-./$NAME/data/logs}")"
if [ "$DB_MODE" = "local" ]; then
  PGDATA_DIR="$(abs_dir "${PGDATA_DIR:-./$NAME/data/pg}")"
fi
if [ "$REDIS_MODE" = "local" ]; then
  REDISDATA_DIR="$(abs_dir "${REDISDATA_DIR:-./$NAME/data/redis}")"
fi
[ -n "$MARKET_URL" ] || MARKET_URL="$(env_get NOVABLOG_MARKET_URL)"

echo
c_info "===== 镜像 ====="
if [ "$BUILD" = 1 ]; then
  VERSION="$(ask '本地构建的镜像标签' "${VERSION:-dev}")"
else
  VERSION="$(ask '镜像版本号' "${VERSION:-latest}")"
fi
# 镜像仓库不再提问：默认官方镜像，可用 --image 覆盖

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
  # 写超时放宽到 10 分钟：主题安装/更新需从 GitHub 下载制品，同步接口可能耗时较长
  write_timeout: 600s

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
  proxy_url: ""                      # 主题下载与 GitHub API 的 HTTP 代理；空=直连，代理不可用时自动降级直连
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
NOVABLOG_FRONTEND_ENABLED=$([ -n "$FRONTEND_DIR_INPUT" ] && printf 1 || printf 0)

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

# ============================== 端口占用预检 ==============================
# 同机多博客场景：目标端口被其它项目容器发布时提前失败，避免影响正在运行的博客
# （本博客自己的旧容器交给 compose up -d 原地重建，不算冲突）
port_conflicts() { # $1=端口 → 输出占用容器名
  docker ps --format '{{.Names}}\t{{.Ports}}' 2>/dev/null \
    | awk -F'\t' -v p=":${1}->" '$2 ~ p {print $1}' \
    | grep -v "^${NAME}-" || true
}

echo
c_info "===== 端口占用预检 ====="
for port_pair in "$BLOG_PORT:博客前端" "$ADMIN_PORT:CMS 后台"; do
  port="${port_pair%%:*}"
  label="${port_pair#*:}"
  hits="$(port_conflicts "$port")"
  if [ -n "$hits" ]; then
    c_err "$label 端口 $port 已被以下容器占用，为不影响在运行的博客，本次部署终止："
    printf '%s\n' "$hits" | sed 's/^/    - /'
    die "请为博客 $NAME 换用其它端口（--blog-port / --admin-port），或先停止占用端口的容器"
  fi
done
c_info "端口 ${BLOG_PORT}（博客）/ ${ADMIN_PORT}（后台）可用"

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
  c_err "请根据上方日志排查（典型原因：外部数据库/Redis 地址或密码填错），完整日志：./deploy.sh --name $NAME --logs"
  exit 1
fi

# 访客地址优先展示域名形态（走反代/标准端口时不带端口）；本机端口单独列出供反代与直连调试
if [ -n "$DOMAIN" ] && [ "$DOMAIN" != "_" ]; then
  BLOG_URL="http://$DOMAIN/"
else
  BLOG_URL="http://localhost:$BLOG_PORT/"
fi
if [ -n "$ADMIN_DOMAIN" ] && [ "$ADMIN_DOMAIN" != "_" ]; then
  ADMIN_URL="http://$ADMIN_DOMAIN/admin/"
  SETUP_URL="http://$ADMIN_DOMAIN/admin/setup"
else
  ADMIN_URL="http://localhost:$ADMIN_PORT/admin/"
  SETUP_URL="http://localhost:$ADMIN_PORT/admin/setup"
fi
FRONTEND_NOTE="（空目录，使用后台安装的主题）"
[ -n "$FRONTEND_DIR_INPUT" ] && FRONTEND_NOTE="（自备前端，已启用 themes.frontend_dir）"

cat <<EOF

$(c_info '部署完成')

  博客名称     $NAME
  数据目录     ${BLOG_DIR}（.env、config.yaml 与全部挂载数据）
  博客入口     $BLOG_URL
  后台入口     $ADMIN_URL
  首装向导     $SETUP_URL
  本机端口     ${BLOG_PORT}（博客）/ ${ADMIN_PORT}（后台）—— 外层反代目标，访客无需带端口
  镜像版本     $IMAGE_REPO:$VERSION
  前端目录     $FRONTEND_DIR $FRONTEND_NOTE

  常用命令：
    ./deploy.sh --name $NAME --status             查看状态
    ./deploy.sh --name $NAME --logs               查看日志
    ./deploy.sh --name $NAME --down               停止容器
    ./deploy.sh --name $NAME --version <tag>      升级/切换版本（数据保留）

  首次使用：打开后台入口完成安装向导（创建博主账号 → 存储 → 主题）。
  若使用外部 PostgreSQL，请确认库已存在或该账号有建库权限（服务启动会自动建库与建表）。
EOF
