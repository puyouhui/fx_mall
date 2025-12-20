package model

import (
	"database/sql"
	"fmt"
	"time"

	"go_backend/internal/database"
)

// PaymentVerificationRequest 收款审核申请
type PaymentVerificationRequest struct {
	ID                int        `json:"id"`
	OrderID           int        `json:"order_id"`
	OrderNumber       string     `json:"order_number"`
	SalesEmployeeCode string     `json:"sales_employee_code"`
	SalesEmployeeName string     `json:"sales_employee_name"`
	CustomerID        int        `json:"customer_id"`
	CustomerName      string     `json:"customer_name"`
	OrderAmount       float64    `json:"order_amount"`
	RequestReason     string     `json:"request_reason"`
	Status            string     `json:"status"` // pending, approved, rejected
	AdminID           *int       `json:"admin_id,omitempty"`
	AdminName         *string    `json:"admin_name,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	ReviewRemark      *string    `json:"review_remark,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreatePaymentVerificationRequest 创建收款审核申请
func CreatePaymentVerificationRequest(orderID int, salesEmployeeCode string, requestReason string) (*PaymentVerificationRequest, error) {
	// 获取订单信息
	order, err := GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("获取订单失败: %v", err)
	}
	if order == nil {
		return nil, fmt.Errorf("订单不存在")
	}

	// 检查订单状态是否为已送达
	if order.Status != "delivered" && order.Status != "shipped" {
		return nil, fmt.Errorf("订单状态不是已送达，无法提交收款申请")
	}

	// 检查是否已有待审核的申请
	existing, err := GetPendingPaymentVerificationByOrderID(orderID)
	if err == nil && existing != nil {
		return existing, nil // 返回已存在的申请
	}

	// 获取销售员信息
	salesEmployee, err := GetEmployeeByEmployeeCode(salesEmployeeCode)
	if err != nil {
		return nil, fmt.Errorf("获取销售员信息失败: %v", err)
	}
	if salesEmployee == nil {
		return nil, fmt.Errorf("销售员不存在")
	}

	// 获取客户信息
	customer, err := GetMiniAppUserByID(order.UserID)
	if err != nil {
		return nil, fmt.Errorf("获取客户信息失败: %v", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("客户不存在")
	}

	// 处理可能为空的字段
	orderNumber := order.OrderNumber
	if orderNumber == "" {
		orderNumber = fmt.Sprintf("ORDER-%d", order.ID)
	}

	salesEmployeeName := salesEmployee.Name
	if salesEmployeeName == "" {
		salesEmployeeName = salesEmployeeCode
	}

	customerName := customer.Name
	if customerName == "" {
		customerName = "未命名客户"
	}

	query := `
		INSERT INTO payment_verification_requests (
			order_id, order_number, sales_employee_code, sales_employee_name,
			customer_id, customer_name, order_amount, request_reason, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending')
	`

	result, err := database.DB.Exec(
		query,
		order.ID,
		orderNumber,
		salesEmployeeCode,
		salesEmployeeName,
		order.UserID,
		customerName,
		order.TotalAmount,
		requestReason,
	)
	if err != nil {
		return nil, fmt.Errorf("插入收款申请失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取插入ID失败: %v", err)
	}

	return GetPaymentVerificationRequestByID(int(id))
}

// GetPaymentVerificationRequestByID 根据ID获取收款审核申请
func GetPaymentVerificationRequestByID(id int) (*PaymentVerificationRequest, error) {
	query := `
		SELECT id, order_id, order_number, sales_employee_code, sales_employee_name,
		       customer_id, customer_name, order_amount, request_reason, status,
		       admin_id, admin_name, reviewed_at, review_remark, created_at, updated_at
		FROM payment_verification_requests
		WHERE id = ?
	`

	var req PaymentVerificationRequest
	var adminID sql.NullInt64
	var adminName, reviewRemark sql.NullString
	var reviewedAt sql.NullTime

	err := database.DB.QueryRow(query, id).Scan(
		&req.ID, &req.OrderID, &req.OrderNumber, &req.SalesEmployeeCode, &req.SalesEmployeeName,
		&req.CustomerID, &req.CustomerName, &req.OrderAmount, &req.RequestReason, &req.Status,
		&adminID, &adminName, &reviewedAt, &reviewRemark, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询收款审核申请失败: %v", err)
	}

	// 处理可空字段
	if adminID.Valid {
		id := int(adminID.Int64)
		req.AdminID = &id
	}
	if adminName.Valid {
		req.AdminName = &adminName.String
	}
	if reviewedAt.Valid {
		req.ReviewedAt = &reviewedAt.Time
	}
	if reviewRemark.Valid {
		req.ReviewRemark = &reviewRemark.String
	}

	return &req, nil
}

// GetPendingPaymentVerificationByOrderID 根据订单ID获取待审核的申请
func GetPendingPaymentVerificationByOrderID(orderID int) (*PaymentVerificationRequest, error) {
	query := `
		SELECT id, order_id, order_number, sales_employee_code, sales_employee_name,
		       customer_id, customer_name, order_amount, request_reason, status,
		       admin_id, admin_name, reviewed_at, review_remark, created_at, updated_at
		FROM payment_verification_requests
		WHERE order_id = ? AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var req PaymentVerificationRequest
	var adminID sql.NullInt64
	var reviewedAt sql.NullTime

	err := database.DB.QueryRow(query, orderID).Scan(
		&req.ID, &req.OrderID, &req.OrderNumber, &req.SalesEmployeeCode, &req.SalesEmployeeName,
		&req.CustomerID, &req.CustomerName, &req.OrderAmount, &req.RequestReason, &req.Status,
		&adminID, &req.AdminName, &reviewedAt, &req.ReviewRemark, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if adminID.Valid {
		adminIDInt := int(adminID.Int64)
		req.AdminID = &adminIDInt
	}
	if reviewedAt.Valid {
		req.ReviewedAt = &reviewedAt.Time
	}

	return &req, nil
}

// GetPaymentVerificationRequests 获取收款审核申请列表（管理员）
func GetPaymentVerificationRequests(pageNum, pageSize int, status string) ([]*PaymentVerificationRequest, int, error) {
	offset := (pageNum - 1) * pageSize

	where := "1=1"
	args := []interface{}{}

	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM payment_verification_requests WHERE " + where
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	query := `
		SELECT id, order_id, order_number, sales_employee_code, sales_employee_name,
		       customer_id, customer_name, order_amount, request_reason, status,
		       admin_id, admin_name, reviewed_at, review_remark, created_at, updated_at
		FROM payment_verification_requests
		WHERE ` + where + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []*PaymentVerificationRequest
	for rows.Next() {
		var req PaymentVerificationRequest
		var adminID sql.NullInt64
		var adminName, reviewRemark sql.NullString
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&req.ID, &req.OrderID, &req.OrderNumber, &req.SalesEmployeeCode, &req.SalesEmployeeName,
			&req.CustomerID, &req.CustomerName, &req.OrderAmount, &req.RequestReason, &req.Status,
			&adminID, &adminName, &reviewedAt, &reviewRemark, &req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// 处理可空字段
		if adminID.Valid {
			id := int(adminID.Int64)
			req.AdminID = &id
		}
		if adminName.Valid {
			req.AdminName = &adminName.String
		}
		if reviewedAt.Valid {
			req.ReviewedAt = &reviewedAt.Time
		}
		if reviewRemark.Valid {
			req.ReviewRemark = &reviewRemark.String
		}

		requests = append(requests, &req)
	}

	return requests, total, nil
}

// ReviewPaymentVerificationRequest 审核收款申请
func ReviewPaymentVerificationRequest(requestID int, adminID int, adminName string, approved bool, reviewRemark string) error {
	// 获取申请信息
	req, err := GetPaymentVerificationRequestByID(requestID)
	if err != nil {
		return err
	}

	if req.Status != "pending" {
		return sql.ErrNoRows // 申请已被审核
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 更新申请状态
	status := "rejected"
	if approved {
		status = "approved"
	}

	updateQuery := `
		UPDATE payment_verification_requests
		SET status = ?, admin_id = ?, admin_name = ?, reviewed_at = NOW(), review_remark = ?
		WHERE id = ?
	`
	_, err = tx.Exec(updateQuery, status, adminID, adminName, reviewRemark, requestID)
	if err != nil {
		return err
	}

	// 如果审核通过，更新订单状态为已收款
	if approved {
		orderUpdateQuery := `
			UPDATE orders
			SET status = 'paid', updated_at = NOW()
			WHERE id = ? AND status IN ('delivered', 'shipped')
		`
		result, err := tx.Exec(orderUpdateQuery, req.OrderID)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			// 订单状态已改变，回滚事务
			return sql.ErrNoRows
		}
	}

	return tx.Commit()
}
