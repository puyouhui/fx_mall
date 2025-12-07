import '../utils/request.dart';

class OrderApi {
  // 获取待配送订单列表（订单池）
  static Future<ApiResponse<Map<String, dynamic>>> getOrderPool({
    int pageNum = 1,
    int pageSize = 20,
    String? status,
  }) async {
    final response = await Request.get<Map<String, dynamic>>(
      '/employee/delivery/orders',
      queryParams: {
        'pageNum': pageNum.toString(),
        'pageSize': pageSize.toString(),
        if (status != null && status.isNotEmpty) 'status': status,
      },
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 接单（接受配送订单）
  static Future<ApiResponse<Map<String, dynamic>>> acceptOrder(
    int orderId,
  ) async {
    final response = await Request.put<Map<String, dynamic>>(
      '/employee/delivery/orders/$orderId/accept',
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 批量接单
  static Future<ApiResponse<Map<String, dynamic>>> acceptOrders(
    List<int> orderIds,
  ) async {
    // 如果后端支持批量接单接口，使用批量接口
    // 否则循环调用单个接单接口
    final results = <Map<String, dynamic>>[];
    String? lastError;

    for (final orderId in orderIds) {
      final response = await acceptOrder(orderId);
      if (response.isSuccess) {
        results.add({'order_id': orderId, 'success': true});
      } else {
        lastError = response.message;
        results.add({
          'order_id': orderId,
          'success': false,
          'message': response.message,
        });
      }
    }

    final allSuccess = results.every((r) => r['success'] == true);
    return ApiResponse<Map<String, dynamic>>(
      code: allSuccess ? 200 : 400,
      message: allSuccess
          ? '接单成功'
          : (lastError ?? '部分订单接单失败'),
      data: {'results': results},
    );
  }

  // 获取订单详情
  static Future<ApiResponse<Map<String, dynamic>>> getOrderDetail(
    int orderId,
  ) async {
    final response = await Request.get<Map<String, dynamic>>(
      '/employee/delivery/orders/$orderId',
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 完成配送
  static Future<ApiResponse<Map<String, dynamic>>> completeOrder(
    int orderId,
  ) async {
    final response = await Request.put<Map<String, dynamic>>(
      '/employee/delivery/orders/$orderId/complete',
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 问题上报
  static Future<ApiResponse<Map<String, dynamic>>> reportOrderIssue({
    required int orderId,
    required String issueType,
    required String description,
    String? contactPhone,
  }) async {
    final response = await Request.post<Map<String, dynamic>>(
      '/employee/delivery/orders/$orderId/report',
      body: {
        'issue_type': issueType,
        'description': description,
        if (contactPhone != null && contactPhone.isNotEmpty)
          'contact_phone': contactPhone,
      },
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }
}

