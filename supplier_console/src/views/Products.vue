<template>
  <div class="products-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-title">
            <span>商品管理</span>
            <el-tag type="info" size="small" style="margin-left: 10px;">
              共 {{ pagination.total }} 件
            </el-tag>
          </div>
          <div class="header-actions">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索商品名称"
              clearable
              style="width: 300px; margin-right: 10px;"
              @clear="handleSearch"
              @keyup.enter="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-button type="primary" @click="handleSearch">
              <el-icon><Search /></el-icon>
              搜索
            </el-button>
          </div>
        </div>
      </template>
      
      <el-table 
        :data="products" 
        style="width: 100%" 
        v-loading="loading"
        stripe
      >
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column label="商品图片" width="100" align="center">
          <template #default="scope">
            <el-image
              v-if="scope.row.images && scope.row.images.length > 0"
              :src="scope.row.images[0]"
              :preview-src-list="scope.row.images"
              fit="cover"
              style="width: 50px; height: 50px; border-radius: 4px;"
              :preview-teleported="true"
            />
            <span v-else class="no-image">暂无图片</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="商品名称" min-width="200" align="center" show-overflow-tooltip />
        <el-table-column prop="description" label="商品描述" min-width="200" align="center" show-overflow-tooltip />
        <el-table-column label="规格数量" width="100" align="center">
          <template #default="scope">
            <span>{{ scope.row.spec_count || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="价格范围" width="180" align="center">
          <template #default="scope">
            <div>
              <span v-if="scope.row.min_cost_price !== undefined && scope.row.max_cost_price !== undefined" class="cost-price">
                <span v-if="scope.row.min_cost_price === scope.row.max_cost_price">
                  ¥{{ formatPrice(scope.row.min_cost_price) }}
                </span>
                <span v-else>
                  ¥{{ formatPrice(scope.row.min_cost_price) }}-¥{{ formatPrice(scope.row.max_cost_price) }}
                </span>
              </span>
              <span v-else class="cost-price">¥{{ formatPrice(getCostPrice(scope.row)) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="是否精选" width="100" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.is_special ? 'success' : 'info'">
              {{ scope.row.is_special ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'">
              {{ scope.row.status === 1 ? '上架' : '下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" align="center">
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template #default="scope">
            <el-button 
              type="primary" 
              size="small" 
              link
              @click="handleViewDetail(scope.row)"
            >
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 商品详情抽屉 -->
    <el-drawer
      v-model="detailDrawerVisible"
      title="商品详情"
      :size="600"
      direction="rtl"
    >
      <div v-if="currentProduct" class="product-detail">
        <!-- 商品图片 -->
        <div class="detail-section">
          <h3>商品图片</h3>
          <div v-if="currentProduct.images && currentProduct.images.length > 0" class="image-gallery">
            <el-image
              v-for="(image, index) in currentProduct.images"
              :key="index"
              :src="image"
              :preview-src-list="currentProduct.images"
              :initial-index="index"
              fit="cover"
              class="detail-image"
              :preview-teleported="true"
            />
          </div>
          <span v-else class="no-image">暂无图片</span>
        </div>

        <!-- 基本信息 -->
        <div class="detail-section">
          <h3>基本信息</h3>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="商品ID">{{ currentProduct.id }}</el-descriptions-item>
            <el-descriptions-item label="商品名称">{{ currentProduct.name }}</el-descriptions-item>
            <el-descriptions-item label="商品描述">
              <div class="description-text">{{ currentProduct.description || '暂无描述' }}</div>
            </el-descriptions-item>
            <el-descriptions-item label="是否精选">
              <el-tag :type="currentProduct.is_special ? 'success' : 'info'">
                {{ currentProduct.is_special ? '是' : '否' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="currentProduct.status === 1 ? 'success' : 'danger'">
                {{ currentProduct.status === 1 ? '上架' : '下架' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatDateTime(currentProduct.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">
              {{ formatDateTime(currentProduct.updated_at) }}
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 商品规格 -->
        <div v-if="currentProduct.specs && currentProduct.specs.length > 0" class="detail-section">
          <h3>商品规格</h3>
          <el-table :data="currentProduct.specs" border style="width: 100%">
            <el-table-column prop="name" label="规格名称" width="150" align="center" />
            <el-table-column prop="description" label="规格描述" width="200" align="center" show-overflow-tooltip />
            <el-table-column label="价格" width="120" align="center">
              <template #default="scope">
                <span v-if="scope.row.cost">¥{{ formatPrice(scope.row.cost) }}</span>
                <span v-else>-</span>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
      <div v-else class="loading-container">
        <el-skeleton :rows="5" animated />
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getProducts, getProductDetail } from '../api/products'

const loading = ref(false)
const products = ref([])
const searchKeyword = ref('')
const detailDrawerVisible = ref(false)
const currentProduct = ref(null)

const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0
})

// 加载商品列表
const loadProducts = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.page,
      page_size: pagination.value.pageSize
    }
    
    if (searchKeyword.value) {
      // 如果后端支持搜索，可以添加搜索参数
      // params.keyword = searchKeyword.value
    }
    
    const response = await getProducts(params)
    
    if (response.code === 200 && response.data) {
      products.value = response.data.list || []
      pagination.value.total = response.data.total || 0
    } else {
      ElMessage.error(response.message || '获取商品列表失败')
    }
  } catch (error) {
    console.error('获取商品列表失败:', error)
    ElMessage.error(error.message || '获取商品列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

// 搜索商品
const handleSearch = () => {
  pagination.value.page = 1
  loadProducts()
}

// 查看商品详情
const handleViewDetail = async (product) => {
  detailDrawerVisible.value = true
  currentProduct.value = null
  
  try {
    const response = await getProductDetail(product.id)
    if (response.code === 200 && response.data) {
      currentProduct.value = response.data
    } else {
      ElMessage.error(response.message || '获取商品详情失败')
      // 如果获取详情失败，使用列表中的基本信息
      currentProduct.value = product
    }
  } catch (error) {
    console.error('获取商品详情失败:', error)
    ElMessage.error(error.message || '获取商品详情失败，请稍后再试')
    // 如果获取详情失败，使用列表中的基本信息
    currentProduct.value = product
  }
}

// 分页大小改变
const handleSizeChange = (size) => {
  pagination.value.pageSize = size
  pagination.value.page = 1
  loadProducts()
}

// 页码改变
const handlePageChange = (page) => {
  pagination.value.page = page
  loadProducts()
}

// 获取成本价（从规格中取最小值，如果没有规格则使用商品价格）
const getCostPrice = (product) => {
  if (!product) return 0
  // 如果有规格，取所有规格中成本价的最小值
  if (product.specs && product.specs.length > 0) {
    let minCost = 0
    for (const spec of product.specs) {
      if (spec.cost && spec.cost > 0) {
        if (minCost === 0 || spec.cost < minCost) {
          minCost = spec.cost
        }
      }
    }
    if (minCost > 0) {
      return minCost
    }
  }
  // 如果没有规格或规格中没有成本价，使用商品价格字段
  return product.price || 0
}

// 格式化价格
const formatPrice = (price) => {
  if (price === undefined || price === null) return '0.00'
  return Number(price).toFixed(2)
}

// 格式化日期时间
const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  const date = new Date(dateTime)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

onMounted(() => {
  loadProducts()
})
</script>

<style scoped>
.products-page {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.no-image {
  color: #909399;
  font-size: 12px;
}

.price-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.original-price {
  text-decoration: line-through;
  color: #909399;
  font-size: 12px;
}

.cost-price {
  color: #409eff;
  font-weight: 500;
}

.product-detail {
  padding: 0;
}

.detail-section {
  margin-bottom: 30px;
}

.detail-section h3 {
  margin-bottom: 15px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  border-left: 4px solid #409eff;
  padding-left: 10px;
}

.image-gallery {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.detail-image {
  width: 120px;
  height: 120px;
  border-radius: 4px;
  cursor: pointer;
}

.description-text {
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}

.loading-container {
  padding: 20px;
}
</style>
