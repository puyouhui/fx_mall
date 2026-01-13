import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:screenshot/screenshot.dart';
import 'package:gal/gal.dart';
import 'package:path_provider/path_provider.dart';
import 'package:share_plus/share_plus.dart';
import 'dart:io';
import '../utils/storage.dart';

/// 待取货预览页面（用于生成取货清单图片）
class PickupPreviewPage extends StatefulWidget {
  final String supplierName;
  final List<dynamic> items;

  const PickupPreviewPage({
    super.key,
    required this.supplierName,
    required this.items,
  });

  @override
  State<PickupPreviewPage> createState() => _PickupPreviewPageState();
}

class _PickupPreviewPageState extends State<PickupPreviewPage> {
  final ScreenshotController _screenshotController = ScreenshotController();
  bool _isSaving = false;
  Map<String, dynamic>? _employeeInfo;

  @override
  void initState() {
    super.initState();
    _loadEmployeeInfo();
  }

  Future<void> _loadEmployeeInfo() async {
    final employeeInfo = await Storage.getEmployeeInfo();
    if (mounted) {
      setState(() {
        _employeeInfo = employeeInfo;
      });
    }
  }

  /// 复制商品信息到剪贴板
  Future<void> _copyGoodsInfo() async {
    try {
      // 获取配送员信息
      final employeeName = _employeeInfo?['name'] as String? ?? '配送员';

      // 获取当前日期
      final now = DateTime.now();
      final dateStr = DateFormat('yyyy-MM-dd').format(now);

      // 统计货物种类和数量
      final totalTypes = widget.items.length;
      final totalQuantity = widget.items.fold<int>(0, (sum, item) {
        final rawItem = item as Map;
        final quantity = (rawItem['quantity'] as num?)?.toInt() ?? 0;
        return sum + quantity;
      });

      // 构建商品列表信息
      final buffer = StringBuffer();
      buffer.writeln(widget.supplierName);
      buffer.writeln('日期：$dateStr');
      buffer.writeln('配送员：$employeeName');
      buffer.writeln('货物种类：$totalTypes 种');
      buffer.writeln('数量统计：$totalQuantity 件');
      buffer.writeln('');
      buffer.writeln('取货列表：');

      for (var i = 0; i < widget.items.length; i++) {
        final rawItem = widget.items[i] as Map;
        final item = Map<String, dynamic>.from(rawItem);
        final productName = item['product_name'] as String? ?? '';
        final specName = item['spec_name'] as String? ?? '';
        final quantity = item['quantity'] as int? ?? 0;

        buffer.write('${i + 1}. $productName');
        if (specName.isNotEmpty) {
          buffer.write(' - $specName');
        }
        buffer.writeln(' × $quantity');
      }

      // 复制到剪贴板
      await Clipboard.setData(ClipboardData(text: buffer.toString()));

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('商品信息已复制到剪贴板'),
            backgroundColor: Colors.green,
            duration: Duration(seconds: 2),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('复制失败: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  /// 保存图片到相册并分享
  Future<void> _saveAndShare() async {
    // 防止重复点击
    if (_isSaving) return;
    if (!mounted) return;

    // 标记为正在保存，不调用 setState
    _isSaving = true;

    try {
      // 等待一小段时间确保UI渲染完成
      await Future.delayed(const Duration(milliseconds: 200));
      if (!mounted) {
        _isSaving = false;
        return;
      }

      // 捕获截图
      final Uint8List? imageBytes = await _screenshotController.capture(
        delay: const Duration(milliseconds: 100),
        pixelRatio: 2.0,
      );

      if (!mounted || imageBytes == null) {
        _isSaving = false;
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('截图失败，请重试'),
              backgroundColor: Colors.orange,
            ),
          );
        }
        return;
      }

      // 获取应用文档目录
      final directory = await getApplicationDocumentsDirectory();
      final timestamp = DateTime.now().millisecondsSinceEpoch;
      final fileName = '待取货_${widget.supplierName}_$timestamp.png';
      final filePath = '${directory.path}/$fileName';

      // 保存文件
      final file = File(filePath);
      await file.writeAsBytes(imageBytes);

      // 保存到相册
      await Gal.putImage(filePath);

      if (!mounted) {
        _isSaving = false;
        return;
      }

      // 重置状态
      _isSaving = false;

      // 延迟一小段时间后打开分享对话框，避免状态冲突
      await Future.delayed(const Duration(milliseconds: 100));

      if (!mounted) return;

      // 打开分享对话框
      final supplierName = widget.supplierName;
      Share.shareXFiles([
        XFile(filePath),
      ], text: '${supplierName}待取货清单').catchError((error) {
        print('分享失败: $error');
        return error;
      });
    } catch (e) {
      _isSaving = false;
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('保存失败: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final employeeName = _employeeInfo?['name'] as String? ?? '配送员';
    final now = DateTime.now();
    final dateStr = DateFormat('yyyy-MM-dd').format(now);

    // 计算总数量
    final totalQuantity = widget.items.fold<int>(0, (sum, item) {
      final rawItem = item as Map;
      final quantity = (rawItem['quantity'] as num?)?.toInt() ?? 0;
      return sum + quantity;
    });

    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: const Text('待取货预览'),
        centerTitle: true,
        backgroundColor: const Color(0xFF20CB6B),
        elevation: 0,
        iconTheme: const IconThemeData(color: Colors.white),
        titleTextStyle: const TextStyle(
          color: Colors.white,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      body: Column(
        children: [
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Screenshot(
                controller: _screenshotController,
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: Container(
                    decoration: BoxDecoration(
                      color: Colors.white,
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.05),
                          blurRadius: 10,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // 头部信息（绿色渐变背景）
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.symmetric(
                            horizontal: 20,
                            vertical: 12,
                          ),
                          decoration: const BoxDecoration(
                            gradient: LinearGradient(
                              begin: Alignment.topLeft,
                              end: Alignment.bottomRight,
                              colors: [Color(0xFF20CB6B), Color(0xFF18B85A)],
                            ),
                          ),
                          child: Column(
                            children: [
                              Text(
                                '待取货（${widget.supplierName}）',
                                style: const TextStyle(
                                  fontSize: 18,
                                  fontWeight: FontWeight.w700,
                                  color: Colors.white,
                                ),
                              ),
                              const SizedBox(height: 6),
                              Text(
                                '日期：$dateStr',
                                style: TextStyle(
                                  fontSize: 13,
                                  color: Colors.white.withOpacity(0.9),
                                ),
                              ),
                            ],
                          ),
                        ),

                        // 配送员信息
                        Padding(
                          padding: const EdgeInsets.all(20),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Container(
                                    padding: const EdgeInsets.all(8),
                                    decoration: BoxDecoration(
                                      color: const Color(
                                        0xFF20CB6B,
                                      ).withOpacity(0.1),
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: const Icon(
                                      Icons.person,
                                      size: 20,
                                      color: Color(0xFF20CB6B),
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        const Text(
                                          '配送员',
                                          style: TextStyle(
                                            fontSize: 14,
                                            color: Color(0xFF8C92A4),
                                          ),
                                        ),
                                        const SizedBox(height: 4),
                                        Text(
                                          employeeName,
                                          style: const TextStyle(
                                            fontSize: 16,
                                            fontWeight: FontWeight.w600,
                                            color: Color(0xFF20253A),
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ),

                        const Divider(height: 1, color: Color(0xFFE5E7F0)),

                        // 商品列表
                        Padding(
                          padding: const EdgeInsets.all(20),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                '取货列表',
                                style: TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.w600,
                                  color: Color(0xFF20253A),
                                ),
                              ),
                              const SizedBox(height: 16),
                              if (widget.items.isEmpty)
                                const Padding(
                                  padding: EdgeInsets.symmetric(vertical: 20),
                                  child: Center(
                                    child: Text(
                                      '暂无商品',
                                      style: TextStyle(
                                        fontSize: 14,
                                        color: Color(0xFF8C92A4),
                                      ),
                                    ),
                                  ),
                                )
                              else
                                ...widget.items.asMap().entries.map((entry) {
                                  final index = entry.key;
                                  final rawItem = entry.value;
                                  final item = Map<String, dynamic>.from(
                                    rawItem as Map,
                                  );
                                  final productName =
                                      (item['product_name'] as String?) ?? '';
                                  final specName =
                                      (item['spec_name'] as String?) ?? '';
                                  final quantity =
                                      (item['quantity'] as int?) ?? 0;
                                  final image =
                                      (item['image'] as String?) ?? '';

                                  return Container(
                                    margin: EdgeInsets.only(
                                      bottom: index < widget.items.length - 1
                                          ? 16
                                          : 0,
                                    ),
                                    padding: const EdgeInsets.all(12),
                                    decoration: BoxDecoration(
                                      color: const Color(0xFFF5F7FA),
                                      borderRadius: BorderRadius.circular(12),
                                      border: Border.all(
                                        color: const Color(0xFFE5E7F0),
                                        width: 1,
                                      ),
                                    ),
                                    child: Row(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.center,
                                      children: [
                                        // 商品图片
                                        ClipRRect(
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                          child: image.isNotEmpty
                                              ? Image.network(
                                                  image,
                                                  width: 80,
                                                  height: 80,
                                                  fit: BoxFit.cover,
                                                  errorBuilder:
                                                      (
                                                        context,
                                                        error,
                                                        stackTrace,
                                                      ) {
                                                        return Container(
                                                          width: 80,
                                                          height: 80,
                                                          color:
                                                              Colors.grey[200],
                                                          child: const Icon(
                                                            Icons
                                                                .image_not_supported,
                                                            color: Colors.grey,
                                                          ),
                                                        );
                                                      },
                                                )
                                              : Container(
                                                  width: 80,
                                                  height: 80,
                                                  decoration: BoxDecoration(
                                                    color: Colors.grey[200],
                                                    borderRadius:
                                                        BorderRadius.circular(
                                                          8,
                                                        ),
                                                  ),
                                                  child: const Icon(
                                                    Icons.image_not_supported,
                                                    color: Colors.grey,
                                                  ),
                                                ),
                                        ),
                                        const SizedBox(width: 12),
                                        // 商品信息
                                        Expanded(
                                          child: Column(
                                            crossAxisAlignment:
                                                CrossAxisAlignment.start,
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              Text(
                                                productName,
                                                style: const TextStyle(
                                                  fontSize: 16,
                                                  fontWeight: FontWeight.w600,
                                                  color: Color(0xFF20253A),
                                                ),
                                                maxLines: 2,
                                                overflow: TextOverflow.ellipsis,
                                              ),
                                              if (specName.isNotEmpty) ...[
                                                const SizedBox(height: 6),
                                                Text(
                                                  '规格：$specName',
                                                  style: const TextStyle(
                                                    fontSize: 15,
                                                    color: Color(0xFF40475C),
                                                  ),
                                                ),
                                              ],
                                            ],
                                          ),
                                        ),
                                        // 数量（右侧居中）
                                        Text(
                                          '$quantity 件',
                                          style: const TextStyle(
                                            fontSize: 18,
                                            fontWeight: FontWeight.w600,
                                            color: Color(0xFF20CB6B),
                                          ),
                                        ),
                                      ],
                                    ),
                                  );
                                }),
                            ],
                          ),
                        ),

                        // 汇总信息
                        Container(
                          padding: const EdgeInsets.all(20),
                          decoration: BoxDecoration(
                            color: Colors.white,
                            border: Border(
                              top: BorderSide(
                                color: const Color(0xFFE5E7F0),
                                width: 1,
                              ),
                            ),
                          ),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              const Text(
                                '总件数',
                                style: TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.w600,
                                  color: Color(0xFF20253A),
                                ),
                              ),
                              Text(
                                '$totalQuantity 件',
                                style: const TextStyle(
                                  fontSize: 20,
                                  fontWeight: FontWeight.bold,
                                  color: Color(0xFF20CB6B),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ),

          // 底部保存按钮
          Container(
            padding: EdgeInsets.fromLTRB(
              16,
              12,
              16,
              12 + MediaQuery.of(context).padding.bottom,
            ),
            decoration: BoxDecoration(
              color: Colors.white,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.05),
                  blurRadius: 10,
                  offset: const Offset(0, -2),
                ),
              ],
            ),
            child: SafeArea(
              top: false,
              child: Row(
                children: [
                  // 复制按钮
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: _copyGoodsInfo,
                      icon: const Icon(Icons.copy, size: 20),
                      label: const Text(
                        '复制',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      style: OutlinedButton.styleFrom(
                        foregroundColor: const Color(0xFF20CB6B),
                        side: const BorderSide(
                          color: Color(0xFF20CB6B),
                          width: 1.5,
                        ),
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  // 保存并分享按钮
                  Expanded(
                    flex: 2,
                    child: ElevatedButton.icon(
                      onPressed: _saveAndShare,
                      icon: _isSaving
                          ? const SizedBox(
                              height: 20,
                              width: 20,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                valueColor: AlwaysStoppedAnimation<Color>(
                                  Colors.white,
                                ),
                              ),
                            )
                          : const Icon(Icons.save, size: 20),
                      label: Text(
                        _isSaving ? '保存中...' : '保存并分享',
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF20CB6B),
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        elevation: 0,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
