package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetSupplierPaymentsStats 获取供应商付款统计列表
func GetSupplierPaymentsStats(c *gin.Context) {
	// 获取查询参数
	supplierIDStr := c.Query("supplier_id")
	status := c.Query("status") // pending/paid（统计列表不使用状态筛选，但保留参数以兼容）
	pageNum := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNum = p
		}
	}
	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	// 统计列表不进行时间筛选，显示全部数据
	_ = status // 统计列表不使用状态筛选

	// 先获取所有供应商
	supplierQuery := "SELECT id, name FROM suppliers WHERE status = 1"
	if supplierIDStr != "" {
		supplierID, err := strconv.Atoi(supplierIDStr)
		if err == nil {
			supplierQuery += fmt.Sprintf(" AND id = %d", supplierID)
		}
	}
	supplierRows, err := database.DB.Query(supplierQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询供应商失败: " + err.Error(),
		})
		return
	}
	defer supplierRows.Close()

	// 获取所有供应商ID用于分页
	allSupplierIDs := []int{}
	allSupplierNames := make(map[int]string)
	for supplierRows.Next() {
		var supplierID int
		var supplierName string
		if err := supplierRows.Scan(&supplierID, &supplierName); err == nil {
			allSupplierIDs = append(allSupplierIDs, supplierID)
			allSupplierNames[supplierID] = supplierName
		}
	}
	supplierRows.Close()

	// 分页处理
	totalSuppliers := len(allSupplierIDs)
	start := (pageNum - 1) * pageSize
	end := start + pageSize
	if start > totalSuppliers {
		start = totalSuppliers
	}
	if end > totalSuppliers {
		end = totalSuppliers
	}

	// 获取当前页的供应商ID
	currentPageSupplierIDs := allSupplierIDs[start:end]

	stats := []map[string]interface{}{}
	for _, supplierID := range currentPageSupplierIDs {
		supplierName := allSupplierNames[supplierID]

		// 获取已付款的订单商品ID列表和已付款总金额
		paidItems, err := model.GetPaidOrderItemIDs(&supplierID)
		if err != nil {
			continue
		}

		// 查询已付款总金额（从付款记录表，统计列表显示全部时间的已付款金额）
		var paidAmount float64
		paidAmountQuery := `
			SELECT COALESCE(SUM(payment_amount), 0)
			FROM supplier_payments
			WHERE supplier_id = ? AND status = 1
		`
		database.DB.QueryRow(paidAmountQuery, supplierID).Scan(&paidAmount)

		// 查询该供应商的所有已取货商品（通过 order_items.is_picked = 1）
		// 注意：只要商品被取走就要付款，不等待订单结算
		// 统计列表显示全部时间的已取货商品
		orderQuery := `
			SELECT DISTINCT oi.order_id
			FROM order_items oi
			INNER JOIN products p ON oi.product_id = p.id
			INNER JOIN orders o ON oi.order_id = o.id
			WHERE p.supplier_id = ? 
				AND oi.is_picked = 1
				AND o.status != 'cancelled'
		`

		orderRows, err := database.DB.Query(orderQuery, supplierID)
		if err != nil {
			continue
		}

		var orderIDs []int
		for orderRows.Next() {
			var orderID int
			if err := orderRows.Scan(&orderID); err == nil {
				orderIDs = append(orderIDs, orderID)
			}
		}
		orderRows.Close()

		// 计算所有已取货商品的总金额（包括已付款和未付款）
		totalAmount := 0.0
		pendingAmount := 0.0

		for _, orderID := range orderIDs {
			// 查询该订单中该供应商的所有已取货商品成本（包括已付款和未付款）
			itemsQuery := `
				SELECT 
					oi.id as order_item_id,
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

			orderCost := 0.0
			orderPendingCost := 0.0
			for itemRows.Next() {
				var orderItemID int
				var specName string
				var quantity int
				var productSpecsJSON sql.NullString

				if err := itemRows.Scan(&orderItemID, &specName, &quantity, &productSpecsJSON); err != nil {
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

				itemCost := costPrice * float64(quantity)
				orderCost += itemCost

				// 如果商品未付款，算作待付款（在时间范围内的所有未付款商品）
				if !paidItems[orderItemID] {
					orderPendingCost += itemCost
				}
			}
			itemRows.Close()

			totalAmount += orderCost
			pendingAmount += orderPendingCost
		}

		// 如果查询的是已付款状态，但已付款金额为0，则跳过
		if status == "paid" && paidAmount == 0 {
			continue
		}
		// 如果查询的是待付款状态，但待付款金额为0，则跳过
		if status == "pending" && pendingAmount == 0 {
			continue
		}
		// 如果查询全部，但总金额为0，则跳过
		if status == "" && totalAmount == 0 {
			continue
		}

		// 计算付款状态
		paymentStatus := "pending" // 默认待付款
		if paidAmount > 0 && pendingAmount == 0 {
			paymentStatus = "paid" // 全部已付款
		} else if paidAmount > 0 && pendingAmount > 0 {
			paymentStatus = "partial" // 部分已付款
		}

		stat := map[string]interface{}{
			"supplier_id":    supplierID,
			"supplier_name":  supplierName,
			"total_amount":   totalAmount,
			"order_count":    len(orderIDs),
			"pending_amount": pendingAmount,
			"paid_amount":    paidAmount,
			"payment_status": paymentStatus,
		}

		stats = append(stats, stat)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":      stats,
			"total":     totalSuppliers,
			"page":      pageNum,
			"page_size": pageSize,
		},
		"message": "success",
	})
}

// GetSupplierPaymentDetail 获取供应商详细付款清单
func GetSupplierPaymentDetail(c *gin.Context) {
	supplierIDStr := c.Param("id")
	supplierID, err := strconv.Atoi(supplierIDStr)
	if err != nil || supplierID <= 0 {
		badRequestResponse(c, "无效的供应商ID")
		return
	}

	// 获取供应商信息
	supplier, err := model.GetSupplierByID(database.DB, supplierID)
	if err != nil || supplier == nil {
		notFoundResponse(c, "供应商不存在")
		return
	}

	// 获取查询参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	status := c.Query("status") // pending/paid
	pageNum := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNum = p
		}
	}
	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
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
	// 注意：只要商品被取走就要付款，不等待订单结算
	orderQuery := `
		SELECT DISTINCT
			o.id,
			o.order_number,
			o.status,
			o.created_at,
			o.updated_at,
			o.settlement_date,
			COALESCE(
				(SELECT action_time 
				 FROM delivery_logs 
				 WHERE order_id = o.id 
				   AND action = 'pickup_completed' 
				 ORDER BY action_time DESC 
				 LIMIT 1),
				o.updated_at
			) as pickup_time,
			a.name AS address_name
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		LEFT JOIN mini_app_addresses a ON o.address_id = a.id
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

	// 注意：不再根据取货时间过滤订单，而是查询所有订单，然后在商品明细层面根据实际付款状态过滤
	orderQuery += " ORDER BY pickup_time DESC, o.id DESC"

	// 先查询所有符合条件的订单ID（用于分页）
	allOrderRows, err := database.DB.Query(orderQuery, orderArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询订单失败: " + err.Error(),
		})
		return
	}

	// 获取所有订单ID（用于后续过滤和分页）
	allOrderIDs := []int{}
	allOrderData := make(map[int]map[string]interface{})
	for allOrderRows.Next() {
		var orderID int
		var orderNumber, orderStatus string
		var createdAt, updatedAt, pickupTime time.Time
		var settlementDate sql.NullTime
		var addressName sql.NullString

		err := allOrderRows.Scan(&orderID, &orderNumber, &orderStatus, &createdAt, &updatedAt, &settlementDate, &pickupTime, &addressName)
		if err != nil {
			continue
		}

		allOrderIDs = append(allOrderIDs, orderID)
		allOrderData[orderID] = map[string]interface{}{
			"order_id":        orderID,
			"order_number":    orderNumber,
			"status":          orderStatus,
			"created_at":      createdAt,
			"updated_at":      updatedAt,
			"pickup_time":     pickupTime,
			"settlement_date": settlementDate,
			"address_name":    addressName,
		}
	}
	allOrderRows.Close()

	// 获取已付款的订单商品ID列表
	paidItems, err := model.GetPaidOrderItemIDs(&supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询已付款商品失败: " + err.Error(),
		})
		return
	}

	// 过滤订单（根据状态和商品明细）
	filteredOrderIDs := []int{}
	for _, orderID := range allOrderIDs {
		// 查询该订单中该供应商的所有已取货商品明细
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

		orderPaidCost := 0.0
		orderPendingCost := 0.0
		hasItems := false

		for itemRows.Next() {
			var orderItemID, productID, quantity int
			var productName, specName string
			var productSpecsJSON sql.NullString

			err := itemRows.Scan(&orderItemID, &productID, &productName, &specName, &quantity, &productSpecsJSON)
			if err != nil {
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
			isPaid := paidItems[orderItemID]

			// 根据status参数过滤
			if status == "paid" && !isPaid {
				continue // 查询已付款时，跳过未付款的商品
			}
			if status == "pending" && isPaid {
				continue // 查询待付款时，跳过已付款的商品
			}

			hasItems = true
			if isPaid {
				orderPaidCost += subtotal
			} else {
				orderPendingCost += subtotal
			}
		}
		itemRows.Close()

		// 如果查询已付款，但订单中没有已付款商品，则跳过
		if status == "paid" && orderPaidCost == 0 {
			continue
		}
		// 如果查询待付款，但订单中没有待付款商品，则跳过
		if status == "pending" && orderPendingCost == 0 {
			continue
		}
		// 如果查询全部，但订单中没有商品（被过滤后），则跳过
		if !hasItems {
			continue
		}

		filteredOrderIDs = append(filteredOrderIDs, orderID)
	}

	// 分页处理
	totalOrders := len(filteredOrderIDs)
	start := (pageNum - 1) * pageSize
	end := start + pageSize
	if start > totalOrders {
		start = totalOrders
	}
	if end > totalOrders {
		end = totalOrders
	}

	// 获取当前页的订单ID
	currentPageOrderIDs := filteredOrderIDs[start:end]

	orders := []map[string]interface{}{}
	totalAmount := 0.0

	for _, orderID := range currentPageOrderIDs {
		orderData := allOrderData[orderID]
		orderIDInt := orderData["order_id"].(int)
		orderNumber := orderData["order_number"].(string)
		orderStatus := orderData["status"].(string)
		createdAt := orderData["created_at"].(time.Time)
		pickupTime := orderData["pickup_time"].(time.Time)
		settlementDate := orderData["settlement_date"].(sql.NullTime)
		addressName := orderData["address_name"].(sql.NullString)

		// 查询该订单中该供应商的所有已取货商品明细（包括已付款和未付款）
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
		orderPaidCost := 0.0
		orderPendingCost := 0.0

		for itemRows.Next() {
			var orderItemID, productID, quantity int
			var productName, specName string
			var productSpecsJSON sql.NullString

			err := itemRows.Scan(&orderItemID, &productID, &productName, &specName, &quantity, &productSpecsJSON)
			if err != nil {
				continue
			}

			// 计算成本价
			costPrice := 0.0
			if productSpecsJSON.Valid && productSpecsJSON.String != "" {
				var specs []model.Spec
				if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
					// 优先根据规格名称匹配
					for _, spec := range specs {
						if spec.Name == specName {
							costPrice = spec.Cost
							break
						}
					}
					// 如果没找到匹配的规格，使用第一个规格的成本价
					if costPrice == 0 && len(specs) > 0 {
						costPrice = specs[0].Cost
					}
				}
			}

			subtotal := costPrice * float64(quantity)
			orderTotalCost += subtotal

			// 判断商品是否已付款
			isPaid := paidItems[orderItemID]
			if isPaid {
				orderPaidCost += subtotal
			} else {
				orderPendingCost += subtotal
			}

			// 根据status参数过滤
			if status == "paid" && !isPaid {
				continue // 查询已付款时，跳过未付款的商品
			}
			if status == "pending" && isPaid {
				continue // 查询待付款时，跳过已付款的商品
			}

			items = append(items, map[string]interface{}{
				"order_item_id": orderItemID,
				"product_id":    productID,
				"product_name":  productName,
				"spec_name":     specName,
				"quantity":      quantity,
				"cost_price":    costPrice,
				"subtotal":      subtotal,
				"is_paid":       isPaid,
			})
		}
		itemRows.Close()

		// 如果查询已付款，但订单中没有已付款商品，则跳过
		if status == "paid" && orderPaidCost == 0 {
			continue
		}
		// 如果查询待付款，但订单中没有待付款商品，则跳过
		if status == "pending" && orderPendingCost == 0 {
			continue
		}
		// 如果查询全部，但订单中没有商品（被过滤后），则跳过
		if len(items) == 0 {
			continue
		}

		// 根据status参数决定使用哪个金额
		var orderDisplayCost float64
		if status == "paid" {
			orderDisplayCost = orderPaidCost
		} else if status == "pending" {
			orderDisplayCost = orderPendingCost
		} else {
			orderDisplayCost = orderTotalCost
		}

		// 计算订单的付款状态
		orderPaymentStatus := "pending" // 默认待付款
		if orderPaidCost > 0 && orderPendingCost == 0 {
			orderPaymentStatus = "paid" // 全部已付款
		} else if orderPaidCost > 0 && orderPendingCost > 0 {
			orderPaymentStatus = "partial" // 部分已付款
		}

		totalAmount += orderDisplayCost

		orderDataMap := map[string]interface{}{
			"order_id":        orderIDInt,
			"order_number":    orderNumber,
			"order_date":      createdAt.Format("2006-01-02 15:04:05"),
			"pickup_date":     pickupTime.Format("2006-01-02 15:04:05"), // 取货时间（从 delivery_logs 获取，或使用订单更新时间）
			"settlement_date": nil,
			"status":          orderStatus,
			"items":           items,
			"total_cost":      orderDisplayCost,
			"paid_cost":       orderPaidCost,
			"pending_cost":    orderPendingCost,
			"payment_status":  orderPaymentStatus, // 订单对供应商的付款状态
			"address_name": func() string {
				if addressName.Valid {
					return addressName.String
				}
				return ""
			}(),
		}

		if settlementDate.Valid {
			orderDataMap["settlement_date"] = settlementDate.Time.Format("2006-01-02 15:04:05")
		}

		orders = append(orders, orderDataMap)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"supplier_id":   supplier.ID,
			"supplier_name": supplier.Name,
			"total_amount":  totalAmount,
			"order_count":   totalOrders, // 使用过滤后的总数
			"orders":        orders,
			"total":         totalOrders,
			"page":          pageNum,
			"page_size":     pageSize,
		},
		"message": "success",
	})
}

// CreateSupplierPayment 创建供应商付款记录（管理员）
func CreateSupplierPayment(c *gin.Context) {
	// 从上下文获取管理员信息（AuthMiddleware 已设置）
	adminUsernameInterface, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}
	adminUsername := adminUsernameInterface.(string)

	var req struct {
		SupplierID     int     `json:"supplier_id" binding:"required"`
		PaymentDate    string  `json:"payment_date" binding:"required"`
		PaymentAmount  float64 `json:"payment_amount" binding:"required"`
		PaymentMethod  *string `json:"payment_method"`
		PaymentAccount *string `json:"payment_account"`
		PaymentReceipt *string `json:"payment_receipt"`
		Remark         *string `json:"remark"`
		OrderItems     []struct {
			OrderID     int     `json:"order_id" binding:"required"`
			OrderItemID int     `json:"order_item_id" binding:"required"`
			ProductID   int     `json:"product_id" binding:"required"`
			ProductName string  `json:"product_name" binding:"required"`
			SpecName    string  `json:"spec_name"`
			Quantity    int     `json:"quantity" binding:"required"`
			CostPrice   float64 `json:"cost_price" binding:"required"`
			Subtotal    float64 `json:"subtotal" binding:"required"`
		} `json:"order_items" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// 解析付款日期
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		badRequestResponse(c, "付款日期格式错误")
		return
	}

	// 验证金额
	calculatedAmount := 0.0
	for _, item := range req.OrderItems {
		calculatedAmount += item.Subtotal
	}
	if calculatedAmount != req.PaymentAmount {
		badRequestResponse(c, fmt.Sprintf("付款金额不匹配，计算金额: %.2f，提交金额: %.2f", calculatedAmount, req.PaymentAmount))
		return
	}

	// 检查订单商品是否已付款
	for _, item := range req.OrderItems {
		paid, err := model.CheckOrderItemPaid(item.OrderItemID)
		if err != nil {
			internalErrorResponse(c, "检查付款状态失败: "+err.Error())
			return
		}
		if paid {
			badRequestResponse(c, fmt.Sprintf("订单商品 %d 已付款，不能重复付款", item.OrderItemID))
			return
		}
	}

	// 创建付款记录
	payment := &model.SupplierPayment{
		SupplierID:     req.SupplierID,
		PaymentDate:    paymentDate,
		PaymentAmount:  req.PaymentAmount,
		PaymentMethod:  req.PaymentMethod,
		PaymentAccount: req.PaymentAccount,
		PaymentReceipt: req.PaymentReceipt,
		Remark:         req.Remark,
		CreatedBy:      &adminUsername,
		Status:         1,
	}

	items := make([]model.SupplierPaymentItem, len(req.OrderItems))
	for i, item := range req.OrderItems {
		items[i] = model.SupplierPaymentItem{
			OrderID:     item.OrderID,
			OrderItemID: item.OrderItemID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			SpecName:    item.SpecName,
			Quantity:    item.Quantity,
			CostPrice:   item.CostPrice,
			Subtotal:    item.Subtotal,
		}
	}

	paymentID, err := model.CreateSupplierPayment(payment, items)
	if err != nil {
		internalErrorResponse(c, "创建付款记录失败: "+err.Error())
		return
	}

	// 获取完整的付款记录（包含ID和明细）
	createdPayment, err := model.GetSupplierPaymentByID(int(paymentID))
	if err != nil || createdPayment == nil {
		internalErrorResponse(c, "获取付款记录失败: "+err.Error())
		return
	}

	// 获取付款明细
	paymentItems, _ := model.GetSupplierPaymentItems(int(paymentID))
	responseData := map[string]interface{}{
		"id":              createdPayment.ID,
		"supplier_id":     createdPayment.SupplierID,
		"payment_date":    createdPayment.PaymentDate.Format("2006-01-02"),
		"payment_amount":  createdPayment.PaymentAmount,
		"payment_method":  createdPayment.PaymentMethod,
		"payment_account": createdPayment.PaymentAccount,
		"payment_receipt": createdPayment.PaymentReceipt,
		"remark":          createdPayment.Remark,
		"created_by":      createdPayment.CreatedBy,
		"status":          createdPayment.Status,
		"created_at":      createdPayment.CreatedAt.Format("2006-01-02 15:04:05"),
		"items":           paymentItems,
	}

	successResponse(c, responseData, "付款记录创建成功")
}

// GetSupplierPayments 获取供应商付款记录列表（管理员）
func GetSupplierPayments(c *gin.Context) {
	supplierIDStr := c.Query("supplier_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	pageNum := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNum = p
		}
	}
	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	var supplierID *int
	if supplierIDStr != "" {
		if id, err := strconv.Atoi(supplierIDStr); err == nil {
			supplierID = &id
		}
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

	payments, total, err := model.GetSupplierPayments(supplierID, startDate, endDate, pageNum, pageSize)
	if err != nil {
		internalErrorResponse(c, "获取付款记录失败: "+err.Error())
		return
	}

	// 获取每个付款记录的明细
	paymentList := make([]map[string]interface{}, len(payments))
	for i, payment := range payments {
		items, _ := model.GetSupplierPaymentItems(payment.ID)
		paymentList[i] = map[string]interface{}{
			"id":              payment.ID,
			"supplier_id":     payment.SupplierID,
			"payment_date":    payment.PaymentDate.Format("2006-01-02"),
			"payment_amount":  payment.PaymentAmount,
			"payment_method":  payment.PaymentMethod,
			"payment_account": payment.PaymentAccount,
			"payment_receipt": payment.PaymentReceipt,
			"remark":          payment.Remark,
			"created_by":      payment.CreatedBy,
			"status":          payment.Status,
			"created_at":      payment.CreatedAt.Format("2006-01-02 15:04:05"),
			"items":           items,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":      paymentList,
			"total":     total,
			"page":      pageNum,
			"page_size": pageSize,
		},
		"message": "success",
	})
}

// CancelSupplierPayment 撤销付款（管理员）
func CancelSupplierPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		badRequestResponse(c, "无效的付款记录ID")
		return
	}

	if err := model.CancelSupplierPayment(id); err != nil {
		internalErrorResponse(c, "撤销付款失败: "+err.Error())
		return
	}

	successResponse(c, nil, "撤销付款成功")
}
