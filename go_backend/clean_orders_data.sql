-- 清理所有业务数据（保留商品、系统设置、优惠券模板）
-- 注意：此脚本会删除所有用户、订单、配送、员工、销售分成等相关数据
-- 执行前请确保已备份重要数据！

-- 禁用外键检查（临时）
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- 1. 清理销售分成相关数据
-- ============================================
TRUNCATE TABLE sales_commission_monthly_stats;
TRUNCATE TABLE sales_commissions;
TRUNCATE TABLE sales_commission_config;

-- ============================================
-- 2. 清理配送相关数据
-- ============================================
TRUNCATE TABLE employee_location_history;
TRUNCATE TABLE delivery_route_orders;
TRUNCATE TABLE delivery_logs;
TRUNCATE TABLE delivery_records;

-- ============================================
-- 3. 清理订单相关数据
-- ============================================
TRUNCATE TABLE order_items;
TRUNCATE TABLE orders;

-- ============================================
-- 4. 清理用户相关数据
-- ============================================
-- 先清理采购单（有外键依赖）
TRUNCATE TABLE purchase_list_items;
-- 清理用户地址
TRUNCATE TABLE mini_app_addresses;
-- 清理用户优惠券（包括已使用和未使用的）
TRUNCATE TABLE user_coupons;
-- 清理优惠券发放记录
TRUNCATE TABLE coupon_issue_logs;
-- 清理用户表
TRUNCATE TABLE mini_app_users;

-- ============================================
-- 5. 清理员工数据
-- ============================================
TRUNCATE TABLE employees;

-- ============================================
-- 6. 清理管理员数据（可选，如果需要保留管理员可以注释掉）
-- ============================================
-- TRUNCATE TABLE admins;

-- 重新启用外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 显示清理结果
-- ============================================
SELECT 
    'sales_commission_monthly_stats' AS table_name, COUNT(*) AS remaining_count FROM sales_commission_monthly_stats
UNION ALL
SELECT 
    'sales_commissions', COUNT(*) FROM sales_commissions
UNION ALL
SELECT 
    'sales_commission_config', COUNT(*) FROM sales_commission_config
UNION ALL
SELECT 
    'employee_location_history', COUNT(*) FROM employee_location_history
UNION ALL
SELECT 
    'delivery_route_orders', COUNT(*) FROM delivery_route_orders
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
    'purchase_list_items', COUNT(*) FROM purchase_list_items
UNION ALL
SELECT 
    'mini_app_addresses', COUNT(*) FROM mini_app_addresses
UNION ALL
SELECT 
    'user_coupons', COUNT(*) FROM user_coupons
UNION ALL
SELECT 
    'coupon_issue_logs', COUNT(*) FROM coupon_issue_logs
UNION ALL
SELECT 
    'mini_app_users', COUNT(*) FROM mini_app_users
UNION ALL
SELECT 
    'employees', COUNT(*) FROM employees;

-- ============================================
-- 保留的数据表（不会被清理）
-- ============================================
-- products（商品表）
-- product_specs（商品规格表）
-- categories（分类表）
-- system_settings（系统设置表）
-- coupons（优惠券模板表）
-- admins（管理员表，如需清理请取消注释上面的TRUNCATE语句）
