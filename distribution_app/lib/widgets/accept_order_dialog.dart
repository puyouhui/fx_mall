import 'package:flutter/material.dart';

/// 接单确认对话框
class AcceptOrderDialog extends StatelessWidget {
  final String storeName;
  final String address;
  final double riderPayableFee;
  final double totalAmount;
  final int itemCount;
  final bool isUrgent;

  const AcceptOrderDialog({
    super.key,
    required this.storeName,
    required this.address,
    required this.riderPayableFee,
    required this.totalAmount,
    required this.itemCount,
    this.isUrgent = false,
  });

  static Future<bool?> show(
    BuildContext context, {
    required String storeName,
    required String address,
    required double riderPayableFee,
    required double totalAmount,
    required int itemCount,
    bool isUrgent = false,
  }) {
    return showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (context) => AcceptOrderDialog(
        storeName: storeName,
        address: address,
        riderPayableFee: riderPayableFee,
        totalAmount: totalAmount,
        itemCount: itemCount,
        isUrgent: isUrgent,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: Container(
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(20),
          color: Colors.white,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 标题
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: const Color(0xFF20CB6B).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Icon(
                    Icons.info_outline,
                    color: Color(0xFF20CB6B),
                    size: 24,
                  ),
                ),
                const SizedBox(width: 12),
                const Expanded(
                  child: Text(
                    '接单提示',
                    style: TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.w700,
                      color: Color(0xFF20253A),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 20),
            // 重要提示
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: const Color(0xFFFF9500).withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: const Color(0xFFFF9500).withOpacity(0.3),
                  width: 1,
                ),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(
                    Icons.warning_amber_rounded,
                    color: Color(0xFFFF9500),
                    size: 20,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '接单后不可取消，请务必尽快送达',
                      style: TextStyle(
                        fontSize: 13,
                        color: const Color(0xFF20253A).withOpacity(0.8),
                        height: 1.4,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 20),
            // 信息核对标题
            const Text(
              '信息核对',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: Color(0xFF20253A),
              ),
            ),
            const SizedBox(height: 12),
            // 地址信息
            _buildInfoRow(
              icon: Icons.location_on,
              iconColor: const Color(0xFF20CB6B),
              label: '收货地址',
              value: storeName,
              value2: address,
            ),
            const SizedBox(height: 12),
            // 商品数量
            _buildInfoRow(
              icon: Icons.shopping_cart,
              iconColor: const Color(0xFF8C92A4),
              label: '商品数量',
              value: '$itemCount 件',
            ),
            const SizedBox(height: 12),
            // 配送费
            _buildInfoRow(
              icon: Icons.local_shipping,
              iconColor: const Color(0xFF20CB6B),
              label: '配送费',
              value: '¥${riderPayableFee.toStringAsFixed(2)}',
              isHighlight: true,
            ),
            if (isUrgent) ...[
              const SizedBox(height: 12),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFFFF6B6B).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(
                      Icons.flash_on,
                      size: 16,
                      color: Color(0xFFFF6B6B),
                    ),
                    const SizedBox(width: 4),
                    const Text(
                      '加急订单',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFFFF6B6B),
                      ),
                    ),
                  ],
                ),
              ),
            ],
            const SizedBox(height: 24),
            // 按钮
            Row(
              children: [
                Expanded(
                  child: OutlinedButton(
                    onPressed: () => Navigator.of(context).pop(false),
                    style: OutlinedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      side: const BorderSide(
                        color: Color(0xFFE5E7EB),
                        width: 1,
                      ),
                    ),
                    child: const Text(
                      '取消',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF8C92A4),
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  flex: 2,
                  child: ElevatedButton(
                    onPressed: () => Navigator.of(context).pop(true),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF20CB6B),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      elevation: 0,
                    ),
                    child: const Text(
                      '确认接单',
                      style: TextStyle(
                        fontSize: 16,
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
    );
  }

  Widget _buildInfoRow({
    required IconData icon,
    required Color iconColor,
    required String label,
    required String value,
    String? value2,
    bool isHighlight = false,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 18, color: iconColor),
        const SizedBox(width: 10),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: TextStyle(fontSize: 13, color: const Color(0xFF8C92A4)),
              ),
              const SizedBox(height: 4),
              Text(
                value,
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: isHighlight ? FontWeight.w700 : FontWeight.w500,
                  color: isHighlight
                      ? const Color(0xFF20CB6B)
                      : const Color(0xFF20253A),
                ),
              ),
              if (value2 != null && value2.isNotEmpty) ...[
                const SizedBox(height: 2),
                Text(
                  value2,
                  style: TextStyle(
                    fontSize: 13,
                    color: const Color(0xFF40475C).withOpacity(0.8),
                    height: 1.3,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }
}
