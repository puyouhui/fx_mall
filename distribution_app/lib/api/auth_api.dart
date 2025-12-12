import '../models/employee.dart';
import '../utils/request.dart';
import '../utils/storage.dart';

/// 配送端登录/认证相关接口（与员工端接口保持一致）
class AuthApi {
  // 员工/配送员登录
  static Future<ApiResponse<LoginResponse>> login({
    required String phone,
    required String password,
  }) async {
    final response = await Request.post<Map<String, dynamic>>(
      '/employee/login',
      body: {
        'phone': phone,
        'password': password,
      },
      needAuth: false,
    );

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final loginResponse = LoginResponse.fromJson(data);

      // 保存 token 和员工信息
      if (loginResponse.token.isNotEmpty) {
        await Storage.saveToken(loginResponse.token);
        await Storage.saveEmployeeInfo(loginResponse.employee.toJson());
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

  // 获取当前员工/配送员信息
  static Future<ApiResponse<Employee>> getCurrentEmployeeInfo() async {
    final response = await Request.get<Map<String, dynamic>>(
      '/employee/info',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (response.isSuccess && response.data != null) {
      final employee = Employee.fromJson(response.data!);
      // 更新本地存储的员工信息
      await Storage.saveEmployeeInfo(employee.toJson());
      return ApiResponse<Employee>(
        code: response.code,
        message: response.message,
        data: employee,
      );
    }

    return ApiResponse<Employee>(
      code: response.code,
      message: response.message,
    );
  }

  // 登出：清除本地登录信息
  static Future<void> logout() async {
    await Storage.clearAll();
  }

  // 获取WebSocket配置
  static Future<ApiResponse<Map<String, dynamic>>> getWebSocketConfig() async {
    final response = await Request.get<Map<String, dynamic>>(
      '/employee/websocket-config',
      needAuth: false,
      parser: (data) => data as Map<String, dynamic>,
    );

    if (response.isSuccess && response.data != null) {
      return ApiResponse<Map<String, dynamic>>(
        code: response.code,
        message: response.message,
        data: response.data,
      );
    }

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
    );
  }
}

class LoginResponse {
  final String token;
  final Employee employee;

  LoginResponse({
    required this.token,
    required this.employee,
  });

  factory LoginResponse.fromJson(Map<String, dynamic> json) {
    return LoginResponse(
      token: json['token'] as String? ?? '',
      employee: Employee.fromJson(json['employee'] as Map<String, dynamic>),
    );
  }
}



