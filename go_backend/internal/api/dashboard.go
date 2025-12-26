package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"go_backend/internal/database"

	"github.com/gin-gonic/gin"
)

// GetDashboardStats 获取仪表盘统计数据
func GetDashboardStats(c *gin.Context) {
	// 获取时间范围参数（可选：today, week, month, custom）
	timeRange := c.DefaultQuery("time_range", "today")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// 计算时间范围
	var start, end time.Time
	now := time.Now()
	switch timeRange {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = now
	case "week":
		// 本周开始（周一）
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = now.AddDate(0, 0, -weekday+1)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		end = now
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = now
	case "custom":
		if startDate != "" && endDate != "" {
			var err error
			start, err = time.Parse("2006-01-02", startDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "开始日期格式错误"})
				return
			}
			end, err = time.Parse("2006-01-02", endDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "结束日期格式错误"})
				return
			}
			end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
		} else {
			start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			end = now
		}
	default:
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = now
	}

	// 获取核心业务指标
	orderStats := getOrderStats(start, end)
	revenueStats := getRevenueStats(start, end)
	userStats := getUserStats(start, end)
	productStats := getProductStats(start, end)
	deliveryStats := getDeliveryStats(start, end)
	salesStats := getSalesStats(start, end)

	// 获取趋势数据（对比上一个周期）
	var prevStart, prevEnd time.Time
	periodDays := int(end.Sub(start).Hours() / 24)
	if periodDays > 0 {
		prevEnd = start.Add(-time.Second)
		prevStart = prevEnd.AddDate(0, 0, -periodDays+1)
	} else {
		prevStart = start.AddDate(0, 0, -1)
		prevEnd = start.Add(-time.Second)
	}

	prevOrderStats := getOrderStats(prevStart, prevEnd)
	prevRevenueStats := getRevenueStats(prevStart, prevEnd)

	// 计算环比
	orderGrowth := calculateGrowth(orderStats["total_orders"].(int), prevOrderStats["total_orders"].(int))
	revenueGrowth := calculateGrowth(revenueStats["total_revenue"].(float64), prevRevenueStats["total_revenue"].(float64))

	// 获取订单趋势数据（按日）
	orderTrend := getOrderTrend(start, end)
	revenueTrend := getRevenueTrend(start, end)

	// 获取订单状态分布
	orderStatusDistribution := getOrderStatusDistribution()

	// 获取热销商品
	hotProducts := getHotProducts(start, end, 10)

	// 获取配送员绩效排名
	deliveryRanking := getDeliveryRanking(start, end, 10)

	// 获取销售员绩效排名
	salesRanking := getSalesRanking(start, end, 10)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"time_range": gin.H{
				"start": start.Format("2006-01-02 15:04:05"),
				"end":   end.Format("2006-01-02 15:04:05"),
			},
			"order_stats": gin.H{
				"total_orders":     orderStats["total_orders"],
				"pending_delivery": orderStats["pending_delivery"],
				"delivering":       orderStats["delivering"],
				"delivered":        orderStats["delivered"],
				"paid":             orderStats["paid"],
				"cancelled":        orderStats["cancelled"],
				"total_amount":     orderStats["total_amount"],
				"total_profit":     orderStats["total_profit"],
				"net_profit":       orderStats["net_profit"],
				"growth":           orderGrowth,
			},
			"revenue_stats": gin.H{
				"total_revenue":        revenueStats["total_revenue"],
				"goods_revenue":        revenueStats["goods_revenue"],
				"delivery_fee_revenue": revenueStats["delivery_fee_revenue"],
				"urgent_fee_revenue":   revenueStats["urgent_fee_revenue"],
				"total_cost":           revenueStats["total_cost"],
				"goods_cost":           revenueStats["goods_cost"],
				"delivery_cost":        revenueStats["delivery_cost"],
				"sales_commission":     revenueStats["sales_commission"],
				"net_profit":           revenueStats["net_profit"],
				"growth":               revenueGrowth,
			},
			"user_stats":                userStats,
			"product_stats":             productStats,
			"delivery_stats":            deliveryStats,
			"sales_stats":               salesStats,
			"order_trend":               orderTrend,
			"revenue_trend":             revenueTrend,
			"order_status_distribution": orderStatusDistribution,
			"hot_products":              hotProducts,
			"delivery_ranking":          deliveryRanking,
			"sales_ranking":             salesRanking,
		},
		"message": "获取成功",
	})
}

// getOrderStats 获取订单统计
func getOrderStats(start, end time.Time) map[string]interface{} {
	query := `
		SELECT 
			COUNT(*) as total_orders,
			COALESCE(SUM(CASE WHEN status = 'pending_delivery' THEN 1 ELSE 0 END), 0) as pending_delivery,
			COALESCE(SUM(CASE WHEN status = 'pending_pickup' THEN 1 ELSE 0 END), 0) as pending_pickup,
			COALESCE(SUM(CASE WHEN status = 'delivering' THEN 1 ELSE 0 END), 0) as delivering,
			COALESCE(SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END), 0) as delivered,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END), 0) as paid,
			COALESCE(SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END), 0) as cancelled,
			COALESCE(SUM(total_amount), 0) as total_amount,
			COALESCE(SUM(order_profit), 0) as total_profit,
			COALESCE(SUM(net_profit), 0) as net_profit
		FROM orders
		WHERE created_at >= ? AND created_at <= ?
	`

	var stats map[string]interface{}
	var totalOrders, pendingDelivery, pendingPickup, delivering, delivered, paid, cancelled int
	var totalAmount, totalProfit, netProfit sql.NullFloat64

	err := database.DB.QueryRow(query, start, end).Scan(
		&totalOrders, &pendingDelivery, &pendingPickup, &delivering, &delivered, &paid, &cancelled,
		&totalAmount, &totalProfit, &netProfit,
	)

	if err != nil {
		log.Printf("获取订单统计失败: %v", err)
		return map[string]interface{}{
			"total_orders": 0, "pending_delivery": 0, "delivering": 0, "delivered": 0, "paid": 0, "cancelled": 0,
			"total_amount": 0.0, "total_profit": 0.0, "net_profit": 0.0,
		}
	}

	stats = map[string]interface{}{
		"total_orders":     totalOrders,
		"pending_delivery": pendingDelivery + pendingPickup, // 合并待配送和待取货
		"delivering":       delivering,
		"delivered":        delivered,
		"paid":             paid,
		"cancelled":        cancelled,
		"total_amount":     0.0,
		"total_profit":     0.0,
		"net_profit":       0.0,
	}

	if totalAmount.Valid {
		stats["total_amount"] = totalAmount.Float64
	}
	if totalProfit.Valid {
		stats["total_profit"] = totalProfit.Float64
	}
	if netProfit.Valid {
		stats["net_profit"] = netProfit.Float64
	}

	return stats
}

// getRevenueStats 获取收入统计
func getRevenueStats(start, end time.Time) map[string]interface{} {
	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(goods_amount), 0) as goods_revenue,
			COALESCE(SUM(delivery_fee), 0) as delivery_fee_revenue,
			COALESCE(SUM(urgent_fee), 0) as urgent_fee_revenue,
			COALESCE(SUM(goods_amount - COALESCE(order_profit, 0)), 0) as goods_cost,
			COALESCE(SUM(CAST(JSON_EXTRACT(delivery_fee_calculation, '$.total_platform_cost') AS DECIMAL(10,2))), 0) as delivery_cost
		FROM orders
		WHERE created_at >= ? AND created_at <= ?
	`

	var totalRevenue, goodsRevenue, deliveryFeeRevenue, urgentFeeRevenue, goodsCost, deliveryCost sql.NullFloat64

	err := database.DB.QueryRow(query, start, end).Scan(
		&totalRevenue, &goodsRevenue, &deliveryFeeRevenue, &urgentFeeRevenue, &goodsCost, &deliveryCost,
	)

	if err != nil {
		log.Printf("获取收入统计失败: %v", err)
		return map[string]interface{}{
			"total_revenue": 0.0, "goods_revenue": 0.0, "delivery_fee_revenue": 0.0, "urgent_fee_revenue": 0.0,
			"total_cost": 0.0, "goods_cost": 0.0, "delivery_cost": 0.0, "sales_commission": 0.0, "net_profit": 0.0,
		}
	}

	// 获取销售分成支出（已结算订单）
	salesCommissionQuery := `
		SELECT COALESCE(SUM(total_commission), 0)
		FROM sales_commissions
		WHERE settlement_date >= ? AND settlement_date <= ?
	`
	var salesCommission sql.NullFloat64
	_ = database.DB.QueryRow(salesCommissionQuery, start, end).Scan(&salesCommission)

	totalCost := 0.0
	if goodsCost.Valid {
		totalCost += goodsCost.Float64
	}
	if deliveryCost.Valid {
		totalCost += deliveryCost.Float64
	}
	if salesCommission.Valid {
		totalCost += salesCommission.Float64
	}

	netProfit := 0.0
	if totalRevenue.Valid {
		netProfit = totalRevenue.Float64 - totalCost
	}

	return map[string]interface{}{
		"total_revenue":        getFloatValue(totalRevenue),
		"goods_revenue":        getFloatValue(goodsRevenue),
		"delivery_fee_revenue": getFloatValue(deliveryFeeRevenue),
		"urgent_fee_revenue":   getFloatValue(urgentFeeRevenue),
		"total_cost":           totalCost,
		"goods_cost":           getFloatValue(goodsCost),
		"delivery_cost":        getFloatValue(deliveryCost),
		"sales_commission":     getFloatValue(salesCommission),
		"net_profit":           netProfit,
	}
}

// getUserStats 获取用户统计
func getUserStats(start, end time.Time) map[string]interface{} {
	// 总用户数
	var totalUsers int
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM mini_app_users").Scan(&totalUsers)

	// 新增用户数
	var newUsers int
	query := "SELECT COUNT(*) FROM mini_app_users WHERE created_at >= ? AND created_at <= ?"
	_ = database.DB.QueryRow(query, start, end).Scan(&newUsers)

	// 活跃用户数（有订单的用户）
	var activeUsers int
	activeQuery := `
		SELECT COUNT(DISTINCT user_id)
		FROM orders
		WHERE created_at >= ? AND created_at <= ?
	`
	_ = database.DB.QueryRow(activeQuery, start, end).Scan(&activeUsers)

	// 用户类型分布
	var retailUsers, wholesaleUsers, unknownUsers int
	typeQuery := `
		SELECT 
			SUM(CASE WHEN user_type = 'retail' THEN 1 ELSE 0 END) as retail,
			SUM(CASE WHEN user_type = 'wholesale' THEN 1 ELSE 0 END) as wholesale,
			SUM(CASE WHEN user_type = 'unknown' OR user_type = '' THEN 1 ELSE 0 END) as unknown
		FROM mini_app_users
	`
	_ = database.DB.QueryRow(typeQuery).Scan(&retailUsers, &wholesaleUsers, &unknownUsers)

	// 有销售员绑定的用户数
	var usersWithSales int
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM mini_app_users WHERE sales_code IS NOT NULL AND sales_code != ''").Scan(&usersWithSales)

	return map[string]interface{}{
		"total_users":         totalUsers,
		"new_users":           newUsers,
		"active_users":        activeUsers,
		"retail_users":        retailUsers,
		"wholesale_users":     wholesaleUsers,
		"unknown_users":       unknownUsers,
		"users_with_sales":    usersWithSales,
		"users_without_sales": totalUsers - usersWithSales,
	}
}

// getProductStats 获取商品统计
func getProductStats(_ time.Time, _ time.Time) map[string]interface{} {
	// 商品总数
	var totalProducts int
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)

	// 分类数
	var totalCategories int
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&totalCategories)

	return map[string]interface{}{
		"total_products":   totalProducts,
		"total_categories": totalCategories,
	}
}

// getDeliveryStats 获取配送统计
func getDeliveryStats(start, end time.Time) map[string]interface{} {
	// 配送订单数
	var deliveryOrders int
	query := `
		SELECT COUNT(*)
		FROM orders
		WHERE delivery_employee_code IS NOT NULL
			AND delivery_employee_code != ''
			AND created_at >= ? AND created_at <= ?
	`
	_ = database.DB.QueryRow(query, start, end).Scan(&deliveryOrders)

	// 已完成配送订单数
	var completedDelivery int
	completedQuery := `
		SELECT COUNT(*)
		FROM orders
		WHERE delivery_employee_code IS NOT NULL
			AND delivery_employee_code != ''
			AND status IN ('delivered', 'paid')
			AND created_at >= ? AND created_at <= ?
	`
	_ = database.DB.QueryRow(completedQuery, start, end).Scan(&completedDelivery)

	// 活跃配送员数
	var activeDeliveryEmployees int
	activeQuery := `
		SELECT COUNT(DISTINCT delivery_employee_code)
		FROM orders
		WHERE delivery_employee_code IS NOT NULL
			AND delivery_employee_code != ''
			AND created_at >= ? AND created_at <= ?
	`
	_ = database.DB.QueryRow(activeQuery, start, end).Scan(&activeDeliveryEmployees)

	// 孤立订单数
	var isolatedOrders int
	isolatedQuery := `
		SELECT COUNT(*)
		FROM orders
		WHERE is_isolated = 1
			AND created_at >= ? AND created_at <= ?
	`
	_ = database.DB.QueryRow(isolatedQuery, start, end).Scan(&isolatedOrders)

	// 加急订单数
	var urgentOrders int
	urgentQuery := `
		SELECT COUNT(*)
		FROM orders
		WHERE is_urgent = 1
			AND created_at >= ? AND created_at <= ?
	`
	_ = database.DB.QueryRow(urgentQuery, start, end).Scan(&urgentOrders)

	completionRate := 0.0
	if deliveryOrders > 0 {
		completionRate = float64(completedDelivery) / float64(deliveryOrders) * 100
	}

	return map[string]interface{}{
		"delivery_orders":           deliveryOrders,
		"completed_delivery":        completedDelivery,
		"completion_rate":           completionRate,
		"active_delivery_employees": activeDeliveryEmployees,
		"isolated_orders":           isolatedOrders,
		"urgent_orders":             urgentOrders,
	}
}

// getSalesStats 获取销售统计
func getSalesStats(start, end time.Time) map[string]interface{} {
	// 活跃销售员数
	var activeSalesEmployees int
	activeQuery := `
		SELECT COUNT(DISTINCT employee_code)
		FROM sales_commissions
		WHERE calculation_month >= ? AND calculation_month <= ?
	`
	startMonth := start.Format("2006-01")
	endMonth := end.Format("2006-01")
	_ = database.DB.QueryRow(activeQuery, startMonth, endMonth).Scan(&activeSalesEmployees)

	// 销售分成总额
	var totalCommission sql.NullFloat64
	commissionQuery := `
		SELECT COALESCE(SUM(total_commission), 0)
		FROM sales_commissions
		WHERE settlement_date >= ? AND settlement_date <= ?
	`
	_ = database.DB.QueryRow(commissionQuery, start, end).Scan(&totalCommission)

	// 新客数
	var newCustomers int
	newCustomerQuery := `
		SELECT COUNT(DISTINCT user_id)
		FROM sales_commissions
		WHERE is_new_customer_order = 1
			AND settlement_date >= ? AND settlement_date <= ?
	`
	_ = database.DB.QueryRow(newCustomerQuery, start, end).Scan(&newCustomers)

	return map[string]interface{}{
		"active_sales_employees": activeSalesEmployees,
		"total_commission":       getFloatValue(totalCommission),
		"new_customers":          newCustomers,
	}
}

// getOrderTrend 获取订单趋势（按日）
func getOrderTrend(start, end time.Time) []map[string]interface{} {
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as order_count,
			COALESCE(SUM(total_amount), 0) as total_amount
		FROM orders
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`

	rows, err := database.DB.Query(query, start, end)
	if err != nil {
		log.Printf("获取订单趋势失败: %v", err)
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var trend []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var orderCount int
		var totalAmount sql.NullFloat64

		if err := rows.Scan(&date, &orderCount, &totalAmount); err != nil {
			continue
		}

		trend = append(trend, map[string]interface{}{
			"date":         date.Format("2006-01-02"),
			"order_count":  orderCount,
			"total_amount": getFloatValue(totalAmount),
		})
	}

	return trend
}

// getRevenueTrend 获取收入趋势（按日）
func getRevenueTrend(start, end time.Time) []map[string]interface{} {
	query := `
		SELECT 
			DATE(created_at) as date,
			COALESCE(SUM(total_amount), 0) as revenue,
			COALESCE(SUM(order_profit), 0) as profit,
			COALESCE(SUM(net_profit), 0) as net_profit
		FROM orders
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`

	rows, err := database.DB.Query(query, start, end)
	if err != nil {
		log.Printf("获取收入趋势失败: %v", err)
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var trend []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var revenue, profit, netProfit sql.NullFloat64

		if err := rows.Scan(&date, &revenue, &profit, &netProfit); err != nil {
			continue
		}

		trend = append(trend, map[string]interface{}{
			"date":       date.Format("2006-01-02"),
			"revenue":    getFloatValue(revenue),
			"profit":     getFloatValue(profit),
			"net_profit": getFloatValue(netProfit),
		})
	}

	return trend
}

// getOrderStatusDistribution 获取订单状态分布
func getOrderStatusDistribution() []map[string]interface{} {
	query := `
		SELECT 
			status,
			COUNT(*) as count
		FROM orders
		GROUP BY status
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("获取订单状态分布失败: %v", err)
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var distribution []map[string]interface{}
	for rows.Next() {
		var status string
		var count int

		if err := rows.Scan(&status, &count); err != nil {
			continue
		}

		distribution = append(distribution, map[string]interface{}{
			"status": status,
			"count":  count,
		})
	}

	return distribution
}

// getHotProducts 获取热销商品
func getHotProducts(start, end time.Time, limit int) []map[string]interface{} {
	query := `
		SELECT 
			oi.product_id,
			oi.product_name,
			oi.image,
			SUM(oi.quantity) as total_quantity,
			SUM(oi.subtotal) as total_amount
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.created_at >= ? AND o.created_at <= ?
		GROUP BY oi.product_id, oi.product_name, oi.image
		ORDER BY total_amount DESC
		LIMIT ?
	`

	rows, err := database.DB.Query(query, start, end, limit)
	if err != nil {
		log.Printf("获取热销商品失败: %v", err)
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var productID int
		var productName, image sql.NullString
		var totalQuantity int
		var totalAmount sql.NullFloat64

		if err := rows.Scan(&productID, &productName, &image, &totalQuantity, &totalAmount); err != nil {
			continue
		}

		products = append(products, map[string]interface{}{
			"product_id":     productID,
			"product_name":   getStringValue(productName),
			"image":          getStringValue(image),
			"total_quantity": totalQuantity,
			"total_amount":   getFloatValue(totalAmount),
		})
	}

	return products
}

// getDeliveryRanking 获取配送员绩效排名
func getDeliveryRanking(start, end time.Time, limit int) []map[string]interface{} {
	query := `
		SELECT 
			o.delivery_employee_code,
			e.name as employee_name,
			COUNT(*) as order_count,
			COALESCE(SUM(CAST(JSON_EXTRACT(o.delivery_fee_calculation, '$.rider_payable_fee') AS DECIMAL(10,2))), 0) as total_fee
		FROM orders o
		LEFT JOIN employees e ON o.delivery_employee_code = e.employee_code
		WHERE o.delivery_employee_code IS NOT NULL
			AND o.delivery_employee_code != ''
			AND o.created_at >= ? AND o.created_at <= ?
			AND o.status IN ('delivered', 'paid')
		GROUP BY o.delivery_employee_code, e.name
		ORDER BY order_count DESC, total_fee DESC
		LIMIT ?
	`

	rows, err := database.DB.Query(query, start, end, limit)
	if err != nil {
		log.Printf("获取配送员排名失败: %v", err)
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var ranking []map[string]interface{}
	for rows.Next() {
		var employeeCode sql.NullString
		var employeeName sql.NullString
		var orderCount int
		var totalFee sql.NullFloat64

		if err := rows.Scan(&employeeCode, &employeeName, &orderCount, &totalFee); err != nil {
			continue
		}

		ranking = append(ranking, map[string]interface{}{
			"employee_code": getStringValue(employeeCode),
			"employee_name": getStringValue(employeeName),
			"order_count":   orderCount,
			"total_fee":     getFloatValue(totalFee),
		})
	}

	return ranking
}

// getSalesRanking 获取销售员绩效排名
func getSalesRanking(start, end time.Time, limit int) []map[string]interface{} {
	query := `
		SELECT 
			sc.employee_code,
			e.name as employee_name,
			COUNT(DISTINCT sc.order_id) as order_count,
			COALESCE(SUM(sc.order_amount), 0) as total_sales,
			COALESCE(SUM(sc.total_commission), 0) as total_commission,
			SUM(CASE WHEN sc.is_new_customer_order = 1 THEN 1 ELSE 0 END) as new_customer_count
		FROM sales_commissions sc
		LEFT JOIN employees e ON sc.employee_code = e.employee_code
		WHERE sc.settlement_date >= ? AND sc.settlement_date <= ?
		GROUP BY sc.employee_code, e.name
		ORDER BY total_sales DESC
		LIMIT ?
	`

	rows, err := database.DB.Query(query, start, end, limit)
	if err != nil {
		log.Printf("获取销售员排名失败: %v", err)
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var ranking []map[string]interface{}
	for rows.Next() {
		var employeeCode string
		var employeeName sql.NullString
		var orderCount int
		var totalSales, totalCommission sql.NullFloat64
		var newCustomerCount int

		if err := rows.Scan(&employeeCode, &employeeName, &orderCount, &totalSales, &totalCommission, &newCustomerCount); err != nil {
			continue
		}

		ranking = append(ranking, map[string]interface{}{
			"employee_code":      employeeCode,
			"employee_name":      getStringValue(employeeName),
			"order_count":        orderCount,
			"total_sales":        getFloatValue(totalSales),
			"total_commission":   getFloatValue(totalCommission),
			"new_customer_count": newCustomerCount,
		})
	}

	return ranking
}

// 辅助函数
func calculateGrowth(current, previous interface{}) float64 {
	var curr, prev float64

	switch v := current.(type) {
	case int:
		curr = float64(v)
	case float64:
		curr = v
	default:
		return 0
	}

	switch v := previous.(type) {
	case int:
		prev = float64(v)
	case float64:
		prev = v
	default:
		return 0
	}

	if prev == 0 {
		if curr > 0 {
			return 100.0
		}
		return 0
	}

	return ((curr - prev) / prev) * 100
}

func getFloatValue(v sql.NullFloat64) float64 {
	if v.Valid {
		return v.Float64
	}
	return 0.0
}

func getStringValue(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}
