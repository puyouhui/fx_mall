package model

import (
	"fmt"
	"time"
)

// AvailableCouponInfo 可用优惠券信息
type AvailableCouponInfo struct {
	UserCouponID  int     `json:"user_coupon_id"`
	CouponID      int     `json:"coupon_id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"` // delivery_fee 或 amount
	DiscountValue float64 `json:"discount_value"`
	MinAmount     float64 `json:"min_amount"`
	CategoryIDs   []int   `json:"category_ids"`
	IsAvailable   bool    `json:"is_available"`     // 是否满足使用条件
	Reason        string  `json:"reason,omitempty"` // 不可用原因
}

// BestCouponCombination 最佳优惠券组合
type BestCouponCombination struct {
	DeliveryFeeCoupon *AvailableCouponInfo `json:"delivery_fee_coupon,omitempty"` // 选中的免配送费券
	AmountCoupon      *AvailableCouponInfo `json:"amount_coupon,omitempty"`       // 选中的金额券
	TotalDiscount     float64              `json:"total_discount"`                // 总优惠金额
	DeliveryFeeSaved  float64              `json:"delivery_fee_saved"`            // 节省的配送费
	AmountSaved       float64              `json:"amount_saved"`                  // 节省的金额
}

// GetAvailableCouponsForPurchaseList 获取采购单可用的优惠券列表
func GetAvailableCouponsForPurchaseList(userID int, orderAmount float64, categoryIDs []int, deliveryFee float64, isFreeShipping bool) ([]AvailableCouponInfo, error) {
	// 获取用户所有未使用的优惠券
	userCoupons, err := GetUserCoupons(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var availableCoupons []AvailableCouponInfo

	for _, uc := range userCoupons {
		// 只处理未使用的优惠券
		if uc.Status != "unused" {
			continue
		}

		// 检查优惠券是否过期
		if uc.ExpiresAt != nil && now.After(*uc.ExpiresAt) {
			continue
		}

		// 检查优惠券本身的有效期
		if uc.Coupon == nil {
			continue
		}

		coupon := uc.Coupon
		if now.Before(coupon.ValidFrom.ToTime()) || now.After(coupon.ValidTo.ToTime()) {
			continue
		}

		// 检查优惠券是否启用
		if coupon.Status != 1 {
			continue
		}

		info := AvailableCouponInfo{
			UserCouponID:  uc.ID,
			CouponID:      coupon.ID,
			Name:          coupon.Name,
			Type:          coupon.Type,
			DiscountValue: coupon.DiscountValue,
			MinAmount:     coupon.MinAmount,
			CategoryIDs:   coupon.CategoryIDs,
			IsAvailable:   false,
		}

		// 检查金额门槛
		if coupon.MinAmount > 0 && orderAmount < coupon.MinAmount {
			info.Reason = fmt.Sprintf("订单金额需满¥%.2f", coupon.MinAmount)
			availableCoupons = append(availableCoupons, info)
			continue
		}

		// 检查分类限制
		if len(coupon.CategoryIDs) > 0 {
			hasMatchingCategory := false
			for _, orderCatID := range categoryIDs {
				for _, couponCatID := range coupon.CategoryIDs {
					if orderCatID == couponCatID {
						hasMatchingCategory = true
						break
					}
				}
				if hasMatchingCategory {
					break
				}
			}
			if !hasMatchingCategory {
				info.Reason = "订单中不包含适用分类的商品"
				availableCoupons = append(availableCoupons, info)
				continue
			}
		}

		// 对于免配送费券，如果已经免配送费，则不需要使用
		if coupon.Type == "delivery_fee" && isFreeShipping {
			info.Reason = "订单已满足免配送费条件"
			availableCoupons = append(availableCoupons, info)
			continue
		}

		// 可用
		info.IsAvailable = true
		availableCoupons = append(availableCoupons, info)
	}

	return availableCoupons, nil
}

// CalculateBestCouponCombination 计算最佳优惠券组合
func CalculateBestCouponCombination(availableCoupons []AvailableCouponInfo, orderAmount float64, deliveryFee float64, isFreeShipping bool) BestCouponCombination {
	return calculateCouponCombination(availableCoupons, orderAmount, deliveryFee, isFreeShipping, 0, 0)
}

// CalculateCouponCombinationWithSelection 根据指定的优惠券ID优先组合，若无效则回退到最佳组合
func CalculateCouponCombinationWithSelection(availableCoupons []AvailableCouponInfo, orderAmount float64, deliveryFee float64, isFreeShipping bool, deliveryCouponID int, amountCouponID int) BestCouponCombination {
	return calculateCouponCombination(availableCoupons, orderAmount, deliveryFee, isFreeShipping, deliveryCouponID, amountCouponID)
}

func calculateCouponCombination(availableCoupons []AvailableCouponInfo, orderAmount float64, deliveryFee float64, isFreeShipping bool, deliveryCouponID int, amountCouponID int) BestCouponCombination {
	result := BestCouponCombination{}

	// 分离免配送费券和金额券
	var deliveryFeeCoupons []AvailableCouponInfo
	var amountCoupons []AvailableCouponInfo

	for _, coupon := range availableCoupons {
		if !coupon.IsAvailable {
			continue
		}
		if coupon.Type == "delivery_fee" {
			deliveryFeeCoupons = append(deliveryFeeCoupons, coupon)
		} else if coupon.Type == "amount" {
			amountCoupons = append(amountCoupons, coupon)
		}
	}

	// 如果用户指定了免配送费券，优先使用
	if deliveryCouponID > 0 && !isFreeShipping {
		if coupon := findCouponByID(deliveryFeeCoupons, deliveryCouponID); coupon != nil {
			result.DeliveryFeeCoupon = coupon
			result.DeliveryFeeSaved = deliveryFee
		}
	}

	// 如果用户指定了金额券，优先使用
	if amountCouponID > 0 {
		if coupon := findCouponByID(amountCoupons, amountCouponID); coupon != nil && coupon.MinAmount <= orderAmount {
			result.AmountCoupon = coupon
			result.AmountSaved = coupon.DiscountValue
		}
	}

	// 如果用户未指定或指定无效，则回退到默认策略
	if result.DeliveryFeeCoupon == nil && !isFreeShipping && len(deliveryFeeCoupons) > 0 {
		result.DeliveryFeeCoupon = &deliveryFeeCoupons[0]
		result.DeliveryFeeSaved = deliveryFee
	}

	if result.AmountCoupon == nil && len(amountCoupons) > 0 {
		var bestAmountCoupon *AvailableCouponInfo
		maxDiscount := 0.0

		for i := range amountCoupons {
			coupon := &amountCoupons[i]
			if coupon.MinAmount <= orderAmount {
				if coupon.DiscountValue > maxDiscount {
					maxDiscount = coupon.DiscountValue
					bestAmountCoupon = coupon
				}
			}
		}

		if bestAmountCoupon != nil {
			result.AmountCoupon = bestAmountCoupon
			result.AmountSaved = bestAmountCoupon.DiscountValue
		}
	}

	result.TotalDiscount = result.DeliveryFeeSaved + result.AmountSaved
	return result
}

func findCouponByID(coupons []AvailableCouponInfo, userCouponID int) *AvailableCouponInfo {
	if userCouponID <= 0 {
		return nil
	}
	for i := range coupons {
		if coupons[i].UserCouponID == userCouponID && coupons[i].IsAvailable {
			return &coupons[i]
		}
	}
	return nil
}
