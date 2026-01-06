import 'package:flutter/material.dart';
import 'package:super_app/api/auth_api.dart';
import 'package:super_app/pages/statistics/statistics_page.dart';
import 'package:super_app/pages/orders/orders_page.dart';
import 'package:super_app/pages/products/create_product_page.dart';
import 'package:super_app/pages/suppliers/suppliers_page.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int _currentIndex = 0;

  final List<Widget> _pages = const [
    StatisticsPage(),
    OrdersPage(),
    CreateProductPage(),
    SuppliersPage(),
  ];

  final List<String> _titles = const [
    '数据统计',
    '订单管理',
    '商品管理',
    '供应商管理',
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_titles[_currentIndex]),
        backgroundColor: const Color(0xFF20CB6B),
        foregroundColor: Colors.white,
        actions: [
          if (_currentIndex == 0)
            IconButton(
              icon: const Icon(Icons.logout),
              tooltip: '退出登录',
              onPressed: () async {
                // 确认退出
                final confirm = await showDialog<bool>(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text('确认退出'),
                    content: const Text('确定要退出登录吗？'),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.of(context).pop(false),
                        child: const Text('取消'),
                      ),
                      TextButton(
                        onPressed: () => Navigator.of(context).pop(true),
                        child: const Text('确定'),
                      ),
                    ],
                  ),
                );

                if (confirm == true) {
                  await AuthApi.logout();
                  if (context.mounted) {
                    Navigator.of(context).pushNamedAndRemoveUntil(
                      '/login',
                      (route) => false,
                    );
                  }
                }
              },
            ),
          if (_currentIndex == 2)
            IconButton(
              icon: const Icon(Icons.add),
              tooltip: '添加商品',
              onPressed: () {
                Navigator.of(context).pushNamed('/edit_product');
              },
            ),
        ],
      ),
      body: IndexedStack(
        index: _currentIndex,
        children: _pages,
      ),
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.08),
              blurRadius: 12,
              offset: const Offset(0, -2),
            ),
          ],
        ),
        child: NavigationBar(
          selectedIndex: _currentIndex,
          onDestinationSelected: (index) {
            setState(() {
              _currentIndex = index;
            });
          },
          backgroundColor: Colors.transparent,
          elevation: 0,
          indicatorColor: const Color(0xFF20CB6B).withOpacity(0.15),
          labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
          labelTextStyle: MaterialStateProperty.resolveWith((states) {
            if (states.contains(MaterialState.selected)) {
              return const TextStyle(
                color: Color(0xFF20CB6B),
                fontSize: 12,
                fontWeight: FontWeight.w600,
              );
            }
            return const TextStyle(
              color: Color(0xFF8C92A4),
              fontSize: 12,
              fontWeight: FontWeight.normal,
            );
          }),
          destinations: const [
            NavigationDestination(
              icon: Icon(
                Icons.bar_chart_outlined,
                color: Color(0xFF8C92A4),
                size: 24,
              ),
              selectedIcon: Icon(
                Icons.bar_chart,
                color: Color(0xFF20CB6B),
                size: 24,
              ),
              label: '统计',
            ),
            NavigationDestination(
              icon: Icon(
                Icons.receipt_long_outlined,
                color: Color(0xFF8C92A4),
                size: 24,
              ),
              selectedIcon: Icon(
                Icons.receipt_long,
                color: Color(0xFF20CB6B),
                size: 24,
              ),
              label: '订单',
            ),
            NavigationDestination(
              icon: Icon(
                Icons.inventory_2_outlined,
                color: Color(0xFF8C92A4),
                size: 24,
              ),
              selectedIcon: Icon(
                Icons.inventory_2,
                color: Color(0xFF20CB6B),
                size: 24,
              ),
              label: '商品',
            ),
            NavigationDestination(
              icon: Icon(
                Icons.business_outlined,
                color: Color(0xFF8C92A4),
                size: 24,
              ),
              selectedIcon: Icon(
                Icons.business,
                color: Color(0xFF20CB6B),
                size: 24,
              ),
              label: '供应商',
            ),
          ],
        ),
      ),
    );
  }
}

