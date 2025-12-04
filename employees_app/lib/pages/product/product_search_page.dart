import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:employees_app/utils/request.dart';
import 'package:employees_app/pages/product/product_detail_page.dart';

/// 产品查询页面
class ProductSearchPage extends StatefulWidget {
  const ProductSearchPage({super.key});

  @override
  State<ProductSearchPage> createState() => _ProductSearchPageState();
}

class _ProductSearchPageState extends State<ProductSearchPage> {
  final _searchController = TextEditingController();
  final _scrollController = ScrollController();

  List<Map<String, dynamic>> _products = [];
  List<Map<String, dynamic>> _categories = []; // 一级分类列表
  int? _selectedParentCategoryId; // 选中的一级分类ID
  int? _selectedCategoryId; // 选中的分类ID（可能是一级或二级）
  List<Map<String, dynamic>> _subCategories = []; // 当前一级分类下的二级分类列表
  int _pageNum = 1;
  final int _pageSize = 20;
  bool _hasMore = true;
  bool _isLoading = false;
  bool _isLoadingMore = false;
  String _searchKeyword = '';

  @override
  void initState() {
    super.initState();
    _loadCategories();
    _loadProducts();
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _searchController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
            _scrollController.position.maxScrollExtent - 200 &&
        !_isLoadingMore &&
        _hasMore) {
      _loadMoreProducts();
    }
  }

  Future<void> _loadCategories() async {
    final response = await Request.get<List<dynamic>>(
      '/categories',
      parser: (data) => data as List<dynamic>,
    );

    if (response.isSuccess && response.data != null) {
      setState(() {
        _categories = response.data!
            .map((item) => item as Map<String, dynamic>)
            .toList();
      });
    }
  }

  Future<void> _loadProducts({bool reset = false}) async {
    if (_isLoading) return;

    if (reset) {
      setState(() {
        _pageNum = 1;
        _products = [];
        _hasMore = true;
      });
    }

    setState(() {
      _isLoading = true;
    });

    final queryParams = <String, String>{
      'pageNum': '$_pageNum',
      'pageSize': '$_pageSize',
    };

    if (_searchKeyword.isNotEmpty) {
      queryParams['keyword'] = _searchKeyword;
    }
    if (_selectedCategoryId != null) {
      queryParams['categoryId'] = '$_selectedCategoryId';
    }

    final response = await Request.get<Map<String, dynamic>>(
      '/employee/sales/products',
      queryParams: queryParams,
      parser: (data) => data as Map<String, dynamic>,
    );

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final data = response.data!;
      final list = (data['list'] as List<dynamic>? ?? [])
          .cast<Map<String, dynamic>>();
      final total = data['total'] as int? ?? 0;

      setState(() {
        if (reset) {
          _products = list;
        } else {
          _products.addAll(list);
        }
        _hasMore = _products.length < total;
        _isLoading = false;
      });
    } else {
      setState(() {
        _isLoading = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(response.message)));
      }
    }
  }

  Future<void> _loadMoreProducts() async {
    if (_isLoadingMore || !_hasMore) return;

    setState(() {
      _isLoadingMore = true;
      _pageNum++;
    });

    await _loadProducts(reset: false);

    setState(() {
      _isLoadingMore = false;
    });
  }

  void _onSearch() {
    final keyword = _searchController.text.trim();
    setState(() {
      _searchKeyword = keyword;
    });
    _loadProducts(reset: true);
  }

  void _onParentCategorySelected(int? parentCategoryId) {
    setState(() {
      _selectedParentCategoryId = parentCategoryId;

      // 获取该一级分类下的二级分类
      if (parentCategoryId != null) {
        final parentCategory = _categories.firstWhere(
          (cat) => (cat['id'] as int) == parentCategoryId,
          orElse: () => {},
        );
        if (parentCategory.isNotEmpty) {
          final children = parentCategory['children'] as List<dynamic>? ?? [];
          _subCategories = children
              .map((item) => item as Map<String, dynamic>)
              .toList();
          // 选择一级分类时，默认选中该一级分类（会查询该分类及其所有子分类）
          _selectedCategoryId = parentCategoryId;
        } else {
          _subCategories = [];
          _selectedCategoryId = parentCategoryId;
        }
      } else {
        // 选择"全部"时，清除所有分类筛选
        _subCategories = [];
        _selectedCategoryId = null;
      }
    });
    _loadProducts(reset: true);
  }

  void _onSubCategorySelected(int? subCategoryId) {
    setState(() {
      if (subCategoryId == null) {
        // 在二级分类中选择"全部"，则查询该一级分类及其所有子分类
        _selectedCategoryId = _selectedParentCategoryId;
      } else {
        // 选择具体的二级分类
        _selectedCategoryId = subCategoryId;
      }
    });
    _loadProducts(reset: true);
  }

  String _getPriceRange(List<dynamic> specs) {
    if (specs.isEmpty) return '暂无价格';
    final prices = <double>[];
    for (var spec in specs) {
      final specMap = spec as Map<String, dynamic>;
      final retailPrice = (specMap['retail_price'] as num?)?.toDouble() ?? 0.0;
      final wholesalePrice =
          (specMap['wholesale_price'] as num?)?.toDouble() ?? 0.0;
      if (retailPrice > 0) prices.add(retailPrice);
      if (wholesalePrice > 0) prices.add(wholesalePrice);
    }
    if (prices.isEmpty) return '暂无价格';
    prices.sort();
    if (prices.first == prices.last) {
      return '¥${prices.first.toStringAsFixed(2)}';
    }
    return '¥${prices.first.toStringAsFixed(2)} - ¥${prices.last.toStringAsFixed(2)}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      extendBody: true, // 让body延伸到系统操作条下方
      appBar: AppBar(
        title: const Text('销售产品查询', style: TextStyle(color: Colors.white)),
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
          bottom: false, // 底部不使用SafeArea，让内容延伸到系统操作条
          child: Column(
            children: [
              // 搜索和分类筛选区域
              Container(
                padding: const EdgeInsets.all(16),
                child: Column(
                  children: [
                    // 搜索框
                    Row(
                      children: [
                        Expanded(
                          child: TextField(
                            controller: _searchController,
                            decoration: InputDecoration(
                              hintText: '搜索商品名称',
                              filled: true,
                              fillColor: Colors.white,
                              border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(12),
                                borderSide: BorderSide.none,
                              ),
                              contentPadding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 12,
                              ),
                              prefixIcon: const Icon(Icons.search),
                            ),
                            onSubmitted: (_) => _onSearch(),
                          ),
                        ),
                        const SizedBox(width: 12),
                        ElevatedButton(
                          onPressed: _onSearch,
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.white,
                            foregroundColor: const Color(0xFF20253A),
                            padding: const EdgeInsets.symmetric(
                              horizontal: 24,
                              vertical: 12,
                            ),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(12),
                            ),
                            elevation: 0,
                          ),
                          child: const Text(
                            '搜索',
                            style: TextStyle(fontWeight: FontWeight.w600),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    // 分类筛选选择器
                    Row(
                      children: [
                        // 一级分类选择器
                        Expanded(
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 12),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: DropdownButton<int?>(
                              value: _selectedParentCategoryId,
                              isExpanded: true,
                              hint: const Text(
                                '选择一级分类',
                                style: TextStyle(
                                  color: Color(0xFF8C92A4),
                                  fontSize: 14,
                                ),
                              ),
                              underline: const SizedBox(),
                              icon: const Icon(
                                Icons.arrow_drop_down,
                                color: Color(0xFF20CB6B),
                              ),
                              dropdownColor: Colors.white,
                              style: const TextStyle(
                                color: Color(0xFF20253A),
                                fontSize: 14,
                              ),
                              itemHeight: 48,
                              items: [
                                const DropdownMenuItem<int?>(
                                  value: null,
                                  child: Text(
                                    '全部分类',
                                    style: TextStyle(color: Color(0xFF20253A)),
                                  ),
                                ),
                                ..._categories.map((category) {
                                  final id = category['id'] as int;
                                  final name =
                                      category['name'] as String? ?? '';
                                  return DropdownMenuItem<int?>(
                                    value: id,
                                    child: Text(
                                      name,
                                      style: const TextStyle(
                                        color: Color(0xFF20253A),
                                      ),
                                    ),
                                  );
                                }),
                              ],
                              onChanged: (value) {
                                _onParentCategorySelected(value);
                              },
                            ),
                          ),
                        ),
                        // 二级分类选择器（当选择了一级分类时显示）
                        if (_selectedParentCategoryId != null &&
                            _subCategories.isNotEmpty) ...[
                          const SizedBox(width: 12),
                          Expanded(
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 12,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: DropdownButton<int?>(
                                value:
                                    _selectedCategoryId ==
                                        _selectedParentCategoryId
                                    ? null
                                    : _selectedCategoryId,
                                isExpanded: true,
                                hint: const Text(
                                  '选择二级分类',
                                  style: TextStyle(
                                    color: Color(0xFF8C92A4),
                                    fontSize: 14,
                                  ),
                                ),
                                underline: const SizedBox(),
                                icon: const Icon(
                                  Icons.arrow_drop_down,
                                  color: Color(0xFF20CB6B),
                                ),
                                dropdownColor: Colors.white,
                                style: const TextStyle(
                                  color: Color(0xFF20253A),
                                  fontSize: 14,
                                ),
                                itemHeight: 48,
                                items: [
                                  const DropdownMenuItem<int?>(
                                    value: null,
                                    child: Text(
                                      '全部',
                                      style: TextStyle(
                                        color: Color(0xFF20253A),
                                      ),
                                    ),
                                  ),
                                  ..._subCategories.map((category) {
                                    final id = category['id'] as int;
                                    final name =
                                        category['name'] as String? ?? '';
                                    return DropdownMenuItem<int?>(
                                      value: id,
                                      child: Text(
                                        name,
                                        style: const TextStyle(
                                          color: Color(0xFF20253A),
                                        ),
                                      ),
                                    );
                                  }),
                                ],
                                onChanged: (value) {
                                  _onSubCategorySelected(value);
                                },
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                  ],
                ),
              ),
              // 商品列表
              Expanded(
                child: _isLoading && _products.isEmpty
                    ? const Center(
                        child: CircularProgressIndicator(
                          valueColor: AlwaysStoppedAnimation<Color>(
                            Colors.white,
                          ),
                        ),
                      )
                    : _products.isEmpty
                    ? Center(
                        child: Text(
                          _searchKeyword.isNotEmpty ||
                                  _selectedCategoryId != null
                              ? '暂无商品'
                              : '暂无商品数据',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 14,
                          ),
                        ),
                      )
                    : ListView.builder(
                        controller: _scrollController,
                        padding: EdgeInsets.fromLTRB(
                          16,
                          0,
                          16,
                          16 +
                              MediaQuery.of(
                                context,
                              ).padding.bottom, // 添加底部安全区域内边距
                        ),
                        itemCount: _products.length + (_hasMore ? 1 : 0),
                        itemBuilder: (context, index) {
                          if (index == _products.length) {
                            return const Padding(
                              padding: EdgeInsets.symmetric(vertical: 16),
                              child: Center(
                                child: SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                    valueColor: AlwaysStoppedAnimation<Color>(
                                      Colors.white,
                                    ),
                                  ),
                                ),
                              ),
                            );
                          }
                          return _buildProductCard(_products[index]);
                        },
                      ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildProductCard(Map<String, dynamic> product) {
    final id = product['id'] as int;
    final name = product['name'] as String? ?? '';
    final description = product['description'] as String? ?? '';
    final images = (product['images'] as List<dynamic>? ?? [])
        .map((e) => e.toString())
        .toList();
    final specs = product['specs'] as List<dynamic>? ?? [];
    final isSpecial = product['is_special'] as bool? ?? false;
    final imageUrl = images.isNotEmpty ? images[0] : '';

    return InkWell(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => ProductDetailPage(productId: id)),
        );
      },
      borderRadius: BorderRadius.circular(16),
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(14),
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
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 商品图片
            ClipRRect(
              borderRadius: BorderRadius.circular(12),
              child: imageUrl.isNotEmpty
                  ? Image.network(
                      imageUrl,
                      width: 100,
                      height: 100,
                      fit: BoxFit.cover,
                      errorBuilder: (context, error, stack) {
                        return Container(
                          width: 100,
                          height: 100,
                          color: Colors.grey.shade200,
                          alignment: Alignment.center,
                          child: const Icon(
                            Icons.image_not_supported,
                            size: 32,
                            color: Colors.grey,
                          ),
                        );
                      },
                    )
                  : Container(
                      width: 100,
                      height: 100,
                      color: Colors.grey.shade200,
                      alignment: Alignment.center,
                      child: const Icon(
                        Icons.image_not_supported,
                        size: 32,
                        color: Colors.grey,
                      ),
                    ),
            ),
            const SizedBox(width: 12),
            // 商品信息
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
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
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      if (isSpecial)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 6,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: const Color(0xFFFF5A5F).withOpacity(0.1),
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: const Text(
                            '精选',
                            style: TextStyle(
                              fontSize: 10,
                              color: Color(0xFFFF5A5F),
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                    ],
                  ),
                  if (description.isNotEmpty) ...[
                    const SizedBox(height: 4),
                    Text(
                      description,
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFF8C92A4),
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                  const SizedBox(height: 6),
                  Row(
                    children: [
                      Text(
                        '${specs.length}个规格',
                        style: const TextStyle(
                          fontSize: 12,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    _getPriceRange(specs),
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: Color(0xFFFF5A5F),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
