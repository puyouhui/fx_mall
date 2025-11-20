<template>
	<!-- 自定义头部 - 按照参考文章实现 -->
	<view class="custom-header">
		<!-- 固定导航栏 -->
		<view class="navbar-fixed" :style="{ backgroundColor: isHeaderBgVisible ? '#20CB6B' : 'transparent' }">
			<!-- 状态栏撑起高度 -->
			<view :style="{ height: statusBarHeight + 'px' }"></view>
			<!-- 导航栏内容区域 -->
			<view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
				<!-- 搜索框区域 -->
					<view class="navbar-search" @click="goToSearch">
						<view class="navbar-search_icon">
							<uni-icons type="search" size="16" color="#999"></uni-icons>
						</view>
						<view class="navbar-search_text">
							<input type="text" placeholder="请输入产品名称查询" placeholder-style="color: #999;" disabled/>
						</view>
					</view>
			</view>
		</view>
		<!-- 添加占位符高度，避免内容被遮挡 -->
		<!-- <view :style="{ height: statusBarHeight + navBarHeight + 'px' }"></view> -->
	</view>

	<view class="container">

		<!-- 轮播图 -->
		<view class="carousel">
			<swiper class="swiper" autoplay indicator-dots circular>
				<block v-for="(item, index) in carousels" :key="index">
					<swiper-item>
						<image :src="item.image" mode="aspectFill" class="carousel-image"
							@click="navigateTo(item.link)"></image>
					</swiper-item>
				</block>
			</swiper>
		</view>

		<view class="main-container">
			<!-- 特色标签 -->
			<view class="feature-tags">
				<view class="tag-item">
					<image src="/static/icon/coin.png" class="tag-icon"></image>
					<text class="tag-text">价格实惠</text>
				</view>
				<view class="tag-item">
					<image src="/static/icon/coin1.png" class="tag-icon"></image>
					<text class="tag-text">购一件送</text>
				</view>
				<view class="tag-item">
					<image src="/static/icon/tag-2.png" class="tag-icon"></image>
					<text class="tag-text">品类齐全</text>
				</view>
			</view>

			<!-- 分类 -->
			<view class="categories">
				<view class="category-item" v-for="(category, index) in categories" :key="index"
					@click="goToCategory(category.id)">
					<view class="category-icon-bg">
						<image v-if="category.icon" :src="category.icon" class="category-icon"></image>
						<image v-else src="/static/icon/nav_icon4.png" class="category-icon"></image>
					</view>
					<text class="category-name">{{ category.name }}</text>
				</view>
			</view>
		</view>

		<!-- 热销商品 -->
		<view class="hot-products">
			<view class="section-title">
				<view>
					<text class="hot-tag">HOT</text>
					<text class="section-name">热销</text>
				</view>
				<text class="more">更多 &gt;</text>
			</view>
			<scroll-view scroll-x class="hot-scroll">
				<view class="hot-product-item" v-for="(product, index) in hotProducts" :key="index"
					@click="goToProductDetail(product.id)">
					<!-- <image :src="product.images[0]" class="hot-product-image" mode="aspectFill"></image> -->
					<image
						src="http://113.44.164.151:9000/selected/carousel_1758513218.png?Content-Disposition=attachment%3B%20filename%3D%22carousel_1758513218.png%22&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=admin%2F20250922%2F%2Fs3%2Faws4_request&X-Amz-Date=20250922T044311Z&X-Amz-Expires=432000&X-Amz-SignedHeaders=host&X-Amz-Signature=2e4d2c6635b426786b69cb563d6784d7a2b5e9f2abef6a978112dc8210327f6e"
						class="hot-product-image" mode="aspectFill"></image>
					<view class="hot-product-price">
						<text class="price-symbol">¥</text>
						<text class="price-number">{{ product.displayPrice || product.price }}</text>
					</view>
				</view>
			</scroll-view>
		</view>

		<!-- 限时特价 -->
		<view class="special-offers">
			<view class="section-title">
				<text class="section-name">限时特价产品</text>
				<!-- <text class="time-info">剩余: 05:23:45</text> -->
			</view>
			<view class="special-product-list">
				<view class="special-product-item" v-for="(product, index) in specialProducts" :key="index"
					@click="goToProductDetail(product.id)">
					<image :src="product.images[0]" class="special-product-image" mode="aspectFill"></image>
					<view class="special-product-info">
						<text class="product-name">{{ product.name }}</text>
						<view class="price-and-btn">
							<view class="price-container">
								<text class="current-price">¥{{ product.displayPrice || '暂无价格' }}</text>
								<!-- <text class="original-price">{{ product.originalPrice }}</text> -->
							</view>
							<!-- <view class="buy-btn">
							<text>抢购</text>
						</view> -->
							<view class="add-btn" @click.stop="onAddBtnClick(product)">
								<uni-icons type="plusempty" size="20" color="#fff"></uni-icons>
							</view>
						</view>
					</view>
				</view>
			</view>
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
					<view class="spec-item" v-for="(spec, index) in selectedProduct.specs" :key="index"
						:class="{ 'selected': selectedSpec && selectedSpec.name === spec.name }"
						@click.stop="selectSpec(spec)">
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
</template>

<script>
import { getCarousels, getCategories, getSpecialProducts } from '../../api/index';
import { getProductDetail } from '../../api/products';
export default {
	data() {
		return {
			statusBarHeight: 20, // 状态栏高度（默认值）
			navBarHeight: 45, // 导航栏高度（默认值）
			windowWidth: 375, // 窗口宽度（默认值）
			isHeaderBgVisible: false,
			lastScrollTop: 0,
			carousels: [],
			categories: [],
			specialProducts: [],
			hotProducts: [],
			sections: [
				{
					title: '餐饮常用',
					products: []
				},
				{
					title: '酒店民宿',
					products: []
				},
				{
					title: '烧烤',
					products: []
				}
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
		this.loadCarousels();
		this.loadCategories();
		this.loadSpecialProducts();
		this.loadHotProducts();
		this.loadSectionProducts();

		// 获取设备信息
		const info = uni.getSystemInfoSync();
		// 设置状态栏高度
		this.statusBarHeight = info.statusBarHeight;
		this.windowWidth = info.windowWidth;

		// 获取胶囊按钮信息，实现精确对齐
		this.getMenuButtonInfo();
	},
	// 页面滚动事件
	onPageScroll(e) {
		const scrollTop = e.scrollTop;
		// 向上滑动且超过一定距离时显示背景色
		if (scrollTop > 50 && scrollTop > this.lastScrollTop) {
			this.isHeaderBgVisible = true;
		} else if (scrollTop < 50 || scrollTop < this.lastScrollTop - 20) {
			this.isHeaderBgVisible = false;
		}

		this.lastScrollTop = scrollTop;
	},
	methods: {
		// 跳转到搜索页面
		goToSearch() {
			uni.navigateTo({
				url: '/pages/search/search'
			});
		},
		
		// 获取胶囊按钮信息并计算导航栏高度
		getMenuButtonInfo() {
			try {
				// #ifndef H5 || APP-PLUS || MP-ALIPAY
				// 获取胶囊的位置信息
				const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
				// 按照参考文章的公式计算导航栏高度：
				// (胶囊底部高度 - 状态栏的高度) + (胶囊顶部高度 - 状态栏内的高度) = 导航栏的高度
				this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight);
				this.windowWidth = menuButtonInfo.left;
				// #endif
			} catch (error) {
				console.error('获取胶囊按钮信息失败:', error);
			}
		},


		// 加载轮播图
		async loadCarousels() {
			try {
				const res = await getCarousels();
				if (res.code === 200) {
					this.carousels = res.data;
				}
			} catch (error) {
				console.error('加载轮播图失败:', error);
			}
		},

		// 加载分类
		async loadCategories() {
			try {
				const res = await getCategories();
				if (res.code === 200) {
					this.categories = res.data;
				}
			} catch (error) {
				console.error('加载分类失败:', error);
			}
		},

		// 加载特价商品
		async loadSpecialProducts() {
			try {
				// 调用真实接口获取特价商品数据
				const res = await getSpecialProducts({ pageNum: 1, pageSize: 10 });
				console.log('特价商品接口返回:', res);

				// 处理接口返回的数据
				if (res.code === 200) {
					// 确保数据格式正确，兼容不同的返回结构
					this.specialProducts = Array.isArray(res.data) ? res.data : res.data.list || [];
					console.log('处理后的特价商品数据:', this.specialProducts);

					// 为每个商品处理数据并计算价格范围
					this.specialProducts.forEach(product => {
						// 处理数据结构差异，确保有必要的字段
						if (!product.images || !Array.isArray(product.images)) {
							product.images = product.image ? [product.image] : [];
						}

						// 确保价格字段正确
						if (product.price !== undefined) {
							product.price = parseFloat(product.price) || 0;
						}

						// 处理规格数据
						if (!product.specs && product.specifications && Array.isArray(product.specifications)) {
							if (product.specifications.length > 0 && product.specifications[0].price === undefined) {
								// 如果specifications没有价格信息，创建默认的specs结构
								product.specs = product.specifications.map((spec, index) => ({
									id: index + 1,
									name: spec.name,
									description: spec.value || '',
									price: parseFloat(product.price) || 0
								}));
							} else {
								// 直接使用specifications作为specs
								product.specs = product.specifications;
							}
						}

						// 确保所有规格都有id
						if (product.specs && Array.isArray(product.specs)) {
							product.specs.forEach((spec, index) => {
								if (spec.id === undefined) {
									spec.id = index + 1;
								}
								// 确保规格价格为数字
								if (spec.price !== undefined) {
									spec.price = parseFloat(spec.price) || 0;
								}
							});
						}

						// 计算价格范围
						this.calculateProductPriceRange(product);
					});

					// 如果没有获取到数据，显示空状态
					if (this.specialProducts.length === 0) {
						console.log('当前没有特价商品');
					}
				} else {
					console.error('特价商品接口返回错误:', res.message || '未知错误');
					this.specialProducts = [];
					// 可以考虑显示错误提示
					// uni.showToast({
					// 	title: res.message || '获取特价商品失败',
					// 	icon: 'none'
					// });
				}
			} catch (error) {
				console.error('加载特价商品时发生异常:', error);
				this.specialProducts = [];
				// 显示加载失败提示
				// uni.showToast({
				// 	title: '网络异常，请稍后重试',
				// 	icon: 'none'
				// });
			}
		},

		// 计算单个商品的价格范围
		calculateProductPriceRange(product) {
			if (!product.specs || !Array.isArray(product.specs) || product.specs.length === 0) {
				// 如果没有规格数据，使用原价格
				product.displayPrice = product.price || '0.00';
				return;
			}

			// 过滤出有价格的规格
			const pricedSpecs = product.specs.filter(spec => spec.price > 0);
			if (pricedSpecs.length === 0) {
				product.displayPrice = product.price || '0.00';
				return;
			}

			// 计算最小和最大价格
			const minPrice = Math.min(...pricedSpecs.map(spec => spec.price));
			const maxPrice = Math.max(...pricedSpecs.map(spec => spec.price));

			// 设置价格范围显示
			if (minPrice === maxPrice) {
				product.displayPrice = minPrice.toFixed(2);
			} else {
				product.displayPrice = minPrice.toFixed(2) + '~' + maxPrice.toFixed(2);
			}
		},

		// 加载热销商品
		loadHotProducts() {
			// 模拟数据
			this.hotProducts = [
				{ id: 1, images: ['/static/test/product1.jpg'], price: '98.99' },
				{ id: 2, images: ['/static/test/product2.jpg'], price: '98.99' },
				{ id: 3, images: ['/static/test/product3.jpg'], price: '98.99' },
				{ id: 4, images: ['/static/test/product4.jpg'], price: '98.99' }
			];
			// 为模拟数据计算价格范围
			this.hotProducts.forEach(product => {
				this.calculateProductPriceRange(product);
			});
		},

		// 加载各分类区块商品
		loadSectionProducts() {
			// 模拟数据
			this.sections.forEach(section => {
				section.products = [
					{ id: 5, images: ['/static/test/product5.jpg'], price: '128~298' },
					{ id: 6, images: ['/static/test/product6.jpg'], price: '128~298' }
				];
			});
		},

		// 导航到指定链接
		navigateTo(link) {
			if (link.startsWith('product/')) {
				const productId = link.split('/')[1];
				uni.navigateTo({
					url: '/pages/product/detail?id=' + productId
				});
			} else if (link.startsWith('category/')) {
				const categoryId = link.split('/')[1];
				uni.navigateTo({
					url: '/pages/category/category?id=' + categoryId
				});
			}
		},

		// 跳转到分类页面
		goToCategory(categoryId) {
			// 使用globalData传递分类ID
			getApp().globalData.targetCategoryId = categoryId;
			// 跳转到分类页面
			uni.switchTab({
				url: '/pages/category/category'
			});
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
/* 自定义头部样式 - 按照参考文章实现 */
.custom-header {
	/* 头部容器样式 */
}

.navbar-fixed {
	position: fixed;
	top: 0;
	left: 0;
	z-index: 999;
	width: 100%;
	transition: background-color 0.3s ease;
}

.navbar-content {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0 30rpx;
	box-sizing: border-box;
}

.navbar-search {
	display: flex;
	align-items: center;
	padding: 0 10px;
	width: 65%;
	height: 30px;
	border-radius: 30px;
	background-color: #f5f5f5;
}

.navbar-search_icon {
	margin-right: 10px;
}

.navbar-search_text {
	width: 100%;
}

.navbar-search_text input {
	width: 100%;
	font-size: 28rpx;
	color: #999;
	background-color: transparent;
	border: none;
	outline: none;
}

.container {
	background-color: #f5f5f5;
	height: 100vh;
	padding-bottom: 100rpx;
	/* 为底部tabbar留出空间 */
}

.main-container {
	width: 96%;
	margin: 0 auto;
	margin-top: -60px;
	z-index: 9;
	display: block;
	position: relative;
}

/* 轮播图样式 */
.carousel {
	width: 100%;
	height: 360px;
	background-color: #fff;
	z-index: 0;
}

.swiper {
	width: 100%;
	height: 360px;
}

.carousel-image {
	width: 100%;
	height: 360px;
}

/* 特色标签 */
.feature-tags {
	display: flex;
	padding: 10rpx 20rpx;
	background-color: #F3FBF7;
	justify-content: space-around;
	border-radius: 20rpx 20rpx 0 0;
}

.tag-item {
	display: flex;
	align-items: center;
	justify-content: center;
}

.tag-icon {
	width: 28rpx;
	height: 28rpx;
}

.tag-text {
	font-size: 24rpx;
	color: rgba(64, 71, 92, 0.4);
	padding-left: 10rpx;
}

/* 分类样式 */
.categories {
	display: flex;
	flex-wrap: wrap;
	padding: 20rpx 20rpx 0 20rpx;
	background-color: #fff;
	border-radius: 0 0 20rpx 20rpx;
}

.category-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	width: 25%;
	margin-bottom: 20rpx;
}


.category-icon-bg {
	width: 100rpx;
	height: 100rpx;
}

.category-icon {
	width: 100rpx;
	height: 100rpx;
}

.category-name {
	font-size: 26rpx;
	color: #40475C;
}

/* 热销商品样式 */
.hot-products {
	width: 96%;
	background-color: #fff;
	margin-top: 20rpx;
	padding: 20rpx;
	box-sizing: border-box;
	border-radius: 20rpx;
	margin: 20rpx auto 0 auto;
}

.section-title {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 20rpx;
}

.section-left {
	display: flex;
	align-items: center;
}

.hot-tag {
	background-color: #ff4d4f;
	color: #fff;
	font-size: 20rpx;
	padding: 2rpx 8rpx;
	border-radius: 4rpx;
	margin-right: 10rpx;
}

.section-name {
	font-size: 32rpx;
	font-weight: bold;
	color: #333;
	margin-right: 10rpx;
}

.section-subtitle {
	font-size: 24rpx;
	color: #999;
}

.more {
	font-size: 26rpx;
	color: #999;
}

.hot-scroll {
	width: 100%;
	height: 300rpx;
	overflow: hidden;
	white-space: nowrap
}

.hot-scroll::-webkit-scrollbar {
	display: none;
}

.hot-product-item {
	width: 200rpx;
	margin-right: 20rpx;
	display: inline-block;
	transition: transform 0.2s;
}

.hot-product-image {
	width: 200rpx;
	height: 200rpx;
	border-radius: 10rpx;
}

.hot-product-price {
	display: flex;
	align-items: baseline;
	margin-top: 10rpx;
}

.price-symbol {
	font-size: 24rpx;
	color: #ff4d4f;
}

.price-number {
	font-size: 28rpx;
	color: #ff4d4f;
	font-weight: bold;
}

/* 商品分类区块样式 */
.category-sections {
	margin-top: 20rpx;
}

.section {
	background-color: #fff;
	margin-bottom: 20rpx;
	padding: 20rpx;
}

.section-products {
	display: flex;
	justify-content: space-between;
}

.product-item {
	width: 48%;
	display: flex;
	flex-direction: column;
}

.product-image {
	width: 100%;
	height: 240rpx;
	border-radius: 10rpx;
}

.product-price {
	font-size: 28rpx;
	color: #ff4d4f;
	margin-top: 10rpx;
	font-weight: bold;
}

/* 限时特价样式 */
.special-offers {
	background-color: #fff;
	margin-top: 20rpx;
	padding: 20rpx;
}

.time-info {
	font-size: 24rpx;
	color: #999;
}

.special-product-list {
	display: flex;
	flex-wrap: wrap;
	justify-content: space-between;
	margin-top: 20rpx;
}

.special-product-item {
	width: 48%;
	background-color: #fff;
	border-radius: 12rpx;
	overflow: hidden;
	margin-bottom: 20rpx;
	box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.05);
	display: flex;
	flex-direction: column;
}

.special-product-image {
	width: 100%;
	height: 300rpx;
	border-top-left-radius: 12rpx;
	border-top-right-radius: 12rpx;
}

.special-product-info {
	padding: 20rpx;
	display: flex;
	flex-direction: column;
	position: relative;
	flex: 1;
}

.product-name {
	height: 44px;
	font-size: 28rpx;
	color: #333;
	line-height: 40rpx;
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
	margin-bottom: 15rpx;
	font-weight: bold;
}

.price-and-btn {
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.price-container {
	display: flex;
	align-items: baseline;
}

.current-price {
	font-size: 32rpx;
	color: #ff4d4f;
	font-weight: bold;
}

.original-price {
	font-size: 24rpx;
	color: #999;
	text-decoration: line-through;
	margin-left: 10rpx;
}

.buy-btn {
	align-self: flex-end;
	background-color: #19B95F;
	color: #fff;
	padding: 10rpx 30rpx;
	border-radius: 20rpx;
	font-size: 24rpx;
}

.add-btn {
	width: 30px;
	height: 30px;
	background-color: #20CB6B;
	color: #fff;
	font-size: 36rpx;
	border-radius: 30rpx;
	font-weight: bold;
	display: flex;
	justify-content: center;
	align-items: center;
	box-sizing: border-box;
}

.product-modal {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	z-index: 999;
	display: flex;
	justify-content: flex-end;
	align-items: flex-end;
}

.modal-overlay {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background-color: rgba(0, 0, 0, 0.5);
}

.modal-content {
	width: 100%;
	max-height: 70vh;
	background-color: #fff;
	border-radius: 30rpx 30rpx 0 0;
	position: relative;
	z-index: 1;
	overflow-y: auto;
}

.modal-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20rpx 30rpx;
	border-bottom: 1rpx solid #eee;
}

.modal-title {
	font-size: 32rpx;
	font-weight: bold;
	color: #333;
}

.modal-close {
	padding: 10rpx;
}

.product-modal-info {
	display: flex;
	padding: 20rpx;
	border-bottom: 1rpx solid #f0f0f0;
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
	font-size: 32rpx;
	color: #333;
	line-height: 48rpx;
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
}

.modal-product-price {
	font-size: 36rpx;
	color: #ff4d4f;
	font-weight: bold;
	margin-top: 20rpx;
}

.specs-section {
	padding: 20rpx;
}

.specs-title {
	font-size: 28rpx;
	color: #333;
	margin-bottom: 20rpx;
	display: block;
}

.specs-list {
	display: flex;
	flex-wrap: wrap;
}

.spec-item {
	padding: 15rpx 30rpx;
	border: 1rpx solid #ddd;
	border-radius: 50rpx;
	margin-right: 20rpx;
	margin-bottom: 20rpx;
	display: flex;
	align-items: center;
	font-size: 28rpx;
	color: #333;
}

.spec-item.selected {
	border-color: #20CB6B;
	color: #20CB6B;
	background-color: #f0fff4;
}

.spec-description {
	margin: 0 10rpx;
	color: #666;
}

.spec-price {
	color: #ff4d4f;
	margin-left: 10rpx;
	font-weight: bold;
}

.quantity-section {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20rpx;
	border-top: 1rpx solid #f0f0f0;
	border-bottom: 1rpx solid #f0f0f0;
}

.quantity-title {
	font-size: 28rpx;
	color: #333;
}

.quantity-selector {
	display: flex;
	align-items: center;
}

.minus-btn,
.plus-btn {
	width: 60rpx;
	height: 60rpx;

	border-radius: 50%;
	display: flex;
	justify-content: center;
	align-items: center;
}

.minus-btn {
	margin-right: 30rpx;
	border: 1rpx solid #ddd;
}

.plus-btn {
	margin-left: 30rpx;
	background-color: #20CB6B !important;
}

.quantity-text {
	font-size: 32rpx;
	color: #333;
	font-weight: bold;
}

.modal-bottom {
	padding: 20rpx;
}

.buy-btn {
	background-color: #20CB6B !important;
	color: #fff;
	font-size: 32rpx;
	font-weight: bold;
	text-align: center;
	padding: 25rpx 0;
	border-radius: 80rpx;
}

.minus-btn-icon {
	width: 16px;
	height: 16px;
}
</style>
