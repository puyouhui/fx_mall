import 'package:shared_preferences/shared_preferences.dart';
import 'dart:convert';

class Storage {
  static const String _keyToken = 'admin_token';
  static const String _keyAdminInfo = 'admin_info';
  static const String _keyRememberPassword = 'remember_password';
  static const String _keySavedUsername = 'saved_username';
  static const String _keySavedPassword = 'saved_password';

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

  // 保存管理员信息
  static Future<void> saveAdminInfo(Map<String, dynamic> adminInfo) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyAdminInfo, jsonEncode(adminInfo));
  }

  // 获取管理员信息
  static Future<Map<String, dynamic>?> getAdminInfo() async {
    final prefs = await SharedPreferences.getInstance();
    final infoStr = prefs.getString(_keyAdminInfo);
    if (infoStr == null) return null;
    try {
      return jsonDecode(infoStr) as Map<String, dynamic>;
    } catch (e) {
      return null;
    }
  }

  // 删除管理员信息
  static Future<void> removeAdminInfo() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_keyAdminInfo);
  }

  // 清除所有登录信息
  static Future<void> clearAll() async {
    await removeToken();
    await removeAdminInfo();
  }

  // 保存记住密码状态
  static Future<void> saveRememberPassword(bool remember) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_keyRememberPassword, remember);
  }

  // 获取记住密码状态
  static Future<bool> getRememberPassword() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getBool(_keyRememberPassword) ?? false;
  }

  // 保存账号密码（仅在勾选记住密码时使用）
  static Future<void> saveCredentials(String username, String password) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keySavedUsername, username);
    await prefs.setString(_keySavedPassword, password);
  }

  // 获取保存的账号
  static Future<String?> getSavedUsername() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keySavedUsername);
  }

  // 获取保存的密码
  static Future<String?> getSavedPassword() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keySavedPassword);
  }

  // 清除保存的账号密码
  static Future<void> clearSavedCredentials() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_keySavedUsername);
    await prefs.remove(_keySavedPassword);
  }
}

