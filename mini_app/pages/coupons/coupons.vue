<template>
  <view class="coupons-page">
    <!-- 顶部标签页 -->
    <view class="tabs">
      <view 
        class="tab" 
        :class="{ active: activeTab === 'unused' }"
        @click="switchTab('unused')"
      >
        未使用 ({{ unusedCount }})
      </view>
      <view 
        class="tab" 
        :class="{ active: activeTab === 'used' }"
        @click="switchTab('used')"
      >
        已使用 ({{ usedCount }})
      </view>
      <view 
        class="tab" 
        :class="{ active: activeTab === 'expired' }"
        @click="switchTab('expired')"
      >
        已过期 ({{ expiredCount }})
      </view>
    </view>

    <!-- 优惠券列表 -->
    <view class="coupons-list" v-if="loading">
      <view class="loading-wrapper">
        <text class="loading-text">加载中...</text>
      </view>
    </view>
    
    <view class="coupons-list" v-else-if="filteredCoupons.length > 0">
      <view 
        class="coupon-item" 
        v-for="item in filteredCoupons" 
        :key="item.id"
        :class="getCouponItemClass(item)"
      >
        <view class="coupon-content">
          <view class="coupon-left">
            <view class="coupon-name">{{ item.coupon?.name || '优惠券' }}</view>
            <view class="coupon-type">
              <text v-if="item.coupon?.type === 'delivery_fee'" class="type-tag delivery-tag">配送费券</text>
              <text v-else class="type-tag amount-tag">金额券</text>
            </view>
            <view class="coupon-value" v-if="item.coupon?.type === 'amount'">
              <text class="value-symbol">¥</text>
              <text class="value-amount">{{ (item.coupon?.discount_value || 0).toFixed(2) }}</text>
            </view>
            <view class="coupon-value" v-else>
              <text class="value-text">免配送费</text>
            </view>
            <view class="coupon-conditions" v-if="item.coupon">
              <text v-if="item.coupon.min_amount > 0">满¥{{ item.coupon.min_amount.toFixed(2) }}可用</text>
              <text v-else>无门槛</text>
            </view>
          </view>
          <view class="coupon-right">
            <view class="coupon-status">
              <text class="status-text">{{ getStatusText(item.status) }}</text>
            </view>
          </view>
        </view>
        
        <view class="coupon-footer">
          <view class="validity-info">
            <text class="validity-label">有效期：</text>
            <text class="validity-text">{{ getValidityText(item) }}</text>
          </view>
          <view class="coupon-time" v-if="item.created_at">
            <text class="time-text">发放时间：{{ formatDateTime(item.created_at) }}</text>
          </view>
        </view>
      </view>
    </view>
    
    <view class="empty-state" v-else>
      <view class="empty-icon">
        <uni-icons type="wallet" size="60" color="#ddd"></uni-icons>
      </view>
      <text class="empty-text">{{ getEmptyText() }}</text>
    </view>
  </view>
</template>

<script>
import { getUserCoupons } from '../../api/index.js';

export default {
  data() {
    return {
      coupons: [],
      activeTab: 'unused',
      loading: false,
      token: ''
    };
  },
  computed: {
    unusedCoupons() {
      return this.coupons.filter(item => item.status === 'unused');
    },
    usedCoupons() {
      return this.coupons.filter(item => item.status === 'used');
    },
    expiredCoupons() {
      return this.coupons.filter(item => item.status === 'expired');
    },
    unusedCount() {
      return this.unusedCoupons.length;
    },
    usedCount() {
      return this.usedCoupons.length;
    },
    expiredCount() {
      return this.expiredCoupons.length;
    },
    filteredCoupons() {
      switch (this.activeTab) {
        case 'unused':
          return this.unusedCoupons;
        case 'used':
          return this.usedCoupons;
        case 'expired':
          return this.expiredCoupons;
        default:
          return [];
      }
    }
  },
  onLoad() {
    this.token = uni.getStorageSync('miniUserToken');
    if (!this.token) {
      uni.showToast({
        title: '请先登录',
        icon: 'none'
      });
      setTimeout(() => {
        uni.navigateBack();
      }, 1500);
      return;
    }
    this.loadCoupons();
  },
  onShow() {
    // 每次显示页面时刷新数据
    if (this.token) {
      this.loadCoupons();
    }
  },
  methods: {
    async loadCoupons() {
      this.loading = true;
      try {
        const res = await getUserCoupons(this.token);
        if (res && res.code === 200 && Array.isArray(res.data)) {
          this.coupons = res.data;
        } else {
          this.coupons = [];
        }
      } catch (error) {
        console.error('获取优惠券列表失败:', error);
        uni.showToast({
          title: '获取优惠券失败',
          icon: 'none'
        });
        this.coupons = [];
      } finally {
        this.loading = false;
      }
    },
    
    switchTab(tab) {
      this.activeTab = tab;
    },
    
    getCouponItemClass(item) {
      return {
        'coupon-unused': item.status === 'unused',
        'coupon-used': item.status === 'used',
        'coupon-expired': item.status === 'expired'
      };
    },
    
    getStatusText(status) {
      const statusMap = {
        'unused': '未使用',
        'used': '已使用',
        'expired': '已过期'
      };
      return statusMap[status] || '未知';
    },
    
    getValidityText(item) {
      // 优先显示发放时设置的有效期
      if (item.expires_at) {
        const expiresDate = new Date(item.expires_at);
        const now = new Date();
        if (expiresDate < now && item.status === 'unused') {
          return '已过期';
        }
        return `至 ${this.formatDate(expiresDate)}`;
      }
      
      // 如果没有设置有效期，显示优惠券本身的有效期
      if (item.coupon) {
        if (item.coupon.valid_from && item.coupon.valid_to) {
          return `${this.formatDate(new Date(item.coupon.valid_from))} 至 ${this.formatDate(new Date(item.coupon.valid_to))}`;
        }
      }
      
      return '不限制';
    },
    
    formatDate(date) {
      if (!date) return '-';
      const d = new Date(date);
      const year = d.getFullYear();
      const month = String(d.getMonth() + 1).padStart(2, '0');
      const day = String(d.getDate()).padStart(2, '0');
      return `${year}-${month}-${day}`;
    },
    
    formatDateTime(dateStr) {
      if (!dateStr) return '-';
      const date = new Date(dateStr);
      if (isNaN(date.getTime())) return dateStr;
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      return `${year}-${month}-${day} ${hours}:${minutes}`;
    },
    
    getEmptyText() {
      switch (this.activeTab) {
        case 'unused':
          return '暂无未使用的优惠券';
        case 'used':
          return '暂无已使用的优惠券';
        case 'expired':
          return '暂无已过期的优惠券';
        default:
          return '暂无优惠券';
      }
    }
  }
};
</script>

<style scoped>
.coupons-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 40rpx;
}

/* 标签页 */
.tabs {
  display: flex;
  background-color: #fff;
  padding: 0 20rpx;
  border-bottom: 1rpx solid #eee;
}

.tab {
  flex: 1;
  text-align: center;
  padding: 30rpx 0;
  font-size: 28rpx;
  color: #666;
  position: relative;
  transition: color 0.3s;
}

.tab.active {
  color: #20CB6B;
  font-weight: 600;
}

.tab.active::after {
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

/* 优惠券列表 */
.coupons-list {
  padding: 20rpx;
}

.loading-wrapper {
  text-align: center;
  padding: 100rpx 0;
}

.loading-text {
  font-size: 28rpx;
  color: #999;
}

/* 优惠券卡片 */
.coupon-item {
  background-color: #fff;
  border-radius: 16rpx;
  margin-bottom: 20rpx;
  padding: 30rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
  position: relative;
  overflow: hidden;
}

.coupon-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 8rpx;
  background-color: #20CB6B;
}

.coupon-item.coupon-used::before {
  background-color: #999;
}

.coupon-item.coupon-expired::before {
  background-color: #ccc;
}

.coupon-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20rpx;
}

.coupon-left {
  flex: 1;
}

.coupon-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 12rpx;
}

.coupon-type {
  margin-bottom: 12rpx;
}

.type-tag {
  display: inline-block;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  font-weight: 500;
}

.delivery-tag {
  background-color: #E8F8F0;
  color: #20CB6B;
}

.amount-tag {
  background-color: #FFF4E6;
  color: #FF9500;
}

.coupon-value {
  margin-bottom: 12rpx;
}

.value-symbol {
  font-size: 28rpx;
  color: #ff4d4f;
  font-weight: 600;
}

.value-amount {
  font-size: 48rpx;
  color: #ff4d4f;
  font-weight: 700;
}

.value-text {
  font-size: 36rpx;
  color: #20CB6B;
  font-weight: 600;
}

.coupon-conditions {
  font-size: 24rpx;
  color: #999;
}

.coupon-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.coupon-status {
  padding: 8rpx 16rpx;
  border-radius: 20rpx;
  background-color: #f5f5f5;
}

.coupon-item.coupon-unused .coupon-status {
  background-color: #E8F8F0;
}

.coupon-item.coupon-unused .status-text {
  color: #20CB6B;
}

.coupon-item.coupon-used .status-text {
  color: #999;
}

.coupon-item.coupon-expired .status-text {
  color: #ccc;
}

.status-text {
  font-size: 24rpx;
  font-weight: 500;
}

.coupon-footer {
  border-top: 1rpx solid #f5f5f5;
  padding-top: 20rpx;
  margin-top: 20rpx;
}

.validity-info {
  display: flex;
  align-items: center;
  margin-bottom: 8rpx;
}

.validity-label {
  font-size: 24rpx;
  color: #999;
  margin-right: 8rpx;
}

.validity-text {
  font-size: 24rpx;
  color: #666;
}

.coupon-time {
  margin-top: 8rpx;
}

.time-text {
  font-size: 22rpx;
  color: #999;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 200rpx 40rpx;
}

.empty-icon {
  margin-bottom: 30rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999;
}
</style>

