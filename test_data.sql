-- 测试数据SQL语句
-- 图片地址统一使用: http://113.44.164.151:9000/selected/product_1763694599.jpg

-- 插入商品测试数据
-- 注意：假设已有分类ID（category_id）和供应商ID（supplier_id），如果没有请先插入分类和供应商数据

-- 商品1：纸巾
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('心相印云感柔肤纸巾', '适合家用，柔软舒适，3层加厚设计', NULL, NULL, 1, 1, 1, 
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"3包装","wholesale_price":25.00,"retail_price":30.00,"description":"≈10元/包"},{"name":"6包装","wholesale_price":45.00,"retail_price":55.00,"description":"≈9.2元/包"},{"name":"12包装","wholesale_price":80.00,"retail_price":100.00,"description":"≈8.3元/包"}]',
 1, NOW(), NOW());

-- 商品2：矿泉水
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('农夫山泉天然矿泉水', '天然弱碱性水，适合日常饮用', NULL, NULL, 1, 1, 0,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"24瓶装","wholesale_price":18.00,"retail_price":24.00,"description":"≈1元/瓶"},{"name":"12瓶装","wholesale_price":10.00,"retail_price":14.00,"description":"≈1.2元/瓶"}]',
 1, NOW(), NOW());

-- 商品3：方便面
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('康师傅红烧牛肉面', '经典口味，方便快捷', NULL, NULL, 1, 1, 1,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"5包装","wholesale_price":12.00,"retail_price":18.00,"description":"≈3.6元/包"},{"name":"12包装","wholesale_price":25.00,"retail_price":38.00,"description":"≈3.2元/包"}]',
 1, NOW(), NOW());

-- 商品4：牛奶
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('蒙牛纯牛奶', '优质奶源，营养丰富', NULL, NULL, 1, 1, 0,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"250ml*12盒","wholesale_price":35.00,"retail_price":48.00,"description":"≈4元/盒"},{"name":"250ml*24盒","wholesale_price":65.00,"retail_price":88.00,"description":"≈3.7元/盒"}]',
 1, NOW(), NOW());

-- 商品5：洗发水
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('海飞丝去屑洗发水', '强效去屑，清爽控油', NULL, NULL, 1, 1, 1,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"400ml","wholesale_price":28.00,"retail_price":38.00,"description":"单瓶装"},{"name":"750ml","wholesale_price":45.00,"retail_price":65.00,"description":"大瓶装"}]',
 1, NOW(), NOW());

-- 商品6：洗衣液
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('蓝月亮洗衣液', '深层洁净，护色增艳', NULL, NULL, 1, 1, 0,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"1kg装","wholesale_price":15.00,"retail_price":22.00,"description":"小包装"},{"name":"3kg装","wholesale_price":38.00,"retail_price":55.00,"description":"大包装"}]',
 1, NOW(), NOW());

-- 商品7：牙膏
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('佳洁士全优7效牙膏', '全面护齿，清新口气', NULL, NULL, 1, 1, 1,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"120g","wholesale_price":8.00,"retail_price":12.00,"description":"单支装"},{"name":"120g*3支","wholesale_price":20.00,"retail_price":32.00,"description":"三支装"}]',
 1, NOW(), NOW());

-- 商品8：面包
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('达利园小面包', '松软香甜，营养早餐', NULL, NULL, 1, 1, 0,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"500g装","wholesale_price":12.00,"retail_price":18.00,"description":"约20个"},{"name":"1kg装","wholesale_price":22.00,"retail_price":32.00,"description":"约40个"}]',
 1, NOW(), NOW());

-- 商品9：饮料
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('可口可乐', '经典碳酸饮料，冰镇更爽', NULL, NULL, 1, 1, 1,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"330ml*24罐","wholesale_price":35.00,"retail_price":48.00,"description":"≈2元/罐"},{"name":"500ml*12瓶","wholesale_price":28.00,"retail_price":38.00,"description":"≈3.2元/瓶"}]',
 1, NOW(), NOW());

-- 商品10：零食
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('乐事薯片', '香脆可口，多种口味', NULL, NULL, 1, 1, 0,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"70g装","wholesale_price":4.50,"retail_price":7.00,"description":"单包"},{"name":"145g装","wholesale_price":8.00,"retail_price":12.00,"description":"大包"}]',
 1, NOW(), NOW());

-- 商品11：食用油
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('金龙鱼调和油', '营养均衡，烹饪首选', NULL, NULL, 1, 1, 1,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"5L装","wholesale_price":45.00,"retail_price":65.00,"description":"大桶装"},{"name":"2.5L装","wholesale_price":25.00,"retail_price":38.00,"description":"中桶装"}]',
 1, NOW(), NOW());

-- 商品12：大米
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('五常大米', '优质东北大米，粒粒饱满', NULL, NULL, 1, 1, 0,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"10kg装","wholesale_price":55.00,"retail_price":78.00,"description":"大包装"},{"name":"5kg装","wholesale_price":30.00,"retail_price":42.00,"description":"中包装"}]',
 1, NOW(), NOW());

-- 商品13：酱油
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('海天生抽', '鲜味十足，烹饪必备', NULL, NULL, 1, 1, 1,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"500ml","wholesale_price":6.00,"retail_price":9.00,"description":"标准装"},{"name":"1L装","wholesale_price":10.00,"retail_price":15.00,"description":"大瓶装"}]',
 1, NOW(), NOW());

-- 商品14：醋
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('山西老陈醋', '传统工艺，酸香醇厚', NULL, NULL, 1, 1, 0,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"500ml","wholesale_price":5.00,"retail_price":8.00,"description":"标准装"},{"name":"1L装","wholesale_price":9.00,"retail_price":14.00,"description":"大瓶装"}]',
 1, NOW(), NOW());

-- 商品15：盐
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('加碘精制盐', '日常调味，健康加碘', NULL, NULL, 1, 1, 1,
 '["http://113.44.164.151:9000/selected/product_1763694599.jpg"]',
 '[{"name":"400g装","wholesale_price":2.00,"retail_price":3.50,"description":"标准装"},{"name":"1kg装","wholesale_price":4.50,"retail_price":7.00,"description":"大包装"}]',
 1, NOW(), NOW());

-- 插入热销产品关联数据（假设商品ID从1开始）
-- 将前6个商品设置为热销产品
INSERT INTO hot_products (product_id, sort, status, created_at, updated_at) VALUES
(1, 1, 1, NOW(), NOW()),
(2, 2, 1, NOW(), NOW()),
(3, 3, 1, NOW(), NOW()),
(4, 4, 1, NOW(), NOW()),
(5, 5, 1, NOW(), NOW()),
(6, 6, 1, NOW(), NOW());

