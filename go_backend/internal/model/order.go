package model

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"go_backend/internal/database"
)

// Order 订单主表
// 目前先用于创建订单，后续可扩展状态流转、支付等逻辑
type Order struct {
	ID                   int        `json:"id"`
	OrderNumber          string     `json:"order_number"` // 订单编号
	UserID               int        `json:"user_id"`
	AddressID            int        `json:"address_id"`
	Status               string     `json:"status"`                           // pending_delivery/delivering/delivered/paid/cancelled
	DeliveryEmployeeCode *string    `json:"delivery_employee_code,omitempty"` // 配送员员工码（接单时记录）
	GoodsAmount          float64    `json:"goods_amount"`                     // 商品总金额
	DeliveryFee          float64    `json:"delivery_fee"`                     // 配送费
	PointsDiscount       float64    `json:"points_discount"`                  // 积分抵扣金额
	CouponDiscount       float64    `json:"coupon_discount"`                  // 优惠券抵扣金额
	IsUrgent             bool       `json:"is_urgent"`                        // 是否加急订单
	UrgentFee            float64    `json:"urgent_fee"`                       // 加急费用
	TotalAmount          float64    `json:"total_amount"`                     // 实际应付金额
	Remark               string     `json:"remark"`                           // 备注
	OutOfStockStrategy   string     `json:"out_of_stock_strategy"`            // 缺货处理：cancel_item/ship_available/contact_me
	TrustReceipt         bool       `json:"trust_receipt"`                    // 是否信任签收
	HidePrice            bool       `json:"hide_price"`                       // 是否隐藏价格
	RequirePhoneContact  bool       `json:"require_phone_contact"`            // 是否要求配送时电话联系
	ExpectedDeliveryAt   *time.Time `json:"expected_delivery_at"`             // 预计送达时间（可为空）
	WeatherInfo          *string    `json:"weather_info,omitempty"`           // 天气信息（JSON格式）
	IsIsolated           bool       `json:"is_isolated"`                      // 是否孤立订单（8公里内无其他订单）
	DeliveryFeeSettled   bool       `json:"delivery_fee_settled"`            // 配送费是否已结算
	SettlementDate       *time.Time `json:"settlement_date,omitempty"`       // 结算日期
	IsLocked             bool       `json:"is_locked"`                       // 是否被锁定（修改中）
	LockedBy             *string    `json:"locked_by,omitempty"`             // 锁定者员工码
	LockedAt             *time.Time  `json:"locked_at,omitempty"`             // 锁定时间
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
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
	IsPicked    bool    `json:"is_picked"` // 是否已取货
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
	IsUrgent            bool
	UrgentFee           float64
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
	isUrgent := opts.IsUrgent
	urgentFee := opts.UrgentFee

	// 计算金额
	goodsAmount := summary.TotalAmount
	deliveryFee := summary.DeliveryFee
	if summary.IsFreeShipping {
		deliveryFee = 0
		log.Printf("订单创建: 满足免费配送条件，配送费设置为0 (商品金额: %.2f, 免费配送阈值: %.2f)\n", goodsAmount, summary.FreeShippingThreshold)
	}
	if pointsDiscount < 0 {
		pointsDiscount = 0
	}
	if couponDiscount < 0 {
		couponDiscount = 0
	}
	if urgentFee < 0 {
		urgentFee = 0
	}
	if !isUrgent {
		urgentFee = 0
	}

	totalAmount := goodsAmount + deliveryFee + urgentFee - pointsDiscount - couponDiscount
	if totalAmount < 0 {
		totalAmount = 0
	}
	
	log.Printf("订单创建: 商品金额=%.2f, 配送费=%.2f, 积分抵扣=%.2f, 优惠券抵扣=%.2f, 加急费=%.2f, 实付=%.2f\n",
		goodsAmount, deliveryFee, pointsDiscount, couponDiscount, urgentFee, totalAmount)

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
			order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount, coupon_discount, is_urgent, urgent_fee, total_amount,
			remark, out_of_stock_strategy, trust_receipt, hide_price, require_phone_contact, expected_delivery_at,
			created_at, updated_at
		) VALUES (?, ?, ?, 'pending_delivery', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NOW(), NOW())
	`,
		orderNumber, userID, addressID,
		goodsAmount, deliveryFee, pointsDiscount, couponDiscount, boolToTinyInt(isUrgent), urgentFee, totalAmount,
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

	// 注意：不再在创建订单时清空采购单
	// 采购单的清空和恢复由调用方（API层）处理，以便区分用户自己添加的商品和销售员添加的商品

	if err = tx.Commit(); err != nil {
		return nil, nil, err
	}

	// 记录配送流程日志：订单创建
	remark := "订单创建"
	deliveryLog := &DeliveryLog{
		OrderID:    orderID,
		Action:     DeliveryLogActionCreated,
		ActionTime: time.Now(),
		Remark:     &remark,
	}
	_ = CreateDeliveryLog(deliveryLog) // 记录日志失败不影响主流程

	// 计算并存储配送费计算结果和利润信息
	// 注意：这里在事务外执行，避免阻塞订单创建
	go func() {
		// 先更新订单的配送相关信息（孤立状态、天气信息等）
		_ = UpdateOrderDeliveryInfo(orderID)

		// 然后计算并存储配送费计算结果和利润
		_ = CalculateAndStoreOrderProfit(orderID)
	}()

	// 查询刚插入的订单记录
	var order Order
	var expectedDelivery sql.NullTime
	var weatherInfo sql.NullString
	var isUrgentTinyInt, hidePriceTinyInt, trustReceiptTinyInt, requirePhoneContactTinyInt, isIsolatedTinyInt int
	err = database.DB.QueryRow(`
		SELECT id, order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount,
		       coupon_discount, is_urgent, urgent_fee, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, weather_info, is_isolated, created_at, updated_at
		FROM orders WHERE id = ?
	`, orderID).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &order.GoodsAmount, &order.DeliveryFee,
		&order.PointsDiscount, &order.CouponDiscount, &isUrgentTinyInt, &order.UrgentFee, &order.TotalAmount, &order.Remark,
		&order.OutOfStockStrategy, &trustReceiptTinyInt, &hidePriceTinyInt, &requirePhoneContactTinyInt,
		&expectedDelivery, &weatherInfo, &isIsolatedTinyInt, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}
	order.IsUrgent = isUrgentTinyInt == 1
	order.TrustReceipt = trustReceiptTinyInt == 1
	order.HidePrice = hidePriceTinyInt == 1
	order.RequirePhoneContact = requirePhoneContactTinyInt == 1
	order.IsIsolated = isIsolatedTinyInt == 1
	if weatherInfo.Valid {
		order.WeatherInfo = &weatherInfo.String
	}
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
		       coupon_discount, is_urgent, urgent_fee, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, weather_info, is_isolated, created_at, updated_at
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
		var weatherInfo sql.NullString
		var isUrgentTinyInt, hidePriceTinyInt, trustReceiptTinyInt, requirePhoneContactTinyInt, isIsolatedTinyInt int

		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &order.GoodsAmount, &order.DeliveryFee,
			&order.PointsDiscount, &order.CouponDiscount, &isUrgentTinyInt, &order.UrgentFee, &order.TotalAmount, &order.Remark,
			&order.OutOfStockStrategy, &trustReceiptTinyInt, &hidePriceTinyInt, &requirePhoneContactTinyInt,
			&expectedDelivery, &weatherInfo, &isIsolatedTinyInt, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		order.IsUrgent = isUrgentTinyInt == 1
		order.TrustReceipt = trustReceiptTinyInt == 1
		order.HidePrice = hidePriceTinyInt == 1
		order.RequirePhoneContact = requirePhoneContactTinyInt == 1
		order.IsIsolated = isIsolatedTinyInt == 1
		if expectedDelivery.Valid {
			t := expectedDelivery.Time
			order.ExpectedDeliveryAt = &t
		}
		if weatherInfo.Valid {
			order.WeatherInfo = &weatherInfo.String
		}

		orders = append(orders, order)
	}

	return orders, total, nil
}

// GetOrderByID 根据ID获取订单详情
func GetOrderByID(id int) (*Order, error) {
	var order Order
	var expectedDelivery sql.NullTime
	var weatherInfo sql.NullString

	query := `
		SELECT id, order_number, user_id, address_id, status, delivery_employee_code, goods_amount, delivery_fee, points_discount,
		       coupon_discount, is_urgent, urgent_fee, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, weather_info, is_isolated, 
		       is_locked, locked_by, locked_at, created_at, updated_at
		FROM orders WHERE id = ?
	`
	var isUrgentTinyInt, hidePriceTinyInt, trustReceiptTinyInt, requirePhoneContactTinyInt, isIsolatedTinyInt, isLockedTinyInt int
	var deliveryEmployeeCode, lockedBy sql.NullString
	var lockedAt sql.NullTime
	err := database.DB.QueryRow(query, id).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &deliveryEmployeeCode, &order.GoodsAmount, &order.DeliveryFee,
		&order.PointsDiscount, &order.CouponDiscount, &isUrgentTinyInt, &order.UrgentFee, &order.TotalAmount, &order.Remark,
		&order.OutOfStockStrategy, &trustReceiptTinyInt, &hidePriceTinyInt, &requirePhoneContactTinyInt,
		&expectedDelivery, &weatherInfo, &isIsolatedTinyInt, &isLockedTinyInt, &lockedBy, &lockedAt, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	order.IsUrgent = isUrgentTinyInt == 1
	order.TrustReceipt = trustReceiptTinyInt == 1
	order.HidePrice = hidePriceTinyInt == 1
	order.RequirePhoneContact = requirePhoneContactTinyInt == 1
	order.IsIsolated = isIsolatedTinyInt == 1
	order.IsLocked = isLockedTinyInt == 1
	if expectedDelivery.Valid {
		t := expectedDelivery.Time
		order.ExpectedDeliveryAt = &t
	}
	if weatherInfo.Valid {
		order.WeatherInfo = &weatherInfo.String
	}
	if deliveryEmployeeCode.Valid {
		order.DeliveryEmployeeCode = &deliveryEmployeeCode.String
	}
	if lockedBy.Valid {
		order.LockedBy = &lockedBy.String
	}
	if lockedAt.Valid {
		t := lockedAt.Time
		order.LockedAt = &t
	}

	return &order, nil
}

// GetOrderItemsByOrderID 根据订单ID获取订单明细
func GetOrderItemsByOrderID(orderID int) ([]OrderItem, error) {
	var items []OrderItem

	query := `
		SELECT id, order_id, product_id, product_name, spec_name, quantity, unit_price, subtotal, image, is_picked
		FROM order_items WHERE order_id = ? ORDER BY id
	`
	rows, err := database.DB.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		var isPickedTinyInt int
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.SpecName,
			&item.Quantity, &item.UnitPrice, &item.Subtotal, &item.Image, &isPickedTinyInt,
		)
		if err != nil {
			return nil, err
		}
		item.IsPicked = isPickedTinyInt == 1
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
	// 先检查字段是否存在，如果不存在则不查询这些字段
	var hasLockFields bool
	checkLockFieldsQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME IN ('is_locked', 'locked_by')`
	var lockFieldsCount int
	if err := database.DB.QueryRow(checkLockFieldsQuery).Scan(&lockFieldsCount); err == nil {
		hasLockFields = lockFieldsCount >= 2
	}

	var query string
	if hasLockFields {
		query = `
		SELECT
			o.id,
			o.order_number,
			o.status,
			o.total_amount,
			o.created_at,
				o.is_locked,
				o.locked_by,
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
	} else {
		// 字段不存在时，使用默认值
		query = `
			SELECT
				o.id,
				o.order_number,
				o.status,
				o.total_amount,
				o.created_at,
				0 AS is_locked,
				NULL AS locked_by,
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
	}
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
			isLockedTinyInt int
			lockedBy     sql.NullString
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
			&isLockedTinyInt,
			&lockedBy,
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
			"is_locked":     isLockedTinyInt == 1,
			"locked_by":     getStringValue(lockedBy),
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

// GetOrderItemCountByOrderID 根据订单ID获取订单商品数量（总件数）
func GetOrderItemCountByOrderID(orderID int) (int, error) {
	var count sql.NullInt64
	query := `SELECT COALESCE(SUM(quantity), 0) FROM order_items WHERE order_id = ?`
	err := database.DB.QueryRow(query, orderID).Scan(&count)
	if err != nil {
		return 0, err
	}
	if !count.Valid {
		return 0, nil
	}
	return int(count.Int64), nil
}

// GetOrderItemCountsByOrderIDs 批量获取订单商品数量
func GetOrderItemCountsByOrderIDs(orderIDs []int) (map[int]int, error) {
	if len(orderIDs) == 0 {
		return make(map[int]int), nil
	}

	// 构建 IN 查询
	placeholders := make([]string, len(orderIDs))
	args := make([]interface{}, len(orderIDs))
	for i, id := range orderIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT order_id, COALESCE(SUM(quantity), 0) as count
		FROM order_items
		WHERE order_id IN (%s)
		GROUP BY order_id`, strings.Join(placeholders, ","))

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[int]int)
	for rows.Next() {
		var orderID, count int
		if err := rows.Scan(&orderID, &count); err != nil {
			continue
		}
		counts[orderID] = count
	}

	return counts, nil
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
	if err != nil {
		return err
	}

	// 更新订单状态后，重新计算配送费和利润（异步执行，避免阻塞）
	go func() {
		_ = CalculateAndStoreOrderProfit(orderID)
	}()

	return nil
}

// UpdateOrderStatusWithDeliveryEmployee 更新订单状态并记录配送员信息
func UpdateOrderStatusWithDeliveryEmployee(orderID int, newStatus string, deliveryEmployeeCode string) error {
	// 在更新时检查订单是否被锁定，防止在修改期间被接单
	query := `
		UPDATE orders 
		SET status = ?, delivery_employee_code = ?, updated_at = NOW() 
		WHERE id = ? 
		  AND (is_locked = 0 OR is_locked IS NULL)
	`
	result, err := database.DB.Exec(query, newStatus, deliveryEmployeeCode, orderID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// 检查订单是否被锁定
		var isLockedTinyInt int
		err := database.DB.QueryRow("SELECT is_locked FROM orders WHERE id = ?", orderID).Scan(&isLockedTinyInt)
		if err == nil {
			if isLockedTinyInt == 1 {
				return fmt.Errorf("订单正在被修改中，暂时无法接单")
			}
		}
		return fmt.Errorf("更新订单状态失败，订单可能不存在或已被锁定")
	}

	// 更新订单状态后，重新计算配送费和利润（异步执行，避免阻塞）
	go func() {
		_ = CalculateAndStoreOrderProfit(orderID)
	}()

	return nil
}

// LockOrder 锁定订单（防止修改时被接单）
// 使用数据库行锁确保原子性
func LockOrder(orderID int, employeeCode string) error {
	// 使用 SELECT ... FOR UPDATE 获取行锁，然后更新
	// 如果订单已被锁定，返回错误
	// 如果锁定超过5分钟，自动解锁（防止异常情况）
	query := `
		UPDATE orders 
		SET is_locked = 1, locked_by = ?, locked_at = NOW() 
		WHERE id = ? 
		  AND (is_locked = 0 OR is_locked IS NULL OR locked_at < DATE_SUB(NOW(), INTERVAL 5 MINUTE))
		  AND status IN ('pending_delivery', 'pending')
	`
	result, err := database.DB.Exec(query, employeeCode, orderID)
	if err != nil {
		return fmt.Errorf("锁定订单失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取锁定结果失败: %v", err)
	}

	if rowsAffected == 0 {
		// 检查订单是否存在且状态正确
		var status string
		var isLocked bool
		err := database.DB.QueryRow("SELECT status, is_locked FROM orders WHERE id = ?", orderID).Scan(&status, &isLocked)
		if err != nil {
			return fmt.Errorf("订单不存在")
		}
		if status != "pending_delivery" && status != "pending" {
			return fmt.Errorf("订单状态不允许修改（当前状态：%s）", status)
		}
		if isLocked {
			return fmt.Errorf("订单正在被其他员工修改中，请稍后再试")
		}
		return fmt.Errorf("锁定订单失败，请重试")
	}

	return nil
}

// UnlockOrder 解锁订单
func UnlockOrder(orderID int, employeeCode string) error {
	// 只有锁定者才能解锁
	query := `
		UPDATE orders 
		SET is_locked = 0, locked_by = NULL, locked_at = NULL 
		WHERE id = ? AND locked_by = ?
	`
	_, err := database.DB.Exec(query, orderID, employeeCode)
	if err != nil {
		return fmt.Errorf("解锁订单失败: %v", err)
	}
	return nil
}

// UnlockOrderForce 强制解锁订单（用于超时或异常情况）
func UnlockOrderForce(orderID int) error {
	query := `
		UPDATE orders 
		SET is_locked = 0, locked_by = NULL, locked_at = NULL 
		WHERE id = ?
	`
	_, err := database.DB.Exec(query, orderID)
	if err != nil {
		return fmt.Errorf("强制解锁订单失败: %v", err)
	}
	return nil
}
