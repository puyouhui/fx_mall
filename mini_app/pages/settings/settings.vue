<template>
  <view class="settings-page">
    <!-- 自定义导航栏 -->
    <view class="custom-header">
      <view class="navbar-fixed" style="background-color: #E8F8F0;">
        <!-- 状态栏撑起高度 -->
        <view :style="{ height: statusBarHeight + 'px' }"></view>
        <!-- 导航栏内容区域 -->
        <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
          <view class="navbar-left" @click="goBack">
            <uni-icons type="left" size="20" color="#333"></uni-icons>
          </view>
          <view class="navbar-title">
            <text class="navbar-title-text">设置</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 设置内容 -->
    <view class="settings-content">
      <!-- 个人资料（已登录时显示） -->
      <view class="settings-group" v-if="isLoggedIn">
        <view class="settings-item" @click="goToProfile">
          <view class="item-left">
            <view class="item-icon profile-icon">
              <uni-icons type="person" size="22" color="#20CB6B"></uni-icons>
            </view>
            <text class="item-text">个人资料</text>
          </view>
          <uni-icons type="right" size="16" color="#ddd"></uni-icons>
        </view>
      </view>

      <!-- 关于我们 -->
      <view class="settings-group">
        <view class="settings-item" @click="goToAbout">
          <view class="item-left">
            <view class="item-icon about-icon">
              <uni-icons type="info" size="22" color="#20CB6B"></uni-icons>
            </view>
            <text class="item-text">关于我们</text>
          </view>
          <uni-icons type="right" size="16" color="#ddd"></uni-icons>
        </view>
      </view>

      <!-- 其他设置项（预留） -->
      <view class="settings-group">
        <view class="settings-item" @click="handlePrivacy">
          <view class="item-left">
            <view class="item-icon privacy-icon">
              <uni-icons type="locked" size="22" color="#20CB6B"></uni-icons>
            </view>
            <text class="item-text">隐私政策</text>
          </view>
          <uni-icons type="right" size="16" color="#ddd"></uni-icons>
        </view>
        <view class="settings-item" @click="handleTerms">
          <view class="item-left">
            <view class="item-icon terms-icon">
              <uni-icons type="paperplane" size="22" color="#20CB6B"></uni-icons>
            </view>
            <text class="item-text">用户协议</text>
          </view>
          <uni-icons type="right" size="16" color="#ddd"></uni-icons>
        </view>
      </view>

      <!-- 退出登录（已登录时显示） -->
      <view class="logout-section" v-if="isLoggedIn">
        <view class="logout-btn" @click="handleLogout">
          <text class="logout-text">退出登录</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
export default {
  data() {
    return {
      statusBarHeight: 20, // 状态栏高度（默认值）
      navBarHeight: 45, // 导航栏高度（默认值）
      isLoggedIn: false
    };
  },
  onLoad() {
    // 获取设备信息
    const info = uni.getSystemInfoSync();
    // 设置状态栏高度
    this.statusBarHeight = info.statusBarHeight;
    
    // 获取胶囊按钮信息并计算导航栏高度
    this.getMenuButtonInfo();
    
    // 检查登录状态
    this.checkLoginStatus();
  },
  methods: {
    // 获取胶囊按钮信息并计算导航栏高度
    getMenuButtonInfo() {
      try {
        // #ifndef H5 || APP-PLUS || MP-ALIPAY
        // 获取胶囊的位置信息
        const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
        // 按照参考文章的公式计算导航栏高度：
        // (胶囊底部高度 - 状态栏的高度) + (胶囊顶部高度 - 状态栏内的高度) = 导航栏的高度
        this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight);
        // #endif
      } catch (error) {
        console.error('获取胶囊按钮信息失败:', error);
      }
    },
    
    // 检查登录状态
    checkLoginStatus() {
      const token = uni.getStorageSync('miniUserToken');
      this.isLoggedIn = !!token;
    },
    
    // 返回上一页
    goBack() {
      uni.navigateBack();
    },
    
    // 跳转到个人资料
    goToProfile() {
      const token = uni.getStorageSync('miniUserToken');
      if (!token) {
        uni.showToast({
          title: '请先登录',
          icon: 'none'
        });
        return;
      }
      uni.navigateTo({
        url: '/pages/profile/profile'
      });
    },
    
    // 跳转到关于我们
    goToAbout() {
      uni.showModal({
        title: '关于我们',
        content: '商品选购小程序\n\n版本：1.0.0\n\n我们致力于为您提供优质的商品选购服务，让您的采购更加便捷高效。',
        showCancel: false,
        confirmText: '知道了',
        confirmColor: '#20CB6B'
      });
      // TODO: 可以跳转到专门的关于我们页面
    },
    
    // 隐私政策
    handlePrivacy() {
      uni.showModal({
        title: '隐私政策',
        content: '我们非常重视您的隐私保护。我们会严格保护您的个人信息，不会向第三方泄露。',
        showCancel: false,
        confirmText: '知道了',
        confirmColor: '#20CB6B'
      });
      // TODO: 可以跳转到隐私政策页面
    },
    
    // 用户协议
    handleTerms() {
      uni.showModal({
        title: '用户协议',
        content: '欢迎使用我们的服务。使用本服务即表示您同意遵守相关条款和条件。',
        showCancel: false,
        confirmText: '知道了',
        confirmColor: '#20CB6B'
      });
      // TODO: 可以跳转到用户协议页面
    },
    
    // 退出登录
    handleLogout() {
      uni.showModal({
        title: '提示',
        content: '确定要退出登录吗？',
        confirmText: '退出',
        cancelText: '取消',
        confirmColor: '#20CB6B',
        success: (res) => {
          if (res.confirm) {
            // 清除登录信息
            uni.removeStorageSync('miniUserToken');
            uni.removeStorageSync('miniUserInfo');
            uni.removeStorageSync('miniUserUniqueId');
            this.isLoggedIn = false;
            
            uni.showToast({
              title: '已退出登录',
              icon: 'success'
            });
            
            // 延迟返回上一页
            setTimeout(() => {
              uni.navigateBack();
            }, 1500);
          }
        }
      });
    }
  }
};
</script>

<style scoped>
.settings-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

/* 自定义导航栏 */
.custom-header {
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
  padding: 0 30rpx;
}

.navbar-left {
  width: 60rpx;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.navbar-title {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.navbar-title-text {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}

.navbar-right {
  width: 60rpx;
  height: 100%;
}

/* 设置内容 */
.settings-content {
  padding: 20rpx;
}

.settings-group {
  background-color: #fff;
  border-radius: 20rpx;
  margin-bottom: 20rpx;
  overflow: hidden;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.settings-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
  transition: background-color 0.2s;
}

.settings-item:last-child {
  border-bottom: none;
}

.settings-item:active {
  background-color: #f8f8f8;
}

.item-left {
  display: flex;
  align-items: center;
  flex: 1;
}

.item-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 24rpx;
}

.profile-icon {
  background-color: #E8F8F0;
}

.about-icon {
  background-color: #E8F8F0;
}

.privacy-icon {
  background-color: #E8F8F0;
}

.terms-icon {
  background-color: #E8F8F0;
}

.item-text {
  font-size: 28rpx;
  color: #333;
}

/* 退出登录 */
.logout-section {
  margin-top: 40rpx;
}

.logout-btn {
  width: 100%;
  height: 88rpx;
  background-color: #fff;
  border-radius: 44rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.logout-btn:active {
  opacity: 0.8;
  transform: scale(0.98);
}

.logout-text {
  font-size: 30rpx;
  color: #ff4d4f;
  font-weight: 500;
}
</style>

