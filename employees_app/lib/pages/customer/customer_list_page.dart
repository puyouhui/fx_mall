import 'package:flutter/material.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/customer/customer_profile_page.dart';
import 'package:employees_app/pages/customer/customer_detail_page.dart';
import 'package:intl/intl.dart';

/// 我的客户列表页面（销售员）
class CustomerListPage extends StatefulWidget {
  /// 是否作为“选择客户”弹出，true 时点击客户会返回选中的客户数据
  final bool pickMode;

  const CustomerListPage({super.key, this.pickMode = false});

  @override
  State<CustomerListPage> createState() => _CustomerListPageState();
}

class _CustomerListPageState extends State<CustomerListPage> {
  final TextEditingController _searchController = TextEditingController();
  final ScrollController _scrollController = ScrollController();

  final List<Map<String, dynamic>> _customers = [];
  bool _isLoading = false;
  bool _isLoadingMore = false;
  bool _hasMore = true;
  int _pageNum = 1;
  final int _pageSize = 20;
  String _keyword = '';

  @override
  void initState() {
    super.initState();
    _loadCustomers(reset: true);
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _searchController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
            _scrollController.position.maxScrollExtent - 200 &&
        !_isLoadingMore &&
        _hasMore &&
        !_isLoading) {
      _loadCustomers();
    }
  }

  Future<void> _loadCustomers({bool reset = false}) async {
    if (_isLoading || _isLoadingMore) return;

    if (reset) {
      setState(() {
        _isLoading = true;
        _pageNum = 1;
        _hasMore = true;
        _customers.clear();
      });
    } else {
      setState(() {
        _isLoadingMore = true;
      });
    }

    final response = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customers',
      queryParams: {
        'pageNum': _pageNum.toString(),
        'pageSize': _pageSize.toString(),
        if (_keyword.isNotEmpty) 'keyword': _keyword,
      },
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final List<dynamic> list = (data['list'] as List<dynamic>? ?? []);
      final customers = list.cast<Map<String, dynamic>>();

      setState(() {
        if (reset) {
          _customers
            ..clear()
            ..addAll(customers);
        } else {
          _customers.addAll(customers);
        }
        final total = data['total'] as int? ?? _customers.length;
        _hasMore = _customers.length < total;
        if (_hasMore) {
          _pageNum++;
        }
      });
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              response.message.isNotEmpty ? response.message : '获取客户列表失败',
            ),
          ),
        );
      }
    }

    if (mounted) {
      setState(() {
        _isLoading = false;
        _isLoadingMore = false;
      });
    }
  }

  void _onSearch() {
    _keyword = _searchController.text.trim();
    _loadCustomers(reset: true);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      extendBody: true, // 让body延伸到系统操作条下方
      appBar: AppBar(
        title: const Text('我的客户'),
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
          bottom: false, // 底部不使用SafeArea，让内容延伸到系统操作条
          child: Column(
            children: [
              // 搜索栏
              Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 12,
                ),
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
                          onSubmitted: (_) => _onSearch(),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    ElevatedButton(
                      onPressed: _onSearch,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.white,
                        foregroundColor: const Color(0xFF20CB6B),
                        padding: const EdgeInsets.symmetric(
                          horizontal: 14,
                          vertical: 10,
                        ),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(20),
                        ),
                        elevation: 0,
                      ),
                      child: const Text(
                        '搜索',
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: _isLoading && _customers.isEmpty
                    ? const Center(
                        child: CircularProgressIndicator(
                          valueColor: AlwaysStoppedAnimation<Color>(
                            Colors.white,
                          ),
                        ),
                      )
                    : RefreshIndicator(
                        onRefresh: () => _loadCustomers(reset: true),
                        child: ListView.builder(
                          controller: _scrollController,
                          padding: EdgeInsets.fromLTRB(
                            16,
                            0,
                            16,
                            16 +
                                MediaQuery.of(
                                  context,
                                ).padding.bottom, // 添加底部安全区域内边距
                          ),
                          itemCount: _customers.length + (_hasMore ? 1 : 0),
                          itemBuilder: (context, index) {
                            if (index >= _customers.length) {
                              // 底部加载更多
                              return const Padding(
                                padding: EdgeInsets.symmetric(vertical: 16),
                                child: Center(
                                  child: CircularProgressIndicator(
                                    valueColor: AlwaysStoppedAnimation<Color>(
                                      Color(0xFF20CB6B),
                                    ),
                                  ),
                                ),
                              );
                            }
                            final customer = _customers[index];
                            return _buildCustomerCard(customer);
                          },
                        ),
                      ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCustomerCard(Map<String, dynamic> customer) {
    final name = (customer['name'] as String?) ?? '未填写名称';
    final phone = (customer['phone'] as String?) ?? '';
    final userCode = (customer['user_code'] as String?) ?? '';
    final createdAt = customer['created_at']?.toString() ?? '';
    final defaultAddress = customer['default_address'] as Map<String, dynamic>?;
    final addressName = defaultAddress?['name'] as String? ?? '';
    final addressText = defaultAddress?['address'] as String? ?? '';
    final contact = defaultAddress?['contact'] as String? ?? '';
    final addrPhone = defaultAddress?['phone'] as String? ?? '';
    final orderCount = (customer['order_count'] as int?) ?? 0;
    final addressCount = (customer['address_count'] as int?) ?? 0;

    final id = customer['id'] as int? ?? 0;

    return InkWell(
      onTap: id <= 0
          ? null
          : () {
              if (widget.pickMode) {
                // 作为选择客户使用，直接返回选中的客户信息
                Navigator.of(context).pop<Map<String, dynamic>>(customer);
              } else {
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (_) =>
                        CustomerDetailPage(customerId: id, customerName: name),
                  ),
                );
              }
            },
      borderRadius: BorderRadius.circular(16),
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
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
          // 上、左、右保留内边距，底部去掉内边距，让按钮区紧贴卡片底部
          padding: const EdgeInsets.fromLTRB(14, 14, 14, 0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      name,
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF20253A),
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  if (userCode.isNotEmpty)
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFF20CB6B).withOpacity(0.08),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        '编号 $userCode',
                        style: const TextStyle(
                          fontSize: 11,
                          color: Color(0xFF20CB6B),
                        ),
                      ),
                    ),
                ],
              ),
              if (addressText.isNotEmpty) ...[
                const SizedBox(height: 8),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // 地址名称 + 默认地址提示
                    Text.rich(
                      TextSpan(
                        children: [
                          TextSpan(
                            text: addressName.isNotEmpty ? addressName : '地址',
                            style: const TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.w600,
                              color: Color(0xFF20253A),
                            ),
                          ),
                          const TextSpan(
                            text: '  默认地址',
                            style: TextStyle(
                              fontSize: 11,
                              color: Color(0xFF20CB6B),
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      addressText,
                      style: const TextStyle(
                        fontSize: 13,
                        color: Color(0xFF8C92A4),
                      ),
                    ),
                    if (contact.isNotEmpty || addrPhone.isNotEmpty)
                      Padding(
                        padding: const EdgeInsets.only(top: 2),
                        child: Text(
                          '$contact  $addrPhone',
                          style: const TextStyle(
                            fontSize: 13,
                            color: Color(0xFF8C92A4),
                          ),
                        ),
                      ),
                  ],
                ),
              ],
              const SizedBox(height: 10),
              // 下单次数 & 地址数量（水平居中，各占一半）
              Row(
                children: [
                  Expanded(
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(
                          Icons.shopping_bag_outlined,
                          size: 14,
                          color: Color(0xFF8C92A4),
                        ),
                        const SizedBox(width: 4),
                        Text(
                          '下单次数：$orderCount',
                          style: const TextStyle(
                            fontSize: 13,
                            color: Color(0xFF8C92A4),
                          ),
                        ),
                      ],
                    ),
                  ),
                  Expanded(
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(
                          Icons.home_work_outlined,
                          size: 14,
                          color: Color(0xFF8C92A4),
                        ),
                        const SizedBox(width: 4),
                        Text(
                          '地址数量：$addressCount',
                          style: const TextStyle(
                            fontSize: 13,
                            color: Color(0xFF8C92A4),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              if (createdAt.isNotEmpty) ...[
                const SizedBox(height: 6),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(
                      Icons.access_time,
                      size: 13,
                      color: Color(0xFFB0B4C3),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      '绑定时间：${_formatBindTime(createdAt)}',
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFFB0B4C3),
                      ),
                    ),
                  ],
                ),
              ],
              const SizedBox(height: 6),
              const Divider(
                height: 1,
                thickness: 0.5,
                color: Color(0xFFE5E7F0),
              ),
              // 按钮区域：固定高度内垂直居中，整体靠右
              SizedBox(
                height: 50,
                child: Align(
                  alignment: Alignment.centerRight,
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.center,
                    children: [
                      Container(
                        decoration: BoxDecoration(
                          color: const Color(0xFF4C8DF6).withOpacity(0.06),
                          borderRadius: BorderRadius.circular(16),
                        ),
                        child: TextButton.icon(
                          onPressed: () {
                            Navigator.of(context).push(
                              MaterialPageRoute(
                                builder: (_) => CustomerProfilePage(
                                  initialUserCode: userCode,
                                ),
                              ),
                            );
                          },
                          style: TextButton.styleFrom(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 10,
                              vertical: 4,
                            ),
                            minimumSize: Size.zero,
                            tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                          ),
                          icon: const Icon(
                            Icons.edit,
                            size: 16,
                            color: Color(0xFF4C8DF6),
                          ),
                          label: const Text(
                            '编辑',
                            style: TextStyle(
                              fontSize: 13,
                              color: Color(0xFF4C8DF6),
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 16),
                      Container(
                        decoration: BoxDecoration(
                          color: const Color(0xFF20CB6B).withOpacity(0.06),
                          borderRadius: BorderRadius.circular(16),
                        ),
                        child: TextButton.icon(
                          onPressed: phone.isEmpty
                              ? null
                              : () {
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    SnackBar(content: Text('请拨打：$phone')),
                                  );
                                },
                          style: TextButton.styleFrom(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 10,
                              vertical: 4,
                            ),
                            minimumSize: Size.zero,
                            tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                          ),
                          icon: const Icon(
                            Icons.phone_in_talk,
                            size: 16,
                            color: Color(0xFF20CB6B),
                          ),
                          label: const Text(
                            '拨打',
                            style: TextStyle(
                              fontSize: 13,
                              color: Color(0xFF20CB6B),
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

String _formatBindTime(String raw) {
  try {
    final dt = DateTime.tryParse(raw);
    if (dt == null) return raw;
    // 显示为：yyyy-MM-dd HH:mm
    return DateFormat('yyyy-MM-dd HH:mm').format(dt.toLocal());
  } catch (_) {
    return raw;
  }
}
