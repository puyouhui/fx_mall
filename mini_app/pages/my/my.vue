<template>
	<view>
		11231231231221123
	</view>
</template>

<script>
	import { getMiniUserInfo } from '../../api/index.js';
	
	export default {
		data() {
			return {
				
			}
		},
		// 页面显示时更新用户信息
		onShow() {
			this.updateUserInfo();
		},
		methods: {
			// 更新用户信息
			async updateUserInfo() {
				try {
					const token = uni.getStorageSync('miniUserToken');
					if (!token) {
						// 未登录，不获取用户信息
						return;
					}

					const res = await getMiniUserInfo(token);
					if (res && res.code === 200 && res.data) {
						// 更新本地存储的用户信息
						uni.setStorageSync('miniUserInfo', res.data);
						if (res.data.unique_id) {
							uni.setStorageSync('miniUserUniqueId', res.data.unique_id);
						}
					}
				} catch (error) {
					console.error('获取用户信息失败:', error);
					// 静默失败，不显示错误提示
				}
			}
		}
	}
</script>

<style>

</style>
