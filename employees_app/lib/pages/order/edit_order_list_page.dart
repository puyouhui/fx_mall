import 'package:flutter/material.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/order/order_detail_page.dart';

/// 修改订单列表页面（只显示可修改的订单：pending_delivery状态）
class EditOrderListPage extends StatefulWidget {
  const EditOrderListPage({super.key});

  @override
  State<EditOrderListPage> createState() => _EditOrderListPageState();
}

class _EditOrderListPageState extends State<EditOrderListPage> {
  final ScrollController _scrollController = ScrollController();

  final List<Map<String, dynamic>> _orders = [];
  bool _isLoading = false;
  bool _isLoadingMore = false;
  bool _hasMore = true;
  int _pageNum = 1;
  final int _pageSize = 20;

  @override
  void initState() {
    super.initState();
    _loadOrders(reset: true);
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
            _scrollController.position.maxScrollExtent - 200 &&
        !_isLoadingMore &&
        _hasMore &&
        !_isLoading) {
      _loadOrders();
    }
  }

  Future<void> _loadOrders({bool reset = false}) async {
    if (_isLoading || _isLoadingMore) return;

    if (reset) {
      setState(() {
        _isLoading = true;
        _pageNum = 1;
        _hasMore = true;
        _orders.clear();
      });
    } else {
      setState(() {
        _isLoadingMore = true;
      });
    }

    final response = await Request.get<Map<String, dynamic>>(
      '/employee/sales/orders',
      queryParams: {
        'pageNum': _pageNum.toString(),
        'pageSize': _pageSize.toString(),
        'status': 'pending_delivery', // 只获取待配送的订单
      },
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final List<dynamic> list = (data['list'] as List<dynamic>? ?? []);
      final orders = list.cast<Map<String, dynamic>>();

      setState(() {
        if (reset) {
          _orders
            ..clear()
            ..addAll(orders);
        } else {
          _orders.addAll(orders);
        }
        final total = data['total'] as int? ?? _orders.length;
        _hasMore = _orders.length < total;
        if (_hasMore) {
          _pageNum++;
        }
        _isLoading = false;
        _isLoadingMore = false;
      });
    } else {
      setState(() {
        _isLoading = false;
        _isLoadingMore = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              response.message.isNotEmpty ? response.message : '加载订单列表失败',
            ),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  String _formatStatus(String? status) {
    switch (status) {
      case 'pending_delivery':
        return '待配送';
      case 'pending':
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
        return status ?? '未知';
    }
  }

  String _formatMoney(num? value) {
    final v = (value ?? 0).toDouble();
    return v.toStringAsFixed(2);
  }

  String _formatDateTime(String? dateTimeStr) {
    if (dateTimeStr == null || dateTimeStr.isEmpty) return '';
    try {
      final dt = DateTime.parse(dateTimeStr);
      return '${dt.month.toString().padLeft(2, '0')}-${dt.day.toString().padLeft(2, '0')} ${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
    } catch (_) {
      return dateTimeStr;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF5F6FA),
      appBar: AppBar(
        title: const Text('修改订单'),
        backgroundColor: const Color(0xFF20CB6B),
        foregroundColor: Colors.white,
        elevation: 0,
      ),
      body: RefreshIndicator(
        onRefresh: () => _loadOrders(reset: true),
        child: _isLoading && _orders.isEmpty
            ? const Center(
                child: CircularProgressIndicator(
                  valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
                ),
              )
            : _orders.isEmpty
            ? Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      Icons.edit_note_outlined,
                      size: 64,
                      color: Colors.grey[400],
                    ),
                    const SizedBox(height: 16),
                    Text(
                      '暂无可修改的订单',
                      style: TextStyle(fontSize: 16, color: Colors.grey[600]),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      '只有待配送状态的订单可以修改',
                      style: TextStyle(fontSize: 14, color: Colors.grey[500]),
                    ),
                  ],
                ),
              )
            : ListView.builder(
                controller: _scrollController,
                padding: const EdgeInsets.all(16),
                itemCount: _orders.length + (_hasMore ? 1 : 0),
                itemBuilder: (context, index) {
                  if (index >= _orders.length) {
                    return const Padding(
                      padding: EdgeInsets.symmetric(vertical: 16),
                      child: Center(
                        child: CircularProgressIndicator(
                          valueColor: AlwaysStoppedAnimation<Color>(
                            Color(0xFF20CB6B),
                          ),
                        ),
                      ),
                    );
                  }

                  final order = _orders[index];
                  final orderId = order['id'] as int?;
                  final orderNumber = order['order_number'] as String? ?? '';
                  final totalAmount = order['total_amount'] as num? ?? 0;
                  final status = order['status'] as String? ?? '';
                  final createdAt = order['created_at'] as String?;
                  final storeName = order['store_name'] as String? ?? '未知门店';
                  final address = order['address'] as String? ?? '';
                  final isLocked = (order['is_locked'] as bool?) ?? false;
                  final lockedBy = order['locked_by'] as String?;
                  final isUrgent = (order['is_urgent'] as bool?) ?? false;

                  return InkWell(
                    onTap: orderId != null
                        ? () {
                            Navigator.of(context)
                                .push(
                                  MaterialPageRoute(
                                    builder: (_) =>
                                        OrderDetailPage(orderId: orderId),
                                  ),
                                )
                                .then((_) {
                                  // 返回时刷新列表
                                  _loadOrders(reset: true);
                                });
                          }
                        : null,
                    borderRadius: BorderRadius.circular(16),
                    child: Container(
                      margin: const EdgeInsets.only(bottom: 12),
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(16),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withOpacity(0.04),
                            blurRadius: 8,
                            offset: const Offset(0, 2),
                          ),
                        ],
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          // 订单编号和状态
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Expanded(
                                child: Row(
                                  children: [
                                    Expanded(
                                      child: Text(
                                        '订单号：$orderNumber',
                                        style: const TextStyle(
                                          fontSize: 14,
                                          fontWeight: FontWeight.w600,
                                          color: Color(0xFF20253A),
                                        ),
                                      ),
                                    ),
                                    if (isUrgent) ...[
                                      const SizedBox(width: 6),
                                      Container(
                                        padding: const EdgeInsets.symmetric(
                                          horizontal: 6,
                                          vertical: 2,
                                        ),
                                        decoration: BoxDecoration(
                                          color: const Color(0xFFFF6B6B),
                                          borderRadius: BorderRadius.circular(
                                            4,
                                          ),
                                        ),
                                        child: const Text(
                                          '加急',
                                          style: TextStyle(
                                            fontSize: 10,
                                            fontWeight: FontWeight.w600,
                                            color: Colors.white,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                              ),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                  vertical: 4,
                                ),
                                decoration: BoxDecoration(
                                  color: const Color(
                                    0xFFFFA940,
                                  ).withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                child: Text(
                                  _formatStatus(status),
                                  style: const TextStyle(
                                    fontSize: 12,
                                    color: Color(0xFFFFA940),
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 12),
                          // 门店名称
                          Row(
                            children: [
                              const Icon(
                                Icons.store,
                                size: 16,
                                color: Color(0xFF8C92A4),
                              ),
                              const SizedBox(width: 6),
                              Expanded(
                                child: Text(
                                  storeName,
                                  style: const TextStyle(
                                    fontSize: 14,
                                    color: Color(0xFF20253A),
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                              ),
                            ],
                          ),
                          if (address.isNotEmpty) ...[
                            const SizedBox(height: 6),
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                const Icon(
                                  Icons.location_on_outlined,
                                  size: 16,
                                  color: Color(0xFF8C92A4),
                                ),
                                const SizedBox(width: 6),
                                Expanded(
                                  child: Text(
                                    address,
                                    style: const TextStyle(
                                      fontSize: 13,
                                      color: Color(0xFF8C92A4),
                                    ),
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                ),
                              ],
                            ),
                          ],
                          const SizedBox(height: 12),
                          // 金额和时间
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(
                                '¥${_formatMoney(totalAmount)}',
                                style: const TextStyle(
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
                                  color: Color(0xFF20CB6B),
                                ),
                              ),
                              Text(
                                _formatDateTime(createdAt),
                                style: const TextStyle(
                                  fontSize: 12,
                                  color: Color(0xFF8C92A4),
                                ),
                              ),
                            ],
                          ),
                          // 锁定提示
                          if (isLocked) ...[
                            const SizedBox(height: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 6,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.orange.withOpacity(0.1),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: Row(
                                children: [
                                  const Icon(
                                    Icons.lock_outline,
                                    size: 14,
                                    color: Colors.orange,
                                  ),
                                  const SizedBox(width: 6),
                                  Text(
                                    lockedBy != null
                                        ? '正在被修改中（员工：$lockedBy）'
                                        : '正在被修改中',
                                    style: const TextStyle(
                                      fontSize: 12,
                                      color: Colors.orange,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  );
                },
              ),
      ),
    );
  }
}
