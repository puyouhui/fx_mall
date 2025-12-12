import 'dart:async';
import 'dart:convert';
import 'package:geolocator/geolocator.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'config.dart';
import 'storage.dart';
import 'location_service.dart';
import '../api/auth_api.dart';

/// 位置上报服务
class LocationReportService {
  static LocationReportService? _instance;
  WebSocketChannel? _channel;
  StreamSubscription<Position>? _positionSubscription;
  StreamSubscription<dynamic>? _messageSubscription; // WebSocket消息订阅
  Timer? _reportTimer;
  Timer? _heartbeatTimer; // 心跳定时器
  bool _isRunning = false;
  bool _isConnecting = false; // 是否正在连接中
  String? _token;
  String? _websocketUrl; // 缓存的WebSocket URL
  DateTime? _lastMessageTime; // 最后收到消息的时间

  LocationReportService._();

  static LocationReportService get instance {
    _instance ??= LocationReportService._();
    return _instance!;
  }

  /// 启动位置上报
  Future<void> start() async {
    if (_isRunning) {
      print('[LocationReportService] 位置上报服务已在运行');
      return;
    }

    // 获取token
    _token = await Storage.getToken();
    if (_token == null || _token!.isEmpty) {
      print('[LocationReportService] Token不存在，无法启动位置上报');
      return;
    }

    // 检查定位权限
    final hasPermission = await LocationService.checkAndRequestPermission();
    if (!hasPermission) {
      print('[LocationReportService] 定位权限未授予，无法启动位置上报');
      print('[LocationReportService] 请到应用设置中授予定位权限');
      return;
    }

    // 检查定位服务是否启用
    final serviceEnabled = await LocationService.checkLocationServiceEnabled();
    if (!serviceEnabled) {
      print('[LocationReportService] 定位服务未启用，无法启动位置上报');
      print('[LocationReportService] 请开启系统定位服务');
      return;
    }

    _isRunning = true;
    print('[LocationReportService] 启动位置上报服务');

    // 建立WebSocket连接
    await _connectWebSocket();

    // 开始监听位置变化
    _startLocationTracking();

    // 启动心跳检测
    _startHeartbeat();
  }

  /// 停止位置上报
  Future<void> stop() async {
    if (!_isRunning) {
      return;
    }

    _isRunning = false;
    print('[LocationReportService] 停止位置上报服务');

    // 取消重连定时器
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    _reconnectAttempts = 0;

    // 停止心跳定时器
    _heartbeatTimer?.cancel();
    _heartbeatTimer = null;

    // 停止WebSocket消息监听
    await _messageSubscription?.cancel();
    _messageSubscription = null;

    // 停止位置监听
    await _positionSubscription?.cancel();
    _positionSubscription = null;

    // 停止定时器
    _reportTimer?.cancel();
    _reportTimer = null;

    // 关闭WebSocket连接
    try {
      await _channel?.sink.close();
    } catch (e) {
      print('[LocationReportService] 关闭WebSocket连接时出错: $e');
    }
    _channel = null;
  }

  /// 建立WebSocket连接
  Future<void> _connectWebSocket() async {
    // 如果正在连接中，避免重复连接
    if (_isConnecting) {
      print('[LocationReportService] 正在连接中，跳过重复连接');
      return;
    }

    _isConnecting = true;

    // 先关闭旧连接（如果存在）
    if (_channel != null) {
      try {
        await _messageSubscription?.cancel();
        _messageSubscription = null;
        await _channel!.sink.close();
        print('[LocationReportService] 已关闭旧WebSocket连接');
      } catch (e) {
        print('[LocationReportService] 关闭旧连接时出错: $e');
      }
      _channel = null;
    }

    try {
      // 获取token
      if (_token == null || _token!.isEmpty) {
        _token = await Storage.getToken();
      }
      if (_token == null || _token!.isEmpty) {
        print('[LocationReportService] Token不存在，无法建立WebSocket连接');
        return;
      }

      // 获取WebSocket URL（从API获取或使用缓存的）
      if (_websocketUrl == null || _websocketUrl!.isEmpty) {
        await _fetchWebSocketUrl();
      }

      if (_websocketUrl == null || _websocketUrl!.isEmpty) {
        print('[LocationReportService] 无法获取WebSocket URL');
        return;
      }

      // 构建完整的WebSocket URL（通过URL参数传递token）
      final wsUrl = _getWebSocketUrl();
      print('[LocationReportService] 连接WebSocket: $wsUrl');

      _channel = WebSocketChannel.connect(Uri.parse(wsUrl));

      // 监听连接状态
      _messageSubscription = _channel!.stream.listen(
        (message) {
          try {
            _lastMessageTime = DateTime.now(); // 更新最后收到消息的时间
            final data = jsonDecode(message as String) as Map<String, dynamic>;
            final type = data['type'] as String?;
            if (type == 'location_received') {
              print('[LocationReportService] 位置上报成功');
              // 收到消息说明连接正常，重置重连计数器
              _reconnectAttempts = 0;
            } else if (type == 'pong') {
              print('[LocationReportService] 收到心跳响应');
              // 收到pong说明连接正常，重置重连计数器
              _reconnectAttempts = 0;
            }
          } catch (e) {
            print('[LocationReportService] 解析WebSocket消息失败: $e');
          }
        },
        onError: (error) {
          print('[LocationReportService] WebSocket错误: $error');
          _isConnecting = false;
          // 尝试重连
          if (_isRunning) {
            _reconnect();
          }
        },
        onDone: () {
          print('[LocationReportService] WebSocket连接已关闭');
          _isConnecting = false;
          _messageSubscription = null;
          // 尝试重连
          if (_isRunning) {
            _reconnect();
          }
        },
        cancelOnError: false, // 不自动取消，让重连机制处理
      );

      // 连接成功后重置重连计数器和最后消息时间
      _reconnectAttempts = 0;
      _lastMessageTime = DateTime.now();
      _isConnecting = false;
      print('[LocationReportService] WebSocket连接成功');
    } catch (e) {
      print('[LocationReportService] 建立WebSocket连接失败: $e');
      _isConnecting = false;
      // 延迟后重试
      Future.delayed(const Duration(seconds: 5), () {
        if (_isRunning) {
          _connectWebSocket();
        }
      });
    }
  }

  /// 从API获取WebSocket URL
  Future<void> _fetchWebSocketUrl() async {
    try {
      print('[LocationReportService] 开始从API获取WebSocket URL...');
      final response = await AuthApi.getWebSocketConfig();
      print(
        '[LocationReportService] API响应: code=${response.code}, message=${response.message}',
      );

      if (response.isSuccess && response.data != null) {
        print('[LocationReportService] API返回数据: ${response.data}');
        _websocketUrl = response.data!['employee_location_url'] as String?;
        print('[LocationReportService] 原始WebSocket URL: $_websocketUrl');

        if (_websocketUrl != null && _websocketUrl!.isNotEmpty) {
          // 确保URL是ws://或wss://格式（API返回的可能已经是ws://格式）
          if (!_websocketUrl!.startsWith('ws://') &&
              !_websocketUrl!.startsWith('wss://')) {
            // 将http/https转换为ws/wss
            print('[LocationReportService] 转换URL格式: $_websocketUrl');
            _websocketUrl = _websocketUrl!
                .replaceFirst('http://', 'ws://')
                .replaceFirst('https://', 'wss://');
            print('[LocationReportService] 转换后URL: $_websocketUrl');
          }
          print('[LocationReportService] 获取WebSocket URL成功: $_websocketUrl');
        } else {
          print('[LocationReportService] WebSocket URL为空');
        }
      } else {
        print('[LocationReportService] 获取WebSocket URL失败: ${response.message}');
      }
    } catch (e) {
      print('[LocationReportService] 获取WebSocket URL异常: $e');
      print('[LocationReportService] 异常堆栈: ${StackTrace.current}');
    }
  }

  /// 获取WebSocket URL（包含token参数）
  String _getWebSocketUrl() {
    String finalUrl;

    if (_websocketUrl == null || _websocketUrl!.isEmpty) {
      // 如果API获取失败，使用默认URL
      final baseUrl = Config.baseUrl;
      print('[LocationReportService] 使用默认URL，baseUrl: $baseUrl');

      // 确保转换为ws://或wss://格式
      if (baseUrl.startsWith('http://')) {
        finalUrl = baseUrl.replaceFirst('http://', 'ws://');
      } else if (baseUrl.startsWith('https://')) {
        finalUrl = baseUrl.replaceFirst('https://', 'wss://');
      } else if (baseUrl.startsWith('ws://') || baseUrl.startsWith('wss://')) {
        finalUrl = baseUrl;
      } else {
        // 如果没有协议，默认使用ws://
        finalUrl = 'ws://$baseUrl';
      }

      finalUrl = '$finalUrl/api/mini/employee/location/ws';
      _websocketUrl = finalUrl;
      print('[LocationReportService] 构建默认WebSocket URL: $_websocketUrl');
    } else {
      finalUrl = _websocketUrl!;
      // 再次确保URL格式正确（防止API返回http://格式）
      if (finalUrl.startsWith('http://')) {
        finalUrl = finalUrl.replaceFirst('http://', 'ws://');
        _websocketUrl = finalUrl;
      } else if (finalUrl.startsWith('https://')) {
        finalUrl = finalUrl.replaceFirst('https://', 'wss://');
        _websocketUrl = finalUrl;
      }
    }

    // 通过URL参数传递token（因为WebSocket不支持自定义请求头）
    final tokenParam = _token != null
        ? '?token=${Uri.encodeComponent(_token!)}'
        : '';
    final result = '$finalUrl$tokenParam';
    print('[LocationReportService] 最终WebSocket URL: $result');
    return result;
  }

  /// 重连WebSocket（带指数退避）
  Timer? _reconnectTimer;
  int _reconnectAttempts = 0;
  static const int _maxReconnectAttempts = 10; // 最大重连次数
  static const int _maxReconnectDelay = 60; // 最大重连延迟（秒）

  Future<void> _reconnect() async {
    if (!_isRunning) {
      return;
    }

    // 取消之前的重连定时器
    _reconnectTimer?.cancel();

    // 如果超过最大重连次数，重置计数器（避免无限增长）
    if (_reconnectAttempts >= _maxReconnectAttempts) {
      _reconnectAttempts = 0;
    }

    // 计算重连延迟（指数退避：5秒、10秒、20秒...最多60秒）
    final delay = (_reconnectAttempts < 3)
        ? 5 * (_reconnectAttempts + 1)
        : _maxReconnectDelay;
    _reconnectAttempts++;

    print(
      '[LocationReportService] ${delay}秒后尝试重连WebSocket... (第$_reconnectAttempts次)',
    );

    _reconnectTimer = Timer(Duration(seconds: delay), () async {
      if (_isRunning) {
        await _connectWebSocket();
      }
    });
  }

  /// 确保WebSocket连接（如果未连接则连接，如果已连接则检查连接状态）
  Future<void> ensureConnected() async {
    if (!_isRunning) {
      // 如果服务未运行，尝试启动
      print('[LocationReportService] 服务未运行，尝试启动...');
      await start();
      // 启动后立即上报当前位置
      await _getCurrentLocationAndReport();
      return;
    }

    // 检查连接状态
    if (_channel == null) {
      print('[LocationReportService] WebSocket连接不存在，重新连接...');
      await _connectWebSocket();
      // 连接后立即上报当前位置
      await Future.delayed(const Duration(milliseconds: 500));
      await _getCurrentLocationAndReport();
      return;
    }

    // 检查连接是否正常（尝试发送测试消息）
    try {
      // 尝试发送一个ping消息来测试连接
      final pingMessage = jsonEncode({'type': 'ping'});
      _channel!.sink.add(pingMessage);

      // 如果发送成功，重置重连计数器
      _reconnectAttempts = 0;
      print('[LocationReportService] WebSocket连接正常');

      // 连接正常，立即上报当前位置
      await _getCurrentLocationAndReport();
    } catch (e) {
      print('[LocationReportService] WebSocket连接异常，尝试重连: $e');
      // 连接异常，关闭旧连接并重新连接
      try {
        await _channel?.sink.close();
      } catch (_) {
        // 忽略关闭错误
      }
      _channel = null;
      await _connectWebSocket();
      // 重连后立即上报当前位置
      await Future.delayed(const Duration(milliseconds: 500));
      await _getCurrentLocationAndReport();
    }
  }

  /// 开始监听位置变化
  void _startLocationTracking() {
    print('[LocationReportService] 开始监听位置变化...');

    // 使用位置流监听位置变化
    _positionSubscription =
        Geolocator.getPositionStream(
          locationSettings: const LocationSettings(
            accuracy: LocationAccuracy.high,
            distanceFilter: 10, // 每移动10米更新一次
          ),
        ).listen(
          (Position position) {
            print(
              '[LocationReportService] 位置流更新: ${position.latitude}, ${position.longitude}',
            );
            _reportLocation(position);
          },
          onError: (error) {
            print('[LocationReportService] 位置监听错误: $error');
            // 如果权限被拒绝，停止服务
            if (error.toString().contains('permission') ||
                error.toString().contains('权限')) {
              print('[LocationReportService] 定位权限问题，停止位置上报服务');
              stop();
            }
          },
          cancelOnError: false, // 不因错误自动取消
        );

    print('[LocationReportService] 位置流监听已启动');

    // 立即获取一次位置并上报
    _getCurrentLocationAndReport();
  }

  /// 获取当前位置并上报
  Future<void> _getCurrentLocationAndReport() async {
    try {
      final position = await LocationService.getCurrentLocation();
      if (position != null) {
        _reportLocation(position);
      }
    } catch (e) {
      print('[LocationReportService] 获取当前位置失败: $e');
    }
  }

  /// 上报位置
  void _reportLocation(Position position) {
    if (!_isRunning) {
      print('[LocationReportService] 服务未运行，跳过位置上报');
      return;
    }

    if (_channel == null) {
      print('[LocationReportService] WebSocket连接不存在，跳过位置上报');
      // 尝试重连
      _reconnect();
      return;
    }

    try {
      final message = jsonEncode({
        'type': 'location',
        'latitude': position.latitude,
        'longitude': position.longitude,
        'accuracy': position.accuracy,
      });

      _channel!.sink.add(message);
      print(
        '[LocationReportService] 上报位置成功: ${position.latitude}, ${position.longitude}, 精度: ${position.accuracy}米',
      );
    } catch (e) {
      print('[LocationReportService] 上报位置失败: $e');
      // 如果WebSocket连接有问题，尝试重连
      if (e.toString().contains('closed') || e.toString().contains('连接')) {
        print('[LocationReportService] WebSocket连接异常，尝试重连...');
        _reconnect();
      }
    }
  }

  /// 启动心跳检测（每30秒检查一次连接状态）
  void _startHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimer = Timer.periodic(const Duration(seconds: 30), (timer) {
      if (!_isRunning) {
        timer.cancel();
        return;
      }

      // 检查连接是否存在
      if (_channel == null) {
        print('[LocationReportService] 心跳检测：连接不存在，尝试重连');
        _connectWebSocket();
        return;
      }

      // 检查是否长时间未收到消息（超过60秒）
      if (_lastMessageTime != null) {
        final timeSinceLastMessage = DateTime.now().difference(
          _lastMessageTime!,
        );
        if (timeSinceLastMessage.inSeconds > 60) {
          print('[LocationReportService] 心跳检测：超过60秒未收到消息，尝试重连');
          _reconnect();
          return;
        }
      }

      // 尝试发送心跳（ping）
      try {
        final pingMessage = jsonEncode({'type': 'ping'});
        _channel!.sink.add(pingMessage);
        print('[LocationReportService] 发送心跳ping');
      } catch (e) {
        print('[LocationReportService] 发送心跳失败: $e');
        // 如果发送失败，说明连接有问题，尝试重连
        _reconnect();
      }
    });
  }
}
