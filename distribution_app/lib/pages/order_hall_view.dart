import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'order_list_tab.dart';
import 'route_planning_view.dart';
import '../utils/location_service.dart';

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
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  final List<GlobalKey> _tabKeys = [GlobalKey(), GlobalKey(), GlobalKey()];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  void _onOrderAccepted() {
    // 接单成功后，刷新待取货Tab
    // 这里可以通过Key来刷新对应的Tab
  }

  void _navigateToRoutePlanning() {
    Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (_) => const RoutePlanningView()));
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
            color: const Color(0xFF20CB6B),
            child: SafeArea(
              bottom: false,
              child: Column(
                children: [
                  Row(
                    children: [
                      const Text(
                        '接单大厅',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 20,
                          fontWeight: FontWeight.w600,
                        ),
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
                  // 定位信息显示
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: Colors.white.withOpacity(0.2),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Row(
                      children: [
                        if (widget.isLoadingLocation)
                          const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                            ),
                          )
                        else
                          Icon(
                            widget.currentPosition != null
                                ? Icons.location_on
                                : Icons.location_off,
                            color: Colors.white,
                            size: 16,
                          ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            widget.isLoadingLocation
                                ? '正在获取定位...'
                                : widget.locationError ?? 
                                  (widget.currentPosition != null
                                      ? LocationService.formatLocation(widget.currentPosition)
                                      : '定位未获取'),
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 12,
                            ),
                          ),
                        ),
                        if (!widget.isLoadingLocation)
                          Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              if (widget.locationError != null)
                                TextButton(
                                  onPressed: () async {
                                    // 根据错误类型打开不同的设置页面
                                    if (widget.locationError!.contains('定位服务未启用') ||
                                        widget.locationError!.contains('GPS')) {
                                      await LocationService.openLocationSettings();
                                    } else {
                                      await LocationService.openAppSettingsPage();
                                    }
                                    // 延迟一下再刷新，给用户时间操作
                                    Future.delayed(const Duration(seconds: 1), () {
                                      if (widget.onRefreshLocation != null) {
                                        widget.onRefreshLocation!();
                                      }
                                    });
                                  },
                                  style: TextButton.styleFrom(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 4,
                                    ),
                                    minimumSize: Size.zero,
                                    tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                                  ),
                                  child: Text(
                                    widget.locationError!.contains('定位服务未启用') ||
                                            widget.locationError!.contains('GPS')
                                        ? '开启GPS'
                                        : '去设置',
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 11,
                                      decoration: TextDecoration.underline,
                                    ),
                                  ),
                                ),
                              if (widget.onRefreshLocation != null)
                                IconButton(
                                  icon: const Icon(
                                    Icons.refresh,
                                    color: Colors.white,
                                    size: 18,
                                  ),
                                  onPressed: widget.onRefreshLocation,
                                  padding: EdgeInsets.zero,
                                  constraints: const BoxConstraints(),
                                ),
                            ],
                          ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          // Tab栏
          Container(
            color: Colors.white,
            child: TabBar(
              controller: _tabController,
              labelColor: const Color(0xFF20CB6B),
              unselectedLabelColor: const Color(0xFF8C92A4),
              indicatorColor: const Color(0xFF20CB6B),
              indicatorWeight: 3,
              labelStyle: const TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w600,
              ),
              unselectedLabelStyle: const TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.normal,
              ),
              tabs: const [
                Tab(text: '新任务'),
                Tab(text: '待取货'),
                Tab(text: '配送中'),
              ],
            ),
          ),
          // Tab内容（可滑动切换）
          Expanded(
            child: TabBarView(
              controller: _tabController,
              children: [
                // 新任务（待接单订单）
                OrderListTab(
                  key: _tabKeys[0],
                  status: null, // null表示新任务（待接单）
                  onOrderAccepted: _onOrderAccepted,
                ),
                // 待取货（已接单但未取货）
                OrderListTab(
                  key: _tabKeys[1],
                  status: 'pending_pickup',
                  onOrderAccepted: _onOrderAccepted,
                ),
                // 配送中（已取货正在配送）
                OrderListTab(
                  key: _tabKeys[2],
                  status: 'delivering',
                  onOrderAccepted: _onOrderAccepted,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
