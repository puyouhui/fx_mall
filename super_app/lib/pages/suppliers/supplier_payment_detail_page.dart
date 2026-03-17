import 'package:flutter/material.dart';
import 'package:super_app/api/suppliers_api.dart';
import 'package:super_app/models/supplier_payment.dart';

class SupplierPaymentDetailPage extends StatefulWidget {
  final int supplierId;
  final String supplierName;

  const SupplierPaymentDetailPage({
    super.key,
    required this.supplierId,
    required this.supplierName,
  });

  @override
  State<SupplierPaymentDetailPage> createState() =>
      _SupplierPaymentDetailPageState();
}

class _SupplierPaymentDetailPageState extends State<SupplierPaymentDetailPage>
    with SingleTickerProviderStateMixin {
  SupplierPaymentDetail? _detail;
  bool _isLoading = false;
  int _currentPage = 1;
  final int _pageSize = 20;
  String? _timeRange;
  String? _status;

  // 视图模式：按订单 / 按天
  int _tabIndex = 0; // 0: 按订单, 1: 按天

  // 按订单视图：选择模式相关
  bool _isSelectionMode = false;
  Set<int> _selectedOrderItemIds = {};

  // 按天视图相关
  List<SupplierDailyStat> _dailyStats = [];
  String _dailyCenterDate =
      DateTime.now().toIso8601String().substring(0, 10); // YYYY-MM-DD
  String _dailySelectedDate =
      DateTime.now().toIso8601String().substring(0, 10); // YYYY-MM-DD
  SupplierDailyDetail? _dailyDetail;

  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(
      length: 2,
      vsync: this,
      initialIndex: _tabIndex,
    );
    _tabController.addListener(() {
      if (!_tabController.indexIsChanging) {
        setState(() {
          _tabIndex = _tabController.index;
          // 切到按天时退出选择模式
          if (_tabIndex == 1) {
            _isSelectionMode = false;
            _selectedOrderItemIds.clear();
          }
        });
      }
    });
    _loadDetail();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _loadDetail() async {
    setState(() => _isLoading = true);

    try {
      final response = await SuppliersApi.getPaymentDetail(
        widget.supplierId,
        timeRange: _timeRange,
        status: _status,
        pageNum: _currentPage,
        pageSize: _pageSize,
      );

      if (!mounted) return;

      if (response.isSuccess && response.data != null) {
        setState(() {
          _detail = response.data;
          // 退出选择模式
          _isSelectionMode = false;
          _selectedOrderItemIds.clear();
        });
        // 首次进入时，初始化按天视图
        await _loadDailyStats();
        await _loadDailyDetail(_dailySelectedDate);
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(response.message)),
          );
        }
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('加载失败: ${e.toString()}')),
      );
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  // ---------- 按天视图逻辑 ----------

  /// 获取当前中心日期附近7天的日期列表
  List<String> get _dailyDates {
    final center = DateTime.parse(_dailyCenterDate);
    final List<String> dates = [];
    for (int offset = -3; offset <= 3; offset++) {
      final d = center.add(Duration(days: offset));
      dates.add(_fmtDate(d));
    }
    return dates;
  }

  String _fmtDate(DateTime d) {
    final y = d.year.toString().padLeft(4, '0');
    final m = d.month.toString().padLeft(2, '0');
    final day = d.day.toString().padLeft(2, '0');
    return '$y-$m-$day';
  }

  Future<void> _loadDailyStats() async {
    try {
      final center = DateTime.parse(_dailyCenterDate);
      final start = center.add(const Duration(days: -3));
      final end = center.add(const Duration(days: 3));
      final res = await SuppliersApi.getDailyStats(
        widget.supplierId,
        startDate: _fmtDate(start),
        endDate: _fmtDate(end),
      );
      if (!mounted) return;
      if (res.isSuccess && res.data != null) {
        setState(() {
          _dailyStats = res.data!;
        });
      }
    } catch (_) {}
  }

  SupplierDailyStat _getDailyStat(String date) {
    return _dailyStats.firstWhere(
      (e) => e.date == date,
      orElse: () => SupplierDailyStat(
        date: date,
        totalAmount: 0,
        paidAmount: 0,
        pendingAmount: 0,
      ),
    );
  }

  String _getDailyStatusText(String date) {
    final s = _getDailyStat(date);
    return s.statusText;
  }

  Color _getDailyStatusColor(String date) {
    final s = _getDailyStat(date);
    if (s.totalAmount == 0) return const Color(0xFF9E9E9E);
    if (s.pendingAmount <= 0 && s.paidAmount > 0) {
      return const Color(0xFF20CB6B); // 已付完：绿
    }
    if (s.paidAmount <= 0 && s.pendingAmount > 0) {
      return const Color(0xFFFF6B6B); // 待付款：红
    }
    return const Color(0xFFFFA726); // 部分已付：橙
  }

  Future<void> _loadDailyDetail(String date) async {
    try {
      final res = await SuppliersApi.getDailyDetail(widget.supplierId, date);
      if (!mounted) return;
      if (res.isSuccess && res.data != null) {
        setState(() {
          _dailyDetail = res.data;
          _dailySelectedDate = date;
        });
      }
    } catch (_) {}
  }

  void _shiftDailyRange(int offsetDays) async {
    final center = DateTime.parse(_dailyCenterDate).add(
      Duration(days: offsetDays),
    );
    setState(() {
      _dailyCenterDate = _fmtDate(center);
    });
    await _loadDailyStats();
  }

  void _toggleSelectionMode() {
    setState(() {
      _isSelectionMode = !_isSelectionMode;
      if (!_isSelectionMode) {
        _selectedOrderItemIds.clear();
      }
    });
  }

  void _toggleItemSelection(int orderItemId, bool isPaid) {
    if (!_isSelectionMode) return;
    // 只能选择待付款的订单
    if (isPaid) return;

    setState(() {
      if (_selectedOrderItemIds.contains(orderItemId)) {
        _selectedOrderItemIds.remove(orderItemId);
      } else {
        _selectedOrderItemIds.add(orderItemId);
      }
    });
  }

  Future<void> _markAsPaid() async {
    if (_selectedOrderItemIds.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请先选择要标记为已付款的订单')),
      );
      return;
    }

    // 收集选中的订单项信息
    List<Map<String, dynamic>> selectedItems = [];
    double totalAmount = 0.0;

    for (final order in _detail!.orders) {
      for (final item in order.items) {
        if (_selectedOrderItemIds.contains(item.orderItemId) && !item.isPaid) {
          selectedItems.add({
            'order_id': order.orderId,
            'order_item_id': item.orderItemId,
            'product_id': item.productId,
            'product_name': item.productName,
            'spec_name': item.specName,
            'quantity': item.quantity,
            'cost_price': item.costPrice,
            'subtotal': item.subtotal,
          });
          totalAmount += item.subtotal;
        }
      }
    }

    if (selectedItems.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('没有可标记的订单项')),
      );
      return;
    }

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('确认标记'),
        content: Text(
          '确定要将选中的 ${selectedItems.length} 个订单项标记为已付款吗？\n'
          '总金额：¥${totalAmount.toStringAsFixed(2)}',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () => Navigator.of(context).pop(true),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF20CB6B),
              foregroundColor: Colors.white,
            ),
            child: const Text('确认'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    setState(() => _isLoading = true);

    try {
      final response = await SuppliersApi.createSupplierPayment(
        widget.supplierId,
        selectedItems,
        totalAmount,
      );

      if (!mounted) return;

      if (response.isSuccess) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(response.message),
            backgroundColor: Colors.green,
          ),
        );
        // 重新加载数据
        await _loadDetail();
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(response.message),
            backgroundColor: Colors.red,
          ),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('标记失败: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('${widget.supplierName} - 付款详情'),
        backgroundColor: const Color(0xFF20CB6B),
        foregroundColor: Colors.white,
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(44),
          child: Row(
            children: [
              Expanded(
                child: TabBar(
                  labelColor: Colors.white,
                  unselectedLabelColor: Colors.white70,
                  indicatorColor: Colors.white,
                  indicatorSize: TabBarIndicatorSize.label,
                  tabs: const [
                    Tab(text: '按订单'),
                    Tab(text: '按天'),
                  ],
                  controller: _tabController,
                ),
              ),
              if (_tabIndex == 0 &&
                  _detail != null &&
                  _detail!.orders.isNotEmpty)
                IconButton(
                  icon: Icon(_isSelectionMode ? Icons.close : Icons.checklist),
                  onPressed: _toggleSelectionMode,
                  tooltip: _isSelectionMode ? '取消选择' : '批量选择',
                ),
            ],
          ),
        ),
      ),
      body: _isLoading && _detail == null
          ? const Center(child: CircularProgressIndicator())
          : (_tabIndex == 0 && _detail == null)
              ? const Center(child: Text('暂无数据'))
              : _tabIndex == 0
                  ? _buildOrderModeBody()
                  : _buildDailyModeBody(),
    );
  }

  Widget _buildOrderModeBody() {
    return Stack(
      children: [
        RefreshIndicator(
          onRefresh: _loadDetail,
          child: ListView(
            padding: EdgeInsets.fromLTRB(
              16,
              16,
              16,
              _isSelectionMode && _selectedOrderItemIds.isNotEmpty ? 80 : 16,
            ),
            children: [
              // 统计信息卡片
              _buildStatsCard(),
              const SizedBox(height: 16),
              // 订单列表
              ..._detail!.orders.map((order) => _buildOrderCard(order)),
            ],
          ),
        ),
        // 底部按钮
        if (_isSelectionMode && _selectedOrderItemIds.isNotEmpty)
          Positioned(
            bottom: 0,
            left: 0,
            right: 0,
            child: Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white,
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.1),
                    blurRadius: 8,
                    offset: const Offset(0, -2),
                  ),
                ],
              ),
              child: SafeArea(
                child: SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _markAsPaid,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF20CB6B),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    child: Text(
                      '标记为已付款 (${_selectedOrderItemIds.length})',
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }

  Widget _buildDailyModeBody() {
    final detail = _dailyDetail;
    return RefreshIndicator(
      onRefresh: () async {
        await _loadDailyStats();
        await _loadDailyDetail(_dailySelectedDate);
      },
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _buildDailyHeader(),
          const SizedBox(height: 12),
          if (detail != null) _buildDailyDetailCard(detail) else _buildDailyEmpty(),
        ],
      ),
    );
  }

  Widget _buildStatsCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: [
          Column(
            children: [
              const Text(
                '总金额',
                style: TextStyle(
                  fontSize: 12,
                  color: Color(0xFF8C92A4),
                ),
              ),
              const SizedBox(height: 4),
              Text(
                '¥${_detail!.totalAmount.toStringAsFixed(2)}',
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF20253A),
                ),
              ),
            ],
          ),
          Container(
            width: 1,
            height: 40,
            color: Colors.grey.shade300,
          ),
          Column(
            children: [
              const Text(
                '订单数量',
                style: TextStyle(
                  fontSize: 12,
                  color: Color(0xFF8C92A4),
                ),
              ),
              const SizedBox(height: 4),
              Text(
                '${_detail!.orderCount}',
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF20253A),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildOrderCard(SupplierPaymentOrder order) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: ExpansionTile(
        title: Row(
          children: [
            Expanded(
              child: Text(
                order.addressName ?? '地址未填写',
                style: const TextStyle(
                  fontWeight: FontWeight.w600,
                  fontSize: 15,
                ),
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
        subtitle: Padding(
          padding: const EdgeInsets.only(top: 4),
          child: Text(
            _formatDateTime(order.createdAt),
            style: const TextStyle(
              fontSize: 12,
              color: Color(0xFF8C92A4),
            ),
          ),
        ),
        children: [
          ...order.items.map((item) => _buildOrderItem(item, order.orderId)),
        ],
      ),
    );
  }

  Widget _buildOrderItem(SupplierPaymentItem item, int orderId) {
    final isSelected = _selectedOrderItemIds.contains(item.orderItemId);
    final canSelect = _isSelectionMode && !item.isPaid;

    return InkWell(
      onTap: canSelect
          ? () => _toggleItemSelection(item.orderItemId, item.isPaid)
          : item.isPaid
              ? null
              : () {
                  // 点击产品进入订单详情
                  Navigator.of(context).pushNamed(
                    '/order_detail',
                    arguments: orderId,
                  );
                },
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: canSelect
              ? (isSelected
                  ? const Color(0xFF20CB6B).withOpacity(0.1)
                  : Colors.grey.shade50)
              : Colors.grey.shade50,
          border: Border(
            top: BorderSide(color: Colors.grey.shade200),
            left: canSelect && isSelected
                ? const BorderSide(color: Color(0xFF20CB6B), width: 3)
                : BorderSide.none,
          ),
        ),
        child: Row(
          children: [
            if (canSelect) ...[
              Checkbox(
                value: isSelected,
                onChanged: (value) =>
                    _toggleItemSelection(item.orderItemId, item.isPaid),
                activeColor: const Color(0xFF20CB6B),
              ),
              const SizedBox(width: 8),
            ],
            Expanded(
              flex: 3,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    item.productName,
                    style: const TextStyle(
                      fontWeight: FontWeight.w500,
                      fontSize: 14,
                    ),
                  ),
                  if (item.specName.isNotEmpty) ...[
                    const SizedBox(height: 2),
                    Text(
                      '规格: ${item.specName}',
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFF8C92A4),
                      ),
                    ),
                  ],
                ],
              ),
            ),
            Expanded(
              flex: 2,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text(
                    '数量: ${item.quantity}',
                    style: const TextStyle(fontSize: 13),
                  ),
                  Text(
                    '成本: ¥${item.costPrice.toStringAsFixed(2)}',
                    style: const TextStyle(
                      fontSize: 12,
                      color: Color(0xFF8C92A4),
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              flex: 2,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text(
                    '小计: ¥${item.subtotal.toStringAsFixed(2)}',
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: item.isPaid
                          ? const Color(0xFF20CB6B).withOpacity(0.1)
                          : const Color(0xFFFF6B6B).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      item.isPaid ? '已付款' : '待付款',
                      style: TextStyle(
                        fontSize: 11,
                        color: item.isPaid
                            ? const Color(0xFF20CB6B)
                            : const Color(0xFFFF6B6B),
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _formatDateTime(DateTime dateTime) {
    return '${dateTime.year}-${dateTime.month.toString().padLeft(2, '0')}-${dateTime.day.toString().padLeft(2, '0')} '
        '${dateTime.hour.toString().padLeft(2, '0')}:${dateTime.minute.toString().padLeft(2, '0')}';
  }

  Widget _buildDailyHeader() {
    final dates = _dailyDates;
    return Row(
      children: [
        IconButton(
          icon: const Icon(Icons.chevron_left),
          onPressed: () => _shiftDailyRange(-7),
        ),
        Expanded(
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: dates.map((d) {
                final stat = _getDailyStat(d);
                final isActive = d == _dailySelectedDate;
                final color = _getDailyStatusColor(d);
                return GestureDetector(
                  onTap: () => _loadDailyDetail(d),
                  child: Container(
                    width: 110,
                    margin: const EdgeInsets.symmetric(horizontal: 4),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 6,
                    ),
                    decoration: BoxDecoration(
                      color: isActive
                          ? const Color(0xFFE3F2FD)
                          : Colors.white,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: isActive ? color : Colors.grey.shade300,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.03),
                          blurRadius: 4,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          d.substring(5), // MM-DD
                          style: const TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          '应付 ¥${stat.totalAmount.toStringAsFixed(2)}',
                          style: const TextStyle(
                            fontSize: 11,
                            color: Color(0xFF20CB6B),
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          _getDailyStatusText(d),
                          style: TextStyle(
                            fontSize: 11,
                            color: color,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
                );
              }).toList(),
            ),
          ),
        ),
        IconButton(
          icon: const Icon(Icons.chevron_right),
          onPressed: () => _shiftDailyRange(7),
        ),
      ],
    );
  }

  Widget _buildDailyDetailCard(SupplierDailyDetail detail) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.white,
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.04),
                blurRadius: 8,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _buildDailyStatItem('应付', detail.totalAmount, const Color(0xFF20CB6B)),
              _buildDailyStatItem('已付', detail.paidAmount, const Color(0xFF2196F3)),
              _buildDailyStatItem('未付', detail.pendingAmount, const Color(0xFFFF6B6B)),
            ],
          ),
        ),
        const SizedBox(height: 16),
        const Text(
          '待付款明细',
          style: TextStyle(
            fontSize: 15,
            fontWeight: FontWeight.w600,
          ),
        ),
        const SizedBox(height: 8),
        if (detail.pendingItems.isEmpty)
          const Text(
            '暂无待付款数据',
            style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
          )
        else
          ...detail.pendingItems.map(_buildDailyItemTile),
        const SizedBox(height: 16),
        const Text(
          '已付款明细',
          style: TextStyle(
            fontSize: 15,
            fontWeight: FontWeight.w600,
          ),
        ),
        const SizedBox(height: 8),
        if (detail.paidItems.isEmpty)
          const Text(
            '暂无已付款数据',
            style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
          )
        else
          ...detail.paidItems.map(_buildDailyItemTile),
      ],
    );
  }

  Widget _buildDailyStatItem(String label, double amount, Color color) {
    return Column(
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 12,
            color: Color(0xFF8C92A4),
          ),
        ),
        const SizedBox(height: 4),
        Text(
          '¥${amount.toStringAsFixed(2)}',
          style: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
            color: color,
          ),
        ),
      ],
    );
  }

  Widget _buildDailyItemTile(SupplierDailyItem item) {
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            flex: 3,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  item.productName,
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                if (item.specName.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 2),
                    child: Text(
                      '规格: ${item.specName}',
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFF8C92A4),
                      ),
                    ),
                  ),
                Padding(
                  padding: const EdgeInsets.only(top: 2),
                  child: Text(
                    '订单号: ${item.orderNumber}',
                    style: const TextStyle(
                      fontSize: 11,
                      color: Color(0xFF8C92A4),
                    ),
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            flex: 2,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  '数量: ${item.quantity}',
                  style: const TextStyle(fontSize: 12),
                ),
                Text(
                  '成本: ¥${item.costPrice.toStringAsFixed(2)}',
                  style: const TextStyle(
                    fontSize: 12,
                    color: Color(0xFF8C92A4),
                  ),
                ),
                Text(
                  '小计: ¥${item.subtotal.toStringAsFixed(2)}',
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20CB6B),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDailyEmpty() {
    return const Padding(
      padding: EdgeInsets.symmetric(vertical: 40),
      child: Center(
        child: Text(
          '请选择上方日期查看明细',
          style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
        ),
      ),
    );
  }
}
