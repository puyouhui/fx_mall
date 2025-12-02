import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/customer/customer_profile_page.dart';
import 'package:employees_app/pages/order/sales_create_order_page.dart';

/// 客户详情页面：展示客户基础信息、统计信息、地址列表、最近订单等
class CustomerDetailPage extends StatefulWidget {
  final int customerId;
  final String? customerName;

  const CustomerDetailPage({
    super.key,
    required this.customerId,
    this.customerName,
  });

  @override
  State<CustomerDetailPage> createState() => _CustomerDetailPageState();
}

class _CustomerDetailPageState extends State<CustomerDetailPage> {
  bool _isLoading = true;
  String? _error;
  Map<String, dynamic>? _user;
  List<dynamic> _addresses = [];
  List<dynamic> _recentOrders = [];
  int _addressCount = 0;
  int _orderCount = 0;
  double _totalAmount = 0;

  @override
  void initState() {
    super.initState();
    _loadDetail();
  }

  Future<void> _loadDetail() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    final resp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customers/${widget.customerId}',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      final data = resp.data!;
      setState(() {
        _user = data['user'] as Map<String, dynamic>?;
        _addresses = (data['addresses'] as List<dynamic>? ?? []);
        _recentOrders = (data['recent_orders'] as List<dynamic>? ?? []);
        _addressCount = data['address_count'] as int? ?? _addresses.length;
        _orderCount = data['order_count'] as int? ?? 0;
        _totalAmount = (data['total_amount'] as num?)?.toDouble() ?? 0.0;
        _isLoading = false;
      });
    } else {
      setState(() {
        _error = resp.message.isNotEmpty ? resp.message : '获取客户详情失败';
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final name =
        widget.customerName ??
        (_user != null ? (_user!['name'] as String? ?? '客户详情') : '客户详情');

    return Scaffold(
      appBar: AppBar(
        title: Text(name),
        centerTitle: true,
        backgroundColor: const Color(0xFF20CB6B),
        elevation: 0,
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
          child: _isLoading
              ? const Center(
                  child: CircularProgressIndicator(
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                )
              : _error != null
              ? Center(
                  child: Text(
                    _error!,
                    style: const TextStyle(color: Colors.white),
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _loadDetail,
                  child: ListView(
                    padding: const EdgeInsets.all(16),
                    children: [
                      _buildBasicInfoCard(),
                      const SizedBox(height: 12),
                      _buildStatsCard(),
                      const SizedBox(height: 12),
                      _buildAddressCard(),
                      const SizedBox(height: 12),
                      _buildRecentOrdersCard(),
                    ],
                  ),
                ),
        ),
      ),
      // 底部操作区：改客户信息 + 创建新订单
      bottomNavigationBar: _buildBottomActions(),
    );
  }

  Widget _buildBasicInfoCard() {
    final user = _user ?? {};
    final name = user['name'] as String? ?? '未填写名称';
    final phone = user['phone'] as String? ?? '';
    final userCode = user['user_code'] as String? ?? '';
    final storeType = user['store_type'] as String? ?? '';
    final userType = user['user_type'] as String? ?? '';
    final createdAt = user['created_at']?.toString() ?? '';

    return _buildCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  name,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                ),
              ),
              if (userCode.isNotEmpty)
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xFF20CB6B).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Text(
                    '编号 $userCode',
                    style: const TextStyle(
                      fontSize: 11,
                      color: Color(0xFF20CB6B),
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: 8),
          if (phone.isNotEmpty)
            Row(
              children: [
                const Icon(Icons.phone, size: 16, color: Color(0xFF8C92A4)),
                const SizedBox(width: 4),
                Text(
                  phone,
                  style: const TextStyle(
                    fontSize: 14,
                    color: Color(0xFF40475C),
                  ),
                ),
              ],
            ),
          const SizedBox(height: 6),
          if (storeType.isNotEmpty || userType.isNotEmpty)
            Row(
              children: [
                if (storeType.isNotEmpty)
                  Text(
                    storeType,
                    style: const TextStyle(
                      fontSize: 13,
                      color: Color(0xFF8C92A4),
                    ),
                  ),
                if (storeType.isNotEmpty && userType.isNotEmpty)
                  const SizedBox(width: 12),
                if (userType.isNotEmpty)
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFF20CB6B).withOpacity(0.08),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      userType == 'wholesale' ? '批发客户' : '零售客户',
                      style: const TextStyle(
                        fontSize: 11,
                        color: Color(0xFF20CB6B),
                      ),
                    ),
                  ),
              ],
            ),
          if (createdAt.isNotEmpty) ...[
            const SizedBox(height: 8),
            Row(
              children: [
                const Icon(
                  Icons.access_time,
                  size: 14,
                  color: Color(0xFFB0B4C3),
                ),
                const SizedBox(width: 4),
                Text(
                  '绑定时间：${_formatTime(createdAt)}',
                  style: const TextStyle(
                    fontSize: 12,
                    color: Color(0xFFB0B4C3),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildStatsCard() {
    return _buildCard(
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  '历史下单总金额',
                  style: TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
                ),
                const SizedBox(height: 4),
                Text(
                  '¥${_totalAmount.toStringAsFixed(2)}',
                  style: const TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.w700,
                    color: Color(0xFFFF5A5F),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  '订单数量',
                  style: TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
                ),
                const SizedBox(height: 4),
                Text(
                  '$_orderCount 单',
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  '地址数量：$_addressCount',
                  style: const TextStyle(
                    fontSize: 12,
                    color: Color(0xFF8C92A4),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAddressCard() {
    if (_addresses.isEmpty) {
      return _buildCard(
        child: const Text(
          '暂无地址信息',
          style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
        ),
      );
    }

    return _buildCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '收货地址',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 8),
          ..._addresses.map((addr) {
            final map = addr as Map<String, dynamic>;
            final name = map['name'] as String? ?? '';
            final contact = map['contact'] as String? ?? '';
            final phone = map['phone'] as String? ?? '';
            final address = map['address'] as String? ?? '';
            final isDefault = (map['is_default'] as bool?) ?? false;

            return Container(
              margin: const EdgeInsets.only(bottom: 10),
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: const Color(0xFFF7F8FA),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          name.isNotEmpty ? name : '地址',
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                      ),
                      if (isDefault)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 6,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: const Color(0xFF20CB6B).withOpacity(0.1),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: const Text(
                            '默认地址',
                            style: TextStyle(
                              fontSize: 10,
                              color: Color(0xFF20CB6B),
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    address,
                    style: const TextStyle(
                      fontSize: 13,
                      color: Color(0xFF8C92A4),
                    ),
                  ),
                  if (contact.isNotEmpty || phone.isNotEmpty) ...[
                    const SizedBox(height: 2),
                    Text(
                      '$contact  $phone',
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFF8C92A4),
                      ),
                    ),
                  ],
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  Widget _buildRecentOrdersCard() {
    if (_recentOrders.isEmpty) {
      return _buildCard(
        child: const Text(
          '暂无历史订单',
          style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
        ),
      );
    }

    return _buildCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '最近购买记录（3单）',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 8),
          ..._recentOrders.map((item) {
            final order = item as Map<String, dynamic>;
            final orderNo = order['order_number'] as String? ?? '';
            final total = (order['total_amount'] as num?)?.toDouble() ?? 0.0;
            final status = order['status'] as String? ?? '';
            final createdAt = order['created_at']?.toString() ?? '';

            return Container(
              margin: const EdgeInsets.only(bottom: 8),
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: const Color(0xFFF7F8FA),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          orderNo,
                          style: const TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                      ),
                      Text(
                        '¥${total.toStringAsFixed(2)}',
                        style: const TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFFFF5A5F),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      Text(
                        _formatTime(createdAt),
                        style: const TextStyle(
                          fontSize: 11,
                          color: Color(0xFFB0B4C3),
                        ),
                      ),
                      const SizedBox(width: 12),
                      if (status.isNotEmpty)
                        Text(
                          status,
                          style: const TextStyle(
                            fontSize: 11,
                            color: Color(0xFF8C92A4),
                          ),
                        ),
                    ],
                  ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  /// 底部两个主要操作按钮
  Widget _buildBottomActions() {
    final user = _user;
    final userId = user?['id'] as int?;
    final userCode = user?['user_code'] as String? ?? '';

    final canEdit = userId != null && userId > 0;

    return Container(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
      decoration: const BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Color(0x14000000),
            blurRadius: 8,
            offset: Offset(0, -2),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: OutlinedButton.icon(
              onPressed: canEdit
                  ? () {
                      // 跳转到新客资料编辑页面（带上编号）
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) =>
                              CustomerProfilePage(initialUserCode: userCode),
                        ),
                      );
                    }
                  : null,
              icon: const Icon(Icons.edit, size: 18, color: Color(0xFF4C8DF6)),
              label: const Text(
                '改客户信息',
                style: TextStyle(
                  fontSize: 14,
                  color: Color(0xFF4C8DF6),
                  fontWeight: FontWeight.w500,
                ),
              ),
              style: OutlinedButton.styleFrom(
                side: const BorderSide(color: Color(0xFF4C8DF6)),
                padding: const EdgeInsets.symmetric(vertical: 10),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(22),
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: ElevatedButton.icon(
              onPressed: canEdit
                  ? () {
                      final user = _user;
                      if (user == null || user['id'] == null) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(content: Text('客户信息未加载完成')),
                        );
                        return;
                      }
                      final id = user['id'] as int;
                      final name = (user['name'] as String?) ?? widget.customerName;
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) => SalesCreateOrderPage(
                            customerId: id,
                            customerName: name,
                          ),
                        ),
                      );
                    }
                  : null,
              icon: const Icon(Icons.add_shopping_cart, size: 18),
              label: const Text(
                '创建新订单',
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF20CB6B),
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 12),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(22),
                ),
                elevation: 0,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCard({required Widget child}) {
    return Container(
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
      child: Padding(padding: const EdgeInsets.all(14), child: child),
    );
  }

  String _formatTime(String raw) {
    try {
      final dt = DateTime.tryParse(raw);
      if (dt == null) return raw;
      return DateFormat('yyyy-MM-dd HH:mm').format(dt.toLocal());
    } catch (_) {
      return raw;
    }
  }
}
