# 批量改价工具

按分类批量调整商品**成本**、**批发价**、**零售价**，带 GUI 界面，通过后台 API 批量请求。

## 功能

- 选择分类：选择一级或二级分类，对该分类下所有商品进行批量改价
- 价格调整：支持成本、批发价、零售价三个维度的加减
  - 正数 = 涨价
  - 负数 = 降价
- 批量请求：自动分页获取该分类下商品，逐个调用 `PUT /admin/products/:id` 更新

## 环境

- Python 3.9+
- 依赖：`requests`（tkinter 为 Python 自带）

## 安装

```bash
cd scripts/batch_price_tool
pip install -r requirements.txt
```

## 运行

```bash
python batch_price_tool.py
```

## 使用步骤

1. **API 地址**：默认 `https://mall.sscchh.com/api_mall/mini`，可按实际后端地址修改
2. **登录**：填写管理员用户名、密码，点击「登录」
3. **选择分类**：登录成功后点击「刷新分类」，在下拉框中选择要改价的分类
4. **填写调价金额**：
   - 成本调整：如 `0.5` 表示每个规格成本+0.5 元，`-0.3` 表示-0.3 元
   - 批发价调整：同上
   - 零售价调整：同上
5. **开始批量改价**：点击按钮执行，执行日志会显示每个商品的更新结果

## 接口说明

- 登录：`POST /admin/login`，Body: `{username, password}`
- 分类：`GET /admin/categories`，需登录
- 商品：`GET /admin/products?categoryId=X&pageNum=1&pageSize=200`，需登录
- 更新商品：`PUT /admin/products/:id`，Body: 完整商品对象（含修改后的 specs），需登录

## 注意事项

- 价格调整后若小于 0，会自动置为 0
- 每个商品有多个规格时，所有规格均会应用相同的调价
- 无规格的商品会被跳过
- 建议先在测试环境验证后再在生产环境使用
