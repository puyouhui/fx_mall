package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
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
	DeliveryFeeSettled   bool       `json:"delivery_fee_settled"`             // 配送费是否已结算
	SettlementDate       *time.Time `json:"settlement_date,omitempty"`        // 结算日期
	OrderProfit          *float64   `json:"order_profit,omitempty"`           // 订单利润（商品金额-商品成本）
	IsLocked             bool       `json:"is_locked"`                        // 是否被锁定（修改中）
	LockedBy             *string    `json:"locked_by,omitempty"`              // 锁定者员工码
	LockedAt             *time.Time `json:"locked_at,omitempty"`              // 锁定时间
	PaymentMethod        string     `json:"payment_method"`                   // 支付方式: online-在线支付, cod-货到付款（老数据默认cod）
	PaidAt               *time.Time `json:"paid_at,omitempty"`                // 支付完成时间（老数据为NULL）
	WechatTransactionID  *string    `json:"wechat_transaction_id,omitempty"`  // 微信支付单号（有值表示通过微信支付，取消时可退款）
	RefundStatus         *string    `json:"refund_status,omitempty"`          // 退款状态: NULL-无, processing-处理中, success-成功, failed-失败
	WechatRefundID       *string    `json:"wechat_refund_id,omitempty"`       // 微信退款单号
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// OrderItem 订单明细
type OrderItem struct {
	ID                      int                  `json:"id"`
	OrderID                 int                  `json:"order_id"`
	ProductID               int                  `json:"product_id"`
	ProductName             string               `json:"product_name"`
	SpecName                string               `json:"spec_name"`
	SpecSnapshot            *PurchaseSpecSnapshot `json:"spec_snapshot,omitempty"` // 规格快照（保存下单时的完整规格信息）
	Quantity                int                  `json:"quantity"`
	UnitPrice               float64              `json:"unit_price"`                // 成交单价（改价后的价格）
	Subtotal                float64              `json:"subtotal"`                  // 小计
	Image                   string               `json:"image"`
	IsPicked                bool                 `json:"is_picked"`                 // 是否已取货
	OriginalUnitPrice       *float64             `json:"original_unit_price,omitempty"` // 原始单价（从规格快照获取）
	IsPriceModified         bool                 `json:"is_price_modified"`         // 是否改价
	PriceModificationReason *string              `json:"price_modification_reason,omitempty"` // 改价原因
}

// PriceModificationInfo 改价信息
type PriceModificationInfo struct {
	UnitPrice float64 // 改价后的单价
	Reason    string  // 改价原因
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
	PriceModifications  map[int]PriceModificationInfo // 改价映射（采购单项ID -> 改价信息）
	DeliveryFeeCouponID int                           // 免配送费券的用户优惠券ID（在事务内处理）
	AmountCouponID      int                           // 金额券的用户优惠券ID（在事务内处理）
	PaymentMethod       string                        // 支付方式: online-在线支付, cod-货到付款（默认cod兼容老流程）
}

// GenerateOrderNumber 生成订单编号
// 格式：YYYYMMDDHHmmss + 订单ID后4位（不足补0） + 随机数2位
// 例如：20240101120000012345（20位）
// 注意：此函数需要在获取订单ID后调用
func GenerateOrderNumber(orderID int) string {
	now := time.Now()
	// 日期时间部分：YYYYMMDDHHmmss (14位)
	timePart := now.Format("20060102150405")

	// 订单ID后4位（不足补0）
	orderIDPart := fmt.Sprintf("%04d", orderID%10000)

	// 随机数2位
	rand.Seed(time.Now().UnixNano())
	randomPart := fmt.Sprintf("%02d", rand.Intn(100))

	return timePart + orderIDPart + randomPart
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
	paymentMethod := opts.PaymentMethod
	if paymentMethod == "" || (paymentMethod != "online" && paymentMethod != "cod") {
		paymentMethod = "cod" // 默认为货到付款，兼容老数据
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

	// 在线支付订单初始为待支付，支付成功后才进入配送流程；货到付款直接进入待配送
	initialStatus := "pending_delivery"
	if paymentMethod == "online" {
		initialStatus = "pending_payment"
	}

	// 先插入订单主表（不包含订单编号，使用NULL，兼容老数据）
	// 注意：order_number字段在CREATE TABLE中没有NOT NULL约束，允许NULL值
	res, err := tx.Exec(`
		INSERT INTO orders (
			order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount, coupon_discount, is_urgent, urgent_fee, total_amount,
			remark, out_of_stock_strategy, trust_receipt, hide_price, require_phone_contact, expected_delivery_at,
			payment_method, created_at, updated_at
		) VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, NOW(), NOW())
	`,
		userID, addressID, initialStatus,
		goodsAmount, deliveryFee, pointsDiscount, couponDiscount, boolToTinyInt(isUrgent), urgentFee, totalAmount,
		opts.Remark, outOfStockStrategy, boolToTinyInt(trustReceipt), boolToTinyInt(hidePrice), boolToTinyInt(requirePhoneContact),
		paymentMethod,
	)
	if err != nil {
		return nil, nil, err
	}

	orderID64, err := res.LastInsertId()
	if err != nil {
		return nil, nil, err
	}
	orderID := int(orderID64)

	// 使用订单ID生成唯一的订单编号（包含订单ID后4位，确保唯一性）
	orderNumber := GenerateOrderNumber(orderID)

	// 更新订单编号
	_, err = tx.Exec("UPDATE orders SET order_number = ? WHERE id = ?", orderNumber, orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("更新订单编号失败: %v", err)
	}

	// 插入订单明细
	orderItems := make([]OrderItem, 0, len(items))
	itemStmt, err := tx.Prepare(`
		INSERT INTO order_items (
			order_id, product_id, product_name, spec_name, spec_snapshot, quantity, unit_price, subtotal, image,
			original_unit_price, is_price_modified, price_modification_reason
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = itemStmt.Close()
	}()

	hasPriceModification := false
	for _, it := range items {
		// 计算原始价格（从规格快照获取）
		var originalPrice float64
		if userType == "wholesale" {
			originalPrice = it.SpecSnapshot.WholesalePrice
			if originalPrice <= 0 {
				originalPrice = it.SpecSnapshot.RetailPrice
			}
		} else {
			originalPrice = it.SpecSnapshot.RetailPrice
			if originalPrice <= 0 {
				originalPrice = it.SpecSnapshot.WholesalePrice
			}
		}
		if originalPrice <= 0 {
			originalPrice = it.SpecSnapshot.Cost
		}
		if originalPrice < 0 {
			originalPrice = 0
		}

		// 检查是否有改价
		var price float64
		var isPriceModified bool
		var priceModReason *string
		if opts.PriceModifications != nil {
			if mod, hasMod := opts.PriceModifications[it.ID]; hasMod {
				price = mod.UnitPrice
				if price < 0 {
					price = 0
				}
				isPriceModified = price != originalPrice
				if isPriceModified {
					hasPriceModification = true
					if mod.Reason != "" {
						priceModReason = &mod.Reason
					}
				} else {
					price = originalPrice
				}
			} else {
				price = originalPrice
			}
		} else {
			price = originalPrice
		}

		subtotal := price * float64(it.Quantity)

		// 准备插入SQL，包含改价相关字段
		var originalPricePtr *float64
		if isPriceModified {
			originalPricePtr = &originalPrice
		}

		// 序列化规格快照为JSON
		specSnapshotJSON, err := json.Marshal(it.SpecSnapshot)
		if err != nil {
			return nil, nil, fmt.Errorf("序列化规格快照失败: %v", err)
		}

		if _, err = itemStmt.Exec(
			orderID,
			it.ProductID,
			it.ProductName,
			it.SpecName,
			string(specSnapshotJSON),
			it.Quantity,
			price,
			subtotal,
			it.ProductImage,
			originalPricePtr,
			boolToTinyInt(isPriceModified),
			priceModReason,
		); err != nil {
			return nil, nil, err
		}

		orderItems = append(orderItems, OrderItem{
			OrderID:                 orderID,
			ProductID:               it.ProductID,
			ProductName:             it.ProductName,
			SpecName:                it.SpecName,
			SpecSnapshot:            &it.SpecSnapshot,
			Quantity:                it.Quantity,
			UnitPrice:               price,
			Subtotal:                subtotal,
			Image:                   it.ProductImage,
			OriginalUnitPrice:       originalPricePtr,
			IsPriceModified:         isPriceModified,
			PriceModificationReason: priceModReason,
		})
	}

	// 如果订单包含改价商品，更新订单表的has_price_modification字段
	if hasPriceModification {
		_, err = tx.Exec(`UPDATE orders SET has_price_modification = 1 WHERE id = ?`, orderID)
		if err != nil {
			return nil, nil, err
		}
	}

	// 在事务内处理优惠券使用（确保订单创建和优惠券标记在同一事务中）
	// 处理免配送费券
	if opts.DeliveryFeeCouponID > 0 {
		if err := UseCouponByUserCouponIDInTx(tx, opts.DeliveryFeeCouponID, orderID); err != nil {
			return nil, nil, fmt.Errorf("使用免配送费券失败: %v", err)
		}
	}
	// 处理金额券
	if opts.AmountCouponID > 0 {
		if err := UseCouponByUserCouponIDInTx(tx, opts.AmountCouponID, orderID); err != nil {
			return nil, nil, fmt.Errorf("使用金额券失败: %v", err)
		}
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

	// 检查是否是首次下单，如果是且有referrer_id，记录推荐奖励
	go func() {
		// 检查是否是首次下单（排除当前订单）
		var count int
		checkQuery := `
			SELECT COUNT(*) FROM orders
			WHERE user_id = ? AND id != ? AND status != 'cancelled'
		`
		err := database.DB.QueryRow(checkQuery, userID, orderID).Scan(&count)
		if err == nil && count == 0 {
			// 这是首次下单，获取用户信息，检查是否有referrer_id
			user, err := GetMiniAppUserByID(userID)
			if err == nil && user != nil && user.ReferrerID != nil {
				// 创建推荐奖励记录（待发放状态）
				_ = CreateReferralReward(*user.ReferrerID, userID, orderID, orderNumber)
			}
		}
	}()

	// 计算并存储配送费计算结果和利润信息
	// 注意：这里在事务外执行，避免阻塞订单创建
	go func() {
		// 先更新订单的配送相关信息（孤立状态、天气信息等）
		_ = UpdateOrderDeliveryInfo(orderID)

		// 然后计算并存储配送费计算结果和利润（带重试机制）
		_ = CalculateAndStoreOrderProfitWithRetry(orderID, 3)
	}()

	// 查询刚插入的订单记录
	var order Order
	var expectedDelivery sql.NullTime
	var weatherInfo sql.NullString
	var paidAt sql.NullTime
	var paymentMethodVal string
	var isUrgentTinyInt, hidePriceTinyInt, trustReceiptTinyInt, requirePhoneContactTinyInt, isIsolatedTinyInt int
	err = database.DB.QueryRow(`
		SELECT id, order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount,
		       coupon_discount, is_urgent, urgent_fee, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, weather_info, is_isolated,
		       payment_method, paid_at, created_at, updated_at
		FROM orders WHERE id = ?
	`, orderID).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &order.GoodsAmount, &order.DeliveryFee,
		&order.PointsDiscount, &order.CouponDiscount, &isUrgentTinyInt, &order.UrgentFee, &order.TotalAmount, &order.Remark,
		&order.OutOfStockStrategy, &trustReceiptTinyInt, &hidePriceTinyInt, &requirePhoneContactTinyInt,
		&expectedDelivery, &weatherInfo, &isIsolatedTinyInt,
		&paymentMethodVal, &paidAt, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}
	order.IsUrgent = isUrgentTinyInt == 1
	order.TrustReceipt = trustReceiptTinyInt == 1
	order.HidePrice = hidePriceTinyInt == 1
	order.RequirePhoneContact = requirePhoneContactTinyInt == 1
	order.IsIsolated = isIsolatedTinyInt == 1
	order.PaymentMethod = paymentMethodVal
	if paymentMethodVal == "" {
		order.PaymentMethod = "cod" // 老数据兼容
	}
	if weatherInfo.Valid {
		order.WeatherInfo = &weatherInfo.String
	}
	if paidAt.Valid {
		t := paidAt.Time
		order.PaidAt = &t
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

// CreateOrderFromCachedPrepay 从预支付缓存创建订单（支付回调时调用）
// orderNumber 即 out_trade_no；transactionID 为微信支付单号；直接创建为已支付、待配送状态
func CreateOrderFromCachedPrepay(orderNumber, transactionID string, entry *CachedPrepayEntry) (*Order, []OrderItem, error) {
	if entry == nil || len(entry.Items) == 0 {
		return nil, nil, fmt.Errorf("缓存数据无效")
	}
	if entry.Summary == nil {
		return nil, nil, fmt.Errorf("配送费汇总为空")
	}
	userID := entry.UserID
	addressID := entry.AddressID
	items := entry.Items
	summary := entry.Summary
	opts := entry.Options
	userType := entry.UserType
	if userType == "" {
		userType = "retail"
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
	if !isUrgent {
		urgentFee = 0
	}
	totalAmount := goodsAmount + deliveryFee + urgentFee - pointsDiscount - couponDiscount
	if totalAmount < 0 {
		totalAmount = 0
	}

	now := time.Now()
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 使用与 CreateOrderFromPurchaseList 相同的列顺序，避免列数不匹配错误
	// 先插入基础字段，再 UPDATE 设置 paid_at 和 wechat_transaction_id
	res, err := tx.Exec(`
		INSERT INTO orders (
			order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount, coupon_discount, is_urgent, urgent_fee, total_amount,
			remark, out_of_stock_strategy, trust_receipt, hide_price, require_phone_contact, expected_delivery_at,
			payment_method, created_at, updated_at
		) VALUES (?, ?, ?, 'pending_delivery', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, 'online', NOW(), NOW())
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

	// 立即更新为已支付状态
	_, err = tx.Exec(`UPDATE orders SET paid_at = ?, wechat_transaction_id = ? WHERE id = ?`, now, transactionID, orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("更新支付信息失败: %v", err)
	}

	itemStmt, err := tx.Prepare(`
		INSERT INTO order_items (
			order_id, product_id, product_name, spec_name, spec_snapshot, quantity, unit_price, subtotal, image,
			original_unit_price, is_price_modified, price_modification_reason
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = itemStmt.Close() }()

	orderItems := make([]OrderItem, 0, len(items))
	for _, it := range items {
		var originalPrice float64
		if userType == "wholesale" {
			originalPrice = it.SpecSnapshot.WholesalePrice
			if originalPrice <= 0 {
				originalPrice = it.SpecSnapshot.RetailPrice
			}
		} else {
			originalPrice = it.SpecSnapshot.RetailPrice
			if originalPrice <= 0 {
				originalPrice = it.SpecSnapshot.WholesalePrice
			}
		}
		if originalPrice <= 0 {
			originalPrice = it.SpecSnapshot.Cost
		}
		if originalPrice < 0 {
			originalPrice = 0
		}
		price := originalPrice
		subtotal := price * float64(it.Quantity)
		specSnapshotJSON, _ := json.Marshal(it.SpecSnapshot)
		_, err = itemStmt.Exec(orderID, it.ProductID, it.ProductName, it.SpecName, string(specSnapshotJSON), it.Quantity, price, subtotal, it.ProductImage, &originalPrice, 0, nil)
		if err != nil {
			return nil, nil, err
		}
		orderItems = append(orderItems, OrderItem{
			OrderID: orderID, ProductID: it.ProductID, ProductName: it.ProductName, SpecName: it.SpecName,
			SpecSnapshot: &it.SpecSnapshot, Quantity: it.Quantity, UnitPrice: price, Subtotal: subtotal, Image: it.ProductImage,
		})
	}

	if opts.DeliveryFeeCouponID > 0 {
		if err := UseCouponByUserCouponIDInTx(tx, opts.DeliveryFeeCouponID, orderID); err != nil {
			return nil, nil, fmt.Errorf("使用免配送费券失败: %v", err)
		}
	}
	if opts.AmountCouponID > 0 {
		if err := UseCouponByUserCouponIDInTx(tx, opts.AmountCouponID, orderID); err != nil {
			return nil, nil, fmt.Errorf("使用金额券失败: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, nil, err
	}

	remark := "订单创建（支付成功）"
	_ = CreateDeliveryLog(&DeliveryLog{OrderID: orderID, Action: DeliveryLogActionCreated, ActionTime: now, Remark: &remark})

	go func() {
		var count int
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM orders WHERE user_id = ? AND id != ? AND status != 'cancelled'", userID, orderID).Scan(&count)
		if count == 0 {
			user, _ := GetMiniAppUserByID(userID)
			if user != nil && user.ReferrerID != nil {
				_ = CreateReferralReward(*user.ReferrerID, userID, orderID, orderNumber)
			}
		}
	}()

	go func() {
		_ = UpdateOrderDeliveryInfo(orderID)
		_ = CalculateAndStoreOrderProfitWithRetry(orderID, 3)
	}()

	order, err := GetOrderByID(orderID)
	if err != nil || order == nil {
		return nil, orderItems, fmt.Errorf("订单创建成功但查询失败")
	}
	return order, orderItems, nil
}

// CachedPrepayEntry 预支付缓存条目（model 层结构，与 api.PrepayCacheEntry 对应）
type CachedPrepayEntry struct {
	UserID    int
	AddressID int
	UserType  string
	Items     []PurchaseListItem
	Summary   *DeliveryFeeSummary
	Options   OrderCreationOptions
	ItemIDs   []int
}

// GetOrdersWithPagination 获取订单列表（支持分页和搜索）
func GetOrdersWithPagination(pageNum, pageSize int, keyword string, status string) ([]Order, int, error) {
	return GetOrdersWithPaginationAdvanced(pageNum, pageSize, keyword, status, nil, "", "")
}

// GetOrdersWithPaginationAdvanced 获取订单列表（支持分页、搜索、配送员筛选和日期筛选）
// deliveryEmployeeCodes: 配送员员工码列表（nil表示不筛选）
// startDate: 开始日期（格式：YYYY-MM-DD，空字符串表示不筛选）
// endDate: 结束日期（格式：YYYY-MM-DD，空字符串表示不筛选）
func GetOrdersWithPaginationAdvanced(pageNum, pageSize int, keyword string, status string, deliveryEmployeeCodes []string, startDate, endDate string) ([]Order, int, error) {
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

	// 状态筛选（支持多个状态，用逗号分隔）
	if status != "" {
		statuses := strings.Split(status, ",")
		if len(statuses) == 1 {
		where += " AND status = ?"
			args = append(args, strings.TrimSpace(statuses[0]))
		} else {
			// 多个状态使用 IN 查询
			placeholders := make([]string, len(statuses))
			for i, s := range statuses {
				placeholders[i] = "?"
				args = append(args, strings.TrimSpace(s))
			}
			where += " AND status IN (" + strings.Join(placeholders, ",") + ")"
		}
	}

	// 配送员筛选
	if deliveryEmployeeCodes != nil && len(deliveryEmployeeCodes) > 0 {
		// 过滤空字符串
		validCodes := make([]string, 0, len(deliveryEmployeeCodes))
		for _, code := range deliveryEmployeeCodes {
			trimmed := strings.TrimSpace(code)
			if trimmed != "" {
				validCodes = append(validCodes, trimmed)
			}
		}
		if len(validCodes) > 0 {
			placeholders := make([]string, len(validCodes))
			for i, code := range validCodes {
				placeholders[i] = "?"
				args = append(args, code)
			}
			where += " AND delivery_employee_code IN (" + strings.Join(placeholders, ",") + ")"
		}
	}

	// 日期筛选
	if startDate != "" {
		where += " AND DATE(created_at) >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		where += " AND DATE(created_at) <= ?"
		args = append(args, endDate)
	}

	// 获取总数量
	countQuery := "SELECT COUNT(*) FROM orders WHERE " + where
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	query := `
		SELECT id, order_number, user_id, address_id, status, delivery_employee_code, goods_amount, delivery_fee, points_discount,
		       coupon_discount, is_urgent, urgent_fee, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, weather_info, is_isolated,
		       payment_method, paid_at, created_at, updated_at
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
		var deliveryEmployeeCode sql.NullString
		var paidAt sql.NullTime
		var paymentMethodVal string
		var isUrgentTinyInt, hidePriceTinyInt, trustReceiptTinyInt, requirePhoneContactTinyInt, isIsolatedTinyInt int

		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &deliveryEmployeeCode, &order.GoodsAmount, &order.DeliveryFee,
			&order.PointsDiscount, &order.CouponDiscount, &isUrgentTinyInt, &order.UrgentFee, &order.TotalAmount, &order.Remark,
			&order.OutOfStockStrategy, &trustReceiptTinyInt, &hidePriceTinyInt, &requirePhoneContactTinyInt,
			&expectedDelivery, &weatherInfo, &isIsolatedTinyInt,
			&paymentMethodVal, &paidAt, &order.CreatedAt, &order.UpdatedAt,
		)
		if deliveryEmployeeCode.Valid {
			code := deliveryEmployeeCode.String
			order.DeliveryEmployeeCode = &code
		}
		if err != nil {
			return nil, 0, err
		}
		order.PaymentMethod = paymentMethodVal
		if order.PaymentMethod == "" {
			order.PaymentMethod = "cod"
		}
		if paidAt.Valid {
			t := paidAt.Time
			order.PaidAt = &t
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
	var paidAt sql.NullTime
	var paymentMethodVal string
	var refundStatus, wechatRefundID, wechatTransactionID sql.NullString

	query := `
		SELECT id, order_number, user_id, address_id, status, delivery_employee_code, goods_amount, delivery_fee, points_discount,
		       coupon_discount, is_urgent, urgent_fee, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, weather_info, is_isolated, 
		       is_locked, locked_by, locked_at, order_profit, settlement_date, delivery_fee_settled,
		       payment_method, paid_at, wechat_transaction_id, refund_status, wechat_refund_id, created_at, updated_at
		FROM orders WHERE id = ?
	`
	var isUrgentTinyInt, hidePriceTinyInt, trustReceiptTinyInt, requirePhoneContactTinyInt, isIsolatedTinyInt, isLockedTinyInt, deliveryFeeSettledTinyInt int
	var deliveryEmployeeCode, lockedBy sql.NullString
	var lockedAt, settlementDate sql.NullTime
	var orderProfit sql.NullFloat64
	err := database.DB.QueryRow(query, id).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &deliveryEmployeeCode, &order.GoodsAmount, &order.DeliveryFee,
		&order.PointsDiscount, &order.CouponDiscount, &isUrgentTinyInt, &order.UrgentFee, &order.TotalAmount, &order.Remark,
		&order.OutOfStockStrategy, &trustReceiptTinyInt, &hidePriceTinyInt, &requirePhoneContactTinyInt,
		&expectedDelivery, &weatherInfo, &isIsolatedTinyInt, &isLockedTinyInt, &lockedBy, &lockedAt,
		&orderProfit, &settlementDate, &deliveryFeeSettledTinyInt,
		&paymentMethodVal, &paidAt, &wechatTransactionID, &refundStatus, &wechatRefundID, &order.CreatedAt, &order.UpdatedAt,
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
	order.DeliveryFeeSettled = deliveryFeeSettledTinyInt == 1
	order.PaymentMethod = paymentMethodVal
	if order.PaymentMethod == "" {
		order.PaymentMethod = "cod"
	}
	if paidAt.Valid {
		t := paidAt.Time
		order.PaidAt = &t
	}
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
	if orderProfit.Valid {
		profit := orderProfit.Float64
		order.OrderProfit = &profit
	}
	if settlementDate.Valid {
		t := settlementDate.Time
		order.SettlementDate = &t
	}
	if wechatTransactionID.Valid && wechatTransactionID.String != "" {
		s := wechatTransactionID.String
		order.WechatTransactionID = &s
	}
	if refundStatus.Valid {
		s := refundStatus.String
		order.RefundStatus = &s
	}
	if wechatRefundID.Valid {
		s := wechatRefundID.String
		order.WechatRefundID = &s
	}

	return &order, nil
}

// NeedWechatRefundOnCancel 判断取消订单时是否需要发起微信退款
// 条件：已支付且为微信支付（在线支付或货到付款提前通过去付款支付）
func (o *Order) NeedWechatRefundOnCancel() bool {
	if o.PaidAt == nil || o.TotalAmount <= 0 {
		return false
	}
	// 在线支付订单：只能通过微信支付，一定退款
	if o.PaymentMethod == "online" {
		return true
	}
	// 货到付款订单：仅当有微信支付单号时（用户通过去付款支付）才退款
	return o.WechatTransactionID != nil && *o.WechatTransactionID != ""
}

// GetOrderByOrderNumber 根据订单编号获取订单
func GetOrderByOrderNumber(orderNumber string) (*Order, error) {
	if orderNumber == "" {
		return nil, nil
	}
	var order Order
	var expectedDelivery sql.NullTime
	var weatherInfo sql.NullString
	var paidAt sql.NullTime
	var paymentMethodVal string
	var refundStatus, wechatRefundID, wechatTransactionID sql.NullString

	query := `
		SELECT id, order_number, user_id, address_id, status, delivery_employee_code, goods_amount, delivery_fee, points_discount,
		       coupon_discount, is_urgent, urgent_fee, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, weather_info, is_isolated, 
		       is_locked, locked_by, locked_at, order_profit, settlement_date, delivery_fee_settled,
		       payment_method, paid_at, wechat_transaction_id, refund_status, wechat_refund_id, created_at, updated_at
		FROM orders WHERE order_number = ?
	`
	var isUrgentTinyInt, hidePriceTinyInt, trustReceiptTinyInt, requirePhoneContactTinyInt, isIsolatedTinyInt, isLockedTinyInt, deliveryFeeSettledTinyInt int
	var deliveryEmployeeCode, lockedBy sql.NullString
	var lockedAt, settlementDate sql.NullTime
	var orderProfit sql.NullFloat64
	err := database.DB.QueryRow(query, orderNumber).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &deliveryEmployeeCode, &order.GoodsAmount, &order.DeliveryFee,
		&order.PointsDiscount, &order.CouponDiscount, &isUrgentTinyInt, &order.UrgentFee, &order.TotalAmount, &order.Remark,
		&order.OutOfStockStrategy, &trustReceiptTinyInt, &hidePriceTinyInt, &requirePhoneContactTinyInt,
		&expectedDelivery, &weatherInfo, &isIsolatedTinyInt, &isLockedTinyInt, &lockedBy, &lockedAt,
		&orderProfit, &settlementDate, &deliveryFeeSettledTinyInt,
		&paymentMethodVal, &paidAt, &wechatTransactionID, &refundStatus, &wechatRefundID, &order.CreatedAt, &order.UpdatedAt,
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
	order.DeliveryFeeSettled = deliveryFeeSettledTinyInt == 1
	order.PaymentMethod = paymentMethodVal
	if order.PaymentMethod == "" {
		order.PaymentMethod = "cod"
	}
	if paidAt.Valid {
		t := paidAt.Time
		order.PaidAt = &t
	}
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
	if orderProfit.Valid {
		profit := orderProfit.Float64
		order.OrderProfit = &profit
	}
	if settlementDate.Valid {
		t := settlementDate.Time
		order.SettlementDate = &t
	}
	if wechatTransactionID.Valid && wechatTransactionID.String != "" {
		s := wechatTransactionID.String
		order.WechatTransactionID = &s
	}
	if refundStatus.Valid {
		s := refundStatus.String
		order.RefundStatus = &s
	}
	if wechatRefundID.Valid {
		s := wechatRefundID.String
		order.WechatRefundID = &s
	}

	return &order, nil
}

// MarkOrderPaidByWechatPay 微信支付回调成功后标记订单已支付
// 更新 paid_at，若状态为 delivered/shipped 则设为 paid 并触发结算
func MarkOrderPaidByWechatPay(orderID int, transactionID string) error {
	order, err := GetOrderByID(orderID)
	if err != nil || order == nil {
		return fmt.Errorf("订单不存在")
	}
	if order.PaidAt != nil {
		return nil // 已支付，幂等
	}

	now := time.Now()
	// 更新 paid_at、wechat_transaction_id；若为待支付则进入配送流程(pending_delivery)，若已送达则设为 paid
	_, err = database.DB.Exec(`
		UPDATE orders
		SET paid_at = ?, wechat_transaction_id = ?, status = CASE
			WHEN status IN ('delivered', 'shipped') THEN 'paid'
			WHEN status = 'pending_payment' THEN 'pending_delivery'
			ELSE status
		END, updated_at = NOW()
		WHERE id = ? AND (paid_at IS NULL)
	`, now, transactionID, orderID)
	if err != nil {
		return err
	}

	// 若状态变为 paid，需触发结算（与 order_admin 一致）
	order, _ = GetOrderByID(orderID)
	if order != nil && order.Status == "paid" {
		// 设置 settlement_date
		_, _ = database.DB.Exec("UPDATE orders SET settlement_date = ? WHERE id = ? AND settlement_date IS NULL", now, orderID)
		// 异步执行结算逻辑
		go func(oid int) {
			time.Sleep(100 * time.Millisecond)
			ord, _ := GetOrderByID(oid)
			if ord == nil {
				return
			}
			_ = ProcessOrderSettlement(oid)
			_ = ProcessReferralReward(oid)
			_ = AddPointsForOrder(ord.UserID, oid, ord.OrderNumber, ord.TotalAmount)
		}(orderID)
	}

	return nil
}

// RequestWechatRefundForOrder 发起微信退款后更新订单退款信息
func RequestWechatRefundForOrder(orderID int, wechatRefundID string) error {
	_, err := database.DB.Exec(`
		UPDATE orders
		SET refund_status = 'processing', wechat_refund_id = ?, updated_at = NOW()
		WHERE id = ?
	`, wechatRefundID, orderID)
	return err
}

// MarkOrderRefundSuccess 标记订单退款成功（退款回调或查询确认后调用）
func MarkOrderRefundSuccess(orderID int) error {
	_, err := database.DB.Exec(`
		UPDATE orders SET refund_status = 'success', updated_at = NOW() WHERE id = ?
	`, orderID)
	return err
}

// MarkOrderRefundFailed 标记订单退款失败
func MarkOrderRefundFailed(orderID int) error {
	_, err := database.DB.Exec(`
		UPDATE orders SET refund_status = 'failed', updated_at = NOW() WHERE id = ?
	`, orderID)
	return err
}

// GetOrderItemsByOrderID 根据订单ID获取订单明细
func GetOrderItemsByOrderID(orderID int) ([]OrderItem, error) {
	var items []OrderItem

	// 使用缓存的字段存在性结果（避免每次查询都检查）
	hasSpecSnapshotField := database.HasSpecSnapshotField()

	// 根据字段是否存在构建不同的查询
	var query string
	if hasSpecSnapshotField {
		query = `
			SELECT id, order_id, product_id, product_name, spec_name, spec_snapshot, quantity, unit_price, subtotal, image, is_picked,
			       original_unit_price, is_price_modified, price_modification_reason
			FROM order_items WHERE order_id = ? ORDER BY id
		`
	} else {
		// 兼容老数据：如果字段不存在，不查询该字段
		query = `
			SELECT id, order_id, product_id, product_name, spec_name, NULL as spec_snapshot, quantity, unit_price, subtotal, image, is_picked,
			       original_unit_price, is_price_modified, price_modification_reason
			FROM order_items WHERE order_id = ? ORDER BY id
		`
	}

	rows, err := database.DB.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		var isPickedTinyInt int
		var originalPrice sql.NullFloat64
		var isPriceModifiedTinyInt int
		var priceModReason sql.NullString
		var specSnapshotJSON sql.NullString
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.SpecName, &specSnapshotJSON,
			&item.Quantity, &item.UnitPrice, &item.Subtotal, &item.Image, &isPickedTinyInt,
			&originalPrice, &isPriceModifiedTinyInt, &priceModReason,
		)
		if err != nil {
			return nil, err
		}
		item.IsPicked = isPickedTinyInt == 1
		item.IsPriceModified = isPriceModifiedTinyInt == 1
		if originalPrice.Valid {
			item.OriginalUnitPrice = &originalPrice.Float64
		}
		if priceModReason.Valid && priceModReason.String != "" {
			item.PriceModificationReason = &priceModReason.String
		}
		// 解析规格快照（如果存在）
		if specSnapshotJSON.Valid && specSnapshotJSON.String != "" {
			var snapshot PurchaseSpecSnapshot
			if err := json.Unmarshal([]byte(specSnapshotJSON.String), &snapshot); err == nil {
				item.SpecSnapshot = &snapshot
			}
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
			o.is_urgent,
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
				0 AS is_urgent,
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
			id              int
			orderNumber     string
			statusVal       string
			totalAmount     float64
			createdAt       time.Time
			isLockedTinyInt int
			lockedBy        sql.NullString
			isUrgentTinyInt int
			userName        sql.NullString
			userCode        sql.NullString
			storeName       sql.NullString
			address         sql.NullString
			contactPhone    sql.NullString
		)

		if err := rows.Scan(
			&id,
			&orderNumber,
			&statusVal,
			&totalAmount,
			&createdAt,
			&isLockedTinyInt,
			&lockedBy,
			&isUrgentTinyInt,
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
			"is_urgent":     isUrgentTinyInt == 1,
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

// GetUnpaidOrdersBySalesCode 获取销售员名下客户的未收款订单列表（分页，status != 'paid' AND status != 'cancelled'）
func GetUnpaidOrdersBySalesCode(employeeCode string, pageNum, pageSize int) ([]map[string]interface{}, int, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (pageNum - 1) * pageSize

	// 统计总数（排除已收款和已取消的订单）
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM orders o
		JOIN mini_app_users u ON o.user_id = u.id
		WHERE u.sales_code = ? AND o.status != 'paid' AND o.status != 'cancelled'
	`
	if err := database.DB.QueryRow(countQuery, employeeCode).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询分页数据（关联查询地址名称）
	query := `
		SELECT
			o.id,
			o.order_number,
			o.status,
			o.total_amount,
			o.goods_amount,
			o.delivery_fee,
			o.order_profit,
			o.delivery_fee_calculation,
			o.created_at,
			o.user_id,
			a.name AS address_name
		FROM orders o
		JOIN mini_app_users u ON o.user_id = u.id
		LEFT JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE u.sales_code = ? AND o.status != 'paid' AND o.status != 'cancelled'
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
		var id, userID int
		var orderNumber, status string
		var totalAmount, goodsAmount, deliveryFee float64
		var orderProfit sql.NullFloat64
		var deliveryFeeCalculation sql.NullString
		var createdAt time.Time
		var addressName sql.NullString

		err := rows.Scan(
			&id, &orderNumber, &status, &totalAmount, &goodsAmount,
			&deliveryFee, &orderProfit, &deliveryFeeCalculation, &createdAt, &userID,
			&addressName,
		)
		if err != nil {
			continue
		}

		orderData := map[string]interface{}{
			"id":                      int64(id),
			"order_number":            orderNumber,
			"status":                  status,
			"total_amount":            totalAmount,
			"goods_amount":            goodsAmount,
			"delivery_fee":            deliveryFee,
			"created_at":             createdAt,
			"user_id":                userID,
		}

		if orderProfit.Valid {
			orderData["order_profit"] = orderProfit.Float64
		}
		if deliveryFeeCalculation.Valid {
			orderData["delivery_fee_calculation"] = deliveryFeeCalculation.String
		}
		if addressName.Valid {
			orderData["address_name"] = addressName.String
		}

		orders = append(orders, orderData)
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
			o.is_urgent,
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
		var isUrgentTinyInt int
		var storeName sql.NullString
		var address sql.NullString
		var contactPhone sql.NullString

		if err := rows.Scan(&id, &orderNumber, &status, &totalAmount, &createdAt, &isUrgentTinyInt, &storeName, &address, &contactPhone); err != nil {
			return nil, 0, err
		}

		itemCount, _ := GetOrderItemCountByOrderID(id)

		order := map[string]interface{}{
			"id":            id,
			"order_number":  orderNumber,
			"status":        status,
			"total_amount":  totalAmount,
			"created_at":    createdAt,
			"is_urgent":     isUrgentTinyInt == 1,
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

// GetUncompletedOrdersBySalesCode 获取销售员名下客户的待完成订单列表（分页，收款之前的订单）
// 包括：pending_delivery, pending, pending_pickup, delivering, delivered, shipped
func GetUncompletedOrdersBySalesCode(employeeCode string, pageNum, pageSize int) ([]map[string]interface{}, int, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (pageNum - 1) * pageSize

	// 统计总数（排除已收款和已取消的订单）
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM orders o
		JOIN mini_app_users u ON o.user_id = u.id
		WHERE u.sales_code = ? AND o.status != 'paid' AND o.status != 'cancelled'
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
			o.is_urgent,
			a.name AS store_name,
			a.address,
			a.phone AS contact_phone
		FROM orders o
		JOIN mini_app_users u ON o.user_id = u.id
		LEFT JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE u.sales_code = ? AND o.status != 'paid' AND o.status != 'cancelled'
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
		var isUrgentTinyInt int
		var storeName sql.NullString
		var address sql.NullString
		var contactPhone sql.NullString

		if err := rows.Scan(&id, &orderNumber, &status, &totalAmount, &createdAt, &isUrgentTinyInt, &storeName, &address, &contactPhone); err != nil {
			return nil, 0, err
		}

		itemCount, _ := GetOrderItemCountByOrderID(id)

		order := map[string]interface{}{
			"id":            id,
			"order_number":  orderNumber,
			"status":        status,
			"total_amount":  totalAmount,
			"created_at":    createdAt,
			"is_urgent":     isUrgentTinyInt == 1,
			"store_name":    getStringValue(storeName),
			"address":       getStringValue(address),
			"contact_phone": getStringValue(contactPhone),
			"item_count":    itemCount,
		}

		orders = append(orders, order)
	}

	return orders, total, nil
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
		SELECT o.id, o.order_number, o.total_amount, o.status, o.created_at, o.address_id,
		       a.name as address_name
		FROM orders o
		LEFT JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE o.user_id = ?
		ORDER BY o.created_at DESC, o.id DESC
		LIMIT ?
	`

	rows, err := database.DB.Query(query, userID, limit)
	if err != nil {
		log.Printf("[GetRecentOrdersByUserID] 查询失败: userID=%d, error=%v", userID, err)
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
			addressID   sql.NullInt64
			addressName sql.NullString
		)

		if err := rows.Scan(&id, &orderNumber, &totalAmount, &status, &createdAt, &addressID, &addressName); err != nil {
			log.Printf("[GetRecentOrdersByUserID] 扫描数据失败: userID=%d, error=%v", userID, err)
			return nil, err
		}

		order := map[string]interface{}{
			"id":           id,
			"order_number": orderNumber,
			"total_amount": totalAmount,
			"status":       status,
			"created_at":   createdAt,
		}
		if addressName.Valid {
			order["address_name"] = addressName.String
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// CancelExpiredPendingPaymentOrders 取消超时的待支付订单（供定时任务调用）
func CancelExpiredPendingPaymentOrders() (int, error) {
	timeoutMin, _ := GetSystemSetting("order_pending_payment_timeout")
	if timeoutMin == "" {
		timeoutMin = "15"
	}
	minutes, err := strconv.Atoi(timeoutMin)
	if err != nil || minutes <= 0 {
		minutes = 15
	}
	// 查询超时的待支付订单
	rows, err := database.DB.Query(`
		SELECT id FROM orders
		WHERE status = 'pending_payment'
		  AND created_at < DATE_SUB(NOW(), INTERVAL ? MINUTE)
	`, minutes)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var orderIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		orderIDs = append(orderIDs, id)
	}

	for _, orderID := range orderIDs {
		_ = UpdateOrderStatus(orderID, "cancelled")
		// 清理分成记录
		go func(oid int) {
			_ = CancelOrderCommissions(oid)
		}(orderID)
	}
	return len(orderIDs), nil
}

// UpdateOrderStatus 更新订单状态
func UpdateOrderStatus(orderID int, newStatus string) error {
	query := "UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ?"
	_, err := database.DB.Exec(query, newStatus, orderID)
	if err != nil {
		return err
	}

	// 如果订单被取消，更新受影响订单的孤立状态
	if newStatus == "cancelled" {
		go func() {
			// 更新受影响订单的孤立状态（因为当前订单被取消）
			_ = updateAffectedOrdersIsolatedStatus(orderID)
		}()
	}

	// 更新订单状态后，重新计算配送费和利润（异步执行，避免阻塞）
	go func() {
		_ = CalculateAndStoreOrderProfitWithRetry(orderID, 3)
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
		err := database.DB.QueryRow("SELECT COALESCE(is_locked, 0) FROM orders WHERE id = ?", orderID).Scan(&isLockedTinyInt)
		if err == nil && isLockedTinyInt == 1 {
			return fmt.Errorf("订单已被锁定，无法接单")
		}
		return fmt.Errorf("订单不存在或已被其他配送员接单")
	}

	// 如果状态变为 pending_pickup（接单），立即计算并锁定配送费
	// 使用基于该配送员批次的判断（与预览时一致）
	if newStatus == "pending_pickup" {
		// 立即计算并存储配送费（锁定配送费，确保接单前后一致）
		go func() {
			// 使用基于配送员批次的判断重新计算孤立状态和配送费
			calculator, err := NewDeliveryFeeCalculatorForEmployee(orderID, deliveryEmployeeCode)
			if err == nil {
				// 重新计算孤立状态（基于配送员批次）
				address, err := GetAddressByID(calculator.order.AddressID)
				if err == nil && address != nil && address.Latitude != nil && address.Longitude != nil {
					isolatedDistance := calculator.getConfigFloat("delivery_isolated_distance", 8.0)
					nearbyOrders, err := calculator.getNearbyOrders(*address.Latitude, *address.Longitude, isolatedDistance)
					if err == nil {
						validNearby := calculator.filterNearbyOrders(nearbyOrders)
						isIsolated := len(validNearby) == 0
						
						// 更新孤立状态
						_, _ = database.DB.Exec(`
							UPDATE orders 
							SET is_isolated = ?, updated_at = NOW()
							WHERE id = ?
						`, isIsolated, orderID)
					}
				}
				
				// 计算并锁定配送费
				_ = CalculateAndStoreOrderProfitWithCalculator(calculator, orderID)
			}
			
			// 更新受影响订单的孤立状态（因为当前订单从"未接单"变为"已接单"）
			_ = updateAffectedOrdersIsolatedStatus(orderID)
		}()
	}

	return nil
}

// LockOrder 锁定订单（防止修改时被接单）
// 使用数据库行锁确保原子性
func LockOrder(orderID int, employeeCode string) error {
	// 使用事务确保原子性
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 先检查订单状态和锁定状态
	var status string
	var isLockedTinyInt int
	var lockedBy sql.NullString
	err = tx.QueryRow("SELECT status, is_locked, locked_by FROM orders WHERE id = ? FOR UPDATE", orderID).Scan(&status, &isLockedTinyInt, &lockedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("订单不存在")
		}
		return fmt.Errorf("查询订单失败: %v", err)
	}

	// 检查订单状态
	if status != "pending_payment" && status != "pending_delivery" && status != "pending" {
		return fmt.Errorf("订单状态不允许修改（当前状态：%s）", status)
	}

	// 检查是否已被锁定
	isLocked := isLockedTinyInt == 1
	if isLocked {
		// 如果锁定超过5分钟，自动解锁
		var lockedAt sql.NullTime
		err = tx.QueryRow("SELECT locked_at FROM orders WHERE id = ?", orderID).Scan(&lockedAt)
		if err == nil && lockedAt.Valid {
			// 检查是否超过5分钟
			now := time.Now()
			if now.Sub(lockedAt.Time) < 5*time.Minute {
				// 检查是否被当前员工锁定
				if lockedBy.Valid && lockedBy.String == employeeCode {
					// 被当前员工锁定，更新锁定时间
					_, err = tx.Exec("UPDATE orders SET locked_at = NOW() WHERE id = ?", orderID)
					if err != nil {
						return fmt.Errorf("更新锁定时间失败: %v", err)
					}
					if err = tx.Commit(); err != nil {
						return fmt.Errorf("提交事务失败: %v", err)
					}
					return nil
				} else {
					return fmt.Errorf("订单正在被其他员工修改中，请稍后再试")
				}
			}
			// 超过5分钟，自动解锁并重新锁定
		} else {
			// 没有锁定时间，可能是数据不一致，允许重新锁定
		}
	}

	// 锁定订单
	_, err = tx.Exec(`
		UPDATE orders 
		SET is_locked = 1, locked_by = ?, locked_at = NOW() 
		WHERE id = ?
	`, employeeCode, orderID)
	if err != nil {
		return fmt.Errorf("锁定订单失败: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
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
