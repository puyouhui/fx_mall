<template>
  <view class="login-modal-overlay" v-if="visible" @click="handleCancel">
    <view class="login-modal-content" @click.stop>
      <!-- 登录前状态 -->
      <view v-if="!loginSuccess">
        <view class="login-modal-header">
          <!-- <view class="login-modal-icon-wrapper">
            <uni-icons type="person-filled" size="48" color="#20CB6B"></uni-icons>
          </view> -->
          <text class="login-modal-title">登录提示</text>
        </view>
        <view class="login-modal-body">
          <text class="login-modal-text">{{ '需要登录后体验小程序完整功能' }}</text>
          <text class="login-modal-text">{{ '是否继续登录？' }}</text>
        </view>
        <view class="login-modal-footer">
          <view class="login-modal-btn cancel-btn" @click="handleCancel">
            <text class="login-modal-btn-text">取消</text>
          </view>
          <view class="login-modal-btn confirm-btn" @click="handleConfirm" :class="{ 'loading': isLoading }">
            <text class="login-modal-btn-text" v-if="!isLoading">确认</text>
            <text class="login-modal-btn-text" v-else>登录中...</text>
          </view>
        </view>
      </view>
      
      <!-- 登录成功状态 -->
      <view v-else class="login-success-content">
        <view class="success-header">
          <view class="success-icon-wrapper">
            <uni-icons type="checkmarkempty" size="60" color="rgba(255, 255, 255, 0.8)"></uni-icons>
          </view>
          <text class="success-title">登录成功</text>
        </view>
        <view class="success-body">
          <view class="user-code-section">
            <text class="user-code-label">您的用户编号</text>
            <view class="user-code-display" @click="copyUserCode">
              <text class="user-code-text">{{ userCode || '暂无' }}</text>
              <uni-icons type="copy" size="20" color="#20CB6B" class="copy-icon"></uni-icons>
            </view>
          </view>
          <view class="tip-section">
            <text class="tip-text">若您有指定销售员，请记住你的用户码，并告诉业务员</text>
          </view>
        </view>
        <view class="success-footer">
          <view class="copy-close-btn" @click="handleCopyAndClose">
            <text class="copy-close-btn-text">复制并关闭</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { miniLogin } from '../api/index.js';

export default {
  name: 'LoginModal',
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    message: {
      type: String,
      default: ''
    }
  },
  data() {
    return {
      isLoading: false,
      loginSuccess: false,
      userCode: ''
    };
  },
  watch: {
    visible(newVal) {
      // 当弹框关闭时，重置状态
      if (!newVal) {
        this.loginSuccess = false;
        this.userCode = '';
        this.isLoading = false;
      }
    }
  },
  methods: {
    // 执行小程序登录（和规格选择器一样的登录逻辑）
    async performMiniLogin() {
      uni.showLoading({
        title: '登录中...',
        mask: true
      });
      this.isLoading = true;
      
      try {
        const loginRes = await new Promise((resolve, reject) => {
          uni.login({
            provider: 'weixin',
            success: resolve,
            fail: reject
          });
        });

        if (!loginRes || !loginRes.code) {
          throw new Error('未获取到登录凭证');
        }

        // 获取本地存储的分享者ID
        const shareReferrerId = uni.getStorageSync('shareReferrerId');
        let referrerId = null;
        if (shareReferrerId) {
          const id = parseInt(shareReferrerId);
          if (!isNaN(id) && id > 0) {
            referrerId = id;
          }
        }

        const resp = await miniLogin(loginRes.code, referrerId);
        
        // 登录成功后，清除分享者ID（只绑定一次）
        if (referrerId) {
          uni.removeStorageSync('shareReferrerId');
        }
        const data = resp?.data || {};
        const user = data.user || {};
        const token = data.token || '';
        const uniqueId = user.unique_id || user.uniqueId;

        if (!uniqueId) {
          throw new Error('未返回用户唯一ID');
        }

        // 保存用户信息
        if (user) {
          uni.setStorageSync('miniUserInfo', user);
          if (uniqueId) {
            uni.setStorageSync('miniUserUniqueId', uniqueId);
          }
        }

        if (token) {
          uni.setStorageSync('miniUserToken', token);
        }

        // 保存用户编号
        this.userCode = user.user_code || user.userCode || '';
        
        // 检查用户是否已绑定业务员
        const hasSalesEmployee = (user.sales_code && user.sales_code.trim() !== '') || 
                                 (user.sales_employee_id && user.sales_employee_id > 0) ||
                                 (user.salesCode && user.salesCode.trim() !== '') ||
                                 (user.salesEmployeeId && user.salesEmployeeId > 0);
        
        // 如果已绑定业务员，直接关闭弹框
        if (hasSalesEmployee) {
          // 通知父组件登录成功
          this.$emit('login-success', {
            user,
            token,
            uniqueId
          });
          
          // 延迟关闭，让用户看到登录成功的提示
          setTimeout(() => {
            this.$emit('update:visible', false);
            this.$emit('close');
            // 重置状态
            this.loginSuccess = false;
            this.userCode = '';
          }, 800);
        } else {
          // 未绑定业务员，显示用户码页面
          this.loginSuccess = true;
          
          // 通知父组件登录成功
          this.$emit('login-success', {
            user,
            token,
            uniqueId
          });
        }
      } catch (error) {
        console.error('登录失败:', error);
        uni.showToast({
          title: error?.message || '登录失败，请稍后重试',
          icon: 'none'
        });
        this.$emit('login-error', error);
      } finally {
        this.isLoading = false;
        uni.hideLoading();
      }
    },
    
    // 确认登录
    handleConfirm() {
      if (this.isLoading) return;
      this.performMiniLogin();
    },
    
    // 取消登录
    handleCancel() {
      if (this.isLoading) return;
      if (this.loginSuccess) return; // 登录成功后不允许点击遮罩关闭
      this.$emit('update:visible', false);
      this.$emit('cancel');
    },
    
    // 复制用户编号
    copyUserCode() {
      return new Promise((resolve, reject) => {
        if (!this.userCode) {
          uni.showToast({
            title: '用户编号为空',
            icon: 'none'
          });
          reject(new Error('用户编号为空'));
          return;
        }
        
        uni.setClipboardData({
          data: this.userCode,
          success: () => {
            uni.showToast({
              title: '已复制到剪贴板',
              icon: 'success'
            });
            resolve();
          },
          fail: () => {
            uni.showToast({
              title: '复制失败',
              icon: 'none'
            });
            reject(new Error('复制失败'));
          }
        });
      });
    },
    
    // 复制并关闭
    async handleCopyAndClose() {
      try {
        if (this.userCode) {
          await this.copyUserCode();
          // 等待一下让用户看到复制成功的提示
          await new Promise(resolve => setTimeout(resolve, 300));
        }
      } catch (error) {
        console.error('复制失败:', error);
      } finally {
        // 关闭弹框
        this.$emit('update:visible', false);
        this.$emit('close');
        // 重置状态
        this.loginSuccess = false;
        this.userCode = '';
      }
    }
  }
};
</script>

<style scoped>
/* 登录弹框遮罩层 */
.login-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

/* 弹框内容 */
.login-modal-content {
  width: 640rpx;
  background-color: #fff;
  border-radius: 24rpx;
  overflow: hidden;
  box-shadow: 0 8rpx 32rpx rgba(0, 0, 0, 0.12);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from {
    transform: translateY(50rpx);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

/* ========== 登录前状态样式 ========== */

/* 弹框头部 */
.login-modal-header {
  padding: 50rpx 30rpx 0 30rpx;
  text-align: center;
  background: linear-gradient(180deg, #E8F8F0 0%, #fff 100%);
}

.login-modal-icon-wrapper {
  width: 120rpx;
  height: 120rpx;
  margin: 0 auto 24rpx;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 16rpx rgba(32, 203, 107, 0.3);
}

.login-modal-title {
  font-size: 40rpx;
  font-weight: 600;
  color: #333;
  display: block;
}

/* 弹框主体 */
.login-modal-body {
  padding: 40rpx 50rpx;
  text-align: center;
}

.login-modal-text {
  font-size: 30rpx;
  color: #666;
  line-height: 1.8;
  display: block;
}

/* 弹框底部按钮 */
.login-modal-footer {
  display: flex;
  border-top: 1rpx solid #f0f0f0;
}

.login-modal-btn {
  flex: 1;
  height: 110rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  transition: all 0.2s;
}

.login-modal-btn:active {
  opacity: 0.7;
}

.login-modal-btn:first-child {
  border-right: 1rpx solid #f0f0f0;
}

.cancel-btn {
  background-color: #fff;
}

.confirm-btn {
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  position: relative;
  overflow: hidden;
}

.confirm-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: left 0.5s;
}

.confirm-btn:active::before {
  left: 100%;
}

.confirm-btn.loading {
  opacity: 0.8;
}

.login-modal-btn-text {
  font-size: 32rpx;
  color: #333;
  font-weight: 500;
}

.confirm-btn .login-modal-btn-text {
  color: #fff;
  font-weight: 600;
}

/* ========== 登录成功状态样式 ========== */

.login-success-content {
  display: flex;
  flex-direction: column;
}

/* 成功头部 */
.success-header {
  padding: 60rpx 30rpx 0 30rpx;
  text-align: center;
  background: linear-gradient(180deg, #E8F8F0 0%, #fff 100%);
}

.success-icon-wrapper {
  width: 140rpx;
  height: 140rpx;
  margin: 0 auto 30rpx;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6rpx 20rpx rgba(32, 203, 107, 0.4);
  animation: scaleIn 0.4s ease;
}

@keyframes scaleIn {
  from {
    transform: scale(0);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

.success-title {
  font-size: 44rpx;
  font-weight: 600;
  color: #20CB6B;
  display: block;
}

/* 成功主体 */
.success-body {
  padding: 50rpx 40rpx;
}

.user-code-section {
  margin-bottom: 40rpx;
}

.user-code-label {
  font-size: 28rpx;
  color: #999;
  display: block;
  text-align: center;
  margin-bottom: 24rpx;
}

.user-code-display {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 40rpx;
  background: linear-gradient(135deg, #E8F8F0 0%, #F0FBF5 100%);
  border: 2rpx solid #20CB6B;
  border-radius: 16rpx;
  transition: all 0.3s;
  cursor: pointer;
}

.user-code-display:active {
  background: linear-gradient(135deg, #D8F5E8 0%, #E8F8F0 100%);
  transform: scale(0.98);
  box-shadow: 0 4rpx 12rpx rgba(32, 203, 107, 0.2);
}

.user-code-text {
  font-size: 64rpx;
  font-weight: 700;
  color: #20CB6B;
  letter-spacing: 4rpx;
  font-family: 'Courier New', monospace;
  flex: 1;
  text-align: center;
  line-height: 1.4;
}

.copy-icon {
  flex-shrink: 0;
  opacity: 0.7;
  transition: opacity 0.3s;
}

.user-code-display:active .copy-icon {
  opacity: 1;
}

.tip-section {
  padding: 20rpx 0;
  text-align: center;
}

.tip-text {
  font-size: 24rpx;
  color: #999;
  line-height: 1.6;
  display: block;
}

/* 成功底部按钮 */
.success-footer {
  padding: 0 40rpx 40rpx;
}

.copy-close-btn {
  width: 100%;
  height: 100rpx;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  border-radius: 50rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 16rpx rgba(32, 203, 107, 0.3);
  transition: all 0.3s;
  position: relative;
  overflow: hidden;
}

.copy-close-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.3), transparent);
  transition: left 0.5s;
}

.copy-close-btn:active {
  transform: scale(0.98);
  box-shadow: 0 2rpx 8rpx rgba(32, 203, 107, 0.2);
}

.copy-close-btn:active::before {
  left: 100%;
}

.copy-close-btn-text {
  font-size: 34rpx;
  font-weight: 600;
  color: #fff;
  letter-spacing: 2rpx;
}
</style>
