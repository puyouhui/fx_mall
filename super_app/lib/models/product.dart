class Product {
  final int id;
  final String name;
  final String? description;
  final List<String> images;
  final List<ProductSpec> specs;
  final bool isSpecial;
  final int? categoryId;
  final String? categoryName;
  final int? supplierId;
  final String? supplierName;
  final int status;
  final DateTime createdAt;
  final DateTime updatedAt;

  Product({
    required this.id,
    required this.name,
    this.description,
    required this.images,
    required this.specs,
    required this.isSpecial,
    this.categoryId,
    this.categoryName,
    this.supplierId,
    this.supplierName,
    required this.status,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Product.fromJson(Map<String, dynamic> json) {
    final specsData = json['specs'] as List<dynamic>? ?? [];
    
    // 安全解析日期
    DateTime parseDateTime(dynamic value) {
      if (value is String) {
        try {
          return DateTime.parse(value);
        } catch (e) {
          return DateTime.now();
        }
      } else if (value is DateTime) {
        return value;
      }
      return DateTime.now();
    }

    return Product(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: json['name'] as String? ?? '',
      description: json['description'] as String?,
      images: (json['images'] as List<dynamic>? ?? [])
          .map((e) => e.toString())
          .toList(),
      specs: specsData
          .where((s) => s is Map<String, dynamic>)
          .map((s) => ProductSpec.fromJson(s as Map<String, dynamic>))
          .toList(),
      isSpecial: json['is_special'] as bool? ?? false,
      categoryId: (json['category_id'] as num?)?.toInt(),
      categoryName: json['category_name'] as String?,
      supplierId: (json['supplier_id'] as num?)?.toInt(),
      supplierName: json['supplier_name'] as String?,
      status: (json['status'] as num?)?.toInt() ?? 1,
      createdAt: parseDateTime(json['created_at']),
      updatedAt: parseDateTime(json['updated_at']),
    );
  }

  String getPriceRange() {
    if (specs.isEmpty) return '暂无价格';
    final prices = <double>[];
    for (var spec in specs) {
      if (spec.retailPrice > 0) prices.add(spec.retailPrice);
      if (spec.wholesalePrice > 0) prices.add(spec.wholesalePrice);
    }
    if (prices.isEmpty) return '暂无价格';
    prices.sort();
    if (prices.first == prices.last) {
      return '¥${prices.first.toStringAsFixed(2)}';
    }
    return '¥${prices.first.toStringAsFixed(2)} - ¥${prices.last.toStringAsFixed(2)}';
  }
}

class ProductSpec {
  final int id;
  final String name;
  final String? description;
  final double retailPrice;
  final double wholesalePrice;
  final int? stock;
  final int? minOrderQuantity;

  ProductSpec({
    required this.id,
    required this.name,
    this.description,
    required this.retailPrice,
    required this.wholesalePrice,
    this.stock,
    this.minOrderQuantity,
  });

  factory ProductSpec.fromJson(Map<String, dynamic> json) {
    return ProductSpec(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: json['name'] as String? ?? '',
      description: json['description'] as String?,
      retailPrice: (json['retail_price'] as num?)?.toDouble() ?? 0.0,
      wholesalePrice: (json['wholesale_price'] as num?)?.toDouble() ?? 0.0,
      stock: (json['stock'] as num?)?.toInt(),
      minOrderQuantity: (json['min_order_quantity'] as num?)?.toInt(),
    );
  }
}

class ProductListResponse {
  final List<Product> list;
  final int total;

  ProductListResponse({
    required this.list,
    required this.total,
  });

  factory ProductListResponse.fromJson(Map<String, dynamic> json) {
    final listData = json['list'] as List<dynamic>? ?? [];
    return ProductListResponse(
      list: listData.map((item) => Product.fromJson(item as Map<String, dynamic>)).toList(),
      total: json['total'] as int? ?? 0,
    );
  }
}

