<template>
  <view class="order-detail-page">
    <!-- 自定义导航栏 - 绿色背景 -->
    <view class="custom-navbar">
      <view class="navbar-fixed" style="background-color: #20CB6B;">
        <view :style="{ height: statusBarHeight + 'px' }"></view>
        <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
          <view class="navbar-left" @click="goBack">
            <uni-icons type="left" size="20" color="#fff"></uni-icons>
          </view>
          <view class="navbar-title">
            <text class="navbar-title-text">订单详情</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <view 
      class="detail-content" 
      v-if="orderDetail"
      :style="{ 
        paddingTop: `${statusBarHeight + navBarHeight}px`,
        paddingBottom: canCancelOrder ? '140rpx' : '40rpx'
      }"
    >
      <!-- 地图（仅在配送中状态显示，配送员取货后才显示） -->
      <view class="section map-section" v-if="showMap">
        <map
          :latitude="mapCenter.latitude"
          :longitude="mapCenter.longitude"
          :markers="mapMarkers"
          :scale="mapScale"
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
          <text class="info-label">订单编号</text>
          <text class="info-value">{{ orderDetail.order?.order_number }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">订单状态</text>
          <text class="info-value" :class="getStatusClass(orderDetail.order?.status)">
            {{ formatStatus(orderDetail.order?.status) }}
          </text>
        </view>
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
    </view>

    <!-- 取消订单按钮（仅在配送员接单之前显示） -->
    <view class="cancel-order-footer" v-if="canCancelOrder">
      <view class="cancel-btn" @click="handleCancelOrder">
        <text>取消订单</text>
      </view>
    </view>

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
import { getOrderDetail, getDeliveryEmployeeLocation, cancelOrder } from '../../api/index.js'

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
      mapMarkers: [],
      mapScale: 14 // 地图缩放级别
    }
  },
  computed: {
    showMap() {
      const status = this.orderDetail?.order?.status
      // 地图只在配送中状态显示（配送员取货后才显示）
      return status === 'delivering'
    },
    // 是否可以取消订单（配送员接单之前：pending_delivery 或 pending_pickup）
    canCancelOrder() {
      const status = this.orderDetail?.order?.status
      return status === 'pending_delivery' || status === 'pending' || status === 'pending_pickup'
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
          // 如果是配送中状态，加载配送员位置
          if (this.orderDetail.order?.status === 'delivering' && this.orderDetail.delivery_employee?.employee_code) {
            this.loadDeliveryEmployeeLocation()
          }
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
          iconPath: '/static/marker-destination.png', // 目的地标记图标
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
    async loadDeliveryEmployeeLocation() {
      const employeeCode = this.orderDetail?.delivery_employee?.employee_code
      if (!employeeCode) {
        console.log('配送员员工码不存在，无法获取位置')
        return
      }
      
      try {
        const res = await getDeliveryEmployeeLocation(this.token, employeeCode)
        if (res && res.code === 200 && res.data) {
          const location = res.data
          if (location.latitude && location.longitude) {
            // 添加配送员位置标记
            const deliveryMarker = {
              id: 2,
              latitude: location.latitude,
              longitude: location.longitude,
              title: '配送员位置',
              iconPath: '/static/marker-delivery.png', // 配送员标记图标
              width: 30,
              height: 30,
              callout: {
                content: `配送员${location.is_realtime ? '（实时）' : '（历史位置）'}`,
                color: '#333',
                fontSize: 12,
                borderRadius: 4,
                bgColor: location.is_realtime ? '#20CB6B' : '#FFA500',
                padding: 8,
                display: 'ALWAYS'
              }
            }
            
            // 添加到地图标记数组
            this.mapMarkers.push(deliveryMarker)
            
            // 调整地图视野，同时显示收货地址和配送员位置
            if (this.orderDetail?.address?.latitude && this.orderDetail?.address?.longitude) {
              const lat1 = this.orderDetail.address.latitude
              const lng1 = this.orderDetail.address.longitude
              const lat2 = location.latitude
              const lng2 = location.longitude
              
              // 计算中心点
              const centerLat = (lat1 + lat2) / 2
              const centerLng = (lng1 + lng2) / 2
              
              // 计算距离，调整缩放级别
              const distance = this.calculateDistance(lat1, lng1, lat2, lng2)
              let scale = 14
              if (distance > 10000) scale = 11
              else if (distance > 5000) scale = 12
              else if (distance > 2000) scale = 13
              else if (distance > 1000) scale = 14
              else scale = 15
              
              this.mapCenter = {
                latitude: centerLat,
                longitude: centerLng
              }
              
              // 更新地图缩放级别
              this.mapScale = scale
            }
            
            console.log('配送员位置已加载:', location)
          }
        } else {
          console.log('获取配送员位置失败:', res?.message)
        }
      } catch (error) {
        console.error('获取配送员位置失败:', error)
      }
    },
    // 计算两点间距离（米）
    calculateDistance(lat1, lng1, lat2, lng2) {
      const R = 6371000 // 地球半径（米）
      const dLat = (lat2 - lat1) * Math.PI / 180
      const dLng = (lng2 - lng1) * Math.PI / 180
      const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
                Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
                Math.sin(dLng / 2) * Math.sin(dLng / 2)
      const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
      return R * c
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
    getStatusClass(status) {
      const classMap = {
        'pending': 'status-pending',
        'pending_delivery': 'status-pending',
        'pending_pickup': 'status-pending',
        'delivering': 'status-delivering',
        'delivered': 'status-delivered',
        'shipped': 'status-delivered',
        'paid': 'status-paid',
        'completed': 'status-paid',
        'cancelled': 'status-cancelled'
      }
      return classMap[status] || ''
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
    },
    // 取消订单
    async handleCancelOrder() {
      const orderNumber = this.orderDetail?.order?.order_number || ''
      const salesPhone = this.orderDetail?.sales_employee?.phone || ''
      
      // 构建提示内容
      let content = `确定要取消订单吗？\n\n`
      if (salesPhone) {
        content += `如需修改订单，可联系销售员：${salesPhone}\n\n`
      } else if (this.orderDetail?.sales_employee) {
        content += `如需修改订单，可联系销售员修改\n\n`
      }
      content += `取消后订单将无法恢复，是否仍要取消？`
      
      // 显示确认对话框
      const confirmed = await new Promise((resolve) => {
        uni.showModal({
          title: '确认取消订单',
          content: content,
          confirmText: '仍要取消',
          cancelText: '我再想想',
          confirmColor: '#ff4d4f',
          success: (res) => {
            resolve(res.confirm)
          },
          fail: () => {
            resolve(false)
          }
        })
      })
      
      if (!confirmed) {
        return
      }
      
      try {
        uni.showLoading({ title: '取消中...' })
        const res = await cancelOrder(this.token, this.orderId)
        
        if (res && res.code === 200) {
          uni.showToast({
            title: '订单已取消',
            icon: 'success',
            duration: 2000
          })
          
          // 延迟返回，让用户看到成功提示
          setTimeout(() => {
            uni.navigateBack()
          }, 1500)
        } else {
          uni.showToast({
            title: res?.message || '取消订单失败',
            icon: 'none',
            duration: 2000
          })
        }
      } catch (error) {
        console.error('取消订单失败:', error)
        uni.showToast({
          title: '取消订单失败，请重试',
          icon: 'none',
          duration: 2000
        })
      } finally {
        uni.hideLoading()
      }
    }
  }
}
</script>

<style scoped>
.order-detail-page {
  width: 100%;
  min-height: 100vh;
  background: #f5f5f5;
}

.custom-navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
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
  cursor: pointer;
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
  color: #fff;
}

.navbar-right {
  width: 60rpx;
  flex-shrink: 0;
}

.detail-content {
  width: 100%;
  box-sizing: border-box;
  padding: 20rpx;
  min-height: calc(100vh - var(--nav-height, 0px));
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
  border-radius: 8rpx;
  border: 1rpx solid #e8e8e8;
}

.section:last-child {
  margin-bottom: 40rpx;
}

.section-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 24rpx;
}

.address-content {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 20rpx;
  background: #f9f9f9;
  border-radius: 8rpx;
  border: 1rpx solid #e8e8e8;
}

.address-section{
  margin-top: 20rpx !important;
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
  background: #f9f9f9;
  border-radius: 8rpx;
  border: 1rpx solid #e8e8e8;
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

.amount-row:not(:last-child) {
  border-bottom: 1rpx solid #e8e8e8;
}

.amount-row.total {
  border-top: 2rpx solid #e8e8e8;
  margin-top: 20rpx;
  padding-top: 28rpx;
  font-size: 34rpx;
  font-weight: 600;
  color: #333;
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

.info-row:not(:last-child) {
  border-bottom: 1rpx solid #e8e8e8;
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

.status-pending {
  color: #ff4d4f;
}

.status-delivering {
  color: #1890ff;
}

.status-delivered {
  color: #fa8c16;
}

.status-paid {
  color: #52c41a;
}

.status-cancelled {
  color: #999;
}

.sales-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx;
  background: #f9f9f9;
  border-radius: 8rpx;
  border: 1rpx solid #e8e8e8;
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
  background: #20CB6B;
  border-radius: 8rpx;
  font-size: 28rpx;
  color: #fff;
  font-weight: 500;
}

.contact-btn:active {
  background-color: #1AB85A;
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
  background: #f9f9f9;
  border-radius: 8rpx;
  border: 1rpx solid #e8e8e8;
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

.cancel-order-footer {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: #fff;
  padding: 20rpx 20rpx 0 20rpx;
  padding-bottom: calc(env(safe-area-inset-bottom));
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.08);
  z-index: 999;
  box-sizing: border-box;
}

.cancel-btn {
  width: 100%;
  height: 88rpx;
  background-color: #20CB6B;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  font-weight: 500;
  color: #fff;
}

.cancel-btn:active {
  background-color: #1AB85A;
}
</style>




