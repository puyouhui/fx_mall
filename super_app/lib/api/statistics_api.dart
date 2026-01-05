import 'package:super_app/utils/request.dart';

class StatisticsApi {
  // 获取仪表盘统计数据
  static Future<ApiResponse<Map<String, dynamic>>> getDashboardStats({
    String timeRange = 'today',
    String? startDate,
    String? endDate,
  }) async {
    final queryParams = <String, String>{
      'time_range': timeRange,
    };
    
    if (startDate != null && endDate != null) {
      queryParams['start_date'] = startDate;
      queryParams['end_date'] = endDate;
    }

    final response = await Request.get<Map<String, dynamic>>(
      '/admin/dashboard/stats',
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

