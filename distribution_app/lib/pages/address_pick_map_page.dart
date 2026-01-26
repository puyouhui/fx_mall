import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'dart:async';
import '../utils/request.dart';

/// 地址选点页面（配送员端）
class AddressPickMapPage extends StatefulWidget {
  final LatLng? initialCenter;
  final LatLng? initialSelected;

  const AddressPickMapPage({
    super.key,
    this.initialCenter,
    this.initialSelected,
  });

  @override
  State<AddressPickMapPage> createState() => _AddressPickMapPageState();
}

class _AddressPickMapPageState extends State<AddressPickMapPage> {
  final MapController _mapController = MapController();
  LatLng _currentCenter = const LatLng(25.0389, 102.7183); // 默认昆明

  LatLng? _myLocation;
  bool _isLocating = false;
  String? _locateError;
  bool _isDialogOpen = false;

  StreamSubscription<Position>? _posSub;
  late final _LifecycleObserver _lifecycleObserver;

  // 搜索相关
  final TextEditingController _searchController = TextEditingController();
  bool _isSearching = false;
  List<Map<String, dynamic>> _searchResults = [];
  LatLng? _selectedSearchResult;

  @override
  void initState() {
    super.initState();
    _currentCenter =
        widget.initialSelected ??
        widget.initialCenter ??
        const LatLng(25.0389, 102.7183);

    _lifecycleObserver = _LifecycleObserver(onResumed: _onAppResumed);
    WidgetsBinding.instance.addObserver(_lifecycleObserver);
    
    if (widget.initialSelected != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _mapController.move(widget.initialSelected!, 16);
      });
    } else {
      _initMyLocationAndCenter();
    }
  }

  @override
  void dispose() {
    _posSub?.cancel();
    _searchController.dispose();
    WidgetsBinding.instance.removeObserver(_lifecycleObserver);
    super.dispose();
  }

  void _onAppResumed() {
    if (!mounted) return;
    if (_isLocating) return;
    if (_myLocation == null || _locateError != null) {
      _initMyLocationAndCenter();
    }
  }

  void _confirm() {
    final result = _selectedSearchResult ?? _mapController.camera.center;
    Navigator.of(context).pop<LatLng>(result);
  }

  // 搜索POI
  Future<void> _searchPOI(String keyword) async {
    if (keyword.trim().isEmpty) {
      setState(() {
        _searchResults = [];
        _selectedSearchResult = null;
      });
      return;
    }

    setState(() {
      _isSearching = true;
      _searchResults = [];
      _selectedSearchResult = null;
    });

    try {
      final center = _mapController.camera.center;
      final location = '${center.longitude},${center.latitude}';

      final response = await Request.post<Map<String, dynamic>>(
        '/employee/addresses/search-poi',
        body: {
          'keyword': keyword,
          'location': location,
        },
        parser: (data) => data as Map<String, dynamic>,
      );

      if (!mounted) return;

      if (response.isSuccess && response.data != null) {
        final data = response.data!;
        final success = data['success'] as bool? ?? false;
        final results = data['results'] as List<dynamic>? ?? [];

        if (success && results.isNotEmpty) {
          setState(() {
            _searchResults = results
                .map((r) => r as Map<String, dynamic>)
                .toList();
            if (_searchResults.isNotEmpty) {
              final first = _searchResults[0];
              final lat = first['latitude'] as num?;
              final lng = first['longitude'] as num?;
              if (lat != null && lng != null) {
                _selectedSearchResult = LatLng(lat.toDouble(), lng.toDouble());
                _mapController.move(_selectedSearchResult!, 16);
              }
            }
          });
        } else {
          setState(() {
            _searchResults = [];
            _selectedSearchResult = null;
          });
        }
      } else {
        setState(() {
          _searchResults = [];
          _selectedSearchResult = null;
        });
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _searchResults = [];
        _selectedSearchResult = null;
      });
    } finally {
      if (mounted) {
        setState(() {
          _isSearching = false;
        });
      }
    }
  }

  void _selectSearchResult(Map<String, dynamic> result) {
    final lat = result['latitude'] as num?;
    final lng = result['longitude'] as num?;
    if (lat != null && lng != null) {
      setState(() {
        _selectedSearchResult = LatLng(lat.toDouble(), lng.toDouble());
      });
      _mapController.move(_selectedSearchResult!, 16);
    }
  }

  TileProvider _createTiandituTileProvider() {
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

  Future<void> _initMyLocationAndCenter() async {
    setState(() {
      _isLocating = true;
      _locateError = null;
    });

    try {
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
        if (widget.initialSelected == null) {
          _currentCenter = my;
          WidgetsBinding.instance.addPostFrameCallback((_) {
            _mapController.move(my, 16);
          });
        }
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
    try {
      final last = await Geolocator.getLastKnownPosition();
      if (last != null) return last;
    } catch (_) {}

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
      } catch (_) {}
    }

    final completer = Completer<Position?>();
    try {
      await _posSub?.cancel();
      _posSub = Geolocator.getPositionStream(
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
        backgroundColor: const Color(0xFF20CB6B),
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
                userAgentPackageName: 'com.example.distribution_app',
                maxNativeZoom: 18,
                maxZoom: 18,
                tileProvider: _createTiandituTileProvider(),
              ),
              TileLayer(
                urlTemplate: tiandituLabelUrlTemplate,
                subdomains: const ['0', '1', '2', '3', '4', '5', '6', '7'],
                userAgentPackageName: 'com.example.distribution_app',
                maxNativeZoom: 18,
                maxZoom: 18,
                tileProvider: _createTiandituTileProvider(),
              ),
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
              if (_searchResults.isNotEmpty)
                MarkerLayer(
                  markers: _searchResults.map((result) {
                    final lat = result['latitude'] as num?;
                    final lng = result['longitude'] as num?;
                    if (lat == null || lng == null) return null;
                    final point = LatLng(lat.toDouble(), lng.toDouble());
                    final isSelected = _selectedSearchResult == point;

                    return Marker(
                      point: point,
                      width: isSelected ? 28 : 24,
                      height: isSelected ? 28 : 24,
                      alignment: Alignment.center,
                      child: Container(
                        width: isSelected ? 28 : 24,
                        height: isSelected ? 28 : 24,
                        decoration: BoxDecoration(
                          color: isSelected
                              ? const Color(0xFFFF9800)
                              : const Color(0xFF2196F3),
                          shape: BoxShape.circle,
                          border: Border.all(
                            color: Colors.white,
                            width: isSelected ? 2.5 : 2,
                          ),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.25),
                              blurRadius: 6,
                              offset: const Offset(0, 2),
                            ),
                          ],
                        ),
                        child: const Icon(
                          Icons.location_on,
                          color: Colors.white,
                          size: 16,
                        ),
                      ),
                    );
                  }).whereType<Marker>().toList(),
                ),
            ],
          ),
          const IgnorePointer(
            ignoring: true,
            child: Center(child: _CenterPickMarker()),
          ),
          Positioned(
            top: 16,
            left: 16,
            right: 16,
            child: Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(12),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.1),
                    blurRadius: 10,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  TextField(
                    controller: _searchController,
                    decoration: InputDecoration(
                      hintText: '搜索地址、地点',
                      hintStyle: const TextStyle(
                        fontSize: 14,
                        color: Color(0xFF8C92A4),
                      ),
                      prefixIcon: const Icon(
                        Icons.search,
                        color: Color(0xFF8C92A4),
                        size: 20,
                      ),
                      suffixIcon: _searchController.text.isNotEmpty
                          ? IconButton(
                              icon: _isSearching
                                  ? const SizedBox(
                                      width: 20,
                                      height: 20,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2,
                                        valueColor: AlwaysStoppedAnimation<Color>(
                                          Color(0xFF8C92A4),
                                        ),
                                      ),
                                    )
                                  : const Icon(
                                      Icons.clear,
                                      color: Color(0xFF8C92A4),
                                      size: 20,
                                    ),
                              onPressed: _isSearching
                                  ? null
                                  : () {
                                      _searchController.clear();
                                      setState(() {
                                        _searchResults = [];
                                        _selectedSearchResult = null;
                                      });
                                    },
                            )
                          : null,
                      border: InputBorder.none,
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 12,
                      ),
                    ),
                    style: const TextStyle(
                      fontSize: 14,
                      color: Color(0xFF20253A),
                    ),
                    onChanged: (value) {
                      setState(() {});
                      Future.delayed(const Duration(milliseconds: 500), () {
                        if (_searchController.text == value && value.isNotEmpty) {
                          _searchPOI(value);
                        }
                      });
                    },
                    onSubmitted: (value) {
                      if (value.isNotEmpty) {
                        _searchPOI(value);
                      }
                    },
                  ),
                  if (_searchResults.isNotEmpty)
                    Container(
                      constraints: const BoxConstraints(maxHeight: 200),
                      decoration: const BoxDecoration(
                        border: Border(
                          top: BorderSide(
                            color: Color(0xFFE5E7F0),
                            width: 1,
                          ),
                        ),
                      ),
                      child: ListView.builder(
                        shrinkWrap: true,
                        itemCount: _searchResults.length,
                        itemBuilder: (context, index) {
                          final result = _searchResults[index];
                          final name = result['name'] as String? ?? '';
                          final address = result['address'] as String? ?? '';
                          final lat = result['latitude'] as num?;
                          final lng = result['longitude'] as num?;
                          final point = (lat != null && lng != null)
                              ? LatLng(lat.toDouble(), lng.toDouble())
                              : null;
                          final isSelected = _selectedSearchResult == point;

                          return InkWell(
                            onTap: () => _selectSearchResult(result),
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 12,
                              ),
                              decoration: BoxDecoration(
                                color: isSelected
                                    ? const Color(0xFFFF9800).withOpacity(0.1)
                                    : Colors.transparent,
                                border: Border(
                                  bottom: BorderSide(
                                    color: const Color(0xFFE5E7F0),
                                    width: index < _searchResults.length - 1 ? 1 : 0,
                                  ),
                                ),
                              ),
                              child: Row(
                                children: [
                                  Icon(
                                    Icons.location_on,
                                    color: isSelected
                                        ? const Color(0xFFFF9800)
                                        : const Color(0xFF2196F3),
                                    size: 20,
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          name,
                                          style: TextStyle(
                                            fontSize: 14,
                                            fontWeight: isSelected
                                                ? FontWeight.w600
                                                : FontWeight.w500,
                                            color: const Color(0xFF20253A),
                                          ),
                                        ),
                                        if (address.isNotEmpty) ...[
                                          const SizedBox(height: 4),
                                          Text(
                                            address,
                                            style: const TextStyle(
                                              fontSize: 12,
                                              color: Color(0xFF8C92A4),
                                            ),
                                            maxLines: 1,
                                            overflow: TextOverflow.ellipsis,
                                          ),
                                        ],
                                      ],
                                    ),
                                  ),
                                  if (isSelected)
                                    const Icon(
                                      Icons.check_circle,
                                      color: Color(0xFFFF9800),
                                      size: 20,
                                    ),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
                    ),
                ],
              ),
            ),
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

    final radius = size.width * 0.38;
    canvas.drawCircle(center.translate(0, 1.2), radius, shadowPaint);
    canvas.drawCircle(center, radius, ringPaint);

    final gap = 6.5;
    final len = radius + 6;
    canvas.drawLine(
      Offset(center.dx, center.dy - len),
      Offset(center.dx, center.dy - gap),
      linePaint,
    );
    canvas.drawLine(
      Offset(center.dx, center.dy + gap),
      Offset(center.dx, center.dy + len),
      linePaint,
    );
    canvas.drawLine(
      Offset(center.dx - len, center.dy),
      Offset(center.dx - gap, center.dy),
      linePaint,
    );
    canvas.drawLine(
      Offset(center.dx + gap, center.dy),
      Offset(center.dx + len, center.dy),
      linePaint,
    );
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
