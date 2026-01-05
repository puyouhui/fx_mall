import 'package:super_app/models/product.dart';
import 'package:super_app/utils/request.dart';

class ProductsApi {
  // 获取商品列表（管理员）
  static Future<ApiResponse<ProductListResponse>> getProducts({
    int pageNum = 1,
    int pageSize = 20,
    String? keyword,
    int? categoryId,
  }) async {
    final queryParams = <String, String>{
      'pageNum': pageNum.toString(),
      'pageSize': pageSize.toString(),
    };

    if (keyword != null && keyword.isNotEmpty) {
      queryParams['keyword'] = keyword;
    }
    if (categoryId != null) {
      queryParams['categoryId'] = categoryId.toString();
    }

    final response = await Request.get<Map<String, dynamic>>(
      '/admin/products',
      queryParams: queryParams,
      parser: (data) => data as Map<String, dynamic>,
    );

    if (response.isSuccess && response.data != null) {
      // 后端返回格式：{"data": [...products...], "total": ...}
      // 需要转换为：{"list": [...products...], "total": ...}
      final data = response.data!;
      final productsList = data['data'] as List<dynamic>? ?? [];
      final total = data['total'] as int? ?? 0;

      final productListResponse = ProductListResponse(
        list: productsList
            .map((item) => Product.fromJson(item as Map<String, dynamic>))
            .toList(),
        total: total,
      );

      return ApiResponse<ProductListResponse>(
        code: response.code,
        message: response.message,
        data: productListResponse,
      );
    }

    return ApiResponse<ProductListResponse>(
      code: response.code,
      message: response.message,
      data: null,
    );
  }

  // 获取商品详情
  static Future<ApiResponse<Product>> getProductDetail(int productId) async {
    final response = await Request.get<Map<String, dynamic>>(
      '/products/$productId',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (response.isSuccess && response.data != null) {
      final product = Product.fromJson(response.data!);
      return ApiResponse<Product>(
        code: response.code,
        message: response.message,
        data: product,
      );
    }

    return ApiResponse<Product>(
      code: response.code,
      message: response.message,
      data: null,
    );
  }

  // 获取分类列表
  static Future<ApiResponse<List<Map<String, dynamic>>>> getCategories() async {
    final response = await Request.get<dynamic>(
      '/categories',
      parser: (data) => data,
    );

    if (response.isSuccess && response.data != null) {
      List<Map<String, dynamic>> categories = [];
      if (response.data is List) {
        categories = (response.data as List<dynamic>)
            .where((item) => item is Map<String, dynamic>)
            .map((item) => item as Map<String, dynamic>)
            .toList();
      }
      return ApiResponse<List<Map<String, dynamic>>>(
        code: response.code,
        message: response.message,
        data: categories,
      );
    }

    return ApiResponse<List<Map<String, dynamic>>>(
      code: response.code,
      message: response.message,
      data: [],
    );
  }
}

