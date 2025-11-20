-- 数据清空脚本
-- 此脚本会清空商品选购小程序的所有业务数据，但保留管理员账户信息
-- 注意：执行此脚本前请确认已备份重要数据
-- 执行方式：在MySQL命令行中使用 source 命令运行，或在phpMyAdmin等工具中导入运行

-- 使用product_shop数据库
USE product_shop;

-- 查看当前数据库中的所有表
SHOW TABLES;

-- ===================== 清空业务数据表 =====================
-- 注意：保留 admins 表不做清空操作

-- 1. 清空分类表数据
TRUNCATE TABLE categories;

-- 2. 清空商品表数据
TRUNCATE TABLE products;

-- 3. 清空轮播图表数据
TRUNCATE TABLE carousels;

-- 4. 查找并清空其他可能存在的表（除了admins表）
-- 生成清空其他表的SQL语句（不包括admins表）
SET group_concat_max_len = 1000000;
SELECT CONCAT(
    'TRUNCATE TABLE ',
    GROUP_CONCAT(table_name SEPARATOR '; TRUNCATE TABLE '),
    ';'
) AS truncate_statements
FROM information_schema.tables
WHERE table_schema = 'product_shop' 
  AND table_name != 'admins';

-- ===================== 操作说明 =====================
-- 1. 上述TRUNCATE语句会重置表的自增ID并清空所有数据
-- 2. 如果需要保留自增ID序列，可以将TRUNCATE改为DELETE FROM table_name;
-- 3. 执行完脚本后，您可能需要重新添加基础分类数据以保证系统正常运行
-- 4. 管理员账户数据将被完整保留，无需重新创建管理员账户