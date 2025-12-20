-- 创建收款审核申请表
CREATE TABLE IF NOT EXISTS payment_verification_requests (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL COMMENT '订单ID',
    order_number VARCHAR(50) NOT NULL COMMENT '订单号',
    sales_employee_code VARCHAR(50) NOT NULL COMMENT '销售员代码',
    sales_employee_name VARCHAR(100) COMMENT '销售员姓名',
    customer_id INT NOT NULL COMMENT '客户ID',
    customer_name VARCHAR(100) COMMENT '客户姓名',
    order_amount DECIMAL(10, 2) NOT NULL COMMENT '订单金额',
    request_reason TEXT COMMENT '申请原因/备注',
    status ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending' COMMENT '审核状态：pending-待审核，approved-已通过，rejected-已拒绝',
    admin_id INT COMMENT '审核管理员ID',
    admin_name VARCHAR(100) COMMENT '审核管理员姓名',
    reviewed_at DATETIME COMMENT '审核时间',
    review_remark TEXT COMMENT '审核备注',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_order_id (order_id),
    INDEX idx_sales_employee_code (sales_employee_code),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收款审核申请表';

