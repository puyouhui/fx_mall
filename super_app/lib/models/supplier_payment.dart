class SupplierPaymentStats {
  final int supplierId;
  final String supplierName;
  final double totalAmount;
  final double pendingAmount;
  final double paidAmount;
  final int orderCount;
  final String paymentStatus; // 'all_paid', 'partially_paid', 'pending'

  SupplierPaymentStats({
    required this.supplierId,
    required this.supplierName,
    required this.totalAmount,
    required this.pendingAmount,
    required this.paidAmount,
    required this.orderCount,
    required this.paymentStatus,
  });

  factory SupplierPaymentStats.fromJson(Map<String, dynamic> json) {
    return SupplierPaymentStats(
      supplierId: (json['supplier_id'] as num?)?.toInt() ?? 0,
      supplierName: json['supplier_name'] as String? ?? '',
      totalAmount: (json['total_amount'] as num?)?.toDouble() ?? 0.0,
      pendingAmount: (json['pending_amount'] as num?)?.toDouble() ?? 0.0,
      paidAmount: (json['paid_amount'] as num?)?.toDouble() ?? 0.0,
      orderCount: (json['order_count'] as num?)?.toInt() ?? 0,
      paymentStatus: json['payment_status'] as String? ?? 'pending',
    );
  }

  String getPaymentStatusText() {
    switch (paymentStatus) {
      case 'all_paid':
        return '已结清';
      case 'partially_paid':
        return '部分付款';
      case 'pending':
        return '待付款';
      default:
        return paymentStatus;
    }
  }
}

class SupplierPaymentDetail {
  final int supplierId;
  final String supplierName;
  final double totalAmount;
  final int orderCount;
  final List<SupplierPaymentOrder> orders;

  SupplierPaymentDetail({
    required this.supplierId,
    required this.supplierName,
    required this.totalAmount,
    required this.orderCount,
    required this.orders,
  });

  factory SupplierPaymentDetail.fromJson(Map<String, dynamic> json) {
    final ordersData = json['orders'] as List<dynamic>? ?? [];
    return SupplierPaymentDetail(
      supplierId: (json['supplier_id'] as num?)?.toInt() ?? 0,
      supplierName: json['supplier_name'] as String? ?? '',
      totalAmount: (json['total_amount'] as num?)?.toDouble() ?? 0.0,
      orderCount: (json['order_count'] as num?)?.toInt() ?? 0,
      orders: ordersData
          .map((item) => SupplierPaymentOrder.fromJson(item as Map<String, dynamic>))
          .toList(),
    );
  }
}

/// 管理员 App 按天视图：某天汇总
class SupplierDailyStat {
  final String date; // YYYY-MM-DD
  final double totalAmount;
  final double paidAmount;
  final double pendingAmount;

  SupplierDailyStat({
    required this.date,
    required this.totalAmount,
    required this.paidAmount,
    required this.pendingAmount,
  });

  factory SupplierDailyStat.fromJson(Map<String, dynamic> json) {
    return SupplierDailyStat(
      date: json['date'] as String? ?? '',
      totalAmount: (json['total_amount'] as num?)?.toDouble() ?? 0.0,
      paidAmount: (json['paid_amount'] as num?)?.toDouble() ?? 0.0,
      pendingAmount: (json['pending_amount'] as num?)?.toDouble() ?? 0.0,
    );
  }

  String get statusText {
    if (totalAmount == 0) return '无数据';
    if (pendingAmount <= 0 && paidAmount > 0) return '已付完';
    if (paidAmount <= 0 && pendingAmount > 0) return '待付款';
    return '部分已付';
  }
}

/// 管理员 App 按天视图：某天明细
class SupplierDailyDetail {
  final String date;
  final double totalAmount;
  final double paidAmount;
  final double pendingAmount;
  final List<SupplierDailyItem> pendingItems;
  final List<SupplierDailyItem> paidItems;

  SupplierDailyDetail({
    required this.date,
    required this.totalAmount,
    required this.paidAmount,
    required this.pendingAmount,
    required this.pendingItems,
    required this.paidItems,
  });

  factory SupplierDailyDetail.fromJson(Map<String, dynamic> json) {
    final pending = (json['pending_items'] as List<dynamic>? ?? [])
        .where((e) => e is Map<String, dynamic>)
        .map((e) => SupplierDailyItem.fromJson(e as Map<String, dynamic>))
        .toList();
    final paid = (json['paid_items'] as List<dynamic>? ?? [])
        .where((e) => e is Map<String, dynamic>)
        .map((e) => SupplierDailyItem.fromJson(e as Map<String, dynamic>))
        .toList();
    return SupplierDailyDetail(
      date: json['date'] as String? ?? '',
      totalAmount: (json['total_amount'] as num?)?.toDouble() ?? 0.0,
      paidAmount: (json['paid_amount'] as num?)?.toDouble() ?? 0.0,
      pendingAmount: (json['pending_amount'] as num?)?.toDouble() ?? 0.0,
      pendingItems: pending,
      paidItems: paid,
    );
  }
}

class SupplierDailyItem {
  final int orderId;
  final String orderNumber;
  final int orderItemId;
  final int productId;
  final String productName;
  final String specName;
  final int quantity;
  final double costPrice;
  final double subtotal;

  SupplierDailyItem({
    required this.orderId,
    required this.orderNumber,
    required this.orderItemId,
    required this.productId,
    required this.productName,
    required this.specName,
    required this.quantity,
    required this.costPrice,
    required this.subtotal,
  });

  factory SupplierDailyItem.fromJson(Map<String, dynamic> json) {
    return SupplierDailyItem(
      orderId: (json['order_id'] as num?)?.toInt() ?? 0,
      orderNumber: json['order_number'] as String? ?? '',
      orderItemId: (json['order_item_id'] as num?)?.toInt() ?? 0,
      productId: (json['product_id'] as num?)?.toInt() ?? 0,
      productName: json['product_name'] as String? ?? '',
      specName: json['spec_name'] as String? ?? '',
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
      costPrice: (json['cost_price'] as num?)?.toDouble() ?? 0.0,
      subtotal: (json['subtotal'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class SupplierPaymentOrder {
  final int orderId;
  final String orderNumber;
  final String? addressName;
  final DateTime createdAt;
  final double totalAmount;
  final String status;
  final List<SupplierPaymentItem> items;

  SupplierPaymentOrder({
    required this.orderId,
    required this.orderNumber,
    this.addressName,
    required this.createdAt,
    required this.totalAmount,
    required this.status,
    required this.items,
  });

  factory SupplierPaymentOrder.fromJson(Map<String, dynamic> json) {
    final itemsData = json['items'] as List<dynamic>? ?? [];
    return SupplierPaymentOrder(
      orderId: (json['order_id'] as num?)?.toInt() ?? 0,
      orderNumber: json['order_number'] as String? ?? '',
      addressName: json['address_name'] as String?,
      createdAt: json['created_at'] is String
          ? DateTime.parse(json['created_at'])
          : (json['order_date'] is String
              ? DateTime.parse(json['order_date'])
              : (json['created_at'] is DateTime
                  ? json['created_at'] as DateTime
                  : DateTime.now())),
      totalAmount: (json['total_cost'] as num?)?.toDouble() ??
          (json['total_amount'] as num?)?.toDouble() ?? 0.0,
      status: json['status'] as String? ?? '',
      items: itemsData
          .map((item) => SupplierPaymentItem.fromJson(item as Map<String, dynamic>))
          .toList(),
    );
  }
}

class SupplierPaymentItem {
  final int orderItemId;
  final int orderId;
  final int productId;
  final String productName;
  final String specName;
  final int quantity;
  final double costPrice;
  final double subtotal;
  final bool isPaid;

  SupplierPaymentItem({
    required this.orderItemId,
    required this.orderId,
    required this.productId,
    required this.productName,
    required this.specName,
    required this.quantity,
    required this.costPrice,
    required this.subtotal,
    required this.isPaid,
  });

  factory SupplierPaymentItem.fromJson(Map<String, dynamic> json) {
    // 从订单数据中获取 order_id（如果 item 中没有）
    final orderId = (json['order_id'] as num?)?.toInt();
    
    return SupplierPaymentItem(
      orderItemId: (json['order_item_id'] as num?)?.toInt() ?? 0,
      orderId: orderId ?? 0,
      productId: (json['product_id'] as num?)?.toInt() ?? 0,
      productName: json['product_name'] as String? ?? '',
      specName: json['spec_name'] as String? ?? '',
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
      costPrice: (json['cost_price'] as num?)?.toDouble() ?? 0.0,
      subtotal: (json['subtotal'] as num?)?.toDouble() ?? 0.0,
      isPaid: json['is_paid'] as bool? ?? false,
    );
  }
}

