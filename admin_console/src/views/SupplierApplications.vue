<template>
  <div class="supplier-applications-page">
    <el-card class="supplier-applications-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">供应商合作申请管理</span>
          <span class="sub">查看和管理供应商合作申请</span>
        </div>
        <div class="actions">
          <el-select v-model="statusFilter" placeholder="申请状态" clearable style="width: 150px; margin-right: 10px;"
            @change="handleSearch">
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="applications" border stripe class="supplier-applications-table" empty-text="暂无申请数据" row-key="id">
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
            <span v-else-if="scope.row.user_id">用户ID: {{ scope.row.user_id }}</span>
            <span v-else style="color: #c0c4cc;">未登录用户</span>
          </template>
        </el-table-column>
        <el-table-column prop="company_name" label="公司名称" min-width="200" />
        <el-table-column prop="contact_name" label="联系人" width="120" />
        <el-table-column prop="contact_phone" label="联系电话" width="130" />
        <el-table-column prop="email" label="邮箱" width="180">
          <template #default="scope">
            {{ scope.row.email || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="main_category" label="主营类目" width="120" />
        <el-table-column prop="address" label="公司地址" min-width="200" show-overflow-tooltip>
          <template #default="scope">
            {{ scope.row.address || '-' }}
          </template>
        </el-table-column>
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
            <el-button type="success" link size="small" @click="handleUpdateStatus(scope.row, 'approved')"
              v-if="scope.row.status === 'pending'">
              通过
            </el-button>
            <el-button type="danger" link size="small" @click="handleUpdateStatus(scope.row, 'rejected')"
              v-if="scope.row.status === 'pending'">
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
      title="申请详情"
      width="800px"
      :close-on-click-modal="false"
    >
      <div v-loading="detailLoading" class="detail-content">
        <el-descriptions :column="2" border v-if="currentApplication">
          <el-descriptions-item label="申请ID">{{ currentApplication.id }}</el-descriptions-item>
          <el-descriptions-item label="用户ID">{{ currentApplication.user_id || '未登录用户' }}</el-descriptions-item>
          <el-descriptions-item label="用户姓名">{{ currentApplication.user_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="用户编号">用户{{ currentApplication.user_code || currentApplication.user_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="用户电话">{{ currentApplication.user_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentApplication.status)">
              {{ formatStatus(currentApplication.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="公司名称" :span="2">{{ currentApplication.company_name }}</el-descriptions-item>
          <el-descriptions-item label="联系人">{{ currentApplication.contact_name }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ currentApplication.contact_phone }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ currentApplication.email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="主营类目">{{ currentApplication.main_category }}</el-descriptions-item>
          <el-descriptions-item label="公司地址" :span="2">{{ currentApplication.address || '-' }}</el-descriptions-item>
          <el-descriptions-item label="公司简介" :span="2">
            <div style="white-space: pre-wrap;">{{ currentApplication.company_intro || '-' }}</div>
          </el-descriptions-item>
          <el-descriptions-item label="合作意向说明" :span="2">
            <div style="white-space: pre-wrap;">{{ currentApplication.cooperation_intent || '-' }}</div>
          </el-descriptions-item>
          <el-descriptions-item label="管理员备注" :span="2">
            <div style="white-space: pre-wrap;">{{ currentApplication.admin_remark || '-' }}</div>
          </el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ formatDate(currentApplication.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDate(currentApplication.updated_at) }}</el-descriptions-item>
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
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
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
import { ElMessage } from 'element-plus'
import { getSupplierApplications, updateSupplierApplicationStatus } from '../api/supplierApplications'

const loading = ref(false)
const applications = ref([])
const statusFilter = ref('')
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 详情对话框
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const currentApplication = ref(null)

// 状态更新对话框
const statusDialogVisible = ref(false)
const updating = ref(false)
const statusForm = reactive({
  status: '',
  admin_remark: ''
})

const loadApplications = async () => {
  loading.value = true
  try {
    const res = await getSupplierApplications({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      status: statusFilter.value
    })
    if (res.code === 200) {
      applications.value = Array.isArray(res.data) ? res.data : []
      pagination.total = res.total || applications.value.length
    }
  } catch (error) {
    console.error('获取申请列表失败:', error)
    ElMessage.error('获取申请列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.pageNum = 1
  loadApplications()
}

const handlePageChange = (page) => {
  pagination.pageNum = page
  loadApplications()
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.pageNum = 1
  loadApplications()
}

const formatDate = (value) => {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

const formatStatus = (status) => {
  const statusMap = {
    'pending': '待审核',
    'approved': '已通过',
    'rejected': '已拒绝'
  }
  return statusMap[status] || status
}

const getStatusType = (status) => {
  const typeMap = {
    'pending': 'info',
    'approved': 'success',
    'rejected': 'danger'
  }
  return typeMap[status] || ''
}

const handleViewDetail = (row) => {
  currentApplication.value = row
  detailDialogVisible.value = true
}

const handleUpdateStatus = (row, status) => {
  currentApplication.value = row
  statusForm.status = status
  statusForm.admin_remark = row.admin_remark || ''
  statusDialogVisible.value = true
}

const handleUpdateStatusDialog = () => {
  if (!currentApplication.value) return
  statusForm.status = currentApplication.value.status
  statusForm.admin_remark = currentApplication.value.admin_remark || ''
  statusDialogVisible.value = true
}

const handleConfirmUpdateStatus = async () => {
  if (!statusForm.status) {
    ElMessage.warning('请选择状态')
    return
  }

  if (!currentApplication.value) {
    ElMessage.error('未选择申请')
    return
  }

  updating.value = true
  try {
    const res = await updateSupplierApplicationStatus(currentApplication.value.id, {
      status: statusForm.status,
      admin_remark: statusForm.admin_remark
    })
    
    if (res.code === 200) {
      ElMessage.success('更新成功')
      statusDialogVisible.value = false
      // 刷新列表
      loadApplications()
      // 如果详情对话框打开，刷新详情
      if (detailDialogVisible.value) {
        const index = applications.value.findIndex(a => a.id === currentApplication.value.id)
        if (index !== -1) {
          currentApplication.value = applications.value[index]
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
  loadApplications()
})
</script>

<style scoped>
.supplier-applications-page {
  padding: 20px;
}

.supplier-applications-card {
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

.supplier-applications-table {
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

