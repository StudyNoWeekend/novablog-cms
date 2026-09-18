# NovaBlog nginx 统一入口镜像
# 构建上下文为仓库根目录：
#   docker build -f deploy/nginx.Dockerfile -t novablog-nginx .
#
# 包含管理后台静态文件（/admin）与统一入口配置模板；
# 博客页面与 API 由后端 Go 服务提供（nginx 反代）。

# ---------- 管理后台前端构建 ----------
FROM node:20-alpine AS frontend-builder
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
ARG VITE_BASE_PATH=/admin/
ENV VITE_BASE_PATH=$VITE_BASE_PATH
RUN npx vite build

# ---------- nginx 运行层 ----------
FROM nginx:1.27-alpine
COPY --from=frontend-builder /src/frontend/dist /usr/share/nginx/html/admin
COPY deploy/nginx/novablog.conf.template /etc/nginx/templates/default.conf.template
# nginx 官方镜像启动时对 /etc/nginx/templates/*.template 做 envsubst 后输出到 conf.d
# 双入口端口/域名默认值（可在 compose environment 中覆盖）
ENV NOVABLOG_DOMAIN=_ \
    NOVABLOG_ADMIN_PORT=8080 \
    NOVABLOG_ADMIN_DOMAIN=_ \
    NOVABLOG_ADMIN_BASE=/admin/
EXPOSE 80 8080
