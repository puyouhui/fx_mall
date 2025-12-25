<template>
  <view class="profile-page">
    <!-- 自定义导航栏 -->
    <view class="custom-header">
      <view class="navbar-fixed" style="background-color: #E8F8F0;">
        <!-- 状态栏撑起高度 -->
        <view :style="{ height: statusBarHeight + 'px' }"></view>
        <!-- 导航栏内容区域 -->
        <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
          <view class="navbar-left" @click="goBack">
            <uni-icons type="left" size="20" color="#333"></uni-icons>
          </view>
          <view class="navbar-title">
            <text class="navbar-title-text">个人资料</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 个人资料内容 -->
    <view class="profile-content">
      <!-- 用户信息卡片 -->
      <view class="info-card">
        <view class="card-header">
          <text class="card-title">基本信息</text>
        </view>

        <!-- 头像 -->
        <view class="info-item avatar-item" @click="editAvatar">
          <text class="info-label">头像</text>
          <view class="avatar-wrapper">
            <image v-if="userInfo.avatar" :src="userInfo.avatar" class="avatar-image" mode="aspectFill"></image>
            <view v-else class="avatar-placeholder">
              <uni-icons type="person-filled" size="40" color="#ccc"></uni-icons>
            </view>
            <uni-icons type="right" size="16" color="#ddd" style="margin-left: 20rpx;"></uni-icons>
          </view>
        </view>

        <!-- 用户姓名 -->
        <view class="info-item editable" @click="editName">
          <text class="info-label">姓名</text>
          <view class="info-value-wrapper">
            <text class="info-value">{{ displayName }}</text>
            <uni-icons type="right" size="16" color="#ddd"></uni-icons>
          </view>
        </view>

        <!-- 电话 -->
        <view class="info-item editable" @click="editPhone">
          <text class="info-label">电话</text>
          <view class="info-value-wrapper">
            <text class="info-value">{{ userInfo.phone || '未设置' }}</text>
            <uni-icons type="right" size="16" color="#ddd"></uni-icons>
          </view>
        </view>


      </view>
      <!-- 用户编号（底部，灰色） -->
      <view class="info-item user-code-item">
        <text class="info-label user-code-label">用户编号</text>
        <text class="info-value user-code-value">{{ userInfo.user_code ? `${userInfo.user_code}` : '未设置' }}</text>
      </view>
    </view>

    <!-- 编辑姓名弹窗 -->
    <view class="modal-overlay" v-if="showNameModal" @click="closeNameModal">
      <view class="modal-content" @click.stop>
        <view class="modal-header">
          <text class="modal-title">编辑姓名</text>
          <view class="modal-close" @click="closeNameModal">
            <uni-icons type="close" size="20" color="#999"></uni-icons>
          </view>
        </view>
        <view class="modal-body">
          <input class="name-input" v-model="editNameValue" placeholder="请输入姓名" maxlength="50"
            :focus="nameInputFocus" />
          <text class="input-tip">最多50个字符</text>
        </view>
        <view class="modal-footer">
          <view class="modal-btn cancel-btn" @click="closeNameModal">
            <text>取消</text>
          </view>
          <view class="modal-btn confirm-btn" @click="saveName">
            <text>保存</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 编辑电话弹窗 -->
    <view class="modal-overlay" v-if="showPhoneModal" @click="closePhoneModal">
      <view class="modal-content" @click.stop>
        <view class="modal-header">
          <text class="modal-title">编辑电话</text>
          <view class="modal-close" @click="closePhoneModal">
            <uni-icons type="close" size="20" color="#999"></uni-icons>
          </view>
        </view>
        <view class="modal-body">
          <input class="name-input" v-model="editPhoneValue" placeholder="请输入电话" type="number" maxlength="20"
            :focus="phoneInputFocus" />
          <text class="input-tip">最多20个字符</text>
        </view>
        <view class="modal-footer">
          <view class="modal-btn cancel-btn" @click="closePhoneModal">
            <text>取消</text>
          </view>
          <view class="modal-btn confirm-btn" @click="savePhone">
            <text>保存</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { getMiniUserInfo, updateMiniUserName, updateMiniUserPhone, uploadMiniUserAvatar } from '../../api/index.js';

export default {
  data() {
    return {
      statusBarHeight: 20,
      navBarHeight: 45,
      userInfo: {},
      showNameModal: false,
      editNameValue: '',
      nameInputFocus: false,
      showPhoneModal: false,
      editPhoneValue: '',
      phoneInputFocus: false
    };
  },
  computed: {
    displayName() {
      if (this.userInfo.name && this.userInfo.name.trim()) {
        return this.userInfo.name;
      }
      return this.userInfo.user_code ? `${this.userInfo.user_code}` : '未设置';
    },
    userTypeText() {
      if (!this.userInfo.user_type) {
        return '未设置';
      }
      const typeMap = {
        'retail': '零售用户',
        'wholesale': '批发用户'
      };
      return typeMap[this.userInfo.user_type] || '未知';
    },
    userTypeClass() {
      return {
        'type-retail': this.userInfo.user_type === 'retail',
        'type-wholesale': this.userInfo.user_type === 'wholesale'
      };
    }
  },
  onLoad() {
    // 获取设备信息
    const info = uni.getSystemInfoSync();
    this.statusBarHeight = info.statusBarHeight;
    this.getMenuButtonInfo();

    // 加载用户信息
    this.loadUserInfo();
  },
  onShow() {
    // 页面显示时重新加载用户信息
    this.loadUserInfo();
  },
  methods: {
    getMenuButtonInfo() {
      try {
        // #ifndef H5 || APP-PLUS || MP-ALIPAY
        const info = uni.getSystemInfoSync();
        const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
        this.navBarHeight = (menuButtonInfo.bottom - info.statusBarHeight) + (menuButtonInfo.top - info.statusBarHeight);
        // #endif
      } catch (e) {
        console.log('获取胶囊按钮信息失败', e);
      }
    },
    async loadUserInfo() {
      const token = uni.getStorageSync('miniUserToken');
      if (!token) {
        uni.showToast({
          title: '请先登录',
          icon: 'none'
        });
        setTimeout(() => {
          uni.navigateBack();
        }, 1500);
        return;
      }

      try {
        const res = await getMiniUserInfo(token);
        if (res && res.code === 200 && res.data) {
          this.userInfo = res.data;
        } else {
          uni.showToast({
            title: '获取用户信息失败',
            icon: 'none'
          });
        }
      } catch (error) {
        console.error('加载用户信息失败:', error);
        uni.showToast({
          title: '加载失败，请重试',
          icon: 'none'
        });
      }
    },
    editName() {
      this.editNameValue = this.userInfo.name || '';
      this.showNameModal = true;
      this.$nextTick(() => {
        this.nameInputFocus = true;
      });
    },
    closeNameModal() {
      this.showNameModal = false;
      this.nameInputFocus = false;
      this.editNameValue = '';
    },
    async saveName() {
      const name = this.editNameValue.trim();

      if (name.length > 50) {
        uni.showToast({
          title: '姓名长度不能超过50个字符',
          icon: 'none'
        });
        return;
      }

      const token = uni.getStorageSync('miniUserToken');
      if (!token) {
        uni.showToast({
          title: '请先登录',
          icon: 'none'
        });
        return;
      }

      try {
        uni.showLoading({
          title: '保存中...'
        });

        const res = await updateMiniUserName(name, token);
        uni.hideLoading();

        if (res && res.code === 200) {
          uni.showToast({
            title: '保存成功',
            icon: 'success'
          });
          this.closeNameModal();
          // 重新加载用户信息
          await this.loadUserInfo();
        } else {
          uni.showToast({
            title: res.message || '保存失败',
            icon: 'none'
          });
        }
      } catch (error) {
        uni.hideLoading();
        console.error('保存姓名失败:', error);
        uni.showToast({
          title: '保存失败，请重试',
          icon: 'none'
        });
      }
    },
    editAvatar() {
      uni.showActionSheet({
        itemList: ['拍照', '从相册选择'],
        success: (res) => {
          if (res.tapIndex === 0) {
            // 拍照
            this.chooseImage('camera');
          } else if (res.tapIndex === 1) {
            // 从相册选择
            this.chooseImage('album');
          }
        }
      });
    },
    chooseImage(sourceType) {
      uni.chooseImage({
        count: 1,
        sizeType: ['compressed'], // 使用压缩模式
        sourceType: sourceType === 'camera' ? ['camera'] : ['album'],
        success: async (res) => {
          const tempFilePath = res.tempFilePaths[0];
          const token = uni.getStorageSync('miniUserToken');
          if (!token) {
            uni.showToast({
              title: '请先登录',
              icon: 'none'
            });
            return;
          }

          try {
            uni.showLoading({
              title: '压缩中...'
            });

            // 压缩图片
            const compressedPath = await this.compressImage(tempFilePath);
            
            uni.showLoading({
              title: '上传中...'
            });

            const uploadRes = await uploadMiniUserAvatar(compressedPath, token);
            uni.hideLoading();

            if (uploadRes && uploadRes.code === 200) {
              uni.showToast({
                title: '上传成功',
                icon: 'success'
              });
              // 重新加载用户信息
              await this.loadUserInfo();
            } else {
              uni.showToast({
                title: uploadRes.message || '上传失败',
                icon: 'none'
              });
            }
          } catch (error) {
            uni.hideLoading();
            console.error('上传头像失败:', error);
            uni.showToast({
              title: '上传失败，请重试',
              icon: 'none'
            });
          }
        },
        fail: (err) => {
          console.error('选择图片失败:', err);
        }
      });
    },
    editPhone() {
      this.editPhoneValue = this.userInfo.phone || '';
      this.showPhoneModal = true;
      this.$nextTick(() => {
        this.phoneInputFocus = true;
      });
    },
    closePhoneModal() {
      this.showPhoneModal = false;
      this.phoneInputFocus = false;
      this.editPhoneValue = '';
    },
    async savePhone() {
      const phone = this.editPhoneValue.trim();

      if (phone.length > 20) {
        uni.showToast({
          title: '电话长度不能超过20个字符',
          icon: 'none'
        });
        return;
      }

      const token = uni.getStorageSync('miniUserToken');
      if (!token) {
        uni.showToast({
          title: '请先登录',
          icon: 'none'
        });
        return;
      }

      try {
        uni.showLoading({
          title: '保存中...'
        });

        const res = await updateMiniUserPhone(phone, token);
        uni.hideLoading();

        if (res && res.code === 200) {
          uni.showToast({
            title: '保存成功',
            icon: 'success'
          });
          this.closePhoneModal();
          // 重新加载用户信息
          await this.loadUserInfo();
        } else {
          uni.showToast({
            title: res.message || '保存失败',
            icon: 'none'
          });
        }
      } catch (error) {
        uni.hideLoading();
        console.error('保存电话失败:', error);
        uni.showToast({
          title: '保存失败，请重试',
          icon: 'none'
        });
      }
    },
    goBack() {
      uni.navigateBack();
    },
    // 压缩图片
    compressImage(filePath) {
      return new Promise((resolve, reject) => {
        uni.compressImage({
          src: filePath,
          quality: 60, // 压缩质量，值越小压缩越多
          success: (res) => {
            resolve(res.tempFilePath);
          },
          fail: (err) => {
            console.error('压缩图片失败:', err);
            // 压缩失败时使用原图
            resolve(filePath);
          }
        });
      });
    }
  }
};
</script>

<style scoped>
.profile-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #E8F8F0 0%, #E8F8F0 20%, #f5f5f5 40%, #f5f5f5 100%);
}

/* 自定义导航栏 */
.custom-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
}

.navbar-fixed {
  position: relative;
}

.navbar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 30rpx;
}

.navbar-left,
.navbar-right {
  width: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.navbar-title {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.navbar-title-text {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}

/* 个人资料内容 */
.profile-content {
  padding: 30rpx 24rpx;
  position: relative;
  z-index: 1;
}

.info-card {
  background: #fff;
  border-radius: 24rpx;
  overflow: hidden;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.04);
}

.card-header {
  padding: 36rpx 30rpx 28rpx;
  border-bottom: 1rpx solid #f0f0f0;
  background: linear-gradient(135deg, #fafafa 0%, #ffffff 100%);
}

.card-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #333;
  letter-spacing: 0.5rpx;
}

.info-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 32rpx 30rpx;
  border-bottom: 1rpx solid #f0f0f0;
  transition: background-color 0.2s;
}

.info-item:last-child {
  border-bottom: none;
}

.info-item.editable:active,
.info-item.avatar-item:active {
  background-color: #f8f8f8;
}

.info-item.editable,
.info-item.avatar-item {
  cursor: pointer;
}

.info-item.user-code-item {
  margin-top: 20rpx;
  border-bottom: none;
  padding: 24rpx 30rpx;
  justify-content: center;
}

.avatar-wrapper {
  display: flex;
  align-items: center;
}

.avatar-image {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  border: 3rpx solid #f0f0f0;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.08);
}

.avatar-placeholder {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #f5f5f5 0%, #e8e8e8 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 3rpx solid #f0f0f0;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.08);
}

.info-label {
  font-size: 30rpx;
  color: #333;
  font-weight: 500;
  min-width: 120rpx;
}

.info-value-wrapper {
  display: flex;
  align-items: center;
  gap: 20rpx;
  justify-content: flex-end;
}

.info-value {
  font-size: 30rpx;
  color: #333;
}

.user-code-label {
  font-size: 26rpx;
  color: #999;
  font-weight: 400;
}

.user-code-value {
  font-size: 26rpx;
  color: #999;
}

/* 编辑姓名弹窗 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.modal-content {
  width: 640rpx;
  background: #fff;
  border-radius: 24rpx;
  overflow: hidden;
  box-shadow: 0 8rpx 40rpx rgba(0, 0, 0, 0.12);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 30rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.modal-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}

.modal-close {
  width: 40rpx;
  height: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-body {
  padding: 30rpx;
}

.name-input {
  width: 100%;
  height: 88rpx;
  padding: 0 24rpx;
  background: #f8f8f8;
  border-radius: 12rpx;
  font-size: 30rpx;
  color: #333;
  border: 2rpx solid #f0f0f0;
  box-sizing: border-box;
  transition: border-color 0.2s;
}

.name-input:focus {
  border-color: #20CB6B;
  background: #fff;
}

.input-tip {
  display: block;
  margin-top: 20rpx;
  font-size: 24rpx;
  color: #999;
}

.modal-footer {
  display: flex;
  border-top: 1rpx solid #f0f0f0;
}

.modal-btn {
  flex: 1;
  height: 108rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  transition: background-color 0.2s;
}

.cancel-btn {
  color: #666;
  border-right: 1rpx solid #f0f0f0;
}

.cancel-btn:active {
  background-color: #f5f5f5;
}

.confirm-btn {
  color: #20CB6B;
  font-weight: 600;
}

.confirm-btn:active {
  background-color: #f0fdf4;
}
</style>
