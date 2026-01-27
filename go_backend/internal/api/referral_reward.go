package api

import (
	"net/http"
	"strconv"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetReferralRewardConfig 获取推荐奖励活动配置
func GetReferralRewardConfig(c *gin.Context) {
	config, err := model.GetReferralRewardConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配置失败: " + err.Error()})
		return
	}

	if config == nil {
		// 如果不存在，返回默认配置
		config = &model.ReferralRewardConfig{
			IsEnabled:   false,
			RewardType:  "points",
			RewardValue: 0,
			Description: "老用户推荐新用户首次下单奖励活动",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    config,
	})
}

// UpdateReferralRewardConfig 更新推荐奖励活动配置（支持创建和更新）
func UpdateReferralRewardConfig(c *gin.Context) {
	var req struct {
		ID          int     `json:"id"` // ID 可选，如果为0或不存在则创建新配置
		IsEnabled   bool   `json:"is_enabled"`
		RewardType  string `json:"reward_type" binding:"required,oneof=points coupon amount"`
		RewardValue float64 `json:"reward_value" binding:"required,min=0"`
		CouponID    *int   `json:"coupon_id,omitempty"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 如果奖励类型是coupon，必须提供coupon_id
	if req.RewardType == "coupon" && req.CouponID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "优惠券类型奖励必须提供优惠券ID"})
		return
	}

	// 验证优惠券是否存在（如果是coupon类型）
	if req.RewardType == "coupon" && req.CouponID != nil {
		coupon, err := model.GetCouponByID(*req.CouponID)
		if err != nil || coupon == nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "优惠券不存在"})
			return
		}
	}

	config := &model.ReferralRewardConfig{
		ID:          req.ID,
		IsEnabled:   req.IsEnabled,
		RewardType:  req.RewardType,
		RewardValue: req.RewardValue,
		CouponID:    req.CouponID,
		Description: req.Description,
	}

	// 如果 ID 为 0 或不存在，先检查是否有配置，没有则创建，有则更新
	if req.ID == 0 {
		existingConfig, err := model.GetReferralRewardConfig()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配置失败: " + err.Error()})
			return
		}
		
		if existingConfig == nil {
			// 创建新配置
			if err := model.CreateReferralRewardConfig(config); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建配置失败: " + err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "创建成功",
			})
			return
		} else {
			// 使用现有配置的 ID 进行更新
			config.ID = existingConfig.ID
		}
	}

	// 更新配置
	if err := model.UpdateReferralRewardConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新配置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// GetReferralRewards 获取推荐奖励记录列表
func GetReferralRewards(c *gin.Context) {
	pageNumStr := c.DefaultQuery("page_num", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	referrerIDStr := c.Query("referrer_id")
	newUserIDStr := c.Query("new_user_id")
	status := c.Query("status")

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var referrerID, newUserID *int
	if referrerIDStr != "" {
		id, err := strconv.Atoi(referrerIDStr)
		if err == nil && id > 0 {
			referrerID = &id
		}
	}

	if newUserIDStr != "" {
		id, err := strconv.Atoi(newUserIDStr)
		if err == nil && id > 0 {
			newUserID = &id
		}
	}

	rewards, total, err := model.GetReferralRewards(pageNum, pageSize, referrerID, newUserID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取奖励记录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":      rewards,
			"total":     total,
			"page_num":  pageNum,
			"page_size": pageSize,
		},
	})
}

