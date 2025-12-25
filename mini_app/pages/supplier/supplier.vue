<template>
  <view class="supplier-page">
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
            <text class="navbar-title-text">供应商合作</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 顶部提示 -->
    <view class="tip-banner">
      <view class="tip-content">
        <uni-icons type="info" size="18" color="#20CB6B"></uni-icons>
        <text class="tip-text">目前平台正在招募以下类目供应商：生鲜食材、调料干货、日用消耗品、清洁用品、办公用品、包装材料等，欢迎优质供应商加入合作！</text>
      </view>
    </view>

    <!-- 表单内容 -->
    <view class="form-container">
      <!-- 公司名称 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">公司名称 <text class="required">*</text></text>
          <input 
            v-model="formData.company_name" 
            class="form-input" 
            placeholder="请输入公司名称" 
            maxlength="255"
          />
        </view>
      </view>

      <!-- 联系人 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">联系人 <text class="required">*</text></text>
          <input 
            v-model="formData.contact_name" 
            class="form-input" 
            placeholder="请输入联系人姓名" 
            maxlength="100"
          />
        </view>
      </view>

      <!-- 联系电话 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">联系电话 <text class="required">*</text></text>
          <input 
            v-model="formData.contact_phone" 
            class="form-input" 
            type="number"
            placeholder="请输入联系电话" 
            maxlength="20"
          />
        </view>
      </view>

      <!-- 邮箱 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">邮箱</text>
          <input 
            v-model="formData.email" 
            class="form-input" 
            type="email"
            placeholder="请输入邮箱（选填）" 
            maxlength="100"
          />
        </view>
      </view>

      <!-- 公司地址 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">公司地址</text>
          <input 
            v-model="formData.address" 
            class="form-input" 
            placeholder="请输入公司地址（选填）" 
            maxlength="500"
          />
        </view>
      </view>

      <!-- 主营类目 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">主营类目 <text class="required">*</text></text>
          <view class="category-group">
            <view 
              v-for="category in categories" 
              :key="category"
              class="category-item" 
              :class="{ 'active': formData.main_category === category }"
              @click="formData.main_category = category"
            >
              <text class="category-text">{{ category }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 公司简介 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">公司简介</text>
          <textarea 
            v-model="formData.company_intro" 
            class="form-textarea" 
            placeholder="请简要介绍您的公司（选填）" 
            maxlength="1000"
            :auto-height="true"
          ></textarea>
          <view class="char-count">{{ formData.company_intro.length }}/1000</view>
        </view>
      </view>

      <!-- 合作意向说明 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">合作意向说明</text>
          <textarea 
            v-model="formData.cooperation_intent" 
            class="form-textarea" 
            placeholder="请描述您的合作意向（选填）" 
            maxlength="1000"
            :auto-height="true"
          ></textarea>
          <view class="char-count">{{ formData.cooperation_intent.length }}/1000</view>
        </view>
      </view>
    </view>

    <!-- 底部提交按钮 -->
    <view class="bottom-button">
      <view class="submit-btn" @click="handleSubmit" :class="{ 'loading': submitting }">
        <text class="submit-btn-text">{{ submitting ? '提交中...' : '提交申请' }}</text>
      </view>
    </view>
  </view>
</template>

<script>
import { createSupplierApplication } from '../../api/index.js';

export default {
  data() {
    return {
      statusBarHeight: 20,
      navBarHeight: 45,
      token: '',
      submitting: false,
      categories: ['生鲜', '调料干货', '消耗品', '其他'],
      formData: {
        company_name: '',
        contact_name: '',
        contact_phone: '',
        email: '',
        address: '',
        main_category: '',
        company_intro: '',
        cooperation_intent: ''
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
    
    // 获取token（可选，未登录也可以提交）
    this.token = uni.getStorageSync('miniUserToken') || '';
  },
  methods: {
    goBack() {
      uni.navigateBack();
    },
    
    async handleSubmit() {
      // 验证必填字段
      if (!this.formData.company_name || this.formData.company_name.trim() === '') {
        uni.showToast({
          title: '请输入公司名称',
          icon: 'none'
        });
        return;
      }

      if (!this.formData.contact_name || this.formData.contact_name.trim() === '') {
        uni.showToast({
          title: '请输入联系人',
          icon: 'none'
        });
        return;
      }

      if (!this.formData.contact_phone || this.formData.contact_phone.trim() === '') {
        uni.showToast({
          title: '请输入联系电话',
          icon: 'none'
        });
        return;
      }

      // 验证手机号格式
      if (!/^1[3-9]\d{9}$/.test(this.formData.contact_phone.trim())) {
        uni.showToast({
          title: '请输入正确的手机号码',
          icon: 'none'
        });
        return;
      }

      if (!this.formData.main_category || this.formData.main_category.trim() === '') {
        uni.showToast({
          title: '请选择主营类目',
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
          company_name: this.formData.company_name.trim(),
          contact_name: this.formData.contact_name.trim(),
          contact_phone: this.formData.contact_phone.trim(),
          email: this.formData.email.trim(),
          address: this.formData.address.trim(),
          main_category: this.formData.main_category,
          company_intro: this.formData.company_intro.trim(),
          cooperation_intent: this.formData.cooperation_intent.trim()
        };

        const res = await createSupplierApplication(this.token, requestData);
        
        if (res.code === 200) {
          uni.showToast({
            title: '提交成功',
            icon: 'success'
          });
          
          // 清空表单
          this.formData = {
            company_name: '',
            contact_name: '',
            contact_phone: '',
            email: '',
            address: '',
            main_category: '',
            company_intro: '',
            cooperation_intent: ''
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
        console.error('提交供应商合作申请失败:', error);
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
.supplier-page {
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

/* 顶部提示 */
.tip-banner {
  background: linear-gradient(135deg, #E8F8F0 0%, #F0FDF4 100%);
  padding: 24rpx 30rpx;
  margin: 0 30rpx 24rpx 30rpx;
  border-radius: 16rpx;
  border-left: 6rpx solid #20CB6B;
}

.tip-content {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
}

.tip-text {
  flex: 1;
  font-size: 26rpx;
  color: #333;
  line-height: 1.8;
}

/* 表单容器 */
.form-container {
  padding: 0 30rpx;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
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

/* 类目选择 */
.category-group {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
}

.category-item {
  padding: 16rpx 32rpx;
  background-color: #f8f8f8;
  border-radius: 12rpx;
  border: 2rpx solid transparent;
  transition: all 0.3s;
}

.category-item.active {
  background-color: #E8F8F0;
  border-color: #20CB6B;
}

.category-text {
  font-size: 28rpx;
  color: #333;
}

.category-item.active .category-text {
  color: #20CB6B;
  font-weight: 600;
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

