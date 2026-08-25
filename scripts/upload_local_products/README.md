# 本地批量上传 products 商品

读取仓库根目录 `products/*/info.json` 与本地 `img_*` 图片，写入本地后端。

## 前置

1. 启动本地 `go_backend` 与 MinIO（**图片去重修复后需重启后端**）
2. 先跑分类种子：

```bash
python3 scripts/seed_local_categories/seed_local_categories.py
```

## 运行

```bash
cd scripts/upload_local_products
python3 upload_local_products.py --dry-run
python3 upload_local_products.py
```

## 修复已导入商品的重复图片

此前因 MinIO 文件名只有秒级时间戳，同秒多图会互相覆盖。修复后端后执行：

```bash
# 重启 go_backend 后再跑
python3 upload_local_products.py --fix-images
```

会用每个商品目录里的本地 `img_*` 重新上传，并更新商品 `images`。

## 说明

- 只用本地 `img_*`，不再回退远程 URL
- 同目录内容相同的图片会去重
- 已导入记录在 `upload_state.json`
