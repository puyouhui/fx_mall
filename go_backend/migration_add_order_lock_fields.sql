-- 添加订单锁定相关字段
-- 执行此脚本以添加 is_locked, locked_by, locked_at 字段

-- 检查并添加 is_locked 字段
SET @exist := (SELECT COUNT(*) FROM information_schema.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'is_locked');
SET @sqlstmt := IF(@exist = 0, 
    'ALTER TABLE orders ADD COLUMN is_locked TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''是否被锁定（修改中）''',
    'SELECT ''is_locked字段已存在'' AS message');
PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加 locked_by 字段
SET @exist := (SELECT COUNT(*) FROM information_schema.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'locked_by');
SET @sqlstmt := IF(@exist = 0, 
    'ALTER TABLE orders ADD COLUMN locked_by VARCHAR(10) DEFAULT NULL COMMENT ''锁定者员工码''',
    'SELECT ''locked_by字段已存在'' AS message');
PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加 locked_at 字段
SET @exist := (SELECT COUNT(*) FROM information_schema.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'locked_at');
SET @sqlstmt := IF(@exist = 0, 
    'ALTER TABLE orders ADD COLUMN locked_at DATETIME DEFAULT NULL COMMENT ''锁定时间''',
    'SELECT ''locked_at字段已存在'' AS message');
PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

