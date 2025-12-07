package api

import (
	"net/http"
	"strconv"
	"strings"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// AddPurchaseListItemRequest 请求参数
type AddPurchaseListItemRequest struct {
	ProductID int    `json:"product_id" binding:"required"`
	SpecName  string `json:"spec_name"`
	Quantity  int    `json:"quantity"`
}

// UpdatePurchaseListItemRequest 更新数量请求
type UpdatePurchaseListItemRequest struct {
	Quantity int `json:"quantity" binding:"required"`
}

func getMiniUserFromContext(c *gin.Context) (*model.MiniAppUser, bool) {
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

// AddPurchaseListItem 添加采购单项
func AddPurchaseListItem(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	var req AddPurchaseListItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if req.ProductID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "商品ID不合法"})
		return
	}
	if req.Quantity == 0 {
		req.Quantity = 1
	}
	if req.Quantity < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "数量必须大于0"})
		return
	}

	product, err := model.GetProductByID(req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品失败: " + err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	specName := strings.TrimSpace(req.SpecName)
	if specName == "" {
		if len(product.Specs) == 1 {
			specName = product.Specs[0].Name
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择规格"})
			return
		}
	}

	var matchedSpec *model.Spec
	for i := range product.Specs {
		if product.Specs[i].Name == specName {
			matchedSpec = &product.Specs[i]
			break
		}
	}
	if matchedSpec == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "规格不存在"})
		return
	}

	specSnapshot := model.PurchaseSpecSnapshot{
		Name:           matchedSpec.Name,
		Description:    matchedSpec.Description,
		Cost:           matchedSpec.Cost,
		WholesalePrice: matchedSpec.WholesalePrice,
		RetailPrice:    matchedSpec.RetailPrice,
	}

	image := ""
	if len(product.Images) > 0 {
		image = product.Images[0]
	}

	item := &model.PurchaseListItem{
		UserID:       user.ID,
		ProductID:    product.ID,
		ProductName:  product.Name,
		ProductImage: image,
		SpecName:     specName,
		SpecSnapshot: specSnapshot,
		Quantity:     req.Quantity,
		IsSpecial:    product.IsSpecial,
	}

	result, err := model.AddOrUpdatePurchaseListItem(item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存采购单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "message": "添加成功"})
}

// GetPurchaseListItems 获取采购单列表
func GetPurchaseListItems(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	items, err := model.GetPurchaseListItemsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": items, "message": "获取成功"})
}

// GetPurchaseListSummary 获取采购单及配送费摘要
func GetPurchaseListSummary(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	itemIDsParam := c.Query("item_ids")
	itemIDsFilter := make(map[int]struct{})
	deliveryCouponID := parseQueryInt(c, "delivery_coupon_id", 0)
	amountCouponID := parseQueryInt(c, "amount_coupon_id", 0)
	if itemIDsParam != "" {
		for _, idStr := range strings.Split(itemIDsParam, ",") {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
				itemIDsFilter[id] = struct{}{}
			}
		}
	}

	items, err := model.GetPurchaseListItemsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	if len(itemIDsFilter) > 0 {
		filteredItems := make([]model.PurchaseListItem, 0, len(itemIDsFilter))
		for _, item := range items {
			if _, ok := itemIDsFilter[item.ID]; ok {
				filteredItems = append(filteredItems, item)
			}
		}
		items = filteredItems
	}

	if len(items) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"items":   []model.PurchaseListItem{},
				"summary": nil,
			},
			"message": "获取成功",
		})
		return
	}

	// 获取用户类型，默认为零售
	userType := user.UserType
	if userType == "" || userType == "unknown" {
		userType = "retail"
	}

	summary, err := model.CalculateDeliveryFee(items, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	// 计算订单金额和获取商品分类ID
	orderAmount := 0.0
	categoryIDSet := make(map[int]struct{})
	productIDs := make([]int, 0)

	for _, item := range items {
		// 根据用户类型计算商品金额
		var price float64
		if userType == "wholesale" {
			price = item.SpecSnapshot.WholesalePrice
		if price <= 0 {
				price = item.SpecSnapshot.RetailPrice
			}
		} else {
			price = item.SpecSnapshot.RetailPrice
			if price <= 0 {
				price = item.SpecSnapshot.WholesalePrice
			}
		}
		if price <= 0 {
			price = item.SpecSnapshot.Cost
		}
		if price < 0 {
			price = 0
		}
		orderAmount += price * float64(item.Quantity)

		// 收集商品ID用于查询分类
		productIDs = append(productIDs, item.ProductID)
	}

	// 获取商品分类信息
	categoryInfo, err := model.FetchProductCategoryInfo(productIDs)
	if err == nil {
		for _, info := range categoryInfo {
			if info.CategoryID > 0 {
				categoryIDSet[info.CategoryID] = struct{}{}
			}
			if info.ParentID > 0 {
				categoryIDSet[info.ParentID] = struct{}{}
			}
		}
	}

	categoryIDs := make([]int, 0, len(categoryIDSet))
	for id := range categoryIDSet {
		categoryIDs = append(categoryIDs, id)
	}

	// 获取可用优惠券和最佳组合
	availableCoupons, err := model.GetAvailableCouponsForPurchaseList(
		user.ID,
		orderAmount,
		categoryIDs,
		summary.DeliveryFee,
		summary.IsFreeShipping,
	)
	if err != nil {
		// 如果获取优惠券失败，不影响主流程，只记录错误
		availableCoupons = []model.AvailableCouponInfo{}
	}

	bestCombination := model.CalculateBestCouponCombination(
		availableCoupons,
		orderAmount,
		summary.DeliveryFee,
		summary.IsFreeShipping,
	)

	appliedCombination := model.CalculateCouponCombinationWithSelection(
		availableCoupons,
		orderAmount,
		summary.DeliveryFee,
		summary.IsFreeShipping,
		deliveryCouponID,
		amountCouponID,
	)

	// 获取加急费用（从系统设置）
	urgentFeeStr, _ := model.GetSystemSetting("order_urgent_fee")
	urgentFee := 0.0
	if urgentFeeStr != "" {
		if fee, err := strconv.ParseFloat(urgentFeeStr, 64); err == nil && fee > 0 {
			urgentFee = fee
		}
	}

	result := gin.H{
		"items":               items,
		"summary":             summary,
		"available_coupons":   availableCoupons,
		"best_combination":    bestCombination,
		"applied_combination": appliedCombination,
		"urgent_fee":          urgentFee, // 返回加急费用供前端显示
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "message": "获取成功"})
}

// UpdatePurchaseListItem 更新采购单项
func UpdatePurchaseListItem(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	itemID, err := strconv.Atoi(idStr)
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的采购单项ID"})
		return
	}

	var req UpdatePurchaseListItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := model.UpdatePurchaseListItemQuantity(itemID, user.ID, req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	items, err := model.GetPurchaseListItemsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": items, "message": "更新成功"})
}

// DeletePurchaseListItem 删除采购单项
func DeletePurchaseListItem(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	itemID, err := strconv.Atoi(idStr)
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的采购单项ID"})
		return
	}

	if err := model.DeletePurchaseListItem(itemID, user.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ClearPurchaseList 清空采购单
func ClearPurchaseList(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	if err := model.ClearPurchaseList(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "清空采购单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "采购单已清空"})
}
