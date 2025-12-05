import 'package:geolocator/geolocator.dart';
import 'package:permission_handler/permission_handler.dart';

/// 定位服务工具类
class LocationService {
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
    try {
      print('[LocationService] 直接获取定位（不检查服务状态）...');

      // 检查并请求权限
      bool hasPermission = await checkAndRequestPermission();
      if (!hasPermission) {
        print('[LocationService] 没有定位权限，无法获取位置');
        return null;
      }

      print('[LocationService] 正在获取位置信息（使用最低精度，优先网络定位）...');
      // 尝试获取当前位置，即使定位服务可能未启用
      // 使用最低精度，优先使用网络定位（NETWORK_PROVIDER），在室内也能工作
      // 强制使用 Android 原生 LocationManager，避免依赖 Google Location Service
      // 网络定位通常很快，设置较短的超时时间
      Position position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.lowest, // 使用最低精度，优先网络定位
        timeLimit: const Duration(seconds: 15), // 网络定位通常很快，15秒足够
        forceAndroidLocationManager: true, // 强制使用 Android 原生 LocationManager
      );

      print(
        '[LocationService] 直接定位成功: ${position.latitude}, ${position.longitude}, 精度: ${position.accuracy}米',
      );
      return position;
    } catch (e) {
      print('[LocationService] 直接获取定位失败: $e');
      print('[LocationService] 异常类型: ${e.runtimeType}');
      print('[LocationService] 异常详情: ${e.toString()}');

      // 如果第一次尝试失败，可能是网络定位不可用，尝试等待更长时间
      // 但通常如果网络定位可用，第一次就应该成功
      print('[LocationService] 网络定位失败，可能是定位服务未启用或网络不可用');

      return null;
    }
  }

  /// 获取当前位置
  static Future<Position?> getCurrentLocation() async {
    try {
      print('[LocationService] 开始获取定位...');

      // 先检查定位服务是否启用
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
        return null;
      }

      // 检查并请求权限
      bool hasPermission = await checkAndRequestPermission();
      if (!hasPermission) {
        print('[LocationService] 没有定位权限，无法获取位置');
        // 再次检查 Geolocator 的权限状态
        final geolocatorPermission = await Geolocator.checkPermission();
        print('[LocationService] Geolocator 权限状态: $geolocatorPermission');
        return null;
      }

      print('[LocationService] 正在获取位置信息...');
      // 获取当前位置
      // 强制使用 Android 原生 LocationManager，避免依赖 Google Location Service
      // 这样可以确保在国内没有 Google 服务的情况下也能正常定位
      Position position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
        timeLimit: const Duration(seconds: 20),
        forceAndroidLocationManager: true, // 强制使用 Android 原生 LocationManager
      );

      print(
        '[LocationService] 定位成功: ${position.latitude}, ${position.longitude}, 精度: ${position.accuracy}米',
      );
      return position;
    } catch (e) {
      print('[LocationService] 获取定位失败: $e');
      print('[LocationService] 异常类型: ${e.runtimeType}');
      print('[LocationService] 异常堆栈: ${StackTrace.current}');

      // 如果是超时错误，尝试使用低精度定位
      if (e.toString().contains('timeout') ||
          e.toString().contains('TimeoutException')) {
        print('[LocationService] 高精度定位超时，尝试使用低精度定位...');
        try {
          Position position = await Geolocator.getCurrentPosition(
            desiredAccuracy: LocationAccuracy.low,
            timeLimit: const Duration(seconds: 10),
            forceAndroidLocationManager: true, // 强制使用 Android 原生 LocationManager
          );
          print(
            '[LocationService] 低精度定位成功: ${position.latitude}, ${position.longitude}',
          );
          return position;
        } catch (e2) {
          print('[LocationService] 低精度定位也失败: $e2');
        }
      }

      return null;
    }
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
