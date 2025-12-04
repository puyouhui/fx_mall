import 'package:flutter/material.dart';
import 'pages/login_page.dart';
import 'pages/main_shell.dart';

/// 根组件：定义路由与主题，控制从登录页跳转到主页面
class DistributionApp extends StatelessWidget {
  const DistributionApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: '配送员端',
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
      ),
      initialRoute: '/login',
      onGenerateRoute: (RouteSettings settings) {
        switch (settings.name) {
          case '/login':
            return MaterialPageRoute(
              builder: (_) => const LoginPage(),
              settings: settings,
            );
          case '/main':
            final phone = settings.arguments as String? ?? '';
            return MaterialPageRoute(
              builder: (_) => MainShell(courierPhone: phone),
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


