/// 优惠券选择页面
import 'package:flutter/material.dart';

class CouponSelectionPage extends StatefulWidget {
  final int customerId;
  final List<dynamic> coupons;
  final Map<String, dynamic>? selectedCoupon;
  final double orderAmount;

  const CouponSelectionPage({
    super.key,
    required this.customerId,
    required this.coupons,
    this.selectedCoupon,
    required this.orderAmount,
  });

  @override
  State<CouponSelectionPage> createState() => _CouponSelectionPageState();
}

class _CouponSelectionPageState extends State<CouponSelectionPage> {
  Map<String, dynamic>? _selectedCoupon;
  int _currentTab = 0; // 0: 可用, 1: 已使用/过期

  @override
  void initState() {
    super.initState();
    _selectedCoupon = widget.selectedCoupon;
  }

  /// 判断优惠券是否可用（未使用且未过期）
  bool _isCouponUnusedAndValid(Map<String, dynamic> coupon) {
    final status = coupon['status'] as String?;
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
  }

  /// 获取可用优惠券列表
  List<Map<String, dynamic>> _getAvailableCoupons() {
    return widget.coupons
        .where((raw) {
          final coupon = raw as Map<String, dynamic>;
          return _isCouponUnusedAndValid(coupon);
        })
        .map((raw) => raw as Map<String, dynamic>)
        .toList();
  }

  /// 获取已使用/过期优惠券列表
  List<Map<String, dynamic>> _getUsedOrExpiredCoupons() {
    return widget.coupons
        .where((raw) {
          final coupon = raw as Map<String, dynamic>;
          return !_isCouponUnusedAndValid(coupon);
        })
        .map((raw) => raw as Map<String, dynamic>)
        .toList();
  }

  /// 检查优惠券是否可用
  bool _isCouponAvailable(Map<String, dynamic> coupon) {
    // 检查优惠券状态
    final status = coupon['status'] as String?;
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
        // 解析失败，继续检查其他条件
      }
    }

    // 如果传入的是 userCoupon，需要从 coupon 字段获取优惠券信息
    Map<String, dynamic> couponData = coupon;
    if (coupon['coupon'] != null) {
      couponData = coupon['coupon'] as Map<String, dynamic>;
    }

    // 检查优惠券本身的有效期（valid_to）
    final validToStr = couponData['valid_to'] as String?;
    if (validToStr != null && validToStr.isNotEmpty) {
      try {
        final validTo = DateTime.parse(validToStr);
        final now = DateTime.now();
        if (now.isAfter(validTo)) {
          return false; // 优惠券本身已过期
        }
      } catch (e) {
        // 解析失败，继续检查其他条件
      }
    }

    // 检查最低金额要求
    final minAmount = (couponData['min_amount'] as num?)?.toDouble() ?? 0.0;
    if (minAmount > 0 && widget.orderAmount < minAmount) {
      return false;
    }
    return true;
  }

  /// 获取优惠券描述
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
  String _getDiscountAmountText(Map<String, dynamic> coupon) {
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
      return '¥${discountValue.toStringAsFixed(2)}';
    }
    return '';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('选择优惠券'),
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
          child: widget.coupons.isEmpty
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
                          Icons.local_offer_outlined,
                          size: 64,
                          color: Color(0xFF8C92A4),
                        ),
                        const SizedBox(height: 16),
                        const Text(
                          '暂无优惠券',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                      ],
                    ),
                  ),
                )
              : Column(
                  children: [
                    // Tab切换
                    Container(
                      margin: const EdgeInsets.fromLTRB(16, 16, 16, 12),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(12),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withOpacity(0.08),
                            blurRadius: 10,
                            offset: const Offset(0, 2),
                          ),
                        ],
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: InkWell(
                              onTap: () {
                                setState(() {
                                  _currentTab = 0;
                                });
                              },
                              borderRadius: const BorderRadius.only(
                                topLeft: Radius.circular(12),
                                bottomLeft: Radius.circular(12),
                              ),
                              child: Container(
                                padding: const EdgeInsets.symmetric(
                                  vertical: 14,
                                ),
                                child: Center(
                                  child: Column(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Text(
                                        '可用(${_getAvailableCoupons().length})',
                                        style: TextStyle(
                                          fontSize: 15,
                                          fontWeight: FontWeight.w600,
                                          color: _currentTab == 0
                                              ? const Color(0xFF20CB6B)
                                              : const Color(0xFF8C92A4),
                                        ),
                                      ),
                                      const SizedBox(height: 6),
                                      if (_currentTab == 0)
                                        Container(
                                          width: 40,
                                          height: 3,
                                          decoration: BoxDecoration(
                                            color: const Color(0xFF20CB6B),
                                            borderRadius: BorderRadius.circular(
                                              2,
                                            ),
                                          ),
                                        )
                                      else
                                        const SizedBox(height: 3),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                          ),
                          Expanded(
                            child: InkWell(
                              onTap: () {
                                setState(() {
                                  _currentTab = 1;
                                });
                              },
                              borderRadius: const BorderRadius.only(
                                topRight: Radius.circular(12),
                                bottomRight: Radius.circular(12),
                              ),
                              child: Container(
                                padding: const EdgeInsets.symmetric(
                                  vertical: 14,
                                ),
                                child: Center(
                                  child: Column(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Text(
                                        '已使用/过期(${_getUsedOrExpiredCoupons().length})',
                                        style: TextStyle(
                                          fontSize: 15,
                                          fontWeight: FontWeight.w600,
                                          color: _currentTab == 1
                                              ? const Color(0xFF20CB6B)
                                              : const Color(0xFF8C92A4),
                                        ),
                                      ),
                                      const SizedBox(height: 6),
                                      if (_currentTab == 1)
                                        Container(
                                          width: 40,
                                          height: 3,
                                          decoration: BoxDecoration(
                                            color: const Color(0xFF20CB6B),
                                            borderRadius: BorderRadius.circular(
                                              2,
                                            ),
                                          ),
                                        )
                                      else
                                        const SizedBox(height: 3),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                    // 优惠券列表
                    Expanded(
                      child: ListView(
                        padding: const EdgeInsets.all(16),
                        children: [
                          // 不使用优惠券选项（仅在可用tab显示）
                          if (_currentTab == 0 && _selectedCoupon != null)
                            Container(
                              margin: const EdgeInsets.only(bottom: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(16),
                                border: Border.all(
                                  color: const Color(0xFFE5E7F0),
                                  width: 1,
                                ),
                              ),
                              child: InkWell(
                                onTap: () {
                                  Navigator.of(
                                    context,
                                  ).pop<Map<String, dynamic>>({'remove': true});
                                },
                                borderRadius: BorderRadius.circular(16),
                                child: Padding(
                                  padding: const EdgeInsets.all(16),
                                  child: Row(
                                    children: [
                                      Container(
                                        width: 24,
                                        height: 24,
                                        margin: const EdgeInsets.only(
                                          right: 12,
                                        ),
                                        decoration: BoxDecoration(
                                          shape: BoxShape.circle,
                                          border: Border.all(
                                            color: const Color(0xFFE5E7F0),
                                            width: 2,
                                          ),
                                        ),
                                      ),
                                      const Expanded(
                                        child: Text(
                                          '不使用优惠券',
                                          style: TextStyle(
                                            fontSize: 15,
                                            fontWeight: FontWeight.w600,
                                            color: Color(0xFF20253A),
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                          ...(_currentTab == 0
                                  ? _getAvailableCoupons()
                                  : _getUsedOrExpiredCoupons())
                              .map((userCoupon) {
                                // 优惠券信息在 coupon 字段中
                                final couponData =
                                    userCoupon['coupon']
                                        as Map<String, dynamic>?;
                                if (couponData == null)
                                  return const SizedBox.shrink();

                                // 使用优惠券数据
                                final coupon = couponData;
                                final id = coupon['id'] as int?;
                                final name = coupon['name'] as String? ?? '优惠券';
                                final isAvailable = _currentTab == 0
                                    ? _isCouponAvailable(userCoupon)
                                    : false;

                                // 检查选中状态：比较 userCoupon 的 id
                                final userCouponId = userCoupon['id'] as int?;
                                bool isSelected = false;
                                if (_currentTab == 0 &&
                                    _selectedCoupon != null &&
                                    userCouponId != null) {
                                  final selectedId =
                                      _selectedCoupon!['id'] as int?;
                                  isSelected = selectedId == userCouponId;
                                }

                                // 获取状态文本
                                final status = userCoupon['status'] as String?;
                                String statusText = '';
                                Color statusColor = const Color(0xFF8C92A4);
                                if (status == 'used') {
                                  statusText = '已使用';
                                  statusColor = const Color(0xFF8C92A4);
                                } else if (status == 'expired') {
                                  statusText = '已过期';
                                  statusColor = const Color(0xFFFF5722);
                                } else {
                                  // 检查是否过期
                                  final expiresAtStr =
                                      userCoupon['expires_at'] as String?;
                                  if (expiresAtStr != null &&
                                      expiresAtStr.isNotEmpty) {
                                    try {
                                      final expiresAt = DateTime.parse(
                                        expiresAtStr,
                                      );
                                      final now = DateTime.now();
                                      if (now.isAfter(expiresAt)) {
                                        statusText = '已过期';
                                        statusColor = const Color(0xFFFF5722);
                                      }
                                    } catch (e) {
                                      // 解析失败
                                    }
                                  }
                                }

                                if (id == null) return const SizedBox.shrink();

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
                                    onTap: (_currentTab == 0 && isAvailable)
                                        ? () {
                                            setState(() {
                                              _selectedCoupon =
                                                  userCoupon; // 保存完整的 userCoupon，包含 coupon 信息
                                            });
                                            Navigator.of(
                                              context,
                                            ).pop<Map<String, dynamic>>({
                                              'coupon':
                                                  userCoupon, // 返回完整的 userCoupon
                                            });
                                          }
                                        : null,
                                    borderRadius: BorderRadius.circular(16),
                                    child: Opacity(
                                      opacity: (_currentTab == 0 && isAvailable)
                                          ? 1.0
                                          : 0.5,
                                      child: Padding(
                                        padding: const EdgeInsets.all(16),
                                        child: Row(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.start,
                                          children: [
                                            // 选中图标
                                            Container(
                                              width: 24,
                                              height: 24,
                                              margin: const EdgeInsets.only(
                                                right: 12,
                                              ),
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
                                            // 优惠券信息
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
                                                          style:
                                                              const TextStyle(
                                                                fontSize: 15,
                                                                fontWeight:
                                                                    FontWeight
                                                                        .w600,
                                                                color: Color(
                                                                  0xFF20253A,
                                                                ),
                                                              ),
                                                        ),
                                                      ),
                                                      // 优惠金额显示
                                                      Container(
                                                        padding:
                                                            const EdgeInsets.symmetric(
                                                              horizontal: 8,
                                                              vertical: 4,
                                                            ),
                                                        decoration: BoxDecoration(
                                                          color: const Color(
                                                            0xFFFFF4E6,
                                                          ).withOpacity(0.8),
                                                          borderRadius:
                                                              BorderRadius.circular(
                                                                8,
                                                              ),
                                                        ),
                                                        child: Text(
                                                          _getDiscountAmountText(
                                                            coupon,
                                                          ),
                                                          style:
                                                              const TextStyle(
                                                                fontSize: 14,
                                                                fontWeight:
                                                                    FontWeight
                                                                        .w700,
                                                                color: Color(
                                                                  0xFFFF9800,
                                                                ),
                                                              ),
                                                        ),
                                                      ),
                                                      if (!isAvailable &&
                                                          _currentTab == 0)
                                                        Container(
                                                          margin:
                                                              const EdgeInsets.only(
                                                                left: 6,
                                                              ),
                                                          padding:
                                                              const EdgeInsets.symmetric(
                                                                horizontal: 6,
                                                                vertical: 2,
                                                              ),
                                                          decoration: BoxDecoration(
                                                            color: const Color(
                                                              0xFFF5F6FA,
                                                            ),
                                                            borderRadius:
                                                                BorderRadius.circular(
                                                                  10,
                                                                ),
                                                          ),
                                                          child: const Text(
                                                            '不满足条件',
                                                            style: TextStyle(
                                                              fontSize: 10,
                                                              color: Color(
                                                                0xFF8C92A4,
                                                              ),
                                                            ),
                                                          ),
                                                        )
                                                      else if (_currentTab ==
                                                              1 &&
                                                          statusText.isNotEmpty)
                                                        Container(
                                                          margin:
                                                              const EdgeInsets.only(
                                                                left: 6,
                                                              ),
                                                          padding:
                                                              const EdgeInsets.symmetric(
                                                                horizontal: 6,
                                                                vertical: 2,
                                                              ),
                                                          decoration: BoxDecoration(
                                                            color: statusColor
                                                                .withOpacity(
                                                                  0.1,
                                                                ),
                                                            borderRadius:
                                                                BorderRadius.circular(
                                                                  10,
                                                                ),
                                                          ),
                                                          child: Text(
                                                            statusText,
                                                            style: TextStyle(
                                                              fontSize: 10,
                                                              color:
                                                                  statusColor,
                                                            ),
                                                          ),
                                                        ),
                                                    ],
                                                  ),
                                                  const SizedBox(height: 6),
                                                  Text(
                                                    _getCouponDescription(
                                                      coupon,
                                                    ),
                                                    style: const TextStyle(
                                                      fontSize: 13,
                                                      color: Color(0xFF40475C),
                                                    ),
                                                  ),
                                                  const SizedBox(height: 4),
                                                  Text(
                                                    _formatValidPeriod(coupon),
                                                    style: const TextStyle(
                                                      fontSize: 11,
                                                      color: Color(0xFF8C92A4),
                                                    ),
                                                  ),
                                                ],
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ),
                                  ),
                                );
                              })
                              .toList(),
                        ],
                      ),
                    ),
                  ],
                ),
        ),
      ),
    );
  }
}
