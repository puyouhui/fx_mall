import 'dart:convert';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:employees_app/utils/storage.dart';
import 'package:employees_app/utils/config.dart';

class ApiResponse<T> {
  final int code;
  final String message;
  final T? data;

  ApiResponse({required this.code, required this.message, this.data});

  bool get isSuccess => code == 200;
}

class Request {
  // 获取请求头
  static Future<Map<String, String>> _getHeaders({
    bool needAuth = true,
    Map<String, String>? extraHeaders,
  }) async {
    final headers = <String, String>{
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    };

    if (needAuth) {
      final token = await Storage.getToken();
      if (token != null && token.isNotEmpty) {
        headers['Authorization'] = 'Bearer $token';
      }
    }

    if (extraHeaders != null) {
      headers.addAll(extraHeaders);
    }

    return headers;
  }

  // GET 请求
  static Future<ApiResponse<T>> get<T>(
    String path, {
    Map<String, String>? queryParams,
    bool needAuth = true,
    T Function(dynamic)? parser,
  }) async {
    try {
      var uri = Uri.parse('${Config.apiBaseUrl}$path');
      if (queryParams != null && queryParams.isNotEmpty) {
        uri = uri.replace(queryParameters: queryParams);
      }

      final response = await http.get(
        uri,
        headers: await _getHeaders(needAuth: needAuth),
      );

      return _handleResponse<T>(response, parser: parser);
    } catch (e) {
      return ApiResponse<T>(code: 500, message: '网络请求失败: ${e.toString()}');
    }
  }

  // POST 请求
  static Future<ApiResponse<T>> post<T>(
    String path, {
    Map<String, dynamic>? body,
    bool needAuth = true,
    T Function(dynamic)? parser,
  }) async {
    try {
      final uri = Uri.parse('${Config.apiBaseUrl}$path');
      final response = await http.post(
        uri,
        headers: await _getHeaders(needAuth: needAuth),
        body: body != null ? jsonEncode(body) : null,
      );

      return _handleResponse<T>(response, parser: parser);
    } catch (e) {
      return ApiResponse<T>(code: 500, message: '网络请求失败: ${e.toString()}');
    }
  }

  // PUT 请求
  static Future<ApiResponse<T>> put<T>(
    String path, {
    Map<String, dynamic>? body,
    bool needAuth = true,
    T Function(dynamic)? parser,
  }) async {
    try {
      final uri = Uri.parse('${Config.apiBaseUrl}$path');
      final response = await http.put(
        uri,
        headers: await _getHeaders(needAuth: needAuth),
        body: body != null ? jsonEncode(body) : null,
      );

      return _handleResponse<T>(response, parser: parser);
    } catch (e) {
      return ApiResponse<T>(code: 500, message: '网络请求失败: ${e.toString()}');
    }
  }

  // DELETE 请求
  static Future<ApiResponse<T>> delete<T>(
    String path, {
    bool needAuth = true,
    T Function(dynamic)? parser,
  }) async {
    try {
      final uri = Uri.parse('${Config.apiBaseUrl}$path');
      final response = await http.delete(
        uri,
        headers: await _getHeaders(needAuth: needAuth),
      );

      return _handleResponse<T>(response, parser: parser);
    } catch (e) {
      return ApiResponse<T>(code: 500, message: '网络请求失败: ${e.toString()}');
    }
  }

  /// 文件上传（Multipart），用于上传图片等
  static Future<ApiResponse<Map<String, dynamic>>> uploadFile(
    String path,
    File file, {
    String fieldName = 'file',
    bool needAuth = true,
  }) async {
    try {
      final uri = Uri.parse('${Config.apiBaseUrl}$path');
      final request = http.MultipartRequest('POST', uri);

      if (needAuth) {
        final token = await Storage.getToken();
        if (token != null && token.isNotEmpty) {
          request.headers['Authorization'] = 'Bearer $token';
        }
      }
      request.headers['Accept'] = 'application/json';

      request.files.add(
        await http.MultipartFile.fromPath(fieldName, file.path),
      );

      final streamed = await request.send();
      final response = await http.Response.fromStream(streamed);
      return _handleResponse<Map<String, dynamic>>(response);
    } catch (e) {
      return ApiResponse<Map<String, dynamic>>(
        code: 500,
        message: '文件上传失败: ${e.toString()}',
      );
    }
  }

  // 处理响应
  static ApiResponse<T> _handleResponse<T>(
    http.Response response, {
    T Function(dynamic)? parser,
  }) {
    // 检查响应状态码
    if (response.statusCode == 404) {
      return ApiResponse<T>(
        code: 404,
        message: '接口不存在，请检查路径是否正确或后端服务是否已重启',
      );
    }

    // 检查响应内容类型
    final contentType = response.headers['content-type'] ?? '';
    if (!contentType.contains('application/json') && response.body.isNotEmpty) {
      // 如果不是 JSON 格式，可能是 HTML 错误页面
      return ApiResponse<T>(
        code: response.statusCode,
        message: '服务器返回了非 JSON 格式的响应（可能是错误页面）',
      );
    }

    try {
      // 如果响应体为空，返回空响应
      if (response.body.isEmpty) {
        return ApiResponse<T>(
          code: response.statusCode,
          message: '服务器返回空响应',
        );
      }

      final data = jsonDecode(response.body) as Map<String, dynamic>;
      final code = data['code'] as int? ?? response.statusCode;
      final message = data['message'] as String? ?? '请求失败';
      final responseData = data['data'];

      if (code == 401) {
        // Token 失效，清除本地存储
        Storage.clearAll();
      } else if (code == 403) {
        // 账号被禁用或其他状态导致不能使用，清除本地存储并跳转登录页
        Storage.clearAll();
        // 使用全局导航键跳转到登录页
        _navigateToLogin();
      }

      T? parsedData;
      if (responseData != null && parser != null) {
        parsedData = parser(responseData);
      } else if (responseData != null) {
        parsedData = responseData as T?;
      }

      return ApiResponse<T>(code: code, message: message, data: parsedData);
    } catch (e) {
      // 如果解析失败，返回更详细的错误信息
      String errorMessage = '响应解析失败: ${e.toString()}';
      if (response.body.isNotEmpty && response.body.length < 200) {
        errorMessage += '\n响应内容: ${response.body}';
      }
      return ApiResponse<T>(
        code: response.statusCode,
        message: errorMessage,
      );
    }
  }

  // 全局导航键（用于在请求拦截器中跳转）
  static GlobalKey<NavigatorState>? navigatorKey;

  // 设置全局导航键
  static void setNavigatorKey(GlobalKey<NavigatorState> key) {
    navigatorKey = key;
  }

  // 跳转到登录页
  static void _navigateToLogin() {
    if (navigatorKey?.currentState != null) {
      navigatorKey!.currentState!.pushNamedAndRemoveUntil(
        '/login',
        (route) => false,
      );
    }
  }
}
