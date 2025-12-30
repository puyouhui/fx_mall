class Config {
  // 这里直接复用员工端的配置方式，保持同一套后端环境

  // 真机调试（局域网 IP）
  static const String devBaseUrl = 'https://mall.sscchh.com';

  // Android 模拟器（访问宿主机）
  static const String emulatorBaseUrl = 'http://10.0.2.2:8082';

  // 生产环境
  static const String prodBaseUrl = 'https://mall.sscchh.com';

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

  // API 路径前缀，与员工端保持一致
  static const String apiPrefix = '/api/mini';

  // 完整的 API 基础 URL
  static String get apiBaseUrl => '$baseUrl$apiPrefix';
}
