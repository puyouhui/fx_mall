<template>
  <div class="hot-products-container">
    <el-card>
      <h2 class="page-title">热销产品管理</h2>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-right">
          <el-button type="primary" @click="handleAddHotProduct">
            <el-icon>
              <Plus />
            </el-icon>
            添加热销产品
          </el-button>
        </div>
      </div>

      <!-- 热销产品列表 -->
      <el-card class="products-card">
        <el-table :data="hotProducts" stripe row-key="id">
          <el-table-column type="index" label="序号" width="60" align="center" />
          <el-table-column label="排序" width="120" align="center">
            <template #default="scope">
              <el-button-group>
                <el-button size="small" :disabled="scope.$index === 0" @click="moveUp(scope.$index)">
                  <el-icon><ArrowUp /></el-icon>
                </el-button>
                <el-button size="small" :disabled="scope.$index === hotProducts.length - 1" @click="moveDown(scope.$index)">
                  <el-icon><ArrowDown /></el-icon>
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
          <el-table-column label="商品信息" min-width="200">
            <template #default="scope">
              <div class="product-info">
                <el-image
                  v-if="scope.row.product && scope.row.product.images && scope.row.product.images.length > 0"
                  :src="getImageUrl(scope.row.product.images[0])"
                  style="width: 60px; height: 60px; margin-right: 10px;"
                  fit="cover"
                />
                <div>
                  <div class="product-name">{{ scope.row.product?.name || '未知商品' }}</div>
                  <div class="product-id">商品ID: {{ scope.row.product_id }}</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="sort" label="排序值" width="100" align="center" />
          <el-table-column label="状态" width="100" align="center">
            <template #default="scope">
              <el-switch
                v-model="scope.row.status"
                :active-value="1"
                :inactive-value="0"
                @change="handleStatusChange(scope.row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="180" align="center">
            <template #default="scope">
              {{ formatDate(scope.row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" width="150" align="center">
            <template #default="scope">
              <el-button type="danger" size="small" @click="handleDelete(scope.row.id)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 空状态 -->
        <div v-if="hotProducts.length === 0" class="empty-state">
          <el-empty description="暂无热销产品" />
        </div>
      </el-card>

      <!-- 添加热销产品弹窗 -->
      <el-dialog v-model="dialogVisible" title="添加热销产品" width="90%" :close-on-click-modal="false" class="hot-product-dialog">
        <div class="product-selector">
          <!-- 搜索和筛选栏 -->
          <div class="filter-bar">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索商品名称"
              clearable
              style="width: 300px; margin-right: 20px;"
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-select
              v-model="selectedCategoryId"
              placeholder="选择分类"
              clearable
              style="width: 200px; margin-right: 20px;"
              @change="handleCategoryChange"
            >
              <el-option label="全部分类" value="" />
              <el-option
                v-for="category in categories"
                :key="category.id"
                :label="category.name"
                :value="category.id"
              />
            </el-select>
            <el-button type="primary" @click="loadProducts">刷新</el-button>
          </div>

          <!-- 商品卡片列表 -->
          <div class="product-grid" v-loading="productLoading">
            <div
              v-for="product in filteredProducts"
              :key="product.id"
              class="product-card"
              :class="{ 'selected': form.product_id === product.id }"
              @click="selectProduct(product.id)"
            >
              <div class="card-image">
                <el-image
                  v-if="product.images && product.images.length > 0"
                  :src="getImageUrl(product.images[0])"
                  fit="cover"
                  class="product-image"
                />
                <div v-else class="image-placeholder">
                  <el-icon><Picture /></el-icon>
                  <span>暂无图片</span>
                </div>
                <div v-if="form.product_id === product.id" class="selected-badge">
                  <el-icon><Check /></el-icon>
                </div>
              </div>
              <div class="card-content">
                <div class="product-title">{{ product.name }}</div>
                <div class="product-meta">
                  <span v-if="getCategoryName(product.categoryId)" class="product-category">
                    {{ getCategoryName(product.categoryId) }}
                  </span>
                </div>
                <div class="product-price">
                  {{ calculatePriceRange(product.specs) }}
                </div>
                <div v-if="product.description" class="product-description">
                  {{ truncateText(product.description, 50) }}
                </div>
              </div>
            </div>

            <!-- 空状态 -->
            <div v-if="filteredProducts.length === 0 && !productLoading" class="empty-products">
              <el-empty description="暂无商品" />
            </div>
          </div>
        </div>

        <template #footer>
          <div class="dialog-footer-content">
            <!-- 分页 -->
            <div class="pagination-container" v-if="productPagination.total > 0">
              <el-pagination
                v-model:current-page="productPagination.pageNum"
                v-model:page-size="productPagination.pageSize"
                :page-sizes="[12, 24, 48, 96]"
                layout="total, sizes, prev, pager, next, jumper"
                :total="productPagination.total"
                @size-change="handleProductSizeChange"
                @current-change="handleProductPageChange"
              />
            </div>
            <!-- 操作按钮 -->
            <div class="footer-actions">
              <div class="selected-info" v-if="form.product_id">
                <span>已选择商品ID: {{ form.product_id }}</span>
              </div>
              <div>
                <el-button @click="dialogVisible = false">取消</el-button>
                <el-button type="primary" @click="handleSubmit" :disabled="!form.product_id">确定</el-button>
              </div>
            </div>
          </div>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, ArrowUp, ArrowDown, Search, Picture, Check } from '@element-plus/icons-vue'
import { getAllHotProducts, createHotProduct, updateHotProduct, deleteHotProduct, updateHotProductSort } from '../api/hotProduct'
import { getProductList } from '../api/product'
import { getCategoryList } from '../api/category'
import { formatDate } from '../utils/time-format'

// 热销产品列表
const hotProducts = ref([])
const dialogVisible = ref(false)
const formRef = ref(null)
const productLoading = ref(false)
const products = ref([])
const categories = ref([])
const searchKeyword = ref('')
const selectedCategoryId = ref('')

const form = reactive({
  product_id: null
})

const productPagination = reactive({
  pageNum: 1,
  pageSize: 12,
  total: 0
})

// 计算价格范围
const calculatePriceRange = (specs) => {
  if (!specs || !Array.isArray(specs) || specs.length === 0) {
    return '暂无价格'
  }

  const allPrices = []
  specs.forEach(spec => {
    const wholesalePrice = spec.wholesale_price || spec.wholesalePrice
    if (wholesalePrice && wholesalePrice > 0) {
      allPrices.push(wholesalePrice)
    }
    const retailPrice = spec.retail_price || spec.retailPrice
    if (retailPrice && retailPrice > 0) {
      allPrices.push(retailPrice)
    }
  })

  if (allPrices.length === 0) {
    return '暂无价格'
  }

  const minPrice = Math.min(...allPrices)
  const maxPrice = Math.max(...allPrices)

  if (minPrice === maxPrice) {
    return `¥${minPrice.toFixed(2)}`
  } else {
    return `¥${minPrice.toFixed(2)} - ¥${maxPrice.toFixed(2)}`
  }
}

// 截断文本
const truncateText = (text, length) => {
  if (!text) return ''
  return text.length > length ? text.substring(0, length) + '...' : text
}

// 获取分类名称
const getCategoryName = (categoryId) => {
  const category = categories.value.find(cat => cat.id === categoryId)
  return category ? category.name : ''
}

// 过滤后的商品列表
const filteredProducts = computed(() => {
  let result = products.value

  // 按关键词过滤
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(product => 
      product.name.toLowerCase().includes(keyword) ||
      (product.description && product.description.toLowerCase().includes(keyword))
    )
  }

  // 按分类过滤
  if (selectedCategoryId.value) {
    result = result.filter(product => product.categoryId === selectedCategoryId.value)
  }

  return result
})

// 获取图片URL
const getImageUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url
  }
  return `http://113.44.164.151:9000${url}`
}

// 初始化数据
const initData = async () => {
  try {
    const response = await getAllHotProducts()
    if (response && response.code === 200) {
      hotProducts.value = response.data || []
      // 按排序值排序
      hotProducts.value.sort((a, b) => a.sort - b.sort)
    } else {
      ElMessage.error(response?.message || '加载热销产品失败')
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载热销产品失败')
  }
}

// 加载分类列表
const loadCategories = async () => {
  try {
    const response = await getCategoryList()
    if (response && response.code === 200) {
      categories.value = response.data || []
    }
  } catch (error) {
    console.error('加载分类失败:', error)
  }
}

// 加载商品列表
const loadProducts = async () => {
  productLoading.value = true
  try {
    const params = {
      pageNum: productPagination.pageNum,
      pageSize: productPagination.pageSize
    }
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }
    if (selectedCategoryId.value) {
      params.categoryId = selectedCategoryId.value
    }

    const response = await getProductList(params)
    if (response && response.code === 200 && response.data) {
      products.value = response.data.list || response.data || []
      productPagination.total = response.data.total || products.value.length
    } else {
      products.value = []
      productPagination.total = 0
    }
  } catch (error) {
    console.error('加载商品列表失败:', error)
    products.value = []
    productPagination.total = 0
  } finally {
    productLoading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  productPagination.pageNum = 1
  loadProducts()
}

// 分类筛选处理
const handleCategoryChange = () => {
  productPagination.pageNum = 1
  loadProducts()
}

// 分页大小改变
const handleProductSizeChange = (size) => {
  productPagination.pageSize = size
  productPagination.pageNum = 1
  loadProducts()
}

// 页码改变
const handleProductPageChange = (page) => {
  productPagination.pageNum = page
  loadProducts()
}

// 选择商品
const selectProduct = (productId) => {
  form.product_id = productId
}

// 打开添加弹窗
const handleAddHotProduct = async () => {
  dialogVisible.value = true
  form.product_id = null
  searchKeyword.value = ''
  selectedCategoryId.value = ''
  productPagination.pageNum = 1
  await loadCategories()
  await loadProducts()
}

// 提交表单
const handleSubmit = async () => {
  if (!form.product_id) {
    ElMessage.warning('请选择商品')
    return
  }
  try {
    const response = await createHotProduct({ product_id: form.product_id })
    if (response && response.code === 200) {
      ElMessage.success('添加成功')
      dialogVisible.value = false
      await initData()
    } else {
      ElMessage.error(response?.message || '添加失败')
    }
  } catch (error) {
    console.error('添加失败:', error)
    ElMessage.error('添加失败')
  }
}

// 删除
const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个热销产品吗？', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const response = await deleteHotProduct(id)
    if (response && response.code === 200) {
      ElMessage.success('删除成功')
      await initData()
    } else {
      ElMessage.error(response?.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 状态变更
const handleStatusChange = async (row) => {
  try {
    const response = await updateHotProduct(row.id, { status: row.status })
    if (response && response.code === 200) {
      ElMessage.success('更新成功')
    } else {
      // 恢复原状态
      row.status = row.status === 1 ? 0 : 1
      ElMessage.error(response?.message || '更新失败')
    }
  } catch (error) {
    // 恢复原状态
    row.status = row.status === 1 ? 0 : 1
    console.error('更新失败:', error)
    ElMessage.error('更新失败')
  }
}

// 上移
const moveUp = async (index) => {
  if (index === 0) return
  const temp = hotProducts.value[index]
  hotProducts.value[index] = hotProducts.value[index - 1]
  hotProducts.value[index - 1] = temp
  await saveSort()
}

// 下移
const moveDown = async (index) => {
  if (index === hotProducts.value.length - 1) return
  const temp = hotProducts.value[index]
  hotProducts.value[index] = hotProducts.value[index + 1]
  hotProducts.value[index + 1] = temp
  await saveSort()
}

// 保存排序
const saveSort = async () => {
  try {
    const items = hotProducts.value.map((item, index) => ({
      id: item.id,
      sort: index + 1
    }))
    const response = await updateHotProductSort(items)
    if (response && response.code === 200) {
      // 更新本地排序值
      hotProducts.value.forEach((item, index) => {
        item.sort = index + 1
      })
      ElMessage.success('排序已更新')
    } else {
      ElMessage.error(response?.message || '更新排序失败')
      await initData() // 重新加载数据
    }
  } catch (error) {
    console.error('更新排序失败:', error)
    ElMessage.error('更新排序失败')
    await initData() // 重新加载数据
  }
}

onMounted(() => {
  initData()
})
</script>

<style scoped>
.hot-products-container {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px 0;
  font-size: 24px;
  font-weight: bold;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}

.products-card {
  margin-top: 20px;
}

.product-info {
  display: flex;
  align-items: center;
}

.product-name {
  font-weight: bold;
  margin-bottom: 5px;
}

.product-id {
  font-size: 12px;
  color: #999;
}

.product-option {
  display: flex;
  align-items: center;
}

.empty-state {
  padding: 40px;
  text-align: center;
}

/* 商品选择器样式 */
.product-selector {
  display: flex;
  flex-direction: column;
  min-height: 500px;
}

.filter-bar {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #eee;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 12px;
  margin-bottom: 20px;
  align-items: start;
  flex: 1;
}

.product-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fff;
  display: flex;
  flex-direction: column;
  height: auto;
  max-height: none;
}

.product-card:hover {
  border-color: #409eff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
  transform: translateY(-2px);
}

.product-card.selected {
  border-color: #409eff;
  border-width: 2px;
  background: linear-gradient(to bottom, #ecf5ff 0%, #fff 20%);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.2);
}

.card-image {
  position: relative;
  width: 100%;
  height: 100px;
  background: #f5f7fa;
  flex-shrink: 0;
}

.product-image {
  width: 100%;
  height: 100%;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.image-placeholder {
  font-size: 12px;
}

.image-placeholder .el-icon {
  font-size: 28px;
  margin-bottom: 6px;
  color: #c0c4cc;
}

.image-placeholder span {
  font-size: 11px;
}

.selected-badge {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 22px;
  height: 22px;
  background: #409eff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 12px;
  box-shadow: 0 2px 6px rgba(64, 158, 255, 0.4);
  z-index: 10;
}

.card-content {
  padding: 8px;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.product-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.4;
}

.product-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 6px;
}

.product-id {
  font-size: 11px;
  color: #909399;
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  display: inline-block;
  width: fit-content;
}

.product-category {
  font-size: 11px;
  color: #409eff;
  background: #ecf5ff;
  padding: 2px 6px;
  border-radius: 3px;
  white-space: nowrap;
  display: inline-block;
  width: fit-content;
}

.product-price {
  font-size: 14px;
  font-weight: 600;
  color: #f56c6c;
  margin-bottom: 4px;
}

.product-description {
  font-size: 11px;
  color: #909399;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  margin-top: auto;
}

.empty-products {
  grid-column: 1 / -1;
  padding: 60px 20px;
  text-align: center;
}

.dialog-footer-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-top: 16px;
  border-top: 1px solid #eee;
}

.pagination-container {
  display: flex;
  justify-content: center;
  flex-shrink: 0;
}

.footer-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.selected-info {
  color: #409eff;
  font-size: 14px;
}

/* 弹窗样式调整 */
:deep(.hot-product-dialog .el-dialog__body) {
  padding-bottom: 0;
}

:deep(.hot-product-dialog .el-dialog__footer) {
  padding-top: 16px;
  border-top: 1px solid #eee;
}
</style>

