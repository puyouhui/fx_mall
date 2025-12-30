<template>
  <view>
    <view class="ps-modal" v-if="isVisible" @click="closeModal" @touchmove.stop.prevent>
      <view class="ps-overlay" @touchmove.stop.prevent></view>
      <view class="ps-content" @click.stop>
        <view class="ps-header">
          <text class="ps-title">选择规格</text>
          <view class="ps-close" @click.stop="closeModal">
            <uni-icons type="close" size="24" color="#999"></uni-icons>
          </view>
        </view>

        <view class="ps-info" v-if="selectedProduct">
          <image :src="selectedProduct.images?.[0] || ''" class="ps-image" mode="aspectFill"></image>
          <view class="ps-details">
            <text class="ps-name">{{ selectedProduct?.name || '暂无名称' }}</text>
            <view class="ps-meta">
              <text class="ps-desc">{{ selectedProduct?.description || '适合家用，柔软舒适' }}</text>
              <text class="ps-meta-sep">·</text>
              <text class="ps-spec-count">
                {{ selectedProduct?.specs?.length || 0 }} 个可选规格
              </text>
            </view>
            <text class="ps-price-range">¥{{ selectedProduct?.displayPrice || '0.00' }}</text>
          </view>
        </view>

      <view class="ps-specs" v-if="selectedProduct?.specs?.length">
        <text class="ps-section-title">选择规格</text>
        <scroll-view class="ps-specs-list" scroll-y>
          <view class="ps-specs-container">
            <view
              class="ps-spec"
              v-for="(spec, index) in selectedProduct.specs"
              :key="index"
              :class="{ active: getSpecQuantity(spec) > 0 }"
            >
              <view class="ps-spec-info">
                <view class="ps-spec-header">
                  <text class="ps-spec-name">{{ spec.name }}</text>
                  <text class="ps-spec-desc" v-if="spec.description">{{ spec.description }}</text>
                </view>
                <view class="ps-spec-price-container" :class="{ 'wholesale-layout': isWholesaleUser }">
                  <text v-if="isWholesaleUser" class="ps-spec-price">
                    ¥{{ formatSpecPrice(spec, 'wholesale') }}
                  </text>
                  <text v-else class="ps-spec-price">
                    ¥{{ formatSpecPrice(spec, 'retail') }}
                  </text>
                  <!-- 批发用户显示零售价（灰色） -->
                  <text v-if="isWholesaleUser" class="ps-spec-retail-price">
                    零售价：¥{{ formatSpecPrice(spec, 'retail') }}
                  </text>
                </view>
              </view>
              <view class="ps-spec-actions">
                <view
                  class="ps-spec-add"
                  v-if="getSpecQuantity(spec) === 0"
                  @click.stop="increaseSpecQuantity(spec)"
                >
                  <uni-icons type="plusempty" size="22" color="#fff"></uni-icons>
                </view>
                <view v-else class="ps-spec-qty">
                  <view class="ps-spec-btn" @click.stop="decreaseSpecQuantity(spec)">
                    <image src="/static/icon/minus.png" class="ps-minus-icon" />
                  </view>
                  <text class="ps-spec-qty-text">{{ getSpecQuantity(spec) }}</text>
                  <view class="ps-spec-btn plus" @click.stop="increaseSpecQuantity(spec)">
                    <uni-icons type="plusempty" size="18" color="#fff"></uni-icons>
                  </view>
                </view>
              </view>
            </view>
          </view>
        </scroll-view>
      </view>
      <view class="ps-empty-specs" v-else>
        暂无规格信息
      </view>

      <view v-if="!selectedProduct?.specs?.length" class="ps-quantity-single">
        <text class="ps-section-title">数量</text>
        <view class="ps-qty-selector">
          <view class="ps-minus" :class="{ disabled: singleQuantity <= 1 }" @click.stop="decreaseSingleQuantity">
              <image src="/static/icon/minus.png" class="ps-minus-icon"></image>
            </view>
          <text class="ps-qty-text">{{ singleQuantity }}</text>
          <view class="ps-plus" @click.stop="increaseSingleQuantity">
              <uni-icons type="plusempty" size="18" color="#fff"></uni-icons>
            </view>
          </view>
        </view>

        <view class="ps-footer">
          <view class="ps-confirm" @click.stop="addToCart">
            <text>{{ loading ? '处理中...' : '加入采购单' }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { getProductDetail } from '../api/products'
import { miniLogin } from '../api/index'
import { addItemToPurchaseList } from '../utils/purchaseList'

const PROFILE_FORM_PAGE = '/pages/profile/form'

const isVisible = ref(false)
const loading = ref(false)
const selectedProduct = ref(null)
const quantityMap = ref({})
const singleQuantity = ref(1)
const identityNavigating = ref(false)

const userState = ref({
  info: null,
  token: '',
  uniqueId: ''
})

const refreshUserStateFromStorage = () => {
  const storedInfo = uni.getStorageSync('miniUserInfo')
  const storedToken = uni.getStorageSync('miniUserToken')
  const storedUniqueId = uni.getStorageSync('miniUserUniqueId')

  userState.value = {
    info: storedInfo || null,
    token: storedToken || '',
    uniqueId:
      storedUniqueId ||
      storedInfo?.unique_id ||
      storedInfo?.uniqueId ||
      ''
  }
  return userState.value
}

refreshUserStateFromStorage()

const persistUserState = (info, token = '') => {
  if (info) {
    userState.value.info = info
    userState.value.uniqueId = info.unique_id || info.uniqueId || ''
    uni.setStorageSync('miniUserInfo', info)
    if (userState.value.uniqueId) {
      uni.setStorageSync('miniUserUniqueId', userState.value.uniqueId)
    }
  }

  userState.value.token = token || ''
  if (token) {
    uni.setStorageSync('miniUserToken', token)
  }
}

const performMiniLogin = async () => {
  uni.showLoading({
    title: '登录中...',
    mask: true
  })
  try {
    const loginRes = await new Promise((resolve, reject) => {
      uni.login({
        provider: 'weixin',
        success: resolve,
        fail: reject
      })
    })

    if (!loginRes || !loginRes.code) {
      throw new Error('未获取到登录凭证')
    }

    // 获取本地存储的分享者ID
    const shareReferrerId = uni.getStorageSync('shareReferrerId')
    let referrerId = null
    if (shareReferrerId) {
      const id = parseInt(shareReferrerId)
      if (!isNaN(id) && id > 0) {
        referrerId = id
      }
    }

    const resp = await miniLogin(loginRes.code, referrerId)
    const data = resp?.data || {}
    const user = data.user || {}
    const token = data.token || ''
    const uniqueId = user.unique_id || user.uniqueId

    if (!uniqueId) {
      throw new Error('未返回用户唯一ID')
    }

    // 登录成功后，清除分享者ID（只绑定一次）
    if (referrerId) {
      uni.removeStorageSync('shareReferrerId')
    }

    persistUserState(user, token)
    return userState.value
  } finally {
    uni.hideLoading()
  }
}

const ensureMiniUserInfo = async () => {
  const current = refreshUserStateFromStorage()
  if (current.info && current.uniqueId) {
    return current
  }

  return await performMiniLogin()
}

const navigateToProfilePage = () => {
  if (identityNavigating.value) {
    return
  }

  identityNavigating.value = true
  uni.navigateTo({
    url: PROFILE_FORM_PAGE,
    complete: () => {
      setTimeout(() => {
        identityNavigating.value = false
      }, 300)
    }
  })
}

const ensureUserReady = async () => {
  try {
    const state = await ensureMiniUserInfo()
    const info = state.info
    if (!info) {
      return false
    }

    const profileCompleted = info.profile_completed || info.profileCompleted || false
    if (!profileCompleted) {
      uni.showToast({
        title: '请先完善资料',
        icon: 'none'
      })
      setTimeout(() => {
        navigateToProfilePage()
      }, 500)
      return false
    }

    return true
  } catch (error) {
    console.error('登录校验失败:', error)
    uni.showToast({
      title: error?.message || '登录失败，请稍后重试',
      icon: 'none'
    })
    return false
  }
}

const open = async (product) => {
  if (!product || !product.id) {
    uni.showToast({ title: '商品信息有误', icon: 'none' })
    return
  }

  // 刷新用户状态，确保获取最新信息（特别是从资料填写页面返回后）
  refreshUserStateFromStorage()

  loading.value = true
  uni.showLoading({
    title: '加载中',
    mask: true
  })

  try {
    const res = await getProductDetail(parseInt(product.id))
    if (res.code === 200 && res.data) {
      const detail = normalizeProductDetail(res.data)
      selectedProduct.value = detail
      if (detail.specs && detail.specs.length > 0) {
        quantityMap.value = {}
        detail.specs.forEach(spec => {
          quantityMap.value[spec.id] = 0
        })
      } else {
        singleQuantity.value = 1
      }
      isVisible.value = true
    } else {
      uni.showToast({ title: '未找到商品', icon: 'none' })
    }
  } catch (error) {
    console.error('加载商品详情失败:', error)
    uni.showToast({ title: '加载失败，请稍后再试', icon: 'none' })
  } finally {
    loading.value = false
    uni.hideLoading()
  }
}

const normalizeProductDetail = (product) => {
  const detail = { ...product }
  if (!Array.isArray(detail.images) && typeof detail.images === 'string') {
    try {
      detail.images = JSON.parse(detail.images)
    } catch (error) {
      detail.images = [detail.images]
    }
  }
  if (!Array.isArray(detail.images)) {
    detail.images = []
  }

  if (!detail.specs && detail.specifications && Array.isArray(detail.specifications)) {
    detail.specs = detail.specifications
  }

  if (!Array.isArray(detail.specs)) {
    detail.specs = []
  }

  detail.specs = detail.specs.map((spec, index) => {
    const normalized = { ...spec }
    if (normalized.id === undefined) {
      normalized.id = index + 1
    }
    if (normalized.wholesale_price !== undefined) {
      normalized.wholesale_price = parseFloat(normalized.wholesale_price) || 0
    }
    if (normalized.wholesalePrice !== undefined) {
      normalized.wholesalePrice = parseFloat(normalized.wholesalePrice) || 0
    }
    if (normalized.retail_price !== undefined) {
      normalized.retail_price = parseFloat(normalized.retail_price) || 0
    }
    if (normalized.retailPrice !== undefined) {
      normalized.retailPrice = parseFloat(normalized.retailPrice) || 0
    }
    if (normalized.price !== undefined) {
      normalized.price = parseFloat(normalized.price) || 0
    }
    return normalized
  })

  detail.displayPrice = calculateProductPriceRange(detail)
  return detail
}

const calculateProductPriceRange = (product) => {
  if (!product.specs || !Array.isArray(product.specs) || product.specs.length === 0) {
    return (product.price || 0).toFixed(2)
  }

  // 根据用户类型决定显示哪种价格
  const info = userState.value.info
  const isWholesale = info && info.user_type === 'wholesale'

  const prices = []
  product.specs.forEach(spec => {
    if (isWholesale) {
      // 批发用户：显示批发价
      const wholesalePrice = spec.wholesale_price || spec.wholesalePrice
      if (wholesalePrice && wholesalePrice > 0) {
        prices.push(parseFloat(wholesalePrice))
      }
    } else {
      // 未登录或零售用户：显示零售价
      const retailPrice = spec.retail_price || spec.retailPrice
      if (retailPrice && retailPrice > 0) {
        prices.push(parseFloat(retailPrice))
      }
    }
  })

  // 如果没有找到对应类型的价格，使用另一种价格作为后备
  if (prices.length === 0) {
    product.specs.forEach(spec => {
      if (isWholesale) {
        // 批发用户找不到批发价，使用零售价作为后备
        const retailPrice = spec.retail_price || spec.retailPrice
        if (retailPrice && retailPrice > 0) {
          prices.push(parseFloat(retailPrice))
        }
      } else {
        // 零售用户找不到零售价，使用批发价作为后备
        const wholesalePrice = spec.wholesale_price || spec.wholesalePrice
        if (wholesalePrice && wholesalePrice > 0) {
          prices.push(parseFloat(wholesalePrice))
        }
      }
      // 最后使用通用价格字段
      if (spec.price && spec.price > 0 && prices.length === 0) {
        prices.push(parseFloat(spec.price))
      }
    })
  }

  if (prices.length === 0) {
    return (product.price || 0).toFixed(2)
  }

  // 显示价格范围（最低价~最高价）
  const minPrice = Math.min(...prices)
  const maxPrice = Math.max(...prices)
  
  if (minPrice === maxPrice) {
    // 如果所有规格价格相同，只显示一个价格
    return minPrice.toFixed(2)
  } else {
    // 显示价格范围
    return `${minPrice.toFixed(2)}~${maxPrice.toFixed(2)}`
  }
}

const getSpecQuantity = (spec) => {
  if (!spec || !spec.id) return 0
  return quantityMap.value[spec.id] || 0
}

// 计算是否为批发用户
const isWholesaleUser = computed(() => {
  const info = userState.value.info
  return info && info.user_type === 'wholesale'
})

const formatSpecPrice = (spec, priceType = 'retail') => {
  let price = 0
  if (priceType === 'wholesale') {
    // 批发价
    price = spec?.wholesale_price ?? spec?.wholesalePrice ?? 0
  } else {
    // 零售价
    price = spec?.retail_price ?? spec?.retailPrice ?? 0
  }
  
  // 如果指定类型的价格不存在，使用通用价格字段作为后备
  if (!price || price === 0) {
    price = spec?.price ?? selectedProduct.value?.price ?? 0
  }
  
  return parseFloat(price || 0).toFixed(2)
}

const vibrate = () => {
  uni.vibrateShort({ type: 'light' })
}

const increaseSpecQuantity = async (spec) => {
  if (!spec || !spec.id) return
  const ready = await ensureUserReady()
  if (!ready) return
  vibrate()
  quantityMap.value = {
    ...quantityMap.value,
    [spec.id]: getSpecQuantity(spec) + 1
  }
}

const decreaseSpecQuantity = (spec) => {
  if (!spec || !spec.id) return
  const current = getSpecQuantity(spec)
  if (current <= 0) return
  vibrate()
  quantityMap.value = {
    ...quantityMap.value,
    [spec.id]: current - 1
  }
}

const increaseSingleQuantity = async () => {
  const ready = await ensureUserReady()
  if (!ready) return
  vibrate()
  singleQuantity.value++
}

const decreaseSingleQuantity = () => {
  if (singleQuantity.value > 1) {
    vibrate()
    singleQuantity.value--
  }
}

const addToCart = async () => {
  const ready = await ensureUserReady()
  if (!ready) return
  if (!selectedProduct.value) {
    uni.showToast({ title: '商品信息缺失', icon: 'none' })
    return
  }

  const token = userState.value.token || uni.getStorageSync('miniUserToken')
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    return
  }

  try {
    loading.value = true
  if (selectedProduct.value.specs && selectedProduct.value.specs.length > 0) {
    const selectedSpecs = selectedProduct.value.specs.filter(spec => getSpecQuantity(spec) > 0)
    if (!selectedSpecs.length) {
      uni.showToast({ title: '请选择至少一个规格数量', icon: 'none' })
      return
    }
      for (const spec of selectedSpecs) {
        await addItemToPurchaseList({
          token,
        productId: selectedProduct.value.id,
          specName: spec.name || '默认规格',
        quantity: getSpecQuantity(spec)
        })
      }
  } else {
    if (singleQuantity.value <= 0) {
      uni.showToast({ title: '请选择数量', icon: 'none' })
      return
    }
      await addItemToPurchaseList({
        token,
      productId: selectedProduct.value.id,
        specName: '默认规格',
      quantity: singleQuantity.value
      })
    }
  uni.vibrateShort({ type: 'medium' })
  uni.showToast({ title: '已添加到采购单', icon: 'success' })
  closeModal()
  } catch (error) {
    console.error('添加采购单失败:', error)
    uni.showToast({ title: '添加失败，请稍后再试', icon: 'none' })
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  isVisible.value = false
}

defineExpose({
  open,
  close: closeModal
})
</script>

<style scoped>
.ps-modal {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  justify-content: flex-end;
  align-items: flex-end;
}

.ps-overlay {
  position: absolute;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
}

.ps-content {
  width: 100%;
  max-height: 85vh;
  background-color: #fff;
  border-radius: 30rpx 30rpx 0 0;
  position: relative;
  z-index: 1;
  padding: 30rpx;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.ps-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.ps-title {
  font-size: 32rpx;
  font-weight: bold;
}

.ps-close {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ps-info {
  display: flex;
  margin: 24rpx 0;
  flex-shrink: 0;
}

.ps-image {
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  margin-right: 20rpx;
}

.ps-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #222;
}

.ps-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin: 8rpx 0;
  font-size: 24rpx;
  color: #888;
  align-items: center;
}

.ps-meta-sep {
  color: #dcdcdc;
}

.ps-spec-count {
  color: #20CB6B;
}

.ps-price-range {
  font-size: 30rpx;
  font-weight: 600;
  color: #ff4d4f;
}
.ps-specs{
  padding: 0;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  margin-bottom: 20rpx;
}
.ps-quantity {
  padding: 24rpx 0;
}

.ps-section-title {
  font-size: 28rpx;
  font-weight: 600;
  margin-bottom: 16rpx;
  color: #333;
  display: block;
  flex-shrink: 0;
}

.ps-specs-list {
  flex: 1;
  min-height: 0;
  width: 100%;
  box-sizing: border-box;
  /* 增加最大高度，可以显示更多规格 */
  max-height: 550rpx;
}

.ps-specs-container {
  display: flex;
  flex-direction: column;
  padding: 4rpx 0;
}

.ps-spec {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-radius: 16rpx;
  border: 1rpx solid #f0f0f0;
  padding: 16rpx 20rpx;
  background: #fff;
  box-shadow: 0 6rpx 12rpx rgba(0, 0, 0, 0.02);
  box-sizing: border-box;
  min-height: 140rpx;
  margin-bottom: 20rpx;
}

.ps-spec:last-child {
  margin-bottom: 0;
}

.ps-spec.active {
  border-color: rgba(32, 203, 107, 0.5);
  background: rgba(32, 203, 107, 0.04);
}

.ps-spec-info {
  flex: 1;
  padding-right: 20rpx;
}

.ps-spec-header {
  display: flex;
  align-items: baseline;
  gap: 8rpx;
  margin-bottom: 8rpx;
}

.ps-spec-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #222;
}

.ps-spec-desc {
  font-size: 24rpx;
  color: #8a8a8a;
}

.ps-spec-price-container {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.ps-spec-price-container.wholesale-layout {
  flex-direction: row;
  align-items: baseline;
  gap: 12rpx;
}

.ps-spec-price {
  color: #f00;
  font-size: 32rpx;
  font-weight: bold;
}

.ps-spec-retail-price {
  color: #999;
  font-size: 24rpx;
  text-decoration: line-through;
}

.ps-empty-specs {
  padding: 40rpx 0;
  text-align: center;
  color: #999;
  font-size: 26rpx;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ps-spec-actions {
  display: flex;
  align-items: center;
  margin-left: 20rpx;
}

.ps-spec-add {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #20CB6B;
}

.ps-spec-qty {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.ps-spec-btn {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  /* border: 1rpx solid rgba(32, 203, 107, 0.3); */
  display: flex;
  align-items: center;
  justify-content: center;
  /* background: rgba(32, 203, 107, 0.08); */
  background-color: #F7F8F9;
}

.ps-spec-btn.plus {
  background: #20CB6B;
  border-color: transparent;
}

.ps-spec-qty-text {
  font-size: 32rpx;
  font-weight: 600;
  min-width: 48rpx;
  text-align: center;
}

.ps-quantity-single {
  margin-top: 10rpx;
  flex-shrink: 0;
}

.ps-qty-selector {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.ps-minus,
.ps-plus {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background-color: #f5f5f5;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1rpx solid transparent;
}

.ps-minus.disabled {
  opacity: 0.4;
}

.ps-plus {
  background-color: #20CB6B;
  border-color: rgba(32, 203, 107, 0.2);
}

.ps-minus-icon {
  width: 30rpx;
  height: 30rpx;
}

.ps-qty-text {
  font-size: 32rpx;
  font-weight: bold;
}

.ps-footer {
  flex-shrink: 0;
}

.ps-confirm {
  width: 100%;
  background: linear-gradient(90deg, #20CB6B, #12a458);
  border-radius: 999rpx;
  text-align: center;
  padding: 28rpx 0;
  color: #fff;
  font-size: 30rpx;
  font-weight: 600;
}
</style>

