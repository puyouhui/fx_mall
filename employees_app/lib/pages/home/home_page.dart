import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:employees_app/api/auth_api.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/customer/customer_profile_page.dart';
import 'package:employees_app/pages/customer/customer_list_page.dart';
import 'package:employees_app/pages/order/sales_create_order_page.dart';
import 'package:employees_app/pages/coupon/coupon_send_page.dart';
import 'package:employees_app/pages/product/product_search_page.dart';
import 'package:employees_app/pages/order/order_list_page.dart';
import 'package:employees_app/pages/order/order_detail_page.dart';
import 'package:employees_app/pages/order/edit_order_list_page.dart';

/// 员工端首页（总览 + 配送）
class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  bool _isLoading = true;
  String? _errorMessage;
  Map<String, dynamic>? _dashboard;

  @override
  void initState() {
    super.initState();
    _loadDashboard();
  }

  @override
  void dispose() {
    super.dispose();
  }

  Future<void> _loadDashboard() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final response = await AuthApi.getDashboard();
    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      setState(() {
        _dashboard = response.data;
        _isLoading = false;
      });
    } else {
      setState(() {
        _errorMessage = response.message;
        _isLoading = false;
      });
    }
  }

  String _getGreeting() {
    final hour = DateTime.now().hour;
    if (hour < 11) return '上午好';
    if (hour < 13) return '中午好';
    if (hour < 18) return '下午好';
    return '晚上好';
  }

  @override
  Widget build(BuildContext context) {
    final name = (_dashboard?['name'] as String?) ?? '员工';
    final employeeCode = (_dashboard?['employee_code'] as String?) ?? '';

    return Scaffold(
      extendBody: true, // 让body延伸到系统操作条下方
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
          child: _isLoading
              ? const Center(
                  child: CircularProgressIndicator(
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                )
              : _errorMessage != null
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(
                        Icons.error_outline,
                        color: Colors.white,
                        size: 40,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        _errorMessage!,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 14,
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 12),
                      ElevatedButton(
                        onPressed: _loadDashboard,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.white,
                          foregroundColor: const Color(0xFF20CB6B),
                        ),
                        child: const Text('重试'),
                      ),
                    ],
                  ),
                )
              : Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // 顶部问候
                    Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 16,
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            '$name，${_getGreeting()}',
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 22,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            employeeCode.isNotEmpty ? '工号：$employeeCode' : '',
                            style: const TextStyle(
                              color: Colors.white70,
                              fontSize: 13,
                            ),
                          ),
                        ],
                      ),
                    ),

                    // 内容区域
                    Expanded(
                      child: OverviewTab(
                        dashboard: _dashboard,
                        onRefreshDashboard: _loadDashboard,
                      ),
                    ),
                  ],
                ),
        ),
      ),
    );
  }
}

class StatTile extends StatelessWidget {
  final String label;
  final String value;
  final Color accentColor;

  const StatTile({
    super.key,
    required this.label,
    required this.value,
    required this.accentColor,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFF7F8FA),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: const TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
          ),
          const SizedBox(height: 4),
          Row(
            children: [
              Text(
                value,
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: accentColor,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

/// 总览 Tab
class OverviewTab extends StatefulWidget {
  final Map<String, dynamic>? dashboard;
  final Future<void> Function()? onRefreshDashboard;

  const OverviewTab({
    super.key,
    required this.dashboard,
    this.onRefreshDashboard,
  });

  @override
  State<OverviewTab> createState() => _OverviewTabState();
}

class _OverviewTabState extends State<OverviewTab> {
  final List<Map<String, dynamic>> _pendingOrders = [];
  int _pageNum = 1;
  final int _pageSize = 10;
  bool _hasMore = true;
  bool _isLoadingMore = false;
  bool _initialized = false;

  @override
  void initState() {
    super.initState();
    _loadMorePendingOrders();
  }

  /// 刷新待配送订单列表（下拉刷新时调用）
  Future<void> _refreshPendingOrders() async {
    try {
      // 同时刷新 dashboard 数据
      if (widget.onRefreshDashboard != null) {
        await widget.onRefreshDashboard!();
      }

      // 刷新待配送订单列表
      setState(() {
        _pendingOrders.clear();
        _pageNum = 1;
        _hasMore = true;
        _isLoadingMore = false;
      });
      await _loadMorePendingOrders();
    } catch (e) {
      // 如果刷新失败，确保刷新指示器能够正确关闭
      if (mounted) {
        setState(() {
          _isLoadingMore = false;
        });
      }
      rethrow;
    }
  }

  Future<void> _loadMorePendingOrders() async {
    if (_isLoadingMore || !_hasMore) return;

    setState(() {
      _isLoadingMore = true;
    });

    final response = await Request.get<Map<String, dynamic>>(
      '/employee/sales/pending-orders',
      queryParams: {'pageNum': '$_pageNum', 'pageSize': '$_pageSize'},
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final list = (data['list'] as List<dynamic>? ?? [])
          .cast<Map<String, dynamic>>();
      final total = data['total'] as int? ?? 0;

      setState(() {
        _pendingOrders.addAll(list);
        _pageNum++;
        _hasMore = _pendingOrders.length < total;
        _isLoadingMore = false;
        _initialized = true;
      });
    } else {
      setState(() {
        _isLoadingMore = false;
        _initialized = true;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final dashboard = widget.dashboard;
    final isSales = dashboard?['is_sales'] == true;
    final isDelivery = dashboard?['is_delivery'] == true;

    final customerCount = (dashboard?['customer_count'] as int?) ?? 0;
    final orderTotal = (dashboard?['order_total'] as int?) ?? 0;
    final orderPendingDelivery =
        (dashboard?['order_pending_delivery'] as int?) ?? 0;
    final orderToday = (dashboard?['order_today'] as int?) ?? 0;

    return NotificationListener<ScrollNotification>(
      onNotification: (notification) {
        if (notification.metrics.pixels >=
                notification.metrics.maxScrollExtent - 80 &&
            !_isLoadingMore &&
            _hasMore) {
          _loadMorePendingOrders();
        }
        return false;
      },
      child: RefreshIndicator(
        onRefresh: _refreshPendingOrders,
        color: const Color(0xFF20CB6B),
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(), // 确保即使内容不够高也能下拉刷新
          padding: EdgeInsets.fromLTRB(
            16,
            12,
            16,
            16 + MediaQuery.of(context).padding.bottom, // 添加底部安全区域内边距
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // 统计区
              Row(
                children: [
                  Expanded(
                    child: StatTile(
                      label: '我的客户',
                      value: customerCount.toString(),
                      accentColor: const Color(0xFF20CB6B),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: StatTile(
                      label: '我的订单总数',
                      value: orderTotal.toString(),
                      accentColor: const Color(0xFF4C8DF6),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: StatTile(
                      label: '待配送订单',
                      value: orderPendingDelivery.toString(),
                      accentColor: const Color(0xFFFFA940),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: StatTile(
                      label: '今日新增订单',
                      value: orderToday.toString(),
                      accentColor: const Color(0xFFFF5A5F),
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 20),

              // 常用功能
              Container(
                width: double.infinity,
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 14,
                ),
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
                      '常用功能',
                      style: TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF20253A),
                      ),
                    ),
                    const SizedBox(height: 10),
                    Column(
                      children: [
                        // 第一行：4个按钮（1. 新客资料 2. 产品查询 3. 销售开单 4. 修改订单）
                        Row(
                          children: [
                            // 1. 新客资料
                            QuickActionItem(
                              icon: Icons.person_add_alt_1_outlined,
                              iconColor: const Color(0xFFFFA940),
                              label: '新客资料',
                              onTap: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (_) => const CustomerProfilePage(),
                                  ),
                                );
                              },
                            ),
                            const SizedBox(width: 8),
                            // 2. 产品查询
                            QuickActionItem(
                              icon: Icons.search,
                              iconColor: const Color(0xFF20CB6B),
                              label: '产品查询',
                              onTap: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (_) => const ProductSearchPage(),
                                  ),
                                );
                              },
                            ),
                            const SizedBox(width: 8),
                            // 3. 销售开单
                            QuickActionItem(
                              icon: Icons.receipt_long_outlined,
                              iconColor: const Color(0xFF4C8DF6),
                              label: '销售开单',
                              onTap: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (_) =>
                                        const SalesCreateOrderPage(),
                                  ),
                                );
                              },
                            ),
                            const SizedBox(width: 8),
                            // 4. 修改订单
                            QuickActionItem(
                              icon: Icons.edit_note_outlined,
                              iconColor: const Color(0xFF7C4DFF),
                              label: '修改订单',
                              onTap: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (_) => const EditOrderListPage(),
                                  ),
                                );
                              },
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        // 第二行：我的客户 + 送优惠券 + 订单查询 + 收益查询
                        Row(
                          children: [
                            QuickActionItem(
                              icon: Icons.people_alt_outlined,
                              iconColor: const Color(0xFF20CB6B),
                              label: '我的客户',
                              onTap: () {
                                ScaffoldMessenger.of(context).showSnackBar(
                                  const SnackBar(
                                    content: Text('正在进入我的客户列表...'),
                                    duration: Duration(milliseconds: 800),
                                  ),
                                );
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (_) => const CustomerListPage(),
                                  ),
                                );
                              },
                            ),
                            const SizedBox(width: 8),
                            QuickActionItem(
                              icon: Icons.card_giftcard,
                              iconColor: const Color(0xFFFF5A5F),
                              label: '送优惠券',
                              onTap: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (_) => const CouponSendPage(),
                                  ),
                                );
                              },
                            ),
                            const SizedBox(width: 8),
                            QuickActionItem(
                              icon: Icons.receipt_long,
                              iconColor: const Color(0xFF4C8DF6),
                              label: '订单查询',
                              onTap: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (_) => const OrderListPage(),
                                  ),
                                );
                              },
                            ),
                            const SizedBox(width: 8),
                            const QuickActionItem(
                              icon: Icons.trending_up,
                              iconColor: Color(0xFFFFA940),
                              label: '分成查询',
                              // TODO: 接收益查询页面
                            ),
                          ],
                        ),
                      ],
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 20),

              // 待配送订单列表
              if (isSales) ...[
                const Text(
                  '待配送订单',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: Colors.white,
                  ),
                ),
                const SizedBox(height: 8),
                if (_pendingOrders.isEmpty)
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    decoration: BoxDecoration(
                      color: Colors.white.withOpacity(0.9),
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Center(
                      child: Text(
                        _initialized ? '暂无待配送订单' : '加载中...',
                        style: const TextStyle(
                          fontSize: 13,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ),
                  )
                else
                  Column(
                    children: [
                      ..._pendingOrders.map(
                        (order) => OrderPreviewRow(order: order),
                      ),
                      if (_isLoadingMore)
                        const Padding(
                          padding: EdgeInsets.symmetric(vertical: 8),
                          child: Center(
                            child: SizedBox(
                              width: 18,
                              height: 18,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            ),
                          ),
                        ),
                      if (!_hasMore && _pendingOrders.isNotEmpty)
                        const Padding(
                          padding: EdgeInsets.only(top: 8),
                          child: Center(
                            child: Text(
                              '没有更多了',
                              style: TextStyle(
                                fontSize: 11,
                                color: Color(0xFF8C92A4),
                              ),
                            ),
                          ),
                        ),
                    ],
                  ),
              ],

              if (!isSales && !isDelivery) ...[
                const SizedBox(height: 8),
                const Text(
                  '当前账号未配置销售员或配送员角色，统计数据有限。',
                  style: TextStyle(fontSize: 12, color: Colors.white70),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

/// 待配送订单列表中的一行
class OrderPreviewRow extends StatelessWidget {
  final Map<String, dynamic> order;

  const OrderPreviewRow({super.key, required this.order});

  @override
  Widget build(BuildContext context) {
    final storeName = order['store_name'] as String? ?? '未知门店';
    final totalAmountDynamic = order['total_amount'];
    final itemCount = order['item_count'] as int? ?? 0;
    final address = order['address'] as String? ?? '';
    final createdAtRaw = order['created_at']?.toString() ?? '';
    final isUrgent = (order['is_urgent'] as bool?) ?? false;

    String createdTimeText = createdAtRaw;
    if (createdAtRaw.isNotEmpty) {
      // 简单从时间字符串中截取「MM-DD HH:mm」部分，避免引入额外依赖
      // 例如 "2025-12-01T16:45:03Z" -> "12-01 16:45"
      try {
        if (createdAtRaw.length >= 16) {
          createdTimeText =
              '${createdAtRaw.substring(5, 10)} ${createdAtRaw.substring(11, 16)}';
        }
      } catch (_) {
        createdTimeText = createdAtRaw;
      }
    }

    String totalAmountText;
    if (totalAmountDynamic is num) {
      totalAmountText = totalAmountDynamic.toStringAsFixed(2);
    } else {
      totalAmountText = totalAmountDynamic?.toString() ?? '0.00';
    }

    final orderId = order['id'] as int?;

    return InkWell(
      onTap: orderId != null
          ? () {
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => OrderDetailPage(orderId: orderId),
                ),
              );
            }
          : null,
      borderRadius: BorderRadius.circular(18),
      child: Container(
        margin: const EdgeInsets.only(bottom: 14),
        padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 14),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(18),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.03),
              blurRadius: 10,
              offset: const Offset(0, 6),
            ),
          ],
        ),
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: const Color(0xFFFFA940).withOpacity(0.1),
                borderRadius: BorderRadius.circular(20),
              ),
              child: const Icon(
                Icons.local_shipping_outlined,
                color: Color(0xFFFFA940),
                size: 22,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // 门店名称
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          storeName,
                          style: const TextStyle(
                            fontSize: 17,
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
                    ],
                  ),
                  const SizedBox(height: 6),
                  // 地址 + 下单时间（图标左侧对齐）
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (address.isNotEmpty)
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Icon(
                              Icons.location_on_outlined,
                              size: 14,
                              color: Color(0xFF8C92A4),
                            ),
                            const SizedBox(width: 4),
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
                      const SizedBox(height: 4),
                      Row(
                        children: [
                          const Icon(
                            Icons.access_time,
                            size: 12,
                            color: Color(0xFF8C92A4),
                          ),
                          const SizedBox(width: 4),
                          Text(
                            '下单时间：$createdTimeText',
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  // 底部：左侧商品数量，右侧总价（右下角）
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        '共 $itemCount 件商品',
                        style: const TextStyle(
                          fontSize: 12,
                          color: Color(0xFF5C6478),
                        ),
                      ),
                      Text(
                        '¥$totalAmountText',
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: Color(0xFFFF5A5F),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// 常用功能入口按钮
class QuickActionItem extends StatelessWidget {
  final IconData icon;
  final Color iconColor;
  final String label;
  final VoidCallback? onTap;

  const QuickActionItem({
    super.key,
    required this.icon,
    required this.iconColor,
    required this.label,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 8),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 56,
                height: 56,
                decoration: BoxDecoration(
                  color: iconColor.withOpacity(0.08),
                  borderRadius: BorderRadius.circular(28),
                ),
                child: Icon(icon, color: iconColor, size: 28),
              ),
              const SizedBox(height: 8),
              Text(
                label,
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                  color: Color(0xFF40475C),
                ),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
