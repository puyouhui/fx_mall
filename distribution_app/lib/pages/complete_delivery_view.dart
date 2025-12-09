import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:permission_handler/permission_handler.dart';
import 'dart:io';
import '../api/order_api.dart';

/// 配送完成页面：上传货物照片和门牌照片
class CompleteDeliveryView extends StatefulWidget {
  const CompleteDeliveryView({super.key});

  @override
  State<CompleteDeliveryView> createState() => _CompleteDeliveryViewState();
}

class _CompleteDeliveryViewState extends State<CompleteDeliveryView> {
  final ImagePicker _imagePicker = ImagePicker();
  File? _productImage; // 货物照片
  File? _doorplateImage; // 门牌照片
  bool _isSubmitting = false;

  int? _orderId;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // 从路由参数获取订单ID
    final args = ModalRoute.of(context)?.settings.arguments as Map<String, dynamic>?;
    _orderId = args?['orderId'] as int?;
  }

  // 检查并请求相机权限
  Future<bool> _checkCameraPermission() async {
    final status = await Permission.camera.status;
    if (status.isGranted) {
      return true;
    }

    if (status.isPermanentlyDenied) {
      if (mounted) {
        final shouldOpen = await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: const Text('需要相机权限'),
            content: const Text('拍照功能需要相机权限，请到设置中开启'),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(context).pop(false),
                child: const Text('取消'),
              ),
              ElevatedButton(
                onPressed: () => Navigator.of(context).pop(true),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF20CB6B),
                  foregroundColor: Colors.white,
                ),
                child: const Text('去设置'),
              ),
            ],
          ),
        );
        if (shouldOpen == true) {
          await openAppSettings();
        }
      }
      return false;
    }

    final result = await Permission.camera.request();
    return result.isGranted;
  }

  // 选择货物照片
  Future<void> _pickProductImage() async {
    // 先检查相机权限
    final hasPermission = await _checkCameraPermission();
    if (!hasPermission) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('需要相机权限才能拍照'),
            backgroundColor: Colors.orange,
          ),
        );
      }
      return;
    }

    try {
      final XFile? image = await _imagePicker.pickImage(
        source: ImageSource.camera,
        imageQuality: 70, // 降低质量以减少文件大小
        preferredCameraDevice: CameraDevice.rear, // 优先使用后置摄像头
      );
      if (image != null) {
        setState(() {
          _productImage = File(image.path);
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('拍照失败: $e'),
            backgroundColor: Colors.red,
            action: SnackBarAction(
              label: '重试',
              textColor: Colors.white,
              onPressed: _pickProductImage,
            ),
          ),
        );
      }
    }
  }

  // 选择门牌照片
  Future<void> _pickDoorplateImage() async {
    // 先检查相机权限
    final hasPermission = await _checkCameraPermission();
    if (!hasPermission) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('需要相机权限才能拍照'),
            backgroundColor: Colors.orange,
          ),
        );
      }
      return;
    }

    try {
      final XFile? image = await _imagePicker.pickImage(
        source: ImageSource.camera,
        imageQuality: 70, // 降低质量以减少文件大小
        preferredCameraDevice: CameraDevice.rear, // 优先使用后置摄像头
      );
      if (image != null) {
        setState(() {
          _doorplateImage = File(image.path);
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('拍照失败: $e'),
            backgroundColor: Colors.red,
            action: SnackBarAction(
              label: '重试',
              textColor: Colors.white,
              onPressed: _pickDoorplateImage,
            ),
          ),
        );
      }
    }
  }

  // 删除货物照片
  void _removeProductImage() {
    setState(() {
      _productImage = null;
    });
  }

  // 删除门牌照片
  void _removeDoorplateImage() {
    setState(() {
      _doorplateImage = null;
    });
  }

  // 提交配送完成
  Future<void> _submitDelivery() async {
    if (_orderId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('订单ID缺失'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    if (_productImage == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('请拍摄货物照片'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    if (_doorplateImage == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('请拍摄门牌照片'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    setState(() {
      _isSubmitting = true;
    });

    try {
      final response = await OrderApi.completeOrderWithImages(
        orderId: _orderId!,
        productImage: _productImage!,
        doorplateImage: _doorplateImage!,
      );

      if (!mounted) return;

      if (response.isSuccess) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('配送完成'),
            backgroundColor: Color(0xFF20CB6B),
          ),
        );
        // 返回上一页，并传递成功标识
        Navigator.of(context).pop(true);
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              response.message.isNotEmpty
                  ? response.message
                  : '提交失败，请稍后重试',
            ),
            backgroundColor: Colors.red,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('提交失败: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isSubmitting = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('配送完成'),
        backgroundColor: const Color(0xFF20CB6B),
        foregroundColor: Colors.white,
        elevation: 0,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // 提示信息
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: const Color(0xFFF0F9FF),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: const Color(0xFF20CB6B), width: 1),
              ),
              child: Row(
                children: [
                  const Icon(
                    Icons.info_outline,
                    color: Color(0xFF20CB6B),
                    size: 20,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '请拍摄货物照片和门牌照片，用于配送记录',
                      style: TextStyle(
                        fontSize: 14,
                        color: Colors.grey[700],
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 24),

            // 货物照片
            _buildImageSection(
              title: '货物照片',
              image: _productImage,
              onPick: _pickProductImage,
              onRemove: _removeProductImage,
              icon: Icons.inventory_2_outlined,
            ),
            const SizedBox(height: 24),

            // 门牌照片
            _buildImageSection(
              title: '门牌照片',
              image: _doorplateImage,
              onPick: _pickDoorplateImage,
              onRemove: _removeDoorplateImage,
              icon: Icons.home_outlined,
            ),
            const SizedBox(height: 32),

            // 提交按钮
            SizedBox(
              height: 50,
              child: ElevatedButton(
                onPressed: _isSubmitting ? null : _submitDelivery,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF20CB6B),
                  foregroundColor: Colors.white,
                  disabledBackgroundColor: const Color(0xFF9EDFB9),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  elevation: 0,
                ),
                child: _isSubmitting
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                        ),
                      )
                    : const Text(
                        '提交配送完成',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  // 构建图片选择区域
  Widget _buildImageSection({
    required String title,
    required File? image,
    required VoidCallback onPick,
    required VoidCallback onRemove,
    required IconData icon,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, color: const Color(0xFF20CB6B), size: 20),
            const SizedBox(width: 8),
            Text(
              title,
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: Color(0xFF40475C),
              ),
            ),
            const Text(
              ' *',
              style: TextStyle(color: Colors.red, fontSize: 16),
            ),
          ],
        ),
        const SizedBox(height: 12),
        if (image != null)
          // 显示已选择的图片
          Stack(
            children: [
              Container(
                width: double.infinity,
                height: 200,
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.grey[300]!, width: 1),
                ),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: Image.file(
                    image,
                    fit: BoxFit.cover,
                  ),
                ),
              ),
              Positioned(
                top: 8,
                right: 8,
                child: Material(
                  color: Colors.black54,
                  borderRadius: BorderRadius.circular(20),
                  child: InkWell(
                    onTap: onRemove,
                    borderRadius: BorderRadius.circular(20),
                    child: const Padding(
                      padding: EdgeInsets.all(8),
                      child: Icon(
                        Icons.close,
                        color: Colors.white,
                        size: 20,
                      ),
                    ),
                  ),
                ),
              ),
            ],
          )
        else
          // 显示拍照按钮
          InkWell(
            onTap: onPick,
            child: Container(
              width: double.infinity,
              height: 200,
              decoration: BoxDecoration(
                color: Colors.grey[100],
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: Colors.grey[300]!,
                  width: 1,
                  style: BorderStyle.solid,
                ),
              ),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.camera_alt_outlined,
                    size: 48,
                    color: Colors.grey[400],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    '点击拍照',
                    style: TextStyle(
                      fontSize: 14,
                      color: Colors.grey[600],
                    ),
                  ),
                ],
              ),
            ),
          ),
      ],
    );
  }
}

