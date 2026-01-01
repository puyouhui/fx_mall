import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:permission_handler/permission_handler.dart';
import 'order_hall_view.dart';
import 'profile_view.dart';
import '../utils/location_service.dart';

/// 登录后的主框架：底部两个 Tab（接单大厅 / 我的）
class MainShell extends StatefulWidget {
  const MainShell({super.key, required this.courierPhone});

  final String courierPhone;

  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _currentIndex = 0;
  Position? _currentPosition;
  bool _isLoadingLocation = false;
  String? _locationError;
  final OrderHallViewKey _orderHallViewKey = OrderHallViewKey();

  @override
  void initState() {
    super.initState();
    // 立即开始获取位置，不等待UI渲染完成，以优化接单时的响应速度
    _getLocation();
    // 初始状态：接单大厅可见
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _orderHallViewKey.setPageVisible(true);
    });
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
                '为了提供配送服务，需要开启系统定位服务。\n\n请点击"去设置"打开系统定位设置。',
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

  /// 显示定位权限对话框
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
                '为了提供配送服务，需要获取您的位置信息。\n\n请点击"去设置"手动开启定位权限。',
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

  /// 获取定位信息
  Future<void> _getLocation() async {
    print('[MainShell] 开始获取定位信息...');

    if (mounted) {
      setState(() {
        _isLoadingLocation = true;
        _locationError = null;
      });
    }

    // 先检查定位服务是否启用
    final serviceEnabled = await LocationService.checkLocationServiceEnabled();
    if (!serviceEnabled) {
      if (mounted) {
        setState(() {
          _isLoadingLocation = false;
          _locationError = '定位服务未启用，请先开启系统定位服务';
        });
        // 显示对话框引导用户打开系统设置
        final shouldOpenSettings = await _showLocationServiceDialog();
        if (shouldOpenSettings) {
          await LocationService.openLocationSettings();
          // 延迟一下，等待用户操作
          await Future.delayed(const Duration(seconds: 2));
          // 重新检查定位服务
          _getLocation();
        }
      }
      return;
    }

    // 检查并请求权限
    final hasPermission = await LocationService.checkAndRequestPermission();
    print('[MainShell] 权限检查结果: $hasPermission');

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
        // 小米手机可能没有弹出对话框
        errorMsg = '定位权限未授予（小米手机请到设置中手动开启）';
        needShowDialog = true;
      }

      if (mounted) {
        setState(() {
          _isLoadingLocation = false;
          _locationError = errorMsg;
        });

        // 显示对话框引导用户打开应用设置
        if (needShowDialog) {
          final shouldOpenSettings = await _showPermissionDialog();
          if (shouldOpenSettings) {
            await LocationService.openAppSettingsPage();
            // 延迟一下，等待用户操作
            await Future.delayed(const Duration(seconds: 2));
            // 重新检查权限
            _getLocation();
          }
        }
      }
      return;
    }

    // 获取位置
    final position = await LocationService.getCurrentLocation();
    print('[MainShell] 定位结果: ${position != null ? "成功" : "失败"}');

    if (mounted) {
      setState(() {
        _isLoadingLocation = false;
        if (position == null) {
          _locationError = '定位失败，请检查GPS设置';
        } else {
          _currentPosition = position;
          _locationError = null;
        }
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: _currentIndex == 0
          ? null // 接单大厅使用自定义头部
          : AppBar(
              title: const Text('我的'),
              centerTitle: true,
              backgroundColor: const Color(0xFF20CB6B),
              elevation: 0,
              automaticallyImplyLeading: false,
              iconTheme: const IconThemeData(color: Colors.white),
              titleTextStyle: const TextStyle(
                color: Colors.white,
                fontSize: 18,
                fontWeight: FontWeight.w600,
              ),
            ),
      extendBody: true,
      body: IndexedStack(
        index: _currentIndex,
        children: [
          OrderHallView(
            key: _orderHallViewKey,
            currentPosition: _currentPosition,
            isLoadingLocation: _isLoadingLocation,
            locationError: _locationError,
            onRefreshLocation: _getLocation,
          ),
          ProfileView(courierPhone: widget.courierPhone),
        ],
      ),
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.08),
              blurRadius: 12,
              offset: const Offset(0, -2),
            ),
          ],
        ),
        child: NavigationBar(
          selectedIndex: _currentIndex,
          onDestinationSelected: (index) {
            // 更新页面可见性
            _orderHallViewKey.setPageVisible(index == 0);
            setState(() {
              _currentIndex = index;
            });
          },
          backgroundColor: Colors.transparent,
          elevation: 0,
          indicatorColor: const Color(0xFF20CB6B).withOpacity(0.15),
          labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
          labelTextStyle: MaterialStateProperty.resolveWith((states) {
            if (states.contains(MaterialState.selected)) {
              return const TextStyle(
                color: Color(0xFF20CB6B),
                fontSize: 12,
                fontWeight: FontWeight.w600,
              );
            }
            return const TextStyle(
              color: Color(0xFF8C92A4),
              fontSize: 12,
              fontWeight: FontWeight.normal,
            );
          }),
          destinations: [
            NavigationDestination(
              icon: const Icon(
                Icons.list_alt_outlined,
                color: Color(0xFF8C92A4),
                size: 24,
              ),
              selectedIcon: const Icon(
                Icons.list_alt,
                color: Color(0xFF20CB6B),
                size: 24,
              ),
              label: '任务',
            ),
            NavigationDestination(
              icon: const Icon(
                Icons.person_outline,
                color: Color(0xFF8C92A4),
                size: 24,
              ),
              selectedIcon: const Icon(
                Icons.person,
                color: Color(0xFF20CB6B),
                size: 24,
              ),
              label: '我的',
            ),
          ],
        ),
      ),
    );
  }
}
