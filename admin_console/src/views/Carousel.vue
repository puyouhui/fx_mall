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
            <el-card class="item-card">
              <img v-if="item.image" :src="item.image" alt="轮播图" class="carousel-image" />
              <div v-else class="image-placeholder">
                <el-icon>
                  <Picture />
                </el-icon>
                <span>暂无图片</span>
              </div>

              <div class="item-info">
                <div class="info-row">
                  <span class="label">标题:</span>
                  <span class="value">{{ item.title || '未设置' }}</span>
                </div>
                <div class="info-row">
                  <span class="label">链接:</span>
                  <span class="value">{{ item.link || '未设置' }}</span>
                </div>
                <div class="info-row">
                  <span class="label">排序:</span>
                  <!-- <el-input-number v-model="item.sort" :min="1" :max="999" :step="1" @change="handleSortChange(item)"
                    style="width: 80px;" /> -->
                    <span>{{item.sort}}</span>
                </div>
                <div class="info-row">
                  <span class="label">状态:</span>
                  <el-switch v-model="item.status" @change="handleStatusChange(item)" />
                </div>
                <div class="info-row">
                  <span class="label">创建时间:</span>
                  <span class="value">{{ formatDate(item.created_at) }}</span>
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
            <el-upload class="avatar-uploader" action="" :show-file-list="false" :on-success="handleImageUploadSuccess"
              :before-upload="beforeUpload">
              <img v-if="carouselForm.imageUrl" :src="carouselForm.imageUrl" class="avatar" />
              <div v-else class="avatar-uploader-icon">
                <el-icon>
                  <Upload />
                </el-icon>
              </div>
            </el-upload>
            <div class="upload-hint">点击上传轮播图，建议尺寸：1200x400</div>
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
import { Plus, Upload, Picture } from '@element-plus/icons-vue'
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

    // 转换后端数据格式，确保前端正确显示
    carousels.value = response.data.map(item => ({
      ...item,
      status: item.status === 1, // 将后端的整数状态(1/0)转换为前端的布尔值(true/false)
      title: item.title || ''   // 保留后端返回的title值，如果没有则设为空字符串
    }))
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载轮播图数据失败')
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

    if (dialogType.value === 'add') {
      // 创建新对象，不包含id字段，并且将imageUrl改为image，与后端结构体字段名保持一致
      const carouselData = {
        image: carouselForm.imageUrl, // 注意：后端字段名是image，不是imageUrl
        title: carouselForm.title,    // 添加title字段，确保标题数据被保存
        link: carouselForm.link,
        sort: carouselForm.sort,
        status: carouselForm.status ? 1 : 0 // 后端status是整数类型：1-启用，0-禁用
      }
      // 调用真实的新增API
      await createCarousel(carouselData)

      // 重新加载数据
      await initData()

      ElMessage.success('新增成功')
    } else {
      // 创建更新对象，将布尔值status转换为整数类型
      const updateData = {
        image: carouselForm.imageUrl, // 确保使用正确的字段名
        title: carouselForm.title,    // 明确添加title字段
        link: carouselForm.link,
        sort: carouselForm.sort,
        status: carouselForm.status ? 1 : 0 // 后端status是整数类型：1-启用，0-禁用
      }
      // 调用真实的更新API
      await updateCarousel(carouselForm.id, updateData)

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
    // 向后端发送排序更新请求
    await updateCarousel(item.id, { sort: item.sort })

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
    // 将前端的布尔值状态转换为后端需要的整数(1/0)，然后发送请求
    await updateCarousel(item.id, { status: item.status ? 1 : 0 })

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
}

.carousel-image {
  width: 100%;
  height: 200px;
  object-fit: cover;
  margin-bottom: 15px;
}

.image-placeholder {
  width: 100%;
  height: 200px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #f5f7fa;
  border: 1px dashed #dcdfe6;
  margin-bottom: 15px;
}

.image-placeholder .el-icon {
  font-size: 48px;
  color: #c0c4cc;
  margin-bottom: 10px;
}

.image-placeholder span {
  color: #909399;
}

.item-info {
  margin-bottom: 15px;
}

.info-row {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.label {
  width: 80px;
  color: #606266;
}

.value {
  flex: 1;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
.avatar-uploader .avatar {
  width: 100%;
  height: 200px;
  display: block;
  object-fit: cover;
}

.avatar-uploader-icon {
  width: 100%;
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f7fa;
  border: 1px dashed #dcdfe6;
  font-size: 28px;
  color: #c0c4cc;
  cursor: pointer;
}

.upload-hint {
  margin-top: 8px;
  color: #909399;
  font-size: 12px;
}

/* 响应式布局 */
@media (max-width: 768px) {
  .carousel-list {
    grid-template-columns: 1fr;
  }
}
</style>