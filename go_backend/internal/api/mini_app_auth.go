package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go_backend/internal/config"
	"go_backend/internal/database"
	"go_backend/internal/model"
	"go_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type miniAppLoginRequest struct {
	Code       string `json:"code" binding:"required"`
	ReferrerID *int   `json:"referrer_id,omitempty"` // 分享者用户ID（可选）
}

type weChatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// MiniAppLogin 小程序登录，仅记录用户唯一ID（openid）
func MiniAppLogin(c *gin.Context) {
	var req miniAppLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "登录参数错误"})
		return
	}

	sessionInfo, err := fetchWeChatSessionInfo(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户凭证失败: " + err.Error()})
		return
	}

	if sessionInfo.OpenID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "未能获取到用户唯一ID"})
		return
	}

	user, err := model.GetMiniAppUserByUniqueID(sessionInfo.OpenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询用户信息失败: " + err.Error()})
		return
	}

	if user == nil {
		// 新用户，绑定分享者（如果提供）
		user, err = model.CreateMiniAppUser(sessionInfo.OpenID, req.ReferrerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败: " + err.Error()})
			return
		}

		// 新客户登录即送奖励（使用 activity_type = new_customer 的活动配置），异步发放，避免阻塞登录
		go func(userID int) {
			if err := model.GrantNewCustomerLoginReward(userID); err != nil {
				log.Printf("[MiniAppLogin] 发放新客户登录奖励失败(user_id=%d): %v", userID, err)
			}
		}(user.ID)
	} else {
		// 如果用户已存在但没有编号，生成一个
		if user.UserCode == "" {
			userCode, genErr := model.GenerateUserCode()
			if genErr == nil {
				_, execErr := database.DB.Exec(`
					UPDATE mini_app_users SET user_code = ?, updated_at = NOW() WHERE unique_id = ?
				`, userCode, sessionInfo.OpenID)
				if execErr == nil {
					user.UserCode = userCode
				}
			}
		}

		// 检查用户资料完善状态，如果已完善但用户类型是unknown，自动设置为retail
		if user.ProfileCompleted && (user.UserType == "" || user.UserType == "unknown") {
			_, _ = database.DB.Exec(`
				UPDATE mini_app_users SET user_type = 'retail', updated_at = NOW() WHERE unique_id = ?
			`, sessionInfo.OpenID)
			user.UserType = "retail"
		}
	}

	token, err := utils.GenerateMiniAppToken(user.UniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成登录凭证失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"unique_id": user.UniqueID,
			"token":     token,
			"user":      user,
		},
	})
}

// GetUserReferralStats 获取用户拉新统计（后台管理使用）
func GetUserReferralStats(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供用户ID"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	// 统计该用户拉取的新用户数量
	var count int
	err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM mini_app_users WHERE referrer_id = ?
	`, userID).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "统计失败: " + err.Error()})
		return
	}

	// 获取该用户拉取的新用户列表
	referrals, err := model.GetMiniAppUsersByReferrerID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取新用户列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"total":     count,
			"referrals": referrals,
		},
	})
}

// GetMiniAppUsers 获取小程序用户（后台管理使用）
func GetMiniAppUsers(c *gin.Context) {
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	keyword := c.Query("keyword")

	users, total, err := model.GetMiniAppUsers(pageNum, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户列表失败: " + err.Error()})
		return
	}

	// 为每个用户添加销售员信息和优惠券数量
	usersWithSalesEmployee := make([]map[string]interface{}, 0)
	for _, user := range users {
		userData := map[string]interface{}{
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
			"is_sales_employee": user.IsSalesEmployee,
			"created_at":        user.CreatedAt,
			"updated_at":        user.UpdatedAt,
		}

		// 如果用户绑定了销售员，获取销售员信息
		if user.SalesCode != "" {
			employee, err := model.GetEmployeeByEmployeeCode(user.SalesCode)
			if err == nil && employee != nil {
				userData["sales_employee"] = map[string]interface{}{
					"id":            employee.ID,
					"employee_code": employee.EmployeeCode,
					"name":          employee.Name,
					"phone":         employee.Phone,
				}
			} else {
				userData["sales_employee"] = nil
			}
		} else {
			userData["sales_employee"] = nil
		}

		// 统计用户优惠券数量（未使用的）
		var couponCount int
		err := database.DB.QueryRow("SELECT COUNT(*) FROM user_coupons WHERE user_id = ? AND status = 'unused'", user.ID).Scan(&couponCount)
		if err == nil {
			userData["coupon_count"] = couponCount
		} else {
			userData["coupon_count"] = 0
		}

		// 获取用户的默认地址
		defaultAddress, _ := model.GetDefaultAddressByUserID(user.ID)
		if defaultAddress != nil {
			userData["default_address"] = map[string]interface{}{
				"id":      defaultAddress.ID,
				"name":    defaultAddress.Name,
				"contact": defaultAddress.Contact,
				"phone":   defaultAddress.Phone,
				"address": defaultAddress.Address,
			}
		} else {
			userData["default_address"] = nil
		}

		usersWithSalesEmployee = append(usersWithSalesEmployee, userData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"data":  usersWithSalesEmployee,
		"total": total,
	})
}

// GetMiniAppCurrentUser 获取当前登录的小程序用户信息
func GetMiniAppCurrentUser(c *gin.Context) {
	// 从中间件设置的上下文中获取 openID
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	// 如果用户没有编号，生成一个
	if user.UserCode == "" {
		userCode, genErr := model.GenerateUserCode()
		if genErr == nil {
			_, execErr := database.DB.Exec(`
				UPDATE mini_app_users SET user_code = ?, updated_at = NOW() WHERE unique_id = ?
			`, userCode, uniqueID)
			if execErr == nil {
				user.UserCode = userCode
			}
		}
	}

	// 构建返回数据
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
		"points":            user.Points,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
	}

	// 如果用户绑定了销售员，获取销售员信息
	if user.SalesCode != "" {
		employee, err := model.GetEmployeeByEmployeeCode(user.SalesCode)
		if err == nil && employee != nil {
			responseData["sales_employee"] = map[string]interface{}{
				"id":            employee.ID,
				"employee_code": employee.EmployeeCode,
				"name":          employee.Name,
				"phone":         employee.Phone,
			}
		} else {
			responseData["sales_employee"] = nil
		}
	} else {
		responseData["sales_employee"] = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    responseData,
	})
}

// GetMiniAppUserDetail 获取小程序用户详情（后台管理使用）
func GetMiniAppUserDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供用户ID"})
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	user, err := model.GetMiniAppUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户详情失败: " + err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	// 获取用户的默认地址
	defaultAddress, _ := model.GetDefaultAddressByUserID(user.ID)

	// 获取用户的所有地址
	allAddresses, _ := model.GetAddressesByUserID(user.ID)

	// 获取用户的发票抬头
	invoice, _ := model.GetInvoiceByUserID(user.ID)

	// 构建返回数据
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
		"is_sales_employee": user.IsSalesEmployee,
		"points":            user.Points,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
		"default_address":   defaultAddress, // 默认地址信息（保留兼容）
		"addresses":         allAddresses,   // 所有地址列表
		"invoice":           invoice,         // 发票抬头信息
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    responseData,
	})
}

// SaveAdminInvoice 管理员保存发票抬头
func SaveAdminInvoice(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供用户ID"})
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	user, err := model.GetMiniAppUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	var req struct {
		InvoiceType    string `json:"invoice_type" binding:"required"`
		Title          string `json:"title" binding:"required"`
		TaxNumber      string `json:"tax_number"`
		CompanyAddress string `json:"company_address"`
		CompanyPhone   string `json:"company_phone"`
		BankName       string `json:"bank_name"`
		BankAccount    string `json:"bank_account"`
		IsDefault      bool   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "发票抬头不能为空"})
		return
	}

	// 如果是企业类型，纳税人识别号必填
	if req.InvoiceType == "company" && req.TaxNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "企业发票纳税人识别号不能为空"})
		return
	}

	invoiceData := map[string]interface{}{
		"invoice_type":    req.InvoiceType,
		"title":           req.Title,
		"tax_number":      req.TaxNumber,
		"company_address": req.CompanyAddress,
		"company_phone":   req.CompanyPhone,
		"bank_name":       req.BankName,
		"bank_account":    req.BankAccount,
		"is_default":      req.IsDefault,
	}

	invoice, err := model.CreateOrUpdateInvoice(user.ID, invoiceData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存发票抬头失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "保存成功",
		"data":    invoice,
	})
}

// GetAdminAddressByID 管理员获取地址详情
func GetAdminAddressByID(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供地址ID"})
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址ID格式错误"})
		return
	}

	address, err := model.GetAddressByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取地址详情失败: " + err.Error()})
		return
	}

	if address == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "地址不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    address,
	})
}

type updateAdminAddressRequest struct {
	Name      string `json:"name"`
	Contact   string `json:"contact"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	Avatar    string `json:"avatar"`
	StoreType string `json:"storeType"`
	// SalesCode 已移除，销售员绑定到用户而不是地址
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	IsDefault bool     `json:"isDefault"`
}

// UpdateAdminAddress 管理员更新地址
func UpdateAdminAddress(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供地址ID"})
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址ID格式错误"})
		return
	}

	// 先获取地址信息，确认地址存在并获取userID
	address, err := model.GetAddressByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取地址信息失败: " + err.Error()})
		return
	}
	if address == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "地址不存在"})
		return
	}

	var req updateAdminAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 构建更新数据
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
	// 不再处理SalesCode，销售员绑定到用户而不是地址
	if req.Latitude != nil {
		addressData["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		addressData["longitude"] = *req.Longitude
	}
	addressData["is_default"] = req.IsDefault

	// 更新地址
	err = model.UpdateAddress(id, address.UserID, addressData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新地址失败: " + err.Error()})
		return
	}

	// 获取更新后的地址信息
	updatedAddress, err := model.GetAddressByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取更新后的地址信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    updatedAddress,
	})
}

// DeleteAdminAddress 管理员删除地址
// 注意：为避免历史订单详情丢失地址信息，这里会阻止删除已被订单引用的地址。
func DeleteAdminAddress(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供地址ID"})
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址ID格式错误"})
		return
	}

	// 先获取地址信息，确认地址存在并获取userID
	address, err := model.GetAddressByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取地址信息失败: " + err.Error()})
		return
	}
	if address == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "地址不存在"})
		return
	}

	// 如果地址被订单引用，则不允许删除（避免历史订单展示缺失）
	var usedCount int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM orders WHERE address_id = ?`, id).Scan(&usedCount); err == nil {
		if usedCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该地址已被订单引用，无法删除"})
			return
		}
	}

	// 执行删除
	if err := model.DeleteAddress(id, address.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除地址失败: " + err.Error()})
		return
	}

	// 如果删掉的是默认地址，且用户还有其他地址但没有默认，则自动把最新地址设为默认
	remaining, _ := model.GetAddressesByUserID(address.UserID)
	if len(remaining) > 0 {
		hasDefault := false
		for _, a := range remaining {
			if a.IsDefault {
				hasDefault = true
				break
			}
		}
		if !hasDefault {
			_ = model.SetDefaultAddress(remaining[0].ID, address.UserID)
		}
	}

	// 重新统计地址数量，更新资料完善状态（与小程序端保持一致）
	user, _ := model.GetMiniAppUserByID(address.UserID)
	addressCount, err := model.CountAddressesByUserID(address.UserID)
	if err == nil && user != nil {
		profileCompleted := addressCount >= 1
		userType := user.UserType
		if !profileCompleted && userType == "retail" {
			userType = "unknown"
		} else if profileCompleted && (userType == "" || userType == "unknown") {
			userType = "retail"
		}
		database.DB.Exec(`
			UPDATE mini_app_users 
			SET profile_completed = ?, user_type = ?, updated_at = NOW()
			WHERE id = ?
		`, profileCompleted, userType, address.UserID)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

type updateMiniAppUserByAdminRequest struct {
	Name             string  `json:"name"`
	Phone            string  `json:"phone"`
	StoreType        string  `json:"storeType"`
	SalesCode        *string `json:"salesCode"`        // 销售员代码（员工码），使用指针以区分是否传递
	SalesEmployeeID  *int    `json:"salesEmployeeId"` // 销售员ID（优先使用）
	Avatar           string  `json:"avatar"`
	UserType         string  `json:"userType"`
	ProfileCompleted *bool   `json:"profileCompleted,omitempty"`
	IsSalesEmployee  *bool   `json:"isSalesEmployee,omitempty"`  // 是否是销售员
}

// UpdateMiniAppUserByAdmin 管理员更新小程序用户信息（可修改所有字段包括用户类型）
func UpdateMiniAppUserByAdmin(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供用户ID"})
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	var req updateMiniAppUserByAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证用户类型
	if req.UserType != "" {
		req.UserType = strings.ToLower(strings.TrimSpace(req.UserType))
		if req.UserType != "retail" && req.UserType != "wholesale" && req.UserType != "unknown" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户类型仅支持 retail、wholesale 或 unknown"})
			return
		}
	}

	// 构建更新数据
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

	// 处理销售员绑定：优先使用SalesEmployeeID，如果没有则使用SalesCode
	if req.SalesEmployeeID != nil && *req.SalesEmployeeID > 0 {
		// 通过员工ID获取员工码
		employee, err := model.GetEmployeeByID(*req.SalesEmployeeID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "获取销售员信息失败: " + err.Error()})
			return
		}
		if employee == nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "销售员不存在"})
			return
		}
		if !employee.IsSales {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该员工不是销售员"})
			return
		}
		updateData["salesCode"] = employee.EmployeeCode
	} else if req.SalesCode != nil {
		// SalesCode 字段被传递了
		salesCodeValue := strings.TrimSpace(*req.SalesCode)
		if salesCodeValue != "" {
			// 验证销售员代码是否存在且是销售员
			employees, _ := model.GetSalesEmployees()
			found := false
			for _, emp := range employees {
				if emp.EmployeeCode == salesCodeValue {
					found = true
					break
				}
			}
			if !found {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "销售员代码不存在或不是有效的销售员"})
				return
			}
			updateData["salesCode"] = salesCodeValue
		} else {
			// 如果传空字符串，表示清除绑定
			updateData["salesCode"] = ""
		}
	}
	// 如果 SalesCode 和 SalesEmployeeID 都没有传递，则不更新销售员绑定

	if req.Avatar != "" {
		updateData["avatar"] = strings.TrimSpace(req.Avatar)
	}
	if req.UserType != "" {
		updateData["userType"] = req.UserType
	}
	if req.ProfileCompleted != nil {
		updateData["profileCompleted"] = *req.ProfileCompleted
	}
	if req.IsSalesEmployee != nil {
		updateData["isSalesEmployee"] = *req.IsSalesEmployee
	}

	// 更新用户信息
	if err := model.UpdateMiniAppUserByAdmin(id, updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户信息失败: " + err.Error()})
		return
	}

	// 返回更新后的用户信息
	user, err := model.GetMiniAppUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	// 构建返回数据，确保包含所有字段
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
		"is_sales_employee": user.IsSalesEmployee,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
	}

	// 如果用户绑定了销售员，获取销售员信息
	if user.SalesCode != "" {
		employee, err := model.GetEmployeeByEmployeeCode(user.SalesCode)
		if err == nil && employee != nil {
			responseData["sales_employee"] = map[string]interface{}{
				"id":            employee.ID,
				"employee_code": employee.EmployeeCode,
				"name":          employee.Name,
				"phone":         employee.Phone,
			}
		} else {
			responseData["sales_employee"] = nil
		}
	} else {
		responseData["sales_employee"] = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    responseData,
	})
}

func fetchWeChatSessionInfo(code string) (*weChatSessionResponse, error) {
	appID := config.Config.MiniApp.AppID
	appSecret := config.Config.MiniApp.AppSecret
	if appID == "" || appSecret == "" {
		return nil, fmt.Errorf("未配置小程序AppID或AppSecret")
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID,
		appSecret,
		code,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var session weChatSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, err
	}

	if session.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败: %s", session.ErrMsg)
	}

	return &session, nil
}

type updateMiniUserTypeRequest struct {
	UserType string `json:"user_type" binding:"required"`
}

// UpdateMiniAppUserType 更新小程序用户类型
func UpdateMiniAppUserType(c *gin.Context) {
	token := extractBearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	claims, err := utils.ParseMiniAppToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "登录状态已失效，请重新登录"})
		return
	}

	var req updateMiniUserTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	req.UserType = strings.ToLower(strings.TrimSpace(req.UserType))
	if req.UserType != "retail" && req.UserType != "wholesale" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户类型仅支持 retail 或 wholesale"})
		return
	}

	if err := model.UpdateMiniAppUserType(claims.OpenID, req.UserType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户类型失败: " + err.Error()})
		return
	}

	user, err := model.GetMiniAppUserByUniqueID(claims.OpenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    user,
	})
}

func extractBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	if len(header) > 7 && strings.EqualFold(header[0:6], "bearer") {
		return strings.TrimSpace(header[6:])
	}
	return header
}

// MiniAppAuthMiddleware 小程序用户认证中间件
func MiniAppAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
			c.Abort()
			return
		}

		claims, err := utils.ParseMiniAppToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "登录状态已失效，请重新登录"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("openID", claims.OpenID)
		c.Next()
	}
}

type updateMiniUserProfileRequest struct {
	AddressID *int     `json:"address_id,omitempty"` // 地址ID，为空表示新增，不为空表示编辑
	Name      string   `json:"name"`
	Contact   string   `json:"contact"`
	Phone     string   `json:"phone"`
	Address   string   `json:"address"`
	Avatar    string   `json:"avatar,omitempty"` // 地址照片（门头照片）
	StoreType string   `json:"storeType"`
	SalesCode string   `json:"salesCode"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	IsDefault bool     `json:"is_default"` // 是否设置为默认地址
}

// UpdateMiniAppUserProfile 创建或更新地址，根据地址数量判断资料是否完善
func UpdateMiniAppUserProfile(c *gin.Context) {
	token := extractBearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	claims, err := utils.ParseMiniAppToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "登录状态已失效，请重新登录"})
		return
	}

	var req updateMiniUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证必填字段
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "店铺名称不能为空"})
		return
	}
	if strings.TrimSpace(req.Contact) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "联系人不能为空"})
		return
	}
	if strings.TrimSpace(req.Phone) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "手机号码不能为空"})
		return
	}
	if strings.TrimSpace(req.Address) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址不能为空"})
		return
	}

	// 获取用户信息
	user, err := model.GetMiniAppUserByUniqueID(claims.OpenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	// 构建地址数据（不再包含sales_code，销售员绑定到用户）
	addressData := map[string]interface{}{
		"name":       strings.TrimSpace(req.Name),
		"contact":    strings.TrimSpace(req.Contact),
		"phone":      strings.TrimSpace(req.Phone),
		"address":    strings.TrimSpace(req.Address),
		"store_type": strings.TrimSpace(req.StoreType),
		"is_default": req.IsDefault,
	}

	// 如果传了销售员代码，更新用户表的sales_code（而不是地址表）
	if req.SalesCode != "" {
		salesCode := strings.TrimSpace(req.SalesCode)
		// 验证销售员代码是否存在且是销售员
		employees, _ := model.GetSalesEmployees()
		found := false
		for _, emp := range employees {
			if emp.EmployeeCode == salesCode {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "销售员代码不存在或不是有效的销售员"})
			return
		}
		// 更新用户表的sales_code
		userUpdateData := map[string]interface{}{
			"salesCode": salesCode,
		}
		err = model.UpdateMiniAppUserByAdmin(user.ID, userUpdateData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新销售员绑定失败: " + err.Error()})
			return
		}
	}
	if req.Avatar != "" {
		addressData["avatar"] = strings.TrimSpace(req.Avatar)
	}
	if req.Latitude != nil {
		addressData["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		addressData["longitude"] = *req.Longitude
	}

	// 如果经纬度为空，尝试自动解析地址
	if req.Latitude == nil && req.Longitude == nil && strings.TrimSpace(req.Address) != "" {
		// 获取地图API Key
		amapKey, _ := model.GetSystemSetting("map_amap_key")
		tencentKey, _ := model.GetSystemSetting("map_tencent_key")

		geocodeResult, err := utils.GeocodeAddress(strings.TrimSpace(req.Address), amapKey, tencentKey)
		if err == nil && geocodeResult.Success {
			addressData["latitude"] = geocodeResult.Latitude
			addressData["longitude"] = geocodeResult.Longitude
		}
		// 如果解析失败，不阻止保存，但记录日志
	}

	var address *model.Address
	// 判断是新增还是编辑
	if req.AddressID == nil || *req.AddressID == 0 {
		// 新增地址
		address, err = model.CreateAddress(user.ID, addressData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建地址失败: " + err.Error()})
			return
		}
	} else {
		// 更新地址
		err = model.UpdateAddress(*req.AddressID, user.ID, addressData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新地址失败: " + err.Error()})
			return
		}
		address, err = model.GetAddressByID(*req.AddressID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取地址信息失败: " + err.Error()})
			return
		}
	}

	// 统计地址数量，判断资料是否完善
	addressCount, err := model.CountAddressesByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "统计地址数量失败: " + err.Error()})
		return
	}

	// 如果地址数量>=1，标记资料已完善
	profileCompleted := addressCount >= 1

	// 如果资料已完善，且用户类型是unknown，自动设置为retail（零售用户）
	userType := user.UserType
	if profileCompleted && (userType == "" || userType == "unknown") {
		userType = "retail"
	}

	_, err = database.DB.Exec(`
		UPDATE mini_app_users 
		SET profile_completed = ?, user_type = ?, updated_at = NOW()
		WHERE id = ?
	`, profileCompleted, userType, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新资料完善状态失败: " + err.Error()})
		return
	}

	// 返回地址信息
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "地址保存成功",
		"data":    address,
	})
}

// UploadMiniAppUserAvatar 上传小程序用户头像
func UploadMiniAppUserAvatar(c *gin.Context) {
	// 从中间件获取用户信息
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	// 检查是否有文件上传
	file, headers, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片: " + err.Error()})
		return
	}
	defer file.Close()

	if headers.Size > 15*1024*1024 { // 限制文件大小为15MB
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过15MB"})
		return
	}

	// 检查文件类型
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
	}
	extension := ""
	for i := len(headers.Filename) - 1; i >= 0; i-- {
		if headers.Filename[i] == '.' {
			extension = headers.Filename[i:]
			break
		}
	}
	if !imageExtensions[strings.ToLower(extension)] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传JPG、PNG或GIF格式的图片"})
		return
	}

	// 上传图片到MinIO，用户头像存到users目录
	fileURL, err := utils.UploadFile("mini-user-avatar", c.Request, "users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败: " + err.Error()})
		return
	}

	// 写入数据库索引
	SaveImageIndex(fileURL, "users", headers.Filename, headers.Size, headers.Header.Get("Content-Type"))

	// 更新用户头像
	uniqueID := openID.(string)
	if err := model.UpdateMiniAppUserAvatar(uniqueID, fileURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户头像失败: " + err.Error()})
		return
	}

	// 返回更新后的用户信息
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "头像上传成功",
		"data": gin.H{
			"avatar": fileURL,
			"user":   user,
		},
	})
}

// UploadMiniAppUserAvatarByAdmin 管理员上传小程序用户头像
func UploadMiniAppUserAvatarByAdmin(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供用户ID"})
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	// 检查用户是否存在
	user, err := model.GetMiniAppUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	// 检查是否有文件上传
	file, headers, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片: " + err.Error()})
		return
	}
	defer file.Close()

	if headers.Size > 15*1024*1024 { // 限制文件大小为15MB
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过15MB"})
		return
	}

	// 检查文件类型
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
	}
	extension := ""
	for i := len(headers.Filename) - 1; i >= 0; i-- {
		if headers.Filename[i] == '.' {
			extension = headers.Filename[i:]
			break
		}
	}
	if !imageExtensions[strings.ToLower(extension)] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传JPG、PNG或GIF格式的图片"})
		return
	}

	// 上传图片到MinIO，用户头像存到users目录
	fileURL, err := utils.UploadFile("mini-user-avatar", c.Request, "users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败: " + err.Error()})
		return
	}

	// 写入数据库索引
	SaveImageIndex(fileURL, "users", headers.Filename, headers.Size, headers.Header.Get("Content-Type"))

	// 更新用户头像
	if err := model.UpdateMiniAppUserAvatar(user.UniqueID, fileURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户头像失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "头像上传成功",
		"data": gin.H{
			"avatar": fileURL,
		},
	})
}

// UploadAddressAvatar 上传地址头像（门头照片），只上传到MinIO并返回URL，不更新任何表
func UploadAddressAvatar(c *gin.Context) {
	// 从中间件获取用户信息（用于验证身份）
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}
	_ = openID // 仅用于验证，不实际使用

	// 检查是否有文件上传
	file, headers, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片: " + err.Error()})
		return
	}
	defer file.Close()

	if headers.Size > 15*1024*1024 { // 限制文件大小为15MB
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过15MB"})
		return
	}

	// 检查文件类型
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
	}
	extension := ""
	for i := len(headers.Filename) - 1; i >= 0; i-- {
		if headers.Filename[i] == '.' {
			extension = headers.Filename[i:]
			break
		}
	}
	if !imageExtensions[strings.ToLower(extension)] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传JPG、PNG或GIF格式的图片"})
		return
	}

	// 上传图片到MinIO，地址头像存到users目录
	fileURL, err := utils.UploadFile("mini-address-avatar", c.Request, "users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败: " + err.Error()})
		return
	}

	// 写入数据库索引
	SaveImageIndex(fileURL, "users", headers.Filename, headers.Size, headers.Header.Get("Content-Type"))

	// 只返回URL，不更新任何表
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "图片上传成功",
		"data": gin.H{
			"avatar":   fileURL,
			"imageUrl": fileURL, // 兼容前端可能使用的字段名
			"url":      fileURL, // 兼容更多字段名
		},
	})
}

// UploadAddressAvatarByAdmin 管理员上传地址头像（门头照片），只上传到MinIO并返回URL，不更新任何表
func UploadAddressAvatarByAdmin(c *gin.Context) {
	// 检查是否有文件上传
	file, headers, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片: " + err.Error()})
		return
	}
	defer file.Close()

	if headers.Size > 15*1024*1024 { // 限制文件大小为15MB
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过15MB"})
		return
	}

	// 检查文件类型
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
	}
	extension := ""
	for i := len(headers.Filename) - 1; i >= 0; i-- {
		if headers.Filename[i] == '.' {
			extension = headers.Filename[i:]
			break
		}
	}
	if !imageExtensions[strings.ToLower(extension)] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传JPG、PNG或GIF格式的图片"})
		return
	}

	// 上传图片到MinIO，地址头像存到users目录
	fileURL, err := utils.UploadFile("mini-address-avatar", c.Request, "users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败: " + err.Error()})
		return
	}

	// 写入数据库索引
	SaveImageIndex(fileURL, "users", headers.Filename, headers.Size, headers.Header.Get("Content-Type"))

	// 只返回URL，不更新任何表
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "图片上传成功",
		"data": gin.H{
			"avatar":   fileURL,
			"imageUrl": fileURL, // 兼容前端可能使用的字段名
			"url":      fileURL, // 兼容更多字段名
		},
	})
}

// GetMiniAppAddresses 获取用户的所有地址
func GetMiniAppAddresses(c *gin.Context) {
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	addresses, err := model.GetAddressesByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取地址列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    addresses,
	})
}

// GetMiniAppDefaultAddress 获取用户的默认地址
func GetMiniAppDefaultAddress(c *gin.Context) {
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	address, err := model.GetDefaultAddressByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取默认地址失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    address,
	})
}

// GetMiniAppInvoice 获取用户的发票抬头
func GetMiniAppInvoice(c *gin.Context) {
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	invoice, err := model.GetInvoiceByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取发票抬头失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    invoice,
	})
}

// SaveMiniAppInvoice 保存发票抬头
func SaveMiniAppInvoice(c *gin.Context) {
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	var req struct {
		InvoiceType    string `json:"invoice_type" binding:"required"`
		Title          string `json:"title" binding:"required"`
		TaxNumber      string `json:"tax_number"`
		CompanyAddress string `json:"company_address"`
		CompanyPhone   string `json:"company_phone"`
		BankName       string `json:"bank_name"`
		BankAccount    string `json:"bank_account"`
		IsDefault      bool   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "发票抬头不能为空"})
		return
	}

	// 如果是企业类型，纳税人识别号必填
	if req.InvoiceType == "company" && req.TaxNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "企业发票纳税人识别号不能为空"})
		return
	}

	invoiceData := map[string]interface{}{
		"invoice_type":    req.InvoiceType,
		"title":            req.Title,
		"tax_number":       req.TaxNumber,
		"company_address":  req.CompanyAddress,
		"company_phone":    req.CompanyPhone,
		"bank_name":        req.BankName,
		"bank_account":     req.BankAccount,
		"is_default":       req.IsDefault,
	}

	invoice, err := model.CreateOrUpdateInvoice(user.ID, invoiceData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存发票抬头失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "保存成功",
		"data":    invoice,
	})
}

// DeleteMiniAppAddress 删除地址
func DeleteMiniAppAddress(c *gin.Context) {
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	addressIDStr := c.Param("id")
	addressID, err := strconv.Atoi(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址ID无效"})
		return
	}

	err = model.DeleteAddress(addressID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除地址失败: " + err.Error()})
		return
	}

	// 重新统计地址数量，更新资料完善状态
	addressCount, err := model.CountAddressesByUserID(user.ID)
	if err == nil {
		profileCompleted := addressCount >= 1
		// 如果地址数量为0，且用户类型是retail，重置为unknown
		// 如果地址数量>=1，且用户类型是unknown，设置为retail
		userType := user.UserType
		if !profileCompleted && userType == "retail" {
			userType = "unknown"
		} else if profileCompleted && (userType == "" || userType == "unknown") {
			userType = "retail"
		}
		database.DB.Exec(`
			UPDATE mini_app_users 
			SET profile_completed = ?, user_type = ?, updated_at = NOW()
			WHERE id = ?
		`, profileCompleted, userType, user.ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// SetDefaultMiniAppAddress 设置默认地址
func SetDefaultMiniAppAddress(c *gin.Context) {
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	addressIDStr := c.Param("id")
	addressID, err := strconv.Atoi(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址ID无效"})
		return
	}

	err = model.SetDefaultAddress(addressID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "设置默认地址失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "设置成功",
	})
}

// UpdateMiniAppUserName 更新用户姓名
func UpdateMiniAppUserName(c *gin.Context) {
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 验证姓名长度
	if len(req.Name) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "姓名长度不能超过50个字符"})
		return
	}

	err := model.UpdateMiniAppUserName(uniqueID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新姓名失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// UpdateMiniAppUserPhone 更新用户电话
func UpdateMiniAppUserPhone(c *gin.Context) {
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 验证电话格式（简单验证，只检查长度）
	phone := strings.TrimSpace(req.Phone)
	if len(phone) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "电话长度不能超过20个字符"})
		return
	}

	err := model.UpdateMiniAppUserPhone(uniqueID, phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新电话失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// GeocodeAddress 地址解析接口（将地址文本转换为经纬度）
func GeocodeAddress(c *gin.Context) {
	type geocodeRequest struct {
		Address string `json:"address" binding:"required"`
	}

	var req geocodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 获取地图API Key
	amapKey, _ := model.GetSystemSetting("map_amap_key")
	tencentKey, _ := model.GetSystemSetting("map_tencent_key")

	result, err := utils.GeocodeAddress(req.Address, amapKey, tencentKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "地址解析失败: " + err.Error(),
		})
		return
	}

	if !result.Success {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": result.Message,
			"data":    result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "地址解析成功",
		"data":    result,
	})
}

// ReverseGeocode 逆地理编码接口（根据经纬度获取地址）
func ReverseGeocode(c *gin.Context) {
	type reverseGeocodeRequest struct {
		Longitude float64 `json:"longitude" binding:"required"`
		Latitude  float64 `json:"latitude" binding:"required"`
	}

	var req reverseGeocodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 获取地图API Key
	amapKey, _ := model.GetSystemSetting("map_amap_key")
	tencentKey, _ := model.GetSystemSetting("map_tencent_key")

	result, err := utils.ReverseGeocode(req.Longitude, req.Latitude, amapKey, tencentKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "逆地理编码失败: " + err.Error(),
		})
		return
	}

	if !result.Success {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": result.Message,
			"data":    result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "逆地理编码成功",
		"data":    result,
	})
}

// SearchPOI POI搜索接口（使用高德地图API搜索地址）
func SearchPOI(c *gin.Context) {
	// 添加panic恢复
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[SearchPOI] 发生panic: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": fmt.Sprintf("POI搜索发生错误: %v", r),
			})
		}
	}()

	type poiSearchRequest struct {
		Keyword  string  `json:"keyword" binding:"required"`  // 搜索关键词
		City     string  `json:"city,omitempty"`              // 城市（可选）
		Location *string `json:"location,omitempty"`          // 中心点坐标，格式"经度,纬度"（可选）
	}

	var req poiSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	log.Printf("[SearchPOI] 收到搜索请求: keyword=%s, city=%s, location=%v", req.Keyword, req.City, req.Location)

	// 获取地图API Key
	amapKey, _ := model.GetSystemSetting("map_amap_key")
	if amapKey == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "未配置高德地图API Key，无法进行POI搜索",
			"data":    utils.POISearchResponse{Success: false, Message: "未配置高德地图API Key"},
		})
		return
	}

	var location string
	if req.Location != nil {
		location = *req.Location
	}

	result, err := utils.SearchPOI(req.Keyword, req.City, location, amapKey)
	if err != nil {
		log.Printf("[SearchPOI] 搜索失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "POI搜索失败: " + err.Error(),
		})
		return
	}

	if !result.Success {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": result.Message,
			"data":    result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "搜索成功",
		"data":    result,
	})
}

// GetMiniAppReferralUsers 获取当前用户的拉新用户列表（小程序端）
func GetMiniAppReferralUsers(c *gin.Context) {
	// 从中间件设置的上下文中获取 openID
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	// 获取分页参数
	pageNum := parseQueryInt(c, "page_num", 1)
	pageSize := parseQueryInt(c, "page_size", 10)

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 获取拉新用户列表
	users, total, err := model.GetReferralUsersWithOrderStatus(user.ID, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取拉新用户列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  users,
			"total": total,
		},
	})
}

// GetMiniAppReferralStats 获取当前用户的拉新统计数据（小程序端）
func GetMiniAppReferralStats(c *gin.Context) {
	// 从中间件设置的上下文中获取 openID
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	uniqueID := openID.(string)
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	// 获取统计数据
	stats, err := model.GetReferralStats(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取统计数据失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// GetMiniAppReferralActivityInfo 获取拉新活动说明（小程序端）
func GetMiniAppReferralActivityInfo(c *gin.Context) {
	// 返回活动说明（可以从数据库或配置文件读取，这里先返回默认值）
	activityInfo := map[string]interface{}{
		"rules": []string{
			"分享小程序给好友，邀请他们注册成为新用户",
			"好友通过您的分享链接注册后，将显示在您的拉新列表中",
			"当好友完成首次下单后，您将获得相应奖励",
			"拉新奖励以实际活动规则为准",
		},
		"title": "分享有礼活动",
		"description": "邀请好友注册下单，即可获得丰厚奖励",
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    activityInfo,
	})
}