package api

import (
	"net/http"
	"strings"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetDeliveryFeeSettings 获取配送费基础设置
func GetDeliveryFeeSettings(c *gin.Context) {
	setting, err := model.GetDeliveryFeeSetting()
	if err != nil {
		internalErrorResponse(c, "获取配送费设置失败: "+err.Error())
		return
	}

	if setting == nil {
		setting = &model.DeliveryFeeSetting{
			BaseFee:               0,
			FreeShippingThreshold: 0,
			Description:           "",
		}
	}

	successResponse(c, setting, "")
}

// UpdateDeliveryFeeSettings 更新配送费设置
func UpdateDeliveryFeeSettings(c *gin.Context) {
	var req struct {
		BaseFee               float64 `json:"base_fee"`
		FreeShippingThreshold float64 `json:"free_shipping_threshold"`
		Description           string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if req.BaseFee < 0 {
		badRequestResponse(c, "基础配送费不能为负数")
		return
	}
	if req.FreeShippingThreshold < 0 {
		badRequestResponse(c, "免配送费阈值不能为负数")
		return
	}

	description := strings.TrimSpace(req.Description)

	setting, err := model.GetDeliveryFeeSetting()
	if err != nil {
		internalErrorResponse(c, "获取配送费设置失败: "+err.Error())
		return
	}
	if setting == nil {
		setting = &model.DeliveryFeeSetting{}
	}
	setting.BaseFee = req.BaseFee
	setting.FreeShippingThreshold = req.FreeShippingThreshold
	setting.Description = description

	if err := model.UpsertDeliveryFeeSetting(setting); err != nil {
		internalErrorResponse(c, "保存配送费设置失败: "+err.Error())
		return
	}

	successResponse(c, setting, "保存成功")
}

// ListDeliveryFeeExclusions 获取排除项列表
func ListDeliveryFeeExclusions(c *gin.Context) {
	exclusions, err := model.DeliveryFeeExclusionList()
	if err != nil {
		internalErrorResponse(c, "获取排除项失败: "+err.Error())
		return
	}

	successResponse(c, exclusions, "")
}

// CreateDeliveryFeeExclusion 创建排除项
func CreateDeliveryFeeExclusion(c *gin.Context) {
	var req struct {
		ItemType           string `json:"item_type" binding:"required"`
		TargetID           int    `json:"target_id" binding:"required"`
		MinQuantityForFree *int   `json:"min_quantity_for_free"`
		Remark             string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	req.ItemType = strings.ToLower(strings.TrimSpace(req.ItemType))
	if req.ItemType != "category" && req.ItemType != "product" {
		badRequestResponse(c, "item_type 只能是 category 或 product")
		return
	}
	if req.TargetID <= 0 {
		badRequestResponse(c, "target_id 无效")
		return
	}

	if req.ItemType == "category" {
		category, err := model.GetCategoryByID(req.TargetID)
		if err != nil {
			internalErrorResponse(c, "校验分类失败: "+err.Error())
			return
		}
		if category == nil {
			badRequestResponse(c, "分类不存在")
			return
		}
	} else {
		product, err := model.GetProductByID(req.TargetID)
		if err != nil {
			internalErrorResponse(c, "校验商品失败: "+err.Error())
			return
		}
		if product == nil {
			badRequestResponse(c, "商品不存在")
			return
		}
	}

	if req.MinQuantityForFree != nil && *req.MinQuantityForFree <= 0 {
		badRequestResponse(c, "免配送费数量必须大于0")
		return
	}

	exists, err := model.GetDeliveryFeeExclusionByScope(req.ItemType, req.TargetID)
	if err != nil {
		internalErrorResponse(c, "检查排除项失败: "+err.Error())
		return
	}
	if exists != nil {
		badRequestResponse(c, "该对象已配置排除规则")
		return
	}

	exclusion := &model.DeliveryFeeExclusion{
		ItemType:           req.ItemType,
		TargetID:           req.TargetID,
		MinQuantityForFree: req.MinQuantityForFree,
		Remark:             strings.TrimSpace(req.Remark),
	}

	if err := model.CreateDeliveryFeeExclusion(exclusion); err != nil {
		internalErrorResponse(c, "创建排除项失败: "+err.Error())
		return
	}

	successResponse(c, exclusion, "创建成功")
}

// UpdateDeliveryFeeExclusion 更新排除项
func UpdateDeliveryFeeExclusion(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var req struct {
		MinQuantityForFree *int   `json:"min_quantity_for_free"`
		Remark             string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	exclusion, err := model.GetDeliveryFeeExclusionByID(id)
	if err != nil {
		internalErrorResponse(c, "获取排除项失败: "+err.Error())
		return
	}
	if exclusion == nil {
		notFoundResponse(c, "排除项不存在")
		return
	}

	if req.MinQuantityForFree != nil && *req.MinQuantityForFree <= 0 {
		badRequestResponse(c, "免配送费数量必须大于0")
		return
	}
	exclusion.MinQuantityForFree = req.MinQuantityForFree
	exclusion.Remark = strings.TrimSpace(req.Remark)

	if err := model.UpdateDeliveryFeeExclusion(exclusion); err != nil {
		internalErrorResponse(c, "更新排除项失败: "+err.Error())
		return
	}

	successResponse(c, exclusion, "更新成功")
}

// DeleteDeliveryFeeExclusion 删除排除项
func DeleteDeliveryFeeExclusion(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	exclusion, err := model.GetDeliveryFeeExclusionByID(id)
	if err != nil {
		internalErrorResponse(c, "获取排除项失败: "+err.Error())
		return
	}
	if exclusion == nil {
		notFoundResponse(c, "排除项不存在")
		return
	}

	if err := model.DeleteDeliveryFeeExclusion(id); err != nil {
		internalErrorResponse(c, "删除排除项失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
