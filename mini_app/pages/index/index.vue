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
						<input type="text" placeholder="请输入产品名称查询" placeholder-style="color: #999;" disabled />
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
					<text class="tag-text">价格透明</text>
				</view>
				<view class="tag-item">
					<image src="/static/icon/coin1.png" class="tag-icon"></image>
					<text class="tag-text">一件起送</text>
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
			<view class="section-title hot-section-title">
				<view class="section-left">
					<view class="hot-tag">HOT</view>
					<view class="section-title-text">
						<text class="section-name">热销产品</text>
						<!-- <text class="section-subtitle">人气精选 · 限量推荐</text> -->
					</view>
				</view>
				<text class="more more-link">查看更多</text>
			</view>
			<scroll-view scroll-x class="hot-scroll">
				<view class="hot-product-item" v-for="(product, index) in hotProducts" :key="index"
					@click="goToProductDetail(product.id)">
					<!-- <image :src="product.images[0]" class="hot-product-image" mode="aspectFill"></image> -->
					<image v-if="product.images && product.images.length > 0" :src="product.images[0]"
						class="hot-product-image" mode="aspectFill"></image>
					<image v-else src="/static/icon/nav_icon4.png" class="hot-product-image" mode="aspectFill"></image>
					<view class="hot-product-price">
						<view class="price-pill">
							<!-- <text class="price-icon">⚡</text> -->
							<text class="price-symbol">¥</text>
							<text class="price-number">{{ formatHotPrice(product.displayPrice || product.price)
								}}</text>
						</view>
					</view>
				</view>
			</scroll-view>
		</view>

		<!-- 限时特价 -->
		<view class="special-offers">
			<view class="section-title special-section-title">
				<view class="section-left">
					<view class="section-title-text">
						<text class="section-name">精选推荐</text>
					</view>
				</view>
				<text class="more-link special-more">更多好物</text>
			</view>
			<view class="special-product-list">
				<view class="special-product-item" v-for="(product, index) in specialProducts" :key="index"
					@click="goToProductDetail(product.id)">
					<view class="special-product-image-wrapper">
						<image :src="product.images[0]" class="special-product-image" mode="aspectFill"></image>
					</view>
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

		<!-- 测试登录按钮 -->
	</view>

	<ProductSelector ref="productSelector" />

	<!-- 用户编号提示弹窗 -->
	<view class="user-code-modal-overlay" v-if="showUserCodeModal" @click.stop>
		<view class="user-code-modal-content" @click.stop>
			<view class="user-code-success-header">
				<view class="user-code-success-icon-wrapper">
					<uni-icons type="checkmarkempty" size="60" color="rgba(255, 255, 255, 0.8)"></uni-icons>
				</view>
				<text class="user-code-success-title">登录成功</text>
			</view>
			<view class="user-code-success-body">
				<view class="user-code-section">
					<text class="user-code-label">您的用户编号</text>
					<view class="user-code-display" @click="copyUserCode">
						<text class="user-code-text">{{ currentUserCode || '暂无' }}</text>
						<uni-icons type="copy" size="20" color="#20CB6B" class="copy-icon"></uni-icons>
					</view>
				</view>
				<view class="tip-section">
					<text class="tip-text">请把你的编号告诉业务员，便于帮助您完善信息</text>
				</view>
			</view>
			<view class="user-code-success-footer">
				<view class="user-code-btn cancel-btn" @click="handleUserCodeModalCancel">
					<text class="user-code-btn-text">自己填写</text>
				</view>
				<view class="user-code-btn confirm-btn" @click="handleUserCodeModalConfirm">
					<text class="user-code-btn-text">我知道了</text>
				</view>
			</view>
		</view>
	</view>
</template>

<script>
import { getCarousels, getCategories, getSpecialProducts, getHotProducts, getMiniUserInfo, miniLogin } from '../../api/index';
import ProductSelector from '../../components/ProductSelector.vue';
import { getShareConfig, buildSharePath } from '../../utils/shareConfig.js';
import { updatePurchaseListTabBarBadge } from '../../utils/purchaseList.js';
export default {
	components: {
		ProductSelector
	},
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
			isAutoLogging: false, // 是否正在自动登录
			showUserCodeModal: false, // 是否显示用户编号提示弹窗
			currentUserCode: '' // 当前用户编号
		};
	},
	onLoad() {
		// 初始化用户类型
		this.initUserType();

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

		// 启用分享功能
		uni.showShareMenu({
			withShareTicket: true,
			menus: ['shareAppMessage', 'shareTimeline']
		});

		// 检查并自动登录
		this.checkAndAutoLogin();
	},
	// 页面显示时更新用户信息
	onShow() {
		this.updateUserInfo();
		updatePurchaseListTabBarBadge();
		// 如果弹窗正在显示，检查用户是否已经完善了资料
		if (this.showUserCodeModal) {
			const userInfo = uni.getStorageSync('miniUserInfo');
			const profileCompleted = userInfo && (userInfo.profile_completed || userInfo.profileCompleted);
			if (profileCompleted) {
				// 如果已经完善了资料，关闭弹窗
				this.showUserCodeModal = false;
			}
		}
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
	// 分享小程序
	onShareAppMessage(options) {
		// 使用 shareConfig 获取分享配置
		const shareConfig = getShareConfig('index');

		// 构建分享路径，添加分享者ID
		const path = buildSharePath('/pages/index/index');

		// 确保 imageUrl 正确传递（如果配置中有值就使用，否则使用空字符串）
		const shareImageUrl = shareConfig.imageUrl ? shareConfig.imageUrl : '';

		// 调试信息
		console.log('首页分享配置:', shareConfig);
		console.log('分享路径:', path);
		console.log('分享图片URL:', shareImageUrl);

		return {
			title: shareConfig.title,
			path: path,
			imageUrl: shareImageUrl
		};
	},

	methods: {
		// 检查并自动登录
		async checkAndAutoLogin() {
			// 检查是否已登录
			const token = uni.getStorageSync('miniUserToken');
			const userInfo = uni.getStorageSync('miniUserInfo');
			const uniqueId = uni.getStorageSync('miniUserUniqueId');

			// 如果已有完整的登录信息，不需要重新登录
			if (token && userInfo && uniqueId) {
				return;
			}

			// 如果正在登录中，避免重复触发
			if (this.isAutoLogging) {
				return;
			}

			// 延迟一点执行，让页面先加载完成
			setTimeout(async () => {
				try {
					this.isAutoLogging = true;
					await this.performAutoLogin();
				} catch (error) {
					console.error('自动登录失败:', error);
					// 静默失败，不打扰用户
				} finally {
					this.isAutoLogging = false;
				}
			}, 500);
		},

		// 执行自动登录
		async performAutoLogin() {
			uni.showLoading({
				title: '加载中...',
				mask: true
			});

			try {
				// 调用微信登录
				const loginRes = await new Promise((resolve, reject) => {
					uni.login({
						provider: 'weixin',
						success: resolve,
						fail: reject
					});
				});

				if (!loginRes || !loginRes.code) {
					throw new Error('未获取到登录凭证');
				}

				// 获取本地存储的分享者ID
				const shareReferrerId = uni.getStorageSync('shareReferrerId');
				let referrerId = null;
				if (shareReferrerId) {
					const id = parseInt(shareReferrerId);
					if (!isNaN(id) && id > 0) {
						referrerId = id;
					}
				}

				// 调用登录API
				const resp = await miniLogin(loginRes.code, referrerId);
				const data = resp?.data || {};
				const user = data.user || {};
				const token = data.token || '';
				const uniqueId = user.unique_id || user.uniqueId;

				if (!uniqueId) {
					throw new Error('未返回用户唯一ID');
				}

				// 登录成功后，清除分享者ID（只绑定一次）
				if (referrerId) {
					uni.removeStorageSync('shareReferrerId');
				}

				// 保存用户信息
				if (user) {
					uni.setStorageSync('miniUserInfo', user);
					if (uniqueId) {
						uni.setStorageSync('miniUserUniqueId', uniqueId);
					}
				}

				if (token) {
					uni.setStorageSync('miniUserToken', token);
				}

				// 更新用户类型
				this.userType = user.user_type || null;

				// 重新计算产品价格
				this.recalculateAllPrices();

				// 保存用户编号
				const userCode = user.user_code || user.userCode || '';
				this.currentUserCode = userCode;

				// 检查是否是新用户（未完善资料）
				const profileCompleted = user.profile_completed || user.profileCompleted || false;
				if (!profileCompleted && userCode) {
					// 延迟显示用户编号提示弹窗，让登录提示先消失
					setTimeout(() => {
						this.showUserCodeModal = true;
					}, 300);
				}
			} catch (error) {
				console.error('自动登录失败:', error);
				// 静默失败，不显示错误提示，避免打扰用户
			} finally {
				uni.hideLoading();
			}
		},

		// 处理用户编号弹窗 - 自己填写（灰色按钮）
		handleUserCodeModalCancel() {
			this.showUserCodeModal = false;
			// 跳转到资料填写页面
			uni.navigateTo({
				url: '/pages/profile/form'
			});
		},

		// 处理用户编号弹窗 - 我知道了（绿色按钮）
		async handleUserCodeModalConfirm() {
			// 先复制用户编号
			await this.copyUserCode();
			this.showUserCodeModal = false;
		},

		// 复制用户编号
		copyUserCode() {
			return new Promise((resolve, reject) => {
				if (!this.currentUserCode) {
					uni.showToast({
						title: '用户编号为空',
						icon: 'none'
					});
					reject(new Error('用户编号为空'));
					return;
				}

				uni.setClipboardData({
					data: this.currentUserCode,
					success: () => {
						resolve();
					},
					fail: () => {
						reject(new Error('复制失败'));
					}
				});
			});
		},

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
			// 重新计算热销产品价格
			if (this.hotProducts && this.hotProducts.length > 0) {
				this.hotProducts.forEach(product => {
					this.calculateProductPriceRange(product);
				});
			}
			// 重新计算特价产品价格
			if (this.specialProducts && this.specialProducts.length > 0) {
				this.specialProducts.forEach(product => {
					this.calculateProductPriceRange(product);
				});
			}
			// 重新计算分类产品价格
			if (this.sections && this.sections.length > 0) {
				this.sections.forEach(section => {
					if (section.products && section.products.length > 0) {
						section.products.forEach(product => {
							this.calculateProductPriceRange(product);
						});
					}
				});
			}
		},

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

		// 计算单个商品的最低价格（使用批发价和零售价）
		calculateProductPriceRange(product) {
			if (!product.specs || !Array.isArray(product.specs) || product.specs.length === 0) {
				// 如果没有规格数据，使用原价格
				product.displayPrice = product.price || '0.00';
				return;
			}

			// 根据用户类型决定显示哪种价格
			const isWholesaleUser = this.userType === 'wholesale';

			// 收集价格
			const prices = [];
			product.specs.forEach(spec => {
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

			if (prices.length === 0) {
				// 如果没有找到对应类型的价格，尝试使用另一种价格作为后备
				product.specs.forEach(spec => {
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
				});
			}

			if (prices.length === 0) {
				product.displayPrice = product.price || '0.00';
				return;
			}

			// 计算最低价格
			const minPrice = Math.min(...prices);
			product.displayPrice = minPrice.toFixed(2);
		},

		// 热销价格展示格式化（控制小数位）
		formatHotPrice(price) {
			if (price === undefined || price === null || price === '') {
				return '0.0';
			}

			const num = Number(price);
			if (Number.isNaN(num)) {
				return price;
			}

			const absNum = Math.abs(num);

			if (absNum < 100) {
				return num.toFixed(2);
			}

			return num.toFixed(1);
		},

		// 加载热销商品
		async loadHotProducts() {
			try {
				const res = await getHotProducts();
				if (res.code === 200) {
					this.hotProducts = Array.isArray(res.data) ? res.data : [];
					// 为每个商品处理数据并计算价格
					this.hotProducts.forEach(product => {
						// 处理数据结构差异，确保有必要的字段
						if (!product.images || !Array.isArray(product.images)) {
							product.images = product.image ? [product.image] : [];
						}
						// 计算价格
						this.calculateProductPriceRange(product);
					});
				} else {
					console.error('获取热销商品失败:', res.message || '未知错误');
					this.hotProducts = [];
				}
			} catch (error) {
				console.error('加载热销商品时发生异常:', error);
				this.hotProducts = [];
			}
		},

		// 加载各分类区块商品
		loadSectionProducts() {
			// 模拟数据
			this.sections.forEach(section => {
				section.products = [
					{ id: 5, images: ['https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'], price: '128~298' },
					{ id: 6, images: ['https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'], price: '128~298' }
				];
			});
		},

		// 导航到指定链接
		navigateTo(link) {
			console.log('navigateTo', link);

			if (!link || link.trim() === '') {
				return;
			}

			// 处理外部链接（http:// 或 https://）
			if (link.startsWith('http://') || link.startsWith('https://')) {
				// #ifdef H5
				window.open(link, '_blank');
				// #endif
				// #ifndef H5
				uni.showToast({
					title: '外部链接暂不支持',
					icon: 'none'
				});
				// #endif
				return;
			}

			// 处理完整的小程序路径（以 /pages/ 开头）
			if (link.startsWith('/pages/')) {
				// 判断是否是 tabBar 页面
				const tabBarPages = ['/pages/index/index', '/pages/category/category', '/pages/cart/cart', '/pages/my/my'];
				if (tabBarPages.includes(link.split('?')[0])) {
					uni.switchTab({
						url: link.split('?')[0]
					});
				} else {
					uni.navigateTo({
						url: link
					});
				}
				return;
			}

			// 处理商品详情页：product/xxx 或 product?id=xxx
			if (link.startsWith('product/')) {
				const productId = link.split('/')[1];
				uni.navigateTo({
					url: '/pages/product/detail?id=' + productId
				});
				return;
			}

			// 处理分类页面：category/xxx 或 category?id=xxx
			if (link.startsWith('category/')) {
				const categoryId = link.split('/')[1];
				// 使用globalData传递分类ID
				getApp().globalData.targetCategoryId = categoryId;
				uni.switchTab({
					url: '/pages/category/category'
				});
				return;
			}

			// 处理富文本页面：rich-content/xxx 或 rich-content?id=xxx
			if (link.startsWith('rich-content/')) {
				const contentId = link.split('/')[1];
				uni.navigateTo({
					url: '/pages/rich-content/rich-content?id=' + contentId
				});
				return;
			}

			// 处理带查询参数的格式：page?key=value
			if (link.includes('?')) {
				const [page, params] = link.split('?');
				// 尝试匹配已知的页面路径
				if (page === 'product' || page === 'product/detail') {
					const idMatch = params.match(/id=(\d+)/);
					if (idMatch) {
						uni.navigateTo({
							url: '/pages/product/detail?id=' + idMatch[1]
						});
						return;
					}
				} else if (page === 'category') {
					const idMatch = params.match(/id=(\d+)/);
					if (idMatch) {
						getApp().globalData.targetCategoryId = idMatch[1];
						uni.switchTab({
							url: '/pages/category/category'
						});
						return;
					}
				} else if (page === 'rich-content') {
					const idMatch = params.match(/id=(\d+)/);
					if (idMatch) {
						uni.navigateTo({
							url: '/pages/rich-content/rich-content?id=' + idMatch[1]
						});
						return;
					}
				}
			}

			// 如果都不匹配，尝试作为完整路径处理
			if (link.startsWith('/')) {
				uni.navigateTo({
					url: link
				});
			} else {
				// 未知格式，提示用户
				console.warn('未知的跳转链接格式:', link);
				uni.showToast({
					title: '链接格式不正确',
					icon: 'none'
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
		onAddBtnClick(product) {
			this.$refs.productSelector?.open(product);
		},

		// 测试登录按钮
	}
};
</script>

<style>
/* 自定义头部样式 - 按照参考文章实现 */
.custom-header {
	position: relative;
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

.special-section-title {
	padding: 10rpx 20rpx;
	/* background: linear-gradient(90deg, rgba(32, 203, 107, 0.08), rgba(32, 203, 107, 0.02)); */
	border-radius: 16rpx;
	/* border: 1rpx solid rgba(32, 203, 107, 0.15); */
}

.hot-section-title {
	padding: 10rpx 20rpx;
	/* background: linear-gradient(90deg, rgba(32, 203, 107, 0.12), rgba(32, 203, 107, 0.02)); */
	border-radius: 16rpx;
	/* box-shadow: 0 10rpx 20rpx rgba(32, 203, 107, 0.12); */
}

.section-left {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.hot-tag {
	background: linear-gradient(135deg, #20CB6B, #10b05a);
	color: #fff;
	font-size: 22rpx;
	padding: 4rpx 14rpx;
	border-radius: 40rpx;
	font-weight: 600;
	letter-spacing: 1rpx;
	margin-right: 6rpx;
	box-shadow: 0 6rpx 16rpx rgba(32, 203, 107, 0.2);
}

.section-name {
	font-size: 32rpx;
	font-weight: bold;
	color: #333;
}

.section-subtitle {
	font-size: 24rpx;
	color: #999;
	margin-top: 4rpx;
}

.special-tag {
	background: linear-gradient(135deg, #20CB6B, #12a458);
	color: #fff;
	font-size: 20rpx;
	padding: 4rpx 14rpx;
	border-radius: 40rpx;
	font-weight: 600;
	letter-spacing: 1rpx;
	box-shadow: 0 6rpx 16rpx rgba(32, 203, 107, 0.2);
}

.special-subtitle {
	color: #4f9c72;
}

.more {
	font-size: 26rpx;
	color: #999;
}

.more-link {
	color: #20CB6B;
	font-weight: 500;
	display: flex;
	align-items: center;
	gap: 6rpx;
}

.special-more {
	color: #20CB6B;
	font-weight: 600;
}

.hot-scroll {
	width: 100%;
	/* height: 300rpx; */
	overflow: hidden;
	white-space: nowrap
}

.hot-scroll::-webkit-scrollbar {
	display: none;
}

.hot-product-item {
	width: 160rpx;
	margin-right: 16rpx;
	display: inline-block;
	text-align: center;
	transition: transform 0.2s;
}

.hot-product-image {
	width: 160rpx;
	height: 160rpx;
	border-radius: 10rpx;
}

.hot-product-price {
	display: flex;
	align-items: center;
	justify-content: center;
	margin-top: 10rpx;
}

.price-pill {
	display: inline-flex;
	align-items: center;
	justify-content: flex-end;
	box-sizing: border-box;
	/* padding: 0 20rpx 0 30rpx; */
	padding-right: 10rpx;
	padding-top: 5rpx;
	min-width: 120rpx;
	height: 48rpx;
	border-radius: 36rpx;
	background-image: url('/static/icon/hot_icon.png');
	background-size: 100% 100%;
	background-repeat: no-repeat;
	background-position: center;
	color: #fff;
	box-shadow: none;
	line-height: 48rpx;
}

.price-icon {
	font-size: 20rpx;
	margin-right: 4rpx;
	color: #fff;
}

.price-symbol,
.price-number {
	color: #fff;
	font-weight: bold;
	padding-top: 3rpx;
}

.price-symbol {
	font-size: 20rpx;
	margin-right: 2rpx;
}

.price-number {
	font-size: 24rpx;
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
	/* background-color: #fff; */
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
	width: 49%;
	background-color: #fff;
	border-radius: 12rpx;
	overflow: hidden;
	margin-bottom: 20rpx;
	box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.05);
	display: flex;
	flex-direction: column;
}

.special-product-image-wrapper {
	width: 100%;
	position: relative;
	padding-top: 100%;
	border-top-left-radius: 12rpx;
	border-top-right-radius: 12rpx;
	overflow: hidden;
}

.special-product-image {
	position: absolute;
	top: 0;
	left: 0;
	width: 100%;
	height: 100%;
}

.special-product-info {
	padding: 20rpx;
	display: flex;
	flex-direction: column;
	position: relative;
	flex: 1;
}

.product-name {
	font-size: 28rpx;
	color: #333;
	line-height: 38rpx;
	display: -webkit-box;
	line-clamp: 2;
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

/* 用户编号提示弹窗样式 */
.user-code-modal-overlay {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background-color: rgba(0, 0, 0, 0.5);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 9999;
	animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
	from {
		opacity: 0;
	}
	to {
		opacity: 1;
	}
}

.user-code-modal-content {
	width: 640rpx;
	background-color: #fff;
	border-radius: 24rpx;
	overflow: hidden;
	box-shadow: 0 8rpx 32rpx rgba(0, 0, 0, 0.12);
	animation: slideUp 0.3s ease;
}

@keyframes slideUp {
	from {
		transform: translateY(50rpx);
		opacity: 0;
	}
	to {
		transform: translateY(0);
		opacity: 1;
	}
}

/* 成功头部 */
.user-code-success-header {
	padding: 60rpx 30rpx 0 30rpx;
	text-align: center;
	background: linear-gradient(180deg, #E8F8F0 0%, #fff 100%);
}

.user-code-success-icon-wrapper {
	width: 140rpx;
	height: 140rpx;
	margin: 0 auto 30rpx;
	background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 6rpx 20rpx rgba(32, 203, 107, 0.4);
	animation: scaleIn 0.4s ease;
}

@keyframes scaleIn {
	from {
		transform: scale(0);
		opacity: 0;
	}
	to {
		transform: scale(1);
		opacity: 1;
	}
}

.user-code-success-title {
	font-size: 44rpx;
	font-weight: 600;
	color: #20CB6B;
	display: block;
}

/* 成功主体 */
.user-code-success-body {
	padding: 50rpx 40rpx;
}

.user-code-section {
	margin-bottom: 40rpx;
}

.user-code-label {
	font-size: 28rpx;
	color: #999;
	display: block;
	text-align: center;
	margin-bottom: 24rpx;
}

.user-code-display {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 20rpx 40rpx;
	background: linear-gradient(135deg, #E8F8F0 0%, #F0FBF5 100%);
	border: 2rpx solid #20CB6B;
	border-radius: 16rpx;
	transition: all 0.3s;
}

.user-code-display:active {
	background: linear-gradient(135deg, #D8F5E8 0%, #E8F8F0 100%);
	transform: scale(0.98);
	box-shadow: 0 4rpx 12rpx rgba(32, 203, 107, 0.2);
}

.user-code-text {
	font-size: 64rpx;
	font-weight: 700;
	color: #20CB6B;
	letter-spacing: 4rpx;
	font-family: 'Courier New', monospace;
	flex: 1;
	text-align: center;
	line-height: 1.4;
}

.copy-icon {
	flex-shrink: 0;
	opacity: 0.7;
	transition: opacity 0.3s;
}

.user-code-display:active .copy-icon {
	opacity: 1;
}

.tip-section {
	padding: 20rpx 0;
	text-align: center;
}

.tip-text {
	font-size: 24rpx;
	color: #999;
	line-height: 1.6;
	display: block;
}

/* 成功底部按钮 */
.user-code-success-footer {
	display: flex;
	border-top: 1rpx solid #f0f0f0;
}

.user-code-btn {
	flex: 1;
	height: 100rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 32rpx;
	transition: all 0.2s;
}

.user-code-btn:active {
	opacity: 0.7;
}

.user-code-btn.cancel-btn {
	color: #666;
	background-color: #f5f5f5;
	border-right: 1rpx solid #f0f0f0;
}

.user-code-btn.confirm-btn {
	color: #fff;
	background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
	font-weight: 600;
}

.user-code-btn-text {
	font-size: 32rpx;
}

.add-btn {
	width: 32px;
	height: 32px;
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
</style>
