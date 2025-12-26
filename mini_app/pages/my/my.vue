<template>
  <view class="my-page">
    <!-- 个人信息区域 -->
    <view class="user-section" :style="{ paddingTop: (statusBarHeight + 50) + 'px' }">
      <view class="user-header" v-if="isLoggedIn">
        <view class="avatar-wrapper" @click="goToProfile">
          <image v-if="userInfo.avatar" :src="userInfo.avatar" class="avatar" mode="aspectFill"></image>
          <view v-else class="avatar-placeholder">
            <uni-icons type="person-filled" size="40" color="#fff"></uni-icons>
          </view>
        </view>
        <view class="user-info">
          <view class="user-name-row">
            <text class="user-name">{{ displayName }}</text>
            <text class="user-code" v-if="userInfo.user_code">客户编号：{{ userInfo.user_code }}</text>
          </view>
          <view class="user-meta">
            <view class="user-type" :class="userTypeClass">
              <image v-if="userInfo.user_type === 'wholesale'" src="/static/icon/vip.png" class="user-type-icon"
                mode="aspectFit" @error="handleIconError"></image>
              <image v-else-if="userInfo.user_type === 'retail'" src="/static/icon/zx.png" class="user-type-icon"
                mode="aspectFit" @error="handleIconError"></image>
              <text class="user-type-text">{{ userTypeText }}</text>
            </view>
          </view>
        </view>
        <view class="service-icon" @click="goToCustomerService">
          <image src="/static/icon/zx.png" class="service-icon-image" mode="aspectFit"></image>
        </view>
      </view>
      <view class="login-prompt" v-else @click="goToLogin">
        <view class="avatar-wrapper">
          <view class="avatar-placeholder">
            <uni-icons type="person-filled" size="40" color="#20CB6B"></uni-icons>
          </view>
        </view>
        <view class="user-info">
          <text class="login-text">点击登录</text>
          <text class="login-tip">登录后享受更多服务</text>
        </view>
        <view class="service-icon" @click="goToCustomerService">
          <image src="/static/icon/zx.png" class="service-icon-image" mode="aspectFit"></image>
        </view>
      </view>
    </view>

    <!-- 账户概览卡片 -->
    <view class="account-card">
      <view class="account-item">
        <text class="account-label">余额(元)</text>
        <text class="account-value">{{ userBalance.toFixed(2) }}</text>
      </view>
      <view class="account-item" @click="goToCoupons">
        <text class="account-label">优惠券</text>
        <text class="account-value">{{ couponCount }}</text>
      </view>
      <!-- <view class="account-item">
        <text class="account-label">锁价(货)单</text>
        <text class="account-value">{{ lockedOrdersCount }}</text>
      </view> -->
      <view class="account-item">
        <text class="account-label">积分</text>
        <text class="account-value">{{ userPoints }}</text>
      </view>
    </view>

    <!-- 订单信息 -->
    <view class="order-section">
      <view class="section-header" @click="goToOrderList">
        <text class="section-title">订单信息</text>
        <!-- <view class="section-more">
          <text class="more-text">查看全部</text>
          <uni-icons type="right" size="14" color="#999"></uni-icons>
        </view> -->
      </view>
      <view class="order-tabs">
        <view class="order-tab" @click="goToOrderList('pending_delivery')">
          <view class="tab-icon-wrapper">
            <image src="/static/icon/1.png" class="tab-icon" mode="aspectFit"></image>
            <view class="badge" v-if="orderCounts.pending_delivery > 0">{{ orderCounts.pending_delivery > 99 ? '99+' :
              orderCounts.pending_delivery }}</view>
          </view>
          <text class="tab-text">待配送</text>
        </view>
        <view class="order-tab" @click="goToOrderList('delivering')">
          <view class="tab-icon-wrapper">
            <image src="/static/icon/2.png" class="tab-icon" mode="aspectFit"></image>
            <view class="badge" v-if="orderCounts.delivering > 0">{{ orderCounts.delivering > 99 ? '99+' :
              orderCounts.delivering }}</view>
          </view>
          <text class="tab-text">配送中</text>
        </view>
        <view class="order-tab" @click="goToOrderList('delivered')">
          <view class="tab-icon-wrapper">
            <image src="/static/icon/3.png" class="tab-icon" mode="aspectFit"></image>
            <view class="badge" v-if="orderCounts.delivered > 0">{{ orderCounts.delivered > 99 ? '99+' :
              orderCounts.delivered }}</view>
          </view>
          <text class="tab-text">已送达</text>
        </view>
        <view class="order-tab" @click="goToOrderList()">
          <view class="tab-icon-wrapper">
            <image src="/static/icon/4.png" class="tab-icon" mode="aspectFit"></image>
          </view>
          <text class="tab-text">全部订单</text>
        </view>
      </view>
    </view>


    <!-- 我的功能 -->
    <view class="functions-section">
      <view class="section-header">
        <text class="section-title">常用功能</text>
      </view>
      <view class="functions-grid">
        <view class="function-item" v-for="(func, index) in functions" :key="index" @click="handleFunctionClick(func)">
          <view class="function-icon-wrapper">
            <image :src="func.iconPath || '/static/icon/About.png'" class="function-icon" mode="aspectFit"></image>
            <view class="function-badge" v-if="func.badge">{{ func.badge }}</view>
          </view>
          <text class="function-text">{{ func.name }}</text>
        </view>
      </view>
    </view>



    <!-- 轮播图 -->
    <!-- <view class="carousel-section" v-if="carousels.length > 0">
      <swiper class="carousel-swiper" :indicator-dots="true" :autoplay="true" :interval="3000" :duration="500" circular>
        <swiper-item v-for="(item, index) in carousels" :key="index" @click="handleCarouselClick(item)">
          <image :src="item.image" class="carousel-image" mode="aspectFill"></image>
        </swiper-item>
      </swiper>
    </view> -->

    <!-- 登录弹框组件 -->
    <LoginModal :visible="showLoginModal" @update:visible="showLoginModal = $event" @login-success="handleLoginSuccess"
      @close="handleLoginModalClose" />
  </view>
</template>

<script>
import { getMiniUserInfo, getMiniUserDefaultAddress, getUserCoupons, getUserOrders, getCarousels } from '../../api/index.js';
import LoginModal from '../../components/LoginModal.vue';

export default {
  components: {
    LoginModal
  },
  data() {
    return {
      statusBarHeight: 20, // 状态栏高度（默认值）
      userInfo: {},
      isLoggedIn: false,
      defaultAddress: null,
      orderCounts: {
        pending_delivery: 0,
        delivering: 0,
        delivered: 0
      },
      couponCount: 0,
      userBalance: 0.00,
      lockedOrdersCount: 0,
      userPoints: 0,
      carousels: [],
      showLoginModal: false,
      functions: [
        { name: '地址管理', icon: 'location', iconPath: '/static/icon/address.png', path: '/pages/address/address', color: '#20CB6B' },
        { name: '收藏商品', icon: 'star-filled', iconPath: '/static/icon/favorite.png', path: '/pages/favorites/favorites', color: '#20CB6B' },
        // { name: '我的账单', icon: 'wallet', iconPath: '/static/icon/bills.png', path: '/pages/bill/bill', color: '#20CB6B' },
        { name: '客服与帮助', icon: 'chatbubble', iconPath: '/static/icon/customer_service.png', path: '/pages/customer-service/customer-service', color: '#20CB6B' },
        { name: '供应商合作', icon: 'shop', iconPath: '/static/icon/suppliers.png', path: '/pages/supplier/supplier', color: '#20CB6B' },
        { name: '新品需求', icon: 'star', iconPath: '/static/icon/new.png', path: '/pages/product-request/product-request', color: '#20CB6B' },
        { name: '系统设置', icon: 'gear', iconPath: '/static/icon/set.png', path: '/pages/settings/settings', color: '#20CB6B' },
        { name: '发票抬头', icon: 'paperplane', iconPath: '/static/icon/invoice.png', path: '/pages/invoice/invoice', color: '#20CB6B' },
        { name: '关于我们', icon: 'information', iconPath: '/static/icon/About.png', path: '/pages/about-us/about-us', color: '#20CB6B' }
      ]
    };
  },
  computed: {
    // 显示名称：有name显示name，否则显示用户XXXX
    displayName() {
      if (this.userInfo.name && this.userInfo.name.trim()) {
        return this.userInfo.name;
      }
      if (this.userInfo.user_code) {
        return `用户${this.userInfo.user_code}`;
      }
      return '未设置';
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
    // 获取设备信息
    const info = uni.getSystemInfoSync();
    // 设置状态栏高度
    this.statusBarHeight = info.statusBarHeight || 20;

    this.checkLoginStatus();
    this.loadUserInfo();
    this.loadOrderCounts();
    this.loadCarousels();
  },
  onShow() {
    this.checkLoginStatus();
    this.updateUserInfo();
    this.loadDefaultAddress();
    this.loadOrderCounts();
    this.loadCouponCount();
    this.loadUserBalance();
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
          uni.setStorageSync('miniUserInfo', res.data);
          if (res.data.unique_id) {
            uni.setStorageSync('miniUserUniqueId', res.data.unique_id);
          }
          this.userInfo = res.data;
          this.isLoggedIn = true;
        }
      } catch (error) {
        console.error('获取用户信息失败:', error);
      }
    },

    // 加载用户信息
    loadUserInfo() {
      const storedUserInfo = uni.getStorageSync('miniUserInfo');
      if (storedUserInfo) {
        this.userInfo = storedUserInfo;
        // 从用户信息中获取余额和积分
        this.userBalance = parseFloat(storedUserInfo.balance || 0);
        this.userPoints = parseInt(storedUserInfo.points || 0);
      }
    },

    // 加载用户余额
    async loadUserBalance() {
      if (!this.isLoggedIn) {
        this.userBalance = 0;
        this.userPoints = 0;
        return;
      }
      // 余额和积分从用户信息中获取，已在updateUserInfo中更新
      const storedUserInfo = uni.getStorageSync('miniUserInfo');
      if (storedUserInfo) {
        this.userBalance = parseFloat(storedUserInfo.balance || 0);
        this.userPoints = parseInt(storedUserInfo.points || 0);
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

    // 加载订单数量
    async loadOrderCounts() {
      if (!this.isLoggedIn) {
        this.orderCounts = {
          pending_delivery: 0,
          delivering: 0,
          delivered: 0
        };
        return;
      }

      try {
        const token = uni.getStorageSync('miniUserToken');
        if (!token) {
          return;
        }

        const statuses = ['pending_delivery', 'delivering', 'delivered'];
        const counts = {};

        for (const status of statuses) {
          try {
            const res = await getUserOrders(token, { pageNum: 1, pageSize: 1, status });
            if (res && res.code === 200 && res.data) {
              counts[status] = res.data.total || 0;
            } else {
              counts[status] = 0;
            }
          } catch (error) {
            console.error(`获取${status}订单数量失败:`, error);
            counts[status] = 0;
          }
        }

        this.orderCounts = counts;
      } catch (error) {
        console.error('加载订单数量失败:', error);
        this.orderCounts = {
          pending_delivery: 0,
          delivering: 0,
          delivered: 0
        };
      }
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
          this.couponCount = res.data.filter(coupon => coupon.status === 'unused').length;
        } else {
          this.couponCount = 0;
        }
      } catch (error) {
        console.error('加载优惠券数量失败:', error);
        this.couponCount = 0;
      }
    },

    // 加载轮播图
    async loadCarousels() {
      try {
        const res = await getCarousels();
        if (res && res.code === 200 && Array.isArray(res.data)) {
          this.carousels = res.data.filter(item => item.type === 'my' || !item.type);
        }
      } catch (error) {
        console.error('加载轮播图失败:', error);
        this.carousels = [];
      }
    },

    // 显示登录弹框
    goToLogin() {
      this.showLoginModal = true;
    },

    // 处理登录成功
    async handleLoginSuccess({ user, token, uniqueId }) {
      // 更新用户信息
      if (user) {
        this.userInfo = user;
      }
      this.isLoggedIn = true;

      // 刷新用户信息
      await this.updateUserInfo();
      await this.loadOrderCounts();
      await this.loadCouponCount();
      await this.loadUserBalance();
    },

    // 处理登录弹框关闭
    handleLoginModalClose() {
      this.showLoginModal = false;
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
      const statusParam = status ? `?status=${status}` : '';
      uni.navigateTo({
        url: `/pages/order/list${statusParam}`
      });
    },

    // 跳转到客服页面
    goToCustomerService() {
      if (!this.isLoggedIn) {
        this.goToLogin();
        return;
      }
      uni.navigateTo({
        url: '/pages/customer-service/customer-service'
      });
    },

    // 跳转到优惠券页面
    goToCoupons() {
      console.log(
        11
      );

      if (!this.isLoggedIn) {
        this.goToLogin();
        return;
      }
      uni.navigateTo({
        url: '/pages/coupons/coupons'
      });
    },

    // 处理轮播图点击
    handleCarouselClick(item) {
      if (!item.link || item.link.trim() === '') {
        return;
      }

      const link = item.link;

      // 处理外部链接（http:// 或 https://）
      if (link.startsWith('http://') || link.startsWith('https://')) {
        // #ifdef H5
        window.open(link, '_blank');
        // #endif
        // #ifndef H5
        uni.showToast({
          title: '外部链接暂不支持',
          icon: 'none'
        });
        // #endif
        return;
      }

      // 处理完整的小程序路径（以 /pages/ 开头）
      if (link.startsWith('/pages/')) {
        // 判断是否是 tabBar 页面
        const tabBarPages = ['/pages/index/index', '/pages/category/category', '/pages/cart/cart', '/pages/my/my'];
        if (tabBarPages.includes(link.split('?')[0])) {
          uni.switchTab({
            url: link.split('?')[0]
          });
        } else {
          uni.navigateTo({
            url: link
          });
        }
        return;
      }

      // 处理商品详情页：product/xxx 或 product?id=xxx
      if (link.startsWith('product/')) {
        const productId = link.split('/')[1];
        uni.navigateTo({
          url: '/pages/product/detail?id=' + productId
        });
        return;
      }

      // 处理分类页面：category/xxx 或 category?id=xxx
      if (link.startsWith('category/')) {
        const categoryId = link.split('/')[1];
        // 使用globalData传递分类ID
        getApp().globalData.targetCategoryId = categoryId;
        uni.switchTab({
          url: '/pages/category/category'
        });
        return;
      }

      // 处理富文本页面：rich-content/xxx 或 rich-content?id=xxx
      if (link.startsWith('rich-content/')) {
        const contentId = link.split('/')[1];
        uni.navigateTo({
          url: '/pages/rich-content/rich-content?id=' + contentId
        });
        return;
      }

      // 处理带查询参数的格式：page?key=value
      if (link.includes('?')) {
        const [page, params] = link.split('?');
        // 尝试匹配已知的页面路径
        if (page === 'product' || page === 'product/detail') {
          const idMatch = params.match(/id=(\d+)/);
          if (idMatch) {
            uni.navigateTo({
              url: '/pages/product/detail?id=' + idMatch[1]
            });
            return;
          }
        } else if (page === 'category') {
          const idMatch = params.match(/id=(\d+)/);
          if (idMatch) {
            getApp().globalData.targetCategoryId = idMatch[1];
            uni.switchTab({
              url: '/pages/category/category'
            });
            return;
          }
        } else if (page === 'rich-content') {
          const idMatch = params.match(/id=(\d+)/);
          if (idMatch) {
            uni.navigateTo({
              url: '/pages/rich-content/rich-content?id=' + idMatch[1]
            });
            return;
          }
        }
      }

      // 如果都不匹配，尝试作为完整路径处理
      if (link.startsWith('/')) {
        uni.navigateTo({
          url: link
        });
      } else {
        // 未知格式，提示用户
        console.warn('未知的跳转链接格式:', link);
        uni.showToast({
          title: '链接格式不正确',
          icon: 'none'
        });
      }
      if (item.link) {
        // 根据link类型跳转
        if (item.link.startsWith('/')) {
          uni.navigateTo({ url: item.link });
        } else if (item.link.startsWith('http')) {
          // 外部链接，可以打开webview
          uni.showToast({
            title: '外部链接',
            icon: 'none'
          });
        }
      }
    },

    // 处理功能点击
    handleFunctionClick(func) {
      // 关于我们不需要登录
      if (func.path === '/pages/about-us/about-us') {
        uni.navigateTo({
          url: func.path
        });
        return;
      }

      // 其他功能都需要登录
      if (!this.isLoggedIn) {
        this.goToLogin();
        return;
      }

      if (func.path) {
        uni.navigateTo({
          url: func.path
        });
      } else if (func.handler) {
        func.handler();
      }
    },

    // 处理图标加载错误
    handleIconError() {
      // 图标加载失败时静默处理，不显示图标
    }
  }
};
</script>

<style scoped>
.my-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #E8F8F0 0%, #E8F8F0 20%, #f5f5f5 40%, #f5f5f5 100%);
}

/* 个人信息区域 */
.user-section {
  /* background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%); */
  padding: 0 30rpx 90rpx 30rpx;
  border-radius: 0 0 0 0;
}

.user-header,
.login-prompt {
  display: flex;
  align-items: center;
  position: relative;
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
  background-color: rgba(32, 203, 107, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.user-name-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 12rpx;
  flex-wrap: wrap;
}

.user-name-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 12rpx;
  flex-wrap: wrap;
}

.user-name {
  font-size: 36rpx;
  font-weight: 600;
  color: #333;
}

.user-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  flex-wrap: wrap;
}

.user-type {
  display: flex;
  align-items: center;
  gap: 8rpx;
  font-size: 24rpx;
  padding: 0;
  border-radius: 20rpx;
  background-color: #E8F8F0;
  color: #20CB6B;
  height: 40rpx;
  line-height: 40rpx;
}

.user-type-icon {
  width: 30rpx;
  height: 30rpx;
  flex-shrink: 0;
}

.user-type-icon-placeholder {
  width: 30rpx;
  height: 30rpx;
  flex-shrink: 0;
}

.user-type-text {
  color: inherit;
}

.type-retail {
  background-color: #E8F8F0;
  color: #20CB6B;
}

.type-wholesale {
  background-color: #FFF3DA;
  color: #D4A574;
}

.user-code {
  font-size: 24rpx;
  color: #333;
}

.service-icon {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.service-icon-image {
  width: 48rpx;
  height: 48rpx;
}

.login-text {
  font-size: 36rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 8rpx;
}

.login-tip {
  font-size: 24rpx;
  color: #333;
}

/* 账户概览卡片 */
.account-card {
  background-color: #373D52;
  margin: -40rpx 20rpx 20rpx;
  border-radius: 20rpx;
  padding: 32rpx 28rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
  display: flex;
  justify-content: space-around;
}

.account-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  transition: opacity 0.2s;
}

.account-item:active {
  opacity: 0.7;
}

.account-label {
  font-size: 24rpx;
  color: #FFF3DA;
  margin-bottom: 8rpx;
}

.account-value {
  font-size: 32rpx;
  font-weight: 600;
  color: #FFF3DA;
}

/* 订单信息 */
.order-section {
  background-color: #fff;
  margin: 0 20rpx 20rpx;
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
  font-size: 28rpx;
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
  margin-bottom: 12rpx;
}

.tab-icon {
  width: 74rpx;
  height: 74rpx;
}

.badge {
  position: absolute;
  top: -8rpx;
  right: -8rpx;
  min-width: 40rpx;
  height: 40rpx;
  padding: 0 8rpx;
  background-color: #ff4d4f;
  color: #fff;
  font-size: 20rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2rpx solid #fff;
  box-sizing: border-box;
}

.tab-text {
  font-size: 26rpx;
  color: #555;
}

/* 轮播图 */
.carousel-section {
  margin: 20rpx 20rpx 20rpx;
  border-radius: 20rpx;
  overflow: hidden;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.carousel-swiper {
  width: 100%;
  height: 200rpx;
}

.carousel-image {
  width: 100%;
  height: 100%;
}

/* 我的功能 */
.functions-section {
  background-color: #fff;
  margin: 0 20rpx;
  border-radius: 20rpx;
  padding: 30rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.functions-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 32rpx 20rpx;
}

.function-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.function-icon-wrapper {
  position: relative;
  width: 88rpx;
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 20rpx;
  margin-bottom: 6rpx;
}

.function-icon {
  width: 56rpx;
  height: 56rpx;
}

.function-badge {
  position: absolute;
  top: -8rpx;
  right: -8rpx;
  min-width: 32rpx;
  height: 32rpx;
  padding: 0 8rpx;
  background-color: #ff4d4f;
  color: #fff;
  font-size: 20rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2rpx solid #fff;
  box-sizing: border-box;
}

.function-text {
  font-size: 24rpx;
  color: #555;
  text-align: center;
}
</style>
