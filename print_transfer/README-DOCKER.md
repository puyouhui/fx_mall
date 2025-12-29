# Docker 构建说明

## ⚠️ 重要：如果遇到 403 Forbidden 错误

Docker Desktop 的代理配置可能没有生效。请按以下步骤操作：

### 方法 1: 配置 Docker Desktop 代理并重启（推荐）

1. 打开 Docker Desktop
2. 进入 **Settings** → **Resources** → **Proxies**
3. 启用 **Manual proxy configuration**
4. 设置代理：
   - Web Server (HTTP): `http://127.0.0.1:7897`
   - Secure Web Server (HTTPS): `http://127.0.0.1:7897`
5. **重要：点击 Apply & Restart，完全重启 Docker Desktop**
6. 重启后，验证代理是否生效：
   ```bash
   docker info | grep -i proxy
   ```
   应该看到你的代理地址

7. 然后构建：
   ```bash
   docker build -t node-hiprint-transit:latest .
   ```

### 方法 2: 使用构建和推送脚本（推荐）

使用 `build-and-push.sh` 脚本，它会自动处理代理和镜像拉取：

```bash
# 构建并推送镜像（自动生成时间戳标签）
./build-and-push.sh

# 或指定标签
./build-and-push.sh latest
```

详细说明请查看 [BUILD.md](./BUILD.md)

### 方法 3: 先手动拉取镜像（如果方法1不行）

```bash
# 使用代理拉取镜像
export HTTP_PROXY=http://127.0.0.1:7897
export HTTPS_PROXY=http://127.0.0.1:7897
docker pull node:16-alpine

# 拉取成功后，再构建
docker build -t node-hiprint-transit:latest .
```

### 方法 4: 手动配置 Docker daemon.json

编辑或创建 `~/.docker/daemon.json`：
```json
{
  "proxies": {
    "http-proxy": "http://127.0.0.1:7897",
    "https-proxy": "http://127.0.0.1:7897",
    "no-proxy": "localhost,127.0.0.1"
  }
}
```

然后重启 Docker Desktop。

### 方法 5: 使用国内镜像源

如果代理不可用，可以使用国内镜像源：

```bash
docker build -f Dockerfile.mirror -t node-hiprint-transit:latest .
```

## 构建后查看镜像大小

```bash
docker images | grep node-hiprint-transit
```

## 运行容器

```bash
docker run -d \
  -p 17521:17521 \
  -v /var/hiprint/config.json:/node-hiprint-transit/config.json \
  -v /var/hiprint/logs:/node-hiprint-transit/logs \
  -v /var/hiprint/ssl.key:/node-hiprint-transit/dist/src/ssl.key \
  -v /var/hiprint/ssl.pem:/node-hiprint-transit/dist/src/ssl.pem \
  --name node-hiprint-transit \
  node-hiprint-transit:latest
```

或使用 docker-compose：
```bash
docker-compose up -d
```

