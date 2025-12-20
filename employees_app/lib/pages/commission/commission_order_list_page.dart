import 'package:flutter/material.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/order/order_detail_page.dart';

/// 全部订单查询页面
class CommissionOrderListPage extends StatefulWidget {
  const CommissionOrderListPage({super.key});

  @override
  State<CommissionOrderListPage> createState() =>
      _CommissionOrderListPageState();
}

class _CommissionOrderListPageState extends State<CommissionOrderListPage> {
  final ScrollController _scrollController = ScrollController();
  final List<Map<String, dynamic>> _orders = [];
  bool _isLoading = false;
  bool _isLoadingMore = false;
  bool _hasMore = true;
  int _pageNum = 1;
  final int _pageSize = 20;
  String _selectedStatus =
      'all'; // all, invalid, unaccounted, accounted, settled

  final List<Map<String, String>> _statusOptions = [
    {'value': 'all', 'label': '全部'},
    {'value': 'invalid', 'label': '无效订单'},
    {'value': 'unaccounted', 'label': '未计入'},
    {'value': 'accounted', 'label': '已计入'},
    {'value': 'settled', 'label': '已结算'},
  ];

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

    // 如果是"未计入"状态，调用未收款订单接口
    if (_selectedStatus == 'unaccounted') {
      final response = await Request.get<Map<String, dynamic>>(
        '/employee/sales/commission/unpaid-orders',
        queryParams: {
          'pageNum': _pageNum.toString(),
          'pageSize': _pageSize.toString(),
        },
        parser: (data) => data as Map<String, dynamic>,
      );

      if (!mounted) return;

      if (response.isSuccess && response.data != null) {
        final data = response.data!;
        final List<dynamic> list = (data['list'] as List<dynamic>? ?? []);
        final orders = list.cast<Map<String, dynamic>>();

        // 转换数据格式，添加commission_preview信息
        final transformedOrders = orders.map((order) {
          final preview = order['commission_preview'] as Map<String, dynamic>?;
          return {
            ...order,
            'order_date': order['order_date'] ?? order['created_at'],
            'total_commission': preview?['total_commission'] ?? 0,
            'base_commission': preview?['base_commission'] ?? 0,
            'new_customer_bonus': preview?['new_customer_bonus'] ?? 0,
            'tier_commission': preview?['tier_commission'] ?? 0,
            'tier_level': preview?['tier_level'] ?? 0,
            'is_new_customer_order': preview?['is_new_customer_order'] ?? false,
            'is_valid_order': preview?['is_valid_order'] ?? true,
            'order_profit': preview?['order_profit'] ?? 0,
            'is_accounted': false,
            'is_settled': false,
            'is_accounted_cancelled': false,
          };
        }).toList();

        setState(() {
          if (reset) {
            _orders
              ..clear()
              ..addAll(transformedOrders);
          } else {
            _orders.addAll(transformedOrders);
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
              content: Text(response.message),
              backgroundColor: Colors.red,
              behavior: SnackBarBehavior.floating,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
          );
        }
      }
      return;
    }

    // 其他状态调用原来的接口
    final status = _selectedStatus == 'all' ? null : _selectedStatus;

    final response = await Request.get<Map<String, dynamic>>(
      '/employee/sales/commission/list',
      queryParams: {
        'pageNum': _pageNum.toString(),
        'pageSize': _pageSize.toString(),
        if (status != null) 'status': status,
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
            content: Text(response.message),
            backgroundColor: Colors.red,
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        );
      }
    }
  }

  void _onStatusChanged(String? value) {
    if (value != null && value != _selectedStatus) {
      setState(() {
        _selectedStatus = value;
      });
      _loadOrders(reset: true);
    }
  }

  String _formatMoney(dynamic value) {
    if (value == null) return '0.00';
    if (value is num) {
      return value.toStringAsFixed(2);
    }
    return value.toString();
  }

  String _getStatusText(Map<String, dynamic> order) {
    // 如果是未计入状态，显示订单状态
    if (_selectedStatus == 'unaccounted' ||
        (order['is_accounted'] == false &&
            order['is_accounted_cancelled'] != true)) {
      final status = order['status'] as String? ?? '';
      if (status == 'pending' || status == 'pending_delivery') {
        return '待配送';
      } else if (status == 'delivered' || status == 'shipped') {
        return '已配送';
      } else if (status == 'cancelled') {
        return '已取消';
      } else if (status == 'paid') {
        return '已收款';
      }
      return '未计入';
    }
    if (order['is_accounted_cancelled'] == true) {
      return '计入已取消';
    }
    if (order['is_settled'] == true) {
      return '已结算';
    }
    if (order['is_accounted'] == true) {
      return '已计入';
    }
    if (order['is_valid_order'] == false) {
      return '无效订单';
    }
    return '未计入';
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case '已结算':
        return const Color(0xFF20CB6B);
      case '已计入':
        return const Color(0xFF4C8DF6);
      case '未计入':
      case '待配送':
      case '已配送':
        return const Color(0xFFFFA940);
      case '计入已取消':
        return const Color(0xFFFF5A5F);
      case '无效订单':
      case '已取消':
        return const Color(0xFF8C92A4);
      case '已收款':
        return const Color(0xFF4C8DF6);
      default:
        return const Color(0xFF8C92A4);
    }
  }

  String _formatDate(String? dateStr) {
    if (dateStr == null || dateStr.isEmpty) return '-';
    try {
      if (dateStr.length >= 10) {
        return dateStr.substring(0, 10);
      }
      return dateStr;
    } catch (_) {
      return dateStr;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      extendBody: true,
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [Color(0xFF20CB6B), Color(0xFFEFF7F2)],
          ),
        ),
        child: SafeArea(
          bottom: false,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // 顶部标题区域（与首页风格一致）
              Padding(
                padding: const EdgeInsets.fromLTRB(20, 16, 20, 16),
                child: Row(
                  children: [
                    IconButton(
                      icon: const Icon(Icons.arrow_back, color: Colors.white),
                      onPressed: () => Navigator.of(context).pop(),
                      padding: EdgeInsets.zero,
                      constraints: const BoxConstraints(),
                    ),
                    const SizedBox(width: 12),
                    const Expanded(
                      child: Text(
                        '全部订单查询',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 22,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ],
                ),
              ),

              // 内容区域（白色背景，圆角顶部）
              Expanded(
                child: Container(
                  decoration: const BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.only(
                      topLeft: Radius.circular(24),
                      topRight: Radius.circular(24),
                    ),
                  ),
                  child: Column(
                    children: [
                      // 状态筛选栏
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 12,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: const BorderRadius.only(
                            topLeft: Radius.circular(24),
                            topRight: Radius.circular(24),
                          ),
                        ),
                        child: SingleChildScrollView(
                          scrollDirection: Axis.horizontal,
                          child: Row(
                            children: _statusOptions.map((option) {
                              final isSelected =
                                  _selectedStatus == option['value'];
                              return Padding(
                                padding: const EdgeInsets.only(right: 8),
                                child: FilterChip(
                                  label: Text(option['label']!),
                                  selected: isSelected,
                                  onSelected: (selected) {
                                    if (selected) {
                                      _onStatusChanged(option['value']);
                                    }
                                  },
                                  selectedColor: const Color(0xFF20CB6B),
                                  backgroundColor: const Color(0xFFF7F8FA),
                                  labelStyle: TextStyle(
                                    color: isSelected
                                        ? Colors.white
                                        : const Color(0xFF20253A),
                                    fontSize: 13,
                                    fontWeight: FontWeight.w500,
                                  ),
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 14,
                                    vertical: 6,
                                  ),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(18),
                                    side: BorderSide(
                                      color: isSelected
                                          ? const Color(0xFF20CB6B)
                                          : const Color(0xFFE5E7EB),
                                      width: 1,
                                    ),
                                  ),
                                ),
                              );
                            }).toList(),
                          ),
                        ),
                      ),

                      // 订单列表
                      Expanded(
                        child: _isLoading && _orders.isEmpty
                            ? Center(
                                child: Column(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    const CircularProgressIndicator(
                                      valueColor: AlwaysStoppedAnimation<Color>(
                                        Color(0xFF20CB6B),
                                      ),
                                    ),
                                    const SizedBox(height: 16),
                                    Text(
                                      '加载中...',
                                      style: TextStyle(
                                        color: Colors.grey[400],
                                        fontSize: 14,
                                      ),
                                    ),
                                  ],
                                ),
                              )
                            : _orders.isEmpty
                            ? Center(
                                child: Column(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Container(
                                      padding: const EdgeInsets.all(24),
                                      decoration: BoxDecoration(
                                        color: Colors.white,
                                        shape: BoxShape.circle,
                                        boxShadow: [
                                          BoxShadow(
                                            color: Colors.black.withOpacity(
                                              0.05,
                                            ),
                                            blurRadius: 20,
                                            offset: const Offset(0, 4),
                                          ),
                                        ],
                                      ),
                                      child: Icon(
                                        Icons.inbox_outlined,
                                        color: Colors.grey[300],
                                        size: 64,
                                      ),
                                    ),
                                    const SizedBox(height: 24),
                                    Text(
                                      '暂无订单',
                                      style: TextStyle(
                                        fontSize: 16,
                                        color: Colors.grey[500],
                                        fontWeight: FontWeight.w500,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    Text(
                                      '切换筛选条件试试',
                                      style: TextStyle(
                                        fontSize: 14,
                                        color: Colors.grey[400],
                                      ),
                                    ),
                                  ],
                                ),
                              )
                            : RefreshIndicator(
                                onRefresh: () => _loadOrders(reset: true),
                                color: const Color(0xFF20CB6B),
                                child: ListView.builder(
                                  controller: _scrollController,
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 16,
                                    vertical: 12,
                                  ),
                                  itemCount:
                                      _orders.length + (_hasMore ? 1 : 0),
                                  itemBuilder: (context, index) {
                                    if (index >= _orders.length) {
                                      return Padding(
                                        padding: const EdgeInsets.symmetric(
                                          vertical: 20,
                                        ),
                                        child: Center(
                                          child: Column(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              const CircularProgressIndicator(
                                                valueColor:
                                                    AlwaysStoppedAnimation<
                                                      Color
                                                    >(Color(0xFF20CB6B)),
                                                strokeWidth: 2,
                                              ),
                                              const SizedBox(height: 12),
                                              Text(
                                                '加载更多...',
                                                style: TextStyle(
                                                  color: Colors.grey[400],
                                                  fontSize: 12,
                                                ),
                                              ),
                                            ],
                                          ),
                                        ),
                                      );
                                    }
                                    return Padding(
                                      padding: const EdgeInsets.only(
                                        bottom: 10,
                                      ),
                                      child: _buildOrderCard(_orders[index]),
                                    );
                                  },
                                ),
                              ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildOrderCard(Map<String, dynamic> order) {
    final status = _getStatusText(order);
    final statusColor = _getStatusColor(status);
    final orderNumber = order['order_number'] as String? ?? '-';
    final addressName =
        order['address_name'] as String? ??
        order['store_name'] as String? ??
        '-';
    final orderDate = _formatDate(
      order['order_date']?.toString() ?? order['created_at']?.toString(),
    );
    final totalCommission = _formatMoney(order['total_commission'] ?? 0);
    final orderAmount = _formatMoney(order['order_amount'] ?? 0);
    final isPreview =
        _selectedStatus == 'unaccounted' || order['is_preview'] == true;

    // 获取订单ID（可能是 order_id 或 id）
    final orderId =
        order['order_id'] as int? ??
        order['id'] as int? ??
        (order['order_id'] is num
            ? (order['order_id'] as num).toInt()
            : null) ??
        (order['id'] is num ? (order['id'] as num).toInt() : null);

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
      borderRadius: BorderRadius.circular(16),
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.05),
              blurRadius: 10,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          addressName,
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          orderNumber,
                          style: const TextStyle(
                            fontSize: 12,
                            color: Color(0xFF8C92A4),
                          ),
                        ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      status,
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: statusColor,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          if (isPreview) ...[
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 2,
                              ),
                              margin: const EdgeInsets.only(right: 8),
                              decoration: BoxDecoration(
                                color: const Color(0xFFFFA940).withOpacity(0.1),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: const Text(
                                '分润预览',
                                style: TextStyle(
                                  fontSize: 11,
                                  color: Color(0xFFFFA940),
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          ],
                          const Text(
                            '订单金额',
                            style: TextStyle(
                              fontSize: 14,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                          const SizedBox(width: 4),
                          Text(
                            '¥$orderAmount',
                            style: const TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w700,
                              color: Color(0xFF20253A),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 6),
                      Row(
                        children: [
                          const Text(
                            '分润金额',
                            style: TextStyle(
                              fontSize: 14,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                          const SizedBox(width: 4),
                          Text(
                            '+¥$totalCommission',
                            style: const TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w700,
                              color: Color(0xFF20CB6B),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                  Text(
                    orderDate,
                    style: const TextStyle(
                      fontSize: 12,
                      color: Color(0xFF8C92A4),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
