import 'package:flutter/material.dart';
import 'package:employees_app/utils/request.dart';
import 'package:intl/intl.dart';

/// 送优惠券页面（销售员）
/// 步骤：1）选择客户 2）查看客户已有优惠券 3）选择要发放的优惠券 4）确认发放
class CouponSendPage extends StatefulWidget {
  const CouponSendPage({super.key});

  @override
  State<CouponSendPage> createState() => _CouponSendPageState();
}

class _CouponSendPageState extends State<CouponSendPage> {
  final TextEditingController _searchController = TextEditingController();
  bool _isSearchingCustomer = false;
  bool _isLoadingCoupons = false;
  bool _isIssuing = false;

  // 发放原因预设选项
  final List<String> _reasons = const ['潜在客户', '优质客户', '老客户关怀', '活动赠送', '售后补偿'];
  String? _selectedReason;

  List<Map<String, dynamic>> _customerResults = [];
  Map<String, dynamic>? _selectedCustomer;
  int _customerCouponCount = 0;

  List<Map<String, dynamic>> _coupons = [];
  Map<String, dynamic>? _selectedCoupon;
  int _quantity = 1;

  // 有效期设置
  String _expireType = 'none'; // none, days, date
  int _expireDays = 30;
  DateTime? _expireDateTime;

  @override
  void initState() {
    super.initState();
    _loadCoupons();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadCoupons() async {
    setState(() {
      _isLoadingCoupons = true;
    });

    final resp = await Request.get<List<dynamic>>(
      '/employee/sales/coupons',
      parser: (data) => data as List<dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      setState(() {
        _coupons = resp.data!
            .cast<Map<String, dynamic>>()
            .where((e) => (e['status'] as int? ?? 0) == 1)
            .toList();
        if (_coupons.isNotEmpty) {
          _selectedCoupon = _coupons.first;
        }
      });
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '获取优惠券列表失败'),
        ),
      );
    }

    if (mounted) {
      setState(() {
        _isLoadingCoupons = false;
      });
    }
  }

  Future<void> _searchCustomer() async {
    final keyword = _searchController.text.trim();
    if (keyword.isEmpty) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('请输入客户名称 / 电话 / 编号')));
      return;
    }

    setState(() {
      _isSearchingCustomer = true;
      _customerResults.clear();
      _selectedCustomer = null;
      _customerCouponCount = 0;
    });

    final resp = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customers',
      queryParams: {'pageNum': '1', 'pageSize': '20', 'keyword': keyword},
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      final list = (resp.data!['list'] as List<dynamic>? ?? [])
          .cast<Map<String, dynamic>>();
      setState(() {
        _customerResults = list;
      });
      if (list.isEmpty) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(const SnackBar(content: Text('未找到相关客户')));
      }
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '搜索客户失败'),
        ),
      );
    }

    if (mounted) {
      setState(() {
        _isSearchingCustomer = false;
      });
    }
  }

  Future<void> _loadCustomerCoupons(Map<String, dynamic> customer) async {
    final id = customer['id'] as int?;
    if (id == null || id <= 0) return;

    setState(() {
      _selectedCustomer = customer;
      _customerCouponCount = 0;
    });

    final resp = await Request.get<List<dynamic>>(
      '/employee/sales/customers/$id/coupons',
      parser: (data) => data as List<dynamic>,
    );

    if (!mounted) return;

    if (resp.isSuccess && resp.data != null) {
      setState(() {
        _customerCouponCount = resp.data!.length;
      });
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '获取客户优惠券失败'),
        ),
      );
    }
  }

  Future<void> _issueCoupon() async {
    if (_selectedCustomer == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('请先选择客户')));
      return;
    }
    if (_selectedCoupon == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('请选择要发放的优惠券')));
      return;
    }
    if (_quantity <= 0) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('发放数量必须大于0')));
      return;
    }

    if (_selectedReason == null || _selectedReason!.isEmpty) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('请选择发放原因')));
      return;
    }

    final userId = _selectedCustomer!['id'] as int? ?? 0;
    final couponId = _selectedCoupon!['id'] as int? ?? 0;

    setState(() {
      _isIssuing = true;
    });

    final body = <String, dynamic>{
      'user_id': userId,
      'coupon_id': couponId,
      'quantity': _quantity,
      'reason': _selectedReason ?? '',
    };

    // 有效期参数
    if (_expireType == 'days') {
      body['expires_in'] = _expireDays;
    } else if (_expireType == 'date' && _expireDateTime != null) {
      body['expires_at'] =
          DateFormat('yyyy-MM-dd HH:mm:ss').format(_expireDateTime!);
    }

    final resp = await Request.post<dynamic>(
      '/employee/sales/coupons/issue',
      body: body,
    );

    if (!mounted) return;

    if (resp.isSuccess) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('优惠券发放成功')));
      // 成功后刷新客户的优惠券数量
      await _loadCustomerCoupons(_selectedCustomer!);
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.message.isNotEmpty ? resp.message : '发放优惠券失败'),
        ),
      );
    }

    if (mounted) {
      setState(() {
        _isIssuing = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('送优惠券'),
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
              _buildCustomerSearchBar(),
              Expanded(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                  child: Column(
                    children: [
                      _buildCustomerSection(),
                      const SizedBox(height: 12),
                      _buildCouponSection(),
                    ],
                  ),
                ),
              ),
              _buildBottomButton(),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCustomerSearchBar() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
      child: Row(
        children: [
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(24),
              ),
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: TextField(
                controller: _searchController,
                decoration: const InputDecoration(
                  hintText: '输入客户名称 / 电话 / 编号',
                  border: InputBorder.none,
                  icon: Icon(Icons.search, color: Color(0xFF8C92A4)),
                ),
                onSubmitted: (_) => _searchCustomer(),
              ),
            ),
          ),
          const SizedBox(width: 8),
          ElevatedButton(
            onPressed: _isSearchingCustomer ? null : _searchCustomer,
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.white,
              foregroundColor: const Color(0xFF20CB6B),
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(20),
              ),
              elevation: 0,
            ),
            child: _isSearchingCustomer
                ? const SizedBox(
                    width: 14,
                    height: 14,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      valueColor: AlwaysStoppedAnimation<Color>(
                        Color(0xFF20CB6B),
                      ),
                    ),
                  )
                : const Text(
                    '搜索',
                    style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildCustomerSection() {
    return Container(
      width: double.infinity,
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
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Text(
                  '选择客户',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                ),
                const SizedBox(width: 8),
                if (_selectedCustomer != null)
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFF20CB6B).withOpacity(0.08),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      '该客户现有$_customerCouponCount 张优惠券',
                      style: const TextStyle(
                        fontSize: 11,
                        color: Color(0xFF20CB6B),
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            if (_customerResults.isEmpty)
              const Text(
                '请先通过上方搜索找到客户',
                style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
              )
            else
              Column(
                children: _customerResults.map((c) {
                  final id = c['id'] as int? ?? 0;
                  final name = (c['name'] as String?) ?? '未填写名称';
                  final phone = (c['phone'] as String?) ?? '';
                  final userCode = (c['user_code'] as String?) ?? '';
                  final createdAt = c['created_at']?.toString() ?? '';
                  final isSelected =
                      _selectedCustomer != null &&
                      _selectedCustomer!['id'] == id;

                  return InkWell(
                    onTap: () => _loadCustomerCoupons(c),
                    borderRadius: BorderRadius.circular(10),
                    child: Container(
                      margin: const EdgeInsets.only(top: 8),
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: isSelected
                            ? const Color(0xFF20CB6B).withOpacity(0.06)
                            : const Color(0xFFF7F8FA),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
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
                                    if (userCode.isNotEmpty)
                                      Container(
                                        padding: const EdgeInsets.symmetric(
                                          horizontal: 6,
                                          vertical: 2,
                                        ),
                                        decoration: BoxDecoration(
                                          color: const Color(
                                            0xFF20CB6B,
                                          ).withOpacity(0.08),
                                          borderRadius: BorderRadius.circular(
                                            10,
                                          ),
                                        ),
                                        child: Text(
                                          '编号 $userCode',
                                          style: const TextStyle(
                                            fontSize: 10,
                                            color: Color(0xFF20CB6B),
                                          ),
                                        ),
                                      ),
                                  ],
                                ),
                                const SizedBox(height: 2),
                                if (phone.isNotEmpty)
                                  Text(
                                    phone,
                                    style: const TextStyle(
                                      fontSize: 13,
                                      color: Color(0xFF40475C),
                                    ),
                                  ),
                                if (createdAt.isNotEmpty)
                                  Text(
                                    '绑定时间：${_formatTime(createdAt)}',
                                    style: const TextStyle(
                                      fontSize: 11,
                                      color: Color(0xFFB0B4C3),
                                    ),
                                  ),
                              ],
                            ),
                          ),
                          const SizedBox(width: 8),
                          Icon(
                            isSelected
                                ? Icons.radio_button_checked
                                : Icons.radio_button_unchecked,
                            size: 18,
                            color: const Color(0xFF20CB6B),
                          ),
                        ],
                      ),
                    ),
                  );
                }).toList(),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildCouponSection() {
    return Container(
      width: double.infinity,
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
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  '选择优惠券',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                ),
                if (_isLoadingCoupons)
                  const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      valueColor: AlwaysStoppedAnimation<Color>(
                        Color(0xFF20CB6B),
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            if (_coupons.isEmpty && !_isLoadingCoupons)
              const Text(
                '暂无可发放的优惠券，请先在后台配置优惠券。',
                style: TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
              )
            else if (_coupons.isNotEmpty)
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  DropdownButton<Map<String, dynamic>>(
                    value: _selectedCoupon,
                    isExpanded: true,
                    underline: const SizedBox(),
                    icon: const Icon(
                      Icons.arrow_drop_down,
                      color: Color(0xFF20CB6B),
                    ),
                    items: _coupons.map((coupon) {
                      final name = coupon['name'] as String? ?? '';
                      final type = coupon['type'] as String? ?? '';
                      final discount =
                          (coupon['discount_value'] as num?)?.toDouble() ?? 0.0;
                      final minAmount =
                          (coupon['min_amount'] as num?)?.toDouble() ?? 0.0;

                      final typeLabel = type == 'delivery_fee'
                          ? '免配送费券'
                          : '金额券';

                      final title =
                          '$name（$typeLabel，满${minAmount.toStringAsFixed(0)}减${discount.toStringAsFixed(0)}）';

                      return DropdownMenuItem<Map<String, dynamic>>(
                        value: coupon,
                        child: Text(
                          title,
                          style: const TextStyle(
                            fontSize: 13,
                            color: Color(0xFF20253A),
                          ),
                        ),
                      );
                    }).toList(),
                    onChanged: (value) {
                      setState(() {
                        _selectedCoupon = value;
                      });
                    },
                  ),
                  const SizedBox(height: 8),
                  if (_selectedCoupon != null)
                    _buildCouponDetail(_selectedCoupon!),
                  const SizedBox(height: 12),
                  _buildQuantitySelector(),
                  const SizedBox(height: 12),
                  _buildExpireSelector(),
                  const SizedBox(height: 12),
                  _buildReasonSelector(),
                ],
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildCouponDetail(Map<String, dynamic> coupon) {
    final description = coupon['description'] as String? ?? '';
    final validFrom = coupon['valid_from']?.toString() ?? '';
    final validTo = coupon['valid_to']?.toString() ?? '';

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (description.isNotEmpty)
          Text(
            description,
            style: const TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
          ),
        const SizedBox(height: 4),
        if (validFrom.isNotEmpty && validTo.isNotEmpty)
          Text(
            '有效期：${_formatTime(validFrom)} ~ ${_formatTime(validTo)}',
            style: const TextStyle(fontSize: 11, color: Color(0xFFB0B4C3)),
          ),
      ],
    );
  }

  Widget _buildQuantitySelector() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.start,
      children: [
        const Text(
          '发放数量',
          style: TextStyle(fontSize: 13, color: Color(0xFF20253A)),
        ),
        const SizedBox(width: 12),
        Container(
          decoration: BoxDecoration(
            color: const Color(0xFFF7F8FA),
            borderRadius: BorderRadius.circular(18),
          ),
          child: Row(
            children: [
              IconButton(
                onPressed: _quantity > 1
                    ? () {
                        setState(() {
                          _quantity--;
                        });
                      }
                    : null,
                icon: const Icon(Icons.remove, size: 18),
                constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                padding: EdgeInsets.zero,
              ),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 8),
                child: Text(
                  '$_quantity',
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
              IconButton(
                onPressed: () {
                  setState(() {
                    _quantity++;
                  });
                },
                icon: const Icon(Icons.add, size: 18),
                constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                padding: EdgeInsets.zero,
              ),
            ],
          ),
        ),
      ],
    );
  }

  /// 有效期设置
  Widget _buildExpireSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          '有效期设置',
          style: TextStyle(fontSize: 13, color: Color(0xFF20253A)),
        ),
        const SizedBox(height: 8),
        Row(
          children: [
            Expanded(
              child: Row(
                children: [
                  Radio<String>(
                    value: 'none',
                    groupValue: _expireType,
                    activeColor: const Color(0xFF20CB6B),
                    onChanged: (v) {
                      if (v == null) return;
                      setState(() {
                        _expireType = v;
                      });
                    },
                  ),
                  const Text(
                    '不限制',
                    style: TextStyle(fontSize: 13, color: Color(0xFF40475C)),
                  ),
                ],
              ),
            ),
            Expanded(
              child: Row(
                children: [
                  Radio<String>(
                    value: 'days',
                    groupValue: _expireType,
                    activeColor: const Color(0xFF20CB6B),
                    onChanged: (v) {
                      if (v == null) return;
                      setState(() {
                        _expireType = v;
                      });
                    },
                  ),
                  const Text(
                    'N天后过期',
                    style: TextStyle(fontSize: 13, color: Color(0xFF40475C)),
                  ),
                ],
              ),
            ),
            Expanded(
              child: Row(
                children: [
                  Radio<String>(
                    value: 'date',
                    groupValue: _expireType,
                    activeColor: const Color(0xFF20CB6B),
                    onChanged: (v) {
                      if (v == null) return;
                      setState(() {
                        _expireType = v;
                      });
                    },
                  ),
                  const Text(
                    '指定日期',
                    style: TextStyle(fontSize: 13, color: Color(0xFF40475C)),
                  ),
                ],
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        if (_expireType == 'days')
          Row(
            children: [
              const Text(
                '天数',
                style: TextStyle(fontSize: 13, color: Color(0xFF20253A)),
              ),
              const SizedBox(width: 12),
              Container(
                decoration: BoxDecoration(
                  color: const Color(0xFFF7F8FA),
                  borderRadius: BorderRadius.circular(18),
                ),
                child: Row(
                  children: [
                    IconButton(
                      onPressed: _expireDays > 1
                          ? () {
                              setState(() {
                                _expireDays--;
                              });
                            }
                          : null,
                      icon: const Icon(Icons.remove, size: 18),
                      constraints: const BoxConstraints(
                        minWidth: 32,
                        minHeight: 32,
                      ),
                      padding: EdgeInsets.zero,
                    ),
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 8),
                      child: Text(
                        '$_expireDays 天',
                        style: const TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                    IconButton(
                      onPressed: () {
                        setState(() {
                          _expireDays++;
                        });
                      },
                      icon: const Icon(Icons.add, size: 18),
                      constraints: const BoxConstraints(
                        minWidth: 32,
                        minHeight: 32,
                      ),
                      padding: EdgeInsets.zero,
                    ),
                  ],
                ),
              ),
            ],
          )
        else if (_expireType == 'date')
          GestureDetector(
            onTap: () async {
              final now = DateTime.now();
              final date = await showDatePicker(
                context: context,
                initialDate: _expireDateTime ?? now,
                firstDate: now,
                lastDate: now.add(const Duration(days: 365 * 2)),
              );
              if (date == null) return;
              final time = await showTimePicker(
                context: context,
                initialTime: TimeOfDay.fromDateTime(_expireDateTime ?? now),
              );
              DateTime result = date;
              if (time != null) {
                result = DateTime(
                  date.year,
                  date.month,
                  date.day,
                  time.hour,
                  time.minute,
                );
              }
              setState(() {
                _expireDateTime = result;
              });
            },
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              decoration: BoxDecoration(
                color: const Color(0xFFF7F8FA),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Row(
                children: [
                  const Icon(Icons.event, size: 18, color: Color(0xFF8C92A4)),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _expireDateTime == null
                          ? '请选择过期日期时间'
                          : DateFormat('yyyy-MM-dd HH:mm')
                              .format(_expireDateTime!),
                      style: TextStyle(
                        fontSize: 13,
                        color: _expireDateTime == null
                            ? const Color(0xFF8C92A4)
                            : const Color(0xFF20253A),
                      ),
                    ),
                  ),
                  const Icon(
                    Icons.arrow_drop_down,
                    size: 20,
                    color: Color(0xFF8C92A4),
                  ),
                ],
              ),
            ),
          )
        else
          const Text(
            '不限制表示使用优惠券本身的有效期。',
            style: TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
          ),
      ],
    );
  }

  /// 发放原因选择
  Widget _buildReasonSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          '发放原因',
          style: TextStyle(fontSize: 13, color: Color(0xFF20253A)),
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: _reasons.map((reason) {
            final selected = _selectedReason == reason;
            return ChoiceChip(
              label: Text(
                reason,
                style: TextStyle(
                  fontSize: 13,
                  color: selected ? Colors.white : const Color(0xFF40475C),
                ),
              ),
              selected: selected,
              selectedColor: const Color(0xFF20CB6B),
              backgroundColor: const Color(0xFFF7F8FA),
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
              onSelected: (_) {
                setState(() {
                  _selectedReason = reason;
                });
              },
            );
          }).toList(),
        ),
      ],
    );
  }

  Widget _buildBottomButton() {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
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
      child: SizedBox(
        width: double.infinity,
        child: ElevatedButton(
          onPressed: _isIssuing ? null : _issueCoupon,
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFF20CB6B),
            foregroundColor: Colors.white,
            padding: const EdgeInsets.symmetric(vertical: 12),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(24),
            ),
            elevation: 0,
          ),
          child: _isIssuing
              ? const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                )
              : const Text(
                  '确认发放',
                  style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
                ),
        ),
      ),
    );
  }

  String _formatTime(String raw) {
    try {
      final dt = DateTime.tryParse(raw);
      if (dt == null) return raw;
      return DateFormat('yyyy-MM-dd HH:mm').format(dt.toLocal());
    } catch (_) {
      return raw;
    }
  }
}
