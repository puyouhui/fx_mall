package api

import (
	"net/http"
	"time"

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
		OrderAmount  float64 `json:"order_amount" binding:"required"`  // 平台总收入
		GoodsCost    float64 `json:"goods_cost" binding:"required"`  // 商品总成本
		DeliveryCost float64 `json:"delivery_cost" binding:"required"` // 配送成本
		UserID       int     `json:"user_id" binding:"required"`     // 客户用户ID
		OrderID      int     `json:"order_id"`                        // 订单ID（可选，用于判断是否新客户）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
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

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	commissions, total, err := model.GetSalesCommissionsByEmployee(employee.EmployeeCode, month, pageNum, pageSize)
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

	commissions, total, err := model.GetSalesCommissionsByEmployee(employeeCode, month, pageNum, pageSize)
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

