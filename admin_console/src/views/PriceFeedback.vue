<template>
  <div class="price-feedback-page">
    <el-card class="price-feedback-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">价格反馈管理</span>
          <span class="sub">查看和管理用户提交的价格反馈</span>
        </div>
        <div class="actions">
          <el-select v-model="statusFilter" placeholder="反馈状态" clearable style="width: 150px; margin-right: 10px;"
            @change="handleSearch">
            <el-option label="待处理" value="pending" />
            <el-option label="已处理" value="processed" />
          </el-select>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="feedbacks" border stripe class="price-feedback-table" empty-text="暂无反馈数据" row-key="id">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column label="用户信息" min-width="150" align="center">
          <template #default="scope">
            <div v-if="scope.row.user_name || scope.row.user_code">
              <div>{{ scope.row.user_name || '未命名' }}</div>
              <div style="color: #909399; font-size: 12px;">
                用户{{ scope.row.user_code || scope.row.user_id }}
              </div>
              <div v-if="scope.row.user_phone" style="color: #909399; font-size: 12px;">
                {{ scope.row.user_phone }}
              </div>
            </div>
            <span v-else>未登录用户</span>
          </template>
        </el-table-column>
        <el-table-column prop="product_name" label="商品名称" min-width="200" align="center" />
        <el-table-column label="价格对比" width="200" align="center">
          <template #default="scope">
            <div style="color: #FF4D4F; font-weight: 600;">
              平台：
              <span v-if="scope.row.platform_price_min === scope.row.platform_price_max">
                ¥{{ scope.row.platform_price_min }}
              </span>
              <span v-else>
                ¥{{ scope.row.platform_price_min }} - ¥{{ scope.row.platform_price_max }}
              </span>
            </div>
            <div style="color: #20CB6B; font-weight: 600; margin-top: 4px;">
              反馈：¥{{ scope.row.competitor_price }}
            </div>
            <div style="color: #909399; font-size: 12px; margin-top: 4px;">
              差价：¥{{ (scope.row.competitor_price - scope.row.platform_price_min).toFixed(2) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="价格截图" width="150" align="center">
          <template #default="scope">
            <div v-if="scope.row.images && scope.row.images.length > 0" class="image-preview">
              <el-image
                v-for="(img, index) in scope.row.images.slice(0, 3)"
                :key="index"
                :src="img"
                :preview-src-list="scope.row.images"
                :initial-index="index"
                fit="cover"
                style="width: 40px; height: 40px; margin-right: 4px; border-radius: 4px;"
                :preview-teleported="true"
              />
            </div>
            <span v-else style="color: #909399;">无图片</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注说明" min-width="200" show-overflow-tooltip align="center" />
        <el-table-column prop="status" label="状态" width="120" align="center">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ formatStatus(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" width="180" align="center">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="scope">
            <el-button type="primary" link size="small" @click="handleViewDetail(scope.row)">
              查看详情
            </el-button>
            <el-button type="success" link size="small" @click="handleUpdateStatus(scope.row, 'processed')"
              v-if="scope.row.status === 'pending'">
              标记已处理
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.pageNum"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="价格反馈详情" width="800px" :close-on-click-modal="false">
      <div v-loading="detailLoading" class="detail-content">
        <el-descriptions :column="2" border v-if="currentFeedback">
          <el-descriptions-item label="反馈ID">{{ currentFeedback.id }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentFeedback.status)">
              {{ formatStatus(currentFeedback.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="用户信息" :span="2">
            <div v-if="currentFeedback.user_name || currentFeedback.user_code">
              <div>姓名：{{ currentFeedback.user_name || '未命名' }}</div>
              <div>用户码：{{ currentFeedback.user_code || currentFeedback.user_id }}</div>
              <div v-if="currentFeedback.user_phone">电话：{{ currentFeedback.user_phone }}</div>
            </div>
            <span v-else>未登录用户</span>
          </el-descriptions-item>
          <el-descriptions-item label="商品ID">{{ currentFeedback.product_id }}</el-descriptions-item>
          <el-descriptions-item label="商品名称" :span="1">
            {{ currentFeedback.product_name }}
          </el-descriptions-item>
          <el-descriptions-item label="平台价格">
            <span style="color: #FF4D4F; font-weight: 600;">
              <span v-if="currentFeedback.platform_price_min === currentFeedback.platform_price_max">
                ¥{{ currentFeedback.platform_price_min }}
              </span>
              <span v-else>
                ¥{{ currentFeedback.platform_price_min }} - ¥{{ currentFeedback.platform_price_max }}
              </span>
            </span>
          </el-descriptions-item>
          <el-descriptions-item label="反馈价格">
            <span style="color: #20CB6B; font-weight: 600;">¥{{ currentFeedback.competitor_price }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="价格差">
            <span style="color: #909399;">
              ¥{{ (currentFeedback.competitor_price - currentFeedback.platform_price_min).toFixed(2) }}
            </span>
          </el-descriptions-item>
          <el-descriptions-item label="备注说明" :span="2">
            {{ currentFeedback.remark || '无' }}
          </el-descriptions-item>
          <el-descriptions-item label="价格截图" :span="2" v-if="currentFeedback.images && currentFeedback.images.length > 0">
            <div class="detail-images">
              <el-image
                v-for="(img, index) in currentFeedback.images"
                :key="index"
                :src="img"
                :preview-src-list="currentFeedback.images"
                :initial-index="index"
                fit="cover"
                style="width: 150px; height: 150px; margin-right: 10px; margin-bottom: 10px; border-radius: 8px;"
                :preview-teleported="true"
              />
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="管理员备注" :span="2">
            {{ currentFeedback.admin_remark || '无' }}
          </el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ formatDate(currentFeedback.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDate(currentFeedback.updated_at) }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleUpdateStatusDialog" v-if="currentFeedback && currentFeedback.status === 'pending'">
          标记已处理
        </el-button>
      </template>
    </el-dialog>

    <!-- 状态更新对话框 -->
    <el-dialog v-model="statusDialogVisible" title="更新反馈状态" width="500px" :close-on-click-modal="false">
      <el-form :model="statusForm" label-width="100px">
        <el-form-item label="状态">
          <el-select v-model="statusForm.status" style="width: 100%;">
            <el-option label="待处理" value="pending" />
            <el-option label="已处理" value="processed" />
          </el-select>
        </el-form-item>
        <el-form-item label="管理员备注">
          <el-input
            v-model="statusForm.admin_remark"
            type="textarea"
            :rows="4"
            placeholder="请输入管理员备注（选填）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmUpdateStatus" :loading="updating">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPriceFeedbacks, updatePriceFeedbackStatus } from '../api/priceFeedback'

const loading = ref(false)
const feedbacks = ref([])
const statusFilter = ref('')
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 详情对话框
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const currentFeedback = ref(null)

// 状态更新对话框
const statusDialogVisible = ref(false)
const updating = ref(false)
const statusForm = reactive({
  status: '',
  admin_remark: ''
})

const loadFeedbacks = async () => {
  loading.value = true
  try {
    const res = await getPriceFeedbacks({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      status: statusFilter.value
    })
    if (res.code === 200) {
      feedbacks.value = Array.isArray(res.data) ? res.data : []
      pagination.total = res.total || feedbacks.value.length
    }
  } catch (error) {
    console.error('获取反馈列表失败:', error)
    ElMessage.error('获取反馈列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.pageNum = 1
  loadFeedbacks()
}

const handlePageChange = (page) => {
  pagination.pageNum = page
  loadFeedbacks()
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.pageNum = 1
  loadFeedbacks()
}

const formatDate = (value) => {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

const formatStatus = (status) => {
  const statusMap = {
    'pending': '待处理',
    'processed': '已处理'
  }
  return statusMap[status] || status
}

const getStatusType = (status) => {
  const typeMap = {
    'pending': 'warning',
    'processed': 'success'
  }
  return typeMap[status] || ''
}

const handleViewDetail = (row) => {
  currentFeedback.value = row
  detailDialogVisible.value = true
}

const handleUpdateStatus = (row, status) => {
  currentFeedback.value = row
  statusForm.status = status
  statusForm.admin_remark = row.admin_remark || ''
  statusDialogVisible.value = true
}

const handleUpdateStatusDialog = () => {
  if (!currentFeedback.value) return
  statusForm.status = 'processed'
  statusForm.admin_remark = currentFeedback.value.admin_remark || ''
  statusDialogVisible.value = true
}

const handleConfirmUpdateStatus = async () => {
  if (!statusForm.status) {
    ElMessage.warning('请选择状态')
    return
  }

  if (!currentFeedback.value) {
    ElMessage.error('未选择反馈')
    return
  }

  updating.value = true
  try {
    const res = await updatePriceFeedbackStatus(currentFeedback.value.id, {
      status: statusForm.status,
      admin_remark: statusForm.admin_remark
    })
    
    if (res.code === 200) {
      ElMessage.success('更新成功')
      statusDialogVisible.value = false
      // 刷新列表
      loadFeedbacks()
      // 如果详情对话框打开，刷新详情
      if (detailDialogVisible.value) {
        const index = feedbacks.value.findIndex(f => f.id === currentFeedback.value.id)
        if (index !== -1) {
          currentFeedback.value = feedbacks.value[index]
        }
      }
    } else {
      ElMessage.error(res.message || '更新失败')
    }
  } catch (error) {
    console.error('更新状态失败:', error)
    ElMessage.error('更新失败，请稍后再试')
  } finally {
    updating.value = false
  }
}

onMounted(() => {
  loadFeedbacks()
})
</script>

<style scoped>
.price-feedback-page {
  padding: 20px;
}

.price-feedback-card {
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title .main {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin-right: 10px;
}

.title .sub {
  font-size: 14px;
  color: #909399;
}

.price-feedback-table {
  margin-top: 20px;
}

.image-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
}

.detail-content {
  padding: 10px 0;
}

.detail-images {
  display: flex;
  flex-wrap: wrap;
  margin-top: 10px;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

