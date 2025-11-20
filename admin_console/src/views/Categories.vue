<template>
  <div class="categories-container">
    <el-card>
      <h2 class="page-title">分类管理</h2>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" @click="handleAddCategory">
            <el-icon>
              <plus />
            </el-icon>
            新增分类
          </el-button>
        </div>
        <div class="toolbar-right">
          <!-- <el-input v-model="searchName" placeholder="请输入分类名称" clearable /> -->
        </div>
      </div>

      <!-- 分类列表 -->
      <el-card class="categories-card">
        <el-table :data="categoriesData" stripe>
          <el-table-column type="index" label="id" align="center">
            <template #default="scope">
              {{ scope.row.id }}
            </template>
          </el-table-column>
          <el-table-column prop="name" label="分类名称" align="center">
            <template #default="scope">
              <div v-if="scope.row.parent_id === 0" class="category-level-1">{{ scope.row.name }}</div>
              <div v-else class="category-level-2">{{ '└─ ' + scope.row.name }}</div>
            </template>
          </el-table-column>

          <el-table-column prop="icon" label="分类图标" align="center">
            <template #default="scope">
              <img v-if="scope.row.icon" :src="scope.row.icon" alt="分类图标" style="width: 30px; height: 30px;">
              <span v-else>未上传</span>
            </template>
          </el-table-column>
          <el-table-column prop="parentName" label="所属分类" align="center">
            <template #default="scope">
              <span v-if="scope.row.parent_id === 0">无</span>
              <span v-else>{{ scope.row.parentName }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" align="center">
            <template #default="scope">
              <el-switch v-model="scope.row.status" @change="handleStatusChange(scope.row)" />
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" align="center">
            <template #default="scope">
              {{ formatDate(scope.row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" align="center" fixed="right">
            <template #default="scope">
              <el-button type="primary" size="small" @click="handleEditCategory(scope.row)">
                编辑
              </el-button>
              <el-button type="danger" size="small" @click="handleDeleteCategory(scope.row.id)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-card>

    <!-- 新增/编辑分类弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogType === 'add' ? '新增分类' : '编辑分类'" width="400px">
      <el-form ref="categoryFormRef" :model="categoryForm" :rules="categoryRules" label-width="100px">
        <el-form-item label="分类名称" prop="name">
          <el-input v-model="categoryForm.name" placeholder="请输入分类名称" />
        </el-form-item>
        <el-form-item label="父分类" prop="parent_id">
          <el-select v-model="categoryForm.parent_id" placeholder="请选择父分类（不选为一级分类）">
            <el-option label="无" :value="0" />
            <el-option v-for="cat in level1Categories" :key="cat.id" :label="cat.name" :value="cat.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="categoryForm.status" />
        </el-form-item>

        <el-form-item label="分类图标">
          <el-upload class="avatar-uploader" :show-file-list="false" :before-upload="beforeIconUpload" accept="image/*"
            :http-request="handleIconHttpRequest">
            <img v-if="categoryForm.icon" :src="categoryForm.icon" class="avatar" />
            <el-icon v-else class="avatar-uploader-icon">
              <Plus />
            </el-icon>
          </el-upload>
          <div v-if="categoryForm.icon" class="icon-upload-tip">
            <el-button type="text" size="small" @click="removeIcon">移除图标</el-button>
          </div>
          <div class="icon-upload-tip">支持jpg、jpeg、png格式，建议尺寸: 80x80px</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getCategoryList, createCategory, updateCategory, deleteCategory, uploadCategoryImage } from '../api/category'
import { formatDate } from '../utils/time-format'

// 原始分类数据（从后端获取的树形结构）
const originalCategories = ref([])

// 扁平处理后的分类数据（用于表格显示）
const categoriesData = ref([])

// 一级分类列表（用于选择父分类）
const level1Categories = ref([])

// 弹窗相关
const dialogVisible = ref(false)
const dialogType = ref('add')
const categoryFormRef = ref(null)
const categoryForm = reactive({
  id: '',
  name: '',
  parent_id: 0,
  status: true,
  icon: ''
})

// 表单验证规则
const categoryRules = {
  name: [
    { required: true, message: '请输入分类名称', trigger: 'blur' },
    { min: 2, max: 20, message: '分类名称长度在 2 到 20 个字符', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        // 检查名称是否重复
        if (value) {
          // 找到所有分类的扁平列表
          const allCategories = categoriesData.value
          const isDuplicate = allCategories.some(item =>
            item.name === value && item.parent_id === categoryForm.parent_id && item.id !== categoryForm.id
          )
          if (isDuplicate) {
            callback(new Error('该父分类下已存在同名分类'))
            return
          }
        }
        callback()
      },
      trigger: 'blur'
    }
  ]
}

// 初始化数据
const initData = async () => {
  try {
    // 加载分类数据
    const response = await getCategoryList()

    if (response.code === 200 && response.data) {
      originalCategories.value = response.data

      // 处理扁平数据用于表格显示
      const flattened = []
      const categoryMap = new Map()

      // 先构建所有分类的映射表
      const buildCategoryMap = (categories) => {
        categories.forEach(category => {
          categoryMap.set(category.id, category)
          if (category.children && category.children.length > 0) {
            buildCategoryMap(category.children)
          }
        })
      }

      // 先构建映射表
      buildCategoryMap(originalCategories.value)

      // 扁平化分类数据
      const flattenCategories = (categories) => {
        categories.forEach(category => {
          // 为每个分类添加父分类名称
          const item = { ...category }
          // 显式将status从整数转换为布尔值
          item.status = item.status === 1

          if (item.parent_id && item.parent_id !== 0) {
            const parent = categoryMap.get(item.parent_id)
            if (parent) {
              item.parentName = parent.name
            }
          } else {
            item.parentName = '无'
          }

          // 添加到扁平列表
          flattened.push(item)

          // 递归处理子分类
          if (item.children && item.children.length > 0) {
            flattenCategories(item.children)
          }
        })
      }

      // 执行扁平化
      flattenCategories(originalCategories.value)

      // 设置扁平数据
      categoriesData.value = flattened

      // 设置一级分类列表（用于选择父分类）
      level1Categories.value = originalCategories.value.filter(cat => cat.parent_id === 0)
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败')
  }
}

// 打开新增分类弹窗
const handleAddCategory = () => {
  dialogType.value = 'add'
  // 重置表单
  if (categoryFormRef.value) {
    categoryFormRef.value.resetFields()
  }
  // 清空表单数据 - 创建分类时不设置id（由后端自动生成）
  Object.assign(categoryForm, {
    name: '',
    parent_id: 0,
    status: true,
    icon: ''
  })
  dialogVisible.value = true
}

// 打开编辑分类弹窗
const handleEditCategory = (row) => {
  dialogType.value = 'edit'
  // 复制行数据到表单
  Object.assign(categoryForm, {
    id: row.id,
    name: row.name,
    parent_id: row.parent_id || 0,
    status: row.status !== undefined ? row.status : true,
    icon: row.icon || ''
  })
  dialogVisible.value = true
}

// 删除分类
const handleDeleteCategory = async (id) => {
  try {
    // 检查是否有二级分类
    const hasChildren = categoriesData.value.some(item => item.parent_id === id)
    if (hasChildren) {
      ElMessage.warning('该分类下有子分类，不能删除')
      return
    }

    await ElMessageBox.confirm('确定要删除这个分类吗？', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    // 调用后端API删除分类
    const response = await deleteCategory(id)

    if (response.code === 200) {
      // 重新加载数据
      await initData()
      ElMessage.success('删除成功')
    } else {
      ElMessage.error(response.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    // 验证表单
    await categoryFormRef.value.validate()

    let response
    if (dialogType.value === 'add') {
      // 调用后端API创建分类
      response = await createCategory(categoryForm)
    } else {
      // 调用后端API更新分类
      response = await updateCategory(categoryForm.id, categoryForm)
    }

    if (response.code === 200) {
      // 重新加载数据
      await initData()
      ElMessage.success(dialogType.value === 'add' ? '新增成功' : '更新成功')
      dialogVisible.value = false
    } else {
      ElMessage.error(response.message || '操作失败')
    }
  } catch (error) {
    console.error('提交失败:', error)
    ElMessage.error('操作失败')
  }
}

// 更新状态
const handleStatusChange = async (row) => {
  try {
    // 调用后端API更新状态，传递完整的分类数据以避免其他字段被覆盖
    const response = await updateCategory(row.id, {
      name: row.name,
      parent_id: row.parent_id,
      sort: row.sort || 0,
      status: row.status,
      icon: row.icon || ''
    })

    if (response.code === 200) {
      ElMessage.success('状态已更新')
    } else {
      ElMessage.error(response.message || '更新状态失败')
      // 回滚状态
      row.status = !row.status
    }
  } catch (error) {
    console.error('更新状态失败:', error)
    ElMessage.error('更新状态失败')
    // 回滚状态
    row.status = !row.status
  }
}

// 图标上传处理
const handleIconHttpRequest = async (options) => {
  try {
    const { file } = options

    // 直接传递file对象给uploadCategoryImage函数，它内部会创建FormData
    const response = await uploadCategoryImage(file)

    if (response.code === 200) {
      categoryForm.icon = response.data.url
      ElMessage.success('图标上传成功')
      // 调用success回调
      if (options.onSuccess) {
        options.onSuccess(response)
      }
    } else {
      ElMessage.error(response.message || '图标上传失败')
      if (options.onError) {
        options.onError(response)
      }
    }
  } catch (error) {
    console.error('图标上传失败:', error)
    ElMessage.error('图标上传失败，请稍后再试')
    if (options.onError) {
      options.onError(error)
    }
  }
}

// 上传前验证
const beforeIconUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  if (!isImage) {
    ElMessage.error('只能上传图片格式的文件')
    return false
  }
  const isLt1M = file.size / 1024 / 1024 < 1
  if (!isLt1M) {
    ElMessage.error('上传图片大小不能超过1MB')
    return false
  }
  return true
}

// 移除图标
const removeIcon = () => {
  categoryForm.icon = ''
}

// 组件挂载时
onMounted(() => {
  initData()
})
</script>

<style scoped>
.categories-container {
  padding: 0 0 20px 0;
}

.page-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #333;
}

.toolbar-card {
  margin-bottom: 20px;
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

.categories-card {
  margin-bottom: 20px;
}

.category-level-1 {
  font-weight: 500;
}

.category-level-2 {
  color: #666;
  padding-left: 20px;
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

/* 图标上传组件样式 */
.avatar-uploader {
  display: flex;
  align-items: center;
}

.avatar {
  width: 80px;
  height: 80px;
  border-radius: 4px;
  object-fit: cover;
}

.avatar-uploader-icon {
  width: 80px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed #d9d9d9;
  border-radius: 4px;
  font-size: 28px;
  color: #8c8c8c;
  background-color: #f5f5f5;
  cursor: pointer;
  transition: all 0.3s;
}

.avatar-uploader-icon:hover {
  border-color: #409eff;
  color: #409eff;
  background-color: #e6f7ff;
}

.icon-upload-tip {
  margin-top: 8px;
  margin-left: 10px;
  color: #8c8c8c;
  font-size: 12px;
}
</style>