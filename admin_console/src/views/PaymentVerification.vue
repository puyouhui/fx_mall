<template>
  <div class="payment-verification-page">
    <el-card class="verification-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">收款审核管理</span>
          <span class="sub">审核销售员的收款申请</span>
        </div>
        <div class="actions">
          <el-select
            v-model="statusFilter"
            placeholder="审核状态"
            clearable
            style="width: 150px; margin-right: 10px;"
            @change="handleSearch"
          >
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleRefresh">刷新</el-button>
        </div>
      </div>

      <!-- 统计信息 -->
      <div class="stats-summary" v-if="stats">
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="stat-item">
              <div class="stat-label">待审核</div>
              <div class="stat-value pending">{{ stats.pending || 0 }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-item">
              <div class="stat-label">已通过</div>
              <div class="stat-value approved">{{ stats.approved || 0 }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-item">
              <div class="stat-label">已拒绝</div>
              <div class="stat-value rejected">{{ stats.rejected || 0 }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-item">
              <div class="stat-label">总申请数</div>
              <div class="stat-value total">{{ stats.total || 0 }}</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <el-table
        v-loading="loading"
        :data="requests"
        border
        stripe
        class="verification-table"
        empty-text="暂无收款审核申请"
        row-key="id"
      >
        <el-table-column prop="id" label="申请ID" width="80" />
        <el-table-column prop="order_number" label="订单编号" width="180" />
        <el-table-column label="销售员" width="150">
          <template #default="scope">
            <div>
              <div style="font-weight: 600;">{{ scope.row.sales_employee_name }}</div>
              <div style="color: #909399; font-size: 12px;">{{ scope.row.sales_employee_code }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="客户信息" width="150">
          <template #default="scope">
            <div>
              <div>{{ scope.row.customer_name }}</div>
              <div style="color: #909399; font-size: 12px;">ID: {{ scope.row.customer_id }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="订单金额" width="120" align="right">
          <template #default="scope">
            <span style="color: #ff4d4f; font-weight: 600;">
              ¥{{ formatMoney(scope.row.order_amount) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="申请原因" min-width="200">
          <template #default="scope">
            <span>{{ scope.row.request_reason || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="审核状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ formatStatus(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="审核信息" min-width="180">
          <template #default="scope">
            <div v-if="scope.row.status !== 'pending'">
              <div v-if="scope.row.admin_name" style="font-size: 12px;">
                <span style="color: #909399;">审核人：</span>{{ scope.row.admin_name }}
              </div>
              <div v-if="scope.row.reviewed_at" style="color: #909399; font-size: 12px;">
                {{ formatDateTime(scope.row.reviewed_at) }}
              </div>
              <div v-if="scope.row.review_remark" style="color: #606266; font-size: 12px; margin-top: 4px;">
                备注：{{ scope.row.review_remark }}
              </div>
            </div>
            <span v-else style="color: #c0c4cc;">待审核</span>
          </template>
        </el-table-column>
        <el-table-column label="申请时间" width="160">
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="scope">
            <div v-if="scope.row.status === 'pending'">
              <el-button
                type="success"
                size="small"
                @click="handleApprove(scope.row)"
              >
                通过
              </el-button>
              <el-button
                type="danger"
                size="small"
                @click="handleReject(scope.row)"
              >
                拒绝
              </el-button>
            </div>
            <el-button
              v-else
              type="info"
              size="small"
              @click="handleViewDetail(scope.row)"
            >
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="total > 0"
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        class="pagination"
      />
    </el-card>

    <!-- 审核对话框 -->
    <el-dialog
      v-model="reviewDialogVisible"
      :title="reviewAction === 'approve' ? '通过收款申请' : '拒绝收款申请'"
      width="500px"
    >
      <el-form :model="reviewForm" label-width="100px">
        <el-form-item label="订单编号">
          <span>{{ currentRequest?.order_number }}</span>
        </el-form-item>
        <el-form-item label="销售员">
          <span>{{ currentRequest?.sales_employee_name }} ({{ currentRequest?.sales_employee_code }})</span>
        </el-form-item>
        <el-form-item label="订单金额">
          <span style="color: #ff4d4f; font-weight: 600;">
            ¥{{ formatMoney(currentRequest?.order_amount) }}
          </span>
        </el-form-item>
        <el-form-item label="申请原因">
          <span>{{ currentRequest?.request_reason || '-' }}</span>
        </el-form-item>
        <el-form-item label="审核备注">
          <el-input
            v-model="reviewForm.remark"
            type="textarea"
            :rows="4"
            :placeholder="reviewAction === 'approve' ? '请输入审核备注（可选）' : '请输入拒绝原因'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewDialogVisible = false">取消</el-button>
        <el-button
          :type="reviewAction === 'approve' ? 'success' : 'danger'"
          @click="handleConfirmReview"
          :loading="submitting"
        >
          确认{{ reviewAction === 'approve' ? '通过' : '拒绝' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="收款申请详情"
      width="600px"
    >
      <el-descriptions :column="2" border v-if="currentRequest">
        <el-descriptions-item label="申请ID">{{ currentRequest.id }}</el-descriptions-item>
        <el-descriptions-item label="订单编号">{{ currentRequest.order_number }}</el-descriptions-item>
        <el-descriptions-item label="销售员">
          {{ currentRequest.sales_employee_name }} ({{ currentRequest.sales_employee_code }})
        </el-descriptions-item>
        <el-descriptions-item label="客户">
          {{ currentRequest.customer_name }} (ID: {{ currentRequest.customer_id }})
        </el-descriptions-item>
        <el-descriptions-item label="订单金额">
          <span style="color: #ff4d4f; font-weight: 600;">
            ¥{{ formatMoney(currentRequest.order_amount) }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="getStatusType(currentRequest.status)">
            {{ formatStatus(currentRequest.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="申请原因" :span="2">
          {{ currentRequest.request_reason || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="申请时间" :span="2">
          {{ formatDateTime(currentRequest.created_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="审核人" v-if="currentRequest.admin_name">
          {{ currentRequest.admin_name }}
        </el-descriptions-item>
        <el-descriptions-item label="审核时间" v-if="currentRequest.reviewed_at">
          {{ formatDateTime(currentRequest.reviewed_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="审核备注" :span="2" v-if="currentRequest.review_remark">
          {{ currentRequest.review_remark }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPaymentVerificationRequests, reviewPaymentVerification } from '../api/paymentVerification'
import { formatDateTime } from '../utils/time-format'

// 数据
const loading = ref(false)
const requests = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const statusFilter = ref('pending') // 默认显示待审核
const stats = ref(null)

// 审核对话框
const reviewDialogVisible = ref(false)
const reviewAction = ref('approve') // approve 或 reject
const currentRequest = ref(null)
const reviewForm = ref({
  remark: ''
})
const submitting = ref(false)

// 详情对话框
const detailDialogVisible = ref(false)

// 统计数据
const calculateStats = () => {
  const pending = requests.value.filter(r => r.status === 'pending').length
  const approved = requests.value.filter(r => r.status === 'approved').length
  const rejected = requests.value.filter(r => r.status === 'rejected').length
  stats.value = {
    pending,
    approved,
    rejected,
    total: total.value
  }
}

// 获取列表
const fetchRequests = async () => {
  loading.value = true
  try {
    const params = {
      pageNum: currentPage.value,
      pageSize: pageSize.value,
      status: statusFilter.value
    }
    const res = await getPaymentVerificationRequests(params)
    if (res.code === 200) {
      requests.value = res.data.list || []
      total.value = res.data.total || 0
      calculateStats()
    } else {
      ElMessage.error(res.message || '获取列表失败')
    }
  } catch (error) {
    console.error('获取收款审核列表失败:', error)
    ElMessage.error('获取列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  fetchRequests()
}

// 刷新
const handleRefresh = () => {
  fetchRequests()
}

// 分页
const handlePageChange = (page) => {
  currentPage.value = page
  fetchRequests()
}

const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  fetchRequests()
}

// 通过申请
const handleApprove = (row) => {
  currentRequest.value = row
  reviewAction.value = 'approve'
  reviewForm.value.remark = ''
  reviewDialogVisible.value = true
}

// 拒绝申请
const handleReject = (row) => {
  currentRequest.value = row
  reviewAction.value = 'reject'
  reviewForm.value.remark = ''
  reviewDialogVisible.value = true
}

// 确认审核
const handleConfirmReview = async () => {
  if (reviewAction.value === 'reject' && !reviewForm.value.remark) {
    ElMessage.warning('请输入拒绝原因')
    return
  }

  submitting.value = true
  try {
    const data = {
      request_id: currentRequest.value.id,
      approved: reviewAction.value === 'approve',
      review_remark: reviewForm.value.remark
    }
    const res = await reviewPaymentVerification(currentRequest.value.id, data)
    if (res.code === 200) {
      ElMessage.success(reviewAction.value === 'approve' ? '已通过申请' : '已拒绝申请')
      reviewDialogVisible.value = false
      fetchRequests()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (error) {
    console.error('审核失败:', error)
    ElMessage.error('操作失败')
  } finally {
    submitting.value = false
  }
}

// 查看详情
const handleViewDetail = (row) => {
  currentRequest.value = row
  detailDialogVisible.value = true
}

// 格式化金额
const formatMoney = (amount) => {
  if (amount === null || amount === undefined) return '0.00'
  return Number(amount).toFixed(2)
}

// 格式化状态
const formatStatus = (status) => {
  const statusMap = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝'
  }
  return statusMap[status] || status
}

// 获取状态类型
const getStatusType = (status) => {
  const typeMap = {
    pending: 'warning',
    approved: 'success',
    rejected: 'danger'
  }
  return typeMap[status] || 'info'
}

// 初始化
onMounted(() => {
  fetchRequests()
})
</script>

<style scoped>
.payment-verification-page {
  padding: 20px;
}

.verification-card {
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title {
  display: flex;
  flex-direction: column;
}

.title .main {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.title .sub {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.actions {
  display: flex;
  align-items: center;
}

.stats-summary {
  margin-bottom: 20px;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
}

.stat-item {
  text-align: center;
  color: white;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 32px;
  font-weight: 600;
}

.stat-value.pending {
  color: #ffd700;
}

.stat-value.approved {
  color: #67c23a;
}

.stat-value.rejected {
  color: #f56c6c;
}

.stat-value.total {
  color: #ffffff;
}

.verification-table {
  margin-top: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

