# Docker 部署说明

## 快速开始

### 只构建镜像（不推送，tag 自动使用当前时间）

```bash
cd admin_console
./docker-build.sh
# 镜像名称：mall_admin:202412282345 (示例，实际为当前时间)
```

### 构建并推送到 Docker Hub（tag 使用当前时间）

```bash
# 先登录 Docker Hub
docker login

# 构建并推送（tag 自动使用当前时间，如：202412282345）
./docker-build.sh --push
```

### 使用指定 tag

```bash
# 使用 latest 作为 tag
./docker-build.sh latest --push

# 使用自定义 tag
./docker-build.sh v1.0.0 --push
```

### 构建并推送到阿里云容器镜像服务

```bash
# 先登录阿里云
docker login --username=your-username registry.cn-hangzhou.aliyuncs.com

# 构建并推送（替换 your-namespace 为你的命名空间）
./docker-build.sh latest --push --registry=registry.cn-hangzhou.aliyuncs.com/your-namespace
```

### 构建并推送到腾讯云容器镜像服务

```bash
# 先登录腾讯云
docker login ccr.ccs.tencentyun.com

# 构建并推送（替换 your-namespace 为你的命名空间）
./docker-build.sh latest --push --registry=ccr.ccs.tencentyun.com/your-namespace
```

## 镜像名称和 Tag

- 镜像名称：`mall_admin`
- 默认 tag：使用当前时间的年月日时分格式（如：`202412282345`）
- 可以手动指定 tag：`./docker-build.sh latest`

## 详细步骤

1. **构建镜像**
   ```bash
   ./docker-build.sh
   ```
   脚本会自动：
   - 本地构建前端（`npm run build`）
   - 使用 Docker 打包成镜像（自动使用本地代理 7897 端口）
   - 镜像名称：`mall_admin:latest`

2. **运行容器**
   ```bash
   docker run -d \
     --name admin-console \
     -p 15173:5173 \
     --restart unless-stopped \
     youhuipu/mall_admin:latest
   ```
   
   **参数说明**：
   - `-p 15173:5173`：端口映射（宿主机:容器）
   - `--name admin-console`：容器名称
   - `--restart unless-stopped`：自动重启策略
   
   **注意**：容器内只提供静态文件服务，API 代理需要在宝塔 Nginx 中配置

3. **访问**
   - 本地访问：`http://localhost:15173`
   - 服务器访问：`http://your-server-ip:15173`
   - 或通过域名访问（需要在宝塔中配置反向代理）

## 推送到镜像仓库

### Docker Hub

```bash
# 登录
docker login

# 构建并推送
./docker-build.sh latest --push
```

### 阿里云容器镜像服务

```bash
# 登录（替换为你的用户名）
docker login --username=your-username registry.cn-hangzhou.aliyuncs.com

# 构建并推送（替换 your-namespace）
./docker-build.sh latest --push --registry=registry.cn-hangzhou.aliyuncs.com/your-namespace
```

### 腾讯云容器镜像服务

```bash
# 登录
docker login ccr.ccs.tencentyun.com

# 构建并推送（替换 your-namespace）
./docker-build.sh latest --push --registry=ccr.ccs.tencentyun.com/your-namespace
```

## 拉取镜像

```bash
# 从 Docker Hub 拉取（使用你推送的 tag）
docker pull youhuipu/mall_admin:202512282356

# 或者拉取最新版本（如果标记了 latest）
docker pull youhuipu/mall_admin:latest
```

## 导出镜像（如果不想使用镜像仓库）

```bash
docker save youhuipu/mall_admin:latest | gzip > mall_admin.tar.gz
```

## 在服务器上导入

```bash
gunzip -c mall_admin.tar.gz | docker load
```

## 更新部署

当有新版本时：

```bash
# 1. 停止并删除旧容器
docker stop admin-console
docker rm admin-console

# 2. 拉取新镜像
docker pull youhuipu/mall_admin:新的时间戳

# 3. 运行新容器
docker run -d \
  --name admin-console \
  -p 15173:5173 \
  --restart unless-stopped \
  youhuipu/mall_admin:新的时间戳
```

## 注意事项

### 构建相关
- 默认使用本地代理 `127.0.0.1:7897`，确保代理服务正在运行
- 如需修改代理端口，编辑 `Dockerfile` 中的 `ARG HTTP_PROXY` 和 `ARG HTTPS_PROXY`

### 容器运行
- **容器内只提供静态文件服务，不包含 API 代理配置**
- 运行容器时**不需要** `--add-host` 参数
- 端口映射：宿主机 `15173` → 容器 `5173`

### API 代理配置（重要）

**前端请求地址**：`/api_mall/mini/...`

**宝塔 Nginx 配置示例**：
```nginx
# 前端静态文件（代理到容器）
# 注意：proxy_pass 末尾的 / 很重要，确保路径正确转发
location /admin/ {
    proxy_pass http://127.0.0.1:15173/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Port $server_port;
    
    # 超时设置
    proxy_connect_timeout 30s;
    proxy_send_timeout 30s;
    proxy_read_timeout 30s;
    
    # 禁用缓冲（可选，有助于实时响应）
    proxy_buffering off;
    
    # 处理 WebSocket（如果需要）
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}

# API 代理（代理到后端服务）
# 注意：proxy_pass 末尾的 /api/ 很重要，确保路径正确转发
location /api_mall/ {
    proxy_pass http://127.0.0.1:8082/api/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # 超时设置
    proxy_connect_timeout 30s;
    proxy_send_timeout 30s;
    proxy_read_timeout 30s;
}
```

**说明**：
- 前端请求 `/api_mall/mini/admin/login`
- Nginx 会代理到 `http://127.0.0.1:8082/api/mini/admin/login`
- 确保 `proxy_pass` 末尾有 `/api/`，这样 `/api_mall/` 会被替换为 `/api/`

### 页面空白问题解决

**问题**：访问 `/admin/` 后页面空白，没有任何内容显示。

**原因**：通常是静态资源路径配置问题。

**解决方案**：

1. **确保 `vite.config.js` 中的 `base` 配置正确**：
   ```javascript
   base: '/admin/', // 必须与 Nginx 的 location /admin/ 匹配
   ```
   这确保所有静态资源（JS、CSS）使用正确的路径 `/admin/assets/...`

2. **检查浏览器控制台**：
   - 打开浏览器开发者工具（F12）
   - 查看 Console 和 Network 标签
   - 检查是否有 404 错误（资源加载失败）
   - 检查 JS、CSS 文件的路径是否正确（应该是 `/admin/assets/...`）

3. **如果修改了 `vite.config.js`，需要重新构建**：
   ```bash
   cd admin_console
   ./docker-build.sh --push
   ```
   然后在服务器上拉取新镜像并重新部署。

4. **验证静态资源路径**：
   - 访问 `https://mall.sscchh.com/admin/`
   - 在浏览器开发者工具的 Network 标签中，检查所有资源是否成功加载
   - 如果看到 404，检查资源路径是否正确

### 刷新页面 502 问题解决

**问题**：访问 `/admin/` 正常，但刷新页面后出现 502 错误。

**原因**：这是单页应用（SPA）的常见问题，刷新时浏览器请求 `/admin/some-route`，需要容器内的 Nginx 正确处理。

**解决方案**：

1. **确保 `proxy_pass` 末尾有 `/`**（你的配置已经正确）：
   ```nginx
   location /admin/ {
       proxy_pass http://127.0.0.1:15173/;  # 末尾必须有 /
   }
   ```
   这样 `/admin/dashboard` 会被转发到容器的 `/dashboard`，容器内的 Nginx 会通过 `try_files` 返回 `index.html`。

2. **容器内的 Nginx 已配置**：
   - `try_files $uri $uri/ /index.html;` - 处理所有路由
   - 如果修改了 `nginx.conf`，需要重新构建镜像

3. **如果还是不行，检查**：
   ```bash
   # 检查容器日志
   docker logs admin-console
   
   # 测试容器是否正常
   curl http://127.0.0.1:15173/
   curl http://127.0.0.1:15173/dashboard  # 应该返回 index.html
   
   # 重启 Nginx（在服务器上）
   nginx -s reload
   ```

4. **验证**：访问以下 URL 都应该正常：
   - `https://mall.sscchh.com/admin/` - 首页
   - `https://mall.sscchh.com/admin/dashboard` - 刷新后应该正常
   - `https://mall.sscchh.com/admin/users` - 刷新后应该正常

