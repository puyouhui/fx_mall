import 'package:flutter/material.dart';

/// 接单大厅页（示例版：展示可接单列表占位）
class OrderHallView extends StatelessWidget {
  const OrderHallView({super.key});

  @override
  Widget build(BuildContext context) {
    // 这里先用假数据占位，后续接入真实接口
    final mockOrders = List.generate(10, (index) {
      return '订单 #${1000 + index}';
    });

    return ListView.separated(
      padding: const EdgeInsets.all(16),
      itemBuilder: (context, index) {
        final orderNo = mockOrders[index];
        return Card(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          child: ListTile(
            leading: const Icon(Icons.local_shipping_outlined),
            title: Text(orderNo),
            subtitle: const Text('示例门店 · 示例收货地址'),
            trailing: const Text(
              '待接单',
              style: TextStyle(
                color: Colors.orange,
                fontWeight: FontWeight.w500,
              ),
            ),
            onTap: () {
              // TODO: 进入订单详情或选择订单逻辑
            },
          ),
        );
      },
      separatorBuilder: (_, __) => const SizedBox(height: 12),
      itemCount: mockOrders.length,
    );
  }
}


