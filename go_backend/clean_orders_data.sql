-- 清理所有订单相关数据的SQL脚本
-- 注意：此脚本会删除所有订单、订单明细、配送记录、配送日志、路线排序数据和优惠券使用记录
-- 执行前请确保已备份重要数据！

-- 禁用外键检查（临时）
SET FOREIGN_KEY_CHECKS = 0;

-- 1. 删除配送路线排序表数据
TRUNCATE TABLE delivery_route_orders;

-- 2. 删除配送流程日志表数据
TRUNCATE TABLE delivery_logs;

-- 3. 删除配送记录表数据
TRUNCATE TABLE delivery_records;

-- 4. 删除订单明细表数据
TRUNCATE TABLE order_items;

-- 5. 删除订单主表数据
TRUNCATE TABLE orders;

-- 6. 删除优惠券使用记录（已使用的优惠券）
-- 注意：这里只删除已使用的优惠券记录，未使用的优惠券会保留
-- 清理所有已使用的优惠券（包括有订单ID和没有订单ID的，以防数据异常）
DELETE FROM user_coupons WHERE status = 'used';

-- 7. 清理优惠券发放记录表数据（可选，如果需要保留发放记录可以注释掉）
TRUNCATE TABLE coupon_issue_logs;

-- 重新启用外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- 显示清理结果
SELECT 
    'delivery_route_orders' AS table_name, COUNT(*) AS remaining_count FROM delivery_route_orders
UNION ALL
SELECT 
    'delivery_logs', COUNT(*) FROM delivery_logs
UNION ALL
SELECT 
    'delivery_records', COUNT(*) FROM delivery_records
UNION ALL
SELECT 
    'order_items', COUNT(*) FROM order_items
UNION ALL
SELECT 
    'orders', COUNT(*) FROM orders
UNION ALL
SELECT 
    'user_coupons (used)', COUNT(*) FROM user_coupons WHERE status = 'used'
UNION ALL
SELECT 
    'coupon_issue_logs', COUNT(*) FROM coupon_issue_logs;

