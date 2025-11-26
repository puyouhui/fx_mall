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
				<view class="tab active">今日添加</view>
				<view class="tab">我常买</view>
				<view class="tab" :class="{ active: isEditing }" @click="toggleEdit">{{ isEditing ? '完成' : '编辑' }}
				</view>
			</view>

			<view class="cart-list" v-if="cartItems.length > 0">
				<view class="cart-item" v-for="item in cartItems" :key="item.id">
					<view class="item-select" @click="toggleSelect(item)">
						<view :class="['select-dot', { active: selectedIds.includes(item.id) }]"></view>
					</view>
					<image :src="item.product_image || defaultImage" class="item-image" mode="aspectFill"></image>
					<view class="item-info">
						<view class="item-title-row">
							<text class="item-name">{{ item.product_name }}</text>
							<text class="item-price">¥{{ getDisplayPrice(item).toFixed(2) }}</text>
						</view>
						<text class="item-spec" v-if="item.spec_name">{{ item.spec_name }}</text>
					<view class="blocked-badge" v-if="isItemBlocked(item)">
						<text class="icon">!</text>
						<text>该商品不参与免配送费</text>
					</view>
						<text class="item-tags">支持采购 · 现货供应</text>
						<view class="item-actions">
							<view class="qty-control">
								<view class="qty-btn" @click="decreaseItem(item)">-</view>
								<text class="qty-value">{{ item.quantity }}</text>
								<view class="qty-btn" @click="increaseItem(item)">+</view>
							</view>
							<view class="item-delete" v-if="isEditing" @click="deleteItem(item)">
								<uni-icons type="trash" size="16" color="#ff4d4f"></uni-icons>
								<text>移除</text>
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
			<view class="assistant-card" v-if="cartItems.length > 0 && deliverySummary && !showFeeDetail">
				<view class="assistant-header">
					<text class="assistant-title">配送费</text>
					<text class="assistant-amount" :class="{ free: actualDeliveryFee === 0 }">
						{{ actualDeliveryFee === 0 ? '免配送费' : '¥' + deliveryFeeText }}
					</text>
				</view>
				<text class="assistant-desc" v-if="deliverySummary.is_free_shipping">
					已满足满 ¥{{ freeShippingThresholdText }} 免配送费
				</text>
				<text class="assistant-desc" v-else-if="selectedDeliveryFeeCoupon">
					已使用免配送费券，当前订单免配送费
				</text>
				<text class="assistant-desc" v-else>
					还差 ¥{{ shortOfAmount }} 可免配送费
				</text>
				<view class="assistant-tips" v-if="deliveryTips.length">
					<view class="assistant-tip" v-for="tip in deliveryTips" :key="tipKey(tip)">
						<text class="tip-name">{{ formatTipName(tip) }}</text>
						<text class="tip-qty" v-if="tip.required_quantity">
							{{ tip.current_quantity || 0 }} / {{ tip.required_quantity }} 件
						</text>
					</view>
				</view>
			</view>

			<view class="bottom-bar" v-if="cartItems.length > 0" @click="openFeeDetail">
				<view class="bottom-left">
					<view class="select-all" @click.stop="selectAll">
						<view class="select-main">
							<view
								:class="['select-dot', { active: selectedIds.length === cartItems.length && cartItems.length > 0 }]">
							</view>
							<text>全选</text>
						</view>
						<text class="selected-count">已选 {{ selectedQuantity }} 件</text>
					</view>
					<view class="bottom-total">
						<text class="bottom-amount">¥{{ finalAmount }}</text>
						<text class="bottom-discount" v-if="actualDeliveryFee > 0">（含配送费 ¥{{ actualDeliveryFee.toFixed(2) }}）</text>
						<text class="bottom-discount" v-else-if="totalDiscount > 0">（已优惠 ¥{{ totalDiscount }}）</text>
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
					<scroll-view scroll-y style="max-height: 400rpx; margin-bottom: 20rpx;" v-if="cartItems.length">
						<view class="fee-item" v-for="item in cartItems" :key="item.id">
							<view class="fee-item-info">
								<text class="fee-item-name">{{ item.product_name }}</text>
								<text class="fee-item-spec" v-if="item.spec_name">{{ item.spec_name }}</text>
							</view>
							<view class="fee-item-amount">
								<text class="fee-item-price">¥{{ getDisplayPrice(item).toFixed(2) }} × {{ item.quantity }}</text>
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
									{{ formatCouponName(selectedAmountCoupon) }} (减¥{{ formatCouponDiscount(selectedAmountCoupon) }})
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
							<view 
								class="coupon-option" 
								:class="{ active: !selectedDeliveryFeeCoupon }"
								@click="selectDeliveryFeeCoupon(null)"
							>
								<text>不使用</text>
							</view>
							<view 
								class="coupon-option" 
								v-for="coupon in availableDeliveryFeeCoupons" 
								:key="coupon.user_coupon_id"
								:class="{ active: selectedDeliveryFeeCoupon?.user_coupon_id === coupon.user_coupon_id }"
								@click="selectDeliveryFeeCoupon(coupon)"
							>
								<text class="coupon-option-name">{{ coupon.name }}</text>
								<text class="coupon-option-desc" v-if="coupon.reason">{{ coupon.reason }}</text>
							</view>
						</view>
					</view>
					<!-- 金额券 -->
					<view class="coupon-type-section" v-if="availableAmountCoupons.length > 0">
						<text class="coupon-type-title">金额券</text>
						<view class="coupon-list">
							<view 
								class="coupon-option" 
								:class="{ active: !selectedAmountCoupon }"
								@click="selectAmountCoupon(null)"
							>
								<text>不使用</text>
							</view>
							<view 
								class="coupon-option" 
								v-for="coupon in availableAmountCoupons" 
								:key="coupon.user_coupon_id"
								:class="{ active: selectedAmountCoupon?.user_coupon_id === coupon.user_coupon_id, disabled: !coupon.is_available }"
								@click="coupon.is_available && selectAmountCoupon(coupon)"
							>
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
				const previousSelection = new Set(this.selectedIds);
				const { items, summary, availableCoupons, bestCombination } = await fetchPurchaseList(this.token);
				this.cartItems = items;
				this.deliverySummary = summary || null;
				this.blockedItemIds = summary?.blocked_item_ids || [];
				if (items.length === 0) {
					this.selectedIds = [];
				} else if (previousSelection.size > 0) {
					const retained = items
						.filter(item => previousSelection.has(item.id))
						.map(item => item.id);
					this.selectedIds = retained.length > 0 ? retained : items.map(item => item.id);
				} else {
					this.selectedIds = items.map(item => item.id);
				}
				this.availableCoupons = availableCoupons || [];
				// 自动应用最佳优惠券组合
				if (bestCombination) {
					this.selectedDeliveryFeeCoupon = bestCombination.delivery_fee_coupon || null;
					this.selectedAmountCoupon = bestCombination.amount_coupon || null;
				} else {
					this.selectedDeliveryFeeCoupon = null;
					this.selectedAmountCoupon = null;
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
		},
		selectAll() {
			if (this.selectedIds.length === this.cartItems.length) {
				this.selectedIds = []
			} else {
				this.selectedIds = this.cartItems.map(item => item.id)
			}
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
		}
	}
};
</script>

<style>
.cart-page {
	height: 100vh;
	display: flex;
	flex-direction: column;
	background: linear-gradient(180deg, #e7fff3 0%, #f8f9fb 40%, #f5f6f8 100%);
}

.cart-body {
	flex: 1;
	display: flex;
	flex-direction: column;
	padding: 24rpx;
	box-sizing: border-box;
	min-height: 0;
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
	flex-direction:row;
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
	gap: 16rpx;
	margin-bottom: 24rpx;
}

.tab {
	padding: 12rpx 32rpx;
	border-radius: 999rpx;
	font-size: 26rpx;
	color: #666;
	background: #fff;
}

.tab.active {
	background: #dff9ef;
	color: #20cb6b;
	font-weight: 600;
}

.cart-list {
	flex: 1;
	overflow-y: auto;
	padding-bottom: 40rpx;
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
	background: #f7f8fa;
	border-radius: 999rpx;
}

.qty-btn {
	width: 56rpx;
	height: 56rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 32rpx;
	color: #333;
}

.qty-value {
	font-size: 28rpx;
	font-weight: 600;
	color: #1a1a1a;
	padding: 0 10rpx;
}

.item-delete {
	display: flex;
	align-items: center;
	gap: 6rpx;
	font-size: 24rpx;
	color: #ff4d4f;
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
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 12rpx 24rpx rgba(24,39,75,0.06);
	display: flex;
	flex-direction: column;
	gap: 12rpx;
}

.assistant-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.assistant-title {
	font-weight: 600;
	font-size: 28rpx;
	color: #303133;
}

.assistant-amount {
	font-size: 36rpx;
	font-weight: 600;
	color: #ff4d4f;
}

.assistant-amount.free {
	color: #20cb6b;
}

.assistant-desc {
	font-size: 26rpx;
	color: #606266;
}

.assistant-tips {
	background: #f5f7fb;
	border-radius: 16rpx;
	padding: 12rpx 16rpx;
	display: flex;
	flex-direction: column;
	gap: 10rpx;
}

.assistant-tip {
	display: flex;
	justify-content: space-between;
	font-size: 24rpx;
	color: #606266;
}

.tip-name {
	font-weight: 500;
}

.tip-qty {
	color: #909399;
}

.bottom-bar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	background: #f7f8fa;
	padding: 24rpx 0;
}

.bottom-left {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 24rpx;
}

.select-all {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	gap: 6rpx;
	font-size: 26rpx;
	color: #606266;
}

.select-main {
	display: flex;
	align-items: center;
	gap: 8rpx;
}

.selected-count {
	font-size: 24rpx;
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

.bottom-actions {
	display: flex;
	gap: 16rpx;
	align-items: center;
}

.checkout-btn {
	background: linear-gradient(135deg, #20cb6b, #17b76b);
	color: #fff;
	font-weight: 600;
	flex: 1;
	height: 86rpx;
	line-height: 86rpx;
	border-radius: 999rpx;
	border: none;
	font-size: 30rpx;
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