# App 打包配置说明

## 一、生产环境配置

两个App的生产环境API地址已配置为：`https://mall.sscchh.com/api/mini`

### 配置文件位置
- **员工端**: `employees_app/lib/utils/config.dart`
- **配送端**: `distribution_app/lib/utils/config.dart`

### 环境切换说明
App支持三种环境，通过编译时参数 `APP_ENV` 切换：

- `APP_ENV=emulator`：模拟器环境（默认）
  - API地址：`http://10.0.2.2:8082/api/mini`
- `APP_ENV=device`：真机调试环境
  - 员工端：`http://192.168.2.207:8082/api/mini`
  - 配送端：`http://192.168.31.110:8082/api/mini`
- `APP_ENV=prod`：生产环境
  - API地址：`https://mall.sscchh.com/api/mini`

## 二、打包生产环境APK

### Windows 打包脚本

#### 员工端App
```bash
cd employees_app
build-prod.bat
```

#### 配送端App
```bash
cd distribution_app
build-prod.bat
```

打包脚本会自动：
1. 清理之前的构建文件
2. 获取依赖包
3. 构建生产环境APK（使用 `APP_ENV=prod`）

### 手动打包命令

如果不想使用脚本，可以手动执行：

**员工端**:
```bash
cd employees_app
flutter clean
flutter pub get
flutter build apk --release --dart-define=APP_ENV=prod
```

**配送端**:
```bash
cd distribution_app
flutter clean
flutter pub get
flutter build apk --release --dart-define=APP_ENV=prod
```

### 打包输出

APK文件位置：`build/app/outputs/flutter-apk/app-release.apk`

## 三、后端接口说明

### API基础地址
- 生产环境：`https://mall.sscchh.com/api/mini`
- Nginx配置路径：`location /api/mini/`

### WebSocket配置
- 配送员位置上报：`wss://mall.sscchh.com/api/mini/employee/location/ws`
- 管理后台位置查看：`wss://mall.sscchh.com/api/mini/admin/employee-locations/ws`

WebSocket URL通过接口 `/api/mini/employee/websocket-config` 获取，后端会根据请求协议（HTTP/HTTPS）自动返回对应的 ws:// 或 wss:// 格式。

## 四、注意事项

1. **打包前确认**：确保后端服务已部署并正常运行在 `https://mall.sscchh.com`
2. **网络连接**：App需要能够访问HTTPS接口，确保设备网络正常
3. **证书验证**：生产环境使用HTTPS，确保服务器SSL证书有效
4. **WebSocket连接**：配送端App会自动通过HTTPS升级为WSS连接

## 五、测试建议

打包完成后，建议进行以下测试：
1. 登录功能测试
2. 接口请求测试（确保API地址正确）
3. WebSocket连接测试（配送端位置上报功能）
4. 地图功能测试（如果使用了地图服务）


