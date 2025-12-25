<template>
  <div class="product-requests-page">
    <el-card class="product-requests-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">新品需求管理</span>
          <span class="sub">查看和管理用户提交的新品需求</span>
        </div>
        <div class="actions">
          <el-select v-model="statusFilter" placeholder="需求状态" clearable style="width: 150px; margin-right: 10px;"
            @change="handleSearch">
            <el-option label="待处理" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="已完成" value="completed" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="requests" border stripe class="product-requests-table" empty-text="暂无需求数据" row-key="id">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="用户信息" min-width="150">
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
            <span v-else>用户ID: {{ scope.row.user_id }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="product_name" label="需求产品" min-width="200" />
        <el-table-column prop="brand" label="品牌" width="120">
          <template #default="scope">
            {{ scope.row.brand || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="monthly_quantity" label="月消耗数量" width="120" align="center">
          <template #default="scope">
            {{ scope.row.monthly_quantity || 0 }}
          </template>
        </el-table-column>
        <el-table-column prop="description" label="需求说明" min-width="250" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ formatStatus(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button type="primary" link size="small" @click="handleViewDetail(scope.row)">
              查看详情
            </el-button>
            <el-button type="success" link size="small" @click="handleUpdateStatus(scope.row, 'processing')"
              v-if="scope.row.status === 'pending'">
              处理中
            </el-button>
            <el-button type="success" link size="small" @click="handleUpdateStatus(scope.row, 'completed')"
              v-if="scope.row.status === 'processing'">
              已完成
            </el-button>
            <el-button type="danger" link size="small" @click="handleUpdateStatus(scope.row, 'rejected')"
              v-if="scope.row.status === 'pending' || scope.row.status === 'processing'">
              拒绝
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
    <el-dialog
      v-model="detailDialogVisible"
      title="需求详情"
      width="800px"
      :close-on-click-modal="false"
    >
      <div v-loading="detailLoading" class="detail-content">
        <el-descriptions :column="2" border v-if="currentRequest">
          <el-descriptions-item label="需求ID">{{ currentRequest.id }}</el-descriptions-item>
          <el-descriptions-item label="用户ID">{{ currentRequest.user_id }}</el-descriptions-item>
          <el-descriptions-item label="用户姓名">{{ currentRequest.user_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="用户编号">用户{{ currentRequest.user_code || currentRequest.user_id }}</el-descriptions-item>
          <el-descriptions-item label="用户电话">{{ currentRequest.user_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentRequest.status)">
              {{ formatStatus(currentRequest.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="需求产品" :span="2">{{ currentRequest.product_name }}</el-descriptions-item>
          <el-descriptions-item label="品牌">{{ currentRequest.brand || '-' }}</el-descriptions-item>
          <el-descriptions-item label="月消耗数量">{{ currentRequest.monthly_quantity || 0 }}</el-descriptions-item>
          <el-descriptions-item label="需求说明" :span="2">
            <div style="white-space: pre-wrap;">{{ currentRequest.description || '-' }}</div>
          </el-descriptions-item>
          <el-descriptions-item label="管理员备注" :span="2">
            <div style="white-space: pre-wrap;">{{ currentRequest.admin_remark || '-' }}</div>
          </el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ formatDate(currentRequest.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDate(currentRequest.updated_at) }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="detailDialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="handleUpdateStatusDialog">更新状态</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 更新状态对话框 -->
    <el-dialog
      v-model="statusDialogVisible"
      title="更新状态"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="statusForm" label-width="100px">
        <el-form-item label="状态" required>
          <el-select v-model="statusForm.status" placeholder="请选择状态">
            <el-option label="待处理" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="已完成" value="completed" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item label="管理员备注">
          <el-input
            v-model="statusForm.admin_remark"
            type="textarea"
            :rows="4"
            placeholder="请输入备注信息（选填）"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="statusDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleConfirmUpdateStatus" :loading="updating">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getProductRequests, updateProductRequestStatus } from '../api/productRequests'

const loading = ref(false)
const requests = ref([])
const statusFilter = ref('')
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 详情对话框
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const currentRequest = ref(null)

// 状态更新对话框
const statusDialogVisible = ref(false)
const updating = ref(false)
const statusForm = reactive({
  status: '',
  admin_remark: ''
})

const loadRequests = async () => {
  loading.value = true
  try {
    const res = await getProductRequests({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      status: statusFilter.value
    })
    if (res.code === 200) {
      requests.value = Array.isArray(res.data) ? res.data : []
      pagination.total = res.total || requests.value.length
    }
  } catch (error) {
    console.error('获取需求列表失败:', error)
    ElMessage.error('获取需求列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.pageNum = 1
  loadRequests()
}

const handlePageChange = (page) => {
  pagination.pageNum = page
  loadRequests()
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.pageNum = 1
  loadRequests()
}

const formatDate = (value) => {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

const formatStatus = (status) => {
  const statusMap = {
    'pending': '待处理',
    'processing': '处理中',
    'completed': '已完成',
    'rejected': '已拒绝'
  }
  return statusMap[status] || status
}

const getStatusType = (status) => {
  const typeMap = {
    'pending': 'info',
    'processing': 'warning',
    'completed': 'success',
    'rejected': 'danger'
  }
  return typeMap[status] || ''
}

const handleViewDetail = (row) => {
  currentRequest.value = row
  detailDialogVisible.value = true
}

const handleUpdateStatus = (row, status) => {
  currentRequest.value = row
  statusForm.status = status
  statusForm.admin_remark = row.admin_remark || ''
  statusDialogVisible.value = true
}

const handleUpdateStatusDialog = () => {
  if (!currentRequest.value) return
  statusForm.status = currentRequest.value.status
  statusForm.admin_remark = currentRequest.value.admin_remark || ''
  statusDialogVisible.value = true
}

const handleConfirmUpdateStatus = async () => {
  if (!statusForm.status) {
    ElMessage.warning('请选择状态')
    return
  }

  if (!currentRequest.value) {
    ElMessage.error('未选择需求')
    return
  }

  updating.value = true
  try {
    const res = await updateProductRequestStatus(currentRequest.value.id, {
      status: statusForm.status,
      admin_remark: statusForm.admin_remark
    })
    
    if (res.code === 200) {
      ElMessage.success('更新成功')
      statusDialogVisible.value = false
      // 刷新列表
      loadRequests()
      // 如果详情对话框打开，刷新详情
      if (detailDialogVisible.value) {
        const index = requests.value.findIndex(r => r.id === currentRequest.value.id)
        if (index !== -1) {
          currentRequest.value = requests.value[index]
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
  loadRequests()
})
</script>

<style scoped>
.product-requests-page {
  padding: 20px;
}

.product-requests-card {
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

.actions {
  display: flex;
  align-items: center;
}

.product-requests-table {
  margin-top: 20px;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.detail-content {
  padding: 20px 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>

