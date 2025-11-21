<template>
  <view class="identity-page">
    <view class="identity-header">
      <text class="identity-title">请选择您的身份类型</text>
      <text class="identity-subtitle">完成身份确认后即可继续加购商品</text>
    </view>

    <view class="identity-options">
      <view
        class="identity-card"
        :class="{ active: selectedType === 'retail' }"
        @click="selectType('retail')"
      >
        <view class="card-icon retail"></view>
        <view class="card-content">
          <text class="card-title">零售用户</text>
          <text class="card-desc">适用于门店、餐饮、便利店等实体零售终端</text>
        </view>
        <view class="card-check">
          <view class="check-dot" v-if="selectedType === 'retail'"></view>
        </view>
      </view>

      <view
        class="identity-card"
        :class="{ active: selectedType === 'wholesale' }"
        @click="selectType('wholesale')"
      >
        <view class="card-icon wholesale"></view>
        <view class="card-content">
          <text class="card-title">批发用户</text>
          <text class="card-desc">适用于批发商、经销商或连锁渠道客户</text>
        </view>
        <view class="card-check">
          <view class="check-dot" v-if="selectedType === 'wholesale'"></view>
        </view>
      </view>
    </view>

    <view class="tip-box">
      <text>如需更换身份类型，可联系销售人员协助处理。</text>
    </view>

    <button
      class="confirm-btn"
      :class="{ disabled: !selectedType || submitting }"
      :disabled="!selectedType || submitting"
      @click="handleConfirm"
    >
      {{ submitting ? '提交中...' : '确认并继续' }}
    </button>
  </view>
</template>

<script>
import { updateMiniUserType } from '../../api/index';

export default {
  data() {
    return {
      selectedType: '',
      submitting: false,
      userToken: '',
      userInfo: null
    };
  },
  onLoad() {
    const token = uni.getStorageSync('miniUserToken') || '';
    const info = uni.getStorageSync('miniUserInfo') || null;
    if (!token) {
      uni.showToast({
        title: '请先完成登录',
        icon: 'none'
      });
      setTimeout(() => {
        uni.navigateBack({ delta: 1 });
      }, 800);
      return;
    }

    this.userToken = token;
    this.userInfo = info;

    const currentType = (info?.user_type || info?.userType || '').toLowerCase();
    if (currentType && currentType !== 'unknown') {
      this.selectedType = currentType;
    }
  },
  methods: {
    selectType(type) {
      this.selectedType = type;
    },
    async handleConfirm() {
      if (!this.selectedType) {
        uni.showToast({
          title: '请选择身份类型',
          icon: 'none'
        });
        return;
      }

      if (!this.userToken) {
        uni.showToast({
          title: '登录状态已过期，请重新登录',
          icon: 'none'
        });
        return;
      }

      const currentType = (this.userInfo?.user_type || this.userInfo?.userType || '').toLowerCase();
      if (currentType === this.selectedType) {
        uni.showToast({
          title: '身份类型已设置',
          icon: 'none'
        });
        setTimeout(() => {
          uni.navigateBack({ delta: 1 });
        }, 600);
        return;
      }

      this.submitting = true;
      try {
        const res = await updateMiniUserType(this.selectedType, this.userToken);
        if (res && res.data) {
          const user = res.data;
          uni.setStorageSync('miniUserInfo', user);
          if (user.unique_id) {
            uni.setStorageSync('miniUserUniqueId', user.unique_id);
          }
          uni.showToast({
            title: '设置成功',
            icon: 'success'
          });
          setTimeout(() => {
            uni.navigateBack({ delta: 1 });
          }, 600);
        }
      } catch (error) {
        console.error('更新身份类型失败:', error);
        uni.showToast({
          title: error?.message || '提交失败，请稍后重试',
          icon: 'none'
        });
      } finally {
        this.submitting = false;
      }
    }
  }
};
</script>

<style>
.identity-page {
  min-height: 100vh;
  background-color: #f5f7f8;
  padding: 40rpx 32rpx 60rpx 32rpx;
  box-sizing: border-box;
}

.identity-header {
  margin-bottom: 40rpx;
}

.identity-title {
  display: block;
  font-size: 40rpx;
  font-weight: 600;
  color: #222;
}

.identity-subtitle {
  margin-top: 12rpx;
  display: block;
  font-size: 26rpx;
  color: #888;
}

.identity-options {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.identity-card {
  background-color: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
  display: flex;
  align-items: center;
  box-shadow: 0 10rpx 30rpx rgba(0, 0, 0, 0.04);
  border: 2rpx solid transparent;
}

.identity-card.active {
  border-color: #20cb6b;
  box-shadow: 0 12rpx 36rpx rgba(32, 203, 107, 0.18);
}

.card-icon {
  width: 80rpx;
  height: 80rpx;
  border-radius: 20rpx;
  margin-right: 28rpx;
  background: linear-gradient(135deg, #f0f4ff, #dfe9ff);
}

.card-icon.retail {
  background: linear-gradient(135deg, #f0fdf6, #c7f3da);
}

.card-icon.wholesale {
  background: linear-gradient(135deg, #f2f9ff, #cfe5ff);
}

.card-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.card-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #111;
}

.card-desc {
  font-size: 26rpx;
  color: #666;
}

.card-check {
  width: 40rpx;
  height: 40rpx;
  border-radius: 999rpx;
  border: 2rpx solid #e0e5ec;
  display: flex;
  align-items: center;
  justify-content: center;
}

.identity-card.active .card-check {
  border-color: #20cb6b;
  background-color: rgba(32, 203, 107, 0.12);
}

.check-dot {
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  background-color: #20cb6b;
}

.tip-box {
  margin: 40rpx 0 80rpx 0;
  background-color: #fff;
  border-radius: 20rpx;
  padding: 24rpx;
  color: #888;
  font-size: 24rpx;
  line-height: 1.5;
}

.confirm-btn {
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  border-radius: 999rpx;
  background: linear-gradient(120deg, #20cb6b, #16b35d);
  color: #fff;
  font-size: 32rpx;
  font-weight: 600;
  border: none;
}

.confirm-btn.disabled {
  background: #c7e9d3;
  color: #7cb895;
}
</style>

