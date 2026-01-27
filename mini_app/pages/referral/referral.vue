<template>
  <view class="referral-page">
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
            <text class="navbar-title-text">推荐朋友</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 内容区域 -->
    <view class="content-container">
      <!-- 统计数据卡片 -->
      <view class="stats-card">
        <view class="stats-item">
          <text class="stats-value">{{ stats.totalReferrals || 0 }}</text>
          <text class="stats-label">累计推荐</text>
        </view>
        <view class="stats-divider"></view>
        <view class="stats-item">
          <text class="stats-value">{{ stats.orderedReferrals || 0 }}</text>
          <text class="stats-label">已下单好友</text>
        </view>
        <view class="stats-divider"></view>
        <view class="stats-item">
          <text class="stats-value">{{ stats.pendingReferrals || 0 }}</text>
          <text class="stats-label">待下单好友</text>
        </view>
      </view>

      <!-- 活动说明 -->
      <view class="activity-section">
        <view class="section-header">
          <view class="section-title-wrapper">
            <uni-icons type="info" size="20" color="#20CB6B"></uni-icons>
            <text class="section-title">活动说明</text>
          </view>
        </view>
        <view class="activity-content">
          <view class="activity-item" v-for="(item, index) in activityRules" :key="index">
            <view class="activity-dot"></view>
            <text class="activity-text">{{ item }}</text>
          </view>
        </view>
      </view>

      <!-- 用户列表 -->
      <view class="users-section">
        <view class="section-header">
          <text class="section-title">我拉取的用户</text>
          <text class="section-count">共{{ userList.length }}人</text>
        </view>

        <!-- 加载中 -->
        <view v-if="loading" class="loading-container">
          <uni-icons type="spinner-cycle" size="40" color="#20CB6B"></uni-icons>
          <text class="loading-text">加载中...</text>
        </view>

        <!-- 空状态 -->
        <view v-else-if="userList.length === 0" class="empty-container">
          <uni-icons type="person" size="60" color="#C0C4CC"></uni-icons>
          <text class="empty-text">暂无拉取的用户</text>
          <text class="empty-tip">分享小程序给好友，邀请他们注册下单吧！</text>
        </view>

        <!-- 用户列表 -->
        <view v-else class="user-list">
          <view 
            class="user-item" 
            v-for="(user, index) in userList" 
            :key="user.id || index"
          >
            <view class="user-avatar-wrapper">
              <image 
                v-if="user.avatar" 
                :src="user.avatar" 
                class="user-avatar" 
                mode="aspectFill"
              ></image>
              <view v-else class="user-avatar-placeholder">
                <uni-icons type="person-filled" size="24" color="#fff"></uni-icons>
              </view>
            </view>
            <view class="user-info">
              <view class="user-name-row">
                <text class="user-name">{{ user.name || user.nickname || '未设置昵称' }}</text>
                <view class="user-status" :class="user.has_ordered ? 'ordered' : 'pending'">
                  <text class="status-text">{{ user.has_ordered ? '已下单' : '待下单' }}</text>
                </view>
              </view>
              <view class="user-meta">
                <text class="user-meta-item" v-if="user.phone">手机：{{ user.phone }}</text>
                <text class="user-meta-item" v-if="user.registered_at">
                  注册时间：{{ formatDate(user.registered_at) }}
                </text>
                <text class="user-meta-item" v-if="user.first_order_at && user.has_ordered">
                  首次下单：{{ formatDate(user.first_order_at) }}
                </text>
              </view>
            </view>
          </view>
        </view>

        <!-- 加载更多 -->
        <view v-if="hasMore && !loading" class="load-more" @click="loadMore">
          <text class="load-more-text">加载更多</text>
        </view>

        <!-- 没有更多了 -->
        <view v-if="!hasMore && userList.length > 0" class="no-more">
          <text class="no-more-text">没有更多了</text>
        </view>
      </view>
    </view>
    
    <!-- 悬浮分享按钮 -->
    <button class="share-button" open-type="share" hover-class="none">
      <view class="share-button-content">
        <!-- <uni-icons type="share" size="24" color="#fff"></uni-icons> -->
        <text class="share-button-text">分享</text>
      </view>
    </button>
  </view>
</template>

<script>
import { getReferralUsers, getReferralActivityInfo, getReferralStats } from '../../api/referral.js';
import { getShareConfig, buildSharePath } from '../../utils/shareConfig.js';

export default {
  data() {
    return {
      statusBarHeight: 20,
      navBarHeight: 45,
      token: '',
      loading: false,
      userList: [],
      stats: {
        totalReferrals: 0,
        orderedReferrals: 0,
        pendingReferrals: 0
      },
      activityRules: [
        '活动时间：即日起至2026年6月30日',
        '邀请好友通过您的分享链接注册并完成首次下单',
        '好友完成首次下单后，您将获得5元现金奖励（线下发放）',
        '奖励形式三选一：5元现金（线下发放）、5元无门槛抵扣券、抽纸一提（3包装）',
        '奖励将在好友首次下单后7个工作日内发放，如选择实物奖励，会再您下次进货时赠送！',
        '活动最终解释权归平台所有'
      ],
      pageNum: 1,
      pageSize: 10,
      hasMore: true
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
    
    // 加载数据
    this.loadData();
  },
  onPullDownRefresh() {
    this.pageNum = 1;
    this.userList = [];
    this.hasMore = true;
    this.loadData().finally(() => {
      uni.stopPullDownRefresh();
    });
  },
  onReachBottom() {
    if (this.hasMore && !this.loading) {
      this.loadMore();
    }
  },
  // 分享小程序（推荐页面）
  onShareAppMessage(options) {
    // 使用 shareConfig 获取分享配置
    const shareConfig = getShareConfig('referral');
    
    // 构建分享路径，添加分享者ID，分享后进入首页
    const path = buildSharePath('/pages/index/index');
    
    return {
      title: shareConfig.title,
      path: path,
      imageUrl: shareConfig.imageUrl || ''
    };
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
    
    // 加载数据
    async loadData() {
      this.loading = true;
      try {
        // 并行加载统计数据和用户列表
        await Promise.all([
          this.loadStats(),
          this.loadUserList()
        ]);
      } catch (error) {
        console.error('加载数据失败:', error);
        uni.showToast({
          title: '加载失败，请稍后再试',
          icon: 'none'
        });
      } finally {
        this.loading = false;
      }
    },
    
    // 加载统计数据
    async loadStats() {
      try {
        const res = await getReferralStats(this.token);
        if (res && res.code === 200 && res.data) {
          this.stats = {
            totalReferrals: res.data.total_referrals || 0,
            orderedReferrals: res.data.ordered_referrals || 0,
            pendingReferrals: res.data.pending_referrals || 0
          };
        }
      } catch (error) {
        console.error('加载统计数据失败:', error);
        // 如果接口不存在，使用默认值
      }
    },
    
    // 加载活动说明
    async loadActivityInfo() {
      try {
        const res = await getReferralActivityInfo(this.token);
        if (res && res.code === 200 && res.data && res.data.rules) {
          this.activityRules = res.data.rules;
        }
      } catch (error) {
        console.error('加载活动说明失败:', error);
        // 使用默认活动说明
      }
    },
    
    // 加载用户列表
    async loadUserList() {
      try {
        const res = await getReferralUsers(this.token, {
          page_num: this.pageNum,
          page_size: this.pageSize
        });
        
        console.log('[loadUserList] API返回数据:', res);
        
        if (res && res.code === 200) {
          // 处理返回的数据结构
          let list = [];
          if (res.data) {
            if (Array.isArray(res.data)) {
              // 如果 data 直接是数组
              list = res.data;
            } else if (res.data.list && Array.isArray(res.data.list)) {
              // 如果 data 是对象，包含 list 字段
              list = res.data.list;
            } else if (res.data.referrals && Array.isArray(res.data.referrals)) {
              // 兼容其他可能的字段名
              list = res.data.referrals;
            }
          }
          
          console.log('[loadUserList] 解析后的用户列表:', list);
          console.log('[loadUserList] 用户列表长度:', list.length);
          
          if (this.pageNum === 1) {
            this.userList = list;
          } else {
            this.userList = [...this.userList, ...list];
          }
          
          // 判断是否还有更多数据
          const total = res.data?.total || res.data?.count || 0;
          this.hasMore = this.userList.length < total && list.length === this.pageSize;
          
          console.log('[loadUserList] 总数:', total);
          console.log('[loadUserList] 当前列表长度:', this.userList.length);
          console.log('[loadUserList] 是否还有更多:', this.hasMore);
        } else {
          console.error('[loadUserList] API返回错误:', res);
          uni.showToast({
            title: res?.message || '加载失败',
            icon: 'none'
          });
        }
      } catch (error) {
        console.error('[loadUserList] 加载用户列表失败:', error);
        // 如果接口不存在，使用模拟数据
        if (this.pageNum === 1) {
          this.userList = [];
        }
      }
    },
    
    // 加载更多
    async loadMore() {
      if (this.loading || !this.hasMore) return;
      
      this.pageNum++;
      this.loading = true;
      try {
        await this.loadUserList();
      } catch (error) {
        console.error('加载更多失败:', error);
        this.pageNum--; // 失败时回退页码
      } finally {
        this.loading = false;
      }
    },
    
    // 格式化日期
    formatDate(dateString) {
      if (!dateString) return '';
      const date = new Date(dateString);
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      return `${year}-${month}-${day} ${hours}:${minutes}`;
    },
    
  }
};
</script>

<style scoped>
.referral-page {
  min-height: 100vh;
  background-color: #F5F6FA;
  padding-bottom: 40rpx;
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

/* 内容容器 */
.content-container {
  padding: 0 24rpx;
}

/* 统计数据卡片 */
.stats-card {
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  border-radius: 20rpx;
  padding: 40rpx 30rpx;
  margin-bottom: 20rpx;
  display: flex;
  align-items: center;
  justify-content: space-around;
  box-shadow: 0 4rpx 16rpx rgba(32, 203, 107, 0.3);
}

.stats-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
}

.stats-value {
  font-size: 48rpx;
  font-weight: 700;
  color: #fff;
}

.stats-label {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.9);
}

.stats-divider {
  width: 2rpx;
  height: 60rpx;
  background-color: rgba(255, 255, 255, 0.3);
}

/* 活动说明 */
.activity-section {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 32rpx 28rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.section-title-wrapper {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
}

.section-count {
  font-size: 24rpx;
  color: #999;
}

.activity-content {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.activity-item {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;
}

.activity-dot {
  width: 12rpx;
  height: 12rpx;
  border-radius: 50%;
  background-color: #20CB6B;
  margin-top: 8rpx;
  flex-shrink: 0;
}

.activity-text {
  flex: 1;
  font-size: 28rpx;
  color: #666;
  line-height: 1.6;
}

/* 用户列表 */
.users-section {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 32rpx 28rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.user-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.user-item {
  display: flex;
  align-items: flex-start;
  gap: 20rpx;
  padding: 24rpx;
  background-color: #F5F6FA;
  border-radius: 16rpx;
}

.user-avatar-wrapper {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
}

.user-avatar {
  width: 100%;
  height: 100%;
}

.user-avatar-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.user-name-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.user-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  flex: 1;
}

.user-status {
  padding: 6rpx 16rpx;
  border-radius: 20rpx;
  flex-shrink: 0;
}

.user-status.ordered {
  background-color: #E6F7FF;
}

.user-status.pending {
  background-color: #FFF4E6;
}

.status-text {
  font-size: 22rpx;
  font-weight: 500;
}

.user-status.ordered .status-text {
  color: #1890FF;
}

.user-status.pending .status-text {
  color: #FF9500;
}

.user-meta {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.user-meta-item {
  font-size: 24rpx;
  color: #999;
}

/* 加载状态 */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80rpx 0;
  gap: 20rpx;
}

.loading-text {
  font-size: 28rpx;
  color: #999;
}

/* 空状态 */
.empty-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 120rpx 0;
  gap: 20rpx;
}

.empty-text {
  font-size: 30rpx;
  color: #666;
  font-weight: 500;
}

.empty-tip {
  font-size: 24rpx;
  color: #999;
  text-align: center;
  padding: 0 40rpx;
  line-height: 1.6;
}

/* 加载更多 */
.load-more {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx 0;
}

.load-more-text {
  font-size: 28rpx;
  color: #20CB6B;
}

.no-more {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx 0;
}

.no-more-text {
  font-size: 24rpx;
  color: #999;
}

/* 悬浮分享按钮 */
.share-button {
  position: fixed;
  bottom: 120rpx;
  right: 30rpx;
  z-index: 999;
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  box-shadow: 0 8rpx 24rpx rgba(32, 203, 107, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  touch-action: manipulation;
  border: none;
  padding: 0;
  margin: 0;
  line-height: 1;
  font-size: 28rpx;
}

.share-button::after {
  border: none;
}

.share-button:active {
  transform: scale(0.95);
  opacity: 0.9;
}

.share-button-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
}

.share-button-text {
  font-size: 22rpx;
  color: #fff;
  font-weight: 500;
}
</style>
