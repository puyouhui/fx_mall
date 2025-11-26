package model

import (
	"database/sql"
	"fmt"
	"time"

	"go_backend/internal/database"
)

// Order 订单主表
// 目前先用于创建订单，后续可扩展状态流转、支付等逻辑
type Order struct {
	ID                  int        `json:"id"`
	UserID              int        `json:"user_id"`
	AddressID           int        `json:"address_id"`
	Status              string     `json:"status"`                // pending/paid/shipped/completed/cancelled
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

// CreateOrderFromPurchaseList 从采购单创建订单（包含事务和明细落库）
func CreateOrderFromPurchaseList(userID, addressID int, items []PurchaseListItem, summary *DeliveryFeeSummary, opts OrderCreationOptions) (*Order, []OrderItem, error) {
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
			user_id, address_id, status, goods_amount, delivery_fee, points_discount, coupon_discount, total_amount,
			remark, out_of_stock_strategy, trust_receipt, hide_price, require_phone_contact, expected_delivery_at,
			created_at, updated_at
		) VALUES (?, ?, 'pending', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NOW(), NOW())
	`,
		userID, addressID,
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
		// 使用与配送/购物车一致的价格逻辑
		price := it.SpecSnapshot.WholesalePrice
		if price <= 0 {
			price = it.SpecSnapshot.RetailPrice
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
		SELECT id, user_id, address_id, status, goods_amount, delivery_fee, points_discount,
		       coupon_discount, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, created_at, updated_at
		FROM orders WHERE id = ?
	`, orderID).Scan(
		&order.ID, &order.UserID, &order.AddressID, &order.Status, &order.GoodsAmount, &order.DeliveryFee,
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



