import 'package:super_app/utils/request.dart';

class PaymentVerificationResponse {
  final List<Map<String, dynamic>> list;
  final int total;

  PaymentVerificationResponse({
    required this.list,
    required this.total,
  });
}

class PaymentVerificationApi {
  // 获取收款审核列表
  static Future<ApiResponse<PaymentVerificationResponse>> getPaymentVerifications({
    String? status,
    int pageNum = 1,
    int pageSize = 20,
  }) async {
    final queryParams = <String, String>{
      'pageNum': pageNum.toString(),
      'pageSize': pageSize.toString(),
    };

    if (status != null && status.isNotEmpty) {
      queryParams['status'] = status;
    }

    final response = await Request.get<dynamic>(
      '/admin/payment-verification',
      parser: (data) => data,
      queryParams: queryParams.isEmpty ? null : queryParams,
    );

    if (response.isSuccess && response.data != null) {
      List<Map<String, dynamic>> verificationList = [];
      int total = 0;
      dynamic data = response.data;
      
      // 处理不同的响应格式
      if (data is Map<String, dynamic>) {
        if (data['list'] != null) {
          final listData = data['list'] as List<dynamic>? ?? [];
          verificationList = listData
              .where((item) => item is Map<String, dynamic>)
              .map((item) => item as Map<String, dynamic>)
              .toList();
        }
        total = (data['total'] as num?)?.toInt() ?? verificationList.length;
      } else if (data is List) {
        verificationList = data
            .where((item) => item is Map<String, dynamic>)
            .map((item) => item as Map<String, dynamic>)
            .toList();
        total = verificationList.length;
      }

      return ApiResponse<PaymentVerificationResponse>(
        code: response.code,
        message: response.message,
        data: PaymentVerificationResponse(
          list: verificationList,
          total: total,
        ),
      );
    }

    return ApiResponse<PaymentVerificationResponse>(
      code: response.code,
      message: response.message,
      data: PaymentVerificationResponse(list: [], total: 0),
    );
  }

  // 审核收款申请
  static Future<ApiResponse<void>> reviewPaymentVerification(
    int requestId,
    bool approved,
    String? remark,
  ) async {
    final response = await Request.post<dynamic>(
      '/admin/payment-verification/review',
      body: {
        'request_id': requestId,
        'approved': approved,
        'review_remark': remark ?? '',
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

