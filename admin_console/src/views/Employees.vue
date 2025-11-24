<template>
  <div class="employees-page">
    <el-card class="employees-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">员工管理</span>
          <span class="sub">管理配送员和销售员信息</span>
        </div>
        <div class="actions">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索员工码 / 手机号 / 姓名"
            clearable
            @keyup.enter="handleSearch"
            style="width: 250px;"
          />
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button type="success" @click="handleAdd">添加员工</el-button>
        </div>
      </div>

      <el-table
        v-loading="loading"
        :data="employees"
        border
        stripe
        class="employees-table"
      >
        <el-table-column prop="id" label="ID" min-width="60" />
        <el-table-column prop="employee_code" label="员工码" min-width="100">
          <template #default="scope">
            <span class="employee-code-text">{{ scope.row.employee_code }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="姓名" min-width="100">
          <template #default="scope">
            {{ scope.row.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" min-width="130" />
        <el-table-column prop="is_delivery" label="配送员" min-width="100">
          <template #default="scope">
            <el-tag :type="scope.row.is_delivery ? 'success' : 'info'">
              {{ scope.row.is_delivery ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_sales" label="销售员" min-width="100">
          <template #default="scope">
            <el-tag :type="scope.row.is_sales ? 'success' : 'info'">
              {{ scope.row.is_sales ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="customer_count" label="绑定客户数" min-width="120" v-if="employees.some(e => e.is_sales)">
          <template #default="scope">
            <span v-if="scope.row.is_sales">
              <el-button 
                type="primary" 
                link 
                @click="handleViewCustomers(scope.row)"
              >
                {{ scope.row.customer_count || 0 }} 个
              </el-button>
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status ? 'success' : 'danger'">
              {{ scope.row.status ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="160">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="150" fixed="right">
          <template #default="scope">
            <el-button type="primary" link @click="handleEdit(scope.row)">
              编辑
            </el-button>
            <el-button type="danger" link @click="handleDelete(scope.row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

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

    <!-- 添加/编辑员工对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="120px"
      >
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码（至少6位）"
            show-password
          />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model="form.name" placeholder="请输入姓名（可选）" />
        </el-form-item>
        <el-form-item label="角色" prop="roles">
          <el-checkbox-group v-model="form.roles">
            <el-checkbox label="delivery">配送员</el-checkbox>
            <el-checkbox label="sales">销售员</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="状态" v-if="isEdit">
          <el-switch
            v-model="form.status"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">
            保存
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 客户列表对话框 -->
    <el-dialog
      v-model="customersDialogVisible"
      :title="`销售员 ${selectedEmployee?.name || selectedEmployee?.employee_code || ''} 的绑定客户`"
      width="800px"
      :close-on-click-modal="false"
    >
      <div v-if="selectedEmployee">
        <div style="margin-bottom: 16px; color: #606266;">
          <span>员工码：<strong>{{ selectedEmployee.employee_code }}</strong></span>
          <span style="margin-left: 20px;">绑定客户数：<strong>{{ selectedEmployee.customer_count || 0 }}</strong></span>
        </div>
        <el-table
          :data="selectedEmployee.customers || []"
          border
          stripe
          v-loading="customersLoading"
          max-height="500"
        >
          <!-- <el-table-column prop="id" label="ID" width="80" /> -->
          <el-table-column prop="user_code" label="用户编号" min-width="100" />
          <!-- <el-table-column prop="name" label="姓名" min-width="120">
            <template #default="scope">
              {{ scope.row.name || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="phone" label="手机号" min-width="130">
            <template #default="scope">
              {{ scope.row.phone || '-' }}
            </template>
          </el-table-column> -->
          <el-table-column prop="default_address" label="默认地址" min-width="200">
            <template #default="scope">
              <div v-if="scope.row.default_address" class="address-info">
                <div class="address-item">
                  <span class="address-label">店铺：</span>
                  <span class="address-value">{{ scope.row.default_address.name || '-' }}</span>
                </div>
                <div class="address-item">
                  <span class="address-label">联系人：</span>
                  <span class="address-value">{{ scope.row.default_address.contact || '-' }}</span>
                </div>
                <div class="address-item">
                  <span class="address-label">电话：</span>
                  <span class="address-value">{{ scope.row.default_address.phone || '-' }}</span>
                </div>
                <div class="address-item">
                  <span class="address-label">地址：</span>
                  <span class="address-value">{{ scope.row.default_address.address || '-' }}</span>
                </div>
                <div class="address-item" v-if="scope.row.default_address.store_type">
                  <span class="address-label">类型：</span>
                  <span class="address-value">{{ scope.row.default_address.store_type }}</span>
                </div>
              </div>
              <span v-else class="no-address">暂无默认地址</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="绑定时间" min-width="100">
            <template #default="scope">
              {{ formatDate(scope.row.created_at) }}
            </template>
          </el-table-column>
        </el-table>
        <div v-if="!selectedEmployee.customers || selectedEmployee.customers.length === 0" 
             style="text-align: center; padding: 40px; color: #909399;">
          暂无绑定客户
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="customersDialogVisible = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getEmployees, createEmployee, updateEmployee, deleteEmployee } from '../api/employees'

const loading = ref(false)
const employees = ref([])
const searchKeyword = ref('')
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 对话框相关
const dialogVisible = ref(false)
const dialogTitle = ref('添加员工')
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  phone: '',
  password: '',
  name: '',
  roles: [],
  status: true
})

const formRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    {
      pattern: /^1[3-9]\d{9}$/,
      message: '请输入正确的手机号码',
      trigger: 'blur'
    }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  roles: [
    {
      validator: (rule, value, callback) => {
        if (value.length === 0) {
          callback(new Error('至少需要选择一种角色'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ]
}

const loadEmployees = async () => {
  loading.value = true
  try {
    const res = await getEmployees({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      keyword: searchKeyword.value
    })
    if (res.code === 200) {
      employees.value = Array.isArray(res.data) ? res.data : []
      pagination.total = res.total || employees.value.length
    }
  } catch (error) {
    console.error('获取员工列表失败:', error)
    ElMessage.error('获取员工列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.pageNum = 1
  loadEmployees()
}

const handlePageChange = (page) => {
  pagination.pageNum = page
  loadEmployees()
}

const formatDate = (value) => {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

const handleAdd = () => {
  dialogTitle.value = '添加员工'
  isEdit.value = false
  form.phone = ''
  form.password = ''
  form.name = ''
  form.roles = []
  form.status = true
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑员工'
  isEdit.value = true
  form.phone = row.phone
  form.password = '' // 编辑时不显示密码
  form.name = row.name || ''
  form.roles = []
  if (row.is_delivery) form.roles.push('delivery')
  if (row.is_sales) form.roles.push('sales')
  form.status = row.status
  form.id = row.id
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      const data = {
        phone: form.phone,
        name: form.name,
        is_delivery: form.roles.includes('delivery'),
        is_sales: form.roles.includes('sales')
      }

      if (isEdit.value) {
        // 编辑
        if (form.password) {
          data.password = form.password
        }
        data.status = form.status

        const res = await updateEmployee(form.id, data)
        if (res.code === 200) {
          ElMessage.success('更新成功')
          dialogVisible.value = false
          await loadEmployees()
        } else {
          ElMessage.error(res.message || '更新失败')
        }
      } else {
        // 添加
        data.password = form.password

        const res = await createEmployee(data)
        if (res.code === 200) {
          ElMessage.success('创建成功')
          dialogVisible.value = false
          await loadEmployees()
        } else {
          ElMessage.error(res.message || '创建失败')
        }
      }
    } catch (error) {
      console.error('保存员工失败:', error)
      ElMessage.error('保存失败，请稍后再试')
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除员工 "${row.name || row.phone}" 吗？`,
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
    .then(async () => {
      try {
        const res = await deleteEmployee(row.id)
        if (res.code === 200) {
          ElMessage.success('删除成功')
          await loadEmployees()
        } else {
          ElMessage.error(res.message || '删除失败')
        }
      } catch (error) {
        console.error('删除员工失败:', error)
        ElMessage.error('删除失败，请稍后再试')
      }
    })
    .catch(() => {})
}

// 客户列表对话框
const customersDialogVisible = ref(false)
const selectedEmployee = ref(null)
const customersLoading = ref(false)

const handleViewCustomers = (employee) => {
  selectedEmployee.value = employee
  customersDialogVisible.value = true
  // 如果客户数据不存在，可以在这里重新加载
  // 但后端已经在 GetEmployees 中返回了 customers 数据，所以不需要重新加载
}

onMounted(() => {
  loadEmployees()
})
</script>

<style scoped>
.employees-page {
  padding: 20px 0;
}

.employees-card {
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
  align-items: center;
}

.employees-table {
  margin-top: 10px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.employee-code-text {
  font-weight: 600;
  color: #409eff;
  font-size: 15px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.address-info {
  padding: 8px 0;
  font-size: 13px;
  line-height: 1.6;
}

.address-item {
  margin-bottom: 4px;
  display: flex;
  align-items: flex-start;
}

.address-item:last-child {
  margin-bottom: 0;
}

.address-label {
  color: #909399;
  font-weight: 500;
  min-width: 60px;
  flex-shrink: 0;
}

.address-value {
  color: #606266;
  flex: 1;
  word-break: break-all;
}

.no-address {
  color: #c0c4cc;
  font-style: italic;
}
</style>

