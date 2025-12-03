import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:employees_app/utils/request.dart';

/// 新客资料完善页面
class CustomerProfilePage extends StatefulWidget {
  /// 可选的初始用户编号，用于从其它页面跳转时自动加载客户信息
  final String? initialUserCode;

  const CustomerProfilePage({super.key, this.initialUserCode});

  @override
  State<CustomerProfilePage> createState() => _CustomerProfilePageState();
}

class _CustomerProfilePageState extends State<CustomerProfilePage> {
  final _searchController = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  bool _isSearching = false;
  bool _isSaving = false;
  Map<String, dynamic>? _user;
  Map<String, dynamic>? _defaultAddress;
  List<Map<String, dynamic>> _addresses = [];

  // 可编辑字段
  final _nameController = TextEditingController();
  final _phoneController = TextEditingController();
  final _addrContactController = TextEditingController();
  final _addrPhoneController = TextEditingController();
  final _addrAddressController = TextEditingController();

  // 店铺类型（下拉选择，和小程序保持一致）
  final List<String> _storeTypeOptions = <String>[
    '零售店',
    '批发部',
    '商超',
    '餐饮店',
    '其它',
  ];
  String? _selectedStoreType;

  final ImagePicker _imagePicker = ImagePicker();

  // 用户类型：retail（零售） / wholesale（批发）
  String? _userType;

  @override
  void initState() {
    super.initState();
    final code = widget.initialUserCode;
    if (code != null && code.isNotEmpty) {
      // 从其他页面跳转过来时，自动根据编号查询并加载客户信息
      _searchController.text = code;
      // 延迟到首帧后执行，避免与构建过程冲突
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _searchByUserCode();
      });
    }
  }

  @override
  void dispose() {
    _searchController.dispose();
    _nameController.dispose();
    _phoneController.dispose();
    _addrContactController.dispose();
    _addrPhoneController.dispose();
    _addrAddressController.dispose();
    super.dispose();
  }

  Future<void> _searchByUserCode() async {
    final code = _searchController.text.trim();
    if (code.isEmpty) return;

    setState(() {
      _isSearching = true;
      _user = null;
      _defaultAddress = null;
      _addresses = [];
    });

    // 员工端专用接口：通过用户编号查询自己名下的客户
    final response = await Request.get<Map<String, dynamic>>(
      '/employee/sales/customer-by-code',
      queryParams: {'userCode': code},
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final user = response.data!;
      setState(() {
        _user = user;
        _defaultAddress =
            (user['default_address'] as Map<String, dynamic>? ?? {});
        _addresses = (user['addresses'] as List<dynamic>? ?? [])
            .cast<Map<String, dynamic>>();

        _nameController.text = user['name'] as String? ?? '';
        _phoneController.text = user['phone'] as String? ?? '';
        _selectedStoreType = user['store_type'] as String?;
        _userType = user['user_type'] as String?;

        _addrContactController.text =
            _defaultAddress?['contact'] as String? ?? '';
        _addrPhoneController.text = _defaultAddress?['phone'] as String? ?? '';
        _addrAddressController.text =
            _defaultAddress?['address'] as String? ?? '';
      });
    } else {
      setState(() {
        _user = null;
        _defaultAddress = null;
        _addresses = [];
      });

      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(response.message)));
      }
    }

    if (mounted) {
      setState(() {
        _isSearching = false;
      });
    }
  }

  Future<void> _save() async {
    if (_user == null) return;
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _isSaving = true;
    });

    final userId = _user!['id'] as int;

    // 更新用户基础资料（员工端专用接口）
    final userBody = <String, dynamic>{
      'name': _nameController.text.trim(),
      'phone': _phoneController.text.trim(),
      'storeType': _selectedStoreType ?? '',
      'userType': _userType ?? '',
    };

    await Request.put<dynamic>(
      '/employee/sales/customers/$userId/profile',
      body: userBody,
    );

    // 地址改为单独维护，这里不再直接提交地址信息

    if (mounted) {
      setState(() {
        _isSaving = false;
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('资料已保存')));
    }
  }

  Future<void> _setDefaultAddress(int addrId) async {
    if (_user == null) return;

    await Request.put<dynamic>(
      '/employee/sales/addresses/$addrId',
      body: const {'isDefault': true},
    );

    // 重新拉取用户信息以刷新地址列表和默认地址
    await _searchByUserCode();
  }

  Future<void> _showAddressDialog({Map<String, dynamic>? address}) async {
    if (_user == null) return;

    final userId = _user!['id'] as int;
    final nameController = TextEditingController(
      text: address?['name'] as String? ?? '',
    );
    final contactController = TextEditingController(
      text: address?['contact'] as String? ?? '',
    );
    final phoneController = TextEditingController(
      text: address?['phone'] as String? ?? '',
    );
    final addrController = TextEditingController(
      text: address?['address'] as String? ?? '',
    );
    final avatarController = TextEditingController(
      text: address?['avatar'] as String? ?? '',
    );
    String avatarUrl = avatarController.text;
    // 如果是新增地址且用户没有地址，默认选中"默认地址"
    bool isDefault = address != null
        ? (address['is_default'] as bool?) ?? false
        : _addresses.isEmpty;
    bool isUploading = false;

    await showDialog<void>(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (ctx, setStateDialog) {
            return Dialog(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(20),
              ),
              child: Container(
                constraints: const BoxConstraints(maxWidth: 400),
                padding: const EdgeInsets.all(20),
                child: SingleChildScrollView(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        address == null ? '新增地址' : '编辑地址',
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF20253A),
                        ),
                      ),
                      const SizedBox(height: 20),
                      TextField(
                        controller: nameController,
                        decoration: InputDecoration(
                          labelText: '地址名称',
                          hintText: '例如：西山区-南亚傣味火锅店',
                          filled: true,
                          fillColor: const Color(0xFFF7F8FA),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(12),
                            borderSide: BorderSide.none,
                          ),
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 12,
                          ),
                        ),
                      ),
                      const SizedBox(height: 12),
                      TextField(
                        controller: contactController,
                        decoration: InputDecoration(
                          labelText: '收货人',
                          filled: true,
                          fillColor: const Color(0xFFF7F8FA),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(12),
                            borderSide: BorderSide.none,
                          ),
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 12,
                          ),
                        ),
                      ),
                      const SizedBox(height: 12),
                      TextField(
                        controller: phoneController,
                        keyboardType: TextInputType.phone,
                        decoration: InputDecoration(
                          labelText: '收货电话',
                          filled: true,
                          fillColor: const Color(0xFFF7F8FA),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(12),
                            borderSide: BorderSide.none,
                          ),
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 12,
                          ),
                        ),
                      ),
                      const SizedBox(height: 12),
                      TextField(
                        controller: addrController,
                        maxLines: 2,
                        decoration: InputDecoration(
                          labelText: '详细地址',
                          filled: true,
                          fillColor: const Color(0xFFF7F8FA),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(12),
                            borderSide: BorderSide.none,
                          ),
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 12,
                          ),
                        ),
                      ),
                      const SizedBox(height: 16),
                      const Text(
                        '门头照片',
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF20253A),
                        ),
                      ),
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          Expanded(
                            child: OutlinedButton.icon(
                              onPressed: isUploading
                                  ? null
                                  : () async {
                                      final picked = await _imagePicker
                                          .pickImage(
                                            source: ImageSource.camera,
                                          );
                                      if (picked == null) return;

                                      setStateDialog(() {
                                        isUploading = true;
                                      });

                                      final uploadResp =
                                          await Request.uploadFile(
                                            '/employee/upload/address-avatar',
                                            File(picked.path),
                                          );

                                      setStateDialog(() {
                                        isUploading = false;
                                      });

                                      if (uploadResp.isSuccess &&
                                          uploadResp.data != null) {
                                        final data = uploadResp.data!;
                                        final url =
                                            (data['avatar'] ?? data['imageUrl'])
                                                as String;
                                        setStateDialog(() {
                                          avatarUrl = url;
                                          avatarController.text = url;
                                        });
                                      } else {
                                        if (context.mounted) {
                                          ScaffoldMessenger.of(
                                            context,
                                          ).showSnackBar(
                                            SnackBar(
                                              content: Text(
                                                uploadResp.message.isNotEmpty
                                                    ? uploadResp.message
                                                    : '图片上传失败',
                                              ),
                                            ),
                                          );
                                        }
                                      }
                                    },
                              icon: const Icon(
                                Icons.camera_alt,
                                size: 18,
                                color: Color(0xFF20CB6B),
                              ),
                              label: const Text(
                                '拍照',
                                style: TextStyle(color: Color(0xFF20CB6B)),
                              ),
                              style: OutlinedButton.styleFrom(
                                side: const BorderSide(
                                  color: Color(0xFF20CB6B),
                                ),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                padding: const EdgeInsets.symmetric(
                                  vertical: 12,
                                ),
                              ),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: OutlinedButton.icon(
                              onPressed: isUploading
                                  ? null
                                  : () async {
                                      final picked = await _imagePicker
                                          .pickImage(
                                            source: ImageSource.gallery,
                                          );
                                      if (picked == null) return;

                                      setStateDialog(() {
                                        isUploading = true;
                                      });

                                      final uploadResp =
                                          await Request.uploadFile(
                                            '/employee/upload/address-avatar',
                                            File(picked.path),
                                          );

                                      setStateDialog(() {
                                        isUploading = false;
                                      });

                                      if (uploadResp.isSuccess &&
                                          uploadResp.data != null) {
                                        final data = uploadResp.data!;
                                        final url =
                                            (data['avatar'] ?? data['imageUrl'])
                                                as String;
                                        setStateDialog(() {
                                          avatarUrl = url;
                                          avatarController.text = url;
                                        });
                                      } else {
                                        if (context.mounted) {
                                          ScaffoldMessenger.of(
                                            context,
                                          ).showSnackBar(
                                            SnackBar(
                                              content: Text(
                                                uploadResp.message.isNotEmpty
                                                    ? uploadResp.message
                                                    : '图片上传失败',
                                              ),
                                            ),
                                          );
                                        }
                                      }
                                    },
                              icon: const Icon(
                                Icons.photo_library,
                                size: 18,
                                color: Color(0xFF20CB6B),
                              ),
                              label: const Text(
                                '相册选择',
                                style: TextStyle(color: Color(0xFF20CB6B)),
                              ),
                              style: OutlinedButton.styleFrom(
                                side: const BorderSide(
                                  color: Color(0xFF20CB6B),
                                ),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                padding: const EdgeInsets.symmetric(
                                  vertical: 12,
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                      if (isUploading) ...[
                        const SizedBox(height: 12),
                        const LinearProgressIndicator(
                          minHeight: 2,
                          valueColor: AlwaysStoppedAnimation<Color>(
                            Color(0xFF20CB6B),
                          ),
                        ),
                      ],
                      if (avatarUrl.isNotEmpty) ...[
                        const SizedBox(height: 12),
                        ClipRRect(
                          borderRadius: BorderRadius.circular(12),
                          child: Image.network(
                            avatarUrl,
                            height: 100,
                            fit: BoxFit.cover,
                          ),
                        ),
                      ],
                      const SizedBox(height: 16),
                      Row(
                        children: [
                          Checkbox(
                            value: isDefault,
                            activeColor: const Color(0xFF20CB6B),
                            onChanged: (val) {
                              setStateDialog(() {
                                isDefault = val ?? false;
                              });
                            },
                          ),
                          const Text(
                            '设为默认地址',
                            style: TextStyle(
                              fontSize: 14,
                              color: Color(0xFF40475C),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 20),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          TextButton(
                            onPressed: () => Navigator.of(ctx).pop(),
                            style: TextButton.styleFrom(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 20,
                                vertical: 12,
                              ),
                            ),
                            child: const Text(
                              '取消',
                              style: TextStyle(color: Color(0xFF8C92A4)),
                            ),
                          ),
                          const SizedBox(width: 12),
                          ElevatedButton(
                            onPressed: () async {
                              // 验证必填字段
                              if (contactController.text.trim().isEmpty) {
                                ScaffoldMessenger.of(ctx).showSnackBar(
                                  const SnackBar(content: Text('请输入收货人')),
                                );
                                return;
                              }
                              if (phoneController.text.trim().isEmpty) {
                                ScaffoldMessenger.of(ctx).showSnackBar(
                                  const SnackBar(content: Text('请输入收货电话')),
                                );
                                return;
                              }
                              if (addrController.text.trim().isEmpty) {
                                ScaffoldMessenger.of(ctx).showSnackBar(
                                  const SnackBar(content: Text('请输入详细地址')),
                                );
                                return;
                              }

                              // 先保存外层的客户基础信息（姓名、电话、店铺类型、客户类型）
                              // 确保新增/编辑地址后，再次查询不会把已填写的基础资料清空
                              try {
                                if (_user != null) {
                                  final profileBody = <String, dynamic>{
                                    'name': _nameController.text.trim(),
                                    'phone': _phoneController.text.trim(),
                                    'storeType': _selectedStoreType ?? '',
                                    'userType': _userType ?? '',
                                  };
                                  final profileResp = await Request.put<dynamic>(
                                    '/employee/sales/customers/$userId/profile',
                                    body: profileBody,
                                  );
                                  if (!profileResp.isSuccess) {
                                    if (mounted) {
                                      ScaffoldMessenger.of(ctx).showSnackBar(
                                        SnackBar(
                                          content: Text(
                                            profileResp.message.isNotEmpty
                                                ? profileResp.message
                                                : '基础资料保存失败',
                                          ),
                                        ),
                                      );
                                    }
                                    return;
                                  }
                                }
                              } catch (e) {
                                if (mounted) {
                                  ScaffoldMessenger.of(ctx).showSnackBar(
                                    SnackBar(
                                      content: Text(
                                        '基础资料保存失败: ${e.toString()}',
                                      ),
                                    ),
                                  );
                                }
                                return;
                              }

                              final body = <String, dynamic>{
                                'name': nameController.text.trim(),
                                'contact': contactController.text.trim(),
                                'phone': phoneController.text.trim(),
                                'address': addrController.text.trim(),
                                'avatar': avatarController.text.trim(),
                                'isDefault': isDefault,
                              };

                              try {
                                if (address == null) {
                                  final response = await Request.post<dynamic>(
                                    '/employee/sales/customers/$userId/addresses',
                                    body: body,
                                  );
                                  if (!response.isSuccess) {
                                    if (mounted) {
                                      ScaffoldMessenger.of(ctx).showSnackBar(
                                        SnackBar(
                                          content: Text(
                                            response.message.isNotEmpty
                                                ? response.message
                                                : '地址保存失败',
                                          ),
                                        ),
                                      );
                                    }
                                    return;
                                  }
                                } else {
                                  final addrId = address['id'] as int;
                                  final response = await Request.put<dynamic>(
                                    '/employee/sales/addresses/$addrId',
                                    body: body,
                                  );
                                  if (!response.isSuccess) {
                                    if (mounted) {
                                      ScaffoldMessenger.of(ctx).showSnackBar(
                                        SnackBar(
                                          content: Text(
                                            response.message.isNotEmpty
                                                ? response.message
                                                : '地址更新失败',
                                          ),
                                        ),
                                      );
                                    }
                                    return;
                                  }
                                }

                                if (mounted) {
                                  Navigator.of(ctx).pop();
                                  // 重新查询用户信息以刷新地址列表
                                  await _searchByUserCode();
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    SnackBar(
                                      content: Text(
                                        address == null ? '地址新增成功' : '地址更新成功',
                                      ),
                                    ),
                                  );
                                }
                              } catch (e) {
                                if (mounted) {
                                  ScaffoldMessenger.of(ctx).showSnackBar(
                                    SnackBar(
                                      content: Text('操作失败: ${e.toString()}'),
                                    ),
                                  );
                                }
                              }
                            },
                            style: ElevatedButton.styleFrom(
                              backgroundColor: const Color(0xFF20CB6B),
                              foregroundColor: Colors.white,
                              padding: const EdgeInsets.symmetric(
                                horizontal: 24,
                                vertical: 12,
                              ),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(12),
                              ),
                              elevation: 0,
                            ),
                            child: const Text('保存'),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildAddressList() {
    if (_addresses.isEmpty) {
      return Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(vertical: 20),
        decoration: BoxDecoration(
          color: const Color(0xFFF7F8FA),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Center(
          child: Text(
            '当前暂无地址，请点击"新增地址"添加。',
            style: const TextStyle(fontSize: 13, color: Color(0xFF8C92A4)),
          ),
        ),
      );
    }

    return Column(
      children: _addresses.map((addr) {
        final isDefault = (addr['is_default'] as bool?) ?? false;
        final name = addr['name'] as String? ?? '';
        final contact = addr['contact'] as String? ?? '';
        final phone = addr['phone'] as String? ?? '';
        final address = addr['address'] as String? ?? '';
        final avatar = addr['avatar'] as String? ?? '';

        return Container(
          margin: const EdgeInsets.only(bottom: 12),
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: const Color(0xFFF7F8FA),
            borderRadius: BorderRadius.circular(12),
            border: isDefault
                ? Border.all(
                    color: const Color(0xFF20CB6B).withOpacity(0.3),
                    width: 1.5,
                  )
                : null,
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (avatar.isNotEmpty)
                    ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: Image.network(
                        avatar,
                        width: 50,
                        height: 50,
                        fit: BoxFit.cover,
                        errorBuilder: (context, error, stack) {
                          return Container(
                            width: 50,
                            height: 50,
                            decoration: BoxDecoration(
                              color: Colors.grey.shade200,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            alignment: Alignment.center,
                            child: const Icon(
                              Icons.store_mall_directory,
                              size: 24,
                              color: Colors.grey,
                            ),
                          );
                        },
                      ),
                    ),
                  if (avatar.isNotEmpty) const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // 地址名称（如果有）
                        if (name.isNotEmpty) ...[
                          Row(
                            children: [
                              Expanded(
                                child: Text(
                                  name,
                                  style: const TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.w600,
                                    color: Color(0xFF20253A),
                                  ),
                                ),
                              ),
                              if (isDefault)
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 8,
                                    vertical: 4,
                                  ),
                                  decoration: BoxDecoration(
                                    color: const Color(
                                      0xFF20CB6B,
                                    ).withOpacity(0.1),
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: const Text(
                                    '默认地址',
                                    style: TextStyle(
                                      fontSize: 11,
                                      color: Color(0xFF20CB6B),
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                ),
                            ],
                          ),
                          const SizedBox(height: 6),
                        ] else ...[
                          // 如果没有地址名称，默认地址标签显示在第一行
                          if (isDefault)
                            Row(
                              mainAxisAlignment: MainAxisAlignment.end,
                              children: [
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 8,
                                    vertical: 4,
                                  ),
                                  decoration: BoxDecoration(
                                    color: const Color(
                                      0xFF20CB6B,
                                    ).withOpacity(0.1),
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: const Text(
                                    '默认地址',
                                    style: TextStyle(
                                      fontSize: 11,
                                      color: Color(0xFF20CB6B),
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                        ],
                        // 联系人信息
                        Text(
                          '$contact  $phone',
                          style: const TextStyle(
                            fontSize: 15,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                        ),
                        const SizedBox(height: 6),
                        // 详细地址
                        Text(
                          address,
                          style: const TextStyle(
                            fontSize: 13,
                            color: Color(0xFF8C92A4),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  if (!isDefault)
                    TextButton(
                      onPressed: () {
                        final id = addr['id'] as int;
                        _setDefaultAddress(id);
                      },
                      style: TextButton.styleFrom(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                      ),
                      child: const Text(
                        '设为默认',
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(0xFF20CB6B),
                        ),
                      ),
                    ),
                  const SizedBox(width: 8),
                  TextButton(
                    onPressed: () => _showAddressDialog(address: addr),
                    style: TextButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 6,
                      ),
                    ),
                    child: const Text(
                      '编辑',
                      style: TextStyle(fontSize: 13, color: Color(0xFF40475C)),
                    ),
                  ),
                ],
              ),
            ],
          ),
        );
      }).toList(),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('新客资料完善', style: TextStyle(color: Colors.white)),
        backgroundColor: Colors.transparent,
        elevation: 0,
        iconTheme: const IconThemeData(color: Colors.white),
      ),
      extendBodyBehindAppBar: true,
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [Color(0xFF20CB6B), Color(0xFFEFF7F2)],
          ),
        ),
        child: SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // 查询区域 - 白色卡片
                Container(
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
                  child: Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _searchController,
                          decoration: InputDecoration(
                            labelText: '用户编号',
                            hintText: '请输入客户编号',
                            filled: true,
                            fillColor: const Color(0xFFF7F8FA),
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(12),
                              borderSide: BorderSide.none,
                            ),
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 12,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      ElevatedButton(
                        onPressed: _isSearching ? null : _searchByUserCode,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFF20CB6B),
                          foregroundColor: Colors.white,
                          padding: const EdgeInsets.symmetric(
                            horizontal: 24,
                            vertical: 12,
                          ),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                        ),
                        child: _isSearching
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  valueColor: AlwaysStoppedAnimation<Color>(
                                    Colors.white,
                                  ),
                                ),
                              )
                            : const Text('查询'),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 16),
                if (_user != null)
                  Expanded(
                    child: SingleChildScrollView(
                      child: Form(
                        key: _formKey,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            // 用户资料卡片
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
                                    '用户资料',
                                    style: TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.w600,
                                      color: Color(0xFF20253A),
                                    ),
                                  ),
                                  const SizedBox(height: 16),
                                  TextFormField(
                                    controller: _nameController,
                                    decoration: InputDecoration(
                                      labelText: '姓名',
                                      filled: true,
                                      fillColor: const Color(0xFFF7F8FA),
                                      border: OutlineInputBorder(
                                        borderRadius: BorderRadius.circular(12),
                                        borderSide: BorderSide.none,
                                      ),
                                      contentPadding:
                                          const EdgeInsets.symmetric(
                                            horizontal: 16,
                                            vertical: 12,
                                          ),
                                    ),
                                  ),
                                  const SizedBox(height: 12),
                                  TextFormField(
                                    controller: _phoneController,
                                    keyboardType: TextInputType.phone,
                                    decoration: InputDecoration(
                                      labelText: '手机号',
                                      filled: true,
                                      fillColor: const Color(0xFFF7F8FA),
                                      border: OutlineInputBorder(
                                        borderRadius: BorderRadius.circular(12),
                                        borderSide: BorderSide.none,
                                      ),
                                      contentPadding:
                                          const EdgeInsets.symmetric(
                                            horizontal: 16,
                                            vertical: 12,
                                          ),
                                    ),
                                  ),
                                  const SizedBox(height: 12),
                                  // 店铺类型（与小程序一致）
                                  DropdownButtonFormField<String>(
                                    initialValue:
                                        (_selectedStoreType != null &&
                                            _selectedStoreType!.isNotEmpty)
                                        ? _selectedStoreType
                                        : null,
                                    decoration: InputDecoration(
                                      labelText: '店铺类型（可选）',
                                      filled: true,
                                      fillColor: const Color(0xFFF7F8FA),
                                      border: OutlineInputBorder(
                                        borderRadius: BorderRadius.circular(12),
                                        borderSide: BorderSide.none,
                                      ),
                                      contentPadding:
                                          const EdgeInsets.symmetric(
                                            horizontal: 16,
                                            vertical: 12,
                                          ),
                                    ),
                                    items: _storeTypeOptions
                                        .map(
                                          (v) => DropdownMenuItem<String>(
                                            value: v,
                                            child: Text(v),
                                          ),
                                        )
                                        .toList(),
                                    onChanged: (val) {
                                      setState(() {
                                        _selectedStoreType = val;
                                      });
                                    },
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 16),

                            // 客户类型模块：零售 / 批发 - 独立卡片
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
                                  Row(
                                    children: [
                                      Container(
                                        padding: const EdgeInsets.symmetric(
                                          horizontal: 8,
                                          vertical: 4,
                                        ),
                                        decoration: BoxDecoration(
                                          color: const Color(
                                            0xFFFF5A5F,
                                          ).withOpacity(0.1),
                                          borderRadius: BorderRadius.circular(
                                            6,
                                          ),
                                        ),
                                        child: const Text(
                                          '重要',
                                          style: TextStyle(
                                            fontSize: 11,
                                            color: Color(0xFFFF5A5F),
                                            fontWeight: FontWeight.w600,
                                          ),
                                        ),
                                      ),
                                      const SizedBox(width: 8),
                                      const Text(
                                        '客户类型',
                                        style: TextStyle(
                                          fontSize: 16,
                                          fontWeight: FontWeight.w600,
                                          color: Color(0xFF20253A),
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 12),
                                  Row(
                                    children: [
                                      Expanded(
                                        child: ChoiceChip(
                                          label: const Text('零售客户'),
                                          selected: _userType == 'retail',
                                          selectedColor: const Color(
                                            0xFF20CB6B,
                                          ).withOpacity(0.2),
                                          labelStyle: TextStyle(
                                            color: _userType == 'retail'
                                                ? const Color(0xFF20CB6B)
                                                : const Color(0xFF40475C),
                                            fontWeight: _userType == 'retail'
                                                ? FontWeight.w600
                                                : FontWeight.normal,
                                          ),
                                          onSelected: (selected) {
                                            setState(() {
                                              _userType = selected
                                                  ? 'retail'
                                                  : _userType;
                                            });
                                          },
                                        ),
                                      ),
                                      const SizedBox(width: 12),
                                      Expanded(
                                        child: ChoiceChip(
                                          label: const Text('批发客户'),
                                          selected: _userType == 'wholesale',
                                          selectedColor: const Color(
                                            0xFF20CB6B,
                                          ).withOpacity(0.2),
                                          labelStyle: TextStyle(
                                            color: _userType == 'wholesale'
                                                ? const Color(0xFF20CB6B)
                                                : const Color(0xFF40475C),
                                            fontWeight: _userType == 'wholesale'
                                                ? FontWeight.w600
                                                : FontWeight.normal,
                                          ),
                                          onSelected: (selected) {
                                            setState(() {
                                              _userType = selected
                                                  ? 'wholesale'
                                                  : _userType;
                                            });
                                          },
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 8),
                                  const Text(
                                    '用于区分用户是零售终端还是批发客户，请务必准确选择。',
                                    style: TextStyle(
                                      fontSize: 12,
                                      color: Color(0xFF8C92A4),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 16),

                            // 多地址管理卡片
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
                                  Row(
                                    mainAxisAlignment:
                                        MainAxisAlignment.spaceBetween,
                                    children: [
                                      const Text(
                                        '全部地址',
                                        style: TextStyle(
                                          fontSize: 16,
                                          fontWeight: FontWeight.w600,
                                          color: Color(0xFF20253A),
                                        ),
                                      ),
                                      TextButton.icon(
                                        onPressed: () => _showAddressDialog(),
                                        icon: const Icon(
                                          Icons.add,
                                          size: 18,
                                          color: Color(0xFF20CB6B),
                                        ),
                                        label: const Text(
                                          '新增地址',
                                          style: TextStyle(
                                            color: Color(0xFF20CB6B),
                                          ),
                                        ),
                                        style: TextButton.styleFrom(
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 12,
                                            vertical: 8,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 8),
                                  _buildAddressList(),
                                ],
                              ),
                            ),
                            const SizedBox(height: 24),

                            // 保存按钮
                            SizedBox(
                              width: double.infinity,
                              child: ElevatedButton(
                                onPressed: _isSaving ? null : _save,
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: const Color(0xFF20CB6B),
                                  foregroundColor: Colors.white,
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 16,
                                  ),
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
                                          valueColor:
                                              AlwaysStoppedAnimation<Color>(
                                                Colors.white,
                                              ),
                                        ),
                                      )
                                    : const Text(
                                        '保存资料',
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
