import 'package:flutter/material.dart';
import 'package:super_app/api/products_api.dart';
import 'package:super_app/models/product.dart';
import 'package:super_app/pages/products/product_detail_page.dart';

class CreateProductPage extends StatefulWidget {
  const CreateProductPage({super.key});

  @override
  State<CreateProductPage> createState() => _CreateProductPageState();
}

class _CreateProductPageState extends State<CreateProductPage> {
  final _searchController = TextEditingController();
  final _scrollController = ScrollController();

  List<Product> _products = [];
  List<Map<String, dynamic>> _categories = [];
  int? _selectedParentCategoryId;
  int? _selectedCategoryId;
  List<Map<String, dynamic>> _subCategories = [];
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
        _hasMore &&
        !_isLoading) {
      _loadMoreProducts();
    }
  }

  Future<void> _loadCategories() async {
    final response = await ProductsApi.getCategories();

    if (response.isSuccess && response.data != null) {
      setState(() {
        _categories = response.data!;
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

    try {
      final response = await ProductsApi.getProducts(
        pageNum: _pageNum,
        pageSize: _pageSize,
        keyword: _searchKeyword.isNotEmpty ? _searchKeyword : null,
        categoryId: _selectedCategoryId,
      );

      if (!mounted) return;

      if (response.isSuccess && response.data != null) {
        final newProducts = response.data!.list;
        final total = response.data!.total;

        setState(() {
          if (reset) {
            _products = newProducts;
          } else {
            _products.addAll(newProducts);
          }
          _hasMore = _products.length < total;
          if (_hasMore && !reset) {
            _pageNum++;
          }
          _isLoading = false;
        });
      } else {
        setState(() {
          _isLoading = false;
        });
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(response.message)),
          );
        }
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('加载失败: ${e.toString()}')),
        );
      }
    }
  }

  Future<void> _loadMoreProducts() async {
    if (_isLoadingMore || !_hasMore) return;

    setState(() {
      _isLoadingMore = true;
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
          _selectedCategoryId = parentCategoryId;
        } else {
          _subCategories = [];
          _selectedCategoryId = parentCategoryId;
        }
      } else {
        _subCategories = [];
        _selectedCategoryId = null;
      }
    });
    _loadProducts(reset: true);
  }

  void _onSubCategorySelected(int? subCategoryId) {
    setState(() {
      if (subCategoryId == null) {
        _selectedCategoryId = _selectedParentCategoryId;
      } else {
        _selectedCategoryId = subCategoryId;
      }
    });
    _loadProducts(reset: true);
  }
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      extendBody: true,
      appBar: AppBar(
        title: const Text('商品管理', style: TextStyle(color: Colors.white)),
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
          bottom: false,
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
                                  final name = category['name'] as String? ?? '';
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
                                value: _selectedCategoryId ==
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
                        : RefreshIndicator(
                            onRefresh: () => _loadProducts(reset: true),
                            child: ListView.builder(
                              controller: _scrollController,
                              padding: EdgeInsets.fromLTRB(
                                16,
                                0,
                                16,
                                16 +
                                    MediaQuery.of(context).padding.bottom,
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
                                          valueColor:
                                              AlwaysStoppedAnimation<Color>(
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
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildProductCard(Product product) {
    final imageUrl = product.images.isNotEmpty ? product.images[0] : '';

    return InkWell(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (_) => ProductDetailPage(productId: product.id),
          ),
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
                          product.name,
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFF20253A),
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      if (product.isSpecial)
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
                  if (product.description != null &&
                      product.description!.isNotEmpty) ...[
                    const SizedBox(height: 4),
                    Text(
                      product.description!,
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
                        '${product.specs.length}个规格',
                        style: const TextStyle(
                          fontSize: 12,
                          color: Color(0xFF8C92A4),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    product.getPriceRange(),
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

