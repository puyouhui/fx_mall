# 支持代理配置（通过 build-arg 传入）
ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG http_proxy
ARG https_proxy

# 构建阶段 - 使用 Alpine 版本的 node:16 镜像（更小）
# 注意：如果遇到 403 错误，请先手动拉取镜像：
# HTTP_PROXY=http://127.0.0.1:7897 HTTPS_PROXY=http://127.0.0.1:7897 docker pull node:16-alpine
FROM node:16-alpine AS builder

# 设置代理环境变量
ENV HTTP_PROXY=${HTTP_PROXY}
ENV HTTPS_PROXY=${HTTPS_PROXY}
ENV http_proxy=${http_proxy}
ENV https_proxy=${https_proxy}

# 设置工作目录
WORKDIR /build

# 复制 package.json 和 package-lock.json（如果存在）
COPY package*.json ./

# 安装所有依赖（包括 devDependencies，用于构建）
# 使用 npm install 因为可能没有 package-lock.json
RUN npm install && \
    npm cache clean --force

# 复制源代码和配置文件
COPY . .

# 构建项目（这会生成 dist/ 目录并复制 SSL 证书）
RUN npm run build

# 生产阶段 - 使用 Alpine 版本的 node:16 镜像（更小）
FROM node:16-alpine

# 设置代理环境变量（用于拉取镜像和安装依赖）
ENV HTTP_PROXY=${HTTP_PROXY}
ENV HTTPS_PROXY=${HTTPS_PROXY}
ENV http_proxy=${http_proxy}
ENV https_proxy=${https_proxy}

# 设置工作目录
WORKDIR /node-hiprint-transit

# 只复制 package.json 用于安装生产依赖
COPY package*.json ./

# 只安装生产依赖，并清理缓存
# 使用 npm install 因为可能没有 package-lock.json
RUN npm install --only=production && \
    npm cache clean --force && \
    rm -rf /tmp/* /var/cache/apk/* /root/.npm && \
    find ./node_modules -name "*.md" -delete 2>/dev/null || true && \
    find ./node_modules -name "*.txt" -delete 2>/dev/null || true && \
    find ./node_modules -name "test" -type d -exec rm -rf {} + 2>/dev/null || true && \
    find ./node_modules -name "tests" -type d -exec rm -rf {} + 2>/dev/null || true && \
    find ./node_modules -name "*.test.js" -delete 2>/dev/null || true && \
    find ./node_modules -name "*.spec.js" -delete 2>/dev/null || true

# 从构建阶段复制构建产物（只复制必要的文件）
COPY --from=builder /build/dist ./dist

# 创建日志目录
RUN mkdir -p logs

# 暴露应用运行的端口
EXPOSE 17521

# 启动应用
CMD ["node", "dist/index.js"]
