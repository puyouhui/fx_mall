<template>
  <view class="favorites-page">
    <!-- 头部 -->
    <view class="page-header" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="header-content" :style="{ height: navBarHeight + 'px' }">
        <view class="back-btn" @click="goBack">
          <uni-icons type="left" size="24" color="#fff"></uni-icons>
        </view>
        <text class="header-title">我的收藏</text>
        <view class="header-right"></view>
      </view>
    </view>

    <!-- 占位区域 -->
    <view class="header-placeholder" :style="{ height: (statusBarHeight + navBarHeight) + 'px' }"></view>

    <!-- 商品列表 -->
    <view class="product-list" v-if="favorites.length > 0">
      <view class="product-item" v-for="item in favorites" :key="item.id" @click="goToProductDetail(item.product_id)">
        <image :src="item.product_image || (item.product && item.product.images && item.product.images[0]) || 'https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'" class="product-image" mode="aspectFill"></image>
        <view class="product-info">
          <text class="product-name">{{ item.product_name || (item.product && item.product.name) || '商品名称' }}</text>
          <text class="product-desc" v-if="item.product && item.product.description">{{ item.product.description }}</text>
          <view class="product-bottom">
            <view class="price-row">
              <text class="product-price">¥{{ getDisplayPrice(item) }}</text>
            </view>
            <view class="action-buttons">
              <view class="add-btn" @click.stop="onAddBtnClick(item)">
                <uni-icons type="plusempty" size="18" color="#fff"></uni-icons>
              </view>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 空状态 -->
    <view class="empty-state" v-else-if="!loading">
      <image src="https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg" class="empty-icon" mode="aspectFit"></image>
      <text class="empty-text">暂无收藏商品</text>
      <text class="empty-tip">快去收藏您喜欢的商品吧~</text>
    </view>

    <!-- 加载中 -->
    <view class="loading-state" v-if="loading">
      <text>加载中...</text>
    </view>

    <ProductSelector ref="productSelector" />
  </view>
</template>

<script>
import { getUserFavorites, deleteFavorite } from '../../api/index';
import ProductSelector from '../../components/ProductSelector.vue';

export default {
  components: {
    ProductSelector
  },
  data() {
    return {
      statusBarHeight: 20,
      navBarHeight: 44,
      favorites: [],
      loading: true
    };
  },
  onLoad() {
    const systemInfo = uni.getSystemInfoSync();
    this.statusBarHeight = systemInfo.statusBarHeight || 20;
    
    // 获取胶囊按钮信息
    try {
      const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
      this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight);
    } catch (e) {
      this.navBarHeight = 44;
    }
    
    this.loadFavorites();
  },
  onShow() {
    // 页面显示时重新加载收藏列表
    this.loadFavorites();
  },
  methods: {
    goBack() {
      uni.navigateBack({
        fail: () => {
          uni.switchTab({ url: '/pages/my/my' });
        }
      });
    },
    
    async loadFavorites() {
      this.loading = true;
      try {
        const token = uni.getStorageSync('miniUserToken');
        if (!token) {
          this.favorites = [];
          uni.showToast({
            title: '请先登录',
            icon: 'none'
          });
          return;
        }
        
        const res = await getUserFavorites(token);
        
        if (res && res.code === 200 && res.data) {
          this.favorites = Array.isArray(res.data) ? res.data : [];
        } else {
          this.favorites = [];
        }
      } catch (error) {
        console.error('获取收藏列表失败:', error);
        this.favorites = [];
      } finally {
        this.loading = false;
      }
    },
    
    // 获取显示价格
    getDisplayPrice(item) {
      // 优先使用商品详情中的规格价格
      if (item.product && item.product.specs && item.product.specs.length > 0) {
        const userInfo = uni.getStorageSync('miniUserInfo');
        const isWholesale = userInfo && userInfo.user_type === 'wholesale';
        
        // 取最低价
        const prices = item.product.specs.map(s => {
          if (isWholesale) {
            return parseFloat(s.wholesale_price || s.wholesalePrice || s.retail_price || s.retailPrice || s.price || 0);
          } else {
            return parseFloat(s.retail_price || s.retailPrice || s.wholesale_price || s.wholesalePrice || s.price || 0);
          }
        }).filter(p => p > 0);
        
        if (prices.length > 0) {
          return Math.min(...prices).toFixed(2);
        }
      }
      
      // 使用商品本身的价格
      if (item.product && item.product.price) {
        return parseFloat(item.product.price).toFixed(2);
      }
      
      return '0.00';
    },
    
    goToProductDetail(productId) {
      uni.navigateTo({
        url: '/pages/product/detail?id=' + productId
      });
    },
    
    async deleteFavorite(favoriteId) {
      uni.showModal({
        title: '提示',
        content: '确定要取消收藏吗？',
        success: async (res) => {
          if (res.confirm) {
            try {
              const token = uni.getStorageSync('miniUserToken');
              if (!token) {
                uni.showToast({
                  title: '请先登录',
                  icon: 'none'
                });
                return;
              }
              
              const result = await deleteFavorite(token, favoriteId);
              if (result && result.code === 200) {
                uni.showToast({
                  title: '取消收藏成功',
                  icon: 'success'
                });
                // 重新加载收藏列表
                await this.loadFavorites();
              } else {
                uni.showToast({
                  title: result.message || '取消收藏失败',
                  icon: 'none'
                });
              }
            } catch (error) {
              console.error('删除收藏失败:', error);
              uni.showToast({
                title: '操作失败，请稍后再试',
                icon: 'none'
              });
            }
          }
        }
      });
    },
    
    onAddBtnClick(item) {
      // 构造商品对象传给ProductSelector
      const product = item.product || {
        id: item.product_id,
        name: item.product_name,
        images: item.product_image ? [item.product_image] : []
      };
      this.$refs.productSelector?.open(product);
    }
  }
};
</script>

<style scoped>
.favorites-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.page-header {
  background-color: #20CB6B;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 30rpx;
}

.back-btn {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-title {
  font-size: 32rpx;
  font-weight: 500;
  color: #ffffff;
}

.header-right {
  width: 60rpx;
}

.header-placeholder {
  width: 100%;
}

.product-list {
  padding: 20rpx;
}

.product-item {
  display: flex;
  align-items: flex-start;
  background-color: #fff;
  border-radius: 16rpx;
  padding: 20rpx;
  margin-bottom: 20rpx;
}

.product-image {
  width: 160rpx;
  height: 160rpx;
  border-radius: 12rpx;
  flex-shrink: 0;
}

.product-info {
  flex: 1;
  margin-left: 20rpx;
  min-width: 0;
  display: flex;
  flex-direction: column;
  min-height: 160rpx;
}

.product-name {
  font-size: 28rpx;
  color: #333;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  line-height: 1.4;
  font-weight: 500;
}

.product-desc {
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  overflow: hidden;
}

.product-bottom {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-top: auto;
}

.price-row {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.product-price {
  font-size: 32rpx;
  color: #ff4d4f;
  font-weight: bold;
}

.action-buttons {
  display: flex;
  align-items: center;
}

.add-btn {
  width: 56rpx;
  height: 56rpx;
  background-color: #20CB6B;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 300rpx;
}

.empty-icon {
  width: 200rpx;
  height: 200rpx;
  opacity: 0.5;
}

.empty-text {
  font-size: 32rpx;
  color: #999;
  margin-top: 30rpx;
}

.empty-tip {
  font-size: 26rpx;
  color: #ccc;
  margin-top: 16rpx;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding-top: 300rpx;
  color: #999;
}
</style>

