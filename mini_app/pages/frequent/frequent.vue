<template>
  <view class="frequent-page">
    <!-- 头部 -->
    <view class="page-header" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="header-content" :style="{ height: navBarHeight + 'px' }">
        <view class="back-btn" @click="goBack">
          <uni-icons type="left" size="24" color="#fff"></uni-icons>
        </view>
        <text class="header-title">我常买</text>
        <view class="header-right"></view>
      </view>
    </view>

    <!-- 占位区域 -->
    <view class="header-placeholder" :style="{ height: (statusBarHeight + navBarHeight) + 'px' }"></view>

    <!-- 商品列表 -->
    <view class="product-list" v-if="products.length > 0">
      <view class="product-item" v-for="item in products" :key="item.product_id + '-' + item.spec_name" @click="goToProductDetail(item.product_id)">
        <image :src="item.image || '/static/test/product1.jpg'" class="product-image" mode="aspectFill"></image>
        <view class="product-info">
          <text class="product-name">{{ item.product_name }}</text>
          <text class="product-spec" v-if="item.spec_name">{{ item.spec_name }}</text>
          <view class="product-bottom">
            <view class="price-row">
              <text class="product-price">¥{{ getDisplayPrice(item) }}</text>
              <text class="buy-count">已买{{ item.buy_count }}次</text>
            </view>
            <view class="add-btn" @click.stop="onAddBtnClick(item)">
              <uni-icons type="plusempty" size="18" color="#fff"></uni-icons>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 空状态 -->
    <view class="empty-state" v-else-if="!loading">
      <image src="/static/icon/empty-cart.png" class="empty-icon" mode="aspectFit"></image>
      <text class="empty-text">暂无常购商品</text>
      <text class="empty-tip">下单后这里会显示您购买过的商品</text>
    </view>

    <!-- 加载中 -->
    <view class="loading-state" v-if="loading">
      <text>加载中...</text>
    </view>

    <ProductSelector ref="productSelector" />
  </view>
</template>

<script>
import { get } from '../../api/request';
import ProductSelector from '../../components/ProductSelector.vue';

export default {
  components: {
    ProductSelector
  },
  data() {
    return {
      statusBarHeight: 20,
      navBarHeight: 44,
      products: [],
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
    
    this.loadFrequentProducts();
  },
  methods: {
    goBack() {
      uni.navigateBack({
        fail: () => {
          uni.switchTab({ url: '/pages/cart/cart' });
        }
      });
    },
    
    async loadFrequentProducts() {
      this.loading = true;
      try {
        const token = uni.getStorageSync('miniUserToken');
        if (!token) {
          this.products = [];
          return;
        }
        
        const res = await get('/mini-app/users/frequent-products', {}, {
          header: { Authorization: `Bearer ${token}` }
        });
        
        if (res && res.code === 200 && res.data) {
          this.products = res.data || [];
        }
      } catch (error) {
        console.error('获取常购商品失败:', error);
      } finally {
        this.loading = false;
      }
    },
    
    // 获取显示价格
    getDisplayPrice(item) {
      // 优先使用商品详情中的规格价格
      if (item.product && item.product.specs) {
        const userInfo = uni.getStorageSync('miniUserInfo');
        const isWholesale = userInfo && userInfo.user_type === 'wholesale';
        
        // 找到对应规格
        const spec = item.product.specs.find(s => s.name === item.spec_name);
        if (spec) {
          if (isWholesale) {
            return (spec.wholesale_price || spec.retail_price || 0).toFixed(2);
          } else {
            return (spec.retail_price || spec.wholesale_price || 0).toFixed(2);
          }
        }
        
        // 取最低价
        const prices = item.product.specs.map(s => 
          isWholesale ? (s.wholesale_price || s.retail_price || 0) : (s.retail_price || s.wholesale_price || 0)
        ).filter(p => p > 0);
        
        if (prices.length > 0) {
          return Math.min(...prices).toFixed(2);
        }
      }
      
      return '0.00';
    },
    
    goToProductDetail(productId) {
      uni.navigateTo({
        url: '/pages/product/detail?id=' + productId
      });
    },
    
    onAddBtnClick(item) {
      // 构造商品对象传给ProductSelector
      const product = item.product || {
        id: item.product_id,
        name: item.product_name,
        images: [item.image]
      };
      this.$refs.productSelector?.open(product);
    }
  }
};
</script>

<style scoped>
.frequent-page {
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
}

.product-spec {
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
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

.buy-count {
  font-size: 22rpx;
  color: #20CB6B;
  background-color: #e8f8ef;
  padding: 4rpx 12rpx;
  border-radius: 20rpx;
  display: inline-block;
  width: fit-content;
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

