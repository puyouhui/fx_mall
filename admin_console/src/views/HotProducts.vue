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
        <el-table ref="tableRef" :data="hotProducts" stripe row-key="id">
          <el-table-column type="index" label="序号" width="60" align="center" />
          <el-table-column label="排序" width="140" align="center">
            <template #default="scope">
              <div class="sort-actions">
                <el-icon class="drag-handle" title="拖动排序"><Rank /></el-icon>
                <el-button-group>
                  <el-button size="small" :disabled="scope.$index === 0" @click="moveUp(scope.$index)">
                    <el-icon><ArrowUp /></el-icon>
                  </el-button>
                  <el-button size="small" :disabled="scope.$index === hotProducts.length - 1" @click="moveDown(scope.$index)">
                    <el-icon><ArrowDown /></el-icon>
                  </el-button>
                </el-button-group>
              </div>
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
      <el-dialog v-model="dialogVisible" title="添加热销产品" width="700px" :close-on-click-modal="false" class="hot-product-dialog">
        <div class="product-selector">
          <!-- 搜索和筛选栏 -->
          <div class="filter-bar">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索商品名称"
              clearable
              style="width: 220px; margin-right: 12px;"
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-cascader
              v-model="selectedCategoryId"
              :options="treeCategories"
              :props="{ checkStrictly: true, label: 'name', value: 'id', children: 'children', emitPath: false }"
              placeholder="选择分类（一级或二级）"
              clearable
              style="width: 220px; margin-right: 12px;"
              @change="handleCategoryChange"
            />
            <el-button type="primary" @click="handleSearch">搜索</el-button>
          </div>

          <!-- 分页（置于列表上方，确保可见） -->
          <div class="pagination-bar" v-if="productPagination.total > 0">
            <el-pagination
              v-model:current-page="productPagination.pageNum"
              v-model:page-size="productPagination.pageSize"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next"
              :total="productPagination.total"
              small
              @size-change="handleProductSizeChange"
              @current-change="handleProductPageChange"
            />
          </div>

          <!-- 商品列表 -->
          <div class="product-list-wrap" v-loading="productLoading">
            <div
              v-for="product in products"
              :key="product.id"
              class="product-list-row"
              :class="{ 'selected': form.product_ids.includes(product.id) }"
              @click="toggleProduct(product.id)"
            >
              <div class="list-row-image">
                <el-image
                  v-if="product.images && product.images.length > 0"
                  :src="getImageUrl(product.images[0])"
                  fit="cover"
                  class="thumb"
                />
                <div v-else class="thumb-placeholder">
                  <el-icon><Picture /></el-icon>
                </div>
              </div>
              <div class="list-row-content">
                <div class="list-row-name">{{ product.name }}</div>
                <div class="list-row-meta">
                  <span v-if="(product.category_name || getCategoryName(product.categoryId || product.category_id))" class="category-tag">
                    {{ product.category_name || getCategoryName(product.categoryId || product.category_id) }}
                  </span>
                  <span class="price-tag">{{ calculatePriceRange(product.specs) }}</span>
                </div>
              </div>
              <div class="list-row-action" @click.stop="toggleProduct(product.id)">
                <el-checkbox :model-value="form.product_ids.includes(product.id)">
                  {{ form.product_ids.includes(product.id) ? '已选' : '选择' }}
                </el-checkbox>
              </div>
            </div>
            <div v-if="products.length === 0 && !productLoading" class="empty-products">
              <el-empty description="暂无商品" :image-size="60" />
            </div>
          </div>
        </div>

        <template #footer>
          <div class="dialog-footer-content">
            <div class="selected-info" v-if="form.product_ids.length > 0">
              已选择 {{ form.product_ids.length }} 个商品
            </div>
            <div class="footer-btns">
              <el-button @click="dialogVisible = false">取消</el-button>
              <el-button type="primary" @click="handleSubmit" :disabled="form.product_ids.length === 0">确定添加</el-button>
            </div>
          </div>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, ArrowUp, ArrowDown, Search, Picture, Rank } from '@element-plus/icons-vue'
import { getAllHotProducts, createHotProduct, updateHotProduct, deleteHotProduct, updateHotProductSort } from '../api/hotProduct'
import { getProductList } from '../api/product'
import { getCategoryList } from '../api/category'
import { formatDate } from '../utils/time-format'
import Sortable from 'sortablejs'

// 热销产品列表
const hotProducts = ref([])
const tableRef = ref(null)
const dialogVisible = ref(false)
const formRef = ref(null)
const productLoading = ref(false)
const products = ref([])
const categories = ref([])
const treeCategories = ref([])
const searchKeyword = ref('')
const selectedCategoryId = ref('')

const form = reactive({
  product_ids: []
})

const productPagination = reactive({
  pageNum: 1,
  pageSize: 10,
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

// 获取图片URL
const getImageUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url
  }
  return `http://113.44.164.151:9000${url}`
}

// 拖拽实例（用于销毁重建）
let sortableInstance = null

// 初始化拖拽排序
const initSortable = () => {
  nextTick(() => {
    if (sortableInstance) {
      sortableInstance.destroy()
      sortableInstance = null
    }
    if (!tableRef.value || hotProducts.value.length === 0) return
    const tbody = tableRef.value.$el.querySelector('.el-table__body-wrapper tbody')
    if (!tbody) return
    sortableInstance = Sortable.create(tbody, {
      handle: '.drag-handle',
      animation: 150,
      onEnd: async (evt) => {
        const { oldIndex, newIndex } = evt
        if (oldIndex === newIndex) return
        const moved = hotProducts.value.splice(oldIndex, 1)[0]
        hotProducts.value.splice(newIndex, 0, moved)
        await saveSort()
      }
    })
  })
}

// 初始化数据
const initData = async () => {
  try {
    const response = await getAllHotProducts()
    if (response && response.code === 200) {
      hotProducts.value = response.data || []
      // 按排序值排序
      hotProducts.value.sort((a, b) => a.sort - b.sort)
      initSortable()
    } else {
      ElMessage.error(response?.message || '加载热销产品失败')
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载热销产品失败')
  }
}

// 扁平分展开（用于 getCategoryName 等）
const flattenCategories = (list) => {
  let result = []
  ;(list || []).forEach(item => {
    result.push({ id: item.id, name: item.name, parent_id: item.parent_id })
    if (item.children?.length) {
      result = result.concat(flattenCategories(item.children))
    }
  })
  return result
}

// 将扁平列表构建为树形（用于级联选择器）
const buildCategoryTree = (flatList) => {
  const map = new Map()
  const roots = []
  ;(flatList || []).forEach(cat => {
    map.set(cat.id, { ...cat, children: [] })
  })
  flatList.forEach(cat => {
    if (!cat.parent_id || cat.parent_id === 0) {
      roots.push(map.get(cat.id))
    } else {
      const parent = map.get(cat.parent_id)
      if (parent) {
        parent.children.push(map.get(cat.id))
      }
    }
  })
  return roots
}

// 加载分类列表
const loadCategories = async () => {
  try {
    const response = await getCategoryList()
    if (response && response.code === 200) {
      const raw = response.data || []
      const isTree = raw.length > 0 && Array.isArray(raw[0]?.children)
      treeCategories.value = isTree ? raw : buildCategoryTree(raw)
      categories.value = isTree ? flattenCategories(raw) : raw
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
    if (selectedCategoryId.value != null && selectedCategoryId.value !== '') {
      params.categoryId = selectedCategoryId.value
    }

    const response = await getProductList(params)
    if (response && response.code === 200) {
      const list = Array.isArray(response.data) ? response.data : (response.data?.list || [])
      products.value = list
      productPagination.total = response.total ?? list.length
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

// 切换商品选中状态
const toggleProduct = (productId) => {
  const idx = form.product_ids.indexOf(productId)
  if (idx >= 0) {
    form.product_ids.splice(idx, 1)
  } else {
    form.product_ids.push(productId)
  }
}

// 打开添加弹窗
const handleAddHotProduct = async () => {
  dialogVisible.value = true
  form.product_ids = []
  searchKeyword.value = ''
  selectedCategoryId.value = null
  productPagination.pageNum = 1
  await loadCategories()
  await loadProducts()
}

// 提交表单
const handleSubmit = async () => {
  if (form.product_ids.length === 0) {
    ElMessage.warning('请选择商品')
    return
  }
  let successCount = 0
  let failCount = 0
  try {
    for (const productId of form.product_ids) {
      try {
        const response = await createHotProduct({ product_id: productId })
        if (response && response.code === 200) {
          successCount++
        } else {
          failCount++
        }
      } catch {
        failCount++
      }
    }
    if (successCount > 0) {
      ElMessage.success(`成功添加 ${successCount} 个热销产品${failCount > 0 ? `，${failCount} 个已存在跳过` : ''}`)
      dialogVisible.value = false
      await initData()
    } else {
      ElMessage.warning(failCount > 0 ? '所选商品可能已是热销产品，请勿重复添加' : '添加失败')
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

.sort-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.sort-actions .drag-handle {
  cursor: grab;
  color: #909399;
  font-size: 18px;
}

.sort-actions .drag-handle:hover {
  color: #409eff;
}

.sort-actions .drag-handle:active {
  cursor: grabbing;
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

/* 商品选择器样式 - 列表模式 */
.product-selector {
  display: flex;
  flex-direction: column;
}

.filter-bar {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
}

.product-list-wrap {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
}

.product-list-row {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
  transition: background 0.2s;
}

.product-list-row:last-child {
  border-bottom: none;
}

.product-list-row:hover {
  background: #f5f7fa;
}

.product-list-row.selected {
  background: #ecf5ff;
}

.list-row-image {
  flex-shrink: 0;
  margin-right: 12px;
}

.list-row-image .thumb {
  width: 48px;
  height: 48px;
  border-radius: 4px;
  object-fit: cover;
}

.thumb-placeholder {
  width: 48px;
  height: 48px;
  background: #f5f7fa;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #c0c4cc;
}

.list-row-content {
  flex: 1;
  min-width: 0;
}

.list-row-name {
  font-size: 14px;
  color: #303133;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.list-row-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.category-tag {
  color: #409eff;
  background: #ecf5ff;
  padding: 2px 6px;
  border-radius: 3px;
}

.price-tag {
  color: #f56c6c;
  font-weight: 500;
}

.list-row-action {
  flex-shrink: 0;
}

.empty-products {
  padding: 40px 20px;
  text-align: center;
}

.dialog-footer-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.footer-btns {
  display: flex;
  gap: 8px;
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

