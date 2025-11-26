<template>
  <view class="my-page">
    <!-- 个人信息区域 -->
    <view class="user-section">
      <view class="user-header" v-if="isLoggedIn">
        <view class="avatar-wrapper" @click="goToProfile">
          <image v-if="userInfo.avatar" :src="userInfo.avatar" class="avatar" mode="aspectFill"></image>
          <view v-else class="avatar-placeholder">
            <uni-icons type="person-filled" size="40" color="#fff"></uni-icons>
          </view>
        </view>
        <view class="user-info">
          <text class="user-name">{{ displayName }}</text>
          <view class="user-meta">
            <text class="user-type" :class="userTypeClass">{{ userTypeText }}</text>
          </view>
        </view>
        <view class="edit-btn" @click="goToProfile">
          <uni-icons type="right" size="18" color="#999"></uni-icons>
        </view>
      </view>
      <view class="login-prompt" v-else @click="goToLogin">
        <view class="avatar-wrapper">
          <view class="avatar-placeholder">
            <uni-icons type="person-filled" size="40" color="#fff"></uni-icons>
          </view>
        </view>
        <view class="user-info">
          <text class="login-text">点击登录</text>
          <text class="login-tip">登录后享受更多服务</text>
        </view>
        <view class="edit-btn">
          <uni-icons type="right" size="18" color="#999"></uni-icons>
        </view>
      </view>
    </view>

    <!-- 订单入口 -->
    <view class="order-section">
      <view class="section-header" @click="goToOrderList">
        <text class="section-title">我的订单</text>
        <view class="section-more">
          <text class="more-text">查看全部</text>
          <uni-icons type="right" size="14" color="#999"></uni-icons>
        </view>
      </view>
      <view class="order-tabs">
        <view class="order-tab" @click="goToOrderList('pending_payment')">
          <view class="tab-icon-wrapper">
            <uni-icons type="wallet" size="24" color="#20CB6B"></uni-icons>
            <view class="badge" v-if="orderCounts.pending_payment > 0">{{ orderCounts.pending_payment > 99 ? '99+' : orderCounts.pending_payment }}</view>
          </view>
          <text class="tab-text">待付款</text>
        </view>
        <view class="order-tab" @click="goToOrderList('pending_delivery')">
          <view class="tab-icon-wrapper">
            <uni-icons type="shop" size="24" color="#20CB6B"></uni-icons>
            <view class="badge" v-if="orderCounts.pending_delivery > 0">{{ orderCounts.pending_delivery > 99 ? '99+' : orderCounts.pending_delivery }}</view>
          </view>
          <text class="tab-text">待发货</text>
        </view>
        <view class="order-tab" @click="goToOrderList('pending_receipt')">
          <view class="tab-icon-wrapper">
            <uni-icons type="car" size="24" color="#20CB6B"></uni-icons>
            <view class="badge" v-if="orderCounts.pending_receipt > 0">{{ orderCounts.pending_receipt > 99 ? '99+' : orderCounts.pending_receipt }}</view>
          </view>
          <text class="tab-text">待收货</text>
        </view>
        <view class="order-tab" @click="goToOrderList('completed')">
          <view class="tab-icon-wrapper">
            <uni-icons type="checkmarkempty" size="24" color="#20CB6B"></uni-icons>
          </view>
          <text class="tab-text">已完成</text>
        </view>
      </view>
    </view>

    <!-- 常用功能 -->
    <view class="functions-section">
      <view class="function-group">
        <view class="function-item" @click="goToCoupons">
          <view class="function-icon coupon-icon">
            <uni-icons type="wallet" size="22" color="#20CB6B"></uni-icons>
          </view>
          <text class="function-text">我的优惠券</text>
          <view class="function-right">
            <text class="coupon-count" v-if="couponCount > 0">{{ couponCount }}</text>
            <uni-icons type="right" size="16" color="#ddd"></uni-icons>
          </view>
        </view>
        <view class="function-item" @click="goToCustomerService">
          <view class="function-icon customer-service-icon">
            <uni-icons type="chatbubble" size="22" color="#20CB6B"></uni-icons>
          </view>
          <text class="function-text">我的客服</text>
          <uni-icons type="right" size="16" color="#ddd"></uni-icons>
        </view>
        <view class="function-item" @click="goToAddress">
          <view class="function-icon address-icon">
            <uni-icons type="location" size="22" color="#20CB6B"></uni-icons>
          </view>
          <text class="function-text">收货地址</text>
          <uni-icons type="right" size="16" color="#ddd"></uni-icons>
        </view>
        <view class="function-item" @click="goToProfile">
          <view class="function-icon profile-icon">
            <uni-icons type="person" size="22" color="#20CB6B"></uni-icons>
          </view>
          <text class="function-text">个人资料</text>
          <uni-icons type="right" size="16" color="#ddd"></uni-icons>
        </view>
      </view>
      
      <view class="function-group">
        <view class="function-item" @click="goToSettings">
          <view class="function-icon settings-icon">
            <uni-icons type="gear" size="22" color="#20CB6B"></uni-icons>
          </view>
          <text class="function-text">设置</text>
          <uni-icons type="right" size="16" color="#ddd"></uni-icons>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { getMiniUserInfo, getMiniUserDefaultAddress, getUserCoupons } from '../../api/index.js';

export default {
  data() {
    return {
      userInfo: {},
      isLoggedIn: false,
      defaultAddress: null, // 默认地址
      orderCounts: {
        pending_payment: 0,
        pending_delivery: 0,
        pending_receipt: 0,
        completed: 0
      },
      couponCount: 0
    };
  },
  computed: {
    // 显示名称：有name显示name，否则显示user_code
    displayName() {
      if (this.userInfo.name && this.userInfo.name.trim()) {
        return this.userInfo.name;
      }
      return this.userInfo.user_code ? `用户${this.userInfo.user_code}` : '未设置';
    },
    // 用户类型文本
    userTypeText() {
      if (!this.userInfo.user_type) {
        return '未设置';
      }
      const typeMap = {
        'retail': '普通用户',
        'wholesale': '会员用户'
      };
      return typeMap[this.userInfo.user_type] || '信息未完善';
    },
    // 用户类型样式类
    userTypeClass() {
      return {
        'type-retail': this.userInfo.user_type === 'retail',
        'type-wholesale': this.userInfo.user_type === 'wholesale'
      };
    }
  },
  onLoad() {
    this.checkLoginStatus();
    this.loadUserInfo();
    this.loadOrderCounts();
  },
  // 页面显示时更新用户信息
  onShow() {
    this.checkLoginStatus();
    this.updateUserInfo();
    this.loadDefaultAddress();
    this.loadOrderCounts();
    this.loadCouponCount();
  },
  methods: {
    // 检查登录状态
    checkLoginStatus() {
      const token = uni.getStorageSync('miniUserToken');
      this.isLoggedIn = !!token;
      if (this.isLoggedIn) {
        const storedUserInfo = uni.getStorageSync('miniUserInfo');
        if (storedUserInfo) {
          this.userInfo = storedUserInfo;
        }
      } else {
        this.userInfo = {};
      }
    },
    
    // 更新用户信息
    async updateUserInfo() {
      try {
        const token = uni.getStorageSync('miniUserToken');
        if (!token) {
          this.isLoggedIn = false;
          this.userInfo = {};
          return;
        }

        const res = await getMiniUserInfo(token);
        if (res && res.code === 200 && res.data) {
          // 更新本地存储的用户信息
          uni.setStorageSync('miniUserInfo', res.data);
          if (res.data.unique_id) {
            uni.setStorageSync('miniUserUniqueId', res.data.unique_id);
          }
          this.userInfo = res.data;
          this.isLoggedIn = true;
        }
      } catch (error) {
        console.error('获取用户信息失败:', error);
        // 静默失败，不显示错误提示
      }
    },
    
    // 加载用户信息
    loadUserInfo() {
      const storedUserInfo = uni.getStorageSync('miniUserInfo');
      if (storedUserInfo) {
        this.userInfo = storedUserInfo;
      }
    },
    
    // 加载默认地址
    async loadDefaultAddress() {
      if (!this.isLoggedIn) {
        this.defaultAddress = null;
        return;
      }
      
      try {
        const token = uni.getStorageSync('miniUserToken');
        if (!token) {
          return;
        }
        
        const res = await getMiniUserDefaultAddress(token);
        if (res && res.code === 200 && res.data) {
          this.defaultAddress = res.data;
        } else {
          this.defaultAddress = null;
        }
      } catch (error) {
        console.error('加载默认地址失败:', error);
        this.defaultAddress = null;
      }
    },
    
    // 加载订单数量（模拟数据）
    loadOrderCounts() {
      // TODO: 调用真实API获取订单数量
      // 这里使用模拟数据
      if (this.isLoggedIn) {
        // 模拟订单数量
        this.orderCounts = {
          pending_payment: 2,
          pending_delivery: 1,
          pending_receipt: 3,
          completed: 10
        };
      } else {
        this.orderCounts = {
          pending_payment: 0,
          pending_delivery: 0,
          pending_receipt: 0,
          completed: 0
        };
      }
    },
    
    // 跳转到登录页
    goToLogin() {
      // TODO: 跳转到登录页面
      uni.showToast({
        title: '请先登录',
        icon: 'none'
      });
    },
    
    // 跳转到个人资料页面
    goToProfile() {
      if (!this.isLoggedIn) {
        this.goToLogin();
        return;
      }
      uni.navigateTo({
        url: '/pages/profile/profile'
      });
    },
    
    // 跳转到订单列表
    goToOrderList(status) {
      if (!this.isLoggedIn) {
        this.goToLogin();
        return;
      }
      // TODO: 跳转到订单列表页面
      uni.showToast({
        title: '订单功能开发中',
        icon: 'none'
      });
    },
    
    // 跳转到客服页面
    goToCustomerService() {
      uni.navigateTo({
        url: '/pages/customer-service/customer-service'
      });
    },
    
    // 跳转到收货地址
    goToAddress() {
      if (!this.isLoggedIn) {
        this.goToLogin();
        return;
      }
      uni.navigateTo({
        url: '/pages/address/address'
      });
    },
    
    // 跳转到设置
    goToSettings() {
      uni.navigateTo({
        url: '/pages/settings/settings'
      });
    },
    
    // 跳转到优惠券页面
    goToCoupons() {
      if (!this.isLoggedIn) {
        this.goToLogin();
        return;
      }
      uni.navigateTo({
        url: '/pages/coupons/coupons'
      });
    },
    
    // 加载优惠券数量
    async loadCouponCount() {
      if (!this.isLoggedIn) {
        this.couponCount = 0;
        return;
      }
      
      try {
        const token = uni.getStorageSync('miniUserToken');
        if (!token) {
          return;
        }
        
        const res = await getUserCoupons(token);
        if (res && res.code === 200 && Array.isArray(res.data)) {
          // 统计未使用的优惠券数量
          this.couponCount = res.data.filter(coupon => coupon.status === 'unused').length;
        } else {
          this.couponCount = 0;
        }
      } catch (error) {
        console.error('加载优惠券数量失败:', error);
        this.couponCount = 0;
      }
    }
  }
};
</script>

<style scoped>
.my-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 40rpx;
}

/* 个人信息区域 */
.user-section {
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  padding: 60rpx 30rpx 40rpx;
}

.user-header,
.login-prompt {
  display: flex;
  align-items: center;
}

.avatar-wrapper {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  margin-right: 24rpx;
  overflow: hidden;
  border: 4rpx solid rgba(255, 255, 255, 0.3);
  flex-shrink: 0;
}

.avatar {
  width: 100%;
  height: 100%;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  background-color: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.user-name {
  font-size: 36rpx;
  font-weight: 600;
  color: #fff;
  margin-bottom: 12rpx;
}

.user-meta {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.user-type {
  font-size: 24rpx;
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
  background-color: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.type-retail {
  background-color: rgba(255, 255, 255, 0.25);
}

.type-wholesale {
  background-color: rgba(255, 255, 255, 0.3);
}


.login-text {
  font-size: 36rpx;
  font-weight: 600;
  color: #fff;
  margin-bottom: 8rpx;
}

.login-tip {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.8);
}

.edit-btn {
  padding: 10rpx;
}

/* 订单入口 */
.order-section {
  background-color: #fff;
  margin: -30rpx 20rpx 20rpx;
  border-radius: 20rpx;
  padding: 30rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}

.section-more {
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.more-text {
  font-size: 26rpx;
  color: #999;
}

.order-tabs {
  display: flex;
  justify-content: space-around;
}

.order-tab {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  position: relative;
}

.tab-icon-wrapper {
  position: relative;
  width: 80rpx;
  height: 80rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #F0F9F4;
  border-radius: 50%;
  margin-bottom: 12rpx;
}

.badge {
  position: absolute;
  top: -8rpx;
  right: -8rpx;
  min-width: 32rpx;
  height: 32rpx;
  padding: 0 8rpx;
  background-color: #ff4d4f;
  color: #fff;
  font-size: 20rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2rpx solid #fff;
}

.tab-text {
  font-size: 24rpx;
  color: #666;
}

/* 常用功能 */
.functions-section {
  margin: 0 20rpx;
}

.function-group {
  background-color: #fff;
  border-radius: 20rpx;
  margin-bottom: 20rpx;
  overflow: hidden;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.function-item {
  display: flex;
  align-items: center;
  padding: 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
  transition: background-color 0.2s;
}

.function-item:last-child {
  border-bottom: none;
}

.function-item:active {
  background-color: #f8f8f8;
}

.function-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 24rpx;
}

.customer-service-icon {
  background-color: #E8F8F0;
}

.address-icon {
  background-color: #E8F8F0;
}

.profile-icon {
  background-color: #E8F8F0;
}

.settings-icon {
  background-color: #E8F8F0;
}

.about-icon {
  background-color: #E8F8F0;
}

.function-text {
  flex: 1;
  font-size: 28rpx;
  color: #333;
}

.function-right {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.coupon-icon {
  background-color: #E8F8F0;
}

.coupon-count {
  font-size: 24rpx;
  color: #ff4d4f;
  font-weight: 500;
}

</style>
