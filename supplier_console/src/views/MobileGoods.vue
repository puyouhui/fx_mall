<template>
  <div class="mobile-goods-page">
    <div class="header">
      <h1>待备货</h1>
    </div>
    
    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>
    
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
    </div>
    
    <div v-else-if="goodsList.length === 0" class="empty">
      <p>暂无待备货货物</p>
    </div>
    
    <div v-else class="goods-list">
      <div 
        v-for="(goods, index) in goodsList" 
        :key="index" 
        class="goods-item"
      >
        <div class="goods-image">
          <img 
            v-if="goods.image" 
            :src="goods.image" 
            :alt="goods.product_name"
            @error="handleImageError"
          />
          <div v-else class="no-image">暂无图片</div>
        </div>
        <div class="goods-info">
          <h3 class="goods-name">{{ goods.product_name }}</h3>
          <div class="goods-details">
            <span class="goods-spec">规格：{{ goods.spec_name }}</span>
            <span class="quantity">{{ goods.quantity }} 件</span>
          </div>
        </div>
      </div>
      
      <div class="summary">
        <div class="summary-item">
          <span>总件数：</span>
          <strong>{{ totalQuantity }} 件</strong>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { getMobilePendingGoods } from '../api/mobile'

const route = useRoute()
const loading = ref(false)
const error = ref('')
const goodsList = ref([])

// 从URL参数获取 name 和 ID
const supplierName = ref('')
const supplierId = ref('')

// 计算总件数
const totalQuantity = computed(() => {
  return goodsList.value.reduce((sum, goods) => sum + (goods.quantity || 0), 0)
})

// 图片加载错误处理
const handleImageError = (event) => {
  event.target.style.display = 'none'
  const parent = event.target.parentElement
  if (parent && !parent.querySelector('.no-image')) {
    const noImage = document.createElement('div')
    noImage.className = 'no-image'
    noImage.textContent = '暂无图片'
    parent.appendChild(noImage)
  }
}

// 加载待备货货物列表
const loadGoods = async () => {
  loading.value = true
  error.value = ''
  
  try {
    // 从URL参数获取
    supplierName.value = route.query.name || ''
    supplierId.value = route.query.ID || route.query.id || ''
    
    if (!supplierName.value || !supplierId.value) {
      error.value = '缺少必要参数：name 和 ID'
      loading.value = false
      return
    }
    
    const response = await getMobilePendingGoods(supplierName.value, supplierId.value)
    
    if (response.code === 200 && response.data) {
      goodsList.value = response.data.list || []
    } else {
      error.value = response.message || '获取数据失败'
    }
  } catch (err) {
    console.error('加载货物列表失败:', err)
    error.value = err.message || '加载失败，请稍后再试'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadGoods()
})
</script>

<style scoped>
.mobile-goods-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 20px;
}

.header {
  background: #22BF57;
  color: white;
  padding: 20px;
  text-align: center;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.loading, .error, .empty {
  text-align: center;
  padding: 60px 20px;
  color: #666;
  font-size: 18px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error {
  color: #f56c6c;
}

.goods-list {
  padding: 15px;
}

.goods-item {
  background: white;
  border-radius: 12px;
  padding: 15px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  display: flex;
  gap: 15px;
}

.goods-image {
  width: 100px;
  height: 100px;
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
  background-color: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.goods-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-image {
  color: #999;
  font-size: 12px;
  text-align: center;
  padding: 10px;
}

.goods-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.goods-name {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #333;
  line-height: 1.4;
}

.goods-details {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
  gap: 15px;
}

.goods-spec {
  margin: 0;
  font-size: 18px;
  color: #666;
  flex: 1;
}

.quantity {
  font-size: 24px;
  color: #409eff;
  font-weight: 700;
  white-space: nowrap;
}

.summary {
  background: white;
  border-radius: 12px;
  padding: 20px;
  margin-top: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.summary-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 20px;
}

.summary-item:not(:last-child) {
  border-bottom: 1px solid #eee;
}

.summary-item span {
  color: #666;
  font-size: 20px;
}

.summary-item strong {
  color: #333;
  font-size: 22px;
  font-weight: 600;
}

/* 移动端适配 */
@media (max-width: 480px) {
  .header {
    padding: 15px;
  }
  
  .header h1 {
    font-size: 18px;
  }
  
  .goods-item {
    padding: 12px;
    gap: 12px;
  }
  
  .goods-image {
    width: 80px;
    height: 80px;
  }
  
  .goods-name {
    font-size: 18px;
  }
  
  .goods-spec {
    font-size: 16px;
  }
  
  .quantity {
    font-size: 20px;
  }
  
  .summary-item {
    font-size: 18px;
  }
  
  .summary-item span {
    font-size: 18px;
  }
  
  .summary-item strong {
    font-size: 20px;
  }
}
</style>

