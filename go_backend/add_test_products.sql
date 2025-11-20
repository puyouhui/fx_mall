-- 使用数据库
USE product_shop;

-- 为商品表添加测试数据
INSERT INTO products (name, description, original_price, price, category_id, is_special, images, specs, status, created_at, updated_at)
VALUES 
    ('高级智能手机', '最新款智能手机，具有超强性能和先进摄像系统', 5999.00, 4999.00, 4, true, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "颜色", "value": "黑色"}, {"name": "存储容量", "value": "128GB"}]', 1, NOW(), NOW()),
    ('超薄笔记本电脑', '轻薄便携，高性能处理器，超长续航', 8999.00, 7999.00, 5, true, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "CPU", "value": "Intel i7"}, {"name": "内存", "value": "16GB"}]', 1, NOW(), NOW()),
    ('智能平板电脑', '高清屏幕，多任务处理，适合工作和娱乐', 3299.00, 2799.00, 6, true, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "尺寸", "value": "10.9英寸"}, {"name": "存储", "value": "256GB"}]', 1, NOW(), NOW()),
    ('高端无线耳机', '主动降噪，高保真音质，舒适佩戴', 1999.00, 1599.00, 4, true, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "颜色", "value": "白色"}, {"name": "续航", "value": "30小时"}]', 1, NOW(), NOW()),
    ('智能手表', '健康监测，运动追踪，时尚设计', 2499.00, 1999.00, 4, false, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "材质", "value": "不锈钢"}, {"name": "防水", "value": "50米"}]', 1, NOW(), NOW()),
    ('便携式蓝牙音箱', '360°环绕音效，防水设计，长效续航', 1299.00, 999.00, 4, false, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "颜色", "value": "蓝色"}, {"name": "电池", "value": "20小时"}]', 1, NOW(), NOW()),
    ('机械键盘', '青轴开关，RGB背光，游戏办公两用', 699.00, 499.00, 5, false, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "轴体", "value": "青轴"}, {"name": "布局", "value": "104键"}]', 1, NOW(), NOW()),
    ('游戏鼠标', '高精度传感器，可自定义按键，RGB灯效', 399.00, 299.00, 5, false, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "DPI", "value": "16000"}, {"name": "按键数", "value": "7键"}]', 1, NOW(), NOW()),
    ('高端咖啡机', '意式浓缩，自动奶泡，智能控制', 4999.00, 3999.00, 7, true, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "类型", "value": "全自动"}, {"name": "容量", "value": "1.5L"}]', 1, NOW(), NOW()),
    ('智能扫地机器人', '激光导航，自动充电，智能避障', 2999.00, 2499.00, 9, true, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "导航", "value": "激光导航"}, {"name": "吸力", "value": "3000Pa"}]', 1, NOW(), NOW()),
    ('高级羽绒服', '90%白鸭绒，防风防水，保暖舒适', 1299.00, 899.00, 10, true, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "颜色", "value": "藏青色"}, {"name": "尺码", "value": "L"}]', 1, NOW(), NOW()),
    ('运动鞋', '轻便舒适，减震防滑，适合多种运动', 899.00, 599.00, 10, false, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "颜色", "value": "黑色"}, {"name": "尺码", "value": "42"}]', 1, NOW(), NOW()),
    ('高级护肤套装', '天然成分，深层滋养，适合所有肤质', 1599.00, 1199.00, 11, true, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "类型", "value": "保湿套装"}, {"name": "适用肤质", "value": "所有肤质"}]', 1, NOW(), NOW()),
    ('儿童安全座椅', 'ISOFIX接口，360°旋转，适合0-7岁', 2499.00, 1899.00, 12, false, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "接口", "value": "ISOFIX"}, {"name": "适用年龄", "value": "0-7岁"}]', 1, NOW(), NOW()),
    ('专业摄影相机', '全画幅传感器，4K视频，高速连拍', 12999.00, 10999.00, 4, false, '["http://113.44.164.151:9000/selected/product_1758517912.png"]', '[{"name": "传感器", "value": "全画幅"}, {"name": "像素", "value": "4500万"}]', 1, NOW(), NOW());

-- 为轮播图添加测试数据
INSERT INTO carousels (image, title, link, sort, status, created_at, updated_at)
VALUES 
    ('http://113.44.164.151:9000/selected/product_1758517912.png', '限时特惠 - 智能手机', '/pages/product/detail?id=1', 1, 1, NOW(), NOW()),
    ('http://113.44.164.151:9000/selected/product_1758517912.png', '新品上市 - 笔记本电脑', '/pages/product/detail?id=2', 2, 1, NOW(), NOW()),
    ('http://113.44.164.151:9000/selected/product_1758517912.png', '智能家居特惠', '/pages/category/list?id=9', 3, 1, NOW(), NOW());

-- 显示插入结果
SELECT '成功插入15条商品测试数据和3条轮播图测试数据';