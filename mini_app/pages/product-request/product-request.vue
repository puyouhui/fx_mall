<template>
  <view class="product-request-page">
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
            <text class="navbar-title-text">新品需求</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 表单内容 -->
    <view class="form-container">
      <!-- 需求产品 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">需求产品 <text class="required">*</text></text>
          <input 
            v-model="formData.product_name" 
            class="form-input" 
            placeholder="请输入需求产品名称" 
            maxlength="255"
          />
        </view>
      </view>

      <!-- 品牌 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">品牌</text>
          <input 
            v-model="formData.brand" 
            class="form-input" 
            placeholder="请输入品牌（选填）" 
            maxlength="100"
          />
        </view>
      </view>

      <!-- 月消耗数量 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">月消耗数量</text>
          <input 
            v-model="formData.monthly_quantity" 
            class="form-input" 
            type="number"
            placeholder="请输入月消耗数量（选填）" 
          />
        </view>
      </view>

      <!-- 需求说明 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">需求说明</text>
          <textarea 
            v-model="formData.description" 
            class="form-textarea" 
            placeholder="请详细描述您的需求（选填）" 
            maxlength="1000"
            :auto-height="true"
          ></textarea>
          <view class="char-count">{{ formData.description.length }}/1000</view>
        </view>
      </view>
    </view>

    <!-- 底部提交按钮 -->
    <view class="bottom-button">
      <view class="submit-btn" @click="handleSubmit" :class="{ 'loading': submitting }">
        <text class="submit-btn-text">{{ submitting ? '提交中...' : '提交需求' }}</text>
      </view>
    </view>
  </view>
</template>

<script>
import { createProductRequest } from '../../api/index.js';

export default {
  data() {
    return {
      statusBarHeight: 20,
      navBarHeight: 45,
      token: '',
      submitting: false,
      formData: {
        product_name: '',
        brand: '',
        monthly_quantity: 0,
        description: ''
      }
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
    
    // 获取token
    this.token = uni.getStorageSync('miniUserToken');
    if (!this.token) {
      uni.showToast({
        title: '请先登录',
        icon: 'none'
      });
      setTimeout(() => {
        uni.navigateBack();
      }, 1500);
    }
  },
  methods: {
    goBack() {
      uni.navigateBack();
    },
    
    async handleSubmit() {
      // 验证必填字段
      if (!this.formData.product_name || this.formData.product_name.trim() === '') {
        uni.showToast({
          title: '请输入需求产品名称',
          icon: 'none'
        });
        return;
      }

      if (this.submitting) {
        return;
      }

      this.submitting = true;

      try {
        const requestData = {
          product_name: this.formData.product_name.trim(),
          brand: this.formData.brand.trim(),
          monthly_quantity: parseInt(this.formData.monthly_quantity) || 0,
          description: this.formData.description.trim()
        };

        const res = await createProductRequest(this.token, requestData);
        
        if (res.code === 200) {
          uni.showToast({
            title: '提交成功',
            icon: 'success'
          });
          
          // 清空表单
          this.formData = {
            product_name: '',
            brand: '',
            monthly_quantity: 0,
            description: ''
          };
          
          // 延迟返回
          setTimeout(() => {
            uni.navigateBack();
          }, 1500);
        } else {
          uni.showToast({
            title: res.message || '提交失败',
            icon: 'none'
          });
        }
      } catch (error) {
        console.error('提交新品需求失败:', error);
        uni.showToast({
          title: '提交失败，请稍后再试',
          icon: 'none'
        });
      } finally {
        this.submitting = false;
      }
    }
  }
};
</script>

<style scoped>
.product-request-page {
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

/* 表单容器 */
.form-container {
  padding: 0 30rpx;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
  margin-top: -24rpx;
}

.form-section {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 30rpx;
  margin-top: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.form-item {
  display: flex;
  flex-direction: column;
}

.form-label {
  font-size: 28rpx;
  font-weight: 500;
  color: #333;
  margin-bottom: 20rpx;
}

.required {
  color: #ff4757;
}

.form-input {
  width: 100%;
  height: 88rpx;
  background-color: #f8f8f8;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}

.form-textarea {
  width: 100%;
  min-height: 200rpx;
  background-color: #f8f8f8;
  border-radius: 12rpx;
  padding: 24rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
  line-height: 1.6;
}

.char-count {
  text-align: right;
  font-size: 24rpx;
  color: #999;
  margin-top: 12rpx;
}

/* 底部按钮 */
.bottom-button {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 20rpx 30rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background-color: #fff;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.08);
  z-index: 100;
}

.submit-btn {
  width: 100%;
  height: 88rpx;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  border-radius: 44rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 12rpx rgba(32, 203, 107, 0.3);
}

.submit-btn.loading {
  opacity: 0.7;
}

.submit-btn-text {
  font-size: 32rpx;
  font-weight: 600;
  color: #fff;
}
</style>

