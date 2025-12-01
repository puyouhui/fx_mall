import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:employees_app/utils/request.dart';

/// 新客资料完善页面
class CustomerProfilePage extends StatefulWidget {
  const CustomerProfilePage({super.key});

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
    bool isDefault = (address?['is_default'] as bool?) ?? false;
    bool isUploading = false;

    await showDialog<void>(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (ctx, setStateDialog) {
            return AlertDialog(
              title: Text(address == null ? '新增地址' : '编辑地址'),
              content: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    TextField(
                      controller: contactController,
                      decoration: const InputDecoration(labelText: '收货人'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: phoneController,
                      keyboardType: TextInputType.phone,
                      decoration: const InputDecoration(labelText: '收货电话'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: addrController,
                      maxLines: 2,
                      decoration: const InputDecoration(labelText: '详细地址'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: avatarController,
                      decoration: const InputDecoration(
                        labelText: '门头照片链接（可选）',
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
                                    final picked = await _imagePicker.pickImage(
                                      source: ImageSource.camera,
                                    );
                                    if (picked == null) return;

                                    setStateDialog(() {
                                      isUploading = true;
                                    });

                                    final uploadResp = await Request.uploadFile(
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
                            icon: const Icon(Icons.camera_alt, size: 18),
                            label: const Text('拍照'),
                          ),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: OutlinedButton.icon(
                            onPressed: isUploading
                                ? null
                                : () async {
                                    final picked = await _imagePicker.pickImage(
                                      source: ImageSource.gallery,
                                    );
                                    if (picked == null) return;

                                    setStateDialog(() {
                                      isUploading = true;
                                    });

                                    final uploadResp = await Request.uploadFile(
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
                            icon: const Icon(Icons.photo_library, size: 18),
                            label: const Text('相册选择'),
                          ),
                        ),
                      ],
                    ),
                    if (isUploading) ...[
                      const SizedBox(height: 8),
                      const LinearProgressIndicator(minHeight: 2),
                    ],
                    if (avatarUrl.isNotEmpty) ...[
                      const SizedBox(height: 8),
                      ClipRRect(
                        borderRadius: BorderRadius.circular(6),
                        child: Image.network(
                          avatarUrl,
                          height: 80,
                          fit: BoxFit.cover,
                        ),
                      ),
                    ],
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        Checkbox(
                          value: isDefault,
                          onChanged: (val) {
                            setStateDialog(() {
                              isDefault = val ?? false;
                            });
                          },
                        ),
                        const Text('设为默认地址'),
                      ],
                    ),
                  ],
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(ctx).pop(),
                  child: const Text('取消'),
                ),
                ElevatedButton(
                  onPressed: () async {
                    final body = <String, dynamic>{
                      'contact': contactController.text.trim(),
                      'phone': phoneController.text.trim(),
                      'address': addrController.text.trim(),
                      'avatar': avatarController.text.trim(),
                      'isDefault': isDefault,
                    };

                    if (address == null) {
                      await Request.post<dynamic>(
                        '/employee/sales/customers/$userId/addresses',
                        body: body,
                      );
                    } else {
                      final addrId = address['id'] as int;
                      await Request.put<dynamic>(
                        '/employee/sales/addresses/$addrId',
                        body: body,
                      );
                    }

                    if (mounted) {
                      Navigator.of(ctx).pop();
                      await _searchByUserCode();
                    }
                  },
                  child: const Text('保存'),
                ),
              ],
            );
          },
        );
      },
    );
  }

  Widget _buildAddressList() {
    if (_addresses.isEmpty) {
      return const Text(
        '当前暂无地址，请点击右上角“新增地址”添加。',
        style: TextStyle(fontSize: 13, color: Colors.grey),
      );
    }

    return Column(
      children: _addresses.map((addr) {
        final isDefault = (addr['is_default'] as bool?) ?? false;
        final contact = addr['contact'] as String? ?? '';
        final phone = addr['phone'] as String? ?? '';
        final address = addr['address'] as String? ?? '';
        final avatar = addr['avatar'] as String? ?? '';

        return Card(
          margin: const EdgeInsets.only(bottom: 8),
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    if (avatar.isNotEmpty)
                      ClipRRect(
                        borderRadius: BorderRadius.circular(6),
                        child: Image.network(
                          avatar,
                          width: 46,
                          height: 46,
                          fit: BoxFit.cover,
                          errorBuilder: (context, error, stack) {
                            return Container(
                              width: 46,
                              height: 46,
                              color: Colors.grey.shade200,
                              alignment: Alignment.center,
                              child: const Icon(
                                Icons.store_mall_directory,
                                size: 22,
                                color: Colors.grey,
                              ),
                            );
                          },
                        ),
                      ),
                    if (avatar.isNotEmpty) const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Expanded(
                                child: Text(
                                  '$contact  $phone',
                                  style: const TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                              ),
                              if (isDefault)
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 8,
                                    vertical: 2,
                                  ),
                                  decoration: BoxDecoration(
                                    color: Colors.green.shade50,
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  child: const Text(
                                    '默认地址',
                                    style: TextStyle(
                                      fontSize: 11,
                                      color: Colors.green,
                                    ),
                                  ),
                                ),
                            ],
                          ),
                          const SizedBox(height: 4),
                          Text(
                            address,
                            style: const TextStyle(
                              fontSize: 13,
                              color: Colors.black87,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    if (!isDefault)
                      TextButton(
                        onPressed: () {
                          final id = addr['id'] as int;
                          _setDefaultAddress(id);
                        },
                        child: const Text('设为默认'),
                      ),
                    TextButton(
                      onPressed: () => _showAddressDialog(address: addr),
                      child: const Text('编辑'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        );
      }).toList(),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('新客资料完善')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 查询区域
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _searchController,
                    decoration: const InputDecoration(
                      labelText: '用户编号',
                      hintText: '请输入客户编号',
                      border: OutlineInputBorder(),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                ElevatedButton(
                  onPressed: _isSearching ? null : _searchByUserCode,
                  child: _isSearching
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('查询'),
                ),
              ],
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
                        // 用户资料
                        const Text(
                          '用户资料',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        const SizedBox(height: 8),
                        TextFormField(
                          controller: _nameController,
                          decoration: const InputDecoration(
                            labelText: '姓名',
                            border: OutlineInputBorder(),
                          ),
                        ),
                        const SizedBox(height: 10),
                        TextFormField(
                          controller: _phoneController,
                          keyboardType: TextInputType.phone,
                          decoration: const InputDecoration(
                            labelText: '手机号',
                            border: OutlineInputBorder(),
                          ),
                        ),
                        const SizedBox(height: 10),
                        // 店铺类型（与小程序一致）
                        DropdownButtonFormField<String>(
                          value:
                              (_selectedStoreType != null &&
                                  _selectedStoreType!.isNotEmpty)
                              ? _selectedStoreType
                              : null,
                          decoration: const InputDecoration(
                            labelText: '店铺类型（可选）',
                            border: OutlineInputBorder(),
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
                        const SizedBox(height: 16),

                        // 用户类型模块：零售 / 批发
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.all(12),
                          margin: const EdgeInsets.only(bottom: 20),
                          decoration: BoxDecoration(
                            color: Colors.grey.shade50,
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(
                              color: Colors.grey.shade300,
                              width: 0.8,
                            ),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                '客户类型（很重要）',
                                style: TextStyle(
                                  fontSize: 14,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                              const SizedBox(height: 8),
                              Row(
                                children: [
                                  ChoiceChip(
                                    label: const Text('零售客户'),
                                    selected: _userType == 'retail',
                                    onSelected: (selected) {
                                      setState(() {
                                        _userType = selected
                                            ? 'retail'
                                            : _userType;
                                      });
                                    },
                                  ),
                                  const SizedBox(width: 12),
                                  ChoiceChip(
                                    label: const Text('批发客户'),
                                    selected: _userType == 'wholesale',
                                    onSelected: (selected) {
                                      setState(() {
                                        _userType = selected
                                            ? 'wholesale'
                                            : _userType;
                                      });
                                    },
                                  ),
                                ],
                              ),
                              const SizedBox(height: 6),
                              const Text(
                                '用于区分用户是零售终端还是批发客户，请务必准确选择。',
                                style: TextStyle(
                                  fontSize: 12,
                                  color: Colors.grey,
                                ),
                              ),
                            ],
                          ),
                        ),

                        const SizedBox(height: 4),

                        // 多地址管理
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text(
                              '全部地址',
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            TextButton.icon(
                              onPressed: () => _showAddressDialog(),
                              icon: const Icon(Icons.add, size: 18),
                              label: const Text('新增地址'),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        _buildAddressList(),
                        const SizedBox(height: 24),

                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton(
                            onPressed: _isSaving ? null : _save,
                            style: ElevatedButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 12),
                            ),
                            child: _isSaving
                                ? const SizedBox(
                                    width: 16,
                                    height: 16,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                    ),
                                  )
                                : const Text('保存资料'),
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
    );
  }
}
