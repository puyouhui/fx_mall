import 'package:flutter/material.dart';
import '../api/order_api.dart';
import 'order_detail_view.dart';

/// 订单列表Tab组件：用于显示不同状态的订单列表
class OrderListTab extends StatefulWidget {
  const OrderListTab({
    super.key,
    required this.status,
    required this.onOrderAccepted,
  });

  final String? status; // 订单状态：null=新任务, 'pending_pickup'=待取货, 'delivering'=配送中
  final VoidCallback onOrderAccepted; // 接单成功回调

  @override
  State<OrderListTab> createState() => _OrderListTabState();
}

class _OrderListTabState extends State<OrderListTab> {
  final ScrollController _scrollController = ScrollController();
  final List<Map<String, dynamic>> _orders = [];
  final Map<int, bool> _acceptingOrders = {};
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

    final response = await OrderApi.getOrderPool(
      pageNum: _pageNum,
      pageSize: _pageSize,
      status: widget.status,
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
      });
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              response.message.isNotEmpty ? response.message : '获取订单列表失败',
            ),
          ),
        );
      }
    }

    if (mounted) {
      setState(() {
        _isLoading = false;
        _isLoadingMore = false;
      });
    }
  }

  Future<void> _acceptOrder(int orderId) async {
    if (_acceptingOrders[orderId] == true) return;

    setState(() {
      _acceptingOrders[orderId] = true;
    });

    final response = await OrderApi.acceptOrder(orderId);

    if (!mounted) return;

    if (response.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('接单成功'),
          backgroundColor: Color(0xFF20CB6B),
        ),
      );
      // 从列表中移除已接单的订单
      setState(() {
        _orders.removeWhere(
          (order) => (order['id'] as num?)?.toInt() == orderId,
        );
        _acceptingOrders.remove(orderId);
      });
      // 通知父组件刷新
      widget.onOrderAccepted();
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(response.message), backgroundColor: Colors.red),
      );
      setState(() {
        _acceptingOrders.remove(orderId);
      });
    }
  }

  void _viewOrderItems(Map<String, dynamic> order) {
    final orderId = (order['id'] as num?)?.toInt();

    if (orderId == null) return;

    // 跳转到订单详情页面
    Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => OrderDetailView(orderId: orderId)),
    );
  }

  @override
  Widget build(BuildContext context) {
    return _isLoading && _orders.isEmpty
        ? const Center(
            child: CircularProgressIndicator(
              valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
            ),
          )
        : RefreshIndicator(
            onRefresh: () => _loadOrders(reset: true),
            child: _orders.isEmpty
                ? LayoutBuilder(
                    builder: (context, constraints) {
                      return SingleChildScrollView(
                        padding: EdgeInsets.fromLTRB(
                          16,
                          16,
                          16,
                          16 + MediaQuery.of(context).padding.bottom,
                        ),
                        child: ConstrainedBox(
                          constraints: BoxConstraints(
                            minHeight:
                                constraints.maxHeight -
                                32 -
                                MediaQuery.of(context).padding.bottom,
                          ),
                          child: Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                const Icon(
                                  Icons.inbox_outlined,
                                  size: 42,
                                  color: Color(0xFF20CB6B),
                                ),
                                const Text(
                                  '暂无订单',
                                  style: TextStyle(
                                    fontSize: 18,
                                    fontWeight: FontWeight.w600,
                                    color: Color.fromARGB(193, 255, 255, 255),
                                  ),
                                ),
                                const SizedBox(height: 8),
                                Text(
                                  widget.status == null
                                      ? '新的订单将在这里显示'
                                      : widget.status == 'pending_pickup'
                                      ? '暂无待取货的订单'
                                      : '暂无配送中的订单',
                                  style: const TextStyle(
                                    fontSize: 14,
                                    color: Color.fromARGB(193, 255, 255, 255),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      );
                    },
                  )
                : ListView.builder(
                    controller: _scrollController,
                    padding: EdgeInsets.fromLTRB(
                      16,
                      0,
                      16,
                      16 + MediaQuery.of(context).padding.bottom,
                    ),
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
                      return _buildOrderCard(order);
                    },
                  ),
          );
  }

  Widget _buildOrderCard(Map<String, dynamic> order) {
    final orderId = (order['id'] as num?)?.toInt();
    final itemCount = (order['item_count'] as int?) ?? 0;
    final addressData = order['address'] as Map<String, dynamic>?;
    final storeName = addressData?['name'] as String? ?? '门店名称未填写';
    final address = addressData?['address'] as String? ?? '';

    // 加急状态
    final isUrgent = (order['is_urgent'] as bool?) ?? false;
    final urgentFee = (order['urgent_fee'] as num?)?.toDouble() ?? 0.0;

    // 配送费计算结果
    final deliveryFeeCalc =
        order['delivery_fee_calculation'] as Map<String, dynamic>?;
    final riderPayableFee =
        (deliveryFeeCalc?['rider_payable_fee'] as num?)?.toDouble() ?? 0.0;

    final isAccepting = orderId != null && (_acceptingOrders[orderId] == true);
    final showAcceptButton = widget.status == null; // 只有新任务显示接单按钮

    return Container(
      margin: const EdgeInsets.only(top: 12),
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
          // 点击查看商品列表
          InkWell(
            onTap: () => _viewOrderItems(order),
            borderRadius: const BorderRadius.only(
              topLeft: Radius.circular(16),
              topRight: Radius.circular(16),
            ),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // 地址（第一行，突出显示）
                  if (address.isNotEmpty)
                    Text(
                      address,
                      style: const TextStyle(
                        fontSize: 17,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF20253A),
                        height: 1.4,
                      ),
                      maxLines: 3,
                      overflow: TextOverflow.ellipsis,
                    ),
                  const SizedBox(height: 12),
                  // 门店名称
                  Text(
                    storeName,
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                      color: Color(0xFF40475C),
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 12),
                  // 商品数量和查看提示
                  Row(
                    children: [
                      const Icon(
                        Icons.shopping_cart_outlined,
                        size: 14,
                        color: Color(0xFF8C92A4),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '共$itemCount件商品',
                        style: const TextStyle(
                          fontSize: 13,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                      const Spacer(),
                      const Text(
                        '查看详情',
                        style: TextStyle(
                          fontSize: 12,
                          color: Color(0xFF20CB6B),
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                      const SizedBox(width: 2),
                      const Icon(
                        Icons.chevron_right,
                        size: 16,
                        color: Color(0xFF20CB6B),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  // 配送金额和加急状态（突出显示）
                  Row(
                    children: [
                      // 配送金额（突出显示）
                      if (deliveryFeeCalc != null)
                        Expanded(
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Icon(
                                Icons.local_shipping,
                                size: 18,
                                color: Color(0xFF20CB6B),
                              ),
                              const SizedBox(width: 6),
                              Text(
                                '¥${riderPayableFee.toStringAsFixed(2)}',
                                style: const TextStyle(
                                  fontSize: 20,
                                  fontWeight: FontWeight.w700,
                                  color: Color(0xFF20CB6B),
                                ),
                              ),
                            ],
                          ),
                        ),
                      // 加急状态标签
                      if (isUrgent) ...[
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 8,
                          ),
                          decoration: BoxDecoration(
                            color: const Color(0xFFFF6B6B).withOpacity(0.1),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Icon(
                                Icons.flash_on,
                                size: 16,
                                color: Color(0xFFFF6B6B),
                              ),
                              const SizedBox(width: 4),
                              Text(
                                urgentFee > 0
                                    ? '加急 +¥${urgentFee.toStringAsFixed(2)}'
                                    : '加急',
                                style: const TextStyle(
                                  fontSize: 13,
                                  fontWeight: FontWeight.w700,
                                  color: Color(0xFFFF6B6B),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ],
                  ),
                ],
              ),
            ),
          ),
          // 接单按钮（仅新任务显示）
          if (showAcceptButton)
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 0, 12, 12),
              child: SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: orderId != null && !isAccepting
                      ? () => _acceptOrder(orderId)
                      : null,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF20CB6B),
                    foregroundColor: Colors.white,
                    disabledBackgroundColor: const Color(0xFF9EDFB9),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    elevation: 0,
                  ),
                  child: isAccepting
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            valueColor: AlwaysStoppedAnimation<Color>(
                              Colors.white,
                            ),
                          ),
                        )
                      : const Text(
                          '接单',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
