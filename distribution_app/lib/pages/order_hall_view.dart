import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'order_list_tab.dart';
import 'route_planning_view.dart';
import '../utils/storage.dart';
import '../api/order_api.dart';

/// 接单大厅页：包含三个Tab（新任务、待取货、配送中）
class OrderHallView extends StatefulWidget {
  const OrderHallView({
    super.key,
    this.currentPosition,
    this.isLoadingLocation = false,
    this.locationError,
    this.onRefreshLocation,
  });

  final Position? currentPosition;
  final bool isLoadingLocation;
  final String? locationError;
  final VoidCallback? onRefreshLocation;

  @override
  State<OrderHallView> createState() => _OrderHallViewState();
}

class _OrderHallViewState extends State<OrderHallView>
    with SingleTickerProviderStateMixin, WidgetsBindingObserver {
  late TabController _tabController;
  final List<OrderListTabKey> _tabKeys = [
    OrderListTabKey(),
    OrderListTabKey(),
    OrderListTabKey(),
  ];
  String _employeeName = '配送员';

  // 每个tab的订单数量
  final List<int> _orderCounts = [0, 0, 0];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    // 监听 tab 切换，确保切换到的 tab 数量已更新
    _tabController.addListener(_onTabChanged);
    _loadEmployeeInfo();
    // 监听应用生命周期，当应用从后台返回时刷新列表
    WidgetsBinding.instance.addObserver(this);
    // 页面初始化时，直接调用 API 获取所有 tab 的数量
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _loadAllTabCounts();
    });
  }

  /// Tab 切换时的回调
  void _onTabChanged() {
    if (!_tabController.indexIsChanging) {
      // Tab 切换完成，如果当前 tab 已加载完成，则更新数量
      // 否则保持原有数量（通过 API 获取的数量）
      _updateOrderCountsSafely();
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _tabController.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    // 当应用从后台返回前台时，刷新所有Tab的订单列表
    if (state == AppLifecycleState.resumed) {
      // 直接调用 API 更新数量
      _loadAllTabCounts();

      // 同时刷新所有 tab（不等待）
      for (final key in _tabKeys) {
        if (key.currentState != null) {
          key.refresh(); // 不等待，异步刷新
        }
      }
    }
  }

  Future<void> _loadEmployeeInfo() async {
    final employeeInfo = await Storage.getEmployeeInfo();
    if (employeeInfo != null && mounted) {
      setState(() {
        _employeeName = employeeInfo['name'] as String? ?? '配送员';
      });
    }
  }

  String _getGreeting() {
    final hour = DateTime.now().hour;
    if (hour >= 5 && hour < 12) {
      return '早上好';
    } else if (hour >= 12 && hour < 18) {
      return '下午好';
    } else if (hour >= 18 && hour < 22) {
      return '晚上好';
    } else {
      return '晚上好';
    }
  }

  void _onOrderAccepted() async {
    // 接单成功后，直接调用 API 更新所有 tab 的数量
    await _loadAllTabCounts();

    // 同时刷新所有Tab（因为订单状态可能变化）
    // 不等待刷新完成，让它们在后台加载
    for (final key in _tabKeys) {
      if (key.currentState != null) {
        key.refresh(); // 不等待，异步刷新
      }
    }
  }

  /// 直接调用 API 获取所有 tab 的数量（不依赖 tab 构建）
  Future<void> _loadAllTabCounts() async {
    try {
      // 并行获取三个状态的订单数量
      final futures = [
        OrderApi.getOrderPool(pageNum: 1, pageSize: 1, status: null), // 新任务
        OrderApi.getOrderPool(
          pageNum: 1,
          pageSize: 1,
          status: 'pending_pickup',
        ), // 待取货
        OrderApi.getOrderPool(
          pageNum: 1,
          pageSize: 1,
          status: 'delivering',
        ), // 配送中
      ];

      final results = await Future.wait(futures);

      if (mounted) {
        setState(() {
          // 更新每个 tab 的数量
          for (int i = 0; i < results.length; i++) {
            if (results[i].isSuccess && results[i].data != null) {
              final total = results[i].data!['total'] as int? ?? 0;
              _orderCounts[i] = total;
            }
          }
        });
      }
    } catch (e) {
      // 如果 API 调用失败，回退到从 tab 获取数量
      if (mounted) {
        _updateOrderCounts();
      }
    }
  }

  /// 更新所有tab的订单数量（从已构建的 tab 获取）
  /// 只有当 tab 已加载完成时才更新，避免覆盖通过 API 获取的正确数量
  void _updateOrderCountsSafely() {
    if (mounted) {
      setState(() {
        for (int i = 0; i < _tabKeys.length; i++) {
          final count = _tabKeys[i].getOrderCount();
          // 只有当新数量 >= 当前数量时才更新，避免用 0 覆盖通过 API 获取的正确数量
          // 如果新数量 < 当前数量，可能是 tab 还没加载完成，保持原有数量
          if (count >= _orderCounts[i]) {
            _orderCounts[i] = count;
          }
        }
      });
    }
  }

  /// 更新所有tab的订单数量（从已构建的 tab 获取）
  /// 用于 API 调用失败时的回退方案
  void _updateOrderCounts() {
    if (mounted) {
      setState(() {
        for (int i = 0; i < _tabKeys.length; i++) {
          _orderCounts[i] = _tabKeys[i].getOrderCount();
        }
      });
    }
  }

  void _navigateToRoutePlanning() {
    Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (_) => const RoutePlanningView()));
  }

  /// 构建带徽标的Tab
  Widget _buildTabWithBadge(String text, int count) {
    return Tab(
      child: Stack(
        clipBehavior: Clip.none,
        children: [
          // Tab文本
          Padding(
            padding: EdgeInsets.only(right: count > 0 ? 12 : 0),
            child: Text(text),
          ),
          // 徽标（绝对定位）
          if (count > 0)
            Positioned(
              right: -8,
              top: -6,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(color: const Color(0xFF20CB6B), width: 1),
                ),
                constraints: const BoxConstraints(minWidth: 18, minHeight: 18),
                child: Center(
                  child: Text(
                    count > 99 ? '99+' : count.toString(),
                    style: const TextStyle(
                      color: Color(0xFF20CB6B),
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                      height: 1.2,
                    ),
                    textAlign: TextAlign.center,
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [Color(0xFF20CB6B), Color(0xFFEFF7F2)],
        ),
      ),
      child: Column(
        children: [
          // 自定义头部：标题 + 路线规划按钮
          Container(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
            color: Colors.transparent,
            child: SafeArea(
              bottom: false,
              child: Column(
                children: [
                  Row(
                    children: [
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            '${_getGreeting()}，$_employeeName',
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 20,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                      const Spacer(),
                      ElevatedButton(
                        onPressed: _navigateToRoutePlanning,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.white,
                          foregroundColor: const Color(0xFF20CB6B),
                          padding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 10,
                          ),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(8),
                          ),
                          elevation: 0,
                        ),
                        child: const Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(Icons.route, size: 18),
                            SizedBox(width: 4),
                            Text(
                              '路线规划',
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          // Tab栏
          Container(
            color: Colors.transparent,
            child: TabBar(
              controller: _tabController,
              labelColor: Colors.white,
              unselectedLabelColor: Colors.white.withOpacity(0.7),
              indicatorColor: Colors.white,
              indicatorWeight: 3,
              dividerColor: Colors.transparent,
              labelStyle: const TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w600,
              ),
              unselectedLabelStyle: const TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.normal,
              ),
              tabs: [
                _buildTabWithBadge('新任务', _orderCounts[0]),
                _buildTabWithBadge('待取货', _orderCounts[1]),
                _buildTabWithBadge('配送中', _orderCounts[2]),
              ],
            ),
          ),
          // Tab内容（可滑动切换）
          Expanded(
            child: TabBarView(
              controller: _tabController,
              // 预加载所有 tab，确保数量能及时更新
              physics: const AlwaysScrollableScrollPhysics(),
              children: [
                // 新任务（待接单订单）
                OrderListTab(
                  key: _tabKeys[0],
                  status: null, // null表示新任务（待接单）
                  onOrderAccepted: _onOrderAccepted,
                  onOrderCountChanged: () {
                    if (mounted) {
                      final count = _tabKeys[0].getOrderCount();
                      // 只有当新数量 >= 当前数量时才更新，避免用 0 覆盖通过 API 获取的正确数量
                      if (count >= _orderCounts[0]) {
                        setState(() {
                          _orderCounts[0] = count;
                        });
                      }
                    }
                  },
                ),
                // 待取货（已接单但未取货）
                OrderListTab(
                  key: _tabKeys[1],
                  status: 'pending_pickup',
                  onOrderAccepted: _onOrderAccepted,
                  onOrderCountChanged: () {
                    if (mounted) {
                      final count = _tabKeys[1].getOrderCount();
                      // 只有当新数量 >= 当前数量时才更新，避免用 0 覆盖通过 API 获取的正确数量
                      if (count >= _orderCounts[1]) {
                        setState(() {
                          _orderCounts[1] = count;
                        });
                      }
                    }
                  },
                ),
                // 配送中（已取货正在配送）
                OrderListTab(
                  key: _tabKeys[2],
                  status: 'delivering',
                  onOrderAccepted: _onOrderAccepted,
                  onOrderCountChanged: () {
                    if (mounted) {
                      final count = _tabKeys[2].getOrderCount();
                      // 只有当新数量 >= 当前数量时才更新，避免用 0 覆盖通过 API 获取的正确数量
                      if (count >= _orderCounts[2]) {
                        setState(() {
                          _orderCounts[2] = count;
                        });
                      }
                    }
                  },
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
