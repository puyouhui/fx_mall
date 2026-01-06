import 'package:flutter/material.dart';
import 'package:super_app/api/suppliers_api.dart';
import 'package:super_app/models/supplier_payment.dart';
import 'package:super_app/pages/suppliers/supplier_payment_detail_page.dart';

class SuppliersPage extends StatefulWidget {
  const SuppliersPage({super.key});

  @override
  State<SuppliersPage> createState() => _SuppliersPageState();
}

class _SuppliersPageState extends State<SuppliersPage> {
  List<SupplierPaymentStats> _statsList = [];
  bool _isLoading = false;
  int _currentPage = 1;
  final int _pageSize = 20;
  bool _hasMore = true;
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _loadStats();
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
            _scrollController.position.maxScrollExtent - 200 &&
        !_isLoading &&
        _hasMore) {
      _loadMoreStats();
    }
  }

  Future<void> _loadStats({bool reset = false}) async {
    if (_isLoading) return;

    if (reset) {
      setState(() {
        _currentPage = 1;
        _statsList = [];
        _hasMore = true;
      });
    }

    setState(() => _isLoading = true);

    try {
      final response = await SuppliersApi.getPaymentStats(
        pageNum: _currentPage,
        pageSize: _pageSize,
      );

      if (!mounted) return;

      if (response.isSuccess && response.data != null) {
        final newStats = response.data!;
        setState(() {
          if (reset) {
            _statsList = newStats;
          } else {
            _statsList.addAll(newStats);
          }
          _hasMore = newStats.length >= _pageSize;
          if (_hasMore && !reset) {
            _currentPage++;
          }
          _isLoading = false;
        });
      } else {
        setState(() => _isLoading = false);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(response.message)),
          );
        }
      }
    } catch (e) {
      if (!mounted) return;
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('加载失败: ${e.toString()}')),
        );
      }
    }
  }

  Future<void> _loadMoreStats() async {
    if (!_hasMore || _isLoading) return;
    await _loadStats(reset: false);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: RefreshIndicator(
        onRefresh: () => _loadStats(reset: true),
        child: _isLoading && _statsList.isEmpty
            ? const Center(child: CircularProgressIndicator())
            : _statsList.isEmpty
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.business_outlined,
                          size: 64,
                          color: Colors.grey.shade300,
                        ),
                        const SizedBox(height: 16),
                        Text(
                          '暂无供应商付款数据',
                          style: TextStyle(
                            fontSize: 16,
                            color: Colors.grey.shade600,
                          ),
                        ),
                      ],
                    ),
                  )
                : ListView.builder(
                    controller: _scrollController,
                    padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                    itemCount: _statsList.length + (_hasMore ? 1 : 0),
                    itemBuilder: (context, index) {
                      if (index >= _statsList.length) {
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

                      final stats = _statsList[index];
                      return _buildSupplierCard(stats);
                    },
                  ),
      ),
    );
  }

  Widget _buildSupplierCard(SupplierPaymentStats stats) {
    return InkWell(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (_) => SupplierPaymentDetailPage(
              supplierId: stats.supplierId,
              supplierName: stats.supplierName,
            ),
          ),
        );
      },
      child: Container(
        margin: const EdgeInsets.only(top: 12),
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: Colors.white,
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
            // 供应商名称
            Text(
              stats.supplierName,
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: Color(0xFF20253A),
              ),
            ),
            const SizedBox(height: 12),
            // 金额信息
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        '应付款总额',
                        style: TextStyle(
                          fontSize: 12,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        '¥${stats.totalAmount.toStringAsFixed(2)}',
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: Color(0xFF20253A),
                        ),
                      ),
                    ],
                  ),
                ),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        '待付款',
                        style: TextStyle(
                          fontSize: 12,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        '¥${stats.pendingAmount.toStringAsFixed(2)}',
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: Color(0xFFFF6B6B),
                        ),
                      ),
                    ],
                  ),
                ),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        '已付款',
                        style: TextStyle(
                          fontSize: 12,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        '¥${stats.paidAmount.toStringAsFixed(2)}',
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: Color(0xFF20CB6B),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            // 订单数量和查看详情
            Row(
              children: [
                Text(
                  '订单数量: ${stats.orderCount}',
                  style: const TextStyle(
                    fontSize: 13,
                    color: Color(0xFF8C92A4),
                  ),
                ),
                const Spacer(),
                const Text(
                  '查看详情 >',
                  style: TextStyle(
                    fontSize: 13,
                    color: Color(0xFF20CB6B),
                    fontWeight: FontWeight.w500,
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

