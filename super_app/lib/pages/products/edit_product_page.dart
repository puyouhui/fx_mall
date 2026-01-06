import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:super_app/api/products_api.dart';
import 'package:super_app/api/suppliers_api.dart';
import 'package:super_app/models/product.dart';
import 'package:super_app/utils/request.dart';

class EditProductPage extends StatefulWidget {
  final int? productId; // null表示新增，有值表示编辑

  const EditProductPage({super.key, this.productId});

  @override
  State<EditProductPage> createState() => _EditProductPageState();
}

class _EditProductPageState extends State<EditProductPage> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _descriptionController = TextEditingController();
  
  List<Map<String, dynamic>> _categories = [];
  List<Map<String, dynamic>> _suppliers = [];
  List<int> _selectedCategoryPath = []; // 级联分类路径 [parentId, childId]
  int? _selectedSupplierId;
  bool _isSpecial = false;
  List<String> _images = [];
  List<Map<String, dynamic>> _specs = [];
  
  bool _isLoading = false;
  bool _isSaving = false;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  @override
  void dispose() {
    _nameController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);

    try {
      // 加载分类和供应商
      final categoriesResponse = await ProductsApi.getCategories();
      final suppliersResponse = await SuppliersApi.getSuppliers();

      if (!mounted) return;

      if (categoriesResponse.isSuccess && categoriesResponse.data != null) {
        setState(() => _categories = categoriesResponse.data!);
      }

      if (suppliersResponse.isSuccess && suppliersResponse.data != null) {
        setState(() => _suppliers = suppliersResponse.data!);
      }

      // 如果是编辑模式，加载商品数据
      if (widget.productId != null) {
        await _loadProduct();
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('加载数据失败: ${e.toString()}')),
      );
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  Future<void> _loadProduct() async {
    final response = await ProductsApi.getProductDetail(widget.productId!);
    
    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final product = response.data!;
      _nameController.text = product.name;
      _descriptionController.text = product.description ?? '';
      _isSpecial = product.isSpecial;
      _images = List.from(product.images);
      _selectedSupplierId = product.supplierId;
      
      // 转换规格数据
      // 需要从API响应的原始JSON中获取cost和delivery_count，因为ProductSpec模型中没有这些字段
      // 直接使用Request.get获取原始JSON数据
      final rawResponse = await Request.get<Map<String, dynamic>>(
        '/products/${widget.productId}',
        parser: (data) => data as Map<String, dynamic>,
      );
      
      if (rawResponse.isSuccess && rawResponse.data != null) {
        final rawData = rawResponse.data!;
        final rawSpecs = rawData['specs'] as List<dynamic>? ?? [];
        
        // 从原始JSON中提取specs的完整数据（包括cost和delivery_count）
        _specs = rawSpecs.map((rawSpec) {
          final specMap = rawSpec as Map<String, dynamic>;
          return {
            'name': specMap['name'] as String? ?? '',
            'description': specMap['description'] as String? ?? '',
            'retail_price': (specMap['retail_price'] as num?)?.toDouble() ?? 0.0,
            'wholesale_price': (specMap['wholesale_price'] as num?)?.toDouble() ?? 0.0,
            'cost': (specMap['cost'] as num?)?.toDouble() ?? 0.0,
            'delivery_count': (specMap['delivery_count'] as num?)?.toDouble() ?? 1.0,
          };
        }).toList();
      } else {
        // 如果获取原始数据失败，使用product对象的specs数据（不包含cost）
        _specs = product.specs.map((spec) {
          return {
            'name': spec.name,
            'description': spec.description ?? '',
            'retail_price': spec.retailPrice,
            'wholesale_price': spec.wholesalePrice,
            'cost': 0.0, // 默认值，需要用户填写
            'delivery_count': 1.0, // 默认值
          };
        }).toList();
      }
      
      // 尝试从原始API响应中获取cost和delivery_count
      // 由于ProductsApi使用了Product.fromJson，我们需要获取原始JSON
      // 这里我们需要修改API来返回原始JSON或者同时返回解析后的对象

      // 设置分类路径
      if (product.categoryId != null && _categories.isNotEmpty) {
        _setCategoryPath(product.categoryId!);
      }

      setState(() {});
    }
  }

  void _setCategoryPath(int categoryId) {
    // 查找分类路径（一级和二级）
    for (var category in _categories) {
      final categoryIdValue = category['id'] as int;
      if (categoryIdValue == categoryId) {
        _selectedCategoryPath = [categoryId];
        return;
      }
      final children = category['children'] as List<dynamic>? ?? [];
      for (var child in children) {
        final childMap = child as Map<String, dynamic>;
        if (childMap['id'] as int == categoryId) {
          _selectedCategoryPath = [categoryIdValue, categoryId];
          return;
        }
      }
    }
  }

  int? _getFinalCategoryId() {
    if (_selectedCategoryPath.isEmpty) return null;
    // 如果选择了二级分类，使用二级分类ID；否则使用一级分类ID
    return _selectedCategoryPath.length > 1 
        ? _selectedCategoryPath[1]
        : _selectedCategoryPath[0];
  }

  Future<void> _saveProduct() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    if (_selectedCategoryPath.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请选择商品分类')),
      );
      return;
    }

    if (_specs.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请至少添加一个商品规格')),
      );
      return;
    }

    // 验证规格数据（所有字段除了描述都是必填，且成本价必须大于0）
    for (var i = 0; i < _specs.length; i++) {
      final spec = _specs[i];
      final name = spec['name'] as String? ?? '';
      final cost = (spec['cost'] as num?)?.toDouble() ?? 0.0;
      final retailPrice = (spec['retail_price'] as num?)?.toDouble() ?? 0.0;
      final wholesalePrice = (spec['wholesale_price'] as num?)?.toDouble() ?? 0.0;
      final deliveryCount = (spec['delivery_count'] as num?)?.toDouble() ?? 0.0;
      
      if (name.isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('规格${i + 1}：请输入规格名称')),
        );
        return;
      }
      
      if (cost <= 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('规格"$name"：成本价必须大于0，当前值为${cost.toStringAsFixed(2)}')),
        );
        return;
      }
      
      if (retailPrice <= 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('规格"$name"：零售价必须大于0')),
        );
        return;
      }
      
      if (wholesalePrice <= 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('规格"$name"：批发价必须大于0')),
        );
        return;
      }
      
      if (deliveryCount <= 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('规格"$name"：配送计件数必须大于0')),
        );
        return;
      }
    }

    setState(() => _isSaving = true);

    try {
      final productData = {
        'name': _nameController.text.trim(),
        'description': _descriptionController.text.trim(),
        'category_id': _getFinalCategoryId(),
        'supplier_id': _selectedSupplierId,
        'original_price': 0,
        'price': 0,
        'is_special': _isSpecial,
        'images': _images,
        'specs': _specs,
        'status': 1,
      };

      ApiResponse<Product> response;
      if (widget.productId == null) {
        // 新增
        response = await ProductsApi.createProduct(productData);
      } else {
        // 更新
        productData['id'] = widget.productId;
        response = await ProductsApi.updateProduct(widget.productId!, productData);
      }

      if (!mounted) return;

      if (response.isSuccess) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(widget.productId == null ? '创建成功' : '更新成功')),
        );
        Navigator.of(context).pop(true); // 返回true表示成功
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(response.message)),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('保存失败: ${e.toString()}')),
      );
    } finally {
      if (mounted) {
        setState(() => _isSaving = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.productId == null ? '添加商品' : '编辑商品'),
        backgroundColor: const Color(0xFF20CB6B),
        foregroundColor: Colors.white,
      ),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(
                valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF20CB6B)),
              ),
            )
          : Form(
              key: _formKey,
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  // 商品名称
                  TextFormField(
                    controller: _nameController,
                    decoration: InputDecoration(
                      labelText: '商品名称 *',
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: BorderSide(color: Colors.grey.shade300),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: const BorderSide(
                          color: Color(0xFF20CB6B),
                          width: 2,
                        ),
                      ),
                    ),
                    validator: (value) {
                      if (value == null || value.trim().isEmpty) {
                        return '请输入商品名称';
                      }
                      if (value.trim().length < 2 || value.trim().length > 50) {
                        return '商品名称长度在 2 到 50 个字符';
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 16),

                  // 商品描述
                  TextFormField(
                    controller: _descriptionController,
                    decoration: InputDecoration(
                      labelText: '商品描述',
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: BorderSide(color: Colors.grey.shade300),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: const BorderSide(
                          color: Color(0xFF20CB6B),
                          width: 2,
                        ),
                      ),
                    ),
                    maxLines: 3,
                  ),
                  const SizedBox(height: 16),

                  // 分类选择
                  _buildCategorySelector(),
                  const SizedBox(height: 16),

                  // 供应商选择
                  _buildSupplierSelector(),
                  const SizedBox(height: 16),

                  // 精选商品
                  Container(
                    decoration: BoxDecoration(
                      border: Border.all(color: Colors.grey.shade300),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: SwitchListTile(
                      title: const Text(
                        '精选商品',
                        style: TextStyle(fontWeight: FontWeight.w500),
                      ),
                      value: _isSpecial,
                      activeColor: const Color(0xFF20CB6B),
                      onChanged: (value) => setState(() => _isSpecial = value),
                    ),
                  ),
                  const SizedBox(height: 16),

                  // 商品图片
                  _buildImageSection(),
                  const SizedBox(height: 16),

                  // 商品规格
                  _buildSpecsSection(),
                  const SizedBox(height: 24),

                  // 保存按钮
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: _isSaving ? null : _saveProduct,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF20CB6B),
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        elevation: 2,
                      ),
                      child: _isSaving
                          ? const SizedBox(
                              width: 20,
                              height: 20,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                valueColor:
                                    AlwaysStoppedAnimation<Color>(Colors.white),
                              ),
                            )
                          : Text(
                              widget.productId == null ? '创建商品' : '更新商品',
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
    );
  }

  Widget _buildCategorySelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('商品分类 *', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
        const SizedBox(height: 8),
        // 一级分类
        DropdownButtonFormField<int?>(
          value: _selectedCategoryPath.isNotEmpty ? _selectedCategoryPath[0] : null,
          decoration: InputDecoration(
            labelText: '一级分类',
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade300),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(
                color: Color(0xFF20CB6B),
                width: 2,
              ),
            ),
          ),
          items: [
            const DropdownMenuItem<int?>(value: null, child: Text('请选择')),
            ..._categories.map((category) => DropdownMenuItem<int?>(
                  value: category['id'] as int,
                  child: Text(category['name'] as String? ?? ''),
                )),
          ],
          onChanged: (value) {
            setState(() {
              if (value == null) {
                _selectedCategoryPath = [];
              } else {
                final children = _categories
                    .firstWhere((c) => c['id'] == value, orElse: () => {})['children']
                    as List<dynamic>?;
                if (children != null && children.isNotEmpty) {
                  // 有子分类，需要选择二级分类
                  _selectedCategoryPath = [value];
                } else {
                  // 没有子分类，直接使用一级分类
                  _selectedCategoryPath = [value];
                }
              }
            });
          },
        ),
        // 二级分类（如果一级分类有子分类）
        if (_selectedCategoryPath.isNotEmpty) ...[
          const SizedBox(height: 12),
          Builder(
            builder: (context) {
              final parentId = _selectedCategoryPath[0];
              final parentCategory = _categories.firstWhere(
                (c) => c['id'] == parentId,
                orElse: () => {},
              );
              final children = parentCategory['children'] as List<dynamic>? ?? [];
              
              if (children.isEmpty) {
                return const SizedBox.shrink();
              }

              return DropdownButtonFormField<int?>(
                value: _selectedCategoryPath.length > 1 ? _selectedCategoryPath[1] : null,
                decoration: InputDecoration(
                  labelText: '二级分类',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: Colors.grey.shade300),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(
                      color: Color(0xFF20CB6B),
                      width: 2,
                    ),
                  ),
                ),
                items: [
                  const DropdownMenuItem<int?>(value: null, child: Text('请选择')),
                  ...children.map((child) => DropdownMenuItem<int?>(
                        value: child['id'] as int,
                        child: Text(child['name'] as String? ?? ''),
                      )),
                ],
                onChanged: (value) {
                  setState(() {
                    if (value == null) {
                      _selectedCategoryPath = [parentId];
                    } else {
                      _selectedCategoryPath = [parentId, value];
                    }
                  });
                },
              );
            },
          ),
        ],
      ],
    );
  }

  Widget _buildSupplierSelector() {
    return DropdownButtonFormField<int?>(
      value: _selectedSupplierId,
      decoration: InputDecoration(
        labelText: '供应商（可选）',
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.grey.shade300),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(
            color: Color(0xFF20CB6B),
            width: 2,
          ),
        ),
      ),
      items: [
        const DropdownMenuItem<int?>(value: null, child: Text('不选择（使用自营）')),
        ..._suppliers.map((supplier) => DropdownMenuItem<int?>(
              value: supplier['id'] as int,
              child: Text(supplier['name'] as String? ?? ''),
            )),
      ],
      onChanged: (value) => setState(() => _selectedSupplierId = value),
    );
  }

  Widget _buildImageSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          '商品图片',
          style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
        ),
        const SizedBox(height: 12),
        // 图片网格
        if (_images.isNotEmpty)
          GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 3,
              crossAxisSpacing: 12,
              mainAxisSpacing: 12,
              childAspectRatio: 1.0, // 确保是方形
            ),
            itemCount: _images.length,
            itemBuilder: (context, index) {
              return Stack(
                children: [
                  Container(
                    decoration: BoxDecoration(
                      color: const Color(0xFFF5F6FA),
                      border: Border.all(
                        color: Colors.grey.shade300,
                        width: 1,
                      ),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    clipBehavior: Clip.antiAlias,
                    child: AspectRatio(
                      aspectRatio: 1.0,
                      child: Image.network(
                        _images[index],
                        fit: BoxFit.cover,
                        errorBuilder: (context, error, stackTrace) {
                          return Container(
                            color: const Color(0xFFF5F6FA),
                            child: const Icon(
                              Icons.broken_image,
                              color: Color(0xFFB0B4C3),
                              size: 40,
                            ),
                          );
                        },
                        loadingBuilder: (context, child, loadingProgress) {
                          if (loadingProgress == null) return child;
                          return Container(
                            color: const Color(0xFFF5F6FA),
                            child: Center(
                              child: CircularProgressIndicator(
                                value: loadingProgress.expectedTotalBytes != null
                                    ? loadingProgress.cumulativeBytesLoaded /
                                        loadingProgress.expectedTotalBytes!
                                    : null,
                                strokeWidth: 2,
                                valueColor: const AlwaysStoppedAnimation<Color>(
                                  Color(0xFF20CB6B),
                                ),
                              ),
                            ),
                          );
                        },
                      ),
                    ),
                  ),
                  Positioned(
                    top: 6,
                    right: 6,
                    child: GestureDetector(
                      onTap: () {
                        setState(() => _images.removeAt(index));
                      },
                      child: Container(
                        padding: const EdgeInsets.all(4),
                        decoration: BoxDecoration(
                          color: Colors.red,
                          shape: BoxShape.circle,
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.2),
                              blurRadius: 4,
                              offset: const Offset(0, 2),
                            ),
                          ],
                        ),
                        child: const Icon(
                          Icons.close,
                          color: Colors.white,
                          size: 16,
                        ),
                      ),
                    ),
                  ),
                ],
              );
            },
          ),
        const SizedBox(height: 12),
        ElevatedButton.icon(
          onPressed: _showImagePicker,
          icon: const Icon(Icons.add_photo_alternate),
          label: const Text('添加图片'),
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFF20CB6B),
            foregroundColor: Colors.white,
            padding: const EdgeInsets.symmetric(
              horizontal: 20,
              vertical: 12,
            ),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildSpecsSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            const Text('商品规格 *', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
            ElevatedButton.icon(
              onPressed: () => _showAddSpecDialog(),
              icon: const Icon(Icons.add, size: 20),
              label: const Text('添加规格'),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF20CB6B),
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 10,
                ),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        if (_specs.isEmpty)
          const Padding(
            padding: EdgeInsets.all(16),
            child: Center(child: Text('暂无规格，请添加')),
          )
        else
          ...List.generate(_specs.length, (index) {
            final spec = _specs[index];
            return Card(
              margin: const EdgeInsets.only(bottom: 12),
              elevation: 1,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
                side: BorderSide(color: Colors.grey.shade200),
              ),
              child: ListTile(
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
                title: Text(
                  spec['name'] as String? ?? '',
                  style: const TextStyle(
                    fontWeight: FontWeight.w600,
                    fontSize: 15,
                  ),
                ),
                subtitle: Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // 规格描述
                      if ((spec['description'] as String?) != null &&
                          (spec['description'] as String?)!.isNotEmpty) ...[
                        Text(
                          spec['description'] as String? ?? '',
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.grey.shade700,
                          ),
                        ),
                        const SizedBox(height: 6),
                      ],
                      // 价格信息
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Expanded(
                                child: Text(
                                  '成本价: ¥${(spec['cost'] as num?)?.toStringAsFixed(2) ?? '0.00'}',
                                  style: TextStyle(
                                    fontSize: 13,
                                    color: Colors.grey.shade600,
                                  ),
                                ),
                              ),
                              Expanded(
                                child: Text(
                                  '零售价: ¥${(spec['retail_price'] as num?)?.toStringAsFixed(2) ?? '0.00'}',
                                  style: const TextStyle(fontSize: 13),
                                ),
                              ),
                              Expanded(
                                child: Text(
                                  '批发价: ¥${(spec['wholesale_price'] as num?)?.toStringAsFixed(2) ?? '0.00'}',
                                  style: const TextStyle(fontSize: 13),
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                      const SizedBox(height: 4),
                      Text(
                        '配送计件数: ${(spec['delivery_count'] as num?)?.toStringAsFixed(2) ?? '1.00'}',
                        style: TextStyle(
                          fontSize: 13,
                          color: Colors.grey.shade600,
                        ),
                      ),
                    ],
                  ),
                ),
                trailing: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    IconButton(
                      icon: const Icon(Icons.edit, color: Color(0xFF20CB6B)),
                      onPressed: () => _showEditSpecDialog(index),
                      tooltip: '编辑',
                    ),
                    IconButton(
                      icon: const Icon(Icons.delete, color: Colors.red),
                      onPressed: () {
                        setState(() => _specs.removeAt(index));
                      },
                      tooltip: '删除',
                    ),
                  ],
                ),
              ),
            );
          }),
      ],
    );
  }

  void _showAddSpecDialog() {
    _showSpecDialog(-1);
  }

  void _showEditSpecDialog(int index) {
    _showSpecDialog(index);
  }

  void _showSpecDialog(int index) {
    final isEdit = index >= 0;
    final spec = isEdit ? Map<String, dynamic>.from(_specs[index]) : {
      'name': '',
      'description': '',
      'retail_price': 0.0,
      'wholesale_price': 0.0,
      'cost': 0.0,
      'delivery_count': 1.0,
    };

    final nameController = TextEditingController(text: spec['name'] as String? ?? '');
    final descriptionController = TextEditingController(text: spec['description'] as String? ?? '');
    final costController = TextEditingController(
      text: (spec['cost'] as num?)?.toString() ?? '0',
    );
    final retailPriceController = TextEditingController(
      text: (spec['retail_price'] as num?)?.toString() ?? '0',
    );
    final wholesalePriceController = TextEditingController(
      text: (spec['wholesale_price'] as num?)?.toString() ?? '0',
    );
    final deliveryCountController = TextEditingController(
      text: (spec['delivery_count'] as num?)?.toString() ?? '1.0',
    );

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(isEdit ? '编辑规格' : '添加规格'),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: nameController,
                decoration: InputDecoration(
                  labelText: '规格名称 *',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: Colors.grey.shade300),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(
                      color: Color(0xFF20CB6B),
                      width: 2,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: descriptionController,
                decoration: InputDecoration(
                  labelText: '规格描述',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: Colors.grey.shade300),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(
                      color: Color(0xFF20CB6B),
                      width: 2,
                    ),
                  ),
                ),
                maxLines: 2,
              ),
              const SizedBox(height: 12),
              TextField(
                controller: costController,
                decoration: InputDecoration(
                  labelText: '成本价 *',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: Colors.grey.shade300),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(
                      color: Color(0xFF20CB6B),
                      width: 2,
                    ),
                  ),
                  helperText: '用于计算利润',
                ),
                keyboardType: const TextInputType.numberWithOptions(decimal: true),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: retailPriceController,
                decoration: InputDecoration(
                  labelText: '零售价 *',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: Colors.grey.shade300),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(
                      color: Color(0xFF20CB6B),
                      width: 2,
                    ),
                  ),
                ),
                keyboardType: const TextInputType.numberWithOptions(decimal: true),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: wholesalePriceController,
                decoration: InputDecoration(
                  labelText: '批发价 *',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: Colors.grey.shade300),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(
                      color: Color(0xFF20CB6B),
                      width: 2,
                    ),
                  ),
                ),
                keyboardType: const TextInputType.numberWithOptions(decimal: true),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: deliveryCountController,
                decoration: InputDecoration(
                  labelText: '配送计件数 *',
                  hintText: '例如：1.0（1件装），0.1（10包=1件）',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: Colors.grey.shade300),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(
                      color: Color(0xFF20CB6B),
                      width: 2,
                    ),
                  ),
                  helperText: '用于计算配送费，默认1.0',
                ),
                keyboardType: const TextInputType.numberWithOptions(decimal: true),
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () {
              final name = nameController.text.trim();
              final cost = double.tryParse(costController.text) ?? 0.0;
              final retailPrice = double.tryParse(retailPriceController.text) ?? 0.0;
              final wholesalePrice = double.tryParse(wholesalePriceController.text) ?? 0.0;
              final deliveryCount = double.tryParse(deliveryCountController.text) ?? 1.0;

              // 验证规格名称（必填）
              if (name.isEmpty) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('请输入规格名称')),
                );
                return;
              }

              // 验证成本价（必填且必须大于0）
              final costText = costController.text.trim();
              if (costText.isEmpty) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('请输入成本价')),
                );
                return;
              }
              if (cost <= 0) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('成本价必须大于0')),
                );
                return;
              }

              // 验证零售价（必填且必须大于0）
              final retailPriceText = retailPriceController.text.trim();
              if (retailPriceText.isEmpty) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('请输入零售价')),
                );
                return;
              }
              if (retailPrice <= 0) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('零售价必须大于0')),
                );
                return;
              }

              // 验证批发价（必填且必须大于0）
              final wholesalePriceText = wholesalePriceController.text.trim();
              if (wholesalePriceText.isEmpty) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('请输入批发价')),
                );
                return;
              }
              if (wholesalePrice <= 0) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('批发价必须大于0')),
                );
                return;
              }

              // 验证配送计件数（必填且必须大于0）
              final deliveryCountText = deliveryCountController.text.trim();
              if (deliveryCountText.isEmpty) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('请输入配送计件数')),
                );
                return;
              }
              if (deliveryCount <= 0) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('配送计件数必须大于0')),
                );
                return;
              }

              setState(() {
                final newSpec = {
                  'name': name,
                  'description': descriptionController.text.trim(),
                  'cost': cost,
                  'retail_price': retailPrice,
                  'wholesale_price': wholesalePrice,
                  'delivery_count': deliveryCount,
                };

                if (isEdit) {
                  _specs[index] = newSpec;
                } else {
                  _specs.add(newSpec);
                }
              });

              Navigator.of(context).pop();
            },
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  void _showImagePicker() {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.photo_library),
              title: const Text('从相册选择'),
              onTap: () {
                Navigator.pop(context);
                _pickImage(ImageSource.gallery);
              },
            ),
            ListTile(
              leading: const Icon(Icons.camera_alt),
              title: const Text('拍照'),
              onTap: () {
                Navigator.pop(context);
                _pickImage(ImageSource.camera);
              },
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _pickImage(ImageSource source) async {
    try {
      final ImagePicker picker = ImagePicker();
      final XFile? image = await picker.pickImage(
        source: source,
        maxWidth: 1920,
        maxHeight: 1920,
        imageQuality: 85,
      );

      if (image == null) return;

      // 检查文件大小（5MB限制）
      final file = File(image.path);
      final fileSize = await file.length();
      if (fileSize > 5 * 1024 * 1024) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('图片大小不能超过5MB')),
        );
        return;
      }

      // 显示上传进度
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Row(
            children: [
              SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
              SizedBox(width: 12),
              Text('正在上传图片...'),
            ],
          ),
          duration: Duration(seconds: 30),
        ),
      );

      // 上传图片
      final response = await ProductsApi.uploadProductImage(file);

      if (!mounted) return;

      // 关闭进度提示
      ScaffoldMessenger.of(context).hideCurrentSnackBar();

      if (response.isSuccess && response.data != null) {
        setState(() {
          _images.add(response.data!);
        });
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('图片上传成功')),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('上传失败: ${response.message}')),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).hideCurrentSnackBar();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('选择图片失败: ${e.toString()}')),
      );
    }
  }
}

