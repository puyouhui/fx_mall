package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetSupplierPaidItems 供应商获取已付款清单
func GetSupplierPaidItems(c *gin.Context) {
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 获取查询参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	orderNumber := c.Query("order_number")

	// 构建查询条件
	whereClause := "WHERE sp.supplier_id = ? AND sp.status = 1"
	args := []interface{}{supplierID}

	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			whereClause += " AND sp.payment_date >= ?"
			args = append(args, t)
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			whereClause += " AND sp.payment_date <= ?"
			args = append(args, t)
		}
	}

	// 查询付款记录
	query := `
		SELECT 
			sp.id,
			sp.payment_date,
			sp.payment_amount,
			sp.payment_method,
			sp.payment_account,
			sp.payment_receipt,
			sp.remark,
			sp.created_at
		FROM supplier_payments sp
		` + whereClause + `
		ORDER BY sp.payment_date DESC, sp.id DESC
	`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}
	defer rows.Close()

	payments := []map[string]interface{}{}
	for rows.Next() {
		var paymentID int
		var paymentDate, createdAt time.Time
		var paymentAmount float64
		var paymentMethod, paymentAccount, paymentReceipt, remark sql.NullString

		if err := rows.Scan(&paymentID, &paymentDate, &paymentAmount, &paymentMethod, &paymentAccount, &paymentReceipt, &remark, &createdAt); err != nil {
			continue
		}

		// 获取付款明细
		items, err := model.GetSupplierPaymentItems(paymentID)
		if err != nil {
			continue
		}

		// 如果指定了订单号，过滤订单
		if orderNumber != "" {
			filteredItems := []model.SupplierPaymentItem{}
			for _, item := range items {
				var orderNum string
				database.DB.QueryRow("SELECT order_number FROM orders WHERE id = ?", item.OrderID).Scan(&orderNum)
				if orderNum == orderNumber {
					filteredItems = append(filteredItems, item)
				}
			}
			if len(filteredItems) == 0 {
				continue
			}
			items = filteredItems
		}

		paymentData := map[string]interface{}{
			"payment_id":      paymentID,
			"payment_date":    paymentDate.Format("2006-01-02"),
			"payment_amount":  paymentAmount,
			"payment_method":  nil,
			"payment_account": nil,
			"payment_receipt": nil,
			"remark":          nil,
			"created_at":      createdAt.Format("2006-01-02 15:04:05"),
			"items":           items,
		}

		if paymentMethod.Valid {
			paymentData["payment_method"] = paymentMethod.String
		}
		if paymentAccount.Valid {
			paymentData["payment_account"] = paymentAccount.String
		}
		if paymentReceipt.Valid {
			paymentData["payment_receipt"] = paymentReceipt.String
		}
		if remark.Valid {
			paymentData["remark"] = remark.String
		}

		payments = append(payments, paymentData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    payments,
		"message": "success",
	})
}

// GetSupplierPendingItems 供应商获取待付款清单
func GetSupplierPendingItems(c *gin.Context) {
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 获取查询参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// 获取已付款的订单商品ID列表
	paidItems, err := model.GetPaidOrderItemIDs(&supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}

	// 构建时间范围
	var startDate, endDate time.Time
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
		}
	}

	// 查询订单列表（包含已取货商品的订单）
	orderQuery := `
		SELECT DISTINCT
			o.id,
			o.order_number,
			o.status,
			o.created_at,
			COALESCE(
				(SELECT action_time 
				 FROM delivery_logs 
				 WHERE order_id = o.id 
				   AND action = 'pickup_completed' 
				 ORDER BY action_time DESC 
				 LIMIT 1),
				o.updated_at
			) as pickup_time
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
			AND oi.is_picked = 1
			AND o.status != 'cancelled'
	`

	orderArgs := []interface{}{supplierID}
	if !startDate.IsZero() {
		orderQuery += " AND o.created_at >= ?"
		orderArgs = append(orderArgs, startDate)
	}
	if !endDate.IsZero() {
		orderQuery += " AND o.created_at <= ?"
		orderArgs = append(orderArgs, endDate)
	}

	orderQuery += " ORDER BY pickup_time DESC, o.id DESC"

	orderRows, err := database.DB.Query(orderQuery, orderArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询订单失败: " + err.Error()})
		return
	}
	defer orderRows.Close()

	orders := []map[string]interface{}{}
	totalAmount := 0.0

	for orderRows.Next() {
		var orderID int
		var orderNumber, orderStatus string
		var createdAt, pickupTime time.Time

		err := orderRows.Scan(&orderID, &orderNumber, &orderStatus, &createdAt, &pickupTime)
		if err != nil {
			continue
		}

		// 查询该订单中该供应商的已取货商品明细（未付款的）
		itemsQuery := `
			SELECT 
				oi.id as order_item_id,
				oi.product_id,
				oi.product_name,
				oi.spec_name,
				oi.quantity,
				p.specs as product_specs
			FROM order_items oi
			INNER JOIN products p ON oi.product_id = p.id
			WHERE oi.order_id = ? 
				AND p.supplier_id = ?
				AND oi.is_picked = 1
		`
		itemRows, err := database.DB.Query(itemsQuery, orderID, supplierID)
		if err != nil {
			continue
		}

		items := []map[string]interface{}{}
		orderTotalCost := 0.0

		for itemRows.Next() {
			var orderItemID, productID, quantity int
			var productName, specName string
			var productSpecsJSON sql.NullString

			err := itemRows.Scan(&orderItemID, &productID, &productName, &specName, &quantity, &productSpecsJSON)
			if err != nil {
				continue
			}

			// 跳过已付款的商品
			if paidItems[orderItemID] {
				continue
			}

			// 计算成本价
			costPrice := 0.0
			if productSpecsJSON.Valid && productSpecsJSON.String != "" {
				var specs []model.Spec
				if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
					for _, spec := range specs {
						if spec.Name == specName {
							costPrice = spec.Cost
							break
						}
					}
					if costPrice == 0 && len(specs) > 0 {
						costPrice = specs[0].Cost
					}
				}
			}

			subtotal := costPrice * float64(quantity)
			orderTotalCost += subtotal

			items = append(items, map[string]interface{}{
				"order_item_id": orderItemID,
				"product_id":    productID,
				"product_name":  productName,
				"spec_name":     specName,
				"quantity":      quantity,
				"cost_price":    costPrice,
				"subtotal":      subtotal,
			})
		}
		itemRows.Close()

		if len(items) == 0 {
			continue
		}

		totalAmount += orderTotalCost

		orderData := map[string]interface{}{
			"order_id":     orderID,
			"order_number": orderNumber,
			"order_date":   createdAt.Format("2006-01-02 15:04:05"),
			"pickup_date":  pickupTime.Format("2006-01-02 15:04:05"),
			"status":       orderStatus,
			"items":        items,
			"total_cost":   orderTotalCost,
		}

		orders = append(orders, orderData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"total_amount": totalAmount,
			"order_count":  len(orders),
			"orders":       orders,
		},
		"message": "success",
	})
}

// GetSupplierPaymentStats 供应商获取对账统计
func GetSupplierPaymentStats(c *gin.Context) {
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 获取查询参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// 构建时间范围
	var startDate, endDate time.Time
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
		}
	}

	// 获取已付款的订单商品ID列表
	paidItems, err := model.GetPaidOrderItemIDs(&supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}

	// 计算已付款总额
	paidAmountQuery := `
		SELECT COALESCE(SUM(payment_amount), 0)
		FROM supplier_payments
		WHERE supplier_id = ? AND status = 1
	`
	paidArgs := []interface{}{supplierID}
	if !startDate.IsZero() {
		paidAmountQuery += " AND payment_date >= ?"
		paidArgs = append(paidArgs, startDate)
	}
	if !endDate.IsZero() {
		paidAmountQuery += " AND payment_date <= ?"
		paidArgs = append(paidArgs, endDate)
	}

	var paidAmount float64
	database.DB.QueryRow(paidAmountQuery, paidArgs...).Scan(&paidAmount)

	// 计算待付款总额（已取货但未付款的商品）
	pendingQuery := `
		SELECT 
			oi.id,
			oi.spec_name,
			oi.quantity,
			p.specs as product_specs
		FROM order_items oi
		INNER JOIN products p ON oi.product_id = p.id
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE p.supplier_id = ? 
			AND oi.is_picked = 1
			AND o.status != 'cancelled'
	`
	pendingArgs := []interface{}{supplierID}
	if !startDate.IsZero() {
		pendingQuery += " AND o.created_at >= ?"
		pendingArgs = append(pendingArgs, startDate)
	}
	if !endDate.IsZero() {
		pendingQuery += " AND o.created_at <= ?"
		pendingArgs = append(pendingArgs, endDate)
	}

	pendingRows, err := database.DB.Query(pendingQuery, pendingArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}
	defer pendingRows.Close()

	pendingAmount := 0.0
	for pendingRows.Next() {
		var orderItemID, quantity int
		var specName string
		var productSpecsJSON sql.NullString

		if err := pendingRows.Scan(&orderItemID, &specName, &quantity, &productSpecsJSON); err != nil {
			continue
		}

		// 跳过已付款的商品
		if paidItems[orderItemID] {
			continue
		}

		// 计算成本价
		costPrice := 0.0
		if productSpecsJSON.Valid && productSpecsJSON.String != "" {
			var specs []model.Spec
			if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
				for _, spec := range specs {
					if spec.Name == specName {
						costPrice = spec.Cost
						break
					}
				}
				if costPrice == 0 && len(specs) > 0 {
					costPrice = specs[0].Cost
				}
			}
		}

		pendingAmount += costPrice * float64(quantity)
	}

	totalAmount := paidAmount + pendingAmount

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"total_amount":   totalAmount,
			"paid_amount":    paidAmount,
			"pending_amount": pendingAmount,
		},
		"message": "success",
	})
}
