import 'package:flutter/material.dart';
import '../api/order_api.dart';
import 'package:map_launcher/map_launcher.dart';

/// 批量取货页面
class BatchPickupView extends StatefulWidget {
  const BatchPickupView({super.key});

  @override
  State<BatchPickupView> createState() => _BatchPickupViewState();
}

class _BatchPickupViewState extends State<BatchPickupView> {
  bool _isLoadingSuppliers = true;
  bool _isLoadingItems = false;
  bool _isMarkingPicked = false;
  List<dynamic> _suppliers = [];
  List<dynamic> _items = [];
  int? _selectedSupplierId;
  String? _selectedSupplierName;
  Set<int> _selectedItemIds = {};

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

  Future<void> _loadItems(int supplierId) async {
    setState(() {
      _isLoadingItems = true;
      _selectedItemIds.clear();
    });

    try {
      final response = await OrderApi.getPickupItemsBySupplier(supplierId);
      if (response.isSuccess && response.data != null) {
        setState(() {
          _items = response.data!;
          // 默认选中所有商品
          _selectedItemIds = _items.map((item) => item['id'] as int).toSet();
        });
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                response.message.isNotEmpty ? response.message : '获取商品列表失败',
              ),
              backgroundColor: Colors.red,
            ),
          );
        }
      }
      setState(() {
        _isLoadingItems = false;
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('获取商品列表失败: $e'), backgroundColor: Colors.red),
        );
      }
      setState(() {
        _isLoadingItems = false;
      });
    }
  }

  Future<void> _markItemsAsPicked() async {
    if (_selectedItemIds.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('请至少选择一个商品'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('确认取货'),
        content: Text('确认已取货 ${_selectedItemIds.length} 件商品？'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('确认'),
          ),
        ],
      ),
    );

    if (confirm != true) return;

    setState(() {
      _isMarkingPicked = true;
    });

    try {
      final response = await OrderApi.markItemsAsPicked(
        _selectedItemIds.toList(),
      );
      if (response.isSuccess) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('标记取货成功'),
              backgroundColor: Colors.green,
            ),
          );
        }
        // 重新加载商品列表
        if (_selectedSupplierId != null) {
          await _loadItems(_selectedSupplierId!);
        }
        // 重新加载供应商列表
        await _loadSuppliers();
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                response.message.isNotEmpty ? response.message : '标记取货失败',
              ),
              backgroundColor: Colors.red,
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('标记取货失败: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isMarkingPicked = false;
        });
      }
    }
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
      body: Column(
        children: [
          // 供应商列表或商品列表
          Expanded(
            child: _selectedSupplierId == null
                ? _buildSuppliersList()
                : _buildItemsList(),
          ),

          // 已全部取货按钮（仅在选择了供应商时显示）
          if (_selectedSupplierId != null && _items.isNotEmpty)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
              decoration: BoxDecoration(
                color: Colors.white,
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.08),
                    blurRadius: 12,
                    offset: const Offset(0, -4),
                  ),
                ],
              ),
              child: SafeArea(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (_selectedItemIds.length < _items.length)
                      Container(
                        margin: const EdgeInsets.only(bottom: 12),
                        padding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 10,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.orange[50],
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(
                            color: Colors.orange[200]!,
                            width: 1,
                          ),
                        ),
                        child: Row(
                          children: [
                            Icon(
                              Icons.info_outline,
                              size: 18,
                              color: Colors.orange[700],
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                '已选择 ${_selectedItemIds.length}/${_items.length} 件商品',
                                style: TextStyle(
                                  fontSize: 13,
                                  color: Colors.orange[700],
                                  fontWeight: FontWeight.w500,
                                ),
                              ),
                            ),
                            TextButton(
                              onPressed: () {
                                setState(() {
                                  _selectedItemIds = _items
                                      .map((item) => item['id'] as int)
                                      .toSet();
                                });
                              },
                              child: Text(
                                '全选',
                                style: TextStyle(
                                  fontSize: 13,
                                  color: Colors.orange[700],
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ElevatedButton(
                      onPressed: _isMarkingPicked || _selectedItemIds.isEmpty
                          ? null
                          : _markItemsAsPicked,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: _selectedItemIds.isEmpty
                            ? Colors.grey[300]
                            : const Color(0xFF20CB6B),
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 18),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16),
                        ),
                        elevation: _selectedItemIds.isEmpty ? 0 : 4,
                      ),
                      child: _isMarkingPicked
                          ? const SizedBox(
                              height: 22,
                              width: 22,
                              child: CircularProgressIndicator(
                                strokeWidth: 2.5,
                                valueColor: AlwaysStoppedAnimation<Color>(
                                  Colors.white,
                                ),
                              ),
                            )
                          : Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                const Icon(
                                  Icons.check_circle_outline,
                                  size: 22,
                                ),
                                const SizedBox(width: 8),
                                Text(
                                  '已全部取货 (${_selectedItemIds.length}件)',
                                  style: const TextStyle(
                                    fontSize: 17,
                                    fontWeight: FontWeight.w700,
                                    letterSpacing: 0.5,
                                  ),
                                ),
                              ],
                            ),
                    ),
                  ],
                ),
              ),
            ),
        ],
      ),
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
                onTap: () {
                  setState(() {
                    _selectedSupplierId = supplier['id'] as int?;
                    _selectedSupplierName = name;
                  });
                  _loadItems(_selectedSupplierId!);
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

  Widget _buildItemsList() {
    return Column(
      children: [
        // 返回按钮和供应商名称
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: Colors.white,
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.04),
                blurRadius: 4,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: Row(
            children: [
              Container(
                decoration: BoxDecoration(
                  color: Colors.grey[100],
                  borderRadius: BorderRadius.circular(8),
                ),
                child: IconButton(
                  icon: const Icon(Icons.arrow_back, color: Color(0xFF20253A)),
                  onPressed: () {
                    setState(() {
                      _selectedSupplierId = null;
                      _selectedSupplierName = null;
                      _items = [];
                      _selectedItemIds.clear();
                    });
                  },
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      _selectedSupplierName ?? '商品列表',
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.w700,
                        color: Color(0xFF20253A),
                      ),
                    ),
                    if (_items.isNotEmpty)
                      Text(
                        '共 ${_items.length} 件商品',
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
        ),

        // 商品列表
        Expanded(
          child: _isLoadingItems
              ? const Center(
                  child: CircularProgressIndicator(
                    valueColor: AlwaysStoppedAnimation<Color>(
                      Color(0xFF20CB6B),
                    ),
                  ),
                )
              : _items.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Container(
                        width: 120,
                        height: 120,
                        decoration: BoxDecoration(
                          color: Colors.grey[100],
                          shape: BoxShape.circle,
                        ),
                        child: Icon(
                          Icons.shopping_bag_outlined,
                          size: 64,
                          color: Colors.grey[400],
                        ),
                      ),
                      const SizedBox(height: 24),
                      Text(
                        '该供应商暂无待取货商品',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w600,
                          color: Colors.grey[700],
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        '所有商品已取货完成',
                        style: TextStyle(fontSize: 14, color: Colors.grey[500]),
                      ),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: _items.length,
                  itemBuilder: (context, index) {
                    final item = _items[index];
                    final itemId = item['id'] as int;
                    final productName = item['product_name'] as String? ?? '';
                    final specName = item['spec_name'] as String? ?? '';
                    final quantity = item['quantity'] as int? ?? 0;
                    final image = item['image'] as String? ?? '';
                    final orderNumber = item['order_number'] as String? ?? '';
                    final isSelected = _selectedItemIds.contains(itemId);

                    return Container(
                      margin: const EdgeInsets.only(bottom: 12),
                      decoration: BoxDecoration(
                        color: isSelected
                            ? const Color(0xFFE8F8F0)
                            : Colors.white,
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                          color: isSelected
                              ? const Color(0xFF20CB6B)
                              : Colors.grey[200]!,
                          width: isSelected ? 2 : 1,
                        ),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withOpacity(0.04),
                            blurRadius: 4,
                            offset: const Offset(0, 2),
                          ),
                        ],
                      ),
                      child: Material(
                        color: Colors.transparent,
                        child: InkWell(
                          onTap: () {
                            setState(() {
                              if (isSelected) {
                                _selectedItemIds.remove(itemId);
                              } else {
                                _selectedItemIds.add(itemId);
                              }
                            });
                          },
                          borderRadius: BorderRadius.circular(16),
                          child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                // 复选框
                                Container(
                                  width: 24,
                                  height: 24,
                                  margin: const EdgeInsets.only(top: 2),
                                  decoration: BoxDecoration(
                                    shape: BoxShape.circle,
                                    border: Border.all(
                                      color: isSelected
                                          ? const Color(0xFF20CB6B)
                                          : Colors.grey[400]!,
                                      width: 2,
                                    ),
                                    color: isSelected
                                        ? const Color(0xFF20CB6B)
                                        : Colors.transparent,
                                  ),
                                  child: isSelected
                                      ? const Icon(
                                          Icons.check,
                                          size: 16,
                                          color: Colors.white,
                                        )
                                      : null,
                                ),
                                const SizedBox(width: 12),
                                // 商品图片
                                ClipRRect(
                                  borderRadius: BorderRadius.circular(12),
                                  child: image.isNotEmpty
                                      ? Image.network(
                                          image,
                                          width: 80,
                                          height: 80,
                                          fit: BoxFit.cover,
                                          errorBuilder:
                                              (context, error, stackTrace) {
                                                return Container(
                                                  width: 80,
                                                  height: 80,
                                                  color: Colors.grey[200],
                                                  child: const Icon(
                                                    Icons.image_not_supported,
                                                    color: Colors.grey,
                                                  ),
                                                );
                                              },
                                        )
                                      : Container(
                                          width: 80,
                                          height: 80,
                                          decoration: BoxDecoration(
                                            color: Colors.grey[200],
                                            borderRadius: BorderRadius.circular(
                                              12,
                                            ),
                                          ),
                                          child: const Icon(
                                            Icons.image_not_supported,
                                            color: Colors.grey,
                                          ),
                                        ),
                                ),
                                const SizedBox(width: 12),
                                // 商品信息
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        productName,
                                        style: const TextStyle(
                                          fontSize: 16,
                                          fontWeight: FontWeight.w600,
                                          color: Color(0xFF20253A),
                                        ),
                                        maxLines: 2,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                      if (specName.isNotEmpty) ...[
                                        const SizedBox(height: 6),
                                        Container(
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 8,
                                            vertical: 4,
                                          ),
                                          decoration: BoxDecoration(
                                            color: const Color(
                                              0xFF20CB6B,
                                            ).withOpacity(0.1),
                                            borderRadius: BorderRadius.circular(
                                              6,
                                            ),
                                          ),
                                          child: Text(
                                            specName,
                                            style: const TextStyle(
                                              fontSize: 12,
                                              color: Color(0xFF20CB6B),
                                              fontWeight: FontWeight.w500,
                                            ),
                                          ),
                                        ),
                                      ],
                                      const SizedBox(height: 8),
                                      Row(
                                        children: [
                                          Container(
                                            padding: const EdgeInsets.symmetric(
                                              horizontal: 8,
                                              vertical: 4,
                                            ),
                                            decoration: BoxDecoration(
                                              color: Colors.blue[50],
                                              borderRadius:
                                                  BorderRadius.circular(6),
                                            ),
                                            child: Text(
                                              '数量: $quantity',
                                              style: TextStyle(
                                                fontSize: 12,
                                                color: Colors.blue[700],
                                                fontWeight: FontWeight.w600,
                                              ),
                                            ),
                                          ),
                                          const SizedBox(width: 8),
                                          Expanded(
                                            child: Text(
                                              '订单: $orderNumber',
                                              style: const TextStyle(
                                                fontSize: 12,
                                                color: Color(0xFF8C92A4),
                                              ),
                                              overflow: TextOverflow.ellipsis,
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
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }
}
