import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:screenshot/screenshot.dart';
import 'package:gal/gal.dart';
import 'package:path_provider/path_provider.dart';
import 'package:share_plus/share_plus.dart';
import 'dart:io';

/// 订单预览页面（用于生成销售单图片）
class OrderPreviewPage extends StatefulWidget {
  final Map<String, dynamic> order;
  final Map<String, dynamic>? user;
  final Map<String, dynamic>? address;
  final List<dynamic> items;

  const OrderPreviewPage({
    super.key,
    required this.order,
    this.user,
    this.address,
    required this.items,
  });

  @override
  State<OrderPreviewPage> createState() => _OrderPreviewPageState();
}

class _OrderPreviewPageState extends State<OrderPreviewPage> {
  final ScreenshotController _screenshotController = ScreenshotController();
  bool _isSaving = false; // 不通过 setState 更新，仅用于防止重复点击

  /// 根据订单状态获取头部提示文案
  String _getStatusMessage() {
    final status = widget.order['status']?.toString() ?? '';

    // 待取货之前（pending_delivery, pending）
    if (status == 'pending_delivery' || status == 'pending') {
      return '您的订单已下单，请及时付款！';
    }
    // 配送员取货状态到配送完成前（pending_pickup, delivering）
    else if (status == 'pending_pickup' || status == 'delivering') {
      return '订单配送中，请及时付款！';
    }
    // 配送完成后（delivered, shipped）
    else if (status == 'delivered' || status == 'shipped') {
      return '订单已配送完成，请及时付款！';
    }
    // 配送完成订单（paid, completed）
    else if (status == 'paid' || status == 'completed') {
      return '订单已配送完成！';
    }

    return '您的订单信息';
  }

  String _formatDateTime(dynamic raw) {
    if (raw == null) return '';
    try {
      final dt = raw is DateTime ? raw : DateTime.tryParse(raw.toString());
      if (dt == null) return raw.toString();
      return DateFormat('yyyy-MM-dd HH:mm').format(dt.toLocal());
    } catch (_) {
      return raw.toString();
    }
  }

  /// 保存图片到相册
  Future<void> _saveImage() async {
    // 防止重复点击
    if (_isSaving) return;
    if (!mounted) return;

    // 标记为正在保存，不调用 setState
    _isSaving = true;

    try {
      // 等待一小段时间确保UI渲染完成
      await Future.delayed(const Duration(milliseconds: 200));
      if (!mounted) {
        _isSaving = false;
        return;
      }

      // 捕获截图
      final Uint8List? imageBytes = await _screenshotController.capture(
        delay: const Duration(milliseconds: 100),
        pixelRatio: 2.0,
      );

      if (!mounted || imageBytes == null) {
        _isSaving = false;
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('截图失败，请重试'),
              backgroundColor: Colors.orange,
            ),
          );
        }
        return;
      }

      // 获取应用文档目录
      final directory = await getApplicationDocumentsDirectory();
      final timestamp = DateTime.now().millisecondsSinceEpoch;
      final fileName =
          '订单_${widget.order['order_number'] ?? timestamp}_$timestamp.png';
      final filePath = '${directory.path}/$fileName';

      // 保存文件
      final file = File(filePath);
      await file.writeAsBytes(imageBytes);

      // 保存到相册
      await Gal.putImage(filePath);

      if (!mounted) {
        _isSaving = false;
        return;
      }

      // 重置状态
      _isSaving = false;

      // 延迟一小段时间后打开分享对话框，避免状态冲突
      await Future.delayed(const Duration(milliseconds: 100));

      if (!mounted) return;

      // 打开分享对话框
      final orderNumber = widget.order['order_number']?.toString() ?? '订单';
      Share.shareXFiles([
        XFile(filePath),
      ], text: '${orderNumber}订单详情').catchError((error) {
        print('分享失败: $error');
        return error; // 返回错误对象
      });
    } catch (e) {
      _isSaving = false;
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('保存失败: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final order = widget.order;
    final address = widget.address ?? {};
    final orderNumber = order['order_number']?.toString() ?? '';
    final createdAt = _formatDateTime(order['created_at']);

    final addrName = (address['name'] as String?) ?? '';
    final addrText = (address['address'] as String?) ?? '';

    final pointsDiscount =
        (order['points_discount'] as num?)?.toDouble() ?? 0.0;
    final couponDiscount =
        (order['coupon_discount'] as num?)?.toDouble() ?? 0.0;

    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: const Text('订单预览'),
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
      body: Column(
        children: [
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Screenshot(
                controller: _screenshotController,
                child: Container(
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(12),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.05),
                        blurRadius: 10,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // 头部状态提示（绿色渐变背景）
                      Container(
                        width: double.infinity,
                        padding: const EdgeInsets.all(20),
                        decoration: const BoxDecoration(
                          gradient: LinearGradient(
                            begin: Alignment.topLeft,
                            end: Alignment.bottomRight,
                            colors: [Color(0xFF20CB6B), Color(0xFF18B85A)],
                          ),
                          borderRadius: BorderRadius.only(
                            topLeft: Radius.circular(12),
                            topRight: Radius.circular(12),
                          ),
                        ),
                        child: Column(
                          children: [
                            Text(
                              _getStatusMessage(),
                              style: const TextStyle(
                                fontSize: 18,
                                fontWeight: FontWeight.w600,
                                color: Colors.white,
                              ),
                            ),
                            if (orderNumber.isNotEmpty) ...[
                              const SizedBox(height: 8),
                              Text(
                                '订单号：$orderNumber',
                                style: TextStyle(
                                  fontSize: 14,
                                  color: Colors.white.withOpacity(0.9),
                                ),
                              ),
                            ],
                            if (createdAt.isNotEmpty) ...[
                              const SizedBox(height: 4),
                              Text(
                                '下单时间：$createdAt',
                                style: TextStyle(
                                  fontSize: 12,
                                  color: Colors.white.withOpacity(0.8),
                                ),
                              ),
                            ],
                          ],
                        ),
                      ),

                      // 地址信息
                      Padding(
                        padding: const EdgeInsets.all(20),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Container(
                                  padding: const EdgeInsets.all(8),
                                  decoration: BoxDecoration(
                                    color: const Color(
                                      0xFF20CB6B,
                                    ).withOpacity(0.1),
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: const Icon(
                                    Icons.location_on,
                                    size: 20,
                                    color: Color(0xFF20CB6B),
                                  ),
                                ),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        addrName.isNotEmpty ? addrName : '收货地址',
                                        style: const TextStyle(
                                          fontSize: 16,
                                          fontWeight: FontWeight.w600,
                                          color: Color(0xFF20253A),
                                        ),
                                      ),
                                      if (addrText.isNotEmpty) ...[
                                        const SizedBox(height: 4),
                                        Text(
                                          addrText,
                                          style: const TextStyle(
                                            fontSize: 14,
                                            color: Color(0xFF40475C),
                                            height: 1.4,
                                          ),
                                        ),
                                      ],
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),

                      const Divider(height: 1, color: Color(0xFFE5E7F0)),

                      // 商品列表
                      Padding(
                        padding: const EdgeInsets.all(20),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text(
                              '商品明细',
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.w600,
                                color: Color(0xFF20253A),
                              ),
                            ),
                            const SizedBox(height: 16),
                            if (widget.items.isEmpty)
                              const Padding(
                                padding: EdgeInsets.symmetric(vertical: 20),
                                child: Center(
                                  child: Text(
                                    '暂无商品',
                                    style: TextStyle(
                                      fontSize: 14,
                                      color: Color(0xFF8C92A4),
                                    ),
                                  ),
                                ),
                              )
                            else
                              ...widget.items.asMap().entries.map((entry) {
                                final index = entry.key;
                                final raw = entry.value;
                                final item = raw as Map<String, dynamic>;
                                final name =
                                    (item['product_name'] as String?) ?? '';
                                final spec =
                                    (item['spec_name'] as String?) ?? '';
                                final qty = (item['quantity'] as int?) ?? 0;
                                final unitPrice =
                                    (item['unit_price'] as num?)?.toDouble() ??
                                    0.0;
                                final subtotal =
                                    (item['subtotal'] as num?)?.toDouble() ??
                                    0.0;
                                final image =
                                    (item['product_image'] as String?) ??
                                    (item['image'] as String?) ??
                                    '';

                                // 改价信息
                                final originalUnitPrice =
                                    (item['original_unit_price'] as num?)
                                        ?.toDouble();
                                final isPriceModified =
                                    (item['is_price_modified'] as bool?) ??
                                    false;

                                return Container(
                                  margin: EdgeInsets.only(
                                    bottom: index < widget.items.length - 1
                                        ? 16
                                        : 0,
                                  ),
                                  child: Row(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      // 商品图片
                                      Container(
                                        width: 60,
                                        height: 60,
                                        margin: const EdgeInsets.only(
                                          right: 12,
                                        ),
                                        decoration: BoxDecoration(
                                          color: const Color(0xFFF5F6FA),
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                        ),
                                        clipBehavior: Clip.antiAlias,
                                        child: image.isNotEmpty
                                            ? Image.network(
                                                image,
                                                fit: BoxFit.cover,
                                                errorBuilder:
                                                    (
                                                      context,
                                                      error,
                                                      stackTrace,
                                                    ) {
                                                      return const Icon(
                                                        Icons
                                                            .image_not_supported,
                                                        color: Color(
                                                          0xFFB0B4C3,
                                                        ),
                                                        size: 24,
                                                      );
                                                    },
                                              )
                                            : const Icon(
                                                Icons.image,
                                                color: Color(0xFFB0B4C3),
                                                size: 24,
                                              ),
                                      ),
                                      Expanded(
                                        child: Column(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.start,
                                          children: [
                                            Text(
                                              name,
                                              style: const TextStyle(
                                                fontSize: 15,
                                                fontWeight: FontWeight.w600,
                                                color: Color(0xFF20253A),
                                              ),
                                            ),
                                            if (spec.isNotEmpty) ...[
                                              const SizedBox(height: 4),
                                              Text(
                                                spec,
                                                style: const TextStyle(
                                                  fontSize: 13,
                                                  color: Color(0xFF8C92A4),
                                                ),
                                              ),
                                            ],
                                            const SizedBox(height: 6),
                                            Row(
                                              children: [
                                                Row(
                                                  mainAxisSize:
                                                      MainAxisSize.min,
                                                  children: [
                                                    if (isPriceModified &&
                                                        originalUnitPrice !=
                                                            null &&
                                                        unitPrice <
                                                            originalUnitPrice) ...[
                                                      Text(
                                                        '¥${originalUnitPrice.toStringAsFixed(2)}',
                                                        style: TextStyle(
                                                          fontSize: 11,
                                                          color:
                                                              Colors.grey[400],
                                                          decoration:
                                                              TextDecoration
                                                                  .lineThrough,
                                                        ),
                                                      ),
                                                      const SizedBox(width: 4),
                                                    ],
                                                    Text(
                                                      '¥${unitPrice.toStringAsFixed(2)}',
                                                      style: TextStyle(
                                                        fontSize: 14,
                                                        fontWeight:
                                                            FontWeight.w600,
                                                        color: isPriceModified
                                                            ? const Color(
                                                                0xFFFF5A5F,
                                                              )
                                                            : const Color(
                                                                0xFF20CB6B,
                                                              ),
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                                const SizedBox(width: 12),
                                                Text(
                                                  'x$qty',
                                                  style: const TextStyle(
                                                    fontSize: 13,
                                                    color: Color(0xFF8C92A4),
                                                  ),
                                                ),
                                              ],
                                            ),
                                          ],
                                        ),
                                      ),
                                      Text(
                                        '¥${subtotal.toStringAsFixed(2)}',
                                        style: const TextStyle(
                                          fontSize: 16,
                                          fontWeight: FontWeight.w600,
                                          color: Color(0xFF20253A),
                                        ),
                                      ),
                                    ],
                                  ),
                                );
                              }),
                          ],
                        ),
                      ),

                      const Divider(height: 1, color: Color(0xFFE5E7F0)),

                      // 金额信息
                      Padding(
                        padding: const EdgeInsets.all(20),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text(
                              '金额信息',
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.w600,
                                color: Color(0xFF20253A),
                              ),
                            ),
                            const SizedBox(height: 16),
                            _buildAmountRow(
                              '商品金额',
                              (order['goods_amount'] as num?)?.toDouble() ??
                                  0.0,
                            ),
                            const SizedBox(height: 10),
                            _buildAmountRow(
                              '配送费用',
                              (order['delivery_fee'] as num?)?.toDouble() ??
                                  0.0,
                              isHighlight: true,
                            ),
                            if ((order['is_urgent'] as bool?) ?? false) ...[
                              const SizedBox(height: 10),
                              _buildAmountRow(
                                '加急费用',
                                (order['urgent_fee'] as num?)?.toDouble() ??
                                    0.0,
                                isHighlight: true,
                              ),
                            ],
                            if (pointsDiscount > 0) ...[
                              const SizedBox(height: 10),
                              _buildAmountRow(
                                '积分抵扣',
                                -pointsDiscount,
                                isDiscount: true,
                              ),
                            ],
                            if (couponDiscount > 0) ...[
                              const SizedBox(height: 10),
                              _buildAmountRow(
                                '优惠券抵扣',
                                -couponDiscount,
                                isDiscount: true,
                              ),
                            ],
                            const SizedBox(height: 16),
                            const Divider(height: 1, color: Color(0xFFE5E7F0)),
                            const SizedBox(height: 16),
                            Row(
                              children: [
                                const Text(
                                  '实付金额',
                                  style: TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.w600,
                                    color: Color(0xFF20253A),
                                  ),
                                ),
                                const Spacer(),
                                Text(
                                  '¥${((order['total_amount'] as num?)?.toDouble() ?? 0.0).toStringAsFixed(2)}',
                                  style: const TextStyle(
                                    fontSize: 24,
                                    fontWeight: FontWeight.bold,
                                    color: Color(0xFF20CB6B),
                                  ),
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
            ),
          ),

          // 底部保存按钮
          Container(
            padding: EdgeInsets.fromLTRB(
              16,
              12,
              16,
              12 + MediaQuery.of(context).padding.bottom,
            ),
            decoration: BoxDecoration(
              color: Colors.white,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.05),
                  blurRadius: 10,
                  offset: const Offset(0, -2),
                ),
              ],
            ),
            child: SafeArea(
              top: false,
              child: SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _saveImage,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF20CB6B),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    elevation: 0,
                  ),
                  child: _isSaving
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            valueColor: AlwaysStoppedAnimation<Color>(
                              Colors.white,
                            ),
                          ),
                        )
                      : const Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.save, size: 20),
                            SizedBox(width: 8),
                            Text(
                              '保存并分享',
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ],
                        ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAmountRow(
    String label,
    double value, {
    bool isHighlight = false,
    bool isDiscount = false,
  }) {
    final isNegative = value < 0;
    final display = value.abs();

    Color labelColor;
    Color valueColor;

    if (isDiscount) {
      // 优惠项：绿色
      labelColor = const Color(0xFF20CB6B);
      valueColor = const Color(0xFF20CB6B);
    } else if (isHighlight) {
      // 费用项：橙色
      labelColor = const Color(0xFFFF9800);
      valueColor = const Color(0xFFFF9800);
    } else {
      // 普通项
      labelColor = const Color(0xFF40475C);
      valueColor = const Color(0xFF40475C);
    }

    return Row(
      children: [
        Text(
          label,
          style: TextStyle(
            fontSize: 15,
            color: labelColor,
            fontWeight: FontWeight.w500,
          ),
        ),
        const Spacer(),
        Text(
          '${isNegative ? '-' : ''}¥${display.toStringAsFixed(2)}',
          style: TextStyle(
            fontSize: 15,
            color: valueColor,
            fontWeight: FontWeight.w600,
          ),
        ),
      ],
    );
  }
}
