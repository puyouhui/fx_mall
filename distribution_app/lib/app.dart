import 'package:flutter/material.dart';
import 'pages/login_page.dart';
import 'pages/main_shell.dart';
import 'pages/batch_pickup_view.dart';
import 'pages/complete_delivery_view.dart';
import 'utils/storage.dart';
import 'utils/request.dart';
import 'api/auth_api.dart';

/// 根组件：定义路由与主题，启动时检查登录状态，自动登录
class DistributionApp extends StatefulWidget {
  const DistributionApp({super.key});

  @override
  State<DistributionApp> createState() => _DistributionAppState();
}

class _DistributionAppState extends State<DistributionApp> {
  bool _isLoading = true;
  String _initialRoute = '/login'; // 默认值设为登录页
  String _courierPhone = '';
  final GlobalKey<NavigatorState> _navigatorKey = GlobalKey<NavigatorState>();

  @override
  void initState() {
    super.initState();
    // 设置全局导航键
    Request.setNavigatorKey(_navigatorKey);
    _checkLoginStatus();
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
