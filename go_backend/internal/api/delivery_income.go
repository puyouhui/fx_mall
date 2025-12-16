package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// DeliveryIncomeStats 配送员收入统计
type DeliveryIncomeStats struct {
	SettledFee   float64 `json:"settled_fee"`   // 已结算配送费
	UnsettledFee float64 `json:"unsettled_fee"` // 未结算配送费
	TotalFee     float64 `json:"total_fee"`     // 总配送费
	OrderCount   int     `json:"order_count"`   // 已完成订单数量
}

// GetDeliveryIncomeStats 获取配送员收入统计（配送员端）
func GetDeliveryIncomeStats(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	// 验证是否是配送员
	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	// 查询已完成订单（delivered或paid状态）的配送费统计
	// 只统计该配送员配送的订单
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN o.delivery_fee_settled = 1 THEN CAST(JSON_EXTRACT(o.delivery_fee_calculation, '$.rider_payable_fee') AS DECIMAL(10,2)) ELSE 0 END), 0) as settled_fee,
			COALESCE(SUM(CASE WHEN o.delivery_fee_settled = 0 THEN CAST(JSON_EXTRACT(o.delivery_fee_calculation, '$.rider_payable_fee') AS DECIMAL(10,2)) ELSE 0 END), 0) as unsettled_fee,
			COALESCE(SUM(CAST(JSON_EXTRACT(o.delivery_fee_calculation, '$.rider_payable_fee') AS DECIMAL(10,2))), 0) as total_fee,
			COUNT(*) as order_count
		FROM orders o
		WHERE o.delivery_employee_code = ?
			AND o.status IN ('delivered', 'paid')
			AND o.delivery_fee_calculation IS NOT NULL
	`

	var stats DeliveryIncomeStats
	var settledFee, unsettledFee, totalFee sql.NullFloat64
	var orderCount int

	err := database.DB.QueryRow(query, employee.EmployeeCode).Scan(
		&settledFee,
		&unsettledFee,
		&totalFee,
		&orderCount,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取收入统计失败: " + err.Error(),
		})
		return
	}

	stats.SettledFee = 0
	stats.UnsettledFee = 0
	stats.TotalFee = 0

	if settledFee.Valid {
		stats.SettledFee = settledFee.Float64
	}
	if unsettledFee.Valid {
		stats.UnsettledFee = unsettledFee.Float64
	}
	if totalFee.Valid {
		stats.TotalFee = totalFee.Float64
	}
	stats.OrderCount = orderCount

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// GetDeliveryIncomeDetails 获取配送员收入明细（配送员端）
func GetDeliveryIncomeDetails(c *gin.Context) {
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
	pageSize := parseQueryInt(c, "pageSize", 20)
	settled := c.Query("settled") // 可选：true/false，筛选已结算/未结算

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// 构建查询条件
	where := "o.delivery_employee_code = ? AND o.status IN ('delivered', 'paid') AND o.delivery_fee_calculation IS NOT NULL"
	args := []interface{}{employee.EmployeeCode}

	if settled == "true" {
		where += " AND o.delivery_fee_settled = 1"
	} else if settled == "false" {
		where += " AND o.delivery_fee_settled = 0"
	}

	// 获取总数量
	var total int
	countQuery := "SELECT COUNT(*) FROM orders o WHERE " + where
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
		SELECT 
			o.id,
			o.order_number,
			o.status,
			o.delivery_fee_settled,
			o.settlement_date,
			CAST(JSON_EXTRACT(o.delivery_fee_calculation, '$.rider_payable_fee') AS DECIMAL(10,2)) as rider_payable_fee,
			o.created_at,
			o.updated_at,
			a.name as address_name
		FROM orders o
		LEFT JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE ` + where + `
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	orders := make([]map[string]interface{}, 0)
	for rows.Next() {
		var (
			id                 int
			orderNumber        string
			status             string
			deliveryFeeSettled bool
			settlementDate     sql.NullTime
			riderPayableFee    sql.NullFloat64
			createdAt          time.Time
			updatedAt          time.Time
			addressName        sql.NullString
		)

		if err := rows.Scan(
			&id,
			&orderNumber,
			&status,
			&deliveryFeeSettled,
			&settlementDate,
			&riderPayableFee,
			&createdAt,
			&updatedAt,
			&addressName,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "扫描订单数据失败: " + err.Error()})
			return
		}

		order := map[string]interface{}{
			"id":                   id,
			"order_number":         orderNumber,
			"status":               status,
			"delivery_fee_settled": deliveryFeeSettled,
			"rider_payable_fee":    0.0,
			"created_at":           createdAt,
			"updated_at":           updatedAt,
		}

		if riderPayableFee.Valid {
			order["rider_payable_fee"] = riderPayableFee.Float64
		}

		if settlementDate.Valid {
			order["settlement_date"] = settlementDate.Time.Format("2006-01-02 15:04:05")
		}

		if addressName.Valid {
			order["address_name"] = addressName.String
		} else {
			order["address_name"] = ""
		}

		orders = append(orders, order)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":     orders,
			"total":    total,
			"pageNum":  pageNum,
			"pageSize": pageSize,
		},
	})
}

// BatchSettleDeliveryFees 批量结算配送费（管理员端）
func BatchSettleDeliveryFees(c *gin.Context) {
	// 从上下文中获取管理员ID（需要AuthMiddleware配合）
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	// 验证管理员ID
	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	var req struct {
		EmployeeCode   string `json:"employee_code" binding:"required"`   // 配送员员工码
		OrderIDs       []int  `json:"order_ids,omitempty"`                // 可选：指定订单ID列表，为空则结算所有未结算订单
		SettlementDate string `json:"settlement_date" binding:"required"` // 结算日期（格式：YYYY-MM-DD）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 解析结算日期
	settlementDate, err := time.Parse("2006-01-02", req.SettlementDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "结算日期格式错误，应为 YYYY-MM-DD"})
		return
	}

	// 构建更新条件
	where := "delivery_employee_code = ? AND status IN ('delivered', 'paid') AND delivery_fee_settled = 0 AND delivery_fee_calculation IS NOT NULL"
	args := []interface{}{req.EmployeeCode}

	if len(req.OrderIDs) > 0 {
		// 如果指定了订单ID列表，只结算这些订单
		placeholders := ""
		for i, id := range req.OrderIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, id)
		}
		where += " AND id IN (" + placeholders + ")"
	}

	// 执行批量结算
	updateQuery := `
		UPDATE orders 
		SET delivery_fee_settled = 1, 
		    settlement_date = ?,
		    updated_at = NOW()
		WHERE ` + where
	args = append([]interface{}{settlementDate}, args...)

	result, err := database.DB.Exec(updateQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "批量结算失败: " + err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取更新行数失败: " + err.Error()})
		return
	}

	// 记录操作日志（可选）
	_ = ok // 管理员已验证

	// 对于已结算的订单，如果状态是paid，计算销售分成
	// 查询已结算且状态为paid的订单ID（使用更新后的条件）
	settledOrderQuery := "delivery_employee_code = ? AND status = 'paid' AND delivery_fee_settled = 1 AND settlement_date IS NOT NULL"
	settledOrderArgs := []interface{}{req.EmployeeCode}
	if len(req.OrderIDs) > 0 {
		placeholders := ""
		for i, id := range req.OrderIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			settledOrderArgs = append(settledOrderArgs, id)
		}
		settledOrderQuery += " AND id IN (" + placeholders + ")"
	}
	
	settledOrderRows, err := database.DB.Query("SELECT id FROM orders WHERE "+settledOrderQuery, settledOrderArgs...)
	if err == nil {
		defer settledOrderRows.Close()
		for settledOrderRows.Next() {
			var orderID int
			if err := settledOrderRows.Scan(&orderID); err == nil {
				// 异步处理销售分成计算（避免阻塞）
				go func(id int) {
					if err := model.ProcessOrderSettlement(id); err != nil {
						log.Printf("处理订单 %d 的销售分成失败: %v", id, err)
					}
				}(orderID)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "批量结算成功",
		"data": gin.H{
			"settled_count":   rowsAffected,
			"settlement_date": req.SettlementDate,
		},
	})
}

// GetDeliveryIncomeStatsForAdmin 获取配送员收入统计（管理员端，可查看所有配送员）
func GetDeliveryIncomeStatsForAdmin(c *gin.Context) {
	// 从上下文中获取管理员ID（需要AuthMiddleware配合）
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	// 验证管理员ID
	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	_ = ok // 管理员已验证

	employeeCode := c.Query("employee_code") // 可选：指定配送员员工码

	// 构建查询条件
	where := "o.status IN ('delivered', 'paid') AND o.delivery_fee_calculation IS NOT NULL"
	args := []interface{}{}

	if employeeCode != "" {
		where += " AND o.delivery_employee_code = ?"
		args = append(args, employeeCode)
	}

	// 查询统计信息（关联员工表获取姓名）
	query := `
		SELECT 
			o.delivery_employee_code,
			e.name as employee_name,
			COALESCE(SUM(CASE WHEN o.delivery_fee_settled = 1 THEN CAST(JSON_EXTRACT(o.delivery_fee_calculation, '$.rider_payable_fee') AS DECIMAL(10,2)) ELSE 0 END), 0) as settled_fee,
			COALESCE(SUM(CASE WHEN o.delivery_fee_settled = 0 THEN CAST(JSON_EXTRACT(o.delivery_fee_calculation, '$.rider_payable_fee') AS DECIMAL(10,2)) ELSE 0 END), 0) as unsettled_fee,
			COALESCE(SUM(CAST(JSON_EXTRACT(o.delivery_fee_calculation, '$.rider_payable_fee') AS DECIMAL(10,2))), 0) as total_fee,
			COUNT(*) as order_count
		FROM orders o
		LEFT JOIN employees e ON o.delivery_employee_code = e.employee_code
		WHERE ` + where + `
		GROUP BY o.delivery_employee_code, e.name
		ORDER BY o.delivery_employee_code
	`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取收入统计失败: " + err.Error()})
		return
	}
	defer rows.Close()

	statsList := make([]map[string]interface{}, 0)
	for rows.Next() {
		var (
			empCode      sql.NullString
			empName      sql.NullString
			settledFee   sql.NullFloat64
			unsettledFee sql.NullFloat64
			totalFee     sql.NullFloat64
			orderCount   int
		)

		if err := rows.Scan(&empCode, &empName, &settledFee, &unsettledFee, &totalFee, &orderCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "扫描统计数据失败: " + err.Error()})
			return
		}

		// 计算平均每单配送成本
		avgFeePerOrder := 0.0
		if orderCount > 0 && totalFee.Valid {
			avgFeePerOrder = totalFee.Float64 / float64(orderCount)
		}

		stat := map[string]interface{}{
			"employee_code":     "",
			"employee_name":     "",
			"settled_fee":       0.0,
			"unsettled_fee":     0.0,
			"total_fee":         0.0,
			"order_count":       orderCount,
			"avg_fee_per_order": avgFeePerOrder,
		}

		if empCode.Valid {
			stat["employee_code"] = empCode.String
		}
		if empName.Valid {
			stat["employee_name"] = empName.String
		}
		if settledFee.Valid {
			stat["settled_fee"] = settledFee.Float64
		}
		if unsettledFee.Valid {
			stat["unsettled_fee"] = unsettledFee.Float64
		}
		if totalFee.Valid {
			stat["total_fee"] = totalFee.Float64
		}

		statsList = append(statsList, stat)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    statsList,
	})
}
