import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';

/// 配送端本地存储：保存 token 和员工/配送员信息
class Storage {
  static const String _keyToken = 'distribution_token';
  static const String _keyEmployeeInfo = 'distribution_employee_info';

  // 保存 Token
  static Future<void> saveToken(String token) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyToken, token);
  }

  // 获取 Token
  static Future<String?> getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keyToken);
  }

  // 删除 Token
  static Future<void> removeToken() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_keyToken);
  }

  // 保存员工/配送员信息
  static Future<void> saveEmployeeInfo(
      Map<String, dynamic> employeeInfo) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyEmployeeInfo, jsonEncode(employeeInfo));
  }

  // 获取员工/配送员信息
  static Future<Map<String, dynamic>?> getEmployeeInfo() async {
    final prefs = await SharedPreferences.getInstance();
    final infoStr = prefs.getString(_keyEmployeeInfo);
    if (infoStr == null) return null;
    try {
      return jsonDecode(infoStr) as Map<String, dynamic>;
    } catch (e) {
      return null;
    }
  }

  // 删除员工/配送员信息
  static Future<void> removeEmployeeInfo() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_keyEmployeeInfo);
  }

  // 清除所有登录信息
  static Future<void> clearAll() async {
    await removeToken();
    await removeEmployeeInfo();
  }
}



