package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// PreviewSalesCommission 预览销售分成（开单时）
func PreviewSalesCommission(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	var req struct {
		OrderAmount  float64 `json:"order_amount"`  // 平台总收入
		GoodsCost    float64 `json:"goods_cost"`    // 商品总成本
		DeliveryCost float64 `json:"delivery_cost"` // 配送成本
		UserID       int     `json:"user_id"`       // 客户用户ID
		OrderID      int     `json:"order_id"`       // 订单ID（可选，用于判断是否新客户）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 验证必需字段（允许0值，但不允许缺失）
	if req.UserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户用户ID不能为空"})
		return
	}

	// 如果订单金额、商品成本、配送成本都为0，说明还没有商品，返回空结果
	if req.OrderAmount == 0 && req.GoodsCost == 0 && req.DeliveryCost == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": map[string]interface{}{
				"total_commission": 0.0,
				"base_commission":  0.0,
				"bonus_commission": 0.0,
				"tier_commission":  0.0,
			},
			"message": "预览成功",
		})
		return
	}

	// 判断是否新客户（如果有order_id，查询该订单之前是否有已结算订单）
	isNewCustomer := false
	if req.OrderID > 0 {
		var err error
		isNewCustomer, err = model.IsNewCustomerOrder(req.UserID, req.OrderID)
		if err != nil {
			// 查询失败，默认为非新客户
			isNewCustomer = false
		}
	} else {
		// 没有order_id，查询该用户是否有已结算订单
		// 如果没有已结算订单，则认为是新客户
		hasSettledOrder, err := model.HasSettledOrder(req.UserID)
		if err == nil && !hasSettledOrder {
			isNewCustomer = true
		}
	}

	// 获取当月有效订单总金额（用于计算阶梯提成）
	currentMonth := time.Now().Format("2006-01")
	monthTotalSales, err := model.GetMonthlyTotalSales(employee.EmployeeCode, currentMonth)
	if err != nil {
		// 查询失败，使用0
		monthTotalSales = 0
	}

	// 计算分成（预览模式，不保存）
	calcResult, err := model.CalculateSalesCommission(
		employee.EmployeeCode,
		req.OrderAmount,
		req.GoodsCost,
		req.DeliveryCost,
		isNewCustomer,
		monthTotalSales,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算分成失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": calcResult,
		"message": "预览成功",
	})
}

// GetSalesCommissions 获取销售员的分成记录列表
func GetSalesCommissions(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	month := c.Query("month") // YYYY-MM格式
	status := c.Query("status") // all, accounted, settled, unaccounted, unsettled
	startDateStr := c.Query("start_date") // YYYY-MM-DD格式
	endDateStr := c.Query("end_date") // YYYY-MM-DD格式

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}

	commissions, total, err := model.GetSalesCommissionsByEmployee(employee.EmployeeCode, month, status, startDate, endDate, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取分成记录失败: " + err.Error()})
		return
	}

	// 为每个分成记录添加地址名称（批量查询优化）
	orderIDs := make([]int, 0)
	orderIDMap := make(map[int]*model.SalesCommission)
	for i := range commissions {
		if commissions[i].OrderID > 0 {
			orderIDs = append(orderIDs, commissions[i].OrderID)
			orderIDMap[commissions[i].OrderID] = &commissions[i]
		}
	}

	// 批量查询订单的地址ID
	addressMap := make(map[int]string) // orderID -> addressName
	if len(orderIDs) > 0 {
		placeholders := ""
		args := make([]interface{}, len(orderIDs))
		for i, id := range orderIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args[i] = id
		}

		query := fmt.Sprintf(`
			SELECT o.id, a.name
			FROM orders o
			LEFT JOIN mini_app_addresses a ON o.address_id = a.id
			WHERE o.id IN (%s)
		`, placeholders)

		rows, err := database.DB.Query(query, args...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var orderID int
				var addressName sql.NullString
				if err := rows.Scan(&orderID, &addressName); err == nil {
					if addressName.Valid {
						addressMap[orderID] = addressName.String
					}
				}
			}
		}
	}

	// 构建返回数据
	commissionsWithAddress := make([]map[string]interface{}, 0, len(commissions))
	for _, commission := range commissions {
		commissionMap := map[string]interface{}{
			"id":                    commission.ID,
			"order_id":              commission.OrderID,
			"employee_code":         commission.EmployeeCode,
			"user_id":               commission.UserID,
			"order_number":          commission.OrderNumber,
			"order_date":            commission.OrderDate,
			"settlement_date":       commission.SettlementDate,
			"is_valid_order":        commission.IsValidOrder,
			"is_new_customer_order": commission.IsNewCustomerOrder,
			"order_amount":          commission.OrderAmount,
			"goods_cost":            commission.GoodsCost,
			"delivery_cost":          commission.DeliveryCost,
			"order_profit":          commission.OrderProfit,
			"base_commission":       commission.BaseCommission,
			"new_customer_bonus":    commission.NewCustomerBonus,
			"tier_commission":       commission.TierCommission,
			"total_commission":      commission.TotalCommission,
			"tier_level":            commission.TierLevel,
			"calculation_month":     commission.CalculationMonth,
			"is_accounted":          commission.IsAccounted,
			"accounted_at":          commission.AccountedAt,
			"is_settled":            commission.IsSettled,
			"settled_at":            commission.SettledAt,
			"is_accounted_cancelled": commission.IsAccountedCancelled,
			"created_at":            commission.CreatedAt,
			"updated_at":            commission.UpdatedAt,
		}

		// 添加地址名称
		if commission.OrderID > 0 {
			if name, ok := addressMap[commission.OrderID]; ok {
				commissionMap["address_name"] = name
			}
		}

		commissionsWithAddress = append(commissionsWithAddress, commissionMap)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  commissionsWithAddress,
			"total": total,
		},
		"message": "获取成功",
	})
}

// GetSalesCommissionMonthlyStats 获取销售员的分成月统计
func GetSalesCommissionMonthlyStats(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	month := c.Query("month") // YYYY-MM格式
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	stats, err := model.GetSalesCommissionMonthlyStats(employee.EmployeeCode, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取月统计失败: " + err.Error()})
		return
	}

	if stats == nil {
		// 没有统计数据，返回空数据
		stats = &model.SalesCommissionMonthlyStats{
			EmployeeCode: employee.EmployeeCode,
			StatMonth:    month,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": stats,
		"message": "获取成功",
	})
}

// GetSalesCommissionConfig 获取销售员的分成配置
func GetSalesCommissionConfig(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	config, err := model.GetSalesCommissionConfig(employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": config,
		"message": "获取成功",
	})
}

// GetSalesCommissionOverview 获取销售员的分成总览统计
func GetSalesCommissionOverview(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	// 获取总览统计
	overview, err := model.GetSalesCommissionOverview(employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取总览统计失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": overview,
		"message": "获取成功",
	})
}

// GetUnpaidOrdersWithCommissionPreview 获取未收款订单及其分润预览（未计入的订单）
func GetUnpaidOrdersWithCommissionPreview(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 获取未收款订单列表（status != 'paid'）
	orders, total, err := model.GetUnpaidOrdersBySalesCode(employee.EmployeeCode, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取未收款订单失败: " + err.Error()})
		return
	}

	// 为每个订单计算分润预览
	result := make([]map[string]interface{}, 0)
	for _, order := range orders {
		orderID := int(order["id"].(int64))
		orderObj, err := model.GetOrderByID(orderID)
		if err != nil || orderObj == nil {
			continue
		}

		// 获取用户信息
		user, err := model.GetMiniAppUserByID(orderObj.UserID)
		if err != nil || user == nil || user.SalesCode != employee.EmployeeCode {
			continue
		}

		// 计算订单利润
		var orderProfit float64
		if orderObj.OrderProfit != nil {
			orderProfit = *orderObj.OrderProfit
		} else {
			// 如果没有order_profit，使用默认值0
			orderProfit = 0
		}

		// 获取配送费计算结果（从查询结果中获取）
		var deliveryFeeResult *model.DeliveryFeeCalculationResult
		if deliveryFeeCalcStr, ok := order["delivery_fee_calculation"].(string); ok && deliveryFeeCalcStr != "" {
			var calcResult model.DeliveryFeeCalculationResult
			if err := json.Unmarshal([]byte(deliveryFeeCalcStr), &calcResult); err == nil {
				deliveryFeeResult = &calcResult
			}
		}

		// 计算分润预览（使用已存在的函数，它会从订单的用户信息中获取salesCode）
		commissionPreview := calculateSalesCommissionPreview(orderObj, deliveryFeeResult, orderProfit)

		// 构建返回数据
		orderData := map[string]interface{}{
			"id":            orderObj.ID,
			"order_number":  orderObj.OrderNumber,
			"order_date":    orderObj.CreatedAt,
			"status":        orderObj.Status,
			"order_amount":  orderObj.TotalAmount,
			"goods_amount":  orderObj.GoodsAmount,
			"delivery_fee":  orderObj.DeliveryFee,
			"total_amount":  orderObj.TotalAmount,
		}

		// 添加地址名称（从查询结果中获取）
		if addressName, ok := order["address_name"].(string); ok && addressName != "" {
			orderData["address_name"] = addressName
		}

		if commissionPreview != nil {
			orderData["commission_preview"] = commissionPreview
		}

		result = append(result, orderData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  result,
			"total": total,
		},
		"message": "获取成功",
	})
}

// AdminGetSalesCommissionStats 管理员获取销售员的分成统计（可查看所有销售员）
func AdminGetSalesCommissionStats(c *gin.Context) {
	// 验证管理员权限
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	employeeCode := c.Query("employee_code")
	month := c.Query("month") // YYYY-MM格式
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	if employeeCode == "" {
		// 获取所有销售员的统计
		stats, err := model.GetAllSalesCommissionMonthlyStats(month)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取统计失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": stats,
			"message": "获取成功",
		})
		return
	}

	// 获取指定销售员的统计
	stats, err := model.GetSalesCommissionMonthlyStats(employeeCode, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取统计失败: " + err.Error()})
		return
	}

	if stats == nil {
		stats = &model.SalesCommissionMonthlyStats{
			EmployeeCode: employeeCode,
			StatMonth:    month,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": stats,
		"message": "获取成功",
	})
}

// AdminGetSalesCommissions 管理员获取销售员的分成记录列表
func AdminGetSalesCommissions(c *gin.Context) {
	// 验证管理员权限
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	employeeCode := c.Query("employee_code")
	month := c.Query("month")
	status := c.Query("status") // all, accounted, settled, unaccounted, unsettled
	startDateStr := c.Query("start_date") // YYYY-MM-DD格式
	endDateStr := c.Query("end_date") // YYYY-MM-DD格式
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)

	if employeeCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供销售员员工码"})
		return
	}

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}

	commissions, total, err := model.GetSalesCommissionsByEmployee(employeeCode, month, status, startDate, endDate, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取分成记录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  commissions,
			"total": total,
		},
		"message": "获取成功",
	})
}

// AdminGetSalesCommissionConfig 管理员获取销售员的分成配置
func AdminGetSalesCommissionConfig(c *gin.Context) {
	// 验证管理员权限
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	employeeCode := c.Query("employee_code")
	if employeeCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供销售员员工码"})
		return
	}

	config, err := model.GetSalesCommissionConfig(employeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": config,
		"message": "获取成功",
	})
}

// AdminUpdateSalesCommissionConfig 管理员更新销售员的分成配置
func AdminUpdateSalesCommissionConfig(c *gin.Context) {
	// 验证管理员权限
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	var req struct {
		EmployeeCode         string  `json:"employee_code" binding:"required"`
		BaseCommissionRate   float64 `json:"base_commission_rate"`
		NewCustomerBonusRate float64 `json:"new_customer_bonus_rate"`
		Tier1Threshold       float64 `json:"tier1_threshold"`
		Tier1Rate            float64 `json:"tier1_rate"`
		Tier2Threshold       float64 `json:"tier2_threshold"`
		Tier2Rate            float64 `json:"tier2_rate"`
		Tier3Threshold       float64 `json:"tier3_threshold"`
		Tier3Rate            float64 `json:"tier3_rate"`
		MinProfitThreshold   float64 `json:"min_profit_threshold"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 获取现有配置或创建默认配置
	config, err := model.GetSalesCommissionConfig(req.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配置失败: " + err.Error()})
		return
	}

	// 更新配置值（如果提供了新值）
	if req.BaseCommissionRate > 0 {
		config.BaseCommissionRate = req.BaseCommissionRate
	}
	if req.NewCustomerBonusRate > 0 {
		config.NewCustomerBonusRate = req.NewCustomerBonusRate
	}
	if req.Tier1Threshold > 0 {
		config.Tier1Threshold = req.Tier1Threshold
	}
	if req.Tier1Rate > 0 {
		config.Tier1Rate = req.Tier1Rate
	}
	if req.Tier2Threshold > 0 {
		config.Tier2Threshold = req.Tier2Threshold
	}
	if req.Tier2Rate > 0 {
		config.Tier2Rate = req.Tier2Rate
	}
	if req.Tier3Threshold > 0 {
		config.Tier3Threshold = req.Tier3Threshold
	}
	if req.Tier3Rate > 0 {
		config.Tier3Rate = req.Tier3Rate
	}
	if req.MinProfitThreshold > 0 {
		config.MinProfitThreshold = req.MinProfitThreshold
	}

	err = model.UpdateSalesCommissionConfig(req.EmployeeCode, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新配置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "更新成功",
	})
}

// AdminAccountSalesCommissions 管理员批量计入销售分成
func AdminAccountSalesCommissions(c *gin.Context) {
	// 验证管理员权限
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	var req struct {
		CommissionIDs []int    `json:"commission_ids"` // 分成记录ID列表（可选，如果提供则按ID计入）
		EmployeeCode string   `json:"employee_code"`  // 销售员员工码（可选）
		StartDate    string   `json:"start_date"`     // 开始日期 YYYY-MM-DD（可选）
		EndDate      string   `json:"end_date"`       // 结束日期 YYYY-MM-DD（可选）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "开始日期格式错误，应为YYYY-MM-DD"})
			return
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "结束日期格式错误，应为YYYY-MM-DD"})
			return
		}
	}

	// 如果既没有提供ID列表，也没有提供其他条件，则返回错误
	if len(req.CommissionIDs) == 0 && req.EmployeeCode == "" && startDate == nil && endDate == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供分成记录ID列表或其他筛选条件"})
		return
	}

	affected, err := model.AccountSalesCommissions(req.CommissionIDs, req.EmployeeCode, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计入失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "计入成功",
		"affected": affected,
	})
}

// AdminSettleSalesCommissions 管理员批量结算销售分成
func AdminSettleSalesCommissions(c *gin.Context) {
	// 验证管理员权限
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	var req struct {
		CommissionIDs []int    `json:"commission_ids"` // 分成记录ID列表（可选，如果提供则按ID结算）
		EmployeeCode  string   `json:"employee_code"`  // 销售员员工码（可选）
		StartDate     string   `json:"start_date"`     // 开始日期 YYYY-MM-DD（可选）
		EndDate       string   `json:"end_date"`       // 结束日期 YYYY-MM-DD（可选）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "开始日期格式错误，应为YYYY-MM-DD"})
			return
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "结束日期格式错误，应为YYYY-MM-DD"})
			return
		}
	}

	// 如果既没有提供ID列表，也没有提供其他条件，则返回错误
	if len(req.CommissionIDs) == 0 && req.EmployeeCode == "" && startDate == nil && endDate == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供分成记录ID列表或其他筛选条件"})
		return
	}

	affected, err := model.SettleSalesCommissions(req.CommissionIDs, req.EmployeeCode, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "结算失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "结算成功",
		"affected": affected,
	})
}

// AdminCancelAccountSalesCommissions 管理员取消计入销售分成
func AdminCancelAccountSalesCommissions(c *gin.Context) {
	// 验证管理员权限
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	var req struct {
		CommissionIDs []int `json:"commission_ids" binding:"required"` // 分成记录ID列表
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	if len(req.CommissionIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供要取消计入的记录ID列表"})
		return
	}

	affected, err := model.CancelAccountSalesCommissions(req.CommissionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "取消计入失败: " + err.Error()})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "没有符合条件的记录可以取消计入（只能取消已计入未结算的记录）"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "取消计入成功",
		"affected": affected,
	})
}

// AdminResetAccountSalesCommissions 管理员重新计入销售分成（重置分成）
func AdminResetAccountSalesCommissions(c *gin.Context) {
	// 验证管理员权限
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	var req struct {
		CommissionIDs []int `json:"commission_ids" binding:"required"` // 分成记录ID列表
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	if len(req.CommissionIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供要重新计入的记录ID列表"})
		return
	}

	affected, err := model.ResetAccountSalesCommissions(req.CommissionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "重新计入失败: " + err.Error()})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "没有符合条件的记录可以重新计入（只能重新计入已取消的记录）"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "重新计入成功",
		"affected": affected,
	})
}

