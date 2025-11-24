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
        <el-table-column prop="user_code" label="用户编号" min-width="120">
          <template #default="scope">
            <span v-if="scope.row.user_code" class="user-code-text">用户{{ scope.row.user_code }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" min-width="130">
          <template #default="scope">
            {{ scope.row.phone || '-' }}
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
        <el-table-column prop="sales_employee" label="绑定销售员" min-width="180">
          <template #default="scope">
            <div v-if="scope.row.sales_employee" class="sales-employee-info">
              <span class="sales-employee-name">
                {{ scope.row.sales_employee.name || '未命名' }}
              </span>
              <el-tag size="small" type="info" style="margin-left: 8px;">
                {{ scope.row.sales_employee.employee_code }}
              </el-tag>
            </div>
            <span v-else class="text-muted">-</span>
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
        :lock-scroll="true"
        :modal="true"
        class="user-detail-dialog"
      >
        <div v-loading="detailLoading" class="user-detail">
          <div v-if="userDetail" class="detail-content">
            <!-- 基本信息（包含头像） -->
            <div class="detail-section">
              <div class="section-title">基本信息</div>
              <div class="section-content">
                <div class="basic-info-wrapper">
                  <!-- 头像 -->
                  <div class="avatar-container">
                    <el-image
                      v-if="userDetail.avatar"
                      :src="userDetail.avatar"
                      class="avatar-image-small"
                      fit="cover"
                      :preview-src-list="[userDetail.avatar]"
                    />
                    <div v-else class="no-avatar-small">
                      <el-icon :size="20"><Picture /></el-icon>
                    </div>
                  </div>
                  <!-- 信息列表 -->
                  <div class="info-list">
                    <el-descriptions :column="2" border class="custom-descriptions">
                      <el-descriptions-item label="用户ID" label-class-name="desc-label">
                        <span class="desc-value">{{ userDetail.id }}</span>
                      </el-descriptions-item>
                      <el-descriptions-item label="用户编号" label-class-name="desc-label">
                        <span v-if="userDetail.user_code" class="desc-value user-code-text">用户{{ userDetail.user_code }}</span>
                        <span v-else class="desc-value">-</span>
                      </el-descriptions-item>
                      <el-descriptions-item label="唯一ID" label-class-name="desc-label">
                        <span class="desc-value unique-id">{{ userDetail.unique_id }}</span>
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
                      <el-descriptions-item label="资料完善" label-class-name="desc-label">
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
              </div>
            </div>

            <!-- 地址列表 -->
            <div class="detail-section">
              <div class="section-title">
                <span>收货地址</span>
                <span class="address-count" v-if="userDetail.addresses && userDetail.addresses.length > 0">
                  (共{{ userDetail.addresses.length }}个)
                </span>
              </div>
              <div class="section-content">
                <div v-if="userDetail.addresses && userDetail.addresses.length > 0" class="addresses-list">
                  <div 
                    v-for="address in userDetail.addresses" 
                    :key="address.id" 
                    class="address-item"
                    :class="{ 'is-default': address.is_default }"
                  >
                    <div class="address-header">
                      <el-tag v-if="address.is_default" type="success" size="small">默认地址</el-tag>
                      <el-button 
                        type="primary" 
                        link 
                        size="small" 
                        @click="handleEditAddress(address)"
                        style="margin-left: auto;"
                      >
                        编辑
                      </el-button>
                    </div>
                    <el-descriptions :column="2" border class="address-descriptions">
                      <el-descriptions-item label="地址名称" label-class-name="desc-label">
                        <span class="desc-value">{{ address.name || '-' }}</span>
                      </el-descriptions-item>
                      <el-descriptions-item label="联系人" label-class-name="desc-label">
                        <span class="desc-value">{{ address.contact || '-' }}</span>
                      </el-descriptions-item>
                      <el-descriptions-item label="手机号" label-class-name="desc-label">
                        <span class="desc-value phone-number">{{ address.phone || '-' }}</span>
                      </el-descriptions-item>
                      <el-descriptions-item label="店铺类型" label-class-name="desc-label">
                        <span class="desc-value">{{ address.store_type || '-' }}</span>
                      </el-descriptions-item>
                      <el-descriptions-item label="详细地址" label-class-name="desc-label" :span="2">
                        <span class="desc-value address-text">{{ address.address || '-' }}</span>
                      </el-descriptions-item>
                      <el-descriptions-item label="经纬度" label-class-name="desc-label">
                        <span v-if="address.latitude && address.longitude" class="desc-value coordinates">
                          {{ address.latitude }}, {{ address.longitude }}
                        </span>
                        <span v-else class="desc-value">-</span>
                      </el-descriptions-item>
                      <el-descriptions-item label="门头照片" label-class-name="desc-label" v-if="address.avatar" :span="2">
                        <el-image
                          :src="address.avatar"
                          class="address-avatar-image"
                          fit="cover"
                          :preview-src-list="[address.avatar]"
                        />
                      </el-descriptions-item>
                    </el-descriptions>
                  </div>
                </div>
                <el-empty v-else description="暂无地址" :image-size="80" />
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
          <el-form-item label="手机号" prop="phone">
            <el-input v-model="editForm.phone" placeholder="请输入手机号" />
          </el-form-item>
          <el-form-item label="店铺类型" prop="storeType">
            <el-input v-model="editForm.storeType" placeholder="请输入店铺类型" />
          </el-form-item>
          <el-form-item label="绑定销售员" prop="salesEmployeeId">
            <el-select 
              v-model="editForm.salesEmployeeId" 
              placeholder="请选择销售员" 
              clearable
              style="width: 100%"
              @change="handleSalesEmployeeChange"
            >
              <el-option
                v-for="emp in salesEmployees"
                :key="emp.id"
                :label="`${emp.name || emp.employee_code} (${emp.employee_code})`"
                :value="emp.id"
              >
                <span>{{ emp.name || '未命名' }}</span>
                <span style="color: #8492a6; font-size: 13px; margin-left: 8px;">{{ emp.employee_code }}</span>
              </el-option>
            </el-select>
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

      <!-- 编辑地址对话框 -->
      <el-dialog
        v-model="addressEditDialogVisible"
        title="编辑地址"
        width="700px"
        :close-on-click-modal="false"
      >
        <el-form
          ref="addressEditFormRef"
          :model="addressEditForm"
          :rules="addressEditFormRules"
          label-width="120px"
        >
          <el-form-item label="地址名称" prop="name">
            <el-input v-model="addressEditForm.name" placeholder="请输入地址名称" />
          </el-form-item>
          <el-form-item label="联系人" prop="contact">
            <el-input v-model="addressEditForm.contact" placeholder="请输入联系人" />
          </el-form-item>
          <el-form-item label="手机号" prop="phone">
            <el-input v-model="addressEditForm.phone" placeholder="请输入手机号" />
          </el-form-item>
          <el-form-item label="详细地址" prop="address">
            <el-input 
              v-model="addressEditForm.address" 
              type="textarea" 
              :rows="3"
              placeholder="请输入详细地址" 
            />
          </el-form-item>
          <el-form-item label="店铺类型">
            <el-input v-model="addressEditForm.storeType" placeholder="请输入店铺类型" />
          </el-form-item>
          <el-form-item label="门头照片URL">
            <el-input v-model="addressEditForm.avatar" placeholder="请输入门头照片URL" />
            <div v-if="addressEditForm.avatar" style="margin-top: 10px;">
              <el-image
                :src="addressEditForm.avatar"
                style="width: 100px; height: 100px; border-radius: 8px;"
                fit="cover"
              />
            </div>
          </el-form-item>
          <el-form-item label="经度">
            <el-input-number 
              v-model="addressEditForm.longitude" 
              :precision="6"
              :step="0.000001"
              placeholder="请输入经度"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="纬度">
            <el-input-number 
              v-model="addressEditForm.latitude" 
              :precision="6"
              :step="0.000001"
              placeholder="请输入纬度"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="设为默认地址">
            <el-switch
              v-model="addressEditForm.isDefault"
              active-text="是"
              inactive-text="否"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <span class="dialog-footer">
            <el-button @click="addressEditDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="addressEditSubmitting" @click="handleSaveAddressEdit">
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
import { getMiniUsers, getMiniUserDetail, updateMiniUser, getAdminAddressDetail, updateAdminAddress, getSalesEmployees } from '../api/miniUsers'

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
  phone: '',
  storeType: '',
  salesCode: '',
  salesEmployeeId: null,
  avatar: '',
  userType: 'unknown',
  profileCompleted: false
})
const salesEmployees = ref([])

const editFormRules = {
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

// 地址编辑相关
const addressEditDialogVisible = ref(false)
const addressEditSubmitting = ref(false)
const addressEditFormRef = ref(null)
const addressEditForm = reactive({
  name: '',
  contact: '',
  phone: '',
  address: '',
  avatar: '',
  storeType: '',
  latitude: null,
  longitude: null,
  isDefault: false
})
const editingAddressId = ref(null)

const addressEditFormRules = {
  name: [
    { required: true, message: '请输入地址名称', trigger: 'blur' }
  ],
  contact: [
    { required: true, message: '请输入联系人', trigger: 'blur' }
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (!value || value.trim() === '') {
          callback(new Error('请输入手机号'))
        } else if (!/^1[3-9]\d{9}$/.test(value)) {
          callback(new Error('请输入正确的手机号码'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  address: [
    { required: true, message: '请输入详细地址', trigger: 'blur' }
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

const loadSalesEmployees = async () => {
  try {
    const res = await getSalesEmployees()
    if (res.code === 200) {
      salesEmployees.value = res.data || []
    }
  } catch (error) {
    console.error('获取销售员列表失败:', error)
  }
}

const handleEdit = async () => {
  if (!userDetail.value) return
  
  // 加载销售员列表
  await loadSalesEmployees()
  
  // 填充编辑表单
  editForm.phone = userDetail.value.phone || ''
  editForm.storeType = userDetail.value.store_type || ''
  editForm.salesCode = userDetail.value.sales_code || ''
  editForm.salesEmployeeId = null
  
  // 根据sales_code找到对应的销售员ID
  if (editForm.salesCode && salesEmployees.value.length > 0) {
    const salesEmployee = salesEmployees.value.find(emp => emp.employee_code === editForm.salesCode)
    if (salesEmployee) {
      editForm.salesEmployeeId = salesEmployee.id
    }
  }
  
  editForm.avatar = userDetail.value.avatar || ''
  editForm.userType = userDetail.value.user_type || 'unknown'
  editForm.profileCompleted = userDetail.value.profile_completed || false
  
  editDialogVisible.value = true
}

const handleSalesEmployeeChange = (employeeId) => {
  if (employeeId) {
    const employee = salesEmployees.value.find(emp => emp.id === employeeId)
    if (employee) {
      editForm.salesCode = employee.employee_code
    }
  } else {
    editForm.salesCode = ''
  }
}

const handleSaveEdit = async () => {
  if (!editFormRef.value) return
  
  await editFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    if (!userDetail.value) return
    
    editSubmitting.value = true
    try {
      const updateData = {
        phone: editForm.phone,
        storeType: editForm.storeType,
        avatar: editForm.avatar,
        userType: editForm.userType,
        profileCompleted: editForm.profileCompleted
      }
      
      // 处理销售员绑定
      if (editForm.salesEmployeeId) {
        updateData.salesEmployeeId = editForm.salesEmployeeId
      } else {
        // 如果清空了选择，清除绑定
        updateData.salesCode = ''
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

const handleEditAddress = async (address) => {
  editingAddressId.value = address.id
  addressEditForm.name = address.name || ''
  addressEditForm.contact = address.contact || ''
  addressEditForm.phone = address.phone || ''
  addressEditForm.address = address.address || ''
  addressEditForm.avatar = address.avatar || ''
  addressEditForm.storeType = address.store_type || ''
  addressEditForm.latitude = address.latitude || null
  addressEditForm.longitude = address.longitude || null
  addressEditForm.isDefault = address.is_default || false
  
  addressEditDialogVisible.value = true
}

const handleSaveAddressEdit = async () => {
  if (!addressEditFormRef.value) return
  
  await addressEditFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    if (!editingAddressId.value) return
    
    addressEditSubmitting.value = true
    try {
      const updateData = {
        name: addressEditForm.name,
        contact: addressEditForm.contact,
        phone: addressEditForm.phone,
        address: addressEditForm.address,
        avatar: addressEditForm.avatar,
        storeType: addressEditForm.storeType,
        latitude: addressEditForm.latitude,
        longitude: addressEditForm.longitude,
        isDefault: addressEditForm.isDefault
      }
      
      const res = await updateAdminAddress(editingAddressId.value, updateData)
      if (res.code === 200) {
        ElMessage.success('地址更新成功')
        addressEditDialogVisible.value = false
        // 刷新用户详情
        if (userDetail.value) {
          await handleViewDetail(userDetail.value.id)
        }
      } else {
        ElMessage.error(res.message || '更新失败')
      }
    } catch (error) {
      console.error('更新地址失败:', error)
      ElMessage.error('更新地址失败，请稍后再试')
    } finally {
      addressEditSubmitting.value = false
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
  .el-dialog {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    margin: 0;
  }
  
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

/* 优化整体间距 */
:deep(.user-detail-dialog) {
  .el-dialog__body {
    padding: 20px 24px;
  }
}

.detail-section {
  margin-bottom: 24px;
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  border: 1px solid #ebeef5;
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

/* 基本信息包装器 */
.basic-info-wrapper {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.avatar-container {
  flex-shrink: 0;
}

/* 头像样式（小尺寸） */
.avatar-image-small {
  width: 50px;
  height: 50px;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
  cursor: pointer;
  transition: all 0.3s;
}

.avatar-image-small:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  transform: scale(1.05);
}

.no-avatar-small {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 50px;
  height: 50px;
  border-radius: 6px;
  border: 1px dashed #dcdfe6;
  background: #f5f7fa;
  color: #909399;
}

.info-list {
  flex: 1;
  min-width: 0;
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
    width: 120px;
    padding: 10px 14px;
    border-right: 1px solid #ebeef5;
    font-size: 13px;
  }
  
  .el-descriptions__content {
    padding: 10px 14px;
    color: #303133;
    background: #fff;
    font-size: 13px;
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

.user-code-text {
  font-weight: 600;
  color: #409eff;
  font-size: 15px;
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

/* 地址列表样式 */
.addresses-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.address-item {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  background: #fff;
  transition: all 0.3s;
}

.address-item:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border-color: #c0c4cc;
}

.address-item.is-default {
  border-color: #67c23a;
  background: linear-gradient(135deg, #f0f9ff 0%, #f0fdf4 100%);
}

.address-header {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.address-count {
  font-size: 14px;
  color: #909399;
  font-weight: normal;
  margin-left: 8px;
}

:deep(.address-descriptions) {
  margin-top: 0;
}

:deep(.address-descriptions .el-descriptions__label) {
  width: 100px;
  font-size: 13px;
}

:deep(.address-descriptions .el-descriptions__content) {
  font-size: 13px;
}

.address-avatar-image {
  width: 80px;
  height: 80px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  cursor: pointer;
  transition: all 0.3s;
}

.address-avatar-image:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  transform: scale(1.05);
}

.sales-employee-info {
  display: flex;
  align-items: center;
}

.sales-employee-name {
  font-weight: 500;
  color: #303133;
}
</style>

