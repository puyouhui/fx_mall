<template>
	<view class="container">
		<!-- 自定义悬浮透明顶部组件 -->
		<view class="custom-header" :class="{ 'custom-header-white': hasScrolled }"
			:style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="header-buttons">
				<view class="header-buttons-left">
					<view class="header-fg"></view>
					<view class="header-btn-left" @click="goBack">
						<uni-icons type="left" size="20" color="#2E2E2E"></uni-icons>
					</view>
					<view class="header-btn-left" @click="searchProduct">
						<uni-icons type="search" size="20" color="#2E2E2E"></uni-icons>
					</view>
				</view>
				<view class="header-share" @click="shareProduct">
					<image src="/static/font/wechat.svg" mode="aspectFit" class="share-icon"></image>
					<text class="share-text">分享</text>
				</view>
			</view>
		</view>

		<view class="product-image-container">
			<!-- 商品图片轮播 -->
			<swiper class="image-swiper" v-if="product.images && product.images.length > 0" autoplay indicator-dots
				circular @change="onSwiperChange">
				<swiper-item v-for="(image, index) in product.images" :key="index">
					<image :src="image" mode="aspectFill" class="product-image"></image>
				</swiper-item>
			</swiper>
			<view class="image-count" v-if="product.images && product.images.length > 0">{{ currentImageIndex + 1 }}/{{
				product.images.length }}</view>
		</view>



		<!-- 商品信息 -->
		<view class="product-info">
			<!-- 买贵反馈提示 -->
			<view class="price-feedback">
				<text class="feedback-text">价格贵了？点击反馈有惊喜哦~</text>
				<text class="feedback-arrow">
					<uni-icons type="right" size="16" color="#ffffff"></uni-icons>
				</text>
			</view>
			<view class="product-info-content">
				<view class="price-container">
					<text class="price-symbol"></text>
					<text class="product-price">{{ product.priceRange || '0.00' }}</text>
				</view>
				<text class="product-name">{{ product.name }}</text>
				<text class="product-desc">{{ product.description }}</text>
			</view>
			<view class="quality-guarantee" v-if="product.isSpecial">
				<text class="guarantee-icon">✓</text>
				<text class="guarantee-text">保质期权益</text>
			</view>
		</view>

		<!-- 商品规格 - 新增选择卡片 -->
		<view class="product-specs" v-if="product.specs && product.specs.length > 0">
			<view class="section-title">可选规格</view>
			<view class="spec-list">
				<view class="spec-card" v-for="(spec, index) in product.specs" :key="index">
					<view class="spec-info">
						<view class="spec-header">
							<text class="spec-name">{{ spec.name }}</text>
							<text class="spec-description" v-if="spec.description && userType === 'wholesale'"> ({{ spec.description }})</text>
						</view>
						<view class="spec-prices">
							<view class="spec-price-container" :class="{ 'wholesale-layout': userType === 'wholesale' }">
								<text v-if="userType === 'wholesale'" class="spec-price">
									批发价: ￥{{ formatSpecPrice(spec, 'wholesale') }}
								</text>
								<text v-else class="spec-price">
									￥{{ formatSpecPrice(spec, 'retail') }}
								</text>
								<!-- 批发用户显示零售价（灰色） -->
								<text v-if="userType === 'wholesale'" class="spec-retail-price">
									零售价: ￥{{ formatSpecPrice(spec, 'retail') }}
								</text>
							</view>
						</view>
					</view>
					<view class="spec-action">
						<view class="stepper" v-if="spec.quantity > 0">
							<view class="minus-btn" @click="decreaseQuantity(spec)">
								<image src="/static/icon/minus.png" class="category-icon"></image>
							</view>
							<text class="quantity">{{ spec.quantity }}</text>
							<view class="plus-btn" @click="increaseQuantity(spec)">
								<uni-icons type="plusempty" size="20" color="#fff"></uni-icons>
							</view>
						</view>
						<view class="add-btn" v-else @click="addSpecToCart(spec)">
							<uni-icons type="plusempty" size="20" color="#fff"></uni-icons>
						</view>
					</view>
				</view>
			</view>
		</view>

		<!-- 常见问题卡片 -->
		<view class="faq-card">
			<view class="faq-title">常见问题</view>
			<view class="faq-content">
				<view class="faq-item">
					<view class="faq-question">
						<text class="faq-question-icon">•</text>
						<text class="faq-question-text">什么时候送到？</text>
					</view>
					<text class="faq-answer">我们的配送时间是早上8:00 - 21:00，如您需要加急，请联系您的客户经理，无法指定时间送达，望您谅解。</text>
				</view>
				<view class="faq-item">
					<view class="faq-question">
						<text class="faq-question-icon">•</text>
						<text class="faq-question-text">今日购买的商品什么时候送到？</text>
					</view>
					<text class="faq-answer">每个区域配送到达时间不同，昆明主城区每日多次配送，若下单后当日无法送达，您的客户经理会与您取得联系！</text>
				</view>
				<view class="faq-item">
					<view class="faq-question">
						<text class="faq-question-icon">•</text>
						<text class="faq-question-text">平台下面售后有什么条件？</text>
					</view>
					<text class="faq-answer">因一次性用品商品特殊性，如签收后未拆封，不影响二次销售，平台提供免费退换服务，如遇质量问题请于2天内联系客户经理申请售后。</text>
				</view>
				<view class="faq-item">
					<view class="faq-question">
						<text class="faq-question-icon">•</text>
						<text class="faq-question-text">申请退款后，多长时间处理售后订单？</text>
					</view>
					<text class="faq-answer">仓库收到退回商品后，会在当日16:00-22:00间由系统原路返还到您的付款账户中。</text>
				</view>
				<view class="faq-note">
					<text>小程序上的图片以收到实物为准，产品描述仅供参考，并不完全准确，最终解释权归本公司所有。</text>
				</view>
			</view>
		</view>

		<view style="padding-bottom: 100px;">

		</view>

		<!-- 商品详情 -->
		<!-- <view class="product-details" v-if="product.details">
			<view class="section-title">产品规格</view>
			<view class="details-content">
				<view class="detail-row">
					<text class="detail-label">【产品规格】</text>
					<text class="detail-value">{{product.specifications ? product.specifications.map(s =>
						s.value).join(' ') : ''}}</text>
				</view>
				<view class="detail-row">
					<text class="detail-label">【产品名称】</text>
					<text class="detail-value">{{ product.name }}</text>
				</view>
			</view>
		</view> -->

		<!-- 商品信息卡片 -->
		<!-- <view class="info-cards">
			<view class="info-card">
				<text class="card-label">储存方式</text>
				<text class="card-value">常温</text>
			</view>
			<view class="info-card">
				<text class="card-label">保质期</text>
				<text class="card-value">12个月</text>
			</view>
			<view class="info-card">
				<text class="card-label">商品分类</text>
				<text class="card-value">{{ product.categoryName || '食品饮料' }}</text>
			</view>
		</view> -->

		<!-- 配送信息 -->
		<!-- <view class="delivery-info">
			<view class="section-title">配送</view>
			<view class="delivery-content">
				<text class="delivery-text">{{ deliveryAddress || '未配置收货地址' }}</text>
				<text class="delivery-arrow">→</text>
			</view>
		</view> -->

		<!-- 售后信息 -->
		<!-- <view class="after-sale-info">
			<view class="section-title">售后</view>
			<view class="after-sale-content">
				<text class="after-sale-text">非质量问题{{ refundDays }}天内反馈 质量问题{{ refundDays }}天内反馈 轻</text>
			</view>
		</view> -->

		<!-- 底部固定操作栏 -->
		<view class="bottom-action">
			<view class="bottom-action-container">
				<view class="action-buttons">
					<view class="action-btn" @click="collectProduct">
						<uni-icons type="star" size="28" color="#2C2C2C"></uni-icons>
						<text class="action-text">收藏</text>
					</view>
					<view class="action-btn" @click="collectProduct">
						<uni-icons type="cart" size="28" color="#2C2C2C"></uni-icons>
						<text class="action-text">采购单</text>
					</view>
					<view class="action-btn" @click="collectProduct">
						<uni-icons type="phone" size="28" color="#2C2C2C"></uni-icons>
						<text class="action-text">客服</text>
					</view>
				</view>
				<view class="right-actions">
					<view class="add-to-cart-btn" @click="addToCart">
						<uni-icons type="plusempty" color="#fff" size="20"
							style="padding-right: 14rpx;padding-top: 2rpx;"></uni-icons>
						加入采购单
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script>
import { getProductDetail } from '../../api/products';
import { getMiniUserInfo } from '../../api/index';
export default {
	data() {
		return {
			product: {
				id: 0,
				name: '',
				description: '',
				price: 0, // 保留字段但不再使用
				originalPrice: 0, // 保留字段但不再使用
				priceRange: '',
				minPrice: 0,
				maxPrice: 0,
				categoryId: 0,
				categoryName: '',
				isSpecial: false,
				images: [],
				specifications: [], // 保留字段
				specs: [], // 实际使用的规格数组
				stock: 0,
				sales: 0,
				details: ''
			},
			currentImageIndex: 0,
			isCollected: false,
			deliveryAddress: '',
			refundDays: 6,
			statusBarHeight: 0,
			cartCount: 0, // 购物车商品总数
			scrollTop: 0, // 滚动距离
			hasScrolled: false, // 是否已经滚动
			userType: null // 用户类型：'retail' | 'wholesale' | null（未登录）
		};
	},
	onLoad(options) {
		// 初始化用户类型
		this.initUserType();
		
		// 获取设备信息，特别是状态栏高度
		const systemInfo = uni.getSystemInfoSync();
		this.statusBarHeight = systemInfo.statusBarHeight;

		if (options && options.id) {
			this.loadProductDetail(options.id);
		}
	},
	// 页面显示时更新用户信息
	onShow() {
		this.updateUserInfo();
	},
	onPageScroll(e) {
		// 监听页面滚动
		const currentScrollTop = e.scrollTop;
		// 当向上滚动超过50px时，顶部变为白色
		if (currentScrollTop > 50 && currentScrollTop > this.scrollTop) {
			this.hasScrolled = true;
		} else if (currentScrollTop <= 20) {
			this.hasScrolled = false;
		}
		this.scrollTop = currentScrollTop;
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
					// 重新计算价格
					if (this.product && this.product.id) {
						this.calculatePriceRange();
					}
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
					// 重新计算价格
					if (this.product && this.product.id) {
						this.calculatePriceRange();
					}
				}
			} catch (error) {
				console.error('获取用户信息失败:', error);
				// 静默失败，不显示错误提示
				this.userType = null;
			}
		},
		
		// 加载商品详情
		async loadProductDetail(productId) {
			try {
				// 显示加载动画
				uni.showLoading({
					title: '加载中',
					mask: true
				});

				const res = await getProductDetail(parseInt(productId));
				if (res.code === 200 && res.data) {
					this.product = res.data;
					// 如果后端没有返回priceRange，前端计算
					if (!this.product.priceRange && this.product.specs && this.product.specs.length > 0) {
						this.calculatePriceRange();
					}
					// 处理后端返回的数据结构差异
					// 转换下划线命名为驼峰命名
					if (this.product.original_price !== undefined) {
						this.product.originalPrice = this.product.original_price;
					}
					if (this.product.category_id !== undefined) {
						this.product.categoryId = this.product.category_id;
					}
					if (this.product.category_name !== undefined) {
						this.product.categoryName = this.product.category_name;
					}

					// 如果后端返回的是specifications而不是specs，兼容处理
					if (!this.product.specs && this.product.specifications && this.product.specifications.length > 0) {
						// 检查specifications的结构，进行适当转换
						if (this.product.specifications[0].price === undefined) {
							// 如果specifications没有价格信息，创建默认的specs结构
							this.product.specs = this.product.specifications.map((spec, index) => ({
								id: index + 1,
								name: spec.name,
								description: spec.value || '',
								price: parseFloat(this.product.price) || 0
							}));
						} else {
							// 直接使用specifications作为specs
							this.product.specs = this.product.specifications;
						}
					}

					// 确保所有规格都有id，并处理价格字段
					if (this.product.specs && this.product.specs.length > 0) {
						this.product.specs.forEach((spec, index) => {
							if (spec.id === undefined) {
								spec.id = index + 1;
							}
							// 确保批发价和零售价为数字
							if (spec.wholesale_price !== undefined) {
								spec.wholesale_price = parseFloat(spec.wholesale_price) || 0;
							}
							if (spec.wholesalePrice !== undefined) {
								spec.wholesalePrice = parseFloat(spec.wholesalePrice) || 0;
							}
							if (spec.retail_price !== undefined) {
								spec.retail_price = parseFloat(spec.retail_price) || 0;
							}
							if (spec.retailPrice !== undefined) {
								spec.retailPrice = parseFloat(spec.retailPrice) || 0;
							}
						});
					}
					// 初始化规格数量并从本地存储同步
					this.initSpecQuantities();
					// 更新购物车数量显示
					this.updateCartCount();
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
				uni.hideLoading();
			}
		},

		// 图片切换事件
		onSwiperChange(e) {
			this.currentImageIndex = e.detail.current;
		},

		// 计算价格范围（根据用户类型显示批发价或零售价）
		calculatePriceRange() {
			// 处理没有规格的情况
			if (!this.product.specs || this.product.specs.length === 0) {
				// 使用商品本身的价格，如果有
				const productPrice = parseFloat(this.product.price) || 0;
				this.product.minPrice = productPrice;
				this.product.maxPrice = productPrice;
				this.product.priceRange = '¥' + productPrice.toFixed(2);
				return;
			}

			// 根据用户类型决定显示哪种价格
			const isWholesaleUser = this.userType === 'wholesale';
			
			// 收集价格
			const prices = [];
			this.product.specs.forEach(spec => {
				if (isWholesaleUser) {
					// 批发用户：显示批发价
					const wholesalePrice = spec.wholesale_price || spec.wholesalePrice;
					if (wholesalePrice && wholesalePrice > 0) {
						prices.push(parseFloat(wholesalePrice));
					}
				} else {
					// 未登录或零售用户：显示零售价
					const retailPrice = spec.retail_price || spec.retailPrice;
					if (retailPrice && retailPrice > 0) {
						prices.push(parseFloat(retailPrice));
					}
				}
			});

			// 如果没有找到对应类型的价格，使用另一种价格作为后备
			if (prices.length === 0) {
				this.product.specs.forEach(spec => {
					if (isWholesaleUser) {
						// 批发用户找不到批发价，使用零售价作为后备
						const retailPrice = spec.retail_price || spec.retailPrice;
						if (retailPrice && retailPrice > 0) {
							prices.push(parseFloat(retailPrice));
						}
					} else {
						// 零售用户找不到零售价，使用批发价作为后备
						const wholesalePrice = spec.wholesale_price || spec.wholesalePrice;
						if (wholesalePrice && wholesalePrice > 0) {
							prices.push(parseFloat(wholesalePrice));
						}
					}
					// 最后使用通用价格字段
					if (prices.length === 0 && spec.price && spec.price > 0) {
						prices.push(parseFloat(spec.price));
					}
				});
			}

			// 处理所有价格为0的情况
			if (prices.length === 0) {
				const fallbackPrice = parseFloat(this.product.price) || 0;
				this.product.minPrice = fallbackPrice;
				this.product.maxPrice = fallbackPrice;
				this.product.priceRange = '¥' + fallbackPrice.toFixed(2);
				return;
			}

			// 只显示最低价格
			this.product.minPrice = Math.min(...prices);
			this.product.maxPrice = Math.max(...prices);
			this.product.priceRange = '¥' + this.product.minPrice.toFixed(2);
		},

		// 格式化规格价格
		formatSpecPrice(spec, priceType = 'retail') {
			let price = 0;
			if (priceType === 'wholesale') {
				// 批发价
				price = spec.wholesale_price || spec.wholesalePrice || 0;
			} else {
				// 零售价
				price = spec.retail_price || spec.retailPrice || 0;
			}
			
			// 如果指定类型的价格不存在，使用通用价格字段作为后备
			if (!price || price === 0) {
				price = spec.price || this.product.price || 0;
			}
			
			return parseFloat(price || 0).toFixed(2);
		},

		// 初始化规格数量并从本地存储同步
		initSpecQuantities() {
			// 获取现有采购单数据
			const cart = uni.getStorageSync('cart') || [];

			// 为每个规格初始化quantity为0
			this.product.specs.forEach(spec => {
				spec.quantity = 0;

				// 检查该规格是否已在购物车中
				const specKey = spec.name + (spec.description ? ':' + spec.description : '');
				const cartItem = cart.find(item =>
					item.productId === this.product.id && item.specKey === specKey
				);

				// 如果在购物车中，更新数量
				if (cartItem) {
					spec.quantity = cartItem.quantity;
				}
			});
		},

		// 添加规格到采购单
		addSpecToCart(spec) {
			// 获取现有采购单数据
			let cart = uni.getStorageSync('cart') || [];

			// 为规格创建唯一标识，即使信息不完整也能正确处理
			const specName = spec.name || '默认规格';
			const specDesc = spec.description || spec.value || '';
			const specKey = specName + (specDesc ? ':' + specDesc : '');
			
			// 根据用户类型选择价格
			let specPrice = 0;
			if (this.userType === 'wholesale') {
				// 批发用户使用批发价
				specPrice = parseFloat(spec.wholesale_price || spec.wholesalePrice || spec.price || this.product.price || 0);
			} else {
				// 零售用户或未登录用户使用零售价
				specPrice = parseFloat(spec.retail_price || spec.retailPrice || spec.price || this.product.price || 0);
			}

			// 检查商品是否已在采购单中
			const index = cart.findIndex(item =>
				item.productId === this.product.id && item.specKey === specKey
			);

			if (index > -1) {
				// 商品已存在，增加数量
				cart[index].quantity += 1;
			} else {
				// 商品不存在，添加新商品
				cart.push({
					productId: this.product.id,
					quantity: 1,
					name: this.product.name,
					specName: specName,
					specDescription: specDesc,
					specKey: specKey,
					wholesalePrice: parseFloat(spec.wholesale_price || spec.wholesalePrice) || 0,
					retailPrice: parseFloat(spec.retail_price || spec.retailPrice) || 0,
					price: specPrice, // 根据用户类型选择的价格
					image: this.product.images && this.product.images.length > 0 ? this.product.images[0] : '',
					isSpecial: this.product.isSpecial
				});
			}

			// 保存到本地存储
			uni.setStorageSync('cart', cart);

			// 更新规格数量
			spec.quantity = (spec.quantity || 0) + 1;

			// 更新购物车数量
			this.updateCartCount();

			// 提示成功
			uni.showToast({
				title: '已添加到采购单',
				icon: 'success',
				duration: 2000
			});
		},

		// 增加规格数量
		increaseQuantity(spec) {
			// 获取现有采购单数据
			let cart = uni.getStorageSync('cart') || [];

			// 为规格创建唯一标识，即使信息不完整也能正确处理
			const specName = spec.name || '默认规格';
			const specDesc = spec.description || spec.value || '';
			const specKey = specName + (specDesc ? ':' + specDesc : '');

			// 查找采购单中的对应项
			const index = cart.findIndex(item =>
				item.productId === this.product.id && item.specKey === specKey
			);

			if (index > -1) {
				// 增加采购单中的数量
				cart[index].quantity += 1;
				// 保存到本地存储
				uni.setStorageSync('cart', cart);
				// 更新规格数量
				spec.quantity = (spec.quantity || 0) + 1;
				// 更新购物车数量
				this.updateCartCount();
			}
		},

		// 减少规格数量
		decreaseQuantity(spec) {
			// 确保数量是有效的数字
			const currentQuantity = parseInt(spec.quantity) || 0;

			if (currentQuantity <= 1) {
				// 如果数量为1或更少，减少后从购物车移除
				this.removeSpecFromCart(spec);
			} else {
				// 获取现有采购单数据
				let cart = uni.getStorageSync('cart') || [];

				// 为规格创建唯一标识，即使信息不完整也能正确处理
				const specName = spec.name || '默认规格';
				const specDesc = spec.description || spec.value || '';
				const specKey = specName + (specDesc ? ':' + specDesc : '');

				// 查找采购单中的对应项
				const index = cart.findIndex(item =>
					item.productId === this.product.id && item.specKey === specKey
				);

				if (index > -1) {
					// 减少采购单中的数量
					cart[index].quantity -= 1;
					// 保存到本地存储
					uni.setStorageSync('cart', cart);
					// 更新规格数量
					spec.quantity = currentQuantity - 1;
					// 更新购物车数量
					this.updateCartCount();
				}
			}
		},

		// 从采购单移除规格
		removeSpecFromCart(spec) {
			// 获取现有采购单数据
			let cart = uni.getStorageSync('cart') || [];

			// 为规格创建唯一标识，即使信息不完整也能正确处理
			const specName = spec.name || '默认规格';
			const specDesc = spec.description || spec.value || '';
			const specKey = specName + (specDesc ? ':' + specDesc : '');

			// 查找采购单中的对应项并过滤掉
			cart = cart.filter(item =>
				!(item.productId === this.product.id && item.specKey === specKey)
			);

			// 保存到本地存储
			uni.setStorageSync('cart', cart);

			// 更新规格数量
			spec.quantity = 0;

			// 更新购物车数量
			this.updateCartCount();
		},

		// 更新购物车商品总数
		updateCartCount() {
			// 获取现有采购单数据
			const cart = uni.getStorageSync('cart') || [];
			// 计算商品总数
			this.cartCount = cart.reduce((total, item) => total + item.quantity, 0);
		},

		// 添加到采购单（兼容旧调用）
		addToCart() {
			// 提示用户直接点击规格添加
			uni.showToast({
				title: '请点击具体规格添加',
				icon: 'none',
				duration: 2000
			});
		},

		// 收藏商品
		collectProduct() {
			// 模拟收藏功能
			this.isCollected = !this.isCollected;
			uni.showToast({
				title: this.isCollected ? '收藏成功' : '取消收藏',
				icon: 'none',
				duration: 2000
			});
		},

		// 跳转到购物车
		goToCart() {
			uni.navigateTo({
				url: '/pages/cart/index'
			});
		},

		// 跳转到登录页面
		goToLogin() {
			// 模拟登录跳转
			uni.showToast({
				title: '跳转到登录页面',
				icon: 'none',
				duration: 2000
			});
		},

		// 返回上一页
		goBack() {
			uni.navigateBack();
		},

		// 搜索商品
		searchProduct() {
			uni.showToast({
				title: '打开搜索页面',
				icon: 'none',
				duration: 2000
			});
		},

		// 分享商品
		shareProduct() {
			uni.showShareMenu({
				withShareTicket: true,
				menus: ['shareAppMessage', 'shareTimeline']
			});
		}
	}
};
</script>

<style>
.container {
	display: flex;
	flex-direction: column;
	height: 100vh;
	background-color: #f5f5f5;
}

/* 自定义悬浮透明顶部组件样式 */
.custom-header {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	z-index: 999;
	padding-bottom: 20rpx;
	padding-right: 25%;
	transition: background-color 0.3s ease;
}

/* 滚动后顶部变为白色的样式 */
.custom-header-white {
	background-color: #fff;
	box-shadow: 0 2rpx 10rpx rgba(0, 0, 0, 0.1);
}

.custom-header-white .header-buttons-left {
	background-color: #fff;
	box-shadow: 0 2rpx 10rpx rgba(0, 0, 0, 0.1);
}

.custom-header-white .header-share {
	background-color: #fff;
	box-shadow: 0 2rpx 10rpx rgba(0, 0, 0, 0.1);
}

.header-buttons {
	display: flex;
	justify-content: space-between;
	padding: 0 30rpx;
	align-items: center;
}

.header-buttons-left {
	display: flex;
	width: 100px;
	height: 40px;
	border-radius: 20px;
	justify-content: space-between;
	background-color: rgba(255, 255, 255, 0.8);
	position: relative;
}

.header-fg {
	width: 1px;
	height: 20px;
	background-color: #ccc;
	position: absolute;
	top: 50%;
	left: 50%;
	transform: translate(-50%, -50%);
}

.header-btn-left {
	width: 50%;
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 8rpx 0;
}

.header-btn-left .uni-icons {
	display: flex;
	justify-content: center;
	align-items: center;
}

/* 调整左右按钮组与右侧分享按钮的垂直对齐 */
.header-buttons>view {
	display: flex;
	align-items: center;
}

.header-share {
	height: 40px;
	display: flex;
	justify-content: center;
	align-items: center;
	background-color: rgba(255, 255, 255, 0.8);
	padding: 0 20rpx;
	border-radius: 20px;
}

.share-icon {
	width: 40rpx;
	height: 40rpx;
}

.share-text {
	font-size: 24rpx;
	color: #2E2E2E;
	margin-left: 8rpx;
	font-weight: 500;
}

/* 商品图片轮播样式 */
.image-swiper {
	min-height: 375px;
	width: 100%;
	background-color: #fff;
}

.product-image {
	height: 375px;
	width: 100%;
}

.product-image-container {
	position: relative;
}

.image-count {
	position: absolute;
	bottom: 20rpx;
	right: 20rpx;
	background-color: rgba(0, 0, 0, 0.5);
	color: #fff;
	font-size: 24rpx;
	padding: 8rpx 16rpx;
	border-radius: 20rpx;
}

/* 买贵反馈提示 */
.price-feedback {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 16rpx 20rpx;
	background: linear-gradient(90deg, #4CAF50, #8BC34A);
	color: #fff;
	border-radius: 20rpx 20rpx 0 0;
}

.feedback-text {
	font-size: 26rpx;
}

.feedback-arrow {
	font-size: 28rpx;
}

/* 商品信息样式 */
.product-info {
	width: 96%;
	/* padding: 20rpx 20rpx 20rpx 20rpx; */
	background-color: #fff;
	margin: 20rpx auto;
	box-sizing: border-box;
	border-radius: 20rpx;
}

.product-info-content {
	padding: 20rpx 20rpx 20rpx 20rpx;
}

.price-container {
	display: flex;
	align-items: baseline;
	margin-bottom: 15rpx;
}

.price-symbol {
	font-size: 28rpx;
	color: #f00;
	font-weight: bold;
	margin-right: 4rpx;
}

.product-price {
	font-size: 40rpx;
	color: #f00;
	font-weight: bold;
}

.product-name {
	font-size: 36rpx;
	font-weight: bold;
	color: #333;
	line-height: 50rpx;
	display: block;
	margin-bottom: 15rpx;
}

.product-desc {
	font-size: 28rpx;
	color: #666;
	line-height: 40rpx;
	margin-bottom: 15rpx;
}

.quality-guarantee {
	display: flex;
	align-items: center;
}

.guarantee-icon {
	color: #4CAF50;
	margin-right: 8rpx;
}

.guarantee-text {
	font-size: 26rpx;
	color: #4CAF50;
}

/* 商品规格样式 */
.product-specs {
	width: 96%;
	background-color: #fff;
	padding: 20rpx;
	margin: 0 auto 20rpx auto;
	box-sizing: border-box;
	border-radius: 20rpx;
}

.spec-card {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20rpx;
	margin-bottom: 15rpx;
	background-color: #f9f9f9;
	border-radius: 15rpx;
	transition: all 0.3s ease;
}

.spec-card:last-child {
	margin-bottom: 0;
}

.spec-info {
	flex: 1;
}

.spec-name {
	font-size: 28rpx;
	color: #333;
	font-weight: 500;
}

.spec-description {
	font-size: 26rpx;
	color: #666;
	line-height: 45rpx;
}

.spec-price {
	font-size: 32rpx;
	color: #f00;
	font-weight: bold;
	/* margin-left: 10rpx; */
}

.spec-original-price {
	font-size: 24rpx;
	color: #999;
	text-decoration: line-through;
}

.spec-action {
	display: flex;
	align-items: center;
	padding-right: 20rpx;
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
	padding: 2rpx 0 0 2rpx;
	box-sizing: border-box;
}

.stepper {
	display: flex;
	align-items: center;
}

.minus-btn,
.plus-btn {
	width: 30px;
	height: 30px;
	background-color: #20CB6B;
	color: #333;
	font-size: 24rpx;
	display: flex;
	justify-content: center;
	align-items: center;
	border-radius: 50%;
}

.minus-btn {
	margin-right: 20rpx;
	background-color: #dfe6e9;
	font-size: 20rpx;
	font-weight: bold;
	line-height: 50rpx;
	text-align: center;
}

.minus-btn image {
	width: 30rpx;
	height: 30rpx;
	padding-top: 1rpx;
}

.plus-btn {
	margin-left: 20rpx;
}

.quantity {
	font-size: 28rpx;
	color: #333;
	font-weight: bold;
}

/* 商品详情样式 */
.product-details {
	width: 96%;
	background-color: #fff;
	padding: 20rpx;
	margin-bottom: 20rpx;
	margin: 0 auto;
	box-sizing: border-box;
	border-radius: 20rpx;
}

.details-content {
	padding: 10rpx;
}

.detail-row {
	margin-bottom: 10rpx;
}

.detail-label {
	font-size: 28rpx;
	color: #666;
}

.detail-value {
	font-size: 28rpx;
	color: #333;
}

.section-title {
	font-size: 32rpx;
	font-weight: bold;
	color: #333;
	margin-bottom: 20rpx;
	padding-bottom: 10rpx;
	border-bottom: 1rpx solid #eee;
}


.spec-item {
	display: flex;
	padding: 15rpx 0;
	border-bottom: 1rpx solid #f0f0f0;
}

.spec-item:last-child {
	border-bottom: none;
}

.spec-name {
	font-size: 32rpx;
	color: #333;
	margin-right: 10rpx;
	font-weight: 700;
}

.spec-value {
	font-size: 28rpx;
	color: #333;
}

.spec-description {
	font-size: 26rpx;
	color: #999;
}

.spec-option-description {
	font-size: 26rpx;
	color: #999;
}

/* 信息卡片样式 */
.info-cards {
	display: flex;
	background-color: #fff;
	margin-bottom: 20rpx;
}

.info-card {
	flex: 1;
	padding: 20rpx;
	display: flex;
	flex-direction: column;
	/* align-items: center; */
	border-right: 1rpx solid #f0f0f0;
}

.info-card:last-child {
	border-right: none;
}

.card-label {
	font-size: 26rpx;
	color: #666;
	margin-bottom: 8rpx;
}

.card-value {
	font-size: 28rpx;
	color: #333;
	font-weight: 500;
}

/* 规格信息样式 */
.spec-info {
	flex-direction: column;
	align-items: flex-start;
}

.spec-header {
	width: 100%;
	display: flex;
}

.spec-prices {
	width: 100%;
	display: flex;
	align-items: center;
	margin-top: 10rpx;
}

.spec-price-container {
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.spec-price-container.wholesale-layout {
	flex-direction: row;
	align-items: baseline;
	gap: 12rpx;
}

.spec-price {
	font-size: 32rpx;
	color: #f00;
	font-weight: bold;
}

.spec-retail-price {
	font-size: 24rpx;
	color: #999;
	text-decoration: line-through;
}

.spec-original-price {
	margin-left: 15rpx;
}

/* 规格选择弹窗样式 */
.spec-modal {
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

.spec-modal-overlay {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background-color: rgba(0, 0, 0, 0.5);
}

.spec-modal-content {
	width: 100%;
	max-height: 60vh;
	background-color: #fff;
	border-radius: 30rpx 30rpx 0 0;
	position: relative;
	z-index: 1;
	overflow-y: auto;
}

.spec-modal-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20rpx 30rpx;
	border-bottom: 1rpx solid #eee;
}

.spec-modal-title {
	font-size: 32rpx;
	font-weight: bold;
}

.spec-modal-close {
	font-size: 48rpx;
	color: #999;
}

.spec-modal-body {
	padding: 20rpx;
}

.spec-option {
	padding: 20rpx;
	border-bottom: 1rpx solid #f0f0f0;
}

.spec-option:last-child {
	border-bottom: none;
}

.spec-info {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 10rpx;
}

.spec-option-name {
	font-size: 30rpx;
	color: #333;
}

.spec-option-price {
	font-size: 32rpx;
	color: #f00;
	font-weight: bold;
}

.spec-option-original-price {
	font-size: 26rpx;
	color: #999;
	text-decoration: line-through;
}

/* 配送信息样式 */
.delivery-info {
	background-color: #fff;
	padding: 20rpx;
	margin-bottom: 20rpx;
}

.delivery-content {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 10rpx;
}

.delivery-text {
	font-size: 28rpx;
	color: #333;
}

.delivery-arrow {
	font-size: 28rpx;
	color: #999;
}

/* 售后信息样式 */
.after-sale-info {
	background-color: #fff;
	padding: 20rpx;
	margin-bottom: 120rpx;
}

.after-sale-content {
	padding: 10rpx;
}

.after-sale-text {
	font-size: 28rpx;
	color: #333;
}

/* 底部操作栏样式 */
.bottom-action {
	width: 100%;
	background-color: #fff;
	padding: 15rpx 0 10rpx 0;
	position: fixed;
	bottom: 0;
	border-top: 1rpx solid #eee;
}

.bottom-action-container {
	width: 100%;
	/* height: 60px; */
	display: flex;
	align-items: center;
	padding-bottom: env(safe-area-inset-bottom);
	/* padding: 15rpx 20rpx; */
	background-color: #fff;
	touch-action: manipulation;
}

.action-buttons {
	width: 50%;
	height: 48px;
	/* background-color: #666; */
	display: flex;
	margin-right: 20rpx;
	padding-left: 30rpx;
}

.right-actions {
	width: 50%;
	height: 48px;
	display: flex;
	justify-content: flex-end;
	padding-right: 30rpx;
}

.action-btn {
	display: flex;
	flex-direction: column;
	align-items: center;
	margin-right: 40rpx;
}

.btn-icon {
	font-size: 40rpx;
	margin-bottom: 8rpx;
}

.btn-text {
	font-size: 22rpx;
	color: #666;
}

.add-to-cart-btn {
	width: 90%;
	background-color: #20CB6B;
	color: #fff;
	font-size: 32rpx;
	font-weight: bold;
	padding: 20rpx 0;
	border-radius: 60rpx;
	border: none;
	touch-action: manipulation;
	margin-right: 15rpx;
	display: flex;
	justify-content: center;
}

.login-btn {
	width: 180rpx;
	background-color: #007AFF;
	color: #fff;
	font-size: 32rpx;
	font-weight: bold;
	padding: 20rpx 0;
	border-radius: 60rpx;
	border: none;
	touch-action: manipulation;
}

/* 常见问题卡片样式 */
.faq-card {
	width: 96%;
	background-color: #fff;
	padding: 20rpx;
	margin: 0 auto 20rpx auto;
	box-sizing: border-box;
	border-radius: 20rpx;
}

.faq-title {
	width: 100%;
	text-align: center;
	color: #20CB6B;
	font-size: 32rpx;
	font-weight: bold;
	margin-bottom: 20rpx;
}

.faq-content {
	padding: 10rpx 0;
}

.faq-item {
	margin-bottom: 20rpx;
}

.faq-item:last-child {
	margin-bottom: 10rpx;
}

.faq-question {
	display: flex;
	align-items: flex-start;
	margin-bottom: 8rpx;
}

.faq-question-icon {
	color: #4CAF50;
	font-size: 28rpx;
	margin-right: 10rpx;
	font-weight: bold;
}

.faq-question-text {
	font-size: 28rpx;
	color: #333;
	font-weight: 500;
	flex: 1;
}

.faq-answer {
	font-size: 26rpx;
	color: #666;
	line-height: 40rpx;
}

.faq-note {
	padding: 15rpx;
	background-color: #f9f9f9;
	border-radius: 10rpx;
	margin-top: 10rpx;
}

.faq-note text {
	font-size: 24rpx;
	color: #999;
	line-height: 36rpx;
}

.action-text {
	padding-top: -10rpx;
}
</style>