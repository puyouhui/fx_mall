import 'package:geolocator/geolocator.dart';
import 'package:permission_handler/permission_handler.dart';

/// 定位服务工具类
class LocationService {
  // 缓存的位置信息
  static Position? _cachedPosition;
  static DateTime? _cachedPositionTime;
  static const Duration _cacheValidDuration = Duration(minutes: 5); // 缓存有效期5分钟

  /// 获取缓存的位置（如果可用且未过期）
  static Position? getCachedPosition() {
    if (_cachedPosition != null && _cachedPositionTime != null) {
      final age = DateTime.now().difference(_cachedPositionTime!);
      if (age < _cacheValidDuration) {
        print('[LocationService] 使用缓存的位置，缓存时间: ${age.inSeconds}秒前');
        return _cachedPosition;
      } else {
        print('[LocationService] 缓存的位置已过期，清除缓存');
        _cachedPosition = null;
        _cachedPositionTime = null;
      }
    }
    return null;
  }

  /// 设置缓存的位置
  static void setCachedPosition(Position position) {
    _cachedPosition = position;
    _cachedPositionTime = DateTime.now();
    print('[LocationService] 位置已缓存: ${position.latitude}, ${position.longitude}');
  }

  /// 清除缓存的位置
  static void clearCache() {
    _cachedPosition = null;
    _cachedPositionTime = null;
    print('[LocationService] 位置缓存已清除');
  }
  /// 检查定位服务是否启用
  static Future<bool> checkLocationServiceEnabled() async {
    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      print('[LocationService] 定位服务状态: $serviceEnabled');
      return serviceEnabled;
    } catch (e) {
      print('[LocationService] 检查定位服务状态异常: $e');
      return false;
    }
  }

  /// 打开定位服务设置页面
  static Future<bool> openLocationSettings() async {
    try {
      return await Geolocator.openLocationSettings();
    } catch (e) {
      print('[LocationService] 打开定位设置失败: $e');
      return false;
    }
  }

  /// 打开应用设置页面（用于手动授予权限）
  static Future<bool> openAppSettingsPage() async {
    try {
      // 使用 permission_handler 打开应用设置
      return await openAppSettings();
    } catch (e) {
      print('[LocationService] 打开应用设置失败: $e');
      // 如果失败，尝试使用 geolocator 的方法
      try {
        return await Geolocator.openAppSettings();
      } catch (e2) {
        print('[LocationService] 使用 geolocator 打开应用设置也失败: $e2');
        return false;
      }
    }
  }

  /// 检查并请求定位权限（使用 permission_handler）
  /// 针对小米手机（MIUI）优化：如果权限请求没有弹出对话框，引导用户到设置页面
  static Future<bool> checkAndRequestPermission() async {
    try {
      // 使用 permission_handler 检查权限状态
      PermissionStatus status = await Permission.location.status;
      print('[LocationService] 当前权限状态: $status');

      // 如果权限已授予，直接返回
      if (status.isGranted) {
        print('[LocationService] 定位权限已授予');
        return true;
      }

      // 如果权限被永久拒绝
      if (status.isPermanentlyDenied) {
        print('[LocationService] 定位权限被永久拒绝，需要到设置中手动开启');
        return false;
      }

      // 如果权限被拒绝或未授予，请求权限
      if (status.isDenied || status.isLimited) {
        print('[LocationService] 权限未授予，开始请求定位权限...');

        // 请求权限（这会显示系统权限对话框）
        // 注意：在小米手机上，这个对话框可能不会弹出
        status = await Permission.location.request();
        print('[LocationService] 权限请求结果: $status');

        // 如果权限仍然是拒绝状态（可能是小米手机没有弹出对话框）
        if (status.isDenied) {
          print('[LocationService] 权限请求后仍为拒绝状态');
          print('[LocationService] 可能是小米手机（MIUI）未弹出权限对话框');
          // 再次检查权限状态，确保准确性
          await Future.delayed(const Duration(milliseconds: 500));
          status = await Permission.location.status;
          print('[LocationService] 延迟后再次检查权限状态: $status');

          if (status.isDenied) {
            print('[LocationService] 用户拒绝了定位权限或系统未弹出对话框');
            return false;
          }
        }

        // 如果权限被永久拒绝
        if (status.isPermanentlyDenied) {
          print('[LocationService] 定位权限被永久拒绝，需要到设置中手动开启');
          return false;
        }
      }

      // 检查最终权限状态
      if (status.isGranted || status.isLimited) {
        print('[LocationService] 定位权限已授予: $status');
        return true;
      }

      print('[LocationService] 未知的权限状态: $status');
      return false;
    } catch (e) {
      print('[LocationService] 权限检查异常: $e');
      print('[LocationService] 异常堆栈: ${StackTrace.current}');
      return false;
    }
  }

  /// 获取当前位置（即使定位服务未启用也尝试获取，某些设备可能仍能获取到网络定位）
  static Future<Position?> getCurrentLocationDirect() async {
    print('[LocationService] 直接获取定位（不检查服务状态，优先网络定位）...');

    // 检查并请求权限
    bool hasPermission = await checkAndRequestPermission();
    if (!hasPermission) {
      print('[LocationService] 没有定位权限，无法获取位置');
      return null;
    }

    // 尝试从低精度到最低精度，每种精度重试2次
    final accuracyLevels = [
      LocationAccuracy.low,
      LocationAccuracy.lowest,
    ];

    for (int level = 0; level < accuracyLevels.length; level++) {
      final accuracy = accuracyLevels[level];
      
      // 每种精度重试2次
      for (int retry = 0; retry < 2; retry++) {
        try {
          print('[LocationService] 尝试网络定位 - 精度: $accuracy, 重试: ${retry + 1}/2');
          
          // 尝试获取当前位置，即使定位服务可能未启用
          // 使用低精度，优先使用网络定位（NETWORK_PROVIDER），在室内也能工作
          // 强制使用 Android 原生 LocationManager，避免依赖 Google Location Service
          Position position = await Geolocator.getCurrentPosition(
            desiredAccuracy: accuracy,
            timeLimit: const Duration(seconds: 15), // 网络定位通常很快，15秒足够
            forceAndroidLocationManager: true, // 强制使用 Android 原生 LocationManager
          );

          print(
            '[LocationService] 直接定位成功: ${position.latitude}, ${position.longitude}, 精度: ${position.accuracy}米',
          );
          // 缓存位置信息
          setCachedPosition(position);
          return position;
        } catch (e) {
          final errorStr = e.toString();
          print('[LocationService] 网络定位失败 (精度: $accuracy, 重试: ${retry + 1}/2): $e');
          
          // 如果是权限错误，直接返回
          if (errorStr.contains('permission') || 
              errorStr.contains('Permission') ||
              errorStr.contains('权限')) {
            print('[LocationService] 权限错误，停止尝试');
            return null;
          }
          
          // 如果是最后一次重试，继续下一级精度
          if (retry == 1) {
            print('[LocationService] 当前精度级别失败，降级到下一级精度');
            break;
          }
          
          // 等待一下再重试
          await Future.delayed(const Duration(milliseconds: 500));
        }
      }
    }

    print('[LocationService] 网络定位失败，可能是定位服务未启用或网络不可用');
    return null;
  }

  /// 获取当前位置（带多级降级策略和重试机制）
  static Future<Position?> getCurrentLocation() async {
    // 先检查并请求权限
    bool hasPermission = await checkAndRequestPermission();
    if (!hasPermission) {
      print('[LocationService] 没有定位权限，无法获取位置');
      // 再次检查 Geolocator 的权限状态
      final geolocatorPermission = await Geolocator.checkPermission();
      print('[LocationService] Geolocator 权限状态: $geolocatorPermission');
      return null;
    }

    // 检查定位服务是否启用
    bool serviceEnabled = await checkLocationServiceEnabled();
    print('[LocationService] 定位服务状态: $serviceEnabled');

    // 如果定位服务未启用，尝试直接获取（某些设备可能仍能获取到网络定位）
    if (!serviceEnabled) {
      print('[LocationService] 定位服务未启用，尝试直接获取位置（网络定位）...');
      final position = await getCurrentLocationDirect();
      if (position != null) {
        print('[LocationService] 即使定位服务未启用，仍成功获取到位置（可能是网络定位）');
        return position;
      }
      print('[LocationService] 定位服务未启用且无法获取位置');
      // 即使定位服务未启用，也继续尝试多级降级策略
    }

    // 多级降级策略：优先使用网络定位（低精度），在中国更可靠
    // 从低精度到高精度，优先网络定位，GPS作为备选
    final accuracyLevels = [
      LocationAccuracy.low,        // 优先：网络定位（WiFi + 基站）
      LocationAccuracy.lowest,     // 备选：最低精度网络定位
      LocationAccuracy.medium,     // 备选：中等精度（GPS + 网络）
      LocationAccuracy.high,       // 最后：高精度GPS（在中国可能失败）
    ];

    final timeLimits = [
      const Duration(seconds: 15), // 网络定位通常很快，15秒足够
      const Duration(seconds: 10), // 最低精度10秒
      const Duration(seconds: 20), // 中等精度20秒
      const Duration(seconds: 30), // 高精度GPS给30秒（可能超时）
    ];

    for (int level = 0; level < accuracyLevels.length; level++) {
      final accuracy = accuracyLevels[level];
      final timeLimit = timeLimits[level];
      
      // 每种精度重试2次
      for (int retry = 0; retry < 2; retry++) {
        try {
          print('[LocationService] 尝试获取定位 - 精度: $accuracy, 超时: ${timeLimit.inSeconds}秒, 重试: ${retry + 1}/2');
          
          Position position = await Geolocator.getCurrentPosition(
            desiredAccuracy: accuracy,
            timeLimit: timeLimit,
            forceAndroidLocationManager: true, // 强制使用 Android 原生 LocationManager
          );

          print(
            '[LocationService] 定位成功: ${position.latitude}, ${position.longitude}, 精度: ${position.accuracy}米, 使用精度级别: $accuracy',
          );
          // 缓存位置信息
          setCachedPosition(position);
          return position;
        } catch (e) {
          final errorStr = e.toString();
          print('[LocationService] 定位失败 (精度: $accuracy, 重试: ${retry + 1}/2): $e');
          
          // 如果是权限错误，直接返回，不需要继续尝试
          if (errorStr.contains('permission') || 
              errorStr.contains('Permission') ||
              errorStr.contains('权限')) {
            print('[LocationService] 权限错误，停止尝试');
            return null;
          }
          
          // 如果是超时错误，继续下一级精度或重试
          if (errorStr.contains('timeout') || 
              errorStr.contains('TimeoutException') ||
              errorStr.contains('超时')) {
            print('[LocationService] 定位超时，继续尝试...');
            // 如果是最后一次重试，继续下一级精度
            if (retry == 1) {
              print('[LocationService] 当前精度级别失败，降级到下一级精度');
              break; // 跳出重试循环，继续下一级精度
            }
            // 否则等待一下再重试
            await Future.delayed(const Duration(milliseconds: 500));
            continue;
          }
          
          // 其他错误，也尝试下一级精度
          print('[LocationService] 其他错误，降级到下一级精度');
          break; // 跳出重试循环，继续下一级精度
        }
      }
    }

    print('[LocationService] 所有精度级别都失败，定位失败');
    return null;
  }

  /// 获取位置信息字符串
  static String formatLocation(Position? position) {
    if (position == null) {
      return '定位失败';
    }

    return '纬度: ${position.latitude.toStringAsFixed(6)}\n'
        '经度: ${position.longitude.toStringAsFixed(6)}\n'
        '精度: ${position.accuracy.toStringAsFixed(2)}米';
  }
}
