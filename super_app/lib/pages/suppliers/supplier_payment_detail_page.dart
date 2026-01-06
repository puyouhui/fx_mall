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

class _SupplierPaymentDetailPageState
    extends State<SupplierPaymentDetailPage> {
  SupplierPaymentDetail? _detail;
  bool _isLoading = false;
  int _currentPage = 1;
  final int _pageSize = 20;
  String? _timeRange;
  String? _status;

  // 选择模式相关
  bool _isSelectionMode = false;
  Set<int> _selectedOrderItemIds = {};

  @override
  void initState() {
    super.initState();
    _loadDetail();
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
        actions: [
          if (_detail != null && _detail!.orders.isNotEmpty)
            IconButton(
              icon: Icon(_isSelectionMode ? Icons.close : Icons.checklist),
              onPressed: _toggleSelectionMode,
              tooltip: _isSelectionMode ? '取消选择' : '批量选择',
            ),
        ],
      ),
      body: _isLoading && _detail == null
          ? const Center(child: CircularProgressIndicator())
          : _detail == null
              ? const Center(child: Text('暂无数据'))
              : Stack(
                  children: [
                    RefreshIndicator(
                      onRefresh: _loadDetail,
                      child: ListView(
                        padding: EdgeInsets.fromLTRB(
                          16,
                          16,
                          16,
                          _isSelectionMode && _selectedOrderItemIds.isNotEmpty
                              ? 80
                              : 16,
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
}
