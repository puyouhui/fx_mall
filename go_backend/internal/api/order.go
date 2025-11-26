package api

import (
	"net/http"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// CreateOrderRequest 小程序创建订单请求体
// 订单金额相关目前由后端根据采购单和配送费重新计算，前端传的仅作展示
type CreateOrderRequest struct {
	AddressID           int     `json:"address_id" binding:"required"`
	Remark              string  `json:"remark"`
	ItemIDs             []int   `json:"item_ids"`              // 只对勾选的采购单项下单
	OutOfStockStrategy  string  `json:"out_of_stock_strategy"` // cancel_item / ship_available / contact_me
	TrustReceipt        bool    `json:"trust_receipt"`         // 信任签收
	HidePrice           bool    `json:"hide_price"`            // 是否隐藏价格
	RequirePhoneContact bool    `json:"require_phone_contact"` // 配送时是否电话联系
	ExpectedDeliveryAt  *string `json:"expected_delivery_at"`  // 预留，暂不解析
	PointsDiscount      float64 `json:"points_discount"`       // 预留：积分抵扣金额
	CouponDiscount      float64 `json:"coupon_discount"`       // 预留：优惠券抵扣金额（当前已在购物车+确认页计算）
	DeliveryCouponID    int     `json:"delivery_coupon_id"`    // 指定免配送费券
	AmountCouponID      int     `json:"amount_coupon_id"`      // 指定金额券
}

// CreateOrderFromCart 从当前采购单创建订单
func CreateOrderFromCart(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	if req.AddressID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择收货地址"})
		return
	}

	// 校验地址归属
	address, err := model.GetAddressByID(req.AddressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取地址信息失败: " + err.Error()})
		return
	}
	if address == nil || address.UserID != user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "收货地址无效"})
		return
	}

	// 获取当前采购单
	items, err := model.GetPurchaseListItemsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单为空，无法创建订单"})
		return
	}

	// 只针对指定的 item_ids 下单
	if len(req.ItemIDs) > 0 {
		filter := make(map[int]struct{}, len(req.ItemIDs))
		for _, id := range req.ItemIDs {
			if id > 0 {
				filter[id] = struct{}{}
			}
		}
		filteredItems := make([]model.PurchaseListItem, 0, len(filter))
		for _, item := range items {
			if _, ok := filter[item.ID]; ok {
				filteredItems = append(filteredItems, item)
			}
		}
		items = filteredItems
	}

	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单为空，无法创建订单"})
		return
	}

	// 计算配送费和金额汇总
	summary, err := model.CalculateDeliveryFee(items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败: " + err.Error()})
		return
	}

	// 计算订单金额和分类信息用于优惠券筛选
	orderAmount := 0.0
	categoryIDSet := make(map[int]struct{})
	productIDs := make([]int, 0, len(items))
	for _, item := range items {
		price := item.SpecSnapshot.WholesalePrice
		if price <= 0 {
			price = item.SpecSnapshot.RetailPrice
		}
		if price <= 0 {
			price = item.SpecSnapshot.Cost
		}
		if price < 0 {
			price = 0
		}
		orderAmount += price * float64(item.Quantity)
		productIDs = append(productIDs, item.ProductID)
	}

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

	availableCoupons, err := model.GetAvailableCouponsForPurchaseList(
		user.ID,
		orderAmount,
		categoryIDs,
		summary.DeliveryFee,
		summary.IsFreeShipping,
	)
	if err != nil {
		availableCoupons = []model.AvailableCouponInfo{}
	}

	appliedCombination := model.CalculateCouponCombinationWithSelection(
		availableCoupons,
		orderAmount,
		summary.DeliveryFee,
		summary.IsFreeShipping,
		req.DeliveryCouponID,
		req.AmountCouponID,
	)

	// 使用模型层事务函数创建订单并落库
	options := model.OrderCreationOptions{
		Remark:              req.Remark,
		OutOfStockStrategy:  req.OutOfStockStrategy,
		TrustReceipt:        req.TrustReceipt,
		HidePrice:           req.HidePrice,
		RequirePhoneContact: req.RequirePhoneContact,
		PointsDiscount:      req.PointsDiscount,
		CouponDiscount:      appliedCombination.TotalDiscount,
	}

	order, orderItems, err := model.CreateOrderFromPurchaseList(user.ID, req.AddressID, items, summary, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建订单失败: " + err.Error()})
		return
	}

	// 返回创建成功的订单概要
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"order":       order,
			"order_items": orderItems,
		},
		"message": "创建订单成功",
	})
}
