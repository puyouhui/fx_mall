package model

import (
	"database/sql"
	"time"

	"go_backend/internal/database"
)

// SupplierPayment 供应商付款记录
type SupplierPayment struct {
	ID             int       `json:"id"`
	SupplierID     int       `json:"supplier_id"`
	PaymentDate    time.Time `json:"payment_date"`
	PaymentAmount  float64   `json:"payment_amount"`
	PaymentMethod  *string   `json:"payment_method,omitempty"`
	PaymentAccount *string   `json:"payment_account,omitempty"`
	PaymentReceipt *string   `json:"payment_receipt,omitempty"`
	Remark         *string   `json:"remark,omitempty"`
	CreatedBy      *string   `json:"created_by,omitempty"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SupplierPaymentItem 供应商付款明细
type SupplierPaymentItem struct {
	ID          int       `json:"id"`
	PaymentID   int       `json:"payment_id"`
	OrderID     int       `json:"order_id"`
	OrderItemID int       `json:"order_item_id"`
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name"`
	SpecName    string    `json:"spec_name"`
	Quantity    int       `json:"quantity"`
	CostPrice   float64   `json:"cost_price"`
	Subtotal    float64   `json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateSupplierPayment 创建供应商付款记录，返回创建的付款记录ID
func CreateSupplierPayment(payment *SupplierPayment, items []SupplierPaymentItem) (int64, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 插入付款记录
	paymentQuery := `
		INSERT INTO supplier_payments 
		(supplier_id, payment_date, payment_amount, payment_method, payment_account, payment_receipt, remark, created_by, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.Exec(
		paymentQuery,
		payment.SupplierID,
		payment.PaymentDate,
		payment.PaymentAmount,
		payment.PaymentMethod,
		payment.PaymentAccount,
		payment.PaymentReceipt,
		payment.Remark,
		payment.CreatedBy,
		payment.Status,
	)
	if err != nil {
		return 0, err
	}

	paymentID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 插入付款明细
	itemQuery := `
		INSERT INTO supplier_payment_items 
		(payment_id, order_id, order_item_id, product_id, product_name, spec_name, quantity, cost_price, subtotal, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`
	for _, item := range items {
		_, err = tx.Exec(
			itemQuery,
			paymentID,
			item.OrderID,
			item.OrderItemID,
			item.ProductID,
			item.ProductName,
			item.SpecName,
			item.Quantity,
			item.CostPrice,
			item.Subtotal,
		)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return paymentID, nil
}

// GetSupplierPaymentByID 根据ID获取付款记录
func GetSupplierPaymentByID(id int) (*SupplierPayment, error) {
	var payment SupplierPayment
	query := `
		SELECT id, supplier_id, payment_date, payment_amount, payment_method, payment_account, 
		       payment_receipt, remark, created_by, status, created_at, updated_at
		FROM supplier_payments
		WHERE id = ?
	`
	err := database.DB.QueryRow(query, id).Scan(
		&payment.ID,
		&payment.SupplierID,
		&payment.PaymentDate,
		&payment.PaymentAmount,
		&payment.PaymentMethod,
		&payment.PaymentAccount,
		&payment.PaymentReceipt,
		&payment.Remark,
		&payment.CreatedBy,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// GetSupplierPayments 获取供应商付款记录列表
func GetSupplierPayments(supplierID *int, startDate, endDate *time.Time, pageNum, pageSize int) ([]SupplierPayment, int, error) {
	offset := (pageNum - 1) * pageSize

	// 构建查询条件
	whereClause := "WHERE status = 1"
	args := []interface{}{}
	countArgs := []interface{}{}

	if supplierID != nil {
		whereClause += " AND supplier_id = ?"
		args = append(args, *supplierID)
		countArgs = append(countArgs, *supplierID)
	}
	if startDate != nil {
		whereClause += " AND payment_date >= ?"
		args = append(args, *startDate)
		countArgs = append(countArgs, *startDate)
	}
	if endDate != nil {
		whereClause += " AND payment_date <= ?"
		args = append(args, *endDate)
		countArgs = append(countArgs, *endDate)
	}

	// 获取总数
	var total int
	countQuery := "SELECT COUNT(*) FROM supplier_payments " + whereClause
	err := database.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT id, supplier_id, payment_date, payment_amount, payment_method, payment_account, 
		       payment_receipt, remark, created_by, status, created_at, updated_at
		FROM supplier_payments
		` + whereClause + `
		ORDER BY payment_date DESC, id DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	payments := []SupplierPayment{}
	for rows.Next() {
		var payment SupplierPayment
		err := rows.Scan(
			&payment.ID,
			&payment.SupplierID,
			&payment.PaymentDate,
			&payment.PaymentAmount,
			&payment.PaymentMethod,
			&payment.PaymentAccount,
			&payment.PaymentReceipt,
			&payment.Remark,
			&payment.CreatedBy,
			&payment.Status,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			continue
		}
		payments = append(payments, payment)
	}

	return payments, total, nil
}

// GetSupplierPaymentItems 获取付款明细
func GetSupplierPaymentItems(paymentID int) ([]SupplierPaymentItem, error) {
	query := `
		SELECT id, payment_id, order_id, order_item_id, product_id, product_name, spec_name, 
		       quantity, cost_price, subtotal, created_at
		FROM supplier_payment_items
		WHERE payment_id = ?
		ORDER BY id
	`
	rows, err := database.DB.Query(query, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []SupplierPaymentItem{}
	for rows.Next() {
		var item SupplierPaymentItem
		err := rows.Scan(
			&item.ID,
			&item.PaymentID,
			&item.OrderID,
			&item.OrderItemID,
			&item.ProductID,
			&item.ProductName,
			&item.SpecName,
			&item.Quantity,
			&item.CostPrice,
			&item.Subtotal,
			&item.CreatedAt,
		)
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

// CheckOrderItemPaid 检查订单商品是否已付款
func CheckOrderItemPaid(orderItemID int) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM supplier_payment_items spi
		INNER JOIN supplier_payments sp ON spi.payment_id = sp.id
		WHERE spi.order_item_id = ? AND sp.status = 1
	`
	err := database.DB.QueryRow(query, orderItemID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetPaidOrderItemIDs 获取已付款的订单商品ID列表
func GetPaidOrderItemIDs(supplierID *int) (map[int]bool, error) {
	query := `
		SELECT DISTINCT spi.order_item_id
		FROM supplier_payment_items spi
		INNER JOIN supplier_payments sp ON spi.payment_id = sp.id
		WHERE sp.status = 1
	`
	args := []interface{}{}
	if supplierID != nil {
		query += " AND sp.supplier_id = ?"
		args = append(args, *supplierID)
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	paidItems := make(map[int]bool)
	for rows.Next() {
		var orderItemID int
		if err := rows.Scan(&orderItemID); err == nil {
			paidItems[orderItemID] = true
		}
	}

	return paidItems, nil
}

// CancelSupplierPayment 撤销付款（软删除）
func CancelSupplierPayment(id int) error {
	query := "UPDATE supplier_payments SET status = 0, updated_at = NOW() WHERE id = ?"
	_, err := database.DB.Exec(query, id)
	return err
}
