# 配送端App WebSocket连接调试指南

## 一、调试方法

### 1. 查看日志输出

配送端App的WebSocket连接过程会输出详细的日志，包括：

#### 启动服务时的日志
```
[LocationReportService] 启动位置上报服务
[LocationReportService] ========== 获取WebSocket配置 ==========
[LocationReportService] API地址: https://mall.sscchh.com/api/mini/employee/websocket-config
[LocationReportService] API响应状态码: 200
[LocationReportService] API返回数据: {...}
[LocationReportService] ✓ 获取WebSocket URL成功: wss://mall.sscchh.com/api/mini/employee/location/ws
[LocationReportService] ========== WebSocket连接信息 ==========
[LocationReportService] URL: wss://mall.sscchh.com/api/mini/employee/location/ws?token=...
[LocationReportService] Token存在: true
[LocationReportService] Token长度: xxx
[LocationReportService] ========================================
[LocationReportService] 正在创建WebSocket通道...
[LocationReportService] WebSocket通道已创建
[LocationReportService] ✓ WebSocket连接建立成功，开始监听消息
```

#### 连接成功后的消息
```
[LocationReportService] 收到WebSocket消息: {"type":"location_received"}
[LocationReportService] 消息类型: location_received
[LocationReportService] ✓ 位置上报成功（服务器确认）
```

#### 连接失败时的日志
```
[LocationReportService] ✗ WebSocket流错误: ...
[LocationReportService] 错误类型: ...
[LocationReportService] 堆栈: ...
[LocationReportService] 将在重连延迟后尝试重新连接...
```

### 2. 使用Android Studio或VS Code查看日志

#### Android Studio
1. 连接设备或启动模拟器
2. 在底部选择 `Logcat`
3. 搜索 `LocationReportService` 查看相关日志

#### VS Code（Flutter插件）
1. 按 `F5` 启动调试
2. 在 `Debug Console` 中查看输出
3. 或者在终端中运行 `flutter run` 查看日志

### 3. 使用adb命令查看日志（Android）

```bash
# 查看所有日志
adb logcat

# 只查看包含 LocationReportService 的日志
adb logcat | grep LocationReportService

# 保存日志到文件
adb logcat | grep LocationReportService > websocket_debug.log
```

## 二、常见问题排查

### 问题1: 无法获取WebSocket URL

**症状：**
```
[LocationReportService] ✗ 获取WebSocket URL失败: ...
[LocationReportService] 使用默认URL构建连接
```

**原因：**
- API接口 `/api/mini/employee/websocket-config` 调用失败
- 网络连接问题
- 后端服务未正常运行

**解决方法：**
1. 检查API地址是否正确（查看日志中的 `API地址`）
2. 检查网络连接是否正常
3. 检查后端服务是否运行在 `https://mall.sscchh.com`
4. 使用浏览器或Postman测试API接口：
   ```
   GET https://mall.sscchh.com/api/mini/employee/websocket-config
   ```
   应该返回：
   ```json
   {
     "code": 200,
     "message": "获取成功",
     "data": {
       "employee_location_url": "wss://mall.sscchh.com/api/mini/employee/location/ws",
       "admin_location_url": "wss://mall.sscchh.com/api/mini/admin/employee-locations/ws"
     }
   }
   ```

### 问题2: WebSocket连接失败

**症状：**
```
[LocationReportService] ✗ WebSocket流错误: ...
[LocationReportService] WebSocket连接已关闭（onDone）
```

**可能的原因和解决方法：**

#### a) 网络问题
- 检查设备网络连接
- 确保能够访问 `https://mall.sscchh.com`
- 检查防火墙是否阻止WebSocket连接

#### b) SSL证书问题（生产环境）
- 确保服务器SSL证书有效
- 如果使用自签名证书，需要在Android中添加证书信任（不建议）

#### c) Token无效
- 检查Token是否存在：`[LocationReportService] Token存在: true`
- 检查Token是否过期，尝试重新登录
- 检查URL中的Token是否正确传递

#### d) Nginx配置问题
- 检查Nginx是否支持WebSocket升级
- 确保Nginx配置包含：
  ```nginx
  proxy_http_version 1.1;
  proxy_set_header Upgrade $http_upgrade;
  proxy_set_header Connection "upgrade";
  proxy_connect_timeout 7d;
  proxy_send_timeout 7d;
  proxy_read_timeout 7d;
  ```

### 问题3: 连接建立但收不到消息

**症状：**
```
[LocationReportService] ✓ WebSocket连接建立成功，开始监听消息
[LocationReportService] 心跳检测：超过60秒未收到消息，尝试重连
```

**原因：**
- 服务器未发送确认消息
- 网络不稳定导致消息丢失
- 后端处理位置上报的代码有问题

**解决方法：**
1. 检查后端WebSocket处理代码是否正常
2. 检查后端日志，看是否收到位置上报数据
3. 检查网络稳定性

### 问题4: 调试环境连接正常，生产环境连接失败

**可能原因：**

#### a) URL配置错误
- 检查 `Config.baseUrl` 是否正确设置为生产环境URL
- 检查是否使用了正确的环境变量 `APP_ENV=prod` 打包

#### b) 生产环境Nginx配置
- 确保生产环境Nginx配置了WebSocket支持
- 检查Nginx日志是否有502或其他错误

#### c) 证书问题
- 生产环境使用HTTPS，确保SSL证书有效
- 检查浏览器访问 `https://mall.sscchh.com` 是否正常

## 三、测试WebSocket连接

### 1. 使用浏览器测试（需要先登录获取Token）

```javascript
// 在浏览器控制台执行
const token = 'YOUR_TOKEN_HERE'; // 替换为实际的token
const ws = new WebSocket(`wss://mall.sscchh.com/api/mini/employee/location/ws?token=${token}`);

ws.onopen = () => {
  console.log('WebSocket连接成功');
  // 发送测试位置
  ws.send(JSON.stringify({
    type: 'location',
    latitude: 39.9042,
    longitude: 116.4074,
    accuracy: 10
  }));
};

ws.onmessage = (event) => {
  console.log('收到消息:', event.data);
};

ws.onerror = (error) => {
  console.error('WebSocket错误:', error);
};

ws.onclose = (event) => {
  console.log('WebSocket关闭:', event.code, event.reason);
};
```

### 2. 使用wscat工具测试（需要Node.js）

```bash
# 安装wscat
npm install -g wscat

# 测试连接（替换YOUR_TOKEN为实际token）
wscat -c "wss://mall.sscchh.com/api/mini/employee/location/ws?token=YOUR_TOKEN"
```

### 3. 检查后端日志

在后端服务器上查看日志，确认是否收到WebSocket连接请求：

```bash
# 查看Go后端日志
tail -f /path/to/backend/logs/app.log | grep -i websocket
```

## 四、环境配置说明

### 调试环境
- **API地址**: `http://192.168.31.110:8082/api/mini`（真机调试）或 `http://10.0.2.2:8082/api/mini`（模拟器）
- **WebSocket**: `ws://192.168.31.110:8082/api/mini/employee/location/ws`（真机）或 `ws://10.0.2.2:8082/api/mini/employee/location/ws`（模拟器）
- **打包方式**: `flutter build apk --release --dart-define=APP_ENV=device`

### 生产环境
- **API地址**: `https://mall.sscchh.com/api/mini`
- **WebSocket**: `wss://mall.sscchh.com/api/mini/employee/location/ws`
- **打包方式**: `flutter build apk --release --dart-define=APP_ENV=prod` 或使用 `build-prod.bat`

## 五、调试检查清单

在报告WebSocket连接问题前，请检查以下项目：

- [ ] 查看App日志，确认连接过程
- [ ] 检查Token是否存在且有效
- [ ] 检查网络连接是否正常
- [ ] 检查API接口 `/api/mini/employee/websocket-config` 是否可访问
- [ ] 检查WebSocket URL是否正确（ws:// 或 wss://）
- [ ] 检查后端服务是否正常运行
- [ ] 检查Nginx配置是否支持WebSocket
- [ ] 检查生产环境SSL证书是否有效
- [ ] 检查防火墙是否阻止WebSocket连接
- [ ] 使用浏览器或wscat工具测试WebSocket连接

## 六、联系支持

如果以上方法都无法解决问题，请提供以下信息：

1. 完整的日志输出（从启动到连接失败）
2. 使用的环境（调试/生产）
3. 网络环境（WiFi/4G/5G）
4. 设备信息（Android版本、型号）
5. 后端日志（如果可访问）

