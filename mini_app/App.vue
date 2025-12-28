<script>
	export default {
		globalData: {
			targetCategoryId: null
		},
		onLaunch: function(options) {
			console.log('App Launch', options)
			// 检查分享参数
			if (options && options.query) {
				const referrerId = options.query.referrer_id
				if (referrerId) {
					// 保存分享者ID到本地存储
					uni.setStorageSync('shareReferrerId', referrerId)
					console.log('保存分享者ID:', referrerId)
				}
			}
		},
		onShow: function(options) {
			console.log('App Show', options)
			// 检查分享参数（从其他小程序或分享卡片打开）
			if (options && options.query) {
				const referrerId = options.query.referrer_id
				if (referrerId) {
					// 保存分享者ID到本地存储
					uni.setStorageSync('shareReferrerId', referrerId)
					console.log('保存分享者ID:', referrerId)
				}
			}
		},
		onHide: function() {
			console.log('App Hide')
		}
	}
</script>

<style>
	/* 全局样式重置 */
/* 	* {
		margin: 0;
		padding: 0;
		box-sizing: border-box;
	} */

	/* 全局字体设置 */
	page {
		font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', sans-serif;
		font-size: 28rpx;
		color: #333;
		background-color: #f8f8f8;
	}

	/* 滚动条样式 */
	::-webkit-scrollbar {
		display: none;
	}

	/* 按钮基础样式 */
	button {
		font-size: 28rpx;
	}

	/* 链接样式 */
	a {
		color: #007AFF;
	}

	/* 表单元素样式 */
	input, textarea {
		font-size: 28rpx;
	}

	/* 防止页面被拖拽 */
	body {
		user-select: none;
	}
</style>
