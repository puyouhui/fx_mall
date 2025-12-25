package api

import (
	"net/http"
	"strconv"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// getMiniUserFromContext 从上下文获取小程序用户信息（复用 purchase_list.go 中的函数）
func getMiniUserFromContextForProductRequest(c *gin.Context) (*model.MiniAppUser, bool) {
	openIDValue, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return nil, false
	}

	openID := openIDValue.(string)
	user, err := model.GetMiniAppUserByUniqueID(openID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return nil, false
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return nil, false
	}
	return user, true
}

// CreateProductRequest 创建新品需求（小程序端）
func CreateProductRequest(c *gin.Context) {
	user, ok := getMiniUserFromContextForProductRequest(c)
	if !ok {
		return
	}

	var req struct {
		ProductName     string `json:"product_name" binding:"required"`
		Brand           string `json:"brand"`
		MonthlyQuantity int    `json:"monthly_quantity"`
		Description     string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	productRequest, err := model.CreateProductRequest(
		user.ID,
		req.ProductName,
		req.Brand,
		req.MonthlyQuantity,
		req.Description,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "提交成功",
		"data": productRequest,
	})
}

// GetUserProductRequests 获取用户的新品需求列表（小程序端）
func GetUserProductRequests(c *gin.Context) {
	user, ok := getMiniUserFromContextForProductRequest(c)
	if !ok {
		return
	}

	requests, err := model.GetProductRequestsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": requests,
	})
}

// GetAllProductRequests 获取所有新品需求列表（管理员）
func GetAllProductRequests(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	requests, total, err := model.GetAllProductRequests(pageNum, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": requests,
		"total": total,
		"pageNum": pageNum,
		"pageSize": pageSize,
	})
}

// UpdateProductRequestStatus 更新新品需求状态（管理员）
func UpdateProductRequestStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var req struct {
		Status      string `json:"status" binding:"required"`
		AdminRemark string `json:"admin_remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证状态值
	validStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
		"completed":  true,
		"rejected":   true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的状态值"})
		return
	}

	if err := model.UpdateProductRequestStatus(id, req.Status, req.AdminRemark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "更新成功",
	})
}

