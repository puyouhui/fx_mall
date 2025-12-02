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
			"id":      address.ID,
			"name":    address.Name,
			"contact": address.Contact,
			"phone":   address.Phone,
			"address": address.Address,
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

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"order":       order,
			"order_items": items,
			"address":     addressData,
			"user":        userData,
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
	recentOrders, _ := model.GetRecentOrdersByUserID(id, 3)

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

	// 获取总数量（这里使用用户整体订单数量，忽略筛选条件）
	total, _ := model.CountOrdersByUserID(userID)

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
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "加入采购单失败: " + err.Error()})
		return
	}

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
		products, total, err = model.GetProductsByCategoryWithPagination(categoryID, pageNum, pageSize)
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

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
