# NovaBlog 单镜像（nginx + Go 后端 + 管理后台前端）
# 构建上下文为仓库根目录：
#   docker build -f deploy/Dockerfile --build-arg VERSION=v1.0.0 -t novablog .
#
# 运行层由 supervisord 同时拉起：
#   - nginx     统一入口，:80 博客入口 / :8080 管理后台入口（反代 127.0.0.1:8111）
#   - novablog  Go 后端：业务 API + 博客前端静态托管（激活主题或挂载的自备前端）

ARG VERSION=dev

# ---------- ① 管理后台前端构建 ----------
FROM node:20-alpine AS frontend-builder
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
# 管理后台部署在 /admin 子路径（nginx 统一入口），如独立域名改为 /
ARG VITE_BASE_PATH=/admin/
ENV VITE_BASE_PATH=$VITE_BASE_PATH
RUN npx vite build

# ---------- ② Go 后端构建 ----------
FROM golang:1.26-alpine AS backend-builder
WORKDIR /src/backend
ENV GOPROXY=https://goproxy.cn,direct
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
ARG VERSION
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/novablog ./cmd/api

# ---------- ③ 运行层（nginx 入口 + Go 后端） ----------
FROM nginx:1.27-alpine

ARG VERSION
LABEL org.opencontainers.image.title="NovaBlog CMS" \
      org.opencontainers.image.description="面向独立博主的内容管理系统：Go 后端 + 管理后台 + 博客主题托管，单镜像交付" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.source="https://github.com/StudyNoWeekend/novablog-admin" \
      org.opencontainers.image.url="https://github.com/StudyNoWeekend/novablog-admin" \
      org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache supervisor tzdata curl

WORKDIR /app
COPY --from=backend-builder /out/novablog ./novablog
# 内置一份默认配置，生产经挂载覆盖（见 deploy/README.md）
COPY backend/config/config.example.yaml ./config/config.yaml
# 管理后台静态文件（base 路径须与构建期 VITE_BASE_PATH 一致，默认 /admin/）
COPY --from=frontend-builder /src/frontend/dist /usr/share/nginx/html/admin
# nginx 配置模板（由 deploy/entrypoint.sh 在启动时渲染到 /etc/nginx/conf.d/）
COPY deploy/nginx/novablog.conf.template /etc/nginx/templates/default.conf.template
COPY deploy/supervisord.conf /etc/supervisor/conf.d/novablog.conf
COPY deploy/entrypoint.sh /usr/local/bin/novablog-entrypoint.sh
RUN chmod +x /usr/local/bin/novablog-entrypoint.sh

# 运行期目录：自备博客前端挂载点 + 数据目录
RUN mkdir -p /app/uploads /app/data/themes /app/logs /app/blog-frontend

# 入口端口/域名默认值（envsubst 不支持 ${VAR:-default}，改默认值请改这里或在运行期覆盖）
ENV TZ=Asia/Shanghai \
    NOVABLOG_DOMAIN=_ \
    NOVABLOG_ADMIN_PORT=8080 \
    NOVABLOG_ADMIN_DOMAIN=_ \
    NOVABLOG_ADMIN_BASE=/admin/ \
    NOVABLOG_BACKEND_ADDR=127.0.0.1:8111

EXPOSE 80 8080
VOLUME ["/app/uploads", "/app/data", "/app/logs"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD curl -fsS http://127.0.0.1:8111/health || exit 1

ENTRYPOINT ["/usr/local/bin/novablog-entrypoint.sh"]
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/novablog.conf"]
