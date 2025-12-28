package api

import (
	"database/sql"
	"encoding/json"
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
						_ = model.CalculateAndStoreOrderProfit(order.ID)
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
					_ = model.CalculateAndStoreOrderProfit(id)
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
					"goods_cost":            commission.GoodsCost,
					"delivery_cost":         commission.DeliveryCost,
					"calculation_month":     commission.CalculationMonth,
				}
			} else {
				result["sales_commission"] = nil
			}
		}
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

	// 允许的状态流转
	transitions := map[string][]string{
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

	// 判断是否新客户首单（正确查询数据库）
	isNewCustomer, err := model.IsNewCustomerOrder(order.UserID, order.ID)
	if err != nil {
		log.Printf("判断是否新客户失败: %v", err)
		isNewCustomer = false
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
