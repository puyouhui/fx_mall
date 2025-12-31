# 供应商后台 Docker 部署说明

## 快速开始

### 1. 构建镜像

```bash
cd supplier_console
docker-build.bat
```

### 2. 推送镜像（可选）

```bash
docker-build.bat --push
```

### 3. 在服务器上部署

```bash
# 拉取镜像
docker pull youhuipu/mall_supplier:时间戳

# 运行容器
docker run -d \
  --name supplier-console \
  -p 15174:5174 \
  --restart unless-stopped \
  youhuipu/mall_supplier:时间戳
```

## 注意事项

### 构建相关
- 默认使用本地代理 `127.0.0.1:7897`，确保代理服务正在运行
- 如需修改代理端口，编辑 `Dockerfile` 中的 `ARG HTTP_PROXY` 和 `ARG HTTPS_PROXY`

### 容器运行
- **容器内只提供静态文件服务，不包含 API 代理配置**
- 运行容器时**不需要** `--add-host` 参数
- 端口映射：宿主机 `15174` → 容器 `5174`

### API 代理配置（重要）

**前端请求地址**：`/api_mall/mini/supplier/...`

**宝塔 Nginx 配置示例**：
```nginx
# 前端静态文件（代理到容器）
# 注意：proxy_pass 末尾的 / 很重要，确保路径正确转发
location /supplier/ {
    proxy_pass http://127.0.0.1:15174/;
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
- 前端请求 `/api_mall/mini/supplier/login`
- Nginx 会代理到 `http://127.0.0.1:8082/api/mini/supplier/login`
- 确保 `proxy_pass` 末尾有 `/api/`，这样 `/api_mall/` 会被替换为 `/api/`

### 页面空白问题解决

**问题**：访问 `/supplier/` 后页面空白，没有任何内容显示。

**原因**：通常是静态资源路径配置问题。

**解决方案**：

1. **确保 `vite.config.js` 中的 `base` 配置正确**：
   ```javascript
   base: '/supplier/', // 必须与 Nginx 的 location /supplier/ 匹配
   ```
   这确保所有静态资源（JS、CSS）使用正确的路径 `/supplier/assets/...`

2. **检查浏览器控制台**：
   - 打开浏览器开发者工具（F12）
   - 查看 Console 和 Network 标签
   - 检查是否有 404 错误（资源加载失败）
   - 检查 JS、CSS 文件的路径是否正确（应该是 `/supplier/assets/...`）

3. **如果修改了 `vite.config.js`，需要重新构建**：
   ```bash
   cd supplier_console
   docker-build.bat --push
   ```
   然后在服务器上拉取新镜像并重新部署。

4. **验证静态资源路径**：
   - 访问 `https://mall.sscchh.com/supplier/`
   - 在浏览器开发者工具的 Network 标签中，检查所有资源是否成功加载
   - 如果看到 404，检查资源路径是否正确

### 刷新页面 502 问题解决

**问题**：访问 `/supplier/` 正常，但刷新页面后出现 502 错误。

**原因**：这是单页应用（SPA）的常见问题，刷新时浏览器请求 `/supplier/some-route`，需要容器内的 Nginx 正确处理。

**解决方案**：

1. **确保 `proxy_pass` 末尾有 `/`**（你的配置应该正确）：
   ```nginx
   location /supplier/ {
       proxy_pass http://127.0.0.1:15174/;  # 末尾必须有 /
   }
   ```
   这样 `/supplier/dashboard` 会被转发到容器的 `/dashboard`，容器内的 Nginx 会通过 `try_files` 返回 `index.html`。

2. **容器内的 Nginx 已配置**：
   - `try_files $uri $uri/ /index.html;` - 处理所有路由
   - 如果修改了 `nginx.conf`，需要重新构建镜像

3. **如果还是不行，检查**：
   ```bash
   # 检查容器日志
   docker logs supplier-console
   
   # 测试容器是否正常
   curl http://127.0.0.1:15174/
   curl http://127.0.0.1:15174/dashboard  # 应该返回 index.html
   
   # 重启 Nginx（在服务器上）
   nginx -s reload
   ```

4. **验证**：访问以下 URL 都应该正常：
   - `https://mall.sscchh.com/supplier/` - 首页
   - `https://mall.sscchh.com/supplier/dashboard` - 刷新后应该正常
   - `https://mall.sscchh.com/supplier/orders` - 刷新后应该正常

### API 502 错误排查

**问题**：API 请求返回 502 Bad Gateway。

**可能原因和解决方案**：

1. **后端服务未运行**：
   ```bash
   # 检查后端服务是否运行
   curl http://127.0.0.1:8082/api/mini/supplier/login
   # 或者
   ps aux | grep go_backend
   ```

2. **Nginx 配置错误**：
   - 检查 `proxy_pass` 是否正确指向后端服务
   - 确保 `proxy_pass` 末尾有 `/api/`
   - 检查后端服务端口是否为 `8082`

3. **路径不匹配**：
   - 前端请求：`/api_mall/mini/supplier/login`
   - Nginx 应转发到：`http://127.0.0.1:8082/api/mini/supplier/login`
   - 确保 Nginx 配置中的 `proxy_pass` 为 `http://127.0.0.1:8082/api/`

4. **检查 Nginx 错误日志**：
   ```bash
   # 查看 Nginx 错误日志
   tail -f /var/log/nginx/error.log
   # 或宝塔面板的日志路径
   tail -f /www/wwwlogs/域名_error.log
   ```

