class Config {
  // ============================================
  // 配置说明：
  // 1. Android 模拟器：使用 http://10.0.2.2:8082
  // 2. 真机调试：使用电脑的局域网 IP，例如 http://192.168.2.196:8082
  // 3. 生产环境：使用正式域名
  // 4. 通过 --dart-define=APP_ENV=xxx 自动切换
  //    - APP_ENV=emulator  -> 模拟器
  //    - APP_ENV=device    -> 真机调试
  //    - APP_ENV=prod      -> 生产
  // ============================================

  // 真机调试（局域网 IP）
  static const String devBaseUrl = 'http://192.168.2.207:8082';

  // Android 模拟器（访问宿主机）
  static const String emulatorBaseUrl = 'http://10.0.2.2:8082';

  // 生产环境
  static const String prodBaseUrl = 'https://your-production-server.com';

  // 编译期环境变量：APP_ENV
  // 不传时默认当成 emulator
  static const String _env = String.fromEnvironment(
    'APP_ENV',
    defaultValue: 'emulator',
  );

  // 当前使用的 BASE_URL（对外只用这个）
  static String get baseUrl {
    switch (_env) {
      case 'device':
        return devBaseUrl;
      case 'prod':
        return prodBaseUrl;
      case 'emulator':
      default:
        return emulatorBaseUrl;
    }
  }

  // API 路径前缀
  static const String apiPrefix = '/api/mini';

  // 完整的 API 基础 URL
  static String get apiBaseUrl => '$baseUrl$apiPrefix';
}
