# 故障排除指南

## 问题：访问 https://mall.sscchh.com:17521/ 一直转圈

### 原因分析

1. **应用只提供 WebSocket 服务**
   - 这个应用是一个 Socket.io 中转服务
   - 不提供传统的 HTTP 页面
   - 需要通过 Socket.io 客户端连接，不能直接用浏览器访问

2. **可能的问题**
   - 容器未正常启动
   - 端口映射配置错误
   - 防火墙/安全组未开放端口
   - SSL 证书配置错误

### 解决步骤

#### 1. 检查容器状态

```bash
# 查看容器状态
docker ps -a | grep print

# 查看容器日志
docker logs print

# 查看实时日志
docker logs -f print
```

#### 2. 检查端口映射

在宝塔面板中：
1. 进入 **Docker** → **容器**
2. 找到 `print` 容器
3. 点击 **容器详情** → **端口映射**
4. 确认端口映射：`17521:17521`

或使用命令行：
```bash
docker port print
```

#### 3. 检查防火墙和安全组

**服务器防火墙：**
```bash
# 检查防火墙状态
sudo ufw status

# 如果防火墙开启，需要开放端口
sudo ufw allow 17521/tcp
```

**云服务器安全组：**
- 登录云服务器控制台
- 进入安全组配置
- 添加入站规则：端口 17521，协议 TCP

#### 4. 测试端口连通性

```bash
# 在服务器上测试本地端口
curl -k https://localhost:17521

# 从外部测试（在另一台机器上）
curl -k https://mall.sscchh.com:17521
```

#### 5. 检查 SSL 证书

```bash
# 检查证书文件是否存在
ls -la /var/hiprint/ssl.*

# 检查证书内容
sudo openssl x509 -in /var/hiprint/ssl.pem -text -noout | head -20

# 检查私钥
sudo openssl rsa -in /var/hiprint/ssl.key -check
```

#### 6. 检查容器网络

```bash
# 检查容器网络配置
docker inspect print | grep -A 20 "NetworkSettings"

# 测试容器内部端口
docker exec print wget -O- https://localhost:17521
```

### 正确的使用方式

这个服务**不能**直接用浏览器访问，需要通过 Socket.io 客户端连接：

**前端代码示例：**
```javascript
import { hiprint } from 'vue-plugin-hiprint'

hiprint.init({
  host: 'https://mall.sscchh.com:17521',
  token: 'vue-plugin-hiprint'
})
```

### 健康检查

修复后，访问 `https://mall.sscchh.com:17521/` 应该能看到一个简单的状态页面，显示：
- 服务运行状态
- 版本信息
- 连接方式说明

如果还是转圈，检查：
1. 容器日志是否有错误
2. 端口是否正确映射
3. 防火墙是否开放

### 常见错误

#### 错误 1: 容器启动失败

**症状：** 容器状态为 `Exited`

**解决：**
```bash
# 查看错误日志
docker logs print

# 常见原因：
# - SSL 证书文件不存在或格式错误
# - 配置文件错误
# - 端口被占用
```

#### 错误 2: 连接被拒绝

**症状：** `Connection refused`

**解决：**
- 检查容器是否运行：`docker ps | grep print`
- 检查端口映射：`docker port print`
- 检查防火墙和安全组

#### 错误 3: SSL 证书错误

**症状：** `SSL certificate problem` 或 `certificate verify failed`

**解决：**
- 确保证书文件存在：`ls -la /var/hiprint/ssl.*`
- 检查证书格式：`openssl x509 -in /var/hiprint/ssl.pem -text`
- 确保证书和私钥匹配

### 调试命令

```bash
# 进入容器内部
docker exec -it print sh

# 查看进程
docker exec print ps aux

# 查看网络连接
docker exec print netstat -tlnp

# 测试服务
docker exec print wget -O- https://localhost:17521
```

