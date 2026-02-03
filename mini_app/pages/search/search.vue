<template>
  <view class="search-page">
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
            <text class="navbar-title-text">搜索</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>
    
    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 搜索内容区域 -->
    <view class="search-content">
      <!-- 搜索输入区域 -->
      <view class="search-input-area">
        <view class="search-wrapper">
          <view class="search-input-container">
            <uni-icons type="search" size="16" color="#999" class="search-icon"></uni-icons>
            <input 
              ref="searchInput" 
              type="text" 
              v-model="searchText" 
              placeholder="下拉抽纸" 
              placeholder-style="color: #999;" 
              @confirm="performSearch" 
              @input="onSearchInput"
              :focus="isInputFocused"
            />
            <view class="clear-btn" v-if="searchText.trim()" @click.stop="clearSearchText">
              <uni-icons type="clear" size="16" color="#999"></uni-icons>
            </view>
          </view>
          <view class="search-btn" @click.stop="performSearch">搜索</view>
        </view>
      </view>
      <!-- 搜索建议列表 -->
      <view class="search-suggestions" v-if="searchText.trim() && suggestions.length > 0">
        <view class="suggestions-list">
          <view class="suggestion-item" v-for="(suggestion, index) in suggestions" :key="index" @click="selectSuggestion(suggestion)">
            <view class="suggestion-icon-wrapper">
              <uni-icons type="search" size="16" color="#20CB6B" class="suggestion-icon"></uni-icons>
            </view>
            <text class="suggestion-text">{{ suggestion }}</text>
            <view class="suggestion-arrow">
              <uni-icons type="right" size="14" color="#ccc"></uni-icons>
            </view>
          </view>
        </view>
      </view>

      <!-- 搜索历史 -->
      <view class="search-history" v-if="!searchText.trim() && searchHistory.length > 0">
        <view class="section-title">
          <text class="title-text">搜索历史</text>
          <view class="clear-history" @click="clearSearchHistory">
            <uni-icons type="trash" size="16" color="#999"></uni-icons>
            <text class="clear-text">清除</text>
          </view>
        </view>
        <view class="history-tags">
          <view class="tag-item history-item" v-for="(item, index) in searchHistory" :key="index" @click="selectHistoryItem(item)">
            {{ item }}
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
                <image :src="product.images[0] || 'https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'" class="product-image" mode="aspectFill"></image>
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

          <!-- 特惠推荐 -->
          <view class="processing-products">
            <view class="section-title processing-title">
              <view class="title-left">
                <text class="title-text">特惠推荐</text>
              </view>
              <view class="more">
                <text class="more-text">全部</text>
                <uni-icons type="right" size="14" color="#999"></uni-icons>
              </view>
            </view>
            <view class="processing-product-list">
              <view class="product-item" v-for="(product, index) in featuredProducts" :key="index" @click="goToProductDetail(product.id)">
                <image :src="product.images[0] || 'https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'" class="product-image" mode="aspectFill"></image>
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

    <ProductSelector ref="productSelector" />
  </view>
</template>

<script>
import { getProductDetail, searchProducts, searchProductSuggestions, getHotSearchKeywords } from '../../api/products';
import { getHotProducts, getSpecialProducts } from '../../api/index';
import { addItemToPurchaseList, updatePurchaseListTabBarBadge } from '../../utils/purchaseList';
import ProductSelector from '../../components/ProductSelector.vue';

export default {
  components: {
    ProductSelector
  },
  data() {
    return {
      statusBarHeight: 0, // 状态栏高度
      navBarHeight: 45, // 导航栏高度（默认值）
      windowWidth: 375, // 胶囊按钮左侧宽度（默认值，用于计算搜索框宽度）
      systemInfo: {}, // 系统信息
      searchText: '',
      suggestions: [], // 搜索建议列表
      searchTimer: null, // 搜索防抖定时器
      hotSearchTags: [],
      specialProducts: [],
      featuredProducts: [],
      searchHistory: [], // 搜索历史
      isInputFocused: false, // 输入框聚焦状态
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
    
    // 加载搜索历史
    this.loadSearchHistory();
    
    // 获取热门搜索关键词
    this.loadHotSearchKeywords();
    
    // 获取热销产品作为超值推荐
    this.loadHotProducts();
    
    // 获取精选商品作为特惠推荐
    this.loadFeaturedProducts();
  },
  onReady() {
    // 页面渲染完成后自动聚焦输入框
    this.$nextTick(() => {
      this.isInputFocused = true;
    });
  },
  onShow() {
    // 页面显示时重新加载搜索历史（从搜索结果页返回时）
    this.loadSearchHistory();
    // 重新聚焦输入框
    this.$nextTick(() => {
      this.isInputFocused = true;
    });
  },
  methods: {
    // 加载热门搜索关键词
    async loadHotSearchKeywords() {
      try {
        const res = await getHotSearchKeywords();
        if (res && res.code === 200 && res.data) {
          this.hotSearchTags = Array.isArray(res.data) ? res.data : [];
        }
      } catch (error) {
        console.error('获取热门搜索关键词失败:', error);
      }
    },
    
    // 加载热销产品作为超值推荐
    async loadHotProducts() {
      try {
        const res = await getHotProducts();
        if (res && res.code === 200 && res.data) {
          const products = Array.isArray(res.data) ? res.data : [];
          // 计算价格范围
          products.forEach(product => {
            this.calculateProductPriceRange(product);
          });
          // 前5个作为超值推荐
          this.specialProducts = products.slice(0, 5);
        }
      } catch (error) {
        console.error('获取热销产品失败:', error);
      }
    },
    
    // 加载精选商品作为特惠推荐
    async loadFeaturedProducts() {
      try {
        const res = await getSpecialProducts({ pageNum: 1, pageSize: 5 });
        if (res && res.code === 200 && res.data) {
          const products = Array.isArray(res.data) ? res.data : res.data.list || [];
          // 计算价格范围
          products.forEach(product => {
            this.calculateProductPriceRange(product);
          });
          this.featuredProducts = products.slice(0, 5);
        }
      } catch (error) {
        console.error('获取精选商品失败:', error);
      }
    },
    
    // 输入框输入事件（获取搜索建议）- 添加防抖处理
    onSearchInput() {
      // 清除之前的定时器
      if (this.searchTimer) {
        clearTimeout(this.searchTimer);
      }
      
      const keyword = this.searchText.trim();
      if (keyword) {
        // 防抖：延迟300ms后执行搜索建议请求
        this.searchTimer = setTimeout(() => {
          this.getSearchSuggestions(keyword);
        }, 300);
      } else {
        // 清空建议列表
        this.suggestions = [];
      }
    },
    
    // 清除搜索输入内容
    clearSearchText() {
      this.searchText = '';
      this.suggestions = [];
      // 清除防抖定时器
      if (this.searchTimer) {
        clearTimeout(this.searchTimer);
        this.searchTimer = null;
      }
      // 重新聚焦输入框
      this.$nextTick(() => {
        this.isInputFocused = true;
      });
    },
    
    // 获取搜索建议
    async getSearchSuggestions(keyword) {
      // 如果关键词为空，直接返回
      if (!keyword || !keyword.trim()) {
        this.suggestions = [];
        return;
      }
      
      try {
        const res = await searchProductSuggestions(keyword.trim(), 10);
        if (res && res.code === 200 && res.data) {
          // 确保返回的是数组
          this.suggestions = Array.isArray(res.data) ? res.data : [];
        } else {
          this.suggestions = [];
        }
      } catch (error) {
        console.error('获取搜索建议失败:', error);
        // 静默失败，不显示错误提示，避免影响用户体验
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
        // 添加到搜索历史
        this.addToSearchHistory(keyword);
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
    
    // 加载搜索历史
    loadSearchHistory() {
      try {
        const history = uni.getStorageSync('searchHistory');
        if (history && Array.isArray(history)) {
          this.searchHistory = history;
        } else {
          this.searchHistory = [];
        }
      } catch (error) {
        console.error('加载搜索历史失败:', error);
        this.searchHistory = [];
      }
    },
    
    // 添加到搜索历史
    addToSearchHistory(keyword) {
      if (!keyword || !keyword.trim()) {
        return;
      }
      const trimmedKeyword = keyword.trim();
      
      // 移除已存在的相同关键词
      this.searchHistory = this.searchHistory.filter(item => item !== trimmedKeyword);
      
      // 添加到开头
      this.searchHistory.unshift(trimmedKeyword);
      
      // 限制历史记录数量（最多保存10条）
      if (this.searchHistory.length > 10) {
        this.searchHistory = this.searchHistory.slice(0, 10);
      }
      
      // 保存到本地存储
      try {
        uni.setStorageSync('searchHistory', this.searchHistory);
      } catch (error) {
        console.error('保存搜索历史失败:', error);
      }
    },
    
    // 点击历史记录项
    selectHistoryItem(keyword) {
      this.searchText = keyword;
      this.performSearch();
    },
    
    // 清除搜索历史
    clearSearchHistory() {
      uni.showModal({
        title: '提示',
        content: '确定要清除所有搜索历史吗？',
        success: (res) => {
          if (res.confirm) {
            this.searchHistory = [];
            try {
              uni.removeStorageSync('searchHistory');
            } catch (error) {
              console.error('清除搜索历史失败:', error);
            }
          }
        }
      });
    },
    
    // 生成模拟搜索结果
    generateMockSearchResults(keyword) {
      const mockResults = [
        { id: 101, name: `精选${keyword}1`, price: '128.00', images: ['https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'] },
        { id: 102, name: `优质${keyword}2`, price: '98.00', images: ['https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'] },
        { id: 103, name: `新鲜${keyword}3`, price: '158.00', images: ['https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'] }
      ];
      mockResults.forEach(product => {
        this.calculateProductPriceRange(product);
      });
      return mockResults;
    },
    
    // 计算商品价格范围（与首页一致）
    calculateProductPriceRange(product) {
      if (!product.specs || !Array.isArray(product.specs) || product.specs.length === 0) {
        product.displayPrice = product.price || '0.00';
        return;
      }

      // 获取用户类型
      const userInfo = uni.getStorageSync('miniUserInfo');
      const isWholesaleUser = userInfo && userInfo.user_type === 'wholesale';
      
      // 收集价格
      const prices = [];
      product.specs.forEach(spec => {
        if (isWholesaleUser) {
          const wholesalePrice = spec.wholesale_price || spec.wholesalePrice;
          if (wholesalePrice && wholesalePrice > 0) {
            prices.push(parseFloat(wholesalePrice));
          }
        } else {
          const retailPrice = spec.retail_price || spec.retailPrice;
          if (retailPrice && retailPrice > 0) {
            prices.push(parseFloat(retailPrice));
          }
        }
      });

      // 如果没有找到对应类型的价格，使用另一种价格作为后备
      if (prices.length === 0) {
        product.specs.forEach(spec => {
          if (isWholesaleUser) {
            const retailPrice = spec.retail_price || spec.retailPrice;
            if (retailPrice && retailPrice > 0) {
              prices.push(parseFloat(retailPrice));
            }
          } else {
            const wholesalePrice = spec.wholesale_price || spec.wholesalePrice;
            if (wholesalePrice && wholesalePrice > 0) {
              prices.push(parseFloat(wholesalePrice));
            }
          }
        });
      }

      if (prices.length === 0) {
        product.displayPrice = product.price || '0.00';
        return;
      }

      const minPrice = Math.min(...prices);
      product.displayPrice = minPrice.toFixed(2);
    },
    
    // 跳转到商品详情
    goToProductDetail(productId) {
      uni.navigateTo({
        url: '/pages/product/detail?id=' + productId
      });
    },
    
    // 显示商品选择弹窗（使用ProductSelector组件，与首页一致）
    onAddBtnClick(product) {
      this.$refs.productSelector?.open(product);
    },
    
    // 原来的弹窗逻辑（保留但不再使用）
    async onAddBtnClickOld(product) {
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
    
    // 添加到采购单
    async addToCart() {
      if (!this.selectedSpec) {
        uni.showToast({
          title: '请选择商品规格',
          icon: 'none'
        });
        return;
      }
      const token = uni.getStorageSync('miniUserToken');
      if (!token) {
        uni.showToast({
          title: '请先登录',
          icon: 'none'
        });
        return;
      }

      try {
        await addItemToPurchaseList({
          token,
        productId: this.selectedProduct.id,
        specName: this.selectedSpec.name,
        quantity: this.quantity
        });
      uni.showToast({
        title: '已添加到采购单',
        icon: 'success'
      });
      updatePurchaseListTabBarBadge();
      this.closeProductModal();
      } catch (error) {
        console.error('添加采购单失败:', error);
        uni.showToast({
          title: '添加失败，请稍后再试',
          icon: 'none'
        });
      }
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
  padding-top: 0 !important;
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
  position: relative;
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
  padding-right: 40rpx;
}

.clear-btn {
  position: absolute;
  right: 20rpx;
  width: 32rpx;
  height: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  cursor: pointer;
  z-index: 10;
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

/* 搜索历史样式 */
.search-history {
  padding: 20rpx;
  border-radius: 20rpx;
  margin-bottom: 20rpx;
}

.clear-history {
  display: flex;
  align-items: center;
  gap: 8rpx;
  cursor: pointer;
}

.clear-text {
  font-size: 24rpx;
  color: #999;
}

.history-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
}

.history-item {
  cursor: pointer;
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
  font-size: 26rpx;
  font-weight: bold;
  padding-left: 10rpx;
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
  padding: 0;
  margin: 20rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.suggestions-list {
  display: flex;
  flex-direction: column;
}

.suggestion-item {
  display: flex;
  align-items: center;
  padding: 24rpx 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
  cursor: pointer;
  transition: background-color 0.2s;
  position: relative;
}

.suggestion-item:active {
  background-color: #f8f8f8;
}

.suggestion-item:last-child {
  border-bottom: none;
}

.suggestion-icon-wrapper {
  width: 48rpx;
  height: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f0f9f4;
  border-radius: 50%;
  margin-right: 20rpx;
  flex-shrink: 0;
}

.suggestion-icon {
  flex-shrink: 0;
}

.suggestion-text {
  flex: 1;
  font-size: 30rpx;
  color: #333;
  font-weight: 400;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.suggestion-arrow {
  margin-left: 15rpx;
  flex-shrink: 0;
  opacity: 0.5;
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
  border-radius: 16rpx;
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
  min-width: 0;
  overflow: hidden;
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
  word-break: break-all;
}

/* 超值推荐和特惠推荐中的商品名称固定宽度并允许换行 */
.special-product-list .product-name,
.processing-product-list .product-name {
  white-space: normal;
  word-wrap: break-word;
  word-break: break-all;
  display: block;
  -webkit-line-clamp: unset;
  -webkit-box-orient: unset;
  overflow: visible;
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