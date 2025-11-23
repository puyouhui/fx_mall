<template>
  <div class="products-container">
    <el-card>
      <h2 class="page-title">商品管理</h2>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-right">
          <el-button type="primary" @click="handleAddProduct">
            <el-icon>
              <plus />
            </el-icon>
            新增商品
          </el-button>
        </div>

        <div class="toolbar-left">
          <el-input v-model="searchForm.keyword" placeholder="请输入商品名称" :prefix-icon="Search"
            style="width: 300px; margin-right: 20px;" @input="handleSearch" />
          <el-cascader
            v-model="searchForm.categoryIds"
            :options="treeCategories"
            :props="{ checkStrictly: true, label: 'name', value: 'id', children: 'children', emitPath: false }"
            placeholder="选择分类（一级或二级）"
            clearable
            style="width: 250px; margin-right: 20px;"
            @change="handleCategoryFilterChange"
            collapse-tags
            collapse-tags-tooltip
          />
          <el-button type="primary" @click="handleSearch">
            搜索
          </el-button>
        </div>

      </div>

      <!-- 商品列表 -->
      <el-card class="products-card">
        <el-table :data="products" stripe>
          <!-- <el-table-column type="index" label="序号" /> -->
          <el-table-column prop="id" label="商品ID" align="center" width="100" />
          <el-table-column prop="name" label="商品名称" align="center" />
          <el-table-column label="分类" align="center">
            <template #default="scope">
              {{ getCategoryName(scope.row.categoryId) }}
            </template>
          </el-table-column>
          <el-table-column label="供应商" align="center">
            <template #default="scope">
              <span v-if="scope.row.supplier_name">{{ scope.row.supplier_name }}</span>
              <span v-else style="color: #999;">未绑定</span>
            </template>
          </el-table-column>
          <el-table-column label="价格范围" align="center">
            <template #default="scope">
              {{ calculatePriceRange(scope.row.specs) }}
            </template>
          </el-table-column>
          <el-table-column prop="isSpecial" label="精选" align="center" width="100">
            <template #default="scope">
              <el-switch 
                v-model="scope.row.isSpecial" 
                @change="handleSpecialStatusChange(scope.row)"
                :loading="scope.row.updatingSpecial"
              />
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" align="center">
            <template #default="scope">
              <el-tooltip :content="scope.row.description" placement="top">
                <span>{{ truncateText(scope.row.description, 20) }}</span>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column prop="images" label="图片" align="center">
            <template #default="scope">
              <el-image v-if="scope.row.images && scope.row.images.length > 0" :src="getImageUrl(scope.row.images[0])"
                :preview-src-list="getImageUrlList(scope.row.images)" style="width: 40px; height: 40px;" fit="cover" />
              <span v-else>暂无</span>
            </template>
          </el-table-column>
          <el-table-column prop="specs" label="规格" align="center">
            <template #default="scope">
              <el-tooltip :content="formatSpecs(scope.row.specs)" placement="top">
                <span>{{ formatSpecsBrief(scope.row.specs) }}</span>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" align="center">
            <template #default="scope">
              {{ formatDate(scope.row.updated_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" align="center">
            <template #default="scope">
              <el-button type="primary" size="small" @click="handleEditProduct(scope.row)">
                编辑
              </el-button>
              <el-button type="danger" size="small" @click="handleDeleteProduct(scope.row.id)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <div class="pagination-container">
          <el-pagination v-model:current-page="pagination.pageNum" v-model:page-size="pagination.pageSize"
            :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" :total="pagination.total"
            @size-change="handleSizeChange" @current-change="handleCurrentChange" />
        </div>
      </el-card>

      <!-- 新增/编辑商品弹窗 -->
      <el-dialog v-model="dialogVisible" :title="dialogType === 'add' ? '新增商品' : '编辑商品'" width="600px">
        <el-form ref="productFormRef" :model="productForm" :rules="productRules" label-width="100px">
          <el-form-item label="商品名称" prop="name">
            <el-input v-model="productForm.name" placeholder="请输入商品名称" />
          </el-form-item>
          <el-form-item label="所属分类" prop="categoryIds">
            <el-cascader v-model="productForm.categoryIds" :options="treeCategories"
              :props="{ checkStrictly: false, label: 'name', value: 'id', children: 'children' }" placeholder="请选择所属分类"
              style="width: 100%;" />
          </el-form-item>
          <el-form-item label="供应商" prop="supplierId">
            <el-select v-model="productForm.supplierId" placeholder="请选择供应商" style="width: 100%;">
              <el-option v-for="supplier in suppliers" :key="supplier.id" :label="supplier.name" :value="supplier.id" />
            </el-select>
          </el-form-item>
          <!-- 商品本身价格字段已废弃，实际使用规格价格 -->
          <el-form-item label="商品规格">
            <div class="specs-container">
              <div class="specs-input-group">
                <el-row :gutter="12" style="margin-bottom: 10px;">
                  <el-col :span="12">
                    <el-input v-model="currentSpec.name" placeholder="规格名称（如：3瓶装）" />
                  </el-col>
                  <el-col :span="12">
                    <el-input v-model="currentSpec.description" placeholder="规格描述（如：≈1.5元/瓶）" />
                  </el-col>
                </el-row>
                <el-row :gutter="12" style="margin-bottom: 10px;">
                  <el-col :span="12">
                    <el-input-number 
                      v-model="currentSpec.wholesalePrice" 
                      :min="0.01" 
                      :step="0.01" 
                      placeholder="批发价" 
                      style="width: 100%;"
                    />
                  </el-col>
                  <el-col :span="12">
                    <el-input-number 
                      v-model="currentSpec.retailPrice" 
                      :min="0.01" 
                      :step="0.01" 
                      placeholder="零售价" 
                      style="width: 100%;"
                    />
                  </el-col>
                </el-row>
                <div style="text-align: center; margin-top: 10px;">
                  <el-button type="primary" @click="addSpec">添加规格</el-button>
                  <el-tooltip effect="dark" content="请填写所有必填项后添加规格" placement="top">
                    <el-button type="text" size="small" style="margin-left: 10px;">添加说明</el-button>
                  </el-tooltip>
                </div>
              </div>
              <div class="specs-list" v-if="productForm.specs.length > 0">
                <div v-for="(spec, index) in productForm.specs" :key="index" class="spec-item">
                  <span>{{ spec.name }} ({{ spec.description || '-' }}) (批发价: ¥{{ spec.wholesale_price || spec.wholesalePrice }}, 零售价: ¥{{ spec.retail_price || spec.retailPrice }})</span>
                  <el-button type="text" danger @click="removeSpec(index)">删除</el-button>
                </div>
              </div>
            </div>
          </el-form-item>
          <el-form-item label="是否精选" prop="isSpecial">
            <el-switch v-model="productForm.isSpecial" />
          </el-form-item>
          <el-form-item label="商品描述" prop="description">
            <el-input v-model="productForm.description" type="textarea" placeholder="请输入商品描述" rows="3" />
          </el-form-item>
          <!-- 商品规格已移到上方价格输入位置 -->
          <el-form-item label="商品图片">
            <el-upload class="upload-demo" action="" :show-file-list="true" :on-remove="handleRemove"
              :before-upload="beforeUpload" :http-request="handleHttpRequest" multiple limit="5"
              :on-exceed="handleExceed" :file-list="productForm.images" list-type="picture-card"
              :on-preview="handleUploadPreview">
              <el-button type="primary">
                <el-icon>
                  <upload />
                </el-icon>
                上传图片
              </el-button>
            </el-upload>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElTooltip, ElImage } from 'element-plus'
import {
  Search,
  Plus,
  Upload
} from '@element-plus/icons-vue'
// 导入相关API函数
import { getProductList, createProduct, updateProduct, deleteProduct, uploadProductImage, updateProductSpecialStatus } from '../api/product'
import { getCategoryList } from '../api/category'
import { getAllSuppliers } from '../api/suppliers'
import { formatDate } from '../utils/time-format'

// 截断文本函数
const truncateText = (text, length) => {
  if (!text) return ''
  return text.length > length ? text.substring(0, length) + '...' : text
}

// 格式化规格显示
const formatSpecs = (specs) => {
  if (!specs || !Array.isArray(specs)) return '暂无规格'
  return specs.map(spec => `${spec.name}`).join('\n')
}

// 格式化规格简要显示
const formatSpecsBrief = (specs) => {
  if (!specs || !Array.isArray(specs)) return '暂无规格'
  if (specs.length === 0) return '暂无规格'
  return truncateText(specs[0].name, 15) + (specs.length > 1 ? ` +${specs.length - 1}` : '')
}

// 计算价格范围（前端计算，避免后端资源浪费）
// 显示所有规格的批发价和零售价中的最低价到最高价
const calculatePriceRange = (specs) => {
  if (!specs || !Array.isArray(specs) || specs.length === 0) {
    return '暂无价格'
  }

  // 收集所有规格的批发价和零售价
  const allPrices = []
  
  specs.forEach(spec => {
    // 获取批发价
    const wholesalePrice = spec.wholesale_price || spec.wholesalePrice
    if (wholesalePrice && wholesalePrice > 0) {
      allPrices.push(wholesalePrice)
    }
    
    // 获取零售价
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

// 商品列表
const products = ref([])

// 分类列表
const categories = ref([])

// 树形分类列表（用于级联选择器）
const treeCategories = ref([])

// 供应商列表
const suppliers = ref([])

// 搜索表单
const searchForm = reactive({
  keyword: '',
  categoryId: '', // 保留用于API调用
  categoryIds: null // 级联选择器的值
})

// 分页信息
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 100
})

// 弹窗相关
const dialogVisible = ref(false)
const dialogType = ref('add')
const productFormRef = ref(null)
const productForm = reactive({
  id: null,
  name: '',
  categoryIds: [], // 改为数组存储级联选择的分类ID
  categoryId: '', // 保留原字段用于提交
  supplierId: null, // 供应商ID
  originalPrice: 0,
  price: 0,
  isSpecial: false,
  description: '',
  images: [],
  specs: []
})

// 规格相关
const currentSpec = reactive({
  name: '',
  wholesalePrice: null,
  retailPrice: null,
  description: ''
})

// 表单验证规则
const productRules = {
  name: [
    { required: true, message: '请输入商品名称', trigger: 'blur' },
    { min: 2, max: 50, message: '商品名称长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  categoryIds: [
    {
      required: true,
      validator: (rule, value, callback) => {
        if (!value || value.length === 0) {
          callback(new Error('请选择所属分类'))
          return
        }

        // 如果是二级分类，直接通过
        if (value.length === 2) {
          callback()
          return
        }

        // 检查当前一级分类是否有子分类
        const firstLevelId = value[0]
        const category = categories.value.find(c => c.id === firstLevelId)
        if (category) {
          const hasChildren = categories.value.some(c => c.parent_id === firstLevelId)
          if (hasChildren) {
            callback(new Error('当前分类下有子分类，请选择二级分类'))
          } else {
            callback()
          }
        } else {
          callback(new Error('分类不存在'))
        }
      },
      trigger: 'change'
    }
  ],
  description: [
    { max: 500, message: '商品描述不能超过 500 个字符', trigger: 'blur' }
  ],
  supplierId: [
    // 供应商ID为可选，后端会自动分配默认供应商
    // 如果用户未选择供应商，后端会使用默认的自营供应商
  ]
}

// 根据分类ID获取分类名称
const getCategoryName = (categoryId) => {
  if (!categoryId) return '未分类'
  const category = categories.value.find(cat => Number(cat.id) === Number(categoryId))
  return category ? category.name : '未分类'
}

// 将扁平分类数据转换为树形结构
defineProps({
  flatCategories: {
    type: Array,
    default: () => []
  }
})

const convertToTree = (flatCategories) => {
  const map = new Map()
  const roots = []

  // 首先将所有分类添加到映射中
  flatCategories.forEach(category => {
    map.set(category.id, {
      ...category,
      label: category.name,
      value: category.id,
      children: []
    })
  })

  // 构建树形结构
  flatCategories.forEach(category => {
    if (category.parent_id === 0 || !category.parent_id) {
      // 一级分类
      roots.push(map.get(category.id))
    } else {
      // 二级分类，添加到父分类的children中
      const parent = map.get(category.parent_id)
      if (parent) {
        parent.children.push(map.get(category.id))
      }
    }
  })

  return roots
}

// 初始化数据
const initData = async () => {
  try {
    // 加载分类数据
    const categoryResponse = await getCategoryList()
    // 打印原始分类数据用于调试
    console.log('原始分类数据:', categoryResponse)
    // 处理分类数据结构 - 支持嵌套data字段和直接数据
    let rawCategories = []
    if (categoryResponse && categoryResponse.data && Array.isArray(categoryResponse.data)) {
      rawCategories = categoryResponse.data
    } else if (Array.isArray(categoryResponse)) {
      rawCategories = categoryResponse
    }

    // 检测数据是否已经是树形结构（包含children字段）
    const isTreeStructure = rawCategories.length > 0 && rawCategories[0].children !== undefined

    if (isTreeStructure) {
      // 如果已经是树形结构，直接使用，并转换为级联选择器需要的格式
      treeCategories.value = rawCategories.map(category => ({
        ...category,
        label: category.name,
        value: category.id,
        children: category.children.map(child => ({
          ...child,
          label: child.name,
          value: child.id
        }))
      }))

      // 提取所有分类到扁平数组用于搜索
      const flattenCategories = (categories) => {
        let result = []
        categories.forEach(category => {
          result.push(category)
          if (category.children && category.children.length > 0) {
            result = result.concat(flattenCategories(category.children))
          }
        })
        return result
      }
      categories.value = flattenCategories(rawCategories)
    } else {
      // 如果是扁平结构，按原来的方式处理
      categories.value = rawCategories
      treeCategories.value = convertToTree(rawCategories)
    }

    console.log('树形分类数据:', treeCategories.value);

    // 加载供应商数据
    try {
      const supplierResponse = await getAllSuppliers()
      if (supplierResponse.code === 200 && supplierResponse.data) {
        suppliers.value = supplierResponse.data
      }
    } catch (error) {
      console.error('加载供应商数据失败:', error)
    }

    // 加载商品数据
    const productResponse = await getProductList({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword,
      categoryId: searchForm.categoryId || ''
    })
    // 处理商品数据结构 - 支持嵌套data字段和直接数据
    let productData = []
    if (productResponse && productResponse.data && Array.isArray(productResponse.data)) {
      productData = productResponse.data
    } else if (Array.isArray(productResponse)) {
      productData = productResponse
    }

    // 更新分页总数（使用后端返回的total，如果没有则使用数组长度）
    if (productResponse && productResponse.total !== undefined) {
      pagination.total = productResponse.total
    } else {
      pagination.total = productData.length
    }

    if (productData.length > 0) {
      // 确保数据格式正确
      products.value = productData.map(product => ({
        ...product,
        id: Number(product.id),
        categoryId: Number(product.categoryId || product.category_id || 0),
        categoryName: getCategoryName(Number(product.categoryId || product.category_id || 0)), // 添加分类名称
        supplierId: product.supplier_id || product.supplierId || null, // 供应商ID
        supplier_name: product.supplier_name || '', // 供应商名称
        originalPrice: Number(product.originalPrice || product.original_price || 0),
        price: Number(product.price || 0),
        isSpecial: product.isSpecial === true || product.isSpecial === 'true' || product.is_special === true || product.is_special === 'true',
        updatingSpecial: false, // 添加更新状态标记
        images: Array.isArray(product.images) ? product.images : [],
        specs: Array.isArray(product.specs) ? product.specs : []
      }))
    } else {
      products.value = []
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败，请稍后再试')
  }
}

// 搜索商品
// 分类筛选变更处理（选中后立即筛选）
const handleCategoryFilterChange = (value) => {
  // 级联选择器设置了 emitPath: false，所以返回的是单个分类ID
  if (value && value !== null && value !== '') {
    searchForm.categoryId = value
  } else {
    searchForm.categoryId = ''
  }
  // 重置到第一页并立即筛选
  pagination.pageNum = 1
  initData()
}

const handleSearch = () => {
  pagination.pageNum = 1
  initData()
}

// 分页处理
const handleSizeChange = (size) => {
  pagination.pageSize = size
  initData()
}

const handleCurrentChange = (current) => {
  pagination.pageNum = current
  initData()
}

// 打开新增商品弹窗
const handleAddProduct = () => {
  dialogType.value = 'add'
  // 重置表单
  if (productFormRef.value) {
    productFormRef.value.resetFields()
  }
  // 清空表单数据，尝试默认选择自营供应商
  // 如果供应商列表还未加载完成，supplierId 将为 null，后端会自动分配默认供应商
  const selfOperatedSupplier = suppliers.value.find(s => s.username === 'self_operated')
  Object.assign(productForm, {
    id: null,
    name: '',
    categoryIds: [],
    categoryId: '',
    supplierId: selfOperatedSupplier ? selfOperatedSupplier.id : null, // 如果找到自营供应商则使用，否则为 null（后端会自动分配）
    originalPrice: 0, // 保留字段但不再使用
    price: 0, // 保留字段但不再使用
    isSpecial: false,
    description: '',
    images: [],
    specs: []
  })
  // 清空当前规格
  Object.assign(currentSpec, {
    name: '',
    value: '',
    price: 0,
    originalPrice: 0
  })
  dialogVisible.value = true
}

// 获取分类的完整路径
const getCategoryPath = (categoryId) => {
  const path = []
  let currentId = categoryId

  // 先找到当前分类
  while (currentId) {
    const category = categories.value.find(cat => Number(cat.id) === Number(currentId))
    if (!category) break

    // 插入到路径的开头
    path.unshift(Number(currentId))

    // 如果没有父分类或父分类是0，则退出循环
    if (!category.parent_id || category.parent_id === 0) {
      break
    }

    // 继续查找父分类
    currentId = category.parent_id
  }

  return path
}

// 打开编辑商品弹窗
const handleEditProduct = (row) => {
  dialogType.value = 'edit'
  // 复制行数据到表单
  const categoryId = Number(row.categoryId || 0) // 确保分类ID是数字类型
  const categoryPath = getCategoryPath(categoryId)

  // 如果没有供应商，默认选择自营供应商
  let supplierId = row.supplier_id || row.supplierId || null
  if (!supplierId) {
    const selfOperatedSupplier = suppliers.value.find(s => s.username === 'self_operated')
    if (selfOperatedSupplier) {
      supplierId = selfOperatedSupplier.id
    }
  }

  Object.assign(productForm, {
    ...row,
    id: Number(row.id), // 确保id是数字类型
    categoryId: categoryId,
    categoryIds: categoryPath, // 设置级联分类路径
    supplierId: supplierId, // 设置供应商ID
    images: row.images && Array.isArray(row.images) ?
      row.images.map(img => {
        // 确保img是有效类型
        if (!img) {
          return { name: 'image', url: '' }
        }
        // 根据img的类型处理
        if (typeof img === 'string') {
          try {
            return {
              name: img.substring(img.lastIndexOf('/') + 1),
              url: img
            }
          } catch (e) {
            // 防止字符串操作出错
            return { name: 'image', url: img }
          }
        } else if (typeof img === 'object') {
          return {
            name: img.name || (img.url ? (typeof img.url === 'string' ? img.url.substring(img.url.lastIndexOf('/') + 1) : 'image') : 'image'),
            url: img.url || ''
          }
        }
        // 默认返回空图片对象
        return { name: 'image', url: '' }
      }) : [],
    specs: Array.isArray(row.specs) ? [...row.specs] : []
  })
  // 清空当前规格
  Object.assign(currentSpec, {
    name: '',
    wholesalePrice: null,
    retailPrice: null,
    description: ''
  })
  dialogVisible.value = true
}

// 处理精选状态变更
const handleSpecialStatusChange = async (row) => {
  try {
    // 设置更新状态，防止重复点击
    if (!row.updatingSpecial) {
      row.updatingSpecial = true
    }
    
    const response = await updateProductSpecialStatus(row.id, row.isSpecial)
    
    if (response.code === 200) {
      ElMessage.success(row.isSpecial ? '已设置为精选商品' : '已取消精选')
    } else {
      // 如果更新失败，恢复原状态
      row.isSpecial = !row.isSpecial
      ElMessage.error(response.message || '更新失败')
    }
  } catch (error) {
    // 如果更新失败，恢复原状态
    row.isSpecial = !row.isSpecial
    console.error('更新精选状态失败:', error)
    if (error.response && error.response.data && error.response.data.message) {
      ElMessage.error(error.response.data.message)
    } else {
      ElMessage.error('更新精选状态失败')
    }
  } finally {
    row.updatingSpecial = false
  }
}

// 删除商品
const handleDeleteProduct = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个商品吗？', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    // 调用删除API
    await deleteProduct(id)

    // 从列表中移除
    const index = products.value.findIndex(item => item.id === id)
    if (index > -1) {
      products.value.splice(index, 1)
    }

    ElMessage.success('删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 添加规格
const addSpec = () => {
  if (!currentSpec.name || !currentSpec.wholesalePrice || currentSpec.wholesalePrice <= 0 || !currentSpec.retailPrice || currentSpec.retailPrice <= 0) {
    ElMessage.warning('请输入规格名称、批发价和零售价')
    return
  }

  // 检查是否已存在相同的规格
  const exists = productForm.specs.some(spec =>
    spec.name === currentSpec.name
  )

  if (exists) {
    ElMessage.warning('该规格已存在')
    return
  }

  // 添加规格，使用正确的字段名
  productForm.specs.push({
    name: currentSpec.name,
    wholesale_price: currentSpec.wholesalePrice,
    retail_price: currentSpec.retailPrice,
    description: currentSpec.description || ''
  })

  // 清空当前输入
  Object.assign(currentSpec, {
    name: '',
    wholesalePrice: null,
    retailPrice: null,
    description: ''
  })
}

// 移除规格
const removeSpec = (index) => {
  productForm.specs.splice(index, 1)
}

// 提交表单
const handleSubmit = async () => {
  try {
    // 验证表单
    await productFormRef.value.validate()

    // 验证规格数据
    if (!productForm.specs || productForm.specs.length === 0) {
      ElMessage.warning('请至少添加一个商品规格')
      return
    }

    for (const spec of productForm.specs) {
      const wholesalePrice = spec.wholesale_price || spec.wholesalePrice
      const retailPrice = spec.retail_price || spec.retailPrice
      if (!spec.name || !wholesalePrice || wholesalePrice <= 0 || !retailPrice || retailPrice <= 0) {
        ElMessage.warning('所有规格都必须有名称、批发价和零售价')
        return
      }
    }

    // 根据级联选择的分类ID设置最终提交的categoryId
    // 如果选择了二级分类，使用二级分类ID；否则使用一级分类ID
    const categoryId = productForm.categoryIds.length > 1
      ? productForm.categoryIds[1]
      : productForm.categoryIds[0]

    // 处理图片数据和规格数据
    const formData = {
      // 只包含需要提交的字段，避免多余字段导致的问题
      id: productForm.id, // 添加id字段
      name: productForm.name,
      category_id: Number(categoryId), // 转换为数字类型
      supplier_id: productForm.supplierId || null, // 供应商ID（可选）
      original_price: 0, // 设置为0，实际使用规格价格
      price: 0, // 设置为0，实际使用规格价格
      is_special: productForm.isSpecial, // 转换为后端期望的字段名
      description: productForm.description,
      images: productForm.images.map(img => img.url),
      specs: productForm.specs || [], // 确保规格数据存在且格式正确
      status: 1 // 默认状态为启用
    }

    // 确保id是数字类型
    if (formData.id) {
      formData.id = Number(formData.id)
    }

    if (dialogType.value === 'add') {
      // 调用新增API
      const response = await createProduct(formData)

      // 重新加载商品列表
      initData()

      ElMessage.success('新增成功')
    } else {
      // 调用更新API
      const response = await updateProduct(formData.id, formData)

      // 重新加载商品列表
      initData()

      ElMessage.success('更新成功')
    }

    dialogVisible.value = false
  } catch (error) {
    console.error('提交失败:', error)
    ElMessage.error('提交失败，请稍后再试')
  }
}

// 获取单个图片URL的辅助函数
const getImageUrl = (img) => {
  if (!img) return ''
  if (typeof img === 'string') return img
  if (typeof img === 'object' && img.url) return img.url
  return ''
}

// 获取图片URL列表的辅助函数
const getImageUrlList = (images) => {
  if (!images || !Array.isArray(images)) return []
  return images.map(img => getImageUrl(img)).filter(url => url !== '')
}

// 上传组件的预览处理函数
const handleUploadPreview = (uploadFile) => {
  // 创建一个临时的图片预览弹窗
  const imageUrl = getImageUrl(uploadFile)
  if (imageUrl) {
    // 使用Element Plus的image-viewer组件显示图片预览
    const viewer = document.createElement('div')
    viewer.className = 'image-viewer-overlay'
    viewer.innerHTML = `
      <div class="image-viewer-content" style="display: flex; justify-content: center; align-items: center; height: 100%;">
        <img src="${imageUrl}" style="max-width: 90%; max-height: 90%;" />
      </div>
    `
    viewer.onclick = () => {
      document.body.removeChild(viewer)
    }
    // 添加样式
    viewer.style.position = 'fixed'
    viewer.style.top = '0'
    viewer.style.left = '0'
    viewer.style.width = '100%'
    viewer.style.height = '100%'
    viewer.style.backgroundColor = 'rgba(0, 0, 0, 0.8)'
    viewer.style.zIndex = '9999'
    viewer.style.cursor = 'pointer'
    viewer.style.display = 'flex'
    viewer.style.justifyContent = 'center'
    viewer.style.alignItems = 'center'
    document.body.appendChild(viewer)
  }
}

// 上传相关函数
const handleRemove = (file, fileList) => {
  productForm.images = fileList
}

const handlePreview = (uploadFile) => {
  console.log('预览文件:', uploadFile)
}

const handleExceed = (files, fileList) => {
  ElMessage.warning(`最多只能上传 ${fileList.length} 个文件`)
}

// 文件类型和大小校验
const beforeUpload = (file) => {
  // 文件类型校验
  const isImage = /\.(jpg|jpeg|png|gif)$/i.test(file.name)
  if (!isImage) {
    ElMessage.error('请上传JPG、PNG或GIF格式的图片')
    return false
  }

  // 文件大小校验
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过5MB')
    return false
  }

  return true
}

// 自定义上传函数
const handleHttpRequest = async (options) => {
  try {
    const { file } = options

    // 显示上传中状态
    ElMessage({ message: '图片上传中...', type: 'info' })

    // 创建FormData对象
    const formData = new FormData()
    formData.append('file', file)

    // 调用上传API
    const response = await uploadProductImage(formData)

    // 处理上传结果
    if (response.code === 200 && response.data && response.data.imageUrl) {
      // 成功回调
      if (options.onSuccess) {
        options.onSuccess(response)
      }

      // 将图片添加到列表中
      const uploadedFile = {
        name: file.name,
        url: response.data.imageUrl,
        status: 'success' // 标记为成功状态
      }

      // 确保不会重复添加
      const exists = productForm.images.some(img => img.url === uploadedFile.url)
      if (!exists) {
        productForm.images.push(uploadedFile)
      }

      ElMessage.success('图片上传成功')
    } else {
      // 失败回调
      if (options.onError) {
        options.onError(response)
      }
      ElMessage.error('图片上传失败: ' + (response.message || '未知错误'))
    }
  } catch (error) {
    // 错误回调
    if (options.onError) {
      options.onError(error)
    }
    console.error('图片上传失败:', error)
    ElMessage.error('图片上传失败，请稍后再试')
  }
}

// 组件挂载时
onMounted(() => {
  initData()
})
</script>

<style scoped>
.products-container {
  padding: 0 0 20px 0;
}

.page-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #333;
}

.toolbar-card {
  margin-bottom: 10px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  margin-bottom: 20px;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-right {
  display: flex;
  align-items: center;
}

.products-card {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

/* 规格相关样式 */
.specs-container {
  width: 100%;
}

.specs-input-group {
  margin-bottom: 15px;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
}

.specs-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 10px;
}

.spec-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  margin-bottom: 8px;
  background-color: #f9f9f9;
  border-radius: 4px;
}

.spec-item:last-child {
  margin-bottom: 0;
}

.spec-item span {
  flex: 1;
  font-size: 14px;
  line-height: 1.5;
}

/* 价格输入框样式优化 */
.el-input-number {
  .el-input__wrapper {
    padding: 0 32px 0 12px;
  }
}

/* 响应式布局 */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-left,
  .toolbar-right {
    margin-bottom: 10px;
    justify-content: center;
  }

  .toolbar-left {
    flex-direction: column;
  }

  .toolbar-left>* {
    width: 100%;
    margin-right: 0 !important;
    margin-bottom: 10px;
  }
}
</style>