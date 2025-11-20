<template>
	<!-- 自定义头部 - 按照参考文章实现 -->
	<view class="custom-header">
		<!-- 固定导航栏 -->
		<view class="navbar-fixed">
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
	</view>

	<!-- 分类容器 - 保持固定高度 -->
	<view class="categories-container" :style="{ top: statusBarHeight + navBarHeight + 'px' }">
		<!-- 一级分类左右滑动选择 - 仅在分类未展开时显示 -->
		<view class="primary-categories" v-if="!isCategoriesExpanded">
			<scroll-view scroll-x enable-flex class="primary-scroll">
				<view class="primary-category-item" v-for="category in primaryCategories" :key="category.id"
					:class="{ active: selectedPrimaryCategoryId === category.id }"
					@click="selectPrimaryCategory(category.id)">
					<image :src="category.icon" class="primary-category-icon"></image>
					<text class="primary-category-name">{{ category.name }}</text>
				</view>
				<view style="width: 40px;height: 40px;">占位</view>
				<!-- <view class="primary-category-item"></view> -->
			</scroll-view>
		</view>

		<!-- 展开/收起分类按钮 - 始终显示 -->
		<view class="expand-categories-btn" @click="toggleCategoriesExpand">
			<text>展</text>
			<text style="margin-top: 5rpx;">开</text>
			<uni-icons :type="isCategoriesExpanded ? 'down' : 'down'" size="14" color="#20CB6B" class="expand-icon"
				style="margin-left: -2rpx;"></uni-icons>
		</view>
	</view>

	<!-- 展开的全部分类弹窗 - 使用fixed定位 -->
	<view class="all-categories-popup" :style="{ top: statusBarHeight + navBarHeight + 'px' }"
		v-if="isCategoriesExpanded">
		<!-- <view class="all-categories-popup" v-if="true"> -->
		<view class="all-categories">
			<view class="all-categories-header">
				<text class="all-categories-title">全部分类</text>
			</view>
			<scroll-view scroll-y class="all-categories-scroll">
				<view class="category-grid-container">
					<view class="category-grid-item" v-for="category in primaryCategories" :key="category.id"
						:class="{ active: selectedPrimaryCategoryId === category.id }"
						@click="selectPrimaryCategory(category.id)">
						<image :src="category.icon" class="category-grid-icon"></image>
						<text class="category-grid-name">{{ category.name }}</text>
					</view>
				</view>
			</scroll-view>
			<!-- 收齐图标 -->
			<view class="collapse-icon" @click="toggleCategoriesExpand">
				收起
				<uni-icons type="up" size="14" color="#999" style="margin-left: 10rpx;"></uni-icons>

			</view>
		</view>
	</view>

	<!-- 主体区域 - 通过计算高度设置 -->
	<view class="container"
		:style="{ height: containerHeight + 'px', top: statusBarHeight + navBarHeight + 90 + 'px' }">
		<!-- 调试信息 (临时添加) -->
		<!-- <view class="debug-info" v-if="debugInfo.show">
			<text class="debug-title">调试信息</text>
			<text class="debug-text">当前分类ID: {{ debugInfo.categoryId }}</text>
			<text class="debug-text">API调用结果: {{ debugInfo.apiResult }}</text>
		</view> -->

		<!-- 主体内容 -->
		<view class="main-content">
			<!-- 二级分类列表 - 左侧垂直滚动 -->
			<view class="secondary-categories">
				<scroll-view scroll-y enable-flex class="secondary-scroll">
					<view class="secondary-category-item" v-for="category in secondaryCategories" :key="category.id"
						:class="{ active: selectedSecondaryCategoryId === category.id }"
						@click="selectSecondaryCategory(category.id)">
						{{ category.name }}
					</view>
				</scroll-view>
			</view>

			<!-- 商品列表 - 右侧两列布局 -->
			<view class="products-container">
				<view class="category-title">{{ currentCategoryName }}</view>
				<view class="product-list" v-if="currentProducts.length > 0">
					<view class="product-item" v-for="product in currentProducts" :key="product.id" @click="goToProductDetail(product.id)">
						<image :src="product.images[0]" class="product-image"></image>
						<view class="product-content">
							<view class="product-name">{{ product.name }}</view>
							<view class="product-desc">{{ product.description || '极速送达，品质稳定' }}</view>
							<view class="product-bottom">
								<view class="price-container">
									<view class="product-price"><text
											style="font-size: 24rpx;margin-right: 6rpx;">¥</text>{{ product.price_range
											}}</view>
									<view class="original-price" v-if="product.originalPrice">¥{{ product.originalPrice
									}}
									</view>
								</view>
								<view class="add-cart-btn" @click.stop="addToCart(product)">
									<uni-icons type="plusempty" size="18" color="#fff"></uni-icons>
								</view>
							</view>
						</view>
					</view>
				</view>
				<view class="empty-tip" v-else>
					暂无商品
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
			<view class="modal-product-info">
				<image :src="selectedProduct?.images[0] || ''" class="modal-product-image" mode="aspectFill"></image>
				<view class="modal-product-details">
					<view class="modal-product-name">{{ selectedProduct.name }}</view>
					<view class="modal-product-price">¥{{ selectedProduct.displayPrice }}</view>
				</view>
			</view>

			<!-- 规格选择 -->
			<view class="modal-specs-section" v-if="selectedProduct.specs && selectedProduct.specs.length > 0">
				<view class="modal-section-title">选择规格</view>
				<view class="modal-specs-list">
					<view class="modal-spec-item" :class="{ active: selectedSpec && selectedSpec.id === spec.id }"
						v-for="spec in selectedProduct.specs" :key="spec.id" @click.stop="selectSpec(spec)">
						<text>{{ spec.name }}</text>
						<text class="modal-spec-description" v-if="spec.description">({{ spec.description }})</text>
						<text class="modal-spec-price" v-if="spec.price && spec.price !== selectedProduct.price">¥{{ spec.price }}</text>
					</view>
				</view>
			</view>

			<!-- 数量选择 -->
			<view class="modal-quantity-section">
				<view class="modal-section-title">数量</view>
				<view class="modal-quantity-control">
					<view class="modal-decrease-btn" @click.stop="decreaseQuantity" :class="{ disabled: quantity <= 1 }">
						<image src="/static/icon/minus.png" class="minus-btn-icon"></image>
					</view>
					<view class="modal-quantity">{{ quantity }}</view>
					<view class="modal-increase-btn" @click.stop="increaseQuantity">
						<uni-icons type="plusempty" size="18" color="#fff"></uni-icons>
					</view>
				</view>
			</view>

			<!-- 底部按钮 -->
			<view class="modal-bottom">
				<view class="modal-add-to-cart-btn" @click.stop="addToPurchaseCart">
					添加到采购单
				</view>
			</view>
		</view>
	</view>
</template>

<script>
import { log } from 'console';
import { getCategories } from '../../api/index.js';
import { getProductsByCategory } from '../../api/products.js';
import { getProductDetail } from '../../api/products.js';

export default {
	data() {
		return {
			// 自定义头部相关
			statusBarHeight: 0,
			navBarHeight: 44,
			isHeaderBgVisible: true,

			// 分类相关
			primaryCategories: [], // 一级分类
			secondaryCategories: [], // 二级分类
			selectedPrimaryCategoryId: 0,
			selectedSecondaryCategoryId: 0,
			currentCategoryName: '',
			isCategoriesExpanded: false,

			// 商品相关
			currentProducts: [],

			// 布局相关
			categoriesContainerHeight: 100, // 分类容器固定高度
			containerHeight: 0, // 主体区域高度（通过计算设置）
			screenHeight: 0, // 屏幕高度
			tabBarHeight: 100, // 底部tabbar高度（rpx单位）

			// 调试信息
			debugInfo: {
				show: true,
				categoryId: '',
				apiResult: ''
			},

			// 商品选择弹窗相关状态
			showProductModal: false,
			selectedProduct: null,
			selectedSpec: null,
			quantity: 1,
			loadingProduct: false
		};
	},

	// 生命周期函数 - 监听页面加载
	onLoad(options) {
		// 获取设备信息，设置状态栏高度
		const systemInfo = uni.getSystemInfoSync();
		this.statusBarHeight = systemInfo.statusBarHeight;
		this.screenHeight = systemInfo.windowHeight;

		// 计算布局高度
		this.calculateLayoutHeight();

		// 加载分类数据
		this.loadPrimaryCategories();

		// 如果有传入分类ID，则选中对应分类
		if (options && options.id) {
			this.selectedPrimaryCategoryId = parseInt(options.id);
			// 只加载二级分类，商品加载会在loadSecondaryCategories方法中自动处理
			this.loadSecondaryCategories(this.selectedPrimaryCategoryId);
		}
	},

	// 生命周期函数 - 监听页面显示
	onShow() {
		// 只在分类数据为空时才重新加载
		if (this.primaryCategories.length === 0) {
			this.loadPrimaryCategories();
		}
		
		// 从globalData中获取目标分类ID
		const app = getApp();
		const targetCategoryId = app.globalData.targetCategoryId;
		
		// 如果有目标分类ID，则选中对应分类
		if (targetCategoryId) {
			const categoryId = parseInt(targetCategoryId);
			if (categoryId !== this.selectedPrimaryCategoryId) {
				this.selectedPrimaryCategoryId = categoryId;
				// 加载对应二级分类和商品
				this.loadSecondaryCategories(this.selectedPrimaryCategoryId);
			}
			// 清空globalData中的目标分类ID，避免下次打开时继续使用
			app.globalData.targetCategoryId = null;
		}
	},

	// 生命周期函数 - 监听页面尺寸变化
	onResize() {
		// 重新计算布局高度
		this.calculateLayoutHeight();
	},

	methods: {
		// 跳转到搜索页面
		goToSearch() {
			uni.navigateTo({
				url: '/pages/search/search'
			});
		},
		
		// 加载一级分类
		async loadPrimaryCategories() {
			try {
				const res = await getCategories();
				// 过滤出一级分类（parent_id为0或不存在）
				this.primaryCategories = res.data.filter(category =>
					category.parent_id === 0 || category.parentId === 0
				).map(category => ({
					...category,
					parentId: category.parent_id || 0, // 统一键名
					icon: category.icon || '/static/images/default-category.png' // 设置默认图标
				}));

				// 设置默认选中第一个分类
				if (this.primaryCategories.length > 0 && this.selectedPrimaryCategoryId === 0) {
					this.selectPrimaryCategory(this.primaryCategories[0].id);
				}

				// 构建分类网格数据
				this.buildCategoriesGrid();
			} catch (error) {
				console.error('加载分类失败:', error);
				// 使用模拟数据
				this.useMockCategories();
			}
		},

		// 使用模拟分类数据
		useMockCategories() {
			this.primaryCategories = [
				{ id: 1, name: '热门推荐', parentId: 0, icon: '/static/images/category1.png' },
				{ id: 2, name: '新鲜水果', parentId: 0, icon: '/static/images/category2.png' },
				{ id: 3, name: '休闲零食', parentId: 0, icon: '/static/images/category3.png' },
				{ id: 4, name: '生鲜蔬菜', parentId: 0, icon: '/static/images/category4.png' }
			];

			if (this.primaryCategories.length > 0 && this.selectedPrimaryCategoryId === 0) {
				this.selectPrimaryCategory(this.primaryCategories[0].id);
			}

			this.buildCategoriesGrid();
		},

		// 构建分类网格 - 现在直接使用primaryCategories数组
		buildCategoriesGrid() {
			// 不再进行分组，直接使用原始数组
			this.allCategoriesGrid = [];
		},

		// 加载二级分类
		async loadSecondaryCategories(primaryCategoryId) {
			try {
				const res = await getCategories();
				const allCategories = res.data;

				// 查找当前一级分类
				const primaryCategory = allCategories.find(cat => cat.id === primaryCategoryId);

				// 如果找到了一级分类且有子分类
				if (primaryCategory && primaryCategory.children && primaryCategory.children.length > 0) {
					// 过滤出启用的子分类
					this.secondaryCategories = primaryCategory.children
						.filter(child => child.status === 1 || child.status === undefined)
						.map(child => ({
							...child,
							parentId: child.parent_id || primaryCategoryId // 统一键名
						}));
				} else {
					// 否则使用模拟数据
					this.secondaryCategories = this.generateMockSecondaryCategories(primaryCategoryId);
				}
			} catch (error) {
				console.error('加载二级分类失败:', error);
				// 使用模拟数据
				this.secondaryCategories = this.generateMockSecondaryCategories(primaryCategoryId);
			}

			// 确保至少有一个二级分类
			if (this.secondaryCategories.length === 0) {
				// 如果没有二级分类，创建一个默认的分类项
				this.secondaryCategories = [{
					id: primaryCategoryId * 10 + 1,
					name: '全部',
					parentId: primaryCategoryId
				}];
			}

			// 自动选择第一个二级分类
			this.selectedSecondaryCategoryId = this.secondaryCategories[0].id;
			// 使用二级分类ID加载对应商品
			this.loadProductsByCategory(this.selectedSecondaryCategoryId);
		},

		// 显示商品选择弹窗
		async addToCart(product) {
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

		// 添加到采购单
		addToPurchaseCart() {
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

		// 生成模拟二级分类数据
		generateMockSecondaryCategories(primaryCategoryId) {
			const primaryCategory = this.primaryCategories.find(cat => cat.id === primaryCategoryId);
			const categoryName = primaryCategory ? primaryCategory.name : '热门';

			// 根据一级分类ID生成不同的二级分类名称
			if (primaryCategoryId === 4 || categoryName.includes('蔬菜')) {
				// 叶菜类应该显示具体的蔬菜品种
				if (this.selectedPrimaryCategoryId === 4 && this.selectedSecondaryCategoryId === 0) {
					// 默认显示叶菜类的具体品种
					return [
						{ id: primaryCategoryId * 100 + 1, name: '生菜', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 2, name: '菜心', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 3, name: '芹菜', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 4, name: '油麦菜', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 5, name: '大白菜', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 6, name: '娃娃菜', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 7, name: '芥兰', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 8, name: '枸杞叶', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 9, name: '西洋菜', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 10, name: '香菜', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 11, name: '紫苏', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 100 + 12, name: '韭菜', parentId: primaryCategoryId }
					];
				} else {
					return [
						{ id: primaryCategoryId * 10 + 1, name: '叶菜类', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 10 + 2, name: '根茎类', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 10 + 3, name: '豆芽豆类', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 10 + 4, name: '茄瓜类', parentId: primaryCategoryId },
						{ id: primaryCategoryId * 10 + 5, name: '葱姜蒜', parentId: primaryCategoryId }
					];
				}
			} else {
				return [
					{ id: primaryCategoryId * 10 + 1, name: categoryName + '精选', parentId: primaryCategoryId },
					{ id: primaryCategoryId * 10 + 2, name: '新品上架', parentId: primaryCategoryId },
					{ id: primaryCategoryId * 10 + 3, name: '特价优惠', parentId: primaryCategoryId },
					{ id: primaryCategoryId * 10 + 4, name: '销量排行', parentId: primaryCategoryId },
					{ id: primaryCategoryId * 10 + 5, name: '新品尝鲜', parentId: primaryCategoryId }
				];
			}
		},

		// 加载分类商品
		async loadProductsByCategory(categoryId) {
			try {
				console.log('加载分类商品，分类ID:', categoryId);
				// 更新调试信息
				this.debugInfo.categoryId = categoryId;
				this.debugInfo.apiResult = '加载中...';

				// 显示加载状态
				uni.showLoading({
					title: '加载中...'
				});

				// 正确传递参数对象，包含categoryId、pageNum和pageSize
				const res = await getProductsByCategory({ categoryId, pageNum: 1, pageSize: 10 });
				console.log('获取分类商品结果:', res);
				// 更新调试信息
				this.debugInfo.apiResult = JSON.stringify(res, null, 2);
				// 从响应的data.list中获取商品列表
				this.currentProducts = res.data && res.data.list ? res.data.list : [];

				// 如果没有商品数据，清空列表以显示空状态提示
				if (!this.currentProducts || this.currentProducts.length === 0) {
					console.log('没有找到商品数据，显示空状态');
					this.currentProducts = [];
				}
				console.log('最终显示的商品数据:', this.currentProducts);
			} catch (error) {
				console.error('加载商品失败:', error);
				// 出错时也显示空状态
				this.currentProducts = [];
			} finally {
				// 隐藏加载状态
				uni.hideLoading();
			}
		},

		// 生成模拟商品数据
		generateMockProducts(categoryId) {
			const primaryCategory = this.primaryCategories.find(cat => cat.id === categoryId);
			const categoryName = primaryCategory ? primaryCategory.name : '热门';

			// 为蔬菜分类提供更真实的模拟数据
			if (categoryId === 4 || categoryName.includes('蔬菜')) {
				return [
					{
						id: 1,
						name: '优质红苋菜2斤',
						description: '新鲜',
						price: 8.8,
						originalPrice: 10.8,
						image: '/static/test/vegetable1.jpg'
					},
					{
						id: 2,
						name: '普通空心菜',
						description: '绿色健康 新鲜蔬菜',
						price: 18.8,
						originalPrice: 24.8,
						image: '/static/test/vegetable2.jpg'
					},
					{
						id: 3,
						name: '普通生菜5斤',
						description: '新鲜 2.5KG',
						price: 13.8,
						originalPrice: 16.8,
						image: '/static/test/vegetable3.jpg'
					},
					{
						id: 4,
						name: '广东菜心2斤',
						description: '新鲜采摘',
						price: 12.8,
						originalPrice: 15.8,
						image: '/static/test/vegetable4.jpg'
					}
				];
			} else {
				// 通用模拟数据
				return Array(4).fill().map((_, index) => ({
					id: index + 1,
					name: `${categoryName}商品${index + 1}`,
					description: '优质商品，品质保证',
					price: Math.floor(Math.random() * 50) + 10,
					originalPrice: Math.floor(Math.random() * 30) + 60,
					image: `/static/test/product${index + 1}.jpg`
				}));
			}
		},

		// 选择一级分类
		selectPrimaryCategory(categoryId) {
			this.selectedPrimaryCategoryId = categoryId;
			this.selectedSecondaryCategoryId = 0; // 重置二级分类选中状态

			// 获取当前分类名称
			const selectedCategory = this.primaryCategories.find(cat => cat.id === categoryId);
			this.currentCategoryName = selectedCategory ? selectedCategory.name : '';

			// 加载二级分类（商品加载会在loadSecondaryCategories方法中自动处理）
			this.loadSecondaryCategories(categoryId);

			// 收起分类弹窗
			this.isCategoriesExpanded = false;

			// 滚动到顶部
			uni.pageScrollTo({
				scrollTop: 0,
				duration: 300
			});
		},

		// 选择二级分类
		selectSecondaryCategory(categoryId) {
			console.log(categoryId);

			this.selectedSecondaryCategoryId = categoryId;
			// 使用二级分类ID加载对应商品
			this.loadProductsByCategory(categoryId);
		},

		// 切换分类展开/收起
		toggleCategoriesExpand() {
			this.isCategoriesExpanded = !this.isCategoriesExpanded;
		},

		// 计算布局高度
		calculateLayoutHeight() {
			// 头部总高度（状态栏 + 导航栏）
			const headerTotalHeight = this.statusBarHeight + this.navBarHeight;

			// 转换tabBar高度（从rpx到px）
			const systemInfo = uni.getSystemInfoSync();
			const tabBarHeightInPx = (this.tabBarHeight * systemInfo.windowWidth) / 750;

			// 主体区域高度 = 屏幕高度 - 头部高度 - 分类容器高度(90px) - tabBar高度
			this.containerHeight = this.screenHeight - headerTotalHeight - tabBarHeightInPx;
		},

		// 跳转到商品详情页
		goToProductDetail(productId) {
			uni.navigateTo({
				url: `/pages/product/detail?id=${productId}`
			});
		}
	}
}
</script>

<style>
page {
	background-color: #EEF7F4;
}

/* 商品选择弹窗样式 */
.product-modal {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	display: flex;
	flex-direction: column;
	justify-content: flex-end;
	z-index: 9999;
}

.product-modal-content {
	background: #fff;
	border-top-left-radius: 20rpx;
	border-top-right-radius: 20rpx;
	padding: 20rpx;
	max-height: 70vh;
	overflow-y: auto;
	box-sizing: border-box;
}

.product-info {
	display: flex;
	padding: 20rpx 0;
	border-bottom: 1px solid #f0f0f0;
}

.product-image {
	width: 200rpx;
	height: 200rpx;
	background: #f5f5f5;
	border-radius: 10rpx;
}

.product-details {
	flex: 1;
	margin-left: 20rpx;
	display: flex;
	flex-direction: column;
	justify-content: space-between;
}

.product-name {
	font-size: 28rpx;
	color: #333;
	line-height: 40rpx;
}

.product-price {
	font-size: 32rpx;
	color: #ff6b35;
	font-weight: bold;
}

.product-desc {
	font-size: 24rpx;
	color: #999;
	line-height: 32rpx;
}

.close-btn {
	font-size: 48rpx;
	color: #999;
	padding: 0 10rpx;
}

.specs-section,
.quantity-section {
	padding: 20rpx 0;
	border-bottom: 1px solid #f0f0f0;
}

.section-title {
	font-size: 28rpx;
	color: #333;
	margin-bottom: 20rpx;
}

.specs-list {
	display: flex;
	flex-wrap: wrap;
}

.spec-item {
	padding: 15rpx 25rpx;
	border: 1px solid #ddd;
	border-radius: 10rpx;
	margin-right: 20rpx;
	margin-bottom: 20rpx;
	font-size: 26rpx;
	color: #666;
}

.spec-item.active {
	border-color: #ff6b35;
	color: #ff6b35;
}

.quantity-control {
	display: flex;
	align-items: center;
	width: 200rpx;
}

.decrease-btn,
.increase-btn {
	width: 60rpx;
	height: 60rpx;
	border: 1px solid #ddd;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 32rpx;
	color: #666;
}

.decrease-btn.disabled {
	color: #ccc;
}

.quantity {
	width: 80rpx;
	height: 60rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 28rpx;
}

.modal-footer {
	background: #fff;
	padding: 20rpx;
	border-top: 1px solid #f0f0f0;
}

.add-to-cart-btn {
	background: #ff6b35;
	color: #fff;
	text-align: center;
	padding: 25rpx;
	border-radius: 10rpx;
	font-size: 32rpx;
	font-weight: bold;
}

/* 商品选择弹窗样式 */
.product-modal {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	display: flex;
	flex-direction: column;
	justify-content: flex-end;
	z-index: 9999;
}

.product-modal-content {
	background: #fff;
	border-top-left-radius: 20rpx;
	border-top-right-radius: 20rpx;
	padding: 20rpx;
	max-height: 70vh;
	overflow-y: auto;
}

.product-info {
	display: flex;
	padding: 20rpx 0;
	border-bottom: 1px solid #f0f0f0;
}

.product-image {
	width: 200rpx;
	height: 200rpx;
	background: #f5f5f5;
	border-radius: 10rpx;
}

.product-details {
	flex: 1;
	margin-left: 20rpx;
	display: flex;
	flex-direction: column;
	justify-content: space-between;
}

.product-name {
	font-size: 28rpx;
	color: #333;
	line-height: 40rpx;
}

.product-price {
	font-size: 32rpx;
	color: #ff6b35;
	font-weight: bold;
}

.product-desc {
	font-size: 24rpx;
	color: #999;
	line-height: 32rpx;
}

.close-btn {
	font-size: 48rpx;
	color: #999;
	padding: 0 10rpx;
}

.specs-section,
.quantity-section {
	padding: 20rpx 0;
	border-bottom: 1px solid #f0f0f0;
}

.section-title {
	font-size: 28rpx;
	color: #333;
	margin-bottom: 20rpx;
}

.specs-list {
	display: flex;
	flex-wrap: wrap;
}

.spec-item {
	padding: 15rpx 25rpx;
	border: 1px solid #ddd;
	border-radius: 10rpx;
	margin-right: 20rpx;
	margin-bottom: 20rpx;
	font-size: 26rpx;
	color: #666;
}

.spec-item.active {
	border-color: #ff6b35;
	color: #ff6b35;
}

.quantity-control {
	display: flex;
	align-items: center;
	width: 200rpx;
}

.decrease-btn,
.increase-btn {
	width: 60rpx;
	height: 60rpx;
	border: 1px solid #ddd;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 32rpx;
	color: #666;
}

.decrease-btn.disabled {
	color: #ccc;
}

.quantity {
	width: 80rpx;
	height: 60rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 28rpx;
}

.modal-footer {
	background: #fff;
	padding: 20rpx;
	border-top: 1px solid #f0f0f0;
}

.add-to-cart-btn {
	background: #ff6b35;
	color: #fff;
	text-align: center;
	padding: 25rpx;
	border-radius: 10rpx;
	font-size: 32rpx;
	font-weight: bold;
}
</style>

<style lang="scss" scoped>
// 自定义头部样式
.custom-header {
	position: fixed;
	top: 0;
	left: 0;
	width: 100%;
	z-index: 100;
}

.navbar-fixed {
	background-color: #EEF7F4;
}

.navbar-content {
	display: flex;
	align-items: center;
	justify-content: flex-start;
	padding-left: 30rpx;
	box-sizing: border-box;
}

.navbar-search {
	display: flex;
	align-items: center;
	width: 60%;
	height: 30px;
	background-color: #fff;
	border-radius: 16px;
	padding: 0 12px;
	border: 1rpx solid #20CB6B;
}

.navbar-search_icon {
	margin-right: 8px;
}

.navbar-search_text {
	flex: 1;
}

.navbar-search_text input {
	font-size: 14px;
	color: #333;
	background-color: transparent;
}

// 分类容器样式 - 固定高度
.categories-container {
	position: fixed;
	left: 0;
	right: 0;
	height: 90px;
	/* 固定高度 */
	background-color: #EEF7F4;
	z-index: 90;
	overflow: hidden;
}

// 一级分类样式
.primary-categories {
	height: 100%;
	display: flex;
	flex-direction: row;
}

.primary-scroll {
	flex: 1;
	padding: 10px 0;
	display: flex;
	white-space: nowrap;
}

.primary-category-item {
	min-width: 85px;
	height: 90px;
	display: inline-flex;
	flex-direction: column;
	align-items: center;
	// margin: 0 15px;
}

.primary-category-icon {
	width: 88rpx;
	height: 88rpx;
	border-radius: 16rpx;
	// margin-bottom: 8rpx;
}

.primary-category-name {
	font-size: 24rpx;
	color: #333;
	text-align: center;
}

.primary-category-item.active {
	color: #20CB6B;
}

.primary-category-item.active .primary-category-name {
	color: #fff;
	font-weight: bold;
	background-color: #20CB6B;
	padding: 0 12rpx;
	border-radius: 16rpx;
}

// 展开全部分类按钮
.expand-categories-btn {
	width: 40px;
	height: 100%;
	background-color: #EEF7F4;
	position: absolute;
	right: 0;
	top: 50%;
	transform: translateY(-50%);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	font-size: 26rpx;
	color: #20CB6B;
	z-index: 10;
}

.expand-icon {
	margin-left: 8rpx;
}

// 全部分类弹窗样式
.all-categories-popup {
	width: 100%;
	height: 100vh;
	position: fixed;
	background-color: rgba(0, 0, 0, 0.2);
	z-index: 1000;
	display: flex;
	align-items: flex-start;
	justify-content: center;
}

.all-categories {
	width: 100%;
	background-color: #EEF7F4;
	background-image: linear-gradient(#EEF7F4, #f5f7f6);
	border-radius: 0 0 30rpx 30rpx;
	overflow: hidden;
}

.all-categories-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20rpx 30rpx 0 30rpx;
}

.all-categories-title {
	font-size: 32rpx;
	font-weight: bold;
	color: #333;
}

.collapse-icon {
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0 0 20rpx 0;
	border-radius: 16rpx;
	font-size: 28rpx;
	color: #999;
}

.all-categories-scroll {
	// height: 500rpx;
	padding: 20rpx;
	box-sizing: border-box;
}

.category-grid-container {
	display: flex;
	flex-wrap: wrap;
	justify-content: flex-start;
}

.category-grid-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	width: 20%;
	margin-bottom: 8px;
}

.category-grid-icon {
	width: 100rpx;
	height: 100rpx;
	border-radius: 20rpx;
	margin-bottom: 10rpx;
}

.category-grid-name {
	font-size: 24rpx;
	color: #333;
	text-align: center;

}

.category-grid-item.active .category-grid-name {
	color: #fff;
	font-weight: bold;
	background-color: #20CB6B;
	padding: 0 12rpx;
	border-radius: 16rpx;
}

// 容器样式 - 固定位置并设置高度
.container {
	position: fixed;
	left: 0;
	right: 0;
	bottom: 0;
	overflow: hidden;
	border-top-left-radius: 30rpx;
	border-top-right-radius: 30rpx;
	// 阴影
	box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.1);
}

// 主体内容样式 - 左右布局
.main-content {
	height: 100%;
	display: flex;
	flex-direction: row;
}

// 二级分类样式 - 左侧固定宽度
.secondary-categories {
	width: 180rpx;
	background-color: #fff;
	border-right: 1rpx solid #eee;
}

.secondary-scroll {
	height: 100%;
}

.secondary-category-item {
	padding: 30rpx 10rpx;
	text-align: center;
	font-size: 28rpx;
	color: #333;
	position: relative;
}

.secondary-category-item.active {
	color: #20CB6B;
	font-weight: bold;
	background-color: #EEF7F4;
}

.secondary-category-item.active::after {
	content: '';
	position: absolute;
	left: 0;
	top: 50%;
	transform: translateY(-50%);
	width: 8rpx;
	height: 40rpx;
	border-radius: 20rpx;
	background-color: #20CB6B;
}

// 商品列表容器 - 右侧占据剩余空间
.products-container {
	flex: 1;
	overflow-y: auto;
	padding: 20rpx;
	background-color: #fff;
}

.category-title {
	width: 100%;
	height: 20px;
	border-radius: 10rpx;
	background-color: #FAFAFA;
	font-size: 24rpx;
	color: #5F5F5F;
	padding-bottom: 10rpx;
	// border-bottom: 1rpx solid #eee;
	position: relative;
	display: flex;
	align-items: center;
	justify-content: center;
}

// .category-title::after {
// 	content: '';
// 	position: absolute;
// 	left: 0;
// 	bottom: -1rpx;
// 	width: 60rpx;
// 	height: 4rpx;
// 	background-color: #20CB6B;
// }

/* 商品列表样式 */
.product-list {
	display: flex;
	flex-direction: column;
	padding: 20rpx 0;
}

.product-item {
	width: 100%;
	height: 200rpx;
	display: flex;
	position: relative;
	box-sizing: border-box;
	// padding: 20rpx;
	margin-bottom: 20rpx;
	border-radius: 10rpx;
	touch-action: manipulation;
}

.product-image {
	width: 200rpx;
	height: 200rpx;
	border-radius: 16rpx;
	object-fit: cover;
	flex-shrink: 0;
}

.product-content {
	flex: 1;
	margin-left: 10rpx;
	display: flex;
	flex-direction: column;
	justify-content: space-between;
	padding: 0 10rpx;
}

.product-name {
	font-weight: 600;
	font-size: 28rpx;
	color: #333;
	line-height: 44rpx;
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
	margin-bottom: 10rpx;
}

.product-desc {
	font-size: 24rpx;
	color: #666;
	line-height: 32rpx;
	display: -webkit-box;
	-webkit-line-clamp: 1;
	-webkit-box-orient: vertical;
	overflow: hidden;
	margin-bottom: 15rpx;
}

.product-bottom {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding-right: 10rpx;
	box-sizing: border-box;
}

.price-container {
	display: flex;
	align-items: center;
}

.product-price {
	font-size: 32rpx;
	color: #FF6B6B;
	font-weight: bold;
	margin-right: 10rpx;
}

.original-price {
	font-size: 24rpx;
	color: #999;
	text-decoration: line-through;
}

.add-cart-btn {
	width: 50rpx;
	height: 50rpx;
	background-color: #20CB6B;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	touch-action: manipulation;
}

.empty-tip {
	text-align: center;
	color: #999;
	font-size: 28rpx;
	margin-top: 100rpx;
	padding: 40rpx;
}

// 调试信息样式
.debug-info {
	background-color: #FFF5F5;
	border: 1px solid #FFCCCC;
	border-radius: 8px;
	padding: 12px;
	margin-bottom: 16px;
}

.debug-title {
	display: block;
	font-size: 14px;
	font-weight: bold;
	color: #FF3B30;
	margin-bottom: 8px;
}

.debug-text {
	display: block;
	font-size: 12px;
	color: #666666;
	line-height: 1.5;
	margin-bottom: 4px;
	word-wrap: break-word;
}

/* 商品选择弹窗样式 - 与首页保持一致 */
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

/* 商品信息 */
.modal-product-info {
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

/* 规格选择 */
.modal-specs-section {
	padding: 20rpx;
}

.modal-section-title {
	font-size: 28rpx;
	color: #333;
	margin-bottom: 20rpx;
	display: block;
}

.modal-specs-list {
	display: flex;
	flex-wrap: wrap;
}

.modal-spec-item {
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

.modal-spec-item {
		padding: 15rpx 30rpx;
		border: 1rpx solid #ddd;
		border-radius: 50rpx;
		margin-right: 20rpx;
		margin-bottom: 20rpx;
		font-size: 28rpx;
		color: #333;
	}

	.modal-spec-item.active {
		border-color: #20CB6B;
		color: #20CB6B;
		background-color: #f0fff4;
	}

	.modal-spec-description {
		margin: 0 10rpx;
		color: #666;
	}

	.modal-spec-price {
		color: #ff4d4f;
		margin-left: 10rpx;
		font-weight: bold;
	}

/* 数量选择 */
.modal-quantity-section {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20rpx;
	border-top: 1rpx solid #f0f0f0;
	border-bottom: 1rpx solid #f0f0f0;
}

.modal-quantity-control {
	display: flex;
	align-items: center;
}

.modal-decrease-btn,
.modal-increase-btn {
	width: 60rpx;
	height: 60rpx;
	border-radius: 50%;
	display: flex;
	justify-content: center;
	align-items: center;
}

.modal-decrease-btn {
	margin-right: 30rpx;
	border: 1rpx solid #ddd;
}

.modal-increase-btn {
	margin-left: 30rpx;
	background-color: #20CB6B !important;
}

.modal-decrease-btn.disabled {
	color: #ccc;
}

.modal-quantity {
	font-size: 32rpx;
	color: #333;
	font-weight: bold;
}

/* 底部按钮 */
.modal-bottom {
	padding: 20rpx;
}

.modal-add-to-cart-btn {
	background-color: #20CB6B !important;
	color: #fff;
	font-size: 32rpx;
	font-weight: bold;
	text-align: center;
	padding: 25rpx 0;
	border-radius: 80rpx;
}

.minus-btn-icon{
	width: 32rpx;
	height: 32rpx;
}
</style>