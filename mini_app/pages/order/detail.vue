<template>
  <view class="order-detail-page">
    <!-- 自定义导航栏 - 绿色背景 -->
    <view class="custom-navbar">
      <view class="navbar-fixed" style="background-color: #20CB6B;">
        <view :style="{ height: statusBarHeight + 'px' }"></view>
        <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
          <view class="navbar-left" @click="goBack">
            <uni-icons type="left" size="20" color="#fff"></uni-icons>
          </view>
          <view class="navbar-title">
            <text class="navbar-title-text">订单详情</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 顶部渐变背景或地图区域 -->
    <view 
      class="top-gradient-section"
      :style="{ 
        paddingTop: `${statusBarHeight + navBarHeight}px`
      }"
      v-if="orderDetail"
    >
      <!-- 地图（仅在配送中状态显示，配送员取货后才显示） -->
      <view class="map-section" v-if="showMap">
        <map
          :latitude="mapCenter.latitude"
          :longitude="mapCenter.longitude"
          :markers="mapMarkers"
          :scale="mapScale"
          class="map-container"
          :show-location="true"
          :enable-zoom="true"
          :enable-scroll="true"
        ></map>
        <view class="map-refresh-btn" @click="refreshDeliveryLocation">
          <uni-icons type="reload" size="20" color="#20CB6B"></uni-icons>
        </view>
      </view>
      
      <!-- 渐变背景（不显示地图时） -->
      <view class="gradient-background" v-else>
        <view class="status-content">
          <view class="status-left">
            <view class="status-icon-circle">
              <uni-icons :type="getStatusIcon(orderDetail.order?.status)" size="30" color="#fff"></uni-icons>
            </view>
            <view class="status-text-group">
              <text class="status-main-text">{{ formatStatus(orderDetail.order?.status) }}</text>
              <view class="status-tag" v-if="orderDetail.order?.order_type">
                <text>{{ orderDetail.order.order_type }}</text>
              </view>
              <view class="payment-countdown" v-if="showPaymentCountdown">
                <text class="countdown-label">剩余支付时间</text>
                <text class="countdown-value">{{ paymentCountdownText }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>
    </view>

    <view 
      class="detail-content" 
      v-if="orderDetail"
      :style="{ 
        paddingTop: showMap ? '0' : '2rpx',
        paddingBottom: showActionFooter ? '180rpx' : '80rpx'
      }"
    >
      <!-- 收货地址 -->
      <view class="section address-section" v-if="orderDetail.address">
        <view class="address-content">
          <view class="address-main">
            <view class="address-header">
              <view class="address-title-row">
                <text class="address-store">{{ orderDetail.address.name || '收货地址' }}</text>
              </view>
              <view class="address-contact-row">
                <text class="address-contact">{{ orderDetail.address.contact }}</text>
                <text class="address-phone">{{ orderDetail.address.phone }}</text>
              </view>
              <view class="address-detail">{{ orderDetail.address.address }}</view>
            </view>
          </view>
        </view>
      </view>

      <!-- 商品列表 -->
      <view class="section goods-section">
        <view class="section-title">商品信息</view>
        <view class="goods-list">
          <view 
            class="goods-item" 
            v-for="(item, index) in orderDetail.order_items" 
            :key="index"
          >
            <image :src="item.image || defaultImage" class="goods-image" mode="aspectFill" />
            <view class="goods-info">
              <view class="goods-name-row">
                <text class="goods-name">{{ item.product_name }}</text>
              </view>
              <text class="goods-spec" v-if="item.spec_name">{{ item.spec_name }}</text>
              <view class="goods-bottom">
                <text class="goods-price">¥{{ formatMoney(item.unit_price) }}</text>
                <text class="goods-qty">× {{ item.quantity }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- 加急订单显示 -->
      <view class="section urgent-section" v-if="orderDetail.order?.is_urgent">
        <view class="urgent-container urgent-active">
          <view class="urgent-left">
            <view class="urgent-header">
              <text class="urgent-title">加急订单</text>
              <text class="urgent-tag">平台将为您加急配送</text>
            </view>
          </view>
          <view class="urgent-right">
            <view class="urgent-price-wrapper" v-if="orderDetail.order?.urgent_fee > 0">
              <text class="urgent-price">¥{{ formatMoney(orderDetail.order?.urgent_fee) }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 金额明细 -->
      <view class="section amount-section">
        <view class="section-title">金额明细</view>
        <view class="amount-row">
          <text class="amount-label">商品金额</text>
          <text class="amount-value">¥{{ formatMoney(orderDetail.order?.goods_amount) }}</text>
        </view>
        <view class="amount-row">
          <text class="amount-label">配送费</text>
          <text class="amount-value">¥{{ formatMoney(orderDetail.order?.delivery_fee) }}</text>
        </view>
        <view class="amount-row" v-if="orderDetail.order?.points_discount > 0">
          <text class="amount-label">积分抵扣</text>
          <text class="amount-value discount-text">-¥{{ formatMoney(orderDetail.order?.points_discount) }}</text>
        </view>
        <view class="amount-row" v-if="orderDetail.order?.coupon_discount > 0">
          <text class="amount-label">优惠券</text>
          <text class="amount-value discount-text">-¥{{ formatMoney(orderDetail.order?.coupon_discount) }}</text>
        </view>
        <view class="amount-row urgent-fee-row" v-if="orderDetail.order?.is_urgent && orderDetail.order?.urgent_fee > 0">
          <view class="urgent-fee-label-wrapper">
            <text class="amount-label urgent-fee-label">加急费用</text>
            <text class="urgent-fee-tag">将优先为您配送</text>
          </view>
          <text class="amount-value urgent-fee-value">¥{{ formatMoney(orderDetail.order?.urgent_fee) }}</text>
        </view>
        <view class="amount-divider"></view>
        <view class="amount-row total-row">
          <text class="amount-label total-label">实付金额</text>
          <text class="amount-value total-value">¥{{ formatMoney(orderDetail.order?.total_amount) }}</text>
        </view>
      </view>

      <!-- 订单信息 -->
      <view class="section info-section">
        <view class="section-title">订单信息</view>
        <view class="info-row">
          <text class="info-label">订单编号</text>
          <text class="info-value">{{ orderDetail.order?.order_number }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">订单状态</text>
          <text class="info-value" :class="getStatusClass(orderDetail.order?.status)">
            {{ formatStatusShort(orderDetail.order?.status) }}
          </text>
        </view>
        <view class="info-row">
          <text class="info-label">下单时间</text>
          <text class="info-value">{{ formatDate(orderDetail.order?.created_at) }}</text>
        </view>
      </view>

      <!-- 订单备注 -->
      <view class="section remark-section" v-if="orderDetail.order?.remark">
        <view class="remark-header">
          <text class="section-title">订单备注</text>
        </view>
        <view class="remark-content">
          <text class="remark-text">{{ orderDetail.order?.remark }}</text>
        </view>
      </view>

      <!-- 其他选项 -->
      <view class="section options-section">
        <view class="section-title">其他选项</view>
        <view class="option-row" v-if="orderDetail.order?.out_of_stock_strategy">
          <view class="option-text">
            <text class="option-title">缺货处理</text>
            <text class="option-desc">遇到缺货时的处理方式</text>
          </view>
          <view class="option-status">
            <text class="option-status-value">{{ getOutOfStockStrategyText(orderDetail.order?.out_of_stock_strategy) }}</text>
          </view>
        </view>
        <view class="option-row" v-if="orderDetail.order?.trust_receipt !== undefined">
          <view class="option-text">
            <text class="option-title">信任签收</text>
            <text class="option-desc">配送电话联系不上时，允许放门口或指定位置</text>
          </view>
          <view class="option-status">
            <text v-if="orderDetail.order?.trust_receipt" class="option-status-active">已开启</text>
            <text v-else class="option-status-text">未开启</text>
          </view>
        </view>
        <view class="option-row" v-if="orderDetail.order?.hide_price !== undefined">
          <view class="option-text">
            <text class="option-title">隐藏价格</text>
            <text class="option-desc">选择后，小票中将不显示商品价格</text>
          </view>
          <view class="option-status">
            <text v-if="orderDetail.order?.hide_price" class="option-status-active">已开启</text>
            <text v-else class="option-status-text">未开启</text>
          </view>
        </view>
        <view class="option-row" v-if="orderDetail.order?.require_phone_contact !== undefined">
          <view class="option-text">
            <text class="option-title">配送时电话联系</text>
            <text class="option-desc">建议保持电话畅通，方便配送员联系</text>
          </view>
          <view class="option-status">
            <text v-if="orderDetail.order?.require_phone_contact" class="option-status-active">已开启</text>
            <text v-else class="option-status-text">未开启</text>
          </view>
        </view>
      </view>

      <!-- 客服提示 -->
      <view class="customer-service-tip" @click="goToCustomerService">
        <view class="service-avatar">
          <uni-icons type="chatbubble" size="20" color="#20CB6B"></uni-icons>
        </view>
        <text class="service-text">有问题不能解决？点我试试~</text>
      </view>

    </view>

    <!-- 底部操作按钮：左侧图标操作，右侧主按钮（按状态统一展示） -->
    <view class="action-footer" v-if="orderDetail && showActionFooter">
      <view class="action-footer-container">
        <view class="action-footer-left">
          <view class="action-icon-btn" @click="goToCustomerService">
            <uni-icons type="chat" size="28" color="#2C2C2C"></uni-icons>
            <text class="action-icon-text">客服</text>
          </view>
          <view 
            v-if="canCancelOrder"
            class="action-icon-btn" 
            @click="handleCancelOrder"
          >
            <uni-icons type="closeempty" size="28" color="#2C2C2C"></uni-icons>
            <text class="action-icon-text">取消</text>
          </view>
        </view>
        <view class="action-footer-right">
          <template v-if="hasMainAction">
            <view 
              class="action-main-btn" 
              :class="mainBtnClass"
              v-if="showPayBtn"
              @click="handlePayOrder"
            >
              <text>{{ paying ? '支付中...' : '去付款' }}</text>
            </view>
            <view 
              class="action-main-btn" 
              :class="mainBtnClass"
              v-else-if="showContactDeliveryBtn"
              @click="contactDelivery"
            >
              <text>联系配送员</text>
            </view>
            <view 
              class="action-main-btn" 
              :class="mainBtnClass"
              v-else-if="showConfirmReceiveBtn"
              @click="handleOpenConfirmReceive"
            >
              <text>{{ confirmReceiveLoading ? '打开中...' : '确认收货' }}</text>
            </view>
          </template>
          <view v-else class="action-main-btn" @click="goToCustomerService">
            <text>联系我们</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 加载中提示（仅在订单详情未加载时显示） -->
    <view 
      class="loading" 
      v-if="!orderDetail"
      :style="{ 
        height: `calc(100vh - ${statusBarHeight + navBarHeight}px)`,
        marginTop: `${statusBarHeight + navBarHeight}px`
      }"
    >
      <text>加载中...</text>
    </view>
  </view>
</template>

<script>
import { getOrderDetail, getDeliveryEmployeeLocation, cancelOrder, getWechatPayPrepay, getWechatConfirmReceiveInfo } from '../../api/index.js'
import { getShareConfig, buildSharePath } from '../../utils/shareConfig.js'

export default {
  data() {
    return {
      statusBarHeight: 0,
      navBarHeight: 44,
      orderDetail: null,
      orderId: 0,
      token: '',
      defaultImage: 'https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg',
      mapCenter: {
        latitude: 39.90864,
        longitude: 116.39750
      },
      mapMarkers: [],
      mapScale: 6, // 地图缩放级别
      paying: false, // 支付中
      paymentDeadlineAt: null, // 支付截止时间 ISO 字符串
      paymentCountdownText: '--:--',
      countdownTimer: null,
      confirmReceiveLoading: false,
      confirmReceiveDone: false, // 确认收货成功后不再显示底部栏
      fromPayment: false, // 支付成功跳转，需轮询等待订单创建
      fromSubmit: false, // 从提交订单/支付成功进入，返回时回首页
      paymentPollTimer: null,
      paymentPollCount: 0
    }
  },
  computed: {
    showPaymentCountdown() {
      return this.orderDetail?.order?.status === 'pending_payment' && this.paymentDeadlineAt
    },
    showMap() {
      const status = this.orderDetail?.order?.status
      // 地图只在配送中状态显示（配送员取货后才显示）
      return status === 'delivering'
    },
    // 是否可以取消订单（待支付、配送员接单之前：pending_payment、pending_delivery、pending_pickup）
    canCancelOrder() {
      const status = this.orderDetail?.order?.status
      return status === 'pending_payment' || status === 'pending_delivery' || status === 'pending' || status === 'pending_pickup'
    },
    // 是否显示配送员信息（接单后到配送完时显示）
    showDeliveryEmployee() {
      const status = this.orderDetail?.order?.status
      // 配送员接单后的状态：pending_pickup, delivering, delivered, shipped, paid, completed
      return status === 'pending_pickup' || 
             status === 'delivering' || 
             status === 'delivered' || 
             status === 'shipped' || 
             status === 'paid' || 
             status === 'completed'
    },
    // 是否显示联系配送员按钮（仅配送中状态）
    showContactDeliveryBtn() {
      const status = this.orderDetail?.order?.status
      // 只有 delivering 状态显示联系配送员按钮
      return status === 'delivering' && 
             this.orderDetail?.delivery_employee?.phone
    },
    // 是否显示去付款按钮（未支付且未取消的订单）
    showPayBtn() {
      const order = this.orderDetail?.order
      if (!order) return false
      if (order.status === 'cancelled') return false
      if (order.status === 'paid' || order.paid_at) return false
      return Number(order.total_amount || 0) > 0
    },
    // 是否显示确认收货按钮（仅微信支付订单且已送达/已收款，且未完成确认收货）
    showConfirmReceiveBtn() {
      const order = this.orderDetail?.order
      if (!order || this.confirmReceiveDone) return false
      const statusOk = order.status === 'delivered' || order.status === 'shipped' || order.status === 'paid'
      const isWechatPay = order.payment_method === 'online'
      return statusOk && isWechatPay
    },
    // 是否显示底部操作按钮
    showActionFooter() {
      return this.canCancelOrder || 
             this.showContactDeliveryBtn ||
             this.showPayBtn ||
             this.showConfirmReceiveBtn
    },
    // 是否有右侧主按钮（用于底部栏布局与样式统一）
    hasMainAction() {
      return this.showPayBtn || this.showContactDeliveryBtn || this.showConfirmReceiveBtn
    },
    // 主按钮统一 class，便于按状态扩展样式
    mainBtnClass() {
      if (this.showPayBtn) return 'action-main-btn--pay'
      if (this.showContactDeliveryBtn) return 'action-main-btn--contact'
      if (this.showConfirmReceiveBtn) return 'action-main-btn--confirm'
      return ''
    }
  },
  onLoad(options) {
    const systemInfo = uni.getSystemInfoSync()
    this.statusBarHeight = systemInfo.statusBarHeight || 0
    
    // 计算导航栏高度（状态栏高度 + 导航栏内容高度）
    try {
      const menuButtonInfo = uni.getMenuButtonBoundingClientRect()
      this.navBarHeight = (menuButtonInfo.bottom - this.statusBarHeight) + (menuButtonInfo.top - this.statusBarHeight)
    } catch (e) {
      this.navBarHeight = 44
    }
    
    this.token = uni.getStorageSync('miniUserToken')
    // 支持订单ID（数字）或订单编号（从「小程序购物订单」跳转时微信用 order_number 作为 id）
    const idParam = options.id || options.scene || ''
    this.orderId = idParam
    this.fromPayment = options.fromPayment === '1'
    this.fromSubmit = options.fromSubmit === '1'
    if (!idParam) {
      uni.showToast({
        title: '订单参数无效',
        icon: 'none'
      })
      setTimeout(() => {
        this.backOrToHome()
      }, 1500)
      return
    }
    
    this.loadOrderDetail()
    // 监听微信确认收货组件回调
    uni.$on('wechatConfirmReceiveDone', this.onWechatConfirmReceiveDone)
  },
  onUnload() {
    this.clearCountdownTimer()
    this.clearPaymentPoll()
    uni.$off('wechatConfirmReceiveDone', this.onWechatConfirmReceiveDone)
  },
  // 分享小程序（订单详情页）
  onShareAppMessage(options) {
    // 使用 shareConfig 获取分享配置
    const shareConfig = getShareConfig('order', {
      orderNumber: this.orderDetail?.order?.order_number || ''
    });
    
    // 构建分享路径，优先使用订单编号（与「小程序购物订单」跳转一致），否则用 id
    const shareId = this.orderDetail?.order?.order_number || this.orderId
    const path = buildSharePath(`/pages/order/detail?id=${shareId}`)
    
    return {
      title: shareConfig.title,
      path: path,
      imageUrl: shareConfig.imageUrl || ''
    };
  },
  methods: {
    goBack() {
      this.clearCountdownTimer()
      this.backOrToHome()
    },
    /** 从提交/支付进入则回首页，否则返回上一页 */
    backOrToHome() {
      if (this.fromSubmit || this.fromPayment) {
        uni.reLaunch({ url: '/pages/index/index' })
      } else {
        uni.navigateBack()
      }
    },
    onWechatConfirmReceiveDone(payload) {
      if (!payload || !this.orderDetail?.order) return
      const orderNumber = this.orderDetail.order.order_number
      const match = payload.merchant_trade_no === orderNumber || String(this.orderId) === String(orderNumber)
      if (!match) return
      if (payload.status === 'success') {
        this.confirmReceiveDone = true
        uni.showToast({ title: '确认收货成功', icon: 'success' })
        this.loadOrderDetail()
      } else if (payload.status === 'fail') {
        uni.showToast({ title: payload.errormsg || '确认收货失败', icon: 'none' })
      }
    },
    startPaymentCountdown() {
      this.clearCountdownTimer()
      if (!this.paymentDeadlineAt || this.orderDetail?.order?.status !== 'pending_payment') return
      const dateStr = String(this.paymentDeadlineAt).replace(/-/g, '/').replace('T', ' ')
      const deadline = new Date(dateStr)
      if (isNaN(deadline.getTime())) {
        this.paymentCountdownText = '--:--'
        return
      }
      const updateCountdown = () => {
        const now = new Date()
        const diff = Math.max(0, Math.floor((deadline - now) / 1000))
        if (diff <= 0) {
          this.paymentCountdownText = '已超时'
          this.clearCountdownTimer()
          this.loadOrderDetail()
          return
        }
        const m = Math.floor(diff / 60)
        const s = diff % 60
        this.paymentCountdownText = `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
      }
      updateCountdown()
      this.countdownTimer = setInterval(updateCountdown, 1000)
    },
    clearCountdownTimer() {
      if (this.countdownTimer) {
        clearInterval(this.countdownTimer)
        this.countdownTimer = null
      }
    },
    async loadOrderDetail() {
      try {
        uni.showLoading({ title: this.fromPayment ? '订单生成中...' : '加载中...' })
        const res = await getOrderDetail(this.token, this.orderId, this.fromPayment ? { silent: true } : {})
        if (res && res.code === 200 && res.data) {
          this.orderDetail = res.data
          this.paymentDeadlineAt = res.data.payment_deadline_at || null
          this.startPaymentCountdown()
          // 调试：打印订单详情
          console.log('订单详情:', JSON.stringify(this.orderDetail, null, 2))
          console.log('订单状态:', this.orderDetail?.order?.status)
          console.log('配送员信息:', this.orderDetail?.delivery_employee)
          // 初始化地图
          this.initMap()
          // 如果是配送中状态，加载配送员位置
          if (this.orderDetail.order?.status === 'delivering' && this.orderDetail.delivery_employee?.employee_code) {
            this.loadDeliveryEmployeeLocation()
          }
        } else {
          if (this.fromPayment) {
            this.startPaymentPoll()
          } else {
            uni.showToast({ title: res?.message || '获取订单详情失败', icon: 'none' })
            setTimeout(() => this.backOrToHome(), 1500)
          }
        }
      } catch (error) {
        console.error('获取订单详情失败:', error)
        if (this.fromPayment) {
          this.startPaymentPoll()
        } else {
          uni.showToast({ title: '获取订单详情失败', icon: 'none' })
          setTimeout(() => this.backOrToHome(), 1500)
        }
      } finally {
        uni.hideLoading()
      }
    },
    startPaymentPoll() {
      this.clearPaymentPoll()
      this.paymentPollCount = 0
      const maxAttempts = 15
      uni.showLoading({ title: '订单生成中...' })
      const doPoll = async () => {
        this.paymentPollCount++
        try {
          const res = await getOrderDetail(this.token, this.orderId, { silent: true })
          if (res && res.code === 200 && res.data) {
            this.clearPaymentPoll()
            this.orderDetail = res.data
            this.paymentDeadlineAt = res.data.payment_deadline_at || null
            this.startPaymentCountdown()
            this.initMap()
            if (this.orderDetail?.order?.status === 'delivering' && this.orderDetail?.delivery_employee?.employee_code) {
              this.loadDeliveryEmployeeLocation()
            }
            uni.hideLoading()
            uni.showToast({ title: '订单已生成', icon: 'success' })
            this.fromPayment = false
            return
          }
        } catch (e) {
          console.log('轮询获取订单失败:', e)
        }
        if (this.paymentPollCount >= maxAttempts) {
          this.clearPaymentPoll()
          uni.hideLoading()
          uni.showToast({ title: '订单生成较慢，请稍后从订单列表查看', icon: 'none', duration: 3000 })
          this.fromPayment = false
          setTimeout(() => this.backOrToHome(), 2000)
          return
        }
        this.paymentPollTimer = setTimeout(doPoll, 2000)
      }
      this.paymentPollTimer = setTimeout(doPoll, 2000)
    },
    clearPaymentPoll() {
      if (this.paymentPollTimer) {
        clearTimeout(this.paymentPollTimer)
        this.paymentPollTimer = null
      }
    },
    initMap() {
      if (!this.orderDetail?.address) return
      
      const address = this.orderDetail.address
      if (address.latitude && address.longitude) {
        // 设置地图中心点为收货地址
        this.mapCenter = {
          latitude: address.latitude,
          longitude: address.longitude
        }
        
        // 添加收货地址标记（绿色原点）
        this.mapMarkers = [{
          id: 1,
          latitude: address.latitude,
          longitude: address.longitude,
          title: '收货地址',
          iconPath: '/static/icon/marker-customer-green.png', // 客户绿色原点图标
          width: 24,
          height: 24,
          callout: {
            content: address.name || '收货地址',
            color: '#333',
            fontSize: 12,
            borderRadius: 4,
            bgColor: '#fff',
            padding: 8,
            display: 'ALWAYS'
          }
        }]
      }
    },
    async loadDeliveryEmployeeLocation() {
      const employeeCode = this.orderDetail?.delivery_employee?.employee_code
      if (!employeeCode) {
        console.log('配送员员工码不存在，无法获取位置')
        return
      }
      
      try {
        const res = await getDeliveryEmployeeLocation(this.token, employeeCode)
        if (res && res.code === 200 && res.data) {
          const location = res.data
          if (location.latitude && location.longitude) {
            // 添加配送员位置标记（绿色车辆）
            const deliveryMarker = {
              id: 2,
              latitude: location.latitude,
              longitude: location.longitude,
              title: '配送员位置',
              iconPath: '/static/icon/marker-delivery-car-green.png', // 配送员绿色车辆图标
              width: 30,
              height: 30,
              callout: {
                content: `配送员${location.is_realtime ? '（实时）' : '（历史位置）'}`,
                color: '#fff',
                fontSize: 12,
                borderRadius: 4,
                bgColor: location.is_realtime ? '#20cb6b' : '#20cb6b',
                padding: 8,
                display: 'ALWAYS'
              }
            }
            
            // 更新或添加配送员位置标记
            const existingIndex = this.mapMarkers.findIndex(m => m.id === 2)
            if (existingIndex >= 0) {
              // 如果已存在，更新位置
              this.mapMarkers[existingIndex] = deliveryMarker
              // 触发视图更新
              this.$forceUpdate()
            } else {
              // 如果不存在，添加新标记
              this.mapMarkers.push(deliveryMarker)
            }
            
            // 调整地图视野，同时显示收货地址和配送员位置
            if (this.orderDetail?.address?.latitude && this.orderDetail?.address?.longitude) {
              const lat1 = this.orderDetail.address.latitude
              const lng1 = this.orderDetail.address.longitude
              const lat2 = location.latitude
              const lng2 = location.longitude
              
              // 计算中心点
              const centerLat = (lat1 + lat2) / 2
              const centerLng = (lng1 + lng2) / 2
              
              // 计算距离，调整缩放级别
              const distance = this.calculateDistance(lat1, lng1, lat2, lng2)
              let scale = 14
              if (distance > 10000) scale = 10
              else if (distance > 5000) scale = 11
              else if (distance > 2000) scale = 12
              else if (distance > 1000) scale = 13
              else scale = 14
              
              this.mapCenter = {
                latitude: centerLat,
                longitude: centerLng
              }
              
              // 更新地图缩放级别
              this.mapScale = scale
            }
            
            console.log('配送员位置已加载:', location)
          }
        } else {
          console.log('获取配送员位置失败:', res?.message)
        }
      } catch (error) {
        console.error('获取配送员位置失败:', error)
      }
    },
    // 计算两点间距离（米）
    calculateDistance(lat1, lng1, lat2, lng2) {
      const R = 6371000 // 地球半径（米）
      const dLat = (lat2 - lat1) * Math.PI / 180
      const dLng = (lng2 - lng1) * Math.PI / 180
      const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
                Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
                Math.sin(dLng / 2) * Math.sin(dLng / 2)
      const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
      return R * c
    },
    contactDelivery() {
      if (!this.orderDetail?.delivery_employee?.phone) {
        uni.showToast({
          title: '配送员联系方式不可用',
          icon: 'none'
        })
        return
      }
      
      uni.makePhoneCall({
        phoneNumber: this.orderDetail.delivery_employee.phone,
        fail: (err) => {
          console.error('拨打电话失败:', err)
          uni.showToast({
            title: '拨打电话失败',
            icon: 'none'
          })
        }
      })
    },
    // 打开微信确认收货组件（wx.openBusinessView）
    async handleOpenConfirmReceive() {
      if (this.confirmReceiveLoading || !this.orderId || !this.token) return
      this.confirmReceiveLoading = true
      try {
        const res = await getWechatConfirmReceiveInfo(this.orderId, this.token)
        if (!res || res.code !== 200 || !res.data) {
          uni.showToast({ title: res?.message || '获取失败', icon: 'none' })
          return
        }
        const { transaction_id, merchant_id, merchant_trade_no } = res.data
        const wxObj = typeof wx !== 'undefined' ? wx : uni
        if (!wxObj.openBusinessView) {
          uni.showToast({ title: '当前环境不支持确认收货', icon: 'none' })
          return
        }
        wxObj.openBusinessView({
          businessType: 'weappOrderConfirm',
          extraData: {
            transaction_id,
            merchant_id,
            merchant_trade_no
          },
          success: () => {
            // 组件关闭后会触发 App.onShow，由 App 处理回调并刷新
          },
          fail: (err) => {
            console.error('打开确认收货组件失败:', err)
            uni.showToast({ title: err.errMsg || '打开失败', icon: 'none' })
          }
        })
      } catch (e) {
        uni.showToast({ title: e?.message || '操作失败', icon: 'none' })
      } finally {
        this.confirmReceiveLoading = false
      }
    },
    async handlePayOrder() {
      if (this.paying || !this.orderId || !this.token) return
      this.paying = true
      try {
        const res = await getWechatPayPrepay(this.orderId, this.token)
        if (!res || res.code !== 200 || !res.data) {
          uni.showToast({ title: res?.message || '获取支付参数失败', icon: 'none' })
          return
        }
        const { timeStamp, nonceStr, package: packageVal, signType, paySign } = res.data
        await new Promise((resolve, reject) => {
          uni.requestPayment({
            provider: 'wxpay',
            timeStamp: String(timeStamp),
            nonceStr,
            package: packageVal,
            signType: signType || 'RSA',
            paySign,
            success: () => resolve(),
            fail: (err) => {
              if (err.errMsg && err.errMsg.includes('cancel')) {
                uni.showToast({ title: '已取消支付', icon: 'none' })
              } else {
                uni.showToast({ title: err.errMsg || '支付失败', icon: 'none' })
              }
              reject(err)
            }
          })
        })
        uni.showToast({ title: '支付成功', icon: 'success' })
        this.loadOrderDetail()
      } catch (e) {
        console.error('支付失败:', e)
      } finally {
        this.paying = false
      }
    },
    formatStatus(status) {
      const statusMap = {
        'pending': '订单正在中心分拣中...',
        'pending_payment': '待支付',
        'pending_delivery': '订单正在中心分拣中...',
        'pending_pickup': '分拣已完成，待配送',
        'delivering': '正在配送中...',
        'delivered': '订单已送达',
        'shipped': '订单已送达',
        'paid': '订单已完成',
        'completed': '订单已完成',
        'cancelled': '订单已取消'
      }
      return statusMap[status] || status
    },
    formatStatusShort(status) {
      const statusMap = {
        'pending': '分拣中',
        'pending_payment': '待支付',
        'pending_delivery': '分拣中',
        'pending_pickup': '待配送',
        'delivering': '配送中',
        'delivered': '已送达',
        'shipped': '已送达',
        'paid': '已收款',
        'completed': '已收款',
        'cancelled': '已取消'
      }
      return statusMap[status] || status
    },
    getStatusIcon(status) {
      const iconMap = {
        'pending': 'shop',
        'pending_payment': 'wallet',
        'pending_delivery': 'shop',
        'delivering': 'car',
        'delivered': 'checkmarkempty',
        'shipped': 'checkmarkempty',
        'paid': 'wallet',
        'completed': 'wallet',
        'cancelled': 'close'
      }
      return iconMap[status] || 'shop'
    },
    getStatusColor(status) {
      const colorMap = {
        'pending': '#ff4d4f',
        'pending_payment': '#fa8c16',
        'pending_delivery': '#ff4d4f',
        'delivering': '#1890ff',
        'delivered': '#fa8c16',
        'shipped': '#fa8c16',
        'paid': '#52c41a',
        'completed': '#52c41a',
        'cancelled': '#999'
      }
      return colorMap[status] || '#666'
    },
    getStatusClass(status) {
      const classMap = {
        'pending': 'status-green',
        'pending_delivery': 'status-green',
        'pending_pickup': 'status-yellow',
        'delivering': 'status-green',
        'delivered': 'status-green',
        'shipped': 'status-green',
        'paid': 'status-green',
        'completed': 'status-green',
        'cancelled': 'status-cancelled'
      }
      return classMap[status] || ''
    },
    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hour = String(date.getHours()).padStart(2, '0')
      const minute = String(date.getMinutes()).padStart(2, '0')
      return `${year}-${month}-${day} ${hour}:${minute}`
    },
    formatMoney(amount) {
      if (amount === null || amount === undefined) return '0.00'
      return Number(amount).toFixed(2)
    },
    getOutOfStockStrategyText(strategy) {
      const strategyMap = {
        'cancel_item': '缺货商品不要，其他正常发货',
        'ship_available': '有货的先发，缺货的后续补发',
        'contact_me': '缺货时联系我'
      }
      return strategyMap[strategy] || '缺货时联系我'
    },
    // 取消订单
    async handleCancelOrder() {
      const order = this.orderDetail?.order || {}
      const salesPhone = this.orderDetail?.sales_employee?.phone || ''
      const totalAmount = Number(order.total_amount || 0)
      const isPaid = !!order.paid_at

      // 构建提示内容
      let content = ``
      if (totalAmount > 0) {
        if (isPaid) {
          content += `订单已支付，取消后将原路退款 ¥${this.formatMoney(totalAmount)}，预计1-3工作日到账，是否仍要取消？`
        }
      }
      
      // 显示确认对话框
      const confirmed = await new Promise((resolve) => {
        uni.showModal({
          title: '确认取消订单',
          content: content,
          confirmText: '仍要取消',
          cancelText: '我再想想',
          confirmColor: '#ff4d4f',
          success: (res) => {
            resolve(res.confirm)
          },
          fail: () => {
            resolve(false)
          }
        })
      })
      
      if (!confirmed) {
        return
      }
      
      try {
        uni.showLoading({ title: '取消中...' })
        const orderIdForCancel = (this.orderDetail?.order?.id != null) ? this.orderDetail.order.id : this.orderId
        const res = await cancelOrder(this.token, orderIdForCancel)
        
        if (res && res.code === 200) {
          uni.showToast({
            title: '订单已取消',
            icon: 'success',
            duration: 2000
          })
          
          // 刷新订单详情，更新订单状态
          await this.loadOrderDetail()
          
          // 延迟返回，让用户看到成功提示和更新后的状态
          setTimeout(() => {
            uni.navigateBack()
          }, 1500)
        } else {
          uni.showToast({
            title: res?.message || '取消订单失败',
            icon: 'none',
            duration: 2000
          })
        }
      } catch (error) {
        console.error('取消订单失败:', error)
        uni.showToast({
          title: '取消订单失败，请重试',
          icon: 'none',
          duration: 2000
        })
      } finally {
        uni.hideLoading()
      }
    },
    // 修改订单
    handleModifyOrder() {
      const salesPhone = this.orderDetail?.sales_employee?.phone || ''
      if (salesPhone) {
        uni.makePhoneCall({
          phoneNumber: salesPhone,
          fail: (err) => {
            console.error('拨打电话失败:', err)
            uni.showToast({
              title: '拨打电话失败',
              icon: 'none'
            })
          }
        })
      } else {
        uni.showToast({
          title: '暂无销售员联系方式',
          icon: 'none'
        })
      }
    },
    // 查看更多
    handleViewMore() {
      // 滚动到商品列表
      uni.pageScrollTo({
        scrollTop: 0,
        duration: 300
      })
    },
    // 跳转到客服中心
    goToCustomerService() {
      uni.navigateTo({
        url: '/pages/customer-service/customer-service'
      })
    },
    // 刷新配送员位置
    async refreshDeliveryLocation() {
      if (!this.orderDetail?.delivery_employee?.employee_code) {
        uni.showToast({
          title: '配送员信息不可用',
          icon: 'none'
        })
        return
      }
      
      uni.showLoading({ title: '刷新中...' })
      await this.loadDeliveryEmployeeLocation()
      uni.hideLoading()
      uni.showToast({
        title: '位置已更新',
        icon: 'success',
        duration: 1500
      })
    }
  }
}
</script>

<style scoped>
.order-detail-page {
  width: 100%;
  min-height: 100vh;
  background: #f5f5f5;
}

.custom-navbar {
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
  padding: 0 20rpx;
  box-sizing: border-box;
}

.navbar-left {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  cursor: pointer;
}

.navbar-title {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.navbar-title-text {
  font-size: 32rpx;
  font-weight: 500;
  color: #fff;
}

.navbar-right {
  width: 60rpx;
  flex-shrink: 0;
}

.detail-content {
  width: 100%;
  box-sizing: border-box;
  padding: 20rpx 24rpx;
  min-height: calc(100vh - var(--nav-height, 0px));
  background: #F5F6FA;
  margin-top: -148rpx;
  position: relative;
  z-index: 10;
  border-radius: 40rpx 40rpx 0 0;
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  font-size: 28rpx;
  color: #999;
}

.status-section {
  background: linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%);
  padding: 60rpx 40rpx 50rpx;
  text-align: center;
  margin: 20rpx 0;
  border-radius: 24rpx;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.06);
  position: relative;
  overflow: hidden;
}

.status-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 6rpx;
  background: linear-gradient(90deg, #20CB6B 0%, #1AB85A 100%);
}

.status-icon {
  margin-bottom: 24rpx;
  display: flex;
  justify-content: center;
  align-items: center;
}

.status-text {
  display: block;
  font-size: 36rpx;
  font-weight: 700;
  color: #20253A;
  margin-bottom: 20rpx;
  letter-spacing: 0.5rpx;
  line-height: 1.6;
  padding: 0 20rpx;
  word-break: break-all;
}

.order-number {
  display: block;
  font-size: 26rpx;
  color: #8C92A4;
  background-color: #f5f5f5;
  padding: 12rpx 24rpx;
  border-radius: 20rpx;
  display: inline-block;
  margin-top: 8rpx;
}

.section {
  background-color: #fff;
  padding: 32rpx 28rpx;
  margin-bottom: 20rpx;
  border-radius: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
  box-sizing: border-box;
}

.section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 32rpx;
  padding: 0;
  letter-spacing: 0.5rpx;
}

.address-section {
  margin-top: 24rpx;
}

.address-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.address-main {
  flex: 1;
  /* padding-right: 24rpx; */
}

.address-header {
  width: 100%;
}

.address-title-row {
  display: flex;
  align-items: center;
  margin-bottom: 12rpx;
}

.address-store {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  line-height: 1.4;
}

.address-contact-row {
  display: flex;
  align-items: center;
  margin-bottom: 8rpx;
}

.address-contact {
  font-size: 28rpx;
  color: #666;
  margin-right: 16rpx;
  font-weight: 500;
}

.address-phone {
  font-size: 28rpx;
  color: #666;
}

.address-detail {
  font-size: 26rpx;
  color: #909399;
  line-height: 1.5;
  margin-top: 4rpx;
}

.goods-list {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.goods-item {
  display: flex;
  align-items: flex-start;
}

.goods-image {
  width: 140rpx;
  height: 140rpx;
  border-radius: 16rpx;
  margin-right: 20rpx;
  background-color: #F5F5F5;
  flex-shrink: 0;
}

.goods-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 140rpx;
  padding-top: 4rpx;
}

.goods-name-row {
  margin-bottom: 8rpx;
}

.goods-name {
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.goods-spec {
  font-size: 24rpx;
  color: #909399;
  margin-bottom: 12rpx;
  line-height: 1.4;
}

.goods-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
}

.goods-price {
  font-size: 30rpx;
  color: #FF4D4F;
  font-weight: 600;
}

.goods-qty {
  font-size: 28rpx;
  color: #666;
  font-weight: 500;
}

.amount-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
  min-height: 44rpx;
}

.amount-label {
  font-size: 28rpx;
  color: #666;
  font-weight: 400;
}

.amount-value {
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
  text-align: right;
}

.amount-value.discount-text {
  color: #20CB6B;
  font-weight: 500;
}

.amount-divider {
  height: 1rpx;
  background-color: #F0F0F0;
  margin: 20rpx 0;
}

.amount-row.total-row {
  margin-top: 8rpx;
  margin-bottom: 0;
  padding-top: 12rpx;
}

.total-label {
  font-size: 32rpx;
  color: #333;
  font-weight: 600;
}

.total-value {
  font-size: 36rpx;
  color: #FF4D4F;
  font-weight: 600;
}

/* 加急订单模块样式 */
.urgent-section {
  margin-top: 0;
}

.urgent-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-radius: 12rpx;
  transition: all 0.3s ease;
}

.urgent-container.urgent-active {
  background-color: #E8F8F0;
  padding: 20rpx;
}

.urgent-left {
  flex: 1;
  display: flex;
  align-items: center;
}

.urgent-header {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.urgent-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #20CB6B;
}

.urgent-tag {
  display: inline-block;
  padding: 4rpx 12rpx;
  color: #20CB6B;
  font-size: 24rpx;
  font-weight: 500;
  border-radius: 12rpx;
  background-color: #fff;
}

.urgent-right {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.urgent-price-wrapper {
  display: flex;
  align-items: baseline;
  padding: 8rpx 16rpx;
  border-radius: 8rpx;
}

.urgent-price {
  font-size: 32rpx;
  font-weight: 700;
  color: #20CB6B;
  line-height: 1;
}

/* 加急费用突出显示 */
.urgent-fee-row {
  border-radius: 12rpx;
  margin: 16rpx 0;
}

.urgent-fee-label-wrapper {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.urgent-fee-tag {
  display: inline-block;
  padding: 4rpx 12rpx;
  background-color: #E8F8F0;
  color: #20CB6B;
  font-size: 20rpx;
  border-radius: 8rpx;
  font-weight: 500;
}

.urgent-fee-value {
  color: #20CB6B;
  font-size: 30rpx;
}

/* 备注部分 */
.remark-section {
  margin-top: 0;
}

.remark-header {
  margin-bottom: 20rpx;
}

.remark-content {
  padding: 20rpx;
  background-color: #F5F6FA;
  border-radius: 12rpx;
  font-size: 28rpx;
  color: #333;
  min-height: 80rpx;
  line-height: 1.6;
}

.remark-text {
  display: block;
  word-break: break-all;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 20rpx 0;
  min-height: 44rpx;
}

.info-row:not(:last-child) {
  border-bottom: 1rpx solid #F0F0F0;
}

.info-label {
  color: #666;
  min-width: 160rpx;
  font-weight: 400;
  font-size: 28rpx;
}

.info-value {
  flex: 1;
  text-align: right;
  color: #333;
  font-weight: 500;
  word-break: break-all;
  font-size: 28rpx;
}

/* 其他选项部分 */
.options-section {
  margin-top: 0;
}

.options-section .option-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx 0;
  min-height: 80rpx;
}

.options-section .option-row:not(:last-child) {
  border-bottom: 1rpx solid #F0F0F0;
}

.option-text {
  flex: 1;
  padding-right: 24rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.option-title {
  font-size: 30rpx;
  color: #333;
  font-weight: 500;
  line-height: 1.4;
}

.option-desc {
  font-size: 24rpx;
  color: #909399;
  line-height: 1.5;
}

.option-status {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex-shrink: 0;
}

.option-status-text {
  font-size: 26rpx;
  color: #909399;
}

.option-status-active {
  font-size: 26rpx;
  color: #20CB6B;
}

.option-status-value {
  font-size: 26rpx;
  color: #333;
  text-align: right;
  max-width: 300rpx;
  word-break: break-all;
}

.status-yellow {
  color: #faad14;
}

.status-green {
  color: #20CB6B;
}

.status-cancelled {
  color: #999;
}

.sales-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx;
  background: #f9f9f9;
  border-radius: 8rpx;
}

.sales-info {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  flex: 1;
}

.sales-name {
  font-size: 32rpx;
  font-weight: 700;
  color: #20253A;
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.sales-name::before {
  content: '👤';
  font-size: 28rpx;
}

.sales-code {
  font-size: 26rpx;
  color: #8C92A4;
  background-color: #fff;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
  display: inline-block;
  width: fit-content;
}

.contact-btn {
  display: flex;
  align-items: center;
  gap: 10rpx;
  padding: 20rpx 36rpx;
  background: #20CB6B;
  border-radius: 8rpx;
  font-size: 28rpx;
  color: #fff;
  font-weight: 500;
}

.contact-btn:active {
  background-color: #1AB85A;
}

.contact-buttons-section {
  padding: 32rpx 0;
}

.contact-buttons {
  display: flex;
  gap: 20rpx;
  justify-content: space-between;
  padding: 0;
}

.contact-btn-small {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  padding: 16rpx 32rpx;
  background: #20CB6B;
  border: 1rpx solid #20CB6B;
  border-radius: 12rpx;
  font-size: 26rpx;
  color: #fff;
  font-weight: 400;
  flex: 1;
  transition: all 0.2s;
}

.contact-btn-small:active {
  background-color: #1AB85A;
  border-color: #1AB85A;
}

.top-gradient-section {
  width: 100%;
  position: relative;
}

.map-section {
  width: 100%;
  height: 600rpx;
  overflow: hidden;
  position: relative;
}

.map-container {
  width: 100%;
  height: 100%;
  margin: 0;
}

.map-refresh-btn {
  position: absolute;
  right: 20rpx;
  bottom: 166rpx;
  width: 72rpx;
  height: 72rpx;
  background: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.15);
  z-index: 100;
}

.map-refresh-btn:active {
  background: #f5f5f5;
}

.gradient-background {
  width: 100%;
  background: linear-gradient(180deg, #20CB6B 0%, #1AB85A 30%, rgba(26, 184, 90, 0.6) 70%, rgba(245, 245, 245, 1) 100%);
  padding: 40rpx 30rpx 120rpx;
  box-sizing: border-box;
  position: relative;
  z-index: 1;
  border-radius: 0 0 40rpx 40rpx;
}

.status-content {
  display: flex;
  align-items: flex-start;
  margin-bottom: 60rpx;
}

.status-left {
  display: flex;
  align-items: center;
  gap: 24rpx;
  flex: 1;
}

.status-icon-circle {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.status-text-group {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  flex: 1;
}

.status-main-text {
  font-size: 32rpx;
  font-weight: 700;
  color: #fff;
  line-height: 1.4;
}

.status-tag {
  display: inline-block;
  padding: 8rpx 20rpx;
  background: rgba(255, 255, 255, 0.25);
  border-radius: 20rpx;
  font-size: 24rpx;
  color: #fff;
  width: fit-content;
}

.payment-countdown {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
  margin-top: 8rpx;
}

.countdown-label {
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.85);
}

.countdown-value {
  font-size: 36rpx;
  font-weight: 700;
  color: #fff;
  letter-spacing: 4rpx;
}

.status-actions {
  display: flex;
  gap: 20rpx;
  justify-content: center;
  margin-top: 40rpx;
}

.action-btn {
  flex: 1;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 44rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #20CB6B;
  font-weight: 500;
}

.action-btn:active {
  background: rgba(255, 255, 255, 0.85);
}

.delivery-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx;
  background: #f9f9f9;
  border-radius: 8rpx;
}

.delivery-info {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  flex: 1;
}

.delivery-name {
  font-size: 32rpx;
  font-weight: 700;
  color: #20253A;
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.delivery-name::before {
  content: '🚚';
  font-size: 28rpx;
}

.delivery-code {
  font-size: 26rpx;
  color: #8C92A4;
  background-color: #fff;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
  display: inline-block;
  width: fit-content;
}

/* 底部操作栏：统一高度与内边距，各状态视觉一致 */
.action-footer {
  width: 100%;
  background-color: #fff;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  border-top: 1rpx solid #eee;
  padding: 24rpx 0;
  padding-bottom: calc(24rpx + env(safe-area-inset-bottom));
  z-index: 999;
  box-sizing: border-box;
  min-height: 120rpx;
}

.action-footer-container {
  width: 100%;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 30rpx;
  box-sizing: border-box;
}

.action-footer-left {
  flex-shrink: 0;
  width: 180rpx;
  min-width: 180rpx;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 14rpx;
}

.action-icon-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 72rpx;
  padding: 8rpx 0;
}

.action-icon-text {
  font-size: 22rpx;
  color: #2C2C2C;
  margin-top: 6rpx;
}

.action-footer-right {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

/* 主按钮统一尺寸与圆角，略宽以减少中间空隙感 */
.action-main-btn {
  flex-shrink: 0;
  min-width: 500rpx;
  height: 80rpx;
  line-height: 80rpx;
  background-color: #20CB6B;
  color: #fff;
  font-size: 30rpx;
  font-weight: 600;
  padding: 0 56rpx;
  border-radius: 40rpx;
  display: flex;
  justify-content: center;
  align-items: center;
  box-sizing: border-box;
}

.action-main-btn text {
  white-space: nowrap;
}

.action-main-btn:active {
  background-color: #1AB85A;
}


.customer-service-tip {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  padding: 24rpx 30rpx;
  margin-bottom: 20rpx;
  /* margin: 40rpx 0 20rpx; */
  /* background: #f5f5f5; */
  /* border-radius: 16rpx; */
  /* border: 1rpx solid #e8e8e8; */
  cursor: pointer;
}

.order-notice {
  margin: 32rpx 0;
  padding: 24rpx 32rpx;
  background: #f0e8ff;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.notice-text {
  font-size: 26rpx;
  color: #6b46c1;
  line-height: 1.5;
  text-align: center;
}

.service-avatar {
  width: 48rpx;
  height: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.service-text {
  font-size: 28rpx;
  color: #20CB6B;
}
</style>




