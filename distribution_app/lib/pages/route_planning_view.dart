import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'package:permission_handler/permission_handler.dart';
import '../utils/location_service.dart';

/// 路线规划页面：使用 flutter_map 显示天地图图层
class RoutePlanningView extends StatefulWidget {
  const RoutePlanningView({super.key});

  @override
  State<RoutePlanningView> createState() => _RoutePlanningViewState();
}

class _RoutePlanningViewState extends State<RoutePlanningView> {
  // 地图控制器
  final MapController _mapController = MapController();

  // 用户位置
  Position? _userPosition;
  bool _isLoadingLocation = false;
  String? _locationError;

  // 地图初始中心点（北京天安门）
  final LatLng _initialCenter = const LatLng(39.90864, 116.39750);
  final double _initialZoom = 15.0;

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
    // 页面加载时获取用户位置
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _getUserLocation();
    });
  }

  @override
  void dispose() {
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

  /// 获取用户位置
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
          _userPosition = position;
          _locationError = null;
          // 将地图中心移动到用户位置
          _mapController.move(
            LatLng(_userPosition!.latitude, _userPosition!.longitude),
            _initialZoom,
          );
        }
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('路线规划'),
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
              // 用户位置标记图层
              if (_userPosition != null)
                MarkerLayer(
                  markers: [
                    Marker(
                      point: LatLng(
                        _userPosition!.latitude,
                        _userPosition!.longitude,
                      ),
                      width: 40,
                      height: 40,
                      child: Container(
                        decoration: BoxDecoration(
                          color: const Color(0xFF20CB6B),
                          shape: BoxShape.circle,
                          border: Border.all(color: Colors.white, width: 3),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.3),
                              blurRadius: 8,
                              spreadRadius: 2,
                            ),
                          ],
                        ),
                        child: const Icon(
                          Icons.location_on,
                          color: Colors.white,
                          size: 24,
                        ),
                      ),
                    ),
                  ],
                ),
              // 版权信息
              RichAttributionWidget(
                attributions: [TextSourceAttribution('天地图', onTap: () {})],
              ),
            ],
          ),
          // 顶部提示信息
          Positioned(
            top: 16,
            left: 16,
            right: 16,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(8),
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
                  if (_isLoadingLocation)
                    const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        valueColor: AlwaysStoppedAnimation<Color>(
                          Color(0xFF20CB6B),
                        ),
                      ),
                    )
                  else if (_locationError != null)
                    const Icon(Icons.error_outline, size: 16, color: Colors.red)
                  else if (_userPosition != null)
                    const Icon(
                      Icons.location_on,
                      size: 16,
                      color: Color(0xFF20CB6B),
                    )
                  else
                    const Icon(
                      Icons.info_outline,
                      size: 16,
                      color: Color(0xFF20CB6B),
                    ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _isLoadingLocation
                          ? '正在获取位置...'
                          : _locationError ??
                                (_userPosition != null
                                    ? '位置: ${_userPosition!.latitude.toStringAsFixed(6)}, ${_userPosition!.longitude.toStringAsFixed(6)}'
                                    : '天地图已集成，路线规划功能开发中...'),
                      style: TextStyle(
                        fontSize: 12,
                        color: _locationError != null
                            ? Colors.red
                            : const Color(0xFF40475C),
                      ),
                    ),
                  ),
                  if (_locationError != null || _userPosition == null)
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        if (_locationError != null &&
                            (_locationError!.contains('定位服务未启用') ||
                                _locationError!.contains('GPS')))
                          TextButton(
                            onPressed: () async {
                              // 打开定位设置页面
                              await LocationService.openLocationSettings();
                              // 延迟一下，等待用户操作
                              await Future.delayed(const Duration(seconds: 3));
                              // 重新获取位置（会自动检查定位服务状态）
                              _getUserLocation();
                            },
                            style: TextButton.styleFrom(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              minimumSize: Size.zero,
                              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                            ),
                            child: const Text(
                              '开启GPS',
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 11,
                                decoration: TextDecoration.underline,
                              ),
                            ),
                          ),
                        IconButton(
                          icon: const Icon(Icons.refresh, size: 18),
                          onPressed: _getUserLocation,
                          padding: EdgeInsets.zero,
                          constraints: const BoxConstraints(),
                          color: Colors.white,
                        ),
                      ],
                    ),
                ],
              ),
            ),
          ),
          // 定位按钮（右下角）
          Positioned(
            bottom: 16,
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
        ],
      ),
    );
  }
}
