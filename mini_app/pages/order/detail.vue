<template>
  <view class="order-detail-page">
    <!-- 自定义导航栏 -->
    <view class="custom-navbar">
      <view :style="{ height: statusBarHeight + 'px' }"></view>
      <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
        <view class="navbar-left" @click="goBack">
          <uni-icons type="left" size="20" color="#333"></uni-icons>
        </view>
        <view class="navbar-title">订单详情</view>
        <view class="navbar-right"></view>
      </view>
    </view>

    <scroll-view class="detail-scroll" scroll-y v-if="orderDetail">
      <!-- 订单状态 -->
      <view class="status-section">
        <view class="status-icon">
          <uni-icons :type="getStatusIcon(orderDetail.order?.status)" size="60" :color="getStatusColor(orderDetail.order?.status)"></uni-icons>
        </view>
        <text class="status-text">{{ formatStatus(orderDetail.order?.status) }}</text>
        <text class="order-number">订单编号：{{ orderDetail.order?.order_number }}</text>
      </view>

      <!-- 收货地址 -->
      <view class="section address-section" v-if="orderDetail.address">
        <view class="section-title">收货地址</view>
        <view class="address-content">
          <text class="address-name">{{ orderDetail.address.name }}</text>
          <text class="address-contact">{{ orderDetail.address.contact }} {{ orderDetail.address.phone }}</text>
          <text class="address-detail">{{ orderDetail.address.address }}</text>
        </view>
      </view>

      <!-- 商品列表 -->
      <view class="section goods-section">
        <view class="section-title">商品信息</view>
        <view class="goods-list">
          <view 
            class="goods-item" 
            v-for="(item, index) in orderDetail.order_items" 
            :key="index"
          >
            <image :src="item.image || defaultImage" class="goods-image" mode="aspectFill" />
            <view class="goods-info">
              <text class="goods-name">{{ item.product_name }}</text>
              <text class="goods-spec" v-if="item.spec_name">{{ item.spec_name }}</text>
              <view class="goods-bottom">
                <text class="goods-price">¥{{ formatMoney(item.unit_price) }}</text>
                <text class="goods-qty">× {{ item.quantity }}</text>
              </view>
            </view>
            <view class="goods-subtotal">
              <text>¥{{ formatMoney(item.subtotal) }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 金额明细 -->
      <view class="section amount-section">
        <view class="section-title">金额明细</view>
        <view class="amount-row">
          <text>商品金额</text>
          <text>¥{{ formatMoney(orderDetail.order?.goods_amount) }}</text>
        </view>
        <view class="amount-row">
          <text>配送费</text>
          <text>¥{{ formatMoney(orderDetail.order?.delivery_fee) }}</text>
        </view>
        <view class="amount-row" v-if="orderDetail.order?.points_discount > 0">
          <text>积分抵扣</text>
          <text class="discount">-¥{{ formatMoney(orderDetail.order?.points_discount) }}</text>
        </view>
        <view class="amount-row" v-if="orderDetail.order?.coupon_discount > 0">
          <text>优惠券</text>
          <text class="discount">-¥{{ formatMoney(orderDetail.order?.coupon_discount) }}</text>
        </view>
        <view class="amount-row total">
          <text>实付金额</text>
          <text class="total-amount">¥{{ formatMoney(orderDetail.order?.total_amount) }}</text>
        </view>
      </view>

      <!-- 订单信息 -->
      <view class="section info-section">
        <view class="section-title">订单信息</view>
        <view class="info-row">
          <text class="info-label">下单时间</text>
          <text class="info-value">{{ formatDate(orderDetail.order?.created_at) }}</text>
        </view>
        <view class="info-row" v-if="orderDetail.order?.remark">
          <text class="info-label">订单备注</text>
          <text class="info-value">{{ orderDetail.order?.remark }}</text>
        </view>
      </view>

      <!-- 销售员信息 -->
      <view class="section sales-section" v-if="orderDetail.sales_employee">
        <view class="section-title">销售员</view>
        <view class="sales-content">
          <view class="sales-info">
            <text class="sales-name">{{ orderDetail.sales_employee.name || orderDetail.sales_employee.employee_code }}</text>
            <text class="sales-code" v-if="orderDetail.sales_employee.employee_code">
              工号：{{ orderDetail.sales_employee.employee_code }}
            </text>
          </view>
          <view 
            class="contact-btn" 
            v-if="orderDetail.sales_employee.phone"
            @click="contactSales"
          >
            <uni-icons type="phone" size="18" color="#20CB6B"></uni-icons>
            <text>联系销售员</text>
          </view>
        </view>
      </view>
    </scroll-view>

    <view class="loading" v-else>
      <text>加载中...</text>
    </view>
  </view>
</template>

<script>
import { getOrderDetail } from '../../api/index.js'

export default {
  data() {
    return {
      statusBarHeight: 0,
      navBarHeight: 44,
      orderDetail: null,
      orderId: 0,
      token: '',
      defaultImage: '/static/default-product.png'
    }
  },
  onLoad(options) {
    const systemInfo = uni.getSystemInfoSync()
    this.statusBarHeight = systemInfo.statusBarHeight || 0
    
    this.token = uni.getStorageSync('miniUserToken')
    this.orderId = parseInt(options.id) || 0
    
    if (!this.orderId) {
      uni.showToast({
        title: '订单ID无效',
        icon: 'none'
      })
      setTimeout(() => {
        uni.navigateBack()
      }, 1500)
      return
    }
    
    this.loadOrderDetail()
  },
  methods: {
    goBack() {
      uni.navigateBack()
    },
    async loadOrderDetail() {
      try {
        uni.showLoading({ title: '加载中...' })
        const res = await getOrderDetail(this.token, this.orderId)
        if (res && res.code === 200 && res.data) {
          this.orderDetail = res.data
        } else {
          uni.showToast({
            title: res?.message || '获取订单详情失败',
            icon: 'none'
          })
          setTimeout(() => {
            uni.navigateBack()
          }, 1500)
        }
      } catch (error) {
        console.error('获取订单详情失败:', error)
        uni.showToast({
          title: '获取订单详情失败',
          icon: 'none'
        })
        setTimeout(() => {
          uni.navigateBack()
        }, 1500)
      } finally {
        uni.hideLoading()
      }
    },
    contactSales() {
      if (!this.orderDetail?.sales_employee?.phone) {
        uni.showToast({
          title: '销售员联系方式不可用',
          icon: 'none'
        })
        return
      }
      
      uni.makePhoneCall({
        phoneNumber: this.orderDetail.sales_employee.phone,
        fail: (err) => {
          console.error('拨打电话失败:', err)
          uni.showToast({
            title: '拨打电话失败',
            icon: 'none'
          })
        }
      })
    },
    formatStatus(status) {
      const statusMap = {
        'pending': '待配送',
        'pending_delivery': '待配送',
        'delivering': '配送中',
        'delivered': '已送达',
        'shipped': '已送达',
        'paid': '已收款',
        'completed': '已收款',
        'cancelled': '已取消'
      }
      return statusMap[status] || status
    },
    getStatusIcon(status) {
      const iconMap = {
        'pending': 'shop',
        'pending_delivery': 'shop',
        'delivering': 'car',
        'delivered': 'checkmarkempty',
        'shipped': 'checkmarkempty',
        'paid': 'wallet',
        'completed': 'wallet',
        'cancelled': 'close'
      }
      return iconMap[status] || 'shop'
    },
    getStatusColor(status) {
      const colorMap = {
        'pending': '#ff4d4f',
        'pending_delivery': '#ff4d4f',
        'delivering': '#1890ff',
        'delivered': '#fa8c16',
        'shipped': '#fa8c16',
        'paid': '#52c41a',
        'completed': '#52c41a',
        'cancelled': '#999'
      }
      return colorMap[status] || '#666'
    },
    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hour = String(date.getHours()).padStart(2, '0')
      const minute = String(date.getMinutes()).padStart(2, '0')
      return `${year}-${month}-${day} ${hour}:${minute}`
    },
    formatMoney(amount) {
      if (amount === null || amount === undefined) return '0.00'
      return Number(amount).toFixed(2)
    }
  }
}
</script>

<style scoped>
.order-detail-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.custom-navbar {
  background-color: #fff;
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
}

.navbar-left {
  width: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.navbar-title {
  flex: 1;
  text-align: center;
  font-size: 36rpx;
  font-weight: 600;
  color: #333;
}

.navbar-right {
  width: 60rpx;
}

.detail-scroll {
  height: calc(100vh - var(--status-bar-height, 0px) - 88rpx);
  margin-top: calc(var(--status-bar-height, 0px) + 88rpx);
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  height: calc(100vh - var(--status-bar-height, 0px) - 88rpx);
  margin-top: calc(var(--status-bar-height, 0px) + 88rpx);
  font-size: 28rpx;
  color: #999;
}

.status-section {
  background-color: #fff;
  padding: 60rpx 40rpx;
  text-align: center;
  margin-bottom: 20rpx;
}

.status-icon {
  margin-bottom: 20rpx;
}

.status-text {
  display: block;
  font-size: 36rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 16rpx;
}

.order-number {
  display: block;
  font-size: 24rpx;
  color: #999;
}

.section {
  background-color: #fff;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 24rpx;
}

.address-content {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.address-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
}

.address-contact {
  font-size: 26rpx;
  color: #666;
}

.address-detail {
  font-size: 26rpx;
  color: #666;
  line-height: 1.6;
}

.goods-list {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.goods-item {
  display: flex;
  gap: 20rpx;
}

.goods-image {
  width: 120rpx;
  height: 120rpx;
  border-radius: 12rpx;
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
  margin-bottom: 8rpx;
}

.goods-spec {
  font-size: 24rpx;
  color: #999;
  margin-bottom: 12rpx;
}

.goods-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.goods-price {
  font-size: 26rpx;
  color: #ff4d4f;
}

.goods-qty {
  font-size: 24rpx;
  color: #999;
}

.goods-subtotal {
  display: flex;
  align-items: center;
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
}

.amount-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16rpx 0;
  font-size: 28rpx;
  color: #666;
}

.amount-row.total {
  border-top: 1px solid #f0f0f0;
  margin-top: 16rpx;
  padding-top: 24rpx;
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}

.discount {
  color: #52c41a;
}

.total-amount {
  color: #ff4d4f;
  font-size: 36rpx;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 16rpx 0;
  font-size: 28rpx;
}

.info-label {
  color: #666;
  min-width: 160rpx;
}

.info-value {
  flex: 1;
  text-align: right;
  color: #333;
}

.sales-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sales-info {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  flex: 1;
}

.sales-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
}

.sales-code {
  font-size: 24rpx;
  color: #999;
}

.contact-btn {
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 16rpx 32rpx;
  background-color: #f0f9f4;
  border-radius: 8rpx;
  font-size: 26rpx;
  color: #20CB6B;
}
</style>




