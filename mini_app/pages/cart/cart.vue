<template>
	<view class="cart-page">
		<view class="cart-header">
			<view :style="{ height: statusBarHeight + 'px' }"></view>
			<view class="cart-header-content" :style="{ height: navBarHeight + 'px', paddingRight: menuRightPadding }">
				<view class="header-left">
					<text class="nav-title">我的采购单</text>
					<text class="nav-subtitle">共 {{ cartItems.length }} 款商品</text>
				</view>
				<view class="header-right">
				</view>
			</view>
		</view>
		<view class="cart-body">
			<view class="tabs">
				<view class="tabs-left">
					<view class="tab ">全部产品</view>
					<view class="tab_edit" :class="{ active: isEditing }" @click="toggleEdit">{{ isEditing ? '完成' : '编辑'
						}}</view>
					<view class="tab_delete" v-if="isEditing" @click="batchDelete">批量删除</view>
				</view>
				<view class="tab tab-frequent" @click="goToFrequent">我常买</view>
			</view>

			<view class="cart-list" v-if="cartItems.length > 0">
				<view class="cart-item" v-for="item in cartItems" :key="item.id">
					<view class="item-select" @click.stop="toggleSelect(item)">
						<view :class="['select-dot', { active: selectedIds.includes(item.id) }]"></view>
					</view>
					<image :src="item.product_image || defaultImage" class="item-image" mode="aspectFill" @click="goToProductDetail(item.product_id)"></image>
					<view class="item-info" @click="goToProductDetail(item.product_id)">
						<view class="item-title-row">
							<text class="item-name">{{ item.product_name }}</text>
							<text class="item-price">¥{{ getDisplayPrice(item).toFixed(2) }}</text>
						</view>
						<text class="item-spec" v-if="item.spec_name">{{ item.spec_name }}</text>
						<view class="blocked-badge" v-if="isItemBlocked(item) && actualDeliveryFee > 0">
							<text class="icon">!</text>
							<text>该商品不参与免配送费</text>
						</view>
						<text class="item-tags">支持采购 · 现货供应</text>
						<view class="item-actions">
							<view class="qty-control" v-if="!isEditing">
								<view class="qty-btn minus" @click.stop="decreaseItem(item)">
									<image src="/static/icon/minus.png" class="qty-icon" />
								</view>
								<text class="qty-value">{{ item.quantity }}</text>
								<view class="qty-btn plus" @click.stop="increaseItem(item)">
									<uni-icons type="plusempty" size="16" color="#fff"></uni-icons>
								</view>
							</view>
							<view class="item-delete" v-if="isEditing" @click.stop="deleteItem(item)">
								<uni-icons type="trash" size="22" color="#fff"></uni-icons>
							</view>
						</view>
					</view>
				</view>
			</view>

			<!-- 空采购单提示 -->
			<view class="empty-cart" v-else>
				<image src="/static/empty-cart.png" class="empty-icon"></image>
				<text class="empty-text">您的采购单还是空的</text>
				<text class="empty-subtext">快去选购商品吧~</text>
			</view>

			<!-- 配送费信息 -->
			<view class="assistant-card" v-if="cartItems.length > 0 && deliverySummary && !showFeeDetail"
				@touchmove.stop.prevent>
				<view class="assistant-row">
					<view class="assistant-left">
						<text class="assistant-title"><text class="title-red">凑单</text><text
								class="title-black">助手</text></text>
						<text class="assistant-divider">|</text>
						<text class="assistant-hint"
							v-if="!deliverySummary.is_free_shipping && !selectedDeliveryFeeCoupon">
							还差<text class="assistant-amount">{{ shortOfAmount }}元</text>免基础配送费
						</text>
						<text class="assistant-hint free" v-else>您已享受基础配送费优惠</text>
					</view>
					<view class="assistant-btn" @click.stop="goToIndex">
						<text>{{ (deliverySummary.is_free_shipping || selectedDeliveryFeeCoupon) ? '再看看' : '去凑单'
							}}</text>
					</view>
				</view>
			</view>

			<view class="bottom-bar" v-if="cartItems.length > 0" @click="openFeeDetail" @touchmove.stop.prevent>
				<view class="bottom-left">
					<view class="select-all" @click.stop="selectAll">
						<!-- <text class="selected-count">共{{ selectedQuantity }} 件</text> -->
						<view class="select-main">
							<view
								:class="['select-dot', { active: selectedIds.length === cartItems.length && cartItems.length > 0 }]">
							</view>
							<text>全选</text>
						</view>
					</view>
					<view class="bottom-total">
						<text class="bottom-amount">¥{{ finalAmount }}</text>
						<text class="bottom-discount" v-if="actualDeliveryFee > 0">含配送费 ¥{{ actualDeliveryFee.toFixed(2)
							}}</text>
						<text class="bottom-discount-amount" v-else-if="totalDiscount > 0">已优惠 ¥{{ totalDiscount
							}}</text>
					</view>
				</view>
				<button class="checkout-btn" @click.stop="goCheckout">去下单</button>
			</view>
		</view>

		<view v-if="showFeeDetail" class="fee-modal-container">
			<view class="fee-modal-mask" @click="closeFeeDetail"></view>
			<view class="fee-modal">
				<view class="fee-modal-header">
					<text>费用详情</text>
					<text class="fee-modal-close" @click="closeFeeDetail">×</text>
				</view>
				<view class="fee-modal-body">
					<scroll-view scroll-y style="max-height: 400rpx; margin-bottom: 20rpx;"
						v-if="selectedCartItems.length">
						<view class="fee-item" v-for="item in selectedCartItems" :key="item.id">
							<view class="fee-item-info">
								<text class="fee-item-name">{{ item.product_name }}</text>
								<text class="fee-item-spec" v-if="item.spec_name">{{ item.spec_name }}</text>
							</view>
							<view class="fee-item-amount">
								<text class="fee-item-price">¥{{ getDisplayPrice(item).toFixed(2) }} × {{ item.quantity
									}}</text>
								<text class="fee-item-total">¥{{ getItemTotal(item) }}</text>
							</view>
						</view>
					</scroll-view>
					<view class="fee-row" v-if="deliverySummary">
						<text>配送费</text>
						<text>{{ actualDeliveryFeeText }}</text>
					</view>
					<!-- 优惠券选择 -->
					<view class="coupon-section" v-if="availableCoupons.length > 0">
						<view class="coupon-row" v-if="availableDeliveryFeeCoupons.length > 0">
							<text class="coupon-label">免配送费券</text>
							<view class="coupon-selector" @click="showCouponSelector = true">
								<text v-if="selectedDeliveryFeeCoupon" class="coupon-selected">
									{{ formatCouponName(selectedDeliveryFeeCoupon) }}
								</text>
								<text v-else class="coupon-placeholder">选择优惠券</text>
								<text class="coupon-change">切换</text>
							</view>
						</view>
						<view class="coupon-row" v-if="availableAmountCoupons.length > 0">
							<text class="coupon-label">金额券</text>
							<view class="coupon-selector" @click="showCouponSelector = true">
								<text v-if="selectedAmountCoupon" class="coupon-selected">
									{{ formatCouponName(selectedAmountCoupon) }} (减¥{{
										formatCouponDiscount(selectedAmountCoupon) }})
								</text>
								<text v-else class="coupon-placeholder">选择优惠券</text>
								<text class="coupon-change">切换</text>
							</view>
						</view>
					</view>
					<view class="fee-row" v-if="totalDiscount > 0">
						<text>优惠</text>
						<text class="discount">- ¥{{ totalDiscount }}</text>
					</view>
					<view class="fee-row total">
						<text>合计</text>
						<text>¥{{ finalAmount }}</text>
					</view>
				</view>
			</view>
		</view>

		<!-- 优惠券选择弹窗 -->
		<view v-if="showCouponSelector" class="coupon-modal-container">
			<view class="coupon-modal-mask" @click="showCouponSelector = false"></view>
			<view class="coupon-modal">
				<view class="coupon-modal-header">
					<text>选择优惠券</text>
					<text class="coupon-modal-close" @click="showCouponSelector = false">×</text>
				</view>
				<view class="coupon-modal-body">
					<!-- 免配送费券 -->
					<view class="coupon-type-section" v-if="availableDeliveryFeeCoupons.length > 0">
						<text class="coupon-type-title">免配送费券</text>
						<view class="coupon-list">
							<view class="coupon-option" :class="{ active: !selectedDeliveryFeeCoupon }"
								@click="selectDeliveryFeeCoupon(null)">
								<text>不使用</text>
							</view>
							<view class="coupon-option" v-for="coupon in availableDeliveryFeeCoupons"
								:key="coupon.user_coupon_id"
								:class="{ active: selectedDeliveryFeeCoupon?.user_coupon_id === coupon.user_coupon_id }"
								@click="selectDeliveryFeeCoupon(coupon)">
								<text class="coupon-option-name">{{ coupon.name }}</text>
								<text class="coupon-option-desc" v-if="coupon.reason">{{ coupon.reason }}</text>
							</view>
						</view>
					</view>
					<!-- 金额券 -->
					<view class="coupon-type-section" v-if="availableAmountCoupons.length > 0">
						<text class="coupon-type-title">金额券</text>
						<view class="coupon-list">
							<view class="coupon-option" :class="{ active: !selectedAmountCoupon }"
								@click="selectAmountCoupon(null)">
								<text>不使用</text>
							</view>
							<view class="coupon-option" v-for="coupon in availableAmountCoupons"
								:key="coupon.user_coupon_id"
								:class="{ active: selectedAmountCoupon?.user_coupon_id === coupon.user_coupon_id, disabled: !coupon.is_available }"
								@click="coupon.is_available && selectAmountCoupon(coupon)">
								<view class="coupon-option-content">
									<text class="coupon-option-name">{{ coupon.name }}</text>
									<text class="coupon-option-value">减¥{{ coupon.discount_value.toFixed(2) }}</text>
									<text class="coupon-option-condition" v-if="coupon.min_amount > 0">
										满¥{{ coupon.min_amount.toFixed(2) }}可用
									</text>
									<text class="coupon-option-condition" v-else>无门槛</text>
								</view>
								<text class="coupon-option-reason" v-if="coupon.reason">{{ coupon.reason }}</text>
							</view>
						</view>
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script>
import { fetchPurchaseList, deletePurchaseListItemById, clearPurchaseListByToken, updatePurchaseListQuantity } from '../../utils/purchaseList'

export default {
	data() {
		return {
			cartItems: [],
			deliverySummary: null,
			blockedItemIds: [],
			showFeeDetail: false,
			selectedDiscount: '0.00',
			userType: 'unknown',
			loading: false,
			token: '',
			defaultImage: '/static/empty-cart.png',
			statusBarHeight: 0,
			navBarHeight: 44,
			menuButtonRect: null,
			selectedIds: [],
			isEditing: false,
			availableCoupons: [],
			selectedDeliveryFeeCoupon: null, // 选中的免配送费券
			selectedAmountCoupon: null, // 选中的金额券
			showCouponSelector: false // 是否显示优惠券选择器
		};
	},
	onLoad() {
		const info = uni.getSystemInfoSync()
		const menuButton = uni.getMenuButtonBoundingClientRect ? uni.getMenuButtonBoundingClientRect() : null
		this.statusBarHeight = info.statusBarHeight || 0
		if (menuButton) {
			this.navBarHeight = menuButton.height + (menuButton.top - this.statusBarHeight) * 2
			this.menuButtonRect = menuButton
		}
	},
	onShow() {
		this.userType = uni.getStorageSync('miniUserInfo')?.user_type || 'unknown';
		this.loadCart();
	},
	watch: {
		// 监听选中商品变化，重新计算配送费
		selectedIds: {
			handler(newIds, oldIds) {
				// 如果选中商品发生变化，重新获取配送费摘要
				if (newIds && newIds.length > 0 && this.cartItems.length > 0) {
					this.updateDeliveryFeeForSelected()
				} else if (newIds && newIds.length === 0) {
					// 如果没有选中商品，清空配送费摘要
					this.deliverySummary = null
				}
			},
			immediate: false,
			deep: true
		}
	},
	computed: {
		totalQuantity() {
			return this.cartItems.reduce((total, item) => total + (item.quantity || 0), 0);
		},
		totalAmount() {
			return this.cartItems.reduce((total, item) => {
				return total + this.getDisplayPrice(item) * (item.quantity || 0)
			}, 0).toFixed(2)
		},
		selectedQuantity() {
			return this.cartItems
				.filter(item => this.selectedIds.includes(item.id))
				.reduce((total, item) => total + (item.quantity || 0), 0)
		},
		selectedAmount() {
			return this.cartItems
				.filter(item => this.selectedIds.includes(item.id))
				.reduce((total, item) => total + this.getDisplayPrice(item) * (item.quantity || 0), 0)
				.toFixed(2)
		},
		selectedCartItems() {
			return this.cartItems.filter(item => this.selectedIds.includes(item.id))
		},
		deliveryFeeText() {
			if (!this.deliverySummary) return '0.00'
			const fee = Number(this.deliverySummary.delivery_fee || 0)
			return fee.toFixed(2)
		},
		shortOfAmount() {
			if (!this.deliverySummary) return '0.00'
			const value = Number(this.deliverySummary.short_of_amount || 0)
			if (value <= 0) return '0.00'
			return value.toFixed(2)
		},
		freeShippingThresholdText() {
			if (!this.deliverySummary) return '0.00'
			return Number(this.deliverySummary.free_shipping_threshold || 0).toFixed(2)
		},
		deliveryTips() {
			return (this.deliverySummary && Array.isArray(this.deliverySummary.tips)) ? this.deliverySummary.tips : []
		},
		finalAmount() {
			const total = Number(this.selectedAmount || 0)
			const deliveryFee = Number(this.actualDeliveryFee || 0)
			const amountDiscount = this.amountCouponDiscountAmount
			return (total + deliveryFee - amountDiscount).toFixed(2)
		},
		// 可用免配送费券列表
		availableDeliveryFeeCoupons() {
			return this.availableCoupons.filter(c => c.type === 'delivery_fee')
		},
		// 可用金额券列表
		availableAmountCoupons() {
			return this.availableCoupons.filter(c => c.type === 'amount')
		},
		// 实际配送费（考虑免配送费券）
		actualDeliveryFee() {
			if (!this.deliverySummary) return 0
			const base = this.deliverySummary.is_free_shipping ? 0 : Number(this.deliverySummary.delivery_fee || 0)
			return Math.max(base - this.deliveryFeeSavedAmount, 0)
		},
		actualDeliveryFeeText() {
			if (this.deliverySummary?.is_free_shipping || this.selectedDeliveryFeeCoupon) {
				return '免配送费'
			}
			return '¥' + this.deliveryFeeText
		},
		// 计算总优惠金额
		deliveryFeeSavedAmount() {
			if (!this.deliverySummary || this.deliverySummary.is_free_shipping) {
				return 0
			}
			if (this.selectedDeliveryFeeCoupon) {
				return Number(this.deliverySummary.delivery_fee || 0)
			}
			return 0
		},
		amountCouponDiscountAmount() {
			if (this.selectedAmountCoupon) {
				return Number(this.selectedAmountCoupon.discount_value || 0)
			}
			return 0
		},
		totalDiscount() {
			return this.amountCouponDiscountAmount.toFixed(2)
		},
		menuRightPadding() {
			if (this.menuButtonRect) {
				const info = uni.getSystemInfoSync()
				const safeRight = info.windowWidth - this.menuButtonRect.right
				return `${this.menuButtonRect.width + safeRight + 16}px`
			}
			return '80rpx'
		}
	},
	methods: {
		async loadCart() {
			this.loading = true;
			try {
				this.token = uni.getStorageSync('miniUserToken') || '';
				if (!this.token) {
					this.cartItems = [];
					this.deliverySummary = null;
					this.blockedItemIds = [];
					return;
				}
				// 从本地存储读取之前保存的选中状态和已知商品ID
				const savedSelectedIds = uni.getStorageSync('cartSelectedIds') || [];
				const savedKnownIds = uni.getStorageSync('cartKnownIds') || [];
				const previousSelection = new Set(savedSelectedIds);
				const previousKnown = new Set(savedKnownIds);

				// 先获取所有商品列表（不传item_ids，获取全部）
				const { items } = await fetchPurchaseList(this.token);
				this.cartItems = items;
				this.blockedItemIds = [];
				if (items.length === 0) {
					this.selectedIds = [];
					this.saveSelectedIds([], []);
					this.deliverySummary = null;
					this.availableCoupons = [];
					this.selectedDeliveryFeeCoupon = null;
					this.selectedAmountCoupon = null;
				} else {
					// 获取当前商品ID集合
					const currentItemIds = items.map(item => item.id);
					// 找出新加入的商品（在当前列表中但不在之前已知的商品列表中）
					const newItemIds = items
						.filter(item => !previousKnown.has(item.id))
						.map(item => item.id);

					if (previousKnown.size > 0) {
						// 保留之前保存的选中状态（只保留仍存在的商品）
						const retained = items
							.filter(item => previousSelection.has(item.id))
							.map(item => item.id);
						// 保留之前的选中状态，并自动选中新加入的商品
						this.selectedIds = [...retained, ...newItemIds];
					} else {
						// 首次加载或没有之前的选中状态，自动选中所有商品
						this.selectedIds = items.map(item => item.id);
					}

					// 保存选中状态和已知商品ID
					this.saveSelectedIds(this.selectedIds, currentItemIds);

					// 根据选中的商品重新获取配送费摘要（包括优惠券信息）
					if (this.selectedIds.length > 0) {
						await this.updateDeliveryFeeForSelected();
					}
				}
			} catch (error) {
				console.error('加载采购单失败:', error);
				uni.showToast({
					title: '加载失败，请稍后再试',
					icon: 'none'
				});
			} finally {
				this.loading = false;
			}
		},
		saveSelectedIds(selectedIds, knownIds) {
			uni.setStorageSync('cartSelectedIds', selectedIds);
			if (knownIds !== undefined) {
				uni.setStorageSync('cartKnownIds', knownIds);
			}
		},
		getDisplayPrice(item) {
			const snapshot = item.spec_snapshot || {};
			const wholesale = Number(snapshot.wholesale_price || 0);
			const retail = Number(snapshot.retail_price || 0);
			if (this.userType === 'wholesale') {
				return wholesale || retail;
			}
			return retail || wholesale;
		},
		toggleSelect(item) {
			const index = this.selectedIds.indexOf(item.id)
			if (index > -1) {
				this.selectedIds.splice(index, 1)
			} else {
				this.selectedIds.push(item.id)
			}
			this.saveSelectedIds(this.selectedIds)
		},
		selectAll() {
			if (this.selectedIds.length === this.cartItems.length) {
				this.selectedIds = []
			} else {
				this.selectedIds = this.cartItems.map(item => item.id)
			}
			this.saveSelectedIds(this.selectedIds)
		},
		getItemTotal(item) {
			if (!item) return '0.00'
			const total = this.getDisplayPrice(item) * (item.quantity || 0)
			return total.toFixed(2)
		},
		goCheckout() {
			if (!this.cartItems.length) {
				uni.showToast({ title: '采购单为空', icon: 'none' })
				return
			}
			const selected = this.selectedIds && this.selectedIds.length ? this.selectedIds : []
			if (!selected.length) {
				uni.showToast({ title: '请选择要下单的商品', icon: 'none' })
				return
			}
			const query = [`item_ids=${selected.join(',')}`]
			if (this.selectedDeliveryFeeCoupon?.user_coupon_id) {
				query.push(`delivery_coupon_id=${this.selectedDeliveryFeeCoupon.user_coupon_id}`)
			}
			if (this.selectedAmountCoupon?.user_coupon_id) {
				query.push(`amount_coupon_id=${this.selectedAmountCoupon.user_coupon_id}`)
			}
			const url = `/pages/order/confirm?${query.join('&')}`
			uni.navigateTo({ url })
		},
		formatCouponName(coupon) {
			if (!coupon) return ''
			if (typeof coupon === 'string') return coupon
			return coupon.name || (coupon.coupon ? coupon.coupon.name : '')
		},
		formatCouponDiscount(coupon) {
			if (!coupon) return '0.00'
			const value = coupon.discount_value || (coupon.coupon ? coupon.coupon.discount_value : 0)
			return Number(value || 0).toFixed(2)
		},
		openFeeDetail() {
			if (!this.cartItems.length) return
			this.showFeeDetail = true
		},
		closeFeeDetail() {
			this.showFeeDetail = false
		},
		// 选择免配送费券
		selectDeliveryFeeCoupon(coupon) {
			this.selectedDeliveryFeeCoupon = coupon
			this.showCouponSelector = false
		},
		// 选择金额券
		selectAmountCoupon(coupon) {
			this.selectedAmountCoupon = coupon
			this.showCouponSelector = false
		},
		tipKey(tip) {
			if (!tip) return ''
			return `${tip.item_type || 'unknown'}-${tip.target_id || 0}`
		},
		isItemBlocked(item) {
			if (!item || !Array.isArray(this.blockedItemIds)) return false
			return this.blockedItemIds.includes(item.id)
		},
		formatTipName(tip) {
			if (!tip) return '特殊商品'
			if ((tip.item_type || '') === 'product') {
				return tip.target_name || tip.product_name || '指定商品'
			}
			return tip.target_name || '指定分类'
		},
		toggleEdit() {
			this.isEditing = !this.isEditing
		},
		goToFrequent() {
			uni.navigateTo({
				url: '/pages/frequent/frequent'
			});
		},
		goToIndex() {
			uni.switchTab({
				url: '/pages/index/index'
			});
		},
		goToProductDetail(productId) {
			if (!productId) {
				uni.showToast({
					title: '商品信息错误',
					icon: 'none'
				});
				return;
			}
			uni.navigateTo({
				url: `/pages/product/detail?id=${productId}`
			});
		},
		async increaseItem(item) {
			if (!item) return
			await this.handleQuantityChange(item, (item.quantity || 0) + 1)
		},
		async decreaseItem(item) {
			if (!item) return
			const next = (item.quantity || 0) - 1
			if (next <= 0) {
				this.deleteItem(item)
				return
			}
			await this.handleQuantityChange(item, next)
		},
		async handleQuantityChange(item, quantity) {
			if (!item || !this.token) return
			uni.vibrateShort({ type: 'light' })
			try {
				await updatePurchaseListQuantity({ token: this.token, itemId: item.id, quantity })
				await this.loadCart()
			} catch (error) {
				console.error('更新采购单数量失败:', error)
				uni.showToast({ title: '操作失败，请稍后再试', icon: 'none' })
			}
		},
		deleteItem(item) {
			if (!item) return;
			uni.showModal({
				title: '确认删除',
				content: '确定要从采购单中删除这个商品吗？',
				success: async (res) => {
					if (res.confirm) {
						try {
							await deletePurchaseListItemById({ token: this.token, itemId: item.id });
							uni.showToast({ title: '删除成功', icon: 'success' });
							this.loadCart();
						} catch (error) {
							console.error('删除采购单项失败:', error);
							uni.showToast({ title: '删除失败，请稍后再试', icon: 'none' });
						}
					}
				}
			});
		},
		batchDelete() {
			if (!this.selectedIds.length) {
				uni.showToast({ title: '请先选择商品', icon: 'none' });
				return;
			}
			uni.showModal({
				title: '确认删除',
				content: `确定要删除选中的 ${this.selectedIds.length} 个商品吗？`,
				success: async (res) => {
					if (res.confirm) {
						try {
							for (const itemId of this.selectedIds) {
								await deletePurchaseListItemById({ token: this.token, itemId });
							}
							uni.showToast({ title: '删除成功', icon: 'success' });
							this.selectedIds = [];
							this.saveSelectedIds([]);
							this.loadCart();
						} catch (error) {
							console.error('批量删除失败:', error);
							uni.showToast({ title: '删除失败，请稍后再试', icon: 'none' });
						}
					}
				}
			});
		},
		copyCart() {
			if (!this.cartItems.length) return;
			let copyText = '我的采购单：\n';
			this.cartItems.forEach((item, index) => {
				const price = this.getDisplayPrice(item);
				copyText += `${index + 1}. ${item.product_name}`;
				copyText += item.spec_name ? `（${item.spec_name}）` : '';
				if (price > 0) {
					copyText += ` - ¥${price.toFixed(2)}`;
				}
				copyText += ` (数量：${item.quantity})\n`;
			});
			uni.setClipboardData({
				data: copyText,
				success: () => {
					uni.showToast({
						title: '已复制到剪贴板',
						icon: 'success',
						duration: 2000
					});
				}
			});
		},
		clearCart() {
			if (!this.cartItems.length) return;
			uni.showModal({
				title: '确认清空',
				content: '确定要清空整个采购单吗？',
				success: async (res) => {
					if (res.confirm) {
						try {
							await clearPurchaseListByToken(this.token);
							uni.showToast({ title: '采购单已清空', icon: 'success' });
							this.loadCart();
						} catch (error) {
							console.error('清空采购单失败:', error);
							uni.showToast({ title: '清空失败，请稍后再试', icon: 'none' });
						}
					}
				}
			});
		},
		// 根据选中的商品更新配送费摘要
		async updateDeliveryFeeForSelected() {
			if (!this.token || !this.selectedIds || this.selectedIds.length === 0) {
				this.deliverySummary = null;
				return;
			}
			try {
				const { summary, availableCoupons, bestCombination } = await fetchPurchaseList(this.token, this.selectedIds);
				this.deliverySummary = summary || null;
				this.blockedItemIds = summary?.blocked_item_ids || [];
				// 更新可用优惠券（基于选中商品）
				this.availableCoupons = availableCoupons || [];
				// 自动应用最佳优惠券组合
				if (bestCombination) {
					this.selectedDeliveryFeeCoupon = bestCombination.delivery_fee_coupon || null;
					this.selectedAmountCoupon = bestCombination.amount_coupon || null;
				} else {
					// 如果最佳组合不适用，清空选中的优惠券
					this.selectedDeliveryFeeCoupon = null;
					this.selectedAmountCoupon = null;
				}
			} catch (error) {
				console.error('更新配送费失败:', error);
				// 失败时不更新，保持原有状态
			}
		}
	}
};
</script>

<style>
body {
	overflow: hidden;
}

.cart-page {
	height: 100vh;
	display: flex;
	flex-direction: column;
	background: linear-gradient(180deg, #e7fff3 0%, #f8f9fb 40%, #f5f6f8 100%);
	overflow: hidden;
}

.cart-body {
	flex: 1;
	display: flex;
	flex-direction: column;
	padding: 24rpx;
	box-sizing: border-box;
	min-height: 0;
	overflow: hidden;
}

.cart-header {
	/* background: #fff; */
	border-bottom-left-radius: 32rpx;
	border-bottom-right-radius: 32rpx;
	/* box-shadow: 0 20rpx 40rpx rgba(0, 0, 0, 0.08); */
	overflow: hidden;
	padding-bottom: 12rpx;
}

.cart-header-content {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0 24rpx;
}

.header-left {
	display: flex;
	flex-direction: row;
	align-items: flex-end;
	justify-content: flex-end;
}

.nav-title {
	font-size: 40rpx;
	font-weight: 600;
	color: #1a1a1a;
}

.nav-subtitle {
	font-size: 26rpx;
	color: #8f9aad;
	margin-left: 20rpx;
	padding-bottom: 4rpx;
}

.header-right {
	display: flex;
	align-items: center;
}

.header-icon {
	width: 64rpx;
	height: 64rpx;
	border-radius: 30rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(32, 203, 107, 0.12);
	box-shadow: inset 0 0 0 1px rgba(32, 203, 107, 0.25);
}

.tabs {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 24rpx;
}

.tabs-left {
	display: flex;
	gap: 16rpx;
}

.tab {
	padding: 12rpx 32rpx;
	border-radius: 999rpx;
	font-size: 26rpx;
	color: #666;
	background: #fff;
}

.tab_all {
	padding: 12rpx 32rpx;
}

.tab_edit {
	padding: 12rpx 32rpx;
	border-radius: 999rpx;
	font-size: 26rpx;
	color: #20cb6b;
	background: #fff;
	font-weight: 600;
}

.tab_edit.active {
	background: #20cb6b;
	color: #fff;
}

.tab_delete {
	padding: 12rpx 32rpx;
	border-radius: 999rpx;
	font-size: 26rpx;
	color: #fff;
	background: #ff4d4f;
	font-weight: 500;
}

.tab.active {
	background: #dff9ef;
}

.tab-frequent {
	background: transparent;
	color: #20CB6B;
	color: #20cb6b;
	font-weight: 600;
}

.cart-list {
	flex: 1;
	overflow-y: auto;
	padding-bottom: 260rpx;
}

.cart-item {
	display: flex;
	background: #fff;
	border-radius: 24rpx;
	padding: 20rpx;
	margin-bottom: 20rpx;
	box-shadow: 0 12rpx 24rpx rgba(24, 39, 75, 0.05);
}

.item-select {
	width: 40rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.select-dot {
	width: 32rpx;
	height: 32rpx;
	border-radius: 50%;
	border: 2rpx solid #dcdfe6;
	transition: all .2s;
}

.select-dot.active {
	background: linear-gradient(135deg, #20cb6b, #17b76b);
	border-color: transparent;
	box-shadow: 0 4rpx 8rpx rgba(32, 203, 107, 0.35);
	position: relative;
}

.select-dot.active::after {
	content: '';
	position: absolute;
	left: 50%;
	top: 50%;
	transform: translate(-50%, -50%);
	width: 40%;
	height: 40%;
	background-color: #fff;
	border-radius: 50%;
}

.item-image {
	width: 160rpx;
	height: 160rpx;
	border-radius: 20rpx;
	margin: 0 20rpx;
	background: #f6f7fb;
}

.item-info {
	flex: 1;
	display: flex;
	flex-direction: column;
}

.item-title-row {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
	margin-bottom: 10rpx;
}

.item-name {
	font-size: 30rpx;
	font-weight: 600;
	color: #1a1a1a;
	flex: 1;
	margin-right: 12rpx;
}

.item-price {
	font-size: 32rpx;
	color: #ff4d4f;
	font-weight: 700;
}

.item-spec {
	font-size: 26rpx;
	color: #909399;
	margin-bottom: 8rpx;
}

.blocked-badge {
	display: inline-flex;
	align-items: center;
	gap: 8rpx;
	padding: 6rpx 14rpx;
	border-radius: 999rpx;
	background: rgba(255, 77, 79, 0.12);
	color: #ff4d4f;
	font-size: 24rpx;
	margin-bottom: 8rpx;
}

.blocked-badge .icon {
	width: 28rpx;
	height: 28rpx;
	border-radius: 50%;
	background: rgba(255, 77, 79, 0.2);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 20rpx;
	font-weight: 600;
}

.item-tags {
	font-size: 24rpx;
	color: #20cb6b;
	padding: 6rpx 18rpx;
	background: #ddf8ed;
	border-radius: 999rpx;
	width: fit-content;
}

.item-actions {
	margin-top: auto;
	display: flex;
	justify-content: flex-end;
	align-items: center;
	gap: 20rpx;
}

.qty-control {
	display: flex;
	align-items: center;
	gap: 10rpx;
}

.qty-btn {
	width: 56rpx;
	height: 56rpx;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
}

.qty-btn.minus {
	background: #F7F8F9;
	/* border: 1rpx solid rgba(32, 203, 107, 0.3); */
}

.qty-btn.plus {
	background: #20CB6B;
}

.qty-icon {
	width: 28rpx;
	height: 28rpx;
}

.qty-value {
	font-size: 28rpx;
	font-weight: 600;
	color: #1a1a1a;
	min-width: 40rpx;
	text-align: center;
}

.item-delete {
	width: 56rpx;
	height: 56rpx;
	border-radius: 50%;
	background: #ff4d4f;
	display: flex;
	align-items: center;
	justify-content: center;
}

.empty-cart {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	color: #909399;
}

.empty-icon {
	width: 240rpx;
	height: 240rpx;
	margin-bottom: 24rpx;
	opacity: 0.6;
}

.assistant-card {
	position: fixed;
	left: 24rpx;
	right: 24rpx;
	bottom: 130rpx;
	background: #FFEFF0;
	border-radius: 16rpx 16rpx 0 0;
	padding: 16rpx 24rpx;
	z-index: 99;
	box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.06);
}

.assistant-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.assistant-left {
	display: flex;
	align-items: center;
	gap: 12rpx;
	flex: 1;
}

.assistant-title {
	font-size: 28rpx;
	font-weight: 700;
}

.title-red {
	color: #e74c3c;
}

.title-black {
	color: #333;
}

.assistant-divider {
	font-size: 24rpx;
	color: #ddd;
	padding: 0 10rpx;
}

.assistant-hint {
	font-size: 24rpx;
	color: #666;
}

.assistant-hint.free {
	color: #555;
}

.assistant-amount {
	color: #e74c3c;
	font-weight: 600;
}

.assistant-btn {
	padding: 8rpx 32rpx;
	border: 2rpx solid #e74c3c;
	border-radius: 999rpx;
	color: #e74c3c;
	font-size: 26rpx;
	font-weight: 500;
}

.bottom-bar {
	position: fixed;
	left: 0;
	bottom: 0;
	width: 100%;
	display: flex;
	align-items: center;
	justify-content: space-between;
	background: #fff;
	padding: 24rpx;
	box-sizing: border-box;
	box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.06);
	z-index: 100;
	border-bottom: 1rpx solid #F0F0F0;
}

.bottom-left {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 24rpx;
	flex: 1;
	padding-right: 20rpx;
}

.select-all {
	width: 100rpx;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	gap: 6rpx;
	font-size: 26rpx;
	color: #606266;
	height: 100%;
}

.select-main {
	display: flex;
	align-items: center;
	gap: 8rpx;
}

.selected-count {
	font-size: 22rpx;
	color: #909399;
	margin-left: 8rpx;
}

.bottom-total {
	display: flex;
	flex-direction: column;
	align-items: baseline;
	gap: 8rpx;
}

.bottom-label {
	font-size: 28rpx;
	color: #303133;
}

.bottom-amount {
	font-size: 40rpx;
	color: #ff4d4f;
	font-weight: 700;
}

.bottom-discount {
	font-size: 24rpx;
	color: #c0c4cc;
}

.bottom-discount-amount {
	font-size: 24rpx;
	color: #20cb6b;
}

.bottom-actions {
	display: flex;
	gap: 16rpx;
	align-items: center;
}

.checkout-btn {
	background: linear-gradient(135deg, #20cb6b, #17b76b);
	color: #fff;
	font-weight: 600;
	width: 280rpx;
	height: 86rpx;
	line-height: 86rpx;
	border-radius: 999rpx;
	border: none;
	font-size: 30rpx;
	margin: 0;
}

.fee-modal-container {
	position: fixed;
	left: 0;
	bottom: 0;
	width: 100%;
	z-index: 1000;
	display: flex;
	align-items: flex-end;
	justify-content: center;
	border-bottom: 1rpx solid #F0F0F0;
}

.fee-modal-mask {
	position: fixed;
	left: 0;
	top: 0;
	width: 100%;
	height: 100%;
	background: rgba(0, 0, 0, 0.35);
}

.fee-modal {
	width: 100%;
	background: #fff;
	border-top-left-radius: 32rpx;
	border-top-right-radius: 32rpx;
	padding: 32rpx 40rpx 60rpx;
	box-sizing: border-box;
	position: relative;
}

.fee-modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	font-size: 32rpx;
	font-weight: 600;
	color: #1a1a1a;
	margin-bottom: 24rpx;
}

.fee-modal-close {
	font-size: 40rpx;
	color: #909399;
	padding: 0 10rpx;
}

.fee-modal-body {
	display: flex;
	flex-direction: column;
	gap: 20rpx;
}

.fee-item {
	display: flex;
	justify-content: space-between;
	padding: 16rpx 0;
	border-bottom: 1px solid #f2f3f5;
}

.fee-item:last-child {
	border-bottom: none;
}

.fee-item-info {
	display: flex;
	flex-direction: column;
	gap: 4rpx;
	color: #303133;
	font-size: 26rpx;
	flex: 1;
	padding-right: 20rpx;
}

.fee-item-name {
	font-weight: 600;
}

.fee-item-spec {
	color: #909399;
	font-size: 24rpx;
}

.fee-item-amount {
	text-align: right;
	font-size: 24rpx;
	color: #606266;
	display: flex;
	flex-direction: column;
	gap: 6rpx;
}

.fee-item-price {
	color: #909399;
}

.fee-item-total {
	font-size: 26rpx;
	color: #303133;
	font-weight: 600;
}


.fee-row {
	display: flex;
	justify-content: space-between;
	font-size: 28rpx;
	color: #303133;
}

.fee-row.total {
	font-size: 34rpx;
	font-weight: 600;
	color: #ff4d4f;
	margin-top: 10rpx;
}

.fee-row .discount {
	color: #ff4d4f;
	font-weight: 600;
}

/* 优惠券相关样式 */
.coupon-section {
	margin-top: 20rpx;
	padding-top: 20rpx;
	border-top: 1rpx solid #f2f3f5;
}

.coupon-row {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 16rpx;
}

.coupon-label {
	font-size: 26rpx;
	color: #303133;
}

.coupon-selector {
	display: flex;
	align-items: center;
	gap: 12rpx;
	padding: 8rpx 16rpx;
	background-color: #f5f7fa;
	border-radius: 8rpx;
}

.coupon-selected {
	font-size: 24rpx;
	color: #20CB6B;
	font-weight: 500;
}

.coupon-placeholder {
	font-size: 24rpx;
	color: #909399;
}

.coupon-change {
	font-size: 22rpx;
	color: #20CB6B;
}

/* 优惠券选择弹窗 */
.coupon-modal-container {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	z-index: 1000;
	display: flex;
	align-items: flex-end;
}

.coupon-modal-mask {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background-color: rgba(0, 0, 0, 0.5);
}

.coupon-modal {
	position: relative;
	width: 100%;
	background-color: #fff;
	border-radius: 32rpx 32rpx 0 0;
	max-height: 80vh;
	display: flex;
	flex-direction: column;
	z-index: 1001;
}

.coupon-modal-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 32rpx;
	border-bottom: 1rpx solid #f2f3f5;
}

.coupon-modal-header text:first-child {
	font-size: 32rpx;
	font-weight: 600;
	color: #303133;
}

.coupon-modal-close {
	font-size: 40rpx;
	color: #909399;
	padding: 0 10rpx;
}

.coupon-modal-body {
	flex: 1;
	overflow-y: auto;
	padding: 32rpx;
}

.coupon-type-section {
	margin-bottom: 40rpx;
}

.coupon-type-title {
	font-size: 28rpx;
	font-weight: 600;
	color: #303133;
	margin-bottom: 20rpx;
	display: block;
}

.coupon-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

.coupon-option {
	padding: 24rpx;
	border: 2rpx solid #e4e7ed;
	border-radius: 16rpx;
	background-color: #fff;
	transition: all 0.3s;
}

.coupon-option.active {
	border-color: #20CB6B;
	background-color: #f0f9f4;
}

.coupon-option.disabled {
	opacity: 0.5;
}

.coupon-option-content {
	display: flex;
	flex-direction: column;
	gap: 8rpx;
}

.coupon-option-name {
	font-size: 28rpx;
	font-weight: 600;
	color: #303133;
}

.coupon-option-value {
	font-size: 32rpx;
	font-weight: 700;
	color: #ff4d4f;
}

.coupon-option-condition {
	font-size: 24rpx;
	color: #909399;
}

.coupon-option-reason {
	font-size: 22rpx;
	color: #f56c6c;
	margin-top: 8rpx;
}
</style>