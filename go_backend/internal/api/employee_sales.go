package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

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

	// 只能查询自己名下的客户
	if user.SalesCode != employee.EmployeeCode {
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

	// 校验客户是否属于当前销售员
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取客户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
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

	// 校验客户是否属于当前销售员
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取客户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "客户不存在"})
		return
	}
	if user.SalesCode != employee.EmployeeCode {
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
	if user.SalesCode != employee.EmployeeCode {
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
			"items":   items,
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
		UserID    int    `json:"user_id" binding:"required"`
		AddressID int    `json:"address_id" binding:"required"`
		ItemIDs   []int  `json:"item_ids"` // 采购单项ID列表，为空则使用全部
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
