import 'package:flutter/material.dart';
import 'package:super_app/pages/login/login_page.dart';
import 'package:super_app/pages/home/home_page.dart';
import 'package:super_app/utils/storage.dart';
import 'package:super_app/utils/request.dart';
import 'package:super_app/api/auth_api.dart';

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
      title: '管理员应用',
      navigatorKey: navigatorKey,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF20CB6B),
        ),
        useMaterial3: true,
      ),
      home: const AuthWrapper(),
      routes: {
        '/login': (context) => const LoginPage(),
        '/home': (context) => const HomePage(),
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
      // 有token，验证管理员状态
      try {
        final response = await AuthApi.getCurrentAdminInfo();
        if (response.isSuccess && response.data != null) {
          // 状态正常，保持登录状态
          if (mounted) {
            setState(() {
              _isLoggedIn = true;
              _isLoading = false;
            });
          }
        } else {
          // API调用失败（可能是401或其他错误），清理登录信息
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
      return const Scaffold(
        body: Center(
          child: CircularProgressIndicator(),
        ),
      );
    }

    return _isLoggedIn ? const HomePage() : const LoginPage();
  }
}
