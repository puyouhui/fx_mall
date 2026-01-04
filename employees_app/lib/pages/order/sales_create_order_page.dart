import 'package:flutter/material.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/customer/customer_list_page.dart';
import 'package:employees_app/pages/order/sales_create_order_page_coupon.dart';
import 'package:employees_app/pages/order/order_detail_page.dart';

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
  Map<String, dynamic>? _riderDeliveryFeePreview; // 配送员配送费预览
  Map<String, dynamic>? _salesCommissionPreview; // 销售分润预览
  List<dynamic>?
  _purchaseListBackup; // 用户原来的采购单备份（从GetSalesCustomerPurchaseList获取）

  // 改价信息：采购单项ID -> {unit_price: 改价后的单价}
  Map<int, Map<String, dynamic>> _priceModifications = {};

  List<dynamic> _coupons = []; // 客户优惠券列表
  Map<String, dynamic>? _selectedCoupon; // 选中的优惠券

  double _urgentFee = 0.0; // 加急费用（从系统设置获取）

  // 防抖相关
  DateTime? _lastSalesCommissionPreviewTime; // 上次调用销售分润预览的时间
  bool _isLoadingSalesCommissionPreview = false; // 是否正在加载销售分润预览

  // 预览卡片显示/隐藏
  bool _showPreviewCards = true; // 是否显示预估配送费和销售分润卡片

  final TextEditingController _remarkController = TextEditingController();

  // 缺货处理策略和选项
  String _outOfStockStrategy = 'contact_me'; // 默认：由客服或配送员联系我确认
  bool _trustReceipt = false; // 信任签收
  bool _hidePrice = false; // 隐藏价格
  bool _requirePhoneContact = true; // 配送时电话联系（默认开启）
  bool _isUrgent = false; // 是否加急

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
        _coupons = [];
        _selectedCoupon = null;
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
      // 传递 is_urgent=true，确保返回正确的加急费用配置
      queryParams: {'is_urgent': 'true'},
      parser: (data) => data as Map<String, dynamic>,
    );

    // 3. 获取客户优惠券列表
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
    final coupons = couponsResp.isSuccess && couponsResp.data != null
        ? couponsResp.data!
        : <dynamic>[];

    final user = detail['user'] as Map<String, dynamic>? ?? detail;
    final addrList = (detail['addresses'] as List<dynamic>? ?? []);
    final items = (purchase['items'] as List<dynamic>? ?? []);
    final summary = purchase['summary'] as Map<String, dynamic>?;
    // 获取加急费用：优先使用接口返回的字段，其次尝试summary中的字段
    double urgentFee = (purchase['urgent_fee'] as num?)?.toDouble() ?? 0.0;
    if (urgentFee <= 0 && summary != null) {
      urgentFee = (summary['urgent_fee'] as num?)?.toDouble() ?? 0.0;
    }
    // 保存备份数据（用户进入开单页面时的原始采购单，不包含销售员后续的操作）
    final backup = (purchase['backup'] as List<dynamic>? ?? []);

    int? selectedAddressId;
    // 如果已经有选中的地址，先检查该地址是否还在地址列表中
    if (_selectedAddressId != null && addrList.isNotEmpty) {
      bool addressExists = false;
      for (final a in addrList) {
        final m = a as Map<String, dynamic>;
        if (m['id'] == _selectedAddressId) {
          addressExists = true;
          selectedAddressId = _selectedAddressId; // 保留用户之前选择的地址
          break;
        }
      }
      // 如果之前选中的地址不存在了，则重新选择
      if (!addressExists) {
        selectedAddressId = null; // 重置，下面会重新选择
      }
    }

    // 如果没有选中的地址（或之前选中的地址已不存在），则选择默认地址
    if (selectedAddressId == null && addrList.isNotEmpty) {
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
      _coupons = coupons;
      _selectedCoupon = null; // 不默认选中优惠券
      _urgentFee = urgentFee; // 设置加急费用
      // 只在第一次加载时保存备份（如果备份为空），后续调用不覆盖备份
      // 这样确保备份是用户进入开单页面时的原始状态，不包含销售员后续的任何操作
      if (_purchaseListBackup == null || _purchaseListBackup!.isEmpty) {
        _purchaseListBackup = backup; // 保存备份数据（用户进入开单页面时的原始采购单）
      }
      _isLoading = false;
    });
    // 加载配送员配送费预览，等待完成后再加载销售分润预览
    // 因为销售分润预览需要配送成本（从配送员配送费预览中获取）
    await _loadRiderDeliveryFeePreview();

    // 如果有改价信息，需要重新计算汇总信息（使用改价后的价格）
    // 因为后端返回的 summary 是基于原价计算的
    if (_priceModifications.isNotEmpty) {
      _recalculateSummary();
    }

    // 加载销售分润预览（确保配送员配送费预览已完成）
    _loadSalesCommissionPreview();
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

    // 构建改价信息列表
    final priceModifications = _priceModifications.entries.map((entry) {
      return {
        'purchase_list_item_id': entry.key,
        'unit_price': entry.value['unit_price'],
      };
    }).toList();

    final body = <String, dynamic>{
      'user_id': _customerId,
      'address_id': _selectedAddressId,
      // 不传 item_ids，后端会使用该客户采购单中的所有条目
      'item_ids': <int>[],
      'remark': _remarkController.text.trim(),
      'out_of_stock_strategy': _outOfStockStrategy,
      'trust_receipt': _trustReceipt,
      'hide_price': _hidePrice,
      'require_phone_contact': _requirePhoneContact,
      'is_urgent': _isUrgent,
      if (_selectedCoupon != null) 'coupon_id': _getCouponId(_selectedCoupon!),
      if (_purchaseListBackup != null)
        'purchase_list_backup': _purchaseListBackup, // 传入备份数据
      if (priceModifications.isNotEmpty)
        'price_modifications': priceModifications,
    };

    final resp = await Request.post<Map<String, dynamic>>(
      '/employee/sales/orders',
      body: body,
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    setState(() {
      _isSubmitting = false;
    });

    if (resp.isSuccess) {
      // 获取订单ID
      int? orderId;
      if (resp.data != null) {
        final order = resp.data!['order'] as Map<String, dynamic>?;
        if (order != null) {
          orderId = order['id'] as int?;
        }
      }

      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('创建订单成功')));

      // 如果有订单ID，跳转到订单详情页面
      if (orderId != null) {
        // 先关闭当前页面
        Navigator.of(context).pop();
        // 然后跳转到订单详情页面
        Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => OrderDetailPage(orderId: orderId!)),
        );
      } else {
        // 如果没有订单ID，只返回上一页
        Navigator.of(context).pop(true);
      }
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
      extendBody: true, // 让body延伸到系统操作条下方
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
        actions: [
          IconButton(
            icon: Icon(
              _showPreviewCards ? Icons.visibility_off : Icons.visibility,
              color: Colors.white,
            ),
            onPressed: () {
              setState(() {
                _showPreviewCards = !_showPreviewCards;
              });
            },
            tooltip: _showPreviewCards ? '隐藏预览' : '显示预览',
          ),
        ],
      ),
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [Color(0xFF20CB6B), Color(0xFFEFF7F2)],
          ),
        ),
        child: _isLoading
            ? SafeArea(
                child: const Center(
                  child: CircularProgressIndicator(
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                ),
              )
            : Column(
                children: [
                  Expanded(
                    child: SafeArea(
                      bottom: false, // 底部不使用SafeArea，让内容延伸到系统操作条
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
                            _buildCouponCard(),
                            const SizedBox(height: 12),
                            _buildUrgentCard(),
                            if (_customerId != null) ...[
                              const SizedBox(height: 12),
                              _buildSummaryCard(),
                              const SizedBox(height: 12),
                              _buildOrderOptionsCard(),
                              const SizedBox(height: 12),
                              _buildRemarkCard(),
                              if (_showPreviewCards) ...[
                                const SizedBox(height: 12),
                                _buildRiderDeliveryFeeCard(),
                                const SizedBox(height: 12),
                                _buildSalesCommissionCard(),
                              ],
                            ],
                          ],
                        ),
                      ),
                    ),
                  ),
                  _buildBottomBar(),
                ],
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
        // 更新销售分润预览
        _loadSalesCommissionPreview();
      } else {
        // 选择优惠券
        final coupon = result['coupon'] as Map<String, dynamic>?;
        if (coupon != null) {
          setState(() {
            _selectedCoupon = coupon;
          });
          // 更新销售分润预览
          _loadSalesCommissionPreview();
        }
      }
    }
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
      // 更新配送员配送费预览，等待完成后再更新销售分润预览
      await _loadRiderDeliveryFeePreview();
      // 更新销售分润预览（确保配送员配送费预览已完成）
      _loadSalesCommissionPreview();
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

  /// 构建加急订单卡片
  Widget _buildUrgentCard() {
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
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Text(
                      '加急配送',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF20253A),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      _urgentFee > 0
                          ? '¥${_urgentFee.toStringAsFixed(2)}'
                          : '¥0.00',
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w700,
                        color: Color(0xFFFF6B6B),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 4),
                const Text(
                  '将优先为您配送',
                  style: TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
                ),
              ],
            ),
          ),
          const SizedBox(width: 12),
          Switch(
            value: _isUrgent,
            onChanged: (value) async {
              setState(() {
                _isUrgent = value;
              });
              // 更新配送员配送费预览，等待完成后再更新销售分润预览
              await _loadRiderDeliveryFeePreview();
              // 更新销售分润预览（确保配送员配送费预览已完成）
              _loadSalesCommissionPreview();
            },
            activeColor: const Color(0xFFFF6B6B),
          ),
        ],
      ),
    );
  }

  Widget _buildCouponCard() {
    // 计算可用优惠券数量（只统计未使用且未过期的）
    final availableCount = _coupons.where((coupon) {
      final status = (coupon as Map<String, dynamic>)['status'] as String?;
      if (status == 'used' || status == 'expired') {
        return false;
      }

      // 检查用户优惠券的有效期（expires_at）
      final expiresAtStr = coupon['expires_at'] as String?;
      if (expiresAtStr != null && expiresAtStr.isNotEmpty) {
        try {
          final expiresAt = DateTime.parse(expiresAtStr);
          final now = DateTime.now();
          if (now.isAfter(expiresAt)) {
            return false; // 已过期
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
                            // 优惠金额显示
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: const Color(0xFFFFF4E6).withOpacity(0.8),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: Text(
                                _getCouponDiscountText(_selectedCoupon!),
                                style: const TextStyle(
                                  fontSize: 13,
                                  fontWeight: FontWeight.w700,
                                  color: Color(0xFFFF9800),
                                ),
                              ),
                            ),
                            const SizedBox(width: 8),
                            InkWell(
                              onTap: () {
                                setState(() {
                                  _selectedCoupon = null;
                                });
                                // 更新销售分润预览
                                _loadSalesCommissionPreview();
                              },
                              child: const Icon(
                                Icons.close,
                                size: 18,
                                color: Color(0xFF8C92A4),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 6),
                        Text(
                          _getCouponDescription(_selectedCoupon!),
                          style: const TextStyle(
                            fontSize: 13,
                            color: Color(0xFF40475C),
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          _formatValidPeriod(_selectedCoupon!),
                          style: const TextStyle(
                            fontSize: 11,
                            color: Color(0xFF8C92A4),
                          ),
                        ),
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

  String _getCouponDescription(Map<String, dynamic> coupon) {
    // 如果传入的是 userCoupon，需要从 coupon 字段获取
    Map<String, dynamic> couponData = coupon;
    if (coupon['coupon'] != null) {
      couponData = coupon['coupon'] as Map<String, dynamic>;
    }

    final type = couponData['type'] as String? ?? '';
    final discountValue =
        (couponData['discount_value'] as num?)?.toDouble() ?? 0.0;
    final minAmount = (couponData['min_amount'] as num?)?.toDouble() ?? 0.0;

    if (type == 'delivery_fee') {
      return '免配送费';
    } else if (type == 'amount') {
      if (minAmount > 0) {
        return '满¥${minAmount.toStringAsFixed(2)}减¥${discountValue.toStringAsFixed(2)}';
      } else {
        return '减¥${discountValue.toStringAsFixed(2)}';
      }
    }
    return '';
  }

  /// 格式化有效期
  String _formatValidPeriod(Map<String, dynamic> coupon) {
    // 如果传入的是 userCoupon，优先使用 expires_at，否则使用 coupon 的 valid_to
    Map<String, dynamic> couponData = coupon;
    if (coupon['coupon'] != null) {
      couponData = coupon['coupon'] as Map<String, dynamic>;
    }

    // 优先使用 userCoupon 的 expires_at
    final expiresAt = coupon['expires_at'] as String?;
    if (expiresAt != null && expiresAt.isNotEmpty) {
      try {
        final expires = DateTime.parse(expiresAt);
        final expiresStr =
            '${expires.year}-${expires.month.toString().padLeft(2, '0')}-${expires.day.toString().padLeft(2, '0')}';
        return '有效期至 $expiresStr';
      } catch (e) {
        return '有效期至 $expiresAt';
      }
    }

    // 否则使用 coupon 的 valid_to
    final validTo = couponData['valid_to'] as String?;
    if (validTo != null && validTo.isNotEmpty) {
      try {
        final to = DateTime.parse(validTo);
        final toStr =
            '${to.year}-${to.month.toString().padLeft(2, '0')}-${to.day.toString().padLeft(2, '0')}';
        return '有效期至 $toStr';
      } catch (e) {
        return '有效期至 $validTo';
      }
    }

    return '';
  }

  /// 获取优惠金额显示
  String _getCouponDiscountText(Map<String, dynamic> coupon) {
    // 如果传入的是 userCoupon，需要从 coupon 字段获取
    Map<String, dynamic> couponData = coupon;
    if (coupon['coupon'] != null) {
      couponData = coupon['coupon'] as Map<String, dynamic>;
    }

    final type = couponData['type'] as String? ?? '';
    final discountValue =
        (couponData['discount_value'] as num?)?.toDouble() ?? 0.0;

    if (type == 'delivery_fee') {
      return '免配送费';
    } else if (type == 'amount') {
      return '减¥${discountValue.toStringAsFixed(2)}';
    }
    return '';
  }

  /// 获取优惠券名称
  String _getCouponName(Map<String, dynamic> coupon) {
    // 如果传入的是 userCoupon，需要从 coupon 字段获取
    if (coupon['coupon'] != null) {
      final couponData = coupon['coupon'] as Map<String, dynamic>;
      return couponData['name'] as String? ?? '优惠券';
    }
    return coupon['name'] as String? ?? '优惠券';
  }

  /// 获取优惠券ID（用于提交订单）
  /// 返回 user_coupon_id（用户优惠券ID），而不是 coupon_id（优惠券本身ID）
  int? _getCouponId(Map<String, dynamic> coupon) {
    // 如果传入的是 userCoupon，使用 userCoupon 的 id（即 user_coupon_id）
    if (coupon['id'] != null) {
      return coupon['id'] as int?;
    }
    return null;
  }

  /// 计算优惠券实际折扣金额
  double _calculateCouponDiscount(
    Map<String, dynamic>? coupon,
    double goodsAmount,
    double deliveryFee,
  ) {
    if (coupon == null) return 0.0;

    // 如果传入的是 userCoupon，需要从 coupon 字段获取
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
              // 计算原始价格
              double originalUnitPrice;
              if (userType == 'wholesale') {
                originalUnitPrice = wholesalePrice > 0
                    ? wholesalePrice
                    : retailPrice;
              } else {
                originalUnitPrice = retailPrice > 0
                    ? retailPrice
                    : wholesalePrice;
              }
              if (originalUnitPrice <= 0) {
                originalUnitPrice = cost > 0 ? cost : 0.0;
              }

              // 检查是否有改价
              double unitPrice = originalUnitPrice;
              bool isPriceModified = false;
              if (itemId != null && _priceModifications.containsKey(itemId)) {
                final mod = _priceModifications[itemId]!;
                final modifiedPrice = (mod['unit_price'] as num?)?.toDouble();
                if (modifiedPrice != null && modifiedPrice >= 0) {
                  unitPrice = modifiedPrice;
                  isPriceModified = unitPrice != originalUnitPrice;
                }
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
                        // 左侧图片 + 右下角删除按钮（红色圆形）
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
                              // 左侧：名称 + 规格 + 单价
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
                                    Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        if (isPriceModified) ...[
                                          Text(
                                            '¥${originalUnitPrice.toStringAsFixed(2)}',
                                            style: TextStyle(
                                              fontSize: 11,
                                              color: Colors.grey[400],
                                              decoration:
                                                  TextDecoration.lineThrough,
                                            ),
                                          ),
                                          const SizedBox(width: 4),
                                        ],
                                        Text(
                                          '¥${unitPrice.toStringAsFixed(2)}',
                                          style: TextStyle(
                                            fontSize: 13,
                                            fontWeight: FontWeight.w600,
                                            color: isPriceModified
                                                ? const Color(0xFFFF5A5F)
                                                : const Color(0xFF20CB6B),
                                          ),
                                        ),
                                      ],
                                    ),
                                    if (isPriceModified && itemId != null) ...[
                                      const SizedBox(height: 2),
                                      InkWell(
                                        onTap: () => _modifyPrice(
                                          itemId,
                                          originalUnitPrice,
                                          unitPrice,
                                          cost,
                                        ),
                                        child: Container(
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 6,
                                            vertical: 2,
                                          ),
                                          decoration: BoxDecoration(
                                            color: const Color(
                                              0xFFFF5A5F,
                                            ).withOpacity(0.1),
                                            borderRadius: BorderRadius.circular(
                                              4,
                                            ),
                                          ),
                                          child: const Text(
                                            '已改价',
                                            style: TextStyle(
                                              fontSize: 10,
                                              color: Color(0xFFFF5A5F),
                                            ),
                                          ),
                                        ),
                                      ),
                                    ] else if (itemId != null) ...[
                                      const SizedBox(height: 2),
                                      InkWell(
                                        onTap: () => _modifyPrice(
                                          itemId,
                                          originalUnitPrice,
                                          unitPrice,
                                          cost,
                                        ),
                                        child: Container(
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 6,
                                            vertical: 2,
                                          ),
                                          decoration: BoxDecoration(
                                            color: const Color(
                                              0xFF20CB6B,
                                            ).withOpacity(0.1),
                                            borderRadius: BorderRadius.circular(
                                              4,
                                            ),
                                          ),
                                          child: const Text(
                                            '改价',
                                            style: TextStyle(
                                              fontSize: 10,
                                              color: Color(0xFF20CB6B),
                                            ),
                                          ),
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                              ),
                              const SizedBox(width: 8),
                              // 右侧：总价
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
                    // 数量选择模块放在右下角
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

    // 计算加急费用
    final urgentFee = _isUrgent ? _urgentFee : 0.0;

    final finalTotal =
        totalAmount + totalDeliveryFee + urgentFee - couponDiscount;

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
            if (_isUrgent) ...[
              const SizedBox(height: 10),
              // 加急配送费行
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        '加急配送费',
                        style: TextStyle(
                          fontSize: 14,
                          color: Color(0xFF40475C),
                        ),
                      ),
                      const SizedBox(height: 2),
                      const Text(
                        '将优先为您配送',
                        style: TextStyle(
                          fontSize: 11,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ],
                  ),
                  Text(
                    '¥${urgentFee.toStringAsFixed(2)}',
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFFFF6B6B),
                    ),
                  ),
                ],
              ),
            ],
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
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: const Color(0xFF20CB6B).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: const Icon(
                  Icons.note_outlined,
                  size: 16,
                  color: Color(0xFF20CB6B),
                ),
              ),
              const SizedBox(width: 8),
              const Text(
                '订单备注',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const SizedBox(width: 6),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: const Color(0xFFF5F6FA),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: const Text(
                  '选填',
                  style: TextStyle(
                    fontSize: 11,
                    color: Color(0xFF8C92A4),
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Container(
            decoration: BoxDecoration(
              color: const Color(0xFFF5F6FA),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: const Color(0xFFE5E7F0), width: 1),
            ),
            child: TextField(
              controller: _remarkController,
              maxLines: 4,
              minLines: 3,
              maxLength: 200,
              buildCounter:
                  (
                    context, {
                    required currentLength,
                    required isFocused,
                    maxLength,
                  }) => null, // 隐藏默认计数器，使用自定义的
              style: const TextStyle(
                fontSize: 14,
                color: Color(0xFF20253A),
                height: 1.5,
              ),
              decoration: const InputDecoration(
                hintText: '请输入订单备注信息，例如：帮客户电话确认后再发货、某些商品缺货时电话沟通等',
                hintStyle: TextStyle(
                  fontSize: 13,
                  color: Color(0xFF8C92A4),
                  height: 1.5,
                ),
                border: InputBorder.none,
                enabledBorder: InputBorder.none,
                focusedBorder: InputBorder.none,
                contentPadding: EdgeInsets.all(12),
              ),
            ),
          ),
          const SizedBox(height: 8),
          Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              ValueListenableBuilder<TextEditingValue>(
                valueListenable: _remarkController,
                builder: (context, value, child) {
                  final length = value.text.length;
                  return Text(
                    '$length/200',
                    style: TextStyle(
                      fontSize: 12,
                      color: length > 200
                          ? const Color(0xFFFF5722)
                          : const Color(0xFF8C92A4),
                    ),
                  );
                },
              ),
            ],
          ),
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
                  fontWeight: FontWeight.w500,
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
                  fontSize: 15,
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
        const SizedBox(width: 12),
        Switch(
          value: value,
          onChanged: onChanged,
          activeColor: const Color(0xFF20CB6B),
        ),
      ],
    );
  }

  Widget _buildBottomBar() {
    final summary = _summary ?? {};
    final goodsAmount = (summary['total_amount'] as num?)?.toDouble() ?? 0.0;
    final deliveryFee = (summary['delivery_fee'] as num?)?.toDouble() ?? 0.0;

    // 使用统一的方法计算优惠券折扣
    final couponDiscount = _calculateCouponDiscount(
      _selectedCoupon,
      goodsAmount,
      deliveryFee,
    );

    // 计算加急费用
    final urgentFee = _isUrgent ? _urgentFee : 0.0;

    final total = goodsAmount + deliveryFee + urgentFee - couponDiscount;
    final bottomPadding = MediaQuery.of(context).padding.bottom;

    return Container(
      // 外层Container：白色背景延伸到系统操作条区域
      color: Colors.white,
      child: Container(
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
        padding: EdgeInsets.fromLTRB(16, 8, 16, 12 + bottomPadding),
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
      // 更新配送员配送费预览，等待完成后再更新销售分润预览
      await _loadRiderDeliveryFeePreview();
      // 更新销售分润预览（确保配送员配送费预览已完成）
      _loadSalesCommissionPreview();
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
        // 删除改价信息（如果该商品被删除）
        _priceModifications.remove(itemId);
      });
      // 更新配送员配送费预览，等待完成后再更新销售分润预览
      await _loadRiderDeliveryFeePreview();
      // 更新销售分润预览（确保配送员配送费预览已完成）
      _loadSalesCommissionPreview();
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '删除商品失败'),
        ),
      );
    }
  }

  /// 改价对话框
  Future<void> _modifyPrice(
    int itemId,
    double originalPrice,
    double currentPrice,
    double costPrice,
  ) async {
    final priceController = TextEditingController(
      text: currentPrice.toStringAsFixed(2),
    );

    final result = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext) {
        String? errorMessage;

        return StatefulBuilder(
          builder: (context, setDialogState) => AlertDialog(
            title: const Text('修改价格'),
            content: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '原价：¥${originalPrice.toStringAsFixed(2)}',
                    style: const TextStyle(
                      fontSize: 14,
                      color: Color(0xFF8C92A4),
                    ),
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: priceController,
                    decoration: const InputDecoration(
                      labelText: '新价格',
                      hintText: '请输入新价格',
                      border: OutlineInputBorder(),
                      prefixText: '¥',
                    ),
                    keyboardType: const TextInputType.numberWithOptions(
                      decimal: true,
                    ),
                    autofocus: true,
                  ),
                  if (errorMessage != null) ...[
                    const SizedBox(height: 12),
                    Text(
                      errorMessage!,
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFFFF5A5F),
                      ),
                    ),
                  ],
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () {
                  Navigator.of(dialogContext).pop(false);
                },
                child: const Text('取消'),
              ),
              TextButton(
                onPressed: () {
                  final newPriceText = priceController.text.trim();

                  if (newPriceText.isEmpty) {
                    setDialogState(() {
                      errorMessage = '请输入新价格';
                    });
                    return;
                  }

                  final newPrice = double.tryParse(newPriceText);
                  if (newPrice == null || newPrice < 0) {
                    setDialogState(() {
                      errorMessage = '价格格式不正确';
                    });
                    return;
                  }

                  // 验证不能低于成本价
                  if (costPrice > 0 && newPrice < costPrice) {
                    setDialogState(() {
                      errorMessage = '价格不能低于成本价';
                    });
                    return;
                  }

                  Navigator.of(dialogContext).pop(true);
                },
                child: const Text('确定'),
              ),
            ],
          ),
        );
      },
    );

    // 延迟 dispose，确保对话框完全关闭
    await Future.delayed(const Duration(milliseconds: 150));
    priceController.dispose();

    if (result == true) {
      final newPriceText = priceController.text.trim();
      final newPrice = double.tryParse(newPriceText);

      if (newPrice != null && newPrice >= 0) {
        // 延迟更新状态，确保对话框完全关闭
        await Future.delayed(const Duration(milliseconds: 100));

        if (mounted) {
          setState(() {
            if (newPrice == originalPrice) {
              // 如果改回原价，删除改价信息
              _priceModifications.remove(itemId);
            } else {
              // 保存改价信息
              _priceModifications[itemId] = {'unit_price': newPrice};
            }
          });

          // 重新计算配送费和销售提成，等待配送费预览完成后再计算销售分润
          await _loadRiderDeliveryFeePreview();
          // 更新销售分润预览（确保配送员配送费预览已完成）
          _loadSalesCommissionPreview();
          // 重新计算汇总信息（使用改价后的价格）
          _recalculateSummary();
        }
      }
    }
  }

  /// 构建配送员配送费卡片（独立模块，页面底部）
  Widget _buildRiderDeliveryFeeCard() {
    if (_customerId == null ||
        _items.isEmpty ||
        _riderDeliveryFeePreview == null) {
      return const SizedBox.shrink();
    }

    final preview = _riderDeliveryFeePreview!;
    final riderPayableFee =
        (preview['rider_payable_fee'] as num?)?.toDouble() ?? 0.0;
    final baseFee = (preview['base_fee'] as num?)?.toDouble() ?? 0.0;
    final isolatedFee = (preview['isolated_fee'] as num?)?.toDouble() ?? 0.0;
    final itemFee = (preview['item_fee'] as num?)?.toDouble() ?? 0.0;
    final urgentFee = (preview['urgent_fee'] as num?)?.toDouble() ?? 0.0;
    final weatherFee = (preview['weather_fee'] as num?)?.toDouble() ?? 0.0;
    final profitShare = (preview['profit_share'] as num?)?.toDouble() ?? 0.0;

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
                '预估配送费',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFA940).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Text(
                  '预览',
                  style: TextStyle(
                    fontSize: 12,
                    color: Color(0xFFFFA940),
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          if (baseFee > 0) _buildDeliveryFeeRow('基础配送费', baseFee),
          if (isolatedFee > 0) _buildDeliveryFeeRow('孤立点补贴', isolatedFee),
          if (itemFee > 0) _buildDeliveryFeeRow('商品件数补贴', itemFee),
          if (urgentFee > 0)
            _buildDeliveryFeeRow('加急费用', urgentFee, highlight: true),
          if (weatherFee > 0) _buildDeliveryFeeRow('恶劣天气补贴', weatherFee),
          if (profitShare > 0)
            _buildDeliveryFeeRow('额外奖励', profitShare, highlight: true),
          const Divider(height: 20, thickness: 0.5, color: Color(0xFFE5E7F0)),
          Row(
            children: [
              const Text(
                '配送员收入',
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              Text(
                '¥${riderPayableFee.toStringAsFixed(2)}',
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF4C8DF6),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildDeliveryFeeRow(
    String label,
    double value, {
    bool highlight = false,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          Text(
            label,
            style: TextStyle(
              fontSize: 13,
              color: highlight
                  ? const Color(0xFF4C8DF6)
                  : const Color(0xFF40475C),
              fontWeight: highlight ? FontWeight.w600 : FontWeight.normal,
            ),
          ),
          const Spacer(),
          Text(
            '¥${value.toStringAsFixed(2)}',
            style: TextStyle(
              fontSize: 13,
              color: highlight
                  ? const Color(0xFF4C8DF6)
                  : const Color(0xFF40475C),
              fontWeight: highlight ? FontWeight.w600 : FontWeight.normal,
            ),
          ),
        ],
      ),
    );
  }

  /// 构建销售分润预览卡片
  Widget _buildSalesCommissionCard() {
    if (_customerId == null ||
        _items.isEmpty ||
        _salesCommissionPreview == null) {
      return const SizedBox.shrink();
    }

    final preview = _salesCommissionPreview!;
    final baseCommission =
        (preview['base_commission'] as num?)?.toDouble() ?? 0.0;
    final newCustomerBonus =
        (preview['new_customer_bonus'] as num?)?.toDouble() ?? 0.0;
    final tierCommission =
        (preview['tier_commission'] as num?)?.toDouble() ?? 0.0;
    final totalCommission =
        (preview['total_commission'] as num?)?.toDouble() ?? 0.0;
    final tierLevel = (preview['tier_level'] as int?) ?? 0;
    final isValidOrder = (preview['is_valid_order'] as bool?) ?? false;
    final isNewCustomerOrder =
        (preview['is_new_customer_order'] as bool?) ?? false;

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
                '预计销售分润',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFA940).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Text(
                  '预览',
                  style: TextStyle(
                    fontSize: 12,
                    color: Color(0xFFFFA940),
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          if (!isValidOrder) ...[
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xFFFF5A5F).withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Row(
                children: [
                  Icon(Icons.info_outline, size: 16, color: Color(0xFFFF5A5F)),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '无效订单（利润需>5元才计入有效分成）',
                      style: TextStyle(fontSize: 12, color: Color(0xFFFF5A5F)),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
          ],
          _buildCommissionRow('基础提成', baseCommission),
          if (isNewCustomerOrder) ...[
            const SizedBox(height: 8),
            _buildCommissionRow('新客开发激励', newCustomerBonus, highlight: true),
          ],
          if (tierLevel > 0) ...[
            const SizedBox(height: 8),
            _buildCommissionRow(
              '阶梯提成（阶梯$tierLevel）',
              tierCommission,
              highlight: true,
            ),
          ],
          const Divider(height: 20, thickness: 0.5, color: Color(0xFFE5E7F0)),
          Row(
            children: [
              const Text(
                '我的预计总分成',
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              Text(
                '¥${totalCommission.toStringAsFixed(2)}',
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFFFFA940),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            '订单收款后才会正式计入销售分成',
            style: TextStyle(fontSize: 12, color: Colors.grey[600]),
          ),
        ],
      ),
    );
  }

  Widget _buildCommissionRow(
    String label,
    double value, {
    bool highlight = false,
  }) {
    return Row(
      children: [
        Text(
          label,
          style: TextStyle(
            fontSize: 13,
            color: highlight
                ? const Color(0xFF20CB6B)
                : const Color(0xFF40475C),
            fontWeight: highlight ? FontWeight.w600 : FontWeight.normal,
          ),
        ),
        const Spacer(),
        Text(
          '¥${value.toStringAsFixed(2)}',
          style: TextStyle(
            fontSize: 13,
            color: highlight
                ? const Color(0xFF20CB6B)
                : const Color(0xFF40475C),
            fontWeight: highlight ? FontWeight.w600 : FontWeight.normal,
          ),
        ),
      ],
    );
  }

  /// 加载销售分润预览
  Future<void> _loadSalesCommissionPreview() async {
    // 防抖：如果距离上次调用不到300ms，则取消本次调用
    final now = DateTime.now();
    if (_lastSalesCommissionPreviewTime != null &&
        now.difference(_lastSalesCommissionPreviewTime!).inMilliseconds < 300) {
      return;
    }

    // 如果正在加载，则取消本次调用
    if (_isLoadingSalesCommissionPreview) {
      return;
    }

    if (_customerId == null || _items.isEmpty) {
      setState(() {
        _salesCommissionPreview = null;
      });
      return;
    }

    // 需要地址ID和汇总信息才能计算
    if (_selectedAddressId == null || _summary == null) {
      setState(() {
        _salesCommissionPreview = null;
      });
      return;
    }

    // 更新调用时间
    _lastSalesCommissionPreviewTime = now;
    _isLoadingSalesCommissionPreview = true;

    try {
      // 计算订单金额、商品成本、配送成本
      final totalAmount =
          (_summary!['total_amount'] as num?)?.toDouble() ?? 0.0;
      final totalDeliveryFee =
          (_summary!['delivery_fee'] as num?)?.toDouble() ?? 0.0;

      // 计算商品成本（从商品明细中计算）
      double goodsCost = 0.0;
      for (var item in _items) {
        final itemMap = item as Map<String, dynamic>;
        final specSnapshot =
            itemMap['spec_snapshot'] as Map<String, dynamic>? ?? {};
        final cost = (specSnapshot['cost'] as num?)?.toDouble() ?? 0.0;
        final quantity = (itemMap['quantity'] as int?) ?? 0;
        goodsCost += cost * quantity;
      }

      // 计算配送成本（从配送员配送费预览中获取）
      double deliveryCost = 0.0;
      if (_riderDeliveryFeePreview != null) {
        deliveryCost =
            (_riderDeliveryFeePreview!['rider_payable_fee'] as num?)
                ?.toDouble() ??
            0.0;
      }

      // 计算优惠券折扣
      final couponDiscount = _calculateCouponDiscount(
        _selectedCoupon,
        totalAmount,
        totalDeliveryFee,
      );

      // 计算加急费用
      final urgentFee = _isUrgent ? _urgentFee : 0.0;

      // 平台总收入 = 商品金额 + 配送费 + 加急费 - 优惠券折扣
      final orderAmount =
          totalAmount + totalDeliveryFee + urgentFee - couponDiscount;

      // 确保所有值都是有效的数字（不能为 null）
      final safeOrderAmount = orderAmount.isNaN ? 0.0 : orderAmount;
      final safeGoodsCost = goodsCost.isNaN ? 0.0 : goodsCost;
      final safeDeliveryCost = deliveryCost.isNaN ? 0.0 : deliveryCost;

      // 调用预览API
      final resp = await Request.post<Map<String, dynamic>>(
        '/employee/sales/commission/preview',
        body: {
          'order_amount': safeOrderAmount,
          'goods_cost': safeGoodsCost,
          'delivery_cost': safeDeliveryCost,
          'user_id': _customerId,
        },
        parser: (data) => data as Map<String, dynamic>,
      );

      if (!mounted) {
        _isLoadingSalesCommissionPreview = false;
        return;
      }

      if (resp.isSuccess && resp.data != null) {
        setState(() {
          _salesCommissionPreview = resp.data;
        });
      } else {
        setState(() {
          _salesCommissionPreview = null;
        });
      }
    } catch (e) {
      // 静默失败，不影响主流程
      if (mounted) {
        setState(() {
          _salesCommissionPreview = null;
        });
      }
    } finally {
      _isLoadingSalesCommissionPreview = false;
    }
  }

  /// 重新计算汇总信息（根据改价后的价格）
  void _recalculateSummary() {
    if (_items.isEmpty) {
      setState(() {
        _summary = null;
      });
      return;
    }

    // 计算商品总金额（使用改价后的价格）
    double totalAmount = 0.0;
    int totalQuantity = 0;

    for (final raw in _items) {
      final item = raw as Map<String, dynamic>;
      final itemId = item['id'] as int?;
      final qty = (item['quantity'] as int?) ?? 0;
      final snapshot = item['spec_snapshot'] as Map<String, dynamic>? ?? {};
      final retailPrice = (snapshot['retail_price'] as num?)?.toDouble() ?? 0.0;
      final wholesalePrice =
          (snapshot['wholesale_price'] as num?)?.toDouble() ?? 0.0;
      final cost = (snapshot['cost'] as num?)?.toDouble() ?? 0.0;

      final userType = (_user?['user_type'] as String?) ?? 'retail';
      // 计算原始价格
      double originalUnitPrice;
      if (userType == 'wholesale') {
        originalUnitPrice = wholesalePrice > 0 ? wholesalePrice : retailPrice;
      } else {
        originalUnitPrice = retailPrice > 0 ? retailPrice : wholesalePrice;
      }
      if (originalUnitPrice <= 0) {
        originalUnitPrice = cost > 0 ? cost : 0.0;
      }

      // 检查是否有改价
      double unitPrice = originalUnitPrice;
      if (itemId != null && _priceModifications.containsKey(itemId)) {
        final mod = _priceModifications[itemId]!;
        final modifiedPrice = (mod['unit_price'] as num?)?.toDouble();
        if (modifiedPrice != null && modifiedPrice >= 0) {
          unitPrice = modifiedPrice;
        }
      }

      totalAmount += unitPrice * qty;
      totalQuantity += qty;
    }

    // 获取原有的配送费信息（从 _summary 中获取，如果没有则从后端重新获取）
    final oldSummary = _summary ?? {};
    final deliveryFee = (oldSummary['delivery_fee'] as num?)?.toDouble() ?? 0.0;
    final freeShippingThreshold =
        (oldSummary['free_shipping_threshold'] as num?)?.toDouble() ?? 0.0;
    final isFreeShipping =
        totalAmount >= freeShippingThreshold && freeShippingThreshold > 0;

    // 更新汇总信息
    setState(() {
      _summary = {
        ...oldSummary, // 先展开旧数据
        'total_amount': totalAmount, // 然后覆盖为新计算的金额（改价后）
        'delivery_fee': isFreeShipping ? 0.0 : deliveryFee,
        'free_shipping_threshold': freeShippingThreshold,
        'is_free_shipping': isFreeShipping,
        'total_quantity': totalQuantity,
      };
    });

    // 如果配送费需要重新计算（因为商品金额变化可能影响免配送费判断），重新获取
    if (isFreeShipping != (oldSummary['is_free_shipping'] as bool? ?? false)) {
      // 重新获取采购单数据以更新配送费
      _reloadPurchaseListForSummary();
    }
  }

  /// 重新加载采购单数据以更新汇总信息（仅用于更新配送费）
  Future<void> _reloadPurchaseListForSummary() async {
    if (_customerId == null) return;

    try {
      final purchaseResp = await Request.get<Map<String, dynamic>>(
        '/employee/sales/customers/$_customerId/purchase-list',
        queryParams: {'is_urgent': _isUrgent.toString()},
        parser: (data) => data as Map<String, dynamic>,
      );

      if (!mounted) return;

      if (purchaseResp.isSuccess && purchaseResp.data != null) {
        final purchase = purchaseResp.data!;
        final purchaseSummary = purchase['summary'] as Map<String, dynamic>?;

        if (purchaseSummary != null) {
          // 只更新配送费相关字段，商品金额使用前端计算的（改价后的）
          final oldSummary = _summary ?? {};
          final calculatedTotalAmount =
              (oldSummary['total_amount'] as num?)?.toDouble() ?? 0.0;
          final calculatedTotalQuantity =
              (oldSummary['total_quantity'] as int?) ?? 0;

          setState(() {
            _summary = {
              ...purchaseSummary,
              'total_amount': calculatedTotalAmount, // 使用前端计算的商品金额（改价后）
              'total_quantity': calculatedTotalQuantity, // 使用前端计算的数量
            };
          });
        }
      }
    } catch (e) {
      // 静默失败，不影响主流程
      print('重新加载采购单汇总失败: $e');
    }
  }

  /// 加载配送员配送费预览
  Future<void> _loadRiderDeliveryFeePreview() async {
    if (_customerId == null || _items.isEmpty) {
      setState(() {
        _riderDeliveryFeePreview = null;
      });
      return;
    }

    // 需要地址ID才能计算孤立订单补贴和天气补贴
    if (_selectedAddressId == null) {
      setState(() {
        _riderDeliveryFeePreview = null;
      });
      return;
    }

    try {
      final resp = await Request.get<Map<String, dynamic>>(
        '/employee/sales/customers/$_customerId/rider-delivery-fee-preview',
        queryParams: {
          'address_id': _selectedAddressId.toString(),
          'is_urgent': _isUrgent.toString(),
        },
        parser: (data) => data as Map<String, dynamic>,
      );

      if (!mounted) return;

      if (resp.isSuccess && resp.data != null) {
        setState(() {
          _riderDeliveryFeePreview = resp.data;
        });
      } else {
        setState(() {
          _riderDeliveryFeePreview = null;
        });
      }
    } catch (e) {
      // 静默失败，不影响主流程
      if (mounted) {
        setState(() {
          _riderDeliveryFeePreview = null;
        });
      }
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

    if (!mounted) return;

    // 如果返回了数据，直接使用返回的数据更新
    if (result != null) {
      final items = result['items'] as List<dynamic>?;
      final summary = result['summary'] as Map<String, dynamic>?;
      if (items != null) {
        setState(() {
          _items = items;
          _summary = summary;
        });
        // 更新配送员配送费预览，等待完成后再更新销售分润预览
        await _loadRiderDeliveryFeePreview();
        // 更新销售分润预览（确保配送员配送费预览已完成）
        _loadSalesCommissionPreview();
      }
    } else {
      // 如果没有返回数据（用户可能直接返回），重新加载数据以确保同步
      await _loadData();
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
  List<Map<String, dynamic>> _frequentProducts = [];
  bool _isLoadingFrequent = false;
  int _selectedTab = 0; // 0: 全部商品, 1: 常购商品

  @override
  void initState() {
    super.initState();
    _searchController.addListener(_onSearchChanged);
    _loadFrequentProducts();
    _loadProducts();
  }

  void _onSearchChanged() {
    setState(() {
      // 搜索内容变化时触发重新加载
    });
  }

  @override
  void dispose() {
    _searchController.removeListener(_onSearchChanged);
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadFrequentProducts() async {
    setState(() {
      _isLoadingFrequent = true;
    });

    final resp = await Request.get<List<dynamic>>(
      '/employee/sales/customers/${widget.customerId}/frequent-products',
      parser: (data) => data as List<dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      setState(() {
        _frequentProducts = resp.data!.cast<Map<String, dynamic>>();
        _isLoadingFrequent = false;
      });
    } else {
      setState(() {
        _isLoadingFrequent = false;
      });
      // 常购商品加载失败不显示错误提示，因为可能客户没有常购商品
    }
  }

  Future<void> _loadProducts() async {
    setState(() {
      _isLoading = true;
    });

    final resp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/products',
      queryParams: {
        'pageNum': '1',
        'pageSize': '100', // 增加页面大小以显示更多商品
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

  void _clearSearch() {
    _searchController.clear();
    _loadProducts();
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

  // 快速添加常购商品到采购单
  Future<void> _quickAddFrequentProduct(
    Map<String, dynamic> frequentProduct,
  ) async {
    final product = frequentProduct['product'] as Map<String, dynamic>?;
    if (product == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('商品信息不完整')));
      return;
    }

    final productId = product['id'] as int?;
    final specName = frequentProduct['spec_name'] as String? ?? '';
    final specs = product['specs'] as List<dynamic>? ?? [];

    if (productId == null || specName.isEmpty) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('商品或规格信息不完整')));
      return;
    }

    // 查找对应的规格
    Map<String, dynamic>? targetSpec;
    for (var spec in specs) {
      final s = spec as Map<String, dynamic>;
      if ((s['name'] as String? ?? '') == specName) {
        targetSpec = s;
        break;
      }
    }

    if (targetSpec == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('规格不存在')));
      return;
    }

    // 直接添加到采购单，数量默认为1
    await _addToPurchaseList(product, targetSpec, 1);
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
                  Container(
                    padding: const EdgeInsets.fromLTRB(20, 16, 12, 12),
                    decoration: const BoxDecoration(
                      border: Border(
                        bottom: BorderSide(color: Color(0xFFE5E7F0), width: 1),
                      ),
                    ),
                    child: Row(
                      children: [
                        Expanded(
                          child: Text(
                            product['name'] as String? ?? '',
                            style: const TextStyle(
                              fontSize: 17,
                              fontWeight: FontWeight.w600,
                              color: Color(0xFF20253A),
                            ),
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        IconButton(
                          icon: const Icon(
                            Icons.close,
                            color: Color(0xFF8C92A4),
                          ),
                          onPressed: () => Navigator.of(context).pop(),
                        ),
                      ],
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.fromLTRB(20, 16, 20, 20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          '选择规格',
                          style: TextStyle(
                            fontSize: 15,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                        const SizedBox(height: 12),
                        Wrap(
                          spacing: 10,
                          runSpacing: 10,
                          children: specs.map((raw) {
                            final spec = raw as Map<String, dynamic>;
                            final name = spec['name'] as String? ?? '';
                            final desc = spec['description'] as String? ?? '';
                            final isSelected = identical(spec, selectedSpec);
                            return InkWell(
                              onTap: () {
                                setState(() {
                                  selectedSpec = spec;
                                });
                              },
                              borderRadius: BorderRadius.circular(12),
                              child: Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 16,
                                  vertical: 12,
                                ),
                                decoration: BoxDecoration(
                                  color: isSelected
                                      ? const Color(0xFF20CB6B)
                                      : const Color(0xFFF5F6FA),
                                  borderRadius: BorderRadius.circular(12),
                                  border: Border.all(
                                    color: isSelected
                                        ? const Color(0xFF20CB6B)
                                        : const Color(0xFFE5E7F0),
                                    width: isSelected ? 2 : 1,
                                  ),
                                ),
                                child: Text(
                                  desc.isNotEmpty ? '$name（$desc）' : name,
                                  style: TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w600,
                                    color: isSelected
                                        ? Colors.white
                                        : const Color(0xFF40475C),
                                  ),
                                ),
                              ),
                            );
                          }).toList(),
                        ),
                        const SizedBox(height: 20),
                        const Text(
                          '购买数量',
                          style: TextStyle(
                            fontSize: 15,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                        const SizedBox(height: 8),
                        Row(
                          children: [
                            Container(
                              decoration: BoxDecoration(
                                color: const Color(0xFFF5F6FA),
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(
                                  color: const Color(0xFFE5E7F0),
                                  width: 1,
                                ),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  InkWell(
                                    onTap: quantity > 1
                                        ? () {
                                            setState(() {
                                              quantity--;
                                            });
                                          }
                                        : null,
                                    borderRadius: const BorderRadius.only(
                                      topLeft: Radius.circular(12),
                                      bottomLeft: Radius.circular(12),
                                    ),
                                    child: Container(
                                      width: 44,
                                      height: 44,
                                      alignment: Alignment.center,
                                      child: Icon(
                                        Icons.remove,
                                        size: 20,
                                        color: quantity > 1
                                            ? const Color(0xFF40475C)
                                            : const Color(0xFFB0B4C3),
                                      ),
                                    ),
                                  ),
                                  Container(
                                    width: 60,
                                    height: 44,
                                    alignment: Alignment.center,
                                    decoration: const BoxDecoration(
                                      border: Border.symmetric(
                                        vertical: BorderSide(
                                          color: Color(0xFFE5E7F0),
                                          width: 1,
                                        ),
                                      ),
                                    ),
                                    child: Text(
                                      '$quantity',
                                      style: const TextStyle(
                                        fontSize: 16,
                                        fontWeight: FontWeight.w600,
                                        color: Color(0xFF20253A),
                                      ),
                                    ),
                                  ),
                                  InkWell(
                                    onTap: () {
                                      setState(() {
                                        quantity++;
                                      });
                                    },
                                    borderRadius: const BorderRadius.only(
                                      topRight: Radius.circular(12),
                                      bottomRight: Radius.circular(12),
                                    ),
                                    child: Container(
                                      width: 44,
                                      height: 44,
                                      alignment: Alignment.center,
                                      child: const Icon(
                                        Icons.add,
                                        size: 20,
                                        color: Color(0xFF20CB6B),
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 20),
                        SizedBox(
                          width: double.infinity,
                          height: 48,
                          child: ElevatedButton(
                            onPressed: selectedSpec == null
                                ? null
                                : () => _addToPurchaseList(
                                    product,
                                    selectedSpec!,
                                    quantity,
                                  ),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: selectedSpec == null
                                  ? const Color(0xFFE5E7F0)
                                  : const Color(0xFF20CB6B),
                              foregroundColor: selectedSpec == null
                                  ? const Color(0xFF8C92A4)
                                  : Colors.white,
                              padding: const EdgeInsets.symmetric(vertical: 12),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(12),
                              ),
                              elevation: 0,
                            ),
                            child: const Text(
                              '加入采购单',
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
          child: Column(
            children: [
              // 搜索栏 - 固定在绿色背景区域
              Container(
                color: const Color(0xFF20CB6B),
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
                child: Container(
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(12),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.06),
                        blurRadius: 8,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Row(
                    children: [
                      const SizedBox(width: 12),
                      const Icon(
                        Icons.search,
                        color: Color(0xFF8C92A4),
                        size: 20,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: TextField(
                          controller: _searchController,
                          decoration: const InputDecoration(
                            hintText: '输入商品名称或编码搜索',
                            hintStyle: TextStyle(
                              fontSize: 14,
                              color: Color(0xFF8C92A4),
                            ),
                            border: InputBorder.none,
                            contentPadding: EdgeInsets.symmetric(vertical: 12),
                          ),
                          textInputAction: TextInputAction.search,
                          onSubmitted: (_) => _loadProducts(),
                          onChanged: (_) {
                            setState(() {});
                            if (_searchController.text.trim().isEmpty) {
                              _loadProducts();
                            }
                          },
                          style: const TextStyle(
                            fontSize: 14,
                            color: Color(0xFF20253A),
                          ),
                        ),
                      ),
                      // 清除按钮
                      if (_searchController.text.isNotEmpty)
                        InkWell(
                          onTap: _clearSearch,
                          borderRadius: BorderRadius.circular(12),
                          child: Container(
                            padding: const EdgeInsets.all(8),
                            child: const Icon(
                              Icons.clear,
                              size: 18,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                        ),
                      InkWell(
                        onTap: _loadProducts,
                        borderRadius: BorderRadius.circular(12),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 8,
                          ),
                          child: const Text(
                            '搜索',
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: Color(0xFF20CB6B),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),
                    ],
                  ),
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
                  clipBehavior: Clip.antiAlias,
                  child: _searchController.text.trim().isNotEmpty
                      ? _buildSearchProductsList() // 搜索时直接显示搜索结果
                      : Column(
                          children: [
                            // 标签页切换（仅在非搜索状态显示）
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 12,
                              ),
                              decoration: const BoxDecoration(
                                border: Border(
                                  bottom: BorderSide(
                                    color: Color(0xFFE5E7F0),
                                    width: 1,
                                  ),
                                ),
                              ),
                              child: Row(
                                children: [
                                  Expanded(
                                    child: InkWell(
                                      onTap: () {
                                        setState(() {
                                          _selectedTab = 0;
                                        });
                                      },
                                      child: Container(
                                        padding: const EdgeInsets.symmetric(
                                          vertical: 8,
                                        ),
                                        decoration: BoxDecoration(
                                          color: _selectedTab == 0
                                              ? const Color(0xFF20CB6B)
                                              : Colors.transparent,
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                        ),
                                        child: Center(
                                          child: Text(
                                            '全部商品',
                                            style: TextStyle(
                                              fontSize: 14,
                                              fontWeight: FontWeight.w600,
                                              color: _selectedTab == 0
                                                  ? Colors.white
                                                  : const Color(0xFF8C92A4),
                                            ),
                                          ),
                                        ),
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: InkWell(
                                      onTap: () {
                                        setState(() {
                                          _selectedTab = 1;
                                        });
                                      },
                                      child: Container(
                                        padding: const EdgeInsets.symmetric(
                                          vertical: 8,
                                        ),
                                        decoration: BoxDecoration(
                                          color: _selectedTab == 1
                                              ? const Color(0xFF20CB6B)
                                              : Colors.transparent,
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                        ),
                                        child: Center(
                                          child: Text(
                                            '常购商品',
                                            style: TextStyle(
                                              fontSize: 14,
                                              fontWeight: FontWeight.w600,
                                              color: _selectedTab == 1
                                                  ? Colors.white
                                                  : const Color(0xFF8C92A4),
                                            ),
                                          ),
                                        ),
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            // 内容区域
                            Expanded(
                              child: _selectedTab == 0
                                  ? _buildSearchProductsList() // 全部商品
                                  : _buildFrequentProductsList(), // 常购商品
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

  Widget _buildFrequentProductsList() {
    if (_isLoadingFrequent) {
      return const Center(
        child: CircularProgressIndicator(
          valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
        ),
      );
    }

    if (_frequentProducts.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.shopping_bag_outlined,
              size: 64,
              color: Color(0xFFB0B4C3),
            ),
            SizedBox(height: 16),
            Text(
              '暂无常购商品',
              style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
            ),
            SizedBox(height: 4),
            Text(
              '客户下单后这里会显示常购商品',
              style: TextStyle(fontSize: 12, color: Color(0xFFB0B4C3)),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: _frequentProducts.length,
      itemBuilder: (context, index) {
        final frequentProduct = _frequentProducts[index];
        final product = frequentProduct['product'] as Map<String, dynamic>?;
        final productName = frequentProduct['product_name'] as String? ?? '';
        final specName = frequentProduct['spec_name'] as String? ?? '';
        final image = frequentProduct['image'] as String? ?? '';
        final buyCount = frequentProduct['buy_count'] as int? ?? 0;

        if (product == null) {
          return const SizedBox.shrink();
        }

        final images = product['images'] as List<dynamic>? ?? [];
        final productImage = images.isNotEmpty ? images[0] as String? : image;

        return Container(
          margin: const EdgeInsets.only(bottom: 12),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
              color: const Color(0xFF20CB6B).withOpacity(0.2),
              width: 1,
            ),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.04),
                blurRadius: 10,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: InkWell(
            onTap: () => _quickAddFrequentProduct(frequentProduct),
            borderRadius: BorderRadius.circular(16),
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Row(
                children: [
                  // 商品图片
                  Container(
                    width: 80,
                    height: 80,
                    decoration: BoxDecoration(
                      color: const Color(0xFFF5F6FA),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    clipBehavior: Clip.antiAlias,
                    child: (productImage ?? '').isNotEmpty
                        ? Image.network(
                            productImage!,
                            fit: BoxFit.cover,
                            errorBuilder: (context, error, stackTrace) {
                              return const Icon(
                                Icons.image_not_supported,
                                color: Color(0xFFB0B4C3),
                                size: 32,
                              );
                            },
                          )
                        : const Icon(
                            Icons.image,
                            color: Color(0xFFB0B4C3),
                            size: 32,
                          ),
                  ),
                  const SizedBox(width: 12),
                  // 商品信息
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          productName,
                          style: const TextStyle(
                            fontSize: 15,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                        if (specName.isNotEmpty) ...[
                          const SizedBox(height: 4),
                          Text(
                            specName,
                            style: const TextStyle(
                              fontSize: 13,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                        ],
                        const SizedBox(height: 8),
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: const Color(0xFF20CB6B).withOpacity(0.1),
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: Text(
                                '已买$buyCount次',
                                style: const TextStyle(
                                  fontSize: 11,
                                  color: Color(0xFF20CB6B),
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                            const Spacer(),
                            const Icon(
                              Icons.add_circle,
                              color: Color(0xFF20CB6B),
                              size: 24,
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildSearchProductsList() {
    if (_isLoading) {
      return const Center(
        child: CircularProgressIndicator(
          valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
        ),
      );
    }

    if (_products.isEmpty) {
      return const Center(
        child: Text(
          '暂无商品数据',
          style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: _products.length,
      itemBuilder: (context, index) {
        final product = _products[index];
        final name = (product['name'] as String?) ?? '';
        final desc = (product['description'] as String?) ?? '';
        final images = product['images'] as List<dynamic>? ?? [];
        final image = images.isNotEmpty ? images[0] as String? : '';
        final isSpecial = (product['is_special'] as bool?) ?? false;
        return Container(
          margin: const EdgeInsets.only(bottom: 12),
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
          child: InkWell(
            onTap: () => _openSpecSelector(product),
            borderRadius: BorderRadius.circular(16),
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Row(
                children: [
                  // 商品图片
                  Container(
                    width: 80,
                    height: 80,
                    decoration: BoxDecoration(
                      color: const Color(0xFFF5F6FA),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    clipBehavior: Clip.antiAlias,
                    child: (image ?? '').isNotEmpty
                        ? Image.network(
                            image!,
                            fit: BoxFit.cover,
                            errorBuilder: (context, error, stackTrace) {
                              return const Icon(
                                Icons.image_not_supported,
                                color: Color(0xFFB0B4C3),
                                size: 32,
                              );
                            },
                          )
                        : const Icon(
                            Icons.image,
                            color: Color(0xFFB0B4C3),
                            size: 32,
                          ),
                  ),
                  const SizedBox(width: 12),
                  // 商品信息
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            if (isSpecial)
                              Container(
                                margin: const EdgeInsets.only(right: 6),
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 6,
                                  vertical: 2,
                                ),
                                decoration: BoxDecoration(
                                  color: const Color(
                                    0xFFFF9800,
                                  ).withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(4),
                                ),
                                child: const Text(
                                  '精选',
                                  style: TextStyle(
                                    fontSize: 10,
                                    color: Color(0xFFFF9800),
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                              ),
                            Expanded(
                              child: Text(
                                name,
                                style: const TextStyle(
                                  fontSize: 15,
                                  fontWeight: FontWeight.w600,
                                  color: Color(0xFF20253A),
                                ),
                                maxLines: 2,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          ],
                        ),
                        if (desc.isNotEmpty) ...[
                          const SizedBox(height: 4),
                          Text(
                            desc,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF8C92A4),
                            ),
                          ),
                        ],
                        const SizedBox(height: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: const Color(0xFF20CB6B).withOpacity(0.1),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: const Text(
                            '选择规格',
                            style: TextStyle(
                              fontSize: 12,
                              color: Color(0xFF20CB6B),
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  const Icon(
                    Icons.chevron_right,
                    color: Color(0xFF8C92A4),
                    size: 20,
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}

/// 地址选择页面
class AddressSelectionPage extends StatefulWidget {
  final int customerId;
  final List<dynamic> addresses;
  final int? selectedAddressId;

  const AddressSelectionPage({
    super.key,
    required this.customerId,
    required this.addresses,
    this.selectedAddressId,
  });

  @override
  State<AddressSelectionPage> createState() => _AddressSelectionPageState();
}

class _AddressSelectionPageState extends State<AddressSelectionPage> {
  int? _selectedAddressId;

  @override
  void initState() {
    super.initState();
    _selectedAddressId = widget.selectedAddressId;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('选择收货地址'),
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
          child: widget.addresses.isEmpty
              ? Center(
                  child: Container(
                    margin: const EdgeInsets.all(16),
                    padding: const EdgeInsets.all(24),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(
                          Icons.location_off,
                          size: 64,
                          color: Color(0xFF8C92A4),
                        ),
                        const SizedBox(height: 16),
                        const Text(
                          '该客户暂无地址',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                        const SizedBox(height: 8),
                        const Text(
                          '请先在"新客资料"中为客户添加地址',
                          style: TextStyle(
                            fontSize: 13,
                            color: Color(0xFF8C92A4),
                          ),
                          textAlign: TextAlign.center,
                        ),
                      ],
                    ),
                  ),
                )
              : ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    ...widget.addresses.map((raw) {
                      final addr = raw as Map<String, dynamic>;
                      final id = addr['id'] as int?;
                      final name = (addr['name'] as String?) ?? '收货地址';
                      final text = (addr['address'] as String?) ?? '';
                      final contact = (addr['contact'] as String?) ?? '';
                      final phone = (addr['phone'] as String?) ?? '';
                      final isDefault = (addr['is_default'] as bool?) ?? false;

                      if (id == null) return const SizedBox.shrink();

                      final isSelected = _selectedAddressId == id;

                      return Container(
                        margin: const EdgeInsets.only(bottom: 12),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(
                            color: isSelected
                                ? const Color(0xFF20CB6B)
                                : const Color(0xFFE5E7F0),
                            width: isSelected ? 2 : 1,
                          ),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.04),
                              blurRadius: 10,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                        child: InkWell(
                          onTap: () {
                            setState(() {
                              _selectedAddressId = id;
                            });
                            Navigator.of(context).pop<Map<String, dynamic>>({
                              'id': id,
                              'address': addr,
                            });
                          },
                          borderRadius: BorderRadius.circular(16),
                          child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                // 选中图标
                                Container(
                                  width: 24,
                                  height: 24,
                                  margin: const EdgeInsets.only(right: 12),
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
                                      ? const Icon(
                                          Icons.check,
                                          size: 16,
                                          color: Colors.white,
                                        )
                                      : null,
                                ),
                                // 地址信息
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
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
                                          if (isDefault)
                                            Container(
                                              margin: const EdgeInsets.only(
                                                left: 6,
                                              ),
                                              padding:
                                                  const EdgeInsets.symmetric(
                                                    horizontal: 6,
                                                    vertical: 2,
                                                  ),
                                              decoration: BoxDecoration(
                                                color: const Color(
                                                  0xFF20CB6B,
                                                ).withOpacity(0.08),
                                                borderRadius:
                                                    BorderRadius.circular(10),
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
                                      const SizedBox(height: 6),
                                      if (text.isNotEmpty)
                                        Text(
                                          text,
                                          style: const TextStyle(
                                            fontSize: 13,
                                            color: Color(0xFF40475C),
                                          ),
                                        ),
                                      if (contact.isNotEmpty ||
                                          phone.isNotEmpty) ...[
                                        const SizedBox(height: 4),
                                        Text(
                                          '$contact  $phone',
                                          style: const TextStyle(
                                            fontSize: 12,
                                            color: Color(0xFF8C92A4),
                                          ),
                                        ),
                                      ],
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      );
                    }).toList(),
                  ],
                ),
        ),
      ),
    );
  }
}
