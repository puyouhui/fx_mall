-- 为 orders 表添加 order_number 字段的迁移脚本
-- 如果字段已存在，此脚本会报错，可以忽略

-- 添加 order_number 字段
ALTER TABLE orders 
ADD COLUMN order_number VARCHAR(32) UNIQUE COMMENT '订单编号' AFTER id;

-- 添加索引
ALTER TABLE orders 
ADD INDEX idx_order_number (order_number);

-- 为已有订单生成订单编号（可选，如果需要为历史订单生成编号）
-- UPDATE orders SET order_number = CONCAT(
--   DATE_FORMAT(created_at, '%Y%m%d%H%i%s'),
--   LPAD(user_id % 1000, 3, '0'),
--   LPAD(FLOOR(RAND() * 1000), 3, '0')
-- ) WHERE order_number IS NULL OR order_number = '';

UPDATE orders SET order_number = CONCAT(
  DATE_FORMAT(created_at, '%Y%m%d%H%i%s'),
  LPAD(user_id % 1000, 3, '0'),
  LPAD(FLOOR(RAND() * 1000), 3, '0')
) WHERE order_number IS NULL OR order_number = '';
