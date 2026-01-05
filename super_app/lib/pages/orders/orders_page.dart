import 'package:flutter/material.dart';
import 'package:super_app/api/orders_api.dart';
import 'package:super_app/models/order.dart';

class OrdersPage extends StatefulWidget {
  const OrdersPage({super.key});

  @override
  State<OrdersPage> createState() => _OrdersPageState();
}

class _OrdersPageState extends State<OrdersPage>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  int _currentPage = 1;
  final int _pageSize = 20;
  bool _isLoading = false;
  bool _hasMore = true;
  List<Order> _orders = [];
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
    _tabController.addListener(() {
      if (!_tabController.indexIsChanging) {
        _refreshOrders();
      }
    });
    _scrollController.addListener(_onScroll);
    _loadOrders();
  }

  @override
  void dispose() {
    _tabController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    // 触底加载（距离底部200px时触发）
    if (_scrollController.position.pixels >=
            _scrollController.position.maxScrollExtent - 200 &&
        !_isLoading &&
        _hasMore) {
      _loadMoreOrders();
    }
  }

  Future<void> _refreshOrders() async {
    setState(() {
      _currentPage = 1;
      _hasMore = true;
      _orders.clear();
    });
    await _loadOrders();
  }

  Future<void> _loadOrders() async {
    // 防止重复加载（但允许第一页加载）
    if (_isLoading) return;

    setState(() {
      _isLoading = true;
    });

    try {
      final tabIndex = _tabController.index;
      List<Order> newOrders = [];

      if (tabIndex == 1) {
        // 待审核：pending_delivery 和 pending_pickup
        // 由于API只支持单个status，需要分别请求两个状态并合并
        final response1 = await OrdersApi.getOrders(
          pageNum: _currentPage,
          pageSize: _pageSize,
          status: 'pending_delivery',
        );
        final response2 = await OrdersApi.getOrders(
          pageNum: _currentPage,
          pageSize: _pageSize,
          status: 'pending_pickup',
        );

        if (!mounted) return;

        if (response1.isSuccess && response1.data != null) {
          newOrders.addAll(response1.data!.list);
        }
        if (response2.isSuccess && response2.data != null) {
          newOrders.addAll(response2.data!.list);
        }

        // 去重并按创建时间倒序排序
        final orderMap = <int, Order>{};
        for (var order in newOrders) {
          if (!orderMap.containsKey(order.id)) {
            orderMap[order.id] = order;
          }
        }
        newOrders = orderMap.values.toList()
          ..sort((a, b) => b.createdAt.compareTo(a.createdAt));
      } else if (tabIndex == 2) {
        // 已通过：delivered 和 paid
        final response1 = await OrdersApi.getOrders(
          pageNum: _currentPage,
          pageSize: _pageSize,
          status: 'delivered',
        );
        final response2 = await OrdersApi.getOrders(
          pageNum: _currentPage,
          pageSize: _pageSize,
          status: 'paid',
        );

        if (!mounted) return;

        if (response1.isSuccess && response1.data != null) {
          newOrders.addAll(response1.data!.list);
        }
        if (response2.isSuccess && response2.data != null) {
          newOrders.addAll(response2.data!.list);
        }

        // 去重并按创建时间倒序排序
        final orderMap = <int, Order>{};
        for (var order in newOrders) {
          if (!orderMap.containsKey(order.id)) {
            orderMap[order.id] = order;
          }
        }
        newOrders = orderMap.values.toList()
          ..sort((a, b) => b.createdAt.compareTo(a.createdAt));
      } else {
        // 全部或已拒绝：使用单个status
        String? status;
        if (tabIndex == 3) {
          status = 'cancelled';
        }

        final response = await OrdersApi.getOrders(
          pageNum: _currentPage,
          pageSize: _pageSize,
          status: status,
        );

        if (!mounted) return;

        if (response.isSuccess && response.data != null) {
          newOrders = response.data!.list;
        } else {
          setState(() {
            _isLoading = false;
          });
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(response.message)),
            );
          }
          return;
        }
      }

      if (!mounted) return;

      setState(() {
        if (_currentPage == 1) {
          _orders.clear();
        }
        _orders.addAll(newOrders);
        // 对于复合状态的tab，判断是否有更多数据比较困难，简单判断是否返回了数据
        _hasMore = newOrders.length >= _pageSize;
        if (_hasMore) {
          _currentPage++;
        }
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('加载失败: ${e.toString()}')),
        );
      }
    }
  }

  Future<void> _loadMoreOrders() async {
    if (_hasMore && !_isLoading) {
      _currentPage++;
      await _loadOrders();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          // 标签栏
          Container(
            color: Colors.white,
            child: TabBar(
              controller: _tabController,
              labelColor: const Color(0xFF20CB6B),
              unselectedLabelColor: const Color(0xFF8C92A4),
              indicatorColor: const Color(0xFF20CB6B),
              tabs: const [
                Tab(text: '全部'),
                Tab(text: '待审核'),
                Tab(text: '已通过'),
                Tab(text: '已拒绝'),
              ],
            ),
          ),
          
          // 订单列表
          Expanded(
            child: RefreshIndicator(
              onRefresh: _refreshOrders,
              child: _buildOrderList(),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildOrderList() {
    if (_isLoading && _orders.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_orders.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.receipt_long_outlined,
              size: 64,
              color: Colors.grey[300],
            ),
            const SizedBox(height: 16),
            Text(
              '暂无订单',
              style: TextStyle(
                fontSize: 16,
                color: Colors.grey[600],
              ),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
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
    );
  }

  Widget _buildOrderCard(Order order) {
    final address = order.address;
    final addressName = address?['name'] as String? ?? '';
    final addressContact = address?['contact'] as String? ?? '';
    final addressPhone = address?['phone'] as String? ?? '';
    final addressText = address?['address'] as String? ?? '';
    final contact = addressContact.isNotEmpty ? addressContact : addressPhone;

    return InkWell(
      onTap: () {
        // TODO: 跳转到订单详情
      },
      borderRadius: BorderRadius.circular(16),
      child: Container(
        margin: const EdgeInsets.only(top: 12),
        padding: const EdgeInsets.all(14),
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
            // 地址名称和订单号
            Row(
              children: [
                Expanded(
                  child: Text(
                    addressName.isNotEmpty ? addressName : '地址名称未填写',
                    style: const TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF20253A),
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                if (order.isUrgent) ...[
                  const SizedBox(width: 6),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFF6B6B),
                      borderRadius: BorderRadius.circular(4),
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
                if (order.orderNumber.isNotEmpty) ...[
                  const SizedBox(width: 8),
                  Text(
                    order.orderNumber,
                    style: const TextStyle(
                      fontSize: 11,
                      color: Color(0xFF8C92A4),
                    ),
                  ),
                ],
              ],
            ),
            const SizedBox(height: 6),
            // 联系电话
            if (contact.isNotEmpty)
              Text(
                contact,
                style: const TextStyle(
                  fontSize: 13,
                  color: Color(0xFF40475C),
                ),
              ),
            // 地址
            if (addressText.isNotEmpty) ...[
              const SizedBox(height: 4),
              Text(
                addressText,
                style: const TextStyle(
                  fontSize: 12,
                  color: Color(0xFF8C92A4),
                ),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ],
            const SizedBox(height: 8),
            // 商品数量和创建时间
            Row(
              children: [
                Text(
                  '共${order.itemCount ?? 0}件商品',
                  style: const TextStyle(
                    fontSize: 12,
                    color: Color(0xFF8C92A4),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    _formatDateTime(order.createdAt),
                    style: const TextStyle(
                      fontSize: 11,
                      color: Color(0xFFB0B4C3),
                    ),
                    textAlign: TextAlign.right,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            // 状态和实付金额
            Row(
              children: [
                _buildOrderStatus(order.status),
                const Spacer(),
                const Text(
                  '实付金额：',
                  style: TextStyle(
                    fontSize: 12,
                    color: Color(0xFF8C92A4),
                  ),
                ),
                Text(
                  '¥${order.totalAmount.toStringAsFixed(2)}',
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF20CB6B),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildOrderStatus(String status) {
    Color color;
    String text;

    switch (status) {
      case 'pending_delivery':
        color = const Color(0xFFFFA940); // 橙色
        text = '待配送';
        break;
      case 'pending_pickup':
        color = const Color(0xFF4C8DF6); // 蓝色
        text = '待取货';
        break;
      case 'delivering':
        color = const Color(0xFF20CB6B); // 绿色
        text = '配送中';
        break;
      case 'delivered':
        color = const Color(0xFF20CB6B); // 绿色
        text = '已送达';
        break;
      case 'paid':
        color = const Color(0xFF7C4DFF); // 紫色
        text = '已收款';
        break;
      case 'cancelled':
        color = const Color(0xFFB0B4C3); // 灰色
        text = '已取消';
        break;
      default:
        color = const Color(0xFF8C92A4); // 默认灰色
        text = status;
    }

    return Text(
      text,
      style: TextStyle(
        fontSize: 13,
        color: color,
        fontWeight: FontWeight.w600,
      ),
    );
  }

  String _formatDateTime(DateTime dateTime) {
    // 格式：MM-DD HH:mm
    return '${dateTime.month.toString().padLeft(2, '0')}-${dateTime.day.toString().padLeft(2, '0')} '
        '${dateTime.hour.toString().padLeft(2, '0')}:${dateTime.minute.toString().padLeft(2, '0')}';
  }
}

