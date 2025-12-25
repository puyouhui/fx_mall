package api

import (
	"net/http"
	"strconv"
	"strings"

	"go_backend/internal/model"
	"go_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// extractBearerToken 从 Authorization 头中提取 token
func extractBearerTokenForSupplier(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	if len(header) > 7 && strings.EqualFold(header[0:7], "Bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return header
}

// CreateSupplierApplication 创建供应商合作申请（小程序端）
func CreateSupplierApplication(c *gin.Context) {
	// 获取用户ID（如果已登录）
	var userID *int
	
	// 尝试从请求头获取 token（支持可选认证）
	token := extractBearerTokenForSupplier(c.GetHeader("Authorization"))
	if token != "" {
		// 尝试解析 token 获取用户信息
		claims, err := utils.ParseMiniAppToken(token)
		if err == nil {
			user, err := model.GetMiniAppUserByUniqueID(claims.OpenID)
			if err == nil && user != nil {
				userID = &user.ID
			}
		}
	}

	var req struct {
		CompanyName       string `json:"company_name" binding:"required"`
		ContactName       string `json:"contact_name" binding:"required"`
		ContactPhone      string `json:"contact_phone" binding:"required"`
		Email             string `json:"email"`
		Address           string `json:"address"`
		MainCategory      string `json:"main_category" binding:"required"`
		CompanyIntro      string `json:"company_intro"`
		CooperationIntent string `json:"cooperation_intent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	application, err := model.CreateSupplierApplication(
		userID,
		req.CompanyName,
		req.ContactName,
		req.ContactPhone,
		req.Email,
		req.Address,
		req.MainCategory,
		req.CompanyIntro,
		req.CooperationIntent,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "提交成功",
		"data": application,
	})
}

// GetUserSupplierApplications 获取用户的申请列表（小程序端）
func GetUserSupplierApplications(c *gin.Context) {
	user, ok := getMiniUserFromContextForProductRequest(c)
	if !ok {
		return
	}

	applications, err := model.GetSupplierApplicationsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": applications,
	})
}

// GetAllSupplierApplications 获取所有申请列表（管理员）
func GetAllSupplierApplications(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	applications, total, err := model.GetAllSupplierApplications(pageNum, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": applications,
		"total": total,
		"pageNum": pageNum,
		"pageSize": pageSize,
	})
}

// UpdateSupplierApplicationStatus 更新申请状态（管理员）
func UpdateSupplierApplicationStatus(c *gin.Context) {
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
		"pending":  true,
		"approved": true,
		"rejected": true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的状态值"})
		return
	}

	if err := model.UpdateSupplierApplicationStatus(id, req.Status, req.AdminRemark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "更新成功",
	})
}

