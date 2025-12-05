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

  @override
  void initState() {
    super.initState();
    // 延迟一下，确保UI已经渲染完成后再请求权限
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _getLocation();
    });
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
          _locationError = '定位服务未启用，请先开启GPS';
        });
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
      if (permission == LocationPermission.deniedForever ||
          permissionHandlerStatus.isPermanentlyDenied) {
        errorMsg = '定位权限被永久拒绝，请到设置中开启';
      } else if (permission == LocationPermission.denied ||
          permissionHandlerStatus.isDenied) {
        // 小米手机可能没有弹出对话框
        errorMsg = '定位权限未授予（小米手机请到设置中手动开启）';
      }

      if (mounted) {
        setState(() {
          _isLoadingLocation = false;
          _locationError = errorMsg;
        });
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
            currentPosition: _currentPosition,
            isLoadingLocation: _isLoadingLocation,
            locationError: _locationError,
            onRefreshLocation: _getLocation,
          ),
          ProfileView(courierPhone: widget.courierPhone),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentIndex,
        onDestinationSelected: (index) {
          setState(() {
            _currentIndex = index;
          });
        },
        backgroundColor: Colors.white,
        indicatorColor: const Color(0xFF20CB6B).withOpacity(0.1),
        labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.list_alt_outlined),
            selectedIcon: Icon(Icons.list_alt),
            label: '接单大厅',
          ),
          NavigationDestination(
            icon: Icon(Icons.person_outline),
            selectedIcon: Icon(Icons.person),
            label: '我的',
          ),
        ],
      ),
    );
  }
}
