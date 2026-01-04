import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_map_location_marker/flutter_map_location_marker.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:map_launcher/map_launcher.dart';
import 'dart:async';
import 'dart:math' as math;
import '../utils/location_service.dart';
import '../utils/coordinate_transform.dart';
import '../api/order_api.dart';
import 'batch_pickup_items_view.dart';
import 'order_detail_view.dart';

/// 路线规划页面：使用 flutter_map 显示天地图图层
class RoutePlanningView extends StatefulWidget {
  const RoutePlanningView({super.key});

  @override
  State<RoutePlanningView> createState() => _RoutePlanningViewState();
}

class _RoutePlanningViewState extends State<RoutePlanningView>
    with WidgetsBindingObserver {
  // 地图控制器
  final MapController _mapController = MapController();

  // 拖拽控制器（用于监听拖拽状态）
  final DraggableScrollableController _draggableController =
      DraggableScrollableController();

  // 底部悬浮框是否展开（用于控制定位图标和提醒框的显示）
  bool _isSheetExpanded = false;

  // 用户位置（保留用于显示状态）
  Position? _userPosition;
  // ignore: unused_field
  bool _isLoadingLocation = false;
  // ignore: unused_field
  String? _locationError;

  // 位置流控制器（用于 flutter_map_location_marker）
  final StreamController<LocationMarkerPosition?> _locationStreamController =
      StreamController<LocationMarkerPosition?>.broadcast();

  // 位置更新订阅
  StreamSubscription<Position>? _positionStreamSubscription;

  // 地图初始中心点（北京天安门）
  final LatLng _initialCenter = const LatLng(39.90864, 116.39750);
  final double _initialZoom = 15.0;

  // 配送订单相关
  List<Map<String, dynamic>> _deliveringOrders = [];
  bool _isLoadingOrders = false;

  // 待取货供应商相关
  List<Map<String, dynamic>> _pickupSuppliers = [];
  bool _isLoadingSuppliers = false;
  bool _hasPendingPickup = false; // 是否有待取货订单
  bool _hasCompletedAllPickup = false; // 是否已完成全部取货

  // 天地图瓦片服务 URL 模板（Web墨卡托投影）
  // 使用WMTS格式，与您提供的正确URL格式一致
  // 影像底图（img_w）：Web墨卡托投影的影像底图
  // 注意：使用您提供的新密钥
  static const String _tiandituTileUrlTemplate =
      'https://t{s}.tianditu.gov.cn/img_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=img&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=d95864378581051adb04fe26acb13ecf';

  // 天地图影像标注图层（可选，叠加在底图上）
  // 使用Web墨卡托投影的影像标注图层
  static const String _tiandituLabelUrlTemplate =
      'https://t{s}.tianditu.gov.cn/cia_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=cia&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=d95864378581051adb04fe26acb13ecf';

  /// 创建自定义的TileProvider，添加浏览器请求头以解决403错误
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
    // 监听应用生命周期
    WidgetsBinding.instance.addObserver(this);
    // 监听拖拽状态变化
    _draggableController.addListener(_onDraggableChanged);
    // 页面加载时获取用户位置并开始监听位置更新
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _startLocationTracking();
      _checkAndLoadPickupSuppliers();
    });
  }

  /// 监听拖拽状态变化
  void _onDraggableChanged() {
    final isExpanded = _draggableController.size > 0.5; // 超过50%认为已展开
    if (_isSheetExpanded != isExpanded) {
      setState(() {
        _isSheetExpanded = isExpanded;
      });
    }
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    // 当应用从后台返回前台时，重新检查状态并刷新路线
    if (state == AppLifecycleState.resumed) {
      _checkAndLoadPickupSuppliers();
      // 如果已完成全部取货且有配送中订单，重新加载订单列表以刷新路线
      if (_hasCompletedAllPickup && _userPosition != null) {
        _loadDeliveringOrders();
      }
    }
  }

  /// 检查是否有待取货订单，如果有则加载供应商列表
  Future<void> _checkAndLoadPickupSuppliers() async {
    // 先检查是否有待取货订单
    final pickupResponse = await OrderApi.getOrderPool(
      pageNum: 1,
      pageSize: 1,
      status: 'pending_pickup',
    );

    if (!mounted) return;

    if (pickupResponse.isSuccess && pickupResponse.data != null) {
      final data = pickupResponse.data!;
      final total = data['total'] as int? ?? 0;

      if (total > 0) {
        // 有待取货订单，加载供应商列表
        setState(() {
          _hasPendingPickup = true;
          _hasCompletedAllPickup = false;
        });
        await _loadPickupSuppliers();
      } else {
        // 没有待取货订单，检查是否有配送中的订单
        setState(() {
          _hasPendingPickup = false;
        });
        await _loadDeliveringOrders();
        // 如果有配送中的订单，说明已完成全部取货
        if (mounted) {
          final hasOrders = _deliveringOrders.isNotEmpty;
          print(
            '[RoutePlanningView] 加载配送中订单完成 - 订单数量: ${_deliveringOrders.length}, 用户位置: ${_userPosition != null}',
          );
          setState(() {
            _hasCompletedAllPickup = hasOrders;
          });
        }
      }
    } else {
      // 加载失败，尝试加载配送中的订单
      await _loadDeliveringOrders();
      if (mounted) {
        final hasOrders = _deliveringOrders.isNotEmpty;
        print(
          '[RoutePlanningView] 加载配送中订单完成（失败分支） - 订单数量: ${_deliveringOrders.length}, 用户位置: ${_userPosition != null}',
        );
        setState(() {
          _hasCompletedAllPickup = hasOrders;
        });
      }
    }
  }

  /// 加载待取货供应商列表
  Future<void> _loadPickupSuppliers() async {
    if (_isLoadingSuppliers) return;

    setState(() {
      _isLoadingSuppliers = true;
    });

    // 传递配送员当前位置作为起点
    final response = await OrderApi.getPickupSuppliers(
      latitude: _userPosition?.latitude,
      longitude: _userPosition?.longitude,
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      List<Map<String, dynamic>> suppliers = (response.data as List<dynamic>)
          .cast<Map<String, dynamic>>();

      // 后端已经按照路线优化算法排序，直接使用
      // 如果有用户位置，计算距离（用于显示，但不改变排序）
      if (_userPosition != null) {
        for (var supplier in suppliers) {
          final lat = (supplier['latitude'] as num?)?.toDouble();
          final lng = (supplier['longitude'] as num?)?.toDouble();
          if (lat != null && lng != null) {
            // 将GCJ-02坐标转换为WGS84坐标后再计算距离
            final wgs84Point = CoordinateTransform.gcj02ToWgs84(lat, lng);
            final distance = Geolocator.distanceBetween(
              _userPosition!.latitude,
              _userPosition!.longitude,
              wgs84Point.latitude,
              wgs84Point.longitude,
            );
            supplier['distance'] = distance;
          }
        }
      }

      setState(() {
        _pickupSuppliers = suppliers;
        _isLoadingSuppliers = false;
      });

      // 更新地图显示
      _updateMapForSuppliers();
    } else {
      setState(() {
        _isLoadingSuppliers = false;
      });
    }
  }

  /// 按距离排序供应商（由近到远）
  List<Map<String, dynamic>> _sortSuppliersByDistance(
    List<Map<String, dynamic>> suppliers,
  ) {
    if (_userPosition == null) return suppliers;

    final sorted = List<Map<String, dynamic>>.from(suppliers);
    sorted.sort((a, b) {
      final latA = (a['latitude'] as num?)?.toDouble();
      final lngA = (a['longitude'] as num?)?.toDouble();
      final latB = (b['latitude'] as num?)?.toDouble();
      final lngB = (b['longitude'] as num?)?.toDouble();

      if (latA == null || lngA == null) return 1;
      if (latB == null || lngB == null) return -1;

      // 将GCJ-02坐标转换为WGS84坐标后再计算距离
      final wgs84A = CoordinateTransform.gcj02ToWgs84(latA, lngA);
      final wgs84B = CoordinateTransform.gcj02ToWgs84(latB, lngB);

      final distanceA = Geolocator.distanceBetween(
        _userPosition!.latitude,
        _userPosition!.longitude,
        wgs84A.latitude,
        wgs84A.longitude,
      );
      final distanceB = Geolocator.distanceBetween(
        _userPosition!.latitude,
        _userPosition!.longitude,
        wgs84B.latitude,
        wgs84B.longitude,
      );

      return distanceA.compareTo(distanceB);
    });

    // 添加距离信息到供应商数据中
    for (var supplier in sorted) {
      final lat = (supplier['latitude'] as num?)?.toDouble();
      final lng = (supplier['longitude'] as num?)?.toDouble();
      if (lat != null && lng != null && _userPosition != null) {
        // 将GCJ-02坐标转换为WGS84坐标后再计算距离
        final wgs84Point = CoordinateTransform.gcj02ToWgs84(lat, lng);
        final distance = Geolocator.distanceBetween(
          _userPosition!.latitude,
          _userPosition!.longitude,
          wgs84Point.latitude,
          wgs84Point.longitude,
        );
        supplier['distance'] = distance;
      }
    }

    return sorted;
  }

  /// 更新地图以显示供应商路线
  void _updateMapForSuppliers() {
    if (_pickupSuppliers.isEmpty || _userPosition == null) return;

    // 计算所有供应商的边界，以便调整地图视野
    double? minLat, maxLat, minLng, maxLng;

    // 先添加用户位置（WGS84）
    minLat = maxLat = _userPosition!.latitude;
    minLng = maxLng = _userPosition!.longitude;

    // 添加所有供应商位置（转换为WGS84）
    for (var supplier in _pickupSuppliers) {
      final lat = (supplier['latitude'] as num?)?.toDouble();
      final lng = (supplier['longitude'] as num?)?.toDouble();
      if (lat != null && lng != null) {
        // 将GCJ-02坐标转换为WGS84坐标
        final wgs84Point = CoordinateTransform.gcj02ToWgs84(lat, lng);
        minLat = math.min(minLat!, wgs84Point.latitude);
        maxLat = math.max(maxLat!, wgs84Point.latitude);
        minLng = math.min(minLng!, wgs84Point.longitude);
        maxLng = math.max(maxLng!, wgs84Point.longitude);
      }
    }

    // 调整地图视野以包含所有点
    if (minLat != null && maxLat != null && minLng != null && maxLng != null) {
      final centerLat = (minLat + maxLat) / 2;
      final centerLng = (minLng + maxLng) / 2;
      final latDiff = maxLat - minLat;
      final lngDiff = maxLng - minLng;
      final maxDiff = math.max(latDiff, lngDiff);

      // 计算合适的缩放级别
      double zoom = 15.0;
      if (maxDiff > 0.1) {
        zoom = 12.0;
      } else if (maxDiff > 0.05) {
        zoom = 13.0;
      } else if (maxDiff > 0.02) {
        zoom = 14.0;
      }

      _mapController.move(LatLng(centerLat, centerLng), zoom);
    }
  }

  /// 加载配送中的订单（使用排序后的订单列表）
  Future<void> _loadDeliveringOrders() async {
    if (_isLoadingOrders) return;

    setState(() {
      _isLoadingOrders = true;
    });

    // 使用新的排序后订单API
    final response = await OrderApi.getRouteOrders();

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final List<dynamic> list = (data['list'] as List<dynamic>? ?? []);
      final orders = list.cast<Map<String, dynamic>>();

      // 检查是否所有订单都是已送达状态
      final allDelivered =
          orders.isNotEmpty &&
          orders.every((order) {
            final status = order['status'] as String? ?? '';
            return status == 'delivered' || status == 'shipped';
          });

      // 如果所有订单都是已送达，清空列表以显示"暂无配送订单"
      // 否则，显示所有订单（包括已送达的订单，它们会被标记为"已送达"）
      final shouldClearOrders = allDelivered;
      setState(() {
        _deliveringOrders = shouldClearOrders ? [] : orders;
        _isLoadingOrders = false;
      });

      // 如果清空了订单列表且拖拽框是展开的，将其折叠
      if (shouldClearOrders && _isSheetExpanded && mounted) {
        _draggableController.animateTo(
          0.4, // 折叠到默认高度
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      }
    } else {
      // 如果获取排序订单失败，回退到原来的方式
      print('[RoutePlanningView] 获取排序订单失败，使用原方式: ${response.message}');
      final fallbackResponse = await OrderApi.getOrderPool(
        pageNum: 1,
        pageSize: 100,
        status: 'delivering',
      );
      if (fallbackResponse.isSuccess && fallbackResponse.data != null) {
        final fallbackData = fallbackResponse.data!;
        final List<dynamic> fallbackList =
            (fallbackData['list'] as List<dynamic>? ?? []);
        setState(() {
          _deliveringOrders = fallbackList.cast<Map<String, dynamic>>();
          _isLoadingOrders = false;
        });
      } else {
        setState(() {
          _isLoadingOrders = false;
        });
      }
    }
  }

  /// 导航到供应商位置
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

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _draggableController.removeListener(_onDraggableChanged);
    _draggableController.dispose();
    _positionStreamSubscription?.cancel();
    _locationStreamController.close();
    _mapController.dispose();
    super.dispose();
  }

  /// 显示定位服务未启用对话框
  Future<bool> _showLocationServiceDialog() async {
    return await showDialog<bool>(
          context: context,
          barrierDismissible: false,
          builder: (BuildContext context) {
            return AlertDialog(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
              title: Row(
                children: [
                  Icon(Icons.location_off, color: Colors.orange[700]),
                  const SizedBox(width: 8),
                  const Text(
                    '定位服务未启用',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                ],
              ),
              content: const Text(
                '为了提供路线规划服务，需要开启系统定位服务。\n\n请点击"去设置"打开系统定位设置。',
                style: TextStyle(fontSize: 15, color: Color(0xFF40475C)),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: const Text(
                    '取消',
                    style: TextStyle(color: Color(0xFF8C92A4)),
                  ),
                ),
                ElevatedButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF20CB6B),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                  child: const Text(
                    '去设置',
                    style: TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            );
          },
        ) ??
        false;
  }

  /// 显示权限说明对话框（针对小米手机等可能不弹出系统对话框的情况）
  Future<bool> _showPermissionDialog() async {
    return await showDialog<bool>(
          context: context,
          barrierDismissible: false,
          builder: (BuildContext context) {
            return AlertDialog(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
              title: Row(
                children: [
                  Icon(Icons.location_disabled, color: Colors.orange[700]),
                  const SizedBox(width: 8),
                  const Text(
                    '需要定位权限',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                ],
              ),
              content: const Text(
                '为了提供路线规划服务，需要获取您的位置信息。\n\n请点击"去设置"手动开启定位权限。',
                style: TextStyle(fontSize: 15, color: Color(0xFF40475C)),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: const Text(
                    '取消',
                    style: TextStyle(color: Color(0xFF8C92A4)),
                  ),
                ),
                ElevatedButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF20CB6B),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                  child: const Text(
                    '去设置',
                    style: TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            );
          },
        ) ??
        false;
  }

  /// 开始位置跟踪（使用 flutter_map_location_marker）
  Future<void> _startLocationTracking() async {
    print('[RoutePlanningView] 开始位置跟踪...');

    if (mounted) {
      setState(() {
        _isLoadingLocation = true;
        _locationError = null;
      });
    }

    // 检查定位服务是否启用
    final serviceEnabled = await LocationService.checkLocationServiceEnabled();
    print('[RoutePlanningView] 定位服务状态: $serviceEnabled');

    // 检查并请求权限
    final hasPermission = await LocationService.checkAndRequestPermission();
    print('[RoutePlanningView] 权限检查结果: $hasPermission');

    if (!hasPermission) {
      // 检查权限状态，给出更具体的提示
      final permission = await Geolocator.checkPermission();
      final permissionHandlerStatus = await Permission.location.status;

      String errorMsg = '定位权限未授予';
      bool needShowDialog = false;

      if (permission == LocationPermission.deniedForever ||
          permissionHandlerStatus.isPermanentlyDenied) {
        errorMsg = '定位权限被永久拒绝，请到设置中开启';
        needShowDialog = true;
      } else if (permission == LocationPermission.denied ||
          permissionHandlerStatus.isDenied) {
        errorMsg = '定位权限未授予（小米手机请到设置中手动开启）';
        needShowDialog = true;
      }

      if (mounted) {
        setState(() {
          _isLoadingLocation = false;
          _locationError = errorMsg;
        });
      }

      if (needShowDialog && mounted) {
        final shouldOpenSettings = await _showPermissionDialog();
        if (shouldOpenSettings) {
          await LocationService.openAppSettingsPage();
          await Future.delayed(const Duration(seconds: 2));
          _startLocationTracking();
        }
      }
      return;
    }

    // 即使定位服务未启用，也尝试使用网络定位
    // 开始监听位置更新（使用降级策略）
    try {
      // 如果定位服务未启用，先尝试获取一次网络定位
      if (!serviceEnabled) {
        print('[RoutePlanningView] 定位服务未启用，尝试使用网络定位...');
        final networkPosition = await LocationService.getCurrentLocation();
        if (networkPosition != null && mounted) {
          print('[RoutePlanningView] 网络定位成功，继续使用网络定位');
          // 立即将初始位置发送到流中，以便 CurrentLocationLayer 能显示
          _locationStreamController.add(
            LocationMarkerPosition(
              latitude: networkPosition.latitude,
              longitude: networkPosition.longitude,
              accuracy: networkPosition.accuracy,
            ),
          );
          final wasFirstLocation = _userPosition == null;
          setState(() {
            _userPosition = networkPosition;
            _isLoadingLocation = false;
            _locationError = null;
          });
          print(
            '[RoutePlanningView] 网络定位成功 - 首次定位: $wasFirstLocation, 坐标: ${networkPosition.latitude}, ${networkPosition.longitude}',
          );
          print(
            '[RoutePlanningView] 当前状态 - 待取货: $_hasPendingPickup, 已完成取货: $_hasCompletedAllPickup, 配送中订单: ${_deliveringOrders.length}',
          );
          _mapController.move(
            LatLng(networkPosition.latitude, networkPosition.longitude),
            _initialZoom,
          );
          // 网络定位成功，启动定位流（使用低精度）
          _startPositionStreamWithFallback();
          return;
        } else {
          // 网络定位也失败，提示用户开启GPS
          if (mounted) {
            setState(() {
              _isLoadingLocation = false;
              _locationError = '定位服务未启用，请先开启GPS';
            });
            // 显示对话框引导用户打开系统设置
            final shouldOpenSettings = await _showLocationServiceDialog();
            if (shouldOpenSettings) {
              await LocationService.openLocationSettings();
              // 延迟一下，等待用户操作
              await Future.delayed(const Duration(seconds: 2));
              // 重新检查定位服务
              _startLocationTracking();
            }
          }
          return;
        }
      }

      // 定位服务已启用，正常启动定位流
      _startPositionStreamWithFallback();

      // 立即获取一次位置（用于初始定位）
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
        final wasFirstLocation = _userPosition == null;
        print(
          '[RoutePlanningView] 初始位置获取成功 - 首次定位: $wasFirstLocation, 坐标: ${initialPosition.latitude}, ${initialPosition.longitude}',
        );
        print(
          '[RoutePlanningView] 当前状态 - 待取货: $_hasPendingPickup, 已完成取货: $_hasCompletedAllPickup, 配送中订单: ${_deliveringOrders.length}',
        );
        setState(() {
          _userPosition = initialPosition;
          _isLoadingLocation = false;
          _locationError = null;
        });
        _mapController.move(
          LatLng(initialPosition.latitude, initialPosition.longitude),
          _initialZoom,
        );
        // 获取到位置后，重新计算路线
        _recalculateRoutesWithNewLocation();
      } else if (mounted) {
        // 如果获取位置失败，但定位流已启动，等待定位流更新
        setState(() {
          _isLoadingLocation = true;
          _locationError = null;
        });
      }
    } catch (e) {
      print('[RoutePlanningView] 启动位置跟踪失败: $e');
      if (mounted) {
        setState(() {
          _isLoadingLocation = false;
          _locationError = '启动位置跟踪失败: ${e.toString()}';
        });
      }
    }
  }

  /// 启动定位流（带降级策略，优先网络定位）
  void _startPositionStreamWithFallback() {
    // 优先使用网络定位（低精度），在中国更可靠
    _tryStartPositionStream(LocationAccuracy.low, () {
      // 如果低精度失败，尝试最低精度
      print('[RoutePlanningView] 网络定位流失败，尝试最低精度');
      Future.delayed(const Duration(seconds: 1), () {
        if (mounted) {
          _tryStartPositionStream(LocationAccuracy.lowest, () {
            // 如果最低精度也失败，尝试中等精度（GPS + 网络）
            print('[RoutePlanningView] 最低精度定位流失败，尝试中等精度（GPS+网络）');
            Future.delayed(const Duration(seconds: 1), () {
              if (mounted) {
                _tryStartPositionStream(LocationAccuracy.medium, () {
                  // 最后尝试高精度GPS（在中国可能失败）
                  print('[RoutePlanningView] 中等精度定位流失败，尝试高精度GPS');
                  Future.delayed(const Duration(seconds: 1), () {
                    if (mounted) {
                      _tryStartPositionStream(LocationAccuracy.high, () {
                        print('[RoutePlanningView] 所有精度级别都失败，定位流无法启动');
                        if (mounted) {
                          setState(() {
                            _isLoadingLocation = false;
                            _locationError = '定位失败，请检查网络和GPS设置';
                          });
                        }
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
              distanceFilter: 10, // 每移动10米更新一次
            ),
          ).listen(
            (Position position) {
              print(
                '[RoutePlanningView] 位置更新: ${position.latitude}, ${position.longitude}, 精度: ${position.accuracy}米',
              );

              // 更新位置流
              _locationStreamController.add(
                LocationMarkerPosition(
                  latitude: position.latitude,
                  longitude: position.longitude,
                  accuracy: position.accuracy,
                ),
              );

              // 更新状态
              final wasFirstLocation = _userPosition == null;
              if (mounted) {
                setState(() {
                  _userPosition = position;
                  _isLoadingLocation = false;
                  _locationError = null;
                });
              }

              print(
                '[RoutePlanningView] 位置流更新 - 首次定位: $wasFirstLocation, 坐标: ${position.latitude}, ${position.longitude}',
              );
              print(
                '[RoutePlanningView] 当前状态 - 待取货: $_hasPendingPickup, 已完成取货: $_hasCompletedAllPickup, 配送中订单: ${_deliveringOrders.length}',
              );

              // 首次定位时，将地图中心移动到用户位置，并重新计算路线
              if (wasFirstLocation) {
                _mapController.move(
                  LatLng(position.latitude, position.longitude),
                  _initialZoom,
                );
                // 获取到位置后，重新计算路线
                _recalculateRoutesWithNewLocation();
              } else {
                // 位置更新后，如果有待取货供应商，重新计算距离并排序
                if (_hasPendingPickup && _pickupSuppliers.isNotEmpty) {
                  // 重新计算距离（因为用户位置已更新）
                  final sorted = _sortSuppliersByDistance(_pickupSuppliers);
                  if (mounted) {
                    setState(() {
                      _pickupSuppliers = sorted;
                    });
                  }
                }
                // 如果已完成全部取货且有配送中订单，触发 setState 以重新计算路线显示顺序
                // 路线会根据新的配送员位置自动重新排序（在 build 方法中计算）
                if (_hasCompletedAllPickup &&
                    _deliveringOrders.isNotEmpty &&
                    mounted) {
                  setState(() {
                    // 不需要修改数据，只是触发重新构建，路线会根据新位置自动重新排序
                  });
                }
              }
            },
            onError: (error) {
              print('[RoutePlanningView] 位置流错误 (精度: $accuracy): $error');
              // 如果当前精度失败，尝试降级
              onError();
            },
            cancelOnError: false, // 不因错误而取消流
          );
      print('[RoutePlanningView] 定位流启动成功 (精度: $accuracy)');
    } catch (e) {
      print('[RoutePlanningView] 启动定位流失败 (精度: $accuracy): $e');
      onError();
    }
  }

  /// 获取用户位置（保留用于兼容）
  Future<void> _getUserLocation() async {
    print('[RoutePlanningView] 开始获取用户位置...');

    if (mounted) {
      setState(() {
        _isLoadingLocation = true;
        _locationError = null;
      });
    }

    // 检查定位服务是否启用
    final serviceEnabled = await LocationService.checkLocationServiceEnabled();
    print('[RoutePlanningView] 定位服务状态: $serviceEnabled');

    // 检查并请求权限
    final hasPermission = await LocationService.checkAndRequestPermission();
    print('[RoutePlanningView] 权限检查结果: $hasPermission');

    if (!hasPermission) {
      // 检查权限状态，给出更具体的提示
      final permission = await Geolocator.checkPermission();
      final permissionHandlerStatus = await Permission.location.status;

      String errorMsg = '定位权限未授予';
      bool needShowDialog = false;

      if (permission == LocationPermission.deniedForever ||
          permissionHandlerStatus.isPermanentlyDenied) {
        errorMsg = '定位权限被永久拒绝，请到设置中开启';
        needShowDialog = true;
      } else if (permission == LocationPermission.denied ||
          permissionHandlerStatus.isDenied) {
        // 小米手机可能没有弹出对话框，显示自定义对话框引导用户
        errorMsg = '定位权限未授予（小米手机请到设置中手动开启）';
        needShowDialog = true;
      }

      if (mounted) {
        setState(() {
          _isLoadingLocation = false;
          _locationError = errorMsg;
        });
      }

      // 如果是权限被拒绝或未授予，显示对话框引导用户
      if (needShowDialog && mounted) {
        final shouldOpenSettings = await _showPermissionDialog();
        if (shouldOpenSettings) {
          await LocationService.openAppSettingsPage();
          // 延迟一下，等待用户操作
          await Future.delayed(const Duration(seconds: 2));
          // 重新检查权限
          _getUserLocation();
        }
      }
      return;
    }

    // 如果定位服务未启用，先尝试使用网络定位（某些设备可能仍能获取到位置）
    // 使用 forceAndroidLocationManager: true 时，即使定位服务未启用，也可能通过网络定位获取位置
    if (!serviceEnabled) {
      print('[RoutePlanningView] 定位服务未启用，尝试使用网络定位...');
      try {
        // 尝试直接获取位置（使用网络定位，不依赖GPS）
        final position = await LocationService.getCurrentLocationDirect();
        if (position != null) {
          print('[RoutePlanningView] 网络定位成功，即使定位服务未启用');
          final wasFirstLocation = _userPosition == null;
          if (mounted) {
            setState(() {
              _isLoadingLocation = false;
              _userPosition = position;
              _locationError = null;
              // 将地图中心移动到用户位置
              _mapController.move(
                LatLng(_userPosition!.latitude, _userPosition!.longitude),
                _initialZoom,
              );
            });
            print(
              '[RoutePlanningView] 网络定位成功（_getUserLocation） - 首次定位: $wasFirstLocation, 坐标: ${position.latitude}, ${position.longitude}',
            );
            print(
              '[RoutePlanningView] 当前状态 - 待取货: $_hasPendingPickup, 已完成取货: $_hasCompletedAllPickup, 配送中订单: ${_deliveringOrders.length}',
            );
            // 获取到位置后，重新计算路线
            _recalculateRoutesWithNewLocation();
          }
          return;
        } else {
          print('[RoutePlanningView] 网络定位也失败，需要开启GPS');
        }
      } catch (e) {
        print('[RoutePlanningView] 网络定位异常: $e');
      }

      // 如果网络定位也失败，提示用户开启GPS
      if (mounted) {
        setState(() {
          _isLoadingLocation = false;
          _locationError = '定位服务未启用，请先开启GPS';
        });
        // 显示对话框引导用户打开系统设置
        final shouldOpenSettings = await _showLocationServiceDialog();
        if (shouldOpenSettings) {
          await LocationService.openLocationSettings();
          // 延迟一下，等待用户操作
          await Future.delayed(const Duration(seconds: 2));
          // 重新检查定位服务
          _getUserLocation();
        }
      }
      return;
    }

    // 定位服务已启用，正常获取位置
    final position = await LocationService.getCurrentLocation();
    print('[RoutePlanningView] 定位结果: ${position != null ? "成功" : "失败"}');

    if (mounted) {
      setState(() {
        _isLoadingLocation = false;
        if (position == null) {
          if (!serviceEnabled) {
            _locationError = '定位服务未启用，请先开启GPS';
          } else {
            _locationError = '定位失败，请检查GPS设置';
          }
        } else {
          final wasFirstLocation = _userPosition == null;
          print(
            '[RoutePlanningView] 获取位置成功（_getUserLocation） - 首次定位: $wasFirstLocation, 坐标: ${position.latitude}, ${position.longitude}',
          );
          print(
            '[RoutePlanningView] 当前状态 - 待取货: $_hasPendingPickup, 已完成取货: $_hasCompletedAllPickup, 配送中订单: ${_deliveringOrders.length}',
          );
          _userPosition = position;
          _locationError = null;
          // 将地图中心移动到用户位置
          _mapController.move(
            LatLng(_userPosition!.latitude, _userPosition!.longitude),
            _initialZoom,
          );
          // 获取到位置后，重新计算路线
          _recalculateRoutesWithNewLocation();
        }
      });
    }
  }

  /// 根据新的配送员位置重新计算路线
  Future<void> _recalculateRoutesWithNewLocation() async {
    if (_userPosition == null) return;

    print(
      '[RoutePlanningView] 开始重新计算路线 - 配送员位置: ${_userPosition!.latitude}, ${_userPosition!.longitude}',
    );

    if (_hasPendingPickup) {
      // 如果有待取货订单，重新加载供应商列表（会传递新位置，后端会重新计算）
      print('[RoutePlanningView] 重新加载供应商列表（使用新位置）');
      await _loadPickupSuppliers();
    } else if (_hasCompletedAllPickup && _deliveringOrders.isNotEmpty) {
      // 如果已完成全部取货且有配送中订单，重新计算配送路线
      print('[RoutePlanningView] 重新计算配送路线（使用新位置）');
      try {
        // 调用后端API重新计算路线
        await OrderApi.calculateRoute(
          latitude: _userPosition!.latitude,
          longitude: _userPosition!.longitude,
        );
        // 等待一小段时间让后端完成计算
        await Future.delayed(const Duration(milliseconds: 500));
        // 重新加载排序后的订单列表
        await _loadDeliveringOrders();
      } catch (e) {
        print('[RoutePlanningView] 重新计算配送路线失败: $e');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('路线规划'),
            if (_userPosition != null) ...[
              const SizedBox(width: 8),
              Text(
                '${_userPosition!.latitude.toStringAsFixed(6)}, ${_userPosition!.longitude.toStringAsFixed(6)}',
                style: const TextStyle(
                  color: Colors.white70,
                  fontSize: 11,
                  fontWeight: FontWeight.normal,
                ),
              ),
            ],
          ],
        ),
        backgroundColor: const Color(0xFF20CB6B),
        iconTheme: const IconThemeData(color: Colors.white),
        titleTextStyle: const TextStyle(
          color: Colors.white,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      body: Stack(
        children: [
          // 高德地图
          FlutterMap(
            mapController: _mapController,
            options: MapOptions(
              initialCenter: _initialCenter,
              initialZoom: _initialZoom,
              minZoom: 3.0,
              maxZoom: 18.0,
            ),
            children: [
              // 天地图影像底图图层（Web墨卡托投影）
              TileLayer(
                urlTemplate: _tiandituTileUrlTemplate,
                subdomains: const ['0', '1', '2', '3', '4', '5', '6', '7'],
                userAgentPackageName: 'com.example.distribution_app',
                maxNativeZoom: 18,
                maxZoom: 18,
                tileProvider: createTiandituTileProvider(),
              ),
              // 天地图影像标注图层（叠加在底图上，显示地名、道路等信息）
              TileLayer(
                urlTemplate: _tiandituLabelUrlTemplate,
                subdomains: const ['0', '1', '2', '3', '4', '5', '6', '7'],
                userAgentPackageName: 'com.example.distribution_app',
                maxNativeZoom: 18,
                maxZoom: 18,
                tileProvider: createTiandituTileProvider(),
              ),
              // 用户位置标记图层（使用 flutter_map_location_marker）
              CurrentLocationLayer(
                positionStream: _locationStreamController.stream,
              ),
              // 配送路线（已完成取货时显示）- 根据配送员当前位置重新规划最近的待配送路线
              // 注意：序号（route_sequence）保持不变，只是路线显示顺序根据距离重新规划
              // 只连接未完成的订单，已送达订单只显示marker，不参与路线计算
              if (_hasCompletedAllPickup &&
                  _deliveringOrders.isNotEmpty &&
                  _userPosition != null)
                PolylineLayer(
                  polylines: [
                    Polyline(
                      points: [
                        // 起点：配送员位置
                        LatLng(
                          _userPosition!.latitude,
                          _userPosition!.longitude,
                        ),
                        // 根据配送员当前位置对未完成订单进行距离排序，然后显示路线
                        // 序号（route_sequence）保持不变，只是路线显示顺序改变
                        ...(() {
                          // 获取所有未完成的订单
                          final incompleteOrders = _deliveringOrders.where((
                            order,
                          ) {
                            final lat = order['latitude'] as num?;
                            final lng = order['longitude'] as num?;
                            if (lat == null || lng == null) return false;
                            // 只包含未完成的订单
                            final status = order['status'] as String? ?? '';
                            return status != 'delivered' && status != 'shipped';
                          }).toList();

                          // 如果只有一个或没有未完成订单，直接返回
                          if (incompleteOrders.length <= 1) {
                            return incompleteOrders
                                .map((order) {
                                  final lat = (order['latitude'] as num?)
                                      ?.toDouble();
                                  final lng = (order['longitude'] as num?)
                                      ?.toDouble();
                                  if (lat == null || lng == null) return null;
                                  // 将GCJ-02坐标转换为WGS84坐标
                                  final wgs84Point =
                                      CoordinateTransform.gcj02ToWgs84(
                                        lat,
                                        lng,
                                      );
                                  return wgs84Point;
                                })
                                .where((p) => p != null)
                                .cast<LatLng>();
                          }

                          // 根据配送员当前位置计算距离并排序
                          final sortedOrders = List<Map<String, dynamic>>.from(
                            incompleteOrders,
                          );
                          sortedOrders.sort((a, b) {
                            final latA = (a['latitude'] as num?)?.toDouble();
                            final lngA = (a['longitude'] as num?)?.toDouble();
                            final latB = (b['latitude'] as num?)?.toDouble();
                            final lngB = (b['longitude'] as num?)?.toDouble();

                            if (latA == null || lngA == null) return 1;
                            if (latB == null || lngB == null) return -1;

                            // 计算到配送员位置的距离
                            final distanceA = Geolocator.distanceBetween(
                              _userPosition!.latitude,
                              _userPosition!.longitude,
                              latA,
                              lngA,
                            );
                            final distanceB = Geolocator.distanceBetween(
                              _userPosition!.latitude,
                              _userPosition!.longitude,
                              latB,
                              lngB,
                            );

                            return distanceA.compareTo(distanceB);
                          });

                          // 返回排序后的坐标点
                          return sortedOrders
                              .map((order) {
                                final lat = (order['latitude'] as num?)
                                    ?.toDouble();
                                final lng = (order['longitude'] as num?)
                                    ?.toDouble();
                                if (lat == null || lng == null) return null;
                                // 将GCJ-02坐标转换为WGS84坐标
                                final wgs84Point =
                                    CoordinateTransform.gcj02ToWgs84(lat, lng);
                                return wgs84Point;
                              })
                              .where((p) => p != null)
                              .cast<LatLng>();
                        })(),
                      ],
                      strokeWidth: 4,
                      color: const Color(0xFF20CB6B).withOpacity(0.7),
                    ),
                  ],
                ),
              // 取货路线（待取货时显示）- 先渲染线条，避免遮挡标记
              if (_hasPendingPickup &&
                  _pickupSuppliers.isNotEmpty &&
                  _userPosition != null)
                PolylineLayer(
                  polylines: [
                    Polyline(
                      points: [
                        // 起点：配送员位置
                        LatLng(
                          _userPosition!.latitude,
                          _userPosition!.longitude,
                        ),
                        // 依次连接各个供应商（已按距离排序）
                        ..._pickupSuppliers
                            .where(
                              (supplier) =>
                                  supplier['latitude'] != null &&
                                  supplier['longitude'] != null,
                            )
                            .map((supplier) {
                              final lat = (supplier['latitude'] as num)
                                  .toDouble();
                              final lng = (supplier['longitude'] as num)
                                  .toDouble();
                              // 将GCJ-02坐标转换为WGS84坐标
                              final wgs84Point =
                                  CoordinateTransform.gcj02ToWgs84(lat, lng);
                              return wgs84Point;
                            }),
                      ],
                      strokeWidth: 4,
                      color: const Color(0xFF20CB6B).withOpacity(0.7),
                    ),
                  ],
                ),
              // 配送订单位置标记（已完成取货时显示）- 按照后台排序顺序显示
              if (_hasCompletedAllPickup &&
                  _deliveringOrders.isNotEmpty &&
                  _userPosition != null)
                MarkerLayer(
                  markers: _deliveringOrders
                      .where(
                        (order) =>
                            order['latitude'] != null &&
                            order['longitude'] != null,
                      )
                      .map((order) {
                        final lat = (order['latitude'] as num?)?.toDouble();
                        final lng = (order['longitude'] as num?)?.toDouble();
                        if (lat == null || lng == null) return null;

                        // 将GCJ-02坐标转换为WGS84坐标（天地图使用WGS84）
                        final wgs84Point = CoordinateTransform.gcj02ToWgs84(
                          lat,
                          lng,
                        );

                        // 获取排序序号（从 route_sequence 字段）
                        final routeSequence =
                            (order['route_sequence'] as num?)?.toInt() ?? 0;

                        // 获取订单状态，判断是否为已送达
                        final status = order['status'] as String? ?? '';
                        final isDelivered =
                            status == 'delivered' || status == 'shipped';

                        // 已送达订单使用灰色，配送中订单使用绿色
                        final markerColor = isDelivered
                            ? (Colors.grey[600] ?? Colors.grey)
                            : const Color(0xFF20CB6B);

                        return Marker(
                          point: wgs84Point,
                          width: 27,
                          height: 27,
                          alignment: Alignment.center,
                          child: Container(
                            decoration: BoxDecoration(
                              color: markerColor,
                              shape: BoxShape.circle,
                              border: Border.all(
                                color: Colors.white,
                                width: 1.5,
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withOpacity(0.2),
                                  blurRadius: 3,
                                  offset: const Offset(0, 1),
                                ),
                              ],
                            ),
                            child: Center(
                              child: Text(
                                '$routeSequence',
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 11,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                          ),
                        );
                      })
                      .where((m) => m != null)
                      .cast<Marker>()
                      .toList(),
                ),
              // 供应商位置标记（待取货时显示）- 后渲染标记，显示在线条上方
              if (_hasPendingPickup && _pickupSuppliers.isNotEmpty)
                MarkerLayer(
                  markers: _pickupSuppliers
                      .where((supplier) {
                        final lat = supplier['latitude'] as num?;
                        final lng = supplier['longitude'] as num?;
                        return lat != null && lng != null;
                      })
                      .map((supplier) {
                        final lat = supplier['latitude'] as num?;
                        final lng = supplier['longitude'] as num?;
                        // 将GCJ-02坐标转换为WGS84坐标（天地图使用WGS84）
                        final wgs84Point = CoordinateTransform.gcj02ToWgs84(
                          lat!.toDouble(),
                          lng!.toDouble(),
                        );
                        // 使用后端返回的 sequence 字段，如果没有则使用索引+1
                        final sequence =
                            (supplier['sequence'] as num?)?.toInt() ??
                            (_pickupSuppliers.indexOf(supplier) + 1);
                        return Marker(
                          point: wgs84Point,
                          width: 24,
                          height: 24,
                          alignment: Alignment.center,
                          child: Container(
                            width: 24,
                            height: 24,
                            decoration: BoxDecoration(
                              color: const Color(0xFF20CB6B), // 绿色背景
                              shape: BoxShape.circle,
                              border: Border.all(
                                color: Colors.white, // 白色边框
                                width: 1.5,
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withOpacity(0.2),
                                  blurRadius: 3,
                                  offset: const Offset(0, 1),
                                ),
                              ],
                            ),
                            child: Center(
                              child: Text(
                                '$sequence',
                                style: const TextStyle(
                                  color: Colors.white, // 白色文字
                                  fontSize: 11,
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                            ),
                          ),
                        );
                      })
                      .toList(),
                ),
              // 版权信息
              RichAttributionWidget(
                attributions: [TextSourceAttribution('天地图', onTap: () {})],
              ),
            ],
          ),
          // 顶部提示信息（优化样式，展开时隐藏）
          if (!_isSheetExpanded &&
              _hasPendingPickup &&
              _pickupSuppliers.isNotEmpty)
            Positioned(
              top: 12,
              left: 16,
              right: 16,
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 12,
                ),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.orange[300]!, width: 1.5),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.orange.withOpacity(0.15),
                      blurRadius: 12,
                      offset: const Offset(0, 4),
                      spreadRadius: 0,
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(6),
                      decoration: BoxDecoration(
                        color: Colors.orange[50],
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Icon(
                        Icons.info_outline,
                        size: 20,
                        color: Colors.orange[700],
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        '请在到达第一个取货点时检查新订单，没有顺路新订单后再开始取货！',
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.orange[900],
                          fontWeight: FontWeight.w600,
                          height: 1.4,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            )
          else if (!_isSheetExpanded &&
              _hasCompletedAllPickup &&
              _deliveringOrders.isNotEmpty)
            Positioned(
              top: 12,
              left: 16,
              right: 16,
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 12,
                ),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.green[300]!, width: 1.5),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.green.withOpacity(0.15),
                      blurRadius: 12,
                      offset: const Offset(0, 4),
                      spreadRadius: 0,
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(6),
                      decoration: BoxDecoration(
                        color: Colors.green[50],
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Icon(
                        Icons.check_circle_outline,
                        size: 20,
                        color: Colors.green[700],
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        '你已完成所有订单取货，可以开始配送，配送过程中请注意人身安全！',
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.green[900],
                          fontWeight: FontWeight.w600,
                          height: 1.4,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          // 定位按钮（右上角，优化样式，避免与提示信息重叠，缩小尺寸，展开时隐藏）
          if (!_isSheetExpanded)
            Positioned(
              top:
                  (_hasPendingPickup && _pickupSuppliers.isNotEmpty) ||
                      (_hasCompletedAllPickup && _deliveringOrders.isNotEmpty)
                  ? 100 // 如果有提示信息，往下移动更多
                  : 20, // 如果没有提示信息，往下移动一些
              right: 16,
              child: Material(
                elevation: 4,
                shadowColor: Colors.black.withOpacity(0.2),
                borderRadius: BorderRadius.circular(24),
                child: InkWell(
                  onTap: () {
                    if (_userPosition != null) {
                      // 移动到用户位置
                      _mapController.move(
                        LatLng(
                          _userPosition!.latitude,
                          _userPosition!.longitude,
                        ),
                        _initialZoom,
                      );
                    } else {
                      // 重新获取位置
                      _getUserLocation();
                    }
                  },
                  borderRadius: BorderRadius.circular(24),
                  child: Container(
                    width: 48,
                    height: 48,
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(24),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.15),
                          blurRadius: 8,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Icon(
                      _userPosition != null
                          ? Icons.my_location
                          : Icons.location_searching,
                      color: const Color(0xFF20CB6B),
                      size: 22,
                    ),
                  ),
                ),
              ),
            ),
          // 底部悬浮框（可拖拽，显示配送订单信息）
          DraggableScrollableSheet(
            controller: _draggableController, // 添加控制器以监听拖拽状态
            initialChildSize: 0.4, // 初始高度为屏幕的40%（默认展开）
            minChildSize: 0.4, // 最小高度为屏幕的40%（与默认展开一致）
            maxChildSize: 0.85, // 最大高度为屏幕的85%
            snap: true, // 启用吸附效果
            snapSizes: const [0.4, 0.85], // 吸附位置：默认/收起、完全展开
            builder: (context, scrollController) {
              return Container(
                margin: const EdgeInsets.symmetric(horizontal: 0), // 减少边距，增加宽度
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
                    Expanded(
                      child: _hasPendingPickup
                          ? _buildPickupSuppliersList(scrollController)
                          : _buildDeliveringOrdersList(scrollController),
                    ),
                  ],
                ),
              );
            },
          ),
        ],
      ),
    );
  }

  /// 构建待取货供应商列表
  Widget _buildPickupSuppliersList(ScrollController scrollController) {
    if (_isLoadingSuppliers) {
      return const Padding(
        padding: EdgeInsets.all(20),
        child: Center(
          child: CircularProgressIndicator(
            valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
          ),
        ),
      );
    }

    if (_pickupSuppliers.isEmpty) {
      return Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.inbox_outlined, size: 48, color: Colors.grey[400]),
            const SizedBox(height: 12),
            Text(
              '暂无待取货订单',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: Colors.grey[600],
              ),
            ),
          ],
        ),
      );
    }

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 标题（优化样式）
        Padding(
          padding: const EdgeInsets.fromLTRB(18, 18, 18, 14), // 减少顶部间距
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: const Color(0xFF20CB6B).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.store_outlined,
                  size: 20,
                  color: Color(0xFF20CB6B),
                ),
              ),
              const SizedBox(width: 10),
              const Text(
                '取货路线规划',
                style: TextStyle(
                  fontSize: 17,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF20253A),
                  letterSpacing: 0.2,
                ),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 4,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFF20CB6B).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  '共${_pickupSuppliers.length}个',
                  style: const TextStyle(
                    fontSize: 13,
                    color: Color(0xFF20CB6B),
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
        ),
        // 供应商列表
        Expanded(
          child: ListView.builder(
            controller: scrollController,
            // 不设置 physics，让 DraggableScrollableSheet 处理滚动和拖拽
            padding: const EdgeInsets.fromLTRB(18, 0, 18, 18),
            itemCount: _pickupSuppliers.length,
            itemBuilder: (context, index) {
              final supplier = _pickupSuppliers[index];
              // 使用后端返回的 sequence 字段，如果没有则使用索引+1
              final sequence =
                  (supplier['sequence'] as num?)?.toInt() ?? (index + 1);
              return _buildSupplierListItem(supplier, sequence);
            },
          ),
        ),
      ],
    );
  }

  /// 构建供应商列表项
  Widget _buildSupplierListItem(Map<String, dynamic> supplier, int index) {
    final name = supplier['name'] as String? ?? '取货点';
    final address = supplier['address'] as String? ?? '';
    final latitude = (supplier['latitude'] as num?)?.toDouble();
    final longitude = (supplier['longitude'] as num?)?.toDouble();
    final distance = supplier['distance'] as double?;

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 6,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () async {
            // 跳转到批量取货商品列表页面
            final supplierId = supplier['id'] as int?;
            if (supplierId != null) {
              final result = await Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (context) => BatchPickupItemsView(
                    supplierId: supplierId,
                    supplierName: name,
                    supplierLatitude: latitude,
                    supplierLongitude: longitude,
                  ),
                ),
              );
              // 如果取货成功，重新检查状态（可能已经完成全部取货）
              if (result == true && mounted) {
                await _checkAndLoadPickupSuppliers();
              }
            }
          },
          borderRadius: BorderRadius.circular(14),
          child: Padding(
            padding: const EdgeInsets.all(14),
            child: Row(
              children: [
                // 序号
                Container(
                  width: 36,
                  height: 36,
                  decoration: BoxDecoration(
                    color: const Color(0xFF20CB6B),
                    shape: BoxShape.circle,
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF20CB6B).withOpacity(0.3),
                        blurRadius: 4,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Center(
                    child: Text(
                      '$index',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 16,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 14),
                // 供应商信息
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        name,
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF20253A),
                          height: 1.3,
                        ),
                      ),
                      if (address.isNotEmpty) ...[
                        const SizedBox(height: 6),
                        Text(
                          address,
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.grey[600],
                            height: 1.4,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                      if (distance != null && _userPosition != null) ...[
                        const SizedBox(height: 6),
                        Row(
                          children: [
                            Icon(
                              Icons.location_on,
                              size: 16,
                              color: Colors.green[600],
                            ),
                            const SizedBox(width: 4),
                            Text(
                              distance < 1000
                                  ? '距离: ${distance.toStringAsFixed(0)}m'
                                  : '距离: ${(distance / 1000).toStringAsFixed(2)}km',
                              style: TextStyle(
                                fontSize: 13,
                                color: Colors.green[700],
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ],
                        ),
                      ] else if (_userPosition == null) ...[
                        const SizedBox(height: 6),
                        Row(
                          children: [
                            Icon(
                              Icons.location_searching,
                              size: 16,
                              color: Colors.grey[400],
                            ),
                            const SizedBox(width: 4),
                            Text(
                              '距离: 定位中...',
                              style: TextStyle(
                                fontSize: 13,
                                color: Colors.grey[400],
                              ),
                            ),
                          ],
                        ),
                      ],
                    ],
                  ),
                ),
                // 导航按钮
                if (latitude != null && longitude != null) ...[
                  const SizedBox(width: 8),
                  Material(
                    color: Colors.transparent,
                    child: InkWell(
                      onTap: () =>
                          _navigateToSupplier(latitude, longitude, name),
                      borderRadius: BorderRadius.circular(10),
                      child: Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: const Color(0xFF20CB6B).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: const Icon(
                          Icons.navigation,
                          size: 22,
                          color: Color(0xFF20CB6B),
                        ),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }

  /// 构建配送中订单列表
  Widget _buildDeliveringOrdersList(ScrollController scrollController) {
    if (_isLoadingOrders) {
      return const Padding(
        padding: EdgeInsets.all(20),
        child: Center(
          child: CircularProgressIndicator(
            valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
          ),
        ),
      );
    }

    if (_deliveringOrders.isEmpty) {
      return SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 30),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.center,
            mainAxisSize: MainAxisSize.min,
            children: [
              // 图标
              Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(
                  color: Colors.grey[100],
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  Icons.inbox_outlined,
                  size: 48,
                  color: Colors.grey[400],
                ),
              ),
              const SizedBox(height: 20),
              // 文字
              Text(
                '暂无配送订单',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w600,
                  color: Colors.grey[700],
                ),
              ),
              const SizedBox(height: 6),
              Text(
                '所有订单已完成配送',
                style: TextStyle(fontSize: 14, color: Colors.grey[500]),
              ),
              const SizedBox(height: 24),
              // 去接单按钮
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () {
                    // 导航回到首页
                    Navigator.of(context).popUntil((route) => route.isFirst);
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF20CB6B),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    elevation: 0,
                  ),
                  child: const Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.add_task, size: 20),
                      SizedBox(width: 8),
                      Text(
                        '去接单',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                        ),
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

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 标题（优化样式）
        Padding(
          padding: const EdgeInsets.fromLTRB(18, 18, 18, 14), // 减少顶部间距
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: const Color(0xFF20CB6B).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.local_shipping_outlined,
                  size: 20,
                  color: Color(0xFF20CB6B),
                ),
              ),
              const SizedBox(width: 10),
              const Text(
                '推荐配送顺序',
                style: TextStyle(
                  fontSize: 17,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF20253A),
                  letterSpacing: 0.2,
                ),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 4,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFF20CB6B).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  '共${_deliveringOrders.length}个',
                  style: const TextStyle(
                    fontSize: 13,
                    color: Color(0xFF20CB6B),
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
        ),
        // 列表内容
        Expanded(
          child: ListView.builder(
            controller: scrollController,
            // 不设置 physics，让 DraggableScrollableSheet 处理滚动和拖拽
            padding: const EdgeInsets.fromLTRB(18, 0, 18, 18),
            itemCount: _deliveringOrders.length,
            itemBuilder: (context, index) {
              final order = _deliveringOrders[index];
              return _buildOrderListItem(order);
            },
          ),
        ),
      ],
    );
  }

  /// 构建订单列表项（显示排序序号）
  Widget _buildOrderListItem(Map<String, dynamic> order) {
    final orderId = (order['id'] as num?)?.toInt();
    final receiverName = order['name'] as String? ?? ''; // 使用 name 字段（地址名称）
    final receiverAddress = order['address'] as String? ?? ''; // 使用 address 字段
    final contact = order['contact'] as String? ?? ''; // 联系人
    final phone = order['phone'] as String? ?? ''; // 联系电话
    final itemCount = (order['item_count'] as num?)?.toInt() ?? 0; // 商品件数
    final isUrgent = (order['is_urgent'] as bool?) ?? false;
    final routeSequence = (order['route_sequence'] as num?)?.toInt(); // 排序序号
    final latitude = (order['latitude'] as num?)?.toDouble();
    final longitude = (order['longitude'] as num?)?.toDouble();

    // 获取订单状态，判断是否为已送达
    final status = order['status'] as String? ?? '';
    final isDelivered = status == 'delivered' || status == 'shipped';

    // 已送达订单使用灰色，配送中订单使用绿色
    final sequenceColor = isDelivered
        ? (Colors.grey[600] ?? Colors.grey)
        : const Color(0xFF20CB6B);

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: isUrgent
              ? Colors.orange[300]!
              : isDelivered
              ? Colors.grey[300]!
              : const Color(0xFFE5E7EB),
          width: isUrgent ? 1.5 : 1,
        ),
        boxShadow: [
          BoxShadow(
            color: isDelivered
                ? Colors.grey.withOpacity(0.04)
                : Colors.black.withOpacity(0.04),
            blurRadius: 6,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () async {
            if (orderId != null) {
              // 从订单详情返回时，刷新路线规划页面
              await Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (context) => OrderDetailView(orderId: orderId),
                ),
              );
              // 无论返回什么，都刷新路线规划页面
              if (mounted) {
                _checkAndLoadPickupSuppliers();
              }
            }
          },
          borderRadius: BorderRadius.circular(14),
          child: Padding(
            padding: const EdgeInsets.all(14),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // 左侧内容
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          // 排序序号
                          if (routeSequence != null)
                            Container(
                              width: 32,
                              height: 32,
                              decoration: BoxDecoration(
                                color: sequenceColor,
                                shape: BoxShape.circle,
                                border: Border.all(
                                  color: Colors.white,
                                  width: 2,
                                ),
                                boxShadow: [
                                  BoxShadow(
                                    color: sequenceColor.withOpacity(0.3),
                                    blurRadius: 4,
                                    offset: const Offset(0, 2),
                                  ),
                                ],
                              ),
                              child: Center(
                                child: Text(
                                  '$routeSequence',
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 15,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ),
                          if (routeSequence != null) const SizedBox(width: 10),
                          // 收货人名称
                          Expanded(
                            child: Text(
                              receiverName.isNotEmpty ? receiverName : '收货地址',
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.w600,
                                color: isDelivered
                                    ? Colors.grey[600]
                                    : const Color(0xFF20253A),
                                height: 1.3,
                              ),
                            ),
                          ),
                          // 已送达标签
                          if (isDelivered)
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.grey[100],
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Text(
                                '已送达',
                                style: TextStyle(
                                  fontSize: 11,
                                  color: Colors.grey[700],
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          // 加急标签
                          if (isUrgent && !isDelivered)
                            Container(
                              margin: const EdgeInsets.only(left: 6),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.orange[50],
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Text(
                                '加急',
                                style: TextStyle(
                                  fontSize: 11,
                                  color: Colors.orange[800],
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                        ],
                      ),
                      const SizedBox(height: 10),
                      // 联系人信息
                      if (contact.isNotEmpty || phone.isNotEmpty) ...[
                        Row(
                          children: [
                            Icon(
                              Icons.person_outline,
                              size: 16,
                              color: Colors.grey[500],
                            ),
                            const SizedBox(width: 6),
                            Text(
                              contact.isNotEmpty ? contact : phone,
                              style: TextStyle(
                                fontSize: 14,
                                color: Colors.grey[700],
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                            if (contact.isNotEmpty && phone.isNotEmpty) ...[
                              const SizedBox(width: 10),
                              Text(
                                phone,
                                style: TextStyle(
                                  fontSize: 14,
                                  color: Colors.grey[500],
                                ),
                              ),
                            ],
                          ],
                        ),
                        const SizedBox(height: 8),
                      ],
                      // 收货地址
                      if (receiverAddress.isNotEmpty)
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Icon(
                              Icons.location_on_outlined,
                              size: 16,
                              color: Colors.grey[500],
                            ),
                            const SizedBox(width: 6),
                            Expanded(
                              child: Text(
                                receiverAddress,
                                style: TextStyle(
                                  fontSize: 13,
                                  color: Colors.grey[600],
                                  height: 1.4,
                                ),
                                maxLines: 2,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          ],
                        ),
                      const SizedBox(height: 8),
                      // 商品件数
                      Row(
                        children: [
                          Icon(
                            Icons.shopping_cart_outlined,
                            size: 16,
                            color: Colors.grey[500],
                          ),
                          const SizedBox(width: 6),
                          Text(
                            '商品件数: $itemCount',
                            style: TextStyle(
                              fontSize: 13,
                              color: Colors.grey[600],
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                // 右侧导航按钮（垂直居中）
                if (latitude != null && longitude != null) ...[
                  const SizedBox(width: 10),
                  Material(
                    color: Colors.transparent,
                    child: InkWell(
                      onTap: () {
                        // 阻止事件冒泡，避免触发订单详情跳转
                        _navigateToOrder(
                          latitude,
                          longitude,
                          receiverName.isNotEmpty
                              ? receiverName
                              : receiverAddress,
                        );
                      },
                      borderRadius: BorderRadius.circular(10),
                      child: Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: const Color(0xFF20CB6B).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: const Icon(
                          Icons.navigation,
                          size: 22,
                          color: Color(0xFF20CB6B),
                        ),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }

  /// 导航到订单地址
  Future<void> _navigateToOrder(
    double latitude,
    double longitude,
    String name,
  ) async {
    try {
      // 优先使用高德地图
      final isAmapAvailable = await MapLauncher.isMapAvailable(MapType.amap);
      if (isAmapAvailable == true) {
        await MapLauncher.showDirections(
          mapType: MapType.amap,
          destination: Coords(latitude, longitude),
          destinationTitle: name,
        );
        return;
      }

      // 如果没有高德地图，使用其他可用地图
      final availableMaps = await MapLauncher.installedMaps;
      if (availableMaps.isNotEmpty) {
        await MapLauncher.showDirections(
          mapType: availableMaps.first.mapType,
          destination: Coords(latitude, longitude),
          destinationTitle: name,
        );
      } else {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('未安装地图应用，请先安装高德地图或其他地图应用'),
            backgroundColor: Colors.orange,
          ),
        );
      }
    } catch (e) {
      print('[RoutePlanningView] 导航失败: $e');
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('导航失败: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }
}
