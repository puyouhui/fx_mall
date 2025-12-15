import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_map_location_marker/flutter_map_location_marker.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'package:map_launcher/map_launcher.dart';
import 'package:url_launcher/url_launcher.dart';
import 'dart:async';
import '../api/order_api.dart';
import '../utils/location_service.dart';
import '../utils/coordinate_transform.dart';
import '../widgets/accept_order_dialog.dart';
import 'route_planning_view.dart';

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

  // 路线预览相关（仅用于待接单订单）
  List<Map<String, dynamic>> _routePreviewOrders = []; // 已有订单的路线预览
  bool _isLoadingRoutePreview = false; // 是否正在加载路线预览

  // 可拖拽底部面板控制器
  final DraggableScrollableController _draggableController =
      DraggableScrollableController();
  bool _isSheetExpanded = false; // 底部面板是否展开

  // 地图相关
  final MapController _mapController = MapController();
  Position? _userPosition;
  final StreamController<LocationMarkerPosition?> _locationStreamController =
      StreamController<LocationMarkerPosition?>.broadcast();
  StreamSubscription<Position>? _positionStreamSubscription;

  // 天地图瓦片服务 URL 模板（Web墨卡托投影）
  // 使用WMTS格式，与路线规划页面保持一致
  // 影像底图（img_w）：Web墨卡托投影的影像底图
  static const String _tiandituTileUrlTemplate =
      'https://t{s}.tianditu.gov.cn/img_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=img&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=d95864378581051adb04fe26acb13ecf';

  // 天地图影像标注图层（可选，叠加在底图上）
  // 使用Web墨卡托投影的影像标注图层
  static const String _tiandituLabelUrlTemplate =
      'https://t{s}.tianditu.gov.cn/cia_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=cia&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=d95864378581051adb04fe26acb13ecf';

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
    // 立即尝试使用缓存的位置，确保位置图标能立即显示
    _tryUseCachedPosition();
    // 然后异步启动位置跟踪
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _startLocationTracking();
    });
    // 监听底部面板展开/收起状态
    _draggableController.addListener(() {
      final size = _draggableController.size;
      final isExpanded = size > 0.5; // 如果展开超过50%，认为已展开
      if (_isSheetExpanded != isExpanded) {
        setState(() {
          _isSheetExpanded = isExpanded;
        });
      }
    });
  }

  /// 尝试使用缓存的位置，确保位置图标能立即显示
  void _tryUseCachedPosition() {
    final cachedPosition = LocationService.getCachedPosition();
    if (cachedPosition != null) {
      print(
        '[OrderDetailView] 使用缓存的位置立即显示位置图标: ${cachedPosition.latitude}, ${cachedPosition.longitude}',
      );
      // 立即将缓存位置发送到流中，以便 CurrentLocationLayer 能显示
      _locationStreamController.add(
        LocationMarkerPosition(
          latitude: cachedPosition.latitude,
          longitude: cachedPosition.longitude,
          accuracy: cachedPosition.accuracy,
        ),
      );
      setState(() {
        _userPosition = cachedPosition;
      });
    } else {
      print('[OrderDetailView] 没有缓存的位置，等待异步获取');
    }
  }

  @override
  void dispose() {
    _positionStreamSubscription?.cancel();
    _locationStreamController.close();
    _mapController.dispose();
    _draggableController.dispose();
    super.dispose();
  }

  Future<void> _startLocationTracking() async {
    final hasPermission = await LocationService.checkAndRequestPermission();
    if (!hasPermission) {
      print('[OrderDetailView] 没有定位权限，无法启动位置跟踪');
      // 即使没有权限，也尝试使用缓存位置（如果有的话）
      if (_userPosition == null) {
        _tryUseCachedPosition();
      }
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
          // 即使网络定位失败，也尝试使用缓存位置（如果有的话）
          if (_userPosition == null) {
            _tryUseCachedPosition();
          }
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
      } else {
        // 如果获取位置失败，也尝试使用缓存位置（如果有的话）
        if (_userPosition == null) {
          _tryUseCachedPosition();
        }
      }
    } catch (e) {
      print('[OrderDetailView] 启动位置跟踪失败: $e');
      // 错误处理：尝试使用低精度重新启动
      if (mounted) {
        // 如果还没有位置，先尝试使用缓存位置
        if (_userPosition == null) {
          _tryUseCachedPosition();
        }
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

              // 如果有客户地址，调整地图视野以同时显示所有位置
              final addressData =
                  _orderData?['address'] as Map<String, dynamic>?;
              final customerLat = addressData?['latitude'] as num?;
              final customerLng = addressData?['longitude'] as num?;

              // 判断订单状态
              final order = _orderData?['order'] as Map<String, dynamic>?;
              final status = order?['status'] as String? ?? '';
              final isPendingDelivery =
                  status == 'pending_delivery' || status == 'pending';
              final isPendingPickup = status == 'pending_pickup';
              final showRoutePreview = isPendingDelivery || isPendingPickup;

              if (customerLat != null &&
                  customerLng != null &&
                  _userPosition != null) {
                // 将GCJ-02坐标转换为WGS84坐标（天地图使用WGS84）
                final customerWgs84 = CoordinateTransform.gcj02ToWgs84(
                  customerLat.toDouble(),
                  customerLng.toDouble(),
                );

                // 如果是待接单或待取货状态且有路线预览，调整地图视野以显示所有点
                if (showRoutePreview && _routePreviewOrders.isNotEmpty) {
                  // 获取供应商位置（待取货状态时）
                  final suppliers = isPendingPickup
                      ? (_orderData?['suppliers'] as List<dynamic>?)
                                ?.cast<Map<String, dynamic>>() ??
                            []
                      : [];

                  final allPoints = <LatLng>[
                    LatLng(_userPosition!.latitude, _userPosition!.longitude),
                    customerWgs84,
                    ..._routePreviewOrders
                        .where((order) {
                          final lat = order['latitude'] as num?;
                          final lng = order['longitude'] as num?;
                          return lat != null && lng != null;
                        })
                        .map((order) {
                          final lat = order['latitude'] as num?;
                          final lng = order['longitude'] as num?;
                          return CoordinateTransform.gcj02ToWgs84(
                            lat!.toDouble(),
                            lng!.toDouble(),
                          );
                        }),
                    // 添加供应商位置（待取货状态时）
                    ...suppliers
                        .where((supplier) {
                          final lat = supplier['latitude'] as num?;
                          final lng = supplier['longitude'] as num?;
                          return lat != null && lng != null;
                        })
                        .map((supplier) {
                          final lat = supplier['latitude'] as num?;
                          final lng = supplier['longitude'] as num?;
                          return CoordinateTransform.gcj02ToWgs84(
                            lat!.toDouble(),
                            lng!.toDouble(),
                          );
                        }),
                  ];
                  _adjustMapBoundsForMultiplePoints(allPoints);
                } else {
                  // 只有两个点时，使用原来的方法
                  _adjustMapBounds(
                    customerWgs84,
                    LatLng(_userPosition!.latitude, _userPosition!.longitude),
                  );
                }
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

  /// 调整地图视野以显示多个点（用于路线预览）
  void _adjustMapBoundsForMultiplePoints(List<LatLng> points) {
    if (points.isEmpty) return;

    // 计算所有点的边界
    double minLat = points.first.latitude;
    double maxLat = points.first.latitude;
    double minLng = points.first.longitude;
    double maxLng = points.first.longitude;

    for (final point in points) {
      if (point.latitude < minLat) minLat = point.latitude;
      if (point.latitude > maxLat) maxLat = point.latitude;
      if (point.longitude < minLng) minLng = point.longitude;
      if (point.longitude > maxLng) maxLng = point.longitude;
    }

    // 计算中心点
    final centerLat = (minLat + maxLat) / 2;
    final centerLng = (minLng + maxLng) / 2;
    final center = LatLng(centerLat, centerLng);

    // 计算最大距离以确定合适的缩放级别
    double maxDistance = 0;
    for (int i = 0; i < points.length; i++) {
      for (int j = i + 1; j < points.length; j++) {
        final distance = const Distance().as(
          LengthUnit.Meter,
          points[i],
          points[j],
        );
        if (distance > maxDistance) {
          maxDistance = distance;
        }
      }
    }

    double zoom = 15.0;
    if (maxDistance > 10000) {
      zoom = 11.0;
    } else if (maxDistance > 5000) {
      zoom = 12.0;
    } else if (maxDistance > 2000) {
      zoom = 13.0;
    } else if (maxDistance > 1000) {
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

      // 调试：打印完整订单数据和销售员信息
      print('[OrderDetailView] 完整订单数据 keys: ${_orderData?.keys.toList()}');
      final salesEmployee =
          _orderData?['sales_employee'] as Map<String, dynamic>?;
      print('[OrderDetailView] 销售员信息: $salesEmployee');
      if (salesEmployee != null) {
        print('[OrderDetailView] 销售员姓名: ${salesEmployee['name']}');
        print('[OrderDetailView] 销售员电话: ${salesEmployee['phone']}');
        print('[OrderDetailView] 销售员工号: ${salesEmployee['employee_code']}');
      } else {
        print('[OrderDetailView] 销售员信息为null');
      }

      // 如果是待接单、待取货或配送中状态，加载路线预览
      final order = _orderData?['order'] as Map<String, dynamic>?;
      final status = order?['status'] as String? ?? '';
      if (status == 'pending_delivery' ||
          status == 'pending' ||
          status == 'pending_pickup' ||
          status == 'delivering') {
        _loadRoutePreview();
      }
    } else {
      setState(() {
        _errorMessage = response.message.isNotEmpty
            ? response.message
            : '获取订单详情失败';
        _isLoading = false;
      });
    }
  }

  /// 静默加载订单详情（不显示加载状态，用于接单后刷新）
  Future<void> _loadOrderDetailSilently() async {
    final response = await OrderApi.getOrderDetail(widget.orderId);

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      setState(() {
        _orderData = response.data;
      });

      // 如果是待接单、待取货或配送中状态，加载路线预览
      final order = _orderData?['order'] as Map<String, dynamic>?;
      final status = order?['status'] as String? ?? '';
      if (status == 'pending_delivery' ||
          status == 'pending' ||
          status == 'pending_pickup' ||
          status == 'delivering') {
        // 清除旧的路线预览数据，重新加载
        setState(() {
          _routePreviewOrders = [];
        });
        _loadRoutePreview();
      } else {
        // 如果状态不是这些，清除路线预览
        setState(() {
          _routePreviewOrders = [];
        });
      }
    }
  }

  /// 加载路线预览（用于待接单和待取货订单）
  Future<void> _loadRoutePreview() async {
    if (_isLoadingRoutePreview) return;

    setState(() {
      _isLoadingRoutePreview = true;
    });

    try {
      final response = await OrderApi.getRouteOrders();

      if (!mounted) return;

      if (response.isSuccess && response.data != null) {
        final data = response.data!;
        final List<dynamic> list = (data['list'] as List<dynamic>? ?? []);
        final orders = list.cast<Map<String, dynamic>>();

        // 获取当前订单ID和状态
        final currentOrderId = widget.orderId;
        final order = _orderData?['order'] as Map<String, dynamic>?;
        final currentStatus = order?['status'] as String? ?? '';
        final isDelivering = currentStatus == 'delivering';

        // 如果是配送中状态，包含所有订单（包括当前订单），以显示完整路线
        // 如果是待接单或待取货状态，过滤掉当前订单（因为当前订单会单独显示）
        final filteredOrders = orders.where((order) {
          final orderId = order['id'] as int?;
          final status = order['status'] as String? ?? '';

          // 配送中状态：包含所有订单（包括当前订单和已完成的订单）
          if (isDelivering) {
            return status == 'delivering' ||
                status == 'delivered' ||
                status == 'shipped';
          }

          // 待接单或待取货状态：过滤掉当前订单
          if (orderId == null || orderId == currentOrderId) {
            return false;
          }
          // 只显示未完成的订单
          return status == 'delivering' || status == 'pending_pickup';
        }).toList();

        setState(() {
          _routePreviewOrders = filteredOrders;
          _isLoadingRoutePreview = false;
        });

        // 路线预览加载完成后，调整地图视野以显示所有点
        if (mounted && orders.isNotEmpty && _userPosition != null) {
          final addressData = _orderData?['address'] as Map<String, dynamic>?;
          final customerLat = addressData?['latitude'] as num?;
          final customerLng = addressData?['longitude'] as num?;

          final order = _orderData?['order'] as Map<String, dynamic>?;
          final status = order?['status'] as String? ?? '';
          final isPendingPickup = status == 'pending_pickup';

          // 获取供应商位置（待取货状态时）
          final suppliers = isPendingPickup
              ? (_orderData?['suppliers'] as List<dynamic>?)
                        ?.cast<Map<String, dynamic>>() ??
                    []
              : [];

          if (customerLat != null && customerLng != null) {
            final customerWgs84 = CoordinateTransform.gcj02ToWgs84(
              customerLat.toDouble(),
              customerLng.toDouble(),
            );
            final allPoints = <LatLng>[
              LatLng(_userPosition!.latitude, _userPosition!.longitude),
              customerWgs84,
              ...orders
                  .where((order) {
                    final lat = order['latitude'] as num?;
                    final lng = order['longitude'] as num?;
                    return lat != null && lng != null;
                  })
                  .map((order) {
                    final lat = order['latitude'] as num?;
                    final lng = order['longitude'] as num?;
                    return CoordinateTransform.gcj02ToWgs84(
                      lat!.toDouble(),
                      lng!.toDouble(),
                    );
                  }),
              // 添加供应商位置（待取货状态时）
              ...suppliers
                  .where((supplier) {
                    final lat = supplier['latitude'] as num?;
                    final lng = supplier['longitude'] as num?;
                    return lat != null && lng != null;
                  })
                  .map((supplier) {
                    final lat = supplier['latitude'] as num?;
                    final lng = supplier['longitude'] as num?;
                    return CoordinateTransform.gcj02ToWgs84(
                      lat!.toDouble(),
                      lng!.toDouble(),
                    );
                  }),
            ];
            _adjustMapBoundsForMultiplePoints(allPoints);
          }
        }
      } else {
        setState(() {
          _routePreviewOrders = [];
          _isLoadingRoutePreview = false;
        });
      }
    } catch (e) {
      print('[OrderDetailView] 加载路线预览失败: $e');
      if (mounted) {
        setState(() {
          _routePreviewOrders = [];
          _isLoadingRoutePreview = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    // 判断是否是待取货或配送中状态
    final order = _orderData?['order'] as Map<String, dynamic>?;
    final status = order?['status'] as String? ?? '';
    final isPendingPickup = status == 'pending_pickup';
    final isDelivering = status == 'delivering';

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
        actions: [
          // 待取货或配送中状态时显示规划按钮
          if (isPendingPickup || isDelivering)
            IconButton(
              icon: const Icon(Icons.route, color: Colors.white),
              onPressed: _navigateToRoutePlanning,
              tooltip: '路线规划',
            ),
        ],
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
                // 全屏地图
                _buildFullScreenMap(),
                // 底部可拖拽面板（订单信息）
                DraggableScrollableSheet(
                  controller: _draggableController,
                  initialChildSize: 0.4, // 初始高度为屏幕的40%
                  minChildSize: 0.4, // 最小高度为屏幕的40%
                  maxChildSize: 0.85, // 最大高度为屏幕的85%
                  snap: true, // 启用吸附效果
                  snapSizes: const [0.4, 0.85], // 吸附位置：默认/收起、完全展开
                  builder: (context, scrollController) {
                    return Container(
                      margin: const EdgeInsets.symmetric(horizontal: 0),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: const BorderRadius.only(
                          topLeft: Radius.circular(20),
                          topRight: Radius.circular(20),
                        ),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withOpacity(0.12),
                            blurRadius: 16,
                            offset: const Offset(0, -4),
                            spreadRadius: 0,
                          ),
                        ],
                      ),
                      child: Column(
                        children: [
                          // 拖拽指示器
                          Container(
                            margin: const EdgeInsets.only(top: 8, bottom: 4),
                            width: 40,
                            height: 4,
                            decoration: BoxDecoration(
                              color: Colors.grey[300],
                              borderRadius: BorderRadius.circular(2),
                            ),
                          ),
                          // 可滚动内容区域
                          Expanded(
                            child: _buildOrderInfoContent(scrollController),
                          ),
                          // 固定的操作按钮区域
                          _buildActionBar(),
                        ],
                      ),
                    );
                  },
                ),
              ],
            ),
    );
  }

  /// 构建全屏地图
  Widget _buildFullScreenMap() {
    final addressData = _orderData?['address'] as Map<String, dynamic>?;
    final customerLat = addressData?['latitude'] as num?;
    final customerLng = addressData?['longitude'] as num?;

    // 判断订单状态
    final order = _orderData?['order'] as Map<String, dynamic>?;
    final status = order?['status'] as String? ?? '';
    final isPendingDelivery =
        status == 'pending_delivery' || status == 'pending';
    final isPendingPickup = status == 'pending_pickup';
    final isDelivering = status == 'delivering';
    // 待取货状态不显示路线预览，只显示位置点
    final showRoutePreview = isPendingDelivery || isDelivering;

    // 获取供应商列表（仅在待取货状态时）
    final suppliers = _isPendingPickup()
        ? (_orderData?['suppliers'] as List<dynamic>?)
                  ?.cast<Map<String, dynamic>>() ??
              []
        : [];

    // 确定地图初始中心点（使用WGS84坐标）
    LatLng initialCenter = const LatLng(39.90864, 116.39750); // 默认北京
    if (customerLat != null && customerLng != null) {
      // 将GCJ-02坐标转换为WGS84坐标（天地图使用WGS84）
      final wgs84Point = CoordinateTransform.gcj02ToWgs84(
        customerLat.toDouble(),
        customerLng.toDouble(),
      );
      initialCenter = wgs84Point;
    } else if (_userPosition != null) {
      initialCenter = LatLng(_userPosition!.latitude, _userPosition!.longitude);
    }

    return SizedBox.expand(
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
              // 路线预览：已有订单的路线（待接单或配送中状态时显示，待取货状态不显示路线）
              if (showRoutePreview &&
                  _routePreviewOrders.isNotEmpty &&
                  _userPosition != null)
                PolylineLayer(
                  polylines: [
                    Polyline(
                      points: [
                        // 从配送员位置开始
                        LatLng(
                          _userPosition!.latitude,
                          _userPosition!.longitude,
                        ),
                        // 添加已有订单的位置点（配送中状态时按 route_sequence 排序）
                        // 注意：只连接未完成的订单，已送达订单只显示marker，不参与路线计算
                        ...(() {
                          if (isDelivering) {
                            // 过滤掉已送达的订单，只保留未完成的订单
                            final sortedOrders = _routePreviewOrders.where((
                              order,
                            ) {
                              final lat = order['latitude'] as num?;
                              final lng = order['longitude'] as num?;
                              if (lat == null || lng == null) return false;
                              // 只包含未完成的订单
                              final status = order['status'] as String? ?? '';
                              return status != 'delivered' &&
                                  status != 'shipped';
                            }).toList();
                            sortedOrders.sort((a, b) {
                              final seqA =
                                  (a['route_sequence'] as num?)?.toInt() ?? 0;
                              final seqB =
                                  (b['route_sequence'] as num?)?.toInt() ?? 0;
                              return seqA.compareTo(seqB);
                            });
                            return sortedOrders;
                          } else {
                            // 待接单状态：只包含未完成的订单
                            return _routePreviewOrders.where((order) {
                              final lat = order['latitude'] as num?;
                              final lng = order['longitude'] as num?;
                              if (lat == null || lng == null) return false;
                              // 只包含未完成的订单
                              final status = order['status'] as String? ?? '';
                              return status != 'delivered' &&
                                  status != 'shipped';
                            }).toList();
                          }
                        }()).map((order) {
                          final lat = order['latitude'] as num?;
                          final lng = order['longitude'] as num?;
                          // 将GCJ-02坐标转换为WGS84坐标
                          return CoordinateTransform.gcj02ToWgs84(
                            lat!.toDouble(),
                            lng!.toDouble(),
                          );
                        }),
                      ],
                      strokeWidth: 4,
                      color: const Color(
                        0xFF20CB6B,
                      ).withOpacity(0.7), // 增加透明度，使路线更明显
                    ),
                  ],
                ),
              // 路线预览：已有订单的标记点（待接单或配送中状态时显示）
              // 配送中状态：显示所有订单的marker（包括当前订单），当前订单使用蓝色覆盖在对应位置
              if (showRoutePreview && _routePreviewOrders.isNotEmpty)
                MarkerLayer(
                  markers:
                      () {
                        // 配送中状态：按 route_sequence 排序，当前订单放在最后以确保覆盖显示
                        if (isDelivering) {
                          final sortedOrders = _routePreviewOrders.where((
                            order,
                          ) {
                            final lat = order['latitude'] as num?;
                            final lng = order['longitude'] as num?;
                            return lat != null && lng != null;
                          }).toList();
                          sortedOrders.sort((a, b) {
                            final orderIdA = a['id'] as int?;
                            final orderIdB = b['id'] as int?;
                            final isCurrentA = orderIdA == widget.orderId;
                            final isCurrentB = orderIdB == widget.orderId;

                            // 当前订单放在最后
                            if (isCurrentA) return 1;
                            if (isCurrentB) return -1;

                            // 其他订单按 route_sequence 排序
                            final seqA =
                                (a['route_sequence'] as num?)?.toInt() ?? 0;
                            final seqB =
                                (b['route_sequence'] as num?)?.toInt() ?? 0;
                            return seqA.compareTo(seqB);
                          });
                          return sortedOrders;
                        } else {
                          // 待接单状态：不过滤，不排序
                          return _routePreviewOrders.where((order) {
                            final lat = order['latitude'] as num?;
                            final lng = order['longitude'] as num?;
                            return lat != null && lng != null;
                          }).toList();
                        }
                      }().asMap().entries.map((entry) {
                        final index = entry.key;
                        final order = entry.value;
                        final orderId = order['id'] as int?;
                        final lat = order['latitude'] as num?;
                        final lng = order['longitude'] as num?;
                        final sequence = order['route_sequence'] as num?;
                        final status = order['status'] as String? ?? '';
                        final isCurrentOrder = orderId == widget.orderId;
                        final isDeliveredOrder =
                            status == 'delivered' || status == 'shipped';

                        // 将GCJ-02坐标转换为WGS84坐标
                        final wgs84Point = CoordinateTransform.gcj02ToWgs84(
                          lat!.toDouble(),
                          lng!.toDouble(),
                        );

                        // 配送中状态：当前订单使用蓝色，其他订单使用绿色，已送达订单使用灰色
                        final markerColor = isDelivering
                            ? (isCurrentOrder
                                  ? const Color(0xFF2196F3) // 蓝色表示当前订单
                                  : isDeliveredOrder
                                  ? (Colors.grey[600] ?? Colors.grey) // 灰色表示已送达
                                  : const Color(0xFF20CB6B)) // 绿色表示其他配送中订单
                            : const Color(0xFF20CB6B); // 待接单状态使用绿色

                        return Marker(
                          point: wgs84Point,
                          width: 28,
                          height: 28,
                          alignment: Alignment.center,
                          child: Stack(
                            alignment: Alignment.center,
                            children: [
                              // 圆形背景
                              Container(
                                width: 28,
                                height: 28,
                                decoration: BoxDecoration(
                                  color: markerColor,
                                  shape: BoxShape.circle,
                                  border: Border.all(
                                    color: Colors.white,
                                    width: isDelivering && isCurrentOrder
                                        ? 2.5
                                        : 2,
                                  ),
                                  boxShadow: [
                                    BoxShadow(
                                      color: Colors.black.withOpacity(0.3),
                                      blurRadius: isDelivering && isCurrentOrder
                                          ? 5
                                          : 3,
                                      offset: const Offset(0, 2),
                                    ),
                                  ],
                                ),
                              ),
                              // 序号文本（显示在内部）
                              Text(
                                isDelivering && isCurrentOrder
                                    ? '送' // 当前订单显示"送"
                                    : '${sequence ?? (index + 1)}',
                                style: TextStyle(
                                  fontSize: isDelivering && isCurrentOrder
                                      ? 12
                                      : 11,
                                  fontWeight: FontWeight.w700,
                                  color: Colors.white,
                                ),
                              ),
                            ],
                          ),
                        );
                      }).toList(),
                ),
              // 当前订单位置标记（待接单、待取货状态时显示，配送中状态不显示，因为已经在路线marker中显示）
              if (customerLat != null && customerLng != null && !isDelivering)
                MarkerLayer(
                  markers: [
                    Marker(
                      point: CoordinateTransform.gcj02ToWgs84(
                        customerLat.toDouble(),
                        customerLng.toDouble(),
                      ),
                      width: 75,
                      height: 60,
                      alignment: Alignment.topCenter,
                      child: // 其他状态：显示外部标签和圆形点
                      Column(
                        mainAxisSize: MainAxisSize.min,
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: [
                          // 外部标签（当前订单特有）
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 6,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: isPendingDelivery
                                  ? const Color(0xFFFF9800) // 橙色表示待接单
                                  : isPendingPickup
                                  ? const Color(0xFF20CB6B) // 待取货状态使用绿色
                                  : const Color(0xFFFF6B6B), // 红色表示其他状态
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
                              isPendingDelivery
                                  ? '新订单'
                                  : isPendingPickup
                                  ? '客户位置' // 待取货状态显示"客户位置"
                                  : '客户位置',
                              style: const TextStyle(
                                fontSize: 11, // 与供应商marker保持一致
                                fontWeight: FontWeight.w600,
                                color: Colors.white,
                              ),
                              maxLines: 1, // 确保文本在一行显示
                              overflow: TextOverflow.ellipsis, // 超出部分显示省略号
                            ),
                          ),
                          const SizedBox(height: 2),
                          // 圆形标记点（待取货状态不显示内部文字）
                          Stack(
                            alignment: Alignment.center,
                            children: [
                              Container(
                                width: 28,
                                height: 28,
                                decoration: BoxDecoration(
                                  color: isPendingDelivery
                                      ? const Color(0xFFFF9800) // 橙色表示待接单
                                      : isPendingPickup
                                      ? const Color(0xFF20CB6B) // 待取货状态使用绿色
                                      : const Color(0xFFFF6B6B), // 红色表示其他状态
                                  shape: BoxShape.circle,
                                  border: Border.all(
                                    color: Colors.white,
                                    width: 2,
                                  ),
                                  boxShadow: [
                                    BoxShadow(
                                      color: Colors.black.withOpacity(0.3),
                                      blurRadius: 3,
                                      offset: const Offset(0, 2),
                                    ),
                                  ],
                                ),
                              ),
                              // 文本标签（待取货状态不显示内部文字）
                              if (!isPendingPickup)
                                Text(
                                  isPendingDelivery ? '新' : '客',
                                  style: const TextStyle(
                                    fontSize: 11,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.white,
                                  ),
                                ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              // 供应商位置标记（待取货状态时显示，与路线预览一起显示）
              if (isPendingPickup && suppliers.isNotEmpty)
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
                        // 将GCJ-02坐标转换为WGS84坐标（天地图使用WGS84）
                        final wgs84Point = CoordinateTransform.gcj02ToWgs84(
                          lat!.toDouble(),
                          lng!.toDouble(),
                        );
                        return Marker(
                          point: wgs84Point,
                          width: 110,
                          height: 60,
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
                                  width: 28,
                                  height: 28,
                                  decoration: BoxDecoration(
                                    color: const Color(0xFFFF5722),
                                    shape: BoxShape.circle,
                                    border: Border.all(
                                      color: Colors.white,
                                      width: 2,
                                    ),
                                    boxShadow: [
                                      BoxShadow(
                                        color: Colors.black.withOpacity(0.3),
                                        blurRadius: 3,
                                        offset: const Offset(0, 2),
                                      ),
                                    ],
                                  ),
                                  child: const Icon(
                                    Icons.warehouse,
                                    color: Colors.white,
                                    size: 16,
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
                  Icon(
                    isPendingDelivery
                        ? Icons.add_location_alt
                        : isPendingPickup
                        ? Icons.location_searching
                        : isDelivering
                        ? Icons.local_shipping
                        : Icons.location_on,
                    size: 16,
                    color: isPendingDelivery
                        ? const Color(0xFFFF9800)
                        : isPendingPickup
                        ? const Color(0xFF2196F3)
                        : isDelivering
                        ? const Color(0xFF20CB6B)
                        : const Color(0xFFFF6B6B),
                  ),
                  const SizedBox(width: 4),
                  Text(
                    isPendingDelivery
                        ? '新订单'
                        : isPendingPickup
                        ? '当前订单'
                        : isDelivering
                        ? '当前订单'
                        : '客户',
                    style: const TextStyle(
                      fontSize: 12,
                      color: Color(0xFF20253A),
                    ),
                  ),
                  // 待接单或待取货状态时显示已有订单图例
                  if (showRoutePreview && _routePreviewOrders.isNotEmpty) ...[
                    const SizedBox(width: 12),
                    Container(
                      width: 16,
                      height: 16,
                      decoration: const BoxDecoration(
                        color: Color(0xFF20CB6B),
                        shape: BoxShape.circle,
                      ),
                    ),
                    const SizedBox(width: 4),
                    const Text(
                      '已有订单',
                      style: TextStyle(fontSize: 12, color: Color(0xFF20253A)),
                    ),
                  ],
                  // 待取货状态时显示供应商图例
                  if (isPendingPickup && suppliers.isNotEmpty) ...[
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
                      style: TextStyle(fontSize: 12, color: Color(0xFF20253A)),
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
    );
  }

  /// 构建订单信息内容（用于DraggableScrollableSheet）
  Widget _buildOrderInfoContent(ScrollController scrollController) {
    return ListView(
      controller: scrollController,
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
      physics:
          const ClampingScrollPhysics(), // 使用ClampingScrollPhysics以配合DraggableScrollableSheet
      children: [
        // 待取货状态时，供应商列表优先显示
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
        if (_isUrgent()) const SizedBox(height: 12),
        // 订单基本信息
        _buildOrderInfoCard(),
        const SizedBox(height: 12),
        // 订单选项（备注、缺货处理、信任签收等）
        _buildOrderOptionsCard(),
        const SizedBox(height: 12),
        // 销售员信息
        _buildSalesEmployeeCard(),
        const SizedBox(height: 16), // 底部留出空间给固定按钮
      ],
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
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 6,
            offset: const Offset(0, 2),
            spreadRadius: 0,
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

    // 导航到客户位置
    Future<void> _navigateToCustomer() async {
      if (customerLat == null || customerLng == null) return;

      try {
        // 检查是否安装了高德地图
        final isAmapAvailable = await MapLauncher.isMapAvailable(MapType.amap);
        if (isAmapAvailable == true) {
          // 使用高德地图导航
          await MapLauncher.showDirections(
            mapType: MapType.amap,
            destination: Coords(customerLat.toDouble(), customerLng.toDouble()),
            destinationTitle: name.isNotEmpty ? name : '客户位置',
          );
        } else {
          // 如果没有高德地图，检查其他可用的地图应用
          final availableMaps = await MapLauncher.installedMaps;
          if (availableMaps.isNotEmpty) {
            // 使用第一个可用的地图应用
            await availableMaps.first.showDirections(
              destination: Coords(
                customerLat.toDouble(),
                customerLng.toDouble(),
              ),
              destinationTitle: name.isNotEmpty ? name : '客户位置',
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

    // 拨打电话
    Future<void> _callCustomer() async {
      if (phone.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('客户未提供联系电话'),
              duration: Duration(seconds: 2),
            ),
          );
        }
        return;
      }

      try {
        // 使用原生平台通道直接调用 Android Intent
        const platform = MethodChannel('com.example.distribution_app/phone');
        await platform.invokeMethod('dialPhone', {'phone': phone});
      } catch (e) {
        // 如果原生方法失败，尝试使用 url_launcher
        try {
          final uri = Uri.parse('tel:$phone');
          await launchUrl(uri, mode: LaunchMode.externalApplication);
        } catch (e2) {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text('拨打电话失败，请手动拨打: $phone'),
                duration: const Duration(seconds: 2),
              ),
            );
          }
        }
      }
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 6,
            offset: const Offset(0, 2),
            spreadRadius: 0,
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 地址名称（标题）
          Row(
            children: [
              const Icon(Icons.location_on, size: 18, color: Color(0xFF20CB6B)),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  name.isNotEmpty ? name : '收货地址',
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20253A),
                  ),
                ),
              ),
              // 导航按钮
              if (customerLat != null && customerLng != null)
                InkWell(
                  onTap: _navigateToCustomer,
                  borderRadius: BorderRadius.circular(8),
                  child: Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: const Color(0xFF20CB6B).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Icon(
                      Icons.navigation,
                      size: 18,
                      color: Color(0xFF20CB6B),
                    ),
                  ),
                ),
              // 拨打电话按钮
              if (phone.isNotEmpty) ...[
                const SizedBox(width: 8),
                InkWell(
                  onTap: _callCustomer,
                  borderRadius: BorderRadius.circular(8),
                  child: Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: const Color(0xFF20CB6B).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Icon(
                      Icons.phone,
                      size: 18,
                      color: Color(0xFF20CB6B),
                    ),
                  ),
                ),
              ],
            ],
          ),
          // 地址详情
          if (address.isNotEmpty) ...[
            const SizedBox(height: 12),
            Text(
              address,
              style: const TextStyle(
                fontSize: 14,
                color: Color(0xFF20253A),
                height: 1.5,
              ),
            ),
          ],
          // 联系人和电话
          if (contact.isNotEmpty || phone.isNotEmpty) ...[
            const SizedBox(height: 12),
            Row(
              children: [
                const Icon(
                  Icons.person_outline,
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
        ],
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
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 6,
            offset: const Offset(0, 2),
            spreadRadius: 0,
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
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 6,
            offset: const Offset(0, 2),
            spreadRadius: 0,
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
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.03),
              blurRadius: 6,
              offset: const Offset(0, 2),
              spreadRadius: 0,
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
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 6,
            offset: const Offset(0, 2),
            spreadRadius: 0,
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
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: const Color(0xFFFF6B6B).withOpacity(0.25),
          width: 1.5,
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

  /// 构建销售员信息卡片
  Widget _buildSalesEmployeeCard() {
    final salesEmployee =
        _orderData?['sales_employee'] as Map<String, dynamic>?;

    // 调试：打印销售员信息
    print(
      '[OrderDetailView] _buildSalesEmployeeCard - salesEmployee: $salesEmployee',
    );
    print(
      '[OrderDetailView] _buildSalesEmployeeCard - _orderData keys: ${_orderData?.keys.toList()}',
    );

    // 如果销售员信息不存在，或者没有姓名和电话，则不显示
    if (salesEmployee == null) {
      print(
        '[OrderDetailView] _buildSalesEmployeeCard - salesEmployee is null, returning empty',
      );
      return const SizedBox.shrink();
    }

    final name = salesEmployee['name'] as String? ?? '';
    final phone = salesEmployee['phone'] as String? ?? '';
    final employeeCode = salesEmployee['employee_code'] as String? ?? '';

    print(
      '[OrderDetailView] _buildSalesEmployeeCard - name: $name, phone: $phone, employeeCode: $employeeCode',
    );

    // 如果既没有姓名也没有电话，则不显示
    if (name.isEmpty && phone.isEmpty) {
      print(
        '[OrderDetailView] _buildSalesEmployeeCard - name and phone are empty, returning empty',
      );
      return const SizedBox.shrink();
    }

    // 拨打电话
    Future<void> _callSalesEmployee() async {
      if (phone.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('销售员未提供联系电话'),
              duration: Duration(seconds: 2),
            ),
          );
        }
        return;
      }

      try {
        // 使用原生平台通道直接调用 Android Intent
        const platform = MethodChannel('com.example.distribution_app/phone');
        await platform.invokeMethod('dialPhone', {'phone': phone});
      } catch (e) {
        // 如果原生方法失败，尝试使用 url_launcher
        try {
          final uri = Uri.parse('tel:$phone');
          await launchUrl(uri, mode: LaunchMode.externalApplication);
        } catch (e2) {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text('拨打电话失败，请手动拨打: $phone'),
                duration: const Duration(seconds: 2),
              ),
            );
          }
        }
      }
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 6,
            offset: const Offset(0, 2),
            spreadRadius: 0,
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(
                Icons.person_outline,
                size: 18,
                color: Color(0xFF20CB6B),
              ),
              const SizedBox(width: 6),
              const Text(
                '销售员',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              // 拨打电话按钮
              if (phone.isNotEmpty)
                InkWell(
                  onTap: _callSalesEmployee,
                  borderRadius: BorderRadius.circular(8),
                  child: Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: const Color(0xFF20CB6B).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Icon(
                      Icons.phone,
                      size: 18,
                      color: Color(0xFF20CB6B),
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: 12),
          if (name.isNotEmpty) ...[
            Row(
              children: [
                const Text(
                  '姓名：',
                  style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
                ),
                Text(
                  name,
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                    color: Color(0xFF20253A),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
          ],
          if (employeeCode.isNotEmpty) ...[
            Row(
              children: [
                const Text(
                  '工号：',
                  style: TextStyle(fontSize: 14, color: Color(0xFF8C92A4)),
                ),
                Text(
                  employeeCode,
                  style: const TextStyle(
                    fontSize: 14,
                    color: Color(0xFF20253A),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
          ],
          if (phone.isNotEmpty) ...[
            Row(
              children: [
                const Icon(
                  Icons.phone_outlined,
                  size: 14,
                  color: Color(0xFF8C92A4),
                ),
                const SizedBox(width: 6),
                Text(
                  phone,
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

  /// 构建订单选项卡片（显示备注、缺货处理、信任签收等）
  Widget _buildOrderOptionsCard() {
    final order = _orderData?['order'] as Map<String, dynamic>?;
    final remark = (order?['remark'] as String?) ?? '';
    final outOfStockStrategy =
        (order?['out_of_stock_strategy'] as String?) ?? '';
    final trustReceipt = (order?['trust_receipt'] as bool?) ?? false;
    final hidePrice = (order?['hide_price'] as bool?) ?? false;
    final requirePhoneContact =
        (order?['require_phone_contact'] as bool?) ?? true;
    final isUrgent = (order?['is_urgent'] as bool?) ?? false;

    // 缺货处理策略文本
    String outOfStockText = '';
    switch (outOfStockStrategy) {
      case 'cancel_item':
        outOfStockText = '缺货商品不要，其他正常发货';
        break;
      case 'ship_available':
        outOfStockText = '有货就发，缺货商品不发';
        break;
      case 'contact_me':
        outOfStockText = '由客服或配送员联系我确认';
        break;
      default:
        if (outOfStockStrategy.isNotEmpty) {
          outOfStockText = outOfStockStrategy;
        }
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 6,
            offset: const Offset(0, 2),
            spreadRadius: 0,
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '订单选项',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Color(0xFF20253A),
            ),
          ),
          const SizedBox(height: 12),
          // 加急订单 - 放在最前面，红色突出
          if (isUrgent) ...[
            _buildOptionRow(
              label: '加急订单',
              value: '优先为客户配送',
              icon: Icons.flash_on_outlined,
              highlightColor: const Color(0xFFFF5A5F), // 红色
            ),
            const SizedBox(height: 12),
          ],
          // 订单备注 - 橙色突出
          if (remark.isNotEmpty) ...[
            _buildOptionRow(
              label: '订单备注',
              value: remark,
              icon: Icons.note_outlined,
              highlightColor: const Color(0xFFFFA940), // 橙色
            ),
            const SizedBox(height: 12),
          ],
          // 缺货处理策略
          if (outOfStockText.isNotEmpty) ...[
            _buildOptionRow(
              label: '遇到缺货时',
              value: outOfStockText,
              icon: Icons.warning_amber_outlined,
            ),
            const SizedBox(height: 12),
          ],
          // 其他选项
          _buildOptionSwitch(
            label: '信任签收',
            value: trustReceipt,
            description: '配送电话联系不上时，允许放门口或指定位置',
            icon: Icons.verified_user_outlined,
          ),
          const SizedBox(height: 12),
          _buildOptionSwitch(
            label: '隐藏价格',
            value: hidePrice,
            description: '小票中将不显示商品价格',
            icon: Icons.visibility_off_outlined,
          ),
          const SizedBox(height: 12),
          _buildOptionSwitch(
            label: '配送时电话联系',
            value: requirePhoneContact,
            description: '建议保持电话畅通，方便配送员联系',
            icon: Icons.phone_outlined,
          ),
        ],
      ),
    );
  }

  /// 构建选项行（文本类型）
  Widget _buildOptionRow({
    required String label,
    required String value,
    required IconData icon,
    Color? highlightColor,
  }) {
    final color = highlightColor ?? const Color(0xFF4C8DF6);

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.all(6),
          decoration: BoxDecoration(
            color: color.withOpacity(0.1),
            borderRadius: BorderRadius.circular(6),
          ),
          child: Icon(icon, size: 16, color: color),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: TextStyle(
                  fontSize: 13,
                  color: highlightColor != null
                      ? color
                      : const Color(0xFF8C92A4),
                  fontWeight: highlightColor != null
                      ? FontWeight.w600
                      : FontWeight.normal,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                value,
                style: const TextStyle(
                  fontSize: 14,
                  color: Color(0xFF20253A),
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  /// 构建选项开关（布尔类型）
  Widget _buildOptionSwitch({
    required String label,
    required bool value,
    required String description,
    required IconData icon,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.all(6),
          decoration: BoxDecoration(
            color: value
                ? const Color(0xFF20CB6B).withOpacity(0.1)
                : const Color(0xFF8C92A4).withOpacity(0.1),
            borderRadius: BorderRadius.circular(6),
          ),
          child: Icon(
            icon,
            size: 16,
            color: value ? const Color(0xFF20CB6B) : const Color(0xFF8C92A4),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Text(
                    label,
                    style: TextStyle(
                      fontSize: 14,
                      color: value
                          ? const Color(0xFF20253A)
                          : const Color(0xFF8C92A4),
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: value
                          ? const Color(0xFF20CB6B).withOpacity(0.1)
                          : const Color(0xFF8C92A4).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      value ? '已开启' : '未开启',
                      style: TextStyle(
                        fontSize: 11,
                        color: value
                            ? const Color(0xFF20CB6B)
                            : const Color(0xFF8C92A4),
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 4),
              Text(
                description,
                style: const TextStyle(fontSize: 12, color: Color(0xFF8C92A4)),
              ),
            ],
          ),
        ),
      ],
    );
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

  // 跳转到路线规划页面
  void _navigateToRoutePlanning() {
    Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (_) => const RoutePlanningView()));
  }

  // 构建底部操作栏（固定在DraggableScrollableSheet底部）
  Widget _buildActionBar() {
    final order = _orderData?['order'] as Map<String, dynamic>?;
    final status = order?['status'] as String? ?? '';

    // 待配送订单：显示接单按钮
    if (status == 'pending_delivery' || status == 'pending') {
      return Container(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border(top: BorderSide(color: Colors.grey[200]!, width: 1)),
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
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border(top: BorderSide(color: Colors.grey[200]!, width: 1)),
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
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border(top: BorderSide(color: Colors.grey[200]!, width: 1)),
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

    if (_orderData == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('订单信息未加载完成，请稍后再试'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    // 提取订单信息
    final addressData = _orderData?['address'] as Map<String, dynamic>?;
    final storeName = addressData?['name'] as String? ?? '门店名称未填写';
    final address = addressData?['address'] as String? ?? '';
    final orderItems = _orderData?['order_items'] as List<dynamic>? ?? [];
    final itemCount = orderItems.length;
    final order = _orderData?['order'] as Map<String, dynamic>?;
    final totalAmount = (order?['total_amount'] as num?)?.toDouble() ?? 0.0;
    final isUrgent = (order?['is_urgent'] as bool?) ?? false;
    final deliveryFeeCalc =
        _orderData?['delivery_fee_calculation'] as Map<String, dynamic>?;
    final riderPayableFee =
        (deliveryFeeCalc?['rider_payable_fee'] as num?)?.toDouble() ?? 0.0;

    // 显示确认对话框
    final confirmed = await AcceptOrderDialog.show(
      context,
      storeName: storeName,
      address: address,
      riderPayableFee: riderPayableFee,
      totalAmount: totalAmount,
      itemCount: itemCount,
      isUrgent: isUrgent,
    );

    if (confirmed != true) {
      return; // 用户取消接单
    }

    // 接单前检查位置权限和获取位置
    try {
      // 检查定位权限
      final hasPermission = await LocationService.checkAndRequestPermission();
      if (!hasPermission) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('无法获取位置信息，无法接单。请前往设置开启定位权限'),
            backgroundColor: Colors.red,
            duration: Duration(seconds: 3),
          ),
        );
        return;
      }

      // 获取当前位置（使用缓存的位置，如果可用）
      Position? position = LocationService.getCachedPosition();
      if (position == null) {
        position = await LocationService.getCurrentLocation();
      }

      if (position == null) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('无法获取当前位置，请确保已开启GPS定位服务'),
            backgroundColor: Colors.red,
            duration: Duration(seconds: 3),
          ),
        );
        return;
      }

      setState(() {
        _isProcessing = true;
      });

      // 传递位置信息接单
      final response = await OrderApi.acceptOrder(
        widget.orderId,
        latitude: position.latitude,
        longitude: position.longitude,
      );

      if (!mounted) return;

      if (response.isSuccess) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('接单成功'),
            backgroundColor: Color(0xFF20CB6B),
          ),
        );
        // 清除路线预览数据（确保不显示上一趟的路线）
        setState(() {
          _routePreviewOrders = [];
        });
        // 等待一小段时间，让后端有时间计算新路线（异步计算需要时间）
        await Future.delayed(const Duration(milliseconds: 800));
        // 静默刷新订单详情（不显示加载状态，避免白屏）
        await _loadOrderDetailSilently();
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
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('接单失败: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
      if (mounted) {
        setState(() {
          _isProcessing = false;
        });
      }
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
    // 无论返回什么，都刷新订单详情（因为取货操作可能已完成）
    if (mounted) {
      await _loadOrderDetail();
      // 从批量取货页面返回后，强制重新启动位置跟踪，确保 CurrentLocationLayer 能显示
      print('[OrderDetailView] 从批量取货页面返回，强制重新启动位置跟踪');
      // 先尝试使用缓存位置立即显示
      _tryUseCachedPosition();
      // 如果已有位置，立即发送到流中，确保 CurrentLocationLayer 能显示
      if (_userPosition != null) {
        print(
          '[OrderDetailView] 从批量取货页面返回，立即发送位置到流中: ${_userPosition!.latitude}, ${_userPosition!.longitude}',
        );
        _locationStreamController.add(
          LocationMarkerPosition(
            latitude: _userPosition!.latitude,
            longitude: _userPosition!.longitude,
            accuracy: _userPosition!.accuracy,
          ),
        );
      }
      // 无论位置流是否还在运行，都重新启动位置跟踪（确保位置流持续运行）
      // 先取消旧的订阅（如果存在），然后重新启动
      _positionStreamSubscription?.cancel();
      _positionStreamSubscription = null;
      // 使用延迟确保页面已完全恢复，然后重新启动位置跟踪
      Future.delayed(const Duration(milliseconds: 300), () {
        if (mounted) {
          print('[OrderDetailView] 从批量取货页面返回，重新启动位置跟踪');
          // 重新启动位置跟踪，确保位置流持续运行
          _startLocationTracking();
        }
      });
      // 如果订单状态变为配送中，返回true通知列表刷新
      final status = _orderData?['status'] as String?;
      if (status == 'delivering') {
        Navigator.of(context).pop(true);
      } else if (result == true) {
        // 如果取货成功但状态未变，也返回true以触发刷新
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
