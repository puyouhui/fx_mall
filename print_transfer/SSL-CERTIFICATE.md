# 使用服务器 SSL 证书配置说明

## 问题说明

自签名证书会导致浏览器显示"您的连接不是私密连接"的警告。使用服务器正式证书可以解决这个问题。

## 证书文件要求

应用需要两个证书文件：
- `ssl.key` - 私钥文件
- `ssl.pem` - 证书文件（通常是 .crt 或 .pem 格式）

## 获取服务器证书

### 方法 1: 从 Nginx/Apache 配置目录获取

如果你的服务器使用 Nginx 或 Apache，证书通常在以下位置：

**Nginx:**
```bash
# 证书文件通常在
/etc/nginx/ssl/your-domain.com.crt
/etc/nginx/ssl/your-domain.com.key

# 或者
/etc/letsencrypt/live/your-domain.com/fullchain.pem
/etc/letsencrypt/live/your-domain.com/privkey.pem
```

**Apache:**
```bash
/etc/apache2/ssl/your-domain.com.crt
/etc/apache2/ssl/your-domain.com.key
```

### 方法 2: 从宝塔面板获取

1. 登录宝塔面板
2. 进入 **网站** → 选择你的网站 → **SSL**
3. 下载证书文件：
   - 私钥（.key 文件）
   - 证书（.crt 或 .pem 文件）

### 方法 3: Let's Encrypt 证书

如果使用 Let's Encrypt：
```bash
# 证书文件位置
/etc/letsencrypt/live/mall.sscchh.com/fullchain.pem  # 证书链
/etc/letsencrypt/live/mall.sscchh.com/privkey.pem   # 私钥
```

## 配置步骤

### 1. 准备证书文件

将证书文件复制到服务器上的指定目录：

```bash
# 创建目录
sudo mkdir -p /var/hiprint

# 复制证书文件（根据你的实际情况修改路径）
# 方式 1: 如果证书是 .crt 格式，需要转换为 .pem
sudo cp /etc/nginx/ssl/mall.sscchh.com.crt /var/hiprint/ssl.pem
sudo cp /etc/nginx/ssl/mall.sscchh.com.key /var/hiprint/ssl.key

# 方式 2: 如果证书已经是 .pem 格式
sudo cp /etc/letsencrypt/live/mall.sscchh.com/fullchain.pem /var/hiprint/ssl.pem
sudo cp /etc/letsencrypt/live/mall.sscchh.com/privkey.pem /var/hiprint/ssl.key

# 设置权限
sudo chmod 600 /var/hiprint/ssl.key
sudo chmod 644 /var/hiprint/ssl.pem
```

### 2. 证书格式转换（如果需要）

如果证书是 `.crt` 格式，需要转换为 `.pem`：

```bash
# .crt 转 .pem（通常不需要，直接重命名即可）
sudo cp /path/to/certificate.crt /var/hiprint/ssl.pem

# 或者使用 openssl 转换
sudo openssl x509 -in /path/to/certificate.crt -out /var/hiprint/ssl.pem -outform PEM
```

### 3. 验证证书文件

```bash
# 检查证书内容
sudo openssl x509 -in /var/hiprint/ssl.pem -text -noout

# 检查私钥
sudo openssl rsa -in /var/hiprint/ssl.key -check

# 验证证书和私钥是否匹配
sudo openssl x509 -noout -modulus -in /var/hiprint/ssl.pem | openssl md5
sudo openssl rsa -noout -modulus -in /var/hiprint/ssl.key | openssl md5
# 两个 MD5 值应该相同
```

### 4. 使用 Docker Compose 运行

确保 `docker-compose.yml` 中的证书挂载路径正确：

```yaml
volumes:
  - /var/hiprint/ssl.key:/node-hiprint-transit/dist/src/ssl.key
  - /var/hiprint/ssl.pem:/node-hiprint-transit/dist/src/ssl.pem
```

然后启动容器：

```bash
docker-compose up -d
```

### 5. 使用 Docker 命令运行

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

## 证书文件说明

### ssl.pem（证书文件）

可以是以下格式之一：
- 单个证书文件（.crt 或 .pem）
- 证书链文件（包含中间证书）
- Let's Encrypt 的 `fullchain.pem`（推荐，包含完整证书链）

### ssl.key（私钥文件）

- 必须是私钥文件（.key 或 .pem）
- 确保私钥文件权限为 600（仅所有者可读写）

## 常见问题

### Q: 证书和私钥不匹配怎么办？

A: 确保使用同一域名的证书和私钥文件。

### Q: 证书过期了怎么办？

A: 更新证书文件后重启容器：
```bash
# 更新证书文件
sudo cp /path/to/new-cert.pem /var/hiprint/ssl.pem
sudo cp /path/to/new-key.key /var/hiprint/ssl.key

# 重启容器
docker restart node-hiprint-transit
```

### Q: 如何自动更新 Let's Encrypt 证书？

A: 可以创建一个脚本定期更新证书并重启容器，或者使用 certbot 的钩子脚本。

### Q: 证书包含多个文件怎么办？

A: 如果证书包含多个文件（如证书链），将它们合并到一个 .pem 文件中：
```bash
cat certificate.crt intermediate.crt root.crt > /var/hiprint/ssl.pem
```

## 验证配置

配置完成后，访问 `https://mall.sscchh.com:17521` 应该不再显示安全警告。


