<template>
  <view class="price-feedback-page">
    <!-- 自定义导航栏 -->
    <view class="custom-header">
      <view class="navbar-fixed" style="background-color: #20CB6B;">
        <!-- 状态栏撑起高度 -->
        <view :style="{ height: statusBarHeight + 'px' }"></view>
        <!-- 导航栏内容区域 -->
        <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
          <view class="navbar-left" @click="goBack">
            <uni-icons type="left" size="20" color="#fff"></uni-icons>
          </view>
          <view class="navbar-title">
            <text class="navbar-title-text">价格反馈</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 表单内容 -->
    <view class="form-container">
      <!-- 商品信息 -->
      <view class="form-section">
        <view class="section-title">商品信息</view>
        <view class="product-info-display">
          <text class="product-name">{{ productName }}</text>
          <text class="product-price">平台价格：{{ platformPriceDisplay }}</text>
        </view>
      </view>

      <!-- 反馈价格 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">您看到的价格 <text class="required">*</text></text>
          <input 
            v-model="formData.competitor_price" 
            class="form-input" 
            type="digit"
            placeholder="其他供应商价格（必填）" 
            maxlength="20"
          />
        </view>
      </view>

      <!-- 上传图片 -->
      <view class="form-section">
        <view class="section-title">上传其他供应商价格（最多3张）</view>
        <view class="image-upload-container">
          <view 
            class="image-item" 
            v-for="(image, index) in formData.images" 
            :key="index"
          >
            <image :src="image" mode="aspectFill" class="uploaded-image"></image>
            <view class="image-delete" @click="removeImage(index)">
              <uni-icons type="closeempty" size="20" color="#fff"></uni-icons>
            </view>
          </view>
          <view 
            class="image-upload-btn" 
            v-if="formData.images.length < 3"
            @click="chooseImage"
          >
            <uni-icons type="plusempty" size="40" color="#C0C4CC"></uni-icons>
            <text class="upload-text">上传图片</text>
          </view>
        </view>
      </view>

      <!-- 备注说明 -->
      <view class="form-section">
        <view class="form-item">
          <text class="form-label">备注说明</text>
          <textarea 
            v-model="formData.remark" 
            class="form-textarea" 
            placeholder="请输入其他说明信息（选填）" 
            maxlength="500"
            :auto-height="true"
          ></textarea>
        </view>
      </view>

      <!-- 提示信息 -->
      <view class="tip-section">
        <view class="tip-content">
          <uni-icons type="info" size="16" color="#FF9500"></uni-icons>
          <text class="tip-text">提交价格反馈后，我们会尽快核实并给您反馈，感谢您的支持！</text>
        </view>
      </view>
    </view>

    <!-- 底部提交按钮 -->
    <view class="bottom-button">
      <view class="submit-btn" @click="handleSubmit" :class="{ 'loading': submitting }">
        <text class="submit-btn-text">{{ submitting ? '提交中...' : '提交反馈' }}</text>
      </view>
    </view>
  </view>
</template>

<script>
import { createPriceFeedback } from '../../api/index.js';
import { uploadAddressAvatar } from '../../api/index.js';

export default {
  data() {
    return {
      statusBarHeight: 20,
      navBarHeight: 45,
      token: '',
      submitting: false,
      productId: 0,
      productName: '',
      platformPriceMin: 0,
      platformPriceMax: 0,
      platformPriceDisplay: '¥0.00',
      formData: {
        competitor_price: '',
        images: [],
        remark: ''
      }
    };
  },
  onLoad(options) {
    // 获取设备信息
    const info = uni.getSystemInfoSync();
    this.statusBarHeight = info.statusBarHeight;
    
    // 获取胶囊按钮信息并计算导航栏高度
    this.getMenuButtonInfo();
    
    // 检查登录状态
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
    this.token = token;
    
    // 获取传递的商品信息
    if (options.productId) {
      this.productId = parseInt(options.productId);
    }
    if (options.productName) {
      this.productName = decodeURIComponent(options.productName);
    }
    if (options.priceMin) {
      this.platformPriceMin = parseFloat(options.priceMin);
    }
    if (options.priceMax) {
      this.platformPriceMax = parseFloat(options.priceMax);
    }
    if (options.price) {
      this.platformPriceDisplay = decodeURIComponent(options.price);
    } else {
      // 如果没有传递显示价格，根据min和max生成
      if (this.platformPriceMin === this.platformPriceMax) {
        this.platformPriceDisplay = `¥${this.platformPriceMin.toFixed(2)}`;
      } else {
        this.platformPriceDisplay = `¥${this.platformPriceMin.toFixed(2)} - ¥${this.platformPriceMax.toFixed(2)}`;
      }
    }
  },
  methods: {
    // 获取胶囊按钮信息并计算导航栏高度
    getMenuButtonInfo() {
      try {
        // #ifndef H5 || APP-PLUS || MP-ALIPAY
        const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
        this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight);
        // #endif
      } catch (error) {
        console.error('获取胶囊按钮信息失败:', error);
      }
    },
    
    // 返回上一页
    goBack() {
      uni.navigateBack();
    },
    
    // 选择图片（参考个人资料页面的方式）
    chooseImage() {
      const remaining = 3 - this.formData.images.length;
      if (remaining <= 0) {
        uni.showToast({
          title: '最多只能上传3张图片',
          icon: 'none'
        });
        return;
      }
      
      uni.showActionSheet({
        itemList: ['拍摄', '从相册选择'],
        success: (res) => {
          const sourceType = res.tapIndex === 0 ? ['camera'] : ['album'];
          uni.chooseImage({
            count: remaining,
            sizeType: ['compressed'], // 使用压缩模式
            sourceType: sourceType,
            success: (chooseRes) => {
              // 上传图片
              this.compressAndUploadImages(chooseRes.tempFilePaths);
            },
            fail: (err) => {
              console.error('选择图片失败:', err);
            }
          });
        }
      });
    },
    
    // 上传图片（压缩后上传）
    async compressAndUploadImages(tempFilePaths) {
      uni.showLoading({ title: '处理中...' });
      
      try {
        for (const tempFilePath of tempFilePaths) {
          // 先压缩图片
          uni.showLoading({ title: '压缩中...' });
          const compressedPath = await this.compressImage(tempFilePath);
          
          // 再上传图片
          uni.showLoading({ title: '上传中...' });
          const imageUrl = await this.uploadImage(compressedPath);
          
          if (imageUrl) {
            this.formData.images.push(imageUrl);
          }
        }
        uni.showToast({
          title: '上传成功',
          icon: 'success'
        });
      } catch (error) {
        console.error('上传图片失败:', error);
        uni.showToast({
          title: error?.message || '图片上传失败，请重试',
          icon: 'none'
        });
      } finally {
        uni.hideLoading();
      }
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
    },
    
    // 上传图片到服务器（参考个人资料页面的上传方式）
    async uploadImage(filePath) {
      try {
        // 使用地址头像上传接口（这个接口可以正常上传图片）
        const res = await uploadAddressAvatar(filePath, this.token);
        if (res && res.code === 200 && res.data) {
          return res.data.avatar || res.data.imageUrl || res.data.url || '';
        } else {
          throw new Error(res?.message || '上传失败');
        }
      } catch (error) {
        console.error('上传图片失败:', error);
        throw error;
      }
    },
    
    // 删除图片
    removeImage(index) {
      this.formData.images.splice(index, 1);
    },
    
    // 提交反馈
    async handleSubmit() {
      if (this.submitting) return;
      
      // 验证必填字段
      if (!this.formData.competitor_price || this.formData.competitor_price.trim() === '') {
        uni.showToast({
          title: '请输入您看到的价格',
          icon: 'none'
        });
        return;
      }
      
      const price = parseFloat(this.formData.competitor_price);
      if (isNaN(price) || price <= 0) {
        uni.showToast({
          title: '请输入有效的价格',
          icon: 'none'
        });
        return;
      }
      
      if (!this.productId) {
        uni.showToast({
          title: '商品信息错误',
          icon: 'none'
        });
        return;
      }
      
      this.submitting = true;
      try {
        const res = await createPriceFeedback(this.token, {
          product_id: this.productId,
          product_name: this.productName,
          platform_price_min: this.platformPriceMin,
          platform_price_max: this.platformPriceMax,
          competitor_price: price,
          images: this.formData.images,
          remark: this.formData.remark.trim()
        });
        
        if (res && res.code === 200) {
          uni.showToast({
            title: '提交成功',
            icon: 'success'
          });
          setTimeout(() => {
            uni.navigateBack();
          }, 1500);
        } else {
          uni.showToast({
            title: res.message || '提交失败',
            icon: 'none'
          });
        }
      } catch (error) {
        console.error('提交价格反馈失败:', error);
        uni.showToast({
          title: '提交失败，请稍后再试',
          icon: 'none'
        });
      } finally {
        this.submitting = false;
      }
    }
  }
};
</script>

<style scoped>
.price-feedback-page {
  min-height: 100vh;
  background-color: #F5F6FA;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
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
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  background-color: #fff;
}

.navbar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 30rpx;
}

.navbar-left {
  width: 60rpx;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: flex-start;
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
  color: #fff;
}

.navbar-right {
  width: 60rpx;
  height: 100%;
}

/* 表单容器 */
.form-container {
  padding: 0 24rpx 20rpx 24rpx;
  padding-bottom: calc(60rpx + env(safe-area-inset-bottom));
}

.form-section {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 32rpx 28rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 24rpx;
}

/* 商品信息显示 */
.product-info-display {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.product-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
}

.product-price {
  font-size: 28rpx;
  color: #FF4D4F;
  font-weight: 600;
}

/* 表单项 */
.form-item {
  margin-bottom: 32rpx;
}

.form-item:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
  margin-bottom: 16rpx;
}

.required {
  color: #FF4D4F;
}

.form-input {
  width: 100%;
  height: 88rpx;
  padding: 0 24rpx;
  background-color: #F5F6FA;
  border-radius: 12rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}

.form-input::placeholder {
  color: #909399;
}

.form-textarea {
  width: 100%;
  min-height: 200rpx;
  padding: 24rpx;
  background-color: #F5F6FA;
  border-radius: 12rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
  line-height: 1.6;
}

.form-textarea::placeholder {
  color: #909399;
}

/* 图片上传 */
.image-upload-container {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
}

.image-item {
  position: relative;
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  overflow: hidden;
}

.uploaded-image {
  width: 100%;
  height: 100%;
}

.image-delete {
  position: absolute;
  top: 8rpx;
  right: 8rpx;
  width: 48rpx;
  height: 48rpx;
  background-color: rgba(0, 0, 0, 0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-upload-btn {
  width: 200rpx;
  height: 200rpx;
  border: 2rpx dashed #C0C4CC;
  border-radius: 12rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  background-color: #F5F6FA;
}

.upload-text {
  font-size: 24rpx;
  color: #909399;
}

/* 提示信息 */
.tip-section {
  margin-top: 20rpx;
  margin-bottom: 20rpx;
}

.tip-content {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
  padding: 24rpx;
  background-color: #FFF4E6;
  border-radius: 12rpx;
  border-left: 4rpx solid #FF9500;
}

.tip-text {
  flex: 1;
  font-size: 26rpx;
  color: #666;
  line-height: 1.6;
}

/* 底部按钮 */
.bottom-button {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 20rpx 30rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background-color: #fff;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}

.submit-btn {
  width: 100%;
  height: 88rpx;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  border-radius: 44rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 16rpx rgba(32, 203, 107, 0.3);
  transition: all 0.3s;
}

.submit-btn:active {
  opacity: 0.9;
  transform: scale(0.98);
}

.submit-btn.loading {
  opacity: 0.7;
}

.submit-btn-text {
  font-size: 32rpx;
  color: #fff;
  font-weight: 600;
}
</style>

