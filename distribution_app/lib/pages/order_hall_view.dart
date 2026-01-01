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

/// 用于刷新订单大厅的GlobalKey
typedef OrderHallViewKey = GlobalKey<_OrderHallViewState>;

extension OrderHallViewKeyExtension on OrderHallViewKey {
  /// 刷新订单列表和角标数量
  Future<void> refreshAll() async {
    await currentState?._refreshAll();
  }
  
  /// 设置页面可见性
  void setPageVisible(bool visible) {
    currentState?._setPageVisible(visible);
  }
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

  // 每个tab的订单数量（初始值为null，表示未加载）
  List<int?> _orderCounts = [null, null, null];
  // 是否正在加载数量
  bool _isLoadingCounts = false;
  // 防抖定时器：避免频繁调用 _loadAllTabCounts
  DateTime? _lastLoadTime;
  static const Duration _minLoadInterval = Duration(seconds: 5); // 最小间隔5秒，减少请求频率
  // 标记页面是否可见（用于判断是否应该刷新）
  bool _isPageVisible = true;
  
  /// 设置页面可见性（由外部调用，比如 main_shell）
  void _setPageVisible(bool visible) {
    if (_isPageVisible != visible) {
      setState(() {
        _isPageVisible = visible;
      });
      // 如果页面变为可见，且距离上次加载时间超过最小间隔，则刷新
      if (visible) {
        final now = DateTime.now();
        if (_lastLoadTime == null ||
            now.difference(_lastLoadTime!) >= _minLoadInterval) {
          _loadAllTabCounts();
        }
      }
    }
  }

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _loadEmployeeInfo();
    // 监听应用生命周期，当应用从后台返回时刷新列表
    WidgetsBinding.instance.addObserver(this);
    // 页面初始化时，立即加载所有 tab 的数量（不等待postFrameCallback）
    _loadAllTabCounts();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _tabController.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    // 当应用从后台返回前台时，且页面可见时，才刷新所有Tab的订单列表
    if (state == AppLifecycleState.resumed && _isPageVisible) {
      _refreshAll();
    }
  }

  /// 刷新所有订单列表和角标数量
  Future<void> _refreshAll() async {
    // 先刷新数量
    await _loadAllTabCounts();

    // 然后刷新所有 tab
    final refreshFutures = <Future>[];
    for (final key in _tabKeys) {
      if (key.currentState != null) {
        refreshFutures.add(key.refresh());
      }
    }
    await Future.wait(refreshFutures);

    // 刷新完成后，再次更新数量（确保数量准确）
    if (mounted) {
      await _loadAllTabCounts();
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
    // 接单成功后，强制刷新（忽略防抖限制和页面可见性检查）
    // 先刷新数量
    await _loadAllTabCounts(force: true);

    // 然后刷新所有Tab（因为订单状态可能变化）
    final refreshFutures = <Future>[];
    for (final key in _tabKeys) {
      if (key.currentState != null) {
        refreshFutures.add(key.refresh());
      }
    }
    await Future.wait(refreshFutures);

    // 刷新完成后，再次更新数量（确保数量准确）
    if (mounted) {
      await _loadAllTabCounts(force: true);
    }
  }

  /// 直接调用 API 获取所有 tab 的数量（不依赖 tab 构建）
  /// 这是角标数量的唯一数据源，确保不会闪烁
  Future<void> _loadAllTabCounts({bool force = false}) async {
    // 如果页面不可见，不加载（避免在详情页面时后台刷新）
    if (!_isPageVisible && !force) return;
    
    // 如果正在加载，避免重复请求
    if (_isLoadingCounts) return;

    // 防抖：如果距离上次加载时间太短，跳过本次请求（除非强制刷新）
    if (!force) {
      final now = DateTime.now();
      if (_lastLoadTime != null &&
          now.difference(_lastLoadTime!) < _minLoadInterval) {
        return;
      }
    }

    _lastLoadTime = DateTime.now();
    _isLoadingCounts = true;
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
        // 先计算新数量，只有在成功获取到数据时才更新，否则保持旧值
        final newCounts = List<int?>.from(_orderCounts); // 先复制旧值
        for (int i = 0; i < results.length; i++) {
          if (results[i].isSuccess && results[i].data != null) {
            final total = results[i].data!['total'] as int? ?? 0;
            newCounts[i] = total; // 只有成功时才更新
          }
          // 如果失败，保持旧值不变（newCounts[i] 已经是旧值）
        }
        // 一次性更新所有数量，避免中间状态
        setState(() {
          for (int i = 0; i < newCounts.length; i++) {
            _orderCounts[i] = newCounts[i];
          }
        });
      }
    } catch (e) {
      // 如果 API 调用失败，保持原值不变，不进行任何更新
      // 这样可以避免闪烁，因为数量不会突然变为0
    } finally {
      _isLoadingCounts = false;
    }
  }

  void _navigateToRoutePlanning() {
    Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (_) => const RoutePlanningView()));
  }

  /// 构建带徽标的Tab
  Widget _buildTabWithBadge(String text, int? count) {
    // 如果数量为null，表示还未加载，不显示角标
    final displayCount = count ?? 0;
    return Tab(
      child: Stack(
        clipBehavior: Clip.none,
        children: [
          // Tab文本
          Padding(
            padding: EdgeInsets.only(right: displayCount > 0 ? 12 : 0),
            child: Text(text),
          ),
          // 徽标（绝对定位）
          if (displayCount > 0)
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
                    displayCount > 99 ? '99+' : displayCount.toString(),
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
                  onOrderCountChanged: () async {
                    // 延迟一下，确保后端已处理状态变更
                    await Future.delayed(const Duration(milliseconds: 500));
                    if (mounted) {
                      // 强制刷新，确保角标能及时更新
                      await _loadAllTabCounts(force: true);
                    }
                  },
                ),
                // 待取货（已接单但未取货）
                OrderListTab(
                  key: _tabKeys[1],
                  status: 'pending_pickup',
                  onOrderAccepted: _onOrderAccepted,
                  onOrderCountChanged: () async {
                    // 延迟一下，确保后端已处理状态变更
                    await Future.delayed(const Duration(milliseconds: 500));
                    if (mounted) {
                      // 强制刷新，确保角标能及时更新
                      await _loadAllTabCounts(force: true);
                    }
                  },
                ),
                // 配送中（已取货正在配送）
                OrderListTab(
                  key: _tabKeys[2],
                  status: 'delivering',
                  onOrderAccepted: _onOrderAccepted,
                  onOrderCountChanged: () async {
                    // 延迟一下，确保后端已处理状态变更
                    await Future.delayed(const Duration(milliseconds: 500));
                    if (mounted) {
                      // 强制刷新，确保角标能及时更新
                      await _loadAllTabCounts(force: true);
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
