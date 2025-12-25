package api

import (
	"net/http"
	"strconv"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// CreatePriceFeedback 创建价格反馈（小程序端）
func CreatePriceFeedback(c *gin.Context) {
	// 获取用户ID（如果已登录）
	var userID *int
	openIDValue, exists := c.Get("openID")
	if exists {
		openID := openIDValue.(string)
		user, err := model.GetMiniAppUserByUniqueID(openID)
		if err == nil && user != nil {
			userID = &user.ID
		}
	}

	var req struct {
		ProductID        int      `json:"product_id" binding:"required"`
		ProductName      string   `json:"product_name" binding:"required"`
		PlatformPriceMin float64  `json:"platform_price_min" binding:"required"`
		PlatformPriceMax float64  `json:"platform_price_max" binding:"required"`
		CompetitorPrice  float64  `json:"competitor_price" binding:"required"`
		Images           []string `json:"images"`
		Remark           string   `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证价格
	if req.CompetitorPrice <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "竞争对手价格必须大于0"})
		return
	}

	// 限制图片数量
	if len(req.Images) > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "最多只能上传3张图片"})
		return
	}

	// 验证价格范围
	if req.PlatformPriceMin < 0 || req.PlatformPriceMax < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "平台价格不能为负数"})
		return
	}
	if req.PlatformPriceMin > req.PlatformPriceMax {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "最低价格不能大于最高价格"})
		return
	}

	feedback, err := model.CreatePriceFeedback(
		userID,
		req.ProductID,
		req.ProductName,
		req.PlatformPriceMin,
		req.PlatformPriceMax,
		req.CompetitorPrice,
		req.Images,
		req.Remark,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "提交成功",
		"data":    feedback,
	})
}

// GetAllPriceFeedbacks 获取所有价格反馈列表（管理员）
func GetAllPriceFeedbacks(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	feedbacks, total, err := model.GetAllPriceFeedbacks(pageNum, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    feedbacks,
		"total":   total,
		"pageNum": pageNum,
		"pageSize": pageSize,
	})
}

// UpdatePriceFeedbackStatus 更新价格反馈状态（管理员）
func UpdatePriceFeedbackStatus(c *gin.Context) {
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
		"pending":   true,
		"processed": true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的状态值"})
		return
	}

	if err := model.UpdatePriceFeedbackStatus(id, req.Status, req.AdminRemark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

