-- 设置字符集
SET NAMES utf8mb4;

-- 创建热门搜索关键词表
CREATE TABLE IF NOT EXISTS hot_search_keywords (
    id INT AUTO_INCREMENT PRIMARY KEY,
    keyword VARCHAR(100) NOT NULL COMMENT '关键词',
    sort INT DEFAULT 0 COMMENT '排序（越小越靠前）',
    status TINYINT DEFAULT 1 COMMENT '状态：1启用，0禁用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='热门搜索关键词';

-- 插入默认的热门搜索关键词
INSERT INTO hot_search_keywords (keyword, sort, status) VALUES
('火锅食材', 1, 1),
('调味品', 2, 1),
('饮料', 3, 1),
('零食', 4, 1),
('水果', 5, 1),
('蔬菜', 6, 1),
('肉类', 7, 1),
('乳制品', 8, 1);


