# 本地批量添加分类

写入 **8 个一级分类**。小程序打开任一分类时，左侧固定「全部」，展示该一级下全部商品。

## 运行

```bash
# 若之前建过「抽纸/卷纸…」分组，加 --cleanup-groups 清掉
python3 seed_local_categories.py --cleanup-groups
```

会生成 `tag_category_map.json`（`tagId -> 一级 categoryId`），供商品上传脚本使用。
