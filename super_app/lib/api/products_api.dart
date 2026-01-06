import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:super_app/models/product.dart';
import 'package:super_app/utils/request.dart';
import 'package:super_app/utils/config.dart';
import 'package:super_app/utils/storage.dart';

class ProductsApi {
  // 获取商品列表（管理员）
  static Future<ApiResponse<ProductListResponse>> getProducts({
    int pageNum = 1,
    int pageSize = 20,
    String? keyword,
    int? categoryId,
  }) async {
    try {
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

      var uri = Uri.parse('${Config.apiBaseUrl}/admin/products');
      uri = uri.replace(queryParameters: queryParams);

      final token = await Storage.getToken();
      final headers = <String, String>{
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      };
      if (token != null && token.isNotEmpty) {
        headers['Authorization'] = 'Bearer $token';
      }

      final response = await http.get(uri, headers: headers);

      if (response.statusCode == 401) {
        Storage.clearAll();
        return ApiResponse<ProductListResponse>(
          code: 401,
          message: '缺少身份凭证，请重新登录',
        );
      }

      if (response.body.isEmpty) {
        return ApiResponse<ProductListResponse>(
          code: response.statusCode,
          message: '服务器返回空响应',
        );
      }

      final data = jsonDecode(response.body) as Map<String, dynamic>;
      final code = data['code'] as int? ?? response.statusCode;
      final message = data['message'] as String? ?? '请求失败';
      
      // 后端返回格式：{"code": 200, "data": [...products...], "total": ...}
      final productsList = data['data'] as List<dynamic>? ?? [];
      final total = data['total'] as int? ?? 0;

      if (code == 200) {
        final productListResponse = ProductListResponse(
          list: productsList
              .where((item) => item is Map<String, dynamic>)
              .map((item) => Product.fromJson(item as Map<String, dynamic>))
              .toList(),
          total: total,
        );

        return ApiResponse<ProductListResponse>(
          code: code,
          message: message,
          data: productListResponse,
        );
      }

      return ApiResponse<ProductListResponse>(
        code: code,
        message: message,
        data: null,
      );
    } catch (e) {
      return ApiResponse<ProductListResponse>(
        code: 500,
        message: '网络请求失败: ${e.toString()}',
      );
    }
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

  // 创建商品
  static Future<ApiResponse<Product>> createProduct(Map<String, dynamic> productData) async {
    final response = await Request.post<Map<String, dynamic>>(
      '/admin/products',
      body: productData,
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

  // 更新商品
  static Future<ApiResponse<Product>> updateProduct(int productId, Map<String, dynamic> productData) async {
    final response = await Request.put<Map<String, dynamic>>(
      '/admin/products/$productId',
      body: productData,
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

  // 上传商品图片
  static Future<ApiResponse<String>> uploadProductImage(File imageFile) async {
    try {
      final uri = Uri.parse('${Config.apiBaseUrl}/admin/products/upload');
      final token = await Storage.getToken();
      
      final request = http.MultipartRequest('POST', uri);
      request.headers['Authorization'] = 'Bearer $token';
      
      // 添加文件
      request.files.add(
        await http.MultipartFile.fromPath('file', imageFile.path),
      );

      final streamedResponse = await request.send();
      final response = await http.Response.fromStream(streamedResponse);

      if (response.statusCode == 401) {
        Storage.clearAll();
        return ApiResponse<String>(
          code: 401,
          message: '缺少身份凭证，请重新登录',
        );
      }

      if (response.body.isEmpty) {
        return ApiResponse<String>(
          code: response.statusCode,
          message: '服务器返回空响应',
        );
      }

      final data = jsonDecode(response.body) as Map<String, dynamic>;
      final code = data['code'] as int? ?? response.statusCode;
      final message = data['message'] as String? ?? '请求失败';

      if (code == 200) {
        final responseData = data['data'] as Map<String, dynamic>?;
        final imageUrl = responseData?['imageUrl'] as String? ?? '';
        return ApiResponse<String>(
          code: code,
          message: message,
          data: imageUrl,
        );
      }

      return ApiResponse<String>(
        code: code,
        message: message,
        data: null,
      );
    } catch (e) {
      return ApiResponse<String>(
        code: 500,
        message: '上传失败: ${e.toString()}',
      );
    }
  }
}

