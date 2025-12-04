package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetAllCoupons 获取所有优惠券（后台管理）
func GetAllCoupons(c *gin.Context) {
	coupons, err := model.GetAllCoupons()
	if err != nil {
		internalErrorResponse(c, "获取优惠券列表失败: "+err.Error())
		return
	}
	// 确保返回空数组而不是 nil
	if coupons == nil {
		coupons = []model.CouponWithStats{}
	}
	successResponse(c, coupons, "")
}

// GetCouponByID 根据ID获取优惠券详情
func GetCouponByID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	coupon, err := model.GetCouponByID(id)
	if err != nil {
		internalErrorResponse(c, "获取优惠券失败: "+err.Error())
		return
	}
	if coupon == nil {
		notFoundResponse(c, "优惠券不存在")
		return
	}

	successResponse(c, coupon, "")
}

// CreateCoupon 创建优惠券
func CreateCoupon(c *gin.Context) {
	var req struct {
		Name          string  `json:"name" binding:"required"`
		Type          string  `json:"type" binding:"required,oneof=delivery_fee amount"`
		DiscountValue float64 `json:"discount_value" binding:"min=0"`
		MinAmount     float64 `json:"min_amount" binding:"min=0"`
		CategoryIDs   []int   `json:"category_ids"`
		TotalCount    int     `json:"total_count" binding:"min=0"`
		Status        int     `json:"status"`
		ValidFrom     string  `json:"valid_from" binding:"required"`
		ValidTo       string  `json:"valid_to" binding:"required"`
		Description   string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 解析日期时间（支持 "YYYY-MM-DD HH:mm:ss" 和 RFC3339 格式）
	parseTime := func(timeStr string) (time.Time, error) {
		// 尝试解析 "YYYY-MM-DD HH:mm:ss" 格式，使用本地时区
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local); err == nil {
			return t, nil
		}
		// 尝试解析 RFC3339 格式（带时区信息）
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			return t, nil
		}
		// 尝试解析 "YYYY-MM-DDTHH:mm:ss" 格式，使用本地时区
		if t, err := time.ParseInLocation("2006-01-02T15:04:05", timeStr, time.Local); err == nil {
			return t, nil
		}
		return time.Time{}, fmt.Errorf("无效的日期时间格式: %s", timeStr)
	}

	validFrom, err := parseTime(req.ValidFrom)
	if err != nil {
		badRequestResponse(c, "开始时间格式错误: "+err.Error())
		return
	}

	validTo, err := parseTime(req.ValidTo)
	if err != nil {
		badRequestResponse(c, "结束时间格式错误: "+err.Error())
		return
	}

	// 验证有效期
	if validTo.Before(validFrom) {
		badRequestResponse(c, "有效期结束时间必须晚于开始时间")
		return
	}

	// 验证配送费券的discount_value应该为0
	if req.Type == "delivery_fee" && req.DiscountValue != 0 {
		badRequestResponse(c, "配送费券的优惠值必须为0（表示全免）")
		return
	}

	// 验证金额券的discount_value必须大于0
	if req.Type == "amount" && req.DiscountValue <= 0 {
		badRequestResponse(c, "金额券的优惠值必须大于0")
		return
	}

	coupon := &model.Coupon{
		Name:          req.Name,
		Type:          req.Type,
		DiscountValue: req.DiscountValue,
		MinAmount:     req.MinAmount,
		CategoryIDs:   req.CategoryIDs,
		TotalCount:    req.TotalCount,
		Status:        req.Status,
		ValidFrom:     model.FromTime(validFrom),
		ValidTo:       model.FromTime(validTo),
		Description:   req.Description,
	}

	if coupon.Status == 0 {
		coupon.Status = 1 // 默认启用
	}

	if err := model.CreateCoupon(coupon); err != nil {
		internalErrorResponse(c, "创建优惠券失败: "+err.Error())
		return
	}

	successResponse(c, coupon, "创建成功")
}

// UpdateCoupon 更新优惠券
func UpdateCoupon(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var req struct {
		Name          string  `json:"name" binding:"required"`
		Type          string  `json:"type" binding:"required,oneof=delivery_fee amount"`
		DiscountValue float64 `json:"discount_value" binding:"min=0"`
		MinAmount     float64 `json:"min_amount" binding:"min=0"`
		CategoryIDs   []int   `json:"category_ids"`
		TotalCount    int     `json:"total_count" binding:"min=0"`
		Status        int     `json:"status"`
		ValidFrom     string  `json:"valid_from" binding:"required"`
		ValidTo       string  `json:"valid_to" binding:"required"`
		Description   string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 解析日期时间（支持 "YYYY-MM-DD HH:mm:ss" 和 RFC3339 格式）
	parseTime := func(timeStr string) (time.Time, error) {
		// 尝试解析 "YYYY-MM-DD HH:mm:ss" 格式，使用本地时区
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local); err == nil {
			return t, nil
		}
		// 尝试解析 RFC3339 格式（带时区信息）
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			return t, nil
		}
		// 尝试解析 "YYYY-MM-DDTHH:mm:ss" 格式，使用本地时区
		if t, err := time.ParseInLocation("2006-01-02T15:04:05", timeStr, time.Local); err == nil {
			return t, nil
		}
		return time.Time{}, fmt.Errorf("无效的日期时间格式: %s", timeStr)
	}

	validFrom, err := parseTime(req.ValidFrom)
	if err != nil {
		badRequestResponse(c, "开始时间格式错误: "+err.Error())
		return
	}

	validTo, err := parseTime(req.ValidTo)
	if err != nil {
		badRequestResponse(c, "结束时间格式错误: "+err.Error())
		return
	}

	// 验证有效期
	if validTo.Before(validFrom) {
		badRequestResponse(c, "有效期结束时间必须晚于开始时间")
		return
	}

	// 验证配送费券的discount_value应该为0
	if req.Type == "delivery_fee" && req.DiscountValue != 0 {
		badRequestResponse(c, "配送费券的优惠值必须为0（表示全免）")
		return
	}

	// 验证金额券的discount_value必须大于0
	if req.Type == "amount" && req.DiscountValue <= 0 {
		badRequestResponse(c, "金额券的优惠值必须大于0")
		return
	}

	// 检查优惠券是否存在
	existingCoupon, err := model.GetCouponByID(id)
	if err != nil {
		internalErrorResponse(c, "获取优惠券失败: "+err.Error())
		return
	}
	if existingCoupon == nil {
		notFoundResponse(c, "优惠券不存在")
		return
	}

	coupon := &model.Coupon{
		ID:            id,
		Name:          req.Name,
		Type:          req.Type,
		DiscountValue: req.DiscountValue,
		MinAmount:     req.MinAmount,
		CategoryIDs:   req.CategoryIDs,
		TotalCount:    req.TotalCount,
		Status:        req.Status,
		ValidFrom:     model.FromTime(validFrom),
		ValidTo:       model.FromTime(validTo),
		Description:   req.Description,
		UsedCount:     existingCoupon.UsedCount, // 保留已使用数量
	}

	if err := model.UpdateCoupon(coupon); err != nil {
		internalErrorResponse(c, "更新优惠券失败: "+err.Error())
		return
	}

	successResponse(c, coupon, "更新成功")
}

// DeleteCoupon 删除优惠券
func DeleteCoupon(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	// 检查优惠券是否存在
	coupon, err := model.GetCouponByID(id)
	if err != nil {
		internalErrorResponse(c, "获取优惠券失败: "+err.Error())
		return
	}
	if coupon == nil {
		notFoundResponse(c, "优惠券不存在")
		return
	}

	if err := model.DeleteCoupon(id); err != nil {
		internalErrorResponse(c, "删除优惠券失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// GetAvailableCoupons 获取可用优惠券列表（小程序用）
func GetAvailableCoupons(c *gin.Context) {
	openIDValue, exists := c.Get("openID")
	if !exists {
		unauthorizedResponse(c, "未登录")
		return
	}

	openID := openIDValue.(string)
	user, err := model.GetMiniAppUserByUniqueID(openID)
	if err != nil {
		internalErrorResponse(c, "获取用户信息失败: "+err.Error())
		return
	}
	if user == nil {
		unauthorizedResponse(c, "用户不存在")
		return
	}

	userID := user.ID

	// 获取订单金额和分类ID（从查询参数）
	orderAmountStr := c.Query("order_amount")
	orderAmount := 0.0
	if orderAmountStr != "" {
		var err error
		orderAmount, err = strconv.ParseFloat(orderAmountStr, 64)
		if err != nil {
			badRequestResponse(c, "订单金额格式错误")
			return
		}
	}

	categoryIDsStr := c.Query("category_ids")
	var categoryIDs []int
	if categoryIDsStr != "" {
		// 解析分类ID（逗号分隔）
		idStrs := strings.Split(categoryIDsStr, ",")
		for _, idStr := range idStrs {
			if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				categoryIDs = append(categoryIDs, id)
			}
		}
	}

	coupons, err := model.GetAvailableCouponsForUser(userID, orderAmount, categoryIDs)
	if err != nil {
		internalErrorResponse(c, "获取可用优惠券失败: "+err.Error())
		return
	}

	successResponse(c, coupons, "")
}

// GetUserCoupons 获取用户的优惠券列表
func GetUserCoupons(c *gin.Context) {
	openIDValue, exists := c.Get("openID")
	if !exists {
		unauthorizedResponse(c, "未登录")
		return
	}

	openID := openIDValue.(string)
	user, err := model.GetMiniAppUserByUniqueID(openID)
	if err != nil {
		internalErrorResponse(c, "获取用户信息失败: "+err.Error())
		return
	}
	if user == nil {
		unauthorizedResponse(c, "用户不存在")
		return
	}

	userID := user.ID

	userCoupons, err := model.GetUserCoupons(userID)
	if err != nil {
		internalErrorResponse(c, "获取用户优惠券失败: "+err.Error())
		return
	}

	successResponse(c, userCoupons, "")
}

// GetAdminUserCoupons 管理员获取用户的优惠券列表
func GetAdminUserCoupons(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		badRequestResponse(c, "请提供用户ID")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		badRequestResponse(c, "用户ID格式错误")
		return
	}

	// 验证用户是否存在
	user, err := model.GetMiniAppUserByID(userID)
	if err != nil {
		internalErrorResponse(c, "获取用户信息失败: "+err.Error())
		return
	}
	if user == nil {
		notFoundResponse(c, "用户不存在")
		return
	}

	// 获取用户优惠券列表
	userCoupons, err := model.GetUserCoupons(userID)
	if err != nil {
		internalErrorResponse(c, "获取用户优惠券失败: "+err.Error())
		return
	}

	successResponse(c, userCoupons, "")
}

// IssueCouponToUser 发放优惠券给用户（管理员操作）
func IssueCouponToUser(c *gin.Context) {
	var req struct {
		UserID    int    `json:"user_id" binding:"required"`
		CouponID  int    `json:"coupon_id" binding:"required"`
		Quantity  int    `json:"quantity"`   // 发放数量，默认为1
		ExpiresIn int    `json:"expires_in"` // 有效期天数，0表示不限制
		ExpiresAt string `json:"expires_at"` // 有效期截止时间（可选，优先级高于expires_in）
		Reason    string `json:"reason"`     // 发放原因（潜在客户、优质客户等），目前仅用于记录，不参与计算
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认数量为1
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	// 处理有效期
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		// 如果提供了 expires_at，解析它
		parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", req.ExpiresAt, time.Local)
		if err != nil {
			// 尝试其他格式
			parsedTime, err = time.ParseInLocation("2006-01-02", req.ExpiresAt, time.Local)
			if err != nil {
				badRequestResponse(c, "有效期时间格式错误")
				return
			}
		}
		expiresAt = &parsedTime
	} else if req.ExpiresIn > 0 {
		// 如果提供了 expires_in，计算到期时间
		calculatedTime := time.Now().AddDate(0, 0, req.ExpiresIn)
		expiresAt = &calculatedTime
	}

	// 验证用户是否存在
	user, err := model.GetMiniAppUserByID(req.UserID)
	if err != nil {
		internalErrorResponse(c, "获取用户信息失败: "+err.Error())
		return
	}
	if user == nil {
		badRequestResponse(c, "用户不存在")
		return
	}

	// 验证优惠券是否存在
	coupon, err := model.GetCouponByID(req.CouponID)
	if err != nil {
		internalErrorResponse(c, "获取优惠券信息失败: "+err.Error())
		return
	}
	if coupon == nil {
		badRequestResponse(c, "优惠券不存在")
		return
	}

	// 发放优惠券
	if err := model.IssueCouponToUser(req.UserID, req.CouponID, req.Quantity, expiresAt); err != nil {
		internalErrorResponse(c, "发放优惠券失败: "+err.Error())
		return
	}

	// 记录发放日志（区分管理员 / 员工）
	operatorType := "admin"
	operatorID := 0
	operatorName := ""

	// 如果是员工端调用，该中间件已经把 employee 放到 context 里
	if emp, ok := getEmployeeFromContext(c); ok {
		operatorType = "employee"
		operatorID = emp.ID
		operatorName = emp.Name
	}

	log := &model.CouponIssueLog{
		UserID:       user.ID,
		CouponID:     coupon.ID,
		CouponName:   coupon.Name,
		Quantity:     req.Quantity,
		Reason:       req.Reason,
		OperatorType: operatorType,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		ExpiresAt:    expiresAt,
	}
	_ = model.CreateCouponIssueLog(log)

	successResponse(c, nil, fmt.Sprintf("成功发放 %d 张优惠券", req.Quantity))
}

// GetCouponIssueLogs 获取优惠券发放记录列表（后台管理）
func GetCouponIssueLogs(c *gin.Context) {
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 20)
	keyword := strings.TrimSpace(c.Query("keyword"))
	couponID := parseQueryInt(c, "couponId", 0)

	logs, total, err := model.GetCouponIssueLogs(pageNum, pageSize, keyword, couponID)
	if err != nil {
		internalErrorResponse(c, "获取优惠券发放记录失败: "+err.Error())
		return
	}

	if logs == nil {
		logs = []model.CouponIssueLog{}
	}

	data := gin.H{
		"list":     logs,
		"total":    total,
		"pageNum":  pageNum,
		"pageSize": pageSize,
	}

	successResponse(c, data, "")
}

// GetCouponUsageLogs 获取优惠券使用记录列表（后台管理）
func GetCouponUsageLogs(c *gin.Context) {
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 20)
	keyword := strings.TrimSpace(c.Query("keyword"))
	couponID := parseQueryInt(c, "couponId", 0)

	logs, total, err := model.GetCouponUsageLogs(pageNum, pageSize, keyword, couponID)
	if err != nil {
		internalErrorResponse(c, "获取优惠券使用记录失败: "+err.Error())
		return
	}

	if logs == nil {
		logs = []model.CouponUsageLog{}
	}

	data := gin.H{
		"list":     logs,
		"total":    total,
		"pageNum":  pageNum,
		"pageSize": pageSize,
	}

	successResponse(c, data, "")
}