<template>
  <view class="profile-form-page">
    <!-- 地图背景 -->
    <map class="map-background" :latitude="mapLocation.latitude" :longitude="mapLocation.longitude" :scale="15"
      :show-location="true" :enable-zoom="false" :enable-scroll="false" :markers="markers"></map>

    <!-- 表单卡片 -->
    <view class="form-card">
      <!-- 选择店铺定位按钮 -->
      <view class="location-selector" @click="selectLocation">
        <uni-icons type="location" size="20" color="#20CB6B" class="location-icon"></uni-icons>
        <view class="location-content">
          <text class="location-text">{{ formData.address || '点击选择店铺定位' }}</text>
          <uni-icons type="right" size="18" color="#20CB6B" style="margin-top: 6rpx;"></uni-icons>
        </view>
      </view>

      <!-- 表单内容 -->
      <view class="form-content">
        <!-- 店铺名称 -->
        <view class="form-item">
          <view class="form-item-header">
            <text class="form-label">店铺名称 <text class="required">(必填)</text></text>
            <view class="photo-wrapper" @click="uploadStorePhoto">
              <text v-if="!formData.avatar" class="photo-label">门头照片</text>
              <image v-else :src="formData.avatar" class="photo-thumbnail" mode="aspectFill"></image>
            </view>
          </view>
          <input v-model="formData.name" class="form-input" placeholder="例如:王记烤鸭店" maxlength="100" />
        </view>

        <!-- 联系人 -->
        <view class="form-item">
          <text class="form-label">联系人 <text class="required">(必填)</text></text>
          <input v-model="formData.contact" class="form-input" placeholder="例如:孙先生" maxlength="50" />
        </view>

        <!-- 手机号 -->
        <view class="form-item">
          <text class="form-label">手机号 <text class="required">(必填)</text></text>
          <input v-model="formData.phone" class="form-input" type="number" placeholder="请输入收货人电话" maxlength="20" />
        </view>

        <!-- 店铺地址 - 选择定位后显示 -->
        <view class="form-item" v-if="showAddressFields">
          <text class="form-label">店铺地址 <text class="required">(必填)</text></text>
          <textarea v-model="formData.address" class="form-textarea" placeholder="请完善详细地址信息（如：门牌号、楼层等）" maxlength="255"
            :auto-height="true" :show-confirm-bar="false" @input="onAddressInput" />
          <view class="address-tip" v-if="showAddressFields && formData.address">
            <uni-icons type="info" size="14" color="#909399"></uni-icons>
            <text class="tip-text">您可以继续完善地址信息，如门牌号、楼层，以便准确送达。</text>
          </view>
        </view>

        <!-- 店铺类型 - 选择定位后显示 -->
        <view class="form-item" v-if="showAddressFields">
          <text class="form-label">店铺类型 <text class="required">(非必填)</text></text>
          <view class="picker-wrapper" @click="showStoreTypePicker = true">
            <input v-model="formData.storeType" class="form-input picker-input" placeholder="请输入店铺类型" disabled />
            <uni-icons type="bottom" size="16" color="#999" class="picker-icon"></uni-icons>
          </view>
        </view>

        <!-- 销售员代码 - 只在用户未绑定销售员时显示 -->
        <view class="form-item sales-code-item" v-if="showAddressFields && !userHasSalesCode">
          <view v-if="!showSalesCodeInput" class="sales-code-link" @click="showSalesCodeInput = true">
            <text class="link-text">绑定业务员</text>
            <uni-icons type="right" size="14" color="#20CB6B" style="margin-top: 4rpx;"></uni-icons>
          </view>
          <view v-else class="sales-code-wrapper">
            <view class="sales-code-header">
              <text class="form-label">业务员代码</text>
              <view class="sales-code-close" @click="closeSalesCodeInput">
                <uni-icons type="close" size="18" color="#999"></uni-icons>
              </view>
            </view>
            <view class="sales-code-inputs">
              <input v-for="(code, index) in salesCodeArray" :key="index" :id="`sales-code-${index}`"
                v-model="salesCodeArray[index]" class="sales-code-box" type="number" maxlength="1"
                :focus="salesCodeFocusIndex === index" @input="onSalesCodeInput(index, $event)"
                @focus="onSalesCodeFocus(index)" @blur="onSalesCodeBlur(index)" />
            </view>
          </view>
        </view>

        <!-- 设置为默认地址 -->
        <view class="form-item default-address-item" v-if="showAddressFields">
          <view class="default-address-switch">
            <text class="form-label">设置为默认地址</text>
            <switch :checked="formData.isDefault" @change="onDefaultAddressChange" color="#20CB6B" />
          </view>
        </view>
      </view>

      <!-- 底部按钮 -->
      <view class="form-footer">
        <!-- 编辑状态下显示删除文字 -->
        <view class="delete-text-wrapper" v-if="formData.addressId">
          <text class="delete-text" @click="handleDeleteAddress" v-if="!deleting">删除地址</text>
          <text class="delete-text deleting" v-else>删除中...</text>
        </view>
        <button class="form-btn submit-btn" :class="{ disabled: !canSubmit || submitting }" @click="handleSubmit">
          {{ submitting ? '提交中...' : '确认信息' }}
        </button>
        <button class="form-btn import-btn" @click="importWeChatAddress" v-if="!formData.addressId">
          导入微信收货地址
        </button>
      </view>
    </view>

    <!-- 店铺类型选择器弹窗 -->
    <view class="popup-overlay" v-if="showStoreTypePicker" @click="showStoreTypePicker = false">
      <view class="popup-content" @click.stop>
        <view class="popup-header">
          <text class="popup-title">选择店铺类型</text>
          <view class="popup-close" @click="showStoreTypePicker = false">
            <uni-icons type="close" size="20" color="#666"></uni-icons>
          </view>
        </view>
        <scroll-view scroll-y class="popup-list">
          <view v-for="(type, index) in storeTypeOptions" :key="index" class="popup-item"
            :class="{ active: formData.storeType === type }" @click="selectStoreType(type)">
            <text>{{ type }}</text>
            <uni-icons v-if="formData.storeType === type" type="checkmarkempty" size="18" color="#20CB6B"></uni-icons>
          </view>
        </scroll-view>
      </view>
    </view>
  </view>
</template>

<script>
import { updateMiniUserProfile, uploadAddressAvatar, getMiniUserAddresses, getMiniUserDefaultAddress, getMiniUserInfo, geocodeAddress, deleteMiniUserAddress } from '../../api/index';

export default {
  data() {
    return {
      formData: {
        addressId: null, // 地址ID，用于编辑
        name: '',
        contact: '',
        phone: '',
        address: '',
        storeType: '',
        salesCode: '',
        latitude: null,
        longitude: null,
        avatar: '',
        isDefault: false // 是否设置为默认地址
      },
      salesCodeArray: ['', '', '', '', ''],
      submitting: false,
      deleting: false, // 删除中状态
      userToken: '',
      userInfo: null,
      userHasSalesCode: false, // 用户是否已绑定销售员
      showStoreTypePicker: false,
      showSalesCodeInput: false,
      salesCodeFocusIndex: -1,
      showAddressFields: false,
      mapLocation: {
        latitude: 39.908823,
        longitude: 116.397470
      },
      markers: [], // 地图标记点
      storeTypeOptions: [

        '餐饮店',
        '酒店',
        '民宿',
        '便利店',
        '超市',
        '酒吧/KTV',
        '其他'
      ]
    };
  },
  computed: {
    canSubmit() {
      return (
        this.formData.name.trim() &&
        this.formData.contact.trim() &&
        this.formData.phone.trim() &&
        this.formData.address.trim()
      );
    }
  },
  async onLoad(options) {
    const token = uni.getStorageSync('miniUserToken') || '';
    if (!token) {
      uni.showToast({
        title: '请先完成登录',
        icon: 'none'
      });
      setTimeout(() => {
        uni.navigateBack({ delta: 1 });
      }, 800);
      return;
    }

    this.userToken = token;

    // 先获取用户信息，检查是否已绑定销售员
    this.loadUserInfo();

    // 判断是新增还是编辑（通过URL参数address_id）
    if (options.address_id) {
      // 编辑模式：加载地址数据（会设置经纬度）
      this.formData.addressId = parseInt(options.address_id);
      // 先加载地址数据，确保经纬度正确
      await this.loadAddressData(parseInt(options.address_id));
    } else {
      // 新增模式：不显示地址字段，需要先选择定位
      this.showAddressFields = false;
      // 新增模式才获取当前位置作为默认值
      this.getCurrentLocation();
    }
  },
  methods: {
    // 加载用户信息，检查是否已绑定销售员
    async loadUserInfo() {
      try {
        const res = await getMiniUserInfo(this.userToken);
        if (res && res.code === 200 && res.data) {
          this.userInfo = res.data;
          // 检查用户是否已绑定销售员
          this.userHasSalesCode = !!(res.data.sales_code && res.data.sales_code.trim());
        }
      } catch (error) {
        console.error('获取用户信息失败:', error);
        // 静默失败，不影响其他功能
      }
    },
    // 获取当前位置
    getCurrentLocation() {
      uni.getLocation({
        type: 'gcj02',
        success: (res) => {
          this.mapLocation = {
            latitude: res.latitude,
            longitude: res.longitude
          };
          this.formData.latitude = res.latitude;
          this.formData.longitude = res.longitude;
          // 更新地图标记
          this.updateMarkers();
        },
        fail: () => {
          console.log('获取位置失败');
        }
      });
    },
    // 更新地图标记点
    updateMarkers() {
      if (this.formData.latitude && this.formData.longitude) {
        this.markers = [{
          id: 1,
          latitude: this.formData.latitude,
          longitude: this.formData.longitude,
          // 不设置iconPath，使用默认的红色标记点
          width: 30,
          height: 45,
          // 不设置width和height，使用系统默认大小，避免变形
          anchor: {
            x: 0.5,
            y: 1
          },
          callout: {
            content: this.formData.name || '店铺位置',
            color: '#333',
            fontSize: 14,
            borderRadius: 4,
            bgColor: '#fff',
            padding: 8,
            display: 'BYCLICK' // 点击时显示
          }
        }];
      } else {
        this.markers = [];
      }
    },
    // 选择店铺定位
    selectLocation() {
      uni.chooseLocation({
        success: (res) => {
          console.log(res);
          this.mapLocation = {
            latitude: res.latitude,
            longitude: res.longitude
          };
          this.formData.latitude = res.latitude;
          this.formData.longitude = res.longitude;
          // 更新地图标记
          this.updateMarkers();
          // 选择定位后显示地址和店铺类型字段
          this.showAddressFields = true;
          // 延迟设置地址，确保textarea已经渲染
          this.$nextTick(() => {
            this.formData.address = res.address || res.name;
            // 提示用户完善详细地址
            uni.showToast({
              title: '您可以继续完善详细地址信息',
              icon: 'none',
              duration: 2000
            });
            // 检查用户是否已有默认地址，如果没有则默认选中"设置为默认地址"
            this.checkDefaultAddress();
          });
        },
        fail: (err) => {
          console.log('选择位置失败:', err);
          uni.showToast({
            title: '选择位置失败',
            icon: 'none'
          });
        }
      });
    },
    // 地址输入事件
    onAddressInput(event) {
      // 触发高度重新计算
      this.$nextTick(() => {
        // 确保高度正确计算
      });
    },
    // 上传门头照片
    uploadStorePhoto() {
      uni.showActionSheet({
        itemList: ['拍摄', '从相册选择'],
        success: (res) => {
          const sourceType = res.tapIndex === 0 ? ['camera'] : ['album'];
          uni.chooseImage({
            count: 1,
            sizeType: ['compressed'],
            sourceType: sourceType,
            success: async (chooseRes) => {
              const tempFilePath = chooseRes.tempFilePaths[0];
              // 上传图片
              await this.uploadAvatar(tempFilePath);
            },
            fail: (err) => {
              console.log('选择图片失败:', err);
            }
          });
        }
      });
    },
    // 上传地址照片到服务器（门头照片）
    async uploadAvatar(filePath) {
      if (!this.userToken) {
        uni.showToast({
          title: '请先登录',
          icon: 'none'
        });
        return;
      }

      uni.showLoading({
        title: '上传中...',
        mask: true
      });

      try {
        // 调用地址头像上传接口（只上传图片，不更新用户表）
        const res = await uploadAddressAvatar(filePath, this.userToken);
        if (res && res.code === 200 && res.data) {
          // 更新本地头像（这是地址的照片，只保存在formData中，提交时保存到地址表）
          this.formData.avatar = res.data.avatar || res.data.imageUrl || '';
          uni.showToast({
            title: '上传成功',
            icon: 'success'
          });
        } else {
          throw new Error(res?.message || '上传失败');
        }
      } catch (error) {
        console.error('上传地址照片失败:', error);
        uni.showToast({
          title: error?.message || '上传失败，请稍后重试',
          icon: 'none'
        });
      } finally {
        uni.hideLoading();
      }
    },
    // 选择店铺类型
    selectStoreType(type) {
      this.formData.storeType = type;
      this.showStoreTypePicker = false;
    },
    // 销售员代码输入
    onSalesCodeInput(index, event) {
      let value = event.detail.value;
      // 只允许输入数字，如果输入了多个字符，只取最后一个
      if (value) {
        const numbers = value.replace(/\D/g, '');
        if (numbers.length > 0) {
          value = numbers.slice(-1); // 只取最后一个数字
        } else {
          value = '';
        }
        this.salesCodeArray[index] = value;
      }

      // 更新销售员代码
      this.formData.salesCode = this.salesCodeArray.join('');

      // 如果输入了数字且不是最后一个输入框，自动跳转到下一个
      if (value && index < 4) {
        this.salesCodeFocusIndex = index + 1;
      }
    },
    // 销售员代码输入框聚焦
    onSalesCodeFocus(index) {
      this.salesCodeFocusIndex = index;
    },
    // 销售员代码输入框失焦
    onSalesCodeBlur(index) {
      // 延迟一下，避免与自动跳转冲突
      setTimeout(() => {
        if (this.salesCodeFocusIndex === index) {
          this.salesCodeFocusIndex = -1;
        }
      }, 100);
    },
    // 关闭销售员代码输入
    closeSalesCodeInput() {
      // 如果所有输入框都为空，则直接关闭
      if (this.salesCodeArray.every(code => !code)) {
        this.showSalesCodeInput = false;
        this.formData.salesCode = '';
        this.salesCodeFocusIndex = -1;
      } else {
        uni.showModal({
          title: '提示',
          content: '已输入销售员代码，确定要关闭吗？',
          success: (res) => {
            if (res.confirm) {
              this.showSalesCodeInput = false;
              this.salesCodeArray = ['', '', '', '', ''];
              this.formData.salesCode = '';
              this.salesCodeFocusIndex = -1;
            }
          }
        });
      }
    },
    // 加载地址数据（编辑模式）
    async loadAddressData(addressId) {
      try {
        const res = await getMiniUserAddresses(this.userToken);
        if (res && res.code === 200 && res.data) {
          const address = res.data.find(addr => addr.id === addressId);
          if (address) {
            this.formData.name = address.name || '';
            this.formData.contact = address.contact || '';
            this.formData.phone = address.phone || '';
            this.formData.address = address.address || '';
            this.formData.storeType = address.store_type || '';
            // 不再从地址中读取销售员代码，因为销售员已改为和用户绑定
            this.formData.avatar = address.avatar || '';
            this.formData.isDefault = address.is_default || false;
            if (address.latitude) {
              this.mapLocation.latitude = address.latitude;
              this.formData.latitude = address.latitude;
            }
            if (address.longitude) {
              this.mapLocation.longitude = address.longitude;
              this.formData.longitude = address.longitude;
            }
            // 更新地图标记
            this.updateMarkers();
            this.showAddressFields = true;
          }
        }
      } catch (error) {
        console.error('加载地址数据失败:', error);
        uni.showToast({
          title: '加载地址数据失败',
          icon: 'none'
        });
      }
    },

    // 检查用户是否已有默认地址
    async checkDefaultAddress() {
      try {
        const res = await getMiniUserDefaultAddress(this.userToken);
        if (res && res.code === 200) {
          // 如果用户没有默认地址（res.data 为 null 或空），则默认选中"设置为默认地址"
          if (!res.data) {
            this.formData.isDefault = true;
          } else {
            // 如果用户已有默认地址，则默认不选中
            this.formData.isDefault = false;
          }
        } else {
          // API 调用失败，默认选中（第一次完善资料的情况）
          this.formData.isDefault = true;
        }
      } catch (error) {
        console.error('检查默认地址失败:', error);
        // 出错时默认选中（第一次完善资料的情况）
        this.formData.isDefault = true;
      }
    },

    // 默认地址开关变化
    onDefaultAddressChange(e) {
      this.formData.isDefault = e.detail.value;
    },

    // 导入微信收货地址
    importWeChatAddress() {
      uni.chooseAddress({
        success: (res) => {
          this.formData.name = res.userName || this.formData.name;
          this.formData.phone = res.telNumber || this.formData.phone;
          this.formData.address = res.detailInfo || this.formData.address;
          // 导入地址后显示地址和店铺类型字段
          this.showAddressFields = true;
          // 尝试获取地址的经纬度
          // 注意：微信小程序需要配置相关权限
        },
        fail: (err) => {
          console.log('导入地址失败:', err);
          uni.showToast({
            title: '导入地址失败',
            icon: 'none'
          });
        }
      });
    },
    async handleSubmit() {
      if (!this.canSubmit) {
        uni.showToast({
          title: '请填写必填项',
          icon: 'none'
        });
        return;
      }

      // 验证手机号格式
      const phoneRegex = /^1[3-9]\d{9}$/;
      if (!phoneRegex.test(this.formData.phone.trim())) {
        uni.showToast({
          title: '请输入正确的手机号码',
          icon: 'none'
        });
        return;
      }

      if (!this.userToken) {
        uni.showToast({
          title: '登录状态已过期，请重新登录',
          icon: 'none'
        });
        return;
      }

      // 如果地址已填写但没有经纬度，尝试自动解析
      // 注意：编辑模式下，如果用户没有重新选择定位，应该保持原有经纬度
      // 只有在新增模式或用户修改了地址但经纬度丢失的情况下才需要解析
      if (this.formData.address.trim() && (!this.formData.latitude || !this.formData.longitude)) {
        // 如果是编辑模式，提示用户重新选择定位，而不是自动解析（避免经纬度错误）
        if (this.formData.addressId) {
          uni.showModal({
            title: '提示',
            content: '检测到地址位置信息丢失，为了确保配送准确，请点击"选择店铺定位"按钮重新选择位置',
            showCancel: false,
            confirmText: '我知道了'
          });
          return;
        }

        // 新增模式才尝试自动解析
        uni.showLoading({ title: '正在解析地址...', mask: true });
        try {
          const geocodeRes = await geocodeAddress(this.formData.address.trim(), this.userToken);
          if (geocodeRes && geocodeRes.code === 200 && geocodeRes.data && geocodeRes.data.success) {
            this.formData.latitude = geocodeRes.data.latitude;
            this.formData.longitude = geocodeRes.data.longitude;
            uni.showToast({
              title: '地址解析成功',
              icon: 'success',
              duration: 1500
            });
          } else {
            // 解析失败，提示用户选择定位
            uni.hideLoading();
            uni.showModal({
              title: '提示',
              content: '无法自动获取地址位置，请点击"选择店铺定位"按钮手动选择位置，以确保配送准确',
              showCancel: false,
              confirmText: '我知道了'
            });
            return;
          }
        } catch (error) {
          console.error('地址解析失败:', error);
          uni.hideLoading();
          uni.showModal({
            title: '提示',
            content: '地址解析失败，请点击"选择店铺定位"按钮手动选择位置',
            showCancel: false,
            confirmText: '我知道了'
          });
          return;
        }
        uni.hideLoading();
      }

      // 如果仍然没有经纬度，强制要求用户选择定位
      if (!this.formData.latitude || !this.formData.longitude) {
        uni.showModal({
          title: '提示',
          content: '为了确保配送准确，请点击"选择店铺定位"按钮选择位置',
          showCancel: false,
          confirmText: '我知道了'
        });
        return;
      }

      this.submitting = true;
      try {
        const submitData = {
          address_id: this.formData.addressId || null,
          name: this.formData.name.trim(),
          contact: this.formData.contact.trim(),
          phone: this.formData.phone.trim(),
          address: this.formData.address.trim(),
          storeType: this.formData.storeType.trim(),
          is_default: this.formData.isDefault
        };

        // 如果用户未绑定销售员且输入了销售员代码，则绑定到用户
        if (!this.userHasSalesCode && this.salesCodeArray.join('').trim()) {
          submitData.salesCode = this.salesCodeArray.join('');
        }

        if (this.formData.avatar) {
          submitData.avatar = this.formData.avatar;
        }
        // 确保经纬度已设置
        if (this.formData.latitude) {
          submitData.latitude = this.formData.latitude;
        }
        if (this.formData.longitude) {
          submitData.longitude = this.formData.longitude;
        }

        const res = await updateMiniUserProfile(submitData, this.userToken);
        if (res && res.code === 200 && res.data) {
          uni.showToast({
            title: this.formData.addressId ? '更新成功' : '添加成功',
            icon: 'success'
          });
          setTimeout(() => {
            // 返回上一页
            uni.navigateBack({ delta: 1 });
          }, 600);
        }
      } catch (error) {
        console.error('提交资料失败:', error);
        uni.showToast({
          title: error?.message || '提交失败，请稍后重试',
          icon: 'none'
        });
      } finally {
        this.submitting = false;
      }
    },
    // 删除地址
    async handleDeleteAddress() {
      if (!this.formData.addressId) {
        return;
      }

      uni.showModal({
        title: '提示',
        content: '确定要删除这个地址吗？删除后无法恢复。',
        confirmText: '删除',
        cancelText: '取消',
        confirmColor: '#ff4d4f',
        success: async (res) => {
          if (res.confirm) {
            this.deleting = true;
            try {
              const result = await deleteMiniUserAddress(this.formData.addressId, this.userToken);
              if (result && result.code === 200) {
                uni.showToast({
                  title: '删除成功',
                  icon: 'success',
                  duration: 1500
                });
                // 延迟返回地址列表页面
                setTimeout(() => {
                  uni.navigateBack();
                }, 1500);
              } else {
                throw new Error(result?.message || '删除失败');
              }
            } catch (error) {
              console.error('删除地址失败:', error);
              uni.showToast({
                title: error?.message || '删除失败，请稍后再试',
                icon: 'none'
              });
            } finally {
              this.deleting = false;
            }
          }
        }
      });
    }
  }
};
</script>

<style scoped>
.profile-form-page {
  position: relative;
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

.map-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 20vh;
  z-index: 0;
}

.form-card {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: #fff;
  border-radius: 40rpx 40rpx 0 0;
  padding: 40rpx 32rpx;
  box-sizing: border-box;
  max-height: 85vh;
  overflow-y: auto;
  z-index: 10;
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.1);
}

.location-selector {
  display: flex;
  align-items: center;
  padding: 20rpx 32rpx;
  border: 2rpx solid #20CB6B;
  border-radius: 16rpx;
  margin-bottom: 40rpx;
  background-color: #fff;
  min-width: 0;
  box-sizing: border-box;
}

.location-icon {
  margin-right: 20rpx;
  flex-shrink: 0;
  margin-top: 4rpx;
}

.location-content {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  gap: 12rpx;
  min-width: 0;
  overflow: hidden;
}

.location-text {
  font-size: 32rpx;
  font-weight: 600;
  color: #20CB6B;
  flex: 1;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  max-width: 100%;
}


.form-item {
  margin-bottom: 32rpx;
}

.sales-code-item {
  margin-bottom: 20rpx;
  box-sizing: border-box;
}

.form-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.form-label {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 10rpx;
}

.required {
  color: #999;
  margin-left: 4rpx;
  font-weight: normal;
  font-size: 28rpx;
}

.photo-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
}

.photo-label {
  font-size: 28rpx;
  color: #20CB6B;
  font-weight: 500;
}

.photo-thumbnail {
  width: 60rpx;
  height: 60rpx;
  border-radius: 8rpx;
  border: 1rpx solid #e8e8e8;
}

.form-input {
  width: 100%;
  height: 88rpx;
  padding: 0 28rpx;
  background-color: #f8f9fa;
  border-radius: 16rpx;
  font-size: 30rpx;
  color: #222;
  box-sizing: border-box;
  border: 1rpx solid #e8e8e8;
}

.form-input::placeholder {
  color: #999;
  font-size: 30rpx;
}

.form-textarea {
  width: 100%;
  min-height: 120rpx;
  max-height: 300rpx;
  padding: 24rpx 28rpx;
  background-color: #f8f9fa;
  border-radius: 16rpx;
  font-size: 30rpx;
  color: #222;
  box-sizing: border-box;
  line-height: 1.6;
  border: 1rpx solid #e8e8e8;
  overflow-y: auto;
}

.form-textarea::placeholder {
  color: #999;
  font-size: 30rpx;
}

.picker-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.picker-input {
  padding-right: 60rpx;
}

.picker-icon {
  position: absolute;
  right: 28rpx;
}

.sales-code-link {
  width: 180rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10rpx 0;
  margin: 20rpx auto 0 auto;
  color: #20CB6B;
}

.link-text {
  font-size: 30rpx;
}

.sales-code-wrapper {
  margin-top: 20rpx;
  margin-bottom: 32rpx;
}

.sales-code-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
}

.sales-code-inputs {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16rpx;
}

.sales-code-box {
  width: 88rpx;
  height: 88rpx;
  background-color: #f8f9fa;
  border: 1rpx solid #e8e8e8;
  border-radius: 16rpx;
  text-align: center;
  font-size: 36rpx;
  font-weight: 600;
  color: #222;
  box-sizing: border-box;
}

.sales-code-close {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.default-address-item {
  margin-top: 20rpx;
}

.default-address-switch {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.form-footer {
  margin-top: 12rpx;
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.form-btn {
  width: 100%;
  height: 96rpx;
  line-height: 96rpx;
  text-align: center;
  border-radius: 48rpx;
  font-size: 34rpx;
  border: none;
}

.submit-btn {
  background: linear-gradient(120deg, #20cb6b, #16b35d);
  color: #fff;
}

.import-btn {
  background-color: #F8F8F8;
  color: #333;
}

.delete-text-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 8rpx;
}

.delete-text {
  font-size: 26rpx;
  color: #ff4d4f;
  text-decoration: none;
  cursor: pointer;
  padding: 8rpx 16rpx;
}

.delete-text:active {
  opacity: 0.7;
}

.delete-text.deleting {
  color: #ff9999;
  cursor: not-allowed;
}

/* 店铺类型选择器弹窗 */
.popup-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
}

.popup-content {
  width: 100%;
  background-color: #fff;
  border-radius: 40rpx 40rpx 0 0;
  padding: 40rpx 0;
  max-height: 60vh;
}

.popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 32rpx 30rpx 32rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.popup-title {
  font-size: 36rpx;
  font-weight: 600;
  color: #222;
}

.popup-close {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.popup-list {
  padding: 20rpx 0;
  max-height: 50vh;
}

.popup-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 32rpx;
  font-size: 32rpx;
  color: #333;
  border-bottom: 1rpx solid #f5f5f5;
}

.popup-item.active {
  color: #20CB6B;
  background-color: #f0fdf6;
}

/* 地址提示样式 */
.address-tip {
  display: flex;
  align-items: flex-start;
  gap: 8rpx;
  margin-top: 12rpx;
  padding: 16rpx;
  background-color: #f0fdf6;
  border-radius: 8rpx;
  border-left: 4rpx solid #20CB6B;
}

.tip-text {
  font-size: 24rpx;
  color: #606266;
  line-height: 1.6;
  flex: 1;
}
</style>
