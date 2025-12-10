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

	// 获取配送员当前位置（可选参数，但强烈建议提供）
	var employeeLat, employeeLng *float64
	if latStr := c.Query("latitude"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			employeeLat = &lat
		}
	}
	if lngStr := c.Query("longitude"); lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
			employeeLng = &lng
		}
	}

	// 如果未提供位置，返回错误（要求必须提供位置才能接单）
	if employeeLat == nil || employeeLng == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "接单需要提供当前位置信息，请确保已开启定位权限并获取到位置",
		})
		return
	}

	fmt.Printf("[AcceptDeliveryOrder] 接单 - 订单ID: %d, 配送员: %s, 位置: %.6f, %.6f\n", id, employee.EmployeeCode, *employeeLat, *employeeLng)

	// 更新订单状态为待取货，并记录配送员信息
	err = model.UpdateOrderStatusWithDeliveryEmployee(id, "pending_pickup", employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "接单失败: " + err.Error()})
		return
	}

	// 记录配送流程日志：接单
	deliveryLog := &model.DeliveryLog{
		OrderID:              id,
		Action:               model.DeliveryLogActionAccepted,
		DeliveryEmployeeCode: &employee.EmployeeCode,
		ActionTime:           time.Now(),
	}
	_ = model.CreateDeliveryLog(deliveryLog) // 记录日志失败不影响主流程

	// 异步触发路线规划计算（使用配送员当前位置）
	go func() {
		err := CalculateAndUpdateRoute(employee.EmployeeCode, employeeLat, employeeLng)
		if err != nil {
			fmt.Printf("[AcceptDeliveryOrder] 异步计算路线失败: %v\n", err)
		} else {
			fmt.Printf("[AcceptDeliveryOrder] 异步计算路线成功，配送员: %s\n", employee.EmployeeCode)
		}
	}()

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
		OrderID:              id,
		Action:               model.DeliveryLogActionDeliveringStarted,
		DeliveryEmployeeCode: &employee.EmployeeCode,
		ActionTime:           time.Now(),
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
		OrderID:              id,
		DeliveryEmployeeCode: employee.EmployeeCode,
		ProductImageURL:      productImageURL,
		DoorplateImageURL:    doorplateImageURL,
		CompletedAt:          time.Now(),
	}

	// 记录配送流程日志：配送完成
	remark := "配送完成，已上传货物照片和门牌照片"
	deliveryLog := &model.DeliveryLog{
		OrderID:              id,
		Action:               model.DeliveryLogActionDeliveringCompleted,
		DeliveryEmployeeCode: &employee.EmployeeCode,
		ActionTime:           time.Now(),
		Remark:               &remark,
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

// GetPickupSuppliers 获取配送员待取货订单的供应商列表（按路线优化顺序）
func GetPickupSuppliers(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	// 获取配送员当前位置（可选参数）
	var employeeLat, employeeLng *float64
	if latStr := c.Query("latitude"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			employeeLat = &lat
		}
	}
	if lngStr := c.Query("longitude"); lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
			employeeLng = &lng
		}
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
		  AND s.latitude IS NOT NULL
		  AND s.longitude IS NOT NULL
	`

	rows, err := database.DB.Query(query, employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取供应商列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	type SupplierLocation struct {
		ID        int
		Name      string
		Contact   *string
		Phone     *string
		Email     *string
		Address   *string
		Latitude  float64
		Longitude float64
	}

	suppliers := make([]SupplierLocation, 0)
	for rows.Next() {
		var supplier SupplierLocation
		var latitude, longitude sql.NullFloat64
		var contact, phone, email, address sql.NullString
		err := rows.Scan(
			&supplier.ID,
			&supplier.Name,
			&contact,
			&phone,
			&email,
			&address,
			&latitude,
			&longitude,
		)
		if err != nil {
			continue
		}

		if !latitude.Valid || !longitude.Valid {
			continue
		}

		supplier.Latitude = latitude.Float64
		supplier.Longitude = longitude.Float64
		if contact.Valid {
			supplier.Contact = &contact.String
		}
		if phone.Valid {
			supplier.Phone = &phone.String
		}
		if email.Valid {
			supplier.Email = &email.String
		}
		if address.Valid {
			supplier.Address = &address.String
		}
		suppliers = append(suppliers, supplier)
	}

	if len(suppliers) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"data":    []interface{}{},
			"message": "暂无待取货供应商",
		})
		return
	}

	// 如果只有一家供应商，直接返回
	if len(suppliers) == 1 {
		supplier := suppliers[0]
		supplierData := map[string]interface{}{
			"id":        supplier.ID,
			"name":      supplier.Name,
			"contact":   supplier.Contact,
			"phone":     supplier.Phone,
			"email":     supplier.Email,
			"address":   supplier.Address,
			"latitude":  supplier.Latitude,
			"longitude": supplier.Longitude,
			"sequence":  1, // 排序序号
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"data":    []map[string]interface{}{supplierData},
			"message": "获取成功",
		})
		return
	}

	// 获取配送员当前位置作为起点
	var startLat, startLng float64
	if employeeLat != nil && employeeLng != nil {
		// 使用前端传递的配送员当前位置
		startLat = *employeeLat
		startLng = *employeeLng
	} else if len(suppliers) > 0 {
		// 如果没有提供位置，使用第一个供应商位置作为起点（回退方案）
		startLat = suppliers[0].Latitude
		startLng = suppliers[0].Longitude
	} else {
		// 没有供应商，无法规划路线
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"data":    []interface{}{},
			"message": "暂无待取货供应商",
		})
		return
	}

	// 构建位置列表（起点 + 供应商位置）
	locations := make([]struct {
		Name string
		Lat  float64
		Lng  float64
	}, 0, len(suppliers)+1)

	// 添加起点
	startPointName := fmt.Sprintf("start_pickup_%s", employee.EmployeeCode)
	locations = append(locations, struct {
		Name string
		Lat  float64
		Lng  float64
	}{
		Name: startPointName,
		Lat:  startLat,
		Lng:  startLng,
	})

	// 添加供应商位置
	supplierIDToIndex := make(map[int]int) // 供应商ID -> 在locations中的索引
	for i, supplier := range suppliers {
		supplierName := fmt.Sprintf("supplier_%d", supplier.ID)
		locations = append(locations, struct {
			Name string
			Lat  float64
			Lng  float64
		}{
			Name: supplierName,
			Lat:  supplier.Latitude,
			Lng:  supplier.Longitude,
		})
		supplierIDToIndex[supplier.ID] = i + 1 // +1 因为起点占索引0
	}

	// 创建优化器并添加位置
	optimizer := utils.NewDeliveryRouteOptimizer()
	for _, loc := range locations {
		optimizer.AddLocation(loc.Name, loc.Lat, loc.Lng)
	}

	// 执行路线优化（迭代300次）
	optimizedRoute, _, err := optimizer.OptimizeRoute(startPointName, 300)
	if err != nil {
		// 如果优化失败，按原始顺序返回
		fmt.Printf("[GetPickupSuppliers] 路线优化失败: %v，使用原始顺序\n", err)
		suppliersList := make([]map[string]interface{}, 0)
		for i, supplier := range suppliers {
			supplierData := map[string]interface{}{
				"id":        supplier.ID,
				"name":      supplier.Name,
				"contact":   supplier.Contact,
				"phone":     supplier.Phone,
				"email":     supplier.Email,
				"address":   supplier.Address,
				"latitude":  supplier.Latitude,
				"longitude": supplier.Longitude,
				"sequence":  i + 1,
			}
			suppliersList = append(suppliersList, supplierData)
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"data":    suppliersList,
			"message": "获取成功",
		})
		return
	}

	// 构建排序后的供应商列表（排除起点，只保留供应商）
	sortedSuppliers := make([]SupplierLocation, 0)
	sequenceMap := make(map[int]int) // 供应商ID -> 排序序号

	sequence := 1
	for _, routeIdx := range optimizedRoute {
		loc := optimizer.GetLocationByIndex(routeIdx)
		if loc == nil {
			continue
		}

		// 跳过起点
		if loc.Name == startPointName {
			continue
		}

		// 解析供应商ID
		var supplierID int
		_, err := fmt.Sscanf(loc.Name, "supplier_%d", &supplierID)
		if err != nil {
			continue
		}

		// 找到对应的供应商
		for _, supplier := range suppliers {
			if supplier.ID == supplierID {
				sortedSuppliers = append(sortedSuppliers, supplier)
				sequenceMap[supplierID] = sequence
				sequence++
				break
			}
		}
	}

	// 构建返回数据
	suppliersList := make([]map[string]interface{}, 0)
	for _, supplier := range sortedSuppliers {
		seq := sequenceMap[supplier.ID]
		supplierData := map[string]interface{}{
			"id":        supplier.ID,
			"name":      supplier.Name,
			"contact":   supplier.Contact,
			"phone":     supplier.Phone,
			"email":     supplier.Email,
			"address":   supplier.Address,
			"latitude":  supplier.Latitude,
			"longitude": supplier.Longitude,
			"sequence":  seq, // 排序序号
		}
		suppliersList = append(suppliersList, supplierData)
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
							OrderID:              orderID,
							Action:               model.DeliveryLogActionPickupCompleted,
							DeliveryEmployeeCode: &deliveryEmployeeCode.String,
							ActionTime:           time.Now(),
							Remark:               &remark,
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

// CalculateAndUpdateRoute 计算并更新配送员的路线排序
// employeeLat, employeeLng: 配送员当前位置（可选，如果为nil则使用第一个订单位置作为起点）
func CalculateAndUpdateRoute(employeeCode string, employeeLat, employeeLng *float64) error {
	// 查询该配送员所有配送中的订单（包括 pending_pickup 和 delivering）
	query := `
		SELECT o.id, o.order_number, a.latitude, a.longitude, a.address, a.name
		FROM orders o
		INNER JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE o.delivery_employee_code = ? 
		  AND o.status IN ('pending_pickup', 'delivering')
		  AND a.latitude IS NOT NULL 
		  AND a.longitude IS NOT NULL
		ORDER BY o.created_at ASC
	`

	rows, err := database.DB.Query(query, employeeCode)
	if err != nil {
		return fmt.Errorf("查询配送订单失败: %w", err)
	}
	defer rows.Close()

	type OrderLocation struct {
		OrderID   int
		Latitude  float64
		Longitude float64
	}

	var orders []OrderLocation
	var orderIDs []int
	for rows.Next() {
		var ol OrderLocation
		var latitude, longitude sql.NullFloat64
		var orderNumber, address, name sql.NullString

		err := rows.Scan(&ol.OrderID, &orderNumber, &latitude, &longitude, &address, &name)
		if err != nil {
			continue
		}

		if !latitude.Valid || !longitude.Valid {
			continue
		}

		ol.Latitude = latitude.Float64
		ol.Longitude = longitude.Float64
		orders = append(orders, ol)
		orderIDs = append(orderIDs, ol.OrderID)
	}

	if len(orders) == 0 {
		// 没有订单，清空排序记录
		_ = model.DeleteRouteOrdersByEmployee(employeeCode)
		return nil
	}

	// 如果只有一单，直接设置排序，不需要优化
	if len(orders) == 1 {
		orderSequences := []struct {
			OrderID  int
			Sequence int
			Distance *float64
		}{
			{
				OrderID:  orders[0].OrderID,
				Sequence: 1,
				Distance: nil, // 单订单没有距离概念
			},
		}
		err := model.UpdateRouteSequence(employeeCode, orderSequences)
		if err != nil {
			return fmt.Errorf("更新路线排序失败: %w", err)
		}
		return nil
	}

	// 获取配送员当前位置作为起点
	var startLat, startLng float64
	if employeeLat != nil && employeeLng != nil {
		// 使用传入的配送员当前位置
		startLat = *employeeLat
		startLng = *employeeLng
		fmt.Printf("[CalculateAndUpdateRoute] 使用配送员当前位置作为起点: %.6f, %.6f\n", startLat, startLng)
	} else if len(orders) > 0 {
		// 如果没有提供位置，使用第一个订单位置作为起点（回退方案）
		startLat = orders[0].Latitude
		startLng = orders[0].Longitude
		fmt.Printf("[CalculateAndUpdateRoute] 警告: 未提供配送员位置，使用第一个订单位置作为起点: %.6f, %.6f (订单ID: %d)\n", startLat, startLng, orders[0].OrderID)
	} else {
		// 没有订单，清空排序记录
		_ = model.DeleteRouteOrdersByEmployee(employeeCode)
		return nil
	}

	// 构建位置列表（起点 + 订单位置）
	locations := make([]struct {
		Name string
		Lat  float64
		Lng  float64
	}, 0, len(orders)+1)

	// 添加起点（使用特殊名称）- 必须第一个添加，确保索引为0
	startPointName := fmt.Sprintf("start_%s", employeeCode)
	locations = append(locations, struct {
		Name string
		Lat  float64
		Lng  float64
	}{
		Name: startPointName,
		Lat:  startLat,
		Lng:  startLng,
	})
	fmt.Printf("[CalculateAndUpdateRoute] 起点已添加: %s (%.6f, %.6f)，索引: 0\n", startPointName, startLat, startLng)

	// 添加订单位置
	for i, order := range orders {
		orderName := fmt.Sprintf("order_%d", order.OrderID)
		locations = append(locations, struct {
			Name string
			Lat  float64
			Lng  float64
		}{
			Name: orderName,
			Lat:  order.Latitude,
			Lng:  order.Longitude,
		})
		// 更新 orders 中的索引映射
		orders[i].OrderID = order.OrderID
	}

	// 创建优化器并添加位置
	optimizer := utils.NewDeliveryRouteOptimizer()
	for _, loc := range locations {
		optimizer.AddLocation(loc.Name, loc.Lat, loc.Lng)
	}

	// 执行路线优化（迭代300次）
	optimizedRoute, totalDistance, err := optimizer.OptimizeRoute(startPointName, 300)
	if err != nil {
		return fmt.Errorf("路线优化失败: %w", err)
	}
	fmt.Printf("[CalculateAndUpdateRoute] 路线优化完成，总距离: %.2f 公里，路线点数: %d\n", totalDistance, len(optimizedRoute))
	if len(optimizedRoute) > 0 {
		firstLoc := optimizer.GetLocationByIndex(optimizedRoute[0])
		if firstLoc != nil {
			fmt.Printf("[CalculateAndUpdateRoute] 优化路线第一个点: %s (%.6f, %.6f)\n", firstLoc.Name, firstLoc.Lat, firstLoc.Lng)
		}
		if len(optimizedRoute) > 1 {
			secondLoc := optimizer.GetLocationByIndex(optimizedRoute[1])
			if secondLoc != nil {
				fmt.Printf("[CalculateAndUpdateRoute] 优化路线第二个点: %s (%.6f, %.6f)\n", secondLoc.Name, secondLoc.Lat, secondLoc.Lng)
			}
		}
	}

	// 构建排序记录（排除起点，只保留订单）
	orderSequences := make([]struct {
		OrderID  int
		Sequence int
		Distance *float64
	}, 0)

	sequence := 1
	startIdx := -1
	// 先找到起点的索引
	for i, routeIdx := range optimizedRoute {
		loc := optimizer.GetLocationByIndex(routeIdx)
		if loc != nil && loc.Name == startPointName {
			startIdx = i
			fmt.Printf("[CalculateAndUpdateRoute] 起点在优化路线中的索引: %d\n", startIdx)
			break
		}
	}

	if startIdx == -1 {
		return fmt.Errorf("在优化路线中未找到起点")
	}

	// 确保起点是第一个点（NearestNeighbor算法应该保证这一点）
	if startIdx != 0 {
		fmt.Printf("[CalculateAndUpdateRoute] 警告: 起点不在第一个位置 (索引: %d)，这可能导致路线计算错误\n", startIdx)
	}

	for i, routeIdx := range optimizedRoute {
		loc := optimizer.GetLocationByIndex(routeIdx)
		if loc == nil {
			continue
		}

		// 跳过起点
		if loc.Name == startPointName {
			continue
		}

		// 解析订单ID
		var orderID int
		_, err := fmt.Sscanf(loc.Name, "order_%d", &orderID)
		if err != nil {
			continue
		}

		// 计算距离（从前一个点到当前点的距离）
		// 如果前一个点是起点，则计算从起点到当前点的距离
		// 否则计算从前一个订单到当前订单的距离
		var distance *float64
		if i > 0 {
			prevIdx := optimizedRoute[i-1]
			prevLoc := optimizer.GetLocationByIndex(prevIdx)
			// 如果前一个点是起点，计算从起点到当前点的距离
			if prevLoc != nil && prevLoc.Name == startPointName {
				dist := optimizer.CalculateDistance(prevIdx, routeIdx)
				distance = &dist
			} else {
				// 否则计算从前一个订单到当前订单的距离
				dist := optimizer.CalculateDistance(prevIdx, routeIdx)
				distance = &dist
			}
		} else if i == startIdx+1 {
			// 如果当前点是起点后的第一个点，计算从起点到当前点的距离
			dist := optimizer.CalculateDistance(optimizedRoute[startIdx], routeIdx)
			distance = &dist
		}

		orderSequences = append(orderSequences, struct {
			OrderID  int
			Sequence int
			Distance *float64
		}{
			OrderID:  orderID,
			Sequence: sequence,
			Distance: distance,
		})
		sequence++
	}

	// 更新数据库中的排序
	err = model.UpdateRouteSequence(employeeCode, orderSequences)
	if err != nil {
		return fmt.Errorf("更新路线排序失败: %w", err)
	}

	return nil
}

// CalculateRoute 手动触发路线规划计算（API接口）
func CalculateRoute(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	// 获取配送员当前位置（可选参数）
	var employeeLat, employeeLng *float64
	if latStr := c.Query("latitude"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			employeeLat = &lat
		}
	}
	if lngStr := c.Query("longitude"); lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
			employeeLng = &lng
		}
	}

	// 异步计算路线
	go func() {
		err := CalculateAndUpdateRoute(employee.EmployeeCode, employeeLat, employeeLng)
		if err != nil {
			fmt.Printf("[CalculateRoute] 路线规划计算失败: %v\n", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "路线规划计算已开始",
	})
}

// GetRouteOrders 获取排序后的订单列表（用于路线规划页面）
func GetRouteOrders(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	// 获取排序记录
	routeOrders, err := model.GetRouteOrdersByEmployee(employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取路线排序失败: " + err.Error(),
		})
		return
	}

	// 如果没有排序记录，返回空列表
	if len(routeOrders) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"list":  []interface{}{},
				"total": 0,
			},
			"message": "暂无排序记录",
		})
		return
	}

	// 获取订单详情
	orderList := make([]map[string]interface{}, 0)
	for _, ro := range routeOrders {
		order, err := model.GetOrderByID(ro.OrderID)
		if err != nil || order == nil {
			continue
		}

		// 获取订单地址
		address, err := model.GetAddressByID(order.AddressID)
		if err != nil || address == nil {
			continue
		}

		// 获取订单商品数量
		itemCount, _ := model.GetOrderItemCountByOrderID(order.ID)

		orderData := map[string]interface{}{
			"id":                  order.ID,
			"order_number":        order.OrderNumber,
			"status":              order.Status,
			"route_sequence":      ro.RouteSequence,
			"calculated_distance": ro.CalculatedDistance,
			"address":             address.Address,
			"name":                address.Name,
			"contact":             address.Contact, // 联系人
			"phone":               address.Phone,   // 联系电话
			"latitude":            address.Latitude,
			"longitude":           address.Longitude,
			"total_amount":        order.TotalAmount,
			"is_urgent":           order.IsUrgent,
			"item_count":          itemCount, // 商品件数
			"created_at":          order.CreatedAt,
		}

		orderList = append(orderList, orderData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  orderList,
			"total": len(orderList),
		},
		"message": "获取成功",
	})
}
