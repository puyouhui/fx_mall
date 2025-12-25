<template>
  <view class="about-us-page">
    <!-- 自定义导航栏 -->
    <view class="custom-header">
      <view class="navbar-fixed" style="background-color: #20CB6B;">
        <!-- 状态栏撑起高度 -->
        <view :style="{ height: statusBarHeight + 'px' }"></view>
        <!-- 导航栏内容区域 -->
        <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
          <view class="navbar-left" @click="goBack">
            <uni-icons type="left" size="20" color="#fff"></uni-icons>
          </view>
          <view class="navbar-title">
            <text class="navbar-title-text">关于我们</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 内容区域 -->
    <view class="content-container">
      <!-- Logo/品牌区域 -->
      <view class="brand-section">
        <view class="logo-wrapper">
          <view class="logo-placeholder">
            <text class="logo-text">商城</text>
          </view>
        </view>
        <text class="app-name">商城小程序</text>
        <text class="app-version">Version 1.0.0</text>
      </view>

      <!-- 公司介绍 -->
      <view class="info-section">
        <view class="section-title">公司简介</view>
        <view class="section-content">
          <text class="content-text">
            我们是一家专业的B2B电商平台，致力于为批发和零售客户提供优质的商品和服务。
            平台涵盖生鲜食材、调料干货、日用消耗品、清洁用品、办公用品、包装材料等多个品类，
            通过高效的供应链管理和专业的配送服务，为客户创造价值。
          </text>
        </view>
      </view>

      <!-- 服务理念 -->
      <view class="info-section">
        <view class="section-title">服务理念</view>
        <view class="section-content">
          <view class="service-item">
            <view class="service-icon">✓</view>
            <text class="service-text">品质保证：严格筛选供应商，确保商品质量</text>
          </view>
          <view class="service-item">
            <view class="service-icon">✓</view>
            <text class="service-text">高效配送：专业的配送团队，快速送达</text>
          </view>
          <view class="service-item">
            <view class="service-icon">✓</view>
            <text class="service-text">贴心服务：7×24小时客服支持，随时为您服务</text>
          </view>
          <view class="service-item">
            <view class="service-icon">✓</view>
            <text class="service-text">诚信经营：透明价格，诚信交易</text>
          </view>
        </view>
      </view>

      <!-- 联系方式 -->
      <view class="info-section">
        <view class="section-title">联系我们</view>
        <view class="section-content">
          <view class="contact-item" @click="makePhoneCall('400-000-0000')">
            <view class="contact-icon-wrapper">
              <uni-icons type="phone" size="20" color="#20CB6B"></uni-icons>
            </view>
            <view class="contact-info">
              <text class="contact-label">客服热线</text>
              <text class="contact-value">400-000-0000</text>
            </view>
            <uni-icons type="right" size="16" color="#999"></uni-icons>
          </view>
          <view class="contact-item">
            <view class="contact-icon-wrapper">
              <uni-icons type="email" size="20" color="#20CB6B"></uni-icons>
            </view>
            <view class="contact-info">
              <text class="contact-label">邮箱地址</text>
              <text class="contact-value">service@example.com</text>
            </view>
          </view>
          <view class="contact-item">
            <view class="contact-icon-wrapper">
              <uni-icons type="location" size="20" color="#20CB6B"></uni-icons>
            </view>
            <view class="contact-info">
              <text class="contact-label">公司地址</text>
              <text class="contact-value">北京市朝阳区xxx街道xxx号</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 版权信息 -->
      <view class="copyright-section">
        <text class="copyright-text">© 2024 商城小程序 版权所有</text>
        <text class="copyright-text">All Rights Reserved</text>
      </view>
    </view>
  </view>
</template>

<script>
export default {
  data() {
    return {
      statusBarHeight: 20,
      navBarHeight: 45
    };
  },
  onLoad() {
    // 获取系统信息
    const systemInfo = uni.getSystemInfoSync();
    this.statusBarHeight = systemInfo.statusBarHeight || 20;
    
    // 获取导航栏高度
    const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
    if (menuButtonInfo) {
      this.navBarHeight = (menuButtonInfo.top - systemInfo.statusBarHeight) * 2 + menuButtonInfo.height;
    }
  },
  methods: {
    goBack() {
      uni.navigateBack();
    },
    
    makePhoneCall(phone) {
      uni.makePhoneCall({
        phoneNumber: phone,
        fail: (err) => {
          console.error('拨打电话失败:', err);
          uni.showToast({
            title: '拨打电话失败',
            icon: 'none'
          });
        }
      });
    }
  }
};
</script>

<style scoped>
.about-us-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #E8F8F0 0%, #E8F8F0 30%, #f5f5f5 50%, #f5f5f5 100%);
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
  background-color: #20CB6B;
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
  font-size: 36rpx;
  font-weight: 600;
  color: #fff;
}

.navbar-right {
  width: 60rpx;
}

/* 内容容器 */
.content-container {
  padding: 0 30rpx 60rpx 30rpx;
}

/* 品牌区域 */
.brand-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60rpx 0 40rpx 0;
}

.logo-wrapper {
  margin-bottom: 30rpx;
}

.logo-placeholder {
  width: 160rpx;
  height: 160rpx;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  border-radius: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8rpx 24rpx rgba(32, 203, 107, 0.3);
}

.logo-text {
  font-size: 48rpx;
  font-weight: 700;
  color: #fff;
}

.app-name {
  font-size: 40rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 12rpx;
}

.app-version {
  font-size: 24rpx;
  color: #999;
}

/* 信息区块 */
.info-section {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 40rpx 30rpx;
  margin-top: 30rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 24rpx;
  padding-bottom: 20rpx;
  border-bottom: 2rpx solid #f0f0f0;
}

.section-content {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.content-text {
  font-size: 28rpx;
  color: #666;
  line-height: 1.8;
  text-align: justify;
}

/* 服务项 */
.service-item {
  display: flex;
  align-items: flex-start;
  gap: 20rpx;
}

.service-icon {
  width: 40rpx;
  height: 40rpx;
  background-color: #E8F8F0;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 4rpx;
  font-size: 24rpx;
  color: #20CB6B;
  font-weight: 600;
}

.service-text {
  flex: 1;
  font-size: 28rpx;
  color: #666;
  line-height: 1.8;
}

/* 联系方式 */
.contact-item {
  display: flex;
  align-items: center;
  padding: 24rpx 0;
  border-bottom: 1rpx solid #f5f5f5;
}

.contact-item:last-child {
  border-bottom: none;
}

.contact-icon-wrapper {
  width: 64rpx;
  height: 64rpx;
  background-color: #E8F8F0;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 24rpx;
}

.contact-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.contact-label {
  font-size: 24rpx;
  color: #999;
}

.contact-value {
  font-size: 28rpx;
  color: #333;
}

/* 版权信息 */
.copyright-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60rpx 0 40rpx 0;
  gap: 12rpx;
}

.copyright-text {
  font-size: 24rpx;
  color: #999;
}
</style>

