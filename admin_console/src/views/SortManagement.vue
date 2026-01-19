<template>
  <div class="sort-management-container">
    <el-card>
      <h2 class="page-title">排序管理</h2>

      <div class="content-wrapper">
        <!-- 左侧：一级分类列表（简化版） -->
        <div class="left-sidebar">
          <div class="sidebar-title">一级分类</div>
          <div class="category-list">
            <div
              v-for="category in primaryCategories"
              :key="category.id"
              class="category-item"
              :class="{ active: selectedPrimaryCategoryId === category.id }"
              @click="handlePrimaryCategoryRowClick(category)"
            >
              <el-image
                v-if="category.icon"
                :src="category.icon"
                class="category-icon"
                fit="cover"
              />
              <div v-else class="category-icon-placeholder">
                <el-icon><Grid /></el-icon>
              </div>
              <span class="category-name-text">{{ category.name }}</span>
            </div>
            <el-empty v-if="primaryCategories.length === 0" description="暂无一级分类" />
          </div>
        </div>

        <!-- 中间：二级分类排序 -->
        <div class="middle-content">
          <div class="sort-section">
            <div class="section-header">
              <h3>
                二级分类
                <span v-if="selectedPrimaryCategoryName" class="category-name">
                  ({{ selectedPrimaryCategoryName }})
                </span>
              </h3>
              <el-button type="primary" @click="saveCategorySort" :loading="categorySaving" :disabled="!selectedPrimaryCategoryId">
                保存
              </el-button>
            </div>
            <el-table
              v-if="selectedPrimaryCategoryId"
              :data="secondaryCategories"
              stripe
              row-key="id"
              class="sort-table"
              highlight-current-row
              @row-click="handleSecondaryCategoryRowClick"
            >
              <el-table-column type="index" label="序号" width="80" align="center" />
              <el-table-column label="排序操作" width="150" align="center">
                <template #default="scope">
                  <el-button-group>
                    <el-button
                      size="small"
                      :disabled="scope.$index === 0"
                      @click.stop="moveCategoryUp(scope.$index)"
                    >
                      <el-icon><ArrowUp /></el-icon>
                    </el-button>
                    <el-button
                      size="small"
                      :disabled="scope.$index === secondaryCategories.length - 1"
                      @click.stop="moveCategoryDown(scope.$index)"
                    >
                      <el-icon><ArrowDown /></el-icon>
                    </el-button>
                  </el-button-group>
                </template>
              </el-table-column>
              <el-table-column prop="name" label="分类名称" min-width="200" />
              <el-table-column prop="sort" label="排序值" width="120" align="center">
                <template #default="scope">
                  <el-input-number
                    v-model="scope.row.sort"
                    :min="0"
                    size="small"
                    style="width: 100px;"
                  />
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="100" align="center">
                <template #default="scope">
                  <el-tag :type="scope.row.status === 1 ? 'success' : 'info'">
                    {{ scope.row.status === 1 ? '启用' : '禁用' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!selectedPrimaryCategoryId" description="请选择左侧一级分类" />
            <el-empty v-else-if="selectedPrimaryCategoryId && secondaryCategories.length === 0" description="该分类下暂无二级分类" />
          </div>
        </div>

        <!-- 右侧：商品排序 -->
        <div class="right-content">
          <el-tabs v-model="productSortTab" class="product-sort-tabs">
            <!-- 分类商品排序 -->
            <el-tab-pane label="分类商品排序" name="category">
              <div class="sort-section">
                <div class="section-header">
                  <h3>
                    商品排序
                    <span v-if="selectedSecondaryCategoryName" class="category-name">
                      ({{ selectedSecondaryCategoryName }})
                    </span>
                  </h3>
                </div>
                <div class="section-actions">
                  <el-input
                    v-model="productSearchKeyword"
                    placeholder="搜索商品名称"
                    clearable
                    style="width: 200px; margin-right: 10px;"
                    @input="handleProductSearch"
                    :disabled="!selectedSecondaryCategoryId"
                  >
                    <template #prefix>
                      <el-icon><Search /></el-icon>
                    </template>
                  </el-input>
                  <el-button type="primary" @click="saveProductSort" :loading="productSaving" :disabled="!selectedSecondaryCategoryId">
                    保存商品排序
                  </el-button>
                </div>
                <el-table
                  v-if="selectedSecondaryCategoryId && filteredProducts.length > 0"
                  :data="filteredProducts"
                  stripe
                  row-key="id"
                  class="sort-table"
                >
                  <el-table-column type="index" label="序号" width="80" align="center" />
                  <el-table-column label="排序操作" width="150" align="center">
                    <template #default="scope">
                      <el-button-group>
                        <el-button
                          size="small"
                          :disabled="scope.$index === 0"
                          @click.stop="moveProductUp(scope.$index)"
                        >
                          <el-icon><ArrowUp /></el-icon>
                        </el-button>
                        <el-button
                          size="small"
                          :disabled="scope.$index === filteredProducts.length - 1"
                          @click.stop="moveProductDown(scope.$index)"
                        >
                          <el-icon><ArrowDown /></el-icon>
                        </el-button>
                      </el-button-group>
                    </template>
                  </el-table-column>
                  <el-table-column label="商品信息" min-width="200">
                    <template #default="scope">
                      <div class="product-info">
                        <el-image
                          v-if="scope.row.images && scope.row.images.length > 0"
                          :src="getImageUrl(scope.row.images[0])"
                          style="width: 50px; height: 50px; margin-right: 10px;"
                          fit="cover"
                        />
                        <div>
                          <div class="product-name">{{ scope.row.name }}</div>
                          <div class="product-id">ID: {{ scope.row.id }}</div>
                        </div>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column prop="sort" label="排序值" width="120" align="center">
                    <template #default="scope">
                      <el-input-number
                        v-model="scope.row.sort"
                        :min="0"
                        size="small"
                        style="width: 100px;"
                      />
                    </template>
                  </el-table-column>
                  <el-table-column prop="status" label="状态" width="100" align="center">
                    <template #default="scope">
                      <el-tag :type="scope.row.status === 1 ? 'success' : 'info'">
                        {{ scope.row.status === 1 ? '启用' : '禁用' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                </el-table>
                <template v-if="!selectedSecondaryCategoryId">
                  <el-empty description="请选择中间二级分类" />
                </template>
                <template v-else-if="selectedSecondaryCategoryId && filteredProducts.length === 0">
                  <el-empty description="该分类下暂无商品" />
                </template>
              </div>
            </el-tab-pane>

            <!-- 精选商品排序 -->
            <el-tab-pane label="精选商品排序" name="special">
              <div class="sort-section">
                <div class="section-header">
                  <h3>精选商品排序</h3>
                  <el-button type="primary" @click="saveSpecialProductSort" :loading="specialProductSaving">
                    保存精选商品排序
                  </el-button>
                </div>
                <el-table
                  :data="specialProducts"
                  stripe
                  row-key="id"
                  class="sort-table"
                >
                  <el-table-column type="index" label="序号" width="80" align="center" />
                  <el-table-column label="排序操作" width="150" align="center">
                    <template #default="scope">
                      <el-button-group>
                        <el-button
                          size="small"
                          :disabled="scope.$index === 0"
                          @click.stop="moveSpecialProductUp(scope.$index)"
                        >
                          <el-icon><ArrowUp /></el-icon>
                        </el-button>
                        <el-button
                          size="small"
                          :disabled="scope.$index === specialProducts.length - 1"
                          @click.stop="moveSpecialProductDown(scope.$index)"
                        >
                          <el-icon><ArrowDown /></el-icon>
                        </el-button>
                      </el-button-group>
                    </template>
                  </el-table-column>
                  <el-table-column label="商品信息" min-width="250">
                    <template #default="scope">
                      <div class="product-info">
                        <el-image
                          v-if="scope.row.images && scope.row.images.length > 0"
                          :src="getImageUrl(scope.row.images[0])"
                          style="width: 60px; height: 60px; margin-right: 10px;"
                          fit="cover"
                        />
                        <div>
                          <div class="product-name">{{ scope.row.name }}</div>
                          <div class="product-id">ID: {{ scope.row.id }}</div>
                        </div>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column prop="special_sort" label="排序值" width="120" align="center">
                    <template #default="scope">
                      <el-input-number
                        v-model="scope.row.special_sort"
                        :min="0"
                        size="small"
                        style="width: 100px;"
                      />
                    </template>
                  </el-table-column>
                  <el-table-column prop="status" label="状态" width="100" align="center">
                    <template #default="scope">
                      <el-tag :type="scope.row.status === 1 ? 'success' : 'info'">
                        {{ scope.row.status === 1 ? '启用' : '禁用' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="specialProducts.length === 0" description="暂无精选商品" />
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowUp, ArrowDown, Search, Grid } from '@element-plus/icons-vue'
import { getCategories } from '../api/categories'
import { getProductsByCategory } from '../api/products'
import { batchUpdateCategorySort, batchUpdateProductSort, getAllSpecialProducts, batchUpdateSpecialProductSort } from '../api/sort'

// 数据
const primaryCategories = ref([])
const secondaryCategories = ref([])
const allProducts = ref([])
const selectedPrimaryCategoryId = ref(null)
const selectedPrimaryCategoryName = ref('')
const selectedSecondaryCategoryId = ref(null)
const selectedSecondaryCategoryName = ref('')
const productSearchKeyword = ref('')
const categorySaving = ref(false)
const productSaving = ref(false)
const productSortTab = ref('category') // 商品排序标签：'category' 分类商品，'special' 精选商品
const specialProducts = ref([])
const specialProductSaving = ref(false)

// 计算属性：过滤后的商品列表
const filteredProducts = computed(() => {
  if (!productSearchKeyword.value) {
    return allProducts.value
  }
  const keyword = productSearchKeyword.value.toLowerCase()
  return allProducts.value.filter(product =>
    product.name.toLowerCase().includes(keyword)
  )
})

// 扁平化分类树形结构
const flattenCategories = (categories, result = []) => {
  categories.forEach(cat => {
    // 确保复制完整的分类对象，保留所有字段
    const categoryItem = {
      id: cat.id,
      name: cat.name,
      parent_id: cat.parent_id,
      sort: cat.sort || 0,
      status: cat.status,
      icon: cat.icon,
      created_at: cat.created_at,
      updated_at: cat.updated_at
    }
    result.push(categoryItem)
    
    // 递归处理子分类
    if (cat.children && Array.isArray(cat.children) && cat.children.length > 0) {
      flattenCategories(cat.children, result)
    }
  })
  return result
}

// 初始化数据
const initData = async () => {
  try {
    const categories = await getCategories()
    
    // getCategories() 返回的直接就是数组（树形结构）
    if (Array.isArray(categories) && categories.length > 0) {
      // 后端返回的是树形结构，需要扁平化处理
      const allCategories = flattenCategories(categories)
      
      // 获取一级分类（parent_id === 0）
      primaryCategories.value = allCategories.filter(cat => cat.parent_id === 0)
    }
  } catch (error) {
    console.error('加载分类失败:', error)
    ElMessage.error('加载分类失败')
  }
}

// 选择一级分类（从表格行点击触发）
const handlePrimaryCategoryRowClick = async (row) => {
  if (!row || !row.id) return
  
  selectedPrimaryCategoryId.value = row.id
  selectedPrimaryCategoryName.value = row.name
  selectedSecondaryCategoryId.value = null
  selectedSecondaryCategoryName.value = ''
  allProducts.value = []
  
  await loadSecondaryCategories(selectedPrimaryCategoryId.value)
}

// 兼容方法（用于保存后重新选中）
const handlePrimaryCategorySelect = async (row) => {
  await handlePrimaryCategoryRowClick(row)
}

// 加载二级分类
const loadSecondaryCategories = async (primaryCategoryId) => {
  try {
    const categories = await getCategories()
    
    if (Array.isArray(categories) && categories.length > 0) {
      // 找到对应的一级分类
      const primaryCategory = categories.find(cat => cat.id === primaryCategoryId)
      
      if (primaryCategory && primaryCategory.children) {
        // 直接从一级分类的children数组中获取二级分类
        secondaryCategories.value = (primaryCategory.children || [])
          .map(cat => ({ ...cat, sort: cat.sort || 0 }))
      } else {
        secondaryCategories.value = []
      }
    }
  } catch (error) {
    console.error('加载二级分类失败:', error)
    ElMessage.error('加载二级分类失败')
  }
}

// 选择二级分类（从表格行点击触发）
const handleSecondaryCategoryRowClick = async (row) => {
  if (!row || !row.id) return
  
  selectedSecondaryCategoryId.value = row.id
  selectedSecondaryCategoryName.value = row.name
  
  await loadProductsByCategory(selectedSecondaryCategoryId.value)
}

// 兼容方法（用于保存后重新选中）
const handleSecondaryCategorySelect = async (row) => {
  await handleSecondaryCategoryRowClick(row)
}

// 加载指定二级分类下的商品
const loadProductsByCategory = async (categoryId) => {
  try {
    console.log('开始加载商品，分类ID:', categoryId)
    const response = await getProductsByCategory(categoryId, 1, 1000)
    console.log('商品数据响应:', response)
    
    // 处理不同的响应格式
    let productList = []
    if (response) {
      // 如果响应有 code 和 data.list 结构
      if (response.code === 200 && response.data && response.data.list) {
        productList = response.data.list
      }
      // 如果响应直接是 data 对象，且有 list
      else if (response.data && response.data.list) {
        productList = response.data.list
      }
      // 如果响应直接是数组
      else if (Array.isArray(response)) {
        productList = response
      }
      // 如果响应直接是 { list: [...] } 结构
      else if (response.list && Array.isArray(response.list)) {
        productList = response.list
      }
    }
    
    console.log('处理后的商品列表:', productList)
    
    allProducts.value = productList.map(product => ({
      ...product,
      sort: product.sort || 0
    }))
      .sort((a, b) => {
        // 先按sort排序，再按id排序
        if (a.sort !== b.sort) {
          return a.sort - b.sort
        }
        return a.id - b.id
      })
    
    console.log('最终商品列表:', allProducts.value)
  } catch (error) {
    console.error('加载商品失败:', error)
    ElMessage.error('加载商品失败: ' + (error.message || '未知错误'))
    allProducts.value = []
  }
}



// 商品搜索
const handleProductSearch = () => {
  // 搜索逻辑已在computed中实现
}

// 二级分类排序操作
const moveCategoryUp = (index) => {
  if (index === 0) return
  const temp = secondaryCategories.value[index]
  secondaryCategories.value[index] = secondaryCategories.value[index - 1]
  secondaryCategories.value[index - 1] = temp
  // 更新sort值
  updateCategorySortValues()
}

const moveCategoryDown = (index) => {
  if (index === secondaryCategories.value.length - 1) return
  const temp = secondaryCategories.value[index]
  secondaryCategories.value[index] = secondaryCategories.value[index + 1]
  secondaryCategories.value[index + 1] = temp
  // 更新sort值
  updateCategorySortValues()
}

const updateCategorySortValues = () => {
  secondaryCategories.value.forEach((cat, index) => {
    cat.sort = index + 1
  })
}

// 商品排序操作
const moveProductUp = (index) => {
  if (index === 0) return
  const temp = filteredProducts.value[index]
  const actualIndex = allProducts.value.indexOf(temp)
  allProducts.value[actualIndex] = allProducts.value[actualIndex - 1]
  allProducts.value[actualIndex - 1] = temp
  // 更新sort值
  updateProductSortValues()
}

const moveProductDown = (index) => {
  if (index === filteredProducts.value.length - 1) return
  const temp = filteredProducts.value[index]
  const actualIndex = allProducts.value.indexOf(temp)
  allProducts.value[actualIndex] = allProducts.value[actualIndex + 1]
  allProducts.value[actualIndex + 1] = temp
  // 更新sort值
  updateProductSortValues()
}

const updateProductSortValues = () => {
  allProducts.value.forEach((product, index) => {
    product.sort = index + 1
  })
}


// 保存二级分类排序
const saveCategorySort = async () => {
  if (secondaryCategories.value.length === 0) {
    ElMessage.warning('没有需要排序的二级分类')
    return
  }

  categorySaving.value = true
  try {
    const items = secondaryCategories.value.map((cat, index) => ({
      id: cat.id,
      sort: cat.sort || index + 1
    }))

    const response = await batchUpdateCategorySort(items)
    if (response.code === 200) {
      ElMessage.success('二级分类排序保存成功')
      // 重新加载二级分类数据
      if (selectedPrimaryCategoryId.value) {
        await loadSecondaryCategories(selectedPrimaryCategoryId.value)
        // 如果之前选中了二级分类，重新选中
        if (selectedSecondaryCategoryId.value) {
          const row = secondaryCategories.value.find(cat => cat.id === selectedSecondaryCategoryId.value)
          if (row) {
            await handleSecondaryCategorySelect(row)
          }
        }
      }
    } else {
      ElMessage.error(response.message || '保存失败')
    }
  } catch (error) {
    console.error('保存二级分类排序失败:', error)
    ElMessage.error('保存二级分类排序失败')
  } finally {
    categorySaving.value = false
  }
}

// 保存商品排序
const saveProductSort = async () => {
  if (allProducts.value.length === 0) {
    ElMessage.warning('没有需要排序的商品')
    return
  }

  productSaving.value = true
  try {
    const items = allProducts.value.map((product, index) => ({
      id: product.id,
      sort: product.sort || index + 1
    }))

    const response = await batchUpdateProductSort(items)
    if (response.code === 200) {
      ElMessage.success('商品排序保存成功')
      // 重新加载商品数据
      if (selectedSecondaryCategoryId.value) {
        await loadProductsByCategory(selectedSecondaryCategoryId.value)
      }
    } else {
      ElMessage.error(response.message || '保存失败')
    }
  } catch (error) {
    console.error('保存商品排序失败:', error)
    ElMessage.error('保存商品排序失败')
  } finally {
    productSaving.value = false
  }
}

// 获取图片URL
const getImageUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url
  }
  return url
}

// 加载精选商品
const loadSpecialProducts = async () => {
  try {
    const response = await getAllSpecialProducts()
    if (response && response.code === 200) {
      specialProducts.value = (response.data || []).map(product => ({
        ...product,
        special_sort: product.special_sort || 0
      }))
        .sort((a, b) => {
          if (a.special_sort !== b.special_sort) {
            return a.special_sort - b.special_sort
          }
          return a.id - b.id
        })
    }
  } catch (error) {
    console.error('加载精选商品失败:', error)
    ElMessage.error('加载精选商品失败')
  }
}

// 精选商品排序操作
const moveSpecialProductUp = (index) => {
  if (index === 0) return
  const temp = specialProducts.value[index]
  specialProducts.value[index] = specialProducts.value[index - 1]
  specialProducts.value[index - 1] = temp
  updateSpecialProductSortValues()
}

const moveSpecialProductDown = (index) => {
  if (index === specialProducts.value.length - 1) return
  const temp = specialProducts.value[index]
  specialProducts.value[index] = specialProducts.value[index + 1]
  specialProducts.value[index + 1] = temp
  updateSpecialProductSortValues()
}

const updateSpecialProductSortValues = () => {
  specialProducts.value.forEach((product, index) => {
    product.special_sort = index + 1
  })
}

// 保存精选商品排序
const saveSpecialProductSort = async () => {
  if (specialProducts.value.length === 0) {
    ElMessage.warning('没有需要排序的精选商品')
    return
  }

  specialProductSaving.value = true
  try {
    const items = specialProducts.value.map((product, index) => ({
      id: product.id,
      special_sort: product.special_sort || index + 1
    }))

    const response = await batchUpdateSpecialProductSort(items)
    if (response.code === 200) {
      ElMessage.success('精选商品排序保存成功')
      await loadSpecialProducts()
    } else {
      ElMessage.error(response.message || '保存失败')
    }
  } catch (error) {
    console.error('保存精选商品排序失败:', error)
    ElMessage.error('保存精选商品排序失败')
  } finally {
    specialProductSaving.value = false
  }
}

// 页面加载时初始化
onMounted(() => {
  initData()
  loadSpecialProducts()
})
</script>

<style scoped>
.sort-management-container {
  padding: 20px;
}

.page-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #333;
}

.content-wrapper {
  display: flex;
  gap: 20px;
  min-height: 600px;
}

/* 左侧一级分类列表 */
.left-sidebar {
  width: 200px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 10px 0;
  flex-shrink: 0;
}

.sidebar-title {
  padding: 10px 15px;
  font-weight: bold;
  color: #333;
  border-bottom: 1px solid #e4e7ed;
  margin-bottom: 10px;
  font-size: 14px;
}

.category-list {
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

.category-item {
  display: flex;
  align-items: center;
  padding: 12px 15px;
  cursor: pointer;
  transition: all 0.3s;
  border-left: 3px solid transparent;
}

.category-item:hover {
  background-color: #f5f7fa;
}

.category-item.active {
  background-color: #ecf5ff;
  border-left-color: #409eff;
  color: #409eff;
}

.category-icon {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  margin-right: 10px;
  flex-shrink: 0;
}

.category-icon-placeholder {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  margin-right: 10px;
  background-color: #f5f7fa;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  flex-shrink: 0;
}

.category-name-text {
  font-size: 14px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 中间和右侧内容 */
.middle-content,
.right-content {
  flex: 1;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 20px;
  min-width: 0;
}

.category-name {
  color: #409eff;
  font-size: 14px;
  font-weight: normal;
  margin-left: 10px;
}

.sort-section {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.section-header {
  margin-bottom: 15px;
  flex-shrink: 0;
}

.section-header h3 {
  margin: 0;
  font-size: 18px;
  color: #333;
}

.section-actions {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.sort-table {
  flex: 1;
  overflow: auto;
}

.product-info {
  display: flex;
  align-items: center;
}

.product-name {
  font-weight: 500;
  margin-bottom: 5px;
  color: #333;
}

.product-id {
  font-size: 12px;
  color: #999;
}

.product-sort-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.product-sort-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: auto;
}

.product-sort-tabs :deep(.el-tab-pane) {
  height: 100%;
}
</style>
