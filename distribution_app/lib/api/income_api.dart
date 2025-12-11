import '../utils/request.dart';

class IncomeApi {
  // 获取配送员收入统计
  static Future<ApiResponse<Map<String, dynamic>>> getIncomeStats() async {
    final response = await Request.get<Map<String, dynamic>>(
      '/employee/delivery/income/stats',
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 获取配送员收入明细
  static Future<ApiResponse<Map<String, dynamic>>> getIncomeDetails({
    int pageNum = 1,
    int pageSize = 20,
    String? settled, // 'true' 或 'false'
  }) async {
    final queryParams = <String, String>{
      'pageNum': pageNum.toString(),
      'pageSize': pageSize.toString(),
    };
    if (settled != null && settled.isNotEmpty) {
      queryParams['settled'] = settled;
    }

    final response = await Request.get<Map<String, dynamic>>(
      '/employee/delivery/income/details',
      queryParams: queryParams,
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }
}

