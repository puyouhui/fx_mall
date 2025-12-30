<template>
  <div class="image-library-container">
    <el-card>
      <h2 class="page-title">图库管理</h2>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
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
          <el-upload
            class="upload-btn"
            action=""
            :show-file-list="false"
            :before-upload="beforeUpload"
            :http-request="handleUpload"
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
            <img :src="image.url" :alt="image.name" class="image-preview" />
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  Upload,
  Delete,
  Refresh,
  ZoomIn
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

// 分页信息
const pagination = reactive({
  pageNum: 1,
  pageSize: 30 // 每页30个（3行 x 10列）
})

// 初始化数据
const initData = async () => {
  loading.value = true
  try {
    // 如果有搜索关键词，获取所有数据用于搜索
    // 否则使用分页参数
    let params = {}
    if (searchKeyword.value) {
      // 搜索时获取所有数据（第一页，但pageSize很大）
      params = {
        pageNum: 1,
        pageSize: 10000 // 获取所有数据用于搜索
      }
    } else {
      params = {
        pageNum: pagination.pageNum,
        pageSize: pagination.pageSize
      }
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

// 刷新
const handleRefresh = () => {
  selectedImages.value = []
  searchKeyword.value = ''
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

// 上传前验证
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

// 上传图片
const handleUpload = async (options) => {
  try {
    const { file } = options
    ElMessage({ message: '图片上传中...', type: 'info' })

    const formData = new FormData()
    formData.append('file', file)

    const response = await uploadImage(formData)
    if (response.code === 200 && response.data && response.data.imageUrl) {
      ElMessage.success('图片上传成功')
      // 刷新列表
      await initData()
    } else {
      ElMessage.error('图片上传失败: ' + (response.message || '未知错误'))
    }
  } catch (error) {
    console.error('图片上传失败:', error)
    ElMessage.error('图片上传失败，请稍后再试')
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

