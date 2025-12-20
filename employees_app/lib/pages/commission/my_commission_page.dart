import 'package:flutter/material.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/commission/commission_order_list_page.dart';
import 'package:employees_app/pages/order/order_detail_page.dart';

class MyCommissionPage extends StatefulWidget {
  const MyCommissionPage({super.key});

  @override
  State<MyCommissionPage> createState() => _MyCommissionPageState();
}

class _MyCommissionPageState extends State<MyCommissionPage> {
  bool _isLoading = true;
  String? _errorMessage;
  Map<String, dynamic>? _overview;
  List<Map<String, dynamic>> _recentOrders = [];
  String _selectedPeriod = 'all'; // all, today, week, month, year

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  // 获取时间范围
  Map<String, String?> _getDateRange() {
    final now = DateTime.now();
    String? startDate;
    String? endDate;

    switch (_selectedPeriod) {
      case 'today':
        startDate = _formatDateForApi(now);
        endDate = _formatDateForApi(now);
        break;
      case 'week':
        final weekStart = now.subtract(Duration(days: now.weekday - 1));
        startDate = _formatDateForApi(weekStart);
        endDate = _formatDateForApi(now);
        break;
      case 'month':
        startDate = _formatDateForApi(DateTime(now.year, now.month, 1));
        endDate = _formatDateForApi(now);
        break;
      case 'year':
        startDate = _formatDateForApi(DateTime(now.year, 1, 1));
        endDate = _formatDateForApi(now);
        break;
      default:
        startDate = null;
        endDate = null;
    }

    return {'start_date': startDate, 'end_date': endDate};
  }

  String _formatDateForApi(DateTime date) {
    return '${date.year.toString().padLeft(4, '0')}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
  }

  Future<void> _loadData() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      // 获取时间范围
      final dateRange = _getDateRange();
      final queryParams = <String, String>{};
      if (dateRange['start_date'] != null) {
        queryParams['start_date'] = dateRange['start_date']!;
      }
      if (dateRange['end_date'] != null) {
        queryParams['end_date'] = dateRange['end_date']!;
      }

      // 获取总览统计
      final overviewResponse = await Request.get<Map<String, dynamic>>(
        '/employee/sales/commission/overview',
        queryParams: queryParams,
        parser: (data) => data as Map<String, dynamic>,
      );

      if (overviewResponse.isSuccess && overviewResponse.data != null) {
        setState(() {
          _overview = overviewResponse.data;
        });
      }

      // 获取最近订单
      final ordersResponse = await Request.get<Map<String, dynamic>>(
        '/employee/sales/commission/list',
        queryParams: {'pageNum': '1', 'pageSize': '5'},
        parser: (data) => data as Map<String, dynamic>,
      );

      if (ordersResponse.isSuccess && ordersResponse.data != null) {
        final list = (ordersResponse.data!['list'] as List<dynamic>? ?? [])
            .cast<Map<String, dynamic>>();
        setState(() {
          _recentOrders = list;
        });
      }

      setState(() {
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isLoading = false;
      });
    }
  }

  String _formatMoney(dynamic amount) {
    if (amount == null) return '0.00';
    if (amount is num) {
      return amount.toDouble().toStringAsFixed(2);
    }
    return amount.toString();
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

  String _getPeriodLabel(String period) {
    switch (period) {
      case 'today':
        return '今天';
      case 'week':
        return '7日';
      case 'month':
        return '月';
      case 'year':
        return '年';
      default:
        return '全部';
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
                      Container(
                        padding: const EdgeInsets.all(20),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          shape: BoxShape.circle,
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.1),
                              blurRadius: 20,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                        child: const Icon(
                          Icons.error_outline,
                          color: Color(0xFFFF5A5F),
                          size: 48,
                        ),
                      ),
                      const SizedBox(height: 24),
                      Text(
                        _errorMessage!,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 15,
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 24),
                      ElevatedButton.icon(
                        onPressed: _loadData,
                        icon: const Icon(Icons.refresh, size: 18),
                        label: const Text('重试'),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.white,
                          foregroundColor: const Color(0xFF20CB6B),
                          padding: const EdgeInsets.symmetric(
                            horizontal: 24,
                            vertical: 12,
                          ),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(8),
                          ),
                        ),
                      ),
                    ],
                  ),
                )
              : Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // 顶部标题和总金额区域（与首页风格一致）
                    Padding(
                      padding: const EdgeInsets.fromLTRB(20, 16, 20, 20),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          // 返回按钮和标题
                          Row(
                            children: [
                              IconButton(
                                icon: const Icon(
                                  Icons.arrow_back,
                                  color: Colors.white,
                                ),
                                onPressed: () => Navigator.of(context).pop(),
                                padding: EdgeInsets.zero,
                                constraints: const BoxConstraints(),
                              ),
                              const SizedBox(width: 12),
                              const Expanded(
                                child: Text(
                                  '我的分润',
                                  style: TextStyle(
                                    color: Colors.white,
                                    fontSize: 22,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                              // 时间筛选选择器
                              PopupMenuButton<String>(
                                initialValue: _selectedPeriod,
                                onSelected: (value) {
                                  setState(() {
                                    _selectedPeriod = value;
                                  });
                                  _loadData();
                                },
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 12,
                                    vertical: 6,
                                  ),
                                  decoration: BoxDecoration(
                                    color: Colors.white.withOpacity(0.2),
                                    borderRadius: BorderRadius.circular(20),
                                    border: Border.all(
                                      color: Colors.white.withOpacity(0.3),
                                      width: 1,
                                    ),
                                  ),
                                  child: Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Text(
                                        _getPeriodLabel(_selectedPeriod),
                                        style: const TextStyle(
                                          color: Colors.white,
                                          fontSize: 14,
                                          fontWeight: FontWeight.w500,
                                        ),
                                      ),
                                      const SizedBox(width: 4),
                                      const Icon(
                                        Icons.arrow_drop_down,
                                        color: Colors.white,
                                        size: 20,
                                      ),
                                    ],
                                  ),
                                ),
                                itemBuilder: (context) => [
                                  const PopupMenuItem<String>(
                                    value: 'all',
                                    child: Text('全部'),
                                  ),
                                  const PopupMenuItem<String>(
                                    value: 'today',
                                    child: Text('今天'),
                                  ),
                                  const PopupMenuItem<String>(
                                    value: 'week',
                                    child: Text('7日'),
                                  ),
                                  const PopupMenuItem<String>(
                                    value: 'month',
                                    child: Text('月'),
                                  ),
                                  const PopupMenuItem<String>(
                                    value: 'year',
                                    child: Text('年'),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),

                    // 内容区域（白色背景，圆角顶部）
                    Expanded(
                      child: RefreshIndicator(
                        onRefresh: _loadData,
                        color: const Color(0xFF20CB6B),
                        child: Container(
                          decoration: const BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.only(
                              topLeft: Radius.circular(24),
                              topRight: Radius.circular(24),
                            ),
                          ),
                          child: SingleChildScrollView(
                            physics: const AlwaysScrollableScrollPhysics(),
                            padding: EdgeInsets.fromLTRB(
                              16,
                              16,
                              16,
                              16 + MediaQuery.of(context).padding.bottom,
                            ),
                            child: Column(
                              children: [
                                // 总金额卡片（简约白色卡片）
                                Container(
                                  decoration: BoxDecoration(
                                    color: Colors.white,
                                    borderRadius: BorderRadius.circular(20),
                                    boxShadow: [
                                      BoxShadow(
                                        color: Colors.black.withOpacity(0.06),
                                        blurRadius: 20,
                                        offset: const Offset(0, 4),
                                        spreadRadius: 0,
                                      ),
                                    ],
                                  ),
                                  padding: const EdgeInsets.all(24),
                                  child: Column(
                                    children: [
                                      const Text(
                                        '预计总分成金额',
                                        style: TextStyle(
                                          fontSize: 16,
                                          color: Color(0xFF8C92A4),
                                          fontWeight: FontWeight.w500,
                                        ),
                                      ),
                                      const SizedBox(height: 8),
                                      Text(
                                        '¥${_formatMoney(_overview?['total_amount'] ?? 0)}',
                                        style: const TextStyle(
                                          fontSize: 36,
                                          fontWeight: FontWeight.w700,
                                          color: Color(0xFF20253A),
                                        ),
                                      ),
                                      const SizedBox(height: 16),
                                      Row(
                                        children: [
                                          Expanded(
                                            child: _buildStatItem(
                                              '未计入',
                                              _formatMoney(
                                                _overview?['unaccounted_amount'] ??
                                                    0,
                                              ),
                                              const Color(0xFFFFA940),
                                            ),
                                          ),
                                          const SizedBox(width: 16),
                                          Expanded(
                                            child: _buildStatItem(
                                              '已计入',
                                              _formatMoney(
                                                _overview?['accounted_amount'] ??
                                                    0,
                                              ),
                                              const Color(0xFF4C8DF6),
                                            ),
                                          ),
                                        ],
                                      ),
                                      const SizedBox(height: 12),
                                      Row(
                                        children: [
                                          Expanded(
                                            child: _buildStatItem(
                                              '已结算',
                                              _formatMoney(
                                                _overview?['settled_amount'] ??
                                                    0,
                                              ),
                                              const Color(0xFF20CB6B),
                                            ),
                                          ),
                                          const SizedBox(width: 16),
                                          Expanded(
                                            child: _buildStatItem(
                                              '取消计入',
                                              _formatMoney(
                                                _overview?['cancelled_amount'] ??
                                                    0,
                                              ),
                                              const Color(0xFFFF5A5F),
                                            ),
                                          ),
                                        ],
                                      ),
                                      if ((_overview?['invalid_order_count'] ??
                                              0) >
                                          0) ...[
                                        const SizedBox(height: 16),
                                        Container(
                                          padding: const EdgeInsets.all(12),
                                          decoration: BoxDecoration(
                                            color: const Color(0xFFF5F7FA),
                                            borderRadius: BorderRadius.circular(
                                              12,
                                            ),
                                          ),
                                          child: Row(
                                            mainAxisAlignment:
                                                MainAxisAlignment.center,
                                            children: [
                                              const Icon(
                                                Icons.block,
                                                size: 18,
                                                color: Color(0xFF8C92A4),
                                              ),
                                              const SizedBox(width: 8),
                                              Text(
                                                '无效订单：${_overview?['invalid_order_count'] ?? 0}单',
                                                style: const TextStyle(
                                                  fontSize: 14,
                                                  color: Color(0xFF8C92A4),
                                                ),
                                              ),
                                            ],
                                          ),
                                        ),
                                      ],
                                    ],
                                  ),
                                ),
                                const SizedBox(height: 20),

                                // 最近订单区域
                                if (_recentOrders.isNotEmpty) ...[
                                  Row(
                                    mainAxisAlignment:
                                        MainAxisAlignment.spaceBetween,
                                    children: [
                                      const Text(
                                        '最近订单',
                                        style: TextStyle(
                                          fontSize: 18,
                                          fontWeight: FontWeight.w600,
                                          color: Color(0xFF20253A),
                                        ),
                                      ),
                                      TextButton(
                                        onPressed: () {
                                          Navigator.of(context).push(
                                            MaterialPageRoute(
                                              builder: (_) =>
                                                  const CommissionOrderListPage(),
                                            ),
                                          );
                                        },
                                        child: const Text(
                                          '查看全部',
                                          style: TextStyle(
                                            color: Color(0xFF20CB6B),
                                            fontSize: 14,
                                            fontWeight: FontWeight.w500,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 12),
                                  ..._recentOrders.map(
                                    (order) => _buildOrderCard(order),
                                  ),
                                ],
                              ],
                            ),
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
        ),
      ),
    );
  }

  Widget _buildStatItem(String label, String value, Color color) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        children: [
          Text(
            label,
            style: TextStyle(
              fontSize: 14,
              color: color,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            '¥$value',
            style: TextStyle(
              fontSize: 20,
              color: color,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildOrderCard(Map<String, dynamic> order) {
    final status = _getStatusText(order);
    final statusColor = _getStatusColor(status);
    final orderNumber = order['order_number'] as String? ?? '-';
    final addressName = order['address_name'] as String? ?? '-';
    final orderDate = _formatDate(order['order_date']?.toString());
    final totalCommission = _formatMoney(order['total_commission'] ?? 0);
    final orderAmount = _formatMoney(order['order_amount'] ?? 0);
    
    // 获取订单ID（可能是 order_id 或 id）
    final orderId = order['order_id'] as int? ?? 
                    order['id'] as int? ?? 
                    (order['order_id'] is num ? (order['order_id'] as num).toInt() : null) ??
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

  String _getStatusText(Map<String, dynamic> order) {
    if (order['is_accounted_cancelled'] == true) {
      return '取消计入';
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
        return const Color(0xFFFFA940);
      case '取消计入':
        return const Color(0xFFFF5A5F);
      case '无效订单':
        return Colors.grey;
      default:
        return const Color(0xFF8C92A4);
    }
  }
}
