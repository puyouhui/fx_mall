package model

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go_backend/internal/database"
)

// Order 订单主表
// 目前先用于创建订单，后续可扩展状态流转、支付等逻辑
type Order struct {
	ID                  int        `json:"id"`
	OrderNumber         string     `json:"order_number"` // 订单编号
	UserID              int        `json:"user_id"`
	AddressID           int        `json:"address_id"`
	Status              string     `json:"status"`                // pending_delivery/delivering/delivered/paid/cancelled
	GoodsAmount         float64    `json:"goods_amount"`          // 商品总金额
	DeliveryFee         float64    `json:"delivery_fee"`          // 配送费
	PointsDiscount      float64    `json:"points_discount"`       // 积分抵扣金额
	CouponDiscount      float64    `json:"coupon_discount"`       // 优惠券抵扣金额
	TotalAmount         float64    `json:"total_amount"`          // 实际应付金额
	Remark              string     `json:"remark"`                // 备注
	OutOfStockStrategy  string     `json:"out_of_stock_strategy"` // 缺货处理：cancel_item/ship_available/contact_me
	TrustReceipt        bool       `json:"trust_receipt"`         // 是否信任签收
	HidePrice           bool       `json:"hide_price"`            // 是否隐藏价格
	RequirePhoneContact bool       `json:"require_phone_contact"` // 是否要求配送时电话联系
	ExpectedDeliveryAt  *time.Time `json:"expected_delivery_at"`  // 预计送达时间（可为空）
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// OrderItem 订单明细
type OrderItem struct {
	ID          int     `json:"id"`
	OrderID     int     `json:"order_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	SpecName    string  `json:"spec_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"` // 成交单价
	Subtotal    float64 `json:"subtotal"`   // 小计
	Image       string  `json:"image"`
}

// OrderCreationOptions 创建订单时的附加参数
type OrderCreationOptions struct {
	Remark              string
	OutOfStockStrategy  string
	TrustReceipt        bool
	HidePrice           bool
	RequirePhoneContact bool
	PointsDiscount      float64
	CouponDiscount      float64
}

// GenerateOrderNumber 生成订单编号
// 格式：YYYYMMDDHHmmss + 用户ID后3位（不足补0） + 随机数3位
// 例如：20240101120000123456（20位）
func GenerateOrderNumber(userID int) string {
	now := time.Now()
	// 日期时间部分：YYYYMMDDHHmmss (14位)
	timePart := now.Format("20060102150405")

	// 用户ID后3位（不足补0）
	userIDPart := fmt.Sprintf("%03d", userID%1000)

	// 随机数3位
	rand.Seed(time.Now().UnixNano())
	randomPart := fmt.Sprintf("%03d", rand.Intn(1000))

	return timePart + userIDPart + randomPart
}

// generateUniqueOrderNumber 生成唯一的订单编号（如果重复则重试）
func generateUniqueOrderNumber(userID int, maxRetries int) (string, error) {
	for i := 0; i < maxRetries; i++ {
		orderNumber := GenerateOrderNumber(userID)

		// 检查订单编号是否已存在
		var exists int
		err := database.DB.QueryRow("SELECT COUNT(*) FROM orders WHERE order_number = ?", orderNumber).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("检查订单编号失败: %v", err)
		}

		if exists == 0 {
			return orderNumber, nil
		}

		// 如果重复，等待一小段时间后重试
		time.Sleep(time.Millisecond * 10)
	}

	return "", fmt.Errorf("生成唯一订单编号失败，已重试 %d 次", maxRetries)
}

// CreateOrderFromPurchaseList 从采购单创建订单（包含事务和明细落库）
// userType: "wholesale" 表示批发客户，使用批发价；"retail" 或其他值表示零售客户，使用零售价
func CreateOrderFromPurchaseList(userID, addressID int, items []PurchaseListItem, summary *DeliveryFeeSummary, opts OrderCreationOptions, userType string) (*Order, []OrderItem, error) {
	if len(items) == 0 {
		return nil, nil, fmt.Errorf("采购单为空")
	}
	if summary == nil {
		return nil, nil, fmt.Errorf("配送费汇总为空")
	}

	outOfStockStrategy := opts.OutOfStockStrategy
	if outOfStockStrategy == "" {
		outOfStockStrategy = "contact_me"
	}
	trustReceipt := opts.TrustReceipt
	hidePrice := opts.HidePrice
	requirePhoneContact := opts.RequirePhoneContact
	pointsDiscount := opts.PointsDiscount
	couponDiscount := opts.CouponDiscount

	// 计算金额
	goodsAmount := summary.TotalAmount
	deliveryFee := summary.DeliveryFee
	if summary.IsFreeShipping {
		deliveryFee = 0
	}
	if pointsDiscount < 0 {
		pointsDiscount = 0
	}
	if couponDiscount < 0 {
		couponDiscount = 0
	}

	totalAmount := goodsAmount + deliveryFee - pointsDiscount - couponDiscount
	if totalAmount < 0 {
		totalAmount = 0
	}

	// 生成唯一的订单编号
	orderNumber, err := generateUniqueOrderNumber(userID, 5)
	if err != nil {
		return nil, nil, fmt.Errorf("生成订单编号失败: %v", err)
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		// 如果 err 不为 nil，则回滚
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 插入订单主表
	res, err := tx.Exec(`
		INSERT INTO orders (
			order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount, coupon_discount, total_amount,
			remark, out_of_stock_strategy, trust_receipt, hide_price, require_phone_contact, expected_delivery_at,
			created_at, updated_at
		) VALUES (?, ?, ?, 'pending_delivery', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NOW(), NOW())
	`,
		orderNumber, userID, addressID,
		goodsAmount, deliveryFee, pointsDiscount, couponDiscount, totalAmount,
		opts.Remark, outOfStockStrategy, boolToTinyInt(trustReceipt), boolToTinyInt(hidePrice), boolToTinyInt(requirePhoneContact),
	)
	if err != nil {
		return nil, nil, err
	}

	orderID64, err := res.LastInsertId()
	if err != nil {
		return nil, nil, err
	}
	orderID := int(orderID64)

	// 插入订单明细
	orderItems := make([]OrderItem, 0, len(items))
	itemStmt, err := tx.Prepare(`
		INSERT INTO order_items (
			order_id, product_id, product_name, spec_name, quantity, unit_price, subtotal, image
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = itemStmt.Close()
	}()

	for _, it := range items {
		// 根据用户类型计算价格，与配送费计算保持一致
		var price float64
		if userType == "wholesale" {
			price = it.SpecSnapshot.WholesalePrice
			if price <= 0 {
				price = it.SpecSnapshot.RetailPrice
			}
		} else {
			price = it.SpecSnapshot.RetailPrice
			if price <= 0 {
				price = it.SpecSnapshot.WholesalePrice
			}
		}
		if price <= 0 {
			price = it.SpecSnapshot.Cost
		}
		if price < 0 {
			price = 0
		}
		subtotal := price * float64(it.Quantity)

		if _, err = itemStmt.Exec(
			orderID,
			it.ProductID,
			it.ProductName,
			it.SpecName,
			it.Quantity,
			price,
			subtotal,
			it.ProductImage,
		); err != nil {
			return nil, nil, err
		}

		orderItems = append(orderItems, OrderItem{
			OrderID:     orderID,
			ProductID:   it.ProductID,
			ProductName: it.ProductName,
			SpecName:    it.SpecName,
			Quantity:    it.Quantity,
			UnitPrice:   price,
			Subtotal:    subtotal,
			Image:       it.ProductImage,
		})
	}

	// 创建成功后可以清空采购单
	if _, err = tx.Exec(`DELETE FROM purchase_list_items WHERE user_id = ?`, userID); err != nil {
		return nil, nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, nil, err
	}

	// 查询刚插入的订单记录
	var order Order
	var expectedDelivery sql.NullTime
	err = database.DB.QueryRow(`
		SELECT id, order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount,
		       coupon_discount, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, created_at, updated_at
		FROM orders WHERE id = ?
	`, orderID).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &order.GoodsAmount, &order.DeliveryFee,
		&order.PointsDiscount, &order.CouponDiscount, &order.TotalAmount, &order.Remark,
		&order.OutOfStockStrategy, &order.TrustReceipt, &order.HidePrice, &order.RequirePhoneContact,
		&expectedDelivery, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}
	if expectedDelivery.Valid {
		t := expectedDelivery.Time
		order.ExpectedDeliveryAt = &t
	}

	return &order, orderItems, nil
}

func boolToTinyInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// GetOrdersWithPagination 获取订单列表（支持分页和搜索）
func GetOrdersWithPagination(pageNum, pageSize int, keyword string, status string) ([]Order, int, error) {
	var orders []Order
	var total int

	// 计算偏移量
	offset := (pageNum - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 构建查询条件
	where := "1=1"
	args := []interface{}{}

	// 关键词搜索：订单ID、订单编号、用户ID
	if keyword != "" {
		where += " AND (id = ? OR order_number LIKE ? OR user_id = ?)"
		// 尝试将关键词转换为数字
		var idValue int
		keywordPattern := "%" + keyword + "%"
		if _, err := fmt.Sscanf(keyword, "%d", &idValue); err == nil {
			args = append(args, idValue, keywordPattern, idValue)
		} else {
			// 如果不是数字，使用0（不会匹配任何ID）
			args = append(args, 0, keywordPattern, 0)
		}
	}

	// 状态筛选
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	// 获取总数量
	countQuery := "SELECT COUNT(*) FROM orders WHERE " + where
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	query := `
		SELECT id, order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount,
		       coupon_discount, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, created_at, updated_at
		FROM orders WHERE ` + where + ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var order Order
		var expectedDelivery sql.NullTime

		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &order.GoodsAmount, &order.DeliveryFee,
			&order.PointsDiscount, &order.CouponDiscount, &order.TotalAmount, &order.Remark,
			&order.OutOfStockStrategy, &order.TrustReceipt, &order.HidePrice, &order.RequirePhoneContact,
			&expectedDelivery, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if expectedDelivery.Valid {
			t := expectedDelivery.Time
			order.ExpectedDeliveryAt = &t
		}

		orders = append(orders, order)
	}

	return orders, total, nil
}

// GetOrderByID 根据ID获取订单详情
func GetOrderByID(id int) (*Order, error) {
	var order Order
	var expectedDelivery sql.NullTime

	query := `
		SELECT id, order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount,
		       coupon_discount, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, created_at, updated_at
		FROM orders WHERE id = ?
	`
	err := database.DB.QueryRow(query, id).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &order.GoodsAmount, &order.DeliveryFee,
		&order.PointsDiscount, &order.CouponDiscount, &order.TotalAmount, &order.Remark,
		&order.OutOfStockStrategy, &order.TrustReceipt, &order.HidePrice, &order.RequirePhoneContact,
		&expectedDelivery, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if expectedDelivery.Valid {
		t := expectedDelivery.Time
		order.ExpectedDeliveryAt = &t
	}

	return &order, nil
}

// GetOrderItemsByOrderID 根据订单ID获取订单明细
func GetOrderItemsByOrderID(orderID int) ([]OrderItem, error) {
	var items []OrderItem

	query := `
		SELECT id, order_id, product_id, product_name, spec_name, quantity, unit_price, subtotal, image
		FROM order_items WHERE order_id = ? ORDER BY id
	`
	rows, err := database.DB.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.SpecName,
			&item.Quantity, &item.UnitPrice, &item.Subtotal, &item.Image,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// CountOrdersBySalesCode 统计某个销售员名下客户的订单总数、待配送数量、今日新增订单数量
func CountOrdersBySalesCode(employeeCode string) (total int, pendingDelivery int, todayTotal int, err error) {
	query := `
		SELECT 
			COUNT(*) AS total,
			SUM(CASE WHEN o.status IN ('pending_delivery', 'pending') THEN 1 ELSE 0 END) AS pending_delivery,
			SUM(CASE WHEN DATE(o.created_at) = CURRENT_DATE THEN 1 ELSE 0 END) AS today_total
		FROM orders o
		JOIN mini_app_users u ON o.user_id = u.id
		WHERE u.sales_code = ?
	`

	err = database.DB.QueryRow(query, employeeCode).Scan(&total, &pendingDelivery, &todayTotal)
	if err != nil {
		return 0, 0, 0, err
	}
	return total, pendingDelivery, todayTotal, nil
}

// GetOrdersBySalesCode 获取销售员名下客户的订单列表（分页，支持状态 & 关键字搜索）
func GetOrdersBySalesCode(employeeCode string, pageNum, pageSize int, status, keyword string) ([]map[string]interface{}, int, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (pageNum - 1) * pageSize

	where := "u.sales_code = ?"
	args := []interface{}{employeeCode}

	if status != "" {
		// 状态筛选逻辑与客户订单接口保持一致
		if status == "pending_delivery" || status == "pending" {
			where += " AND (o.status = ? OR o.status = 'pending')"
			args = append(args, "pending_delivery")
		} else if status == "delivered" || status == "shipped" {
			where += " AND (o.status = ? OR o.status = 'shipped')"
			args = append(args, "delivered")
		} else {
			where += " AND o.status = ?"
			args = append(args, status)
		}
	}

	if keyword != "" {
		kw := "%" + strings.TrimSpace(keyword) + "%"
		where += " AND (o.order_number LIKE ? OR u.name LIKE ? OR u.user_code LIKE ? OR a.name LIKE ?)"
		args = append(args, kw, kw, kw, kw)
	}

	// 统计总数
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM orders o
		JOIN mini_app_users u ON o.user_id = u.id
		LEFT JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE ` + where
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	query := `
		SELECT
			o.id,
			o.order_number,
			o.status,
			o.total_amount,
			o.created_at,
			u.name AS user_name,
			u.user_code,
			a.name AS store_name,
			a.address,
			a.phone AS contact_phone
		FROM orders o
		JOIN mini_app_users u ON o.user_id = u.id
		LEFT JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE ` + where + `
		ORDER BY o.id DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]map[string]interface{}, 0)
	for rows.Next() {
		var (
			id           int
			orderNumber  string
			statusVal    string
			totalAmount  float64
			createdAt    time.Time
			userName     sql.NullString
			userCode     sql.NullString
			storeName    sql.NullString
			address      sql.NullString
			contactPhone sql.NullString
		)

		if err := rows.Scan(
			&id,
			&orderNumber,
			&statusVal,
			&totalAmount,
			&createdAt,
			&userName,
			&userCode,
			&storeName,
			&address,
			&contactPhone,
		); err != nil {
			return nil, 0, err
		}

		itemCount, _ := GetOrderItemCountByOrderID(id)

		order := map[string]interface{}{
			"id":            id,
			"order_number":  orderNumber,
			"status":        statusVal,
			"total_amount":  totalAmount,
			"created_at":    createdAt,
			"user_name":     getStringValue(userName),
			"user_code":     getStringValue(userCode),
			"store_name":    getStringValue(storeName),
			"address":       getStringValue(address),
			"contact_phone": getStringValue(contactPhone),
			"item_count":    itemCount,
		}

		orders = append(orders, order)
	}

	return orders, total, nil
}

// GetPendingOrdersBySalesCode 获取销售员名下客户的待配送订单列表（分页）
func GetPendingOrdersBySalesCode(employeeCode string, pageNum, pageSize int) ([]map[string]interface{}, int, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (pageNum - 1) * pageSize

	// 统计总数
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM orders o
		JOIN mini_app_users u ON o.user_id = u.id
		WHERE u.sales_code = ? AND o.status IN ('pending_delivery', 'pending')
	`
	if err := database.DB.QueryRow(countQuery, employeeCode).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	query := `
		SELECT
			o.id,
			o.order_number,
			o.status,
			o.total_amount,
			o.created_at,
			a.name AS store_name,
			a.address,
			a.phone AS contact_phone
		FROM orders o
		JOIN mini_app_users u ON o.user_id = u.id
		LEFT JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE u.sales_code = ? AND o.status IN ('pending_delivery', 'pending')
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := database.DB.Query(query, employeeCode, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var orderNumber string
		var status string
		var totalAmount float64
		var createdAt time.Time
		var storeName sql.NullString
		var address sql.NullString
		var contactPhone sql.NullString

		if err := rows.Scan(&id, &orderNumber, &status, &totalAmount, &createdAt, &storeName, &address, &contactPhone); err != nil {
			return nil, 0, err
		}

		itemCount, _ := GetOrderItemCountByOrderID(id)

		order := map[string]interface{}{
			"id":            id,
			"order_number":  orderNumber,
			"status":        status,
			"total_amount":  totalAmount,
			"created_at":    createdAt,
			"store_name":    getStringValue(storeName),
			"address":       getStringValue(address),
			"contact_phone": getStringValue(contactPhone),
			"item_count":    itemCount,
		}

		orders = append(orders, order)
	}

	return orders, total, nil
}

// GetOrderItemCountByOrderID 根据订单ID获取订单商品数量
func GetOrderItemCountByOrderID(orderID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM order_items WHERE order_id = ?`
	err := database.DB.QueryRow(query, orderID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountOrdersByUserID 统计指定用户的订单数量
func CountOrdersByUserID(userID int) (int, error) {
	var count int
	err := database.DB.QueryRow(`
		SELECT COUNT(*)
		FROM orders
		WHERE user_id = ?
	`, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetOrderSummaryByUserID 获取用户历史订单汇总（总金额 & 订单数量）
func GetOrderSummaryByUserID(userID int) (float64, int, error) {
	var totalAmount float64
	var orderCount int
	err := database.DB.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM orders
		WHERE user_id = ?
	`, userID).Scan(&totalAmount, &orderCount)
	if err != nil {
		return 0, 0, err
	}
	return totalAmount, orderCount, nil
}

// GetRecentOrdersByUserID 获取用户最近的订单列表（按创建时间倒序）
func GetRecentOrdersByUserID(userID int, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 3
	}

	query := `
		SELECT id, order_number, total_amount, status, created_at
		FROM orders
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`

	rows, err := database.DB.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]map[string]interface{}, 0)
	for rows.Next() {
		var (
			id          int
			orderNumber string
			totalAmount float64
			status      string
			createdAt   time.Time
		)

		if err := rows.Scan(&id, &orderNumber, &totalAmount, &status, &createdAt); err != nil {
			return nil, err
		}

		order := map[string]interface{}{
			"id":           id,
			"order_number": orderNumber,
			"total_amount": totalAmount,
			"status":       status,
			"created_at":   createdAt,
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// UpdateOrderStatus 更新订单状态
func UpdateOrderStatus(orderID int, newStatus string) error {
	query := "UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ?"
	_, err := database.DB.Exec(query, newStatus, orderID)
	return err
}
