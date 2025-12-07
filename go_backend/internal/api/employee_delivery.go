package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetDeliveryOrders 获取待配送订单列表（配送员）
func GetDeliveryOrders(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	// 验证是否是配送员
	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	status := c.Query("status") // 可选：pending_delivery, delivering, delivered

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 构建查询条件
	where := "status IN ('pending_delivery', 'pending')" // 待配送订单
	args := []interface{}{}

	// 如果指定了状态，使用指定状态
	if status != "" {
		if status == "pending_delivery" || status == "pending" {
			where = "(status = ? OR status = 'pending')"
			args = append(args, "pending_delivery")
		} else if status == "delivering" {
			where = "status = ?"
			args = append(args, status)
		} else if status == "delivered" || status == "shipped" {
			where = "(status = ? OR status = 'shipped')"
			args = append(args, "delivered")
		} else {
			where = "status = ?"
			args = append(args, status)
		}
	}

	// 获取总数量
	var total int
	countQuery := "SELECT COUNT(*) FROM orders WHERE " + where
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

	// 获取分页数据
	query := `
		SELECT id, order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount,
		       coupon_discount, is_urgent, urgent_fee, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, weather_info, is_isolated, created_at, updated_at
		FROM orders WHERE ` + where + ` ORDER BY created_at ASC LIMIT ? OFFSET ?`
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
		var isUrgentTinyInt, trustReceiptTinyInt, hidePriceTinyInt, requirePhoneContactTinyInt, isIsolatedTinyInt int

		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &order.GoodsAmount, &order.DeliveryFee,
			&order.PointsDiscount, &order.CouponDiscount, &isUrgentTinyInt, &order.UrgentFee, &order.TotalAmount, &order.Remark,
			&order.OutOfStockStrategy, &trustReceiptTinyInt, &hidePriceTinyInt, &requirePhoneContactTinyInt,
			&expectedDelivery, &weatherInfo, &isIsolatedTinyInt, &order.CreatedAt, &order.UpdatedAt,
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

		// 从订单表中读取已存储的配送费计算结果（配送员视图）
		var deliveryFeeCalcJSON sql.NullString
		err = database.DB.QueryRow(`
			SELECT delivery_fee_calculation
			FROM orders WHERE id = ?
		`, order.ID).Scan(&deliveryFeeCalcJSON)

		var deliveryFeeResult model.DeliveryFeeCalculationResult
		var riderUrgentFee float64 // 配送员能拿到的加急费用

		if err == nil && deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
			if json.Unmarshal([]byte(deliveryFeeCalcJSON.String), &deliveryFeeResult) == nil {
				// 配送员能拿到的加急费用是配送费计算中的urgent_fee
				riderUrgentFee = deliveryFeeResult.UrgentFee
			}
		} else {
			// 如果订单中没有存储的数据，则尝试计算（这种情况应该很少）
			calculator, calcErr := model.NewDeliveryFeeCalculator(order.ID)
			if calcErr == nil {
				result, calcErr := calculator.Calculate(false) // false表示配送员视图
				if calcErr == nil {
					deliveryFeeResult = *result
					riderUrgentFee = deliveryFeeResult.UrgentFee
					// 异步存储计算结果，下次查询时可以直接使用
					go func() {
						_ = model.CalculateAndStoreOrderProfit(order.ID)
					}()
				}
			}
		}

		// 构建配送费明细map（包含利润提成，但用"绩效奖励"名称）
		deliveryFeeMap := map[string]interface{}{
			"base_fee":                    deliveryFeeResult.BaseFee,
			"isolated_fee":                deliveryFeeResult.IsolatedFee,
			"item_fee":                    deliveryFeeResult.ItemFee,
			"urgent_fee":                  deliveryFeeResult.UrgentFee,
			"weather_fee":                 deliveryFeeResult.WeatherFee,
			"delivery_fee_without_profit": deliveryFeeResult.DeliveryFeeWithoutProfit,
			"performance_bonus":           deliveryFeeResult.ProfitShare, // 利润提成用"绩效奖励"名称
			"rider_payable_fee":           deliveryFeeResult.RiderPayableFee,
			"total_platform_cost":         deliveryFeeResult.TotalPlatformCost,
		}

		orderData := map[string]interface{}{
			"id":                       order.ID,
			"order_number":             order.OrderNumber,
			"status":                   order.Status,
			"goods_amount":             order.GoodsAmount,
			"delivery_fee":             order.DeliveryFee,
			"total_amount":             order.TotalAmount,
			"item_count":               itemCount,
			"address":                  addressData,
			"is_urgent":                order.IsUrgent,
			"urgent_fee":               riderUrgentFee, // 配送员能拿到的加急费用
			"weather_info":             order.WeatherInfo,
			"is_isolated":              order.IsIsolated,
			"created_at":               order.CreatedAt,
			"updated_at":               order.UpdatedAt,
			"delivery_fee_calculation": deliveryFeeMap,
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

// GetDeliveryOrderDetail 获取订单详情（配送员）
func GetDeliveryOrderDetail(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
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
			"id":        address.ID,
			"name":      address.Name,
			"contact":   address.Contact,
			"phone":     address.Phone,
			"address":   address.Address,
			"latitude":  address.Latitude,
			"longitude": address.Longitude,
		}
	}

	// 获取用户信息
	user, _ := model.GetMiniAppUserByID(order.UserID)
	userData := map[string]interface{}{}
	if user != nil {
		userData = map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"phone": user.Phone,
		}
	}

	// 从订单表中读取已存储的配送费计算结果（配送员视图）
	var deliveryFeeCalcJSON sql.NullString
	var deliveryFeeCalculation map[string]interface{}
	err = database.DB.QueryRow(`
		SELECT delivery_fee_calculation
		FROM orders WHERE id = ?
	`, id).Scan(&deliveryFeeCalcJSON)

	var deliveryFeeResult model.DeliveryFeeCalculationResult
	var riderUrgentFee float64 // 配送员能拿到的加急费用

	if err == nil && deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
		if json.Unmarshal([]byte(deliveryFeeCalcJSON.String), &deliveryFeeResult) == nil {
			// 配送员能拿到的加急费用是配送费计算中的urgent_fee
			riderUrgentFee = deliveryFeeResult.UrgentFee
		}
	} else {
		// 如果订单中没有存储的数据，则尝试计算（这种情况应该很少）
		calculator, calcErr := model.NewDeliveryFeeCalculator(id)
		if calcErr == nil {
			result, calcErr := calculator.Calculate(false) // false表示配送员视图
			if calcErr == nil {
				deliveryFeeResult = *result
				riderUrgentFee = deliveryFeeResult.UrgentFee
				// 异步存储计算结果，下次查询时可以直接使用
				go func() {
					_ = model.CalculateAndStoreOrderProfit(id)
				}()
			}
		}
	}

	// 构建配送费明细map（包含利润提成，但用"绩效奖励"名称）
	deliveryFeeCalculation = map[string]interface{}{
		"base_fee":                    deliveryFeeResult.BaseFee,
		"isolated_fee":                deliveryFeeResult.IsolatedFee,
		"item_fee":                    deliveryFeeResult.ItemFee,
		"urgent_fee":                  deliveryFeeResult.UrgentFee,
		"weather_fee":                 deliveryFeeResult.WeatherFee,
		"delivery_fee_without_profit": deliveryFeeResult.DeliveryFeeWithoutProfit,
		"performance_bonus":           deliveryFeeResult.ProfitShare, // 利润提成用"绩效奖励"名称
		"rider_payable_fee":           deliveryFeeResult.RiderPayableFee,
		"total_platform_cost":         deliveryFeeResult.TotalPlatformCost,
	}

	// 配送员能拿到的加急费用
	orderData := map[string]interface{}{
		"id":           order.ID,
		"order_number": order.OrderNumber,
		"status":       order.Status,
		"goods_amount": order.GoodsAmount,
		"delivery_fee": order.DeliveryFee,
		"total_amount": order.TotalAmount,
		"is_urgent":    order.IsUrgent,
		"urgent_fee":   riderUrgentFee, // 配送员能拿到的加急费用
		"remark":       order.Remark,
		"created_at":   order.CreatedAt,
		"updated_at":   order.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"order":                    orderData,
			"order_items":              items,
			"address":                  addressData,
			"user":                     userData,
			"delivery_fee_calculation": deliveryFeeCalculation,
		},
		"message": "获取成功",
	})
}

// AcceptDeliveryOrder 接单（配送员接单）
func AcceptDeliveryOrder(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
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

	// 验证订单状态
	if order.Status != "pending_delivery" && order.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只能接待配送的订单"})
		return
	}

	// 更新订单状态为配送中
	err = model.UpdateOrderStatus(id, "delivering")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "接单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "接单成功",
	})
}

// CompleteDeliveryOrder 完成配送
func CompleteDeliveryOrder(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
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

	// 验证订单状态
	if order.Status != "delivering" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只能完成配送中的订单"})
		return
	}

	// 更新订单状态为已送达
	err = model.UpdateOrderStatus(id, "delivered")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "完成配送失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "配送完成",
	})
}

// ReportOrderIssue 问题上报
func ReportOrderIssue(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	var req struct {
		IssueType    string `json:"issue_type" binding:"required"`  // 问题类型
		Description  string `json:"description" binding:"required"` // 问题描述
		ContactPhone string `json:"contact_phone,omitempty"`        // 联系电话（可选）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
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

	// 验证订单状态（只有配送中的订单可以上报问题）
	if order.Status != "delivering" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只能上报配送中订单的问题"})
		return
	}

	// TODO: 这里可以将问题信息保存到数据库或发送通知
	// 目前先简单返回成功，后续可以扩展为保存到问题表或发送通知给管理员

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "问题上报成功，我们会尽快处理",
	})
}
