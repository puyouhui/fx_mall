import 'package:flutter/material.dart';
import 'app.dart';

/// 应用入口：确保插件初始化后再启动应用
Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const DistributionApp());
}
