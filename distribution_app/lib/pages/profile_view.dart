import 'package:flutter/material.dart';
import '../utils/storage.dart';
import 'income_stats_view.dart';
import 'order_history_view.dart';

/// 我的 页面视图
class ProfileView extends StatefulWidget {
  const ProfileView({super.key, required this.courierPhone});

  final String courierPhone;

  @override
  State<ProfileView> createState() => _ProfileViewState();
}

class _ProfileViewState extends State<ProfileView> {
  String _employeeName = '配送员';

  @override
  void initState() {
    super.initState();
    _loadEmployeeInfo();
  }

  Future<void> _loadEmployeeInfo() async {
    final employeeInfo = await Storage.getEmployeeInfo();
    if (employeeInfo != null && mounted) {
      setState(() {
        _employeeName = employeeInfo['name'] as String? ?? '配送员';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [Color(0xFF20CB6B), Color(0xFFEFF7F2)],
          stops: [0.0, 0.3],
        ),
      ),
      child: SafeArea(
        child: LayoutBuilder(
          builder: (context, constraints) {
            return SingleChildScrollView(
              child: ConstrainedBox(
                constraints: BoxConstraints(minHeight: constraints.maxHeight),
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // 用户信息卡片
                      Container(
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(20),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.06),
                              blurRadius: 20,
                              offset: const Offset(0, 4),
                              spreadRadius: 0,
                            ),
                          ],
                        ),
                        child: Padding(
                          padding: const EdgeInsets.all(24),
                          child: Row(
                            children: [
                              Container(
                                width: 72,
                                height: 72,
                                decoration: BoxDecoration(
                                  gradient: LinearGradient(
                                    begin: Alignment.topLeft,
                                    end: Alignment.bottomRight,
                                    colors: [
                                      const Color(0xFF20CB6B).withOpacity(0.15),
                                      const Color(0xFF20CB6B).withOpacity(0.08),
                                    ],
                                  ),
                                  shape: BoxShape.circle,
                                ),
                                child: const Icon(
                                  Icons.person,
                                  size: 40,
                                  color: Color(0xFF20CB6B),
                                ),
                              ),
                              const SizedBox(width: 20),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      _employeeName,
                                      style: const TextStyle(
                                        fontSize: 22,
                                        fontWeight: FontWeight.w700,
                                        color: Color(0xFF20253A),
                                        letterSpacing: 0.2,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    Row(
                                      children: [
                                        Icon(
                                          Icons.phone_outlined,
                                          size: 16,
                                          color: const Color(0xFF8C92A4),
                                        ),
                                        const SizedBox(width: 6),
                                        Text(
                                          widget.courierPhone,
                                          style: const TextStyle(
                                            fontSize: 15,
                                            color: Color(0xFF8C92A4),
                                            fontWeight: FontWeight.w400,
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
                      const SizedBox(height: 28),
                      // 常用功能标题
                      const Padding(
                        padding: EdgeInsets.symmetric(horizontal: 4),
                        child: Text(
                          '常用功能',
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.w700,
                            color: Color(0xFF20253A),
                            letterSpacing: 0.3,
                          ),
                        ),
                      ),
                      const SizedBox(height: 16),
                      // 功能菜单卡片
                      Container(
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(20),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.06),
                              blurRadius: 20,
                              offset: const Offset(0, 4),
                              spreadRadius: 0,
                            ),
                          ],
                        ),
                        child: Column(
                          children: [
                            _buildMenuItem(
                              icon: Icons.shopping_cart_outlined,
                              title: '批量取货',
                              onTap: () {
                                Navigator.of(
                                  context,
                                ).pushNamed('/batch-pickup');
                              },
                            ),
                            Divider(
                              height: 1,
                              indent: 64,
                              thickness: 0.5,
                              color: const Color(0xFFE5E7EB),
                            ),
                            _buildMenuItem(
                              icon: Icons.history_outlined,
                              title: '历史订单',
                              onTap: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (context) =>
                                        const OrderHistoryView(),
                                  ),
                                );
                              },
                            ),
                            Divider(
                              height: 1,
                              indent: 64,
                              thickness: 0.5,
                              color: const Color(0xFFE5E7EB),
                            ),
                            _buildMenuItem(
                              icon: Icons.account_balance_wallet_outlined,
                              title: '收入统计',
                              onTap: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (context) =>
                                        const IncomeStatsView(),
                                  ),
                                );
                              },
                            ),
                            Divider(
                              height: 1,
                              indent: 64,
                              thickness: 0.5,
                              color: const Color(0xFFE5E7EB),
                            ),
                            _buildMenuItem(
                              icon: Icons.logout_rounded,
                              title: '退出登录',
                              iconColor: const Color(0xFFEF4444),
                              textColor: const Color(0xFFEF4444),
                              onTap: () {
                                _confirmLogout(context);
                              },
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
        ),
      ),
    );
  }

  Widget _buildMenuItem({
    required IconData icon,
    required String title,
    required VoidCallback onTap,
    Color? iconColor,
    Color? textColor,
  }) {
    final defaultIconColor = iconColor ?? const Color(0xFF20CB6B);
    final defaultTextColor = textColor ?? const Color(0xFF20253A);

    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(20),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
          child: Row(
            children: [
              Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: defaultIconColor.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(icon, color: defaultIconColor, size: 22),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Text(
                  title,
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: defaultTextColor,
                    letterSpacing: 0.2,
                  ),
                ),
              ),
              Icon(
                Icons.chevron_right_rounded,
                color: const Color(0xFF8C92A4).withOpacity(0.6),
                size: 22,
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _confirmLogout(BuildContext context) async {
    final result = await showDialog<bool>(
      context: context,
      builder: (context) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          title: const Text(
            '确认退出登录？',
            style: TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.w700,
              color: Color(0xFF20253A),
            ),
          ),
          content: const Text(
            '退出后需要重新登录才能继续接单。',
            style: TextStyle(
              fontSize: 15,
              color: Color(0xFF8C92A4),
              height: 1.5,
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 12,
                ),
              ),
              child: const Text(
                '取消',
                style: TextStyle(
                  fontSize: 16,
                  color: Color(0xFF8C92A4),
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
            TextButton(
              onPressed: () => Navigator.of(context).pop(true),
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 12,
                ),
              ),
              child: const Text(
                '退出登录',
                style: TextStyle(
                  fontSize: 16,
                  color: Color(0xFFEF4444),
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          ],
        );
      },
    );

    if (result == true && context.mounted) {
      // 清除登录信息
      await Storage.clearAll();

      // 退出登录：返回到登录页并清空之前的页面栈
      Navigator.of(context).pushNamedAndRemoveUntil('/login', (route) => false);
    }
  }
}
