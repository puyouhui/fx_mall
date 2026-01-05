import 'package:super_app/models/order.dart';
import 'package:super_app/utils/request.dart';

class OrdersApi {
  // 获取订单列表
  static Future<ApiResponse<OrderListResponse>> getOrders({
    int pageNum = 1,
    int pageSize = 10,
    String? keyword,
    String? status,
    String? startDate,
    String? endDate,
  }) async {
    final queryParams = <String, String>{
      'pageNum': pageNum.toString(),
      'pageSize': pageSize.toString(),
    };

    if (keyword != null && keyword.isNotEmpty) {
      queryParams['keyword'] = keyword;
    }
    if (status != null && status.isNotEmpty) {
      queryParams['status'] = status;
    }
    if (startDate != null && startDate.isNotEmpty) {
      queryParams['start_date'] = startDate;
    }
    if (endDate != null && endDate.isNotEmpty) {
      queryParams['end_date'] = endDate;
    }

    final response = await Request.get<OrderListResponse>(
      '/admin/orders',
      queryParams: queryParams,
      parser: (data) => OrderListResponse.fromJson(data as Map<String, dynamic>),
    );

    return ApiResponse<OrderListResponse>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 获取订单详情
  static Future<ApiResponse<Map<String, dynamic>>> getOrderDetail(int orderId) async {
    final response = await Request.get<Map<String, dynamic>>(
      '/admin/orders/$orderId',
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }
}

