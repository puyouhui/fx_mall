<template>
  <view class="search-page">
    <!-- 搜索头部 -->
    <view class="search-header">
      <!-- 搜索输入区域 -->
      <view class="search-input-area">
        <view class="search-wrapper">
          <view class="search-input-container">
            <uni-icons type="search" size="16" color="#999" class="search-icon"></uni-icons>
            <input type="text" v-model="searchText" placeholder="9折热销推荐" placeholder-style="color: #999;" @confirm="performSearch" @input="onSearchInput" />
          </view>
          <view class="search-btn" @click.stop="performSearch">搜索</view>
        </view>
      </view>
    </view>

    <!-- 搜索内容区域 -->
    <view class="search-content">
      <!-- 搜索建议列表 -->
      <view class="search-suggestions" v-if="searchText.trim() && suggestions.length > 0">
        <view class="suggestions-list">
          <view class="suggestion-item" v-for="(suggestion, index) in suggestions" :key="index" @click="selectSuggestion(suggestion)">
            <uni-icons type="search" size="14" color="#999" class="suggestion-icon"></uni-icons>
            <text class="suggestion-text">{{ suggestion }}</text>
          </view>
        </view>
      </view>

      <!-- 热门搜索 -->
      <view class="hot-search" v-if="!searchText.trim()">
        <view class="section-title">
          <text class="title-text">热门搜索</text>
        </view>
        <view class="hot-tags">
          <view class="tag-item" v-for="(tag, index) in hotSearchTags" :key="index" @click="searchByTag(tag)">
            {{ tag }}
          </view>
        </view>
      </view>

      <!-- 商品推荐区域 - 并排布局，可滑动 -->
      <scroll-view class="recommendation-scroll" scroll-x="true" v-if="!searchText.trim()" show-scrollbar="false">
        <view class="recommendation-container">
          <!-- 超值推荐 -->
          <view class="special-offers">
            <view class="section-title special-title">
              <view class="title-left">
                <text class="title-text">超值推荐</text>
              </view>
              <view class="more">
                <text class="more-text">全部</text>
                <uni-icons type="right" size="14" color="#999"></uni-icons>
              </view>
            </view>
            <view class="special-product-list">
              <view class="product-item" v-for="(product, index) in specialProducts" :key="index" @click="goToProductDetail(product.id)">
                <image :src="product.images[0] || '/static/test/product1.jpg'" class="product-image" mode="aspectFill"></image>
                <view class="product-info">
                  <text class="product-name">{{ product.name }}</text>
                  <view class="product-bottom-info">
                    <text class="product-price">¥{{ product.displayPrice || product.price }}</text>
                    <view class="add-btn" @click.stop="onAddBtnClick(product)">
                      <uni-icons type="plusempty" size="18" color="#fff"></uni-icons>
                    </view>
                  </view>
                </view>
              </view>
            </view>
          </view>

          <!-- 处理商品 -->
          <view class="processing-products">
            <view class="section-title processing-title">
              <view class="title-left">
                <!-- <view class="bag-icon">
                  <text class="bag-text">🛍</text>
                </view> -->
                <text class="title-text">处理商品</text>
              </view>
              <view class="more">
                <text class="more-text">全部</text>
                <uni-icons type="right" size="14" color="#999"></uni-icons>
              </view>
            </view>
            <view class="processing-product-list">
              <view class="product-item" v-for="(product, index) in processingProducts" :key="index" @click="goToProductDetail(product.id)">
                <image :src="product.images[0] || '/static/test/product1.jpg'" class="product-image" mode="aspectFill"></image>
                <view class="product-info">
                  <view class="product-name-row">
                    <text class="product-name">{{ product.name }}</text>
                    <view class="trust-badge">放心购</view>
                  </view>
                  <view class="product-bottom-info">
                    <text class="product-price">¥{{ product.displayPrice || product.price }}</text>
                    <view class="add-btn" @click.stop="onAddBtnClick(product)">
                      <uni-icons type="plusempty" size="18" color="#fff"></uni-icons>
                    </view>
                  </view>
                </view>
              </view>
            </view>
          </view>
        </view>
      </scroll-view>
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
          <image :src="selectedProduct?.images[0] || ''" class="modal-product-image" mode="aspectFill"></image>
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
import { getProductDetail, searchProducts, searchProductSuggestions } from '../../api/products';

export default {
  data() {
    return {
      statusBarHeight: 0, // 状态栏高度
      navBarHeight: 45, // 导航栏高度（默认值）
      windowWidth: 375, // 胶囊按钮左侧宽度（默认值，用于计算搜索框宽度）
      systemInfo: {}, // 系统信息
      searchText: '',
      suggestions: [], // 搜索建议列表
      hotSearchTags: ['火锅食材', '调味品', '饮料', '零食', '水果', '蔬菜', '肉类', '乳制品'],
      specialProducts: [
        { id: 1, name: '精选澳洲牛肉卷', price: '98.99', images: ['/static/test/product1.jpg'] },
        { id: 2, name: '有机蔬菜礼盒', price: '128.00', images: ['/static/test/product2.jpg'] },
        { id: 3, name: '进口水果拼盘', price: '168.00', images: ['/static/test/product3.jpg'] },
        { id: 4, name: '优质大米5kg', price: '88.00', images: ['/static/test/product4.jpg'] }
      ],
      processingProducts: [
        { id: 5, name: '临期面包组合', price: '38.00', images: ['/static/test/product5.jpg'] },
        { id: 6, name: '促销酸奶12盒', price: '59.90', images: ['/static/test/product6.jpg'] },
        { id: 7, name: '打折巧克力礼盒', price: '79.00', images: ['/static/test/product7.jpg'] },
        { id: 8, name: '特价坚果礼盒', price: '99.00', images: ['/static/test/product8.jpg'] }
      ],
      // 弹窗相关状态
      showProductModal: false,
      selectedProduct: null,
      selectedSpec: null,
      quantity: 1,
      loadingProduct: false
    };
  },
  onLoad() {
    // 获取设备信息，设置状态栏高度
    const systemInfo = uni.getSystemInfoSync();
    this.systemInfo = systemInfo;
    this.statusBarHeight = systemInfo.statusBarHeight || 0;
    this.windowWidth = systemInfo.windowWidth || 375;
    
    // 获取胶囊按钮信息，计算搜索框可用宽度
    this.getMenuButtonInfo();
    
    // 计算价格范围
    this.specialProducts.forEach(product => {
      this.calculateProductPriceRange(product);
    });
    this.processingProducts.forEach(product => {
      this.calculateProductPriceRange(product);
    });
  },
  methods: {
    // 输入框输入事件（获取搜索建议）
    async onSearchInput() {
      const keyword = this.searchText.trim();
      if (keyword) {
        // 获取搜索建议
        await this.getSearchSuggestions(keyword);
      } else {
        // 清空建议列表
        this.suggestions = [];
      }
    },
    
    // 获取搜索建议
    async getSearchSuggestions(keyword) {
      try {
        const res = await searchProductSuggestions(keyword, 10);
        if (res.code === 200 && res.data) {
          this.suggestions = res.data || [];
        } else {
          this.suggestions = [];
        }
      } catch (error) {
        console.error('获取搜索建议失败:', error);
        this.suggestions = [];
      }
    },
    
    // 选择搜索建议
    selectSuggestion(suggestion) {
      this.searchText = suggestion;
      this.suggestions = [];
      // 执行搜索
      this.performSearch();
    },
    
    // 执行搜索（点击搜索按钮或回车时触发，跳转到搜索结果页面）
    performSearch() {
      const keyword = this.searchText.trim();
      if (keyword) {
        // 跳转到搜索结果页面
        uni.navigateTo({
          url: `/pages/search/results?keyword=${encodeURIComponent(keyword)}`
        });
      } else {
        uni.showToast({
          title: '请输入搜索关键词',
          icon: 'none',
          duration: 2000
        });
      }
    },
    
    // 根据热门标签搜索
    searchByTag(tag) {
      this.searchText = tag;
      this.performSearch();
    },
    
    // 生成模拟搜索结果
    generateMockSearchResults(keyword) {
      const mockResults = [
        { id: 101, name: `精选${keyword}1`, price: '128.00', images: ['/static/test/product1.jpg'] },
        { id: 102, name: `优质${keyword}2`, price: '98.00', images: ['/static/test/product2.jpg'] },
        { id: 103, name: `新鲜${keyword}3`, price: '158.00', images: ['/static/test/product3.jpg'] }
      ];
      mockResults.forEach(product => {
        this.calculateProductPriceRange(product);
      });
      return mockResults;
    },
    
    // 计算商品价格范围
    calculateProductPriceRange(product) {
      if (product.specs && product.specs.length > 0) {
        // 过滤出有价格的规格
        const pricedSpecs = product.specs.filter(spec => spec.price !== undefined && spec.price !== null);
        
        if (pricedSpecs.length > 0) {
          const minPrice = Math.min(...pricedSpecs.map(spec => spec.price));
          const maxPrice = Math.max(...pricedSpecs.map(spec => spec.price));
          
          // 设置价格范围显示
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
        // 显示加载状态
        this.loadingProduct = true;
        uni.showLoading({
          title: '加载中',
          mask: true
        });
        
        // 调用接口获取完整的商品详情
        const res = await getProductDetail(parseInt(product.id));
        if (res.code === 200 && res.data) {
          // 处理返回的商品数据
          const productDetail = res.data;
          
          // 处理数据结构差异，确保有规格数据
          if (!productDetail.specs && productDetail.specifications && productDetail.specifications.length > 0) {
            if (productDetail.specifications[0].price === undefined) {
              // 如果specifications没有价格信息，创建默认的specs结构
              productDetail.specs = productDetail.specifications.map((spec, index) => ({
                id: index + 1,
                name: spec.name,
                description: spec.value || '',
                price: parseFloat(productDetail.price) || 0
              }));
            } else {
              // 直接使用specifications作为specs
              productDetail.specs = productDetail.specifications;
            }
          }
          
          // 确保所有规格都有id
          if (productDetail.specs && productDetail.specs.length > 0) {
            productDetail.specs.forEach((spec, index) => {
              if (spec.id === undefined) {
                spec.id = index + 1;
              }
            });
          }
          
          // 确保价格字段正确
          if (productDetail.price !== undefined) {
            productDetail.price = parseFloat(productDetail.price) || 0;
          }
          
          // 计算价格范围
          this.calculateProductPriceRange(productDetail);
          
          // 设置选中的商品
          this.selectedProduct = productDetail;
          // 默认选择第一个规格
          this.selectedSpec = productDetail.specs && productDetail.specs.length > 0 ? productDetail.specs[0] : null;
          // 重置数量
          this.quantity = 1;
          // 显示弹窗
          this.showProductModal = true;
        } else {
          // 商品不存在，显示错误提示
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
        // 隐藏加载动画
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
    
    // 获取胶囊按钮信息并计算导航栏高度
    getMenuButtonInfo() {
      try {
        // #ifndef H5 || APP-PLUS || MP-ALIPAY
        // 获取胶囊的位置信息
        const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
        // 计算导航栏高度
        this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight);
        // 胶囊按钮左侧的宽度，用于限制搜索框宽度
        this.windowWidth = menuButtonInfo.left;
        // #endif
      } catch (error) {
        console.error('获取胶囊按钮信息失败:', error);
      }
    },
    
    // 返回上一页
    goBack() {
      uni.navigateBack({
        fail: () => {
          // 如果无法返回，则跳转到首页
          uni.switchTab({
            url: '/pages/index/index'
          });
        }
      });
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
      
      // 获取购物车数据
      let cart = uni.getStorageSync('cart') || [];
      
      // 构建商品信息
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
      
      // 检查商品是否已在购物车中
      const existingItemIndex = cart.findIndex(item => 
        item.productId === cartItem.productId && item.specKey === cartItem.specKey
      );
      
      if (existingItemIndex >= 0) {
        // 已存在则增加数量
        cart[existingItemIndex].quantity += cartItem.quantity;
      } else {
        // 不存在则添加新商品
        cart.push(cartItem);
      }
      
      // 保存到本地存储
      uni.setStorageSync('cart', cart);
      
      // 显示成功提示
      uni.showToast({
        title: '已添加到采购单',
        icon: 'success'
      });
      
      // 关闭弹窗
      this.closeProductModal();
    }
  }
};
</script>

<style>
page{
  background-color: #F8F8F8;
}
</style>

<style scoped>
.search-page {
  min-height: 100vh;
  background-color: #F8F8F8;
  display: flex;
  flex-direction: column;
}

/* 搜索头部样式 */
.search-header {
  border-bottom: 1rpx solid #eee;
  position: sticky;
  top: 0;
  z-index: 100;
  box-sizing: border-box;
}

/* 导航栏样式 */
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10rpx 20rpx;
  height: 88rpx;
  box-sizing: border-box;
}

.back-btn {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  cursor: pointer;
}

.navbar-title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-size: 32rpx;
  font-weight: 500;
  color: #333;
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 10rpx;
  flex-shrink: 0;
}

.navbar-icon-btn {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #fff;
  border-radius: 50%;
  cursor: pointer;
}

.icon-dots {
  font-size: 36rpx;
  color: #333;
  line-height: 1;
  font-weight: normal;
}

.icon-scan {
  font-size: 28rpx;
  color: #333;
  line-height: 1;
  font-weight: normal;
}

/* 搜索输入区域样式 */
.search-input-area {
  padding: 20rpx;
  padding-top: 10rpx;
  box-sizing: border-box;
}

.search-wrapper {
  display: flex;
  align-items: center;
  height: 70rpx;
  background-color: #fff;
  border: 1rpx solid #eee;
  border-radius: 10rpx;
  overflow: hidden;
  box-sizing: border-box;
}

.search-input-container {
  flex: 1;
  display: flex;
  align-items: center;
  height: 100%;
  padding: 0 20rpx;
  background-color: transparent;
  min-width: 0;
}

.search-icon {
  margin-right: 10rpx;
  flex-shrink: 0;
}

.search-input-container input {
  flex: 1;
  height: 100%;
  font-size: 28rpx;
  color: #333;
  background-color: transparent;
  border: none;
  outline: none;
}

.search-btn {
  height: 100%;
  padding: 0 30rpx;
  background-color: #20CB6B;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #fff;
  flex-shrink: 0;
  cursor: pointer;
  white-space: nowrap;
  border-left: 1rpx solid rgba(255, 255, 255, 0.3);
}


/* 搜索内容区域样式 */
.search-content {
  padding-top: 0;
  flex: 1;
  overflow-y: auto;
}

/* 热门搜索样式 */
.hot-search {
  padding: 20rpx;
  border-radius: 20rpx;
  padding: 20rpx;
  margin-bottom: 20rpx;
}

.section-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.title-left {
  display: flex;
  align-items: center;
  gap: 10rpx;
}

.title-text {
  font-size: 30rpx;
  font-weight: bold;
}

.special-title .title-text {
  color: #ff4d4f;
}

.processing-title .title-text {
  color: #20CB6B;
}

.hot-badge {
  width: 50rpx;
  height: 50rpx;
  background-color: #ff4d4f;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.hot-text {
  font-size: 20rpx;
  color: #fff;
  font-weight: bold;
}

.bag-icon {
  width: 50rpx;
  height: 50rpx;
  background-color: #f0f9f4;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bag-text {
  font-size: 24rpx;
}

.more {
  display: flex;
  align-items: center;
  gap: 5rpx;
  font-size: 26rpx;
  color: #999;
}

.more-text {
  font-size: 26rpx;
  color: #999;
}

.hot-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
}

.tag-item {
  padding: 10rpx 25rpx;
  background-color: #fff;
  border-radius: 20rpx;
  font-size: 26rpx;
  color: #666;
}

/* 搜索建议样式 */
.search-suggestions {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 20rpx;
  padding-left: 0;
  margin-bottom: 20rpx;
}

.suggestions-list {
  display: flex;
  flex-direction: column;
}

.suggestion-item {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
  cursor: pointer;
}

.suggestion-item:last-child {
  border-bottom: none;
}

.suggestion-icon {
  margin-right: 15rpx;
  flex-shrink: 0;
}

.suggestion-text {
  flex: 1;
  font-size: 28rpx;
  color: #333;
}

/* 搜索结果样式 */
.search-results {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 20rpx;
  margin-bottom: 20rpx;
}

/* 搜索结果列表布局 - 保持原两行布局 */
.search-result-list {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
}

/* 超值推荐和处理商品列表布局 - 单行布局 */
.special-product-list,
.processing-product-list {
  display: flex;
  flex-direction: column;
}

/* 商品项基础样式 */
.product-item {
  background-color: #fff;
  border-radius: 15rpx;
  margin-bottom: 20rpx;
  position: relative;
}

/* 搜索结果中的商品项 - 保持原两行布局 */
.search-result-list .product-item {
  width: 48%;
}

/* 超值推荐和处理商品中的商品项 - 单行布局 */
.special-product-list .product-item,
.processing-product-list .product-item {
  width: 100%;
  display: flex;
  min-height: 180rpx;
  border: 1rpx solid #f0f0f0;
  border-radius: 10rpx;
  margin-bottom: 15rpx;
  overflow: hidden;
  background-color: #fff;
}

.special-product-list .product-item:last-child,
.processing-product-list .product-item:last-child {
  margin-bottom: 0;
}

/* 搜索结果中的商品图片样式 */
.search-result-list .product-image {
  width: 100%;
  height: 280rpx;
  border-radius: 15rpx 15rpx 0 0;
}

/* 超值推荐和处理商品中的图片样式 */
.special-product-list .product-image,
.processing-product-list .product-image {
  width: 180rpx;
  height: 180rpx;
  border-radius: 0;
  flex-shrink: 0;
  object-fit: cover;
}

/* 搜索结果中的商品信息样式 */
.search-result-list .product-info {
  padding: 15rpx;
  position: relative;
}

/* 超值推荐和处理商品中的信息样式 */
.special-product-list .product-info,
.processing-product-list .product-info {
  flex: 1;
  padding: 15rpx;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  position: relative;
}

.product-name {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  font-size: 26rpx;
  color: #333;
  line-height: 1.4;
  margin-bottom: 10rpx;
}

.product-name-row {
  display: flex;
  align-items: flex-start;
  gap: 10rpx;
  margin-bottom: 10rpx;
}

.product-name-row .product-name {
  flex: 1;
  margin-bottom: 0;
}

.trust-badge {
  padding: 4rpx 12rpx;
  background-color: #f5f5f5;
  border-radius: 4rpx;
  font-size: 20rpx;
  color: #999;
  flex-shrink: 0;
  margin-top: 2rpx;
}

.product-bottom-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.product-price {
  font-size: 28rpx;
  color: #ff4d4f;
  font-weight: bold;
}

.add-btn {
  width: 50rpx;
  height: 50rpx;
  background-color: #20CB6B;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

/* 商品推荐滚动容器 */
.recommendation-scroll {
  width: 100%;
  white-space: nowrap;
  /* margin-left: -20rpx;
  padding-left: 20rpx; */
}

/* 商品推荐容器样式 - 横向并排布局 */
.recommendation-container {
  display: inline-flex;
  gap: 20rpx;
  padding: 0 20rpx 20rpx 20rpx;
}

/* 超值推荐样式 */
.special-offers {
  width: 450rpx;
  min-width: 450rpx;
  background-color: #fff5f5;
  border-radius: 20rpx;
  padding: 20rpx;
  flex-shrink: 0;
  box-sizing: border-box;
}

/* 处理商品样式 */
.processing-products {
  width: 450rpx;
  min-width: 450rpx;
  background-color: #f0f9f4;
  border-radius: 20rpx;
  padding: 20rpx;
  flex-shrink: 0;
  box-sizing: border-box;
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