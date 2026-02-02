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
			
			// 检查小程序版本更新（仅微信小程序）
			// #ifdef MP-WEIXIN
			this.checkForUpdate()
			// #endif
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
			// 微信确认收货组件回调（appId: wx1183b055aeec94d1）
			if (options && options.referrerInfo && options.referrerInfo.appId === 'wx1183b055aeec94d1') {
				const extra = options.referrerInfo.extraData || {}
				const status = extra.status || ''
				const reqData = extra.req_extradata || {}
				uni.$emit('wechatConfirmReceiveDone', {
					status,
					errormsg: extra.errormsg,
					merchant_trade_no: reqData.merchant_trade_no,
					transaction_id: reqData.transaction_id
				})
			}
		},
		onHide: function() {
			console.log('App Hide')
		},
		methods: {
			// 检查小程序版本更新（仅微信小程序）
			checkForUpdate() {
				// #ifdef MP-WEIXIN
				if (typeof wx !== 'undefined' && wx.getUpdateManager) {
					const updateManager = wx.getUpdateManager()
					
					// 检查是否有新版本
					updateManager.onCheckForUpdate(function (res) {
						// 请求完新版本信息的回调
						console.log('检查更新结果:', res.hasUpdate)
						if (res.hasUpdate) {
							console.log('发现新版本，开始下载...')
						}
					})
					
					// 新版本下载完成
					updateManager.onUpdateReady(function () {
						wx.showModal({
							title: '更新提示',
							content: '新版本已经准备好，是否重启应用？',
							showCancel: true,
							cancelText: '稍后',
							confirmText: '立即重启',
							success(res) {
								if (res.confirm) {
									// 新的版本已经下载好，调用 applyUpdate 应用新版本并重启
									updateManager.applyUpdate()
								}
							}
						})
					})
					
					// 新版本下载失败
					updateManager.onUpdateFailed(function () {
						console.error('新版本下载失败')
						wx.showToast({
							title: '更新失败，请稍后重试',
							icon: 'none',
							duration: 2000
						})
					})
				}
				// #endif
			}
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
