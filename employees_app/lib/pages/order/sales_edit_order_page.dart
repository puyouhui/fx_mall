import 'package:flutter/material.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/order/sales_create_order_page.dart';
import 'package:employees_app/pages/order/sales_create_order_page_coupon.dart';

/// 修改订单页面（销售员）
/// 基于创建订单页面，但加载已有订单数据
class SalesEditOrderPage extends StatefulWidget {
  final int orderId;

  const SalesEditOrderPage({super.key, required this.orderId});

  @override
  State<SalesEditOrderPage> createState() => _SalesEditOrderPageState();
}

class _SalesEditOrderPageState extends State<SalesEditOrderPage>
    with WidgetsBindingObserver {
  bool _isLoading = true;
  bool _isSubmitting = false;
  bool _isLocked = false; // 订单是否已锁定
  bool _hasUnlocked = false; // 是否已经解锁（防止重复解锁）

  int? _customerId;
  Map<String, dynamic>? _user;
  List<dynamic> _addresses = [];
  int? _selectedAddressId;

  List<dynamic> _items = []; // 采购单商品
  Map<String, dynamic>? _summary; // 运费汇总
  Map<String, dynamic>? _riderDeliveryFeePreview; // 配送员配送费预览
  List<dynamic>?
  _purchaseListBackup; // 用户原来的采购单备份（从SyncOrderItemsToPurchaseList获取）

  List<dynamic> _coupons = [];
  Map<String, dynamic>? _selectedCoupon;

  final TextEditingController _remarkController = TextEditingController();

  String _outOfStockStrategy = 'contact_me';
  bool _trustReceipt = false;
  bool _hidePrice = false;
  bool _requirePhoneContact = true;
  bool _isUrgent = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _loadOrderData();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _remarkController.dispose();
    // 注意：不在dispose中自动解锁，避免应用生命周期变化时误解锁
    // 只在用户主动退出（WillPopScope）或取消修改时解锁
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    super.didChangeAppLifecycleState(state);
    // 注意：不要因为应用生命周期变化而解锁订单
    // 只在用户主动退出（WillPopScope）或取消修改时解锁
  }

  /// 解锁订单
  Future<void> _unlockOrder() async {
    if (_hasUnlocked) return;
    _hasUnlocked = true;
    try {
      await Request.post('/employee/sales/orders/${widget.orderId}/unlock');
    } catch (e) {
      // 忽略解锁错误
      print('解锁订单失败: $e');
    }
  }

  /// 加载订单数据
  Future<void> _loadOrderData() async {
    setState(() {
      _isLoading = true;
    });

    // 1. 获取订单详情
    final orderResp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/orders/${widget.orderId}',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (!orderResp.isSuccess || orderResp.data == null) {
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            orderResp.message.isNotEmpty ? orderResp.message : '获取订单信息失败',
          ),
        ),
      );
      Navigator.of(context).pop(false);
      return;
    }

    final orderData = orderResp.data!;
    final order = orderData['order'] as Map<String, dynamic>?;
    final user = orderData['user'] as Map<String, dynamic>?;
    final address = orderData['address'] as Map<String, dynamic>?;

    if (order == null || user == null) {
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('订单数据不完整')));
      // 如果订单数据不完整，不解锁订单（订单可能已经被锁定，不应该因为数据获取失败而解锁）
      Navigator.of(context).pop(false);
      return;
    }

    // 检查订单是否已锁定
    final isLocked = (order['is_locked'] as bool?) ?? false;
    if (!isLocked) {
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('订单未锁定，请先锁定订单')));
      // 返回 null 表示订单已解锁，不需要再次解锁
      // 在 _handleEditOrder 中，我们需要区分 null（订单已解锁）和 false（用户取消）
      Navigator.of(context).pop(null);
      return;
    }

    // 设置锁定状态
    setState(() {
      _isLocked = true;
    });

    _customerId = user['id'] as int?;
    if (_customerId == null) {
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('客户信息缺失')));
      // 客户信息缺失，不解锁订单，直接返回（订单保持锁定状态）
      Navigator.of(context).pop(false);
      return;
    }

    // 2. 将订单商品同步到采购单（确保修改订单时能正确显示商品）
    // 注意：这个API会检查订单是否已锁定，如果未锁定会返回错误
    final syncResp = await Request.post<Map<String, dynamic>>(
      '/employee/sales/orders/${widget.orderId}/sync-to-purchase-list',
      parser: (data) => data as Map<String, dynamic>,
    );

    // 如果同步失败，检查失败原因
    if (!syncResp.isSuccess) {
      setState(() {
        _isLoading = false;
        _isLocked = false; // 同步失败，订单可能未锁定
      });

      // 检查失败原因并显示错误信息
      final errorMessage = syncResp.message.isNotEmpty
          ? syncResp.message
          : '同步订单商品失败';

      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(errorMessage)));

      // 注意：这里不解锁订单，因为：
      // 1. 如果是因为"订单未锁定"失败，订单本来就没锁定，不需要解锁
      // 2. 如果是因为其他原因失败（比如订单已被其他员工锁定），订单可能已经被其他操作解锁了
      // 3. 我们只更新前端状态，不解锁后端订单，避免误操作
      Navigator.of(context).pop(false);
      return;
    }

    // 3. 获取客户详情（含地址列表）
    final detailResp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customers/$_customerId',
      parser: (data) => data as Map<String, dynamic>,
    );

    // 4. 获取客户采购单（同步后应该包含订单商品）
    final purchaseResp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customers/$_customerId/purchase-list',
      parser: (data) => data as Map<String, dynamic>,
    );

    // 5. 获取客户优惠券列表
    final couponsResp = await Request.get<List<dynamic>>(
      '/employee/sales/customers/$_customerId/coupons',
      parser: (data) => data as List<dynamic>,
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
      // 如果同步失败，订单可能未锁定，需要先解锁再返回
      if (!syncResp.isSuccess) {
        // 同步失败可能是因为订单未锁定，尝试解锁（如果已锁定）
        try {
          await Request.post('/employee/sales/orders/${widget.orderId}/unlock');
        } catch (e) {
          // 忽略解锁错误
        }
      }
      Navigator.of(context).pop(false);
      return;
    }

    final detail = detailResp.data!;
    // 如果同步成功，使用同步后的采购单；否则使用原始采购单
    final purchase = syncResp.isSuccess && syncResp.data != null
        ? syncResp.data!
        : (purchaseResp.isSuccess && purchaseResp.data != null
              ? purchaseResp.data!
              : <String, dynamic>{});
    final coupons = couponsResp.isSuccess && couponsResp.data != null
        ? couponsResp.data!
        : <dynamic>[];

    final addrList = (detail['addresses'] as List<dynamic>? ?? []);
    final purchaseItems = (purchase['items'] as List<dynamic>? ?? []);
    final purchaseSummary = purchase['summary'] as Map<String, dynamic>?;
    // 保存备份数据（用户进入修改订单页面时的原始采购单，不包含销售员后续的操作）
    // 注意：备份数据只在 syncResp 成功时才有，如果同步失败，备份数据可能为空
    final backup = syncResp.isSuccess && syncResp.data != null
        ? (syncResp.data!['backup'] as List<dynamic>? ?? [])
        : <dynamic>[]; // 如果同步失败，备份为空数组

    // 设置选中的地址（使用订单中的地址）
    int? selectedAddressId;
    if (address != null && address['id'] != null) {
      selectedAddressId = address['id'] as int?;
    } else if (addrList.isNotEmpty) {
      // 如果没有订单地址，使用默认地址
      for (final a in addrList) {
        final m = a as Map<String, dynamic>;
        final isDefault = (m['is_default'] as bool?) ?? false;
        if (isDefault && m['id'] != null) {
          selectedAddressId = m['id'] as int;
          break;
        }
      }
      selectedAddressId ??=
          (addrList.first as Map<String, dynamic>)['id'] as int?;
    }

    // 设置订单数据
    _remarkController.text = order['remark'] as String? ?? '';
    _outOfStockStrategy =
        order['out_of_stock_strategy'] as String? ?? 'contact_me';
    _trustReceipt = (order['trust_receipt'] as bool?) ?? false;
    _hidePrice = (order['hide_price'] as bool?) ?? false;
    _requirePhoneContact = (order['require_phone_contact'] as bool?) ?? true;
    _isUrgent = (order['is_urgent'] as bool?) ?? false;

    // 查找订单使用的优惠券
    final couponDiscount =
        (order['coupon_discount'] as num?)?.toDouble() ?? 0.0;
    Map<String, dynamic>? selectedCoupon;
    if (couponDiscount > 0) {
      // 尝试从优惠券列表中找到匹配的优惠券
      for (final coupon in coupons) {
        final couponData = coupon as Map<String, dynamic>;
        final couponInfo = couponData['coupon'] as Map<String, dynamic>?;
        if (couponInfo != null) {
          final discountValue =
              (couponInfo['discount_value'] as num?)?.toDouble() ?? 0.0;
          final type = couponInfo['type'] as String? ?? '';
          if (type == 'delivery_fee' && couponDiscount > 0) {
            // 可能是免配送费券
            selectedCoupon = couponData;
            break;
          } else if (type == 'amount' && discountValue == couponDiscount) {
            // 金额券
            selectedCoupon = couponData;
            break;
          }
        }
      }
    }

    setState(() {
      _user = user;
      _addresses = addrList;
      _selectedAddressId = selectedAddressId;
      _items = purchaseItems;
      _summary = purchaseSummary;
      _coupons = coupons;
      _selectedCoupon = selectedCoupon;
      _purchaseListBackup = backup; // 保存备份数据（用户进入修改订单页面时的原始采购单）
      _isLoading = false;
      _isLocked = true; // 订单已锁定
    });

    // 加载配送员配送费预览
    _loadRiderDeliveryFeePreview();
  }

  /// 加载配送员配送费预览
  Future<void> _loadRiderDeliveryFeePreview() async {
    if (_customerId == null || _selectedAddressId == null || _items.isEmpty) {
      return;
    }

    try {
      final resp = await Request.get<Map<String, dynamic>>(
        '/employee/sales/customers/$_customerId/rider-delivery-fee-preview?address_id=$_selectedAddressId&is_urgent=$_isUrgent',
        parser: (data) => data as Map<String, dynamic>,
      );

      if (!mounted) return;

      if (resp.isSuccess && resp.data != null) {
        setState(() {
          _riderDeliveryFeePreview = resp.data;
        });
      }
    } catch (e) {
      // 忽略错误
      print('加载配送员配送费预览失败: $e');
    }
  }

  /// 提交修改订单
  Future<void> _submitOrder() async {
    if (_customerId == null || _user == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('客户信息加载失败，无法修改订单')));
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
      'address_id': _selectedAddressId,
      'item_ids': <int>[], // 使用全部采购单商品
      'remark': _remarkController.text.trim(),
      'out_of_stock_strategy': _outOfStockStrategy,
      'trust_receipt': _trustReceipt,
      'hide_price': _hidePrice,
      'require_phone_contact': _requirePhoneContact,
      'is_urgent': _isUrgent,
      if (_selectedCoupon != null) 'coupon_id': _getCouponId(_selectedCoupon!),
      if (_purchaseListBackup != null)
        'purchase_list_backup': _purchaseListBackup, // 传入备份数据
    };

    final resp = await Request.put<Map<String, dynamic>>(
      '/employee/sales/orders/${widget.orderId}',
      body: body,
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    setState(() {
      _isSubmitting = false;
    });

    if (resp.isSuccess) {
      // 标记已解锁（后端会在修改成功后自动解锁）
      _hasUnlocked = true;

      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('修改订单成功')));

      // 返回true表示修改成功
      Navigator.of(context).pop(true);
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '修改订单失败'),
        ),
      );
    }
  }

  /// 获取优惠券ID
  int? _getCouponId(Map<String, dynamic> coupon) {
    if (coupon['id'] != null) {
      return coupon['id'] as int?;
    }
    return null;
  }

  @override
  Widget build(BuildContext context) {
    final titleName = _user?['name'] as String? ?? '修改订单';

    return WillPopScope(
      onWillPop: () async {
        // 如果订单已锁定且未解锁，先解锁
        // 注意：只有在用户主动返回时才解锁
        if (_isLocked && !_hasUnlocked) {
          await _unlockOrder();
        }
        return true;
      },
      child: Scaffold(
        extendBody: true,
        appBar: AppBar(
          title: Text('修改订单 - $titleName'),
          backgroundColor: const Color(0xFF20CB6B),
          foregroundColor: Colors.white,
        ),
        body: _isLoading
            ? const Center(child: CircularProgressIndicator())
            : _buildContent(),
      ),
    );
  }

  Widget _buildContent() {
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
          Expanded(
            child: SafeArea(
              bottom: false,
              child: RefreshIndicator(
                onRefresh: _loadOrderData,
                child: ListView(
                  padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
                  children: [
                    _buildCustomerInfoCard(),
                    const SizedBox(height: 12),
                    _buildAddressCard(),
                    const SizedBox(height: 12),
                    _buildItemsCard(),
                    const SizedBox(height: 12),
                    _buildCouponCard(),
                    if (_customerId != null) ...[
                      const SizedBox(height: 12),
                      _buildSummaryCard(),
                      const SizedBox(height: 12),
                      _buildOrderOptionsCard(),
                      const SizedBox(height: 12),
                      _buildRemarkCard(),
                      const SizedBox(height: 12),
                      _buildRiderDeliveryFeeCard(),
                    ],
                  ],
                ),
              ),
            ),
          ),
          _buildBottomBar(),
        ],
      ),
    );
  }

  // 以下方法需要从创建订单页面复制，这里先创建占位方法
  // 实际使用时需要从 sales_create_order_page.dart 复制完整的实现

  Widget _buildCustomerInfoCard() {
    final user = _user ?? {};
    final hasCustomer = _customerId != null && _user != null;
    final name = hasCustomer ? (user['name'] as String?) ?? '未填写名称' : '客户信息';
    final phone = hasCustomer ? (user['phone'] as String?) ?? '' : '';
    final userCode = hasCustomer ? (user['user_code'] as String?) ?? '' : '';

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
            '客户信息',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 12),
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
    );
  }

  /// 选择/切换收货地址
  Future<void> _selectAddress() async {
    if (_customerId == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('请先选择客户')));
      return;
    }

    final result = await Navigator.of(context).push<Map<String, dynamic>>(
      MaterialPageRoute(
        builder: (_) => AddressSelectionPage(
          customerId: _customerId!,
          addresses: _addresses,
          selectedAddressId: _selectedAddressId,
        ),
      ),
    );

    if (!mounted || result == null) return;

    final selectedId = result['id'] as int?;
    if (selectedId != null) {
      setState(() {
        _selectedAddressId = selectedId;
      });
      // 更新配送员配送费预览
      _loadRiderDeliveryFeePreview();
    }
  }

  Widget _buildAddressCard() {
    // 获取当前选中的地址
    Map<String, dynamic>? selectedAddress;
    if (_selectedAddressId != null) {
      for (final raw in _addresses) {
        final addr = raw as Map<String, dynamic>;
        final addrId = addr['id'] as int?;
        if (addrId != null && addrId == _selectedAddressId) {
          selectedAddress = addr;
          break;
        }
      }
    }

    return InkWell(
      onTap: _customerId != null ? _selectAddress : null,
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
                  '收货地址',
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
                  )
                else
                  TextButton(
                    onPressed: _selectAddress,
                    style: TextButton.styleFrom(
                      padding: const EdgeInsets.symmetric(horizontal: 8),
                      minimumSize: const Size(0, 32),
                      tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    ),
                    child: const Text(
                      '选择地址',
                      style: TextStyle(fontSize: 14, color: Color(0xFF4C8DF6)),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            if (_customerId == null)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFF5F6FA),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Row(
                  children: [
                    Icon(
                      Icons.info_outline,
                      size: 16,
                      color: Color(0xFF8C92A4),
                    ),
                    SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        '请选择客户后，再选择收货地址',
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ),
                  ],
                ),
              )
            else if (_addresses.isEmpty)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFF5F6FA),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Row(
                  children: [
                    Icon(
                      Icons.info_outline,
                      size: 16,
                      color: Color(0xFF8C92A4),
                    ),
                    SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        '该客户暂无地址，请先在"新客资料"中为客户添加地址',
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ),
                  ],
                ),
              )
            else if (selectedAddress == null)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFF4E6),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: const Color(0xFFFFE0B2), width: 1),
                ),
                child: const Row(
                  children: [
                    Icon(
                      Icons.location_off,
                      size: 16,
                      color: Color(0xFFFF9800),
                    ),
                    SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        '请选择收货地址',
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(0xFFFF9800),
                        ),
                      ),
                    ),
                    Icon(
                      Icons.chevron_right,
                      size: 20,
                      color: Color(0xFF8C92A4),
                    ),
                  ],
                ),
              )
            else ...[
              // 显示选中的地址
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(
                    Icons.location_on,
                    size: 20,
                    color: Color(0xFF20CB6B),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Text(
                                selectedAddress['name'] as String? ?? '收货地址',
                                style: const TextStyle(
                                  fontSize: 15,
                                  fontWeight: FontWeight.w600,
                                  color: Color(0xFF20253A),
                                ),
                              ),
                            ),
                            if ((selectedAddress['is_default'] as bool?) ??
                                false)
                              Container(
                                margin: const EdgeInsets.only(left: 6),
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 6,
                                  vertical: 2,
                                ),
                                decoration: BoxDecoration(
                                  color: const Color(
                                    0xFF20CB6B,
                                  ).withOpacity(0.08),
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
                        const SizedBox(height: 4),
                        if ((selectedAddress['address'] as String?)
                                ?.isNotEmpty ??
                            false)
                          Text(
                            selectedAddress['address'] as String? ?? '',
                            style: const TextStyle(
                              fontSize: 13,
                              color: Color(0xFF40475C),
                            ),
                          ),
                        if (((selectedAddress['contact'] as String?)
                                    ?.isNotEmpty ??
                                false) ||
                            ((selectedAddress['phone'] as String?)
                                    ?.isNotEmpty ??
                                false)) ...[
                          const SizedBox(height: 2),
                          Text(
                            '${selectedAddress['contact'] as String? ?? ''}  ${selectedAddress['phone'] as String? ?? ''}',
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  const Icon(
                    Icons.chevron_right,
                    size: 20,
                    color: Color(0xFF8C92A4),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }

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

  /// 更新商品数量
  Future<void> _updateItemQuantity(int itemId, int newQuantity) async {
    if (newQuantity <= 0) {
      await _deleteItem(itemId);
      return;
    }

    final resp = await Request.put<Map<String, dynamic>>(
      '/employee/sales/customers/$_customerId/purchase-list/$itemId',
      body: {'quantity': newQuantity},
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      final purchase = resp.data!;
      setState(() {
        _items = (purchase['items'] as List<dynamic>? ?? []);
        _summary = purchase['summary'] as Map<String, dynamic>?;
      });
      // 刷新配送员配送费预览
      _loadRiderDeliveryFeePreview();
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '更新数量失败'),
        ),
      );
    }
  }

  /// 删除商品
  Future<void> _deleteItem(int itemId) async {
    final resp = await Request.delete(
      '/employee/sales/customers/$_customerId/purchase-list/$itemId',
    );

    if (!mounted) return;

    if (resp.isSuccess) {
      // 重新加载采购单
      final purchaseResp = await Request.get<Map<String, dynamic>>(
        '/employee/sales/customers/$_customerId/purchase-list',
        parser: (data) => data as Map<String, dynamic>,
      );

      if (!mounted) return;

      if (purchaseResp.isSuccess && purchaseResp.data != null) {
        final purchase = purchaseResp.data!;
        setState(() {
          _items = (purchase['items'] as List<dynamic>? ?? []);
          _summary = purchase['summary'] as Map<String, dynamic>?;
        });
        // 刷新配送员配送费预览
        _loadRiderDeliveryFeePreview();
      }
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '删除失败'),
        ),
      );
    }
  }

  /// 打开添加商品页面
  Future<void> _openAddProductPage() async {
    if (_customerId == null) {
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

    if (!mounted) return;

    // 无论是否有返回数据，都重新加载采购单数据，确保数据是最新的
    // 这样可以避免数据不同步的问题
    final purchaseResp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customers/$_customerId/purchase-list',
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (purchaseResp.isSuccess && purchaseResp.data != null) {
      final purchase = purchaseResp.data!;
      final purchaseItems = (purchase['items'] as List<dynamic>? ?? []);
      final purchaseSummary = purchase['summary'] as Map<String, dynamic>?;

      setState(() {
        _items = purchaseItems;
        _summary = purchaseSummary;
      });
      // 刷新配送员配送费预览
      _loadRiderDeliveryFeePreview();

      // 如果添加商品成功返回了数据，显示成功提示
      if (result != null && result['items'] != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('商品已添加到采购单'),
            backgroundColor: Color(0xFF20CB6B),
          ),
        );
      }
    } else {
      // 如果获取采购单失败，尝试使用返回的数据（如果有）
      if (result != null) {
        final items = result['items'] as List<dynamic>?;
        final summary = result['summary'] as Map<String, dynamic>?;
        if (items != null) {
          setState(() {
            _items = items;
            _summary = summary;
          });
          _loadRiderDeliveryFeePreview();
        }
      }
    }
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
                child: Stack(
                  children: [
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // 左侧图片 + 右下角删除按钮
                        Stack(
                          alignment: Alignment.bottomRight,
                          children: [
                            Container(
                              width: 90,
                              height: 90,
                              decoration: BoxDecoration(
                                color: const Color(0xFFF5F6FA),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              clipBehavior: Clip.antiAlias,
                              child: image.isNotEmpty
                                  ? Image.network(
                                      image,
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
                            if (itemId != null)
                              Positioned(
                                right: 2,
                                bottom: 2,
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
                          child: Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  mainAxisSize: MainAxisSize.min,
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
                                    const SizedBox(height: 4),
                                    if (spec.isNotEmpty) ...[
                                      Text(
                                        spec,
                                        style: const TextStyle(
                                          fontSize: 12,
                                          fontWeight: FontWeight.w500,
                                          color: Color(0xFF8C92A4),
                                        ),
                                      ),
                                      const SizedBox(height: 4),
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
                              Column(
                                crossAxisAlignment: CrossAxisAlignment.end,
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Text(
                                    '¥${subtotal.toStringAsFixed(2)}',
                                    style: const TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.w700,
                                      color: Color(0xFF20CB6B),
                                    ),
                                  ),
                                  if (itemId == null) ...[
                                    const SizedBox(height: 4),
                                    Text(
                                      'x$qty',
                                      style: const TextStyle(
                                        fontSize: 12,
                                        color: Color(0xFF8C92A4),
                                      ),
                                    ),
                                  ],
                                ],
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    // 数量选择模块
                    if (itemId != null)
                      Positioned(
                        right: 0,
                        bottom: 0,
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            _buildRoundQtyButton(
                              icon: Icons.remove,
                              backgroundColor: const Color(0xFFF0F1F5),
                              iconColor: const Color(0xFF8C92A4),
                              onTap: qty > 1
                                  ? () => _updateItemQuantity(itemId, qty - 1)
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
                              backgroundColor: const Color(0xFF20CB6B),
                              iconColor: Colors.white,
                              onTap: () => _updateItemQuantity(itemId, qty + 1),
                            ),
                          ],
                        ),
                      ),
                  ],
                ),
              );
            }).toList(),
          const SizedBox(height: 12),
          Center(
            child: SizedBox(
              width: 180,
              height: 44,
              child: ElevatedButton.icon(
                onPressed: _customerId == null ? null : _openAddProductPage,
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

  /// 选择/切换优惠券
  Future<void> _selectCoupon() async {
    if (_customerId == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('请先选择客户')));
      return;
    }

    if (_coupons.isEmpty) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('该客户暂无可用优惠券')));
      return;
    }

    // 计算当前订单金额（用于筛选可用优惠券）
    final goodsAmount = (_summary?['total_amount'] as num?)?.toDouble() ?? 0.0;

    final result = await Navigator.of(context).push<Map<String, dynamic>>(
      MaterialPageRoute(
        builder: (_) => CouponSelectionPage(
          customerId: _customerId!,
          coupons: _coupons,
          selectedCoupon: _selectedCoupon,
          orderAmount: goodsAmount,
        ),
      ),
    );

    if (!mounted) return;

    if (result != null) {
      if (result['remove'] == true) {
        // 移除优惠券
        setState(() {
          _selectedCoupon = null;
        });
      } else {
        // 选择优惠券
        final coupon = result['coupon'] as Map<String, dynamic>?;
        if (coupon != null) {
          setState(() {
            _selectedCoupon = coupon;
          });
        }
      }
    }
  }

  /// 获取优惠券名称
  String _getCouponName(Map<String, dynamic> coupon) {
    if (coupon['coupon'] != null) {
      final couponData = coupon['coupon'] as Map<String, dynamic>;
      return couponData['name'] as String? ?? '优惠券';
    }
    return coupon['name'] as String? ?? '优惠券';
  }

  /// 计算优惠券实际折扣金额
  double _calculateCouponDiscount(
    Map<String, dynamic>? coupon,
    double goodsAmount,
    double deliveryFee,
  ) {
    if (coupon == null) return 0.0;

    Map<String, dynamic> couponData = coupon;
    if (coupon['coupon'] != null) {
      couponData = coupon['coupon'] as Map<String, dynamic>;
    }

    final couponType = couponData['type'] as String? ?? '';
    final discountValue =
        (couponData['discount_value'] as num?)?.toDouble() ?? 0.0;
    final minAmount = (couponData['min_amount'] as num?)?.toDouble() ?? 0.0;

    // 检查是否满足使用条件
    bool canUse = true;
    if (minAmount > 0 && goodsAmount < minAmount) {
      canUse = false;
    }

    if (!canUse) return 0.0;

    if (couponType == 'delivery_fee') {
      // 免配送费
      return deliveryFee;
    } else if (couponType == 'amount') {
      // 满减券
      double discount = discountValue;
      // 确保折扣不超过商品金额
      if (discount > goodsAmount) {
        discount = goodsAmount;
      }
      return discount;
    }
    return 0.0;
  }

  Widget _buildCouponCard() {
    // 计算可用优惠券数量
    final availableCount = _coupons.where((coupon) {
      final status = (coupon as Map<String, dynamic>)['status'] as String?;
      if (status == 'used' || status == 'expired') {
        return false;
      }

      final expiresAtStr = coupon['expires_at'] as String?;
      if (expiresAtStr != null && expiresAtStr.isNotEmpty) {
        try {
          final expiresAt = DateTime.parse(expiresAtStr);
          final now = DateTime.now();
          if (now.isAfter(expiresAt)) {
            return false;
          }
        } catch (e) {
          // 解析失败，继续检查
        }
      }

      return true;
    }).length;

    return InkWell(
      onTap: _customerId != null && availableCount > 0 ? _selectCoupon : null,
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
                  '优惠券',
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
                  )
                else if (availableCount == 0)
                  const Text(
                    '暂无优惠券',
                    style: TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
                  )
                else
                  TextButton(
                    onPressed: _selectCoupon,
                    style: TextButton.styleFrom(
                      padding: const EdgeInsets.symmetric(horizontal: 8),
                      minimumSize: const Size(0, 32),
                      tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    ),
                    child: const Text(
                      '选择优惠券',
                      style: TextStyle(fontSize: 14, color: Color(0xFF4C8DF6)),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            if (_customerId == null)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFF5F6FA),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Row(
                  children: [
                    Icon(
                      Icons.info_outline,
                      size: 16,
                      color: Color(0xFF8C92A4),
                    ),
                    SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        '请选择客户后，再选择优惠券',
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ),
                  ],
                ),
              )
            else if (availableCount == 0)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFF5F6FA),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Row(
                  children: [
                    Icon(
                      Icons.local_offer_outlined,
                      size: 16,
                      color: Color(0xFF8C92A4),
                    ),
                    SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        '该客户暂无可用优惠券',
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ),
                  ],
                ),
              )
            else if (_selectedCoupon == null)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFF4E6),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: const Color(0xFFFFE0B2), width: 1),
                ),
                child: Row(
                  children: [
                    const Icon(
                      Icons.local_offer,
                      size: 16,
                      color: Color(0xFFFF9800),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        '客户有 $availableCount 张优惠券可用',
                        style: const TextStyle(
                          fontSize: 13,
                          color: Color(0xFFFF9800),
                        ),
                      ),
                    ),
                    const Icon(
                      Icons.chevron_right,
                      size: 20,
                      color: Color(0xFF8C92A4),
                    ),
                  ],
                ),
              )
            else ...[
              // 显示选中的优惠券
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(
                    Icons.local_offer,
                    size: 20,
                    color: Color(0xFF20CB6B),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Text(
                                _getCouponName(_selectedCoupon!),
                                style: const TextStyle(
                                  fontSize: 15,
                                  fontWeight: FontWeight.w600,
                                  color: Color(0xFF20253A),
                                ),
                              ),
                            ),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: const Color(
                                  0xFF20CB6B,
                                ).withOpacity(0.08),
                                borderRadius: BorderRadius.circular(10),
                              ),
                              child: Text(
                                '-¥${_calculateCouponDiscount(_selectedCoupon, (_summary?['total_amount'] as num?)?.toDouble() ?? 0.0, (_summary?['delivery_fee'] as num?)?.toDouble() ?? 0.0).toStringAsFixed(2)}',
                                style: const TextStyle(
                                  fontSize: 12,
                                  fontWeight: FontWeight.w600,
                                  color: Color(0xFF20CB6B),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  InkWell(
                    onTap: () {
                      setState(() {
                        _selectedCoupon = null;
                      });
                    },
                    child: const Icon(
                      Icons.close,
                      size: 18,
                      color: Color(0xFF8C92A4),
                    ),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }

  String _formatMoney(num? value) {
    final v = (value ?? 0).toDouble();
    return v.toStringAsFixed(2);
  }

  Widget _buildSummaryCard() {
    final summary = _summary ?? {};
    final goodsAmount = _formatMoney(summary['total_amount'] as num?);
    final deliveryFee = _formatMoney(summary['delivery_fee'] as num?);
    final freeThreshold = _formatMoney(
      summary['free_shipping_threshold'] as num?,
    );
    final isFree = (summary['is_free_shipping'] as bool?) ?? false;
    final totalQuantity = summary['total_quantity'] as int? ?? 0;
    final totalAmount = (summary['total_amount'] as num?)?.toDouble() ?? 0.0;
    final totalDeliveryFee =
        (summary['delivery_fee'] as num?)?.toDouble() ?? 0.0;

    // 使用统一的方法计算优惠券折扣
    final couponDiscount = _calculateCouponDiscount(
      _selectedCoupon,
      totalAmount,
      totalDeliveryFee,
    );

    final finalTotal = totalAmount + totalDeliveryFee - couponDiscount;

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
          const SizedBox(height: 12),
          if (_customerId == null)
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: const Color(0xFFF5F6FA),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Row(
                children: [
                  Icon(Icons.info_outline, size: 16, color: Color(0xFF8C92A4)),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '请选择客户并添加商品后，将自动计算金额与运费',
                      style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
                    ),
                  ),
                ],
              ),
            )
          else ...[
            // 商品金额行
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  '商品金额',
                  style: TextStyle(fontSize: 14, color: Color(0xFF40475C)),
                ),
                Row(
                  children: [
                    Text(
                      '¥$goodsAmount',
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF20253A),
                      ),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      '（共 $totalQuantity 件）',
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFF8C92A4),
                      ),
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 10),
            // 运费信息行
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  '配送费用',
                  style: TextStyle(fontSize: 14, color: Color(0xFF40475C)),
                ),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text(
                      isFree ? '¥0.00' : '¥$deliveryFee',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: isFree
                            ? const Color(0xFF20CB6B)
                            : const Color(0xFF20253A),
                      ),
                    ),
                    if (!isFree) ...[
                      const SizedBox(height: 2),
                      Text(
                        '满 ¥$freeThreshold 包邮',
                        style: const TextStyle(
                          fontSize: 11,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ] else ...[
                      const SizedBox(height: 2),
                      const Text(
                        '已满足包邮条件',
                        style: TextStyle(
                          fontSize: 11,
                          color: Color(0xFF20CB6B),
                        ),
                      ),
                    ],
                  ],
                ),
              ],
            ),
            if (couponDiscount > 0) ...[
              const SizedBox(height: 10),
              // 优惠券折扣行
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text(
                    '优惠券折扣',
                    style: TextStyle(fontSize: 14, color: Color(0xFF40475C)),
                  ),
                  Text(
                    '-¥${couponDiscount.toStringAsFixed(2)}',
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF20CB6B),
                    ),
                  ),
                ],
              ),
            ],
            const Divider(height: 20, color: Color(0xFFE5E7F0)),
            // 合计行
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  '应付合计',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                ),
                Text(
                  '¥${finalTotal.toStringAsFixed(2)}',
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w700,
                    color: Color(0xFF20CB6B),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildOrderOptionsCard() {
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
          // 缺货处理策略
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: const Color(0xFF20CB6B).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: const Icon(
                  Icons.inventory_2_outlined,
                  size: 16,
                  color: Color(0xFF20CB6B),
                ),
              ),
              const SizedBox(width: 8),
              const Text(
                '遇到缺货时',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          _buildOutOfStockOption(
            value: 'cancel_item',
            label: '缺货商品不要，其他正常发货',
            icon: Icons.cancel_outlined,
          ),
          const SizedBox(height: 8),
          _buildOutOfStockOption(
            value: 'ship_available',
            label: '有货就发，缺货商品不发',
            icon: Icons.local_shipping_outlined,
          ),
          const SizedBox(height: 8),
          _buildOutOfStockOption(
            value: 'contact_me',
            label: '由客服或配送员联系我确认',
            icon: Icons.phone_outlined,
          ),
          const SizedBox(height: 20),
          const Divider(height: 1, color: Color(0xFFE5E7F0)),
          const SizedBox(height: 16),
          // 其他选项
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: const Color(0xFF4C8DF6).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: const Icon(
                  Icons.settings_outlined,
                  size: 16,
                  color: Color(0xFF4C8DF6),
                ),
              ),
              const SizedBox(width: 8),
              const Text(
                '其他选项',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          _buildSwitchOption(
            title: '信任签收',
            description: '配送电话联系不上时，允许放门口或指定位置',
            value: _trustReceipt,
            onChanged: (value) {
              setState(() {
                _trustReceipt = value;
              });
            },
          ),
          const SizedBox(height: 12),
          _buildSwitchOption(
            title: '隐藏价格',
            description: '选择后，小票中将不显示商品价格',
            value: _hidePrice,
            onChanged: (value) {
              setState(() {
                _hidePrice = value;
              });
            },
          ),
          const SizedBox(height: 12),
          _buildSwitchOption(
            title: '配送时电话联系',
            description: '建议保持电话畅通，方便配送员联系',
            value: _requirePhoneContact,
            onChanged: (value) {
              setState(() {
                _requirePhoneContact = value;
              });
            },
          ),
          const SizedBox(height: 12),
          _buildSwitchOption(
            title: '加急订单',
            description: '选择后，将产生加急费用',
            value: _isUrgent,
            onChanged: (value) {
              setState(() {
                _isUrgent = value;
              });
              // 更新配送员配送费预览
              _loadRiderDeliveryFeePreview();
            },
          ),
        ],
      ),
    );
  }

  Widget _buildOutOfStockOption({
    required String value,
    required String label,
    required IconData icon,
  }) {
    final isSelected = _outOfStockStrategy == value;
    return InkWell(
      onTap: () {
        setState(() {
          _outOfStockStrategy = value;
        });
      },
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isSelected
              ? const Color(0xFF20CB6B).withOpacity(0.08)
              : const Color(0xFFF5F6FA),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: isSelected
                ? const Color(0xFF20CB6B)
                : const Color(0xFFE5E7F0),
            width: isSelected ? 2 : 1,
          ),
        ),
        child: Row(
          children: [
            Container(
              width: 24,
              height: 24,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                border: Border.all(
                  color: isSelected
                      ? const Color(0xFF20CB6B)
                      : const Color(0xFFE5E7F0),
                  width: 2,
                ),
                color: isSelected
                    ? const Color(0xFF20CB6B)
                    : Colors.transparent,
              ),
              child: isSelected
                  ? const Icon(Icons.check, size: 16, color: Colors.white)
                  : null,
            ),
            const SizedBox(width: 12),
            Icon(
              icon,
              size: 18,
              color: isSelected
                  ? const Color(0xFF20CB6B)
                  : const Color(0xFF8C92A4),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                label,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
                  color: isSelected
                      ? const Color(0xFF20253A)
                      : const Color(0xFF40475C),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSwitchOption({
    required String title,
    required String description,
    required bool value,
    required ValueChanged<bool> onChanged,
  }) {
    return Row(
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const SizedBox(height: 4),
              Text(
                description,
                style: const TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
              ),
            ],
          ),
        ),
        Switch(
          value: value,
          onChanged: onChanged,
          activeColor: const Color(0xFF20CB6B),
        ),
      ],
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
            '备注',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _remarkController,
            maxLines: 3,
            decoration: const InputDecoration(
              hintText: '请输入备注信息（选填）',
              border: OutlineInputBorder(),
              contentPadding: EdgeInsets.all(12),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRiderDeliveryFeeCard() {
    final preview = _riderDeliveryFeePreview;
    if (preview == null) {
      return const SizedBox.shrink();
    }

    final riderPayableFee =
        (preview['rider_payable_fee'] as num?)?.toDouble() ?? 0.0;

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
      child: Row(
        children: [
          const Text(
            '配送员配送费预估',
            style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
          ),
          const Spacer(),
          Text(
            '¥${riderPayableFee.toStringAsFixed(2)}',
            style: const TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
          ),
        ],
      ),
    );
  }

  /// 取消修改订单
  Future<void> _handleCancelEdit() async {
    // 显示美化的确认对话框
    final shouldCancel = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (context) => Dialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // 信息图标
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  color: const Color(0xFF4C8DF6).withOpacity(0.1),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.info_outline,
                  color: Color(0xFF4C8DF6),
                  size: 32,
                ),
              ),
              const SizedBox(height: 20),
              // 标题
              const Text(
                '确认取消修改',
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const SizedBox(height: 12),
              // 内容
              const Text(
                '确定要取消修改订单吗？\n\n取消后订单将解锁，配送员可以接单。',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  color: Color(0xFF8C92A4),
                  height: 1.5,
                ),
              ),
              const SizedBox(height: 24),
              // 按钮
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: () => Navigator.of(context).pop(false),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        side: const BorderSide(color: Color(0xFFE5E7EB)),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      child: const Text(
                        '继续修改',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton(
                      onPressed: () => Navigator.of(context).pop(true),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFFFF5A5F),
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        elevation: 0,
                      ),
                      child: const Text(
                        '取消修改',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );

    if (shouldCancel == true && mounted) {
      // 解锁订单
      await _unlockOrder();
      // 返回上一页
      Navigator.of(context).pop(false);
    }
  }

  Widget _buildBottomBar() {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 10,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: SafeArea(
        top: false,
        child: Row(
          children: [
            // 取消修改按钮
            Expanded(
              child: OutlinedButton(
                onPressed: _isSubmitting ? null : _handleCancelEdit,
                style: OutlinedButton.styleFrom(
                  foregroundColor: const Color(0xFF8C92A4),
                  side: const BorderSide(color: Color(0xFFE5E7EB)),
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(24),
                  ),
                ),
                child: const Text(
                  '取消修改',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
                ),
              ),
            ),
            const SizedBox(width: 12),
            // 保存修改按钮
            Expanded(
              flex: 2,
              child: ElevatedButton(
                onPressed: _isSubmitting ? null : _submitOrder,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF20CB6B),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(24),
                  ),
                ),
                child: _isSubmitting
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          valueColor: AlwaysStoppedAnimation<Color>(
                            Colors.white,
                          ),
                        ),
                      )
                    : const Text(
                        '保存修改',
                        style: TextStyle(
                          fontSize: 16,
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
}
