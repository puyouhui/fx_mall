<template>
  <view class="points-logs-page">
    <!-- 导航栏 -->
    <view class="navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="nav-content">
        <view class="nav-left" @click="goBack">
          <uni-icons type="left" size="20" color="#333"></uni-icons>
        </view>
        <view class="nav-title">积分明细</view>
        <view class="nav-right"></view>
      </view>
    </view>

    <!-- 积分概览 -->
    <view class="points-overview">
      <view class="points-value">{{ userPoints }}</view>
      <view class="points-label">当前积分</view>
    </view>

    <!-- 积分明细列表 -->
    <view class="logs-container">
      <view v-if="loading" class="loading-wrapper">
        <uni-load-more status="loading"></uni-load-more>
      </view>
      <view v-else-if="logs.length === 0" class="empty-wrapper">
        <text class="empty-text">暂无积分明细</text>
      </view>
      <view v-else class="logs-list">
        <view class="log-item" v-for="(log, index) in logs" :key="log.id">
          <view class="log-left">
            <view class="log-type">{{ getTypeText(log.type) }}</view>
            <view class="log-desc">{{ log.description || '-' }}</view>
            <view class="log-time">{{ formatTime(log.created_at) }}</view>
          </view>
          <view class="log-right">
            <text class="log-points" :class="{ 'positive': log.points > 0, 'negative': log.points < 0 }">
              {{ log.points > 0 ? '+' : '' }}{{ log.points }}
            </text>
            <view class="log-balance">余额: {{ log.balance_after }}</view>
          </view>
        </view>
      </view>
    </view>

    <!-- 加载更多 -->
    <view v-if="hasMore && !loading" class="load-more-wrapper">
      <uni-load-more status="more" @clickLoadMore="loadMore"></uni-load-more>
    </view>
    <view v-if="!hasMore && logs.length > 0" class="load-more-wrapper">
      <uni-load-more status="noMore"></uni-load-more>
    </view>
  </view>
</template>

<script>
import { getPointsLogs } from '../../api/index.js';

export default {
  data() {
    return {
      statusBarHeight: 20,
      userPoints: 0,
      logs: [],
      pageNum: 1,
      pageSize: 10,
      total: 0,
      loading: false,
      hasMore: true
    };
  },
  onLoad() {
    // 获取设备信息
    const info = uni.getSystemInfoSync();
    this.statusBarHeight = info.statusBarHeight || 20;
    
    // 获取用户积分
    const userInfo = uni.getStorageSync('miniUserInfo');
    if (userInfo && userInfo.points !== undefined) {
      this.userPoints = parseInt(userInfo.points) || 0;
    }
    
    // 加载积分明细
    this.loadLogs();
  },
  onPullDownRefresh() {
    this.pageNum = 1;
    this.logs = [];
    this.hasMore = true;
    this.loadLogs().finally(() => {
      uni.stopPullDownRefresh();
    });
  },
  onReachBottom() {
    if (this.hasMore && !this.loading) {
      this.loadMore();
    }
  },
  methods: {
    goBack() {
      uni.navigateBack();
    },
    async loadLogs() {
      const token = uni.getStorageSync('miniUserToken');
      if (!token) {
        uni.showToast({
          title: '请先登录',
          icon: 'none'
        });
        setTimeout(() => {
          uni.navigateTo({
            url: '/pages/login/login'
          });
        }, 1500);
        return;
      }

      this.loading = true;
      try {
        const res = await getPointsLogs(token, {
          page_num: this.pageNum,
          page_size: this.pageSize
        });

        if (res && res.code === 200 && res.data) {
          const newLogs = res.data.list || [];
          this.logs = this.pageNum === 1 ? newLogs : [...this.logs, ...newLogs];
          this.total = res.data.total || 0;
          this.hasMore = this.logs.length < this.total;
        } else {
          uni.showToast({
            title: res.message || '获取积分明细失败',
            icon: 'none'
          });
        }
      } catch (error) {
        console.error('加载积分明细失败:', error);
        uni.showToast({
          title: '加载失败，请重试',
          icon: 'none'
        });
      } finally {
        this.loading = false;
      }
    },
    loadMore() {
      if (this.hasMore && !this.loading) {
        this.pageNum++;
        this.loadLogs();
      }
    },
    getTypeText(type) {
      const typeMap = {
        'order_reward': '订单奖励',
        'referral_reward': '推荐奖励',
        'points_discount': '积分抵扣',
        'admin_adjust': '管理员调整'
      };
      return typeMap[type] || type;
    },
    formatTime(timeStr) {
      if (!timeStr) return '-';
      const date = new Date(timeStr);
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hour = String(date.getHours()).padStart(2, '0');
      const minute = String(date.getMinutes()).padStart(2, '0');
      return `${year}-${month}-${day} ${hour}:${minute}`;
    }
  }
};
</script>

<style scoped>
.points-logs-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.navbar {
  background-color: #fff;
  border-bottom: 1px solid #eee;
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-content {
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 15px;
}

.nav-left {
  width: 40px;
  display: flex;
  align-items: center;
}

.nav-title {
  flex: 1;
  text-align: center;
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.nav-right {
  width: 40px;
}

.points-overview {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 30px 20px;
  text-align: center;
  color: #fff;
}

.points-value {
  font-size: 48px;
  font-weight: bold;
  margin-bottom: 10px;
}

.points-label {
  font-size: 14px;
  opacity: 0.9;
}

.logs-container {
  padding: 10px 15px;
}

.loading-wrapper,
.empty-wrapper {
  padding: 40px 0;
  text-align: center;
}

.empty-text {
  color: #999;
  font-size: 14px;
}

.logs-list {
  background-color: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.log-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  border-bottom: 1px solid #f0f0f0;
}

.log-item:last-child {
  border-bottom: none;
}

.log-left {
  flex: 1;
}

.log-type {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-bottom: 5px;
}

.log-desc {
  font-size: 13px;
  color: #666;
  margin-bottom: 5px;
}

.log-time {
  font-size: 12px;
  color: #999;
}

.log-right {
  text-align: right;
}

.log-points {
  font-size: 18px;
  font-weight: 600;
  display: block;
  margin-bottom: 5px;
}

.log-points.positive {
  color: #20CB6B;
}

.log-points.negative {
  color: #ff4757;
}

.log-balance {
  font-size: 12px;
  color: #999;
}

.load-more-wrapper {
  padding: 20px 0;
}
</style>

