import 'dart:io';

import 'package:employees_app/utils/request.dart';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:latlong2/latlong.dart';
import 'package:employees_app/pages/customer/customer_address_pick_map_page.dart';
import 'package:employees_app/utils/coordinate_transform.dart';

/// 客户地址新增/编辑页（用于替代弹框，提升可操作性）
class CustomerAddressEditPage extends StatefulWidget {
  final int userId;
  final Map<String, dynamic>? address;

  /// 从资料完善页带过来的“当前客户资料草稿”，用于在保存地址前先保存客户基础资料，
  /// 避免返回刷新导致资料被旧数据覆盖。
  final Map<String, dynamic>? profileDraft;

  /// 新增地址时：是否默认选中“设为默认地址”
  final bool defaultSelected;

  const CustomerAddressEditPage({
    super.key,
    required this.userId,
    this.address,
    this.profileDraft,
    this.defaultSelected = false,
  });

  @override
  State<CustomerAddressEditPage> createState() =>
      _CustomerAddressEditPageState();
}

class _CustomerAddressEditPageState extends State<CustomerAddressEditPage> {
  final _formKey = GlobalKey<FormState>();

  final _nameController = TextEditingController();
  final _contactController = TextEditingController();
  final _phoneController = TextEditingController();
  final _addressController = TextEditingController();

  final ImagePicker _imagePicker = ImagePicker();

  bool _isSaving = false;
  bool _isUploading = false;
  bool _isGeocoding = false;
  bool _isDefault = false;
  String _avatarUrl = '';

  // 地址坐标（存储为 GCJ-02，用于后端/高德/腾讯）
  double? _latitudeGcj;
  double? _longitudeGcj;

  bool get _isEdit => widget.address != null;

  @override
  void initState() {
    super.initState();

    final addr = widget.address;
    _nameController.text = addr?['name'] as String? ?? '';
    _contactController.text = addr?['contact'] as String? ?? '';
    _phoneController.text = addr?['phone'] as String? ?? '';
    _addressController.text = addr?['address'] as String? ?? '';
    _avatarUrl = addr?['avatar'] as String? ?? '';

    final lat = addr?['latitude'];
    final lng = addr?['longitude'];
    if (lat is num && lng is num) {
      _latitudeGcj = lat.toDouble();
      _longitudeGcj = lng.toDouble();
    }

    _isDefault = _isEdit
        ? ((addr?['is_default'] as bool?) ?? false)
        : widget.defaultSelected;
  }

  @override
  void dispose() {
    _nameController.dispose();
    _contactController.dispose();
    _phoneController.dispose();
    _addressController.dispose();
    super.dispose();
  }

  void _showSnack(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
  }

  Future<void> _pickAndUpload(ImageSource source) async {
    if (_isUploading) return;
    // 统一压缩（相册/拍照都生效）：降低质量并限制尺寸，减少上传体积
    // imageQuality: 0-100，数值越小压缩越狠（体积更小）
    final picked = await _imagePicker.pickImage(
      source: source,
      imageQuality: 30,
      maxWidth: 1600,
      maxHeight: 1600,
    );
    if (picked == null) return;

    setState(() {
      _isUploading = true;
    });

    try {
      final uploadResp = await Request.uploadFile(
        '/employee/upload/address-avatar',
        File(picked.path),
      );

      if (uploadResp.isSuccess && uploadResp.data != null) {
        final data = uploadResp.data!;
        final url = (data['avatar'] ?? data['imageUrl']) as String?;
        if (url != null && url.isNotEmpty) {
          setState(() {
            _avatarUrl = url;
          });
        } else {
          _showSnack('图片上传失败：返回数据缺少URL');
        }
      } else {
        _showSnack(
          uploadResp.message.isNotEmpty ? uploadResp.message : '图片上传失败',
        );
      }
    } catch (e) {
      _showSnack('图片上传失败: ${e.toString()}');
    } finally {
      if (mounted) {
        setState(() {
          _isUploading = false;
        });
      }
    }
  }

  Future<bool> _saveProfileIfNeeded() async {
    final profile = widget.profileDraft;
    if (profile == null) return true;

    try {
      final profileResp = await Request.put<dynamic>(
        '/employee/sales/customers/${widget.userId}/profile',
        body: profile,
      );
      if (!profileResp.isSuccess) {
        _showSnack(
          profileResp.message.isNotEmpty ? profileResp.message : '基础资料保存失败',
        );
        return false;
      }
      return true;
    } catch (e) {
      _showSnack('基础资料保存失败: ${e.toString()}');
      return false;
    }
  }

  Future<void> _save() async {
    if (_isSaving) return;
    if (!_formKey.currentState!.validate()) return;
    if (_latitudeGcj == null || _longitudeGcj == null) {
      _showSnack('请先选择地址位置');
      return;
    }

    setState(() {
      _isSaving = true;
    });

    try {
      // 先保存资料完善页上正在编辑的基础资料（避免返回刷新覆盖）
      final ok = await _saveProfileIfNeeded();
      if (!ok) {
        return;
      }

      final body = <String, dynamic>{
        'name': _nameController.text.trim(),
        'contact': _contactController.text.trim(),
        'phone': _phoneController.text.trim(),
        'address': _addressController.text.trim(),
        'avatar': _avatarUrl,
        'isDefault': _isDefault,
        'latitude': _latitudeGcj,
        'longitude': _longitudeGcj,
      };

      if (_isEdit) {
        final addrId = widget.address!['id'] as int;
        final resp = await Request.put<dynamic>(
          '/employee/sales/addresses/$addrId',
          body: body,
        );
        if (!resp.isSuccess) {
          _showSnack(resp.message.isNotEmpty ? resp.message : '地址更新失败');
          return;
        }
      } else {
        final resp = await Request.post<dynamic>(
          '/employee/sales/customers/${widget.userId}/addresses',
          body: body,
        );
        if (!resp.isSuccess) {
          _showSnack(resp.message.isNotEmpty ? resp.message : '地址新增失败');
          return;
        }
      }

      if (!mounted) return;
      Navigator.of(context).pop(true);
    } catch (e) {
      _showSnack('操作失败: ${e.toString()}');
    } finally {
      if (mounted) {
        setState(() {
          _isSaving = false;
        });
      }
    }
  }

  InputDecoration _inputDecoration(String label, {String? hintText}) {
    return InputDecoration(
      labelText: label,
      hintText: hintText,
      filled: true,
      fillColor: const Color(0xFFF7F8FA),
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: BorderSide.none,
      ),
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
    );
  }

  LatLng? _getSelectedWgs84() {
    if (_latitudeGcj == null || _longitudeGcj == null) return null;
    return CoordinateTransform.gcj02ToWgs84(_latitudeGcj!, _longitudeGcj!);
  }

  Future<void> _openPickMap() async {
    final selectedWgs84 = _getSelectedWgs84();
    final initialCenter = selectedWgs84 ?? const LatLng(25.0389, 102.7183);

    final pickedWgs84 = await Navigator.of(context).push<LatLng>(
      MaterialPageRoute(
        builder: (_) => CustomerAddressPickMapPage(
          initialCenter: initialCenter,
          initialSelected: selectedWgs84,
        ),
      ),
    );
    if (pickedWgs84 == null) return;

    // 天地图为 WGS84；后端/高德/腾讯为 GCJ-02，因此先转 GCJ 再做逆地理编码与保存
    final gcj = CoordinateTransform.wgs84ToGcj02(
      pickedWgs84.latitude,
      pickedWgs84.longitude,
    );

    setState(() {
      _latitudeGcj = gcj.latitude;
      _longitudeGcj = gcj.longitude;
    });

    await _reverseGeocodeFillAddress();
  }

  Future<void> _reverseGeocodeFillAddress() async {
    if (_latitudeGcj == null || _longitudeGcj == null) return;
    if (_isGeocoding) return;

    setState(() {
      _isGeocoding = true;
    });

    try {
      final resp = await Request.post<Map<String, dynamic>>(
        '/employee/addresses/reverse-geocode',
        body: {'longitude': _longitudeGcj, 'latitude': _latitudeGcj},
        parser: (data) => data as Map<String, dynamic>,
      );

      if (resp.isSuccess && resp.data != null) {
        final addrText = resp.data!['address'] as String?;
        if (addrText != null && addrText.trim().isNotEmpty) {
          _addressController.text = addrText.trim();
        } else {
          _showSnack('地址解析成功，但未返回详细地址');
        }
      } else {
        _showSnack(resp.message.isNotEmpty ? resp.message : '地址解析失败');
      }
    } catch (e) {
      _showSnack('地址解析失败: ${e.toString()}');
    } finally {
      if (mounted) {
        setState(() {
          _isGeocoding = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      extendBodyBehindAppBar: true,
      appBar: AppBar(
        title: Text(
          _isEdit ? '编辑地址' : '新增地址',
          style: const TextStyle(color: Colors.white),
        ),
        backgroundColor: Colors.transparent,
        elevation: 0,
        iconTheme: const IconThemeData(color: Colors.white),
      ),
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [Color(0xFF20CB6B), Color(0xFFEFF7F2)],
          ),
        ),
        child: SafeArea(
          bottom: false,
          child: Padding(
            padding: EdgeInsets.fromLTRB(
              16,
              12,
              16,
              16 + MediaQuery.of(context).padding.bottom,
            ),
            child: Column(
              children: [
                Expanded(
                  child: SingleChildScrollView(
                    child: Form(
                      key: _formKey,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          // 基本信息卡片
                          Container(
                            width: double.infinity,
                            padding: const EdgeInsets.all(16),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(16),
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withOpacity(0.04),
                                  blurRadius: 10,
                                  offset: const Offset(0, 4),
                                ),
                              ],
                            ),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                const Text(
                                  '地址信息',
                                  style: TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.w600,
                                    color: Color(0xFF20253A),
                                  ),
                                ),
                                const SizedBox(height: 16),
                                TextFormField(
                                  controller: _nameController,
                                  decoration: _inputDecoration(
                                    '地址名称',
                                    hintText: '例如：西山区-南亚傣味火锅店',
                                  ),
                                  validator: (v) {
                                    if (v == null || v.trim().isEmpty)
                                      return '请输入地址名称';
                                    return null;
                                  },
                                ),
                                const SizedBox(height: 12),
                                TextFormField(
                                  controller: _contactController,
                                  decoration: _inputDecoration('收货人'),
                                  validator: (v) {
                                    if (v == null || v.trim().isEmpty)
                                      return '请输入收货人';
                                    return null;
                                  },
                                ),
                                const SizedBox(height: 12),
                                TextFormField(
                                  controller: _phoneController,
                                  keyboardType: TextInputType.phone,
                                  decoration: _inputDecoration('收货电话'),
                                  validator: (v) {
                                    if (v == null || v.trim().isEmpty)
                                      return '请输入收货电话';
                                    return null;
                                  },
                                ),
                                const SizedBox(height: 12),
                                // 详细地址：先按钮选点，选点后后端解析回填，再显示可编辑输入框
                                Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      children: [
                                        const Expanded(
                                          child: Text(
                                            '详细地址',
                                            style: TextStyle(
                                              fontSize: 14,
                                              fontWeight: FontWeight.w600,
                                              color: Color(0xFF20253A),
                                            ),
                                          ),
                                        ),
                                        if (_latitudeGcj != null &&
                                            _longitudeGcj != null)
                                          TextButton(
                                            onPressed: _isGeocoding
                                                ? null
                                                : _openPickMap,
                                            child: const Text('重新选择'),
                                          ),
                                      ],
                                    ),
                                    const SizedBox(height: 8),
                                    if (_latitudeGcj == null ||
                                        _longitudeGcj == null)
                                      Center(
                                        child: SizedBox(
                                          width: 260,
                                          child: ElevatedButton.icon(
                                            onPressed: _isGeocoding
                                                ? null
                                                : _openPickMap,
                                            icon: _isGeocoding
                                                ? const SizedBox(
                                                    width: 16,
                                                    height: 16,
                                                    child: CircularProgressIndicator(
                                                      strokeWidth: 2,
                                                      valueColor:
                                                          AlwaysStoppedAnimation<
                                                            Color
                                                          >(Colors.white),
                                                    ),
                                                  )
                                                : const Icon(
                                                    Icons.map,
                                                    size: 18,
                                                    color: Colors.white,
                                                  ),
                                            label: const Text(
                                              '点击选择位置',
                                              style: TextStyle(
                                                fontWeight: FontWeight.w600,
                                              ),
                                            ),
                                            style: ElevatedButton.styleFrom(
                                              backgroundColor: const Color(
                                                0xFF20CB6B,
                                              ),
                                              foregroundColor: Colors.white,
                                              elevation: 0,
                                              padding:
                                                  const EdgeInsets.symmetric(
                                                    vertical: 12,
                                                  ),
                                              shape: RoundedRectangleBorder(
                                                borderRadius:
                                                    BorderRadius.circular(12),
                                              ),
                                            ),
                                          ),
                                        ),
                                      )
                                    else
                                      Container(
                                        width: double.infinity,
                                        padding: const EdgeInsets.symmetric(
                                          horizontal: 12,
                                          vertical: 10,
                                        ),
                                        decoration: BoxDecoration(
                                          color: const Color(
                                            0xFF20CB6B,
                                          ).withOpacity(0.06),
                                          borderRadius: BorderRadius.circular(
                                            12,
                                          ),
                                          border: Border.all(
                                            color: const Color(
                                              0xFF20CB6B,
                                            ).withOpacity(0.15),
                                          ),
                                        ),
                                        child: Row(
                                          children: [
                                            Container(
                                              width: 20,
                                              height: 20,
                                              decoration: const BoxDecoration(
                                                color: Color(0xFF20CB6B),
                                                shape: BoxShape.circle,
                                              ),
                                              child: const Icon(
                                                Icons.check,
                                                color: Colors.white,
                                                size: 14,
                                              ),
                                            ),
                                            const SizedBox(width: 8),
                                            Expanded(
                                              child: Text(
                                                _isGeocoding
                                                    ? '已选点，正在解析地址...'
                                                    : '已选择位置',
                                                style: const TextStyle(
                                                  fontSize: 13,
                                                  color: Color(0xFF40475C),
                                                  fontWeight: FontWeight.w600,
                                                ),
                                              ),
                                            ),
                                            if (_latitudeGcj != null &&
                                                _longitudeGcj != null)
                                              Text(
                                                '${_latitudeGcj!.toStringAsFixed(6)}, ${_longitudeGcj!.toStringAsFixed(6)}',
                                                style: const TextStyle(
                                                  fontSize: 11,
                                                  color: Color(0xFF8C92A4),
                                                ),
                                              ),
                                          ],
                                        ),
                                      ),
                                    if (_latitudeGcj != null &&
                                        _longitudeGcj != null) ...[
                                      const SizedBox(height: 14),
                                      TextFormField(
                                        controller: _addressController,
                                        maxLines: 2,
                                        decoration: _inputDecoration(
                                          '详细地址（可修改）',
                                        ),
                                        validator: (v) {
                                          if (v == null || v.trim().isEmpty)
                                            return '请输入详细地址';
                                          return null;
                                        },
                                      ),
                                    ],
                                  ],
                                ),
                                const SizedBox(height: 16),
                                Container(
                                  padding: const EdgeInsets.all(12),
                                  decoration: BoxDecoration(
                                    color: const Color(0xFFF7F8FA),
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  child: Row(
                                    children: [
                                      Switch(
                                        value: _isDefault,
                                        activeColor: const Color(0xFF20CB6B),
                                        onChanged: (val) {
                                          setState(() {
                                            _isDefault = val;
                                          });
                                        },
                                      ),
                                      const SizedBox(width: 8),
                                      const Expanded(
                                        child: Text(
                                          '设为默认地址',
                                          style: TextStyle(
                                            fontSize: 14,
                                            color: Color(0xFF40475C),
                                            fontWeight: FontWeight.w500,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: 16),

                          // 门头照片卡片
                          Container(
                            width: double.infinity,
                            padding: const EdgeInsets.all(16),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(16),
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withOpacity(0.04),
                                  blurRadius: 10,
                                  offset: const Offset(0, 4),
                                ),
                              ],
                            ),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                const Text(
                                  '门头照片',
                                  style: TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.w600,
                                    color: Color(0xFF20253A),
                                  ),
                                ),
                                const SizedBox(height: 12),
                                Row(
                                  children: [
                                    Expanded(
                                      child: OutlinedButton.icon(
                                        onPressed: _isUploading
                                            ? null
                                            : () => _pickAndUpload(
                                                ImageSource.camera,
                                              ),
                                        icon: const Icon(
                                          Icons.camera_alt,
                                          size: 18,
                                          color: Color(0xFF20CB6B),
                                        ),
                                        label: const Text(
                                          '拍照',
                                          style: TextStyle(
                                            color: Color(0xFF20CB6B),
                                          ),
                                        ),
                                        style: OutlinedButton.styleFrom(
                                          side: const BorderSide(
                                            color: Color(0xFF20CB6B),
                                          ),
                                          shape: RoundedRectangleBorder(
                                            borderRadius: BorderRadius.circular(
                                              12,
                                            ),
                                          ),
                                          padding: const EdgeInsets.symmetric(
                                            vertical: 12,
                                          ),
                                        ),
                                      ),
                                    ),
                                    const SizedBox(width: 10),
                                    Expanded(
                                      child: OutlinedButton.icon(
                                        onPressed: _isUploading
                                            ? null
                                            : () => _pickAndUpload(
                                                ImageSource.gallery,
                                              ),
                                        icon: const Icon(
                                          Icons.photo_library,
                                          size: 18,
                                          color: Color(0xFF20CB6B),
                                        ),
                                        label: const Text(
                                          '相册选择',
                                          style: TextStyle(
                                            color: Color(0xFF20CB6B),
                                          ),
                                        ),
                                        style: OutlinedButton.styleFrom(
                                          side: const BorderSide(
                                            color: Color(0xFF20CB6B),
                                          ),
                                          shape: RoundedRectangleBorder(
                                            borderRadius: BorderRadius.circular(
                                              12,
                                            ),
                                          ),
                                          padding: const EdgeInsets.symmetric(
                                            vertical: 12,
                                          ),
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                                if (_isUploading) ...[
                                  const SizedBox(height: 12),
                                  const LinearProgressIndicator(
                                    minHeight: 2,
                                    valueColor: AlwaysStoppedAnimation<Color>(
                                      Color(0xFF20CB6B),
                                    ),
                                  ),
                                ],
                                if (_avatarUrl.isNotEmpty) ...[
                                  const SizedBox(height: 12),
                                  ClipRRect(
                                    borderRadius: BorderRadius.circular(12),
                                    child: Image.network(
                                      _avatarUrl,
                                      height: 160,
                                      width: double.infinity,
                                      fit: BoxFit.cover,
                                      errorBuilder: (context, error, stack) {
                                        return Container(
                                          height: 160,
                                          width: double.infinity,
                                          decoration: BoxDecoration(
                                            color: Colors.grey.shade200,
                                            borderRadius: BorderRadius.circular(
                                              12,
                                            ),
                                          ),
                                          alignment: Alignment.center,
                                          child: const Text(
                                            '图片加载失败',
                                            style: TextStyle(
                                              color: Color(0xFF8C92A4),
                                            ),
                                          ),
                                        );
                                      },
                                    ),
                                  ),
                                  const SizedBox(height: 10),
                                  Align(
                                    alignment: Alignment.centerRight,
                                    child: TextButton.icon(
                                      onPressed: _isUploading
                                          ? null
                                          : () {
                                              setState(() {
                                                _avatarUrl = '';
                                              });
                                            },
                                      icon: const Icon(
                                        Icons.delete_outline,
                                        size: 18,
                                      ),
                                      label: const Text('移除照片'),
                                      style: TextButton.styleFrom(
                                        foregroundColor: const Color(
                                          0xFF8C92A4,
                                        ),
                                      ),
                                    ),
                                  ),
                                ] else ...[
                                  const SizedBox(height: 8),
                                  const Text(
                                    '建议拍摄门头照片，便于配送员识别。',
                                    style: TextStyle(
                                      fontSize: 12,
                                      color: Color(0xFF8C92A4),
                                    ),
                                  ),
                                ],
                              ],
                            ),
                          ),
                          const SizedBox(height: 24),
                        ],
                      ),
                    ),
                  ),
                ),

                // 底部保存按钮
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _isSaving ? null : _save,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF20CB6B),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      elevation: 0,
                    ),
                    child: _isSaving
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              valueColor: AlwaysStoppedAnimation<Color>(
                                Colors.white,
                              ),
                            ),
                          )
                        : Text(
                            _isEdit ? '保存修改' : '保存地址',
                            style: const TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
