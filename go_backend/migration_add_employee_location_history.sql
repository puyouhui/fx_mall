-- 创建配送员位置历史表
CREATE TABLE IF NOT EXISTS employee_location_history (
    id INT PRIMARY KEY AUTO_INCREMENT,
    employee_id INT NOT NULL COMMENT '员工ID',
    employee_code VARCHAR(20) NOT NULL COMMENT '员工码',
    latitude DECIMAL(10, 8) NOT NULL COMMENT '纬度',
    longitude DECIMAL(11, 8) NOT NULL COMMENT '经度',
    accuracy DECIMAL(10, 2) DEFAULT NULL COMMENT '精度（米）',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    INDEX idx_employee_id (employee_id),
    INDEX idx_employee_code (employee_code),
    INDEX idx_created_at (created_at),
    KEY idx_employee_created (employee_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配送员位置历史表';

