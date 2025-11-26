<template>
  <view class="order-confirm-page">
    <!-- 地址栏 -->
    <view class="section address-section" @click="goSelectAddress">
      <view class="address-left">
        <view v-if="defaultAddress" class="address-main">
          <view class="address-row">
            <text class="address-name">{{ defaultAddress.name || '收货地址' }}</text>
            <text class="address-tag" v-if="defaultAddress.is_default">默认</text>
          </view>
          <view class="address-row">
            <text class="address-contact">{{ defaultAddress.contact }}</text>
            <text class="address-phone">{{ defaultAddress.phone }}</text>
          </view>
          <view class="address-detail">{{ defaultAddress.address }}</view>
        </view>
        <view v-else class="address-empty">
          <text>请选择收货地址</text>
        </view>
      </view>
      <view class="address-right">
        <uni-icons type="right" size="20" color="#ccc"></uni-icons>
      </view>
    </view>

    <!-- 商品信息和预计送达时间 -->
    <view class="section goods-section" v-if="items.length">
      <view class="goods-header">
        <text class="section-title">商品信息</text>
        <text class="delivery-time">预计送达：{{ expectedDeliveryText }}</text>
      </view>
      <view class="goods-list">
        <view class="goods-item" v-for="item in items" :key="item.id">
          <image :src="item.product_image || defaultImage" class="goods-image" mode="aspectFill" />
          <view class="goods-info">
            <text class="goods-name">{{ item.product_name }}</text>
            <text class="goods-spec" v-if="item.spec_name">{{ item.spec_name }}</text>
            <view class="goods-bottom">
              <text class="goods-price">¥{{ getDisplayPrice(item).toFixed(2) }}</text>
              <text class="goods-qty">× {{ item.quantity }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 金额信息 -->
    <view class="section amount-section">
      <view class="amount-row">
        <text>商品金额</text>
        <text>¥{{ goodsAmount }}</text>
      </view>
      <view class="amount-row">
        <text>配送费</text>
        <view class="amount-right">
          <text>{{ deliveryFeeText }}</text>
          <text class="delivery-note" v-if="deliveryFeeNote">{{ deliveryFeeNote }}</text>
        </view>
      </view>
      <view class="amount-row">
        <text>积分抵扣</text>
        <text class="muted">暂未使用</text>
      </view>
      <view class="amount-row" v-if="couponDiscountText">
        <text>优惠券</text>
        <text class="success">{{ couponDiscountText }}</text>
      </view>
      <view class="amount-row total">
        <text>小计</text>
        <text>¥{{ totalAmount }}</text>
      </view>
    </view>

    <!-- 备注和缺货处理 -->
    <view class="section remark-section">
      <view class="remark-row">
        <text class="section-title">订单备注</text>
      </view>
      <textarea
        class="remark-input"
        v-model="remark"
        placeholder="如有特殊要求可在此说明，例如需要纸箱包装"
        auto-height
      />

      <view class="sub-section">
        <text class="sub-title">遇到缺货时</text>
        <!-- uni-app 的 radio-group 不支持 v-model，这里用 @change 手动更新 -->
        <radio-group class="strategy-group" @change="onOutOfStockChange">
          <label class="strategy-item" v-for="item in outOfStockOptions" :key="item.value">
            <radio :value="item.value" :checked="outOfStockStrategy === item.value" />
            <text class="strategy-text">{{ item.label }}</text>
          </label>
        </radio-group>
      </view>
    </view>

    <!-- 其他选项 -->
    <view class="section options-section">
      <view class="option-row">
        <view class="option-text">
          <text class="option-title">信任签收</text>
          <text class="option-desc">配送电话联系不上时，允许放门口或指定位置</text>
        </view>
        <switch :checked="trustReceipt" @change="trustReceipt = $event.detail.value" />
      </view>

      <view class="option-row">
        <view class="option-text">
          <text class="option-title">隐藏价格</text>
          <text class="option-desc">选择后，小票中将不显示商品价格</text>
        </view>
        <switch :checked="hidePrice" @change="hidePrice = $event.detail.value" />
      </view>

      <view class="option-row">
        <view class="option-text">
          <text class="option-title">配送时电话联系</text>
          <text class="option-desc">建议保持电话畅通，方便配送员联系</text>
        </view>
        <switch :checked="requirePhoneContact" @change="requirePhoneContact = $event.detail.value" />
      </view>
    </view>

    <!-- 底部提交栏 -->
    <view class="bottom-bar">
      <view class="bottom-left">
        <text class="bottom-label">应付：</text>
        <text class="bottom-amount">¥{{ totalAmount }}</text>
      </view>
      <button class="submit-btn" @click="submitOrder">提交订单</button>
    </view>
  </view>
</template>

<script>
import { getPurchaseListSummary, getMiniUserDefaultAddress, createOrder } from '../../api/index.js'

export default {
  data() {
    return {
      token: '',
      itemIds: [],
      userType: 'unknown',
      preDeliveryCouponId: 0,
      preAmountCouponId: 0,
      deliveryCouponId: 0,
      amountCouponId: 0,
      items: [],
      summary: null,
      defaultAddress: null,
      remark: '',
      outOfStockStrategy: 'contact_me',
      outOfStockOptions: [
        { value: 'cancel_item', label: '缺货商品不要，其他正常发货' },
        { value: 'ship_available', label: '有货就发，缺货商品不发' },
        { value: 'contact_me', label: '由客服或配送员联系我确认' }
      ],
      trustReceipt: false,
      hidePrice: false,
      requirePhoneContact: true,
      expectedDeliveryText: '尽快送达',
      defaultImage: '/static/empty-cart.png',
      pointsDiscount: 0,
      couponDiscount: 0,
      deliveryFeeSaved: 0,
      amountCouponSaved: 0,
      submitting: false
    }
  },
  computed: {
    goodsAmount() {
      if (!this.items.length) return '0.00'
      const value = this.items.reduce((sum, item) => {
        return sum + this.getDisplayPrice(item) * (item.quantity || 0)
      }, 0)
      return value.toFixed(2)
    },
    deliveryFeeText() {
      if (!this.summary) return '¥0.00'
      const fee = Number(this.summary.delivery_fee || 0)
      const base = this.summary.is_free_shipping ? 0 : fee
      const actual = Math.max(base - this.deliveryFeeSaved, 0)
      if (actual <= 0) {
        return '免配送费'
      }
      return '¥' + actual.toFixed(2)
    },
    deliveryFeeNote() {
      if (!this.summary) return ''
      if (this.summary.is_free_shipping) {
        return `已满足满 ¥${this.freeShippingThresholdText} 免配送费`
      }
      if (this.deliveryFeeSaved > 0) {
        return `已使用免配送费券减免 ¥${this.deliveryFeeSaved.toFixed(2)}`
      }
      return ''
    },
    totalAmount() {
      const goods = Number(this.goodsAmount || 0)
      if (!this.summary) return goods.toFixed(2)
      const fee = this.summary.is_free_shipping ? 0 : Number(this.summary.delivery_fee || 0)
      const actualFee = Math.max(fee - this.deliveryFeeSaved, 0)
      const total = goods + actualFee - Number(this.pointsDiscount || 0) - Number(this.amountCouponSaved || 0)
      return total.toFixed(2)
    },
    couponDiscountText() {
      if (!this.amountCouponSaved) return ''
      return `-¥${Number(this.amountCouponSaved).toFixed(2)}`
    },
    totalDiscountText() {
      const total = Number(this.amountCouponSaved || 0)
      if (!total) return ''
      return `-¥${total.toFixed(2)}`
    }
  },
  onLoad(options) {
    this.token = uni.getStorageSync('miniUserToken') || ''
    const miniUserInfo = uni.getStorageSync('miniUserInfo')
    if (miniUserInfo && miniUserInfo.user_type) {
      this.userType = miniUserInfo.user_type
    }
    if (!this.token) {
      uni.showToast({ title: '请先登录', icon: 'none' })
      setTimeout(() => {
        uni.navigateBack()
      }, 1500)
      return
    }
    if (options && options.item_ids) {
      const decoded = decodeURIComponent(options.item_ids)
      this.itemIds = decoded
        .split(',')
        .map(id => parseInt(id, 10))
        .filter(id => id > 0)
    }
    if (options && options.delivery_coupon_id) {
      const decodedDelivery = decodeURIComponent(options.delivery_coupon_id)
      const parsed = parseInt(decodedDelivery, 10)
      this.preDeliveryCouponId = isNaN(parsed) ? 0 : parsed
    }
    if (options && options.amount_coupon_id) {
      const decodedAmount = decodeURIComponent(options.amount_coupon_id)
      const parsed = parseInt(decodedAmount, 10)
      this.preAmountCouponId = isNaN(parsed) ? 0 : parsed
    }
    this.loadData()
  },
  methods: {
    async loadData() {
      try {
        await Promise.all([this.loadAddress(), this.loadPurchaseSummary()])
      } catch (e) {
        console.error('加载确认订单数据失败:', e)
      }
    },
    async loadAddress() {
      try {
        const res = await getMiniUserDefaultAddress(this.token)
        if (res && res.code === 200) {
          this.defaultAddress = res.data || null
        }
      } catch (e) {
        console.error('获取默认地址失败:', e)
      }
    },
    async loadPurchaseSummary() {
      try {
        const params = {}
        if (this.itemIds.length) {
          params.item_ids = this.itemIds.join(',')
        }
        if (this.preDeliveryCouponId) {
          params.delivery_coupon_id = this.preDeliveryCouponId
        }
        if (this.preAmountCouponId) {
          params.amount_coupon_id = this.preAmountCouponId
        }
        const res = await getPurchaseListSummary(this.token, params)
        if (res && res.code === 200 && res.data) {
          this.items = Array.isArray(res.data.items) ? res.data.items : []
          this.summary = res.data.summary || null
          if (!this.itemIds.length && this.items.length) {
            this.itemIds = this.items.map(item => item.id)
          }
          if (!this.items.length || !this.summary) {
            uni.showToast({ title: '请选择要下单的商品', icon: 'none' })
            setTimeout(() => uni.navigateBack(), 800)
            return
          }
          const combination = res.data.applied_combination || res.data.best_combination
          if (combination) {
            this.deliveryFeeSaved = Number(combination.delivery_fee_saved || 0)
            this.amountCouponSaved = Number(combination.amount_saved || 0)
            this.deliveryCouponId = combination.delivery_fee_coupon ? combination.delivery_fee_coupon.user_coupon_id : 0
            this.amountCouponId = combination.amount_coupon ? combination.amount_coupon.user_coupon_id : 0
            this.preDeliveryCouponId = this.deliveryCouponId || 0
            this.preAmountCouponId = this.amountCouponId || 0
            this.couponDiscount = this.deliveryFeeSaved + this.amountCouponSaved
          } else {
            this.deliveryFeeSaved = 0
            this.amountCouponSaved = 0
            this.deliveryCouponId = 0
            this.amountCouponId = 0
            this.preDeliveryCouponId = 0
            this.preAmountCouponId = 0
            this.couponDiscount = 0
          }
        }
      } catch (e) {
        console.error('获取采购单汇总失败:', e)
      }
    },
    getDisplayPrice(item) {
      const snapshot = item.spec_snapshot || {}
      const wholesale = Number(snapshot.wholesale_price || 0)
      const retail = Number(snapshot.retail_price || 0)
      if (this.userType === 'wholesale') {
        return wholesale || retail || Number(snapshot.cost || 0)
      }
      return retail || wholesale || Number(snapshot.cost || 0)
    },
    goSelectAddress() {
      uni.navigateTo({ url: '/pages/address/address' })
    },
    onOutOfStockChange(e) {
      this.outOfStockStrategy = e.detail.value
    },
    async submitOrder() {
      if (!this.defaultAddress) {
        uni.showToast({ title: '请先选择收货地址', icon: 'none' })
        return
      }
      if (!this.items.length) {
        uni.showToast({ title: '采购单为空', icon: 'none' })
        return
      }
      if (this.submitting) return
      this.submitting = true
      try {
        const payload = {
          address_id: this.defaultAddress.id,
          remark: this.remark,
          out_of_stock_strategy: this.outOfStockStrategy,
          trust_receipt: this.trustReceipt,
          hide_price: this.hidePrice,
          require_phone_contact: this.requirePhoneContact,
          expected_delivery_at: null,
          points_discount: this.pointsDiscount,
          coupon_discount: this.couponDiscount,
          item_ids: this.itemIds,
          delivery_coupon_id: this.deliveryCouponId || null,
          amount_coupon_id: this.amountCouponId || null
        }
        const res = await createOrder(payload, this.token)
        if (res && res.code === 200) {
          uni.showToast({ title: '下单成功', icon: 'success' })
          // 下单成功后返回采购单页面或跳转到订单详情（预留）
          setTimeout(() => {
            uni.switchTab({ url: '/pages/cart/cart' })
          }, 800)
        } else {
          uni.showToast({ title: res.message || '下单失败', icon: 'none' })
        }
      } catch (e) {
        console.error('提交订单失败:', e)
        uni.showToast({ title: '提交失败，请稍后再试', icon: 'none' })
      } finally {
        this.submitting = false
      }
    }
  }
}
</script>

<style scoped>
.order-confirm-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 120rpx;
}

.section {
  background-color: #fff;
  margin: 20rpx;
  border-radius: 16rpx;
  padding: 24rpx 24rpx 20rpx;
}

.address-section {
  display: flex;
  align-items: center;
}

.address-left {
  flex: 1;
}

.address-right {
  padding-left: 20rpx;
}

.address-main .address-row {
  display: flex;
  align-items: center;
  margin-bottom: 8rpx;
}

.address-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}

.address-tag {
  margin-left: 12rpx;
  padding: 2rpx 10rpx;
  font-size: 20rpx;
  color: #20CB6B;
  border-radius: 20rpx;
  border: 1px solid #20CB6B;
}

.address-contact,
.address-phone {
  font-size: 26rpx;
  color: #666;
  margin-right: 16rpx;
}

.address-detail {
  font-size: 24rpx;
  color: #909399;
}

.address-empty {
  font-size: 28rpx;
  color: #909399;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
}

.goods-section .goods-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.delivery-time {
  font-size: 24rpx;
  color: #909399;
}

.goods-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.goods-item {
  display: flex;
}

.goods-image {
  width: 120rpx;
  height: 120rpx;
  border-radius: 12rpx;
  margin-right: 16rpx;
  background-color: #f5f5f5;
}

.goods-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.goods-name {
  font-size: 28rpx;
  color: #333;
}

.goods-spec {
  font-size: 24rpx;
  color: #909399;
}

.goods-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.goods-price {
  font-size: 28rpx;
  color: #ff4d4f;
  font-weight: 600;
}

.goods-qty {
  font-size: 24rpx;
  color: #666;
}

.amount-section .amount-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
  font-size: 26rpx;
  color: #333;
}

.amount-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.delivery-note {
  font-size: 24rpx;
  color: #909399;
  margin-top: 4rpx;
}

.amount-section .amount-row .muted {
  color: #909399;
}

.amount-section .amount-row .success {
  color: #20CB6B;
}

.amount-section .amount-row.total {
  margin-top: 10rpx;
  font-size: 30rpx;
  font-weight: 600;
}

.remark-input {
  margin-top: 16rpx;
  padding: 16rpx;
  background-color: #f5f5f5;
  border-radius: 12rpx;
  font-size: 26rpx;
  min-height: 120rpx;
}

.sub-section {
  margin-top: 24rpx;
}

.sub-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
}

.strategy-group {
  margin-top: 12rpx;
}

.strategy-item {
  display: flex;
  align-items: center;
  margin-bottom: 12rpx;
}

.strategy-text {
  margin-left: 12rpx;
  font-size: 26rpx;
  color: #555;
}

.options-section .option-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10rpx 0;
}

.option-text {
  flex: 1;
}

.option-title {
  font-size: 28rpx;
  color: #333;
}

.option-desc {
  font-size: 24rpx;
  color: #909399;
}

.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: 100rpx;
  background-color: #fff;
  border-top: 1rpx solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24rpx;
}

.bottom-left {
  display: flex;
  align-items: baseline;
}

.bottom-label {
  font-size: 26rpx;
  color: #666;
}

.bottom-amount {
  font-size: 34rpx;
  color: #ff4d4f;
  font-weight: 600;
  margin-left: 8rpx;
}

.submit-btn {
  background-color: #20CB6B;
  color: #fff;
  font-size: 28rpx;
  padding: 12rpx 40rpx;
  border-radius: 40rpx;
}
</style>


