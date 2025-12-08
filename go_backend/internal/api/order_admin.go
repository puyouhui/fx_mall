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

// GetAllOrdersForAdmin 获取所有订单（后台管理）
func GetAllOrdersForAdmin(c *gin.Context) {
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	keyword := c.Query("keyword")
	status := c.Query("status")

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	orders, total, err := model.GetOrdersWithPagination(pageNum, pageSize, keyword, status)
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

	// 组装订单数据
	ordersWithDetails := make([]map[string]interface{}, 0, len(orders))
	for _, order := range orders {
		orderData := map[string]interface{}{
			"id":                  order.ID,
			"order_number":        order.OrderNumber,
			"user_id":             order.UserID,
			"address_id":          order.AddressID,
			"status":              order.Status,
			"goods_amount":        order.GoodsAmount,
			"delivery_fee":        order.DeliveryFee,
			"points_discount":     order.PointsDiscount,
			"coupon_discount":     order.CouponDiscount,
			"total_amount":        order.TotalAmount,
			"remark":              order.Remark,
			"out_of_stock_strategy": order.OutOfStockStrategy,
			"trust_receipt":       order.TrustReceipt,
			"hide_price":          order.HidePrice,
			"require_phone_contact": order.RequirePhoneContact,
			"created_at":          order.CreatedAt,
			"updated_at":          order.UpdatedAt,
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
		
		if err == nil {
			// 解析配送费计算结果JSON
			if deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
				var deliveryFeeResult model.DeliveryFeeCalculationResult
				if json.Unmarshal([]byte(deliveryFeeCalcJSON.String), &deliveryFeeResult) == nil {
					orderData["delivery_fee_calculation"] = deliveryFeeResult
				}
			}
			
			// 读取利润信息
			if orderProfit.Valid {
				orderData["order_profit"] = orderProfit.Float64
			}
			if netProfit.Valid {
				orderData["net_profit"] = netProfit.Float64
			}
		}
		
		// 如果订单中没有存储的数据，则尝试计算（这种情况应该很少）
		if _, hasCalc := orderData["delivery_fee_calculation"]; !hasCalc {
			calculator, calcErr := model.NewDeliveryFeeCalculator(order.ID)
			if calcErr == nil {
				deliveryFeeResult, calcErr := calculator.Calculate(true) // true表示管理员视图
				if calcErr == nil {
					orderData["delivery_fee_calculation"] = deliveryFeeResult
					
					// 计算订单利润
					orderProfit := calculator.CalculateOrderProfit()
					orderData["order_profit"] = orderProfit
					
					// 计算减去配送费后的利润（平台实际利润）
					netProfit := orderProfit - deliveryFeeResult.TotalPlatformCost
					orderData["net_profit"] = netProfit
					
					// 异步存储计算结果，下次查询时可以直接使用
					go func() {
						_ = model.CalculateAndStoreOrderProfit(order.ID)
					}()
				}
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
			"id":         user.ID,
			"user_code":  user.UserCode,
			"name":       user.Name,
			"phone":      user.Phone,
			"user_type":  user.UserType,
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
	
	if err == nil {
		// 解析配送费计算结果JSON
		if deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
			var deliveryFeeResult model.DeliveryFeeCalculationResult
			if json.Unmarshal([]byte(deliveryFeeCalcJSON.String), &deliveryFeeResult) == nil {
				deliveryFeeCalculation = map[string]interface{}{
					"base_fee":                  deliveryFeeResult.BaseFee,
					"isolated_fee":             deliveryFeeResult.IsolatedFee,
					"item_fee":                 deliveryFeeResult.ItemFee,
					"urgent_fee":               deliveryFeeResult.UrgentFee,
					"weather_fee":              deliveryFeeResult.WeatherFee,
					"delivery_fee_without_profit": deliveryFeeResult.DeliveryFeeWithoutProfit,
					"profit_share":             deliveryFeeResult.ProfitShare,
					"rider_payable_fee":        deliveryFeeResult.RiderPayableFee,
					"total_platform_cost":      deliveryFeeResult.TotalPlatformCost,
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
	if len(deliveryFeeCalculation) == 0 {
		calculator, calcErr := model.NewDeliveryFeeCalculator(id)
		if calcErr == nil {
			deliveryFeeResult, calcErr := calculator.Calculate(true) // true表示管理员视图
			if calcErr == nil {
				deliveryFeeCalculation = map[string]interface{}{
					"base_fee":                  deliveryFeeResult.BaseFee,
					"isolated_fee":             deliveryFeeResult.IsolatedFee,
					"item_fee":                 deliveryFeeResult.ItemFee,
					"urgent_fee":               deliveryFeeResult.UrgentFee,
					"weather_fee":              deliveryFeeResult.WeatherFee,
					"delivery_fee_without_profit": deliveryFeeResult.DeliveryFeeWithoutProfit,
					"profit_share":             deliveryFeeResult.ProfitShare,
					"rider_payable_fee":        deliveryFeeResult.RiderPayableFee,
					"total_platform_cost":      deliveryFeeResult.TotalPlatformCost,
				}
				
				// 计算订单利润
				orderProfit = calculator.CalculateOrderProfit()
				
				// 计算减去配送费后的利润（平台实际利润）
				netProfit = orderProfit - deliveryFeeResult.TotalPlatformCost
				
				// 异步存储计算结果，下次查询时可以直接使用
				go func() {
					_ = model.CalculateAndStoreOrderProfit(id)
				}()
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"order":                  order,
			"order_items":            items,
			"user":                   userData,
			"address":                addressData,
			"delivery_fee_calculation": deliveryFeeCalculation,
			"order_profit":           orderProfit,
			"net_profit":             netProfit,
		},
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
		"delivering":       {"delivered", "cancelled"},
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

