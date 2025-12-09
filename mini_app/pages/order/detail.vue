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

    <scroll-view 
      class="detail-scroll" 
      scroll-y 
      v-if="orderDetail"
      :style="{ 
        height: `calc(100vh - ${statusBarHeight + navBarHeight}px)`,
        marginTop: `${statusBarHeight + navBarHeight}px`,
        paddingBottom: '40rpx'
      }"
    >
      <!-- 订单状态 -->
      <view class="status-section">
        <view class="status-icon">
          <uni-icons :type="getStatusIcon(orderDetail.order?.status)" size="60" :color="getStatusColor(orderDetail.order?.status)"></uni-icons>
        </view>
        <text class="status-text">{{ formatStatus(orderDetail.order?.status) }}</text>
        <text class="order-number">订单编号：{{ orderDetail.order?.order_number }}</text>
      </view>

      <!-- 地图（仅在配送中状态显示，配送员取货后才显示） -->
      <view class="section map-section" v-if="showMap">
        <view class="section-title">配送地图</view>
        <map
          :latitude="mapCenter.latitude"
          :longitude="mapCenter.longitude"
          :markers="mapMarkers"
          :scale="14"
          class="map-container"
          :show-location="true"
          :enable-zoom="true"
          :enable-scroll="true"
        ></map>
      </view>

      <!-- 配送员信息（仅在待取货或配送中状态显示） -->
      <view class="section delivery-section" v-if="orderDetail.delivery_employee && (orderDetail.delivery_employee.id || orderDetail.delivery_employee.employee_code)">
        <view class="section-title">配送员信息</view>
        <view class="delivery-content">
          <view class="delivery-info">
            <text class="delivery-name">{{ orderDetail.delivery_employee.name || orderDetail.delivery_employee.employee_code }}</text>
            <text class="delivery-code" v-if="orderDetail.delivery_employee.employee_code">
              工号：{{ orderDetail.delivery_employee.employee_code }}
            </text>
          </view>
          <view 
            class="contact-btn" 
            v-if="orderDetail.delivery_employee.phone"
            @click="contactDelivery"
          >
            <uni-icons type="phone" size="18" color="#20CB6B"></uni-icons>
            <text>联系配送员</text>
          </view>
        </view>
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

    <view 
      class="loading" 
      v-else
      :style="{ 
        height: `calc(100vh - ${statusBarHeight + navBarHeight}px)`,
        marginTop: `${statusBarHeight + navBarHeight}px`
      }"
    >
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
      defaultImage: '/static/default-product.png',
      mapCenter: {
        latitude: 39.90864,
        longitude: 116.39750
      },
      mapMarkers: []
    }
  },
  computed: {
    showMap() {
      const status = this.orderDetail?.order?.status
      // 地图只在配送中状态显示（配送员取货后才显示）
      return status === 'delivering'
    }
  },
  onLoad(options) {
    const systemInfo = uni.getSystemInfoSync()
    this.statusBarHeight = systemInfo.statusBarHeight || 0
    
    // 计算导航栏高度（状态栏高度 + 导航栏内容高度）
    try {
      const menuButtonInfo = uni.getMenuButtonBoundingClientRect()
      this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight)
    } catch (e) {
      this.navBarHeight = 44
    }
    
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
          // 调试：打印订单详情
          console.log('订单详情:', JSON.stringify(this.orderDetail, null, 2))
          console.log('订单状态:', this.orderDetail?.order?.status)
          console.log('配送员信息:', this.orderDetail?.delivery_employee)
          // 初始化地图
          this.initMap()
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
    initMap() {
      if (!this.orderDetail?.address) return
      
      const address = this.orderDetail.address
      if (address.latitude && address.longitude) {
        // 设置地图中心点为收货地址
        this.mapCenter = {
          latitude: address.latitude,
          longitude: address.longitude
        }
        
        // 添加收货地址标记
        this.mapMarkers = [{
          id: 1,
          latitude: address.latitude,
          longitude: address.longitude,
          title: '收货地址',
          width: 30,
          height: 30,
          callout: {
            content: address.name || '收货地址',
            color: '#333',
            fontSize: 12,
            borderRadius: 4,
            bgColor: '#fff',
            padding: 8,
            display: 'ALWAYS'
          }
        }]
      }
    },
    contactDelivery() {
      if (!this.orderDetail?.delivery_employee?.phone) {
        uni.showToast({
          title: '配送员联系方式不可用',
          icon: 'none'
        })
        return
      }
      
      uni.makePhoneCall({
        phoneNumber: this.orderDetail.delivery_employee.phone,
        fail: (err) => {
          console.error('拨打电话失败:', err)
          uni.showToast({
            title: '拨打电话失败',
            icon: 'none'
          })
        }
      })
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
        'pending': '订单正在中心仓库分拣，请耐心等待',
        'pending_delivery': '订单正在中心仓库分拣，请耐心等待',
        'pending_pickup': '中心分拣完成，待配送',
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
  width: 100%;
  height: 100vh;
  background: linear-gradient(180deg, #f8f9fa 0%, #f5f5f5 100%);
  overflow: hidden;
}

.custom-navbar {
  background-color: #fff;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.navbar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 30rpx;
}

.navbar-left {
  width: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 10rpx;
}

.navbar-title {
  flex: 1;
  text-align: center;
  font-size: 36rpx;
  font-weight: 600;
  color: #20253A;
}

.navbar-right {
  width: 60rpx;
}

.detail-scroll {
  width: 100%;
  box-sizing: border-box;
  padding: 0 20rpx;
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  font-size: 28rpx;
  color: #999;
}

.status-section {
  background: linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%);
  padding: 60rpx 40rpx 50rpx;
  text-align: center;
  margin: 20rpx 0;
  border-radius: 24rpx;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.06);
  position: relative;
  overflow: hidden;
}

.status-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 6rpx;
  background: linear-gradient(90deg, #20CB6B 0%, #1AB85A 100%);
}

.status-icon {
  margin-bottom: 24rpx;
  display: flex;
  justify-content: center;
  align-items: center;
}

.status-text {
  display: block;
  font-size: 36rpx;
  font-weight: 700;
  color: #20253A;
  margin-bottom: 20rpx;
  letter-spacing: 0.5rpx;
  line-height: 1.6;
  padding: 0 20rpx;
  word-break: break-all;
}

.order-number {
  display: block;
  font-size: 26rpx;
  color: #8C92A4;
  background-color: #f5f5f5;
  padding: 12rpx 24rpx;
  border-radius: 20rpx;
  display: inline-block;
  margin-top: 8rpx;
}

.section {
  background-color: #fff;
  padding: 32rpx 30rpx;
  margin-bottom: 24rpx;
  box-sizing: border-box;
  border-radius: 20rpx;
  box-shadow: 0 2rpx 16rpx rgba(0, 0, 0, 0.04);
  border: 1rpx solid #f0f0f0;
}

.section:last-child {
  margin-bottom: 40rpx;
}

.section-title {
  font-size: 34rpx;
  font-weight: 700;
  color: #20253A;
  margin-bottom: 28rpx;
  display: flex;
  align-items: center;
  position: relative;
  padding-bottom: 16rpx;
}

.section-title::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  width: 60rpx;
  height: 4rpx;
  background: linear-gradient(90deg, #20CB6B 0%, #1AB85A 100%);
  border-radius: 2rpx;
}

.address-content {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 20rpx;
  background: linear-gradient(135deg, #F0FDF6 0%, #E8F8F0 100%);
  border-radius: 16rpx;
  border: 1rpx solid #E0F5E8;
}

.address-name {
  font-size: 32rpx;
  font-weight: 700;
  color: #20253A;
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.address-name::before {
  content: '📍';
  font-size: 28rpx;
}

.address-contact {
  font-size: 28rpx;
  color: #40475C;
  font-weight: 500;
}

.address-detail {
  font-size: 28rpx;
  color: #666;
  line-height: 1.8;
  word-break: break-all;
}

.goods-list {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.goods-item {
  display: flex;
  gap: 24rpx;
  padding: 20rpx;
  background: #fafafa;
  border-radius: 16rpx;
  border: 1rpx solid #f0f0f0;
  transition: all 0.3s ease;
}

.goods-image {
  width: 140rpx;
  height: 140rpx;
  border-radius: 16rpx;
  background-color: #f5f5f5;
  flex-shrink: 0;
  border: 1rpx solid #e8e8e8;
}

.goods-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-width: 0;
}

.goods-name {
  font-size: 30rpx;
  color: #20253A;
  margin-bottom: 12rpx;
  font-weight: 600;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.goods-spec {
  font-size: 26rpx;
  color: #8C92A4;
  margin-bottom: 16rpx;
  padding: 6rpx 12rpx;
  background-color: #f5f5f5;
  border-radius: 8rpx;
  display: inline-block;
  width: fit-content;
}

.goods-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
}

.goods-price {
  font-size: 28rpx;
  color: #ff4d4f;
  font-weight: 600;
}

.goods-qty {
  font-size: 26rpx;
  color: #8C92A4;
  background-color: #f5f5f5;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}

.goods-subtotal {
  display: flex;
  align-items: center;
  font-size: 32rpx;
  font-weight: 700;
  color: #20253A;
  min-width: 120rpx;
  justify-content: flex-end;
}

.amount-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 0;
  font-size: 30rpx;
  color: #40475C;
  position: relative;
}

.amount-row:not(:last-child)::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1rpx;
  background: linear-gradient(90deg, transparent 0%, #f0f0f0 20%, #f0f0f0 80%, transparent 100%);
}

.amount-row.total {
  border-top: 2rpx solid #f0f0f0;
  margin-top: 20rpx;
  padding-top: 28rpx;
  font-size: 34rpx;
  font-weight: 700;
  color: #20253A;
  background: linear-gradient(135deg, #fafafa 0%, #f5f5f5 100%);
  margin-left: -30rpx;
  margin-right: -30rpx;
  padding-left: 30rpx;
  padding-right: 30rpx;
  border-radius: 0 0 20rpx 20rpx;
}

.amount-row.total::after {
  display: none;
}

.discount {
  color: #20CB6B;
  font-weight: 600;
}

.total-amount {
  color: #ff4d4f;
  font-size: 40rpx;
  font-weight: 700;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 20rpx 0;
  font-size: 30rpx;
  position: relative;
}

.info-row:not(:last-child)::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1rpx;
  background: linear-gradient(90deg, transparent 0%, #f0f0f0 20%, #f0f0f0 80%, transparent 100%);
}

.info-label {
  color: #8C92A4;
  min-width: 180rpx;
  font-weight: 500;
}

.info-value {
  flex: 1;
  text-align: right;
  color: #20253A;
  font-weight: 500;
  word-break: break-all;
}

.sales-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx;
  background: linear-gradient(135deg, #F0FDF6 0%, #E8F8F0 100%);
  border-radius: 16rpx;
  border: 1rpx solid #E0F5E8;
}

.sales-info {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  flex: 1;
}

.sales-name {
  font-size: 32rpx;
  font-weight: 700;
  color: #20253A;
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.sales-name::before {
  content: '👤';
  font-size: 28rpx;
}

.sales-code {
  font-size: 26rpx;
  color: #8C92A4;
  background-color: #fff;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
  display: inline-block;
  width: fit-content;
}

.contact-btn {
  display: flex;
  align-items: center;
  gap: 10rpx;
  padding: 20rpx 36rpx;
  background: linear-gradient(135deg, #20CB6B 0%, #1AB85A 100%);
  border-radius: 12rpx;
  font-size: 28rpx;
  color: #fff;
  font-weight: 600;
  box-shadow: 0 4rpx 12rpx rgba(32, 203, 107, 0.3);
  transition: all 0.3s ease;
}

.contact-btn:active {
  transform: scale(0.95);
  box-shadow: 0 2rpx 8rpx rgba(32, 203, 107, 0.2);
}

.map-section {
  padding: 0;
  overflow: hidden;
}

.map-container {
  width: 100%;
  height: 400rpx;
  margin: 0;
}

.delivery-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx;
  background: linear-gradient(135deg, #F0F9FF 0%, #E0F2FE 100%);
  border-radius: 16rpx;
  border: 1rpx solid #D0E7F8;
}

.delivery-info {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  flex: 1;
}

.delivery-name {
  font-size: 32rpx;
  font-weight: 700;
  color: #20253A;
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.delivery-name::before {
  content: '🚚';
  font-size: 28rpx;
}

.delivery-code {
  font-size: 26rpx;
  color: #8C92A4;
  background-color: #fff;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
  display: inline-block;
  width: fit-content;
}
</style>




