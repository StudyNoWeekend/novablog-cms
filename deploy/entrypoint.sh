#!/bin/sh
# NovaBlog 镜像入口脚本
#
# 作用：渲染 nginx 配置模板后启动 supervisord。
# 官方 nginx 入口脚本仅在容器命令为 nginx/nginx-debug 时才会执行 /docker-entrypoint.d/
# 下的模板渲染（envsubst），而本镜像以 supervisord 为 PID 1，因此在此显式渲染。
set -e

render_templates() {
    [ -d /etc/nginx/templates ] || return 0
    for tpl in /etc/nginx/templates/*.template; do
        [ -f "$tpl" ] || continue
        out="/etc/nginx/conf.d/$(basename "$tpl" .template)"
        # 仅替换环境中真实存在的变量，避免误伤 $uri / $host 等 nginx 内置变量
        # shellcheck disable=SC2016
        envsubst "$(printf '${%s} ' $(env | cut -d= -f1))" < "$tpl" > "$out"
        echo "[entrypoint] 渲染 nginx 配置：$out"
    done
}

render_templates

if ! nginx -t 2>/dev/null; then
    echo "[entrypoint] nginx 配置校验失败：" >&2
    nginx -t >&2 || true
    exit 1
fi

exec "$@"
