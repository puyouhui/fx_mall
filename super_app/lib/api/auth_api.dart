import 'package:super_app/utils/request.dart';
import 'package:super_app/models/admin.dart';
import 'package:super_app/utils/storage.dart';

class AuthApi {
  // 管理员登录
  static Future<ApiResponse<LoginResponse>> login({
    required String username,
    required String password,
  }) async {
    final response = await Request.post<Map<String, dynamic>>(
      '/admin/login',
      body: {
        'username': username,
        'password': password,
      },
      needAuth: false,
    );

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final loginResponse = LoginResponse.fromJson(data);

      // 保存 token 和管理员信息
      if (loginResponse.token.isNotEmpty) {
        await Storage.saveToken(loginResponse.token);
        await Storage.saveAdminInfo(loginResponse.admin.toJson());
      }

      return ApiResponse<LoginResponse>(
        code: response.code,
        message: response.message,
        data: loginResponse,
      );
    }

    return ApiResponse<LoginResponse>(
      code: response.code,
      message: response.message,
    );
  }

  // 获取当前管理员信息
  static Future<ApiResponse<Admin>> getCurrentAdminInfo() async {
    final response = await Request.get<Map<String, dynamic>>(
      '/admin/info',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (response.isSuccess && response.data != null) {
      final admin = Admin.fromJson(response.data!);
      // 更新本地存储的管理员信息
      await Storage.saveAdminInfo(admin.toJson());
      return ApiResponse<Admin>(
        code: response.code,
        message: response.message,
        data: admin,
      );
    }

    return ApiResponse<Admin>(
      code: response.code,
      message: response.message,
    );
  }

  // 登出
  static Future<void> logout() async {
    await Storage.clearAll();
  }
}

class LoginResponse {
  final String token;
  final Admin admin;

  LoginResponse({
    required this.token,
    required this.admin,
  });

  factory LoginResponse.fromJson(Map<String, dynamic> json) {
    return LoginResponse(
      token: json['token'] as String? ?? '',
      admin: Admin.fromJson(json['admin'] as Map<String, dynamic>),
    );
  }
}

