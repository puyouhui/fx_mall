package api

import (
	"net/http"
	"strconv"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// SubmitPaymentVerificationRequest 销售员提交收款申请
func SubmitPaymentVerificationRequest(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	var req struct {
		OrderID       int    `json:"order_id" binding:"required"`
		RequestReason string `json:"request_reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 创建收款审核申请
	verificationReq, err := model.CreatePaymentVerificationRequest(req.OrderID, employee.EmployeeCode, req.RequestReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交收款申请失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": verificationReq,
		"message": "提交成功，等待审核",
	})
}

// GetPaymentVerificationRequests 管理员获取收款审核列表
func GetPaymentVerificationRequests(c *gin.Context) {
	// 验证管理员权限
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	_, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	status := c.Query("status") // pending, approved, rejected

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	requests, total, err := model.GetPaymentVerificationRequests(pageNum, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取收款审核列表失败: " + err.Error()})
		return
	}

	// 转换为map格式返回
	requestsList := make([]map[string]interface{}, 0, len(requests))
	for _, req := range requests {
		reqMap := map[string]interface{}{
			"id":                 req.ID,
			"order_id":           req.OrderID,
			"order_number":       req.OrderNumber,
			"sales_employee_code": req.SalesEmployeeCode,
			"sales_employee_name": req.SalesEmployeeName,
			"customer_id":        req.CustomerID,
			"customer_name":      req.CustomerName,
			"order_amount":       req.OrderAmount,
			"request_reason":     req.RequestReason,
			"status":             req.Status,
			"admin_id":           req.AdminID,
			"admin_name":         req.AdminName,
			"reviewed_at":         req.ReviewedAt,
			"review_remark":      req.ReviewRemark,
			"created_at":         req.CreatedAt,
			"updated_at":         req.UpdatedAt,
		}
		requestsList = append(requestsList, reqMap)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  requestsList,
			"total": total,
		},
		"message": "获取成功",
	})
}

// ReviewPaymentVerificationRequest 管理员审核收款申请
func ReviewPaymentVerificationRequest(c *gin.Context) {
	// 验证管理员权限
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	adminID, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的管理员信息"})
		return
	}

	// 获取管理员姓名
	adminNameInterface, exists := c.Get("adminName")
	adminName := ""
	if exists {
		if name, ok := adminNameInterface.(string); ok {
			adminName = name
		}
	}

	var req struct {
		RequestID    int    `json:"request_id" binding:"required"`
		Approved     *bool  `json:"approved" binding:"required"`
		ReviewRemark string `json:"review_remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if req.Approved == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "approved 参数不能为空"})
		return
	}

	// 审核收款申请
	approved := *req.Approved
	err := model.ReviewPaymentVerificationRequest(req.RequestID, adminID, adminName, approved, req.ReviewRemark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "审核失败: " + err.Error()})
		return
	}

	message := "已拒绝"
	if approved {
		message = "审核通过，订单已标记为已收款"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": message,
	})
}

// GetPaymentVerificationRequestByOrderID 根据订单ID获取收款申请状态（销售员）
func GetPaymentVerificationRequestByOrderID(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	if !employee.IsSales {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是销售员，无权访问此功能"})
		return
	}

	orderIDStr := c.Param("orderId")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的订单ID"})
		return
	}

	req, err := model.GetPendingPaymentVerificationByOrderID(orderID)
	if err != nil {
		// 没有待审核的申请，返回null
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": nil,
			"message": "获取成功",
		})
		return
	}

	reqMap := map[string]interface{}{
		"id":                 req.ID,
		"order_id":           req.OrderID,
		"order_number":       req.OrderNumber,
		"sales_employee_code": req.SalesEmployeeCode,
		"sales_employee_name": req.SalesEmployeeName,
		"customer_id":        req.CustomerID,
		"customer_name":      req.CustomerName,
		"order_amount":       req.OrderAmount,
		"request_reason":     req.RequestReason,
		"status":             req.Status,
		"admin_id":           req.AdminID,
		"admin_name":         req.AdminName,
		"reviewed_at":         req.ReviewedAt,
		"review_remark":      req.ReviewRemark,
		"created_at":         req.CreatedAt,
		"updated_at":         req.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": reqMap,
		"message": "获取成功",
	})
}

