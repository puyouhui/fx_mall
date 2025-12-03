import 'package:flutter/material.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/customer/customer_list_page.dart';

/// 销售开单页面（员工代客下单）
/// 支持：
/// 1）从任意入口打开页面，先选择客户，再下单
/// 2）也可以从客户详情页带着 customerId 进来，默认选中该客户
class SalesCreateOrderPage extends StatefulWidget {
  /// 可选的初始客户 ID，从客户详情页跳转时传入
  final int? customerId;
  final String? customerName;

  const SalesCreateOrderPage({super.key, this.customerId, this.customerName});

  @override
  State<SalesCreateOrderPage> createState() => _SalesCreateOrderPageState();
}

class _SalesCreateOrderPageState extends State<SalesCreateOrderPage> {
  bool _isLoading = true;
  bool _isSubmitting = false;

  int? _customerId; // 当前选中的客户 ID

  Map<String, dynamic>? _user; // 客户信息
  List<dynamic> _addresses = []; // 地址列表
  int? _selectedAddressId;

  List<dynamic> _items = []; // 采购单商品
  Map<String, dynamic>? _summary; // 运费汇总

  final TextEditingController _remarkController = TextEditingController();

  /// 圆形加减按钮
  Widget _buildRoundQtyButton({
    required IconData icon,
    required Color backgroundColor,
    required Color iconColor,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        width: 28,
        height: 28,
        decoration: BoxDecoration(
          color: backgroundColor,
          shape: BoxShape.circle,
        ),
        child: Icon(icon, size: 16, color: iconColor),
      ),
    );
  }

  /// 选择/切换客户
  Future<void> _selectCustomer() async {
    final result = await Navigator.of(context).push<Map<String, dynamic>>(
      MaterialPageRoute(builder: (_) => const CustomerListPage(pickMode: true)),
    );

    if (!mounted || result == null) return;

    final id = result['id'] as int?;
    if (id == null) return;

    setState(() {
      _customerId = id;
      _user = result;
      _addresses = [];
      _selectedAddressId = null;
      _items = [];
      _summary = null;
    });

    await _loadData();
  }

  @override
  void initState() {
    super.initState();
    _customerId = widget.customerId;
    if (_customerId != null) {
      _loadData();
    } else {
      // 没有预选客户时，直接显示页面，让用户先选择客户
      _isLoading = false;
    }
  }

  @override
  void dispose() {
    _remarkController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    if (_customerId == null) {
      // 还未选择客户，不请求数据
      setState(() {
        _isLoading = false;
        _user = null;
        _addresses = [];
        _selectedAddressId = null;
        _items = [];
        _summary = null;
      });
      return;
    }

    setState(() {
      _isLoading = true;
    });

    // 1. 获取客户详情（含地址列表）
    final detailResp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customers/$_customerId',
      parser: (data) => data as Map<String, dynamic>,
    );

    // 2. 获取客户采购单 & 运费汇总
    final purchaseResp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customers/$_customerId/purchase-list',
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
            detailResp.message.isNotEmpty ? detailResp.message : '获取客户信息失败',
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
      selectedAddressId ??=
          (addrList.first as Map<String, dynamic>)['id'] as int?;
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
    if (_customerId == null || _user == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('客户信息加载失败，无法下单')));
      return;
    }

    if (_items.isEmpty) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('客户采购单为空，请先在小程序添加商品')));
      return;
    }

    if (_selectedAddressId == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('请选择收货地址')));
      return;
    }

    setState(() {
      _isSubmitting = true;
    });

    final body = <String, dynamic>{
      'user_id': _customerId,
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
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('创建订单成功')));
      Navigator.of(context).pop(true);
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '创建订单失败'),
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
                            if (_customerId != null) ...[
                              const SizedBox(height: 12),
                              _buildSummaryCard(),
                              const SizedBox(height: 12),
                              _buildRemarkCard(),
                            ],
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
    final hasCustomer = _customerId != null && _user != null;
    final name = hasCustomer ? (user['name'] as String?) ?? '未填写名称' : '请选择客户';
    final phone = hasCustomer ? (user['phone'] as String?) ?? '' : '';
    final userCode = hasCustomer ? (user['user_code'] as String?) ?? '' : '';

    return InkWell(
      onTap: _selectCustomer,
      borderRadius: BorderRadius.circular(16),
      child: Container(
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
                  '当前客户',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                ),
                const Spacer(),
                TextButton(
                  onPressed: _selectCustomer,
                  style: TextButton.styleFrom(
                    padding: const EdgeInsets.symmetric(horizontal: 8),
                    minimumSize: const Size(0, 32),
                    tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  ),
                  child: Text(
                    hasCustomer ? '更换客户' : '选择客户',
                    style: const TextStyle(
                      fontSize: 14,
                      color: Color(0xFF4C8DF6),
                    ),
                  ),
                ),
              ],
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
                style: const TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
              ),
            ],
          ],
        ),
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
            children: [
              const Text(
                '选择收货地址',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              if (_customerId == null)
                const Text(
                  '请先选择客户',
                  style: TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
                ),
            ],
          ),
          const SizedBox(height: 8),
          if (_customerId == null)
            const Text(
              '请选择客户后，再选择收货地址',
              style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
            )
          else if (_addresses.isEmpty)
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
          const Text(
            '商品列表',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 10),
          if (_customerId == null)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 4),
              child: Text(
                '请选择客户后，再为客户添加商品',
                style: TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
              ),
            )
          else if (_items.isEmpty)
            const SizedBox.shrink()
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
                padding: const EdgeInsets.only(bottom: 12),
                decoration: const BoxDecoration(
                  border: Border(
                    bottom: BorderSide(color: Color(0xFFE5E7F0), width: 0.6),
                  ),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // 左侧图片 + 右下角删除按钮（红色圆形）
                    Stack(
                      alignment: Alignment.bottomRight,
                      children: [
                        Container(
                          width: 72,
                          height: 72,
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
                              : const Icon(
                                  Icons.image,
                                  color: Color(0xFFB0B4C3),
                                ),
                        ),
                        if (itemId != null)
                          Positioned(
                            right: -2,
                            bottom: -2,
                            child: InkWell(
                              onTap: () => _deleteItem(itemId),
                              borderRadius: BorderRadius.circular(14),
                              child: Container(
                                width: 24,
                                height: 24,
                                decoration: const BoxDecoration(
                                  color: Color(0xFFFF5A5F),
                                  shape: BoxShape.circle,
                                ),
                                child: const Icon(
                                  Icons.delete_outline,
                                  size: 14,
                                  color: Colors.white,
                                ),
                              ),
                            ),
                          ),
                      ],
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: SizedBox(
                        height: 72,
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            // 左侧：名称 + 规格 + 单价（底部与图片对齐）
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
                                  const Spacer(),
                                  if (spec.isNotEmpty) ...[
                                    Text(
                                      spec,
                                      style: const TextStyle(
                                        fontSize: 12,
                                        fontWeight: FontWeight.w500,
                                        color: Color(0xFF8C92A4),
                                      ),
                                    ),
                                  ],
                                  Text(
                                    '¥${unitPrice.toStringAsFixed(2)}',
                                    style: const TextStyle(
                                      fontSize: 13,
                                      fontWeight: FontWeight.w600,
                                      color: Color(0xFF20CB6B),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(width: 8),
                            // 右侧列：上方总价，下方数量模块（整体与总价右对齐，并与图片底部对齐）
                            Column(
                              crossAxisAlignment: CrossAxisAlignment.end,
                              children: [
                                Text(
                                  '¥${subtotal.toStringAsFixed(2)}',
                                  style: const TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.w700,
                                    color: Color(0xFF20253A),
                                  ),
                                ),
                                const Spacer(),
                                if (itemId != null)
                                  Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      _buildRoundQtyButton(
                                        icon: Icons.remove,
                                        backgroundColor: const Color(
                                          0xFFF0F1F5,
                                        ),
                                        iconColor: const Color(0xFF8C92A4),
                                        onTap: qty > 1
                                            ? () => _updateItemQuantity(
                                                itemId,
                                                qty - 1,
                                              )
                                            : () => _deleteItem(itemId),
                                      ),
                                      const SizedBox(width: 6),
                                      SizedBox(
                                        width: 32,
                                        child: Center(
                                          child: Text(
                                            '$qty',
                                            style: const TextStyle(
                                              fontSize: 13,
                                              color: Color(0xFF40475C),
                                            ),
                                          ),
                                        ),
                                      ),
                                      const SizedBox(width: 6),
                                      _buildRoundQtyButton(
                                        icon: Icons.add,
                                        backgroundColor: const Color(
                                          0xFF20CB6B,
                                        ),
                                        iconColor: Colors.white,
                                        onTap: () => _updateItemQuantity(
                                          itemId,
                                          qty + 1,
                                        ),
                                      ),
                                    ],
                                  )
                                else
                                  Text(
                                    'x$qty',
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
                  ],
                ),
              );
            }),
          const SizedBox(height: 12),
          Center(
            child: SizedBox(
              width: 180,
              height: 44,
              child: ElevatedButton.icon(
                onPressed: _customerId == null
                    ? null
                    : () => _openAddProductPage(),
                icon: const Icon(Icons.add, size: 20),
                label: const Text(
                  '添加商品',
                  style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
                ),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.white,
                  foregroundColor: const Color(0xFF4C8DF6),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(22),
                    side: const BorderSide(color: Color(0xFF4C8DF6)),
                  ),
                  elevation: 0,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryCard() {
    final summary = _summary ?? {};
    final goodsAmount = _formatMoney(summary['total_amount'] as num?);
    final deliveryFee = _formatMoney(summary['delivery_fee'] as num?);
    final baseFee = _formatMoney(summary['base_fee'] as num?);
    final freeThreshold = _formatMoney(
      summary['free_shipping_threshold'] as num?,
    );
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
          if (_customerId == null)
            const Text(
              '请选择客户并添加商品后，将自动计算金额与运费',
              style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
            )
          else ...[
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
    final goodsAmount = (summary['total_amount'] as num?)?.toDouble() ?? 0.0;
    final deliveryFee = (summary['delivery_fee'] as num?)?.toDouble() ?? 0.0;
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
                onPressed: _customerId == null || _isSubmitting
                    ? null
                    : _submitOrder,
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
    if (_customerId == null) return;

    final resp = await Request.put<Map<String, dynamic>>(
      '/employee/sales/customers/$_customerId/purchase-list/$itemId',
      body: {'quantity': quantity},
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      final data = resp.data!;
      setState(() {
        _items = data['items'] as List<dynamic>? ?? [];
        _summary = data['summary'] as Map<String, dynamic>?;
      });
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '更新商品数量失败'),
        ),
      );
    }
  }

  Future<void> _deleteItem(int itemId) async {
    if (_customerId == null) return;

    final resp = await Request.delete<Map<String, dynamic>>(
      '/employee/sales/customers/$_customerId/purchase-list/$itemId',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      final data = resp.data!;
      setState(() {
        _items = data['items'] as List<dynamic>? ?? [];
        _summary = data['summary'] as Map<String, dynamic>?;
      });
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '删除商品失败'),
        ),
      );
    }
  }

  /// 打开添加商品页面：搜索商品 -> 选择规格 -> 加入采购单
  Future<void> _openAddProductPage() async {
    if (_user == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('客户信息未加载完成')));
      return;
    }

    final result = await Navigator.of(context).push<Map<String, dynamic>>(
      MaterialPageRoute(
        builder: (_) => AddProductPage(customerId: _customerId!),
      ),
    );

    if (!mounted || result == null) return;

    final items = result['items'] as List<dynamic>?;
    final summary = result['summary'] as Map<String, dynamic>?;
    if (items != null) {
      setState(() {
        _items = items;
        _summary = summary;
      });
    }
  }
}

/// 添加商品页面：搜索商品 + 选择规格 + 数量 + 加入采购单
class AddProductPage extends StatefulWidget {
  final int customerId;

  const AddProductPage({required this.customerId});

  @override
  State<AddProductPage> createState() => _AddProductPageState();
}

class _AddProductPageState extends State<AddProductPage> {
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
          content: Text(resp.message.isNotEmpty ? resp.message : '加载商品失败'),
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
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('商品或规格信息不完整')));
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
      Navigator.of(context).pop<Map<String, dynamic>>({
        'items': items,
        'summary': summary,
      }); // 关闭页面并把最新采购单返回
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('已加入采购单')));
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '加入采购单失败'),
        ),
      );
    }
  }

  void _openSpecSelector(Map<String, dynamic> product) {
    final specs = product['specs'] as List<dynamic>? ?? [];
    if (specs.isEmpty) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('该商品暂无规格，无法加入采购单')));
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
        Map<String, dynamic>? selectedSpec =
            specs.first as Map<String, dynamic>;

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
                                  color: const Color(0xFFE5E7F0),
                                ),
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
                                    icon: const Icon(
                                      Icons.remove,
                                      color: Color(0xFF8C92A4),
                                    ),
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
                                    icon: const Icon(
                                      Icons.add,
                                      color: Color(0xFF20CB6B),
                                    ),
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
                              padding: const EdgeInsets.symmetric(vertical: 12),
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
    return Scaffold(
      appBar: AppBar(
        title: const Text('添加商品'),
        centerTitle: true,
        backgroundColor: const Color(0xFF20CB6B),
        elevation: 0,
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
          child: Column(
            children: [
              // 搜索栏
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                child: Row(
                  children: [
                    Expanded(
                      child: Container(
                        height: 40,
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(20),
                        ),
                        padding: const EdgeInsets.symmetric(horizontal: 12),
                        child: TextField(
                          controller: _searchController,
                          decoration: const InputDecoration(
                            hintText: '输入商品名称 / 编码搜索',
                            border: InputBorder.none,
                            icon: Icon(
                              Icons.search,
                              color: Color(0xFF8C92A4),
                              size: 18,
                            ),
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
                          backgroundColor: Colors.white,
                          foregroundColor: const Color(0xFF20CB6B),
                          padding: const EdgeInsets.symmetric(horizontal: 12),
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
              Expanded(
                child: Container(
                  decoration: const BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.vertical(
                      top: Radius.circular(16),
                    ),
                  ),
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
                            final name = (product['name'] as String?) ?? '';
                            final desc =
                                (product['description'] as String?) ?? '';
                            final images =
                                product['images'] as List<dynamic>? ?? [];
                            final image = images.isNotEmpty
                                ? images[0] as String?
                                : '';
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
                                    : const Icon(
                                        Icons.image,
                                        color: Color(0xFFB0B4C3),
                                      ),
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
              ),
            ],
          ),
        ),
      ),
    );
  }
}
