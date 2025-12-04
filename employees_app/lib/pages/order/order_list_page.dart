import 'package:flutter/material.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/order/order_detail_page.dart';

/// 订单列表页面（员工查看名下客户的所有订单）
class OrderListPage extends StatefulWidget {
  const OrderListPage({super.key});

  @override
  State<OrderListPage> createState() => _OrderListPageState();
}

class _OrderListPageState extends State<OrderListPage> {
  final TextEditingController _searchController = TextEditingController();
  final ScrollController _scrollController = ScrollController();

  final List<Map<String, dynamic>> _orders = [];
  bool _isLoading = false;
  bool _isLoadingMore = false;
  bool _hasMore = true;
  int _pageNum = 1;
  final int _pageSize = 20;
  String _keyword = '';
  String _status = ''; // 全部

  @override
  void initState() {
    super.initState();
    _loadOrders(reset: true);
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
      _loadOrders();
    }
  }

  Future<void> _loadOrders({bool reset = false}) async {
    if (_isLoading || _isLoadingMore) return;

    if (reset) {
      setState(() {
        _isLoading = true;
        _pageNum = 1;
        _hasMore = true;
        _orders.clear();
      });
    } else {
      setState(() {
        _isLoadingMore = true;
      });
    }

    final response = await Request.get<Map<String, dynamic>>(
      '/employee/sales/orders',
      queryParams: {
        'pageNum': _pageNum.toString(),
        'pageSize': _pageSize.toString(),
        if (_keyword.isNotEmpty) 'keyword': _keyword,
        if (_status.isNotEmpty) 'status': _status,
      },
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final List<dynamic> list = (data['list'] as List<dynamic>? ?? []);
      final orders = list.cast<Map<String, dynamic>>();

      setState(() {
        if (reset) {
          _orders
            ..clear()
            ..addAll(orders);
        } else {
          _orders.addAll(orders);
        }
        final total = data['total'] as int? ?? _orders.length;
        _hasMore = _orders.length < total;
        if (_hasMore) {
          _pageNum++;
        }
      });
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            response.message.isNotEmpty ? response.message : '获取订单列表失败',
          ),
        ),
      );
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
    _loadOrders(reset: true);
  }

  void _onStatusChanged(String? value) {
    setState(() {
      _status = value ?? '';
    });
    _loadOrders(reset: true);
  }

  /// 顶部状态 Tab（全部 / 待配送 / 已送达 / 已取消）
  Widget _buildStatusTabs() {
    return Container(
      // 使用透明背景，让下面的渐变背景透出，避免与页面整体风格不一致
      color: Colors.transparent,
      child: TabBar(
        isScrollable: false,
        indicatorSize: TabBarIndicatorSize.tab,
        // 自定义一个短一点的下划线，并去掉整行的下边框
        indicator: const UnderlineTabIndicator(
          borderSide: BorderSide(width: 2, color: Colors.white),
          insets: EdgeInsets.symmetric(horizontal: 24),
        ),
        labelPadding: EdgeInsets.zero,
        indicatorColor: Colors.transparent,
        indicatorWeight: 0,
        dividerColor: Colors.transparent, // 取消整行底部的分割线
        labelColor: Colors.white,
        unselectedLabelColor: const Color(0xFFE0F5EB),
        onTap: (index) {
          const keys = ['', 'pending_delivery', 'delivered', 'cancelled'];
          _onStatusChanged(keys[index]);
        },
        tabs: const [
          Tab(text: '全部'),
          Tab(text: '待配送'),
          Tab(text: '已送达'),
          Tab(text: '已取消'),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: 4,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('订单查询'),
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
        extendBody: true, // 让body延伸到系统操作条下方
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
                _buildSearchBar(),
                _buildStatusTabs(),
                Expanded(
                  child: _isLoading && _orders.isEmpty
                      ? const Center(
                          child: CircularProgressIndicator(
                            valueColor: AlwaysStoppedAnimation<Color>(
                              Colors.white,
                            ),
                          ),
                        )
                      : RefreshIndicator(
                          onRefresh: () => _loadOrders(reset: true),
                          child: _orders.isEmpty
                              ? ListView(
                                  padding: EdgeInsets.fromLTRB(
                                    16,
                                    40,
                                    16,
                                    16 +
                                        MediaQuery.of(
                                          context,
                                        ).padding.bottom, // 添加底部安全区域内边距
                                  ),
                                  children: const [
                                    Center(
                                      child: Text(
                                        '暂无订单数据',
                                        style: TextStyle(
                                          fontSize: 14,
                                          color: Colors.white,
                                        ),
                                      ),
                                    ),
                                  ],
                                )
                              : ListView.builder(
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
                                  itemCount:
                                      _orders.length + (_hasMore ? 1 : 0),
                                  itemBuilder: (context, index) {
                                    if (index >= _orders.length) {
                                      return const Padding(
                                        padding: EdgeInsets.symmetric(
                                          vertical: 16,
                                        ),
                                        child: Center(
                                          child: CircularProgressIndicator(
                                            valueColor:
                                                AlwaysStoppedAnimation<Color>(
                                                  Color(0xFF20CB6B),
                                                ),
                                          ),
                                        ),
                                      );
                                    }
                                    final order = _orders[index];
                                    return _buildOrderCard(order);
                                  },
                                ),
                        ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSearchBar() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        children: [
          Expanded(
            child: SizedBox(
              height: 44,
              child: Container(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(24),
                ),
                padding: const EdgeInsets.symmetric(horizontal: 0),
                child: TextField(
                  controller: _searchController,
                  textAlignVertical: TextAlignVertical.center,
                  decoration: const InputDecoration(
                    hintText: '输入订单信息查询',
                    border: InputBorder.none,
                    isDense: true,
                    contentPadding: EdgeInsets.zero,
                    prefixIcon: Icon(Icons.search, color: Color(0xFF8C92A4)),
                  ),
                  onSubmitted: (_) => _onSearch(),
                ),
              ),
            ),
          ),
          const SizedBox(width: 8),
          SizedBox(
            height: 44,
            width: 96,
            child: ElevatedButton(
              onPressed: _onSearch,
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.white,
                foregroundColor: const Color(0xFF20CB6B),
                padding: const EdgeInsets.symmetric(horizontal: 0),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(24),
                ),
                elevation: 0,
              ),
              child: const Text(
                '搜索',
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildOrderCard(Map<String, dynamic> order) {
    final orderNumber = order['order_number'] as String? ?? '';
    final status = order['status'] as String? ?? '';
    final totalAmount = (order['total_amount'] as num?)?.toDouble() ?? 0.0;
    final itemCount = (order['item_count'] as int?) ?? 0;
    final userName = order['user_name'] as String? ?? '';
    final userCode = order['user_code'] as String? ?? '';
    final storeName = order['store_name'] as String? ?? '';
    final address = order['address'] as String? ?? '';
    final createdAt = order['created_at']?.toString() ?? '';

    String statusText = status;
    Color statusColor = const Color(0xFF8C92A4);
    if (status == 'pending_delivery' || status == 'pending') {
      statusText = '待配送';
      statusColor = const Color(0xFFFFA940);
    } else if (status == 'delivered' || status == 'shipped') {
      statusText = '已送达';
      statusColor = const Color(0xFF20CB6B);
    } else if (status == 'cancelled') {
      statusText = '已取消';
      statusColor = const Color(0xFFB0B4C3);
    }

    return InkWell(
      onTap: () {
        if (order['id'] != null) {
          final id = order['id'] as int;
          Navigator.of(context).push(
            MaterialPageRoute(builder: (_) => OrderDetailPage(orderId: id)),
          );
        }
      },
      borderRadius: BorderRadius.circular(16),
      child: Container(
        margin: const EdgeInsets.only(top: 12),
        padding: const EdgeInsets.all(14),
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
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    storeName.isNotEmpty ? storeName : '门店名称未填写',
                    style: const TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF20253A),
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                if (orderNumber.isNotEmpty)
                  Text(
                    orderNumber,
                    style: const TextStyle(
                      fontSize: 11,
                      color: Color(0xFF8C92A4),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 6),
            if (userName.isNotEmpty || userCode.isNotEmpty)
              Text(
                userCode.isNotEmpty ? '$userName（编号 $userCode）' : userName,
                style: const TextStyle(fontSize: 13, color: Color(0xFF40475C)),
              ),
            if (address.isNotEmpty) ...[
              const SizedBox(height: 4),
              Text(
                address,
                style: const TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ],
            const SizedBox(height: 8),
            Row(
              children: [
                Text(
                  '共$itemCount件商品',
                  style: const TextStyle(
                    fontSize: 12,
                    color: Color(0xFF8C92A4),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    createdAt,
                    style: const TextStyle(
                      fontSize: 11,
                      color: Color(0xFFB0B4C3),
                    ),
                    textAlign: TextAlign.right,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Text(
                  statusText,
                  style: TextStyle(
                    fontSize: 13,
                    color: statusColor,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const Spacer(),
                const Text(
                  '实付金额：',
                  style: TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
                ),
                Text(
                  '¥${totalAmount.toStringAsFixed(2)}',
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF20CB6B),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
