<template>
  <view class="order-confirm-page">
    <!-- 地址栏 -->
    <view class="section address-section" @click="goSelectAddress">
      <view class="address-content">
        <view v-if="defaultAddress" class="address-main">
          <view class="address-header">
            <view class="address-title-row">
              <text class="address-store">{{ defaultAddress.name || '收货地址' }}</text>
            <text class="address-tag" v-if="defaultAddress.is_default">默认</text>
          </view>
            <view class="address-contact-row">
            <text class="address-contact">{{ defaultAddress.contact }}</text>
            <text class="address-phone">{{ defaultAddress.phone }}</text>
          </view>
          <view class="address-detail">{{ defaultAddress.address }}</view>
          </view>
        </view>
        <view v-else class="address-empty">
          <text class="empty-text">请选择收货地址</text>
        </view>
        <view class="address-arrow">
          <uni-icons type="right" size="18" color="#C0C4CC"></uni-icons>
        </view>
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
            <view class="goods-name-row">
            <text class="goods-name">{{ item.product_name }}</text>
            </view>
            <text class="goods-spec" v-if="item.spec_name">{{ item.spec_name }}</text>
            <view class="goods-bottom">
              <text class="goods-price">¥{{ getDisplayPrice(item).toFixed(2) }}</text>
              <text class="goods-qty">× {{ item.quantity }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 加急订单模块 -->
    <view class="section urgent-section">
      <view class="urgent-container" :class="{ 'urgent-active': isUrgent }">
        <view class="urgent-left">
          <view class="urgent-header">
            <text class="urgent-title">加急订单</text>
            <text class="urgent-tag">平台将为您加急配送</text>
          </view>
        </view>
        <view class="urgent-right">
          <view class="urgent-price-wrapper" v-if="urgentFee > 0">
            <text class="urgent-price">¥{{ urgentFee.toFixed(2) }}</text>
          </view>
          <switch 
            :checked="isUrgent" 
            @change="onUrgentChange"
            color="#20CB6B"
            class="urgent-switch"
          />
        </view>
      </view>
    </view>

    <!-- 金额信息 -->
    <view class="section amount-section">
      <view class="amount-row">
        <text class="amount-label">商品金额</text>
        <text class="amount-value">¥{{ goodsAmount }}</text>
      </view>
      <view class="amount-row">
        <text class="amount-label">配送费</text>
        <text class="amount-value" :class="deliveryFeeText === '免配送费' ? 'free-text' : ''">{{ deliveryFeeText }}</text>
      </view>
      <!-- 积分抵扣暂时隐藏 -->
      <!-- <view class="amount-row">
        <text class="amount-label">积分抵扣</text>
        <text class="amount-value muted">暂未使用</text>
      </view> -->
      <view class="amount-row" v-if="couponDiscountText">
        <text class="amount-label">优惠券</text>
        <text class="amount-value discount-text">{{ couponDiscountText }}</text>
      </view>
      <view class="amount-row urgent-fee-row" v-if="isUrgent">
        <view class="urgent-fee-label-wrapper">
          <text class="amount-label urgent-fee-label">加急费用</text>
          <text class="urgent-fee-tag">将优先为您配送</text>
        </view>
        <text class="amount-value urgent-fee-value">¥{{ (urgentFee || 0).toFixed(2) }}</text>
      </view>
      <view class="amount-divider"></view>
      <view class="amount-row total-row">
        <text class="amount-label total-label">小计</text>
        <text class="amount-value total-value">¥{{ totalAmount }}</text>
      </view>
    </view>

    <!-- 备注和缺货处理 -->
    <view class="section remark-section">
      <view class="remark-header">
        <text class="section-title">订单备注</text>
      </view>
      <textarea
        class="remark-input"
        v-model="remark"
        placeholder="如有特殊要求可在此说明，例如需要纸箱包装"
        auto-height
        maxlength="200"
      />

      <view class="sub-section">
        <text class="sub-title">遇到缺货时</text>
        <radio-group class="strategy-group" @change="onOutOfStockChange">
          <label 
            class="strategy-item" 
            v-for="item in outOfStockOptions" 
            :key="item.value"
            :class="{ 'strategy-item-active': outOfStockStrategy === item.value }"
          >
            <radio 
              :value="item.value" 
              :checked="outOfStockStrategy === item.value"
              color="#20CB6B"
            />
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
        <switch 
          :checked="trustReceipt" 
          @change="trustReceipt = $event.detail.value"
          color="#20CB6B"
        />
      </view>

      <view class="option-row">
        <view class="option-text">
          <text class="option-title">隐藏价格</text>
          <text class="option-desc">选择后，小票中将不显示商品价格</text>
        </view>
        <switch 
          :checked="hidePrice" 
          @change="hidePrice = $event.detail.value"
          color="#20CB6B"
        />
      </view>

      <view class="option-row">
        <view class="option-text">
          <text class="option-title">配送时电话联系</text>
          <text class="option-desc">建议保持电话畅通，方便配送员联系</text>
        </view>
        <switch 
          :checked="requirePhoneContact" 
          @change="requirePhoneContact = $event.detail.value"
          color="#20CB6B"
        />
      </view>
    </view>

    <!-- 底部提交栏 -->
    <view class="bottom-bar">
      <view class="bottom-left">
        <text class="bottom-label">合计：</text>
        <text class="bottom-amount">¥{{ totalAmount }}</text>
      </view>
      <button 
        class="submit-btn" 
        :class="{ 'submit-btn-disabled': submitting }"
        :disabled="submitting"
        @click="goToCheckout"
      >
        {{ submitting ? '提交中...' : '提交订单' }}
      </button>
    </view>
  </view>
</template>

<script>
import { getPurchaseListSummary, getMiniUserDefaultAddress } from '../../api/index.js'

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
      trustReceipt: true,
      hidePrice: false,
      requirePhoneContact: true,
      expectedDeliveryText: '尽快送达',
      defaultImage: 'https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg',
      pointsDiscount: 0,
      couponDiscount: 0,
      deliveryFeeSaved: 0,
      amountCouponSaved: 0,
      submitting: false,
      isUrgent: false,
      urgentFee: 0,
      showConfirm: false
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
      const urgent = this.isUrgent ? (this.urgentFee || 0) : 0
      const total = goods + actualFee + urgent - Number(this.pointsDiscount || 0) - Number(this.amountCouponSaved || 0)
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
    },
    // 第一个商品名称
    firstGoodsName() {
      if (!this.items.length) return ''
      return this.items[0].product_name || '商品'
    },
    // 商品总数量
    totalQuantity() {
      if (!this.items.length) return 0
      return this.items.reduce((sum, item) => sum + (item.quantity || 0), 0)
    }
  },
  onLoad(options) {
    // 监听地址选择事件
    uni.$on('addressSelected', this.onAddressSelected);
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
          // 获取加急费用（从API返回的数据中获取）
          if (res.data.urgent_fee !== undefined) {
            this.urgentFee = Number(res.data.urgent_fee || 0)
          }
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
      uni.navigateTo({ url: '/pages/address/address?selectMode=true' })
    },
    onOutOfStockChange(e) {
      this.outOfStockStrategy = e.detail.value
    },
    // 加急订单开关变化
    onUrgentChange(e) {
      this.isUrgent = e.detail.value
      // 如果开启加急但费用为0，尝试从API获取
      if (this.isUrgent && this.urgentFee === 0) {
        // 加急费用应该已经从 loadPurchaseSummary 中获取了
        // 如果还是0，说明系统设置中加急费用为0
        console.log('加急费用:', this.urgentFee)
      }
    },
    // 跳转到收银台
    goToCheckout() {
      if (!this.defaultAddress) {
        uni.showToast({ title: '请先选择收货地址', icon: 'none' })
        return
      }
      if (!this.items.length) {
        uni.showToast({ title: '采购单为空', icon: 'none' })
        return
      }
      if (this.submitting) return

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
        amount_coupon_id: this.amountCouponId || null,
        is_urgent: this.isUrgent
      }

      const checkoutData = {
        payload,
        firstGoodsName: this.firstGoodsName,
        totalQuantity: this.totalQuantity,
        totalAmount: this.totalAmount
      }
      uni.setStorageSync('checkoutOrderData', JSON.stringify(checkoutData))
      uni.navigateTo({ url: '/pages/order/checkout' })
    },
    // 地址选择回调
    onAddressSelected(address) {
      this.defaultAddress = address
      // 重新加载采购单汇总（地址变化可能影响配送费）
      this.loadPurchaseSummary()
    }
  },
  onUnload() {
    // 页面卸载时移除事件监听
    uni.$off('addressSelected', this.onAddressSelected)
  }
}
</script>

<style scoped>
.order-confirm-page {
  min-height: 100vh;
  background-color: #F5F6FA;
  padding-bottom: 180rpx;
}

.section {
  background-color: #fff;
  margin: 20rpx 24rpx;
  border-radius: 20rpx;
  padding: 32rpx 28rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.address-section {
  margin-top: 24rpx;
}

.address-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.address-main {
  flex: 1;
  padding-right: 24rpx;
}

.address-header {
  width: 100%;
}

.address-title-row {
  display: flex;
  align-items: center;
  margin-bottom: 12rpx;
}

.address-store {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  line-height: 1.4;
}

.address-tag {
  margin-left: 12rpx;
  padding: 4rpx 12rpx;
  font-size: 20rpx;
  color: #20CB6B;
  background-color: #E8F8F0;
  border-radius: 4rpx;
  line-height: 1.2;
}

.address-contact-row {
  display: flex;
  align-items: center;
  margin-bottom: 8rpx;
}

.address-contact {
  font-size: 28rpx;
  color: #666;
  margin-right: 16rpx;
  font-weight: 500;
}

.address-phone {
  font-size: 28rpx;
  color: #666;
}

.address-detail {
  font-size: 26rpx;
  color: #909399;
  line-height: 1.5;
  margin-top: 4rpx;
}

.address-empty {
  flex: 1;
  padding-right: 24rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #909399;
}

.address-arrow {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  line-height: 1.4;
}

.goods-section .goods-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
  padding-bottom: 20rpx;
  border-bottom: 1rpx solid #F0F0F0;
}

.delivery-time {
  font-size: 24rpx;
  color: #909399;
}

.goods-list {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.goods-item {
  display: flex;
  align-items: flex-start;
}

.goods-image {
  width: 140rpx;
  height: 140rpx;
  border-radius: 16rpx;
  margin-right: 20rpx;
  background-color: #F5F5F5;
  flex-shrink: 0;
}

.goods-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 140rpx;
  padding-top: 4rpx;
}

.goods-name-row {
  margin-bottom: 8rpx;
}

.goods-name {
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.goods-spec {
  font-size: 24rpx;
  color: #909399;
  margin-bottom: 12rpx;
  line-height: 1.4;
}

.goods-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
}

.goods-price {
  font-size: 30rpx;
  color: #FF4D4F;
  font-weight: 600;
}

.goods-qty {
  font-size: 26rpx;
  color: #666;
  font-weight: 500;
}

/* 加急订单模块样式 */
.urgent-section {
  margin-top: 0;
}

.urgent-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-radius: 12rpx;
  transition: all 0.3s ease;
}

.urgent-left {
  flex: 1;
  display: flex;
  align-items: center;
}

.urgent-header {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.urgent-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #20CB6B;
}

.urgent-tag {
  display: inline-block;
  padding: 4rpx 12rpx;
  color: #20CB6B;
  font-size: 24rpx;
  font-weight: 500;
  border-radius: 12rpx;
  background-color: #E8F8F0;
}

.urgent-right {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.urgent-price-wrapper {
  display: flex;
  align-items: baseline;
  padding: 8rpx 16rpx;
  border-radius: 8rpx;
}

.urgent-price {
  font-size: 32rpx;
  font-weight: 700;
  color: #20CB6B;
  line-height: 1;
}

.urgent-switch {
  transform: scale(0.95);
}

.amount-section .amount-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
  min-height: 44rpx;
}

.amount-label {
  font-size: 28rpx;
  color: #666;
  font-weight: 400;
}

.amount-value {
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
  text-align: right;
}

.amount-value.free-text {
  color: #20CB6B;
}

.amount-value.muted {
  color: #909399;
}

.amount-value.discount-text {
  color: #20CB6B;
  font-weight: 500;
}

/* 加急费用突出显示 */
.urgent-fee-row {
  /* padding: 20rpx 24rpx; */
  border-radius: 12rpx;
  margin: 16rpx 0;
}

.urgent-fee-label-wrapper {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.urgent-fee-tag {
  display: inline-block;
  padding: 4rpx 12rpx;
  background-color: #E8F8F0;
  color: #20CB6B;
  font-size: 20rpx;
  border-radius: 8rpx;
  font-weight: 500;
}

.urgent-fee-value {
  color: #20CB6B;
  font-size: 30rpx;
}

.amount-divider {
  height: 1rpx;
  background-color: #F0F0F0;
  margin: 20rpx 0;
}

.amount-section .total-row {
  margin-top: 8rpx;
  margin-bottom: 0;
  padding-top: 12rpx;
}

.total-label {
  font-size: 32rpx;
  color: #333;
  font-weight: 600;
}

.total-value {
  font-size: 36rpx;
  color: #FF4D4F;
  font-weight: 600;
}

.remark-header {
  margin-bottom: 20rpx;
}

.remark-input {
  margin-top: 16rpx;
  padding: 20rpx;
  background-color: #F5F6FA;
  border-radius: 12rpx;
  font-size: 28rpx;
  color: #333;
  min-height: 140rpx;
  line-height: 1.6;
  width: 100%;
  box-sizing: border-box;
}

.sub-section {
  margin-top: 32rpx;
  padding-top: 32rpx;
  border-top: 1rpx solid #F0F0F0;
}

.sub-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 20rpx;
}

.strategy-group {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  margin-top: 16rpx;
}

.strategy-item {
  display: flex;
  align-items: flex-start;
  padding: 16rpx;
  border-radius: 12rpx;
  background-color: #F5F6FA;
  border: 1rpx solid transparent;
  transition: all 0.3s;
  box-sizing: border-box;
}

.strategy-item-active {
  background-color: #E8F8F0;
  border-color: #20CB6B;
}

.strategy-text {
  margin-left: 16rpx;
  font-size: 28rpx;
  color: #333;
  line-height: 1.5;
  flex: 1;
}

.options-section .option-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx 0;
  min-height: 80rpx;
}

.options-section .option-row:not(:last-child) {
  border-bottom: 1rpx solid #F0F0F0;
}

.option-text {
  flex: 1;
  padding-right: 24rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.option-title {
  font-size: 30rpx;
  color: #333;
  font-weight: 500;
  line-height: 1.4;
}

.option-desc {
  font-size: 24rpx;
  color: #909399;
  line-height: 1.5;
}

.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: 120rpx;
  background-color: #fff;
  border-top: 1rpx solid #F0F0F0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 32rpx;
  padding-bottom: env(safe-area-inset-bottom);
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}

.bottom-left {
  display: flex;
  align-items: baseline;
  flex: 1;
}

.bottom-label {
  font-size: 28rpx;
  color: #666;
  font-weight: 400;
}

.bottom-amount {
  font-size: 40rpx;
  color: #FF4D4F;
  font-weight: 600;
  margin-left: 8rpx;
}

.submit-btn {
  width: 50%;
  height: 48px;
  line-height: 48px;
  background-color: #20CB6B;
  color: #fff;
  font-size: 32rpx;
  font-weight: 600;
  padding: 0 60rpx;
  border-radius: 50rpx;
  border: none;
  box-shadow: 0 4rpx 16rpx rgba(32, 203, 107, 0.3);
  transition: all 0.3s;
  box-sizing: border-box;
}

.submit-btn::after {
  border: none;
}

.submit-btn-disabled {
  background-color: #CCE8D9;
  box-shadow: none;
  opacity: 0.7;
}
</style>


