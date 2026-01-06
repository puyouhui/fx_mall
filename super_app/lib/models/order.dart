class Order {
  final int id;
  final String orderNumber;
  final int userId;
  final int addressId;
  final String status;
  final double goodsAmount;
  final double deliveryFee;
  final double pointsDiscount;
  final double couponDiscount;
  final bool isUrgent;
  final double urgentFee;
  final double totalAmount;
  final String? remark;
  final String? outOfStockStrategy;
  final bool trustReceipt;
  final bool hidePrice;
  final bool requirePhoneContact;
  final DateTime createdAt;
  final DateTime updatedAt;
  final int? itemCount;
  final Map<String, dynamic>? user;
  final Map<String, dynamic>? address;
  final Map<String, dynamic>? deliveryEmployee;

  Order({
    required this.id,
    required this.orderNumber,
    required this.userId,
    required this.addressId,
    required this.status,
    required this.goodsAmount,
    required this.deliveryFee,
    required this.pointsDiscount,
    required this.couponDiscount,
    required this.isUrgent,
    required this.urgentFee,
    required this.totalAmount,
    this.remark,
    this.outOfStockStrategy,
    required this.trustReceipt,
    required this.hidePrice,
    required this.requirePhoneContact,
    required this.createdAt,
    required this.updatedAt,
    this.itemCount,
    this.user,
    this.address,
    this.deliveryEmployee,
  });

  factory Order.fromJson(Map<String, dynamic> json) {
    // 安全解析日期
    DateTime parseDateTime(dynamic value) {
      if (value == null) {
        return DateTime.now();
      }
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

    return Order(
      id: (json['id'] as num?)?.toInt() ?? 0,
      orderNumber: json['order_number'] as String? ?? '',
      userId: (json['user_id'] as num?)?.toInt() ?? 0,
      addressId: (json['address_id'] as num?)?.toInt() ?? 0,
      status: json['status'] as String? ?? '',
      goodsAmount: (json['goods_amount'] as num?)?.toDouble() ?? 0.0,
      deliveryFee: (json['delivery_fee'] as num?)?.toDouble() ?? 0.0,
      pointsDiscount: (json['points_discount'] as num?)?.toDouble() ?? 0.0,
      couponDiscount: (json['coupon_discount'] as num?)?.toDouble() ?? 0.0,
      isUrgent: json['is_urgent'] as bool? ?? false,
      urgentFee: (json['urgent_fee'] as num?)?.toDouble() ?? 0.0,
      totalAmount: (json['total_amount'] as num?)?.toDouble() ?? 0.0,
      remark: json['remark'] as String?,
      outOfStockStrategy: json['out_of_stock_strategy'] as String?,
      trustReceipt: json['trust_receipt'] as bool? ?? false,
      hidePrice: json['hide_price'] as bool? ?? false,
      requirePhoneContact: json['require_phone_contact'] as bool? ?? false,
      createdAt: parseDateTime(json['created_at']),
      updatedAt: parseDateTime(json['updated_at']),
      itemCount: (json['item_count'] as num?)?.toInt(),
      user: json['user'] as Map<String, dynamic>?,
      address: json['address'] as Map<String, dynamic>?,
      deliveryEmployee: json['delivery_employee'] as Map<String, dynamic>?,
    );
  }

  String get statusText {
    switch (status) {
      case 'pending_delivery':
        return '待配送';
      case 'pending_pickup':
        return '待取货';
      case 'delivering':
        return '配送中';
      case 'delivered':
        return '已送达';
      case 'paid':
        return '已收款';
      case 'cancelled':
        return '已取消';
      default:
        return status;
    }
  }

  bool get isPendingReview {
    return status == 'pending_delivery' || status == 'pending_pickup';
  }

  bool get isApproved {
    return status == 'delivered' || status == 'paid';
  }

  bool get isRejected {
    return status == 'cancelled';
  }
}

class OrderListResponse {
  final List<Order> list;
  final int total;

  OrderListResponse({
    required this.list,
    required this.total,
  });

  factory OrderListResponse.fromJson(Map<String, dynamic> json) {
    final listData = json['list'] as List<dynamic>? ?? [];
    return OrderListResponse(
      list: listData.map((item) => Order.fromJson(item as Map<String, dynamic>)).toList(),
      total: json['total'] as int? ?? 0,
    );
  }
}

