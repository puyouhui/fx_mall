import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'dart:async';

/// 地址选点页面（固定中心点选址）：
/// - 中心“位置图标”固定在屏幕中心（用于选点，返回它的坐标）
/// - “绿色原点”为我的位置（属于地图图层，拖动地图时会跟随地图变化）
/// - 进入页面后自动把“我的位置”设为地图中心
class CustomerAddressPickMapPage extends StatefulWidget {
  final LatLng? initialCenter;
  final LatLng? initialSelected;

  const CustomerAddressPickMapPage({
    super.key,
    this.initialCenter,
    this.initialSelected,
  });

  @override
  State<CustomerAddressPickMapPage> createState() =>
      _CustomerAddressPickMapPageState();
}

class _CustomerAddressPickMapPageState
    extends State<CustomerAddressPickMapPage> {
  final MapController _mapController = MapController();
  LatLng _currentCenter = const LatLng(25.0389, 102.7183); // 默认昆明

  LatLng? _myLocation;
  bool _isLocating = false;
  String? _locateError;
  bool _isDialogOpen = false;

  StreamSubscription<Position>? _posSub;
  late final _LifecycleObserver _lifecycleObserver;

  @override
  void initState() {
    super.initState();
    _currentCenter =
        widget.initialSelected ??
        widget.initialCenter ??
        const LatLng(25.0389, 102.7183);

    _lifecycleObserver = _LifecycleObserver(onResumed: _onAppResumed);
    WidgetsBinding.instance.addObserver(_lifecycleObserver);
    _initMyLocationAndCenter();
  }

  @override
  void dispose() {
    _posSub?.cancel();
    WidgetsBinding.instance.removeObserver(_lifecycleObserver);
    super.dispose();
  }

  void _onAppResumed() {
    // 从系统设置返回后，自动再尝试一次（避免仍卡在“正在定位/权限被关”状态）
    if (!mounted) return;
    if (_isLocating) return;
    if (_myLocation == null || _locateError != null) {
      _initMyLocationAndCenter();
    }
  }

  void _confirm() {
    // 返回中心“位置图标”的坐标（即地图中心点）
    Navigator.of(context).pop<LatLng>(_mapController.camera.center);
  }

  TileProvider _createTiandituTileProvider() {
    return NetworkTileProvider(
      // 注意：flutter_map 内部会对 headers 做 putIfAbsent，不能使用 const Map（不可变）
      headers: {
        'User-Agent':
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Referer': 'https://lbs.tianditu.gov.cn/',
        'Accept': 'image/webp,image/apng,image/*,*/*;q=0.8',
        'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8',
      },
    );
  }

  Future<void> _initMyLocationAndCenter() async {
    setState(() {
      _isLocating = true;
      _locateError = null;
    });

    try {
      // 用户要求“网络定位”，不优先使用 GPS：
      // 这里不强制要求定位服务开关为“开启”，先尝试网络定位（Android 下 forceAndroidLocationManager=true 更偏向 NETWORK_PROVIDER）。
      final serviceEnabled = await Geolocator.isLocationServiceEnabled();

      var permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
      }
      if (permission == LocationPermission.denied ||
          permission == LocationPermission.deniedForever) {
        setState(() {
          _locateError = permission == LocationPermission.deniedForever
              ? '定位权限已被系统关闭，请到系统设置中开启'
              : '定位权限被拒绝';
        });
        _showPermissionDialog(permission == LocationPermission.deniedForever);
        return;
      }

      final pos = await _getPositionWithFallback();
      if (pos == null) {
        setState(() {
          _locateError = serviceEnabled
              ? '网络定位失败/超时，请检查网络或权限后重试'
              : '定位服务未开启，且网络定位失败，请到系统设置中开启定位服务';
        });
        if (!serviceEnabled) {
          _showLocationServiceDialog();
        }
        return;
      }
      final my = LatLng(pos.latitude, pos.longitude);

      if (!mounted) return;
      setState(() {
        _myLocation = my;
        _currentCenter = my;
      });

      // 地图控制器 move 建议放到下一帧，保证已挂载
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _mapController.move(my, 16);
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _locateError = '定位失败: ${e.toString()}';
      });
    } finally {
      if (mounted) {
        setState(() {
          _isLocating = false;
        });
      }
    }
  }

  Future<Position?> _getPositionWithFallback() async {
    // 1) 先取最后一次已知位置（通常秒回，体验最快）
    try {
      final last = await Geolocator.getLastKnownPosition();
      if (last != null) return last;
    } catch (_) {
      // ignore and fallback
    }

    // 2) 网络定位优先：低精度/最低精度（避免触发慢 GPS）
    // Android：forceAndroidLocationManager=true 更偏向原生 LocationManager（NETWORK_PROVIDER）
    final accuracyLevels = <LocationAccuracy>[
      LocationAccuracy.low,
      LocationAccuracy.lowest,
    ];
    final timeLimits = <Duration>[
      const Duration(seconds: 6),
      const Duration(seconds: 4),
    ];

    for (int i = 0; i < accuracyLevels.length; i++) {
      try {
        final pos = await Geolocator.getCurrentPosition(
          desiredAccuracy: accuracyLevels[i],
          timeLimit: timeLimits[i],
          forceAndroidLocationManager: true,
        );
        return pos;
      } catch (_) {
        // continue
      }
    }

    // 3) 再兜底：订阅一次定位流，取第一个位置（加超时）
    final completer = Completer<Position?>();
    try {
      await _posSub?.cancel();
      _posSub =
          Geolocator.getPositionStream(
            locationSettings: const LocationSettings(
              accuracy: LocationAccuracy.low,
              distanceFilter: 0,
            ),
          ).listen(
            (p) {
              if (!completer.isCompleted) completer.complete(p);
            },
            onError: (_) {
              if (!completer.isCompleted) completer.complete(null);
            },
          );

      return await completer.future.timeout(
        const Duration(seconds: 8),
        onTimeout: () => null,
      );
    } finally {
      await _posSub?.cancel();
      _posSub = null;
    }
  }

  void _showLocationServiceDialog() {
    if (_isDialogOpen || !mounted) return;
    _isDialogOpen = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      showDialog<void>(
        context: context,
        builder: (ctx) {
          return AlertDialog(
            title: const Text('定位服务未开启'),
            content: const Text('请在系统设置中开启定位服务后再尝试。'),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: const Text('取消'),
              ),
              ElevatedButton(
                onPressed: () async {
                  Navigator.of(ctx).pop();
                  await Geolocator.openLocationSettings();
                },
                child: const Text('去开启'),
              ),
            ],
          );
        },
      ).whenComplete(() {
        _isDialogOpen = false;
      });
    });
  }

  void _showPermissionDialog(bool deniedForever) {
    if (_isDialogOpen || !mounted) return;
    _isDialogOpen = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      showDialog<void>(
        context: context,
        builder: (ctx) {
          return AlertDialog(
            title: const Text('无法获取定位权限'),
            content: Text(
              deniedForever
                  ? '定位权限已被系统关闭，请到系统设置中为本应用开启定位权限。'
                  : '请允许定位权限，否则无法自动定位到当前位置。',
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: const Text('知道了'),
              ),
              ElevatedButton(
                onPressed: () async {
                  Navigator.of(ctx).pop();
                  await Geolocator.openAppSettings();
                },
                child: const Text('去设置'),
              ),
            ],
          );
        },
      ).whenComplete(() {
        _isDialogOpen = false;
      });
    });
  }

  @override
  Widget build(BuildContext context) {
    // 天地图瓦片（Web墨卡托，WGS84）
    const String tiandituTileUrlTemplate =
        'https://t{s}.tianditu.gov.cn/img_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=img&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=d95864378581051adb04fe26acb13ecf';
    const String tiandituLabelUrlTemplate =
        'https://t{s}.tianditu.gov.cn/cia_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=cia&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=d95864378581051adb04fe26acb13ecf';

    return Scaffold(
      appBar: AppBar(
        title: const Text(
          '选择地址位置',
          style: TextStyle(color: Colors.white, fontWeight: FontWeight.w600),
        ),
        centerTitle: true,
        backgroundColor: Color(0xFF20CB6B),
        elevation: 0,
        iconTheme: const IconThemeData(color: Colors.white),
        flexibleSpace: const DecoratedBox(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [Color(0xFF20CB6B), Color(0xFF10B05A)],
            ),
          ),
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 12, top: 8, bottom: 8),
            child: FilledButton(
              onPressed: _confirm,
              style: FilledButton.styleFrom(
                backgroundColor: Colors.white,
                foregroundColor: const Color(0xFF20CB6B),
                elevation: 0,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                shape: const StadiumBorder(),
              ),
              child: const Text(
                '确定',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
            ),
          ),
        ],
      ),
      body: Stack(
        children: [
          FlutterMap(
            mapController: _mapController,
            options: MapOptions(
              initialCenter: _currentCenter,
              initialZoom: 16,
              minZoom: 3,
              maxZoom: 18,
              onMapEvent: (event) {
                // 固定中心点选址：拖动/缩放结束后取当前中心点
                if (event is MapEventMoveEnd ||
                    event is MapEventFlingAnimationEnd ||
                    event is MapEventDoubleTapZoomEnd ||
                    event is MapEventScrollWheelZoom ||
                    event is MapEventRotateEnd) {
                  setState(() {
                    _currentCenter = event.camera.center;
                  });
                }
              },
            ),
            children: [
              TileLayer(
                urlTemplate: tiandituTileUrlTemplate,
                subdomains: const ['0', '1', '2', '3', '4', '5', '6', '7'],
                userAgentPackageName: 'com.example.employees_app',
                maxNativeZoom: 18,
                maxZoom: 18,
                tileProvider: _createTiandituTileProvider(),
              ),
              TileLayer(
                urlTemplate: tiandituLabelUrlTemplate,
                subdomains: const ['0', '1', '2', '3', '4', '5', '6', '7'],
                userAgentPackageName: 'com.example.employees_app',
                maxNativeZoom: 18,
                maxZoom: 18,
                tileProvider: _createTiandituTileProvider(),
              ),
              // 我的位置：绿色原点（地图图层，拖动地图会跟随变化）
              if (_myLocation != null)
                MarkerLayer(
                  markers: [
                    Marker(
                      point: _myLocation!,
                      width: 16,
                      height: 16,
                      alignment: Alignment.center,
                      child: Container(
                        width: 14,
                        height: 14,
                        decoration: BoxDecoration(
                          color: const Color(0xFF20CB6B),
                          shape: BoxShape.circle,
                          border: Border.all(color: Colors.white, width: 2),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.25),
                              blurRadius: 6,
                              offset: const Offset(0, 3),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
            ],
          ),
          // 固定在屏幕中心的“位置图标”（用于选点，不拦截手势）
          const IgnorePointer(
            ignoring: true,
            child: Center(child: _CenterPickMarker()),
          ),
          Positioned(
            left: 16,
            right: 16,
            bottom: 16 + MediaQuery.of(context).padding.bottom,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(12),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.08),
                    blurRadius: 12,
                    offset: const Offset(0, 6),
                  ),
                ],
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      _isLocating
                          ? '正在定位...'
                          : (_locateError != null
                                ? _locateError!
                                : '中心点：${_currentCenter.latitude.toStringAsFixed(6)}, ${_currentCenter.longitude.toStringAsFixed(6)}'),
                      style: const TextStyle(
                        fontSize: 13,
                        color: Color(0xFF40475C),
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  if (!_isLocating)
                    TextButton(
                      onPressed: _initMyLocationAndCenter,
                      child: const Text('重试'),
                    ),
                  ElevatedButton(
                    onPressed: _confirm,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF20CB6B),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 10,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                      ),
                      elevation: 0,
                    ),
                    child: const Text('确定'),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _CenterPickMarker extends StatelessWidget {
  const _CenterPickMarker();

  @override
  Widget build(BuildContext context) {
    // 中心点准星（绿色）：固定在屏幕中心，用于选点
    return SizedBox(
      width: 44,
      height: 44,
      child: CustomPaint(
        painter: _CrosshairPainter(color: const Color(0xFF20CB6B)),
      ),
    );
  }
}

class _CrosshairPainter extends CustomPainter {
  final Color color;

  const _CrosshairPainter({required this.color});

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);

    // 阴影（轻微立体感）
    final shadowPaint = Paint()
      ..color = Colors.black.withOpacity(0.18)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 3;

    final ringPaint = Paint()
      ..color = color.withOpacity(0.95)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2.6
      ..strokeCap = StrokeCap.round;

    final linePaint = Paint()
      ..color = color.withOpacity(0.95)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 3.2
      ..strokeCap = StrokeCap.round;

    final dotPaint = Paint()
      ..color = color
      ..style = PaintingStyle.fill;

    // 外圈
    final radius = size.width * 0.38;
    canvas.drawCircle(
      center.translate(0, 1.2),
      radius,
      shadowPaint,
    ); // shadow offset
    canvas.drawCircle(center, radius, ringPaint);

    // 十字线（中间留空，避免遮挡“我的位置”绿色原点）
    final gap = 6.5;
    final len = radius + 6;

    // 上
    canvas.drawLine(
      Offset(center.dx, center.dy - len),
      Offset(center.dx, center.dy - gap),
      linePaint,
    );
    // 下
    canvas.drawLine(
      Offset(center.dx, center.dy + gap),
      Offset(center.dx, center.dy + len),
      linePaint,
    );
    // 左
    canvas.drawLine(
      Offset(center.dx - len, center.dy),
      Offset(center.dx - gap, center.dy),
      linePaint,
    );
    // 右
    canvas.drawLine(
      Offset(center.dx + gap, center.dy),
      Offset(center.dx + len, center.dy),
      linePaint,
    );

    // 中心小点（点很小，主要用于视觉聚焦）
    canvas.drawCircle(center, 2.2, dotPaint);
  }

  @override
  bool shouldRepaint(covariant _CrosshairPainter oldDelegate) {
    return oldDelegate.color != color;
  }
}

class _LifecycleObserver extends WidgetsBindingObserver {
  final VoidCallback onResumed;

  _LifecycleObserver({required this.onResumed});

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      onResumed();
    }
  }
}
