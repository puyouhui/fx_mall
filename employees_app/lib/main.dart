import 'package:flutter/material.dart';
import 'package:employees_app/pages/login/login_page.dart';
import 'package:employees_app/pages/home/home_page.dart';
import 'package:employees_app/utils/storage.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/api/auth_api.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();
    // 设置全局导航键
    Request.setNavigatorKey(navigatorKey);

    return MaterialApp(
      title: '员工端应用',
      navigatorKey: navigatorKey,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
        useMaterial3: true,
      ),
      home: const AuthWrapper(),
      routes: {
        '/login': (context) => const LoginPage(),
        '/home': (context) => const HomePage(),
        // 后续添加配送和销售页面路由
        '/delivery/orders': (context) => const HomePage(), // 临时，后续替换
        '/sales/customers': (context) => const HomePage(), // 临时，后续替换
      },
    );
  }
}

class AuthWrapper extends StatefulWidget {
  const AuthWrapper({super.key});

  @override
  State<AuthWrapper> createState() => _AuthWrapperState();
}

class _AuthWrapperState extends State<AuthWrapper> {
  bool _isLoading = true;
  bool _isLoggedIn = false;

  @override
  void initState() {
    super.initState();
    _checkAuthStatus();
  }

  Future<void> _checkAuthStatus() async {
    final token = await Storage.getToken();
    if (token != null && token.isNotEmpty) {
      // 有token，验证员工状态
      try {
        final response = await AuthApi.getCurrentEmployeeInfo();
        if (response.isSuccess && response.data != null) {
          final employee = response.data!;
          // 检查员工状态
          if (!employee.status) {
            // 账号被禁用，清理登录信息并跳转登录页
            await Storage.clearAll();
            if (mounted) {
              setState(() {
                _isLoggedIn = false;
                _isLoading = false;
              });
            }
            return;
          }
          // 状态正常，保持登录状态
          if (mounted) {
            setState(() {
              _isLoggedIn = true;
              _isLoading = false;
            });
          }
        } else {
          // API调用失败（可能是403或其他错误），清理登录信息
          await Storage.clearAll();
          if (mounted) {
            setState(() {
              _isLoggedIn = false;
              _isLoading = false;
            });
          }
        }
      } catch (e) {
        // 发生异常，清理登录信息
        await Storage.clearAll();
        if (mounted) {
          setState(() {
            _isLoggedIn = false;
            _isLoading = false;
          });
        }
      }
    } else {
      // 没有token
      if (mounted) {
        setState(() {
          _isLoggedIn = false;
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }

    return _isLoggedIn ? const HomePage() : const LoginPage();
  }
}
