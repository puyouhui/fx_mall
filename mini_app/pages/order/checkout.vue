<template>
  <view class="checkout-page">
    <!-- 自定义导航栏 -->
    <view class="custom-navbar">
      <view class="navbar-fixed" style="background-color: #fff; border-bottom: 1rpx solid #eee;">
        <view :style="{ height: statusBarHeight + 'px' }"></view>
        <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
          <view class="navbar-left" @click="goBack">
            <uni-icons type="left" size="20" color="#333"></uni-icons>
          </view>
          <view class="navbar-title">
            <text class="navbar-title-text">收银台</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <view class="checkout-content" :style="{ paddingTop: (statusBarHeight + navBarHeight) + 'px' }">
      <!-- 订单信息 -->
      <view class="section order-summary-section">
        <view class="section-title">订单信息</view>
        <view class="order-goods-info">
          <text class="goods-name">{{ firstGoodsName }}</text>
          <text class="goods-count" v-if="totalQuantity > 1">等{{ totalQuantity }}件商品</text>
          <text class="goods-count" v-else>1件商品</text>
        </view>
        <view class="order-amount-row">
          <text class="amount-label">订单金额</text>
          <text class="amount-value">¥{{ totalAmount }}</text>
        </view>
      </view>

      <!-- 支付方式选择 -->
      <view class="section payment-method-section">
        <view class="section-title">选择支付方式</view>
        <view 
          class="payment-item" 
          :class="{ 'payment-item-active': paymentMethod === 'online' }"
          @click="paymentMethod = 'online'"
        >
          <view class="payment-left">
            <view class="payment-icon wechat-icon">
              <text class="icon-text">微</text>
            </view>
            <view class="payment-info">
              <text class="payment-name">在线支付</text>
              <text class="payment-desc">使用微信支付，支付成功后订单立即生效</text>
            </view>
          </view>
          <view class="payment-right">
            <view class="radio-outer" :class="{ 'radio-checked': paymentMethod === 'online' }">
              <view class="radio-inner" v-if="paymentMethod === 'online'"></view>
            </view>
          </view>
        </view>
        <view 
          class="payment-item" 
          :class="{ 'payment-item-active': paymentMethod === 'cod' }"
          @click="paymentMethod = 'cod'"
        >
          <view class="payment-left">
            <view class="payment-icon cod-icon">
              <uni-icons type="location" size="20" color="#20CB6B"></uni-icons>
            </view>
            <view class="payment-info">
              <text class="payment-name">货到付款</text>
              <text class="payment-desc">配送完成后付款，过程中可随时转为在线支付</text>
            </view>
          </view>
          <view class="payment-right">
            <view class="radio-outer" :class="{ 'radio-checked': paymentMethod === 'cod' }">
              <view class="radio-inner" v-if="paymentMethod === 'cod'"></view>
            </view>
          </view>
        </view>
      </view>

      <!-- 底部支付栏 -->
      <view class="checkout-footer">
        <view class="footer-left">
          <text class="footer-label">实付金额</text>
          <text class="footer-amount">¥{{ totalAmount }}</text>
        </view>
        <button 
          class="pay-btn" 
          :class="{ 'pay-btn-disabled': submitting }"
          :disabled="submitting"
          @click="handleConfirmPay"
        >
          {{ submitting ? '处理中...' : (paymentMethod === 'online' ? '确认支付' : '提交订单') }}
        </button>
      </view>
    </view>
  </view>
</template>

<script>
import { createOrder, getWechatPrepayFromCheckout } from '../../api/index.js'

const CHECKOUT_STORAGE_KEY = 'checkoutOrderData'

export default {
  data() {
    return {
      statusBarHeight: 0,
      navBarHeight: 44,
      token: '',
      checkoutData: null,
      paymentMethod: 'online', // online | cod
      submitting: false,
      firstGoodsName: '',
      totalQuantity: 0,
      totalAmount: '0.00'
    }
  },
  onLoad() {
    const systemInfo = uni.getSystemInfoSync()
    this.statusBarHeight = systemInfo.statusBarHeight || 0
    this.getMenuButtonInfo()
    this.token = uni.getStorageSync('miniUserToken') || ''
    
    if (!this.token) {
      uni.showToast({ title: '请先登录', icon: 'none' })
      setTimeout(() => uni.navigateBack(), 1500)
      return
    }

    let data = uni.getStorageSync(CHECKOUT_STORAGE_KEY)
    if (!data) {
      uni.showToast({ title: '订单数据已过期，请重新提交', icon: 'none' })
      setTimeout(() => uni.navigateBack(), 1500)
      return
    }

    try {
      this.checkoutData = typeof data === 'string' ? JSON.parse(data) : data
      if (!this.checkoutData || !this.checkoutData.payload) {
        uni.showToast({ title: '订单数据已过期，请重新提交', icon: 'none' })
        setTimeout(() => uni.navigateBack(), 1500)
        return
      }
      this.firstGoodsName = this.checkoutData.firstGoodsName || '商品'
      this.totalQuantity = this.checkoutData.totalQuantity || 0
      this.totalAmount = this.checkoutData.totalAmount || '0.00'
    } catch (e) {
      console.error('解析订单数据失败:', e)
      uni.showToast({ title: '订单数据异常', icon: 'none' })
      setTimeout(() => uni.navigateBack(), 1500)
    }
  },
  methods: {
    getMenuButtonInfo() {
      try {
        // #ifndef H5 || APP-PLUS || MP-ALIPAY
        const menuButtonInfo = uni.getMenuButtonBoundingClientRect()
        this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight)
        // #endif
      } catch (e) {
        console.error('获取胶囊按钮信息失败:', e)
      }
    },
    goBack() {
      uni.navigateBack()
    },
    async handleConfirmPay() {
      if (this.submitting || !this.checkoutData) return
      
      this.submitting = true
      try {
        const payload = {
          ...this.checkoutData.payload,
          payment_method: this.paymentMethod
        }

        // 在线支付：不创建订单，直接预支付，支付成功后在回调中创建
        if (this.paymentMethod === 'online') {
          const res = await getWechatPrepayFromCheckout(payload, this.token)
          if (!res || res.code !== 200 || !res.data) {
            uni.showToast({ title: res?.message || '获取支付参数失败', icon: 'none' })
            return
          }
          try {
            await this.doWechatPayWithParams(res.data)
          } catch (e) {
            // 用户取消或支付失败，不做任何操作，可再次点击支付
            return
          }
          const outTradeNo = res.data?.out_trade_no || ''
          uni.removeStorageSync(CHECKOUT_STORAGE_KEY)
          uni.showToast({ title: '支付成功', icon: 'success', duration: 1500 })
          // 订单由支付回调异步创建，直接跳转订单详情（用 order_number 即 out_trade_no），详情页将轮询直到订单出现
          setTimeout(() => {
            if (outTradeNo) {
              uni.redirectTo({ url: `/pages/order/detail?id=${encodeURIComponent(outTradeNo)}&fromPayment=1` })
            } else {
              uni.redirectTo({ url: '/pages/order/list' })
            }
          }, 1500)
          return
        }

        // 货到付款：创建订单
        const res = await createOrder(payload, this.token)
        if (!res || res.code !== 200) {
          uni.showToast({ title: res?.message || '下单失败', icon: 'none' })
          return
        }
        const orderData = res.data?.order || {}
        const orderId = orderData.id || orderData.order_id
        if (!orderId) {
          uni.showToast({ title: '订单创建异常', icon: 'none' })
          return
        }
        uni.removeStorageSync(CHECKOUT_STORAGE_KEY)
        uni.showToast({ title: '下单成功', icon: 'success', duration: 1500 })
        setTimeout(() => {
          uni.redirectTo({ url: `/pages/order/detail?id=${orderId}` })
        }, 1500)
      } catch (e) {
        console.error('支付/下单失败:', e)
        uni.showToast({ title: e?.message || '操作失败，请重试', icon: 'none' })
      } finally {
        this.submitting = false
      }
    },
    async doWechatPayWithParams(prepayData) {
      const { timeStamp, nonceStr, package: packageVal, signType, paySign } = prepayData
      return new Promise((resolve, reject) => {
        uni.requestPayment({
          provider: 'wxpay',
          timeStamp: String(timeStamp),
          nonceStr,
          package: packageVal,
          signType: signType || 'RSA',
          paySign,
          success: () => resolve(),
          fail: (err) => {
            if (err.errMsg && err.errMsg.includes('cancel')) {
              uni.showToast({ title: '已取消支付', icon: 'none' })
              reject(new Error('用户取消支付'))
            } else {
              uni.showToast({ title: err.errMsg || '支付失败', icon: 'none' })
              reject(err)
            }
          }
        })
      })
    }
  }
}
</script>

<style scoped>
.checkout-page {
  min-height: 100vh;
  background-color: #F5F6FA;
}

.custom-navbar {
  position: relative;
  z-index: 1000;
}

.navbar-fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
}

.navbar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20rpx;
  box-sizing: border-box;
}

.navbar-left {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.navbar-title {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.navbar-title-text {
  font-size: 32rpx;
  font-weight: 500;
  color: #333;
}

.navbar-right {
  width: 60rpx;
  flex-shrink: 0;
}

.checkout-content {
  padding-bottom: 180rpx;
}

.section {
  background: #fff;
  margin: 24rpx;
  border-radius: 20rpx;
  padding: 32rpx 28rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 24rpx;
}

.order-goods-info {
  margin-bottom: 20rpx;
}

.goods-name {
  font-size: 30rpx;
  color: #333;
  font-weight: 500;
  margin-right: 8rpx;
}

.goods-count {
  font-size: 28rpx;
  color: #909399;
}

.order-amount-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 20rpx;
  border-top: 1rpx solid #F0F0F0;
}

.amount-label {
  font-size: 28rpx;
  color: #666;
}

.amount-value {
  font-size: 36rpx;
  font-weight: 600;
  color: #FF4D4F;
}

.payment-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28rpx 24rpx;
  border-radius: 16rpx;
  background: #F8F9FA;
  margin-bottom: 20rpx;
  border: 2rpx solid transparent;
  transition: all 0.2s;
}

.payment-item:last-child {
  margin-bottom: 0;
}

.payment-item-active {
  background: #E8F8F0;
  border-color: #20CB6B;
}

.payment-left {
  display: flex;
  align-items: center;
  flex: 1;
}

.payment-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 24rpx;
}

.wechat-icon {
  background: linear-gradient(135deg, #07C160 0%, #06AD56 100%);
}

.icon-text {
  font-size: 32rpx;
  font-weight: 600;
  color: #fff;
}

.cod-icon {
  background: #E8F8F0;
}

.payment-info {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.payment-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}

.payment-desc {
  font-size: 24rpx;
  color: #909399;
  line-height: 1.4;
}

.radio-outer {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  border: 2rpx solid #dcdfe6;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.radio-checked {
  border-color: #20CB6B;
  background: #20CB6B;
}

.radio-inner {
  width: 12rpx;
  height: 12rpx;
  border-radius: 50%;
  background: #fff;
}

.checkout-footer {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: 120rpx;
  padding-bottom: env(safe-area-inset-bottom);
  background: #fff;
  border-top: 1rpx solid #F0F0F0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 32rpx;
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}

.footer-left {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.footer-label {
  font-size: 24rpx;
  color: #909399;
}

.footer-amount {
  font-size: 40rpx;
  font-weight: 600;
  color: #FF4D4F;
}

.pay-btn {
  width: 50%;
  height: 88rpx;
  line-height: 88rpx;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  color: #fff;
  font-size: 32rpx;
  font-weight: 600;
  border-radius: 50rpx;
  border: none;
}

.pay-btn::after {
  border: none;
}

.pay-btn-disabled {
  opacity: 0.7;
}
</style>
