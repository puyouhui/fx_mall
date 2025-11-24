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

			<!-- 底部操作栏 -->
			<view class="assistant-card" v-if="cartItems.length > 0">
				<text class="assistant-title">凑单助手</text>
				<text class="assistant-desc">再加 ¥20.00 可享包邮</text>
			</view>

			<view class="bottom-bar" v-if="cartItems.length > 0">
				<view class="bottom-left">
					<view class="select-all" @click="selectAll">
						<view
							:class="['select-dot', { active: selectedIds.length === cartItems.length && cartItems.length > 0 }]">
						</view>
						<text>全选</text>
					</view>
					<view class="bottom-total">
						<text class="bottom-label">已选 {{ selectedQuantity }} 件，应付：</text>
						<text class="bottom-amount">¥{{ selectedAmount }}</text>
						<text class="bottom-discount">（仅供参考）</text>
					</view>
				</view>
				<button class="checkout-btn">去下单</button>
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
			userType: 'unknown',
			loading: false,
			token: '',
			defaultImage: '/static/empty-cart.png',
			statusBarHeight: 0,
			navBarHeight: 44,
			menuButtonRect: null,
			selectedIds: [],
			isEditing: false
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
					return;
				}
				const list = await fetchPurchaseList(this.token);
				this.cartItems = list;
				this.selectedIds = list.map(item => item.id)
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
	padding: 20rpx 24rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 12rpx 24rpx rgba(24,39,75,0.06);
	display: flex;
	justify-content: space-between;
	align-items: center;
	color: #20cb6b;
	font-size: 26rpx;
}

.assistant-title {
	font-weight: 600;
}

.assistant-desc {
	color: #ff893a;
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
	align-items: center;
	gap: 12rpx;
	font-size: 26rpx;
	color: #606266;
}

.bottom-total {
	display: flex;
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
</style>