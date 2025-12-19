package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/model"
	"go_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetSalesCustomers 获取我的客户列表（销售员）
func GetSalesCustomers(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	// 验证是否是销售员
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	keyword := c.Query("keyword")

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 获取客户列表
	customers, err := model.GetCustomersByEmployeeCode(employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取客户列表失败: " + err.Error()})
		return
	}

	// 关键词搜索
	if keyword != "" {
		filtered := make([]map[string]interface{}, 0)
		for _, customer := range customers {
			name := ""
			phone := ""
			userCode := ""
			if n, ok := customer["name"].(string); ok {
				name = n
			}
			if p, ok := customer["phone"].(string); ok {
				phone = p
			}
			if uc, ok := customer["user_code"].(string); ok {
				userCode = uc
			}

			if contains(name, keyword) || contains(phone, keyword) || contains(userCode, keyword) {
				filtered = append(filtered, customer)
			}
		}
		customers = filtered
	}

	// 分页
	total := len(customers)
	start := (pageNum - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var pageData []map[string]interface{}
	if start < end {
		pageData = customers[start:end]
	} else {
		pageData = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  pageData,
			"total": total,
		},
		"message": "获取成功",
	})
}

// GetSalesCustomerByCode 销售员根据用户编号查询自己的客户（含资料与地址）
func GetSalesCustomerByCode(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	userCode := strings.TrimSpace(c.Query("userCode"))
	if userCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供用户编号"})
		return
	}

	user, err := model.GetMiniAppUserByUserCode(userCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询用户失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	// 业务逻辑：
	// 1. 如果客户是新客且资料未完善，任何销售员都可以查询
	// 2. 如果客户已完善资料且不属于当前销售员，才返回错误
	isNewCustomer := !user.ProfileCompleted || user.SalesCode == ""
	if !isNewCustomer && user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "该客户不属于当前销售员"})
		return
	}

	// 获取默认地址和所有地址
	defaultAddress, _ := model.GetDefaultAddressByUserID(user.ID)
	allAddresses, _ := model.GetAddressesByUserID(user.ID)

	responseData := map[string]interface{}{
		"id":                user.ID,
		"unique_id":         user.UniqueID,
		"user_code":         user.UserCode,
		"name":              user.Name,
		"avatar":            user.Avatar,
		"phone":             user.Phone,
		"sales_code":        user.SalesCode,
		"store_type":        user.StoreType,
		"user_type":         user.UserType,
		"profile_completed": user.ProfileCompleted,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
		"default_address":   defaultAddress,
		"addresses":         allAddresses,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    responseData,
	})
}

// GetMyPendingOrders 获取当前销售员名下客户的待配送订单列表（分页）
func GetMyPendingOrders(c *gin.Context) {
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

	orders, total, err := model.GetPendingOrdersBySalesCode(employee.EmployeeCode, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取待配送订单失败: " + err.Error()})
		return
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

// GetSalesOrders 获取当前销售员名下客户的订单列表（分页，支持状态和搜索）
func GetSalesOrders(c *gin.Context) {
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
	status := strings.TrimSpace(c.Query("status"))
	keyword := strings.TrimSpace(c.Query("keyword"))

	orders, total, err := model.GetOrdersBySalesCode(employee.EmployeeCode, pageNum, pageSize, status, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":     orders,
			"total":    total,
			"pageNum":  pageNum,
			"pageSize": pageSize,
		},
		"message": "获取成功",
	})
}

// GetSalesOrderDetail 获取订单详情（销售员查看）
// 返回：订单基本信息、订单商品明细、收货地址、客户信息
func GetSalesOrderDetail(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

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
	// #region agent log - 记录订单锁定状态
	log.Printf("[GetSalesOrderDetail] 查询订单详情: 订单ID=%d, is_locked=%v, locked_by=%v, locked_at=%v", id, order.IsLocked, order.LockedBy, order.LockedAt)
	// #endregion

	// 获取客户信息，并校验是否属于当前销售员
	user, err := model.GetMiniAppUserByID(order.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取客户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权查看该订单"})
		return
	}

	// 获取订单明细
	items, err := model.GetOrderItemsByOrderID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单明细失败: " + err.Error()})
		return
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

	// 获取配送员信息（如果有）
	deliveryEmployeeData := map[string]interface{}{}
	if order.DeliveryEmployeeCode != nil && *order.DeliveryEmployeeCode != "" {
		deliveryEmployee, err := model.GetEmployeeByEmployeeCode(*order.DeliveryEmployeeCode)
		if err == nil && deliveryEmployee != nil {
			deliveryEmployeeData = map[string]interface{}{
				"id":            deliveryEmployee.ID,
				"employee_code": deliveryEmployee.EmployeeCode,
				"name":          deliveryEmployee.Name,
				"phone":         deliveryEmployee.Phone,
			}
		}
	}

	// 客户信息摘要
	userData := map[string]interface{}{
		"id":         user.ID,
		"user_code":  user.UserCode,
		"name":       user.Name,
		"phone":      user.Phone,
		"store_type": user.StoreType,
		"user_type":  user.UserType,
	}

	// 获取配送费计算结果（用于提取配送员费用）
	var deliveryFeeCalculation map[string]interface{}
	var riderPayableFee float64
	var deliveryFeeResult *model.DeliveryFeeCalculationResult

	var deliveryFeeCalcJSON sql.NullString
	err = database.DB.QueryRow(`
		SELECT delivery_fee_calculation
		FROM orders WHERE id = ?
	`, id).Scan(&deliveryFeeCalcJSON)

	if err == nil && deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
		var result model.DeliveryFeeCalculationResult
		if json.Unmarshal([]byte(deliveryFeeCalcJSON.String), &result) == nil {
			deliveryFeeResult = &result
			riderPayableFee = result.RiderPayableFee
			deliveryFeeCalculation = map[string]interface{}{
				"rider_payable_fee":   result.RiderPayableFee,
				"base_fee":            result.BaseFee,
				"isolated_fee":        result.IsolatedFee,
				"item_fee":            result.ItemFee,
				"urgent_fee":          result.UrgentFee,
				"weather_fee":         result.WeatherFee,
				"profit_share":        result.ProfitShare,
				"total_platform_cost": result.TotalPlatformCost,
			}
		}
	}

	// 获取订单利润
	var orderProfit float64
	var orderProfitVal sql.NullFloat64
	err = database.DB.QueryRow(`
		SELECT order_profit
		FROM orders WHERE id = ?
	`, id).Scan(&orderProfitVal)
	if err == nil && orderProfitVal.Valid {
		orderProfit = orderProfitVal.Float64
	}

	// 添加销售分成信息
	var salesCommissionPreview map[string]interface{}
	var salesCommission map[string]interface{}

	if user.SalesCode != "" {
		// 计算预览分成（所有订单都显示）
		previewCommission := calculateSalesCommissionPreviewForSales(order, deliveryFeeResult, orderProfit, user.SalesCode)
		if previewCommission != nil {
			salesCommissionPreview = previewCommission
		}

		// 已收款订单：从数据库查询已计入的分成
		if order.Status == "paid" {
			commissions, err := model.GetSalesCommissionsByOrderIDs([]int{id})
			if err == nil && len(commissions) > 0 && commissions[0] != nil {
				commission := commissions[0]
				salesCommission = map[string]interface{}{
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
			}
		}
	}

	result := gin.H{
		"order":             order,
		"order_items":       items,
		"address":           addressData,
		"user":              userData,
		"delivery_employee": deliveryEmployeeData,
	}

	// 添加配送员费用信息
	if deliveryFeeCalculation != nil {
		result["delivery_fee_calculation"] = deliveryFeeCalculation
		result["rider_payable_fee"] = riderPayableFee
	}

	// 添加销售分成信息
	if salesCommissionPreview != nil {
		result["sales_commission_preview"] = salesCommissionPreview
	}
	if salesCommission != nil {
		result["sales_commission"] = salesCommission
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    result,
		"message": "获取成功",
	})
}

// LockOrderForEdit 锁定订单用于修改（防止被接单）
func LockOrderForEdit(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	// 获取订单并校验权限
	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败: " + err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	// 校验客户归属
	user, err := model.GetMiniAppUserByID(order.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权修改该订单"})
		return
	}

	// 锁定订单
	err = model.LockOrder(id, employee.EmployeeCode)
	if err != nil {
		log.Printf("[LockOrderForEdit] 锁定订单失败: 订单ID=%d, 员工=%s, 错误=%v", id, employee.EmployeeCode, err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 验证锁定是否成功
	lockedOrder, err := model.GetOrderByID(id)
	if err != nil || lockedOrder == nil {
		log.Printf("[LockOrderForEdit] 验证锁定状态失败: 订单ID=%d, 错误=%v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证锁定状态失败"})
		return
	}
	if !lockedOrder.IsLocked {
		log.Printf("[LockOrderForEdit] 订单锁定失败: 订单ID=%d, is_locked=%v", id, lockedOrder.IsLocked)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "订单锁定失败，请重试"})
		return
	}

	log.Printf("[LockOrderForEdit] 订单锁定成功: 订单ID=%d, 锁定者=%s, is_locked=%v", id, employee.EmployeeCode, lockedOrder.IsLocked)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "订单已锁定，可以开始修改",
	})
}

// UnlockOrderAfterEdit 修改完成后解锁订单
func UnlockOrderAfterEdit(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	// 解锁订单
	err = model.UnlockOrder(id, employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "订单已解锁",
	})
}

// SyncOrderItemsToPurchaseList 将订单商品同步到采购单（用于修改订单）
func SyncOrderItemsToPurchaseList(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	// 获取订单并校验权限
	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败: " + err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	// 验证客户归属
	user, err := model.GetMiniAppUserByID(order.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问该订单"})
		return
	}

	// 检查订单是否已锁定（必须已锁定才能进入修改页面）
	if !order.IsLocked {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单未锁定，请先锁定订单"})
		return
	}
	// 检查订单是否被当前员工锁定
	if order.LockedBy == nil || *order.LockedBy != employee.EmployeeCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单已被其他员工锁定，无法修改"})
		return
	}

	// 获取订单商品
	orderItems, err := model.GetOrderItemsByOrderID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单商品失败: " + err.Error()})
		return
	}

	if len(orderItems) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "订单没有商品，无需同步",
			"data":    gin.H{"items": []interface{}{}, "summary": nil},
		})
		return
	}

	// 修改订单时，不直接修改用户的采购单，而是创建一个临时的采购单视图
	// 将订单商品添加到采购单（用于修改订单时显示和编辑）
	// 在同步订单商品到采购单之前，先备份用户的原始采购单（用于修改订单完成后恢复）
	// 注意：即使采购单为空（比如用户之前已经下过单），也要备份，这样恢复时才能正确恢复空状态
	userPurchaseListBackup, err := model.BackupPurchaseList(order.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "备份采购单失败: " + err.Error()})
		return
	}
	log.Printf("[SyncOrderItemsToPurchaseList] 备份采购单，备份商品数量: %d", len(userPurchaseListBackup))

	// 先清空该用户的采购单（用于修改订单）
	// 注意：这里清空的是用户当前的采购单，可能是空的（如果用户之前已经下过单），也可能包含用户在小程序中添加的商品
	_, err = database.DB.Exec("DELETE FROM purchase_list_items WHERE user_id = ?", order.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "清空采购单失败: " + err.Error()})
		return
	}

	// 将订单商品添加到采购单
	for _, orderItem := range orderItems {
		// 获取商品信息（包括规格信息）
		product, err := model.GetProductByID(orderItem.ProductID)
		if err != nil || product == nil {
			log.Printf("[SyncOrderItemsToPurchaseList] 获取商品失败: 商品ID=%d, 错误=%v", orderItem.ProductID, err)
			continue
		}

		// 查找对应的规格
		var specSnapshot model.PurchaseSpecSnapshot
		found := false
		for _, spec := range product.Specs {
			if spec.Name == orderItem.SpecName {
				specSnapshot = model.PurchaseSpecSnapshot{
					Name:           spec.Name,
					Description:    spec.Description,
					Cost:           spec.Cost,
					WholesalePrice: spec.WholesalePrice,
					RetailPrice:    spec.RetailPrice,
				}
				found = true
				break
			}
		}

		// 如果找不到规格，使用订单中的价格信息构造快照
		if !found {
			specSnapshot = model.PurchaseSpecSnapshot{
				Name:           orderItem.SpecName,
				Description:    "",
				Cost:           0,
				WholesalePrice: orderItem.UnitPrice,
				RetailPrice:    orderItem.UnitPrice,
			}
		}

		// 使用订单中的图片，如果没有则使用商品首图
		productImage := orderItem.Image
		if productImage == "" && len(product.Images) > 0 {
			productImage = product.Images[0]
		}

		item := &model.PurchaseListItem{
			UserID:       order.UserID,
			ProductID:    orderItem.ProductID,
			ProductName:  orderItem.ProductName,
			ProductImage: productImage,
			SpecName:     orderItem.SpecName,
			SpecSnapshot: specSnapshot,
			Quantity:     orderItem.Quantity,
			IsSpecial:    product.IsSpecial,
		}

		if _, err := model.AddOrUpdatePurchaseListItem(item); err != nil {
			log.Printf("[SyncOrderItemsToPurchaseList] 添加采购单项失败: 商品ID=%d, 错误=%v", orderItem.ProductID, err)
			continue
		}

	}

	// 返回最新的采购单（包含配送费汇总）
	items, err := model.GetPurchaseListItemsByUserID(order.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	// 计算配送费汇总
	userType := user.UserType
	if userType == "" {
		userType = "retail"
	}
	summary, err := model.CalculateDeliveryFee(items, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	// 获取加急费用（从系统设置）
	urgentFeeStr, _ := model.GetSystemSetting("order_urgent_fee")
	urgentFee := 0.0
	if urgentFeeStr != "" {
		if fee, parseErr := strconv.ParseFloat(urgentFeeStr, 64); parseErr == nil && fee > 0 {
			urgentFee = fee
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "订单商品已同步到采购单",
		"data": gin.H{
			"items":      items,
			"summary":    summary,
			"backup":     userPurchaseListBackup, // 返回备份，供修改订单完成后恢复使用（即使为空也要返回）
			"urgent_fee": urgentFee,              // 返回加急费用供前端显示
		},
	})
	log.Printf("[SyncOrderItemsToPurchaseList] 返回备份数据，备份商品数量: %d", len(userPurchaseListBackup))
}

// UpdateOrderForCustomer 修改订单（销售员）
func UpdateOrderForCustomer(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	var req struct {
		AddressID           int                      `json:"address_id" binding:"required"`
		ItemIDs             []int                    `json:"item_ids"`              // 采购单项ID列表，为空则使用全部
		Remark              string                   `json:"remark"`                // 订单备注
		CouponID            int                      `json:"coupon_id"`             // 用户优惠券ID（user_coupon_id），可选
		OutOfStockStrategy  string                   `json:"out_of_stock_strategy"` // 缺货处理：cancel_item / ship_available / contact_me
		TrustReceipt        bool                     `json:"trust_receipt"`         // 信任签收
		HidePrice           bool                     `json:"hide_price"`            // 是否隐藏价格
		RequirePhoneContact bool                     `json:"require_phone_contact"` // 配送时是否电话联系
		IsUrgent            bool                     `json:"is_urgent"`             // 是否加急订单
		PurchaseListBackup  []model.PurchaseListItem `json:"purchase_list_backup"`  // 用户原来的采购单备份（从 SyncOrderItemsToPurchaseList 获取）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 获取订单并校验权限和锁定状态
	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败: " + err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	// 检查订单状态是否允许修改
	if order.Status != "pending_delivery" && order.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单状态不允许修改（当前状态：" + order.Status + "）"})
		return
	}

	// 检查订单是否已锁定（必须已锁定才能修改）
	if !order.IsLocked {
		log.Printf("[UpdateOrderForCustomer] 订单未锁定，拒绝修改: 订单ID=%d, 员工=%s", id, employee.EmployeeCode)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单未锁定，请先锁定订单"})
		return
	}
	// 检查订单是否被当前员工锁定
	if order.LockedBy == nil || *order.LockedBy != employee.EmployeeCode {
		log.Printf("[UpdateOrderForCustomer] 订单被其他员工锁定，拒绝修改: 订单ID=%d, 当前员工=%s, 锁定者=%v", id, employee.EmployeeCode, order.LockedBy)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单已被其他员工锁定，无法修改"})
		return
	}
	log.Printf("[UpdateOrderForCustomer] 订单锁定检查通过: 订单ID=%d, 锁定者=%s", id, employee.EmployeeCode)

	// 验证客户归属
	user, err := model.GetMiniAppUserByID(order.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权修改该订单"})
		return
	}

	// 验证地址归属
	address, err := model.GetAddressByID(req.AddressID)
	if err != nil || address == nil || address.UserID != order.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "收货地址无效"})
		return
	}

	// 获取采购单
	items, err := model.GetPurchaseListItemsByUserID(order.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	// 筛选指定的 item_ids
	if len(req.ItemIDs) > 0 {
		filter := make(map[int]struct{}, len(req.ItemIDs))
		for _, itemID := range req.ItemIDs {
			if itemID > 0 {
				filter[itemID] = struct{}{}
			}
		}
		filteredItems := make([]model.PurchaseListItem, 0, len(filter))
		for _, item := range items {
			if _, ok := filter[item.ID]; ok {
				filteredItems = append(filteredItems, item)
			}
		}
		items = filteredItems
	}

	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单为空，无法修改订单"})
		return
	}

	// 获取用户类型
	userType := user.UserType
	if userType == "" || userType == "unknown" {
		userType = "retail"
	}

	// 计算配送费
	summary, err := model.CalculateDeliveryFee(items, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	// 计算订单金额和分类信息用于优惠券筛选
	orderAmount := 0.0
	categoryIDSet := make(map[int]struct{})
	productIDs := make([]int, 0, len(items))
	for _, item := range items {
		var price float64
		if userType == "wholesale" {
			price = item.SpecSnapshot.WholesalePrice
			if price <= 0 {
				price = item.SpecSnapshot.RetailPrice
			}
		} else {
			price = item.SpecSnapshot.RetailPrice
			if price <= 0 {
				price = item.SpecSnapshot.WholesalePrice
			}
		}
		if price <= 0 {
			price = item.SpecSnapshot.Cost
		}
		if price < 0 {
			price = 0
		}
		orderAmount += price * float64(item.Quantity)
		productIDs = append(productIDs, item.ProductID)
	}

	categoryInfo, err := model.FetchProductCategoryInfo(productIDs)
	if err == nil {
		for _, info := range categoryInfo {
			if info.CategoryID > 0 {
				categoryIDSet[info.CategoryID] = struct{}{}
			}
			if info.ParentID > 0 {
				categoryIDSet[info.ParentID] = struct{}{}
			}
		}
	}
	categoryIDs := make([]int, 0, len(categoryIDSet))
	for id := range categoryIDSet {
		categoryIDs = append(categoryIDs, id)
	}

	// 获取可用优惠券并计算折扣
	couponDiscount := 0.0
	var selectedCoupon *model.AvailableCouponInfo
	if req.CouponID > 0 {
		availableCoupons, err := model.GetAvailableCouponsForPurchaseList(
			order.UserID,
			orderAmount,
			categoryIDs,
			summary.DeliveryFee,
			summary.IsFreeShipping,
		)
		if err == nil {
			for i := range availableCoupons {
				if availableCoupons[i].UserCouponID == req.CouponID && availableCoupons[i].IsAvailable {
					selectedCoupon = &availableCoupons[i]
					if selectedCoupon.Type == "delivery_fee" {
						couponDiscount = summary.DeliveryFee
					} else if selectedCoupon.Type == "amount" {
						couponDiscount = selectedCoupon.DiscountValue
					}
					break
				}
			}
		}
	}

	// 处理缺货策略
	outOfStockStrategy := req.OutOfStockStrategy
	if outOfStockStrategy == "" {
		outOfStockStrategy = "contact_me"
	}

	// 获取加急费用
	urgentFee := 0.0
	if req.IsUrgent {
		urgentFeeStr, err := model.GetSystemSetting("order_urgent_fee")
		if err == nil && urgentFeeStr != "" {
			if fee, parseErr := strconv.ParseFloat(urgentFeeStr, 64); parseErr == nil && fee > 0 {
				urgentFee = fee
			}
		}
	}

	// 计算新的订单金额
	goodsAmount := summary.TotalAmount
	deliveryFee := summary.DeliveryFee
	if summary.IsFreeShipping {
		deliveryFee = 0
	}
	if !req.IsUrgent {
		urgentFee = 0
	}
	totalAmount := goodsAmount + deliveryFee + urgentFee - couponDiscount
	if totalAmount < 0 {
		totalAmount = 0
	}

	// 开始事务更新订单
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "开始事务失败: " + err.Error()})
		return
	}
	defer tx.Rollback()

	// 更新订单主表
	_, err = tx.Exec(`
		UPDATE orders SET
			address_id = ?,
			goods_amount = ?,
			delivery_fee = ?,
			coupon_discount = ?,
			is_urgent = ?,
			urgent_fee = ?,
			total_amount = ?,
			remark = ?,
			out_of_stock_strategy = ?,
			trust_receipt = ?,
			hide_price = ?,
			require_phone_contact = ?,
			updated_at = NOW()
		WHERE id = ? AND is_locked = 1 AND locked_by = ?
	`, req.AddressID, goodsAmount, deliveryFee, couponDiscount, boolToTinyInt(req.IsUrgent), urgentFee, totalAmount,
		req.Remark, outOfStockStrategy, boolToTinyInt(req.TrustReceipt), boolToTinyInt(req.HidePrice), boolToTinyInt(req.RequirePhoneContact),
		id, employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新订单失败: " + err.Error()})
		return
	}

	// 删除旧的订单商品
	_, err = tx.Exec("DELETE FROM order_items WHERE order_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除旧订单商品失败: " + err.Error()})
		return
	}

	// 插入新的订单商品
	itemStmt, err := tx.Prepare(`
		INSERT INTO order_items (
			order_id, product_id, product_name, spec_name, quantity, unit_price, subtotal, image
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "准备插入订单商品失败: " + err.Error()})
		return
	}
	defer itemStmt.Close()

	for _, it := range items {
		var price float64
		if userType == "wholesale" {
			price = it.SpecSnapshot.WholesalePrice
			if price <= 0 {
				price = it.SpecSnapshot.RetailPrice
			}
		} else {
			price = it.SpecSnapshot.RetailPrice
			if price <= 0 {
				price = it.SpecSnapshot.WholesalePrice
			}
		}
		if price <= 0 {
			price = it.SpecSnapshot.Cost
		}
		if price < 0 {
			price = 0
		}
		subtotal := price * float64(it.Quantity)

		_, err = itemStmt.Exec(
			id,
			it.ProductID,
			it.ProductName,
			it.SpecName,
			it.Quantity,
			price,
			subtotal,
			it.ProductImage,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "插入订单商品失败: " + err.Error()})
			return
		}
	}

	// 解锁订单
	_, err = tx.Exec(`
		UPDATE orders SET is_locked = 0, locked_by = NULL, locked_at = NULL 
		WHERE id = ? AND locked_by = ?
	`, id, employee.EmployeeCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "解锁订单失败: " + err.Error()})
		return
	}

	// 如果使用了新的优惠券，标记为已使用
	// 使用 user_coupon_id 精确更新，避免用户有多张相同优惠券时误更新
	if selectedCoupon != nil && req.CouponID > 0 {
		if err := model.UseCouponByUserCouponID(req.CouponID, id); err != nil {
			// 记录错误但不影响订单修改
			log.Printf("[UpdateOrderForCustomer] 标记优惠券为已使用失败 (userCouponID=%d, orderID=%d): %v", req.CouponID, id, err)
		} else {
			log.Printf("[UpdateOrderForCustomer] 成功标记优惠券为已使用 (userCouponID=%d, orderID=%d)", req.CouponID, id)
		}
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交事务失败: " + err.Error()})
		return
	}

	// 恢复用户原来的采购单（无条件恢复备份，不管销售员做了什么操作）
	// 备份是用户进入修改订单页面时的原始采购单，不包含销售员后续的任何操作
	// 无论销售员增加了商品、减少了商品、修改了数量，都恢复到备份的状态
	if len(req.PurchaseListBackup) > 0 {
		// 直接恢复备份，不需要任何过滤
		// 因为备份就是用户进入修改订单页面时的原始状态
		log.Printf("[UpdateOrderForCustomer] 开始恢复采购单，备份商品数量: %d", len(req.PurchaseListBackup))
		if err := model.RestorePurchaseList(order.UserID, req.PurchaseListBackup); err != nil {
			log.Printf("[UpdateOrderForCustomer] 恢复采购单失败: %v", err)
		} else {
			log.Printf("[UpdateOrderForCustomer] 成功恢复采购单，恢复商品数量: %d", len(req.PurchaseListBackup))
		}
	} else {
		// 如果没有提供备份，说明前端没有保存备份，这是不正确的行为
		// 但为了兼容，直接清空采购单（因为修改订单时，采购单中只有订单商品）
		log.Printf("[UpdateOrderForCustomer] 警告：前端未传入备份数据，直接清空采购单（可能不准确）")
		_, err = database.DB.Exec("DELETE FROM purchase_list_items WHERE user_id = ?", order.UserID)
		if err != nil {
			log.Printf("[UpdateOrderForCustomer] 清空采购单失败: %v", err)
		}
	}

	// 异步重新计算配送费和利润
	go func() {
		_ = model.UpdateOrderDeliveryInfo(id)
		_ = model.CalculateAndStoreOrderProfit(id)
	}()

	// 获取更新后的订单信息
	updatedOrder, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取更新后的订单失败: " + err.Error()})
		return
	}

	orderItems, err := model.GetOrderItemsByOrderID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单商品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"order":       updatedOrder,
			"order_items": orderItems,
		},
		"message": "修改订单成功",
	})
}

// CancelSalesOrder 销售员取消订单
func CancelSalesOrder(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	// 获取订单并校验权限
	order, err := model.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败: " + err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	// 校验客户归属
	user, err := model.GetMiniAppUserByID(order.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权取消该订单"})
		return
	}

	// 验证订单状态是否可以取消
	if order.Status != "pending_delivery" && order.Status != "pending" && order.Status != "pending_pickup" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单状态不允许取消（只能取消待配送或待取货状态的订单）",
		})
		return
	}

	// 更新订单状态为已取消
	err = model.UpdateOrderStatus(id, "cancelled")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "取消订单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "订单已取消",
	})
}

// UpdateSalesCustomerProfile 销售员更新自己名下客户的基础资料
type updateSalesCustomerProfileRequest struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	StoreType string `json:"storeType"`
	UserType  string `json:"userType"`
}

func UpdateSalesCustomerProfile(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户ID格式错误"})
		return
	}

	// 校验客户权限
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取客户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}

	// 业务逻辑：
	// 1. 如果客户是新客且资料未完善，任何销售员都可以完善资料
	// 2. 如果客户已完善资料且不属于当前销售员，才返回错误
	isNewCustomer := !user.ProfileCompleted || user.SalesCode == ""
	if !isNewCustomer && user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权修改该客户资料"})
		return
	}

	var req updateSalesCustomerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	updateData := map[string]interface{}{}
	if req.Name != "" {
		updateData["name"] = strings.TrimSpace(req.Name)
	}
	if req.Phone != "" {
		updateData["phone"] = strings.TrimSpace(req.Phone)
	}
	if req.StoreType != "" {
		updateData["storeType"] = strings.TrimSpace(req.StoreType)
	}
	// 用户类型 retail / wholesale
	if req.UserType != "" {
		ut := strings.ToLower(strings.TrimSpace(req.UserType))
		if ut == "retail" || ut == "wholesale" {
			updateData["userType"] = ut
		}
	}
	// 员工完善资料时，标记为已完善
	updateData["profileCompleted"] = true
	// 如果客户是新客（没有绑定销售员），则绑定当前销售员
	if user.SalesCode == "" {
		updateData["salesCode"] = employee.EmployeeCode
	}

	if err := model.UpdateMiniAppUserByAdmin(userID, updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新客户资料失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "资料更新成功",
	})
}

// 销售员维护客户地址
type upsertSalesAddressRequest struct {
	Name      string   `json:"name"`
	Contact   string   `json:"contact"`
	Phone     string   `json:"phone"`
	Address   string   `json:"address"`
	Avatar    string   `json:"avatar"`
	StoreType string   `json:"storeType"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	IsDefault bool     `json:"isDefault"`
}

// CreateSalesCustomerAddress 为客户新增地址
func CreateSalesCustomerAddress(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户ID格式错误"})
		return
	}

	// 校验客户权限
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取客户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}

	// 业务逻辑：
	// 1. 如果客户是新客且资料未完善，任何销售员都可以为其新增地址
	// 2. 如果客户已完善资料且不属于当前销售员，才返回错误
	isNewCustomer := !user.ProfileCompleted || user.SalesCode == ""
	if !isNewCustomer && user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权为该客户新增地址"})
		return
	}

	var req upsertSalesAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	addressData := map[string]interface{}{
		"name":       strings.TrimSpace(req.Name),
		"contact":    strings.TrimSpace(req.Contact),
		"phone":      strings.TrimSpace(req.Phone),
		"address":    strings.TrimSpace(req.Address),
		"avatar":     strings.TrimSpace(req.Avatar),
		"store_type": strings.TrimSpace(req.StoreType),
		"is_default": req.IsDefault,
	}
	if req.Latitude != nil {
		addressData["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		addressData["longitude"] = *req.Longitude
	}

	newAddr, err := model.CreateAddress(userID, addressData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "新增地址失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "地址新增成功",
		"data":    newAddr,
	})
}

// UpdateSalesCustomerAddress 更新客户地址
func UpdateSalesCustomerAddress(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	addrID, err := strconv.Atoi(idStr)
	if err != nil || addrID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址ID格式错误"})
		return
	}

	// 先查地址，拿到 userID 并校验客户归属
	addr, err := model.GetAddressByID(addrID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取地址信息失败: " + err.Error()})
		return
	}
	if addr == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "地址不存在"})
		return
	}

	user, err := model.GetMiniAppUserByID(addr.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取客户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}

	// 业务逻辑：
	// 1. 如果客户是新客且资料未完善，任何销售员都可以修改其地址
	// 2. 如果客户已完善资料且不属于当前销售员，才返回错误
	isNewCustomer := !user.ProfileCompleted || user.SalesCode == ""
	if !isNewCustomer && user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权修改该客户地址"})
		return
	}

	var req upsertSalesAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	addressData := map[string]interface{}{}
	if req.Name != "" {
		addressData["name"] = strings.TrimSpace(req.Name)
	}
	if req.Contact != "" {
		addressData["contact"] = strings.TrimSpace(req.Contact)
	}
	if req.Phone != "" {
		addressData["phone"] = strings.TrimSpace(req.Phone)
	}
	if req.Address != "" {
		addressData["address"] = strings.TrimSpace(req.Address)
	}
	if req.Avatar != "" {
		addressData["avatar"] = strings.TrimSpace(req.Avatar)
	}
	if req.StoreType != "" {
		addressData["store_type"] = strings.TrimSpace(req.StoreType)
	}
	if req.Latitude != nil {
		addressData["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		addressData["longitude"] = *req.Longitude
	}
	addressData["is_default"] = req.IsDefault

	if err := model.UpdateAddress(addrID, addr.UserID, addressData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新地址失败: " + err.Error()})
		return
	}

	updatedAddr, err := model.GetAddressByID(addrID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取更新后的地址失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "地址更新成功",
		"data":    updatedAddr,
	})
}

// UploadAddressAvatarByEmployee 员工端上传门头照片，返回图片 URL
func UploadAddressAvatarByEmployee(c *gin.Context) {
	// 通过中间件校验员工身份
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	_ = employee // 目前不需要具体信息，只要通过鉴权即可

	// 检查是否有文件上传
	file, headers, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片: " + err.Error()})
		return
	}
	defer file.Close()

	if headers.Size > 15*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过15MB"})
		return
	}

	// 上传到 MinIO（与小程序地址头像共用 bucket 前缀）
	fileURL, err := utils.UploadFile("mini-address-avatar", c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "图片上传成功",
		"data": gin.H{
			"avatar":   fileURL,
			"imageUrl": fileURL,
		},
	})
}

// GetSalesCustomerDetail 获取客户详情（销售员）
func GetSalesCustomerDetail(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户ID格式错误"})
		return
	}

	// 获取用户信息
	user, err := model.GetMiniAppUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取客户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}

	// 验证客户是否属于当前销售员
	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问此客户信息"})
		return
	}

	// 获取客户地址列表
	addresses, _ := model.GetAddressesByUserID(id)

	// 地址数量
	addressCount, _ := model.CountAddressesByUserID(id)

	// 订单汇总信息（总金额 & 数量）
	totalAmount, orderCount, _ := model.GetOrderSummaryByUserID(id)

	// 最近三笔订单
	recentOrders, err := model.GetRecentOrdersByUserID(id, 3)
	if err != nil {
		log.Printf("[GetSalesCustomerDetail] 获取最近订单失败: userID=%d, error=%v", id, err)
		recentOrders = []map[string]interface{}{} // 如果出错，返回空数组
	}
	log.Printf("[GetSalesCustomerDetail] 获取最近订单成功: userID=%d, count=%d", id, len(recentOrders))

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"user":          user,
			"addresses":     addresses,
			"address_count": addressCount,
			"order_count":   orderCount,
			"total_amount":  totalAmount,
			"recent_orders": recentOrders,
		},
		"message": "获取成功",
	})
}

// GetSalesCustomerOrders 获取客户的订单列表（销售员）
func GetSalesCustomerOrders(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户ID格式错误"})
		return
	}

	// 验证客户是否属于当前销售员
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}

	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问此客户信息"})
		return
	}

	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	status := c.Query("status")

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 构建查询条件
	where := "user_id = ?"
	args := []interface{}{userID}

	// 状态筛选
	if status != "" {
		if status == "pending_delivery" || status == "pending" {
			where += " AND (status = ? OR status = 'pending')"
			args = append(args, "pending_delivery")
		} else if status == "delivered" || status == "shipped" {
			where += " AND (status = ? OR status = 'shipped')"
			args = append(args, "delivered")
		} else {
			where += " AND status = ?"
			args = append(args, status)
		}
	}

	// 计算偏移量
	offset := (pageNum - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// 获取分页数据
	query := `
		SELECT id, order_number, user_id, address_id, status, goods_amount, delivery_fee, points_discount,
		       coupon_discount, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, created_at, updated_at
		FROM orders WHERE ` + where + ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询订单失败: " + err.Error()})
		return
	}
	defer rows.Close()

	var orders []map[string]interface{}
	for rows.Next() {
		var order model.Order
		var expectedDeliveryAt sql.NullTime
		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID,
			&order.Status, &order.GoodsAmount, &order.DeliveryFee, &order.PointsDiscount,
			&order.CouponDiscount, &order.TotalAmount, &order.Remark,
			&order.OutOfStockStrategy, &order.TrustReceipt, &order.HidePrice,
			&order.RequirePhoneContact, &expectedDeliveryAt, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			log.Printf("扫描订单数据失败: %v", err)
			continue
		}

		if expectedDeliveryAt.Valid {
			order.ExpectedDeliveryAt = &expectedDeliveryAt.Time
		}

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
			"total_amount":          order.TotalAmount,
			"remark":                order.Remark,
			"out_of_stock_strategy": order.OutOfStockStrategy,
			"trust_receipt":         order.TrustReceipt,
			"hide_price":           order.HidePrice,
			"require_phone_contact": order.RequirePhoneContact,
			"expected_delivery_at":  order.ExpectedDeliveryAt,
			"created_at":            order.CreatedAt,
			"updated_at":            order.UpdatedAt,
		}
		orders = append(orders, orderData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  orders,
			"total": len(orders), // 简化处理，实际应该查询总数
		},
		"message": "获取成功",
	})
}

// GetSalesCustomerFrequentProducts 获取客户的常购商品列表（销售员）
func GetSalesCustomerFrequentProducts(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户ID格式错误"})
		return
	}

	// 验证客户是否属于当前销售员
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}

	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问此客户信息"})
		return
	}

	// 查询客户的订单明细，按商品+规格分组统计购买次数
	query := `
		SELECT 
			oi.product_id,
			oi.product_name,
			oi.spec_name,
			oi.image,
			COUNT(*) as buy_count,
			MAX(o.created_at) as last_buy_at
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.user_id = ?
		GROUP BY oi.product_id, oi.product_name, oi.spec_name, oi.image
		ORDER BY buy_count DESC, last_buy_at DESC
		LIMIT 50
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		log.Printf("查询客户常购商品失败: userID=%d, error=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询常购商品失败"})
		return
	}
	defer rows.Close()

	var products []FrequentProduct
	for rows.Next() {
		var p FrequentProduct
		var lastBuyAt string // To scan MAX(o.created_at)
		if err := rows.Scan(&p.ProductID, &p.ProductName, &p.SpecName, &p.Image, &p.BuyCount, &lastBuyAt); err != nil {
			log.Printf("扫描常购商品结果失败: %v", err)
			continue
		}

		// 获取商品详情（包含规格信息用于价格显示）
		product, err := model.GetProductByID(p.ProductID)
		if err == nil && product != nil {
			p.Product = product
		}

		products = append(products, p)
	}

	if products == nil {
		products = []FrequentProduct{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    products,
		"message": "获取成功",
	})
}

// GetSalesCustomerPurchaseList 获取客户的采购单（销售员）
func GetSalesCustomerPurchaseList(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户ID格式错误"})
		return
	}

	// 验证客户是否属于当前销售员
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}

	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问此客户信息"})
		return
	}

	// 在获取采购单时，备份用户的原始采购单（用于后续创建订单时恢复）
	// 这样备份的就是用户原来的采购单，不包含销售员后续添加的新商品
	// 注意：每次调用都会备份，前端应该保存第一次调用时返回的备份
	userPurchaseListBackup, err := model.BackupPurchaseList(userID)
	if err != nil {
		// 备份失败不影响获取采购单，只记录日志
		log.Printf("[GetSalesCustomerPurchaseList] 备份采购单失败: %v", err)
		userPurchaseListBackup = []model.PurchaseListItem{}
	}

	// 获取采购单
	items, err := model.GetPurchaseListItemsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	// 获取用户类型
	userType := user.UserType
	if userType == "" || userType == "unknown" {
		userType = "retail"
	}

	// 计算配送费
	summary, err := model.CalculateDeliveryFee(items, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	// 获取加急费用（从系统设置）
	urgentFeeStr, _ := model.GetSystemSetting("order_urgent_fee")
	urgentFee := 0.0
	if urgentFeeStr != "" {
		if fee, parseErr := strconv.ParseFloat(urgentFeeStr, 64); parseErr == nil && fee > 0 {
			urgentFee = fee
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"items":      items,
			"summary":    summary,
			"backup":     userPurchaseListBackup, // 返回备份，前端必须保存并在创建订单时传入
			"urgent_fee": urgentFee,              // 返回加急费用供前端显示
		},
		"message": "获取成功",
	})
}

// UpdateSalesCustomerPurchaseItem 更新客户采购单中某一条目的数量（销售员代客调整购物车）
func UpdateSalesCustomerPurchaseItem(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户ID格式错误"})
		return
	}

	itemIDStr := c.Param("itemId")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单项ID格式错误"})
		return
	}

	// 校验客户归属
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权修改该客户采购单"})
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "数量必须大于0"})
		return
	}

	if err := model.UpdatePurchaseListItemQuantity(itemID, userID, req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 修改后重新返回该客户采购单及配送费汇总
	items, err := model.GetPurchaseListItemsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	userType := user.UserType
	if userType == "" || userType == "unknown" {
		userType = "retail"
	}

	summary, err := model.CalculateDeliveryFee(items, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"items":   items,
			"summary": summary,
		},
		"message": "更新成功",
	})
}

// DeleteSalesCustomerPurchaseItem 删除客户采购单中的某一条目（销售员代客删商品）
func DeleteSalesCustomerPurchaseItem(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户ID格式错误"})
		return
	}

	itemIDStr := c.Param("itemId")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单项ID格式错误"})
		return
	}

	// 校验客户归属
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权修改该客户采购单"})
		return
	}

	if err := model.DeletePurchaseListItem(itemID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 删除后重新返回该客户采购单及配送费汇总
	items, err := model.GetPurchaseListItemsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	userType := user.UserType
	if userType == "" || userType == "unknown" {
		userType = "retail"
	}

	summary, err := model.CalculateDeliveryFee(items, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"items":   items,
			"summary": summary,
		},
		"message": "删除成功",
	})
}

// AddSalesCustomerPurchaseItem 为客户的采购单新增一条商品（销售员代客加购）
func AddSalesCustomerPurchaseItem(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户ID格式错误"})
		return
	}

	// 校验客户归属
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权修改该客户采购单"})
		return
	}

	// 请求参数：商品ID、规格名称、数量
	var req struct {
		ProductID int    `json:"product_id" binding:"required"`
		SpecName  string `json:"spec_name" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "数量必须大于0"})
		return
	}

	// 获取商品及规格信息
	product, err := model.GetProductByID(req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品信息失败: " + err.Error()})
		return
	}
	if product == nil || product.Status == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在或已下架"})
		return
	}

	specName := strings.TrimSpace(req.SpecName)
	if specName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "规格名称不能为空"})
		return
	}

	var specSnapshot model.PurchaseSpecSnapshot
	found := false
	for _, spec := range product.Specs {
		if spec.Name == specName {
			specSnapshot = model.PurchaseSpecSnapshot{
				Name:           spec.Name,
				Description:    spec.Description,
				Cost:           spec.Cost,
				WholesalePrice: spec.WholesalePrice,
				RetailPrice:    spec.RetailPrice,
			}
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "指定规格不存在"})
		return
	}

	// 取商品首图作为采购单图片
	productImage := ""
	if len(product.Images) > 0 {
		productImage = product.Images[0]
	}

	item := &model.PurchaseListItem{
		UserID:       userID,
		ProductID:    product.ID,
		ProductName:  product.Name,
		ProductImage: productImage,
		SpecName:     specName,
		SpecSnapshot: specSnapshot,
		Quantity:     req.Quantity,
		IsSpecial:    product.IsSpecial,
	}

	if _, err := model.AddOrUpdatePurchaseListItem(item); err != nil {
		log.Printf("[AddSalesCustomerPurchaseItem] 加入采购单失败: 用户ID=%d, 商品ID=%d, 规格=%s, 错误=%v", userID, item.ProductID, item.SpecName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "加入采购单失败: " + err.Error()})
		return
	}
	log.Printf("[AddSalesCustomerPurchaseItem] 成功添加商品到采购单: 用户ID=%d, 商品ID=%d, 规格=%s, 数量=%d", userID, item.ProductID, item.SpecName, item.Quantity)

	// 返回最新的采购单及配送费汇总
	items, err := model.GetPurchaseListItemsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	userType := user.UserType
	if userType == "" || userType == "unknown" {
		userType = "retail"
	}

	summary, err := model.CalculateDeliveryFee(items, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"items":   items,
			"summary": summary,
		},
		"message": "添加成功",
	})
}

// CreateOrderForCustomer 为客户创建订单（销售员）
func CreateOrderForCustomer(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	var req struct {
		UserID              int                      `json:"user_id" binding:"required"`
		AddressID           int                      `json:"address_id" binding:"required"`
		ItemIDs             []int                    `json:"item_ids"`              // 采购单项ID列表，为空则使用全部
		Remark              string                   `json:"remark"`                // 订单备注
		CouponID            int                      `json:"coupon_id"`             // 用户优惠券ID（user_coupon_id），可选
		OutOfStockStrategy  string                   `json:"out_of_stock_strategy"` // 缺货处理：cancel_item / ship_available / contact_me
		TrustReceipt        bool                     `json:"trust_receipt"`         // 信任签收
		HidePrice           bool                     `json:"hide_price"`            // 是否隐藏价格
		RequirePhoneContact bool                     `json:"require_phone_contact"` // 配送时是否电话联系
		IsUrgent            bool                     `json:"is_urgent"`             // 是否加急订单
		PurchaseListBackup  []model.PurchaseListItem `json:"purchase_list_backup"`  // 用户原来的采购单备份（从GetSalesCustomerPurchaseList获取，必须传入）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 验证客户是否属于当前销售员
	user, err := model.GetMiniAppUserByID(req.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}

	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权为此客户创建订单"})
		return
	}

	// 验证地址归属
	address, err := model.GetAddressByID(req.AddressID)
	if err != nil || address == nil || address.UserID != req.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "收货地址无效"})
		return
	}

	// 使用前端传入的备份数据（在GetSalesCustomerPurchaseList时获取的）
	// 这个备份是用户进入开单页面时的原始采购单，不包含销售员后续的任何操作
	userPurchaseListBackup := req.PurchaseListBackup
	if len(userPurchaseListBackup) == 0 {
		// 如果没有传入备份，说明前端没有保存备份，这是不正确的行为
		// 但为了兼容，我们在创建订单前备份（这时可能已经包含销售员添加的商品了）
		// 注意：这种情况下，恢复后可能仍然包含销售员添加的商品，这是不正确的
		backup, err := model.BackupPurchaseList(req.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "备份采购单失败: " + err.Error()})
			return
		}
		userPurchaseListBackup = backup
		log.Printf("[CreateOrderForCustomer] 警告：前端未传入备份数据，使用创建订单前的备份（可能包含销售员添加的商品，恢复后可能不准确）")
	} else {
		log.Printf("[CreateOrderForCustomer] 使用前端传入的备份数据，备份商品数量: %d", len(userPurchaseListBackup))
	}

	// 获取采购单（此时可能包含销售员在开单时添加的新商品、修改的数量等）
	items, err := model.GetPurchaseListItemsByUserID(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单为空，无法创建订单"})
		return
	}

	// 筛选指定的 item_ids
	selectedItemIDs := make(map[int]struct{})
	if len(req.ItemIDs) > 0 {
		for _, id := range req.ItemIDs {
			if id > 0 {
				selectedItemIDs[id] = struct{}{}
			}
		}
		filteredItems := make([]model.PurchaseListItem, 0, len(selectedItemIDs))
		for _, item := range items {
			if _, ok := selectedItemIDs[item.ID]; ok {
				filteredItems = append(filteredItems, item)
			}
		}
		items = filteredItems
	}

	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单为空，无法创建订单"})
		return
	}

	// 获取用户类型
	userType := user.UserType
	if userType == "" || userType == "unknown" {
		userType = "retail"
	}

	// 计算配送费
	summary, err := model.CalculateDeliveryFee(items, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	// 计算订单金额和分类信息用于优惠券筛选
	orderAmount := 0.0
	categoryIDSet := make(map[int]struct{})
	productIDs := make([]int, 0, len(items))
	for _, item := range items {
		// 根据用户类型计算商品金额
		var price float64
		if userType == "wholesale" {
			price = item.SpecSnapshot.WholesalePrice
			if price <= 0 {
				price = item.SpecSnapshot.RetailPrice
			}
		} else {
			price = item.SpecSnapshot.RetailPrice
			if price <= 0 {
				price = item.SpecSnapshot.WholesalePrice
			}
		}
		if price <= 0 {
			price = item.SpecSnapshot.Cost
		}
		if price < 0 {
			price = 0
		}
		orderAmount += price * float64(item.Quantity)
		productIDs = append(productIDs, item.ProductID)
	}

	categoryInfo, err := model.FetchProductCategoryInfo(productIDs)
	if err == nil {
		for _, info := range categoryInfo {
			if info.CategoryID > 0 {
				categoryIDSet[info.CategoryID] = struct{}{}
			}
			if info.ParentID > 0 {
				categoryIDSet[info.ParentID] = struct{}{}
			}
		}
	}
	categoryIDs := make([]int, 0, len(categoryIDSet))
	for id := range categoryIDSet {
		categoryIDs = append(categoryIDs, id)
	}

	// 获取可用优惠券并计算折扣
	couponDiscount := 0.0
	var selectedCoupon *model.AvailableCouponInfo
	if req.CouponID > 0 {
		availableCoupons, err := model.GetAvailableCouponsForPurchaseList(
			req.UserID,
			orderAmount,
			categoryIDs,
			summary.DeliveryFee,
			summary.IsFreeShipping,
		)
		if err == nil {
			// 查找指定的优惠券
			for i := range availableCoupons {
				if availableCoupons[i].UserCouponID == req.CouponID && availableCoupons[i].IsAvailable {
					selectedCoupon = &availableCoupons[i]
					// 计算折扣金额
					if selectedCoupon.Type == "delivery_fee" {
						couponDiscount = summary.DeliveryFee
					} else if selectedCoupon.Type == "amount" {
						couponDiscount = selectedCoupon.DiscountValue
					}
					break
				}
			}
		}
	}

	// 处理缺货策略，默认为 contact_me
	outOfStockStrategy := req.OutOfStockStrategy
	if outOfStockStrategy == "" {
		outOfStockStrategy = "contact_me"
	}

	// 获取加急费用（从系统设置）
	urgentFee := 0.0
	if req.IsUrgent {
		urgentFeeStr, err := model.GetSystemSetting("order_urgent_fee")
		if err == nil && urgentFeeStr != "" {
			if fee, parseErr := strconv.ParseFloat(urgentFeeStr, 64); parseErr == nil && fee > 0 {
				urgentFee = fee
			}
		}
	}

	// 创建订单
	options := model.OrderCreationOptions{
		Remark:              req.Remark,
		OutOfStockStrategy:  outOfStockStrategy,
		TrustReceipt:        req.TrustReceipt,
		HidePrice:           req.HidePrice,
		RequirePhoneContact: req.RequirePhoneContact,
		PointsDiscount:      0,
		CouponDiscount:      couponDiscount,
		IsUrgent:            req.IsUrgent,
		UrgentFee:           urgentFee,
	}

	// 创建订单（注意：CreateOrderFromPurchaseList 不再清空采购单）
	order, orderItems, err := model.CreateOrderFromPurchaseList(req.UserID, req.AddressID, items, summary, options, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建订单失败: " + err.Error()})
		return
	}

	// 创建订单成功后，删除用于创建订单的商品，然后恢复用户原来的采购单
	// 删除用于创建订单的商品
	if len(selectedItemIDs) > 0 {
		// 删除指定的商品
		itemIDList := make([]interface{}, 0, len(selectedItemIDs))
		for id := range selectedItemIDs {
			itemIDList = append(itemIDList, id)
		}
		placeholders := strings.Repeat("?,", len(itemIDList))
		placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
		query := fmt.Sprintf("DELETE FROM purchase_list_items WHERE user_id = ? AND id IN (%s)", placeholders)
		args := append([]interface{}{req.UserID}, itemIDList...)
		_, err = database.DB.Exec(query, args...)
		if err != nil {
			log.Printf("[CreateOrderForCustomer] 删除已下单商品失败: %v", err)
		}
	} else {
		// 如果没有指定 item_ids，删除所有商品（因为创建订单时使用了所有商品）
		_, err = database.DB.Exec("DELETE FROM purchase_list_items WHERE user_id = ?", req.UserID)
		if err != nil {
			log.Printf("[CreateOrderForCustomer] 清空采购单失败: %v", err)
		}
	}

	// 恢复用户原来的采购单（无条件恢复备份，不管销售员做了什么操作）
	// 备份是用户进入开单页面时的原始采购单，不包含销售员后续的任何操作
	// 无论销售员增加了商品、减少了商品、修改了数量，都恢复到备份的状态
	if len(userPurchaseListBackup) > 0 {
		// 直接恢复备份，不需要任何过滤
		// 因为备份就是用户进入开单页面时的原始状态
		log.Printf("[CreateOrderForCustomer] 开始恢复采购单，备份商品数量: %d", len(userPurchaseListBackup))
		if err := model.RestorePurchaseList(req.UserID, userPurchaseListBackup); err != nil {
			log.Printf("[CreateOrderForCustomer] 恢复采购单失败: %v", err)
		} else {
			log.Printf("[CreateOrderForCustomer] 成功恢复采购单，恢复商品数量: %d", len(userPurchaseListBackup))
		}
	} else {
		log.Printf("[CreateOrderForCustomer] 警告：备份数据为空，无法恢复采购单")
	}

	// 创建订单成功后，使用优惠券（标记为已使用并关联订单ID）
	// 使用 user_coupon_id 精确更新，避免用户有多张相同优惠券时误更新
	if selectedCoupon != nil && req.CouponID > 0 {
		if err := model.UseCouponByUserCouponID(req.CouponID, order.ID); err != nil {
			// 如果使用失败，记录错误但不影响订单创建
			log.Printf("[CreateOrderForCustomer] 标记优惠券为已使用失败 (userCouponID=%d, orderID=%d): %v", req.CouponID, order.ID, err)
		} else {
			log.Printf("[CreateOrderForCustomer] 成功标记优惠券为已使用 (userCouponID=%d, orderID=%d)", req.CouponID, order.ID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"order":       order,
			"order_items": orderItems,
		},
		"message": "创建订单成功",
	})
}

// GetSalesProducts 获取商品列表（销售员查询商品用，支持搜索和分类筛选）
func GetSalesProducts(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	// 解析查询参数
	keyword := strings.TrimSpace(c.Query("keyword"))
	categoryIDStr := strings.TrimSpace(c.Query("categoryId"))
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 20)

	// 限制分页参数范围
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // 最大每页100条
	}

	var products []model.Product
	var total int
	var err error

	// 解析分类ID
	var categoryID int
	if categoryIDStr != "" {
		if id, err := strconv.Atoi(categoryIDStr); err == nil && id > 0 {
			categoryID = id
		}
	}

	// 根据条件查询商品
	if categoryID > 0 && keyword != "" {
		// 同时有分类和关键词：先按分类查询，再在结果中筛选关键词
		products, _, err = model.GetProductsByCategoryWithPagination(categoryID, pageNum, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品失败: " + err.Error()})
			return
		}
		// 在结果中进一步筛选关键词
		filteredProducts := []model.Product{}
		keywordLower := strings.ToLower(keyword)
		for _, product := range products {
			if strings.Contains(strings.ToLower(product.Name), keywordLower) ||
				strings.Contains(strings.ToLower(product.Description), keywordLower) {
				filteredProducts = append(filteredProducts, product)
			}
		}
		products = filteredProducts
		total = len(filteredProducts)
	} else if categoryID > 0 {
		// 只有分类筛选
		products, total, err = model.GetProductsByCategoryWithPagination(categoryID, pageNum, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品失败: " + err.Error()})
			return
		}
	} else if keyword != "" {
		// 只有关键词搜索
		products, total, err = model.SearchProductsWithPagination(keyword, pageNum, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "搜索商品失败: " + err.Error()})
			return
		}
	} else {
		// 没有筛选条件，获取所有商品（带分页）
		products, total, err = model.GetAllProductsWithPagination(pageNum, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品失败: " + err.Error()})
			return
		}
	}

	// 构建返回数据结构
	result := map[string]interface{}{
		"list":     products,
		"total":    total,
		"pageNum":  pageNum,
		"pageSize": pageSize,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    result,
	})
}

// PreviewRiderDeliveryFee 预览配送员配送费（基于采购单和地址）
func PreviewRiderDeliveryFee(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}
	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "客户ID格式错误"})
		return
	}

	// 从查询参数获取地址ID和是否加急
	addressIDStr := c.Query("address_id")
	var addressID int
	if addressIDStr != "" {
		addressID, err = strconv.Atoi(addressIDStr)
		if err != nil || addressID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址ID格式错误"})
			return
		}
	}

	isUrgent := c.Query("is_urgent") == "true"

	// 校验客户归属
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问此客户信息"})
		return
	}

	// 获取采购单
	items, err := model.GetPurchaseListItemsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	if len(items) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"rider_payable_fee": 0.0,
				"base_fee":          0.0,
				"isolated_fee":      0.0,
				"item_fee":          0.0,
				"urgent_fee":        0.0,
				"weather_fee":       0.0,
				"profit_share":      0.0,
			},
			"message": "采购单为空",
		})
		return
	}

	// 获取用户类型
	userType := user.UserType
	if userType == "" || userType == "unknown" {
		userType = "retail"
	}

	// 计算配送员配送费
	result, err := model.CalculateRiderDeliveryFeePreview(items, addressID, isUrgent, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    result,
		"message": "获取成功",
	})
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// boolToTinyInt 将布尔值转换为TinyInt（0或1）
func boolToTinyInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// calculateSalesCommissionPreviewForSales 计算销售分成预览（销售员端）
func calculateSalesCommissionPreviewForSales(order *model.Order, deliveryFeeResult *model.DeliveryFeeCalculationResult, orderProfit float64, salesCode string) map[string]interface{} {
	// 计算订单金额、商品成本、配送成本
	orderAmount := order.TotalAmount             // 平台总收入
	goodsCost := order.GoodsAmount - orderProfit // 商品总成本
	deliveryCost := 0.0
	if deliveryFeeResult != nil {
		deliveryCost = deliveryFeeResult.TotalPlatformCost
	}

	// 判断是否新客户
	isNewCustomer, _ := model.IsNewCustomerOrder(order.UserID, order.ID)

	// 获取当月有效订单总金额（用于计算阶梯提成）
	currentMonth := time.Now().Format("2006-01")
	monthTotalSales, _ := model.GetMonthlyTotalSales(salesCode, currentMonth)

	// 计算分成
	calcResult, err := model.CalculateSalesCommission(
		salesCode,
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
		"order_profit":          orderProfit,
		"order_amount":          orderAmount,
		"goods_cost":            goodsCost,
		"delivery_cost":         deliveryCost,
	}
}
