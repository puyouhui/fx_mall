import 'dart:io';
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
    int orderId, {
    double? latitude,
    double? longitude,
  }) async {
    final queryParams = <String, String>{};
    if (latitude != null) {
      queryParams['latitude'] = latitude.toString();
    }
    if (longitude != null) {
      queryParams['longitude'] = longitude.toString();
    }

    final response = await Request.put<Map<String, dynamic>>(
      '/employee/delivery/orders/$orderId/accept',
      queryParams: queryParams.isNotEmpty ? queryParams : null,
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
      message: allSuccess ? '接单成功' : (lastError ?? '部分订单接单失败'),
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

  // 获取排序后的配送订单列表（用于路线规划）
  static Future<ApiResponse<Map<String, dynamic>>> getRouteOrders() async {
    final response = await Request.get<Map<String, dynamic>>(
      '/employee/delivery/route/orders',
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 手动触发路线规划计算
  static Future<ApiResponse<Map<String, dynamic>>> calculateRoute({
    double? latitude,
    double? longitude,
  }) async {
    final queryParams = <String, String>{};
    if (latitude != null) {
      queryParams['latitude'] = latitude.toString();
    }
    if (longitude != null) {
      queryParams['longitude'] = longitude.toString();
    }

    final response = await Request.post<Map<String, dynamic>>(
      '/employee/delivery/route/calculate',
      queryParams: queryParams,
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 完成配送（带图片上传）
  static Future<ApiResponse<Map<String, dynamic>>> completeOrderWithImages({
    required int orderId,
    required File productImage,
    required File doorplateImage,
  }) async {
    // 使用 multipart/form-data 上传图片
    final response = await Request.postMultipart<Map<String, dynamic>>(
      '/employee/delivery/orders/$orderId/complete',
      files: {'product_image': productImage, 'doorplate_image': doorplateImage},
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 完成配送（旧接口，保留兼容性）
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

  // 获取待取货供应商列表
  static Future<ApiResponse<List<dynamic>>> getPickupSuppliers({
    double? latitude,
    double? longitude,
  }) async {
    final queryParams = <String, String>{};
    if (latitude != null) {
      queryParams['latitude'] = latitude.toString();
    }
    if (longitude != null) {
      queryParams['longitude'] = longitude.toString();
    }

    final response = await Request.get<List<dynamic>>(
      '/employee/delivery/pickup/suppliers',
      queryParams: queryParams,
      parser: (data) => data as List<dynamic>,
    );

    return ApiResponse<List<dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 获取供应商的待取货商品列表
  static Future<ApiResponse<List<dynamic>>> getPickupItemsBySupplier(
    int supplierId,
  ) async {
    final response = await Request.get<List<dynamic>>(
      '/employee/delivery/pickup/suppliers/$supplierId/items',
      parser: (data) => data as List<dynamic>,
    );

    return ApiResponse<List<dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 标记商品已取货
  static Future<ApiResponse<Map<String, dynamic>>> markItemsAsPicked(
    List<int> itemIds,
  ) async {
    final response = await Request.post<Map<String, dynamic>>(
      '/employee/delivery/pickup/mark-picked',
      body: {'item_ids': itemIds},
      parser: (data) => data as Map<String, dynamic>,
    );

    return ApiResponse<Map<String, dynamic>>(
      code: response.code,
      message: response.message,
      data: response.data,
    );
  }

  // 获取历史订单（已完成的订单：delivered、paid状态）
  static Future<ApiResponse<Map<String, dynamic>>> getHistoryOrders({
    int pageNum = 1,
    int pageSize = 20,
  }) async {
    final response = await Request.get<Map<String, dynamic>>(
      '/employee/delivery/orders',
      queryParams: {
        'pageNum': pageNum.toString(),
        'pageSize': pageSize.toString(),
        'status': 'completed', // 使用completed状态获取所有已完成的订单
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
