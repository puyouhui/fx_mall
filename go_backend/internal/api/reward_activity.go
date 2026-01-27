package api

import (
	"net/http"
	"strconv"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetRewardActivities 获取奖励活动列表
func GetRewardActivities(c *gin.Context) {
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

	// 获取筛选参数
	activityType := c.Query("activity_type") // referral 或 new_customer

	// 获取活动列表
	activities, total, err := model.GetRewardActivities(pageNum, pageSize, activityType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取活动列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":      activities,
			"total":     total,
			"page_num":  pageNum,
			"page_size": pageSize,
		},
	})
}

// GetRewardActivity 获取单个奖励活动
func GetRewardActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "活动ID格式错误"})
		return
	}

	activity, err := model.GetRewardActivityByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取活动失败: " + err.Error()})
		return
	}

	if activity == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "活动不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    activity,
	})
}

// CreateRewardActivity 创建奖励活动
func CreateRewardActivity(c *gin.Context) {
	var req struct {
		ActivityName string   `json:"activity_name" binding:"required"`
		ActivityType string   `json:"activity_type" binding:"required,oneof=referral new_customer"`
		IsEnabled    bool     `json:"is_enabled"`
		RewardType   string   `json:"reward_type" binding:"required,oneof=points coupon amount"`
		RewardValue  float64  `json:"reward_value" binding:"min=0"`
		CouponIDs    []int    `json:"coupon_ids,omitempty"`
		Description  string   `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 如果创建时设置为启用，检查是否已存在同类型启用中的活动，避免重复导致系统错误
	if req.IsEnabled {
		existingActivity, err := model.GetEnabledRewardActivityByType(req.ActivityType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "检查活动唯一性失败: " + err.Error()})
			return
		}
		if existingActivity != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "已存在同类型且启用中的活动，请先停用或修改原有活动后再创建",
			})
			return
		}
	}

	// 如果奖励类型是coupon，必须提供至少一个优惠券ID
	if req.RewardType == "coupon" {
		if len(req.CouponIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "优惠券类型奖励必须至少选择一张优惠券"})
			return
		}
		// 验证每个优惠券是否存在
		for _, cid := range req.CouponIDs {
			if cid <= 0 {
				continue
			}
			coupon, err := model.GetCouponByID(cid)
			if err != nil || coupon == nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "优惠券不存在或已失效"})
				return
			}
		}
	}

	activity := &model.RewardActivity{
		ActivityName: req.ActivityName,
		ActivityType: req.ActivityType,
		IsEnabled:    req.IsEnabled,
		RewardType:   req.RewardType,
		RewardValue:  req.RewardValue,
		CouponIDs:    req.CouponIDs,
		Description:  req.Description,
	}

	if err := model.CreateRewardActivity(activity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建活动失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    activity,
	})
}

// UpdateRewardActivity 更新奖励活动
func UpdateRewardActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "活动ID格式错误"})
		return
	}

	var req struct {
		ActivityName string   `json:"activity_name" binding:"required"`
		ActivityType string   `json:"activity_type" binding:"required,oneof=referral new_customer"`
		IsEnabled    bool     `json:"is_enabled"`
		RewardType   string   `json:"reward_type" binding:"required,oneof=points coupon amount"`
		RewardValue  float64  `json:"reward_value" binding:"min=0"`
		CouponIDs    []int    `json:"coupon_ids,omitempty"`
		Description  string   `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 检查活动是否存在
	existingActivity, err := model.GetRewardActivityByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取活动失败: " + err.Error()})
		return
	}
	if existingActivity == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "活动不存在"})
		return
	}

	// 如果奖励类型是coupon，必须提供至少一个优惠券ID
	if req.RewardType == "coupon" {
		if len(req.CouponIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "优惠券类型奖励必须至少选择一张优惠券"})
			return
		}
		// 验证每个优惠券是否存在
		for _, cid := range req.CouponIDs {
			if cid <= 0 {
				continue
			}
			coupon, err := model.GetCouponByID(cid)
			if err != nil || coupon == nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "优惠券不存在或已失效"})
				return
			}
		}
	}

	activity := &model.RewardActivity{
		ID:           id,
		ActivityName: req.ActivityName,
		ActivityType: req.ActivityType,
		IsEnabled:    req.IsEnabled,
		RewardType:   req.RewardType,
		RewardValue:  req.RewardValue,
		CouponIDs:    req.CouponIDs,
		Description:  req.Description,
	}

	if err := model.UpdateRewardActivity(activity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新活动失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    activity,
	})
}

// DeleteRewardActivity 删除奖励活动
func DeleteRewardActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "活动ID格式错误"})
		return
	}

	// 检查活动是否存在
	activity, err := model.GetRewardActivityByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取活动失败: " + err.Error()})
		return
	}
	if activity == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "活动不存在"})
		return
	}

	if err := model.DeleteRewardActivity(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除活动失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}
