import 'package:super_app/models/supplier_payment.dart';
import 'package:super_app/utils/request.dart';

class SuppliersApi {
  // 获取供应商列表
  static Future<ApiResponse<List<Map<String, dynamic>>>> getSuppliers() async {
    final response = await Request.get<dynamic>(
      '/admin/suppliers',
      parser: (data) => data,
    );

    if (response.isSuccess && response.data != null) {
      List<Map<String, dynamic>> suppliers = [];
      if (response.data is List) {
        suppliers = (response.data as List<dynamic>)
            .where((item) => item is Map<String, dynamic>)
            .map((item) => item as Map<String, dynamic>)
            .toList();
      }
      return ApiResponse<List<Map<String, dynamic>>>(
        code: response.code,
        message: response.message,
        data: suppliers,
      );
    }

    return ApiResponse<List<Map<String, dynamic>>>(
      code: response.code,
      message: response.message,
      data: [],
    );
  }

  // 获取供应商付款统计列表
  static Future<ApiResponse<List<SupplierPaymentStats>>> getPaymentStats({
    String? timeRange,
    String? status,
    int pageNum = 1,
    int pageSize = 20,
  }) async {
    final queryParams = <String, String>{
      'page': pageNum.toString(),
      'page_size': pageSize.toString(),
    };

    if (timeRange != null && timeRange.isNotEmpty) {
      queryParams['time_range'] = timeRange;
    }
    if (status != null && status.isNotEmpty) {
      queryParams['status'] = status;
    }

    final response = await Request.get<dynamic>(
      '/admin/suppliers/payments/stats',
      parser: (data) => data,
      queryParams: queryParams.isEmpty ? null : queryParams,
    );

    if (response.isSuccess && response.data != null) {
      List<SupplierPaymentStats> statsList = [];
      dynamic data = response.data;
      
      // 处理不同的响应格式
      if (data is Map<String, dynamic> && data['list'] != null) {
        final listData = data['list'] as List<dynamic>? ?? [];
        statsList = listData
            .where((item) => item is Map<String, dynamic>)
            .map((item) => SupplierPaymentStats.fromJson(item as Map<String, dynamic>))
            .toList();
      } else if (data is List) {
        statsList = data
            .where((item) => item is Map<String, dynamic>)
            .map((item) => SupplierPaymentStats.fromJson(item as Map<String, dynamic>))
            .toList();
      }

      return ApiResponse<List<SupplierPaymentStats>>(
        code: response.code,
        message: response.message,
        data: statsList,
      );
    }

    return ApiResponse<List<SupplierPaymentStats>>(
      code: response.code,
      message: response.message,
      data: [],
    );
  }

  // 获取供应商付款详情
  static Future<ApiResponse<SupplierPaymentDetail>> getPaymentDetail(
    int supplierId, {
    String? timeRange,
    String? status,
    int pageNum = 1,
    int pageSize = 20,
  }) async {
    final queryParams = <String, String>{
      'page': pageNum.toString(),
      'page_size': pageSize.toString(),
    };

    if (timeRange != null && timeRange.isNotEmpty) {
      queryParams['time_range'] = timeRange;
    }
    if (status != null && status.isNotEmpty) {
      queryParams['status'] = status;
    }

    final response = await Request.get<Map<String, dynamic>>(
      '/admin/suppliers/$supplierId/payments/detail',
      parser: (data) => data as Map<String, dynamic>,
      queryParams: queryParams.isEmpty ? null : queryParams,
    );

    if (response.isSuccess && response.data != null) {
      final detail = SupplierPaymentDetail.fromJson(response.data!);
      return ApiResponse<SupplierPaymentDetail>(
        code: response.code,
        message: response.message,
        data: detail,
      );
    }

    return ApiResponse<SupplierPaymentDetail>(
      code: response.code,
      message: response.message,
      data: null,
    );
  }

  // 创建供应商付款记录（标记订单项为已付款）
  static Future<ApiResponse<void>> createSupplierPayment(
    int supplierId,
    List<Map<String, dynamic>> orderItems,
    double paymentAmount,
  ) async {
    final now = DateTime.now();
    final paymentDate = '${now.year}-${now.month.toString().padLeft(2, '0')}-${now.day.toString().padLeft(2, '0')}';

    final response = await Request.post<dynamic>(
      '/admin/suppliers/payments',
      body: {
        'supplier_id': supplierId,
        'payment_date': paymentDate,
        'payment_amount': paymentAmount,
        'payment_method': '批量标记',
        'order_items': orderItems,
      },
      parser: (data) => data,
    );

    return ApiResponse<void>(
      code: response.code,
      message: response.message,
      data: null,
    );
  }
}

