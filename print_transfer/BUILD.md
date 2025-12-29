# 构建和推送说明

## 快速开始

### 1. 构建并推送镜像

```bash
# 构建并推送时间戳标签（自动生成，格式：YYYYMMDDHHmm，例如：202512291337）
./build-and-push.sh

# 构建并推送 latest 标签
./build-and-push.sh latest

# 构建并推送指定标签
./build-and-push.sh v0.0.6
./build-and-push.sh 202512291337
```

### 2. 仅构建镜像（不推送）

```bash
# 使用代理构建
export HTTP_PROXY=http://127.0.0.1:7897
export HTTPS_PROXY=http://127.0.0.1:7897
docker build -t youhuipu/node-hiprint-transit:latest .
```

## 脚本功能

`build-and-push.sh` 脚本会自动完成以下操作：

1. ✅ 检查 Docker 是否运行
2. ✅ 检查并拉取基础镜像（node:16-alpine）
3. ✅ 构建 Docker 镜像
4. ✅ 推送镜像到 Docker Hub

## 配置

### 代理设置

脚本默认使用 `http://127.0.0.1:7897` 作为代理。如果需要修改，可以：

1. 编辑脚本中的 `PROXY_HOST` 和 `PROXY_PORT` 变量
2. 或设置环境变量：
   ```bash
   export HTTP_PROXY=http://your-proxy:port
   export HTTPS_PROXY=http://your-proxy:port
   ./build-and-push.sh
   ```

### Docker Hub 登录

首次推送前需要登录 Docker Hub：

```bash
docker login
```

输入你的 Docker Hub 用户名和密码。

## 镜像信息

- **镜像名称**: `youhuipu/node-hiprint-transit`
- **默认标签**: 时间戳格式（YYYYMMDDHHmm，例如：202512291337）
- **基础镜像**: `node:16-alpine`
- **镜像大小**: 约 136MB（优化后）

## 标签说明

- **时间戳标签**（默认）：格式为 `YYYYMMDDHHmm`，例如 `202512291337` 表示 2025年12月29日13时37分
- **latest 标签**：手动指定 `./build-and-push.sh latest`
- **版本标签**：可以指定任意标签，如 `v0.0.6`、`1.0.0` 等

## 使用镜像

### 拉取镜像

```bash
# 拉取时间戳标签的镜像（例如：202512291337）
docker pull youhuipu/node-hiprint-transit:202512291337

# 拉取 latest 标签
docker pull youhuipu/node-hiprint-transit:latest
```

### 运行容器

**重要：使用服务器证书**

在运行容器前，需要将服务器的 SSL 证书文件复制到 `/var/hiprint/` 目录：

```bash
# 创建目录
sudo mkdir -p /var/hiprint

# 复制证书文件（根据实际情况修改路径）
# 示例：从 Nginx 配置目录复制
sudo cp /etc/nginx/ssl/mall.sscchh.com.crt /var/hiprint/ssl.pem
sudo cp /etc/nginx/ssl/mall.sscchh.com.key /var/hiprint/ssl.key

# 或从 Let's Encrypt 复制
sudo cp /etc/letsencrypt/live/mall.sscchh.com/fullchain.pem /var/hiprint/ssl.pem
sudo cp /etc/letsencrypt/live/mall.sscchh.com/privkey.pem /var/hiprint/ssl.key

# 设置权限
sudo chmod 600 /var/hiprint/ssl.key
sudo chmod 644 /var/hiprint/ssl.pem
```

详细说明请查看 [SSL-CERTIFICATE.md](./SSL-CERTIFICATE.md)

然后运行容器：

```bash
docker run -d \
  -p 17521:17521 \
  -v /var/hiprint/config.json:/node-hiprint-transit/config.json \
  -v /var/hiprint/logs:/node-hiprint-transit/logs \
  -v /var/hiprint/ssl.key:/node-hiprint-transit/dist/src/ssl.key \
  -v /var/hiprint/ssl.pem:/node-hiprint-transit/dist/src/ssl.pem \
  --name node-hiprint-transit \
  youhuipu/node-hiprint-transit:latest
```

或使用 docker-compose：

```bash
docker-compose up -d
```

## 故障排除

### 1. 基础镜像拉取失败

如果遇到 403 错误，请先手动拉取：

```bash
export HTTP_PROXY=http://127.0.0.1:7897
export HTTPS_PROXY=http://127.0.0.1:7897
docker pull node:16-alpine
```

### 2. 推送失败

- 确保已登录 Docker Hub：`docker login`
- 检查网络连接
- 确认有推送权限

### 3. 构建失败

- 检查代理设置是否正确
- 确保所有依赖文件存在
- 查看构建日志中的错误信息

