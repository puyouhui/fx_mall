<template>
  <view class="results-page">
    <!-- 搜索结果列表 -->
    <view class="results-container">
      <!-- 商品列表 -->
      <view class="product-list" v-if="searchResults.length > 0">
        <view class="product-item" v-for="(product, index) in searchResults" :key="index" @click="goToProductDetail(product.id)">
          <image :src="product.images && product.images.length > 0 ? product.images[0] : '/static/test/product1.jpg'" class="product-image" mode="aspectFill"></image>
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

    <!-- 商品选择弹窗 -->
    <view class="product-modal" v-if="showProductModal" @click="closeProductModal">
      <view class="modal-overlay"></view>
      <view class="modal-content" @click.stop>
        <!-- 弹窗头部 -->
        <view class="modal-header">
          <text class="modal-title">选择规格</text>
          <view class="modal-close" @click.stop="closeProductModal">
            <uni-icons type="close" size="24" color="#999"></uni-icons>
          </view>
        </view>

        <!-- 商品信息 -->
        <view class="product-modal-info">
          <image :src="selectedProduct?.images && selectedProduct.images.length > 0 ? selectedProduct.images[0] : ''" class="modal-product-image" mode="aspectFill"></image>
          <view class="modal-product-details">
            <text class="modal-product-name">{{ selectedProduct?.name }}</text>
            <text class="modal-product-price">¥{{ selectedProduct?.displayPrice || '暂无价格' }}</text>
          </view>
        </view>

        <!-- 规格选择 -->
        <view class="specs-section" v-if="selectedProduct?.specs && selectedProduct.specs.length > 0">
          <text class="specs-title">选择规格</text>
          <view class="specs-list">
            <view class="spec-item" v-for="(spec, index) in selectedProduct.specs" :key="index" :class="{ 'selected': selectedSpec && selectedSpec.name === spec.name }" @click.stop="selectSpec(spec)">
              <text class="spec-name">{{ spec.name }}</text>
              <text class="spec-description" v-if="spec.description">({{ spec.description }})</text>
              <text class="spec-price" v-if="spec.price">¥{{ spec.price.toFixed(2) }}</text>
            </view>
          </view>
        </view>

        <!-- 数量选择 -->
        <view class="quantity-section">
          <text class="quantity-title">数量</text>
          <view class="quantity-selector">
            <view class="minus-btn" @click.stop="decreaseQuantity">
              <image src="/static/icon/minus.png" class="minus-btn-icon"></image>
            </view>
            <text class="quantity-text">{{ quantity }}</text>
            <view class="plus-btn" @click.stop="increaseQuantity">
              <uni-icons type="plusempty" size="18" color="#fff"></uni-icons>
            </view>
          </view>
        </view>

        <!-- 底部按钮 -->
        <view class="modal-bottom">
          <view class="buy-btn" @click.stop="addToCart">
            <text>加入采购单</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { getProductDetail, searchProducts } from '../../api/products';

export default {
  data() {
    return {
      keyword: '',
      searchResults: [],
      total: 0,
      loading: false,
      // 弹窗相关状态
      showProductModal: false,
      selectedProduct: null,
      selectedSpec: null,
      quantity: 1,
      loadingProduct: false
    };
  },
  onLoad(options) {
    if (options.keyword) {
      this.keyword = decodeURIComponent(options.keyword);
      this.performSearch();
    }
  },
  methods: {
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

    // 计算商品价格范围
    calculateProductPriceRange(product) {
      if (product.specs && product.specs.length > 0) {
        const pricedSpecs = product.specs.filter(spec => spec.price !== undefined && spec.price !== null);
        
        if (pricedSpecs.length > 0) {
          const minPrice = Math.min(...pricedSpecs.map(spec => spec.price));
          const maxPrice = Math.max(...pricedSpecs.map(spec => spec.price));
          
          if (minPrice === maxPrice) {
            product.displayPrice = minPrice.toFixed(2);
          } else {
            product.displayPrice = minPrice.toFixed(2) + '~' + maxPrice.toFixed(2);
          }
        } else if (product.price) {
          product.displayPrice = parseFloat(product.price).toFixed(2);
        }
      } else if (product.price) {
        product.displayPrice = parseFloat(product.price).toFixed(2);
      }
    },

    // 跳转到商品详情
    goToProductDetail(productId) {
      uni.navigateTo({
        url: '/pages/product/detail?id=' + productId
      });
    },

    // 显示商品选择弹窗
    async onAddBtnClick(product) {
      try {
        this.loadingProduct = true;
        uni.showLoading({
          title: '加载中',
          mask: true
        });
        
        const res = await getProductDetail(parseInt(product.id));
        if (res.code === 200 && res.data) {
          const productDetail = res.data;
          
          if (!productDetail.specs && productDetail.specifications && productDetail.specifications.length > 0) {
            if (productDetail.specifications[0].price === undefined) {
              productDetail.specs = productDetail.specifications.map((spec, index) => ({
                id: index + 1,
                name: spec.name,
                description: spec.value || '',
                price: parseFloat(productDetail.price) || 0
              }));
            } else {
              productDetail.specs = productDetail.specifications;
            }
          }
          
          if (productDetail.specs && productDetail.specs.length > 0) {
            productDetail.specs.forEach((spec, index) => {
              if (spec.id === undefined) {
                spec.id = index + 1;
              }
            });
          }
          
          if (productDetail.price !== undefined) {
            productDetail.price = parseFloat(productDetail.price) || 0;
          }
          
          this.calculateProductPriceRange(productDetail);
          this.selectedProduct = productDetail;
          this.selectedSpec = productDetail.specs && productDetail.specs.length > 0 ? productDetail.specs[0] : null;
          this.quantity = 1;
          this.showProductModal = true;
        } else {
          uni.showToast({
            title: '商品不存在',
            icon: 'none',
            duration: 2000
          });
        }
      } catch (error) {
        console.error('加载商品详情失败:', error);
        uni.showToast({
          title: '加载失败，请重试',
          icon: 'none',
          duration: 2000
        });
      } finally {
        this.loadingProduct = false;
        uni.hideLoading();
      }
    },

    // 关闭弹窗
    closeProductModal() {
      this.showProductModal = false;
    },

    // 选择规格
    selectSpec(spec) {
      this.selectedSpec = spec;
    },

    // 增加数量
    increaseQuantity() {
      this.quantity++;
    },

    // 减少数量
    decreaseQuantity() {
      if (this.quantity > 1) {
        this.quantity--;
      }
    },

    // 添加到购物车
    addToCart() {
      if (!this.selectedSpec) {
        uni.showToast({
          title: '请选择商品规格',
          icon: 'none'
        });
        return;
      }
      
      let cart = uni.getStorageSync('cart') || [];
      
      const cartItem = {
        productId: this.selectedProduct.id,
        productName: this.selectedProduct.name,
        productImage: this.selectedProduct.images && this.selectedProduct.images.length > 0 ? this.selectedProduct.images[0] : '',
        specKey: this.selectedSpec.name + (this.selectedSpec.description ? ':' + this.selectedSpec.description : ''),
        specName: this.selectedSpec.name,
        specDescription: this.selectedSpec.description,
        price: this.selectedSpec.price || this.selectedProduct.price,
        quantity: this.quantity
      };
      
      const existingItemIndex = cart.findIndex(item => 
        item.productId === cartItem.productId && item.specKey === cartItem.specKey
      );
      
      if (existingItemIndex >= 0) {
        cart[existingItemIndex].quantity += cartItem.quantity;
      } else {
        cart.push(cartItem);
      }
      
      uni.setStorageSync('cart', cart);
      
      uni.showToast({
        title: '已添加到采购单',
        icon: 'success'
      });
      
      this.closeProductModal();
    }
  }
};
</script>

<style scoped>
.results-page {
  min-height: 100vh;
  background-color: #f5f5f5;
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
  border-radius: 15rpx;
  display: flex;
  padding: 20rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.05);
}

.product-image {
  width: 200rpx;
  height: 200rpx;
  border-radius: 10rpx;
  flex-shrink: 0;
  margin-right: 20rpx;
}

.product-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.product-name {
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  margin-bottom: 10rpx;
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
  color: #ff4d4f;
  font-weight: bold;
}

.add-btn {
  width: 60rpx;
  height: 60rpx;
  background-color: #ff4d4f;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
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
  padding: 100rpx 20rpx;
}

.empty-text {
  font-size: 32rpx;
  color: #999;
  margin-bottom: 20rpx;
}

.empty-tip {
  font-size: 26rpx;
  color: #ccc;
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

/* 商品选择弹窗样式 */
.product-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 999;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
}

.modal-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
}

.modal-content {
  background-color: #fff;
  border-radius: 20rpx 20rpx 0 0;
  padding: 20rpx;
  position: relative;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.modal-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.modal-product-info {
  display: flex;
  margin-bottom: 30rpx;
}

.modal-product-image {
  width: 200rpx;
  height: 200rpx;
  border-radius: 10rpx;
  margin-right: 20rpx;
  flex-shrink: 0;
}

.modal-product-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.modal-product-name {
  font-size: 28rpx;
  color: #333;
  margin-bottom: 10rpx;
}

.modal-product-price {
  font-size: 32rpx;
  color: #ff4d4f;
  font-weight: bold;
}

.specs-section {
  margin-bottom: 30rpx;
}

.specs-title {
  font-size: 28rpx;
  color: #333;
  margin-bottom: 15rpx;
  display: block;
}

.specs-list {
  display: flex;
  flex-wrap: wrap;
  gap: 15rpx;
}

.spec-item {
  padding: 10rpx 20rpx;
  border: 1rpx solid #ddd;
  border-radius: 10rpx;
  font-size: 26rpx;
  color: #666;
}

.spec-item.selected {
  border-color: #ff4d4f;
  color: #ff4d4f;
  background-color: #fff5f5;
}

.spec-description {
  color: #999;
  margin: 0 5rpx;
}

.spec-price {
  color: #ff4d4f;
  margin-left: 10rpx;
}

.quantity-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30rpx;
}

.quantity-title {
  font-size: 28rpx;
  color: #333;
}

.quantity-selector {
  display: flex;
  align-items: center;
}

.minus-btn {
  width: 60rpx;
  height: 60rpx;
  background-color: #f5f5f5;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 20rpx;
}

.minus-btn-icon {
  width: 30rpx;
  height: 6rpx;
}

.quantity-text {
  font-size: 30rpx;
  color: #333;
  margin-right: 20rpx;
}

.plus-btn {
  width: 60rpx;
  height: 60rpx;
  background-color: #ff4d4f;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-bottom {
  margin-top: 20rpx;
}

.buy-btn {
  width: 100%;
  height: 90rpx;
  background-color: #ff4d4f;
  border-radius: 45rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  color: #fff;
}
</style>

