import 'dart:async';
import 'package:flutter/material.dart';
import 'package:super_app/api/orders_api.dart';
import 'package:super_app/api/payment_verification_api.dart';
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
  final TextEditingController _searchController = TextEditingController();
  String _searchKeyword = '';
  Timer? _debounce;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
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
    _searchController.dispose();
    _debounce?.cancel();
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

  void _onSearchChanged(String value) {
    // 取消之前的防抖任务
    _debounce?.cancel();

    // 防抖：延迟500ms后执行搜索
    _debounce = Timer(const Duration(milliseconds: 500), () {
      if (mounted && _searchController.text == value) {
        final keyword = value.trim();
        if (_searchKeyword != keyword) {
          setState(() {
            _searchKeyword = keyword;
            _currentPage = 1;
            _hasMore = true;
            _orders.clear();
          });
          _loadOrders();
        }
      }
    });
  }

  void _clearSearch() {
    _searchController.clear();
    setState(() {
      _searchKeyword = '';
      _currentPage = 1;
      _hasMore = true;
      _orders.clear();
    });
    _loadOrders();
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
        // 待审核：只显示有收款审核申请的订单（status='pending'）
        // 获取待审核的收款审核申请，获取订单ID列表
        final verificationResponse =
            await PaymentVerificationApi.getPaymentVerifications(
              status: 'pending',
              pageNum: _currentPage,
              pageSize: _pageSize,
            );

        if (!mounted) return;

        if (verificationResponse.isSuccess &&
            verificationResponse.data != null) {
          final verificationData = verificationResponse.data!;
          final verificationList = verificationData.list;

          // 调试信息：打印收款审核申请列表
          debugPrint('收款审核申请总数: ${verificationData.total}');
          debugPrint('当前页申请数量: ${verificationList.length}');
          if (verificationList.isNotEmpty) {
            debugPrint('第一个申请数据: ${verificationList.first}');
          }

          // 提取订单ID列表
          final orderIds = <int>[];
          for (final v in verificationList) {
            // 尝试多种可能的字段名
            final orderId = v['order_id'] ?? v['orderId'] ?? v['order_ID'];
            int? id;
            if (orderId is num) {
              id = orderId.toInt();
            } else if (orderId is int) {
              id = orderId;
            }
            if (id != null && id > 0) {
              orderIds.add(id);
            }
          }

          if (orderIds.isNotEmpty) {
            // 并行查询所有订单详情
            final orderDetailFutures = orderIds.map((orderId) async {
              try {
                final orderDetailResponse = await OrdersApi.getOrderDetail(
                  orderId,
                );
                if (orderDetailResponse.isSuccess &&
                    orderDetailResponse.data != null) {
                  final responseData = orderDetailResponse.data!;

                  // 订单详情API返回的数据结构：{"order": {...}, "address": {...}, "order_items": [...], ...}
                  // 需要从 order 字段中获取订单数据，并合并 address 和 item_count
                  Map<String, dynamic> orderData;
                  if (responseData.containsKey('order') &&
                      responseData['order'] is Map<String, dynamic>) {
                    orderData = Map<String, dynamic>.from(
                      responseData['order'] as Map<String, dynamic>,
                    );

                    // 合并地址信息（在响应的顶层）
                    if (responseData.containsKey('address') &&
                        responseData['address'] is Map<String, dynamic>) {
                      orderData['address'] = responseData['address'];
                    }

                    // 合并 item_count（如果在响应的顶层，或从 order_items 计算）
                    if (responseData.containsKey('item_count')) {
                      orderData['item_count'] = responseData['item_count'];
                    } else if (responseData.containsKey('order_items') &&
                        responseData['order_items'] is List) {
                      final orderItems = responseData['order_items'] as List;
                      orderData['item_count'] = orderItems.length;
                    }
                  } else {
                    // 如果没有order字段，尝试直接使用data（兼容不同的API响应格式）
                    orderData = responseData;
                  }

                  try {
                    return Order.fromJson(orderData);
                  } catch (e) {
                    // 解析失败，打印错误
                    if (mounted) {
                      debugPrint('订单 $orderId 解析失败: $e');
                      debugPrint('订单数据: $orderData');
                      debugPrint('完整响应: $responseData');
                    }
                    return null;
                  }
                } else {
                  if (mounted) {
                    debugPrint(
                      '订单 $orderId 查询失败: ${orderDetailResponse.message}',
                    );
                  }
                }
              } catch (e) {
                // 查询失败，打印错误
                if (mounted) {
                  debugPrint('订单 $orderId 查询异常: $e');
                }
              }
              return null;
            }).toList();

            // 等待所有查询完成
            final orderResults = await Future.wait(orderDetailFutures);

            // 添加非null的订单
            for (var order in orderResults) {
              if (order != null) {
                newOrders.add(order);
              }
            }
          } else {
            // 如果没有提取到订单ID，打印调试信息
            if (mounted && verificationList.isNotEmpty) {
              debugPrint('收款审核申请列表不为空，但未提取到订单ID');
              debugPrint('申请数量: ${verificationList.length}');
              debugPrint('第一个申请的数据: ${verificationList.first}');
              if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(
                      '收款审核申请: ${verificationList.length}条，但无法获取订单ID',
                    ),
                    duration: const Duration(seconds: 3),
                  ),
                );
              }
            } else if (mounted && verificationList.isEmpty) {
              debugPrint('收款审核申请列表为空');
            }
          }

          // 判断是否有更多数据（基于收款审核申请的总数）
          final totalOrders = _currentPage == 1
              ? newOrders.length
              : _orders.length + newOrders.length;
          _hasMore = totalOrders < verificationData.total;
        } else {
          setState(() {
            _isLoading = false;
          });
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(verificationResponse.message)),
            );
          }
          return;
        }
      } else {
        // 全部订单：不筛选状态，但可以使用搜索关键词
        final keyword = _searchKeyword.isNotEmpty ? _searchKeyword : null;
        debugPrint('搜索订单，关键词: $keyword, 页码: $_currentPage');

        final response = await OrdersApi.getOrders(
          pageNum: _currentPage,
          pageSize: _pageSize,
          status: null,
          keyword: keyword,
        );

        if (!mounted) return;

        debugPrint(
          '订单搜索响应: code=${response.code}, message=${response.message}',
        );
        debugPrint('订单搜索结果数量: ${response.data?.list.length ?? 0}');

        if (response.isSuccess && response.data != null) {
          newOrders = response.data!.list;
          debugPrint('获取到 ${newOrders.length} 个订单');
        } else {
          setState(() {
            _isLoading = false;
          });
          if (mounted) {
            ScaffoldMessenger.of(
              context,
            ).showSnackBar(SnackBar(content: Text(response.message)));
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

        // 对于待审核tab，_hasMore已经在上面设置了
        // 对于全部订单tab，判断是否有更多数据
        if (tabIndex == 0) {
          _hasMore = newOrders.length >= _pageSize;
        }

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
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('加载失败: ${e.toString()}')));
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
          // 搜索框
          Container(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            color: Colors.white,
            child: TextField(
              controller: _searchController,
              onChanged: (value) {
                setState(() {}); // 更新UI以显示/隐藏清除按钮
                _onSearchChanged(value);
              },
              decoration: InputDecoration(
                hintText: '搜索订单号、收货人、电话',
                prefixIcon: const Icon(Icons.search, color: Color(0xFF8C92A4)),
                suffixIcon: _searchController.text.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.clear, color: Color(0xFF8C92A4)),
                        onPressed: () {
                          _clearSearch();
                        },
                      )
                    : null,
                filled: true,
                fillColor: Colors.grey.shade100,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide.none,
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 12,
                ),
              ),
            ),
          ),
          // 标签栏
          Container(
            color: Colors.white,
            child: TabBar(
              controller: _tabController,
              labelColor: const Color(0xFF20CB6B),
              unselectedLabelColor: const Color(0xFF8C92A4),
              indicatorColor: const Color(0xFF20CB6B),
              tabs: const [
                Tab(text: '全部订单'),
                Tab(text: '待审核'),
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
              style: TextStyle(fontSize: 16, color: Colors.grey[600]),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.all(16),
      itemCount: _orders.length + (_hasMore ? 1 : 0),
      itemBuilder: (context, index) {
        if (index >= _orders.length) {
          return const Padding(
            padding: EdgeInsets.symmetric(vertical: 16),
            child: Center(
              child: CircularProgressIndicator(
                valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
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

    // 根据订单状态获取背景色
    Color statusBgColor = _getStatusBackgroundColor(order.status);

    return InkWell(
      onTap: () {
        Navigator.of(context).pushNamed('/order_detail', arguments: order.id);
      },
      borderRadius: BorderRadius.circular(12),
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: statusBgColor,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.shade200, width: 1),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 4,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 加急标签
            if (order.isUrgent) ...[
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFFFF6B6B),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: const Text(
                  '加急',
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                    color: Colors.white,
                  ),
                ),
              ),
              const SizedBox(height: 12),
            ],
            // 地址名称
            Text(
              addressName.isNotEmpty ? addressName : '地址名称未填写',
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: Color(0xFF20253A),
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
            const SizedBox(height: 8),
            // 联系电话和地址
            if (contact.isNotEmpty) ...[
              Text(
                contact,
                style: const TextStyle(fontSize: 14, color: Color(0xFF40475C)),
              ),
              const SizedBox(height: 4),
            ],
            if (addressText.isNotEmpty) ...[
              Text(
                addressText,
                style: const TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ],
            const SizedBox(height: 12),
            // 商品数量和创建时间
            Row(
              children: [
                Text(
                  '共${order.itemCount ?? 0}件商品',
                  style: const TextStyle(
                    fontSize: 13,
                    color: Color(0xFF8C92A4),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    _formatDateTime(order.createdAt),
                    style: const TextStyle(
                      fontSize: 12,
                      color: Color(0xFFB0B4C3),
                    ),
                    textAlign: TextAlign.right,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            // 状态和实付金额
            Row(
              children: [
                _buildOrderStatus(order.status),
                const Spacer(),
                const Text(
                  '实付金额：',
                  style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
                ),
                Text(
                  '¥${order.totalAmount.toStringAsFixed(2)}',
                  style: const TextStyle(
                    fontSize: 18,
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

  // 根据订单状态获取背景色（浅色背景）
  Color _getStatusBackgroundColor(String status) {
    switch (status) {
      case 'pending_delivery':
        return const Color(0xFFFFF4E6); // 橙色浅背景
      case 'pending_pickup':
        return const Color(0xFFE6F2FF); // 蓝色浅背景
      case 'delivering':
        return const Color(0xFFE6F9F0); // 绿色浅背景
      case 'delivered':
        return const Color(0xFFE6F9F0); // 绿色浅背景
      case 'paid':
        return const Color(0xFFF3EFFF); // 紫色浅背景
      case 'cancelled':
        return const Color(0xFFF5F5F5); // 灰色浅背景
      default:
        return Colors.white;
    }
  }

  // 根据订单状态获取文字颜色
  Color _getStatusTextColor(String status) {
    switch (status) {
      case 'pending_delivery':
        return const Color(0xFFFFA940); // 橙色
      case 'pending_pickup':
        return const Color(0xFF4C8DF6); // 蓝色
      case 'delivering':
        return const Color(0xFF20CB6B); // 绿色
      case 'delivered':
        return const Color(0xFF20CB6B); // 绿色
      case 'paid':
        return const Color(0xFF7C4DFF); // 紫色
      case 'cancelled':
        return const Color(0xFFB0B4C3); // 灰色
      default:
        return const Color(0xFF8C92A4);
    }
  }

  Widget _buildOrderStatus(String status) {
    Color color = _getStatusTextColor(status);
    String text;

    switch (status) {
      case 'pending_delivery':
        text = '待配送';
        break;
      case 'pending_pickup':
        text = '待取货';
        break;
      case 'delivering':
        text = '配送中';
        break;
      case 'delivered':
        text = '已送达';
        break;
      case 'paid':
        text = '已收款';
        break;
      case 'cancelled':
        text = '已取消';
        break;
      default:
        text = status;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withOpacity(0.15),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: color.withOpacity(0.3), width: 1),
      ),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 13,
          color: color,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  String _formatDateTime(DateTime dateTime) {
    // 格式：MM-DD HH:mm
    return '${dateTime.month.toString().padLeft(2, '0')}-${dateTime.day.toString().padLeft(2, '0')} '
        '${dateTime.hour.toString().padLeft(2, '0')}:${dateTime.minute.toString().padLeft(2, '0')}';
  }
}
