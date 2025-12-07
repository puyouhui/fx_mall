package api

import (
	"net/http"
	"strconv"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetDeliveryFeeCalculation 获取配送费计算结果（管理员可见利润分成）
func GetDeliveryFeeCalculation(c *gin.Context) {
	// 检查是否为管理员（通过中间件验证，这里只需要检查是否已认证）
	// 管理员认证由 AdminAuthMiddleware 处理

	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	// 创建计算器
	calculator, err := model.NewDeliveryFeeCalculator(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建计算器失败: " + err.Error()})
		return
	}

	// 计算配送费（管理员可见利润分成）
	result, err := calculator.Calculate(true)
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

// GetDeliveryFeeCalculationForRider 获取配送费计算结果（配送员视图）
// 配送员实际所得包含利润分成，但不显示利润分成明细
func GetDeliveryFeeCalculationForRider(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	if !employee.IsDelivery {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您不是配送员，无权访问此功能"})
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	// 创建计算器
	calculator, err := model.NewDeliveryFeeCalculator(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建计算器失败: " + err.Error()})
		return
	}

	// 计算配送费（配送员视图：不显示利润分成明细，但金额已包含利润分成）
	result, err := calculator.Calculate(false)
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

// GetDeliveryFeeSettings 获取配送费基础设置
func GetDeliveryFeeSettings(c *gin.Context) {
	setting, err := model.GetDeliveryFeeSetting()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配送费设置失败: " + err.Error()})
		return
	}

	if setting == nil {
		// 如果没有设置，返回默认值
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"id":                      0,
				"base_fee":                0,
				"free_shipping_threshold": 0,
				"description":             "",
			},
			"message": "获取成功",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    setting,
		"message": "获取成功",
	})
}

// UpdateDeliveryFeeSettings 更新配送费基础设置
func UpdateDeliveryFeeSettings(c *gin.Context) {
	var req struct {
		BaseFee               float64 `json:"base_fee" binding:"required"`
		FreeShippingThreshold float64 `json:"free_shipping_threshold" binding:"required"`
		Description           string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 获取现有设置
	existingSetting, err := model.GetDeliveryFeeSetting()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配送费设置失败: " + err.Error()})
		return
	}

	setting := &model.DeliveryFeeSetting{
		BaseFee:               req.BaseFee,
		FreeShippingThreshold: req.FreeShippingThreshold,
		Description:           req.Description,
	}

	if existingSetting != nil {
		setting.ID = existingSetting.ID
	}

	err = model.UpsertDeliveryFeeSetting(setting)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新配送费设置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    setting,
		"message": "更新成功",
	})
}

// ListDeliveryFeeExclusions 获取配送费排除项列表
func ListDeliveryFeeExclusions(c *gin.Context) {
	exclusions, err := model.DeliveryFeeExclusionList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取排除项列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    exclusions,
		"message": "获取成功",
	})
}

// CreateDeliveryFeeExclusion 创建配送费排除项
func CreateDeliveryFeeExclusion(c *gin.Context) {
	var req struct {
		ItemType           string `json:"item_type" binding:"required"`
		TargetID           int    `json:"target_id" binding:"required"`
		MinQuantityForFree *int   `json:"min_quantity_for_free"`
		Remark             string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 验证 item_type
	if req.ItemType != "category" && req.ItemType != "product" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "item_type 必须是 'category' 或 'product'"})
		return
	}

	exclusion := &model.DeliveryFeeExclusion{
		ItemType:           req.ItemType,
		TargetID:           req.TargetID,
		MinQuantityForFree: req.MinQuantityForFree,
		Remark:             req.Remark,
	}

	err := model.CreateDeliveryFeeExclusion(exclusion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建排除项失败: " + err.Error()})
		return
	}

	// 重新获取完整的排除项信息（包含名称等）
	exclusion, err = model.GetDeliveryFeeExclusionByID(exclusion.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取排除项详情失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    exclusion,
		"message": "创建成功",
	})
}

// UpdateDeliveryFeeExclusion 更新配送费排除项
func UpdateDeliveryFeeExclusion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "排除项ID格式错误"})
		return
	}

	var req struct {
		MinQuantityForFree *int   `json:"min_quantity_for_free"`
		Remark             string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 获取现有排除项
	exclusion, err := model.GetDeliveryFeeExclusionByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取排除项失败: " + err.Error()})
		return
	}
	if exclusion == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "排除项不存在"})
		return
	}

	// 更新字段
	if req.MinQuantityForFree != nil {
		exclusion.MinQuantityForFree = req.MinQuantityForFree
	}
	if req.Remark != "" {
		exclusion.Remark = req.Remark
	}

	err = model.UpdateDeliveryFeeExclusion(exclusion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新排除项失败: " + err.Error()})
		return
	}

	// 重新获取完整的排除项信息
	exclusion, err = model.GetDeliveryFeeExclusionByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取排除项详情失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    exclusion,
		"message": "更新成功",
	})
}

// DeleteDeliveryFeeExclusion 删除配送费排除项
func DeleteDeliveryFeeExclusion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "排除项ID格式错误"})
		return
	}

	// 检查排除项是否存在
	exclusion, err := model.GetDeliveryFeeExclusionByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取排除项失败: " + err.Error()})
		return
	}
	if exclusion == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "排除项不存在"})
		return
	}

	err = model.DeleteDeliveryFeeExclusion(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除排除项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}
