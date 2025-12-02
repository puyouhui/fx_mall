import 'package:flutter/material.dart';
import 'package:employees_app/utils/request.dart';

/// 销售开单页面（员工代客下单）
/// 流程：1）展示当前客户信息 2）加载客户采购单和运费汇总 3）选择收货地址 4）填写备注并提交创建订单
class SalesCreateOrderPage extends StatefulWidget {
  final int customerId;
  final String? customerName;

  const SalesCreateOrderPage({
    super.key,
    required this.customerId,
    this.customerName,
  });

  @override
  State<SalesCreateOrderPage> createState() => _SalesCreateOrderPageState();
}

class _SalesCreateOrderPageState extends State<SalesCreateOrderPage> {
  bool _isLoading = true;
  bool _isSubmitting = false;

  Map<String, dynamic>? _user; // 客户信息
  List<dynamic> _addresses = []; // 地址列表
  int? _selectedAddressId;

  List<dynamic> _items = []; // 采购单商品
  Map<String, dynamic>? _summary; // 运费汇总

  final TextEditingController _remarkController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  @override
  void dispose() {
    _remarkController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    setState(() {
      _isLoading = true;
    });

    // 1. 获取客户详情（含地址列表）
    final detailResp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customers/${widget.customerId}',
      parser: (data) => data as Map<String, dynamic>,
    );

    // 2. 获取客户采购单 & 运费汇总
    final purchaseResp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customers/${widget.customerId}/purchase-list',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (!detailResp.isSuccess || detailResp.data == null) {
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            detailResp.message.isNotEmpty
                ? detailResp.message
                : '获取客户信息失败',
          ),
        ),
      );
      return;
    }

    if (!purchaseResp.isSuccess || purchaseResp.data == null) {
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            purchaseResp.message.isNotEmpty
                ? purchaseResp.message
                : '获取客户采购单失败',
          ),
        ),
      );
      return;
    }

    final detail = detailResp.data!;
    final purchase = purchaseResp.data!;

    final user = detail['user'] as Map<String, dynamic>? ?? detail;
    final addrList = (detail['addresses'] as List<dynamic>? ?? []);
    final items = (purchase['items'] as List<dynamic>? ?? []);
    final summary = purchase['summary'] as Map<String, dynamic>?;

    int? selectedAddressId;
    if (addrList.isNotEmpty) {
      // 优先选默认地址
      for (final a in addrList) {
        final m = a as Map<String, dynamic>;
        final isDefault = (m['is_default'] as bool?) ?? false;
        if (isDefault && m['id'] != null) {
          selectedAddressId = m['id'] as int;
          break;
        }
      }
      // 如果没有默认地址，则选第一个
      selectedAddressId ??= (addrList.first as Map<String, dynamic>)['id'] as int?;
    }

    setState(() {
      _user = user;
      _addresses = addrList;
      _selectedAddressId = selectedAddressId;
      _items = items;
      _summary = summary;
      _isLoading = false;
    });
  }

  Future<void> _submitOrder() async {
    if (_user == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('客户信息加载失败，无法下单')),
      );
      return;
    }

    if (_items.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('客户采购单为空，请先在小程序添加商品')),
      );
      return;
    }

    if (_selectedAddressId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请选择收货地址')),
      );
      return;
    }

    setState(() {
      _isSubmitting = true;
    });

    final body = <String, dynamic>{
      'user_id': widget.customerId,
      'address_id': _selectedAddressId,
      // 不传 item_ids，后端会使用该客户采购单中的所有条目
      'item_ids': <int>[],
      'remark': _remarkController.text.trim(),
    };

    final resp = await Request.post<dynamic>(
      '/employee/sales/orders',
      body: body,
    );

    if (!mounted) return;

    setState(() {
      _isSubmitting = false;
    });

    if (resp.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('创建订单成功')),
      );
      Navigator.of(context).pop(true);
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            resp.message.isNotEmpty ? resp.message : '创建订单失败',
          ),
        ),
      );
    }
  }

  String _formatMoney(num? value) {
    final v = (value ?? 0).toDouble();
    return v.toStringAsFixed(2);
  }

  @override
  Widget build(BuildContext context) {
    final titleName = widget.customerName ?? _user?['name'] as String? ?? '';

    return Scaffold(
      appBar: AppBar(
        title: Text(titleName.isNotEmpty ? '销售开单 - $titleName' : '销售开单'),
        centerTitle: true,
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
          ),
        ),
        child: SafeArea(
          child: _isLoading
              ? const Center(
                  child: CircularProgressIndicator(
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                )
              : Column(
                  children: [
                    Expanded(
                      child: RefreshIndicator(
                        onRefresh: _loadData,
                        child: ListView(
                          padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
                          children: [
                            _buildCustomerInfoCard(),
                            const SizedBox(height: 12),
                            _buildAddressCard(),
                            const SizedBox(height: 12),
                            _buildItemsCard(),
                            const SizedBox(height: 12),
                            _buildSummaryCard(),
                            const SizedBox(height: 12),
                            _buildRemarkCard(),
                          ],
                        ),
                      ),
                    ),
                    _buildBottomBar(),
                  ],
                ),
        ),
      ),
    );
  }

  Widget _buildCustomerInfoCard() {
    final user = _user ?? {};
    final name = (user['name'] as String?) ?? '未填写名称';
    final phone = (user['phone'] as String?) ?? '';
    final userCode = (user['user_code'] as String?) ?? '';

    return Container(
      padding: const EdgeInsets.all(16),
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
          const Text(
            '当前客户',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: Text(
                  name,
                  style: const TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (phone.isNotEmpty) ...[
                const SizedBox(width: 8),
                const Icon(Icons.phone, size: 14, color: Color(0xFF8C92A4)),
                const SizedBox(width: 4),
                Text(
                  phone,
                  style: const TextStyle(
                    fontSize: 13,
                    color: Color(0xFF40475C),
                  ),
                ),
              ],
            ],
          ),
          if (userCode.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              '客户编号：$userCode',
              style: const TextStyle(
                fontSize: 12,
                color: Color(0xFF8C92A4),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildAddressCard() {
    return Container(
      padding: const EdgeInsets.all(16),
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
          Row(
            children: const [
              Text(
                '选择收货地址',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          if (_addresses.isEmpty)
            const Text(
              '该客户暂无地址，请先在“新客资料”中为客户添加地址',
              style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
            )
          else
            Column(
              children: _addresses.map((raw) {
                final addr = raw as Map<String, dynamic>;
                final id = addr['id'] as int?;
                final name = (addr['name'] as String?) ?? '收货地址';
                final text = (addr['address'] as String?) ?? '';
                final contact = (addr['contact'] as String?) ?? '';
                final phone = (addr['phone'] as String?) ?? '';
                final isDefault = (addr['is_default'] as bool?) ?? false;

                if (id == null) return const SizedBox.shrink();

                return RadioListTile<int>(
                  value: id,
                  groupValue: _selectedAddressId,
                  onChanged: (value) {
                    setState(() {
                      _selectedAddressId = value;
                    });
                  },
                  contentPadding: EdgeInsets.zero,
                  title: Row(
                    children: [
                      Expanded(
                        child: Text(
                          name,
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                      ),
                      if (isDefault)
                        Container(
                          margin: const EdgeInsets.only(left: 6),
                          padding: const EdgeInsets.symmetric(
                            horizontal: 6,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: const Color(0xFF20CB6B).withOpacity(0.08),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Text(
                            '默认',
                            style: TextStyle(
                              fontSize: 10,
                              color: Color(0xFF20CB6B),
                            ),
                          ),
                        ),
                    ],
                  ),
                  subtitle: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (text.isNotEmpty)
                        Padding(
                          padding: const EdgeInsets.only(top: 2),
                          child: Text(
                            text,
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF40475C),
                            ),
                          ),
                        ),
                      if (contact.isNotEmpty || phone.isNotEmpty)
                        Padding(
                          padding: const EdgeInsets.only(top: 2),
                          child: Text(
                            '$contact  $phone',
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                        ),
                    ],
                  ),
                );
              }).toList(),
            ),
        ],
      ),
    );
  }

  Widget _buildItemsCard() {
    return Container(
      padding: const EdgeInsets.all(16),
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
          Row(
            children: [
              const Text(
                '商品列表',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              TextButton.icon(
                onPressed: () {
                  _openAddProductSheet();
                },
                icon: const Icon(Icons.add, size: 18, color: Color(0xFF4C8DF6)),
                label: const Text(
                  '添加商品',
                  style: TextStyle(
                    fontSize: 13,
                    color: Color(0xFF4C8DF6),
                  ),
                ),
                style: TextButton.styleFrom(
                  padding: EdgeInsets.zero,
                  minimumSize: const Size(0, 32),
                  tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          if (_items.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 4),
              child: Text(
                '当前采购单为空，请点击右上角“添加商品”',
                style: TextStyle(
                  fontSize: 12,
                  color: Color(0xFF8C92A4),
                ),
              ),
            )
          else
            ..._items.map((raw) {
            final item = raw as Map<String, dynamic>;
            final itemId = item['id'] as int?;
            final name = (item['product_name'] as String?) ?? '';
            final spec = (item['spec_name'] as String?) ?? '';
            final qty = (item['quantity'] as int?) ?? 0;
            final snapshot =
                item['spec_snapshot'] as Map<String, dynamic>? ?? {};
            final retailPrice =
                (snapshot['retail_price'] as num?)?.toDouble() ?? 0.0;
            final wholesalePrice =
                (snapshot['wholesale_price'] as num?)?.toDouble() ?? 0.0;
            final cost = (snapshot['cost'] as num?)?.toDouble() ?? 0.0;

            final userType = (_user?['user_type'] as String?) ?? 'retail';
            double unitPrice;
            if (userType == 'wholesale') {
              unitPrice = wholesalePrice > 0 ? wholesalePrice : retailPrice;
            } else {
              unitPrice = retailPrice > 0 ? retailPrice : wholesalePrice;
            }
            if (unitPrice <= 0) {
              unitPrice = cost > 0 ? cost : 0.0;
            }
            final subtotal = unitPrice * qty;
            final image = (item['product_image'] as String?) ?? '';

            return Container(
              margin: const EdgeInsets.only(bottom: 12),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    width: 56,
                    height: 56,
                    decoration: BoxDecoration(
                      color: const Color(0xFFF5F6FA),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    clipBehavior: Clip.antiAlias,
                    child: image.isNotEmpty
                        ? Image.network(
                            image,
                            fit: BoxFit.cover,
                            errorBuilder: (context, error, stackTrace) {
                              return const Icon(
                                Icons.image_not_supported,
                                color: Color(0xFFB0B4C3),
                              );
                            },
                          )
                        : const Icon(Icons.image, color: Color(0xFFB0B4C3)),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          name,
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                        if (spec.isNotEmpty) ...[
                          const SizedBox(height: 2),
                          Text(
                            spec,
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                        ],
                        const SizedBox(height: 6),
                        Row(
                          children: [
                            Text(
                              '¥${unitPrice.toStringAsFixed(2)}',
                              style: const TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                                color: Color(0xFF20CB6B),
                              ),
                            ),
                            const SizedBox(width: 12),
                            // 数量调整区：减号、数量、加号
                            if (itemId != null)
                              Container(
                                decoration: BoxDecoration(
                                  borderRadius: BorderRadius.circular(16),
                                  border: Border.all(
                                      color: const Color(0xFFE5E7F0)),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    IconButton(
                                      iconSize: 18,
                                      padding: EdgeInsets.zero,
                                      constraints: const BoxConstraints(
                                        minWidth: 32,
                                        minHeight: 32,
                                      ),
                                      icon: const Icon(Icons.remove,
                                          color: Color(0xFF8C92A4)),
                                      onPressed: qty > 1
                                          ? () => _updateItemQuantity(
                                                itemId,
                                                qty - 1,
                                              )
                                          : () => _deleteItem(itemId),
                                    ),
                                    Padding(
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 6,
                                      ),
                                      child: Text(
                                        '$qty',
                                        style: const TextStyle(
                                          fontSize: 13,
                                          color: Color(0xFF40475C),
                                        ),
                                      ),
                                    ),
                                    IconButton(
                                      iconSize: 18,
                                      padding: EdgeInsets.zero,
                                      constraints: const BoxConstraints(
                                        minWidth: 32,
                                        minHeight: 32,
                                      ),
                                      icon: const Icon(Icons.add,
                                          color: Color(0xFF20CB6B)),
                                      onPressed: () => _updateItemQuantity(
                                        itemId,
                                        qty + 1,
                                      ),
                                    ),
                                  ],
                                ),
                              )
                            else
                              Text(
                                'x$qty',
                                style: const TextStyle(
                                  fontSize: 12,
                                  color: Color(0xFF8C92A4),
                                ),
                              ),
                            const Spacer(),
                            Text(
                              '¥${subtotal.toStringAsFixed(2)}',
                              style: const TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                                color: Color(0xFF20253A),
                              ),
                            ),
                            if (itemId != null) ...[
                              const SizedBox(width: 8),
                              IconButton(
                                iconSize: 18,
                                padding: EdgeInsets.zero,
                                constraints: const BoxConstraints(
                                  minWidth: 32,
                                  minHeight: 32,
                                ),
                                icon: const Icon(
                                  Icons.delete_outline,
                                  color: Color(0xFFFF5A5F),
                                ),
                                onPressed: () => _deleteItem(itemId),
                              ),
                            ],
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  Widget _buildSummaryCard() {
    final summary = _summary ?? {};
    final goodsAmount = _formatMoney(summary['total_amount'] as num?);
    final deliveryFee = _formatMoney(summary['delivery_fee'] as num?);
    final baseFee = _formatMoney(summary['base_fee'] as num?);
    final freeThreshold =
        _formatMoney(summary['free_shipping_threshold'] as num?);
    final isFree = (summary['is_free_shipping'] as bool?) ?? false;
    final totalQuantity = summary['total_quantity'] as int? ?? 0;

    return Container(
      padding: const EdgeInsets.all(16),
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
          const Text(
            '汇总信息',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 8),
          Text(
            '商品金额：¥$goodsAmount （共 $totalQuantity 件）',
            style: const TextStyle(fontSize: 13, color: Color(0xFF40475C)),
          ),
          const SizedBox(height: 4),
          Text(
            '基础运费：¥$baseFee，满 ¥$freeThreshold 包邮',
            style: const TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
          ),
          const SizedBox(height: 4),
          Text(
            isFree ? '当前已满足包邮条件，配送费用为 ¥0.00' : '当前配送费用：¥$deliveryFee',
            style: TextStyle(
              fontSize: 13,
              color: isFree
                  ? const Color(0xFF20CB6B)
                  : const Color(0xFF40475C),
              fontWeight: isFree ? FontWeight.w600 : FontWeight.normal,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRemarkCard() {
    return Container(
      padding: const EdgeInsets.all(16),
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
          const Text(
            '订单备注（选填）',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 8),
          TextField(
            controller: _remarkController,
            maxLines: 3,
            decoration: const InputDecoration(
              hintText: '例如：帮客户电话确认后再发货、某些商品缺货时电话沟通等',
              border: OutlineInputBorder(
                borderRadius: BorderRadius.all(Radius.circular(12)),
                borderSide: BorderSide(color: Color(0xFFE5E7F0)),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.all(Radius.circular(12)),
                borderSide: BorderSide(color: Color(0xFF20CB6B)),
              ),
              contentPadding: EdgeInsets.symmetric(
                horizontal: 12,
                vertical: 10,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBottomBar() {
    final summary = _summary ?? {};
    final goodsAmount =
        (summary['total_amount'] as num?)?.toDouble() ?? 0.0;
    final deliveryFee =
        (summary['delivery_fee'] as num?)?.toDouble() ?? 0.0;
    final total = goodsAmount + deliveryFee;

    return Container(
      decoration: const BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Color(0x14000000),
            blurRadius: 8,
            offset: Offset(0, -2),
          ),
        ],
      ),
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
      child: SafeArea(
        top: false,
        child: Row(
          children: [
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  '应付合计',
                  style: TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
                ),
                const SizedBox(height: 2),
                Text(
                  '¥${total.toStringAsFixed(2)}',
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF20CB6B),
                  ),
                ),
              ],
            ),
            const Spacer(),
            SizedBox(
              width: 150,
              child: ElevatedButton(
                onPressed: _isSubmitting ? null : _submitOrder,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF20CB6B),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(22),
                  ),
                  elevation: 0,
                ),
                child: Text(
                  _isSubmitting ? '提交中...' : '确认代客下单',
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _updateItemQuantity(int itemId, int quantity) async {
    setState(() {
      _isLoading = true;
    });

    final resp = await Request.put<Map<String, dynamic>>(
      '/employee/sales/customers/${widget.customerId}/purchase-list/$itemId',
      body: {'quantity': quantity},
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      final data = resp.data!;
      setState(() {
        _items = data['items'] as List<dynamic>? ?? [];
        _summary = data['summary'] as Map<String, dynamic>?;
        _isLoading = false;
      });
    } else {
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            resp.message.isNotEmpty ? resp.message : '更新商品数量失败',
          ),
        ),
      );
    }
  }

  Future<void> _deleteItem(int itemId) async {
    setState(() {
      _isLoading = true;
    });

    final resp = await Request.delete<Map<String, dynamic>>(
      '/employee/sales/customers/${widget.customerId}/purchase-list/$itemId',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      final data = resp.data!;
      setState(() {
        _items = data['items'] as List<dynamic>? ?? [];
        _summary = data['summary'] as Map<String, dynamic>?;
        _isLoading = false;
      });
    } else {
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            resp.message.isNotEmpty ? resp.message : '删除商品失败',
          ),
        ),
      );
    }
  }

  /// 打开添加商品的底部弹层：搜索商品 -> 选择规格 -> 加入采购单
  Future<void> _openAddProductSheet() async {
    if (_user == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('客户信息未加载完成')),
      );
      return;
    }

    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) {
        return _AddProductBottomSheet(
          customerId: widget.customerId,
          onUpdated: (items, summary) {
            setState(() {
              _items = items;
              _summary = summary;
            });
          },
        );
      },
    );
  }
}

/// 底部添加商品弹层：搜索商品 + 选择规格 + 数量 + 加入采购单
class _AddProductBottomSheet extends StatefulWidget {
  final int customerId;
  final void Function(List<dynamic> items, Map<String, dynamic>? summary)
      onUpdated;

  const _AddProductBottomSheet({
    required this.customerId,
    required this.onUpdated,
  });

  @override
  State<_AddProductBottomSheet> createState() => _AddProductBottomSheetState();
}

class _AddProductBottomSheetState extends State<_AddProductBottomSheet> {
  final TextEditingController _searchController = TextEditingController();
  bool _isLoading = false;
  List<Map<String, dynamic>> _products = [];

  @override
  void initState() {
    super.initState();
    _loadProducts();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadProducts() async {
    setState(() {
      _isLoading = true;
    });

    final resp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/products',
      queryParams: {
        'pageNum': '1',
        'pageSize': '20',
        if (_searchController.text.trim().isNotEmpty)
          'keyword': _searchController.text.trim(),
      },
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      final list = resp.data!['list'] as List<dynamic>? ?? [];
      setState(() {
        _products = list.cast<Map<String, dynamic>>();
        _isLoading = false;
      });
    } else {
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            resp.message.isNotEmpty ? resp.message : '加载商品失败',
          ),
        ),
      );
    }
  }

  Future<void> _addToPurchaseList(
    Map<String, dynamic> product,
    Map<String, dynamic> spec,
    int quantity,
  ) async {
    final productId = product['id'] as int?;
    final specName = spec['name'] as String? ?? '';

    if (productId == null || specName.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('商品或规格信息不完整')),
      );
      return;
    }

    setState(() {
      _isLoading = true;
    });

    final resp = await Request.post<Map<String, dynamic>>(
      '/employee/sales/customers/${widget.customerId}/purchase-list',
      body: {
        'product_id': productId,
        'spec_name': specName,
        'quantity': quantity,
      },
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    setState(() {
      _isLoading = false;
    });

    if (resp.isSuccess && resp.data != null) {
      final data = resp.data!;
      final items = data['items'] as List<dynamic>? ?? [];
      final summary = data['summary'] as Map<String, dynamic>?;
      widget.onUpdated(items, summary);
      Navigator.of(context).pop(); // 关闭弹层
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('已加入采购单')),
      );
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            resp.message.isNotEmpty ? resp.message : '加入采购单失败',
          ),
        ),
      );
    }
  }

  void _openSpecSelector(Map<String, dynamic> product) {
    final specs = product['specs'] as List<dynamic>? ?? [];
    if (specs.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('该商品暂无规格，无法加入采购单')),
      );
      return;
    }

    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) {
        int quantity = 1;
        Map<String, dynamic>? selectedSpec = specs.first as Map<String, dynamic>;

        return StatefulBuilder(
          builder: (context, setState) {
            return Padding(
              padding: EdgeInsets.only(
                bottom: MediaQuery.of(context).viewInsets.bottom,
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                    child: Row(
                      children: [
                        Expanded(
                          child: Text(
                            product['name'] as String? ?? '',
                            style: const TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                              color: Color(0xFF20253A),
                            ),
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        IconButton(
                          icon: const Icon(Icons.close),
                          onPressed: () => Navigator.of(context).pop(),
                        ),
                      ],
                    ),
                  ),
                  const Divider(height: 1, color: Color(0xFFE5E7F0)),
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          '选择规格',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                        const SizedBox(height: 8),
                        Wrap(
                          spacing: 8,
                          runSpacing: 8,
                          children: specs.map((raw) {
                            final spec = raw as Map<String, dynamic>;
                            final name = spec['name'] as String? ?? '';
                            final desc = spec['description'] as String? ?? '';
                            final isSelected = identical(spec, selectedSpec);
                            return ChoiceChip(
                              label: Text(
                                desc.isNotEmpty ? '$name（$desc）' : name,
                                style: TextStyle(
                                  fontSize: 12,
                                  color: isSelected
                                      ? Colors.white
                                      : const Color(0xFF40475C),
                                ),
                              ),
                              selected: isSelected,
                              selectedColor: const Color(0xFF20CB6B),
                              onSelected: (_) {
                                setState(() {
                                  selectedSpec = spec;
                                });
                              },
                            );
                          }).toList(),
                        ),
                        const SizedBox(height: 16),
                        const Text(
                          '购买数量',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                        const SizedBox(height: 8),
                        Row(
                          children: [
                            Container(
                              decoration: BoxDecoration(
                                borderRadius: BorderRadius.circular(16),
                                border: Border.all(
                                    color: const Color(0xFFE5E7F0)),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  IconButton(
                                    iconSize: 18,
                                    padding: EdgeInsets.zero,
                                    constraints: const BoxConstraints(
                                      minWidth: 32,
                                      minHeight: 32,
                                    ),
                                    icon: const Icon(Icons.remove,
                                        color: Color(0xFF8C92A4)),
                                    onPressed: quantity > 1
                                        ? () {
                                            setState(() {
                                              quantity--;
                                            });
                                          }
                                        : null,
                                  ),
                                  Padding(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                    ),
                                    child: Text(
                                      '$quantity',
                                      style: const TextStyle(
                                        fontSize: 14,
                                        color: Color(0xFF40475C),
                                      ),
                                    ),
                                  ),
                                  IconButton(
                                    iconSize: 18,
                                    padding: EdgeInsets.zero,
                                    constraints: const BoxConstraints(
                                      minWidth: 32,
                                      minHeight: 32,
                                    ),
                                    icon: const Icon(Icons.add,
                                        color: Color(0xFF20CB6B)),
                                    onPressed: () {
                                      setState(() {
                                        quantity++;
                                      });
                                    },
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 16),
                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton(
                            onPressed: selectedSpec == null
                                ? null
                                : () => _addToPurchaseList(
                                      product,
                                      selectedSpec!,
                                      quantity,
                                    ),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: const Color(0xFF20CB6B),
                              foregroundColor: Colors.white,
                              padding:
                                  const EdgeInsets.symmetric(vertical: 12),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(22),
                              ),
                              elevation: 0,
                            ),
                            child: const Text(
                              '加入采购单',
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedPadding(
      duration: const Duration(milliseconds: 200),
      padding: EdgeInsets.only(
        bottom: MediaQuery.of(context).viewInsets.bottom,
      ),
      child: Container(
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
        ),
        child: SafeArea(
          top: false,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                child: Row(
                  children: [
                    const Text(
                      '添加商品',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF20253A),
                      ),
                    ),
                    const Spacer(),
                    IconButton(
                      icon: const Icon(Icons.close),
                      onPressed: () => Navigator.of(context).pop(),
                    ),
                  ],
                ),
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: Row(
                  children: [
                    Expanded(
                      child: Container(
                        height: 40,
                        decoration: BoxDecoration(
                          color: const Color(0xFFF5F6FA),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        padding: const EdgeInsets.symmetric(horizontal: 12),
                        child: TextField(
                          controller: _searchController,
                          decoration: const InputDecoration(
                            hintText: '输入商品名称 / 编码搜索',
                            border: InputBorder.none,
                            icon: Icon(Icons.search,
                                color: Color(0xFF8C92A4), size: 18),
                          ),
                          textInputAction: TextInputAction.search,
                          onSubmitted: (_) => _loadProducts(),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    SizedBox(
                      height: 40,
                      child: ElevatedButton(
                        onPressed: _loadProducts,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFF20CB6B),
                          foregroundColor: Colors.white,
                          padding:
                              const EdgeInsets.symmetric(horizontal: 12),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(20),
                          ),
                          elevation: 0,
                        ),
                        child: const Text(
                          '搜索',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const Divider(height: 1, color: Color(0xFFE5E7F0)),
              SizedBox(
                height: 360,
                child: _isLoading
                    ? const Center(
                        child: CircularProgressIndicator(
                          valueColor: AlwaysStoppedAnimation<Color>(
                            Color(0xFF20CB6B),
                          ),
                        ),
                      )
                    : _products.isEmpty
                        ? const Center(
                            child: Text(
                              '暂无商品数据',
                              style: TextStyle(
                                fontSize: 13,
                                color: Color(0xFF8C92A4),
                              ),
                            ),
                          )
                        : ListView.builder(
                            itemCount: _products.length,
                            itemBuilder: (context, index) {
                              final product = _products[index];
                              final name =
                                  (product['name'] as String?) ?? '';
                              final desc =
                                  (product['description'] as String?) ?? '';
                              final images =
                                  product['images'] as List<dynamic>? ?? [];
                              final image =
                                  images.isNotEmpty ? images[0] as String? : '';
                              return ListTile(
                                leading: Container(
                                  width: 44,
                                  height: 44,
                                  decoration: BoxDecoration(
                                    color: const Color(0xFFF5F6FA),
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  clipBehavior: Clip.antiAlias,
                                  child: (image ?? '').isNotEmpty
                                      ? Image.network(
                                          image!,
                                          fit: BoxFit.cover,
                                          errorBuilder:
                                              (context, error, stackTrace) {
                                            return const Icon(
                                              Icons.image_not_supported,
                                              color: Color(0xFFB0B4C3),
                                            );
                                          },
                                        )
                                      : const Icon(Icons.image,
                                          color: Color(0xFFB0B4C3)),
                                ),
                                title: Text(
                                  name,
                                  style: const TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w600,
                                    color: Color(0xFF20253A),
                                  ),
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                subtitle: desc.isNotEmpty
                                    ? Text(
                                        desc,
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                        style: const TextStyle(
                                          fontSize: 12,
                                          color: Color(0xFF8C92A4),
                                        ),
                                      )
                                    : null,
                                trailing: TextButton(
                                  onPressed: () => _openSpecSelector(product),
                                  child: const Text(
                                    '选择规格',
                                    style: TextStyle(
                                      fontSize: 12,
                                      color: Color(0xFF4C8DF6),
                                    ),
                                  ),
                                ),
                              );
                            },
                          ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}


