<template>
  <view class="invoice-page">
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
            <text class="navbar-title-text">发票抬头</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 表单内容 -->
    <view class="form-container">
      <!-- 发票类型 -->
      <view class="form-section">
        <view class="section-title">发票类型</view>
        <view class="invoice-type-group">
          <view 
            class="type-item" 
            :class="{ 'active': formData.invoice_type === 'company' }"
            @click="formData.invoice_type = 'company'"
          >
            <text class="type-text">企业</text>
          </view>
          <view 
            class="type-item" 
            :class="{ 'active': formData.invoice_type === 'personal' }"
            @click="formData.invoice_type = 'personal'"
          >
            <text class="type-text">个人</text>
          </view>
        </view>
      </view>

      <!-- 发票抬头 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">发票抬头 <text class="required">*</text></text>
          <input 
            v-model="formData.title" 
            class="form-input" 
            placeholder="请输入发票抬头" 
            maxlength="200"
          />
        </view>
      </view>

      <!-- 企业发票额外字段 -->
      <view class="form-section" v-if="formData.invoice_type === 'company'">
        <view class="form-item">
          <text class="form-label">纳税人识别号 <text class="required">*</text></text>
          <input 
            v-model="formData.tax_number" 
            class="form-input" 
            placeholder="请输入纳税人识别号" 
            maxlength="50"
          />
        </view>
        <view class="form-item">
          <text class="form-label">公司地址</text>
          <input 
            v-model="formData.company_address" 
            class="form-input" 
            placeholder="请输入公司地址（选填）" 
            maxlength="255"
          />
        </view>
        <view class="form-item">
          <text class="form-label">公司电话</text>
          <input 
            v-model="formData.company_phone" 
            class="form-input" 
            type="number"
            placeholder="请输入公司电话（选填）" 
            maxlength="50"
          />
        </view>
        <view class="form-item">
          <text class="form-label">开户银行</text>
          <input 
            v-model="formData.bank_name" 
            class="form-input" 
            placeholder="请输入开户银行（选填）" 
            maxlength="100"
          />
        </view>
        <view class="form-item">
          <text class="form-label">银行账号</text>
          <input 
            v-model="formData.bank_account" 
            class="form-input" 
            type="number"
            placeholder="请输入银行账号（选填）" 
            maxlength="100"
          />
        </view>
      </view>

      <!-- 提示信息 -->
      <view class="tip-section">
        <view class="tip-content">
          <uni-icons type="info" size="16" color="#FF9500"></uni-icons>
          <text class="tip-text">如果需要开发票的订单，请在下单前联系客服</text>
        </view>
      </view>
    </view>

    <!-- 底部保存按钮 -->
    <view class="bottom-button">
      <view class="save-btn" @click="handleSave" :class="{ 'loading': saving }">
        <text class="save-btn-text">{{ saving ? '保存中...' : '保存' }}</text>
      </view>
    </view>
  </view>
</template>

<script>
import { getMiniUserInvoice, saveMiniUserInvoice } from '../../api/index.js';

export default {
  data() {
    return {
      statusBarHeight: 20,
      navBarHeight: 45,
      token: '',
      saving: false,
      formData: {
        invoice_type: 'company',
        title: '',
        tax_number: '',
        company_address: '',
        company_phone: '',
        bank_name: '',
        bank_account: ''
      }
    };
  },
  onLoad() {
    // 获取设备信息
    const info = uni.getSystemInfoSync();
    this.statusBarHeight = info.statusBarHeight;
    
    // 获取胶囊按钮信息并计算导航栏高度
    this.getMenuButtonInfo();
    
    // 检查登录状态
    const token = uni.getStorageSync('miniUserToken');
    if (!token) {
      uni.showToast({
        title: '请先登录',
        icon: 'none'
      });
      setTimeout(() => {
        uni.navigateBack();
      }, 1500);
      return;
    }
    this.token = token;
    
    // 加载发票抬头信息
    this.loadInvoice();
  },
  methods: {
    // 获取胶囊按钮信息并计算导航栏高度
    getMenuButtonInfo() {
      try {
        // #ifndef H5 || APP-PLUS || MP-ALIPAY
        const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
        this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight);
        // #endif
      } catch (error) {
        console.error('获取胶囊按钮信息失败:', error);
      }
    },
    
    // 返回上一页
    goBack() {
      uni.navigateBack();
    },
    
    // 加载发票抬头信息
    async loadInvoice() {
      try {
        uni.showLoading({ title: '加载中...' });
        const res = await getMiniUserInvoice(this.token);
        if (res && res.code === 200 && res.data) {
          this.formData = {
            invoice_type: res.data.invoice_type || 'personal',
            title: res.data.title || '',
            tax_number: res.data.tax_number || '',
            company_address: res.data.company_address || '',
            company_phone: res.data.company_phone || '',
            bank_name: res.data.bank_name || '',
            bank_account: res.data.bank_account || ''
          };
        }
      } catch (error) {
        console.error('获取发票抬头失败:', error);
      } finally {
        uni.hideLoading();
      }
    },
    
    // 保存发票抬头
    async handleSave() {
      if (this.saving) return;
      
      // 验证必填字段
      if (!this.formData.title || this.formData.title.trim() === '') {
        uni.showToast({
          title: '请输入发票抬头',
          icon: 'none'
        });
        return;
      }
      
      // 如果是企业类型，验证纳税人识别号
      if (this.formData.invoice_type === 'company' && (!this.formData.tax_number || this.formData.tax_number.trim() === '')) {
        uni.showToast({
          title: '请输入纳税人识别号',
          icon: 'none'
        });
        return;
      }
      
      this.saving = true;
      try {
        const res = await saveMiniUserInvoice(this.token, {
          invoice_type: this.formData.invoice_type,
          title: this.formData.title.trim(),
          tax_number: this.formData.tax_number.trim(),
          company_address: this.formData.company_address.trim(),
          company_phone: this.formData.company_phone.trim(),
          bank_name: this.formData.bank_name.trim(),
          bank_account: this.formData.bank_account.trim(),
          is_default: true
        });
        
        if (res && res.code === 200) {
          uni.showToast({
            title: '保存成功',
            icon: 'success'
          });
          setTimeout(() => {
            uni.navigateBack();
          }, 1500);
        } else {
          uni.showToast({
            title: res.message || '保存失败',
            icon: 'none'
          });
        }
      } catch (error) {
        console.error('保存发票抬头失败:', error);
        uni.showToast({
          title: '保存失败，请稍后再试',
          icon: 'none'
        });
      } finally {
        this.saving = false;
      }
    }
  }
};
</script>

<style scoped>
.invoice-page {
  min-height: 100vh;
  background-color: #F5F6FA;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
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
  background-color: #fff;
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
  color: #fff;
}

.navbar-right {
  width: 60rpx;
  height: 100%;
}

/* 表单容器 */
.form-container {
  padding: 0 24rpx 20rpx 24rpx;
  padding-bottom: calc(60rpx + env(safe-area-inset-bottom));
}

.form-section {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 32rpx 28rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 24rpx;
}

/* 发票类型选择 */
.invoice-type-group {
  display: flex;
  gap: 20rpx;
}

.type-item {
  flex: 1;
  height: 80rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2rpx solid #E0E0E0;
  border-radius: 12rpx;
  background-color: #fff;
  transition: all 0.3s;
}

.type-item.active {
  border-color: #20CB6B;
  background-color: #E8F8F0;
}

.type-text {
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
}

.type-item.active .type-text {
  color: #20CB6B;
  font-weight: 600;
}

/* 表单项 */
.form-item {
  margin-bottom: 32rpx;
}

.form-item:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
  margin-bottom: 16rpx;
}

.required {
  color: #FF4D4F;
}

.form-input {
  width: 100%;
  height: 88rpx;
  padding: 0 24rpx;
  background-color: #F5F6FA;
  border-radius: 12rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}

.form-input::placeholder {
  color: #909399;
}

/* 提示信息 */
.tip-section {
  margin-top: 20rpx;
  margin-bottom: 20rpx;
}

.tip-content {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
  padding: 24rpx;
  background-color: #FFF4E6;
  border-radius: 12rpx;
  border-left: 4rpx solid #FF9500;
}

.tip-text {
  flex: 1;
  font-size: 26rpx;
  color: #666;
  line-height: 1.6;
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
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}

.save-btn {
  width: 100%;
  height: 88rpx;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  border-radius: 44rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 16rpx rgba(32, 203, 107, 0.3);
  transition: all 0.3s;
}

.save-btn:active {
  opacity: 0.9;
  transform: scale(0.98);
}

.save-btn.loading {
  opacity: 0.7;
}

.save-btn-text {
  font-size: 32rpx;
  color: #fff;
  font-weight: 600;
}
</style>

