<template>
  <view class="address-page">
    <!-- 自定义导航栏 -->
    <view class="custom-header">
      <view class="navbar-fixed" style="background-color: #FFFFFF;">
        <!-- 状态栏撑起高度 -->
        <view :style="{ height: statusBarHeight + 'px' }"></view>
        <!-- 导航栏内容区域 -->
        <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
          <view class="navbar-left" @click="goBack">
            <uni-icons type="left" size="20" color="#333"></uni-icons>
          </view>
          <view class="navbar-title">
            <text class="navbar-title-text">我的收货地址</text>
          </view>
          <view class="navbar-right">
            <uni-icons type="more-filled" size="20" color="#333" style="margin-right: 20rpx;"></uni-icons>
            <uni-icons type="gear" size="20" color="#333"></uni-icons>
          </view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 地址列表 -->
    <view class="address-list">
      <view 
        class="address-item" 
        v-for="(address, index) in addressList" 
        :key="index"
        :class="{ 'default-address': address.is_default }"
      >
        <view class="address-content" @click="handleSetDefault(address)">
          <view class="address-header">
            <text class="address-name" :class="{ 'default-name': address.is_default }">{{ address.name }}</text>
            <view class="default-badge" v-if="address.is_default">
              <text class="default-text">默认</text>
            </view>
          </view>
          <text class="address-detail">{{ address.address }}</text>
          <text class="address-contact">{{ address.contact }} {{ address.phone }}</text>
        </view>
        <view class="address-actions">
          <view class="address-action" @click.stop="handleEdit(address)">
            <uni-icons type="compose" size="20" color="#20CB6B"></uni-icons>
            <text class="action-text">编辑</text>
          </view>
          <view class="address-action delete-action" @click.stop="handleDelete(address, index)">
            <uni-icons type="trash" size="20" color="#ff4d4f"></uni-icons>
            <text class="action-text">删除</text>
          </view>
        </view>
      </view>
      
      <!-- 空状态 -->
      <view class="empty-state" v-if="addressList.length === 0">
        <text class="empty-text">暂无收货地址</text>
        <text class="empty-tip">点击下方按钮添加地址</text>
      </view>
    </view>

    <!-- 底部添加地址按钮 -->
    <view class="bottom-button">
      <view class="add-btn" @click="handleAddAddress">
        <text class="add-btn-text">添加地址</text>
      </view>
    </view>
  </view>
</template>

<script>
import { getMiniUserAddresses, deleteMiniUserAddress, setDefaultMiniUserAddress } from '../../api/index';

export default {
  data() {
    return {
      statusBarHeight: 20, // 状态栏高度（默认值）
      navBarHeight: 45, // 导航栏高度（默认值）
      addressList: [],
      userToken: ''
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
    
    this.userToken = token;
    this.loadAddresses();
  },
  onShow() {
    // 页面显示时重新加载地址列表
    if (this.userToken) {
      this.loadAddresses();
    }
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
    
    // 返回上一页
    goBack() {
      uni.navigateBack();
    },
    
    // 加载地址列表
    async loadAddresses() {
      try {
        const res = await getMiniUserAddresses(this.userToken);
        if (res && res.code === 200 && res.data) {
          this.addressList = res.data || [];
        }
      } catch (error) {
        console.error('加载地址列表失败:', error);
        uni.showToast({
          title: '加载地址列表失败',
          icon: 'none'
        });
      }
    },
    
    // 编辑地址
    handleEdit(address) {
      uni.navigateTo({
        url: `/pages/profile/form?address_id=${address.id}`
      });
    },
    
    // 添加地址
    handleAddAddress() {
      uni.navigateTo({
        url: '/pages/profile/form'
      });
    },
    
    // 删除地址
    async handleDelete(address, index) {
      uni.showModal({
        title: '提示',
        content: '确定要删除这个地址吗？',
        confirmText: '删除',
        cancelText: '取消',
        confirmColor: '#ff4d4f',
        success: async (res) => {
          if (res.confirm) {
            try {
              const result = await deleteMiniUserAddress(address.id, this.userToken);
              if (result && result.code === 200) {
                uni.showToast({
                  title: '删除成功',
                  icon: 'success'
                });
                // 重新加载地址列表
                this.loadAddresses();
              }
            } catch (error) {
              console.error('删除地址失败:', error);
              uni.showToast({
                title: error?.message || '删除失败',
                icon: 'none'
              });
            }
          }
        }
      });
    },
    
    // 设置默认地址
    async handleSetDefault(address) {
      if (address.is_default) {
        return; // 已经是默认地址，不需要操作
      }
      
      try {
        const res = await setDefaultMiniUserAddress(address.id, this.userToken);
        if (res && res.code === 200) {
          uni.showToast({
            title: '设置成功',
            icon: 'success'
          });
          // 重新加载地址列表
          this.loadAddresses();
        }
      } catch (error) {
        console.error('设置默认地址失败:', error);
        uni.showToast({
          title: error?.message || '设置失败',
          icon: 'none'
        });
      }
    }
  }
};
</script>

<style scoped>
.address-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 120rpx;
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
  background-color: #FFFFFF;
  border-bottom: 1rpx solid #f0f0f0;
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
  display: flex;
  align-items: center;
  justify-content: flex-end;
  width: 120rpx;
  height: 100%;
}

/* 地址列表 */
.address-list {
  padding: 20rpx;
}

.address-item {
  background-color: #FFFFFF;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.04);
}

.address-item.default-address {
  background-color: #F0F9F4;
}

.address-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.address-header {
  display: flex;
  align-items: center;
  margin-bottom: 16rpx;
}

.address-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  margin-right: 16rpx;
}

.address-name.default-name {
  color: #20CB6B;
}

.default-badge {
  background-color: #ff4d4f;
  border-radius: 4rpx;
  padding: 4rpx 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.default-text {
  font-size: 20rpx;
  color: #FFFFFF;
  font-weight: 500;
}

.address-detail {
  font-size: 28rpx;
  color: #666;
  line-height: 1.6;
  margin-bottom: 12rpx;
}

.address-contact {
  font-size: 26rpx;
  color: #999;
}

.address-actions {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  margin-left: 20rpx;
  flex-shrink: 0;
}

.address-action {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 60rpx;
  padding: 10rpx;
}

.action-text {
  font-size: 20rpx;
  color: #666;
  margin-top: 4rpx;
}

.delete-action .action-text {
  color: #ff4d4f;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 120rpx 20rpx;
}

.empty-text {
  font-size: 32rpx;
  color: #999;
  margin-bottom: 16rpx;
}

.empty-tip {
  font-size: 26rpx;
  color: #ccc;
}

/* 底部添加地址按钮 */
.bottom-button {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 20rpx 30rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background-color: #FFFFFF;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}

.add-btn {
  width: 100%;
  height: 88rpx;
  background-color: #20CB6B;
  border-radius: 44rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 16rpx rgba(32, 203, 107, 0.3);
}

.add-btn:active {
  opacity: 0.9;
  transform: scale(0.98);
}

.add-btn-text {
  font-size: 32rpx;
  color: #FFFFFF;
  font-weight: 600;
}
</style>

