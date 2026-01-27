<template>
  <view class="results-page">
    <!-- 自定义头部 - 绿色背景 -->
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
            <text class="navbar-title-text">搜索结果</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>
    
    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>
    
    <!-- 搜索结果列表 -->
    <view class="results-container">
      <!-- 商品列表 -->
      <view class="product-list" v-if="searchResults.length > 0">
        <view class="product-item" v-for="(product, index) in searchResults" :key="index" @click="goToProductDetail(product.id)">
          <image :src="product.images && product.images.length > 0 ? product.images[0] : 'https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'" class="product-image" mode="aspectFill"></image>
          <view class="product-info">
            <text class="product-name">{{ product.name }}</text>
            <text class="product-desc" v-if="product.description">{{ product.description }}</text>
            <view class="product-bottom">
              <view class="price-container">
                <text class="product-price">¥{{ product.displayPrice || product.price_range || '暂无价格' }}</text>
              </view>
              <view class="add-btn" @click.stop="onAddBtnClick(product)">
                <uni-icons type="plusempty" size="18" color="#fff"></uni-icons>
              </view>
            </view>
          </view>
        </view>
        <!-- 没有更多商品提示 -->
        <view class="no-more-tip">
          <text class="no-more-text">没有更多商品了</text>
        </view>
      </view>

      <!-- 空状态 -->
      <view class="empty-state" v-else-if="!loading">
        <text class="empty-text">未找到相关商品</text>
        <text class="empty-tip">试试其他关键词吧</text>
      </view>

      <!-- 加载中 -->
      <view class="loading-state" v-if="loading">
        <text class="loading-text">搜索中...</text>
      </view>
    </view>

    <!-- 商品选择器组件 -->
    <ProductSelector ref="productSelector" />
  </view>
</template>

<script>
import { searchProducts } from '../../api/products';
import { getMiniUserInfo } from '../../api/index';
import ProductSelector from '../../components/ProductSelector.vue';

export default {
  components: {
    ProductSelector
  },
  data() {
    return {
      statusBarHeight: 0, // 状态栏高度
      navBarHeight: 45, // 导航栏高度（默认值）
      keyword: '',
      searchResults: [],
      total: 0,
      loading: false,
      userType: null, // 用户类型：'retail' | 'wholesale' | null（未登录）
    };
  },
  onLoad(options) {
    // 获取设备信息，设置状态栏高度
    const systemInfo = uni.getSystemInfoSync();
    this.statusBarHeight = systemInfo.statusBarHeight || 0;
    
    // 获取胶囊按钮信息，计算导航栏高度
    this.getMenuButtonInfo();
    
    // 初始化用户类型
    this.initUserType();
    
    if (options.keyword) {
      this.keyword = decodeURIComponent(options.keyword);
      this.performSearch();
    }
  },
  // 页面显示时更新用户信息
  onShow() {
    this.updateUserInfo();
  },
  methods: {
    // 初始化用户类型
    initUserType() {
      const userInfo = uni.getStorageSync('miniUserInfo');
      if (userInfo && userInfo.user_type) {
        this.userType = userInfo.user_type;
      } else {
        this.userType = null;
      }
    },
    
    // 更新用户信息
    async updateUserInfo() {
      try {
        const token = uni.getStorageSync('miniUserToken');
        if (!token) {
          // 未登录，不获取用户信息
          this.userType = null;
          // 重新计算产品价格
          this.recalculateAllPrices();
          return;
        }

        const res = await getMiniUserInfo(token);
        if (res && res.code === 200 && res.data) {
          // 更新本地存储的用户信息
          uni.setStorageSync('miniUserInfo', res.data);
          if (res.data.unique_id) {
            uni.setStorageSync('miniUserUniqueId', res.data.unique_id);
          }
          // 更新用户类型
          this.userType = res.data.user_type || null;
          // 重新计算产品价格
          this.recalculateAllPrices();
        }
      } catch (error) {
        console.error('获取用户信息失败:', error);
        // 静默失败，不显示错误提示
        this.userType = null;
      }
    },
    
    // 重新计算所有产品价格
    recalculateAllPrices() {
      if (this.searchResults && this.searchResults.length > 0) {
        this.searchResults.forEach(product => {
          this.calculateProductPriceRange(product);
        });
      }
    },
    
    // 执行搜索
    async performSearch() {
      if (!this.keyword.trim()) {
        return;
      }

      this.loading = true;
      this.searchResults = [];

      try {
        const res = await searchProducts(this.keyword, 1, 20);
        if (res.code === 200 && res.data) {
          const productList = res.data.list || [];
          this.total = res.data.total || 0;
          
          if (productList.length > 0) {
            this.searchResults = productList;
            this.searchResults.forEach(product => {
              this.calculateProductPriceRange(product);
            });
          }
        }
      } catch (error) {
        console.error('搜索商品失败:', error);
        uni.showToast({
          title: '搜索失败，请重试',
          icon: 'none',
          duration: 2000
        });
      } finally {
        this.loading = false;
      }
    },

    // 计算商品价格范围（根据用户类型显示批发价或零售价）
    calculateProductPriceRange(product) {
      let specs = product.specs;
      if (typeof specs === 'string') {
        try {
          specs = JSON.parse(specs);
        } catch (error) {
          specs = [];
        }
      }

      if (!Array.isArray(specs) || specs.length === 0) {
        // 如果没有规格数据，根据用户类型返回零售价或批发价
        const basePrice = product.price || 0;
        product.displayPrice = basePrice.toFixed(2);
        return;
      }

      // 根据用户类型决定显示哪种价格
      const isWholesaleUser = this.userType === 'wholesale';
      
      const prices = [];
      specs.forEach(spec => {
        if (isWholesaleUser) {
          // 批发用户：显示批发价
          const wholesalePrice = spec.wholesale_price ?? spec.wholesalePrice;
          if (wholesalePrice > 0) prices.push(parseFloat(wholesalePrice));
        } else {
          // 未登录或零售用户：显示零售价
          const retailPrice = spec.retail_price ?? spec.retailPrice;
          if (retailPrice > 0) prices.push(parseFloat(retailPrice));
        }
      });

      // 如果没有找到对应类型的价格，使用另一种价格作为后备
      if (prices.length === 0) {
        specs.forEach(spec => {
          if (isWholesaleUser) {
            // 批发用户找不到批发价，使用零售价作为后备
            const retailPrice = spec.retail_price ?? spec.retailPrice;
            if (retailPrice > 0) prices.push(parseFloat(retailPrice));
          } else {
            // 零售用户找不到零售价，使用批发价作为后备
            const wholesalePrice = spec.wholesale_price ?? spec.wholesalePrice;
            if (wholesalePrice > 0) prices.push(parseFloat(wholesalePrice));
          }
          // 最后使用通用价格字段
          if (spec.price && spec.price > 0 && prices.length === 0) {
            prices.push(parseFloat(spec.price));
          }
        });
      }

      if (prices.length === 0) {
        // 如果没有找到对应用户类型的价格，回退到通用价格
        product.displayPrice = (product.price || 0).toFixed(2);
        return;
      }

      // 显示价格范围（最低价~最高价）
      const minPrice = Math.min(...prices);
      const maxPrice = Math.max(...prices);
      
      if (minPrice === maxPrice) {
        // 如果所有规格价格相同，只显示一个价格
        product.displayPrice = minPrice.toFixed(2);
      } else {
        // 显示价格范围
        product.displayPrice = `${minPrice.toFixed(2)}~${maxPrice.toFixed(2)}`;
      }
    },

    // 跳转到商品详情
    goToProductDetail(productId) {
      uni.navigateTo({
        url: '/pages/product/detail?id=' + productId
      });
    },

    // 显示商品选择弹窗（使用 ProductSelector 组件）
    onAddBtnClick(product) {
      // 使用 ProductSelector 组件打开商品选择器
      this.$refs.productSelector?.open(product);
    },
    
    // 获取胶囊按钮信息并计算导航栏高度
    getMenuButtonInfo() {
      try {
        // #ifndef H5 || APP-PLUS || MP-ALIPAY
        // 获取胶囊的位置信息
        const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
        // 计算导航栏高度
        this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight);
        // #endif
      } catch (error) {
        console.error('获取胶囊按钮信息失败:', error);
      }
    },
    
    // 返回上一页
    goBack() {
      uni.navigateBack({
        fail: () => {
          // 如果无法返回，则跳转到搜索页
          uni.navigateTo({
            url: '/pages/search/search'
          });
        }
      });
    }
  }
};
</script>

<style scoped>
.results-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

/* 自定义头部样式 */
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
}

.navbar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20rpx;
  box-sizing: border-box;
}

.navbar-left {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  cursor: pointer;
}

.navbar-title {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.navbar-title-text {
  font-size: 32rpx;
  font-weight: 500;
  color: #fff;
}

.navbar-right {
  width: 60rpx;
  flex-shrink: 0;
}

.results-container {
  padding: 20rpx;
}

.product-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.product-item {
  background-color: #fff;
  border-radius: 20rpx;
  display: flex;
  padding: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.08);
  transition: transform 0.2s, box-shadow 0.2s;
}

.product-item:active {
  transform: scale(0.98);
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.12);
}

.product-image {
  width: 200rpx;
  height: 200rpx;
  border-radius: 16rpx;
  flex-shrink: 0;
  margin-right: 24rpx;
  background-color: #f5f5f5;
}

.product-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.product-name {
  font-size: 30rpx;
  color: #333;
  font-weight: 500;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  margin-bottom: 12rpx;
  line-height: 1.5;
}

.product-desc {
  font-size: 24rpx;
  color: #999;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  overflow: hidden;
  margin-bottom: 15rpx;
}

.product-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.price-container {
  display: flex;
  align-items: baseline;
}

.product-price {
  font-size: 32rpx;
  color: #f00;
  font-weight: bold;
}

.add-btn {
  width: 60rpx;
  height: 60rpx;
  background-color: #20CB6B;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4rpx 12rpx rgba(32, 203, 107, 0.3);
}

.no-more-tip {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40rpx 20rpx;
  margin-top: 20rpx;
}

.no-more-text {
  font-size: 26rpx;
  color: #999;
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
  color: #666;
  margin-bottom: 16rpx;
  font-weight: 500;
}

.empty-tip {
  font-size: 26rpx;
  color: #999;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 100rpx 20rpx;
}

.loading-text {
  font-size: 28rpx;
  color: #999;
}

</style>

