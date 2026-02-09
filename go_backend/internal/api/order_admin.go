package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetAllOrdersForAdmin 获取所有订单（后台管理）
func GetAllOrdersForAdmin(c *gin.Context) {
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	keyword := c.Query("keyword")
	status := c.Query("status")
	deliveryEmployeeIDsStr := c.Query("delivery_employee_ids") // 配送员ID列表（逗号分隔）
	startDate := c.Query("start_date")                           // 开始日期（YYYY-MM-DD）
	endDate := c.Query("end_date")                              // 结束日期（YYYY-MM-DD）

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 处理配送员ID筛选
	var deliveryEmployeeCodes []string
	if deliveryEmployeeIDsStr != "" {
		// 解析配送员ID列表
		idStrs := strings.Split(deliveryEmployeeIDsStr, ",")
		employeeIDs := make([]int, 0, len(idStrs))
		for _, idStr := range idStrs {
			idStr = strings.TrimSpace(idStr)
			if idStr != "" {
				if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
					employeeIDs = append(employeeIDs, id)
				}
			}
		}

		// 根据ID获取员工码
		if len(employeeIDs) > 0 {
			deliveryEmployeeCodes = make([]string, 0, len(employeeIDs))
			for _, employeeID := range employeeIDs {
				employee, err := model.GetEmployeeByID(employeeID)
				if err == nil && employee != nil && employee.IsDelivery {
					deliveryEmployeeCodes = append(deliveryEmployeeCodes, employee.EmployeeCode)
				}
			}
		}
	}

	orders, total, err := model.GetOrdersWithPaginationAdvanced(pageNum, pageSize, keyword, status, deliveryEmployeeCodes, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单列表失败: " + err.Error()})
		return
	}

	// 如果没有订单，直接返回空列表
	if len(orders) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"list":  []map[string]interface{}{},
				"total": 0,
			},
			"message": "获取成功",
		})
		return
	}

	// 批量获取用户ID和地址ID
	userIDs := make([]int, 0, len(orders))
	addressIDs := make([]int, 0, len(orders))
	orderIDs := make([]int, 0, len(orders))
	userIDSet := make(map[int]bool)
	addressIDSet := make(map[int]bool)

	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
		if !userIDSet[order.UserID] {
			userIDs = append(userIDs, order.UserID)
			userIDSet[order.UserID] = true
		}
		if !addressIDSet[order.AddressID] {
			addressIDs = append(addressIDs, order.AddressID)
			addressIDSet[order.AddressID] = true
		}
	}

	// 批量查询用户信息
	usersMap, _ := model.GetMiniAppUsersByIDs(userIDs)

	// 批量查询地址信息
	addressesMap, _ := model.GetAddressesByIDs(addressIDs)

	// 批量查询订单商品数量
	itemCountsMap, _ := model.GetOrderItemCountsByOrderIDs(orderIDs)

	// 收集所有销售员代码
	salesCodes := make([]string, 0)
	salesCodeSet := make(map[string]bool)
	for _, user := range usersMap {
		if user != nil && user.SalesCode != "" && !salesCodeSet[user.SalesCode] {
			salesCodes = append(salesCodes, user.SalesCode)
			salesCodeSet[user.SalesCode] = true
		}
	}

	// 批量查询销售员信息
	employeesMap, _ := model.GetEmployeesByEmployeeCodes(salesCodes)

	// 收集所有配送员代码
	deliveryCodes := make([]string, 0)
	deliveryCodeSet := make(map[string]bool)
	for _, order := range orders {
		if order.DeliveryEmployeeCode != nil && *order.DeliveryEmployeeCode != "" && !deliveryCodeSet[*order.DeliveryEmployeeCode] {
			deliveryCodes = append(deliveryCodes, *order.DeliveryEmployeeCode)
			deliveryCodeSet[*order.DeliveryEmployeeCode] = true
		}
	}

	// 批量查询配送员信息
	deliveryEmployeesMap, _ := model.GetEmployeesByEmployeeCodes(deliveryCodes)

	// 批量查询销售分成信息（仅已收款订单）
	salesCommissionsMap := make(map[int]*model.SalesCommission)
	paidOrderIDs := make([]int, 0)
	for _, order := range orders {
		if order.Status == "paid" {
			paidOrderIDs = append(paidOrderIDs, order.ID)
		}
	}
	if len(paidOrderIDs) > 0 {
		commissions, _ := model.GetSalesCommissionsByOrderIDs(paidOrderIDs)
		for _, commission := range commissions {
			salesCommissionsMap[commission.OrderID] = commission
		}
	}

	// 组装订单数据
	ordersWithDetails := make([]map[string]interface{}, 0, len(orders))
	for _, order := range orders {
		orderData := map[string]interface{}{
			"id":                    order.ID,
			"order_number":          order.OrderNumber,
			"user_id":               order.UserID,
			"address_id":            order.AddressID,
			"status":                order.Status,
			"payment_method":        order.PaymentMethod,
			"paid_at":               order.PaidAt,
			"wechat_transaction_id": order.WechatTransactionID,
			"goods_amount":          order.GoodsAmount,
			"delivery_fee":          order.DeliveryFee,
			"points_discount":       order.PointsDiscount,
			"coupon_discount":       order.CouponDiscount,
			"is_urgent":             order.IsUrgent,
			"urgent_fee":            order.UrgentFee,
			"total_amount":          order.TotalAmount,
			"remark":                order.Remark,
			"out_of_stock_strategy": order.OutOfStockStrategy,
			"trust_receipt":         order.TrustReceipt,
			"hide_price":            order.HidePrice,
			"require_phone_contact": order.RequirePhoneContact,
			"created_at":            order.CreatedAt,
			"updated_at":            order.UpdatedAt,
		}

		// 获取用户信息（从批量查询结果中获取）
		if user, ok := usersMap[order.UserID]; ok && user != nil {
			userData := map[string]interface{}{
				"id":         user.ID,
				"user_code":  user.UserCode,
				"name":       user.Name,
				"phone":      user.Phone,
				"user_type":  user.UserType,
				"sales_code": user.SalesCode,
			}

			// 获取销售员信息（从批量查询结果中获取）
			if user.SalesCode != "" {
				if employee, ok := employeesMap[user.SalesCode]; ok && employee != nil {
					userData["sales_employee"] = map[string]interface{}{
						"id":            employee.ID,
						"employee_code": employee.EmployeeCode,
						"name":          employee.Name,
						"phone":         employee.Phone,
					}
				}
			}

			orderData["user"] = userData
		}

		// 获取订单商品数量（从批量查询结果中获取）
		if itemCount, ok := itemCountsMap[order.ID]; ok {
			orderData["item_count"] = itemCount
		} else {
			orderData["item_count"] = 0
		}

		// 获取地址信息（从批量查询结果中获取）
		if address, ok := addressesMap[order.AddressID]; ok && address != nil {
			orderData["address"] = map[string]interface{}{
				"id":      address.ID,
				"name":    address.Name,
				"contact": address.Contact,
				"phone":   address.Phone,
				"address": address.Address,
			}
		}

		// 从订单表中读取已存储的配送费计算结果和利润信息（避免实时计算）
		// 如果订单中没有存储，则尝试计算（这种情况应该很少，因为订单创建时会计算）
		var deliveryFeeCalcJSON sql.NullString
		var orderProfit, netProfit sql.NullFloat64
		err = database.DB.QueryRow(`
			SELECT delivery_fee_calculation, order_profit, net_profit
			FROM orders WHERE id = ?
		`, order.ID).Scan(&deliveryFeeCalcJSON, &orderProfit, &netProfit)

		var deliveryFeeResult *model.DeliveryFeeCalculationResult
		var orderProfitVal, netProfitVal float64

		if err == nil {
			// 解析配送费计算结果JSON
			if deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
				var result model.DeliveryFeeCalculationResult
				if json.Unmarshal([]byte(deliveryFeeCalcJSON.String), &result) == nil {
					deliveryFeeResult = &result
					orderData["delivery_fee_calculation"] = result
				}
			}

			// 读取利润信息
			if orderProfit.Valid {
				orderProfitVal = orderProfit.Float64
				orderData["order_profit"] = orderProfitVal
			}
			if netProfit.Valid {
				netProfitVal = netProfit.Float64
				orderData["net_profit"] = netProfitVal
			}
		}

		// 如果订单中没有存储的数据，则尝试计算（这种情况应该很少）
		if deliveryFeeResult == nil {
			calculator, calcErr := model.NewDeliveryFeeCalculator(order.ID)
			if calcErr == nil {
				result, calcErr := calculator.Calculate(true) // true表示管理员视图
				if calcErr == nil {
					deliveryFeeResult = result
					orderData["delivery_fee_calculation"] = result

					// 计算订单利润
					orderProfitVal = calculator.CalculateOrderProfit()
					orderData["order_profit"] = orderProfitVal

					// 计算减去配送费后的利润（平台实际利润）
					netProfitVal = orderProfitVal - result.TotalPlatformCost
					orderData["net_profit"] = netProfitVal

					// 异步存储计算结果，下次查询时可以直接使用
					go func() {
						_ = model.CalculateAndStoreOrderProfitWithRetry(order.ID, 3)
					}()
				}
			}
		}

		// 简化利润计算（更直观的方式）
		if deliveryFeeResult != nil {
			// 平台总收入 = 实付金额
			platformRevenue := order.TotalAmount
			// 商品总成本 = 商品总金额 - 订单利润
			goodsCost := order.GoodsAmount - orderProfitVal
			// 毛利润 = 平台总收入 - 商品总成本
			grossProfit := platformRevenue - goodsCost
			// 配送成本 = 配送员实际所得
			deliveryCost := deliveryFeeResult.TotalPlatformCost
			// 净利润 = 平台总收入 - 商品总成本 - 配送成本
			netProfitSimplified := platformRevenue - goodsCost - deliveryCost

			orderData["simplified_profit"] = map[string]interface{}{
				"platform_revenue": platformRevenue,
				"goods_cost":       goodsCost,
				"gross_profit":     grossProfit,
				"delivery_cost":    deliveryCost,
				"net_profit":       netProfitSimplified,
			}
		}

		// 添加销售分成信息
		// 所有订单都计算预览分成，已收款订单额外显示已计入的分成
		if user, ok := usersMap[order.UserID]; ok && user != nil && user.SalesCode != "" {
			// 计算预览分成（所有订单都显示）
			previewCommission := calculateSalesCommissionPreview(&order, deliveryFeeResult, orderProfitVal)
			if previewCommission != nil {
				orderData["sales_commission_preview"] = previewCommission
			}
		}

		// 已收款订单：从数据库查询已计入的分成
		if order.Status == "paid" {
			if commission, ok := salesCommissionsMap[order.ID]; ok && commission != nil {
				orderData["sales_commission"] = map[string]interface{}{
					"total_commission":   commission.TotalCommission,
					"base_commission":    commission.BaseCommission,
					"new_customer_bonus": commission.NewCustomerBonus,
					"tier_commission":    commission.TierCommission,
					"is_valid_order":     commission.IsValidOrder,
					"is_settled":         commission.SettlementDate != nil,
				}
			} else {
				// 已收款但没有分成记录（可能是历史订单或没有销售员）
				orderData["sales_commission"] = nil
			}
		}

		// 添加配送员信息和配送费
		if order.DeliveryEmployeeCode != nil && *order.DeliveryEmployeeCode != "" {
			if deliveryEmployee, ok := deliveryEmployeesMap[*order.DeliveryEmployeeCode]; ok && deliveryEmployee != nil {
				orderData["delivery_employee"] = map[string]interface{}{
					"id":            deliveryEmployee.ID,
					"employee_code": deliveryEmployee.EmployeeCode,
					"name":          deliveryEmployee.Name,
					"phone":         deliveryEmployee.Phone,
				}
			} else {
				// 配送员信息不存在，只显示员工码
				orderData["delivery_employee"] = map[string]interface{}{
					"employee_code": *order.DeliveryEmployeeCode,
				}
			}

			// 从配送费计算结果中提取配送员实际所得
			if deliveryFeeResult != nil && deliveryFeeResult.RiderPayableFee > 0 {
				orderData["rider_payable_fee"] = deliveryFeeResult.RiderPayableFee
			}
		}

		ordersWithDetails = append(ordersWithDetails, orderData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  ordersWithDetails,
			"total": total,
		},
		"message": "获取成功",
	})
}

// GetOrderByIDForAdmin 获取订单详情（后台管理）
func GetOrderByIDForAdmin(c *gin.Context) {
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

	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单详情失败: " + err.Error()})
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

	// 获取用户信息
	user, _ := model.GetMiniAppUserByID(order.UserID)
	userData := map[string]interface{}{}
	if user != nil {
		userData = map[string]interface{}{
			"id":        user.ID,
			"user_code": user.UserCode,
			"name":      user.Name,
			"phone":     user.Phone,
			"user_type": user.UserType,
			"sales_code": user.SalesCode,
		}

		// 获取销售员信息
		if user.SalesCode != "" {
			employee, err := model.GetEmployeeByEmployeeCode(user.SalesCode)
			if err == nil && employee != nil {
				userData["sales_employee"] = map[string]interface{}{
					"id":            employee.ID,
					"employee_code": employee.EmployeeCode,
					"name":          employee.Name,
					"phone":         employee.Phone,
				}
			}
		}
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

	// 从订单表中读取已存储的配送费计算结果和利润信息（避免实时计算）
	deliveryFeeCalculation := map[string]interface{}{}
	var orderProfit, netProfit float64

	var deliveryFeeCalcJSON sql.NullString
	var orderProfitVal, netProfitVal sql.NullFloat64
	err = database.DB.QueryRow(`
		SELECT delivery_fee_calculation, order_profit, net_profit
		FROM orders WHERE id = ?
	`, id).Scan(&deliveryFeeCalcJSON, &orderProfitVal, &netProfitVal)

	var deliveryFeeResult *model.DeliveryFeeCalculationResult
	if err == nil {
		// 解析配送费计算结果JSON
		if deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
			var result model.DeliveryFeeCalculationResult
			if json.Unmarshal([]byte(deliveryFeeCalcJSON.String), &result) == nil {
				deliveryFeeResult = &result
				deliveryFeeCalculation = map[string]interface{}{
					"base_fee":                    result.BaseFee,
					"isolated_fee":                result.IsolatedFee,
					"item_fee":                    result.ItemFee,
					"urgent_fee":                  result.UrgentFee,
					"weather_fee":                 result.WeatherFee,
					"delivery_fee_without_profit": result.DeliveryFeeWithoutProfit,
					"profit_share":                result.ProfitShare,
					"rider_payable_fee":           result.RiderPayableFee,
					"total_platform_cost":         result.TotalPlatformCost,
				}
			}
		}

		// 读取利润信息
		if orderProfitVal.Valid {
			orderProfit = orderProfitVal.Float64
		}
		if netProfitVal.Valid {
			netProfit = netProfitVal.Float64
		}
	}

	// 如果订单中没有存储的数据，则尝试计算（这种情况应该很少）
	if deliveryFeeResult == nil {
		calculator, calcErr := model.NewDeliveryFeeCalculator(id)
		if calcErr == nil {
			result, calcErr := calculator.Calculate(true) // true表示管理员视图
			if calcErr == nil {
				deliveryFeeResult = result
				deliveryFeeCalculation = map[string]interface{}{
					"base_fee":                    result.BaseFee,
					"isolated_fee":                result.IsolatedFee,
					"item_fee":                    result.ItemFee,
					"urgent_fee":                  result.UrgentFee,
					"weather_fee":                 result.WeatherFee,
					"delivery_fee_without_profit": result.DeliveryFeeWithoutProfit,
					"profit_share":                result.ProfitShare,
					"rider_payable_fee":           result.RiderPayableFee,
					"total_platform_cost":         result.TotalPlatformCost,
				}

				// 计算订单利润
				orderProfit = calculator.CalculateOrderProfit()

				// 计算减去配送费后的利润（平台实际利润）
				netProfit = orderProfit - result.TotalPlatformCost

				// 异步存储计算结果，下次查询时可以直接使用
				go func() {
					_ = model.CalculateAndStoreOrderProfitWithRetry(id, 3)
				}()
			}
		}
	}

	// 简化利润计算（更直观的方式）
	// 平台总收入 = 实付金额（已包含所有收入和扣减）
	// 商品总成本 = 商品总金额 - 订单利润
	// 毛利润 = 平台总收入 - 商品总成本
	// 配送成本 = 配送员实际所得
	// 净利润 = 平台总收入 - 商品总成本 - 配送成本 = 毛利润 - 配送成本
	var simplifiedProfit map[string]interface{}
	if deliveryFeeResult != nil {
		// 平台总收入 = 实付金额（商品金额 + 配送费 + 加急费 - 优惠券抵扣 - 积分抵扣）
		platformRevenue := order.TotalAmount
		// 商品总成本 = 商品总金额 - 订单利润
		goodsCost := order.GoodsAmount - orderProfit
		// 毛利润 = 平台总收入 - 商品总成本
		grossProfit := platformRevenue - goodsCost
		// 配送成本 = 配送员实际所得
		deliveryCost := deliveryFeeResult.TotalPlatformCost
		// 净利润 = 平台总收入 - 商品总成本 - 配送成本
		netProfitSimplified := platformRevenue - goodsCost - deliveryCost

		// 简化的利润分析
		simplifiedProfit = map[string]interface{}{
			"platform_revenue": platformRevenue,     // 平台总收入（实付金额）
			"goods_cost":       goodsCost,           // 商品总成本
			"gross_profit":     grossProfit,         // 毛利润
			"delivery_cost":    deliveryCost,        // 配送成本
			"net_profit":       netProfitSimplified, // 净利润
		}
	}

	result := gin.H{
			"order":                    order,
			"order_items":              items,
			"user":                     userData,
			"address":                  addressData,
			"delivery_fee_calculation": deliveryFeeCalculation,
			"order_profit":             orderProfit,
			"net_profit":               netProfit,
			"simplified_profit":        simplifiedProfit, // 简化的利润分析（平台总收入、商品总成本、毛利润、配送成本、净利润）
	}

	// 添加销售分成信息
	if user != nil && user.SalesCode != "" {
		// 计算预览分成（所有订单都显示）
		previewCommission := calculateSalesCommissionPreview(order, deliveryFeeResult, orderProfit)
		if previewCommission != nil {
			result["sales_commission_preview"] = previewCommission
		}

		// 已收款订单：从数据库查询已计入的分成
		if order.Status == "paid" {
			commissions, err := model.GetSalesCommissionsByOrderIDs([]int{id})
			if err == nil && len(commissions) > 0 && commissions[0] != nil {
				commission := commissions[0]
				result["sales_commission"] = map[string]interface{}{
					"total_commission":      commission.TotalCommission,
					"base_commission":       commission.BaseCommission,
					"new_customer_bonus":    commission.NewCustomerBonus,
					"tier_commission":       commission.TierCommission,
					"tier_level":            commission.TierLevel,
					"is_valid_order":        commission.IsValidOrder,
					"is_new_customer_order": commission.IsNewCustomerOrder,
					"is_settled":            commission.SettlementDate != nil,
					"settlement_date":       commission.SettlementDate,
					"order_profit":          commission.OrderProfit,
					"order_amount":          commission.OrderAmount,
					"goods_cost":               commission.GoodsCost,
					"delivery_cost":         commission.DeliveryCost,
					"calculation_month":     commission.CalculationMonth,
				}
			} else {
				result["sales_commission"] = nil
			}
		}
	}

	// 获取配送记录（包含图片）
	deliveryRecord, err := model.GetDeliveryRecordByOrderID(id)
	if err == nil && deliveryRecord != nil {
		deliveryRecordData := map[string]interface{}{
			"id":                     deliveryRecord.ID,
			"order_id":               deliveryRecord.OrderID,
			"delivery_employee_code": deliveryRecord.DeliveryEmployeeCode,
			"completed_at":           deliveryRecord.CompletedAt,
			"created_at":             deliveryRecord.CreatedAt,
			"updated_at":             deliveryRecord.UpdatedAt,
		}
		if deliveryRecord.ProductImageURL != nil {
			deliveryRecordData["product_image_url"] = *deliveryRecord.ProductImageURL
		}
		if deliveryRecord.DoorplateImageURL != nil {
			deliveryRecordData["doorplate_image_url"] = *deliveryRecord.DoorplateImageURL
		}
		result["delivery_record"] = deliveryRecordData
	}

	// 获取配送日志
	deliveryLogs, err := model.GetDeliveryLogsByOrderID(id)
	if err == nil && len(deliveryLogs) > 0 {
		logsData := make([]map[string]interface{}, 0, len(deliveryLogs))
		for _, log := range deliveryLogs {
			logData := map[string]interface{}{
				"id":          log.ID,
				"order_id":    log.OrderID,
				"action":      log.Action,
				"action_time": log.ActionTime,
				"created_at":  log.CreatedAt,
			}
			if log.DeliveryEmployeeCode != nil {
				logData["delivery_employee_code"] = *log.DeliveryEmployeeCode
			}
			if log.Remark != nil {
				logData["remark"] = *log.Remark
			}
			logsData = append(logsData, logData)
		}
		result["delivery_logs"] = logsData
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    result,
		"message": "获取成功",
	})
}

// UpdateOrderStatus 更新订单状态
func UpdateOrderStatus(c *gin.Context) {
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

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 验证状态值
	validStatuses := map[string]bool{
		"pending_delivery": true,
		"pending_pickup":   true,
		"delivering":       true,
		"delivered":        true,
		"paid":             true,
		"cancelled":        true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的订单状态"})
		return
	}

	// 检查订单是否存在
	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败: " + err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	// 验证状态流转是否合法
	if !isValidStatusTransition(order.Status, req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "状态流转不合法"})
		return
	}

	// 更新订单状态
	err = model.UpdateOrderStatus(id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新订单状态失败: " + err.Error()})
		return
	}

	// 如果订单状态变为 cancelled（已取消），清理该订单的分成记录
	// 特别是新客订单记录，这样其他订单就可以重新计算新客激励了
	if req.Status == "cancelled" {
		// 异步处理，避免阻塞
		go func(orderID int) {
			if err := model.CancelOrderCommissions(orderID); err != nil {
				log.Printf("取消订单 %d 的分成记录失败: %v", orderID, err)
			} else {
				log.Printf("订单 %d 的分成记录已清理", orderID)
			}
		}(id)
	}

	// 如果订单状态变为 paid（已收款），需要计算销售分成和处理推荐奖励
	if req.Status == "paid" {
		// 异步处理销售分成计算和推荐奖励（避免阻塞）
		go func(orderID int) {
			// 等待一下，确保订单状态和结算日期已更新
			time.Sleep(100 * time.Millisecond)

			// 重新获取订单信息，检查是否有结算日期
			order, err := model.GetOrderByID(orderID)
			if err != nil {
				log.Printf("获取订单 %d 信息失败: %v", orderID, err)
				return
			}
			if order == nil {
				log.Printf("订单 %d 不存在", orderID)
				return
			}

			// 如果订单有结算日期，计算销售分成
			// 如果没有结算日期，设置结算日期为当前时间
			if order.SettlementDate == nil {
				now := time.Now()
				_, err = database.DB.Exec("UPDATE orders SET settlement_date = ? WHERE id = ?", now, orderID)
				if err != nil {
					log.Printf("设置订单 %d 结算日期失败: %v", orderID, err)
					return
				}
				// 更新订单对象的结算日期
				order.SettlementDate = &now
			}

			// 计算销售分成
			if err := model.ProcessOrderSettlement(orderID); err != nil {
				log.Printf("处理订单 %d 的销售分成失败: %v", orderID, err)
			} else {
				log.Printf("订单 %d 的销售分成计算成功", orderID)
			}

			// 处理推荐奖励（订单完成付款后发放奖励给老用户）
			if err := model.ProcessReferralReward(orderID); err != nil {
				log.Printf("处理订单 %d 的推荐奖励失败: %v", orderID, err)
			} else {
				log.Printf("订单 %d 的推荐奖励处理完成", orderID)
			}

			// 处理订单积分奖励（订单完成付款后发放积分）
			// 积分规则：每消费1元奖励1积分，四舍五入
			if err := model.AddPointsForOrder(order.UserID, orderID, order.OrderNumber, order.TotalAmount); err != nil {
				log.Printf("处理订单 %d 的积分奖励失败: %v", orderID, err)
			} else {
				log.Printf("订单 %d 的积分奖励处理完成", orderID)
			}
		}(id)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// isValidStatusTransition 验证状态流转是否合法
// 流程：pending_delivery -> pending_pickup -> delivering -> delivered -> paid
// 可以取消：pending_delivery, pending_pickup, delivering -> cancelled
// 兼容旧状态：pending 视为 pending_delivery
func isValidStatusTransition(currentStatus, newStatus string) bool {
	// 将旧状态 pending 转换为新状态 pending_delivery
	if currentStatus == "pending" {
		currentStatus = "pending_delivery"
	}

	// 允许的状态流转（pending_payment 在线支付待支付，支付成功后变为 pending_delivery）
	transitions := map[string][]string{
		"pending_payment":  {"pending_delivery", "cancelled"}, // 待支付：可手动标记已支付进入配送，或取消
		"pending_delivery": {"pending_pickup", "cancelled"},
		"pending_pickup":   {"delivering", "cancelled"},
		"delivering":       {"delivered"}, // 配送中不能取消
		"delivered":        {"paid"},
		"shipped":          {"paid"}, // 兼容旧状态 shipped -> paid
		"paid":             {},       // 已收款是最终状态，不能再流转
		"completed":        {},       // 兼容旧状态 completed，不能再流转
		"cancelled":        {},       // 已取消是最终状态，不能再流转
	}

	allowed, exists := transitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}

	return false
}

// AdminManualRefund 管理员手动退款（用于支付回调未同步等异常场景）
// POST /api/admin/orders/:id/manual-refund
func AdminManualRefund(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败: " + err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	if order.TotalAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单金额为0，无需退款"})
		return
	}

	// 发起微信退款（无论 paid_at 是否已同步，都尝试退款，用于支付回调未同步的场景）
	refundID, refundErr := RequestWechatRefund(order, "支付回调未同步，管理员手动退款")
	if refundErr != nil {
		log.Printf("[AdminManualRefund] 订单 %d 微信退款失败: %v", id, refundErr)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "退款失败: " + refundErr.Error() + "。请确认用户已通过微信支付，或检查微信商户平台",
		})
		return
	}

	// 更新退款状态
	if err := model.RequestWechatRefundForOrder(id, refundID); err != nil {
		log.Printf("[AdminManualRefund] 更新订单退款状态失败: %v", err)
	}

	// 更新订单状态为已取消
	if err := model.UpdateOrderStatus(id, "cancelled"); err != nil {
		log.Printf("[AdminManualRefund] 更新订单状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "退款已受理，但更新订单状态失败"})
		return
	}

	// 清理分成记录
	go func(orderID int) {
		if err := model.CancelOrderCommissions(orderID); err != nil {
			log.Printf("手动退款-取消订单 %d 的分成记录失败: %v", orderID, err)
		}
	}(id)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "退款已受理，预计1-3工作日到账。订单已取消。",
	})
}

// AdminRefundWithDetailsReq 售后退款请求
type AdminRefundWithDetailsReq struct {
	RefundAmount float64 `json:"refund_amount"` // 退款金额（元），0 或空表示全额
	Reason       string  `json:"reason"`        // 退款原因和详情描述
	CancelOrder  bool    `json:"cancel_order"`  // 是否同时取消订单（仅全额退款时有效）
}

// AdminRefundWithDetails 售后退款：支持指定金额、自定义原因，可选取消订单
// POST /api/admin/orders/:id/refund-with-details
func AdminRefundWithDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	var req AdminRefundWithDetailsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败: " + err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	if order.TotalAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单金额为0，无需退款"})
		return
	}

	// 货到付款且未通过微信支付的订单无法退款（避免无效 API 调用）
	if order.PaymentMethod == "cod" && (order.WechatTransactionID == nil || strings.TrimSpace(*order.WechatTransactionID) == "") {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该订单为货到付款且未通过微信支付，无法退款。如有疑问请联系客服。",
		})
		return
	}

	refundAmount := req.RefundAmount
	if refundAmount <= 0 {
		refundAmount = order.TotalAmount
	}
	if refundAmount > order.TotalAmount {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("退款金额不能超过订单实付金额 ¥%.2f", order.TotalAmount)})
		return
	}

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "售后退款"
	}

	refundID, refundErr := RequestWechatRefundWithOptions(order, RefundOptions{
		RefundAmount: refundAmount,
		Reason:       reason,
	})
	if refundErr != nil {
		log.Printf("[AdminRefundWithDetails] 订单 %d 微信退款失败: %v", id, refundErr)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "退款失败: " + refundErr.Error(),
		})
		return
	}

	if err := model.RequestWechatRefundForOrder(id, refundID); err != nil {
		log.Printf("[AdminRefundWithDetails] 更新订单退款状态失败: %v", err)
	}

	// 全额退款且勾选了取消订单时，更新状态为已取消
	if req.CancelOrder && refundAmount >= order.TotalAmount-0.01 {
		if err := model.UpdateOrderStatus(id, "cancelled"); err != nil {
			log.Printf("[AdminRefundWithDetails] 更新订单状态失败: %v", err)
		} else {
			go func(orderID int) {
				if err := model.CancelOrderCommissions(orderID); err != nil {
					log.Printf("售后退款-取消订单 %d 的分成记录失败: %v", orderID, err)
				}
			}(id)
		}
	}

	msg := "退款已受理，预计1-3工作日到账"
	if req.CancelOrder && refundAmount >= order.TotalAmount-0.01 {
		msg += "。订单已取消。"
	}
	msg += "。"

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": msg,
	})
}

// RecalculateOrderProfit 强制重新计算订单利润（用于修复老订单）
func RecalculateOrderProfit(c *gin.Context) {
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

	// 检查订单是否存在
	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败: " + err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	// 获取订单明细，用于诊断
	items, err := model.GetOrderItemsByOrderID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单明细失败: " + err.Error()})
		return
	}

	// 诊断信息：检查每个商品的成本计算情况
	diagnosis := make([]map[string]interface{}, 0, len(items))
	totalCost := 0.0
	for _, item := range items {
		itemInfo := map[string]interface{}{
			"order_item_id": item.ID,
			"product_id":    item.ProductID,
			"product_name":  item.ProductName,
			"spec_name":      item.SpecName,
			"quantity":       item.Quantity,
		}

		var cost float64
		var costSource string

		// 优先从订单快照中获取成本
		if item.SpecSnapshot != nil {
			cost = item.SpecSnapshot.Cost
			costSource = "快照"
			itemInfo["has_snapshot"] = true
			itemInfo["snapshot_cost"] = cost
		} else {
			itemInfo["has_snapshot"] = false
			// 如果没有快照，从商品规格JSON中获取成本（兼容旧订单）
			var specsJSON sql.NullString
			err := database.DB.QueryRow(`SELECT specs FROM products WHERE id = ?`, item.ProductID).Scan(&specsJSON)
			if err != nil {
				itemInfo["error"] = "商品不存在或已删除: " + err.Error()
				itemInfo["cost"] = 0
				diagnosis = append(diagnosis, itemInfo)
				continue
			}

			if !specsJSON.Valid {
				itemInfo["error"] = "商品规格JSON为空"
				itemInfo["cost"] = 0
				diagnosis = append(diagnosis, itemInfo)
				continue
			}

			// 解析规格JSON
			var specs []struct {
				Name           string  `json:"name"`
				WholesalePrice float64 `json:"wholesale_price"`
				RetailPrice    float64 `json:"retail_price"`
				Cost           float64 `json:"cost"`
			}
			if err := json.Unmarshal([]byte(specsJSON.String), &specs); err != nil {
				itemInfo["error"] = "解析规格JSON失败: " + err.Error()
				itemInfo["cost"] = 0
				diagnosis = append(diagnosis, itemInfo)
				continue
			}

			itemInfo["available_specs"] = len(specs)
			itemInfo["spec_names"] = make([]string, 0, len(specs))
			for _, spec := range specs {
				itemInfo["spec_names"] = append(itemInfo["spec_names"].([]string), spec.Name)
			}

			// 查找匹配的规格
			matched := false
			for _, spec := range specs {
				if spec.Name == item.SpecName {
					cost = spec.Cost
					costSource = "规格匹配"
					matched = true
					itemInfo["matched_spec_name"] = spec.Name
					break
				}
			}

			// 如果没找到匹配的规格，使用第一个规格的成本价作为回退
			if !matched && len(specs) > 0 {
				cost = specs[0].Cost
				costSource = "回退到第一个规格"
				itemInfo["fallback_spec_name"] = specs[0].Name
				itemInfo["warning"] = fmt.Sprintf("规格名称 '%s' 不匹配，使用第一个规格 '%s' 的成本", item.SpecName, specs[0].Name)
			} else if !matched {
				itemInfo["error"] = "未找到匹配的规格，且没有可用规格"
				itemInfo["cost"] = 0
				diagnosis = append(diagnosis, itemInfo)
				continue
			}
		}

		if cost < 0 {
			cost = 0
		}

		itemCost := cost * float64(item.Quantity)
		totalCost += itemCost

		itemInfo["cost"] = cost
		itemInfo["cost_source"] = costSource
		itemInfo["item_cost"] = itemCost
		diagnosis = append(diagnosis, itemInfo)
	}

	// 计算利润
	profit := order.GoodsAmount - totalCost
	if profit < 0 {
		profit = 0
	}

	// 强制重新计算并存储（忽略已接单的限制）
	calculator, err := model.NewDeliveryFeeCalculator(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      500,
			"message":   "创建计算器失败: " + err.Error(),
			"diagnosis": diagnosis,
		})
		return
	}

	// 计算订单利润
	calculatedProfit := calculator.CalculateOrderProfit()

	// 强制重新计算配送费和利润（即使已接单也重新计算）
	err = model.CalculateAndStoreOrderProfitWithCalculator(calculator, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      500,
			"message":   "重新计算失败: " + err.Error(),
			"diagnosis": diagnosis,
			"calculated_profit": calculatedProfit,
			"total_cost":        totalCost,
		})
		return
	}

	// 重新获取订单，获取更新后的利润和净利润
	var storedProfit, storedNetProfit sql.NullFloat64
	err = database.DB.QueryRow(`
		SELECT order_profit, net_profit
		FROM orders WHERE id = ?
	`, id).Scan(&storedProfit, &storedNetProfit)

	if err == nil {
		var profitVal, netProfitVal *float64
		if storedProfit.Valid {
			profit := storedProfit.Float64
			profitVal = &profit
		}
		if storedNetProfit.Valid {
			netProfit := storedNetProfit.Float64
			netProfitVal = &netProfit
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "重新计算成功",
			"data": map[string]interface{}{
				"order_id":          id,
				"goods_amount":      order.GoodsAmount,
				"calculated_cost":   totalCost,
				"calculated_profit": calculatedProfit,
				"stored_profit":     profitVal,
				"stored_net_profit": netProfitVal,
			},
			"diagnosis": diagnosis,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "重新计算完成，但获取更新后的订单失败: " + err.Error(),
			"data": map[string]interface{}{
				"order_id":          id,
				"goods_amount":      order.GoodsAmount,
				"calculated_cost":   totalCost,
				"calculated_profit": calculatedProfit,
			},
			"diagnosis": diagnosis,
		})
	}
}

// calculateSalesCommissionPreview 计算销售分成预览（未收款订单）
func calculateSalesCommissionPreview(order *model.Order, deliveryFeeResult *model.DeliveryFeeCalculationResult, orderProfit float64) map[string]interface{} {
	// 获取用户信息
	user, err := model.GetMiniAppUserByID(order.UserID)
	if err != nil || user == nil || user.SalesCode == "" {
		return nil
	}

	// 计算订单金额、商品成本、配送成本
	orderAmount := order.TotalAmount             // 平台总收入
	goodsCost := order.GoodsAmount - orderProfit // 商品总成本
	deliveryCost := 0.0
	if deliveryFeeResult != nil {
		deliveryCost = deliveryFeeResult.TotalPlatformCost
	}

	// 对于已收款订单，如果已有销售分成记录，优先从记录中读取新客状态
	// 这样可以确保第一单的新客状态不会因为创建第二单而改变
	var isNewCustomer bool
	if order.Status == "paid" {
		commissions, err := model.GetSalesCommissionsByOrderIDs([]int{order.ID})
		if err == nil && len(commissions) > 0 && commissions[0] != nil {
			// 已有销售分成记录，使用记录中的新客状态
			isNewCustomer = commissions[0].IsNewCustomerOrder
		} else {
			// 没有销售分成记录，重新判断
			isNewCustomer, _ = model.IsNewCustomerOrder(order.UserID, order.ID)
		}
	} else {
		// 未收款订单：检查是否有已计入分成的新客订单记录
		// 如果有，说明该用户已经有新客订单，当前订单不是新客
		query := `
			SELECT COUNT(*) 
			FROM sales_commissions sc
			INNER JOIN orders o ON sc.order_id = o.id
			WHERE sc.user_id = ?
			  AND sc.order_id != ?
			  AND sc.is_new_customer_order = 1
			  AND sc.is_accounted = 1
			  AND sc.is_accounted_cancelled = 0
			  AND o.status != 'cancelled'
		`
		var count int
		err := database.DB.QueryRow(query, order.UserID, order.ID).Scan(&count)
		if err == nil && count > 0 {
			// 已有新客订单记录，当前订单不是新客
			isNewCustomer = false
		} else {
			// 没有新客订单记录，检查是否是第一个订单
			// 检查该订单是否是该用户的第一个订单（基于订单ID，排除取消的）
			checkQuery := `
				SELECT COUNT(*) 
				FROM orders o
				WHERE o.user_id = ?
				  AND o.id < ?
				  AND o.status != 'cancelled'
			`
			var orderCount int
			err = database.DB.QueryRow(checkQuery, order.UserID, order.ID).Scan(&orderCount)
			if err == nil && orderCount == 0 {
				// 这是第一个订单，是新客
				isNewCustomer = true
			} else {
				// 不是第一个订单，不是新客
				isNewCustomer = false
			}
		}
	}

	// 确定计算月份：
	// 1. 如果订单已结算，使用结算月份（与已计入的分成保持一致）
	// 2. 如果订单未结算，使用当前月份
	var calculationMonth string
	if order.Status == "paid" && order.SettlementDate != nil {
		// 已收款订单：使用结算月份，确保预览分成与已计入分成一致
		calculationMonth = order.SettlementDate.Format("2006-01")
	} else {
		// 未收款订单：使用当前月份
		calculationMonth = time.Now().Format("2006-01")
	}

	// 获取指定月份的有效订单总金额（用于计算阶梯提成）
	monthTotalSales, _ := model.GetMonthlyTotalSales(user.SalesCode, calculationMonth)

	// 计算分成
	calcResult, err := model.CalculateSalesCommission(
		user.SalesCode,
		orderAmount,
		goodsCost,
		deliveryCost,
		isNewCustomer,
		monthTotalSales,
	)
	if err != nil {
		return nil
	}

	return map[string]interface{}{
		"total_commission":      calcResult.TotalCommission,
		"base_commission":       calcResult.BaseCommission,
		"new_customer_bonus":    calcResult.NewCustomerBonus,
		"tier_commission":       calcResult.TierCommission,
		"tier_level":            calcResult.TierLevel,
		"is_valid_order":        calcResult.IsValidOrder,
		"is_new_customer_order": calcResult.IsNewCustomerOrder,
		"is_preview":            true,
		"calculation_month":     calculationMonth,
	}
}
