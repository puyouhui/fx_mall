<template>
  <view class="order-list-page">
    <!-- 自定义导航栏 - 绿色背景 -->
    <view class="custom-navbar">
      <view class="navbar-fixed" style="background-color: #20CB6B;">
        <view :style="{ height: statusBarHeight + 'px' }"></view>
        <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
          <view class="navbar-left" @click="goBack">
            <uni-icons type="left" size="20" color="#fff"></uni-icons>
          </view>
          <view class="navbar-title">
            <text class="navbar-title-text">我的订单</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 状态标签 -->
    <view class="status-tabs" :style="{ top: navbarTotalHeight + 'px' }">
      <view 
        class="status-tab" 
        :class="{ active: currentStatus === '' }"
        @click="switchStatus('')"
      >
        全部
      </view>
      <view 
        class="status-tab" 
        :class="{ active: currentStatus === 'pending_delivery' }"
        @click="switchStatus('pending_delivery')"
      >
        待配送
      </view>
      <view 
        class="status-tab" 
        :class="{ active: currentStatus === 'delivering' }"
        @click="switchStatus('delivering')"
      >
        配送中
      </view>
      <view 
        class="status-tab" 
        :class="{ active: currentStatus === 'delivered' }"
        @click="switchStatus('delivered')"
      >
        已送达
      </view>
      <view 
        class="status-tab" 
        :class="{ active: currentStatus === 'paid' }"
        @click="switchStatus('paid')"
      >
        已完成
      </view>
    </view>

    <!-- 订单列表 -->
    <scroll-view 
      class="order-scroll" 
      :style="scrollViewStyle"
      scroll-y 
      @scrolltolower="loadMore"
      :lower-threshold="100"
      :enable-back-to-top="true"
    >
      <view class="order-list" v-if="orders.length > 0">
        <view 
          class="order-item" 
          v-for="order in orders" 
          :key="order.id"
          @click="goToDetail(order.id)"
        >
          <view class="order-header">
            <text class="order-title">{{ getAddressName(order) }}</text>
            <text class="order-status" :class="getStatusClass(order.status)">
              {{ formatStatus(order.status) }}
            </text>
          </view>
          <view class="order-info">
            <view class="order-info-left">
              <text class="order-time">{{ formatDate(order.created_at) }}</text>
              <text class="order-number">订单编号：{{ order.order_number }}</text>
            </view>
            <text class="order-amount">¥{{ formatMoney(order.total_amount) }}</text>
          </view>
          <view class="order-footer">
            <view class="item-count-wrapper">
              <view class="goods-images-stack" v-if="getOrderImages(order).length > 0">
                <view 
                  class="goods-image-item" 
                  v-for="(img, idx) in getOrderImages(order)" 
                  :key="idx"
                  :style="{ zIndex: getOrderImages(order).length - idx }"
                >
                  <image :src="img || '/static/default-product.png'" class="goods-thumb" mode="aspectFill"></image>
                </view>
              </view>
              <text class="item-count">{{ order.item_count || 0 }} 件商品</text>
            </view>
            <view class="action-buttons">
              <text class="action-btn" @click.stop="goToDetail(order.id)">查看详情</text>
            </view>
          </view>
        </view>
      </view>
      <view class="empty-state" v-else-if="!loading">
        <text class="empty-text">暂无订单</text>
      </view>
      <view class="loading-more" v-if="loadingMore">
        <text>加载中...</text>
      </view>
      <view class="no-more" v-if="hasMore === false && orders.length > 0">
        <text>没有更多了</text>
      </view>
    </scroll-view>
  </view>
</template>

<script>
import { getUserOrders } from '../../api/index.js'

export default {
  data() {
    return {
      statusBarHeight: 0,
      navBarHeight: 44,
      navbarTotalHeight: 0,
      tabsHeight: 82, // 状态标签高度（rpx）
      windowHeight: 0, // 窗口高度
      orders: [],
      currentStatus: '',
      loading: false,
      loadingMore: false,
      pageNum: 1,
      pageSize: 10,
      hasMore: true,
      token: ''
    }
  },
  onLoad(options) {
    // 获取状态栏高度和屏幕信息
    const systemInfo = uni.getSystemInfoSync()
    this.statusBarHeight = systemInfo.statusBarHeight || 0
    this.windowHeight = systemInfo.windowHeight || 0
    
    // 获取胶囊按钮信息，计算导航栏高度
    this.getMenuButtonInfo()
    
    // 计算导航栏总高度（状态栏 + 导航栏内容）
    this.navbarTotalHeight = this.statusBarHeight + this.navBarHeight
    
    // 获取token
    this.token = uni.getStorageSync('miniUserToken')
    
    // 如果有传入状态，设置当前状态
    if (options.status) {
      this.currentStatus = options.status
    }
    
    this.loadOrders()
  },
  onReady() {
    // 页面渲染完成后，重新获取窗口信息以确保准确
    const systemInfo = uni.getSystemInfoSync()
    this.windowHeight = systemInfo.windowHeight || 0
  },
  methods: {
    goBack() {
      uni.navigateBack()
    },
    switchStatus(status) {
      if (this.currentStatus === status) return
      this.currentStatus = status
      this.pageNum = 1
      this.orders = []
      this.hasMore = true
      this.loadOrders()
    },
    async loadOrders() {
      if (this.loading || this.loadingMore) return
      
      if (this.pageNum === 1) {
        this.loading = true
      } else {
        this.loadingMore = true
      }
      
      try {
        const params = {
          pageNum: this.pageNum,
          pageSize: this.pageSize
        }
        if (this.currentStatus) {
          params.status = this.currentStatus
        }
        
        const res = await getUserOrders(this.token, params)
        if (res && res.code === 200 && res.data) {
          const newOrders = res.data.list || []
          if (this.pageNum === 1) {
            this.orders = newOrders
          } else {
            this.orders = [...this.orders, ...newOrders]
          }
          
          this.hasMore = newOrders.length >= this.pageSize
          if (this.hasMore) {
            this.pageNum++
          }
        } else {
          uni.showToast({
            title: res?.message || '获取订单列表失败',
            icon: 'none'
          })
        }
      } catch (error) {
        console.error('获取订单列表失败:', error)
        uni.showToast({
          title: '获取订单列表失败',
          icon: 'none'
        })
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    loadMore() {
      if (this.hasMore && !this.loadingMore) {
        this.loadOrders()
      }
    },
    goToDetail(orderId) {
      uni.navigateTo({
        url: `/pages/order/detail?id=${orderId}`
      })
    },
    formatStatus(status) {
      const statusMap = {
        'pending': '待配送',
        'pending_delivery': '待配送',
        'pending_pickup': '待配送',
        'delivering': '配送中',
        'delivered': '已送达',
        'shipped': '已送达',
        'paid': '已完成',
        'completed': '已完成',
        'cancelled': '已取消'
      }
      return statusMap[status] || status
    },
    getStatusClass(status) {
      const classMap = {
        'pending': 'status-green',
        'pending_delivery': 'status-green',
        'pending_pickup': 'status-yellow',
        'delivering': 'status-green',
        'delivered': 'status-green',
        'shipped': 'status-green',
        'paid': 'status-green',
        'completed': 'status-green',
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
    // 获取地址名称
    getAddressName(order) {
      // 优先使用 address.name
      if (order.address && order.address.name) {
        return order.address.name
      }
      // 其次使用 address_name
      if (order.address_name) {
        return order.address_name
      }
      // 如果都没有，返回默认值
      return '未设置地址'
    },
    // 获取订单商品图片（最多显示3张）
    getOrderImages(order) {
      if (!order.order_items || !Array.isArray(order.order_items)) {
        return []
      }
      // 获取前3张商品图片
      return order.order_items
        .filter(item => item.image)
        .slice(0, 3)
        .map(item => item.image)
    },
    // 获取胶囊按钮信息并计算导航栏高度
    getMenuButtonInfo() {
      try {
        // #ifndef H5 || APP-PLUS || MP-ALIPAY
        // 获取胶囊的位置信息
        const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
        // 计算导航栏高度
        this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight);
        // #endif
      } catch (error) {
        console.error('获取胶囊按钮信息失败:', error);
      }
    }
  },
  computed: {
    // 计算滚动区域样式
    scrollViewStyle() {
      if (!this.windowHeight || !this.navbarTotalHeight) {
        return {
          height: '0px',
          marginTop: '0px'
        }
      }
      
      // 获取屏幕宽度用于rpx转px
      const systemInfo = uni.getSystemInfoSync()
      const screenWidth = systemInfo.windowWidth || 375
      
      // 状态标签高度：rpx转px
      const tabsHeightPx = (this.tabsHeight / 750) * screenWidth
      
      // 计算可用高度：窗口高度 - 导航栏高度 - 状态标签高度
      const availableHeight = this.windowHeight - this.navbarTotalHeight - tabsHeightPx
      
      // 计算上边距：导航栏高度 + 状态标签高度
      const marginTop = this.navbarTotalHeight + tabsHeightPx
      
      return {
        height: Math.max(0, availableHeight) + 'px',
        marginTop: marginTop + 'px'
      }
    }
  }
}
</script>

<style scoped>
.order-list-page {
  background-color: #f5f5f5;
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

.status-tabs {
  display: flex;
  background-color: #fff;
  padding: 10rpx 0;
  border-bottom: 1px solid #eee;
  position: fixed;
  left: 0;
  right: 0;
  z-index: 999;
  height: 82rpx;
  box-sizing: border-box;
}

.status-tab {
  flex: 1;
  text-align: center;
  font-size: 28rpx;
  color: #666;
  padding: 10rpx 0;
  position: relative;
}

.status-tab.active {
  color: #20CB6B;
  font-weight: 600;
}

.status-tab.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 60rpx;
  height: 4rpx;
  background-color: #20CB6B;
  border-radius: 2rpx;
}

.order-scroll {
  width: 100%;
  box-sizing: border-box;
}

.order-list {
  padding: 20rpx;
}

.order-item {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.order-title {
  flex: 1;
  font-size: 28rpx;
  font-weight: 500;
  color: #333;
}

.order-number {
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
}

.order-status {
  font-size: 24rpx;
  font-weight: 600;
}

.status-yellow {
  color: #faad14;
}

.status-green {
  color: #20CB6B;
}

.status-cancelled {
  color: #999;
}

.order-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.order-info-left {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.order-time {
  font-size: 26rpx;
  color: #999;
}

.order-amount {
  font-size: 30rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 16rpx;
  border-top: 1px solid #f0f0f0;
}

.item-count-wrapper {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.goods-images-stack {
  display: flex;
  align-items: center;
  position: relative;
  height: 48rpx;
}

.goods-image-item {
  position: relative;
  width: 48rpx;
  height: 48rpx;
  border-radius: 6rpx;
  overflow: hidden;
  border: 2rpx solid #fff;
  background: #f5f5f5;
  margin-left: -12rpx;
  overflow: hidden;
  box-shadow: 0 2rpx 4rpx rgba(0, 0, 0, 0.1);
}

.goods-image-item:first-child {
  margin-left: 0;
}

.goods-thumb {
  width: 100%;
  height: 100%;
}

.item-count {
  font-size: 26rpx;
  color: #666;
}

.action-buttons {
  display: flex;
  gap: 16rpx;
}

.action-btn {
  padding: 10rpx 28rpx;
  font-size: 24rpx;
  color: #20CB6B;
  border: 1px solid #20CB6B;
  border-radius: 8rpx;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 100rpx 0;
}

.empty-text {
  font-size: 28rpx;
  color: #999;
}

.loading-more,
.no-more {
  text-align: center;
  padding: 40rpx 0;
  font-size: 24rpx;
  color: #999;
}
</style>

