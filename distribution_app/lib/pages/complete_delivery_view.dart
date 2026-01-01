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

  // 检查并请求相册权限
  Future<bool> _checkPhotoPermission() async {
    if (Platform.isAndroid) {
      // Android 13+ 使用 READ_MEDIA_IMAGES，Android 12 及以下使用 READ_EXTERNAL_STORAGE
      // permission_handler 会自动处理版本差异
      Permission permission;
      try {
        // 尝试使用 photos 权限（Android 13+）
        permission = Permission.photos;
        final status = await permission.status;
        if (status.isGranted) {
          return true;
        }
        if (status.isPermanentlyDenied) {
          if (mounted) {
            final shouldOpen = await showDialog<bool>(
              context: context,
              builder: (context) => AlertDialog(
                title: const Text('需要相册权限'),
                content: const Text('选择图片需要相册权限，请到设置中开启'),
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
        // 如果权限被拒绝但未永久拒绝，请求权限
        if (status.isDenied) {
          final result = await permission.request();
          if (result.isGranted) {
            return true;
          }
          if (result.isPermanentlyDenied) {
            if (mounted) {
              final shouldOpen = await showDialog<bool>(
                context: context,
                builder: (context) => AlertDialog(
                  title: const Text('需要相册权限'),
                  content: const Text('选择图片需要相册权限，请到设置中开启'),
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
          return false;
        }
        return false;
      } catch (e) {
        // 如果 photos 权限不可用（Android 12 及以下），使用 storage 权限
        permission = Permission.storage;
        final status = await permission.status;
        if (status.isGranted) {
          return true;
        }
        if (status.isPermanentlyDenied) {
          if (mounted) {
            final shouldOpen = await showDialog<bool>(
              context: context,
              builder: (context) => AlertDialog(
                title: const Text('需要存储权限'),
                content: const Text('选择图片需要存储权限，请到设置中开启'),
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
        // 如果权限被拒绝但未永久拒绝，请求权限
        if (status.isDenied) {
          final result = await permission.request();
          if (result.isGranted) {
            return true;
          }
          if (result.isPermanentlyDenied) {
            if (mounted) {
              final shouldOpen = await showDialog<bool>(
                context: context,
                builder: (context) => AlertDialog(
                  title: const Text('需要存储权限'),
                  content: const Text('选择图片需要存储权限，请到设置中开启'),
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
          return false;
        }
        return false;
      }
    }
    // iOS 不需要额外权限（image_picker 会自动处理）
    return true;
  }


  // 从相册选择货物照片
  Future<void> _pickProductImageFromGallery() async {
    await _pickProductImageFromSource(ImageSource.gallery);
  }

  // 拍摄货物照片
  Future<void> _pickProductImageFromCamera() async {
    await _pickProductImageFromSource(ImageSource.camera);
  }

  // 从指定来源选择货物照片
  Future<void> _pickProductImageFromSource(ImageSource source) async {
    try {
      // 检查权限
      if (source == ImageSource.camera) {
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
      } else if (source == ImageSource.gallery) {
        // 对于相册选择，image_picker 会自动处理权限
        // 我们只检查权限状态，如果被永久拒绝则提示用户
        if (Platform.isAndroid) {
          try {
            final photosStatus = await Permission.photos.status;
            if (photosStatus.isPermanentlyDenied) {
              if (mounted) {
                final shouldOpen = await showDialog<bool>(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text('需要相册权限'),
                    content: const Text('选择图片需要相册权限，请到设置中开启'),
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
              return;
            }
            // 如果权限被拒绝但未永久拒绝，尝试请求权限
            if (photosStatus.isDenied) {
              await Permission.photos.request();
            }
          } catch (e) {
            // 如果 photos 权限不可用（Android 12 及以下），尝试 storage 权限
            try {
              final storageStatus = await Permission.storage.status;
              if (storageStatus.isPermanentlyDenied) {
                if (mounted) {
                  final shouldOpen = await showDialog<bool>(
                    context: context,
                    builder: (context) => AlertDialog(
                      title: const Text('需要存储权限'),
                      content: const Text('选择图片需要存储权限，请到设置中开启'),
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
                return;
              }
              // 如果权限被拒绝但未永久拒绝，尝试请求权限
              if (storageStatus.isDenied) {
                await Permission.storage.request();
              }
            } catch (e2) {
              // 如果权限请求失败，继续尝试选择图片（image_picker 可能会自己处理）
              print('权限检查失败: $e2');
            }
          }
        }
      }

      final XFile? image;
      if (source == ImageSource.camera) {
        image = await _imagePicker.pickImage(
          source: source,
          imageQuality: 70, // 降低质量以减少文件大小
          preferredCameraDevice: CameraDevice.rear, // 优先使用后置摄像头
        );
      } else {
        image = await _imagePicker.pickImage(
          source: source,
          imageQuality: 70, // 降低质量以减少文件大小
        );
      }
      if (image != null) {
        setState(() {
          _productImage = File(image!.path);
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              source == ImageSource.camera ? '拍照失败: $e' : '选择图片失败: $e',
            ),
            backgroundColor: Colors.red,
            action: SnackBarAction(
              label: '重试',
              textColor: Colors.white,
              onPressed: () => _pickProductImageFromSource(source),
            ),
          ),
        );
      }
    }
  }

  // 从相册选择门牌照片
  Future<void> _pickDoorplateImageFromGallery() async {
    await _pickDoorplateImageFromSource(ImageSource.gallery);
  }

  // 拍摄门牌照片
  Future<void> _pickDoorplateImageFromCamera() async {
    await _pickDoorplateImageFromSource(ImageSource.camera);
  }

  // 从指定来源选择门牌照片
  Future<void> _pickDoorplateImageFromSource(ImageSource source) async {
    try {
      // 检查权限
      if (source == ImageSource.camera) {
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
      } else if (source == ImageSource.gallery) {
        // 对于相册选择，image_picker 会自动处理权限
        // 我们只检查权限状态，如果被永久拒绝则提示用户
        if (Platform.isAndroid) {
          try {
            final photosStatus = await Permission.photos.status;
            if (photosStatus.isPermanentlyDenied) {
              if (mounted) {
                final shouldOpen = await showDialog<bool>(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text('需要相册权限'),
                    content: const Text('选择图片需要相册权限，请到设置中开启'),
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
              return;
            }
            // 如果权限被拒绝但未永久拒绝，尝试请求权限
            if (photosStatus.isDenied) {
              await Permission.photos.request();
            }
          } catch (e) {
            // 如果 photos 权限不可用（Android 12 及以下），尝试 storage 权限
            try {
              final storageStatus = await Permission.storage.status;
              if (storageStatus.isPermanentlyDenied) {
                if (mounted) {
                  final shouldOpen = await showDialog<bool>(
                    context: context,
                    builder: (context) => AlertDialog(
                      title: const Text('需要存储权限'),
                      content: const Text('选择图片需要存储权限，请到设置中开启'),
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
                return;
              }
              // 如果权限被拒绝但未永久拒绝，尝试请求权限
              if (storageStatus.isDenied) {
                await Permission.storage.request();
              }
            } catch (e2) {
              // 如果权限请求失败，继续尝试选择图片（image_picker 可能会自己处理）
              print('权限检查失败: $e2');
            }
          }
        }
      }

      final XFile? image;
      if (source == ImageSource.camera) {
        image = await _imagePicker.pickImage(
          source: source,
          imageQuality: 70, // 降低质量以减少文件大小
          preferredCameraDevice: CameraDevice.rear, // 优先使用后置摄像头
        );
      } else {
        image = await _imagePicker.pickImage(
          source: source,
          imageQuality: 70, // 降低质量以减少文件大小
        );
      }
      if (image != null) {
        setState(() {
          _doorplateImage = File(image!.path);
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              source == ImageSource.camera ? '拍照失败: $e' : '选择图片失败: $e',
            ),
            backgroundColor: Colors.red,
            action: SnackBarAction(
              label: '重试',
              textColor: Colors.white,
              onPressed: () => _pickDoorplateImageFromSource(source),
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
                      '请拍摄或选择货物照片和门牌照片，用于配送记录',
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
              onPickFromGallery: _pickProductImageFromGallery,
              onPickFromCamera: _pickProductImageFromCamera,
              onRemove: _removeProductImage,
              icon: Icons.inventory_2_outlined,
            ),
            const SizedBox(height: 24),

            // 门牌照片
            _buildImageSection(
              title: '门牌照片',
              image: _doorplateImage,
              onPickFromGallery: _pickDoorplateImageFromGallery,
              onPickFromCamera: _pickDoorplateImageFromCamera,
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
    required VoidCallback onPickFromGallery,
    required VoidCallback onPickFromCamera,
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
          // 显示两个按钮：选择和拍摄
          Container(
            width: double.infinity,
            height: 200,
            decoration: BoxDecoration(
              color: Colors.grey[100],
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: Colors.grey[200]!,
                width: 1,
                style: BorderStyle.solid,
              ),
            ),
            child: Row(
              children: [
                // 左侧：从相册选择按钮
                Expanded(
                  child: Material(
                    color: Colors.transparent,
                    child: InkWell(
                      onTap: onPickFromGallery,
                      borderRadius: const BorderRadius.only(
                        topLeft: Radius.circular(12),
                        bottomLeft: Radius.circular(12),
                      ),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            Icons.photo_library,
                            size: 48,
                            color: Colors.grey[400],
                          ),
                          const SizedBox(height: 8),
                          Text(
                            '选择',
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                              color: Colors.grey[500],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
                // 中间分隔线
                Container(
                  width: 1,
                  height: 120,
                  color: Colors.grey[300]!.withOpacity(0.3),
                ),
                // 右侧：拍照按钮
                Expanded(
                  child: Material(
                    color: Colors.transparent,
                    child: InkWell(
                      onTap: onPickFromCamera,
                      borderRadius: const BorderRadius.only(
                        topRight: Radius.circular(12),
                        bottomRight: Radius.circular(12),
                      ),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            Icons.camera_alt,
                            size: 48,
                            color: Colors.grey[400],
                          ),
                          const SizedBox(height: 8),
                          Text(
                            '拍摄',
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                              color: Colors.grey[500],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
      ],
    );
  }
}

