<template>
  <div class="carousel-container">
    <el-card>
      <h2 class="page-title">轮播图管理</h2>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-right">
          <el-button type="primary" @click="handleAddCarousel">
            <el-icon>
              <plus />
            </el-icon>
            新增轮播图
          </el-button>
        </div>
      </div>

      <!-- 轮播图列表 -->
      <div class="carousel-list">
        <div v-for="item in carousels" :key="item.id" class="carousel-item">
          <el-card class="item-card" shadow="hover">
            <div class="card-media">
              <img v-if="item.image" :src="item.image" alt="轮播图" class="carousel-image" />
              <div v-else class="image-placeholder">
                <el-icon>
                  <Picture />
                </el-icon>
                <span>暂无图片</span>
              </div>

              <div class="media-overlay">
                <el-tag :type="item.status ? 'success' : 'info'" size="small">
                  {{ item.status ? '已启用' : '已停用' }}
                </el-tag>
                <span class="sort-pill">排序 {{ item.sort }}</span>
              </div>

              <div class="media-actions">
                <el-tooltip content="编辑" placement="top">
                  <el-button
                    circle
                    size="small"
                    type="primary"
                    @click.stop="handleEditCarousel(item)"
                  >
                    <el-icon>
                      <Edit />
                    </el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top">
                  <el-button
                    circle
                    size="small"
                    type="danger"
                    @click.stop="handleDeleteCarousel(item.id)"
                  >
                    <el-icon>
                      <Delete />
                    </el-icon>
                  </el-button>
                </el-tooltip>
              </div>
            </div>

            <div class="item-content">
              <div class="title-row">
                <div>
                  <p class="carousel-title">{{ item.title || '未设置标题' }}</p>
                  <p class="carousel-subtitle">创建于 {{ formatDate(item.created_at) }}</p>
                </div>
                <el-switch v-model="item.status" @change="handleStatusChange(item)" />
              </div>

              <div v-if="item.link" class="carousel-link">
                <el-icon>
                  <Link />
                </el-icon>
                <a :href="item.link" target="_blank" rel="noopener">{{ item.link }}</a>
              </div>
              <div v-else class="carousel-link muted">
                <el-icon>
                  <Link />
                </el-icon>
                <span>未设置跳转链接</span>
              </div>

              <div class="meta-row">
                <div class="meta-pill">
                  <span>排序</span>
                  <strong>{{ item.sort }}</strong>
                </div>
                <div class="meta-pill">
                  <span>ID</span>
                  <strong>{{ item.id }}</strong>
                </div>
              </div>

              <div class="item-actions">
                <el-button type="primary" size="small" @click="handleEditCarousel(item)">
                  编辑
                </el-button>
                <el-button type="danger" size="small" @click="handleDeleteCarousel(item.id)">
                  删除
                </el-button>
              </div>
            </div>
          </el-card>
        </div>

        <!-- 空状态 -->
        <div v-if="carousels.length === 0" class="empty-state">
          <el-empty description="暂无轮播图数据" />
        </div>
      </div>
      </el-card>

      <!-- 新增/编辑轮播图弹窗 -->
      <el-dialog v-model="dialogVisible" :title="dialogType === 'add' ? '新增轮播图' : '编辑轮播图'" width="500px">
        <el-form ref="carouselFormRef" :model="carouselForm" :rules="carouselRules" label-width="80px">
          <el-form-item label="图片" prop="imageUrl">
            <div class="upload-card">
              <el-upload
                class="avatar-uploader"
                action=""
                :show-file-list="false"
                :on-success="handleImageUploadSuccess"
                :before-upload="beforeUpload"
              >
                <img v-if="carouselForm.imageUrl" :src="carouselForm.imageUrl" class="avatar" />
                <div v-else class="avatar-uploader-icon">
                  <el-icon>
                    <Upload />
                  </el-icon>
                  <span class="upload-text">上传轮播图</span>
                </div>
              </el-upload>
              <div class="upload-meta">
                <p>建议尺寸 1200×400， JPG/PNG 不超过 2MB</p>
                <p>上传后可实时预览，支持重新选择</p>
              </div>
            </div>
          </el-form-item>
          <el-form-item label="标题" prop="title">
            <el-input v-model="carouselForm.title" placeholder="请输入轮播图标题" />
          </el-form-item>
          <el-form-item label="链接" prop="link">
            <el-input v-model="carouselForm.link" placeholder="请输入跳转链接" />
          </el-form-item>
          <el-form-item label="排序" prop="sort">
            <el-input-number v-model="carouselForm.sort" :min="1" :max="999" :step="1" placeholder="请输入排序号" />
          </el-form-item>
          <el-form-item label="状态" prop="status">
            <el-switch v-model="carouselForm.status" />
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Upload, Picture, Link, Edit, Delete } from '@element-plus/icons-vue'
import { getCarouselList, createCarousel, updateCarousel, deleteCarousel, uploadCarouselImage } from '../api/carousel'
import { formatDate } from '../utils/time-format'

// 轮播图列表
const carousels = ref([])

// 弹窗相关
const dialogVisible = ref(false)
const dialogType = ref('add')
const carouselFormRef = ref(null)
const carouselForm = reactive({
  id: '',
  title: '',
  imageUrl: '',
  link: '',
  sort: 1,
  status: true
})

// 构建请求负载
const buildRequestPayload = (source) => ({
  image: source.imageUrl || source.image || '',
  title: source.title || '',
  link: source.link || '',
  sort: Number(source.sort) || 1,
  status: source.status ? 1 : 0
})

// 表单验证规则
const carouselRules = {
  imageUrl: [
    { required: true, message: '请上传轮播图', trigger: 'change' }
  ],
  title: [
    { max: 50, message: '标题不能超过 50 个字符', trigger: 'blur' }
  ],
  link: [
    { max: 255, message: '链接不能超过 255 个字符', trigger: 'blur' }
  ],
  sort: [
    { required: true, message: '请输入排序号', trigger: 'blur' },
    { type: 'number', min: 1, max: 999, message: '排序号范围在 1 到 999', trigger: 'blur' }
  ]
}

// 初始化数据
const initData = async () => {
  try {
    // 加载轮播图数据
    const response = await getCarouselList()

    if (response && response.code === 200 && Array.isArray(response.data)) {
      // 转换后端数据格式，确保前端正确显示
      carousels.value = response.data.map(item => ({
        ...item,
        status: item.status === 1, // 将后端的整数状态(1/0)转换为前端的布尔值(true/false)
        title: item.title || ''   // 保留后端返回的title值，如果没有则设为空字符串
      }))
    } else {
      const msg = response?.message || '加载轮播图数据失败'
      console.warning(msg, response)
      ElMessage.warning(msg)
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.warning('加载轮播图数据失败')
  }
}

// 打开新增轮播图弹窗
const handleAddCarousel = () => {
  dialogType.value = 'add'
  // 重置表单
  if (carouselFormRef.value) {
    carouselFormRef.value.resetFields()
  }
  // 清空表单数据
  Object.assign(carouselForm, {
    id: '',
    title: '',
    imageUrl: '',
    link: '',
    sort: 1,
    status: true
  })
  dialogVisible.value = true
}

// 打开编辑轮播图弹窗
const handleEditCarousel = (row) => {
  dialogType.value = 'edit'
  // 复制行数据到表单，并确保image字段正确映射到imageUrl
  Object.assign(carouselForm, {
    ...row,
    imageUrl: row.image, // 重要：后端返回的是image字段，需要映射到前端表单的imageUrl
    title: row.title || '' // 确保title有默认值
  })
  dialogVisible.value = true
}

// 删除轮播图
const handleDeleteCarousel = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个轮播图吗？', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    // 调用真实的删除API
    await deleteCarousel(id)

    // 重新加载数据
    await initData()

    ElMessage.success('删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败，请重试')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    // 验证表单
    await carouselFormRef.value.validate()

    const payload = buildRequestPayload(carouselForm)

    if (dialogType.value === 'add') {
      await createCarousel(payload)

      // 重新加载数据
      await initData()

      ElMessage.success('新增成功')
    } else {
      // 创建更新对象，将布尔值status转换为整数类型
      await updateCarousel(carouselForm.id, payload)

      // 重新加载数据
      await initData()

      ElMessage.success('更新成功')
    }

    dialogVisible.value = false
  } catch (error) {
    console.error('提交失败:', error)
    ElMessage.error('操作失败，请重试')
  }
}

// 更新排序
const handleSortChange = async (item) => {
  try {
    await updateCarousel(item.id, buildRequestPayload(item))

    // 重新加载数据以确保排序正确显示
    await initData()

    ElMessage.success('排序已更新')
  } catch (error) {
    console.error('更新排序失败:', error)
    ElMessage.error('更新排序失败')
  }
}

// 更新状态
const handleStatusChange = async (item) => {
  try {
    await updateCarousel(item.id, buildRequestPayload(item))

    ElMessage.success('状态已更新')
  } catch (error) {
    console.error('更新状态失败:', error)
    ElMessage.error('更新状态失败')
    // 回滚状态
    item.status = !item.status
  }
}

// 图片上传相关函数
const beforeUpload = async (file) => {
  // 检查文件类型
  const isImage = file.type.indexOf('image/') !== -1
  if (!isImage) {
    ElMessage.error('只能上传图片文件!')
    return false
  }

  // 检查文件大小（2MB）
  const isLt2M = file.size / 1024 / 1024 < 2
  if (!isLt2M) {
    ElMessage.error('上传图片大小不能超过 2MB!')
    return false
  }

  // 创建FormData对象
  const formData = new FormData()
  formData.append('file', file)

  try {
    // 调用真实的上传API
    const response = await uploadCarouselImage(formData)
    carouselForm.imageUrl = response.data.imageUrl
    ElMessage.success('图片上传成功')
  } catch (error) {
    console.error('图片上传失败:', error)
    ElMessage.error('图片上传失败')
  }

  return false // 阻止默认上传
}

const handleImageUploadSuccess = (response, file) => {
  // 真实环境下的上传成功回调
  console.log('上传成功', response)
}

// 组件挂载时
onMounted(() => {
  initData()
})
</script>

<style scoped>
.carousel-container {
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

.carousel-card {
  margin-bottom: 20px;
}

.carousel-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.carousel-item {
  transition: transform 0.3s ease;
}

.carousel-item:hover {
  transform: translateY(-5px);
}

.item-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.card-media {
  position: relative;
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 12px;
}

.carousel-image {
  width: 100%;
  height: 220px;
  object-fit: cover;
  display: block;
}

.image-placeholder {
  width: 100%;
  height: 220px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #eef2ff, #f8fbff);
  border: 1px dashed #dcdfe6;
  color: #a0a7b4;
  gap: 8px;
}

.image-placeholder .el-icon {
  font-size: 48px;
}

.media-overlay {
  position: absolute;
  top: 16px;
  left: 16px;
  display: flex;
  gap: 8px;
  align-items: center;
}

.media-actions {
  position: absolute;
  top: 16px;
  right: 16px;
  display: flex;
  gap: 8px;
}

.sort-pill {
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.35);
  color: #fff;
}

.item-content {
  padding: 16px 4px 4px;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 10px;
}

.carousel-title {
  font-size: 18px;
  font-weight: 600;
  color: #1f2d3d;
  margin: 0;
}

.carousel-subtitle {
  margin: 4px 0 0;
  color: #9aa5b1;
  font-size: 13px;
}

.carousel-link {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 8px 0 16px;
  font-size: 14px;
  color: #409eff;
  word-break: break-all;
}

.carousel-link.muted {
  color: #c0c4cc;
}

.carousel-link a {
  color: inherit;
  text-decoration: none;
}

.carousel-link a:hover {
  text-decoration: underline;
}

.meta-row {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.meta-pill {
  flex: 1;
  background: #f5f7fa;
  border-radius: 10px;
  padding: 10px 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #606266;
  font-size: 13px;
}

.meta-pill strong {
  font-size: 16px;
  color: #303133;
}

.item-actions {
  margin-top: auto;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.empty-state {
  grid-column: 1 / -1;
  padding: 60px 0;
}

/* 上传组件样式 */
.upload-card {
  width: 100%;
  border: 1px dashed #dcdfe6;
  border-radius: 12px;
  padding: 16px;
  background: #f9fbff;
}

.avatar-uploader {
  width: 100%;
  display: block;
}

.avatar-uploader .avatar {
  width: 100%;
  height: 220px;
  display: block;
  object-fit: cover;
  border-radius: 8px;
}

:deep(.el-upload) {
  width: 100%;
  display: block;
}

.avatar-uploader-icon {
  width: 100%;
  height: 220px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #f1f5ff;
  border-radius: 8px;
  color: #8690a3;
  cursor: pointer;
  border: 2px dashed #cfd8ff;
  gap: 8px;
  transition: border-color 0.3s ease, color 0.3s ease;
}

.avatar-uploader-icon:hover {
  border-color: #8da2ff;
  color: #5c6ddc;
}

.upload-text {
  font-size: 14px;
}

.upload-meta {
  margin-top: 12px;
  color: #8c939d;
  font-size: 13px;
  line-height: 1.6;
}

/* 响应式布局 */
@media (max-width: 768px) {
  .carousel-list {
    grid-template-columns: 1fr;
  }
}
</style>