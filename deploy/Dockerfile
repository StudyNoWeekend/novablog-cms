# NovaBlog CMS 后端镜像（多阶段构建）
# 构建上下文为仓库根目录：
#   docker build -f deploy/Dockerfile -t novablog-backend .
#
# 运行层不含 Node 工具链。

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
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/novablog ./cmd/api

# ---------- ③ 运行层（alpine，无 Node） ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl \
    && addgroup -g 10001 nova \
    && adduser -D -u 10001 -G nova nova

WORKDIR /app
COPY --from=backend-builder /out/novablog ./novablog
# 内置一份默认配置，生产经卷挂载覆盖（见 deploy/README.md）
COPY backend/config/config.example.yaml ./config/config.yaml

RUN mkdir -p /app/data/themes /app/uploads /app/logs \
    && chown -R nova:nova /app
USER nova

ENV TZ=Asia/Shanghai
EXPOSE 8111
VOLUME ["/app/data", "/app/uploads"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD curl -fsS http://localhost:8111/health || exit 1

ENTRYPOINT ["./novablog"]
