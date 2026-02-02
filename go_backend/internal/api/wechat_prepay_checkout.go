package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

// WeChatPrepayFromCheckout 在线支付预支付（不创建订单，支付成功后在回调中创建）
// POST /mini-app/wechat-pay/prepay-from-checkout
// Body: 与 CreateOrderFromCart 相同的 CreateOrderRequest，且 payment_method 必须为 online
func WeChatPrepayFromCheckout(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}
	if strings.TrimSpace(req.PaymentMethod) != "online" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "此接口仅支持在线支付"})
		return
	}
	if req.AddressID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择收货地址"})
		return
	}

	address, err := model.GetAddressByID(req.AddressID)
	if err != nil || address == nil || address.UserID != user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "收货地址无效"})
		return
	}

	items, err := model.GetPurchaseListItemsByUserID(user.ID)
	if err != nil || len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单为空，无法支付"})
		return
	}
	if len(req.ItemIDs) > 0 {
		filter := make(map[int]struct{}, len(req.ItemIDs))
		for _, id := range req.ItemIDs {
			if id > 0 {
				filter[id] = struct{}{}
			}
		}
		filtered := make([]model.PurchaseListItem, 0, len(filter))
		for _, it := range items {
			if _, ok := filter[it.ID]; ok {
				filtered = append(filtered, it)
			}
		}
		items = filtered
	}
	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "采购单为空，无法支付"})
		return
	}

	for _, item := range items {
		if item.SpecSnapshot.Cost <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "当前订单异常，不能支付，请联系管理员"})
			return
		}
	}

	userType := user.UserType
	if userType == "" || userType == "unknown" {
		userType = "retail"
	}

	summary, err := model.CalculateDeliveryFee(items, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "计算配送费失败"})
		return
	}

	orderAmount := 0.0
	categoryIDSet := make(map[int]struct{})
	productIDs := make([]int, 0, len(items))
	for _, item := range items {
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
		productIDs = append(productIDs, item.ProductID)
	}

	categoryInfo, _ := model.FetchProductCategoryInfo(productIDs)
	for _, info := range categoryInfo {
		if info.CategoryID > 0 {
			categoryIDSet[info.CategoryID] = struct{}{}
		}
		if info.ParentID > 0 {
			categoryIDSet[info.ParentID] = struct{}{}
		}
	}
	categoryIDs := make([]int, 0, len(categoryIDSet))
	for id := range categoryIDSet {
		categoryIDs = append(categoryIDs, id)
	}

	availableCoupons, _ := model.GetAvailableCouponsForPurchaseList(user.ID, orderAmount, categoryIDs, summary.DeliveryFee, summary.IsFreeShipping)
	appliedCombination := model.CalculateCouponCombinationWithSelection(
		availableCoupons, orderAmount, summary.DeliveryFee, summary.IsFreeShipping,
		req.DeliveryCouponID, req.AmountCouponID,
	)

	urgentFee := 0.0
	if req.IsUrgent {
		if s, err := model.GetSystemSetting("order_urgent_fee"); err == nil && s != "" {
			if f, e := strconv.ParseFloat(s, 64); e == nil && f > 0 {
				urgentFee = f
			}
		}
	}

	options := model.OrderCreationOptions{
		Remark:              req.Remark,
		OutOfStockStrategy:  req.OutOfStockStrategy,
		TrustReceipt:        req.TrustReceipt,
		HidePrice:           req.HidePrice,
		RequirePhoneContact: req.RequirePhoneContact,
		PointsDiscount:      req.PointsDiscount,
		CouponDiscount:      appliedCombination.TotalDiscount,
		IsUrgent:            req.IsUrgent,
		UrgentFee:           urgentFee,
		DeliveryFeeCouponID: 0,
		AmountCouponID:      0,
		PaymentMethod:       "online",
	}
	if appliedCombination.DeliveryFeeCoupon != nil && appliedCombination.DeliveryFeeCoupon.UserCouponID > 0 {
		options.DeliveryFeeCouponID = appliedCombination.DeliveryFeeCoupon.UserCouponID
	}
	if appliedCombination.AmountCoupon != nil && appliedCombination.AmountCoupon.UserCouponID > 0 {
		options.AmountCouponID = appliedCombination.AmountCoupon.UserCouponID
	}

	goodsAmount := summary.TotalAmount
	deliveryFee := summary.DeliveryFee
	if summary.IsFreeShipping {
		deliveryFee = 0
	}
	pointsDiscount := options.PointsDiscount
	couponDiscount := options.CouponDiscount
	if pointsDiscount < 0 {
		pointsDiscount = 0
	}
	if couponDiscount < 0 {
		couponDiscount = 0
	}
	if !req.IsUrgent {
		urgentFee = 0
	}
	totalAmount := goodsAmount + deliveryFee + urgentFee - pointsDiscount - couponDiscount
	if totalAmount < 0 {
		totalAmount = 0
	}
	if totalAmount < 0.01 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单金额异常"})
		return
	}

	outTradeNo := GenerateOutTradeNo()
	cacheEntry := &model.CachedPrepayEntry{
		UserID:    user.ID,
		AddressID: req.AddressID,
		UserType:  userType,
		Items:     items,
		Summary:   summary,
		Options:   options,
		ItemIDs:   req.ItemIDs,
	}
	SetPrepayCache(outTradeNo, cacheEntry)
	log.Printf("[WeChatPrepayFromCheckout] 预支付缓存已写入 out_trade_no=%s", outTradeNo)

	cfg, err := getWechatPayConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	mchPrivateKey, err := utils.LoadPrivateKey(cfg.PrivateKeyPEM)
	if err != nil {
		log.Printf("[WeChatPrepayFromCheckout] 加载私钥失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "微信支付配置错误"})
		return
	}

	ctx := c.Request.Context()
	var opts []core.ClientOption
	if cfg.PublicKeyID != "" && cfg.PublicKeyPEM != "" {
		wechatPubKey, err := utils.LoadPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "微信支付公钥配置错误"})
			return
		}
		opts = []core.ClientOption{
			option.WithWechatPayPublicKeyAuthCipher(cfg.MchID, cfg.SerialNo, mchPrivateKey, cfg.PublicKeyID, wechatPubKey),
		}
	} else {
		opts = []core.ClientOption{
			option.WithWechatPayAutoAuthCipher(cfg.MchID, cfg.SerialNo, mchPrivateKey, cfg.APIv3Key),
		}
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "微信支付服务异常"})
		return
	}

	amountFen := int64(totalAmount * 100)
	if amountFen < 1 {
		amountFen = 1
	}
	description := buildPayDescriptionFromItems(items)

	svc := jsapi.JsapiApiService{Client: client}
	resp, _, err := svc.PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid:       core.String(cfg.AppID),
		Mchid:       core.String(cfg.MchID),
		Description: core.String(description),
		OutTradeNo:  core.String(outTradeNo),
		NotifyUrl:   core.String(cfg.NotifyURL),
		Amount:      &jsapi.Amount{Total: core.Int64(amountFen)},
		Payer:       &jsapi.Payer{Openid: core.String(user.UniqueID)},
	})
	if err != nil {
		log.Printf("[WeChatPrepayFromCheckout] 预支付失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "发起支付失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"out_trade_no": outTradeNo,
			"timeStamp":    resp.TimeStamp,
			"nonceStr":     resp.NonceStr,
			"package":      resp.Package,
			"signType":     resp.SignType,
			"paySign":      resp.PaySign,
		},
	})
}

// buildPayDescriptionFromItems 从商品列表构建支付描述
func buildPayDescriptionFromItems(items []model.PurchaseListItem) string {
	var parts []string
	for _, it := range items {
		part := it.ProductName
		if it.SpecName != "" {
			part += " " + it.SpecName
		}
		part += "×" + strconv.Itoa(it.Quantity)
		parts = append(parts, part)
	}
	desc := strings.Join(parts, ";")
	if len(desc) > 127 {
		desc = desc[:124] + "..."
	}
	return desc
}

// ClearPurchaseListByItemIDs 根据 item_ids 删除采购单项（支付回调创建订单后调用）
func ClearPurchaseListByItemIDs(userID int, itemIDs []int) {
	if len(itemIDs) == 0 {
		_, _ = database.DB.Exec("DELETE FROM purchase_list_items WHERE user_id = ?", userID)
		return
	}
	validIDs := make([]interface{}, 0, len(itemIDs)+1)
	validIDs = append(validIDs, userID)
	ph := ""
	for _, id := range itemIDs {
		if id <= 0 {
			continue
		}
		if ph != "" {
			ph += ","
		}
		ph += "?"
		validIDs = append(validIDs, id)
	}
	// 若筛选后无有效 ID（如前端传了非法数据），退化为清空该用户全部采购单，避免漏删
	if len(validIDs) <= 1 {
		_, _ = database.DB.Exec("DELETE FROM purchase_list_items WHERE user_id = ?", userID)
		return
	}
	query := "DELETE FROM purchase_list_items WHERE user_id = ? AND id IN (" + ph + ")"
	_, _ = database.DB.Exec(query, validIDs...)
}
