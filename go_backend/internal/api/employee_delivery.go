package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/model"
	"go_backend/internal/utils"

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
		} else if status == "pending_pickup" {
			where = "status = ?"
			args = append(args, status)
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

	// 获取订单商品对应的供应商信息，并关联订单商品
	// 批量查询商品对应的供应商信息（去重）
	suppliersList := make([]map[string]interface{}, 0)
	if len(items) > 0 {
		// 收集所有商品ID
		productIDs := make([]int, 0, len(items))
		for _, item := range items {
			productIDs = append(productIDs, item.ProductID)
		}

		// 批量查询商品对应的供应商信息
		placeholders := make([]string, len(productIDs))
		args := make([]interface{}, len(productIDs))
		for i, pid := range productIDs {
			placeholders[i] = "?"
			args[i] = pid
		}

		query := fmt.Sprintf(`
			SELECT DISTINCT s.id, s.name, s.contact, s.phone, s.email, s.address, s.latitude, s.longitude
			FROM products p
			INNER JOIN suppliers s ON p.supplier_id = s.id
			WHERE p.id IN (%s) AND p.supplier_id IS NOT NULL
			ORDER BY s.id
		`, strings.Join(placeholders, ","))

		rows, err := database.DB.Query(query, args...)
		if err == nil {
			defer rows.Close()
			supplierIDSet := make(map[int]bool)
			supplierMap := make(map[int]*model.Supplier) // supplier_id -> Supplier
			for rows.Next() {
				var supplier model.Supplier
				var latitude, longitude sql.NullFloat64
				err := rows.Scan(
					&supplier.ID,
					&supplier.Name,
					&supplier.Contact,
					&supplier.Phone,
					&supplier.Email,
					&supplier.Address,
					&latitude,
					&longitude,
				)
				if err == nil && !supplierIDSet[supplier.ID] {
					supplierIDSet[supplier.ID] = true
					if latitude.Valid {
						supplier.Latitude = &latitude.Float64
					}
					if longitude.Valid {
						supplier.Longitude = &longitude.Float64
					}
					supplierMap[supplier.ID] = &supplier
				}
			}

			// 为每个供应商查询对应的订单商品
			for supplierID, supplier := range supplierMap {
				// 查询该供应商对应的订单商品
				supplierItems := make([]map[string]interface{}, 0)
				for _, item := range items {
					// 查询该商品是否属于该供应商
					var productSupplierID sql.NullInt64
					err := database.DB.QueryRow("SELECT supplier_id FROM products WHERE id = ?", item.ProductID).Scan(&productSupplierID)
					if err == nil && productSupplierID.Valid && int(productSupplierID.Int64) == supplierID {
						supplierItems = append(supplierItems, map[string]interface{}{
							"product_name": item.ProductName,
							"spec_name":    item.SpecName,
							"quantity":     item.Quantity,
							"image":        item.Image,
							"is_picked":    item.IsPicked, // 添加取货状态
						})
					}
				}

				// 构建供应商数据（包含商品列表）
				supplierData := map[string]interface{}{
					"id":        supplier.ID,
					"name":      supplier.Name,
					"contact":   supplier.Contact,
					"phone":     supplier.Phone,
					"email":     supplier.Email,
					"address":   supplier.Address,
					"latitude":  supplier.Latitude,
					"longitude": supplier.Longitude,
					"items":     supplierItems, // 该供应商对应的订单商品列表
				}
				suppliersList = append(suppliersList, supplierData)
			}
		}
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
			"suppliers":                suppliersList, // 供应商列表
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

	// 更新订单状态为待取货，并记录配送员信息
	err = model.UpdateOrderStatusWithDeliveryEmployee(id, "pending_pickup", employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "接单失败: " + err.Error()})
		return
	}

	// 记录配送流程日志：接单
	deliveryLog := &model.DeliveryLog{
		OrderID:             id,
		Action:              model.DeliveryLogActionAccepted,
		DeliveryEmployeeCode: &employee.EmployeeCode,
		ActionTime:          time.Now(),
	}
	_ = model.CreateDeliveryLog(deliveryLog) // 记录日志失败不影响主流程

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "接单成功",
	})
}

// StartDeliveryOrder 开始配送（从待取货状态转为配送中）
func StartDeliveryOrder(c *gin.Context) {
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
	if order.Status != "pending_pickup" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只能开始待取货的订单"})
		return
	}

	// 更新订单状态为配送中
	err = model.UpdateOrderStatus(id, "delivering")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "开始配送失败: " + err.Error()})
		return
	}

	// 记录配送流程日志：开始配送
	deliveryLog := &model.DeliveryLog{
		OrderID:             id,
		Action:              model.DeliveryLogActionDeliveringStarted,
		DeliveryEmployeeCode: &employee.EmployeeCode,
		ActionTime:          time.Now(),
	}
	_ = model.CreateDeliveryLog(deliveryLog) // 记录日志失败不影响主流程

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "开始配送成功",
	})
}

// CompleteDeliveryOrder 完成配送（支持上传图片）
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

	// 验证配送员是否匹配
	if order.DeliveryEmployeeCode == nil || *order.DeliveryEmployeeCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您无权完成此订单的配送"})
		return
	}

	var productImageURL, doorplateImageURL *string

	// 上传货物照片
	if _, _, err := c.Request.FormFile("product_image"); err == nil {
		url, uploadErr := utils.UploadFileByFieldName("product_image", "delivery_product", c.Request)
		if uploadErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "上传货物照片失败: " + uploadErr.Error()})
			return
		}
		productImageURL = &url
	}

	// 上传门牌照片
	if _, _, err := c.Request.FormFile("doorplate_image"); err == nil {
		url, uploadErr := utils.UploadFileByFieldName("doorplate_image", "delivery_doorplate", c.Request)
		if uploadErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "上传门牌照片失败: " + uploadErr.Error()})
			return
		}
		doorplateImageURL = &url
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "开启事务失败: " + err.Error()})
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 更新订单状态为已送达
	_, err = tx.Exec("UPDATE orders SET status = 'delivered', updated_at = NOW() WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新订单状态失败: " + err.Error()})
		return
	}

	// 创建配送记录
	deliveryRecord := &model.DeliveryRecord{
		OrderID:             id,
		DeliveryEmployeeCode: employee.EmployeeCode,
		ProductImageURL:     productImageURL,
		DoorplateImageURL:   doorplateImageURL,
		CompletedAt:         time.Now(),
	}

	// 记录配送流程日志：配送完成
	remark := "配送完成，已上传货物照片和门牌照片"
	deliveryLog := &model.DeliveryLog{
		OrderID:             id,
		Action:              model.DeliveryLogActionDeliveringCompleted,
		DeliveryEmployeeCode: &employee.EmployeeCode,
		ActionTime:          time.Now(),
		Remark:              &remark,
	}
	_ = model.CreateDeliveryLog(deliveryLog) // 记录日志失败不影响主流程

	_, err = tx.Exec(`
		INSERT INTO delivery_records (
			order_id, delivery_employee_code, product_image_url, doorplate_image_url, completed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`, deliveryRecord.OrderID, deliveryRecord.DeliveryEmployeeCode,
		deliveryRecord.ProductImageURL, deliveryRecord.DoorplateImageURL, deliveryRecord.CompletedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建配送记录失败: " + err.Error()})
		return
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交事务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "配送完成",
		"data": gin.H{
			"delivery_record": deliveryRecord,
		},
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

// GetPickupSuppliers 获取配送员待取货订单的供应商列表
func GetPickupSuppliers(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	// 查询该配送员所有待取货订单的商品对应的供应商（去重）
	query := `
		SELECT DISTINCT s.id, s.name, s.contact, s.phone, s.email, s.address, s.latitude, s.longitude
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		INNER JOIN suppliers s ON p.supplier_id = s.id
		WHERE o.delivery_employee_code = ? 
		  AND o.status = 'pending_pickup'
		  AND oi.is_picked = 0
		  AND p.supplier_id IS NOT NULL
		ORDER BY s.id
	`

	rows, err := database.DB.Query(query, employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取供应商列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	suppliersList := make([]map[string]interface{}, 0)
	for rows.Next() {
		var supplier model.Supplier
		var latitude, longitude sql.NullFloat64
		err := rows.Scan(
			&supplier.ID,
			&supplier.Name,
			&supplier.Contact,
			&supplier.Phone,
			&supplier.Email,
			&supplier.Address,
			&latitude,
			&longitude,
		)
		if err == nil {
			if latitude.Valid {
				supplier.Latitude = &latitude.Float64
			}
			if longitude.Valid {
				supplier.Longitude = &longitude.Float64
			}
			supplierData := map[string]interface{}{
				"id":        supplier.ID,
				"name":      supplier.Name,
				"contact":   supplier.Contact,
				"phone":     supplier.Phone,
				"email":     supplier.Email,
				"address":   supplier.Address,
				"latitude":  supplier.Latitude,
				"longitude": supplier.Longitude,
			}
			suppliersList = append(suppliersList, supplierData)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    suppliersList,
		"message": "获取成功",
	})
}

// GetPickupItemsBySupplier 获取指定供应商的待取货商品列表
func GetPickupItemsBySupplier(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	supplierIDStr := c.Param("supplierId")
	supplierID, err := strconv.Atoi(supplierIDStr)
	if err != nil || supplierID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "供应商ID格式错误"})
		return
	}

	// 查询该配送员在该供应商的待取货商品
	query := `
		SELECT 
			oi.id,
			oi.order_id,
			oi.product_id,
			oi.product_name,
			oi.spec_name,
			oi.quantity,
			oi.unit_price,
			oi.subtotal,
			oi.image,
			oi.is_picked,
			o.order_number,
			o.id as order_id
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE o.delivery_employee_code = ? 
		  AND o.status = 'pending_pickup'
		  AND oi.is_picked = 0
		  AND p.supplier_id = ?
		ORDER BY o.id, oi.id
	`

	rows, err := database.DB.Query(query, employee.EmployeeCode, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	itemsList := make([]map[string]interface{}, 0)
	for rows.Next() {
		var itemID, orderID, productID, quantity, isPickedTinyInt int
		var productName, specName, image, orderNumber string
		var unitPrice, subtotal float64

		err := rows.Scan(
			&itemID, &orderID, &productID, &productName, &specName,
			&quantity, &unitPrice, &subtotal, &image, &isPickedTinyInt,
			&orderNumber, &orderID,
		)
		if err == nil {
			itemData := map[string]interface{}{
				"id":           itemID,
				"order_id":     orderID,
				"order_number": orderNumber,
				"product_id":   productID,
				"product_name": productName,
				"spec_name":    specName,
				"quantity":     quantity,
				"unit_price":   unitPrice,
				"subtotal":     subtotal,
				"image":        image,
				"is_picked":    isPickedTinyInt == 1,
			}
			itemsList = append(itemsList, itemData)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    itemsList,
		"message": "获取成功",
	})
}

// MarkItemsAsPicked 标记商品已取货
func MarkItemsAsPicked(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	var req struct {
		ItemIDs []int `json:"item_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	if len(req.ItemIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请至少选择一个商品"})
		return
	}

	// 验证这些商品是否属于该配送员的待取货订单
	placeholders := make([]string, len(req.ItemIDs))
	args := make([]interface{}, len(req.ItemIDs)+1)
	args[0] = employee.EmployeeCode
	for i, itemID := range req.ItemIDs {
		placeholders[i] = "?"
		args[i+1] = itemID
	}

	checkQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.delivery_employee_code = ? 
		  AND o.status = 'pending_pickup'
		  AND oi.id IN (%s)
	`, strings.Join(placeholders, ","))

	var count int
	err := database.DB.QueryRow(checkQuery, args...).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证商品失败: " + err.Error()})
		return
	}

	if count != len(req.ItemIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部分商品不属于您的待取货订单"})
		return
	}

	// 批量更新商品状态为已取货
	updatePlaceholders := make([]string, len(req.ItemIDs))
	updateArgs := make([]interface{}, len(req.ItemIDs))
	for i, itemID := range req.ItemIDs {
		updatePlaceholders[i] = "?"
		updateArgs[i] = itemID
	}
	updateQuery := fmt.Sprintf(`
		UPDATE order_items 
		SET is_picked = 1 
		WHERE id IN (%s)
	`, strings.Join(updatePlaceholders, ","))

	_, err = database.DB.Exec(updateQuery, updateArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "标记取货失败: " + err.Error()})
		return
	}

	// 获取这些商品所属的订单ID（去重）
	orderIDQuery := fmt.Sprintf(`
		SELECT DISTINCT order_id
		FROM order_items
		WHERE id IN (%s)
	`, strings.Join(updatePlaceholders, ","))

	orderIDRows, err := database.DB.Query(orderIDQuery, updateArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单ID失败: " + err.Error()})
		return
	}
	defer orderIDRows.Close()

	orderIDs := make([]int, 0)
	for orderIDRows.Next() {
		var orderID int
		if err := orderIDRows.Scan(&orderID); err == nil {
			orderIDs = append(orderIDs, orderID)
		}
	}

	// 检查每个订单是否所有商品都已取货，如果是则更新订单状态为 delivering
	for _, orderID := range orderIDs {
		// 检查该订单是否所有商品都已取货
		var totalItems, pickedItems int
		checkAllPickedQuery := `
			SELECT 
				COUNT(*) as total_items,
				SUM(CASE WHEN is_picked = 1 THEN 1 ELSE 0 END) as picked_items
			FROM order_items
			WHERE order_id = ?
		`
		err := database.DB.QueryRow(checkAllPickedQuery, orderID).Scan(&totalItems, &pickedItems)
		if err != nil {
			continue // 跳过出错的订单
		}

		// 如果所有商品都已取货，且订单状态为 pending_pickup，则更新为 delivering
		if totalItems > 0 && totalItems == pickedItems {
			// 先检查订单状态是否为 pending_pickup
			var currentStatus string
			var deliveryEmployeeCode sql.NullString
			err := database.DB.QueryRow("SELECT status, delivery_employee_code FROM orders WHERE id = ?", orderID).Scan(&currentStatus, &deliveryEmployeeCode)
			if err == nil && currentStatus == "pending_pickup" {
				// 更新订单状态为 delivering，同时更新 updated_at
				_, err = database.DB.Exec("UPDATE orders SET status = 'delivering', updated_at = NOW() WHERE id = ?", orderID)
				if err != nil {
					// 记录错误但不影响整体流程
					fmt.Printf("更新订单 %d 状态失败: %v\n", orderID, err)
				} else {
					// 记录配送流程日志：取货完成
					if deliveryEmployeeCode.Valid {
						remark := "所有商品已取货，自动转为配送中"
						deliveryLog := &model.DeliveryLog{
							OrderID:             orderID,
							Action:              model.DeliveryLogActionPickupCompleted,
							DeliveryEmployeeCode: &deliveryEmployeeCode.String,
							ActionTime:          time.Now(),
							Remark:              &remark,
						}
						_ = model.CreateDeliveryLog(deliveryLog) // 记录日志失败不影响主流程
					}

					// 订单状态更新后，重新计算配送费和利润（异步执行，避免阻塞）
					go func(orderID int) {
						_ = model.CalculateAndStoreOrderProfit(orderID)
					}(orderID)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "标记取货成功",
	})
}

// PlanDeliveryRoute 规划配送路线
func PlanDeliveryRoute(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	// 获取请求参数：配送员当前位置
	var req struct {
		OriginLatitude  float64 `json:"origin_latitude" binding:"required"`  // 起点纬度
		OriginLongitude float64 `json:"origin_longitude" binding:"required"` // 起点经度
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[PlanDeliveryRoute] 请求参数错误: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	fmt.Printf("[PlanDeliveryRoute] 收到路线规划请求 - 配送员: %s, 起点: %.6f,%.6f\n", employee.EmployeeCode, req.OriginLatitude, req.OriginLongitude)

	// 获取高德地图API Key
	mapSettings, err := model.GetMapSettings()
	if err != nil {
		fmt.Printf("[PlanDeliveryRoute] 获取地图配置失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取地图配置失败: " + err.Error(),
		})
		return
	}

	amapKey := mapSettings["amap_key"]
	if amapKey == "" {
		fmt.Printf("[PlanDeliveryRoute] 未配置高德地图API Key\n")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "未配置高德地图API Key，无法进行路线规划",
		})
		return
	}
	fmt.Printf("[PlanDeliveryRoute] 高德地图API Key已配置（长度: %d）\n", len(amapKey))

	// 查询该配送员所有配送中的订单
	query := `
		SELECT o.id, o.order_number, o.address_id, a.latitude, a.longitude, a.address, a.name
		FROM orders o
		INNER JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE o.delivery_employee_code = ? 
		  AND o.status = 'delivering'
		  AND a.latitude IS NOT NULL 
		  AND a.longitude IS NOT NULL
		ORDER BY o.created_at ASC
	`

	fmt.Printf("[PlanDeliveryRoute] 查询配送订单 - SQL: %s, 配送员: %s\n", query, employee.EmployeeCode)
	rows, err := database.DB.Query(query, employee.EmployeeCode)
	if err != nil {
		fmt.Printf("[PlanDeliveryRoute] 查询配送订单失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询配送订单失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	// 收集所有客户位置
	type CustomerLocation struct {
		OrderID   int
		OrderNumber string
		Latitude  float64
		Longitude float64
		Address   string
		Name      string
	}

	customers := make([]CustomerLocation, 0)
	customerCount := 0
	for rows.Next() {
		customerCount++
		var customer CustomerLocation
		var latitude, longitude sql.NullFloat64
		var address, name sql.NullString

		err := rows.Scan(
			&customer.OrderID,
			&customer.OrderNumber,
			&sql.NullInt64{}, // address_id
			&latitude,
			&longitude,
			&address,
			&name,
		)
		if err != nil {
			fmt.Printf("[PlanDeliveryRoute] 扫描订单数据失败 (第%d条): %v\n", customerCount, err)
			continue
		}
		
		if !latitude.Valid || !longitude.Valid {
			fmt.Printf("[PlanDeliveryRoute] 订单 %d (订单号: %s) 缺少经纬度信息\n", customer.OrderID, customer.OrderNumber)
			continue
		}
		
		customer.Latitude = latitude.Float64
		customer.Longitude = longitude.Float64
		if address.Valid {
			customer.Address = address.String
		}
		if name.Valid {
			customer.Name = name.String
		}
		customers = append(customers, customer)
		fmt.Printf("[PlanDeliveryRoute] 找到客户 - 订单ID: %d, 订单号: %s, 地址: %s, 坐标: %.6f,%.6f\n", 
			customer.OrderID, customer.OrderNumber, customer.Address, customer.Latitude, customer.Longitude)
	}

	fmt.Printf("[PlanDeliveryRoute] 共查询到 %d 条订单记录，有效客户 %d 个\n", customerCount, len(customers))

	if len(customers) == 0 {
		fmt.Printf("[PlanDeliveryRoute] 没有有效的配送订单（所有订单都缺少经纬度）\n")
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "暂无配送订单",
			"data":    nil,
		})
		return
	}

	// 计算起点到每个客户的距离，找到最远的客户作为目的地
	origin := utils.Coordinate{
		Latitude:  req.OriginLatitude,
		Longitude: req.OriginLongitude,
	}

	fmt.Printf("[PlanDeliveryRoute] 起点坐标: %.6f,%.6f\n", origin.Latitude, origin.Longitude)

	var farthestCustomer CustomerLocation
	var maxDistance float64

	for _, customer := range customers {
		// 计算距离（使用简单的欧几里得距离，实际应该使用地理距离）
		latDiff := customer.Latitude - origin.Latitude
		lngDiff := customer.Longitude - origin.Longitude
		distance := latDiff*latDiff + lngDiff*lngDiff

		if distance > maxDistance {
			maxDistance = distance
			farthestCustomer = customer
		}
	}

	fmt.Printf("[PlanDeliveryRoute] 最远客户 - 订单ID: %d, 订单号: %s, 坐标: %.6f,%.6f\n", 
		farthestCustomer.OrderID, farthestCustomer.OrderNumber, farthestCustomer.Latitude, farthestCustomer.Longitude)

	// 构建途经点列表（排除目的地）
	waypoints := make([]utils.Coordinate, 0)
	for _, customer := range customers {
		if customer.OrderID != farthestCustomer.OrderID {
			waypoints = append(waypoints, utils.Coordinate{
				Latitude:  customer.Latitude,
				Longitude: customer.Longitude,
			})
		}
	}

	fmt.Printf("[PlanDeliveryRoute] 途经点数量: %d\n", len(waypoints))
	for i, wp := range waypoints {
		fmt.Printf("[PlanDeliveryRoute] 途经点 %d: %.6f,%.6f\n", i+1, wp.Latitude, wp.Longitude)
	}

	// 目的地：最远的客户位置
	destination := utils.Coordinate{
		Latitude:  farthestCustomer.Latitude,
		Longitude: farthestCustomer.Longitude,
	}

	fmt.Printf("[PlanDeliveryRoute] 目的地坐标: %.6f,%.6f\n", destination.Latitude, destination.Longitude)
	fmt.Printf("[PlanDeliveryRoute] 开始调用高德地图路线规划API...\n")

	// 调用路线规划API
	routeResult, err := utils.PlanRoute(origin, destination, waypoints, amapKey)
	if err != nil {
		fmt.Printf("[PlanDeliveryRoute] 路线规划API调用失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "路线规划失败: " + err.Error(),
		})
		return
	}

	if !routeResult.Success {
		fmt.Printf("[PlanDeliveryRoute] 路线规划失败: %s\n", routeResult.Message)
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": routeResult.Message,
			"data":    nil,
		})
		return
	}

	fmt.Printf("[PlanDeliveryRoute] 路线规划成功 - 距离: %d米, 耗时: %d秒, 坐标点数: %d\n", 
		routeResult.Distance, routeResult.Duration, len(strings.Split(routeResult.Polyline, ";")))

	// 构建返回数据
	responseData := map[string]interface{}{
		"route": map[string]interface{}{
			"distance": routeResult.Distance,
			"duration": routeResult.Duration,
			"polyline": routeResult.Polyline,
			"steps":    routeResult.Steps,
		},
		"destination": map[string]interface{}{
			"order_id":     farthestCustomer.OrderID,
			"order_number": farthestCustomer.OrderNumber,
			"latitude":     farthestCustomer.Latitude,
			"longitude":    farthestCustomer.Longitude,
			"address":      farthestCustomer.Address,
			"name":         farthestCustomer.Name,
		},
		"waypoints": make([]map[string]interface{}, 0),
	}

	// 添加途经点信息（使用API返回的优化后的顺序，如果API返回了的话）
	// 如果API返回了优化后的途经点顺序，使用API返回的顺序；否则使用原始顺序
	if routeResult.Waypoints != nil && len(routeResult.Waypoints) > 0 {
		// 使用API返回的优化后的途经点顺序
		fmt.Printf("[PlanDeliveryRoute] 使用API返回的优化后的途经点顺序，数量: %d\n", len(routeResult.Waypoints))
		for i, wpInfo := range routeResult.Waypoints {
			// 解析途经点坐标
			coords := strings.Split(wpInfo.Location, ",")
			if len(coords) == 2 {
				var wpLng, wpLat float64
				if _, err := fmt.Sscanf(coords[0], "%f", &wpLng); err == nil {
					if _, err := fmt.Sscanf(coords[1], "%f", &wpLat); err == nil {
						// 找到对应的客户信息
						for _, customer := range customers {
							// 允许小的坐标误差（0.0001度，约11米）
							latDiff := customer.Latitude - wpLat
							lngDiff := customer.Longitude - wpLng
							if latDiff < 0 {
								latDiff = -latDiff
							}
							if lngDiff < 0 {
								lngDiff = -lngDiff
							}
							if lngDiff < 0.0001 && latDiff < 0.0001 {
								responseData["waypoints"] = append(responseData["waypoints"].([]map[string]interface{}), map[string]interface{}{
									"order_id":     customer.OrderID,
									"order_number": customer.OrderNumber,
									"latitude":     customer.Latitude,
									"longitude":    customer.Longitude,
									"address":      customer.Address,
									"name":         customer.Name,
									"index":        i + 1, // 途经点序号（优化后的顺序）
								})
								break
							}
						}
					}
				}
			}
		}
	} else {
		// 使用原始途经点顺序
		fmt.Printf("[PlanDeliveryRoute] 使用原始途经点顺序，数量: %d\n", len(waypoints))
		for i, wp := range waypoints {
			// 找到对应的客户信息
			for _, customer := range customers {
				if customer.Latitude == wp.Latitude && customer.Longitude == wp.Longitude {
					responseData["waypoints"] = append(responseData["waypoints"].([]map[string]interface{}), map[string]interface{}{
						"order_id":     customer.OrderID,
						"order_number": customer.OrderNumber,
						"latitude":     customer.Latitude,
						"longitude":    customer.Longitude,
						"address":      customer.Address,
						"name":         customer.Name,
						"index":        i + 1, // 途经点序号
					})
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "路线规划成功",
		"data":    responseData,
	})
}
