import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import '../api/order_api.dart';
import '../utils/location_service.dart';
import '../widgets/accept_order_dialog.dart';
import 'order_detail_view.dart';

/// 订单列表Tab组件：用于显示不同状态的订单列表
class OrderListTab extends StatefulWidget {
  const OrderListTab({
    super.key,
    required this.status,
    required this.onOrderAccepted,
    this.onOrderCountChanged,
  });

  final String? status; // 订单状态：null=新任务, 'pending_pickup'=待取货, 'delivering'=配送中
  final VoidCallback onOrderAccepted; // 接单成功回调
  final Future<void> Function()? onOrderCountChanged; // 订单数量变化回调

  @override
  State<OrderListTab> createState() => _OrderListTabState();
}

/// 用于刷新订单列表的GlobalKey
typedef OrderListTabKey = GlobalKey<_OrderListTabState>;

extension OrderListTabKeyExtension on OrderListTabKey {
  Future<void> refresh() async {
    await currentState?.refresh();
  }

  int getOrderCount() {
    // 使用 API 返回的 total，而不是当前已加载的订单数
    return currentState?._total ?? 0;
  }

  /// 等待刷新完成
  Future<void> waitForRefresh() async {
    await currentState?._waitForRefresh();
  }
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
  int _total = 0; // 订单总数（从 API 返回）

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

  /// 刷新订单列表（供外部调用）
  Future<void> refresh() async {
    await _loadOrders(reset: true);
  }

  /// 等待刷新完成
  Future<void> _waitForRefresh() async {
    if (_isLoading) {
      // 如果正在加载，等待加载完成
      while (_isLoading && mounted) {
        await Future.delayed(const Duration(milliseconds: 100));
      }
    }
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
      final total = data['total'] as int? ?? 0;

      setState(() {
        if (reset) {
          _orders
            ..clear()
            ..addAll(orders);
        } else {
          _orders.addAll(orders);
        }
        _total = total; // 更新总数
        _hasMore = _orders.length < total;
        if (_hasMore) {
          _pageNum++;
        }
        _isLoading = false;
        _isLoadingMore = false;
      });
      // 只在重置加载（首次加载或手动刷新）时才通知订单数量变化
      // 避免滚动加载更多时频繁触发回调
      if (reset && widget.onOrderCountChanged != null) {
        // 使用微任务确保 setState 完成后再触发回调
        Future.microtask(() {
          if (mounted && widget.onOrderCountChanged != null) {
            widget.onOrderCountChanged!();
          }
        });
      }
    } else {
      // API调用失败时，不触发数量变化回调，保持当前数量不变，避免闪烁
      if (mounted) {
        setState(() {
          _isLoading = false;
          _isLoadingMore = false;
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

  Future<void> _acceptOrder(int orderId) async {
    if (_acceptingOrders[orderId] == true) return;

    // 找到订单数据
    Map<String, dynamic>? orderData;
    for (var order in _orders) {
      final id = (order['id'] as num?)?.toInt();
      if (id == orderId) {
        orderData = order;
        break;
      }
    }

    if (orderData == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('订单信息不存在'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    // 提取订单信息
    final addressData = orderData['address'] as Map<String, dynamic>?;
    final storeName = addressData?['name'] as String? ?? '门店名称未填写';
    final address = addressData?['address'] as String? ?? '';
    final itemCount = (orderData['item_count'] as int?) ?? 0;
    final totalAmount = (orderData['total_amount'] as num?)?.toDouble() ?? 0.0;
    final isUrgent = (orderData['is_urgent'] as bool?) ?? false;
    final deliveryFeeCalc =
        orderData['delivery_fee_calculation'] as Map<String, dynamic>?;
    final riderPayableFee =
        (deliveryFeeCalc?['rider_payable_fee'] as num?)?.toDouble() ?? 0.0;

    // 显示确认对话框
    final confirmed = await AcceptOrderDialog.show(
      context,
      storeName: storeName,
      address: address,
      riderPayableFee: riderPayableFee,
      totalAmount: totalAmount,
      itemCount: itemCount,
      isUrgent: isUrgent,
    );

    if (confirmed != true) {
      return; // 用户取消接单
    }

    // 接单前检查位置权限和获取位置
    try {
      // 检查定位权限
      final hasPermission = await LocationService.checkAndRequestPermission();
      if (!hasPermission) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('无法获取位置信息，无法接单。请前往设置开启定位权限'),
            backgroundColor: Colors.red,
            duration: Duration(seconds: 3),
          ),
        );
        return;
      }

      // 获取当前位置（使用缓存的位置，如果可用）
      Position? position = LocationService.getCachedPosition();
      if (position == null) {
        position = await LocationService.getCurrentLocation();
      }
      
      if (position == null) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('无法获取当前位置，请确保已开启GPS定位服务'),
            backgroundColor: Colors.red,
            duration: Duration(seconds: 3),
          ),
        );
        return;
      }

      setState(() {
        _acceptingOrders[orderId] = true;
      });

      // 传递位置信息接单
      final response = await OrderApi.acceptOrder(
        orderId,
        latitude: position.latitude,
        longitude: position.longitude,
      );

      if (!mounted) return;

      if (response.isSuccess) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('接单成功'),
            backgroundColor: Color(0xFF20CB6B),
          ),
        );
        // 先刷新当前列表（因为接单后订单状态会变化，需要重新加载）
        await _loadOrders(reset: true);
        // 然后通知父组件更新数量（接单后订单会从新任务移到待取货）
        // 延迟一下，确保列表刷新完成后再更新角标
        await Future.delayed(const Duration(milliseconds: 300));
        widget.onOrderAccepted();
      } else {
        // 检查是否是订单已被接走等错误，如果是则自动返回首页
        final errorMessage = response.message.toLowerCase();
        final shouldReturnHome = errorMessage.contains('已被') ||
            errorMessage.contains('接走') ||
            errorMessage.contains('已被接') ||
            errorMessage.contains('其他配送员') ||
            errorMessage.contains('已被其他');
        
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(response.message),
            backgroundColor: Colors.red,
            duration: const Duration(seconds: 2),
          ),
        );
        
        setState(() {
          _acceptingOrders.remove(orderId);
        });
        
        // 如果是订单已被接走等错误，自动返回首页
        if (shouldReturnHome && mounted) {
          await Future.delayed(const Duration(milliseconds: 500));
          if (mounted) {
            Navigator.of(context).popUntil((route) => route.isFirst);
          }
        }
      }
    } catch (e) {
      if (!mounted) return;
      final errorMessage = e.toString().toLowerCase();
      final shouldReturnHome = errorMessage.contains('已被') ||
          errorMessage.contains('接走') ||
          errorMessage.contains('已被接') ||
          errorMessage.contains('其他配送员') ||
          errorMessage.contains('已被其他');
      
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('接单失败: ${e.toString()}'),
          backgroundColor: Colors.red,
          duration: const Duration(seconds: 2),
        ),
      );
      
      setState(() {
        _acceptingOrders.remove(orderId);
      });
      
      // 如果是订单已被接走等错误，自动返回首页
      if (shouldReturnHome && mounted) {
        await Future.delayed(const Duration(milliseconds: 500));
        if (mounted) {
          Navigator.of(context).popUntil((route) => route.isFirst);
        }
      }
    }
  }

  void _viewOrderItems(Map<String, dynamic> order) async {
    final orderId = (order['id'] as num?)?.toInt();

    if (orderId == null) return;

    // 跳转到订单详情页面，并等待返回结果
    final result = await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => OrderDetailView(orderId: orderId)),
    );

    // 返回首页时，如果订单状态发生了变化，刷新列表和角标
    if (mounted) {
      if (result == true) {
        // 如果订单状态发生了变化，先等待一下确保后端已处理
        await Future.delayed(const Duration(milliseconds: 300));
        // 然后刷新列表
        await _loadOrders(reset: true);
        // 等待列表刷新完成后再更新角标
        await Future.delayed(const Duration(milliseconds: 300));
        // 更新角标数量（强制刷新）
        if (widget.onOrderCountChanged != null) {
          await widget.onOrderCountChanged!();
        }
      } else {
        // 即使没有变化，也更新一次角标数量（确保角标准确）
        // 但延迟更短，避免不必要的等待
        await Future.delayed(const Duration(milliseconds: 200));
        if (widget.onOrderCountChanged != null) {
          await widget.onOrderCountChanged!();
        }
      }
    }
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
            child: ListView.builder(
              controller: _scrollController,
              // 设置为始终可滚动，即使内容为空也能下拉刷新
              physics: const AlwaysScrollableScrollPhysics(),
              padding: EdgeInsets.fromLTRB(
                16,
                0,
                16,
                16 + MediaQuery.of(context).padding.bottom,
              ),
              itemCount: _orders.isEmpty
                  ? 1 // 空状态时显示一个空状态项
                  : _orders.length + (_hasMore ? 1 : 0),
              itemBuilder: (context, index) {
                // 空状态显示
                if (_orders.isEmpty) {
                  return SizedBox(
                    height: MediaQuery.of(context).size.height * 0.6,
                    child: Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(
                            Icons.inbox_outlined,
                            size: 42,
                            color: Color(0xFF20CB6B),
                          ),
                          const SizedBox(height: 16),
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
                  );
                }

                // 加载更多指示器
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

                // 订单卡片
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
                  // 门店名称（地址名称，第一行，突出显示）
                  Text(
                    storeName,
                    style: const TextStyle(
                      fontSize: 17,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF20253A),
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 8),
                  // 地址（第二行）
                  if (address.isNotEmpty)
                    Text(
                      address,
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w400,
                        color: Color(0xFF40475C),
                        height: 1.4,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  const SizedBox(height: 8),
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
