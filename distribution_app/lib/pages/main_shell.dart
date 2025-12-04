import 'package:flutter/material.dart';
import 'order_hall_view.dart';
import 'profile_view.dart';

/// 登录后的主框架：底部两个 Tab（接单大厅 / 我的）
class MainShell extends StatefulWidget {
  const MainShell({
    super.key,
    required this.courierPhone,
  });

  final String courierPhone;

  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _currentIndex = 0;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_currentIndex == 0 ? '接单大厅' : '我的'),
      ),
      body: IndexedStack(
        index: _currentIndex,
        children: [
          const OrderHallView(),
          ProfileView(
            courierPhone: widget.courierPhone,
          ),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentIndex,
        onDestinationSelected: (index) {
          setState(() {
            _currentIndex = index;
          });
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.list_alt_outlined),
            selectedIcon: Icon(Icons.list_alt),
            label: '接单大厅',
          ),
          NavigationDestination(
            icon: Icon(Icons.person_outline),
            selectedIcon: Icon(Icons.person),
            label: '我的',
          ),
        ],
      ),
    );
  }
}


