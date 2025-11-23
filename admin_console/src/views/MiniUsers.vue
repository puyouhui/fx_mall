<template>
  <div class="mini-users-page">
    <el-card class="mini-users-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">小程序用户</span>
          <span class="sub">查看登录用户唯一ID与基础信息</span>
        </div>
        <div class="actions">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索唯一ID / 姓名 / 电话"
            clearable
            @keyup.enter="handleSearch"
          />
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </div>
      </div>

      <el-table
        v-loading="loading"
        :data="users"
        border
        stripe
        class="mini-users-table"
      >
        <!-- <el-table-column prop="unique_id" label="唯一ID" min-width="220" /> -->
        <el-table-column prop="id" label="ID" min-width="60" />
        <el-table-column prop="name" label="名称" min-width="120">
          <template #default="scope">
            {{ scope.row.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="contact" label="联系人" min-width="120">
          <template #default="scope">
            {{ scope.row.contact || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" min-width="130">
          <template #default="scope">
            {{ scope.row.phone || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="180" show-overflow-tooltip>
          <template #default="scope">
            {{ scope.row.address || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="经纬度" min-width="140">
          <template #default="scope">
            <span v-if="scope.row.latitude && scope.row.longitude">
              {{ scope.row.latitude }}, {{ scope.row.longitude }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="user_type" label="用户类型" min-width="120">
          <template #default="scope">
            <el-tag :type="scope.row.user_type === 'wholesale' ? 'warning' : 'success'">
              {{ formatUserType(scope.row.user_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="profile_completed" label="资料完善" min-width="110">
          <template #default="scope">
            <el-tag :type="scope.row.profile_completed ? 'success' : 'info'">
              {{ scope.row.profile_completed ? '已完善' : '未完善' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sales_code" label="销售员代码" min-width="120">
          <template #default="scope">
            {{ scope.row.sales_code || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="store_type" label="店铺类型" min-width="120">
          <template #default="scope">
            {{ scope.row.store_type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="首次登录时间" min-width="160">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="100" fixed="right">
          <template #default="scope">
            <el-button type="primary" link @click="handleViewDetail(scope.row.id)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 用户详情对话框 -->
      <el-dialog
        v-model="detailDialogVisible"
        title="用户详情"
        width="900px"
        :close-on-click-modal="false"
        class="user-detail-dialog"
      >
        <div v-loading="detailLoading" class="user-detail">
          <div v-if="userDetail" class="detail-content">
            <!-- 用户头像 -->
            <div class="detail-section avatar-section">
              <div class="section-title">头像</div>
              <div class="section-content">
                <el-image
                  v-if="userDetail.avatar"
                  :src="userDetail.avatar"
                  class="avatar-image"
                  fit="cover"
                  :preview-src-list="[userDetail.avatar]"
                />
                <div v-else class="no-avatar">
                  <el-icon :size="40"><Picture /></el-icon>
                  <span>未上传</span>
                </div>
              </div>
            </div>

            <!-- 基本信息 -->
            <div class="detail-section">
              <div class="section-title">基本信息</div>
              <div class="section-content">
                <el-descriptions :column="2" border class="custom-descriptions">
                  <el-descriptions-item label="用户ID" label-class-name="desc-label">
                    <span class="desc-value">{{ userDetail.id }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="唯一ID" label-class-name="desc-label">
                    <span class="desc-value unique-id">{{ userDetail.unique_id }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="店铺名称" label-class-name="desc-label">
                    <span class="desc-value">{{ userDetail.name || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="联系人" label-class-name="desc-label">
                    <span class="desc-value">{{ userDetail.contact || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="手机号" label-class-name="desc-label">
                    <span class="desc-value phone-number">{{ userDetail.phone || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="用户类型" label-class-name="desc-label">
                    <el-tag 
                      :type="userDetail.user_type === 'wholesale' ? 'warning' : (userDetail.user_type === 'retail' ? 'success' : 'info')"
                      class="user-type-tag"
                    >
                      {{ formatUserType(userDetail.user_type) }}
                    </el-tag>
                  </el-descriptions-item>
                  <el-descriptions-item label="资料完善" :span="2" label-class-name="desc-label">
                    <el-tag 
                      :type="userDetail.profile_completed ? 'success' : 'info'"
                      class="profile-tag"
                    >
                      {{ userDetail.profile_completed ? '已完善' : '未完善' }}
                    </el-tag>
                  </el-descriptions-item>
                </el-descriptions>
              </div>
            </div>

            <!-- 地址信息 -->
            <div class="detail-section">
              <div class="section-title">地址信息</div>
              <div class="section-content">
                <el-descriptions :column="1" border class="custom-descriptions">
                  <el-descriptions-item label="详细地址" label-class-name="desc-label">
                    <span class="desc-value address-text">{{ userDetail.address || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="经纬度" label-class-name="desc-label">
                    <span v-if="userDetail.latitude && userDetail.longitude" class="desc-value coordinates">
                      {{ userDetail.latitude }}, {{ userDetail.longitude }}
                    </span>
                    <span v-else class="desc-value">-</span>
                  </el-descriptions-item>
                </el-descriptions>
              </div>
            </div>

            <!-- 店铺信息 -->
            <div class="detail-section">
              <div class="section-title">店铺信息</div>
              <div class="section-content">
                <el-descriptions :column="2" border class="custom-descriptions">
                  <el-descriptions-item label="店铺类型" label-class-name="desc-label">
                    <span class="desc-value">{{ userDetail.store_type || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="销售员代码" label-class-name="desc-label">
                    <span class="desc-value">{{ userDetail.sales_code || '-' }}</span>
                  </el-descriptions-item>
                </el-descriptions>
              </div>
            </div>

            <!-- 时间信息 -->
            <div class="detail-section">
              <div class="section-title">时间信息</div>
              <div class="section-content">
                <el-descriptions :column="2" border class="custom-descriptions">
                  <el-descriptions-item label="首次登录时间" label-class-name="desc-label">
                    <span class="desc-value time-text">{{ formatDate(userDetail.created_at) }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="最近更新时间" label-class-name="desc-label">
                    <span class="desc-value time-text">{{ formatDate(userDetail.updated_at) }}</span>
                  </el-descriptions-item>
                </el-descriptions>
              </div>
            </div>
          </div>
        </div>
        <template #footer>
          <div class="dialog-footer">
            <el-button @click="detailDialogVisible = false">关闭</el-button>
            <el-button type="primary" @click="handleEdit">编辑</el-button>
          </div>
        </template>
      </el-dialog>

      <!-- 编辑用户对话框 -->
      <el-dialog
        v-model="editDialogVisible"
        title="编辑用户信息"
        width="700px"
        :close-on-click-modal="false"
      >
        <el-form
          ref="editFormRef"
          :model="editForm"
          :rules="editFormRules"
          label-width="120px"
        >
          <el-form-item label="店铺名称" prop="name">
            <el-input v-model="editForm.name" placeholder="请输入店铺名称" />
          </el-form-item>
          <el-form-item label="联系人" prop="contact">
            <el-input v-model="editForm.contact" placeholder="请输入联系人" />
          </el-form-item>
          <el-form-item label="手机号" prop="phone">
            <el-input v-model="editForm.phone" placeholder="请输入手机号" />
          </el-form-item>
          <el-form-item label="详细地址" prop="address">
            <el-input
              v-model="editForm.address"
              type="textarea"
              :rows="3"
              placeholder="请输入详细地址"
            />
          </el-form-item>
          <el-form-item label="店铺类型" prop="storeType">
            <el-input v-model="editForm.storeType" placeholder="请输入店铺类型" />
          </el-form-item>
          <el-form-item label="销售员代码" prop="salesCode">
            <el-input v-model="editForm.salesCode" placeholder="请输入销售员代码" maxlength="5" />
          </el-form-item>
          <el-form-item label="头像URL" prop="avatar">
            <el-input v-model="editForm.avatar" placeholder="请输入头像URL" />
            <div v-if="editForm.avatar" style="margin-top: 10px;">
              <el-image
                :src="editForm.avatar"
                style="width: 100px; height: 100px; border-radius: 8px;"
                fit="cover"
              />
            </div>
          </el-form-item>
          <el-form-item label="经度" prop="latitude">
            <el-input-number
              v-model="editForm.latitude"
              :precision="6"
              :step="0.000001"
              placeholder="请输入经度"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="纬度" prop="longitude">
            <el-input-number
              v-model="editForm.longitude"
              :precision="6"
              :step="0.000001"
              placeholder="请输入纬度"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="用户类型" prop="userType">
            <el-select v-model="editForm.userType" placeholder="请选择用户类型" style="width: 100%">
              <el-option label="未选择" value="unknown" />
              <el-option label="零售用户" value="retail" />
              <el-option label="批发用户" value="wholesale" />
            </el-select>
          </el-form-item>
          <el-form-item label="资料完善" prop="profileCompleted">
            <el-switch
              v-model="editForm.profileCompleted"
              active-text="已完善"
              inactive-text="未完善"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <span class="dialog-footer">
            <el-button @click="editDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="editSubmitting" @click="handleSaveEdit">
              保存
            </el-button>
          </span>
        </template>
      </el-dialog>

      <div class="pagination">
        <el-pagination
          background
          layout="total, prev, pager, next, jumper"
          :page-size="pagination.pageSize"
          :current-page="pagination.pageNum"
          :total="pagination.total"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Picture } from '@element-plus/icons-vue'
import { getMiniUsers, getMiniUserDetail, updateMiniUser } from '../api/miniUsers'

const loading = ref(false)
const users = ref([])
const searchKeyword = ref('')
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 用户详情相关
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const userDetail = ref(null)

// 编辑相关
const editDialogVisible = ref(false)
const editSubmitting = ref(false)
const editFormRef = ref(null)
const editForm = reactive({
  name: '',
  contact: '',
  phone: '',
  address: '',
  storeType: '',
  salesCode: '',
  avatar: '',
  latitude: null,
  longitude: null,
  userType: 'unknown',
  profileCompleted: false
})

const editFormRules = {
  name: [{ required: false, message: '请输入店铺名称', trigger: 'blur' }],
  contact: [{ required: false, message: '请输入联系人', trigger: 'blur' }],
  phone: [
    { required: false, message: '请输入手机号', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (!value || value.trim() === '') {
          callback()
        } else if (!/^1[3-9]\d{9}$/.test(value)) {
          callback(new Error('请输入正确的手机号码'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await getMiniUsers({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      keyword: searchKeyword.value
    })
    if (res.code === 200) {
      users.value = Array.isArray(res.data) ? res.data : []
      pagination.total = res.total || users.value.length
    }
  } catch (error) {
    console.error('获取用户失败:', error)
    ElMessage.error('获取用户列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.pageNum = 1
  loadUsers()
}

const handlePageChange = (page) => {
  pagination.pageNum = page
  loadUsers()
}

const formatDate = (value) => {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

const formatUserType = (type) => {
  if (type === 'wholesale') return '批发用户'
  if (type === 'retail') return '零售用户'
  return '未选择'
}

const handleViewDetail = async (id) => {
  detailDialogVisible.value = true
  detailLoading.value = true
  userDetail.value = null
  
  try {
    const res = await getMiniUserDetail(id)
    if (res.code === 200) {
      userDetail.value = res.data
    } else {
      ElMessage.error(res.message || '获取用户详情失败')
      detailDialogVisible.value = false
    }
  } catch (error) {
    console.error('获取用户详情失败:', error)
    ElMessage.error('获取用户详情失败，请稍后再试')
    detailDialogVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

const handleEdit = () => {
  if (!userDetail.value) return
  
  // 填充编辑表单
  editForm.name = userDetail.value.name || ''
  editForm.contact = userDetail.value.contact || ''
  editForm.phone = userDetail.value.phone || ''
  editForm.address = userDetail.value.address || ''
  editForm.storeType = userDetail.value.store_type || ''
  editForm.salesCode = userDetail.value.sales_code || ''
  editForm.avatar = userDetail.value.avatar || ''
  editForm.latitude = userDetail.value.latitude || null
  editForm.longitude = userDetail.value.longitude || null
  editForm.userType = userDetail.value.user_type || 'unknown'
  editForm.profileCompleted = userDetail.value.profile_completed || false
  
  editDialogVisible.value = true
}

const handleSaveEdit = async () => {
  if (!editFormRef.value) return
  
  await editFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    if (!userDetail.value) return
    
    editSubmitting.value = true
    try {
      const updateData = {
        name: editForm.name,
        contact: editForm.contact,
        phone: editForm.phone,
        address: editForm.address,
        storeType: editForm.storeType,
        salesCode: editForm.salesCode,
        avatar: editForm.avatar,
        userType: editForm.userType,
        profileCompleted: editForm.profileCompleted
      }
      
      if (editForm.latitude !== null) {
        updateData.latitude = editForm.latitude
      }
      if (editForm.longitude !== null) {
        updateData.longitude = editForm.longitude
      }
      
      const res = await updateMiniUser(userDetail.value.id, updateData)
      if (res.code === 200) {
        ElMessage.success('更新成功')
        editDialogVisible.value = false
        // 刷新用户详情
        await handleViewDetail(userDetail.value.id)
        // 刷新用户列表
        await loadUsers()
      } else {
        ElMessage.error(res.message || '更新失败')
      }
    } catch (error) {
      console.error('更新用户失败:', error)
      ElMessage.error('更新用户失败，请稍后再试')
    } finally {
      editSubmitting.value = false
    }
  })
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.mini-users-page {
  padding: 20px 0;
}

.mini-users-card {
  border: none;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.04);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.title .main {
  font-size: 20px;
  font-weight: 600;
  margin-right: 12px;
}

.title .sub {
  color: #909399;
  font-size: 14px;
}

.actions {
  display: flex;
  gap: 12px;
  min-width: 320px;
}

.mini-users-table {
  margin-top: 10px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

/* 用户详情对话框样式 */
:deep(.user-detail-dialog) {
  .el-dialog__body {
    padding: 24px;
    max-height: 70vh;
    overflow-y: auto;
  }
  
  .el-dialog__header {
    padding: 20px 24px 16px;
    border-bottom: 1px solid #f0f0f0;
  }
  
  .el-dialog__title {
    font-size: 18px;
    font-weight: 600;
    color: #303133;
  }
  
  .el-dialog__footer {
    padding: 16px 24px;
    border-top: 1px solid #f0f0f0;
  }
}

.user-detail {
  min-height: 200px;
}

.detail-content {
  padding: 0;
}

.detail-section {
  margin-bottom: 32px;
  background: #fff;
}

.detail-section:last-child {
  margin-bottom: 0;
}

.avatar-section {
  margin-bottom: 28px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 10px;
  /* border-bottom: 2px solid #409eff; */
  position: relative;
}

.section-title::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 60px;
  height: 2px;
  background: #409eff;
}

.section-content {
  margin-top: 16px;
}

/* 头像样式 */
.avatar-image {
  width: 120px;
  height: 120px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  cursor: pointer;
  transition: all 0.3s;
}

.avatar-image:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  transform: scale(1.05);
}

.no-avatar {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 120px;
  height: 120px;
  border-radius: 8px;
  border: 1px dashed #dcdfe6;
  background: #f5f7fa;
  color: #909399;
  font-size: 14px;
  gap: 8px;
}

/* 描述列表样式 */
:deep(.custom-descriptions) {
  .el-descriptions__table {
    border-collapse: separate;
    border-spacing: 0;
  }
  
  .el-descriptions__label {
    background: #f8f9fa;
    font-weight: 500;
    color: #606266;
    width: 140px;
    padding: 12px 16px;
    border-right: 1px solid #ebeef5;
  }
  
  .el-descriptions__content {
    padding: 12px 16px;
    color: #303133;
    background: #fff;
  }
  
  .el-descriptions__cell {
    border-bottom: 1px solid #ebeef5;
  }
  
  .el-descriptions__cell:last-child {
    border-bottom: none;
  }
  
  .desc-label {
    background: #f8f9fa !important;
  }
}

.desc-value {
  color: #303133;
  font-size: 14px;
  word-break: break-all;
}

.unique-id {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #606266;
}

.phone-number {
  font-weight: 500;
  color: #303133;
}

.address-text {
  line-height: 1.6;
  color: #303133;
}

.coordinates {
  font-family: 'Courier New', monospace;
  color: #606266;
}

.time-text {
  color: #606266;
  font-size: 13px;
}

.user-type-tag,
.profile-tag {
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 12px;
}

/* 对话框底部按钮 */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.dialog-footer .el-button {
  padding: 10px 20px;
  font-size: 14px;
  border-radius: 4px;
}

.dialog-footer .el-button--primary {
  background: #409eff;
  border-color: #409eff;
}

.dialog-footer .el-button--primary:hover {
  background: #66b1ff;
  border-color: #66b1ff;
}
</style>

