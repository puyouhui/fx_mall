import 'package:employees_app/pages/customer/customer_address_edit_page.dart';
import 'package:flutter/material.dart';
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
        // 确保店铺类型在选项列表中，否则设为 null
        final storeType = user['store_type'] as String?;
        _selectedStoreType = (storeType != null && 
            _storeTypeOptions.contains(storeType)) 
            ? storeType 
            : null;
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

  Future<void> _openAddressEditPage({Map<String, dynamic>? address}) async {
    if (_user == null) return;

    final userId = _user!['id'] as int;
    final initialIsDefault = address == null
        ? _addresses.isEmpty
        : ((address['is_default'] as bool?) ?? false);

    final profileDraft = <String, dynamic>{
      'name': _nameController.text.trim(),
      'phone': _phoneController.text.trim(),
      'storeType': _selectedStoreType ?? '',
      'userType': _userType ?? '',
    };

    final changed = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => CustomerAddressEditPage(
          userId: userId,
          address: address,
          profileDraft: profileDraft,
          defaultSelected: initialIsDefault,
        ),
      ),
    );

    if (changed == true) {
      await _searchByUserCode();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(address == null ? '地址新增成功' : '地址更新成功')),
        );
      }
    }
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
                    onPressed: () => _openAddressEditPage(address: addr),
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
      extendBodyBehindAppBar: true, // 让背景延伸到AppBar下方
      appBar: AppBar(
        title: const Text(
          '新客资料完善',
          style: TextStyle(color: Colors.white),
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
          bottom: false, // 底部不使用SafeArea，让内容延伸到系统操作条
          child: Padding(
            padding: EdgeInsets.fromLTRB(
              16,
              12,
              16,
              16 + MediaQuery.of(context).padding.bottom, // 添加底部安全区域内边距
            ),
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
                                    value: _selectedStoreType,
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
                                        onPressed: () => _openAddressEditPage(),
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
