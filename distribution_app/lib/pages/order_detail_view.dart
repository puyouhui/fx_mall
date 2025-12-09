import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_map_location_marker/flutter_map_location_marker.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'package:map_launcher/map_launcher.dart';
import 'dart:async';
import '../api/order_api.dart';
import '../utils/location_service.dart';

/// 订单详情页面：显示订单的完整信息，包括配送费、加急状态等
class OrderDetailView extends StatefulWidget {
  const OrderDetailView({super.key, required this.orderId});

  final int orderId;

  @override
  State<OrderDetailView> createState() => _OrderDetailViewState();
}

class _OrderDetailViewState extends State<OrderDetailView> {
  bool _isLoading = true;
  Map<String, dynamic>? _orderData;
  String? _errorMessage;
  bool _isProcessing = false; // 是否正在处理操作（接单/完成配送/问题上报）

  // 地图相关
  final MapController _mapController = MapController();
  Position? _userPosition;
  final StreamController<LocationMarkerPosition?> _locationStreamController =
      StreamController<LocationMarkerPosition?>.broadcast();
  StreamSubscription<Position>? _positionStreamSubscription;

  // 天地图瓦片服务 URL 模板（Web墨卡托投影 vec_w，因为 flutter_map 对经纬度投影支持有限）
  // 使用 Web 墨卡托投影，与路线规划页面保持一致，确保地图能正常显示
  static const String _tiandituTileUrlTemplate =
      'https://t{s}.tianditu.gov.cn/vec_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=vec&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=d95864378581051adb04fe26acb13ecf';
  static const String _tiandituLabelUrlTemplate =
      'https://t{s}.tianditu.gov.cn/cva_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=cva&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=d95864378581051adb04fe26acb13ecf';

  static TileProvider createTiandituTileProvider() {
    return NetworkTileProvider(
      headers: {
        'User-Agent':
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Referer': 'https://lbs.tianditu.gov.cn/',
        'Accept': 'image/webp,image/apng,image/*,*/*;q=0.8',
        'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8',
      },
    );
  }

  @override
  void initState() {
    super.initState();
    _loadOrderDetail();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _startLocationTracking();
    });
  }

  @override
  void dispose() {
    _positionStreamSubscription?.cancel();
    _locationStreamController.close();
    _mapController.dispose();
    super.dispose();
  }

  Future<void> _startLocationTracking() async {
    final hasPermission = await LocationService.checkAndRequestPermission();
    if (!hasPermission) {
      print('[OrderDetailView] 没有定位权限，无法启动位置跟踪');
      return;
    }

    final serviceEnabled = await LocationService.checkLocationServiceEnabled();

    try {
      // 即使定位服务未启用，也尝试使用网络定位
      if (!serviceEnabled) {
        print('[OrderDetailView] 定位服务未启用，尝试使用网络定位...');
        final networkPosition = await LocationService.getCurrentLocation();
        if (networkPosition != null && mounted) {
          print('[OrderDetailView] 网络定位成功，继续使用网络定位');
          // 立即将初始位置发送到流中，以便 CurrentLocationLayer 能显示
          _locationStreamController.add(
            LocationMarkerPosition(
              latitude: networkPosition.latitude,
              longitude: networkPosition.longitude,
              accuracy: networkPosition.accuracy,
            ),
          );
          setState(() {
            _userPosition = networkPosition;
          });
          // 网络定位成功，启动定位流（使用低精度）
          _startPositionStreamWithFallback();
          return;
        } else {
          print('[OrderDetailView] 网络定位也失败，无法启动位置跟踪');
          return;
        }
      }

      // 定位服务已启用，正常启动定位流
      // 使用多级精度策略启动定位流，从高精度开始，如果失败则降级
      _startPositionStreamWithFallback();

      final initialPosition = await LocationService.getCurrentLocation();
      if (initialPosition != null && mounted) {
        // 立即将初始位置发送到流中，以便 CurrentLocationLayer 能显示
        _locationStreamController.add(
          LocationMarkerPosition(
            latitude: initialPosition.latitude,
            longitude: initialPosition.longitude,
            accuracy: initialPosition.accuracy,
          ),
        );
        setState(() {
          _userPosition = initialPosition;
        });
      }
    } catch (e) {
      print('[OrderDetailView] 启动位置跟踪失败: $e');
      // 错误处理：尝试使用低精度重新启动
      if (mounted) {
        Future.delayed(const Duration(seconds: 2), () {
          if (mounted) {
            _startPositionStreamWithFallback();
          }
        });
      }
    }
  }

  /// 启动定位流（带降级策略，优先网络定位）
  void _startPositionStreamWithFallback() {
    // 优先使用网络定位（低精度），在中国更可靠
    _tryStartPositionStream(LocationAccuracy.low, () {
      // 如果低精度失败，尝试最低精度
      print('[OrderDetailView] 网络定位流失败，尝试最低精度');
      Future.delayed(const Duration(seconds: 1), () {
        if (mounted) {
          _tryStartPositionStream(LocationAccuracy.lowest, () {
            // 如果最低精度也失败，尝试中等精度（GPS + 网络）
            print('[OrderDetailView] 最低精度定位流失败，尝试中等精度（GPS+网络）');
            Future.delayed(const Duration(seconds: 1), () {
              if (mounted) {
                _tryStartPositionStream(LocationAccuracy.medium, () {
                  // 最后尝试高精度GPS（在中国可能失败）
                  print('[OrderDetailView] 中等精度定位流失败，尝试高精度GPS');
                  Future.delayed(const Duration(seconds: 1), () {
                    if (mounted) {
                      _tryStartPositionStream(LocationAccuracy.high, () {
                        print('[OrderDetailView] 所有精度级别都失败，定位流无法启动');
                      });
                    }
                  });
                });
              }
            });
          });
        }
      });
    });
  }

  /// 尝试启动指定精度的定位流
  void _tryStartPositionStream(
    LocationAccuracy accuracy,
    VoidCallback onError,
  ) {
    try {
      _positionStreamSubscription?.cancel();
      _positionStreamSubscription =
          Geolocator.getPositionStream(
            locationSettings: LocationSettings(
              accuracy: accuracy,
              distanceFilter: 10,
              // 注意：LocationSettings 不支持 forceAndroidLocationManager 参数
              // 但可以通过其他方式优化
            ),
          ).listen(
            (Position position) {
              _locationStreamController.add(
                LocationMarkerPosition(
                  latitude: position.latitude,
                  longitude: position.longitude,
                  accuracy: position.accuracy,
                ),
              );

              if (mounted) {
                setState(() {
                  _userPosition = position;
                });
              }

              // 如果有客户地址，调整地图视野以同时显示两个位置
              final addressData =
                  _orderData?['address'] as Map<String, dynamic>?;
              final customerLat = addressData?['latitude'] as num?;
              final customerLng = addressData?['longitude'] as num?;
              if (customerLat != null &&
                  customerLng != null &&
                  _userPosition != null) {
                _adjustMapBounds(
                  LatLng(customerLat.toDouble(), customerLng.toDouble()),
                  LatLng(_userPosition!.latitude, _userPosition!.longitude),
                );
              }
            },
            onError: (error) {
              print('[OrderDetailView] 定位流错误 (精度: $accuracy): $error');
              // 如果当前精度失败，尝试降级
              onError();
            },
            cancelOnError: false, // 不因错误而取消流
          );
      print('[OrderDetailView] 定位流启动成功 (精度: $accuracy)');
    } catch (e) {
      print('[OrderDetailView] 启动定位流失败 (精度: $accuracy): $e');
      onError();
    }
  }

  void _adjustMapBounds(LatLng customerPos, LatLng userPos) {
    // 计算两个位置的中心点和合适的缩放级别
    final centerLat = (customerPos.latitude + userPos.latitude) / 2;
    final centerLng = (customerPos.longitude + userPos.longitude) / 2;
    final center = LatLng(centerLat, centerLng);

    // 计算距离以确定合适的缩放级别
    final distance = const Distance().as(
      LengthUnit.Meter,
      customerPos,
      userPos,
    );
    double zoom = 15.0;
    if (distance > 5000) {
      zoom = 12.0;
    } else if (distance > 2000) {
      zoom = 13.0;
    } else if (distance > 1000) {
      zoom = 14.0;
    }

    _mapController.move(center, zoom);
  }

  Future<void> _loadOrderDetail() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final response = await OrderApi.getOrderDetail(widget.orderId);

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      setState(() {
        _orderData = response.data;
        _isLoading = false;
      });
    } else {
      setState(() {
        _errorMessage = response.message.isNotEmpty
            ? response.message
            : '获取订单详情失败';
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_getAppBarTitle()),
        backgroundColor: const Color(0xFF20CB6B),
        iconTheme: const IconThemeData(color: Colors.white),
        titleTextStyle: const TextStyle(
          color: Colors.white,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      backgroundColor: const Color(0xFFF5F5F5),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(
                valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
              ),
            )
          : _errorMessage != null
          ? Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    _errorMessage!,
                    style: const TextStyle(
                      fontSize: 14,
                      color: Color(0xFF8C92A4),
                    ),
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: _loadOrderDetail,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF20CB6B),
                      foregroundColor: Colors.white,
                    ),
                    child: const Text('重试'),
                  ),
                ],
              ),
            )
          : _orderData == null
          ? const Center(
              child: Text(
                '订单数据为空',
                style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
              ),
            )
          : Stack(
              children: [
                RefreshIndicator(
                  onRefresh: _loadOrderDetail,
                  child: SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: EdgeInsets.zero,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // 地图（顶部）
                        _buildMapCard(),
                        const SizedBox(height: 12),
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 16),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              // 待取货状态时，供应商列表优先显示在地图下面
                              if (_isPendingPickup()) ...[
                                _buildSuppliersCard(),
                                const SizedBox(height: 12),
                              ],
                              // 地址信息
                              _buildAddressCard(),
                              const SizedBox(height: 12),
                              // 商品列表
                              _buildItemsCard(),
                              const SizedBox(height: 12),
                              // 非待取货状态时，供应商列表显示在商品列表下面
                              if (!_isPendingPickup()) ...[
                                _buildSuppliersCard(),
                                const SizedBox(height: 12),
                              ],
                              // 配送费信息
                              _buildDeliveryFeeCard(),
                              const SizedBox(height: 12),
                              // 加急状态
                              if (_isUrgent()) _buildUrgentCard(),
                              const SizedBox(height: 12),
                              // 订单基本信息（移到最下方）
                              _buildOrderInfoCard(),
                              // 底部留出空间给操作按钮（增加padding避免被遮挡）
                              const SizedBox(height: 120),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                // 底部操作栏
                Positioned(
                  bottom: 0,
                  left: 0,
                  right: 0,
                  child: _buildActionBar(),
                ),
              ],
            ),
    );
  }

  Widget _buildMapCard() {
    final addressData = _orderData?['address'] as Map<String, dynamic>?;
    final customerLat = addressData?['latitude'] as num?;
    final customerLng = addressData?['longitude'] as num?;

    // 获取供应商列表（仅在待取货状态时）
    final suppliers = _isPendingPickup()
        ? (_orderData?['suppliers'] as List<dynamic>?)
                  ?.cast<Map<String, dynamic>>() ??
              []
        : [];

    // 确定地图初始中心点
    LatLng initialCenter = const LatLng(39.90864, 116.39750); // 默认北京
    if (customerLat != null && customerLng != null) {
      initialCenter = LatLng(customerLat.toDouble(), customerLng.toDouble());
    } else if (_userPosition != null) {
      initialCenter = LatLng(_userPosition!.latitude, _userPosition!.longitude);
    }

    return Container(
      height: 250,
      decoration: BoxDecoration(color: Colors.white),
      child: ClipRRect(
        child: Stack(
          children: [
            FlutterMap(
              mapController: _mapController,
              options: MapOptions(
                initialCenter: initialCenter,
                initialZoom:
                    customerLat != null &&
                        customerLng != null &&
                        _userPosition != null
                    ? 13.0
                    : 15.0,
                minZoom: 3.0,
                maxZoom: 18.0,
                // 使用默认的 Web 墨卡托投影（EPSG:3857），与路线规划页面保持一致
                // 注意：flutter_map 对 WMTS 经纬度投影的支持有限，使用 Web 墨卡托投影更稳定
              ),
              children: [
                // 天地图矢量底图（经纬度投影）
                TileLayer(
                  urlTemplate: _tiandituTileUrlTemplate,
                  subdomains: const ['0', '1', '2', '3', '4', '5', '6', '7'],
                  userAgentPackageName: 'com.example.distribution_app',
                  maxNativeZoom: 18,
                  maxZoom: 18,
                  tileProvider: createTiandituTileProvider(),
                ),
                // 天地图矢量标注图层（经纬度投影）
                TileLayer(
                  urlTemplate: _tiandituLabelUrlTemplate,
                  subdomains: const ['0', '1', '2', '3', '4', '5', '6', '7'],
                  userAgentPackageName: 'com.example.distribution_app',
                  maxNativeZoom: 18,
                  maxZoom: 18,
                  tileProvider: createTiandituTileProvider(),
                ),
                // 客户位置标记
                if (customerLat != null && customerLng != null)
                  MarkerLayer(
                    markers: [
                      Marker(
                        point: LatLng(
                          customerLat.toDouble(),
                          customerLng.toDouble(),
                        ),
                        width: 120,
                        height: 65,
                        alignment: Alignment.topCenter,
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          mainAxisAlignment: MainAxisAlignment.start,
                          children: [
                            // 文本标签
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 6,
                                vertical: 2,
                              ),
                              decoration: BoxDecoration(
                                color: const Color(0xFFFF6B6B),
                                borderRadius: BorderRadius.circular(4),
                                boxShadow: [
                                  BoxShadow(
                                    color: Colors.black.withOpacity(0.2),
                                    blurRadius: 4,
                                    offset: const Offset(0, 2),
                                  ),
                                ],
                              ),
                              child: const Text(
                                '客户',
                                style: TextStyle(
                                  fontSize: 11,
                                  fontWeight: FontWeight.w600,
                                  color: Colors.white,
                                ),
                              ),
                            ),
                            const SizedBox(height: 2),
                            // 图标
                            Container(
                              width: 40,
                              height: 40,
                              decoration: BoxDecoration(
                                color: const Color(0xFFFF6B6B),
                                shape: BoxShape.circle,
                                border: Border.all(
                                  color: Colors.white,
                                  width: 3,
                                ),
                                boxShadow: [
                                  BoxShadow(
                                    color: Colors.black.withOpacity(0.3),
                                    blurRadius: 6,
                                    offset: const Offset(0, 2),
                                  ),
                                ],
                              ),
                              child: const Icon(
                                Icons.location_on,
                                color: Colors.white,
                                size: 24,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                // 供应商位置标记（仅在待取货状态时显示）
                if (_isPendingPickup() && suppliers.isNotEmpty)
                  MarkerLayer(
                    markers: suppliers
                        .where((supplier) {
                          final lat = supplier['latitude'] as num?;
                          final lng = supplier['longitude'] as num?;
                          return lat != null && lng != null;
                        })
                        .map((supplier) {
                          final lat = supplier['latitude'] as num?;
                          final lng = supplier['longitude'] as num?;
                          final name = supplier['name'] as String? ?? '';
                          return Marker(
                            point: LatLng(lat!.toDouble(), lng!.toDouble()),
                            width: 120,
                            height: 65,
                            alignment: Alignment.topCenter,
                            child: Tooltip(
                              message: name,
                              child: Column(
                                mainAxisSize: MainAxisSize.min,
                                mainAxisAlignment: MainAxisAlignment.start,
                                children: [
                                  // 文本标签
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 6,
                                      vertical: 2,
                                    ),
                                    decoration: BoxDecoration(
                                      color: const Color(0xFFFF5722),
                                      borderRadius: BorderRadius.circular(4),
                                      boxShadow: [
                                        BoxShadow(
                                          color: Colors.black.withOpacity(0.2),
                                          blurRadius: 4,
                                          offset: const Offset(0, 2),
                                        ),
                                      ],
                                    ),
                                    child: Text(
                                      name.length > 6
                                          ? '${name.substring(0, 6)}...'
                                          : name,
                                      style: const TextStyle(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w600,
                                        color: Colors.white,
                                      ),
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  ),
                                  const SizedBox(height: 2),
                                  // 图标（确保居中）
                                  Container(
                                    width: 40,
                                    height: 40,
                                    decoration: BoxDecoration(
                                      color: const Color(0xFFFF5722),
                                      shape: BoxShape.circle,
                                      border: Border.all(
                                        color: Colors.white,
                                        width: 3,
                                      ),
                                      boxShadow: [
                                        BoxShadow(
                                          color: Colors.black.withOpacity(0.3),
                                          blurRadius: 6,
                                          offset: const Offset(0, 2),
                                        ),
                                      ],
                                    ),
                                    child: const Icon(
                                      Icons.warehouse,
                                      color: Colors.white,
                                      size: 24,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          );
                        })
                        .toList(),
                  ),
                // 配送员位置标记（使用 CurrentLocationLayer）
                CurrentLocationLayer(
                  positionStream: _locationStreamController.stream,
                ),
                // 版权信息
                RichAttributionWidget(
                  attributions: [TextSourceAttribution('天地图', onTap: () {})],
                ),
              ],
            ),
            // 图例说明
            Positioned(
              bottom: 8,
              left: 8,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.9),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(
                      Icons.location_on,
                      size: 16,
                      color: Color(0xFFFF6B6B),
                    ),
                    const SizedBox(width: 4),
                    const Text(
                      '客户',
                      style: TextStyle(fontSize: 12, color: Color(0xFF20253A)),
                    ),
                    // 待取货状态时显示供应商图例
                    if (_isPendingPickup() && suppliers.isNotEmpty) ...[
                      const SizedBox(width: 12),
                      Container(
                        width: 16,
                        height: 16,
                        decoration: const BoxDecoration(
                          color: Color(0xFFFF5722),
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(
                          Icons.warehouse,
                          size: 10,
                          color: Colors.white,
                        ),
                      ),
                      const SizedBox(width: 4),
                      const Text(
                        '供应商',
                        style: TextStyle(
                          fontSize: 12,
                          color: Color(0xFF20253A),
                        ),
                      ),
                    ],
                    const SizedBox(width: 12),
                    const Icon(
                      Icons.my_location,
                      size: 16,
                      color: Color(0xFF20CB6B),
                    ),
                    const SizedBox(width: 4),
                    const Text(
                      '我的位置',
                      style: TextStyle(fontSize: 12, color: Color(0xFF20253A)),
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

  Widget _buildOrderInfoCard() {
    final order = _orderData?['order'] as Map<String, dynamic>?;
    final orderNumber = order?['order_number'] as String? ?? '';
    final status = order?['status'] as String? ?? '';
    final createdAt = order?['created_at'] as String? ?? '';

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Text(
                '订单编号：',
                style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
              ),
              Text(
                orderNumber,
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              const Text(
                '订单状态：',
                style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: _getStatusColor(status).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  _formatStatus(status),
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: _getStatusColor(status),
                  ),
                ),
              ),
            ],
          ),
          if (createdAt.isNotEmpty) ...[
            const SizedBox(height: 8),
            Row(
              children: [
                const Text(
                  '下单时间：',
                  style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
                ),
                Text(
                  createdAt,
                  style: const TextStyle(
                    fontSize: 14,
                    color: Color(0xFF20253A),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildAddressCard() {
    final addressData = _orderData?['address'] as Map<String, dynamic>?;
    final name = addressData?['name'] as String? ?? '';
    final address = addressData?['address'] as String? ?? '';
    final contact = addressData?['contact'] as String? ?? '';
    final phone = addressData?['phone'] as String? ?? '';
    final customerLat = addressData?['latitude'] as num?;
    final customerLng = addressData?['longitude'] as num?;

    // 获取订单状态，只有接单后（delivering）才显示联系人和电话
    final order = _orderData?['order'] as Map<String, dynamic>?;
    final status = order?['status'] as String? ?? '';
    final isDelivering = status == 'delivering';

    // 点击地址卡片时，放大地图并移动到客户位置
    void _onAddressCardTap() {
      if (customerLat != null && customerLng != null) {
        final customerLocation = LatLng(
          customerLat.toDouble(),
          customerLng.toDouble(),
        );
        // 移动到客户位置并放大到16级（更详细的视图）
        _mapController.move(customerLocation, 16.0);
      }
    }

    return GestureDetector(
      onTap: _onAddressCardTap,
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 8,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(
                  Icons.location_on,
                  size: 18,
                  color: Color(0xFF20CB6B),
                ),
                const SizedBox(width: 6),
                const Text(
                  '收货地址',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                ),
                const Spacer(),
                if (customerLat != null && customerLng != null)
                  const Icon(Icons.zoom_in, size: 16, color: Color(0xFF8C92A4)),
              ],
            ),
            const SizedBox(height: 12),
            if (name.isNotEmpty)
              Text(
                name,
                style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w500,
                  color: Color(0xFF40475C),
                ),
              ),
            if (address.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                address,
                style: const TextStyle(
                  fontSize: 14,
                  color: Color(0xFF20253A),
                  height: 1.5,
                ),
              ),
            ],
            // 只有接单后（delivering状态）才显示联系人和电话
            if (isDelivering && (contact.isNotEmpty || phone.isNotEmpty)) ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  const Icon(
                    Icons.person_outline,
                    size: 14,
                    color: Color(0xFF8C92A4),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    '$contact $phone',
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
    );
  }

  Widget _buildItemsCard() {
    final items =
        (_orderData?['order_items'] as List<dynamic>?)
            ?.cast<Map<String, dynamic>>() ??
        [];

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(
                Icons.shopping_cart_outlined,
                size: 18,
                color: Color(0xFF20CB6B),
              ),
              SizedBox(width: 6),
              Text(
                '商品列表',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          if (items.isEmpty)
            const Text(
              '暂无商品信息',
              style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
            )
          else
            ...items.map((item) => _buildItemRow(item)),
        ],
      ),
    );
  }

  Widget _buildItemRow(Map<String, dynamic> item) {
    final name = item['product_name'] as String? ?? '';
    final spec = item['spec_name'] as String? ?? '';
    final quantity = (item['quantity'] as num?)?.toInt() ?? 0;
    final image = item['image'] as String? ?? '';
    final isPicked = (item['is_picked'] as bool?) ?? false;
    final status = _orderData?['status'] as String? ?? '';
    final showPickupStatus =
        status == 'pending_pickup' || status == 'delivering';

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: isPicked && showPickupStatus
            ? const Color(0xFFE8F8F0).withOpacity(0.5)
            : const Color(0xFFF8F9FA),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: isPicked && showPickupStatus
              ? const Color(0xFF20CB6B).withOpacity(0.3)
              : const Color(0xFFE5E7EB),
          width: isPicked && showPickupStatus ? 2 : 1,
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 商品图片
          if (image.isNotEmpty)
            ClipRRect(
              borderRadius: BorderRadius.circular(6),
              child: Image.network(
                image,
                width: 70,
                height: 70,
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) {
                  return Container(
                    width: 70,
                    height: 70,
                    decoration: BoxDecoration(
                      color: const Color(0xFFE5E7EB),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: const Icon(
                      Icons.image_not_supported,
                      size: 30,
                      color: Color(0xFF8C92A4),
                    ),
                  );
                },
                loadingBuilder: (context, child, loadingProgress) {
                  if (loadingProgress == null) return child;
                  return Container(
                    width: 70,
                    height: 70,
                    decoration: BoxDecoration(
                      color: const Color(0xFFE5E7EB),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: const Center(
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        valueColor: AlwaysStoppedAnimation<Color>(
                          Color(0xFF20CB6B),
                        ),
                      ),
                    ),
                  );
                },
              ),
            )
          else
            Container(
              width: 70,
              height: 70,
              decoration: BoxDecoration(
                color: const Color(0xFFE5E7EB),
                borderRadius: BorderRadius.circular(6),
              ),
              child: const Icon(
                Icons.image_not_supported,
                size: 30,
                color: Color(0xFF8C92A4),
              ),
            ),
          const SizedBox(width: 12),
          // 商品信息
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  style: const TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 8),
                // 规格和数量并排显示
                Row(
                  children: [
                    // 规格标签
                    if (spec.isNotEmpty) ...[
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: const Color(0xFF20CB6B).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(6),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(
                              Icons.inventory_2_outlined,
                              size: 14,
                              color: Color(0xFF20CB6B),
                            ),
                            const SizedBox(width: 4),
                            Text(
                              spec,
                              style: const TextStyle(
                                fontSize: 13,
                                fontWeight: FontWeight.w600,
                                color: Color(0xFF20CB6B),
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 8),
                    ],
                    // 数量标签
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFF40475C).withOpacity(0.1),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(
                            Icons.numbers,
                            size: 14,
                            color: Color(0xFF40475C),
                          ),
                          const SizedBox(width: 4),
                          Text(
                            '$quantity 件',
                            style: const TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.w600,
                              color: Color(0xFF40475C),
                            ),
                          ),
                        ],
                      ),
                    ),
                    // 取货状态标签（仅在待取货或配送中状态显示）
                    if (showPickupStatus && isPicked) ...[
                      const SizedBox(width: 8),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: const Color(0xFF20CB6B).withOpacity(0.2),
                          borderRadius: BorderRadius.circular(6),
                        ),
                        child: const Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              Icons.check_circle_outline,
                              size: 14,
                              color: Color(0xFF20CB6B),
                            ),
                            SizedBox(width: 4),
                            Text(
                              '已取货',
                              style: TextStyle(
                                fontSize: 13,
                                fontWeight: FontWeight.w600,
                                color: Color(0xFF20CB6B),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSuppliersCard() {
    final suppliers =
        (_orderData?['suppliers'] as List<dynamic>?)
            ?.cast<Map<String, dynamic>>() ??
        [];

    if (suppliers.isEmpty) {
      return const SizedBox.shrink();
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(Icons.store_outlined, size: 18, color: Color(0xFF20CB6B)),
              SizedBox(width: 6),
              Text(
                '取货地址',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          ...suppliers.map((supplier) => _buildSupplierRow(supplier)),
        ],
      ),
    );
  }

  Widget _buildSupplierRow(Map<String, dynamic> supplier) {
    final name = supplier['name'] as String? ?? '';
    final address = supplier['address'] as String? ?? '';
    final contact = supplier['contact'] as String? ?? '';
    final phone = supplier['phone'] as String? ?? '';
    final latitude = supplier['latitude'] as double?;
    final longitude = supplier['longitude'] as double?;
    final items =
        (supplier['items'] as List<dynamic>?)?.cast<Map<String, dynamic>>() ??
        [];

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF8F9FA),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 供应商名称和导航图标
          Row(
            children: [
              const Icon(Icons.business, size: 16, color: Color(0xFF20CB6B)),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  name,
                  style: const TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                ),
              ),
              // 导航图标（如果有经纬度）
              if (latitude != null && longitude != null)
                InkWell(
                  onTap: () => _navigateToSupplier(latitude, longitude, name),
                  child: Container(
                    padding: const EdgeInsets.all(6),
                    decoration: BoxDecoration(
                      color: const Color(0xFF20CB6B).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: const Icon(
                      Icons.navigation,
                      size: 18,
                      color: Color(0xFF20CB6B),
                    ),
                  ),
                ),
            ],
          ),
          // 地址
          if (address.isNotEmpty) ...[
            const SizedBox(height: 8),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(
                  Icons.location_on_outlined,
                  size: 14,
                  color: Color(0xFF8C92A4),
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    address,
                    style: const TextStyle(
                      fontSize: 13,
                      color: Color(0xFF40475C),
                      height: 1.5,
                    ),
                  ),
                ),
              ],
            ),
          ],
          // 联系人和电话
          if (contact.isNotEmpty || phone.isNotEmpty) ...[
            const SizedBox(height: 6),
            Row(
              children: [
                const Icon(
                  Icons.phone_outlined,
                  size: 14,
                  color: Color(0xFF8C92A4),
                ),
                const SizedBox(width: 6),
                Text(
                  contact.isNotEmpty && phone.isNotEmpty
                      ? '$contact $phone'
                      : contact.isNotEmpty
                      ? contact
                      : phone,
                  style: const TextStyle(
                    fontSize: 13,
                    color: Color(0xFF8C92A4),
                  ),
                ),
              ],
            ),
          ],
          // 取货商品列表
          if (items.isNotEmpty) ...[
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(6),
                border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Row(
                    children: [
                      Icon(
                        Icons.shopping_bag_outlined,
                        size: 14,
                        color: Color(0xFF20CB6B),
                      ),
                      SizedBox(width: 6),
                      Text(
                        '取货商品',
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF20253A),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  ...items.map((item) => _buildSupplierItemRow(item)),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  // 构建供应商商品行
  Widget _buildSupplierItemRow(Map<String, dynamic> item) {
    final productName = item['product_name'] as String? ?? '';
    final specName = item['spec_name'] as String? ?? '';
    final quantity = (item['quantity'] as num?)?.toInt() ?? 0;
    final isPicked = (item['is_picked'] as bool?) ?? false;

    return Container(
      margin: const EdgeInsets.only(bottom: 6),
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
      decoration: BoxDecoration(
        color: isPicked
            ? const Color(0xFFE8F8F0).withOpacity(0.5)
            : const Color(0xFFF8F9FA),
        borderRadius: BorderRadius.circular(4),
        border: isPicked
            ? Border.all(
                color: const Color(0xFF20CB6B).withOpacity(0.3),
                width: 1,
              )
            : null,
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // 商品名称
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        productName,
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w500,
                          color: isPicked
                              ? const Color(0xFF20CB6B)
                              : const Color(0xFF20253A),
                        ),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    if (isPicked) ...[
                      const SizedBox(width: 6),
                      const Icon(
                        Icons.check_circle,
                        size: 16,
                        color: Color(0xFF20CB6B),
                      ),
                    ],
                  ],
                ),
                // 规格和数量
                if (specName.isNotEmpty || quantity > 0) ...[
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      // 规格
                      if (specName.isNotEmpty) ...[
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 6,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: const Color(0xFF20CB6B).withOpacity(0.1),
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            specName,
                            style: const TextStyle(
                              fontSize: 11,
                              color: Color(0xFF20CB6B),
                            ),
                          ),
                        ),
                        const SizedBox(width: 6),
                      ],
                      // 数量
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 6,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: const Color(0xFF40475C).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          '数量: $quantity',
                          style: const TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF40475C),
                          ),
                        ),
                      ),
                      // 取货状态
                      if (isPicked) ...[
                        const SizedBox(width: 6),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 6,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: const Color(0xFF20CB6B).withOpacity(0.2),
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: const Text(
                            '已取货',
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.w600,
                              color: Color(0xFF20CB6B),
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  // 导航到供应商位置
  Future<void> _navigateToSupplier(
    double latitude,
    double longitude,
    String name,
  ) async {
    try {
      // 检查是否安装了高德地图
      final isAmapAvailable = await MapLauncher.isMapAvailable(MapType.amap);
      if (isAmapAvailable == true) {
        // 使用高德地图导航
        await MapLauncher.showDirections(
          mapType: MapType.amap,
          destination: Coords(latitude, longitude),
          destinationTitle: name,
        );
      } else {
        // 如果没有高德地图，检查其他可用的地图应用
        final availableMaps = await MapLauncher.installedMaps;
        if (availableMaps.isNotEmpty) {
          // 使用第一个可用的地图应用
          await availableMaps.first.showDirections(
            destination: Coords(latitude, longitude),
            destinationTitle: name,
          );
        } else {
          // 如果没有安装任何地图应用，显示提示
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('未安装地图应用，请先安装高德地图'),
                duration: Duration(seconds: 2),
              ),
            );
          }
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('打开导航失败: $e'),
            duration: const Duration(seconds: 2),
          ),
        );
      }
    }
  }

  Widget _buildDeliveryFeeCard() {
    final deliveryFeeCalc =
        _orderData?['delivery_fee_calculation'] as Map<String, dynamic>?;

    // 如果数据不存在，尝试从订单数据中获取
    if (deliveryFeeCalc == null || deliveryFeeCalc.isEmpty) {
      // 如果订单中没有配送费计算数据，显示提示信息
      return Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 8,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: const Row(
          children: [
            Icon(Icons.local_shipping, size: 18, color: Color(0xFF8C92A4)),
            SizedBox(width: 6),
            Text(
              '配送费信息暂未计算',
              style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
            ),
          ],
        ),
      );
    }

    final baseFee = (deliveryFeeCalc['base_fee'] as num?)?.toDouble() ?? 0.0;
    final isolatedFee =
        (deliveryFeeCalc['isolated_fee'] as num?)?.toDouble() ?? 0.0;
    final itemFee = (deliveryFeeCalc['item_fee'] as num?)?.toDouble() ?? 0.0;
    final urgentFee =
        (deliveryFeeCalc['urgent_fee'] as num?)?.toDouble() ?? 0.0;
    final weatherFee =
        (deliveryFeeCalc['weather_fee'] as num?)?.toDouble() ?? 0.0;
    final performanceBonus =
        (deliveryFeeCalc['performance_bonus'] as num?)?.toDouble() ??
        0.0; // 绩效奖励（利润提成）
    final riderPayableFee =
        (deliveryFeeCalc['rider_payable_fee'] as num?)?.toDouble() ?? 0.0;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(Icons.local_shipping, size: 18, color: Color(0xFF20CB6B)),
              SizedBox(width: 6),
              Text(
                '配送费明细',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          _buildFeeRow('基础配送费', baseFee),
          if (isolatedFee > 0)
            _buildFeeRow('孤立订单补贴', isolatedFee, isSubsidy: true),
          if (itemFee > 0) _buildFeeRow('件数补贴', itemFee, isSubsidy: true),
          if (urgentFee > 0) _buildFeeRow('加急订单补贴', urgentFee, isSubsidy: true),
          if (weatherFee > 0) _buildFeeRow('天气补贴', weatherFee, isSubsidy: true),
          if (performanceBonus > 0)
            _buildFeeRow('平台奖励', performanceBonus, isSubsidy: true),
          const Divider(height: 24),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                '合计配送费',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              Text(
                '¥${riderPayableFee.toStringAsFixed(2)}',
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF20CB6B),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildFeeRow(String label, double amount, {bool isSubsidy = false}) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              if (isSubsidy)
                const Text(
                  '+ ',
                  style: TextStyle(
                    fontSize: 13,
                    color: Color(0xFF20CB6B),
                    fontWeight: FontWeight.w600,
                  ),
                ),
              Text(
                label,
                style: const TextStyle(fontSize: 13, color: Color(0xFF40475C)),
              ),
            ],
          ),
          Text(
            '¥${amount.toStringAsFixed(2)}',
            style: TextStyle(
              fontSize: 13,
              fontWeight: isSubsidy ? FontWeight.w600 : FontWeight.normal,
              color: isSubsidy
                  ? const Color(0xFF20CB6B)
                  : const Color(0xFF20253A),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildUrgentCard() {
    final order = _orderData?['order'] as Map<String, dynamic>?;
    final urgentFee = (order?['urgent_fee'] as num?)?.toDouble() ?? 0.0;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFFF6B6B).withOpacity(0.05),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: const Color(0xFFFF6B6B).withOpacity(0.3),
          width: 1,
        ),
      ),
      child: Row(
        children: [
          const Icon(Icons.flash_on, size: 24, color: Color(0xFFFF6B6B)),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  '加急订单',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFFFF6B6B),
                  ),
                ),
                if (urgentFee > 0) ...[
                  const SizedBox(height: 4),
                  Text(
                    '加急费用：¥${urgentFee.toStringAsFixed(2)}',
                    style: const TextStyle(
                      fontSize: 13,
                      color: Color(0xFFFF6B6B),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  bool _isPendingPickup() {
    final order = _orderData?['order'] as Map<String, dynamic>?;
    final status = order?['status'] as String? ?? '';
    return status == 'pending_pickup';
  }

  bool _isUrgent() {
    final order = _orderData?['order'] as Map<String, dynamic>?;
    return (order?['is_urgent'] as bool?) ?? false;
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'pending_delivery':
      case 'pending':
        return const Color(0xFFFF6B6B);
      case 'delivering':
        return const Color(0xFF20CB6B);
      case 'delivered':
      case 'shipped':
        return const Color(0xFFFFA726);
      case 'paid':
      case 'completed':
        return const Color(0xFF20CB6B);
      default:
        return const Color(0xFF8C92A4);
    }
  }

  String _formatStatus(String status) {
    switch (status) {
      case 'pending_delivery':
      case 'pending':
        return '待配送';
      case 'pending_pickup':
        return '待取货';
      case 'delivering':
        return '配送中';
      case 'delivered':
      case 'shipped':
        return '已送达';
      case 'paid':
      case 'completed':
        return '已收款';
      case 'cancelled':
        return '已取消';
      default:
        return status;
    }
  }

  // 获取AppBar标题（包含订单状态）
  String _getAppBarTitle() {
    if (_orderData == null) {
      return '订单详情';
    }

    final order = _orderData?['order'] as Map<String, dynamic>?;
    final status = order?['status'] as String? ?? '';

    if (status.isEmpty) {
      return '订单详情';
    }

    final statusText = _formatStatusForAppBar(status);
    return '订单详情（$statusText）';
  }

  // 格式化状态文本（用于AppBar显示）
  String _formatStatusForAppBar(String status) {
    switch (status) {
      case 'pending_delivery':
      case 'pending':
        return '待接单';
      case 'pending_pickup':
        return '待取货';
      case 'delivering':
        return '配送中';
      case 'delivered':
      case 'shipped':
        return '已送达';
      case 'paid':
      case 'completed':
        return '已收款';
      case 'cancelled':
        return '已取消';
      default:
        return status;
    }
  }

  // 构建底部操作栏
  Widget _buildActionBar() {
    final order = _orderData?['order'] as Map<String, dynamic>?;
    final status = order?['status'] as String? ?? '';

    // 待配送订单：显示接单按钮
    if (status == 'pending_delivery' || status == 'pending') {
      return Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.1),
              blurRadius: 10,
              offset: const Offset(0, -2),
            ),
          ],
        ),
        child: SafeArea(
          top: false,
          child: SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: _isProcessing ? null : _handleAcceptOrder,
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF20CB6B),
                foregroundColor: Colors.white,
                disabledBackgroundColor: const Color(0xFF9EDFB9),
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                elevation: 0,
              ),
              child: _isProcessing
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                      ),
                    )
                  : const Text(
                      '接单',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
            ),
          ),
        ),
      );
    }

    // 配送中订单：显示配送完成和问题上报按钮
    if (status == 'delivering') {
      return Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.1),
              blurRadius: 10,
              offset: const Offset(0, -2),
            ),
          ],
        ),
        child: SafeArea(
          top: false,
          child: Row(
            children: [
              // 问题上报按钮（左侧）
              Expanded(
                flex: 2,
                child: SizedBox(
                  height: 50, // 固定高度，与配送完成按钮一致
                  child: OutlinedButton(
                    onPressed: _isProcessing ? null : _handleReportIssue,
                    style: OutlinedButton.styleFrom(
                      foregroundColor: const Color(0xFF40475C),
                      side: const BorderSide(color: Color(0xFFE5E7EB)),
                      padding: EdgeInsets.zero, // 移除padding，使用固定高度
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    child: const Text(
                      '问题上报',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              // 配送完成按钮（右侧）
              Expanded(
                flex: 3,
                child: SizedBox(
                  height: 50, // 固定高度，与问题上报按钮一致
                  child: ElevatedButton(
                    onPressed: _isProcessing ? null : _handleCompleteDelivery,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF20CB6B),
                      foregroundColor: Colors.white,
                      disabledBackgroundColor: const Color(0xFF9EDFB9),
                      padding: EdgeInsets.zero, // 移除padding，使用固定高度
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      elevation: 0,
                    ),
                    child: _isProcessing
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              valueColor: AlwaysStoppedAnimation<Color>(
                                Colors.white,
                              ),
                            ),
                          )
                        : const Text(
                            '配送完成',
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                  ),
                ),
              ),
            ],
          ),
        ),
      );
    }

    // 待取货订单：显示批量取货按钮
    if (status == 'pending_pickup') {
      return Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.1),
              blurRadius: 10,
              offset: const Offset(0, -2),
            ),
          ],
        ),
        child: SafeArea(
          top: false,
          child: SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: _handleGoToBatchPickup,
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF20CB6B),
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                elevation: 0,
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: const [
                  Icon(Icons.inventory_2, size: 20),
                  SizedBox(width: 8),
                  Text(
                    '批量取货',
                    style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
                  ),
                ],
              ),
            ),
          ),
        ),
      );
    }

    // 其他状态：不显示操作按钮
    return const SizedBox.shrink();
  }

  // 处理接单
  Future<void> _handleAcceptOrder() async {
    if (_isProcessing) return;

    setState(() {
      _isProcessing = true;
    });

    final response = await OrderApi.acceptOrder(widget.orderId);

    if (!mounted) return;

    if (response.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('接单成功'),
          backgroundColor: Color(0xFF20CB6B),
        ),
      );
      // 重新加载订单详情
      await _loadOrderDetail();
      // 返回上一页
      Navigator.of(context).pop(true);
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            response.message.isNotEmpty ? response.message : '接单失败，请稍后重试',
          ),
          backgroundColor: Colors.red,
        ),
      );
    }

    if (mounted) {
      setState(() {
        _isProcessing = false;
      });
    }
  }

  // 处理配送完成 - 跳转到配送完成页面
  Future<void> _handleCompleteDelivery() async {
    if (_isProcessing) return;

    // 跳转到配送完成页面
    final result = await Navigator.of(
      context,
    ).pushNamed('/complete-delivery', arguments: {'orderId': widget.orderId});

    // 如果返回true，表示配送完成成功，刷新订单详情并返回true通知列表刷新
    if (result == true && mounted) {
      await _loadOrderDetail();
      // 返回true，通知列表页面刷新
      Navigator.of(context).pop(true);
    }
  }

  // 跳转到批量取货页面
  Future<void> _handleGoToBatchPickup() async {
    final result = await Navigator.of(context).pushNamed('/batch-pickup');
    // 如果返回值为true，表示完成了取货操作，需要刷新订单详情
    if (result == true && mounted) {
      await _loadOrderDetail();
      // 如果订单状态变为配送中，返回true通知列表刷新
      final status = _orderData?['status'] as String?;
      if (status == 'delivering') {
        Navigator.of(context).pop(true);
      }
    }
  }

  // 处理问题上报
  Future<void> _handleReportIssue() async {
    if (_isProcessing) return;

    // 显示问题上报对话框
    final result = await showDialog<Map<String, String>>(
      context: context,
      builder: (context) => _ReportIssueDialog(),
    );

    if (result == null) return;

    setState(() {
      _isProcessing = true;
    });

    final response = await OrderApi.reportOrderIssue(
      orderId: widget.orderId,
      issueType: result['issue_type'] ?? '',
      description: result['description'] ?? '',
      contactPhone: result['contact_phone'],
    );

    if (!mounted) return;

    if (response.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('问题上报成功，我们会尽快处理'),
          backgroundColor: Color(0xFF20CB6B),
        ),
      );
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            response.message.isNotEmpty ? response.message : '上报失败，请稍后重试',
          ),
          backgroundColor: Colors.red,
        ),
      );
    }

    if (mounted) {
      setState(() {
        _isProcessing = false;
      });
    }
  }
}

// 问题上报对话框
class _ReportIssueDialog extends StatefulWidget {
  @override
  State<_ReportIssueDialog> createState() => _ReportIssueDialogState();
}

class _ReportIssueDialogState extends State<_ReportIssueDialog> {
  final _formKey = GlobalKey<FormState>();
  String _selectedIssueType = '地址错误';
  final _descriptionController = TextEditingController();
  final _contactPhoneController = TextEditingController();

  final List<String> _issueTypes = ['地址错误', '无法联系客户', '商品损坏', '客户拒收', '其他问题'];

  @override
  void dispose() {
    _descriptionController.dispose();
    _contactPhoneController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('问题上报'),
      content: Form(
        key: _formKey,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // 问题类型
              const Text(
                '问题类型',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const SizedBox(height: 8),
              DropdownButtonFormField<String>(
                value: _selectedIssueType,
                decoration: InputDecoration(
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 12,
                  ),
                ),
                items: _issueTypes.map((type) {
                  return DropdownMenuItem(value: type, child: Text(type));
                }).toList(),
                onChanged: (value) {
                  if (value != null) {
                    setState(() {
                      _selectedIssueType = value;
                    });
                  }
                },
              ),
              const SizedBox(height: 16),
              // 问题描述
              const Text(
                '问题描述',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const SizedBox(height: 8),
              TextFormField(
                controller: _descriptionController,
                maxLines: 4,
                decoration: InputDecoration(
                  hintText: '请详细描述遇到的问题...',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  contentPadding: const EdgeInsets.all(12),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return '请输入问题描述';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              // 联系电话（可选）
              const Text(
                '联系电话（可选）',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const SizedBox(height: 8),
              TextFormField(
                controller: _contactPhoneController,
                keyboardType: TextInputType.phone,
                decoration: InputDecoration(
                  hintText: '请输入联系电话',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  contentPadding: const EdgeInsets.all(12),
                ),
              ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('取消'),
        ),
        ElevatedButton(
          onPressed: () {
            if (_formKey.currentState!.validate()) {
              Navigator.of(context).pop({
                'issue_type': _selectedIssueType,
                'description': _descriptionController.text.trim(),
                'contact_phone': _contactPhoneController.text.trim(),
              });
            }
          },
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFF20CB6B),
            foregroundColor: Colors.white,
          ),
          child: const Text('提交'),
        ),
      ],
    );
  }
}
