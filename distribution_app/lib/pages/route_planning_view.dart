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

class _RoutePlanningViewState extends State<RoutePlanningView> {
  // 地图控制器
  final MapController _mapController = MapController();

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

  // 路线规划相关
  // ignore: unused_field
  Map<String, dynamic>? _routePlanResult; // 路线规划结果（保留用于调试）
  bool _isPlanningRoute = false; // 是否正在规划路线
  List<LatLng>? _routePolyline; // 解析后的路线坐标点
  List<Map<String, dynamic>> _routeWaypoints = []; // 途经点列表（按顺序）
  Map<String, dynamic>? _routeDestination; // 目的地

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
    // 页面加载时获取用户位置并开始监听位置更新
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _startLocationTracking();
      _checkAndLoadPickupSuppliers();
    });
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
          // 如果已完成全部取货且有用户位置，自动规划路线
          if (hasOrders && _userPosition != null) {
            print('[RoutePlanningView] 满足路线规划条件，开始规划');
            _planDeliveryRoute();
          } else {
            print(
              '[RoutePlanningView] 不满足路线规划条件 - 有订单: $hasOrders, 有位置: ${_userPosition != null}',
            );
          }
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
        // 如果已完成全部取货且有用户位置，自动规划路线
        if (hasOrders && _userPosition != null) {
          print('[RoutePlanningView] 满足路线规划条件（失败分支），开始规划');
          _planDeliveryRoute();
        } else {
          print(
            '[RoutePlanningView] 不满足路线规划条件（失败分支） - 有订单: $hasOrders, 有位置: ${_userPosition != null}',
          );
        }
      }
    }
  }

  /// 加载待取货供应商列表
  Future<void> _loadPickupSuppliers() async {
    if (_isLoadingSuppliers) return;

    setState(() {
      _isLoadingSuppliers = true;
    });

    final response = await OrderApi.getPickupSuppliers();

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      List<Map<String, dynamic>> suppliers = (response.data as List<dynamic>)
          .cast<Map<String, dynamic>>();

      // 如果有用户位置，计算距离并排序
      if (_userPosition != null) {
        suppliers = _sortSuppliersByDistance(suppliers);
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

  /// 加载配送中的订单
  Future<void> _loadDeliveringOrders() async {
    if (_isLoadingOrders) return;

    setState(() {
      _isLoadingOrders = true;
    });

    final response = await OrderApi.getOrderPool(
      pageNum: 1,
      pageSize: 100, // 获取所有配送中的订单
      status: 'delivering',
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final List<dynamic> list = (data['list'] as List<dynamic>? ?? []);
      setState(() {
        _deliveringOrders = list.cast<Map<String, dynamic>>();
        _isLoadingOrders = false;
      });
    } else {
      setState(() {
        _isLoadingOrders = false;
      });
    }
  }

  /// 规划配送路线
  Future<void> _planDeliveryRoute() async {
    if (_userPosition == null || _isPlanningRoute) {
      print(
        '[RoutePlanningView] 跳过路线规划 - 用户位置: ${_userPosition != null}, 正在规划: $_isPlanningRoute',
      );
      return;
    }

    print(
      '[RoutePlanningView] 开始规划路线 - 起点: ${_userPosition!.latitude}, ${_userPosition!.longitude}',
    );
    print('[RoutePlanningView] 配送中订单数量: ${_deliveringOrders.length}');

    setState(() {
      _isPlanningRoute = true;
    });

    try {
      final response = await OrderApi.planDeliveryRoute(
        originLatitude: _userPosition!.latitude,
        originLongitude: _userPosition!.longitude,
      );

      if (!mounted) return;

      print(
        '[RoutePlanningView] 路线规划API响应 - 成功: ${response.isSuccess}, 消息: ${response.message}',
      );

      if (response.isSuccess && response.data != null) {
        final data = response.data!;
        final route = data['route'] as Map<String, dynamic>?;
        final waypoints = data['waypoints'] as List<dynamic>?;
        final destination = data['destination'] as Map<String, dynamic>?;

        print(
          '[RoutePlanningView] 路线规划数据 - route: ${route != null}, waypoints: ${waypoints != null ? waypoints.length : 0}, destination: ${destination != null}',
        );

        // 解析polyline坐标点串
        List<LatLng>? polylinePoints;
        if (route != null && route['polyline'] != null) {
          final polylineStr = route['polyline'] as String;
          print('[RoutePlanningView] 开始解析polyline，长度: ${polylineStr.length}');
          polylinePoints = _parsePolyline(polylineStr);
          print(
            '[RoutePlanningView] polyline解析完成，坐标点数: ${polylinePoints.length}',
          );
        } else {
          print('[RoutePlanningView] 警告: route或polyline为空');
        }

        // 按index排序途经点
        List<Map<String, dynamic>> sortedWaypoints = [];
        if (waypoints != null) {
          print('[RoutePlanningView] 途经点数量: ${waypoints.length}');
          sortedWaypoints = waypoints.cast<Map<String, dynamic>>().toList()
            ..sort((a, b) {
              final indexA = (a['index'] as num?)?.toInt() ?? 0;
              final indexB = (b['index'] as num?)?.toInt() ?? 0;
              return indexA.compareTo(indexB);
            });
          print('[RoutePlanningView] 途经点排序完成，数量: ${sortedWaypoints.length}');
        } else {
          print('[RoutePlanningView] 警告: waypoints为空');
        }

        setState(() {
          _routePlanResult = data;
          _routePolyline = polylinePoints;
          _routeWaypoints = sortedWaypoints;
          _routeDestination = destination;
          _isPlanningRoute = false;
        });

        // 更新地图显示
        if (polylinePoints != null && polylinePoints.isNotEmpty) {
          print('[RoutePlanningView] 路线解析成功，坐标点数: ${polylinePoints.length}');
          print(
            '[RoutePlanningView] 途经点数量: ${sortedWaypoints.length}, 目的地: ${destination != null}',
          );
          if (polylinePoints.isNotEmpty) {
            print(
              '[RoutePlanningView] 路线起点: ${polylinePoints.first.latitude}, ${polylinePoints.first.longitude}',
            );
            print(
              '[RoutePlanningView] 路线终点: ${polylinePoints.last.latitude}, ${polylinePoints.last.longitude}',
            );
            // 检查路线是否经过途经点
            if (sortedWaypoints.isNotEmpty) {
              for (final wp in sortedWaypoints) {
                final wpLat = (wp['latitude'] as num?)?.toDouble();
                final wpLng = (wp['longitude'] as num?)?.toDouble();
                if (wpLat != null && wpLng != null) {
                  // 将途经点坐标转换为WGS84
                  final wpWgs84 = CoordinateTransform.gcj02ToWgs84(
                    wpLat,
                    wpLng,
                  );
                  // 检查路线中是否有接近途经点的坐标（允许误差0.0001度，约11米）
                  bool foundNearby = false;
                  for (final point in polylinePoints) {
                    final latDiff = (point.latitude - wpWgs84.latitude).abs();
                    final lngDiff = (point.longitude - wpWgs84.longitude).abs();
                    if (latDiff < 0.0001 && lngDiff < 0.0001) {
                      foundNearby = true;
                      break;
                    }
                  }
                  if (foundNearby) {
                    print(
                      '[RoutePlanningView] 途经点[${wp['index']}] (${wpLat}, ${wpLng}) 在路线中找到',
                    );
                  } else {
                    print(
                      '[RoutePlanningView] 警告: 途经点[${wp['index']}] (${wpLat}, ${wpLng}) 在路线中未找到',
                    );
                  }
                }
              }
            }
          }
          _updateMapForRoute();
        } else {
          print('[RoutePlanningView] 警告: polyline为空，无法显示路线');
        }

        print(
          '[RoutePlanningView] 路线规划成功 - 途经点: ${sortedWaypoints.length}, 目的地: ${destination != null}, 路线点数: ${polylinePoints?.length ?? 0}',
        );
      } else {
        print('[RoutePlanningView] 路线规划失败 - 消息: ${response.message}');
        if (mounted) {
          setState(() {
            _isPlanningRoute = false;
          });
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                response.message.isNotEmpty ? response.message : '路线规划失败',
              ),
              backgroundColor: Colors.orange,
            ),
          );
        }
      }
    } catch (e, stackTrace) {
      print('[RoutePlanningView] 路线规划异常: $e');
      print('[RoutePlanningView] 堆栈跟踪: $stackTrace');
      if (mounted) {
        setState(() {
          _isPlanningRoute = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('路线规划失败: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  /// 解析polyline坐标点串（高德地图格式：经度,纬度;经度,纬度;...）
  List<LatLng> _parsePolyline(String polyline) {
    final points = <LatLng>[];
    final segments = polyline.split(';');
    print('[RoutePlanningView] polyline分段数: ${segments.length}');

    int successCount = 0;
    int failCount = 0;

    for (int i = 0; i < segments.length; i++) {
      final segment = segments[i];
      final coords = segment.split(',');
      if (coords.length == 2) {
        try {
          final lng = double.parse(coords[0].trim());
          final lat = double.parse(coords[1].trim());
          // 将GCJ-02坐标转换为WGS84坐标（天地图使用WGS84）
          final wgs84Point = CoordinateTransform.gcj02ToWgs84(lat, lng);
          points.add(wgs84Point);
          successCount++;
        } catch (e) {
          failCount++;
          if (i < 5 || i >= segments.length - 5) {
            // 只打印前5个和后5个错误，避免日志过多
            print('[RoutePlanningView] 解析坐标点失败 [索引$i]: $segment, 错误: $e');
          }
        }
      } else {
        failCount++;
        if (i < 5 || i >= segments.length - 5) {
          print(
            '[RoutePlanningView] 坐标格式错误 [索引$i]: $segment (期望2个值，实际${coords.length}个)',
          );
        }
      }
    }

    print(
      '[RoutePlanningView] polyline解析结果 - 成功: $successCount, 失败: $failCount, 总计: ${points.length}',
    );
    return points;
  }

  /// 更新地图显示路线
  void _updateMapForRoute() {
    if (_routePolyline == null || _routePolyline!.isEmpty) return;

    // 计算路线边界
    double minLat = _routePolyline!.first.latitude;
    double maxLat = _routePolyline!.first.latitude;
    double minLng = _routePolyline!.first.longitude;
    double maxLng = _routePolyline!.first.longitude;

    for (final point in _routePolyline!) {
      if (point.latitude < minLat) minLat = point.latitude;
      if (point.latitude > maxLat) maxLat = point.latitude;
      if (point.longitude < minLng) minLng = point.longitude;
      if (point.longitude > maxLng) maxLng = point.longitude;
    }

    // 添加用户位置到边界计算
    if (_userPosition != null) {
      if (_userPosition!.latitude < minLat) minLat = _userPosition!.latitude;
      if (_userPosition!.latitude > maxLat) maxLat = _userPosition!.latitude;
      if (_userPosition!.longitude < minLng) minLng = _userPosition!.longitude;
      if (_userPosition!.longitude > maxLng) maxLng = _userPosition!.longitude;
    }

    // 计算中心点和缩放级别
    final centerLat = (minLat + maxLat) / 2;
    final centerLng = (minLng + maxLng) / 2;
    final latDiff = maxLat - minLat;
    final lngDiff = maxLng - minLng;
    final maxDiff = latDiff > lngDiff ? latDiff : lngDiff;

    double zoom = 15.0;
    if (maxDiff > 0.1) {
      zoom = 12.0;
    } else if (maxDiff > 0.05) {
      zoom = 13.0;
    } else if (maxDiff > 0.02) {
      zoom = 14.0;
    }

    // 移动地图到路线中心
    _mapController.move(LatLng(centerLat, centerLng), zoom);
  }

  /// 导航到客户位置
  Future<void> _navigateToCustomer(
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
          SnackBar(content: Text('导航失败: $e'), backgroundColor: Colors.red),
        );
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
    _positionStreamSubscription?.cancel();
    _locationStreamController.close();
    _mapController.dispose();
    super.dispose();
  }

  /// 显示权限说明对话框（针对小米手机等可能不弹出系统对话框的情况）
  Future<bool> _showPermissionDialog() async {
    return await showDialog<bool>(
          context: context,
          builder: (BuildContext context) {
            return AlertDialog(
              title: const Text('需要定位权限'),
              content: const Text(
                '为了提供路线规划服务，需要获取您的位置信息。\n\n'
                '如果您使用的是小米手机，系统可能不会自动弹出权限请求，请点击"去设置"手动开启定位权限。',
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: const Text('取消'),
                ),
                TextButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  child: const Text('去设置'),
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
          // 如果已完成全部取货，自动规划路线
          if (wasFirstLocation &&
              _hasCompletedAllPickup &&
              _deliveringOrders.isNotEmpty) {
            print('[RoutePlanningView] 网络定位成功，触发路线规划');
            _planDeliveryRoute();
          }
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
        // 如果已完成全部取货，自动规划路线
        if (wasFirstLocation &&
            _hasCompletedAllPickup &&
            _deliveringOrders.isNotEmpty) {
          print('[RoutePlanningView] 初始位置获取成功，触发路线规划');
          _planDeliveryRoute();
        }
        _mapController.move(
          LatLng(initialPosition.latitude, initialPosition.longitude),
          _initialZoom,
        );
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

              // 首次定位时，将地图中心移动到用户位置
              if (wasFirstLocation) {
                _mapController.move(
                  LatLng(position.latitude, position.longitude),
                  _initialZoom,
                );
                // 如果有待取货供应商，更新地图显示
                if (_hasPendingPickup && _pickupSuppliers.isNotEmpty) {
                  _updateMapForSuppliers();
                }
                // 如果已完成全部取货，自动规划路线
                if (_hasCompletedAllPickup && _deliveringOrders.isNotEmpty) {
                  print('[RoutePlanningView] 首次定位成功（位置流），触发路线规划');
                  _planDeliveryRoute();
                } else {
                  print(
                    '[RoutePlanningView] 首次定位成功（位置流），但不满足路线规划条件 - 已完成取货: $_hasCompletedAllPickup, 配送中订单: ${_deliveringOrders.length}',
                  );
                }
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
                // 如果已完成全部取货，重新规划路线
                if (_hasCompletedAllPickup && _deliveringOrders.isNotEmpty) {
                  print('[RoutePlanningView] 位置更新（位置流），重新规划路线');
                  _planDeliveryRoute();
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
            // 如果已完成全部取货，自动规划路线
            if (wasFirstLocation &&
                _hasCompletedAllPickup &&
                _deliveringOrders.isNotEmpty) {
              print('[RoutePlanningView] 网络定位成功（_getUserLocation），触发路线规划');
              _planDeliveryRoute();
            }
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
          // 如果已完成全部取货，自动规划路线
          if (wasFirstLocation &&
              _hasCompletedAllPickup &&
              _deliveringOrders.isNotEmpty) {
            print('[RoutePlanningView] 获取位置成功（_getUserLocation），触发路线规划');
            _planDeliveryRoute();
          }
        }
      });
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
              // 配送路线（已完成取货时显示）- 先渲染线条，避免遮挡标记
              if (_hasCompletedAllPickup &&
                  _routePolyline != null &&
                  _routePolyline!.isNotEmpty)
                PolylineLayer(
                  polylines: [
                    // 白色边框（底层）
                    Polyline(
                      points: _routePolyline!,
                      strokeWidth: 8,
                      color: Colors.white,
                    ),
                    // 绿色路线（上层）
                    Polyline(
                      points: _routePolyline!,
                      strokeWidth: 5,
                      color: const Color(0xFF20CB6B),
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
              // 客户位置标记（已完成取货时显示）- 后渲染标记，显示在线条上方
              if (_hasCompletedAllPickup &&
                  (_routeWaypoints.isNotEmpty || _routeDestination != null))
                MarkerLayer(
                  markers: [
                    // 途经点标记
                    ..._routeWaypoints
                        .map((waypoint) {
                          final lat = (waypoint['latitude'] as num?)
                              ?.toDouble();
                          final lng = (waypoint['longitude'] as num?)
                              ?.toDouble();
                          if (lat == null || lng == null) return null;
                          final wgs84Point = CoordinateTransform.gcj02ToWgs84(
                            lat,
                            lng,
                          );
                          final index =
                              (waypoint['index'] as num?)?.toInt() ?? 0;
                          return Marker(
                            point: wgs84Point,
                            width: 40,
                            height: 40,
                            alignment: Alignment.center,
                            child: Container(
                              decoration: BoxDecoration(
                                color: const Color(0xFF20CB6B),
                                shape: BoxShape.circle,
                                border: Border.all(
                                  color: Colors.white,
                                  width: 2,
                                ),
                                boxShadow: [
                                  BoxShadow(
                                    color: Colors.black.withOpacity(0.2),
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
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ),
                          );
                        })
                        .where((m) => m != null)
                        .cast<Marker>(),
                    // 目的地标记
                    if (_routeDestination != null)
                      ...[
                        () {
                          final lat = (_routeDestination!['latitude'] as num?)
                              ?.toDouble();
                          final lng = (_routeDestination!['longitude'] as num?)
                              ?.toDouble();
                          if (lat == null || lng == null) return null;
                          final wgs84Point = CoordinateTransform.gcj02ToWgs84(
                            lat,
                            lng,
                          );
                          final waypointCount = _routeWaypoints.length;
                          return Marker(
                            point: wgs84Point,
                            width: 40,
                            height: 40,
                            alignment: Alignment.center,
                            child: Container(
                              decoration: BoxDecoration(
                                color: Colors.red,
                                shape: BoxShape.circle,
                                border: Border.all(
                                  color: Colors.white,
                                  width: 2,
                                ),
                                boxShadow: [
                                  BoxShadow(
                                    color: Colors.black.withOpacity(0.2),
                                    blurRadius: 4,
                                    offset: const Offset(0, 2),
                                  ),
                                ],
                              ),
                              child: Center(
                                child: Text(
                                  '${waypointCount + 1}',
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 16,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ),
                          );
                        }(),
                      ].where((m) => m != null).cast<Marker>(),
                  ],
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
                        final index = _pickupSuppliers.indexOf(supplier) + 1;
                        return Marker(
                          point: wgs84Point,
                          width: 36,
                          height: 36,
                          alignment: Alignment.center,
                          child: Container(
                            width: 36,
                            height: 36,
                            decoration: BoxDecoration(
                              color: const Color(0xFF20CB6B), // 绿色背景
                              shape: BoxShape.circle,
                              border: Border.all(
                                color: Colors.white, // 白色边框
                                width: 2,
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withOpacity(0.2),
                                  blurRadius: 4,
                                  offset: const Offset(0, 2),
                                ),
                              ],
                            ),
                            child: Center(
                              child: Text(
                                '$index',
                                style: const TextStyle(
                                  color: Colors.white, // 白色文字
                                  fontSize: 16,
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
          // 顶部提示信息
          if (_hasPendingPickup && _pickupSuppliers.isNotEmpty)
            Positioned(
              top: 16,
              left: 16,
              right: 16,
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  color: Colors.orange[50],
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.orange[200]!, width: 1),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.1),
                      blurRadius: 8,
                      offset: const Offset(0, 2),
                    ),
                  ],
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
                        '请在到达第一个取货点时检查新订单，没有顺路新订单后再开始取货！！！',
                        style: TextStyle(
                          fontSize: 13,
                          color: Colors.orange[900],
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            )
          else if (_hasCompletedAllPickup && _deliveringOrders.isNotEmpty)
            Positioned(
              top: 16,
              left: 16,
              right: 16,
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  color: Colors.green[50],
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.green[200]!, width: 1),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.1),
                      blurRadius: 8,
                      offset: const Offset(0, 2),
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    Icon(
                      Icons.check_circle_outline,
                      size: 18,
                      color: Colors.green[700],
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        '你已完成全部取货，配送过程中请注意行车安全！',
                        style: TextStyle(
                          fontSize: 13,
                          color: Colors.green[900],
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          // 定位按钮（右下角，位于悬浮框上方）
          Positioned(
            bottom: 240, // 调整位置，避免被悬浮框遮挡
            right: 16,
            child: FloatingActionButton(
              mini: true,
              backgroundColor: Colors.white,
              onPressed: () {
                if (_userPosition != null) {
                  // 移动到用户位置
                  _mapController.move(
                    LatLng(_userPosition!.latitude, _userPosition!.longitude),
                    _initialZoom,
                  );
                } else {
                  // 重新获取位置
                  _getUserLocation();
                }
              },
              child: Icon(
                _userPosition != null
                    ? Icons.my_location
                    : Icons.location_searching,
                color: const Color(0xFF20CB6B),
              ),
            ),
          ),
          // 底部悬浮框（显示配送订单信息）
          Positioned(
            left: 16,
            right: 16,
            bottom: 16,
            child: _buildBottomFloatingBox(),
          ),
        ],
      ),
    );
  }

  /// 构建底部悬浮框
  Widget _buildBottomFloatingBox() {
    return Container(
      constraints: const BoxConstraints(maxHeight: 300),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 10,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: _hasPendingPickup
          ? _buildPickupSuppliersList()
          : _buildDeliveringOrdersList(),
    );
  }

  /// 构建待取货供应商列表
  Widget _buildPickupSuppliersList() {
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
        // 标题
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 12),
          child: Row(
            children: [
              const Icon(
                Icons.store_outlined,
                size: 20,
                color: Color(0xFF20CB6B),
              ),
              const SizedBox(width: 8),
              const Text(
                '取货路线规划',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              Text(
                '共${_pickupSuppliers.length}个取货点',
                style: TextStyle(fontSize: 12, color: Colors.grey[600]),
              ),
            ],
          ),
        ),
        // 供应商列表
        Flexible(
          child: ListView.builder(
            shrinkWrap: true,
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
            itemCount: _pickupSuppliers.length,
            itemBuilder: (context, index) {
              final supplier = _pickupSuppliers[index];
              return _buildSupplierListItem(supplier, index + 1);
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
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF8F9FA),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
      ),
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
            // 如果取货成功，刷新供应商列表
            if (result == true && mounted) {
              await _loadPickupSuppliers();
            }
          }
        },
        child: Row(
          children: [
            // 序号
            Container(
              width: 32,
              height: 32,
              decoration: BoxDecoration(
                color: const Color(0xFF20CB6B),
                shape: BoxShape.circle,
              ),
              child: Center(
                child: Text(
                  '$index',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 14,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            // 供应商信息
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
                  ),
                  if (address.isNotEmpty) ...[
                    const SizedBox(height: 4),
                    Text(
                      address,
                      style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                  if (distance != null && _userPosition != null) ...[
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        const Icon(
                          Icons.location_on,
                          size: 14,
                          color: Color(0xFF20CB6B),
                        ),
                        const SizedBox(width: 4),
                        Text(
                          distance < 1000
                              ? '距离: ${distance.toStringAsFixed(0)}m'
                              : '距离: ${(distance / 1000).toStringAsFixed(2)}km',
                          style: const TextStyle(
                            fontSize: 13,
                            color: Color(0xFF20CB6B),
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ] else if (_userPosition == null) ...[
                    const SizedBox(height: 4),
                    Text(
                      '距离: 定位中...',
                      style: TextStyle(fontSize: 12, color: Colors.grey[400]),
                    ),
                  ],
                ],
              ),
            ),
            // 导航按钮
            if (latitude != null && longitude != null)
              InkWell(
                onTap: () => _navigateToSupplier(latitude, longitude, name),
                child: Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: const Color(0xFF20CB6B).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Icon(
                    Icons.navigation,
                    size: 20,
                    color: Color(0xFF20CB6B),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  /// 构建配送中订单列表
  Widget _buildDeliveringOrdersList() {
    if (_isLoadingOrders || _isPlanningRoute) {
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
      return Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.inbox_outlined, size: 48, color: Colors.grey[400]),
            const SizedBox(height: 12),
            Text(
              '暂无配送订单',
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

    // 如果有路线规划结果，显示按路线排序的客户列表
    final hasRoutePlan =
        _routeWaypoints.isNotEmpty || _routeDestination != null;

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 标题
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 12),
          child: Row(
            children: [
              const Icon(
                Icons.local_shipping_outlined,
                size: 20,
                color: Color(0xFF20CB6B),
              ),
              const SizedBox(width: 8),
              Text(
                hasRoutePlan ? '配送客户顺序列表' : '配送订单列表',
                style: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF20253A),
                ),
              ),
              const Spacer(),
              Text(
                hasRoutePlan
                    ? '共${_routeWaypoints.length + (_routeDestination != null ? 1 : 0)}个客户'
                    : '共${_deliveringOrders.length}个订单',
                style: TextStyle(fontSize: 12, color: Colors.grey[600]),
              ),
              // 一键导航按钮（导航到第一个客户）
              if (hasRoutePlan && _routeWaypoints.isNotEmpty)
                InkWell(
                  onTap: () {
                    final firstWaypoint = _routeWaypoints[0];
                    final lat = (firstWaypoint['latitude'] as num?)?.toDouble();
                    final lng = (firstWaypoint['longitude'] as num?)
                        ?.toDouble();
                    final name = firstWaypoint['name'] as String? ?? '';
                    final address = firstWaypoint['address'] as String? ?? '';
                    if (lat != null && lng != null) {
                      _navigateToCustomer(
                        lat,
                        lng,
                        name.isNotEmpty ? name : address,
                      );
                    }
                  },
                  child: Container(
                    margin: const EdgeInsets.only(left: 12),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 6,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFF20CB6B),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.navigation, size: 16, color: Colors.white),
                        SizedBox(width: 4),
                        Text(
                          '一键导航',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.white,
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
        // 列表内容
        Flexible(
          child: hasRoutePlan
              ? ListView.builder(
                  shrinkWrap: true,
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                  itemCount:
                      _routeWaypoints.length +
                      (_routeDestination != null ? 1 : 0),
                  itemBuilder: (context, index) {
                    if (index < _routeWaypoints.length) {
                      // 途经点
                      final waypoint = _routeWaypoints[index];
                      return _buildCustomerListItem(waypoint, index + 1);
                    } else {
                      // 目的地
                      return _buildCustomerListItem(
                        _routeDestination!,
                        _routeWaypoints.length + 1,
                        isDestination: true,
                      );
                    }
                  },
                )
              : ListView.builder(
                  shrinkWrap: true,
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
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

  /// 构建客户列表项（按路线规划顺序）
  Widget _buildCustomerListItem(
    Map<String, dynamic> customer,
    int index, {
    bool isDestination = false,
  }) {
    final name = customer['name'] as String? ?? '';
    final address = customer['address'] as String? ?? '';
    final latitude = (customer['latitude'] as num?)?.toDouble();
    final longitude = (customer['longitude'] as num?)?.toDouble();

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: isDestination ? Colors.red[50] : const Color(0xFFF8F9FA),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isDestination ? Colors.red[300]! : const Color(0xFFE5E7EB),
          width: isDestination ? 2 : 1,
        ),
      ),
      child: Row(
        children: [
          // 序号标记
          Container(
            width: 32,
            height: 32,
            decoration: BoxDecoration(
              color: isDestination ? Colors.red : const Color(0xFF20CB6B),
              shape: BoxShape.circle,
              border: Border.all(color: Colors.white, width: 2),
            ),
            child: Center(
              child: Text(
                '$index',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          // 客户信息
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // 地址名称
                if (name.isNotEmpty)
                  Text(
                    name,
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF20253A),
                    ),
                  ),
                if (name.isNotEmpty && address.isNotEmpty)
                  const SizedBox(height: 4),
                // 地址
                if (address.isNotEmpty)
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Icon(
                        Icons.location_on_outlined,
                        size: 14,
                        color: Color(0xFF8C92A4),
                      ),
                      const SizedBox(width: 4),
                      Expanded(
                        child: Text(
                          address,
                          style: const TextStyle(
                            fontSize: 13,
                            color: Color(0xFF40475C),
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
              ],
            ),
          ),
          // 导航按钮
          if (latitude != null && longitude != null)
            InkWell(
              onTap: () => _navigateToCustomer(
                latitude,
                longitude,
                name.isNotEmpty ? name : address,
              ),
              child: Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFF20CB6B).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.navigation,
                  size: 20,
                  color: Color(0xFF20CB6B),
                ),
              ),
            ),
        ],
      ),
    );
  }

  /// 构建订单列表项（保留用于其他场景）
  Widget _buildOrderListItem(Map<String, dynamic> order) {
    final orderId = (order['id'] as num?)?.toInt();
    final orderNumber = order['order_number'] as String? ?? '';
    final receiverName = order['receiver_name'] as String? ?? '';
    final receiverAddress = order['receiver_address'] as String? ?? '';
    final totalAmount = (order['total_amount'] as num?)?.toDouble() ?? 0.0;
    final isUrgent = (order['is_urgent'] as bool?) ?? false;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF8F9FA),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isUrgent ? Colors.orange[300]! : const Color(0xFFE5E7EB),
          width: isUrgent ? 2 : 1,
        ),
      ),
      child: InkWell(
        onTap: () {
          if (orderId != null) {
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (context) => OrderDetailView(orderId: orderId),
              ),
            );
          }
        },
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                // 订单号
                Expanded(
                  child: Text(
                    '订单号: $orderNumber',
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF20253A),
                    ),
                  ),
                ),
                // 加急标签
                if (isUrgent)
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.orange[100],
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      '加急',
                      style: TextStyle(
                        fontSize: 10,
                        color: Colors.orange[800],
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            // 收货人信息
            if (receiverName.isNotEmpty) ...[
              Row(
                children: [
                  const Icon(
                    Icons.person_outline,
                    size: 14,
                    color: Color(0xFF8C92A4),
                  ),
                  const SizedBox(width: 4),
                  Text(
                    receiverName,
                    style: const TextStyle(
                      fontSize: 13,
                      color: Color(0xFF40475C),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 4),
            ],
            // 收货地址
            if (receiverAddress.isNotEmpty) ...[
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(
                    Icons.location_on_outlined,
                    size: 14,
                    color: Color(0xFF8C92A4),
                  ),
                  const SizedBox(width: 4),
                  Expanded(
                    child: Text(
                      receiverAddress,
                      style: const TextStyle(
                        fontSize: 13,
                        color: Color(0xFF40475C),
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
            ],
            // 订单金额
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  '订单金额: ¥${totalAmount.toStringAsFixed(2)}',
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF20CB6B),
                  ),
                ),
                const Icon(
                  Icons.chevron_right,
                  size: 18,
                  color: Color(0xFF8C92A4),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
