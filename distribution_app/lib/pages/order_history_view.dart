import 'package:flutter/material.dart';
import '../api/order_api.dart';
import 'order_detail_view.dart';

/// 历史订单页面：显示配送员的所有已完成订单，分为今日和之前两个tab
class OrderHistoryView extends StatefulWidget {
  const OrderHistoryView({super.key});

  @override
  State<OrderHistoryView> createState() => _OrderHistoryViewState();
}

class _OrderHistoryViewState extends State<OrderHistoryView>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  final ScrollController _todayScrollController = ScrollController();
  final ScrollController _previousScrollController = ScrollController();

  // 今日订单
  final List<Map<String, dynamic>> _todayOrders = [];
  bool _isLoadingToday = false;
  bool _isLoadingMoreToday = false;
  bool _hasMoreToday = true;
  int _pageNumToday = 1;
  final int _pageSize = 20;

  // 之前订单
  final List<Map<String, dynamic>> _previousOrders = [];
  bool _isLoadingPrevious = false;
  bool _isLoadingMorePrevious = false;
  bool _hasMorePrevious = true;
  int _pageNumPrevious = 1;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _tabController.addListener(_onTabChanged);
    _todayScrollController.addListener(
      () => _onScroll(_todayScrollController, true),
    );
    _previousScrollController.addListener(
      () => _onScroll(_previousScrollController, false),
    );
    _loadTodayOrders(reset: true);
  }

  @override
  void dispose() {
    _tabController.removeListener(_onTabChanged);
    _tabController.dispose();
    _todayScrollController.dispose();
    _previousScrollController.dispose();
    super.dispose();
  }

  void _onTabChanged() {
    if (_tabController.index == 0 && _todayOrders.isEmpty && !_isLoadingToday) {
      _loadTodayOrders(reset: true);
    } else if (_tabController.index == 1 &&
        _previousOrders.isEmpty &&
        !_isLoadingPrevious) {
      _loadPreviousOrders(reset: true);
    }
  }

  void _onScroll(ScrollController controller, bool isToday) {
    if (controller.position.pixels >=
            controller.position.maxScrollExtent - 200 &&
        !(isToday ? _isLoadingMoreToday : _isLoadingMorePrevious) &&
        (isToday ? _hasMoreToday : _hasMorePrevious) &&
        !(isToday ? _isLoadingToday : _isLoadingPrevious)) {
      if (isToday) {
        _loadTodayOrders();
      } else {
        _loadPreviousOrders();
      }
    }
  }

  // 获取今日订单
  Future<void> _loadTodayOrders({bool reset = false}) async {
    if (_isLoadingToday || _isLoadingMoreToday) return;

    if (reset) {
      setState(() {
        _isLoadingToday = true;
        _pageNumToday = 1;
        _hasMoreToday = true;
        _todayOrders.clear();
      });
    } else {
      setState(() {
        _isLoadingMoreToday = true;
      });
    }

    // 获取所有已完成的订单，然后在前端筛选今日的
    final response = await OrderApi.getHistoryOrders(
      pageNum: _pageNumToday,
      pageSize: _pageSize,
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final List<dynamic> list = (data['list'] as List<dynamic>? ?? []);
      final allOrders = list.cast<Map<String, dynamic>>();

      // 筛选今日订单
      final today = DateTime.now();
      final todayOrders = allOrders.where((order) {
        final createdAt = order['created_at'] as String?;
        if (createdAt == null) return false;
        try {
          final orderDate = DateTime.parse(createdAt);
          return orderDate.year == today.year &&
              orderDate.month == today.month &&
              orderDate.day == today.day;
        } catch (e) {
          return false;
        }
      }).toList();

      setState(() {
        if (reset) {
          _todayOrders
            ..clear()
            ..addAll(todayOrders);
        } else {
          _todayOrders.addAll(todayOrders);
        }
        // 如果返回的订单数少于pageSize，说明没有更多了
        _hasMoreToday = allOrders.length >= _pageSize && todayOrders.isNotEmpty;
        if (_hasMoreToday) {
          _pageNumToday++;
        }
        _isLoadingToday = false;
        _isLoadingMoreToday = false;
      });
    } else {
      if (mounted) {
        setState(() {
          _isLoadingToday = false;
          _isLoadingMoreToday = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              response.message.isNotEmpty ? response.message : '获取订单列表失败',
            ),
          ),
        );
      }
    }
  }

  // 获取之前订单
  Future<void> _loadPreviousOrders({bool reset = false}) async {
    if (_isLoadingPrevious || _isLoadingMorePrevious) return;

    if (reset) {
      setState(() {
        _isLoadingPrevious = true;
        _pageNumPrevious = 1;
        _hasMorePrevious = true;
        _previousOrders.clear();
      });
    } else {
      setState(() {
        _isLoadingMorePrevious = true;
      });
    }

    // 获取所有已完成的订单，然后在前端筛选之前的
    final response = await OrderApi.getHistoryOrders(
      pageNum: _pageNumPrevious,
      pageSize: _pageSize,
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final List<dynamic> list = (data['list'] as List<dynamic>? ?? []);
      final allOrders = list.cast<Map<String, dynamic>>();

      // 筛选之前的订单（非今日）
      final today = DateTime.now();
      final previousOrders = allOrders.where((order) {
        final createdAt = order['created_at'] as String?;
        if (createdAt == null) return false;
        try {
          final orderDate = DateTime.parse(createdAt);
          return !(orderDate.year == today.year &&
              orderDate.month == today.month &&
              orderDate.day == today.day);
        } catch (e) {
          return false;
        }
      }).toList();

      setState(() {
        if (reset) {
          _previousOrders
            ..clear()
            ..addAll(previousOrders);
        } else {
          _previousOrders.addAll(previousOrders);
        }
        // 如果返回的订单数少于pageSize，说明没有更多了
        _hasMorePrevious =
            allOrders.length >= _pageSize && previousOrders.isNotEmpty;
        if (_hasMorePrevious) {
          _pageNumPrevious++;
        }
        _isLoadingPrevious = false;
        _isLoadingMorePrevious = false;
      });
    } else {
      if (mounted) {
        setState(() {
          _isLoadingPrevious = false;
          _isLoadingMorePrevious = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              response.message.isNotEmpty ? response.message : '获取订单列表失败',
            ),
          ),
        );
      }
    }
  }

  void _viewOrderDetail(Map<String, dynamic> order) async {
    final orderId = (order['id'] as num?)?.toInt();
    if (orderId == null) return;

    await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => OrderDetailView(orderId: orderId)),
    );

    // 返回后刷新当前tab的订单列表
    if (mounted) {
      if (_tabController.index == 0) {
        await _loadTodayOrders(reset: true);
      } else {
        await _loadPreviousOrders(reset: true);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: const Text(
          '历史订单',
          style: TextStyle(
            fontSize: 18,
            fontWeight: FontWeight.w600,
            color: Color(0xFF20253A),
          ),
        ),
        backgroundColor: Colors.white,
        elevation: 0,
        bottom: TabBar(
          controller: _tabController,
          labelColor: const Color(0xFF20CB6B),
          unselectedLabelColor: const Color(0xFF8C92A4),
          indicatorColor: const Color(0xFF20CB6B),
          indicatorWeight: 3,
          labelStyle: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
          ),
          unselectedLabelStyle: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w500,
          ),
          tabs: const [
            Tab(text: '今日'),
            Tab(text: '之前'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildOrderList(_todayOrders, _todayScrollController, true),
          _buildOrderList(_previousOrders, _previousScrollController, false),
        ],
      ),
    );
  }

  Widget _buildOrderList(
    List<Map<String, dynamic>> orders,
    ScrollController scrollController,
    bool isToday,
  ) {
    final isLoading = isToday ? _isLoadingToday : _isLoadingPrevious;
    final hasMore = isToday ? _hasMoreToday : _hasMorePrevious;

    if (isLoading && orders.isEmpty) {
      return const Center(
        child: CircularProgressIndicator(
          valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: () => isToday
          ? _loadTodayOrders(reset: true)
          : _loadPreviousOrders(reset: true),
      color: const Color(0xFF20CB6B),
      child: orders.isEmpty
          ? SingleChildScrollView(
              physics: const AlwaysScrollableScrollPhysics(),
              child: SizedBox(
                height: MediaQuery.of(context).size.height * 0.6,
                child: Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(
                        Icons.inbox_outlined,
                        size: 64,
                        color: Color(0xFF8C92A4),
                      ),
                      const SizedBox(height: 16),
                      Text(
                        isToday ? '今日暂无订单' : '暂无历史订单',
                        style: const TextStyle(
                          fontSize: 16,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            )
          : ListView.builder(
              controller: scrollController,
              physics: const AlwaysScrollableScrollPhysics(),
              padding: EdgeInsets.fromLTRB(
                16,
                16,
                16,
                16 + MediaQuery.of(context).padding.bottom,
              ),
              itemCount: orders.length + (hasMore ? 1 : 0),
              itemBuilder: (context, index) {
                if (index >= orders.length) {
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

                final order = orders[index];
                return _buildOrderCard(order);
              },
            ),
    );
  }

  Widget _buildOrderCard(Map<String, dynamic> order) {
    final orderNumber = order['order_number'] as String? ?? '';
    final status = order['status'] as String? ?? '';
    final itemCount = (order['item_count'] as int?) ?? 0;
    final addressData = order['address'] as Map<String, dynamic>?;
    final storeName = addressData?['name'] as String? ?? '门店名称未填写';
    final address = addressData?['address'] as String? ?? '';
    final createdAt = order['created_at'] as String? ?? '';
    final isUrgent = (order['is_urgent'] as bool?) ?? false;

    // 配送费计算结果
    final deliveryFeeCalc =
        order['delivery_fee_calculation'] as Map<String, dynamic>?;
    final riderPayableFee =
        (deliveryFeeCalc?['rider_payable_fee'] as num?)?.toDouble() ?? 0.0;

    // 格式化时间
    String formattedTime = '';
    if (createdAt.isNotEmpty) {
      try {
        final dateTime = DateTime.parse(createdAt);
        final month = dateTime.month.toString().padLeft(2, '0');
        final day = dateTime.day.toString().padLeft(2, '0');
        final hour = dateTime.hour.toString().padLeft(2, '0');
        final minute = dateTime.minute.toString().padLeft(2, '0');
        formattedTime = '$month-$day $hour:$minute';
      } catch (e) {
        formattedTime = createdAt;
      }
    }

    // 格式化状态
    String statusText = '';
    Color statusColor = const Color(0xFF8C92A4);
    switch (status) {
      case 'delivered':
      case 'shipped':
        statusText = '已送达';
        statusColor = const Color(0xFF20CB6B);
        break;
      case 'paid':
      case 'completed':
        statusText = '已收款';
        statusColor = const Color(0xFF20CB6B);
        break;
      default:
        statusText = status;
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
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
      child: InkWell(
        onTap: () => _viewOrderDetail(order),
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // 订单编号和时间
              Row(
                children: [
                  Expanded(
                    child: Text(
                      '订单号：$orderNumber',
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w500,
                        color: Color(0xFF40475C),
                      ),
                    ),
                  ),
                  if (formattedTime.isNotEmpty)
                    Text(
                      formattedTime,
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFF8C92A4),
                      ),
                    ),
                ],
              ),
              const SizedBox(height: 12),
              // 门店名称
              Text(
                storeName,
                style: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 8),
              // 地址
              if (address.isNotEmpty)
                Text(
                  address,
                  style: const TextStyle(
                    fontSize: 14,
                    color: Color(0xFF8C92A4),
                    height: 1.4,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              const SizedBox(height: 12),
              // 底部信息：商品数量、状态、配送费
              Row(
                children: [
                  // 商品数量
                  Row(
                    children: [
                      const Icon(
                        Icons.shopping_cart_outlined,
                        size: 14,
                        color: Color(0xFF8C92A4),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '共$itemCount件',
                        style: const TextStyle(
                          fontSize: 13,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ],
                  ),
                  const Spacer(),
                  // 状态
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      statusText,
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w500,
                        color: statusColor,
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  // 配送费
                  if (deliveryFeeCalc != null)
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(
                          Icons.local_shipping,
                          size: 16,
                          color: Color(0xFF20CB6B),
                        ),
                        const SizedBox(width: 4),
                        Text(
                          '¥${riderPayableFee.toStringAsFixed(2)}',
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w700,
                            color: Color(0xFF20CB6B),
                          ),
                        ),
                      ],
                    ),
                  // 加急标识
                  if (isUrgent) ...[
                    const SizedBox(width: 8),
                    const Text(
                      '加急',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: Color.fromRGBO(241, 196, 15, 1),
                      ),
                    ),
                  ],
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
