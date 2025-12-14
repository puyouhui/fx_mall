import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:employees_app/utils/request.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import 'package:employees_app/utils/coordinate_transform.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:flutter/services.dart';
import 'package:employees_app/pages/order/sales_edit_order_page.dart';

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
  Map<String, dynamic>? _deliveryEmployee;
  List<dynamic> _items = [];

  // 地图相关
  final MapController _mapController = MapController();
  Map<String, dynamic>? _deliveryEmployeeLocation;
  bool _loadingDeliveryLocation = false;

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
        _deliveryEmployee =
            resp.data!['delivery_employee'] as Map<String, dynamic>?;
        _items = resp.data!['order_items'] as List<dynamic>? ?? [];
        _loading = false;
      });

      // 如果订单状态是待取货或配送中，且有配送员，则获取配送员位置
      final status = _order?['status']?.toString() ?? '';
      if ((status == 'pending_pickup' || status == 'delivering') &&
          _deliveryEmployee != null &&
          _deliveryEmployee!['employee_code'] != null) {
        _loadDeliveryEmployeeLocation();
      }
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
      extendBody: true, // 让body延伸到系统操作条下方
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
          bottom: false, // 底部不使用SafeArea，让内容延伸到系统操作条
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
              : Column(
                  children: [
                    Expanded(
                      child: RefreshIndicator(
                        onRefresh: _loadDetail,
                        child: ListView(
                          padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
                          children: [
                            _buildBaseInfoCard(),
                            const SizedBox(height: 12),
                            _buildCustomerCard(),
                            const SizedBox(height: 12),
                            // 如果有配送员，显示配送员信息
                            if (_deliveryEmployee != null &&
                                _deliveryEmployee!['employee_code'] !=
                                    null) ...[
                              _buildDeliveryEmployeeCard(),
                              const SizedBox(height: 12),
                            ],
                            // 待取货或配送中状态时显示地图
                            if (_shouldShowMap()) ...[
                              _buildMapCard(),
                              const SizedBox(height: 12),
                            ],
                            _buildItemsCard(),
                            const SizedBox(height: 12),
                            _buildAmountSummaryCard(),
                          ],
                        ),
                      ),
                    ),
                    // 底部操作按钮
                    _buildActionButtons(),
                  ],
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
    } else if (status == 'pending_pickup') {
      statusText = '待取货';
      statusColor = const Color(0xFF4C8DF6);
    } else if (status == 'delivering') {
      statusText = '配送中';
      statusColor = const Color(0xFF4C8DF6);
    } else if (status == 'delivered' || status == 'shipped') {
      statusText = '已送达';
      statusColor = const Color(0xFF20CB6B);
    } else if (status == 'paid' || status == 'completed') {
      statusText = '已收款';
      statusColor = const Color(0xFF20CB6B);
    } else if (status == 'cancelled') {
      statusText = '已取消';
      statusColor = const Color(0xFFB0B4C3);
    } else {
      // 未知状态，显示原始状态
      statusText = status;
      statusColor = const Color(0xFF8C92A4);
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
                  fontSize: 16,
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
                InkWell(
                  onTap: () => _makePhoneCall(phone),
                  borderRadius: BorderRadius.circular(20),
                  child: Container(
                    padding: const EdgeInsets.all(6),
                    decoration: BoxDecoration(
                      color: const Color(0xFF20CB6B).withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      Icons.phone,
                      size: 16,
                      color: Color(0xFF20CB6B),
                    ),
                  ),
                ),
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

  /// 构建配送员信息卡片
  Widget _buildDeliveryEmployeeCard() {
    final deliveryEmployee = _deliveryEmployee ?? {};
    final name = (deliveryEmployee['name'] as String?) ?? '';
    final phone = (deliveryEmployee['phone'] as String?) ?? '';
    final employeeCode = (deliveryEmployee['employee_code'] as String?) ?? '';

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
            '配送员信息',
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
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (name.isNotEmpty)
                      Text(
                        name,
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF20253A),
                        ),
                      ),
                    if (employeeCode.isNotEmpty) ...[
                      const SizedBox(height: 4),
                      Text(
                        '员工编号：$employeeCode',
                        style: const TextStyle(
                          fontSize: 12,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              if (phone.isNotEmpty) ...[
                InkWell(
                  onTap: () => _makePhoneCall(phone),
                  borderRadius: BorderRadius.circular(20),
                  child: Container(
                    padding: const EdgeInsets.all(6),
                    decoration: BoxDecoration(
                      color: const Color(0xFF4C8DF6).withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      Icons.phone,
                      size: 16,
                      color: Color(0xFF4C8DF6),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
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
        ],
      ),
    );
  }

  /// 拨打电话
  Future<void> _makePhoneCall(String phone) async {
    try {
      // 使用原生平台通道直接调用 Android Intent
      const platform = MethodChannel('com.example.employees_app/phone');
      await platform.invokeMethod('dialPhone', {'phone': phone});
    } catch (e) {
      // 如果原生方法失败，尝试使用 url_launcher
      try {
        final uri = Uri.parse('tel:$phone');
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } catch (e2) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('拨打电话失败，请手动拨打: $phone'),
              duration: const Duration(seconds: 3),
            ),
          );
        }
      }
    }
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

  /// 判断订单是否应该显示操作按钮（待配送状态就显示，即使被锁定）
  bool _shouldShowActionButtons() {
    final order = _order;
    if (order == null) return false;

    final status = order['status']?.toString() ?? '';
    // 待配送状态的订单都应该显示按钮
    return status == 'pending_delivery' || status == 'pending';
  }

  /// 处理取消修改（解锁订单）
  Future<void> _handleCancelEdit() async {
    final order = _order;
    if (order == null) return;

    // 显示美化的确认对话框
    final confirmed = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (context) => Dialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // 信息图标
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  color: const Color(0xFF4C8DF6).withOpacity(0.1),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.info_outline,
                  color: Color(0xFF4C8DF6),
                  size: 32,
                ),
              ),
              const SizedBox(height: 20),
              // 标题
              const Text(
                '确认取消修改',
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const SizedBox(height: 12),
              // 内容
              const Text(
                '确定要取消修改订单吗？\n\n取消后订单将解锁，配送员可以接单。',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  color: Color(0xFF8C92A4),
                  height: 1.5,
                ),
              ),
              const SizedBox(height: 24),
              // 按钮
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: () => Navigator.of(context).pop(false),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        side: const BorderSide(color: Color(0xFFE5E7EB)),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      child: const Text(
                        '继续修改',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton(
                      onPressed: () => Navigator.of(context).pop(true),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFFFF5A5F),
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        elevation: 0,
                      ),
                      child: const Text(
                        '取消修改',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );

    if (confirmed != true) return;

    // 调用解锁订单API
    final resp = await Request.post<dynamic>(
      '/employee/sales/orders/${widget.orderId}/unlock',
    );

    if (!mounted) return;

    if (resp.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('订单已解锁'),
          backgroundColor: Color(0xFF20CB6B),
        ),
      );
      // 重新加载订单详情
      await _loadDetail();
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '解锁订单失败'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  /// 构建底部操作按钮
  Widget _buildActionButtons() {
    if (!_shouldShowActionButtons()) {
      return const SizedBox.shrink();
    }

    final order = _order;
    final isLocked = (order?['is_locked'] as bool?) ?? false;

    // 如果订单被锁定，显示"取消锁定"和"继续修改"按钮
    if (isLocked) {
      return Container(
        padding: EdgeInsets.fromLTRB(
          16,
          12,
          16,
          0 + MediaQuery.of(context).padding.bottom,
        ),
        decoration: BoxDecoration(
          color: Colors.white,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.1),
              blurRadius: 10,
              offset: const Offset(0, -2),
            ),
          ],
        ),
        child: SafeArea(
          top: false,
          child: Row(
            children: [
              // 取消锁定按钮
              Expanded(
                child: OutlinedButton(
                  onPressed: _handleCancelEdit,
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    side: const BorderSide(
                      color: Color(0xFFFF5A5F),
                      width: 1.5,
                    ),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: const Text(
                    '取消锁定',
                    style: TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFFFF5A5F),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              // 继续修改按钮
              Expanded(
                child: ElevatedButton(
                  onPressed: _handleContinueEdit,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF20CB6B),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    elevation: 0,
                  ),
                  child: const Text(
                    '继续修改',
                    style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
                  ),
                ),
              ),
            ],
          ),
        ),
      );
    }

    // 如果订单未被锁定，显示"取消订单"和"修改订单"按钮
    return Container(
      padding: EdgeInsets.fromLTRB(
        16,
        12,
        16,
        0 + MediaQuery.of(context).padding.bottom,
      ),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 10,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: SafeArea(
        top: false,
        child: Row(
          children: [
            // 取消订单按钮
            Expanded(
              child: OutlinedButton(
                onPressed: _handleCancelOrder,
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  side: const BorderSide(color: Color(0xFFFF5A5F), width: 1.5),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                child: const Text(
                  '取消订单',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFFFF5A5F),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            // 修改订单按钮
            Expanded(
              child: ElevatedButton(
                onPressed: _handleEditOrder,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF20CB6B),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  elevation: 0,
                ),
                child: const Text(
                  '修改订单',
                  style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// 处理取消订单
  Future<void> _handleCancelOrder() async {
    final order = _order;
    if (order == null) return;

    final orderNumber = order['order_number']?.toString() ?? '';

    // 显示美化的确认对话框
    final confirmed = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (context) => Dialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // 警告图标
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  color: const Color(0xFFFF5A5F).withOpacity(0.1),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.warning_amber_rounded,
                  color: Color(0xFFFF5A5F),
                  size: 32,
                ),
              ),
              const SizedBox(height: 20),
              // 标题
              const Text(
                '确认取消订单',
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const SizedBox(height: 12),
              // 内容
              Text(
                '确定要取消订单 $orderNumber 吗？\n\n取消后订单将无法恢复。',
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 14,
                  color: Color(0xFF8C92A4),
                  height: 1.5,
                ),
              ),
              const SizedBox(height: 24),
              // 按钮
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: () => Navigator.of(context).pop(false),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        side: const BorderSide(color: Color(0xFFE5E7EB)),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      child: const Text(
                        '取消',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton(
                      onPressed: () => Navigator.of(context).pop(true),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFFFF5A5F),
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        elevation: 0,
                      ),
                      child: const Text(
                        '确认取消',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );

    if (confirmed != true) return;

    // 调用取消订单API
    final resp = await Request.post<dynamic>(
      '/employee/sales/orders/${widget.orderId}/cancel',
    );

    if (!mounted) return;

    if (resp.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('订单已取消'),
          backgroundColor: Color(0xFF20CB6B),
        ),
      );
      // 重新加载订单详情
      await _loadDetail();
      // 返回上一页
      if (mounted) {
        Navigator.of(context).pop(true);
      }
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '取消订单失败'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  /// 处理继续修改（订单已锁定，直接跳转到修改页面）
  Future<void> _handleContinueEdit() async {
    // 跳转到修改订单页面
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => SalesEditOrderPage(orderId: widget.orderId),
      ),
    );

    // 如果修改订单页面返回false或null，说明用户取消了修改或退出页面，需要解锁订单
    // 如果返回true，说明修改成功，订单已在后端自动解锁
    if (result != true && mounted) {
      // 尝试解锁订单（如果后端已经解锁，这个调用会失败，但不影响）
      await Request.post('/employee/sales/orders/${widget.orderId}/unlock');
    }

    // 刷新订单详情
    if (mounted) {
      _loadDetail();
    }
  }

  /// 处理修改订单
  Future<void> _handleEditOrder() async {
    // 先锁定订单
    final lockResp = await Request.post<Map<String, dynamic>>(
      '/employee/sales/orders/${widget.orderId}/lock',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (!lockResp.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            lockResp.message.isNotEmpty ? lockResp.message : '锁定订单失败',
          ),
        ),
      );
      return;
    }

    // 显示美化的提示对话框
    final shouldProceed = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (context) => Dialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // 图标
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  color: const Color(0xFF20CB6B).withOpacity(0.1),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.info_outline,
                  color: Color(0xFF20CB6B),
                  size: 32,
                ),
              ),
              const SizedBox(height: 20),
              // 标题
              const Text(
                '订单已锁定',
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const SizedBox(height: 12),
              // 内容
              const Text(
                '订单已锁定，配送员无法接单。\n\n修改订单期间请不要退出页面，点击"取消锁定"或"保存修改"时会自动解锁订单。',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  color: Color(0xFF8C92A4),
                  height: 1.5,
                ),
              ),
              const SizedBox(height: 24),
              // 按钮
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: () => Navigator.of(context).pop(false),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        side: const BorderSide(color: Color(0xFFE5E7EB)),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      child: const Text(
                        '取消',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton(
                      onPressed: () => Navigator.of(context).pop(true),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF20CB6B),
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        elevation: 0,
                      ),
                      child: const Text(
                        '确定',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );

    if (!mounted) {
      // 如果用户关闭了对话框但页面已销毁，解锁订单
      await Request.post('/employee/sales/orders/${widget.orderId}/unlock');
      return;
    }

    if (shouldProceed != true) {
      // 用户取消，解锁订单
      await Request.post('/employee/sales/orders/${widget.orderId}/unlock');
      return;
    }

    // 用户点击"确定"，订单保持锁定状态，跳转到修改订单页面
    // 注意：订单在此时保持锁定状态，只有在修改页面点击"取消修改"、"保存修改"或退出页面时才会解锁
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => SalesEditOrderPage(orderId: widget.orderId),
      ),
    );

    // 如果修改订单页面返回false或null，说明用户取消了修改或退出页面，需要解锁订单
    // 如果返回true，说明修改成功，订单已在后端自动解锁
    if (result != true && mounted) {
      // 尝试解锁订单（如果后端已经解锁，这个调用会失败，但不影响）
      await Request.post('/employee/sales/orders/${widget.orderId}/unlock');
    }

    // 刷新订单详情
    if (mounted) {
      _loadDetail();
    }
  }

  /// 判断是否应该显示地图
  bool _shouldShowMap() {
    final status = _order?['status']?.toString() ?? '';
    return status == 'pending_pickup' || status == 'delivering';
  }

  /// 获取配送员位置
  Future<void> _loadDeliveryEmployeeLocation() async {
    if (_deliveryEmployee == null) return;

    final employeeCode = _deliveryEmployee!['employee_code']?.toString();
    if (employeeCode == null || employeeCode.isEmpty) return;

    setState(() {
      _loadingDeliveryLocation = true;
    });

    try {
      final resp = await Request.get<Map<String, dynamic>>(
        '/employee/delivery-employee-location/$employeeCode',
        parser: (data) => data as Map<String, dynamic>,
      );

      if (!mounted) return;

      if (resp.isSuccess && resp.data != null) {
        setState(() {
          _deliveryEmployeeLocation = resp.data;
          _loadingDeliveryLocation = false;
        });

        // 调整地图视野以显示所有标记
        _adjustMapBounds();
      } else {
        setState(() {
          _loadingDeliveryLocation = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _loadingDeliveryLocation = false;
        });
      }
    }
  }

  /// 调整地图视野以显示所有标记
  void _adjustMapBounds() {
    final address = _address;
    final deliveryLocation = _deliveryEmployeeLocation;

    if (address == null || deliveryLocation == null) return;

    final customerLat = address['latitude'];
    final customerLng = address['longitude'];
    final deliveryLat = deliveryLocation['latitude'];
    final deliveryLng = deliveryLocation['longitude'];

    if (customerLat == null ||
        customerLng == null ||
        deliveryLat == null ||
        deliveryLng == null)
      return;

    // 计算边界
    final customerLatNum = customerLat as num;
    final customerLngNum = customerLng as num;
    final deliveryLatNum = deliveryLat as num;
    final deliveryLngNum = deliveryLng as num;

    final minLat = customerLatNum.toDouble() < deliveryLatNum.toDouble()
        ? customerLatNum.toDouble()
        : deliveryLatNum.toDouble();
    final maxLat = customerLatNum.toDouble() > deliveryLatNum.toDouble()
        ? customerLatNum.toDouble()
        : deliveryLatNum.toDouble();
    final minLng = customerLngNum.toDouble() < deliveryLngNum.toDouble()
        ? customerLngNum.toDouble()
        : deliveryLngNum.toDouble();
    final maxLng = customerLngNum.toDouble() > deliveryLngNum.toDouble()
        ? customerLngNum.toDouble()
        : deliveryLngNum.toDouble();

    // 转换为WGS84坐标
    final customerWgs84 = CoordinateTransform.gcj02ToWgs84(minLat, minLng);
    final deliveryWgs84 = CoordinateTransform.gcj02ToWgs84(maxLat, maxLng);

    // 计算中心点和缩放级别
    final centerLat = (customerWgs84.latitude + deliveryWgs84.latitude) / 2;
    final centerLng = (customerWgs84.longitude + deliveryWgs84.longitude) / 2;

    // 计算距离以确定合适的缩放级别
    final latDiff = (customerWgs84.latitude - deliveryWgs84.latitude).abs();
    final lngDiff = (customerWgs84.longitude - deliveryWgs84.longitude).abs();
    final maxDiff = latDiff > lngDiff ? latDiff : lngDiff;

    double zoom = 13.0;
    if (maxDiff > 0.1) {
      zoom = 11.0;
    } else if (maxDiff > 0.05) {
      zoom = 12.0;
    } else if (maxDiff > 0.02) {
      zoom = 13.0;
    } else {
      zoom = 14.0;
    }

    // 移动地图到合适的位置
    _mapController.move(LatLng(centerLat, centerLng), zoom);
  }

  /// 构建地图卡片
  Widget _buildMapCard() {
    final address = _address;
    final deliveryLocation = _deliveryEmployeeLocation;

    // 检查是否有位置信息
    final customerLat = address?['latitude'];
    final customerLng = address?['longitude'];
    final hasCustomerLocation = customerLat != null && customerLng != null;
    final hasDeliveryLocation =
        deliveryLocation != null &&
        deliveryLocation['latitude'] != null &&
        deliveryLocation['longitude'] != null;

    if (!hasCustomerLocation && !hasDeliveryLocation) {
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
        child: const Center(
          child: Text(
            '暂无位置信息',
            style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
          ),
        ),
      );
    }

    // 确定地图初始中心点
    LatLng initialCenter = const LatLng(39.90864, 116.39750); // 默认北京
    if (hasCustomerLocation) {
      final wgs84Point = CoordinateTransform.gcj02ToWgs84(
        (customerLat as num).toDouble(),
        (customerLng as num).toDouble(),
      );
      initialCenter = wgs84Point;
    } else if (hasDeliveryLocation) {
      final deliveryLat = deliveryLocation['latitude'] as num;
      final deliveryLng = deliveryLocation['longitude'] as num;
      final wgs84Point = CoordinateTransform.gcj02ToWgs84(
        deliveryLat.toDouble(),
        deliveryLng.toDouble(),
      );
      initialCenter = wgs84Point;
    }

    // 天地图瓦片服务 URL 模板（Web墨卡托投影）
    const String tiandituTileUrlTemplate =
        'https://t{s}.tianditu.gov.cn/img_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=img&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=d95864378581051adb04fe26acb13ecf';

    const String tiandituLabelUrlTemplate =
        'https://t{s}.tianditu.gov.cn/cia_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=cia&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=d95864378581051adb04fe26acb13ecf';

    TileProvider createTiandituTileProvider() {
      return NetworkTileProvider(
        headers: {
          'User-Agent':
              'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
          'Referer': 'https://lbs.tianditu.gov.cn/',
          'Accept': 'image/webp,image/apng,image/*,*/*;q=0.8',
          'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8',
        },
      );
    }

    return Container(
      height: 300,
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
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Stack(
          children: [
            FlutterMap(
              mapController: _mapController,
              options: MapOptions(
                initialCenter: initialCenter,
                initialZoom: hasCustomerLocation && hasDeliveryLocation
                    ? 13.0
                    : 15.0,
                minZoom: 3.0,
                maxZoom: 18.0,
              ),
              children: [
                // 天地图矢量底图
                TileLayer(
                  urlTemplate: tiandituTileUrlTemplate,
                  subdomains: const ['0', '1', '2', '3', '4', '5', '6', '7'],
                  userAgentPackageName: 'com.example.employees_app',
                  maxNativeZoom: 18,
                  maxZoom: 18,
                  tileProvider: createTiandituTileProvider(),
                ),
                // 天地图矢量标注图层
                TileLayer(
                  urlTemplate: tiandituLabelUrlTemplate,
                  subdomains: const ['0', '1', '2', '3', '4', '5', '6', '7'],
                  userAgentPackageName: 'com.example.employees_app',
                  maxNativeZoom: 18,
                  maxZoom: 18,
                  tileProvider: createTiandituTileProvider(),
                ),
                // 路线（如果有两个位置）- 放在标记图层之前，这样线条会在图标下面
                if (hasCustomerLocation && hasDeliveryLocation)
                  PolylineLayer(
                    polylines: [
                      Polyline(
                        points: [
                          CoordinateTransform.gcj02ToWgs84(
                            (customerLat as num).toDouble(),
                            (customerLng as num).toDouble(),
                          ),
                          CoordinateTransform.gcj02ToWgs84(
                            (deliveryLocation['latitude'] as num).toDouble(),
                            (deliveryLocation['longitude'] as num).toDouble(),
                          ),
                        ],
                        strokeWidth: 3,
                        color: const Color(0xFF20CB6B).withOpacity(0.5),
                      ),
                    ],
                  ),
                // 标记图层
                MarkerLayer(
                  markers: [
                    // 客户位置标记
                    if (hasCustomerLocation)
                      Marker(
                        point: CoordinateTransform.gcj02ToWgs84(
                          (customerLat as num).toDouble(),
                          (customerLng as num).toDouble(),
                        ),
                        width: 28,
                        height: 28,
                        alignment: Alignment.center,
                        child: Container(
                          width: 28,
                          height: 28,
                          decoration: BoxDecoration(
                            color: const Color(0xFF20CB6B),
                            shape: BoxShape.circle,
                            border: Border.all(color: Colors.white, width: 2),
                            boxShadow: [
                              BoxShadow(
                                color: Colors.black.withOpacity(0.3),
                                blurRadius: 3,
                                offset: const Offset(0, 2),
                              ),
                            ],
                          ),
                          child: const Icon(
                            Icons.location_on,
                            color: Colors.white,
                            size: 16,
                          ),
                        ),
                      ),
                    // 配送员位置标记
                    if (hasDeliveryLocation)
                      Marker(
                        point: CoordinateTransform.gcj02ToWgs84(
                          (deliveryLocation['latitude'] as num).toDouble(),
                          (deliveryLocation['longitude'] as num).toDouble(),
                        ),
                        width: 28,
                        height: 28,
                        alignment: Alignment.center,
                        child: Container(
                          width: 28,
                          height: 28,
                          decoration: BoxDecoration(
                            color: const Color(0xFF4C8DF6),
                            shape: BoxShape.circle,
                            border: Border.all(color: Colors.white, width: 2),
                            boxShadow: [
                              BoxShadow(
                                color: Colors.black.withOpacity(0.3),
                                blurRadius: 3,
                                offset: const Offset(0, 2),
                              ),
                            ],
                          ),
                          child: const Icon(
                            Icons.local_shipping,
                            color: Colors.white,
                            size: 16,
                          ),
                        ),
                      ),
                  ],
                ),
              ],
            ),
            // 图例
            Positioned(
              top: 8,
              left: 8,
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 8,
                ),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.9),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (hasCustomerLocation)
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Container(
                            width: 12,
                            height: 12,
                            decoration: const BoxDecoration(
                              color: Color(0xFF20CB6B),
                              shape: BoxShape.circle,
                            ),
                          ),
                          const SizedBox(width: 6),
                          const Text(
                            '客户位置',
                            style: TextStyle(
                              fontSize: 12,
                              color: Color(0xFF20253A),
                            ),
                          ),
                        ],
                      ),
                    if (hasCustomerLocation && hasDeliveryLocation)
                      const SizedBox(height: 6),
                    if (hasDeliveryLocation)
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Container(
                            width: 12,
                            height: 12,
                            decoration: const BoxDecoration(
                              color: Color(0xFF4C8DF6),
                              shape: BoxShape.circle,
                            ),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            '配送员位置${deliveryLocation['is_realtime'] == true ? '（实时）' : '（历史）'}',
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF20253A),
                            ),
                          ),
                        ],
                      ),
                  ],
                ),
              ),
            ),
            // 加载提示
            if (_loadingDeliveryLocation)
              const Positioned.fill(
                child: Center(child: CircularProgressIndicator()),
              ),
          ],
        ),
      ),
    );
  }
}
