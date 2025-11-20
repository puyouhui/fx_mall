-- 使用product_shop数据库
USE product_shop;

-- 向categories表添加icon字段，用于存储分类图标URL
ALTER TABLE categories
ADD COLUMN icon VARCHAR(255) NULL COMMENT '分类图标URL' AFTER updated_at;

-- 查看更新后的表结构
desc categories;