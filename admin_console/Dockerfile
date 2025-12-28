# 独立构建版本：使用已构建好的 dist 目录
# 适用于：先本地构建，再打包成 Docker 镜像
#
# Usage:
#   1. 先本地构建: npm run build
#   2. 再构建镜像: docker build -f Dockerfile -t admin-console:latest .

# 支持代理配置（通过 build-arg 传入）
ARG HTTP_PROXY
ARG HTTPS_PROXY

FROM nginx:alpine

# 设置环境变量（用于容器运行时，如果需要）
ENV HTTP_PROXY=${HTTP_PROXY}
ENV HTTPS_PROXY=${HTTPS_PROXY}

# 复制已构建好的 dist 目录
COPY dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 5173

# 启动 Nginx
CMD ["/bin/sh", "-c", "nginx -g 'daemon off;'"]

