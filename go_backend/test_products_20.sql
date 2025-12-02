-- 生成20条商品测试数据
-- 图片地址统一使用: http://124.223.94.29:9000/fengxing/product_1764294470.jpg
-- 供应商ID: 1 (自营供应商)
-- 每个商品至少2个规格，所有规格都包含成本字段

-- 商品1：心相印云感柔肤纸巾
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('心相印云感柔肤纸巾', '适合家用，柔软舒适，3层加厚设计，亲肤无刺激', NULL, NULL, 9, 1, 1, 
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"3包装","wholesale_price":25.00,"retail_price":30.00,"cost":18.00,"description":"≈10元/包"},{"name":"6包装","wholesale_price":45.00,"retail_price":55.00,"cost":32.00,"description":"≈9.2元/包"},{"name":"12包装","wholesale_price":80.00,"retail_price":100.00,"cost":58.00,"description":"≈8.3元/包"}]',
 1, NOW(), NOW());

-- 商品2：农夫山泉天然矿泉水
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('农夫山泉天然矿泉水', '天然弱碱性水，适合日常饮用，健康纯净', NULL, NULL, 1, 1, 0,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"24瓶装","wholesale_price":18.00,"retail_price":24.00,"cost":12.00,"description":"≈1元/瓶"},{"name":"12瓶装","wholesale_price":10.00,"retail_price":14.00,"cost":6.50,"description":"≈1.2元/瓶"},{"name":"550ml*12瓶","wholesale_price":15.00,"retail_price":20.00,"cost":9.00,"description":"大瓶装"}]',
 1, NOW(), NOW());

-- 商品3：康师傅红烧牛肉面
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('康师傅红烧牛肉面', '经典口味，方便快捷，营养美味', NULL, NULL, 7, 1, 1,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"5包装","wholesale_price":12.00,"retail_price":18.00,"cost":8.00,"description":"≈3.6元/包"},{"name":"12包装","wholesale_price":25.00,"retail_price":38.00,"cost":16.00,"description":"≈3.2元/包"},{"name":"桶装*6桶","wholesale_price":28.00,"retail_price":42.00,"cost":18.00,"description":"桶装方便面"}]',
 1, NOW(), NOW());

-- 商品4：蒙牛纯牛奶
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('蒙牛纯牛奶', '优质奶源，营养丰富，适合全家饮用', NULL, NULL, 1, 1, 0,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"250ml*12盒","wholesale_price":35.00,"retail_price":48.00,"cost":24.00,"description":"≈4元/盒"},{"name":"250ml*24盒","wholesale_price":65.00,"retail_price":88.00,"cost":42.00,"description":"≈3.7元/盒"},{"name":"1L*12盒","wholesale_price":55.00,"retail_price":75.00,"cost":35.00,"description":"大盒装"}]',
 1, NOW(), NOW());

-- 商品5：海飞丝去屑洗发水
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('海飞丝去屑洗发水', '强效去屑，清爽控油，持久留香', NULL, NULL, 9, 1, 1,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"400ml","wholesale_price":28.00,"retail_price":38.00,"cost":18.00,"description":"单瓶装"},{"name":"750ml","wholesale_price":45.00,"retail_price":65.00,"cost":28.00,"description":"大瓶装"},{"name":"200ml*3瓶","wholesale_price":35.00,"retail_price":50.00,"cost":22.00,"description":"三瓶装"}]',
 1, NOW(), NOW());

-- 商品6：蓝月亮洗衣液
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('蓝月亮洗衣液', '深层洁净，护色增艳，温和不伤手', NULL, NULL, 9, 1, 0,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"1kg装","wholesale_price":15.00,"retail_price":22.00,"cost":9.00,"description":"小包装"},{"name":"3kg装","wholesale_price":38.00,"retail_price":55.00,"cost":24.00,"description":"大包装"},{"name":"500g装","wholesale_price":8.00,"retail_price":12.00,"cost":5.00,"description":"试用装"}]',
 1, NOW(), NOW());

-- 商品7：佳洁士全优7效牙膏
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('佳洁士全优7效牙膏', '全面护齿，清新口气，防蛀美白', NULL, NULL, 9, 1, 1,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"120g","wholesale_price":8.00,"retail_price":12.00,"cost":5.00,"description":"单支装"},{"name":"120g*3支","wholesale_price":20.00,"retail_price":32.00,"cost":12.00,"description":"三支装"},{"name":"180g","wholesale_price":12.00,"retail_price":18.00,"cost":7.00,"description":"大支装"}]',
 1, NOW(), NOW());

-- 商品8：达利园小面包
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('达利园小面包', '松软香甜，营养早餐，独立包装', NULL, NULL, 1, 1, 0,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"500g装","wholesale_price":12.00,"retail_price":18.00,"cost":7.00,"description":"约20个"},{"name":"1kg装","wholesale_price":22.00,"retail_price":32.00,"cost":13.00,"description":"约40个"},{"name":"250g装","wholesale_price":6.50,"retail_price":10.00,"cost":4.00,"description":"小包装"}]',
 1, NOW(), NOW());

-- 商品9：可口可乐
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('可口可乐', '经典碳酸饮料，冰镇更爽，畅快解渴', NULL, NULL, 1, 1, 1,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"330ml*24罐","wholesale_price":35.00,"retail_price":48.00,"cost":22.00,"description":"≈2元/罐"},{"name":"500ml*12瓶","wholesale_price":28.00,"retail_price":38.00,"cost":17.00,"description":"≈3.2元/瓶"},{"name":"2L*6瓶","wholesale_price":32.00,"retail_price":45.00,"cost":19.00,"description":"大瓶装"}]',
 1, NOW(), NOW());

-- 商品10：乐事薯片
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('乐事薯片', '香脆可口，多种口味，休闲零食首选', NULL, NULL, 1, 1, 0,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"70g装","wholesale_price":4.50,"retail_price":7.00,"cost":2.80,"description":"单包"},{"name":"145g装","wholesale_price":8.00,"retail_price":12.00,"cost":5.00,"description":"大包"},{"name":"70g*6包","wholesale_price":24.00,"retail_price":38.00,"cost":15.00,"description":"六包装"}]',
 1, NOW(), NOW());

-- 商品11：金龙鱼调和油
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('金龙鱼调和油', '营养均衡，烹饪首选，健康好油', NULL, NULL, 7, 1, 1,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"5L装","wholesale_price":45.00,"retail_price":65.00,"cost":28.00,"description":"大桶装"},{"name":"2.5L装","wholesale_price":25.00,"retail_price":38.00,"cost":15.00,"description":"中桶装"},{"name":"1L装","wholesale_price":12.00,"retail_price":18.00,"cost":7.50,"description":"小桶装"}]',
 1, NOW(), NOW());

-- 商品12：五常大米
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('五常大米', '优质东北大米，粒粒饱满，香甜可口', NULL, NULL, 7, 1, 0,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"10kg装","wholesale_price":55.00,"retail_price":78.00,"cost":35.00,"description":"大包装"},{"name":"5kg装","wholesale_price":30.00,"retail_price":42.00,"cost":19.00,"description":"中包装"},{"name":"2.5kg装","wholesale_price":18.00,"retail_price":25.00,"cost":11.00,"description":"小包装"}]',
 1, NOW(), NOW());

-- 商品13：海天生抽
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('海天生抽', '鲜味十足，烹饪必备，提鲜增香', NULL, NULL, 7, 1, 1,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"500ml","wholesale_price":6.00,"retail_price":9.00,"cost":3.50,"description":"标准装"},{"name":"1L装","wholesale_price":10.00,"retail_price":15.00,"cost":6.00,"description":"大瓶装"},{"name":"250ml*2瓶","wholesale_price":7.00,"retail_price":11.00,"cost":4.50,"description":"双瓶装"}]',
 1, NOW(), NOW());

-- 商品14：山西老陈醋
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('山西老陈醋', '传统工艺，酸香醇厚，调味佳品', NULL, NULL, 7, 1, 0,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"500ml","wholesale_price":5.00,"retail_price":8.00,"cost":3.00,"description":"标准装"},{"name":"1L装","wholesale_price":9.00,"retail_price":14.00,"cost":5.50,"description":"大瓶装"},{"name":"250ml*2瓶","wholesale_price":6.00,"retail_price":10.00,"cost":3.80,"description":"双瓶装"}]',
 1, NOW(), NOW());

-- 商品15：加碘精制盐
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('加碘精制盐', '日常调味，健康加碘，细盐易溶', NULL, NULL, 7, 1, 1,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"400g装","wholesale_price":2.00,"retail_price":3.50,"cost":1.20,"description":"标准装"},{"name":"1kg装","wholesale_price":4.50,"retail_price":7.00,"cost":2.80,"description":"大包装"},{"name":"250g*4袋","wholesale_price":5.00,"retail_price":8.00,"cost":3.20,"description":"四袋装"}]',
 1, NOW(), NOW());

-- 商品16：维达抽纸
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('维达抽纸', '3层加厚，柔韧不易破，家庭必备', NULL, NULL, 9, 1, 0,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"130抽*3包","wholesale_price":18.00,"retail_price":25.00,"cost":12.00,"description":"三包装"},{"name":"130抽*6包","wholesale_price":32.00,"retail_price":45.00,"cost":22.00,"description":"六包装"},{"name":"130抽*12包","wholesale_price":58.00,"retail_price":82.00,"cost":40.00,"description":"十二包装"}]',
 1, NOW(), NOW());

-- 商品17：立白洗衣粉
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('立白洗衣粉', '强效去污，护色增白，经济实惠', NULL, NULL, 9, 1, 1,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"1kg装","wholesale_price":12.00,"retail_price":18.00,"cost":7.50,"description":"小包装"},{"name":"3kg装","wholesale_price":28.00,"retail_price":42.00,"cost":18.00,"description":"大包装"},{"name":"500g装","wholesale_price":6.50,"retail_price":10.00,"cost":4.00,"description":"试用装"}]',
 1, NOW(), NOW());

-- 商品18：飘柔洗发水
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('飘柔洗发水', '柔顺亮泽，修复受损发质，持久留香', NULL, NULL, 9, 1, 0,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"400ml","wholesale_price":26.00,"retail_price":36.00,"cost":16.00,"description":"单瓶装"},{"name":"750ml","wholesale_price":42.00,"retail_price":60.00,"cost":26.00,"description":"大瓶装"},{"name":"200ml*3瓶","wholesale_price":32.00,"retail_price":48.00,"cost":20.00,"description":"三瓶装"}]',
 1, NOW(), NOW());

-- 商品19：舒肤佳香皂
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('舒肤佳香皂', '除菌抑菌，温和洁净，全家适用', NULL, NULL, 9, 1, 1,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"115g*3块","wholesale_price":12.00,"retail_price":18.00,"cost":7.00,"description":"三块装"},{"name":"115g*6块","wholesale_price":20.00,"retail_price":32.00,"cost":12.00,"description":"六块装"},{"name":"115g单块","wholesale_price":4.50,"retail_price":7.00,"cost":2.50,"description":"单块装"}]',
 1, NOW(), NOW());

-- 商品20：清风卷纸
INSERT INTO products (name, description, price, original_price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES
('清风卷纸', '4层加厚，柔韧吸水，经济实惠', NULL, NULL, 9, 1, 0,
 '["http://124.223.94.29:9000/fengxing/product_1764294470.jpg"]',
 '[{"name":"10卷装","wholesale_price":22.00,"retail_price":32.00,"cost":14.00,"description":"十卷装"},{"name":"20卷装","wholesale_price":38.00,"retail_price":58.00,"cost":25.00,"description":"二十卷装"},{"name":"6卷装","wholesale_price":15.00,"retail_price":22.00,"cost":9.50,"description":"六卷装"}]',
 1, NOW(), NOW());









