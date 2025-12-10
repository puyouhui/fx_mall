import 'package:flutter/material.dart';
import '../api/order_api.dart';
import 'package:map_launcher/map_launcher.dart';
import 'batch_pickup_items_view.dart';

/// 批量取货页面
class BatchPickupView extends StatefulWidget {
  const BatchPickupView({super.key});

  @override
  State<BatchPickupView> createState() => _BatchPickupViewState();
}

class _BatchPickupViewState extends State<BatchPickupView> {
  bool _isLoadingSuppliers = true;
  List<dynamic> _suppliers = [];

  @override
  void initState() {
    super.initState();
    _loadSuppliers();
  }

  Future<void> _loadSuppliers() async {
    setState(() {
      _isLoadingSuppliers = true;
    });

    try {
      final response = await OrderApi.getPickupSuppliers();
      if (response.isSuccess && response.data != null) {
        setState(() {
          _suppliers = response.data!;
          _isLoadingSuppliers = false;
        });
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                response.message.isNotEmpty ? response.message : '获取供应商列表失败',
              ),
              backgroundColor: Colors.red,
            ),
          );
        }
      }
      setState(() {
        _isLoadingSuppliers = false;
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('获取供应商列表失败: $e'), backgroundColor: Colors.red),
        );
      }
      setState(() {
        _isLoadingSuppliers = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('批量取货'),
        backgroundColor: const Color(0xFF20CB6B),
        iconTheme: const IconThemeData(color: Colors.white),
        titleTextStyle: const TextStyle(
          color: Colors.white,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      body: _buildSuppliersList(),
    );
  }

  Widget _buildSuppliersList() {
    if (_isLoadingSuppliers) {
      return const Center(
        child: CircularProgressIndicator(
          valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
        ),
      );
    }

    if (_suppliers.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.inventory_2_outlined, size: 64, color: Colors.grey[400]),
            const SizedBox(height: 16),
            Text(
              '暂无待取货供应商',
              style: TextStyle(fontSize: 16, color: Colors.grey[600]),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadSuppliers,
      color: const Color(0xFF20CB6B),
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _suppliers.length,
        itemBuilder: (context, index) {
          final supplier = _suppliers[index];
          final name = supplier['name'] as String? ?? '';
          final address = supplier['address'] as String? ?? '';
          final phone = supplier['phone'] as String? ?? '';
          final latitude = supplier['latitude'] as double?;
          final longitude = supplier['longitude'] as double?;

          return Container(
            margin: const EdgeInsets.only(bottom: 16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.06),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Material(
              color: Colors.transparent,
              child: InkWell(
                onTap: () async {
                  // 跳转到商品列表和取货操作页面
                  final result = await Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (context) => BatchPickupItemsView(
                        supplierId: supplier['id'] as int,
                        supplierName: name,
                        supplierLatitude: latitude,
                        supplierLongitude: longitude,
                      ),
                    ),
                  );
                  // 如果取货成功，刷新供应商列表
                  // 无论是否成功，都刷新列表（确保数据最新）
                  if (mounted) {
                    _loadSuppliers();
                  }
                },
                borderRadius: BorderRadius.circular(16),
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Container(
                            width: 48,
                            height: 48,
                            decoration: BoxDecoration(
                              color: const Color(0xFFFF5722).withOpacity(0.1),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: const Icon(
                              Icons.store,
                              color: Color(0xFFFF5722),
                              size: 28,
                            ),
                          ),
                          const SizedBox(width: 16),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  name,
                                  style: const TextStyle(
                                    fontSize: 18,
                                    fontWeight: FontWeight.w700,
                                    color: Color(0xFF20253A),
                                  ),
                                ),
                                if (phone.isNotEmpty) ...[
                                  const SizedBox(height: 4),
                                  Row(
                                    children: [
                                      const Icon(
                                        Icons.phone_outlined,
                                        size: 14,
                                        color: Color(0xFF8C92A4),
                                      ),
                                      const SizedBox(width: 4),
                                      Text(
                                        phone,
                                        style: const TextStyle(
                                          fontSize: 13,
                                          color: Color(0xFF8C92A4),
                                        ),
                                      ),
                                    ],
                                  ),
                                ],
                              ],
                            ),
                          ),
                          if (latitude != null && longitude != null)
                            Container(
                              decoration: BoxDecoration(
                                color: const Color(0xFF20CB6B).withOpacity(0.1),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: IconButton(
                                icon: const Icon(
                                  Icons.navigation,
                                  color: Color(0xFF20CB6B),
                                  size: 22,
                                ),
                                onPressed: () => _navigateToSupplier(
                                  latitude,
                                  longitude,
                                  name,
                                ),
                                tooltip: '导航',
                              ),
                            ),
                          const Icon(
                            Icons.chevron_right,
                            color: Color(0xFF8C92A4),
                          ),
                        ],
                      ),
                      if (address.isNotEmpty) ...[
                        const SizedBox(height: 12),
                        Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: Colors.grey[50],
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Icon(
                                Icons.location_on_outlined,
                                size: 18,
                                color: Color(0xFF20CB6B),
                              ),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  address,
                                  style: const TextStyle(
                                    fontSize: 14,
                                    color: Color(0xFF40475C),
                                    height: 1.4,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  Future<void> _navigateToSupplier(
    double? latitude,
    double? longitude,
    String name,
  ) async {
    if (latitude == null || longitude == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('供应商地址信息不完整'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    try {
      final isAmapAvailable = await MapLauncher.isMapAvailable(MapType.amap);
      if (isAmapAvailable == true) {
        await MapLauncher.showDirections(
          mapType: MapType.amap,
          destination: Coords(latitude, longitude),
          destinationTitle: name,
        );
      } else {
        final availableMaps = await MapLauncher.installedMaps;
        if (availableMaps.isNotEmpty) {
          await availableMaps.first.showDirections(
            destination: Coords(latitude, longitude),
            destinationTitle: name,
          );
        } else {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('未安装地图应用，请先安装高德地图'),
                backgroundColor: Colors.orange,
              ),
            );
          }
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('打开导航失败: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }
}
