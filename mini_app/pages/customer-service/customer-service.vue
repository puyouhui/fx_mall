<template>
  <view class="customer-service-page">
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
            <text class="navbar-title-text">客服中心</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 顶部问候区域 -->
    <view class="header-section">
      <view class="greeting">
        <text class="greeting-text">Hi, 有什么可以帮您!</text>
      </view>
      <view class="service-time">
        <text class="time-label">人工客服接待时间:</text>
        <text class="time-value">8:30-21:30</text>
      </view>
      <view class="mascot">
        <view class="mascot-placeholder">
          <uni-icons type="chatbubble-filled" size="48" color="#20CB6B"></uni-icons>
          <text class="mascot-hi">hi</text>
        </view>
      </view>
    </view>

    <!-- 三个入口按钮 -->
    <view class="action-buttons">
      <view class="action-btn" @click="handleOnlineService">
        <view class="btn-icon online-service-icon">
          <uni-icons type="chatbubble" size="32" color="#20CB6B"></uni-icons>
        </view>
        <text class="btn-text">在线客服</text>
      </view>
      <view class="action-btn" @click="handleComplaint">
        <view class="btn-icon complaint-icon">
          <uni-icons type="email" size="32" color="#FF9500"></uni-icons>
        </view>
        <text class="btn-text">投诉</text>
      </view>
      <view class="action-btn" @click="handleFeedback">
        <view class="btn-icon feedback-icon">
          <uni-icons type="compose" size="32" color="#5AC8FA"></uni-icons>
        </view>
        <text class="btn-text">功能反馈</text>
      </view>
    </view>

    <!-- 第一联系人：销售员信息 -->
    <view class="sales-employee-section" v-if="salesEmployee">
      <view class="section-header">
        <view class="section-title-wrapper">
          <text class="section-title">专属销售经理</text>
          <text class="section-tip">请先联系销售经理，联系不上再联系平台客服</text>
        </view>
      </view>
      <view class="sales-employee-card" @click="callSalesEmployee">
        <view class="sales-employee-info">
          <view class="sales-employee-avatar">
            <uni-icons type="person-filled" size="32" color="#20CB6B"></uni-icons>
          </view>
          <view class="sales-employee-details">
            <text class="sales-employee-name">{{ salesEmployee.name }}</text>
            <text class="sales-employee-phone">{{ salesEmployee.phone }}</text>
          </view>
        </view>
        <view class="call-button">
          <uni-icons type="phone" size="26" color="#fff"></uni-icons>
        </view>
      </view>
    </view>

    <!-- 分类标签 -->
    <view class="category-tabs">
      <view 
        class="tab-item" 
        v-for="(category, index) in categories" 
        :key="index"
        :class="{ 'active': currentCategoryIndex === index }"
        @click="switchCategory(index)"
      >
        <text class="tab-text">{{ category.name }}</text>
      </view>
    </view>

    <!-- 问题列表 -->
    <view class="questions-section">
      <view 
        class="question-item" 
        v-for="(question, index) in currentQuestions" 
        :key="index"
        @click="handleQuestionClick(question)"
      >
        <text class="question-number">{{ index + 1 }}.</text>
        <text class="question-text">{{ question.title }}</text>
        <uni-icons type="right" size="16" color="#999"></uni-icons>
      </view>
    </view>

    <!-- 底部联系客服按钮 -->
    <view class="bottom-button">
      <view class="contact-btn" @click="handleContactService">
        <text class="contact-btn-text">联系客服</text>
      </view>
    </view>
  </view>
</template>

<script>
import { getMiniUserInfo } from '../../api/index.js';

export default {
  data() {
    return {
      statusBarHeight: 20, // 状态栏高度（默认值）
      navBarHeight: 45, // 导航栏高度（默认值）
      currentCategoryIndex: 0, // 当前选中的分类索引
      salesEmployee: null, // 销售员信息
      categories: [
        {
          name: '常见问题',
          questions: [
            { id: 1, title: '配送时效是怎么样的?', answer: '我们承诺在订单确认后24小时内发货，具体配送时间根据您所在地区而定，一般3-7个工作日可送达。' },
            { id: 2, title: '司机的联系方式是多少?', answer: '配送司机的联系方式会在订单发货后通过短信发送给您，您也可以在订单详情中查看。' },
            { id: 3, title: '有没有资质呀?', answer: '我们拥有完整的营业执照、食品经营许可证等相关资质，所有商品均经过严格的质量检测。' },
            { id: 4, title: '办理商品退货后,多久会退款?', answer: '退货商品确认收货后，我们会在3-5个工作日内完成退款，退款将原路返回到您的支付账户。' },
            { id: 5, title: '怎么开票?', answer: '您可以在下单时选择需要发票，填写发票信息。我们会在发货后7个工作日内将发票寄出。' },
            { id: 6, title: '充值什么时候到账?', answer: '充值成功后，资金会立即到账。如遇延迟，请稍等片刻或联系客服处理。' },
            { id: 7, title: '业务员/司机服务不好,怎么反馈?', answer: '您可以通过"投诉"入口提交反馈，我们会认真处理并及时回复。' },
            { id: 8, title: '之前地址在配送范围,突然又不在了?', answer: '配送范围可能会根据实际情况进行调整。如您的地址不在配送范围内，请联系客服，我们会尽力为您解决。' }
          ]
        },
        {
          name: '商品问题',
          questions: [
            { id: 9, title: '商品质量有问题怎么办?', answer: '如发现商品质量问题，请在收货后7天内联系客服，我们会为您办理退换货。' },
            { id: 10, title: '商品与描述不符怎么办?', answer: '如商品与描述不符，您可以申请退货退款，我们会在核实后为您处理。' },
            { id: 11, title: '如何查看商品保质期?', answer: '商品保质期信息会在商品详情页和商品包装上标注，您可以在下单前查看。' },
            { id: 12, title: '商品缺货怎么办?', answer: '如商品缺货，我们会及时通知您，您可以选择等待补货或申请退款。' },
            { id: 13, title: '可以批量购买吗?', answer: '可以，您可以在商品详情页选择数量，或联系客服咨询批量采购优惠。' }
          ]
        },
        {
          name: '配送物流',
          questions: [
            { id: 14, title: '配送范围包括哪些地区?', answer: '我们目前覆盖全国大部分城市，具体配送范围可在下单时查看，或联系客服咨询。' },
            { id: 15, title: '如何查询物流信息?', answer: '您可以在订单详情中查看物流信息，或通过物流单号在快递公司官网查询。' },
            { id: 16, title: '可以指定配送时间吗?', answer: '您可以在下单时备注期望的配送时间，我们会尽量安排，但不保证一定能满足。' },
            { id: 17, title: '配送费用如何计算?', answer: '配送费用根据订单金额和配送距离计算，具体费用会在结算时显示。' },
            { id: 18, title: '商品损坏了怎么办?', answer: '如商品在配送过程中损坏，请拒收并联系客服，我们会重新发货或退款。' }
          ]
        },
        {
          name: '售后',
          questions: [
            { id: 19, title: '退货流程是什么?', answer: '您可以在订单详情中申请退货，填写退货原因，我们审核通过后会安排退货。' },
            { id: 20, title: '退款多久到账?', answer: '退货商品确认收货后，我们会在3-5个工作日内完成退款，退款将原路返回到您的支付账户。' },
            { id: 21, title: '可以换货吗?', answer: '可以，如商品有质量问题或与描述不符，您可以申请换货，我们会在核实后为您处理。' },
            { id: 22, title: '售后时效是多久?', answer: '我们提供7天无理由退货服务，质量问题可享受更长的售后保障。' },
            { id: 23, title: '如何联系售后?', answer: '您可以通过"在线客服"或"联系客服"按钮联系我们的售后团队，我们会尽快为您处理。' }
          ]
        }
      ]
    };
  },
  onLoad() {
    // 获取设备信息
    const info = uni.getSystemInfoSync();
    // 设置状态栏高度
    this.statusBarHeight = info.statusBarHeight;
    
    // 获取胶囊按钮信息并计算导航栏高度
    this.getMenuButtonInfo();
    
    // 加载用户信息和销售员信息
    this.loadUserInfo();
  },
  computed: {
    // 当前分类的问题列表
    currentQuestions() {
      if (this.categories[this.currentCategoryIndex]) {
        return this.categories[this.currentCategoryIndex].questions || [];
      }
      return [];
    }
  },
  methods: {
    // 获取胶囊按钮信息并计算导航栏高度
    getMenuButtonInfo() {
      try {
        // #ifndef H5 || APP-PLUS || MP-ALIPAY
        // 获取胶囊的位置信息
        const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
        // 按照参考文章的公式计算导航栏高度：
        // (胶囊底部高度 - 状态栏的高度) + (胶囊顶部高度 - 状态栏内的高度) = 导航栏的高度
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
    
    // 切换分类
    switchCategory(index) {
      this.currentCategoryIndex = index;
    },
    
    // 处理问题点击
    handleQuestionClick(question) {
      uni.showModal({
        title: question.title,
        content: question.answer,
        showCancel: false,
        confirmText: '知道了',
        confirmColor: '#20CB6B'
      });
    },
    
    // 在线客服
    handleOnlineService() {
      uni.showToast({
        title: '正在为您转接在线客服...',
        icon: 'none',
        duration: 2000
      });
      // TODO: 接入在线客服系统
    },
    
    // 投诉
    handleComplaint() {
      uni.showModal({
        title: '投诉',
        content: '投诉功能开发中，您可以通过"联系客服"按钮联系我们。',
        showCancel: false,
        confirmText: '知道了',
        confirmColor: '#20CB6B'
      });
      // TODO: 跳转到投诉页面
    },
    
    // 功能反馈
    handleFeedback() {
      uni.showModal({
        title: '功能反馈',
        content: '功能反馈功能开发中，您可以通过"联系客服"按钮联系我们。',
        showCancel: false,
        confirmText: '知道了',
        confirmColor: '#20CB6B'
      });
      // TODO: 跳转到功能反馈页面
    },
    
    // 联系客服
    handleContactService() {
      // 检查当前时间是否在服务时间内
      const now = new Date();
      const hours = now.getHours();
      const minutes = now.getMinutes();
      const currentTime = hours * 60 + minutes;
      const startTime = 8 * 60 + 30; // 8:30
      const endTime = 21 * 60 + 30; // 21:30
      
      if (currentTime >= startTime && currentTime <= endTime) {
        // 服务时间内，跳转到在线客服
        this.handleOnlineService();
      } else {
        // 服务时间外，提示用户
        uni.showModal({
          title: '提示',
          content: '当前不在服务时间内（8:30-21:30），您可以留言，我们会在服务时间内尽快回复您。',
          confirmText: '去留言',
          cancelText: '取消',
          confirmColor: '#20CB6B',
          success: (res) => {
            if (res.confirm) {
              // TODO: 跳转到留言页面
              this.handleOnlineService();
            }
          }
        });
      }
    },
    
    // 加载用户信息和销售员信息
    async loadUserInfo() {
      try {
        const token = uni.getStorageSync('miniUserToken');
        if (!token) {
          return;
        }
        
        const res = await getMiniUserInfo(token);
        if (res && res.code === 200 && res.data) {
          // 获取销售员信息
          if (res.data.sales_employee) {
            this.salesEmployee = res.data.sales_employee;
          }
        }
      } catch (error) {
        console.error('获取用户信息失败:', error);
      }
    },
    
    // 拨打销售员电话
    callSalesEmployee() {
      if (!this.salesEmployee || !this.salesEmployee.phone) {
        uni.showToast({
          title: '销售员电话不可用',
          icon: 'none'
        });
        return;
      }
      
      uni.makePhoneCall({
        phoneNumber: this.salesEmployee.phone,
        success: () => {
          console.log('拨打电话成功');
        },
        fail: (err) => {
          console.error('拨打电话失败:', err);
          uni.showToast({
            title: '拨打电话失败',
            icon: 'none'
          });
        }
      });
    }
  }
};
</script>

<style scoped>
.customer-service-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #E8F8F0 0%, #E8F8F0 15%, #F5FCF8 30%, #FFFFFF 60%);
  padding-bottom: calc(128rpx + env(safe-area-inset-bottom));
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
  color: #333;
}

.navbar-right {
  width: 60rpx;
  height: 100%;
}

/* 顶部问候区域 */
.header-section {
  position: relative;
  padding: 40rpx 30rpx 30rpx;
  background: transparent;
}

.greeting {
  margin-bottom: 20rpx;
}

.greeting-text {
  font-size: 36rpx;
  font-weight: 600;
  color: #333;
  line-height: 1.5;
}

.service-time {
  display: flex;
  align-items: center;
  margin-bottom: 20rpx;
}

.time-label {
  font-size: 26rpx;
  color: #666;
  margin-right: 10rpx;
}

.time-value {
  font-size: 26rpx;
  color: #20CB6B;
  font-weight: 500;
}

.mascot {
  position: absolute;
  right: 30rpx;
  top: 40rpx;
  width: 120rpx;
  height: 120rpx;
}

.mascot-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #E8F8F0 0%, #D4F4E0 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 12rpx rgba(32, 203, 107, 0.2);
  position: relative;
}

.mascot-hi {
  position: absolute;
  top: 48%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 24rpx;
  font-weight: 600;
  color: #fff;
  z-index: 1;
}

/* 三个入口按钮 */
.action-buttons {
  display: flex;
  justify-content: space-around;
  padding: 30rpx;
  background-color: #fff;
  margin: 0 20rpx 20rpx;
  border-radius: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.btn-icon {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16rpx;
}

.online-service-icon {
  background-color: #E8F8F0;
}

.complaint-icon {
  background-color: #FFF4E6;
}

.feedback-icon {
  background-color: #E6F7FF;
}

.btn-text {
  font-size: 26rpx;
  color: #333;
  font-weight: 500;
}

/* 销售员信息区域 */
.sales-employee-section {
  background-color: #fff;
  margin: 0 20rpx 20rpx;
  border-radius: 20rpx;
  padding: 30rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.section-header {
  margin-bottom: 20rpx;
}

.section-title-wrapper {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}

.section-tip {
  font-size: 24rpx;
  color: #999;
  line-height: 1.5;
}

.sales-employee-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx;
  background-color: #F0FDF6;
  border-radius: 16rpx;
  border: 1rpx solid #E8F8F0;
  transition: all 0.3s;
}

.sales-employee-card:active {
  background-color: #E8F8F0;
  transform: scale(0.98);
}

.sales-employee-info {
  display: flex;
  align-items: center;
  flex: 1;
}

.sales-employee-avatar {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background-color: #E8F8F0;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 24rpx;
  flex-shrink: 0;
}

.sales-employee-details {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.sales-employee-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 8rpx;
}

.sales-employee-phone {
  font-size: 26rpx;
  color: #666;
}

.call-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80rpx;
  height: 80rpx;
  background-color: #20CB6B;
  border-radius: 50%;
  flex-shrink: 0;
  transition: all 0.3s;
}

.call-button:active {
  background-color: #18B85A;
  transform: scale(0.95);
}

/* 分类标签 */
.category-tabs {
  display: flex;
  padding: 20rpx 30rpx;
  background-color: #fff;
  margin: 0 20rpx 20rpx;
  border-radius: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
  overflow-x: auto;
  white-space: nowrap;
}

.tab-item {
  display: inline-block;
  padding: 12rpx 24rpx;
  margin-right: 20rpx;
  border-radius: 30rpx;
  background-color: #F5F5F5;
  transition: all 0.3s;
}

.tab-item.active {
  background-color: #20CB6B;
}

.tab-text {
  font-size: 26rpx;
  color: #666;
}

.tab-item.active .tab-text {
  color: #fff;
  font-weight: 500;
}

/* 问题列表 */
.questions-section {
  background-color: #fff;
  margin: 0 20rpx;
  border-radius: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
  overflow: hidden;
}

.question-item {
  display: flex;
  align-items: center;
  padding: 28rpx 30rpx;
  border-bottom: 1rpx solid #F5F5F5;
  transition: background-color 0.2s;
}

.question-item:last-child {
  border-bottom: none;
}

.question-item:active {
  background-color: #F8F8F8;
}

.question-number {
  font-size: 28rpx;
  color: #20CB6B;
  font-weight: 500;
  margin-right: 16rpx;
  min-width: 40rpx;
}

.question-text {
  flex: 1;
  font-size: 28rpx;
  color: #333;
  line-height: 1.6;
}

/* 底部联系客服按钮 */
.bottom-button {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 20rpx 30rpx;
  padding-bottom: calc(env(safe-area-inset-bottom));
  background-color: #fff;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}

.contact-btn {
  width: 100%;
  height: 88rpx;
  background: linear-gradient(135deg, #20CB6B 0%, #18B85A 100%);
  border-radius: 44rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 16rpx rgba(32, 203, 107, 0.3);
}

.contact-btn:active {
  opacity: 0.9;
  transform: scale(0.98);
}

.contact-btn-text {
  font-size: 32rpx;
  color: #fff;
  font-weight: 600;
}
</style>

