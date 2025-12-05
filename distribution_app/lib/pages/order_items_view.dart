import 'package:flutter/material.dart';

/// 订单商品列表页面：显示订单的所有商品，方便配送员规划车辆空间
class OrderItemsView extends StatelessWidget {
  const OrderItemsView({
    super.key,
    required this.orderId,
    required this.items,
    this.orderNumber,
  });

  final int orderId;
  final List<Map<String, dynamic>> items;
  final String? orderNumber;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(orderNumber != null ? '订单 $orderNumber' : '商品列表'),
        backgroundColor: const Color(0xFF20CB6B),
        iconTheme: const IconThemeData(color: Colors.white),
        titleTextStyle: const TextStyle(
          color: Colors.white,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      backgroundColor: const Color(0xFFF5F5F5),
      body: items.isEmpty
          ? const Center(
              child: Text(
                '暂无商品信息',
                style: TextStyle(
                  fontSize: 14,
                  color: Color(0xFF8C92A4),
                ),
              ),
            )
          : ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: items.length,
              itemBuilder: (context, index) {
                final item = items[index];
                return _buildItemCard(item);
              },
            ),
    );
  }

  Widget _buildItemCard(Map<String, dynamic> item) {
    final name = item['name'] as String? ?? '商品名称未填写';
    final quantity = (item['quantity'] as num?)?.toInt() ?? 0;
    final unit = item['unit'] as String? ?? '';
    final image = item['image'] as String?;
    final spec = item['spec'] as String?;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 商品图片
          if (image != null && image.isNotEmpty)
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: Image.network(
                image,
                width: 80,
                height: 80,
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) {
                  return Container(
                    width: 80,
                    height: 80,
                    color: const Color(0xFFF5F5F5),
                    child: const Icon(
                      Icons.image_not_supported,
                      color: Color(0xFFD0D0D0),
                    ),
                  );
                },
              ),
            )
          else
            Container(
              width: 80,
              height: 80,
              decoration: BoxDecoration(
                color: const Color(0xFFF5F5F5),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Icon(
                Icons.image_not_supported,
                color: Color(0xFFD0D0D0),
              ),
            ),
          const SizedBox(width: 12),
          // 商品信息
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  style: const TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                if (spec != null && spec.isNotEmpty) ...[
                  const SizedBox(height: 4),
                  Text(
                    spec,
                    style: const TextStyle(
                      fontSize: 12,
                      color: Color(0xFF8C92A4),
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
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
                        '数量：$quantity${unit.isNotEmpty ? unit : '件'}',
                        style: const TextStyle(
                          fontSize: 13,
                          color: Color(0xFF20CB6B),
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

