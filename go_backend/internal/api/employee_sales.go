package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"go_backend/internal/database"
	"go_backend/internal/model"

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

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"user":      user,
			"addresses": addresses,
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

	// 获取总数量
	var total int
	countQuery := "SELECT COUNT(*) FROM orders WHERE " + where
	err = database.DB.QueryRow(countQuery, args...).Scan(&total)
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
		       coupon_discount, total_amount, remark, out_of_stock_strategy, trust_receipt,
		       hide_price, require_phone_contact, expected_delivery_at, created_at, updated_at
		FROM orders WHERE ` + where + ` ORDER BY id DESC LIMIT ? OFFSET ?`
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

		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID, &order.Status, &order.GoodsAmount, &order.DeliveryFee,
			&order.PointsDiscount, &order.CouponDiscount, &order.TotalAmount, &order.Remark,
			&order.OutOfStockStrategy, &order.TrustReceipt, &order.HidePrice, &order.RequirePhoneContact,
			&expectedDelivery, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if expectedDelivery.Valid {
			t := expectedDelivery.Time
			order.ExpectedDeliveryAt = &t
		}

		itemCount, _ := model.GetOrderItemCountByOrderID(order.ID)

		orderData := map[string]interface{}{
			"id":           order.ID,
			"order_number": order.OrderNumber,
			"status":       order.Status,
			"goods_amount": order.GoodsAmount,
			"delivery_fee": order.DeliveryFee,
			"total_amount": order.TotalAmount,
			"item_count":   itemCount,
			"created_at":   order.CreatedAt,
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

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"items":  items,
			"summary": summary,
		},
		"message": "获取成功",
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
		UserID    int   `json:"user_id" binding:"required"`
		AddressID int   `json:"address_id" binding:"required"`
		ItemIDs   []int `json:"item_ids"` // 采购单项ID列表，为空则使用全部
		Remark    string `json:"remark"`
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

	// 获取采购单
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
	if len(req.ItemIDs) > 0 {
		filter := make(map[int]struct{}, len(req.ItemIDs))
		for _, id := range req.ItemIDs {
			if id > 0 {
				filter[id] = struct{}{}
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

	// 创建订单
	options := model.OrderCreationOptions{
		Remark:              req.Remark,
		OutOfStockStrategy:  "contact_me",
		TrustReceipt:        false,
		HidePrice:           false,
		RequirePhoneContact: true,
		PointsDiscount:      0,
		CouponDiscount:      0,
	}

	order, orderItems, err := model.CreateOrderFromPurchaseList(req.UserID, req.AddressID, items, summary, options, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建订单失败: " + err.Error()})
		return
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

// GetSalesProducts 获取商品列表（销售员创建订单用）
func GetSalesProducts(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	// 可以复用现有的商品列表API，或者创建一个简化的版本
	// 这里先返回成功，具体实现可以根据需求调整
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    []interface{}{},
	})
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

