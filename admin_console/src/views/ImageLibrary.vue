<template>
  <div class="image-library-container">
    <el-card>
      <h2 class="page-title">图库管理</h2>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-select
            v-model="selectedCategory"
            placeholder="选择目录"
            style="width: 150px; margin-right: 20px;"
            @change="handleCategoryChange"
            clearable
          >
            <el-option label="全部" value="" />
            <el-option-group label="商品相关">
              <el-option label="商品图片" value="products" />
              <el-option label="轮播图" value="carousels" />
              <el-option label="分类图标" value="categories" />
            </el-option-group>
            <el-option-group label="用户和系统">
              <el-option label="用户相关" value="users" />
              <el-option label="配送相关" value="delivery" />
            </el-option-group>
            <el-option-group label="其他">
              <el-option label="其他图片" value="others" />
              <el-option label="富文本图片" value="rich-content" />
            </el-option-group>
          </el-select>
          <el-input
            v-model="searchKeyword"
            placeholder="搜索图片名称"
            :prefix-icon="Search"
            style="width: 300px; margin-right: 20px;"
            @input="handleSearch"
            clearable
          />
          <el-button type="primary" @click="handleRefresh">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-select
            v-model="uploadCategory"
            placeholder="选择上传目录"
            style="width: 150px; margin-right: 10px;"
          >
            <el-option-group label="商品相关">
              <el-option label="商品图片" value="products" />
              <el-option label="轮播图" value="carousels" />
              <el-option label="分类图标" value="categories" />
            </el-option-group>
            <el-option-group label="用户和系统">
              <el-option label="用户相关" value="users" />
              <el-option label="配送相关" value="delivery" />
            </el-option-group>
            <el-option-group label="其他">
              <el-option label="其他图片" value="others" />
              <el-option label="富文本图片" value="rich-content" />
            </el-option-group>
          </el-select>
          <el-upload
            ref="uploadRef"
            class="upload-btn"
            action=""
            :show-file-list="false"
            :before-upload="beforeUpload"
            :on-change="handleFileChange"
            :auto-upload="false"
            multiple
          >
            <el-button type="primary">
              <el-icon><Upload /></el-icon>
              上传图片
            </el-button>
          </el-upload>
          <el-button
            type="danger"
            :disabled="selectedImages.length === 0"
            @click="handleBatchDelete"
          >
            <el-icon><Delete /></el-icon>
            批量删除 ({{ selectedImages.length }})
          </el-button>
        </div>
      </div>

      <!-- 图片列表 -->
      <div class="image-grid" v-loading="loading">
        <div
          v-for="image in images"
          :key="image.url"
          class="image-item"
          :class="{ selected: selectedImages.includes(image.url) }"
          @click="toggleSelect(image.url)"
        >
          <div class="image-wrapper">
            <!-- 优化：使用懒加载和占位符 -->
            <img 
              :data-src="image.url" 
              :alt="image.name" 
              class="image-preview lazy-image"
              loading="lazy"
              @load="handleImageLoad"
              @error="handleImageError"
            />
            <!-- 加载占位符 -->
            <div class="image-placeholder" v-if="!imageLoaded[image.url] || imageLoaded[image.url] === 'error'">
              <el-icon class="loading-icon" v-if="imageLoaded[image.url] !== 'error'"><Loading /></el-icon>
            </div>
            <div class="image-overlay">
              <el-checkbox
                :model-value="selectedImages.includes(image.url)"
                @change="toggleSelect(image.url)"
                @click.stop
              />
              <div class="image-actions">
                <el-button
                  type="primary"
                  size="small"
                  circle
                  @click.stop="handlePreview(image)"
                  title="预览"
                >
                  <el-icon><ZoomIn /></el-icon>
                </el-button>
                <el-button
                  type="danger"
                  size="small"
                  circle
                  @click.stop="handleDelete(image.url)"
                  title="删除"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
          </div>
          <div class="image-info">
            <div class="image-name" :title="image.name">{{ truncateText(image.name, 30) }}</div>
            <div class="image-meta">
              <span>{{ formatFileSize(image.size) }}</span>
              <span>{{ image.updatedAt }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <el-empty v-if="!loading && images.length === 0" description="暂无图片" />

      <!-- 分页 -->
      <div class="pagination-container" v-if="total > 0">
        <el-pagination
          v-model:current-page="pagination.pageNum"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[30, 60, 90, 120]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>

      <!-- 图片预览对话框 -->
      <el-dialog v-model="previewVisible" title="图片预览" width="80%" center>
        <div class="preview-container">
          <img :src="previewImage.url" :alt="previewImage.name" class="preview-image" />
          <div class="preview-info">
            <p><strong>名称:</strong> {{ previewImage.name }}</p>
            <p><strong>大小:</strong> {{ formatFileSize(previewImage.size) }}</p>
            <p><strong>更新时间:</strong> {{ previewImage.updatedAt }}</p>
            <p><strong>URL:</strong> <code>{{ previewImage.url }}</code></p>
            <el-button type="primary" @click="handleCopyUrl">复制URL</el-button>
          </div>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  Upload,
  Delete,
  Refresh,
  ZoomIn,
  Loading,
  Picture
} from '@element-plus/icons-vue'
import { getImageList, batchDeleteImages, uploadImage } from '../api/imageLibrary'

// 数据
const loading = ref(false)
const images = ref([])
const searchKeyword = ref('')
const selectedImages = ref([])
const previewVisible = ref(false)
const previewImage = ref({})
const total = ref(0)
const selectedCategory = ref('') // 当前选择的目录
const uploadCategory = ref('others') // 上传时选择的目录，默认为"其他图片"

// 图片加载状态管理
const imageLoaded = ref({})
const imageObserver = ref(null)

// 分页信息
const pagination = reactive({
  pageNum: 1,
  pageSize: 30 // 每页30个（3行 x 10列）
})

// 图片懒加载处理
const setupLazyLoading = () => {
  // 清理旧的观察器
  if (imageObserver.value) {
    imageObserver.value.disconnect()
  }

  // 使用 Intersection Observer 实现更精确的懒加载
  imageObserver.value = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const img = entry.target
        const dataSrc = img.getAttribute('data-src')
        if (dataSrc && !img.src) {
          img.src = dataSrc
          img.removeAttribute('data-src')
        }
        imageObserver.value.unobserve(img)
      }
    })
  }, {
    rootMargin: '50px' // 提前50px开始加载
  })

  // 观察所有懒加载图片
  nextTick(() => {
    const lazyImages = document.querySelectorAll('.lazy-image[data-src]')
    lazyImages.forEach(img => {
      imageObserver.value.observe(img)
    })
  })
}

// 图片加载成功
const handleImageLoad = (event) => {
  const img = event.target
  const src = img.src || img.getAttribute('data-src')
  if (src) {
    imageLoaded.value[src] = true
    img.classList.add('loaded')
  }
}

// 图片加载失败
const handleImageError = (event) => {
  const img = event.target
  const src = img.src || img.getAttribute('data-src')
  if (src) {
    // 标记为加载失败，显示错误状态
    imageLoaded.value[src] = 'error'
    const placeholder = img.nextElementSibling
    if (placeholder && placeholder.classList.contains('image-placeholder')) {
      placeholder.style.color = '#f56c6c'
      // 使用 Vue 的方式更新内容
      const icon = placeholder.querySelector('.loading-icon')
      if (icon) {
        icon.style.display = 'none'
      }
      const errorText = document.createElement('div')
      errorText.style.cssText = 'font-size: 12px; margin-top: 5px; text-align: center;'
      errorText.textContent = '加载失败'
      placeholder.appendChild(errorText)
    }
  }
}

// 初始化数据
const initData = async () => {
  loading.value = true
  // 重置图片加载状态
  imageLoaded.value = {}
  
  try {
    // 优化：搜索时也使用合理的分页大小，避免一次性加载过多数据
    // 由于后端已优化分页性能，我们可以使用较大的pageSize进行搜索，但限制在合理范围内
    let params = {}
    if (searchKeyword.value) {
      // 搜索时使用较大的pageSize（但不超过500），以便在前端进行搜索过滤
      // 如果搜索结果很多，用户可以通过分页查看更多结果
      params = {
        pageNum: 1,
        pageSize: 500 // 从10000减少到500，避免请求过大导致卡死
      }
    } else {
      params = {
        pageNum: pagination.pageNum,
        pageSize: pagination.pageSize
      }
    }
    
    // 如果选择了目录，添加category参数
    if (selectedCategory.value) {
      params.category = selectedCategory.value
    }
    
    const response = await getImageList(params)
    if (response.code === 200) {
      let allImages = response.data || []
      
      // 如果有搜索关键词，进行过滤
      if (searchKeyword.value) {
        const keyword = searchKeyword.value.toLowerCase()
        allImages = allImages.filter(img => img.name.toLowerCase().includes(keyword))
        // 搜索时，total为过滤后的数量
        total.value = allImages.length
        // 对过滤后的结果进行分页
        const start = (pagination.pageNum - 1) * pagination.pageSize
        const end = start + pagination.pageSize
        images.value = allImages.slice(start, end)
      } else {
        images.value = allImages
        total.value = response.total || 0
      }
      
      // 按更新时间倒序排列（后端应该已经排序，这里确保一下）
      images.value.sort((a, b) => {
        return new Date(b.updatedAt) - new Date(a.updatedAt)
      })
      
      // 设置懒加载
      await nextTick()
      setupLazyLoading()
    } else {
      ElMessage.error(response.message || '获取图片列表失败')
    }
  } catch (error) {
    console.error('获取图片列表失败:', error)
    ElMessage.error('获取图片列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  // 搜索时重置到第一页
  pagination.pageNum = 1
  initData()
}

// 分页大小改变
const handleSizeChange = (val) => {
  pagination.pageSize = val
  pagination.pageNum = 1
  initData()
}

// 页码改变
const handlePageChange = (val) => {
  pagination.pageNum = val
  initData()
}

// 目录改变
const handleCategoryChange = () => {
  selectedImages.value = []
  searchKeyword.value = ''
  pagination.pageNum = 1
  initData()
}

// 刷新
const handleRefresh = () => {
  selectedImages.value = []
  searchKeyword.value = ''
  selectedCategory.value = ''
  pagination.pageNum = 1
  initData()
}

// 切换选择
const toggleSelect = (url) => {
  const index = selectedImages.value.indexOf(url)
  if (index > -1) {
    selectedImages.value.splice(index, 1)
  } else {
    selectedImages.value.push(url)
  }
}

// 预览图片
const handlePreview = (image) => {
  previewImage.value = image
  previewVisible.value = true
}

// 复制URL
const handleCopyUrl = async () => {
  try {
    await navigator.clipboard.writeText(previewImage.value.url)
    ElMessage.success('URL已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 删除单个图片
const handleDelete = async (url) => {
  try {
    await ElMessageBox.confirm('确定要删除这张图片吗？', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await batchDeleteImages([url])
    ElMessage.success('删除成功')
    
    // 刷新列表
    await initData()
    // 从选中列表中移除
    const index = selectedImages.value.indexOf(url)
    if (index > -1) {
      selectedImages.value.splice(index, 1)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 批量删除
const handleBatchDelete = async () => {
  if (selectedImages.value.length === 0) {
    ElMessage.warning('请至少选择一张图片')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedImages.value.length} 张图片吗？此操作不可恢复！`,
      '确认批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await batchDeleteImages(selectedImages.value)
    ElMessage.success('删除成功')
    
    // 刷新列表
    await initData()
    // 清空选中列表
    selectedImages.value = []
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error('批量删除失败')
    }
  }
}

// 上传状态管理
const isUploading = ref(false)
const fileChangeTimer = ref(null)

// 上传前验证（仅验证，不阻止上传）
const beforeUpload = (file) => {
  const isImage = /\.(jpg|jpeg|png|gif|webp)$/i.test(file.name)
  if (!isImage) {
    ElMessage.error('请上传JPG、PNG、GIF或WEBP格式的图片')
    return false
  }

  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    ElMessage.error('图片大小不能超过10MB')
    return false
  }

  return true
}

// 处理单个文件上传
const uploadSingleFile = async (file) => {
  const formData = new FormData()
  formData.append('file', file.raw || file) // 兼容 file 对象和 raw 属性

  // 使用选择的目录进行上传
  const response = await uploadImage(formData, uploadCategory.value)
  if (response.code === 200 && response.data && response.data.imageUrl) {
    return { success: true, file: file.name }
  } else {
    throw new Error(response.message || '上传失败')
  }
}

// 处理文件选择变化（批量上传）
const handleFileChange = (file, fileList) => {
  // 如果正在上传，忽略新的文件选择
  if (isUploading.value) {
    return
  }

  // 清除之前的定时器
  if (fileChangeTimer.value) {
    clearTimeout(fileChangeTimer.value)
  }

  // 设置新的定时器，等待所有文件都添加到 fileList
  // Element Plus 会在选择文件时逐个触发 on-change
  fileChangeTimer.value = setTimeout(() => {
    // 获取所有有效的文件（通过验证的，并且有 raw 属性）
    const validFiles = fileList.filter(f => f.status !== 'fail' && f.raw)
    
    // 如果没有有效文件，直接返回
    if (validFiles.length === 0) {
      return
    }

    // 开始上传
    uploadFiles(validFiles)
  }, 200) // 增加延迟，确保所有文件都已添加
}

// 批量上传文件
const uploadFiles = async (files) => {
  if (isUploading.value || files.length === 0) {
    return
  }

  isUploading.value = true
  let successCount = 0
  let failCount = 0

  try {
    // 显示上传进度消息
    const message = ElMessage({
      message: `正在上传图片 (0/${files.length})...`,
      type: 'info',
      duration: 0 // 不自动关闭
    })

    // 串行上传每个文件
    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      try {
        await uploadSingleFile(file)
        successCount++
        // 更新进度消息
        message.message = `正在上传图片 (${i + 1}/${files.length})...`
      } catch (error) {
        console.error(`文件 ${file.name} 上传失败:`, error)
        failCount++
      }
    }

    // 关闭进度消息
    message.close()

    // 显示最终结果
    if (successCount > 0 && failCount === 0) {
      ElMessage.success(`成功上传 ${successCount} 张图片`)
      // 刷新列表
      await initData()
    } else if (successCount > 0 && failCount > 0) {
      ElMessage.warning(`成功上传 ${successCount} 张，失败 ${failCount} 张`)
      // 刷新列表
      await initData()
    } else {
      ElMessage.error(`上传失败，共 ${failCount} 张图片上传失败`)
    }
  } catch (error) {
    console.error('批量上传失败:', error)
    ElMessage.error('批量上传失败，请稍后再试')
  } finally {
    isUploading.value = false
    // 清空上传组件的文件列表
    if (uploadRef.value) {
      uploadRef.value.clearFiles()
    }
  }
}

// 工具函数
const truncateText = (text, length) => {
  if (!text) return ''
  return text.length > length ? text.substring(0, length) + '...' : text
}

const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
}

// 组件挂载时
onMounted(() => {
  initData()
})

// 组件卸载时清理
onUnmounted(() => {
  if (imageObserver.value) {
    imageObserver.value.disconnect()
  }
})
</script>

<style scoped>
.image-library-container {
  padding: 0 0 20px 0;
}

.page-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #333;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 10px 0;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
}

.upload-btn {
  margin-right: 10px;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(10, 1fr); /* 每行10个 */
  gap: 15px;
  margin-top: 20px;
  contain: layout style paint; /* 优化：减少重绘 */
}

@media (max-width: 1600px) {
  .image-grid {
    grid-template-columns: repeat(8, 1fr); /* 小屏幕8个 */
  }
}

@media (max-width: 1200px) {
  .image-grid {
    grid-template-columns: repeat(6, 1fr); /* 更小屏幕6个 */
  }
}

@media (max-width: 768px) {
  .image-grid {
    grid-template-columns: repeat(4, 1fr); /* 移动端4个 */
  }
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.image-item {
  border: 2px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s;
  background: #fff;
  will-change: transform; /* 优化：提示浏览器优化 */
}

.image-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px 0 rgba(64, 158, 255, 0.2);
}

.image-item.selected {
  border-color: #409eff;
  background: #ecf5ff;
}

.image-wrapper {
  position: relative;
  width: 100%;
  padding-top: 100%; /* 1:1 宽高比 */
  overflow: hidden;
  background: #f5f7fa;
}

.image-preview {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  opacity: 0;
  transition: opacity 0.3s ease-in-out;
}

.image-preview.loaded {
  opacity: 1;
}

/* 加载占位符 */
.image-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #909399;
  z-index: 1;
}

.loading-icon {
  font-size: 24px;
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.image-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s;
}

.image-item:hover .image-overlay {
  opacity: 1;
}

.image-overlay .el-checkbox {
  position: absolute;
  top: 10px;
  left: 10px;
}

.image-actions {
  display: flex;
  gap: 10px;
}

.image-info {
  padding: 10px;
}

.image-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #909399;
}

.preview-container {
  display: flex;
  gap: 20px;
}

.preview-image {
  max-width: 60%;
  max-height: 70vh;
  object-fit: contain;
  border-radius: 8px;
}

.preview-info {
  flex: 1;
}

.preview-info p {
  margin-bottom: 10px;
  line-height: 1.6;
}

.preview-info code {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  word-break: break-all;
}
</style>

