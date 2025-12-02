import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:employees_app/utils/request.dart';

/// 订单详情页（销售员）
class OrderDetailPage extends StatefulWidget {
  final int orderId;

  const OrderDetailPage({super.key, required this.orderId});

  @override
  State<OrderDetailPage> createState() => _OrderDetailPageState();
}

class _OrderDetailPageState extends State<OrderDetailPage> {
  bool _loading = true;
  Map<String, dynamic>? _order;
  Map<String, dynamic>? _user;
  Map<String, dynamic>? _address;
  List<dynamic> _items = [];

  @override
  void initState() {
    super.initState();
    _loadDetail();
  }

  Future<void> _loadDetail() async {
    setState(() {
      _loading = true;
    });

    final resp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/orders/${widget.orderId}',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      setState(() {
        _order = resp.data!['order'] as Map<String, dynamic>?;
        _user = resp.data!['user'] as Map<String, dynamic>?;
        _address = resp.data!['address'] as Map<String, dynamic>?;
        _items = resp.data!['order_items'] as List<dynamic>? ?? [];
        _loading = false;
      });
    } else {
      setState(() {
        _loading = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '获取订单详情失败'),
        ),
      );
    }
  }

  String _formatDateTime(dynamic raw) {
    if (raw == null) return '';
    try {
      final dt = raw is DateTime ? raw : DateTime.tryParse(raw.toString());
      if (dt == null) return raw.toString();
      return DateFormat('yyyy-MM-dd HH:mm').format(dt.toLocal());
    } catch (_) {
      return raw.toString();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('订单详情'),
        centerTitle: true,
        backgroundColor: const Color(0xFF20CB6B),
        elevation: 0,
        iconTheme: const IconThemeData(color: Colors.white),
        titleTextStyle: const TextStyle(
          color: Colors.white,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [Color(0xFF20CB6B), Color(0xFFEFF7F2)],
          ),
        ),
        child: SafeArea(
          child: _loading
              ? const Center(
                  child: CircularProgressIndicator(
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                )
              : _order == null
              ? const Center(
                  child: Text(
                    '订单不存在',
                    style: TextStyle(color: Colors.white, fontSize: 14),
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _loadDetail,
                  child: ListView(
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
                    children: [
                      _buildBaseInfoCard(),
                      const SizedBox(height: 12),
                      _buildCustomerCard(),
                      const SizedBox(height: 12),
                      _buildItemsCard(),
                      const SizedBox(height: 12),
                      _buildAmountSummaryCard(),
                    ],
                  ),
                ),
        ),
      ),
    );
  }

  Widget _buildBaseInfoCard() {
    final order = _order ?? {};
    final orderNumber = order['order_number']?.toString() ?? '';
    final status = order['status']?.toString() ?? '';
    final createdAt = _formatDateTime(order['created_at']);

    String statusText = status;
    Color statusColor = const Color(0xFF8C92A4);
    if (status == 'pending_delivery' || status == 'pending') {
      statusText = '待配送';
      statusColor = const Color(0xFFFFA940);
    } else if (status == 'delivering') {
      statusText = '配送中';
      statusColor = const Color(0xFF4C8DF6);
    } else if (status == 'delivered' || status == 'shipped') {
      statusText = '已送达';
      statusColor = const Color(0xFF20CB6B);
    } else if (status == 'paid') {
      statusText = '已收款';
      statusColor = const Color(0xFF20CB6B);
    } else if (status == 'cancelled') {
      statusText = '已取消';
      statusColor = const Color(0xFFB0B4C3);
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Text(
                '订单信息',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              Text(
                statusText,
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: statusColor,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          if (orderNumber.isNotEmpty)
            Text(
              '订单号：$orderNumber',
              style: const TextStyle(fontSize: 13, color: Color(0xFF40475C)),
            ),
          if (createdAt.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              '下单时间：$createdAt',
              style: const TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildCustomerCard() {
    final user = _user ?? {};
    final address = _address ?? {};

    final name = (user['name'] as String?) ?? '未填写名称';
    final phone = (user['phone'] as String?) ?? '';
    final userCode = (user['user_code'] as String?) ?? '';
    final storeType = (user['store_type'] as String?) ?? '';
    final userType = (user['user_type'] as String?) ?? '';

    final addrName = (address['name'] as String?) ?? '';
    final addrText = (address['address'] as String?) ?? '';
    final contact = (address['contact'] as String?) ?? '';
    final addrPhone = (address['phone'] as String?) ?? '';

    String userTypeText = '';
    if (userType == 'wholesale') {
      userTypeText = '批发客户';
    } else if (userType == 'retail') {
      userTypeText = '零售客户';
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '客户与收货信息',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 10),
          Row(
            children: [
              Expanded(
                child: Text(
                  name,
                  style: const TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (phone.isNotEmpty) ...[
                const SizedBox(width: 8),
                const Icon(Icons.phone, size: 14, color: Color(0xFF8C92A4)),
                const SizedBox(width: 4),
                Text(
                  phone,
                  style: const TextStyle(
                    fontSize: 13,
                    color: Color(0xFF40475C),
                  ),
                ),
              ],
            ],
          ),
          const SizedBox(height: 4),
          if (userCode.isNotEmpty)
            Text(
              '客户编号：$userCode',
              style: const TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
            ),
          if (storeType.isNotEmpty || userTypeText.isNotEmpty) ...[
            const SizedBox(height: 4),
            Row(
              children: [
                if (storeType.isNotEmpty)
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFF4C8DF6).withOpacity(0.06),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Text(
                      storeType,
                      style: const TextStyle(
                        fontSize: 11,
                        color: Color(0xFF4C8DF6),
                      ),
                    ),
                  ),
                if (userTypeText.isNotEmpty) ...[
                  const SizedBox(width: 6),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFF20CB6B).withOpacity(0.06),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Text(
                      userTypeText,
                      style: const TextStyle(
                        fontSize: 11,
                        color: Color(0xFF20CB6B),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ],
          const SizedBox(height: 10),
          if (addrText.isNotEmpty) ...[
            Row(
              children: [
                const Icon(
                  Icons.location_on_outlined,
                  size: 18,
                  color: Color(0xFF20CB6B),
                ),
                const SizedBox(width: 4),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        addrName.isNotEmpty ? addrName : '收货地址',
                        style: const TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF20253A),
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        addrText,
                        style: const TextStyle(
                          fontSize: 13,
                          color: Color(0xFF40475C),
                        ),
                      ),
                      if (contact.isNotEmpty || addrPhone.isNotEmpty)
                        Padding(
                          padding: const EdgeInsets.only(top: 2),
                          child: Text(
                            '$contact  $addrPhone',
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildItemsCard() {
    if (_items.isEmpty) {
      return Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: const Text(
          '暂无商品明细',
          style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
        ),
      );
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '商品明细',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 10),
          ..._items.map((raw) {
            final item = raw as Map<String, dynamic>;
            final name = (item['product_name'] as String?) ?? '';
            final spec = (item['spec_name'] as String?) ?? '';
            final qty = (item['quantity'] as int?) ?? 0;
            final unitPrice = (item['unit_price'] as num?)?.toDouble() ?? 0.0;
            final subtotal = (item['subtotal'] as num?)?.toDouble() ?? 0.0;
            final image = (item['image'] as String?) ?? '';

            return Container(
              margin: const EdgeInsets.only(bottom: 12),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    width: 56,
                    height: 56,
                    decoration: BoxDecoration(
                      color: const Color(0xFFF5F6FA),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    clipBehavior: Clip.antiAlias,
                    child: image.isNotEmpty
                        ? Image.network(
                            image,
                            fit: BoxFit.cover,
                            errorBuilder: (context, error, stackTrace) {
                              return const Icon(
                                Icons.image_not_supported,
                                color: Color(0xFFB0B4C3),
                              );
                            },
                          )
                        : const Icon(Icons.image, color: Color(0xFFB0B4C3)),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          name,
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                        if (spec.isNotEmpty) ...[
                          const SizedBox(height: 2),
                          Text(
                            spec,
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                        ],
                        const SizedBox(height: 6),
                        Row(
                          children: [
                            Text(
                              '¥${unitPrice.toStringAsFixed(2)}',
                              style: const TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                                color: Color(0xFF20CB6B),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Text(
                              'x$qty',
                              style: const TextStyle(
                                fontSize: 12,
                                color: Color(0xFF8C92A4),
                              ),
                            ),
                            const Spacer(),
                            Text(
                              '¥${subtotal.toStringAsFixed(2)}',
                              style: const TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                                color: Color(0xFF20253A),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  Widget _buildAmountSummaryCard() {
    final order = _order ?? {};
    final goodsAmount = (order['goods_amount'] as num?)?.toDouble() ?? 0.0;
    final deliveryFee = (order['delivery_fee'] as num?)?.toDouble() ?? 0.0;
    final pointsDiscount =
        (order['points_discount'] as num?)?.toDouble() ?? 0.0;
    final couponDiscount =
        (order['coupon_discount'] as num?)?.toDouble() ?? 0.0;
    final totalAmount = (order['total_amount'] as num?)?.toDouble() ?? 0.0;
    final remark = (order['remark'] as String?) ?? '';

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '金额信息',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 10),
          _buildAmountRow('商品金额', goodsAmount),
          _buildAmountRow('配送费用', deliveryFee),
          if (pointsDiscount > 0) _buildAmountRow('积分抵扣', -pointsDiscount),
          if (couponDiscount > 0) _buildAmountRow('优惠券抵扣', -couponDiscount),
          const Divider(height: 20, thickness: 0.5, color: Color(0xFFE5E7F0)),
          Row(
            children: [
              const Text(
                '实付金额',
                style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
              ),
              const Spacer(),
              Text(
                '¥${totalAmount.toStringAsFixed(2)}',
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF20CB6B),
                ),
              ),
            ],
          ),
          if (remark.isNotEmpty) ...[
            const SizedBox(height: 8),
            Text(
              '备注：$remark',
              style: const TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildAmountRow(String label, double value) {
    final isNegative = value < 0;
    final display = value.abs();
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        children: [
          Text(
            label,
            style: const TextStyle(fontSize: 13, color: Color(0xFF40475C)),
          ),
          const Spacer(),
          Text(
            '${isNegative ? '-' : ''}¥${display.toStringAsFixed(2)}',
            style: TextStyle(
              fontSize: 13,
              color: isNegative
                  ? const Color(0xFFFF5A5F)
                  : const Color(0xFF40475C),
            ),
          ),
        ],
      ),
    );
  }
}
