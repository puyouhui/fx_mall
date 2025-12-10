import 'dart:convert';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'config.dart';
import 'storage.dart';

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
    Map<String, String>? queryParams,
    bool needAuth = true,
    T Function(dynamic)? parser,
  }) async {
    try {
      var uri = Uri.parse('${Config.apiBaseUrl}$path');
      if (queryParams != null && queryParams.isNotEmpty) {
        uri = uri.replace(queryParameters: queryParams);
      }
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
    Map<String, String>? queryParams,
    bool needAuth = true,
    T Function(dynamic)? parser,
  }) async {
    try {
      var uri = Uri.parse('${Config.apiBaseUrl}$path');
      if (queryParams != null && queryParams.isNotEmpty) {
        uri = uri.replace(queryParameters: queryParams);
      }
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

  // 文件上传（如上传图片）
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

  // 多文件上传（multipart/form-data）
  static Future<ApiResponse<T>> postMultipart<T>(
    String path, {
    Map<String, File>? files,
    Map<String, String>? fields,
    bool needAuth = true,
    T Function(dynamic)? parser,
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

      // 添加文件
      if (files != null) {
        for (final entry in files.entries) {
          request.files.add(
            await http.MultipartFile.fromPath(entry.key, entry.value.path),
          );
        }
      }

      // 添加字段
      if (fields != null) {
        request.fields.addAll(fields);
      }

      final streamed = await request.send();
      final response = await http.Response.fromStream(streamed);
      return _handleResponse<T>(response, parser: parser);
    } catch (e) {
      return ApiResponse<T>(code: 500, message: '文件上传失败: ${e.toString()}');
    }
  }

  // 处理响应
  static ApiResponse<T> _handleResponse<T>(
    http.Response response, {
    T Function(dynamic)? parser,
  }) {
    try {
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
      return ApiResponse<T>(
        code: response.statusCode,
        message: '响应解析失败: ${e.toString()}',
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
