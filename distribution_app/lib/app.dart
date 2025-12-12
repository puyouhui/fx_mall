import 'package:flutter/material.dart';
import 'pages/login_page.dart';
import 'pages/main_shell.dart';
import 'pages/batch_pickup_view.dart';
import 'pages/complete_delivery_view.dart';
import 'utils/storage.dart';
import 'utils/request.dart';
import 'utils/location_report_service.dart';
import 'api/auth_api.dart';

/// 根组件：定义路由与主题，启动时检查登录状态，自动登录
class DistributionApp extends StatefulWidget {
  const DistributionApp({super.key});

  @override
  State<DistributionApp> createState() => _DistributionAppState();
}

class _DistributionAppState extends State<DistributionApp>
    with WidgetsBindingObserver {
  bool _isLoading = true;
  String _initialRoute = '/login'; // 默认值设为登录页
  String _courierPhone = '';
  final GlobalKey<NavigatorState> _navigatorKey = GlobalKey<NavigatorState>();

  @override
  void initState() {
    super.initState();
    // 监听应用生命周期
    WidgetsBinding.instance.addObserver(this);
    // 设置全局导航键
    Request.setNavigatorKey(_navigatorKey);
    _checkLoginStatus();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    super.didChangeAppLifecycleState(state);
    if (state == AppLifecycleState.resumed) {
      // 应用从后台返回前台时，立即检查并重连WebSocket
      print('[DistributionApp] 应用进入前台，立即检查并重连WebSocket连接');
      _ensureWebSocketConnected();
    } else if (state == AppLifecycleState.paused) {
      // 应用进入后台时，记录状态但不关闭连接（让系统管理）
      print('[DistributionApp] 应用进入后台（连接保持，等待系统管理）');
    } else if (state == AppLifecycleState.inactive) {
      // 应用处于非活动状态（如来电时）
      print('[DistributionApp] 应用处于非活动状态');
    } else if (state == AppLifecycleState.detached) {
      // 应用被系统终止
      print('[DistributionApp] 应用被系统终止');
    }
  }

  /// 确保WebSocket连接（如果已登录）
  Future<void> _ensureWebSocketConnected() async {
    final token = await Storage.getToken();
    if (token != null && token.isNotEmpty) {
      try {
        print('[DistributionApp] 开始确保WebSocket连接...');
        // 强制检查并重连（应用恢复时，连接可能已被系统关闭）
        await LocationReportService.instance.ensureConnected();
        print('[DistributionApp] WebSocket连接确保完成');
      } catch (e) {
        print('[DistributionApp] 确保WebSocket连接失败: $e');
        // 如果失败，延迟后重试
        Future.delayed(const Duration(seconds: 2), () {
          _ensureWebSocketConnected();
        });
      }
    } else {
      print('[DistributionApp] 未登录，跳过WebSocket连接检查');
    }
  }

  Future<void> _checkLoginStatus() async {
    // 检查是否有token
    final token = await Storage.getToken();
    if (token != null && token.isNotEmpty) {
      // 有token，验证员工状态
      try {
        final response = await AuthApi.getCurrentEmployeeInfo();
        if (response.isSuccess && response.data != null) {
          final employee = response.data!;
          // 检查员工状态和角色
          if (!employee.status) {
            // 账号被禁用，清理登录信息并跳转登录页
            await Storage.clearAll();
            if (mounted) {
              setState(() {
                _initialRoute = '/login';
                _isLoading = false;
              });
            }
            return;
          }
          if (!employee.isDelivery) {
            // 不是配送员，清理登录信息并跳转登录页
            await Storage.clearAll();
            if (mounted) {
              setState(() {
                _initialRoute = '/login';
                _isLoading = false;
              });
            }
            return;
          }
          // 状态正常，跳转到主页面
          final phone = employee.phone;

          // 启动位置上报服务
          try {
            await LocationReportService.instance.start();
          } catch (e) {
            print('启动位置上报服务失败: $e');
          }

          if (mounted) {
            setState(() {
              _initialRoute = '/main';
              _courierPhone = phone;
              _isLoading = false;
            });
          }
        } else {
          // API调用失败（可能是403或其他错误），清理登录信息
          await Storage.clearAll();
          if (mounted) {
            setState(() {
              _initialRoute = '/login';
              _isLoading = false;
            });
          }
        }
      } catch (e) {
        // 发生异常，清理登录信息
        await Storage.clearAll();
        if (mounted) {
          setState(() {
            _initialRoute = '/login';
            _isLoading = false;
          });
        }
      }
    } else {
      // 没有token，跳转到登录页
      if (mounted) {
        setState(() {
          _initialRoute = '/login';
          _isLoading = false;
        });
      }
    }
  }

  Widget _buildHome() {
    if (_initialRoute == '/main') {
      return MainShell(
        courierPhone: _courierPhone.isNotEmpty ? _courierPhone : '配送员',
      );
    }
    return const LoginPage();
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      // 启动时显示加载页面
      return MaterialApp(
        debugShowCheckedModeBanner: false,
        home: Scaffold(
          backgroundColor: const Color(0xFF20CB6B),
          body: const Center(
            child: CircularProgressIndicator(
              valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
            ),
          ),
        ),
      );
    }

    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: '配送员端',
      navigatorKey: _navigatorKey,
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
      ),
      home: _buildHome(),
      onGenerateRoute: (RouteSettings settings) {
        final routeName = settings.name ?? '/login';
        switch (routeName) {
          case '/login':
            return MaterialPageRoute(
              builder: (_) => const LoginPage(),
              settings: settings,
            );
          case '/main':
            // 优先使用传入的参数，否则使用存储的phone
            final phone = (settings.arguments as String?) ?? _courierPhone;
            return MaterialPageRoute(
              builder: (_) =>
                  MainShell(courierPhone: phone.isNotEmpty ? phone : '配送员'),
              settings: settings,
            );
          case '/batch-pickup':
            return MaterialPageRoute(
              builder: (_) => const BatchPickupView(),
              settings: settings,
            );
          case '/complete-delivery':
            return MaterialPageRoute(
              builder: (_) => const CompleteDeliveryView(),
              settings: settings,
            );
          default:
            return MaterialPageRoute(
              builder: (_) => const LoginPage(),
              settings: const RouteSettings(name: '/login'),
            );
        }
      },
    );
  }
}
