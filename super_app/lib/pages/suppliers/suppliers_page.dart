import 'package:flutter/material.dart';

class SuppliersPage extends StatefulWidget {
  const SuppliersPage({super.key});

  @override
  State<SuppliersPage> createState() => _SuppliersPageState();
}

class _SuppliersPageState extends State<SuppliersPage> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: RefreshIndicator(
        onRefresh: () async {
          // TODO: 刷新供应商列表
          await Future.delayed(const Duration(seconds: 1));
        },
        child: CustomScrollView(
          slivers: [
            // 顶部操作栏
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      '供应商管理',
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF20253A),
                      ),
                    ),
                    IconButton(
                      onPressed: () {
                        // TODO: 搜索供应商
                      },
                      icon: const Icon(Icons.search),
                      tooltip: '搜索',
                    ),
                  ],
                ),
              ),
            ),
            
            // 供应商列表
            SliverPadding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              sliver: SliverList(
                delegate: SliverChildBuilderDelegate(
                  (context, index) {
                    // TODO: 从API获取供应商列表
                    return _buildSupplierCard(
                      name: '供应商名称 ${index + 1}',
                      contact: '联系人',
                      phone: '13800138000',
                      status: 'active',
                    );
                  },
                  childCount: 0, // 暂时显示空列表
                ),
              ),
            ),
            
            // 空状态
            const SliverFillRemaining(
              hasScrollBody: false,
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      Icons.business_outlined,
                      size: 64,
                      color: Color(0xFF8C92A4),
                    ),
                    SizedBox(height: 16),
                    Text(
                      '暂无供应商',
                      style: TextStyle(
                        fontSize: 16,
                        color: Color(0xFF8C92A4),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSupplierCard({
    required String name,
    required String contact,
    required String phone,
    required String status,
  }) {
    final isActive = status == 'active';
    
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: Container(
          width: 50,
          height: 50,
          decoration: BoxDecoration(
            color: isActive ? Colors.green[100] : Colors.grey[200],
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: isActive ? Colors.green : Colors.grey,
              width: 2,
            ),
          ),
          child: Icon(
            Icons.business,
            color: isActive ? Colors.green : Colors.grey,
          ),
        ),
        title: Text(
          name,
          style: const TextStyle(
            fontWeight: FontWeight.w600,
          ),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 4),
            Text('联系人: $contact'),
            Text('电话: $phone'),
          ],
        ),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: isActive
                    ? Colors.green.withOpacity(0.1)
                    : Colors.grey.withOpacity(0.1),
                borderRadius: BorderRadius.circular(4),
              ),
              child: Text(
                isActive ? '正常' : '禁用',
                style: TextStyle(
                  color: isActive ? Colors.green : Colors.grey,
                  fontSize: 12,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
            const SizedBox(width: 8),
            PopupMenuButton(
              itemBuilder: (context) => [
                const PopupMenuItem(
                  value: 'edit',
                  child: Row(
                    children: [
                      Icon(Icons.edit, size: 20),
                      SizedBox(width: 8),
                      Text('编辑'),
                    ],
                  ),
                ),
                PopupMenuItem(
                  value: 'toggle',
                  child: Row(
                    children: [
                      Icon(
                        isActive ? Icons.block : Icons.check_circle,
                        size: 20,
                      ),
                      const SizedBox(width: 8),
                      Text(isActive ? '禁用' : '启用'),
                    ],
                  ),
                ),
              ],
              onSelected: (value) {
                // TODO: 处理编辑/启用/禁用操作
              },
            ),
          ],
        ),
        onTap: () {
          // TODO: 跳转到供应商详情
        },
      ),
    );
  }
}

