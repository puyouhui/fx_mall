package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// CreateOrderRequest 小程序创建订单请求体
// 订单金额相关目前由后端根据采购单和配送费重新计算，前端传的仅作展示
type CreateOrderRequest struct {
	AddressID           int     `json:"address_id" binding:"required"`
	Remark              string  `json:"remark"`
	ItemIDs             []int   `json:"item_ids"`              // 只对勾选的采购单项下单
	OutOfStockStrategy  string  `json:"out_of_stock_strategy"` // cancel_item / ship_available / contact_me
	TrustReceipt        bool    `json:"trust_receipt"`         // 信任签收
	HidePrice           bool    `json:"hide_price"`            // 是否隐藏价格
	RequirePhoneContact bool    `json:"require_phone_contact"` // 配送时是否电话联系
	ExpectedDeliveryAt  *string `json:"expected_delivery_at"`  // 预留，暂不解析
	PointsDiscount      float64 `json:"points_discount"`       // 预留：积分抵扣金额
	CouponDiscount      float64 `json:"coupon_discount"`       // 预留：优惠券抵扣金额（当前已在购物车+确认页计算）
	DeliveryCouponID    int     `json:"delivery_coupon_id"`    // 指定免配送费券
	AmountCouponID      int     `json:"amount_coupon_id"`      // 指定金额券
	IsUrgent            bool    `json:"is_urgent"`             // 是否加急订单
}

// CreateOrderFromCart 从当前采购单创建订单
func CreateOrderFromCart(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	if req.AddressID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择收货地址"})
		return
	}

	// 校验地址归属
	address, err := model.GetAddressByID(req.AddressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取地址信息失败: " + err.Error()})
		return
	}
	if address == nil || address.UserID != user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "收货地址无效"})
		return
	}

	// 获取当前采购单
	items, err := model.GetPurchaseListItemsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单为空，无法创建订单"})
		return
	}

	// 只针对指定的 item_ids 下单
	if len(req.ItemIDs) > 0 {
		filter := make(map[int]struct{}, len(req.ItemIDs))
		for _, id := range req.ItemIDs {
			if id > 0 {
				filter[id] = struct{}{}
			}
		}
		filteredItems := make([]model.PurchaseListItem, 0, len(filter))
		for _, item := range items {
			if _, ok := filter[item.ID]; ok {
				filteredItems = append(filteredItems, item)
			}
		}
		items = filteredItems
	}

	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单为空，无法创建订单"})
		return
	}

	// 获取用户类型，默认为零售
	userType := user.UserType
	if userType == "" || userType == "unknown" {
		userType = "retail"
	}

	// 计算配送费和金额汇总
	summary, err := model.CalculateDeliveryFee(items, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	// 计算订单金额和分类信息用于优惠券筛选
	orderAmount := 0.0
	categoryIDSet := make(map[int]struct{})
	productIDs := make([]int, 0, len(items))
	for _, item := range items {
		// 根据用户类型计算商品金额
		var price float64
		if userType == "wholesale" {
			price = item.SpecSnapshot.WholesalePrice
			if price <= 0 {
				price = item.SpecSnapshot.RetailPrice
			}
		} else {
			price = item.SpecSnapshot.RetailPrice
			if price <= 0 {
				price = item.SpecSnapshot.WholesalePrice
			}
		}
		if price <= 0 {
			price = item.SpecSnapshot.Cost
		}
		if price < 0 {
			price = 0
		}
		orderAmount += price * float64(item.Quantity)
		productIDs = append(productIDs, item.ProductID)
	}

	categoryInfo, err := model.FetchProductCategoryInfo(productIDs)
	if err == nil {
		for _, info := range categoryInfo {
			if info.CategoryID > 0 {
				categoryIDSet[info.CategoryID] = struct{}{}
			}
			if info.ParentID > 0 {
				categoryIDSet[info.ParentID] = struct{}{}
			}
		}
	}
	categoryIDs := make([]int, 0, len(categoryIDSet))
	for id := range categoryIDSet {
		categoryIDs = append(categoryIDs, id)
	}

	availableCoupons, err := model.GetAvailableCouponsForPurchaseList(
		user.ID,
		orderAmount,
		categoryIDs,
		summary.DeliveryFee,
		summary.IsFreeShipping,
	)
	if err != nil {
		availableCoupons = []model.AvailableCouponInfo{}
	}

	appliedCombination := model.CalculateCouponCombinationWithSelection(
		availableCoupons,
		orderAmount,
		summary.DeliveryFee,
		summary.IsFreeShipping,
		req.DeliveryCouponID,
		req.AmountCouponID,
	)

	// 获取加急费用（从系统设置）
	urgentFee := 0.0
	if req.IsUrgent {
		urgentFeeStr, err := model.GetSystemSetting("order_urgent_fee")
		if err == nil && urgentFeeStr != "" {
			if fee, parseErr := strconv.ParseFloat(urgentFeeStr, 64); parseErr == nil && fee > 0 {
				urgentFee = fee
			}
		}
	}

	// 使用模型层事务函数创建订单并落库
	options := model.OrderCreationOptions{
		Remark:              req.Remark,
		OutOfStockStrategy:  req.OutOfStockStrategy,
		TrustReceipt:        req.TrustReceipt,
		HidePrice:           req.HidePrice,
		RequirePhoneContact: req.RequirePhoneContact,
		PointsDiscount:      req.PointsDiscount,
		CouponDiscount:      appliedCombination.TotalDiscount,
		IsUrgent:            req.IsUrgent,
		UrgentFee:           urgentFee,
	}

	order, orderItems, err := model.CreateOrderFromPurchaseList(user.ID, req.AddressID, items, summary, options, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建订单失败: " + err.Error()})
		return
	}

	// 创建订单成功后，使用优惠券（标记为已使用并关联订单ID）
	// 处理免配送费券（使用 UserCouponID 精确更新）
	if appliedCombination.DeliveryFeeCoupon != nil && appliedCombination.DeliveryFeeCoupon.UserCouponID > 0 {
		if err := model.UseCouponByUserCouponID(appliedCombination.DeliveryFeeCoupon.UserCouponID, order.ID); err != nil {
			// 如果使用失败，记录错误但不影响订单创建
			log.Printf("标记免配送费券为已使用失败 (用户优惠券ID: %d, 订单ID: %d): %v", appliedCombination.DeliveryFeeCoupon.UserCouponID, order.ID, err)
		} else {
			log.Printf("成功标记免配送费券为已使用 (用户优惠券ID: %d, 订单ID: %d)", appliedCombination.DeliveryFeeCoupon.UserCouponID, order.ID)
		}
	}
	// 处理金额券（使用 UserCouponID 精确更新）
	if appliedCombination.AmountCoupon != nil && appliedCombination.AmountCoupon.UserCouponID > 0 {
		if err := model.UseCouponByUserCouponID(appliedCombination.AmountCoupon.UserCouponID, order.ID); err != nil {
			// 如果使用失败，记录错误但不影响订单创建
			log.Printf("标记金额券为已使用失败 (用户优惠券ID: %d, 订单ID: %d): %v", appliedCombination.AmountCoupon.UserCouponID, order.ID, err)
		} else {
			log.Printf("成功标记金额券为已使用 (用户优惠券ID: %d, 订单ID: %d)", appliedCombination.AmountCoupon.UserCouponID, order.ID)
		}
	}

	// 小程序用户自己下单后，删除已下单的商品（不备份，直接删除）
	// 如果指定了 item_ids，只删除指定的商品；否则删除所有商品（因为使用了所有商品）
	if len(req.ItemIDs) > 0 {
		// 删除指定的商品
		itemIDList := make([]interface{}, 0, len(req.ItemIDs))
		for _, id := range req.ItemIDs {
			if id > 0 {
				itemIDList = append(itemIDList, id)
			}
		}
		if len(itemIDList) > 0 {
			placeholders := strings.Repeat("?,", len(itemIDList))
			placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
			query := fmt.Sprintf("DELETE FROM purchase_list_items WHERE user_id = ? AND id IN (%s)", placeholders)
			args := append([]interface{}{user.ID}, itemIDList...)
			_, err = database.DB.Exec(query, args...)
			if err != nil {
				log.Printf("[CreateOrderFromCart] 删除已下单商品失败: %v", err)
			} else {
				log.Printf("[CreateOrderFromCart] 成功删除已下单商品，删除商品数量: %d", len(itemIDList))
			}
		}
	} else {
		// 如果没有指定 item_ids，删除所有商品（因为创建订单时使用了所有商品）
		_, err = database.DB.Exec("DELETE FROM purchase_list_items WHERE user_id = ?", user.ID)
		if err != nil {
			log.Printf("[CreateOrderFromCart] 清空采购单失败: %v", err)
		} else {
			log.Printf("[CreateOrderFromCart] 成功清空采购单")
		}
	}

	// 返回创建成功的订单概要
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"order":       order,
			"order_items": orderItems,
		},
		"message": "创建订单成功",
	})
}

// GetUserOrders 获取当前用户的订单列表（小程序）
func GetUserOrders(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	status := c.Query("status") // 可选的状态筛选

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 构建查询条件，只查询当前用户的订单
	where := "o.user_id = ?"
	args := []interface{}{user.ID}

	// 状态筛选（兼容旧状态）
	if status != "" {
		// 兼容旧状态：pending_delivery 也包含 pending 和 pending_pickup（待取货也显示在待配送中）
		if status == "pending_delivery" {
			where += " AND (o.status = ? OR o.status = 'pending' OR o.status = 'pending_pickup')"
			args = append(args, status)
		} else if status == "delivered" {
			// 兼容旧状态：delivered 也包含 shipped
			where += " AND (o.status = ? OR o.status = 'shipped')"
			args = append(args, status)
		} else if status == "paid" {
			// 兼容旧状态：paid 也包含 completed
			where += " AND (o.status = ? OR o.status = 'completed')"
			args = append(args, status)
		} else {
			where += " AND o.status = ?"
			args = append(args, status)
		}
	}

	// 获取总数量（需要关联查询以保持 WHERE 条件一致）
	var total int
	countQuery := "SELECT COUNT(*) FROM orders o LEFT JOIN mini_app_addresses a ON o.address_id = a.id WHERE " + where
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单数量失败: " + err.Error()})
		return
	}

	// 计算偏移量
	offset := (pageNum - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// 获取分页数据（关联查询地址表获取地址名称）
	query := `
		SELECT o.id, o.order_number, o.user_id, o.address_id, o.status, o.goods_amount, o.delivery_fee, o.points_discount,
		       o.coupon_discount, o.is_urgent, o.urgent_fee, o.total_amount, o.remark, o.out_of_stock_strategy, o.trust_receipt,
		       o.hide_price, o.require_phone_contact, o.expected_delivery_at, o.weather_info, o.is_isolated, o.created_at, o.updated_at,
		       a.name AS address_name
		FROM orders o
		LEFT JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE ` + where + ` ORDER BY o.id DESC LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	orders := make([]map[string]interface{}, 0)
	for rows.Next() {
		var order model.Order
		var expectedDelivery sql.NullTime
		var weatherInfo sql.NullString
		var addressName sql.NullString
		var isUrgentTinyInt, trustReceiptTinyInt, hidePriceTinyInt, requirePhoneContactTinyInt, isIsolatedTinyInt int

		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &order.GoodsAmount, &order.DeliveryFee,
			&order.PointsDiscount, &order.CouponDiscount, &isUrgentTinyInt, &order.UrgentFee, &order.TotalAmount, &order.Remark,
			&order.OutOfStockStrategy, &trustReceiptTinyInt, &hidePriceTinyInt, &requirePhoneContactTinyInt,
			&expectedDelivery, &weatherInfo, &isIsolatedTinyInt, &order.CreatedAt, &order.UpdatedAt, &addressName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "解析订单数据失败: " + err.Error()})
			return
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

		// 获取订单商品数量
		itemCount, _ := model.GetOrderItemCountByOrderID(order.ID)

		// 获取订单商品列表（用于获取商品图片）
		orderItems, _ := model.GetOrderItemsByOrderID(order.ID)
		orderItemsData := make([]map[string]interface{}, 0)
		for _, item := range orderItems {
			itemData := map[string]interface{}{
				"id":           item.ID,
				"product_id":   item.ProductID,
				"product_name": item.ProductName,
				"spec_name":    item.SpecName,
				"quantity":     item.Quantity,
				"unit_price":   item.UnitPrice,
				"subtotal":     item.Subtotal,
				"image":        item.Image,
			}
			orderItemsData = append(orderItemsData, itemData)
		}

		orderData := map[string]interface{}{
			"id":              order.ID,
			"order_number":    order.OrderNumber,
			"status":          order.Status,
			"goods_amount":    order.GoodsAmount,
			"delivery_fee":    order.DeliveryFee,
			"points_discount": order.PointsDiscount,
			"coupon_discount": order.CouponDiscount,
			"total_amount":    order.TotalAmount,
			"item_count":      itemCount,
			"order_items":     orderItemsData,
			"created_at":      order.CreatedAt,
			"updated_at":      order.UpdatedAt,
		}

		// 添加地址名称
		if addressName.Valid {
			orderData["address_name"] = addressName.String
		}

		orders = append(orders, orderData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  orders,
			"total": total,
		},
		"message": "获取成功",
	})
}

// GetUserOrderDetail 获取订单详情（小程序）
func GetUserOrderDetail(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供订单ID"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	// 获取订单
	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败: " + err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	// 验证订单归属：允许订单创建者或销售员查看
	isOrderOwner := order.UserID == user.ID
	isSalesEmployee := user.IsSalesEmployee && user.SalesEmployeeID != nil

	// 如果是销售员，需要验证该订单是否属于该销售员负责的客户
	if !isOrderOwner && isSalesEmployee {
		// 获取订单创建者信息
		orderUser, err := model.GetMiniAppUserByID(order.UserID)
		if err != nil || orderUser == nil {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问此订单"})
			return
		}
		// 检查订单创建者是否绑定了该销售员
		if orderUser.SalesEmployeeID == nil || *orderUser.SalesEmployeeID != *user.SalesEmployeeID {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问此订单"})
			return
		}
	} else if !isOrderOwner && !isSalesEmployee {
		// 既不是订单创建者，也不是销售员
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问此订单"})
		return
	}

	// 获取订单明细
	items, err := model.GetOrderItemsByOrderID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单明细失败: " + err.Error()})
		return
	}

	// 获取地址信息
	address, _ := model.GetAddressByID(order.AddressID)
	addressData := map[string]interface{}{}
	if address != nil {
		addressData = map[string]interface{}{
			"id":      address.ID,
			"name":    address.Name,
			"contact": address.Contact,
			"phone":   address.Phone,
			"address": address.Address,
		}
	}

	// 获取销售员信息
	salesEmployeeData := map[string]interface{}{}
	if user.SalesCode != "" {
		employee, err := model.GetEmployeeByEmployeeCode(user.SalesCode)
		if err == nil && employee != nil {
			salesEmployeeData = map[string]interface{}{
				"id":            employee.ID,
				"employee_code": employee.EmployeeCode,
				"name":          employee.Name,
				"phone":         employee.Phone,
			}
		}
	}

	// 获取配送员信息（如果订单已被接单）
	var deliveryEmployeeData interface{} = nil
	if order.DeliveryEmployeeCode != nil && *order.DeliveryEmployeeCode != "" {
		deliveryEmployee, err := model.GetEmployeeByEmployeeCode(*order.DeliveryEmployeeCode)
		if err == nil && deliveryEmployee != nil {
			deliveryEmployeeData = map[string]interface{}{
				"id":            deliveryEmployee.ID,
				"employee_code": deliveryEmployee.EmployeeCode,
				"name":          deliveryEmployee.Name,
				"phone":         deliveryEmployee.Phone,
			}
		}
	}

	// 获取地址经纬度（用于地图显示）
	if address != nil {
		addressData["latitude"] = address.Latitude
		addressData["longitude"] = address.Longitude
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"order":             order,
			"order_items":       items,
			"address":           addressData,
			"sales_employee":    salesEmployeeData,
			"delivery_employee": deliveryEmployeeData,
		},
		"message": "获取成功",
	})
}

// CancelUserOrder 小程序用户取消订单
func CancelUserOrder(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供订单ID"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	// 获取订单
	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败: " + err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	// 验证订单归属：只能取消自己的订单
	if order.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权取消该订单"})
		return
	}

	// 验证订单状态是否可以取消（配送员接单之前：pending_delivery 或 pending_pickup）
	if order.Status != "pending_delivery" && order.Status != "pending" && order.Status != "pending_pickup" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单状态不允许取消（只能取消待配送或待取货状态的订单）",
		})
		return
	}

	// 更新订单状态为已取消
	err = model.UpdateOrderStatus(id, "cancelled")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "取消订单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "订单已取消",
	})
}
