<template>
	<view class="container">
		<!-- 顶部导航栏 -->
		<view class="nav-bar">
			<text class="nav-title">我的采购单</text>
		</view>

		<!-- 采购单商品列表 -->
		<view class="cart-list" v-if="cartItems.length > 0">
			<view class="cart-item" v-for="(item, index) in cartItems" :key="index">
				<image :src="item.image" class="item-image"></image>
				<view class="item-info">
					<text class="item-name">{{ item.name }}</text>
					<text class="item-price" v-if="item.isSpecial">¥{{ item.price }}</text>
					<text class="item-quantity">数量：{{ item.quantity }}</text>
				</view>
				<button class="delete-btn" @click="deleteItem(index)">删除</button>
			</view>
		</view>

		<!-- 空采购单提示 -->
		<view class="empty-cart" v-else>
			<image src="/static/empty-cart.png" class="empty-icon"></image>
			<text class="empty-text">您的采购单还是空的</text>
			<text class="empty-subtext">快去选购商品吧~</text>
		</view>

		<!-- 底部操作栏 -->
		<view class="bottom-bar" v-if="cartItems.length > 0">
			<view class="total-info">
				<text>商品数量：{{ totalQuantity }}</text>
			</view>
			<view class="action-buttons">
				<button class="copy-btn" @click="copyCart">一键复制</button>
				<button class="clear-btn" @click="clearCart">清空采购单</button>
			</view>
		</view>
	</view>
</template>

<script>
	export default {
		data() {
			return {
				cartItems: []
			};
		},
		onShow() {
			// 每次显示页面时从本地存储加载采购单数据
			this.loadCart();
		},
		computed: {
			// 计算采购单商品总数
			totalQuantity() {
				return this.cartItems.reduce((total, item) => total + item.quantity, 0);
			}
		},
		methods: {
			// 从本地存储加载采购单
			loadCart() {
				this.cartItems = uni.getStorageSync('cart') || [];
			},

			// 删除指定商品
			deleteItem(index) {
				uni.showModal({
					title: '确认删除',
					content: '确定要从采购单中删除这个商品吗？',
					success: (res) => {
						if (res.confirm) {
							this.cartItems.splice(index, 1);
							// 更新本地存储
							uni.setStorageSync('cart', this.cartItems);
							// 提示删除成功
							uni.showToast({
								title: '删除成功',
								icon: 'success',
								duration: 2000
							});
						}
					}
				});
			},

			// 一键复制采购单内容
			copyCart() {
				let copyText = '我的采购单：\n';
				
				this.cartItems.forEach((item, index) => {
					copyText += `${index + 1}. ${item.name}`;
					if (item.isSpecial) {
						copyText += ` - ¥${item.price}`;
					}
					copyText += ` (数量：${item.quantity})\n`;
				});
				
				// 调用小程序API复制文本
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

			// 清空采购单
			clearCart() {
				uni.showModal({
					title: '确认清空',
					content: '确定要清空整个采购单吗？',
					success: (res) => {
						if (res.confirm) {
							this.cartItems = [];
							// 清空本地存储
							uni.removeStorageSync('cart');
							// 提示清空成功
							uni.showToast({
								title: '采购单已清空',
								icon: 'success',
								duration: 2000
							});
						}
					}
				});
			}
		}
	};
</script>

<style>
	.container {
		height: 100vh;
		display: flex;
		flex-direction: column;
	}

	/* 顶部导航栏样式 */
	.nav-bar {
		height: 88rpx;
		background-color: #007AFF;
		display: flex;
		align-items: center;
		justify-content: center;
		position: relative;
	}

	.nav-title {
		color: #fff;
		font-size: 36rpx;
		font-weight: bold;
	}

	/* 采购单列表样式 */
	.cart-list {
		flex: 1;
		padding: 20rpx;
		overflow-y: auto;
		background-color: #f5f5f5;
	}

	.cart-item {
		display: flex;
		align-items: center;
		background-color: #fff;
		border-radius: 10rpx;
		padding: 20rpx;
		margin-bottom: 20rpx;
		box-shadow: 0 2rpx 10rpx rgba(0, 0, 0, 0.05);
	}

	.item-image {
		width: 140rpx;
		height: 140rpx;
		border-radius: 10rpx;
		margin-right: 20rpx;
	}

	.item-info {
		flex: 1;
	}

	.item-name {
		font-size: 30rpx;
		color: #333;
		line-height: 40rpx;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
		margin-bottom: 10rpx;
	}

	.item-price {
		font-size: 32rpx;
		color: #f00;
		font-weight: bold;
		display: block;
		margin-bottom: 10rpx;
	}

	.item-quantity {
		font-size: 28rpx;
		color: #999;
	}

	.delete-btn {
		background-color: #f00;
		color: #fff;
		font-size: 28rpx;
		padding: 10rpx 30rpx;
		border-radius: 30rpx;
		border: none;
	}

	/* 空采购单样式 */
	.empty-cart {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		background-color: #f5f5f5;
	}

	.empty-icon {
		width: 200rpx;
		height: 200rpx;
		margin-bottom: 30rpx;
		opacity: 0.5;
	}

	.empty-text {
		font-size: 32rpx;
		color: #999;
		margin-bottom: 10rpx;
	}

	.empty-subtext {
		font-size: 28rpx;
		color: #ccc;
	}

	/* 底部操作栏样式 */
	.bottom-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 20rpx;
		background-color: #fff;
		border-top: 1rpx solid #eee;
	}

	.total-info {
		font-size: 30rpx;
		color: #333;
	}

	.action-buttons {
		display: flex;
		gap: 20rpx;
	}

	.copy-btn {
		background-color: #007AFF;
		color: #fff;
		font-size: 28rpx;
		padding: 10rpx 30rpx;
		border-radius: 30rpx;
		border: none;
	}

	.clear-btn {
		background-color: #f00;
		color: #fff;
		font-size: 28rpx;
		padding: 10rpx 30rpx;
		border-radius: 30rpx;
		border: none;
	}
</style>