import 'package:flutter/material.dart';
import 'package:super_app/api/statistics_api.dart';

class StatisticsPage extends StatefulWidget {
  const StatisticsPage({super.key});

  @override
  State<StatisticsPage> createState() => _StatisticsPageState();
}

class _StatisticsPageState extends State<StatisticsPage>
    with AutomaticKeepAliveClientMixin {
  String _timeRange = 'today'; // today, week, month
  bool _isLoading = true;
  Map<String, dynamic>? _statsData;

  @override
  bool get wantKeepAlive => true;

  @override
  void initState() {
    super.initState();
    _loadStatistics();
  }

  Future<void> _loadStatistics() async {
    if (!mounted) return;

    setState(() {
      _isLoading = true;
    });

    try {
      final response = await StatisticsApi.getDashboardStats(
        timeRange: _timeRange,
      );

      if (!mounted) return;

      if (response.isSuccess && response.data != null) {
        if (mounted) {
          setState(() {
            _statsData = response.data;
            _isLoading = false;
          });
        }
      } else {
        setState(() {
          _isLoading = false;
          _statsData = null; // 清除旧数据
        });
        if (mounted) {
          ScaffoldMessenger.of(
            context,
          ).showSnackBar(SnackBar(content: Text(response.message)));
        }
      }
    } catch (e) {
      if (!mounted) return;

      setState(() {
        _isLoading = false;
        _statsData = null; // 清除旧数据
      });
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('加载失败: ${e.toString()}')));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    super.build(context); // 必须调用，因为使用了 AutomaticKeepAliveClientMixin

    return Scaffold(
      body: RefreshIndicator(
        onRefresh: _loadStatistics,
        child: _isLoading
            ? const Center(child: CircularProgressIndicator())
            : SingleChildScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // 日期选择器
                    _buildDateSelector(),
                    const SizedBox(height: 16),

                    // 统计卡片网格
                    _buildStatCards(),
                  ],
                ),
              ),
      ),
    );
  }

  Widget _buildDateSelector() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 4,
            offset: const Offset(0, 1),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: SegmentedButton<String>(
              segments: const [
                ButtonSegment(value: 'today', label: Text('今日')),
                ButtonSegment(value: 'week', label: Text('本周')),
                ButtonSegment(value: 'month', label: Text('本月')),
              ],
              selected: {_timeRange},
              onSelectionChanged: (Set<String> newSelection) {
                if (!mounted) return;

                final newTimeRange = newSelection.first;

                // 如果是相同的选项，不执行操作
                if (newTimeRange == _timeRange) return;

                setState(() {
                  _timeRange = newTimeRange;
                });

                _loadStatistics();
              },
              style: SegmentedButton.styleFrom(
                backgroundColor: Colors.transparent,
                selectedBackgroundColor: const Color(0xFF20CB6B),
                selectedForegroundColor: Colors.white,
                foregroundColor: const Color(0xFF8C92A4),
                side: BorderSide.none,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(6),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatCards() {
    if (_statsData == null) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(32.0),
          child: Text(
            '暂无数据',
            style: TextStyle(color: Color(0xFF8C92A4), fontSize: 14),
          ),
        ),
      );
    }

    final orderStats =
        _statsData!['order_stats'] as Map<String, dynamic>? ?? {};
    final revenueStats =
        _statsData!['revenue_stats'] as Map<String, dynamic>? ?? {};
    final userStats = _statsData!['user_stats'] as Map<String, dynamic>? ?? {};

    return GridView.count(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      crossAxisCount: 2,
      crossAxisSpacing: 16,
      mainAxisSpacing: 16,
      childAspectRatio: 1.5,
      children: [
        _buildStatCard(
          title: '订单总数',
          value: '${orderStats['total_orders'] ?? 0}',
          icon: Icons.shopping_cart_outlined,
          color: Colors.blue,
        ),
        _buildStatCard(
          title: '销售额',
          value: '¥${_formatNumber(revenueStats['total_revenue'] ?? 0.0)}',
          icon: Icons.attach_money,
          color: Colors.green,
        ),
        _buildStatCard(
          title: '待审核订单',
          value: '${orderStats['pending_delivery'] ?? 0}',
          icon: Icons.pending_actions,
          color: Colors.orange,
        ),
        _buildStatCard(
          title: '活跃用户',
          value: '${userStats['active_users'] ?? 0}',
          icon: Icons.people_outline,
          color: Colors.purple,
        ),
      ],
    );
  }

  Widget _buildStatCard({
    required String title,
    required String value,
    required IconData icon,
    required Color color,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: color.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(icon, color: color, size: 24),
              ),
            ],
          ),
          Flexible(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Flexible(
                  child: Text(
                    value,
                    style: TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                      color: color,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 12,
                    color: Color(0xFF8C92A4),
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  String _formatNumber(dynamic value) {
    if (value is num) {
      if (value >= 10000) {
        return '${(value / 10000).toStringAsFixed(1)}万';
      }
      return value.toStringAsFixed(2);
    }
    return '0.00';
  }
}
