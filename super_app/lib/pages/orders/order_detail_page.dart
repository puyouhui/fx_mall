import 'package:flutter/material.dart';
import 'package:super_app/api/orders_api.dart';
import 'package:super_app/api/payment_verification_api.dart';
import 'package:intl/intl.dart';

class OrderDetailPage extends StatefulWidget {
  final int orderId;

  const OrderDetailPage({super.key, required this.orderId});

  @override
  State<OrderDetailPage> createState() => _OrderDetailPageState();
}

class _OrderDetailPageState extends State<OrderDetailPage> {
  Map<String, dynamic>? _orderDetail;
  bool _isLoading = false;
  Map<String, dynamic>? _paymentVerificationRequest; // 收款审核申请

  @override
  void initState() {
    super.initState();
    _loadOrderDetail();
  }

  Future<void> _loadOrderDetail() async {
    setState(() => _isLoading = true);

    try {
      final response = await OrdersApi.getOrderDetail(widget.orderId);

      if (!mounted) return;

      if (response.isSuccess && response.data != null) {
        setState(() => _orderDetail = response.data);
        
        // 查询是否有待审核的收款审核申请
        await _loadPaymentVerificationRequest();
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(response.message)),
          );
        }
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('加载失败: ${e.toString()}')),
      );
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  // 查询收款审核申请
  Future<void> _loadPaymentVerificationRequest() async {
    try {
      // 查询所有待审核的申请，查找是否有当前订单的申请
      final response = await PaymentVerificationApi.getPaymentVerifications(
        status: 'pending',
        pageNum: 1,
        pageSize: 100, // 获取足够多的数据以查找当前订单
      );

      if (!mounted) return;

      if (response.isSuccess && response.data != null) {
        final verificationList = response.data!.list;
        // 查找当前订单的申请
        Map<String, dynamic>? foundRequest;
        try {
          foundRequest = verificationList.firstWhere(
            (item) {
              final orderId = (item['order_id'] as num?)?.toInt();
              return orderId == widget.orderId;
            },
          ) as Map<String, dynamic>?;
        } catch (e) {
          // 没有找到对应的申请
          foundRequest = null;
        }

        if (mounted) {
          setState(() {
            _paymentVerificationRequest = foundRequest;
          });
        }
      } else {
        // 查询失败，清空申请
        if (mounted) {
          setState(() {
            _paymentVerificationRequest = null;
          });
        }
      }
    } catch (e) {
      // 查询失败不影响页面显示
      debugPrint('查询收款审核申请失败: $e');
      if (mounted) {
        setState(() {
          _paymentVerificationRequest = null;
        });
      }
    }
  }

  String _formatDateTime(dynamic value) {
    if (value == null) return '';
    if (value is String) {
      try {
        final dateTime = DateTime.parse(value);
        return DateFormat('yyyy-MM-dd HH:mm:ss').format(dateTime);
      } catch (e) {
        return value.toString();
      }
    } else if (value is DateTime) {
      return DateFormat('yyyy-MM-dd HH:mm:ss').format(value);
    }
    return value.toString();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('订单详情'),
        backgroundColor: const Color(0xFF20CB6B),
        foregroundColor: Colors.white,
      ),
      body: _isLoading && _orderDetail == null
          ? const Center(child: CircularProgressIndicator())
          : _orderDetail == null
              ? const Center(child: Text('暂无数据'))
              : Column(
                  children: [
                    Expanded(
                      child: RefreshIndicator(
                        onRefresh: _loadOrderDetail,
                        color: const Color(0xFF20CB6B),
                        child: ListView(
                          padding: const EdgeInsets.all(16),
                          children: [
                            _buildOrderInfoCard(),
                            const SizedBox(height: 16),
                            _buildProfitCard(),
                            const SizedBox(height: 16),
                            _buildSalesCommissionCard(),
                            const SizedBox(height: 16),
                            _buildDeliveryFeeCard(),
                            const SizedBox(height: 16),
                            _buildItemsCard(),
                            const SizedBox(height: 16),
                            _buildAddressCard(),
                            // 底部留出空间给审核按钮
                            if (_paymentVerificationRequest != null)
                              const SizedBox(height: 80),
                          ],
                        ),
                      ),
                    ),
                    // 审核按钮（固定在底部）
                    if (_paymentVerificationRequest != null)
                      _buildReviewButton(),
                  ],
                ),
    );
  }

  Widget _buildOrderInfoCard() {
    final order = _orderDetail!['order'] as Map<String, dynamic>?;
    if (order == null) return const SizedBox.shrink();

    return _buildCard(
      title: '订单信息',
      child: Column(
        children: [
          _buildInfoRow('订单号', order['order_number']?.toString() ?? ''),
          _buildInfoRow('订单状态', _getStatusText(order['status']?.toString() ?? '')),
          _buildInfoRow('创建时间', _formatDateTime(order['created_at'])),
          _buildInfoRow('商品金额', '¥${(order['goods_amount'] as num?)?.toStringAsFixed(2) ?? '0.00'}'),
          _buildInfoRow('配送费', '¥${(order['delivery_fee'] as num?)?.toStringAsFixed(2) ?? '0.00'}'),
          if (((order['urgent_fee'] as num?)?.toDouble() ?? 0.0) > 0)
            _buildInfoRow('加急费', '¥${(order['urgent_fee'] as num?)?.toStringAsFixed(2) ?? '0.00'}'),
          if (((order['coupon_discount'] as num?)?.toDouble() ?? 0.0) > 0)
            _buildInfoRow('优惠券抵扣', '-¥${(order['coupon_discount'] as num?)?.toStringAsFixed(2) ?? '0.00'}'),
          if (((order['points_discount'] as num?)?.toDouble() ?? 0.0) > 0)
            _buildInfoRow('积分抵扣', '-¥${(order['points_discount'] as num?)?.toStringAsFixed(2) ?? '0.00'}'),
          _buildInfoRow('实付金额', '¥${(order['total_amount'] as num?)?.toStringAsFixed(2) ?? '0.00'}', isHighlight: true),
        ],
      ),
    );
  }

  Widget _buildProfitCard() {
    final simplifiedProfit = _orderDetail!['simplified_profit'] as Map<String, dynamic>?;
    if (simplifiedProfit == null) return const SizedBox.shrink();

    final platformRevenue = (simplifiedProfit['platform_revenue'] as num?)?.toDouble() ?? 0.0;
    final goodsCost = (simplifiedProfit['goods_cost'] as num?)?.toDouble() ?? 0.0;
    final grossProfit = (simplifiedProfit['gross_profit'] as num?)?.toDouble() ?? 0.0;
    final deliveryCost = (simplifiedProfit['delivery_cost'] as num?)?.toDouble() ?? 0.0;
    final netProfit = (simplifiedProfit['net_profit'] as num?)?.toDouble() ?? 0.0;
    final order = _orderDetail!['order'] as Map<String, dynamic>?;

    return _buildCard(
      title: '利润信息',
      child: Column(
        children: [
          _buildProfitRow(
            '平台总收入（实付金额）',
            '¥${platformRevenue.toStringAsFixed(2)}',
            const Color(0xFF409EFF),
            subtitle: _buildRevenueSubtitle(order),
          ),
          const Divider(height: 24),
          _buildProfitRow(
            '商品总成本',
            '¥${goodsCost.toStringAsFixed(2)}',
            const Color(0xFF909399),
          ),
          const Divider(height: 24),
          _buildProfitRow(
            '毛利润（平台总收入 - 商品总成本）',
            '¥${grossProfit.toStringAsFixed(2)}',
            const Color(0xFF67C23A),
            subtitle: Text(
              '= 平台总收入(¥${platformRevenue.toStringAsFixed(2)}) - 商品总成本(¥${goodsCost.toStringAsFixed(2)})',
              style: const TextStyle(fontSize: 12, color: Color(0xFF909399)),
            ),
          ),
          const Divider(height: 24),
          _buildProfitRow(
            '配送成本',
            '¥${deliveryCost.toStringAsFixed(2)}',
            const Color(0xFFE6A23C),
          ),
          const Divider(height: 24),
          _buildProfitRow(
            '净利润（平台总收入 - 商品总成本 - 配送成本）',
            '¥${netProfit.toStringAsFixed(2)}',
            netProfit >= 0 ? const Color(0xFF67C23A) : const Color(0xFFF56C6C),
            subtitle: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '= 平台总收入(¥${platformRevenue.toStringAsFixed(2)}) - 商品总成本(¥${goodsCost.toStringAsFixed(2)}) - 配送成本(¥${deliveryCost.toStringAsFixed(2)})',
                  style: const TextStyle(fontSize: 12, color: Color(0xFF909399)),
                ),
                const SizedBox(height: 4),
                Text(
                  netProfit >= 0 ? '✓ 平台盈利' : '✗ 平台亏损',
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                    color: netProfit >= 0 ? const Color(0xFF67C23A) : const Color(0xFFF56C6C),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRevenueSubtitle(Map<String, dynamic>? order) {
    if (order == null) return const SizedBox.shrink();
    
    final goodsAmount = (order['goods_amount'] as num?)?.toDouble() ?? 0.0;
    final deliveryFee = (order['delivery_fee'] as num?)?.toDouble() ?? 0.0;
    final urgentFee = (order['urgent_fee'] as num?)?.toDouble() ?? 0.0;
    final couponDiscount = (order['coupon_discount'] as num?)?.toDouble() ?? 0.0;
    final pointsDiscount = (order['points_discount'] as num?)?.toDouble() ?? 0.0;

    final parts = <String>[];
    parts.add('商品金额(¥${goodsAmount.toStringAsFixed(2)})');
    parts.add('配送费(¥${deliveryFee.toStringAsFixed(2)})');
    if (urgentFee > 0) parts.add('加急费(¥${urgentFee.toStringAsFixed(2)})');
    if (couponDiscount > 0) parts.add('优惠券(-¥${couponDiscount.toStringAsFixed(2)})');
    if (pointsDiscount > 0) parts.add('积分(-¥${pointsDiscount.toStringAsFixed(2)})');

    return Text(
      parts.join(' + '),
      style: const TextStyle(fontSize: 12, color: Color(0xFF909399)),
    );
  }

  Widget _buildDeliveryFeeCard() {
    final deliveryFeeCalc = _orderDetail!['delivery_fee_calculation'] as Map<String, dynamic>?;
    if (deliveryFeeCalc == null || deliveryFeeCalc.isEmpty) return const SizedBox.shrink();

    final baseFee = (deliveryFeeCalc['base_fee'] as num?)?.toDouble() ?? 0.0;
    final itemFee = (deliveryFeeCalc['item_fee'] as num?)?.toDouble() ?? 0.0;
    final urgentFee = (deliveryFeeCalc['urgent_fee'] as num?)?.toDouble() ?? 0.0;
    final isolatedFee = (deliveryFeeCalc['isolated_fee'] as num?)?.toDouble() ?? 0.0;
    final weatherFee = (deliveryFeeCalc['weather_fee'] as num?)?.toDouble() ?? 0.0;
    final profitShare = (deliveryFeeCalc['profit_share'] as num?)?.toDouble() ?? 0.0;
    final riderPayableFee = (deliveryFeeCalc['rider_payable_fee'] as num?)?.toDouble() ?? 0.0;
    final totalPlatformCost = (deliveryFeeCalc['total_platform_cost'] as num?)?.toDouble() ?? 0.0;

    return _buildCard(
      title: '配送费计算详情',
      child: Column(
        children: [
          if (baseFee > 0) _buildInfoRow('基础配送费', '¥${baseFee.toStringAsFixed(2)}'),
          if (itemFee > 0) _buildInfoRow('计件费', '¥${itemFee.toStringAsFixed(2)}'),
          if (urgentFee > 0) _buildInfoRow('加急费', '¥${urgentFee.toStringAsFixed(2)}'),
          if (isolatedFee > 0) _buildInfoRow('隔离费', '¥${isolatedFee.toStringAsFixed(2)}'),
          if (weatherFee > 0) _buildInfoRow('天气附加费', '¥${weatherFee.toStringAsFixed(2)}'),
          if (profitShare > 0)
            _buildInfoRow(
              '利润分成',
              '¥${profitShare.toStringAsFixed(2)}',
              subtitle: const Text(
                '(已包含在配送员实际所得中)',
                style: TextStyle(fontSize: 12, color: Color(0xFF909399)),
              ),
            ),
          const Divider(height: 24),
          _buildInfoRow('配送员实际所得', '¥${riderPayableFee.toStringAsFixed(2)}', isHighlight: true),
          _buildInfoRow('平台总成本', '¥${totalPlatformCost.toStringAsFixed(2)}', isHighlight: true),
        ],
      ),
    );
  }

  Widget _buildItemsCard() {
    final items = _orderDetail!['order_items'] as List<dynamic>?;
    if (items == null || items.isEmpty) return const SizedBox.shrink();

    return _buildCard(
      title: '商品清单',
      child: Column(
        children: items.map<Widget>((item) {
          final itemMap = item as Map<String, dynamic>;
          return Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (itemMap['image'] != null && itemMap['image'].toString().isNotEmpty)
                  Container(
                    width: 60,
                    height: 60,
                    margin: const EdgeInsets.only(right: 12),
                    decoration: BoxDecoration(
                      borderRadius: BorderRadius.circular(8),
                      color: Colors.grey.shade200,
                    ),
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: Image.network(
                        itemMap['image'].toString(),
                        fit: BoxFit.cover,
                        errorBuilder: (context, error, stackTrace) {
                          return const Icon(Icons.image_not_supported, color: Colors.grey);
                        },
                      ),
                    ),
                  ),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        itemMap['product_name']?.toString() ?? '',
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF20253A),
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        itemMap['spec_name']?.toString() ?? '',
                        style: const TextStyle(
                          fontSize: 13,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                      const SizedBox(height: 4),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            'x${itemMap['quantity'] ?? 0}',
                            style: const TextStyle(
                              fontSize: 13,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                          Text(
                            '¥${((itemMap['subtotal'] as num?)?.toDouble() ?? 0.0).toStringAsFixed(2)}',
                            style: const TextStyle(
                              fontSize: 15,
                              fontWeight: FontWeight.w600,
                              color: Color(0xFF20CB6B),
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
        }).toList(),
      ),
    );
  }

  Widget _buildAddressCard() {
    final address = _orderDetail!['address'] as Map<String, dynamic>?;
    if (address == null) return const SizedBox.shrink();

    return _buildCard(
      title: '收货信息',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildInfoRow('收货人', address['name']?.toString() ?? ''),
          _buildInfoRow('联系电话', address['phone']?.toString() ?? ''),
          if (address['contact']?.toString().isNotEmpty ?? false)
            _buildInfoRow('联系人', address['contact']?.toString() ?? ''),
          _buildInfoRow('收货地址', address['address']?.toString() ?? ''),
        ],
      ),
    );
  }

  Widget _buildCard({
    required String title,
    required Widget child,
    Widget? titleAction,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200, width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(
                title,
                style: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              if (titleAction != null) ...[
                const Spacer(),
                titleAction,
              ],
            ],
          ),
          const SizedBox(height: 16),
          child,
        ],
      ),
    );
  }

  Widget _buildInfoRow(String label, String value, {bool isHighlight = false, Widget? subtitle}) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: TextStyle(
                  fontSize: 14,
                  color: Colors.grey.shade600,
                ),
              ),
              Expanded(
                child: Text(
                  value,
                  textAlign: TextAlign.right,
                  style: TextStyle(
                    fontSize: isHighlight ? 16 : 14,
                    fontWeight: isHighlight ? FontWeight.bold : FontWeight.normal,
                    color: isHighlight ? const Color(0xFF20CB6B) : const Color(0xFF20253A),
                  ),
                ),
              ),
            ],
          ),
          if (subtitle != null) ...[
            const SizedBox(height: 4),
            Padding(
              padding: const EdgeInsets.only(left: 0),
              child: subtitle,
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildProfitRow(String label, String value, Color valueColor, {Widget? subtitle}) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Text(
                  label,
                  style: TextStyle(
                    fontSize: 14,
                    color: Colors.grey.shade600,
                  ),
                ),
              ),
              Text(
                value,
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: valueColor,
                ),
              ),
            ],
          ),
          if (subtitle != null) ...[
            const SizedBox(height: 4),
            Padding(
              padding: const EdgeInsets.only(left: 0),
              child: subtitle,
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildSalesCommissionCard() {
    // 优先使用实际分成，如果没有则使用预览分成
    final salesCommission = _orderDetail!['sales_commission'] as Map<String, dynamic>?;
    final salesCommissionPreview = _orderDetail!['sales_commission_preview'] as Map<String, dynamic>?;
    
    final commission = salesCommission ?? salesCommissionPreview;
    if (commission == null) return const SizedBox.shrink();

    final isSettled = (salesCommission != null) && (commission['is_settled'] == true);
    final isValidOrder = commission['is_valid_order'] == true;
    final orderProfit = (commission['order_profit'] as num?)?.toDouble() ?? 0.0;
    final baseCommission = (commission['base_commission'] as num?)?.toDouble() ?? 0.0;
    final newCustomerBonus = (commission['new_customer_bonus'] as num?)?.toDouble() ?? 0.0;
    final tierCommission = (commission['tier_commission'] as num?)?.toDouble() ?? 0.0;
    final tierLevel = (commission['tier_level'] as num?)?.toInt() ?? 0;
    final isNewCustomerOrder = commission['is_new_customer_order'] == true || newCustomerBonus > 0;
    final totalCommission = (commission['total_commission'] as num?)?.toDouble() ?? 0.0;

    // 获取销售员信息
    final user = _orderDetail!['user'] as Map<String, dynamic>?;
    final salesEmployee = user?['sales_employee'] as Map<String, dynamic>?;
    final salesEmployeeName = salesEmployee?['name']?.toString() ?? '未知销售员';

    return _buildCard(
      title: '销售分成明细',
      titleAction: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        decoration: BoxDecoration(
          color: isSettled
              ? const Color(0xFF4C8DF6).withOpacity(0.1)
              : const Color(0xFFFFA940).withOpacity(0.1),
          borderRadius: BorderRadius.circular(6),
        ),
        child: Text(
          isSettled ? '已计入' : '预览',
          style: TextStyle(
            fontSize: 12,
            color: isSettled ? const Color(0xFF4C8DF6) : const Color(0xFFFFA940),
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (salesEmployeeName.isNotEmpty && salesEmployeeName != '未知销售员') ...[
            _buildInfoRow('销售员', salesEmployeeName),
            const SizedBox(height: 12),
          ],
          if (!isValidOrder) ...[
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xFFFF5A5F).withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Row(
                children: [
                  Icon(Icons.info_outline, size: 16, color: Color(0xFFFF5A5F)),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '无效订单（利润需>5元才计入有效分成）',
                      style: TextStyle(fontSize: 12, color: Color(0xFFFF5A5F)),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
          ],
          _buildCommissionRow(
            '订单利润',
            '¥${orderProfit.toStringAsFixed(2)}',
            subtitle: _buildCommissionSubtitle(
              '订单金额 - 商品总成本 - 配送成本',
              orderProfit,
            ),
          ),
          const SizedBox(height: 12),
          _buildCommissionRow(
            '基础提成（45%）',
            '¥${baseCommission.toStringAsFixed(2)}',
            subtitle: _buildCommissionSubtitle(
              '订单利润 × 45%',
              baseCommission,
            ),
          ),
          if (isNewCustomerOrder) ...[
            const SizedBox(height: 12),
            _buildCommissionRow(
              '新客开发激励（20%）',
              '¥${newCustomerBonus.toStringAsFixed(2)}',
              highlight: true,
              subtitle: _buildCommissionSubtitle(
                '订单利润 × 20%',
                newCustomerBonus,
              ),
            ),
          ],
          if (tierLevel > 0) ...[
            const SizedBox(height: 12),
            _buildCommissionRow(
              '阶梯提成（阶梯$tierLevel）',
              '¥${tierCommission.toStringAsFixed(2)}',
              highlight: true,
              subtitle: _buildCommissionSubtitle(
                '根据销售员等级计算的阶梯提成',
                tierCommission,
              ),
            ),
          ],
          const Divider(height: 24, thickness: 0.5, color: Color(0xFFE5E7F0)),
          Row(
            children: [
              Text(
                isSettled ? '总分成' : '预计总分成',
                style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              Text(
                '¥${totalCommission.toStringAsFixed(2)}',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: isSettled
                      ? const Color(0xFF4C8DF6)
                      : const Color(0xFFFFA940),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildCommissionRow(
    String label,
    String value, {
    bool highlight = false,
    Widget? subtitle,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              label,
              style: TextStyle(
                fontSize: 14,
                color: highlight ? const Color(0xFFFFA940) : Colors.grey.shade600,
                fontWeight: highlight ? FontWeight.w600 : FontWeight.normal,
              ),
            ),
            Text(
              value,
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w600,
                color: highlight ? const Color(0xFFFFA940) : const Color(0xFF20253A),
              ),
            ),
          ],
        ),
        if (subtitle != null) ...[
          const SizedBox(height: 4),
          Padding(
            padding: const EdgeInsets.only(left: 0),
            child: subtitle,
          ),
        ],
      ],
    );
  }

  Widget _buildCommissionSubtitle(String text, double value) {
    return Text(
      text,
      style: TextStyle(
        fontSize: 12,
        color: Colors.grey.shade600,
      ),
    );
  }

  Widget _buildReviewButton() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 8,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: SafeArea(
        child: SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: _showReviewDialog,
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF20CB6B),
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(vertical: 16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              elevation: 0,
            ),
            child: const Text(
              '处理审核',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ),
      ),
    );
  }

  void _showReviewDialog() {
    showDialog(
      context: context,
      builder: (context) => _ReviewDialog(
        paymentVerificationRequest: _paymentVerificationRequest,
        onReview: (approved, remark) async {
          Navigator.of(context).pop();
          await _reviewPaymentVerification(approved, remark);
        },
      ),
    );
  }

  Future<void> _reviewPaymentVerification(bool approved, String remark) async {
    if (_paymentVerificationRequest == null) return;

    final requestId = (_paymentVerificationRequest!['id'] as num?)?.toInt();
    if (requestId == null) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('无法获取审核申请ID')),
        );
      }
      return;
    }

    try {
      final response = await PaymentVerificationApi.reviewPaymentVerification(
        requestId,
        approved,
        remark.isEmpty ? null : remark,
      );

      if (!mounted) return;

      if (response.isSuccess) {
        // 先清空审核申请，这样按钮会立即隐藏
        setState(() {
          _paymentVerificationRequest = null;
        });

        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(response.message),
              backgroundColor: Colors.green,
              duration: const Duration(seconds: 2),
            ),
          );
        }

        // 重新加载订单详情（订单状态可能已改变）
        await _loadOrderDetail();
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(response.message),
              backgroundColor: Colors.red,
            ),
          );
        }
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('审核失败: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  String _getStatusText(String status) {
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
}

// 审核对话框
class _ReviewDialog extends StatefulWidget {
  final Map<String, dynamic>? paymentVerificationRequest;
  final Future<void> Function(bool approved, String remark) onReview;

  const _ReviewDialog({
    required this.paymentVerificationRequest,
    required this.onReview,
  });

  @override
  State<_ReviewDialog> createState() => _ReviewDialogState();
}

class _ReviewDialogState extends State<_ReviewDialog> {
  final TextEditingController _remarkController = TextEditingController();
  bool _isSubmitting = false;

  @override
  void dispose() {
    _remarkController.dispose();
    super.dispose();
  }

  Future<void> _handleReview(bool approved) async {
    if (_isSubmitting) return;

    setState(() => _isSubmitting = true);

    try {
      await widget.onReview(approved, _remarkController.text.trim());
    } finally {
      if (mounted) {
        setState(() => _isSubmitting = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('审核收款申请'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (widget.paymentVerificationRequest != null) ...[
              Text(
                '订单号：${widget.paymentVerificationRequest!['order_number'] ?? ''}',
                style: const TextStyle(fontSize: 14),
              ),
              const SizedBox(height: 8),
              Text(
                '订单金额：¥${((widget.paymentVerificationRequest!['order_amount'] as num?)?.toDouble() ?? 0.0).toStringAsFixed(2)}',
                style: const TextStyle(fontSize: 14),
              ),
              if (widget.paymentVerificationRequest!['request_reason'] != null &&
                  widget.paymentVerificationRequest!['request_reason'].toString().isNotEmpty) ...[
                const SizedBox(height: 8),
                Text(
                  '申请原因：${widget.paymentVerificationRequest!['request_reason']}',
                  style: const TextStyle(fontSize: 14),
                ),
              ],
              const SizedBox(height: 16),
            ],
            const Text(
              '审核备注（选填）',
              style: TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _remarkController,
              decoration: InputDecoration(
                hintText: '请输入审核备注',
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
                contentPadding: const EdgeInsets.all(12),
              ),
              maxLines: 3,
              enabled: !_isSubmitting,
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: _isSubmitting ? null : () => Navigator.of(context).pop(),
          child: const Text('取消'),
        ),
        TextButton(
          onPressed: _isSubmitting
              ? null
              : () async {
                  await _handleReview(false);
                },
          style: TextButton.styleFrom(
            foregroundColor: Colors.red,
          ),
          child: const Text('拒绝'),
        ),
        ElevatedButton(
          onPressed: _isSubmitting
              ? null
              : () async {
                  await _handleReview(true);
                },
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFF20CB6B),
            foregroundColor: Colors.white,
          ),
          child: _isSubmitting
              ? const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                )
              : const Text('通过'),
        ),
      ],
    );
  }
}

