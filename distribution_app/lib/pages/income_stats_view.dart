import 'package:flutter/material.dart';
import '../api/income_api.dart';

/// 收入统计页面
class IncomeStatsView extends StatefulWidget {
  const IncomeStatsView({super.key});

  @override
  State<IncomeStatsView> createState() => _IncomeStatsViewState();
}

class _IncomeStatsViewState extends State<IncomeStatsView> {
  Map<String, dynamic>? _stats;
  bool _isLoading = true;
  String? _errorMessage;
  int _currentTab = 0; // 0: 全部, 1: 已结算, 2: 未结算

  @override
  void initState() {
    super.initState();
    _loadStats();
  }

  Future<void> _loadStats() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final response = await IncomeApi.getIncomeStats();

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      setState(() {
        _stats = response.data;
        _isLoading = false;
      });
    } else {
      setState(() {
        _errorMessage = response.message.isNotEmpty
            ? response.message
            : '获取收入统计失败';
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('配送收入统计'),
        backgroundColor: const Color(0xFF20CB6B),
        elevation: 0,
        iconTheme: const IconThemeData(color: Colors.white),
        titleTextStyle: const TextStyle(
          color: Colors.white,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [Color(0xFF20CB6B), Color(0xFFEFF7F2)],
            stops: [0.0, 0.3],
          ),
        ),
        child: _isLoading
            ? const Center(child: CircularProgressIndicator())
            : _errorMessage != null
            ? Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(
                      Icons.error_outline,
                      size: 64,
                      color: Colors.red,
                    ),
                    const SizedBox(height: 16),
                    Text(
                      _errorMessage!,
                      style: const TextStyle(fontSize: 16, color: Colors.red),
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: _loadStats,
                      child: const Text('重试'),
                    ),
                  ],
                ),
              )
            : _buildContent(),
      ),
    );
  }

  Widget _buildContent() {
    // 安全地转换数值，支持 int 和 double
    final settledFeeValue = _stats?['settled_fee'];
    final settledFee = settledFeeValue is int
        ? settledFeeValue.toDouble()
        : (settledFeeValue as num?)?.toDouble() ?? 0.0;

    final unsettledFeeValue = _stats?['unsettled_fee'];
    final unsettledFee = unsettledFeeValue is int
        ? unsettledFeeValue.toDouble()
        : (unsettledFeeValue as num?)?.toDouble() ?? 0.0;

    final totalFeeValue = _stats?['total_fee'];
    final totalFee = totalFeeValue is int
        ? totalFeeValue.toDouble()
        : (totalFeeValue as num?)?.toDouble() ?? 0.0;

    final orderCount = _stats?['order_count'] as int? ?? 0;

    return Column(
      children: [
        // 统计卡片
        Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            children: [
              // 总收入卡片
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
                      '总收入',
                      style: TextStyle(
                        fontSize: 16,
                        color: Color(0xFF8C92A4),
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      '¥${totalFee.toStringAsFixed(2)}',
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
                            '已结算',
                            settledFee,
                            const Color(0xFF20CB6B),
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: _buildStatItem(
                            '未结算',
                            unsettledFee,
                            const Color(0xFFFF9500),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF5F7FA),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(
                            Icons.shopping_bag_outlined,
                            size: 18,
                            color: Color(0xFF8C92A4),
                          ),
                          const SizedBox(width: 8),
                          Text(
                            '已完成订单：$orderCount 单',
                            style: const TextStyle(
                              fontSize: 14,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
        // Tab栏
        Container(
          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          padding: const EdgeInsets.all(4),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
          ),
          child: Row(
            children: [
              Expanded(child: _buildTabButton('全部', 0)),
              Expanded(child: _buildTabButton('已结算', 1)),
              Expanded(child: _buildTabButton('未结算', 2)),
            ],
          ),
        ),
        // 订单列表
        Expanded(child: _IncomeDetailsList(currentTab: _currentTab)),
      ],
    );
  }

  Widget _buildStatItem(String label, double value, Color color) {
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
            '¥${value.toStringAsFixed(2)}',
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

  Widget _buildTabButton(String text, int index) {
    final isSelected = _currentTab == index;
    return InkWell(
      onTap: () {
        setState(() {
          _currentTab = index;
        });
      },
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12),
        decoration: BoxDecoration(
          color: isSelected
              ? const Color(0xFF20CB6B).withOpacity(0.1)
              : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.08),
                    blurRadius: 8,
                    offset: const Offset(0, 2),
                  ),
                ]
              : null,
        ),
        child: Text(
          text,
          textAlign: TextAlign.center,
          style: TextStyle(
            fontSize: 15,
            fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
            color: isSelected
                ? const Color(0xFF20CB6B)
                : const Color(0xFF8C92A4),
          ),
        ),
      ),
    );
  }
}

/// 收入明细列表
class _IncomeDetailsList extends StatefulWidget {
  final int currentTab;

  const _IncomeDetailsList({required this.currentTab});

  @override
  State<_IncomeDetailsList> createState() => _IncomeDetailsListState();
}

class _IncomeDetailsListState extends State<_IncomeDetailsList> {
  List<dynamic> _orders = [];
  bool _isLoading = false;
  bool _hasMore = true;
  int _pageNum = 1;
  final int _pageSize = 20;

  @override
  void initState() {
    super.initState();
    _loadOrders();
  }

  @override
  void didUpdateWidget(_IncomeDetailsList oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.currentTab != widget.currentTab) {
      _refreshOrders();
    }
  }

  Future<void> _loadOrders({bool refresh = false}) async {
    if (_isLoading) return;

    if (refresh) {
      _pageNum = 1;
      _hasMore = true;
    }

    if (!_hasMore) return;

    setState(() {
      _isLoading = true;
    });

    String? settled;
    if (widget.currentTab == 1) {
      settled = 'true';
    } else if (widget.currentTab == 2) {
      settled = 'false';
    }

    final response = await IncomeApi.getIncomeDetails(
      pageNum: _pageNum,
      pageSize: _pageSize,
      settled: settled,
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final list = data['list'] as List<dynamic>? ?? [];
      final total = data['total'] as int? ?? 0;

      setState(() {
        if (refresh) {
          _orders = list;
        } else {
          _orders.addAll(list);
        }
        _hasMore = _orders.length < total;
        _pageNum++;
        _isLoading = false;
      });
    } else {
      setState(() {
        _isLoading = false;
      });
    }
  }

  Future<void> _refreshOrders() async {
    await _loadOrders(refresh: true);
  }

  @override
  Widget build(BuildContext context) {
    if (_orders.isEmpty && !_isLoading) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.receipt_long_outlined,
              size: 64,
              color: Color(0xFF8C92A4),
            ),
            const SizedBox(height: 16),
            const Text(
              '暂无订单',
              style: TextStyle(fontSize: 16, color: Color(0xFF8C92A4)),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _refreshOrders,
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _orders.length + (_hasMore ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == _orders.length) {
            if (_hasMore) {
              _loadOrders();
              return const Center(
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: CircularProgressIndicator(),
                ),
              );
            }
            return const SizedBox.shrink();
          }

          final order = _orders[index];
          return _buildOrderItem(order);
        },
      ),
    );
  }

  Widget _buildOrderItem(Map<String, dynamic> order) {
    final orderNumber = order['order_number'] as String? ?? '';
    final addressName = order['address_name'] as String? ?? '';

    // 安全地转换数值，支持 int 和 double
    final riderPayableFeeValue = order['rider_payable_fee'];
    final riderPayableFee = riderPayableFeeValue is int
        ? riderPayableFeeValue.toDouble()
        : (riderPayableFeeValue as num?)?.toDouble() ?? 0.0;

    final isSettled = order['delivery_fee_settled'] as bool? ?? false;
    final settlementDate = order['settlement_date'] as String?;

    return Container(
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
                        addressName.isNotEmpty ? addressName : '收货地址',
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
                    color: isSettled
                        ? const Color(0xFF20CB6B).withOpacity(0.1)
                        : const Color(0xFFFF9500).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    isSettled ? '已结算' : '未结算',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: isSettled
                          ? const Color(0xFF20CB6B)
                          : const Color(0xFFFF9500),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    const Text(
                      '配送费',
                      style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      '+¥${riderPayableFee.toStringAsFixed(2)}',
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w700,
                        color: Color(0xFF20253A),
                      ),
                    ),
                  ],
                ),
                if (isSettled && settlementDate != null)
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      const Text(
                        '结算日期',
                        style: TextStyle(
                          fontSize: 12,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        settlementDate.contains('T')
                            ? settlementDate.split('T')[0]
                            : settlementDate.split(' ')[0],
                        style: const TextStyle(
                          fontSize: 12,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ],
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
