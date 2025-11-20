-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS product_shop DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE product_shop;

-- 创建管理员表
CREATE TABLE IF NOT EXISTS admins (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员表';

-- 创建分类表
CREATE TABLE IF NOT EXISTS categories (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL COMMENT '分类名称',
    parent_id INT DEFAULT 0 COMMENT '父分类ID，0表示一级分类',
    sort INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    icon VARCHAR(255) NULL COMMENT '分类图标URL',
    UNIQUE KEY uk_name_parent_id (name, parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品分类表';

-- 创建商品表
CREATE TABLE IF NOT EXISTS products (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '商品名称',
    description TEXT COMMENT '商品描述',
    price DECIMAL(10,2) NOT NULL COMMENT '商品价格',
    category_id INT NOT NULL COMMENT '分类ID',
    is_special TINYINT DEFAULT 0 COMMENT '是否特价：1-是，0-否',
    images TEXT COMMENT '商品图片（JSON格式）',
    specs TEXT COMMENT '商品规格（JSON格式）',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    KEY idx_category_id (category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品表';

-- 创建轮播图表
CREATE TABLE IF NOT EXISTS carousels (
    id INT PRIMARY KEY AUTO_INCREMENT,
    image VARCHAR(255) NOT NULL COMMENT '轮播图图片地址',
    title VARCHAR(255) DEFAULT '' COMMENT '轮播图标题',
    link VARCHAR(255) DEFAULT '' COMMENT '链接地址',
    sort INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='轮播图表';

-- 插入管理员测试数据
INSERT INTO admins (username, password, created_at, updated_at)
VALUES ('admin', 'admin123', NOW(), NOW());

-- 插入一级分类测试数据
INSERT INTO categories (name, parent_id, sort, status, created_at, updated_at, icon)
VALUES 
    ('电子产品', 0, 1, 1, NOW(), NOW(), ''),
    ('家居用品', 0, 2, 1, NOW(), NOW(), ''),
    ('服装鞋帽', 0, 3, 1, NOW(), NOW(), '');

-- 插入二级分类测试数据
INSERT INTO categories (name, parent_id, sort, status, created_at, updated_at, icon)
VALUES 
    ('手机', 1, 1, 1, NOW(), NOW(), ''),
    ('电脑', 1, 2, 1, NOW(), NOW(), ''),
    ('平板电脑', 1, 3, 1, NOW(), NOW(), ''),
    ('厨房用品', 2, 1, 1, NOW(), NOW(), ''),
    ('床上用品', 2, 2, 1, NOW(), NOW(), ''),
    ('清洁用品', 2, 3, 1, NOW(), NOW(), ''),
    ('男装', 3, 1, 1, NOW(), NOW(), ''),
    ('女装', 3, 2, 1, NOW(), NOW(), ''),
    ('童装', 3, 3, 1, NOW(), NOW(), '');

-- 插入轮播图测试数据
INSERT INTO carousels (image, title, link, sort, status, created_at, updated_at)
VALUES 
    ('/static/banner1.jpg', '新品特惠', '/pages/product/list', 1, 1, NOW(), NOW()),
    ('/static/banner2.jpg', '限时折扣', '/pages/product/special', 2, 1, NOW(), NOW()),
    ('/static/banner3.jpg', '热门推荐', '/pages/product/hot', 3, 1, NOW(), NOW());

-- 插入商品测试数据
INSERT INTO products (name, description, price, category_id, is_special, images, specs, status, created_at, updated_at)
VALUES 
    ('iPhone 15 Pro', '全新iPhone 15 Pro，搭载A17 Pro芯片，超强性能', 8999.00, 4, 0, '["/static/products/iphone1.jpg", "/static/products/iphone2.jpg"]', '{"color": ["黑色", "白色", "金色"], "storage": ["128GB", "256GB", "512GB"]}', 1, NOW(), NOW()),
    ('MacBook Air', '轻薄笔记本电脑，M2芯片，续航长达18小时', 7999.00, 5, 0, '["/static/products/mac1.jpg", "/static/products/mac2.jpg"]', '{"color": ["银色", "深空灰色"], "storage": ["256GB", "512GB"]}', 1, NOW(), NOW()),
    ('智能手表', '多功能智能手表，支持心率监测、GPS定位', 1299.00, 4, 1, '["/static/products/watch1.jpg", "/static/products/watch2.jpg"]', '{"color": ["黑色", "银色"]}', 1, NOW(), NOW());